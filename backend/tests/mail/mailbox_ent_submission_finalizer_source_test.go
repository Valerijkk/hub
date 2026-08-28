package mail_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailboxSubmissionFinalizationRepository = (*mail.MailboxEntRepository)(nil)
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionFinalizationScanRequest) (mail.MailboxSubmissionFinalizationScanResult, error) = (*mail.MailboxEntRepository).ScanAcceptedMailboxSubmissions
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) = (*mail.MailboxEntRepository).ClaimAcceptedMailboxSubmission
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionFinalizationCompletion) error = (*mail.MailboxEntRepository).CompleteAcceptedMailboxSubmission
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionFinalizationClaim) error = (*mail.MailboxEntRepository).ReleaseAcceptedMailboxSubmission

func mailboxEntSubmissionFinalizerSourceSlice(t *testing.T, source, start, end string) string {
	t.Helper()
	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source is missing start marker %q", start)
	}
	endOffset := strings.Index(source[startIndex+len(start):], end)
	if endOffset < 0 {
		t.Fatalf("source is missing end marker %q after %q", end, start)
	}
	return source[startIndex : startIndex+len(start)+endOffset]
}

func requireMailboxEntSubmissionFinalizerOrdered(t *testing.T, source string, fragments ...string) {
	t.Helper()
	previous := -1
	for _, fragment := range fragments {
		position := strings.Index(source, fragment)
		if position < 0 {
			t.Fatalf("source is missing %q", fragment)
		}
		if position <= previous {
			t.Fatalf("source puts %q before its required predecessor", fragment)
		}
		previous = position
	}
}

func requireMailboxEntSubmissionFinalizerAbsent(t *testing.T, source string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if strings.Contains(source, fragment) {
			t.Fatalf("source contains forbidden %q", fragment)
		}
	}
}

func TestMailboxEntSubmissionFinalizerScanAndClaimSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_finalizer.go")
	if err != nil {
		t.Fatalf("read submission finalizer source: %v", err)
	}
	source := string(sourceBytes)
	scan := mailboxEntSubmissionFinalizerSourceSlice(t, source,
		"func (repository *MailboxEntRepository) ScanAcceptedMailboxSubmissions",
		"func (repository *MailboxEntRepository) ClaimAcceptedMailboxSubmission",
	)
	requireMailboxEntSubmissionFinalizerOrdered(t, scan,
		"repository.withMailboxEntTransaction",
		"mailboxEntDatabaseWallClock",
		"transaction.MailSubmission.Query()",
		"Order(ent.Asc(mailsubmission.FieldUpdatedAt), ent.Asc(mailsubmission.FieldID))",
		"Limit(request.Limit)",
	)
	for _, required := range []string{
		"mailsubmission.StatusEQ(mailsubmission.Status(\"accepted\"))",
		"mailsubmission.UpdatedAtLTE(databaseNow)",
		"mailsubmission.CommandLeaseTokenIsNil()",
		"mailsubmission.CommandLeaseExpiresAtIsNil()",
		"mailsubmission.ProviderSubmissionIDNotNil()",
		"mailsubmission.ProviderMessageIDNotNil()",
		"mailboxEntSubmissionFinalizationScanRowValid",
		"mailboxSubmissionFinalizationScanResultCopy",
	} {
		if !strings.Contains(scan, required) {
			t.Fatalf("scan source is missing %q", required)
		}
	}
	claim := mailboxEntSubmissionFinalizerSourceSlice(t, source,
		"func (repository *MailboxEntRepository) ClaimAcceptedMailboxSubmission",
		"func (repository *MailboxEntRepository) CompleteAcceptedMailboxSubmission",
	)
	requireMailboxEntSubmissionFinalizerOrdered(t, claim,
		"repository.client.MailSubmission.Query()",
		"repository.withMailboxEntTransaction",
		"transaction.MailAccount.Query()",
		"transaction.MailSubmission.Query()",
		"transaction.MailDraft.Query()",
		"mailboxEntDatabaseWallClock",
		"mailboxEntSubmissionFinalizationAcceptedShape",
		"mailboxEntSubmissionFinalizationLockExpectedAttachments",
		"SetMutationLeaseToken(leaseID)",
		"SetCommandLeaseToken(leaseID)",
		"NewMailboxSubmissionFinalizationClaimedResult",
	)
	for _, required := range []string{
		"MailboxEngineCommandLeaseDuration",
		"mailboxEntCheckedTimeAdd",
		"uuid.New()",
		"mailboxEntSubmissionFinalizationAcceptedShape",
		"submission.ProviderSubmissionID",
		"submission.ProviderMessageID",
		"submission.EngineBindingHash",
		"submission.DispatchAuthorizationHash",
		"submission.DispatchAuthorizedAt",
	} {
		if !strings.Contains(claim, required) {
			t.Fatalf("claim source is missing %q", required)
		}
	}
	requireMailboxEntSubmissionFinalizerAbsent(t, claim,
		"BeginEngineCommand(", "MailEngine", ".GetMessage(", ".GetBlob(", ".Submit(", ".ReconcileSubmission(",
		"provider.", "nats.", "http.", "net/http", "time.Now",
	)
}

func TestMailboxEntSubmissionFinalizerCompletionAndReleaseSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_finalizer.go")
	if err != nil {
		t.Fatalf("read submission finalizer source: %v", err)
	}
	source := string(sourceBytes)
	complete := mailboxEntSubmissionFinalizerSourceSlice(t, source,
		"func (repository *MailboxEntRepository) CompleteAcceptedMailboxSubmission",
		"func (repository *MailboxEntRepository) ReleaseAcceptedMailboxSubmission",
	)
	requireMailboxEntSubmissionFinalizerOrdered(t, complete,
		"transaction.MailAccount.Query()",
		"transaction.MailSubmission.Query()",
		"transaction.MailDraft.Query()",
		"mailboxEntSubmissionFinalizationTerminalReplayNoop",
		"mailboxEntSubmissionFinalizationPreReadAncestors",
		"DeriveThreadIdentity",
		"mailboxEntSubmissionFinalizationLockThreads",
		"mailboxEntSubmissionFinalizationResolveLockedThread",
		"mailboxEntSubmissionFinalizationLockMessages",
		"mailboxEntSubmissionFinalizationRevalidateAncestors",
		"mailboxEntSubmissionFinalizationLockAttachments",
		"mailboxEntSubmissionFinalizationUpsertMessage",
		"mailboxEntSubmissionFinalizationTransferAttachments",
		"mailboxEntSubmissionFinalizationResolveFoldersAndLabels",
		"mailboxEntSubmissionFinalizationResolveRoutes",
		"NewMailboxSubmissionSentEnvelopeV1",
		"mailboxEntSubmissionFinalizationAppendOutbox",
		"SetStatus(mailsubmission.Status(\"sent\"))",
		"ClearCommandLeaseToken()",
		"SetStatus(maildraft.Status(\"submitted\"))",
		"ClearMutationLeaseToken()",
		"SetMutationEpoch(nextEpoch)",
	)
	for _, required := range []string{"mailboxEntStableSyncID", "mailboxEntSubmissionFinalizationClaimFenceMatches"} {
		if !strings.Contains(complete, required) {
			t.Fatalf("completion source is missing %q", required)
		}
	}
	for _, required := range []string{
		"mailboxEntSubmissionFinalizationAttachmentsMatch",
		"SetProjectionOrigin(mailmessage.ProjectionOrigin(EngineProjectionOriginConfirmedOwnerSubmission))",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("finalizer source is missing %q", required)
		}
	}
	requireMailboxEntSubmissionFinalizerAbsent(t, complete,
		"mailboxEntApplyMessageAttachments(", "SetEngineBlobID(", "BeginEngineCommand(", "MailEngine", ".GetMessage(", ".GetBlob(",
		".Submit(", ".ReconcileSubmission(", "provider.", "nats.", "http.", "net/http", "time.Now",
	)
	release := mailboxEntSubmissionFinalizerSourceSlice(t, source,
		"func (repository *MailboxEntRepository) ReleaseAcceptedMailboxSubmission",
		"func mailboxEntSubmissionFinalizationScanRowValid",
	)
	requireMailboxEntSubmissionFinalizerOrdered(t, release,
		"transaction.MailAccount.Query()",
		"transaction.MailSubmission.Query()",
		"transaction.MailDraft.Query()",
		"mailboxEntDatabaseWallClock",
		"mailboxEntSubmissionFinalizationClaimFenceMatches",
		"ClearCommandLeaseToken()",
		"ClearMutationLeaseToken()",
		"SetMutationEpoch(nextEpoch)",
	)
	for _, required := range []string{"mailsubmission.Status(\"accepted\")", "ClearResultMailMessageID()", "ClearLastErrorCode()"} {
		if !strings.Contains(release, required) {
			t.Fatalf("release source is missing %q", required)
		}
	}
	requireMailboxEntSubmissionFinalizerAbsent(t, release,
		"SetStatus(mailsubmission.Status(\"queued\"))", "SetStatus(mailsubmission.Status(\"failed\"))", "draft.Update()",
		"BeginEngineCommand(", "MailEngine", ".GetMessage(", ".GetBlob(", ".Submit(", ".ReconcileSubmission(",
		"provider.", "nats.", "http.", "net/http", "time.Now",
	)
}

func TestMailboxEntSubmissionFinalizerReviewFixSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_finalizer.go")
	if err != nil {
		t.Fatalf("read submission finalizer source: %v", err)
	}
	source := string(sourceBytes)
	if !strings.Contains(source, "!submission.UpdatedAt.After(databaseNow)") {
		t.Fatal("scan row validator must reject rows newer than the database clock")
	}

	complete := mailboxEntSubmissionFinalizerSourceSlice(t, source,
		"func (repository *MailboxEntRepository) CompleteAcceptedMailboxSubmission",
		"func (repository *MailboxEntRepository) ReleaseAcceptedMailboxSubmission",
	)
	requireMailboxEntSubmissionFinalizerOrdered(t, complete,
		"mailboxEntSubmissionFinalizationPreReadAncestors",
		"DeriveThreadIdentity",
		"mailboxEntSubmissionFinalizationLockThreads",
		"mailboxEntSubmissionFinalizationResolveLockedThread",
		"mailboxEntSubmissionFinalizationLockMessages",
		"mailboxEntSubmissionFinalizationRevalidateAncestors",
	)

	threads := mailboxEntSubmissionFinalizerSourceSlice(t, source,
		"func mailboxEntSubmissionFinalizationLockThreads",
		"func mailboxEntSubmissionFinalizationResolveLockedThread",
	)
	for _, required := range []string{
		"mailthread.MailAccountIDEQ(account.ID)",
		"Order(ent.Asc(mailthread.FieldID))",
		"ForUpdate()",
		"mailboxEntApplySortedUUIDs",
	} {
		if !strings.Contains(threads, required) {
			t.Fatalf("thread lock source is missing %q", required)
		}
	}
	messages := mailboxEntSubmissionFinalizerSourceSlice(t, source,
		"func mailboxEntSubmissionFinalizationLockMessages",
		"func mailboxEntSubmissionFinalizationRevalidateAncestors",
	)
	requireMailboxEntSubmissionFinalizerOrdered(t, messages,
		"transaction.MailMessage.Query()",
		"mailmessage.MailAccountIDEQ(account.ID)",
		"mailmessage.Or(",
		"Order(ent.Asc(mailmessage.FieldID))",
		"ForUpdate()",
	)

	labels := mailboxEntSubmissionFinalizerSourceSlice(t, source,
		"func mailboxEntSubmissionFinalizationResolveFoldersAndLabels",
		"func mailboxEntSubmissionFinalizationResolveRoutes",
	)
	for _, required := range []string{
		"maillabel.EngineKeywordIn(keywords...)",
		"seenKeywords := make(map[string]struct{}, len(rows))",
		"keywordSet[row.EngineKeyword]",
		"MailboxSyncPolicyLabelState{Label: MailboxEventID{AccountID: accountID, ID: row.ID}, EngineRule: true}",
	} {
		if !strings.Contains(labels, required) {
			t.Fatalf("label resolution source is missing %q", required)
		}
	}
	requireMailboxEntSubmissionFinalizerAbsent(t, labels, "if len(rows) != len(keywords)")

	routes := mailboxEntSubmissionFinalizerSourceSlice(t, source,
		"func mailboxEntSubmissionFinalizationResolveRoutes",
		"func mailboxEntSubmissionFinalizationAppendOutbox",
	)
	for _, required := range []string{
		"maildomain.DomainIn(domainNames...)",
		"mailaccount.StatusEQ(mailaccount.StatusActive)",
		"mailaccount.DesiredEnabledEQ(true)",
		"recipient.MailDomainID != domain.ID",
	} {
		if !strings.Contains(routes, required) {
			t.Fatalf("route resolution source is missing %q", required)
		}
	}
	requireMailboxEntSubmissionFinalizerAbsent(t, routes, "ForUpdate()")
}

func TestMailboxEntSubmissionFinalizerTerminalReplaySourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_finalizer.go")
	if err != nil {
		t.Fatalf("read submission finalizer source: %v", err)
	}
	source := string(sourceBytes)
	complete := mailboxEntSubmissionFinalizerSourceSlice(t, source,
		"func (repository *MailboxEntRepository) CompleteAcceptedMailboxSubmission",
		"func (repository *MailboxEntRepository) ReleaseAcceptedMailboxSubmission",
	)
	requireMailboxEntSubmissionFinalizerOrdered(t, complete,
		"transaction.MailAccount.Query()",
		"transaction.MailSubmission.Query()",
		"transaction.MailDraft.Query()",
		"mailboxEntSubmissionFinalizationTerminalReplayNoop",
		"mailboxEntDatabaseWallClock",
		"mailboxEntSubmissionFinalizationClaimFenceMatches",
	)

	terminal := mailboxEntSubmissionFinalizerSourceSlice(t, source,
		"func mailboxEntSubmissionFinalizationTerminalReplayNoop",
		"func mailboxEntSubmissionFinalizationScanRowValid",
	)
	requireMailboxEntSubmissionFinalizerOrdered(t, terminal,
		"MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusSent",
		"mailboxEntSubmissionFinalizationTerminalClaimMatches",
		"mailboxEntSubmissionFinalizationTerminalMessageMatches",
		"mailboxEntSubmissionFinalizationTerminalOutbox",
		"return true, nil",
	)
	for _, required := range []string{
		"submission.ResultMailMessageID == nil || *submission.ResultMailMessageID != messageID",
		"MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusAccepted",
		"submission.CommandLeaseToken != nil || submission.CommandLeaseExpiresAt != nil",
		"account.MutationEpoch < nextEpoch",
		"mailmessage.MailAccountIDEQ(account.ID)",
		"mailmessage.EngineMessageIDEQ(engineMessageID)",
		"mailmessage.ProjectionOrigin(EngineProjectionOriginConfirmedOwnerSubmission)",
		"mailboxSubmissionSentEventID(submission.ID)",
		"json.Valid(outbox.Payload)",
		"json.Unmarshal(outbox.Payload, &envelope)",
		"mailSentEventJSON",
		"validMailEventTime(payload.SentAt)",
		"payload.SentAt.Zone()",
		"MailboxEventID{AccountID: account.ID, ID: payload.MailboxID}",
		"mailboxEntSubmissionFinalizationTerminalEventAddress(payload.From)",
		"mailboxEntSubmissionFinalizationTerminalEventAddresses(payload.To)",
		"NewMailSentEnvelopeV1",
		"subtle.ConstantTimeCompare(canonical, outbox.Payload)",
		"json.Marshal(payload)",
		"json.Marshal(envelope)",
		"subtle.ConstantTimeCompare(envelopeJSON, outbox.Payload)",
	} {
		if !strings.Contains(terminal, required) {
			t.Fatalf("terminal replay source is missing %q", required)
		}
	}
	requireMailboxEntSubmissionFinalizerAbsent(t, terminal,
		".Update()", ".Create()", "BeginEngineCommand(", "MailEngine", ".GetMessage(", ".GetBlob(", ".Submit(", ".ReconcileSubmission(",
		"provider.", "nats.", "http.", "net/http", "time.Now",
		"account.MutationLeaseToken != nil || account.MutationLeaseUntil != nil", "account.MutationEpoch != nextEpoch",
		"mailboxEntSubmissionFinalizationTerminalThread", "mailboxEntSubmissionFinalizationTerminalAttachments",
		"NewMailboxSubmissionSentEnvelopeV1", "RecipientResolutions: resolutions", "syncEventAddress(", "syncEventAddresses(", "completion.message.Projection",
		"mailattachment.MailMessageIDEQ(messageID)", "mailobject.IDIn(objectIDs...)",
	)
}
