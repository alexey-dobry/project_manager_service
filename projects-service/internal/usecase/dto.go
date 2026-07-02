package usecase

import (
	"time"

	"github.com/google/uuid"

	"github.com/student-pm/projects-service/internal/domain"
)

// Actor — кто инициировал действие. Заполняется handler-ом из JWT.
type Actor struct {
	ID   uuid.UUID
	Role domain.Role
}

// ===== projects =====

type CreateProjectInput struct {
	Title       string
	Description string
	GroupID     uuid.UUID
	Deadline    *time.Time
}

type UpdateProjectInput struct {
	Title       *string
	Description *string
	Deadline    *time.Time
	ClearDeadline bool                  // отдельный флаг — JSON null vs not-set
	Status      *domain.ProjectStatus
}

// ===== tasks =====

type CreateTaskInput struct {
	Title       string
	Description string
	AssigneeID  *uuid.UUID
	Priority    domain.TaskPriority
	DueDate     *time.Time
}

type UpdateTaskInput struct {
	Title         *string
	Description   *string
	AssigneeID    *uuid.UUID
	ClearAssignee bool
	Priority      *domain.TaskPriority
	Status        *domain.TaskStatus
	Order         *int
	DueDate       *time.Time
	ClearDueDate  bool
}

// ===== comments =====

type CreateCommentInput struct {
	Content string
}
