package httperr

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type Body struct {
	Error Detail `json:"error"`
}

type Detail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Send(c *fiber.Ctx, status int, code, msg string) error {
	return c.Status(status).JSON(Body{Error: Detail{Code: code, Message: msg}})
}

// Common helpers.
func Unauthorized(c *fiber.Ctx, msg string) error {
	return Send(c, http.StatusUnauthorized, "unauthorized", msg)
}

func BadGateway(c *fiber.Ctx, msg string) error {
	return Send(c, http.StatusBadGateway, "upstream_error", msg)
}

func NotFound(c *fiber.Ctx, msg string) error {
	return Send(c, http.StatusNotFound, "not_found", msg)
}
