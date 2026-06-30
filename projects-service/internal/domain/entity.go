package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role — глобальная роль пользователя из JWT-claims.
type Role string

const (
	RoleStudent     Role = "student"
	RoleGroupLeader Role = "group_leader"
	RoleTeacher     Role = "teacher"
	RoleAdmin       Role = "admin"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleStudent, RoleGroupLeader, RoleTeacher, RoleAdmin:
		return true
	}
	return false
}

// ProjectStatus — статус проекта (жизненный цикл).
type ProjectStatus string

const (
	ProjectDraft      ProjectStatus = "draft"
	ProjectInProgress ProjectStatus = "in_progress"
	ProjectReview     ProjectStatus = "review"
	ProjectCompleted  ProjectStatus = "completed"
	ProjectArchived   ProjectStatus = "archived"
)

func (s ProjectStatus) IsValid() bool {
	switch s {
	case ProjectDraft, ProjectInProgress, ProjectReview, ProjectCompleted, ProjectArchived:
		return true
	}
	return false
}

// projectTransitions описывает допустимые переходы между статусами проекта.
var projectTransitions = map[ProjectStatus]map[ProjectStatus]bool{
	ProjectDraft:      {ProjectInProgress: true, ProjectArchived: true},
	ProjectInProgress: {ProjectReview: true, ProjectArchived: true},
	ProjectReview:     {ProjectInProgress: true, ProjectCompleted: true, ProjectArchived: true},
	ProjectCompleted:  {ProjectArchived: true},
	ProjectArchived:   {ProjectDraft: true}, // расконсервация
}

// CanTransitionTo — проверка, что переход разрешён.
func (s ProjectStatus) CanTransitionTo(next ProjectStatus) bool {
	if s == next {
		return true // идемпотентность
	}
	allowed, ok := projectTransitions[s]
	return ok && allowed[next]
}

// TaskStatus — статус задачи.
type TaskStatus string

const (
	TaskTodo       TaskStatus = "todo"
	TaskInProgress TaskStatus = "in_progress"
	TaskDone       TaskStatus = "done"
	TaskBlocked    TaskStatus = "blocked"
)

func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskTodo, TaskInProgress, TaskDone, TaskBlocked:
		return true
	}
	return false
}

// taskTransitions — допустимые переходы для задач.
// blocked доступен из любого нетерминального и возвращается в исходный (todo/in_progress).
var taskTransitions = map[TaskStatus]map[TaskStatus]bool{
	TaskTodo:       {TaskInProgress: true, TaskBlocked: true},
	TaskInProgress: {TaskTodo: true, TaskDone: true, TaskBlocked: true},
	TaskDone:       {TaskInProgress: true}, // reopen
	TaskBlocked:    {TaskTodo: true, TaskInProgress: true},
}

func (s TaskStatus) CanTransitionTo(next TaskStatus) bool {
	if s == next {
		return true
	}
	allowed, ok := taskTransitions[s]
	return ok && allowed[next]
}

// TaskPriority — приоритет.
type TaskPriority string

const (
	PriorityLow      TaskPriority = "low"
	PriorityMedium   TaskPriority = "medium"
	PriorityHigh     TaskPriority = "high"
	PriorityCritical TaskPriority = "critical"
)

func (p TaskPriority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		return true
	}
	return false
}

// ===== entities =====

// Project — проект студенческой группы.
type Project struct {
	ID          uuid.UUID     `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	GroupID     uuid.UUID     `json:"group_id"`  // логическая FK -> groups-service
	OwnerID     uuid.UUID     `json:"owner_id"`  // логическая FK -> auth-service.users
	Status      ProjectStatus `json:"status"`
	Deadline    *time.Time    `json:"deadline,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// Task — задача в проекте.
type Task struct {
	ID          uuid.UUID    `json:"id"`
	ProjectID   uuid.UUID    `json:"project_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	AssigneeID  *uuid.UUID   `json:"assignee_id,omitempty"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Comment — комментарий к задаче.
type Comment struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	UserID    uuid.UUID `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ProjectStats — агрегаты для /projects/:id/stats.
type ProjectStats struct {
	ProjectID    uuid.UUID            `json:"project_id"`
	TotalTasks   int                  `json:"total_tasks"`
	ByStatus     map[TaskStatus]int   `json:"by_status"`
	ByPriority   map[TaskPriority]int `json:"by_priority"`
	OverdueCount int                  `json:"overdue_count"`
	DonePercent  float64              `json:"done_percent"` // 0..100
}
