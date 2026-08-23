package auth_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// forbiddenTelegramLoginMarkers — символы/префиксы удалённого Telegram-login surface (ADR-AUTH-01).
// Notification bind живёт в internal/notification и здесь не сканируется.
var forbiddenTelegramLoginMarkers = []string{
	"tglogin:",
	"VerifyTelegramAuth",
	"TelegramBotAuth",
	"handleLogin",
	"GetUserByTelegramID",
	"NewTelegramBotAuth",
	"ExportTelegramHandleLogin",
	"ErrTelegramAuthExpired",
	"ErrTelegramVerifyLimited",
}

// TestAuth_NoTelegramLoginSurface_Task102 — denial-регрессия TASK-102 / A2W6-A2-001..003 + A1-001/002/005/006:
// в prod-пакете auth нет login-бота, Login Widget crypto и tglogin:-ключей.
//
// Мутация: вернуть VerifyTelegramAuth / tglogin: / TelegramBotAuth в internal/auth → тест красный.
func TestAuth_NoTelegramLoginSurface_Task102(t *testing.T) {
	authDir := filepath.Join("..", "..", "internal", "auth")
	entries, err := os.ReadDir(authDir)
	require.NoError(t, err)

	var hits []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		// login-файлы должны отсутствовать целиком
		switch e.Name() {
		case "telegram.go", "telegram_bot.go":
			hits = append(hits, e.Name()+": file must not exist (ADR-AUTH-01)")
			continue
		}
		data, err := os.ReadFile(filepath.Join(authDir, e.Name()))
		require.NoError(t, err)
		src := string(data)
		for _, marker := range forbiddenTelegramLoginMarkers {
			if strings.Contains(src, marker) {
				hits = append(hits, e.Name()+": contains "+marker)
			}
		}
	}
	require.Empty(t, hits, "Telegram login surface residual in auth:\n%s", strings.Join(hits, "\n"))
}

// TestAuth_NoPositiveTelegramLoginTests_Task102 — в testframework/auth нет позитивных
// HandleLogin/VerifyTelegramAuth/tglogin assert'ов (A2W6-A2-001).
func TestAuth_NoPositiveTelegramLoginTests_Task102(t *testing.T) {
	testDir := "."
	var hits []string
	err := filepath.WalkDir(testDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		// этот файл сам перечисляет маркеры — пропускаем
		if filepath.Base(path) == "telegram_login_removed_task102_test.go" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		src := string(data)
		for _, marker := range []string{
			"VerifyTelegramAuth",
			"TelegramBotAuth",
			"ExportTelegramHandleLogin",
			"HandleLogin_StoresCode",
			"tglogin:",
			"ExportTGCodePrefix",
		} {
			if strings.Contains(src, marker) {
				hits = append(hits, path+": contains "+marker)
			}
		}
		return nil
	})
	require.NoError(t, err)
	require.Empty(t, hits, "positive Telegram login tests residual:\n%s", strings.Join(hits, "\n"))
}
