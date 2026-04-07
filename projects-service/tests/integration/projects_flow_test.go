package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	gjwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/student-pm/projects-service/internal/domain"
	httpdelivery "github.com/student-pm/projects-service/internal/delivery/http"
	jwtpkg "github.com/student-pm/projects-service/internal/pkg/jwt"
	"github.com/student-pm/projects-service/internal/pkg/validator"
	"github.com/student-pm/projects-service/internal/usecase"
)

const testSecret = "test-secret-test-secret-1234567890"

// ===== тестовый эмулятор auth-service =====

func issueAccessToken(t *testing.T, userID uuid.UUID, role domain.Role) string {
	t.Helper()
	now := time.Now().UTC()
	tok := gjwt.NewWithClaims(gjwt.SigningMethodHS256, gjwt.MapClaims{
		"sub":  userID.String(),
		"role": string(role),
		"iss":  "test-auth",
		"iat":  now.Unix(),
		"nbf":  now.Unix(),
		"exp":  now.Add(time.Minute).Unix(),
		"jti":  uuid.NewString(),
	})
	signed, err := tok.SignedString([]byte(testSecret))
	require.NoError(t, err)
	return signed
}

// ===== in-memory репо =====

type memProjectRepo struct {
	mu    sync.Mutex
	items map[uuid.UUID]*domain.Project
}

func (m *memProjectRepo) Create(_ context.Context, p *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *p
	m.items[p.ID] = &cp
	return nil
}
func (m *memProjectRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.items[id]
	if !ok {
		return nil, domain.ErrProjectNotFound
	}
	cp := *p
	return &cp, nil
}
func (m *memProjectRepo) List(_ context.Context, f usecase.ProjectListFilter) ([]domain.Project, int, error) {
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
	return out, len(out), nil
}
func (m *memProjectRepo) Update(_ context.Context, p *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[p.ID]; !ok {
		return domain.ErrProjectNotFound
	}
	cp := *p
	m.items[p.ID] = &cp
	return nil
}
func (m *memProjectRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return domain.ErrProjectNotFound
	}
	delete(m.items, id)
	return nil
}

type memTaskRepo struct {
	mu    sync.Mutex
	items map[uuid.UUID]*domain.Task
}

func (m *memTaskRepo) Create(_ context.Context, t *domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *t
	m.items[t.ID] = &cp
	return nil
}
func (m *memTaskRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.items[id]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}
	cp := *t
	return &cp, nil
}
func (m *memTaskRepo) ListByProject(_ context.Context, projectID uuid.UUID, f usecase.TaskListFilter) ([]domain.Task, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Task, 0)
	for _, t := range m.items {
		if t.ProjectID == projectID {
			out = append(out, *t)
		}
	}
	return out, len(out), nil
}
func (m *memTaskRepo) Update(_ context.Context, t *domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[t.ID]; !ok {
		return domain.ErrTaskNotFound
	}
	cp := *t
	m.items[t.ID] = &cp
	return nil
}
func (m *memTaskRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return domain.ErrTaskNotFound
	}
	delete(m.items, id)
	return nil
}
func (m *memTaskRepo) Stats(_ context.Context, projectID uuid.UUID, now time.Time) (*domain.ProjectStats, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stats := &domain.ProjectStats{
		ProjectID:  projectID,
		ByStatus:   map[domain.TaskStatus]int{},
		ByPriority: map[domain.TaskPriority]int{},
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

type memCommentRepo struct {
	mu    sync.Mutex
	items map[uuid.UUID]*domain.Comment
}

func (m *memCommentRepo) Create(_ context.Context, c *domain.Comment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *c
	m.items[c.ID] = &cp
	return nil
}
func (m *memCommentRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Comment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.items[id]
	if !ok {
		return nil, domain.ErrCommentNotFound
	}
	cp := *c
	return &cp, nil
}
func (m *memCommentRepo) ListByTask(_ context.Context, taskID uuid.UUID) ([]domain.Comment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Comment, 0)
	for _, c := range m.items {
		if c.TaskID == taskID {
			out = append(out, *c)
		}
	}
	return out, nil
}
func (m *memCommentRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return domain.ErrCommentNotFound
	}
	delete(m.items, id)
	return nil
}

// ===== app builder =====

func buildApp(t *testing.T) *fiber.App {
	t.Helper()
	pr := &memProjectRepo{items: map[uuid.UUID]*domain.Project{}}
	tr := &memTaskRepo{items: map[uuid.UUID]*domain.Task{}}
	cr := &memCommentRepo{items: map[uuid.UUID]*domain.Comment{}}
	svc := usecase.NewService(pr, tr, cr, time.Now)

	ver, err := jwtpkg.New(testSecret)
	require.NoError(t, err)
	v := validator.New()
	hh := httpdelivery.NewHandler(svc, v)

	app := fiber.New()
	app.Use(httpdelivery.RequestID())
	httpdelivery.RegisterRoutes(app, hh, ver)
	return app
}

func doJSON(t *testing.T, app *fiber.App, method, path, token string, body any) (*http.Response, []byte) {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, data
}

// ===== сценарий =====

func TestFullProjectFlow(t *testing.T) {
	app := buildApp(t)

	owner := uuid.New()
	ownerToken := issueAccessToken(t, owner, domain.RoleStudent)
	groupID := uuid.New()

	// 1) создание проекта
	resp, body := doJSON(t, app, http.MethodPost, "/projects", ownerToken, map[string]any{
		"title": "Курсовая", "description": "ВКР", "group_id": groupID.String(),
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))
	var p httpdelivery.ProjectResponse
	require.NoError(t, json.Unmarshal(body, &p))
	require.Equal(t, "draft", p.Status)

	// 2) создание задачи
	resp, body = doJSON(t, app, http.MethodPost, "/projects/"+p.ID+"/tasks", ownerToken, map[string]any{
		"title": "Подготовить план", "priority": "high",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))
	var task httpdelivery.TaskResponse
	require.NoError(t, json.Unmarshal(body, &task))
	require.Equal(t, "todo", task.Status)

	// 3) перевод задачи в in_progress
	resp, _ = doJSON(t, app, http.MethodPatch, "/projects/"+p.ID+"/tasks/"+task.ID, ownerToken, map[string]any{
		"status": "in_progress",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// 4) попытка перевести задачу из in_progress сразу в blocked → ok
	resp, _ = doJSON(t, app, http.MethodPatch, "/projects/"+p.ID+"/tasks/"+task.ID, ownerToken, map[string]any{
		"status": "blocked",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// 5) blocked → done напрямую запрещено
	resp, body = doJSON(t, app, http.MethodPatch, "/projects/"+p.ID+"/tasks/"+task.ID, ownerToken, map[string]any{
		"status": "done",
	})
	require.Equal(t, http.StatusConflict, resp.StatusCode, string(body))
	require.Contains(t, string(body), "invalid_transition")

	// 6) комментарий к задаче
	resp, body = doJSON(t, app, http.MethodPost, "/tasks/"+task.ID+"/comments", ownerToken, map[string]any{
		"content": "Не успеваем, нужна помощь",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))

	// 7) список комментариев
	resp, body = doJSON(t, app, http.MethodGet, "/tasks/"+task.ID+"/comments", ownerToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var comments []httpdelivery.CommentResponse
	require.NoError(t, json.Unmarshal(body, &comments))
	require.Len(t, comments, 1)

	// 8) stats
	resp, body = doJSON(t, app, http.MethodGet, "/projects/"+p.ID+"/stats", ownerToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var stats httpdelivery.StatsResponse
	require.NoError(t, json.Unmarshal(body, &stats))
	require.Equal(t, 1, stats.TotalTasks)
	require.Equal(t, 1, stats.ByStatus["blocked"])

	// 9) удаление задачи
	resp, _ = doJSON(t, app, http.MethodDelete, "/projects/"+p.ID+"/tasks/"+task.ID, ownerToken, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// 10) удаление проекта
	resp, _ = doJSON(t, app, http.MethodDelete, "/projects/"+p.ID, ownerToken, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestProjects_NoTokenIs401(t *testing.T) {
	app := buildApp(t)
	resp, _ := doJSON(t, app, http.MethodGet, "/projects", "", nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestProjects_StrangerCannotEdit(t *testing.T) {
	app := buildApp(t)
	owner := uuid.New()
	stranger := uuid.New()
	ownerToken := issueAccessToken(t, owner, domain.RoleStudent)
	strangerToken := issueAccessToken(t, stranger, domain.RoleStudent)

	resp, body := doJSON(t, app, http.MethodPost, "/projects", ownerToken, map[string]any{
		"title": "P", "group_id": uuid.NewString(),
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))
	var p httpdelivery.ProjectResponse
	require.NoError(t, json.Unmarshal(body, &p))

	resp, _ = doJSON(t, app, http.MethodPatch, "/projects/"+p.ID, strangerToken, map[string]any{
		"title": "hijacked",
	})
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestHealth(t *testing.T) {
	app := buildApp(t)
	resp, _ := doJSON(t, app, http.MethodGet, "/health", "", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
