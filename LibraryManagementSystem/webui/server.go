package webui

import (
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"library/app"
	"library/domain"
	"log"
	"net/http"
	"sort"
	"strconv"
)

//go:embed templates/*
var templateFS embed.FS

// Server acts as the user-facing web driver adapter.
type Server struct {
	homeTmpl          *template.Template
	memberLoansTmpl   *template.Template
	librarianDashTmpl *template.Template
	loginTmpl         *template.Template
	adminDashTmpl     *template.Template
	registerTmpl      *template.Template

	listBooks      *app.ListBooksHandler
	listMembers    *app.ListMembersHandler
	getMemberLoans *app.GetMemberLoansHandler
	getLoan        *app.GetLoanHandler
	borrowBook     *app.BorrowBookHandler
	returnBook     *app.ReturnBookHandler
	auth           *app.AuthHandler
	dash           *app.AdminDashboardHandler
	addBook        *app.AddBookHandler
	increaseCopies *app.IncreaseBookCopiesHandler
}

func NewServer(
	listBooks *app.ListBooksHandler,
	listMembers *app.ListMembersHandler,
	getMemberLoans *app.GetMemberLoansHandler,
	getLoan *app.GetLoanHandler,
	borrowBook *app.BorrowBookHandler,
	returnBook *app.ReturnBookHandler,
	auth *app.AuthHandler,
	dash *app.AdminDashboardHandler,
	addBook *app.AddBookHandler,
	increaseCopies *app.IncreaseBookCopiesHandler,
) *Server {
	homeTmpl := template.Must(template.ParseFS(templateFS, "templates/layout.html", "templates/home.html"))
	memberLoansTmpl := template.Must(template.ParseFS(templateFS, "templates/layout.html", "templates/member_loans.html"))
	librarianDashTmpl := template.Must(template.ParseFS(templateFS, "templates/layout.html", "templates/librarian_dashboard.html"))
	adminDashTmpl := template.Must(template.ParseFS(templateFS, "templates/layout.html", "templates/admin_dashboard.html"))
	loginTmpl := template.Must(template.ParseFS(templateFS, "templates/layout.html", "templates/login.html"))
	registerTmpl := template.Must(template.ParseFS(templateFS, "templates/layout.html", "templates/register.html"))

	return &Server{
		homeTmpl:          homeTmpl,
		memberLoansTmpl:   memberLoansTmpl,
		librarianDashTmpl: librarianDashTmpl,
		adminDashTmpl:     adminDashTmpl,
		loginTmpl:         loginTmpl,
		registerTmpl:      registerTmpl,
		listBooks:         listBooks,
		listMembers:       listMembers,
		getMemberLoans:    getMemberLoans,
		getLoan:           getLoan,
		borrowBook:        borrowBook,
		returnBook:        returnBook,
		auth:              auth,
		dash:              dash,
		addBook:           addBook,
		increaseCopies:    increaseCopies,
	}
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	// Public Auth Endpoints
	mux.HandleFunc("GET /login", s.loginPageHandler)
	mux.HandleFunc("POST /login", s.loginSubmitHandler)
	mux.HandleFunc("POST /logout", s.logoutHandler)
	mux.HandleFunc("POST /register", s.registerHandler)
	mux.HandleFunc("GET /register", s.registerPageHandler)

	// Member Role Controlled Endpoints
	mux.HandleFunc("GET /", RequireRole(domain.RoleMember, domain.RoleLibrarain, domain.RoleSuperAdmin)(s.homeHandler))
	mux.HandleFunc("POST /borrow", RequireRole(domain.RoleMember)(s.borrowHandler))
	mux.HandleFunc("GET /members/{id}/loans", RequireRole(domain.RoleMember, domain.RoleLibrarain)(s.memberLoansHandler))
	mux.HandleFunc("POST /loans/{id}/return", RequireRole(domain.RoleMember)(s.returnHandler))

	// Librarian & Admin Core Management Endpoints
	mux.HandleFunc("GET /admin/dashboard", RequireRole(domain.RoleLibrarain, domain.RoleSuperAdmin)(s.adminDashboardHandler))
	mux.HandleFunc("POST /librarian/books/add", RequireRole(domain.RoleLibrarain, domain.RoleSuperAdmin)(s.addBookHandler))
	mux.HandleFunc("POST /librarian/books/{id}/increase-copies", RequireRole(domain.RoleLibrarain, domain.RoleSuperAdmin)(s.increaseCopiesHandler))
	mux.HandleFunc("POST /admin/members/delete", RequireRole(domain.RoleLibrarain, domain.RoleSuperAdmin)(s.removeMemberHandler))

	// Admin Only Management Endpoints
	mux.HandleFunc("POST /super/admins/create", RequireRole(domain.RoleSuperAdmin)(s.createLibrarianHandler))
	mux.HandleFunc("POST /super/admins/delete", RequireRole(domain.RoleSuperAdmin)(s.removeAdminHandler))
	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
		cssFile, err := templateFS.ReadFile("templates/style.css")
		if err != nil {
			http.Error(w, "css asset not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/css")
		w.Write(cssFile)
	})

	return mux
}

// Auth Request Route Handlers

func (s *Server) loginPageHandler(w http.ResponseWriter, r *http.Request) {
	s.render(w, s.loginTmpl, map[string]any{"Flash": r.URL.Query().Get("error")})
}

func (s *Server) registerPageHandler(w http.ResponseWriter, r *http.Request) {
	s.render(w, s.registerTmpl, nil)
}

func (s *Server) loginSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/login?error=invalid_form", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := s.auth.HandleLogin(r.Context(), app.LoginCommand{Username: username, Password: password})
	if err != nil {
		http.Redirect(w, r, "/login?error=invalid_credentials", http.StatusSeeOther)
		return
	}

	SetUserSession(w, user.Username, string(user.Role), user.TargetID)

	if user.Role == domain.RoleLibrarain || user.Role == domain.RoleSuperAdmin {
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	ClearUserSession(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Catalog App Handlers

type homeViewData struct {
	Books           []domain.Book
	Members         []domain.Member
	CurrentUser     string
	CurrentRole     string
	CurrentMemberID string
	Flash           string

	Genres         []string
	Languages      []string
	SelectedGenre  string
	SelectedAuthor string
	SelectedLang   string
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	username, role, memberID := GetUserSession(r)

	genre := r.URL.Query().Get("genre")
	author := r.URL.Query().Get("author")
	language := r.URL.Query().Get("language")

	books, err := s.listBooks.Handle(r.Context(), app.ListBooksQuery{
		Genre:    genre,
		Author:   author,
		Language: language,
	})
	if err != nil {
		log.Printf("internal error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	members, err := s.listMembers.Handle(r.Context(), app.ListMembersQuery{})
	if err != nil {
		log.Printf("internal error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := homeViewData{
		Books:           books,
		Members:         members,
		CurrentUser:     username,
		CurrentRole:     role,
		CurrentMemberID: memberID,
		Flash:           r.URL.Query().Get("error"),
		Genres:          distinctGenres(),
		Languages:       domain.Languages(),
		SelectedGenre:   genre,
		SelectedAuthor:  author,
		SelectedLang:    language,
	}
	s.render(w, s.homeTmpl, data)
}

// distinctGenres flattens domain.GenresByLanguage into one deduplicated, sorted list for the filter dropdown.
func distinctGenres() []string {
	seen := map[string]bool{}
	var all []string
	for _, genres := range domain.GenresByLanguage {
		for _, g := range genres {
			if !seen[g] {
				seen[g] = true
				all = append(all, g)
			}
		}
	}
	sort.Strings(all)
	return all
}

func (s *Server) borrowHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	bookID := r.FormValue("bookId")
	memberID := r.FormValue("memberId")

	if _, err := s.borrowBook.Handle(r.Context(), app.BorrowBookCommand{BookID: bookID, MemberID: memberID}); err != nil {
		http.Redirect(w, r, "/?error="+errMessage(err), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/members/"+memberID+"/loans", http.StatusSeeOther)
}

type loanView struct {
	domain.Loan
	BookTitle string
}

type memberLoansViewData struct {
	Member          domain.Member
	Loans           []loanView
	Flash           string
	CurrentRole     string
	CurrentMemberID string
}

func (s *Server) memberLoansHandler(w http.ResponseWriter, r *http.Request) {
	memberID := r.PathValue("id")
	_, role, currentMemberID := GetUserSession(r)

	members, err := s.listMembers.Handle(r.Context(), app.ListMembersQuery{})
	if err != nil {
		log.Printf("internal error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var member domain.Member
	found := false
	for _, m := range members {
		if m.ID == memberID {
			member = m
			found = true
			break
		}
	}
	if !found {
		http.NotFound(w, r)
		return
	}

	loans, err := s.getMemberLoans.Handle(r.Context(), app.GetMemberLoansQuery{MemberID: memberID})
	if err != nil {
		log.Printf("internal error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	books, err := s.listBooks.Handle(r.Context(), app.ListBooksQuery{})
	if err != nil {
		log.Printf("internal error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	titleByID := make(map[string]string, len(books))
	for _, b := range books {
		titleByID[b.ID] = b.Title
	}

	views := make([]loanView, 0, len(loans))
	for _, l := range loans {
		views = append(views, loanView{Loan: l, BookTitle: titleByID[l.BookID]})
	}

	data := memberLoansViewData{
		Member:          member,
		Loans:           views,
		Flash:           r.URL.Query().Get("error"),
		CurrentRole:     role,
		CurrentMemberID: currentMemberID,
	}
	s.render(w, s.memberLoansTmpl, data)
}

func (s *Server) returnHandler(w http.ResponseWriter, r *http.Request) {
	loanID := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	_, _, sessionMemberID := GetUserSession(r)

	loan, err := s.getLoan.Handle(r.Context(), app.GetLoanQuery{LoanID: loanID})
	if err != nil {
		http.Redirect(w, r, "/members/"+sessionMemberID+"/loans?error="+errMessage(err), http.StatusSeeOther)
		return
	}
	if loan.MemberID != sessionMemberID {
		http.Redirect(w, r, "/members/"+sessionMemberID+"/loans?error=not_your_loan", http.StatusSeeOther)
		return
	}

	if err := s.returnBook.Handle(r.Context(), app.ReturnBookCommand{LoanID: loanID}); err != nil {
		http.Redirect(w, r, "/members/"+loan.MemberID+"/loans?error="+errMessage(err), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/members/"+loan.MemberID+"/loans", http.StatusSeeOther)
}

// Administrative Dashboard View Route Handlers

type adminDashViewData struct {
	Metrics     app.AdminDashboardView
	Books       []domain.Book
	Languages   []string
	GenresJSON  template.JS
	IsSuper     bool
	CurrentUser string
	Flash       string
}

func (s *Server) adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	username, role, _ := GetUserSession(r)

	metrics, err := s.dash.Execute(r.Context())
	if err != nil {
		http.Error(w, "internal configuration matrix evaluation mismatch", http.StatusInternalServerError)
		return
	}

	books, err := s.listBooks.Handle(r.Context(), app.ListBooksQuery{})
	if err != nil {
		log.Printf("internal error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	genresJSON, err := json.Marshal(domain.GenresByLanguage)
	if err != nil {
		log.Printf("internal error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := adminDashViewData{
		Metrics:     metrics,
		Books:       books,
		Languages:   domain.Languages(),
		GenresJSON:  template.JS(genresJSON),
		IsSuper:     role == string(domain.RoleSuperAdmin),
		CurrentUser: username,
		Flash:       r.URL.Query().Get("error"),
	}

	if role == string(domain.RoleSuperAdmin) {
		s.render(w, s.adminDashTmpl, data)
	} else {
		s.render(w, s.librarianDashTmpl, data)
	}
}

func (s *Server) addBookHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=invalid_form", http.StatusSeeOther)
		return
	}

	title := r.FormValue("title")
	author := r.FormValue("author")
	language := r.FormValue("language")
	genre := r.FormValue("genre")
	copies, _ := strconv.Atoi(r.FormValue("copies"))

	if _, err := s.addBook.Handle(r.Context(), app.AddBookCommand{
		Title:    title,
		Author:   author,
		Language: language,
		Genre:    genre,
		Copies:   copies,
	}); err != nil {
		http.Redirect(w, r, "/admin/dashboard?error="+errMessage(err), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard?success=book_added", http.StatusSeeOther)
}

func (s *Server) increaseCopiesHandler(w http.ResponseWriter, r *http.Request) {
	bookID := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=invalid_form", http.StatusSeeOther)
		return
	}

	amount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil || amount <= 0 {
		http.Redirect(w, r, "/admin/dashboard?error=invalid_book_details", http.StatusSeeOther)
		return
	}

	if _, err := s.increaseCopies.Handle(r.Context(), app.IncreaseBookCopiesCommand{BookID: bookID, Amount: amount}); err != nil {
		http.Redirect(w, r, "/admin/dashboard?error="+errMessage(err), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard?success=copies_increased", http.StatusSeeOther)
}

func (s *Server) createLibrarianHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=invalid_form", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if err := s.auth.HandleCreatelibrarian(r.Context(), app.CreateLibrarianCommand{Username: username, Password: password}); err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=could_not_create_librarian", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard?success=librarian_created", http.StatusSeeOther)
}

func (s *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/register?error=invalid_form", http.StatusSeeOther)
		return
	}

	fname := r.FormValue("fname")
	lname := r.FormValue("lname")
	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")

	if err := s.auth.HandleRegister(r.Context(), app.RegisterCommand{
		Fname:    fname,
		Lname:    lname,
		Email:    email,
		Username: username,
		Password: password,
	}); err != nil {
		http.Redirect(w, r, "/login?error=could_not_create_member", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/login?success=member_created", http.StatusSeeOther)
}
func (s *Server) removeMemberHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=invalid_form", http.StatusSeeOther)
		return
	}

	memberID := r.FormValue("memberId")
	if err := s.dash.RemoveMember(r.Context(), memberID); err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=could_not_remove_member", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard?success=member_removed", http.StatusSeeOther)
}

func (s *Server) removeAdminHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=invalid_form", http.StatusSeeOther)
		return
	}

	username := r.FormValue("adminId")
	if err := s.auth.HandleRemoveAdmin(r.Context(), username); err != nil {
		http.Redirect(w, r, "/admin/dashboard?error=could_not_remove_admin", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/dashboard?success=admin_removed", http.StatusSeeOther)
}

// Core Helper Functions

func (s *Server) render(w http.ResponseWriter, tmpl *template.Template, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Println("template rendering error:", err)
		http.Error(w, "internal content rendering exception", http.StatusInternalServerError)
	}
}

func errMessage(err error) string {
	switch {
	case errors.Is(err, domain.ErrBookNotFound):
		return "book not found"
	case errors.Is(err, domain.ErrMemberNotFound):
		return "member not found"
	case errors.Is(err, domain.ErrNoCopiesAvailable):
		return "no copies available"
	case errors.Is(err, domain.ErrTooManyActiveLoans):
		return "too many active loans"
	case errors.Is(err, domain.ErrLoanAlreadyReturned):
		return "loan already returned"
	case errors.Is(err, domain.ErrInvalidBookInput):
		return "invalid book details"
	default:
		return "something went wrong"
	}
}
