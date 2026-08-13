package calls_test

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/calls"
)

// TestRaiseHand_RejectsNonParticipant_A2_006 — stranger без записи участника
// получает 403; LiveKit не вызывается (нет 500 от недоступного room SDK).
func TestRaiseHand_RejectsNonParticipant_A2_006(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))

	app := newCallsHandlerApp(f)
	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/raise-hand",
		bytes.NewReader([]byte(`{"raised":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.stranger.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// TestRaiseHand_ServiceRejectsNonParticipant_A2_006 — сервисный слой возвращает
// ErrCallForbidden до LiveKit (mutation: убрать IsActiveParticipant check → FAIL).
func TestRaiseHand_ServiceRejectsNonParticipant_A2_006(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))

	err := f.svc.RaiseHand(ctx, callEnt.ID, f.stranger.ID, true)
	require.ErrorIs(t, err, calls.ErrCallForbidden)
}

// TestRaiseHand_ParticipantReachesLiveKit_A2_006 — активный участник проходит
// authz и доходит до LiveKit (ожидаем ошибку SDK без реального LiveKit).
func TestRaiseHand_ParticipantReachesLiveKit_A2_006(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))

	err := f.svc.RaiseHand(ctx, callEnt.ID, f.owner.ID, true)
	require.Error(t, err)
	assert.NotErrorIs(t, err, calls.ErrCallForbidden)
	assert.NotErrorIs(t, err, calls.ErrCallNotFound)
}
