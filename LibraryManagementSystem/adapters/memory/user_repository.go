package memory

import (
	"context"
	"errors"
	"library/domain"
	"sync"
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[string]domain.User
}

func NewUserRepository() *UserRepository {
	repo := &UserRepository{
		users: make(map[string]domain.User),
	}

	// Hardcoded Admin user
	repo.users["SUPRADMN"] = domain.User{
		Username: "SUPRADMN",
		Password: "password",
		Role:     domain.RoleSuperAdmin,
	}

	return repo
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[username]
	if !exists {
		return domain.User{}, errors.New("user not found")
	}
	return user, nil
}

func (r *UserRepository) Save(ctx context.Context, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.Username] = user
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, username string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, username)
	return nil
}

func (r *UserRepository) FindAllLibrarians(ctx context.Context) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var librarians []domain.User
	for _, u := range r.users {
		if u.Role == domain.RoleLibrarain {
			librarians = append(librarians, u)
		}
	}
	return librarians, nil
}
