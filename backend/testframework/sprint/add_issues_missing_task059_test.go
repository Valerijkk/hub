//go:build hubtest

package sprint_test

// TASK-059 / P3-BE-03 / W1-A4-015: mock SetIssuesSprint rejects missing issueProject.
// Mutation: вернуть silent-skip для !ok → AddIssues без mapping зелёный → FAIL.

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/internal/sprint"
)

func TestAddIssuesToSprint_RejectsMissingIssueMapping_W1A4015(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	projectID := uuid.New()

	sprintID := uuid.New()
	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "S", Status: sprintEnt.StatusPlanning,
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14),
	}
	repo.sprintProject[sprintID] = projectID

	missing := uuid.New() // не в issueProject
	err := svc.AddIssuesToSprint(ctx, projectID, sprintID, []uuid.UUID{missing})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не принадлежат проекту")
	assert.Empty(t, repo.movedIssueIDs)
}
