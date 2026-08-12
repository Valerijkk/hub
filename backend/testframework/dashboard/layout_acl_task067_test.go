//go:build hubtest

package dashboard_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/dashboard"
)

// TestApplyTemplate_StrangerForbidden_Task067 — W7-A4-016 / P3-BE-15:
// ApplyTemplate stranger → ErrForbidden (403 на handler).
// Мутация: убрать IsProjectMember из ApplyTemplate/SaveLayout → stranger OK.
func TestApplyTemplate_StrangerForbidden_Task067(t *testing.T) {
	svc, client, ctx, userID := newDashboardTestEnv(t)

	proj, err := client.Project.Create().SetName("Foreign").SetKey("FRN").Save(ctx)
	require.NoError(t, err)

	stranger, err := client.User.Create().
		SetEmail("stranger-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)
	_ = userID

	err = svc.ApplyTemplate(ctx, stranger.ID, proj.ID, "dev")
	require.Error(t, err)
	assert.True(t, errors.Is(err, dashboard.ErrForbidden))

	// Member succeeds.
	_, err = client.ProjectMember.Create().
		SetUserID(stranger.ID).
		SetProjectID(proj.ID).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, svc.ApplyTemplate(ctx, stranger.ID, proj.ID, "dev"))
}

// TestSaveLayout_StrangerProjectForbidden_Task067 — layout ACL на SaveLayout с project_id.
func TestSaveLayout_StrangerProjectForbidden_Task067(t *testing.T) {
	svc, client, ctx, _ := newDashboardTestEnv(t)
	proj, err := client.Project.Create().SetName("LayoutACL").SetKey("LAC").Save(ctx)
	require.NoError(t, err)

	stranger, err := client.User.Create().
		SetEmail("layout-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	err = svc.SaveLayout(ctx, stranger.ID, dashboard.SaveLayoutRequest{
		DashboardType: "project",
		ProjectID:     &proj.ID,
		Items:         []dashboard.LayoutItem{{I: "w1", X: 0, Y: 0, W: 6, H: 4, WidgetType: "my-tasks"}},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, dashboard.ErrForbidden))
}
