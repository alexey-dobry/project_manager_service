package httpdelivery

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// FlexDate — нестрогая обёртка над датой во входящих запросах.
// Клиенты присылают дедлайны по-разному: полный RFC3339, дату без времени
// (как отдаёт HTML <input type="date">) или пустую строку/null при
// отсутствии значения. Строгий time.Time принимает только RFC3339 и на
// остальном падает прямо в BodyParser, возвращая неинформативный
// "invalid JSON body" — этот тип принимает все три варианта.
type FlexDate struct {
	Time  time.Time
	Valid bool
}

func (d *FlexDate) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		d.Valid = false
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		d.Time, d.Valid = t, true
		return nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		d.Time, d.Valid = t, true
		return nil
	}
	return fmt.Errorf("invalid date %q: expected RFC3339 or YYYY-MM-DD", s)
}

func (d FlexDate) MarshalJSON() ([]byte, error) {
	if !d.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time)
}

// Ptr конвертирует в *time.Time, ожидаемый usecase-слоем: nil, если значение
// не было передано.
func (d FlexDate) Ptr() *time.Time {
	if !d.Valid {
		return nil
	}
	t := d.Time
	return &t
}

// Project

type CreateProjectRequest struct {
	Title       string   `json:"title"        validate:"required,min=2,max=200"`
	Description string   `json:"description"  validate:"max=4000"`
	GroupID     string   `json:"group_id"     validate:"required,uuid"`
	Deadline    FlexDate `json:"deadline,omitempty"`
}

// UnmarshalJSON поддерживает "name" как альтернативное имя поля title —
// некоторые клиенты называют его так.
func (r *CreateProjectRequest) UnmarshalJSON(data []byte) error {
	type alias CreateProjectRequest
	aux := &struct {
		Name *string `json:"name"`
		*alias
	}{alias: (*alias)(r)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if r.Title == "" && aux.Name != nil {
		r.Title = *aux.Name
	}
	return nil
}

// UpdateProjectRequest содержит поля для частичного обновления.
// Флаги Clear* очищают соответствующее nullable-поле.
type UpdateProjectRequest struct {
	Title         *string  `json:"title,omitempty"        validate:"omitempty,min=2,max=200"`
	Description   *string  `json:"description,omitempty"  validate:"omitempty,max=4000"`
	Status        *string  `json:"status,omitempty"       validate:"omitempty,oneof=draft in_progress review completed archived"`
	Deadline      FlexDate `json:"deadline,omitempty"`
	ClearDeadline bool     `json:"clear_deadline,omitempty"`
}

// UnmarshalJSON — тот же алиас "name" → "title", что и у CreateProjectRequest.
func (r *UpdateProjectRequest) UnmarshalJSON(data []byte) error {
	type alias UpdateProjectRequest
	aux := &struct {
		Name *string `json:"name"`
		*alias
	}{alias: (*alias)(r)}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if r.Title == nil && aux.Name != nil {
		r.Title = aux.Name
	}
	return nil
}

// TasksCountResponse — разбивка задач проекта по статусам.
type TasksCountResponse struct {
	Total      int `json:"total"`
	Todo       int `json:"todo"`
	InProgress int `json:"in_progress"`
	Done       int `json:"done"`
	Blocked    int `json:"blocked"`
}

type ProjectResponse struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	Name        string              `json:"name"` // алиас title — некоторые клиенты ждут это имя
	Description string              `json:"description"`
	GroupID     string              `json:"group_id"`
	OwnerID     string              `json:"owner_id"`
	Status      string              `json:"status"`
	Deadline    *time.Time          `json:"deadline,omitempty"`
	Progress    int                 `json:"progress"`
	TasksCount  TasksCountResponse  `json:"tasks_count"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type PaginatedProjects struct {
	Items  []ProjectResponse `json:"items"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

// Task

type CreateTaskRequest struct {
	Title       string   `json:"title"       validate:"required,min=2,max=200"`
	Description string   `json:"description" validate:"max=4000"`
	AssigneeID  *string  `json:"assignee_id,omitempty" validate:"omitempty,uuid"`
	Priority    string   `json:"priority"    validate:"required,oneof=low medium high critical"`
	DueDate     FlexDate `json:"due_date,omitempty"`
	// ProjectID используется только плоским POST /tasks (без :id в пути) —
	// для вложенного POST /projects/:id/tasks достаточно параметра пути.
	ProjectID *string `json:"project_id,omitempty" validate:"omitempty,uuid"`
}

type UpdateTaskRequest struct {
	Title         *string  `json:"title,omitempty"       validate:"omitempty,min=2,max=200"`
	Description   *string  `json:"description,omitempty" validate:"omitempty,max=4000"`
	AssigneeID    *string  `json:"assignee_id,omitempty" validate:"omitempty,uuid"`
	ClearAssignee bool     `json:"clear_assignee,omitempty"`
	Status        *string  `json:"status,omitempty"      validate:"omitempty,oneof=todo in_progress done blocked"`
	Priority      *string  `json:"priority,omitempty"    validate:"omitempty,oneof=low medium high critical"`
	Order         *int     `json:"order,omitempty"`
	DueDate       FlexDate `json:"due_date,omitempty"`
	ClearDueDate  bool     `json:"clear_due_date,omitempty"`
}

// MoveTaskRequest — атомарное перемещение карточки на Kanban-доске между
// колонками (или внутри одной) с указанием новой позиции сортировки.
type MoveTaskRequest struct {
	TaskID    string `json:"task_id"    validate:"required,uuid"`
	NewStatus string `json:"new_status" validate:"required,oneof=todo in_progress done blocked"`
	NewOrder  int    `json:"new_order"`
}

type TaskResponse struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	AssigneeID    *string    `json:"assignee_id,omitempty"`
	Status        string     `json:"status"`
	Priority      string     `json:"priority"`
	DueDate       *time.Time `json:"due_date,omitempty"`
	Order         int        `json:"order"`
	CommentsCount int        `json:"comments_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
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
