package project_test

// TASK-042 / P2-BE-12: Get deleted → 404; javascript: avatar → 400.
import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/project"
	projmod "github.com/valerii/hub/backend/internal/project"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

func TestGetProject_Deleted_NotFound_P2_BE_12(t *testing.T) {
	dsn := "file:projdel_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("p-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
	require.NoError(t, err)
	p, err := client.Project.Create().SetName("Trash").SetKey("TR").SetStatus(project.StatusDeleted).Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().SetUserID(u.ID).SetProjectID(p.ID).SetProjectRole("owner").Save(ctx)
	require.NoError(t, err)

	svc := projmod.NewProjectService(projmod.NewEntRepository(client, nil), nil, nil)
	_, err = svc.GetProject(ctx, p.ID, u.ID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 404, appErr.Code)
}

func TestUpdateProject_JavascriptAvatar_Rejected_P2_BE_12(t *testing.T) {
	dsn := "file:projav_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("a-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
	require.NoError(t, err)
	p, err := client.Project.Create().SetName("Avatar").SetKey("AV").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().SetUserID(u.ID).SetProjectID(p.ID).SetProjectRole("owner").Save(ctx)
	require.NoError(t, err)

	svc := projmod.NewProjectService(projmod.NewEntRepository(client, nil), nil, nil)
	bad := "javascript:alert(1)"
	_, err = svc.UpdateProject(ctx, p.ID, projmod.UpdateProjectRequest{AvatarURL: &bad, Version: 1})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 400, appErr.Code)
}
