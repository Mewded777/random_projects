package ports

// This package defines the interfaces for the repositories that the application layer depends on.
import (
	"context"

	"library/domain"
)

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (domain.User, error)
	Save(ctx context.Context, user domain.User) error
	Delete(ctx context.Context, username string) error
	FindAllLibrarians(ctx context.Context) ([]domain.User, error)
}

type BookRepository interface {
	FindByID(ctx context.Context, id string) (domain.Book, error)
	FindAll(ctx context.Context) ([]domain.Book, error)
	DecrementAvailable(ctx context.Context, id string) (domain.Book, error)
	IncrementAvailable(ctx context.Context, id string) (domain.Book, error)
	Create(ctx context.Context, book domain.Book) (domain.Book, error)
	IncreaseCopies(ctx context.Context, id string, amount int) (domain.Book, error)
}

type MemberRepository interface {
	FindByID(ctx context.Context, id string) (domain.Member, error)
	FindAll(ctx context.Context) ([]domain.Member, error)
	Delete(ctx context.Context, id string) error
	Save(ctx context.Context, member domain.Member) error
}

type LoanRepository interface {
	Create(ctx context.Context, loan domain.Loan) error
	Save(ctx context.Context, loan domain.Loan) error
	FindByID(ctx context.Context, id string) (domain.Loan, error)
	FindByMember(ctx context.Context, memberID string) ([]domain.Loan, error)
	CountActiveByMember(ctx context.Context, memberID string) (int, error)
}
