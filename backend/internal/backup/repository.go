package backup

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/backuprecord"
)

// ErrNotFound — отсутствующая запись backup (W7-A4-011 → HTTP 404).
var ErrNotFound = errors.New("бэкап не найден")

// Repository — персистентность записей backup (ARCH-01).
type Repository interface {
	CreateInProgress(ctx context.Context, s3Key string) (*ent.BackupRecord, error)
	Get(ctx context.Context, id uuid.UUID) (*ent.BackupRecord, error)
	MarkCompleted(ctx context.Context, id uuid.UUID, fileSize int64) (*ent.BackupRecord, error)
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
	SetRestoredAt(ctx context.Context, id uuid.UUID, at time.Time) error
	ListAll(ctx context.Context) ([]*ent.BackupRecord, error)
	ListCompletedNewestFirst(ctx context.Context) ([]*ent.BackupRecord, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// EntRepository — реализация Repository через Ent ORM.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий модуля backup.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) CreateInProgress(ctx context.Context, s3Key string) (*ent.BackupRecord, error) {
	return r.client.BackupRecord.Create().
		SetS3Key(s3Key).
		SetStatus(backuprecord.StatusInProgress).
		Save(ctx)
}

func (r *EntRepository) Get(ctx context.Context, id uuid.UUID) (*ent.BackupRecord, error) {
	return r.client.BackupRecord.Get(ctx, id)
}

func (r *EntRepository) MarkCompleted(ctx context.Context, id uuid.UUID, fileSize int64) (*ent.BackupRecord, error) {
	return r.client.BackupRecord.UpdateOneID(id).
		SetStatus(backuprecord.StatusCompleted).
		SetFileSize(fileSize).
		Save(ctx)
}

func (r *EntRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.client.BackupRecord.UpdateOneID(id).
		SetStatus(backuprecord.StatusFailed).
		SetErrorMessage(errMsg).
		Save(ctx)
	return err
}

// SetRestoredAt фиксирует факт restore в restored_at (W7-A4-018), не в error_message.
func (r *EntRepository) SetRestoredAt(ctx context.Context, id uuid.UUID, at time.Time) error {
	_, err := r.client.BackupRecord.UpdateOneID(id).
		SetRestoredAt(at).
		Save(ctx)
	return err
}

func (r *EntRepository) ListAll(ctx context.Context) ([]*ent.BackupRecord, error) {
	return r.client.BackupRecord.Query().
		Order(ent.Desc(backuprecord.FieldCreatedAt)).
		All(ctx)
}

func (r *EntRepository) ListCompletedNewestFirst(ctx context.Context) ([]*ent.BackupRecord, error) {
	return r.client.BackupRecord.Query().
		Where(backuprecord.StatusEQ(backuprecord.StatusCompleted)).
		Order(ent.Desc(backuprecord.FieldCreatedAt)).
		All(ctx)
}

func (r *EntRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.client.BackupRecord.DeleteOneID(id).Exec(ctx)
}

// IsNotFound — ent not-found для маппинга на уровне service.
func IsNotFound(err error) bool {
	return ent.IsNotFound(err)
}

// NotFoundError — отсутствующая запись backup.
func NotFoundError() error {
	return ErrNotFound
}
