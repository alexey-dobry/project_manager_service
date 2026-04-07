package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/student-pm/projects-service/internal/domain"
)

// ProjectRepository — порт хранения проектов.
type ProjectRepository interface {
	Create(ctx context.Context, p *domain.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	List(ctx context.Context, f ProjectListFilter) ([]domain.Project, int, error)
	Update(ctx context.Context, p *domain.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TaskRepository — порт хранения задач.
type TaskRepository interface {
	Create(ctx context.Context, t *domain.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, f TaskListFilter) ([]domain.Task, int, error)
	Update(ctx context.Context, t *domain.Task) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Stats — агрегации для /projects/:id/stats. Один SQL вместо ListByProject + подсчёт в Go.
	Stats(ctx context.Context, projectID uuid.UUID, now time.Time) (*domain.ProjectStats, error)
}

// CommentRepository — порт хранения комментариев.
type CommentRepository interface {
	Create(ctx context.Context, c *domain.Comment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	ListByTask(ctx context.Context, taskID uuid.UUID) ([]domain.Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProjectListFilter — фильтр и пагинация для списка проектов.
type ProjectListFilter struct {
	GroupID *uuid.UUID
	OwnerID *uuid.UUID
	Status  *domain.ProjectStatus
	Limit   int
	Offset  int
}

// TaskListFilter — фильтр для задач внутри проекта.
type TaskListFilter struct {
	Status     *domain.TaskStatus
	Priority   *domain.TaskPriority
	AssigneeID *uuid.UUID
	Limit      int
	Offset     int
}
