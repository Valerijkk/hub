package audit_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/audit"
)

type phase151UserRevocationRecorder struct {
	effects    []string
	roomErrors []error
}

func (r *phase151UserRevocationRecorder) RevokeUserMessengerAuthority(
	_ context.Context, scope audit.MessengerUserAccessScope,
) error {
	r.effects = append(r.effects, phase151UserEffect("pg", scope))
	return nil
}

func (r *phase151UserRevocationRecorder) CleanupUserMessengerState(
	_ context.Context, scope audit.MessengerUserAccessScope,
) error {
	r.effects = append(r.effects, phase151UserEffect("mongo", scope))
	return nil
}

func (r *phase151UserRevocationRecorder) InvalidateUserMessengerRooms(
	_ context.Context, scope audit.MessengerUserAccessScope,
) error {
	r.effects = append(r.effects, phase151UserEffect("room", scope))
	if len(r.roomErrors) == 0 {
		return nil
	}
	err := r.roomErrors[0]
	r.roomErrors = r.roomErrors[1:]
	return err
}

func TestPhase151UserAccessLoss_UsesPGCleanupRoomSequence(t *testing.T) {
	for _, reason := range []audit.MessengerUserAccessLossReason{
		audit.MessengerUserAccountDisabled,
		audit.MessengerUserAccountDeleted,
		audit.MessengerUserOrganizationRemoved,
	} {
		t.Run(string(reason), func(t *testing.T) {
			tenantID, organizationID, userID := uuid.New(), uuid.New(), uuid.New()
			scope := audit.MessengerUserAccessScope{TenantID: tenantID, OrganizationID: organizationID, UserID: userID, LifecycleGeneration: 7}
			recorder := &phase151UserRevocationRecorder{}
			revoker := audit.NewMessengerUserAccessRevoker(audit.MessengerUserAccessRevokerDependencies{
				Authority: recorder, DurableState: recorder, UserRooms: recorder,
			})

			err := revoker.RevokeUserAccess(context.Background(), audit.MessengerUserAccessLoss{
				TenantID: tenantID, OrganizationID: organizationID, UserID: userID, Reason: reason, LifecycleGeneration: 7,
			})

			require.NoError(t, err)
			assert.Equal(t, []string{
				phase151UserEffect("pg", scope), phase151UserEffect("mongo", scope), phase151UserEffect("room", scope),
			}, recorder.effects)
			for _, effect := range recorder.effects {
				assert.Contains(t, effect, tenantID.String(), "no cross-tenant effect")
			}
		})
	}
}

func TestPhase151UserAccessLoss_RoomFailureLeavesRetryableRevokedResidual(t *testing.T) {
	tenantID, organizationID, userID := uuid.New(), uuid.New(), uuid.New()
	scope := audit.MessengerUserAccessScope{TenantID: tenantID, OrganizationID: organizationID, UserID: userID, LifecycleGeneration: 9}
	recorder := &phase151UserRevocationRecorder{roomErrors: []error{errors.New("room invalidation unavailable")}}
	revoker := audit.NewMessengerUserAccessRevoker(audit.MessengerUserAccessRevokerDependencies{
		Authority: recorder, DurableState: recorder, UserRooms: recorder,
	})
	input := audit.MessengerUserAccessLoss{
		TenantID: tenantID, OrganizationID: organizationID, UserID: userID, Reason: audit.MessengerUserAccountDisabled, LifecycleGeneration: 9,
	}

	err := revoker.RevokeUserAccess(context.Background(), input)
	var residual *audit.MessengerUserAccessRevocationError
	require.ErrorAs(t, err, &residual)
	assert.Equal(t, audit.MessengerUserRoomInvalidationStage, residual.Stage)
	assert.True(t, residual.Retryable)
	assert.Equal(t, []string{
		phase151UserEffect("pg", scope), phase151UserEffect("mongo", scope), phase151UserEffect("room", scope),
	}, recorder.effects)

	require.NoError(t, revoker.RevokeUserAccess(context.Background(), input))
	assert.Equal(t, []string{
		phase151UserEffect("pg", scope), phase151UserEffect("mongo", scope), phase151UserEffect("room", scope),
		phase151UserEffect("pg", scope), phase151UserEffect("mongo", scope), phase151UserEffect("room", scope),
	}, recorder.effects)
}

func phase151UserEffect(stage string, scope audit.MessengerUserAccessScope) string {
	return fmt.Sprintf("%s:%s:%s:%s:%d", stage, scope.TenantID, scope.OrganizationID, scope.UserID, scope.LifecycleGeneration)
}
