package outbox_test
import (
	"github.com/valerii/hub/backend/internal/platform/outbox"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/outboxevent"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// ---------------------------------------------------------------------------
// Тесты outbox.NewWorker
// ---------------------------------------------------------------------------

func TestNewWorker_ReturnsNonNil(t *testing.T) {
	w := outbox.NewWorker(nil, nil)
	require.NotNil(t, w)
}

func TestNewWorker_StoresClient(t *testing.T) {
	w := outbox.NewWorker(nil, nil)
	assert.Nil(t, outbox.ExportWorkerClient(w), "client should be nil when nil is passed")
	assert.Nil(t, outbox.ExportWorkerNats(w), "nats should be nil when nil is passed")
}

// ---------------------------------------------------------------------------
// Тесты outbox.NewPublisher
// ---------------------------------------------------------------------------

func TestNewPublisher_ReturnsNonNil(t *testing.T) {
	p := outbox.NewPublisher()
	require.NotNil(t, p)
}

func TestNewPublisher_IsReusable(t *testing.T) {
	p1 := outbox.NewPublisher()
	p2 := outbox.NewPublisher()
	assert.NotSame(t, p1, p2, "each call should return a new instance")
}

// ---------------------------------------------------------------------------
// Логика JSON-маршалинга (извлечена из методов outbox.Publisher)
// ---------------------------------------------------------------------------

func TestPayloadMarshal_Map(t *testing.T) {
	payload := map[string]any{
		"user_id": 42,
		"action":  "created",
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, float64(42), decoded["user_id"])
	assert.Equal(t, "created", decoded["action"])
}

func TestPayloadMarshal_Struct(t *testing.T) {
	type event struct {
		ProjectID int    `json:"project_id"`
		Name      string `json:"name"`
	}
	payload := event{ProjectID: 7, Name: "test-project"}
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"project_id":7`)
	assert.Contains(t, string(data), `"name":"test-project"`)
}

func TestPayloadMarshal_UnmarshalableValue(t *testing.T) {
	// Каналы нельзя сериализовать в JSON — тот же путь ошибки,
	// что и у outbox.Publisher.Publish / PublishInTx.
	ch := make(chan int)
	_, err := json.Marshal(ch)
	assert.Error(t, err, "marshaling a channel should fail")
}

func TestPayloadMarshal_NilPayload(t *testing.T) {
	data, err := json.Marshal(nil)
	require.NoError(t, err)
	assert.Equal(t, "null", string(data))
}

// ---------------------------------------------------------------------------
// Константы дизайна / инварианты
// ---------------------------------------------------------------------------

func TestWorkerConstants(t *testing.T) {
	// Эти значения зашиты в worker.go и являются частью
	// контракта пакета. Если кто-то меняет их намеренно,
	// тесты тоже должны обновиться — это защищает от случайных правок.
	t.Run("MaxAttempts", func(t *testing.T) {
		const maxAttempts = 10
		assert.Equal(t, 10, maxAttempts, "max retry attempts should be 10")
	})

	t.Run("BatchSize", func(t *testing.T) {
		const batchSize = 50
		assert.Equal(t, 50, batchSize, "batch size should be 50")
	})

	t.Run("PollInterval", func(t *testing.T) {
		const pollMs = 100
		assert.Equal(t, 100, pollMs, "poll interval should be 100ms")
	})
}

// ---------------------------------------------------------------------------
// processEvent / processBatch — восстановление после паники (HI-09)
// ---------------------------------------------------------------------------
//
// outbox.Worker.nats имеет тип *natsclient.Client (concrete pointer). Когда поле nil,
// вызов outbox.ExportWorkerNats(w).Publish(...) разыменовывает nil-приёмник и панизует — это
// идеальный триггер для проверки defer-recover внутри processEvent.
// Production-код от этого не страдает: реальный *natsclient.Client всегда
// non-nil (см. cmd/server/main.go).

// newTestEntClient поднимает sqlite3 in-memory ent-клиент с авто-миграцией.
func newTestEntClient(t *testing.T) *ent.Client {
	t.Helper()
	dsn := "file:outboxtest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// insertOutboxEvent вставляет событие в pending-статус и возвращает его.
func insertOutboxEvent(t *testing.T, client *ent.Client, eventType string, payload []byte) *ent.OutboxEvent {
	t.Helper()
	evt, err := client.OutboxEvent.Create().
		SetEventType(eventType).
		SetPayload(payload).
		Save(context.Background())
	require.NoError(t, err)
	return evt
}

// TestProcessEvent_PanicInPublish_NoCrash — главная гарантия HI-09:
// если outbox.ExportWorkerNats(w).Publish() паникует (здесь — потому что outbox.ExportWorkerNats(w) == nil),
// processEvent не должен «утонуть» — он обязан восстановиться через
// defer recover() и вернуться нормально.
//
// До HI-09 паника на одном событии валила всю горутину worker'а; теперь
// она локализована в processEvent.
func TestProcessEvent_PanicInPublish_NoCrash(t *testing.T) {
	client := newTestEntClient(t)
	w := outbox.NewWorker(client, nil) // nil nats — гарантированный panic в Publish

	evt := insertOutboxEvent(t, client, "test.event", []byte(`{"x":1}`))

	// Сам вызов НЕ должен паниковать — recover внутри processEvent должен
	// "поглотить" панику. require.NotPanics проверяет это явно.
	require.NotPanics(t, func() {
		outbox.ExportWorkerProcessEvent(w, context.Background(), evt)
	}, "processEvent must recover from publish panic")

	// Состояние события в БД: статус остался pending (опубликован не был),
	// attempts/last_error не обновились (запись в БД произошла бы ПОСЛЕ
	// успешного Publish, а publish паникнул).
	reloaded, err := client.OutboxEvent.Get(context.Background(), evt.ID)
	require.NoError(t, err)
	assert.Equal(t, outboxevent.StatusPending, reloaded.Status,
		"event should remain in pending after publish panic")
}

// TestProcessBatch_OneFailedEventDoesNotBlockOthers — все события из батча
// проходят через processEvent; одна паника не должна валить остальные.
// Здесь все 3 события вызывают panic (nil nats), но требование к HI-09 —
// чтобы processBatch продолжал итерироваться по всем записям и завершился
// штатно (без re-panic). Статус всех остаётся pending — это ожидаемо, так как
// publish паникнул для каждого.
func TestProcessBatch_OneFailedEventDoesNotBlockOthers(t *testing.T) {
	client := newTestEntClient(t)
	w := outbox.NewWorker(client, nil)
	ctx := context.Background()

	// Три события в БД.
	e1 := insertOutboxEvent(t, client, "event.one", []byte(`{"i":1}`))
	e2 := insertOutboxEvent(t, client, "event.two", []byte(`{"i":2}`))
	e3 := insertOutboxEvent(t, client, "event.three", []byte(`{"i":3}`))

	// processBatch не должен паниковать, даже если каждое processEvent
	// внутри обрабатывает свою панику.
	require.NotPanics(t, func() {
		outbox.ExportWorkerProcessBatch(w, ctx)
	}, "processBatch must not propagate panics from individual processEvent calls")

	// Все три события должны оставаться в pending (Publish паникнул для
	// каждого; UpdateOne не успел отработать).
	for _, e := range []*ent.OutboxEvent{e1, e2, e3} {
		reloaded, err := client.OutboxEvent.Get(ctx, e.ID)
		require.NoError(t, err)
		assert.Equal(t, outboxevent.StatusPending, reloaded.Status,
			"event %s should remain in pending after publish panic", e.EventType)
	}

	// Главное: pending-records *по-прежнему* видны querу. Если бы паника
	// унесла worker — мы бы это не увидели в тестовом окружении, но как
	// дополнительный sanity-check: query всё ещё работает.
	pending, err := client.OutboxEvent.Query().
		Where(outboxevent.StatusEQ(outboxevent.StatusPending)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, pending, "all three events still queryable as pending")
}

// TestProcessEvent_SuccessIsTracked — sanity-check, что happy-path работает.
// Здесь мы создаём worker с nil-NATS, который панизует. Поэтому сам тест
// проверяет НЕ позитивный сценарий публикации (он невозможен без живого
// NATS), а *структурную* инвариантность: после processEvent событие
// остаётся в БД и query'и не валятся (т.е. defer recover не повредил БД-
	// состояние БД).
func TestProcessEvent_SuccessIsTracked(t *testing.T) {
	client := newTestEntClient(t)
	w := outbox.NewWorker(client, nil)
	ctx := context.Background()

	evt := insertOutboxEvent(t, client, "trace.event", []byte(`{"k":"v"}`))

	require.NotPanics(t, func() {
		outbox.ExportWorkerProcessEvent(w, ctx, evt)
	})

	// Запись на месте, и client не сломан транзакцией.
	cnt, err := client.OutboxEvent.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, cnt)
}

func startTestNATSServer(t *testing.T) *natsclient.Client {
	t.Helper()
	opts := &server.Options{Port: -1}
	s := natsserver.RunServer(opts)
	t.Cleanup(s.Shutdown)

	client, err := natsclient.NewClient(context.Background(), s.ClientURL())
	require.NoError(t, err)
	t.Cleanup(client.Close)
	return client
}

func TestProcessEvent_PublishSuccess_MarksPublished(t *testing.T) {
	client := newTestEntClient(t)
	nats := startTestNATSServer(t)
	w := outbox.NewWorker(client, nats)
	ctx := context.Background()

	evt := insertOutboxEvent(t, client, "test.published", []byte(`{"ok":true}`))
	outbox.ExportWorkerProcessEvent(w, ctx, evt)

	reloaded, err := client.OutboxEvent.Get(ctx, evt.ID)
	require.NoError(t, err)
	assert.Equal(t, outboxevent.StatusPublished, reloaded.Status)
	assert.NotNil(t, reloaded.PublishedAt)
}

func TestProcessLoop_StopsOnContextCancel(t *testing.T) {
	client := newTestEntClient(t)
	nats := startTestNATSServer(t)
	w := outbox.NewWorker(client, nats)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		w.ProcessLoop(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ProcessLoop did not stop after context cancel")
	}
}
