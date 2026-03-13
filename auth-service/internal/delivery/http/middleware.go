package httpdelivery

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/student-pm/auth-service/internal/domain"
	httperr "github.com/student-pm/auth-service/internal/pkg/errors"
	"github.com/student-pm/auth-service/internal/usecase"
)

// Ключи в Locals.
const (
	ctxRequestID = "request_id"
	ctxUserID    = "user_id"
	ctxUserRole  = "user_role"
)

// RequestID присваивает каждому запросу X-Request-ID (или принимает входящий).
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Locals(ctxRequestID, id)
		c.Set("X-Request-ID", id)
		return c.Next()
	}
}

// AccessLog — структурный лог запроса с длительностью и кодом ответа.
func AccessLog(log zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		reqID, _ := c.Locals(ctxRequestID).(string)

		log.Info().
			Str("request_id", reqID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("duration", time.Since(start)).
			Str("ip", c.IP()).
			Str("ua", c.Get("User-Agent")).
			Msg("http_request")

		return err
	}
}

// AuthRequired парсит Authorization: Bearer <token> и кладёт user_id+role в Locals.
func AuthRequired(tp usecase.TokenProvider) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Get("Authorization")
		if !strings.HasPrefix(raw, "Bearer ") {
			return httperr.Send(c, fiber.StatusUnauthorized, "missing_token", "missing bearer token", nil)
		}
		tok := strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
		uid, role, err := tp.ParseAccess(tok)
		if err != nil {
			return httperr.FromDomain(c, domain.ErrInvalidToken)
		}
		c.Locals(ctxUserID, uid)
		c.Locals(ctxUserRole, role)
		return c.Next()
	}
}

// RequireRole — RBAC: пускает только указанные роли.
func RequireRole(roles ...domain.Role) fiber.Handler {
	allow := make(map[domain.Role]struct{}, len(roles))
	for _, r := range roles {
		allow[r] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals(ctxUserRole).(domain.Role)
		if _, ok := allow[role]; !ok {
			return httperr.FromDomain(c, domain.ErrForbidden)
		}
		return c.Next()
	}
}

// userIDFrom извлекает user_id из контекста (для handlers).
func userIDFrom(c *fiber.Ctx) (uuid.UUID, bool) {
	v, ok := c.Locals(ctxUserID).(uuid.UUID)
	return v, ok
}

// roleFrom извлекает роль из контекста.
func roleFrom(c *fiber.Ctx) (domain.Role, bool) {
	v, ok := c.Locals(ctxUserRole).(domain.Role)
	return v, ok
}
