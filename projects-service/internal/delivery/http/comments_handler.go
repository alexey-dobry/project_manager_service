package httpdelivery

import (
	"github.com/gofiber/fiber/v2"

	httperr "github.com/student-pm/projects-service/internal/pkg/errors"
	"github.com/student-pm/projects-service/internal/usecase"
)

// CreateComment godoc
// @Summary      Создать комментарий к задаче
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        task_id path string true "Task UUID"
// @Param        body    body CreateCommentRequest true "Comment payload"
// @Success      201 {object} CommentResponse
// @Router       /tasks/{task_id}/comments [post]
func (h *Handler) CreateComment(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	taskID, err := parseUUIDParam(c, "task_id")
	if err != nil {
		return err
	}
	var req CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}
	cm, err := h.svc.CreateComment(c.UserContext(), actor, taskID, usecase.CreateCommentInput{
		Content: req.Content,
	})
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(toCommentResponse(cm))
}

// ListComments godoc
// @Summary      Список комментариев задачи
// @Tags         comments
// @Produce      json
// @Security     BearerAuth
// @Param        task_id path string true "Task UUID"
// @Success      200 {array} CommentResponse
// @Router       /tasks/{task_id}/comments [get]
func (h *Handler) ListComments(c *fiber.Ctx) error {
	if _, err := h.actorOr401(c); err != nil {
		return err
	}
	taskID, err := parseUUIDParam(c, "task_id")
	if err != nil {
		return err
	}
	cs, err := h.svc.ListComments(c.UserContext(), taskID)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	out := make([]CommentResponse, 0, len(cs))
	for i := range cs {
		out = append(out, toCommentResponse(&cs[i]))
	}
	return c.JSON(out)
}

// DeleteComment godoc
// @Summary      Удалить комментарий
// @Description  Автор или teacher/admin
// @Tags         comments
// @Security     BearerAuth
// @Param        task_id    path string true "Task UUID"
// @Param        comment_id path string true "Comment UUID"
// @Success      204
// @Failure      403 {object} httperr.Body
// @Router       /tasks/{task_id}/comments/{comment_id} [delete]
func (h *Handler) DeleteComment(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	taskID, err := parseUUIDParam(c, "task_id")
	if err != nil {
		return err
	}
	commentID, err := parseUUIDParam(c, "comment_id")
	if err != nil {
		return err
	}
	if err := h.svc.DeleteComment(c.UserContext(), actor, taskID, commentID); err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
