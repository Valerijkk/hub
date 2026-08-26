package messenger_test

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
)

var errPhase151UnreadRedisForbidden = errors.New("phase151 unread Redis authority is forbidden")

// phase151UnreadRedisPoison intercepts every go-redis operation before dial.
// A non-zero count proves production still attempted to use legacy Redis unread truth.
type phase151UnreadRedisPoison struct {
	attempts atomic.Int64
}

func (p *phase151UnreadRedisPoison) DialHook(_ redis.DialHook) redis.DialHook {
	return func(context.Context, string, string) (net.Conn, error) {
		p.attempts.Add(1)
		return nil, errPhase151UnreadRedisForbidden
	}
}

func (p *phase151UnreadRedisPoison) ProcessHook(_ redis.ProcessHook) redis.ProcessHook {
	return func(context.Context, redis.Cmder) error {
		p.attempts.Add(1)
		return errPhase151UnreadRedisForbidden
	}
}

func (p *phase151UnreadRedisPoison) ProcessPipelineHook(_ redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(context.Context, []redis.Cmder) error {
		p.attempts.Add(1)
		return errPhase151UnreadRedisForbidden
	}
}

func (p *phase151UnreadRedisPoison) assertUntouched(t *testing.T, operation string) {
	t.Helper()
	assert.Zerof(t, p.attempts.Load(), "%s must not execute a Redis unread operation", operation)
}

func newPhase151RedisRetirementService(t *testing.T, repo *mockRepository) (*messenger.MessengerService, *phase151UnreadRedisPoison) {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: "phase151-unread-redis-forbidden.invalid:0"})
	poison := &phase151UnreadRedisPoison{}
	client.AddHook(poison)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	return messenger.NewMessengerService(repo, nil, nil, messenger.NewUnreadService(client)), poison
}

func TestPhase151LegacyUnreadReadReceiptFailsClosedAndIgnoresLostRedis(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	actorID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, actorID)
	baseline := messenger.NewMessengerService(repo, nil, nil, nil)
	wantCounts, err := baseline.GetUnreadCounts(ctx, actorID.String())
	require.NoError(t, err)
	require.Error(t, baseline.MarkChannelRead(ctx, actorID.String(), channelID.String()))

	service, poison := newPhase151RedisRetirementService(t, repo)
	gotCounts, err := service.GetUnreadCounts(ctx, actorID.String())
	require.NoError(t, err)
	require.Error(t, service.MarkChannelRead(ctx, actorID.String(), channelID.String()))
	assert.Equal(t, wantCounts, gotCounts)
	poison.assertUntouched(t, "GetUnreadCounts and MarkChannelRead")
}

func TestPhase151CanonicalWritersAndChannelCleanupIgnoreRedisUnread(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID, recipientID := uuid.New(), uuid.New()
	sourceChannelID := setupChannelWithMembers(repo, "public", false, senderID)
	targetChannelID := setupChannelWithMembers(repo, "public", false, senderID, recipientID)
	repo.messages["phase151-redis-forward-source"] = &messenger.MessageDocument{
		ID: "phase151-redis-forward-source", ChannelID: sourceChannelID, SenderID: uuid.New(), ContentText: "forward source",
	}
	service, poison := newPhase151RedisRetirementService(t, repo)

	created, err := service.SendMessage(ctx, targetChannelID, senderID, messenger.SendMessageRequest{
		ContentText: "ordinary post", ClientNonce: "phase151-redis-ordinary-send",
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	forwarded, err := service.ForwardMessage(ctx, sourceChannelID, senderID, "phase151-redis-forward-source", targetChannelID.String())
	require.NoError(t, err)
	require.NotNil(t, forwarded)
	poison.assertUntouched(t, "user post and forward")
}

func TestPhase151BotLeaveAndDeleteIgnoreRedisUnread(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	botID, memberID := uuid.New(), uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, botID, memberID)
	service, poison := newPhase151RedisRetirementService(t, repo)
	service.SetSystemBotID(botID)
	message, err := service.SendBotMessage(ctx, channelID, messenger.SendMessageRequest{ContentText: "bot send"})
	require.NoError(t, err)
	require.NotNil(t, message)
	require.NoError(t, service.LeaveChannel(ctx, channelID, memberID))
	require.NoError(t, service.DeleteChannel(ctx, channelID, botID))
	poison.assertUntouched(t, "bot send, leave, delete")
}

func TestPhase151UnreadCutoverNeverReadsOrImportsRedisCounters(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	actorID := uuid.New()
	setupChannelWithMembers(repo, "public", false, actorID)
	service, poison := newPhase151RedisRetirementService(t, repo)
	result, err := messenger.ExportPhase151RunUnreadCutover(ctx, service, actorID)
	require.Error(t, err)
	assert.False(t, result.BaselinedAuthorizedTargets)
	assert.False(t, result.ImportedRedisCounters)
	poison.assertUntouched(t, "authorized unread cutover")
}

func TestPhase151ServerUnreadDependencyContractKeepsPresenceRedisOnly(t *testing.T) {
	contract := messenger.ExportPhase151UnreadDependencyContract()
	assert.True(t, contract.PresenceRedisAllowed)
	assert.False(t, contract.UnreadRedisAccepted)
	assert.False(t, contract.DashboardUnreadRedisAccepted)
	assert.True(t, contract.DurableUnreadRepositoryRequired)
	assert.True(t, contract.AuthorizedUnreadScopeRequired)
}

type phase151DashboardUnreadRepository struct {
	messageIDs map[messenger.UnreadCategory][]string
	calls      []messenger.UnreadCategory
}

func (r *phase151DashboardUnreadRepository) ListAuthorizedUnreadMessageIDs(
	_ context.Context, _ messenger.AuthorizedUnreadScope, category messenger.UnreadCategory,
) ([]string, error) {
	r.calls = append(r.calls, category)
	return append([]string(nil), r.messageIDs[category]...), nil
}

func TestPhase151DashboardUnreadSummary_DeduplicatesAuthorizedMessagesAcrossCategories(t *testing.T) {
	userID := uuid.New()
	scope := &phase151UnreadScopeFixture{scope: messenger.AuthorizedUnreadScope{TenantID: uuid.New(), UserID: userID}}
	repo := &phase151DashboardUnreadRepository{messageIDs: map[messenger.UnreadCategory][]string{
		messenger.UnreadCategoryChannel: {"m1", "m2"}, messenger.UnreadCategoryDM: {"m2", "m3"},
		messenger.UnreadCategoryMention: {"m1", "m4"}, messenger.UnreadCategoryFollowedThread: {"m4", "m5"},
	}}
	summary, err := messenger.ExportPhase151BuildDashboardUnreadSummary(context.Background(), userID, scope, repo)
	require.NoError(t, err)
	assert.Equal(t, messenger.DashboardUnreadSummary{UniqueMessageCount: 5, HasUnread: true}, summary)
	assert.Equal(t, []uuid.UUID{userID}, scope.calls)
	assert.ElementsMatch(t, []messenger.UnreadCategory{
		messenger.UnreadCategoryChannel, messenger.UnreadCategoryDM,
		messenger.UnreadCategoryMention, messenger.UnreadCategoryFollowedThread,
	}, repo.calls)
}

func TestPhase151DashboardUnreadSummary_DeniedScopeFailsBeforeMongoAndRedis(t *testing.T) {
	userID := uuid.New()
	scope := &phase151UnreadScopeFixture{err: errors.New("unread scope denied")}
	repo := &phase151DashboardUnreadRepository{}
	summary, err := messenger.ExportPhase151BuildDashboardUnreadSummary(context.Background(), userID, scope, repo)
	require.Error(t, err)
	assert.Equal(t, messenger.DashboardUnreadSummary{}, summary)
	assert.Empty(t, repo.calls)
}
