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

func TestMailboxCalendarDeliveryDraftCreatorFactoryAndReaderSourceContract(t *testing.T) {
	creatorSourceBytes, err := os.ReadFile("../../internal/mail/mailbox_calendar_delivery_draft_creator.go")
	if err != nil {
		t.Fatal(err)
	}
	creatorSource := string(creatorSourceBytes)
	for _, term := range []string{
		"func NewCalendarMailDeliveryDraftCreator(",
		"owners *MailboxService",
		"repository *MailboxEntRepository",
		"engine MailboxDraftEngine",
		"engine.(mailboxDraftCalendarCreateEngine)",
		"newMailboxDraftOwnerBoundReader(owners, repository, engine)",
		"newMailboxCalendarDeliveryDraftCreator(owners, repository, repository, repository, calendarEngine, reader)",
	} {
		if !strings.Contains(creatorSource, term) {
			t.Fatalf("missing creator factory term %q", term)
		}
	}
	for _, forbidden := range []string{
		"NewMailboxDraftService(",
		"type CalendarMIMEPart",
		"type MailboxCalendarDeliveryDraftReader",
	} {
		if strings.Contains(creatorSource, forbidden) {
			t.Fatalf("forbidden creator factory surface %q", forbidden)
		}
	}

	readerSourceBytes, err := os.ReadFile("../../internal/mail/mailbox_draft_get.go")
	if err != nil {
		t.Fatal(err)
	}
	readerSource := string(readerSourceBytes)
	for _, term := range []string{
		"type mailboxDraftOwnerBoundReader struct",
		"func newMailboxDraftOwnerBoundReader(",
		"func (reader *mailboxDraftOwnerBoundReader) GetMailboxDraft(",
		"func mailboxDraftGetOwnerBound(",
		"func (service *MailboxDraftService) GetMailboxDraft(",
		"return mailboxDraftGetOwnerBound(ctx, service.owners, service.repository, service.engine, actorID, draftID)",
		"return mailboxDraftGetOwnerBound(ctx, reader.owners, reader.repository, reader.engine, actorID, draftID)",
		"owners.ResolveOwner(ctx, actorID)",
		"repository.GetMailboxDraft(ctx, MailboxDraftGetRequest{",
		"engine.GetDraft(ctx, EngineDraftReadRequest{",
		"repository.RecheckMailboxDraftRead(ctx, MailboxDraftReadRecheckRequest{",
		"mailboxDraftGetNormalizeContent(message.Body)",
	} {
		if !strings.Contains(readerSource, term) {
			t.Fatalf("missing owner-bound reader term %q", term)
		}
	}
}

func TestMailboxCalendarDeliveryDraftCreatorUnknownAndRestartOrder(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_calendar_delivery_draft_creator.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	createBody := mailboxCalendarDeliveryDraftCreatorMethodBody(t, source, "CreateMailboxDraft")
	uploadStart := strings.Index(createBody, "creator.engine.uploadCalendarDraftCreateMIME")
	advanceStart := strings.Index(createBody, "creator.calendar.AdvanceMailboxDraftCalendarCreateToImport")
	importStart := strings.Index(createBody, "creator.engine.importCalendarDraftCreateMIME")
	finishStart := strings.Index(createBody, "creator.calendar.FinishMailboxDraftCalendarCreateApplied")
	if uploadStart < 0 || advanceStart <= uploadStart || importStart <= advanceStart || finishStart <= importStart {
		t.Fatal("calendar MIME flow must remain upload, handoff, import, then finish")
	}
	importBranch := createBody[importStart:finishStart]
	if !strings.Contains(importBranch, "RestartMailboxDraftCalendarCreateUploadAfterBlobNotFound") || !strings.Contains(importBranch, "continue") {
		t.Fatal("persisted blob-not-found must rotate into the fresh bounded upload loop")
	}
	if !strings.Contains(importBranch, "creator.markUnknown(ctx, advanced") {
		t.Fatal("non-restartable import outcomes must be durably marked unknown")
	}
}

func mailboxCalendarDeliveryDraftCreatorMethodBody(t *testing.T, source, method string) string {
	t.Helper()
	declaration := "func (creator *mailboxCalendarDeliveryDraftCreator) " + method + "("
	start := strings.Index(source, declaration)
	if start < 0 {
		t.Fatalf("creator method %q not found", method)
	}
	next := strings.Index(source[start+len(declaration):], "\nfunc (creator *mailboxCalendarDeliveryDraftCreator) ")
	if next < 0 {
		return source[start:]
	}
	return source[start : start+len(declaration)+next]
}
