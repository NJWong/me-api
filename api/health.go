package api

import "github.com/gofiber/fiber/v2"

func AddHealthEndpoints(app *fiber.App) {
	app.Get("/health", handleHealthCheck)
}

func handleHealthCheck(c *fiber.Ctx) error {
	return c.SendString("OK")
}
