package httpdelivery

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/student-pm/groups-service/internal/domain"
	"github.com/student-pm/groups-service/internal/pkg/authclient"
	httperr "github.com/student-pm/groups-service/internal/pkg/errors"
	"github.com/student-pm/groups-service/internal/pkg/validator"
	"github.com/student-pm/groups-service/internal/usecase"
)

type Handler struct {
	groups     *usecase.GroupService
	validator  *validator.Validator
	authClient *authclient.Client
}

func NewHandler(g *usecase.GroupService, v *validator.Validator, ac *authclient.Client) *Handler {
	return &Handler{groups: g, validator: v, authClient: ac}
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

// CreateGroup godoc
// @Summary      Создать группу
// @Description  Доступно teacher/admin
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateGroupRequest true "Group payload"
// @Success      201 {object} GroupResponse
// @Failure      400 {object} httperr.Body
// @Failure      403 {object} httperr.Body
// @Router       /groups [post]
func (h *Handler) CreateGroup(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	var req CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}
	leaderID := actor.ID
	if req.LeaderID != nil {
		parsed, err := uuid.Parse(*req.LeaderID)
		if err != nil {
			return httperr.Send(c, fiber.StatusBadRequest, "invalid_leader_id", "leader_id must be UUID", nil)
		}
		leaderID = parsed
	}
	g, err := h.groups.Create(c.UserContext(), actor, usecase.CreateGroupInput{
		Name: req.Name, Course: req.Course, Faculty: req.Faculty, LeaderID: leaderID,
	})
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(toGroupResponse(g))
}

// ListGroups godoc
// @Summary      Список групп
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Param        faculty query string false "точное совпадение"
// @Param        course  query int    false "номер курса"
// @Param        limit   query int    false "по умолчанию 20, максимум 100"
// @Param        offset  query int    false "по умолчанию 0"
// @Success      200 {object} PaginatedGroupsResponse
// @Router       /groups [get]
func (h *Handler) ListGroups(c *fiber.Ctx) error {
	if _, err := h.actorOr401(c); err != nil {
		return err
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	course, _ := strconv.Atoi(c.Query("course"))

	groups, total, err := h.groups.List(c.UserContext(), usecase.ListFilter{
		Faculty: c.Query("faculty"),
		Course:  course,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return httperr.FromDomain(c, err)
	}

	items := make([]GroupResponse, 0, len(groups))
	for i := range groups {
		items = append(items, toGroupResponse(&groups[i]))
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return c.JSON(PaginatedGroupsResponse{Items: items, Total: total, Limit: limit, Offset: offset})
}

// GetGroup godoc
// @Summary      Получить группу по ID
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Group UUID"
// @Success      200 {object} GroupResponse
// @Failure      404 {object} httperr.Body
// @Router       /groups/{id} [get]
func (h *Handler) GetGroup(c *fiber.Ctx) error {
	if _, err := h.actorOr401(c); err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_id", "id must be UUID", nil)
	}
	g, err := h.groups.GetByID(c.UserContext(), id)
	if err != nil {
		return httperr.FromDomain(c, err)
	}

	memberships, err := h.groups.ListMembers(c.UserContext(), id)
	if err != nil {
		return httperr.FromDomain(c, err)
	}

	token := bearerToken(c)
	members := make([]MemberInfo, 0, len(memberships))
	var leader *MemberInfo
	for i := range memberships {
		info, err := h.authClient.GetUser(c.UserContext(), token, memberships[i].UserID)
		if err != nil {
			// Один недоступный auth-service/пользователь не должен рушить
			// всю страницу группы — пропускаем и отдаём остальных.
			continue
		}
		mi := MemberInfo{ID: info.ID, FullName: info.FullName, Role: info.Role}
		members = append(members, mi)
		if memberships[i].UserID == g.LeaderID {
			leaderCopy := mi
			leader = &leaderCopy
		}
	}
	// Лидер мог не попасть в список participants (например, ещё не
	// добавлен как member) — на этот случай запрашиваем его отдельно.
	if leader == nil {
		if info, err := h.authClient.GetUser(c.UserContext(), token, g.LeaderID); err == nil {
			leader = &MemberInfo{ID: info.ID, FullName: info.FullName, Role: info.Role}
		}
	}

	resp := GroupWithMembersResponse{
		GroupResponse: toGroupResponse(g),
		Leader:        leader,
		Members:       members,
	}
	return c.JSON(resp)
}

// bearerToken достаёт исходный токен из заголовка запроса — нужен для
// проброса в auth-service при обогащении ответа данными участников.
func bearerToken(c *fiber.Ctx) string {
	raw := c.Get("Authorization")
	return strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
}

// UpdateGroup godoc
// @Summary      Обновить группу
// @Description  teacher/admin или текущий лидер группы
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string true "Group UUID"
// @Param        body body UpdateGroupRequest true "Patch payload"
// @Success      200 {object} GroupResponse
// @Failure      403 {object} httperr.Body
// @Router       /groups/{id} [patch]
func (h *Handler) UpdateGroup(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_id", "id must be UUID", nil)
	}
	var req UpdateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}

	in := usecase.UpdateGroupInput{Name: req.Name, Course: req.Course, Faculty: req.Faculty}
	if req.LeaderID != nil {
		lid, err := uuid.Parse(*req.LeaderID)
		if err != nil {
			return httperr.Send(c, fiber.StatusBadRequest, "invalid_leader_id", "leader_id must be UUID", nil)
		}
		in.LeaderID = &lid
	}
	g, err := h.groups.Update(c.UserContext(), actor, id, in)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toGroupResponse(g))
}

// DeleteGroup godoc
// @Summary      Удалить группу
// @Description  Только teacher/admin
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Group UUID"
// @Success      204
// @Failure      403 {object} httperr.Body
// @Router       /groups/{id} [delete]
func (h *Handler) DeleteGroup(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_id", "id must be UUID", nil)
	}
	if err := h.groups.Delete(c.UserContext(), actor, id); err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ListMembers godoc
// @Summary      Список участников группы
// @Tags         memberships
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Group UUID"
// @Success      200 {array} MembershipResponse
// @Router       /groups/{id}/members [get]
func (h *Handler) ListMembers(c *fiber.Ctx) error {
	if _, err := h.actorOr401(c); err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_id", "id must be UUID", nil)
	}
	ms, err := h.groups.ListMembers(c.UserContext(), id)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	out := make([]MembershipResponse, 0, len(ms))
	for i := range ms {
		out = append(out, toMembershipResponse(&ms[i]))
	}
	return c.JSON(out)
}

// AddMember godoc
// @Summary      Добавить участника
// @Description  Лидер группы или teacher/admin
// @Tags         memberships
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string true "Group UUID"
// @Param        body body AddMemberRequest true "Member payload"
// @Success      201 {object} MembershipResponse
// @Failure      403 {object} httperr.Body
// @Failure      409 {object} httperr.Body
// @Router       /groups/{id}/members [post]
func (h *Handler) AddMember(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_id", "id must be UUID", nil)
	}
	var req AddMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_user_id", "user_id must be UUID", nil)
	}
	role := domain.MembershipRole(req.RoleInGroup)
	if req.RoleInGroup == "" {
		role = domain.MembershipMember
	}
	m, err := h.groups.AddMember(c.UserContext(), actor, groupID, usecase.AddMemberInput{
		UserID: uid, Role: role,
	})
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(toMembershipResponse(m))
}

// RemoveMember godoc
// @Summary      Удалить участника
// @Description  Лидер группы / teacher / admin / сам себя
// @Tags         memberships
// @Produce      json
// @Security     BearerAuth
// @Param        id      path string true "Group UUID"
// @Param        user_id path string true "User UUID"
// @Success      204
// @Failure      403 {object} httperr.Body
// @Failure      404 {object} httperr.Body
// @Router       /groups/{id}/members/{user_id} [delete]
func (h *Handler) RemoveMember(c *fiber.Ctx) error {
	actor, err := h.actorOr401(c)
	if err != nil {
		return err
	}
	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_id", "id must be UUID", nil)
	}
	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_user_id", "user_id must be UUID", nil)
	}
	if err := h.groups.RemoveMember(c.UserContext(), actor, groupID, userID); err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Health / Ready ----------

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

// ===== mappers =====

func toGroupResponse(g *domain.Group) GroupResponse {
	return GroupResponse{
		ID:        g.ID.String(),
		Name:      g.Name,
		Course:    g.Course,
		Faculty:   g.Faculty,
		LeaderID:  g.LeaderID.String(),
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}

func toMembershipResponse(m *domain.Membership) MembershipResponse {
	return MembershipResponse{
		UserID:      m.UserID.String(),
		GroupID:     m.GroupID.String(),
		RoleInGroup: string(m.RoleInGroup),
		JoinedAt:    m.JoinedAt,
	}
}
