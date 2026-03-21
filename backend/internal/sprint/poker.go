package sprint

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// fibonacciValues — допустимые значения голосов Planning Poker (Фибоначчи + ?).
var fibonacciValues = map[string]bool{
	"0": true, "1": true, "2": true, "3": true, "5": true,
	"8": true, "13": true, "21": true, "34": true, "55": true,
	"89": true, "?": true,
}

// PokerService управляет сессиями Planning Poker через Redis + NATS.
type PokerService struct {
	redis *redis.Client
	nats  *natsclient.Client
}

// NewPokerService создаёт новый сервис Planning Poker.
func NewPokerService(rdb *redis.Client, nc *natsclient.Client) *PokerService {
	return &PokerService{redis: rdb, nats: nc}
}

// StartPokerSession создаёт новую сессию Planning Poker (SCRM-05).
// Сессия хранится в Redis с TTL 5 минут.
func (p *PokerService) StartPokerSession(ctx context.Context, sprintID, issueID uuid.UUID) (*PokerSessionResponse, error) {
	sessionID := uuid.New().String()
	key := fmt.Sprintf("poker:%s", sessionID)

	pipe := p.redis.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"issue_id":  issueID.String(),
		"sprint_id": sprintID.String(),
		"revealed":  "false",
	})
	pipe.Expire(ctx, key, 5*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, apperrors.NewInternal("ошибка создания сессии Planning Poker")
	}

	p.publishEvent("poker.session_started", map[string]string{
		"session_id": sessionID,
		"issue_id":   issueID.String(),
		"sprint_id":  sprintID.String(),
	})

	return &PokerSessionResponse{
		SessionID: sessionID,
		IssueID:   issueID.String(),
		Revealed:  false,
	}, nil
}

// CastVote записывает голос участника (SCRM-05).
// Значение должно быть из ряда Фибоначчи или "?".
// Голоса скрыты до раскрытия: NATS-событие содержит только факт голосования без значения.
func (p *PokerService) CastVote(ctx context.Context, sessionID string, userID uuid.UUID, value string) error {
	if !fibonacciValues[value] {
		return apperrors.NewBadRequest("недопустимое значение голоса, используйте ряд Фибоначчи: 0,1,2,3,5,8,13,21,34,55,89,?")
	}

	key := fmt.Sprintf("poker:%s", sessionID)
	exists, err := p.redis.Exists(ctx, key).Result()
	if err != nil || exists == 0 {
		return apperrors.NewNotFound("сессия Planning Poker не найдена или истекла")
	}

	// Проверяем что голоса ещё не раскрыты
	revealed, _ := p.redis.HGet(ctx, key, "revealed").Result()
	if revealed == "true" {
		return apperrors.NewBadRequest("голоса уже раскрыты")
	}

	votesKey := fmt.Sprintf("poker:%s:votes", sessionID)
	if err := p.redis.HSet(ctx, votesKey, userID.String(), value).Err(); err != nil {
		return apperrors.NewInternal("ошибка записи голоса")
	}
	// Обновляем TTL
	p.redis.Expire(ctx, votesKey, 5*time.Minute)
	p.redis.Expire(ctx, key, 5*time.Minute)

	// Публикуем событие без значения голоса (голоса скрыты)
	p.publishEvent("poker.vote_cast", map[string]string{
		"session_id": sessionID,
		"user_id":    userID.String(),
		"voted":      "true",
	})

	return nil
}

// RevealVotes раскрывает все голоса сессии (SCRM-05).
func (p *PokerService) RevealVotes(ctx context.Context, sessionID string) (*PokerSessionResponse, error) {
	key := fmt.Sprintf("poker:%s", sessionID)
	exists, err := p.redis.Exists(ctx, key).Result()
	if err != nil || exists == 0 {
		return nil, apperrors.NewNotFound("сессия Planning Poker не найдена или истекла")
	}

	// Устанавливаем revealed=true
	p.redis.HSet(ctx, key, "revealed", "true")

	// Получаем все голоса
	votesKey := fmt.Sprintf("poker:%s:votes", sessionID)
	votes, err := p.redis.HGetAll(ctx, votesKey).Result()
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения голосов")
	}

	issueID, _ := p.redis.HGet(ctx, key, "issue_id").Result()

	// Публикуем событие с полным списком голосов
	p.publishEvent("poker.revealed", map[string]interface{}{
		"session_id": sessionID,
		"issue_id":   issueID,
		"votes":      votes,
	})

	return &PokerSessionResponse{
		SessionID: sessionID,
		IssueID:   issueID,
		Revealed:  true,
		Votes:     votes,
	}, nil
}

// GetSession возвращает текущее состояние сессии Planning Poker.
// Если голоса раскрыты — возвращает значения, иначе только список проголосовавших.
func (p *PokerService) GetSession(ctx context.Context, sessionID string) (*PokerSessionResponse, error) {
	key := fmt.Sprintf("poker:%s", sessionID)
	data, err := p.redis.HGetAll(ctx, key).Result()
	if err != nil || len(data) == 0 {
		return nil, apperrors.NewNotFound("сессия Planning Poker не найдена или истекла")
	}

	votesKey := fmt.Sprintf("poker:%s:votes", sessionID)
	allVotes, _ := p.redis.HGetAll(ctx, votesKey).Result()

	revealed := data["revealed"] == "true"

	resp := &PokerSessionResponse{
		SessionID: sessionID,
		IssueID:   data["issue_id"],
		Revealed:  revealed,
	}

	if revealed {
		resp.Votes = allVotes
	} else {
		// Возвращаем только список проголосовавших без значений
		voters := make([]string, 0, len(allVotes))
		for userID := range allVotes {
			voters = append(voters, userID)
		}
		resp.Voters = voters
	}

	return resp, nil
}

// publishEvent публикует NATS-событие (nil-safe).
func (p *PokerService) publishEvent(subject string, payload interface{}) {
	if p.nats == nil {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("не удалось сериализовать poker event", "error", err)
		return
	}
	if err := p.nats.Publish(subject, data); err != nil {
		slog.Error("не удалось опубликовать poker event", "subject", subject, "error", err)
	}
}
