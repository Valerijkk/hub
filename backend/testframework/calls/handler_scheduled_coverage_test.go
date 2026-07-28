package calls_test
import (
	"github.com/valerii/hub/backend/internal/calls"
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/internal/platform/config"
	"github.com/valerii/hub/backend/internal/rbac"
)

func TestFinalizeCreate_ReturnsTokenAndParticipant(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	resp, err := calls.ExportFinalizeCreate(ctx, calls.ExportCallServiceRoom(f.svc), f.repo, callEnt, f.owner.ID, "Owner")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, callEnt.ID, resp.ID)

	participants, err := f.repo.ListParticipants(ctx, callEnt.ID)
	require.NoError(t, err)
	require.Len(t, participants, 1)
	assert.Equal(t, f.owner.ID, participants[0].UserID)
}

func TestHandler_CheckConflicts_MissingTime(t *testing.T) {
	app := newCallsValidationApp(calls.ExportNewHandlerWithScheduledSvc(calls.ExportNewScheduledCallServiceForTest(nil)))

	req := httptest.NewRequest(fiber.MethodGet, "/calls/conflicts?duration=30", nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CheckConflicts_InvalidDuration(t *testing.T) {
	app := newCallsValidationApp(calls.ExportNewHandlerWithScheduledSvc(calls.ExportNewScheduledCallServiceForTest(nil)))

	when := time.Now().UTC().Format(time.RFC3339)
	req := httptest.NewRequest(fiber.MethodGet, "/calls/conflicts?time="+when+"&duration=0", nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CheckConflicts_EmptyParticipantIDs(t *testing.T) {
	app := newCallsValidationApp(calls.ExportNewHandlerWithScheduledSvc(calls.ExportNewScheduledCallServiceForTest(nil)))

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

	svc := calls.ExportNewRecordingServiceForTest(f.repo, nil, config.S3Config{})
	err = svc.StopRecording(ctx, callEnt.ID, rec.ID, f.owner.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не активна")
}

func TestHandler_StartRecording_ServiceError(t *testing.T) {
	f := newCallsTestFixture(t)
	callEnt := f.createInstantCall(t, f.owner, "nrec1")
	recSvc := calls.ExportNewRecordingServiceForTest(f.repo, nil, config.S3Config{Region: "us-east-1", Bucket: "b"})
	h := calls.ExportNewHandlerWithRecSvc(recSvc)
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

	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).
		SetProjectID(base.ProjectID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	rec, err := client.Recording.Create().
		SetEgressID("egress-handler-processing").
		SetStatus(recording.StatusProcessing).
		SetCallID(base.CallID).
		SetProjectID(base.ProjectID).
		SetStartedBy(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))

	h := calls.ExportNewHandlerWithRecSvcAndRBAC(svc, rbacSvc)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/recordings/"+rec.ID.String()+"/download", nil)
	req.Header.Set("X-Test-User-ID", owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
