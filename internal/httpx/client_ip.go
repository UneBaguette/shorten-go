package httpx

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

func ClientIP(c fiber.Ctx) string {
	ip := strings.TrimSpace(c.Get("X-Real-IP"))
	if ip == "" {
		ip = c.IP()
	}
	return ip
}