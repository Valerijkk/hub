package messenger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// ChannelAuthorityFactSource is the PostgreSQL-only authority port used by
// the production ChannelAccess resolver. Mongo data cannot implement it.
type ChannelAuthorityFactSource interface {
	InstanceOrganizationIDSource
	ResolveChannelAuthorityFacts(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (ChannelAuthorityFacts, error)
}

// PGChannelAccessResolver binds every channel operation to a fresh singleton
// organization, active actor, live channel membership and project proof.
type PGChannelAccessResolver struct {
	source ChannelAuthorityFactSource
}

func NewPGChannelAccessResolver(source ChannelAuthorityFactSource) *PGChannelAccessResolver {
	if source == nil {
		panic("messenger PostgreSQL channel authority source is required")
	}
	return &PGChannelAccessResolver{source: source}
}

func (r *PGChannelAccessResolver) AuthorizeChannelAccess(
	ctx context.Context,
	actorID uuid.UUID,
	channelID uuid.UUID,
	action ChannelAction,
) (ChannelAccess, error) {
	if r == nil || r.source == nil || actorID == uuid.Nil || channelID == uuid.Nil || !isKnownChannelAction(action) {
		return ChannelAccess{}, ErrChannelAccessDenied
	}
	tenantID, err := ExportPhase151ResolveInstanceOrganizationID(ctx, r.source)
	if err != nil {
		return ChannelAccess{}, err
	}
	facts, err := r.source.ResolveChannelAuthorityFacts(ctx, tenantID, actorID, channelID)
	if err != nil {
		return ChannelAccess{}, fmt.Errorf("resolve channel authority facts: %w", err)
	}
	if facts.TenantID != tenantID || facts.ActorID != actorID || facts.Channel == nil || facts.Channel.ID != channelID {
		return ChannelAccess{}, ErrChannelAccessDenied
	}
	return ExportPhase151ResolveChannelAccessFromFacts(facts, action)
}
