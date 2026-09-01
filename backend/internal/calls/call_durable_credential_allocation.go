package calls

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	entcall "github.com/valerii/hub/backend/ent/call"
)

var (
	ErrCallJoinCredentialUnavailable     = errors.New("call join credential unavailable")
	ErrCallJoinCredentialCommitAmbiguous = errors.New("call join credential commit outcome ambiguous")
)

// DurableCallJoinCredentialAllocationSource begins one durable credential
// allocation transaction.
type DurableCallJoinCredentialAllocationSource interface {
	BeginDurableCallJoinCredentialAllocation(context.Context) (DurableCallJoinCredentialAllocationTx, error)
}

// DurableCallJoinCredentialAllocationTx owns the ordered durable allocation
// operations for one authenticated call join.
type DurableCallJoinCredentialAllocationTx interface {
	LockActiveUserName(context.Context, uuid.UUID) (string, error)
	LockActiveCall(context.Context, uuid.UUID) (*ent.Call, error)
	AuthorizeLockedCall(context.Context, *ent.Call, uuid.UUID) error
	CreateUnknownParticipantConnection(context.Context, CallParticipantConnectionKey) error
	Commit() error
	Rollback() error
}

// DurableCallJoinCredentialAllocator allocates one unknown durable connection
// before issuing a provider credential.
type DurableCallJoinCredentialAllocator struct {
	source DurableCallJoinCredentialAllocationSource
	issuer DurableCallJoinCredentialIssuer
}

// NewDurableCallJoinCredentialAllocator constructs a durable allocation seam.
func NewDurableCallJoinCredentialAllocator(
	source DurableCallJoinCredentialAllocationSource,
	issuer DurableCallJoinCredentialIssuer,
) (*DurableCallJoinCredentialAllocator, error) {
	if isNilDurableCredentialAllocationDependency(source) ||
		isNilDurableCredentialAllocationDependency(issuer) {
		return nil, ErrCallJoinCredentialInvalid
	}

	return &DurableCallJoinCredentialAllocator{source: source, issuer: issuer}, nil
}

// IssueForAuthenticatedUser allocates and commits a durable unknown connection
// before exposing a provider join credential.
func (a *DurableCallJoinCredentialAllocator) IssueForAuthenticatedUser(
	ctx context.Context,
	callID uuid.UUID,
	authenticatedUserID uuid.UUID,
) (JoinCredential, error) {
	if a == nil ||
		isNilDurableCredentialAllocationDependency(a.source) ||
		isNilDurableCredentialAllocationDependency(a.issuer) ||
		isNilDurableCredentialAllocationDependency(ctx) ||
		callID == uuid.Nil ||
		authenticatedUserID == uuid.Nil {
		return JoinCredential{}, ErrCallJoinCredentialInvalid
	}
	if ctx.Err() != nil {
		return JoinCredential{}, ErrCallJoinCredentialUnavailable
	}

	tx, err := a.source.BeginDurableCallJoinCredentialAllocation(ctx)
	if err != nil || isNilDurableCredentialAllocationDependency(tx) {
		return JoinCredential{}, ErrCallJoinCredentialUnavailable
	}

	rollback := true
	defer func() {
		if rollback {
			_ = tx.Rollback()
		}
	}()

	serverKnownUserName, err := tx.LockActiveUserName(ctx, authenticatedUserID)
	if err != nil {
		return JoinCredential{}, ErrCallJoinCredentialUnavailable
	}
	serverKnownUserName = strings.TrimSpace(serverKnownUserName)
	if serverKnownUserName == "" {
		return JoinCredential{}, ErrCallJoinCredentialInvalid
	}

	call, err := tx.LockActiveCall(ctx, callID)
	if err != nil {
		return JoinCredential{}, ErrCallJoinCredentialUnavailable
	}
	if !isValidLockedDurableCredentialCall(call, callID) {
		return JoinCredential{}, ErrCallJoinCredentialInvalid
	}

	if err := tx.AuthorizeLockedCall(ctx, call, authenticatedUserID); err != nil {
		if errors.Is(err, ErrCallForbidden) {
			return JoinCredential{}, ErrCallForbidden
		}
		return JoinCredential{}, ErrCallJoinCredentialUnavailable
	}

	connectionKey, err := NewCallParticipantConnectionKey(callID, authenticatedUserID, uuid.New())
	if err != nil {
		return JoinCredential{}, ErrCallJoinCredentialUnavailable
	}
	identity, err := NewProviderParticipantIdentity(connectionKey)
	if err != nil {
		return JoinCredential{}, ErrCallJoinCredentialUnavailable
	}
	if err := tx.CreateUnknownParticipantConnection(ctx, connectionKey); err != nil {
		return JoinCredential{}, ErrCallJoinCredentialUnavailable
	}

	credential, err := a.issuer.IssueDurableCallJoinCredential(call, identity, serverKnownUserName)
	if err != nil || credential == (JoinCredential{}) {
		return JoinCredential{}, ErrCallJoinCredentialUnavailable
	}
	if ctx.Err() != nil {
		return JoinCredential{}, ErrCallJoinCredentialUnavailable
	}

	rollback = false
	if err := tx.Commit(); err != nil {
		return JoinCredential{}, ErrCallJoinCredentialCommitAmbiguous
	}

	return credential, nil
}

func isValidLockedDurableCredentialCall(call *ent.Call, callID uuid.UUID) bool {
	if call == nil ||
		call.ID != callID ||
		call.ID == uuid.Nil ||
		call.Status != entcall.StatusActive ||
		strings.TrimSpace(call.RoomName) == "" ||
		call.RoomName != strings.TrimSpace(call.RoomName) {
		return false
	}

	return (call.ChannelID != nil && *call.ChannelID != uuid.Nil) ||
		(call.ChannelID == nil && call.CreatedBy != uuid.Nil)
}

func isNilDurableCredentialAllocationDependency(value any) bool {
	if value == nil {
		return true
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
