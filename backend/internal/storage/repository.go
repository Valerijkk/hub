package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/ent/issue"
)

// ErrAlreadyConfirmed — A4-006: повторный confirm уже confirmed файла.
var ErrAlreadyConfirmed = errors.New("file already confirmed")

// ErrNotPending — confirm на не-pending статусе.
var ErrNotPending = errors.New("file is not pending")

// CreateFileInput — метаданные pending-загрузки до завершения S3 PUT.
type CreateFileInput struct {
	ID          uuid.UUID
	Filename    string
	ContentType string
	S3Key       string
	S3Bucket    string
	UploaderID  uuid.UUID
	ProjectID   uuid.UUID
	Size        int64
}

// Repository — персистентность метаданных файлов (ARCH-01).
type Repository interface {
	CreatePending(ctx context.Context, input CreateFileInput) (*ent.FileMetadata, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	ConfirmUpload(ctx context.Context, fileID uuid.UUID, entityType, entityID string) error
	ProjectIDForActiveFile(ctx context.Context, fileID uuid.UUID) (uuid.UUID, error)
	UploaderIDForActiveFile(ctx context.Context, fileID uuid.UUID) (uuid.UUID, error)
	GetActiveByID(ctx context.Context, fileID uuid.UUID) (*ent.FileMetadata, error)
	GetPendingByID(ctx context.Context, fileID uuid.UUID) (*ent.FileMetadata, error)
	IssueProjectID(ctx context.Context, issueID uuid.UUID) (uuid.UUID, error)
	// IsChannelMember — W7-A4-002: участник активного (не soft-deleted) канала.
	IsChannelMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
	// ActiveChannelProjectID — project_id активного канала; nil для DM/global.
	ActiveChannelProjectID(ctx context.Context, channelID uuid.UUID) (*uuid.UUID, error)
	SoftDelete(ctx context.Context, fileID uuid.UUID, at time.Time) error
	ListOrphanPending(ctx context.Context, olderThan time.Time, limit int) ([]*ent.FileMetadata, error)
}

// EntRepository — реализация Repository через Ent.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий метаданных файлов.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) CreatePending(ctx context.Context, input CreateFileInput) (*ent.FileMetadata, error) {
	return r.client.FileMetadata.Create().
		SetID(input.ID).
		SetFilename(input.Filename).
		SetContentType(input.ContentType).
		SetSize(input.Size).
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

// ConfirmUpload — A4-006: только status=pending и deleted_at IS NULL;
// повторный confirm → ErrAlreadyConfirmed (не перепривязывает entity).
func (r *EntRepository) ConfirmUpload(ctx context.Context, fileID uuid.UUID, entityType, entityID string) error {
	affected, err := r.client.FileMetadata.Update().
		Where(
			filemetadata.IDEQ(fileID),
			filemetadata.StatusEQ(filemetadata.StatusPending),
			filemetadata.DeletedAtIsNil(),
		).
		SetStatus(filemetadata.StatusConfirmed).
		SetEntityType(entityType).
		SetEntityID(entityID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("confirm upload: %w", err)
	}
	if affected > 0 {
		return nil
	}

	meta, gerr := r.client.FileMetadata.Query().
		Where(filemetadata.IDEQ(fileID), filemetadata.DeletedAtIsNil()).
		Only(ctx)
	if gerr != nil {
		if ent.IsNotFound(gerr) {
			return gerr
		}
		return fmt.Errorf("confirm upload lookup: %w", gerr)
	}
	if meta.Status == filemetadata.StatusConfirmed {
		return ErrAlreadyConfirmed
	}
	return ErrNotPending
}

func (r *EntRepository) ProjectIDForActiveFile(ctx context.Context, fileID uuid.UUID) (uuid.UUID, error) {
	return r.client.FileMetadata.Query().
		Where(filemetadata.IDEQ(fileID), filemetadata.DeletedAtIsNil()).
		QueryProject().
		OnlyID(ctx)
}

// UploaderIDForActiveFile — владелец pending/active файла (edge uploader, не exported FK).
func (r *EntRepository) UploaderIDForActiveFile(ctx context.Context, fileID uuid.UUID) (uuid.UUID, error) {
	return r.client.FileMetadata.Query().
		Where(filemetadata.IDEQ(fileID), filemetadata.DeletedAtIsNil()).
		QueryUploader().
		OnlyID(ctx)
}

// GetPendingByID — метаданные pending-файла (для HEAD/size check перед confirm).
func (r *EntRepository) GetPendingByID(ctx context.Context, fileID uuid.UUID) (*ent.FileMetadata, error) {
	return r.client.FileMetadata.Query().
		Where(
			filemetadata.IDEQ(fileID),
			filemetadata.DeletedAtIsNil(),
			filemetadata.StatusEQ(filemetadata.StatusPending),
		).
		Only(ctx)
}

// IssueProjectID — project_id задачи по UUID (W7-A4-002 entity scope).
func (r *EntRepository) IssueProjectID(ctx context.Context, issueID uuid.UUID) (uuid.UUID, error) {
	return r.client.Issue.Query().
		Where(issue.IDEQ(issueID), issue.DeletedAtIsNil()).
		QueryProject().
		OnlyID(ctx)
}

// IsChannelMember — тот же soft-delete predicate, что messenger.IsMember (NF-02).
func (r *EntRepository) IsChannelMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	return r.client.ChannelMember.Query().
		Where(
			channelmember.ChannelIDEQ(channelID),
			channelmember.UserIDEQ(userID),
			channelmember.HasChannelWith(channel.DeletedAtIsNil()),
		).
		Exist(ctx)
}

// ActiveChannelProjectID возвращает project_id живого канала или NotFound.
func (r *EntRepository) ActiveChannelProjectID(ctx context.Context, channelID uuid.UUID) (*uuid.UUID, error) {
	ch, err := r.client.Channel.Query().
		Where(channel.IDEQ(channelID), channel.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return ch.ProjectID, nil
}

// GetActiveByID — A4-007: download только confirmed + not deleted.
func (r *EntRepository) GetActiveByID(ctx context.Context, fileID uuid.UUID) (*ent.FileMetadata, error) {
	return r.client.FileMetadata.Query().
		Where(
			filemetadata.IDEQ(fileID),
			filemetadata.DeletedAtIsNil(),
			filemetadata.StatusEQ(filemetadata.StatusConfirmed),
		).
		Only(ctx)
}

// SoftDelete — P3-BE-05 / W7-A4-014: только active (DeletedAtIsNil); повторный delete → NotFound.
// Mutation: UpdateOneID без Where → повторный soft-delete «успешен».
func (r *EntRepository) SoftDelete(ctx context.Context, fileID uuid.UUID, at time.Time) error {
	affected, err := r.client.FileMetadata.Update().
		Where(
			filemetadata.IDEQ(fileID),
			filemetadata.DeletedAtIsNil(),
		).
		SetStatus(filemetadata.StatusDeleted).
		SetDeletedAt(at).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	return nil
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
