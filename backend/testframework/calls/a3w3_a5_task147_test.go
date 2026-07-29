//go:build hubtest

package calls_test

// TASK-147 / A3W3-A5-001..004: orphan room on DB fail, conflicts soft-delete,
// nil project_id → 400, title/guest_name rune caps.
import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/internal/calls"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// TestCreateCall_DBFail_DeletesLiveKitRoom_A3W3_A5_001
// Mutation: убрать DeleteRoom после persistChannelCall fail → deleteRoom=0 → FAIL.
func TestCreateCall_DBFail_DeletesLiveKitRoom_A3W3_A5_001(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	var deleteRoom atomic.Int32
	cfg := liveKitTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		if strings.Contains(r.URL.Path, "DeleteRoom") {
			deleteRoom.Add(1)
		}
		w.WriteHeader(http.StatusOK)
	})
	calls.ExportCallServiceSetRoom(f.svc, calls.NewCallRoomService(cfg))
	calls.ExportCallServiceSetCreateCall(f.svc,
		func(context.Context, string, uuid.UUID, uuid.UUID) (*ent.Call, error) {
			return nil, errors.New("db unavailable")
		})

	_, err := f.svc.CreateCall(ctx, f.channel.ID, f.owner.ID, "Owner")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось сохранить звонок")
	assert.GreaterOrEqual(t, deleteRoom.Load(), int32(1), "ожидался DeleteRoom после провала CreateCall DB")

	cnt, qErr := f.client.Call.Query().Where(call.ChannelID(f.channel.ID)).Count(ctx)
	require.NoError(t, qErr)
	assert.Equal(t, 0, cnt, "call row не должна остаться после DB fail")
}

// TestCreateInstantCall_DBFail_DeletesLiveKitRoom_A3W3_A5_001
// Mutation: убрать DeleteRoom после persistInstantCall fail → deleteRoom=0 → FAIL.
func TestCreateInstantCall_DBFail_DeletesLiveKitRoom_A3W3_A5_001(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	var deleteRoom atomic.Int32
	cfg := liveKitTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		if strings.Contains(r.URL.Path, "DeleteRoom") {
			deleteRoom.Add(1)
		}
		w.WriteHeader(http.StatusOK)
	})
	calls.ExportCallServiceSetRoom(f.svc, calls.NewCallRoomService(cfg))
	calls.ExportCallServiceSetCreateInstantCall(f.svc,
		func(context.Context, string, uuid.UUID, string, string) (*ent.Call, error) {
			return nil, errors.New("db unavailable")
		})

	_, err := f.svc.CreateInstantCall(ctx, f.owner.ID, "Owner", "Quick")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось сохранить звонок")
	assert.GreaterOrEqual(t, deleteRoom.Load(), int32(1), "ожидался DeleteRoom после провала CreateInstantCall DB")
}

// TestFindConflictingCalls_SkipsSoftDeletedChannel_A3W3_A5_002
// Mutation: убрать HasChannelWith(DeletedAtIsNil) → soft-deleted снова в conflicts → FAIL.
func TestFindConflictingCalls_SkipsSoftDeletedChannel_A3W3_A5_002(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("SoftConf").SetKey(testCallsProjectKey("S")).Save(ctx)
	require.NoError(t, err)

	at := time.Now().UTC().Add(3 * time.Hour)
	sc, err := f.repo.CreateScheduledCall(ctx, "SoftDeletedStandup", "", at, "", f.channel.ID, proj.ID, f.owner.ID, 30, nil, nil)
	require.NoError(t, err)

	windowStart := at.Add(-30 * time.Minute)
	windowEnd := at.Add(30 * time.Minute)

	before, err := f.repo.FindConflictingCalls(ctx, []uuid.UUID{f.owner.ID}, windowStart, windowEnd)
	require.NoError(t, err)
	require.NotEmpty(t, before)
	found := false
	for _, c := range before {
		if c.ID == sc.ID {
			found = true
			break
		}
	}
	require.True(t, found, "активный канал должен давать conflict")

	_, err = f.client.Channel.UpdateOneID(f.channel.ID).SetDeletedAt(time.Now().UTC()).Save(ctx)
	require.NoError(t, err)

	after, err := f.repo.FindConflictingCalls(ctx, []uuid.UUID{f.owner.ID}, windowStart, windowEnd)
	require.NoError(t, err)
	for _, c := range after {
		assert.NotEqual(t, sc.ID, c.ID, "soft-deleted channel не должен попадать в FindConflictingCalls")
		assert.NotEqual(t, "SoftDeletedStandup", c.Title)
	}
}

// TestGetProjectIDByChannel_NilProject_NotFound_A3W3_A5_003
// Mutation: вернуть (uuid.Nil, nil) → ErrorIs NotFound/ErrInvalidInput краснеет.
func TestGetProjectIDByChannel_NilProject_NotFound_A3W3_A5_003(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	_, err := f.repo.GetProjectIDByChannel(ctx, f.channel.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err), "nil ProjectID должен быть NotFound, не (Nil, nil)")
}

// TestStartRecording_NilProjectID_ErrInvalidInput_A3W3_A5_003
// Mutation: silent Nil из GetProjectIDByChannel без reject → FK/500 или ErrorIs краснеет.
func TestStartRecording_NilProjectID_ErrInvalidInput_A3W3_A5_003(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	// fixture channel без project_id
	callEnt := f.createChannelCall(t, f.owner)

	svc := calls.NewRecordingService(config.LiveKitConfig{
		URL: "http://localhost:7880", APIKey: "testkey", APISecret: "testsecretthatshouldbelongenough",
	}, config.S3Config{
		Region: "us-east-1", AccessKey: "k", SecretKey: "secretkeysecret",
		Bucket: "recordings", Endpoint: "localhost:8333",
	}, f.repo, nil)

	_, err := svc.StartRecording(ctx, callEnt.ID, f.owner.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, calls.ErrInvalidInput)
}

// TestHandler_StartRecording_NilProject_400_A3W3_A5_003
func TestHandler_StartRecording_NilProject_400_A3W3_A5_003(t *testing.T) {
	f := newCallsTestFixture(t)
	callEnt := f.createChannelCall(t, f.owner)

	recSvc := calls.NewRecordingService(config.LiveKitConfig{
		URL: "http://localhost:7880", APIKey: "testkey", APISecret: "testsecretthatshouldbelongenough",
	}, config.S3Config{
		Region: "us-east-1", AccessKey: "k", SecretKey: "secretkeysecret",
		Bucket: "recordings", Endpoint: "localhost:8333",
	}, f.repo, nil)
	h := calls.ExportNewHandlerWithRecSvc(recSvc)
	app := fiber.New()
	app.Use(testAuthMiddleware)
	app.Post("/calls/:callId/recording/start", h.StartRecording)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/recording/start", nil)
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestCreateInstantCall_TitleOverMaxRunes_A3W3_A5_004
// Mutation: убрать rune-check → Ent MaxLen 500 или oversized path → ErrorIs краснеет.
func TestCreateInstantCall_TitleOverMaxRunes_A3W3_A5_004(t *testing.T) {
	f := newCallsTestFixture(t)
	cfg := liveKitTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	calls.ExportCallServiceSetRoom(f.svc, calls.NewCallRoomService(cfg))

	// Cyrillic oversize via escape (Windows encoding-safe): rune-cap before Ent byte MaxLen.
	over := strings.Repeat("\u044F", calls.ExportMaxCallTitleRunes()+1)
	require.Equal(t, calls.ExportMaxCallTitleRunes()+1, utf8.RuneCountInString(over))

	_, err := f.svc.CreateInstantCall(context.Background(), f.owner.ID, "Owner", over)
	require.ErrorIs(t, err, calls.ErrInvalidInput)

	okTitle := "OK title boundary"
	resp, err := f.svc.CreateInstantCall(context.Background(), f.owner.ID, "Owner", okTitle)
	require.NoError(t, err)
	assert.Equal(t, okTitle, resp.Title)
}

// TestHandler_CreateInstantCall_TitleOverMax_400_A3W3_A5_004
func TestHandler_CreateInstantCall_TitleOverMax_400_A3W3_A5_004(t *testing.T) {
	f := newCallsTestFixture(t)
	cfg := liveKitTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	calls.ExportCallServiceSetRoom(f.svc, calls.NewCallRoomService(cfg))
	app := newCallsHandlerApp(f)

	over := strings.Repeat("x", calls.ExportMaxCallTitleRunes()+1)
	body := []byte(`{"title":"` + over + `"}`)
	req := httptest.NewRequest(fiber.MethodPost, "/calls/instant", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	req.Header.Set("X-Test-User-Name", "Owner")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestGuestJoin_GuestNameOverMaxRunes_A3W3_A5_004
// Mutation: убрать cap → oversized JWT path / ErrorIs краснеет.
func TestGuestJoin_GuestNameOverMaxRunes_A3W3_A5_004(t *testing.T) {
	f := newCallsTestFixture(t)
	code := "guestnm01ab"
	f.createInstantCall(t, f.owner, code)
	cfg := liveKitTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	calls.ExportCallServiceSetRoom(f.svc, calls.NewCallRoomService(cfg))

	over := strings.Repeat("\u0413", calls.ExportMaxGuestNameRunes()+1)
	_, err := f.svc.GuestJoin(context.Background(), code, over)
	require.ErrorIs(t, err, calls.ErrInvalidInput)

	ok := strings.Repeat("\u0413", calls.ExportMaxGuestNameRunes())
	resp, err := f.svc.GuestJoin(context.Background(), code, ok)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
}

// TestHandler_GuestJoin_GuestNameOverMax_400_A3W3_A5_004
func TestHandler_GuestJoin_GuestNameOverMax_400_A3W3_A5_004(t *testing.T) {
	f := newCallsTestFixture(t)
	code := "guestnm02cd"
	f.createInstantCall(t, f.owner, code)
	app := newCallsHandlerApp(f)

	over := strings.Repeat("G", calls.ExportMaxGuestNameRunes()+1)
	body := []byte(`{"invite_code":"` + code + `","guest_name":"` + over + `"}`)
	req := httptest.NewRequest(fiber.MethodPost, "/calls/guest-join", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
