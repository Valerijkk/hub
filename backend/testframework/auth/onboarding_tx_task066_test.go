package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/auth"
)

// TestCompleteOnboarding_TxRollback_TASK066 — W6-A1-010: fail mark → profile rollback.
func TestCompleteOnboarding_TxRollback_TASK066(t *testing.T) {
	dsn := "file:onboard_tx_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("onboard-" + uuid.New().String()[:8] + "@example.com").
		SetDisplayName("before").
		Save(ctx)
	require.NoError(t, err)

	repo := auth.NewEntRepository(client)
	display := "ShouldNotPersist"
	_, err = repo.WithEntTx(ctx, func(ctx context.Context, txRepo auth.Repository) (*ent.User, error) {
		if _, err := txRepo.UpdateUserProfile(ctx, u.ID, auth.UpdateProfileRequest{DisplayName: &display}); err != nil {
			return nil, err
		}
		return nil, errors.New("force mark-complete failure")
	})
	require.Error(t, err)

	got, err := repo.GetUser(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "before", got.DisplayName)
	assert.False(t, got.OnboardingCompleted)

	// Happy path через ProfileService — атомарный complete.
	svc := auth.NewProfileService(repo)
	okDisplay := "AfterOnboard"
	completed, err := svc.CompleteOnboarding(ctx, u.ID, auth.OnboardingCompleteRequest{
		ProfileData: auth.UpdateProfileRequest{DisplayName: &okDisplay},
	})
	require.NoError(t, err)
	assert.True(t, completed.OnboardingCompleted)
	assert.Equal(t, okDisplay, completed.DisplayName)
}
