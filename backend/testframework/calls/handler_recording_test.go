package calls_test
import (
	"github.com/valerii/hub/backend/internal/calls"
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/internal/platform/config"
)

func testCallsProjectKey(prefix string) string {
	suffix := strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", ""))[:5]
	return prefix + suffix
}

func TestToRecordingResponse_MapsAllFields(t *testing.T) {
	ended := time.Now().UTC().Truncate(time.Second)
	started := ended.Add(-2 * time.Minute)
	created := started.Add(-time.Minute)

	rec := &ent.Recording{
		ID:              uuid.New(),
		EgressID:        "egress-abc",
		Status:          recording.StatusReady,
		FilePath:        "recordings/p/c.mp4",
		FileSize:        4096,
		DurationSeconds: 120,
		StartedAt:       started,
		EndedAt:         &ended,
		CallID:          uuid.New(),
		ProjectID:       uuid.New(),
		StartedBy:       uuid.New(),
		CreatedAt:       created,
	}

	resp := calls.ExportToRecordingResponse(rec)
	require.NotNil(t, resp)
	assert.Equal(t, rec.ID, resp.ID)
	assert.Equal(t, "egress-abc", resp.EgressID)
	assert.Equal(t, "ready", resp.Status)
	assert.Equal(t, rec.FilePath, resp.FilePath)
	assert.Equal(t, int64(4096), resp.FileSize)
	assert.Equal(t, 120, resp.DurationSeconds)
	assert.Equal(t, started, resp.StartedAt)
	require.NotNil(t, resp.EndedAt)
	assert.Equal(t, ended, *resp.EndedAt)
	assert.Equal(t, rec.CallID, resp.CallID)
	assert.Equal(t, rec.ProjectID, resp.ProjectID)
	assert.Equal(t, rec.StartedBy, resp.StartedBy)
	assert.Equal(t, created, resp.CreatedAt)
}

func TestToRecordingResponse_OmitsEndedAtWhenNil(t *testing.T) {
	rec := &ent.Recording{
		ID:        uuid.New(),
		EgressID:  "egress-live",
		Status:    recording.StatusRecording,
		CallID:    uuid.New(),
		ProjectID: uuid.New(),
		StartedBy: uuid.New(),
		StartedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}

	resp := calls.ExportToRecordingResponse(rec)
	require.NotNil(t, resp)
	assert.Nil(t, resp.EndedAt)
}

func newHandlerValidationApp(h *calls.Handler) *fiber.App {
	app := fiber.New()
	auth := func(c fiber.Ctx) error {
		if uid := c.Get("X-Test-User-ID"); uid != "" {
			c.Locals("userID", uid)
		}
		return c.Next()
	}
	// Fiber v3: Get(path, terminal, middleware...) — middleware runs before terminal.
	app.Post("/calls/:callId/recording/start", h.StartRecording, auth)
	app.Post("/calls/:callId/recording/stop", h.StopRecording, auth)
	app.Get("/projects/:projectId/recordings", h.ListProjectRecordings, auth)
	app.Get("/recordings/:recordingId/download", h.GetRecordingDownloadURL, auth)
	return app
}

func TestParseCallID_Validation(t *testing.T) {
	validID := uuid.New()
	app := fiber.New()
	app.Get("/calls/:callId", func(c fiber.Ctx) error {
		id, err := calls.ExportParseCallID(c)
		if err != nil {
			return err
		}
		return c.SendString(id.String())
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{name: "valid uuid", path: "/" + validID.String(), wantStatus: fiber.StatusOK},
		{name: "invalid format", path: "/not-a-uuid", wantStatus: fiber.StatusBadRequest},
		{name: "nil uuid", path: "/" + uuid.Nil.String(), wantStatus: fiber.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodGet, "/calls"+tt.path, nil)
			resp, err := app.Test(req, 5*time.Second)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestStopRecording_Validation(t *testing.T) {
	h := calls.NewHandler(nil, nil, nil, config.LiveKitConfig{}, nil)
	app := newHandlerValidationApp(h)
	userID := uuid.New()

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "missing recording_id", body: `{}`, wantStatus: fiber.StatusBadRequest},
		{name: "nil recording_id", body: `{"recording_id":"00000000-0000-0000-0000-000000000000"}`, wantStatus: fiber.StatusBadRequest},
		{name: "invalid json", body: `{`, wantStatus: fiber.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodPost, "/calls/"+uuid.New().String()+"/recording/stop", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Test-User-ID", userID.String())

			resp, err := app.Test(req, 5*time.Second)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestStartRecording_InvalidCallID(t *testing.T) {
	h := calls.NewHandler(nil, nil, nil, config.LiveKitConfig{}, nil)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/not-a-uuid/recording/start", nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestListProjectRecordings_InvalidProjectID(t *testing.T) {
	h := calls.NewHandler(nil, nil, nil, config.LiveKitConfig{}, nil)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/projects/bad-id/recordings", nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetRecordingDownloadURL_InvalidRecordingID(t *testing.T) {
	h := calls.NewHandler(nil, nil, nil, config.LiveKitConfig{}, nil)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/recordings/not-uuid/download", nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandlerRecording_Unauthenticated(t *testing.T) {
	h := calls.NewHandler(nil, nil, nil, config.LiveKitConfig{}, nil)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+uuid.New().String()+"/recording/start", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func newRecordingWebhookFixture(t *testing.T) (*calls.Handler, config.LiveKitConfig, string) {
	t.Helper()
	dsn := "file:recwebhook_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@example.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Webhook Project").
		SetKey(testCallsProjectKey("W")).
		Save(ctx)
	require.NoError(t, err)

	c, err := client.Call.Create().
		SetRoomName("room-webhook").
		SetStatus(call.StatusActive).
		SetCreatedBy(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	egressID := "egress-webhook-" + uuid.New().String()[:8]
	_, err = client.Recording.Create().
		SetEgressID(egressID).
		SetStatus(recording.StatusProcessing).
		SetCallID(c.ID).
		SetProjectID(proj.ID).
		SetStartedBy(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	cfg := liveKitWebhookCfg()
	repo := calls.NewEntRepository(client)
	h := calls.ExportNewHandlerForTest(nil, calls.ExportNewRecordingServiceForTest(repo, nil, config.S3Config{}), nil, cfg, nil)

	return h, cfg, egressID
}

func TestLiveKitWebhook_EgressEnded_UpdatesRecording(t *testing.T) {
	h, cfg, egressID := newRecordingWebhookFixture(t)
	app := fiber.New()
	app.Post("/api/calls/webhook", h.HandleWebhook)

	body := []byte(`{"event":"egress_ended","egress_info":{"egress_id":"` + egressID + `","file_results":[{"filename":"rec.mp4","size":2048,"duration":45.7}]}}`)
	token := signLiveKitWebhook(t, body, cfg.APIKey, cfg.APISecret)

	req := httptest.NewRequest(fiber.MethodPost, "/api/calls/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestLiveKitWebhook_EgressEnded_NilRecSvcStillOK(t *testing.T) {
	cfg := liveKitWebhookCfg()
	app := newWebhookApp(t, cfg)

	body := []byte(`{"event":"egress_ended","egress_info":{"egress_id":"missing","file_results":[{"size":1,"duration":1}]}}`)
	token := signLiveKitWebhook(t, body, cfg.APIKey, cfg.APISecret)

	req := httptest.NewRequest(fiber.MethodPost, "/api/calls/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestLiveKitWebhook_InvalidJSONBody_Rejected(t *testing.T) {
	cfg := liveKitWebhookCfg()
	app := newWebhookApp(t, cfg)

	body := []byte(`{broken`)
	token := signLiveKitWebhook(t, body, cfg.APIKey, cfg.APISecret)

	req := httptest.NewRequest(fiber.MethodPost, "/api/calls/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestOnEgressEnded_UpdatesRecordingMetadata(t *testing.T) {
	dsn := "file:egresssvc_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@example.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Egress Project").
		SetKey(testCallsProjectKey("E")).
		Save(ctx)
	require.NoError(t, err)

	c, err := client.Call.Create().
		SetRoomName("room-egress").
		SetStatus(call.StatusActive).
		SetCreatedBy(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	egressID := "egress-svc-" + uuid.New().String()[:8]
	rec, err := client.Recording.Create().
		SetEgressID(egressID).
		SetStatus(recording.StatusProcessing).
		SetCallID(c.ID).
		SetProjectID(proj.ID).
		SetStartedBy(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	svc := calls.ExportNewRecordingServiceForTest(calls.NewEntRepository(client), nil, config.S3Config{})
	err = svc.OnEgressEnded(ctx, egressID, 8192, 90.4)
	require.NoError(t, err)

	updated, err := client.Recording.Get(ctx, rec.ID)
	require.NoError(t, err)
	assert.Equal(t, recording.StatusReady, updated.Status)
	assert.Equal(t, int64(8192), updated.FileSize)
	assert.Equal(t, 90, updated.DurationSeconds)
	require.NotNil(t, updated.EndedAt)
}
