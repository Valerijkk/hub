package audit

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// HTTPAuditEntry — данные HTTP-запроса для журнала аудита.
type HTTPAuditEntry struct {
	Method     string
	Path       string
	Status     int
	DurationMS int64
	RequestID  string
	UserID     *uuid.UUID
	IPAddress  string
	UserAgent  string
}

// stripQuery возвращает путь без query string (ME-03).
func stripQuery(p string) string {
	if i := strings.IndexByte(p, '?'); i >= 0 {
		return p[:i]
	}
	return p
}

// shouldSkipHTTPAudit — health/docs и preflight не пишем в журнал.
func shouldSkipHTTPAudit(method, path string) bool {
	if method == fiber.MethodOptions {
		return true
	}
	switch path {
	case "/api/health", "/api/status", "/api/docs", "/api/docs/doc.json", "/api/webhooks/alertmanager":
		return true
	}
	if strings.HasPrefix(path, "/api/docs/") {
		return true
	}
	return false
}

// HTTPAuditMiddleware записывает HTTP-запросы в журнал аудита через recorder.
// Recorder вызывается асинхронно, чтобы не блокировать ответ клиенту.
func HTTPAuditMiddleware(recorder func(context.Context, HTTPAuditEntry)) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		path := stripQuery(c.Path())
		method := c.Method()
		if shouldSkipHTTPAudit(method, path) {
			return err
		}

		entry := HTTPAuditEntry{
			Method:     method,
			Path:       path,
			Status:     c.Response().StatusCode(),
			DurationMS: time.Since(start).Milliseconds(),
			IPAddress:  c.IP(),
			UserAgent:  c.Get("User-Agent"),
		}
		if rid, ok := c.Locals("requestID").(string); ok {
			entry.RequestID = rid
		}
		if uidStr, ok := c.Locals("userID").(string); ok && uidStr != "" {
			if uid, parseErr := uuid.Parse(uidStr); parseErr == nil {
				entry.UserID = &uid
			}
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			recorder(ctx, entry)
		}()

		return err
	}
}
