package main

import (
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/UneBaguette/shorten-go/internal/handler"
	"github.com/UneBaguette/shorten-go/internal/middleware"
	"github.com/UneBaguette/shorten-go/internal/store"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file, using system env")
	}

	port := os.Getenv("PORT")
	baseURL := os.Getenv("BASE_URL")
	dbPath := os.Getenv("DB_PATH")
	apiKey := os.Getenv("API_KEY")

	allowedOrigins := make(map[string]struct{})
	rawOrigins := os.Getenv("ALLOWED_ORIGINS")

	if rawOrigins == "" {
		log.Fatal("ALLOWED_ORIGINS is not set")
	}

	for _, o := range strings.Split(rawOrigins, ",") {
		allowedOrigins[strings.TrimSpace(o)] = struct{}{}
	}

	ttl, err := strconv.Atoi(os.Getenv("LINK_TTL"))

	if err != nil {
		ttl = 720 // Default to 30 days
	}

	s, err := store.New(dbPath)

	if err != nil {
		log.Fatal("Could not open database:", err)
	}

	defer s.Close()

	h := handler.New(s, baseURL, time.Duration(ttl)*time.Hour, allowedOrigins)

	app := fiber.New()

	os.MkdirAll("./logs", 0755)

	logFile, err := os.OpenFile("./logs/app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatal(err)
	}

	defer logFile.Close()

	// Log every request to both stdout and a file
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${ip} ${method} ${path} ${status} ${latency}\n",
		Stream: io.MultiWriter(os.Stdout, logFile),
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"https://example.com", "https://www.example.com"},
		AllowMethods: []string{"GET", "POST", "DELETE"},
		AllowHeaders: []string{"Content-Type", "X-API-Key"},
	}))

	// Route definitions
	app.Post("/shorten", limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Hour,
	}), middleware.Blacklist, h.Shorten)
	app.Get("/:code", h.Redirect)
	app.Delete("/:code", middleware.ApiKey(apiKey), h.Delete)

	log.Fatal(app.Listen(":" + port))
}
