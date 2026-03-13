package httpdelivery

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/gofiber/swagger"

	"github.com/student-pm/auth-service/internal/usecase"
)

// RegisterRoutes монтирует все эндпоинты сервиса на app.
// Sec: публичные — /auth/register, /auth/login, /auth/refresh, /health, /ready, /swagger/*.
// Остальное — за AuthRequired.
func RegisterRoutes(app *fiber.App, h *Handler, tp usecase.TokenProvider) {
	// system
	app.Get("/health", h.Health)
	app.Get("/ready", h.Ready)
	app.Get("/swagger/*", fiberSwagger.HandlerDefault)

	// public auth
	auth := app.Group("/auth")
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)

	// protected
	authProtected := app.Group("/auth", AuthRequired(tp))
	authProtected.Post("/logout", h.Logout)

	users := app.Group("/users", AuthRequired(tp))
	users.Get("/me", h.Me)
	users.Get("/:id", h.GetUserByID)
	users.Patch("/:id", h.UpdateUser)
}
