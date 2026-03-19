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
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/student-pm/auth-service/internal/domain"
	httpdelivery "github.com/student-pm/auth-service/internal/delivery/http"
	"github.com/student-pm/auth-service/internal/pkg/hasher"
	jwtpkg "github.com/student-pm/auth-service/internal/pkg/jwt"
	"github.com/student-pm/auth-service/internal/pkg/validator"
	"github.com/student-pm/auth-service/internal/usecase"
)

// In-process интеграционный тест: реальный Fiber-app + in-memory репозитории.
// Полноценный e2e на testcontainers-go (реальный Postgres) — заготовка в README.

// ===== in-memory репозитории =====

type memUserRepo struct {
	mu      sync.Mutex
	byID    map[uuid.UUID]*domain.User
	byEmail map[string]*domain.User
}

func newMemUserRepo() *memUserRepo {
	return &memUserRepo{
		byID:    map[uuid.UUID]*domain.User{},
		byEmail: map[string]*domain.User{},
	}
}

func (m *memUserRepo) Create(_ context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byEmail[u.Email]; ok {
		return domain.ErrUserAlreadyExists
	}
	cp := *u
	m.byID[u.ID] = &cp
	m.byEmail[u.Email] = &cp
	return nil
}

func (m *memUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *memUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byEmail[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *memUserRepo) Update(_ context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byID[u.ID]; !ok {
		return domain.ErrUserNotFound
	}
	cp := *u
	m.byID[u.ID] = &cp
	m.byEmail[u.Email] = &cp
	return nil
}

type memTokenRepo struct {
	mu     sync.Mutex
	byHash map[string]*domain.RefreshToken
}

func newMemTokenRepo() *memTokenRepo {
	return &memTokenRepo{byHash: map[string]*domain.RefreshToken{}}
}

func (m *memTokenRepo) Save(_ context.Context, t *domain.RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *t
	m.byHash[t.TokenHash] = &cp
	return nil
}

func (m *memTokenRepo) GetByHash(_ context.Context, hash string) (*domain.RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.byHash[hash]
	if !ok {
		return nil, domain.ErrInvalidToken
	}
	cp := *t
	return &cp, nil
}

func (m *memTokenRepo) Revoke(_ context.Context, id uuid.UUID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.byHash {
		if t.ID == id && t.RevokedAt == nil {
			t.RevokedAt = &at
		}
	}
	return nil
}

func (m *memTokenRepo) RevokeAllForUser(_ context.Context, uid uuid.UUID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.byHash {
		if t.UserID == uid && t.RevokedAt == nil {
			t.RevokedAt = &at
		}
	}
	return nil
}

// ===== app builder =====

func buildApp(t *testing.T) *fiber.App {
	t.Helper()
	users := newMemUserRepo()
	tokens := newMemTokenRepo()
	h := hasher.New(4) // bcrypt MinCost для скорости
	tp, err := jwtpkg.New(jwtpkg.Config{
		Secret:     "test-secret-test-secret-1234567890",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	})
	require.NoError(t, err)
	svc := usecase.NewAuthService(users, tokens, h, tp, time.Now)
	v := validator.New()
	hh := httpdelivery.NewHandler(svc, v)

	app := fiber.New()
	app.Use(httpdelivery.RequestID())
	httpdelivery.RegisterRoutes(app, hh, tp)
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

// ===== сценарии =====

func TestAuthFlow_RegisterLoginMe(t *testing.T) {
	app := buildApp(t)

	// 1) register
	resp, body := doJSON(t, app, http.MethodPost, "/auth/register", "", map[string]string{
		"email":     "ivan@uni.edu",
		"password":  "secret123",
		"full_name": "Ivan I",
		"role":      "student",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, string(body))

	var auth httpdelivery.AuthResponse
	require.NoError(t, json.Unmarshal(body, &auth))
	require.NotEmpty(t, auth.AccessToken)
	require.Equal(t, "ivan@uni.edu", auth.User.Email)

	// 2) /users/me с access-токеном
	resp, body = doJSON(t, app, http.MethodGet, "/users/me", auth.AccessToken, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	var me httpdelivery.UserResponse
	require.NoError(t, json.Unmarshal(body, &me))
	require.Equal(t, auth.User.ID, me.ID)

	// 3) /users/me без токена → 401
	resp, _ = doJSON(t, app, http.MethodGet, "/users/me", "", nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// 4) login с верным паролем
	resp, body = doJSON(t, app, http.MethodPost, "/auth/login", "", map[string]string{
		"email":    "ivan@uni.edu",
		"password": "secret123",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))

	// 5) login с неверным паролем → 401
	resp, _ = doJSON(t, app, http.MethodPost, "/auth/login", "", map[string]string{
		"email":    "ivan@uni.edu",
		"password": "WRONG",
	})
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// 6) refresh — выдаёт новую пару, старый токен после этого не работает
	resp, body = doJSON(t, app, http.MethodPost, "/auth/refresh", "", map[string]string{
		"refresh_token": auth.RefreshToken,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	var refreshed httpdelivery.AuthResponse
	require.NoError(t, json.Unmarshal(body, &refreshed))
	require.NotEqual(t, auth.RefreshToken, refreshed.RefreshToken)

	resp, _ = doJSON(t, app, http.MethodPost, "/auth/refresh", "", map[string]string{
		"refresh_token": auth.RefreshToken,
	})
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "старый refresh должен быть отозван")
}

func TestRegister_ValidationError(t *testing.T) {
	app := buildApp(t)
	resp, body := doJSON(t, app, http.MethodPost, "/auth/register", "", map[string]string{
		"email":     "not-an-email",
		"password":  "x",
		"full_name": "A",
		"role":      "student",
	})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Contains(t, string(body), "validation_failed")
}

func TestHealth(t *testing.T) {
	app := buildApp(t)
	resp, _ := doJSON(t, app, http.MethodGet, "/health", "", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
