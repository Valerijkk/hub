package mail_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestMailboxSyncPolicy_ExportsFrozenABIAndSafeErrors(t *testing.T) {
	policyInputType := reflect.TypeOf(mail.MailboxSyncPolicyInput{})
	requireMailboxSyncFields(t, policyInputType, []string{"Page", "Folders", "Messages", "Destroyed", "Labels", "SyncEpoch", "EventCutoff", "Event"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncPolicyPage{}), []string{"Change", "Enumeration"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncPolicyDestroyed{}), []string{"EngineMessageID", "Projected", "Message", "Thread", "Labels", "ChangedEvent"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncPolicyResult{}), []string{"Advance", "Folders", "Upserts", "Tombstones", "Counts", "Events"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncPolicyError{}), []string{"Kind"})
	for _, testCase := range []struct {
		kind mail.MailboxSyncPolicyErrorKind
		want error
		text string
	}{
		{mail.MailboxSyncPolicyErrorInvalid, mail.ErrMailboxSyncPolicyInvalid, "mailbox sync policy invalid"},
		{mail.MailboxSyncPolicyErrorTooLarge, mail.ErrMailboxSyncPolicyTooLarge, "mailbox sync policy exceeds bounds"},
		{mail.MailboxSyncPolicyErrorDuplicate, mail.ErrMailboxSyncPolicyDuplicate, "mailbox sync policy duplicate"},
		{mail.MailboxSyncPolicyErrorAccountMismatch, mail.ErrMailboxSyncPolicyAccountMismatch, "mailbox sync policy account mismatch"},
	} {
		err := &mail.MailboxSyncPolicyError{Kind: testCase.kind}
		if !errors.Is(err, testCase.want) || err.Error() != testCase.text {
			t.Fatalf("policy error %#v did not preserve safe sentinel contract", err)
		}
	}
	var nilError *mail.MailboxSyncPolicyError
	if nilError.Error() != "mailbox sync policy error" || nilError.Unwrap() != nil {
		t.Fatal("nil policy error must be safe and content-free")
	}
}

func TestClassifyMailboxSyncPage_EmitsFirstInboundReceiptAndDeepCopies(t *testing.T) {
	input := validMailboxSyncInput(t)
	result, err := mail.ClassifyMailboxSyncPage(input)
	if err != nil {
		t.Fatalf("ClassifyMailboxSyncPage() error = %v", err)
	}
	if result.Advance.Change == nil || result.Advance.Enumeration != nil || result.Advance.SyncEpoch != input.SyncEpoch {
		t.Fatalf("advance = %#v, want exact change advance", result.Advance)
	}
	if len(result.Upserts) != 1 || len(result.Events) != 1 || result.Events[0].Subject() != "mail.received.v1" {
		t.Fatalf("result must emit one first-seen inbound receipt: %#v", result)
	}
	if result.Upserts[0].ReceivedEventID == nil || *result.Upserts[0].ReceivedEventID != input.Messages[0].ReceivedEventID {
		t.Fatalf("received marker = %#v, want first-seen marker", result.Upserts[0].ReceivedEventID)
	}
	input.Messages[0].EngineMessageID = "mutated-after-call"
	input.Page.Change.CreatedOrUpdated[0].MessageID = "mutated-after-call"
	input.Folders[0].Snapshot.FolderID = "mutated-after-call"
	if result.Upserts[0].Projection.MessageID != "message-04f" || result.Folders[0].Snapshot.FolderID != "inbox" {
		t.Fatalf("result retained caller-owned mutable state: %#v", result)
	}
}

func TestClassifyMailboxSyncPage_AppliesInboundReceiptCutoffToEnumerationHistory(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		receivedAt    func(time.Time) time.Time
		omitReceivedAt bool
		reservation   bool
		wantEvent     bool
		wantInvalid   bool
	}{
		{"historical before cutoff", func(cutoff time.Time) time.Time { return cutoff.Add(-time.Nanosecond) }, false, false, false, false},
		{"historical at cutoff", func(cutoff time.Time) time.Time { return cutoff }, false, false, false, false},
		{"after cutoff", func(cutoff time.Time) time.Time { return cutoff.Add(time.Nanosecond) }, false, true, true, false},
		{"after cutoff reservation is required", func(cutoff time.Time) time.Time { return cutoff.Add(time.Nanosecond) }, false, false, false, true},
		{"historical reservation is unused", func(cutoff time.Time) time.Time { return cutoff.Add(-time.Nanosecond) }, false, true, false, true},
		{"equal cutoff reservation is unused", func(cutoff time.Time) time.Time { return cutoff }, false, true, false, true},
		{"missing received at", nil, true, false, false, true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := mailboxSyncInputWithEnumeration(t, 1)
			if testCase.omitReceivedAt {
				input.Page.Enumeration.Messages[0].ReceivedAt = nil
			} else {
				receivedAt := testCase.receivedAt(input.EventCutoff)
				input.Page.Enumeration.Messages[0].ReceivedAt = &receivedAt
			}
			if !testCase.reservation {
				input.Messages[0].ReceivedEventID = uuid.Nil
			}

			result, err := mail.ClassifyMailboxSyncPage(input)
			if testCase.wantInvalid {
				if !errors.Is(err, mail.ErrMailboxSyncPolicyInvalid) || !reflect.DeepEqual(result, mail.MailboxSyncPolicyResult{}) {
					t.Fatalf("ClassifyMailboxSyncPage() = %#v, %v; want invalid without output", result, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ClassifyMailboxSyncPage() error = %v", err)
			}
			if len(result.Upserts) != 1 {
				t.Fatalf("upserts = %#v, want one projection", result.Upserts)
			}
			if testCase.wantEvent {
				if len(result.Events) != 1 || result.Events[0].Subject() != "mail.received.v1" || result.Upserts[0].ReceivedEventID == nil || *result.Upserts[0].ReceivedEventID != input.Messages[0].ReceivedEventID {
					t.Fatalf("after-cutoff result = %#v, want one received event and marker", result)
				}
				return
			}
			if len(result.Events) != 0 || result.Upserts[0].ReceivedEventID != nil {
				t.Fatalf("historical enumeration result = %#v, want no received event or marker", result)
			}
		})
	}
}

func TestClassifyMailboxSyncPage_DoesNotEmitReceivedForFirstSeenConfirmedOwner(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		origin mail.EngineProjectionOrigin
		proof  func(*mail.MailboxSyncPolicyProvenance, mail.MailboxSyncPolicyProvenanceProof)
	}{
		{
			name:   "submission",
			origin: mail.EngineProjectionOriginConfirmedOwnerSubmission,
			proof: func(provenance *mail.MailboxSyncPolicyProvenance, proof mail.MailboxSyncPolicyProvenanceProof) {
				provenance.ConfirmedSubmission = &proof
			},
		},
		{
			name:   "draft",
			origin: mail.EngineProjectionOriginConfirmedOwnerDraft,
			proof: func(provenance *mail.MailboxSyncPolicyProvenance, proof mail.MailboxSyncPolicyProvenanceProof) {
				provenance.ConfirmedDraft = &proof
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validMailboxSyncInput(t)
			input.Page.Change.CreatedOrUpdated[0].Origin = testCase.origin
			input.Messages[0].ReceivedEventID = uuid.Nil
			proof := mail.MailboxSyncPolicyProvenanceProof{
				Relation:        mailboxSyncID(input.Event.Mailbox.AccountID, "first-seen-owner-"+testCase.name),
				EngineMessageID: input.Page.Change.CreatedOrUpdated[0].MessageID,
			}
			testCase.proof(&input.Messages[0].Provenance, proof)

			result, err := mail.ClassifyMailboxSyncPage(input)
			if err != nil {
				t.Fatalf("ClassifyMailboxSyncPage() error = %v", err)
			}
			if len(result.Upserts) != 1 || result.Upserts[0].ReceivedEventID != nil || len(result.Events) != 0 {
				t.Fatalf("first-seen confirmed owner result = %#v, want no received marker or event", result)
			}
		})
	}
}

func TestClassifyMailboxSyncPage_UsesAllowlistedDeltaAndDestroyEvent(t *testing.T) {
	input := validMailboxSyncInput(t)
	input.Messages[0].Previous = &mail.MailboxSyncPolicyPreviousMessage{
		Thread: input.Messages[0].Thread.Thread,
		FolderIDs: []mail.MailboxEventID{input.Folders[0].Folder},
		Flags: mail.EngineMessageFlags{},
		Labels: []mail.MailboxSyncPolicyLabelState{},
	}
	input.Messages[0].ReceivedEventID = uuid.Nil
	input.Messages[0].ChangedEvent = mail.MailboxSyncPolicyEventReservation{ID: uuid.MustParse("11111111-1111-4111-8111-111111111111")}
	result, err := mail.ClassifyMailboxSyncPage(input)
	if err != nil {
		t.Fatalf("changed projection error = %v", err)
	}
	if len(result.Events) != 1 || result.Events[0].Subject() != "mail.message.changed.v1" {
		t.Fatalf("changed projection events = %#v", result.Events)
	}

	destroyed := validMailboxSyncInput(t)
	destroyed.Page.Change.CreatedOrUpdated = []mail.EngineMessageProjection{}
	destroyed.Page.Change.DestroyedIDs = []string{"message-04f"}
	destroyed.Page.Change.Bounds.Returned = 1
	destroyed.Messages = []mail.MailboxSyncPolicyMessage{}
	destroyed.Destroyed = []mail.MailboxSyncPolicyDestroyed{{
		EngineMessageID: "message-04f",
		Projected:       true,
		Message: destroyedMessageID(),
		Thread: destroyedThreadID(),
		Labels: []mail.MailboxSyncPolicyLabelState{},
		ChangedEvent: mail.MailboxSyncPolicyEventReservation{ID: uuid.MustParse("12111111-1111-4111-8111-111111111111")},
	}}
	result, err = mail.ClassifyMailboxSyncPage(destroyed)
	if err != nil {
		t.Fatalf("destroyed projection error = %v", err)
	}
	if len(result.Tombstones) != 1 || len(result.Events) != 1 || result.Events[0].Subject() != "mail.message.changed.v1" {
		t.Fatalf("destroyed result = %#v", result)
	}
}

func TestClassifyMailboxSyncPage_IgnoresUnknownDestroyedMessageAndAdvancesExactly(t *testing.T) {
	input := unknownDestroyedMailboxSyncInput(t)
	result, err := mail.ClassifyMailboxSyncPage(input)
	if err != nil {
		t.Fatalf("unknown destroyed message error = %v", err)
	}
	if result.Advance.Change == nil || result.Advance.Enumeration != nil || result.Advance.SyncEpoch != input.SyncEpoch || result.Advance.Change.FromCursor != input.Page.Change.FromCursor || result.Advance.Change.NextCursor != input.Page.Change.NextCursor {
		t.Fatalf("unknown destroyed advance = %#v, want exact change advance", result.Advance)
	}
	if len(result.Folders) != 0 || len(result.Upserts) != 0 || len(result.Tombstones) != 0 || len(result.Counts.Folders) != 0 || len(result.Counts.Threads) != 0 || len(result.Events) != 0 {
		t.Fatalf("unknown destroyed message must be a no-op: %#v", result)
	}
}

func TestClassifyMailboxSyncPage_RejectsHybridUnknownDestroyedState(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailboxSyncPolicyDestroyed, mail.MailboxSyncPolicyInput)
	}{
		{"message identity", func(destroyed *mail.MailboxSyncPolicyDestroyed, input mail.MailboxSyncPolicyInput) { destroyed.Message = mailboxSyncID(input.Event.Mailbox.AccountID, "unknown-message") }},
		{"thread identity", func(destroyed *mail.MailboxSyncPolicyDestroyed, input mail.MailboxSyncPolicyInput) { destroyed.Thread = mailboxSyncID(input.Event.Mailbox.AccountID, "unknown-thread") }},
		{"labels", func(destroyed *mail.MailboxSyncPolicyDestroyed, input mail.MailboxSyncPolicyInput) { destroyed.Labels = []mail.MailboxSyncPolicyLabelState{{Label: mailboxSyncID(input.Event.Mailbox.AccountID, "unknown-label")}} }},
		{"reservation", func(destroyed *mail.MailboxSyncPolicyDestroyed, _ mail.MailboxSyncPolicyInput) { destroyed.ChangedEvent = mail.MailboxSyncPolicyEventReservation{ID: mailboxSyncUUID("unknown-destroyed-reservation")} }},
		{"committed reservation", func(destroyed *mail.MailboxSyncPolicyDestroyed, _ mail.MailboxSyncPolicyInput) { destroyed.ChangedEvent = mail.MailboxSyncPolicyEventReservation{Committed: true} }},
		{"projected without message", func(destroyed *mail.MailboxSyncPolicyDestroyed, _ mail.MailboxSyncPolicyInput) { destroyed.Projected = true }},
		{"projected without thread", func(destroyed *mail.MailboxSyncPolicyDestroyed, input mail.MailboxSyncPolicyInput) { destroyed.Projected = true; destroyed.Message = mailboxSyncID(input.Event.Mailbox.AccountID, "projected-message") }},
		{"projected without reservation", func(destroyed *mail.MailboxSyncPolicyDestroyed, input mail.MailboxSyncPolicyInput) { destroyed.Projected = true; destroyed.Message = mailboxSyncID(input.Event.Mailbox.AccountID, "projected-message"); destroyed.Thread = mailboxSyncID(input.Event.Mailbox.AccountID, "projected-thread") }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := unknownDestroyedMailboxSyncInput(t)
			testCase.mutate(&input.Destroyed[0], input)
			result, err := mail.ClassifyMailboxSyncPage(input)
			if !errors.Is(err, mail.ErrMailboxSyncPolicyInvalid) || !reflect.DeepEqual(result, mail.MailboxSyncPolicyResult{}) {
				t.Fatalf("hybrid destroyed input = %#v, %v; want invalid without output", result, err)
			}
		})
	}
}

func TestClassifyMailboxSyncPage_RejectsUnionBoundsAndProvenanceViolations(t *testing.T) {
	for _, testCase := range []struct {
		name string
		mutate func(*mail.MailboxSyncPolicyInput)
		want error
	}{
		{"missing page", func(input *mail.MailboxSyncPolicyInput) { input.Page.Change = nil }, mail.ErrMailboxSyncPolicyInvalid},
		{"both pages", func(input *mail.MailboxSyncPolicyInput) { input.Page.Enumeration = &mail.EngineEnumerationPage{AccountID: "engine-account", Bounds: mail.EnginePageBounds{Requested: 1, Maximum: 100, Returned: 0}, Messages: []mail.EngineMessageProjection{}} }, mail.ErrMailboxSyncPolicyInvalid},
		{"returned mismatch", func(input *mail.MailboxSyncPolicyInput) { input.Page.Change.Bounds.Returned = 2 }, mail.ErrMailboxSyncPolicyInvalid},
		{"cross account message", func(input *mail.MailboxSyncPolicyInput) { input.Messages[0].Message.AccountID = uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc") }, mail.ErrMailboxSyncPolicyAccountMismatch},
		{"inbound bcc", func(input *mail.MailboxSyncPolicyInput) { input.Page.Change.CreatedOrUpdated[0].BCC = []mail.EngineAddress{{Email: "hidden@example.test"}} }, mail.ErrMailboxSyncPolicyInvalid},
		{"received reservation absent", func(input *mail.MailboxSyncPolicyInput) { input.Messages[0].ReceivedEventID = uuid.Nil }, mail.ErrMailboxSyncPolicyInvalid},
		{"duplicate page id", func(input *mail.MailboxSyncPolicyInput) { input.Page.Change.CreatedOrUpdated = append(input.Page.Change.CreatedOrUpdated, input.Page.Change.CreatedOrUpdated[0]); input.Page.Change.Bounds.Returned = 2 }, mail.ErrMailboxSyncPolicyDuplicate},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validMailboxSyncInput(t)
			testCase.mutate(&input)
			_, err := mail.ClassifyMailboxSyncPage(input)
			if !errors.Is(err, testCase.want) {
				t.Fatalf("ClassifyMailboxSyncPage() error = %v, want errors.Is(%v)", err, testCase.want)
			}
		})
	}
}

func TestClassifyMailboxSyncPage_EnforcesFrozenBoundsSlicesAndPageBudget(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		input func() mail.MailboxSyncPolicyInput
		want  error
	}{
		{"change 100", func() mail.MailboxSyncPolicyInput { return mailboxSyncInputWithLiveMessages(t, 100) }, nil},
		{"change 101", func() mail.MailboxSyncPolicyInput { return mailboxSyncInputWithLiveMessages(t, 101) }, mail.ErrMailboxSyncPolicyTooLarge},
		{"enumeration 100", func() mail.MailboxSyncPolicyInput { return mailboxSyncInputWithEnumeration(t, 100) }, nil},
		{"enumeration 101", func() mail.MailboxSyncPolicyInput { return mailboxSyncInputWithEnumeration(t, 101) }, mail.ErrMailboxSyncPolicyTooLarge},
		{"folders 100", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Folders = mailboxSyncFolders(input, 100)
			return input
		}, nil},
		{"folders 101", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Folders = mailboxSyncFolders(input, 101)
			return input
		}, mail.ErrMailboxSyncPolicyTooLarge},
		{"labels 100", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Labels = mailboxSyncLabels(input, 100)
			return input
		}, nil},
		{"labels 101", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Labels = mailboxSyncLabels(input, 101)
			return input
		}, mail.ErrMailboxSyncPolicyTooLarge},
		{"change projections nil", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Page.Change.CreatedOrUpdated = nil
			return input
		}, mail.ErrMailboxSyncPolicyInvalid},
		{"change destroyed nil", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Page.Change.DestroyedIDs = nil
			return input
		}, mail.ErrMailboxSyncPolicyInvalid},
		{"enumeration messages nil", func() mail.MailboxSyncPolicyInput {
			input := mailboxSyncInputWithEnumeration(t, 1)
			input.Page.Enumeration.Messages = nil
			return input
		}, mail.ErrMailboxSyncPolicyInvalid},
		{"projection recipients nil", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Page.Change.CreatedOrUpdated[0].To = nil
			return input
		}, mail.ErrMailboxSyncPolicyInvalid},
		{"cc 200", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Page.Change.CreatedOrUpdated[0].CC = mailboxSyncAddresses("cc", mail.MaxEngineAddressItems)
			mailboxSyncRefreshThread(t, &input, 0)
			return input
		}, nil},
		{"to 201", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Page.Change.CreatedOrUpdated[0].To = mailboxSyncAddresses("to", mail.MaxEngineAddressItems+1)
			return input
		}, mail.ErrMailboxSyncPolicyTooLarge},
		{"cc 201", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Page.Change.CreatedOrUpdated[0].CC = mailboxSyncAddresses("cc", mail.MaxEngineAddressItems+1)
			return input
		}, mail.ErrMailboxSyncPolicyTooLarge},
		{"bcc 201", func() mail.MailboxSyncPolicyInput {
			input := confirmedOwnerSyncPolicyInput(t)
			input.Page.Change.CreatedOrUpdated[0].BCC = mailboxSyncAddresses("bcc", mail.MaxEngineAddressItems+1)
			return input
		}, mail.ErrMailboxSyncPolicyTooLarge},
		{"reply-to 201", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Page.Change.CreatedOrUpdated[0].ReplyTo = mailboxSyncAddresses("reply", mail.MaxEngineAddressItems+1)
			return input
		}, mail.ErrMailboxSyncPolicyTooLarge},
		{"cross-list address duplicate", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Page.Change.CreatedOrUpdated[0].CC = []mail.EngineAddress{{Email: "owner@example.test"}}
			return input
		}, mail.ErrMailboxSyncPolicyDuplicate},
		{"aggregate page bytes", func() mail.MailboxSyncPolicyInput {
			input := mailboxSyncInputWithLiveMessages(t, 100)
			for projectionIndex := range input.Page.Change.CreatedOrUpdated {
				references := make([]string, mail.MaxEngineReferenceItems)
				for referenceIndex := range references {
					references[referenceIndex] = fmt.Sprintf("%s-%03d-%03d", strings.Repeat("r", 990), projectionIndex, referenceIndex)
				}
				input.Page.Change.CreatedOrUpdated[projectionIndex].References = references
			}
			return input
		}, mail.ErrMailboxSyncPolicyTooLarge},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := mail.ClassifyMailboxSyncPage(testCase.input())
			if testCase.want == nil {
				if err != nil {
					t.Fatalf("ClassifyMailboxSyncPage() error = %v", err)
				}
				return
			}
			if !errors.Is(err, testCase.want) {
				t.Fatalf("ClassifyMailboxSyncPage() error = %v, want errors.Is(%v)", err, testCase.want)
			}
		})
	}
}

func TestClassifyMailboxSyncPage_ClassifiesProjectionFolderIDCardinality(t *testing.T) {
	for _, pageKind := range []string{"change", "enumeration"} {
		for _, testCase := range []struct {
			name      string
			folderIDs func() []string
			want      error
		}{
			{"nil", func() []string { return nil }, mail.ErrMailboxSyncPolicyInvalid},
			{"empty", func() []string { return []string{} }, mail.ErrMailboxSyncPolicyInvalid},
			{"over max", func() []string {
				values := make([]string, mail.MaxEngineFolderItems+1)
				for index := range values {
					values[index] = fmt.Sprintf("folder-%03d", index)
				}
				return values
			}, mail.ErrMailboxSyncPolicyTooLarge},
		} {
			t.Run(pageKind+" "+testCase.name, func(t *testing.T) {
				input := validMailboxSyncInput(t)
				if pageKind == "enumeration" {
					input = mailboxSyncInputWithEnumeration(t, 1)
					input.Page.Enumeration.Messages[0].FolderIDs = testCase.folderIDs()
				} else {
					input.Page.Change.CreatedOrUpdated[0].FolderIDs = testCase.folderIDs()
				}

				result, err := mail.ClassifyMailboxSyncPage(input)
				if !errors.Is(err, testCase.want) {
					t.Fatalf("ClassifyMailboxSyncPage() error = %v, want errors.Is(%v)", err, testCase.want)
				}
				if !reflect.DeepEqual(result, mail.MailboxSyncPolicyResult{}) {
					t.Fatalf("ClassifyMailboxSyncPage() result = %#v, want zero result", result)
				}
			})
		}
	}
}

func TestClassifyMailboxSyncPage_RequiresAllAndOnlyPageStateAtExactLimit(t *testing.T) {
	accepted := mailboxSyncInputWithLiveMessages(t, mail.MaxEnginePageItems)
	result, err := mail.ClassifyMailboxSyncPage(accepted)
	if err != nil {
		t.Fatalf("exact page limit error = %v", err)
	}
	if len(result.Upserts) != mail.MaxEnginePageItems || len(result.Events) != mail.MaxEnginePageItems || len(result.Tombstones) != 0 {
		t.Fatalf("exact page limit result = %#v, want all-and-only 100 live state records", result)
	}

	missing := mailboxSyncInputWithLiveMessages(t, mail.MaxEnginePageItems)
	missing.Messages = missing.Messages[:mail.MaxEnginePageItems-1]
	if _, err := mail.ClassifyMailboxSyncPage(missing); !errors.Is(err, mail.ErrMailboxSyncPolicyInvalid) {
		t.Fatalf("missing exact-limit state error = %v, want invalid", err)
	}

	extra := mailboxSyncInputWithLiveMessages(t, mail.MaxEnginePageItems+1)
	extra.Page.Change.CreatedOrUpdated = extra.Page.Change.CreatedOrUpdated[:mail.MaxEnginePageItems]
	extra.Page.Change.Bounds.Returned = mail.MaxEnginePageItems
	if _, err := mail.ClassifyMailboxSyncPage(extra); !errors.Is(err, mail.ErrMailboxSyncPolicyTooLarge) {
		t.Fatalf("extra exact-limit state error = %v, want too_large", err)
	}

	overLimit := mailboxSyncInputWithLiveMessages(t, mail.MaxEnginePageItems+1)
	if _, err := mail.ClassifyMailboxSyncPage(overLimit); !errors.Is(err, mail.ErrMailboxSyncPolicyTooLarge) {
		t.Fatalf("101 all-and-only state records error = %v, want too_large", err)
	}
}

func TestClassifyMailboxSyncPage_AcceptsMixedPageAtLimitAndRejectsOverLimit(t *testing.T) {
	accepted := mailboxSyncInputWithLiveMessages(t, mail.MaxEnginePageItems-1)
	mailboxSyncAppendDestroyed(&accepted, "destroyed-at-limit")
	result, err := mail.ClassifyMailboxSyncPage(accepted)
	if err != nil {
		t.Fatalf("99 live plus 1 destroyed error = %v", err)
	}
	if len(result.Upserts) != mail.MaxEnginePageItems-1 || len(result.Tombstones) != 1 || len(result.Events) != mail.MaxEnginePageItems {
		t.Fatalf("99 live plus 1 destroyed result = %#v, want matching all-and-only page state", result)
	}

	overLimit := mailboxSyncInputWithLiveMessages(t, mail.MaxEnginePageItems)
	mailboxSyncAppendDestroyed(&overLimit, "destroyed-over-limit")
	if _, err := mail.ClassifyMailboxSyncPage(overLimit); !errors.Is(err, mail.ErrMailboxSyncPolicyTooLarge) {
		t.Fatalf("100 live plus 1 destroyed error = %v, want too_large", err)
	}
}

func TestClassifyMailboxSyncPage_CountsAllAffectedIDsExactlyAndSortsBytewise(t *testing.T) {
	input := mailboxSyncInputWithChangedMessages(t, 2)
	input.Folders = mailboxSyncFolders(input, 3)
	input.Page.Change.CreatedOrUpdated[0].FolderIDs = []string{input.Folders[1].Snapshot.FolderID}
	input.Page.Change.CreatedOrUpdated[1].FolderIDs = []string{input.Folders[2].Snapshot.FolderID}

	accountID := input.Event.Mailbox.AccountID
	currentFirst := mailboxSyncID(accountID, "count-current-first")
	previousFirst := mailboxSyncID(accountID, "count-previous-first")
	currentSecond := mailboxSyncID(accountID, "count-current-second")
	destroyedThread := mailboxSyncID(accountID, "count-destroyed-thread")
	input.Messages[0].Thread.Thread = currentFirst
	input.Messages[0].Previous.Thread = previousFirst
	input.Messages[0].Previous.FolderIDs = []mail.MailboxEventID{input.Folders[2].Folder}
	input.Messages[1].Thread.Thread = currentSecond
	input.Messages[1].Previous.Thread = currentFirst
	input.Messages[1].Previous.FolderIDs = []mail.MailboxEventID{input.Folders[0].Folder}
	mailboxSyncAppendDestroyed(&input, "count-destroyed")
	input.Destroyed[0].Thread = destroyedThread

	result, err := mail.ClassifyMailboxSyncPage(input)
	if err != nil {
		t.Fatalf("count classification error = %v", err)
	}
	requireMailboxSyncExactSortedIDs(t, result.Counts.Folders, mailboxSyncSortedIDs([]mail.MailboxEventID{
		input.Folders[0].Folder,
		input.Folders[1].Folder,
		input.Folders[2].Folder,
	}))
	requireMailboxSyncExactSortedIDs(t, result.Counts.Threads, mailboxSyncSortedIDs([]mail.MailboxEventID{
		currentFirst,
		previousFirst,
		currentSecond,
		destroyedThread,
	}))
}

func TestClassifyMailboxSyncPage_HandlesPerItemPriorStateAndReservationsAtPageLimit(t *testing.T) {
	accepted := mailboxSyncInputWithChangedMessages(t, mail.MaxEnginePageItems)
	result, err := mail.ClassifyMailboxSyncPage(accepted)
	if err != nil {
		t.Fatalf("100 unique prior-state reservations error = %v", err)
	}
	if len(result.Upserts) != mail.MaxEnginePageItems || len(result.Events) != mail.MaxEnginePageItems {
		t.Fatalf("100 unique prior-state reservations result = %#v", result)
	}
	for _, intent := range result.Upserts {
		if intent.ReceivedEventID != nil {
			t.Fatalf("changed intent retained unexpected received marker: %#v", intent)
		}
	}

	overLimit := mailboxSyncInputWithChangedMessages(t, mail.MaxEnginePageItems+1)
	if _, err := mail.ClassifyMailboxSyncPage(overLimit); !errors.Is(err, mail.ErrMailboxSyncPolicyTooLarge) {
		t.Fatalf("101 unique prior-state reservations error = %v, want too_large from combined page bound", err)
	}
}

func TestClassifyMailboxSyncPage_RejectsHugeNestedCollectionWithoutPanic(t *testing.T) {
	input := validMailboxSyncInput(t)
	input.Page.Change.CreatedOrUpdated[0].To = make([]mail.EngineAddress, mail.MaxEngineAddressItems*1024)

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("ClassifyMailboxSyncPage() panicked for oversized nested collection: %v", recovered)
		}
	}()

	_, err := mail.ClassifyMailboxSyncPage(input)
	if !errors.Is(err, mail.ErrMailboxSyncPolicyTooLarge) {
		t.Fatalf("ClassifyMailboxSyncPage() error = %v, want errors.Is(ErrMailboxSyncPolicyTooLarge)", err)
	}
}

func TestClassifyMailboxSyncPage_RejectsPageWideReservationDuplicates(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		input func() mail.MailboxSyncPolicyInput
	}{
		{"two receipts", func() mail.MailboxSyncPolicyInput {
			input := mailboxSyncInputWithLiveMessages(t, 2)
			input.Messages[1].ReceivedEventID = input.Messages[0].ReceivedEventID
			return input
		}},
		{"received and changed", func() mail.MailboxSyncPolicyInput {
			input := mailboxSyncInputWithLiveMessages(t, 2)
			input.Messages[1].Previous = &mail.MailboxSyncPolicyPreviousMessage{
				Thread:    input.Messages[1].Thread.Thread,
				FolderIDs: []mail.MailboxEventID{input.Folders[0].Folder},
				Flags:     mail.EngineMessageFlags{},
				Labels:    []mail.MailboxSyncPolicyLabelState{},
			}
			input.Messages[1].ReceivedEventID = uuid.Nil
			input.Messages[1].ChangedEvent = mail.MailboxSyncPolicyEventReservation{ID: input.Messages[0].ReceivedEventID, Committed: true}
			return input
		}},
		{"received and destroyed", func() mail.MailboxSyncPolicyInput {
			input := mailboxSyncInputWithLiveMessages(t, 1)
			input.Page.Change.DestroyedIDs = []string{"destroyed-04f"}
			input.Page.Change.Bounds.Returned = 2
			input.Destroyed = []mail.MailboxSyncPolicyDestroyed{{
				EngineMessageID: "destroyed-04f",
				Projected:       true,
				Message:         mailboxSyncID(input.Event.Mailbox.AccountID, "destroyed-message"),
				Thread:          input.Messages[0].Thread.Thread,
				Labels:          []mail.MailboxSyncPolicyLabelState{},
				ChangedEvent:    mail.MailboxSyncPolicyEventReservation{ID: input.Messages[0].ReceivedEventID, Committed: true},
			}}
			return input
		}},
		{"prior marker and current changed", func() mail.MailboxSyncPolicyInput {
			input := mailboxSyncInputWithChangedMessages(t, 2)
			marker := mailboxSyncUUID("prior-and-current-changed")
			input.Messages[0].Previous.ReceivedEventID = &marker
			input.Messages[1].ChangedEvent = mail.MailboxSyncPolicyEventReservation{ID: marker}
			return input
		}},
		{"prior marker and current received", func() mail.MailboxSyncPolicyInput {
			input := mailboxSyncInputWithLiveMessages(t, 2)
			marker := mailboxSyncUUID("prior-and-current-received")
			input.Messages[0].Previous = &mail.MailboxSyncPolicyPreviousMessage{
				Thread:          input.Messages[0].Thread.Thread,
				FolderIDs:       []mail.MailboxEventID{input.Folders[0].Folder},
				Flags:           input.Page.Change.CreatedOrUpdated[0].Flags,
				Labels:          []mail.MailboxSyncPolicyLabelState{},
				ReceivedEventID: &marker,
			}
			input.Messages[0].ReceivedEventID = uuid.Nil
			input.Messages[1].ReceivedEventID = marker
			return input
		}},
		{"prior marker and destroyed changed", func() mail.MailboxSyncPolicyInput {
			input := mailboxSyncInputWithLiveMessages(t, 1)
			marker := mailboxSyncUUID("prior-and-destroyed-changed")
			input.Messages[0].Previous = &mail.MailboxSyncPolicyPreviousMessage{
				Thread:          input.Messages[0].Thread.Thread,
				FolderIDs:       []mail.MailboxEventID{input.Folders[0].Folder},
				Flags:           input.Page.Change.CreatedOrUpdated[0].Flags,
				Labels:          []mail.MailboxSyncPolicyLabelState{},
				ReceivedEventID: &marker,
			}
			input.Messages[0].ReceivedEventID = uuid.Nil
			mailboxSyncAppendDestroyed(&input, "prior-and-destroyed")
			input.Destroyed[0].ChangedEvent = mail.MailboxSyncPolicyEventReservation{ID: marker}
			return input
		}},
		{"two prior markers", func() mail.MailboxSyncPolicyInput {
			input := mailboxSyncInputWithChangedMessages(t, 2)
			marker := mailboxSyncUUID("two-prior-markers")
			input.Messages[0].Previous.ReceivedEventID = &marker
			input.Messages[1].Previous.ReceivedEventID = &marker
			return input
		}},
		{"projection and destroyed page identifier", func() mail.MailboxSyncPolicyInput {
			input := validMailboxSyncInput(t)
			input.Page.Change.DestroyedIDs = []string{"message-04f"}
			input.Page.Change.Bounds.Returned = 2
			return input
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := mail.ClassifyMailboxSyncPage(testCase.input())
			if !errors.Is(err, mail.ErrMailboxSyncPolicyDuplicate) {
				t.Fatalf("ClassifyMailboxSyncPage() error = %v, want duplicate reservation error", err)
			}
		})
	}

	zeroPrior := mailboxSyncInputWithLiveMessages(t, 2)
	zero := uuid.Nil
	zeroPrior.Messages[0].Previous = &mail.MailboxSyncPolicyPreviousMessage{
		Thread:          zeroPrior.Messages[0].Thread.Thread,
		FolderIDs:       []mail.MailboxEventID{zeroPrior.Folders[0].Folder},
		Flags:           zeroPrior.Page.Change.CreatedOrUpdated[0].Flags,
		Labels:          []mail.MailboxSyncPolicyLabelState{},
		ReceivedEventID: &zero,
	}
	zeroPrior.Messages[0].ReceivedEventID = uuid.Nil
	zeroPrior.Messages[1].ChangedEvent = mail.MailboxSyncPolicyEventReservation{ID: zeroPrior.Messages[1].ReceivedEventID}
	if _, err := mail.ClassifyMailboxSyncPage(zeroPrior); !errors.Is(err, mail.ErrMailboxSyncPolicyInvalid) {
		t.Fatalf("zero prior marker must fail invalid before later current reservation duplicate: %v", err)
	}
}

func TestClassifyMailboxSyncPage_RequiresStrictStoredMessageIDs(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.EngineMessageProjection)
	}{
		{"unicode internet message id", func(projection *mail.EngineMessageProjection) { projection.InternetMessageID = "<\u0442\u0435\u0441\u0442@example.test>" }},
		{"unbracketed internet message id", func(projection *mail.EngineMessageProjection) { projection.InternetMessageID = "message@example.test" }},
		{"missing at-sign in reply-to", func(projection *mail.EngineMessageProjection) { projection.InReplyTo = "<message.example.test>" }},
		{"whitespace reference", func(projection *mail.EngineMessageProjection) { projection.References = []string{"<one@example.test> "} }},
		{"empty reference", func(projection *mail.EngineMessageProjection) { projection.References = []string{""} }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validMailboxSyncInput(t)
			testCase.mutate(&input.Page.Change.CreatedOrUpdated[0])
			result, err := mail.ClassifyMailboxSyncPage(input)
			if !errors.Is(err, mail.ErrMailboxSyncPolicyInvalid) || !reflect.DeepEqual(result, mail.MailboxSyncPolicyResult{}) {
				t.Fatalf("ClassifyMailboxSyncPage() = %#v, %v; want invalid with no retained output", result, err)
			}
		})
	}

	input := validMailboxSyncInput(t)
	projection := &input.Page.Change.CreatedOrUpdated[0]
	projection.InternetMessageID = "<Preserve.Case+Tag@Example.TEST>"
	projection.InReplyTo = "<Reply.Case@Example.TEST>"
	projection.References = []string{"<First.Ref@Example.TEST>", "<Second+Ref@Sub.Example.TEST>"}
	mailboxSyncRefreshThread(t, &input, 0)
	result, err := mail.ClassifyMailboxSyncPage(input)
	if err != nil {
		t.Fatalf("strict valid message identifiers error = %v", err)
	}
	if len(result.Upserts) != 1 || result.Upserts[0].Projection.InternetMessageID != "<Preserve.Case+Tag@Example.TEST>" || result.Upserts[0].Projection.InReplyTo != "<Reply.Case@Example.TEST>" || !reflect.DeepEqual(result.Upserts[0].Projection.References, []string{"<First.Ref@Example.TEST>", "<Second+Ref@Sub.Example.TEST>"}) {
		t.Fatalf("strict valid identifiers were not byte-preserved: %#v", result.Upserts)
	}
}

func TestClassifyMailboxSyncPage_AllowsOptionalAttachmentMetadataAndRejectsUnsafeValues(t *testing.T) {
	accepted := validMailboxSyncInput(t)
	accepted.Page.Change.CreatedOrUpdated[0].Attachments = []mail.EngineProjectionAttachment{{BlobID: "blob-04f", Filename: "", MediaType: "", Disposition: mail.EngineAttachmentDispositionAttachment}}
	result, err := mail.ClassifyMailboxSyncPage(accepted)
	if err != nil || len(result.Upserts) != 1 || !reflect.DeepEqual(result.Upserts[0].Projection.Attachments, accepted.Page.Change.CreatedOrUpdated[0].Attachments) {
		t.Fatalf("attachment with optional empty metadata = %#v, %v", result, err)
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*mail.EngineProjectionAttachment)
	}{
		{"oversized filename", func(attachment *mail.EngineProjectionAttachment) { attachment.Filename = strings.Repeat("f", mail.MaxEngineMIMEFieldBytes+1) }},
		{"unsafe filename", func(attachment *mail.EngineProjectionAttachment) { attachment.Filename = "unsafe\nname" }},
		{"oversized media type", func(attachment *mail.EngineProjectionAttachment) { attachment.MediaType = strings.Repeat("m", mail.MaxEngineMIMEFieldBytes+1) }},
		{"unsafe media type", func(attachment *mail.EngineProjectionAttachment) { attachment.MediaType = "text/\u0000plain" }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validMailboxSyncInput(t)
			input.Page.Change.CreatedOrUpdated[0].Attachments = []mail.EngineProjectionAttachment{{BlobID: "blob-04f", Filename: "file.txt", MediaType: "text/plain", Disposition: mail.EngineAttachmentDispositionAttachment}}
			testCase.mutate(&input.Page.Change.CreatedOrUpdated[0].Attachments[0])
			result, err := mail.ClassifyMailboxSyncPage(input)
			if !errors.Is(err, mail.ErrMailboxSyncPolicyInvalid) || !reflect.DeepEqual(result, mail.MailboxSyncPolicyResult{}) {
				t.Fatalf("ClassifyMailboxSyncPage() = %#v, %v; want invalid with no retained output", result, err)
			}
		})
	}
}

func TestClassifyMailboxSyncPage_EnforcesProvenancePreviousMarkersAndReplay(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailboxSyncPolicyInput)
		want   error
	}{
		{"zero proof relation", func(input *mail.MailboxSyncPolicyInput) { input.Messages[0].Provenance.ConfirmedSubmission.Relation.ID = uuid.Nil }, mail.ErrMailboxSyncPolicyInvalid},
		{"dual proof", func(input *mail.MailboxSyncPolicyInput) {
			proof := *input.Messages[0].Provenance.ConfirmedSubmission
			input.Messages[0].Provenance.ConfirmedDraft = &proof
		}, mail.ErrMailboxSyncPolicyInvalid},
		{"cross account proof", func(input *mail.MailboxSyncPolicyInput) { input.Messages[0].Provenance.ConfirmedSubmission.Relation.AccountID = mailboxSyncUUID("foreign-account") }, mail.ErrMailboxSyncPolicyAccountMismatch},
		{"proof engine identifier mismatch", func(input *mail.MailboxSyncPolicyInput) { input.Messages[0].Provenance.ConfirmedSubmission.EngineMessageID = "other-message" }, mail.ErrMailboxSyncPolicyInvalid},
		{"proof origin mismatch", func(input *mail.MailboxSyncPolicyInput) { input.Page.Change.CreatedOrUpdated[0].Origin = mail.EngineProjectionOriginConfirmedOwnerDraft }, mail.ErrMailboxSyncPolicyInvalid},
		{"owner origin without proof", func(input *mail.MailboxSyncPolicyInput) { input.Messages[0].Provenance.ConfirmedSubmission = nil }, mail.ErrMailboxSyncPolicyInvalid},
		{"previous zero received marker", func(input *mail.MailboxSyncPolicyInput) {
			zero := uuid.Nil
			input.Messages[0].Previous.ReceivedEventID = &zero
		}, mail.ErrMailboxSyncPolicyInvalid},
		{"owner bcc remains valid only with proof", func(input *mail.MailboxSyncPolicyInput) { input.Page.Change.CreatedOrUpdated[0].BCC = []mail.EngineAddress{{Email: "owner-bcc@example.test"}} }, nil},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := confirmedOwnerSyncPolicyInput(t)
			testCase.mutate(&input)
			_, err := mail.ClassifyMailboxSyncPage(input)
			if testCase.want == nil {
				if err != nil {
					t.Fatalf("ClassifyMailboxSyncPage() error = %v", err)
				}
				return
			}
			if !errors.Is(err, testCase.want) {
				t.Fatalf("ClassifyMailboxSyncPage() error = %v, want errors.Is(%v)", err, testCase.want)
			}
		})
	}

	replay := validMailboxSyncInput(t)
	marker := mailboxSyncUUID("received-replay-marker")
	replay.Messages[0].Previous = &mail.MailboxSyncPolicyPreviousMessage{
		Thread:          replay.Messages[0].Thread.Thread,
		FolderIDs:       []mail.MailboxEventID{replay.Folders[0].Folder},
		Flags:           replay.Page.Change.CreatedOrUpdated[0].Flags,
		Labels:          []mail.MailboxSyncPolicyLabelState{},
		ReceivedEventID: &marker,
	}
	replay.Messages[0].ReceivedEventID = uuid.Nil
	result, err := mail.ClassifyMailboxSyncPage(replay)
	if err != nil {
		t.Fatalf("replay ClassifyMailboxSyncPage() error = %v", err)
	}
	if len(result.Events) != 0 || len(result.Upserts) != 1 || result.Upserts[0].ReceivedEventID == nil || *result.Upserts[0].ReceivedEventID != marker {
		t.Fatalf("replay result = %#v, want retained receipt marker without event", result)
	}

	committed := confirmedOwnerSyncPolicyInput(t)
	committed.Messages[0].Previous.Flags.Seen = false
	committed.Messages[0].ChangedEvent = mail.MailboxSyncPolicyEventReservation{ID: mailboxSyncUUID("committed-change"), Committed: true}
	result, err = mail.ClassifyMailboxSyncPage(committed)
	if err != nil {
		t.Fatalf("committed change ClassifyMailboxSyncPage() error = %v", err)
	}
	if len(result.Events) != 0 {
		t.Fatalf("committed reservation emitted replay event: %#v", result.Events)
	}

	ownerBCC := confirmedOwnerSyncPolicyInput(t)
	ownerBCC.Page.Change.CreatedOrUpdated[0].BCC = []mail.EngineAddress{{Email: "owner-bcc@example.test"}}
	result, err = mail.ClassifyMailboxSyncPage(ownerBCC)
	if err != nil || len(result.Upserts) != 1 || len(result.Upserts[0].Projection.BCC) != 0 {
		t.Fatalf("confirmed BCC leaked from safe projection: result=%#v err=%v", result, err)
	}
}

func TestClassifyMailboxSyncPage_ReplacesEngineLabelsAndPreservesManualLabels(t *testing.T) {
	input := confirmedOwnerSyncPolicyInput(t)
	accountID := input.Event.Mailbox.AccountID
	manualLabel := mail.MailboxEventID{AccountID: accountID, ID: uuid.MustParse("01010101-0101-4101-8101-010101010101")}
	engineLabel := mail.MailboxEventID{AccountID: accountID, ID: uuid.MustParse("02020202-0202-4202-8202-020202020202")}
	input.Labels = []mail.MailboxSyncPolicyLabelDefinition{
		{Label: manualLabel, EngineKeyword: "known-manual"},
		{Label: engineLabel, EngineKeyword: "known-engine"},
	}
	input.Messages[0].Previous.Labels = []mail.MailboxSyncPolicyLabelState{
		{Label: manualLabel, Manual: true, EngineRule: true},
		{Label: engineLabel, EngineRule: true},
	}
	input.Page.Change.CreatedOrUpdated[0].ReservedKeywords = []mail.EngineKeywordMarker{"known-engine", "unknown-keyword"}
	input.Messages[0].ChangedEvent = mail.MailboxSyncPolicyEventReservation{ID: mailboxSyncUUID("label-change")}

	result, err := mail.ClassifyMailboxSyncPage(input)
	if err != nil {
		t.Fatalf("ClassifyMailboxSyncPage() error = %v", err)
	}
	if len(result.Upserts) != 1 || len(result.Upserts[0].Projection.ReservedKeywords) != 0 {
		t.Fatalf("raw engine keywords leaked into result: %#v", result.Upserts)
	}
	if !reflect.DeepEqual(result.Upserts[0].Labels, []mail.MailboxSyncPolicyLabelState{
		{Label: manualLabel, Manual: true},
		{Label: engineLabel, EngineRule: true},
	}) {
		t.Fatalf("labels = %#v, want manual preserved and current known engine rule only", result.Upserts[0].Labels)
	}
	if len(result.Events) != 1 || result.Events[0].Subject() != "mail.message.changed.v1" {
		t.Fatalf("label replacement events = %#v", result.Events)
	}
}

func TestClassifyMailboxSyncPage_IsDeterministicAndDoesNotAliasCallerState(t *testing.T) {
	input := mailboxSyncInputWithLiveMessages(t, 2)
	input.Folders = mailboxSyncFolders(input, 2)
	input.Page.Change.CreatedOrUpdated[1].FolderIDs = []string{input.Folders[1].Snapshot.FolderID}

	first, err := mail.ClassifyMailboxSyncPage(input)
	if err != nil {
		t.Fatalf("first ClassifyMailboxSyncPage() error = %v", err)
	}
	shuffled := mailboxSyncInputWithLiveMessages(t, 2)
	shuffled.Folders = mailboxSyncFolders(shuffled, 2)
	shuffled.Page.Change.CreatedOrUpdated[1].FolderIDs = []string{shuffled.Folders[1].Snapshot.FolderID}
	shuffled.Folders[0], shuffled.Folders[1] = shuffled.Folders[1], shuffled.Folders[0]
	shuffled.Messages[0], shuffled.Messages[1] = shuffled.Messages[1], shuffled.Messages[0]
	shuffled.Page.Change.CreatedOrUpdated[0], shuffled.Page.Change.CreatedOrUpdated[1] = shuffled.Page.Change.CreatedOrUpdated[1], shuffled.Page.Change.CreatedOrUpdated[0]
	second, err := mail.ClassifyMailboxSyncPage(shuffled)
	if err != nil {
		t.Fatalf("shuffled ClassifyMailboxSyncPage() error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("shuffled input changed deterministic result:\nfirst=%#v\nsecond=%#v", first, second)
	}
	if first.Upserts[0].Projection.MessageID >= first.Upserts[1].Projection.MessageID || first.Folders[0].Folder.ID == first.Folders[1].Folder.ID || bytes.Compare(first.Folders[0].Folder.ID[:], first.Folders[1].Folder.ID[:]) >= 0 {
		t.Fatalf("result order is not raw-byte deterministic: %#v", first)
	}

	eventJSON, err := first.Events[0].MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	input.Page.Change.CreatedOrUpdated[0].To[0].Email = "mutated@example.test"
	input.Page.Change.CreatedOrUpdated[0].FolderIDs[0] = "mutated-folder"
	input.Folders[0].Snapshot.Name = "mutated-folder"
	inboxIntent := mailboxSyncFolderIntent(t, first, "inbox")
	if first.Upserts[0].Projection.To[0].Email != "owner@example.test" || first.Upserts[0].Projection.FolderIDs[0] == "mutated-folder" || inboxIntent.Snapshot.Name == "mutated-folder" {
		t.Fatalf("result aliases caller-owned projection or folder state: %#v", first)
	}
	currentEventJSON, err := first.Events[0].MarshalJSON()
	if err != nil || !bytes.Equal(eventJSON, currentEventJSON) {
		t.Fatalf("result event mutated through caller state: err=%v before=%s after=%s", err, eventJSON, currentEventJSON)
	}
}

func TestClassifyMailboxSyncPage_RejectsUnsafeEventContextBeforeEventlessPage(t *testing.T) {
	eventless := validMailboxSyncInput(t)
	eventless.Page.Change.CreatedOrUpdated = []mail.EngineMessageProjection{}
	eventless.Page.Change.DestroyedIDs = []string{}
	eventless.Page.Change.Bounds.Returned = 0
	eventless.Messages = []mail.MailboxSyncPolicyMessage{}
	eventless.Destroyed = []mail.MailboxSyncPolicyDestroyed{}
	result, err := mail.ClassifyMailboxSyncPage(eventless)
	if err != nil || len(result.Events) != 0 || len(result.Upserts) != 0 || len(result.Tombstones) != 0 {
		t.Fatalf("eventless page = %#v, %v", result, err)
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailboxSyncPolicyInput)
	}{
		{"trim-unstable correlation", func(input *mail.MailboxSyncPolicyInput) { input.Event.CorrelationID = " corr-mail-04f " }},
		{"overlong correlation", func(input *mail.MailboxSyncPolicyInput) { input.Event.CorrelationID = strings.Repeat("c", 129) }},
		{"huge valid correlation", func(input *mail.MailboxSyncPolicyInput) { input.Event.CorrelationID = strings.Repeat("c", 1024*1024) }},
		{"zero organization", func(input *mail.MailboxSyncPolicyInput) { input.Event.OrganizationID = uuid.Nil }},
		{"zero actor", func(input *mail.MailboxSyncPolicyInput) { actor := uuid.Nil; input.Event.ActorID = &actor }},
		{"nonzero event offset", func(input *mail.MailboxSyncPolicyInput) { input.Event.OccurredAt = time.Date(2030, time.May, 6, 10, 8, 10, 0, time.FixedZone("plus-one", 3600)) }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("ClassifyMailboxSyncPage() panicked for unsafe event context: %v", recovered)
				}
			}()
			input := eventless
			testCase.mutate(&input)
			_, err := mail.ClassifyMailboxSyncPage(input)
			if !errors.Is(err, mail.ErrMailboxSyncPolicyInvalid) {
				t.Fatalf("ClassifyMailboxSyncPage() error = %v, want invalid event context", err)
			}
		})
	}

	utcEquivalent := eventless
	utcEquivalent.Event.OccurredAt = time.Date(2030, time.May, 6, 7, 8, 10, 0, time.FixedZone("zero", 0))
	if _, err := mail.ClassifyMailboxSyncPage(utcEquivalent); err != nil {
		t.Fatalf("zero-offset event context must match envelope acceptance: %v", err)
	}
}

func TestClassifyMailboxSyncPage_ContainsErrorsAndNeverPanicsOnMalformedInput(t *testing.T) {
	secret := validMailboxSyncInput(t)
	secret.Event.CorrelationID = " secret-input-value "
	_, err := mail.ClassifyMailboxSyncPage(secret)
	if !errors.Is(err, mail.ErrMailboxSyncPolicyInvalid) || strings.Contains(err.Error(), "secret-input-value") {
		t.Fatalf("policy error leaked invalid input: %v", err)
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailboxSyncPolicyInput)
	}{
		{"nil page slices", func(input *mail.MailboxSyncPolicyInput) {
			input.Page.Change.CreatedOrUpdated = nil
			input.Page.Change.DestroyedIDs = nil
		}},
		{"oversized recipients", func(input *mail.MailboxSyncPolicyInput) {
			input.Page.Change.CreatedOrUpdated[0].To = make([]mail.EngineAddress, mail.MaxEngineAddressItems+1)
		}},
		{"zero previous marker", func(input *mail.MailboxSyncPolicyInput) {
			zero := uuid.Nil
			input.Messages[0].Previous = &mail.MailboxSyncPolicyPreviousMessage{Thread: input.Messages[0].Thread.Thread, FolderIDs: []mail.MailboxEventID{input.Folders[0].Folder}, Labels: []mail.MailboxSyncPolicyLabelState{}, ReceivedEventID: &zero}
			input.Messages[0].ReceivedEventID = uuid.Nil
		}},
		{"foreign event mailbox", func(input *mail.MailboxSyncPolicyInput) { input.Event.Mailbox.AccountID = mailboxSyncUUID("foreign-event-account") }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("ClassifyMailboxSyncPage() panicked: %v", recovered)
				}
			}()
			input := validMailboxSyncInput(t)
			testCase.mutate(&input)
			_, _ = mail.ClassifyMailboxSyncPage(input)
		})
	}
}

func validMailboxSyncInput(t *testing.T) mail.MailboxSyncPolicyInput {
	t.Helper()
	account := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	folder := mail.MailboxEventID{AccountID: account, ID: uuid.MustParse("99999999-9999-4999-8999-999999999999")}
	message := destroyedMessageID()
	thread := destroyedThreadID()
	receivedAt := time.Date(2030, time.May, 6, 7, 8, 9, 0, time.UTC)
	projection := mail.EngineMessageProjection{
		AccountID: "engine-account", MessageID: "message-04f", Version: "v1",
		From: mail.EngineAddress{Email: "sender@example.test"}, To: []mail.EngineAddress{{Email: "owner@example.test"}}, CC: []mail.EngineAddress{}, BCC: []mail.EngineAddress{}, ReplyTo: []mail.EngineAddress{}, References: []string{}, ReservedKeywords: []mail.EngineKeywordMarker{}, Attachments: []mail.EngineProjectionAttachment{},
		Subject: "Sync policy", ReceivedAt: &receivedAt, Priority: mail.EnginePriorityNormal,
		FolderIDs: []string{"inbox"}, Flags: mail.EngineMessageFlags{Seen: true}, Origin: mail.EngineProjectionOriginInboundOrUnknown, AuthenticationClass: mail.MailAuthenticationClassUntrusted,
	}
	identity, err := mail.DeriveThreadIdentity(mail.ThreadingInput{Subject: projection.Subject, From: projection.From, To: projection.To, Ancestors: []mail.ThreadAncestor{}, AncestorLookupComplete: true})
	if err != nil {
		t.Fatalf("fixture identity error = %v", err)
	}
	return mail.MailboxSyncPolicyInput{
		Page: mail.MailboxSyncPolicyPage{Change: &mail.EngineChangePage{AccountID: "engine-account", FromCursor: "cursor-a", NextCursor: "cursor-b", Bounds: mail.EnginePageBounds{Requested: 1, Maximum: 100, Returned: 1}, CreatedOrUpdated: []mail.EngineMessageProjection{projection}, DestroyedIDs: []string{}}},
		Folders: []mail.MailboxSyncPolicyFolder{{Snapshot: mail.EngineFolder{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}, Folder: folder}},
		Messages: []mail.MailboxSyncPolicyMessage{{EngineMessageID: "message-04f", Message: message, Thread: mail.MailboxSyncPolicyThreadResolution{Thread: thread, KeyVersion: identity.KeyVersion, Key: identity.Key, Ancestors: []mail.ThreadAncestor{}, AncestorLookupComplete: true}, ReceivedEventID: uuid.MustParse("10111111-1111-4111-8111-111111111111")}},
		Destroyed: []mail.MailboxSyncPolicyDestroyed{}, Labels: []mail.MailboxSyncPolicyLabelDefinition{}, SyncEpoch: 7,
		EventCutoff: time.Date(2030, time.May, 6, 7, 8, 8, 0, time.UTC),
		Event: mail.MailboxSyncPolicyEventContext{Mailbox: mail.MailboxEventID{AccountID: account, ID: uuid.MustParse("55555555-5555-4555-8555-555555555555")}, OrganizationID: uuid.MustParse("22222222-2222-4222-8222-222222222222"), OccurredAt: time.Date(2030, time.May, 6, 7, 8, 10, 0, time.UTC), CorrelationID: "corr-mail-04f"},
	}
}

func mailboxSyncInputWithLiveMessages(t *testing.T, count int) mail.MailboxSyncPolicyInput {
	t.Helper()
	input := validMailboxSyncInput(t)
	baseProjection := input.Page.Change.CreatedOrUpdated[0]
	baseMessage := input.Messages[0]
	projections := make([]mail.EngineMessageProjection, count)
	messages := make([]mail.MailboxSyncPolicyMessage, count)
	for index := range projections {
		projection := baseProjection
		projection.MessageID = fmt.Sprintf("message-%03d", index)
		projection.Version = fmt.Sprintf("version-%03d", index)
		message := baseMessage
		message.EngineMessageID = projection.MessageID
		message.Message = mailboxSyncID(input.Event.Mailbox.AccountID, fmt.Sprintf("message-id-%03d", index))
		message.ReceivedEventID = mailboxSyncUUID(fmt.Sprintf("received-event-%03d", index))
		projections[index] = projection
		messages[index] = message
	}
	input.Page.Change.CreatedOrUpdated = projections
	input.Page.Change.DestroyedIDs = []string{}
	input.Page.Change.Bounds.Returned = count
	input.Messages = messages
	input.Destroyed = []mail.MailboxSyncPolicyDestroyed{}
	return input
}

func mailboxSyncInputWithChangedMessages(t *testing.T, count int) mail.MailboxSyncPolicyInput {
	t.Helper()
	input := mailboxSyncInputWithLiveMessages(t, count)
	for index := range input.Messages {
		input.Messages[index].Previous = &mail.MailboxSyncPolicyPreviousMessage{
			Thread:    input.Messages[index].Thread.Thread,
			FolderIDs: []mail.MailboxEventID{input.Folders[0].Folder},
			Flags:     mail.EngineMessageFlags{Seen: false},
			Labels:    []mail.MailboxSyncPolicyLabelState{},
		}
		input.Messages[index].ReceivedEventID = uuid.Nil
		input.Messages[index].ChangedEvent = mail.MailboxSyncPolicyEventReservation{ID: mailboxSyncUUID(fmt.Sprintf("changed-event-%03d", index))}
	}
	return input
}

func mailboxSyncAppendDestroyed(input *mail.MailboxSyncPolicyInput, name string) {
	engineMessageID := "destroyed-" + name
	input.Page.Change.DestroyedIDs = append(input.Page.Change.DestroyedIDs, engineMessageID)
	input.Page.Change.Bounds.Returned = len(input.Page.Change.CreatedOrUpdated) + len(input.Page.Change.DestroyedIDs)
	input.Destroyed = append(input.Destroyed, mail.MailboxSyncPolicyDestroyed{
		EngineMessageID: engineMessageID,
		Projected:       true,
		Message:         mailboxSyncID(input.Event.Mailbox.AccountID, engineMessageID+"-message"),
		Thread:          mailboxSyncID(input.Event.Mailbox.AccountID, engineMessageID+"-thread"),
		Labels:          []mail.MailboxSyncPolicyLabelState{},
		ChangedEvent:    mail.MailboxSyncPolicyEventReservation{ID: mailboxSyncUUID(engineMessageID + "-event")},
	})
}

func unknownDestroyedMailboxSyncInput(t *testing.T) mail.MailboxSyncPolicyInput {
	t.Helper()
	input := validMailboxSyncInput(t)
	input.Page.Change.CreatedOrUpdated = []mail.EngineMessageProjection{}
	input.Page.Change.DestroyedIDs = []string{"unknown-destroyed-04f"}
	input.Page.Change.Bounds.Returned = 1
	input.Folders = []mail.MailboxSyncPolicyFolder{}
	input.Messages = []mail.MailboxSyncPolicyMessage{}
	input.Destroyed = []mail.MailboxSyncPolicyDestroyed{{EngineMessageID: "unknown-destroyed-04f"}}
	return input
}

func mailboxSyncInputWithEnumeration(t *testing.T, count int) mail.MailboxSyncPolicyInput {
	t.Helper()
	input := mailboxSyncInputWithLiveMessages(t, count)
	input.Page.Enumeration = &mail.EngineEnumerationPage{
		AccountID: input.Page.Change.AccountID, BaseCursor: "cursor-a", Checkpoint: "checkpoint-a", NextCheckpoint: "", Complete: true,
		Bounds: mail.EnginePageBounds{Requested: 1, Maximum: mail.MaxEnginePageItems, Returned: count}, Messages: input.Page.Change.CreatedOrUpdated,
	}
	input.Page.Change = nil
	return input
}

func confirmedOwnerSyncPolicyInput(t *testing.T) mail.MailboxSyncPolicyInput {
	t.Helper()
	input := validMailboxSyncInput(t)
	projection := &input.Page.Change.CreatedOrUpdated[0]
	message := &input.Messages[0]
	projection.Origin = mail.EngineProjectionOriginConfirmedOwnerSubmission
	message.ReceivedEventID = uuid.Nil
	message.Previous = &mail.MailboxSyncPolicyPreviousMessage{
		Thread:    message.Thread.Thread,
		FolderIDs: []mail.MailboxEventID{input.Folders[0].Folder},
		Flags:     projection.Flags,
		Labels:    []mail.MailboxSyncPolicyLabelState{},
	}
	message.Provenance.ConfirmedSubmission = &mail.MailboxSyncPolicyProvenanceProof{
		Relation:        mailboxSyncID(input.Event.Mailbox.AccountID, "confirmed-submission"),
		EngineMessageID: projection.MessageID,
	}
	return input
}

func mailboxSyncFolders(input mail.MailboxSyncPolicyInput, count int) []mail.MailboxSyncPolicyFolder {
	folders := make([]mail.MailboxSyncPolicyFolder, count)
	if count == 0 {
		return folders
	}
	folders[0] = input.Folders[0]
	for index := 1; index < count; index++ {
		folderID := fmt.Sprintf("folder-%03d", index)
		folders[index] = mail.MailboxSyncPolicyFolder{
			Snapshot: mail.EngineFolder{AccountID: input.Page.Change.AccountID, FolderID: folderID, Name: fmt.Sprintf("Folder %03d", index), Role: mail.EngineFolderRoleCustom},
			Folder:   mailboxSyncID(input.Event.Mailbox.AccountID, folderID),
		}
	}
	return folders
}

func mailboxSyncLabels(input mail.MailboxSyncPolicyInput, count int) []mail.MailboxSyncPolicyLabelDefinition {
	labels := make([]mail.MailboxSyncPolicyLabelDefinition, count)
	for index := range labels {
		name := fmt.Sprintf("label-%03d", index)
		labels[index] = mail.MailboxSyncPolicyLabelDefinition{Label: mailboxSyncID(input.Event.Mailbox.AccountID, name), EngineKeyword: mail.EngineKeywordMarker(name)}
	}
	return labels
}

func mailboxSyncAddresses(prefix string, count int) []mail.EngineAddress {
	addresses := make([]mail.EngineAddress, count)
	for index := range addresses {
		addresses[index] = mail.EngineAddress{Email: fmt.Sprintf("%s-%03d@example.test", prefix, index)}
	}
	return addresses
}

func mailboxSyncRefreshThread(t *testing.T, input *mail.MailboxSyncPolicyInput, index int) {
	t.Helper()
	projection := input.Page.Change.CreatedOrUpdated[index]
	resolution := input.Messages[index].Thread
	identity, err := mail.DeriveThreadIdentity(mail.ThreadingInput{
		References: projection.References, InReplyTo: projection.InReplyTo, Subject: projection.Subject,
		From: projection.From, To: projection.To, CC: projection.CC,
		Ancestors: resolution.Ancestors, AncestorLookupComplete: resolution.AncestorLookupComplete,
	})
	if err != nil {
		t.Fatalf("DeriveThreadIdentity() fixture error = %v", err)
	}
	input.Messages[index].Thread.KeyVersion = identity.KeyVersion
	input.Messages[index].Thread.Key = identity.Key
}

func mailboxSyncFolderIntent(t *testing.T, result mail.MailboxSyncPolicyResult, engineFolderID string) mail.MailboxSyncPolicyFolderIntent {
	t.Helper()
	for _, intent := range result.Folders {
		if intent.Snapshot.FolderID == engineFolderID {
			return intent
		}
	}
	t.Fatalf("folder intent %q was not found in %#v", engineFolderID, result.Folders)
	return mail.MailboxSyncPolicyFolderIntent{}
}

func mailboxSyncID(accountID uuid.UUID, name string) mail.MailboxEventID {
	return mail.MailboxEventID{AccountID: accountID, ID: mailboxSyncUUID(name)}
}

func mailboxSyncUUID(name string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("mailbox-sync-policy-test:"+name))
}

func mailboxSyncSortedIDs(values []mail.MailboxEventID) []mail.MailboxEventID {
	result := append([]mail.MailboxEventID(nil), values...)
	for left := range result {
		for right := left + 1; right < len(result); right++ {
			if bytes.Compare(result[right].ID[:], result[left].ID[:]) < 0 {
				result[left], result[right] = result[right], result[left]
			}
		}
	}
	return result
}

func requireMailboxSyncExactSortedIDs(t *testing.T, got, want []mail.MailboxEventID) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("IDs = %#v, want %#v", got, want)
	}
	for index := 1; index < len(got); index++ {
		if bytes.Compare(got[index-1].ID[:], got[index].ID[:]) >= 0 {
			t.Fatalf("IDs must be deduplicated and bytewise sorted: %#v", got)
		}
	}
}

func destroyedMessageID() mail.MailboxEventID {
	return mail.MailboxEventID{AccountID: uuid.MustParse("44444444-4444-4444-8444-444444444444"), ID: uuid.MustParse("66666666-6666-4666-8666-666666666666")}
}

func destroyedThreadID() mail.MailboxEventID {
	return mail.MailboxEventID{AccountID: uuid.MustParse("44444444-4444-4444-8444-444444444444"), ID: uuid.MustParse("77777777-7777-4777-8777-777777777777")}
}

func requireMailboxSyncFields(t *testing.T, valueType reflect.Type, want []string) {
	t.Helper()
	if valueType.NumField() != len(want) {
		t.Fatalf("%s field count = %d, want %d", valueType.Name(), valueType.NumField(), len(want))
	}
	for index, name := range want {
		if field := valueType.Field(index); field.Name != name || !field.IsExported() {
			t.Fatalf("%s field %d = %#v, want exported %q", valueType.Name(), index, field, name)
		}
	}
}
