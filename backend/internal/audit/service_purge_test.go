package audit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type purgeMockRepo struct {
	mockAuditRepo
	purgeCounts map[string]int
	purgeErr    error
}

func (m *purgeMockRepo) PurgeExpiredIssues(_ context.Context, _ time.Time) (int, error) {
	if m.purgeErr != nil {
		return 0, m.purgeErr
	}
	return m.purgeCounts["issues"], nil
}

func (m *purgeMockRepo) PurgeExpiredTestCases(_ context.Context, _ time.Time) (int, error) {
	if m.purgeErr != nil {
		return 0, m.purgeErr
	}
	return m.purgeCounts["test_cases"], nil
}

func (m *purgeMockRepo) PurgeExpiredFiles(_ context.Context, _ time.Time) (int, error) {
	if m.purgeErr != nil {
		return 0, m.purgeErr
	}
	return m.purgeCounts["files"], nil
}

func (m *purgeMockRepo) PurgeExpiredProjects(_ context.Context, _ time.Time) (int, error) {
	if m.purgeErr != nil {
		return 0, m.purgeErr
	}
	return m.purgeCounts["projects"], nil
}

type trashListMockRepo struct {
	mockAuditRepo
	items []TrashItem
	total int
}

func (m *trashListMockRepo) ListDeletedIssues(_ context.Context, _ *uuid.UUID, _ time.Time, _, _ int) ([]TrashItem, int, error) {
	return m.items, m.total, nil
}

func TestAuditService_PurgeExpired_Succeeds(t *testing.T) {
	repo := &purgeMockRepo{
		purgeCounts: map[string]int{
			"issues":     2,
			"test_cases": 1,
			"files":      0,
			"projects":   3,
		},
	}
	svc := NewAuditService(repo, nil)

	err := svc.PurgeExpired(context.Background())
	require.NoError(t, err)
}

func TestAuditService_PurgeExpired_LogsErrorsWithoutFailing(t *testing.T) {
	repo := &purgeMockRepo{purgeErr: errors.New("db unavailable")}
	svc := NewAuditService(repo, nil)

	err := svc.PurgeExpired(context.Background())
	require.NoError(t, err)
}

func TestAuditService_ListTrash_WithItems(t *testing.T) {
	deletedAt := time.Now().Add(-5 * 24 * time.Hour)
	itemID := uuid.New()
	repo := &trashListMockRepo{
		items: []TrashItem{{
			ID:         itemID,
			Name:       "BUG-42",
			EntityType: "issue",
			DeletedAt:  deletedAt,
		}},
		total: 1,
	}
	svc := NewAuditService(repo, nil)

	resp, err := svc.ListTrash(context.Background(), "issue", nil, 1, 50)
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, itemID.String(), resp.Items[0].ID)
	assert.Equal(t, "BUG-42", resp.Items[0].Name)
	assert.Equal(t, 1, resp.Total)
	assert.Greater(t, resp.Items[0].DaysRemaining, 0)
}

func TestAuditService_ListTrash_UnknownTypeReturnsEmpty(t *testing.T) {
	svc := NewAuditService(&mockAuditRepo{}, nil)

	resp, err := svc.ListTrash(context.Background(), "unknown", nil, 0, 200)
	require.NoError(t, err)
	assert.Empty(t, resp.Items)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 50, resp.PerPage)
}

func TestAuditService_SoftDelete_AllTypes(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := NewAuditService(repo, nil)
	ctx := context.Background()
	id := uuid.New()

	for _, entityType := range []string{"issue", "test_case", "file", "project"} {
		require.NoError(t, svc.SoftDelete(ctx, entityType, id), entityType)
	}
}

func TestAuditService_RestoreItem_AllTypes(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := NewAuditService(repo, nil)
	ctx := context.Background()
	id := uuid.New()

	for _, entityType := range []string{"issue", "test_case", "file", "project"} {
		require.NoError(t, svc.RestoreItem(ctx, entityType, id), entityType)
	}
}

func TestAuditService_IsSystemAdmin(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	svc := NewAuditService(repo, nil)
	ctx := context.Background()

	ok, err := svc.IsSystemAdmin(ctx, adminID)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = svc.IsSystemAdmin(ctx, uuid.New())
	require.NoError(t, err)
	assert.False(t, ok)
}
