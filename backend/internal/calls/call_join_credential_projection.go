package calls

import (
	"context"
	"errors"
	"reflect"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	entcall "github.com/valerii/hub/backend/ent/call"
)

// CallJoinCredentialProjectionSource is the server-owned authority used to
// reload and authorize a credential-issuance request. Caller-controlled room,
// display-name, and call facts must not be supplied to the projector.
type CallJoinCredentialProjectionSource interface {
	LoadActiveCall(ctx context.Context, callID uuid.UUID) (*ent.Call, error)
	IsChannelMember(ctx context.Context, channelID, authenticatedUserID uuid.UUID) (bool, error)
	GetServerKnownCallUserName(ctx context.Context, authenticatedUserID uuid.UUID) (string, error)
}

// CallJoinCredentialProjector reloads and authorizes server-owned call facts
// before delegating the final provider-specific issuance to its typed issuer.
type CallJoinCredentialProjector struct {
	source CallJoinCredentialProjectionSource
	issuer CallJoinCredentialIssuer
}

// NewCallJoinCredentialProjector constructs a call join credential projector.
func NewCallJoinCredentialProjector(
	source CallJoinCredentialProjectionSource,
	issuer CallJoinCredentialIssuer,
) (*CallJoinCredentialProjector, error) {
	if isNilCallJoinCredentialProjectionDependency(source) {
		return nil, errors.New("call join credential projection source is required")
	}
	if isNilCallJoinCredentialProjectionDependency(issuer) {
		return nil, errors.New("call join credential issuer is required")
	}

	return &CallJoinCredentialProjector{
		source: source,
		issuer: issuer,
	}, nil
}

// IssueForAuthenticatedUser reloads active call facts, authorizes the caller,
// loads the persisted display name, then delegates credential issuance once.
func (p *CallJoinCredentialProjector) IssueForAuthenticatedUser(
	ctx context.Context,
	callID uuid.UUID,
	authenticatedUserID uuid.UUID,
) (JoinCredential, error) {
	if isNilCallJoinCredentialProjectionDependency(ctx) ||
		callID == uuid.Nil ||
		authenticatedUserID == uuid.Nil {
		return JoinCredential{}, ErrCallJoinCredentialInvalid
	}

	callEntity, err := p.source.LoadActiveCall(ctx, callID)
	if err != nil {
		return JoinCredential{}, err
	}
	if callEntity == nil ||
		callEntity.ID != callID ||
		callEntity.Status != entcall.StatusActive {
		return JoinCredential{}, ErrCallJoinCredentialInvalid
	}

	if callEntity.ChannelID != nil {
		isMember, err := p.source.IsChannelMember(ctx, *callEntity.ChannelID, authenticatedUserID)
		if err != nil {
			return JoinCredential{}, err
		}
		if !isMember {
			return JoinCredential{}, ErrCallForbidden
		}
	} else if callEntity.CreatedBy != authenticatedUserID {
		return JoinCredential{}, ErrCallForbidden
	}

	serverKnownUserName, err := p.source.GetServerKnownCallUserName(ctx, authenticatedUserID)
	if err != nil {
		return JoinCredential{}, err
	}

	credential, err := p.issuer.IssueCallJoinCredential(callEntity, authenticatedUserID, serverKnownUserName)
	if err != nil {
		return JoinCredential{}, err
	}
	return credential, nil
}

func isNilCallJoinCredentialProjectionDependency(value any) bool {
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
