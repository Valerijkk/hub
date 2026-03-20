package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/internal/platform/config"
)

const (
	uploadURLTTL   = 5 * time.Minute
	downloadURLTTL = 1 * time.Hour
)

// UploadResult contains the result of generating an upload URL.
type UploadResult struct {
	FileID    uuid.UUID `json:"fileId"`
	UploadURL string    `json:"uploadUrl"`
	S3Key     string    `json:"s3Key"`
}

// StorageService manages file uploads and downloads via S3-compatible storage (SeaweedFS).
type StorageService struct {
	s3Client      *s3.Client
	presignClient *s3.PresignClient
	entClient     *ent.Client
	bucket        string
}

// NewStorageService creates a storage service configured for SeaweedFS (S3-compatible).
func NewStorageService(cfg config.S3Config, entClient *ent.Client) (*StorageService, error) {
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
		o.UsePathStyle = true // Required for SeaweedFS
	})
	presignClient := s3.NewPresignClient(s3Client)

	return &StorageService{
		s3Client:      s3Client,
		presignClient: presignClient,
		entClient:     entClient,
		bucket:        cfg.Bucket,
	}, nil
}

// GenerateUploadURL creates a presigned PUT URL for file upload.
// It creates a FileMetadata record with status=pending and returns the upload URL.
func (s *StorageService) GenerateUploadURL(ctx context.Context, filename, contentType string, uploaderID, projectID uuid.UUID) (*UploadResult, error) {
	fileID := uuid.New()
	s3Key := fmt.Sprintf("%s/files/%s/%s", projectID.String(), fileID.String(), filename)

	// Create FileMetadata record (status=pending until confirmed)
	fileMeta, err := s.entClient.FileMetadata.Create().
		SetID(fileID).
		SetFilename(filename).
		SetContentType(contentType).
		SetSize(0). // Updated on confirm
		SetS3Key(s3Key).
		SetS3Bucket(s.bucket).
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(uploaderID).
		SetProjectID(projectID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create file metadata: %w", err)
	}

	// Generate presigned PUT URL
	presignResult, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s3Key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(uploadURLTTL))
	if err != nil {
		// Clean up the metadata record on failure
		_ = s.entClient.FileMetadata.DeleteOneID(fileMeta.ID).Exec(ctx)
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &UploadResult{
		FileID:    fileMeta.ID,
		UploadURL: presignResult.URL,
		S3Key:     s3Key,
	}, nil
}

// ConfirmUpload marks a file as confirmed and sets entity type and ID.
func (s *StorageService) ConfirmUpload(ctx context.Context, fileID uuid.UUID, entityType, entityID string) error {
	return s.entClient.FileMetadata.UpdateOneID(fileID).
		SetStatus(filemetadata.StatusConfirmed).
		SetEntityType(entityType).
		SetEntityID(entityID).
		Exec(ctx)
}

// GenerateDownloadURL creates a presigned GET URL for file download.
func (s *StorageService) GenerateDownloadURL(ctx context.Context, fileID uuid.UUID) (string, error) {
	fileMeta, err := s.entClient.FileMetadata.Get(ctx, fileID)
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

// DeleteFile performs a soft delete by setting status to deleted.
func (s *StorageService) DeleteFile(ctx context.Context, fileID uuid.UUID) error {
	return s.entClient.FileMetadata.UpdateOneID(fileID).
		SetStatus(filemetadata.StatusDeleted).
		Exec(ctx)
}
