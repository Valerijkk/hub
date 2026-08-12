package apigateway_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/apigateway"
)

// subscribedRepo возвращает один активный эндпоинт с events=["*"] и привязанный
// к проекту (Edges.Project обязателен после TASK-012 project-scope фильтра).
func subscribedRepo() (*stubWebhookRepo, uuid.UUID) {
	projectID := uuid.New()
	return &stubWebhookRepo{
		endpoints: []*ent.WebhookEndpoint{{
			ID: uuid.New(), URL: "https://example.com/hook", Secret: "s",
			Events: []string{"*"}, IsActive: true,
			Edges: ent.WebhookEndpointEdges{Project: &ent.Project{ID: projectID}},
		}},
	}, projectID
}

func projectIssueSubject(projectID uuid.UUID) string {
	return "project." + projectID.String() + ".issue.created"
}

// APIGW-02: при переполнении воркер-очереди dispatchEvent обязан вернуть
// errWebhookQueueFull, а не «тихо» дропнуть задачу. Только так вызывающий
// consumer узнает, что Ack делать нельзя, и JetStream доставит событие повторно.
func TestWebhookDispatcher_dispatchEvent_QueueFull_ReturnsError(t *testing.T) {
	repo, projectID := subscribedRepo()
	d := apigateway.NewWebhookDispatcher(repo, nil)
	apigateway.ExportWebhookDispatcherInitJobs(d, 1)

	ctx := context.Background()
	body := []byte(`{"k":"v"}`)
	payload := map[string]interface{}{"k": "v"}
	subject := projectIssueSubject(projectID)

	// Первый dispatch заполняет очередь (buffer=1) — задача поставлена, ошибки нет.
	require.NoError(t, apigateway.ExportWebhookDispatcherDispatchEvent(d, ctx, subject, body, payload))

	// Второй dispatch не может поставить задачу — очередь полна.
	err := apigateway.ExportWebhookDispatcherDispatchEvent(d, ctx, subject, body, payload)
	require.Error(t, err)
	assert.ErrorIs(t, err, apigateway.ExportErrWebhookQueueFull())
}

// APIGW-02: handleMessage должен пробросить ошибку переполнения, чтобы consumer
// сделал NAK вместо Ack (предотвращение потери данных при всплеске).
func TestWebhookDispatcher_handleMessage_QueueFull_ReturnsError(t *testing.T) {
	repo, projectID := subscribedRepo()
	d := apigateway.NewWebhookDispatcher(repo, nil)
	apigateway.ExportWebhookDispatcherInitJobs(d, 1)

	ctx := context.Background()
	msg := &stubWebhookJSMsg{subject: projectIssueSubject(projectID), data: []byte(`{"k":"v"}`)}

	// Первое сообщение заполняет очередь.
	require.NoError(t, apigateway.ExportWebhookDispatcherHandleMessage(d, ctx, msg))

	// Второе — очередь полна, ожидаем ошибку (триггер NAK у consumer).
	err := apigateway.ExportWebhookDispatcherHandleMessage(d, ctx, msg)
	require.Error(t, err)
	assert.ErrorIs(t, err, apigateway.ExportErrWebhookQueueFull())
}

// Успешная постановка в очередь не должна возвращать ошибку (consumer делает Ack).
func TestWebhookDispatcher_handleMessage_Success_NoError(t *testing.T) {
	repo, projectID := subscribedRepo()
	d := apigateway.NewWebhookDispatcher(repo, nil)
	apigateway.ExportWebhookDispatcherInitJobs(d, 4)

	msg := &stubWebhookJSMsg{subject: projectIssueSubject(projectID), data: []byte(`{"k":"v"}`)}
	require.NoError(t, apigateway.ExportWebhookDispatcherHandleMessage(d, context.Background(), msg))
}

// Poison message (битый JSON): повтор не поможет — handleMessage возвращает nil,
// чтобы сообщение было подтверждено (Ack), а не зациклилось на NAK.
func TestWebhookDispatcher_handleMessage_InvalidJSON_AcksNoError(t *testing.T) {
	d := apigateway.NewWebhookDispatcher(&stubWebhookRepo{}, nil)
	msg := &stubWebhookJSMsg{subject: "project.x.issue.created", data: []byte(`{invalid`)}
	require.NoError(t, apigateway.ExportWebhookDispatcherHandleMessage(d, context.Background(), msg))
}

// Транзиентная ошибка получения эндпоинтов также должна привести к NAK (ошибка),
// а не к Ack — иначе событие потеряется.
func TestWebhookDispatcher_dispatchEvent_RepoError_ReturnsError(t *testing.T) {
	repo := &errListWebhookRepo{}
	d := apigateway.NewWebhookDispatcher(repo, nil)
	apigateway.ExportWebhookDispatcherInitJobs(d, 4)

	err := apigateway.ExportWebhookDispatcherDispatchEvent(d, context.Background(),
		"project.x.issue.created", []byte(`{}`), map[string]interface{}{})
	require.Error(t, err)
}

// errListWebhookRepo имитирует транзиентный сбой выборки активных эндпоинтов.
type errListWebhookRepo struct{ stubWebhookRepo }

func (r *errListWebhookRepo) ListActiveWebhookEndpoints(context.Context) ([]*ent.WebhookEndpoint, error) {
	return nil, errors.New("db unavailable")
}
