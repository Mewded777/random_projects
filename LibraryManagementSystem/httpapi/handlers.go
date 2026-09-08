package httpapi

import (
	"encoding/json"
	"errors"
	"library/app"
	"library/domain"
	"net/http"
)

type Server struct {
	borrowBook     *app.BorrowBookHandler
	returnBook     *app.ReturnBookHandler
	getLoan        *app.GetLoanHandler
	getMemberLoans *app.GetMemberLoansHandler
}

func NewServer(
	borrowBook *app.BorrowBookHandler,
	returnBook *app.ReturnBookHandler,
	getLoan *app.GetLoanHandler,
	getMemberLoans *app.GetMemberLoansHandler,
) *Server {
	return &Server{
		borrowBook:     borrowBook,
		returnBook:     returnBook,
		getLoan:        getLoan,
		getMemberLoans: getMemberLoans,
	}
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /loans", s.borrowBookHandler)
	mux.HandleFunc("POST /loans/{id}/return", s.returnBookHandler)
	mux.HandleFunc("GET /members/{id}/loans", s.listMemberLoansHandler)
	return mux
}

type borrowRequest struct {
	BookID   string `json:"bookId"`
	MemberID string `json:"memberId"`
}

type loanResponse struct {
	ID       string `json:"id"`
	BookID   string `json:"bookId"`
	MemberID string `json:"memberId"`
	DueDate  string `json:"dueDate"`
	Returned bool   `json:"returned"`
}

func toLoanResponse(l domain.Loan) loanResponse {
	return loanResponse{ID: l.ID, BookID: l.BookID, MemberID: l.MemberID, DueDate: l.DueDate, Returned: l.Returned}
}

func (s *Server) borrowBookHandler(w http.ResponseWriter, r *http.Request) {
	var req borrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	loanID, err := s.borrowBook.Handle(r.Context(), app.BorrowBookCommand{BookID: req.BookID, MemberID: req.MemberID})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBookNotFound):
			http.Error(w, "book not found", http.StatusNotFound)
		case errors.Is(err, domain.ErrMemberNotFound):
			http.Error(w, "member not found", http.StatusNotFound)
		case errors.Is(err, domain.ErrNoCopiesAvailable):
			http.Error(w, "no copies available", http.StatusConflict)
		case errors.Is(err, domain.ErrTooManyActiveLoans):
			http.Error(w, "too many active loans", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	loan, err := s.getLoan.Handle(r.Context(), app.GetLoanQuery{LoanID: loanID})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toLoanResponse(loan))
}

func (s *Server) returnBookHandler(w http.ResponseWriter, r *http.Request) {
	loanID := r.PathValue("id")
	_ = loanID
	err := s.returnBook.Handle(r.Context(), app.ReturnBookCommand{LoanID: loanID})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			http.Error(w, "loan not found", http.StatusNotFound)
		case errors.Is(err, domain.ErrLoanAlreadyReturned):
			http.Error(w, "loan already returned", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	loan, err := s.getLoan.Handle(r.Context(), app.GetLoanQuery{LoanID: loanID})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toLoanResponse(loan))
}

func (s *Server) listMemberLoansHandler(w http.ResponseWriter, r *http.Request) {
	memberID := r.PathValue("id")

	loans, err := s.getMemberLoans.Handle(r.Context(), app.GetMemberLoansQuery{MemberID: memberID})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := make([]loanResponse, 0, len(loans))
	for _, l := range loans {
		resp = append(resp, toLoanResponse(l))
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
