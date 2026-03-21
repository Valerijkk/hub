package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const telegramAuthMaxAge = 5 * time.Minute

// TelegramUser представляет данные пользователя, извлечённые из Telegram initData.
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// VerifyTelegramAuth проверяет HMAC-SHA256 подпись Telegram Web App initData.
//
// Steps:
//  1. Parse initData as URL query params.
//  2. Extract and remove the "hash" parameter.
//  3. Sort remaining params alphabetically by key.
//  4. Build data_check_string as "key=value\n..." (without trailing newline).
//  5. secret = HMAC-SHA256("WebAppData", botToken).
//  6. Compute HMAC-SHA256(secret, dataCheckString) and compare with hash.
//  7. Verify auth_date is not older than 5 minutes.
func VerifyTelegramAuth(botToken, initData string) (*TelegramUser, error) {
	params, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("telegram auth: parse init data: %w", err)
	}

	receivedHash := params.Get("hash")
	if receivedHash == "" {
		return nil, fmt.Errorf("telegram auth: missing hash")
	}
	params.Del("hash")

	// Sort keys and build data_check_string.
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var dataCheckString strings.Builder
	for i, k := range keys {
		if i > 0 {
			dataCheckString.WriteByte('\n')
		}
		dataCheckString.WriteString(k)
		dataCheckString.WriteByte('=')
		dataCheckString.WriteString(params.Get(k))
	}

	// secret = HMAC-SHA256("WebAppData", botToken)
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secret := secretKey.Sum(nil)

	// hash = HMAC-SHA256(secret, dataCheckString)
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(dataCheckString.String()))
	calculatedHash := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(calculatedHash), []byte(receivedHash)) {
		return nil, fmt.Errorf("telegram auth: invalid hash")
	}

	// Verify auth_date freshness.
	authDateStr := params.Get("auth_date")
	if authDateStr == "" {
		return nil, fmt.Errorf("telegram auth: missing auth_date")
	}
	authDateUnix, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("telegram auth: invalid auth_date: %w", err)
	}
	authDate := time.Unix(authDateUnix, 0)
	if time.Since(authDate) > telegramAuthMaxAge {
		return nil, ErrTelegramAuthExpired
	}

	// Parse user JSON from "user" param.
	userJSON := params.Get("user")
	if userJSON == "" {
		return nil, fmt.Errorf("telegram auth: missing user data")
	}
	var tgUser TelegramUser
	if err := json.Unmarshal([]byte(userJSON), &tgUser); err != nil {
		return nil, fmt.Errorf("telegram auth: parse user: %w", err)
	}

	return &tgUser, nil
}

var ErrTelegramAuthExpired = fmt.Errorf("telegram auth: auth_date is too old")
