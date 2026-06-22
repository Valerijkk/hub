package intranet_test
import (
	"github.com/valerii/hub/backend/internal/intranet"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubCoffeeNotifier struct {
	calls int
}

func (s *stubCoffeeNotifier) NotifyCoffeeMatch(_ context.Context, _ uuid.UUID, _ intranet.CoffeePartnerResponse) error {
	s.calls++
	return nil
}

func TestService_CoffeeEnrollmentAndMatch(t *testing.T) {
	svc, adminID, memberID := setupIntranetServiceWithCoffee(t)
	ctx := context.Background()

	_, err := svc.SetCoffeeEnrollment(ctx, memberID, true)
	require.NoError(t, err)
	_, err = svc.SetCoffeeEnrollment(ctx, adminID, true)
	require.NoError(t, err)

	_, err = svc.RunCoffeeMatching(ctx, memberID)
	assert.Error(t, err)

	round, err := svc.RunCoffeeMatching(ctx, adminID)
	require.NoError(t, err)
	assert.Equal(t, 1, round.PairCount)

	match, err := svc.GetMyCoffeeMatch(ctx, memberID)
	require.NoError(t, err)
	require.NotNil(t, match.Partner)

	stats, err := svc.GetCoffeeStats(ctx, adminID)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.ActiveEnrollments)

	rounds, err := svc.ListCoffeeRounds(ctx, adminID)
	require.NoError(t, err)
	assert.Len(t, rounds, 1)
}

func TestService_GetCoffeeEnrollmentDefault(t *testing.T) {
	svc, _, memberID := setupIntranetService(t)
	en, err := svc.GetCoffeeEnrollment(context.Background(), memberID)
	require.NoError(t, err)
	assert.False(t, en.Enrolled)
}
