package audit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditService_ListTrash_Empty(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := NewAuditService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	resp, err := svc.ListTrash(ctx, "issue", &projectID, 1, 50)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Items)
}

func TestAuditService_SoftDelete_UnknownType(t *testing.T) {
	svc := NewAuditService(&mockAuditRepo{}, nil)
	err := svc.SoftDelete(context.Background(), "unknown", uuid.New())
	require.Error(t, err)
}

func TestAuditService_RestoreItem_UnknownType(t *testing.T) {
	svc := NewAuditService(&mockAuditRepo{}, nil)
	err := svc.RestoreItem(context.Background(), "unknown", uuid.New())
	require.Error(t, err)
}

func TestMiddleware_shouldSkipHTTPAudit(t *testing.T) {
	tests := []struct {
		method, path string
		want         bool
	}{
		{"GET", "/api/health", true},
		{"GET", "/api/auth/me", false},
		{"OPTIONS", "/api/x", true},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, shouldSkipHTTPAudit(tt.method, tt.path), tt.method+" "+tt.path)
	}
}

func TestMiddleware_stripQuery(t *testing.T) {
	assert.Equal(t, "/api/issues", stripQuery("/api/issues?page=1"))
}

func TestRepository_SoftDeleteIssue_NotFound(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	err := repo.SoftDeleteIssue(context.Background(), uuid.New(), time.Now().UTC())
	require.Error(t, err)
}
