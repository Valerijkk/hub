//go:build hubtest

package storage

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/rbac"
)

// ExportSanitizeFilename нормализует имя файла для hubtest.
func ExportSanitizeFilename(name string) string {
	return sanitizeFilename(name)
}

// ExportNewStorageServiceForTest создаёт StorageService для hubtest без S3.
func ExportNewStorageServiceForTest(repo Repository, bucket string) *StorageService {
	return &StorageService{repo: repo, bucket: bucket}
}

// ExportNewStorageServiceForTestWithS3 создаёт StorageService с S3-клиентами для hubtest.
func ExportNewStorageServiceForTestWithS3(repo Repository, bucket string, s3Client *s3.Client, presign *s3.PresignClient) *StorageService {
	return &StorageService{
		repo:          repo,
		bucket:        bucket,
		s3Client:      s3Client,
		presignClient: presign,
	}
}

// ExportUploadURLTTL возвращает TTL presigned upload URL для hubtest.
func ExportUploadURLTTL() time.Duration {
	return uploadURLTTL
}

// ExportDownloadURLTTL возвращает TTL presigned download URL для hubtest.
func ExportDownloadURLTTL() time.Duration {
	return downloadURLTTL
}

// ExportUserIDFromCtx извлекает userID из Fiber Locals для hubtest.
func ExportUserIDFromCtx(c fiber.Ctx) (uuid.UUID, error) {
	return userIDFromCtx(c)
}

// ExportHandlerAuthorizeProject вызывает authorizeProject для hubtest.
func ExportHandlerAuthorizeProject(h *Handler, c fiber.Ctx, userID, projectID uuid.UUID, perm rbac.Permission) error {
	return h.authorizeProject(c, userID, projectID, perm)
}

// ExportMaxUploadBytes возвращает лимит размера файла (NF-09) для hubtest.
func ExportMaxUploadBytes() int64 {
	return maxUploadBytes
}

// ExportValidateEntityScope вызывает validateEntityScope для hubtest.
func ExportValidateEntityScope(s *StorageService, ctx context.Context, fileProjectID uuid.UUID, entityType, entityID string) error {
	return s.validateEntityScope(ctx, fileProjectID, entityType, entityID)
}

// ExportValidateUploadContentType — W7-A4-006 MIME guard для hubtest.
func ExportValidateUploadContentType(contentType string) error {
	return validateUploadContentType(contentType)
}

// ExportIsS3NotFound — W7-A4-005 NotFound detect для hubtest.
func ExportIsS3NotFound(err error) bool {
	return isS3NotFound(err)
}
