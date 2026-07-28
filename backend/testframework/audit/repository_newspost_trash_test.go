package audit_test

// FR-INTR-08: newspost в единой админ-корзине (list / restore / retention / purge).
import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/newspost"
	"github.com/valerii/hub/backend/internal/audit"
)

func (e auditTrashEnv) createNewsPost(t *testing.T, title string) *ent.NewsPost {
	t.Helper()
	post, err := e.client.NewsPost.Create().
		SetTitle(title).
		SetBody("body for " + title).
		SetAuthorID(e.userA).
		Save(e.ctx)
	require.NoError(t, err)
	return post
}

// TestRepository_SoftDeleteAndRestoreNewsPost — слой репозитория: soft-delete → list → restore.
func TestRepository_SoftDeleteAndRestoreNewsPost(t *testing.T) {
	env := newAuditTrashEnv(t)
	post := env.createNewsPost(t, "News in trash")
	when := time.Now().UTC()

	require.NoError(t, env.repo.SoftDeleteNewsPost(env.ctx, post.ID, when))

	items, total, err := env.repo.ListDeletedNewsPosts(env.ctx, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, post.ID, items[0].ID)
	assert.Equal(t, "newspost", items[0].EntityType)

	require.NoError(t, env.repo.RestoreNewsPost(env.ctx, post.ID))

	items, total, err = env.repo.ListDeletedNewsPosts(env.ctx, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)

	restored, err := env.client.NewsPost.Get(env.ctx, post.ID)
	require.NoError(t, err)
	assert.Nil(t, restored.DeletedAt)
}

// TestRepository_NewsPost_RetentionHidesOlderThan30Days — NFR-11: старше cutoff не в списке.
func TestRepository_NewsPost_RetentionHidesOlderThan30Days(t *testing.T) {
	env := newAuditTrashEnv(t)
	post := env.createNewsPost(t, "Old news")
	old := time.Now().UTC().Add(-40 * 24 * time.Hour)
	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)

	require.NoError(t, env.repo.SoftDeleteNewsPost(env.ctx, post.ID, old))

	items, total, err := env.repo.ListDeletedNewsPosts(env.ctx, cutoff, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

// TestRepository_PurgeExpiredNewsPosts — hard delete после retention.
func TestRepository_PurgeExpiredNewsPosts(t *testing.T) {
	env := newAuditTrashEnv(t)
	post := env.createNewsPost(t, "Purge me")
	old := time.Now().UTC().Add(-40 * 24 * time.Hour)

	require.NoError(t, env.repo.SoftDeleteNewsPost(env.ctx, post.ID, old))

	n, err := env.repo.PurgeExpiredNewsPosts(env.ctx, time.Now().UTC().Add(-30*24*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	_, err = env.client.NewsPost.Get(env.ctx, post.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

// TestAuditService_RestoreNewsPost_ViaTrashAPI — сервисный путь RestoreItem("newspost")
// бьёт в реальный RestoreNewsPost (не mock). Мутация: убрать case "newspost" из RestoreItem.
func TestAuditService_RestoreNewsPost_ViaTrashAPI(t *testing.T) {
	env := newAuditTrashEnv(t)
	post := env.createNewsPost(t, "Restore via service")
	when := time.Now().UTC()
	require.NoError(t, env.repo.SoftDeleteNewsPost(env.ctx, post.ID, when))

	svc := audit.NewAuditService(env.repo, nil)

	list, err := svc.ListTrash(env.ctx, "newspost", nil, 1, 50)
	require.NoError(t, err)
	require.Equal(t, 1, list.Total)
	require.Len(t, list.Items, 1)
	assert.Equal(t, post.ID.String(), list.Items[0].ID)
	assert.Greater(t, list.Items[0].DaysRemaining, 0)

	require.NoError(t, svc.RestoreItem(env.ctx, "newspost", post.ID))

	got, err := env.client.NewsPost.Query().
		Where(newspost.IDEQ(post.ID), newspost.DeletedAtIsNil()).
		Only(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, "Restore via service", got.Title)

	listAfter, err := svc.ListTrash(env.ctx, "newspost", nil, 1, 50)
	require.NoError(t, err)
	assert.Equal(t, 0, listAfter.Total)
}
