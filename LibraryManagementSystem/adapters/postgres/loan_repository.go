package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"library/domain"
)

// dueDateLayout matches the format domain.NewLoan uses when it stamps a
// loan's DueDate ("2006-01-02"), so it round-trips cleanly through
// Postgres's DATE type.
const dueDateLayout = "2006-01-02"

// LoanRepository is a Postgres-backed implementation of ports.LoanRepository.
type LoanRepository struct {
	db *sql.DB
}

func NewLoanRepository(db *sql.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

func (r *LoanRepository) Create(ctx context.Context, loan domain.Loan) error {
	memberID, err := strconv.Atoi(loan.MemberID)
	if err != nil {
		return fmt.Errorf("postgres: create loan: invalid member id %q: %w", loan.MemberID, err)
	}
	bookID, err := strconv.Atoi(loan.BookID)
	if err != nil {
		return fmt.Errorf("postgres: create loan: invalid book id %q: %w", loan.BookID, err)
	}
	dueDate, err := time.Parse(dueDateLayout, loan.DueDate)
	if err != nil {
		return fmt.Errorf("postgres: create loan: invalid due date %q: %w", loan.DueDate, err)
	}

	const q = `
		INSERT INTO loans (loan_id, user_id, book_id, borrow_date, due_date, return_date, status)
		VALUES ($1, $2, $3, CURRENT_DATE, $4, NULL, 'Borrowed')`
	if _, err := r.db.ExecContext(ctx, q, loan.ID, memberID, bookID, dueDate); err != nil {
		return fmt.Errorf("postgres: create loan: %w", err)
	}
	return nil
}

// Save persists a loan's mutable state (currently just Returned).
//
//	The borrow_date, book, and member never change after a loan is created, so this only ever touches status/return_date.
func (r *LoanRepository) Save(ctx context.Context, loan domain.Loan) error {
	status := "Borrowed"
	if loan.Returned {
		status = "Returned"
	}

	const q = `
		UPDATE loans
		SET status = $2::VARCHAR,
		    return_date = CASE WHEN $2::VARCHAR = 'Returned' THEN CURRENT_DATE ELSE NULL END
		WHERE loan_id = $1`
	res, err := r.db.ExecContext(ctx, q, loan.ID, status)
	if err != nil {
		return fmt.Errorf("postgres: save loan: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: save loan: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *LoanRepository) FindByID(ctx context.Context, id string) (domain.Loan, error) {
	const q = `SELECT loan_id, user_id, book_id, due_date, status FROM loans WHERE loan_id = $1`
	return scanLoan(r.db.QueryRowContext(ctx, q, id))
}

func (r *LoanRepository) FindByMember(ctx context.Context, memberID string) ([]domain.Loan, error) {
	id, err := strconv.Atoi(memberID)
	if err != nil {
		return []domain.Loan{}, nil
	}

	const q = `SELECT loan_id, user_id, book_id, due_date, status FROM loans WHERE user_id = $1 ORDER BY loan_id`
	rows, err := r.db.QueryContext(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("postgres: find loans by member: %w", err)
	}
	defer rows.Close()

	loans := make([]domain.Loan, 0)
	for rows.Next() {
		loan, err := scanLoanRow(rows)
		if err != nil {
			return nil, err
		}
		loans = append(loans, loan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: find loans by member: %w", err)
	}
	return loans, nil
}

func (r *LoanRepository) CountActiveByMember(ctx context.Context, memberID string) (int, error) {
	id, err := strconv.Atoi(memberID)
	if err != nil {
		return 0, nil
	}

	const q = `SELECT COUNT(*) FROM loans WHERE user_id = $1 AND status <> 'Returned'`
	var count int
	if err := r.db.QueryRowContext(ctx, q, id).Scan(&count); err != nil {
		return 0, fmt.Errorf("postgres: count active loans: %w", err)
	}
	return count, nil
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanLoan(row rowScanner) (domain.Loan, error) {
	loan, err := scanLoanRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Loan{}, domain.ErrNotFound
	}
	return loan, err
}

func scanLoanRow(row rowScanner) (domain.Loan, error) {
	var (
		loanID   string
		memberID int
		bookID   int
		dueDate  time.Time
		status   string
	)
	if err := row.Scan(&loanID, &memberID, &bookID, &dueDate, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Loan{}, err
		}
		return domain.Loan{}, fmt.Errorf("postgres: scan loan: %w", err)
	}
	return domain.Loan{
		ID:       loanID,
		BookID:   strconv.Itoa(bookID),
		MemberID: strconv.Itoa(memberID),
		DueDate:  dueDate.Format(dueDateLayout),
		Returned: status == "Returned",
	}, nil
}
