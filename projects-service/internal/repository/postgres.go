package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/student-pm/projects-service/internal/domain"
	"github.com/student-pm/projects-service/internal/usecase"
)

// Repos — фасад над pgx-репозиториями. Один pool на все три репо.
type Repos struct {
	Projects *ProjectRepo
	Tasks    *TaskRepo
	Comments *CommentRepo
}

func NewRepos(pool *pgxpool.Pool) *Repos {
	return &Repos{
		Projects: &ProjectRepo{pool: pool},
		Tasks:    &TaskRepo{pool: pool},
		Comments: &CommentRepo{pool: pool},
	}
}

// =================================================================
// ProjectRepo
// =================================================================

type ProjectRepo struct {
	pool *pgxpool.Pool
}

func (r *ProjectRepo) Create(ctx context.Context, p *domain.Project) error {
	const q = `
		INSERT INTO projects (id, title, description, group_id, owner_id, status, deadline, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, q,
		p.ID, p.Title, p.Description, p.GroupID, p.OwnerID, p.Status, p.Deadline, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *ProjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	const q = `
		SELECT id, title, description, group_id, owner_id, status, deadline, created_at, updated_at
		FROM projects WHERE id = $1
	`
	p := &domain.Project{}
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.Title, &p.Description, &p.GroupID, &p.OwnerID, &p.Status, &p.Deadline, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *ProjectRepo) List(ctx context.Context, f usecase.ProjectListFilter) ([]domain.Project, int, error) {
	conds := []string{"1=1"}
	args := []any{}
	if f.GroupID != nil {
		args = append(args, *f.GroupID)
		conds = append(conds, "group_id = $"+strconv.Itoa(len(args)))
	}
	if f.OwnerID != nil {
		args = append(args, *f.OwnerID)
		conds = append(conds, "owner_id = $"+strconv.Itoa(len(args)))
	}
	if f.Status != nil {
		args = append(args, *f.Status)
		conds = append(conds, "status = $"+strconv.Itoa(len(args)))
	}
	where := strings.Join(conds, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM projects WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	q := "SELECT id, title, description, group_id, owner_id, status, deadline, created_at, updated_at " +
		"FROM projects WHERE " + where +
		" ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)-1) + " OFFSET $" + strconv.Itoa(len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]domain.Project, 0)
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.GroupID, &p.OwnerID,
			&p.Status, &p.Deadline, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

func (r *ProjectRepo) Update(ctx context.Context, p *domain.Project) error {
	const q = `
		UPDATE projects
		SET title = $1, description = $2, status = $3, deadline = $4, updated_at = $5
		WHERE id = $6
	`
	tag, err := r.pool.Exec(ctx, q, p.Title, p.Description, p.Status, p.Deadline, p.UpdatedAt, p.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM projects WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

// =================================================================
// TaskRepo
// =================================================================

type TaskRepo struct {
	pool *pgxpool.Pool
}

func (r *TaskRepo) Create(ctx context.Context, t *domain.Task) error {
	const q = `
		INSERT INTO tasks (id, project_id, title, description, assignee_id, status, priority, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, q,
		t.ID, t.ProjectID, t.Title, t.Description, t.AssigneeID,
		t.Status, t.Priority, t.DueDate, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (r *TaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	const q = `
		SELECT id, project_id, title, description, assignee_id, status, priority, due_date, created_at, updated_at
		FROM tasks WHERE id = $1
	`
	t := &domain.Task{}
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.AssigneeID,
		&t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *TaskRepo) ListByProject(ctx context.Context, projectID uuid.UUID, f usecase.TaskListFilter) ([]domain.Task, int, error) {
	conds := []string{"project_id = $1"}
	args := []any{projectID}
	if f.Status != nil {
		args = append(args, *f.Status)
		conds = append(conds, "status = $"+strconv.Itoa(len(args)))
	}
	if f.Priority != nil {
		args = append(args, *f.Priority)
		conds = append(conds, "priority = $"+strconv.Itoa(len(args)))
	}
	if f.AssigneeID != nil {
		args = append(args, *f.AssigneeID)
		conds = append(conds, "assignee_id = $"+strconv.Itoa(len(args)))
	}
	where := strings.Join(conds, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM tasks WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	q := "SELECT id, project_id, title, description, assignee_id, status, priority, due_date, created_at, updated_at " +
		"FROM tasks WHERE " + where +
		" ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)-1) + " OFFSET $" + strconv.Itoa(len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]domain.Task, 0)
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.AssigneeID,
			&t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, rows.Err()
}

func (r *TaskRepo) Update(ctx context.Context, t *domain.Task) error {
	const q = `
		UPDATE tasks
		SET title = $1, description = $2, assignee_id = $3, status = $4, priority = $5,
		    due_date = $6, updated_at = $7
		WHERE id = $8
	`
	tag, err := r.pool.Exec(ctx, q,
		t.Title, t.Description, t.AssigneeID, t.Status, t.Priority, t.DueDate, t.UpdatedAt, t.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM tasks WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

// Stats — все агрегации одним SQL'ем через FILTER.
// Здесь же считаем overdue: due_date < now AND status != 'done'.
func (r *TaskRepo) Stats(ctx context.Context, projectID uuid.UUID, now time.Time) (*domain.ProjectStats, error) {
	const q = `
		SELECT
		  COUNT(*)                                                          AS total,
		  COUNT(*) FILTER (WHERE status   = 'todo')                         AS s_todo,
		  COUNT(*) FILTER (WHERE status   = 'in_progress')                  AS s_in_progress,
		  COUNT(*) FILTER (WHERE status   = 'done')                         AS s_done,
		  COUNT(*) FILTER (WHERE status   = 'blocked')                      AS s_blocked,
		  COUNT(*) FILTER (WHERE priority = 'low')                          AS p_low,
		  COUNT(*) FILTER (WHERE priority = 'medium')                       AS p_medium,
		  COUNT(*) FILTER (WHERE priority = 'high')                         AS p_high,
		  COUNT(*) FILTER (WHERE priority = 'critical')                     AS p_critical,
		  COUNT(*) FILTER (WHERE due_date IS NOT NULL
		                   AND due_date < $2 AND status != 'done')          AS overdue
		FROM tasks WHERE project_id = $1
	`
	var (
		total, todo, inProgress, done, blocked              int
		pLow, pMedium, pHigh, pCritical, overdue            int
	)
	err := r.pool.QueryRow(ctx, q, projectID, now).Scan(
		&total, &todo, &inProgress, &done, &blocked,
		&pLow, &pMedium, &pHigh, &pCritical, &overdue,
	)
	if err != nil {
		return nil, err
	}

	var pct float64
	if total > 0 {
		pct = float64(done) / float64(total) * 100.0
	}
	return &domain.ProjectStats{
		ProjectID:  projectID,
		TotalTasks: total,
		ByStatus: map[domain.TaskStatus]int{
			domain.TaskTodo:       todo,
			domain.TaskInProgress: inProgress,
			domain.TaskDone:       done,
			domain.TaskBlocked:    blocked,
		},
		ByPriority: map[domain.TaskPriority]int{
			domain.PriorityLow:      pLow,
			domain.PriorityMedium:   pMedium,
			domain.PriorityHigh:     pHigh,
			domain.PriorityCritical: pCritical,
		},
		OverdueCount: overdue,
		DonePercent:  pct,
	}, nil
}

// =================================================================
// CommentRepo
// =================================================================

type CommentRepo struct {
	pool *pgxpool.Pool
}

func (r *CommentRepo) Create(ctx context.Context, c *domain.Comment) error {
	const q = `
		INSERT INTO comments (id, task_id, user_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, q, c.ID, c.TaskID, c.UserID, c.Content, c.CreatedAt)
	return err
}

func (r *CommentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	const q = `SELECT id, task_id, user_id, content, created_at FROM comments WHERE id = $1`
	c := &domain.Comment{}
	err := r.pool.QueryRow(ctx, q, id).Scan(&c.ID, &c.TaskID, &c.UserID, &c.Content, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCommentNotFound
		}
		return nil, err
	}
	return c, nil
}

func (r *CommentRepo) ListByTask(ctx context.Context, taskID uuid.UUID) ([]domain.Comment, error) {
	const q = `
		SELECT id, task_id, user_id, content, created_at
		FROM comments WHERE task_id = $1 ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, q, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Comment, 0)
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CommentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM comments WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCommentNotFound
	}
	return nil
}
