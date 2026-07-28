package marketplace_test

// P3-BE-04 / TASK-060: лимиты title/description по рунам, не по байтам.
import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/marketplace"
)

// TestValidateListingInput_CyrillicTitle_RuneLimit: 150 «я» = 300 байт, 150 рун.
// Старый len(title) > 200 отклонял; по рунам title ≤ 200 — ok.
func TestValidateListingInput_CyrillicTitle_RuneLimit(t *testing.T) {
	title := strings.Repeat("я", 150)
	require.Equal(t, 300, len(title))
	require.Equal(t, 150, utf8.RuneCountInString(title))
	require.Greater(t, len(title), 200, "байтовая проверка ошибочно отклонила бы этот title")

	err := marketplace.ExportValidateListingInput(title, "описание товара", marketplace.CategoryBooks, nil, "")
	assert.NoError(t, err, "150 рун кириллицы должны пройти лимит 200")
}

// TestValidateListingInput_CyrillicTitle_OverRuneLimit: 201 руна → ErrInvalidInput.
func TestValidateListingInput_CyrillicTitle_OverRuneLimit(t *testing.T) {
	title := strings.Repeat("я", 201)
	require.Equal(t, 201, utf8.RuneCountInString(title))

	err := marketplace.ExportValidateListingInput(title, "описание", marketplace.CategoryBooks, nil, "")
	assert.ErrorIs(t, err, marketplace.ErrInvalidInput)
}

// TestValidateListingInput_EmojiDescription_RuneLimit: emoji считаются по рунам.
func TestValidateListingInput_EmojiDescription_RuneLimit(t *testing.T) {
	// 100 emoji = 400 байт, 100 рун — должно пройти description ≤ 2000.
	desc := strings.Repeat("🔥", 100)
	require.Equal(t, 400, len(desc))
	require.Equal(t, 100, utf8.RuneCountInString(desc))

	err := marketplace.ExportValidateListingInput("title", desc, marketplace.CategoryBooks, nil, "")
	assert.NoError(t, err)

	over := strings.Repeat("🔥", 2001)
	err = marketplace.ExportValidateListingInput("title", over, marketplace.CategoryBooks, nil, "")
	assert.ErrorIs(t, err, marketplace.ErrInvalidInput)
}
