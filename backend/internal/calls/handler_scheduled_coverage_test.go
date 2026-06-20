package calls

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/internal/platform/config"
)

func TestFinalizeCreate_ReturnsTokenAndParticipant(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	resp, err := finalizeCreate(ctx, f.svc.room, f.svc.repo, callEnt, f.owner.ID, "Owner")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, callEnt.ID, resp.ID)

	participants, err := f.svc.repo.ListParticipants(ctx, callEnt.ID)
	require.NoError(t, err)
	require.Len(t, participants, 1)
	assert.Equal(t, f.owner.ID, participants[0].UserID)
}

func TestHandler_CheckConflicts_MissingTime(t *testing.T) {
	app := newCallsValidationApp(&Handler{scheduledSvc: &ScheduledCallService{}})

	req := httptest.NewRequest(fiber.MethodGet, "/calls/conflicts?duration=30", nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CheckConflicts_InvalidDuration(t *testing.T) {
	app := newCallsValidationApp(&Handler{scheduledSvc: &ScheduledCallService{}})

	when := time.Now().UTC().Format(time.RFC3339)
	req := httptest.NewRequest(fiber.MethodGet, "/calls/conflicts?time="+when+"&duration=0", nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CheckConflicts_EmptyParticipantIDs(t *testing.T) {
	app := newCallsValidationApp(&Handler{scheduledSvc: &ScheduledCallService{}})

	when := time.Now().UTC().Format(time.RFC3339)
	req := httptest.NewRequest(fiber.MethodGet, "/calls/conflicts?time="+when+"&duration=30&participant_ids=bad,also-bad", nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRecordingService_StopRecording_NotActive(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("RecStop").SetKey("RST").Save(ctx)
	require.NoError(t, err)
	callEnt := f.createChannelCall(t, f.owner)

	rec, err := f.client.Recording.Create().
		SetEgressID("egress-done").
		SetStatus(recording.StatusReady).
		SetCallID(callEnt.ID).
		SetProjectID(proj.ID).
		SetStartedBy(f.owner.ID).
		Save(ctx)
	require.NoError(t, err)

	svc := &RecordingService{repo: f.svc.repo}
	err = svc.StopRecording(ctx, rec.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не активна")
}

func TestHandler_StartRecording_ServiceError(t *testing.T) {
	f := newCallsTestFixture(t)
	callEnt := f.createInstantCall(t, f.owner, "nrec1")
	recSvc := &RecordingService{repo: f.svc.repo, s3Cfg: config.S3Config{Region: "us-east-1", Bucket: "b"}}
	h := &Handler{recSvc: recSvc}
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/recording/start", nil)
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestHandler_GetRecordingDownloadURL_NotReady(t *testing.T) {
	svc, _, client := newRecordingListFixture(t)
	ctx := context.Background()
	owner, err := client.User.Query().First(ctx)
	require.NoError(t, err)
	base, err := client.Recording.Query().First(ctx)
	require.NoError(t, err)

	rec, err := client.Recording.Create().
		SetEgressID("egress-handler-processing").
		SetStatus(recording.StatusProcessing).
		SetCallID(base.CallID).
		SetProjectID(base.ProjectID).
		SetStartedBy(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	h := &Handler{recSvc: svc}
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/recordings/"+rec.ID.String()+"/download", nil)
	req.Header.Set("X-Test-User-ID", owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
