package mail_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

var (
	mailboxSubmissionSentEventAccountID      = uuid.MustParse("01010101-0101-4101-8101-010101010101")
	mailboxSubmissionSentEventOrganizationID = uuid.MustParse("02020202-0202-4202-8202-020202020202")
	mailboxSubmissionSentEventActorID        = uuid.MustParse("03030303-0303-4303-8303-030303030303")
	mailboxSubmissionSentEventSubmissionID   = uuid.MustParse("04040404-0404-4404-8404-040404040404")
	mailboxSubmissionSentEventMessageID      = uuid.MustParse("05050505-0505-4505-8505-050505050505")
	mailboxSubmissionSentEventThreadID       = uuid.MustParse("06060606-0606-4606-8606-060606060606")
)

func TestMailboxSubmissionSentEnvelopeV1_DeterministicReplayUsesSubmissionUUIDv8(t *testing.T) {
	input := mailboxSubmissionSentEventFixture()
	first, err := mail.NewMailboxSubmissionSentEnvelopeV1(input)
	if err != nil {
		t.Fatalf("NewMailboxSubmissionSentEnvelopeV1() first error = %v", err)
	}
	second, err := mail.NewMailboxSubmissionSentEnvelopeV1(input)
	if err != nil {
		t.Fatalf("NewMailboxSubmissionSentEnvelopeV1() second error = %v", err)
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first envelope = %v", err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("marshal second envelope = %v", err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("replay bytes differ: %s != %s", firstJSON, secondJSON)
	}
	if first.EventID() != mailboxSubmissionSentEventExpectedID(input.SubmissionID) {
		t.Fatalf("event id = %s, want submission-derived id", first.EventID())
	}
	if first.EventID().Version() != uuid.Version(8) || first.EventID().Variant() != uuid.RFC4122 {
		t.Fatalf("event id %s is not UUIDv8/RFC4122", first.EventID())
	}
	assertMailboxSubmissionSentEventJSON(t, firstJSON, input, mail.MailDeliveryClassExternal, []string{"recipient@example.test"})
}

func TestMailboxSubmissionSentEnvelopeV1_SubmissionIDIsTheOnlyEventIDInput(t *testing.T) {
	firstInput := mailboxSubmissionSentEventFixture()
	secondInput := firstInput
	secondInput.SubmissionID = uuid.MustParse("07070707-0707-4707-8707-070707070707")
	first, err := mail.NewMailboxSubmissionSentEnvelopeV1(firstInput)
	if err != nil {
		t.Fatalf("first event error = %v", err)
	}
	second, err := mail.NewMailboxSubmissionSentEnvelopeV1(secondInput)
	if err != nil {
		t.Fatalf("second event error = %v", err)
	}
	if first.EventID() == second.EventID() {
		t.Fatalf("distinct submission ids produced event id %s", first.EventID())
	}
	if second.EventID() != mailboxSubmissionSentEventExpectedID(secondInput.SubmissionID) {
		t.Fatalf("second event id = %s, want submission-derived id", second.EventID())
	}
}

func TestMailboxSubmissionSentEnvelopeV1_ReducesRecipientRoutesConservatively(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		mutate  func(*mail.MailboxSubmissionSentEventInput)
		want    mail.MailDeliveryClass
	}{
		{
			name: "mixed routes prefer external",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) {
				input.Projection.CC = []mail.EngineAddress{{Email: "instance@example.test"}}
				input.Projection.BCC = []mail.EngineAddress{{Email: "external@example.net"}}
				input.RecipientResolutions = []mail.MailboxSubmissionSentRecipientResolution{
					{Email: "recipient@example.test", DeliveryClass: mail.MailDeliveryClassInternal},
					{Email: "instance@example.test", DeliveryClass: mail.MailDeliveryClassInstanceDomain},
					{Email: "external@example.net", DeliveryClass: mail.MailDeliveryClassExternal},
				}
			},
			want: mail.MailDeliveryClassExternal,
		},
		{
			name: "all internal",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) {
				input.RecipientResolutions[0].DeliveryClass = mail.MailDeliveryClassInternal
			},
			want: mail.MailDeliveryClassInternal,
		},
		{
			name: "instance domain",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) {
				input.RecipientResolutions[0].DeliveryClass = mail.MailDeliveryClassInstanceDomain
			},
			want: mail.MailDeliveryClassInstanceDomain,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := mailboxSubmissionSentEventFixture()
			testCase.mutate(&input)
			envelope, err := mail.NewMailboxSubmissionSentEnvelopeV1(input)
			if err != nil {
				t.Fatalf("NewMailboxSubmissionSentEnvelopeV1() error = %v", err)
			}
			encoded, err := json.Marshal(envelope)
			if err != nil {
				t.Fatalf("marshal envelope = %v", err)
			}
			assertMailboxSubmissionSentEventJSON(t, encoded, input, testCase.want, []string{"recipient@example.test"})
		})
	}
}

func TestMailboxSubmissionSentEnvelopeV1_UsesExactVisibleToWithoutLeakingCCOrBCC(t *testing.T) {
	input := mailboxSubmissionSentEventFixture()
	input.Projection.To = []mail.EngineAddress{}
	input.Projection.BCC = []mail.EngineAddress{{Email: "blind@example.net"}}
	input.RecipientResolutions = []mail.MailboxSubmissionSentRecipientResolution{{Email: "blind@example.net", DeliveryClass: mail.MailDeliveryClassExternal}}
	envelope, err := mail.NewMailboxSubmissionSentEnvelopeV1(input)
	if err != nil {
		t.Fatalf("NewMailboxSubmissionSentEnvelopeV1() error = %v", err)
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope = %v", err)
	}
	assertMailboxSubmissionSentEventJSON(t, encoded, input, mail.MailDeliveryClassExternal, []string{})
	for _, forbidden := range []string{"blind@example.net", "\"cc\"", "\"bcc\""} {
		if bytes.Contains(encoded, []byte(forbidden)) {
			t.Fatalf("sent envelope leaked private recipient field/value %q: %s", forbidden, encoded)
		}
	}
}

func TestMailboxSubmissionSentEnvelopeV1_RejectsRecipientResolutionCoverageFailures(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailboxSubmissionSentEventInput)
		secret string
	}{
		{
			name:   "nil resolutions",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) { input.RecipientResolutions = nil },
		},
		{
			name:   "missing route",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) { input.RecipientResolutions = []mail.MailboxSubmissionSentRecipientResolution{} },
		},
		{
			name: "extra route",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) {
				input.RecipientResolutions = append(input.RecipientResolutions, mail.MailboxSubmissionSentRecipientResolution{Email: "extra@example.net", DeliveryClass: mail.MailDeliveryClassExternal})
			},
			secret: "extra@example.net",
		},
		{
			name: "duplicate route",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) {
				input.RecipientResolutions = append(input.RecipientResolutions, input.RecipientResolutions[0])
			},
		},
		{
			name: "malformed route email",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) {
				input.RecipientResolutions[0].Email = "NOT-CANONICAL@example.test"
			},
			secret: "NOT-CANONICAL@example.test",
		},
		{
			name: "unknown delivery class",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) {
				input.RecipientResolutions[0].DeliveryClass = mail.MailDeliveryClass("unknown")
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := mailboxSubmissionSentEventFixture()
			testCase.mutate(&input)
			_, err := mail.NewMailboxSubmissionSentEnvelopeV1(input)
			requireMailboxSubmissionSentEventInvalid(t, err, testCase.secret)
		})
	}
}

func TestMailboxSubmissionSentEnvelopeV1_RejectsMissingOrInvalidConfirmedProjection(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailboxSubmissionSentEventInput)
		secret string
	}{
		{
			name: "no recipients",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) {
				input.Projection.To = []mail.EngineAddress{}
				input.Projection.CC = []mail.EngineAddress{}
				input.Projection.BCC = []mail.EngineAddress{}
				input.RecipientResolutions = []mail.MailboxSubmissionSentRecipientResolution{}
			},
		},
		{
			name:   "unconfirmed origin",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) { input.Projection.Origin = mail.EngineProjectionOriginInboundOrUnknown },
		},
		{
			name:   "draft flag",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) { input.Projection.Flags.Draft = true },
		},
		{
			name:   "missing sent at",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) { input.Projection.SentAt = nil },
		},
		{
			name: "non UTC sent at",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) {
				value := time.Date(2001, time.February, 3, 5, 6, 7, 0, time.FixedZone("UTC+1", 3600))
				input.Projection.SentAt = &value
			},
		},
		{
			name: "noncanonical projection recipient",
			mutate: func(input *mail.MailboxSubmissionSentEventInput) {
				input.Projection.To = []mail.EngineAddress{{Email: "RECIPIENT@example.test"}}
			},
			secret: "RECIPIENT@example.test",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := mailboxSubmissionSentEventFixture()
			testCase.mutate(&input)
			_, err := mail.NewMailboxSubmissionSentEnvelopeV1(input)
			requireMailboxSubmissionSentEventInvalid(t, err, testCase.secret)
		})
	}
}

func TestMailboxSubmissionSentEnvelopeV1_RejectsInvalidHubIDsAndDispatchAuthorizationTime(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailboxSubmissionSentEventInput)
	}{
		{"zero account", func(input *mail.MailboxSubmissionSentEventInput) { input.AccountID = uuid.Nil }},
		{"zero organization", func(input *mail.MailboxSubmissionSentEventInput) { input.OrganizationID = uuid.Nil }},
		{"zero actor", func(input *mail.MailboxSubmissionSentEventInput) { input.ActorID = uuid.Nil }},
		{"zero submission", func(input *mail.MailboxSubmissionSentEventInput) { input.SubmissionID = uuid.Nil }},
		{"zero message", func(input *mail.MailboxSubmissionSentEventInput) { input.MessageID = uuid.Nil }},
		{"zero thread", func(input *mail.MailboxSubmissionSentEventInput) { input.ThreadID = uuid.Nil }},
		{"zero authorization time", func(input *mail.MailboxSubmissionSentEventInput) { input.DispatchAuthorizedAt = time.Time{} }},
		{"pre epoch authorization time", func(input *mail.MailboxSubmissionSentEventInput) { input.DispatchAuthorizedAt = time.Date(1969, time.December, 31, 23, 59, 59, 0, time.UTC) }},
		{"non utc authorization time", func(input *mail.MailboxSubmissionSentEventInput) { input.DispatchAuthorizedAt = time.Date(2001, time.February, 3, 4, 5, 6, 0, time.FixedZone("UTC+1", 3600)) }},
		{"subsecond authorization time", func(input *mail.MailboxSubmissionSentEventInput) { input.DispatchAuthorizedAt = time.Date(2001, time.February, 3, 4, 5, 6, 1, time.UTC) }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := mailboxSubmissionSentEventFixture()
			testCase.mutate(&input)
			_, err := mail.NewMailboxSubmissionSentEnvelopeV1(input)
			requireMailboxSubmissionSentEventInvalid(t, err, "")
		})
	}
}

func TestMailboxSubmissionSentEventSource_IsPureAndDeterministic(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	sourcePath := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "internal", "mail", "mailbox_submission_sent_event.go"))
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read source %s: %v", sourcePath, err)
	}
	for _, forbidden := range []string{
		"time.Now(", "uuid.New(", "os.", "http.", "nats.", "ent.", "MailEngine", "GetMessage(", "GetBlob(", "Submit(", "ReconcileSubmission(",
	} {
		if strings.Contains(string(source), forbidden) {
			t.Fatalf("source contains forbidden nondeterministic or I/O marker %q", forbidden)
		}
	}
}

func mailboxSubmissionSentEventFixture() mail.MailboxSubmissionSentEventInput {
	sentAt := time.Date(2001, time.February, 3, 4, 5, 7, 0, time.UTC)
	return mail.MailboxSubmissionSentEventInput{
		AccountID:             mailboxSubmissionSentEventAccountID,
		OrganizationID:        mailboxSubmissionSentEventOrganizationID,
		ActorID:               mailboxSubmissionSentEventActorID,
		SubmissionID:          mailboxSubmissionSentEventSubmissionID,
		MessageID:             mailboxSubmissionSentEventMessageID,
		ThreadID:              mailboxSubmissionSentEventThreadID,
		DispatchAuthorizedAt:  time.Date(2001, time.February, 3, 4, 5, 6, 0, time.UTC),
		Projection: mail.EngineMessageProjection{
			AccountID:         "engine-account",
			MessageID:         "engine-message",
			Version:           "engine-version",
			References:        []string{},
			From:              mail.EngineAddress{Email: "sender@example.test", Name: "Sender"},
			To:                []mail.EngineAddress{{Email: "recipient@example.test"}},
			CC:                []mail.EngineAddress{},
			BCC:               []mail.EngineAddress{},
			ReplyTo:           []mail.EngineAddress{},
			Subject:           "Sent event fixture",
			Snippet:           "fixture",
			SentAt:            &sentAt,
			AuthenticationClass: mail.MailAuthenticationClassUntrusted,
			SizeBytes:         7,
			Priority:          mail.EnginePriorityNormal,
			FolderIDs:         []string{"sent"},
			ReservedKeywords:  []mail.EngineKeywordMarker{},
			Attachments:       []mail.EngineProjectionAttachment{},
			Origin:            mail.EngineProjectionOriginConfirmedOwnerSubmission,
		},
		RecipientResolutions: []mail.MailboxSubmissionSentRecipientResolution{
			{Email: "recipient@example.test", DeliveryClass: mail.MailDeliveryClassExternal},
		},
	}
}

func mailboxSubmissionSentEventExpectedID(submissionID uuid.UUID) uuid.UUID {
	const domain = "hub.mail.submission.sent.event.v1"
	frame := make([]byte, 0, 4+len(domain)+len(submissionID))
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(domain)))
	frame = append(frame, length[:]...)
	frame = append(frame, domain...)
	frame = append(frame, submissionID[:]...)
	digest := sha256.Sum256(frame)
	result := uuid.Must(uuid.FromBytes(digest[:16]))
	result[6] = result[6]&0x0f | 0x80
	result[8] = result[8]&0x3f | 0x80
	return result
}

func assertMailboxSubmissionSentEventJSON(t *testing.T, encoded []byte, input mail.MailboxSubmissionSentEventInput, deliveryClass mail.MailDeliveryClass, visibleTo []string) {
	t.Helper()
	var actual struct {
		Subject        string    `json:"subject"`
		EventID        uuid.UUID `json:"event_id"`
		OccurredAt     time.Time `json:"occurred_at"`
		OrganizationID uuid.UUID `json:"organization_id"`
		ActorID        uuid.UUID `json:"actor_id"`
		CorrelationID  string    `json:"correlation_id"`
		Payload        struct {
			MailboxID     uuid.UUID               `json:"mailbox_id"`
			SubmissionID  uuid.UUID               `json:"submission_id"`
			MessageID     uuid.UUID               `json:"message_id"`
			ThreadID      uuid.UUID               `json:"thread_id"`
			To            []struct{ Email string `json:"email"` } `json:"to"`
			SentAt        time.Time               `json:"sent_at"`
			DeliveryClass mail.MailDeliveryClass  `json:"delivery_class"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(encoded, &actual); err != nil {
		t.Fatalf("unmarshal envelope = %v", err)
	}
	if actual.Subject != "mail.sent.v1" || actual.EventID != mailboxSubmissionSentEventExpectedID(input.SubmissionID) || actual.OccurredAt != input.DispatchAuthorizedAt || actual.OrganizationID != input.OrganizationID || actual.ActorID != input.ActorID || actual.CorrelationID != input.SubmissionID.String() {
		t.Fatalf("envelope header = %#v, want frozen submission event values", actual)
	}
	if actual.Payload.MailboxID != input.AccountID || actual.Payload.SubmissionID != input.SubmissionID || actual.Payload.MessageID != input.MessageID || actual.Payload.ThreadID != input.ThreadID || actual.Payload.SentAt != *input.Projection.SentAt || actual.Payload.DeliveryClass != deliveryClass {
		t.Fatalf("event payload identity = %#v, want input-bound sent event values", actual.Payload)
	}
	if len(actual.Payload.To) != len(visibleTo) {
		t.Fatalf("visible to = %#v, want %v", actual.Payload.To, visibleTo)
	}
	for index, email := range visibleTo {
		if actual.Payload.To[index].Email != email {
			t.Fatalf("visible to[%d] = %q, want %q", index, actual.Payload.To[index].Email, email)
		}
	}
}

func requireMailboxSubmissionSentEventInvalid(t *testing.T, err error, secret string) {
	t.Helper()
	if err == nil || !errors.Is(err, mail.ErrMailEventInvalid) {
		t.Fatalf("error = %v, want content-free mail event invalid", err)
	}
	if secret != "" && strings.Contains(err.Error(), secret) {
		t.Fatalf("error leaked input content %q: %v", secret, err)
	}
}
