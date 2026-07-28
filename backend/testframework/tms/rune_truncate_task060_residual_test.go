package tms_test

// P3-BE-04 / W2-A2-019 / TASK-060 residual: normalizeErrorPattern по рунам.
import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/tms"
)

func TestNormalizeErrorPattern_Cyrillic_NotMidRune(t *testing.T) {
	msg := strings.Repeat("я", 120)
	require.Equal(t, 240, len(msg))

	// 99 байт — mid-rune (каждая «я» = 2 байта).
	byteCut := msg[:99]
	assert.False(t, utf8.ValidString(byteCut))

	got := tms.ExportNormalizeErrorPattern(msg)
	assert.True(t, utf8.ValidString(got))
	assert.Equal(t, 100, utf8.RuneCountInString(got))
	assert.Equal(t, strings.Repeat("я", 100), got)
}

func TestNormalizeErrorPattern_Short_Unchanged(t *testing.T) {
	assert.Equal(t, "короткая", tms.ExportNormalizeErrorPattern("короткая"))
}
