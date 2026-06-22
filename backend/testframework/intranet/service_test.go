package intranet_test
import (
	"github.com/valerii/hub/backend/internal/intranet"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

type stubKudosNotifier struct {
	called bool
}

func (s *stubKudosNotifier) NotifyKudosReceived(_ context.Context, _ uuid.UUID, _ intranet.KudoResponse, _ uuid.UUID) error {
	s.called = true
	return nil
}

func setupIntranetService(t *testing.T) (*intranet.Service, uuid.UUID, uuid.UUID) {
	svc, adminID, memberID := setupIntranetServiceWithNotifiers(t, nil)
	return svc, adminID, memberID
}

func setupIntranetServiceWithCoffee(t *testing.T) (*intranet.Service, uuid.UUID, uuid.UUID) {
	coffeeNotifier := &stubCoffeeNotifier{}
	svc, adminID, memberID := setupIntranetServiceWithNotifiers(t, coffeeNotifier)
	return svc, adminID, memberID
}

func setupIntranetServiceWithNotifiers(t *testing.T, coffeeNotifier CoffeeNotifier) (*intranet.Service, uuid.UUID, uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:intra_svc?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("intra-admin@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	member, err := client.User.Create().
		SetEmail("intra-member@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	notifier := &stubKudosNotifier{}
	return intranet.NewService(intranet.NewEntRepository(client), notifier, coffeeNotifier), admin.ID, member.ID
}

func TestService_UpdateNewsVersionConflict(t *testing.T) {
	svc, adminID, _ := setupIntranetService(t)
	ctx := context.Background()

	created, err := svc.CreateNews(ctx, adminID, intranet.CreateNewsRequest{
		Title:     "Draft",
		Body:      "Content",
		Published: true,
	})
	require.NoError(t, err)

	newTitle := "Updated"
	updated, err := svc.UpdateNews(ctx, adminID, created.ID, intranet.UpdateNewsRequest{
		Title:   &newTitle,
		Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)

	_, err = svc.UpdateNews(ctx, adminID, created.ID, intranet.UpdateNewsRequest{
		Title:   &newTitle,
		Version: created.Version,
	})
	assert.ErrorIs(t, err, intranet.ErrVersionConflict)
}

func TestService_CreateKudoSelfAndInvalid(t *testing.T) {
	svc, adminID, memberID := setupIntranetService(t)
	ctx := context.Background()

	_, err := svc.CreateKudo(ctx, memberID, intranet.CreateKudoRequest{
		RecipientID: memberID,
		Message:     "self",
	})
	assert.ErrorIs(t, err, intranet.ErrSelfKudo)

	_, err = svc.CreateKudo(ctx, memberID, intranet.CreateKudoRequest{
		RecipientID: adminID,
		Message:     "",
	})
	assert.ErrorIs(t, err, intranet.ErrInvalidKudo)
}

func TestService_ListNewsAdminIncludesDraft(t *testing.T) {
	svc, adminID, memberID := setupIntranetService(t)
	ctx := context.Background()

	_, err := svc.CreateNews(ctx, adminID, intranet.CreateNewsRequest{
		Title:     "Draft post",
		Body:      "Not published yet",
		Published: false,
	})
	require.NoError(t, err)

	public, err := svc.ListNews(ctx, 1, 10)
	require.NoError(t, err)
	assert.Empty(t, public.Items)

	adminList, err := svc.ListNewsAdmin(ctx, adminID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, adminList.Items, 1)

	_, err = svc.ListNewsAdmin(ctx, memberID, 1, 10)
	assert.ErrorIs(t, err, intranet.ErrForbidden)
}

func TestService_GetNewsDraftHiddenFromMember(t *testing.T) {
	svc, adminID, memberID := setupIntranetService(t)
	ctx := context.Background()

	draft, err := svc.CreateNews(ctx, adminID, intranet.CreateNewsRequest{
		Title:     "Secret",
		Body:      "Internal only",
		Published: false,
	})
	require.NoError(t, err)

	_, err = svc.GetNews(ctx, memberID, draft.ID, false)
	assert.ErrorIs(t, err, intranet.ErrNotFound)

	got, err := svc.GetNews(ctx, adminID, draft.ID, false)
	require.NoError(t, err)
	assert.Equal(t, "Secret", got.Title)
}
