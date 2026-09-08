package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"library/domain"
)

// MemberRepository is a Postgres-backed implementation of ports.MemberRepository.
type MemberRepository struct {
	db *sql.DB
}

func NewMemberRepository(db *sql.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

const memberSelect = `
	SELECT u.user_id, u.first_name, u.last_name
	FROM users u
	JOIN roles rl ON rl.role_id = u.role_id
	WHERE rl.role_name = 'Student'`

func (r *MemberRepository) FindByID(ctx context.Context, id string) (domain.Member, error) {
	userID, err := strconv.Atoi(id)
	if err != nil {
		return domain.Member{}, domain.ErrNotFound
	}

	row := r.db.QueryRowContext(ctx, memberSelect+` AND u.user_id = $1`, userID)
	var (
		memberID  int
		firstName string
		lastName  string
	)
	err = row.Scan(&memberID, &firstName, &lastName)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Member{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Member{}, fmt.Errorf("postgres: find member by id: %w", err)
	}

	return domain.Member{ID: strconv.Itoa(memberID), Name: joinName(firstName, lastName)}, nil
}

func (r *MemberRepository) FindAll(ctx context.Context) ([]domain.Member, error) {
	rows, err := r.db.QueryContext(ctx, memberSelect+` ORDER BY u.user_id`)
	if err != nil {
		return nil, fmt.Errorf("postgres: find all members: %w", err)
	}
	defer rows.Close()

	var members []domain.Member
	for rows.Next() {
		var (
			memberID  int
			firstName string
			lastName  string
		)
		if err := rows.Scan(&memberID, &firstName, &lastName); err != nil {
			return nil, fmt.Errorf("postgres: scan member: %w", err)
		}
		members = append(members, domain.Member{ID: strconv.Itoa(memberID), Name: joinName(firstName, lastName)})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: find all members: %w", err)
	}
	return members, nil
}

func (r *MemberRepository) Delete(ctx context.Context, id string) error {
	userID, err := strconv.Atoi(id)
	if err != nil {
		return domain.ErrNotFound
	}

	const q = `
		DELETE FROM users
		WHERE user_id = $1
		AND role_id = (SELECT role_id FROM roles WHERE role_name = 'Student')`
	res, err := r.db.ExecContext(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("postgres: delete member: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: delete member: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func joinName(first, last string) string {
	return strings.TrimSpace(first + " " + last)
}

func (r *MemberRepository) Save(ctx context.Context, member domain.Member) error {
	return nil
}
