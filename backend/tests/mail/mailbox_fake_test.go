package mail_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mail "github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailEngine = (*mail.FakeMailEngine)(nil)

func TestFakeMailEngineStateFamiliesCatchMissingInMemoryPorts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("domain-1", "account-1", "folder-1")

	caps, err := fake.Capabilities(ctx)
	if err != nil {
		t.Fatalf("Capabilities: %v", err)
	}
	if !caps.ExclusiveCustomFolderWriter || !caps.StableProjectionEnumeration || !caps.SubmissionReconciliation || !caps.DraftReconciliation {
		t.Fatalf("Capabilities = %#v, want all fake-supported capabilities", caps)
	}

	domain, err := fake.EnsureDomain(ctx, mail.DomainSpec{ProvisioningKey: "domain-key", Domain: "example.test", DesiredGeneration: 7, DesiredHash: "hash-7", Present: true, Enabled: true})
	if err != nil {
		t.Fatalf("EnsureDomain: %v", err)
	}
	if domain.DomainID != "domain-1" || domain.State != mail.EngineDomainVerified || !domain.Applied {
		t.Fatalf("EnsureDomain = %#v, want applied deterministic domain", domain)
	}
	gotDomain, err := fake.GetDomain(ctx, "example.test")
	if err != nil {
		t.Fatalf("GetDomain: %v", err)
	}
	if gotDomain.DomainID != "domain-1" {
		t.Fatalf("GetDomain id = %q, want domain-1", gotDomain.DomainID)
	}

	account, err := fake.EnsureAccount(ctx, mail.AccountSpec{ProvisioningKey: "account-key", LocalPart: "ada", Domain: "example.test", DesiredGeneration: 3, DesiredHash: "account-hash", Enabled: true})
	if err != nil {
		t.Fatalf("EnsureAccount: %v", err)
	}
	if account.AccountID != "account-1" || account.State != mail.EngineAccountActive || account.AppliedGeneration != 3 {
		t.Fatalf("EnsureAccount = %#v, want deterministic active account", account)
	}

	folder, err := fake.CreateFolder(ctx, "account-1", mail.FolderSpec{Name: "Projects"})
	if err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}
	if folder.FolderID != "folder-1" || folder.AccountID != "account-1" || folder.Name != "Projects" {
		t.Fatalf("CreateFolder = %#v, want deterministic account-scoped folder", folder)
	}
	if err := fake.RenameFolder(ctx, "account-1", "folder-1", "Projects 2026"); err != nil {
		t.Fatalf("RenameFolder: %v", err)
	}
	folders, err := fake.ListFolders(ctx, "account-1")
	if err != nil {
		t.Fatalf("ListFolders: %v", err)
	}
	if len(folders) != 1 || folders[0].Name != "Projects 2026" {
		t.Fatalf("ListFolders = %#v, want renamed folder", folders)
	}
	if err := fake.DeleteFolder(ctx, "account-1", "folder-1"); err != nil {
		t.Fatalf("DeleteFolder: %v", err)
	}
	folders, err = fake.ListFolders(ctx, "account-1")
	if err != nil {
		t.Fatalf("ListFolders after delete: %v", err)
	}
	if len(folders) != 0 || folders == nil {
		t.Fatalf("ListFolders after delete = %#v, want concrete empty slice", folders)
	}
}

func TestFakeMailEngineCopiesMessageAndBlobStateCatchAliasing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake()
	message := testMessage("account-a", "message-a", "version-a", "secret body")
	message.Projection.References = []string{"<parent@example.test>"}
	attachmentHash := make([]byte, 32)
	attachmentHash[0] = 1
	message.Projection.Attachments = []mail.EngineProjectionAttachment{{BlobID: "blob-a", Filename: "a.txt", MediaType: "text/plain", Disposition: mail.EngineAttachmentDispositionAttachment, SHA256: attachmentHash}}
	fake.PutMessage(message)
	fake.PutBlob(mail.EngineBlob{AccountID: "account-a", BlobID: "blob-a", MediaType: "text/plain", SizeBytes: 3, Bytes: []byte{1, 2, 3}})

	message.Body.PlainBody = "caller mutation"
	message.Projection.References[0] = "caller reference"
	got, err := fake.GetMessage(ctx, mail.EngineMessageReadRequest{AccountID: "account-a", MessageID: "message-a", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("GetMessage: %v", err)
	}
	if got.Body.PlainBody != "secret body" || got.Projection.References[0] != "<parent@example.test>" {
		t.Fatalf("GetMessage leaked caller mutation: %#v", got)
	}
	got.Body.PlainBody = "returned mutation"
	got.Projection.Attachments[0].SHA256[0] = 99
	again, err := fake.GetMessage(ctx, mail.EngineMessageReadRequest{AccountID: "account-a", MessageID: "message-a", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("GetMessage again: %v", err)
	}
	if again.Body.PlainBody != "secret body" || again.Projection.Attachments[0].SHA256[0] != 1 {
		t.Fatalf("GetMessage leaked returned mutation: %#v", again)
	}

	blob, err := fake.GetBlob(ctx, mail.EngineBlobReadRequest{AccountID: "account-a", BlobID: "blob-a", MaxBytes: 3})
	if err != nil {
		t.Fatalf("GetBlob: %v", err)
	}
	blob.Bytes[0] = 42
	againBlob, err := fake.GetBlob(ctx, mail.EngineBlobReadRequest{AccountID: "account-a", BlobID: "blob-a", MaxBytes: 3})
	if err != nil {
		t.Fatalf("GetBlob again: %v", err)
	}
	if got := againBlob.Bytes[0]; got != 1 {
		t.Fatalf("GetBlob leaked returned byte mutation: got %d, want 1", got)
	}
}

func TestFakeMailEngineStableDraftAndSubmitKeysCatchDuplicateWrites(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("draft-message", "draft-version", "submission", "sent-message")
	draft := testDraft("account-a", "draft-key", "", "", "mutation-1")

	saved, err := fake.SaveDraft(ctx, draft)
	if err != nil {
		t.Fatalf("SaveDraft create: %v", err)
	}
	if saved.Projection.MessageID != "draft-message" || saved.Projection.Version != "draft-version" || !saved.Projection.Flags.Draft {
		t.Fatalf("SaveDraft create = %#v, want deterministic draft projection", saved)
	}

	sent, err := fake.Submit(ctx, mail.EngineSubmission{AccountID: "account-a", DraftKey: "draft-key", MessageID: "draft-message", ExpectedVersion: "draft-version", SubmissionKey: "stable-submit", MaxMessageBytes: 4096})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	first, ok := sent.(mail.EngineSubmissionAccepted)
	if !ok {
		t.Fatalf("Submit result = %T, want EngineSubmissionAccepted", sent)
	}
	if first.SubmissionID != "submission" || first.MessageID != "sent-message" || len(first.SubmissionID) == 0 || len(first.SubmissionID) > mail.MaxEngineOpaqueIDBytes || len(first.MessageID) == 0 || len(first.MessageID) > mail.MaxEngineOpaqueIDBytes {
		t.Fatalf("Submit result = %#v, want two bounded deterministic opaque ids", first)
	}

	again, err := fake.Submit(ctx, mail.EngineSubmission{AccountID: "account-a", DraftKey: "draft-key", MessageID: "draft-message", ExpectedVersion: "draft-version", SubmissionKey: "stable-submit", MaxMessageBytes: 4096})
	if err != nil {
		t.Fatalf("Submit stable retry: %v", err)
	}
	second, ok := again.(mail.EngineSubmissionAccepted)
	if !ok || second != first {
		t.Fatalf("Submit stable retry = %#v, want exactly first visible result %#v", again, first)
	}
	resultMessage, err := fake.GetMessage(ctx, mail.EngineMessageReadRequest{AccountID: "account-a", MessageID: "sent-message", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("GetMessage sent result: %v", err)
	}
	if resultMessage.Projection.Origin != mail.EngineProjectionOriginConfirmedOwnerSubmission {
		t.Fatalf("sent projection origin = %q, want confirmed owner submission", resultMessage.Projection.Origin)
	}
}

func TestFakeMailEngineDraftVersionAndReconciliationCatchMutatingObservePaths(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("draft-message", "draft-version", "next-version")
	created, err := fake.SaveDraft(ctx, testDraft("account-a", "draft-key", "", "", "mutation-1"))
	if err != nil {
		t.Fatalf("SaveDraft create: %v", err)
	}
	_, err = fake.SaveDraft(ctx, testDraft("account-a", "draft-key", created.Projection.MessageID, "wrong-version", "mutation-2"))
	if !errors.Is(err, mail.ErrEngineConflict) {
		t.Fatalf("SaveDraft stale version error = %v, want conflict", err)
	}

	before, err := fake.ReconcileDraft(ctx, mail.EngineDraftReconciliation{AccountID: "account-a", DraftKey: "draft-key", MutationID: "mutation-1", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("ReconcileDraft: %v", err)
	}
	applied, ok := before.(mail.EngineDraftApplied)
	if !ok || applied.Message.Projection.Version != "draft-version" {
		t.Fatalf("ReconcileDraft = %#v, want original applied draft", before)
	}
	updated, err := fake.SaveDraft(ctx, testDraft("account-a", "draft-key", created.Projection.MessageID, created.Projection.Version, "mutation-2"))
	if err != nil {
		t.Fatalf("SaveDraft after reconcile: %v", err)
	}
	if updated.Projection.Version != "next-version" {
		t.Fatalf("SaveDraft after reconcile version = %q, want next-version; reconciliation must not mutate", updated.Projection.Version)
	}
}

func TestFakeMailEngineScriptedPreAndPostFailuresCatchWrongDispatchOrdering(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("draft-message", "draft-version")
	fake.EnqueueFault(mail.FakeMailFault{Operation: mail.FakeMailOperationSaveDraft, Kind: mail.EngineErrorUnavailable})
	_, err := fake.SaveDraft(ctx, testDraft("account-a", "draft-key", "", "", "mutation-1"))
	if !errors.Is(err, mail.ErrEngineUnavailable) {
		t.Fatalf("pre-dispatch SaveDraft error = %v, want unavailable", err)
	}
	pre, err := fake.ReconcileDraft(ctx, mail.EngineDraftReconciliation{AccountID: "account-a", DraftKey: "draft-key", MutationID: "mutation-1", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("ReconcileDraft after pre-dispatch fault: %v", err)
	}
	if _, ok := pre.(mail.EngineDraftAbsent); !ok {
		t.Fatalf("ReconcileDraft after pre-dispatch fault = %T, want EngineDraftAbsent", pre)
	}

	fake.EnqueueFault(mail.FakeMailFault{Operation: mail.FakeMailOperationSaveDraft, Kind: mail.EngineErrorIndeterminate, AfterDispatch: true})
	_, err = fake.SaveDraft(ctx, testDraft("account-a", "draft-key", "", "", "mutation-1"))
	if !errors.Is(err, mail.ErrEngineIndeterminate) {
		t.Fatalf("post-dispatch SaveDraft error = %v, want indeterminate", err)
	}
	post, err := fake.ReconcileDraft(ctx, mail.EngineDraftReconciliation{AccountID: "account-a", DraftKey: "draft-key", MutationID: "mutation-1", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("ReconcileDraft after lost response: %v", err)
	}
	if _, ok := post.(mail.EngineDraftApplied); !ok {
		t.Fatalf("ReconcileDraft after lost response = %T, want EngineDraftApplied", post)
	}
}

func TestFakeMailEngineCrossAccountIsolationCatchForeignDisclosureAndMutation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake()
	fake.PutMessage(testMessage("account-a", "message-a", "version-a", "private body"))
	fake.PutBlob(mail.EngineBlob{AccountID: "account-a", BlobID: "blob-a", MediaType: "text/plain", SizeBytes: 3, Bytes: []byte{1, 2, 3}})

	_, err := fake.GetMessage(ctx, mail.EngineMessageReadRequest{AccountID: "account-b", MessageID: "message-a", MaxContentBytes: 4096})
	if !errors.Is(err, mail.ErrEngineNotFound) {
		t.Fatalf("GetMessage foreign id error = %v, want undisclosing not found", err)
	}
	_, err = fake.GetBlob(ctx, mail.EngineBlobReadRequest{AccountID: "account-b", BlobID: "blob-a", MaxBytes: 3})
	if !errors.Is(err, mail.ErrEngineNotFound) {
		t.Fatalf("GetBlob foreign id error = %v, want undisclosing not found", err)
	}
	if err := fake.SetFlags(ctx, "account-b", []string{"message-a"}, mail.FlagPatch{Seen: mail.FlagValue{Present: true, Value: true}}); err != nil && !errors.Is(err, mail.ErrEngineNotFound) {
		t.Fatalf("SetFlags foreign id error = %v, want safe no-op or not found", err)
	}
	owner, err := fake.GetMessage(ctx, mail.EngineMessageReadRequest{AccountID: "account-a", MessageID: "message-a", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("GetMessage owner after foreign mutation: %v", err)
	}
	if owner.Projection.Flags.Seen {
		t.Fatal("foreign SetFlags mutated owner message")
	}
}

func TestFakeMailEngineScriptsPagesAndMismatchedProjectionCatchSyncAndFinalizationBugs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake()
	projection := testMessage("account-a", "message-a", "version-a", "body").Projection
	fake.ScriptChanges(mail.FakeMailChangeScript{AccountID: "account-a", FromCursor: "cursor-0", Page: mail.EngineChangePage{AccountID: "account-a", FromCursor: "cursor-0", NextCursor: "cursor-1", Bounds: mail.EnginePageBounds{Requested: 1, Maximum: mail.MaxEnginePageItems, Returned: 1}, CreatedOrUpdated: []mail.EngineMessageProjection{projection}}})
	page, err := fake.Changes(ctx, mail.EngineChangeRequest{AccountID: "account-a", FromCursor: "cursor-0", Page: mail.EnginePageRequest{Limit: 1}})
	if err != nil {
		t.Fatalf("Changes scripted page: %v", err)
	}
	page.CreatedOrUpdated[0].Subject = "caller mutation"
	if page.NextCursor != "cursor-1" || page.Bounds.Returned != 1 {
		t.Fatalf("Changes page = %#v, want literal scripted advance", page)
	}
	again, err := fake.Changes(ctx, mail.EngineChangeRequest{AccountID: "account-a", FromCursor: "cursor-0", Page: mail.EnginePageRequest{Limit: 1}})
	if err != nil {
		t.Fatalf("Changes scripted page again: %v", err)
	}
	if again.CreatedOrUpdated[0].Subject != "subject" {
		t.Fatalf("Changes returned aliased projection: %#v", again)
	}
	_, err = fake.Changes(ctx, mail.EngineChangeRequest{AccountID: "account-a", FromCursor: "cursor-secret", Page: mail.EnginePageRequest{Limit: 1}})
	if !errors.Is(err, mail.ErrEngineInvalidCursor) || strings.Contains(fmt.Sprint(err), "cursor-secret") {
		t.Fatalf("Changes invalid cursor error = %v, want typed non-leaking invalid cursor", err)
	}
	fake.ScriptChanges(mail.FakeMailChangeScript{AccountID: "account-a", FromCursor: "cursor-1", Fault: mail.EngineErrorUnavailable})
	_, err = fake.Changes(ctx, mail.EngineChangeRequest{AccountID: "account-a", FromCursor: "cursor-1", Page: mail.EnginePageRequest{Limit: 1}})
	if !errors.Is(err, mail.ErrEngineUnavailable) || strings.Contains(fmt.Sprint(err), "cursor-1") {
		t.Fatalf("Changes transient failure = %v, want typed non-leaking unavailable", err)
	}

	fake.ScriptEnumeration(mail.FakeMailEnumerationScript{AccountID: "account-a", BaseCursor: "", Checkpoint: "", Page: mail.EngineEnumerationPage{AccountID: "account-a", BaseCursor: "base-1", Checkpoint: "", NextCheckpoint: "checkpoint-1", Complete: false, Bounds: mail.EnginePageBounds{Requested: 1, Maximum: mail.MaxEnginePageItems, Returned: 1}, Messages: []mail.EngineMessageProjection{projection}}})
	fake.ScriptEnumeration(mail.FakeMailEnumerationScript{AccountID: "account-a", BaseCursor: "base-1", Checkpoint: "checkpoint-1", Fault: mail.EngineErrorRestartRequired})
	enumeration, err := fake.EnumerateProjection(ctx, mail.EngineEnumerationRequest{AccountID: "account-a", Page: mail.EnginePageRequest{Limit: 1}})
	if err != nil {
		t.Fatalf("EnumerateProjection initial: %v", err)
	}
	if enumeration.BaseCursor != "base-1" || enumeration.NextCheckpoint != "checkpoint-1" || enumeration.Complete {
		t.Fatalf("EnumerateProjection initial = %#v, want literal base anchored page", enumeration)
	}
	_, err = fake.EnumerateProjection(ctx, mail.EngineEnumerationRequest{AccountID: "account-a", BaseCursor: "base-1", Checkpoint: "checkpoint-1", Page: mail.EnginePageRequest{Limit: 1}})
	if !errors.Is(err, mail.ErrEngineRestartRequired) || strings.Contains(fmt.Sprint(err), "checkpoint-1") {
		t.Fatalf("EnumerateProjection restart = %v, want typed non-leaking restart", err)
	}

	fake.PutMessage(testMessage("account-a", "message-a", "version-a", "real body"))
	mismatchProjection := testMessage("account-b", "other-message", "version-b", "mismatch body").Projection
	fake.ScriptMessageProjection("account-a", "message-a", mismatchProjection)
	mismatched, err := fake.GetMessage(ctx, mail.EngineMessageReadRequest{AccountID: "account-a", MessageID: "message-a", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("GetMessage scripted mismatch: %v", err)
	}
	if mismatched.Projection.AccountID != "account-b" || mismatched.Projection.MessageID != "other-message" {
		t.Fatalf("GetMessage scripted mismatch = %#v, want finalization oracle mismatch", mismatched.Projection)
	}
}

func TestFakeMailEngineSieveSpamAndConcurrentCallsCatchUnlockedState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake()
	fake.PutMessage(testMessage("account-a", "message-a", "version-a", "body"))

	sieve, err := fake.UpsertSieve(ctx, "account-a", mail.SieveScript{Name: "hub", Generation: 9, Hash: "sieve-hash"})
	if err != nil {
		t.Fatalf("UpsertSieve: %v", err)
	}
	if !sieve.Active || sieve.State != mail.EngineSieveApplied || sieve.Generation != 9 {
		t.Fatalf("UpsertSieve = %#v, want fixed applied state", sieve)
	}
	gotSieve, err := fake.GetSieve(ctx, "account-a", "hub")
	if err != nil {
		t.Fatalf("GetSieve: %v", err)
	}
	if gotSieve.Hash != "sieve-hash" {
		t.Fatalf("GetSieve = %#v, want saved fixed state", gotSieve)
	}

	training, err := fake.ClassifyMessages(ctx, mail.EngineSpamTraining{AccountID: "account-a", ReceiptID: "receipt-a", CommandID: "command-a", Kind: mail.EngineSpamTrainingSpam, Samples: []mail.EngineSpamTrainingSample{{MessageID: "message-a"}, {MessageID: "missing"}}})
	if err != nil {
		t.Fatalf("ClassifyMessages: %v", err)
	}
	if len(training.Outcomes) != 2 || training.Outcomes[0].Status != mail.EngineSpamTrainingAccepted || training.Outcomes[1].Status != mail.EngineSpamTrainingRejected {
		t.Fatalf("ClassifyMessages = %#v, want accepted known and rejected absent outcomes", training)
	}

	var wait sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for turn := 0; turn < 32; turn++ {
				_, _ = fake.ListMessages(ctx, mail.EngineMessageQuery{AccountID: "account-a", Page: mail.EnginePageRequest{Limit: 1}})
				_ = fake.SetFlags(ctx, "account-a", []string{"message-a"}, mail.FlagPatch{Seen: mail.FlagValue{Present: true, Value: true}})
			}
		}()
	}
	wait.Wait()
	health, err := fake.Health(ctx)
	if err != nil || !health.Available || health.State != mail.EngineHealthReady {
		t.Fatalf("Health = %#v, %v; want ready deterministic fake", health, err)
	}
}

func TestFakeMailEngineSaveDraftRejectsBeforeRetainingAttachmentBlob(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("message-id", "version-id")
	draft := testDraft("account-a", "draft-key", "", "", "mutation-1")
	draft.MaxContentBytes = 1
	draft.Message.Attachments = []mail.EngineDraftAttachment{{AttachmentKey: "residual-blob", Filename: "a.txt", MediaType: "text/plain", Disposition: mail.EngineAttachmentDispositionAttachment, Bytes: []byte{1, 2}}}

	_, err := fake.SaveDraft(ctx, draft)
	if !errors.Is(err, mail.ErrEnginePayloadTooLarge) {
		t.Fatalf("SaveDraft over budget error = %v, want payload too large", err)
	}
	_, err = fake.GetBlob(ctx, mail.EngineBlobReadRequest{AccountID: "account-a", BlobID: "residual-blob", MaxBytes: 2})
	if !errors.Is(err, mail.ErrEngineNotFound) {
		t.Fatalf("GetBlob after rejected SaveDraft = %v, want no residual blob", err)
	}
	reconciled, err := fake.ReconcileDraft(ctx, mail.EngineDraftReconciliation{AccountID: "account-a", DraftKey: "draft-key", MutationID: "mutation-1", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("ReconcileDraft after rejected SaveDraft: %v", err)
	}
	if _, ok := reconciled.(mail.EngineDraftAbsent); !ok {
		t.Fatalf("ReconcileDraft after rejected SaveDraft = %T, want EngineDraftAbsent", reconciled)
	}
}

func TestFakeMailEngineReconcileDraftRequiresExactWriteTupleAndCanBeIndeterminate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("message-id", "version-id")
	_, err := fake.SaveDraft(ctx, testDraft("account-a", "draft-key", "", "", "mutation-1"))
	if err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}

	stale, err := fake.ReconcileDraft(ctx, mail.EngineDraftReconciliation{AccountID: "account-a", DraftKey: "draft-key", MessageID: "message-id", PreviousVersion: "version-id", MutationID: "mutation-1", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("ReconcileDraft stale tuple: %v", err)
	}
	if _, ok := stale.(mail.EngineDraftIndeterminate); !ok {
		t.Fatalf("ReconcileDraft stale tuple = %T, want EngineDraftIndeterminate", stale)
	}

	fake.EnqueueFault(mail.FakeMailFault{Operation: mail.FakeMailOperationReconcileDraft, Kind: mail.EngineErrorIndeterminate})
	indeterminate, err := fake.ReconcileDraft(ctx, mail.EngineDraftReconciliation{AccountID: "account-a", DraftKey: "draft-key", MutationID: "mutation-1", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("ReconcileDraft scripted indeterminate: %v", err)
	}
	if _, ok := indeterminate.(mail.EngineDraftIndeterminate); !ok {
		t.Fatalf("ReconcileDraft scripted indeterminate = %T, want EngineDraftIndeterminate", indeterminate)
	}
}

func TestFakeMailEngineReconcileSubmissionUsesExactBindingAndIndeterminateResult(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("draft-message", "draft-version", "submission-id", "sent-message")
	draft, err := fake.SaveDraft(ctx, testDraft("account-a", "draft-key", "", "", "mutation-1"))
	if err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}
	submission := mail.EngineSubmission{AccountID: "account-a", DraftKey: "draft-key", MessageID: draft.Projection.MessageID, ExpectedVersion: draft.Projection.Version, SubmissionKey: "submit-key", MaxMessageBytes: 4096}
	_, err = fake.Submit(ctx, submission)
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	accepted, err := fake.ReconcileSubmission(ctx, submission)
	if err != nil {
		t.Fatalf("ReconcileSubmission applied: %v", err)
	}
	if _, ok := accepted.(mail.EngineSubmissionAccepted); !ok {
		t.Fatalf("ReconcileSubmission applied = %T, want EngineSubmissionAccepted", accepted)
	}
	foreignSubmission := submission
	foreignSubmission.AccountID = "account-b"
	absent, err := fake.ReconcileSubmission(ctx, foreignSubmission)
	if err != nil {
		t.Fatalf("ReconcileSubmission foreign account: %v", err)
	}
	if _, ok := absent.(mail.EngineSubmissionAbsent); !ok {
		t.Fatalf("ReconcileSubmission foreign account = %T, want EngineSubmissionAbsent", absent)
	}
	fake.EnqueueFault(mail.FakeMailFault{Operation: mail.FakeMailOperationReconcileSubmission, Kind: mail.EngineErrorIndeterminate})
	indeterminate, err := fake.ReconcileSubmission(ctx, submission)
	if err != nil {
		t.Fatalf("ReconcileSubmission scripted indeterminate: %v", err)
	}
	if _, ok := indeterminate.(mail.EngineSubmissionIndeterminate); !ok {
		t.Fatalf("ReconcileSubmission scripted indeterminate = %T, want EngineSubmissionIndeterminate", indeterminate)
	}
}

func TestFakeMailEngineTerminalEnumerationAnchorFeedsNaturalChanges(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake()
	fake.PutMessage(testMessage("account-a", "message-1", "version-1", "body"))

	terminal, err := fake.EnumerateProjection(ctx, mail.EngineEnumerationRequest{AccountID: "account-a", Page: mail.EnginePageRequest{Limit: 1}})
	if err != nil {
		t.Fatalf("EnumerateProjection terminal: %v", err)
	}
	if !terminal.Complete || terminal.NextCheckpoint != "" || terminal.BaseCursor == "" || terminal.Bounds.Returned != 1 {
		t.Fatalf("EnumerateProjection terminal = %#v, want terminal page with reusable base anchor", terminal)
	}
	fake.PutMessage(testMessage("account-a", "message-2", "version-2", "new body"))
	changes, err := fake.Changes(ctx, mail.EngineChangeRequest{AccountID: "account-a", FromCursor: terminal.BaseCursor, Page: mail.EnginePageRequest{Limit: 1}})
	if err != nil {
		t.Fatalf("Changes from terminal base anchor: %v", err)
	}
	if changes.Bounds.Returned != 1 || changes.CreatedOrUpdated[0].MessageID != "message-2" || changes.NextCursor != "" {
		t.Fatalf("Changes from terminal base anchor = %#v, want only post-anchor message", changes)
	}
}

func TestFakeMailEngineFaultQueueIsStrictFIFOAndPostDispatchIsVisible(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("message-id", "version-id")
	fake.EnqueueFault(mail.FakeMailFault{Operation: mail.FakeMailOperationSetFlags, Kind: mail.EngineErrorUnavailable})
	fake.EnqueueFault(mail.FakeMailFault{Operation: mail.FakeMailOperationSaveDraft, Kind: mail.EngineErrorIndeterminate, AfterDispatch: true})

	created, err := fake.SaveDraft(ctx, testDraft("account-a", "draft-key", "", "", "mutation-1"))
	if err != nil {
		t.Fatalf("SaveDraft must not skip earlier queued SetFlags fault: %v", err)
	}
	if err := fake.SetFlags(ctx, "account-a", []string{created.Projection.MessageID}, mail.FlagPatch{Seen: mail.FlagValue{Present: true, Value: true}}); !errors.Is(err, mail.ErrEngineUnavailable) {
		t.Fatalf("SetFlags first queued fault = %v, want unavailable", err)
	}
	_, err = fake.SaveDraft(ctx, testDraft("account-a", "draft-key", "", "", "mutation-1"))
	if !errors.Is(err, mail.ErrEngineIndeterminate) {
		t.Fatalf("SaveDraft post-dispatch queued fault = %v, want indeterminate", err)
	}
}

func TestFakeMailEngineRejectsDuplicateGeneratedIDWithoutClobberOrLockingCallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var fake *mail.FakeMailEngine
	calls := 0
	fake = mail.NewFakeMailEngine(mail.FakeMailEngineConfig{NextID: func() string {
		calls++
		if calls == 1 {
			_, _ = fake.Health(ctx)
		}
		return "duplicate-id"
	}, Now: func() time.Time { return time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC) }})
	first, err := fake.EnsureDomain(ctx, mail.DomainSpec{ProvisioningKey: "domain-key", Domain: "example.test", Present: true, Enabled: true})
	if err != nil {
		t.Fatalf("EnsureDomain first deterministic id: %v", err)
	}
	_, err = fake.EnsureAccount(ctx, mail.AccountSpec{ProvisioningKey: "account-key", LocalPart: "ada", Domain: "example.test", Enabled: true})
	if !errors.Is(err, mail.ErrEngineConflict) {
		t.Fatalf("EnsureAccount duplicate generated id error = %v, want conflict", err)
	}
	still, err := fake.GetDomain(ctx, "example.test")
	if err != nil || still.DomainID != first.DomainID {
		t.Fatalf("GetDomain after duplicate id = %#v, %v; want original retained domain", still, err)
	}
}

func TestFakeMailEngineRejectsControlKeyCollisionsAndInvalidBoundsWithoutDisclosure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake()
	fake.PutMessage(testMessage("account-a", "message-a", "version-a", "body"))
	fake.ScriptChanges(mail.FakeMailChangeScript{AccountID: "account-a", FromCursor: "cursor-a", Page: mail.EngineChangePage{AccountID: "account-a", FromCursor: "cursor-a", Bounds: mail.EnginePageBounds{Requested: 1, Maximum: mail.MaxEnginePageItems, Returned: 0}, CreatedOrUpdated: []mail.EngineMessageProjection{}, DestroyedIDs: []string{}}})

	_, err := fake.Changes(ctx, mail.EngineChangeRequest{AccountID: "account-a\x00other", FromCursor: "cursor-a", Page: mail.EnginePageRequest{Limit: 1}})
	if !errors.Is(err, mail.ErrEngineRejectedPolicy) || strings.Contains(fmt.Sprint(err), "account-a") {
		t.Fatalf("Changes control account error = %v, want safe rejected policy", err)
	}
	_, err = fake.Changes(ctx, mail.EngineChangeRequest{AccountID: "account-a", FromCursor: "cursor-a\x00other", Page: mail.EnginePageRequest{Limit: 1}})
	if !errors.Is(err, mail.ErrEngineInvalidCursor) || strings.Contains(fmt.Sprint(err), "cursor-a") {
		t.Fatalf("Changes control cursor error = %v, want safe invalid cursor", err)
	}
	_, err = fake.ListMessages(ctx, mail.EngineMessageQuery{AccountID: "account-a", Page: mail.EnginePageRequest{Limit: mail.MaxEnginePageItems + 1}})
	if !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("ListMessages invalid limit error = %v, want rejected policy", err)
	}
	got, err := fake.GetMessage(ctx, mail.EngineMessageReadRequest{AccountID: "account-a", MessageID: "message-a", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("GetMessage after invalid inputs: %v", err)
	}
	*got.Projection.ReceivedAt = got.Projection.ReceivedAt.Add(time.Hour)
	again, err := fake.GetMessage(ctx, mail.EngineMessageReadRequest{AccountID: "account-a", MessageID: "message-a", MaxContentBytes: 4096})
	if err != nil || !again.Projection.ReceivedAt.Equal(time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("GetMessage timestamp deep copy = %#v, %v; want original timestamp", again.Projection.ReceivedAt, err)
	}
}

func TestFakeMailEngineMoveKeywordsAndDeleteRejectInvalidMutationsWithoutPartialState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("folder-id")
	fake.PutMessage(testMessage("account-a", "message-a", "version-a", "body"))
	folder, err := fake.CreateFolder(ctx, "account-a", mail.FolderSpec{Name: "Archive"})
	if err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}
	if err := fake.MoveMessages(ctx, "account-a", []string{"message-a"}, folder.FolderID); err != nil {
		t.Fatalf("MoveMessages: %v", err)
	}
	if err := fake.SetKeywords(ctx, "account-a", []string{"message-a"}, mail.KeywordPatch{Add: []mail.EngineKeywordMarker{"$label1", "$label2"}}); err != nil {
		t.Fatalf("SetKeywords: %v", err)
	}
	tooMany := make([]mail.EngineKeywordMarker, mail.MaxEngineKeywordItems+1)
	for index := range tooMany {
		tooMany[index] = mail.EngineKeywordMarker(fmt.Sprintf("$label-%d", index))
	}
	if err := fake.SetKeywords(ctx, "account-a", []string{"message-a"}, mail.KeywordPatch{Add: tooMany}); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("SetKeywords over limit error = %v, want rejected policy", err)
	}
	message, err := fake.GetMessage(ctx, mail.EngineMessageReadRequest{AccountID: "account-a", MessageID: "message-a", MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("GetMessage after MoveMessages/SetKeywords: %v", err)
	}
	if len(message.Projection.FolderIDs) != 1 || message.Projection.FolderIDs[0] != "folder-id" || len(message.Projection.ReservedKeywords) != 2 {
		t.Fatalf("GetMessage after retained mutations = %#v", message.Projection)
	}
	if err := fake.DeleteMessages(ctx, "account-a", []string{"message-a"}, true); err != nil {
		t.Fatalf("DeleteMessages permanent: %v", err)
	}
	_, err = fake.GetMessage(ctx, mail.EngineMessageReadRequest{AccountID: "account-a", MessageID: "message-a", MaxContentBytes: 4096})
	if !errors.Is(err, mail.ErrEngineNotFound) {
		t.Fatalf("GetMessage after DeleteMessages = %v, want not found", err)
	}
}

func TestFakeMailEngineNaturalListMessagesAdvancesLiteralCursorsToTerminalPage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake()
	fake.PutMessage(testMessage("account-a", "message-3", "version-3", "third"))
	fake.PutMessage(testMessage("account-a", "message-1", "version-1", "first"))
	fake.PutMessage(testMessage("account-a", "message-2", "version-2", "second"))

	first, err := fake.ListMessages(ctx, mail.EngineMessageQuery{AccountID: "account-a", Page: mail.EnginePageRequest{Limit: 2}})
	if err != nil {
		t.Fatalf("ListMessages first page: %v", err)
	}
	if first.Cursor != "list:0" || first.NextCursor != "list:2" || first.Bounds != (mail.EnginePageBounds{Requested: 2, Maximum: mail.MaxEnginePageItems, Returned: 2}) || len(first.Messages) != 2 || first.Messages[0].MessageID != "message-1" || first.Messages[1].MessageID != "message-2" {
		t.Fatalf("ListMessages first page = %#v, want literal sorted first page", first)
	}
	terminal, err := fake.ListMessages(ctx, mail.EngineMessageQuery{AccountID: "account-a", Cursor: first.NextCursor, Page: mail.EnginePageRequest{Limit: 2}})
	if err != nil {
		t.Fatalf("ListMessages terminal page: %v", err)
	}
	if terminal.Cursor != "list:2" || terminal.NextCursor != "" || terminal.Bounds != (mail.EnginePageBounds{Requested: 2, Maximum: mail.MaxEnginePageItems, Returned: 1}) || len(terminal.Messages) != 1 || terminal.Messages[0].MessageID != "message-3" {
		t.Fatalf("ListMessages terminal page = %#v, want literal terminal page", terminal)
	}
}

func TestFakeMailEngineDefaultIDsRemainUniqueUnderConcurrentAccountCreation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := mail.NewFakeMailEngine(mail.FakeMailEngineConfig{Now: func() time.Time { return time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC) }})
	const workers = 16
	start := make(chan struct{})
	results := make(chan mail.EngineAccount, workers)
	errorsByWorker := make(chan error, workers)
	var wait sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			<-start
			account, err := fake.EnsureAccount(ctx, mail.AccountSpec{ProvisioningKey: fmt.Sprintf("key-%d", index), LocalPart: fmt.Sprintf("user%d", index), Domain: "example.test", Enabled: true})
			if err != nil {
				errorsByWorker <- err
				return
			}
			results <- account
		}(worker)
	}
	close(start)
	wait.Wait()
	close(results)
	close(errorsByWorker)
	for err := range errorsByWorker {
		t.Fatalf("EnsureAccount concurrent default id: %v", err)
	}
	ids := make(map[string]struct{}, workers)
	for account := range results {
		if _, duplicate := ids[account.AccountID]; duplicate {
			t.Fatalf("EnsureAccount generated duplicate id %q", account.AccountID)
		}
		ids[account.AccountID] = struct{}{}
	}
	if len(ids) != workers {
		t.Fatalf("EnsureAccount distinct ids = %d, want %d", len(ids), workers)
	}
}

func TestFakeMailEngineWriteDecisionsDoNotConsumeCallbacksBeforeStableOutcomes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ids := []string{"draft-message", "draft-version", "submission-id", "sent-message"}
	nextCalls := 0
	nowCalls := 0
	fake := mail.NewFakeMailEngine(mail.FakeMailEngineConfig{
		NextID: func() string {
			nextCalls++
			if nextCalls > len(ids) {
				return ""
			}
			return ids[nextCalls-1]
		},
		Now: func() time.Time {
			nowCalls++
			return time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
		},
	})
	draftRequest := testDraft("account-a", "draft-key", "", "", "mutation-1")
	fake.EnqueueFault(mail.FakeMailFault{Operation: mail.FakeMailOperationSaveDraft, Kind: mail.EngineErrorUnavailable})
	if _, err := fake.SaveDraft(ctx, draftRequest); !errors.Is(err, mail.ErrEngineUnavailable) {
		t.Fatalf("SaveDraft pre-dispatch fault = %v, want unavailable", err)
	}
	if nextCalls != 0 || nowCalls != 0 {
		t.Fatalf("SaveDraft pre-fault callbacks = ids:%d now:%d, want 0:0", nextCalls, nowCalls)
	}
	draft, err := fake.SaveDraft(ctx, draftRequest)
	if err != nil {
		t.Fatalf("SaveDraft applied: %v", err)
	}
	if nextCalls != 2 || nowCalls != 1 {
		t.Fatalf("SaveDraft applied callbacks = ids:%d now:%d, want 2:1", nextCalls, nowCalls)
	}
	replay, err := fake.SaveDraft(ctx, draftRequest)
	if err != nil || replay.Projection.MessageID != draft.Projection.MessageID || nextCalls != 2 || nowCalls != 1 {
		t.Fatalf("SaveDraft replay = %#v, %v, ids:%d now:%d; want stable zero-source replay", replay, err, nextCalls, nowCalls)
	}
	stale := testDraft("account-a", "draft-key", draft.Projection.MessageID, "stale-version", "mutation-2")
	if _, err := fake.SaveDraft(ctx, stale); !errors.Is(err, mail.ErrEngineConflict) || nextCalls != 2 || nowCalls != 1 {
		t.Fatalf("SaveDraft stale version = %v, ids:%d now:%d; want conflict without consuming next version", err, nextCalls, nowCalls)
	}
	submission := mail.EngineSubmission{AccountID: "account-a", DraftKey: "draft-key", MessageID: draft.Projection.MessageID, ExpectedVersion: draft.Projection.Version, SubmissionKey: "submit-key", MaxMessageBytes: 4096}
	sent, err := fake.Submit(ctx, submission)
	if err != nil || nextCalls != 4 || nowCalls != 1 {
		t.Fatalf("Submit applied = %#v, %v, ids:%d now:%d; want 4:1", sent, err, nextCalls, nowCalls)
	}
	repeated, err := fake.Submit(ctx, submission)
	if err != nil || nextCalls != 4 || nowCalls != 1 || fmt.Sprint(repeated) != fmt.Sprint(sent) {
		t.Fatalf("Submit replay = %#v, %v, ids:%d now:%d; want stable zero-source result", repeated, err, nextCalls, nowCalls)
	}
}

func TestFakeMailEngineRejectsCrossAccountScriptPagesWithoutDisclosure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake()
	foreign := testMessage("account-b", "foreign-message", "foreign-version", "foreign-body").Projection
	fake.ScriptChanges(mail.FakeMailChangeScript{AccountID: "account-a", FromCursor: "script-cursor", Page: mail.EngineChangePage{AccountID: "account-a", FromCursor: "script-cursor", Bounds: mail.EnginePageBounds{Requested: 1, Maximum: mail.MaxEnginePageItems, Returned: 1}, CreatedOrUpdated: []mail.EngineMessageProjection{foreign}}})
	_, err := fake.Changes(ctx, mail.EngineChangeRequest{AccountID: "account-a", FromCursor: "script-cursor", Page: mail.EnginePageRequest{Limit: 1}})
	if !errors.Is(err, mail.ErrEngineInvalidCursor) || strings.Contains(fmt.Sprint(err), "account-b") || strings.Contains(fmt.Sprint(err), "foreign-message") {
		t.Fatalf("Changes rejected foreign script = %v, want safe invalid cursor without disclosure", err)
	}
	fake.ScriptEnumeration(mail.FakeMailEnumerationScript{AccountID: "account-a", BaseCursor: "base-a", Page: mail.EngineEnumerationPage{AccountID: "account-a", BaseCursor: "base-a", Complete: true, Bounds: mail.EnginePageBounds{Requested: 1, Maximum: mail.MaxEnginePageItems, Returned: 1}, Messages: []mail.EngineMessageProjection{foreign}}})
	_, err = fake.EnumerateProjection(ctx, mail.EngineEnumerationRequest{AccountID: "account-a", BaseCursor: "base-a", Page: mail.EnginePageRequest{Limit: 1}})
	if !errors.Is(err, mail.ErrEngineRestartRequired) || strings.Contains(fmt.Sprint(err), "account-b") || strings.Contains(fmt.Sprint(err), "foreign-message") {
		t.Fatalf("EnumerateProjection rejected foreign script = %v, want safe restart without disclosure", err)
	}
}

func TestFakeMailEngineRejectsNoncanonicalAddressesBeforeSourceOrRetention(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	nextCalls := 0
	fake := mail.NewFakeMailEngine(mail.FakeMailEngineConfig{NextID: func() string {
		nextCalls++
		if nextCalls == 1 {
			return "message-id"
		}
		return "version-id"
	}, Now: func() time.Time { return time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC) }})
	noncanonical := testDraft("account-a", "draft-key", "", "", "mutation-1")
	noncanonical.Message.From.Email = "Sender@example.test"
	if _, err := fake.SaveDraft(ctx, noncanonical); !errors.Is(err, mail.ErrEngineRejectedPolicy) || nextCalls != 0 {
		t.Fatalf("SaveDraft noncanonical address = %v, source calls:%d; want rejected policy before source", err, nextCalls)
	}
	invalid := testDraft("account-a", "draft-key", "", "", "mutation-1")
	invalid.Message.To[0].Email = "not an address"
	if _, err := fake.SaveDraft(ctx, invalid); !errors.Is(err, mail.ErrEngineRejectedPolicy) || nextCalls != 0 {
		t.Fatalf("SaveDraft invalid address = %v, source calls:%d; want rejected policy before source", err, nextCalls)
	}
	if _, err := fake.SaveDraft(ctx, testDraft("account-a", "draft-key", "", "", "mutation-1")); err != nil || nextCalls != 2 {
		t.Fatalf("SaveDraft canonical address after rejection = %v, source calls:%d; want retained first source values", err, nextCalls)
	}
}

func TestFakeMailEngineBoundsGeneratedAccountIdentityIDsBeforeRetention(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ids := []string{strings.Repeat("a", mail.MaxEngineOpaqueIDBytes), "account-id"}
	index := 0
	fake := mail.NewFakeMailEngine(mail.FakeMailEngineConfig{NextID: func() string {
		id := ids[index]
		index++
		return id
	}, Now: func() time.Time { return time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC) }})
	spec := mail.AccountSpec{ProvisioningKey: "account-key", LocalPart: "ada", Domain: "example.test", Enabled: true}
	if _, err := fake.EnsureAccount(ctx, spec); !errors.Is(err, mail.ErrEngineRejectedPolicy) || index != 1 {
		t.Fatalf("EnsureAccount max generated id = %v, source calls:%d; want bounded rejection", err, index)
	}
	account, err := fake.EnsureAccount(ctx, spec)
	if err != nil || account.AccountID != "account-id" || len(account.PrimaryIdentityID) > mail.MaxEngineOpaqueIDBytes {
		t.Fatalf("EnsureAccount after bounded rejection = %#v, %v; want clean usable account", account, err)
	}
}

func TestFakeMailEngineEnumerationResumeTuplesRestartWithoutPolicyRejection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake()
	fake.PutMessage(testMessage("account-a", "message-1", "version-1", "body-1"))
	fake.PutMessage(testMessage("account-a", "message-2", "version-2", "body-2"))
	first, err := fake.EnumerateProjection(ctx, mail.EngineEnumerationRequest{AccountID: "account-a", Page: mail.EnginePageRequest{Limit: 1}})
	if err != nil || first.BaseCursor == "" || first.NextCheckpoint == "" {
		t.Fatalf("EnumerateProjection initial checkpoint = %#v, %v; want literal resumable page", first, err)
	}
	for name, request := range map[string]mail.EngineEnumerationRequest{
		"base_without_checkpoint": {AccountID: "account-a", BaseCursor: first.BaseCursor, Page: mail.EnginePageRequest{Limit: 1}},
		"checkpoint_without_base": {AccountID: "account-a", Checkpoint: first.NextCheckpoint, Page: mail.EnginePageRequest{Limit: 1}},
		"checkpoint_for_other_base": {AccountID: "account-a", BaseCursor: "sync:0", Checkpoint: first.NextCheckpoint, Page: mail.EnginePageRequest{Limit: 1}},
	} {
		if _, err := fake.EnumerateProjection(ctx, request); !errors.Is(err, mail.ErrEngineRestartRequired) || errors.Is(err, mail.ErrEngineRejectedPolicy) {
			t.Fatalf("EnumerateProjection %s = %v, want restart required only", name, err)
		}
	}
	_, err = fake.EnumerateProjection(ctx, mail.EngineEnumerationRequest{AccountID: "account-a", BaseCursor: "bad\x00cursor", Checkpoint: "checkpoint-a", Page: mail.EnginePageRequest{Limit: 1}})
	if !errors.Is(err, mail.ErrEngineRejectedPolicy) || errors.Is(err, mail.ErrEngineRestartRequired) || strings.Contains(fmt.Sprint(err), "bad") {
		t.Fatalf("EnumerateProjection malformed resume = %v, want safe rejected policy only", err)
	}
}

// This intentionally exercises only the documented source-free Health re-entry.
// Calling a source-consuming operation from either callback is forbidden by config contract.
func TestFakeMailEngineSerializesInjectedSourcesAndPermitsHealthReentry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var fake *mail.FakeMailEngine
	var active int32
	var overlap int32
	var healthFailure int32
	var sequence uint64
	enter := func() {
		if atomic.AddInt32(&active, 1) != 1 {
			atomic.StoreInt32(&overlap, 1)
		}
		health, err := fake.Health(ctx)
		if err != nil || !health.Available || health.State != mail.EngineHealthReady {
			atomic.StoreInt32(&healthFailure, 1)
		}
		atomic.AddInt32(&active, -1)
	}
	fake = mail.NewFakeMailEngine(mail.FakeMailEngineConfig{
		NextID: func() string {
			enter()
			return fmt.Sprintf("source-%d", atomic.AddUint64(&sequence, 1))
		},
		Now: func() time.Time {
			enter()
			return time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
		},
	})
	const workers = 8
	start := make(chan struct{})
	errorsByWorker := make(chan error, workers)
	var wait sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			<-start
			_, err := fake.SaveDraft(ctx, testDraft("account-a", fmt.Sprintf("draft-%d", index), "", "", fmt.Sprintf("mutation-%d", index)))
			if err != nil {
				errorsByWorker <- err
			}
		}(worker)
	}
	close(start)
	wait.Wait()
	close(errorsByWorker)
	for err := range errorsByWorker {
		t.Fatalf("SaveDraft with reentrant source: %v", err)
	}
	if atomic.LoadInt32(&overlap) != 0 || atomic.LoadInt32(&active) != 0 || atomic.LoadInt32(&healthFailure) != 0 || atomic.LoadUint64(&sequence) != workers*3 {
		t.Fatalf("injected source serialization overlap:%d active:%d health:%d calls:%d, want 0:0:0:%d", overlap, active, healthFailure, sequence, workers*3)
	}
}

func TestFakeMailEnginePreflightSubmissionReadsExactActiveIdentityWithoutMutation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("account-a", "draft-message", "draft-version")
	account, err := fake.EnsureAccount(ctx, mail.AccountSpec{
		ProvisioningKey: "account-key",
		LocalPart:       "sender",
		Domain:          "example.test",
		Enabled:         true,
		Primary:         mail.AccountIdentitySpec{LocalPart: "sender", Domain: "example.test", Enabled: true},
	})
	if err != nil {
		t.Fatalf("EnsureAccount: %v", err)
	}
	draft, err := fake.SaveDraft(ctx, testDraft(account.AccountID, "draft-key", "", "", "mutation-1"))
	if err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}
	submission := mail.EngineSubmission{AccountID: account.AccountID, DraftKey: "draft-key", MessageID: draft.Projection.MessageID, ExpectedVersion: draft.Projection.Version, SubmissionKey: "submission-key", MaxMessageBytes: 4096}
	before, err := fake.GetDraft(ctx, mail.EngineDraftReadRequest{AccountID: account.AccountID, DraftKey: "draft-key", MessageID: draft.Projection.MessageID, ExpectedVersion: draft.Projection.Version, MaxContentBytes: 4096})
	if err != nil {
		t.Fatalf("GetDraft before preflight: %v", err)
	}
	first, err := fake.PreflightSubmission(ctx, testSubmissionPreflightRequest(submission))
	if err != nil {
		t.Fatalf("PreflightSubmission: %v", err)
	}
	second, err := fake.PreflightSubmission(ctx, testSubmissionPreflightRequest(submission))
	if err != nil {
		t.Fatalf("repeated PreflightSubmission: %v", err)
	}
	if first.FromEmail != "sender@example.test" || first.RecipientCount != 1 || first.MessageBytes < 1 || first.BindingHash == [32]byte{} || first.PolicyBindingHash == [32]byte{} || !reflect.DeepEqual(first, second) {
		t.Fatalf("PreflightSubmission = %#v / %#v, want stable local projection", first, second)
	}
	after, err := fake.GetDraft(ctx, mail.EngineDraftReadRequest{AccountID: account.AccountID, DraftKey: "draft-key", MessageID: draft.Projection.MessageID, ExpectedVersion: draft.Projection.Version, MaxContentBytes: 4096})
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("PreflightSubmission mutated draft: before=%#v after=%#v err=%v", before, after, err)
	}
	fake.EnqueueFault(mail.FakeMailFault{Operation: mail.FakeMailOperationPreflightSubmission, Kind: mail.EngineErrorUnavailable})
	if _, err := fake.PreflightSubmission(ctx, testSubmissionPreflightRequest(submission)); !errors.Is(err, mail.ErrEngineUnavailable) {
		t.Fatalf("scripted preflight fault = %v, want unavailable", err)
	}
	if _, err := fake.PreflightSubmission(ctx, testSubmissionPreflightRequest(submission)); err != nil {
		t.Fatalf("PreflightSubmission after consumed fault: %v", err)
	}
}

func TestFakeMailEnginePreflightSubmissionRejectsUnknownSuspendedAndMismatchedIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for _, testCase := range []struct {
		name       string
		account    mail.AccountSpec
		fromEmail  string
		wantEnsure bool
	}{
		{name: "unknown account", fromEmail: "sender@example.test"},
		{name: "suspended account", account: mail.AccountSpec{ProvisioningKey: "suspended-key", LocalPart: "sender", Domain: "example.test", Primary: mail.AccountIdentitySpec{LocalPart: "sender", Domain: "example.test", Enabled: true}}, fromEmail: "sender@example.test", wantEnsure: true},
		{name: "mismatched identity", account: mail.AccountSpec{ProvisioningKey: "mismatch-key", LocalPart: "allowed", Domain: "example.test", Enabled: true, Primary: mail.AccountIdentitySpec{LocalPart: "allowed", Domain: "example.test", Enabled: true}}, fromEmail: "sender@example.test", wantEnsure: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fake := newFake("account-a", "draft-message", "draft-version")
			accountID := "account-a"
			if testCase.wantEnsure {
				account, err := fake.EnsureAccount(ctx, testCase.account)
				if err != nil {
					t.Fatalf("EnsureAccount: %v", err)
				}
				accountID = account.AccountID
			}
			draftRequest := testDraft(accountID, "draft-key", "", "", "mutation-1")
			draftRequest.Message.From.Email = testCase.fromEmail
			draft, err := fake.SaveDraft(ctx, draftRequest)
			if err != nil {
				t.Fatalf("SaveDraft: %v", err)
			}
			submission := mail.EngineSubmission{AccountID: accountID, DraftKey: "draft-key", MessageID: draft.Projection.MessageID, ExpectedVersion: draft.Projection.Version, SubmissionKey: "submission-key", MaxMessageBytes: 4096}
			_, err = fake.PreflightSubmission(ctx, testSubmissionPreflightRequest(submission))
			if !errors.Is(err, mail.ErrEngineIdentityUnavailable) {
				t.Fatalf("PreflightSubmission error = %v, want identity unavailable", err)
			}
		})
	}
}

func TestFakeMailEnginePreflightSubmissionEnforcesPolicySnapshot(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("account-a", "draft-message", "draft-version")
	account, err := fake.EnsureAccount(ctx, mail.AccountSpec{
		ProvisioningKey: "account-key",
		LocalPart:       "sender",
		Domain:          "example.test",
		Enabled:         true,
		Primary:         mail.AccountIdentitySpec{LocalPart: "sender", Domain: "example.test", Enabled: true},
	})
	if err != nil {
		t.Fatalf("EnsureAccount: %v", err)
	}
	draft, err := fake.SaveDraft(ctx, testDraft(account.AccountID, "draft-key", "", "", "mutation-1"))
	if err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}
	submission := mail.EngineSubmission{AccountID: account.AccountID, DraftKey: "draft-key", MessageID: draft.Projection.MessageID, ExpectedVersion: draft.Projection.Version, SubmissionKey: "submission-key", MaxMessageBytes: 4096}
	internalOnly := testSubmissionPreflightRequest(submission)
	internalOnly.ExternalMailEnabled = false
	first, err := fake.PreflightSubmission(ctx, internalOnly)
	if err != nil {
		t.Fatalf("internal-only PreflightSubmission: %v", err)
	}
	second, err := fake.PreflightSubmission(ctx, internalOnly)
	if err != nil || second != first {
		t.Fatalf("deterministic internal-only PreflightSubmission = %#v, %v; want %#v, nil", second, err, first)
	}
	changedRecipients := internalOnly
	changedRecipients.MaxRecipients = 2
	changedRecipientResult, err := fake.PreflightSubmission(ctx, changedRecipients)
	if err != nil || changedRecipientResult.PolicyBindingHash == first.PolicyBindingHash {
		t.Fatalf("recipient policy hash = %#v, %v; want success with distinct binding", changedRecipientResult, err)
	}
	changedMaxBytes := internalOnly
	changedMaxBytes.PolicyMaxMessageBytes = 2_097_152
	changedMaxBytesResult, err := fake.PreflightSubmission(ctx, changedMaxBytes)
	if err != nil || changedMaxBytesResult.PolicyBindingHash == first.PolicyBindingHash {
		t.Fatalf("message policy hash = %#v, %v; want success with distinct binding", changedMaxBytesResult, err)
	}
	changedExternalMail := internalOnly
	changedExternalMail.ExternalMailEnabled = true
	changedExternalResult, err := fake.PreflightSubmission(ctx, changedExternalMail)
	if err != nil || changedExternalResult.PolicyBindingHash == first.PolicyBindingHash {
		t.Fatalf("external-mail policy hash = %#v, %v; want success with distinct binding", changedExternalResult, err)
	}
	changedDomains := internalOnly
	changedDomains.InternalDomains = []string{"example.test", "tenant.example.test"}
	changedDomainsResult, err := fake.PreflightSubmission(ctx, changedDomains)
	if err != nil || changedDomainsResult.PolicyBindingHash == first.PolicyBindingHash {
		t.Fatalf("internal-domain policy hash = %#v, %v; want success with distinct binding", changedDomainsResult, err)
	}

	invalidOrder := internalOnly
	invalidOrder.InternalDomains = []string{"tenant.example.test", "example.test"}
	fake.EnqueueFault(mail.FakeMailFault{Operation: mail.FakeMailOperationPreflightSubmission, Kind: mail.EngineErrorUnavailable})
	if _, err := fake.PreflightSubmission(ctx, invalidOrder); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("unsorted domain snapshot error = %v, want rejected policy", err)
	}
	if _, err := fake.PreflightSubmission(ctx, internalOnly); !errors.Is(err, mail.ErrEngineUnavailable) {
		t.Fatalf("valid preflight after malformed snapshot = %v, want unconsumed scripted fault", err)
	}
	duplicateDomain := internalOnly
	duplicateDomain.InternalDomains = []string{"example.test", "example.test"}
	if _, err := fake.PreflightSubmission(ctx, duplicateDomain); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("duplicate domain snapshot error = %v, want rejected policy", err)
	}
	invalidMaxRecipients := internalOnly
	invalidMaxRecipients.MaxRecipients = 0
	if _, err := fake.PreflightSubmission(ctx, invalidMaxRecipients); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("invalid max recipients error = %v, want rejected policy", err)
	}

	externalDraftRequest := testDraft(account.AccountID, "draft-key", draft.Projection.MessageID, draft.Projection.Version, "mutation-2")
	externalDraftRequest.Message.To[0].Email = "recipient@outside.test"
	externalDraft, err := fake.SaveDraft(ctx, externalDraftRequest)
	if err != nil {
		t.Fatalf("SaveDraft external recipient: %v", err)
	}
	externalRequest := testSubmissionPreflightRequest(mail.EngineSubmission{AccountID: account.AccountID, DraftKey: "draft-key", MessageID: externalDraft.Projection.MessageID, ExpectedVersion: externalDraft.Projection.Version, SubmissionKey: "submission-key", MaxMessageBytes: 4096})
	externalRequest.ExternalMailEnabled = false
	if _, err := fake.PreflightSubmission(ctx, externalRequest); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf("external-disabled preflight error = %v, want rejected policy", err)
	}
}

func TestFakeMailEnginePreflightSubmissionUsesLoweredCurrentMessagePolicy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("account-a", "draft-message", "draft-version")
	account, err := fake.EnsureAccount(ctx, mail.AccountSpec{
		ProvisioningKey: "account-key",
		LocalPart:       "sender",
		Domain:          "example.test",
		Enabled:         true,
		Primary:         mail.AccountIdentitySpec{LocalPart: "sender", Domain: "example.test", Enabled: true},
	})
	if err != nil {
		t.Fatalf("EnsureAccount: %v", err)
	}
	draftRequest := testDraft(account.AccountID, "draft-key", "", "", "mutation-1")
	draftRequest.MaxContentBytes = 2_097_152
	draftRequest.Message.PlainBody = strings.Repeat("x", 1_048_577)
	draft, err := fake.SaveDraft(ctx, draftRequest)
	if err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}
	request := testSubmissionPreflightRequest(mail.EngineSubmission{AccountID: account.AccountID, DraftKey: "draft-key", MessageID: draft.Projection.MessageID, ExpectedVersion: draft.Projection.Version, SubmissionKey: "submission-key", MaxMessageBytes: 2_097_152})
	request.PolicyMaxMessageBytes = 1_048_576
	if _, err := fake.PreflightSubmission(ctx, request); !errors.Is(err, mail.ErrEnginePayloadTooLarge) {
		t.Fatalf("lowered current message policy error = %v, want payload too large", err)
	}
}

func testSubmissionPreflightRequest(submission mail.EngineSubmission) mail.EngineSubmissionPreflightRequest {
	return mail.EngineSubmissionPreflightRequest{
		Submission:            submission,
		PolicyMaxMessageBytes: 1_048_576,
		MaxRecipients:         mail.MaxEngineDraftTotalRecipientItems,
		ExternalMailEnabled:   true,
		InternalDomains:       []string{"example.test"},
	}
}

func newFake(ids ...string) *mail.FakeMailEngine {
	index := 0
	return mail.NewFakeMailEngine(mail.FakeMailEngineConfig{
		NextID: func() string {
			if index >= len(ids) {
				return fmt.Sprintf("unused-%d", index)
			}
			id := ids[index]
			index++
			return id
		},
		Now: func() time.Time { return time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC) },
	})
}

func testDraft(accountID, draftKey, messageID, previousVersion, mutationID string) mail.EngineDraft {
	return mail.EngineDraft{
		AccountID:       accountID,
		DraftKey:        draftKey,
		MessageID:       messageID,
		PreviousVersion: previousVersion,
		MutationID:      mutationID,
		MaxContentBytes: 4096,
		Message: mail.EngineMessageContent{
			From:      mail.EngineAddress{Email: "sender@example.test"},
			To:        []mail.EngineAddress{{Email: "recipient@example.test"}},
			Subject:   "subject",
			Format:    mail.EngineBodyPlain,
			PlainBody: "draft body",
			Priority:  mail.EnginePriorityNormal,
		},
	}
}

func testMessage(accountID, messageID, version, body string) mail.EngineMessage {
	when := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	return mail.EngineMessage{
		Projection: mail.EngineMessageProjection{
			AccountID:  accountID,
			MessageID:  messageID,
			Version:    version,
			From:       mail.EngineAddress{Email: "sender@example.test"},
			To:         []mail.EngineAddress{{Email: "recipient@example.test"}},
			Subject:    "subject",
			Snippet:    "snippet",
			ReceivedAt: &when,
			SizeBytes:  int64(len(body)),
			Priority:   mail.EnginePriorityNormal,
			FolderIDs:  []string{"inbox"},
			Origin:     mail.EngineProjectionOriginInboundOrUnknown,
		},
		Body: mail.EngineMessageBody{Format: mail.EngineBodyPlain, PlainBody: body},
	}
}
