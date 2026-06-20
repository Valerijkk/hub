package audit

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestHTTPAuditMiddleware_SkipsHealth(t *testing.T) {
	var recorded int
	app := fiber.New()
	app.Use(HTTPAuditMiddleware(func(_ context.Context, _ HTTPAuditEntry) {
		recorded++
	}))
	app.Get("/api/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/api/health", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	if recorded != 0 {
		t.Fatalf("expected skip, recorded %d", recorded)
	}
}

func TestHTTPAuditMiddleware_RecordsAPI(t *testing.T) {
	done := make(chan HTTPAuditEntry, 1)
	app := fiber.New()
	app.Use(HTTPAuditMiddleware(func(_ context.Context, e HTTPAuditEntry) {
		done <- e
	}))
	app.Get("/api/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/api/test", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	select {
	case e := <-done:
		if e.Path != "/api/test" || e.Method != fiber.MethodGet {
			t.Fatalf("unexpected entry: %+v", e)
		}
	case <-time.After(time.Second):
		t.Fatal("expected audit entry")
	}
}
