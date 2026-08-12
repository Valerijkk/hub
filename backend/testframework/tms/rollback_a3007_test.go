package tms_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestRollbackTestCase_RejectsForeignVersion_A3_007 — service/mock contract:
// GetTestCaseVersion(testCaseID, versionID) rejects foreign ownership.
// Prod Ent ownership is covered by rollback_a3007_ent_test.go
// (HasTestCaseWith on Postgres).
func TestRollbackTestCase_RejectsForeignVersion_A3_007(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	mine, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Mine", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	other, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Other", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)
	otherID, err := uuid.Parse(other.ID)
	require.NoError(t, err)

	foreignVersionID := uuid.New()
	repo.versions[foreignVersionID] = &ent.TestCaseVersion{
		ID:            foreignVersionID,
		VersionNumber: 1,
		Snapshot: map[string]interface{}{
			"title": "Hijacked",
		},
		CreatedAt: time.Now(),
		Edges: ent.TestCaseVersionEdges{
			TestCase: &ent.TestCase{ID: otherID},
		},
	}

	_, err = svc.RollbackTestCase(ctx, projectID, mine.CaseKey, foreignVersionID, userID)
	require.Error(t, err)
}
