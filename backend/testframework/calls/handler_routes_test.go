package calls_test
import (
	"github.com/valerii/hub/backend/internal/calls"
	"bytes"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/config"
)

func TestNewHandler(t *testing.T) {
	f := newCallsTestFixture(t)
	cfg := config.LiveKitConfig{URL: "http://localhost:7880", APIKey: "k", APISecret: "secretsecretsecretsecretsecretsec"}
	scheduled := calls.NewScheduledCallService(f.repo, nil)
	h := calls.NewHandler(f.svc, nil, scheduled, cfg, nil)
	require.NotNil(t, h)
	assert.Equal(t, f.svc, calls.ExportHandlerService(h))
}

func TestHandler_RegisterRoutes_Smoke(t *testing.T) {
	f := newCallsTestFixture(t)
	cfg := config.LiveKitConfig{URL: "http://localhost:7880", APIKey: "k", APISecret: "secretsecretsecretsecretsecretsec"}
	h := calls.NewHandler(f.svc, nil, calls.NewScheduledCallService(f.repo, nil), cfg, nil)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, testAuthMiddleware)

	req := httptest.NewRequest(fiber.MethodGet, "/api/calls/conflicts?time=bad&duration=30", nil)
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_UpdateScheduledCall_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	proj, err := f.client.Project.Create().SetName("Upd Proj").SetKey("UPD").Save(ctx)
	require.NoError(t, err)

	scheduledSvc := calls.NewScheduledCallService(f.repo, nil)
	when := time.Now().UTC().Add(2 * time.Hour)
	created, err := scheduledSvc.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Daily", ChannelID: f.channel.ID, ScheduledAt: when,
	})
	require.NoError(t, err)

	h := calls.ExportNewHandlerWithServiceAndScheduled(f.svc, scheduledSvc, config.LiveKitConfig{
		URL: "http://localhost:7880", APIKey: "testkey", APISecret: "testsecretthatshouldbelongenough",
	})
	app := newCallsValidationApp(h)

	newWhen := when.Add(time.Hour).Format(time.RFC3339)
	body := []byte(`{"title":"Daily Updated","scheduled_at":"` + newWhen + `"}`)
	req := httptest.NewRequest(fiber.MethodPut, "/scheduled-calls/"+created.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_DeleteScheduledCall_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	proj, err := f.client.Project.Create().SetName("Del Proj").SetKey("DEL").Save(ctx)
	require.NoError(t, err)

	scheduledSvc := calls.NewScheduledCallService(f.repo, nil)
	created, err := scheduledSvc.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Remove me", ChannelID: f.channel.ID, ScheduledAt: time.Now().UTC().Add(time.Hour),
	})
	require.NoError(t, err)

	h := calls.ExportNewHandlerWithServiceAndScheduled(f.svc, scheduledSvc, config.LiveKitConfig{
		URL: "http://localhost:7880", APIKey: "testkey", APISecret: "testsecretthatshouldbelongenough",
	})
	app := newCallsValidationApp(h)

	req := httptest.NewRequest(fiber.MethodDelete, "/scheduled-calls/"+created.ID.String(), nil)
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_JoinCall_Unauthenticated(t *testing.T) {
	f := newCallsTestFixture(t)
	callEnt := f.createChannelCall(t, f.owner)
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/join", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
