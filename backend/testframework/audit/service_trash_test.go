package audit_test

// Тесты сервиса аудита: интеграционные и unit-сценарии testframework.
import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/audit"
)

// TestAuditService_ListTrash_Empty проверяет бизнес-логику: audit сервис список корзина пустой.
func TestAuditService_ListTrash_Empty(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	resp, err := svc.ListTrash(ctx, "issue", &projectID, 1, 50)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Items)
}

// TestAuditService_SoftDelete_UnknownType проверяет бизнес-логику: audit сервис soft удаление unknown type.
func TestAuditService_SoftDelete_UnknownType(t *testing.T) {
	svc := audit.NewAuditService(&mockAuditRepo{}, nil)
	err := svc.SoftDelete(context.Background(), "unknown", uuid.New())
	require.Error(t, err)
}

// TestAuditService_RestoreItem_UnknownType проверяет бизнес-логику: audit сервис restore item unknown type.
func TestAuditService_RestoreItem_UnknownType(t *testing.T) {
	svc := audit.NewAuditService(&mockAuditRepo{}, nil)
	err := svc.RestoreItem(context.Background(), "unknown", uuid.New())
	require.Error(t, err)
}

// TestMiddleware_shouldSkipHTTPAudit проверяет сценарий: middleware should skip http audit.
func TestMiddleware_shouldSkipHTTPAudit(t *testing.T) {
	tests := []struct {
		method, path string
		want         bool
	}{
		{"GET", "/api/health", true},
		{"POST", "/api/webhooks/alertmanager", true},
		{"GET", "/api/auth/me", false},
		{"OPTIONS", "/api/x", true},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, audit.ExportShouldSkipHTTPAudit(tt.method, tt.path), tt.method+" "+tt.path)
	}
}

// TestMiddleware_stripQuery проверяет сценарий: middleware strip query.
func TestMiddleware_stripQuery(t *testing.T) {
	assert.Equal(t, "/api/issues", audit.ExportStripQuery("/api/issues?page=1"))
}

// TestRepository_SoftDeleteIssue_NotFound проверяет слой репозитория: репозиторий soft удаление issue not found.
func TestRepository_SoftDeleteIssue_NotFound(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	err := repo.SoftDeleteIssue(context.Background(), uuid.New(), time.Now().UTC())
	require.Error(t, err)
}
