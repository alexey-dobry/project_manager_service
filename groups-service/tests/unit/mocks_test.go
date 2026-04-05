package unit

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/student-pm/groups-service/internal/domain"
	"github.com/student-pm/groups-service/internal/usecase"
)

type MockRepo struct {
	mu      sync.Mutex
	groups  map[uuid.UUID]*domain.Group
	members map[membershipKey]*domain.Membership
}

type membershipKey struct {
	GroupID uuid.UUID
	UserID  uuid.UUID
}

func NewMockRepo() *MockRepo {
	return &MockRepo{
		groups:  map[uuid.UUID]*domain.Group{},
		members: map[membershipKey]*domain.Membership{},
	}
}

func (m *MockRepo) Create(_ context.Context, g *domain.Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ex := range m.groups {
		if ex.Name == g.Name {
			return domain.ErrGroupAlreadyExists
		}
	}
	cp := *g
	m.groups[g.ID] = &cp
	return nil
}

func (m *MockRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Group, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	g, ok := m.groups[id]
	if !ok {
		return nil, domain.ErrGroupNotFound
	}
	cp := *g
	return &cp, nil
}

func (m *MockRepo) List(_ context.Context, f usecase.ListFilter) ([]domain.Group, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Group, 0, len(m.groups))
	for _, g := range m.groups {
		if f.Faculty != "" && g.Faculty != f.Faculty {
			continue
		}
		if f.Course > 0 && g.Course != f.Course {
			continue
		}
		out = append(out, *g)
	}
	total := len(out)
	if f.Offset >= total {
		return []domain.Group{}, total, nil
	}
	end := f.Offset + f.Limit
	if end > total {
		end = total
	}
	return out[f.Offset:end], total, nil
}

func (m *MockRepo) Update(_ context.Context, g *domain.Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.groups[g.ID]; !ok {
		return domain.ErrGroupNotFound
	}
	cp := *g
	m.groups[g.ID] = &cp
	return nil
}

func (m *MockRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.groups[id]; !ok {
		return domain.ErrGroupNotFound
	}
	delete(m.groups, id)
	for k := range m.members {
		if k.GroupID == id {
			delete(m.members, k)
		}
	}
	return nil
}

func (m *MockRepo) AddMember(_ context.Context, mb *domain.Membership) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := membershipKey{GroupID: mb.GroupID, UserID: mb.UserID}
	if _, ok := m.members[k]; ok {
		return domain.ErrMemberAlreadyInGroup
	}
	cp := *mb
	m.members[k] = &cp
	return nil
}

func (m *MockRepo) RemoveMember(_ context.Context, groupID, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := membershipKey{GroupID: groupID, UserID: userID}
	if _, ok := m.members[k]; !ok {
		return domain.ErrMemberNotFound
	}
	delete(m.members, k)
	return nil
}

func (m *MockRepo) ListMembers(_ context.Context, groupID uuid.UUID) ([]domain.Membership, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Membership, 0)
	for k, v := range m.members {
		if k.GroupID == groupID {
			out = append(out, *v)
		}
	}
	return out, nil
}

func (m *MockRepo) GetMember(_ context.Context, groupID, userID uuid.UUID) (*domain.Membership, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.members[membershipKey{GroupID: groupID, UserID: userID}]
	if !ok {
		return nil, domain.ErrMemberNotFound
	}
	cp := *v
	return &cp, nil
}
