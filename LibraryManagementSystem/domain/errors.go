package domain

import "errors"

var (
	ErrNotFound               = errors.New("not found")
	ErrBookNotFound           = errors.New("book not found")
	ErrMemberNotFound         = errors.New("member not found")
	ErrNoCopiesAvailable      = errors.New("no copies available")
	ErrLoanAlreadyReturned    = errors.New("loan already returned")
	ErrTooManyActiveLoans     = errors.New("member has too many active loans")
	ErrAvailableExceedsCopies = errors.New("available copies cannot exceed total copies")
	ErrInvalidBookInput       = errors.New("invalid book input")
)
