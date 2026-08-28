package messenger

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
)

var _ Phase151PendingRecoveryTargetFactSource = (*EntRepository)(nil)

func (r *EntRepository) ResolvePendingRecoveryTargetFacts(
	ctx context.Context,
	tenantID uuid.UUID,
	channelID uuid.UUID,
) (Phase151PendingRecoveryTargetFacts, error) {
	if r == nil || r.client == nil || phase151PostCreateDependencyNil(ctx) {
		return Phase151PendingRecoveryTargetFacts{}, ErrPostCreatePipelineInvalid
	}
	if err := ctx.Err(); err != nil {
		return Phase151PendingRecoveryTargetFacts{}, err
	}
	if tenantID == uuid.Nil || channelID == uuid.Nil {
		return Phase151PendingRecoveryTargetFacts{}, ErrPostCreatePipelineInvalid
	}

	ch, err := r.client.Channel.Query().Where(
		channel.IDEQ(channelID),
		channel.DeletedAtIsNil(),
		channel.Or(
			channel.And(
				channel.ProjectIDNotNil(),
				channel.HasProjectWith(
					project.StatusEQ(project.StatusActive),
					project.DeletedAtIsNil(),
					project.HasOrganizationWith(organization.IDEQ(tenantID)),
				),
			),
			channel.And(
				channel.ProjectIDIsNil(),
				channel.HasCreatorWith(
					user.HasOrganizationWith(organization.IDEQ(tenantID)),
				),
			),
		),
	).Only(ctx)
	if err != nil {
		return Phase151PendingRecoveryTargetFacts{}, err
	}
	if ch == nil {
		return Phase151PendingRecoveryTargetFacts{}, ErrPostCreatePipelineInvalid
	}

	return Phase151PendingRecoveryTargetFacts{
		TenantID:               tenantID,
		ChannelID:              ch.ID,
		ChannelType:            string(ch.ChannelType),
		IsChannelLive:          ch.DeletedAt == nil,
		IsProjectChannel:       ch.ProjectID != nil,
		HasActiveTenantProject: ch.ProjectID != nil,
		HasTenantCreator:       ch.ProjectID == nil,
	}, nil
}
