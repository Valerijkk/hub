package calls

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/ent/user"
)

var _ CallJoinCredentialProjectionSource = (*EntRepository)(nil)

// LoadActiveCall reloads a call that is currently active.
func (r *EntRepository) LoadActiveCall(ctx context.Context, callID uuid.UUID) (*ent.Call, error) {
	if r == nil || r.db == nil || isNilCallJoinCredentialProjectionDependency(ctx) || callID == uuid.Nil {
		return nil, ErrCallJoinCredentialInvalid
	}

	return r.db.Call.Query().
		Where(
			call.IDEQ(callID),
			call.StatusEQ(call.StatusActive),
		).
		Only(ctx)
}

// GetServerKnownCallUserName reloads the active user's server-owned display name.
func (r *EntRepository) GetServerKnownCallUserName(ctx context.Context, authenticatedUserID uuid.UUID) (string, error) {
	if r == nil || r.db == nil || isNilCallJoinCredentialProjectionDependency(ctx) || authenticatedUserID == uuid.Nil {
		return "", ErrCallJoinCredentialInvalid
	}

	userEntity, err := r.db.User.Query().
		Where(
			user.IDEQ(authenticatedUserID),
			user.StatusEQ(user.StatusActive),
		).
		Only(ctx)
	if err != nil {
		return "", err
	}

	if displayName := strings.TrimSpace(userEntity.DisplayName); displayName != "" {
		return displayName, nil
	}
	if name := strings.TrimSpace(userEntity.Name); name != "" {
		return name, nil
	}

	return "", ErrCallJoinCredentialInvalid
}
