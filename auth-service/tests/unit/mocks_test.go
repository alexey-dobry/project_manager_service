package unit

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/student-pm/auth-service/internal/domain"
)

// ===== UserRepo mock =====

type MockUserRepo struct {
	mu       sync.Mutex
	byID     map[uuid.UUID]*domain.User
	byEmail  map[string]*domain.User
	CreateFn func(*domain.User) error
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		byID:    map[uuid.UUID]*domain.User{},
		byEmail: map[string]*domain.User{},
	}
}

func (m *MockUserRepo) Create(_ context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.CreateFn != nil {
		if err := m.CreateFn(u); err != nil {
			return err
		}
	}
	if _, ok := m.byEmail[u.Email]; ok {
		return domain.ErrUserAlreadyExists
	}
	cp := *u
	m.byID[u.ID] = &cp
	m.byEmail[u.Email] = &cp
	return nil
}

func (m *MockUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *MockUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byEmail[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *MockUserRepo) Update(_ context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byID[u.ID]; !ok {
		return domain.ErrUserNotFound
	}
	cp := *u
	m.byID[u.ID] = &cp
	m.byEmail[u.Email] = &cp
	return nil
}

// ===== RefreshTokenRepo mock =====

type MockTokenRepo struct {
	mu     sync.Mutex
	byHash map[string]*domain.RefreshToken
}

func NewMockTokenRepo() *MockTokenRepo {
	return &MockTokenRepo{byHash: map[string]*domain.RefreshToken{}}
}

func (m *MockTokenRepo) Save(_ context.Context, t *domain.RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *t
	m.byHash[t.TokenHash] = &cp
	return nil
}

func (m *MockTokenRepo) GetByHash(_ context.Context, hash string) (*domain.RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.byHash[hash]
	if !ok {
		return nil, domain.ErrInvalidToken
	}
	cp := *t
	return &cp, nil
}

func (m *MockTokenRepo) Revoke(_ context.Context, id uuid.UUID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.byHash {
		if t.ID == id && t.RevokedAt == nil {
			t.RevokedAt = &at
			return nil
		}
	}
	return nil
}

func (m *MockTokenRepo) RevokeAllForUser(_ context.Context, userID uuid.UUID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.byHash {
		if t.UserID == userID && t.RevokedAt == nil {
			t.RevokedAt = &at
		}
	}
	return nil
}

// Count — для проверок в тестах.
func (m *MockTokenRepo) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.byHash)
}
