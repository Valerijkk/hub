package marketplace_test
import (
	"github.com/valerii/hub/backend/internal/marketplace"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

type stubNotifier struct {
	submitted int
	decisions int
}

func (s *stubNotifier) NotifyListingSubmitted(_ context.Context, _ []uuid.UUID, _ ListingResponse, _ uuid.UUID) error {
	s.submitted++
	return nil
}

func (s *stubNotifier) NotifyListingDecision(_ context.Context, _ uuid.UUID, _ bool, _ ListingResponse, _ uuid.UUID) error {
	s.decisions++
	return nil
}

type stubBroadcaster struct {
	called bool
}

func (s *stubBroadcaster) BroadcastApprovedListing(_ context.Context, _ ListingResponse) error {
	s.called = true
	return nil
}

func setupMarketplaceService(t *testing.T) (*Service, uuid.UUID, uuid.UUID, *stubNotifier) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:mp_svc_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	admin, err := client.User.Create().
		SetEmail("mp-admin-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(context.Background())
	require.NoError(t, err)

	member, err := client.User.Create().
		SetEmail("mp-member-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(context.Background())
	require.NoError(t, err)

	notifier := &stubNotifier{}
	svc := NewService(NewEntRepository(client), notifier, &stubBroadcaster{})
	return svc, admin.ID, member.ID, notifier
}

// setupMarketplaceServiceWithClient — как setupMarketplaceService, но отдаёт *ent.Client
// для создания дополнительных пользователей (W7-A5-016 stranger denial).
func setupMarketplaceServiceWithClient(t *testing.T) (*Service, uuid.UUID, uuid.UUID, *ent.Client) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:mp_svc_c_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	admin, err := client.User.Create().
		SetEmail("mp-admin-c-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(context.Background())
	require.NoError(t, err)

	member, err := client.User.Create().
		SetEmail("mp-member-c-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(context.Background())
	require.NoError(t, err)

	svc := NewService(NewEntRepository(client), &stubNotifier{}, &stubBroadcaster{})
	return svc, admin.ID, member.ID, client
}

func createTestListing(t *testing.T, svc *Service, authorID uuid.UUID) ListingResponse {
	t.Helper()
	resp, err := svc.CreateListing(context.Background(), authorID, CreateListingRequest{
		Title:       "Monitor 27",
		Description: "4K IPS display",
		Category:    CategoryEquipment,
	})
	require.NoError(t, err)
	return *resp
}

func TestService_CreateListingInvalidInput(t *testing.T) {
	svc, _, memberID, _ := setupMarketplaceService(t)
	_, err := svc.CreateListing(context.Background(), memberID, CreateListingRequest{
		Title:       "",
		Description: "x",
		Category:    CategoryBooks,
	})
	assert.ErrorIs(t, err, ErrInvalidInput)

	_, err = svc.CreateListing(context.Background(), memberID, CreateListingRequest{
		Title:       "Valid",
		Description: "Valid",
		Category:    ListingCategory("invalid"),
	})
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestService_ApproveRejectSoldWithdraw(t *testing.T) {
	svc, adminID, memberID, notifier := setupMarketplaceService(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)
	assert.Equal(t, 1, notifier.submitted)

	approved, err := svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{Version: created.Version})
	require.NoError(t, err)
	assert.Equal(t, StatusApproved, approved.Status)
	assert.Equal(t, 1, notifier.decisions)

	sold, err := svc.MarkSold(ctx, memberID, approved.ID, VersionRequest{Version: approved.Version})
	require.NoError(t, err)
	assert.Equal(t, StatusSold, sold.Status)

	created2 := createTestListing(t, svc, memberID)
	rejected, err := svc.RejectListing(ctx, adminID, created2.ID, ReviewListingRequest{
		Version:    created2.Version,
		ReviewNote: "spam",
	})
	require.NoError(t, err)
	assert.Equal(t, StatusRejected, rejected.Status)
	assert.Equal(t, "spam", rejected.ReviewNote)

	created3 := createTestListing(t, svc, memberID)
	withdrawn, err := svc.WithdrawListing(ctx, memberID, created3.ID, VersionRequest{Version: created3.Version})
	require.NoError(t, err)
	assert.Equal(t, StatusWithdrawn, withdrawn.Status)
}

func TestService_UpdateApprovedRemoderation(t *testing.T) {
	svc, adminID, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)
	approved, err := svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{Version: created.Version})
	require.NoError(t, err)

	newTitle := "Monitor 27 Pro"
	updated, err := svc.UpdateListing(ctx, memberID, approved.ID, UpdateListingRequest{
		Title:   &newTitle,
		Version: approved.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, StatusPending, updated.Status)
	assert.Equal(t, newTitle, updated.Title)
}

func TestService_ListMineAndApproved(t *testing.T) {
	svc, adminID, memberID, _ := setupMarketplaceService(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)
	_, err := svc.ApproveListing(ctx, adminID, created.ID, ReviewListingRequest{Version: created.Version})
	require.NoError(t, err)

	mine, err := svc.ListMine(ctx, memberID, 1, 25)
	require.NoError(t, err)
	assert.Len(t, mine.Items, 1)

	public, err := svc.ListApproved(ctx, string(CategoryEquipment), 1, 10)
	require.NoError(t, err)
	assert.Len(t, public.Items, 1)
	assert.Equal(t, 1, public.Total)
	assert.False(t, public.HasMore)
}

func TestService_DeleteListingPermissions(t *testing.T) {
	svc, adminID, memberID, client := setupMarketplaceServiceWithClient(t)
	ctx := context.Background()

	created := createTestListing(t, svc, memberID)

	stranger, err := client.User.Create().
		SetEmail("mp-stranger-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	err = svc.DeleteListing(ctx, stranger.ID, created.ID)
	assert.ErrorIs(t, err, marketplace.ErrNotOwner)

	err = svc.DeleteListing(ctx, memberID, created.ID)
	require.NoError(t, err)

	err = svc.DeleteListing(ctx, adminID, created.ID)
	assert.ErrorIs(t, err, marketplace.ErrNotFound)
}

func TestFormatPriceHelpers(t *testing.T) {
	assert.Equal(t, "Договорная", formatPrice(nil, "RUB"))
	zero := int64(0)
	assert.Equal(t, "Бесплатно", formatPrice(&zero, "RUB"))
	cents := int64(150050)
	assert.Equal(t, "RUB 1500.50", formatPrice(&cents, "RUB"))
}
