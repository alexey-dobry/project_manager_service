package httperr

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/student-pm/groups-service/internal/domain"
)

type Body struct {
	Error Detail `json:"error"`
}

type Detail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func Send(c *fiber.Ctx, status int, code, msg string, details map[string]string) error {
	return c.Status(status).JSON(Body{Error: Detail{
		Code: code, Message: msg, Details: details,
	}})
}

func FromDomain(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrGroupNotFound):
		return Send(c, http.StatusNotFound, "group_not_found", "group not found", nil)
	case errors.Is(err, domain.ErrGroupAlreadyExists):
		return Send(c, http.StatusConflict, "group_already_exists", "group with this name already exists", nil)
	case errors.Is(err, domain.ErrMemberAlreadyInGroup):
		return Send(c, http.StatusConflict, "member_already_in_group", "user is already a member of this group", nil)
	case errors.Is(err, domain.ErrMemberNotFound):
		return Send(c, http.StatusNotFound, "member_not_found", "user is not a member of this group", nil)
	case errors.Is(err, domain.ErrForbidden):
		return Send(c, http.StatusForbidden, "forbidden", "forbidden", nil)
	case errors.Is(err, domain.ErrInvalidToken):
		return Send(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token", nil)
	case errors.Is(err, domain.ErrInvalidRole),
		errors.Is(err, domain.ErrInvalidMembership):
		return Send(c, http.StatusBadRequest, "invalid_role", "role is not allowed", nil)
	default:
		return Send(c, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
