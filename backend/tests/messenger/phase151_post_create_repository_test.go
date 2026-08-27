package messenger_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

var _ messenger.Phase151PostCreatePipelineRepository = (*messenger.Phase151PostCreateRepository)(nil)
var _ messenger.Phase151PostCreateMentionResolver = (*messenger.Phase151EligibleMentionResolver)(nil)
var _ messenger.Phase151PostCreateUnreadInvalidator = (*messenger.Phase151UnreadInvalidationEffect)(nil)

type phase151CompositionMentionCall struct {
	scope   messenger.AuthorizedMessageProducerScope
	request messenger.PostCreatePipelineRequest
}

type phase151CompositionMentions struct {
	resolved []uuid.UUID
	err      error
	calls    []phase151CompositionMentionCall
}

func (r *phase151CompositionMentions) ResolveEligibleMentionUserIDs(
	_ context.Context,
	scope messenger.AuthorizedMessageProducerScope,
	request messenger.PostCreatePipelineRequest,
) ([]uuid.UUID, error) {
	r.calls = append(r.calls, phase151CompositionMentionCall{scope: scope, request: request})
	return r.resolved, r.err
}

type phase151CompositionInvalidationCall struct {
	scope   messenger.AuthorizedMessageProducerScope
	message *messenger.MessageDocument
}

type phase151CompositionInvalidation struct {
	err   error
	calls []phase151CompositionInvalidationCall
}

func (e *phase151CompositionInvalidation) PublishUnreadInvalidation(
	_ context.Context,
	scope messenger.AuthorizedMessageProducerScope,
	message *messenger.MessageDocument,
) error {
	e.calls = append(e.calls, phase151CompositionInvalidationCall{scope: scope, message: message})
	return e.err
}

func phase151CompositionCombinedRepository() *messenger.CombinedRepository {
	return messenger.NewCombinedRepository(&messenger.EntRepository{}, &messenger.MongoRepository{})
}

func TestPhase151PostCreateRepositoryComposesAllPortsAndDelegatesExternalEffects(t *testing.T) {
	resolved := []uuid.UUID{uuid.New(), uuid.New()}
	mentions := &phase151CompositionMentions{resolved: resolved}
	invalidation := &phase151CompositionInvalidation{}
	repository := messenger.NewPhase151PostCreateRepository(
		phase151CompositionCombinedRepository(), mentions, invalidation,
	)
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
		Producer: messenger.TrustedMessageProducerUser,
	}
	request := messenger.PostCreatePipelineRequest{
		Producer: messenger.TrustedMessageProducerUser,
		Message: messenger.MessageDocument{
			ChannelID: scope.ChannelID, SenderID: scope.ActorID, ClientNonce: "compose-0001",
		},
	}
	message := &messenger.MessageDocument{ID: uuid.NewString(), ChannelID: scope.ChannelID, SenderID: scope.ActorID}

	actual, err := repository.ResolveEligibleMentionUserIDs(context.Background(), scope, request)
	require.NoError(t, err)
	assert.Equal(t, resolved, actual)
	require.Equal(t, []phase151CompositionMentionCall{{scope: scope, request: request}}, mentions.calls)
	actual[0] = uuid.New()
	assert.NotEqual(t, actual[0], mentions.resolved[0], "composition does not expose a delegate-owned result slice")

	err = repository.PublishUnreadInvalidation(context.Background(), scope, message)
	require.NoError(t, err)
	assert.Equal(t, []phase151CompositionInvalidationCall{{scope: scope, message: message}}, invalidation.calls)
}

func TestPhase151PostCreateRepositoryPreservesDelegateErrors(t *testing.T) {
	mentionErr := errors.New("mention delegate failed")
	invalidationErr := errors.New("invalidation delegate failed")
	mentions := &phase151CompositionMentions{err: mentionErr}
	invalidation := &phase151CompositionInvalidation{err: invalidationErr}
	repository := messenger.NewPhase151PostCreateRepository(
		phase151CompositionCombinedRepository(), mentions, invalidation,
	)

	_, err := repository.ResolveEligibleMentionUserIDs(
		context.Background(), messenger.AuthorizedMessageProducerScope{}, messenger.PostCreatePipelineRequest{},
	)
	require.ErrorIs(t, err, mentionErr)
	err = repository.PublishUnreadInvalidation(
		context.Background(), messenger.AuthorizedMessageProducerScope{}, &messenger.MessageDocument{},
	)
	require.ErrorIs(t, err, invalidationErr)
}

func TestPhase151PostCreateRepositoryRequiresCompleteComposition(t *testing.T) {
	combined := phase151CompositionCombinedRepository()
	mentions := &phase151CompositionMentions{}
	invalidation := &phase151CompositionInvalidation{}

	assert.Panics(t, func() { messenger.NewPhase151PostCreateRepository(nil, mentions, invalidation) })
	assert.Panics(t, func() {
		messenger.NewPhase151PostCreateRepository(
			messenger.NewCombinedRepository(nil, &messenger.MongoRepository{}), mentions, invalidation,
		)
	})
	assert.Panics(t, func() {
		messenger.NewPhase151PostCreateRepository(
			messenger.NewCombinedRepository(&messenger.EntRepository{}, nil), mentions, invalidation,
		)
	})
	assert.Panics(t, func() { messenger.NewPhase151PostCreateRepository(combined, nil, invalidation) })
	assert.Panics(t, func() { messenger.NewPhase151PostCreateRepository(combined, mentions, nil) })
	var typedNilMentions *phase151CompositionMentions
	var typedNilInvalidation *phase151CompositionInvalidation
	assert.Panics(t, func() { messenger.NewPhase151PostCreateRepository(combined, typedNilMentions, invalidation) })
	assert.Panics(t, func() { messenger.NewPhase151PostCreateRepository(combined, mentions, typedNilInvalidation) })
}

func TestPhase151PostCreateRepositoryNilReceiverFailsExplicitPortsClosed(t *testing.T) {
	var repository *messenger.Phase151PostCreateRepository

	resolved, err := repository.ResolveEligibleMentionUserIDs(
		context.Background(), messenger.AuthorizedMessageProducerScope{}, messenger.PostCreatePipelineRequest{},
	)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	assert.Nil(t, resolved)
	err = repository.PublishUnreadInvalidation(
		context.Background(), messenger.AuthorizedMessageProducerScope{}, &messenger.MessageDocument{},
	)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
}
