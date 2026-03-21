package audit

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/ent/systemsetting"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
type RetentionService struct {
	db      *ent.Client
	mongoDB *mongo.Database
}

// NewRetentionService создаёт новый сервис политик хранения данных.
func NewRetentionService(db *ent.Client, mongoDB *mongo.Database) *RetentionService {
	return &RetentionService{db: db, mongoDB: mongoDB}
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

// purgeOldCallRecordings удаляет метаданные записей звонков старше указанного количества дней.
func (s *RetentionService) purgeOldCallRecordings(ctx context.Context, days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)

	n, err := s.db.Recording.Delete().
		Where(recording.CreatedAtLT(cutoff)).
		Exec(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		slog.Info("удалены устаревшие записи звонков", "count", n, "older_than_days", days)
	}
	return nil
}
