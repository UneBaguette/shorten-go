package handler

import (
	"net/url"
	"strings"
	"time"

	"github.com/UneBaguette/shorten-go/internal/httpx"
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

	if c.Method() != "OPTIONS" {
		origin := c.Get("Origin")
		if _, ok := h.allowedOrigins[origin]; !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "forbidden",
			})
		}
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

	ip := httpx.ClientIP(c)

	code := h.store.GenerateCode()
	token := store.GenerateToken()

	shortURL := &model.URL{
		Code:        code,
		Original:    req.URL,
		CreatedAt:   time.Now(),
		IP:          ip,
		DeleteToken: token,
	}

	if err := h.store.Set(shortURL, h.ttl); err != nil {
		if err.Error() == "limit_reached" {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "limit_reached",
			})
		}

		if err.Error() == "daily_limit_reached" {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "daily_limit_reached",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not save url",
		})
	}

	activeLinks, err := h.store.CountByIP(ip)

	if err != nil {
		activeLinks = 0
	}

	dailyCreations, err := h.store.GetCreations(ip)

	if err != nil {
		dailyCreations = 0
	}

	return c.Status(fiber.StatusCreated).JSON(model.ShortenResponse{
		Short:          h.baseURL + "/" + code,
		Code:           code,
		DeleteToken:    token,
		ActiveLinks:    activeLinks,
		DailyCreations: dailyCreations,
		MaxLinks:       store.MaxLinksPerIP,
		MaxDaily:       store.MaxCreationsPerDay,
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

	return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(url.Original)
}

func (h *Handler) Check(c fiber.Ctx) error {
	code := c.Params("code")
	url, err := h.store.Get(code)

	if err != nil || url == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) Delete(c fiber.Ctx) error {
	code := c.Params("code")
	token := c.Get("X-Delete-Token")

	url, err := h.store.Get(code)

	if err != nil || url == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "url not found",
		})
	}

	if url.DeleteToken != token {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	if err := h.store.Delete(code, url.IP); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not delete url",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
