package calls_test
import (
	"github.com/valerii/hub/backend/internal/calls"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/livekit/protocol/livekit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/internal/platform/config"
)

func mockLiveKitEgressHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		info := &livekit.EgressInfo{EgressId: "egress-mock-" + uuid.New().String()[:8]}
		data, err := proto.Marshal(info)
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/protobuf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}

func newRecordingServiceWithMockEgress(t *testing.T, repo *calls.EntRepository) (*calls.RecordingService, config.LiveKitConfig) {
	t.Helper()
	srv := httptest.NewServer(mockLiveKitEgressHandler(t))
	t.Cleanup(srv.Close)
	cfg := config.LiveKitConfig{
		URL:       srv.URL,
		APIKey:    "testkey",
		APISecret: "testsecretthatshouldbelongenough",
	}
	svc := calls.NewRecordingService(cfg, config.S3Config{
		Region: "us-east-1", AccessKey: "k", SecretKey: "secretkeysecret",
		Bucket: "recordings", Endpoint: "localhost:8333", UseSSL: false,
	}, repo, nil)
	return svc, cfg
}

func TestRecordingService_StartRecording_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("RecOK").SetKey(testCallsProjectKey("O")).Save(ctx)
	require.NoError(t, err)
	_, err = f.client.Channel.UpdateOne(f.channel).SetProjectID(proj.ID).Save(ctx)
	require.NoError(t, err)

	callEnt := f.createChannelCall(t, f.owner)
	svc, _ := newRecordingServiceWithMockEgress(t, f.repo)

	resp, err := svc.StartRecording(ctx, callEnt.ID, f.owner.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.EgressID)
	assert.Equal(t, string(recording.StatusRecording), resp.Status)
}

func TestRecordingService_StopRecording_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("RecStop").SetKey(testCallsProjectKey("S")).Save(ctx)
	require.NoError(t, err)

	callEnt := f.createChannelCall(t, f.owner)
	svc, _ := newRecordingServiceWithMockEgress(t, f.repo)

	rec, err := f.client.Recording.Create().
		SetEgressID("egress-stop-" + uuid.New().String()[:6]).
		SetStatus(recording.StatusRecording).
		SetCallID(callEnt.ID).
		SetProjectID(proj.ID).
		SetStartedBy(f.owner.ID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.StopRecording(ctx, callEnt.ID, rec.ID, f.owner.ID))

	updated, err := f.client.Recording.Get(ctx, rec.ID)
	require.NoError(t, err)
	assert.Equal(t, recording.StatusProcessing, updated.Status)
}

func TestCallChatService_EphemeralSendMessage(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	msg, err := calls.ExportCallServiceChat(f.svc).SendMessage(ctx, callEnt.ID.String(), f.owner.ID.String(), "Owner", "ephemeral", true)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "ephemeral", msg.Text)
}

func TestCallChatService_SendMessage_EndedCall(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)
	_, err := f.client.Call.UpdateOne(callEnt).SetStatus("ended").Save(ctx)
	require.NoError(t, err)

	_, err = calls.ExportCallServiceChat(f.svc).SendMessage(ctx, callEnt.ID.String(), f.owner.ID.String(), "Owner", "late", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "завершён")
}
