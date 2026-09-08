package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"library/domain"
	"strconv"
	"strings"
)

// UserRepository is a Postgres-backed implementation of ports.UserRepository.
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	const q = `
		SELECT u.user_id, u.username, u.password_hash, rl.role_name
		FROM users u
		JOIN roles rl ON rl.role_id = u.role_id
		WHERE u.username = $1`

	var (
		userID   int
		uname    string
		password string
		roleStr  string
	)
	err := r.db.QueryRowContext(ctx, q, username).Scan(&userID, &uname, &password, &roleStr)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, errors.New("user not found")
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("postgres: find user by username: %w", err)
	}

	role, err := domainRole(roleStr)
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{Username: uname, Password: password, Role: role}
	// A Student's TargetID points at their own row: they *are* the
	// member. Staff (Librarian/Admin) accounts have no linked member.
	if role == domain.RoleMember {
		user.TargetID = strconv.Itoa(userID)
	}
	return user, nil
}

func (r *UserRepository) Save(ctx context.Context, user domain.User) error {
	roleStr, err := roleName(user.Role)
	if err != nil {
		return err
	}

	email := strings.ToLower(user.Username) + "@library.local"

	const q = `
		INSERT INTO users (username, first_name, last_name, email, password_hash, role_id)
		VALUES ($1, $2, '', $3, $4, (SELECT role_id FROM roles WHERE role_name = $5))
		ON CONFLICT (username) DO UPDATE
		SET password_hash = EXCLUDED.password_hash,
		    role_id = EXCLUDED.role_id`

	_, err = r.db.ExecContext(ctx, q, user.Username, user.Username, email, user.Password, roleStr)
	if err != nil {
		return fmt.Errorf("postgres: save user: %w", err)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, username string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE username = $1`, username)
	if err != nil {
		return fmt.Errorf("postgres: delete user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindAllLibrarians(ctx context.Context) ([]domain.User, error) {
	const q = `
		SELECT u.username, u.password_hash
		FROM users u
		JOIN roles rl ON rl.role_id = u.role_id
		WHERE rl.role_name = 'Librarian'`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("postgres: find all admins: %w", err)
	}
	defer rows.Close()

	var admins []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.Username, &u.Password); err != nil {
			return nil, fmt.Errorf("postgres: scan admin: %w", err)
		}
		u.Role = domain.RoleLibrarain
		admins = append(admins, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: find all admins: %w", err)
	}
	return admins, nil
}
