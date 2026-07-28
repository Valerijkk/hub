package auth_test

// Тесты сервиса аутентификации: интеграционные и unit-сценарии testframework.
import (
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/auth"
)

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
	svc := auth.ExportNewSessionServiceForJWTTest("test-secret-key", 15*time.Minute)

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
	svc1 := auth.ExportNewSessionServiceForJWTTest("secret-1", 15*time.Minute)
	svc2 := auth.ExportNewSessionServiceForJWTTest("secret-2", 15*time.Minute)

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
