package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"library/adapters/memory"
	"library/adapters/postgres"
	"library/app"
	"library/domain"
	"library/httpapi"
	"library/ports"
	"library/webui"
)

func main() {
	ctx := context.Background()

	var (
		bookRepo   ports.BookRepository
		memberRepo ports.MemberRepository
		loanRepo   ports.LoanRepository
		userRepo   ports.UserRepository
		db         *sql.DB
	)

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		var err error
		db, err = postgres.Open(ctx, dsn)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		log.Println("connected to Postgres database")

		bookRepo = postgres.NewBookRepository(db)
		memberRepo = postgres.NewMemberRepository(db)
		loanRepo = postgres.NewLoanRepository(db)
		userRepo = postgres.NewUserRepository(db)
	} else {
		log.Println("DATABASE_URL not set, using in-memory storage")

		memBookRepo := memory.NewBookRepository(map[string]*domain.Book{
			"book-1": {ID: "book-1", Title: "Domain-Driven Design", Author: "Eric Evans", Copies: 2, Available: 2, Language: domain.LanguageEnglish, Genre: "Drama"},
			"book-2": {ID: "book-2", Title: "Clean Architecture", Author: "Robert C. Martin", Copies: 1, Available: 1, Language: domain.LanguageEnglish, Genre: "Mystery"},
		})
		memMemberRepo := memory.NewMemberRepository(map[string]*domain.Member{
			"member-1": {ID: "member-1", Name: "Alice"},
			"member-2": {ID: "member-2", Name: "Bob"},
		})
		memUserRepo := memory.NewUserRepository() // Loads SUPRADMN automatically

		// Seed default login credentials for the members so they can log in.
		_ = memUserRepo.Save(ctx, domain.User{Username: "alice", Password: "password", Role: domain.RoleMember, TargetID: "member-1"})
		_ = memUserRepo.Save(ctx, domain.User{Username: "bob", Password: "password", Role: domain.RoleMember, TargetID: "member-2"})

		bookRepo = memBookRepo
		memberRepo = memMemberRepo
		loanRepo = memory.NewLoanRepository()
		userRepo = memUserRepo
	}

	//  Initialize Core Application Handlers
	borrowBook := app.NewBorrowBookHandler(bookRepo, memberRepo, loanRepo)
	returnBook := app.NewReturnBookHandler(bookRepo, loanRepo)
	getLoan := app.NewGetLoanHandler(loanRepo)
	getMemberLoans := app.NewGetMemberLoansHandler(loanRepo)
	listBooks := app.NewListBooksHandler(bookRepo)
	listMembers := app.NewListMembersHandler(memberRepo)

	// Pass the specialized user credential storage repository to the auth handler
	auth := app.NewAuthHandler(userRepo, memberRepo)
	adminDashboard := app.NewAdminDashboardHandler(bookRepo, memberRepo, loanRepo, userRepo)
	addBook := app.NewAddBookHandler(bookRepo)
	increaseCopies := app.NewIncreaseBookCopiesHandler(bookRepo)

	//  Initialize Driving Web Server Adapters
	web := webui.NewServer(listBooks, listMembers, getMemberLoans, getLoan, borrowBook, returnBook, auth, adminDashboard, addBook, increaseCopies)
	api := httpapi.NewServer(borrowBook, returnBook, getLoan, getMemberLoans)

	// JSON API on :8081, in its own goroutine, so it doesn't block the website from starting on :8080.
	go func() {
		log.Println("JSON API listening on :8081")
		if err := http.ListenAndServe(":8081", api.Routes()); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("website listening on :8080")

	// Automatically open the website interface in your browser targeting the login page
	openBrowser("http://localhost:8080/login")

	if err := http.ListenAndServe(":8080", web.Routes()); err != nil {
		log.Fatal(err)
	}
}

// openBrowser launches the system's default browser targeting the given URL.
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin": // macOS
		err = exec.Command("open", url).Start()
	default:
		log.Printf("unsupported platform, please open manually: %s", url)
		return
	}

	if err != nil {
		log.Printf("failed to automatically open browser: %v", err)
	}
}
