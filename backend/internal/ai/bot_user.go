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

// EnsureBotUser создаёт или возвращает системного бота @hub-ai.
func EnsureBotUser(ctx context.Context, client *ent.Client) (uuid.UUID, error) {
	if client == nil {
		return uuid.Nil, fmt.Errorf("ent client required")
	}
	existing, err := client.User.Query().
		Where(user.EmailEQ(BotEmail)).
		Only(ctx)
	if err == nil {
		return existing.ID, nil
	}
	if !ent.IsNotFound(err) {
		return uuid.Nil, fmt.Errorf("query bot user: %w", err)
	}
	created, err := client.User.Create().
		SetEmail(BotEmail).
		SetName(BotName).
		SetDisplayName(BotDisplayName).
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		SetOnboardingCompleted(true).
		Save(ctx)
	if err != nil {
		// W5-A3-015: parallel EnsureBot — unique conflict / sqlite lock → re-query победителя.
		if existing, qerr := client.User.Query().Where(user.EmailEQ(BotEmail)).Only(ctx); qerr == nil {
			return existing.ID, nil
		}
		return uuid.Nil, fmt.Errorf("create bot user: %w", err)
	}
	return created.ID, nil
}
