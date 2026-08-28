package calendar_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
	"github.com/valerii/hub/backend/internal/platform/events"
)

var (
	inboundReplyOwnerID        = uuid.MustParse("d1111111-1111-4111-8111-111111111111")
	inboundReplyOrganizationID = uuid.MustParse("d2222222-2222-4222-8222-222222222222")
	inboundReplyEventID        = uuid.MustParse("d3333333-3333-4333-8333-333333333333")
	inboundReplyAttendeeID     = uuid.MustParse("d4444444-4444-4444-8444-444444444444")
	inboundReplyEnvelopeID     = uuid.MustParse("d5555555-5555-4555-8555-555555555555")
)

var _ calendar.EventRepository = (*inboundReplyRepositoryFake)(nil)
var _ calendar.EventTransactionBackend = (*inboundReplyBackendFake)(nil)
var _ calendar.CalendarInboundReplyTransactionBackend = (*inboundReplyBackendFake)(nil)

func TestCalendarInboundReply_FrozenABIAndSenderBoundReceiptVector(t *testing.T) {
	assertInboundReplyStructABI(t, reflect.TypeOf(calendar.CalendarInboundReplyInput{}), []string{"OwnerID", "Sender", "Message"})
	assertInboundReplyStructABI(t, reflect.TypeOf(calendar.CalendarInboundReplyAuthorizationRequest{}), []string{"UID", "HasRecurrenceID", "RecurrenceID", "Sender", "EventSequence", "Receipt"})
	assertInboundReplyStructABI(t, reflect.TypeOf(calendar.CalendarInboundReplyAttendeeSnapshot{}), []string{"AttendeeID", "EventID", "Email", "Status", "RecurrenceID", "RespondedAt", "Version", "RemovedAt"})
	assertInboundReplyStructABI(t, reflect.TypeOf(calendar.CalendarInboundReplyAuthorization{}), []string{"Matched", "Current", "Target", "Exception", "ExistingReceipt", "DistinctReceiptCount"})
	for _, opaque := range []reflect.Type{
		reflect.TypeOf(calendar.CalendarInboundReplyResult{}),
		reflect.TypeOf(calendar.CalendarInboundReplyScope{}),
		reflect.TypeOf(calendar.CalendarInboundReplyReceiptKey{}),
		reflect.TypeOf(calendar.CalendarInboundReplyPersistenceCommand{}),
		reflect.TypeOf(calendar.CalendarInboundReplyPersistenceResult{}),
	} {
		for field := range opaque.NumField() {
			if opaque.Field(field).PkgPath == "" {
				t.Fatalf("%s exposes field %q", opaque, opaque.Field(field).Name)
			}
		}
	}
	assertInboundReplyExactType(t, reflect.TypeOf((*calendar.EventService).ApplyInboundITIPReply), reflect.TypeOf(func(*calendar.EventService, context.Context, calendar.CalendarInboundReplyInput) (calendar.CalendarInboundReplyResult, error) { panic("oracle") }))
	assertInboundReplyExactType(t, reflect.TypeOf(calendar.NewCalendarInboundReplyReceiptKey), reflect.TypeOf(func(uint8, string, [sha256.Size]byte) (calendar.CalendarInboundReplyReceiptKey, error) { panic("oracle") }))
	assertInboundReplyExactType(t, reflect.TypeOf(calendar.NewCalendarInboundReplyPersistenceCommand), reflect.TypeOf(func(calendar.CalendarInboundReplyAttendeeSnapshot, *calendar.CalendarInboundReplyAttendeeSnapshot, uint, calendar.CalendarInboundReplyReceiptKey, calendar.CalendarInboundReplyReceiptOutcome, *calendar.EventOutboxDescriptor) (calendar.CalendarInboundReplyPersistenceCommand, error) { panic("oracle") }))
	assertInboundReplyExactType(t, reflect.TypeOf(calendar.NewCalendarInboundReplyPersistenceResult), reflect.TypeOf(func(calendar.CalendarInboundReplyAttendeeSnapshot, uint, calendar.CalendarInboundReplyReceiptKey, calendar.CalendarInboundReplyReceiptOutcome, *calendar.EventOutboxDescriptor) (calendar.CalendarInboundReplyPersistenceResult, error) { panic("oracle") }))
	assertInboundReplyValueMethod(t, reflect.TypeOf(calendar.CalendarInboundReplyResult{}), "Outcome", reflect.TypeOf(func(calendar.CalendarInboundReplyResult) calendar.CalendarInboundReplyOutcome { panic("oracle") }))
	assertInboundReplyValueMethod(t, reflect.TypeOf(calendar.CalendarInboundReplyScope{}), "OwnerID", reflect.TypeOf(func(calendar.CalendarInboundReplyScope) uuid.UUID { panic("oracle") }))
	assertInboundReplyValueMethod(t, reflect.TypeOf(calendar.CalendarInboundReplyScope{}), "OrganizationID", reflect.TypeOf(func(calendar.CalendarInboundReplyScope) uuid.UUID { panic("oracle") }))
	assertInboundReplyValueMethod(t, reflect.TypeOf(calendar.CalendarInboundReplyScope{}), "EventOperation", reflect.TypeOf(func(calendar.CalendarInboundReplyScope) calendar.EventMutationOperation { panic("oracle") }))
	assertInboundReplyValueMethod(t, reflect.TypeOf(calendar.CalendarInboundReplyReceiptKey{}), "CanonicalVersion", reflect.TypeOf(func(calendar.CalendarInboundReplyReceiptKey) uint8 { panic("oracle") }))
	assertInboundReplyValueMethod(t, reflect.TypeOf(calendar.CalendarInboundReplyReceiptKey{}), "PayloadHash", reflect.TypeOf(func(calendar.CalendarInboundReplyReceiptKey) [sha256.Size]byte { panic("oracle") }))
	commandType := reflect.TypeOf(calendar.CalendarInboundReplyPersistenceCommand{})
	resultType := reflect.TypeOf(calendar.CalendarInboundReplyPersistenceResult{})
	for _, method := range []struct {
		name string
		want reflect.Type
	}{
		{name: "EventSequence", want: reflect.TypeOf(func(calendar.CalendarInboundReplyPersistenceCommand) uint { panic("oracle") })},
		{name: "Receipt", want: reflect.TypeOf(func(calendar.CalendarInboundReplyPersistenceCommand) calendar.CalendarInboundReplyReceiptKey { panic("oracle") })},
		{name: "ReceiptOutcome", want: reflect.TypeOf(func(calendar.CalendarInboundReplyPersistenceCommand) calendar.CalendarInboundReplyReceiptOutcome { panic("oracle") })},
		{name: "Outbox", want: reflect.TypeOf(func(calendar.CalendarInboundReplyPersistenceCommand) (calendar.EventOutboxDescriptor, bool) { panic("oracle") })},
	} {
		assertInboundReplyValueMethod(t, commandType, method.name, method.want)
	}
	for _, method := range []struct {
		name string
		want reflect.Type
	}{
		{name: "EventSequence", want: reflect.TypeOf(func(calendar.CalendarInboundReplyPersistenceResult) uint { panic("oracle") })},
		{name: "Receipt", want: reflect.TypeOf(func(calendar.CalendarInboundReplyPersistenceResult) calendar.CalendarInboundReplyReceiptKey { panic("oracle") })},
		{name: "ReceiptOutcome", want: reflect.TypeOf(func(calendar.CalendarInboundReplyPersistenceResult) calendar.CalendarInboundReplyReceiptOutcome { panic("oracle") })},
		{name: "Outbox", want: reflect.TypeOf(func(calendar.CalendarInboundReplyPersistenceResult) (calendar.EventOutboxDescriptor, bool) { panic("oracle") })},
	} {
		assertInboundReplyValueMethod(t, resultType, method.name, method.want)
	}
	assertInboundReplyValueMethod(t, reflect.TypeOf(calendar.CalendarInboundReplyPersistenceCommand{}), "Prior", reflect.TypeOf(func(calendar.CalendarInboundReplyPersistenceCommand) calendar.CalendarInboundReplyAttendeeSnapshot { panic("oracle") }))
	assertInboundReplyValueMethod(t, reflect.TypeOf(calendar.CalendarInboundReplyPersistenceCommand{}), "Updated", reflect.TypeOf(func(calendar.CalendarInboundReplyPersistenceCommand) (calendar.CalendarInboundReplyAttendeeSnapshot, bool) { panic("oracle") }))
	assertInboundReplyValueMethod(t, reflect.TypeOf(calendar.CalendarInboundReplyPersistenceCommand{}), "ExpectedVersion", reflect.TypeOf(func(calendar.CalendarInboundReplyPersistenceCommand) uint { panic("oracle") }))
	assertInboundReplyValueMethod(t, reflect.TypeOf(calendar.CalendarInboundReplyPersistenceResult{}), "Attendee", reflect.TypeOf(func(calendar.CalendarInboundReplyPersistenceResult) calendar.CalendarInboundReplyAttendeeSnapshot { panic("oracle") }))
	port := reflect.TypeOf((*calendar.CalendarInboundReplyTransactionBackend)(nil)).Elem()
	if port.NumMethod() != 3 {
		t.Fatalf("inbound transaction port methods = %d", port.NumMethod())
	}
	for _, method := range []string{"ResolveCalendarInboundReplyOwner", "ResolveCalendarInboundReply", "PersistCalendarInboundReply"} {
		if _, found := port.MethodByName(method); !found {
			t.Fatalf("missing inbound transaction port method %q", method)
		}
	}

	if calendar.EventMutationInboundReply != "inbound_reply" ||
		calendar.CalendarInboundReplyApplied != "applied" ||
		calendar.CalendarInboundReplyObservedNoChange != "observed_no_change" ||
		calendar.CalendarInboundReplyDuplicate != "duplicate" ||
		calendar.CalendarInboundReplyStale != "stale" ||
		calendar.CalendarInboundReplyQuarantined != "quarantined" {
		t.Fatalf("inbound reply constants changed")
	}

	parserPayloadHash := [sha256.Size]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	key, err := calendar.NewCalendarInboundReplyReceiptKey(1, "reply@example.test", parserPayloadHash)
	if err != nil {
		t.Fatalf("NewCalendarInboundReplyReceiptKey() error = %v", err)
	}
	want := [sha256.Size]byte{0x25, 0xf0, 0x05, 0x78, 0xe8, 0xcf, 0x6b, 0x75, 0x9d, 0x31, 0xe3, 0x81, 0x3d, 0x57, 0x5f, 0xdd, 0x5f, 0x12, 0xca, 0xa5, 0x14, 0xa9, 0x34, 0x0f, 0x92, 0xfb, 0x94, 0xd6, 0xbc, 0x68, 0x04, 0xf5}
	if key.CanonicalVersion() != 1 || key.PayloadHash() != want {
		t.Fatalf("sender-bound receipt = version:%d hash:%x", key.CanonicalVersion(), key.PayloadHash())
	}

	var nilService *calendar.EventService
	_, err = nilService.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: uuid.New(), Sender: "reply@example.test"})
	if !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("nil service error = %v, want invalid", err)
	}
}

func TestEventServiceApplyInboundITIPReply_RejectsNilTypedNilAndZeroOwnerBeforeWork(t *testing.T) {
	validBackend := newInboundReplyBackend()
	validRepository := &inboundReplyRepositoryFake{backend: validBackend}
	validClock := &inboundReplyClock{}
	validIDs := &inboundReplyIDs{values: []uuid.UUID{inboundReplyEnvelopeID}}
	var typedRepository *inboundReplyRepositoryFake
	var typedClock *inboundReplyClock
	var typedIDs *inboundReplyIDs
	for _, testCase := range []struct {
		name    string
		service *calendar.EventService
		ownerID uuid.UUID
	}{
		{name: "nil repository", service: calendar.NewEventService(nil, validClock, validIDs, inboundReplySerializerUnused{}), ownerID: inboundReplyOwnerID},
		{name: "typed nil repository", service: calendar.NewEventService(typedRepository, validClock, validIDs, inboundReplySerializerUnused{}), ownerID: inboundReplyOwnerID},
		{name: "nil clock", service: calendar.NewEventService(validRepository, nil, validIDs, inboundReplySerializerUnused{}), ownerID: inboundReplyOwnerID},
		{name: "typed nil clock", service: calendar.NewEventService(validRepository, typedClock, validIDs, inboundReplySerializerUnused{}), ownerID: inboundReplyOwnerID},
		{name: "nil ids", service: calendar.NewEventService(validRepository, validClock, nil, inboundReplySerializerUnused{}), ownerID: inboundReplyOwnerID},
		{name: "typed nil ids", service: calendar.NewEventService(validRepository, validClock, typedIDs, inboundReplySerializerUnused{}), ownerID: inboundReplyOwnerID},
		{name: "zero owner", service: calendar.NewEventService(validRepository, validClock, validIDs, inboundReplySerializerUnused{}), ownerID: uuid.Nil},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := testCase.service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: testCase.ownerID, Sender: "reply@example.test", Message: inboundReplyMessage(calendar.RSVPAccepted)})
			if !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("preflight error:%v", err)
			}
		})
	}
	if validBackend.ownerCalls != 0 || validClock.calls != 0 || validIDs.calls != 0 {
		t.Fatalf("invalid dependency path worked owner:%d clock:%d ids:%d", validBackend.ownerCalls, validClock.calls, validIDs.calls)
	}
}

func TestEventServiceApplyInboundITIPReply_AppliesOneOwnerBoundAttendeeCAS(t *testing.T) {
	backend := newInboundReplyBackend()
	clock := &inboundReplyClock{now: time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)}
	ids := &inboundReplyIDs{values: []uuid.UUID{inboundReplyEnvelopeID}}
	service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, clock, ids, inboundReplySerializerUnused{})
	originalCurrent := backend.authorization.Current
	originalOrganizerID := *backend.authorization.Current.OrganizerID
	originalCurrent.OrganizerID = &originalOrganizerID

	result, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{
		OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: inboundReplyMessage(calendar.RSVPAccepted),
	})
	if err != nil || result.Outcome() != calendar.CalendarInboundReplyApplied {
		t.Fatalf("ApplyInboundITIPReply() result:%q error:%v", result.Outcome(), err)
	}
	if backend.ownerCalls != 1 || backend.resolveCalls != 1 || backend.persistCalls != 1 || clock.calls != 1 || ids.calls != 1 || backend.fallbackCalls != 0 {
		t.Fatalf("calls owner:%d resolve:%d persist:%d clock:%d ids:%d fallback:%d", backend.ownerCalls, backend.resolveCalls, backend.persistCalls, clock.calls, ids.calls, backend.fallbackCalls)
	}
	if !reflect.DeepEqual(backend.authorization.Current, originalCurrent) {
		t.Fatalf("inbound reply mutated authoritative event: before:%#v after:%#v", originalCurrent, backend.authorization.Current)
	}
	if len(backend.persisted) != 1 {
		t.Fatalf("persisted commands = %d", len(backend.persisted))
	}
	command := backend.persisted[0]
	updated, found := command.Updated()
	if !found || updated.AttendeeID != inboundReplyAttendeeID || updated.Status != calendar.RSVPAccepted || updated.Version != 3 || updated.RespondedAt == nil || !updated.RespondedAt.Equal(clock.now) || command.ExpectedVersion() != 2 || command.EventSequence() != 4 || command.ReceiptOutcome() != calendar.CalendarInboundReplyReceiptApplied {
		t.Fatalf("inbound persistence command = %#v updated:%#v found:%t", command, updated, found)
	}
	updated.Status = calendar.RSVPDeclined
	updated.RespondedAt = nil
	again, found := command.Updated()
	if !found || again.Status != calendar.RSVPAccepted || again.RespondedAt == nil || !again.RespondedAt.Equal(clock.now) {
		t.Fatalf("updated getter leaked an alias: %#v found:%t", again, found)
	}
	descriptor, found := command.Outbox()
	if !found || descriptor.Envelope().Subject() != "calendar.rsvp.changed.v1" || descriptor.Envelope().EventID() != inboundReplyEnvelopeID {
		t.Fatalf("inbound outbox = %#v found:%t", descriptor, found)
	}
	encoded, marshalErr := json.Marshal(descriptor.Envelope())
	if marshalErr != nil {
		t.Fatalf("json.Marshal(envelope) error = %v", marshalErr)
	}
	const wantEnvelopeJSON = "{\"subject\":\"calendar.rsvp.changed.v1\",\"schema_version\":1,\"event_id\":\"d5555555-5555-4555-8555-555555555555\",\"occurred_at\":\"2035-04-05T06:07:08Z\",\"organization_id\":\"d2222222-2222-4222-8222-222222222222\",\"actor_id\":null,\"correlation_id\":\"d3333333-3333-4333-8333-333333333333\",\"payload\":{\"event_id\":\"d3333333-3333-4333-8333-333333333333\",\"attendee_id\":\"d4444444-4444-4444-8444-444444444444\",\"attendee_email\":\"reply@example.test\",\"previous_status\":\"needs_action\",\"new_status\":\"accepted\",\"version\":3,\"recurrence_id\":null,\"reply_delivery_status\":\"not_required\"}}"
	if string(encoded) != wantEnvelopeJSON {
		t.Fatalf("inbound envelope JSON = %s", encoded)
	}
	var outer struct {
		OrganizationID uuid.UUID `json:"organization_id"`
		ActorID        *uuid.UUID `json:"actor_id"`
		CorrelationID  string     `json:"correlation_id"`
		Payload       struct {
			UID                 string  `json:"uid"`
			RecurrenceID        *string `json:"recurrence_id"`
			ReplyDeliveryStatus string  `json:"reply_delivery_status"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(encoded, &outer); err != nil || outer.OrganizationID != inboundReplyOrganizationID || outer.ActorID != nil || outer.CorrelationID != inboundReplyEventID.String() || outer.Payload.UID != "" || outer.Payload.RecurrenceID != nil || outer.Payload.ReplyDeliveryStatus != "not_required" {
		t.Fatalf("safe inbound envelope outer:%#v error:%v", outer, err)
	}
}

func TestEventServiceApplyInboundITIPReply_MasksOrganizerBeforeVariableClone(t *testing.T) {
	backend := newInboundReplyBackend()
	backend.authorization.Current.OrganizerEmail = "forged-organizer@example.test"
	backend.authorization.Current.Exdates = make([]string, 1001)
	service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, &inboundReplyClock{}, &inboundReplyIDs{}, inboundReplySerializerUnused{})
	result, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{
		OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: inboundReplyMessage(calendar.RSVPAccepted),
	})
	if err != nil || result.Outcome() != calendar.CalendarInboundReplyQuarantined {
		t.Fatalf("forged organizer result:%q error:%v", result.Outcome(), err)
	}
	if backend.ownerCalls != 1 || backend.resolveCalls != 1 || backend.persistCalls != 0 {
		t.Fatalf("forged organizer calls owner:%d resolve:%d persist:%d", backend.ownerCalls, backend.resolveCalls, backend.persistCalls)
	}
}

func TestEventServiceApplyInboundITIPReply_MasksEveryFixedIdentityBeforeCapClone(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*inboundReplyBackendFake)
	}{
		{name: "owner", mutate: func(backend *inboundReplyBackendFake) { backend.authorization.Current.OwnerID = uuid.New() }},
		{name: "organization", mutate: func(backend *inboundReplyBackendFake) { backend.authorization.Current.OrganizationID = uuid.New() }},
		{name: "uid", mutate: func(backend *inboundReplyBackendFake) { backend.authorization.Current.UID = "foreign@example.test" }},
		{name: "organizer id", mutate: func(backend *inboundReplyBackendFake) { organizerID := uuid.New(); backend.authorization.Current.OrganizerID = &organizerID }},
		{name: "organizer", mutate: func(backend *inboundReplyBackendFake) { backend.authorization.Current.OrganizerEmail = "forged@example.test" }},
		{name: "target event", mutate: func(backend *inboundReplyBackendFake) { backend.authorization.Target.EventID = uuid.New() }},
		{name: "target sender", mutate: func(backend *inboundReplyBackendFake) { backend.authorization.Target.Email = "other@example.test" }},
		{name: "null recurrence distinction", mutate: func(backend *inboundReplyBackendFake) { backend.authorization.Target.RecurrenceID = stringPointer("2035-04-06T09:00:00Z") }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newInboundReplyBackend()
			backend.authorization.Current.Exdates = make([]string, 1001)
			testCase.mutate(backend)
			clock := &inboundReplyClock{}
			service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, clock, &inboundReplyIDs{}, inboundReplySerializerUnused{})
			result, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: inboundReplyMessage(calendar.RSVPAccepted)})
			if err != nil || result.Outcome() != calendar.CalendarInboundReplyQuarantined || backend.persistCalls != 0 || clock.calls != 0 {
				t.Fatalf("fixed identity result:%q error:%v persist:%d clock:%d", result.Outcome(), err, backend.persistCalls, clock.calls)
			}
		})
	}
}

func TestCalendarInboundReplyReceiptKey_BindsVersionSenderLengthAndPayload(t *testing.T) {
	payload := [sha256.Size]byte{1}
	base, err := calendar.NewCalendarInboundReplyReceiptKey(1, "reply@example.test", payload)
	if err != nil {
		t.Fatalf("base key error:%v", err)
	}
	changedPayload := payload
	changedPayload[31] = 1
	changedHash, err := calendar.NewCalendarInboundReplyReceiptKey(1, "reply@example.test", changedPayload)
	if err != nil || changedHash.PayloadHash() == base.PayloadHash() {
		t.Fatalf("payload hash variation error:%v base:%x changed:%x", err, base.PayloadHash(), changedHash.PayloadHash())
	}
	changedSender, err := calendar.NewCalendarInboundReplyReceiptKey(1, "other@example.test", payload)
	if err != nil || changedSender.PayloadHash() == base.PayloadHash() {
		t.Fatalf("sender variation error:%v base:%x changed:%x", err, base.PayloadHash(), changedSender.PayloadHash())
	}
	for _, invalid := range []struct {
		version uint8
		sender  string
	}{{version: 0, sender: "reply@example.test"}, {version: 2, sender: "reply@example.test"}, {version: 1, sender: "REPLY@example.test"}, {version: 1, sender: string([]byte{0xff})}, {version: 1, sender: strings.Repeat("a", 321)}} {
		if _, err := calendar.NewCalendarInboundReplyReceiptKey(invalid.version, invalid.sender, payload); !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("invalid key version:%d sender:%q error:%v", invalid.version, invalid.sender, err)
		}
	}
	responded := time.Date(2035, time.April, 2, 3, 4, 5, 0, time.UTC)
	prior := calendar.CalendarInboundReplyAttendeeSnapshot{AttendeeID: inboundReplyAttendeeID, EventID: inboundReplyEventID, Email: "reply@example.test", Status: calendar.RSVPAccepted, RespondedAt: &responded, Version: 2}
	command, err := calendar.NewCalendarInboundReplyPersistenceCommand(prior, nil, 4, base, calendar.CalendarInboundReplyReceiptObservedNoChange, nil)
	if err != nil || command.ExpectedVersion() != 2 || command.Receipt() != base || command.ReceiptOutcome() != calendar.CalendarInboundReplyReceiptObservedNoChange {
		t.Fatalf("observed persistence constructor command:%#v error:%v", command, err)
	}
	prior.RespondedAt = nil
	if returned := command.Prior(); returned.RespondedAt == nil || !returned.RespondedAt.Equal(responded) {
		t.Fatalf("prior input alias escaped: %#v", returned)
	}
	persistenceResult, err := calendar.NewCalendarInboundReplyPersistenceResult(command.Prior(), command.EventSequence(), command.Receipt(), command.ReceiptOutcome(), nil)
	if err != nil {
		t.Fatalf("observed persistence result error:%v", err)
	}
	resultAttendee := persistenceResult.Attendee()
	resultAttendee.RespondedAt = nil
	if again := persistenceResult.Attendee(); again.RespondedAt == nil || !again.RespondedAt.Equal(responded) {
		t.Fatalf("result attendee getter leaked an alias: %#v", again)
	}
	if _, err := calendar.NewCalendarInboundReplyPersistenceCommand(command.Prior(), nil, command.EventSequence(), command.Receipt(), calendar.CalendarInboundReplyReceiptApplied, nil); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("applied without update/outbox error:%v", err)
	}
}

func TestCalendarInboundReplyReceiptKey_UnicodeUsesU32BEUTF8ByteLength(t *testing.T) {
	payload := [sha256.Size]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	for _, vector := range []struct {
		sender       string
		byteLength   int
		runeLength   int
		payloadHash  [sha256.Size]byte
	}{
		{sender: "a@example.test", byteLength: 14, runeLength: 14, payloadHash: [sha256.Size]byte{0x73, 0x66, 0xa2, 0x28, 0x4c, 0x7e, 0x01, 0x92, 0xfd, 0xcf, 0xd8, 0x03, 0x39, 0x24, 0xce, 0x15, 0x89, 0x1f, 0x80, 0xd9, 0x0e, 0xa0, 0xc9, 0xf7, 0xad, 0xd8, 0xf3, 0x2b, 0x08, 0xf7, 0xc5, 0x1f}},
		{sender: "é@example.test", byteLength: 15, runeLength: 14, payloadHash: [sha256.Size]byte{0x3c, 0x08, 0xa2, 0x24, 0xf7, 0xb2, 0x83, 0x30, 0x11, 0xef, 0x8e, 0x29, 0xe0, 0x10, 0x99, 0x82, 0x0d, 0xa0, 0xd9, 0x5c, 0x6b, 0x71, 0x5f, 0x37, 0xf2, 0x20, 0x6b, 0x2e, 0x5c, 0xc6, 0x07, 0x9b}},
		{sender: "müller@example.test", byteLength: 20, runeLength: 19, payloadHash: [sha256.Size]byte{0x08, 0x46, 0xe7, 0xa8, 0xc6, 0x77, 0x16, 0x76, 0xd4, 0x2f, 0xac, 0x09, 0xc7, 0xc3, 0x6a, 0x5d, 0x0a, 0xc1, 0x30, 0x26, 0xd1, 0x59, 0xee, 0x5d, 0xa4, 0x5a, 0x87, 0xaf, 0xf7, 0xf7, 0x6e, 0xc9}},
	} {
		if len(vector.sender) != vector.byteLength || utf8.RuneCountInString(vector.sender) != vector.runeLength {
			t.Fatalf("invalid literal vector sender:%q bytes:%d runes:%d", vector.sender, len(vector.sender), utf8.RuneCountInString(vector.sender))
		}
		key, err := calendar.NewCalendarInboundReplyReceiptKey(1, vector.sender, payload)
		if err != nil || key.PayloadHash() != vector.payloadHash {
			t.Fatalf("U32BE sender bytes vector sender:%q error:%v hash:%x", vector.sender, err, key.PayloadHash())
		}
	}
}

func TestEventServiceApplyInboundITIPReply_NormalizesSenderBeforeSenderBoundHash(t *testing.T) {
	backend := newInboundReplyBackend()
	clock := &inboundReplyClock{now: time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)}
	service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, clock, &inboundReplyIDs{values: []uuid.UUID{inboundReplyEnvelopeID}}, inboundReplySerializerUnused{})
	message := inboundReplyMessage(calendar.RSVPAccepted)
	message.Attendees[0].Email = "REPLY@example.test"
	identity, identityErr := calendar.CanonicalITIPIdentity(message)
	if identityErr != nil {
		t.Fatalf("canonical identity error:%v", identityErr)
	}
	wantReceipt, receiptErr := calendar.NewCalendarInboundReplyReceiptKey(1, "reply@example.test", identity.PayloadHash)
	if receiptErr != nil {
		t.Fatalf("canonical receipt error:%v", receiptErr)
	}
	result, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "REPLY@example.test", Message: message})
	if err != nil || result.Outcome() != calendar.CalendarInboundReplyApplied || len(backend.persisted) != 1 || backend.persisted[0].Receipt() != wantReceipt {
		t.Fatalf("normalized sender result:%q error:%v commands:%d", result.Outcome(), err, len(backend.persisted))
	}
}

func TestCalendarInboundReply_PublicBoundariesDeepCopyInputAuthorizationCommandResultAndOutbox(t *testing.T) {
	respondedAt := time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)
	recurrenceID := "2035-04-06T09:00:00Z"
	prior := calendar.CalendarInboundReplyAttendeeSnapshot{AttendeeID: inboundReplyAttendeeID, EventID: inboundReplyEventID, Email: "reply@example.test", Status: calendar.RSVPNeedsAction, RecurrenceID: &recurrenceID, Version: 2}
	updated := prior
	updated.Status = calendar.RSVPAccepted
	updated.Version = 3
	updated.RespondedAt = &respondedAt
	payload := map[string]any{"items": []string{"before"}}
	envelope, envelopeErr := events.NewEnvelopeV1("calendar.rsvp.changed.v1", inboundReplyEnvelopeID, respondedAt, inboundReplyOrganizationID, nil, inboundReplyEventID.String(), payload)
	if envelopeErr != nil {
		t.Fatalf("deep copy envelope error:%v", envelopeErr)
	}
	descriptor, descriptorErr := calendar.NewEventOutboxDescriptor(envelope)
	if descriptorErr != nil {
		t.Fatalf("deep copy descriptor error:%v", descriptorErr)
	}
	receipt, receiptErr := calendar.NewCalendarInboundReplyReceiptKey(1, "reply@example.test", [sha256.Size]byte{7})
	if receiptErr != nil {
		t.Fatalf("deep copy receipt error:%v", receiptErr)
	}
	command, commandErr := calendar.NewCalendarInboundReplyPersistenceCommand(prior, &updated, 4, receipt, calendar.CalendarInboundReplyReceiptApplied, &descriptor)
	if commandErr != nil {
		t.Fatalf("deep copy command error:%v", commandErr)
	}
	persistenceResult, resultErr := calendar.NewCalendarInboundReplyPersistenceResult(updated, 4, receipt, calendar.CalendarInboundReplyReceiptApplied, &descriptor)
	if resultErr != nil {
		t.Fatalf("deep copy result error:%v", resultErr)
	}
	frozenPayload, marshalErr := json.Marshal(descriptor.Envelope())
	if marshalErr != nil {
		t.Fatalf("deep copy payload marshal error:%v", marshalErr)
	}
	recurrenceID = "2035-04-07T09:00:00Z"
	respondedAt = respondedAt.Add(time.Hour)
	prior.Email = "mutated-prior@example.test"
	updated.Email = "mutated-updated@example.test"
	payload["items"].([]string)[0] = "mutated"
	returnedPrior := command.Prior()
	returnedUpdated, foundUpdated := command.Updated()
	returnedResult := persistenceResult.Attendee()
	if !foundUpdated || returnedPrior.Email != "reply@example.test" || returnedUpdated.Email != "reply@example.test" || returnedUpdated.RecurrenceID == nil || *returnedUpdated.RecurrenceID != "2035-04-06T09:00:00Z" || returnedUpdated.RespondedAt == nil || !returnedUpdated.RespondedAt.Equal(time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)) || returnedResult.Email != "reply@example.test" {
		t.Fatalf("original alias escaped prior:%#v updated:%#v result:%#v", returnedPrior, returnedUpdated, returnedResult)
	}
	*returnedUpdated.RecurrenceID = "returned-mutation"
	*returnedUpdated.RespondedAt = returnedUpdated.RespondedAt.Add(time.Hour)
	returnedResult.Email = "returned-result@example.test"
	returnedPrior.Email = "returned-prior@example.test"
	if againPrior := command.Prior(); againPrior.Email != "reply@example.test" {
		t.Fatalf("returned prior mutation escaped: %#v", againPrior)
	}
	if againUpdated, found := command.Updated(); !found || againUpdated.Email != "reply@example.test" || againUpdated.RecurrenceID == nil || *againUpdated.RecurrenceID != "2035-04-06T09:00:00Z" || againUpdated.RespondedAt == nil || !againUpdated.RespondedAt.Equal(time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)) {
		t.Fatalf("returned updated mutation escaped: %#v found:%t", againUpdated, found)
	}
	if againResult := persistenceResult.Attendee(); againResult.Email != "reply@example.test" || againResult.RecurrenceID == nil || *againResult.RecurrenceID != "2035-04-06T09:00:00Z" || againResult.RespondedAt == nil || !againResult.RespondedAt.Equal(time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)) {
		t.Fatalf("returned result mutation escaped: %#v", againResult)
	}
	commandOutbox, foundCommandOutbox := command.Outbox()
	resultOutbox, foundResultOutbox := persistenceResult.Outbox()
	if !foundCommandOutbox || !foundResultOutbox {
		t.Fatalf("outbox disappeared command:%t result:%t", foundCommandOutbox, foundResultOutbox)
	}
	commandPayload, commandMarshalErr := json.Marshal(commandOutbox.Envelope())
	resultPayload, resultMarshalErr := json.Marshal(resultOutbox.Envelope())
	if commandMarshalErr != nil || resultMarshalErr != nil || string(commandPayload) != string(frozenPayload) || string(resultPayload) != string(frozenPayload) {
		t.Fatalf("descriptor payload alias command:%s result:%s errors:%v/%v", commandPayload, resultPayload, commandMarshalErr, resultMarshalErr)
	}
	commandPayload[0] = '!'
	resultPayload[0] = '!'
	commandPayloadAgain, _ := json.Marshal(commandOutbox.Envelope())
	resultPayloadAgain, _ := json.Marshal(resultOutbox.Envelope())
	if string(commandPayloadAgain) != string(frozenPayload) || string(resultPayloadAgain) != string(frozenPayload) {
		t.Fatal("outbox envelope returned mutable payload bytes")
	}

	backend := newInboundReplyBackend()
	clock := &inboundReplyClock{now: time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)}
	service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, clock, &inboundReplyIDs{values: []uuid.UUID{inboundReplyEnvelopeID}}, inboundReplySerializerUnused{})
	message := inboundReplyMessage(calendar.RSVPAccepted)
	message.EXDates = []calendar.ITIPRecurrenceID{{Time: time.Date(2035, time.April, 7, 9, 0, 0, 0, time.UTC)}}
	result, applyErr := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: message})
	if applyErr != nil || result.Outcome() != calendar.CalendarInboundReplyApplied || len(backend.persisted) != 1 {
		t.Fatalf("input/auth boundary result:%q error:%v commands:%d", result.Outcome(), applyErr, len(backend.persisted))
	}
	persistedCommand := backend.persisted[0]
	persistedUpdated, persistedFound := persistedCommand.Updated()
	persistedOutbox, persistedOutboxFound := persistedCommand.Outbox()
	persistedPayload, persistedMarshalErr := json.Marshal(persistedOutbox.Envelope())
	if !persistedFound || !persistedOutboxFound || persistedMarshalErr != nil {
		t.Fatalf("missing frozen persisted material updated:%t outbox:%t error:%v", persistedFound, persistedOutboxFound, persistedMarshalErr)
	}
	message.Attendees[0].Email = "mutated-input@example.test"
	message.Attendees = append(message.Attendees, calendar.ITIPParticipant{Email: "extra@example.test", Status: calendar.RSVPDeclined})
	message.Temporal.Timed.StartsAt = message.Temporal.Timed.StartsAt.Add(48 * time.Hour)
	message.EXDates[0].Time = message.EXDates[0].Time.Add(48 * time.Hour)
	backend.authorization.Target.Email = "mutated-authorization@example.test"
	backend.authorization.Target.Status = calendar.RSVPDeclined
	backend.authorization.Current.OrganizerEmail = "mutated-organizer@example.test"
	backend.authorization.Current.Sequence = 999
	if againUpdated, found := persistedCommand.Updated(); !found || !reflect.DeepEqual(againUpdated, persistedUpdated) || againUpdated.Email != "reply@example.test" || againUpdated.Status != calendar.RSVPAccepted {
		t.Fatalf("input or authorization alias escaped updated:%#v found:%t", againUpdated, found)
	}
	if againOutbox, found := persistedCommand.Outbox(); !found {
		t.Fatal("input or authorization alias removed outbox")
	} else if encoded, err := json.Marshal(againOutbox.Envelope()); err != nil || string(encoded) != string(persistedPayload) {
		t.Fatalf("input or authorization alias changed envelope:%s error:%v", encoded, err)
	}
}

func TestEventTransactionInboundReply_RequiresCapabilityBeforeOwnerResolveAndClosesScopes(t *testing.T) {
	unsupported := &eventTransactionBackendFake{organizationID: inboundReplyOrganizationID}
	err := calendar.RunEventTransaction(unsupported, func(transaction *calendar.EventTransaction) error {
		_, found, scopeErr := transaction.ResolveCalendarInboundReplyScope(context.Background(), inboundReplyOwnerID)
		if found || !errors.Is(scopeErr, calendar.ErrEventInvalid) {
			t.Fatalf("unsupported inbound scope found:%t error:%v", found, scopeErr)
		}
		return nil
	})
	if err != nil || len(unsupported.scopes) != 0 {
		t.Fatalf("unsupported transaction error:%v backend calls:%d", err, len(unsupported.scopes))
	}

	var typedNil *inboundReplyBackendFake
	err = calendar.RunEventTransaction(typedNil, func(transaction *calendar.EventTransaction) error {
		_, found, scopeErr := transaction.ResolveCalendarInboundReplyScope(context.Background(), inboundReplyOwnerID)
		if found || !errors.Is(scopeErr, calendar.ErrEventInvalid) {
			t.Fatalf("typed nil inbound scope found:%t error:%v", found, scopeErr)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("typed nil transaction error:%v", err)
	}

	backend := newInboundReplyBackend()
	var escaped *calendar.EventTransaction
	err = calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		escaped = transaction
		scope, found, scopeErr := transaction.ResolveCalendarInboundReplyScope(context.Background(), inboundReplyOwnerID)
		if !found || scopeErr != nil || scope.OwnerID() != inboundReplyOwnerID || scope.OrganizationID() != inboundReplyOrganizationID || scope.EventOperation() != calendar.EventMutationInboundReply {
			t.Fatalf("resolved inbound scope=%#v found:%t error:%v", scope, found, scopeErr)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("supported transaction error:%v", err)
	}
	if _, found, scopeErr := escaped.ResolveCalendarInboundReplyScope(context.Background(), inboundReplyOwnerID); found || !errors.Is(scopeErr, calendar.ErrEventTransactionClosed) {
		t.Fatalf("escaped inbound scope found:%t error:%v", found, scopeErr)
	}
	receipt, receiptErr := calendar.NewCalendarInboundReplyReceiptKey(1, "reply@example.test", [sha256.Size]byte{1})
	if receiptErr != nil {
		t.Fatalf("foreign scope receipt error:%v", receiptErr)
	}
	err = calendar.RunEventTransaction(backend, func(ownerTransaction *calendar.EventTransaction) error {
		scope, found, scopeErr := ownerTransaction.ResolveCalendarInboundReplyScope(context.Background(), inboundReplyOwnerID)
		if !found || scopeErr != nil {
			t.Fatalf("foreign source scope found:%t error:%v", found, scopeErr)
		}
		return calendar.RunEventTransaction(backend, func(foreignTransaction *calendar.EventTransaction) error {
			_, resolveErr := foreignTransaction.ResolveCalendarInboundReply(context.Background(), scope, calendar.CalendarInboundReplyAuthorizationRequest{
				UID: "hub-" + inboundReplyEventID.String(), Sender: "reply@example.test", EventSequence: 4, Receipt: receipt,
			})
			if !errors.Is(resolveErr, calendar.ErrEventForbidden) {
				t.Fatalf("foreign scope resolve error:%v", resolveErr)
			}
			return nil
		})
	})
	if err != nil {
		t.Fatalf("foreign scope transaction error:%v", err)
	}
}

func TestEventServiceApplyInboundITIPReply_SequenceCapReceiptAndObservedStateMachine(t *testing.T) {
	tests := []struct {
		name       string
		configure  func(*inboundReplyBackendFake)
		sequence   int
		status     calendar.RSVPStatus
		want       calendar.CalendarInboundReplyOutcome
		wantErr    error
		writes     int
		clockCalls int
	}{
		{name: "existing receipt precedes cancelled current", configure: func(backend *inboundReplyBackendFake) { backend.authorization.ExistingReceipt = true; backend.authorization.Current.Status = calendar.CalendarEventStatusCancelled }, sequence: 1, status: calendar.RSVPAccepted, want: calendar.CalendarInboundReplyDuplicate},
		{name: "lower sequence is stale", sequence: 3, status: calendar.RSVPAccepted, want: calendar.CalendarInboundReplyStale},
		{name: "higher sequence quarantines", sequence: 5, status: calendar.RSVPAccepted, want: calendar.CalendarInboundReplyQuarantined},
		{name: "eighth existing distinct key applies", configure: func(backend *inboundReplyBackendFake) { backend.authorization.DistinctReceiptCount = 7 }, sequence: 4, status: calendar.RSVPAccepted, want: calendar.CalendarInboundReplyApplied, writes: 1, clockCalls: 1},
		{name: "ninth distinct key quarantines", configure: func(backend *inboundReplyBackendFake) { backend.authorization.DistinctReceiptCount = 8 }, sequence: 4, status: calendar.RSVPAccepted, want: calendar.CalendarInboundReplyQuarantined},
		{name: "count above cap is invalid", configure: func(backend *inboundReplyBackendFake) { backend.authorization.DistinctReceiptCount = 9 }, sequence: 4, status: calendar.RSVPAccepted, wantErr: calendar.ErrEventInvalid},
		{name: "same status records observed receipt only", configure: func(backend *inboundReplyBackendFake) { responded := time.Date(2035, time.April, 1, 1, 1, 1, 0, time.UTC); backend.authorization.Target.Status = calendar.RSVPAccepted; backend.authorization.Target.RespondedAt = &responded }, sequence: 4, status: calendar.RSVPAccepted, want: calendar.CalendarInboundReplyObservedNoChange, writes: 1},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newInboundReplyBackend()
			if testCase.configure != nil {
				testCase.configure(backend)
			}
			clock := &inboundReplyClock{now: time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)}
			ids := &inboundReplyIDs{values: []uuid.UUID{inboundReplyEnvelopeID}}
			service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, clock, ids, inboundReplySerializerUnused{})
			message := inboundReplyMessage(testCase.status)
			message.Sequence = testCase.sequence
			result, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: message})
			if !errors.Is(err, testCase.wantErr) || testCase.wantErr == nil && result.Outcome() != testCase.want {
				t.Fatalf("state result:%q error:%v want:%q wantErr:%v", result.Outcome(), err, testCase.want, testCase.wantErr)
			}
			if backend.persistCalls != testCase.writes || clock.calls != testCase.clockCalls || backend.fallbackCalls != 0 {
				t.Fatalf("state calls persist:%d clock:%d fallback:%d", backend.persistCalls, clock.calls, backend.fallbackCalls)
			}
			if testCase.want == calendar.CalendarInboundReplyObservedNoChange && len(backend.persisted) == 1 {
				if _, found := backend.persisted[0].Updated(); found {
					t.Fatal("observed receipt carried an attendee update")
				}
				if _, found := backend.persisted[0].Outbox(); found {
					t.Fatal("observed receipt carried an outbox descriptor")
				}
			}
		})
	}
}

func TestEventServiceApplyInboundITIPReply_RejectsEveryMalformedDistinctReceiptCountAboveNine(t *testing.T) {
	for count := 10; count <= 255; count++ {
		backend := newInboundReplyBackend()
		backend.authorization.DistinctReceiptCount = uint8(count)
		clock := &inboundReplyClock{now: time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)}
		ids := &inboundReplyIDs{values: []uuid.UUID{inboundReplyEnvelopeID}}
		service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, clock, ids, inboundReplySerializerUnused{})
		_, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: inboundReplyMessage(calendar.RSVPAccepted)})
		if !errors.Is(err, calendar.ErrEventInvalid) || backend.persistCalls != 0 || clock.calls != 0 || ids.calls != 0 {
			t.Fatalf("malformed count:%d error:%v persist:%d clock:%d ids:%d", count, err, backend.persistCalls, clock.calls, ids.calls)
		}
	}
}

func TestEventServiceApplyInboundITIPReply_DeclinedAcceptedRedeliveryStaysDuplicateForSeriesAndOccurrence(t *testing.T) {
	const recurrence = "2035-04-06T09:00:00Z"
	for _, testCase := range []struct {
		name    string
		prepare func(*inboundReplyBackendFake, *calendar.ITIPMessage)
	}{
		{name: "series", prepare: func(_ *inboundReplyBackendFake, _ *calendar.ITIPMessage) {}},
		{name: "occurrence", prepare: func(backend *inboundReplyBackendFake, message *calendar.ITIPMessage) {
			backend.authorization.Current.RRULE = "FREQ=DAILY"
			backend.authorization.Target.RecurrenceID = stringPointer(recurrence)
			message.RecurrenceID = &calendar.ITIPRecurrenceID{Time: time.Date(2035, time.April, 6, 9, 0, 0, 0, time.UTC)}
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newInboundReplyBackend()
			clock := &inboundReplyClock{now: time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)}
			ids := &inboundReplyIDs{values: []uuid.UUID{inboundReplyEnvelopeID, uuid.MustParse("d6666666-6666-4666-8666-666666666666")}}
			service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, clock, ids, inboundReplySerializerUnused{})
			declined := inboundReplyMessage(calendar.RSVPDeclined)
			testCase.prepare(backend, &declined)
			result, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: declined})
			if err != nil || result.Outcome() != calendar.CalendarInboundReplyApplied {
				t.Fatalf("A result:%q error:%v", result.Outcome(), err)
			}
			respondedA := clock.now
			backend.authorization.Target.Status = calendar.RSVPDeclined
			backend.authorization.Target.Version = 3
			backend.authorization.Target.RespondedAt = &respondedA
			accepted := inboundReplyMessage(calendar.RSVPAccepted)
			testCase.prepare(backend, &accepted)
			result, err = service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: accepted})
			if err != nil || result.Outcome() != calendar.CalendarInboundReplyApplied || len(backend.persisted) != 2 || backend.persisted[0].Receipt() == backend.persisted[1].Receipt() {
				t.Fatalf("B result:%q error:%v commands:%d", result.Outcome(), err, len(backend.persisted))
			}
			respondedB := clock.now
			backend.authorization.Target.Status = calendar.RSVPAccepted
			backend.authorization.Target.Version = 4
			backend.authorization.Target.RespondedAt = &respondedB
			backend.authorization.ExistingReceipt = true
			result, err = service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: declined})
			if err != nil || result.Outcome() != calendar.CalendarInboundReplyDuplicate || len(backend.persisted) != 2 || clock.calls != 2 || ids.calls != 2 || backend.fallbackCalls != 0 {
				t.Fatalf("redelivered A result:%q error:%v commands:%d clock:%d ids:%d fallback:%d", result.Outcome(), err, len(backend.persisted), clock.calls, ids.calls, backend.fallbackCalls)
			}
		})
	}
}

func TestEventServiceApplyInboundITIPReply_OccurrenceRequiresMaterializedActiveMember(t *testing.T) {
	const recurrence = "2035-04-06T09:00:00Z"
	configureOccurrence := func(backend *inboundReplyBackendFake) calendar.ITIPMessage {
		backend.authorization.Current.RRULE = "FREQ=DAILY"
		backend.authorization.Target.RecurrenceID = stringPointer(recurrence)
		message := inboundReplyMessage(calendar.RSVPAccepted)
		message.RecurrenceID = &calendar.ITIPRecurrenceID{Time: time.Date(2035, time.April, 6, 9, 0, 0, 0, time.UTC)}
		return message
	}
	for _, testCase := range []struct {
		name         string
		mutate       func(*inboundReplyBackendFake, *calendar.ITIPMessage)
		want         calendar.CalendarInboundReplyOutcome
		wantWrite    bool
		wantSequence uint
	}{
		{name: "materialized active occurrence applies", want: calendar.CalendarInboundReplyApplied, wantWrite: true, wantSequence: 4},
		{name: "missing occurrence quarantines", mutate: func(backend *inboundReplyBackendFake, _ *calendar.ITIPMessage) { backend.authorization.Target.RecurrenceID = nil }, want: calendar.CalendarInboundReplyQuarantined},
		{name: "nonmember quarantines", mutate: func(backend *inboundReplyBackendFake, message *calendar.ITIPMessage) { nonmember := "2035-04-07T09:00:00Z"; backend.authorization.Current.RRULE = "FREQ=DAILY;COUNT=1"; backend.authorization.Target.RecurrenceID = &nonmember; message.RecurrenceID = &calendar.ITIPRecurrenceID{Time: time.Date(2035, time.April, 7, 9, 0, 0, 0, time.UTC)} }, want: calendar.CalendarInboundReplyQuarantined},
		{name: "exdate quarantines", mutate: func(backend *inboundReplyBackendFake, _ *calendar.ITIPMessage) { backend.authorization.Current.Exdates = []string{recurrence} }, want: calendar.CalendarInboundReplyQuarantined},
		{name: "cancelled exception quarantines", mutate: func(backend *inboundReplyBackendFake, _ *calendar.ITIPMessage) { backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: uuid.New(), EventID: inboundReplyEventID, RecurrenceID: recurrence, Sequence: 4, Status: calendar.CalendarEventStatusCancelled, Cancelled: true} }, want: calendar.CalendarInboundReplyQuarantined},
		{name: "active exception supplies effective sequence", mutate: func(backend *inboundReplyBackendFake, message *calendar.ITIPMessage) { backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: uuid.New(), EventID: inboundReplyEventID, RecurrenceID: recurrence, Sequence: 3, Status: calendar.CalendarEventStatusConfirmed}; message.Sequence = 3 }, want: calendar.CalendarInboundReplyApplied, wantWrite: true, wantSequence: 3},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newInboundReplyBackend()
			message := configureOccurrence(backend)
			if testCase.mutate != nil {
				testCase.mutate(backend, &message)
			}
			clock := &inboundReplyClock{now: time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)}
			service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, clock, &inboundReplyIDs{values: []uuid.UUID{inboundReplyEnvelopeID}}, inboundReplySerializerUnused{})
			result, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: message})
			if err != nil || result.Outcome() != testCase.want || testCase.wantWrite != (backend.persistCalls == 1) {
				t.Fatalf("occurrence result:%q error:%v writes:%d", result.Outcome(), err, backend.persistCalls)
			}
			if testCase.wantWrite {
				if backend.persisted[0].EventSequence() != testCase.wantSequence {
					t.Fatalf("occurrence effective sequence:%d want:%d", backend.persisted[0].EventSequence(), testCase.wantSequence)
				}
				descriptor, found := backend.persisted[0].Outbox()
				encoded, marshalErr := json.Marshal(descriptor.Envelope())
				if !found || marshalErr != nil || !strings.Contains(string(encoded), `"recurrence_id":"`+recurrence+`"`) {
					t.Fatalf("occurrence envelope found:%t error:%v json:%s", found, marshalErr, encoded)
				}
			}
		})
	}
}

func TestEventServiceApplyInboundITIPReply_RejectsEchoTamperIDCollisionAndRawErrors(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*inboundReplyBackendFake, *inboundReplyIDs)
	}{
		{name: "application event id collides with event", configure: func(_ *inboundReplyBackendFake, ids *inboundReplyIDs) { ids.values = []uuid.UUID{inboundReplyEventID} }},
		{name: "persistence echo changes sequence", configure: func(backend *inboundReplyBackendFake, _ *inboundReplyIDs) { backend.resultOverride = func(command calendar.CalendarInboundReplyPersistenceCommand) (calendar.CalendarInboundReplyPersistenceResult, error) { attendee, _ := command.Updated(); outbox, _ := command.Outbox(); return calendar.NewCalendarInboundReplyPersistenceResult(attendee, command.EventSequence()+1, command.Receipt(), command.ReceiptOutcome(), &outbox) } }},
		{name: "raw backend error is contained", configure: func(backend *inboundReplyBackendFake, _ *inboundReplyIDs) { backend.persistError = inboundReplyHostileError{} }},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newInboundReplyBackend()
			ids := &inboundReplyIDs{values: []uuid.UUID{inboundReplyEnvelopeID}}
			testCase.configure(backend, ids)
			service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, &inboundReplyClock{now: time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)}, ids, inboundReplySerializerUnused{})
			_, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: inboundReplyMessage(calendar.RSVPAccepted)})
			if !errors.Is(err, calendar.ErrEventInvalid) || strings.Contains(err.Error(), "raw-inbound-secret") || strings.Contains(err.Error(), "hub-") || len(backend.persisted) != 0 {
				t.Fatalf("unsafe failure error:%v persisted:%d", err, len(backend.persisted))
		}
		})
	}
}

func TestEventServiceApplyInboundITIPReply_RollsBackEveryPersistenceEchoMismatch(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		override func(calendar.CalendarInboundReplyPersistenceCommand) (calendar.CalendarInboundReplyPersistenceResult, error)
	}{
		{name: "attendee", override: func(command calendar.CalendarInboundReplyPersistenceCommand) (calendar.CalendarInboundReplyPersistenceResult, error) {
			attendee, _ := command.Updated()
			attendee.Email = "other@example.test"
			outbox, _ := command.Outbox()
			return calendar.NewCalendarInboundReplyPersistenceResult(attendee, command.EventSequence(), command.Receipt(), command.ReceiptOutcome(), &outbox)
		}},
		{name: "sequence", override: func(command calendar.CalendarInboundReplyPersistenceCommand) (calendar.CalendarInboundReplyPersistenceResult, error) {
			attendee, _ := command.Updated()
			outbox, _ := command.Outbox()
			return calendar.NewCalendarInboundReplyPersistenceResult(attendee, command.EventSequence()+1, command.Receipt(), command.ReceiptOutcome(), &outbox)
		}},
		{name: "receipt", override: func(command calendar.CalendarInboundReplyPersistenceCommand) (calendar.CalendarInboundReplyPersistenceResult, error) {
			attendee, _ := command.Updated()
			outbox, _ := command.Outbox()
			tamperedReceipt, err := calendar.NewCalendarInboundReplyReceiptKey(1, "other@example.test", [sha256.Size]byte{9})
			if err != nil {
				return calendar.CalendarInboundReplyPersistenceResult{}, err
			}
			return calendar.NewCalendarInboundReplyPersistenceResult(attendee, command.EventSequence(), tamperedReceipt, command.ReceiptOutcome(), &outbox)
		}},
		{name: "outcome", override: func(command calendar.CalendarInboundReplyPersistenceCommand) (calendar.CalendarInboundReplyPersistenceResult, error) {
			return calendar.NewCalendarInboundReplyPersistenceResult(command.Prior(), command.EventSequence(), command.Receipt(), calendar.CalendarInboundReplyReceiptObservedNoChange, nil)
		}},
		{name: "outbox", override: func(command calendar.CalendarInboundReplyPersistenceCommand) (calendar.CalendarInboundReplyPersistenceResult, error) {
			attendee, _ := command.Updated()
			envelope, err := events.NewEnvelopeV1("calendar.rsvp.changed.v1", uuid.MustParse("d7777777-7777-4777-8777-777777777777"), time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC), inboundReplyOrganizationID, nil, inboundReplyEventID.String(), struct{ Tampered bool }{Tampered: true})
			if err != nil {
				return calendar.CalendarInboundReplyPersistenceResult{}, err
			}
			outbox, err := calendar.NewEventOutboxDescriptor(envelope)
			if err != nil {
				return calendar.CalendarInboundReplyPersistenceResult{}, err
			}
			return calendar.NewCalendarInboundReplyPersistenceResult(attendee, command.EventSequence(), command.Receipt(), command.ReceiptOutcome(), &outbox)
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newInboundReplyBackend()
			backend.resultOverride = testCase.override
			service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, &inboundReplyClock{now: time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)}, &inboundReplyIDs{values: []uuid.UUID{inboundReplyEnvelopeID}}, inboundReplySerializerUnused{})
			_, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: inboundReplyMessage(calendar.RSVPAccepted)})
			if !errors.Is(err, calendar.ErrEventInvalid) || len(backend.persisted) != 0 {
				t.Fatalf("persistence echo mismatch error:%v persisted:%d", err, len(backend.persisted))
			}
		})
	}
}

func TestEventServiceApplyInboundITIPReply_RejectsVersionOverflowAndOwnerMissWithoutWrites(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		configure func(*inboundReplyBackendFake)
		want      calendar.CalendarInboundReplyOutcome
		wantErr   error
	}{
		{name: "owner miss quarantines", configure: func(backend *inboundReplyBackendFake) { backend.ownerFound = false; backend.ownerOrganization = uuid.Nil }, want: calendar.CalendarInboundReplyQuarantined},
		{name: "owner miss malformed organization is invalid", configure: func(backend *inboundReplyBackendFake) { backend.ownerFound = false; backend.ownerOrganization = inboundReplyOrganizationID }, wantErr: calendar.ErrEventInvalid},
		{name: "target version overflow is invalid", configure: func(backend *inboundReplyBackendFake) { backend.authorization.Target.Version = ^uint(0) }, wantErr: calendar.ErrEventInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newInboundReplyBackend()
			testCase.configure(backend)
			clock := &inboundReplyClock{now: time.Date(2035, time.April, 5, 6, 7, 8, 0, time.UTC)}
			service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, clock, &inboundReplyIDs{values: []uuid.UUID{inboundReplyEnvelopeID}}, inboundReplySerializerUnused{})
			result, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: inboundReplyMessage(calendar.RSVPAccepted)})
			if !errors.Is(err, testCase.wantErr) || testCase.wantErr == nil && result.Outcome() != testCase.want || backend.persistCalls != 0 || clock.calls != 0 {
				t.Fatalf("overflow/owner result:%q error:%v writes:%d clock:%d", result.Outcome(), err, backend.persistCalls, clock.calls)
			}
		})
	}
}

func TestEventServiceApplyInboundITIPReply_QuarantinesHostileIngressBeforeTransaction(t *testing.T) {
	backend := newInboundReplyBackend()
	service := calendar.NewEventService(&inboundReplyRepositoryFake{backend: backend}, &inboundReplyClock{}, &inboundReplyIDs{}, inboundReplySerializerUnused{})
	for _, testCase := range []struct {
		name   string
		mutate func(*calendar.ITIPMessage)
	}{
		{name: "request method", mutate: func(message *calendar.ITIPMessage) { message.Method = calendar.ITIPMethodRequest }},
		{name: "no attendee", mutate: func(message *calendar.ITIPMessage) { message.Attendees = nil }},
		{name: "two attendees", mutate: func(message *calendar.ITIPMessage) { message.Attendees = append(message.Attendees, message.Attendees[0]) }},
		{name: "needs action", mutate: func(message *calendar.ITIPMessage) { message.Attendees[0].Status = calendar.RSVPNeedsAction }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			message := inboundReplyMessage(calendar.RSVPAccepted)
			testCase.mutate(&message)
			result, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: message})
			if err != nil || result.Outcome() != calendar.CalendarInboundReplyQuarantined || backend.ownerCalls != 0 || backend.resolveCalls != 0 || backend.persistCalls != 0 {
				t.Fatalf("canonical ingress result:%q error:%v calls owner:%d resolve:%d persist:%d", result.Outcome(), err, backend.ownerCalls, backend.resolveCalls, backend.persistCalls)
			}
		})
	}
	message := inboundReplyMessage(calendar.RSVPAccepted)
	message.Attendees[0].Email = "different@example.test"
	result, err := service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: "reply@example.test", Message: message})
	if err != nil || result.Outcome() != calendar.CalendarInboundReplyQuarantined || backend.ownerCalls != 0 || backend.resolveCalls != 0 || backend.persistCalls != 0 {
		t.Fatalf("mismatched ingress result:%q error:%v calls owner:%d resolve:%d persist:%d", result.Outcome(), err, backend.ownerCalls, backend.resolveCalls, backend.persistCalls)
	}
	result, err = service.ApplyInboundITIPReply(context.Background(), calendar.CalendarInboundReplyInput{OwnerID: inboundReplyOwnerID, Sender: string(make([]byte, 1024*1024)), Message: inboundReplyMessage(calendar.RSVPAccepted)})
	if err != nil || result.Outcome() != calendar.CalendarInboundReplyQuarantined || backend.ownerCalls != 0 {
		t.Fatalf("oversized ingress result:%q error:%v owner:%d", result.Outcome(), err, backend.ownerCalls)
	}
}

func assertInboundReplyStructABI(t *testing.T, typ reflect.Type, want []string) {
	t.Helper()
	if typ.NumField() != len(want) {
		t.Fatalf("%s field count = %d, want %d", typ, typ.NumField(), len(want))
	}
	for index, name := range want {
		field := typ.Field(index)
		if field.Name != name || field.PkgPath != "" {
			t.Fatalf("%s field %d = %#v, want exported %q", typ, index, field, name)
		}
	}
}

func assertInboundReplyExactType(t *testing.T, got, want reflect.Type) {
	t.Helper()
	if got != want {
		t.Fatalf("signature = %s, want %s", got, want)
	}
}

func assertInboundReplyValueMethod(t *testing.T, typ reflect.Type, name string, want reflect.Type) {
	t.Helper()
	method, found := typ.MethodByName(name)
	if !found || method.Type != want {
		t.Fatalf("%s.%s signature = %v, want %s", typ, name, method.Type, want)
	}
}

func inboundReplyMessage(status calendar.RSVPStatus) calendar.ITIPMessage {
	return calendar.ITIPMessage{
		Method: calendar.ITIPMethodReply, UID: "hub-" + inboundReplyEventID.String(), Stamp: time.Date(2035, time.April, 4, 6, 7, 8, 0, time.UTC), Sequence: 4,
		Status: calendar.ITIPStatusConfirmed, Summary: "Inbound RSVP", Description: "", Location: "",
		Temporal: calendar.ITIPTemporal{Timed: &calendar.ITIPTimed{StartsAt: time.Date(2035, time.April, 6, 9, 0, 0, 0, time.UTC), EndsAt: time.Date(2035, time.April, 6, 10, 0, 0, 0, time.UTC), Timezone: "UTC"}},
		Organizer: calendar.ITIPOrganizer{Email: "owner@example.test"},
		Attendees: []calendar.ITIPParticipant{{Email: "reply@example.test", Role: calendar.ITIPParticipantRequired, Status: status, RSVP: true}},
	}
}

type inboundReplyHostileError struct{}

func (inboundReplyHostileError) Error() string { return "raw-inbound-secret" }
func (inboundReplyHostileError) Unwrap() error { panic("hostile unwrap must not be evaluated") }

type inboundReplyRepositoryFake struct{ backend *inboundReplyBackendFake }

func (repository *inboundReplyRepositoryFake) WithEventTransaction(_ context.Context, callback func(*calendar.EventTransaction) error) error {
	persisted := append([]calendar.CalendarInboundReplyPersistenceCommand(nil), repository.backend.persisted...)
	err := calendar.RunEventTransaction(repository.backend, callback)
	if err != nil {
		repository.backend.persisted = persisted
	}
	return err
}

type inboundReplyBackendFake struct {
	ownerCalls        int
	resolveCalls      int
	persistCalls      int
	fallbackCalls     int
	ownerFound        bool
	ownerOrganization uuid.UUID
	ownerError        error
	resolveError      error
	persistError      error
	resultOverride    func(calendar.CalendarInboundReplyPersistenceCommand) (calendar.CalendarInboundReplyPersistenceResult, error)
	authorization calendar.CalendarInboundReplyAuthorization
	persisted    []calendar.CalendarInboundReplyPersistenceCommand
}

func newInboundReplyBackend() *inboundReplyBackendFake {
	return &inboundReplyBackendFake{ownerFound: true, ownerOrganization: inboundReplyOrganizationID, authorization: calendar.CalendarInboundReplyAuthorization{
		Matched: true,
		Current: calendar.CalendarEventCommandInput{
			EventID: inboundReplyEventID, OrganizationID: inboundReplyOrganizationID, OwnerID: inboundReplyOwnerID, UID: "hub-" + inboundReplyEventID.String(),
			Title: "Inbound reply", Temporal: calendar.CalendarEvent{StartsAt: time.Date(2035, time.April, 6, 9, 0, 0, 0, time.UTC), EndsAt: time.Date(2035, time.April, 6, 10, 0, 0, 0, time.UTC), Timezone: "UTC"},
			Sequence: 4, Version: 2, Source: calendar.CalendarEventSourceHub, Status: calendar.CalendarEventStatusConfirmed, OrganizerID: &inboundReplyOwnerID, OrganizerEmail: "owner@example.test",
		},
		Target: calendar.CalendarInboundReplyAttendeeSnapshot{AttendeeID: inboundReplyAttendeeID, EventID: inboundReplyEventID, Email: "reply@example.test", Status: calendar.RSVPNeedsAction, Version: 2},
	}}
}

func (backend *inboundReplyBackendFake) ResolveCalendarInboundReplyOwner(context.Context, uuid.UUID) (uuid.UUID, bool, error) {
	backend.ownerCalls++
	return backend.ownerOrganization, backend.ownerFound, backend.ownerError
}

func (backend *inboundReplyBackendFake) ResolveCalendarInboundReply(_ context.Context, _ calendar.CalendarInboundReplyScope, request calendar.CalendarInboundReplyAuthorizationRequest) (calendar.CalendarInboundReplyAuthorization, error) {
	backend.resolveCalls++
	if backend.resolveError != nil {
		return calendar.CalendarInboundReplyAuthorization{}, backend.resolveError
	}
	if request.UID != "hub-"+inboundReplyEventID.String() || request.Sender != "reply@example.test" {
		return calendar.CalendarInboundReplyAuthorization{}, calendar.ErrEventInvalid
	}
	return backend.authorization, nil
}

func (backend *inboundReplyBackendFake) PersistCalendarInboundReply(_ context.Context, _ calendar.CalendarInboundReplyScope, command calendar.CalendarInboundReplyPersistenceCommand) (calendar.CalendarInboundReplyPersistenceResult, error) {
	backend.persistCalls++
	if backend.persistError != nil {
		return calendar.CalendarInboundReplyPersistenceResult{}, backend.persistError
	}
	backend.persisted = append(backend.persisted, command)
	if backend.resultOverride != nil {
		return backend.resultOverride(command)
	}
	attendee, updated := command.Updated()
	if !updated {
		attendee = command.Prior()
	}
	outbox, found := command.Outbox()
	if !found {
		return calendar.NewCalendarInboundReplyPersistenceResult(attendee, command.EventSequence(), command.Receipt(), command.ReceiptOutcome(), nil)
	}
	return calendar.NewCalendarInboundReplyPersistenceResult(attendee, command.EventSequence(), command.Receipt(), command.ReceiptOutcome(), &outbox)
}

func (backend *inboundReplyBackendFake) ResolveActorOrganization(context.Context, uuid.UUID) (uuid.UUID, error) { backend.fallbackCalls++; return uuid.Nil, calendar.ErrEventInvalid }
func (backend *inboundReplyBackendFake) ResolveCalendarEventCreateAuthorization(context.Context, calendar.EventMutationScope, calendar.CalendarEventCreateAuthorizationRequest) (calendar.CalendarEventCreateAuthorization, error) { backend.fallbackCalls++; return calendar.CalendarEventCreateAuthorization{}, calendar.ErrEventInvalid }
func (backend *inboundReplyBackendFake) ReserveMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptReservation) (calendar.MutationReceiptReservationResult, error) { backend.fallbackCalls++; return calendar.MutationReceiptReservationResult{}, calendar.ErrEventInvalid }
func (backend *inboundReplyBackendFake) CompleteMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptCompletion) error { backend.fallbackCalls++; return calendar.ErrEventInvalid }
func (backend *inboundReplyBackendFake) PersistCalendarEvent(context.Context, calendar.EventMutationScope, calendar.CalendarEventCommand) (calendar.CalendarEventPersistenceResult, error) { backend.fallbackCalls++; return calendar.CalendarEventPersistenceResult{}, calendar.ErrEventInvalid }
func (backend *inboundReplyBackendFake) PersistCalendarEventAttendee(context.Context, calendar.EventMutationScope, calendar.CalendarEventAttendeeCommand) (calendar.CalendarEventAttendeePersistenceResult, error) { backend.fallbackCalls++; return calendar.CalendarEventAttendeePersistenceResult{}, calendar.ErrEventInvalid }
func (backend *inboundReplyBackendFake) PersistCalendarEventException(context.Context, calendar.EventMutationScope, calendar.CalendarEventExceptionCommand) (calendar.CalendarEventExceptionPersistenceResult, error) { backend.fallbackCalls++; return calendar.CalendarEventExceptionPersistenceResult{}, calendar.ErrEventInvalid }
func (backend *inboundReplyBackendFake) AppendOutbox(context.Context, calendar.EventMutationScope, calendar.EventOutboxDescriptor) error { backend.fallbackCalls++; return calendar.ErrEventInvalid }

type inboundReplyClock struct {
	now   time.Time
	calls int
}

func (clock *inboundReplyClock) Now() time.Time {
	clock.calls++
	return clock.now
}

type inboundReplyIDs struct {
	values []uuid.UUID
	calls  int
}

func (ids *inboundReplyIDs) NewEventID() uuid.UUID {
	if ids.calls >= len(ids.values) {
		return uuid.Nil
	}
	value := ids.values[ids.calls]
	ids.calls++
	return value
}

type inboundReplySerializerUnused struct{}

func (inboundReplySerializerUnused) Serialize(calendar.ITIPMessage) ([]byte, error) {
	panic("inbound reply must not serialize outbound iTIP")
}
