package apigateway_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/apigateway"
)

// TASK-011: UI short name issue.created матчит JetStream subject project.<uuid>.issue.created.
func TestContainsEvent_UIShortName_MatchesProjectSubject(t *testing.T) {
	t.Parallel()
	events := []string{"issue.created"}
	projectID := uuid.New()
	subject := "project." + projectID.String() + ".issue.created"

	assert.True(t, apigateway.ExportContainsEvent(events, subject))
	assert.True(t, apigateway.ExportContainsEvent(events, "issue.created")) // exact short
	assert.False(t, apigateway.ExportContainsEvent(events, "project."+projectID.String()+".issue.updated"))
}

func TestContainsEvent_Aliases_TMSAndSprint(t *testing.T) {
	t.Parallel()

	assert.True(t, apigateway.ExportContainsEvent(
		[]string{"test_run.completed"}, "tms.testrun.completed"))
	assert.True(t, apigateway.ExportContainsEvent(
		[]string{"test_run.completed"}, "tms.run.completed."+uuid.New().String()))
	assert.False(t, apigateway.ExportContainsEvent(
		[]string{"test_run.completed"}, "tms.testrun.created"))

	assert.True(t, apigateway.ExportContainsEvent(
		[]string{"sprint.completed"}, "sprint.completed"))
	assert.True(t, apigateway.ExportContainsEvent(
		[]string{"sprint.completed"}, "sprint."+uuid.New().String()+".completed"))

	assert.True(t, apigateway.ExportContainsEvent(
		[]string{"issue.status_changed"},
		"project."+uuid.New().String()+".issue.moved"))
}

func TestContainsEvent_IssueUpdated_Suffix(t *testing.T) {
	t.Parallel()
	events := []string{"issue.updated", "issue.created"}
	pid := uuid.New().String()
	assert.True(t, apigateway.ExportContainsEvent(events, "project."+pid+".issue.updated"))
	assert.True(t, apigateway.ExportContainsEvent(events, "tracker."+pid+".issue.created"))
}

// TASK-011 + enqueue: подписка issue.created → job при subject project.<uuid>.issue.created.
func TestWebhookDispatcher_dispatchEvent_EnqueuesOnUIShortMatch(t *testing.T) {
	projectID := uuid.New()
	epID := uuid.New()
	repo := &stubWebhookRepo{
		endpoints: []*ent.WebhookEndpoint{{
			ID: epID, URL: "https://example.com/hook", Secret: "s",
			Events: []string{"issue.created"}, IsActive: true,
			Edges: ent.WebhookEndpointEdges{Project: &ent.Project{ID: projectID}},
		}},
	}
	d := apigateway.NewWebhookDispatcher(repo, nil)
	apigateway.ExportWebhookDispatcherInitJobs(d, 4)

	subject := "project." + projectID.String() + ".issue.created"
	require.NoError(t, apigateway.ExportWebhookDispatcherDispatchEvent(
		d, context.Background(), subject, []byte(`{}`), map[string]interface{}{}))

	ids := apigateway.ExportWebhookDispatcherDrainJobEndpointIDs(d)
	require.Equal(t, []uuid.UUID{epID}, ids)
}

// TASK-012: эндпоинт проекта A не получает событие проекта B (в т.ч. при *).
func TestWebhookDispatcher_dispatchEvent_ProjectScope_Denial(t *testing.T) {
	projectA := uuid.New()
	projectB := uuid.New()
	epA := uuid.New()
	epB := uuid.New()

	repo := &stubWebhookRepo{
		endpoints: []*ent.WebhookEndpoint{
			{
				ID: epA, URL: "https://example.com/a", Secret: "s",
				Events: []string{"*"}, IsActive: true,
				Edges: ent.WebhookEndpointEdges{Project: &ent.Project{ID: projectA}},
			},
			{
				ID: epB, URL: "https://example.com/b", Secret: "s",
				Events: []string{"*"}, IsActive: true,
				Edges: ent.WebhookEndpointEdges{Project: &ent.Project{ID: projectB}},
			},
		},
	}
	d := apigateway.NewWebhookDispatcher(repo, nil)
	apigateway.ExportWebhookDispatcherInitJobs(d, 8)

	subjectB := "project." + projectB.String() + ".issue.created"
	require.NoError(t, apigateway.ExportWebhookDispatcherDispatchEvent(
		d, context.Background(), subjectB, []byte(`{}`), map[string]interface{}{}))

	ids := apigateway.ExportWebhookDispatcherDrainJobEndpointIDs(d)
	require.Equal(t, []uuid.UUID{epB}, ids, "only project B endpoint must receive event for B")
}

// TASK-012 fail-closed: subject без project UUID и без project_id в payload
// не ставит jobs даже при events=["*"].
func TestWebhookDispatcher_dispatchEvent_NoProject_FailClosed(t *testing.T) {
	projectID := uuid.New()
	repo := &stubWebhookRepo{
		endpoints: []*ent.WebhookEndpoint{{
			ID: uuid.New(), URL: "https://example.com/hook", Secret: "s",
			Events: []string{"*", "test_run.completed"}, IsActive: true,
			Edges: ent.WebhookEndpointEdges{Project: &ent.Project{ID: projectID}},
		}},
	}
	d := apigateway.NewWebhookDispatcher(repo, nil)
	apigateway.ExportWebhookDispatcherInitJobs(d, 4)

	require.NoError(t, apigateway.ExportWebhookDispatcherDispatchEvent(
		d, context.Background(), "tms.testrun.completed", []byte(`{}`), map[string]interface{}{}))

	assert.Empty(t, apigateway.ExportWebhookDispatcherDrainJobEndpointIDs(d))
}

// TMS subject без UUID в path, но project_id в payload — доставляем в нужный проект.
func TestWebhookDispatcher_dispatchEvent_ProjectFromPayload(t *testing.T) {
	projectA := uuid.New()
	projectB := uuid.New()
	epA := uuid.New()
	epB := uuid.New()

	repo := &stubWebhookRepo{
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
	}
	d := apigateway.NewWebhookDispatcher(repo, nil)
	apigateway.ExportWebhookDispatcherInitJobs(d, 4)

	payload := map[string]interface{}{"project_id": projectB.String(), "run_id": uuid.New().String()}
	require.NoError(t, apigateway.ExportWebhookDispatcherDispatchEvent(
		d, context.Background(), "tms.testrun.completed", []byte(`{}`), payload))

	ids := apigateway.ExportWebhookDispatcherDrainJobEndpointIDs(d)
	require.Equal(t, []uuid.UUID{epB}, ids)
}

func TestExtractEventProjectID(t *testing.T) {
	t.Parallel()
	pid := uuid.New()

	got, ok := apigateway.ExportExtractEventProjectID("project."+pid.String()+".issue.created", nil)
	require.True(t, ok)
	assert.Equal(t, pid, got)

	_, ok = apigateway.ExportExtractEventProjectID("project.created", nil)
	assert.False(t, ok)

	got, ok = apigateway.ExportExtractEventProjectID("tms.testrun.completed", map[string]interface{}{
		"project_id": pid.String(),
	})
	require.True(t, ok)
	assert.Equal(t, pid, got)
}
