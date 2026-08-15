package mail_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/platform/events"
)

var (
	eventID        = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	organizationID = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	actorID        = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	accountID      = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	mailboxID      = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	messageID      = uuid.MustParse("66666666-6666-4666-8666-666666666666")
	threadID       = uuid.MustParse("77777777-7777-4777-8777-777777777777")
	submissionID   = uuid.MustParse("88888888-8888-4888-8888-888888888888")
	folderID       = uuid.MustParse("99999999-9999-4999-8999-999999999999")
	labelID        = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	occurredAt     = time.Date(2030, time.May, 6, 7, 8, 9, 123456789, time.UTC)
	payloadTime    = time.Date(2030, time.May, 6, 7, 8, 9, 0, time.UTC)
)

func TestMailEventMappers_SerializeExactCompleteEnvelopes(t *testing.T) {
	calendarMethod := mail.MailCalendarMethodRequest
	for _, testCase := range []struct {
		name string
		mapIt func() ([]byte, error)
		want  string
	}{
		{
			name: "received",
			mapIt: func() ([]byte, error) {
				envelope, err := mail.NewMailReceivedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", receivedPayload(&calendarMethod))
				if err != nil {
					return nil, err
				}
				return json.Marshal(envelope)
			},
			want: `{"subject":"mail.received.v1","schema_version":1,"event_id":"11111111-1111-4111-8111-111111111111","occurred_at":"2030-05-06T07:08:09.123456789Z","organization_id":"22222222-2222-4222-8222-222222222222","actor_id":"33333333-3333-4333-8333-333333333333","correlation_id":"corr-mail-03d","payload":{"mailbox_id":"55555555-5555-4555-8555-555555555555","message_id":"66666666-6666-4666-8666-666666666666","engine_message_id":"engine-message-03d","thread_id":"77777777-7777-4777-8777-777777777777","from":{"email":"sender@example.test","name":"Sender"},"to":[{"email":"recipient@example.test"}],"received_at":"2030-05-06T07:08:09Z","folder_ids":["99999999-9999-4999-8999-999999999999"],"has_attachments":true,"calendar_method":"request"}}`,
		},
		{
			name: "sent",
			mapIt: func() ([]byte, error) {
				envelope, err := mail.NewMailSentEnvelopeV1(eventID, occurredAt, organizationID, nil, "corr-mail-03d", sentPayload())
				if err != nil {
					return nil, err
				}
				return json.Marshal(envelope)
			},
			want: `{"subject":"mail.sent.v1","schema_version":1,"event_id":"11111111-1111-4111-8111-111111111111","occurred_at":"2030-05-06T07:08:09.123456789Z","organization_id":"22222222-2222-4222-8222-222222222222","actor_id":null,"correlation_id":"corr-mail-03d","payload":{"mailbox_id":"55555555-5555-4555-8555-555555555555","submission_id":"88888888-8888-4888-8888-888888888888","message_id":"66666666-6666-4666-8666-666666666666","thread_id":"77777777-7777-4777-8777-777777777777","from":{"email":"sender@example.test","name":"Sender"},"to":[{"email":"recipient@example.test"}],"sent_at":"2030-05-06T07:08:09Z","delivery_class":"external"}}`,
		},
		{
			name: "message changed",
			mapIt: func() ([]byte, error) {
				envelope, err := mail.NewMailMessageChangedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", changedPayload())
				if err != nil {
					return nil, err
				}
				return json.Marshal(envelope)
			},
			want: `{"subject":"mail.message.changed.v1","schema_version":1,"event_id":"11111111-1111-4111-8111-111111111111","occurred_at":"2030-05-06T07:08:09.123456789Z","organization_id":"22222222-2222-4222-8222-222222222222","actor_id":"33333333-3333-4333-8333-333333333333","correlation_id":"corr-mail-03d","payload":{"mailbox_id":"55555555-5555-4555-8555-555555555555","message_id":"66666666-6666-4666-8666-666666666666","changed_fields":["folder_ids","seen","labels"],"folder_ids":["99999999-9999-4999-8999-999999999999"],"flags":["seen","flagged"],"label_ids":["aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"]}}`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := testCase.mapIt()
			if err != nil {
				t.Fatalf("map event error = %v", err)
			}
			if !bytes.Equal(got, []byte(testCase.want)) {
				t.Fatalf("envelope bytes = %s, want %s", got, testCase.want)
			}
			assertMailEnvelopeSchemaShape(t, got)
			assertForbiddenContentAbsent(t, got)
		})
	}
}

func TestMailEventMappers_AreDeterministicAndEventIDIsTheOnlyVaryingField(t *testing.T) {
	calendarMethod := mail.MailCalendarMethodRequest
	first, err := mail.NewMailReceivedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", receivedPayload(&calendarMethod))
	if err != nil {
		t.Fatalf("first mapper error = %v", err)
	}
	second, err := mail.NewMailReceivedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", receivedPayload(&calendarMethod))
	if err != nil {
		t.Fatalf("second mapper error = %v", err)
	}
	firstBytes, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first = %v", err)
	}
	secondBytes, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("marshal second = %v", err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("same application event id produced different bytes: %s != %s", firstBytes, secondBytes)
	}

	otherEventID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	other, err := mail.NewMailReceivedEnvelopeV1(otherEventID, occurredAt, organizationID, &actorID, "corr-mail-03d", receivedPayload(&calendarMethod))
	if err != nil {
		t.Fatalf("different event id mapper error = %v", err)
	}
	otherBytes, err := json.Marshal(other)
	if err != nil {
		t.Fatalf("marshal other = %v", err)
	}
	var firstValue, otherValue map[string]any
	if err := json.Unmarshal(firstBytes, &firstValue); err != nil {
		t.Fatalf("unmarshal first = %v", err)
	}
	if err := json.Unmarshal(otherBytes, &otherValue); err != nil {
		t.Fatalf("unmarshal other = %v", err)
	}
	delete(firstValue, "event_id")
	delete(otherValue, "event_id")
	if !reflect.DeepEqual(firstValue, otherValue) {
		t.Fatalf("distinct event ids changed fields other than event_id: %#v != %#v", firstValue, otherValue)
	}
}

func TestMailEventMappers_RejectSharedExplicitEnvelopeInputs(t *testing.T) {
	calendarMethod := mail.MailCalendarMethodRequest
	sensitiveCorrelation := "SENSITIVE-CORRELATION-03D-" + strings.Repeat("x", 200)
	for _, testCase := range []struct {
		name   string
		mutate func(*uuid.UUID, *time.Time, *uuid.UUID, **uuid.UUID, *string)
		want   error
		marker string
	}{
		{
			name: "zero event id",
			mutate: func(eventID *uuid.UUID, _ *time.Time, _ *uuid.UUID, _ **uuid.UUID, _ *string) {
				*eventID = uuid.Nil
			},
			want: events.ErrInvalidEventID,
		},
		{
			name: "zero occurred at",
			mutate: func(_ *uuid.UUID, occurredAt *time.Time, _ *uuid.UUID, _ **uuid.UUID, _ *string) {
				*occurredAt = time.Time{}
			},
			want: events.ErrInvalidOccurredAt,
		},
		{
			name: "non-UTC occurred at",
			mutate: func(_ *uuid.UUID, occurredAt *time.Time, _ *uuid.UUID, _ **uuid.UUID, _ *string) {
				*occurredAt = time.Date(2030, time.May, 6, 7, 8, 9, 0, time.FixedZone("UTC+1", 3600))
			},
			want: events.ErrInvalidOccurredAt,
		},
		{
			name: "zero organization",
			mutate: func(_ *uuid.UUID, _ *time.Time, organizationID *uuid.UUID, _ **uuid.UUID, _ *string) {
				*organizationID = uuid.Nil
			},
			want: events.ErrInvalidOrganizationID,
		},
		{
			name: "zero actor",
			mutate: func(_ *uuid.UUID, _ *time.Time, _ *uuid.UUID, actorID **uuid.UUID, _ *string) {
				zero := uuid.Nil
				*actorID = &zero
			},
			want: events.ErrInvalidActorID,
		},
		{
			name: "blank correlation",
			mutate: func(_ *uuid.UUID, _ *time.Time, _ *uuid.UUID, _ **uuid.UUID, correlationID *string) {
				*correlationID = " \t "
			},
			want: events.ErrInvalidCorrelationID,
		},
		{
			name: "oversize correlation",
			mutate: func(_ *uuid.UUID, _ *time.Time, _ *uuid.UUID, _ **uuid.UUID, correlationID *string) {
				*correlationID = sensitiveCorrelation
			},
			want:   events.ErrInvalidCorrelationID,
			marker: sensitiveCorrelation,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			gotEventID := eventID
			gotOccurredAt := occurredAt
			gotOrganizationID := organizationID
			gotActorID := actorID
			actor := &gotActorID
			correlationID := "corr-mail-03d"
			testCase.mutate(&gotEventID, &gotOccurredAt, &gotOrganizationID, &actor, &correlationID)
			_, err := mail.NewMailReceivedEnvelopeV1(gotEventID, gotOccurredAt, gotOrganizationID, actor, correlationID, receivedPayload(&calendarMethod))
			requireEnvelopeInputFailure(t, err, testCase.want, testCase.marker)
		})
	}
}

func TestMailReceivedEnvelopeV1_AcceptsNullableCalendarMethodValues(t *testing.T) {
	cancel := mail.MailCalendarMethodCancel
	reply := mail.MailCalendarMethodReply
	for _, testCase := range []struct {
		name    string
		method  *mail.MailCalendarMethod
		wantRaw string
	}{
		{"nil", nil, `"calendar_method":null`},
		{"cancel", &cancel, `"calendar_method":"cancel"`},
		{"reply", &reply, `"calendar_method":"reply"`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			envelope, err := mail.NewMailReceivedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", receivedPayload(testCase.method))
			if err != nil {
				t.Fatalf("calendar method %s mapper error = %v", testCase.name, err)
			}
			encoded, err := json.Marshal(envelope)
			if err != nil {
				t.Fatalf("marshal calendar method %s = %v", testCase.name, err)
			}
			if !bytes.Contains(encoded, []byte(testCase.wantRaw)) {
				t.Fatalf("calendar method %s envelope = %s, missing %s", testCase.name, encoded, testCase.wantRaw)
			}
		})
	}
}

func TestMailSentEnvelopeV1_AcceptsDeliveryClassValues(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		class   mail.MailDeliveryClass
		wantRaw string
	}{
		{"internal", mail.MailDeliveryClassInternal, `"delivery_class":"internal"`},
		{"instance domain", mail.MailDeliveryClassInstanceDomain, `"delivery_class":"instance_domain"`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			payload := sentPayload()
			payload.DeliveryClass = testCase.class
			envelope, err := mail.NewMailSentEnvelopeV1(eventID, occurredAt, organizationID, nil, "corr-mail-03d", payload)
			if err != nil {
				t.Fatalf("delivery class %s mapper error = %v", testCase.name, err)
			}
			encoded, err := json.Marshal(envelope)
			if err != nil {
				t.Fatalf("marshal delivery class %s = %v", testCase.name, err)
			}
			if !bytes.Contains(encoded, []byte(testCase.wantRaw)) {
				t.Fatalf("delivery class %s envelope = %s, missing %s", testCase.name, encoded, testCase.wantRaw)
			}
		})
	}
}

func TestMailMessageChangedEnvelopeV1_AcceptsRemainingChangedFieldsAndFlags(t *testing.T) {
	payload := changedPayload()
	payload.ChangedFields = []mail.MailChangedField{
		mail.MailChangedFieldAnswered,
		mail.MailChangedFieldFlagged,
		mail.MailChangedFieldJunk,
	}
	payload.Flags = []mail.MailMessageFlag{
		mail.MailMessageFlagAnswered,
		mail.MailMessageFlagJunk,
		mail.MailMessageFlagDraft,
	}
	envelope, err := mail.NewMailMessageChangedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", payload)
	if err != nil {
		t.Fatalf("remaining changed field and flag mapper error = %v", err)
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal remaining changed field and flag envelope = %v", err)
	}
	for _, wantRaw := range []string{
		`"changed_fields":["answered","flagged","junk"]`,
		`"flags":["answered","junk","draft"]`,
	} {
		if !bytes.Contains(encoded, []byte(wantRaw)) {
			t.Fatalf("changed envelope = %s, missing %s", encoded, wantRaw)
		}
	}
}

func TestMailEventMappers_AcceptExactPayloadMaxima(t *testing.T) {
	calendarMethod := mail.MailCalendarMethodRequest
	for _, testCase := range []struct {
		name   string
		mapIt  func() (events.EnvelopeV1, error)
		verify func(*testing.T, map[string]json.RawMessage)
	}{
		{
			name: "received engine message and folders",
			mapIt: func() (events.EnvelopeV1, error) {
				payload := receivedPayload(&calendarMethod)
				payload.EngineMessage.ID = strings.Repeat("e", 512)
				payload.FolderIDs = eventIDs(50)
				return mail.NewMailReceivedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", payload)
			},
			verify: func(t *testing.T, payload map[string]json.RawMessage) {
				var engineMessageID string
				var folderIDs []uuid.UUID
				if err := json.Unmarshal(payload["engine_message_id"], &engineMessageID); err != nil {
					t.Fatalf("decode exact engine message max = %v", err)
				}
				if err := json.Unmarshal(payload["folder_ids"], &folderIDs); err != nil {
					t.Fatalf("decode exact received folder max = %v", err)
				}
				if len(engineMessageID) != 512 || len(folderIDs) != 50 {
					t.Fatalf("received maxima = engine:%d folders:%d, want engine:512 folders:50", len(engineMessageID), len(folderIDs))
				}
			},
		},
		{
			name: "changed folders and labels",
			mapIt: func() (events.EnvelopeV1, error) {
				payload := changedPayload()
				payload.FolderIDs = eventIDs(50)
				payload.LabelIDs = eventIDs(50)
				return mail.NewMailMessageChangedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", payload)
			},
			verify: func(t *testing.T, payload map[string]json.RawMessage) {
				var folderIDs, labelIDs []uuid.UUID
				if err := json.Unmarshal(payload["folder_ids"], &folderIDs); err != nil {
					t.Fatalf("decode exact changed folder max = %v", err)
				}
				if err := json.Unmarshal(payload["label_ids"], &labelIDs); err != nil {
					t.Fatalf("decode exact changed label max = %v", err)
				}
				if len(folderIDs) != 50 || len(labelIDs) != 50 {
					t.Fatalf("changed maxima = folders:%d labels:%d, want folders:50 labels:50", len(folderIDs), len(labelIDs))
				}
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			envelope, err := testCase.mapIt()
			if err != nil {
				t.Fatalf("exact payload maximum mapper error = %v", err)
			}
			encoded, err := json.Marshal(envelope)
			if err != nil {
				t.Fatalf("marshal exact payload maximum = %v", err)
			}
			testCase.verify(t, envelopePayload(t, encoded))
		})
	}
}

func TestMailReceivedEnvelopeV1_RejectsRequiredBoundsAndUnsafeInputs(t *testing.T) {
	calendarMethod := mail.MailCalendarMethodRequest
	foreignAccountID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	sensitiveEngineID := "SENSITIVE-ENGINE-ID-03D"
	sensitiveEmail := "SENSITIVE-EMAIL-03D"
	sensitiveName := "SENSITIVE-NAME-03D"
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailReceivedEventPayload)
		kind   mail.MailEventErrorKind
		marker string
	}{
		{"zero mailbox id", func(payload *mail.MailReceivedEventPayload) { payload.Mailbox.ID = uuid.Nil }, mail.MailEventErrorInvalid, ""},
		{"zero mailbox account", func(payload *mail.MailReceivedEventPayload) { payload.Mailbox.AccountID = uuid.Nil }, mail.MailEventErrorInvalid, ""},
		{"cross-account mailbox", func(payload *mail.MailReceivedEventPayload) { payload.Mailbox.AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch, ""},
		{"cross-account message", func(payload *mail.MailReceivedEventPayload) { payload.Message.AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch, ""},
		{"zero message id", func(payload *mail.MailReceivedEventPayload) { payload.Message.ID = uuid.Nil }, mail.MailEventErrorInvalid, ""},
		{"zero message account", func(payload *mail.MailReceivedEventPayload) { payload.Message.AccountID = uuid.Nil }, mail.MailEventErrorInvalid, ""},
		{"cross-account thread", func(payload *mail.MailReceivedEventPayload) { payload.Thread.AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch, ""},
		{"zero thread id", func(payload *mail.MailReceivedEventPayload) { payload.Thread.ID = uuid.Nil }, mail.MailEventErrorInvalid, ""},
		{"zero thread account", func(payload *mail.MailReceivedEventPayload) { payload.Thread.AccountID = uuid.Nil }, mail.MailEventErrorInvalid, ""},
		{"empty engine message id", func(payload *mail.MailReceivedEventPayload) { payload.EngineMessage.ID = "" }, mail.MailEventErrorInvalid, ""},
		{"invalid engine message id", func(payload *mail.MailReceivedEventPayload) { payload.EngineMessage.ID = sensitiveEngineID + string([]byte{0xff}) }, mail.MailEventErrorInvalid, sensitiveEngineID},
		{"oversize engine message id", func(payload *mail.MailReceivedEventPayload) { payload.EngineMessage.ID = sensitiveEngineID + strings.Repeat("x", 513) }, mail.MailEventErrorTooLarge, sensitiveEngineID},
		{"zero engine account", func(payload *mail.MailReceivedEventPayload) { payload.EngineMessage.AccountID = uuid.Nil }, mail.MailEventErrorInvalid, ""},
		{"cross-account engine", func(payload *mail.MailReceivedEventPayload) { payload.EngineMessage.AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch, ""},
		{"empty recipients", func(payload *mail.MailReceivedEventPayload) { payload.To = nil }, mail.MailEventErrorInvalid, ""},
		{"invalid sender email", func(payload *mail.MailReceivedEventPayload) { payload.From.Email = sensitiveEmail }, mail.MailEventErrorInvalid, sensitiveEmail},
		{"oversize sender email", func(payload *mail.MailReceivedEventPayload) { payload.From.Email = sensitiveEmail + strings.Repeat("a", 400) + "@example.test" }, mail.MailEventErrorTooLarge, sensitiveEmail},
		{"invalid sender name", func(payload *mail.MailReceivedEventPayload) { value := sensitiveName + "\x00"; payload.From.Name = &value }, mail.MailEventErrorInvalid, sensitiveName},
		{"oversize sender name", func(payload *mail.MailReceivedEventPayload) { value := sensitiveName + strings.Repeat("a", 700); payload.From.Name = &value }, mail.MailEventErrorTooLarge, sensitiveName},
		{"non-normalized sender", func(payload *mail.MailReceivedEventPayload) { payload.From.Email = "SENDER@example.test" }, mail.MailEventErrorInvalid, ""},
		{"duplicate recipients", func(payload *mail.MailReceivedEventPayload) { payload.To = append(payload.To, payload.To[0]) }, mail.MailEventErrorDuplicate, ""},
		{"too many recipients", func(payload *mail.MailReceivedEventPayload) { payload.To = eventAddresses(201) }, mail.MailEventErrorTooLarge, ""},
		{"empty folders", func(payload *mail.MailReceivedEventPayload) { payload.FolderIDs = nil }, mail.MailEventErrorInvalid, ""},
		{"duplicate folders", func(payload *mail.MailReceivedEventPayload) { payload.FolderIDs = append(payload.FolderIDs, payload.FolderIDs[0]) }, mail.MailEventErrorDuplicate, ""},
		{"zero folder", func(payload *mail.MailReceivedEventPayload) { payload.FolderIDs[0].ID = uuid.Nil }, mail.MailEventErrorInvalid, ""},
		{"zero folder account", func(payload *mail.MailReceivedEventPayload) { payload.FolderIDs[0].AccountID = uuid.Nil }, mail.MailEventErrorInvalid, ""},
		{"cross-account folder", func(payload *mail.MailReceivedEventPayload) { payload.FolderIDs[0].AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch, ""},
		{"too many folders", func(payload *mail.MailReceivedEventPayload) { payload.FolderIDs = eventIDs(51) }, mail.MailEventErrorTooLarge, ""},
		{"invalid calendar method", func(payload *mail.MailReceivedEventPayload) { value := mail.MailCalendarMethod("unknown"); payload.CalendarMethod = &value }, mail.MailEventErrorInvalid, ""},
		{"zero received at", func(payload *mail.MailReceivedEventPayload) { payload.ReceivedAt = time.Time{} }, mail.MailEventErrorInvalid, ""},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			payload := receivedPayload(&calendarMethod)
			testCase.mutate(&payload)
			_, err := mail.NewMailReceivedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", payload)
			requireMailEventFailure(t, err, testCase.kind, testCase.marker)
		})
	}
}

func TestMailSentEnvelopeV1_RejectsRequiredBoundsAndUnsafeInputs(t *testing.T) {
	foreignAccountID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailSentEventPayload)
		kind   mail.MailEventErrorKind
	}{
		{"zero mailbox id", func(payload *mail.MailSentEventPayload) { payload.Mailbox.ID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"zero mailbox account", func(payload *mail.MailSentEventPayload) { payload.Mailbox.AccountID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"cross-account mailbox", func(payload *mail.MailSentEventPayload) { payload.Mailbox.AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch},
		{"zero submission id", func(payload *mail.MailSentEventPayload) { payload.Submission.ID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"zero submission account", func(payload *mail.MailSentEventPayload) { payload.Submission.AccountID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"cross-account submission", func(payload *mail.MailSentEventPayload) { payload.Submission.AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch},
		{"zero message id", func(payload *mail.MailSentEventPayload) { payload.Message.ID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"zero message account", func(payload *mail.MailSentEventPayload) { payload.Message.AccountID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"cross-account message", func(payload *mail.MailSentEventPayload) { payload.Message.AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch},
		{"zero thread id", func(payload *mail.MailSentEventPayload) { payload.Thread.ID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"zero thread account", func(payload *mail.MailSentEventPayload) { payload.Thread.AccountID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"cross-account thread", func(payload *mail.MailSentEventPayload) { payload.Thread.AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch},
		{"empty recipients", func(payload *mail.MailSentEventPayload) { payload.To = nil }, mail.MailEventErrorInvalid},
		{"too many recipients", func(payload *mail.MailSentEventPayload) { payload.To = eventAddresses(201) }, mail.MailEventErrorTooLarge},
		{"duplicate recipient", func(payload *mail.MailSentEventPayload) { payload.To = append(payload.To, payload.To[0]) }, mail.MailEventErrorDuplicate},
		{"non-normalized recipient", func(payload *mail.MailSentEventPayload) { payload.To[0].Email = "RECIPIENT@example.test" }, mail.MailEventErrorInvalid},
		{"unknown delivery class", func(payload *mail.MailSentEventPayload) { payload.DeliveryClass = mail.MailDeliveryClass("unknown") }, mail.MailEventErrorInvalid},
		{"zero sent at", func(payload *mail.MailSentEventPayload) { payload.SentAt = time.Time{} }, mail.MailEventErrorInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			payload := sentPayload()
			testCase.mutate(&payload)
			_, err := mail.NewMailSentEnvelopeV1(eventID, occurredAt, organizationID, nil, "corr-mail-03d", payload)
			requireMailEventFailure(t, err, testCase.kind)
		})
	}
}

func TestMailEventMappers_DoNotLeakSentOrChangedInputMarkers(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mapIt  func(string) error
		kind   mail.MailEventErrorKind
		marker string
	}{
		{
			name: "sent malformed recipient",
			mapIt: func(marker string) error {
				payload := sentPayload()
				payload.To[0].Email = marker
				_, err := mail.NewMailSentEnvelopeV1(eventID, occurredAt, organizationID, nil, "corr-mail-03d", payload)
				return err
			},
			kind:   mail.MailEventErrorInvalid,
			marker: "SENSITIVE-SENT-RECIPIENT-03D",
		},
		{
			name: "sent unknown delivery class",
			mapIt: func(marker string) error {
				payload := sentPayload()
				payload.DeliveryClass = mail.MailDeliveryClass(marker)
				_, err := mail.NewMailSentEnvelopeV1(eventID, occurredAt, organizationID, nil, "corr-mail-03d", payload)
				return err
			},
			kind:   mail.MailEventErrorInvalid,
			marker: "SENSITIVE-SENT-DELIVERY-03D",
		},
		{
			name: "changed unknown field",
			mapIt: func(marker string) error {
				payload := changedPayload()
				payload.ChangedFields = []mail.MailChangedField{mail.MailChangedField(marker)}
				_, err := mail.NewMailMessageChangedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", payload)
				return err
			},
			kind:   mail.MailEventErrorInvalid,
			marker: "SENSITIVE-CHANGED-FIELD-03D",
		},
		{
			name: "changed unknown flag",
			mapIt: func(marker string) error {
				payload := changedPayload()
				payload.Flags = []mail.MailMessageFlag{mail.MailMessageFlag(marker)}
				_, err := mail.NewMailMessageChangedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", payload)
				return err
			},
			kind:   mail.MailEventErrorInvalid,
			marker: "SENSITIVE-CHANGED-FLAG-03D",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			requireMailEventFailure(t, testCase.mapIt(testCase.marker), testCase.kind, testCase.marker)
		})
	}
}

func TestMailEventMappers_AcceptRecipientSetBoundaries(t *testing.T) {
	calendarMethod := mail.MailCalendarMethodRequest
	for _, testCase := range []struct {
		name string
		mapIt func([]mail.MailEventAddress) (events.EnvelopeV1, error)
	}{
		{
			name: "received one and maximum",
			mapIt: func(addresses []mail.MailEventAddress) (events.EnvelopeV1, error) {
				payload := receivedPayload(&calendarMethod)
				payload.To = addresses
				return mail.NewMailReceivedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", payload)
			},
		},
		{
			name: "sent one and maximum",
			mapIt: func(addresses []mail.MailEventAddress) (events.EnvelopeV1, error) {
				payload := sentPayload()
				payload.To = addresses
				return mail.NewMailSentEnvelopeV1(eventID, occurredAt, organizationID, nil, "corr-mail-03d", payload)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			for _, count := range []int{1, 200} {
				if _, err := testCase.mapIt(eventAddresses(count)); err != nil {
					t.Fatalf("recipient count %d mapper error = %v", count, err)
				}
			}
		})
	}
}

func TestMailMessageChangedEnvelopeV1_RejectsInvalidAndDeletedLiveState(t *testing.T) {
	foreignAccountID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailMessageChangedEventPayload)
		kind   mail.MailEventErrorKind
	}{
		{"zero mailbox id", func(payload *mail.MailMessageChangedEventPayload) { payload.Mailbox.ID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"zero mailbox account", func(payload *mail.MailMessageChangedEventPayload) { payload.Mailbox.AccountID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"cross-account mailbox", func(payload *mail.MailMessageChangedEventPayload) { payload.Mailbox.AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch},
		{"zero message id", func(payload *mail.MailMessageChangedEventPayload) { payload.Message.ID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"cross-account message", func(payload *mail.MailMessageChangedEventPayload) { payload.Message.AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch},
		{"zero message account", func(payload *mail.MailMessageChangedEventPayload) { payload.Message.AccountID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"zero folder", func(payload *mail.MailMessageChangedEventPayload) { payload.FolderIDs[0].ID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"zero folder account", func(payload *mail.MailMessageChangedEventPayload) { payload.FolderIDs[0].AccountID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"cross-account folder", func(payload *mail.MailMessageChangedEventPayload) { payload.FolderIDs[0].AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch},
		{"duplicate folder", func(payload *mail.MailMessageChangedEventPayload) { payload.FolderIDs = append(payload.FolderIDs, payload.FolderIDs[0]) }, mail.MailEventErrorDuplicate},
		{"too many folders", func(payload *mail.MailMessageChangedEventPayload) { payload.FolderIDs = eventIDs(51) }, mail.MailEventErrorTooLarge},
		{"cross-account label", func(payload *mail.MailMessageChangedEventPayload) { payload.LabelIDs[0].AccountID = foreignAccountID }, mail.MailEventErrorAccountMismatch},
		{"zero label", func(payload *mail.MailMessageChangedEventPayload) { payload.LabelIDs[0].ID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"zero label account", func(payload *mail.MailMessageChangedEventPayload) { payload.LabelIDs[0].AccountID = uuid.Nil }, mail.MailEventErrorInvalid},
		{"empty changed fields", func(payload *mail.MailMessageChangedEventPayload) { payload.ChangedFields = nil }, mail.MailEventErrorInvalid},
		{"unknown changed field", func(payload *mail.MailMessageChangedEventPayload) { payload.ChangedFields = []mail.MailChangedField{"unknown"} }, mail.MailEventErrorInvalid},
		{"duplicate changed field", func(payload *mail.MailMessageChangedEventPayload) { payload.ChangedFields = []mail.MailChangedField{mail.MailChangedFieldSeen, mail.MailChangedFieldSeen} }, mail.MailEventErrorDuplicate},
		{"too many changed fields", func(payload *mail.MailMessageChangedEventPayload) { payload.ChangedFields = repeatedChangedFields(31) }, mail.MailEventErrorTooLarge},
		{"unknown flag", func(payload *mail.MailMessageChangedEventPayload) { payload.Flags = []mail.MailMessageFlag{"unknown"} }, mail.MailEventErrorInvalid},
		{"duplicate flag", func(payload *mail.MailMessageChangedEventPayload) { payload.Flags = []mail.MailMessageFlag{mail.MailMessageFlagSeen, mail.MailMessageFlagSeen} }, mail.MailEventErrorDuplicate},
		{"too many flags", func(payload *mail.MailMessageChangedEventPayload) { payload.Flags = repeatedFlags(31) }, mail.MailEventErrorTooLarge},
		{"duplicate labels", func(payload *mail.MailMessageChangedEventPayload) { payload.LabelIDs = append(payload.LabelIDs, payload.LabelIDs[0]) }, mail.MailEventErrorDuplicate},
		{"too many labels", func(payload *mail.MailMessageChangedEventPayload) { payload.LabelIDs = eventIDs(51) }, mail.MailEventErrorTooLarge},
		{"deleted with folder state", func(payload *mail.MailMessageChangedEventPayload) { payload.ChangedFields = []mail.MailChangedField{mail.MailChangedFieldDeleted}; payload.FolderIDs = eventIDs(1) }, mail.MailEventErrorInvalid},
		{"deleted with flag state", func(payload *mail.MailMessageChangedEventPayload) { payload.ChangedFields = []mail.MailChangedField{mail.MailChangedFieldDeleted}; payload.Flags = []mail.MailMessageFlag{mail.MailMessageFlagSeen} }, mail.MailEventErrorInvalid},
		{"deleted mixed with live change", func(payload *mail.MailMessageChangedEventPayload) { payload.ChangedFields = []mail.MailChangedField{mail.MailChangedFieldDeleted, mail.MailChangedFieldSeen} }, mail.MailEventErrorInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			payload := changedPayload()
			testCase.mutate(&payload)
			_, err := mail.NewMailMessageChangedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", payload)
			requireMailEventFailure(t, err, testCase.kind)
		})
	}

	deleted := changedPayload()
	deleted.ChangedFields = []mail.MailChangedField{mail.MailChangedFieldDeleted}
	deleted.FolderIDs = []mail.MailboxEventID{}
	deleted.Flags = []mail.MailMessageFlag{}
	envelope, err := mail.NewMailMessageChangedEnvelopeV1(eventID, occurredAt, organizationID, &actorID, "corr-mail-03d", deleted)
	if err != nil {
		t.Fatalf("deleted mapper error = %v", err)
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal deleted envelope = %v", err)
	}
	if !strings.Contains(string(encoded), `"changed_fields":["deleted"]`) || !strings.Contains(string(encoded), `"folder_ids":[]`) || !strings.Contains(string(encoded), `"flags":[]`) || !strings.Contains(string(encoded), `"label_ids":["aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"]`) {
		t.Fatalf("deleted envelope must retain only label state: %s", encoded)
	}
}

func TestMailEventMappers_ExposeOnlySafeBoundedInputShapes(t *testing.T) {
	assertExactExportedFields(t, reflect.TypeOf(mail.MailboxEventID{}), []string{"AccountID", "ID"})
	assertExactExportedFields(t, reflect.TypeOf(mail.MailboxEventOpaqueID{}), []string{"AccountID", "ID"})
	assertExactExportedFields(t, reflect.TypeOf(mail.MailEventAddress{}), []string{"Email", "Name"})
	assertExactExportedFields(t, reflect.TypeOf(mail.MailReceivedEventPayload{}), []string{"Mailbox", "Message", "EngineMessage", "Thread", "From", "To", "ReceivedAt", "FolderIDs", "HasAttachments", "CalendarMethod"})
	assertExactExportedFields(t, reflect.TypeOf(mail.MailSentEventPayload{}), []string{"Mailbox", "Submission", "Message", "Thread", "From", "To", "SentAt", "DeliveryClass"})
	assertExactExportedFields(t, reflect.TypeOf(mail.MailMessageChangedEventPayload{}), []string{"Mailbox", "Message", "ChangedFields", "FolderIDs", "Flags", "LabelIDs"})
	for _, valueType := range []reflect.Type{
		reflect.TypeOf(mail.MailReceivedEventPayload{}),
		reflect.TypeOf(mail.MailSentEventPayload{}),
		reflect.TypeOf(mail.MailMessageChangedEventPayload{}),
	} {
		for index := 0; index < valueType.NumField(); index++ {
			name := strings.ToLower(valueType.Field(index).Name)
			for _, forbidden := range []string{"body", "subject", "snippet", "mime", "header", "attachmentbytes", "filename", "upload", "provider", "credential", "token", "idempotency"} {
				if strings.Contains(name, forbidden) {
					t.Fatalf("%s exposes forbidden content field %q", valueType.Name(), valueType.Field(index).Name)
				}
			}
		}
	}
}

func receivedPayload(calendarMethod *mail.MailCalendarMethod) mail.MailReceivedEventPayload {
	return mail.MailReceivedEventPayload{
		Mailbox:       scopedID(mailboxID),
		Message:       scopedID(messageID),
		EngineMessage: mail.MailboxEventOpaqueID{AccountID: accountID, ID: "engine-message-03d"},
		Thread:        scopedID(threadID),
		From:          mail.MailEventAddress{Email: "sender@example.test", Name: stringPointer("Sender")},
		To:            []mail.MailEventAddress{{Email: "recipient@example.test"}},
		ReceivedAt:    payloadTime,
		FolderIDs:     []mail.MailboxEventID{scopedID(folderID)},
		HasAttachments: true,
		CalendarMethod: calendarMethod,
	}
}

func sentPayload() mail.MailSentEventPayload {
	return mail.MailSentEventPayload{
		Mailbox:       scopedID(mailboxID),
		Submission:    scopedID(submissionID),
		Message:       scopedID(messageID),
		Thread:        scopedID(threadID),
		From:          mail.MailEventAddress{Email: "sender@example.test", Name: stringPointer("Sender")},
		To:            []mail.MailEventAddress{{Email: "recipient@example.test"}},
		SentAt:        payloadTime,
		DeliveryClass: mail.MailDeliveryClassExternal,
	}
}

func changedPayload() mail.MailMessageChangedEventPayload {
	return mail.MailMessageChangedEventPayload{
		Mailbox:       scopedID(mailboxID),
		Message:       scopedID(messageID),
		ChangedFields: []mail.MailChangedField{mail.MailChangedFieldFolderIDs, mail.MailChangedFieldSeen, mail.MailChangedFieldLabels},
		FolderIDs:     []mail.MailboxEventID{scopedID(folderID)},
		Flags:         []mail.MailMessageFlag{mail.MailMessageFlagSeen, mail.MailMessageFlagFlagged},
		LabelIDs:      []mail.MailboxEventID{scopedID(labelID)},
	}
}

func scopedID(id uuid.UUID) mail.MailboxEventID {
	return mail.MailboxEventID{AccountID: accountID, ID: id}
}

func eventAddresses(count int) []mail.MailEventAddress {
	addresses := make([]mail.MailEventAddress, count)
	for index := range addresses {
		addresses[index] = mail.MailEventAddress{Email: "recipient-" + strings.TrimPrefix(uuid.NewSHA1(uuid.Nil, []byte{byte(index)}).String(), "") + "@example.test"}
	}
	return addresses
}

func eventIDs(count int) []mail.MailboxEventID {
	ids := make([]mail.MailboxEventID, count)
	for index := range ids {
		ids[index] = scopedID(uuid.NewSHA1(uuid.Nil, []byte{byte(index + 1)}))
	}
	return ids
}

func repeatedChangedFields(count int) []mail.MailChangedField {
	values := make([]mail.MailChangedField, count)
	for index := range values {
		values[index] = mail.MailChangedField("field-" + uuid.NewSHA1(uuid.Nil, []byte{byte(index)}).String())
	}
	return values
}

func repeatedFlags(count int) []mail.MailMessageFlag {
	values := make([]mail.MailMessageFlag, count)
	for index := range values {
		values[index] = mail.MailMessageFlag("flag-" + uuid.NewSHA1(uuid.Nil, []byte{byte(index)}).String())
	}
	return values
}

func requireMailEventFailure(t *testing.T, err error, wantKind mail.MailEventErrorKind, markers ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("mapper error = nil, want %q", wantKind)
	}
	if !errors.Is(err, mailEventSentinel(wantKind)) {
		t.Fatalf("mapper error = %v, want errors.Is(%v)", err, mailEventSentinel(wantKind))
	}
	var typed *mail.MailEventPolicyError
	if !errors.As(err, &typed) || typed.Kind != wantKind {
		t.Fatalf("mapper error = %#v, want MailEventPolicyError kind %q", err, wantKind)
	}
	for _, marker := range markers {
		assertErrorDoesNotLeakMarker(t, err, marker)
	}
}

func requireEnvelopeInputFailure(t *testing.T, err error, want error, marker string) {
	t.Helper()
	if err == nil {
		t.Fatalf("mapper error = nil, want errors.Is(%v)", want)
	}
	if !errors.Is(err, want) {
		t.Fatalf("mapper error = %v, want errors.Is(%v)", err, want)
	}
	assertErrorDoesNotLeakMarker(t, err, marker)
}

func assertErrorDoesNotLeakMarker(t *testing.T, err error, marker string) {
	t.Helper()
	if marker != "" && strings.Contains(err.Error(), marker) {
		t.Fatalf("mapper error leaked marker %q: %q", marker, err.Error())
	}
}

func mailEventSentinel(kind mail.MailEventErrorKind) error {
	switch kind {
	case mail.MailEventErrorInvalid:
		return mail.ErrMailEventInvalid
	case mail.MailEventErrorTooLarge:
		return mail.ErrMailEventTooLarge
	case mail.MailEventErrorDuplicate:
		return mail.ErrMailEventDuplicate
	case mail.MailEventErrorAccountMismatch:
		return mail.ErrMailEventAccountMismatch
	default:
		return nil
	}
}

func assertMailEnvelopeSchemaShape(t *testing.T, encoded []byte) {
	t.Helper()
	var decoded struct {
		Subject       string          `json:"subject"`
		SchemaVersion int             `json:"schema_version"`
		EventID       uuid.UUID       `json:"event_id"`
		OccurredAt    time.Time       `json:"occurred_at"`
		Organization  uuid.UUID       `json:"organization_id"`
		ActorID       *uuid.UUID      `json:"actor_id"`
		CorrelationID string          `json:"correlation_id"`
		Payload       json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("schema-shaped envelope is invalid JSON: %v", err)
	}
	if decoded.SchemaVersion != 1 || decoded.EventID == uuid.Nil || decoded.Organization == uuid.Nil || decoded.CorrelationID == "" || decoded.Payload == nil {
		t.Fatalf("envelope omitted required schema fields: %#v", decoded)
	}
	if !strings.HasSuffix(decoded.OccurredAt.Format(time.RFC3339Nano), "Z") {
		t.Fatalf("envelope occurred_at is not UTC Z: %s", decoded.OccurredAt.Format(time.RFC3339Nano))
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(decoded.Payload, &payload); err != nil {
		t.Fatalf("payload is not an object: %v", err)
	}
	for _, required := range []string{"mailbox_id", "message_id"} {
		if _, found := payload[required]; !found {
			t.Fatalf("payload for %s omitted required %s: %s", decoded.Subject, required, decoded.Payload)
		}
	}
}

func envelopePayload(t *testing.T, encoded []byte) map[string]json.RawMessage {
	t.Helper()
	var envelope struct {
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		t.Fatalf("decode envelope payload = %v", err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("decode payload object = %v", err)
	}
	return payload
}

func assertForbiddenContentAbsent(t *testing.T, encoded []byte) {
	t.Helper()
	for _, forbidden := range []string{"SENSITIVE-BODY-03D", "SENSITIVE-SUBJECT-03D", "SENSITIVE-SNIPPET-03D", "SENSITIVE-MIME-03D", "SENSITIVE-ATTACHMENT-03D", "SENSITIVE-CREDENTIAL-03D", "SENSITIVE-TOKEN-03D", "SENSITIVE-IDEMPOTENCY-03D"} {
		if bytes.Contains(encoded, []byte(forbidden)) {
			t.Fatalf("serialized envelope contains forbidden content %q: %s", forbidden, encoded)
		}
	}
}

func assertExactExportedFields(t *testing.T, valueType reflect.Type, want []string) {
	t.Helper()
	if valueType.NumField() != len(want) {
		t.Fatalf("%s field count = %d, want %d", valueType.Name(), valueType.NumField(), len(want))
	}
	for index, name := range want {
		field := valueType.Field(index)
		if field.Name != name || !field.IsExported() {
			t.Fatalf("%s field %d = %q exported=%t, want %q exported", valueType.Name(), index, field.Name, field.IsExported(), name)
		}
	}
}

func stringPointer(value string) *string {
	return &value
}
