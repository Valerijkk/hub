//go:build hubtest

package storage_test

import (
	"context"
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
	"github.com/valerii/hub/backend/internal/storage"
)

// softDeleteNotFoundRepo — ProjectIDForActiveFile OK, SoftDelete → NotFound (race/повтор).
type softDeleteNotFoundRepo struct {
	inner     storage.Repository
	projectID uuid.UUID
}

func (r *softDeleteNotFoundRepo) CreatePending(ctx context.Context, input storage.CreateFileInput) (*ent.FileMetadata, error) {
	return r.inner.CreatePending(ctx, input)
}
func (r *softDeleteNotFoundRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.inner.DeleteByID(ctx, id)
}
func (r *softDeleteNotFoundRepo) ConfirmUpload(ctx context.Context, fileID uuid.UUID, entityType, entityID string) error {
	return r.inner.ConfirmUpload(ctx, fileID, entityType, entityID)
}
func (r *softDeleteNotFoundRepo) ProjectIDForActiveFile(ctx context.Context, fileID uuid.UUID) (uuid.UUID, error) {
	return r.projectID, nil
}
func (r *softDeleteNotFoundRepo) UploaderIDForActiveFile(ctx context.Context, fileID uuid.UUID) (uuid.UUID, error) {
	return r.inner.UploaderIDForActiveFile(ctx, fileID)
}
func (r *softDeleteNotFoundRepo) GetActiveByID(ctx context.Context, fileID uuid.UUID) (*ent.FileMetadata, error) {
	return r.inner.GetActiveByID(ctx, fileID)
}
func (r *softDeleteNotFoundRepo) GetPendingByID(ctx context.Context, fileID uuid.UUID) (*ent.FileMetadata, error) {
	return r.inner.GetPendingByID(ctx, fileID)
}
func (r *softDeleteNotFoundRepo) IssueProjectID(ctx context.Context, issueID uuid.UUID) (uuid.UUID, error) {
	return r.inner.IssueProjectID(ctx, issueID)
}
func (r *softDeleteNotFoundRepo) IsChannelMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	return r.inner.IsChannelMember(ctx, channelID, userID)
}
func (r *softDeleteNotFoundRepo) ActiveChannelProjectID(ctx context.Context, channelID uuid.UUID) (*uuid.UUID, error) {
	return r.inner.ActiveChannelProjectID(ctx, channelID)
}
func (r *softDeleteNotFoundRepo) SoftDelete(ctx context.Context, fileID uuid.UUID, at time.Time) error {
	return &ent.NotFoundError{}
}
func (r *softDeleteNotFoundRepo) ListOrphanPending(ctx context.Context, olderThan time.Time, limit int) ([]*ent.FileMetadata, error) {
	return r.inner.ListOrphanPending(ctx, olderThan, limit)
}

// TestHandler_DeleteFile_SoftDeleteNotFound404_A3W74_004 — SoftDelete NotFound → 404, не 500.
// Mutation: убрать ent.IsNotFound map в Delete → StatusInternalServerError.
func TestHandler_DeleteFile_SoftDeleteNotFound404_A3W74_004(t *testing.T) {
	dsn := "file:del404_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
	ctx := context.Background()

	owner, err := client.User.Create().SetEmail("del404-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().SetName("Del404").SetKey("D404").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).SetProjectID(proj.ID).SetProjectRole(projectmember.ProjectRoleOwner).Save(ctx)
	require.NoError(t, err)

	fileID := uuid.New()
	_, err = client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("x.txt").
		SetContentType("text/plain").
		SetSize(1).
		SetS3Key("key").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(owner.ID).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	inner := storage.NewEntRepository(client)
	stub := &softDeleteNotFoundRepo{inner: inner, projectID: proj.ID}
	svc, err := storage.NewStorageService(config.S3Config{
		Region: "us-east-1", Endpoint: "http://127.0.0.1:9999",
		AccessKey: "test", SecretKey: "test", Bucket: "hub-files",
	}, stub)
	require.NoError(t, err)

	h := storage.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)
	t.Cleanup(func() { _ = app.Shutdown() })

	req := httptest.NewRequest(http.MethodDelete, "/api/files/"+fileID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}
