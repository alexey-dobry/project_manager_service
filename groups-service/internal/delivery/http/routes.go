package httpdelivery

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/gofiber/swagger"

	"github.com/student-pm/groups-service/internal/pkg/jwt"
)

func RegisterRoutes(app *fiber.App, h *Handler, v *jwt.Verifier) {
	app.Get("/health", h.Health)
	app.Get("/ready", h.Ready)
	app.Get("/swagger/*", fiberSwagger.HandlerDefault)

	// все эндпоинты требуют валидный JWT
	groups := app.Group("/groups", AuthRequired(v))
	groups.Post("", h.CreateGroup)
	groups.Get("", h.ListGroups)
	groups.Get("/:id", h.GetGroup)
	groups.Patch("/:id", h.UpdateGroup)
	groups.Delete("/:id", h.DeleteGroup)

	groups.Get("/:id/members", h.ListMembers)
	groups.Post("/:id/members", h.AddMember)
	groups.Delete("/:id/members/:user_id", h.RemoveMember)
}
