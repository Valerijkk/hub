//go:build hubtest

package marketplace_test

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/marketplace"
)

// TestListApproved_ReviewNoteHidden_A3W75001 — публичная витрина не светит admin note.
// Mutation: снова копировать ReviewNote в toListingResponseOpts(allowEmail=false) → FAIL.
func TestListApproved_ReviewNoteHidden_A3W75001(t *testing.T) {
	svc, adminID, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)
	note := "internal: PII of seller phone +1-555"
	approved, err := svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{
		Version:    created.Version,
		ReviewNote: note,
	})
	require.NoError(t, err)
	assert.Equal(t, note, approved.ReviewNote, "admin approve response keeps note (private)")

	public, err := svc.ListApproved(ctx, "", 1, 10)
	require.NoError(t, err)
	require.Len(t, public.Items, 1)
	assert.Empty(t, public.Items[0].ReviewNote)
	assert.Empty(t, public.Items[0].ReviewerName)
	assert.Nil(t, public.Items[0].ReviewerID)

	mine, err := svc.ListMine(ctx, memberID, 1, 10)
	require.NoError(t, err)
	require.Len(t, mine.Items, 1)
	assert.Equal(t, note, mine.Items[0].ReviewNote, "ListMine (private) keeps review_note")
}

// TestCurrencyNormalize_A3W75003 — "USD "/"rub" → USD/RUB до Save; invalid → ErrInvalidInput.
// Mutation: писать сырой req.Currency без normalize → "USD " 500/wrong → FAIL.
func TestCurrencyNormalize_A3W75003(t *testing.T) {
	svc, _, memberID, client := setupMarketplaceServiceWithClient(t)
	ctx := context.Background()

	price := int64(1500)
	created, err := svc.CreateListing(ctx, memberID, CreateListingRequest{
		Title:       "Currency pad",
		Description: "trim space",
		Category:    CategoryOther,
		PriceCents:  &price,
		Currency:    "USD ",
	})
	require.NoError(t, err)
	assert.Equal(t, "USD", created.Currency)

	row, err := client.MarketplaceListing.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "USD", row.Currency)

	rub, err := svc.CreateListing(ctx, memberID, CreateListingRequest{
		Title:       "Currency lower",
		Description: "case fold",
		Category:    CategoryOther,
		PriceCents:  &price,
		Currency:    "rub",
	})
	require.NoError(t, err)
	assert.Equal(t, "RUB", rub.Currency)

	_, err = svc.CreateListing(ctx, memberID, CreateListingRequest{
		Title:       "Bad ccy",
		Description: "not allowlist",
		Category:    CategoryOther,
		PriceCents:  &price,
		Currency:    "BTC",
	})
	assert.ErrorIs(t, err, marketplace.ErrInvalidInput)

	assert.Equal(t, "USD", marketplace.ExportNormalizeCurrency("USD "))
	assert.Equal(t, "RUB", marketplace.ExportNormalizeCurrency("rub"))
}

// TestReviewNoteRuneCap_A3W75004 — note > 2000 runes → ErrInvalidInput; короткий проходит.
// Mutation: убрать rune cap в transitionReview → oversized проходит → FAIL.
func TestReviewNoteRuneCap_A3W75004(t *testing.T) {
	svc, adminID, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	cap := marketplace.ExportMaxReviewNoteRunes()
	require.Equal(t, 2000, cap)

	created := createTestListing(t, svc, memberID)
	oversized := strings.Repeat("я", cap+1)
	require.Equal(t, cap+1, utf8.RuneCountInString(oversized))

	_, err := svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{
		Version:    created.Version,
		ReviewNote: oversized,
	})
	assert.ErrorIs(t, err, marketplace.ErrInvalidInput)

	created2 := createTestListing(t, svc, memberID)
	okNote := strings.Repeat("a", cap)
	approved, err := svc.ApproveListing(ctx, adminID, created2.ID, ReviewListingRequest{
		Version:    created2.Version,
		ReviewNote: okNote,
	})
	require.NoError(t, err)
	assert.Equal(t, okNote, approved.ReviewNote)
}
