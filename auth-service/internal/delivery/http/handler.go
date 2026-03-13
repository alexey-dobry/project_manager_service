package httpdelivery

import (
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

	in := usecase.UpdateUserInput{FullName: req.FullName}
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
	// Здесь в реальной системе пингуем зависимости (БД и т.п.).
	return c.JSON(MessageResponse{Message: "ready"})
}

// ===== mappers =====

func toUserResponse(u *domain.User) UserResponse {
	r := UserResponse{
		ID:        u.ID.String(),
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
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
