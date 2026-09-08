package memory

import (
	"context"
	"fmt"
	"library/domain"
	"sync"
	"sync/atomic"
	"time"
)

type BookRepository struct {
	mu    sync.Mutex
	books map[string]*domain.Book
}

func NewBookRepository(seed map[string]*domain.Book) *BookRepository {
	if seed == nil {
		seed = map[string]*domain.Book{}
	}
	return &BookRepository{books: seed}
}

var bookIDCounter int64

func newBookID() string {
	n := atomic.AddInt64(&bookIDCounter, 1)
	return fmt.Sprintf("book-%d-%d", time.Now().UnixNano(), n)
}

func (r *BookRepository) FindByID(ctx context.Context, id string) (domain.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	book, ok := r.books[id]
	if !ok {
		return domain.Book{}, domain.ErrNotFound
	}
	return *book, nil
}

func (r *BookRepository) FindAll(ctx context.Context) ([]domain.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	books := make([]domain.Book, 0, len(r.books))
	for _, book := range r.books {
		books = append(books, *book)
	}
	return books, nil
}

func (r *BookRepository) DecrementAvailable(ctx context.Context, id string) (domain.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	book, ok := r.books[id]
	if !ok {
		return domain.Book{}, domain.ErrNotFound
	}
	if book.Available <= 0 {
		return domain.Book{}, domain.ErrNoCopiesAvailable
	}
	book.Available--
	return *book, nil
}

// IncrementAvailable must refuse to push Available above Copies
func (r *BookRepository) IncrementAvailable(ctx context.Context, id string) (domain.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	book, ok := r.books[id]
	if !ok {
		return domain.Book{}, domain.ErrNotFound
	}
	if book.Available >= book.Copies {
		return domain.Book{}, domain.ErrAvailableExceedsCopies
	}
	book.Available++
	return *book, nil
}

// Create adds a brand new title to the catalog.
func (r *BookRepository) Create(ctx context.Context, book domain.Book) (domain.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	book.ID = newBookID()
	book.Available = book.Copies

	stored := book
	r.books[book.ID] = &stored
	return stored, nil
}

// IncreaseCopies records that the library has acquired amount additional physical copies of an existing book
// both Copies and Available go up together.
func (r *BookRepository) IncreaseCopies(ctx context.Context, id string, amount int) (domain.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	book, ok := r.books[id]
	if !ok {
		return domain.Book{}, domain.ErrNotFound
	}
	if amount <= 0 {
		return domain.Book{}, domain.ErrInvalidBookInput
	}

	book.Copies += amount
	book.Available += amount
	return *book, nil
}
