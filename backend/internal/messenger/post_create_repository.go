package messenger

import (
	"context"
	"reflect"

	"github.com/google/uuid"
)

type Phase151PostCreateMentionResolver interface {
	ResolveEligibleMentionUserIDs(context.Context, AuthorizedMessageProducerScope, PostCreatePipelineRequest) ([]uuid.UUID, error)
}

type Phase151PostCreateUnreadInvalidator interface {
	PublishUnreadInvalidation(context.Context, AuthorizedMessageProducerScope, *MessageDocument) error
}

// Phase151PostCreateRepository composes every concrete canonical create port
// without changing any live writer. Storage methods are promoted from the
// CombinedRepository; the two external effects remain explicit delegates.
type Phase151PostCreateRepository struct {
	*CombinedRepository
	mentions     Phase151PostCreateMentionResolver
	invalidation Phase151PostCreateUnreadInvalidator
}

var _ Phase151PostCreatePipelineRepository = (*Phase151PostCreateRepository)(nil)

func NewPhase151PostCreateRepository(
	combined *CombinedRepository,
	mentions Phase151PostCreateMentionResolver,
	invalidation Phase151PostCreateUnreadInvalidator,
) *Phase151PostCreateRepository {
	if combined == nil || combined.EntRepository == nil || combined.MongoRepository == nil ||
		phase151PostCreateDependencyNil(mentions) || phase151PostCreateDependencyNil(invalidation) {
		panic("complete post-create repository composition is required")
	}
	return &Phase151PostCreateRepository{
		CombinedRepository: combined,
		mentions:           mentions,
		invalidation:       invalidation,
	}
}

func phase151PostCreateDependencyNil(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (r *Phase151PostCreateRepository) ResolveEligibleMentionUserIDs(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	request PostCreatePipelineRequest,
) ([]uuid.UUID, error) {
	if r == nil || r.mentions == nil {
		return nil, ErrPostCreatePipelineInvalid
	}
	resolved, err := r.mentions.ResolveEligibleMentionUserIDs(ctx, scope, request)
	if err != nil {
		return nil, err
	}
	return append([]uuid.UUID(nil), resolved...), nil
}

func (r *Phase151PostCreateRepository) PublishUnreadInvalidation(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
) error {
	if r == nil || r.invalidation == nil {
		return ErrPostCreatePipelineInvalid
	}
	return r.invalidation.PublishUnreadInvalidation(ctx, scope, message)
}
