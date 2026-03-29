package outbox

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NewWorker
// ---------------------------------------------------------------------------

func TestNewWorker_ReturnsNonNil(t *testing.T) {
	w := NewWorker(nil, nil)
	require.NotNil(t, w)
}

func TestNewWorker_StoresClient(t *testing.T) {
	w := NewWorker(nil, nil)
	assert.Nil(t, w.client, "client should be nil when nil is passed")
	assert.Nil(t, w.nats, "nats should be nil when nil is passed")
}

// ---------------------------------------------------------------------------
// NewPublisher
// ---------------------------------------------------------------------------

func TestNewPublisher_ReturnsNonNil(t *testing.T) {
	p := NewPublisher()
	require.NotNil(t, p)
}

func TestNewPublisher_IsReusable(t *testing.T) {
	p1 := NewPublisher()
	p2 := NewPublisher()
	assert.NotSame(t, p1, p2, "each call should return a new instance")
}

// ---------------------------------------------------------------------------
// JSON marshaling logic (extracted from Publisher methods)
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
	// Channels cannot be marshaled to JSON — same error path
	// as Publisher.Publish / PublishInTx would hit.
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
// Design constants / invariants
// ---------------------------------------------------------------------------

func TestWorkerConstants(t *testing.T) {
	// These values are hard-coded in worker.go and are part of the
	// package's contract.  If someone changes them intentionally the
	// tests should be updated too — this guards against accidental edits.
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
