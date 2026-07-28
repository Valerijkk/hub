package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// WS ticket — одноразовый short-lived обмен для WebSocket handshake
// (ADR-WS-JWT-HANDSHAKE-2026). Access JWT не передаётся в query.

const (
	wsTicketRedisPrefix = "hub:ws:ticket:"
	// wsTicketTTL — окно на upgrade после Issue (reconnect/сеть).
	wsTicketTTL = 30 * time.Second
	// wsTicketBytes — 32 байта → 64 hex (достаточно энтропии, без UUID-формата в протоколе).
	wsTicketBytes = 32
)

// WSTicketClaims — данные, связанные с ticket до redeem.
type WSTicketClaims struct {
	UserID    string `json:"uid"`
	SessionID string `json:"sid"`
}

// IssueWSTicket создаёт одноразовый ticket в Redis (TTL 30s).
func (s *AuthService) IssueWSTicket(ctx context.Context, userID, sessionID string) (ticket string, expiresIn int, err error) {
	if s.redis == nil {
		return "", 0, fmt.Errorf("issue ws ticket: redis unavailable")
	}
	raw := make([]byte, wsTicketBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", 0, fmt.Errorf("issue ws ticket: rand: %w", err)
	}
	ticket = hex.EncodeToString(raw)
	payload, err := json.Marshal(WSTicketClaims{UserID: userID, SessionID: sessionID})
	if err != nil {
		return "", 0, fmt.Errorf("issue ws ticket: marshal: %w", err)
	}
	key := wsTicketRedisPrefix + ticket
	if err := s.redis.Set(ctx, key, payload, wsTicketTTL).Err(); err != nil {
		return "", 0, fmt.Errorf("issue ws ticket: redis set: %w", err)
	}
	return ticket, int(wsTicketTTL.Seconds()), nil
}

// RedeemWSTicket атомарно извлекает и удаляет ticket (one-shot).
// Повторный redeem того же ticket → ошибка (как «не найден»).
func (s *AuthService) RedeemWSTicket(ctx context.Context, ticket string) (*WSTicketClaims, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redeem ws ticket: redis unavailable")
	}
	if ticket == "" {
		return nil, fmt.Errorf("redeem ws ticket: empty")
	}
	key := wsTicketRedisPrefix + ticket
	raw, err := s.redis.GetDel(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("redeem ws ticket: %w", err)
	}
	var claims WSTicketClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, fmt.Errorf("redeem ws ticket: unmarshal: %w", err)
	}
	if claims.UserID == "" || claims.SessionID == "" {
		return nil, fmt.Errorf("redeem ws ticket: incomplete claims")
	}
	return &claims, nil
}

// WSTicketTTLSeconds экспортирует TTL для DTO/тестов.
func WSTicketTTLSeconds() int { return int(wsTicketTTL.Seconds()) }
