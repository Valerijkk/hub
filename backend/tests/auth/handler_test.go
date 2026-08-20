package auth_test

// Тесты HTTP-обработчика аутентификации: интеграционные и unit-сценарии testframework.
import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"github.com/valerii/hub/backend/internal/auth"
)

type magicLinkHandlerRepo struct {
	auth.Repository
}

func (*magicLinkHandlerRepo) UserExistsByEmail(context.Context, string) (bool, error) {
	return true, nil
}

// testApp создаёт минимальное приложение Fiber с переданным обработчиком, смонтированным на
// указанном method+path, пригодное для unit-тестирования отдельных обработчиков.
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

// TestSendMagicLink_InvalidEmail проверяет отклонение при отправка magic link некорректный email.
func TestSendMagicLink_InvalidEmail(t *testing.T) {
	h := auth.NewHandler(&auth.AuthService{}, false)
	app := testApp(http.MethodPost, "/auth/magic-link", h.SendMagicLink)

	body, _ := json.Marshal(auth.MagicLinkRequest{Email: "not-an-email"})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5 * time.Second)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 400, got %d: %s", resp.StatusCode, string(respBody))
	}
}

// TestSendMagicLink_EmptyBody проверяет сценарий: отправка magic link пустой body.
func TestSendMagicLink_EmptyBody(t *testing.T) {
	h := auth.NewHandler(&auth.AuthService{}, false)
	app := testApp(http.MethodPost, "/auth/magic-link", h.SendMagicLink)

	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5 * time.Second)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// TestVerifyMagicLink_ShortCode проверяет сценарий: verify magic link short code.
func TestVerifyMagicLink_ShortCode(t *testing.T) {
	h := auth.NewHandler(&auth.AuthService{}, false)
	app := testApp(http.MethodPost, "/auth/magic-link/verify", h.VerifyMagicLink)

	body, _ := json.Marshal(auth.VerifyMagicLinkRequest{Email: "test@example.com", Code: "123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5 * time.Second)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// TestGetMe_Unauthenticated проверяет сценарий: получение me unauthenticated.
func TestGetMe_Unauthenticated(t *testing.T) {
	h := auth.NewHandler(&auth.AuthService{}, false)
	app := testApp(http.MethodGet, "/auth/me", h.GetMe)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

	resp, err := app.Test(req, 5 * time.Second)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// TestLogout_Unauthenticated проверяет сценарий: выход unauthenticated.
func TestLogout_Unauthenticated(t *testing.T) {
	h := auth.NewHandler(&auth.AuthService{}, false)
	app := testApp(http.MethodPost, "/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

	resp, err := app.Test(req, 5 * time.Second)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// TestRefreshToken_EmptyBody проверяет сценарий: refresh token пустой body.
func TestRefreshToken_EmptyBody(t *testing.T) {
	h := auth.NewHandler(&auth.AuthService{}, false)
	app := testApp(http.MethodPost, "/auth/refresh", h.RefreshToken)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5 * time.Second)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSendMagicLink_CodeDisclosureFollowsExplicitDevMode(t *testing.T) {
	for _, tc := range []struct {
		name       string
		devMode    bool
		expectCode bool
	}{
		{name: "production without delivery", devMode: false, expectCode: false},
		{name: "development without delivery", devMode: true, expectCode: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })

			magicLink := auth.NewMagicLinkService(rdb, nil, "")
			service := auth.NewAuthService(
				&magicLinkHandlerRepo{},
				rdb,
				nil,
				"unit-test-jwt-secret-at-least-32-bytes",
				time.Minute,
				time.Hour,
				magicLink,
				nil,
			)
			handler := auth.NewHandler(service, tc.devMode)
			app := testApp(http.MethodPost, "/auth/magic-link", handler.SendMagicLink)
			body, err := json.Marshal(auth.MagicLinkRequest{Email: "user@example.test"})
			if err != nil {
				t.Fatalf("marshal request: %v", err)
			}
			req := httptest.NewRequest(
				http.MethodPost,
				"/auth/magic-link",
				bytes.NewReader(body),
			)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, 5*time.Second)
			if err != nil {
				t.Fatalf("test request failed: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				responseBody, _ := io.ReadAll(resp.Body)
				t.Fatalf("expected 200, got %d: %s", resp.StatusCode, responseBody)
			}
			var payload map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if payload["message"] != "Magic link sent" {
				t.Fatalf("unexpected response message: %q", payload["message"])
			}
			code, hasCode := payload["code"]
			if hasCode != tc.expectCode {
				t.Fatalf("code presence = %v, want %v", hasCode, tc.expectCode)
			}
			if tc.expectCode {
				if len(code) != 6 {
					t.Fatalf("code length = %d, want 6", len(code))
				}
				if _, err := strconv.Atoi(code); err != nil {
					t.Fatalf("code is not numeric: %v", err)
				}
			}
		})
	}
}
