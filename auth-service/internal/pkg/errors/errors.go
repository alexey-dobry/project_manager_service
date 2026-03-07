package httperr

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/student-pm/auth-service/internal/domain"
)

// Body — формат ошибки из ТЗ: { "error": { "code": "...", "message": "..." } }.
type Body struct {
	Error Detail `json:"error"`
}

// Detail — детали ошибки. Details — для валидационных ошибок (поле → причина).
type Detail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Send отдаёт ошибку клиенту в едином формате.
func Send(c *fiber.Ctx, status int, code, msg string, details map[string]string) error {
	return c.Status(status).JSON(Body{Error: Detail{
		Code: code, Message: msg, Details: details,
	}})
}

// FromDomain мапит доменную ошибку в HTTP-ответ.
// Если ошибка незнакома — 500 с обезличенным сообщением (детали — в логи).
func FromDomain(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return Send(c, http.StatusNotFound, "user_not_found", "user not found", nil)
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return Send(c, http.StatusConflict, "user_already_exists", "user with this email already exists", nil)
	case errors.Is(err, domain.ErrInvalidCredentials):
		return Send(c, http.StatusUnauthorized, "invalid_credentials", "invalid email or password", nil)
	case errors.Is(err, domain.ErrInvalidToken),
		errors.Is(err, domain.ErrTokenRevoked):
		return Send(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token", nil)
	case errors.Is(err, domain.ErrForbidden):
		return Send(c, http.StatusForbidden, "forbidden", "forbidden", nil)
	case errors.Is(err, domain.ErrInvalidRole):
		return Send(c, http.StatusBadRequest, "invalid_role", "role is not allowed", nil)
	default:
		return Send(c, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
