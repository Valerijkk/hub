package messenger

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
)

func (r *EntRepository) ListPendingRecoveryTargets(
	ctx context.Context,
	after *Phase151RecoveryTargetCursor,
	limit int,
) (Phase151PendingRecoveryTargetPage, error) {
	if r == nil || r.client == nil || phase151PostCreateDependencyNil(ctx) {
		return Phase151PendingRecoveryTargetPage{}, ErrPostCreatePipelineInvalid
	}
	if err := ctx.Err(); err != nil {
		return Phase151PendingRecoveryTargetPage{}, err
	}
	if limit <= 0 || limit == math.MaxInt {
		return Phase151PendingRecoveryTargetPage{}, ErrPostCreatePipelineInvalid
	}
	if after != nil && (after.TenantID == uuid.Nil ||
		after.ChannelID == uuid.Nil ||
		(after.Kind != "channel" && after.Kind != "dm")) {
		return Phase151PendingRecoveryTargetPage{}, ErrPostCreatePipelineInvalid
	}

	tenantID, err := ExportPhase151ResolveInstanceOrganizationID(ctx, r)
	if err != nil {
		return Phase151PendingRecoveryTargetPage{}, err
	}
	if after != nil && after.TenantID != tenantID {
		return Phase151PendingRecoveryTargetPage{}, ErrPostCreatePipelineInvalid
	}

	query := r.client.Channel.Query().Where(
		channel.DeletedAtIsNil(),
		channel.ChannelTypeIn(
			channel.ChannelTypePublic,
			channel.ChannelTypePrivate,
			channel.ChannelTypeDm,
			channel.ChannelTypeGroup,
		),
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
	)
	if after != nil {
		query = query.Where(channel.IDGT(after.ChannelID))
	}

	channels, err := query.
		Order(ent.Asc(channel.FieldID)).
		Limit(limit + 1).
		All(ctx)
	if err != nil {
		return Phase151PendingRecoveryTargetPage{}, err
	}

	returned := channels
	hasNext := len(channels) > limit
	if hasNext {
		returned = channels[:limit]
	}
	targets := make([]Phase151PendingRecoveryTarget, 0, len(returned))
	for _, ch := range returned {
		if ch == nil {
			return Phase151PendingRecoveryTargetPage{}, ErrPostCreatePipelineInvalid
		}
		kind, ok := phase151PendingRecoveryTargetKind(string(ch.ChannelType))
		if !ok {
			return Phase151PendingRecoveryTargetPage{}, ErrPostCreatePipelineInvalid
		}
		targets = append(targets, Phase151PendingRecoveryTarget{
			TenantID: tenantID, ChannelID: ch.ID, Kind: kind,
		})
	}

	page := Phase151PendingRecoveryTargetPage{Targets: targets}
	if hasNext {
		last := targets[len(targets)-1]
		page.Next = &Phase151RecoveryTargetCursor{
			TenantID: last.TenantID, ChannelID: last.ChannelID, Kind: last.Kind,
		}
	}
	return page, nil
}

func phase151PendingRecoveryTargetKind(channelType string) (string, bool) {
	switch channelType {
	case string(channel.ChannelTypePublic), string(channel.ChannelTypePrivate):
		return "channel", true
	case string(channel.ChannelTypeDm), string(channel.ChannelTypeGroup):
		return "dm", true
	default:
		return "", false
	}
}
