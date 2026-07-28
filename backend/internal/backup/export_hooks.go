//go:build hubtest

package backup

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// ExportToBackupResponse конвертирует ent.BackupRecord для hubtest.
func ExportToBackupResponse(r *ent.BackupRecord) *BackupResponse {
	return toBackupResponse(r)
}

// ExportStrPtr возвращает указатель на строку для hubtest.
func ExportStrPtr(s string) *string {
	return strPtr(s)
}

// ExportBackupExtractZIP распаковывает ZIP backup для hubtest.
func ExportBackupExtractZIP(s *BackupService, zipData []byte) ([]byte, []byte, error) {
	return s.extractZIP(zipData)
}

// ExportBackupMarkFailed помечает backup как failed для hubtest.
func ExportBackupMarkFailed(s *BackupService, ctx context.Context, id uuid.UUID, err error) {
	s.markFailed(ctx, id, err)
}

// ExportSanitizeBackupErrorMessage — W7-A4-007 sanitize для hubtest.
func ExportSanitizeBackupErrorMessage(err error) string {
	return sanitizeBackupErrorMessage(err)
}

// ExportPgDumpArgsEnv — W7-A4-013 argv без пароля для hubtest.
func ExportPgDumpArgsEnv(pgURL string, extraArgs ...string) (args []string, env []string, err error) {
	return pgDumpArgsEnv(pgURL, extraArgs...)
}

// ExportWriteMongoConfig — W7-A4-013 mongo --config для hubtest.
func ExportWriteMongoConfig(dir, mongoURL string) (string, func(), error) {
	return writeMongoConfig(dir, mongoURL)
}

// ExportMongoConfigArg wraps mongoConfigArg for hubtest.
func ExportMongoConfigArg(configPath string) string {
	return mongoConfigArg(configPath)
}

// ExportBackupLifecycleCtx возвращает lifecycle context сервиса для hubtest.
func ExportBackupLifecycleCtx(s *BackupService) context.Context {
	return s.lifecycleCtx
}

// ExportNewBackupServiceForURLTest создаёт BackupService с URL для hubtest.
func ExportNewBackupServiceForURLTest(pgURL, mongoURL string) *BackupService {
	return &BackupService{pgURL: pgURL, mongoURL: mongoURL}
}

// ExportBackupSetURLs задаёт URL подключений для hubtest.
func ExportBackupSetURLs(s *BackupService, pgURL, mongoURL string) {
	s.pgURL = pgURL
	s.mongoURL = mongoURL
}

// ExportBackupDumpMongo вызывает dumpMongo для hubtest.
func ExportBackupDumpMongo(s *BackupService, ctx context.Context) ([]byte, error) {
	return s.dumpMongo(ctx)
}

// ExportBackupDumpPostgres вызывает dumpPostgres для hubtest.
func ExportBackupDumpPostgres(s *BackupService, ctx context.Context) ([]byte, error) {
	return s.dumpPostgres(ctx)
}

// ExportBackupRestorePostgres вызывает restorePostgres для hubtest.
func ExportBackupRestorePostgres(s *BackupService, ctx context.Context, data []byte) error {
	return s.restorePostgres(ctx, data)
}

// ExportBackupRestoreMongo вызывает restoreMongo для hubtest.
func ExportBackupRestoreMongo(s *BackupService, ctx context.Context, data []byte) error {
	return s.restoreMongo(ctx, data)
}

// ExportBackupCreateZIP создаёт ZIP архив backup для hubtest.
func ExportBackupCreateZIP(s *BackupService, pgData, mongoData []byte) ([]byte, error) {
	return s.createZIP(pgData, mongoData)
}

// ExportBackupIsNotFound проверяет ent not-found для hubtest.
func ExportBackupIsNotFound(err error) bool {
	return IsNotFound(err)
}

// ExportBackupSetMaxBackups задаёт лимит хранимых backup для hubtest.
func ExportBackupSetMaxBackups(s *BackupService, n int) {
	s.maxBackups = n
}

// ExportBackupCleanupOldBackups вызывает cleanupOldBackups для hubtest.
func ExportBackupCleanupOldBackups(s *BackupService, ctx context.Context) {
	s.cleanupOldBackups(ctx)
}

// ExportBackupSetMaintenanceMode переключает maintenance mode для hubtest.
func ExportBackupSetMaintenanceMode(s *BackupService, enabled bool) {
	s.setMaintenanceMode(enabled)
}

// ExportBackupIsMaintenanceMode возвращает состояние gate для hubtest (A7-003).
func ExportBackupIsMaintenanceMode(s *BackupService) bool {
	return s.IsMaintenanceMode()
}

// ExportNewBackupServiceWithGate создаёт BackupService с MaintenanceGate для hubtest.
func ExportNewBackupServiceWithGate(rootCtx context.Context, repo Repository, gate *MaintenanceGate) *BackupService {
	return NewBackupServiceWithGate(rootCtx, repo, nil, "", "", "", gate)
}

// ExportBackupSetS3 задаёт S3 клиент и bucket для hubtest.
func ExportBackupSetS3(s *BackupService, bucket string, client *s3.Client) {
	s.s3Bucket = bucket
	s.s3Client = client
}

// ExportBackupS3Bucket возвращает S3 bucket сервиса для hubtest.
func ExportBackupS3Bucket(s *BackupService) string {
	return s.s3Bucket
}
