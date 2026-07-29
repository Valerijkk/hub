//go:build hubtest

package tracker_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/timelog"
	"github.com/valerii/hub/backend/ent/user"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestAutomationConsumer_AckKnownNoOpSubjects_A3W1_A3_001 — updated/comment/link
// Ack без Nak и без роста unknown-metric (нет redelivery storm).
// Mutation: убрать known-ignore / вернуть default→Nak → naks>0 FAIL.
func TestAutomationConsumer_AckKnownNoOpSubjects_A3W1_A3_001(t *testing.T) {
	repo := newMockRepo()
	engine := tracker.NewAutomationEngine(repo, nil, nil)
	c := tracker.NewAutomationConsumer(nil, engine)
	pid := uuid.New().String()
	before := tracker.AutomationUnknownSubjectTotal()

	for _, kind := range []string{"updated", "comment_created", "link_created"} {
		msg := &countingAutomationJSMsg{
			subject: "project." + pid + ".issue." + kind,
			data:    []byte(`{}`),
		}
		tracker.ExportAutomationConsumerProcessMessage(c, context.Background(), msg)
		assert.Equal(t, 1, msg.acks, "kind %s must Ack", kind)
		assert.Equal(t, 0, msg.naks, "kind %s must not Nak", kind)
	}
	assert.Equal(t, before, tracker.AutomationUnknownSubjectTotal(), "known no-op must not bump unknown metric")

	// garbage по-прежнему Nak + metric (TASK-088 не ослаблен).
	garbage := &countingAutomationJSMsg{
		subject: "project." + pid + ".issue.deleted",
		data:    []byte(`{}`),
	}
	tracker.ExportAutomationConsumerProcessMessage(c, context.Background(), garbage)
	assert.Equal(t, 1, garbage.naks)
	assert.Equal(t, 0, garbage.acks)
	assert.Equal(t, before+1, tracker.AutomationUnknownSubjectTotal())
}

// TestAutomationConsumer_FilterSubjects_A3W1_A3_001 — durable consumer подписан
// только на moved/created/assigned (не на широкий issue.>).
// Mutation: вернуть FilterSubject issue.> → assert FAIL.
func TestAutomationConsumer_FilterSubjects_A3W1_A3_001(t *testing.T) {
	want := []string{
		"project.*.issue.moved",
		"project.*.issue.created",
		"project.*.issue.assigned",
	}
	assert.Equal(t, want, tracker.AutomationFilterSubjects)

	nc := startAutomationJetStream(t)
	engine := tracker.NewAutomationEngine(newMockRepo(), nil, nil)
	c := tracker.NewAutomationConsumer(nc, engine)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, c.Start(ctx))
	defer c.Stop()

	// дать consumer создаться
	deadline := time.Now().Add(2 * time.Second)
	var got []string
	for time.Now().Before(deadline) {
		stream, err := nc.JetStream.Stream(ctx, natsclient.StreamName)
		if err != nil {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		cons, err := stream.Consumer(ctx, "automation-dispatcher")
		if err != nil {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		info, err := cons.Info(ctx)
		if err != nil {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		got = info.Config.FilterSubjects
		break
	}
	require.Equal(t, want, got, "JetStream consumer must use AutomationFilterSubjects")
}

// TestAutomationConsumer_NakOnPostMessageFail_A3W1_A3_002 — action Post fail → Nak, не Ack.
// Mutation: executeRule снова return nil / slog-only → acks==1 FAIL.
func TestAutomationConsumer_NakOnPostMessageFail_A3W1_A3_002(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:a3w1_postfail_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("PostNak").SetKey("PN").Save(ctx)
	require.NoError(t, err)
	u, err := client.User.Create().SetEmail("pn@test.com").Save(ctx)
	require.NoError(t, err)
	ch, err := client.Channel.Create().SetName("err").SetChannelType("public").SetProjectID(proj.ID).SetCreatedBy(u.ID).Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	_, err = repo.CreateAutomationRule(ctx, proj.ID, u.ID, tracker.CreateAutomationRuleInput{
		Name: "Post", Enabled: true, TriggerType: tracker.TriggerIssueCreated,
		ActionType:   tracker.ActionPostToChannel,
		ActionConfig: map[string]interface{}{"channel_id": ch.ID.String()},
	})
	require.NoError(t, err)

	engine := tracker.NewAutomationEngine(repo, stubMessengerFail{}, nil)
	c := tracker.NewAutomationConsumer(nil, engine)

	payload, err := json.Marshal(tracker.IssueCreatedEvent{
		IssueKey: "PN-1", IssueID: uuid.New().String(), ProjectID: proj.ID.String(),
		CreatedByID: u.ID.String(), IssueType: "task", Title: "x",
	})
	require.NoError(t, err)

	msg := &countingAutomationJSMsg{
		subject: "project." + proj.ID.String() + ".issue.created",
		data:    payload,
	}
	tracker.ExportAutomationConsumerProcessMessage(c, ctx, msg)
	assert.Equal(t, 1, msg.naks, "PostAutomationMessage fail must Nak")
	assert.Equal(t, 0, msg.acks, "action fail must not Ack")
}

// TestEntRepository_StartTimer_SoftDeleted_A3W1_A3_003 — trash issue → NotFound, нет open timer.
// Mutation: убрать DeletedAtIsNil с StartTimer lock → таймер создаётся FAIL.
func TestEntRepository_StartTimer_SoftDeleted_A3W1_A3_003(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	created, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Trash timer", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)
	require.NoError(t, env.repo.DeleteIssue(env.ctx, created.ID))

	_, err = env.repo.StartTimer(env.ctx, created.ID, env.userID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "expected AppError, got %T: %v", err, err)
	assert.Equal(t, 404, appErr.Code)

	openCount, err := env.client.TimeLog.Query().
		Where(
			timelog.HasIssueWith(issue.IDEQ(created.ID)),
			timelog.HasUserWith(user.IDEQ(env.userID)),
			timelog.EndedAtIsNil(),
			timelog.StartedAtNotNil(),
		).
		Count(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, openCount, "soft-deleted issue must not get an open timer")
}

// TestStartTimer_LockUsesDeletedAtIsNil_Source_A3W1_A3_003 — source guard.
func TestStartTimer_LockUsesDeletedAtIsNil_Source_A3W1_A3_003(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	srcPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "tracker", "repo_time.go")
	raw, err := os.ReadFile(srcPath)
	require.NoError(t, err)
	src := string(raw)

	fnStart := strings.Index(src, "func (r *EntRepository) StartTimer")
	require.GreaterOrEqual(t, fnStart, 0)
	fnEnd := strings.Index(src[fnStart:], "\nfunc (r *EntRepository) StopTimer")
	require.GreaterOrEqual(t, fnEnd, 0)
	chunk := src[fnStart : fnStart+fnEnd]
	assert.Contains(t, chunk, "DeletedAtIsNil", "A3W1-A3-003: StartTimer lock must Where(DeletedAtIsNil)")
}
