package mail_test

import (
	"os"
	"strings"
	"testing"
)

func TestMailboxCalendarDeliveryDraftCreatorSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_calendar_delivery_draft_creator.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	for _, term := range []string{
		"type mailboxCalendarDeliveryDraftCreator struct",
		"func newMailboxCalendarDeliveryDraftCreator",
		"mailboxCalendarDeliveryDraftLocator",
		"mailboxDraftCalendarCreateRepository",
		"mailboxDraftCalendarCreateEngine",
		"mailboxCalendarDeliveryDraftReader",
		"func (creator *mailboxCalendarDeliveryDraftCreator) CreateMailboxDraft",
		"mailboxCalendarDeliveryDraftInputSealed(input)",
		"input.calendar.valid()",
		"len(input.To) != 1",
		"len(input.CC) != 0",
		"len(input.BCC) != 0",
		"len(input.Attachments) != 0",
		"input.ReplyToMessageID != nil",
		"input.ForwardMessageID != nil",
		"calendarMailDeliveryDraftClientID(input.calendar.deliveryID())",
		"calendarMailDeliveryDraftKey(input.calendar.deliveryID())",
		"owner.PrimaryAddress()",
		"ResolveMailboxDraftIdentityProof",
		"identity.from.Email != primary.Email",
		"mailboxDraftCreateRequestHashWithCalendar",
		"mailboxDraftCreateMIMEBindingDigest",
		"LookupMailboxDraftCreatePreHashLocator",
		"LookupMailboxDraftCreateReplay",
		"MailboxDraftServiceErrorIdempotencyConflict",
		"BeginEngineCommand",
		"ReserveMailboxDraftCalendarCreate",
		"uploadCalendarDraftCreateMIME",
		"MarkMailboxDraftCalendarCreateUnknown",
		"mailboxDraftCalendarCreateUnknownStageUpload",
		"mailboxDraftCalendarCreateUnknownStageAdvance",
		"mailboxDraftCalendarCreateUnknownStageImport",
		"AdvanceMailboxDraftCalendarCreateToImport",
		"receivedAt: advanced.Draft.UpdatedAt",
		"importCalendarDraftCreateMIME",
		"errors.Is(err, errMailboxDraftCalendarCreatePersistedBlobNotFound)",
		"state.Attempt < 3",
		"RestartMailboxDraftCalendarCreateUploadAfterBlobNotFound",
		"NextMutationID",
		"NextBindingDigest",
		"FinishMailboxDraftCalendarCreateApplied",
		"creator.get(ctx, actorID, prepared.canonical.DraftClientID, finished.ID)",
	} {
		if !strings.Contains(source, term) {
			t.Fatalf("missing creator term %q", term)
		}
	}
	for _, forbidden := range []string{
		"time.Now",
		"net/http",
		"http.",
		"github.com/nats-io/nats.go",
		"stalwart.",
		"SaveDraft(",
		"ReserveMailboxDraftCreate(",
		"AdvanceMailboxDraftCreateToImport(",
		"mailboxDraftCreateWithAttachments(",
		"mailboxDraftCreateWithoutAttachments(",
		"calendarBytes()",
		"serialized",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("forbidden creator term %q", forbidden)
		}
	}
}

func TestMailboxCalendarDeliveryDraftCreatorUnknownAndRestartOrder(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_calendar_delivery_draft_creator.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	uploadStart := strings.Index(source, "uploadCalendarDraftCreateMIME")
	advanceStart := strings.Index(source, "AdvanceMailboxDraftCalendarCreateToImport")
	importStart := strings.Index(source, "importCalendarDraftCreateMIME")
	finishStart := strings.Index(source, "FinishMailboxDraftCalendarCreateApplied")
	if uploadStart < 0 || advanceStart <= uploadStart || importStart <= advanceStart || finishStart <= importStart {
		t.Fatal("calendar MIME flow must remain upload, handoff, import, then finish")
	}
	importBranch := source[importStart:finishStart]
	if !strings.Contains(importBranch, "RestartMailboxDraftCalendarCreateUploadAfterBlobNotFound") || !strings.Contains(importBranch, "continue") {
		t.Fatal("persisted blob-not-found must rotate into the fresh bounded upload loop")
	}
	if !strings.Contains(importBranch, "creator.markUnknown(ctx, advanced") {
		t.Fatal("non-restartable import outcomes must be durably marked unknown")
	}
}
