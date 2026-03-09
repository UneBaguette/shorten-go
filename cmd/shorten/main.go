package main

import (
	"log"
	"os"
	"time"

	"github.com/UneBaguette/shorten-go/internal/handler"
	"github.com/UneBaguette/shorten-go/internal/middleware"
	"github.com/UneBaguette/shorten-go/internal/store"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
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

	s, err := store.New(dbPath)
	if err != nil {
		log.Fatal("Could not open database:", err)
	}
	defer s.Close()

	h := handler.New(s, baseURL)

	app := fiber.New()

	app.Post("/shorten", limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
	}), h.Shorten)
	app.Get("/:code", h.Redirect)
	app.Delete("/:code", middleware.ApiKey(apiKey), h.Delete)

	log.Fatal(app.Listen(":" + port))
}
