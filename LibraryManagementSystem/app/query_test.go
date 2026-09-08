package app

import (
	"context"
	"errors"
	"library/adapters/memory"
	"library/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLoanHandler(t *testing.T) {
	bookRepo := memory.NewBookRepository(map[string]*domain.Book{
		"book-1": {ID: "book-1", Title: "Domain-Driven Design", Author: "Eric Evans", Copies: 2, Available: 2},
	})
	memberRepo := memory.NewMemberRepository(map[string]*domain.Member{
		"member-1": {ID: "member-1", Name: "Alice"},
	})
	loanRepo := memory.NewLoanRepository()

	borrow := NewBorrowBookHandler(bookRepo, memberRepo, loanRepo)
	getLoan := NewGetLoanHandler(loanRepo)
	ctx := context.Background()

	loanID, err := borrow.Handle(ctx, BorrowBookCommand{BookID: "book-1", MemberID: "member-1"})
	if err != nil {
		t.Fatalf("setup: unexpected error borrowing: %v", err)
	}

	loan, err := getLoan.Handle(ctx, GetLoanQuery{LoanID: loanID})
	assert.NoError(t, err)
	assert.Equal(t, "book-1", loan.BookID)
	assert.Equal(t, "member-1", loan.MemberID)
	_ = loanID
	_ = getLoan
}

func TestGetLoanHandler_NotFound(t *testing.T) {
	loanRepo := memory.NewLoanRepository()
	getLoan := NewGetLoanHandler(loanRepo)
	ctx := context.Background()

	_, err := getLoan.Handle(ctx, GetLoanQuery{LoanID: "does-not-exist"})

	assert.True(t, errors.Is(err, domain.ErrNotFound))
	_ = err
}

func TestGetMemberLoansHandler(t *testing.T) {
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
	if _, err := borrow.Handle(ctx, BorrowBookCommand{BookID: "book-2", MemberID: "member-1"}); err != nil {
		t.Fatalf("setup: unexpected error: %v", err)
	}

	loans, err := getMemberLoans.Handle(ctx, GetMemberLoansQuery{MemberID: "member-1"})
	assert.NoError(t, err)
	assert.Len(t, loans, 2)
	loans, err = getMemberLoans.Handle(ctx, GetMemberLoansQuery{MemberID: "member-2"})
	assert.NoError(t, err)
	assert.Len(t, loans, 0)
	_ = getMemberLoans
}

var _ = assert.Equal
