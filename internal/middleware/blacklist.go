package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

var blacklist = []string{
	"malware.com",
	"phishing.com",
	"spam.com",
}

func Blacklist(c fiber.Ctx) error {
	body := struct {
		URL string `json:"url"`
	}{}

	if err := c.Bind().JSON(&body); err != nil {
		return c.Next()
	}

	for _, domain := range blacklist {
		if strings.Contains(body.URL, domain) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "url is blacklisted",
			})
		}
	}

	return c.Next()
}
