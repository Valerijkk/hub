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

type mailboxReplyDraftBootstrapCapabilities interface {
	CreateMailboxDraft(context.Context, uuid.UUID, mail.MailboxDraftCreateInput) (mail.MailboxDraftView, error)
	GetMailboxDraft(context.Context, uuid.UUID, uuid.UUID) (mail.MailboxDraftView, error)
}

type mailboxForwardDraftBootstrapCapabilities interface {
	CreateMailboxDraft(context.Context, uuid.UUID, mail.MailboxDraftCreateInput) (mail.MailboxDraftView, error)
	GetMailboxDraft(context.Context, uuid.UUID, uuid.UUID) (mail.MailboxDraftView, error)
}

var _ mailboxReplyDraftBootstrapCapabilities = (*mail.MailboxDraftService)(nil)
var _ mailboxForwardDraftBootstrapCapabilities = (*mail.MailboxDraftService)(nil)

func mailboxDraftBootstrapSource(t *testing.T, file string) string {
	t.Helper()

	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	sourcePath := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "internal", "mail", file))
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read Mail source %q: %v", file, err)
	}
	return string(source)
}

func mailboxDraftBootstrapSourceSlice(t *testing.T, source, start, nextDeclaration string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source is missing start declaration %q", start)
	}
	nextIndex := strings.Index(source[startIndex+len(start):], nextDeclaration)
	if nextIndex < 0 {
		t.Fatalf("source is missing next declaration %q", nextDeclaration)
	}
	return source[startIndex : startIndex+len(start)+nextIndex]
}

func requireMailboxDraftBootstrapSource(t *testing.T, source string, terms ...string) {
	t.Helper()

	for _, term := range terms {
		if !strings.Contains(source, term) {
			t.Fatalf("source is missing %q", term)
		}
	}
}

func forbidMailboxDraftBootstrapSource(t *testing.T, source string, terms ...string) {
	t.Helper()

	for _, term := range terms {
		if strings.Contains(source, term) {
			t.Fatalf("source retains forbidden capability %q", term)
		}
	}
}

func TestMailboxReplyDraftBootstrapCapabilitiesSourceContract(t *testing.T) {
	reply := mailboxDraftBootstrapSourceSlice(
		t,
		mailboxDraftBootstrapSource(t, "mailbox_reply_draft_service.go"),
		"type mailboxReplyDrafts interface {",
		"type MailboxReplyDraftInput struct {",
	)
	requireMailboxDraftBootstrapSource(t, reply,
		"CreateMailboxDraft(context.Context, uuid.UUID, MailboxDraftCreateInput) (MailboxDraftView, error)",
		"GetMailboxDraft(context.Context, uuid.UUID, uuid.UUID) (MailboxDraftView, error)",
		"preflightMailboxReplyDraft(context.Context, uuid.UUID, mailboxReplyDraftPreflightInput) (mailboxReplyDraftPreflightResult, error)",
	)
	forbidMailboxDraftBootstrapSource(t, reply, "MailboxDraftServiceAPI", "SaveMailboxDraft", "DeleteMailboxDraft")
}

func TestMailboxForwardDraftBootstrapCapabilitiesSourceContract(t *testing.T) {
	forward := mailboxDraftBootstrapSourceSlice(
		t,
		mailboxDraftBootstrapSource(t, "mailbox_forward_draft_service.go"),
		"type mailboxForwardDrafts interface {",
		"type MailboxForwardDraftInput struct {",
	)
	requireMailboxDraftBootstrapSource(t, forward,
		"CreateMailboxDraft(context.Context, uuid.UUID, MailboxDraftCreateInput) (MailboxDraftView, error)",
		"GetMailboxDraft(context.Context, uuid.UUID, uuid.UUID) (MailboxDraftView, error)",
		"preflightMailboxForwardDraft(context.Context, uuid.UUID, mailboxForwardDraftPreflightInput) (mailboxForwardDraftPreflightResult, error)",
	)
	forbidMailboxDraftBootstrapSource(t, forward, "MailboxDraftServiceAPI", "SaveMailboxDraft", "DeleteMailboxDraft")
}

func TestMailboxReplyAllDraftSharesReplyBootstrapCapabilitiesSourceContract(t *testing.T) {
	replyAll := mailboxDraftBootstrapSourceSlice(
		t,
		mailboxDraftBootstrapSource(t, "mailbox_reply_all_draft_service.go"),
		"func NewMailboxReplyAllDraftService(",
		"func (service *MailboxReplyAllDraftService) CreateMailboxReplyAllDraft(",
	)
	requireMailboxDraftBootstrapSource(t, replyAll, "drafts mailboxReplyDrafts,")
}
