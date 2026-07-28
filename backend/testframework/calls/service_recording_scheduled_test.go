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
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/ent/scheduledcall"
	"github.com/valerii/hub/backend/internal/platform/config"
	"github.com/valerii/hub/backend/internal/rbac"
)

func newRecordingListFixture(t *testing.T) (*calls.RecordingService, uuid.UUID, *ent.Client) {
	t.Helper()
	dsn := "file:reclist_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@example.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Recordings").
		SetKey(testCallsProjectKey("R")).
		Save(ctx)
	require.NoError(t, err)

	c, err := client.Call.Create().
		SetRoomName("room-list").
		SetStatus(call.StatusActive).
		SetCreatedBy(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		_, err = client.Recording.Create().
			SetEgressID("egress-list-" + uuid.New().String()[:6]).
			SetStatus(recording.StatusReady).
			SetCallID(c.ID).
			SetProjectID(proj.ID).
			SetStartedBy(owner.ID).
			SetFilePath("recordings/proj/call-" + uuid.New().String()[:8] + ".mp4").
			SetCreatedAt(time.Now().Add(-time.Duration(i) * time.Hour)).
			Save(ctx)
		require.NoError(t, err)
	}

	svc := calls.ExportNewRecordingServiceForTest(calls.NewEntRepository(client), nil, config.S3Config{})
	return svc, proj.ID, client
}

func TestRecordingService_ListProjectRecordings_Pagination(t *testing.T) {
	svc, projectID, _ := newRecordingListFixture(t)
	ctx := context.Background()

	page1, total, err := svc.ListProjectRecordings(ctx, projectID, 0, 2)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, page1, 2)

	page2, total2, err := svc.ListProjectRecordings(ctx, projectID, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, 3, total2)
	assert.Len(t, page2, 1)
}

func TestRecordingService_RecordingProjectID(t *testing.T) {
	svc, projectID, client := newRecordingListFixture(t)
	ctx := context.Background()

	rec, err := client.Recording.Query().First(ctx)
	require.NoError(t, err)

	got, err := svc.RecordingProjectID(ctx, rec.ID)
	require.NoError(t, err)
	assert.Equal(t, projectID, got)
}

func TestRecordingService_GetRecordingDownloadURL_NoPresign(t *testing.T) {
	svc, _, client := newRecordingListFixture(t)
	ctx := context.Background()

	rec, err := client.Recording.Query().First(ctx)
	require.NoError(t, err)

	_, err = svc.GetRecordingDownloadURL(ctx, rec.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "presign")
}

func TestRecordingService_GetRecordingDownloadURL_NotReady(t *testing.T) {
	svc, _, client := newRecordingListFixture(t)
	ctx := context.Background()

	base, err := client.Recording.Query().First(ctx)
	require.NoError(t, err)

	rec, err := client.Recording.Create().
		SetEgressID("egress-processing").
		SetStatus(recording.StatusProcessing).
		SetCallID(base.CallID).
		SetProjectID(base.ProjectID).
		SetStartedBy(base.StartedBy).
		Save(ctx)
	require.NoError(t, err)

	_, err = svc.GetRecordingDownloadURL(ctx, rec.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не готова")
}

func TestRecordingService_GetRecordingDownloadURL_Success(t *testing.T) {
	svc, _, client := newRecordingListFixture(t)
	ctx := context.Background()

	rec, err := client.Recording.Query().First(ctx)
	require.NoError(t, err)

	svcWithPresign := calls.NewRecordingService(
		config.LiveKitConfig{URL: "http://localhost:7880", APIKey: "k", APISecret: "testsecretthatshouldbelongenough"},
		config.S3Config{
			Region: "us-east-1", AccessKey: "accesskey", SecretKey: "secretkeysecret",
			Bucket: "recordings", Endpoint: "localhost:8333", UseSSL: false,
		},
		calls.ExportRecordingServiceRepo(svc),
		nil,
	)
	require.NotNil(t, calls.ExportRecordingServicePresignClient(svcWithPresign))

	resp, err := svcWithPresign.GetRecordingDownloadURL(ctx, rec.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.URL)
	assert.False(t, resp.ExpiresAt.IsZero())
}

func TestHandler_GetRecordingDownloadURL_ReadyRecording(t *testing.T) {
	svc, projID, client := newRecordingListFixture(t)
	ctx := context.Background()
	owner, err := client.User.Query().First(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).
		SetProjectID(projID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))

	recSvc := calls.NewRecordingService(
		config.LiveKitConfig{URL: "http://localhost:7880", APIKey: "k", APISecret: "testsecretthatshouldbelongenough"},
		config.S3Config{
			Region: "us-east-1", AccessKey: "accesskey", SecretKey: "secretkeysecret",
			Bucket: "recordings", Endpoint: "localhost:8333", UseSSL: false,
		},
		calls.ExportRecordingServiceRepo(svc),
		nil,
	)
	rec, err := client.Recording.Query().First(ctx)
	require.NoError(t, err)

	h := calls.ExportNewHandlerWithRecSvcAndRBAC(recSvc, rbacSvc)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/recordings/"+rec.ID.String()+"/download", nil)
	req.Header.Set("X-Test-User-ID", owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestNewRecordingService_Constructor(t *testing.T) {
	f := newCallsTestFixture(t)
	svc := calls.NewRecordingService(
		calls.ExportCallRoomServiceCfg(calls.ExportCallServiceRoom(f.svc)),
		config.S3Config{Region: "us-east-1", AccessKey: "k", SecretKey: "s", Bucket: "b", Endpoint: "localhost:8333"},
		f.repo,
		nil,
	)
	require.NotNil(t, svc)
}

func TestScheduledCallService_CreateScheduledCall_Validation(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("Sched Val").SetKey("SVL").Save(ctx)
	require.NoError(t, err)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	_, err = scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "название")

	_, err = scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "No channel",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "channel_id")
}

func TestScheduledCallService_CheckConflicts_NoParticipants(t *testing.T) {
	svc := calls.ExportNewScheduledCallServiceForTest(calls.NewEntRepository(newScheduledTestClient(t)))
	resp, err := svc.CheckConflicts(context.Background(), uuid.Nil, time.Now(), 30, nil)
	require.NoError(t, err)
	assert.False(t, resp.HasConflict)
	assert.Empty(t, resp.Conflicts)
}

func TestToScheduledCallResponse_MapsFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	channelID := uuid.New()
	projectID := uuid.New()
	createdBy := uuid.New()

	sc := &ent.ScheduledCall{
		ID:          uuid.New(),
		Title:       "Standup",
		Description: "Daily",
		ScheduledAt: now,
		Rrule:       "FREQ=DAILY",
		ChannelID:   channelID,
		ProjectID:   projectID,
		CreatedBy:   createdBy,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	resp := calls.ExportToScheduledCallResponse(sc)
	require.NotNil(t, resp)
	assert.Equal(t, sc.ID, resp.ID)
	assert.Equal(t, "Standup", resp.Title)
	assert.Equal(t, "Daily", resp.Description)
	assert.Equal(t, now, resp.ScheduledAt)
	assert.Equal(t, "FREQ=DAILY", resp.RRule)
	assert.Equal(t, channelID, resp.ChannelID)
	assert.Equal(t, projectID, resp.ProjectID)
	assert.Equal(t, createdBy, resp.CreatedBy)
	assert.True(t, resp.IsActive)
}

func newScheduledTestClient(t *testing.T) *ent.Client {
	t.Helper()
	dsn := "file:sched_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestToCallResponse_MapsFields(t *testing.T) {
	channelID := uuid.New()
	started := time.Now().UTC()
	callEnt := &ent.Call{
		ID:               uuid.New(),
		RoomName:         "call-room",
		ChannelID:        &channelID,
		Status:           call.StatusActive,
		CreatedBy:        uuid.New(),
		ParticipantCount: 3,
		InviteCode:       "invite01",
		Title:            "Sync",
		StartedAt:        &started,
	}

	resp := calls.ExportToCallResponse(callEnt, "jwt-token")
	require.NotNil(t, resp)
	assert.Equal(t, callEnt.ID, resp.ID)
	assert.Equal(t, "call-room", resp.RoomName)
	assert.Equal(t, channelID, *resp.ChannelID)
	assert.Equal(t, "active", resp.Status)
	assert.Equal(t, 3, resp.ParticipantCount)
	assert.Equal(t, "jwt-token", resp.Token)
	assert.Equal(t, "invite01", resp.InviteCode)
	assert.Equal(t, "Sync", resp.Title)
	require.NotNil(t, resp.StartedAt)
}

// Ensure scheduledcall import is used for compile-time schema reference.
var _ = scheduledcall.Table
