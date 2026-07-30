package marketplace_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/hook"
	"github.com/valerii/hub/backend/ent/marketplacelisting"
	"github.com/valerii/hub/backend/internal/marketplace"
)

// TestListMine_PaginationCap_W7A5003 — ListMine/Pending с Limit; has_more при N>per_page.
// Mutation: убрать Limit из ListListingsByAuthor → при N=3,per_page=2 вернёт 3 → FAIL.
func TestListMine_PaginationCap_W7A5003(t *testing.T) {
	svc, _, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := svc.CreateListing(ctx, memberID, CreateListingRequest{
			Title:       "Item " + string(rune('A'+i)),
			Description: "desc for pagination",
			Category:    CategoryOther,
		})
		require.NoError(t, err)
	}

	page1, err := svc.ListMine(ctx, memberID, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 3, page1.Total)
	assert.Len(t, page1.Items, 2)
	assert.True(t, page1.HasMore)
	assert.Equal(t, 1, page1.Page)
	assert.Equal(t, 2, page1.PerPage)

	page2, err := svc.ListMine(ctx, memberID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2.Items, 1)
	assert.False(t, page2.HasMore)
}

// TestListPending_PaginationCap_W7A5003 — admin pending list capped.
func TestListPending_PaginationCap_W7A5003(t *testing.T) {
	svc, adminID, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := svc.CreateListing(ctx, memberID, CreateListingRequest{
			Title:       "Pend " + string(rune('A'+i)),
			Description: "pending item",
			Category:    CategoryBooks,
		})
		require.NoError(t, err)
	}

	page1, err := svc.ListPending(ctx, adminID, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 3, page1.Total)
	assert.Len(t, page1.Items, 2)
	assert.True(t, page1.HasMore)
}

// TestUpdateApproved_AtomicRemoderation_W7A5004 — один UPDATE: fields+pending.
// Ent hook после каждого UpdateOne: mid-window approved+newTitle запрещён.
// Mutation: split UpdateListing + SetListingStatus → hook видит approved+newTitle → RED.
func TestUpdateApproved_AtomicRemoderation_W7A5004(t *testing.T) {
	svc, adminID, memberID, client := setupMarketplaceServiceWithClient(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)
	approved, err := svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{Version: created.Version})
	require.NoError(t, err)
	assert.Equal(t, StatusApproved, approved.Status)

	newTitle := "Atomic Remoderation Title"
	listingID := approved.ID

	var sawApprovedWithNewTitle atomic.Bool
	client.MarketplaceListing.Use(hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}
			row, gerr := client.MarketplaceListing.Get(ctx, listingID)
			if gerr == nil && row.Status == marketplacelisting.StatusApproved && row.Title == newTitle {
				sawApprovedWithNewTitle.Store(true)
			}
			return v, nil
		})
	}, ent.OpUpdateOne))

	updated, err := svc.UpdateListing(ctx, memberID, approved.ID, UpdateListingRequest{
		Title:   &newTitle,
		Version: approved.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, StatusPending, updated.Status)
	assert.Equal(t, newTitle, updated.Title)

	row, err := client.MarketplaceListing.Get(ctx, approved.ID)
	require.NoError(t, err)
	assert.Equal(t, "pending", string(row.Status))
	assert.Equal(t, newTitle, row.Title)

	assert.False(t, sawApprovedWithNewTitle.Load(),
		"W7-A5-004: mid-window approved+new title must not be visible between Saves")
}

// TestAuthorName_NoEmailOnPublicCatalog_W7A5005 — витрина без email PII.
// Mutation: вернуть email fallback в displayName(public) → author_name = email → FAIL.
func TestAuthorName_NoEmailOnPublicCatalog_W7A5005(t *testing.T) {
	svc, adminID, memberID, client := setupMarketplaceServiceWithClient(t)
	ctx := context.Background()

	_, err := client.User.UpdateOneID(memberID).
		ClearDisplayName().
		ClearName().
		Save(ctx)
	require.NoError(t, err)

	created := createTestListing(t, svc, memberID)
	_, err = svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{Version: created.Version})
	require.NoError(t, err)

	public, err := svc.ListApproved(ctx, "", 1, 10)
	require.NoError(t, err)
	require.Len(t, public.Items, 1)
	assert.Equal(t, "Пользователь", public.Items[0].AuthorName)
	assert.NotContains(t, public.Items[0].AuthorName, "@")

	mine, err := svc.ListMine(ctx, memberID, 1, 10)
	require.NoError(t, err)
	require.Len(t, mine.Items, 1)
	// ListMine — private: email fallback допустим.
	assert.Contains(t, mine.Items[0].AuthorName, "@")
}

// TestValidatePriceCurrency_W7A5006 — price≥0 + currency allowlist.
// Mutation: убрать проверки из validateListingInput → отрицательная цена/XYZ проходят → FAIL.
func TestValidatePriceCurrency_W7A5006(t *testing.T) {
	svc, _, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	neg := int64(-100)
	_, err := svc.CreateListing(ctx, memberID, CreateListingRequest{
		Title:       "Bad price",
		Description: "negative",
		Category:    CategoryOther,
		PriceCents:  &neg,
		Currency:    "RUB",
	})
	assert.ErrorIs(t, err, marketplace.ErrInvalidInput)

	ok := int64(100)
	_, err = svc.CreateListing(ctx, memberID, CreateListingRequest{
		Title:       "Bad currency",
		Description: "xyz",
		Category:    CategoryOther,
		PriceCents:  &ok,
		Currency:    "XYZ",
	})
	assert.ErrorIs(t, err, marketplace.ErrInvalidInput)

	created, err := svc.CreateListing(ctx, memberID, CreateListingRequest{
		Title:       "Good price",
		Description: "valid",
		Category:    CategoryOther,
		PriceCents:  &ok,
		Currency:    "USD",
	})
	require.NoError(t, err)
	assert.Equal(t, "USD", created.Currency)

	assert.True(t, marketplace.ExportIsValidCurrency("rub"))
	assert.False(t, marketplace.ExportIsValidCurrency("BTC"))
}

// TestListApproved_HasMore_W7A5003 — каталог отдаёт has_more.
func TestListApproved_HasMore_W7A5003(t *testing.T) {
	svc, adminID, memberID, _ := setupMarketplaceServiceWithClient(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		created, err := svc.CreateListing(ctx, memberID, CreateListingRequest{
			Title:       "Cat " + string(rune('1'+i)),
			Description: "catalog",
			Category:    CategoryEquipment,
		})
		require.NoError(t, err)
		_, err = svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{Version: created.Version})
		require.NoError(t, err)
	}

	page, err := svc.ListApproved(ctx, string(CategoryEquipment), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 3, page.Total)
	assert.Len(t, page.Items, 2)
	assert.True(t, page.HasMore)
}
