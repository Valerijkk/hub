//go:build hubtest

package ai_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/gofiber/fiber/v3"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/ai"
	trackermod "github.com/valerii/hub/backend/internal/tracker"
)

// A3W5-3-001: UTF-8 / mixed-case prefix trim без panic и с корректным remainder.
func TestNormalizeComposerCompletion_UTF8CaseFold_TASK155(t *testing.T) {
	t.Parallel()
	// Prefix «Тест»; completion с другим регистром — EqualFold trim по rune-границе.
	got := ai.ExportNormalizeComposerCompletion("Тест", "тест продолжение")
	assert.Equal(t, "продолжение", got)

	got = ai.ExportNormalizeComposerCompletion("Тест", "ТЕСТ ок")
	assert.Equal(t, "ок", got)

	got = ai.ExportNormalizeComposerCompletion("Hello", "HELLO world")
	assert.Equal(t, "world", got)
}

// A3W5-3-002: parallel IncrementUsage без unique fail; счётчики согласованы.
func TestIncrementUsage_ParallelRace_TASK155(t *testing.T) {
	repo, _, ctx, userID := newAIRepoEnv(t)
	const n = 20
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = repo.IncrementUsage(ctx, userID, 1, 2)
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		require.NoError(t, err, "goroutine %d", i)
	}
	in, out, req, err := repo.GetDailyUsage(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, n, in)
	assert.Equal(t, n*2, out)
	assert.Equal(t, n, req)
}

// A3W5-3-003: concurrent settings Upsert без unique violation.
func TestEntSettingsStore_ConcurrentUpsert_TASK155(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:aiset_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1&_busy_timeout=5000")
	t.Cleanup(func() { _ = client.Close() })
	store := ai.NewEntSettingsStore(client)
	ctx := context.Background()
	key := ai.ExportSettingBotSystemPrompt()

	const n = 12
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = store.Upsert(ctx, key, "prompt-"+string(rune('a'+i%5)), "bot prompt")
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		require.NoError(t, err, "goroutine %d", i)
	}
	val, err := store.Get(ctx, key)
	require.NoError(t, err)
	assert.NotEmpty(t, val)
	assert.Contains(t, val, "prompt-")
}

// A3W5-3-004: IssueLookup без SetAccessChecker → deny.
func TestTrackerIssueLookup_NilAccessDeny_TASK155(t *testing.T) {
	projID := uuid.New()
	l := ai.NewTrackerIssueLookup(&trackermod.TrackerService{}, stubProjectByKey{
		proj: &ent.Project{ID: projID, Key: "HUB"},
	})
	// без SetAccessChecker
	title, desc, err := l.IssueContext(context.Background(), uuid.New(), "HUB-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "issue access checker not configured")
	assert.Empty(t, title)
	assert.Empty(t, desc)
}

// A3W5-3-005: save preference model-only → allow_external_ai=false.
func TestSaveUserPreference_ModelOnlyAllowFalse_TASK155(t *testing.T) {
	repo, _, ctx, userID := newAIRepoEnv(t)
	model := "gpt-local"
	require.NoError(t, repo.SaveUserPreference(ctx, userID, ai.SaveUserPreferenceInput{
		DefaultModel: &model,
	}))
	rec, err := repo.GetUserPreferenceRecord(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, rec)
	assert.Equal(t, model, rec.DefaultModel)
	assert.False(t, rec.AllowExternalAI)
}

// A3W5-3-006: stale version → conflict; success bumps version.
func TestUpdateProvider_VersionConflict_TASK155(t *testing.T) {
	repo, _, ctx, _ := newAIRepoEnv(t)
	created, err := repo.CreateProvider(ctx, ai.CreateProviderInput{
		Name:         "V",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.example.com/v1",
		Enabled:      true,
	})
	require.NoError(t, err)
	require.Equal(t, 1, created.Version)

	name := "V2"
	updated, err := repo.UpdateProvider(ctx, created.ID, ai.UpdateProviderInput{
		Name: &name, Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Version)

	stale := "stale"
	_, err = repo.UpdateProvider(ctx, created.ID, ai.UpdateProviderInput{
		Name: &stale, Version: 1,
	})
	require.ErrorIs(t, err, ai.ErrVersionConflict)

	status, msg := ai.ExportMapProviderMutateError(ai.ErrVersionConflict)
	assert.Equal(t, fiber.StatusConflict, status)
	assert.Equal(t, "version conflict", msg)
}

func TestSaveUserPreference_StaleVersion_TASK155(t *testing.T) {
	repo, _, ctx, userID := newAIRepoEnv(t)
	model := "m1"
	require.NoError(t, repo.SaveUserPreference(ctx, userID, ai.SaveUserPreferenceInput{
		DefaultModel: &model,
	}))
	rec, err := repo.GetUserPreferenceRecord(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 1, rec.Version)

	model2 := "m2"
	err = repo.SaveUserPreference(ctx, userID, ai.SaveUserPreferenceInput{
		DefaultModel: &model2,
		Version:      99,
	})
	require.ErrorIs(t, err, ai.ErrVersionConflict)

	require.NoError(t, repo.SaveUserPreference(ctx, userID, ai.SaveUserPreferenceInput{
		DefaultModel: &model2,
		Version:      rec.Version,
	}))
	rec2, err := repo.GetUserPreferenceRecord(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, "m2", rec2.DefaultModel)
	assert.Equal(t, 2, rec2.Version)
}

// A3W5-3-007: stream error → фиксированное WS message без внутренней строки.
func TestStartChatStream_ErrorSanitized_TASK155(t *testing.T) {
	leak := "driver: llm-secret-dsn=internal"
	pid := uuid.New()
	repo := &mockRepo{providers: []*ent.AIProvider{{
		ID: pid, Name: "Ollama", BaseURL: "http://127.0.0.1:11434/v1",
		Models: []string{"llama3"}, Enabled: true, Scope: aiprovider.ScopeSystem,
	}}}
	svc := ai.NewService(repo, &mockAdapter{err: errors.New(leak)}, nil)
	streams := &mockStreamPublisher{}
	svc.SetStreamPublisher(streams)

	err := svc.StartChatStream(context.Background(), uuid.New(), ai.ChatCompletionRequest{
		Model:    "llama3",
		StreamID: "stream-err",
		Messages: []ai.ChatMessage{{Role: "user", Content: "hi there!!"}},
	}, "127.0.0.1", "test")
	require.NoError(t, err)

	deadline := time.Now().Add(2 * time.Second)
	for streams.eventCount() < 2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	require.GreaterOrEqual(t, streams.eventCount(), 2)
	var errEvt *streamEvent
	for i := 0; i < streams.eventCount(); i++ {
		ev, ok := streams.eventAt(i)
		require.True(t, ok)
		if ev.event == "error" {
			errEvt = &ev
			break
		}
	}
	require.NotNil(t, errEvt)
	msg, _ := errEvt.payload["message"].(string)
	assert.Equal(t, ai.ExportMsgAIStreamFailed(), msg)
	assert.NotContains(t, msg, leak)
	assert.NotContains(t, msg, "llm-secret")
}
