package audit

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/recording"
)

func testProjectKey(prefix string) string {
	suffix := strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", ""))[:5]
	return prefix + suffix
}

func newRetentionTestClient(t *testing.T) *ent.Client {
	t.Helper()
	dsn := "file:retention_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func seedRecordingFixture(t *testing.T, client *ent.Client, createdAt time.Time) *ent.Recording {
	t.Helper()
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@example.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Retention Project").
		SetKey(testProjectKey("R")).
		Save(ctx)
	require.NoError(t, err)

	c, err := client.Call.Create().
		SetRoomName("room-" + uuid.New().String()[:8]).
		SetStatus(call.StatusActive).
		SetCreatedBy(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	rec, err := client.Recording.Create().
		SetEgressID("egress-" + uuid.New().String()[:8]).
		SetStatus(recording.StatusReady).
		SetCallID(c.ID).
		SetProjectID(proj.ID).
		SetStartedBy(owner.ID).
		SetFilePath("recordings/" + proj.ID.String() + "/" + c.ID.String() + ".mp4").
		SetCreatedAt(createdAt).
		Save(ctx)
	require.NoError(t, err)
	return rec
}

func TestRetentionService_EnsureIndexes_NilMongo(t *testing.T) {
	svc := NewRetentionService(newRetentionTestClient(t), nil, nil, "")
	err := svc.EnsureIndexes(context.Background())
	require.NoError(t, err)
}

func TestRetentionService_PurgeOldMessages_NilMongo(t *testing.T) {
	svc := NewRetentionService(newRetentionTestClient(t), nil, nil, "")
	err := svc.purgeOldMessages(context.Background(), 30)
	require.NoError(t, err)
}

func TestRetentionService_GetRetentionDays_UsesSystemSetting(t *testing.T) {
	client := newRetentionTestClient(t)
	ctx := context.Background()
	svc := NewRetentionService(client, nil, nil, "")

	_, err := client.SystemSetting.Create().
		SetKey(RetentionCallRecordingsDays).
		SetValue("90").
		SetDescription("test").
		Save(ctx)
	require.NoError(t, err)

	got := svc.getRetentionDays(ctx, RetentionCallRecordingsDays, defaultCallRecordingRetention)
	assert.Equal(t, 90, got)
}

func TestRetentionService_GetRetentionDays_InvalidValueFallsBack(t *testing.T) {
	client := newRetentionTestClient(t)
	ctx := context.Background()
	svc := NewRetentionService(client, nil, nil, "")

	_, err := client.SystemSetting.Create().
		SetKey(RetentionAllureResultsDays).
		SetValue("not-a-number").
		SetDescription("test").
		Save(ctx)
	require.NoError(t, err)

	got := svc.getRetentionDays(ctx, RetentionAllureResultsDays, defaultAllureRetentionDays)
	assert.Equal(t, defaultAllureRetentionDays, got)
}

func TestRetentionService_GetRetentionDays_MissingSettingFallsBack(t *testing.T) {
	svc := NewRetentionService(newRetentionTestClient(t), nil, nil, "")
	got := svc.getRetentionDays(context.Background(), RetentionChatMessagesDays, defaultChatRetentionDays)
	assert.Equal(t, defaultChatRetentionDays, got)
}

func TestRetentionService_PurgeOldCallRecordings_WithoutS3(t *testing.T) {
	client := newRetentionTestClient(t)
	ctx := context.Background()
	svc := NewRetentionService(client, nil, nil, "")

	old := seedRecordingFixture(t, client, time.Now().Add(-200*24*time.Hour))
	fresh := seedRecordingFixture(t, client, time.Now().Add(-1*24*time.Hour))

	err := svc.purgeOldCallRecordings(ctx, 180)
	require.NoError(t, err)

	_, err = client.Recording.Get(ctx, old.ID)
	assert.True(t, ent.IsNotFound(err), "expired recording should be deleted")

	still, err := client.Recording.Get(ctx, fresh.ID)
	require.NoError(t, err)
	assert.Equal(t, fresh.ID, still.ID)
}

func TestRetentionService_PurgeOldCallRecordings_NoExpiredRows(t *testing.T) {
	client := newRetentionTestClient(t)
	ctx := context.Background()
	svc := NewRetentionService(client, nil, nil, "")

	seedRecordingFixture(t, client, time.Now().Add(-5*24*time.Hour))

	err := svc.purgeOldCallRecordings(ctx, 180)
	require.NoError(t, err)

	count, err := client.Recording.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRetentionService_RunRetentionCleanup_NilMongoSucceeds(t *testing.T) {
	client := newRetentionTestClient(t)
	ctx := context.Background()
	svc := NewRetentionService(client, nil, nil, "")

	seedRecordingFixture(t, client, time.Now().Add(-200*24*time.Hour))

	err := svc.RunRetentionCleanup(ctx)
	require.NoError(t, err)

	count, err := client.Recording.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
