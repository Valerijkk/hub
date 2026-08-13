package messenger_test

// Тесты репозитория мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/channel"
)

// TestEntRepository_PollLifecycle проверяет слой репозитория: ent репозиторий опрос жизненный цикл.
func TestEntRepository_PollLifecycle(t *testing.T) {
	env := newMessengerRepoEnv(t)
	ch := env.createPublicChannel(t, "poll-lifecycle")

	poll, options, err := env.repo.CreatePollWithOptions(env.ctx, messenger.CreatePollInput{
		ChannelID:     ch.ID,
		CreatorID:     env.owner.ID,
		Question:      "Lunch?",
		IsAnonymous:   true,
		IsMultiChoice: true,
		Options:       []string{"Pizza", "Sushi", "Salad"},
	})
	require.NoError(t, err)
	require.Len(t, options, 3)

	got, err := env.repo.GetPoll(env.ctx, poll.ID)
	require.NoError(t, err)
	assert.Equal(t, poll.ID, got.ID)

	withOpts, opts, err := env.repo.GetPollWithOptions(env.ctx, poll.ID)
	require.NoError(t, err)
	require.Len(t, opts, 3)
	assert.Equal(t, "Lunch?", withOpts.Question)

	require.NoError(t, env.repo.VotePoll(env.ctx, poll.ID, env.peer.ID, []uuid.UUID{options[0].ID, options[1].ID}))
	votes, err := env.repo.ListVotes(env.ctx, poll.ID)
	require.NoError(t, err)
	assert.Len(t, votes, 2)

	require.NoError(t, env.repo.VotePoll(env.ctx, poll.ID, env.peer.ID, nil))
	votes, err = env.repo.ListVotes(env.ctx, poll.ID)
	require.NoError(t, err)
	assert.Empty(t, votes)

	require.NoError(t, env.repo.ClosePoll(env.ctx, poll.ID))
	closed, err := env.repo.GetPoll(env.ctx, poll.ID)
	require.NoError(t, err)
	assert.True(t, closed.IsClosed)
}

// TestEntRepository_FindGlobalChannelByName проверяет слой репозитория: ent репозиторий find global channel by name.
func TestEntRepository_FindGlobalChannelByName(t *testing.T) {
	env := newMessengerRepoEnv(t)

	_, err := env.client.Channel.Create().
		SetName("marketplace").
		SetChannelType(channel.ChannelTypePublic).
		SetCreatedBy(env.owner.ID).
		Save(env.ctx)
	require.NoError(t, err)

	found, err := env.repo.FindGlobalChannelByName(env.ctx, "marketplace")
	require.NoError(t, err)
	assert.Equal(t, "marketplace", found.Name)

	_, err = env.repo.FindGlobalChannelByName(env.ctx, "missing")
	require.Error(t, err)
}

// TestHandler_Wave92_DeleteChannelAndAddMembers проверяет HTTP-слой: HTTP-обработчик wave 92 удаление channel and add members.
func TestHandler_Wave92_DeleteChannelAndAddMembers(t *testing.T) {
	app, repo, userID := setupMessengerHandlerApp(t)
	channelID := setupChannelWithMembers(repo, "public", false, userID)
	newMember := uuid.New()

	body, _ := json.Marshal(map[string]any{"member_ids": []string{newMember.String()}})
	req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	isMember, err := repo.IsMember(t.Context(), channelID, newMember)
	require.NoError(t, err)
	assert.True(t, isMember)

	req = httptest.NewRequest(http.MethodDelete, "/api/messenger/channels/"+channelID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
