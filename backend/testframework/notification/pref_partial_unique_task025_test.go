//go:build hubtest

package notification_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/notification"
)

// TestNotificationPreference_PartialUnique_TASK025 — два Create с NULL project_id
// на один (user, type, channel) → constraint; per-project дубль тоже; global+project — ок.
// Mutation: вернуть обычный Unique(…, project_id) без partial → два NULL Create проходят → FAIL.
func TestNotificationPreference_PartialUnique_TASK025(t *testing.T) {
	_, repo, ctx, userID := newNotificationTestEnv(t)
	client := notification.ExportRepositoryDB(repo)

	mk := func(channel string, projectID *uuid.UUID) error {
		b := client.NotificationPreference.Create().
			SetUserID(userID).
			SetNotificationType(notification.TypeAssigned).
			SetChannel(channel).
			SetEnabled(true)
		if projectID != nil {
			b = b.SetProjectID(*projectID)
		}
		_, err := b.Save(ctx)
		return err
	}

	require.NoError(t, mk(notification.ChannelInApp, nil))
	err := mk(notification.ChannelInApp, nil)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err), "duplicate global pref: %v", err)

	projectID := uuid.New()
	require.NoError(t, mk(notification.ChannelInApp, &projectID))
	err = mk(notification.ChannelInApp, &projectID)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err), "duplicate project pref: %v", err)

	otherProject := uuid.New()
	require.NoError(t, mk(notification.ChannelInApp, &otherProject))
	require.NoError(t, mk(notification.ChannelTelegram, nil))
}
