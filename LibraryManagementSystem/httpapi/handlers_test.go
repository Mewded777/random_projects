package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"library/adapters/memory"
	"library/app"
	"library/domain"

	"github.com/stretchr/testify/assert"
)

type memberRepoAdapter struct {
	*memory.MemberRepository
}

func (m *memberRepoAdapter) Delete(ctx context.Context, id string) error {
	return nil
}

func newTestServer() *Server {
	bookRepo := memory.NewBookRepository(map[string]*domain.Book{
		"book-1": {ID: "book-1", Title: "Domain-Driven Design", Author: "Eric Evans", Copies: 2, Available: 2},
		"book-2": {ID: "book-2", Title: "Clean Architecture", Author: "Robert C. Martin", Copies: 1, Available: 1},
	})
	memberRepo := &memberRepoAdapter{MemberRepository: memory.NewMemberRepository(map[string]*domain.Member{
		"member-1": {ID: "member-1", Name: "Alice"},
		"member-2": {ID: "member-2", Name: "Bob"},
	})}
	loanRepo := memory.NewLoanRepository()

	borrowBook := app.NewBorrowBookHandler(bookRepo, memberRepo, loanRepo)
	returnBook := app.NewReturnBookHandler(bookRepo, loanRepo)
	getLoan := app.NewGetLoanHandler(loanRepo)
	getMemberLoans := app.NewGetMemberLoansHandler(loanRepo)

	return NewServer(borrowBook, returnBook, getLoan, getMemberLoans)
}

func TestBorrowBookHandler_MalformedBody(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/loans", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()
	s.borrowBookHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBorrowBookHandler_Success(t *testing.T) {
	s := newTestServer()

	body := []byte(`{"bookId":"book-1","memberId":"member-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/loans", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.borrowBookHandler(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp loanResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "book-1", resp.BookID)
	assert.Equal(t, "member-1", resp.MemberID)
}

func TestBorrowBookHandler_ErrorStatusCodes(t *testing.T) {
	cases := []struct {
		name       string
		bookID     string
		memberID   string
		wantStatus int
	}{
		{"book not found", "does-not-exist", "member-1", http.StatusNotFound},
		{"member not found", "book-1", "does-not-exist", http.StatusNotFound},
		{"no copies available", "book-2", "member-2", http.StatusConflict},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer() // fresh server per case.
			if tc.name == "no copies available" {
				setupBody := []byte(`{"bookId":"book-2","memberId":"member-1"}`)
				setupReq := httptest.NewRequest(http.MethodPost, "/loans", bytes.NewReader(setupBody))
				setupRec := httptest.NewRecorder()
				s.borrowBookHandler(setupRec, setupReq)
				if setupRec.Code != http.StatusCreated {
					t.Fatalf("setup: unexpected response code: %d", setupRec.Code)
				}
			}
			// special setup only for the "no copies" case .
			body := []byte(`{"bookId":"` + tc.bookID + `","memberId":"` + tc.memberID + `"}`)
			req := httptest.NewRequest(http.MethodPost, "/loans", bytes.NewReader(body))
			rec := httptest.NewRecorder()
			s.borrowBookHandler(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code, tc.name)
		})
	}
}

func TestReturnBookHandler_Success(t *testing.T) {
	s := newTestServer()

	borrowBody := []byte(`{"bookId":"book-2","memberId":"member-1"}`)
	borrowReq := httptest.NewRequest(http.MethodPost, "/loans", bytes.NewReader(borrowBody))
	borrowRec := httptest.NewRecorder()
	s.borrowBookHandler(borrowRec, borrowReq)

	var borrowResp loanResponse
	if err := json.NewDecoder(borrowRec.Body).Decode(&borrowResp); err != nil {
		t.Fatalf("setup: failed to decode borrow response: %v", err)
	}
	if borrowResp.ID == "" {
		t.Fatalf("setup: borrow response had empty loan ID")
	}

	req := httptest.NewRequest(http.MethodPost, "/loans/return", nil)
	req.SetPathValue("id", borrowResp.ID)
	rec := httptest.NewRecorder()
	s.returnBookHandler(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

var _ = assert.Equal
