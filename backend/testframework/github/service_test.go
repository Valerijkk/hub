package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIssueKeyRegex проверяет regex для поиска ключей задач в тексте.
func TestIssueKeyRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "одиночный ключ в сообщении коммита",
			input:    "fix: resolve login issue HUB-123",
			expected: []string{"HUB-123"},
		},
		{
			name:     "несколько ключей",
			input:    "feat: HUB-1 and MOBILE-42 support",
			expected: []string{"HUB-1", "MOBILE-42"},
		},
		{
			name:     "ключ в начале строки",
			input:    "HUB-99 fix bug",
			expected: []string{"HUB-99"},
		},
		{
			name:     "нет ключей",
			input:    "regular commit message without keys",
			expected: nil,
		},
		{
			name:     "ключ с нижним регистром -- не совпадает",
			input:    "hub-123 should not match",
			expected: nil,
		},
		{
			name:     "граница слова -- ключ в скобках",
			input:    "(HUB-5) fix issue",
			expected: []string{"HUB-5"},
		},
		{
			name:     "граница слова -- ключ после двоеточия",
			input:    "fix: HUB-10: description",
			expected: []string{"HUB-10"},
		},
		{
			name:     "дублирование ключей",
			input:    "HUB-1 and HUB-1 again",
			expected: []string{"HUB-1", "HUB-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := github.ExportIssueKeyRegex().FindAllString(tt.input, -1)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateWebhookSignature проверяет HMAC-SHA256 валидацию подписи.
func TestValidateWebhookSignature(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"action":"opened"}`)

	t.Run("корректная подпись", func(t *testing.T) {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		result := github.ValidateWebhookSignature(payload, sig, secret)
		assert.True(t, result)
	})

	t.Run("некорректная подпись", func(t *testing.T) {
		result := github.ValidateWebhookSignature(payload, "sha256=0000000000000000000000000000000000000000000000000000000000000000", secret)
		assert.False(t, result)
	})

	t.Run("отсутствие префикса sha256=", func(t *testing.T) {
		result := github.ValidateWebhookSignature(payload, "invalid-format", secret)
		assert.False(t, result)
	})

	t.Run("пустая подпись", func(t *testing.T) {
		result := github.ValidateWebhookSignature(payload, "", secret)
		assert.False(t, result)
	})
}

// TestSlugify проверяет генерацию URL-безопасных имён веток.
func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "обычный ключ + заголовок",
			input:    "HUB-123-Fix Login Bug",
			expected: "hub-123-fix-login-bug",
		},
		{
			name:     "специальные символы",
			input:    "HUB-1: Add feature! (v2)",
			expected: "hub-1-add-feature-v2",
		},
		{
			name:     "множественные пробелы",
			input:    "HUB-5   extra   spaces",
			expected: "hub-5-extra-spaces",
		},
		{
			name:     "кириллица",
			input:    "HUB-10-Исправить баг",
			expected: "hub-10-исправить-баг",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := github.ExportSlugify(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTruncate проверяет обрезку строк.
func TestTruncate(t *testing.T) {
	t.Run("строка короче лимита", func(t *testing.T) {
		result := github.ExportTruncate("short", 10)
		assert.Equal(t, "short", result)
	})

	t.Run("строка длиннее лимита", func(t *testing.T) {
		result := github.ExportTruncate("this is a long string", 10)
		assert.Equal(t, "this is a ", result)
		require.Len(t, []rune(result), 10)
	})

	t.Run("строка равна лимиту", func(t *testing.T) {
		result := github.ExportTruncate("exact", 5)
		assert.Equal(t, "exact", result)
	})
}
