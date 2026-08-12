//go:build hubtest

package intranet_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/newspost"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/audit"
	"github.com/valerii/hub/backend/internal/intranet"
)

// TestDeleteNews_AuthorCanDeleteOwn_Task098 — FR-INTR-07 / W7-A2-001:
// автор-member удаляет свой пост; чужой member → 403.
// Мутация: вернуть DeleteNews на requireAdmin-only → ErrorIs(ErrForbidden) для автора.
func TestDeleteNews_AuthorCanDeleteOwn_Task098(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:t098_author_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("t098-admin-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	author, err := client.User.Create().
		SetEmail("t098-author-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	stranger, err := client.User.Create().
		SetEmail("t098-stranger-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	svc := intranet.NewService(intranet.NewEntRepository(client), nil, nil)

	post, err := client.NewsPost.Create().
		SetTitle("Mine").
		SetBody("body").
		SetPublished(true).
		SetAuthorID(author.ID).
		Save(ctx)
	require.NoError(t, err)

	require.ErrorIs(t, svc.DeleteNews(ctx, stranger.ID, post.ID), intranet.ErrForbidden)
	require.NoError(t, svc.DeleteNews(ctx, author.ID, post.ID))

	_, err = client.NewsPost.Query().
		Where(newspost.IDEQ(post.ID), newspost.DeletedAtIsNil()).
		Only(ctx)
	require.True(t, ent.IsNotFound(err), "post must be soft-deleted")

	post2, err := client.NewsPost.Create().
		SetTitle("Other").
		SetBody("body").
		SetPublished(true).
		SetAuthorID(author.ID).
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, svc.DeleteNews(ctx, admin.ID, post2.ID))
}

// TestNewsPoll_SoftDeleteRestoreCounters_Task098 — FR-INTR-07/08:
// create+vote → soft-delete скрывает poll → trash restore возвращает счётчики.
// Мутация: убрать author-ветку / HasNewsPostWith(DeletedAtIsNil) → краснеет.
func TestNewsPoll_SoftDeleteRestoreCounters_Task098(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:t098_poll_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("t098-poll-admin-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	voter, err := client.User.Create().
		SetEmail("t098-voter-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	svc := intranet.NewService(intranet.NewEntRepository(client), nil, nil)

	created, err := svc.CreateNews(ctx, admin.ID, intranet.CreateNewsRequest{
		Title: "With poll", Body: "body", Published: true,
		Poll: &intranet.CreatePollPayload{
			Question: "Lunch?",
			Options:  []string{"Pizza", "Sushi"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, created.Poll)
	require.Len(t, created.Poll.Options, 2)
	optPizza := created.Poll.Options[0].ID

	voted, err := svc.VoteNewsPoll(ctx, voter.ID, created.ID, intranet.VoteNewsPollRequest{OptionID: optPizza})
	require.NoError(t, err)
	assert.Equal(t, 1, voted.TotalVotes)
	require.NotNil(t, voted.VotedOptionID)
	assert.Equal(t, optPizza, *voted.VotedOptionID)
	assert.Equal(t, 1, voted.Options[0].VotesCount)

	_, err = svc.VoteNewsPoll(ctx, voter.ID, created.ID, intranet.VoteNewsPollRequest{OptionID: optPizza})
	require.ErrorIs(t, err, intranet.ErrAlreadyVoted)

	require.NoError(t, svc.DeleteNews(ctx, admin.ID, created.ID))

	_, err = svc.GetNewsPoll(ctx, voter.ID, created.ID)
	require.ErrorIs(t, err, intranet.ErrNotFound, "poll must be hidden with soft-deleted post")

	auditRepo := audit.NewRepository(client)
	require.NoError(t, auditRepo.RestoreNewsPost(ctx, created.ID))

	restored, err := svc.GetNewsPoll(ctx, voter.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, restored.TotalVotes)
	assert.Equal(t, 1, restored.Options[0].VotesCount)
	assert.Equal(t, "Lunch?", restored.Question)

	// NFR-11: soft-delete 30d — старше cutoff не в trash (не обходим retention).
	require.NoError(t, svc.DeleteNews(ctx, admin.ID, created.ID))
	old := time.Now().UTC().Add(-40 * 24 * time.Hour)
	_, err = client.NewsPost.UpdateOneID(created.ID).SetDeletedAt(old).Save(ctx)
	require.NoError(t, err)
	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)
	items, total, err := auditRepo.ListDeletedNewsPosts(ctx, cutoff, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)
}

// TestCreateNews_PollValidation_Task098 — FR-INTR-01: <2 options → ErrInvalidInput.
func TestCreateNews_PollValidation_Task098(t *testing.T) {
	svc, adminID, _ := setupIntranetService(t)
	_, err := svc.CreateNews(context.Background(), adminID, intranet.CreateNewsRequest{
		Title: "Bad poll", Body: "x", Published: true,
		Poll: &intranet.CreatePollPayload{Question: "Q?", Options: []string{"only-one"}},
	})
	require.ErrorIs(t, err, intranet.ErrInvalidInput)
}
