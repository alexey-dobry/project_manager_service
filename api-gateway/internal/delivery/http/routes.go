package httpdelivery

import (
	"github.com/gofiber/fiber/v2"

	"github.com/student-pm/api-gateway/internal/pkg/jwt"
	"github.com/student-pm/api-gateway/internal/proxy"
)

// Upstreams содержит три proxy под backend-сервисы.
type Upstreams struct {
	Auth     *proxy.Proxy
	Groups   *proxy.Proxy
	Projects *proxy.Proxy
}

// RegisterRoutes регистрирует маршруты gateway и проксирующие маршруты на backend.
func RegisterRoutes(app *fiber.App, ups Upstreams, v *jwt.Verifier, swagger fiber.Handler) {
	app.Get("/", root)
	app.Get("/health", health)
	app.Get("/ready", health)
	if swagger != nil {
		app.Get("/swagger", redirectToSwaggerIndex)
		app.Get("/swagger/*", swagger)
	}

	app.Post("/auth/register", ups.Auth.Handler())
	app.Post("/auth/login", ups.Auth.Handler())
	app.Post("/auth/refresh", ups.Auth.Handler())

	authMW := AuthRequired(v)

	app.All("/auth/*", authMW, ups.Auth.Handler())
	app.All("/users", authMW, ups.Auth.Handler())
	app.All("/users/*", authMW, ups.Auth.Handler())

	app.All("/groups", authMW, ups.Groups.Handler())
	app.All("/groups/*", authMW, ups.Groups.Handler())

	app.All("/projects", authMW, ups.Projects.Handler())
	app.All("/projects/*", authMW, ups.Projects.Handler())
	app.All("/tasks", authMW, ups.Projects.Handler())
	app.All("/tasks/*", authMW, ups.Projects.Handler())
}

func root(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"service": "api-gateway",
		"docs":    "/swagger/",
	})
}

func health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "ok"})
}

func redirectToSwaggerIndex(c *fiber.Ctx) error {
	return c.Redirect("/swagger/", fiber.StatusFound)
}
