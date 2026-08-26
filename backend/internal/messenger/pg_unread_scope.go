package messenger

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/user"
)

var ErrAuthorizedUnreadScopeDenied = errors.New("authorized unread scope denied")

// PGAuthorizedUnreadScopeResolver resolves the only Mongo-eligible unread
// targets from fresh PostgreSQL authority. It never accepts a tenant or target
// from a caller, cache, or Mongo document.
type PGAuthorizedUnreadScopeResolver struct {
	client *ent.Client
}

var _ AuthorizedUnreadScopeResolver = (*PGAuthorizedUnreadScopeResolver)(nil)

// NewPGAuthorizedUnreadScopeResolver creates the production Ent-backed
// resolver. A nil client is retained as a fail-closed resolver so wiring can
// never receive an implicit allow-all fallback.
func NewPGAuthorizedUnreadScopeResolver(client *ent.Client) *PGAuthorizedUnreadScopeResolver {
	return &PGAuthorizedUnreadScopeResolver{client: client}
}

// AuthorizedUnreadScopeFacts contains only fresh PostgreSQL authority facts.
// It is exported so isolated unit tests can prove the closed projection without
// a Mongo dependency; production facts are loaded only by this adapter.
type AuthorizedUnreadScopeFacts struct {
	TenantID              uuid.UUID
	ActorID               uuid.UUID
	ActorOrganizationID   uuid.UUID
	ActorActive           bool
	MembershipTargetFacts []AuthorizedUnreadMembershipTargetFacts
}

// AuthorizedUnreadMembershipTargetFacts is the PG-only membership proof for
// one current channel membership. A target with any missing current fact is
// omitted rather than being widened from channel metadata.
type AuthorizedUnreadMembershipTargetFacts struct {
	ChannelID                  uuid.UUID
	ChannelType                string
	IsCurrentMember            bool
	IsChannelActive            bool
	IsProjectChannel           bool
	HasActiveProjectMembership bool
	IsProjectActive            bool
}

// ExportPhase151ResolveAuthorizedUnreadScopeFromFacts performs the closed
// projection shared by the Ent adapter and isolated unit tests. It intentionally
// does not admit public pre-join channels: unread/follow/state operations are
// member-bound even though public timeline reads are not.
func ExportPhase151ResolveAuthorizedUnreadScopeFromFacts(
	facts AuthorizedUnreadScopeFacts,
) (AuthorizedUnreadScope, error) {
	if facts.TenantID == uuid.Nil ||
		facts.ActorID == uuid.Nil ||
		facts.ActorOrganizationID != facts.TenantID ||
		!facts.ActorActive {
		return AuthorizedUnreadScope{}, ErrAuthorizedUnreadScopeDenied
	}

	scope := AuthorizedUnreadScope{
		TenantID: facts.TenantID,
		UserID:   facts.ActorID,
		Targets:  make([]AuthorizedUnreadTarget, 0, len(facts.MembershipTargetFacts)),
	}
	seen := make(map[uuid.UUID]struct{}, len(facts.MembershipTargetFacts))
	for _, targetFacts := range facts.MembershipTargetFacts {
		if targetFacts.ChannelID == uuid.Nil ||
			!targetFacts.IsCurrentMember ||
			!targetFacts.IsChannelActive {
			continue
		}
		if targetFacts.IsProjectChannel &&
			(!targetFacts.HasActiveProjectMembership || !targetFacts.IsProjectActive) {
			continue
		}

		kind, allowed := authorizedUnreadTargetKind(targetFacts.ChannelType)
		if !allowed {
			continue
		}
		if _, duplicate := seen[targetFacts.ChannelID]; duplicate {
			continue
		}
		seen[targetFacts.ChannelID] = struct{}{}
		scope.Targets = append(scope.Targets, AuthorizedUnreadTarget{
			ChannelID: targetFacts.ChannelID,
			Kind:      kind,
		})
	}
	return scope, nil
}

func (r *PGAuthorizedUnreadScopeResolver) ResolveAuthorizedUnreadScope(
	ctx context.Context,
	actorID uuid.UUID,
) (AuthorizedUnreadScope, error) {
	if r == nil || r.client == nil || ctx == nil || actorID == uuid.Nil {
		return AuthorizedUnreadScope{}, ErrAuthorizedUnreadScopeDenied
	}

	tenantIDs, err := r.client.Organization.Query().
		Limit(2).
		IDs(ctx)
	if err != nil {
		return AuthorizedUnreadScope{}, fmt.Errorf("list instance organizations: %w", err)
	}
	if len(tenantIDs) != 1 || tenantIDs[0] == uuid.Nil {
		return AuthorizedUnreadScope{}, ErrAuthorizedUnreadScopeDenied
	}
	tenantID := tenantIDs[0]

	actorActive, err := r.client.User.Query().
		Where(
			user.IDEQ(actorID),
			user.StatusEQ(user.StatusActive),
			user.HasOrganizationWith(organization.IDEQ(tenantID)),
		).
		Exist(ctx)
	if err != nil {
		return AuthorizedUnreadScope{}, fmt.Errorf("resolve active tenant actor: %w", err)
	}
	if !actorActive {
		return AuthorizedUnreadScope{}, ErrAuthorizedUnreadScopeDenied
	}

	memberships, err := r.client.Channel.Query().
		Where(
			channel.DeletedAtIsNil(),
			channel.HasMembersWith(channelmember.UserID(actorID)),
			channel.Or(
				channel.ProjectIDIsNil(),
				channel.HasProjectWith(
					project.StatusEQ(project.StatusActive),
					project.DeletedAtIsNil(),
					project.HasOrganizationWith(organization.IDEQ(tenantID)),
					project.HasMembersWith(projectmember.HasUserWith(user.IDEQ(actorID))),
				),
			),
		).
		Limit(userChannelMembershipCap + 1).
		All(ctx)
	if err != nil {
		return AuthorizedUnreadScope{}, fmt.Errorf("list active unread memberships: %w", err)
	}
	if len(memberships) > userChannelMembershipCap {
		return AuthorizedUnreadScope{}, ErrAuthorizedUnreadScopeDenied
	}

	targetFacts := make([]AuthorizedUnreadMembershipTargetFacts, 0, len(memberships))
	for _, membership := range memberships {
		if membership == nil {
			continue
		}
		targetFact := AuthorizedUnreadMembershipTargetFacts{
			ChannelID:        membership.ID,
			ChannelType:      string(membership.ChannelType),
			IsCurrentMember:  true,
			IsChannelActive:  membership.DeletedAt == nil,
			IsProjectChannel: membership.ProjectID != nil,
		}
		if targetFact.IsProjectChannel {
			// The Ent predicate above proves both facts from the current row.
			targetFact.HasActiveProjectMembership = true
			targetFact.IsProjectActive = true
		}
		targetFacts = append(targetFacts, targetFact)
	}

	return ExportPhase151ResolveAuthorizedUnreadScopeFromFacts(AuthorizedUnreadScopeFacts{
		TenantID:              tenantID,
		ActorID:               actorID,
		ActorOrganizationID:   tenantID,
		ActorActive:           true,
		MembershipTargetFacts: targetFacts,
	})
}

func authorizedUnreadTargetKind(channelType string) (string, bool) {
	switch channelType {
	case string(channel.ChannelTypePublic), string(channel.ChannelTypePrivate):
		return "channel", true
	case string(channel.ChannelTypeDm), string(channel.ChannelTypeGroup):
		return "dm", true
	default:
		return "", false
	}
}
