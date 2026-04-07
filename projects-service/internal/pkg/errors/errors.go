package httperr

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/student-pm/projects-service/internal/domain"
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
	return c.Status(status).JSON(Body{Error: Detail{Code: code, Message: msg, Details: details}})
}

func FromDomain(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		return Send(c, http.StatusNotFound, "project_not_found", "project not found", nil)
	case errors.Is(err, domain.ErrTaskNotFound):
		return Send(c, http.StatusNotFound, "task_not_found", "task not found", nil)
	case errors.Is(err, domain.ErrCommentNotFound):
		return Send(c, http.StatusNotFound, "comment_not_found", "comment not found", nil)
	case errors.Is(err, domain.ErrTaskNotInProject):
		return Send(c, http.StatusNotFound, "task_not_in_project", "task does not belong to this project", nil)
	case errors.Is(err, domain.ErrCommentNotInTask):
		return Send(c, http.StatusNotFound, "comment_not_in_task", "comment does not belong to this task", nil)
	case errors.Is(err, domain.ErrInvalidStatus):
		return Send(c, http.StatusBadRequest, "invalid_status", "invalid status", nil)
	case errors.Is(err, domain.ErrInvalidPriority):
		return Send(c, http.StatusBadRequest, "invalid_priority", "invalid priority", nil)
	case errors.Is(err, domain.ErrInvalidTransition):
		return Send(c, http.StatusConflict, "invalid_transition", "status transition is not allowed", nil)
	case errors.Is(err, domain.ErrEmptyContent):
		return Send(c, http.StatusBadRequest, "empty_content", "comment content cannot be empty", nil)
	case errors.Is(err, domain.ErrForbidden):
		return Send(c, http.StatusForbidden, "forbidden", "forbidden", nil)
	case errors.Is(err, domain.ErrInvalidToken):
		return Send(c, http.StatusUnauthorized, "invalid_token", "invalid or expired token", nil)
	default:
		return Send(c, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
