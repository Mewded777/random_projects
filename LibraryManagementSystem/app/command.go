package app

import (
	"context"
	"errors"
	"strings"

	"library/domain"
	"library/ports"
)

// Commands change state. Per CQRS, they return the minimum needed to
// confirm the change happened — not the full resulting object. If a
// caller needs the full state afterward, that's what the Query handlers
// in query.go are for.

type BorrowBookCommand struct {
	BookID   string
	MemberID string
}

type BorrowBookHandler struct {
	Books   ports.BookRepository
	Members ports.MemberRepository
	Loans   ports.LoanRepository
}

func NewBorrowBookHandler(books ports.BookRepository, members ports.MemberRepository, loans ports.LoanRepository) *BorrowBookHandler {
	return &BorrowBookHandler{Books: books, Members: members, Loans: loans}
}

func (h *BorrowBookHandler) Handle(ctx context.Context, cmd BorrowBookCommand) (string, error) {
	if _, err := h.Members.FindByID(ctx, cmd.MemberID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", domain.ErrMemberNotFound
		}
		return "", err
	}

	activeLoans, err := h.Loans.CountActiveByMember(ctx, cmd.MemberID)
	if err != nil {
		return "", err
	}

	loan, err := domain.NewLoan(cmd.BookID, cmd.MemberID, activeLoans)
	if err != nil {
		return "", err
	}

	if _, err := h.Books.DecrementAvailable(ctx, cmd.BookID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", domain.ErrBookNotFound
		}
		return "", err
	}

	if err := h.Loans.Create(ctx, loan); err != nil {
		return "", err
	}

	return loan.ID, nil
}

type ReturnBookCommand struct {
	LoanID string
}

type ReturnBookHandler struct {
	Books ports.BookRepository
	Loans ports.LoanRepository
}

func NewReturnBookHandler(books ports.BookRepository, loans ports.LoanRepository) *ReturnBookHandler {
	return &ReturnBookHandler{Books: books, Loans: loans}
}

// Handle returns only an error.Nothing else to confirm beyond "it worked or it didn't."
func (h *ReturnBookHandler) Handle(ctx context.Context, cmd ReturnBookCommand) error {

	loan, err := h.Loans.FindByID(ctx, cmd.LoanID)

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	if err := loan.MarkAsReturned(); err != nil {
		if errors.Is(err, domain.ErrLoanAlreadyReturned) {
			return domain.ErrLoanAlreadyReturned
		}
		return err
	}

	if err := h.Loans.Save(ctx, loan); err != nil {
		return err
	}

	if _, err := h.Books.IncrementAvailable(ctx, loan.BookID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrBookNotFound
		}
		return err
	}

	return nil
}

// AddBookCommand is issued by a librarian to add a brand new title to the
// catalog. Language and Genre must be one of the supported combinations
// defined in the domain package (the librarian picks language first,
// which determines the genre options offered).
type AddBookCommand struct {
	Title    string
	Author   string
	Language string
	Genre    string
	Copies   int
}

type AddBookHandler struct {
	Books ports.BookRepository
}

func NewAddBookHandler(books ports.BookRepository) *AddBookHandler {
	return &AddBookHandler{Books: books}
}

// Handle validates the librarian's input and creates the book.
func (h *AddBookHandler) Handle(ctx context.Context, cmd AddBookCommand) (domain.Book, error) {
	title := strings.TrimSpace(cmd.Title)
	author := strings.TrimSpace(cmd.Author)

	if title == "" || author == "" {
		return domain.Book{}, domain.ErrInvalidBookInput
	}
	if cmd.Copies <= 0 {
		return domain.Book{}, domain.ErrInvalidBookInput
	}
	if !domain.IsValidLanguage(cmd.Language) {
		return domain.Book{}, domain.ErrInvalidBookInput
	}
	if !domain.IsValidGenre(cmd.Language, cmd.Genre) {
		return domain.Book{}, domain.ErrInvalidBookInput
	}

	return h.Books.Create(ctx, domain.Book{
		Title:    title,
		Author:   author,
		Language: cmd.Language,
		Genre:    cmd.Genre,
		Copies:   cmd.Copies,
	})
}

// IncreaseBookCopiesCommand is issued by a librarian who has acquired more physical copies of a book that's already in the catalog.
type IncreaseBookCopiesCommand struct {
	BookID string
	Amount int
}

type IncreaseBookCopiesHandler struct {
	Books ports.BookRepository
}

func NewIncreaseBookCopiesHandler(books ports.BookRepository) *IncreaseBookCopiesHandler {
	return &IncreaseBookCopiesHandler{Books: books}
}

func (h *IncreaseBookCopiesHandler) Handle(ctx context.Context, cmd IncreaseBookCopiesCommand) (domain.Book, error) {
	if cmd.Amount <= 0 {
		return domain.Book{}, domain.ErrInvalidBookInput
	}

	book, err := h.Books.IncreaseCopies(ctx, cmd.BookID, cmd.Amount)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Book{}, domain.ErrBookNotFound
		}
		return domain.Book{}, err
	}
	return book, nil
}
