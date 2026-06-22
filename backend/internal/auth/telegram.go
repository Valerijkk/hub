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

// VerifyTelegramAuth проверяет подпись Telegram-данных.
// Поддерживает два формата:
//   - Web App initData (user=JSON, secret = HMAC-SHA256("WebAppData", botToken))
//   - Login Widget (id/first_name/... напрямую, secret = SHA256(botToken))
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

	// Сортировка ключей обязательна спецификацией Telegram для data_check_string.
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

	// Определяем формат по наличию поля "user" (Web App) или "id" (Login Widget).
	isWebApp := params.Get("user") != ""

	var secret []byte
	if isWebApp {
		// Web App: secret = HMAC-SHA256("WebAppData", botToken) — отличие от Login Widget.
		secretKey := hmac.New(sha256.New, []byte("WebAppData"))
		secretKey.Write([]byte(botToken))
		secret = secretKey.Sum(nil)
	} else {
		// Login Widget: secret = SHA256(botToken) — иной алгоритм, чем у Web App.
		h := sha256.Sum256([]byte(botToken))
		secret = h[:]
	}

	// hash = HMAC-SHA256(secret, data_check_string) — финальная проверка подписи Telegram.
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(dataCheckString.String()))
	calculatedHash := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(calculatedHash), []byte(receivedHash)) {
		return nil, fmt.Errorf("telegram auth: invalid hash")
	}

	// auth_date старше telegramAuthMaxAge отклоняем — защита от replay initData.
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

	// Web App кладёт user JSON; Login Widget — поля id/first_name напрямую в params.
	var tgUser TelegramUser
	if isWebApp {
		userJSON := params.Get("user")
		if err := json.Unmarshal([]byte(userJSON), &tgUser); err != nil {
			return nil, fmt.Errorf("telegram auth: parse user: %w", err)
		}
	} else {
		// Login Widget: поля напрямую в params.
		idStr := params.Get("id")
		if idStr == "" {
			return nil, fmt.Errorf("telegram auth: missing id")
		}
		tgUser.ID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("telegram auth: invalid id: %w", err)
		}
		tgUser.FirstName = params.Get("first_name")
		tgUser.LastName = params.Get("last_name")
		tgUser.Username = params.Get("username")
	}

	return &tgUser, nil
}

var ErrTelegramAuthExpired = fmt.Errorf("telegram auth: auth_date is too old")
