package middleware

import (
	"regexp"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
)

// maxRequestIDLen — верхняя граница клиентского X-Request-ID (W4-A1-002).
const maxRequestIDLen = 128

// requestIDPattern — безопасный алфавит correlation ID (UUID или [A-Za-z0-9._-]).
var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

// RequestID генерирует уникальный X-Request-ID для каждого запроса и сохраняет в c.Locals.
// Клиентский заголовок принимается только после CopyString + валидации длины/алфавита
// (W4-A1-002); иначе генерируется uuid.NewString() (Pitfall 1: fasthttp buffer reuse).
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		requestID := utils.CopyString(c.Get("X-Request-ID"))
		if !validRequestID(requestID) {
			requestID = uuid.NewString()
		}

		c.Locals("requestID", requestID)
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

// validRequestID — UUID или безопасный ASCII-токен ≤128 рун.
func validRequestID(id string) bool {
	if id == "" {
		return false
	}
	if utf8.RuneCountInString(id) > maxRequestIDLen {
		return false
	}
	if _, err := uuid.Parse(id); err == nil {
		return true
	}
	return requestIDPattern.MatchString(id)
}
