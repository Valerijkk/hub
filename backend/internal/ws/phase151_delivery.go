package ws

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// ChannelEventAudienceResolver resolves the users permitted to receive a channel event.
type ChannelEventAudienceResolver interface {
	ResolveChannelEventAudience(ctx context.Context, channelID uuid.UUID, event string) ([]uuid.UUID, error)
}

// UserChannelEventAuthorizer determines whether a user may emit a channel event.
type UserChannelEventAuthorizer interface {
	AuthorizeUserChannelEvent(ctx context.Context, userID, channelID uuid.UUID, event string) (bool, error)
}

// ExportPhase151ResolveChannelDeliveryRecipients returns authorized subscribers in subscription order.
func ExportPhase151ResolveChannelDeliveryRecipients(
	ctx context.Context,
	channelID uuid.UUID,
	event string,
	subscribers []uuid.UUID,
	resolver ChannelEventAudienceResolver,
) []uuid.UUID {
	if ctx == nil || ctx.Err() != nil || channelID == uuid.Nil || strings.TrimSpace(event) == "" || resolver == nil {
		return []uuid.UUID{}
	}

	audience, err := resolver.ResolveChannelEventAudience(ctx, channelID, event)
	if err != nil {
		return []uuid.UUID{}
	}

	authorized := make(map[uuid.UUID]struct{}, len(audience))
	for _, userID := range audience {
		if userID != uuid.Nil {
			authorized[userID] = struct{}{}
		}
	}

	recipients := make([]uuid.UUID, 0, len(subscribers))
	seen := make(map[uuid.UUID]struct{}, len(subscribers))
	for _, subscriberID := range subscribers {
		if subscriberID == uuid.Nil {
			continue
		}
		if _, alreadyIncluded := seen[subscriberID]; alreadyIncluded {
			continue
		}
		if _, isAuthorized := authorized[subscriberID]; !isAuthorized {
			continue
		}

		seen[subscriberID] = struct{}{}
		recipients = append(recipients, subscriberID)
	}

	return recipients
}

// ExportPhase151ResolveTypingDeliveryRecipients returns recipients for an authorized typing event.
func ExportPhase151ResolveTypingDeliveryRecipients(
	ctx context.Context,
	channelID, senderID uuid.UUID,
	subscribers []uuid.UUID,
	senderAuth UserChannelEventAuthorizer,
	audience ChannelEventAudienceResolver,
) []uuid.UUID {
	if ctx == nil || ctx.Err() != nil || channelID == uuid.Nil || senderID == uuid.Nil || senderAuth == nil || audience == nil {
		return []uuid.UUID{}
	}

	allowed, err := senderAuth.AuthorizeUserChannelEvent(ctx, senderID, channelID, "typing")
	if err != nil || !allowed {
		return []uuid.UUID{}
	}

	return ExportPhase151ResolveChannelDeliveryRecipients(ctx, channelID, "typing", subscribers, audience)
}
