package middleware

import "github.com/gofiber/fiber/v3"

func ApiKey(key string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Get("X-API-Key") != key {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		return c.Next()
	}
}

func DeleteToken(c fiber.Ctx) error {
	token := c.Get("X-Delete-Token")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	return c.Next()
}
