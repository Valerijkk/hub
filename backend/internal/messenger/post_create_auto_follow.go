package messenger

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Phase151AutoThreadFollowPlan struct {
	UserIDs       []uuid.UUID
	ChannelID     uuid.UUID
	RootMessageID string
	Transition    ThreadFollowTransition
}

type Phase151AutoThreadFollowTransitionWriter interface {
	ApplyAuthorizedThreadFollowTransition(
		context.Context,
		AuthorizedUnreadScope,
		uuid.UUID,
		string,
		ThreadFollowTransition,
	) (bool, error)
}

func ExportPhase151BuildAutoThreadFollowPlan(
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
) (Phase151AutoThreadFollowPlan, error) {
	if err := ExportPhase151ValidateTrustedMessageProducer(scope, scope.Producer); err != nil {
		return Phase151AutoThreadFollowPlan{}, err
	}
	if message == nil || message.ID == "" || message.ChannelID != scope.ChannelID ||
		message.SenderID != scope.ActorID || message.CreatedAt.IsZero() ||
		message.DeletedAt != nil || message.PostCreateEffectsState != PostCreateEffectsStatePending ||
		message.PostCreateEffectsCompletedAt != nil {
		return Phase151AutoThreadFollowPlan{}, ErrPostCreatePipelineInvalid
	}

	rootMessageID := message.ID
	if message.ThreadID != "" {
		if message.ThreadID == message.ID {
			return Phase151AutoThreadFollowPlan{}, ErrPostCreatePipelineInvalid
		}
		rootMessageID = message.ThreadID
	}

	userIDs := make([]uuid.UUID, 0, 1+len(message.ResolvedMentionUserIDs))
	seen := make(map[uuid.UUID]struct{}, 1+len(message.ResolvedMentionUserIDs))
	userIDs = append(userIDs, scope.ActorID)
	seen[scope.ActorID] = struct{}{}
	for _, userID := range message.ResolvedMentionUserIDs {
		if userID == uuid.Nil {
			return Phase151AutoThreadFollowPlan{}, ErrPostCreatePipelineInvalid
		}
		if _, duplicate := seen[userID]; duplicate {
			continue
		}
		seen[userID] = struct{}{}
		userIDs = append(userIDs, userID)
	}

	return Phase151AutoThreadFollowPlan{
		UserIDs:       userIDs,
		ChannelID:     scope.ChannelID,
		RootMessageID: rootMessageID,
		Transition: ThreadFollowTransition{
			Mode: ThreadFollowModeAuto,
			Position: MessagePosition{
				CreatedAt: message.CreatedAt.UTC(),
				MessageID: message.ID,
			},
		},
	}, nil
}

func ExportPhase151EnsureAutoThreadFollows(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
	scopeResolver AuthorizedUnreadScopeResolver,
	writer Phase151AutoThreadFollowTransitionWriter,
) error {
	if ctx == nil || scopeResolver == nil || writer == nil {
		return ErrPostCreatePipelineInvalid
	}
	plan, err := ExportPhase151BuildAutoThreadFollowPlan(scope, message)
	if err != nil {
		return err
	}

	for _, userID := range plan.UserIDs {
		authorizedScope, resolveErr := scopeResolver.ResolveAuthorizedUnreadScope(ctx, userID)
		if errors.Is(resolveErr, ErrAuthorizedUnreadScopeDenied) {
			continue
		}
		if resolveErr != nil {
			return fmt.Errorf("resolve automatic thread-follow authority: %w", resolveErr)
		}
		if authorizedScope.TenantID != scope.TenantID || authorizedScope.UserID != userID {
			return ErrPostCreatePipelineInvalid
		}
		if !phase151MongoUnreadScopeAllowsTarget(authorizedScope, plan.ChannelID, "channel") {
			continue
		}
		if _, writeErr := writer.ApplyAuthorizedThreadFollowTransition(
			ctx,
			authorizedScope,
			plan.ChannelID,
			plan.RootMessageID,
			plan.Transition,
		); writeErr != nil {
			return fmt.Errorf("persist automatic thread follow: %w", writeErr)
		}
	}
	return nil
}

// EnsureAutoThreadFollows binds the canonical post-create effect to fresh Ent
// authority and the existing insert-only Mongo transition. The combined
// repository remains unbound to live message writers until all effect ports
// and recovery ownership are complete.
func (r *CombinedRepository) EnsureAutoThreadFollows(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
) error {
	if r == nil || r.EntRepository == nil || r.EntRepository.client == nil || r.MongoRepository == nil {
		return ErrPostCreatePipelineInvalid
	}
	return ExportPhase151EnsureAutoThreadFollows(
		ctx,
		scope,
		message,
		NewPGAuthorizedUnreadScopeResolver(r.EntRepository.client),
		r.MongoRepository,
	)
}
