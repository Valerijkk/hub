package marketplace_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/marketplace"
)

func TestDisplayName_PriorityOrder(t *testing.T) {
	assert.Equal(t, "Display", displayName(&ent.User{DisplayName: "Display", Name: "Name", Email: "e@hub.local"}))
	assert.Equal(t, "Name", displayName(&ent.User{Name: "Name", Email: "e@hub.local"}))
	// W7-A5-005: публичный fallback — не email.
	assert.Equal(t, "Пользователь", displayName(&ent.User{Email: "e@hub.local"}))
	assert.Equal(t, "e@hub.local", marketplace.ExportDisplayNamePrivate(&ent.User{Email: "e@hub.local"}))
}

func TestFormatMinorUnitsAndItoa(t *testing.T) {
	assert.Equal(t, "10", formatMinorUnits(1000))
	assert.Equal(t, "10.05", formatMinorUnits(1005))
	assert.Equal(t, "0", itoa(0))
	assert.Equal(t, "-42", itoa(-42))
	assert.Equal(t, "05", pad2(5))
}

func TestIsValidCategory_RejectUnknown(t *testing.T) {
	assert.False(t, isValidCategory(ListingCategory("invalid")))
	assert.True(t, isValidCategory(CategoryBooks))
}
