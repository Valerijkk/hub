package ws

import (
	"context"

	"github.com/google/uuid"
)

type ChannelRoomAuthorizer interface {
	AuthorizeChannelRoom(ctx context.Context, userID, channelID uuid.UUID) (bool, error)
}

func ExportNewPhase151ChannelRoomAuthorizer(resolver ChannelRoomAuthorizer) Authorizer {
	if resolver == nil {
		panic("channel room authorizer resolver is required")
	}

	return NewAuthorizer(AuthorizerConfig{
		IsChannelMember: MembershipChecker(func(ctx context.Context, entityID, userID uuid.UUID) (bool, error) {
			return resolver.AuthorizeChannelRoom(ctx, userID, entityID)
		}),
	})
}
