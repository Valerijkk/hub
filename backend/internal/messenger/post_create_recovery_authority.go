package messenger

import "github.com/google/uuid"

type Phase151RecoveryServicePrincipal string

const Phase151PostCreateRecoveryServicePrincipal Phase151RecoveryServicePrincipal = "messenger.post_create_recovery"

type Phase151RecoveryTargetAuthority struct {
	Principal Phase151RecoveryServicePrincipal
	TenantID  uuid.UUID
	ChannelID uuid.UUID
	Kind      string
}

func ExportPhase151ValidateRecoveryTargetAuthority(principal Phase151RecoveryServicePrincipal, target Phase151PendingRecoveryTarget, authority Phase151RecoveryTargetAuthority) error {
	if principal != Phase151PostCreateRecoveryServicePrincipal ||
		validatePhase151PendingRecoveryTarget(target) != nil ||
		authority.Principal != principal ||
		authority.TenantID != target.TenantID ||
		authority.ChannelID != target.ChannelID ||
		authority.Kind != target.Kind {
		return ErrPostCreatePipelineInvalid
	}

	return nil
}
