//go:build hubtest

package ws_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valerii/hub/backend/internal/ws"
)

func TestParseWSTicketProtocols(t *testing.T) {
	ticket, negotiated, ok := ws.ParseWSTicketProtocols([]string{
		"chat",
		ws.WSTicketProtocolPrefix + "abc123",
	})
	assert.True(t, ok)
	assert.Equal(t, "abc123", ticket)
	assert.Equal(t, ws.WSTicketProtocolPrefix+"abc123", negotiated)

	_, _, ok = ws.ParseWSTicketProtocols([]string{"chat", "other"})
	assert.False(t, ok)

	_, _, ok = ws.ParseWSTicketProtocols([]string{ws.WSTicketProtocolPrefix})
	assert.False(t, ok)
}

func TestSplitSecWebSocketProtocol(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, ws.SplitSecWebSocketProtocol("a, b"))
	assert.Nil(t, ws.SplitSecWebSocketProtocol(""))
}
