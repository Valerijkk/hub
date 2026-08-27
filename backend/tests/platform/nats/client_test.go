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
	require.Equal(t, []string{
		"project.>",
		"tracker.>",
		"messenger.channel.>",
		"tms.>",
		"calls.channel.>",
		"calls.recording.>",
		"auth.session.>",
		"sprint.>",
		"notification.>",
		"audit.>",
		"calendar.>",
		"mail.>",
	}, natsclient.StreamSubjects)
	require.Len(t, natsclient.StreamSubjects, 12,
		"EVENTS stream must keep the canonical 12 subjects")

	calendarSubjects := 0
	mailSubjects := 0
	for _, subject := range natsclient.StreamSubjects {
		if subject == "calendar.>" {
			calendarSubjects++
		}
		if subject == "mail.>" {
			mailSubjects++
		}
	}
	require.Equal(t, 1, calendarSubjects, "calendar.> must occur exactly once")
	require.Equal(t, 1, mailSubjects, "mail.> must occur exactly once")

	assert.Contains(t, natsclient.StreamSubjects, "notification.>")
	assert.Contains(t, natsclient.StreamSubjects, "calls.recording.>")
	assert.Contains(t, natsclient.StreamSubjects, "audit.>")
}

func TestClientStructFields(t *testing.T) {
	var c natsclient.Client
	assert.Nil(t, c.Conn, "zero-value Client should have nil Conn")
	assert.Nil(t, c.JetStream, "zero-value Client should have nil JetStream")
}
