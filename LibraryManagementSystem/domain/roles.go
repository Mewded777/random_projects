package domain

// Role defines the authorization clearance level of anyone logging in.
type Role string

const (
	RoleMember     Role = "member"
	RoleLibrarain  Role = "librarian"
	RoleSuperAdmin Role = "admin"
)
