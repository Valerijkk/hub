package mail

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

const (
	calendarInboundConsumerSubject           = "mail.received.v1"
	calendarInboundConsumerDurable           = "calendar-mail-inbound-v1"
	calendarInboundConsumerAckWait           = 90 * time.Second
	calendarInboundConsumerMaxAckPending     = 16
	calendarInboundConsumerMaxDeliver        = 20
	calendarInboundConsumerPullWindow        = 8
	calendarInboundConsumerWorkerConcurrency = 4
	calendarInboundConsumerDeliveryTimeout   = 45 * time.Second
	calendarInboundConsumerNakDelay          = 65 * time.Second
)

var (
	errCalendarInboundConsumerInvalid     = errors.New("calendar inbound consumer invalid")
	errCalendarInboundConsumerUnavailable = errors.New("calendar inbound consumer unavailable")
)

// CalendarInboundConsumer dispatches canonical inbound Calendar envelope receipts.
type CalendarInboundConsumer struct {
	client *natsclient.Client
	sink   CalendarInboundSink

	mu      sync.Mutex
	started bool
	running bool
	run     *calendarInboundConsumerRun
}

type calendarInboundConsumerRun struct {
	cancel     context.CancelFunc
	iterator   jetstream.MessagesContext
	done       chan struct{}
	stopParent func() bool
	workers    sync.WaitGroup
}

// NewCalendarInboundConsumer constructs the sealed Mail-owned inbound Calendar consumer.
func NewCalendarInboundConsumer(client *natsclient.Client, sink CalendarInboundSink) (*CalendarInboundConsumer, error) {
	if client == nil || calendarInboundConsumerDependencyNil(client.JetStream) || calendarInboundConsumerDependencyNil(sink) {
		return nil, errCalendarInboundConsumerInvalid
	}
	return &CalendarInboundConsumer{client: client, sink: sink}, nil
}

// Start creates the fixed durable pull consumer and starts its bounded workers.
func (consumer *CalendarInboundConsumer) Start(ctx context.Context) error {
	if consumer == nil || calendarInboundConsumerDependencyNil(ctx) {
		return errCalendarInboundConsumerInvalid
	}

	consumer.mu.Lock()
	defer consumer.mu.Unlock()
	if consumer.started || consumer.running || consumer.run != nil || consumer.client == nil || calendarInboundConsumerDependencyNil(consumer.client.JetStream) || calendarInboundConsumerDependencyNil(consumer.sink) || ctx.Err() != nil {
		return errCalendarInboundConsumerInvalid
	}
	consumer.running = true

	runCtx, cancel := context.WithCancel(ctx)
	stream, err := consumer.client.JetStream.Stream(runCtx, natsclient.StreamName)
	if err != nil {
		cancel()
		consumer.running = false
		return errCalendarInboundConsumerUnavailable
	}
	jetStreamConsumer, err := stream.CreateOrUpdateConsumer(runCtx, jetstream.ConsumerConfig{
		Durable:       calendarInboundConsumerDurable,
		FilterSubject: calendarInboundConsumerSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       calendarInboundConsumerAckWait,
		MaxAckPending: calendarInboundConsumerMaxAckPending,
		MaxDeliver:    calendarInboundConsumerMaxDeliver,
	})
	if err != nil {
		cancel()
		consumer.running = false
		return errCalendarInboundConsumerUnavailable
	}
	iterator, err := jetStreamConsumer.Messages(jetstream.PullMaxMessages(calendarInboundConsumerPullWindow))
	if err != nil {
		cancel()
		consumer.running = false
		return errCalendarInboundConsumerUnavailable
	}
	if runCtx.Err() != nil {
		iterator.Stop()
		cancel()
		consumer.running = false
		return errCalendarInboundConsumerInvalid
	}

	run := &calendarInboundConsumerRun{
		cancel:   cancel,
		iterator: iterator,
		done:     make(chan struct{}),
	}
	run.stopParent = context.AfterFunc(runCtx, iterator.Stop)
	consumer.run = run
	consumer.started = true

	jobs := make(chan jetstream.Msg, calendarInboundConsumerPullWindow)
	for range calendarInboundConsumerWorkerConcurrency {
		run.workers.Add(1)
		go consumer.runWorker(runCtx, run, jobs)
	}
	go consumer.runDispatcher(runCtx, run, jobs)
	slog.Info("calendar inbound consumer started")
	return nil
}

// Stop cancels delivery, stops the iterator, and waits for the started run.
func (consumer *CalendarInboundConsumer) Stop() {
	if consumer == nil {
		return
	}
	consumer.mu.Lock()
	run := consumer.run
	consumer.mu.Unlock()
	if run == nil {
		return
	}
	run.cancel()
	run.iterator.Stop()
	<-run.done
}

func (consumer *CalendarInboundConsumer) runDispatcher(ctx context.Context, run *calendarInboundConsumerRun, jobs chan<- jetstream.Msg) {
	defer func() {
		if recover() != nil {
			slog.Error("calendar inbound consumer dispatcher failed")
		}
		close(jobs)
		run.workers.Wait()
		run.cancel()
		run.iterator.Stop()
		if run.stopParent != nil {
			run.stopParent()
		}
		consumer.finishRun(run)
	}()

	for {
		if ctx.Err() != nil {
			return
		}
		message, err := run.iterator.Next()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, jetstream.ErrMsgIteratorClosed) {
				return
			}
			slog.Error("calendar inbound consumer iterator failed")
			continue
		}
		select {
		case jobs <- message:
		case <-ctx.Done():
			return
		}
	}
}

func (consumer *CalendarInboundConsumer) runWorker(ctx context.Context, run *calendarInboundConsumerRun, jobs <-chan jetstream.Msg) {
	defer run.workers.Done()
	for message := range jobs {
		consumer.handleMessageSafely(ctx, message)
	}
}

func (consumer *CalendarInboundConsumer) finishRun(run *calendarInboundConsumerRun) {
	consumer.mu.Lock()
	if consumer.run == run {
		consumer.run = nil
		consumer.running = false
	}
	consumer.mu.Unlock()
	close(run.done)
}

func calendarInboundConsumerDependencyNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func (consumer *CalendarInboundConsumer) handleMessageSafely(ctx context.Context, message jetstream.Msg) {
	defer func() {
		if recover() != nil {
			calendarInboundConsumerNak(message)
		}
	}()
	consumer.handleMessage(ctx, message)
}

func (consumer *CalendarInboundConsumer) handleMessage(ctx context.Context, message jetstream.Msg) {
	result, err := decodeCalendarInboundEvent(message.Subject(), message.Data())
	if err != nil {
		calendarInboundConsumerTerm(message)
		return
	}
	if result.noCalendarPart {
		calendarInboundConsumerAck(message)
		return
	}

	deliveryContext := natsclient.ExtractTraceContext(ctx, message.Headers())
	deliveryContext, cancel := context.WithTimeout(deliveryContext, calendarInboundConsumerDeliveryTimeout)
	defer cancel()
	if err := consumer.sink.HandleCalendarInbound(deliveryContext, result.delivery); err != nil || deliveryContext.Err() != nil {
		calendarInboundConsumerNak(message)
		return
	}
	calendarInboundConsumerAck(message)
}

func calendarInboundConsumerAck(message jetstream.Msg) {
	if err := message.Ack(); err != nil {
		slog.Error("calendar inbound consumer acknowledgement failed")
	}
}

func calendarInboundConsumerTerm(message jetstream.Msg) {
	if err := message.Term(); err != nil {
		slog.Error("calendar inbound consumer termination failed")
	}
}

func calendarInboundConsumerNak(message jetstream.Msg) {
	if err := message.NakWithDelay(calendarInboundConsumerNakDelay); err != nil {
		slog.Error("calendar inbound consumer retry acknowledgement failed")
	}
}
