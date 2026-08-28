package mail_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	ical "github.com/valerii/hub/backend/internal/ical"
	"github.com/valerii/hub/backend/internal/mail"
)

// TestCalendarMailDeliverySourceContract catches accidental widening of the
// Calendar-to-Mail boundary: the delivery service owns only a sealed durable
// claim and the two ordinary Mail service calls.
func TestCalendarMailDeliverySourceContract(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "internal", "mail"))
	repository := calendarMailDeliverySource(t, filepath.Join(root, "calendar_delivery_repository.go"))
	fake := calendarMailDeliverySource(t, filepath.Join(root, "calendar_delivery_fake.go"))
	service := calendarMailDeliverySource(t, filepath.Join(root, "calendar_delivery_service.go"))
	mimePart := calendarMailDeliverySource(t, filepath.Join(root, "calendar_mime_part.go"))

	for _, required := range []string{
		"AdmitCalendarInviteDeliveries",
		"WakeCalendarReplyDelivery",
		"ScanDueCalendarMailDeliveries",
		"ClaimCalendarMailDelivery",
		"BindCalendarMailDeliverySubmission",
		"ResolveCalendarMailDelivery",
		"CalendarMailDeliveryMethodRequest",
		"CalendarMailDeliveryMethodCancel",
		"CalendarMailDeliveryMethodReply",
		"CalendarMailDeliveryStatePending",
		"CalendarMailDeliveryStateSubmitting",
		"CalendarMailDeliveryStateUnknown",
		"CalendarMailDeliveryStateSent",
		"CalendarMailDeliveryStateSuperseded",
		"CalendarMailDeliveryStateFailedPermanent",
		"CalendarMailDeliveryClaimOutcomeObserved",
		"validCalendarMailDeliveryPreBindResolveCode",
		"senderActorID",
		"accountID",
		"fromIdentityID",
		"recipient",
		"organizationID",
		"leaseUntil",
		"validCalendarMailDeliveryLeaseUntil",
		"calendarMailDeliveryDefaultScanInterval = 30 * time.Second",
		"calendarMailDeliveryDefaultBatchSize = 50",
		"calendarMailDeliveryMaximumAttempts = 8",
		"calendarMailDeliveryRetryMinimum = 30 * time.Second",
		"calendarMailDeliveryRetryMaximum = 5 * time.Minute",
		"CalendarInviteDeliveryKindCreated",
		"CalendarInviteDeliveryKindUpdated",
		"CalendarInviteDeliveryKindCancelled",
		"SourceEnvelopeID",
		"ical.ITIPMessage",
		"ical.NormalizeITIPMessage",
		"ical.CanonicalITIPIdentity",
		"calendarMailDeliveryITIPMessageCopy",
	} {
		if !strings.Contains(repository, required) {
			t.Fatalf("calendar delivery repository source missing %q", required)
		}
	}
	if !strings.Contains(strings.Join(strings.Fields(repository), " "), "calendarMailDeliveryLease = 60 * time.Second") {
		t.Fatal("calendar delivery repository source missing normalized lease declaration")
	}
	for _, required := range []string{
		"hub.calendar.delivery.draft.v1/",
		"hub.calendar.delivery.submission.v1/",
		"calendarMailDeliveryDraftClientID",
		"string(deliveryID.Kind()) + \"/\" + deliveryID.UUID().String()",
		"CreateMailboxDraft",
		"CreateMailboxSubmission",
		"BindCalendarMailDeliverySubmission",
		"ResolveCalendarMailDelivery",
		"CalendarMailMaterializer",
		"SenderActorID",
		"CalendarMailDeliveryClaimOutcomeObserved",
		"safeCreateMailboxDraft",
		"safeCreateMailboxSubmission",
		"safeBind",
		"safeResolve",
		"context.WithoutCancel(ctx)",
		"calendarMailDeliveryResolveTimeout",
		"calendarMailDeliveryExternalContext",
		"calendarMailDeliveryLeaseSafetyMargin",
	} {
		if !strings.Contains(service, required) {
			t.Fatalf("calendar delivery service source missing %q", required)
		}
	}
	if !strings.Contains(mimePart, "calendarMIMEPartHashString(hash, string(sourceDeliveryID.Kind()))") {
		t.Fatal("calendar MIME part binding does not namespace the delivery kind")
	}
	for _, required := range []string{
		"SetNow",
		"calendarMailDeliveryFakeReclaim",
		"claim.claimGeneration++",
		"claim.claimToken = calendarMailDeliveryFakeClaimToken",
		"claim.leaseUntil = leaseUntil",
		"Message                    ical.ITIPMessage",
	} {
		if !strings.Contains(fake, required) {
			t.Fatalf("calendar delivery fake source missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"CalendarMailDeliveryBoundSubmissionDisposition",
		"boundSubmissionID",
		"resolveBoundSubmission",
		"calendarMailDeliverySubmissionResolveCode",
		"CalendarMailDeliveryResolveCodeSubmission",
	} {
		if strings.Contains(repository, forbidden) || strings.Contains(service, forbidden) {
			t.Fatalf("calendar delivery claim source retains forbidden bound-claim symbol %q", forbidden)
		}
	}
	for _, forbidden := range []string{"MailEngine", "Stalwart", "NATS", "entgo.io/ent", "NewMailboxEnt", "provider", "logger", "log."} {
		if strings.Contains(strings.ToLower(repository), strings.ToLower(forbidden)) ||
			strings.Contains(strings.ToLower(service), strings.ToLower(forbidden)) ||
			strings.Contains(strings.ToLower(fake), strings.ToLower(forbidden)) {
			t.Fatalf("calendar delivery source exposes forbidden capability %q", forbidden)
		}
	}
	if calls := strings.Count(service, ".CreateMailboxSubmission("); calls != 1 {
		t.Fatalf("CreateMailboxSubmission call sites = %d, want exactly one", calls)
	}
	if calls := strings.Count(service, ".CreateMailboxDraft("); calls != 1 {
		t.Fatalf("CreateMailboxDraft call sites = %d, want exactly one", calls)
	}
}

func TestCalendarMailDeliveryFakeReturnsNonNilEmptyDue(t *testing.T) {
	repository := mail.NewCalendarMailDeliveryRepositoryFake()

	result, err := repository.ScanDueCalendarMailDeliveries(context.Background(), mail.CalendarMailDeliveryDueScanRequest{Limit: 1})
	if err != nil {
		t.Fatalf("ScanDueCalendarMailDeliveries() error = %v", err)
	}
	if result.DeliveryIDs == nil {
		t.Fatal("ScanDueCalendarMailDeliveries() returned a nil empty due list")
	}
	if len(result.DeliveryIDs) != 0 {
		t.Fatalf("ScanDueCalendarMailDeliveries() due count = %d, want 0", len(result.DeliveryIDs))
	}
}

func TestCalendarMailDeliveryFakeRejectsInvalidSealedClaim(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name   string
		mutate func(*mail.CalendarMailDeliveryFakeClaim)
	}{
		{
			name: "missing organization",
			mutate: func(claim *mail.CalendarMailDeliveryFakeClaim) {
				claim.OrganizationID = uuid.Nil
			},
		},
		{
			name: "missing lease",
			mutate: func(claim *mail.CalendarMailDeliveryFakeClaim) {
				claim.LeaseUntil = time.Time{}
			},
		},
		{
			name: "non canonical lease zone",
			mutate: func(claim *mail.CalendarMailDeliveryFakeClaim) {
				claim.LeaseUntil = claim.LeaseUntil.In(time.FixedZone("non-utc", 0))
			},
		},
		{
			name: "missing message",
			mutate: func(claim *mail.CalendarMailDeliveryFakeClaim) {
				claim.Message = ical.ITIPMessage{}
			},
		},
		{
			name: "invite organizer absent",
			mutate: func(claim *mail.CalendarMailDeliveryFakeClaim) {
				claim.OrganizerID = uuid.Nil
			},
		},
		{
			name: "message sequence differs from delivery",
			mutate: func(claim *mail.CalendarMailDeliveryFakeClaim) {
				claim.EventSequence++
			},
		},
		{
			name: "invite recipient is not an attendee",
			mutate: func(claim *mail.CalendarMailDeliveryFakeClaim) {
				claim.Recipient = mail.EngineAddress{Email: "other-attendee@example.com"}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			deliveryID, repository := calendarDeliveryFixture(t)
			inviteID, err := mail.NewCalendarInviteDeliveryID(uuid.New())
			if err != nil {
				t.Fatalf("NewCalendarInviteDeliveryID() error = %v", err)
			}
			claim := calendarDeliveryFakeClaimFixture(t, deliveryID, inviteID, now.Add(time.Minute))
			test.mutate(&claim)

			if got := calendarDeliveryRepositoryErrorKind(t, repository.SetClaim(claim)); got != mail.CalendarMailDeliveryRepositoryErrorInvalid {
				t.Fatalf("SetClaim() error kind = %q, want %q", got, mail.CalendarMailDeliveryRepositoryErrorInvalid)
			}
		})
	}
}

func TestCalendarMailDeliveryIDNamespacesKindAndRejectsMismatchedClaim(t *testing.T) {
	for _, invalid := range []struct {
		kind  mail.CalendarMailDeliveryKind
		value uuid.UUID
	}{
		{kind: "", value: uuid.New()},
		{kind: mail.CalendarMailDeliveryKindInvite, value: uuid.Nil},
	} {
		if _, err := mail.NewCalendarMailDeliveryID(invalid.kind, invalid.value); err == nil {
			t.Fatalf("NewCalendarMailDeliveryID(%q, %s) error = nil, want invalid identity error", invalid.kind, invalid.value)
		}
	}
	var invalidDeliveryID mail.CalendarMailDeliveryID
	if invalidDeliveryID.Kind() != "" || invalidDeliveryID.UUID() != uuid.Nil {
		t.Fatalf("invalid delivery ID = kind %q, UUID %s, want zero values", invalidDeliveryID.Kind(), invalidDeliveryID.UUID())
	}

	value := uuid.New()
	inviteDeliveryID, err := mail.NewCalendarMailDeliveryID(mail.CalendarMailDeliveryKindInvite, value)
	if err != nil {
		t.Fatalf("NewCalendarMailDeliveryID(invite) error = %v", err)
	}
	replyDeliveryID, err := mail.NewCalendarMailDeliveryID(mail.CalendarMailDeliveryKindReply, value)
	if err != nil {
		t.Fatalf("NewCalendarMailDeliveryID(reply) error = %v", err)
	}
	if inviteDeliveryID == replyDeliveryID {
		t.Fatal("invite and reply delivery IDs with the same UUID compare equal")
	}
	if inviteDeliveryID.Kind() != mail.CalendarMailDeliveryKindInvite || replyDeliveryID.Kind() != mail.CalendarMailDeliveryKindReply {
		t.Fatalf("delivery ID kinds = %q, %q, want invite and reply", inviteDeliveryID.Kind(), replyDeliveryID.Kind())
	}
	if inviteDeliveryID.UUID() != value || replyDeliveryID.UUID() != value {
		t.Fatalf("delivery ID UUIDs = %s, %s, want %s", inviteDeliveryID.UUID(), replyDeliveryID.UUID(), value)
	}

	replyID, err := mail.NewCalendarReplyDeliveryID(uuid.New())
	if err != nil {
		t.Fatalf("NewCalendarReplyDeliveryID() error = %v", err)
	}
	claim := mail.CalendarMailDeliveryFakeClaim{
		DeliveryID: inviteDeliveryID, ReplyID: replyID, Kind: mail.CalendarMailDeliveryKindReply, Method: mail.CalendarMailDeliveryMethodReply,
		EventID: uuid.New(), AttendeeID: uuid.New(), OrganizerID: uuid.Nil, OrganizationID: uuid.New(), AccountID: uuid.New(),
		SenderActorID: uuid.New(), FromIdentityID: uuid.New(), Recipient: mail.EngineAddress{Email: "external-organizer@example.com"},
		EventSequence: 1, AttendeeVersion: 1, Message: calendarDeliveryTimedReplyMessage(), ClaimGeneration: 1, LeaseUntil: time.Now().UTC().Add(time.Minute),
	}
	repository := mail.NewCalendarMailDeliveryRepositoryFake()
	if got := calendarDeliveryRepositoryErrorKind(t, repository.SetClaim(claim)); got != mail.CalendarMailDeliveryRepositoryErrorInvalid {
		t.Fatalf("SetClaim() mismatched delivery kind error kind = %q, want %q", got, mail.CalendarMailDeliveryRepositoryErrorInvalid)
	}
}

func TestCalendarInviteDeliveryAdmissionRequiresKindSourceEnvelopeAndCopies(t *testing.T) {
	repository := mail.NewCalendarMailDeliveryRepositoryFake()
	inviteID, err := mail.NewCalendarInviteDeliveryID(uuid.New())
	if err != nil {
		t.Fatalf("NewCalendarInviteDeliveryID() error = %v", err)
	}
	request := mail.CalendarInviteDeliveryAdmissionRequest{Invites: []mail.CalendarInviteDeliveryAdmission{{
		InviteID: inviteID, SourceEnvelopeID: uuid.New(), Kind: mail.CalendarInviteDeliveryKindCreated,
		EventID: uuid.New(), AttendeeID: uuid.New(), OrganizerID: uuid.New(), Method: mail.CalendarMailDeliveryMethodRequest,
		EventSequence: 1,
	}}}

	for _, test := range []struct {
		name   string
		mutate func(*mail.CalendarInviteDeliveryAdmission)
	}{
		{
			name: "missing source envelope",
			mutate: func(admission *mail.CalendarInviteDeliveryAdmission) {
				admission.SourceEnvelopeID = uuid.Nil
			},
		},
		{
			name: "missing kind",
			mutate: func(admission *mail.CalendarInviteDeliveryAdmission) {
				admission.Kind = ""
			},
		},
		{
			name: "cancel kind with request method",
			mutate: func(admission *mail.CalendarInviteDeliveryAdmission) {
				admission.Kind = mail.CalendarInviteDeliveryKindCancelled
			},
		},
		{
			name: "created kind with cancel method",
			mutate: func(admission *mail.CalendarInviteDeliveryAdmission) {
				admission.Method = mail.CalendarMailDeliveryMethodCancel
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			invalid := request
			invalid.Invites = append([]mail.CalendarInviteDeliveryAdmission(nil), request.Invites...)
			test.mutate(&invalid.Invites[0])
			if got := calendarDeliveryRepositoryErrorKind(t, repository.AdmitCalendarInviteDeliveries(context.Background(), invalid)); got != mail.CalendarMailDeliveryRepositoryErrorInvalid {
				t.Fatalf("AdmitCalendarInviteDeliveries() error kind = %q, want %q", got, mail.CalendarMailDeliveryRepositoryErrorInvalid)
			}
		})
	}

	for _, valid := range []struct {
		kind   mail.CalendarInviteDeliveryKind
		method mail.CalendarMailDeliveryMethod
	}{
		{kind: mail.CalendarInviteDeliveryKindCreated, method: mail.CalendarMailDeliveryMethodRequest},
		{kind: mail.CalendarInviteDeliveryKindUpdated, method: mail.CalendarMailDeliveryMethodRequest},
		{kind: mail.CalendarInviteDeliveryKindCancelled, method: mail.CalendarMailDeliveryMethodCancel},
	} {
		admission := request
		admission.Invites = append([]mail.CalendarInviteDeliveryAdmission(nil), request.Invites...)
		admission.Invites[0].Kind = valid.kind
		admission.Invites[0].Method = valid.method
		if valid.kind != mail.CalendarInviteDeliveryKindCreated {
			admission.Invites[0].SourceEnvelopeID = uuid.New()
		}
		if err := repository.AdmitCalendarInviteDeliveries(context.Background(), admission); err != nil {
			t.Fatalf("AdmitCalendarInviteDeliveries(%s, %s) error = %v", valid.kind, valid.method, err)
		}
	}
	recorded := repository.Admissions()
	if len(recorded) != 3 {
		t.Fatalf("recorded admissions = %d, want 3", len(recorded))
	}
	wantSourceEnvelopeID := request.Invites[0].SourceEnvelopeID
	request.Invites[0].SourceEnvelopeID = uuid.New()
	if recorded[0].Invites[0].SourceEnvelopeID != wantSourceEnvelopeID {
		t.Fatalf("recorded admissions = %#v, want copied source envelope %s", recorded, wantSourceEnvelopeID)
	}
	recorded[0].Invites[0].SourceEnvelopeID = uuid.New()
	if again := repository.Admissions(); len(again) != 3 || again[0].Invites[0].SourceEnvelopeID != wantSourceEnvelopeID {
		t.Fatalf("Admissions() returned aliased source envelope: %#v", again)
	}
}

func TestCalendarInviteDeliveryInitialSequenceZeroAdmissionAndMaterialization(t *testing.T) {
	repository := mail.NewCalendarMailDeliveryRepositoryFake()
	inviteID, err := mail.NewCalendarInviteDeliveryID(uuid.New())
	if err != nil {
		t.Fatalf("NewCalendarInviteDeliveryID() error = %v", err)
	}
	admission := mail.CalendarInviteDeliveryAdmissionRequest{Invites: []mail.CalendarInviteDeliveryAdmission{{
		InviteID: inviteID, SourceEnvelopeID: uuid.New(), Kind: mail.CalendarInviteDeliveryKindCreated,
		EventID: uuid.New(), AttendeeID: uuid.New(), OrganizerID: uuid.New(), Method: mail.CalendarMailDeliveryMethodRequest,
		EventSequence: 0,
	}}}
	if err := repository.AdmitCalendarInviteDeliveries(context.Background(), admission); err != nil {
		t.Fatalf("AdmitCalendarInviteDeliveries(initial SEQUENCE:0) error = %v", err)
	}
	negativeAdmission := admission
	negativeAdmission.Invites = append([]mail.CalendarInviteDeliveryAdmission(nil), admission.Invites...)
	negativeAdmission.Invites[0].EventSequence = -1
	if got := calendarDeliveryRepositoryErrorKind(t, repository.AdmitCalendarInviteDeliveries(context.Background(), negativeAdmission)); got != mail.CalendarMailDeliveryRepositoryErrorInvalid {
		t.Fatalf("AdmitCalendarInviteDeliveries(negative sequence) error kind = %q, want %q", got, mail.CalendarMailDeliveryRepositoryErrorInvalid)
	}

	deliveryID, err := mail.NewCalendarMailDeliveryID(mail.CalendarMailDeliveryKindInvite, uuid.New())
	if err != nil {
		t.Fatalf("NewCalendarMailDeliveryID() error = %v", err)
	}
	message := calendarDeliveryInitialTimedRequestMessage()
	if message.Sequence != 0 {
		t.Fatalf("initial iTIP sequence = %d, want 0", message.Sequence)
	}
	claim := calendarDeliveryFakeClaimFixture(t, deliveryID, inviteID, time.Now().UTC().Add(time.Minute))
	claim.EventSequence = 0
	claim.Message = message
	if err := repository.SetClaim(claim); err != nil {
		t.Fatalf("SetClaim(initial SEQUENCE:0) error = %v", err)
	}
	negativeClaim := claim
	negativeClaim.EventSequence = -1
	negativeClaim.Message.Sequence = -1
	if got := calendarDeliveryRepositoryErrorKind(t, repository.SetClaim(negativeClaim)); got != mail.CalendarMailDeliveryRepositoryErrorInvalid {
		t.Fatalf("SetClaim(negative sequence) error kind = %q, want %q", got, mail.CalendarMailDeliveryRepositoryErrorInvalid)
	}
}

func TestCalendarMailDeliveryReplyAllowsExternalOrganizerOnly(t *testing.T) {
	deliveryID, err := mail.NewCalendarMailDeliveryID(mail.CalendarMailDeliveryKindReply, uuid.New())
	if err != nil {
		t.Fatalf("NewCalendarMailDeliveryID() error = %v", err)
	}
	replyID, err := mail.NewCalendarReplyDeliveryID(uuid.New())
	if err != nil {
		t.Fatalf("NewCalendarReplyDeliveryID() error = %v", err)
	}
	claim := mail.CalendarMailDeliveryFakeClaim{
		DeliveryID: deliveryID, ReplyID: replyID, Kind: mail.CalendarMailDeliveryKindReply, Method: mail.CalendarMailDeliveryMethodReply,
		EventID: uuid.New(), AttendeeID: uuid.New(), OrganizerID: uuid.Nil, OrganizationID: uuid.New(), AccountID: uuid.New(),
		SenderActorID: uuid.New(), FromIdentityID: uuid.New(), Recipient: mail.EngineAddress{Email: "external-organizer@example.com"},
		EventSequence: 1, AttendeeVersion: 1, Message: calendarDeliveryTimedReplyMessage(), ClaimGeneration: 1, LeaseUntil: time.Now().UTC().Add(time.Minute),
	}
	repository := mail.NewCalendarMailDeliveryRepositoryFake()
	if err := repository.SetClaim(claim); err != nil {
		t.Fatalf("SetClaim() external-organizer reply error = %v", err)
	}
	claim.OrganizerID = uuid.New()
	if got := calendarDeliveryRepositoryErrorKind(t, repository.SetClaim(claim)); got != mail.CalendarMailDeliveryRepositoryErrorInvalid {
		t.Fatalf("SetClaim() internal-organizer reply error kind = %q, want %q", got, mail.CalendarMailDeliveryRepositoryErrorInvalid)
	}
	claim.OrganizerID = uuid.Nil
	claim.Recipient = mail.EngineAddress{Email: "not-the-organizer@example.com"}
	if got := calendarDeliveryRepositoryErrorKind(t, repository.SetClaim(claim)); got != mail.CalendarMailDeliveryRepositoryErrorInvalid {
		t.Fatalf("SetClaim() reply recipient error kind = %q, want %q", got, mail.CalendarMailDeliveryRepositoryErrorInvalid)
	}
}

func TestCalendarMailDeliveryFakeCopiesCanonicalMessageSnapshot(t *testing.T) {
	deliveryID, repository := calendarDeliveryFixture(t)
	inviteID, err := mail.NewCalendarInviteDeliveryID(uuid.New())
	if err != nil {
		t.Fatalf("NewCalendarInviteDeliveryID() error = %v", err)
	}
	claim := calendarDeliveryFakeClaimFixture(t, deliveryID, inviteID, time.Now().UTC().Add(time.Minute))
	if err := repository.SetClaim(claim); err != nil {
		t.Fatalf("SetClaim() error = %v", err)
	}
	claim.Message.Attendees[0].Email = "mutated@example.com"

	mailIO := &calendarDeliveryMailProbe{}
	service := calendarDeliveryService(t, repository, mailIO, mailIO, &calendarDeliveryMaterializerProbe{})
	if err := service.Deliver(context.Background(), deliveryID); err != nil {
		t.Fatalf("Deliver() after source message mutation error = %v", err)
	}
	if mailIO.draftCalls != 1 || mailIO.submissionCalls != 1 {
		t.Fatalf("Mail I/O after source message mutation = drafts:%d submissions:%d, want one each", mailIO.draftCalls, mailIO.submissionCalls)
	}
}

func TestCalendarMailDeliveryNoopSkipsMailIO(t *testing.T) {
	repository := mail.NewCalendarMailDeliveryRepositoryFake()
	mailIO := &calendarDeliveryMailProbe{}
	materializer := &calendarDeliveryMaterializerProbe{}
	service := calendarDeliveryService(t, repository, mailIO, mailIO, materializer)
	deliveryID, err := mail.NewCalendarMailDeliveryID(mail.CalendarMailDeliveryKindInvite, uuid.New())
	if err != nil {
		t.Fatalf("NewCalendarMailDeliveryID() error = %v", err)
	}

	if err := service.Deliver(context.Background(), deliveryID); err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if mailIO.draftCalls != 0 || mailIO.submissionCalls != 0 || materializer.calls != 0 {
		t.Fatalf("noop delivery performed Mail I/O: drafts=%d submissions=%d materializations=%d", mailIO.draftCalls, mailIO.submissionCalls, materializer.calls)
	}
}

func TestCalendarMailDeliveryExpiredLeaseSkipsMailIOAndResolves(t *testing.T) {
	leaseUntil := time.Now().UTC().Add(-2 * time.Second)
	deliveryID, repository := calendarDeliveryFixtureWithLease(t, leaseUntil.Add(-time.Second), leaseUntil)
	mailIO := &calendarDeliveryMailProbe{}
	materializer := &calendarDeliveryMaterializerProbe{}
	service := calendarDeliveryService(t, repository, mailIO, mailIO, materializer)

	if err := service.Deliver(context.Background(), deliveryID); err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if mailIO.draftCalls != 0 || mailIO.submissionCalls != 0 || materializer.calls != 0 {
		t.Fatalf("expired lease performed Mail I/O: drafts=%d submissions=%d materializations=%d", mailIO.draftCalls, mailIO.submissionCalls, materializer.calls)
	}
	resolutions := repository.Resolutions()
	if len(resolutions) != 1 || resolutions[0].Code != mail.CalendarMailDeliveryResolveCodeMailboxUnavailable {
		t.Fatalf("expired lease resolutions = %#v, want one mailbox_unavailable resolution", resolutions)
	}
}

func TestCalendarMailDeliveryMailIOUsesSharedLeaseBoundContext(t *testing.T) {
	now := time.Now().UTC()
	leaseUntil := now.Add(time.Minute)
	deliveryID, repository := calendarDeliveryFixtureWithLease(t, now, leaseUntil)
	mailIO := &calendarDeliveryMailProbe{}
	service := calendarDeliveryService(t, repository, mailIO, mailIO, &calendarDeliveryMaterializerProbe{})

	if err := service.Deliver(context.Background(), deliveryID); err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if len(mailIO.draftContexts) != 1 || len(mailIO.submissionContexts) != 1 {
		t.Fatalf("Mail contexts = drafts:%d submissions:%d, want one each", len(mailIO.draftContexts), len(mailIO.submissionContexts))
	}
	if mailIO.draftContexts[0] != mailIO.submissionContexts[0] {
		t.Fatal("draft and submission received different lease-bound contexts")
	}
	wantDeadline := leaseUntil.Add(-time.Second)
	for _, callContext := range []context.Context{mailIO.draftContexts[0], mailIO.submissionContexts[0]} {
		deadline, ok := callContext.Deadline()
		if !ok || !deadline.Equal(wantDeadline) {
			t.Fatalf("Mail I/O deadline = %s (present=%t), want %s", deadline, ok, wantDeadline)
		}
	}
}

func TestCalendarMailDeliveryMailIOPreservesEarlierParentDeadline(t *testing.T) {
	now := time.Now().UTC()
	leaseUntil := now.Add(2 * time.Minute)
	parentDeadline := now.Add(30 * time.Second)
	deliveryID, repository := calendarDeliveryFixtureWithLease(t, now, leaseUntil)
	mailIO := &calendarDeliveryMailProbe{}
	service := calendarDeliveryService(t, repository, mailIO, mailIO, &calendarDeliveryMaterializerProbe{})
	ctx, cancel := context.WithDeadline(context.Background(), parentDeadline)
	defer cancel()

	if err := service.Deliver(ctx, deliveryID); err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if len(mailIO.draftContexts) != 1 || len(mailIO.submissionContexts) != 1 {
		t.Fatalf("Mail contexts = drafts:%d submissions:%d, want one each", len(mailIO.draftContexts), len(mailIO.submissionContexts))
	}
	for _, callContext := range []context.Context{mailIO.draftContexts[0], mailIO.submissionContexts[0]} {
		deadline, ok := callContext.Deadline()
		if !ok || !deadline.Equal(parentDeadline) {
			t.Fatalf("Mail I/O deadline = %s (present=%t), want earlier parent deadline %s", deadline, ok, parentDeadline)
		}
	}
}

func TestCalendarMailDeliveryFakeExpiredReclaimFencesStaleClaim(t *testing.T) {
	now := time.Unix(1, 0).UTC()
	leaseUntil := now.Add(time.Second)
	deliveryID, repository := calendarDeliveryFixtureWithLease(t, now, leaseUntil)

	originalResult, err := repository.ClaimCalendarMailDelivery(context.Background(), deliveryID)
	if err != nil {
		t.Fatalf("initial ClaimCalendarMailDelivery() error = %v", err)
	}
	originalClaim, ok := originalResult.Claim()
	if !ok {
		t.Fatal("initial ClaimCalendarMailDelivery() did not return a claim")
	}
	if err := repository.SetNow(leaseUntil); err != nil {
		t.Fatalf("SetNow() error = %v", err)
	}
	reclaimedResult, err := repository.ClaimCalendarMailDelivery(context.Background(), deliveryID)
	if err != nil {
		t.Fatalf("reclaimed ClaimCalendarMailDelivery() error = %v", err)
	}
	reclaimedClaim, ok := reclaimedResult.Claim()
	if !ok {
		t.Fatal("reclaimed ClaimCalendarMailDelivery() did not return a claim")
	}
	if _, err := repository.BindCalendarMailDeliverySubmission(context.Background(), mail.CalendarMailDeliverySubmissionBindRequest{
		Claim: originalClaim, SubmissionID: uuid.New(),
	}); calendarDeliveryRepositoryErrorKind(t, err) != mail.CalendarMailDeliveryRepositoryErrorInvalid {
		t.Fatalf("stale BindCalendarMailDeliverySubmission() error = %v, want invalid repository error", err)
	}
	if err := repository.ResolveCalendarMailDelivery(context.Background(), mail.CalendarMailDeliveryResolveRequest{
		Claim: originalClaim, Code: mail.CalendarMailDeliveryResolveCodeMailboxUnavailable,
	}); calendarDeliveryRepositoryErrorKind(t, err) != mail.CalendarMailDeliveryRepositoryErrorInvalid {
		t.Fatalf("stale ResolveCalendarMailDelivery() error = %v, want invalid repository error", err)
	}
	bound, err := repository.BindCalendarMailDeliverySubmission(context.Background(), mail.CalendarMailDeliverySubmissionBindRequest{
		Claim: reclaimedClaim, SubmissionID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("reclaimed BindCalendarMailDeliverySubmission() error = %v", err)
	}
	if got := bound.Outcome(); got != mail.CalendarMailDeliverySubmissionBindOutcomeBound {
		t.Fatalf("reclaimed bind outcome = %q, want %q", got, mail.CalendarMailDeliverySubmissionBindOutcomeBound)
	}
}

func TestCalendarMailDeliveryObservedSkipsMailIO(t *testing.T) {
	deliveryID, repository := calendarDeliveryFixture(t)
	mailIO := &calendarDeliveryMailProbe{}
	materializer := &calendarDeliveryMaterializerProbe{}
	service := calendarDeliveryService(t, repository, mailIO, mailIO, materializer)

	if err := service.Deliver(context.Background(), deliveryID); err != nil {
		t.Fatalf("first Deliver() error = %v", err)
	}
	mailIO.draftCalls = 0
	mailIO.submissionCalls = 0
	materializer.calls = 0

	if err := service.Deliver(context.Background(), deliveryID); err != nil {
		t.Fatalf("observed Deliver() error = %v", err)
	}
	if mailIO.draftCalls != 0 || mailIO.submissionCalls != 0 || materializer.calls != 0 {
		t.Fatalf("observed delivery performed Mail I/O: drafts=%d submissions=%d materializations=%d", mailIO.draftCalls, mailIO.submissionCalls, materializer.calls)
	}
	if resolutions := repository.Resolutions(); len(resolutions) != 0 {
		t.Fatalf("observed delivery resolutions = %d, want 0", len(resolutions))
	}
}

func TestCalendarMailDeliveryNamespacesDurableKeysAndMaterializationByKind(t *testing.T) {
	now := time.Now().UTC()
	leaseUntil := now.Add(time.Minute)
	sharedDeliveryUUID := uuid.New()
	inviteDeliveryID, err := mail.NewCalendarMailDeliveryID(mail.CalendarMailDeliveryKindInvite, sharedDeliveryUUID)
	if err != nil {
		t.Fatalf("NewCalendarMailDeliveryID(invite) error = %v", err)
	}
	replyDeliveryID, err := mail.NewCalendarMailDeliveryID(mail.CalendarMailDeliveryKindReply, sharedDeliveryUUID)
	if err != nil {
		t.Fatalf("NewCalendarMailDeliveryID(reply) error = %v", err)
	}
	inviteID, err := mail.NewCalendarInviteDeliveryID(uuid.New())
	if err != nil {
		t.Fatalf("NewCalendarInviteDeliveryID() error = %v", err)
	}
	replyID, err := mail.NewCalendarReplyDeliveryID(uuid.New())
	if err != nil {
		t.Fatalf("NewCalendarReplyDeliveryID() error = %v", err)
	}

	repository := mail.NewCalendarMailDeliveryRepositoryFake()
	if err := repository.SetNow(now); err != nil {
		t.Fatalf("SetNow() error = %v", err)
	}
	inviteClaim := calendarDeliveryFakeClaimFixture(t, inviteDeliveryID, inviteID, leaseUntil)
	inviteClaim.Recipient = mail.EngineAddress{Email: "external-organizer@example.com"}
	inviteClaim.Message.Attendees[0].Email = inviteClaim.Recipient.Email
	if err := repository.SetClaim(inviteClaim); err != nil {
		t.Fatalf("SetClaim(invite) error = %v", err)
	}
	replyClaim := mail.CalendarMailDeliveryFakeClaim{
		DeliveryID:       replyDeliveryID,
		ReplyID:          replyID,
		Kind:             mail.CalendarMailDeliveryKindReply,
		Method:           mail.CalendarMailDeliveryMethodReply,
		EventID:          inviteClaim.EventID,
		AttendeeID:       inviteClaim.AttendeeID,
		OrganizationID:   inviteClaim.OrganizationID,
		AccountID:        inviteClaim.AccountID,
		SenderActorID:    inviteClaim.SenderActorID,
		FromIdentityID:   inviteClaim.FromIdentityID,
		Recipient:        inviteClaim.Recipient,
		EventSequence:    1,
		AttendeeVersion:  1,
		Message:          calendarDeliveryTimedReplyMessage(),
		ClaimGeneration:  1,
		LeaseUntil:       leaseUntil,
	}
	if err := repository.SetClaim(replyClaim); err != nil {
		t.Fatalf("SetClaim(reply) error = %v", err)
	}

	mailIO := &calendarDeliveryMailProbe{}
	materializer := &calendarDeliveryMaterializerProbe{}
	service := calendarDeliveryService(t, repository, mailIO, mailIO, materializer)
	for _, deliveryID := range []mail.CalendarMailDeliveryID{inviteDeliveryID, replyDeliveryID} {
		if err := service.Deliver(context.Background(), deliveryID); err != nil {
			t.Fatalf("Deliver(%s) error = %v", deliveryID.Kind(), err)
		}
	}
	if len(mailIO.draftInputs) != 2 || len(mailIO.submissionInputs) != 2 || len(materializer.inputs) != 2 {
		t.Fatalf("delivery observations = drafts:%d submissions:%d materializations:%d, want two each", len(mailIO.draftInputs), len(mailIO.submissionInputs), len(materializer.inputs))
	}

	wantInviteDraftKey := "hub.calendar.delivery.draft.v1/invite/" + sharedDeliveryUUID.String()
	wantReplyDraftKey := "hub.calendar.delivery.draft.v1/reply/" + sharedDeliveryUUID.String()
	if mailIO.draftInputs[0].IdempotencyKey != wantInviteDraftKey || mailIO.draftInputs[1].IdempotencyKey != wantReplyDraftKey {
		t.Fatalf("draft idempotency keys = %q, %q; want %q, %q", mailIO.draftInputs[0].IdempotencyKey, mailIO.draftInputs[1].IdempotencyKey, wantInviteDraftKey, wantReplyDraftKey)
	}
	if mailIO.draftInputs[0].DraftClientID == mailIO.draftInputs[1].DraftClientID {
		t.Fatal("invite and reply delivery IDs with the same UUID reused a DraftClientID")
	}
	wantInviteSubmissionKey := "hub.calendar.delivery.submission.v1/invite/" + sharedDeliveryUUID.String()
	wantReplySubmissionKey := "hub.calendar.delivery.submission.v1/reply/" + sharedDeliveryUUID.String()
	if mailIO.submissionInputs[0].IdempotencyKey != wantInviteSubmissionKey || mailIO.submissionInputs[1].IdempotencyKey != wantReplySubmissionKey {
		t.Fatalf("submission idempotency keys = %q, %q; want %q, %q", mailIO.submissionInputs[0].IdempotencyKey, mailIO.submissionInputs[1].IdempotencyKey, wantInviteSubmissionKey, wantReplySubmissionKey)
	}
	if reflect.DeepEqual(materializer.inputs[0], materializer.inputs[1]) {
		t.Fatal("invite and reply delivery materialization inputs were cross-replayed")
	}
	if reflect.DeepEqual(mailIO.draftInputs[0], mailIO.draftInputs[1]) {
		t.Fatal("invite and reply MIME-bound draft inputs were cross-replayed")
	}
}

func TestCalendarMailDeliveryBindsKnownSubmissionDespiteErrorWithoutResolve(t *testing.T) {
	deliveryID, repository := calendarDeliveryFixture(t)
	submissionID := uuid.New()
	mailIO := &calendarDeliveryMailProbe{
		submissionFunc: func(context.Context, uuid.UUID, uuid.UUID, mail.MailboxSubmissionCreateInput) (mail.MailboxSubmissionView, error) {
			return mail.MailboxSubmissionView{ID: submissionID}, errors.New("submission response interrupted after durable ID")
		},
	}
	materializer := &calendarDeliveryMaterializerProbe{}
	service := calendarDeliveryService(t, repository, mailIO, mailIO, materializer)

	if err := service.Deliver(context.Background(), deliveryID); err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if resolutions := repository.Resolutions(); len(resolutions) != 0 {
		t.Fatalf("known submission resolutions = %d, want 0", len(resolutions))
	}

	if err := service.Deliver(context.Background(), deliveryID); err != nil {
		t.Fatalf("second Deliver() error = %v", err)
	}
	if mailIO.submissionCalls != 1 {
		t.Fatalf("CreateMailboxSubmission calls = %d, want 1 after observed retry", mailIO.submissionCalls)
	}
	if resolutions := repository.Resolutions(); len(resolutions) != 0 {
		t.Fatalf("observed retry resolutions = %d, want 0", len(resolutions))
	}
}

func TestCalendarMailDeliveryBindErrorLeavesKnownSubmissionUnresolved(t *testing.T) {
	deliveryID, fake := calendarDeliveryFixture(t)
	repository := &calendarDeliveryRepositorySpy{
		CalendarMailDeliveryRepository: fake,
		bindErr:                        errors.New("bind unavailable"),
	}
	mailIO := &calendarDeliveryMailProbe{}
	service := calendarDeliveryService(t, repository, mailIO, mailIO, &calendarDeliveryMaterializerProbe{})

	err := service.Deliver(context.Background(), deliveryID)
	if got := calendarDeliveryServiceErrorKind(t, err); got != mail.CalendarMailDeliveryServiceErrorUnavailable {
		t.Fatalf("Deliver() error kind = %q, want %q", got, mail.CalendarMailDeliveryServiceErrorUnavailable)
	}
	if repository.bindCalls != 1 {
		t.Fatalf("BindCalendarMailDeliverySubmission calls = %d, want 1", repository.bindCalls)
	}
	if resolutions := fake.Resolutions(); len(resolutions) != 0 {
		t.Fatalf("bind-error resolutions = %d, want 0", len(resolutions))
	}
}

func TestCalendarMailDeliveryPanicsResolvePreBindMailboxUnavailable(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*calendarDeliveryMailProbe)
	}{
		{
			name: "draft",
			configure: func(probe *calendarDeliveryMailProbe) {
				probe.draftFunc = func(context.Context, uuid.UUID, mail.MailboxDraftCreateInput) (mail.MailboxDraftView, error) {
					panic("draft panic")
				}
			},
		},
		{
			name: "submission",
			configure: func(probe *calendarDeliveryMailProbe) {
				probe.submissionFunc = func(context.Context, uuid.UUID, uuid.UUID, mail.MailboxSubmissionCreateInput) (mail.MailboxSubmissionView, error) {
					panic("submission panic")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			deliveryID, repository := calendarDeliveryFixture(t)
			mailIO := &calendarDeliveryMailProbe{}
			test.configure(mailIO)
			service := calendarDeliveryService(t, repository, mailIO, mailIO, &calendarDeliveryMaterializerProbe{})

			if err := service.Deliver(context.Background(), deliveryID); err != nil {
				t.Fatalf("Deliver() error = %v", err)
			}
			resolutions := repository.Resolutions()
			if len(resolutions) != 1 {
				t.Fatalf("resolutions = %d, want 1", len(resolutions))
			}
			if resolutions[0].Code != mail.CalendarMailDeliveryResolveCodeMailboxUnavailable {
				t.Fatalf("resolution code = %q, want %q", resolutions[0].Code, mail.CalendarMailDeliveryResolveCodeMailboxUnavailable)
			}
		})
	}
}

func TestCalendarMailDeliveryResolveDetachesCancellationAndPreservesValues(t *testing.T) {
	deliveryID, fake := calendarDeliveryFixture(t)
	repository := &calendarDeliveryRepositorySpy{CalendarMailDeliveryRepository: fake}
	parent := context.WithValue(context.Background(), calendarDeliveryContextKey{}, "delivery-scope")
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	mailIO := &calendarDeliveryMailProbe{
		draftFunc: func(context.Context, uuid.UUID, mail.MailboxDraftCreateInput) (mail.MailboxDraftView, error) {
			cancel()
			return mail.MailboxDraftView{}, errors.New("mailbox unavailable")
		},
	}
	service := calendarDeliveryService(t, repository, mailIO, mailIO, &calendarDeliveryMaterializerProbe{})

	if err := service.Deliver(ctx, deliveryID); err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if repository.resolveCalls != 1 {
		t.Fatalf("ResolveCalendarMailDelivery calls = %d, want 1", repository.resolveCalls)
	}
	if repository.resolveCanceled {
		t.Fatal("ResolveCalendarMailDelivery received the canceled parent context")
	}
	if repository.resolveValue != "delivery-scope" {
		t.Fatalf("ResolveCalendarMailDelivery context value = %v, want delivery-scope", repository.resolveValue)
	}
	if !repository.resolveHasDeadline {
		t.Fatal("ResolveCalendarMailDelivery context has no deadline")
	}
	if !repository.resolveDeadline.After(repository.resolveObservedAt) ||
		repository.resolveDeadline.After(repository.resolveObservedAt.Add(5*time.Second)) {
		t.Fatalf("ResolveCalendarMailDelivery deadline = %s, want a future deadline no later than five seconds", repository.resolveDeadline)
	}
	resolutions := fake.Resolutions()
	if len(resolutions) != 1 || resolutions[0].Code != mail.CalendarMailDeliveryResolveCodeMailboxUnavailable {
		t.Fatalf("resolution = %#v, want one mailbox_unavailable resolution", resolutions)
	}
}

func calendarMailDeliverySource(t *testing.T, path string) string {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read calendar delivery source %q: %v", path, err)
	}
	return string(source)
}

func calendarDeliveryFixture(t *testing.T) (mail.CalendarMailDeliveryID, *mail.CalendarMailDeliveryRepositoryFake) {
	t.Helper()
	now := time.Now().UTC()
	return calendarDeliveryFixtureWithLease(t, now, now.Add(time.Minute))
}

func calendarDeliveryFixtureWithLease(t *testing.T, now, leaseUntil time.Time) (mail.CalendarMailDeliveryID, *mail.CalendarMailDeliveryRepositoryFake) {
	t.Helper()
	deliveryID, err := mail.NewCalendarMailDeliveryID(mail.CalendarMailDeliveryKindInvite, uuid.New())
	if err != nil {
		t.Fatalf("NewCalendarMailDeliveryID() error = %v", err)
	}
	inviteID, err := mail.NewCalendarInviteDeliveryID(uuid.New())
	if err != nil {
		t.Fatalf("NewCalendarInviteDeliveryID() error = %v", err)
	}
	repository := mail.NewCalendarMailDeliveryRepositoryFake()
	if err := repository.SetNow(now); err != nil {
		t.Fatalf("SetNow() error = %v", err)
	}
	if err := repository.SetClaim(calendarDeliveryFakeClaimFixture(t, deliveryID, inviteID, leaseUntil)); err != nil {
		t.Fatalf("SetClaim() error = %v", err)
	}
	return deliveryID, repository
}

func calendarDeliveryFakeClaimFixture(t *testing.T, deliveryID mail.CalendarMailDeliveryID, inviteID mail.CalendarInviteDeliveryID, leaseUntil time.Time) mail.CalendarMailDeliveryFakeClaim {
	t.Helper()
	return mail.CalendarMailDeliveryFakeClaim{
		DeliveryID:       deliveryID,
		InviteID:         inviteID,
		Kind:             mail.CalendarMailDeliveryKindInvite,
		Method:           mail.CalendarMailDeliveryMethodRequest,
		EventID:          uuid.New(),
		AttendeeID:       uuid.New(),
		OrganizerID:      uuid.New(),
		OrganizationID:   uuid.New(),
		AccountID:        uuid.New(),
		SenderActorID:    uuid.New(),
		FromIdentityID:   uuid.New(),
		Recipient:        mail.EngineAddress{Email: "attendee@example.com"},
		EventSequence:    1,
		Message:          calendarDeliveryTimedRequestMessage(),
		ClaimGeneration: 1,
		LeaseUntil:       leaseUntil,
	}
}

func calendarDeliveryTimedRequestMessage() ical.ITIPMessage {
	return ical.ITIPMessage{
		Method:   ical.ITIPMethodRequest,
		UID:      "calendar-delivery-request@example.test",
		Stamp:    time.Date(2026, time.August, 25, 9, 0, 0, 0, time.UTC),
		Sequence: 1,
		Status:   ical.ITIPStatusConfirmed,
		Summary:  "Calendar delivery",
		Temporal: ical.ITIPTemporal{Timed: &ical.ITIPTimed{
			StartsAt: time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC),
			EndsAt:   time.Date(2026, time.August, 26, 10, 0, 0, 0, time.UTC),
			Timezone: "UTC",
		}},
		Organizer: ical.ITIPOrganizer{Email: "organizer@example.com"},
		Attendees: []ical.ITIPParticipant{{
			Email:  "attendee@example.com",
			Role:   ical.ITIPParticipantRequired,
			Status: ical.RSVPNeedsAction,
			RSVP:   true,
		}},
	}
}

func calendarDeliveryInitialTimedRequestMessage() ical.ITIPMessage {
	message := calendarDeliveryTimedRequestMessage()
	message.Sequence = 0
	return message
}

func calendarDeliveryTimedReplyMessage() ical.ITIPMessage {
	return ical.ITIPMessage{
		Method:   ical.ITIPMethodReply,
		UID:      "calendar-delivery-reply@example.test",
		Stamp:    time.Date(2026, time.August, 25, 9, 0, 0, 0, time.UTC),
		Sequence: 1,
		Status:   ical.ITIPStatusConfirmed,
		Summary:  "Calendar reply",
		Temporal: ical.ITIPTemporal{Timed: &ical.ITIPTimed{
			StartsAt: time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC),
			EndsAt:   time.Date(2026, time.August, 26, 10, 0, 0, 0, time.UTC),
			Timezone: "UTC",
		}},
		Organizer: ical.ITIPOrganizer{Email: "external-organizer@example.com"},
		Attendees: []ical.ITIPParticipant{{
			Email:  "attendee@example.com",
			Role:   ical.ITIPParticipantRequired,
			Status: ical.RSVPAccepted,
			RSVP:   true,
		}},
	}
}

func calendarDeliveryService(t *testing.T, repository mail.CalendarMailDeliveryRepository, drafts mail.CalendarMailDeliveryDraftCreator, submissions mail.CalendarMailDeliverySubmissionCreator, materializer mail.CalendarMailMaterializer) *mail.CalendarMailDeliveryService {
	t.Helper()
	service, err := mail.NewCalendarMailDeliveryService(
		repository,
		drafts,
		submissions,
		materializer,
		calendarDeliveryTimer{},
		mail.CalendarMailDeliveryServiceOptions{Enabled: true, ScanInterval: time.Second, BatchSize: 1},
	)
	if err != nil {
		t.Fatalf("NewCalendarMailDeliveryService() error = %v", err)
	}
	return service
}

func calendarDeliveryServiceErrorKind(t *testing.T, err error) mail.CalendarMailDeliveryServiceErrorKind {
	t.Helper()
	var target *mail.CalendarMailDeliveryServiceError
	if !errors.As(err, &target) || target == nil {
		t.Fatalf("error = %v, want CalendarMailDeliveryServiceError", err)
	}
	return target.Kind()
}

func calendarDeliveryRepositoryErrorKind(t *testing.T, err error) mail.CalendarMailDeliveryRepositoryErrorKind {
	t.Helper()
	var target *mail.CalendarMailDeliveryRepositoryError
	if !errors.As(err, &target) || target == nil {
		t.Fatalf("error = %v, want CalendarMailDeliveryRepositoryError", err)
	}
	return target.Kind()
}

type calendarDeliveryTimer struct{}

func (calendarDeliveryTimer) Wait(context.Context, time.Duration) error {
	return nil
}

type calendarDeliveryMailProbe struct {
	draftCalls        int
	submissionCalls   int
	draftContexts     []context.Context
	submissionContexts []context.Context
	draftInputs       []mail.MailboxDraftCreateInput
	submissionInputs  []mail.MailboxSubmissionCreateInput
	draftFunc         func(context.Context, uuid.UUID, mail.MailboxDraftCreateInput) (mail.MailboxDraftView, error)
	submissionFunc    func(context.Context, uuid.UUID, uuid.UUID, mail.MailboxSubmissionCreateInput) (mail.MailboxSubmissionView, error)
}

func (probe *calendarDeliveryMailProbe) CreateMailboxDraft(ctx context.Context, actorID uuid.UUID, input mail.MailboxDraftCreateInput) (mail.MailboxDraftView, error) {
	probe.draftCalls++
	probe.draftContexts = append(probe.draftContexts, ctx)
	probe.draftInputs = append(probe.draftInputs, input)
	if probe.draftFunc != nil {
		return probe.draftFunc(ctx, actorID, input)
	}
	return mail.MailboxDraftView{ID: uuid.New(), DraftClientID: input.DraftClientID, Revision: 1}, nil
}

func (probe *calendarDeliveryMailProbe) CreateMailboxSubmission(ctx context.Context, actorID, draftID uuid.UUID, input mail.MailboxSubmissionCreateInput) (mail.MailboxSubmissionView, error) {
	probe.submissionCalls++
	probe.submissionContexts = append(probe.submissionContexts, ctx)
	probe.submissionInputs = append(probe.submissionInputs, input)
	if probe.submissionFunc != nil {
		return probe.submissionFunc(ctx, actorID, draftID, input)
	}
	return mail.MailboxSubmissionView{ID: uuid.New()}, nil
}

type calendarDeliveryMaterializerProbe struct {
	calls  int
	inputs []mail.CalendarMailDeliveryMaterializationInput
}

func (probe *calendarDeliveryMaterializerProbe) MaterializeCalendarMail(input mail.CalendarMailDeliveryMaterializationInput) (mail.CalendarMailMaterialization, error) {
	probe.calls++
	probe.inputs = append(probe.inputs, input)
	return mail.NewCalendarMailMaterialization("Calendar update", mail.MailboxContentInput{
		SourceFormat: mail.MailboxContentFormatPlain,
		Body:         "Calendar delivery",
	})
}

type calendarDeliveryContextKey struct{}

type calendarDeliveryRepositorySpy struct {
	mail.CalendarMailDeliveryRepository

	bindErr            error
	bindCalls          int
	resolveCalls       int
	resolveCanceled    bool
	resolveValue       any
	resolveDeadline    time.Time
	resolveHasDeadline bool
	resolveObservedAt  time.Time
}

func (spy *calendarDeliveryRepositorySpy) BindCalendarMailDeliverySubmission(ctx context.Context, request mail.CalendarMailDeliverySubmissionBindRequest) (mail.CalendarMailDeliverySubmissionBindResult, error) {
	spy.bindCalls++
	if spy.bindErr != nil {
		return mail.CalendarMailDeliverySubmissionBindResult{}, spy.bindErr
	}
	return spy.CalendarMailDeliveryRepository.BindCalendarMailDeliverySubmission(ctx, request)
}

func (spy *calendarDeliveryRepositorySpy) ResolveCalendarMailDelivery(ctx context.Context, request mail.CalendarMailDeliveryResolveRequest) error {
	spy.resolveCalls++
	spy.resolveCanceled = ctx.Err() != nil
	spy.resolveValue = ctx.Value(calendarDeliveryContextKey{})
	spy.resolveDeadline, spy.resolveHasDeadline = ctx.Deadline()
	spy.resolveObservedAt = time.Now()
	return spy.CalendarMailDeliveryRepository.ResolveCalendarMailDelivery(ctx, request)
}
