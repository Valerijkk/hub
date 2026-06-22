package storage_test
import (
	"github.com/valerii/hub/backend/internal/storage"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/internal/platform/config"
	"github.com/valerii/hub/backend/internal/rbac"
)

type storageRBACEnv struct {
	app       *fiber.App
	client    *ent.Client
	projectID uuid.UUID
	ownerID   uuid.UUID
	viewerID  uuid.UUID
}

func newStorageRBACEnv(t *testing.T, role string) *storageRBACEnv {
	t.Helper()
	dsn := "file:str117_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
	ctx := context.Background()

	owner, err := client.User.Create().SetEmail("own117-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
	require.NoError(t, err)
	viewer, err := client.User.Create().SetEmail("view117-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().SetName("Storage Wave117").SetKey("S117").Save(ctx)
	require.NoError(t, err)

	for _, m := range []struct {
		id   uuid.UUID
		role projectmember.ProjectRole
	}{
		{owner.ID, projectmember.ProjectRoleOwner},
		{viewer.ID, projectmember.ProjectRoleViewer},
	} {
		_, err = client.ProjectMember.Create().
			SetUserID(m.id).SetProjectID(proj.ID).SetProjectRole(m.role).Save(ctx)
		require.NoError(t, err)
	}

	actingID := owner.ID
	switch role {
	case "viewer":
		actingID = viewer.ID
	case "stranger":
		stranger, err := client.User.Create().SetEmail("str117-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
		require.NoError(t, err)
		actingID = stranger.ID
	}

	svc, err := storage.NewStorageService(config.S3Config{
		Region: "us-east-1", Endpoint: "http://127.0.0.1:9999",
		AccessKey: "test", SecretKey: "test", Bucket: "hub-files",
	}, storage.NewEntRepository(client))
	require.NoError(t, err)

	h := storage.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	uid := actingID
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", uid.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	t.Cleanup(func() {
		_ = app.Shutdown()
		_ = client.Close()
	})

	return &storageRBACEnv{
		app: app, client: client, projectID: proj.ID,
		ownerID: owner.ID, viewerID: viewer.ID,
	}
}

func TestHandler_GenerateUploadURL_ForbiddenForViewer(t *testing.T) {
	env := newStorageRBACEnv(t, "viewer")
	payload := fmt.Sprintf(`{"filename":"notes.txt","contentType":"text/plain","projectId":%q}`, env.projectID.String())
	req := httptest.NewRequest(http.MethodPost, "/api/files/upload-url", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_ConfirmUpload_InvalidJSON(t *testing.T) {
	app, projectID, userID, client := newStorageRoutesEnv(t)
	ctx := context.Background()

	fileID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("doc.txt").
		SetContentType("text/plain").
		SetSize(0).
		SetS3Key(projectID.String()+"/files/"+fileID.String()+"/doc.txt").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/files/"+fileID.String()+"/confirm", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_DeleteFile_ForbiddenForViewer(t *testing.T) {
	env := newStorageRBACEnv(t, "viewer")
	ctx := context.Background()

	fileID := uuid.New()
	_, err := env.client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("secret.txt").
		SetContentType("text/plain").
		SetSize(1).
		SetS3Key("key").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(env.ownerID).
		SetProjectID(env.projectID).
		Save(ctx)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/files/"+fileID.String(), nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_GetDownloadURL_NotProjectMember(t *testing.T) {
	env := newStorageRBACEnv(t, "stranger")
	ctx := context.Background()

	fileID := uuid.New()
	_, err := env.client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("doc.txt").
		SetContentType("text/plain").
		SetSize(1).
		SetS3Key("key").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(env.ownerID).
		SetProjectID(env.projectID).
		Save(ctx)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/files/"+fileID.String()+"/url", nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUserIDFromCtx_InvalidUUID(t *testing.T) {
	app := fiber.New()
	t.Cleanup(func() { _ = app.Shutdown() })

	app.Get("/uid", func(c fiber.Ctx) error {
		c.Locals("userID", "not-a-uuid")
		_, err := storage.ExportUserIDFromCtx(c)
		require.Error(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, err.(*fiber.Error).Code)
		return c.SendStatus(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/uid", nil)
	_, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
}
