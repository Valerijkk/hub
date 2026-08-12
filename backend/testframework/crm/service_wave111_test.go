package crm_test
import (
	"github.com/valerii/hub/backend/internal/crm"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
)

func TestService_UpdateContact_LinkedUserSuccess(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_link_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	linked, err := client.User.Create().
		SetEmail("linked-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	svc := crm.NewService(crm.NewEntRepository(client))
	created, err := svc.CreateContact(ctx, owner.ID, crm.CreateContactRequest{
		Name:  "Partner",
		Email: "partner@test.com",
	})
	require.NoError(t, err)

	linkedID := linked.ID
	updated, err := svc.UpdateContact(ctx, owner.ID, created.ID, crm.UpdateContactRequest{
		LinkedUserID: &linkedID,
		Version:      created.Version,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.LinkedUserID)
	assert.Equal(t, linked.ID, *updated.LinkedUserID)
}

func TestService_UpdateContact_InvalidLinkedUser(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_badlink_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("crm-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	svc := crm.NewService(crm.NewEntRepository(client))
	created, err := svc.CreateContact(ctx, owner.ID, crm.CreateContactRequest{Name: "Ghost", Email: "g@test.com"})
	require.NoError(t, err)

	badID := uuid.New()
	_, err = svc.UpdateContact(ctx, owner.ID, created.ID, crm.UpdateContactRequest{
		LinkedUserID: &badID,
		Version:      created.Version,
	})
	require.ErrorIs(t, err, crm.ErrInvalidInput)
}

func TestService_UpdateContact_VersionConflict(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_ver_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("ver-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	svc := crm.NewService(crm.NewEntRepository(client))
	created, err := svc.CreateContact(ctx, owner.ID, crm.CreateContactRequest{Name: "Ver Co", Email: "v@test.com"})
	require.NoError(t, err)

	name := "Renamed"
	_, err = svc.UpdateContact(ctx, owner.ID, created.ID, crm.UpdateContactRequest{
		Name:    &name,
		Version: created.Version + 99,
	})
	require.ErrorIs(t, err, crm.ErrVersionConflict)
}

func TestService_CreateDeal_InvalidContactRef(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_cref_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("cref-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	svc := crm.NewService(crm.NewEntRepository(client))
	badContact := uuid.New()
	_, err = svc.CreateDeal(ctx, owner.ID, crm.CreateDealRequest{
		Title:     "Orphan deal",
		Stage:     crm.StageLead,
		ContactID: &badContact,
	})
	require.ErrorIs(t, err, crm.ErrInvalidInput)
}

func TestService_UpdateDeal_InvalidContactRef(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_ucref_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("uc-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("CRM").
		SetKey("CRM").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().SetProjectID(proj.ID).SetUserID(owner.ID).Save(ctx)
	require.NoError(t, err)

	svc := crm.NewService(crm.NewEntRepository(client))
	deal, err := svc.CreateDeal(ctx, owner.ID, crm.CreateDealRequest{
		Title:     "Base deal",
		Stage:     crm.StageLead,
		ProjectID: &proj.ID,
	})
	require.NoError(t, err)

	badContact := uuid.New()
	_, err = svc.UpdateDeal(ctx, owner.ID, deal.ID, crm.UpdateDealRequest{
		ContactID: &badContact,
		Version:   deal.Version,
	})
	require.ErrorIs(t, err, crm.ErrInvalidInput)
}

func TestService_DeleteContact_NotFound(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_dnf_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	svc := crm.NewService(crm.NewEntRepository(client))
	err := svc.DeleteContact(context.Background(), uuid.New(), uuid.New())
	require.ErrorIs(t, err, crm.ErrNotFound)
}

func TestService_UpdateDeal_ProjectForbidden(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_pforb_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("pf-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	otherProj, err := client.Project.Create().
		SetName("Other").
		SetKey("OTH").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	svc := crm.NewService(crm.NewEntRepository(client))
	deal, err := svc.CreateDeal(ctx, owner.ID, crm.CreateDealRequest{Title: "Open", Stage: crm.StageLead})
	require.NoError(t, err)

	_, err = svc.UpdateDeal(ctx, owner.ID, deal.ID, crm.UpdateDealRequest{
		ProjectID: &otherProj.ID,
		Version:   deal.Version,
	})
	require.ErrorIs(t, err, crm.ErrProjectForbidden)
}
