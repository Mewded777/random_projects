package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUnauthorized       = errors.New("access denied: insufficient permissions")
)

// User handles security credentials for anyone accessing the application.
type User struct {
	Username string
	Password string
	Role     Role   // References the types defined in roles.go
	TargetID string // Links to Member.ID if Role == RoleMember. Empty for admins.
}
