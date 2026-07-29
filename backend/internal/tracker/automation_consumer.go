package tracker

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// errAutomationUnknownSubject — неизвестный/битый subject JetStream (A2W1-A3-002).
var errAutomationUnknownSubject = errors.New("automation: unknown subject")

// automationUnknownSubjectTotal — счётчик неизвестных subject (метрика для тестов/observability).
var automationUnknownSubjectTotal atomic.Uint64

// AutomationUnknownSubjectTotal возвращает число неизвестных subject с момента старта процесса.
func AutomationUnknownSubjectTotal() uint64 {
	return automationUnknownSubjectTotal.Load()
}

// AutomationFilterSubjects — JetStream filter: только kinds, которые обрабатывает engine.
// A3W1-A3-001: широкий project.*.issue.> ловил updated/comment_created/link_created → Nak-шторм.
var AutomationFilterSubjects = []string{
	"project.*.issue.moved",
	"project.*.issue.created",
	"project.*.issue.assigned",
}

// knownNoOpIssueEventKinds — ожидаемые issue-события без automation-правил.
// Ack/Term без метрики; Nak только для мусора (TASK-088 / A2W1-A3-002).
var knownNoOpIssueEventKinds = map[string]struct{}{
	"updated":         {},
	"comment_created": {},
	"link_created":    {},
}

// AutomationConsumer выполняет правила автоматизации по JetStream-событиям трекера.
type AutomationConsumer struct {
	nc     *natsclient.Client
	engine *AutomationEngine

	cancel context.CancelFunc
}

// NewAutomationConsumer создаёт consumer автоматизаций.
func NewAutomationConsumer(nc *natsclient.Client, engine *AutomationEngine) *AutomationConsumer {
	return &AutomationConsumer{nc: nc, engine: engine}
}

// Start подписывается на события задач в общем JetStream-потоке EVENTS.
func (c *AutomationConsumer) Start(ctx context.Context) error {
	if c.nc == nil || c.engine == nil {
		slog.Warn("automation consumer skipped: nil dependencies")
		return nil
	}
	js := c.nc.JetStream
	stream, err := js.Stream(ctx, natsclient.StreamName)
	if err != nil {
		return err
	}
	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:        "automation-dispatcher",
		FilterSubjects: AutomationFilterSubjects,
		AckPolicy:      jetstream.AckExplicitPolicy,
		MaxDeliver:     5,
		AckWait:        30 * time.Second,
	})
	if err != nil {
		return err
	}
	consCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	go func() {
		iter, err := cons.Messages(jetstream.PullMaxMessages(10))
		if err != nil {
			slog.Error("automation: messages iterator", "error", err)
			return
		}
		defer iter.Stop()
		for {
			select {
			case <-consCtx.Done():
				return
			default:
			}
			msg, err := iter.Next()
			if err != nil {
				if consCtx.Err() != nil {
					return
				}
				continue
			}
			c.processMessage(consCtx, msg)
		}
	}()
	slog.Info("automation consumer started")
	return nil
}

// Stop останавливает цикл consumer.
func (c *AutomationConsumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *AutomationConsumer) processMessage(ctx context.Context, msg jetstream.Msg) {
	kind := parseIssueEventKind(msg.Subject())
	var processErr error
	switch kind {
	case "moved":
		var evt IssueMovedEvent
		if err := json.Unmarshal(msg.Data(), &evt); err != nil {
			slog.Error("automation: unmarshal moved", "error", err)
			processErr = err
		} else {
			processErr = c.engine.HandleStatusChanged(ctx, evt)
		}
	case "created":
		var evt IssueCreatedEvent
		if err := json.Unmarshal(msg.Data(), &evt); err != nil {
			slog.Error("automation: unmarshal created", "error", err)
			processErr = err
		} else {
			processErr = c.engine.HandleIssueCreated(ctx, evt)
		}
	case "assigned":
		var evt IssueAssignedEvent
		if err := json.Unmarshal(msg.Data(), &evt); err != nil {
			slog.Error("automation: unmarshal assigned", "error", err)
			processErr = err
		} else {
			processErr = c.engine.HandleIssueAssigned(ctx, evt)
		}
	default:
		if _, knownNoOp := knownNoOpIssueEventKinds[kind]; knownNoOp {
			// A3W1-A3-001: известный no-op (updated/comment/link) — Ack без Nak/метрики.
			processErr = nil
		} else {
			// A2W1-A3-002 / TASK-088: неизвестный subject → Nak + метрика (не молчаливый Ack).
			slog.Warn("automation: unknown subject", "subject", msg.Subject(), "kind", kind)
			automationUnknownSubjectTotal.Add(1)
			processErr = errAutomationUnknownSubject
		}
	}
	if processErr != nil {
		if nakErr := msg.Nak(); nakErr != nil {
			slog.Error("automation: nak failed", "error", nakErr)
		}
		return
	}
	if err := msg.Ack(); err != nil {
		slog.Error("automation: ack failed", "error", err)
	}
}

func parseIssueEventKind(subject string) string {
	// project.{uuid}.issue.{kind}
	parts := strings.Split(subject, ".")
	if len(parts) < 4 || parts[0] != "project" || parts[2] != "issue" {
		return ""
	}
	return parts[3]
}
