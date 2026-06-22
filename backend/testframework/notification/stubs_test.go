package notification_test

// Тесты модуля уведомлений: интеграционные и unit-сценарии testframework.

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type stubJSMsg struct {
	subject string
	data    []byte
	acked   bool
}

func (m *stubJSMsg) Subject() string                           { return m.subject }
func (m *stubJSMsg) Data() []byte                              { return m.data }
func (m *stubJSMsg) Headers() nats.Header                      { return nil }
func (m *stubJSMsg) Metadata() (*jetstream.MsgMetadata, error) { return nil, nil }
func (m *stubJSMsg) Ack() error                                { m.acked = true; return nil }
func (m *stubJSMsg) DoubleAck(context.Context) error           { return nil }
func (m *stubJSMsg) Nak() error                                { return nil }
func (m *stubJSMsg) NakWithDelay(time.Duration) error          { return nil }
func (m *stubJSMsg) InProgress() error                         { return nil }
func (m *stubJSMsg) Term() error                               { return nil }
func (m *stubJSMsg) TermWithReason(string) error               { return nil }
func (m *stubJSMsg) Reply() string                             { return "" }

// stubJetStreamMsg — псевдоним для перенесённых coverage-тестов.
type stubJetStreamMsg = stubJSMsg
