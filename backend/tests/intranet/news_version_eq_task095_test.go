//go:build hubtest

package intranet_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/intranet"
)

// TestUpdateNews_VersionEQ_TASK095 — W7-A2-003: VersionEQ в repo WHERE.
// Mutation: убрать VersionEQ из UpdateNewsPost → stale version снова проходит.
func TestUpdateNews_VersionEQ_TASK095(t *testing.T) {
	svc, adminID, _ := setupIntranetService(t)
	ctx := context.Background()

	created, err := svc.CreateNews(ctx, adminID, intranet.CreateNewsRequest{
		Title:     "OL title",
		Body:      "OL body",
		Published: true,
	})
	require.NoError(t, err)
	require.Equal(t, 1, created.Version)

	_, err = svc.UpdateNews(ctx, adminID, created.ID, intranet.UpdateNewsRequest{
		Version: 0,
	})
	assert.ErrorIs(t, err, intranet.ErrInvalidInput)

	newTitle := "v2"
	updated, err := svc.UpdateNews(ctx, adminID, created.ID, intranet.UpdateNewsRequest{
		Title:   &newTitle,
		Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Version)

	_, err = svc.UpdateNews(ctx, adminID, created.ID, intranet.UpdateNewsRequest{
		Title:   &newTitle,
		Version: created.Version, // stale
	})
	assert.ErrorIs(t, err, intranet.ErrVersionConflict)
}
