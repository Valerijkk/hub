package sprint

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

const (
	pokerSessionTTL   = 5 * time.Minute
	pokerVotesInitKey = "_init" // маркер пустого hash votes (Redis не хранит пустой HSET)
)

// fibonacciValues — допустимые значения голосов Planning Poker (Фибоначчи + ?).
var fibonacciValues = map[string]bool{
	"0": true, "1": true, "2": true, "3": true, "5": true,
	"8": true, "13": true, "21": true, "34": true, "55": true,
	"89": true, "?": true,
}

// castVoteScript — атомарно: fail-closed HGET revealed + HSET голоса + Expire обоих ключей.
// Возврат: 1=ok, 0=уже revealed, -1=сессия отсутствует.
var castVoteScript = redis.NewScript(`
local revealed = redis.call('HGET', KEYS[1], 'revealed')
if not revealed then
  return -1
end
if revealed == 'true' then
  return 0
end
redis.call('HSET', KEYS[2], ARGV[1], ARGV[2])
redis.call('HDEL', KEYS[2], '_init')
local ttl = tonumber(ARGV[3])
redis.call('EXPIRE', KEYS[1], ttl)
redis.call('EXPIRE', KEYS[2], ttl)
return 1
`)

// revealVotesScript — CAS revealed false→true (W1-A4-002).
// Возврат: 1=переключили, 0=уже true, -1=сессия отсутствует.
var revealVotesScript = redis.NewScript(`
local revealed = redis.call('HGET', KEYS[1], 'revealed')
if not revealed then
  return -1
end
if revealed == 'true' then
  return 0
end
redis.call('HSET', KEYS[1], 'revealed', 'true')
return 1
`)

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
// Сессия и votes-hash инициализируются в одном pipeline с TTL (W1-A4-003).
// projectID пишется в Redis для RBAC на vote/reveal/GET без projectId в URL (W1-A4-005).
func (p *PokerService) StartPokerSession(ctx context.Context, projectID, sprintID, issueID uuid.UUID) (*PokerSessionResponse, error) {
	sessionID := uuid.New().String()
	key := fmt.Sprintf("poker:%s", sessionID)
	votesKey := fmt.Sprintf("poker:%s:votes", sessionID)

	pipe := p.redis.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"issue_id":   issueID.String(),
		"sprint_id":  sprintID.String(),
		"project_id": projectID.String(),
		"revealed":   "false",
	})
	pipe.Expire(ctx, key, pokerSessionTTL)
	// Пустой votes hash: Redis не хранит пустой HSET — маркер + Expire сразу при Start.
	pipe.HSet(ctx, votesKey, pokerVotesInitKey, "1")
	pipe.Expire(ctx, votesKey, pokerSessionTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, apperrors.NewInternal("ошибка создания сессии Planning Poker")
	}

	p.nats.PublishJSON("poker.session_started", map[string]string{
		"session_id": sessionID,
		"issue_id":   issueID.String(),
		"sprint_id":  sprintID.String(),
		"project_id": projectID.String(),
	})

	return &PokerSessionResponse{
		SessionID: sessionID,
		IssueID:   issueID.String(),
		Revealed:  false,
	}, nil
}

// SessionProjectID возвращает project_id сессии из Redis (для RBAC).
func (p *PokerService) SessionProjectID(ctx context.Context, sessionID string) (uuid.UUID, error) {
	key := fmt.Sprintf("poker:%s", sessionID)
	projectIDStr, err := p.redis.HGet(ctx, key, "project_id").Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, apperrors.NewNotFound("сессия Planning Poker не найдена или истекла")
		}
		return uuid.Nil, apperrors.NewInternal("ошибка чтения сессии Planning Poker")
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return uuid.Nil, apperrors.NewInternal("повреждённый project_id сессии Planning Poker")
	}
	return projectID, nil
}

// CastVote записывает голос участника (SCRM-05).
// Значение должно быть из ряда Фибоначчи или "?".
// Голоса скрыты до раскрытия: NATS-событие содержит только факт голосования без значения.
// W1-A4-002/003/004: Lua CAS — голос после reveal отклоняется атомарно; TTL обновляется в том же скрипте.
func (p *PokerService) CastVote(ctx context.Context, sessionID string, userID uuid.UUID, value string) error {
	if !fibonacciValues[value] {
		return apperrors.NewBadRequest("недопустимое значение голоса, используйте ряд Фибоначчи: 0,1,2,3,5,8,13,21,34,55,89,?")
	}

	key := fmt.Sprintf("poker:%s", sessionID)
	votesKey := fmt.Sprintf("poker:%s:votes", sessionID)
	ttlSec := int(pokerSessionTTL / time.Second)

	res, err := castVoteScript.Run(ctx, p.redis, []string{key, votesKey}, userID.String(), value, ttlSec).Int64()
	if err != nil {
		return apperrors.NewInternal("ошибка записи голоса")
	}
	switch res {
	case -1:
		return apperrors.NewNotFound("сессия Planning Poker не найдена или истекла")
	case 0:
		return apperrors.NewBadRequest("голоса уже раскрыты")
	}

	p.nats.PublishJSON("poker.vote_cast", map[string]string{
		"session_id": sessionID,
		"user_id":    userID.String(),
		"voted":      "true",
	})

	return nil
}

// RevealVotes раскрывает все голоса сессии (SCRM-05).
// W1-A4-002: CAS через Lua — concurrent CastVote после reveal не проходит.
func (p *PokerService) RevealVotes(ctx context.Context, sessionID string) (*PokerSessionResponse, error) {
	key := fmt.Sprintf("poker:%s", sessionID)

	res, err := revealVotesScript.Run(ctx, p.redis, []string{key}).Int64()
	if err != nil {
		return nil, apperrors.NewInternal("ошибка раскрытия голосов")
	}
	switch res {
	case -1:
		return nil, apperrors.NewNotFound("сессия Planning Poker не найдена или истекла")
	case 0:
		// Уже раскрыто — идемпотентно отдаём текущее состояние.
	}

	votesKey := fmt.Sprintf("poker:%s:votes", sessionID)
	votes, err := p.redis.HGetAll(ctx, votesKey).Result()
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения голосов")
	}
	votes = filterPokerVotes(votes)

	issueID, err := p.redis.HGet(ctx, key, "issue_id").Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, apperrors.NewInternal("ошибка чтения issue_id сессии")
	}

	p.nats.PublishJSON("poker.revealed", map[string]interface{}{
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
	if err != nil {
		return nil, apperrors.NewInternal("ошибка чтения сессии Planning Poker")
	}
	if len(data) == 0 {
		return nil, apperrors.NewNotFound("сессия Planning Poker не найдена или истекла")
	}

	votesKey := fmt.Sprintf("poker:%s:votes", sessionID)
	allVotes, err := p.redis.HGetAll(ctx, votesKey).Result()
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения голосов")
	}
	allVotes = filterPokerVotes(allVotes)

	revealed := data["revealed"] == "true"

	resp := &PokerSessionResponse{
		SessionID: sessionID,
		IssueID:   data["issue_id"],
		Revealed:  revealed,
	}

	if revealed {
		resp.Votes = allVotes
	} else {
		voters := make([]string, 0, len(allVotes))
		for userID := range allVotes {
			voters = append(voters, userID)
		}
		resp.Voters = voters
	}

	return resp, nil
}

func filterPokerVotes(votes map[string]string) map[string]string {
	if votes == nil {
		return map[string]string{}
	}
	delete(votes, pokerVotesInitKey)
	return votes
}
