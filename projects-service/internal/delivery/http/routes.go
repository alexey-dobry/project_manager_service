package httpdelivery

import (
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/gofiber/swagger"

	"github.com/student-pm/projects-service/internal/pkg/jwt"
)

func RegisterRoutes(app *fiber.App, h *Handler, v *jwt.Verifier) {
	app.Get("/health", h.Health)
	app.Get("/ready", h.Ready)
	app.Get("/swagger/*", fiberSwagger.HandlerDefault)

	// projects
	projects := app.Group("/projects", AuthRequired(v))
	projects.Post("", h.CreateProject)
	projects.Get("", h.ListProjects)
	projects.Get("/:id", h.GetProject)
	projects.Patch("/:id", h.UpdateProject)
	projects.Delete("/:id", h.DeleteProject)
	projects.Get("/:id/stats", h.ProjectStats)

	// tasks (вложены в проект)
	projects.Post("/:id/tasks", h.CreateTask)
	projects.Get("/:id/tasks", h.ListTasks)
	projects.Get("/:id/tasks/:task_id", h.GetTask)
	projects.Patch("/:id/tasks/:task_id", h.UpdateTask)
	projects.Delete("/:id/tasks/:task_id", h.DeleteTask)

	// comments
	tasks := app.Group("/tasks", AuthRequired(v))
	tasks.Post("/:task_id/comments", h.CreateComment)
	tasks.Get("/:task_id/comments", h.ListComments)
	tasks.Delete("/:task_id/comments/:comment_id", h.DeleteComment)
}
