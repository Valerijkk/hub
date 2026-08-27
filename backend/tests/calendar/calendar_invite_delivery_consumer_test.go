package calendar_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
	"github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/platform/events"
)

var (
	calendarInviteDeliveryConsumerTestSourceID    = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	calendarInviteDeliveryConsumerTestEventID     = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	calendarInviteDeliveryConsumerTestOrganizerID = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	calendarInviteDeliveryConsumerTestActorID     = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	calendarInviteDeliveryConsumerTestOrgID       = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	calendarInviteDeliveryConsumerTestOccurredAt  = time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
)

func TestProcessCalendarInviteDeliveryAdmissionsMapsV2Kinds(t *testing.T) {
	testCases := []struct {
		name   string
		subject string
		kind   mail.CalendarInviteDeliveryKind
		method mail.CalendarMailDeliveryMethod
	}{
		{name: "created", subject: calendar.CalendarInviteCreatedSubjectV2, kind: mail.CalendarInviteDeliveryKindCreated, method: mail.CalendarMailDeliveryMethodRequest},
		{name: "updated", subject: calendar.CalendarInviteUpdatedSubjectV2, kind: mail.CalendarInviteDeliveryKindUpdated, method: mail.CalendarMailDeliveryMethodRequest},
		{name: "cancelled", subject: calendar.CalendarInviteCancelledSubjectV2, kind: mail.CalendarInviteDeliveryKindCancelled, method: mail.CalendarMailDeliveryMethodCancel},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			envelope, encoded, payload := calendarInviteDeliveryConsumerTestEnvelope(t, testCase.subject, 1, 0)
			repository := &calendarInviteDeliveryConsumerRecordingRepository{}

			got := calendar.ProcessCalendarInviteDeliveryAdmissions(context.Background(), repository, testCase.subject, encoded)
			if got != calendar.CalendarInviteDeliveryConsumerDispositionAck {
				t.Fatalf("disposition = %v, want ACK", got)
			}
			if len(repository.requests) != 1 || len(repository.requests[0].Invites) != 1 {
				t.Fatalf("requests = %#v, want one one-recipient admission", repository.requests)
			}

			admission := repository.requests[0].Invites[0]
			recipient := payload.Deliveries[0]
			if admission.InviteID.UUID() != recipient.InviteID || admission.SourceEnvelopeID != envelope.EventID() || admission.EventID != payload.EventID ||
				admission.AttendeeID != recipient.AttendeeID || admission.OrganizerID != payload.OrganizerID || admission.Kind != testCase.kind ||
				admission.Method != testCase.method || admission.EventSequence != 0 || admission.RecurrenceID != "" {
				t.Fatalf("admission = %#v, payload = %#v", admission, payload)
			}
			if admission.OrganizerID == calendarInviteDeliveryConsumerTestActorID {
				t.Fatal("organizer must come from payload, never envelope actor")
			}
		})
	}
}

func TestProcessCalendarInviteDeliveryAdmissionsBatchesCanonicalDeliveries(t *testing.T) {
	testCases := []struct {
		name      string
		deliveries int
		batches   []int
	}{
		{name: "zero", deliveries: 0, batches: []int{}},
		{name: "fifty_one", deliveries: 51, batches: []int{50, 1}},
		{name: "two_hundred", deliveries: 200, batches: []int{50, 50, 50, 50}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, encoded, payload := calendarInviteDeliveryConsumerTestEnvelope(t, calendar.CalendarInviteCreatedSubjectV2, testCase.deliveries, 0)
			repository := &calendarInviteDeliveryConsumerRecordingRepository{}

			got := calendar.ProcessCalendarInviteDeliveryAdmissions(context.Background(), repository, calendar.CalendarInviteCreatedSubjectV2, encoded)
			if got != calendar.CalendarInviteDeliveryConsumerDispositionAck {
				t.Fatalf("disposition = %v, want ACK", got)
			}
			if len(repository.requests) != len(testCase.batches) {
				t.Fatalf("batch count = %d, want %d", len(repository.requests), len(testCase.batches))
			}
			var flattened []mail.CalendarInviteDeliveryAdmission
			for index, request := range repository.requests {
				if len(request.Invites) != testCase.batches[index] {
					t.Fatalf("batch %d length = %d, want %d", index, len(request.Invites), testCase.batches[index])
				}
				flattened = append(flattened, request.Invites...)
			}
			if len(flattened) != len(payload.Deliveries) {
				t.Fatalf("flattened admissions = %d, payload deliveries = %d", len(flattened), len(payload.Deliveries))
			}
			for index, recipient := range payload.Deliveries {
				if flattened[index].InviteID.UUID() != recipient.InviteID || flattened[index].AttendeeID != recipient.AttendeeID {
					t.Fatalf("admission %d lost canonical delivery order", index)
				}
			}
		})
	}
}

func TestProcessCalendarInviteDeliveryAdmissionsRejectsOversizedCanonicalEnvelope(t *testing.T) {
	encoded := calendarInviteDeliveryConsumerTestOversizedEnvelope(t)
	repository := &calendarInviteDeliveryConsumerRecordingRepository{}

	got := calendar.ProcessCalendarInviteDeliveryAdmissions(context.Background(), repository, calendar.CalendarInviteCreatedSubjectV2, encoded)
	if got != calendar.CalendarInviteDeliveryConsumerDispositionTerm {
		t.Fatalf("disposition = %v, want TERM", got)
	}
	if len(repository.requests) != 0 {
		t.Fatalf("repository called for oversized envelope: %#v", repository.requests)
	}
}

func TestProcessCalendarInviteDeliveryAdmissionsTerminatesInvalidEnvelopes(t *testing.T) {
	_, canonical, _ := calendarInviteDeliveryConsumerTestEnvelope(t, calendar.CalendarInviteCreatedSubjectV2, 1, 0)
	duplicate := calendarInviteDeliveryConsumerTestDuplicateEnvelope(t)
	testCases := []struct {
		name    string
		subject string
		encoded []byte
	}{
		{name: "noncanonical", subject: calendar.CalendarInviteCreatedSubjectV2, encoded: append([]byte(" "), canonical...)},
		{name: "subject_mismatch", subject: calendar.CalendarInviteUpdatedSubjectV2, encoded: canonical},
		{name: "duplicate_delivery", subject: calendar.CalendarInviteCreatedSubjectV2, encoded: duplicate},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &calendarInviteDeliveryConsumerRecordingRepository{}
			got := calendar.ProcessCalendarInviteDeliveryAdmissions(context.Background(), repository, testCase.subject, testCase.encoded)
			if got != calendar.CalendarInviteDeliveryConsumerDispositionTerm {
				t.Fatalf("disposition = %v, want TERM", got)
			}
			if len(repository.requests) != 0 {
				t.Fatalf("repository called for rejected envelope: %#v", repository.requests)
			}
		})
	}
}

func TestProcessCalendarInviteDeliveryAdmissionsReplaysPartialBatchOnTransientFailure(t *testing.T) {
	_, encoded, payload := calendarInviteDeliveryConsumerTestEnvelope(t, calendar.CalendarInviteCreatedSubjectV2, 51, 0)
	transient := calendarInviteDeliveryConsumerTestUnavailableError(t)
	repository := &calendarInviteDeliveryConsumerRecordingRepository{
		admitError: func(call int) error {
			if call == 2 {
				return transient
			}
			return nil
		},
	}

	got := calendar.ProcessCalendarInviteDeliveryAdmissions(context.Background(), repository, calendar.CalendarInviteCreatedSubjectV2, encoded)
	if got != calendar.CalendarInviteDeliveryConsumerDispositionNak {
		t.Fatalf("disposition = %v, want NAK", got)
	}
	if len(repository.requests) != 2 || len(repository.requests[0].Invites) != 50 || len(repository.requests[1].Invites) != 1 {
		t.Fatalf("partial requests = %#v", repository.requests)
	}
	for index, recipient := range payload.Deliveries {
		var admission mail.CalendarInviteDeliveryAdmission
		if index < 50 {
			admission = repository.requests[0].Invites[index]
		} else {
			admission = repository.requests[1].Invites[0]
		}
		if admission.InviteID.UUID() != recipient.InviteID {
			t.Fatalf("partial replay order changed at %d", index)
		}
	}
}

func TestProcessCalendarInviteDeliveryAdmissionsDispositions(t *testing.T) {
	_, encoded, _ := calendarInviteDeliveryConsumerTestEnvelope(t, calendar.CalendarInviteCreatedSubjectV2, 1, 0)
	invalidID, invalid := mail.NewCalendarInviteDeliveryID(uuid.Nil)
	if invalid == nil || invalidID.UUID() != uuid.Nil {
		t.Fatal("construct typed Invalid repository error")
	}

	testCases := []struct {
		name       string
		admitError error
		panicOnCall bool
		want       calendar.CalendarInviteDeliveryConsumerDisposition
	}{
		{name: "typed_invalid", admitError: invalid, want: calendar.CalendarInviteDeliveryConsumerDispositionTerm},
		{name: "typed_unavailable", admitError: calendarInviteDeliveryConsumerTestUnavailableError(t), want: calendar.CalendarInviteDeliveryConsumerDispositionNak},
		{name: "unknown", admitError: errors.New("unknown"), want: calendar.CalendarInviteDeliveryConsumerDispositionNak},
		{name: "deadline", admitError: context.DeadlineExceeded, want: calendar.CalendarInviteDeliveryConsumerDispositionNak},
		{name: "panic", panicOnCall: true, want: calendar.CalendarInviteDeliveryConsumerDispositionNak},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &calendarInviteDeliveryConsumerRecordingRepository{admitError: func(int) error { return testCase.admitError }, panicOnCall: testCase.panicOnCall}
			got := calendar.ProcessCalendarInviteDeliveryAdmissions(context.Background(), repository, calendar.CalendarInviteCreatedSubjectV2, encoded)
			if got != testCase.want {
				t.Fatalf("disposition = %v, want %v", got, testCase.want)
			}
			if len(repository.requests) != 1 {
				t.Fatalf("repository calls = %d, want 1", len(repository.requests))
			}
		})
	}
}

func TestCalendarInviteDeliveryConsumerConflictTerminalSourceContract(t *testing.T) {
	source := readCalendarInviteDeliveryConsumerSource(t)
	requireCalendarInviteDeliveryOrder(t, source,
		"return repositoryErr.Kind() == mail.CalendarMailDeliveryRepositoryErrorInvalid ||",
		"repositoryErr.Kind() == mail.CalendarMailDeliveryRepositoryErrorConflict",
	)
}

func TestCalendarInviteDeliveryConsumerLifecycleSourceContract(t *testing.T) {
	source := readCalendarInviteDeliveryConsumerSource(t)
	compactSource := strings.NewReplacer(" ", "", "\t", "").Replace(source)

	requireCalendarInviteDeliveryConsumerContains(t, source,
		"CalendarInviteCreatedSubjectV2",
		"CalendarInviteUpdatedSubjectV2",
		"CalendarInviteCancelledSubjectV2",
		"calendar-mail-outbound-invite-v2",
		"FilterSubjects:",
		"jetstream.AckExplicitPolicy",
		"bytes.Equal(encoded, canonical)",
		"decoder.DisallowUnknownFields()",
	)
	requireCalendarInviteDeliveryConsumerContains(t, compactSource,
		"calendarInviteDeliveryConsumerPullWindow=8",
		"calendarInviteDeliveryConsumerWorkerConcurrency=4",
		"calendarInviteDeliveryConsumerMaxAckPending=16",
		"calendarInviteDeliveryConsumerAckWait=90*time.Second",
		"calendarInviteDeliveryConsumerDeliveryTimeout=45*time.Second",
		"calendarInviteDeliveryConsumerNakDelay=65*time.Second",
	)
	if strings.Contains(source, "MaxDeliver:") {
		t.Fatal("consumer must keep JetStream redelivery unlimited")
	}
	for _, legacySubject := range []string{
		"calendar.invite.created.v1",
		"calendar.invite.updated.v1",
		"calendar.invite.cancelled.v1",
	} {
		if strings.Contains(source, legacySubject) {
			t.Fatalf("consumer must not admit legacy v1 subject: %s", legacySubject)
		}
	}
}

type calendarInviteDeliveryConsumerRecordingRepository struct {
	requests    []mail.CalendarInviteDeliveryAdmissionRequest
	admitError  func(int) error
	panicOnCall bool
}

var _ mail.CalendarMailDeliveryRepository = (*calendarInviteDeliveryConsumerRecordingRepository)(nil)

func (repository *calendarInviteDeliveryConsumerRecordingRepository) AdmitCalendarInviteDeliveries(_ context.Context, request mail.CalendarInviteDeliveryAdmissionRequest) error {
	copied := mail.CalendarInviteDeliveryAdmissionRequest{Invites: append([]mail.CalendarInviteDeliveryAdmission(nil), request.Invites...)}
	repository.requests = append(repository.requests, copied)
	if repository.panicOnCall {
		panic("admission panic")
	}
	if repository.admitError == nil {
		return nil
	}
	return repository.admitError(len(repository.requests))
}

func (*calendarInviteDeliveryConsumerRecordingRepository) WakeCalendarReplyDelivery(context.Context, mail.CalendarReplyDeliveryWakeRequest) error {
	return nil
}

func (*calendarInviteDeliveryConsumerRecordingRepository) ScanDueCalendarMailDeliveries(context.Context, mail.CalendarMailDeliveryDueScanRequest) (mail.CalendarMailDeliveryDueScanResult, error) {
	return mail.CalendarMailDeliveryDueScanResult{}, nil
}

func (*calendarInviteDeliveryConsumerRecordingRepository) ClaimCalendarMailDelivery(context.Context, mail.CalendarMailDeliveryID) (mail.CalendarMailDeliveryClaimResult, error) {
	return mail.CalendarMailDeliveryClaimResult{}, nil
}

func (*calendarInviteDeliveryConsumerRecordingRepository) BindCalendarMailDeliverySubmission(context.Context, mail.CalendarMailDeliverySubmissionBindRequest) (mail.CalendarMailDeliverySubmissionBindResult, error) {
	return mail.CalendarMailDeliverySubmissionBindResult{}, nil
}

func (*calendarInviteDeliveryConsumerRecordingRepository) ResolveCalendarMailDelivery(context.Context, mail.CalendarMailDeliveryResolveRequest) error {
	return nil
}

func calendarInviteDeliveryConsumerTestEnvelope(t *testing.T, subject string, deliveries int, sequence uint) (events.EnvelopeV1, []byte, calendar.CalendarInviteDeliveryV2Payload) {
	t.Helper()
	attendees := calendarInviteDeliveryConsumerTestAttendeeIDs(deliveries)
	actorID := calendarInviteDeliveryConsumerTestActorID
	envelope, err := calendar.NewCalendarInviteDeliveryEnvelopeV2(subject, calendarInviteDeliveryConsumerTestSourceID, calendarInviteDeliveryConsumerTestOccurredAt,
		calendarInviteDeliveryConsumerTestOrgID, &actorID, calendarInviteDeliveryConsumerTestEventID.String(), calendar.CalendarInviteDeliveryEnvelopeV2Input{
			EventID:       calendarInviteDeliveryConsumerTestEventID,
			EventSequence: sequence,
			OrganizerID:   calendarInviteDeliveryConsumerTestOrganizerID,
			AttendeeIDs:   attendees,
		})
	if err != nil {
		t.Fatalf("build v2 envelope: %v", err)
	}
	encoded, err := envelope.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal v2 envelope: %v", err)
	}
	payload, err := calendar.DecodeCalendarInviteDeliveryEnvelopeV2(envelope)
	if err != nil {
		t.Fatalf("decode v2 envelope: %v", err)
	}
	return envelope, encoded, payload
}

func calendarInviteDeliveryConsumerTestOversizedEnvelope(t *testing.T) []byte {
	t.Helper()
	payload := calendarInviteDeliveryConsumerTestPayload(calendarInviteDeliveryConsumerTestAttendeeIDs(201))
	actorID := calendarInviteDeliveryConsumerTestActorID
	envelope, err := events.NewEnvelopeV1(calendar.CalendarInviteCreatedSubjectV2, calendarInviteDeliveryConsumerTestSourceID,
		calendarInviteDeliveryConsumerTestOccurredAt, calendarInviteDeliveryConsumerTestOrgID, &actorID,
		calendarInviteDeliveryConsumerTestEventID.String(), payload)
	if err != nil {
		t.Fatalf("build oversized canonical envelope: %v", err)
	}
	encoded, err := envelope.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal oversized canonical envelope: %v", err)
	}
	return encoded
}

func calendarInviteDeliveryConsumerTestDuplicateEnvelope(t *testing.T) []byte {
	t.Helper()
	attendee := calendarInviteDeliveryConsumerTestAttendeeIDs(1)[0]
	payload := calendar.CalendarInviteDeliveryV2Payload{
		EventID:       calendarInviteDeliveryConsumerTestEventID,
		EventSequence: 0,
		OrganizerID:   calendarInviteDeliveryConsumerTestOrganizerID,
		Deliveries: []calendar.CalendarInviteDeliveryV2Recipient{
			{AttendeeID: attendee, InviteID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("duplicate"))},
			{AttendeeID: attendee, InviteID: uuid.NewSHA1(uuid.NameSpaceDNS, []byte("duplicate"))},
		},
	}
	actorID := calendarInviteDeliveryConsumerTestActorID
	envelope, err := events.NewEnvelopeV1(calendar.CalendarInviteCreatedSubjectV2, calendarInviteDeliveryConsumerTestSourceID,
		calendarInviteDeliveryConsumerTestOccurredAt, calendarInviteDeliveryConsumerTestOrgID, &actorID,
		calendarInviteDeliveryConsumerTestEventID.String(), payload)
	if err != nil {
		t.Fatalf("build duplicate canonical envelope: %v", err)
	}
	encoded, err := envelope.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal duplicate canonical envelope: %v", err)
	}
	return encoded
}

func calendarInviteDeliveryConsumerTestPayload(attendees []uuid.UUID) calendar.CalendarInviteDeliveryV2Payload {
	namespace := uuid.NewSHA1(uuid.NameSpaceURL, []byte("hub.calendar.invite-delivery.id.v1"))
	deliveries := make([]calendar.CalendarInviteDeliveryV2Recipient, len(attendees))
	for index, attendeeID := range attendees {
		deliveries[index] = calendar.CalendarInviteDeliveryV2Recipient{
			AttendeeID: attendeeID,
			InviteID:   uuid.NewSHA1(namespace, []byte(calendarInviteDeliveryConsumerTestSourceID.String()+"/"+attendeeID.String())),
		}
	}
	sort.Slice(deliveries, func(left, right int) bool {
		return bytes.Compare(deliveries[left].InviteID[:], deliveries[right].InviteID[:]) < 0
	})
	return calendar.CalendarInviteDeliveryV2Payload{
		EventID:       calendarInviteDeliveryConsumerTestEventID,
		EventSequence: 0,
		OrganizerID:   calendarInviteDeliveryConsumerTestOrganizerID,
		Deliveries:    deliveries,
	}
}

func calendarInviteDeliveryConsumerTestAttendeeIDs(count int) []uuid.UUID {
	attendees := make([]uuid.UUID, count)
	for index := range attendees {
		attendees[index] = uuid.NewSHA1(uuid.NameSpaceOID, []byte("calendar-invite-delivery-attendee/"+strconv.Itoa(index)))
	}
	return attendees
}

func calendarInviteDeliveryConsumerTestUnavailableError(t *testing.T) error {
	t.Helper()
	fake := mail.NewCalendarMailDeliveryRepositoryFake()
	first, err := mail.NewCalendarMailDeliveryID(mail.CalendarMailDeliveryKindInvite, uuid.NewSHA1(uuid.NameSpaceDNS, []byte("unavailable-first")))
	if err != nil {
		t.Fatalf("build first due id: %v", err)
	}
	second, err := mail.NewCalendarMailDeliveryID(mail.CalendarMailDeliveryKindInvite, uuid.NewSHA1(uuid.NameSpaceDNS, []byte("unavailable-second")))
	if err != nil {
		t.Fatalf("build second due id: %v", err)
	}
	if err := fake.SetDue([]mail.CalendarMailDeliveryID{first, second}); err != nil {
		t.Fatalf("configure unavailable fake: %v", err)
	}
	_, err = fake.ScanDueCalendarMailDeliveries(context.Background(), mail.CalendarMailDeliveryDueScanRequest{Limit: 1})
	if err == nil {
		t.Fatal("construct typed Unavailable repository error")
	}
	return err
}

func readCalendarInviteDeliveryConsumerSource(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("determine test file path")
	}
	source, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "..", "internal", "calendar", "calendar_invite_delivery_consumer.go"))
	if err != nil {
		t.Fatalf("read consumer source: %v", err)
	}
	return string(source)
}

func requireCalendarInviteDeliveryConsumerContains(t *testing.T, source string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(source, needle) {
			t.Fatalf("consumer source missing %q", needle)
		}
	}
}

func requireCalendarInviteDeliveryOrder(t *testing.T, source string, needles ...string) {
	t.Helper()
	offset := 0
	for _, needle := range needles {
		next := strings.Index(source[offset:], needle)
		if next < 0 {
			t.Fatalf("consumer source missing ordered token %q", needle)
		}
		offset += next + len(needle)
	}
}
