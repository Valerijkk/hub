package mail_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMailboxEntDraftCalendarCreateSourceContract(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	sourcePath := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "internal", "mail", "mailbox_ent_draft_calendar_create.go"))
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read calendar draft-create Ent source: %v", err)
	}
	text := string(source)

	for _, required := range []string{
		"func (repository *MailboxEntRepository) ReserveMailboxDraftCalendarCreate",
		"func (repository *MailboxEntRepository) AdvanceMailboxDraftCalendarCreateToImport",
		"func (repository *MailboxEntRepository) MarkMailboxDraftCalendarCreateUnknown",
		"func (repository *MailboxEntRepository) FinishMailboxDraftCalendarCreateApplied",
		"func (repository *MailboxEntRepository) RestartMailboxDraftCalendarCreateUpload",
		"func (repository *MailboxEntRepository) RestartMailboxDraftCalendarCreateUploadAfterBlobNotFound",
		"func (repository *MailboxEntRepository) AbandonMailboxDraftCalendarCreateUpload",
		"withMailboxEntTransaction",
		"mailboxEntDraftCreateLockAccount",
		"mailboxEntDraftCreateLockIdentityRows",
		"mailboxEntDraftCreateLockConfirmedIdentity",
		"mailboxEntDraftReservationLockDraft",
		"MailboxDraftCreateMIMEKindCalendar",
		"MailboxDraftCreateMIMEPhaseUpload",
		"MailboxDraftCreateMIMEPhaseImport",
		"SetPendingCreateMimeBindingDigest",
		"SetPendingCreateMimeAttempt(1)",
		"SetPendingCreateMimeAttempt(nextAttempt)",
		"ClearPendingCreateMimeKind()",
		"ClearPendingCreateMimePhase()",
		"ClearPendingCreateMimeBindingDigest()",
		"SetStatus(maildraft.StatusUnknown)",
		"SetStatus(maildraft.StatusActive)",
		"SetStatus(maildraft.StatusAbandoned)",
		"SetRevision(1)",
		"SetConfirmedMutationID(request.MutationID)",
		"ClearLastMutationID()",
		"ClearPendingMutationKind()",
		"SetAbandonedMutationID(request.ExpectedMutationID)",
		"mailboxEntDraftCreateRelease",
		"mailboxEntCheckedTimeAdd",
		"mailboxDraftCreateUploadRetryAfterFirstAttempt",
		"mailboxDraftCreateUploadRetryAfterSecondAttempt",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("calendar draft-create Ent source missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"time.Now",
		"MailEngine",
		"provider",
		"raw MIME",
		"raw ICS",
		"MailboxDraftCalendarCreateRepository",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("calendar draft-create Ent source contains forbidden %q", forbidden)
		}
	}

	for _, cas := range []string{
		"mailboxEntDraftCalendarCreateMatches",
		"ExpectedBindingDigest",
		"ExpectedMutationID",
		"ExpectedAttempt",
		"ExpectedBlobID",
		"ExpectedSizeBytes",
		"NextMutationID",
		"NextBindingDigest",
		"IdempotencyKeyHash",
		"RequestHash",
	} {
		if !strings.Contains(text, cas) {
			t.Fatalf("calendar draft-create Ent source missing CAS marker %q", cas)
		}
	}

	for _, contract := range []struct {
		name     string
		markers  []string
	}{
		{"ReserveMailboxDraftCalendarCreate", []string{"mailboxEntDraftCreateLockAccount", "SetPendingCreateMimeKind", "SetPendingCreateMimeAttempt(1)", "MailboxDraftReservationDispatch"}},
		{"AdvanceMailboxDraftCalendarCreateToImport", []string{"mailboxEntDraftCalendarCreateLock", "SetPendingCreateMimeBlobID", "MailboxDraftCreateMIMEPhaseImport", "ExpectedBindingDigest"}},
		{"MarkMailboxDraftCalendarCreateUnknown", []string{"mailboxEntDraftCalendarCreateLock", "mailboxEntDraftCalendarCreateUnknownUploadSchedule", "SetStatus(maildraft.StatusUnknown)", "mailboxEntDraftCreateRelease"}},
		{"FinishMailboxDraftCalendarCreateApplied", []string{"mailboxEntDraftCalendarCreateLock", "SetStatus(maildraft.StatusActive)", "SetConfirmedMutationID(request.MutationID)", "ClearPendingCreateMimeBindingDigest"}},
		{"RestartMailboxDraftCalendarCreateUpload", []string{"mailboxEntDraftCalendarCreateLock", "RestartableUnknown", "SetPendingCreateMimeAttempt(nextAttempt)", "NextBindingDigest"}},
		{"RestartMailboxDraftCalendarCreateUploadAfterBlobNotFound", []string{"mailboxEntDraftCalendarCreateLock", "ExpectedBlobID", "ExpectedSizeBytes", "SetPendingCreateMimeAttempt(nextAttempt)"}},
		{"AbandonMailboxDraftCalendarCreateUpload", []string{"mailboxEntDraftCalendarCreateLock", "SetStatus(maildraft.StatusAbandoned)", "SetAbandonedMutationID(request.ExpectedMutationID)", "ClearPendingCreateMimeKind"}},
	} {
		section := mailboxEntDraftCalendarCreateSourceSection(t, text, contract.name)
		for _, marker := range contract.markers {
			if !strings.Contains(section, marker) {
				t.Fatalf("%s source is missing %q", contract.name, marker)
			}
		}
	}
}

func TestMailboxEntDraftCalendarCreateExistingReplayUsesPersistedMutation(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	sourcePath := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "internal", "mail", "mailbox_ent_draft_calendar_create.go"))
	sourceBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read calendar draft-create Ent source: %v", err)
	}
	section := mailboxEntDraftCalendarCreateClassifierSection(t, string(sourceBytes))

	for _, required := range []string{
		"mailboxEntDraftCreateSameRequest(record, request.Create, engineDraftKey)",
		"record.ConfirmedMutationID == nil || *record.ConfirmedMutationID == uuid.Nil",
		"record.LastMutationID == nil || *record.LastMutationID == uuid.Nil",
		"state.BindingDigest != request.State.BindingDigest",
		"mailboxEntDraftCreateWork(record, MailboxDraftWorkReconcileCreate)",
		"MailboxDraftReservationReplay",
		"MailboxDraftReservationReconcileRequired",
		"MailboxDraftRepositoryErrorIdempotencyConflict",
	} {
		if !strings.Contains(section, required) {
			t.Fatalf("existing replay classifier missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"*record.ConfirmedMutationID != request.Create.MutationID",
		"*record.LastMutationID != request.Create.MutationID",
	} {
		if strings.Contains(section, forbidden) {
			t.Fatalf("existing replay classifier compares ephemeral mutation %q", forbidden)
		}
	}
}

func mailboxEntDraftCalendarCreateSourceSection(t *testing.T, source, method string) string {
	t.Helper()
	needle := "func (repository *MailboxEntRepository) " + method
	start := strings.Index(source, needle)
	if start < 0 {
		t.Fatalf("source method %q is absent", method)
	}
	rest := source[start+len(needle):]
	next := strings.Index(rest, "\nfunc (repository *MailboxEntRepository) ")
	if next < 0 {
		return source[start:]
	}
	return source[start : start+len(needle)+next]
}

func mailboxEntDraftCalendarCreateClassifierSection(t *testing.T, source string) string {
	t.Helper()
	needle := "func mailboxEntDraftCalendarCreateClassifyExisting"
	start := strings.Index(source, needle)
	if start < 0 {
		t.Fatal("calendar create existing-row classifier is absent")
	}
	rest := source[start+len(needle):]
	next := strings.Index(rest, "\nfunc mailboxEntDraftCalendarCreateRecord")
	if next < 0 {
		t.Fatal("calendar create existing-row classifier end is absent")
	}
	return source[start : start+len(needle)+next]
}
