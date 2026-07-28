package messenger

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// Префикс ключа Redis для хранения непрочитанных сообщений.
const unreadKeyPrefix = "unread:"

// UnreadService управляет счётчиками непрочитанных сообщений
// через Redis hash. Ключ: "unread:{userID}", поле: channelID, значение: count.
type UnreadService struct {
	rdb *redis.Client
}

// NewUnreadService создаёт новый сервис непрочитанных сообщений.
func NewUnreadService(rdb *redis.Client) *UnreadService {
	return &UnreadService{rdb: rdb}
}

// Increment увеличивает счётчик непрочитанных на 1 для всех участников канала,
// кроме отправителя (senderID). Использует pipeline для атомарности и производительности.
func (s *UnreadService) Increment(ctx context.Context, channelID string, memberIDs []string, senderID string) error {
	if len(memberIDs) == 0 {
		return nil
	}

	pipe := s.rdb.Pipeline()
	for _, mid := range memberIDs {
		if mid == senderID {
			continue
		}
		pipe.HIncrBy(ctx, unreadKeyPrefix+mid, channelID, 1)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// MarkRead сбрасывает счётчик непрочитанных для конкретного канала пользователя.
func (s *UnreadService) MarkRead(ctx context.Context, userID, channelID string) error {
	return s.rdb.HDel(ctx, unreadKeyPrefix+userID, channelID).Err()
}

// GetAll возвращает все счётчики непрочитанных для пользователя.
// Ключ map — channelID, значение — количество непрочитанных.
func (s *UnreadService) GetAll(ctx context.Context, userID string) (map[string]int64, error) {
	raw, err := s.rdb.HGetAll(ctx, unreadKeyPrefix+userID).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64, len(raw))
	for k, v := range raw {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			// W3-A3-010 / P3-BE-01: битые значения Redis не маскировать в 0 без сигнала.
			slog.WarnContext(ctx, "unread counter corrupt",
				"channel_id", k, "raw", v, "error", err)
			continue
		}
		result[k] = n
	}
	return result, nil
}
