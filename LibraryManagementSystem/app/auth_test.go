package app

import (
	"context"
	"errors"
	"testing"

	"library/adapters/memory"
	"library/domain"

	"github.com/stretchr/testify/assert"
)

func TestHandleRegister_CreatesAUsableMember(t *testing.T) {
	userRepo := memory.NewUserRepository()
	memberRepo := memory.NewMemberRepository(nil)
	bookRepo := memory.NewBookRepository(map[string]*domain.Book{
		"book-1": {ID: "book-1", Title: "Test Book", Copies: 1, Available: 1},
	})
	loanRepo := memory.NewLoanRepository()

	auth := NewAuthHandler(userRepo, memberRepo)
	borrow := NewBorrowBookHandler(bookRepo, memberRepo, loanRepo)
	ctx := context.Background()

	err := auth.HandleRegister(ctx, RegisterCommand{
		Fname:    "Ada",
		Lname:    "Lovelace",
		Username: "ada",
		Password: "secret",
	})
	assert.NoError(t, err)

	user, err := userRepo.FindByUsername(ctx, "ada")
	assert.NoError(t, err)
	assert.Equal(t, domain.RoleMember, user.Role)
	assert.NotEmpty(t, user.TargetID)

	// The whole point: a member created via registration must actually
	// be able to borrow, not just exist as a dangling username.
	_, err = borrow.Handle(ctx, BorrowBookCommand{BookID: "book-1", MemberID: user.TargetID})
	assert.NoError(t, err)
}

func TestHandleRegister_RejectsDuplicateUsername(t *testing.T) {
	userRepo := memory.NewUserRepository()
	memberRepo := memory.NewMemberRepository(nil)
	auth := NewAuthHandler(userRepo, memberRepo)
	ctx := context.Background()

	assert.NoError(t, auth.HandleRegister(ctx, RegisterCommand{Username: "ada", Password: "secret"}))
	err := auth.HandleRegister(ctx, RegisterCommand{Username: "ada", Password: "other"})
	assert.Error(t, err)
}

func TestHandleRegister_RejectsEmptyUsernameOrPassword(t *testing.T) {
	userRepo := memory.NewUserRepository()
	memberRepo := memory.NewMemberRepository(nil)
	auth := NewAuthHandler(userRepo, memberRepo)
	ctx := context.Background()

	assert.Error(t, auth.HandleRegister(ctx, RegisterCommand{Username: "", Password: "secret"}))
	assert.Error(t, auth.HandleRegister(ctx, RegisterCommand{Username: "ada", Password: ""}))
}

func TestBorrowBookHandler_MemberCannotExceedThreeActiveLoans(t *testing.T) {
	bookRepo := memory.NewBookRepository(map[string]*domain.Book{
		"book-1": {ID: "book-1", Title: "Book 1", Copies: 1, Available: 1},
		"book-2": {ID: "book-2", Title: "Book 2", Copies: 1, Available: 1},
		"book-3": {ID: "book-3", Title: "Book 3", Copies: 1, Available: 1},
		"book-4": {ID: "book-4", Title: "Book 4", Copies: 1, Available: 1},
	})
	memberRepo := memory.NewMemberRepository(map[string]*domain.Member{
		"member-1": {ID: "member-1", Name: "Alice"},
	})
	loanRepo := memory.NewLoanRepository()
	borrow := NewBorrowBookHandler(bookRepo, memberRepo, loanRepo)
	ctx := context.Background()

	for _, id := range []string{"book-1", "book-2", "book-3"} {
		_, err := borrow.Handle(ctx, BorrowBookCommand{BookID: id, MemberID: "member-1"})
		assert.NoError(t, err)
	}

	_, err := borrow.Handle(ctx, BorrowBookCommand{BookID: "book-4", MemberID: "member-1"})
	assert.Error(t, err)

	// The 4th book must still be untouched.
	// Available shouldn't have been decremented for a borrow that was rejected.
	book, err2 := bookRepo.FindByID(ctx, "book-4")
	assert.NoError(t, err2)
	assert.Equal(t, 1, book.Available)
	_ = errors.New
}
