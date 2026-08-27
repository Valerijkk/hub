package mail_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFakeMailEngineCalendarCreateSourceContract(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	sourcePath := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "internal", "mail", "mailbox_fake.go"))
	sourceBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read FakeMailEngine source: %v", err)
	}
	source := string(sourceBytes)

	for _, required := range []string{
		"var _ mailboxDraftCalendarCreateEngine = (*FakeMailEngine)(nil)",
		"FakeMailOperationUploadCalendarDraftCreateMIME",
		"FakeMailOperationImportCalendarDraftCreateMIME",
		"calendarDraftCreates map[fakeCalendarDraftCreateKey]fakeCalendarDraftCreateStage",
		"calendarDraftCreates: make(map[fakeCalendarDraftCreateKey]fakeCalendarDraftCreateStage)",
		"for key := range f.calendarDraftCreates",
		"func (f *FakeMailEngine) uploadCalendarDraftCreateMIME",
		"func (f *FakeMailEngine) importCalendarDraftCreateMIME",
		"mailboxDraftMIMEPlanForDraftWithCalendar",
		"plan.writeTo(ctx, digest)",
		"errMailboxDraftCalendarCreatePersistedBlobNotFound",
		"messageFromDraft",
		"request.receivedAt",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("FakeMailEngine calendar create source missing %q", required)
		}
	}
	upload := fakeCalendarCreateSourceSection(t, source, "uploadCalendarDraftCreateMIME")
	for _, required := range []string{
		"ctx == nil",
		"calendar == nil",
		"!calendar.valid()",
		"len(draft.Message.Attachments) != 0",
		"mailboxDraftMIMECalendarRecipientValid",
		"mailboxDraftMIMEPlanValid(plan)",
		"plan.size",
		"calendar.bindingDigest()",
		"calendar.contentDigest()",
		"fakeCalendarDraftCreateStageMatches",
		"f.takeFaultLocked(FakeMailOperationUploadCalendarDraftCreateMIME)",
		"faulted && !fault.AfterDispatch",
		"faulted && fault.AfterDispatch",
		"f.nextOpaqueID()",
		"f.reserveOpaqueIDsLocked(candidateBlobID)",
	} {
		if !strings.Contains(upload, required) {
			t.Fatalf("calendar upload source missing %q", required)
		}
	}

	importSection := fakeCalendarCreateSourceSection(t, source, "importCalendarDraftCreateMIME")
	for _, required := range []string{
		"validMailboxDraftCalendarCreateImport(request)",
		"fakeCalendarDraftCreateImportMatches",
		"errMailboxDraftCalendarCreatePersistedBlobNotFound",
		"fakeCalendarDraftCreateImportedMessageMatches",
		"f.takeFaultLocked(FakeMailOperationImportCalendarDraftCreateMIME)",
		"faulted && !fault.AfterDispatch",
		"faulted && fault.AfterDispatch",
		"f.nextOpaqueID()",
		"f.reserveOpaqueIDsLocked(candidateMessageID, candidateVersion)",
		"messageFromDraft(stage.draft, messageID, candidateVersion, request.receivedAt)",
		"EngineProjectionOriginConfirmedOwnerDraft",
	} {
		if !strings.Contains(importSection, required) {
			t.Fatalf("calendar import source missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"SaveDraft(ctx, draft)",
		"f.currentTime()",
		"calendar.calendarBytes()",
		"serialized: calendar",
	} {
		if strings.Contains(upload+"\n"+importSection, forbidden) {
			t.Fatalf("FakeMailEngine calendar create source contains forbidden %q", forbidden)
		}
	}
}

func fakeCalendarCreateSourceSection(t *testing.T, source, method string) string {
	t.Helper()
	needle := "func (f *FakeMailEngine) " + method
	start := strings.Index(source, needle)
	if start < 0 {
		t.Fatalf("FakeMailEngine method %q is absent", method)
	}
	rest := source[start+len(needle):]
	next := strings.Index(rest, "\nfunc (f *FakeMailEngine) ")
	if next < 0 {
		return source[start:]
	}
	return source[start : start+len(needle)+next]
}
