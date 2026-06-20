package audit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
)

func TestListAuditLogs_MapsUserAndProjectIDs(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	projectID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	repo := &mockAuditRepo{}
	svc := NewAuditService(repo, nil)

	log := &ent.AuditLog{
		ID:         uuid.New(),
		Action:     "update",
		EntityType: "issue",
		EntityID:   "ISSUE-1",
		UserID:     &userID,
		ProjectID:  &projectID,
		Details:    map[string]interface{}{"field": "status"},
		IPAddress:  "10.0.0.1",
		UserAgent:  "test-agent",
		CreatedAt:  now,
	}
	repo.logs = []*ent.AuditLog{log}

	resp, err := svc.ListAuditLogs(ctx, AuditLogFilters{Page: 1, PerPage: 10})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)

	item := resp.Items[0]
	assert.Equal(t, log.ID.String(), item.ID)
	assert.Equal(t, "update", item.Action)
	assert.Equal(t, "issue", item.EntityType)
	assert.Equal(t, "ISSUE-1", item.EntityID)
	require.NotNil(t, item.UserID)
	assert.Equal(t, userID.String(), *item.UserID)
	require.NotNil(t, item.ProjectID)
	assert.Equal(t, projectID.String(), *item.ProjectID)
	assert.Equal(t, log.Details, item.Details)
	assert.Equal(t, "10.0.0.1", item.IPAddress)
	assert.Equal(t, "test-agent", item.UserAgent)
	assert.Equal(t, now, item.CreatedAt)
}

func TestListAuditLogs_OmitsNilOptionalUUIDs(t *testing.T) {
	ctx := context.Background()
	repo := &mockAuditRepo{}
	svc := NewAuditService(repo, nil)

	repo.logs = []*ent.AuditLog{{
		ID: uuid.New(), Action: "login", EntityType: "session",
		EntityID: "sess-1", CreatedAt: time.Now().UTC(),
	}}

	resp, err := svc.ListAuditLogs(ctx, AuditLogFilters{Page: 1, PerPage: 10})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Nil(t, resp.Items[0].UserID)
	assert.Nil(t, resp.Items[0].ProjectID)
}
