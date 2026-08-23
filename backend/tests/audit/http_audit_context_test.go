package audit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/audit"
)

func TestHTTPAuditMiddleware_SkipsOperationalRoutes(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "options preflight", method: http.MethodOptions, path: "/api/tasks"},
		{name: "health", method: http.MethodGet, path: "/api/health"},
		{name: "health with query", method: http.MethodGet, path: "/api/health?verbose=true"},
		{name: "docs root", method: http.MethodGet, path: "/api/docs"},
		{name: "docs document", method: http.MethodGet, path: "/api/docs/doc.json"},
		{name: "docs subpath with query", method: http.MethodGet, path: "/api/docs/swagger/index.html?theme=dark"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorded := make(chan audit.HTTPAuditEntry, 1)
			app := fiber.New()
			app.Use(audit.HTTPAuditMiddleware(func(_ context.Context, entry audit.HTTPAuditEntry) {
				recorded <- entry
			}))
			app.Use(func(c fiber.Ctx) error {
				return c.SendStatus(http.StatusNoContent)
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
			}

			select {
			case entry := <-recorded:
				t.Fatalf("unexpected audit entry: %+v", entry)
			case <-time.After(100 * time.Millisecond):
			}
		})
	}
}

func TestHTTPAuditMiddleware_RecordsRequestContext(t *testing.T) {
	validUserID := uuid.MustParse("44f8b5c4-2d6a-4c5e-9a03-cf71de69a492")
	tests := []struct {
		name        string
		path        string
		status      int
		requestID   string
		userID      string
		wantUserID  *uuid.UUID
		userAgent   string
		wantPath    string
	}{
		{
			name:       "success strips query and preserves valid user",
			path:       "/api/tasks?include=assignee",
			status:     http.StatusCreated,
			requestID:  "req-audit-created",
			userID:     validUserID.String(),
			wantUserID: &validUserID,
			userAgent:  "HubAuditTest/1.0",
			wantPath:   "/api/tasks",
		},
		{
			name:      "non-2xx omits invalid user",
			path:      "/api/auth/refresh?reason=expired",
			status:    http.StatusUnauthorized,
			requestID: "req-audit-denied",
			userID:    "not-a-uuid",
			userAgent: "HubAuditTest/1.0 denied",
			wantPath:  "/api/auth/refresh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorded := make(chan audit.HTTPAuditEntry, 1)
			var wantIP string

			app := fiber.New()
			app.Use(func(c fiber.Ctx) error {
				wantIP = c.IP()
				c.Locals("requestID", c.Get("X-Test-Request-ID"))
				c.Locals("userID", c.Get("X-Test-User-ID"))
				return c.Next()
			})
			app.Use(audit.HTTPAuditMiddleware(func(_ context.Context, entry audit.HTTPAuditEntry) {
				recorded <- entry
			}))
			app.Use(func(c fiber.Ctx) error {
				return c.SendStatus(tt.status)
			})

			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			req.RemoteAddr = "203.0.113.42:443"
			req.Header.Set("User-Agent", tt.userAgent)
			req.Header.Set("X-Test-Request-ID", tt.requestID)
			req.Header.Set("X-Test-User-ID", tt.userID)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tt.status {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.status)
			}

			var entry audit.HTTPAuditEntry
			select {
			case entry = <-recorded:
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for asynchronous audit entry")
			}

			if entry.Method != http.MethodPost {
				t.Errorf("method = %q, want %q", entry.Method, http.MethodPost)
			}
			if entry.Path != tt.wantPath {
				t.Errorf("path = %q, want %q", entry.Path, tt.wantPath)
			}
			if entry.Status != tt.status {
				t.Errorf("status = %d, want %d", entry.Status, tt.status)
			}
			if entry.RequestID != tt.requestID {
				t.Errorf("request ID = %q, want %q", entry.RequestID, tt.requestID)
			}
			if entry.IPAddress == "" {
				t.Error("IP address was not recorded")
			}
			if entry.IPAddress != wantIP {
				t.Errorf("IP address = %q, want %q", entry.IPAddress, wantIP)
			}
			if entry.UserAgent != tt.userAgent {
				t.Errorf("user agent = %q, want %q", entry.UserAgent, tt.userAgent)
			}
			if tt.wantUserID == nil {
				if entry.UserID != nil {
					t.Errorf("user ID = %s, want nil", entry.UserID)
				}
			} else if entry.UserID == nil || *entry.UserID != *tt.wantUserID {
				t.Errorf("user ID = %v, want %s", entry.UserID, tt.wantUserID)
			}
		})
	}
}
