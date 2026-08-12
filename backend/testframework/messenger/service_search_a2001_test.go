package messenger_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

// TestSearchMessages_DeniesNonMember_A2_001 — non-member must not see foreign-channel hits.
func TestSearchMessages_DeniesNonMember_A2_001(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	doc := &messenger.MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    ownerID,
		ContentText: "secret-token-a2001",
		CreatedAt:   time.Now(),
	}
	repo.messages[doc.ID] = doc

	// Explicit channel_id without membership → 403.
	_, err := svc.SearchMessages(ctx, outsiderID, "secret-token-a2001", channelID.String(), 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")

	// Global search without membership must not return the foreign hit.
	results, err := svc.SearchMessages(ctx, outsiderID, "secret-token-a2001", "", 10)
	require.NoError(t, err)
	assert.Empty(t, results)

	// Member still finds it.
	results, err = svc.SearchMessages(ctx, ownerID, "secret-token-a2001", channelID.String(), 10)
	require.NoError(t, err)
	require.Len(t, results, 1)
}
