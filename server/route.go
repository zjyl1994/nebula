package server

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(app *fiber.App) {
	app.Get("/api", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
}
