package apigateway_test
import (
	"github.com/valerii/hub/backend/internal/apigateway"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

type stubWebhookRepo struct {
	endpoints  []*ent.WebhookEndpoint
	deliveries []apigateway.CreateWebhookDeliveryInput
}

func (s *stubWebhookRepo) ListActiveWebhookEndpoints(context.Context) ([]*ent.WebhookEndpoint, error) {
	return s.endpoints, nil
}

func (s *stubWebhookRepo) CreateWebhookDelivery(_ context.Context, input apigateway.CreateWebhookDeliveryInput) error {
	s.deliveries = append(s.deliveries, input)
	return nil
}

func (s *stubWebhookRepo) CreateWebhookEndpoint(context.Context, uuid.UUID, string, string, []string) (*ent.WebhookEndpoint, error) {
	return nil, nil
}

func (s *stubWebhookRepo) ListWebhookEndpointsByProject(context.Context, uuid.UUID) ([]*ent.WebhookEndpoint, error) {
	return nil, nil
}

func (s *stubWebhookRepo) UpdateWebhookEndpoint(context.Context, uuid.UUID, apigateway.UpdateWebhookEndpointInput) (*ent.WebhookEndpoint, error) {
	return nil, nil
}

func (s *stubWebhookRepo) DeleteWebhookEndpoint(context.Context, uuid.UUID) error {
	return nil
}

func (s *stubWebhookRepo) ListWebhookDeliveries(context.Context, uuid.UUID, int) ([]*ent.WebhookDelivery, error) {
	return nil, nil
}

func (s *stubWebhookRepo) WebhookEndpointProjectID(context.Context, uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}

type stubWebhookJSMsg struct {
	subject string
	data    []byte
}

func (m *stubWebhookJSMsg) Subject() string                         { return m.subject }
func (m *stubWebhookJSMsg) Data() []byte                            { return m.data }
func (m *stubWebhookJSMsg) Headers() nats.Header                    { return nil }
func (m *stubWebhookJSMsg) Metadata() (*jetstream.MsgMetadata, error) { return nil, nil }
func (m *stubWebhookJSMsg) Ack() error                              { return nil }
func (m *stubWebhookJSMsg) DoubleAck(context.Context) error         { return nil }
func (m *stubWebhookJSMsg) Nak() error                              { return nil }
func (m *stubWebhookJSMsg) NakWithDelay(time.Duration) error        { return nil }
func (m *stubWebhookJSMsg) InProgress() error                       { return nil }
func (m *stubWebhookJSMsg) Term() error                             { return nil }
func (m *stubWebhookJSMsg) TermWithReason(string) error             { return nil }
func (m *stubWebhookJSMsg) Reply() string                           { return "" }

func startWebhookJetStream(t *testing.T) *natsclient.Client {
	t.Helper()
	opts := &server.Options{
		JetStream: true,
		StoreDir:  t.TempDir(),
		Port:      -1,
	}
	s := natsserver.RunServer(opts)
	t.Cleanup(s.Shutdown)
	client, err := natsclient.NewClient(context.Background(), s.ClientURL())
	require.NoError(t, err)
	t.Cleanup(client.Close)
	require.NoError(t, client.EnsureStreams(context.Background()))
	return client
}

func TestWebhookDispatcher_StartStop_NilNats(t *testing.T) {
	d := apigateway.NewWebhookDispatcher(&stubWebhookRepo{}, nil)
	require.NoError(t, d.Start(context.Background()))
	d.Stop()
}

func TestWebhookDispatcher_StartStop_WithJetStream(t *testing.T) {
	nc := startWebhookJetStream(t)
	repo := &stubWebhookRepo{}
	d := apigateway.NewWebhookDispatcher(repo, nc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, d.Start(ctx))
	time.Sleep(100 * time.Millisecond)
	d.Stop()
}

func TestWebhookDispatcher_handleMessage_InvalidJSON(t *testing.T) {
	d := apigateway.NewWebhookDispatcher(&stubWebhookRepo{}, nil)
	apigateway.ExportWebhookDispatcherHandleMessage(d, context.Background(), &stubWebhookJSMsg{
		subject: "test.event",
		data:    []byte(`{invalid`),
	})
}

func TestWebhookDispatcher_deliverWithRetry_Success(t *testing.T) {
	useUnsafeClient(t)

	var received bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	epID := uuid.New()
	repo := &stubWebhookRepo{
		endpoints: []*ent.WebhookEndpoint{{
			ID: epID, URL: localTestURL(srv), Secret: "secret", Events: []string{"*"}, IsActive: true,
		}},
	}
	d := apigateway.NewWebhookDispatcher(repo, nil)
	body := []byte(`{"issue_id":"1"}`)
	payload := map[string]interface{}{"issue_id": "1"}
	require.NoError(t, json.Unmarshal(body, &payload))

	apigateway.ExportWebhookDispatcherDeliverWithRetry(d, context.Background(), repo.endpoints[0], "project.x.issue.created", body, payload)

	require.True(t, received)
	require.Len(t, repo.deliveries, 1)
	assert.Equal(t, http.StatusOK, repo.deliveries[0].StatusCode)
	assert.Equal(t, epID, repo.deliveries[0].EndpointID)
}

func TestWebhookDispatcher_dispatchEvent_SkipsUnsubscribed(t *testing.T) {
	repo := &stubWebhookRepo{
		endpoints: []*ent.WebhookEndpoint{{
			ID: uuid.New(), URL: "https://example.com/hook", Secret: "s",
			Events: []string{"other.event"}, IsActive: true,
		}},
	}
	d := apigateway.NewWebhookDispatcher(repo, nil)
	apigateway.ExportWebhookDispatcherInitJobs(d, 1)

	apigateway.ExportWebhookDispatcherDispatchEvent(d, context.Background(), "project.x.issue.created", []byte(`{}`), map[string]interface{}{})

	if apigateway.ExportWebhookDispatcherJobAvailable(d) {
		t.Fatal("unexpected job enqueued for unsubscribed event")
	}
}

func TestWebhookService_ListDeliveries(t *testing.T) {
	dsn := "file:whdel_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	repo := apigateway.NewEntRepository(client, apigatewayTestCipher(t))
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Del Proj").SetKey("DP").Save(ctx)
	require.NoError(t, err)

	ep, err := repo.CreateWebhookEndpoint(ctx, proj.ID, "https://example.com/hook", "sec", []string{"*"})
	require.NoError(t, err)

	require.NoError(t, repo.CreateWebhookDelivery(ctx, apigateway.CreateWebhookDeliveryInput{
		EventType: "test.event", Payload: map[string]interface{}{"k": "v"},
		StatusCode: 200, ResponseBody: "ok", EndpointID: ep.ID,
	}))

	svc := apigateway.NewWebhookService(repo)
	list, err := svc.ListDeliveries(ctx, ep.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "test.event", list[0].EventType)
}
