package storage_test
import (
	"github.com/valerii/hub/backend/internal/storage"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/internal/platform/config"
	"github.com/valerii/hub/backend/internal/rbac"
)

func newStorageRoutesEnv(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID, *ent.Client) {
	t.Helper()
	dsn := "file:strroutes_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("stor-" + uuid.New().String()[:6] + "@ex.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Storage Routes").
		SetKey("SR").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	base, err := storage.NewStorageService(config.S3Config{
		Region:    "us-east-1",
		Endpoint:  "http://127.0.0.1:9999",
		AccessKey: "test",
		SecretKey: "test",
		Bucket:    "hub-files",
	}, storage.NewEntRepository(client))
	require.NoError(t, err)
	// Unit-тесты без SeaweedFS: s3Client=nil → Confirm без HEAD; PresignClient остаётся.
	svc := storage.ExportNewStorageServiceForTestWithS3(
		storage.NewEntRepository(client),
		"hub-files",
		nil,
		s3.NewPresignClient(base.S3Client()),
	)

	h := storage.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	t.Cleanup(func() {
		_ = app.Shutdown()
		_ = client.Close()
	})

	return app, proj.ID, owner.ID, client
}

func TestNewHandler_AndRegisterRoutes(t *testing.T) {
	app, projectID, _, client := newStorageRoutesEnv(t)
	_ = client

	payload := fmt.Sprintf(`{"filename":"notes.txt","contentType":"text/plain","projectId":%q,"size":12}`, projectID.String())
	req := httptest.NewRequest(http.MethodPost, "/api/files/upload-url", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result storage.UploadResult
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEqual(t, uuid.Nil, result.FileID)
	assert.NotEmpty(t, result.UploadURL)
}

func TestHandler_RegisterRoutes_ConfirmAndDownload(t *testing.T) {
	app, projectID, userID, client := newStorageRoutesEnv(t)
	ctx := context.Background()

	fileID := uuid.New()
	s3Key := projectID.String() + "/files/" + fileID.String() + "/doc.txt"
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("doc.txt").
		SetContentType("text/plain").
		SetSize(0).
		SetS3Key(s3Key).
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	confirmBody := `{"entityType":"message","entityId":"msg-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/files/"+fileID.String()+"/confirm", bytes.NewReader([]byte(confirmBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/files/"+fileID.String()+"/url", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, "/api/files/"+fileID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_RegisterRoutes_Unauthenticated(t *testing.T) {
	client, _, _, _ := newStorageTestClient(t)
	svc := newStorageServiceWithPresign(t, client)
	h := storage.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() }, nil)

	payload := `{"filename":"a.txt","contentType":"text/plain","projectId":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/files/upload-url", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestNewStorageService_InvalidConfigStillBuilds(t *testing.T) {
	client, _, _, _ := newStorageTestClient(t)
	svc, err := storage.NewStorageService(config.S3Config{
		Region:    "us-east-1",
		Endpoint:  "http://127.0.0.1:9999",
		AccessKey: "ak",
		SecretKey: "sk",
		Bucket:    "hub-files",
	}, storage.NewEntRepository(client))
	require.NoError(t, err)
	require.NotNil(t, svc)
	assert.Equal(t, "hub-files", svc.Bucket())
}
