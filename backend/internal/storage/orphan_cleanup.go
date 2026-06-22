package storage

import (
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const orphanBatchLimit = 500

// OrphanCleanupService удаляет осиротевшие файлы — записи с status=pending,
// созданные более 24 часов назад (клиент не подтвердил загрузку).
type OrphanCleanupService struct {
	repo     Repository
	s3Client *s3.Client
	bucket   string
}

// NewOrphanCleanupService создаёт сервис очистки осиротевших файлов.
func NewOrphanCleanupService(repo Repository, s3Client *s3.Client, bucket string) *OrphanCleanupService {
	return &OrphanCleanupService{repo: repo, s3Client: s3Client, bucket: bucket}
}

// Cleanup находит и удаляет осиротевшие файлы: pending-записи старше 24 часов.
// Удаляет объект из S3 (SeaweedFS), затем запись из PostgreSQL.
func (s *OrphanCleanupService) Cleanup(ctx context.Context) error {
	cutoff := time.Now().Add(-24 * time.Hour)

	orphans, err := s.repo.ListOrphanPending(ctx, cutoff, orphanBatchLimit)
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
		if err := s.repo.DeleteByID(ctx, f.ID); err != nil {
			slog.Error("не удалось удалить запись метаданных",
				"file_id", f.ID, "error", err)
			continue
		}
		deleted++
	}

	slog.Info("осиротевшие файлы удалены", "deleted", deleted, "total", len(orphans))
	return nil
}
