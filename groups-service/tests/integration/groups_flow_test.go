package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	gjwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/student-pm/groups-service/internal/domain"
	httpdelivery "github.com/student-pm/groups-service/internal/delivery/http"
	"github.com/student-pm/groups-service/internal/pkg/authclient"
	jwtpkg "github.com/student-pm/groups-service/internal/pkg/jwt"
	"github.com/student-pm/groups-service/internal/pkg/validator"
	"github.com/student-pm/groups-service/internal/usecase"
)

const testSecret = "test-secret-test-secret-1234567890"

// issueAccessToken генерирует валидный access-токен для тестов.
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


type memRepo struct {
	mu      sync.Mutex
	groups  map[uuid.UUID]*domain.Group
	members map[mkKey]*domain.Membership
}

type mkKey struct {
	G uuid.UUID
	U uuid.UUID
}

func newMemRepo() *memRepo {
	return &memRepo{
		groups:  map[uuid.UUID]*domain.Group{},
		members: map[mkKey]*domain.Membership{},
	}
}

func (m *memRepo) Create(_ context.Context, g *domain.Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ex := range m.groups {
		if ex.Name == g.Name {
			return domain.ErrGroupAlreadyExists
		}
	}
	cp := *g
	m.groups[g.ID] = &cp
	return nil
}
func (m *memRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Group, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	g, ok := m.groups[id]
	if !ok {
		return nil, domain.ErrGroupNotFound
	}
	cp := *g
	return &cp, nil
}
func (m *memRepo) List(_ context.Context, f usecase.ListFilter) ([]domain.Group, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Group, 0, len(m.groups))
	for _, g := range m.groups {
		if f.Faculty != "" && g.Faculty != f.Faculty {
			continue
		}
		if f.Course > 0 && g.Course != f.Course {
			continue
		}
		out = append(out, *g)
	}
	return out, len(out), nil
}
func (m *memRepo) Update(_ context.Context, g *domain.Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.groups[g.ID]; !ok {
		return domain.ErrGroupNotFound
	}
	cp := *g
	m.groups[g.ID] = &cp
	return nil
}
func (m *memRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.groups[id]; !ok {
		return domain.ErrGroupNotFound
	}
	delete(m.groups, id)
	return nil
}
func (m *memRepo) AddMember(_ context.Context, mb *domain.Membership) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := mkKey{G: mb.GroupID, U: mb.UserID}
	if _, ok := m.members[k]; ok {
		return domain.ErrMemberAlreadyInGroup
	}
	cp := *mb
	m.members[k] = &cp
	return nil
}
func (m *memRepo) RemoveMember(_ context.Context, gID, uID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := mkKey{G: gID, U: uID}
	if _, ok := m.members[k]; !ok {
		return domain.ErrMemberNotFound
	}
	delete(m.members, k)
	return nil
}
func (m *memRepo) ListMembers(_ context.Context, gID uuid.UUID) ([]domain.Membership, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Membership, 0)
	for k, v := range m.members {
		if k.G == gID {
			out = append(out, *v)
		}
	}
	return out, nil
}
func (m *memRepo) GetMember(_ context.Context, gID, uID uuid.UUID) (*domain.Membership, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.members[mkKey{G: gID, U: uID}]
	if !ok {
		return nil, domain.ErrMemberNotFound
	}
	cp := *v
	return &cp, nil
}


// fakeAuthService — минимальная имитация GET /users/{id} auth-service.
// Отдаёт детерминированное ФИО/роль по UUID, чтобы проверить обогащение
// ответа GET /groups/{id} в groups-service.
func fakeAuthService(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/users/"):]
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"id":        id,
			"full_name": "Test User " + id[:8],
			"role":      "student",
		})
	})
	return httptest.NewServer(mux)
}

func buildApp(t *testing.T) *fiber.App {
	t.Helper()
	repo := newMemRepo()
	svc := usecase.NewGroupService(repo, time.Now)
	ver, err := jwtpkg.New(testSecret)
	require.NoError(t, err)
	v := validator.New()

	authSrv := fakeAuthService(t)
	t.Cleanup(authSrv.Close)
	authClient := authclient.New(authSrv.URL)

	hh := httpdelivery.NewHandler(svc, v, authClient)

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


func TestGroupsFlow_AdminCreatesAndManages(t *testing.T) {
	app := buildApp(t)

	adminID := uuid.New()
	leaderID := uuid.New()
	studentID := uuid.New()

	adminToken := issueAccessToken(t, adminID, domain.RoleAdmin)
	leaderToken := issueAccessToken(t, leaderID, domain.RoleStudent) // глобально student
	studentToken := issueAccessToken(t, studentID, domain.RoleStudent)

	// 1) admin создаёт группу
	resp, body := doJSON(t, app, http.MethodPost, "/groups", adminToken, map[string]any{
		"name": "БПИ-211", "course": 2, "faculty": "ФИТ", "leader_id": leaderID.String(),
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))
	var g httpdelivery.GroupResponse
	require.NoError(t, json.Unmarshal(body, &g))

	// 2) student не может создать
	resp, _ = doJSON(t, app, http.MethodPost, "/groups", studentToken, map[string]any{
		"name": "x", "course": 1, "faculty": "y", "leader_id": uuid.NewString(),
	})
	require.Equal(t, http.StatusForbidden, resp.StatusCode)

	// 3) лидер добавляет участника
	resp, _ = doJSON(t, app, http.MethodPost, "/groups/"+g.ID+"/members", leaderToken, map[string]any{
		"user_id": studentID.String(), "role_in_group": "member",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// 4) рандомный студент не может добавить
	resp, _ = doJSON(t, app, http.MethodPost, "/groups/"+g.ID+"/members", studentToken, map[string]any{
		"user_id": uuid.NewString(), "role_in_group": "member",
	})
	require.Equal(t, http.StatusForbidden, resp.StatusCode)

	// 5) список участников: leader (создан автоматически) + student
	resp, body = doJSON(t, app, http.MethodGet, "/groups/"+g.ID+"/members", adminToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	var members []httpdelivery.MembershipResponse
	require.NoError(t, json.Unmarshal(body, &members))
	require.Len(t, members, 2)

	// 5а) GET /groups/{id} возвращает leader и members, обогащённые данными
	// из (фейкового) auth-service — регрессия на баг "r.members undefined".
	resp, body = doJSON(t, app, http.MethodGet, "/groups/"+g.ID, adminToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	var full httpdelivery.GroupWithMembersResponse
	require.NoError(t, json.Unmarshal(body, &full))
	require.NotNil(t, full.Leader)
	require.NotEmpty(t, full.Leader.FullName)
	require.Len(t, full.Members, 2)

	// 6) студент сам выходит из группы
	resp, _ = doJSON(t, app, http.MethodDelete, "/groups/"+g.ID+"/members/"+studentID.String(), studentToken, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// 7) admin удаляет группу
	resp, _ = doJSON(t, app, http.MethodDelete, "/groups/"+g.ID, adminToken, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestGroups_NoTokenIs401(t *testing.T) {
	app := buildApp(t)
	resp, _ := doJSON(t, app, http.MethodGet, "/groups", "", nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGroups_ValidationError(t *testing.T) {
	app := buildApp(t)
	tok := issueAccessToken(t, uuid.New(), domain.RoleAdmin)
	resp, body := doJSON(t, app, http.MethodPost, "/groups", tok, map[string]any{
		"name": "x", "course": 0, "faculty": "", "leader_id": "not-a-uuid",
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Contains(t, string(body), "validation_failed")
}

func TestHealth(t *testing.T) {
	app := buildApp(t)
	resp, _ := doJSON(t, app, http.MethodGet, "/health", "", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestAddMember_DefaultsRoleToMember — регрессия под фронтовый вызов
// groupsApi.addMember, который шлёт только { user_id }, без role_in_group.
func TestAddMember_DefaultsRoleToMember(t *testing.T) {
	app := buildApp(t)
	admin := uuid.New()
	adminToken := issueAccessToken(t, admin, domain.RoleAdmin)
	leaderID := uuid.New()
	studentID := uuid.New()

	resp, body := doJSON(t, app, http.MethodPost, "/groups", adminToken, map[string]any{
		"name": "БПИ-212", "course": 2, "faculty": "ФИТ", "leader_id": leaderID.String(),
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))
	var g httpdelivery.GroupResponse
	require.NoError(t, json.Unmarshal(body, &g))

	// Ровно та форма запроса, что шлёт фронт: только user_id.
	resp, body = doJSON(t, app, http.MethodPost, "/groups/"+g.ID+"/members", adminToken, map[string]any{
		"user_id": studentID.String(),
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))
	var m httpdelivery.MembershipResponse
	require.NoError(t, json.Unmarshal(body, &m))
	require.Equal(t, "member", m.RoleInGroup)
}
