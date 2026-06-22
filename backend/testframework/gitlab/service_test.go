package gitlab_test
import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valerii/hub/backend/internal/gitlab"
)

func TestIssueKeyRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single key", "Fixed ABC-123 in this commit", []string{"ABC-123"}},
		{"multiple keys", "PROJ-1 and PROJ-22 are done", []string{"PROJ-1", "PROJ-22"}},
		{"no keys", "no issue keys here", nil},
		{"key at boundaries", "DEMO-42", []string{"DEMO-42"}},
		{"lowercase ignored", "abc-123 should not match", nil},
		{"mixed text", "See HUB-99: refactored login for AUTH-5", []string{"HUB-99", "AUTH-5"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gitlab.ExportIssueKeyRegex().FindAllString(tt.input, -1)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "ABC-123-Hello World", "abc-123-hello-world"},
		{"special chars", "ABC-1 Fix: login & signup!", "abc-1-fix-login-signup"},
		{"trailing dashes", "test---", "test"},
		{"unicode letters", "DEMO-5-Telegram-авторизация", "demo-5-telegram-авторизация"},
		{"empty string", "", ""},
		{"only special chars", "!@#$%", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gitlab.ExportSlugify(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"no truncation needed", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncated", "hello world", 5, "hello"},
		{"empty string", "", 5, ""},
		{"zero max", "hello", 0, ""},
		{"unicode", "Привет мир", 6, "Привет"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gitlab.ExportTruncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}
