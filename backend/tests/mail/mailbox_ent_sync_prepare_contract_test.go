package mail_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) = (*mail.MailboxEntRepository).PrepareMailboxSyncPage

func mailboxEntPrepareSourceSlice(t *testing.T, source, start, end string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex == -1 {
		t.Fatalf("source is missing start marker %q", start)
	}
	endOffset := strings.Index(source[startIndex+len(start):], end)
	if endOffset == -1 {
		t.Fatalf("source is missing end marker %q after %q", end, start)
	}
	return source[startIndex : startIndex+len(start)+endOffset]
}

func requireMailboxEntPrepareOrderedFragments(t *testing.T, source string, fragments ...string) {
	t.Helper()

	previous := -1
	for _, fragment := range fragments {
		position := strings.Index(source, fragment)
		if position == -1 {
			t.Fatalf("source is missing %q", fragment)
		}
		if position <= previous {
			t.Fatalf("source fragment %q is out of order", fragment)
		}
		previous = position
	}
}

func requireMailboxEntPrepareNoExternalIO(t *testing.T, source string) {
	t.Helper()

	for _, forbidden := range []string{"http.", "net/http", "nats.", "NATS", "MailEngine", "stalwart", "os.", "log.", "callback"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("prepare path contains forbidden I/O vocabulary %q", forbidden)
		}
	}
}

func TestMailboxEntSyncPrepareSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_sync_prepare.go")
	if err != nil {
		t.Fatalf("read prepare source: %v", err)
	}
	source := string(sourceBytes)
	entry := mailboxEntPrepareSourceSlice(t, source,
		"func (repository *MailboxEntRepository) PrepareMailboxSyncPage",
		"func mailboxEntPrepareRequestValid",
	)
	input := mailboxEntPrepareSourceSlice(t, source,
		"func mailboxEntPrepareInput",
		"func mailboxEntStableSyncID",
	)

	requireMailboxEntPrepareOrderedFragments(t, entry,
		"!mailboxEntPrepareRequestValid(request)",
		"repository.withMailboxEntTransaction",
		"now, err := repository.mailboxEntNow()",
		"mailaccount.IDEQ(request.Fence.AccountID)",
		"ForUpdate()",
		"mailboxEntPrepareInput",
		"validMailboxSyncPrepareSelection",
		"ClassifyMailboxSyncPage",
	)
	requireMailboxEntPrepareOrderedFragments(t, input,
		"mailboxEntSyncLeaseFence(account)",
		"persistedFence != request.Fence",
		"account.SyncEpoch != request.Fence.SyncEpoch",
		"account.MutationEpoch != request.Fence.MutationEpoch",
		"account.SyncLeaseUntil.After(now)",
	)
	if got := strings.Count(entry, "repository.mailboxEntNow()"); got != 1 {
		t.Fatalf("prepare durable clock reads = %d, want 1", got)
	}
	prepare := entry + input
	for _, forbidden := range []string{".Create()", ".Update()", ".Delete()", ".Exec("} {
		if strings.Contains(prepare, forbidden) {
			t.Fatalf("prepare must be read-only: found %q", forbidden)
		}
	}
	requireMailboxEntPrepareNoExternalIO(t, prepare)
}

func TestMailboxEntSyncPrepareTombstoneSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_sync_prepare.go")
	if err != nil {
		t.Fatalf("read prepare source: %v", err)
	}
	source := string(sourceBytes)
	persisted := mailboxEntPrepareSourceSlice(t, source,
		"func mailboxEntPreparePersistedStates",
		"func mailboxEntPrepareSortedUUIDStrings",
	)
	messages := mailboxEntPrepareSourceSlice(t, source,
		"func mailboxEntPrepareMessages",
		"func mailboxEntPrepareMessageIDs",
	)
	input := mailboxEntPrepareSourceSlice(t, source,
		"func mailboxEntPrepareInput",
		"func mailboxEntStableSyncID",
	)
	requireMailboxEntPrepareOrderedFragments(t, input,
		"stateRows := make(map[string]*ent.MailMessage, len(messageRows))",
		"for _, projection := range projections {",
		"if row, ok := messageRows[projection.MessageID]; ok {",
		"stateRows[projection.MessageID] = row",
		"if page.Change != nil {",
		"for _, engineMessageID := range page.Change.DestroyedIDs {",
		"if row, ok := messageRows[engineMessageID]; ok && row.DeletedAt == nil {",
		"stateRows[engineMessageID] = row",
		"mailboxEntPreparePersistedStates(ctx, transaction, account, stateRows)",
	)
	if strings.Contains(input, "mailboxEntPreparePersistedStates(ctx, transaction, account, messageRows)") {
		t.Fatal("mailboxEntPrepareInput must not pass all message rows to mailboxEntPreparePersistedStates")
	}

	requireMailboxEntPrepareOrderedFragments(t, persisted,
		"existing[row.ID] = row",
		"mailthread.IDIn(threadIDs...)",
		"folderState := make([]MailboxEventID, 0)",
		"flags := EngineMessageFlags{}",
		"ReceivedEventID: syncCopyUUIDPointer(row.ReceivedEventID)",
	)
	requireMailboxEntPrepareOrderedFragments(t, persisted,
		"message, found := existing[link.MailMessageID]",
		"if !found { return nil, mailboxServiceError(MailboxServiceErrorInvalid) }",
		"if message.DeletedAt != nil { continue }",
		"foldersByMessage[link.MailMessageID] = append(foldersByMessage[link.MailMessageID], link)",
		"folderIDs[link.MailFolderID] = struct{}{}",
	)
	if strings.Contains(persisted, "if row.DeletedAt != nil { continue }") {
		t.Fatal("persisted states must include soft-deleted message rows")
	}
	if strings.Contains(persisted, "mailthread.IDIn(threadIDs...), mailthread.DeletedAtIsNil()") {
		t.Fatal("persisted-state thread query must include deleted threads")
	}
	if strings.Contains(messages, "found && row.DeletedAt == nil") {
		t.Fatal("an existing soft-deleted message must use its persisted state")
	}
}

func TestMailboxEntSyncPrepareStableIDSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_sync_prepare.go")
	if err != nil {
		t.Fatalf("read prepare source: %v", err)
	}
	source := string(sourceBytes)
	stableID := mailboxEntPrepareSourceSlice(t, source,
		"func mailboxEntStableSyncID",
		"func mailboxEntPrepareFolders",
	)

	requireMailboxEntPrepareOrderedFragments(t, stableID,
		"sha256.New()",
		"binary.BigEndian.PutUint32",
		"mailboxEntSyncIDDomain",
		"accountID[:]",
		"copy(value[:], digest[:16])",
		"value[6] = (value[6] & 0x0f) | 0x80",
		"value[8] = (value[8] & 0x3f) | 0x80",
		"uuid.FromBytes(value[:])",
	)
	for _, forbidden := range []string{"uuid.New", "rand.", "time.", "repository."} {
		if strings.Contains(stableID, forbidden) {
			t.Fatalf("stable ID must be pure: found %q", forbidden)
		}
	}
}

func TestMailboxEntSyncPrepareResolutionSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_sync_prepare.go")
	if err != nil {
		t.Fatalf("read prepare source: %v", err)
	}
	source := string(sourceBytes)

	for _, required := range []string{
		"mailfolder.MailAccountIDEQ(account.ID)",
		"mailfolder.EngineFolderIDIn(engineFolderIDs...)",
		"mailmessage.MailAccountIDEQ(account.ID)",
		"mailmessage.EngineMessageIDIn(engineMessageIDs...)",
		"mailthread.MailAccountIDEQ(account.ID)",
		"mailthread.ThreadKeyVersionEQ(int32(identity.KeyVersion))",
		"mailthread.ThreadKeyEQ(identity.Key[:])",
		"mailmessage.DeletedAtIsNil()",
		"mailthread.DeletedAtIsNil()",
		"maillabel.DeletedAtIsNil()",
		"mailsubmission.MailAccountIDEQ(account.ID)",
		"maildraft.MailAccountIDEQ(account.ID)",
		"outboxevent.IDIn(candidateIDs...)",
		"Projected: false",
		"ConfirmedSubmission",
		"ConfirmedDraft",
		"sort.Slice",
		"mailboxEntPrepareCopyInput",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("prepare source is missing durable-resolution guard %q", required)
		}
	}
}

func TestMailboxEntSyncPrepareThreadsLiveOnlySourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_sync_prepare.go")
	if err != nil {
		t.Fatalf("read prepare source: %v", err)
	}
	threads := mailboxEntPrepareSourceSlice(t, string(sourceBytes),
		"func mailboxEntPrepareThreads",
		"func mailboxEntPrepareAncestors",
	)

	requireMailboxEntPrepareOrderedFragments(t, threads,
		"mailthread.MailAccountIDEQ(account.ID)",
		"mailthread.Or(threadPredicates...)",
		"mailthread.DeletedAtIsNil()",
		".All(ctx)",
	)
	if !strings.Contains(threads, "row.DeletedAt != nil") {
		t.Fatal("live thread resolution must reject soft-deleted rows")
	}
}
