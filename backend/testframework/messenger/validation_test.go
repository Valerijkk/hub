package messenger_test

// Тесты модуля мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseExpiresIn проверяет сценарий: разбор expires in.
func TestParseExpiresIn(t *testing.T) {
	t.Run("empty is permanent", func(t *testing.T) {
		exp, err := messenger.ExportParseExpiresIn("")
		require.NoError(t, err)
		assert.Nil(t, exp)
	})

	t.Run("supported durations", func(t *testing.T) {
		for _, tc := range []struct {
			in      string
			minDiff time.Duration
			maxDiff time.Duration
		}{
			{"1h", 59 * time.Minute, 61 * time.Minute},
			{"1d", 23 * time.Hour, 25 * time.Hour},
			{"7d", 6 * 24 * time.Hour, 8 * 24 * time.Hour},
			{"30d", 29 * 24 * time.Hour, 31 * 24 * time.Hour},
		} {
			exp, err := messenger.ExportParseExpiresIn(tc.in)
			require.NoError(t, err, tc.in)
			require.NotNil(t, exp)
			diff := exp.Sub(time.Now().UTC())
			assert.Greater(t, diff, tc.minDiff)
			assert.Less(t, diff, tc.maxDiff)
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		_, err := messenger.ExportParseExpiresIn("2w")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "некорректный срок")
	})
}

// TestGenerateInviteToken проверяет сценарий: generate приглашение token.
func TestGenerateInviteToken(t *testing.T) {
	token, err := messenger.ExportGenerateInviteToken()
	require.NoError(t, err)
	assert.Len(t, token, messenger.ExportInviteTokenLength())
	for _, c := range token {
		assert.Contains(t, messenger.ExportBase62Chars(), string(c))
	}

	other, err := messenger.ExportGenerateInviteToken()
	require.NoError(t, err)
	assert.NotEqual(t, token, other)
}

// TestCreateChannel_InvalidProjectID проверяет отклонение при создание channel некорректный project id.
func TestCreateChannel_InvalidProjectID(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.CreateChannel(ctx, uuid.New(), messenger.CreateChannelRequest{
		ChannelType: "public",
		Name:        "bad-project",
		ProjectID:   "not-a-uuid",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project_id")
}

// TestCreatePoll_Validation проверяет сценарий: создание опрос валидация.
func TestCreatePoll_Validation(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	t.Run("too few options", func(t *testing.T) {
		_, err := svc.CreatePoll(ctx, channelID, ownerID, messenger.CreatePollRequest{
			Question: "Q?",
			Options:  []string{"only one"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "минимум 2")
	})

	t.Run("too many options", func(t *testing.T) {
		opts := make([]string, 11)
		for i := range opts {
			opts[i] = "opt"
		}
		_, err := svc.CreatePoll(ctx, channelID, ownerID, messenger.CreatePollRequest{
			Question: "Q?",
			Options:  opts,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "максимум 10")
	})

	t.Run("empty question", func(t *testing.T) {
		_, err := svc.CreatePoll(ctx, channelID, ownerID, messenger.CreatePollRequest{
			Question: "",
			Options:  []string{"a", "b"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "вопрос")
	})

	t.Run("non-member forbidden", func(t *testing.T) {
		_, err := svc.CreatePoll(ctx, channelID, uuid.New(), messenger.CreatePollRequest{
			Question: "Q?",
			Options:  []string{"a", "b"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "участником")
	})
}

// TestCreateInviteLink_InvalidExpiresIn проверяет отклонение при создание приглашение link некорректный expires in.
func TestCreateInviteLink_InvalidExpiresIn(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	_, err := svc.CreateInviteLink(ctx, channelID, ownerID, messenger.CreateInviteLinkRequest{
		ExpiresIn: "forever",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "срок действия")
}
