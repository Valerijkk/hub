//go:build hubtest && !release

package dev_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/dev"
)

func newSeedHandler(t *testing.T) (*dev.Handler, *ent.Client) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:seed_a3w75_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	return dev.NewHandler(client, nil), client
}

// TestSeedWorkflowTransitions_CreateFailPropagates_A3W75002 — FK fail → err, не (n, nil).
// Mutation: вернуть slog+continue + (n, nil) → err==nil → FAIL.
func TestSeedWorkflowTransitions_CreateFailPropagates_A3W75002(t *testing.T) {
	h, client := newSeedHandler(t)
	ctx := context.Background()

	proj, err := client.Project.Create().
		SetName("Seed DEMO").
		SetKey("SDMO").
		Save(ctx)
	require.NoError(t, err)

	// Несуществующие status UUID → Create transition падает по FK.
	statusMap := map[string]uuid.UUID{
		"Todo": uuid.New(),
		"Done": uuid.New(),
	}
	n, err := dev.ExportSeedWorkflowTransitions(h, ctx, proj.ID, statusMap)
	assert.Error(t, err, "Create FK fail must propagate")
	assert.Equal(t, 0, n)
}

// TestSeedIssueLinks_CreateFailPropagates_A3W75002 — fake issue IDs → err.
func TestSeedIssueLinks_CreateFailPropagates_A3W75002(t *testing.T) {
	h, _ := newSeedHandler(t)
	ctx := context.Background()

	issueMap := map[string]uuid.UUID{
		"DEMO-4": uuid.New(),
		"DEMO-2": uuid.New(),
	}
	n, err := dev.ExportSeedIssueLinks(h, ctx, issueMap)
	assert.Error(t, err, "Create link with missing issues must propagate")
	assert.Equal(t, 0, n)
}

// TestSeedChannels_CreateFailPropagates_A3W75002 — без pm@hub.local → CreatedBy nil FK fail → err.
func TestSeedChannels_CreateFailPropagates_A3W75002(t *testing.T) {
	h, client := newSeedHandler(t)
	ctx := context.Background()

	proj, err := client.Project.Create().
		SetName("Seed CH").
		SetKey("SCHN").
		Save(ctx)
	require.NoError(t, err)

	// Пустой userIDs → created_by = uuid.Nil → FK fail на Create канала.
	n, err := dev.ExportSeedChannels(h, ctx, proj.ID, map[string]uuid.UUID{})
	assert.Error(t, err, "Channel Create without pm user must propagate")
	assert.Equal(t, 0, n)
}
