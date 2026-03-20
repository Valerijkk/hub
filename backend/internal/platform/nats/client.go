package nats

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Client wraps a NATS connection with JetStream support.
type Client struct {
	Conn      *nats.Conn
	JetStream jetstream.JetStream
}

// NewClient connects to NATS and initializes JetStream.
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

// Close closes the NATS connection.
func (c *Client) Close() {
	c.Conn.Close()
}

// Publish publishes an event to core NATS (ephemeral, fire-and-forget).
// Use for: typing indicators, presence, real-time UI updates.
func (c *Client) Publish(subject string, data []byte) error {
	return c.Conn.Publish(subject, data)
}

// Subscribe subscribes to a core NATS subject (ephemeral).
func (c *Client) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return c.Conn.Subscribe(subject, handler)
}
