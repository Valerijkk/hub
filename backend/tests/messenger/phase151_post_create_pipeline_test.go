package messenger_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

var _ messenger.Phase151PostCreatePipelineRepository = (*phase151PostCreateRecorder)(nil)

type phase151PostCreateRecorder struct {
	calls              []string
	created            bool
	canonical          *messenger.MessageDocument
	resolvedMentions   []uuid.UUID
	followCalls        int
	replyCountCalls    int
	completeCalls      int
	invalidationCalls  int
}

func newPhase151PostCreateRecorder() *phase151PostCreateRecorder {
	return &phase151PostCreateRecorder{created: true}
}

func (r *phase151PostCreateRecorder) FindCanonicalCreateResult(
	_ context.Context,
	_ messenger.AuthorizedMessageProducerScope,
	clientNonce string,
) (messenger.CreateMessageResult, bool, error) {
	r.calls = append(r.calls, "idempotency:"+clientNonce)
	if r.created {
		return messenger.CreateMessageResult{}, false, nil
	}
	return messenger.CreateMessageResult{Created: false, Message: r.canonical}, true, nil
}

func (r *phase151PostCreateRecorder) ResolveEligibleMentionUserIDs(
	_ context.Context,
	_ messenger.AuthorizedMessageProducerScope,
	_ messenger.PostCreatePipelineRequest,
) ([]uuid.UUID, error) {
	r.calls = append(r.calls, "resolve-mentions")
	return append([]uuid.UUID(nil), r.resolvedMentions...), nil
}

func (r *phase151PostCreateRecorder) PersistPendingCanonicalMessage(
	_ context.Context,
	_ messenger.AuthorizedMessageProducerScope,
	document *messenger.MessageDocument,
) (messenger.CreateMessageResult, error) {
	r.calls = append(r.calls, "persist-pending")
	copyDocument := *document
	copyDocument.ResolvedMentionUserIDs = append([]uuid.UUID(nil), document.ResolvedMentionUserIDs...)
	r.canonical = &copyDocument
	return messenger.CreateMessageResult{Created: true, Message: r.canonical, PostCreateEffectsState: messenger.PostCreateEffectsStatePending}, nil
}

func (r *phase151PostCreateRecorder) EnsureAutoThreadFollows(
	_ context.Context,
	_ messenger.AuthorizedMessageProducerScope,
	_ *messenger.MessageDocument,
) error {
	r.calls = append(r.calls, "auto-follow")
	r.followCalls++
	return nil
}

func (r *phase151PostCreateRecorder) ReconcileReplyCount(
	_ context.Context,
	_ messenger.AuthorizedMessageProducerScope,
	_ *messenger.MessageDocument,
) error {
	r.calls = append(r.calls, "reply-count")
	r.replyCountCalls++
	return nil
}

func (r *phase151PostCreateRecorder) CompletePostCreateEffects(
	_ context.Context,
	_ messenger.AuthorizedMessageProducerScope,
	_ string,
) (bool, error) {
	r.calls = append(r.calls, "complete-cas")
	r.completeCalls++
	return true, nil
}

func (r *phase151PostCreateRecorder) PublishUnreadInvalidation(
	_ context.Context,
	_ messenger.AuthorizedMessageProducerScope,
	_ *messenger.MessageDocument,
) error {
	r.calls = append(r.calls, "recipient-invalidation")
	r.invalidationCalls++
	return nil
}

func TestPhase151AllCanonicalProducersUsePendingEffectsPipeline(t *testing.T) {
	for _, producer := range []messenger.TrustedMessageProducer{
		messenger.TrustedMessageProducerUser,
		messenger.TrustedMessageProducerForward,
		messenger.TrustedMessageProducerSystemBot,
		messenger.TrustedMessageProducerPoll,
		messenger.TrustedMessageProducerPromotionBacklink,
	} {
		t.Run(string(producer), func(t *testing.T) {
			tenantID := uuid.New()
			actorID := uuid.New()
			channelID := uuid.New()
			mentionedID := uuid.New()
			recorder := newPhase151PostCreateRecorder()
			recorder.resolvedMentions = []uuid.UUID{mentionedID}
			scope := messenger.AuthorizedMessageProducerScope{
				TenantID: tenantID, ActorID: actorID, ChannelID: channelID, Producer: producer,
			}
			if producer == messenger.TrustedMessageProducerSystemBot {
				scope.BotKind = messenger.BotKindSystem
				scope.IntegrationID = uuid.New()
			}
			request := messenger.PostCreatePipelineRequest{
				Producer: producer,
				Message: messenger.MessageDocument{
					ChannelID: channelID, SenderID: actorID,
					ContentText: "phase151 canonical producer", ClientNonce: "phase151-" + string(producer),
				},
			}

			result, err := messenger.ExportPhase151ExecutePostCreatePipeline(context.Background(), scope, request, recorder)

			require.NoError(t, err)
			assert.True(t, result.Created)
			require.NotNil(t, recorder.canonical)
			assert.Equal(t, messenger.PostCreateEffectsStatePending, recorder.canonical.PostCreateEffectsState,
				"the canonical document is pending before any effect runs")
			assert.Equal(t, []uuid.UUID{mentionedID}, recorder.canonical.ResolvedMentionUserIDs)
			recorder.resolvedMentions[0] = uuid.New()
			assert.Equal(t, mentionedID, recorder.canonical.ResolvedMentionUserIDs[0], "resolved mentions are persisted immutably")
			assert.Equal(t, []string{
				"idempotency:" + request.Message.ClientNonce,
				"resolve-mentions",
				"persist-pending",
				"auto-follow",
				"complete-cas",
				"recipient-invalidation",
			}, recorder.calls)
			assert.Equal(t, 1, recorder.followCalls)
			assert.Zero(t, recorder.replyCountCalls)
			assert.Equal(t, 1, recorder.completeCalls)
			assert.Equal(t, 1, recorder.invalidationCalls)
		})
	}
}

func TestPhase151DirectReplyRunsReplyEffectsAndKeepsOneCanonicalDocument(t *testing.T) {
	channelID := uuid.New()
	root := &messenger.MessageDocument{
		ID: "phase151-live-root", ChannelID: channelID, SenderID: uuid.New(), CreatedAt: time.Now().UTC(),
	}
	recorder := newPhase151PostCreateRecorder()
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: channelID,
		Producer: messenger.TrustedMessageProducerUser,
	}
	request := messenger.PostCreatePipelineRequest{
		Producer: messenger.TrustedMessageProducerUser,
		Message: messenger.MessageDocument{
			ChannelID: channelID, SenderID: scope.ActorID, ContentText: "direct reply",
			ClientNonce: "phase151-direct-reply", ThreadID: root.ID, AlsoSendToChannel: true,
		},
		Parent: root,
	}

	result, err := messenger.ExportPhase151ExecutePostCreatePipeline(context.Background(), scope, request, recorder)

	require.NoError(t, err)
	assert.True(t, result.Created)
	require.NotNil(t, recorder.canonical)
	assert.Equal(t, root.ID, recorder.canonical.ThreadID)
	assert.True(t, recorder.canonical.AlsoSendToChannel)
	assert.Equal(t, 1, recorder.replyCountCalls)
	assert.Equal(t, 1, recorder.followCalls)
	assert.Equal(t, 1, recorder.completeCalls)
	assert.Equal(t, 1, recorder.invalidationCalls)
	assert.Equal(t, []string{
		"idempotency:" + request.Message.ClientNonce,
		"resolve-mentions",
		"persist-pending",
		"auto-follow",
		"reply-count",
		"complete-cas",
		"recipient-invalidation",
	}, recorder.calls)
}

func TestPhase151ReplyValidationRejectsNestedCrossChannelAndDeletedRootBeforeInsert(t *testing.T) {
	channelID := uuid.New()
	deletedAt := time.Now().UTC()
	for _, tt := range []struct {
		name   string
		parent *messenger.MessageDocument
	}{
		{name: "nested reply", parent: &messenger.MessageDocument{ID: "nested", ChannelID: channelID, ThreadID: "root"}},
		{name: "cross channel", parent: &messenger.MessageDocument{ID: "foreign", ChannelID: uuid.New()}},
		{name: "deleted root", parent: &messenger.MessageDocument{ID: "deleted", ChannelID: channelID, DeletedAt: &deletedAt}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := newPhase151PostCreateRecorder()
			scope := messenger.AuthorizedMessageProducerScope{
				TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: channelID,
				Producer: messenger.TrustedMessageProducerUser,
			}
			request := messenger.PostCreatePipelineRequest{
				Producer: messenger.TrustedMessageProducerUser,
				Message: messenger.MessageDocument{
					ChannelID: channelID, SenderID: scope.ActorID, ContentText: "must fail",
					ClientNonce: "phase151-invalid-" + tt.name, ThreadID: tt.parent.ID,
				},
				Parent: tt.parent,
			}

			result, err := messenger.ExportPhase151ExecutePostCreatePipeline(context.Background(), scope, request, recorder)

			require.Error(t, err)
			assert.False(t, result.Created)
			assert.Nil(t, recorder.canonical)
			assert.NotContains(t, recorder.calls, "persist-pending")
			assert.Zero(t, recorder.completeCalls)
			assert.Zero(t, recorder.invalidationCalls)
		})
	}
}

func TestPhase151InvalidFullPayloadStopsBeforeRepository(t *testing.T) {
	channelID, actorID := uuid.New(), uuid.New()
	cyclicContent := map[string]interface{}{}
	cyclicContent["self"] = cyclicContent
	recorder := newPhase151PostCreateRecorder()
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: actorID, ChannelID: channelID,
		Producer: messenger.TrustedMessageProducerUser,
	}
	request := messenger.PostCreatePipelineRequest{
		Producer: messenger.TrustedMessageProducerUser,
		Message: messenger.MessageDocument{
			ChannelID: channelID, SenderID: actorID, ClientNonce: "phase151-invalid-json",
			ContentJSON: cyclicContent,
		},
	}

	result, err := messenger.ExportPhase151ExecutePostCreatePipeline(context.Background(), scope, request, recorder)

	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	assert.False(t, result.Created)
	assert.Empty(t, recorder.calls, "invalid prepared payload must stop before nonce lookup, mention resolution, or persistence")
}
func TestPhase151CreatedFalseRetryRepeatsNoPostCreateEffect(t *testing.T) {
	channelID := uuid.New()
	canonical := &messenger.MessageDocument{
		ID: "phase151-canonical-retry", ChannelID: channelID, SenderID: uuid.New(),
		PostCreateEffectsState: messenger.PostCreateEffectsStateComplete,
	}
	recorder := newPhase151PostCreateRecorder()
	recorder.created = false
	recorder.canonical = canonical
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: canonical.SenderID, ChannelID: channelID,
		Producer: messenger.TrustedMessageProducerUser,
	}
	request := messenger.PostCreatePipelineRequest{
		Producer: messenger.TrustedMessageProducerUser,
		Message: messenger.MessageDocument{
			ChannelID: channelID, SenderID: canonical.SenderID, ContentText: "retry",
			ClientNonce: "phase151-existing-canonical",
		},
	}

	result, err := messenger.ExportPhase151ExecutePostCreatePipeline(context.Background(), scope, request, recorder)

	require.NoError(t, err)
	assert.False(t, result.Created)
	assert.Same(t, canonical, result.Message)
	assert.Equal(t, []string{"idempotency:" + request.Message.ClientNonce}, recorder.calls)
	assert.Zero(t, recorder.followCalls)
	assert.Zero(t, recorder.replyCountCalls)
	assert.Zero(t, recorder.completeCalls)
	assert.Zero(t, recorder.invalidationCalls)
}

func TestPhase151TrustedProducerScopeRejectsGenericAndBotWidening(t *testing.T) {
	base := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
	}

	require.Error(t, messenger.ExportPhase151ValidateTrustedMessageProducer(base, messenger.TrustedMessageProducer("generic")))

	botScope := base
	botScope.Producer = messenger.TrustedMessageProducerSystemBot
	botScope.BotKind = messenger.BotKindSystem
	botScope.IntegrationID = uuid.New()
	require.NoError(t, messenger.ExportPhase151ValidateTrustedMessageProducer(botScope, messenger.TrustedMessageProducerSystemBot))

	mismatchedChannel := botScope
	mismatchedChannel.ChannelID = uuid.New()
	require.Error(t, messenger.ExportPhase151ValidateAuthorizedBotProducer(botScope, mismatchedChannel))

	mismatchedIntegration := botScope
	mismatchedIntegration.IntegrationID = uuid.New()
	require.Error(t, messenger.ExportPhase151ValidateAuthorizedBotProducer(botScope, mismatchedIntegration))

	disabledBot := botScope
	disabledBot.BotKind = messenger.BotKind("")
	require.Error(t, messenger.ExportPhase151ValidateTrustedMessageProducer(disabledBot, messenger.TrustedMessageProducerSystemBot))
}

var errPhase151PostCreateRecorder = errors.New("phase151 post-create recorder error")
type phase151CompleteCreateRecorder struct {
	*phase151PostCreateRecorder
}

func (r *phase151CompleteCreateRecorder) PersistPendingCanonicalMessage(
	_ context.Context,
	_ messenger.AuthorizedMessageProducerScope,
	document *messenger.MessageDocument,
) (messenger.CreateMessageResult, error) {
	r.calls = append(r.calls, "persist-pending")
	copyDocument := *document
	copyDocument.PostCreateEffectsState = messenger.PostCreateEffectsStateComplete
	r.canonical = &copyDocument
	return messenger.CreateMessageResult{
		Created: true, Message: r.canonical, PostCreateEffectsState: messenger.PostCreateEffectsStateComplete,
	}, nil
}

func TestPhase151CreatedCompleteResultRepeatsNoPostCreateEffect(t *testing.T) {
	channelID, actorID := uuid.New(), uuid.New()
	recorder := &phase151CompleteCreateRecorder{phase151PostCreateRecorder: newPhase151PostCreateRecorder()}
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: actorID, ChannelID: channelID,
		Producer: messenger.TrustedMessageProducerUser,
	}
	request := messenger.PostCreatePipelineRequest{
		Producer: messenger.TrustedMessageProducerUser,
		Message: messenger.MessageDocument{
			ChannelID: channelID, SenderID: actorID,
			ContentText: "complete retry", ClientNonce: "phase151-complete-create-result",
		},
	}

	result, err := messenger.ExportPhase151ExecutePostCreatePipeline(context.Background(), scope, request, recorder)

	require.NoError(t, err)
	assert.True(t, result.Created)
	assert.Equal(t, []string{
		"idempotency:" + request.Message.ClientNonce, "resolve-mentions", "persist-pending",
	}, recorder.calls)
	assert.Zero(t, recorder.followCalls)
	assert.Zero(t, recorder.replyCountCalls)
	assert.Zero(t, recorder.completeCalls)
	assert.Zero(t, recorder.invalidationCalls)
}