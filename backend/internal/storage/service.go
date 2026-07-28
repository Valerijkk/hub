// Package storage реализует файловое хранилище через S3-совместимый бэкенд (SeaweedFS).
package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// Настройки файлового хранилища.
const (
	uploadURLTTL   = 5 * time.Minute
	downloadURLTTL = 1 * time.Hour
	// maxUploadBytes — TZ NF-09: Fiber BodyLimit 25 MB; тот же потолок для presigned PUT.
	maxUploadBytes int64 = 25 * 1024 * 1024
)

var (
	// ErrFileTooLarge — размер загрузки превышает maxUploadBytes.
	ErrFileTooLarge = errors.New("file too large")
	// ErrInvalidFileSize — size <= 0 при явном указании.
	ErrInvalidFileSize = errors.New("invalid file size")
	// ErrInvalidEntityType — entityType вне allowlist.
	ErrInvalidEntityType = errors.New("invalid entity type")
	// ErrInvalidEntity — entityId невалиден / сущность не найдена.
	ErrInvalidEntity = errors.New("invalid entity")
	// ErrEntityProjectMismatch — сущность принадлежит другому проекту, чем файл.
	ErrEntityProjectMismatch = errors.New("entity project mismatch")
	// ErrObjectNotFound — Confirm без объекта в S3 (W7-A4-005 / P2-BE-21).
	ErrObjectNotFound = errors.New("object not found in storage")
	// ErrContentTypeNotAllowed — опасный MIME (W7-A4-006 / P2-BE-21).
	ErrContentTypeNotAllowed = errors.New("content type not allowed")
)

// UploadResult содержит результат генерации presigned URL для загрузки файла.
type UploadResult struct {
	FileID    uuid.UUID `json:"fileId"`
	UploadURL string    `json:"uploadUrl"`
	S3Key     string    `json:"s3Key"`
}

// StorageService управляет загрузкой и скачиванием файлов через S3-совместимое хранилище (SeaweedFS).
type StorageService struct {
	s3Client      *s3.Client
	presignClient *s3.PresignClient
	repo          Repository
	bucket        string
}

// NewStorageService создаёт сервис хранилища, настроенный для SeaweedFS (S3-совместимый).
func NewStorageService(cfg config.S3Config, repo Repository) (*StorageService, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true // требуется для SeaweedFS
	})
	presignClient := s3.NewPresignClient(s3Client)

	return &StorageService{
		s3Client:      s3Client,
		presignClient: presignClient,
		repo:          repo,
		bucket:        cfg.Bucket,
	}, nil
}

// GenerateUploadURL создаёт presigned PUT URL для загрузки файла.
// size > 0 — ContentLength в presign (NF-09); size == 0 — legacy без bind (клиент без size).
// size > maxUploadBytes → ErrFileTooLarge.
func (s *StorageService) GenerateUploadURL(ctx context.Context, filename, contentType string, uploaderID, projectID uuid.UUID, size int64) (*UploadResult, error) {
	if size < 0 {
		return nil, ErrInvalidFileSize
	}
	if size > maxUploadBytes {
		return nil, ErrFileTooLarge
	}

	fileID := uuid.New()
	s3Key := fmt.Sprintf("%s/files/%s/%s", projectID.String(), fileID.String(), sanitizeFilename(filename))

	if err := validateUploadContentType(contentType); err != nil {
		return nil, err
	}

	fileMeta, err := s.repo.CreatePending(ctx, CreateFileInput{
		ID:          fileID,
		Filename:    filename,
		ContentType: contentType,
		S3Key:       s3Key,
		S3Bucket:    s.bucket,
		UploaderID:  uploaderID,
		ProjectID:   projectID,
		Size:        size,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create file metadata: %w", err)
	}

	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s3Key),
		ContentType: aws.String(contentType),
	}
	// W7-A4-001: при известном size подписываем ContentLength — oversized PUT отклонит S3.
	if size > 0 {
		putInput.ContentLength = aws.Int64(size)
	}

	presignResult, err := s.presignClient.PresignPutObject(ctx, putInput, s3.WithPresignExpires(uploadURLTTL))
	if err != nil {
		if delErr := s.repo.DeleteByID(ctx, fileMeta.ID); delErr != nil {
			// W7-A4-019 / P3-BE-01: cleanup fail после ошибка presign.
			slog.WarnContext(ctx, "storage: cleanup pending file after presign fail",
				"file_id", fileMeta.ID, "error", delErr)
		}
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &UploadResult{
		FileID:    fileMeta.ID,
		UploadURL: presignResult.URL,
		S3Key:     s3Key,
	}, nil
}

// ConfirmUpload помечает файл как подтверждённый и устанавливает тип и ID сущности.
// A4-006: только pending → confirmed; ErrAlreadyConfirmed при повторном confirm.
// W7-A4-001/002: size через HEAD (если S3) + entity scope к project файла.
func (s *StorageService) ConfirmUpload(ctx context.Context, fileID uuid.UUID, entityType, entityID string) error {
	fileMeta, err := s.repo.GetPendingByID(ctx, fileID)
	if err != nil {
		if ent.IsNotFound(err) {
			// A4-006 path: уже confirmed / deleted — делегируем repo.ConfirmUpload.
			return s.repo.ConfirmUpload(ctx, fileID, entityType, entityID)
		}
		return err
	}

	if s.s3Client != nil {
		if err := s.assertObjectWithinLimit(ctx, fileMeta); err != nil {
			return err
		}
	}

	projectID, err := s.repo.ProjectIDForActiveFile(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file project: %w", err)
	}
	uploaderID, err := s.repo.UploaderIDForActiveFile(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file uploader: %w", err)
	}
	if err := s.validateEntityScope(ctx, projectID, uploaderID, entityType, entityID); err != nil {
		return err
	}

	if err := s.repo.ConfirmUpload(ctx, fileID, entityType, entityID); err != nil {
		return err
	}
	return nil
}

func (s *StorageService) assertObjectWithinLimit(ctx context.Context, fileMeta *ent.FileMetadata) error {
	out, err := s.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(fileMeta.S3Bucket),
		Key:    aws.String(fileMeta.S3Key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return ErrObjectNotFound
		}
		return fmt.Errorf("head object: %w", err)
	}
	if out.ContentLength != nil && *out.ContentLength > maxUploadBytes {
		if _, delErr := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(fileMeta.S3Bucket),
			Key:    aws.String(fileMeta.S3Key),
		}); delErr != nil {
			slog.WarnContext(ctx, "storage: delete oversized s3 object failed",
				"file_id", fileMeta.ID, "error", delErr)
		}
		if delErr := s.repo.DeleteByID(ctx, fileMeta.ID); delErr != nil {
			slog.WarnContext(ctx, "storage: delete oversized file meta failed",
				"file_id", fileMeta.ID, "error", delErr)
		}
		return ErrFileTooLarge
	}
	return nil
}

// validateUploadContentType — W7-A4-006: блокируем HTML/SVG (XSS via download).
func validateUploadContentType(contentType string) error {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch ct {
	case "text/html", "application/xhtml+xml", "image/svg+xml", "text/javascript", "application/javascript":
		return ErrContentTypeNotAllowed
	}
	if ct == "" {
		return ErrContentTypeNotAllowed
	}
	return nil
}

func isS3NotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "notfound") ||
		strings.Contains(msg, "not found") ||
		strings.Contains(msg, "nosuchkey") ||
		strings.Contains(msg, "statuscode: 404") ||
		strings.Contains(msg, "status code: 404")
}

// validateEntityScope — W7-A4-002: allowlist + ownership/membership per type.
// Mutation: вернуть только entityID!="" для message/avatar/voice/channel → denial RED.
func (s *StorageService) validateEntityScope(ctx context.Context, fileProjectID, uploaderID uuid.UUID, entityType, entityID string) error {
	switch entityType {
	case "issue":
		entityUUID, err := uuid.Parse(entityID)
		if err != nil {
			return ErrInvalidEntity
		}
		issueProjectID, err := s.repo.IssueProjectID(ctx, entityUUID)
		if err != nil {
			if ent.IsNotFound(err) {
				return ErrInvalidEntity
			}
			return fmt.Errorf("issue project: %w", err)
		}
		if issueProjectID != fileProjectID {
			return ErrEntityProjectMismatch
		}
		return nil
	case "avatar":
		// Avatar — только свой: "me" (FE) или UUID uploader'а.
		if entityID == "" {
			return ErrInvalidEntity
		}
		if entityID == "me" || entityID == uploaderID.String() {
			return nil
		}
		return ErrInvalidEntity
	case "channel", "message", "voice":
		// entityID = channel UUID; uploader обязан быть членом активного канала.
		// voice "temp" — отложенная привязка до send (FE voice-recorder).
		if entityID == "" {
			return ErrInvalidEntity
		}
		if entityType == "voice" && entityID == "temp" {
			return nil
		}
		return s.validateChannelMembership(ctx, fileProjectID, uploaderID, entityID)
	default:
		if entityType == "" && entityID == "" {
			return nil
		}
		return ErrInvalidEntityType
	}
}

// validateChannelMembership — channel/message/voice: UUID + membership + project match.
func (s *StorageService) validateChannelMembership(ctx context.Context, fileProjectID, uploaderID uuid.UUID, entityID string) error {
	channelID, err := uuid.Parse(entityID)
	if err != nil {
		return ErrInvalidEntity
	}
	chProjectID, err := s.repo.ActiveChannelProjectID(ctx, channelID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrInvalidEntity
		}
		return fmt.Errorf("channel project: %w", err)
	}
	if chProjectID != nil && *chProjectID != fileProjectID {
		return ErrEntityProjectMismatch
	}
	ok, err := s.repo.IsChannelMember(ctx, channelID, uploaderID)
	if err != nil {
		return fmt.Errorf("channel membership: %w", err)
	}
	if !ok {
		return ErrInvalidEntity
	}
	return nil
}

// FileProjectID возвращает ID проекта, которому принадлежит файл.
// Используется обработчиком для проверки RBAC по реальному владельцу файла
// (а не по projectId из URL/body, который злоумышленник мог подделать).
// Возвращает ошибку, если файл удалён или не найден.
func (s *StorageService) FileProjectID(ctx context.Context, fileID uuid.UUID) (uuid.UUID, error) {
	pid, err := s.repo.ProjectIDForActiveFile(ctx, fileID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("file not found: %w", err)
	}
	return pid, nil
}

// GenerateDownloadURL создаёт presigned GET URL для скачивания файла.
func (s *StorageService) GenerateDownloadURL(ctx context.Context, fileID uuid.UUID) (string, error) {
	fileMeta, err := s.repo.GetActiveByID(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	presignResult, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(fileMeta.S3Bucket),
		Key:    aws.String(fileMeta.S3Key),
	}, s3.WithPresignExpires(downloadURLTTL))
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return presignResult.URL, nil
}

// NewOrphanCleanupFromService создаёт OrphanCleanupService из S3-клиента и bucket текущего StorageService.
func (s *StorageService) NewOrphanCleanupFromService() *OrphanCleanupService {
	return NewOrphanCleanupService(s.repo, s.s3Client, s.bucket)
}

// S3Client возвращает низкоуровневый S3-клиент. Используется другими модулями
// (например, backup), которым нужен прямой доступ к S3-совместимому хранилищу
// без необходимости дублировать конфигурацию credentials/endpoint.
func (s *StorageService) S3Client() *s3.Client {
	return s.s3Client
}

// Bucket возвращает имя bucket'а, используемого текущим StorageService.
func (s *StorageService) Bucket() string {
	return s.bucket
}

// DeleteFile выполняет мягкое удаление файла (устанавливает статус deleted и deleted_at).
func (s *StorageService) DeleteFile(ctx context.Context, fileID uuid.UUID) error {
	return s.repo.SoftDelete(ctx, fileID, time.Now().UTC())
}
