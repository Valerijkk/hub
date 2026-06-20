package calls

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/internal/platform/config"
)

func TestRecordingService_StartRecording_InstantCallUnsupported(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createInstantCall(t, f.owner, "recinst1")

	svc := &RecordingService{
		repo: f.svc.repo,
		s3Cfg: config.S3Config{Region: "us-east-1", Bucket: "b"},
	}

	_, err := svc.StartRecording(ctx, callEnt.ID, f.owner.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "instant")
}

func TestRecordingService_StartRecording_MaxActiveLimit(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("RecLim").SetKey("RLM").Save(ctx)
	require.NoError(t, err)
	_, err = f.client.Channel.UpdateOne(f.channel).SetProjectID(proj.ID).Save(ctx)
	require.NoError(t, err)

	callEnt := f.createChannelCall(t, f.owner)
	svc := &RecordingService{repo: f.svc.repo, s3Cfg: config.S3Config{Region: "us-east-1", Bucket: "b"}}

	for i := 0; i < maxActiveRecordingsPerCall; i++ {
		_, err = f.client.Recording.Create().
			SetEgressID("egress-limit-" + uuid.New().String()[:6]).
			SetStatus(recording.StatusRecording).
			SetCallID(callEnt.ID).
			SetProjectID(proj.ID).
			SetStartedBy(f.owner.ID).
			Save(ctx)
		require.NoError(t, err)
	}

	_, err = svc.StartRecording(ctx, callEnt.ID, f.owner.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "лимит")
}

func TestRecordingService_StartRecording_EndedCall(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)
	_, err := f.client.Call.UpdateOne(callEnt).SetStatus("ended").Save(ctx)
	require.NoError(t, err)

	svc := &RecordingService{repo: f.svc.repo}
	_, err = svc.StartRecording(ctx, callEnt.ID, f.owner.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "завершён")
}

func TestRecordingService_StopRecording_NotFound(t *testing.T) {
	f := newCallsTestFixture(t)
	svc := &RecordingService{repo: f.svc.repo}
	err := svc.StopRecording(context.Background(), uuid.New())
	require.Error(t, err)
}

func TestNewRecordingService_WithS3Endpoint(t *testing.T) {
	f := newCallsTestFixture(t)
	svc := NewRecordingService(
		f.svc.room.cfg,
		config.S3Config{
			Region: "us-east-1", AccessKey: "k", SecretKey: "s",
			Bucket: "hub", Endpoint: "localhost:8333", UseSSL: false,
		},
		f.svc.repo,
		nil,
	)
	require.NotNil(t, svc)
}
