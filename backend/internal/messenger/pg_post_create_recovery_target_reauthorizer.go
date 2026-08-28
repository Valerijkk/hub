package messenger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Phase151PendingRecoveryTargetFacts are the PostgreSQL facts required to
// reauthorize a pending recovery target before its Mongo documents are read.
type Phase151PendingRecoveryTargetFacts struct {
	TenantID               uuid.UUID
	ChannelID              uuid.UUID
	ChannelType            string
	IsChannelLive          bool
	IsProjectChannel       bool
	HasActiveTenantProject bool
	HasTenantCreator       bool
}

// Phase151PendingRecoveryTargetFactSource exposes only fresh PostgreSQL facts
// needed to close a pending target to a concrete recovery authority.
type Phase151PendingRecoveryTargetFactSource interface {
	InstanceOrganizationIDSource
	ResolvePendingRecoveryTargetFacts(context.Context, uuid.UUID, uuid.UUID) (Phase151PendingRecoveryTargetFacts, error)
}

// PGPhase151PendingRecoveryTargetReauthorizer reconstructs the PostgreSQL
// authority boundary before a pending recovery target can scan Mongo data.
type PGPhase151PendingRecoveryTargetReauthorizer struct {
	source Phase151PendingRecoveryTargetFactSource
}

func NewPGPhase151PendingRecoveryTargetReauthorizer(
	source Phase151PendingRecoveryTargetFactSource,
) *PGPhase151PendingRecoveryTargetReauthorizer {
	if phase151PostCreateDependencyNil(source) {
		return nil
	}
	return &PGPhase151PendingRecoveryTargetReauthorizer{source: source}
}

func (r *PGPhase151PendingRecoveryTargetReauthorizer) ReauthorizePendingRecoveryTarget(
	ctx context.Context,
	target Phase151PendingRecoveryTarget,
) (Phase151RecoveryTargetAuthority, error) {
	if r == nil || phase151PostCreateDependencyNil(r.source) || phase151PostCreateDependencyNil(ctx) {
		return Phase151RecoveryTargetAuthority{}, ErrPostCreatePipelineInvalid
	}
	if err := ctx.Err(); err != nil {
		return Phase151RecoveryTargetAuthority{}, err
	}
	if target.TenantID == uuid.Nil || target.ChannelID == uuid.Nil ||
		(target.Kind != "channel" && target.Kind != "dm") {
		return Phase151RecoveryTargetAuthority{}, ErrPostCreatePipelineInvalid
	}

	tenantID, err := ExportPhase151ResolveInstanceOrganizationID(ctx, r.source)
	if err != nil {
		return Phase151RecoveryTargetAuthority{}, err
	}
	if tenantID != target.TenantID {
		return Phase151RecoveryTargetAuthority{}, ErrPostCreatePipelineInvalid
	}

	facts, err := r.source.ResolvePendingRecoveryTargetFacts(ctx, tenantID, target.ChannelID)
	if err != nil {
		return Phase151RecoveryTargetAuthority{}, fmt.Errorf("resolve pending recovery target facts: %w", err)
	}
	if facts.TenantID != tenantID || facts.ChannelID != target.ChannelID || !facts.IsChannelLive ||
		!phase151RecoveryTargetKindMatches(facts.ChannelType, target.Kind) ||
		(facts.IsProjectChannel && !facts.HasActiveTenantProject) ||
		(!facts.IsProjectChannel && !facts.HasTenantCreator) {
		return Phase151RecoveryTargetAuthority{}, ErrPostCreatePipelineInvalid
	}

	authority := Phase151RecoveryTargetAuthority{
		Principal: Phase151PostCreateRecoveryServicePrincipal,
		TenantID:  target.TenantID,
		ChannelID: target.ChannelID,
		Kind:      target.Kind,
	}
	if err := ExportPhase151ValidateRecoveryTargetAuthority(
		Phase151PostCreateRecoveryServicePrincipal,
		target,
		authority,
	); err != nil {
		return Phase151RecoveryTargetAuthority{}, err
	}
	return authority, nil
}

func phase151RecoveryTargetKindMatches(channelType, targetKind string) bool {
	switch channelType {
	case "public", "private":
		return targetKind == "channel"
	case "dm", "group":
		return targetKind == "dm"
	default:
		return false
	}
}
