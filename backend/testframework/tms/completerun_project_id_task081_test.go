//go:build hubtest

package tms_test

// TASK-081 / A2W4-5-001: CompleteTestRun публикует project_id для project-scoped webhook.
import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/apigateway"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestCompleteTestRun_PublishesProjectID_TASK081 — реальный PublishJSON из CompleteTestRun
// содержит project_id; webhook проекта A получает событие, B — нет.
// Mutation: убрать project_id из payload CompleteTestRun → assert RED / fail-closed.
func TestCompleteTestRun_PublishesProjectID_TASK081(t *testing.T) {
	opts := &server.Options{Port: -1}
	ns := natsserver.RunServer(opts)
	t.Cleanup(ns.Shutdown)

	nc, err := natsclient.NewClient(context.Background(), ns.ClientURL())
	require.NoError(t, err)
	t.Cleanup(func() { nc.Close() })

	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nc)
	ctx := context.Background()
	projectA := uuid.New()
	projectB := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectA, uuid.New(), tms.CreateTestRunRequest{
		Title: "Прогон с webhook project_id",
	})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	_, err = svc.StartTestRun(ctx, projectA, runID)
	require.NoError(t, err)

	subject := tms.EventTestRunCompleted
	got := make(chan []byte, 1)
	sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		got <- append([]byte(nil), msg.Data...)
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = sub.Unsubscribe() })
	require.NoError(t, nc.FlushTimeout(time.Second))

	completed, err := svc.CompleteTestRun(ctx, projectA, runID)
	require.NoError(t, err)
	assert.Equal(t, "completed", completed.Status)

	var raw []byte
	select {
	case raw = <-got:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for tms.testrun.completed NATS event")
	}

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &payload))
	require.Contains(t, payload, "project_id", "CompleteTestRun must publish project_id for webhook scope")
	assert.Equal(t, projectA.String(), payload["project_id"])
	assert.Equal(t, runID.String(), payload["run_id"])

	pid, ok := apigateway.ExportExtractEventProjectID(subject, payload)
	require.True(t, ok, "extractEventProjectID must see project_id from CompleteTestRun payload")
	assert.Equal(t, projectA, pid)

	epA := uuid.New()
	epB := uuid.New()
	d := apigateway.NewWebhookDispatcher(&stubWebhookRepoForTMS{
		endpoints: []*ent.WebhookEndpoint{
			{
				ID: epA, URL: "https://example.com/a", Secret: "s",
				Events: []string{"test_run.completed"}, IsActive: true,
				Edges: ent.WebhookEndpointEdges{Project: &ent.Project{ID: projectA}},
			},
			{
				ID: epB, URL: "https://example.com/b", Secret: "s",
				Events: []string{"test_run.completed"}, IsActive: true,
				Edges: ent.WebhookEndpointEdges{Project: &ent.Project{ID: projectB}},
			},
		},
	}, nil)
	apigateway.ExportWebhookDispatcherInitJobs(d, 4)

	require.NoError(t, apigateway.ExportWebhookDispatcherDispatchEvent(
		d, context.Background(), subject, raw, payload))
	ids := apigateway.ExportWebhookDispatcherDrainJobEndpointIDs(d)
	require.Equal(t, []uuid.UUID{epA}, ids, "only project A webhook must receive completed run of A")

	// Без project_id — по-прежнему fail-closed (TASK-012), даже при events=["*"].
	dDrop := apigateway.NewWebhookDispatcher(&stubWebhookRepoForTMS{
		endpoints: []*ent.WebhookEndpoint{{
			ID: uuid.New(), URL: "https://example.com/drop", Secret: "s",
			Events: []string{"*", "test_run.completed"}, IsActive: true,
			Edges: ent.WebhookEndpointEdges{Project: &ent.Project{ID: projectA}},
		}},
	}, nil)
	apigateway.ExportWebhookDispatcherInitJobs(dDrop, 4)
	noProject := map[string]interface{}{"run_id": runID.String(), "title": completed.Title}
	require.NoError(t, apigateway.ExportWebhookDispatcherDispatchEvent(
		dDrop, context.Background(), subject, []byte(`{}`), noProject))
	assert.Empty(t, apigateway.ExportWebhookDispatcherDrainJobEndpointIDs(dDrop))
}

// stubWebhookRepoForTMS — минимальный stub ListActiveWebhookEndpoints для IT dispatch.
type stubWebhookRepoForTMS struct {
	endpoints []*ent.WebhookEndpoint
}

func (s *stubWebhookRepoForTMS) ListActiveWebhookEndpoints(context.Context) ([]*ent.WebhookEndpoint, error) {
	return s.endpoints, nil
}
func (s *stubWebhookRepoForTMS) CreateWebhookDelivery(context.Context, apigateway.CreateWebhookDeliveryInput) error {
	return nil
}
func (s *stubWebhookRepoForTMS) CreateWebhookEndpoint(context.Context, uuid.UUID, string, string, []string) (*ent.WebhookEndpoint, error) {
	return nil, nil
}
func (s *stubWebhookRepoForTMS) ListWebhookEndpointsByProject(context.Context, uuid.UUID) ([]*ent.WebhookEndpoint, error) {
	return nil, nil
}
func (s *stubWebhookRepoForTMS) UpdateWebhookEndpoint(context.Context, uuid.UUID, apigateway.UpdateWebhookEndpointInput) (*ent.WebhookEndpoint, error) {
	return nil, nil
}
func (s *stubWebhookRepoForTMS) DeleteWebhookEndpoint(context.Context, uuid.UUID) error {
	return nil
}
func (s *stubWebhookRepoForTMS) ListWebhookDeliveries(context.Context, uuid.UUID, int) ([]*ent.WebhookDelivery, error) {
	return nil, nil
}
func (s *stubWebhookRepoForTMS) WebhookEndpointProjectID(context.Context, uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}
