package messenger_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151AutoFollowPort interface {
	EnsureAutoThreadFollows(context.Context, messenger.AuthorizedMessageProducerScope, *messenger.MessageDocument) error
}

var _ phase151AutoFollowPort = (*messenger.CombinedRepository)(nil)

type phase151AutoFollowScopeResult struct {
	scope messenger.AuthorizedUnreadScope
	err   error
}

type phase151AutoFollowResolver struct {
	results map[uuid.UUID]phase151AutoFollowScopeResult
	calls   []uuid.UUID
}

func (r *phase151AutoFollowResolver) ResolveAuthorizedUnreadScope(
	_ context.Context,
	userID uuid.UUID,
) (messenger.AuthorizedUnreadScope, error) {
	r.calls = append(r.calls, userID)
	result, ok := r.results[userID]
	if !ok {
		return messenger.AuthorizedUnreadScope{}, messenger.ErrAuthorizedUnreadScopeDenied
	}
	return result.scope, result.err
}

type phase151AutoFollowWrite struct {
	scope      messenger.AuthorizedUnreadScope
	channelID  uuid.UUID
	rootID     string
	transition messenger.ThreadFollowTransition
}

type phase151AutoFollowWriter struct {
	calls     []phase151AutoFollowWrite
	failUser  uuid.UUID
	failError error
}

func (w *phase151AutoFollowWriter) ApplyAuthorizedThreadFollowTransition(
	_ context.Context,
	scope messenger.AuthorizedUnreadScope,
	channelID uuid.UUID,
	rootID string,
	transition messenger.ThreadFollowTransition,
) (bool, error) {
	w.calls = append(w.calls, phase151AutoFollowWrite{
		scope: scope, channelID: channelID, rootID: rootID, transition: transition,
	})
	if scope.UserID == w.failUser {
		return false, w.failError
	}
	return true, nil
}

func phase151AutoFollowScopeAndMessage() (messenger.AuthorizedMessageProducerScope, *messenger.MessageDocument) {
	actorID := uuid.New()
	channelID := uuid.New()
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: actorID, ChannelID: channelID,
		Producer: messenger.TrustedMessageProducerUser,
	}
	message := &messenger.MessageDocument{
		ID: "root-1", ChannelID: channelID, SenderID: actorID,
		CreatedAt:              time.Date(2026, 8, 27, 12, 30, 0, 123, time.FixedZone("phase151", 3*60*60)),
		PostCreateEffectsState: messenger.PostCreateEffectsStatePending,
	}
	return scope, message
}

func phase151AuthorizedAutoFollowScope(
	tenantID, userID, channelID uuid.UUID,
) messenger.AuthorizedUnreadScope {
	return messenger.AuthorizedUnreadScope{
		TenantID: tenantID,
		UserID:   userID,
		Targets:  []messenger.AuthorizedUnreadTarget{{ChannelID: channelID, Kind: "channel"}},
	}
}

func TestPhase151AutoFollowPlanIncludesAuthorAndUniqueResolvedMentions(t *testing.T) {
	scope, message := phase151AutoFollowScopeAndMessage()
	firstMention := uuid.New()
	secondMention := uuid.New()
	message.ResolvedMentionUserIDs = []uuid.UUID{
		firstMention, scope.ActorID, firstMention, secondMention,
	}

	plan, err := messenger.ExportPhase151BuildAutoThreadFollowPlan(scope, message)
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{scope.ActorID, firstMention, secondMention}, plan.UserIDs)
	assert.Equal(t, scope.ChannelID, plan.ChannelID)
	assert.Equal(t, message.ID, plan.RootMessageID)
	assert.Equal(t, messenger.ThreadFollowModeAuto, plan.Transition.Mode)
	assert.Equal(t, messenger.MessagePosition{
		CreatedAt: message.CreatedAt.UTC(), MessageID: message.ID,
	}, plan.Transition.Position)

	message.ID = "reply-1"
	message.ThreadID = "root-1"
	plan, err = messenger.ExportPhase151BuildAutoThreadFollowPlan(scope, message)
	require.NoError(t, err)
	assert.Equal(t, "root-1", plan.RootMessageID)
	assert.Equal(t, "reply-1", plan.Transition.Position.MessageID)
}

func TestPhase151AutoFollowReauthorizesEveryCandidateBeforeMongo(t *testing.T) {
	scope, message := phase151AutoFollowScopeAndMessage()
	firstMention := uuid.New()
	secondMention := uuid.New()
	message.ResolvedMentionUserIDs = []uuid.UUID{firstMention, secondMention}
	resolver := &phase151AutoFollowResolver{results: map[uuid.UUID]phase151AutoFollowScopeResult{}}
	for _, userID := range []uuid.UUID{scope.ActorID, firstMention, secondMention} {
		resolver.results[userID] = phase151AutoFollowScopeResult{
			scope: phase151AuthorizedAutoFollowScope(scope.TenantID, userID, scope.ChannelID),
		}
	}
	writer := &phase151AutoFollowWriter{}

	err := messenger.ExportPhase151EnsureAutoThreadFollows(context.Background(), scope, message, resolver, writer)
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{scope.ActorID, firstMention, secondMention}, resolver.calls)
	require.Len(t, writer.calls, 3)
	for index, userID := range []uuid.UUID{scope.ActorID, firstMention, secondMention} {
		assert.Equal(t, userID, writer.calls[index].scope.UserID)
		assert.Equal(t, scope.TenantID, writer.calls[index].scope.TenantID)
		assert.Equal(t, scope.ChannelID, writer.calls[index].channelID)
		assert.Equal(t, message.ID, writer.calls[index].rootID)
		assert.Equal(t, messenger.ThreadFollowModeAuto, writer.calls[index].transition.Mode)
		assert.Equal(t, message.ID, writer.calls[index].transition.Position.MessageID)
		assert.Equal(t, message.CreatedAt.UTC(), writer.calls[index].transition.Position.CreatedAt)
	}
}

func TestPhase151AutoFollowSkipsFreshDenialAndMissingChannelTarget(t *testing.T) {
	scope, message := phase151AutoFollowScopeAndMessage()
	missingTarget := uuid.New()
	eligibleMention := uuid.New()
	message.ResolvedMentionUserIDs = []uuid.UUID{missingTarget, eligibleMention}
	resolver := &phase151AutoFollowResolver{results: map[uuid.UUID]phase151AutoFollowScopeResult{
		scope.ActorID: {err: fmt.Errorf("actor revoked: %w", messenger.ErrAuthorizedUnreadScopeDenied)},
		missingTarget: {scope: messenger.AuthorizedUnreadScope{
			TenantID: scope.TenantID, UserID: missingTarget,
		}},
		eligibleMention: {
			scope: phase151AuthorizedAutoFollowScope(scope.TenantID, eligibleMention, scope.ChannelID),
		},
	}}
	writer := &phase151AutoFollowWriter{}

	err := messenger.ExportPhase151EnsureAutoThreadFollows(context.Background(), scope, message, resolver, writer)
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{scope.ActorID, missingTarget, eligibleMention}, resolver.calls)
	require.Len(t, writer.calls, 1)
	assert.Equal(t, eligibleMention, writer.calls[0].scope.UserID)
}

func TestPhase151AutoFollowFailsClosedOnAuthorityOrWriterInfrastructure(t *testing.T) {
	scope, message := phase151AutoFollowScopeAndMessage()
	mentionID := uuid.New()
	message.ResolvedMentionUserIDs = []uuid.UUID{mentionID}

	t.Run("authority infrastructure", func(t *testing.T) {
		resolver := &phase151AutoFollowResolver{results: map[uuid.UUID]phase151AutoFollowScopeResult{
			scope.ActorID: {
				scope: phase151AuthorizedAutoFollowScope(scope.TenantID, scope.ActorID, scope.ChannelID),
			},
			mentionID: {err: errors.New("postgres unavailable")},
		}}
		writer := &phase151AutoFollowWriter{}

		err := messenger.ExportPhase151EnsureAutoThreadFollows(context.Background(), scope, message, resolver, writer)
		require.ErrorContains(t, err, "resolve automatic thread-follow authority")
		require.Len(t, writer.calls, 1)
		assert.Equal(t, scope.ActorID, writer.calls[0].scope.UserID)
	})

	for _, scenario := range []struct {
		name  string
		scope messenger.AuthorizedUnreadScope
	}{
		{
			name:  "wrong user",
			scope: phase151AuthorizedAutoFollowScope(scope.TenantID, mentionID, scope.ChannelID),
		},
		{
			name:  "wrong tenant",
			scope: phase151AuthorizedAutoFollowScope(uuid.New(), scope.ActorID, scope.ChannelID),
		},
	} {
		scenario := scenario
		t.Run("malformed authority "+scenario.name, func(t *testing.T) {
			resolver := &phase151AutoFollowResolver{results: map[uuid.UUID]phase151AutoFollowScopeResult{
				scope.ActorID: {scope: scenario.scope},
			}}
			writer := &phase151AutoFollowWriter{}

			err := messenger.ExportPhase151EnsureAutoThreadFollows(context.Background(), scope, message, resolver, writer)
			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Empty(t, writer.calls)
		})
	}

	t.Run("writer infrastructure", func(t *testing.T) {
		resolver := &phase151AutoFollowResolver{results: map[uuid.UUID]phase151AutoFollowScopeResult{
			scope.ActorID: {
				scope: phase151AuthorizedAutoFollowScope(scope.TenantID, scope.ActorID, scope.ChannelID),
			},
		}}
		writer := &phase151AutoFollowWriter{failUser: scope.ActorID, failError: errors.New("mongo unavailable")}

		err := messenger.ExportPhase151EnsureAutoThreadFollows(context.Background(), scope, message, resolver, writer)
		require.ErrorContains(t, err, "persist automatic thread follow")
		require.Len(t, writer.calls, 1)
	})
}

func TestPhase151AutoFollowRejectsUntrustedProducerBeforeAuthority(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		mutate func(*messenger.AuthorizedMessageProducerScope)
	}{
		{name: "unknown", mutate: func(scope *messenger.AuthorizedMessageProducerScope) {
			scope.Producer = messenger.TrustedMessageProducer("unknown")
		}},
		{name: "system bot without bot proof", mutate: func(scope *messenger.AuthorizedMessageProducerScope) {
			scope.Producer = messenger.TrustedMessageProducerSystemBot
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope, message := phase151AutoFollowScopeAndMessage()
			scenario.mutate(&scope)
			resolver := &phase151AutoFollowResolver{results: map[uuid.UUID]phase151AutoFollowScopeResult{}}
			writer := &phase151AutoFollowWriter{}

			err := messenger.ExportPhase151EnsureAutoThreadFollows(context.Background(), scope, message, resolver, writer)
			require.ErrorIs(t, err, messenger.ErrTrustedMessageProducerDenied)
			assert.Empty(t, resolver.calls)
			assert.Empty(t, writer.calls)
		})
	}
}
func TestPhase151AutoFollowRejectsInvalidCanonicalStateBeforeAuthority(t *testing.T) {
	scope, validMessage := phase151AutoFollowScopeAndMessage()
	now := time.Now().UTC()
	for _, scenario := range []struct {
		name   string
		mutate func(*messenger.MessageDocument)
	}{
		{name: "missing id", mutate: func(message *messenger.MessageDocument) { message.ID = "" }},
		{name: "foreign channel", mutate: func(message *messenger.MessageDocument) { message.ChannelID = uuid.New() }},
		{name: "foreign sender", mutate: func(message *messenger.MessageDocument) { message.SenderID = uuid.New() }},
		{name: "self root", mutate: func(message *messenger.MessageDocument) { message.ThreadID = message.ID }},
		{name: "missing created time", mutate: func(message *messenger.MessageDocument) { message.CreatedAt = time.Time{} }},
		{name: "deleted", mutate: func(message *messenger.MessageDocument) { message.DeletedAt = &now }},
		{name: "complete", mutate: func(message *messenger.MessageDocument) {
			message.PostCreateEffectsState = messenger.PostCreateEffectsStateComplete
		}},
		{name: "completed timestamp", mutate: func(message *messenger.MessageDocument) {
			message.PostCreateEffectsCompletedAt = &now
		}},
		{name: "nil resolved mention", mutate: func(message *messenger.MessageDocument) {
			message.ResolvedMentionUserIDs = []uuid.UUID{uuid.Nil}
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			message := *validMessage
			scenario.mutate(&message)
			resolver := &phase151AutoFollowResolver{results: map[uuid.UUID]phase151AutoFollowScopeResult{}}
			writer := &phase151AutoFollowWriter{}

			err := messenger.ExportPhase151EnsureAutoThreadFollows(context.Background(), scope, &message, resolver, writer)
			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Empty(t, resolver.calls)
			assert.Empty(t, writer.calls)
		})
	}
}

func TestPhase151AutoFollowRequiresBothFreshAuthorityDependencies(t *testing.T) {
	scope, message := phase151AutoFollowScopeAndMessage()
	resolver := &phase151AutoFollowResolver{results: map[uuid.UUID]phase151AutoFollowScopeResult{}}
	writer := &phase151AutoFollowWriter{}

	err := messenger.ExportPhase151EnsureAutoThreadFollows(nil, scope, message, resolver, writer)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	err = messenger.ExportPhase151EnsureAutoThreadFollows(context.Background(), scope, message, nil, writer)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	err = messenger.ExportPhase151EnsureAutoThreadFollows(context.Background(), scope, message, resolver, nil)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
}
