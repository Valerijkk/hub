package notification

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSubjectToType проверяет маппинг NATS subject в тип уведомления.
func TestSubjectToType(t *testing.T) {
	tests := []struct {
		subject  string
		expected string
	}{
		{"project.abc.issue.assigned", TypeAssigned},
		{"project.abc.issue.moved", TypeStatusChanged},
		{"project.abc.issue.comment_created", TypeCommentAdded},
		{"project.abc.issue.link_created", TypeBlockerAdded},
		{"messenger.channel.abc.message_sent", TypeMessageReceived},
		{"calls.channel.abc.started", TypeCallStarted},
		{"tms.testrun.completed", TypeRunCompleted},
		{"project.abc.issue.created", ""},
		{"unknown.subject", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			result := SubjectToType(tt.subject)
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
			result, err := parseTimeMinutes(tt.input)
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
			result := escapeMarkdownV2(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatTelegramMessage проверяет форматирование сообщения для Telegram.
func TestFormatTelegramMessage(t *testing.T) {
	t.Run("only title", func(t *testing.T) {
		result := formatTelegramMessage("Test Title", "")
		assert.Equal(t, "*Test Title*", result)
	})

	t.Run("title and body", func(t *testing.T) {
		result := formatTelegramMessage("Title", "Body text")
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
			result := matchSuffix(tt.parts, tt.suffix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShouldDispatch_Debounce проверяет дебаунс-логику consumer.
func TestShouldDispatch_Debounce(t *testing.T) {
	c := &Consumer{
		debounce: make(map[string]time.Time),
	}

	// Первый вызов — разрешён.
	assert.True(t, c.shouldDispatch("user1", TypeAssigned, "issue1"))
	// Повторный вызов в пределах 2 секунд — запрещён.
	assert.False(t, c.shouldDispatch("user1", TypeAssigned, "issue1"))
	// Другой пользователь — разрешён.
	assert.True(t, c.shouldDispatch("user2", TypeAssigned, "issue1"))
	// Другой тип — разрешён.
	assert.True(t, c.shouldDispatch("user1", TypeMentioned, "issue1"))
}
