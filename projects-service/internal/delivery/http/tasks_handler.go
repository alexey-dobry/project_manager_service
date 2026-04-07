package httpdelivery

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/student-pm/projects-service/internal/domain"
	httperr "github.com/student-pm/projects-service/internal/pkg/errors"
	"github.com/student-pm/projects-service/internal/usecase"
)

// CreateTask godoc
// @Summary      Создать задачу в проекте
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string true "Project UUID"
// @Param        body body CreateTaskRequest true "Task payload"
// @Success      201 {object} TaskResponse
// @Failure      403 {object} httperr.Body
// @Router       /projects/{id}/tasks [post]
func (h *Handler) CreateTask(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	projectID, err := parseUUIDParam(c, "id")
	if err != nil {
		return err
	}
	var req CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}
	in := usecase.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Priority:    domain.TaskPriority(req.Priority),
		DueDate:     req.DueDate,
	}
	if req.AssigneeID != nil {
		uid, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			return httperr.Send(c, fiber.StatusBadRequest, "invalid_assignee_id", "assignee_id must be UUID", nil)
		}
		in.AssigneeID = &uid
	}
	t, err := h.svc.CreateTask(c.UserContext(), actor, projectID, in)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(toTaskResponse(t))
}

// ListTasks godoc
// @Summary      Список задач проекта
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id          path  string true  "Project UUID"
// @Param        status      query string false "todo|in_progress|done|blocked"
// @Param        priority    query string false "low|medium|high|critical"
// @Param        assignee_id query string false "UUID исполнителя"
// @Param        limit       query int    false "1..100, default 20"
// @Param        offset      query int    false "default 0"
// @Success      200 {object} PaginatedTasks
// @Router       /projects/{id}/tasks [get]
func (h *Handler) ListTasks(c *fiber.Ctx) error {
	if _, err := h.actorOr401(c); err != nil {
		return err
	}
	projectID, err := parseUUIDParam(c, "id")
	if err != nil {
		return err
	}
	f := usecase.TaskListFilter{}
	if v := c.Query("status"); v != "" {
		s := domain.TaskStatus(v)
		if !s.IsValid() {
			return httperr.FromDomain(c, domain.ErrInvalidStatus)
		}
		f.Status = &s
	}
	if v := c.Query("priority"); v != "" {
		p := domain.TaskPriority(v)
		if !p.IsValid() {
			return httperr.FromDomain(c, domain.ErrInvalidPriority)
		}
		f.Priority = &p
	}
	if v := c.Query("assignee_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return httperr.Send(c, fiber.StatusBadRequest, "invalid_assignee_id", "assignee_id must be UUID", nil)
		}
		f.AssigneeID = &id
	}
	f.Limit, _ = strconv.Atoi(c.Query("limit"))
	f.Offset, _ = strconv.Atoi(c.Query("offset"))

	items, total, err := h.svc.ListTasks(c.UserContext(), projectID, f)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	out := make([]TaskResponse, 0, len(items))
	for i := range items {
		out = append(out, toTaskResponse(&items[i]))
	}
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	return c.JSON(PaginatedTasks{Items: out, Total: total, Limit: limit, Offset: offset})
}

// GetTask godoc
// @Summary      Получить задачу
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id      path string true "Project UUID"
// @Param        task_id path string true "Task UUID"
// @Success      200 {object} TaskResponse
// @Failure      404 {object} httperr.Body
// @Router       /projects/{id}/tasks/{task_id} [get]
func (h *Handler) GetTask(c *fiber.Ctx) error {
	if _, err := h.actorOr401(c); err != nil {
		return err
	}
	projectID, err := parseUUIDParam(c, "id")
	if err != nil {
		return err
	}
	taskID, err := parseUUIDParam(c, "task_id")
	if err != nil {
		return err
	}
	t, err := h.svc.GetTask(c.UserContext(), projectID, taskID)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toTaskResponse(t))
}

// UpdateTask godoc
// @Summary      Обновить задачу
// @Description  Менеджер проекта (owner/teacher/admin) — все поля. Assignee — только status и description.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path string true "Project UUID"
// @Param        task_id path string true "Task UUID"
// @Param        body    body UpdateTaskRequest true "Patch payload"
// @Success      200 {object} TaskResponse
// @Failure      403 {object} httperr.Body
// @Failure      409 {object} httperr.Body
// @Router       /projects/{id}/tasks/{task_id} [patch]
func (h *Handler) UpdateTask(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	projectID, err := parseUUIDParam(c, "id")
	if err != nil {
		return err
	}
	taskID, err := parseUUIDParam(c, "task_id")
	if err != nil {
		return err
	}
	var req UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}
	in := usecase.UpdateTaskInput{
		Title:         req.Title,
		Description:   req.Description,
		ClearAssignee: req.ClearAssignee,
		DueDate:       req.DueDate,
		ClearDueDate:  req.ClearDueDate,
	}
	if req.AssigneeID != nil {
		uid, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			return httperr.Send(c, fiber.StatusBadRequest, "invalid_assignee_id", "assignee_id must be UUID", nil)
		}
		in.AssigneeID = &uid
	}
	if req.Status != nil {
		s := domain.TaskStatus(*req.Status)
		in.Status = &s
	}
	if req.Priority != nil {
		p := domain.TaskPriority(*req.Priority)
		in.Priority = &p
	}
	t, err := h.svc.UpdateTask(c.UserContext(), actor, projectID, taskID, in)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toTaskResponse(t))
}

// DeleteTask godoc
// @Summary      Удалить задачу
// @Tags         tasks
// @Security     BearerAuth
// @Param        id      path string true "Project UUID"
// @Param        task_id path string true "Task UUID"
// @Success      204
// @Failure      403 {object} httperr.Body
// @Router       /projects/{id}/tasks/{task_id} [delete]
func (h *Handler) DeleteTask(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	projectID, err := parseUUIDParam(c, "id")
	if err != nil {
		return err
	}
	taskID, err := parseUUIDParam(c, "task_id")
	if err != nil {
		return err
	}
	if err := h.svc.DeleteTask(c.UserContext(), actor, projectID, taskID); err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
