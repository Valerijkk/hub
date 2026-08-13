//go:build hubtest

package dashboard_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/dashboard"
)

// TestRequireSystemAdmin_Layering_Task062 — W7-A4-017 / P3-BE-07:
// admin-гейт на service, не handler→repo.
// Мутация: Always return nil из RequireSystemAdmin → member Admin handler 200.
func TestRequireSystemAdmin_Layering_Task062(t *testing.T) {
	svc, client, ctx, userID := newDashboardTestEnv(t)

	err := svc.RequireSystemAdmin(ctx, userID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dashboard.ErrForbidden))

	_, err = client.User.UpdateOneID(userID).SetSystemRole(user.SystemRoleAdmin).Save(ctx)
	require.NoError(t, err)
	require.NoError(t, svc.RequireSystemAdmin(ctx, userID))

	err = svc.RequireSystemAdmin(context.Background(), uuid.New())
	require.Error(t, err)
	assert.True(t, errors.Is(err, dashboard.ErrUnauthorized))
}
