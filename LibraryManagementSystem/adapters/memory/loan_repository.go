package memory

import (
	"context"
	"library/domain"
	"sync"
)

type LoanRepository struct {
	mu    sync.Mutex
	loans map[string]*domain.Loan
}

func NewLoanRepository() *LoanRepository {
	return &LoanRepository{loans: map[string]*domain.Loan{}}
}

func (r *LoanRepository) Create(ctx context.Context, loan domain.Loan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	loanCopy := loan
	r.loans[loan.ID] = &loanCopy
	return nil
}

func (r *LoanRepository) Save(ctx context.Context, loan domain.Loan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.loans[loan.ID]; !ok {
		return domain.ErrNotFound
	}
	loanCopy := loan
	r.loans[loan.ID] = &loanCopy
	return nil
}

func (r *LoanRepository) FindByID(ctx context.Context, id string) (domain.Loan, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	loan, ok := r.loans[id]
	if !ok {
		return domain.Loan{}, domain.ErrNotFound
	}
	return *loan, nil
}

func (r *LoanRepository) FindByMember(ctx context.Context, memberID string) ([]domain.Loan, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	loans := make([]domain.Loan, 0)
	for _, loan := range r.loans {
		if loan.MemberID == memberID {
			loans = append(loans, *loan)
		}
	}
	return loans, nil
}

func (r *LoanRepository) CountActiveByMember(ctx context.Context, memberID string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, loan := range r.loans {
		if loan.MemberID == memberID && !loan.Returned {
			count++
		}
	}
	return count, nil
}
