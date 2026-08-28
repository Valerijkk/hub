package mail_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/uuid"

	mail "github.com/valerii/hub/backend/internal/mail"
)

var _ interface {
	ResolveOwner(context.Context, uuid.UUID) (mail.MailboxOwner, error)
} = (*mail.MailboxService)(nil)

func mailboxAttachmentWiringSource(t *testing.T) string {
	t.Helper()

	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	sourcePath := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "cmd", "server", "wiring_calendar_mail.go"))
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read Calendar mail wiring source: %v", err)
	}
	return string(source)
}

func mailboxAttachmentWiringFunctionSlice(t *testing.T, source, declaration string) string {
	t.Helper()

	startIndex := strings.Index(source, declaration)
	if startIndex < 0 {
		t.Fatalf("source is missing function declaration %q", declaration)
	}
	remainder := source[startIndex+len(declaration):]
	if nextIndex := strings.Index(remainder, "\nfunc "); nextIndex >= 0 {
		return source[startIndex : startIndex+len(declaration)+nextIndex]
	}
	return source[startIndex:]
}

func TestMailboxAttachmentOwnerWiringSourceContract(t *testing.T) {
	wiring := mailboxAttachmentWiringFunctionSlice(t, mailboxAttachmentWiringSource(t), "func wireCalendarMail(d *serverDeps) {")
	compact := strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "").Replace(wiring)
	for _, required := range []string{
		"mailboxAttachmentStagingService,err:=mailmod.NewMailboxAttachmentStagingService(mailboxSvc,mailboxRepo,mailboxRepo,mailboxRepo,mailboxObjectStore,mailmod.MailboxAttachmentStagingConfig{",
		"mailboxAttachmentAbandonService,err:=mailmod.NewMailboxAttachmentAbandonService(mailboxSvc,mailboxRepo)",
	} {
		if !strings.Contains(compact, required) {
			t.Fatalf("attachment wiring is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"mailmod.NewMailboxAttachmentStagingService(mailboxRepo,",
		"mailmod.NewMailboxAttachmentAbandonService(mailboxRepo,",
	} {
		if strings.Contains(compact, forbidden) {
			t.Fatalf("attachment wiring retains owner repository position %q", forbidden)
		}
	}
}
