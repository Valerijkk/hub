package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// MessengerProjectAccessScope identifies one user's messenger access to a
// project. LifecycleGeneration is forwarded to every consumer-owned port so a
// delayed side effect cannot be mistaken for a newer project lifecycle.
type MessengerProjectAccessScope struct {
	TenantID             uuid.UUID
	ProjectID            uuid.UUID
	UserID               uuid.UUID
	LifecycleGeneration  int64
}

// ProjectMemberRemoval is emitted when one member loses project membership.
type ProjectMemberRemoval struct {
	TenantID            uuid.UUID
	ProjectID           uuid.UUID
	UserID              uuid.UUID
	LifecycleGeneration int64
}

// MessengerProjectAccessLossReason explains why all project members lose
// messenger access.
type MessengerProjectAccessLossReason string

const (
	MessengerProjectArchived   MessengerProjectAccessLossReason = "archived"
	MessengerProjectDisabled   MessengerProjectAccessLossReason = "disabled"
	MessengerProjectIneligible MessengerProjectAccessLossReason = "ineligible"
)

// MessengerProjectAccessLoss is emitted when access is removed for every
// eligible project member.
type MessengerProjectAccessLoss struct {
	TenantID            uuid.UUID
	ProjectID           uuid.UUID
	Reason              MessengerProjectAccessLossReason
	LifecycleGeneration int64
}

// MessengerProjectAuthority owns project-group authority revocation.
type MessengerProjectAuthority interface {
	RevokeProjectMessengerAuthority(context.Context, MessengerProjectAccessScope) error
}

// MessengerProjectDurableState owns persistent messenger state cleanup.
type MessengerProjectDurableState interface {
	CleanupProjectMessengerState(context.Context, MessengerProjectAccessScope) error
}

// MessengerProjectUserRooms owns invalidation of a user's project rooms.
type MessengerProjectUserRooms interface {
	InvalidateProjectMessengerUserRoom(context.Context, MessengerProjectAccessScope) error
}

// MessengerProjectMembers lists the users that currently have messenger
// access through a project.
type MessengerProjectMembers interface {
	ListProjectMessengerUserIDs(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error)
}

// MessengerAccessRevokerDependencies contains consumer-owned ports only. The
// project package coordinates their order but never imports their owners.
type MessengerAccessRevokerDependencies struct {
	Authority      MessengerProjectAuthority
	DurableState   MessengerProjectDurableState
	UserRooms      MessengerProjectUserRooms
	ProjectMembers MessengerProjectMembers
}

// MessengerAccessRevocationStage pinpoints the failed external side effect.
type MessengerAccessRevocationStage string

const (
	MessengerAccessListStage       MessengerAccessRevocationStage = "list_project_members"
	MessengerAccessAuthorityStage  MessengerAccessRevocationStage = "revoke_project_authority"
	MessengerAccessCleanupStage    MessengerAccessRevocationStage = "cleanup_durable_state"
	MessengerAccessRoomStage       MessengerAccessRevocationStage = "invalidate_user_room"
)

// MessengerAccessRevocationError marks a failed revocation side effect. Port
// errors are retryable because a subsequent delivery must safely replay every
// preceding phase before it reaches the failed one.
type MessengerAccessRevocationError struct {
	Stage     MessengerAccessRevocationStage
	Retryable bool
	Err       error
}

func (e *MessengerAccessRevocationError) Error() string {
	if e == nil {
		return ""
	}

	return fmt.Sprintf("messenger project access revocation failed at %s: %v", e.Stage, e.Err)
}

func (e *MessengerAccessRevocationError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

// MessengerAccessRevoker coordinates messenger cleanup after project access
// is lost.
type MessengerAccessRevoker struct {
	authority      MessengerProjectAuthority
	durableState   MessengerProjectDurableState
	userRooms      MessengerProjectUserRooms
	projectMembers MessengerProjectMembers
}

func NewMessengerAccessRevoker(dependencies MessengerAccessRevokerDependencies) *MessengerAccessRevoker {
	return &MessengerAccessRevoker{
		authority:      dependencies.Authority,
		durableState:   dependencies.DurableState,
		userRooms:      dependencies.UserRooms,
		projectMembers: dependencies.ProjectMembers,
	}
}

// RevokeProjectMember revokes project-group authority before removing durable
// messenger state and invalidating rooms for the removed member.
func (r *MessengerAccessRevoker) RevokeProjectMember(ctx context.Context, removal ProjectMemberRemoval) error {
	if r == nil || r.authority == nil {
		return newMessengerAccessRevocationError(MessengerAccessAuthorityStage, errors.New("messenger project authority is required"))
	}
	if r.durableState == nil {
		return newMessengerAccessRevocationError(MessengerAccessCleanupStage, errors.New("messenger project durable state is required"))
	}
	if r.userRooms == nil {
		return newMessengerAccessRevocationError(MessengerAccessRoomStage, errors.New("messenger project user rooms are required"))
	}
	scope := MessengerProjectAccessScope{
		TenantID:            removal.TenantID,
		ProjectID:           removal.ProjectID,
		UserID:              removal.UserID,
		LifecycleGeneration: removal.LifecycleGeneration,
	}

	if err := r.authority.RevokeProjectMessengerAuthority(ctx, scope); err != nil {
		return newMessengerAccessRevocationError(MessengerAccessAuthorityStage, err)
	}
	if err := r.durableState.CleanupProjectMessengerState(ctx, scope); err != nil {
		return newMessengerAccessRevocationError(MessengerAccessCleanupStage, err)
	}
	if err := r.userRooms.InvalidateProjectMessengerUserRoom(ctx, scope); err != nil {
		return newMessengerAccessRevocationError(MessengerAccessRoomStage, err)
	}

	return nil
}

// RevokeProjectAccess first revokes authority for every member. Only after all
// authority revocations succeed may it clean durable state; room invalidation
// similarly begins only after all durable cleanup succeeds.
func (r *MessengerAccessRevoker) RevokeProjectAccess(ctx context.Context, loss MessengerProjectAccessLoss) error {
	if r == nil || r.projectMembers == nil {
		return newMessengerAccessRevocationError(MessengerAccessListStage, errors.New("messenger project members are required"))
	}
	if r.authority == nil {
		return newMessengerAccessRevocationError(MessengerAccessAuthorityStage, errors.New("messenger project authority is required"))
	}
	if r.durableState == nil {
		return newMessengerAccessRevocationError(MessengerAccessCleanupStage, errors.New("messenger project durable state is required"))
	}
	if r.userRooms == nil {
		return newMessengerAccessRevocationError(MessengerAccessRoomStage, errors.New("messenger project user rooms are required"))
	}
	userIDs, err := r.projectMembers.ListProjectMessengerUserIDs(ctx, loss.TenantID, loss.ProjectID)
	if err != nil {
		return newMessengerAccessRevocationError(MessengerAccessListStage, err)
	}

	scopes := make([]MessengerProjectAccessScope, 0, len(userIDs))
	for _, userID := range userIDs {
		scopes = append(scopes, MessengerProjectAccessScope{
			TenantID:            loss.TenantID,
			ProjectID:           loss.ProjectID,
			UserID:              userID,
			LifecycleGeneration: loss.LifecycleGeneration,
		})
	}

	for _, scope := range scopes {
		if err := r.authority.RevokeProjectMessengerAuthority(ctx, scope); err != nil {
			return newMessengerAccessRevocationError(MessengerAccessAuthorityStage, err)
		}
	}
	for _, scope := range scopes {
		if err := r.durableState.CleanupProjectMessengerState(ctx, scope); err != nil {
			return newMessengerAccessRevocationError(MessengerAccessCleanupStage, err)
		}
	}
	for _, scope := range scopes {
		if err := r.userRooms.InvalidateProjectMessengerUserRoom(ctx, scope); err != nil {
			return newMessengerAccessRevocationError(MessengerAccessRoomStage, err)
		}
	}

	return nil
}

func newMessengerAccessRevocationError(stage MessengerAccessRevocationStage, err error) *MessengerAccessRevocationError {
	return &MessengerAccessRevocationError{
		Stage:     stage,
		Retryable: true,
		Err:       err,
	}
}
