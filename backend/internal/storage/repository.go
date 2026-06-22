package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/filemetadata"
)

// CreateFileInput — метаданные pending-загрузки до завершения S3 PUT.
type CreateFileInput struct {
	ID          uuid.UUID
	Filename    string
	ContentType string
	S3Key       string
	S3Bucket    string
	UploaderID  uuid.UUID
	ProjectID   uuid.UUID
}

// Repository — персистентность метаданных файлов (ARCH-01).
type Repository interface {
	CreatePending(ctx context.Context, input CreateFileInput) (*ent.FileMetadata, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	ConfirmUpload(ctx context.Context, fileID uuid.UUID, entityType, entityID string) error
	ProjectIDForActiveFile(ctx context.Context, fileID uuid.UUID) (uuid.UUID, error)
	GetActiveByID(ctx context.Context, fileID uuid.UUID) (*ent.FileMetadata, error)
	SoftDelete(ctx context.Context, fileID uuid.UUID, at time.Time) error
	ListOrphanPending(ctx context.Context, olderThan time.Time, limit int) ([]*ent.FileMetadata, error)
}

// EntRepository — реализация Repository через Ent ORM.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий модуля storage.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) CreatePending(ctx context.Context, input CreateFileInput) (*ent.FileMetadata, error) {
	return r.client.FileMetadata.Create().
		SetID(input.ID).
		SetFilename(input.Filename).
		SetContentType(input.ContentType).
		SetSize(0).
		SetS3Key(input.S3Key).
		SetS3Bucket(input.S3Bucket).
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(input.UploaderID).
		SetProjectID(input.ProjectID).
		Save(ctx)
}

func (r *EntRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.client.FileMetadata.DeleteOneID(id).Exec(ctx)
}

func (r *EntRepository) ConfirmUpload(ctx context.Context, fileID uuid.UUID, entityType, entityID string) error {
	return r.client.FileMetadata.UpdateOneID(fileID).
		SetStatus(filemetadata.StatusConfirmed).
		SetEntityType(entityType).
		SetEntityID(entityID).
		Exec(ctx)
}

func (r *EntRepository) ProjectIDForActiveFile(ctx context.Context, fileID uuid.UUID) (uuid.UUID, error) {
	return r.client.FileMetadata.Query().
		Where(filemetadata.IDEQ(fileID), filemetadata.DeletedAtIsNil()).
		QueryProject().
		OnlyID(ctx)
}

func (r *EntRepository) GetActiveByID(ctx context.Context, fileID uuid.UUID) (*ent.FileMetadata, error) {
	return r.client.FileMetadata.Query().
		Where(filemetadata.IDEQ(fileID), filemetadata.DeletedAtIsNil()).
		Only(ctx)
}

func (r *EntRepository) SoftDelete(ctx context.Context, fileID uuid.UUID, at time.Time) error {
	return r.client.FileMetadata.UpdateOneID(fileID).
		SetStatus(filemetadata.StatusDeleted).
		SetDeletedAt(at).
		Exec(ctx)
}

func (r *EntRepository) ListOrphanPending(ctx context.Context, olderThan time.Time, limit int) ([]*ent.FileMetadata, error) {
	return r.client.FileMetadata.Query().
		Where(
			filemetadata.StatusEQ(filemetadata.StatusPending),
			filemetadata.CreatedAtLT(olderThan),
			filemetadata.DeletedAtIsNil(),
		).
		Limit(limit).
		All(ctx)
}
