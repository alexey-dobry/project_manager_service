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
	return h.createTask(c, actor, projectID)
}

// CreateTaskFlat godoc
// @Summary      Создать задачу (project_id в теле запроса)
// @Description  Тот же сценарий, что POST /projects/{id}/tasks, но без
// @Description  project_id в пути — для клиентов, которым удобнее плоский
// @Description  REST-ресурс /tasks.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateTaskRequest true "Task payload, включая project_id"
// @Success      201 {object} TaskResponse
// @Failure      400 {object} httperr.Body
// @Router       /tasks [post]
func (h *Handler) CreateTaskFlat(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	// project_id читаем из тела заранее отдельным разведочным парсингом —
	// основное тело парсится ещё раз внутри createTask в CreateTaskRequest.
	var probe struct {
		ProjectID *string `json:"project_id"`
	}
	if err := c.BodyParser(&probe); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if probe.ProjectID == nil || *probe.ProjectID == "" {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_project_id", "project_id is required", nil)
	}
	projectID, err := uuid.Parse(*probe.ProjectID)
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_project_id", "project_id must be UUID", nil)
	}
	return h.createTask(c, actor, projectID)
}

// createTask — общая логика создания задачи для вложенного и плоского маршрутов.
func (h *Handler) createTask(c *fiber.Ctx, actor usecase.Actor, projectID uuid.UUID) error {
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
		DueDate:     req.DueDate.Ptr(),
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
	return c.Status(fiber.StatusCreated).JSON(toTaskResponse(t, 0))
}

// ListTasks godoc
// @Summary      Список задач проекта
// @Description  Возвращает плоский массив задач (без пагинации) —
// @Description  используется Kanban-доской, где нужны все задачи проекта сразу.
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id          path  string true  "Project UUID"
// @Param        status      query string false "todo|in_progress|done|blocked"
// @Param        priority    query string false "low|medium|high|critical"
// @Param        assignee_id query string false "UUID исполнителя"
// @Success      200 {array} TaskResponse
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
	// Kanban-доска показывает все задачи проекта разом — верхний предел
	// на случай, если фильтры не заданы явно клиентом.
	f.Limit = 1000
	if v := c.Query("limit"); v != "" {
		f.Limit, _ = strconv.Atoi(v)
	}
	f.Offset, _ = strconv.Atoi(c.Query("offset"))

	items, _, err := h.svc.ListTasks(c.UserContext(), projectID, f)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	out := make([]TaskResponse, 0, len(items))
	for i := range items {
		out = append(out, toTaskResponse(&items[i], h.commentsCountOrZero(c.UserContext(), items[i].ID)))
	}
	return c.JSON(out)
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
	return c.JSON(toTaskResponse(t, h.commentsCountOrZero(c.UserContext(), t.ID)))
}

// GetTaskFlat godoc
// @Summary      Получить задачу по ID (без project_id в пути)
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task UUID"
// @Success      200 {object} TaskResponse
// @Failure      404 {object} httperr.Body
// @Router       /tasks/{id} [get]
func (h *Handler) GetTaskFlat(c *fiber.Ctx) error {
	if _, err := h.actorOr401(c); err != nil {
		return err
	}
	taskID, err := parseUUIDParam(c, "id")
	if err != nil {
		return err
	}
	t, err := h.svc.GetTaskByID(c.UserContext(), taskID)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toTaskResponse(t, h.commentsCountOrZero(c.UserContext(), t.ID)))
}

// UpdateTask godoc
// @Summary      Обновить задачу
// @Description  Менеджер проекта (owner/teacher/admin) — все поля. Assignee — только status, order и description.
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
	return h.updateTask(c, actor, projectID, taskID)
}

// UpdateTaskFlat godoc
// @Summary      Обновить задачу по ID (без project_id в пути)
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string true "Task UUID"
// @Param        body body UpdateTaskRequest true "Patch payload"
// @Success      200 {object} TaskResponse
// @Failure      403 {object} httperr.Body
// @Router       /tasks/{id} [patch]
func (h *Handler) UpdateTaskFlat(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	taskID, err := parseUUIDParam(c, "id")
	if err != nil {
		return err
	}
	t, err := h.svc.GetTaskByID(c.UserContext(), taskID)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return h.updateTask(c, actor, t.ProjectID, taskID)
}

// updateTask — общая логика обновления задачи для вложенного и плоского маршрутов.
func (h *Handler) updateTask(c *fiber.Ctx, actor usecase.Actor, projectID, taskID uuid.UUID) error {
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
		Order:         req.Order,
		DueDate:       req.DueDate.Ptr(),
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
	return c.JSON(toTaskResponse(t, h.commentsCountOrZero(c.UserContext(), t.ID)))
}

// MoveTask godoc
// @Summary      Переместить задачу на Kanban-доске
// @Description  Атомарно меняет статус (колонку) и позицию сортировки внутри неё.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body MoveTaskRequest true "task_id, new_status, new_order"
// @Success      200 {object} TaskResponse
// @Failure      403 {object} httperr.Body
// @Failure      409 {object} httperr.Body
// @Router       /tasks/move [post]
func (h *Handler) MoveTask(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	var req MoveTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}
	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_task_id", "task_id must be UUID", nil)
	}
	t, err := h.svc.GetTaskByID(c.UserContext(), taskID)
	if err != nil {
		return httperr.FromDomain(c, err)
	}

	status := domain.TaskStatus(req.NewStatus)
	order := req.NewOrder
	updated, err := h.svc.UpdateTask(c.UserContext(), actor, t.ProjectID, taskID, usecase.UpdateTaskInput{
		Status: &status,
		Order:  &order,
	})
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toTaskResponse(updated, h.commentsCountOrZero(c.UserContext(), updated.ID)))
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

// DeleteTaskFlat godoc
// @Summary      Удалить задачу по ID (без project_id в пути)
// @Tags         tasks
// @Security     BearerAuth
// @Param        id path string true "Task UUID"
// @Success      204
// @Failure      403 {object} httperr.Body
// @Router       /tasks/{id} [delete]
func (h *Handler) DeleteTaskFlat(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	taskID, err := parseUUIDParam(c, "id")
	if err != nil {
		return err
	}
	t, err := h.svc.GetTaskByID(c.UserContext(), taskID)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	if err := h.svc.DeleteTask(c.UserContext(), actor, t.ProjectID, taskID); err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
