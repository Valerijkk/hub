package nats_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func TestStreamNameConstant(t *testing.T) {
	assert.Equal(t, "EVENTS", natsclient.StreamName, "stream name should be EVENTS")
}

func TestStreamSubjects(t *testing.T) {
	// A2W4-2-003: unit assert = канон EnsureStreams, не локальный устаревший слайс.
	require.Equal(t, natsclient.StreamSubjects, []string{
		"project.>",
		"tracker.>",
		"messenger.channel.>",
		"tms.>",
		"calls.channel.>",
		"calls.recording.>",
		"auth.session.>",
		"sprint.>",
		"notification.>",
	})
	assert.Equal(t, 9, len(natsclient.StreamSubjects),
		"EVENTS stream must keep ME-07 subjects (notification.>, calls.recording.>)")
	assert.Contains(t, natsclient.StreamSubjects, "notification.>")
	assert.Contains(t, natsclient.StreamSubjects, "calls.recording.>")
}

func TestClientStructFields(t *testing.T) {
	var c natsclient.Client
	assert.Nil(t, c.Conn, "zero-value Client should have nil Conn")
	assert.Nil(t, c.JetStream, "zero-value Client should have nil JetStream")
}
