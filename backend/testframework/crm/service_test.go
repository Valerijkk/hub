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

func setupCRMService(t *testing.T) (*crm.Service, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:crm_svc?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("crm-owner@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	stranger, err := client.User.Create().
		SetEmail("stranger@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("CRM Project").
		SetKey("CRM1").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	return crm.NewService(crm.NewEntRepository(client)), owner.ID, stranger.ID, proj.ID
}

func TestService_ContactLifecycle(t *testing.T) {
	svc, ownerID, _, _ := setupCRMService(t)
	ctx := context.Background()

	_, err := svc.CreateContact(ctx, ownerID, crm.CreateContactRequest{Name: "  "})
	assert.ErrorIs(t, err, crm.ErrInvalidInput)

	created, err := svc.CreateContact(ctx, ownerID, crm.CreateContactRequest{
		Name:    "Acme Corp",
		Email:   "buyer@acme.com",
		Company: "Acme",
	})
	require.NoError(t, err)

	list, err := svc.ListContacts(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	newName := "Acme International"
	updated, err := svc.UpdateContact(ctx, created.ID, crm.UpdateContactRequest{
		Name:    &newName,
		Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)

	_, err = svc.UpdateContact(ctx, created.ID, crm.UpdateContactRequest{
		Name:    &newName,
		Version: created.Version,
	})
	assert.ErrorIs(t, err, crm.ErrVersionConflict)

	err = svc.DeleteContact(ctx, created.ID)
	require.NoError(t, err)

	list, err = svc.ListContacts(ctx)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestService_ContactInvalidLinkedUser(t *testing.T) {
	svc, ownerID, _, _ := setupCRMService(t)
	ctx := context.Background()

	badID := uuid.New()
	_, err := svc.CreateContact(ctx, ownerID, crm.CreateContactRequest{
		Name:         "Linked",
		LinkedUserID: &badID,
	})
	assert.ErrorIs(t, err, crm.ErrInvalidInput)

	_, err = svc.CreateContact(ctx, ownerID, crm.CreateContactRequest{
		Name:         "Valid link",
		LinkedUserID: &ownerID,
	})
	require.NoError(t, err)
}

func TestService_DealValidationAndDelete(t *testing.T) {
	svc, ownerID, strangerID, projID := setupCRMService(t)
	ctx := context.Background()

	_, err := svc.CreateDeal(ctx, ownerID, crm.CreateDealRequest{Title: ""})
	assert.ErrorIs(t, err, crm.ErrInvalidInput)

	badStage := crm.DealStage("invalid")
	_, err = svc.CreateDeal(ctx, ownerID, crm.CreateDealRequest{
		Title: "X",
		Stage: badStage,
	})
	assert.ErrorIs(t, err, crm.ErrInvalidInput)

	_, err = svc.CreateDeal(ctx, strangerID, crm.CreateDealRequest{
		Title:     "Forbidden",
		ProjectID: &projID,
	})
	assert.ErrorIs(t, err, crm.ErrProjectForbidden)

	deal, err := svc.CreateDeal(ctx, ownerID, crm.CreateDealRequest{
		Title:     "Enterprise",
		Stage:     crm.StageLead,
		ProjectID: &projID,
	})
	require.NoError(t, err)

	pipeline, err := svc.GetPipeline(ctx)
	require.NoError(t, err)
	assert.Len(t, pipeline.Stages, len(crm.PipelineStages))

	lost := crm.StageLost
	updated, err := svc.UpdateDeal(ctx, ownerID, deal.ID, crm.UpdateDealRequest{
		Stage:   &lost,
		Version: deal.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, crm.StageLost, updated.Stage)

	err = svc.DeleteDeal(ctx, deal.ID)
	require.NoError(t, err)

	err = svc.DeleteDeal(ctx, uuid.New())
	assert.ErrorIs(t, err, crm.ErrNotFound)
}

func TestService_DealVersionConflict(t *testing.T) {
	svc, ownerID, _, projID := setupCRMService(t)
	ctx := context.Background()

	deal, err := svc.CreateDeal(ctx, ownerID, crm.CreateDealRequest{
		Title:     "Renewal",
		ProjectID: &projID,
	})
	require.NoError(t, err)

	stage := crm.StageProposal
	_, err = svc.UpdateDeal(ctx, ownerID, deal.ID, crm.UpdateDealRequest{
		Stage:   &stage,
		Version: deal.Version,
	})
	require.NoError(t, err)

	_, err = svc.UpdateDeal(ctx, ownerID, deal.ID, crm.UpdateDealRequest{
		Stage:   &stage,
		Version: deal.Version,
	})
	assert.ErrorIs(t, err, crm.ErrVersionConflict)
}

func TestService_UpdateContactNotFound(t *testing.T) {
	svc, ownerID, _, _ := setupCRMService(t)
	ctx := context.Background()
	name := "X"
	_, err := svc.UpdateContact(ctx, uuid.New(), crm.UpdateContactRequest{Name: &name, Version: 1})
	assert.ErrorIs(t, err, crm.ErrNotFound)
	_ = ownerID
}
