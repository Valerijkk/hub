package auth_test

// Тесты сервиса аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestVerifyTelegramAuth_ValidSignature проверяет принятие корректно подписанных initData Telegram Web App.
func TestVerifyTelegramAuth_ValidSignature(t *testing.T) {
	botToken := "test-bot-token-123"
	authDate := strconv.FormatInt(time.Now().Unix(), 10)

	userJSON := `{"id":123456,"first_name":"John","last_name":"Doe","username":"johndoe"}`

	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("user", userJSON)
	params.Set("query_id", "test-query-id")

	// Вычисляем ожидаемый hash для проверки подписи.
	hash := computeTelegramHash(t, botToken, params)
	params.Set("hash", hash)

	tgUser, err := auth.VerifyTelegramAuth(botToken, params.Encode())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tgUser.ID != 123456 {
		t.Errorf("expected user ID 123456, got %d", tgUser.ID)
	}
	if tgUser.FirstName != "John" {
		t.Errorf("expected first name John, got %s", tgUser.FirstName)
	}
	if tgUser.LastName != "Doe" {
		t.Errorf("expected last name Doe, got %s", tgUser.LastName)
	}
	if tgUser.Username != "johndoe" {
		t.Errorf("expected username johndoe, got %s", tgUser.Username)
	}
}

// TestVerifyTelegramAuth_InvalidHash проверяет отклонение подписи при подменённом hash.
func TestVerifyTelegramAuth_InvalidHash(t *testing.T) {
	params := url.Values{}
	params.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("user", `{"id":1}`)
	params.Set("hash", "invalid-hash-value")

	_, err := auth.VerifyTelegramAuth("bot-token", params.Encode())
	if err == nil {
		t.Fatal("expected error for invalid hash, got nil")
	}
}

// TestVerifyTelegramAuth_ExpiredAuth проверяет отклонение auth_date старше 5 минут.
func TestVerifyTelegramAuth_ExpiredAuth(t *testing.T) {
	botToken := "test-bot-token-123"
	// auth_date 10 минут назад — должно быть отклонено.
	authDate := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
	userJSON := `{"id":123456,"first_name":"John","last_name":"Doe","username":"johndoe"}`

	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("user", userJSON)

	hash := computeTelegramHash(t, botToken, params)
	params.Set("hash", hash)

	_, err := auth.VerifyTelegramAuth(botToken, params.Encode())
	if err == nil {
		t.Fatal("expected error for expired auth, got nil")
	}
	if err != auth.ErrTelegramAuthExpired {
		t.Fatalf("expected auth.ErrTelegramAuthExpired, got: %v", err)
	}
}

// TestVerifyTelegramAuth_MissingHash проверяет ошибку при отсутствии hash в initData.
func TestVerifyTelegramAuth_MissingHash(t *testing.T) {
	params := url.Values{}
	params.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("user", `{"id":1}`)

	_, err := auth.VerifyTelegramAuth("bot-token", params.Encode())
	if err == nil {
		t.Fatal("expected error for missing hash")
	}
}

// TestGenerateCode проверяет сценарий: generate code.
func TestGenerateCode(t *testing.T) {
	for i := 0; i < 100; i++ {
		code, err := auth.ExportGenerateCode()
		if err != nil {
			t.Fatalf("generateCode error: %v", err)
		}
		if len(code) != 6 {
			t.Fatalf("expected 6-digit code, got %q (len=%d)", code, len(code))
		}
		// Код должен состоять только из цифр.
		if _, err := strconv.Atoi(code); err != nil {
			t.Fatalf("code %q is not numeric", code)
		}
	}
}

// TestJWTRoundTrip проверяет сценарий: jwt round trip.
func TestJWTRoundTrip(t *testing.T) {
	svc := auth.ExportNewSessionServiceForJWTTest("test-secret-key", 15 * time.Minute)

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	sessionID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")

	tokenStr, err := auth.ExportSessionGenerateAccessToken(svc, userID, sessionID)
	if err != nil {
		t.Fatalf("generateAccessToken error: %v", err)
	}

	claims, err := svc.ParseAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("ParseAccessToken error: %v", err)
	}

	if claims.UserID != userID.String() {
		t.Errorf("expected userID %s, got %s", userID, claims.UserID)
	}
	if claims.SessionID != sessionID.String() {
		t.Errorf("expected sessionID %s, got %s", sessionID, claims.SessionID)
	}
}

// TestJWTExpired проверяет отклонение при jwt просроченный.
func TestJWTExpired(t *testing.T) {
	svc := auth.ExportNewSessionServiceForJWTTest("test-secret-key", -1*time.Second)

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	sessionID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")

	tokenStr, err := auth.ExportSessionGenerateAccessToken(svc, userID, sessionID)
	if err != nil {
		t.Fatalf("generateAccessToken error: %v", err)
	}

	_, err = svc.ParseAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

// TestJWTWrongSecret проверяет отклонение при jwt неверный секрет.
func TestJWTWrongSecret(t *testing.T) {
	svc1 := auth.ExportNewSessionServiceForJWTTest("secret-1", 15 * time.Minute)
	svc2 := auth.ExportNewSessionServiceForJWTTest("secret-2", 15 * time.Minute)

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	sessionID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")

	tokenStr, err := auth.ExportSessionGenerateAccessToken(svc1, userID, sessionID)
	if err != nil {
		t.Fatalf("generateAccessToken error: %v", err)
	}

	_, err = svc2.ParseAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

// --- Вспомогательные функции ---

// computeTelegramHash вычисляет ожидаемый хеш Telegram HMAC-SHA256 для тестовых данных.
// params НЕ должен содержать ключ "hash".
func computeTelegramHash(t *testing.T, botToken string, params url.Values) string {
	t.Helper()

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	// Сортируем ключи параметров (требование Telegram).
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	dataCheckString := ""
	for i, k := range keys {
		if i > 0 {
			dataCheckString += "\n"
		}
		dataCheckString += fmt.Sprintf("%s=%s", k, params.Get(k))
	}

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secret := secretKey.Sum(nil)

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(dataCheckString))
	return hex.EncodeToString(h.Sum(nil))
}
