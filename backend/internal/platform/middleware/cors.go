package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// CORS returns a configured CORS middleware.
// In production, set CORS_ORIGINS env var (comma-separated).
// Defaults to localhost dev server.
func CORS() fiber.Handler {
	origins := []string{"http://localhost:5173", "http://localhost:8080"}
	if env := os.Getenv("CORS_ORIGINS"); env != "" {
		origins = strings.Split(env, ",")
	}
	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-API-Key", "X-Request-ID"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	})
}
