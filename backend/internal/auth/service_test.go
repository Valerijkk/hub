package auth

import (
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

// TestVerifyTelegramAuth_ValidSignature verifies that a correctly signed
// Telegram Web App initData is accepted.
func TestVerifyTelegramAuth_ValidSignature(t *testing.T) {
	botToken := "test-bot-token-123"
	authDate := strconv.FormatInt(time.Now().Unix(), 10)

	userJSON := `{"id":123456,"first_name":"John","last_name":"Doe","username":"johndoe"}`

	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("user", userJSON)
	params.Set("query_id", "test-query-id")

	// Compute expected hash.
	hash := computeTelegramHash(t, botToken, params)
	params.Set("hash", hash)

	tgUser, err := VerifyTelegramAuth(botToken, params.Encode())
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

// TestVerifyTelegramAuth_InvalidHash verifies that a tampered hash is rejected.
func TestVerifyTelegramAuth_InvalidHash(t *testing.T) {
	params := url.Values{}
	params.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("user", `{"id":1}`)
	params.Set("hash", "invalid-hash-value")

	_, err := VerifyTelegramAuth("bot-token", params.Encode())
	if err == nil {
		t.Fatal("expected error for invalid hash, got nil")
	}
}

// TestVerifyTelegramAuth_ExpiredAuth verifies that auth_date older than 5 minutes
// is rejected.
func TestVerifyTelegramAuth_ExpiredAuth(t *testing.T) {
	botToken := "test-bot-token-123"
	// auth_date 10 minutes ago
	authDate := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
	userJSON := `{"id":123456,"first_name":"John","last_name":"Doe","username":"johndoe"}`

	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("user", userJSON)

	hash := computeTelegramHash(t, botToken, params)
	params.Set("hash", hash)

	_, err := VerifyTelegramAuth(botToken, params.Encode())
	if err == nil {
		t.Fatal("expected error for expired auth, got nil")
	}
	if err != ErrTelegramAuthExpired {
		t.Fatalf("expected ErrTelegramAuthExpired, got: %v", err)
	}
}

// TestVerifyTelegramAuth_MissingHash ensures missing hash parameter returns error.
func TestVerifyTelegramAuth_MissingHash(t *testing.T) {
	params := url.Values{}
	params.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("user", `{"id":1}`)

	_, err := VerifyTelegramAuth("bot-token", params.Encode())
	if err == nil {
		t.Fatal("expected error for missing hash")
	}
}

// TestGenerateCode checks that generateCode produces a 6-digit string.
func TestGenerateCode(t *testing.T) {
	for i := 0; i < 100; i++ {
		code, err := generateCode()
		if err != nil {
			t.Fatalf("generateCode error: %v", err)
		}
		if len(code) != 6 {
			t.Fatalf("expected 6-digit code, got %q (len=%d)", code, len(code))
		}
		// Ensure it is numeric
		if _, err := strconv.Atoi(code); err != nil {
			t.Fatalf("code %q is not numeric", code)
		}
	}
}

// TestJWTRoundTrip tests that generated access tokens can be parsed back.
func TestJWTRoundTrip(t *testing.T) {
	svc := &AuthService{
		jwtSecret: []byte("test-secret-key"),
		accessTTL: 15 * time.Minute,
	}

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	sessionID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")

	tokenStr, err := svc.generateAccessToken(userID, sessionID)
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

// TestJWTExpired tests that an expired token is rejected.
func TestJWTExpired(t *testing.T) {
	svc := &AuthService{
		jwtSecret: []byte("test-secret-key"),
		accessTTL: -1 * time.Second, // already expired
	}

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	sessionID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")

	tokenStr, err := svc.generateAccessToken(userID, sessionID)
	if err != nil {
		t.Fatalf("generateAccessToken error: %v", err)
	}

	_, err = svc.ParseAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

// TestJWTWrongSecret tests that a token signed with a different secret is rejected.
func TestJWTWrongSecret(t *testing.T) {
	svc1 := &AuthService{
		jwtSecret: []byte("secret-1"),
		accessTTL: 15 * time.Minute,
	}
	svc2 := &AuthService{
		jwtSecret: []byte("secret-2"),
		accessTTL: 15 * time.Minute,
	}

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	sessionID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")

	tokenStr, err := svc1.generateAccessToken(userID, sessionID)
	if err != nil {
		t.Fatalf("generateAccessToken error: %v", err)
	}

	_, err = svc2.ParseAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

// --- helpers ---

// computeTelegramHash computes the expected Telegram HMAC-SHA256 hash for test data.
// params should NOT contain the "hash" key.
func computeTelegramHash(t *testing.T, botToken string, params url.Values) string {
	t.Helper()

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	// Sort
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
