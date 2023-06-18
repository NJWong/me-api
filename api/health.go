package api

import "github.com/gofiber/fiber/v2"

func AddHealthRoutes(app *fiber.App) {
	apiGroup := app.Group("/api")
	apiGroup.Get("/health", handleHealthCheck)
}

func handleHealthCheck(c *fiber.Ctx) error {
	return c.SendString("OK")
}
