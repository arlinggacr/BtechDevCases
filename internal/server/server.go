package server

import (
	"github.com/arlinggacr/BtechDevCases/internal/module/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func New(authModule *auth.Module) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: errorHandler})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	authModule.RegisterRoutes(app)
	return app
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if fiberErr, ok := err.(*fiber.Error); ok {
		code = fiberErr.Code
	}
	return c.Status(code).JSON(fiber.Map{"error": err.Error()})
}
