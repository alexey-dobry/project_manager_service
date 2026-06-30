package usecase_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/student-pm/projects-service/internal/domain"
	"github.com/student-pm/projects-service/internal/usecase"
)

// ===== минимально нужные mocks =====
// Полные mocks для остального service-API лежат в tests/unit. Здесь только
// то, что нужно для CreateComment.

type memProjects struct {
	mu sync.Mutex
	m  map[uuid.UUID]*domain.Project
}

func (r *memProjects) Create(_ context.Context, p *domain.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *p
	r.m[p.ID] = &cp
	return nil
}
func (r *memProjects) GetByID(_ context.Context, id uuid.UUID) (*domain.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.m[id]
	if !ok {
		return nil, domain.ErrProjectNotFound
	}
	cp := *p
	return &cp, nil
}
func (r *memProjects) List(_ context.Context, _ usecase.ProjectListFilter) ([]domain.Project, int, error) {
	return nil, 0, nil
}
func (r *memProjects) Update(_ context.Context, _ *domain.Project) error  { return nil }
func (r *memProjects) Delete(_ context.Context, _ uuid.UUID) error        { return nil }

type memTasks struct {
	mu sync.Mutex
	m  map[uuid.UUID]*domain.Task
}

func (r *memTasks) Create(_ context.Context, t *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *t
	r.m[t.ID] = &cp
	return nil
}
func (r *memTasks) GetByID(_ context.Context, id uuid.UUID) (*domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.m[id]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}
	cp := *t
	return &cp, nil
}
func (r *memTasks) ListByProject(_ context.Context, _ uuid.UUID, _ usecase.TaskListFilter) ([]domain.Task, int, error) {
	return nil, 0, nil
}
func (r *memTasks) Update(_ context.Context, _ *domain.Task) error                       { return nil }
func (r *memTasks) Delete(_ context.Context, _ uuid.UUID) error                          { return nil }
func (r *memTasks) Stats(_ context.Context, _ uuid.UUID, _ time.Time) (*domain.ProjectStats, error) {
	return &domain.ProjectStats{}, nil
}

type memComments struct {
	mu sync.Mutex
	m  map[uuid.UUID]*domain.Comment
}

func (r *memComments) Create(_ context.Context, c *domain.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *c
	r.m[c.ID] = &cp
	return nil
}
func (r *memComments) GetByID(_ context.Context, id uuid.UUID) (*domain.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.m[id]
	if !ok {
		return nil, domain.ErrCommentNotFound
	}
	cp := *c
	return &cp, nil
}
func (r *memComments) ListByTask(_ context.Context, _ uuid.UUID) ([]domain.Comment, error) {
	return nil, nil
}
func (r *memComments) Delete(_ context.Context, _ uuid.UUID) error { return nil }

// FuzzCreateCommentValidation — гоняет произвольный текст комментария и
// проверяет инварианты:
//   1. Пустой/whitespace-only текст → ErrEmptyContent.
//   2. Текст длиннее MaxCommentLength → ErrContentTooLong (фикс аномалии #4).
//   3. Любой допустимый текст → комментарий сохранён с TrimSpace.
//   4. Сервис не паникует.
//
// Запуск:
//   go test ./internal/usecase -run=- -fuzz=FuzzCreateCommentValidation -fuzztime=20s
func FuzzCreateCommentValidation(f *testing.F) {
	seeds := []string{
		"",
		"   ",
		"hi",
		"  trim me  ",
		strings.Repeat("a", domain.MaxCommentLength),       // ровно на границе
		strings.Repeat("a", domain.MaxCommentLength+1),     // сразу за границей
		// Из отчёта №4 — payload до 10 МБ. Полные 10 МБ как seed не кладём
		// (избыточно для fuzz-кэша); генерируем строкой в 11 KB,
		// чтобы покрыть условие len > MaxCommentLength.
		strings.Repeat("z", 11*1024),
		"\x00\x00\x00",
		"привет",
		"<script>alert(1)</script>",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	owner := usecase.Actor{ID: uuid.New(), Role: domain.RoleStudent}

	f.Fuzz(func(t *testing.T, content string) {
		// Готовим окружение: проект + задача внутри.
		projects := &memProjects{m: map[uuid.UUID]*domain.Project{}}
		tasks := &memTasks{m: map[uuid.UUID]*domain.Task{}}
		comments := &memComments{m: map[uuid.UUID]*domain.Comment{}}
		svc := usecase.NewService(projects, tasks, comments, time.Now)

		project, err := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
			Title: "P", GroupID: uuid.New(),
		})
		if err != nil {
			t.Fatalf("setup project: %v", err)
		}
		task, err := svc.CreateTask(context.Background(), owner, project.ID, usecase.CreateTaskInput{
			Title: "T", Priority: domain.PriorityMedium,
		})
		if err != nil {
			t.Fatalf("setup task: %v", err)
		}

		// Сам fuzz-вход.
		c, err := svc.CreateComment(context.Background(), owner, task.ID,
			usecase.CreateCommentInput{Content: content})

		// Инвариант: либо c != nil + err == nil, либо c == nil + err != nil.
		if (c == nil) == (err == nil) {
			t.Fatalf("invariant broken: c=%v err=%v", c, err)
		}

		trimmed := strings.TrimSpace(content)

		switch {
		case trimmed == "":
			if !errors.Is(err, domain.ErrEmptyContent) {
				t.Fatalf("empty content must return ErrEmptyContent, got %v", err)
			}
		case len(trimmed) > domain.MaxCommentLength:
			if !errors.Is(err, domain.ErrContentTooLong) {
				t.Fatalf("oversize content must return ErrContentTooLong, got %v (len=%d)",
					err, len(trimmed))
			}
		default:
			if err != nil {
				t.Fatalf("valid content rejected: %v (len=%d)", err, len(trimmed))
			}
			if c.Content != trimmed {
				t.Fatalf("Content not trimmed: stored=%q expected=%q", c.Content, trimmed)
			}
		}
	})
}
