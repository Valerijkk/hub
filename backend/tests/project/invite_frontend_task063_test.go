//go:build hubtest

package project_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mutation: CreateInvite с toInviteResponse(..., "") → Link без frontend base → Contains краснеет.
func TestCreateInvite_FrontendBase_P3BE09(t *testing.T) {
	svc, _, _ := newTestService()
	svc.SetFrontendURL("https://hub.example.com")
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Invite Link", "INVLINK", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	inv, err := svc.CreateInvite(ctx, projectID, ownerID, "developer", 5, nil)
	require.NoError(t, err)
	assert.Equal(t, "https://hub.example.com/invite/"+inv.Code, inv.Link)
}

func TestCreateInvite_FrontendURLEnv_P3BE09(t *testing.T) {
	t.Setenv("FRONTEND_URL", "https://env.hub.test/")
	svc, _, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Env Invite", "ENVINV", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	inv, err := svc.CreateInvite(ctx, projectID, ownerID, "viewer", 1, nil)
	require.NoError(t, err)
	assert.Equal(t, "https://env.hub.test/invite/"+inv.Code, inv.Link)
}
