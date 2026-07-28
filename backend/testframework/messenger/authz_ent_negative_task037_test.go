//go:build hubtest

package messenger_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
)

// TASK-037 / P2-BE-02 — Ent IsMember denial (не mockRepository.IsMember=true).
// Mutation: убрать IsMember check в SendMessage/AddBookmark → outsider проходит.

func TestSendMessage_RejectsNonMember_Ent_TASK037(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "ent-send-idor")
	combined := messenger.NewCombinedRepository(env.repo, nil)
	svc := messenger.ExportNewMessengerServiceTest(combined)

	_, err := svc.SendMessage(env.ctx, ch.ID, env.peer.ID, messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "pwn"},
		ContentText: "pwn",
	})
	requireMessengerAppErrorCode(t, err, 403)

	ok, err := env.repo.IsMember(env.ctx, ch.ID, env.peer.ID)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestAddBookmark_RejectsNonMember_Ent_TASK037(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "ent-bm-idor")
	combined := messenger.NewCombinedRepository(env.repo, nil)
	svc := messenger.ExportNewMessengerServiceTest(combined)

	err := svc.AddBookmark(env.ctx, env.peer.ID, "msg-ent-idor", ch.ID)
	requireMessengerAppErrorCode(t, err, 403)

	list, err := env.repo.ListBookmarks(env.ctx, env.peer.ID)
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestIsMember_Ent_RejectsStranger_TASK037(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "ent-ismember")

	ok, err := env.repo.IsMember(env.ctx, ch.ID, env.owner.ID)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = env.repo.IsMember(env.ctx, ch.ID, env.peer.ID)
	require.NoError(t, err)
	require.False(t, ok)
}
