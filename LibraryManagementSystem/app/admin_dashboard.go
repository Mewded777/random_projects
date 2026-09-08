package app

import (
	"context"
	"library/domain"
	"library/ports"
)

type MemberStatus struct {
	Member        domain.Member
	ActiveLoans   int
	ReturnedCount int
}

// Map this to store domain.User
type LibrarianStatus struct {
	Admin domain.User
}

type AdminDashboardView struct {
	TotalBooks     int
	TotalMembers   int
	MemberStatuses []MemberStatus
	AdminStatuses  []LibrarianStatus // Maps to the html loop
}

type AdminDashboardHandler struct {
	Books   ports.BookRepository
	Members ports.MemberRepository
	Loans   ports.LoanRepository
	Users   ports.UserRepository
}

func NewAdminDashboardHandler(
	b ports.BookRepository,
	m ports.MemberRepository,
	l ports.LoanRepository,
	u ports.UserRepository, //  Update constructor
) *AdminDashboardHandler {
	return &AdminDashboardHandler{Books: b, Members: m, Loans: l, Users: u}
}

func (h *AdminDashboardHandler) Execute(ctx context.Context) (AdminDashboardView, error) {
	books, _ := h.Books.FindAll(ctx)
	members, _ := h.Members.FindAll(ctx)
	users, _ := h.Users.FindAllLibrarians(ctx) // To fetch all users from the repository

	view := AdminDashboardView{
		TotalBooks:   len(books),
		TotalMembers: len(members),
	}

	for _, member := range members {
		loans, _ := h.Loans.FindByMember(ctx, member.ID)
		active := 0
		returned := 0
		for _, loan := range loans {
			if loan.Returned {
				returned++
			} else {
				active++
			}
		}
		view.MemberStatuses = append(view.MemberStatuses, MemberStatus{
			Member:        member,
			ActiveLoans:   active,
			ReturnedCount: returned,
		})
	}

	// Loop through all users and filter out regular members
	for _, user := range users {
		// Only display accounts that are admins/librarians
		if user.Role != domain.RoleMember {
			view.AdminStatuses = append(view.AdminStatuses, LibrarianStatus{
				Admin: user,
			})
		}
	}

	return view, nil
}

// RemoveMember deletes a member's account.
func (h *AdminDashboardHandler) RemoveMember(ctx context.Context, memberID string) error {
	return h.Members.Delete(ctx, memberID)
}
