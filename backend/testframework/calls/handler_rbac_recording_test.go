package calls_test
import (
	"github.com/valerii/hub/backend/internal/calls"
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/internal/platform/config"
	"github.com/valerii/hub/backend/internal/rbac"
)

func newRecordingHandlerRBACFixture(t *testing.T) (*calls.Handler, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	dsn := "file:rechrbac_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
	ctx := context.Background()

	owner, err := client.User.Create().SetEmail("own-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
	require.NoError(t, err)
	stranger, err := client.User.Create().SetEmail("str-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().SetName("RecRBAC").SetKey(testCallsProjectKey("B")).Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().SetUserID(owner.ID).SetProjectID(proj.ID).SetProjectRole("owner").Save(ctx)
	require.NoError(t, err)

	c, err := client.Call.Create().SetRoomName("room-rbac").SetStatus("active").SetCreatedBy(owner.ID).Save(ctx)
	require.NoError(t, err)

	rec, err := client.Recording.Create().
		SetEgressID("egress-rbac").
		SetStatus(recording.StatusReady).
		SetCallID(c.ID).
		SetProjectID(proj.ID).
		SetStartedBy(owner.ID).
		SetFilePath("recordings/rbac/file.mp4").
		Save(ctx)
	require.NoError(t, err)

	recSvc := calls.NewRecordingService(
		config.LiveKitConfig{URL: "http://localhost:7880", APIKey: "k", APISecret: "testsecretthatshouldbelongenough"},
		config.S3Config{
			Region: "us-east-1", AccessKey: "accesskey", SecretKey: "secretkeysecret",
			Bucket: "recordings", Endpoint: "localhost:8333", UseSSL: false,
		},
		calls.NewEntRepository(client),
		nil,
	)

	h := calls.ExportNewHandlerForTest(nil, recSvc, nil, config.LiveKitConfig{}, rbacSvc)
	return h, owner.ID, stranger.ID, rec.ID, proj.ID
}

func TestHandler_GetRecordingDownloadURL_ForbiddenNonMember(t *testing.T) {
	h, _, strangerID, recID, _ := newRecordingHandlerRBACFixture(t)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/recordings/"+recID.String()+"/download", nil)
	req.Header.Set("X-Test-User-ID", strangerID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	_ = h
}

func TestHandler_GetRecordingDownloadURL_MemberSuccess(t *testing.T) {
	h, ownerID, _, recID, _ := newRecordingHandlerRBACFixture(t)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/recordings/"+recID.String()+"/download", nil)
	req.Header.Set("X-Test-User-ID", ownerID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_GetRecordingDownloadURL_NotFoundWithRBAC(t *testing.T) {
	h, ownerID, _, _, _ := newRecordingHandlerRBACFixture(t)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/recordings/"+uuid.New().String()+"/download", nil)
	req.Header.Set("X-Test-User-ID", ownerID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestHandler_ListProjectRecordings_ForbiddenNonMember(t *testing.T) {
	h, _, strangerID, _, projectID := newRecordingHandlerRBACFixture(t)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/projects/"+projectID.String()+"/recordings", nil)
	req.Header.Set("X-Test-User-ID", strangerID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_StartRecording_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("HRec").SetKey(testCallsProjectKey("H")).Save(ctx)
	require.NoError(t, err)
	_, err = f.client.Channel.UpdateOne(f.channel).SetProjectID(proj.ID).Save(ctx)
	require.NoError(t, err)
	callEnt := f.createChannelCall(t, f.owner)

	svc, _ := newRecordingServiceWithMockEgress(t, f.repo)
	h := calls.ExportNewHandlerWithRecSvc(svc)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/recording/start", nil)
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

// A2-004: stranger → 403 на start recording.
func TestHandler_StartRecording_ForbiddenNonMember(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("HRecDeny").SetKey(testCallsProjectKey("D")).Save(ctx)
	require.NoError(t, err)
	_, err = f.client.Channel.UpdateOne(f.channel).SetProjectID(proj.ID).Save(ctx)
	require.NoError(t, err)
	callEnt := f.createChannelCall(t, f.owner)

	svc, _ := newRecordingServiceWithMockEgress(t, f.repo)
	h := calls.ExportNewHandlerWithRecSvc(svc)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/recording/start", nil)
	req.Header.Set("X-Test-User-ID", f.stranger.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_StopRecording_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("HStop").SetKey(testCallsProjectKey("P")).Save(ctx)
	require.NoError(t, err)
	callEnt := f.createChannelCall(t, f.owner)
	svc, _ := newRecordingServiceWithMockEgress(t, f.repo)

	rec, err := f.client.Recording.Create().
		SetEgressID("egress-hstop").
		SetStatus(recording.StatusRecording).
		SetCallID(callEnt.ID).
		SetProjectID(proj.ID).
		SetStartedBy(f.owner.ID).
		Save(ctx)
	require.NoError(t, err)

	h := calls.ExportNewHandlerWithRecSvc(svc)
	app := newHandlerValidationApp(h)
	body, _ := json.Marshal(map[string]string{"recording_id": rec.ID.String()})
	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/recording/stop", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// A2-004: stranger → 403 на stop recording.
func TestHandler_StopRecording_ForbiddenNonMember(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("HStopDeny").SetKey(testCallsProjectKey("Q")).Save(ctx)
	require.NoError(t, err)
	callEnt := f.createChannelCall(t, f.owner)
	svc, _ := newRecordingServiceWithMockEgress(t, f.repo)

	rec, err := f.client.Recording.Create().
		SetEgressID("egress-hstop-deny").
		SetStatus(recording.StatusRecording).
		SetCallID(callEnt.ID).
		SetProjectID(proj.ID).
		SetStartedBy(f.owner.ID).
		Save(ctx)
	require.NoError(t, err)

	h := calls.ExportNewHandlerWithRecSvc(svc)
	app := newHandlerValidationApp(h)
	body, _ := json.Marshal(map[string]string{"recording_id": rec.ID.String()})
	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/recording/stop", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.stranger.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_ListProjectRecordings_MemberSuccess(t *testing.T) {
	svc, projectID, client := newRecordingListFixture(t)
	ctx := context.Background()
	owner, err := client.User.Query().First(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).
		SetProjectID(projectID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	h := calls.ExportNewHandlerForTest(nil, svc, nil, config.LiveKitConfig{}, rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb)))
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/projects/"+projectID.String()+"/recordings?offset=0&limit=5", nil)
	req.Header.Set("X-Test-User-ID", owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
