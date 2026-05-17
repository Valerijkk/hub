package audit

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/ent/systemsetting"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	// Ключи настроек периодов хранения в SystemSetting.
	RetentionChatMessagesDays   = "retention.chat_messages_days"
	RetentionAllureResultsDays  = "retention.allure_results_days"
	RetentionCallRecordingsDays = "retention.call_recordings_days"

	// Значения по умолчанию.
	defaultChatRetentionDays      = 730 // 2 года
	defaultAllureRetentionDays    = 365 // 1 год
	defaultCallRecordingRetention = 180 // 6 месяцев
)

// RetentionService выполняет очистку устаревших данных
// в соответствии с настраиваемыми периодами хранения.
//
// BE-H7: при удалении записей звонков сервис также удаляет соответствующие
// .mp4-объекты из S3 (через s3Client/s3Bucket). Раньше row в Postgres
// удалялся, а recording.mp4 оставался в SeaweedFS, что приводило к
// неограниченному росту storage и нарушало политику хранения.
type RetentionService struct {
	db       *ent.Client
	mongoDB  *mongo.Database
	s3Client *s3.Client
	s3Bucket string
}

// NewRetentionService создаёт новый сервис политик хранения данных.
// s3Client/s3Bucket могут быть nil/"" — тогда S3-объекты записей звонков
// не удаляются (логируется WARN). Это поведение fail-soft на случай
// development-окружения без SeaweedFS.
func NewRetentionService(db *ent.Client, mongoDB *mongo.Database, s3Client *s3.Client, s3Bucket string) *RetentionService {
	return &RetentionService{
		db:       db,
		mongoDB:  mongoDB,
		s3Client: s3Client,
		s3Bucket: s3Bucket,
	}
}

// EnsureIndexes создаёт индексы на created_at для коллекций, которые подметает
// retention-loop (messages, allure_results). Без такого индекса DeleteMany по
// $lt:cutoff превращается в COLLSCAN при росте коллекции — на сотни тысяч
// документов это уже видно по latency, на миллионах — фатально.
//
// ME-3: индексы sparse, чтобы не индексировать документы без поля created_at
// (формат может смещаться при миграциях). В MongoDB 4.2+ индексы создаются в
// фоне по умолчанию, явный Background не нужен.
//
// Идемпотентно — CreateOne не падает, если индекс с тем же именем уже есть.
func (s *RetentionService) EnsureIndexes(ctx context.Context) error {
	if s.mongoDB == nil {
		return nil
	}

	collections := []string{"messages", "allure_results"}
	for _, name := range collections {
		coll := s.mongoDB.Collection(name)
		indexName := "retention_created_at_idx"
		_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().
				SetName(indexName).
				SetSparse(true),
		})
		if err != nil {
			slog.Warn("retention: не удалось создать индекс на created_at",
				"collection", name, "index", indexName, "error", err)
			// fail-soft: retention продолжит работать, просто медленнее;
			// не возвращаем ошибку, чтобы не падал startup сервера.
			continue
		}
		slog.Info("retention: индекс на created_at гарантирован",
			"collection", name, "index", indexName)
	}
	return nil
}

// RunRetentionCleanup выполняет очистку всех устаревших данных.
// Вызывается ежедневно из cron-горутины.
func (s *RetentionService) RunRetentionCleanup(ctx context.Context) error {
	slog.Info("запуск ежедневной очистки по политикам хранения")

	chatDays := s.getRetentionDays(ctx, RetentionChatMessagesDays, defaultChatRetentionDays)
	allureDays := s.getRetentionDays(ctx, RetentionAllureResultsDays, defaultAllureRetentionDays)
	callDays := s.getRetentionDays(ctx, RetentionCallRecordingsDays, defaultCallRecordingRetention)

	var errs []error

	if err := s.purgeOldMessages(ctx, chatDays); err != nil {
		errs = append(errs, fmt.Errorf("очистка сообщений: %w", err))
	}
	if err := s.purgeOldAllureResults(ctx, allureDays); err != nil {
		errs = append(errs, fmt.Errorf("очистка Allure результатов: %w", err))
	}
	if err := s.purgeOldCallRecordings(ctx, callDays); err != nil {
		errs = append(errs, fmt.Errorf("очистка записей звонков: %w", err))
	}

	if len(errs) > 0 {
		for _, e := range errs {
			slog.Error("ошибка очистки по retention policy", "error", e)
		}
		return errs[0]
	}

	slog.Info("очистка по политикам хранения завершена")
	return nil
}

// getRetentionDays читает настраиваемый период хранения из SystemSetting.
func (s *RetentionService) getRetentionDays(ctx context.Context, key string, defaultVal int) int {
	setting, err := s.db.SystemSetting.Query().
		Where(systemsetting.KeyEQ(key)).
		Only(ctx)
	if err != nil {
		return defaultVal
	}
	val, err := strconv.Atoi(setting.Value)
	if err != nil || val <= 0 {
		return defaultVal
	}
	return val
}

// purgeOldMessages удаляет сообщения чата старше указанного количества дней из MongoDB.
func (s *RetentionService) purgeOldMessages(ctx context.Context, days int) error {
	if s.mongoDB == nil {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	coll := s.mongoDB.Collection("messages")

	result, err := coll.DeleteMany(ctx, bson.M{
		"created_at": bson.M{"$lt": cutoff},
	})
	if err != nil {
		return err
	}
	if result.DeletedCount > 0 {
		slog.Info("удалены устаревшие сообщения", "count", result.DeletedCount, "older_than_days", days)
	}
	return nil
}

// purgeOldAllureResults удаляет raw-результаты Allure старше указанного количества дней из MongoDB.
func (s *RetentionService) purgeOldAllureResults(ctx context.Context, days int) error {
	if s.mongoDB == nil {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	coll := s.mongoDB.Collection("allure_results")

	result, err := coll.DeleteMany(ctx, bson.M{
		"created_at": bson.M{"$lt": cutoff},
	})
	if err != nil {
		return err
	}
	if result.DeletedCount > 0 {
		slog.Info("удалены устаревшие Allure результаты", "count", result.DeletedCount, "older_than_days", days)
	}
	return nil
}

// purgeOldCallRecordings удаляет метаданные записей звонков старше указанного количества дней
// вместе с соответствующими .mp4-объектами в S3.
//
// BE-H7: сначала достаём список истёкших записей, удаляем каждый S3-объект,
// затем удаляем row из Postgres. Если S3 DeleteObject упадёт — логируем WARN
// и продолжаем (row всё равно удаляется): retention важнее консистентности
// S3, а orphan_cleanup умеет находить такие объекты. Если S3 не настроен
// (s3Client == nil) — ведём себя как раньше, только row-delete.
func (s *RetentionService) purgeOldCallRecordings(ctx context.Context, days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)

	// Если S3 не настроен — fallback на старое поведение (только метаданные).
	if s.s3Client == nil || s.s3Bucket == "" {
		n, err := s.db.Recording.Delete().
			Where(recording.CreatedAtLT(cutoff)).
			Exec(ctx)
		if err != nil {
			return err
		}
		if n > 0 {
			slog.Warn("удалены метаданные записей звонков без очистки S3 (S3 не настроен)",
				"count", n, "older_than_days", days)
		}
		return nil
	}

	// Получаем список истёкших записей, чтобы знать ключи S3-объектов.
	expired, err := s.db.Recording.Query().
		Where(recording.CreatedAtLT(cutoff)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("получение списка истёкших записей: %w", err)
	}
	if len(expired) == 0 {
		return nil
	}

	// Собираем ID именно из снапшота expired — DELETE WHERE id IN (ids)
	// гарантирует, что мы не уберём rows, вставленные между Query и Delete
	// с задним числом (created_at < cutoff). Раньше второй запрос был
	// DELETE WHERE created_at < cutoff и удалял такие "поздние" записи
	// без S3-очистки, порождая orphan-объекты в SeaweedFS.
	ids := make([]uuid.UUID, 0, len(expired))
	s3Deleted := 0
	for _, rec := range expired {
		ids = append(ids, rec.ID)
		if rec.FilePath == "" {
			continue
		}
		_, delErr := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.s3Bucket),
			Key:    aws.String(rec.FilePath),
		})
		if delErr != nil {
			// fail-soft: row всё равно удаляется ниже. orphan_cleanup
			// при следующем проходе подберёт битый объект.
			slog.Warn("не удалось удалить S3-объект записи (retention)",
				"recording_id", rec.ID, "s3_key", rec.FilePath, "error", delErr)
			continue
		}
		s3Deleted++
	}

	if len(ids) == 0 {
		return nil
	}

	n, err := s.db.Recording.Delete().
		Where(recording.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		slog.Info("удалены устаревшие записи звонков",
			"db_rows", n, "s3_objects", s3Deleted, "older_than_days", days)
	}
	return nil
}
