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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func computeLoginWidgetHash(t *testing.T, botToken string, params url.Values) string {
	t.Helper()
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
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
	h := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, h[:])
	mac.Write([]byte(dataCheckString))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyTelegramAuth_LoginWidget(t *testing.T) {
	botToken := "widget-bot-token"
	authDate := strconv.FormatInt(time.Now().Unix(), 10)
	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("id", "987654")
	params.Set("first_name", "Widget")
	params.Set("last_name", "User")
	params.Set("username", "widgetuser")
	params.Set("hash", computeLoginWidgetHash(t, botToken, params))

	user, err := VerifyTelegramAuth(botToken, params.Encode())
	require.NoError(t, err)
	assert.Equal(t, int64(987654), user.ID)
	assert.Equal(t, "Widget", user.FirstName)
	assert.Equal(t, "User", user.LastName)
	assert.Equal(t, "widgetuser", user.Username)
}

func TestVerifyTelegramAuth_InvalidAuthDate(t *testing.T) {
	params := url.Values{}
	params.Set("auth_date", "not-a-number")
	params.Set("user", `{"id":1}`)
	params.Set("hash", "abc")
	_, err := VerifyTelegramAuth("token", params.Encode())
	require.Error(t, err)
}

func TestVerifyTelegramAuth_MissingAuthDate(t *testing.T) {
	botToken := "widget-bot-token"
	params := url.Values{}
	params.Set("id", "1")
	params.Set("first_name", "A")
	params.Set("hash", computeLoginWidgetHash(t, botToken, params))
	_, err := VerifyTelegramAuth(botToken, params.Encode())
	require.Error(t, err)
}

func TestVerifyTelegramAuth_InvalidUserJSON(t *testing.T) {
	botToken := "test-bot-token-123"
	authDate := strconv.FormatInt(time.Now().Unix(), 10)
	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("user", `{invalid`)
	params.Set("query_id", "q1")
	params.Set("hash", computeTelegramHash(t, botToken, params))
	_, err := VerifyTelegramAuth(botToken, params.Encode())
	require.Error(t, err)
}
