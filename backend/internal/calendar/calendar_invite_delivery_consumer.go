package calendar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/platform/events"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

const (
	calendarInviteDeliveryConsumerDurable            = "calendar-mail-outbound-invite-v2"
	calendarInviteDeliveryConsumerAckWait            = 90 * time.Second
	calendarInviteDeliveryConsumerMaxAckPending      = 16
	calendarInviteDeliveryConsumerPullWindow         = 8
	calendarInviteDeliveryConsumerWorkerConcurrency  = 4
	calendarInviteDeliveryConsumerDeliveryTimeout    = 45 * time.Second
	calendarInviteDeliveryConsumerNakDelay           = 65 * time.Second
	calendarInviteDeliveryConsumerRetryInitialDelay  = time.Second
	calendarInviteDeliveryConsumerRetryMaximumDelay  = 30 * time.Second
	calendarInviteDeliveryConsumerMaximumDeliveries  = 200
	calendarInviteDeliveryConsumerAdmissionBatchSize = 50
)

var (
	calendarInviteDeliveryConsumerSubjects = []string{
		CalendarInviteCreatedSubjectV2,
		CalendarInviteUpdatedSubjectV2,
		CalendarInviteCancelledSubjectV2,
	}
	errCalendarInviteDeliveryConsumerInvalid     = errors.New("calendar invite delivery consumer invalid")
	errCalendarInviteDeliveryConsumerUnavailable = errors.New("calendar invite delivery consumer unavailable")
)

// CalendarInviteDeliveryConsumer admits canonical v2 Calendar invite intents into Mail.
type CalendarInviteDeliveryConsumer struct {
	client     *natsclient.Client
	repository mail.CalendarMailDeliveryRepository

	mu      sync.Mutex
	started bool
	running bool
	run     *calendarInviteDeliveryConsumerRun
}

type calendarInviteDeliveryConsumerRun struct {
	cancel     context.CancelFunc
	iterator   jetstream.MessagesContext
	done       chan struct{}
	stopParent func() bool
	workers    sync.WaitGroup
}

type calendarInviteDeliveryConsumerEnvelopeJSON struct {
	Subject        string          `json:"subject"`
	SchemaVersion  int             `json:"schema_version"`
	EventID        uuid.UUID       `json:"event_id"`
	OccurredAt     time.Time       `json:"occurred_at"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	ActorID        *uuid.UUID      `json:"actor_id"`
	CorrelationID  string          `json:"correlation_id"`
	Payload        json.RawMessage `json:"payload"`
}

// NewCalendarInviteDeliveryConsumer constructs the Calendar-owned v2-only admission consumer.
func NewCalendarInviteDeliveryConsumer(client *natsclient.Client, repository mail.CalendarMailDeliveryRepository) (*CalendarInviteDeliveryConsumer, error) {
	if client == nil || calendarInviteDeliveryConsumerDependencyNil(client.JetStream) || calendarInviteDeliveryConsumerDependencyNil(repository) {
		return nil, errCalendarInviteDeliveryConsumerInvalid
	}
	return &CalendarInviteDeliveryConsumer{client: client, repository: repository}, nil
}

// Start creates the fixed durable pull consumer and starts its bounded workers.
func (consumer *CalendarInviteDeliveryConsumer) Start(ctx context.Context) error {
	if consumer == nil || calendarInviteDeliveryConsumerDependencyNil(ctx) {
		return errCalendarInviteDeliveryConsumerInvalid
	}

	consumer.mu.Lock()
	defer consumer.mu.Unlock()
	if consumer.started {
		return nil
	}
	if consumer.running || consumer.run != nil || consumer.client == nil || calendarInviteDeliveryConsumerDependencyNil(consumer.client.JetStream) ||
		calendarInviteDeliveryConsumerDependencyNil(consumer.repository) || ctx.Err() != nil {
		return errCalendarInviteDeliveryConsumerInvalid
	}
	consumer.running = true

	runContext, cancel := context.WithCancel(ctx)
	stream, err := consumer.client.JetStream.Stream(runContext, natsclient.StreamName)
	if err != nil {
		cancel()
		consumer.running = false
		return errCalendarInviteDeliveryConsumerUnavailable
	}
	jetStreamConsumer, err := stream.CreateOrUpdateConsumer(runContext, jetstream.ConsumerConfig{
		Durable:        calendarInviteDeliveryConsumerDurable,
		FilterSubjects: append([]string(nil), calendarInviteDeliveryConsumerSubjects...),
		AckPolicy:      jetstream.AckExplicitPolicy,
		AckWait:        calendarInviteDeliveryConsumerAckWait,
		MaxAckPending:  calendarInviteDeliveryConsumerMaxAckPending,
	})
	if err != nil {
		cancel()
		consumer.running = false
		return errCalendarInviteDeliveryConsumerUnavailable
	}
	iterator, err := jetStreamConsumer.Messages(jetstream.PullMaxMessages(calendarInviteDeliveryConsumerPullWindow))
	if err != nil {
		cancel()
		consumer.running = false
		return errCalendarInviteDeliveryConsumerUnavailable
	}
	if runContext.Err() != nil {
		iterator.Stop()
		cancel()
		consumer.running = false
		return errCalendarInviteDeliveryConsumerInvalid
	}

	run := &calendarInviteDeliveryConsumerRun{
		cancel:   cancel,
		iterator: iterator,
		done:     make(chan struct{}),
	}
	run.stopParent = context.AfterFunc(runContext, iterator.Stop)
	consumer.run = run
	consumer.started = true

	jobs := make(chan jetstream.Msg, calendarInviteDeliveryConsumerPullWindow)
	for range calendarInviteDeliveryConsumerWorkerConcurrency {
		run.workers.Add(1)
		go consumer.runWorker(runContext, run, jobs)
	}
	go consumer.runDispatcher(runContext, run, jobs)
	slog.Info("calendar invite delivery consumer started")
	return nil
}

// Stop cancels delivery, stops the iterator, and waits for the started run.
func (consumer *CalendarInviteDeliveryConsumer) Stop() {
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

func (consumer *CalendarInviteDeliveryConsumer) runDispatcher(ctx context.Context, run *calendarInviteDeliveryConsumerRun, jobs chan<- jetstream.Msg) {
	defer func() {
		if recover() != nil {
			slog.Error("calendar invite delivery consumer dispatcher failed")
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

	retryDelay := calendarInviteDeliveryConsumerRetryInitialDelay
	for {
		if ctx.Err() != nil {
			return
		}
		message, err := run.iterator.Next()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, jetstream.ErrMsgIteratorClosed) {
				return
			}
			slog.Error("calendar invite delivery consumer iterator failed")
			if !calendarInviteDeliveryConsumerWait(ctx, retryDelay) {
				return
			}
			retryDelay = calendarInviteDeliveryConsumerNextRetryDelay(retryDelay)
			continue
		}
		retryDelay = calendarInviteDeliveryConsumerRetryInitialDelay
		select {
		case jobs <- message:
		case <-ctx.Done():
			return
		}
	}
}

func (consumer *CalendarInviteDeliveryConsumer) runWorker(ctx context.Context, run *calendarInviteDeliveryConsumerRun, jobs <-chan jetstream.Msg) {
	defer run.workers.Done()
	for message := range jobs {
		consumer.handleMessageSafely(ctx, message)
	}
}

func (consumer *CalendarInviteDeliveryConsumer) finishRun(run *calendarInviteDeliveryConsumerRun) {
	consumer.mu.Lock()
	if consumer.run == run {
		consumer.run = nil
		consumer.running = false
	}
	consumer.mu.Unlock()
	close(run.done)
}

func (consumer *CalendarInviteDeliveryConsumer) handleMessageSafely(ctx context.Context, message jetstream.Msg) {
	defer func() {
		if recover() != nil {
			calendarInviteDeliveryConsumerNak(message)
		}
	}()
	consumer.handleMessage(ctx, message)
}

func (consumer *CalendarInviteDeliveryConsumer) handleMessage(ctx context.Context, message jetstream.Msg) {
	admissions, err := calendarInviteDeliveryConsumerAdmissions(message.Subject(), message.Data())
	if err != nil {
		calendarInviteDeliveryConsumerTerm(message)
		return
	}
	if len(admissions) == 0 {
		calendarInviteDeliveryConsumerAck(message)
		return
	}

	deliveryContext := natsclient.ExtractTraceContext(ctx, message.Headers())
	deliveryContext, cancel := context.WithTimeout(deliveryContext, calendarInviteDeliveryConsumerDeliveryTimeout)
	defer cancel()
	for start := 0; start < len(admissions); start += calendarInviteDeliveryConsumerAdmissionBatchSize {
		end := min(start+calendarInviteDeliveryConsumerAdmissionBatchSize, len(admissions))
		if err := consumer.repository.AdmitCalendarInviteDeliveries(deliveryContext, mail.CalendarInviteDeliveryAdmissionRequest{Invites: admissions[start:end]}); err != nil {
			if calendarInviteDeliveryConsumerRepositoryTerminal(err) {
				calendarInviteDeliveryConsumerTerm(message)
				return
			}
			calendarInviteDeliveryConsumerNak(message)
			return
		}
		if deliveryContext.Err() != nil {
			calendarInviteDeliveryConsumerNak(message)
			return
		}
	}
	calendarInviteDeliveryConsumerAck(message)
}

func calendarInviteDeliveryConsumerAdmissions(subject string, encoded []byte) ([]mail.CalendarInviteDeliveryAdmission, error) {
	kind, method, valid := calendarInviteDeliveryConsumerKindMethod(subject)
	if !valid {
		return nil, errCalendarInviteDeliveryConsumerInvalid
	}
	envelope, err := calendarInviteDeliveryConsumerDecodeEnvelope(subject, encoded)
	if err != nil {
		return nil, errCalendarInviteDeliveryConsumerInvalid
	}
	payload, err := DecodeCalendarInviteDeliveryEnvelopeV2(envelope)
	if err != nil || len(payload.Deliveries) > calendarInviteDeliveryConsumerMaximumDeliveries || payload.EventSequence > uint(^uint(0)>>1) {
		return nil, errCalendarInviteDeliveryConsumerInvalid
	}

	recurrenceID := ""
	if payload.RecurrenceID != nil {
		recurrenceID = *payload.RecurrenceID
	}
	admissions := make([]mail.CalendarInviteDeliveryAdmission, len(payload.Deliveries))
	for index, recipient := range payload.Deliveries {
		inviteID, err := mail.NewCalendarInviteDeliveryID(recipient.InviteID)
		if err != nil {
			return nil, errCalendarInviteDeliveryConsumerInvalid
		}
		admissions[index] = mail.CalendarInviteDeliveryAdmission{
			InviteID:         inviteID,
			SourceEnvelopeID: envelope.EventID(),
			Kind:             kind,
			EventID:          payload.EventID,
			AttendeeID:       recipient.AttendeeID,
			OrganizerID:      payload.OrganizerID,
			Method:           method,
			EventSequence:    int(payload.EventSequence),
			RecurrenceID:     recurrenceID,
		}
	}
	return admissions, nil
}

func calendarInviteDeliveryConsumerKindMethod(subject string) (mail.CalendarInviteDeliveryKind, mail.CalendarMailDeliveryMethod, bool) {
	switch subject {
	case CalendarInviteCreatedSubjectV2:
		return mail.CalendarInviteDeliveryKindCreated, mail.CalendarMailDeliveryMethodRequest, true
	case CalendarInviteUpdatedSubjectV2:
		return mail.CalendarInviteDeliveryKindUpdated, mail.CalendarMailDeliveryMethodRequest, true
	case CalendarInviteCancelledSubjectV2:
		return mail.CalendarInviteDeliveryKindCancelled, mail.CalendarMailDeliveryMethodCancel, true
	default:
		return "", "", false
	}
}

func calendarInviteDeliveryConsumerDecodeEnvelope(subject string, encoded []byte) (events.EnvelopeV1, error) {
	var value calendarInviteDeliveryConsumerEnvelopeJSON
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil || decoder.Decode(&struct{}{}) != io.EOF || value.Subject != subject {
		return events.EnvelopeV1{}, errCalendarInviteDeliveryConsumerInvalid
	}
	canonicalEnvelope, err := events.NewEnvelopeV1(value.Subject, value.EventID, value.OccurredAt, value.OrganizationID, calendarInviteDeliveryConsumerCopyUUID(value.ActorID), value.CorrelationID, value.Payload)
	if err != nil {
		return events.EnvelopeV1{}, errCalendarInviteDeliveryConsumerInvalid
	}
	canonical, err := canonicalEnvelope.MarshalJSON()
	if err != nil || !bytes.Equal(encoded, canonical) {
		return events.EnvelopeV1{}, errCalendarInviteDeliveryConsumerInvalid
	}
	return canonicalEnvelope, nil
}

func calendarInviteDeliveryConsumerCopyUUID(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func calendarInviteDeliveryConsumerRepositoryTerminal(err error) bool {
	var repositoryErr *mail.CalendarMailDeliveryRepositoryError
	if !errors.As(err, &repositoryErr) {
		return false
	}
	return repositoryErr.Kind() == mail.CalendarMailDeliveryRepositoryErrorInvalid ||
		repositoryErr.Kind() == mail.CalendarMailDeliveryRepositoryErrorConflict
}

func calendarInviteDeliveryConsumerWait(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func calendarInviteDeliveryConsumerNextRetryDelay(delay time.Duration) time.Duration {
	if delay >= calendarInviteDeliveryConsumerRetryMaximumDelay/2 {
		return calendarInviteDeliveryConsumerRetryMaximumDelay
	}
	return delay * 2
}

func calendarInviteDeliveryConsumerAck(message jetstream.Msg) {
	if err := message.Ack(); err != nil {
		slog.Error("calendar invite delivery consumer acknowledgement failed")
	}
}

func calendarInviteDeliveryConsumerTerm(message jetstream.Msg) {
	if err := message.Term(); err != nil {
		slog.Error("calendar invite delivery consumer termination failed")
	}
}

func calendarInviteDeliveryConsumerNak(message jetstream.Msg) {
	if err := message.NakWithDelay(calendarInviteDeliveryConsumerNakDelay); err != nil {
		slog.Error("calendar invite delivery consumer retry acknowledgement failed")
	}
}

func calendarInviteDeliveryConsumerDependencyNil(value any) bool {
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
