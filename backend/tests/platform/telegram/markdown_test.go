package telegram_test
import (
	"github.com/valerii/hub/backend/internal/platform/telegram"
	"testing")

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
		{"a.b.c", "a\\.b\\.c"},
		{"all_special: _*[]()~`>#+-=|{}.!",
			"all\\_special: \\_\\*\\[\\]\\(\\)\\~\\`\\>\\#\\+\\-\\=\\|\\{\\}\\.\\!"},
		{`a\b`, `a\\b`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := telegram.EscapeMarkdownV2(tt.input)
			if got != tt.expected {
				t.Fatalf("telegram.EscapeMarkdownV2(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
