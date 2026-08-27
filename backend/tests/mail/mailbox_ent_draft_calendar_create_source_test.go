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
		"SetPendingCreateMIMEBindingDigest",
		"SetPendingCreateMIMEAttempt(1)",
		"SetPendingCreateMIMEAttempt(nextAttempt)",
		"ClearPendingCreateMIMEKind()",
		"ClearPendingCreateMIMEPhase()",
		"ClearPendingCreateMIMEBindingDigest()",
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
		{"ReserveMailboxDraftCalendarCreate", []string{"mailboxEntDraftCreateLockAccount", "SetPendingCreateMIMEKind", "SetPendingCreateMIMEAttempt(1)", "MailboxDraftReservationDispatch"}},
		{"AdvanceMailboxDraftCalendarCreateToImport", []string{"mailboxEntDraftCalendarCreateLock", "SetPendingCreateMIMEBlobID", "MailboxDraftCreateMIMEPhaseImport", "ExpectedBindingDigest"}},
		{"MarkMailboxDraftCalendarCreateUnknown", []string{"mailboxEntDraftCalendarCreateLock", "mailboxEntDraftCalendarCreateUnknownUploadSchedule", "SetStatus(maildraft.StatusUnknown)", "mailboxEntDraftCreateRelease"}},
		{"FinishMailboxDraftCalendarCreateApplied", []string{"mailboxEntDraftCalendarCreateLock", "SetStatus(maildraft.StatusActive)", "SetConfirmedMutationID(request.MutationID)", "ClearPendingCreateMIMEBindingDigest"}},
		{"RestartMailboxDraftCalendarCreateUpload", []string{"mailboxEntDraftCalendarCreateLock", "RestartableUnknown", "SetPendingCreateMIMEAttempt(nextAttempt)", "NextBindingDigest"}},
		{"RestartMailboxDraftCalendarCreateUploadAfterBlobNotFound", []string{"mailboxEntDraftCalendarCreateLock", "ExpectedBlobID", "ExpectedSizeBytes", "SetPendingCreateMIMEAttempt(nextAttempt)"}},
		{"AbandonMailboxDraftCalendarCreateUpload", []string{"mailboxEntDraftCalendarCreateLock", "SetStatus(maildraft.StatusAbandoned)", "SetAbandonedMutationID(request.ExpectedMutationID)", "ClearPendingCreateMIMEKind"}},
	} {
		section := mailboxEntDraftCalendarCreateSourceSection(t, text, contract.name)
		for _, marker := range contract.markers {
			if !strings.Contains(section, marker) {
				t.Fatalf("%s source is missing %q", contract.name, marker)
			}
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
