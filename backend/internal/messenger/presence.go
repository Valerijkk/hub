package messenger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Ключи Redis для хранения данных о присутствии пользователей.
const (
	// presenceOnlineKey — sorted set: member=userID, score=unix timestamp последнего heartbeat.
	presenceOnlineKey = "presence:online"
	// presenceStatusKeyPrefix — hash-ключ для кастомного статуса пользователя.
	presenceStatusKeyPrefix = "presence:status:"
	// presenceAwayThreshold — порог перехода в статус "away" (5 минут).
	presenceAwayThreshold = 5 * time.Minute
	// presenceStaleThreshold — порог удаления устаревших записей (90 секунд).
	presenceStaleThreshold = 90 * time.Second
	// presenceCleanupInterval — интервал между запусками очистки (60 секунд).
	presenceCleanupInterval = 60 * time.Second
)

// PresenceService отслеживает присутствие пользователей (online/offline/away)
// через Redis sorted set. Score каждого элемента — unix timestamp последнего heartbeat.
type PresenceService struct {
	rdb *redis.Client
}

// NewPresenceService создаёт новый сервис присутствия.
func NewPresenceService(rdb *redis.Client) *PresenceService {
	return &PresenceService{rdb: rdb}
}

// Heartbeat обновляет время последней активности пользователя.
// Вызывается при каждом ping от клиента.
func (s *PresenceService) Heartbeat(ctx context.Context, userID string) error {
	return s.rdb.ZAdd(ctx, presenceOnlineKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: userID,
	}).Err()
}

// GetStatus возвращает статус присутствия пользователя.
// Логика: если score < 5 мин назад -> "online", < 90 сек -> "away", иначе/нет записи -> "offline".
func (s *PresenceService) GetStatus(ctx context.Context, userID string) (string, error) {
	score, err := s.rdb.ZScore(ctx, presenceOnlineKey, userID).Result()
	if err == redis.Nil {
		return "offline", nil
	}
	if err != nil {
		return "offline", err
	}

	lastSeen := time.Unix(int64(score), 0)
	elapsed := time.Since(lastSeen)

	if elapsed < presenceAwayThreshold {
		return "online", nil
	}
	return "away", nil
}

// GetBulkStatus возвращает статусы присутствия для списка пользователей.
func (s *PresenceService) GetBulkStatus(ctx context.Context, userIDs []string) (map[string]string, error) {
	result := make(map[string]string, len(userIDs))

	pipe := s.rdb.Pipeline()
	cmds := make([]*redis.FloatCmd, len(userIDs))
	for i, uid := range userIDs {
		cmds[i] = pipe.ZScore(ctx, presenceOnlineKey, uid)
	}
	_, _ = pipe.Exec(ctx)

	now := time.Now()
	for i, uid := range userIDs {
		score, err := cmds[i].Result()
		if err == redis.Nil || err != nil {
			result[uid] = "offline"
			continue
		}
		lastSeen := time.Unix(int64(score), 0)
		elapsed := now.Sub(lastSeen)
		if elapsed < presenceAwayThreshold {
			result[uid] = "online"
		} else {
			result[uid] = "away"
		}
	}
	return result, nil
}

// ParseStatusDuration преобразует строковое значение длительности в time.Duration.
// Поддерживаемые значения: "30m", "1h", "4h", "today", "1d", "1w", "0" (без ограничения).
// Пустая строка или неизвестное значение возвращает 24 часа (по умолчанию).
func ParseStatusDuration(duration string) time.Duration {
	switch duration {
	case "30m":
		return 30 * time.Minute
	case "1h":
		return time.Hour
	case "4h":
		return 4 * time.Hour
	case "today":
		now := time.Now()
		midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		remaining := time.Until(midnight)
		if remaining < time.Minute {
			remaining = time.Minute
		}
		return remaining
	case "1d":
		return 24 * time.Hour
	case "1w":
		return 7 * 24 * time.Hour
	case "0":
		return 365 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

// SetCustomStatus устанавливает кастомный текстовый статус пользователя.
// duration — строковое значение длительности ("30m", "1h", "4h", "today", "1d", "1w", "0").
func (s *PresenceService) SetCustomStatus(ctx context.Context, userID, text, duration string) error {
	key := presenceStatusKeyPrefix + userID
	if text == "" {
		return s.rdb.Del(ctx, key).Err()
	}
	ttl := ParseStatusDuration(duration)
	return s.rdb.Set(ctx, key, text, ttl).Err()
}

// GetCustomStatus возвращает кастомный текстовый статус пользователя.
func (s *PresenceService) GetCustomStatus(ctx context.Context, userID string) (string, error) {
	val, err := s.rdb.Get(ctx, presenceStatusKeyPrefix+userID).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// GetOnlineUserIDs возвращает ID всех пользователей со статусом "online"
// (heartbeat менее 5 минут назад).
func (s *PresenceService) GetOnlineUserIDs(ctx context.Context) ([]string, error) {
	minScore := fmt.Sprintf("%d", time.Now().Add(-presenceAwayThreshold).Unix())
	maxScore := "+inf"
	return s.rdb.ZRangeByScore(ctx, presenceOnlineKey, &redis.ZRangeBy{
		Min: minScore,
		Max: maxScore,
	}).Result()
}

// CleanupStale удаляет из sorted set записи старше presenceStaleThreshold.
func (s *PresenceService) CleanupStale(ctx context.Context) error {
	maxScore := fmt.Sprintf("%d", time.Now().Add(-presenceStaleThreshold).Unix())
	return s.rdb.ZRemRangeByScore(ctx, presenceOnlineKey, "-inf", maxScore).Err()
}

// StartCleanupLoop запускает фоновую горутину, которая каждые 60 секунд
// очищает устаревшие записи присутствия. Завершается при отмене контекста.
func (s *PresenceService) StartCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(presenceCleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.CleanupStale(ctx); err != nil {
				slog.Error("ошибка очистки устаревших записей присутствия", "error", err)
			}
		}
	}
}
