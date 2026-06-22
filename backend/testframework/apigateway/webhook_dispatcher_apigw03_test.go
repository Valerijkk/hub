package apigateway_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/apigateway"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func TestWebhookFilterSubjects_CoversWebhookEvents(t *testing.T) {
	t.Parallel()

	want := []string{
		"project.*.issue.>",
		"tracker.*.issue.>",
		"project.created",
		"tms.run.completed.>",
		"tms.testrun.completed",
		"sprint.>",
		"messenger.channel.*.message_sent",
	}
	assert.Equal(t, want, natsclient.WebhookFilterSubjects)
	assert.Equal(t, want, apigateway.ExportWebhookFilterSubjects())
}

func TestWebhookDispatcher_ConsumerUsesFilterSubjects(t *testing.T) {
	nc := startWebhookJetStream(t)
	d := apigateway.NewWebhookDispatcher(&stubWebhookRepo{}, nc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, d.Start(ctx))
	defer d.Stop()

	stream, err := nc.JetStream.Stream(ctx, natsclient.StreamName)
	require.NoError(t, err)

	consumer, err := stream.Consumer(ctx, "webhook-dispatcher")
	require.NoError(t, err)

	info, err := consumer.Info(ctx)
	require.NoError(t, err)
	assert.Equal(t, natsclient.WebhookFilterSubjects, info.Config.FilterSubjects)
}

type countingWebhookRepo struct {
	stubWebhookRepo
	listCalls atomic.Int32
}

func (s *countingWebhookRepo) ListActiveWebhookEndpoints(context.Context) ([]*ent.WebhookEndpoint, error) {
	s.listCalls.Add(1)
	return s.endpoints, nil
}

func TestWebhookDispatcher_SkipsNonWebhookSubjects(t *testing.T) {
	nc := startWebhookJetStream(t)
	repo := &countingWebhookRepo{
		stubWebhookRepo: stubWebhookRepo{
			endpoints: []*ent.WebhookEndpoint{{
				ID: uuid.New(), URL: "https://example.com/hook", Secret: "s",
				Events: []string{"*"}, IsActive: true,
			}},
		},
	}
	d := apigateway.NewWebhookDispatcher(repo, nc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, d.Start(ctx))
	defer d.Stop()

	js := nc.JetStream
	_, err := js.Publish(ctx, "notification.user."+uuid.New().String(), []byte(`{"type":"assigned"}`))
	require.NoError(t, err)

	_, err = js.Publish(ctx, "project."+uuid.New().String()+".issue.created", []byte(`{"issue_id":"1"}`))
	require.NoError(t, err)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if repo.listCalls.Load() >= 1 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	assert.Equal(t, int32(1), repo.listCalls.Load(),
		"notification subject must not trigger webhook endpoint scan; only project.issue.created should")
}
