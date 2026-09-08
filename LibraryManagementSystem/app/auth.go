package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"library/domain"
	"library/ports"
)

type LoginCommand struct {
	Username string
	Password string
}

type RegisterCommand struct {
	Fname    string
	Lname    string
	Email    string
	Username string
	Password string
}

type AuthHandler struct {
	Users   ports.UserRepository
	Members ports.MemberRepository
}

type CreateLibrarianCommand struct {
	Username string
	Password string
}

func NewAuthHandler(users ports.UserRepository, members ports.MemberRepository) *AuthHandler {
	return &AuthHandler{Users: users, Members: members}
}

func (h *AuthHandler) HandleLogin(ctx context.Context, cmd LoginCommand) (domain.User, error) {
	user, err := h.Users.FindByUsername(ctx, cmd.Username)
	if err != nil || user.Password != cmd.Password {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	return user, nil
}

var memberIDCounter int64

// newMemberID produces a unique member ID for a freshly registered member.
func newMemberID(username string) string {
	n := atomic.AddInt64(&memberIDCounter, 1)
	return fmt.Sprintf("member-%s-%d-%d", username, time.Now().UnixNano(), n)
}

// HandleRegister creates a new member account. This does two things,
// it saves login credentials (a User), AND
// it creates the member record those credentials are linked to via TargetID
// without that second part, a newly registered member would have nothing to borrow against and every /borrow attempt would fail with "member not found".
func (h *AuthHandler) HandleRegister(ctx context.Context, cmd RegisterCommand) error {
	username := strings.TrimSpace(cmd.Username)
	if username == "" || cmd.Password == "" {
		return errors.New("username and password are required")
	}

	if _, err := h.Users.FindByUsername(ctx, username); err == nil {
		return errors.New("username already taken")
	}

	name := strings.TrimSpace(strings.TrimSpace(cmd.Fname) + " " + strings.TrimSpace(cmd.Lname))
	if name == "" {
		name = username
	}

	memberID := newMemberID(username)
	if err := h.Members.Save(ctx, domain.Member{ID: memberID, Name: name}); err != nil {
		return err
	}

	user := domain.User{
		Username: username,
		Password: cmd.Password,
		Role:     domain.RoleMember,
		TargetID: memberID,
	}
	return h.Users.Save(ctx, user)
}

func (h *AuthHandler) HandleCreatelibrarian(ctx context.Context, cmd CreateLibrarianCommand) error {
	admin := domain.User{
		Username: cmd.Username,
		Password: cmd.Password,
		Role:     domain.RoleLibrarain,
	}
	return h.Users.Save(ctx, admin)
}

func (h *AuthHandler) HandleRemoveAdmin(ctx context.Context, username string) error {
	return h.Users.Delete(ctx, username)
}
