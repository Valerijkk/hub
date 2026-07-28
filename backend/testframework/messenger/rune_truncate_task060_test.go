package messenger_test

// P3-BE-04 / TASK-060: promote-to-issue title truncate по рунам, не по байтам.
import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
)

// TestTruncateRunes_PromoteTitle_Cyrillic: ASCII-префикс + кириллица —
// title[:500] mid-rune; rune truncate(500) — валидные 500 рун.
func TestTruncateRunes_PromoteTitle_Cyrillic(t *testing.T) {
	s := "x" + strings.Repeat("я", 501)
	require.Greater(t, len(s), 500)
	require.Greater(t, utf8.RuneCountInString(s), 500)

	byteCut := s[:500]
	assert.False(t, utf8.ValidString(byteCut), "байтовый title[:500] обязан ломать UTF-8")

	got := messenger.ExportTruncateRunes(s, 500)
	assert.True(t, utf8.ValidString(got))
	assert.Equal(t, 500, utf8.RuneCountInString(got))
	assert.Equal(t, "x"+strings.Repeat("я", 499), got)
}

func TestTruncateRunes_PromoteTitle_Emoji(t *testing.T) {
	s := "x" + strings.Repeat("🚀", 510)
	byteCut := s[:500]
	assert.False(t, utf8.ValidString(byteCut))

	got := messenger.ExportTruncateRunes(s, 500)
	assert.True(t, utf8.ValidString(got))
	assert.Equal(t, 500, utf8.RuneCountInString(got))
}

func TestTruncateRunes_PromoteTitle_Short_Unchanged(t *testing.T) {
	assert.Equal(t, "задача из чата", messenger.ExportTruncateRunes("задача из чата", 500))
}
