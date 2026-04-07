package unit

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/student-pm/projects-service/internal/domain"
	"github.com/student-pm/projects-service/internal/usecase"
)

// ===== ProjectRepo =====

type MockProjectRepo struct {
	mu    sync.Mutex
	items map[uuid.UUID]*domain.Project
}

func NewMockProjectRepo() *MockProjectRepo {
	return &MockProjectRepo{items: map[uuid.UUID]*domain.Project{}}
}

func (m *MockProjectRepo) Create(_ context.Context, p *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *p
	m.items[p.ID] = &cp
	return nil
}

func (m *MockProjectRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.items[id]
	if !ok {
		return nil, domain.ErrProjectNotFound
	}
	cp := *p
	return &cp, nil
}

func (m *MockProjectRepo) List(_ context.Context, f usecase.ProjectListFilter) ([]domain.Project, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Project, 0, len(m.items))
	for _, p := range m.items {
		if f.GroupID != nil && p.GroupID != *f.GroupID {
			continue
		}
		if f.OwnerID != nil && p.OwnerID != *f.OwnerID {
			continue
		}
		if f.Status != nil && p.Status != *f.Status {
			continue
		}
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	total := len(out)
	if f.Offset >= total {
		return []domain.Project{}, total, nil
	}
	end := f.Offset + f.Limit
	if end > total {
		end = total
	}
	return out[f.Offset:end], total, nil
}

func (m *MockProjectRepo) Update(_ context.Context, p *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[p.ID]; !ok {
		return domain.ErrProjectNotFound
	}
	cp := *p
	m.items[p.ID] = &cp
	return nil
}

func (m *MockProjectRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return domain.ErrProjectNotFound
	}
	delete(m.items, id)
	return nil
}

// ===== TaskRepo =====

type MockTaskRepo struct {
	mu    sync.Mutex
	items map[uuid.UUID]*domain.Task
}

func NewMockTaskRepo() *MockTaskRepo {
	return &MockTaskRepo{items: map[uuid.UUID]*domain.Task{}}
}

func (m *MockTaskRepo) Create(_ context.Context, t *domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *t
	m.items[t.ID] = &cp
	return nil
}

func (m *MockTaskRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.items[id]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}
	cp := *t
	return &cp, nil
}

func (m *MockTaskRepo) ListByProject(_ context.Context, projectID uuid.UUID, f usecase.TaskListFilter) ([]domain.Task, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Task, 0)
	for _, t := range m.items {
		if t.ProjectID != projectID {
			continue
		}
		if f.Status != nil && t.Status != *f.Status {
			continue
		}
		if f.Priority != nil && t.Priority != *f.Priority {
			continue
		}
		if f.AssigneeID != nil {
			if t.AssigneeID == nil || *t.AssigneeID != *f.AssigneeID {
				continue
			}
		}
		out = append(out, *t)
	}
	total := len(out)
	if f.Offset >= total {
		return []domain.Task{}, total, nil
	}
	end := f.Offset + f.Limit
	if end > total {
		end = total
	}
	return out[f.Offset:end], total, nil
}

func (m *MockTaskRepo) Update(_ context.Context, t *domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[t.ID]; !ok {
		return domain.ErrTaskNotFound
	}
	cp := *t
	m.items[t.ID] = &cp
	return nil
}

func (m *MockTaskRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return domain.ErrTaskNotFound
	}
	delete(m.items, id)
	return nil
}

func (m *MockTaskRepo) Stats(_ context.Context, projectID uuid.UUID, now time.Time) (*domain.ProjectStats, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stats := &domain.ProjectStats{
		ProjectID:  projectID,
		ByStatus:   map[domain.TaskStatus]int{domain.TaskTodo: 0, domain.TaskInProgress: 0, domain.TaskDone: 0, domain.TaskBlocked: 0},
		ByPriority: map[domain.TaskPriority]int{domain.PriorityLow: 0, domain.PriorityMedium: 0, domain.PriorityHigh: 0, domain.PriorityCritical: 0},
	}
	for _, t := range m.items {
		if t.ProjectID != projectID {
			continue
		}
		stats.TotalTasks++
		stats.ByStatus[t.Status]++
		stats.ByPriority[t.Priority]++
		if t.DueDate != nil && t.DueDate.Before(now) && t.Status != domain.TaskDone {
			stats.OverdueCount++
		}
	}
	if stats.TotalTasks > 0 {
		stats.DonePercent = float64(stats.ByStatus[domain.TaskDone]) / float64(stats.TotalTasks) * 100.0
	}
	return stats, nil
}

// ===== CommentRepo =====

type MockCommentRepo struct {
	mu    sync.Mutex
	items map[uuid.UUID]*domain.Comment
}

func NewMockCommentRepo() *MockCommentRepo {
	return &MockCommentRepo{items: map[uuid.UUID]*domain.Comment{}}
}

func (m *MockCommentRepo) Create(_ context.Context, c *domain.Comment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *c
	m.items[c.ID] = &cp
	return nil
}

func (m *MockCommentRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Comment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.items[id]
	if !ok {
		return nil, domain.ErrCommentNotFound
	}
	cp := *c
	return &cp, nil
}

func (m *MockCommentRepo) ListByTask(_ context.Context, taskID uuid.UUID) ([]domain.Comment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Comment, 0)
	for _, c := range m.items {
		if c.TaskID == taskID {
			out = append(out, *c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (m *MockCommentRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return domain.ErrCommentNotFound
	}
	delete(m.items, id)
	return nil
}
