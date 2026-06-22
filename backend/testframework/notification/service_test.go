package notification_test

// Тесты сервиса уведомлений: интеграционные и unit-сценарии testframework.

import (
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/stretchr/testify/assert"
	telegramutil "github.com/valerii/hub/backend/internal/platform/telegram"
)

// TestSubjectToType проверяет маппинг NATS subject в тип уведомления.
func TestSubjectToType(t *testing.T) {
	tests := []struct {
		subject  string
		expected string
	}{
		{"project.abc.issue.assigned", notification.TypeAssigned},
		{"project.abc.issue.moved", notification.TypeStatusChanged},
		{"project.abc.issue.comment_created", notification.TypeCommentAdded},
		{"project.abc.issue.link_created", notification.TypeBlockerAdded},
		{"messenger.channel.abc.message_sent", notification.TypeMessageReceived},
		{"calls.channel.abc.started", notification.TypeCallStarted},
		{"tms.testrun.completed", notification.TypeRunCompleted},
		{"project.abc.issue.created", ""},
		{"unknown.subject", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			result := notification.ExportSubjectToType(tt.subject)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseTimeMinutes проверяет парсинг строки HH:MM в минуты.
func TestParseTimeMinutes(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"00:00", 0, false},
		{"23:59", 1439, false},
		{"08:30", 510, false},
		{"12:00", 720, false},
		{"invalid", 0, true},
		{"25:00", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := notification.ExportParseTimeMinutes(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestEscapeMarkdownV2 проверяет экранирование специальных символов для Telegram.
func TestEscapeMarkdownV2(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello", "Hello"},
		{"Hello_World", "Hello\\_World"},
		{"test*bold*", "test\\*bold\\*"},
		{"(brackets)", "\\(brackets\\)"},
		{"HUB-123", "HUB\\-123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := telegramutil.EscapeMarkdownV2(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatTelegramMessage проверяет форматирование сообщения для Telegram.
func TestFormatTelegramMessage(t *testing.T) {
	t.Run("only title", func(t *testing.T) {
		result := notification.ExportFormatTelegramMessage("Test Title", "")
		assert.Equal(t, "*Test Title*", result)
	})

	t.Run("title and body", func(t *testing.T) {
		result := notification.ExportFormatTelegramMessage("Title", "Body text")
		assert.Equal(t, "*Title*\n\nBody text", result)
	})
}

// TestMatchSuffix проверяет вспомогательную функцию matchSuffix.
func TestMatchSuffix(t *testing.T) {
	tests := []struct {
		parts    []string
		suffix   string
		expected bool
	}{
		{[]string{"project", "abc", "issue", "assigned"}, "issue.assigned", true},
		{[]string{"project", "abc", "issue", "created"}, "issue.assigned", false},
		{[]string{"short"}, "issue.assigned", false},
	}

	for _, tt := range tests {
		t.Run(tt.suffix, func(t *testing.T) {
			result := notification.ExportMatchSuffix(tt.parts, tt.suffix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShouldDispatch_Debounce проверяет дебаунс-логику notification.Consumer.
func TestShouldDispatch_Debounce(t *testing.T) {
	c := notification.ExportNewTestConsumerWithDebounce()

	// Первый вызов — разрешён.
	assert.True(t, notification.ExportShouldDispatch(c, "user1", notification.TypeAssigned, "issue1"))
	// Повторный вызов в пределах 2 секунд — запрещён.
	assert.False(t, notification.ExportShouldDispatch(c, "user1", notification.TypeAssigned, "issue1"))
	// Другой пользователь — разрешён.
	assert.True(t, notification.ExportShouldDispatch(c, "user2", notification.TypeAssigned, "issue1"))
	// Другой тип — разрешён.
	assert.True(t, notification.ExportShouldDispatch(c, "user1", notification.TypeMentioned, "issue1"))
}
