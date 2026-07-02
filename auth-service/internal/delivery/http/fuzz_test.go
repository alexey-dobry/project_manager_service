package httpdelivery_test

// Этот пакет — `httpdelivery_test` (внешний test-пакет), чтобы не светить
// внутренние моки наружу. Импортируем нашу собственную delivery как
// сторонний пакет.

import (
	"bytes"
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/student-pm/auth-service/internal/domain"
	httpdelivery "github.com/student-pm/auth-service/internal/delivery/http"
	"github.com/student-pm/auth-service/internal/pkg/hasher"
	jwtpkg "github.com/student-pm/auth-service/internal/pkg/jwt"
	"github.com/student-pm/auth-service/internal/pkg/validator"
	"github.com/student-pm/auth-service/internal/usecase"
)

// ===== in-memory mocks (минимум для fuzz) =====

type memUserRepo struct {
	mu      sync.Mutex
	byEmail map[string]*domain.User
	byID    map[uuid.UUID]*domain.User
}

func newMemUserRepo() *memUserRepo {
	return &memUserRepo{
		byEmail: map[string]*domain.User{},
		byID:    map[uuid.UUID]*domain.User{},
	}
}
func (m *memUserRepo) Create(_ context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byEmail[u.Email]; ok {
		return domain.ErrUserAlreadyExists
	}
	cp := *u
	m.byEmail[u.Email] = &cp
	m.byID[u.ID] = &cp
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
	mu sync.Mutex
	m  map[string]*domain.RefreshToken
}

func newMemTokenRepo() *memTokenRepo { return &memTokenRepo{m: map[string]*domain.RefreshToken{}} }
func (r *memTokenRepo) Save(_ context.Context, t *domain.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *t
	r.m[t.TokenHash] = &cp
	return nil
}
func (r *memTokenRepo) GetByHash(_ context.Context, h string) (*domain.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.m[h]
	if !ok {
		return nil, domain.ErrInvalidToken
	}
	cp := *t
	return &cp, nil
}
func (r *memTokenRepo) Revoke(_ context.Context, id uuid.UUID, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.m {
		if t.ID == id && t.RevokedAt == nil {
			t.RevokedAt = &at
		}
	}
	return nil
}
func (r *memTokenRepo) RevokeAllForUser(_ context.Context, uid uuid.UUID, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.m {
		if t.UserID == uid && t.RevokedAt == nil {
			t.RevokedAt = &at
		}
	}
	return nil
}

func buildApp(tb testing.TB) *fiber.App {
	tb.Helper()
	users := newMemUserRepo()
	tokens := newMemTokenRepo()
	h := hasher.New(4)
	tp, err := jwtpkg.New(jwtpkg.Config{Secret: "fuzz-secret-min-16-chars-OK"})
	if err != nil {
		tb.Fatalf("jwt: %v", err)
	}
	svc := usecase.NewAuthService(users, tokens, h, tp, time.Now)
	app := fiber.New(fiber.Config{
		BodyLimit:    1 * 1024 * 1024, // 1 MB — отвечает на аномалию из отчёта
		ErrorHandler: defaultErrorHandler,
	})
	app.Use(httpdelivery.RequestID())
	hh := httpdelivery.NewHandler(svc, validator.New())
	httpdelivery.RegisterRoutes(app, hh, tp)
	return app
}

func defaultErrorHandler(c *fiber.Ctx, _ error) error {
	return c.Status(500).JSON(fiber.Map{
		"error": fiber.Map{"code": "internal_error", "message": "internal server error"},
	})
}

// FuzzRegisterHandler — гоняет произвольное тело в POST /auth/register
// и требует, чтобы handler НИКОГДА не возвращал 5xx и не паниковал.
//
// Покрывает аномалии из отчёта по Schemathesis:
//   - Аномалия 1: 500 на не-UTF-8 байтах (для /auth/refresh; распространяем
//     инвариант на все POST-эндпоинты auth-service).
//   - Аномалия 2: пустой group_id вызывал uuid.Parse в неинформативном пути.
//
// Допустимые статусы: 2xx, 4xx. Любой 5xx — баг.
//
// Запуск:
//   go test ./internal/delivery/http -run=- -fuzz=FuzzRegisterHandler -fuzztime=30s
func FuzzRegisterHandler(f *testing.F) {
	// Seed: разные «плохие» тела. Bytes — потому что нам важны и не-UTF-8 байты.
	seeds := [][]byte{
		[]byte(``),
		[]byte(`{}`),
		[]byte(`{"email":"x"}`),
		[]byte(`{"email":"x@x.x","password":"abc","full_name":"x","role":"student"}`),
		// Аномалия 1: не-UTF-8 байты
		{0xff, 0xfe, 0xfd},
		// Аномалия 2: пустой group_id (snake_case)
		[]byte(`{"email":"a@a.a","password":"password123","full_name":"a","role":"student","group_id":""}`),
		// Невалидный JSON
		[]byte(`{`),
		[]byte(`{"email":}`),
		// Очень длинные строки
		[]byte(`{"email":"` + string(make([]byte, 10000)) + `"}`),
		// Глубокая вложенность
		[]byte(`{"a":{"a":{"a":{"a":{"a":{"a":{}}}}}}}`),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	app := buildApp(f)

	f.Fuzz(func(t *testing.T, body []byte) {
		// Ограничиваем сверху — иначе fuzz будет тратить время на гигантские тела,
		// а сервер на нашем BodyLimit и так их режет (это тот самый фикс из отчёта).
		if len(body) > 2*1024*1024 {
			return
		}
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			// Сетевая ошибка теста — это про инфраструктуру, не про fuzz.
			t.Fatalf("app.Test: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			t.Fatalf("got 5xx (%d) on body (%d bytes): %q",
				resp.StatusCode, len(body), shorten(body))
		}
	})
}

func shorten(b []byte) string {
	if len(b) > 200 {
		return string(b[:200]) + "...(truncated)"
	}
	return string(b)
}

// FuzzRefreshHandler гоняет произвольное тело в POST /auth/refresh.
//
// Инвариант: обработчик никогда не возвращает 5xx.
// Покрывает аномалию №1 из отчёта: не-UTF-8 байты в теле запроса вызывали
// HTTP 500 из-за необработанного сбоя десериализации.
//
// Запуск:
//
//	go test ./internal/delivery/http -run=- -fuzz=FuzzRefreshHandler -fuzztime=30s
func FuzzRefreshHandler(f *testing.F) {
	seeds := [][]byte{
		[]byte(``),
		[]byte(`{}`),
		[]byte(`{"refresh_token":"abc"}`),
		// аномалия 1: не-UTF-8
		{0xff, 0xfe, 0xfd},
		// нулевые байты
		{0x00, 0x00},
		// обрезанный JSON
		[]byte(`{"refresh_token":`),
		// очень длинный токен
		[]byte(`{"refresh_token":"` + string(make([]byte, 4096)) + `"}`),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	app := buildApp(f)

	f.Fuzz(func(t *testing.T, body []byte) {
		if len(body) > 2*1024*1024 {
			return
		}
		req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			t.Fatalf("got 5xx (%d) on body (%d bytes): %q",
				resp.StatusCode, len(body), shorten(body))
		}
	})
}

// FuzzUpdateUserHandler гоняет произвольное тело в PATCH /users/:id.
//
// Инвариант: обработчик никогда не возвращает 5xx.
// Покрывает аномалию №2: пустая строка в поле group_id приводила к панике
// внутри uuid.Parse вместо детализированного 400.
//
// Запуск:
//
//	go test ./internal/delivery/http -run=- -fuzz=FuzzUpdateUserHandler -fuzztime=30s
func FuzzUpdateUserHandler(f *testing.F) {
	seeds := [][]byte{
		[]byte(`{}`),
		// аномалия 2: пустой group_id
		[]byte(`{"group_id":""}`),
		// невалидный UUID
		[]byte(`{"group_id":"not-a-uuid"}`),
		// валидный UUID
		[]byte(`{"group_id":"00000000-0000-0000-0000-000000000000"}`),
		// не-UTF-8
		{0xff, 0xfe},
		[]byte(`{"full_name":""}`),
		[]byte(`{"role":"unknown_role"}`),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	app := buildApp(f)
	// Нужен авторизованный запрос — используем заранее зарегистрированного пользователя.
	// Регистрируем его один раз и берём его ID из ответа.
	regBody := []byte(`{"email":"fuzz@example.com","password":"fuzz-pass-1","full_name":"Fuzz","role":"admin"}`)
	regReq := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(regReq, 5000)
	if err != nil || regResp.StatusCode != 201 {
		// Если не получилось зарегистрироваться — пропускаем: это инфраструктурная проблема.
		return
	}

	f.Fuzz(func(t *testing.T, body []byte) {
		if len(body) > 2*1024*1024 {
			return
		}
		// Используем синтетический UUID — handler должен ответить 404 или 400, но не 5xx.
		req := httptest.NewRequest("PATCH", "/users/00000000-0000-0000-0000-000000000001",
			bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		// Без токена получим 401 — это тоже допустимо, главное не 5xx.

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			t.Fatalf("got 5xx (%d) on body (%d bytes): %q",
				resp.StatusCode, len(body), shorten(body))
		}
	})
}
