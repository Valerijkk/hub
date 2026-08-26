package audit

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// MessengerUserAccessScope identifies one generation of a user's tenant-
// scoped Messenger access. Every downstream effect receives the generation so
// a delayed retry cannot target a newer lifecycle.
type MessengerUserAccessScope struct {
	TenantID            uuid.UUID
	OrganizationID      uuid.UUID
	UserID              uuid.UUID
	LifecycleGeneration int64
}

type MessengerUserAccessLossReason string

const (
	MessengerUserAccountDisabled     MessengerUserAccessLossReason = "account_disabled"
	MessengerUserAccountDeleted      MessengerUserAccessLossReason = "account_deleted"
	MessengerUserOrganizationRemoved MessengerUserAccessLossReason = "organization_removed"
)

type MessengerUserAccessLoss struct {
	TenantID            uuid.UUID
	OrganizationID      uuid.UUID
	UserID              uuid.UUID
	Reason              MessengerUserAccessLossReason
	LifecycleGeneration int64
}

type MessengerUserAuthority interface {
	RevokeUserMessengerAuthority(context.Context, MessengerUserAccessScope) error
}

type MessengerUserDurableState interface {
	CleanupUserMessengerState(context.Context, MessengerUserAccessScope) error
}

type MessengerUserRooms interface {
	InvalidateUserMessengerRooms(context.Context, MessengerUserAccessScope) error
}

type MessengerUserAccessRevokerDependencies struct {
	Authority    MessengerUserAuthority
	DurableState MessengerUserDurableState
	UserRooms    MessengerUserRooms
}

type MessengerUserAccessRevocationStage string

const (
	MessengerUserAuthorityStage        MessengerUserAccessRevocationStage = "revoke_user_authority"
	MessengerUserCleanupStage          MessengerUserAccessRevocationStage = "cleanup_durable_state"
	MessengerUserRoomInvalidationStage MessengerUserAccessRevocationStage = "invalidate_user_rooms"
)

type MessengerUserAccessRevocationError struct {
	Stage     MessengerUserAccessRevocationStage
	Retryable bool
	Err       error
}

func (e *MessengerUserAccessRevocationError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("messenger user access revocation failed at %s: %v", e.Stage, e.Err)
}

func (e *MessengerUserAccessRevocationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type MessengerUserAccessRevoker struct {
	authority    MessengerUserAuthority
	durableState MessengerUserDurableState
	userRooms    MessengerUserRooms
}

func NewMessengerUserAccessRevoker(dependencies MessengerUserAccessRevokerDependencies) *MessengerUserAccessRevoker {
	return &MessengerUserAccessRevoker{
		authority:    dependencies.Authority,
		durableState: dependencies.DurableState,
		userRooms:    dependencies.UserRooms,
	}
}

// RevokeUserAccess is monotonic: PostgreSQL authority is removed first, then
// durable Messenger state, then transient room membership. A failure never
// compensates an earlier revocation; retry safely replays the ordered effects.
func (r *MessengerUserAccessRevoker) RevokeUserAccess(
	ctx context.Context,
	loss MessengerUserAccessLoss,
) error {
	scope := MessengerUserAccessScope{
		TenantID:            loss.TenantID,
		OrganizationID:      loss.OrganizationID,
		UserID:              loss.UserID,
		LifecycleGeneration: loss.LifecycleGeneration,
	}

	if r == nil || r.authority == nil {
		return newMessengerUserAccessRevocationError(MessengerUserAuthorityStage, errors.New("messenger authority is required"))
	}
	if err := r.authority.RevokeUserMessengerAuthority(ctx, scope); err != nil {
		return newMessengerUserAccessRevocationError(MessengerUserAuthorityStage, err)
	}
	if r.durableState == nil {
		return newMessengerUserAccessRevocationError(MessengerUserCleanupStage, errors.New("messenger durable state is required"))
	}
	if err := r.durableState.CleanupUserMessengerState(ctx, scope); err != nil {
		return newMessengerUserAccessRevocationError(MessengerUserCleanupStage, err)
	}
	if r.userRooms == nil {
		return newMessengerUserAccessRevocationError(MessengerUserRoomInvalidationStage, errors.New("messenger user rooms are required"))
	}
	if err := r.userRooms.InvalidateUserMessengerRooms(ctx, scope); err != nil {
		return newMessengerUserAccessRevocationError(MessengerUserRoomInvalidationStage, err)
	}
	return nil
}

func newMessengerUserAccessRevocationError(
	stage MessengerUserAccessRevocationStage,
	err error,
) *MessengerUserAccessRevocationError {
	return &MessengerUserAccessRevocationError{Stage: stage, Retryable: true, Err: err}
}
