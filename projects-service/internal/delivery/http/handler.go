package httpdelivery

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/student-pm/projects-service/internal/domain"
	httperr "github.com/student-pm/projects-service/internal/pkg/errors"
	"github.com/student-pm/projects-service/internal/pkg/validator"
	"github.com/student-pm/projects-service/internal/usecase"
)

// Handler агрегирует все HTTP-эндпоинты сервиса.
type Handler struct {
	svc       *usecase.Service
	validator *validator.Validator
}

func NewHandler(svc *usecase.Service, v *validator.Validator) *Handler {
	return &Handler{svc: svc, validator: v}
}

// actorOr401 извлекает Actor из контекста или возвращает 401-ответ.
func (h *Handler) actorOr401(c *fiber.Ctx) (usecase.Actor, error) {
	uid, ok1 := userIDFrom(c)
	role, ok2 := roleFrom(c)
	if !ok1 || !ok2 {
		return usecase.Actor{}, httperr.FromDomain(c, domain.ErrInvalidToken)
	}
	return usecase.Actor{ID: uid, Role: role}, nil
}

func parseUUIDParam(c *fiber.Ctx, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(name))
	if err != nil {
		return uuid.Nil, httperr.Send(c, fiber.StatusBadRequest, "invalid_id", name+" must be UUID", nil)
	}
	return id, nil
}

// PROJECTS

// CreateProject godoc
// @Summary      Создать проект
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateProjectRequest true "Project payload"
// @Success      201 {object} ProjectResponse
// @Failure      400 {object} httperr.Body
// @Router       /projects [post]
func (h *Handler) CreateProject(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	var req CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}
	groupID, err := uuid.Parse(req.GroupID)
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_group_id", "group_id must be UUID", nil)
	}
	p, err := h.svc.CreateProject(c.UserContext(), actor, usecase.CreateProjectInput{
		Title:       req.Title,
		Description: req.Description,
		GroupID:     groupID,
		Deadline:    req.Deadline,
	})
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(toProjectResponse(p))
}

// ListProjects godoc
// @Summary      Список проектов
// @Tags         projects
// @Produce      json
// @Security     BearerAuth
// @Param        group_id query string false "UUID группы"
// @Param        owner_id query string false "UUID владельца"
// @Param        status   query string false "draft|in_progress|review|completed|archived"
// @Param        limit    query int    false "1..100, default 20"
// @Param        offset   query int    false "default 0"
// @Success      200 {object} PaginatedProjects
// @Router       /projects [get]
func (h *Handler) ListProjects(c *fiber.Ctx) error {
	if _, err := h.actorOr401(c); err != nil {
		return err
	}

	f := usecase.ProjectListFilter{}
	if v := c.Query("group_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return httperr.Send(c, fiber.StatusBadRequest, "invalid_group_id", "group_id must be UUID", nil)
		}
		f.GroupID = &id
	}
	if v := c.Query("owner_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return httperr.Send(c, fiber.StatusBadRequest, "invalid_owner_id", "owner_id must be UUID", nil)
		}
		f.OwnerID = &id
	}
	if v := c.Query("status"); v != "" {
		s := domain.ProjectStatus(v)
		if !s.IsValid() {
			return httperr.FromDomain(c, domain.ErrInvalidStatus)
		}
		f.Status = &s
	}
	f.Limit, _ = strconv.Atoi(c.Query("limit"))
	f.Offset, _ = strconv.Atoi(c.Query("offset"))

	items, total, err := h.svc.ListProjects(c.UserContext(), f)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	out := make([]ProjectResponse, 0, len(items))
	for i := range items {
		out = append(out, toProjectResponse(&items[i]))
	}
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	return c.JSON(PaginatedProjects{Items: out, Total: total, Limit: limit, Offset: offset})
}

// GetProject godoc
// @Summary      Получить проект по ID
// @Tags         projects
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project UUID"
// @Success      200 {object} ProjectResponse
// @Failure      404 {object} httperr.Body
// @Router       /projects/{id} [get]
func (h *Handler) GetProject(c *fiber.Ctx) error {
	if _, err := h.actorOr401(c); err != nil {
		return err
	}
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return err
	}
	p, err := h.svc.GetProject(c.UserContext(), id)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toProjectResponse(p))
}

// UpdateProject godoc
// @Summary      Обновить проект
// @Description  Owner проекта или teacher/admin
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string true "Project UUID"
// @Param        body body UpdateProjectRequest true "Patch payload"
// @Success      200 {object} ProjectResponse
// @Failure      403 {object} httperr.Body
// @Failure      409 {object} httperr.Body
// @Router       /projects/{id} [patch]
func (h *Handler) UpdateProject(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return err
	}
	var req UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}
	in := usecase.UpdateProjectInput{
		Title:         req.Title,
		Description:   req.Description,
		Deadline:      req.Deadline,
		ClearDeadline: req.ClearDeadline,
	}
	if req.Status != nil {
		s := domain.ProjectStatus(*req.Status)
		in.Status = &s
	}
	p, err := h.svc.UpdateProject(c.UserContext(), actor, id, in)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toProjectResponse(p))
}

// DeleteProject godoc
// @Summary      Удалить проект
// @Tags         projects
// @Security     BearerAuth
// @Param        id path string true "Project UUID"
// @Success      204
// @Failure      403 {object} httperr.Body
// @Router       /projects/{id} [delete]
func (h *Handler) DeleteProject(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.svc.DeleteProject(c.UserContext(), actor, id); err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ProjectStats godoc
// @Summary      Статистика по задачам проекта
// @Tags         projects
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project UUID"
// @Success      200 {object} StatsResponse
// @Failure      404 {object} httperr.Body
// @Router       /projects/{id}/stats [get]
func (h *Handler) ProjectStats(c *fiber.Ctx) error {
	if _, err := h.actorOr401(c); err != nil {
		return err
	}
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return err
	}
	stats, err := h.svc.ProjectStats(c.UserContext(), id)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toStatsResponse(stats))
}

// system

// Health godoc
// @Summary Health
// @Tags    system
// @Produce json
// @Success 200 {object} MessageResponse
// @Router  /health [get]
func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(MessageResponse{Message: "ok"})
}

// Ready godoc
// @Summary Ready
// @Tags    system
// @Produce json
// @Success 200 {object} MessageResponse
// @Router  /ready [get]
func (h *Handler) Ready(c *fiber.Ctx) error {
	return c.JSON(MessageResponse{Message: "ready"})
}

// mappers

func toProjectResponse(p *domain.Project) ProjectResponse {
	return ProjectResponse{
		ID:          p.ID.String(),
		Title:       p.Title,
		Description: p.Description,
		GroupID:     p.GroupID.String(),
		OwnerID:     p.OwnerID.String(),
		Status:      string(p.Status),
		Deadline:    p.Deadline,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toTaskResponse(t *domain.Task) TaskResponse {
	r := TaskResponse{
		ID:          t.ID.String(),
		ProjectID:   t.ProjectID.String(),
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		Priority:    string(t.Priority),
		DueDate:     t.DueDate,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	if t.AssigneeID != nil {
		s := t.AssigneeID.String()
		r.AssigneeID = &s
	}
	return r
}

func toCommentResponse(c *domain.Comment) CommentResponse {
	return CommentResponse{
		ID:        c.ID.String(),
		TaskID:    c.TaskID.String(),
		UserID:    c.UserID.String(),
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
	}
}

func toStatsResponse(s *domain.ProjectStats) StatsResponse {
	byStatus := make(map[string]int, len(s.ByStatus))
	for k, v := range s.ByStatus {
		byStatus[string(k)] = v
	}
	byPriority := make(map[string]int, len(s.ByPriority))
	for k, v := range s.ByPriority {
		byPriority[string(k)] = v
	}
	return StatsResponse{
		ProjectID:    s.ProjectID.String(),
		TotalTasks:   s.TotalTasks,
		ByStatus:     byStatus,
		ByPriority:   byPriority,
		OverdueCount: s.OverdueCount,
		DonePercent:  s.DonePercent,
	}
}
