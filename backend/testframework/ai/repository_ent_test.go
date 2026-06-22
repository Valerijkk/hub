package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/ent/enttest"
)

func newAIRepoEnv(t *testing.T) (*ai.EntRepository, *ent.Client, context.Context, uuid.UUID) {
	t.Helper()
	dsn := "file:airepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("ai-" + uuid.New().String()[:8] + "@test.com").
		SetName("AI User").
		Save(ctx)
	require.NoError(t, err)

	return ai.NewEntRepository(client, nil), client, ctx, user.ID
}

func TestEntRepository_ProviderCRUDAndModels(t *testing.T) {
	repo, _, ctx, userID := newAIRepoEnv(t)

	created, err := repo.CreateProvider(ctx, ai.CreateProviderInput{
		Name:         "DeepSeek",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.deepseek.com/v1",
		APIKey:       "sk-test",
		Models:       []string{"deepseek-chat"},
		Enabled:      true,
		Priority:     1,
		MaxRPM:       60,
	})
	require.NoError(t, err)
	assert.Equal(t, "DeepSeek", created.Name)
	assert.Equal(t, "sk-test", created.APIKeyEnc)

	list, err := repo.ListProviders(ctx, "system")
	require.NoError(t, err)
	require.Len(t, list, 1)

	got, err := repo.GetProvider(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "sk-test", got.APIKeyEnc)

	disabled := false
	updated, err := repo.UpdateProvider(ctx, created.ID, ai.UpdateProviderInput{
		Enabled: &disabled,
	})
	require.NoError(t, err)
	assert.False(t, updated.Enabled)

	models, err := repo.ListModels(ctx, &userID)
	require.NoError(t, err)
	assert.Empty(t, models)

	enabled := true
	_, err = repo.UpdateProvider(ctx, created.ID, ai.UpdateProviderInput{Enabled: &enabled})
	require.NoError(t, err)

	models, err = repo.ListModels(ctx, &userID)
	require.NoError(t, err)
	require.Len(t, models, 1)
	assert.Equal(t, "deepseek-chat", models[0].ID)

	found, err := repo.FindProviderForModel(ctx, userID, "deepseek-chat")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)

	_, err = repo.CreateProvider(ctx, ai.CreateProviderInput{
		Name:         "Bad",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "file:///etc/passwd",
	})
	require.Error(t, err)
}

func TestEntRepository_UserProviderAndPreferences(t *testing.T) {
	repo, _, ctx, userID := newAIRepoEnv(t)

	userProv, err := repo.CreateProvider(ctx, ai.CreateProviderInput{
		Name:         "My Key",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.openai.com/v1",
		Models:       []string{"gpt-4o-mini"},
		Enabled:      true,
		Scope:        "user",
		OwnerUserID:  &userID,
	})
	require.NoError(t, err)

	userList, err := repo.ListUserProviders(ctx, userID)
	require.NoError(t, err)
	require.Len(t, userList, 1)

	gotUser, err := repo.GetUserProvider(ctx, userID, userProv.ID)
	require.NoError(t, err)
	assert.Equal(t, userProv.ID, gotUser.ID)

	model := "gpt-4o-mini"
	require.NoError(t, repo.SaveUserPreference(ctx, userID, ai.SaveUserPreferenceInput{
		DefaultModel:      &model,
		DefaultProviderID: &userProv.ID,
	}))

	defModel, provID, err := repo.GetUserPreference(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, "gpt-4o-mini", defModel)
	require.NotNil(t, provID)
	assert.Equal(t, userProv.ID, *provID)
}

func TestEntRepository_GetUserPreference_Empty(t *testing.T) {
	repo, _, ctx, userID := newAIRepoEnv(t)

	model, provID, err := repo.GetUserPreference(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, model)
	assert.Nil(t, provID)

	rec, err := repo.GetUserPreferenceRecord(ctx, userID)
	require.NoError(t, err)
	assert.Nil(t, rec)
}

func TestEntRepository_UpdateProvider_AllFields(t *testing.T) {
	repo, _, ctx, _ := newAIRepoEnv(t)

	created, err := repo.CreateProvider(ctx, ai.CreateProviderInput{
		Name:         "Orig",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.example.com/v1",
		APIKey:       "sk-old",
		Models:       []string{"m1"},
		Enabled:      true,
		Priority:     1,
		MaxRPM:       30,
	})
	require.NoError(t, err)

	newName := "Renamed"
	newURL := "https://api.new.com/v1"
	newModels := []string{"m2", "m3"}
	enabled := false
	priority := 5
	maxRPM := 120
	ru := true
	emptyKey := ""
	updated, err := repo.UpdateProvider(ctx, created.ID, ai.UpdateProviderInput{
		Name:       &newName,
		BaseURL:    &newURL,
		Models:     newModels,
		Enabled:    &enabled,
		Priority:   &priority,
		MaxRPM:     &maxRPM,
		RuResident: &ru,
		APIKey:     &emptyKey,
	})
	require.NoError(t, err)
	assert.Equal(t, "Renamed", updated.Name)
	assert.Equal(t, newURL, updated.BaseURL)
	assert.Equal(t, newModels, []string(updated.Models))
	assert.False(t, updated.Enabled)
	assert.Equal(t, 5, updated.Priority)
	assert.Equal(t, 120, updated.MaxRpm)
	assert.True(t, updated.RuResident)
	assert.Empty(t, updated.APIKeyEnc)

	newKey := "sk-new"
	_, err = repo.UpdateProvider(ctx, created.ID, ai.UpdateProviderInput{APIKey: &newKey})
	require.NoError(t, err)
	got, err := repo.GetProvider(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "sk-new", got.APIKeyEnc)
}

func TestEntRepository_UpdateProvider_InvalidURL(t *testing.T) {
	repo, _, ctx, _ := newAIRepoEnv(t)
	created, err := repo.CreateProvider(ctx, ai.CreateProviderInput{
		Name:         "P",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.example.com/v1",
	})
	require.NoError(t, err)

	bad := "file:///etc/passwd"
	_, err = repo.UpdateProvider(ctx, created.ID, ai.UpdateProviderInput{BaseURL: &bad})
	require.Error(t, err)
}

func TestEntRepository_UsageTracking(t *testing.T) {
	repo, _, ctx, userID := newAIRepoEnv(t)

	require.NoError(t, repo.IncrementUsage(ctx, userID, 100, 50))
	in, out, req, err := repo.GetDailyUsage(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 100, in)
	assert.Equal(t, 50, out)
	assert.Equal(t, 1, req)

	require.NoError(t, repo.IncrementUsage(ctx, userID, 20, 10))
	in, out, req, err = repo.GetDailyUsage(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 120, in)
	assert.Equal(t, 60, out)
	assert.Equal(t, 2, req)

	reqs, tin, tout, users, err := repo.GetUsageSummary(ctx, time.Now().UTC())
	require.NoError(t, err)
	assert.Equal(t, 2, reqs)
	assert.Equal(t, 120, tin)
	assert.Equal(t, 60, tout)
	assert.Equal(t, 1, users)

	history, err := repo.ListUsageHistory(ctx, 7)
	require.NoError(t, err)
	require.Len(t, history, 1)
}
