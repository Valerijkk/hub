package mail_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

type mailboxSubmissionFinalizationRepositoryStub struct{}

func (mailboxSubmissionFinalizationRepositoryStub) ScanAcceptedMailboxSubmissions(context.Context, mail.MailboxSubmissionFinalizationScanRequest) (mail.MailboxSubmissionFinalizationScanResult, error) {
	return mail.MailboxSubmissionFinalizationScanResult{SubmissionIDs: []uuid.UUID{}}, nil
}

func (mailboxSubmissionFinalizationRepositoryStub) ClaimAcceptedMailboxSubmission(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) {
	return mail.NewMailboxSubmissionFinalizationNoopResult(), nil
}

func (mailboxSubmissionFinalizationRepositoryStub) CompleteAcceptedMailboxSubmission(context.Context, mail.MailboxSubmissionFinalizationCompletion) error {
	return nil
}

func (mailboxSubmissionFinalizationRepositoryStub) ReleaseAcceptedMailboxSubmission(context.Context, mail.MailboxSubmissionFinalizationClaim) error {
	return nil
}

var _ mail.MailboxSubmissionFinalizationRepository = mailboxSubmissionFinalizationRepositoryStub{}

type mailboxSubmissionFinalizationRepositorySpy struct {
	scanFn        func(context.Context, mail.MailboxSubmissionFinalizationScanRequest) (mail.MailboxSubmissionFinalizationScanResult, error)
	claimFn       func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error)
	completeFn    func(context.Context, mail.MailboxSubmissionFinalizationCompletion) error
	releaseFn     func(context.Context, mail.MailboxSubmissionFinalizationClaim) error
	scanCalls     int
	claimCalls    int
	completeCalls int
	releaseCalls  int
	claimRequests []mail.MailboxSubmissionFinalizationClaimRequest
	releaseClaims []mail.MailboxSubmissionFinalizationClaim
	releaseCtxs   []context.Context
}

func (spy *mailboxSubmissionFinalizationRepositorySpy) ScanAcceptedMailboxSubmissions(ctx context.Context, request mail.MailboxSubmissionFinalizationScanRequest) (mail.MailboxSubmissionFinalizationScanResult, error) {
	spy.scanCalls++
	if spy.scanFn != nil {
		return spy.scanFn(ctx, request)
	}
	return mail.MailboxSubmissionFinalizationScanResult{SubmissionIDs: []uuid.UUID{}}, nil
}

func (spy *mailboxSubmissionFinalizationRepositorySpy) ClaimAcceptedMailboxSubmission(ctx context.Context, request mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) {
	spy.claimCalls++
	spy.claimRequests = append(spy.claimRequests, request)
	if spy.claimFn == nil {
		return mail.NewMailboxSubmissionFinalizationNoopResult(), nil
	}
	return spy.claimFn(ctx, request)
}

func (spy *mailboxSubmissionFinalizationRepositorySpy) CompleteAcceptedMailboxSubmission(ctx context.Context, completion mail.MailboxSubmissionFinalizationCompletion) error {
	spy.completeCalls++
	if spy.completeFn == nil {
		return nil
	}
	return spy.completeFn(ctx, completion)
}

func (spy *mailboxSubmissionFinalizationRepositorySpy) ReleaseAcceptedMailboxSubmission(ctx context.Context, claim mail.MailboxSubmissionFinalizationClaim) error {
	spy.releaseCalls++
	spy.releaseClaims = append(spy.releaseClaims, claim)
	spy.releaseCtxs = append(spy.releaseCtxs, ctx)
	if spy.releaseFn == nil {
		return nil
	}
	return spy.releaseFn(ctx, claim)
}

type mailboxSubmissionFinalizationEngineSpy struct {
	mail.MailEngine
	messageFn    func(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error)
	blobFn       func(context.Context, mail.EngineBlobReadRequest) (mail.EngineBlob, error)
	messageCalls int
	blobCalls    int
	messageReads []mail.EngineMessageReadRequest
	blobReads    []mail.EngineBlobReadRequest
}

func (spy *mailboxSubmissionFinalizationEngineSpy) GetMessage(ctx context.Context, request mail.EngineMessageReadRequest) (mail.EngineMessage, error) {
	spy.messageCalls++
	spy.messageReads = append(spy.messageReads, request)
	if spy.messageFn == nil {
		return mail.EngineMessage{}, errors.New("unexpected message read")
	}
	return spy.messageFn(ctx, request)
}

func (spy *mailboxSubmissionFinalizationEngineSpy) GetBlob(ctx context.Context, request mail.EngineBlobReadRequest) (mail.EngineBlob, error) {
	spy.blobCalls++
	spy.blobReads = append(spy.blobReads, request)
	if spy.blobFn == nil {
		return mail.EngineBlob{}, errors.New("unexpected blob read")
	}
	return spy.blobFn(ctx, request)
}

func mailboxSubmissionFinalizationFixture(t *testing.T, expected []mail.MailboxSubmissionFinalizationExpectedAttachment) (uuid.UUID, mail.MailboxSubmissionFinalizationClaimResult) {
	t.Helper()
	submissionID := uuid.New()
	input := mail.MailboxSubmissionFinalizationClaimedInput{
		SubmissionID:              submissionID,
		DraftID:                   uuid.New(),
		Fence:                     mail.MailboxEngineCommandFence{AccountID: uuid.New(), LeaseID: uuid.New(), MutationEpoch: 1, LeaseUntil: time.Date(2001, time.February, 3, 4, 6, 5, 0, time.UTC)},
		EngineAccountID:           "account-opaque",
		ProviderSubmissionID:      "provider-submission-opaque",
		ProviderMessageID:         "provider-message-opaque",
		EngineMessageID:           "provider-message-opaque",
		EngineMaxBytes:            1_048_576,
		EngineBindingHash:         mailboxSubmissionFinalizationHash("binding"),
		DispatchAuthorizationHash: mailboxSubmissionFinalizationHash("authorization"),
		DispatchAuthorizedAt:      time.Date(2001, time.February, 3, 4, 5, 0, 0, time.UTC),
		ExpectedAttachments:       expected,
	}
	result, err := mail.NewMailboxSubmissionFinalizationClaimedResult(input)
	if err != nil {
		t.Fatalf("NewMailboxSubmissionFinalizationClaimedResult() error = %v", err)
	}
	return submissionID, result
}

func mailboxSubmissionFinalizationHash(value string) []byte {
	digest := sha256.Sum256([]byte(value))
	return append([]byte(nil), digest[:]...)
}

func mailboxSubmissionFinalizationClaimedInput(expected []mail.MailboxSubmissionFinalizationExpectedAttachment) mail.MailboxSubmissionFinalizationClaimedInput {
	return mail.MailboxSubmissionFinalizationClaimedInput{
		SubmissionID:              uuid.New(),
		DraftID:                   uuid.New(),
		Fence:                     mail.MailboxEngineCommandFence{AccountID: uuid.New(), LeaseID: uuid.New(), MutationEpoch: 1, LeaseUntil: time.Date(2001, time.February, 3, 4, 6, 5, 0, time.UTC)},
		EngineAccountID:           "account-opaque",
		ProviderSubmissionID:      "provider-submission-opaque",
		ProviderMessageID:         "provider-message-opaque",
		EngineMessageID:           "provider-message-opaque",
		EngineMaxBytes:            1_048_576,
		EngineBindingHash:         mailboxSubmissionFinalizationHash("binding"),
		DispatchAuthorizationHash: mailboxSubmissionFinalizationHash("authorization"),
		DispatchAuthorizedAt:      time.Date(2001, time.February, 3, 4, 5, 5, 0, time.UTC),
		ExpectedAttachments:       expected,
	}
}

func mailboxSubmissionFinalizationExpected(t *testing.T, bytes []byte, filename string) mail.MailboxSubmissionFinalizationExpectedAttachment {
	t.Helper()
	return mail.MailboxSubmissionFinalizationExpectedAttachment{
		AttachmentID: uuid.New(),
		ObjectID:     uuid.New(),
		SHA256:       mailboxSubmissionFinalizationHash(string(bytes)),
		SizeBytes:    int64(len(bytes)),
		Filename:     filename,
		MediaType:    "text/plain",
		Disposition:  mail.EngineAttachmentDispositionAttachment,
	}
}

func mailboxSubmissionFinalizationMessage(attachments []mail.EngineProjectionAttachment) mail.EngineMessage {
	sentAt := time.Date(2001, time.February, 3, 4, 5, 7, 0, time.UTC)
	return mail.EngineMessage{
		Projection: mail.EngineMessageProjection{
			AccountID:             "account-opaque",
			MessageID:             "provider-message-opaque",
			Version:               "version-opaque",
			From:                  mail.EngineAddress{Email: "sender@example.test"},
			To:                    []mail.EngineAddress{{Email: "recipient@example.test"}},
			AuthenticationClass:   mail.MailAuthenticationClassUntrusted,
			SentAt:                &sentAt,
			SizeBytes:             4,
			Priority:              mail.EnginePriorityNormal,
			FolderIDs:             []string{"sent"},
			Attachments:           attachments,
			Origin:                mail.EngineProjectionOriginInboundOrUnknown,
		},
		Body: mail.EngineMessageBody{Format: mail.EngineBodyPlain, PlainBody: "body"},
	}
}

func mailboxSubmissionFinalizationAttachment(blobID string, bytes []byte, filename string) mail.EngineProjectionAttachment {
	return mail.EngineProjectionAttachment{BlobID: blobID, Filename: filename, MediaType: "text/plain", SizeBytes: int64(len(bytes)), Disposition: mail.EngineAttachmentDispositionAttachment}
}

func mailboxSubmissionFinalizer(t *testing.T, repository *mailboxSubmissionFinalizationRepositorySpy, engine *mailboxSubmissionFinalizationEngineSpy) *mail.MailboxSubmissionFinalizer {
	t.Helper()
	finalizer, err := mail.NewMailboxSubmissionFinalizer(repository, engine)
	if err != nil {
		t.Fatalf("NewMailboxSubmissionFinalizer() error = %v", err)
	}
	return finalizer
}

func TestMailboxSubmissionFinalizerScanCopiesBoundedDistinctIDs(t *testing.T) {
	first, second := uuid.New(), uuid.New()
	source := []uuid.UUID{first, second}
	repository := &mailboxSubmissionFinalizationRepositorySpy{scanFn: func(_ context.Context, request mail.MailboxSubmissionFinalizationScanRequest) (mail.MailboxSubmissionFinalizationScanResult, error) {
		if request.Limit != 2 {
			t.Fatalf("scan request = %#v, want limit 2", request)
		}
		return mail.MailboxSubmissionFinalizationScanResult{SubmissionIDs: source}, nil
	}}
	result, err := mailboxSubmissionFinalizer(t, repository, &mailboxSubmissionFinalizationEngineSpy{}).ScanAccepted(context.Background(), 2)
	if err != nil || !reflect.DeepEqual(result.SubmissionIDs, source) || result.SubmissionIDs == nil || repository.scanCalls != 1 {
		t.Fatalf("ScanAccepted() = %#v, %v calls=%d", result, err, repository.scanCalls)
	}
	result.SubmissionIDs[0] = uuid.New()
	if source[0] != first {
		t.Fatal("ScanAccepted() leaked repository slice")
	}

	repository.scanFn = func(context.Context, mail.MailboxSubmissionFinalizationScanRequest) (mail.MailboxSubmissionFinalizationScanResult, error) {
		return mail.MailboxSubmissionFinalizationScanResult{SubmissionIDs: []uuid.UUID{first, first}}, nil
	}
	if _, err := mailboxSubmissionFinalizer(t, repository, &mailboxSubmissionFinalizationEngineSpy{}).ScanAccepted(context.Background(), 2); !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorInvalid) {
		t.Fatalf("duplicate scan error = %v, want invalid", err)
	}
}

func TestMailboxSubmissionFinalizerNoopAndBusyDoNotReadOrWrite(t *testing.T) {
	busy, err := mail.NewMailboxSubmissionFinalizationBusyResult(time.Date(2001, time.February, 3, 4, 5, 6, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewMailboxSubmissionFinalizationBusyResult() error = %v", err)
	}
	for _, testCase := range []struct {
		name  string
		claim mail.MailboxSubmissionFinalizationClaimResult
	}{{name: "noop", claim: mail.NewMailboxSubmissionFinalizationNoopResult()}, {name: "busy", claim: busy}} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { return testCase.claim, nil }}
			engine := &mailboxSubmissionFinalizationEngineSpy{}
			if err := mailboxSubmissionFinalizer(t, repository, engine).Finalize(context.Background(), uuid.New()); err != nil {
				t.Fatalf("Finalize() error = %v, want nil", err)
			}
			if repository.claimCalls != 1 || repository.completeCalls != 0 || repository.releaseCalls != 0 || engine.messageCalls != 0 || engine.blobCalls != 0 {
				t.Fatalf("claim/complete/release/message/blob = %d/%d/%d/%d/%d, want 1/0/0/0/0", repository.claimCalls, repository.completeCalls, repository.releaseCalls, engine.messageCalls, engine.blobCalls)
			}
		})
	}
}

func TestMailboxSubmissionFinalizerZeroAttachmentsCompletesExactClaim(t *testing.T) {
	submissionID, claim := mailboxSubmissionFinalizationFixture(t, []mail.MailboxSubmissionFinalizationExpectedAttachment{})
	repository := &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { return claim, nil }}
	engine := &mailboxSubmissionFinalizationEngineSpy{messageFn: func(_ context.Context, request mail.EngineMessageReadRequest) (mail.EngineMessage, error) {
		if want := (mail.EngineMessageReadRequest{AccountID: "account-opaque", MessageID: "provider-message-opaque", MaxContentBytes: 1_048_576}); request != want {
			t.Fatalf("GetMessage request = %#v, want %#v", request, want)
		}
		return mailboxSubmissionFinalizationMessage([]mail.EngineProjectionAttachment{}), nil
	}}
	if err := mailboxSubmissionFinalizer(t, repository, engine).Finalize(context.Background(), submissionID); err != nil {
		t.Fatalf("Finalize() error = %v", err)
	}
	if repository.completeCalls != 1 || repository.releaseCalls != 0 || engine.messageCalls != 1 || engine.blobCalls != 0 {
		t.Fatalf("complete/release/message/blob = %d/%d/%d/%d, want 1/0/1/0", repository.completeCalls, repository.releaseCalls, engine.messageCalls, engine.blobCalls)
	}
}

func TestMailboxSubmissionFinalizerRequiresCanonicalSentAtAndClosesReleaseFailure(t *testing.T) {
	submissionID, claim := mailboxSubmissionFinalizationFixture(t, []mail.MailboxSubmissionFinalizationExpectedAttachment{})
	repository := &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { return claim, nil }}
	engine := &mailboxSubmissionFinalizationEngineSpy{messageFn: func(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error) {
		message := mailboxSubmissionFinalizationMessage([]mail.EngineProjectionAttachment{})
		message.Projection.SentAt = nil
		return message, nil
	}}
	err := mailboxSubmissionFinalizer(t, repository, engine).Finalize(context.Background(), submissionID)
	if !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorFinalizationRequired) || repository.releaseCalls != 1 || repository.completeCalls != 0 {
		t.Fatalf("missing SentAt Finalize()=%v release/complete=%d/%d, want finalization_required/1/0", err, repository.releaseCalls, repository.completeCalls)
	}

	submissionID, claim = mailboxSubmissionFinalizationFixture(t, []mail.MailboxSubmissionFinalizationExpectedAttachment{})
	repository = &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { return claim, nil }, releaseFn: func(context.Context, mail.MailboxSubmissionFinalizationClaim) error { return errors.New("raw release failure") }}
	engine = &mailboxSubmissionFinalizationEngineSpy{messageFn: func(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error) { return mail.EngineMessage{}, errors.New("raw message failure") }}
	err = mailboxSubmissionFinalizer(t, repository, engine).Finalize(context.Background(), submissionID)
	if !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorRepositoryUnavailable) || repository.releaseCalls != 1 || strings.Contains(err.Error(), "raw release failure") {
		t.Fatalf("release failure Finalize()=%v release=%d, want repository_unavailable/1", err, repository.releaseCalls)
	}
}

func TestMailboxSubmissionFinalizerMatchesAttachmentMultisetsWithoutOrder(t *testing.T) {
	firstBytes, secondBytes := []byte("first"), []byte("second")
	first, second := mailboxSubmissionFinalizationExpected(t, firstBytes, "first.txt"), mailboxSubmissionFinalizationExpected(t, secondBytes, "second.txt")
	submissionID, claim := mailboxSubmissionFinalizationFixture(t, []mail.MailboxSubmissionFinalizationExpectedAttachment{first, second})
	message := mailboxSubmissionFinalizationMessage([]mail.EngineProjectionAttachment{
		mailboxSubmissionFinalizationAttachment("blob-second", secondBytes, "second.txt"),
		mailboxSubmissionFinalizationAttachment("blob-first", firstBytes, "first.txt"),
	})
	engine := &mailboxSubmissionFinalizationEngineSpy{
		messageFn: func(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error) { return message, nil },
		blobFn: func(_ context.Context, request mail.EngineBlobReadRequest) (mail.EngineBlob, error) {
			if request.BlobID == "blob-first" { return mail.EngineBlob{AccountID: request.AccountID, BlobID: request.BlobID, SizeBytes: int64(len(firstBytes)), Bytes: firstBytes}, nil }
			return mail.EngineBlob{AccountID: request.AccountID, BlobID: request.BlobID, SizeBytes: int64(len(secondBytes)), Bytes: secondBytes}, nil
		},
	}
	repository := &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { return claim, nil }}
	if err := mailboxSubmissionFinalizer(t, repository, engine).Finalize(context.Background(), submissionID); err != nil { t.Fatalf("Finalize() error = %v", err) }
	if repository.completeCalls != 1 || repository.releaseCalls != 0 || engine.blobCalls != 2 { t.Fatalf("complete/release/blob = %d/%d/%d, want 1/0/2", repository.completeCalls, repository.releaseCalls, engine.blobCalls) }
}

func TestMailboxSubmissionFinalizerKeepsDuplicateAttachmentMultiplicity(t *testing.T) {
	bytes := []byte("same")
	first, second := mailboxSubmissionFinalizationExpected(t, bytes, "same.txt"), mailboxSubmissionFinalizationExpected(t, bytes, "same.txt")
	submissionID, claim := mailboxSubmissionFinalizationFixture(t, []mail.MailboxSubmissionFinalizationExpectedAttachment{first, second})
	message := mailboxSubmissionFinalizationMessage([]mail.EngineProjectionAttachment{mailboxSubmissionFinalizationAttachment("blob-one", bytes, "same.txt"), mailboxSubmissionFinalizationAttachment("blob-two", bytes, "same.txt")})
	engine := &mailboxSubmissionFinalizationEngineSpy{messageFn: func(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error) { return message, nil }, blobFn: func(_ context.Context, request mail.EngineBlobReadRequest) (mail.EngineBlob, error) { return mail.EngineBlob{AccountID: request.AccountID, BlobID: request.BlobID, SizeBytes: int64(len(bytes)), Bytes: bytes}, nil }}
	repository := &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { return claim, nil }}
	if err := mailboxSubmissionFinalizer(t, repository, engine).Finalize(context.Background(), submissionID); err != nil { t.Fatalf("Finalize() error = %v", err) }
	if repository.completeCalls != 1 || repository.releaseCalls != 0 { t.Fatalf("complete/release = %d/%d, want 1/0", repository.completeCalls, repository.releaseCalls) }
}

func TestMailboxSubmissionFinalizerRejectsDroppedDuplicateAndDifferentBytes(t *testing.T) {
	for _, testCase := range []struct { name string; expected []mail.MailboxSubmissionFinalizationExpectedAttachment; attachments []mail.EngineProjectionAttachment; blobs map[string][]byte }{
		{name: "dropped duplicate", expected: func() []mail.MailboxSubmissionFinalizationExpectedAttachment { bytes := []byte("same"); return []mail.MailboxSubmissionFinalizationExpectedAttachment{mailboxSubmissionFinalizationExpected(t, bytes, "same.txt"), mailboxSubmissionFinalizationExpected(t, bytes, "same.txt")} }(), attachments: []mail.EngineProjectionAttachment{mailboxSubmissionFinalizationAttachment("blob-one", []byte("same"), "same.txt")}, blobs: map[string][]byte{"blob-one": []byte("same")}},
		{name: "same filename different bytes", expected: []mail.MailboxSubmissionFinalizationExpectedAttachment{mailboxSubmissionFinalizationExpected(t, []byte("good"), "same.txt")}, attachments: []mail.EngineProjectionAttachment{mailboxSubmissionFinalizationAttachment("blob-one", []byte("evil"), "same.txt")}, blobs: map[string][]byte{"blob-one": []byte("evil")}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			submissionID, claim := mailboxSubmissionFinalizationFixture(t, testCase.expected)
			repository := &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { return claim, nil }}
			engine := &mailboxSubmissionFinalizationEngineSpy{messageFn: func(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error) { return mailboxSubmissionFinalizationMessage(testCase.attachments), nil }, blobFn: func(_ context.Context, request mail.EngineBlobReadRequest) (mail.EngineBlob, error) { bytes := testCase.blobs[request.BlobID]; return mail.EngineBlob{AccountID: request.AccountID, BlobID: request.BlobID, SizeBytes: int64(len(bytes)), Bytes: bytes}, nil }}
			err := mailboxSubmissionFinalizer(t, repository, engine).Finalize(context.Background(), submissionID)
			if !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorFinalizationRequired) { t.Fatalf("Finalize() error = %v, want finalization_required", err) }
			if repository.completeCalls != 0 || repository.releaseCalls != 1 { t.Fatalf("complete/release = %d/%d, want 0/1", repository.completeCalls, repository.releaseCalls) }
		})
	}
}

func TestMailboxSubmissionFinalizerVerifiesProviderSHAAndCachesBlobReads(t *testing.T) {
	bytes := []byte("blob")
	expected := mailboxSubmissionFinalizationExpected(t, bytes, "blob.txt")
	for _, testCase := range []struct { name string; attachment mail.EngineProjectionAttachment; wantErr bool }{
		{name: "empty SHA accepted", attachment: mailboxSubmissionFinalizationAttachment("blob-one", bytes, "blob.txt")},
		{name: "provider SHA mismatch", attachment: func() mail.EngineProjectionAttachment { attachment := mailboxSubmissionFinalizationAttachment("blob-one", bytes, "blob.txt"); attachment.SHA256 = mailboxSubmissionFinalizationHash("different"); return attachment }(), wantErr: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			submissionID, claim := mailboxSubmissionFinalizationFixture(t, []mail.MailboxSubmissionFinalizationExpectedAttachment{expected})
			repository := &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { return claim, nil }}
			engine := &mailboxSubmissionFinalizationEngineSpy{messageFn: func(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error) { return mailboxSubmissionFinalizationMessage([]mail.EngineProjectionAttachment{testCase.attachment}), nil }, blobFn: func(_ context.Context, request mail.EngineBlobReadRequest) (mail.EngineBlob, error) { return mail.EngineBlob{AccountID: request.AccountID, BlobID: request.BlobID, SizeBytes: int64(len(bytes)), Bytes: bytes}, nil }}
			err := mailboxSubmissionFinalizer(t, repository, engine).Finalize(context.Background(), submissionID)
			if testCase.wantErr {
				if !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorFinalizationRequired) || repository.releaseCalls != 1 { t.Fatalf("Finalize() = %v release=%d, want finalization_required/1", err, repository.releaseCalls) }
			} else if err != nil || repository.completeCalls != 1 || repository.releaseCalls != 0 { t.Fatalf("Finalize() = %v complete/release=%d/%d, want nil/1/0", err, repository.completeCalls, repository.releaseCalls) }
		})
	}

	t.Run("repeated blob read once", func(t *testing.T) {
		first, second := mailboxSubmissionFinalizationExpected(t, bytes, "one.txt"), mailboxSubmissionFinalizationExpected(t, bytes, "two.txt")
		submissionID, claim := mailboxSubmissionFinalizationFixture(t, []mail.MailboxSubmissionFinalizationExpectedAttachment{first, second})
		message := mailboxSubmissionFinalizationMessage([]mail.EngineProjectionAttachment{mailboxSubmissionFinalizationAttachment("blob-one", bytes, "one.txt"), mailboxSubmissionFinalizationAttachment("blob-one", bytes, "two.txt")})
		repository := &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { return claim, nil }}
		engine := &mailboxSubmissionFinalizationEngineSpy{messageFn: func(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error) { return message, nil }, blobFn: func(_ context.Context, request mail.EngineBlobReadRequest) (mail.EngineBlob, error) { return mail.EngineBlob{AccountID: request.AccountID, BlobID: request.BlobID, SizeBytes: int64(len(bytes)), Bytes: bytes}, nil }}
		if err := mailboxSubmissionFinalizer(t, repository, engine).Finalize(context.Background(), submissionID); err != nil { t.Fatalf("Finalize() error = %v", err) }
		if engine.blobCalls != 1 || repository.completeCalls != 1 { t.Fatalf("blob/complete = %d/%d, want 1/1", engine.blobCalls, repository.completeCalls) }
	})
}

func TestMailboxSubmissionFinalizerRejectsIncompatibleRepeatedBlobAndForeignMessage(t *testing.T) {
	bytes := []byte("blob")
	expected := mailboxSubmissionFinalizationExpected(t, bytes, "blob.txt")
	for _, testCase := range []struct { name string; message mail.EngineMessage; wantBlobCalls int }{
		{name: "repeated blob sizes differ", message: mailboxSubmissionFinalizationMessage([]mail.EngineProjectionAttachment{mailboxSubmissionFinalizationAttachment("blob-one", bytes, "blob.txt"), func() mail.EngineProjectionAttachment { attachment := mailboxSubmissionFinalizationAttachment("blob-one", bytes, "blob.txt"); attachment.SizeBytes++; return attachment }()}), wantBlobCalls: 0},
		{name: "foreign account", message: func() mail.EngineMessage { message := mailboxSubmissionFinalizationMessage([]mail.EngineProjectionAttachment{}); message.Projection.AccountID = "foreign-opaque"; return message }()},
		{name: "foreign message", message: func() mail.EngineMessage { message := mailboxSubmissionFinalizationMessage([]mail.EngineProjectionAttachment{}); message.Projection.MessageID = "foreign-message-opaque"; return message }()},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			submissionID, claim := mailboxSubmissionFinalizationFixture(t, []mail.MailboxSubmissionFinalizationExpectedAttachment{expected})
			repository := &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { return claim, nil }}
			engine := &mailboxSubmissionFinalizationEngineSpy{messageFn: func(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error) { return testCase.message, nil }, blobFn: func(_ context.Context, request mail.EngineBlobReadRequest) (mail.EngineBlob, error) { return mail.EngineBlob{AccountID: request.AccountID, BlobID: request.BlobID, SizeBytes: int64(len(bytes)), Bytes: bytes}, nil }}
			err := mailboxSubmissionFinalizer(t, repository, engine).Finalize(context.Background(), submissionID)
			if !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorFinalizationRequired) || repository.releaseCalls != 1 { t.Fatalf("Finalize() = %v release=%d, want finalization_required/1", err, repository.releaseCalls) }
			if engine.blobCalls != testCase.wantBlobCalls { t.Fatalf("blob calls = %d, want %d", engine.blobCalls, testCase.wantBlobCalls) }
		})
	}
}

func TestMailboxSubmissionFinalizerCancellationAndCompleteFailureAreClosed(t *testing.T) {
	t.Run("cancel after claim releases detached", func(t *testing.T) {
		submissionID, claim := mailboxSubmissionFinalizationFixture(t, []mail.MailboxSubmissionFinalizationExpectedAttachment{})
		ctx, cancel := context.WithCancel(context.Background())
		repository := &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { cancel(); return claim, nil }, releaseFn: func(releaseCtx context.Context, claim mail.MailboxSubmissionFinalizationClaim) error { if releaseCtx.Err() != nil || claim.SubmissionID() != submissionID { t.Fatalf("release ctx/claim = %v/%s", releaseCtx.Err(), claim.SubmissionID()) }; return nil }}
		engine := &mailboxSubmissionFinalizationEngineSpy{}
		err := mailboxSubmissionFinalizer(t, repository, engine).Finalize(ctx, submissionID)
		if !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorFinalizationRequired) || repository.releaseCalls != 1 || engine.messageCalls != 0 { t.Fatalf("Finalize()=%v release/message=%d/%d, want finalization_required/1/0", err, repository.releaseCalls, engine.messageCalls) }
	})

	t.Run("complete failure releases detached", func(t *testing.T) {
		submissionID, claim := mailboxSubmissionFinalizationFixture(t, []mail.MailboxSubmissionFinalizationExpectedAttachment{})
		originalClaim, claimed := claim.Claim()
		if !claimed {
			t.Fatal("fixture claim was not claimed")
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		released := false
		repository := &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { return claim, nil }, completeFn: func(context.Context, mail.MailboxSubmissionFinalizationCompletion) error { cancel(); return errors.New("raw completion failure recipient@example.test") }, releaseFn: func(releaseCtx context.Context, releasedClaim mail.MailboxSubmissionFinalizationClaim) error {
			if releaseCtx.Err() != nil || releasedClaim.SubmissionID() != originalClaim.SubmissionID() || !reflect.DeepEqual(releasedClaim.Fence(), originalClaim.Fence()) {
				t.Fatalf("release ctx/claim = %v/%s/%#v, want detached/original", releaseCtx.Err(), releasedClaim.SubmissionID(), releasedClaim.Fence())
			}
			released = true
			return nil
		}}
		engine := &mailboxSubmissionFinalizationEngineSpy{messageFn: func(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error) { return mailboxSubmissionFinalizationMessage([]mail.EngineProjectionAttachment{}), nil }}
		err := mailboxSubmissionFinalizer(t, repository, engine).Finalize(ctx, submissionID)
		if !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorFinalizationRequired) || repository.completeCalls != 1 || repository.releaseCalls != 1 || !released || strings.Contains(err.Error(), "recipient@example.test") { t.Fatalf("Finalize()=%v complete/release/released=%d/%d/%t, want closed error/1/1/true", err, repository.completeCalls, repository.releaseCalls, released) }
	})
}

func TestMailboxSubmissionFinalizerRejectsTypedNilAndDoesNotLeakRawFailure(t *testing.T) {
	if finalizer, err := mail.NewMailboxSubmissionFinalizer(nil, nil); finalizer != nil || !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorInvalid) { t.Fatalf("NewMailboxSubmissionFinalizer(nil,nil) = %#v,%v", finalizer, err) }
	var engine *mail.StalwartEngine
	if finalizer, err := mail.NewMailboxSubmissionFinalizer(mailboxSubmissionFinalizationRepositoryStub{}, engine); finalizer != nil || !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorInvalid) { t.Fatalf("NewMailboxSubmissionFinalizer(stub,typed nil) = %#v,%v", finalizer, err) }

	submissionID, claim := mailboxSubmissionFinalizationFixture(t, []mail.MailboxSubmissionFinalizationExpectedAttachment{})
	repository := &mailboxSubmissionFinalizationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) { return claim, nil }}
	engineSpy := &mailboxSubmissionFinalizationEngineSpy{messageFn: func(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error) { return mail.EngineMessage{}, errors.New("raw provider failure secret-message-id") }}
	err := mailboxSubmissionFinalizer(t, repository, engineSpy).Finalize(context.Background(), submissionID)
	if !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorFinalizationRequired) || strings.Contains(err.Error(), "secret-message-id") || strings.Contains(err.Error(), submissionID.String()) || repository.releaseCalls != 1 { t.Fatalf("Finalize() = %v release=%d, want content-free finalization_required/1", err, repository.releaseCalls) }
}

func TestMailboxSubmissionFinalizationFactoryAndSourceContract(t *testing.T) {
	if _, err := mail.NewMailboxSubmissionFinalizationClaimedResult(mail.MailboxSubmissionFinalizationClaimedInput{}); !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorInvalid) { t.Fatalf("zero claimed input error = %v, want invalid", err) }
	if _, err := mail.NewMailboxSubmissionFinalizationBusyResult(time.Time{}); !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorInvalid) { t.Fatalf("zero busy error = %v, want invalid", err) }
	submissionID, claim := mailboxSubmissionFinalizationFixture(t, []mail.MailboxSubmissionFinalizationExpectedAttachment{})
	if got, ok := claim.Claim(); !ok || got.SubmissionID() != submissionID || got.Fence().AccountID == uuid.Nil { t.Fatalf("sealed claim getter = %#v/%t", got, ok) }

	claimType := reflect.TypeOf(mail.MailboxSubmissionFinalizationClaim{})
	for index := 0; index < claimType.NumField(); index++ { if claimType.Field(index).PkgPath == "" { t.Fatalf("claim field %q must remain sealed", claimType.Field(index).Name) } }
	for _, forbidden := range []string{"ProviderSubmissionID", "ProviderMessageID", "ExpectedAttachments", "EngineAccountID"} { if _, found := claimType.MethodByName(forbidden); found { t.Fatalf("claim must not expose %s", forbidden) } }

	source, err := os.ReadFile("../../internal/mail/mailbox_submission_finalizer.go")
	if err != nil { t.Fatalf("read finalizer source: %v", err) }
	text := string(source)
	if strings.Count(text, ".GetMessage(") != 1 || strings.Count(text, ".GetBlob(") != 1 { t.Fatalf("engine source calls GetMessage/GetBlob %d/%d times, want 1/1", strings.Count(text, ".GetMessage("), strings.Count(text, ".GetBlob(")) }
	for _, forbidden := range []string{".Submit(", ".Reconcile", ".Preflight", ".GetDraft(", "NATS", "HTTP", "storage"} { if strings.Contains(strings.ToLower(text), strings.ToLower(forbidden)) { t.Fatalf("finalizer source contains forbidden %q", forbidden) } }
	for _, required := range []string{"context.WithoutCancel(ctx)", "EngineProjectionOriginConfirmedOwnerSubmission", "EvaluateAttachmentPolicy", "sha256.Sum256", "ReleaseAcceptedMailboxSubmission", "CompleteAcceptedMailboxSubmission"} { if !strings.Contains(text, required) { t.Fatalf("finalizer source missing %q", required) } }
}

func TestMailboxSubmissionFinalizationClaimedFactoryValidatesExpectedAttachmentBoundary(t *testing.T) {
	if _, err := mail.NewMailboxSubmissionFinalizationClaimedResult(mailboxSubmissionFinalizationClaimedInput(nil)); !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorInvalid) {
		t.Fatalf("nil expected attachments error = %v, want invalid", err)
	}
	if _, err := mail.NewMailboxSubmissionFinalizationClaimedResult(mailboxSubmissionFinalizationClaimedInput([]mail.MailboxSubmissionFinalizationExpectedAttachment{})); err != nil {
		t.Fatalf("explicit empty expected attachments error = %v, want nil", err)
	}

	first := mailboxSubmissionFinalizationExpected(t, []byte("one"), "one.txt")
	second := mailboxSubmissionFinalizationExpected(t, []byte("two"), "two.txt")
	duplicateAttachmentID := second
	duplicateAttachmentID.AttachmentID = first.AttachmentID
	if _, err := mail.NewMailboxSubmissionFinalizationClaimedResult(mailboxSubmissionFinalizationClaimedInput([]mail.MailboxSubmissionFinalizationExpectedAttachment{first, duplicateAttachmentID})); !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorInvalid) {
		t.Fatalf("duplicate attachment ID error = %v, want invalid", err)
	}

	duplicateObjectID := second
	duplicateObjectID.ObjectID = first.ObjectID
	if _, err := mail.NewMailboxSubmissionFinalizationClaimedResult(mailboxSubmissionFinalizationClaimedInput([]mail.MailboxSubmissionFinalizationExpectedAttachment{first, duplicateObjectID})); !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorInvalid) {
		t.Fatalf("duplicate object ID error = %v, want invalid", err)
	}

	first = mailboxSubmissionFinalizationExpected(t, []byte("aa"), "first.txt")
	second = mailboxSubmissionFinalizationExpected(t, []byte("bb"), "second.txt")
	input := mailboxSubmissionFinalizationClaimedInput([]mail.MailboxSubmissionFinalizationExpectedAttachment{first, second})
	input.EngineMaxBytes = 3
	if _, err := mail.NewMailboxSubmissionFinalizationClaimedResult(input); !mailboxSubmissionFinalizationErrorIs(err, mail.MailboxSubmissionFinalizationErrorInvalid) {
		t.Fatalf("aggregate beyond engine maximum error = %v, want invalid", err)
	}
	input.EngineMaxBytes = 4
	if _, err := mail.NewMailboxSubmissionFinalizationClaimedResult(input); err != nil {
		t.Fatalf("aggregate at engine maximum error = %v, want nil", err)
	}
}

func mailboxSubmissionFinalizationErrorIs(err error, want mail.MailboxSubmissionFinalizationErrorKind) bool {
	var finalizationError *mail.MailboxSubmissionFinalizationError
	return errors.As(err, &finalizationError) && finalizationError != nil && finalizationError.Kind() == want && finalizationError.Error() == "mailbox submission finalization error"
}
