package memory

import (
	"context"
	"sync"

	"library/domain"
)

type MemberRepository struct {
	mu      sync.Mutex
	members map[string]*domain.Member
}

func NewMemberRepository(seed map[string]*domain.Member) *MemberRepository {
	if seed == nil {
		seed = map[string]*domain.Member{}
	}
	return &MemberRepository{members: seed}
}

func (r *MemberRepository) FindByID(ctx context.Context, id string) (domain.Member, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	member, ok := r.members[id]
	if !ok {
		return domain.Member{}, domain.ErrNotFound
	}
	return *member, nil
}

func (r *MemberRepository) FindAll(ctx context.Context) ([]domain.Member, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	members := make([]domain.Member, 0, len(r.members))
	for _, member := range r.members {
		members = append(members, *member)
	}
	return members, nil
}
func (r *MemberRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.members, id)
	return nil
}

// Save creates or updates a member record. Used at registration time so
// a brand-new member account has something to borrow against.
func (r *MemberRepository) Save(ctx context.Context, member domain.Member) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := member
	r.members[member.ID] = &m
	return nil
}
