package httpdelivery

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/student-pm/groups-service/internal/domain"
	httperr "github.com/student-pm/groups-service/internal/pkg/errors"
	"github.com/student-pm/groups-service/internal/pkg/jwt"
)

const (
	ctxRequestID = "request_id"
	ctxUserID    = "user_id"
	ctxUserRole  = "user_role"
)

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
			Msg("http_request")

		return err
	}
}

// AuthRequired — middleware проверки Bearer-токена.
func AuthRequired(v *jwt.Verifier) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Get("Authorization")
		if !strings.HasPrefix(raw, "Bearer ") {
			return httperr.Send(c, fiber.StatusUnauthorized, "missing_token", "missing bearer token", nil)
		}
		tok := strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
		uid, role, err := v.ParseAccess(tok)
		if err != nil {
			return httperr.FromDomain(c, domain.ErrInvalidToken)
		}
		c.Locals(ctxUserID, uid)
		c.Locals(ctxUserRole, role)
		return c.Next()
	}
}

// userIDFrom / roleFrom — экстракторы Actor из контекста.
func userIDFrom(c *fiber.Ctx) (uuid.UUID, bool) {
	v, ok := c.Locals(ctxUserID).(uuid.UUID)
	return v, ok
}

func roleFrom(c *fiber.Ctx) (domain.Role, bool) {
	v, ok := c.Locals(ctxUserRole).(domain.Role)
	return v, ok
}
