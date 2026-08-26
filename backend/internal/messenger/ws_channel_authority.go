package messenger

import (
	"context"

	"github.com/google/uuid"
)

// WSChannelMemberSource supplies only the current PostgreSQL channel-member
// candidates. Membership alone never grants room access or event delivery.
type WSChannelMemberSource interface {
	GetChannelMemberIDs(context.Context, uuid.UUID) ([]uuid.UUID, error)
}

// WSChannelAuthority applies the fresh ChannelAccess boundary to every
// WebSocket room subscription, sender check, and delivery candidate.
type WSChannelAuthority struct {
	resolver ChannelAccessResolver
	members  WSChannelMemberSource
}

func NewWSChannelAuthority(
	resolver ChannelAccessResolver,
	members WSChannelMemberSource,
) *WSChannelAuthority {
	if resolver == nil || members == nil {
		panic("messenger WebSocket channel authority dependencies are required")
	}
	return &WSChannelAuthority{resolver: resolver, members: members}
}

func (a *WSChannelAuthority) AuthorizeChannelRoom(
	ctx context.Context,
	userID, channelID uuid.UUID,
) (bool, error) {
	return a.authorizeReceive(ctx, userID, channelID)
}

func (a *WSChannelAuthority) AuthorizeUserChannelEvent(
	ctx context.Context,
	userID, channelID uuid.UUID,
	_ string,
) (bool, error) {
	return a.authorizeReceive(ctx, userID, channelID)
}

func (a *WSChannelAuthority) ResolveChannelEventAudience(
	ctx context.Context,
	channelID uuid.UUID,
	_ string,
) ([]uuid.UUID, error) {
	if a == nil || a.resolver == nil || a.members == nil || channelID == uuid.Nil {
		return nil, ErrChannelAccessDenied
	}
	candidates, err := a.members.GetChannelMemberIDs(ctx, channelID)
	if err != nil {
		return nil, err
	}
	recipients := make([]uuid.UUID, 0, len(candidates))
	seen := make(map[uuid.UUID]struct{}, len(candidates))
	for _, candidateID := range candidates {
		if candidateID == uuid.Nil {
			continue
		}
		if _, duplicate := seen[candidateID]; duplicate {
			continue
		}
		seen[candidateID] = struct{}{}
		allowed, authErr := a.authorizeReceive(ctx, candidateID, channelID)
		if authErr != nil || !allowed {
			continue
		}
		recipients = append(recipients, candidateID)
	}
	return recipients, nil
}

func (a *WSChannelAuthority) authorizeReceive(
	ctx context.Context,
	userID, channelID uuid.UUID,
) (bool, error) {
	if a == nil || a.resolver == nil || userID == uuid.Nil || channelID == uuid.Nil {
		return false, ErrChannelAccessDenied
	}
	access, err := a.resolver.AuthorizeChannelAccess(
		ctx,
		userID,
		channelID,
		ChannelActionReceiveEvent,
	)
	if err != nil {
		return false, err
	}
	if access.TenantID == uuid.Nil ||
		access.Channel == nil ||
		access.Channel.ID != channelID ||
		!access.CanReceiveEvent {
		return false, ErrChannelAccessDenied
	}
	return true, nil
}
