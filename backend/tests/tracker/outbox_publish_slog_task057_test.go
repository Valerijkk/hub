//go:build hubtest

package tracker_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tracker"
)

// failingOutboxPublisher — мутация P3-BE-01: Publish всегда err.
type failingOutboxPublisher struct{}

func (f *failingOutboxPublisher) PublishInTx(_ context.Context, _ *ent.Tx, _ string, _ any) error {
	return nil
}

func (f *failingOutboxPublisher) Publish(_ context.Context, _ *ent.Client, _ string, _ any) error {
	return errors.New("outbox down")
}

// TestUpdateIssue_PublishFail_LogsWarn_TASK057 — W1-A3-008 / P3-BE-01.
// Мутация: убрать logOutboxPublishErr / вернуть `_ = Publish` → slog без
// "outbox publish failed", тест краснеет.
func TestUpdateIssue_PublishFail_LogsWarn_TASK057(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	env.repo = tracker.NewEntRepository(env.client, &failingOutboxPublisher{})

	created, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Publish slog", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	newTitle := "updated-title"
	updated, err := env.repo.UpdateIssue(env.ctx, created.ID, tracker.UpdateIssueInput{
		Title:   &newTitle,
		Version: created.Version,
	})
	require.NoError(t, err, "post-commit Publish fail must not fail UpdateIssue")
	assert.Equal(t, newTitle, updated.Title)

	logs := buf.String()
	assert.True(t, strings.Contains(logs, "outbox publish failed"),
		"expected slog warn on Publish err, got: %s", logs)
	assert.True(t, strings.Contains(logs, "outbox down"),
		"expected error detail in slog, got: %s", logs)
}
