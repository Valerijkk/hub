package ws

import (
	"strings"
)

// WSTicketProtocolPrefix — канонический префикс Sec-WebSocket-Protocol
// для short-lived ticket (ADR-WS-JWT-HANDSHAKE-2026).
const WSTicketProtocolPrefix = "hub.ws.ticket."

// ParseWSTicketProtocols ищет ticket в списке client subprotocols.
// Возвращает ticket, полный negotiated protocol и ok.
func ParseWSTicketProtocols(protocols []string) (ticket, negotiated string, ok bool) {
	for _, p := range protocols {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, WSTicketProtocolPrefix) {
			t := strings.TrimPrefix(p, WSTicketProtocolPrefix)
			if t == "" {
				continue
			}
			// Subprotocol token не должен содержать пробелы/запятые.
			if strings.ContainsAny(t, " \t,") {
				continue
			}
			return t, p, true
		}
	}
	return "", "", false
}

// SplitSecWebSocketProtocol разбирает заголовок Sec-WebSocket-Protocol.
func SplitSecWebSocketProtocol(header string) []string {
	if header == "" {
		return nil
	}
	parts := strings.Split(header, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
