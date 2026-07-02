package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/student-pm/projects-service/internal/domain"
	"github.com/student-pm/projects-service/internal/usecase"
)

func newSUT() (*usecase.Service, *MockProjectRepo, *MockTaskRepo, *MockCommentRepo) {
	pr := NewMockProjectRepo()
	tr := NewMockTaskRepo()
	cr := NewMockCommentRepo()
	return usecase.NewService(pr, tr, cr, time.Now), pr, tr, cr
}

func actor(role domain.Role) usecase.Actor {
	return usecase.Actor{ID: uuid.New(), Role: role}
}

// ===== Projects =====

func TestCreateProject_DefaultsToDraft(t *testing.T) {
	svc, _, _, _ := newSUT()
	a := actor(domain.RoleStudent)
	p, err := svc.CreateProject(context.Background(), a, usecase.CreateProjectInput{
		Title: "Курсовая", GroupID: uuid.New(),
	})
	require.NoError(t, err)
	require.Equal(t, domain.ProjectDraft, p.Status)
	require.Equal(t, a.ID, p.OwnerID, "owner == actor")
}

func TestUpdateProject_OnlyOwnerOrElevated(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})

	stranger := actor(domain.RoleStudent)
	newTitle := "hacked"
	_, err := svc.UpdateProject(context.Background(), stranger, p.ID, usecase.UpdateProjectInput{Title: &newTitle})
	require.ErrorIs(t, err, domain.ErrForbidden)

	// owner может
	_, err = svc.UpdateProject(context.Background(), owner, p.ID, usecase.UpdateProjectInput{Title: &newTitle})
	require.NoError(t, err)

	// teacher тоже может
	teacher := actor(domain.RoleTeacher)
	_, err = svc.UpdateProject(context.Background(), teacher, p.ID, usecase.UpdateProjectInput{Title: &newTitle})
	require.NoError(t, err)
}

func TestUpdateProject_StateMachine(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})
	// draft → completed запрещён
	completed := domain.ProjectCompleted
	_, err := svc.UpdateProject(context.Background(), owner, p.ID, usecase.UpdateProjectInput{Status: &completed})
	require.ErrorIs(t, err, domain.ErrInvalidTransition)

	// draft → in_progress → review → completed
	inProg := domain.ProjectInProgress
	_, err = svc.UpdateProject(context.Background(), owner, p.ID, usecase.UpdateProjectInput{Status: &inProg})
	require.NoError(t, err)
	review := domain.ProjectReview
	_, err = svc.UpdateProject(context.Background(), owner, p.ID, usecase.UpdateProjectInput{Status: &review})
	require.NoError(t, err)
	_, err = svc.UpdateProject(context.Background(), owner, p.ID, usecase.UpdateProjectInput{Status: &completed})
	require.NoError(t, err)
}

func TestUpdateProject_ClearDeadline(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	dl := time.Now().Add(48 * time.Hour)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(), Deadline: &dl,
	})
	require.NotNil(t, p.Deadline)

	updated, err := svc.UpdateProject(context.Background(), owner, p.ID, usecase.UpdateProjectInput{
		ClearDeadline: true,
	})
	require.NoError(t, err)
	require.Nil(t, updated.Deadline)
}

func TestDeleteProject_Forbidden(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})
	require.ErrorIs(t,
		svc.DeleteProject(context.Background(), actor(domain.RoleStudent), p.ID),
		domain.ErrForbidden)
	require.NoError(t, svc.DeleteProject(context.Background(), owner, p.ID))
}

// ===== Tasks =====

func TestCreateTask_OnlyManager(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})

	_, err := svc.CreateTask(context.Background(), actor(domain.RoleStudent), p.ID, usecase.CreateTaskInput{
		Title: "Task1", Priority: domain.PriorityMedium,
	})
	require.ErrorIs(t, err, domain.ErrForbidden)

	t1, err := svc.CreateTask(context.Background(), owner, p.ID, usecase.CreateTaskInput{
		Title: "Task1", Priority: domain.PriorityMedium,
	})
	require.NoError(t, err)
	require.Equal(t, domain.TaskTodo, t1.Status)
}

func TestUpdateTask_AssigneeCanChangeStatusOnly(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})
	assignee := actor(domain.RoleStudent)
	t1, err := svc.CreateTask(context.Background(), owner, p.ID, usecase.CreateTaskInput{
		Title: "T", AssigneeID: &assignee.ID, Priority: domain.PriorityMedium,
	})
	require.NoError(t, err)

	// assignee может перевести todo → in_progress
	inProg := domain.TaskInProgress
	updated, err := svc.UpdateTask(context.Background(), assignee, p.ID, t1.ID, usecase.UpdateTaskInput{
		Status: &inProg,
	})
	require.NoError(t, err)
	require.Equal(t, domain.TaskInProgress, updated.Status)

	// но не может менять приоритет
	high := domain.PriorityHigh
	_, err = svc.UpdateTask(context.Background(), assignee, p.ID, t1.ID, usecase.UpdateTaskInput{
		Priority: &high,
	})
	require.ErrorIs(t, err, domain.ErrForbidden)

	// и не может назначать другого
	other := uuid.New()
	_, err = svc.UpdateTask(context.Background(), assignee, p.ID, t1.ID, usecase.UpdateTaskInput{
		AssigneeID: &other,
	})
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdateTask_RandomUserForbidden(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})
	t1, _ := svc.CreateTask(context.Background(), owner, p.ID, usecase.CreateTaskInput{
		Title: "T", Priority: domain.PriorityMedium,
	})
	stranger := actor(domain.RoleStudent)
	inProg := domain.TaskInProgress
	_, err := svc.UpdateTask(context.Background(), stranger, p.ID, t1.ID, usecase.UpdateTaskInput{Status: &inProg})
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdateTask_InvalidTransition(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})
	t1, _ := svc.CreateTask(context.Background(), owner, p.ID, usecase.CreateTaskInput{
		Title: "T", Priority: domain.PriorityMedium,
	})
	done := domain.TaskDone
	_, err := svc.UpdateTask(context.Background(), owner, p.ID, t1.ID, usecase.UpdateTaskInput{Status: &done})
	require.ErrorIs(t, err, domain.ErrInvalidTransition)
}

func TestGetTask_IDORProtection(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p1, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "P1", GroupID: uuid.New(),
	})
	p2, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "P2", GroupID: uuid.New(),
	})
	t1, _ := svc.CreateTask(context.Background(), owner, p1.ID, usecase.CreateTaskInput{
		Title: "T", Priority: domain.PriorityMedium,
	})

	// Запрос /projects/p2/tasks/t1 — задача из p1 → 404 Task not in project
	_, err := svc.GetTask(context.Background(), p2.ID, t1.ID)
	require.ErrorIs(t, err, domain.ErrTaskNotInProject)
}

// ===== Comments =====

func TestCreateComment_TrimsAndRequiresContent(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})
	t1, _ := svc.CreateTask(context.Background(), owner, p.ID, usecase.CreateTaskInput{
		Title: "T", Priority: domain.PriorityMedium,
	})

	c, err := svc.CreateComment(context.Background(), actor(domain.RoleStudent), t1.ID,
		usecase.CreateCommentInput{Content: "  hello  "})
	require.NoError(t, err)
	require.Equal(t, "hello", c.Content)

	_, err = svc.CreateComment(context.Background(), actor(domain.RoleStudent), t1.ID,
		usecase.CreateCommentInput{Content: "   "})
	require.ErrorIs(t, err, domain.ErrEmptyContent)
}

func TestDeleteComment_AuthorOrAdmin(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})
	t1, _ := svc.CreateTask(context.Background(), owner, p.ID, usecase.CreateTaskInput{
		Title: "T", Priority: domain.PriorityMedium,
	})
	author := actor(domain.RoleStudent)
	cm, _ := svc.CreateComment(context.Background(), author, t1.ID, usecase.CreateCommentInput{Content: "hi"})

	stranger := actor(domain.RoleStudent)
	require.ErrorIs(t, svc.DeleteComment(context.Background(), stranger, t1.ID, cm.ID), domain.ErrForbidden)
	require.NoError(t, svc.DeleteComment(context.Background(), author, t1.ID, cm.ID))
}

func TestDeleteComment_AdminOverride(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})
	t1, _ := svc.CreateTask(context.Background(), owner, p.ID, usecase.CreateTaskInput{
		Title: "T", Priority: domain.PriorityMedium,
	})
	author := actor(domain.RoleStudent)
	cm, _ := svc.CreateComment(context.Background(), author, t1.ID, usecase.CreateCommentInput{Content: "hi"})

	require.NoError(t, svc.DeleteComment(context.Background(), actor(domain.RoleAdmin), t1.ID, cm.ID))
}

// ===== Stats =====

func TestProjectStats_Aggregates(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})

	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	// Создаём задачу и двигаем по легальной последовательности переходов.
	createAndMove := func(target domain.TaskStatus, prio domain.TaskPriority, due *time.Time) {
		t1, err := svc.CreateTask(context.Background(), owner, p.ID, usecase.CreateTaskInput{
			Title: "T", Priority: prio, DueDate: due,
		})
		require.NoError(t, err)
		path := map[domain.TaskStatus][]domain.TaskStatus{
			domain.TaskTodo:       {},
			domain.TaskInProgress: {domain.TaskInProgress},
			domain.TaskDone:       {domain.TaskInProgress, domain.TaskDone},
			domain.TaskBlocked:    {domain.TaskBlocked},
		}[target]
		for _, s := range path {
			st := s
			_, err := svc.UpdateTask(context.Background(), owner, p.ID, t1.ID, usecase.UpdateTaskInput{Status: &st})
			require.NoError(t, err)
		}
	}

	createAndMove(domain.TaskTodo, domain.PriorityLow, nil)
	createAndMove(domain.TaskTodo, domain.PriorityMedium, &past)         // overdue
	createAndMove(domain.TaskInProgress, domain.PriorityHigh, &future)
	createAndMove(domain.TaskDone, domain.PriorityCritical, nil)
	createAndMove(domain.TaskBlocked, domain.PriorityHigh, &past)        // overdue (не done)

	stats, err := svc.ProjectStats(context.Background(), p.ID)
	require.NoError(t, err)
	require.Equal(t, 5, stats.TotalTasks)
	require.Equal(t, 2, stats.ByStatus[domain.TaskTodo])
	require.Equal(t, 1, stats.ByStatus[domain.TaskInProgress])
	require.Equal(t, 1, stats.ByStatus[domain.TaskDone])
	require.Equal(t, 1, stats.ByStatus[domain.TaskBlocked])
	require.Equal(t, 2, stats.OverdueCount)
	require.InDelta(t, 20.0, stats.DonePercent, 0.01) // 1/5 = 20%
}

func TestProjectStats_NotFound(t *testing.T) {
	svc, _, _, _ := newSUT()
	_, err := svc.ProjectStats(context.Background(), uuid.New())
	require.ErrorIs(t, err, domain.ErrProjectNotFound)
}

// ===== Order (Kanban-позиция) =====

func TestCreateTask_OrderAppendsToEndOfColumn(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})

	t1, err := svc.CreateTask(context.Background(), owner, p.ID, usecase.CreateTaskInput{
		Title: "T1", Priority: domain.PriorityMedium,
	})
	require.NoError(t, err)
	require.Equal(t, 0, t1.Order, "первая задача в колонке — order 0")

	t2, err := svc.CreateTask(context.Background(), owner, p.ID, usecase.CreateTaskInput{
		Title: "T2", Priority: domain.PriorityMedium,
	})
	require.NoError(t, err)
	require.Equal(t, 1, t2.Order, "вторая задача встаёт следом")
}

func TestUpdateTask_AssigneeCanReorder(t *testing.T) {
	svc, _, _, _ := newSUT()
	owner := actor(domain.RoleStudent)
	p, _ := svc.CreateProject(context.Background(), owner, usecase.CreateProjectInput{
		Title: "X", GroupID: uuid.New(),
	})
	assignee := actor(domain.RoleStudent)
	t1, err := svc.CreateTask(context.Background(), owner, p.ID, usecase.CreateTaskInput{
		Title: "T", AssigneeID: &assignee.ID, Priority: domain.PriorityMedium,
	})
	require.NoError(t, err)

	// Drag-and-drop: assignee двигает свою задачу в in_progress с новой позицией.
	inProg := domain.TaskInProgress
	newOrder := 3
	updated, err := svc.UpdateTask(context.Background(), assignee, p.ID, t1.ID, usecase.UpdateTaskInput{
		Status: &inProg,
		Order:  &newOrder,
	})
	require.NoError(t, err)
	require.Equal(t, domain.TaskInProgress, updated.Status)
	require.Equal(t, 3, updated.Order)
}
