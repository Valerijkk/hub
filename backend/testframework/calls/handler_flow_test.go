package calls_test
import (
	"github.com/valerii/hub/backend/internal/calls"
	"bytes"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/config"
)

func newCallsHandlerApp(f *callsTestFixture) *fiber.App {
	h := calls.ExportNewHandlerForTest(f.svc, nil, nil, config.LiveKitConfig{
		URL:       "http://localhost:7880",
		APIKey:    "testkey",
		APISecret: "testsecretthatshouldbelongenough",
	}, nil)
	return newCallsValidationApp(h)
}

func TestHandler_JoinByInviteCode_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	code := "join1234"
	f.createInstantCall(t, f.owner, code)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodGet, "/calls/invite/"+code, nil)
	req.Header.Set("X-Test-User-ID", f.member.ID.String())
	req.Header.Set("X-Test-User-Name", "Member")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_GetActiveCall_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	f.createChannelCall(t, f.owner)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodGet, "/calls/channel/"+f.channel.ID.String(), nil)
	req.Header.Set("X-Test-User-ID", f.member.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_GetActiveCall_Empty(t *testing.T) {
	f := newCallsTestFixture(t)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodGet, "/calls/channel/"+f.channel.ID.String(), nil)
	req.Header.Set("X-Test-User-ID", f.member.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_JoinCall_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	callEnt := f.createChannelCall(t, f.owner)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/join", nil)
	req.Header.Set("X-Test-User-ID", f.member.ID.String())
	req.Header.Set("X-Test-User-Name", "Member")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_JoinCall_Forbidden(t *testing.T) {
	f := newCallsTestFixture(t)
	callEnt := f.createChannelCall(t, f.owner)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/join", nil)
	req.Header.Set("X-Test-User-ID", f.stranger.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_GuestJoinCall_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	code := "guest123"
	f.createInstantCall(t, f.owner, code)
	app := newCallsHandlerApp(f)

	body := []byte(`{"invite_code":"` + code + `","guest_name":"Guest User"}`)
	req := httptest.NewRequest(fiber.MethodPost, "/calls/guest-join", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_SendChatMessage_EphemeralSuccess(t *testing.T) {
	f := newCallsTestFixture(t)
	callEnt := f.createChannelCall(t, f.owner)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/chat", bytes.NewReader([]byte(`{"text":"hello","ephemeral":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	req.Header.Set("X-Test-User-Name", "Owner")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestHandler_CreateScheduledCall_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	proj, err := f.client.Project.Create().SetName("Call Proj").SetKey("CLP").Save(ctx)
	require.NoError(t, err)

	scheduledSvc := calls.NewScheduledCallService(f.repo, nil)
	h := calls.ExportNewHandlerWithServiceAndScheduled(f.svc, scheduledSvc, config.LiveKitConfig{
		URL: "http://localhost:7880", APIKey: "testkey", APISecret: "testsecretthatshouldbelongenough",
	})
	app := newCallsValidationApp(h)

	when := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	body := []byte(`{"title":"Standup","channel_id":"` + f.channel.ID.String() + `","scheduled_at":"` + when + `"}`)
	req := httptest.NewRequest(fiber.MethodPost, "/projects/"+proj.ID.String()+"/scheduled-calls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestHandler_ListScheduledCalls_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	proj, err := f.client.Project.Create().SetName("List Proj").SetKey("LSP").Save(ctx)
	require.NoError(t, err)

	scheduledSvc := calls.NewScheduledCallService(f.repo, nil)
	when := time.Now().UTC().Add(time.Hour)
	_, err = scheduledSvc.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Demo", ChannelID: f.channel.ID, ScheduledAt: when,
	})
	require.NoError(t, err)

	h := calls.ExportNewHandlerWithServiceAndScheduled(f.svc, scheduledSvc, config.LiveKitConfig{
		URL: "http://localhost:7880", APIKey: "testkey", APISecret: "testsecretthatshouldbelongenough",
	})
	app := newCallsValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/channels/"+f.channel.ID.String()+"/scheduled-calls", nil)
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_CheckConflicts_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	app := newCallsHandlerApp(f)

	when := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	req := httptest.NewRequest(fiber.MethodGet, "/calls/conflicts?time="+when+"&duration=30", nil)
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_LeaveCall_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	callEnt := f.createChannelCall(t, f.owner)
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))

	app := newCallsHandlerApp(f)
	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/leave", nil)
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_EndCall_CreatorSuccess(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	callEnt := f.createChannelCall(t, f.owner)
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))

	app := newCallsHandlerApp(f)
	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/end", nil)
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_EndCall_StrangerForbidden(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	callEnt := f.createChannelCall(t, f.owner)
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))

	app := newCallsHandlerApp(f)
	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/end", nil)
	req.Header.Set("X-Test-User-ID", f.stranger.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGenerateGuestToken_ValidJWT(t *testing.T) {
	h := calls.ExportNewHandlerWithCfg(config.LiveKitConfig{
		APIKey:    "testkey",
		APISecret: "testsecretthatshouldbelongenough",
	})
	token, err := calls.ExportGenerateGuestToken(h, "room-abc", "guest-"+uuid.New().String(), "Guest User")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestHandler_ListChatMessages_NotConfigured(t *testing.T) {
	f := newCallsTestFixture(t)
	callEnt := f.createChannelCall(t, f.owner)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodGet, "/calls/"+callEnt.ID.String()+"/chat?limit=10", nil)
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestHandler_RaiseHand_CallNotFound(t *testing.T) {
	f := newCallsTestFixture(t)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+uuid.New().String()+"/raise-hand", bytes.NewReader([]byte(`{"raised":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestHandler_CreateInstantCall_Unauthenticated(t *testing.T) {
	f := newCallsTestFixture(t)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/instant", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_CreateInstantCall_LiveKitUnavailable(t *testing.T) {
	f := newCallsTestFixture(t)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/instant", bytes.NewReader([]byte(`{"title":"Quick sync"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	req.Header.Set("X-Test-User-Name", "Owner")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestHandler_CreateCall_LiveKitUnavailable(t *testing.T) {
	f := newCallsTestFixture(t)
	app := newCallsHandlerApp(f)

	body := []byte(`{"channel_id":"` + f.channel.ID.String() + `"}`)
	req := httptest.NewRequest(fiber.MethodPost, "/calls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	req.Header.Set("X-Test-User-Name", "Owner")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestHandler_ListChatMessages_WithLimit(t *testing.T) {
	f := newCallsTestFixture(t)
	callEnt := f.createChannelCall(t, f.owner)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodGet, "/calls/"+callEnt.ID.String()+"/chat?limit=25", nil)
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestHandler_LeaveCall_NotFound(t *testing.T) {
	f := newCallsTestFixture(t)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+uuid.New().String()+"/leave", nil)
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
