package ws

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// MembershipChecker возвращает true, если пользователь имеет доступ к
// сущности с указанным ID. Реализуется на стороне каждого модуля
// (messenger, calls, project) — это удерживает пакет ws отвязанным от
// конкретных репозиториев.
type MembershipChecker func(ctx context.Context, entityID, userID uuid.UUID) (bool, error)

// AuthorizerConfig инкапсулирует функции проверки членства для разных
// типов комнат. Все три проверки рекомендуется указывать; nil-значения
// приводят к отказу в подписке на соответствующий тип комнаты (fail-closed).
type AuthorizerConfig struct {
	// IsChannelMember — для комнат "channel:{id}" (messenger).
	IsChannelMember MembershipChecker
	// IsCallParticipantOrChannelMember — для комнат "call:{id}" (звонки).
	// Параметр entityID = callID; реализация может либо проверить участие
	// в самом звонке, либо членство в канале звонка.
	IsCallParticipantOrChannelMember MembershipChecker
	// IsProjectMember — для комнат "project:{id}".
	IsProjectMember MembershipChecker
}

// configAuthorizer — Authorizer на базе AuthorizerConfig.
type configAuthorizer struct {
	cfg AuthorizerConfig
}

// NewAuthorizer создаёт production-Authorizer из набора проверок членства.
// nil-checker для конкретного префикса означает fail-closed — подписки
// на этот тип комнат запрещены, что безопаснее, чем allow-all.
func NewAuthorizer(cfg AuthorizerConfig) Authorizer {
	return &configAuthorizer{cfg: cfg}
}

// CanSubscribe реализует Authorizer. Парсит room по префиксу и делегирует
// проверку соответствующему MembershipChecker. Неизвестные префиксы
// отклоняются (fail-closed).
func (a *configAuthorizer) CanSubscribe(ctx context.Context, userIDStr, room string) (bool, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return false, fmt.Errorf("ws authz: invalid user id %q: %w", userIDStr, err)
	}

	prefix, idStr, ok := splitRoom(room)
	if !ok {
		return false, fmt.Errorf("ws authz: malformed room %q", room)
	}
	entityID, err := uuid.Parse(idStr)
	if err != nil {
		return false, fmt.Errorf("ws authz: invalid entity id in room %q: %w", room, err)
	}

	switch prefix {
	case "channel":
		if a.cfg.IsChannelMember == nil {
			return false, nil
		}
		return a.cfg.IsChannelMember(ctx, entityID, userID)
	case "call":
		if a.cfg.IsCallParticipantOrChannelMember == nil {
			return false, nil
		}
		return a.cfg.IsCallParticipantOrChannelMember(ctx, entityID, userID)
	case "project":
		if a.cfg.IsProjectMember == nil {
			return false, nil
		}
		return a.cfg.IsProjectMember(ctx, entityID, userID)
	default:
		// Неизвестный префикс — fail-closed.
		return false, nil
	}
}

// splitRoom парсит room "{prefix}:{id}" на (prefix, id). Возвращает ok=false
// для пустых или некорректно сформированных значений.
func splitRoom(room string) (prefix, id string, ok bool) {
	idx := strings.IndexByte(room, ':')
	if idx <= 0 || idx >= len(room)-1 {
		return "", "", false
	}
	return room[:idx], room[idx+1:], true
}
