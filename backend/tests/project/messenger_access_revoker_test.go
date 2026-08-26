package project_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/project"
)

type phase151ProjectRevocationRecorder struct {
	members       []uuid.UUID
	effects       []string
	cleanupErrors []error
}

func (r *phase151ProjectRevocationRecorder) ListProjectMessengerUserIDs(
	_ context.Context, tenantID, projectID uuid.UUID,
) ([]uuid.UUID, error) {
	r.effects = append(r.effects, "list:"+tenantID.String()+":"+projectID.String())
	return append([]uuid.UUID(nil), r.members...), nil
}

func (r *phase151ProjectRevocationRecorder) RevokeProjectMessengerAuthority(
	_ context.Context, scope project.MessengerProjectAccessScope,
) error {
	r.effects = append(r.effects, phase151ProjectEffect("pg", scope))
	return nil
}

func (r *phase151ProjectRevocationRecorder) CleanupProjectMessengerState(
	_ context.Context, scope project.MessengerProjectAccessScope,
) error {
	r.effects = append(r.effects, phase151ProjectEffect("mongo", scope))
	if len(r.cleanupErrors) == 0 {
		return nil
	}
	err := r.cleanupErrors[0]
	r.cleanupErrors = r.cleanupErrors[1:]
	return err
}

func (r *phase151ProjectRevocationRecorder) InvalidateProjectMessengerUserRoom(
	_ context.Context, scope project.MessengerProjectAccessScope,
) error {
	r.effects = append(r.effects, phase151ProjectEffect("room", scope))
	return nil
}

func TestPhase151ProjectMemberRemoval_RevokesPGBeforeCleanupAndRoomInvalidation(t *testing.T) {
	tenantID, projectID, userID := uuid.New(), uuid.New(), uuid.New()
	scope := project.MessengerProjectAccessScope{TenantID: tenantID, ProjectID: projectID, UserID: userID, LifecycleGeneration: 7}
	recorder := &phase151ProjectRevocationRecorder{}
	revoker := project.NewMessengerAccessRevoker(project.MessengerAccessRevokerDependencies{
		Authority: recorder, DurableState: recorder, UserRooms: recorder, ProjectMembers: recorder,
	})

	err := revoker.RevokeProjectMember(context.Background(), project.ProjectMemberRemoval{
		TenantID: tenantID, ProjectID: projectID, UserID: userID, LifecycleGeneration: 7,
	})

	require.NoError(t, err)
	assert.Equal(t, []string{
		phase151ProjectEffect("pg", scope),
		phase151ProjectEffect("mongo", scope),
		phase151ProjectEffect("room", scope),
	}, recorder.effects)
}

func TestPhase151ProjectLifecycleLoss_RevokesAllPGAuthorityBeforeAnyCleanup(t *testing.T) {
	for _, reason := range []project.MessengerProjectAccessLossReason{
		project.MessengerProjectArchived,
		project.MessengerProjectDisabled,
		project.MessengerProjectIneligible,
	} {
		t.Run(string(reason), func(t *testing.T) {
			tenantID, projectID := uuid.New(), uuid.New()
			firstUserID, secondUserID := uuid.New(), uuid.New()
			recorder := &phase151ProjectRevocationRecorder{members: []uuid.UUID{firstUserID, secondUserID}}
			revoker := project.NewMessengerAccessRevoker(project.MessengerAccessRevokerDependencies{
				Authority: recorder, DurableState: recorder, UserRooms: recorder, ProjectMembers: recorder,
			})

			err := revoker.RevokeProjectAccess(context.Background(), project.MessengerProjectAccessLoss{
				TenantID: tenantID, ProjectID: projectID, Reason: reason, LifecycleGeneration: 11,
			})

			require.NoError(t, err)
			first := project.MessengerProjectAccessScope{TenantID: tenantID, ProjectID: projectID, UserID: firstUserID, LifecycleGeneration: 11}
			second := project.MessengerProjectAccessScope{TenantID: tenantID, ProjectID: projectID, UserID: secondUserID, LifecycleGeneration: 11}
			assert.Equal(t, []string{
				"list:" + tenantID.String() + ":" + projectID.String(),
				phase151ProjectEffect("pg", first), phase151ProjectEffect("pg", second),
				phase151ProjectEffect("mongo", first), phase151ProjectEffect("mongo", second),
				phase151ProjectEffect("room", first), phase151ProjectEffect("room", second),
			}, recorder.effects)
			for _, effect := range recorder.effects[1:] {
				assert.Contains(t, effect, tenantID.String(), "no cross-tenant effect")
			}
		})
	}
}

func TestPhase151ProjectRevocation_CleanupFailureLeavesRetryableRevokedResidual(t *testing.T) {
	tenantID, projectID, userID := uuid.New(), uuid.New(), uuid.New()
	scope := project.MessengerProjectAccessScope{TenantID: tenantID, ProjectID: projectID, UserID: userID, LifecycleGeneration: 13}
	recorder := &phase151ProjectRevocationRecorder{cleanupErrors: []error{errors.New("mongo unavailable")}}
	revoker := project.NewMessengerAccessRevoker(project.MessengerAccessRevokerDependencies{
		Authority: recorder, DurableState: recorder, UserRooms: recorder, ProjectMembers: recorder,
	})
	input := project.ProjectMemberRemoval{TenantID: tenantID, ProjectID: projectID, UserID: userID, LifecycleGeneration: 13}

	err := revoker.RevokeProjectMember(context.Background(), input)
	var residual *project.MessengerAccessRevocationError
	require.ErrorAs(t, err, &residual)
	assert.Equal(t, project.MessengerAccessCleanupStage, residual.Stage)
	assert.True(t, residual.Retryable)
	assert.Equal(t, []string{phase151ProjectEffect("pg", scope), phase151ProjectEffect("mongo", scope)}, recorder.effects)

	require.NoError(t, revoker.RevokeProjectMember(context.Background(), input))
	assert.Equal(t, []string{
		phase151ProjectEffect("pg", scope), phase151ProjectEffect("mongo", scope),
		phase151ProjectEffect("pg", scope), phase151ProjectEffect("mongo", scope), phase151ProjectEffect("room", scope),
	}, recorder.effects)
}

func phase151ProjectEffect(stage string, scope project.MessengerProjectAccessScope) string {
	return fmt.Sprintf("%s:%s:%s:%s:%d", stage, scope.TenantID, scope.ProjectID, scope.UserID, scope.LifecycleGeneration)
}
