package httpdelivery

import (
	"github.com/gofiber/fiber/v2"

	"github.com/student-pm/api-gateway/internal/pkg/jwt"
	"github.com/student-pm/api-gateway/internal/proxy"
)

// Upstreams — три proxy, каждое под свой backend-сервис.
type Upstreams struct {
	Auth     *proxy.Proxy
	Groups   *proxy.Proxy
	Projects *proxy.Proxy
}

// RegisterRoutes монтирует gateway-маршруты.
//
//   Public (без токена):
//     POST /auth/register, /auth/login, /auth/refresh    → auth-service
//     GET  /, /health, /ready, /swagger/*                → gateway сам
//
//   Protected (Bearer):
//     /auth/logout, /users/*                             → auth-service
//     /groups/*                                          → groups-service
//     /projects/*, /tasks/*                              → projects-service
func RegisterRoutes(app *fiber.App, ups Upstreams, v *jwt.Verifier, swagger fiber.Handler) {
	// gateway own.
	app.Get("/", root)
	app.Get("/health", health)
	app.Get("/ready", health)
	if swagger != nil {
		app.Get("/swagger", redirectToSwaggerIndex)
		app.Get("/swagger/*", swagger)
	}

	// Публичные эндпоинты auth-service.
	app.Post("/auth/register", ups.Auth.Handler())
	app.Post("/auth/login", ups.Auth.Handler())
	app.Post("/auth/refresh", ups.Auth.Handler())

	authMW := AuthRequired(v)

	// auth-service protected
	app.All("/auth/logout", authMW, ups.Auth.Handler())
	app.All("/users", authMW, ups.Auth.Handler())
	app.All("/users/*", authMW, ups.Auth.Handler())

	// groups-service
	app.All("/groups", authMW, ups.Groups.Handler())
	app.All("/groups/*", authMW, ups.Groups.Handler())

	// projects-service: /projects/* и /tasks/*
	app.All("/projects", authMW, ups.Projects.Handler())
	app.All("/projects/*", authMW, ups.Projects.Handler())
	app.All("/tasks", authMW, ups.Projects.Handler())
	app.All("/tasks/*", authMW, ups.Projects.Handler())
}

// ===== gateway own handlers =====

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
