package messenger_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151UnreadInvalidationPort interface {
	PublishUnreadInvalidation(context.Context, messenger.AuthorizedMessageProducerScope, *messenger.MessageDocument) error
}

var _ phase151UnreadInvalidationPort = (*messenger.Phase151UnreadInvalidationEffect)(nil)
var _ messenger.Phase151UnreadInvalidationAudience = (*messenger.WSChannelAuthority)(nil)

type phase151UnreadInvalidationAudienceCall struct {
	channelID uuid.UUID
	event     string
}

type phase151UnreadInvalidationAudience struct {
	recipients []uuid.UUID
	err        error
	calls      []phase151UnreadInvalidationAudienceCall
}

func (a *phase151UnreadInvalidationAudience) ResolveChannelEventAudience(
	_ context.Context,
	channelID uuid.UUID,
	event string,
) ([]uuid.UUID, error) {
	a.calls = append(a.calls, phase151UnreadInvalidationAudienceCall{channelID: channelID, event: event})
	return append([]uuid.UUID(nil), a.recipients...), a.err
}

type phase151PublishedEvent struct {
	subject string
	payload any
}

type phase151UnreadInvalidationPublisher struct {
	calls []phase151PublishedEvent
}

func (p *phase151UnreadInvalidationPublisher) PublishJSON(subject string, payload any) {
	p.calls = append(p.calls, phase151PublishedEvent{subject: subject, payload: payload})
}

func phase151UnreadInvalidationScopeAndMessage() (messenger.AuthorizedMessageProducerScope, *messenger.MessageDocument) {
	actorID := uuid.New()
	channelID := uuid.New()
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: actorID, ChannelID: channelID,
		Producer: messenger.TrustedMessageProducerUser,
	}
	message := &messenger.MessageDocument{
		ID: uuid.New().String(), ChannelID: channelID, SenderID: actorID,
		CreatedAt: time.Now().UTC(), PostCreateEffectsState: messenger.PostCreateEffectsStatePending,
	}
	return scope, message
}

func TestPhase151UnreadInvalidationPlanUsesStrictRootOrThreadIdentifiers(t *testing.T) {
	scope, message := phase151UnreadInvalidationScopeAndMessage()

	rootPlan, err := messenger.ExportPhase151BuildUnreadInvalidationPlan(scope, message)
	require.NoError(t, err)
	assert.Equal(t, scope.ChannelID, rootPlan.ChannelID)
	assert.Equal(t, scope.ActorID, rootPlan.SenderID)
	assert.Equal(t, messenger.UnreadChangedEvent{
		ChannelID:  scope.ChannelID,
		TargetKind: "channel",
		Category:   "channel",
	}, rootPlan.Payload)

	rootID := uuid.New()
	message.ThreadID = rootID.String()
	threadPlan, err := messenger.ExportPhase151BuildUnreadInvalidationPlan(scope, message)
	require.NoError(t, err)
	assert.Equal(t, messenger.UnreadChangedEvent{
		ChannelID:     scope.ChannelID,
		TargetKind:    "thread",
		Category:      "followed_thread",
		RootMessageID: &rootID,
	}, threadPlan.Payload)
	encoded, marshalErr := json.Marshal(threadPlan.Payload)
	require.NoError(t, marshalErr)
	var fields map[string]any
	require.NoError(t, json.Unmarshal(encoded, &fields))
	assert.Equal(t, map[string]any{
		"channel_id":      scope.ChannelID.String(),
		"target_kind":     "thread",
		"category":        "followed_thread",
		"root_message_id": rootID.String(),
	}, fields)
}

func TestPhase151UnreadInvalidationPublishesIdentifiersOncePerCurrentRecipient(t *testing.T) {
	scope, message := phase151UnreadInvalidationScopeAndMessage()
	firstRecipient := uuid.New()
	secondRecipient := uuid.New()
	audience := &phase151UnreadInvalidationAudience{recipients: []uuid.UUID{
		scope.ActorID, firstRecipient, uuid.Nil, firstRecipient, secondRecipient,
	}}
	publisher := &phase151UnreadInvalidationPublisher{}

	effect := messenger.NewPhase151UnreadInvalidationEffect(audience, publisher)
	err := effect.PublishUnreadInvalidation(context.Background(), scope, message)
	require.NoError(t, err)
	assert.Equal(t, []phase151UnreadInvalidationAudienceCall{{
		channelID: scope.ChannelID, event: messenger.UserUnreadChangedEvent,
	}}, audience.calls)
	require.Len(t, publisher.calls, 2)
	assert.Equal(t, []string{
		"messenger.user." + firstRecipient.String() + ".unread_changed",
		"messenger.user." + secondRecipient.String() + ".unread_changed",
	}, []string{publisher.calls[0].subject, publisher.calls[1].subject})
	for _, published := range publisher.calls {
		payload, ok := published.payload.(messenger.UnreadChangedEvent)
		require.True(t, ok)
		assert.Equal(t, scope.ChannelID, payload.ChannelID)
		assert.Equal(t, "channel", payload.TargetKind)
		assert.Equal(t, "channel", payload.Category)
		assert.Nil(t, payload.RootMessageID)

		encoded, marshalErr := json.Marshal(payload)
		require.NoError(t, marshalErr)
		var fields map[string]any
		require.NoError(t, json.Unmarshal(encoded, &fields))
		assert.Equal(t, map[string]any{
			"channel_id":  scope.ChannelID.String(),
			"target_kind": "channel",
			"category":    "channel",
		}, fields)
	}
}

func TestPhase151UnreadInvalidationAudienceFailurePublishesNothing(t *testing.T) {
	scope, message := phase151UnreadInvalidationScopeAndMessage()
	audience := &phase151UnreadInvalidationAudience{err: errors.New("postgres audience unavailable")}
	publisher := &phase151UnreadInvalidationPublisher{}

	err := messenger.ExportPhase151PublishUnreadInvalidation(
		context.Background(), scope, message, audience, publisher,
	)
	require.ErrorContains(t, err, "resolve unread invalidation audience")
	require.Len(t, audience.calls, 1)
	assert.Empty(t, publisher.calls)
}

func TestPhase151UnreadInvalidationRejectsInvalidStateBeforeAudience(t *testing.T) {
	validScope, validMessage := phase151UnreadInvalidationScopeAndMessage()
	now := time.Now().UTC()
	for _, scenario := range []struct {
		name        string
		mutateScope func(*messenger.AuthorizedMessageProducerScope)
		mutate      func(*messenger.MessageDocument)
		expected    error
	}{
		{name: "unknown producer", mutateScope: func(scope *messenger.AuthorizedMessageProducerScope) {
			scope.Producer = messenger.TrustedMessageProducer("unknown")
		}, expected: messenger.ErrTrustedMessageProducerDenied},
		{name: "system bot without proof", mutateScope: func(scope *messenger.AuthorizedMessageProducerScope) {
			scope.Producer = messenger.TrustedMessageProducerSystemBot
		}, expected: messenger.ErrTrustedMessageProducerDenied},
		{name: "missing id", mutate: func(message *messenger.MessageDocument) { message.ID = "" }, expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "missing created time", mutate: func(message *messenger.MessageDocument) { message.CreatedAt = time.Time{} }, expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "foreign channel", mutate: func(message *messenger.MessageDocument) { message.ChannelID = uuid.New() }, expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "foreign sender", mutate: func(message *messenger.MessageDocument) { message.SenderID = uuid.New() }, expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "deleted", mutate: func(message *messenger.MessageDocument) { message.DeletedAt = &now }, expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "complete in-memory state", mutate: func(message *messenger.MessageDocument) {
			message.PostCreateEffectsState = messenger.PostCreateEffectsStateComplete
		}, expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "completed timestamp", mutate: func(message *messenger.MessageDocument) {
			message.PostCreateEffectsCompletedAt = &now
		}, expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "self root", mutate: func(message *messenger.MessageDocument) {
			message.ThreadID = message.ID
		}, expected: messenger.ErrPostCreatePipelineInvalid},
		{name: "non UUID root", mutate: func(message *messenger.MessageDocument) {
			message.ThreadID = "legacy-root"
		}, expected: messenger.ErrPostCreatePipelineInvalid},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope := validScope
			message := *validMessage
			if scenario.mutateScope != nil {
				scenario.mutateScope(&scope)
			}
			if scenario.mutate != nil {
				scenario.mutate(&message)
			}
			audience := &phase151UnreadInvalidationAudience{}
			publisher := &phase151UnreadInvalidationPublisher{}

			err := messenger.ExportPhase151PublishUnreadInvalidation(
				context.Background(), scope, &message, audience, publisher,
			)
			require.ErrorIs(t, err, scenario.expected)
			assert.Empty(t, audience.calls)
			assert.Empty(t, publisher.calls)
		})
	}
}

func TestPhase151UnreadInvalidationRequiresContextAudienceAndPublisher(t *testing.T) {
	scope, message := phase151UnreadInvalidationScopeAndMessage()
	audience := &phase151UnreadInvalidationAudience{}
	publisher := &phase151UnreadInvalidationPublisher{}

	assert.Panics(t, func() { messenger.NewPhase151UnreadInvalidationEffect(nil, publisher) })
	assert.Panics(t, func() { messenger.NewPhase151UnreadInvalidationEffect(audience, nil) })

	err := messenger.ExportPhase151PublishUnreadInvalidation(nil, scope, message, audience, publisher)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	cancelledContext, cancel := context.WithCancel(context.Background())
	cancel()
	err = messenger.ExportPhase151PublishUnreadInvalidation(cancelledContext, scope, message, audience, publisher)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	err = messenger.ExportPhase151PublishUnreadInvalidation(context.Background(), scope, message, nil, publisher)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	err = messenger.ExportPhase151PublishUnreadInvalidation(context.Background(), scope, message, audience, nil)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
}
