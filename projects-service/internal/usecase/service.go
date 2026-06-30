package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/student-pm/projects-service/internal/domain"
)

// Service содержит сценарии работы с проектами, задачами и комментариями.
type Service struct {
	projects ProjectRepository
	tasks    TaskRepository
	comments CommentRepository
	now      func() time.Time
}

func NewService(
	p ProjectRepository, t TaskRepository, c CommentRepository,
	now func() time.Time,
) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{projects: p, tasks: t, comments: c, now: now}
}

// PROJECTS

// CreateProject доступен любому авторизованному пользователю.
// Владельцем проекта становится инициатор операции.
func (s *Service) CreateProject(ctx context.Context, actor Actor, in CreateProjectInput) (*domain.Project, error) {
	now := s.now().UTC()
	p := &domain.Project{
		ID:          uuid.New(),
		Title:       strings.TrimSpace(in.Title),
		Description: in.Description,
		GroupID:     in.GroupID,
		OwnerID:     actor.ID,
		Status:      domain.ProjectDraft,
		Deadline:    in.Deadline,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.projects.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) GetProject(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	return s.projects.GetByID(ctx, id)
}

func (s *Service) ListProjects(ctx context.Context, f ProjectListFilter) ([]domain.Project, int, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return s.projects.List(ctx, f)
}

// UpdateProject — может только owner или teacher/admin.
// Смена статуса валидируется через state-machine.
func (s *Service) UpdateProject(ctx context.Context, actor Actor, id uuid.UUID, in UpdateProjectInput) (*domain.Project, error) {
	p, err := s.projects.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canManageProject(actor, p) {
		return nil, domain.ErrForbidden
	}

	if in.Title != nil {
		p.Title = strings.TrimSpace(*in.Title)
	}
	if in.Description != nil {
		p.Description = *in.Description
	}
	if in.ClearDeadline {
		p.Deadline = nil
	} else if in.Deadline != nil {
		d := *in.Deadline
		p.Deadline = &d
	}
	if in.Status != nil {
		if !in.Status.IsValid() {
			return nil, domain.ErrInvalidStatus
		}
		if !p.Status.CanTransitionTo(*in.Status) {
			return nil, domain.ErrInvalidTransition
		}
		p.Status = *in.Status
	}
	p.UpdatedAt = s.now().UTC()

	if err := s.projects.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) DeleteProject(ctx context.Context, actor Actor, id uuid.UUID) error {
	p, err := s.projects.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !canManageProject(actor, p) {
		return domain.ErrForbidden
	}
	return s.projects.Delete(ctx, id)
}

func (s *Service) ProjectStats(ctx context.Context, projectID uuid.UUID) (*domain.ProjectStats, error) {
	// Гарантируем 404 для несуществующего проекта (а не пустой stats).
	if _, err := s.projects.GetByID(ctx, projectID); err != nil {
		return nil, err
	}
	return s.tasks.Stats(ctx, projectID, s.now().UTC())
}

// TASKS

// CreateTask: создавать задачи может тот, кто может управлять проектом
// (owner, teacher, admin).
func (s *Service) CreateTask(ctx context.Context, actor Actor, projectID uuid.UUID, in CreateTaskInput) (*domain.Task, error) {
	p, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if !canManageProject(actor, p) {
		return nil, domain.ErrForbidden
	}
	if !in.Priority.IsValid() {
		return nil, domain.ErrInvalidPriority
	}

	now := s.now().UTC()
	t := &domain.Task{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Title:       strings.TrimSpace(in.Title),
		Description: in.Description,
		AssigneeID:  in.AssigneeID,
		Status:      domain.TaskTodo,
		Priority:    in.Priority,
		DueDate:     in.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.tasks.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// GetTask возвращает задачу с проверкой её принадлежности указанному проекту.
func (s *Service) GetTask(ctx context.Context, projectID, taskID uuid.UUID) (*domain.Task, error) {
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if t.ProjectID != projectID {
		return nil, domain.ErrTaskNotInProject
	}
	return t, nil
}

func (s *Service) ListTasks(ctx context.Context, projectID uuid.UUID, f TaskListFilter) ([]domain.Task, int, error) {
	if _, err := s.projects.GetByID(ctx, projectID); err != nil {
		return nil, 0, err
	}
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return s.tasks.ListByProject(ctx, projectID, f)
}

// UpdateTask: assignee может менять статус и описание; полный апдейт — менеджер проекта.
// Назначать assignee и менять приоритет может только менеджер проекта.
func (s *Service) UpdateTask(ctx context.Context, actor Actor, projectID, taskID uuid.UUID, in UpdateTaskInput) (*domain.Task, error) {
	p, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if t.ProjectID != projectID {
		return nil, domain.ErrTaskNotInProject
	}

	manager := canManageProject(actor, p)
	assignee := t.AssigneeID != nil && *t.AssigneeID == actor.ID
	if !manager && !assignee {
		return nil, domain.ErrForbidden
	}

	// Поля, которые может менять только менеджер проекта.
	if !manager {
		if in.AssigneeID != nil || in.ClearAssignee || in.Priority != nil ||
			in.DueDate != nil || in.ClearDueDate || in.Title != nil {
			return nil, domain.ErrForbidden
		}
	}

	if in.Title != nil {
		t.Title = strings.TrimSpace(*in.Title)
	}
	if in.Description != nil {
		t.Description = *in.Description
	}
	if in.ClearAssignee {
		t.AssigneeID = nil
	} else if in.AssigneeID != nil {
		a := *in.AssigneeID
		t.AssigneeID = &a
	}
	if in.Priority != nil {
		if !in.Priority.IsValid() {
			return nil, domain.ErrInvalidPriority
		}
		t.Priority = *in.Priority
	}
	if in.Status != nil {
		if !in.Status.IsValid() {
			return nil, domain.ErrInvalidStatus
		}
		if !t.Status.CanTransitionTo(*in.Status) {
			return nil, domain.ErrInvalidTransition
		}
		t.Status = *in.Status
	}
	if in.ClearDueDate {
		t.DueDate = nil
	} else if in.DueDate != nil {
		d := *in.DueDate
		t.DueDate = &d
	}
	t.UpdatedAt = s.now().UTC()

	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) DeleteTask(ctx context.Context, actor Actor, projectID, taskID uuid.UUID) error {
	p, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return err
	}
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if t.ProjectID != projectID {
		return domain.ErrTaskNotInProject
	}
	if !canManageProject(actor, p) {
		return domain.ErrForbidden
	}
	return s.tasks.Delete(ctx, taskID)
}

// COMMENTS

// CreateComment: писать комментарий может любой авторизованный пользователь.
func (s *Service) CreateComment(ctx context.Context, actor Actor, taskID uuid.UUID, in CreateCommentInput) (*domain.Comment, error) {
	if _, err := s.tasks.GetByID(ctx, taskID); err != nil {
		return nil, err
	}
	content := strings.TrimSpace(in.Content)
	if content == "" {
		return nil, domain.ErrEmptyContent
	}
	if len(content) > domain.MaxCommentLength {
		return nil, domain.ErrContentTooLong
	}
	c := &domain.Comment{
		ID:        uuid.New(),
		TaskID:    taskID,
		UserID:    actor.ID,
		Content:   content,
		CreatedAt: s.now().UTC(),
	}
	if err := s.comments.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) ListComments(ctx context.Context, taskID uuid.UUID) ([]domain.Comment, error) {
	if _, err := s.tasks.GetByID(ctx, taskID); err != nil {
		return nil, err
	}
	return s.comments.ListByTask(ctx, taskID)
}

// DeleteComment: автор или teacher/admin.
func (s *Service) DeleteComment(ctx context.Context, actor Actor, taskID, commentID uuid.UUID) error {
	c, err := s.comments.GetByID(ctx, commentID)
	if err != nil {
		return err
	}
	if c.TaskID != taskID {
		return domain.ErrCommentNotInTask
	}
	if c.UserID != actor.ID && !isElevated(actor.Role) {
		return domain.ErrForbidden
	}
	return s.comments.Delete(ctx, commentID)
}

// RBAC helpers

// canManageProject — кто может править/удалять проект и его задачи.
// Owner проекта или глобально привилегированный (teacher/admin).
func canManageProject(a Actor, p *domain.Project) bool {
	return p.OwnerID == a.ID || isElevated(a.Role)
}

func isElevated(r domain.Role) bool {
	return r == domain.RoleTeacher || r == domain.RoleAdmin
}
