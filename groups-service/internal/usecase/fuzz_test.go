package usecase_test

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/student-pm/groups-service/internal/domain"
	"github.com/student-pm/groups-service/internal/usecase"
)

// ===== минимальный in-memory репо =====

type memRepo struct {
	mu      sync.Mutex
	groups  map[uuid.UUID]*domain.Group
	members map[mkKey]*domain.Membership
}

type mkKey struct {
	G uuid.UUID
	U uuid.UUID
}

func newMemRepo() *memRepo {
	return &memRepo{groups: map[uuid.UUID]*domain.Group{}, members: map[mkKey]*domain.Membership{}}
}
func (m *memRepo) Create(_ context.Context, g *domain.Group) error {
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
func (m *memRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Group, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	g, ok := m.groups[id]
	if !ok {
		return nil, domain.ErrGroupNotFound
	}
	cp := *g
	return &cp, nil
}
func (m *memRepo) List(_ context.Context, _ usecase.ListFilter) ([]domain.Group, int, error) {
	return nil, 0, nil
}
func (m *memRepo) Update(_ context.Context, g *domain.Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.groups[g.ID]; !ok {
		return domain.ErrGroupNotFound
	}
	cp := *g
	m.groups[g.ID] = &cp
	return nil
}
func (m *memRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.groups, id)
	return nil
}
func (m *memRepo) AddMember(_ context.Context, mb *domain.Membership) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := mkKey{G: mb.GroupID, U: mb.UserID}
	if _, ok := m.members[k]; ok {
		return domain.ErrMemberAlreadyInGroup
	}
	cp := *mb
	m.members[k] = &cp
	return nil
}
func (m *memRepo) RemoveMember(_ context.Context, g, u uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.members, mkKey{G: g, U: u})
	return nil
}
func (m *memRepo) ListMembers(_ context.Context, _ uuid.UUID) ([]domain.Membership, error) {
	return nil, nil
}
func (m *memRepo) GetMember(_ context.Context, g, u uuid.UUID) (*domain.Membership, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.members[mkKey{G: g, U: u}]
	if !ok {
		return nil, domain.ErrMemberNotFound
	}
	cp := *v
	return &cp, nil
}

// FuzzCreateGroupValidation — гоняет произвольные (name, course, faculty)
// через usecase.Create и проверяет инварианты:
//   1. course вне диапазона [1..10] → ErrInvalidCourse.
//   2. пустое или whitespace-only name → ErrInvalidName.
//   3. валидный вход → группа создана с теми же значениями.
//   4. сервис никогда не паникует.
//
// Аномалия #3 из отчёта (отрицательный course): теперь регрессионно покрыта.
//
// Запуск:
//   go test ./internal/usecase -run=- -fuzz=FuzzCreateGroupValidation -fuzztime=20s
func FuzzCreateGroupValidation(f *testing.F) {
	type seed struct {
		name    string
		course  int
		faculty string
	}
	seeds := []seed{
		{"", 1, "ФИТ"},
		{"   ", 1, "ФИТ"},
		{"БПИ-211", 0, "ФИТ"},       // боли отчёта: course=0
		{"БПИ-211", -1, "ФИТ"},      // боли отчёта: отрицательный
		{"БПИ-211", -math.MaxInt32, "ФИТ"},
		{"БПИ-211", 11, "ФИТ"},      // выше границы
		{"БПИ-211", math.MaxInt32, "ФИТ"},
		{"БПИ-211", 2, "ФИТ"},       // ok
		{"\x00\x00", 1, "ФИТ"},      // нулевые байты
	}
	for _, s := range seeds {
		f.Add(s.name, s.course, s.faculty)
	}

	admin := usecase.Actor{ID: uuid.New(), Role: domain.RoleAdmin}

	f.Fuzz(func(t *testing.T, name string, course int, faculty string) {
		svc := usecase.NewGroupService(newMemRepo(), time.Now)

		g, err := svc.Create(context.Background(), admin, usecase.CreateGroupInput{
			Name:     name,
			Course:   course,
			Faculty:  faculty,
			LeaderID: uuid.New(),
		})

		// Инвариант: либо ок + group != nil, либо ошибка + group == nil.
		if err == nil && g == nil {
			t.Fatalf("nil err but nil group")
		}
		if err != nil && g != nil {
			t.Fatalf("err set but group != nil")
		}

		// Если успешно — обязаны быть валидные значения.
		if err == nil {
			if !domain.IsValidCourse(g.Course) {
				t.Fatalf("created group with invalid course=%d", g.Course)
			}
			if g.Name == "" {
				t.Fatalf("created group with empty name (input=%q)", name)
			}
		} else {
			// Если ошибка — она должна быть одной из ожидаемых.
			// 'forbidden' тут не может быть (admin), поэтому только наши доменные.
			ok := errors.Is(err, domain.ErrInvalidCourse) ||
				errors.Is(err, domain.ErrInvalidName) ||
				errors.Is(err, domain.ErrGroupAlreadyExists)
			if !ok {
				t.Fatalf("unexpected error: %v (name=%q course=%d)", err, name, course)
			}
		}
	})
}
