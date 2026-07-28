package ai

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/rbac"
)

// ProjectAccessChecker — membership/RBAC view для bot issue/review (P1-BE-31).
type ProjectAccessChecker interface {
	CanViewProject(ctx context.Context, userID, projectID uuid.UUID) (bool, error)
}

// RBACProjectAccess адаптирует rbac.RBACService к ProjectAccessChecker.
type RBACProjectAccess struct {
	Svc *rbac.RBACService
}

// CanViewProject — true, если user имеет PermProjectView (или system admin).
func (a RBACProjectAccess) CanViewProject(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	if a.Svc == nil {
		return false, nil
	}
	eff, err := a.Svc.GetEffectivePermissions(ctx, userID, projectID)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return eff.Has(rbac.PermProjectView), nil
}
