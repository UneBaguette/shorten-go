package handler

import (
	"net/url"
	"strings"
	"time"

	"github.com/UneBaguette/shorten-go/internal/model"
	"github.com/UneBaguette/shorten-go/internal/store"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	store          *store.Store
	baseURL        string
	ttl            time.Duration
	allowedOrigins map[string]struct{}
}

func New(store *store.Store, baseURL string, ttl time.Duration, allowedOrigins map[string]struct{}) *Handler {
	return &Handler{store: store, baseURL: baseURL, ttl: ttl, allowedOrigins: allowedOrigins}
}

func (h *Handler) Shorten(c fiber.Ctx) error {

	origin := c.Get("Origin")

	if _, ok := h.allowedOrigins[origin]; !ok {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden",
		})
	}
	req := new(model.ShortenRequest)

	if err := c.Bind().JSON(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request",
		})
	}

	if req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "url is required",
		})
	}

	parsed, err := url.Parse(req.URL)

	if err != nil || parsed.Host == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid url",
		})
	}

	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "url must start with http:// or https://",
		})
	}

	code := h.store.GenerateCode()

	url := &model.URL{
		Code:     code,
		Original: req.URL,
	}

	if err := h.store.Set(url, h.ttl); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not save url",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model.ShortenResponse{
		Short: h.baseURL + "/" + code,
		Code:  code,
	})
}

func (h *Handler) Redirect(c fiber.Ctx) error {
	code := c.Params("code")

	url, err := h.store.Get(code)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal error",
		})
	}

	if url == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "url not found",
		})
	}

	return c.Redirect().To(url.Original)
}

func (h *Handler) Delete(c fiber.Ctx) error {
	code := c.Params("code")

	if err := h.store.Delete(code); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not delete url",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
