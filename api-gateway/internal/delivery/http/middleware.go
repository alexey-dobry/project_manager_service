package httpdelivery

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	httperr "github.com/student-pm/api-gateway/internal/pkg/errors"
	"github.com/student-pm/api-gateway/internal/pkg/jwt"
)

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Locals("request_id", id)
		c.Set("X-Request-ID", id)
		return c.Next()
	}
}

func AccessLog(log zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		reqID, _ := c.Locals("request_id").(string)

		log.Info().
			Str("request_id", reqID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("duration", time.Since(start)).
			Str("ip", c.IP()).
			Msg("gateway_request")
		return err
	}
}

// AuthRequired — pre-validation на gateway.
// Defense in depth: реальная авторизация всё равно происходит на backend
// (он сам читает Authorization), но здесь мы:
//   1) отбрасываем мусорные токены до похода на upstream;
//   2) кладём user_id/role в Locals — для X-User-ID/X-User-Role хедеров и логов.
func AuthRequired(v *jwt.Verifier) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Get("Authorization")
		if !strings.HasPrefix(raw, "Bearer ") {
			return httperr.Unauthorized(c, "missing bearer token")
		}
		tok := strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
		uid, role, err := v.ParseAccess(tok)
		if err != nil {
			return httperr.Unauthorized(c, "invalid or expired token")
		}
		c.Locals("user_id", uid.String())
		c.Locals("user_role", role)
		return c.Next()
	}
}
