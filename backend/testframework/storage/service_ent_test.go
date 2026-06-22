package storage_test
import (
	"github.com/valerii/hub/backend/internal/storage"
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/filemetadata"
)

func newStorageTestClient(t *testing.T) (*ent.Client, context.Context, uuid.UUID, uuid.UUID) {
	t.Helper()
	dsn := "file:storagetest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("storage-" + uuid.New().String()[:8] + "@test.com").
		SetName("Uploader").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Storage Project").
		SetKey("STG").
		Save(ctx)
	require.NoError(t, err)

	return client, ctx, user.ID, proj.ID
}

func newStorageServiceWithPresign(t *testing.T, client *ent.Client) *storage.StorageService {
	t.Helper()
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test-access", "test-secret", ""),
		),
	)
	require.NoError(t, err)

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("http://127.0.0.1:9999")
		o.UsePathStyle = true
	})

	return storage.ExportNewStorageServiceForTestWithS3(
		storage.NewEntRepository(client),
		"hub-files",
		s3Client,
		s3.NewPresignClient(s3Client),
	)
}

func storageRepo(client *ent.Client) storage.Repository {
	return storage.NewEntRepository(client)
}

func TestGenerateUploadURL_CreatesPendingMetadata(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := newStorageServiceWithPresign(t, client)

	result, err := svc.GenerateUploadURL(ctx, "report.pdf", "application/pdf", userID, projectID)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.FileID)
	assert.NotEmpty(t, result.UploadURL)
	assert.Contains(t, result.S3Key, projectID.String())
	assert.Contains(t, result.S3Key, "report.pdf")

	meta, err := client.FileMetadata.Get(ctx, result.FileID)
	require.NoError(t, err)
	assert.Equal(t, filemetadata.StatusPending, meta.Status)
	assert.Equal(t, "report.pdf", meta.Filename)

	uploaderID, err := meta.QueryUploader().OnlyID(ctx)
	require.NoError(t, err)
	assert.Equal(t, userID, uploaderID)

	gotProjectID, err := meta.QueryProject().OnlyID(ctx)
	require.NoError(t, err)
	assert.Equal(t, projectID, gotProjectID)
}

func TestGenerateDownloadURL_ReturnsPresignedURL(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := newStorageServiceWithPresign(t, client)

	fileID := uuid.New()
	s3Key := projectID.String() + "/files/" + fileID.String() + "/doc.txt"
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("doc.txt").
		SetContentType("text/plain").
		SetSize(12).
		SetS3Key(s3Key).
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	url, err := svc.GenerateDownloadURL(ctx, fileID)
	require.NoError(t, err)
	assert.Contains(t, url, "hub-files")
	assert.Contains(t, url, s3Key)
}

func TestConfirmUpload_NotFound(t *testing.T) {
	client, ctx, _, _ := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "")

	err := svc.ConfirmUpload(ctx, uuid.New(), "issue", "iss-1")
	require.Error(t, err)
}

func TestDeleteFile_NotFound(t *testing.T) {
	client, ctx, _, _ := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "")

	err := svc.DeleteFile(ctx, uuid.New())
	require.Error(t, err)
}

func TestStorageService_Accessors(t *testing.T) {
	client, _, _, _ := newStorageTestClient(t)
	svc := newStorageServiceWithPresign(t, client)

	assert.Equal(t, "hub-files", svc.Bucket())
	assert.NotNil(t, svc.S3Client())
	assert.NotNil(t, svc.NewOrphanCleanupFromService())
}

func TestConfirmUpload_UpdatesMetadata(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "hub-files")

	fileID := uuid.New()
	s3Key := projectID.String() + "/files/" + fileID.String() + "/report.pdf"
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("report.pdf").
		SetContentType("application/pdf").
		SetSize(0).
		SetS3Key(s3Key).
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.ConfirmUpload(ctx, fileID, "issue", "issue-123"))

	meta, err := client.FileMetadata.Get(ctx, fileID)
	require.NoError(t, err)
	assert.Equal(t, filemetadata.StatusConfirmed, meta.Status)
	assert.Equal(t, "issue", meta.EntityType)
	assert.Equal(t, "issue-123", meta.EntityID)
}

func TestFileProjectID_ReturnsOwnerProject(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "")

	fileID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("doc.txt").
		SetContentType("text/plain").
		SetSize(10).
		SetS3Key("key").
		SetS3Bucket("bucket").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	got, err := svc.FileProjectID(ctx, fileID)
	require.NoError(t, err)
	assert.Equal(t, projectID, got)
}

func TestFileProjectID_NotFound(t *testing.T) {
	client, ctx, _, _ := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "")

	_, err := svc.FileProjectID(ctx, uuid.New())
	require.Error(t, err)
}

func TestDeleteFile_SoftDeletes(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := storage.ExportNewStorageServiceForTest(storageRepo(client), "")

	fileID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("old.txt").
		SetContentType("text/plain").
		SetSize(5).
		SetS3Key("key").
		SetS3Bucket("bucket").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.DeleteFile(ctx, fileID))

	meta, err := client.FileMetadata.Get(ctx, fileID)
	require.NoError(t, err)
	assert.Equal(t, filemetadata.StatusDeleted, meta.Status)
	require.NotNil(t, meta.DeletedAt)
}

func TestHandler_AuthorizeProject_NoRBAC(t *testing.T) {
	h := storage.NewHandler(storage.ExportNewStorageServiceForTest(nil, ""))
	app := fiber.New()
	t.Cleanup(func() { _ = app.Shutdown() })

	var authErr error
	app.Get("/authz", func(c fiber.Ctx) error {
		authErr = storage.ExportHandlerAuthorizeProject(h, c, uuid.New(), uuid.New(), 0)
		return c.SendStatus(200)
	})
	req := httptest.NewRequest(fiber.MethodGet, "/authz", nil)
	_, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	require.Error(t, authErr)
	assert.Equal(t, fiber.StatusInternalServerError, authErr.(*fiber.Error).Code)
}

func TestHandler_GenerateUploadURL_Validation(t *testing.T) {
	h := storage.NewHandler(storage.ExportNewStorageServiceForTest(nil, ""))
	app := fiber.New()
	t.Cleanup(func() { _ = app.Shutdown() })
	app.Post("/upload-url", func(c fiber.Ctx) error {
		c.Locals("userID", uuid.New().String())
		return h.GenerateUploadURL(c)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/upload-url", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUserIDFromCtx_Missing(t *testing.T) {
	app := fiber.New()
	t.Cleanup(func() { _ = app.Shutdown() })

	var gotErr error
	app.Get("/uid", func(c fiber.Ctx) error {
		_, gotErr = storage.ExportUserIDFromCtx(c)
		return c.SendStatus(200)
	})
	req := httptest.NewRequest(fiber.MethodGet, "/uid", nil)
	_, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	require.Error(t, gotErr)
}
