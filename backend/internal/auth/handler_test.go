package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

// testApp creates a minimal Fiber app with the given handler func mounted at
// the specified method+path, suitable for unit-testing individual handlers.
func testApp(method, path string, handler fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{Immutable: true})
	switch method {
	case http.MethodGet:
		app.Get(path, handler)
	case http.MethodPost:
		app.Post(path, handler)
	case http.MethodPut:
		app.Put(path, handler)
	case http.MethodDelete:
		app.Delete(path, handler)
	}
	return app
}

// TestSendMagicLink_InvalidEmail verifies that a request with an invalid email
// returns 400.
func TestSendMagicLink_InvalidEmail(t *testing.T) {
	h := &Handler{service: &AuthService{}}
	app := testApp(http.MethodPost, "/auth/magic-link", h.SendMagicLink)

	body, _ := json.Marshal(MagicLinkRequest{Email: "not-an-email"})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 400, got %d: %s", resp.StatusCode, string(respBody))
	}
}

// TestSendMagicLink_EmptyBody verifies that an empty request body returns 400.
func TestSendMagicLink_EmptyBody(t *testing.T) {
	h := &Handler{service: &AuthService{}}
	app := testApp(http.MethodPost, "/auth/magic-link", h.SendMagicLink)

	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// TestVerifyMagicLink_ShortCode verifies that a code shorter than 6 chars returns 400.
func TestVerifyMagicLink_ShortCode(t *testing.T) {
	h := &Handler{service: &AuthService{}}
	app := testApp(http.MethodPost, "/auth/magic-link/verify", h.VerifyMagicLink)

	body, _ := json.Marshal(VerifyMagicLinkRequest{Email: "test@example.com", Code: "123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// TestGetMe_Unauthenticated verifies that a request without auth returns 401.
func TestGetMe_Unauthenticated(t *testing.T) {
	h := &Handler{service: &AuthService{}}
	app := testApp(http.MethodGet, "/auth/me", h.GetMe)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// TestLogout_Unauthenticated verifies that logout without auth returns 401.
func TestLogout_Unauthenticated(t *testing.T) {
	h := &Handler{service: &AuthService{}}
	app := testApp(http.MethodPost, "/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// TestRefreshToken_EmptyBody verifies that empty body returns 400.
func TestRefreshToken_EmptyBody(t *testing.T) {
	h := &Handler{service: &AuthService{}}
	app := testApp(http.MethodPost, "/auth/refresh", h.RefreshToken)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
