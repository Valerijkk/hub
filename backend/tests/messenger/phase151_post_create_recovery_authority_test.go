package messenger_test

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

func TestPhase151PostCreateRecoveryAuthority(t *testing.T) {
	tenantID := uuid.New()
	channelID := uuid.New()
	target := messenger.Phase151PendingRecoveryTarget{
		TenantID:  tenantID,
		ChannelID: channelID,
		Kind:      "channel_cleanup",
	}
	authority := messenger.Phase151RecoveryTargetAuthority{
		Principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
		TenantID:  tenantID,
		ChannelID: channelID,
		Kind:      target.Kind,
	}

	authorityType := reflect.TypeOf(messenger.Phase151RecoveryTargetAuthority{})
	fieldNames := make([]string, authorityType.NumField())
	for index := range fieldNames {
		fieldNames[index] = authorityType.Field(index).Name
	}
	assert.Equal(t, []string{"Principal", "TenantID", "ChannelID", "Kind"}, fieldNames)

	require.NoError(t, messenger.ExportPhase151ValidateRecoveryTargetAuthority(
		messenger.Phase151PostCreateRecoveryServicePrincipal,
		target,
		authority,
	))

	tests := []struct {
		name      string
		principal messenger.Phase151RecoveryServicePrincipal
		target    messenger.Phase151PendingRecoveryTarget
		authority messenger.Phase151RecoveryTargetAuthority
	}{
		{
			name:      "empty principal",
			principal: "",
			target:    target,
			authority: authority,
		},
		{
			name:      "arbitrary principal",
			principal: messenger.Phase151RecoveryServicePrincipal("messenger.user_actor"),
			target:    target,
			authority: authority,
		},
		{
			name:      "zero target tenant",
			principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
			target: messenger.Phase151PendingRecoveryTarget{
				ChannelID: channelID,
				Kind:      target.Kind,
			},
			authority: authority,
		},
		{
			name:      "zero target channel",
			principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
			target: messenger.Phase151PendingRecoveryTarget{
				TenantID: tenantID,
				Kind:     target.Kind,
			},
			authority: authority,
		},
		{
			name: "empty target kind",
			principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
			target: messenger.Phase151PendingRecoveryTarget{
				TenantID:  tenantID,
				ChannelID: channelID,
			},
			authority: authority,
		},
		{
			name:      "authority principal drift",
			principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
			target:    target,
			authority: messenger.Phase151RecoveryTargetAuthority{
				Principal: messenger.Phase151RecoveryServicePrincipal("messenger.user_actor"),
				TenantID:  tenantID,
				ChannelID: channelID,
				Kind:      target.Kind,
			},
		},
		{
			name:      "authority tenant drift",
			principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
			target:    target,
			authority: messenger.Phase151RecoveryTargetAuthority{
				Principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
				TenantID:  uuid.New(),
				ChannelID: channelID,
				Kind:      target.Kind,
			},
		},
		{
			name:      "authority channel drift",
			principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
			target:    target,
			authority: messenger.Phase151RecoveryTargetAuthority{
				Principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
				TenantID:  tenantID,
				ChannelID: uuid.New(),
				Kind:      target.Kind,
			},
		},
		{
			name:      "authority kind drift",
			principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
			target:    target,
			authority: messenger.Phase151RecoveryTargetAuthority{
				Principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
				TenantID:  tenantID,
				ChannelID: channelID,
				Kind:      "foreign_kind",
			},
		},
		{
			name:      "zero authority tenant",
			principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
			target:    target,
			authority: messenger.Phase151RecoveryTargetAuthority{
				Principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
				ChannelID: channelID,
				Kind:      target.Kind,
			},
		},
		{
			name:      "zero authority channel",
			principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
			target:    target,
			authority: messenger.Phase151RecoveryTargetAuthority{
				Principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
				TenantID:  tenantID,
				Kind:      target.Kind,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := messenger.ExportPhase151ValidateRecoveryTargetAuthority(
				test.principal,
				test.target,
				test.authority,
			)

			require.Error(t, err)
			assert.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
		})
	}
}
