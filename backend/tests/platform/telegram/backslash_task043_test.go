//go:build hubtest

package telegram_test

import (
	"testing"

	"github.com/valerii/hub/backend/internal/platform/telegram"
)

// TestEscapeMarkdownV2_BackslashFirst_W4A4006 — `\` экранируется первым.
func TestEscapeMarkdownV2_BackslashFirst_W4A4006(t *testing.T) {
	got := telegram.EscapeMarkdownV2(`a\b`)
	want := `a\\b`
	if got != want {
		t.Fatalf("EscapeMarkdownV2(%q) = %q, want %q", `a\b`, got, want)
	}
}
