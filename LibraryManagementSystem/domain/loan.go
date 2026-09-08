package domain

import (
	"fmt"
	"time"
)

const MaxActiveLoans = 3

type Loan struct {
	ID       string
	BookID   string
	MemberID string
	DueDate  string
	Returned bool
}

func NewLoan(bookID, memberID string, currentActiveLoans int) (Loan, error) {
	if currentActiveLoans >= MaxActiveLoans {
		return Loan{}, ErrTooManyActiveLoans
	}
	return Loan{
		ID:       newLoanID(),
		BookID:   bookID,
		MemberID: memberID,
		DueDate:  time.Now().Add(14 * 24 * time.Hour).Format("2006-01-02"),
		Returned: false,
	}, nil
}

func (l *Loan) MarkAsReturned() error {
	if l.Returned {
		return ErrLoanAlreadyReturned
	}
	l.Returned = true
	return nil
}

func newLoanID() string {
	return fmt.Sprintf("loan-%d", time.Now().UnixNano())
}
