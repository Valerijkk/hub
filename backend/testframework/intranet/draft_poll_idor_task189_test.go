//go:build hubtest

package intranet_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/intranet"
)

// TestDraftNewsPoll_GetVoteHiddenFromNonOwner_Task189 — BE-V4-003:
// unpublished post: member GET/vote poll → ErrNotFound; author/admin OK.
// Mutation: убрать requirePublishedNewsOrAuthorOrAdmin → member читает draft FAIL.
func TestDraftNewsPoll_GetVoteHiddenFromNonOwner_Task189(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:t189_draft_poll_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("t189-admin-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	author, err := client.User.Create().
		SetEmail("t189-author-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	member, err := client.User.Create().
		SetEmail("t189-member-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	svc := intranet.NewService(intranet.NewEntRepository(client), nil, nil)

	post, err := client.NewsPost.Create().
		SetTitle("Draft poll").
		SetBody("secret").
		SetPublished(false).
		SetAuthorID(author.ID).
		Save(ctx)
	require.NoError(t, err)

	poll, err := client.NewsPoll.Create().
		SetNewsPostID(post.ID).
		SetQuestion("Draft Q?").
		Save(ctx)
	require.NoError(t, err)
	optA, err := client.NewsPollOption.Create().
		SetPollID(poll.ID).
		SetText("Yes").
		SetPosition(0).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.NewsPollOption.Create().
		SetPollID(poll.ID).
		SetText("No").
		SetPosition(1).
		Save(ctx)
	require.NoError(t, err)

	_, err = svc.GetNewsPoll(ctx, member.ID, post.ID)
	assert.ErrorIs(t, err, intranet.ErrNotFound, "member must not read draft poll")

	_, err = svc.VoteNewsPoll(ctx, member.ID, post.ID, intranet.VoteNewsPollRequest{OptionID: optA.ID})
	assert.ErrorIs(t, err, intranet.ErrNotFound, "member must not vote draft poll")

	got, err := svc.GetNewsPoll(ctx, author.ID, post.ID)
	require.NoError(t, err)
	assert.Equal(t, "Draft Q?", got.Question)

	voted, err := svc.VoteNewsPoll(ctx, author.ID, post.ID, intranet.VoteNewsPollRequest{OptionID: optA.ID})
	require.NoError(t, err)
	require.NotNil(t, voted.VotedOptionID)
	assert.Equal(t, optA.ID, *voted.VotedOptionID)

	adminGot, err := svc.GetNewsPoll(ctx, admin.ID, post.ID)
	require.NoError(t, err)
	assert.Equal(t, "Draft Q?", adminGot.Question)
}

// TestPublishedNewsPoll_MemberCanGetAndVote_Task189 — регрессия: published остаётся открытым.
func TestPublishedNewsPoll_MemberCanGetAndVote_Task189(t *testing.T) {
	svc, adminID, memberID := setupIntranetService(t)
	ctx := context.Background()

	created, err := svc.CreateNews(ctx, adminID, intranet.CreateNewsRequest{
		Title:     "Public poll",
		Body:      "body",
		Published: true,
		Poll: &intranet.CreatePollPayload{
			Question: "Public?",
			Options:  []string{"A", "B"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, created.Poll)
	require.Len(t, created.Poll.Options, 2)

	got, err := svc.GetNewsPoll(ctx, memberID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Public?", got.Question)

	voted, err := svc.VoteNewsPoll(ctx, memberID, created.ID, intranet.VoteNewsPollRequest{
		OptionID: created.Poll.Options[0].ID,
	})
	require.NoError(t, err)
	require.NotNil(t, voted.VotedOptionID)
}
