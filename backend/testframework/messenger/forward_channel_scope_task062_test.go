//go:build hubtest

package messenger_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
)

// TestForwardMessage_SourceChannelMismatch_Task062 — W3-A2-018 / P3-BE-07:
// path channelID должен совпадать с source.ChannelID.
// Мутация: убрать проверку → forward из чужого path channel проходит.
func TestForwardMessage_SourceChannelMismatch_Task062(t *testing.T) {
	repo := newMockRepository()
	svc := messenger.ExportNewMessengerServiceTest(repo)
	ctx := context.Background()
	ownerID := uuid.New()
	sourceID := setupChannelWithMembers(repo, "public", false, ownerID)
	otherID := setupChannelWithMembers(repo, "public", false, ownerID)
	targetID := setupChannelWithMembers(repo, "public", false, ownerID)

	sent, err := svc.SendMessage(ctx, sourceID, ownerID, messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"type": "doc"},
		ContentText: "secret",
	})
	require.NoError(t, err)

	_, err = svc.ForwardMessage(ctx, otherID, ownerID, sent.ID, targetID.String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдено")
}
