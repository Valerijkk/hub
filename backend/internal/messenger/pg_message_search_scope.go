package messenger

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

type messageSearchAuthoritySource interface {
	ChannelAuthorityFactSource
	ListMessageSearchCandidateChannelIDs(context.Context, uuid.UUID, uuid.UUID, int) ([]uuid.UUID, error)
}

// PGMessageSearchScopeResolver builds the sole PG-authorized candidate set
// shared by Messenger search and the external global Search message shard.
type PGMessageSearchScopeResolver struct {
	source messageSearchAuthoritySource
}

func NewPGMessageSearchScopeResolver(source messageSearchAuthoritySource) *PGMessageSearchScopeResolver {
	if source == nil {
		panic("messenger PostgreSQL message search authority source is required")
	}
	return &PGMessageSearchScopeResolver{source: source}
}

func (r *PGMessageSearchScopeResolver) ResolveAuthorizedMessageSearchScope(
	ctx context.Context,
	actorID uuid.UUID,
) (AuthorizedMessageSearchScope, error) {
	if r == nil || r.source == nil || actorID == uuid.Nil {
		return AuthorizedMessageSearchScope{}, ErrChannelAccessDenied
	}
	tenantID, err := ExportPhase151ResolveInstanceOrganizationID(ctx, r.source)
	if err != nil {
		return AuthorizedMessageSearchScope{}, err
	}
	channelIDs, err := r.source.ListMessageSearchCandidateChannelIDs(
		ctx, tenantID, actorID, userChannelMembershipCap+1,
	)
	if err != nil {
		return AuthorizedMessageSearchScope{}, fmt.Errorf("list message search candidates: %w", err)
	}
	if len(channelIDs) > userChannelMembershipCap {
		return AuthorizedMessageSearchScope{}, fmt.Errorf("message search scope exceeds %d targets", userChannelMembershipCap)
	}

	scope := AuthorizedMessageSearchScope{TenantID: tenantID, UserID: actorID}
	seen := make(map[uuid.UUID]struct{}, len(channelIDs))
	for _, channelID := range channelIDs {
		if channelID == uuid.Nil {
			return AuthorizedMessageSearchScope{}, ErrChannelAccessDenied
		}
		if _, duplicate := seen[channelID]; duplicate {
			continue
		}
		seen[channelID] = struct{}{}
		facts, factsErr := r.source.ResolveChannelAuthorityFacts(ctx, tenantID, actorID, channelID)
		if factsErr != nil {
			if errors.Is(factsErr, ErrChannelAccessDenied) {
				continue
			}
			return AuthorizedMessageSearchScope{}, fmt.Errorf("resolve message search target: %w", factsErr)
		}
		if facts.TenantID != tenantID || facts.ActorID != actorID || facts.Channel == nil || facts.Channel.ID != channelID {
			return AuthorizedMessageSearchScope{}, ErrChannelAccessDenied
		}
		access, accessErr := ExportPhase151ResolveChannelAccessFromFacts(facts, ChannelActionReadTimeline)
		if accessErr != nil {
			if errors.Is(accessErr, ErrChannelAccessDenied) {
				continue
			}
			return AuthorizedMessageSearchScope{}, accessErr
		}
		kind := "channel"
		if access.Channel.ChannelType == channel.ChannelTypeDm || access.Channel.ChannelType == channel.ChannelTypeGroup {
			kind = "dm"
		}
		scope.Targets = append(scope.Targets, AuthorizedMessageSearchTarget{
			ChannelID: channelID, Kind: kind, IsProjectChannel: access.Channel.ProjectID != nil,
		})
	}
	return scope, nil
}

// ResolveMessageChannelIDs is the narrow adapter consumed by external Search.
func (r *PGMessageSearchScopeResolver) ResolveMessageChannelIDs(
	ctx context.Context,
	actorID uuid.UUID,
) ([]uuid.UUID, error) {
	scope, err := r.ResolveAuthorizedMessageSearchScope(ctx, actorID)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(scope.Targets))
	for _, target := range scope.Targets {
		ids = append(ids, target.ChannelID)
	}
	return ids, nil
}

func (r *EntRepository) ListMessageSearchCandidateChannelIDs(
	ctx context.Context,
	tenantID uuid.UUID,
	actorID uuid.UUID,
	limit int,
) ([]uuid.UUID, error) {
	if r == nil || r.client == nil || tenantID == uuid.Nil || actorID == uuid.Nil || limit <= 0 {
		return nil, ErrChannelAccessDenied
	}
	active, err := r.client.User.Query().Where(
		user.IDEQ(actorID),
		user.StatusEQ(user.StatusActive),
		user.HasOrganizationWith(organization.IDEQ(tenantID)),
	).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, ErrChannelAccessDenied
	}
	return r.client.Channel.Query().Where(channel.DeletedAtIsNil()).Limit(limit).IDs(ctx)
}
