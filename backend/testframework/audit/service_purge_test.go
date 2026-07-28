package audit_test

// Тесты сервиса аудита: интеграционные и unit-сценарии testframework.
import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/audit"
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

func (m *purgeMockRepo) PurgeExpiredNewsPosts(_ context.Context, _ time.Time) (int, error) {
	if m.purgeErr != nil {
		return 0, m.purgeErr
	}
	return m.purgeCounts["newsposts"], nil
}

func (m *purgeMockRepo) PurgeExpiredScheduledCalls(_ context.Context, _ time.Time) (int, error) {
	if m.purgeErr != nil {
		return 0, m.purgeErr
	}
	return m.purgeCounts["scheduled_calls"], nil
}

type trashListMockRepo struct {
	mockAuditRepo
	items        []audit.TrashItem
	total        int
	lastCutoff   time.Time
	cutoffCalled bool
}

func (m *trashListMockRepo) ListDeletedIssues(_ context.Context, _ *uuid.UUID, cutoff time.Time, _, _ int) ([]audit.TrashItem, int, error) {
	m.lastCutoff = cutoff
	m.cutoffCalled = true
	return m.items, m.total, nil
}

// TestAuditService_PurgeExpired_Succeeds проверяет отклонение при audit сервис очистка просроченный succeeds.
func TestAuditService_PurgeExpired_Succeeds(t *testing.T) {
	repo := &purgeMockRepo{
		purgeCounts: map[string]int{
			"issues":     2,
			"test_cases": 1,
			"files":      0,
			"projects":   3,
			"newsposts":  1,
		},
	}
	svc := audit.NewAuditService(repo, nil)

	err := svc.PurgeExpired(context.Background())
	require.NoError(t, err)
}

// TestAuditService_PurgeExpired_LogsErrorsWithoutFailing проверяет отклонение при audit сервис очистка просроченный logs errors without failing.
func TestAuditService_PurgeExpired_LogsErrorsWithoutFailing(t *testing.T) {
	repo := &purgeMockRepo{purgeErr: errors.New("db unavailable")}
	svc := audit.NewAuditService(repo, nil)

	err := svc.PurgeExpired(context.Background())
	require.NoError(t, err)
}

// TestAuditService_ListTrash_WithItems — W6-A5-007 / P3-BE-02: cutoff ≈ now−30d.
func TestAuditService_ListTrash_WithItems(t *testing.T) {
	deletedAt := time.Now().Add(-5 * 24 * time.Hour)
	itemID := uuid.New()
	repo := &trashListMockRepo{
		items: []audit.TrashItem{{
			ID:         itemID,
			Name:       "BUG-42",
			EntityType: "issue",
			DeletedAt:  deletedAt,
		}},
		total: 1,
	}
	svc := audit.NewAuditService(repo, nil)

	before := time.Now().Add(-audit.TrashRetentionDays*24*time.Hour - time.Minute)
	resp, err := svc.ListTrash(context.Background(), "issue", nil, 1, 50)
	after := time.Now().Add(-audit.TrashRetentionDays*24*time.Hour + time.Minute)
	require.NoError(t, err)
	require.True(t, repo.cutoffCalled, "ListDeletedIssues must receive cutoff")
	assert.True(t, !repo.lastCutoff.Before(before) && !repo.lastCutoff.After(after),
		"cutoff %v must be within ±1m of now-%dd", repo.lastCutoff, audit.TrashRetentionDays)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, itemID.String(), resp.Items[0].ID)
	assert.Equal(t, "BUG-42", resp.Items[0].Name)
	assert.Equal(t, 1, resp.Total)
	assert.Greater(t, resp.Items[0].DaysRemaining, 0)
}

// TestAuditService_ListTrash_UnknownTypeReturnsEmpty проверяет бизнес-логику: audit сервис список корзина unknown type returns пустой.
func TestAuditService_ListTrash_UnknownTypeReturnsEmpty(t *testing.T) {
	svc := audit.NewAuditService(&mockAuditRepo{}, nil)

	resp, err := svc.ListTrash(context.Background(), "unknown", nil, 0, 200)
	require.NoError(t, err)
	assert.Empty(t, resp.Items)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 50, resp.PerPage)
}

// TestAuditService_SoftDelete_AllTypes проверяет бизнес-логику: audit сервис soft удаление all types.
func TestAuditService_SoftDelete_AllTypes(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()
	id := uuid.New()

	for _, entityType := range []string{"issue", "test_case", "file", "project", "newspost", "scheduled_call"} {
		require.NoError(t, svc.SoftDelete(ctx, entityType, id), entityType)
	}
}

// TestAuditService_RestoreItem_AllTypes проверяет бизнес-логику: audit сервис restore item all types.
func TestAuditService_RestoreItem_AllTypes(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()
	id := uuid.New()

	for _, entityType := range []string{"issue", "test_case", "file", "project", "newspost", "scheduled_call"} {
		require.NoError(t, svc.RestoreItem(ctx, entityType, id), entityType)
	}
}

// TestAuditService_IsSystemAdmin проверяет бизнес-логику: audit сервис is system админ.
func TestAuditService_IsSystemAdmin(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()

	ok, err := svc.IsSystemAdmin(ctx, adminID)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = svc.IsSystemAdmin(ctx, uuid.New())
	require.NoError(t, err)
	assert.False(t, ok)
}
