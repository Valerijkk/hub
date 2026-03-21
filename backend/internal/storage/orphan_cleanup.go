package storage

import (
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/filemetadata"
)

// OrphanCleanupService удаляет осиротевшие файлы — записи с status=pending,
// созданные более 24 часов назад (клиент не подтвердил загрузку).
type OrphanCleanupService struct {
	db       *ent.Client
	s3Client *s3.Client
	bucket   string
}

// NewOrphanCleanupService создаёт сервис очистки осиротевших файлов.
func NewOrphanCleanupService(db *ent.Client, s3Client *s3.Client, bucket string) *OrphanCleanupService {
	return &OrphanCleanupService{db: db, s3Client: s3Client, bucket: bucket}
}

// Cleanup находит и удаляет осиротевшие файлы: pending-записи старше 24 часов.
// Удаляет объект из S3 (SeaweedFS), затем запись из PostgreSQL.
func (s *OrphanCleanupService) Cleanup(ctx context.Context) error {
	cutoff := time.Now().Add(-24 * time.Hour)

	orphans, err := s.db.FileMetadata.Query().
		Where(
			filemetadata.StatusEQ(filemetadata.StatusPending),
			filemetadata.CreatedAtLT(cutoff),
			filemetadata.DeletedAtIsNil(),
		).
		Limit(500). // Batch ограничение для предотвращения перегрузки
		All(ctx)
	if err != nil {
		return err
	}

	if len(orphans) == 0 {
		return nil
	}

	slog.Info("найдены осиротевшие файлы для удаления", "count", len(orphans))

	var deleted int
	for _, f := range orphans {
		// Удаляем объект из S3
		if s.s3Client != nil && f.S3Key != "" {
			_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(s.bucket),
				Key:    aws.String(f.S3Key),
			})
			if err != nil {
				slog.Error("не удалось удалить объект из S3",
					"s3_key", f.S3Key, "file_id", f.ID, "error", err)
				continue
			}
		}

		// Удаляем запись метаданных из PostgreSQL
		if err := s.db.FileMetadata.DeleteOneID(f.ID).Exec(ctx); err != nil {
			slog.Error("не удалось удалить запись метаданных",
				"file_id", f.ID, "error", err)
			continue
		}
		deleted++
	}

	slog.Info("осиротевшие файлы удалены", "deleted", deleted, "total", len(orphans))
	return nil
}
