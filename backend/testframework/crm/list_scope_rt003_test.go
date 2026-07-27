package crm_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/crm"
)

// TestListContacts_StrangerCannotSeeForeign_RT003 — member не видит чужие контакты;
// admin видит org-wide (product decision RT-003).
func TestListContacts_StrangerCannotSeeForeign_RT003(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_rt003_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)
	stranger, err := client.User.Create().
		SetEmail("stranger-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)
	admin, err := client.User.Create().
		SetEmail("admin-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	svc := crm.NewService(crm.NewEntRepository(client))
	mine, err := svc.CreateContact(ctx, owner.ID, crm.CreateContactRequest{
		Name:  "Secret Lead",
		Email: "secret@acme.test",
	})
	require.NoError(t, err)
	_, err = svc.CreateDeal(ctx, owner.ID, crm.CreateDealRequest{
		Title: "Secret deal",
		Stage: crm.StageLead,
	})
	require.NoError(t, err)

	strangerList, err := svc.ListContacts(ctx, stranger.ID)
	require.NoError(t, err)
	assert.Empty(t, strangerList, "stranger must not see foreign contacts")

	ownerList, err := svc.ListContacts(ctx, owner.ID)
	require.NoError(t, err)
	require.Len(t, ownerList, 1)
	assert.Equal(t, mine.ID, ownerList[0].ID)

	adminList, err := svc.ListContacts(ctx, admin.ID)
	require.NoError(t, err)
	require.Len(t, adminList, 1, "admin sees org-wide contacts")
	assert.Equal(t, mine.ID, adminList[0].ID)

	strangerPipe, err := svc.GetPipeline(ctx, stranger.ID)
	require.NoError(t, err)
	var strangerDeals int
	for _, stage := range strangerPipe.Stages {
		strangerDeals += len(stage.Deals)
	}
	assert.Equal(t, 0, strangerDeals, "stranger must not see foreign deals")

	ownerPipe, err := svc.GetPipeline(ctx, owner.ID)
	require.NoError(t, err)
	var ownerDeals int
	for _, stage := range ownerPipe.Stages {
		ownerDeals += len(stage.Deals)
	}
	assert.Equal(t, 1, ownerDeals)

	adminPipe, err := svc.GetPipeline(ctx, admin.ID)
	require.NoError(t, err)
	var adminDeals int
	for _, stage := range adminPipe.Stages {
		adminDeals += len(stage.Deals)
	}
	assert.Equal(t, 1, adminDeals, "admin sees org-wide pipeline")
}
