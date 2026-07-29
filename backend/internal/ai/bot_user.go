package ai

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
)

const (
	// BotEmail — системный пользователь @hub-ai (M3-009).
	BotEmail       = "hub-ai@hub.local"
	BotDisplayName = "@hub-ai"
	BotName        = "Hub AI"
)

// BotUserRepository — слой доступа к системному боту (без *ent.Client снаружи репо).
type BotUserRepository interface {
	EnsureBotUser(ctx context.Context) (uuid.UUID, error)
}

// EntBotUserRepository — BotUserRepository через Ent.
type EntBotUserRepository struct {
	client *ent.Client
}

// NewEntBotUserRepository создаёт репозиторий бота.
func NewEntBotUserRepository(client *ent.Client) *EntBotUserRepository {
	return &EntBotUserRepository{client: client}
}

// EnsureBotUser создаёт или возвращает системного бота @hub-ai.
// A3W5-3-008: *ent.Client только в EntBotUserRepository, не в сигнатуре пакета.
func (r *EntBotUserRepository) EnsureBotUser(ctx context.Context) (uuid.UUID, error) {
	if r == nil || r.client == nil {
		return uuid.Nil, fmt.Errorf("bot user repository required")
	}
	existing, err := r.client.User.Query().
		Where(user.EmailEQ(BotEmail)).
		Only(ctx)
	if err == nil {
		return existing.ID, nil
	}
	if !ent.IsNotFound(err) {
		return uuid.Nil, fmt.Errorf("query bot user: %w", err)
	}
	created, err := r.client.User.Create().
		SetEmail(BotEmail).
		SetName(BotName).
		SetDisplayName(BotDisplayName).
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		SetOnboardingCompleted(true).
		Save(ctx)
	if err != nil {
		// W5-A3-015 / TASK-113: parallel EnsureBot — unique conflict → re-query победителя.
		if existing, qerr := r.client.User.Query().Where(user.EmailEQ(BotEmail)).Only(ctx); qerr == nil {
			return existing.ID, nil
		}
		return uuid.Nil, fmt.Errorf("create bot user: %w", err)
	}
	return created.ID, nil
}

// EnsureBotUser создаёт или возвращает системного бота через репозиторий.
func EnsureBotUser(ctx context.Context, repo BotUserRepository) (uuid.UUID, error) {
	if repo == nil {
		return uuid.Nil, fmt.Errorf("bot user repository required")
	}
	return repo.EnsureBotUser(ctx)
}
