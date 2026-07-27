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

// TestCallChat_Send_RejectsStranger_A2_003 — чужой user не шлёт в channel-call chat.
// Mutation: убрать authorizeCallChatAccess в SendMessage → этот тест FAIL (станет 201).
func TestCallChat_Send_RejectsStranger_A2_003(t *testing.T) {
	f := newCallsTestFixture(t)
	callEnt := f.createChannelCall(t, f.owner)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/chat",
		bytes.NewReader([]byte(`{"text":"pwn","ephemeral":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.stranger.ID.String())
	req.Header.Set("X-Test-User-Name", "Stranger")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// TestCallChat_List_RejectsStranger_A2_003 — list без membership/participant → 403.
func TestCallChat_List_RejectsStranger_A2_003(t *testing.T) {
	chatRepo := calls.ExportNewMemChatRepository()
	f := newCallsTestFixture(t)
	calls.ExportCallServiceSetChat(f.svc, calls.NewCallChatService(f.repo, chatRepo, calls.ExportCallServicePub(f.svc)))
	callEnt := f.createChannelCall(t, f.owner)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodGet, "/calls/"+callEnt.ID.String()+"/chat?limit=10", nil)
	req.Header.Set("X-Test-User-ID", f.stranger.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// TestCallChat_ChannelMemberAllowed_A2_003 — член канала (не participant) может send.
func TestCallChat_ChannelMemberAllowed_A2_003(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	msg, err := f.svc.SendChatMessage(ctx, callEnt.ID.String(), f.member.ID.String(), "Member", "hi", true)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "hi", msg.Text)
}

// TestCallChat_Instant_CreatedByAllowed_A2_003 — instant: created_by без participant OK.
func TestCallChat_Instant_CreatedByAllowed_A2_003(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createInstantCall(t, f.owner, "chatown1")

	msg, err := f.svc.SendChatMessage(ctx, callEnt.ID.String(), f.owner.ID.String(), "Owner", "instant", true)
	require.NoError(t, err)
	require.NotNil(t, msg)
}

// TestCallChat_Instant_StrangerForbidden_A2_003 — instant: не participant и не creator → forbid.
// Mutation: убрать authorizeCallChatAccess → FAIL (stranger получит msg).
func TestCallChat_Instant_StrangerForbidden_A2_003(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createInstantCall(t, f.owner, "chatstr1")

	_, err := f.svc.SendChatMessage(ctx, callEnt.ID.String(), f.stranger.ID.String(), "Stranger", "nope", true)
	require.ErrorIs(t, err, calls.ErrCallForbidden)

	_, err = f.svc.ListChatMessages(ctx, callEnt.ID.String(), f.stranger.ID.String(), 10)
	require.ErrorIs(t, err, calls.ErrCallForbidden)
}

// TestCallChat_Instant_ParticipantAllowed_A2_003 — guest после JoinByInvite может писать в chat.
func TestCallChat_Instant_ParticipantAllowed_A2_003(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createInstantCall(t, f.owner, "chatpart1")
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.member.ID))

	msg, err := f.svc.SendChatMessage(ctx, callEnt.ID.String(), f.member.ID.String(), "Guest", "joined", true)
	require.NoError(t, err)
	require.NotNil(t, msg)
}
