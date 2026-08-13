package audit_test

// W7-A2-006: crm_contact / crm_deal в единой админ-корзине.
import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/crmcontact"
	"github.com/valerii/hub/backend/ent/crmdeal"
	"github.com/valerii/hub/backend/internal/audit"
)

func (e auditTrashEnv) createCRMContact(t *testing.T, name string) *ent.CrmContact {
	t.Helper()
	c, err := e.client.CrmContact.Create().
		SetName(name).
		SetCreatedByID(e.userA).
		Save(e.ctx)
	require.NoError(t, err)
	return c
}

func (e auditTrashEnv) createCRMDeal(t *testing.T, title string) *ent.CrmDeal {
	t.Helper()
	d, err := e.client.CrmDeal.Create().
		SetTitle(title).
		SetOwnerID(e.userA).
		Save(e.ctx)
	require.NoError(t, err)
	return d
}

func TestRepository_SoftDeleteAndRestoreCRMContact_W7A2006(t *testing.T) {
	env := newAuditTrashEnv(t)
	c := env.createCRMContact(t, "Alice Contact")
	when := time.Now().UTC()

	require.NoError(t, env.repo.SoftDeleteCRMContact(env.ctx, c.ID, when))

	items, total, err := env.repo.ListDeletedCRMContacts(env.ctx, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, c.ID, items[0].ID)
	assert.Equal(t, "crm_contact", items[0].EntityType)

	require.NoError(t, env.repo.RestoreCRMContact(env.ctx, c.ID))

	items, total, err = env.repo.ListDeletedCRMContacts(env.ctx, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)

	restored, err := env.client.CrmContact.Query().
		Where(crmcontact.IDEQ(c.ID), crmcontact.DeletedAtIsNil()).
		Only(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, "Alice Contact", restored.Name)
}

func TestRepository_SoftDeleteAndRestoreCRMDeal_W7A2006(t *testing.T) {
	env := newAuditTrashEnv(t)
	d := env.createCRMDeal(t, "Big Deal")
	when := time.Now().UTC()

	require.NoError(t, env.repo.SoftDeleteCRMDeal(env.ctx, d.ID, when))

	items, total, err := env.repo.ListDeletedCRMDeals(env.ctx, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, "crm_deal", items[0].EntityType)

	require.NoError(t, env.repo.RestoreCRMDeal(env.ctx, d.ID))

	got, err := env.client.CrmDeal.Query().
		Where(crmdeal.IDEQ(d.ID), crmdeal.DeletedAtIsNil()).
		Only(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, "Big Deal", got.Title)
}

// TestRepository_CRMContact_RetentionHidesOlderThan30Days — A3W6-5-003.
// Мутация: снять DeletedAtGT(cutoff) в ListDeletedCRMContacts → total≠0.
func TestRepository_CRMContact_RetentionHidesOlderThan30Days(t *testing.T) {
	env := newAuditTrashEnv(t)
	c := env.createCRMContact(t, "Old Contact Hide")
	old := time.Now().UTC().Add(-40 * 24 * time.Hour)
	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)

	require.NoError(t, env.repo.SoftDeleteCRMContact(env.ctx, c.ID, old))

	items, total, err := env.repo.ListDeletedCRMContacts(env.ctx, cutoff, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

// TestRepository_CRMDeal_RetentionHidesOlderThan30Days — A3W6-5-003.
// Мутация: снять DeletedAtGT(cutoff) в ListDeletedCRMDeals → total≠0.
func TestRepository_CRMDeal_RetentionHidesOlderThan30Days(t *testing.T) {
	env := newAuditTrashEnv(t)
	d := env.createCRMDeal(t, "Old Deal Hide")
	old := time.Now().UTC().Add(-40 * 24 * time.Hour)
	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)

	require.NoError(t, env.repo.SoftDeleteCRMDeal(env.ctx, d.ID, old))

	items, total, err := env.repo.ListDeletedCRMDeals(env.ctx, cutoff, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

func TestRepository_PurgeExpiredCRM_W7A2006(t *testing.T) {
	env := newAuditTrashEnv(t)
	c := env.createCRMContact(t, "Old Contact")
	d := env.createCRMDeal(t, "Old Deal")
	old := time.Now().UTC().Add(-40 * 24 * time.Hour)
	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)

	require.NoError(t, env.repo.SoftDeleteCRMContact(env.ctx, c.ID, old))
	require.NoError(t, env.repo.SoftDeleteCRMDeal(env.ctx, d.ID, old))

	n, err := env.repo.PurgeExpiredCRMContacts(env.ctx, cutoff)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	n, err = env.repo.PurgeExpiredCRMDeals(env.ctx, cutoff)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	_, err = env.client.CrmContact.Get(env.ctx, c.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
	_, err = env.client.CrmDeal.Get(env.ctx, d.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

// TestAuditService_RestoreCRM_ViaTrashAPI_W7A2006 — сервисный путь RestoreItem.
// Mutation: убрать case "crm_contact"/"crm_deal" → ошибка unknown type.
func TestAuditService_RestoreCRM_ViaTrashAPI_W7A2006(t *testing.T) {
	env := newAuditTrashEnv(t)
	c := env.createCRMContact(t, "Restore Contact")
	d := env.createCRMDeal(t, "Restore Deal")
	when := time.Now().UTC()
	require.NoError(t, env.repo.SoftDeleteCRMContact(env.ctx, c.ID, when))
	require.NoError(t, env.repo.SoftDeleteCRMDeal(env.ctx, d.ID, when))

	svc := audit.NewAuditService(env.repo, nil)

	list, err := svc.ListTrash(env.ctx, "crm_contact", nil, 1, 50)
	require.NoError(t, err)
	require.Equal(t, 1, list.Total)
	list, err = svc.ListTrash(env.ctx, "crm_deal", nil, 1, 50)
	require.NoError(t, err)
	require.Equal(t, 1, list.Total)

	require.NoError(t, svc.RestoreItem(env.ctx, "crm_contact", c.ID))
	require.NoError(t, svc.RestoreItem(env.ctx, "crm_deal", d.ID))

	_, err = env.client.CrmContact.Query().
		Where(crmcontact.IDEQ(c.ID), crmcontact.DeletedAtIsNil()).Only(env.ctx)
	require.NoError(t, err)
	_, err = env.client.CrmDeal.Query().
		Where(crmdeal.IDEQ(d.ID), crmdeal.DeletedAtIsNil()).Only(env.ctx)
	require.NoError(t, err)
}
