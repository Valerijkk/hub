// Package backup реализует резервное копирование и восстановление данных (PostgreSQL + MongoDB + SeaweedFS).
package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/backuprecord"
)

// BackupService реализует создание, восстановление и управление резервными копиями.
// Использует pg_dump + mongodump + S3 для полного бэкапа всех данных платформы.
const maxBackupReadBytes = 512 * 1024 * 1024 // 512 MiB — лимит чтения при восстановлении админом (NEW-03)

type BackupService struct {
	repo       Repository
	s3Client   *s3.Client
	s3Bucket   string
	pgURL      string
	mongoURL   string
	maxBackups int
	gate       *MaintenanceGate

	// lifecycleCtx — долгоживущий контекст (rootCtx из main), используемый
	// для фоновых goroutine'ов (например, ротация старых бэкапов в
	// cleanupOldBackups). НЕ используйте request-scoped ctx — он
	// отменяется по завершении HTTP-handler'а и может убить cleanup
	// в середине DELETE.
	lifecycleCtx context.Context
}

// NewBackupService создаёт новый сервис резервного копирования.
//
// rootCtx должен быть долгоживущим контекстом (обычно signal.NotifyContext
// из main) — он используется для фоновых задач сервиса (ротация старых
// бэкапов и т.п.). Передавайте именно rootCtx, а не request-ctx Fiber'а.
//
// gate — опциональный MaintenanceGate (A7-003); nil = только лог (тесты без HTTP).
func NewBackupService(rootCtx context.Context, repo Repository, s3Client *s3.Client, s3Bucket, pgURL, mongoURL string) *BackupService {
	return NewBackupServiceWithGate(rootCtx, repo, s3Client, s3Bucket, pgURL, mongoURL, nil)
}

// NewBackupServiceWithGate как NewBackupService, но с явным MaintenanceGate.
func NewBackupServiceWithGate(rootCtx context.Context, repo Repository, s3Client *s3.Client, s3Bucket, pgURL, mongoURL string, gate *MaintenanceGate) *BackupService {
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	return &BackupService{
		repo:         repo,
		s3Client:     s3Client,
		s3Bucket:     s3Bucket,
		pgURL:        pgURL,
		mongoURL:     mongoURL,
		maxBackups:   4,
		lifecycleCtx: rootCtx,
		gate:         gate,
	}
}

// PerformBackup выполняет полное резервное копирование:
// pg_dump (--format=custom) + mongodump + архивирование в ZIP + загрузка в S3.
// Создаёт BackupRecord со статусом completed/failed.
func (s *BackupService) PerformBackup(ctx context.Context) (*BackupResponse, error) {
	timestamp := time.Now().UTC().Format("20060102-150405")
	s3Key := fmt.Sprintf("backups/%s.zip", timestamp)

	// Создаём запись бэкапа (in_progress)
	record, err := s.repo.CreateInProgress(ctx, s3Key)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запись бэкапа: %w", err)
	}

	// Выполняем pg_dump
	pgData, err := s.dumpPostgres(ctx)
	if err != nil {
		s.markFailed(ctx, record.ID, err)
		// A7-004: не отдавать HTTP success при failed dump.
		return nil, fmt.Errorf("pg_dump: %w", err)
	}

	// Выполняем mongodump
	mongoData, err := s.dumpMongo(ctx)
	if err != nil {
		s.markFailed(ctx, record.ID, err)
		return nil, fmt.Errorf("mongodump: %w", err)
	}

	// Архивируем в ZIP
	zipData, err := s.createZIP(pgData, mongoData)
	if err != nil {
		s.markFailed(ctx, record.ID, err)
		return nil, fmt.Errorf("zip backup: %w", err)
	}

	// Загружаем в S3 (W7-A4-004 / P1-BE-39: без S3 — fail-closed, не MarkCompleted).
	if s.s3Client == nil {
		err := fmt.Errorf("s3 client not configured")
		s.markFailed(ctx, record.ID, err)
		return nil, err
	}
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.s3Bucket,
		Key:    &s3Key,
		Body:   bytes.NewReader(zipData),
	})
	if err != nil {
		s.markFailed(ctx, record.ID, fmt.Errorf("ошибка загрузки в S3: %w", err))
		return nil, fmt.Errorf("s3 upload: %w", err)
	}

	// Обновляем статус
	record, err = s.repo.MarkCompleted(ctx, record.ID, int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("не удалось обновить статус бэкапа: %w", err)
	}

	// Ротация старых бэкапов. Используем lifecycleCtx (rootCtx), чтобы
	// goroutine не была убита в середине DeleteObject/DeleteOneID при
	// завершении HTTP-handler'а, инициировавшего бэкап.
	cleanupCtx := s.lifecycleCtx
	if cleanupCtx == nil {
		cleanupCtx = context.Background()
	}
	go s.cleanupOldBackups(cleanupCtx)

	return toBackupResponse(record), nil
}

// RestoreBackup восстанавливает данные из бэкапа.
// ОПАСНАЯ операция: включает режим обслуживания, создаёт бэкап текущего
// состояния перед восстановлением, извлекает pg_dump + mongodump из ZIP,
// выполняет pg_restore и mongorestore через exec.Command.
func (s *BackupService) RestoreBackup(ctx context.Context, backupID uuid.UUID) error {
	record, err := s.repo.Get(ctx, backupID)
	if err != nil {
		if IsNotFound(err) {
			return NotFoundError()
		}
		return fmt.Errorf("ошибка получения бэкапа: %w", err)
	}

	if record.Status != backuprecord.StatusCompleted {
		return fmt.Errorf("невозможно восстановить бэкап со статусом %s", record.Status)
	}

	// Включаем режим обслуживания (устанавливаем флаг).
	s.setMaintenanceMode(true)
	defer s.setMaintenanceMode(false)

	// Создаём бэкап текущего состояния перед восстановлением (safety net).
	slog.Info("создание бэкапа текущего состояния перед восстановлением")
	if _, preBackupErr := s.PerformBackup(ctx); preBackupErr != nil {
		slog.Error("не удалось создать бэкап перед восстановлением", "error", preBackupErr)
		// A7-009: fail-closed — не разрушаем инстанс без safety-net.
		return fmt.Errorf("pre-restore backup failed: %w", preBackupErr)
	}

	// Скачиваем ZIP из S3.
	if s.s3Client == nil {
		return fmt.Errorf("S3 клиент не настроен, невозможно скачать бэкап")
	}

	result, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.s3Bucket,
		Key:    &record.S3Key,
	})
	if err != nil {
		return fmt.Errorf("ошибка скачивания бэкапа из S3: %w", err)
	}
	defer result.Body.Close()

	zipData, err := io.ReadAll(io.LimitReader(result.Body, maxBackupReadBytes))
	if err != nil {
		return fmt.Errorf("ошибка чтения бэкапа: %w", err)
	}

	slog.Info("бэкап скачан, начинаем извлечение", "s3_key", record.S3Key, "size", len(zipData))

	// Извлекаем pg_dump и mongodump из ZIP.
	pgData, mongoData, err := s.extractZIP(zipData)
	if err != nil {
		return fmt.Errorf("ошибка извлечения ZIP: %w", err)
	}

	// Восстанавливаем PostgreSQL.
	if len(pgData) > 0 {
		if err := s.restorePostgres(ctx, pgData); err != nil {
			return fmt.Errorf("ошибка pg_restore: %w", err)
		}
		slog.Info("PostgreSQL восстановлен из бэкапа")
	}

	// Восстанавливаем MongoDB.
	if len(mongoData) > 0 {
		if err := s.restoreMongo(ctx, mongoData); err != nil {
			return fmt.Errorf("ошибка mongorestore: %w", err)
		}
		slog.Info("MongoDB восстановлен из бэкапа")
	}

	// Обновляем статус записи бэкапа — фиксируем факт восстановления.
	_ = s.repo.SetRestoredNote(ctx, backupID, "restored at "+time.Now().UTC().Format(time.RFC3339))

	slog.Info("восстановление завершено", "s3_key", record.S3Key)
	return nil
}

// strPtr возвращает указатель на строку.
func strPtr(s string) *string { return &s }

// setMaintenanceMode устанавливает/снимает режим обслуживания (A7-003).
// Переключает MaintenanceGate (если задан) и логирует для наблюдаемости.
func (s *BackupService) setMaintenanceMode(enabled bool) {
	if s.gate != nil {
		s.gate.Set(enabled)
	}
	if enabled {
		slog.Warn("включён режим обслуживания для восстановления бэкапа")
	} else {
		slog.Info("режим обслуживания выключен")
	}
}

// IsMaintenanceMode сообщает, активен ли гейт обслуживания.
func (s *BackupService) IsMaintenanceMode() bool {
	if s == nil || s.gate == nil {
		return false
	}
	return s.gate.Enabled()
}

// extractZIP извлекает файлы postgres.dump и mongo.archive из ZIP-архива.
func (s *BackupService) extractZIP(zipData []byte) (pgData, mongoData []byte, err error) {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка открытия ZIP: %w", err)
	}

	for _, f := range reader.File {
		rc, openErr := f.Open()
		if openErr != nil {
			return nil, nil, fmt.Errorf("ошибка открытия файла %s в ZIP: %w", f.Name, openErr)
		}

		data, readErr := io.ReadAll(io.LimitReader(rc, maxBackupReadBytes))
		rc.Close()
		if readErr != nil {
			return nil, nil, fmt.Errorf("ошибка чтения файла %s из ZIP: %w", f.Name, readErr)
		}

		switch f.Name {
		case "postgres.dump":
			pgData = data
		case "mongo.archive":
			mongoData = data
		}
	}

	return pgData, mongoData, nil
}

// restorePostgres выполняет pg_restore --clean --if-exists из дампа.
func (s *BackupService) restorePostgres(ctx context.Context, data []byte) error {
	cmd := exec.CommandContext(ctx, "pg_restore",
		"--clean",
		"--if-exists",
		"--no-owner",
		"--dbname="+s.pgURL,
	)
	cmd.Stdin = bytes.NewReader(data)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// A7-007: не вшивать CombinedOutput (могут быть connection strings) в ошибку клиенту.
		slog.ErrorContext(ctx, "pg_restore failed", "error", err, "output_len", len(output))
		return fmt.Errorf("pg_restore failed")
	}
	return nil
}

// restoreMongo выполняет mongorestore --archive из дампа.
func (s *BackupService) restoreMongo(ctx context.Context, data []byte) error {
	cmd := exec.CommandContext(ctx, "mongorestore",
		"--uri="+s.mongoURL,
		"--archive",
		"--drop",
	)
	cmd.Stdin = bytes.NewReader(data)

	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.ErrorContext(ctx, "mongorestore failed", "error", err, "output_len", len(output))
		return fmt.Errorf("mongorestore failed")
	}
	return nil
}

// ListBackups возвращает список всех бэкапов, отсортированных по дате.
func (s *BackupService) ListBackups(ctx context.Context) ([]BackupResponse, error) {
	records, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список бэкапов: %w", err)
	}

	result := make([]BackupResponse, 0, len(records))
	for _, r := range records {
		result = append(result, *toBackupResponse(r))
	}
	return result, nil
}

// DeleteBackup удаляет запись бэкапа и файл из S3.
func (s *BackupService) DeleteBackup(ctx context.Context, backupID uuid.UUID) error {
	record, err := s.repo.Get(ctx, backupID)
	if err != nil {
		if IsNotFound(err) {
			return NotFoundError()
		}
		return err
	}

	// Удаляем из S3
	if s.s3Client != nil {
		_, _ = s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &s.s3Bucket,
			Key:    &record.S3Key,
		})
	}

	return s.repo.Delete(ctx, backupID)
}

// dumpPostgres выполняет pg_dump --format=custom.
func (s *BackupService) dumpPostgres(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "pg_dump", "--format=custom", s.pgURL)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ошибка pg_dump: %w", err)
	}
	return out, nil
}

// dumpMongo выполняет mongodump --archive.
func (s *BackupService) dumpMongo(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "mongodump", "--uri="+s.mongoURL, "--archive")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ошибка mongodump: %w", err)
	}
	return out, nil
}

// createZIP архивирует данные pg_dump и mongodump в ZIP.
func (s *BackupService) createZIP(pgData, mongoData []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	pgFile, err := w.Create("postgres.dump")
	if err != nil {
		return nil, err
	}
	if _, err := pgFile.Write(pgData); err != nil {
		return nil, err
	}

	mongoFile, err := w.Create("mongo.archive")
	if err != nil {
		return nil, err
	}
	if _, err := mongoFile.Write(mongoData); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// cleanupOldBackups удаляет старые бэкапы, оставляя не более maxBackups.
func (s *BackupService) cleanupOldBackups(ctx context.Context) {
	records, err := s.repo.ListCompletedNewestFirst(ctx)
	if err != nil {
		slog.Error("ошибка получения бэкапов для ротации", "error", err)
		return
	}

	if len(records) <= s.maxBackups {
		return
	}

	for _, r := range records[s.maxBackups:] {
		if s.s3Client != nil {
			_, _ = s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: &s.s3Bucket,
				Key:    &r.S3Key,
			})
		}
		if err := s.repo.Delete(ctx, r.ID); err != nil {
			slog.Error("ошибка удаления старого бэкапа", "id", r.ID, "error", err)
		}
	}

	slog.Info("ротация бэкапов завершена", "удалено", len(records)-s.maxBackups)
}

// markFailed обновляет статус бэкапа на failed.
func (s *BackupService) markFailed(ctx context.Context, id uuid.UUID, err error) {
	if updateErr := s.repo.MarkFailed(ctx, id, err.Error()); updateErr != nil {
		slog.Error("не удалось обновить статус бэкапа на failed", "error", updateErr)
	}
}

// toBackupResponse конвертирует Ent-сущность в DTO ответа.
func toBackupResponse(r *ent.BackupRecord) *BackupResponse {
	errMsg := ""
	if r.ErrorMessage != nil {
		errMsg = *r.ErrorMessage
	}
	return &BackupResponse{
		ID:           r.ID.String(),
		S3Key:        r.S3Key,
		FileSize:     r.FileSize,
		Status:       string(r.Status),
		ErrorMessage: errMsg,
		CreatedAt:    r.CreatedAt,
	}
}
