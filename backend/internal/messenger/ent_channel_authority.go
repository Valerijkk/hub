package messenger

import (
	"context"
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

func (r *EntRepository) ListInstanceOrganizationIDs(ctx context.Context) ([]uuid.UUID, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("messenger Ent client is required")
	}
	return r.client.Organization.Query().Limit(2).IDs(ctx)
}

func (r *EntRepository) ResolveChannelAuthorityFacts(
	ctx context.Context,
	tenantID uuid.UUID,
	actorID uuid.UUID,
	channelID uuid.UUID,
) (ChannelAuthorityFacts, error) {
	if r == nil || r.client == nil || tenantID == uuid.Nil || actorID == uuid.Nil || channelID == uuid.Nil {
		return ChannelAuthorityFacts{}, ErrChannelAccessDenied
	}

	active, err := r.client.User.Query().Where(
		user.IDEQ(actorID),
		user.StatusEQ(user.StatusActive),
		user.HasOrganizationWith(organization.IDEQ(tenantID)),
	).Exist(ctx)
	if err != nil {
		return ChannelAuthorityFacts{}, fmt.Errorf("resolve active channel actor: %w", err)
	}
	if !active {
		return ChannelAuthorityFacts{}, ErrChannelAccessDenied
	}

	ch, err := r.client.Channel.Query().Where(
		channel.IDEQ(channelID),
		channel.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return ChannelAuthorityFacts{}, fmt.Errorf("resolve live channel: %w", err)
	}

	projectEligible := true
	hasActiveProjectMembership := false
	if ch.ProjectID != nil {
		projectEligible, err = r.client.Project.Query().Where(
			project.IDEQ(*ch.ProjectID),
			project.StatusEQ(project.StatusActive),
			project.DeletedAtIsNil(),
			project.HasOrganizationWith(organization.IDEQ(tenantID)),
			project.HasMembersWith(projectmember.HasUserWith(user.IDEQ(actorID))),
		).Exist(ctx)
		hasActiveProjectMembership = projectEligible
	} else {
		projectEligible, err = r.client.User.Query().Where(
			user.IDEQ(ch.CreatedBy),
			user.HasOrganizationWith(organization.IDEQ(tenantID)),
		).Exist(ctx)
	}
	if err != nil {
		return ChannelAuthorityFacts{}, fmt.Errorf("resolve channel tenant and project: %w", err)
	}
	if !projectEligible {
		return ChannelAuthorityFacts{}, ErrChannelAccessDenied
	}

	membership, membershipErr := r.client.ChannelMember.Query().Where(
		channelmember.ChannelIDEQ(channelID),
		channelmember.UserIDEQ(actorID),
	).Only(ctx)
	if membershipErr != nil && !ent.IsNotFound(membershipErr) {
		return ChannelAuthorityFacts{}, fmt.Errorf("resolve channel membership: %w", membershipErr)
	}

	facts := ChannelAuthorityFacts{
		TenantID:                   tenantID,
		ActorID:                    actorID,
		ActorOrganizationID:        tenantID,
		ActorActive:                true,
		Channel:                    ch,
		HasActiveProjectMembership: hasActiveProjectMembership,
	}
	if membership != nil {
		facts.IsMember = true
		facts.Role = string(membership.Role)
	}
	return facts, nil
}
