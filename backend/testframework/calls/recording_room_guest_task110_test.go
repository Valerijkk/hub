package calls_test

// TASK-110 / A2W3-5-001..003: orphan egress StopEgress, finalizeCreate rollback, invite entropy.
import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/livekit/protocol/livekit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/internal/calls"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// TestStartRecording_CreateRecordingFail_StopsEgress_A2W3_5_001
// Mutation: убрать compensateOrphanEgress → StopEgress не вызывается → FAIL.
func TestStartRecording_CreateRecordingFail_StopsEgress_A2W3_5_001(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("OrphanEgress").SetKey(testCallsProjectKey("G")).Save(ctx)
	require.NoError(t, err)
	_, err = f.client.Channel.UpdateOne(f.channel).SetProjectID(proj.ID).Save(ctx)
	require.NoError(t, err)
	callEnt := f.createChannelCall(t, f.owner)

	var stopCalls atomic.Int32
	egressID := "egress-orphan-" + uuid.New().String()[:8]
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		if strings.Contains(r.URL.Path, "StopEgress") {
			stopCalls.Add(1)
			info := &livekit.EgressInfo{EgressId: egressID}
			data, mErr := proto.Marshal(info)
			require.NoError(t, mErr)
			w.Header().Set("Content-Type", "application/protobuf")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
			return
		}
		info := &livekit.EgressInfo{EgressId: egressID}
		data, mErr := proto.Marshal(info)
		require.NoError(t, mErr)
		w.Header().Set("Content-Type", "application/protobuf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)

	svc := calls.NewRecordingService(config.LiveKitConfig{
		URL: srv.URL, APIKey: "testkey", APISecret: "testsecretthatshouldbelongenough",
	}, config.S3Config{
		Region: "us-east-1", AccessKey: "k", SecretKey: "secretkeysecret",
		Bucket: "recordings", Endpoint: "localhost:8333",
	}, f.repo, nil)

	calls.ExportRecordingServiceSetCreateRecording(svc,
		func(context.Context, string, uuid.UUID, uuid.UUID, uuid.UUID, string) (*ent.Recording, error) {
			return nil, errors.New("db unavailable")
		})

	_, err = svc.StartRecording(ctx, callEnt.ID, f.owner.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось сохранить запись")
	assert.GreaterOrEqual(t, stopCalls.Load(), int32(1), "ожидался StopEgress после провала CreateRecording")
}

// TestFinalizeCreate_AddParticipantFail_RollsBackRoomAndCall_A2W3_5_002
// Mutation: вернуть slog-only при AddParticipant fail → call остаётся active + DeleteRoom=0 → FAIL.
func TestFinalizeCreate_AddParticipantFail_RollsBackRoomAndCall_A2W3_5_002(t *testing.T) {
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
	room := calls.NewCallRoomService(cfg)

	callEnt, err := f.repo.CreateCall(ctx, "call-rollback-"+uuid.New().String()[:8], f.channel.ID, f.owner.ID)
	require.NoError(t, err)

	// Несуществующий user → FK fail на AddParticipant → rollback room+call.
	ghost := uuid.New()
	_, err = calls.ExportFinalizeCreate(ctx, room, f.repo, callEnt, ghost, "Ghost")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не удалось добавить создателя")
	assert.GreaterOrEqual(t, deleteRoom.Load(), int32(1), "ожидался DeleteRoom при rollback")

	_, getErr := f.client.Call.Get(ctx, callEnt.ID)
	require.True(t, ent.IsNotFound(getErr), "звонок должен быть hard-delete после rollback")
}

// TestCreateInstantCall_InviteCodeBase62Entropy_A2W3_5_003
// Mutation: вернуть uuid-hex[:12] → либо длина/алфавит не base62-полный, либо только hex → FAIL на charset assert.
func TestCreateInstantCall_InviteCodeBase62Entropy_A2W3_5_003(t *testing.T) {
	f := newCallsTestFixture(t)
	cfg := liveKitTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	calls.ExportCallServiceSetRoom(f.svc, calls.NewCallRoomService(cfg))

	resp, err := f.svc.CreateInstantCall(context.Background(), f.owner.ID, "Owner", "Entropy")
	require.NoError(t, err)
	require.Len(t, resp.InviteCode, calls.ExportInviteCodeLength())
	assert.Regexp(t, `^[0-9A-Za-z]{12}$`, resp.InviteCode)

	// Не усечённый UUID-hex: хотя бы раз за серию встретится A-Z / a-z вне a-f
	// (вероятностно; crypto/rand base62 почти наверняка даст не-hex за 20 попыток).
	seenNonHexLetter := false
	for i := 0; i < 20; i++ {
		code, gErr := calls.ExportGenerateInviteCode()
		require.NoError(t, gErr)
		require.Len(t, code, 12)
		require.Regexp(t, `^[0-9A-Za-z]{12}$`, code)
		if strings.ContainsAny(code, "ghijklmnopqrstuvwxyzGHIJKLMNOPQRSTUVWXYZ") {
			seenNonHexLetter = true
			break
		}
	}
	assert.True(t, seenNonHexLetter, "invite должен использовать base62, не только hex")

	stored, err := f.client.Call.Query().Where(call.ID(resp.ID)).Only(context.Background())
	require.NoError(t, err)
	assert.Equal(t, resp.InviteCode, stored.InviteCode)
}
