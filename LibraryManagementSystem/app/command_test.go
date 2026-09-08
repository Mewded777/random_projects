package app

import (
	"context"
	"errors"
	"testing"

	"library/adapters/memory"
	"library/domain"

	"github.com/stretchr/testify/assert"
)

// newTestHandlers builds a BorrowBookHandler and ReturnBookHandler backed
// by fresh in-memory repositories, seeded the same way main() does.
// Returns the repos too, so tests can inspect state directly (e.g.
// checking Book.Available after a borrow).
func newTestHandlers() (*BorrowBookHandler, *ReturnBookHandler, *memory.BookRepository) {
	bookRepo := memory.NewBookRepository(map[string]*domain.Book{
		"book-1": {ID: "book-1", Title: "Domain-Driven Design", Author: "Eric Evans", Copies: 2, Available: 2},
		"book-2": {ID: "book-2", Title: "Clean Architecture", Author: "Robert C. Martin", Copies: 1, Available: 1},
	})
	memberRepo := memory.NewMemberRepository(map[string]*domain.Member{
		"member-1": {ID: "member-1", Name: "Alice"},
		"member-2": {ID: "member-2", Name: "Bob"},
	})
	loanRepo := memory.NewLoanRepository()

	borrow := NewBorrowBookHandler(bookRepo, memberRepo, loanRepo)
	ret := NewReturnBookHandler(bookRepo, loanRepo)
	return borrow, ret, bookRepo
}

func TestBorrowBookHandler_Success(t *testing.T) {
	borrow, _, books := newTestHandlers()
	ctx := context.Background()
	_ = borrow
	_ = books
	_ = ctx

	loanID, err := borrow.Handle(ctx, BorrowBookCommand{BookID: "book-1", MemberID: "member-1"}) // This should be fine
	assert.NoError(t, err)
	assert.NotEmpty(t, loanID)
	book, err := books.FindByID(ctx, "book-1") // This should also be fine and the Available should be decremented to 1
	assert.NoError(t, err)
	assert.Equal(t, 1, book.Available)
}

func TestBorrowBookHandler_BookNotFound(t *testing.T) {
	borrow, _, _ := newTestHandlers()
	ctx := context.Background()

	_, err := borrow.Handle(ctx, BorrowBookCommand{BookID: "does-not-exist", MemberID: "member-1"}) // This should give a ErrBookNotFound
	assert.True(t, errors.Is(err, domain.ErrBookNotFound))
	_ = err
}

func TestBorrowBookHandler_MemberNotFound(t *testing.T) {
	borrow, _, _ := newTestHandlers()
	ctx := context.Background()
	_, err := borrow.Handle(ctx, BorrowBookCommand{BookID: "book-1", MemberID: "does-not-exist"}) // This should give a ErrMemberNotFound
	assert.True(t, errors.Is(err, domain.ErrMemberNotFound))

}

func TestBorrowBookHandler_NoCopiesAvailable(t *testing.T) {
	borrow, _, _ := newTestHandlers()
	ctx := context.Background()
	_, err := borrow.Handle(ctx, BorrowBookCommand{BookID: "book-2", MemberID: "member-1"})
	assert.NoError(t, err)
	_, err = borrow.Handle(ctx, BorrowBookCommand{BookID: "book-2", MemberID: "member-2"})
	assert.True(t, errors.Is(err, domain.ErrNoCopiesAvailable))
}

func TestBorrowBookHandler_TooManyActiveLoans(t *testing.T) {
	// A member cannot borrow more than 3 books.
	bookRepo := memory.NewBookRepository(map[string]*domain.Book{
		"book-1": {ID: "book-1", Title: "Domain-Driven Design", Author: "Eric Evans", Copies: 2, Available: 2},
		"book-2": {ID: "book-2", Title: "Clean Architecture", Author: "Robert C. Martin", Copies: 1, Available: 1},
	})
	memberRepo := memory.NewMemberRepository(map[string]*domain.Member{
		"member-1": {ID: "member-1", Name: "Alice"},
		"member-2": {ID: "member-2", Name: "Bob"},
	})
	loanRepo := memory.NewLoanRepository()

	borrow := NewBorrowBookHandler(bookRepo, memberRepo, loanRepo)
	getMemberLoans := NewGetMemberLoansHandler(loanRepo)
	ctx := context.Background()

	if _, err := borrow.Handle(ctx, BorrowBookCommand{BookID: "book-1", MemberID: "member-1"}); err != nil {
		t.Fatalf("setup: unexpected error: %v", err)
	}
	if _, err := borrow.Handle(ctx, BorrowBookCommand{BookID: "book-1", MemberID: "member-1"}); err != nil {
		t.Fatalf("setup: unexpected error: %v", err)
	}

	if _, err := borrow.Handle(ctx, BorrowBookCommand{BookID: "book-2", MemberID: "member-1"}); err != nil {
		t.Fatalf("setup: unexpected error: %v", err)
	}

	loans, err := getMemberLoans.Handle(ctx, GetMemberLoansQuery{MemberID: "member-1"})
	assert.NoError(t, err)
	assert.Len(t, loans, 3)

	_, err = borrow.Handle(ctx, BorrowBookCommand{BookID: "book-1", MemberID: "member-1"})
	assert.ErrorIs(t, err, domain.ErrTooManyActiveLoans)
}

func TestReturnBookHandler_Success(t *testing.T) {
	borrow, ret, books := newTestHandlers()
	ctx := context.Background()

	loanID, err := borrow.Handle(ctx, BorrowBookCommand{BookID: "book-2", MemberID: "member-1"})
	if err != nil {
		t.Fatalf("setup: unexpected error borrowing: %v", err)
	}

	if err := ret.Handle(ctx, ReturnBookCommand{LoanID: loanID}); err != nil {
		t.Fatalf("setup: unexpected error returning: %v", err)
	}

	assert.NoError(t, err)
	book, err := books.FindByID(ctx, "book-2")
	assert.NoError(t, err)
	assert.Equal(t, 1, book.Available)
	_ = loanID
	_ = ret
	_ = books
}

func TestReturnBookHandler_TwiceFails(t *testing.T) {
	borrow, ret, _ := newTestHandlers()
	ctx := context.Background()

	loanID, err := borrow.Handle(ctx, BorrowBookCommand{BookID: "book-2", MemberID: "member-1"})
	if err != nil {
		t.Fatalf("setup: unexpected error borrowing: %v", err)
	}

	if err := ret.Handle(ctx, ReturnBookCommand{LoanID: loanID}); err != nil {
		t.Fatalf("setup: unexpected error returning: %v", err)
	}
	assert.NoError(t, err)
	err = ret.Handle(ctx, ReturnBookCommand{LoanID: loanID})
	assert.ErrorIs(t, err, domain.ErrLoanAlreadyReturned)
	_ = loanID
	_ = ret
}

func TestReturnBookHandler_NotFound(t *testing.T) {
	_, ret, _ := newTestHandlers()
	ctx := context.Background()

	err := ret.Handle(ctx, ReturnBookCommand{LoanID: "does-not-exist"})

	assert.ErrorIs(t, err, domain.ErrNotFound)
	_ = err
}

var _ = errors.Is
var _ = assert.Equal
