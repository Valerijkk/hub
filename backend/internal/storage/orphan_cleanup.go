package storage

import (
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// orphanBatchLimit — размер батча ListOrphanPending.
// Var (не const): hubtest может снизить для проверки drain-цикла (P3-BE-15 / W7-A4-015).
var orphanBatchLimit = 500

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
// P3-BE-15 / W7-A4-015: цикл батчей до исчерпания (>500 за один прогон).
func (s *OrphanCleanupService) Cleanup(ctx context.Context) error {
	cutoff := time.Now().Add(-24 * time.Hour)
	var totalDeleted int

	for {
		orphans, err := s.repo.ListOrphanPending(ctx, cutoff, orphanBatchLimit)
		if err != nil {
			return err
		}
		if len(orphans) == 0 {
			break
		}

		slog.Info("найдены осиротевшие файлы для удаления", "count", len(orphans))

		var deleted int
		for _, f := range orphans {
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

			if err := s.repo.DeleteByID(ctx, f.ID); err != nil {
				slog.Error("не удалось удалить запись метаданных",
					"file_id", f.ID, "error", err)
				continue
			}
			deleted++
		}

		totalDeleted += deleted
		slog.Info("осиротевшие файлы удалены (батч)", "deleted", deleted, "batch", len(orphans))

		// Если батч неполный — хвоста больше нет.
		if len(orphans) < orphanBatchLimit {
			break
		}
		// Защита от бесконечного цикла при ошибках delete (записи остаются pending).
		if deleted == 0 {
			slog.Error("orphan cleanup: батч не уменьшился, останавливаемся",
				"remaining", len(orphans))
			break
		}
	}

	if totalDeleted > 0 {
		slog.Info("осиротевшие файлы удалены", "deleted", totalDeleted)
	}
	return nil
}
