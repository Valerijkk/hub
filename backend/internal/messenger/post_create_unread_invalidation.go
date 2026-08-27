package messenger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Phase151UnreadInvalidationAudience interface {
	ResolveChannelEventAudience(context.Context, uuid.UUID, string) ([]uuid.UUID, error)
}

type Phase151UnreadInvalidationPlan struct {
	ChannelID uuid.UUID
	SenderID  uuid.UUID
	Payload   UnreadChangedEvent
}

func ExportPhase151BuildUnreadInvalidationPlan(
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
) (Phase151UnreadInvalidationPlan, error) {
	if err := ExportPhase151ValidateTrustedMessageProducer(scope, scope.Producer); err != nil {
		return Phase151UnreadInvalidationPlan{}, err
	}
	if message == nil || message.ID == "" || message.ChannelID != scope.ChannelID ||
		message.SenderID != scope.ActorID || message.CreatedAt.IsZero() ||
		message.DeletedAt != nil || message.PostCreateEffectsState != PostCreateEffectsStatePending ||
		message.PostCreateEffectsCompletedAt != nil {
		return Phase151UnreadInvalidationPlan{}, ErrPostCreatePipelineInvalid
	}

	payload := UnreadChangedEvent{
		ChannelID:  scope.ChannelID,
		TargetKind: "channel",
		Category:   "channel",
	}
	if message.ThreadID != "" {
		if message.ThreadID == message.ID {
			return Phase151UnreadInvalidationPlan{}, ErrPostCreatePipelineInvalid
		}
		rootMessageID, err := uuid.Parse(message.ThreadID)
		if err != nil || rootMessageID == uuid.Nil {
			return Phase151UnreadInvalidationPlan{}, ErrPostCreatePipelineInvalid
		}
		payload.TargetKind = "thread"
		payload.Category = "followed_thread"
		payload.RootMessageID = &rootMessageID
	}

	return Phase151UnreadInvalidationPlan{
		ChannelID: scope.ChannelID,
		SenderID:  scope.ActorID,
		Payload:   payload,
	}, nil
}

func ExportPhase151PublishUnreadInvalidation(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
	audience Phase151UnreadInvalidationAudience,
	publisher EventPublisher,
) error {
	if ctx == nil || ctx.Err() != nil || audience == nil || publisher == nil {
		return ErrPostCreatePipelineInvalid
	}
	plan, err := ExportPhase151BuildUnreadInvalidationPlan(scope, message)
	if err != nil {
		return err
	}
	recipients, err := audience.ResolveChannelEventAudience(
		ctx,
		plan.ChannelID,
		UserUnreadChangedEvent,
	)
	if err != nil {
		return fmt.Errorf("resolve unread invalidation audience: %w", err)
	}

	seen := make(map[uuid.UUID]struct{}, len(recipients))
	for _, recipientID := range recipients {
		if recipientID == uuid.Nil || recipientID == plan.SenderID {
			continue
		}
		if _, duplicate := seen[recipientID]; duplicate {
			continue
		}
		seen[recipientID] = struct{}{}
		publisher.PublishJSON(
			fmt.Sprintf(SubjectUserUnreadChanged, recipientID.String()),
			plan.Payload,
		)
	}
	return nil
}

// Phase151UnreadInvalidationEffect binds the canonical post-create port to a
// fresh PostgreSQL-backed audience resolver and the current best-effort event
// publisher. The pipeline invokes it only after a successful complete CAS.
type Phase151UnreadInvalidationEffect struct {
	audience  Phase151UnreadInvalidationAudience
	publisher EventPublisher
}

func NewPhase151UnreadInvalidationEffect(
	audience Phase151UnreadInvalidationAudience,
	publisher EventPublisher,
) *Phase151UnreadInvalidationEffect {
	if audience == nil || publisher == nil {
		panic("post-create unread invalidation dependencies are required")
	}
	return &Phase151UnreadInvalidationEffect{audience: audience, publisher: publisher}
}

func (e *Phase151UnreadInvalidationEffect) PublishUnreadInvalidation(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
) error {
	if e == nil {
		return ErrPostCreatePipelineInvalid
	}
	return ExportPhase151PublishUnreadInvalidation(
		ctx,
		scope,
		message,
		e.audience,
		e.publisher,
	)
}
