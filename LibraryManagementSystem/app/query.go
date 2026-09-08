package app

import (
	"context"
	"strings"

	"library/domain"
	"library/ports"
)

// Queries never change state — that's the whole point of splitting them
// from commands. Each one only depends on the repository it actually
// reads from.

type GetLoanQuery struct {
	LoanID string
}

type GetLoanHandler struct {
	Loans ports.LoanRepository
}

func NewGetLoanHandler(loans ports.LoanRepository) *GetLoanHandler {
	return &GetLoanHandler{Loans: loans}
}

func (h *GetLoanHandler) Handle(ctx context.Context, q GetLoanQuery) (domain.Loan, error) {
	return h.Loans.FindByID(ctx, q.LoanID)
}

type GetMemberLoansQuery struct {
	MemberID string
}

type GetMemberLoansHandler struct {
	Loans ports.LoanRepository
}

func NewGetMemberLoansHandler(loans ports.LoanRepository) *GetMemberLoansHandler {
	return &GetMemberLoansHandler{Loans: loans}
}

func (h *GetMemberLoansHandler) Handle(ctx context.Context, q GetMemberLoansQuery) ([]domain.Loan, error) {
	return h.Loans.FindByMember(ctx, q.MemberID)
}

// ListBooksQuery lists the catalog, optionally narrowed down by any combination of Genre, Author, and Language.
type ListBooksQuery struct {
	Genre    string
	Author   string
	Language string
}

type ListBooksHandler struct {
	Books ports.BookRepository
}

func NewListBooksHandler(books ports.BookRepository) *ListBooksHandler {
	return &ListBooksHandler{Books: books}
}

func (h *ListBooksHandler) Handle(ctx context.Context, q ListBooksQuery) ([]domain.Book, error) {
	books, err := h.Books.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	if q.Genre == "" && q.Author == "" && q.Language == "" {
		return books, nil
	}

	filtered := make([]domain.Book, 0, len(books))
	for _, b := range books {
		if q.Genre != "" && b.Genre != q.Genre {
			continue
		}
		if q.Language != "" && b.Language != q.Language {
			continue
		}
		if q.Author != "" && !strings.Contains(strings.ToLower(b.Author), strings.ToLower(q.Author)) {
			continue
		}
		filtered = append(filtered, b)
	}
	return filtered, nil
}

type ListMembersQuery struct{}

type ListMembersHandler struct {
	Members ports.MemberRepository
}

func NewListMembersHandler(members ports.MemberRepository) *ListMembersHandler {
	return &ListMembersHandler{Members: members}
}

func (h *ListMembersHandler) Handle(ctx context.Context, q ListMembersQuery) ([]domain.Member, error) {
	return h.Members.FindAll(ctx)
}
