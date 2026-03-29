// Package nats управляет подключением к NATS и JetStream-стримами.
package nats

import (
	"context"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// StreamName — единый JetStream stream для всех модулей.
const StreamName = "EVENTS"

// Client оборачивает NATS-соединение с поддержкой JetStream.
type Client struct {
	Conn      *nats.Conn
	JetStream jetstream.JetStream
}

// NewClient подключается к NATS и инициализирует JetStream.
func NewClient(ctx context.Context, url string) (*Client, error) {
	nc, err := nats.Connect(url,
		nats.Name("hub-backend"),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, err
	}

	slog.Info("connected to NATS", "url", url)
	return &Client{Conn: nc, JetStream: js}, nil
}

// Close закрывает NATS-соединение.
func (c *Client) Close() {
	c.Conn.Close()
}

// Publish публикует событие в core NATS (эфемерное, fire-and-forget).
// Используется для: индикаторов набора, присутствия, обновлений UI в реальном времени.
func (c *Client) Publish(subject string, data []byte) error {
	return c.Conn.Publish(subject, data)
}

// EnsureStreams создаёт единый JetStream stream EVENTS, агрегирующий события
// всех модулей. Каждый consumer (audit, webhook, notification) создаёт свой
// durable consumer на этом стриме.
func (c *Client) EnsureStreams(ctx context.Context) error {
	_, err := c.JetStream.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name: StreamName,
		Subjects: []string{
			"project.>",
			"tracker.>",
			"messenger.channel.>",
			"tms.>",
			"calls.channel.>",
			"auth.session.>",
			"sprint.>",
		},
		Retention: jetstream.InterestPolicy,
		MaxAge:    7 * 24 * time.Hour,
	})
	if err != nil {
		return err
	}
	slog.Info("JetStream stream ensured", "stream", StreamName)
	return nil
}

// Subscribe подписывается на core NATS subject (эфемерная подписка).
func (c *Client) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return c.Conn.Subscribe(subject, handler)
}
