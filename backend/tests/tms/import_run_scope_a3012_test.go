package tms_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestImportAllureToExistingRun_RejectsForeignProject_A3_012 —
// Mutation: снять HasProjectWith в GetTestRunInProject (или mock GetTestRunInProject
// без project-check) → импорт в чужой run проходит.
func TestImportAllureToExistingRun_RejectsForeignProject_A3_012(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectA := uuid.New()
	projectB := uuid.New()
	userID := uuid.New()

	runID := uuid.New()
	repo.runs[runID] = &ent.TestRun{
		ID: runID,
		Edges: ent.TestRunEdges{
			Project: &ent.Project{ID: projectB},
		},
	}

	_, err := svc.ImportAllureToExistingRun(ctx, projectA, runID, userID, tms.ImportAllureRequest{
		Results: []byte(`[{"name":"t","status":"passed","start":1,"stop":2,"uuid":"` + uuid.New().String() + `"}]`),
		Title:   "x",
	})
	require.Error(t, err)
}
