package search_test

// P3-BE-04 / TASK-060: snippet сообщений truncate по рунам, не по байтам.
import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/search"
)

// TestTruncateRunes_Cyrillic_NotMidRune: ASCII-префикс сдвигает границу —
// s[:100] попадает внутрь «я» (2 байта) → invalid UTF-8; rune helper — ok.
func TestTruncateRunes_Cyrillic_NotMidRune(t *testing.T) {
	s := "x" + strings.Repeat("я", 120)
	require.Greater(t, len(s), 100)
	require.Greater(t, utf8.RuneCountInString(s), 100)

	byteCut := s[:100]
	assert.False(t, utf8.ValidString(byteCut), "байтовый s[:100] обязан ломать UTF-8 на кириллице")

	got := search.ExportTruncateRunes(s, 100)
	assert.True(t, utf8.ValidString(got))
	assert.True(t, strings.HasSuffix(got, "..."))
	assert.Equal(t, 100, utf8.RuneCountInString(strings.TrimSuffix(got, "...")))
}

// TestTruncateRunes_Emoji_KeepsWholeRunes: ASCII + emoji — байтовый [:100]
// режет mid-rune; rune truncate(30) сохраняет целые 🔥.
func TestTruncateRunes_Emoji_KeepsWholeRunes(t *testing.T) {
	s := "x" + strings.Repeat("🔥", 40)
	require.Greater(t, len(s), 100)

	byteCut := s[:100]
	assert.False(t, utf8.ValidString(byteCut), "байтовый s[:100] обязан ломать UTF-8 на emoji")

	got := search.ExportTruncateRunes(s, 30)
	assert.True(t, utf8.ValidString(got))
	assert.Equal(t, "x"+strings.Repeat("🔥", 29)+"...", got)
}

func TestTruncateRunes_ShortString_Unchanged(t *testing.T) {
	assert.Equal(t, "привет", search.ExportTruncateRunes("привет", 100))
	assert.Equal(t, "🔥🔥", search.ExportTruncateRunes("🔥🔥", 10))
}
