package httpdelivery

import "time"

// Project

type CreateProjectRequest struct {
	Title       string     `json:"title"        validate:"required,min=2,max=200"`
	Description string     `json:"description"  validate:"max=4000"`
	GroupID     string     `json:"group_id"     validate:"required,uuid"`
	Deadline    *time.Time `json:"deadline,omitempty"`
}

// UpdateProjectRequest содержит поля для частичного обновления.
// Флаги Clear* очищают соответствующее nullable-поле.
type UpdateProjectRequest struct {
	Title         *string    `json:"title,omitempty"        validate:"omitempty,min=2,max=200"`
	Description   *string    `json:"description,omitempty"  validate:"omitempty,max=4000"`
	Status        *string    `json:"status,omitempty"       validate:"omitempty,oneof=draft in_progress review completed archived"`
	Deadline      *time.Time `json:"deadline,omitempty"`
	ClearDeadline bool       `json:"clear_deadline,omitempty"`
}

type ProjectResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	GroupID     string     `json:"group_id"`
	OwnerID     string     `json:"owner_id"`
	Status      string     `json:"status"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PaginatedProjects struct {
	Items  []ProjectResponse `json:"items"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

// Task

type CreateTaskRequest struct {
	Title       string     `json:"title"       validate:"required,min=2,max=200"`
	Description string     `json:"description" validate:"max=4000"`
	AssigneeID  *string    `json:"assignee_id,omitempty" validate:"omitempty,uuid"`
	Priority    string     `json:"priority"    validate:"required,oneof=low medium high critical"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

type UpdateTaskRequest struct {
	Title         *string    `json:"title,omitempty"       validate:"omitempty,min=2,max=200"`
	Description   *string    `json:"description,omitempty" validate:"omitempty,max=4000"`
	AssigneeID    *string    `json:"assignee_id,omitempty" validate:"omitempty,uuid"`
	ClearAssignee bool       `json:"clear_assignee,omitempty"`
	Status        *string    `json:"status,omitempty"      validate:"omitempty,oneof=todo in_progress done blocked"`
	Priority      *string    `json:"priority,omitempty"    validate:"omitempty,oneof=low medium high critical"`
	DueDate       *time.Time `json:"due_date,omitempty"`
	ClearDueDate  bool       `json:"clear_due_date,omitempty"`
}

type TaskResponse struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PaginatedTasks struct {
	Items  []TaskResponse `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// Comment

type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=4000"`
}

type CommentResponse struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Stats

type StatsResponse struct {
	ProjectID    string         `json:"project_id"`
	TotalTasks   int            `json:"total_tasks"`
	ByStatus     map[string]int `json:"by_status"`
	ByPriority   map[string]int `json:"by_priority"`
	OverdueCount int            `json:"overdue_count"`
	DonePercent  float64        `json:"done_percent"`
}

type MessageResponse struct {
	Message string `json:"message" example:"ok"`
}
