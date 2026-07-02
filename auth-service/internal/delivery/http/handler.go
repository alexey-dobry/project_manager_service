package httpdelivery

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/student-pm/auth-service/internal/domain"
	httperr "github.com/student-pm/auth-service/internal/pkg/errors"
	"github.com/student-pm/auth-service/internal/pkg/validator"
	"github.com/student-pm/auth-service/internal/usecase"
)

// Handler агрегирует HTTP-эндпоинты сервиса.
type Handler struct {
	auth      *usecase.AuthService
	validator *validator.Validator
}

func NewHandler(auth *usecase.AuthService, v *validator.Validator) *Handler {
	return &Handler{auth: auth, validator: v}
}

// Register godoc
// @Summary      Регистрация пользователя
// @Description  Создаёт пользователя и сразу возвращает access+refresh токены
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RegisterRequest true "Данные регистрации"
// @Success      201 {object} AuthResponse
// @Failure      400 {object} httperr.Body
// @Failure      409 {object} httperr.Body
// @Router       /auth/register [post]
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}
	res, err := h.auth.Register(c.UserContext(), usecase.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     domain.Role(req.Role),
	})
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(toAuthResponse(res))
}

// Login godoc
// @Summary      Вход
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body LoginRequest true "Email и пароль"
// @Success      200 {object} AuthResponse
// @Failure      401 {object} httperr.Body
// @Router       /auth/login [post]
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}
	res, err := h.auth.Login(c.UserContext(), usecase.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: c.Get("User-Agent"),
		IPAddress: c.IP(),
	})
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toAuthResponse(res))
}

// Refresh godoc
// @Summary      Обновление токенов
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RefreshRequest true "Refresh-токен"
// @Success      200 {object} AuthResponse
// @Failure      401 {object} httperr.Body
// @Router       /auth/refresh [post]
func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}
	res, err := h.auth.Refresh(c.UserContext(), req.RefreshToken)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toAuthResponse(res))
}

// Logout godoc
// @Summary      Выход (ревокация всех refresh-токенов)
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} MessageResponse
// @Failure      401 {object} httperr.Body
// @Router       /auth/logout [post]
func (h *Handler) Logout(c *fiber.Ctx) error {
	uid, ok := userIDFrom(c)
	if !ok {
		return httperr.FromDomain(c, domain.ErrInvalidToken)
	}
	if err := h.auth.Logout(c.UserContext(), uid); err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(MessageResponse{Message: "ok"})
}

// Me godoc
// @Summary      Текущий пользователь
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} UserResponse
// @Failure      401 {object} httperr.Body
// @Router       /users/me [get]
func (h *Handler) Me(c *fiber.Ctx) error {
	uid, ok := userIDFrom(c)
	if !ok {
		return httperr.FromDomain(c, domain.ErrInvalidToken)
	}
	u, err := h.auth.GetByID(c.UserContext(), uid)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toUserResponse(u))
}

// GetUserByID godoc
// @Summary      Пользователь по ID
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User UUID"
// @Success      200 {object} UserResponse
// @Failure      404 {object} httperr.Body
// @Router       /users/{id} [get]
func (h *Handler) GetUserByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_id", "id must be UUID", nil)
	}
	u, err := h.auth.GetByID(c.UserContext(), id)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toUserResponse(u))
}

// SearchUsersByEmail godoc
// @Summary      Найти пользователей по email
// @Description  Точное совпадение по email. Используется, например, при
// @Description  добавлении участника в группу: находят пользователя по
// @Description  почте, дальше вызывают POST /groups/{id}/members с его ID.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        email query string true "Email пользователя"
// @Success      200 {array} UserResponse
// @Router       /users/search [get]
func (h *Handler) SearchUsersByEmail(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		return httperr.Send(c, fiber.StatusBadRequest, "email_required", "email query param is required", nil)
	}
	u, err := h.auth.FindByEmail(c.UserContext(), email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// Пустой результат — это не ошибка поиска, а "ничего не нашли".
			return c.JSON([]UserResponse{})
		}
		return httperr.FromDomain(c, err)
	}
	return c.JSON([]UserResponse{toUserResponse(u)})
}

// UpdateUser godoc
// @Summary      Обновление пользователя (PATCH)
// @Description  Менять role может только admin. Обычный пользователь может править только себя.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string true "User UUID"
// @Param        body body UpdateUserRequest true "Поля для обновления"
// @Success      200 {object} UserResponse
// @Failure      403 {object} httperr.Body
// @Router       /users/{id} [patch]
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "invalid_id", "id must be UUID", nil)
	}
	return h.applyUserUpdate(c, targetID)
}

// UpdateMe godoc
// @Summary      Обновление собственного профиля
// @Description  Роль сменить нельзя — только admin через /users/{id}.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body UpdateUserRequest true "Поля для обновления"
// @Success      200 {object} UserResponse
// @Router       /auth/me [patch]
func (h *Handler) UpdateMe(c *fiber.Ctx) error {
	currentID, ok := userIDFrom(c)
	if !ok {
		return httperr.FromDomain(c, domain.ErrInvalidToken)
	}
	return h.applyUserUpdate(c, currentID)
}

// applyUserUpdate — общая логика для UpdateUser и UpdateMe.
func (h *Handler) applyUserUpdate(c *fiber.Ctx, targetID uuid.UUID) error {
	currentID, _ := userIDFrom(c)
	currentRole, _ := roleFrom(c)

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}

	// RBAC: чужой профиль можно править только админу.
	if targetID != currentID && currentRole != domain.RoleAdmin {
		return httperr.FromDomain(c, domain.ErrForbidden)
	}
	// Менять роль вправе только админ.
	if req.Role != nil && currentRole != domain.RoleAdmin {
		return httperr.FromDomain(c, domain.ErrForbidden)
	}

	in := usecase.UpdateUserInput{FullName: req.FullName, Department: req.Department}
	if req.GroupID != nil {
		gid, err := uuid.Parse(*req.GroupID)
		if err != nil {
			return httperr.Send(c, fiber.StatusBadRequest, "invalid_group_id", "group_id must be UUID", nil)
		}
		in.GroupID = &gid
	}
	if req.Role != nil {
		r := domain.Role(*req.Role)
		in.Role = &r
	}

	u, err := h.auth.Update(c.UserContext(), targetID, in)
	if err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(toUserResponse(u))
}

// ChangePassword godoc
// @Summary      Смена пароля текущего пользователя
// @Description  Требует текущий пароль. Ревокирует все refresh-токены пользователя.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body ChangePasswordRequest true "Текущий и новый пароль"
// @Success      200 {object} MessageResponse
// @Failure      401 {object} httperr.Body
// @Router       /auth/change-password [post]
func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	currentID, ok := userIDFrom(c)
	if !ok {
		return httperr.FromDomain(c, domain.ErrInvalidToken)
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "bad_request", "invalid JSON body", nil)
	}
	if details, err := h.validator.Validate(req); err != nil {
		return httperr.Send(c, fiber.StatusBadRequest, "validation_failed", "validation failed", details)
	}

	if err := h.auth.ChangePassword(c.UserContext(), currentID, req.CurrentPassword, req.NewPassword); err != nil {
		return httperr.FromDomain(c, err)
	}
	return c.JSON(MessageResponse{Message: "password changed"})
}

// Health godoc
// @Summary      Health-check
// @Tags         system
// @Produce      json
// @Success      200 {object} MessageResponse
// @Router       /health [get]
func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(MessageResponse{Message: "ok"})
}

// Ready godoc
// @Summary      Readiness-check
// @Tags         system
// @Produce      json
// @Success      200 {object} MessageResponse
// @Router       /ready [get]
func (h *Handler) Ready(c *fiber.Ctx) error {
	return c.JSON(MessageResponse{Message: "ready"})
}

// ===== mappers =====

func toUserResponse(u *domain.User) UserResponse {
	r := UserResponse{
		ID:         u.ID.String(),
		Email:      u.Email,
		FullName:   u.FullName,
		Department: u.Department,
		Role:       string(u.Role),
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
	if u.GroupID != nil {
		s := u.GroupID.String()
		r.GroupID = &s
	}
	return r
}

func toAuthResponse(r *usecase.AuthResult) AuthResponse {
	return AuthResponse{
		User:             toUserResponse(r.User),
		AccessToken:      r.Tokens.AccessToken,
		AccessExpiresAt:  r.Tokens.AccessExpiresAt,
		RefreshToken:     r.Tokens.RefreshToken,
		RefreshExpiresAt: r.Tokens.RefreshExpiresAt,
	}
}
