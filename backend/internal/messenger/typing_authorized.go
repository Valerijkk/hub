package messenger

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ExportPhase151PublishTyping authorizes every ephemeral typing emission with
// the same fresh channel authority used by durable Messenger actions.
func ExportPhase151PublishTyping(
	ctx context.Context,
	actorID uuid.UUID,
	channelID uuid.UUID,
	authority ChannelAccessResolver,
	publisher EventPublisher,
) error {
	if authority == nil {
		panic("messenger typing authority is required")
	}
	if publisher == nil {
		panic("messenger typing publisher is required")
	}
	if actorID == uuid.Nil || channelID == uuid.Nil {
		return errors.New("messenger typing actor and channel are required")
	}
	access, err := authority.AuthorizeChannelAccess(ctx, actorID, channelID, ChannelActionTyping)
	if err != nil {
		return fmt.Errorf("authorize messenger typing: %w", err)
	}
	if access.TenantID == uuid.Nil || access.Channel == nil || access.Channel.ID != channelID || !channelActionAllowed(access, ChannelActionTyping) {
		return ErrChannelAccessDenied
	}
	publisher.PublishJSON(fmt.Sprintf(SubjectTyping, channelID.String()), TypingEvent{
		ChannelID: channelID.String(),
		UserID:    actorID.String(),
	})
	return nil
}
