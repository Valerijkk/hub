package messenger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

func (s *MessengerService) resolveChannelAction(
	ctx context.Context,
	actorID uuid.UUID,
	channelID uuid.UUID,
	action ChannelAction,
) (ChannelAccess, bool, error) {
	if s.channelAccess == nil {
		return ChannelAccess{}, false, nil
	}
	access, err := s.channelAccess.AuthorizeChannelAccess(ctx, actorID, channelID, action)
	if err != nil {
		return ChannelAccess{}, true, apperrors.NewForbidden("нет доступа к каналу")
	}
	if access.TenantID == uuid.Nil || access.Channel == nil || access.Channel.ID != channelID || !channelActionAllowed(access, action) {
		return ChannelAccess{}, true, apperrors.NewForbidden("нет доступа к каналу")
	}
	return access, true, nil
}

func requireResolvedChannelCapability(access ChannelAccess, action ChannelAction) error {
	allowed := true
	switch action {
	case ChannelActionReadTimeline, ChannelActionReadThread, ChannelActionReadParent:
		allowed = access.CanRead
	case ChannelActionSend:
		allowed = access.CanPost
	case ChannelActionReply:
		allowed = access.CanReply
	}
	if !allowed {
		return apperrors.NewForbidden(fmt.Sprintf("messenger channel action %s denied", action))
	}
	return nil
}
func (s *MessengerService) recipientChannelResponse(
	ctx context.Context,
	actorID uuid.UUID,
	channel *ent.Channel,
) (*ChannelResponse, error) {
	if channel == nil {
		return nil, apperrors.NewForbidden("нет доступа к каналу")
	}
	access, scoped, err := s.resolveChannelAction(ctx, actorID, channel.ID, ChannelActionReadTimeline)
	if err != nil {
		return nil, err
	}
	if !scoped {
		return toChannelResponse(channel), nil
	}
	if access.Channel == nil || access.Channel.ID != channel.ID {
		return nil, apperrors.NewForbidden("нет доступа к каналу")
	}
	return toRecipientChannelResponse(access)
}