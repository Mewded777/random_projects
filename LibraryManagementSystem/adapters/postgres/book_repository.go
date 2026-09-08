package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"library/domain"
)

// BookRepository is a Postgres-backed implementation of ports.BookRepository.
type BookRepository struct {
	db *sql.DB
}

func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{db: db}
}

const bookSelect = `
	SELECT b.book_id, b.title, b.total_copies, b.available_copies,
	       COALESCE(c.category_name, ''), COALESCE(c.language, ''),
	       COALESCE(STRING_AGG(TRIM(BOTH ' ' FROM CONCAT_WS(' ', a.first_name, a.last_name)), ', ' ORDER BY a.author_id), '')
	FROM books b
	LEFT JOIN categories c ON c.category_id = b.category_id
	LEFT JOIN book_authors ba ON ba.book_id = b.book_id
	LEFT JOIN authors a ON a.author_id = ba.author_id`

const bookGroupBy = ` GROUP BY b.book_id, c.category_name, c.language`

// queryRower is satisfied by both *sql.DB and *sql.Tx, so scanBook can be
// reused inside and outside transactions.
type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func scanBook(ctx context.Context, q queryRower, id int) (domain.Book, error) {
	row := q.QueryRowContext(ctx, bookSelect+` WHERE b.book_id = $1`+bookGroupBy, id)

	var (
		bookID    int
		title     string
		copies    int
		available int
		genre     string
		language  string
		author    string
	)
	err := row.Scan(&bookID, &title, &copies, &available, &genre, &language, &author)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Book{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Book{}, fmt.Errorf("postgres: scan book: %w", err)
	}
	return domain.Book{
		ID:        strconv.Itoa(bookID),
		Title:     title,
		Author:    author,
		Copies:    copies,
		Available: available,
		Genre:     genre,
		Language:  language,
	}, nil
}

func (r *BookRepository) FindByID(ctx context.Context, id string) (domain.Book, error) {
	bookID, err := strconv.Atoi(id)
	if err != nil {
		return domain.Book{}, domain.ErrNotFound
	}
	return scanBook(ctx, r.db, bookID)
}

func (r *BookRepository) FindAll(ctx context.Context) ([]domain.Book, error) {
	rows, err := r.db.QueryContext(ctx, bookSelect+bookGroupBy+` ORDER BY b.book_id`)
	if err != nil {
		return nil, fmt.Errorf("postgres: find all books: %w", err)
	}
	defer rows.Close()

	var books []domain.Book
	for rows.Next() {
		var (
			bookID    int
			title     string
			copies    int
			available int
			genre     string
			language  string
			author    string
		)
		if err := rows.Scan(&bookID, &title, &copies, &available, &genre, &language, &author); err != nil {
			return nil, fmt.Errorf("postgres: scan book: %w", err)
		}
		books = append(books, domain.Book{
			ID:        strconv.Itoa(bookID),
			Title:     title,
			Author:    author,
			Copies:    copies,
			Available: available,
			Genre:     genre,
			Language:  language,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: find all books: %w", err)
	}
	return books, nil
}

// DecrementAvailable and IncrementAvailable both need to check-then-write
// atomically (a plain UPDATE could race two concurrent borrows into
// negative availability), so each runs inside its own transaction with a
// row lock via SELECT ... FOR UPDATE.
func (r *BookRepository) DecrementAvailable(ctx context.Context, id string) (domain.Book, error) {
	return r.adjustAvailable(ctx, id, -1)
}

// IncrementAvailable must refuse to push Available above Copies — that
// would mean the repository is reporting copies of a book that don't
// physically exist.
func (r *BookRepository) IncrementAvailable(ctx context.Context, id string) (domain.Book, error) {
	return r.adjustAvailable(ctx, id, +1)
}

func (r *BookRepository) adjustAvailable(ctx context.Context, id string, delta int) (domain.Book, error) {
	bookID, err := strconv.Atoi(id)
	if err != nil {
		return domain.Book{}, domain.ErrNotFound
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Book{}, fmt.Errorf("postgres: begin tx: %w", err)
	}
	defer tx.Rollback()

	var copies, available int
	err = tx.QueryRowContext(ctx,
		`SELECT total_copies, available_copies FROM books WHERE book_id = $1 FOR UPDATE`, bookID,
	).Scan(&copies, &available)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Book{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Book{}, fmt.Errorf("postgres: lock book: %w", err)
	}

	switch {
	case delta < 0 && available <= 0:
		return domain.Book{}, domain.ErrNoCopiesAvailable
	case delta > 0 && available >= copies:
		return domain.Book{}, domain.ErrAvailableExceedsCopies
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE books SET available_copies = available_copies + $1 WHERE book_id = $2`, delta, bookID,
	); err != nil {
		return domain.Book{}, fmt.Errorf("postgres: update book availability: %w", err)
	}

	book, err := scanBook(ctx, tx, bookID)
	if err != nil {
		return domain.Book{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Book{}, fmt.Errorf("postgres: commit: %w", err)
	}
	return book, nil
}

// Create adds a brand new title to the catalog: it finds-or-creates the
// (genre, language) category row, inserts the book itself with
// available_copies == total_copies, inserts a single author row for the
// (unstructured) Author string, and links the two — all inside one
// transaction so a partial failure never leaves an orphaned category or
// author behind.
func (r *BookRepository) Create(ctx context.Context, book domain.Book) (domain.Book, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Book{}, fmt.Errorf("postgres: begin tx: %w", err)
	}
	defer tx.Rollback()

	var categoryID int
	err = tx.QueryRowContext(ctx,
		`SELECT category_id FROM categories WHERE category_name = $1 AND language = $2`,
		book.Genre, book.Language,
	).Scan(&categoryID)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx,
			`INSERT INTO categories (category_name, language) VALUES ($1, $2) RETURNING category_id`,
			book.Genre, book.Language,
		).Scan(&categoryID)
	}
	if err != nil {
		return domain.Book{}, fmt.Errorf("postgres: find or create category: %w", err)
	}

	var bookID int
	err = tx.QueryRowContext(ctx,
		`INSERT INTO books (title, category_id, total_copies, available_copies)
		 VALUES ($1, $2, $3, $3) RETURNING book_id`,
		book.Title, categoryID, book.Copies,
	).Scan(&bookID)
	if err != nil {
		return domain.Book{}, fmt.Errorf("postgres: insert book: %w", err)
	}

	var authorID int
	err = tx.QueryRowContext(ctx,
		`INSERT INTO authors (first_name, last_name) VALUES ($1, '') RETURNING author_id`,
		book.Author,
	).Scan(&authorID)
	if err != nil {
		return domain.Book{}, fmt.Errorf("postgres: insert author: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO book_authors (book_id, author_id) VALUES ($1, $2)`, bookID, authorID,
	); err != nil {
		return domain.Book{}, fmt.Errorf("postgres: link author: %w", err)
	}

	created, err := scanBook(ctx, tx, bookID)
	if err != nil {
		return domain.Book{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Book{}, fmt.Errorf("postgres: commit: %w", err)
	}
	return created, nil
}

// IncreaseCopies records that the library has acquired amount additional
// physical copies of an existing book: both total_copies and
// available_copies go up together.
func (r *BookRepository) IncreaseCopies(ctx context.Context, id string, amount int) (domain.Book, error) {
	if amount <= 0 {
		return domain.Book{}, domain.ErrInvalidBookInput
	}
	bookID, err := strconv.Atoi(id)
	if err != nil {
		return domain.Book{}, domain.ErrNotFound
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Book{}, fmt.Errorf("postgres: begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE books SET total_copies = total_copies + $1, available_copies = available_copies + $1 WHERE book_id = $2`,
		amount, bookID,
	)
	if err != nil {
		return domain.Book{}, fmt.Errorf("postgres: increase copies: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.Book{}, fmt.Errorf("postgres: increase copies: %w", err)
	}
	if n == 0 {
		return domain.Book{}, domain.ErrNotFound
	}

	book, err := scanBook(ctx, tx, bookID)
	if err != nil {
		return domain.Book{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Book{}, fmt.Errorf("postgres: commit: %w", err)
	}
	return book, nil
}
