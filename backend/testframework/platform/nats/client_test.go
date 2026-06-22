package nats_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func TestStreamNameConstant(t *testing.T) {
	assert.Equal(t, "EVENTS", natsclient.StreamName, "stream name should be EVENTS")
}

func TestStreamSubjects(t *testing.T) {
	expectedPrefixes := []string{
		"project.>",
		"tracker.>",
		"messenger.channel.>",
		"tms.>",
		"calls.channel.>",
		"auth.session.>",
		"sprint.>",
	}

	assert.Equal(t, 7, len(expectedPrefixes), "should have 7 subject prefixes in EVENTS stream")
}

func TestClientStructFields(t *testing.T) {
	var c natsclient.Client
	assert.Nil(t, c.Conn, "zero-value Client should have nil Conn")
	assert.Nil(t, c.JetStream, "zero-value Client should have nil JetStream")
}
