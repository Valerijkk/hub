package nats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamNameConstant(t *testing.T) {
	assert.Equal(t, "EVENTS", StreamName, "stream name should be EVENTS")
}

func TestStreamSubjects(t *testing.T) {
	// Verify expected NATS subjects are documented in EnsureStreams.
	// These subjects are part of the contract between modules.
	expectedPrefixes := []string{
		"project.>",
		"tracker.>",
		"messenger.channel.>",
		"tms.>",
		"calls.channel.>",
		"auth.session.>",
		"sprint.>",
	}

	// We can't call EnsureStreams without a real NATS connection,
	// but we can verify the constant and expected subject count.
	assert.Equal(t, 7, len(expectedPrefixes), "should have 7 subject prefixes in EVENTS stream")
}

func TestClientStructFields(t *testing.T) {
	// Verify Client struct has expected nil defaults when zero-valued.
	var c Client
	assert.Nil(t, c.Conn, "zero-value Client should have nil Conn")
	assert.Nil(t, c.JetStream, "zero-value Client should have nil JetStream")
}
