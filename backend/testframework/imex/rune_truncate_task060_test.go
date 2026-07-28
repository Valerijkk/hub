package imex_test

// P3-BE-04 / TASK-060: PDF cell truncate по рунам, не по байтам.
import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/imex"
)

// TestTruncateCell_Cyrillic_NotMidRune: «я» = 2 байта; [:47] (нечётно) mid-rune.
// rune truncate(50) → 47 рун + "...".
func TestTruncateCell_Cyrillic_NotMidRune(t *testing.T) {
	s := strings.Repeat("я", 60)
	require.Equal(t, 120, len(s))

	byteCut := s[:47]
	assert.False(t, utf8.ValidString(byteCut), "байтовый cellVal[:47] обязан ломать UTF-8")

	got := imex.ExportTruncateCell(s, 50)
	assert.True(t, utf8.ValidString(got))
	assert.Equal(t, strings.Repeat("я", 47)+"...", got)
	assert.Equal(t, 50, utf8.RuneCountInString(got))
}

// TestTruncateCell_Emoji_KeepsWholeRunes: 📦 = 4 байта; [:47] mid-rune.
func TestTruncateCell_Emoji_KeepsWholeRunes(t *testing.T) {
	s := strings.Repeat("📦", 60)
	require.Equal(t, 240, len(s))

	byteCut := s[:47]
	assert.False(t, utf8.ValidString(byteCut))

	got := imex.ExportTruncateCell(s, 50)
	assert.True(t, utf8.ValidString(got))
	assert.Equal(t, strings.Repeat("📦", 47)+"...", got)
}

func TestTruncateCell_Short_Unchanged(t *testing.T) {
	assert.Equal(t, "короткий", imex.ExportTruncateCell("короткий", 50))
}
