// Package postgres contains database-backed implementations of the repository interfaces defined in the ports package.
// They talk to the schema in database/library_management.sql over database/sql, using the vendored github.com/lib/pq driver.
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"library/domain"
)

// Open connects to Postgres using a standard connection string / DSN,
// e.g. "postgres://user:password@localhost:5432/library_management?sslmode=disable".
// It pings the database once so callers find out immediately if the
// connection is unusable, rather than on the first query.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return db, nil
}

// roleName maps the application's domain.Role onto the role_name values
// seeded into the roles table (see database/library_management.sql):
//
//	"Student" = domain.RoleMember = a library member, can borrow books
//	"Librarian" = domain.RoleLibrarian = day-to-day staff/admin dashboard
//	"Admin" = domain.RoleSuperAdmin = can create/remove other admins
//

func roleName(r domain.Role) (string, error) {
	switch r {
	case domain.RoleMember:
		return "Student", nil
	case domain.RoleLibrarain:
		return "Librarian", nil
	case domain.RoleSuperAdmin:
		return "Admin", nil
	default:
		return "", fmt.Errorf("postgres: unknown domain role %q", r)
	}
}

// domainRole is the inverse of roleName.
func domainRole(name string) (domain.Role, error) {
	switch name {
	case "Student":
		return domain.RoleMember, nil
	case "Librarian":
		return domain.RoleLibrarain, nil
	case "Admin":
		return domain.RoleSuperAdmin, nil
	default:
		return "", fmt.Errorf("postgres: unknown role_name %q", name)
	}
}
