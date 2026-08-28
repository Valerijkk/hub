package mail_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/mail"
)

type consumerStub struct{}

func (consumerStub) Capabilities(context.Context) (mail.EngineCapabilities, error) {
	return mail.EngineCapabilities{}, nil
}

func (consumerStub) GetDomain(context.Context, string) (mail.EngineDomainState, error) {
	return mail.EngineDomainState{}, nil
}

func (consumerStub) EnsureDomain(context.Context, mail.DomainSpec) (mail.EngineDomainState, error) {
	return mail.EngineDomainState{}, nil
}

func (consumerStub) EnsureAccount(context.Context, mail.AccountSpec) (mail.EngineAccount, error) {
	return mail.EngineAccount{}, nil
}

func (consumerStub) ListFolders(context.Context, string) ([]mail.EngineFolder, error) {
	return nil, nil
}

func (consumerStub) CreateFolder(context.Context, string, mail.FolderSpec) (mail.EngineFolder, error) {
	return mail.EngineFolder{}, nil
}

func (consumerStub) RenameFolder(context.Context, string, string, string) error {
	return nil
}

func (consumerStub) DeleteFolder(context.Context, string, string) error {
	return nil
}

func (consumerStub) ListMessages(context.Context, mail.EngineMessageQuery) (mail.EngineProjectionPage, error) {
	return mail.EngineProjectionPage{}, nil
}

func (consumerStub) GetMessage(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error) {
	return mail.EngineMessage{}, nil
}

func (consumerStub) GetDraft(context.Context, mail.EngineDraftReadRequest) (mail.EngineMessage, error) {
	return mail.EngineMessage{}, nil
}

func (consumerStub) GetBlob(context.Context, mail.EngineBlobReadRequest) (mail.EngineBlob, error) {
	return mail.EngineBlob{}, nil
}

func (consumerStub) ReadCalendarPart(context.Context, mail.EngineCalendarPartReadRequest) (mail.EngineCalendarPartReadResult, error) {
	return mail.EngineCalendarPartReadResult{}, nil
}

func (consumerStub) Changes(context.Context, mail.EngineChangeRequest) (mail.EngineChangePage, error) {
	return mail.EngineChangePage{}, nil
}

func (consumerStub) EnumerateProjection(context.Context, mail.EngineEnumerationRequest) (mail.EngineEnumerationPage, error) {
	return mail.EngineEnumerationPage{}, nil
}

func (consumerStub) SaveDraft(context.Context, mail.EngineDraft) (mail.EngineMessage, error) {
	return mail.EngineMessage{}, nil
}

func (consumerStub) ReconcileDraft(context.Context, mail.EngineDraftReconciliation) (mail.EngineDraftResult, error) {
	return mail.EngineDraftAbsent{}, nil
}

func (consumerStub) DeleteDraft(context.Context, mail.EngineDraftDeletion) error {
	return nil
}

func (consumerStub) ReconcileDraftDeletion(context.Context, mail.EngineDraftDeletion) (mail.EngineDraftDeletionResult, error) {
	return mail.EngineDraftDeleted{}, nil
}

func (consumerStub) PreflightSubmission(context.Context, mail.EngineSubmissionPreflightRequest) (mail.EngineSubmissionPreflight, error) {
	return mail.EngineSubmissionPreflight{}, nil
}

func (consumerStub) ReconcileSubmission(context.Context, mail.EngineSubmission) (mail.EngineSubmissionResult, error) {
	return mail.EngineSubmissionAbsent{}, nil
}

func (consumerStub) MoveMessages(context.Context, string, []string, string) error {
	return nil
}

func (consumerStub) SetFlags(context.Context, string, []string, mail.FlagPatch) error {
	return nil
}

func (consumerStub) SetKeywords(context.Context, string, []string, mail.KeywordPatch) error {
	return nil
}

func (consumerStub) DeleteMessages(context.Context, string, []string, bool) error {
	return nil
}

func (consumerStub) GetSieve(context.Context, string, string) (mail.EngineSieveState, error) {
	return mail.EngineSieveState{}, nil
}

func (consumerStub) UpsertSieve(context.Context, string, mail.SieveScript) (mail.EngineSieveState, error) {
	return mail.EngineSieveState{}, nil
}

func (consumerStub) ClassifyMessages(context.Context, mail.EngineSpamTraining) (mail.EngineSpamTrainingResult, error) {
	return mail.EngineSpamTrainingResult{}, nil
}

func (consumerStub) Health(context.Context) (mail.EngineHealth, error) {
	return mail.EngineHealth{}, nil
}

var _ mail.MailEngine = (*consumerStub)(nil)

func TestMailEngine_ExternalConsumerImplementsFrozenPort(t *testing.T) {
	var engine mail.MailEngine = consumerStub{}
	if engine == nil {
		t.Fatal("external consumer stub must satisfy MailEngine")
	}
}

func TestMailEngine_DraftAndSubmissionValuesAreTyped(t *testing.T) {
	draft := mail.EngineDraft{
		AccountID:       "account-opaque",
		DraftKey:        "stable-draft-key",
		MessageID:       "message-opaque",
		PreviousVersion: "version-opaque",
		MutationID:      "hub-mutation-id",
		MaxContentBytes: 1024,
		Message: mail.EngineMessageContent{
			Subject:   "Draft subject",
			PlainBody: "Draft body",
		},
	}
	submission := mail.EngineSubmission{
		AccountID:       "account-opaque",
		DraftKey:        draft.DraftKey,
		MessageID:       draft.MessageID,
		ExpectedVersion: draft.PreviousVersion,
		SubmissionKey:   "stable-submission-key",
		MaxMessageBytes: 1024,
	}

	if draft.DraftKey == "" || draft.PreviousVersion == "" || draft.MutationID == "" || submission.SubmissionKey == "" {
		t.Fatal("draft key, version, mutation id and submission key must remain explicit typed fields")
	}
}

func TestMailEngine_ReconciliationResultsKeepAbsenceAndIndeterminateSeparate(t *testing.T) {
	var absentDraft mail.EngineDraftResult = mail.EngineDraftAbsent{}
	var unknownDraft mail.EngineDraftResult = mail.EngineDraftIndeterminate{Code: mail.EngineIndeterminateTimeout}
	var absentSubmission mail.EngineSubmissionResult = mail.EngineSubmissionAbsent{}
	var unknownSubmission mail.EngineSubmissionResult = mail.EngineSubmissionIndeterminate{Code: mail.EngineIndeterminateUnavailable}

	if reflect.TypeOf(absentDraft) == reflect.TypeOf(unknownDraft) {
		t.Fatal("draft absence and indeterminate outcomes must have separate result shapes")
	}
	if reflect.TypeOf(absentSubmission) == reflect.TypeOf(unknownSubmission) {
		t.Fatal("submission absence and indeterminate outcomes must have separate result shapes")
	}
	var zero mail.EngineSubmissionResult
	if zero != nil {
		t.Fatal("zero submission result must not look like sent")
	}
}

func TestMailEngine_FolderAndMessageReadsAreAccountScoped(t *testing.T) {
	query := mail.EngineMessageQuery{
		AccountID: "account-opaque",
		FolderID:  "folder-opaque",
		Page:      mail.EnginePageRequest{Limit: 50},
	}

	if query.AccountID == "" {
		t.Fatal("message query must carry the engine account id")
	}
	var engine mail.MailEngine = consumerStub{}
	if _, err := engine.ListFolders(context.Background(), query.AccountID); err != nil {
		t.Fatalf("ListFolders() unexpected error = %v", err)
	}
	if _, err := engine.ListMessages(context.Background(), query); err != nil {
		t.Fatalf("ListMessages() unexpected error = %v", err)
	}
}

func TestMailEngine_FlagPatchDistinguishesUnsetFromExplicitFalse(t *testing.T) {
	patch := mail.FlagPatch{Seen: mail.FlagValue{Present: true, Value: false}}
	if !patch.Seen.Present || patch.Seen.Value {
		t.Fatal("flag patch must represent an explicit false independently from an absent change")
	}
	if (mail.FlagPatch{}).Seen.Present {
		t.Fatal("zero flag patch must keep a flag unchanged")
	}
}

func TestMailEngine_HealthAndCapabilitiesDoNotExposeCredentials(t *testing.T) {
	assertNoSensitiveFieldNames(t, reflect.TypeOf(mail.EngineHealth{}))
	assertNoSensitiveFieldNames(t, reflect.TypeOf(mail.EngineCapabilities{}))
}

func TestMailEngine_SentinelErrorsAreTypedAndSafe(t *testing.T) {
	for _, testCase := range []struct {
		kind mail.EngineErrorKind
		want error
		text string
	}{
		{mail.EngineErrorInvalidCursor, mail.ErrEngineInvalidCursor, "mail engine invalid cursor"},
		{mail.EngineErrorRateLimited, mail.ErrEngineRateLimited, "mail engine rate limited"},
		{mail.EngineErrorIdentityUnavailable, mail.ErrEngineIdentityUnavailable, "mail engine identity unavailable"},
		{mail.EngineErrorRecipientLimit, mail.ErrEngineRecipientLimit, "mail engine recipient limit"},
	} {
		t.Run(string(testCase.kind), func(t *testing.T) {
			err := &mail.EngineError{Kind: testCase.kind}
			if err.Error() != testCase.text || !errors.Is(err, testCase.want) || errors.Unwrap(err) != testCase.want {
				t.Fatalf("EngineError{%q} did not preserve its fixed sentinel contract", testCase.kind)
			}
			var typed *mail.EngineError
			if !errors.As(err, &typed) || typed.Kind != testCase.kind {
				t.Fatalf("EngineError{%q} must be available through errors.As", testCase.kind)
			}
		})
	}
}

func TestMailEngine_SubmissionPreflightShapeIsConcrete(t *testing.T) {
	binding := sha256.Sum256([]byte("mail-engine-preflight"))
	policyBinding := sha256.Sum256([]byte("mail-engine-preflight-policy"))
	request := mail.EngineSubmissionPreflightRequest{
		Submission: mail.EngineSubmission{
			AccountID:       "account-a",
			DraftKey:        "draft-a",
			MessageID:       "message-a",
			ExpectedVersion: "version-a",
			SubmissionKey:   "submission-a",
			MaxMessageBytes: 1_048_576,
		},
		PolicyMaxMessageBytes: 1_048_576,
		MaxRecipients:         5,
		ExternalMailEnabled:   false,
		InternalDomains:       []string{"example.test"},
	}
	if request.Submission.AccountID == "" || request.PolicyMaxMessageBytes != 1_048_576 || request.MaxRecipients != 5 || request.ExternalMailEnabled || len(request.InternalDomains) != 1 {
		t.Fatalf("EngineSubmissionPreflightRequest = %#v, want explicit content-free policy snapshot", request)
	}
	assertExactFields(t, reflect.TypeOf(mail.EngineSubmissionPreflightRequest{}), reflect.TypeOf(struct {
		Submission             mail.EngineSubmission
		PolicyMaxMessageBytes  int64
		MaxRecipients          int32
		ExternalMailEnabled    bool
		InternalDomains        []string
	}{}))
	preflight := mail.EngineSubmissionPreflight{
		FromEmail:         "sender@example.test",
		RecipientCount:    1,
		MessageBytes:      42,
		BindingHash:       binding,
		PolicyBindingHash: policyBinding,
	}
	if preflight.FromEmail == "" || preflight.RecipientCount != 1 || preflight.MessageBytes != 42 || preflight.BindingHash == [sha256.Size]byte{} || preflight.PolicyBindingHash == [sha256.Size]byte{} {
		t.Fatalf("EngineSubmissionPreflight = %#v, want concrete nonzero projection", preflight)
	}
	assertExactFields(t, reflect.TypeOf(mail.EngineSubmissionPreflight{}), reflect.TypeOf(struct {
		FromEmail         string
		RecipientCount    int32
		MessageBytes      int64
		BindingHash       [sha256.Size]byte
		PolicyBindingHash [sha256.Size]byte
	}{}))
	assertNoFields(t, reflect.TypeOf(mail.EngineSubmissionPreflight{}), "To", "CC", "BCC", "Recipients", "InternalDomains", "Policy")
}

func TestMailEngine_UnknownErrorKindIsGenericAndDoesNotUnwrap(t *testing.T) {
	for _, unknownKind := range []mail.EngineErrorKind{
		"unknown synthetic-secret@example.test",
		"unknown synthetic message body",
		"unknown synthetic-address@example.test",
	} {
		err := &mail.EngineError{Kind: unknownKind}
		if got := err.Error(); got != "mail engine error" {
			t.Errorf("EngineError{%q}.Error() = %q, want fixed generic error", unknownKind, got)
		}
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			t.Errorf("errors.Unwrap(EngineError{%q}) = %v, want nil", unknownKind, unwrapped)
		}
	}
}

func TestMailEngine_IncrementalAndEnumerationPagesCarryBoundsAndTypedRecovery(t *testing.T) {
	changes := mail.EngineChangePage{
		AccountID:  "account-opaque",
		FromCursor: "from-cursor",
		NextCursor: "next-cursor",
		Bounds:     mail.EnginePageBounds{Requested: 50, Maximum: 100, Returned: 1},
	}
	enumeration := mail.EngineEnumerationPage{
		AccountID:      "account-opaque",
		BaseCursor:     "base-cursor",
		Checkpoint:     "checkpoint-1",
		NextCheckpoint: "checkpoint-2",
		Bounds:         mail.EnginePageBounds{Requested: 50, Maximum: 100, Returned: 1},
	}

	if changes.FromCursor == changes.NextCursor || enumeration.BaseCursor == "" || enumeration.NextCheckpoint == "" {
		t.Fatal("pages must preserve exact cursor and checkpoint metadata")
	}
	if !errors.Is(&mail.EngineError{Kind: mail.EngineErrorInvalidCursor}, mail.ErrEngineInvalidCursor) {
		t.Fatal("invalid cursor outcome must remain typed")
	}
	if !errors.Is(&mail.EngineError{Kind: mail.EngineErrorRestartRequired}, mail.ErrEngineRestartRequired) {
		t.Fatal("restart-required outcome must remain typed")
	}
}

func TestMailEngine_ProjectionOriginDoesNotInferOwnerFromAddressOrSentFolder(t *testing.T) {
	addressAndSentOnly := mail.EngineMessageProjection{
		AccountID: "account-opaque",
		MessageID: "message-opaque",
		Version:   "version-opaque",
		FolderIDs: []string{"sent-folder-opaque"},
		From:      mail.EngineAddress{Email: "owner@example.test"},
	}
	if addressAndSentOnly.Origin != "" {
		t.Fatal("From and Sent fields alone must not derive projection origin")
	}

	message := mail.EngineMessageProjection{
		AccountID:        "account-opaque",
		MessageID:        "message-opaque",
		Version:          "version-opaque",
		Origin:           mail.EngineProjectionOriginInboundOrUnknown,
		FolderIDs:        []string{"sent-folder-opaque"},
		From:             mail.EngineAddress{Email: "owner@example.test"},
		ReservedKeywords: []mail.EngineKeywordMarker{"reserved-marker"},
	}

	if message.Origin != mail.EngineProjectionOriginInboundOrUnknown {
		t.Fatal("address and Sent membership must not imply confirmed-owner origin")
	}
	origins := []mail.EngineProjectionOrigin{
		mail.EngineProjectionOriginInboundOrUnknown,
		mail.EngineProjectionOriginConfirmedOwnerSubmission,
		mail.EngineProjectionOriginConfirmedOwnerDraft,
	}
	for left := range origins {
		for right := left + 1; right < len(origins); right++ {
			if origins[left] == origins[right] {
				t.Fatalf("projection origins at indexes %d and %d must be distinct", left, right)
			}
		}
	}
	// Стабильное поведение mapper по подтверждённой provenance проверяет conformance-пакет.
}

func TestMailEngine_ConsumerCompileAssertionRemainsStable(t *testing.T) {
	var _ mail.MailEngine = (*consumerStub)(nil)
}

func TestMailEngine_ProvisioningSpecShapeCarriesOnlyStableDesiredState(t *testing.T) {
	spec := mail.AccountSpec{
		ProvisioningKey:  "stable-provisioning-key",
		LocalPart:        "u-0123456789ab",
		Domain:           "mail.example.test",
		DesiredGeneration: 7,
		DesiredHash:      "desired-hash",
		Enabled:          true,
		Primary:          mail.AccountIdentitySpec{LocalPart: "u-0123456789ab", Domain: "mail.example.test"},
	}
	if spec.ProvisioningKey == "" || spec.DesiredGeneration == 0 || spec.DesiredHash == "" {
		t.Fatal("provisioning shape must carry stable key, generation and desired hash")
	}
	assertNoFields(t, reflect.TypeOf(mail.AccountSpec{}), "Password", "Token", "Secret", "Credential")
	// Повторный ensure и conflict изменённого spec проверяет поведенческий conformance-пакет fake.
}

func TestMailEngine_SieveAndSpamTrainingUseBoundedTypedValues(t *testing.T) {
	script := mail.SieveScript{Name: "hub-managed", Generation: 9, Hash: "script-hash"}
	training := mail.EngineSpamTraining{
		AccountID: "account-opaque",
		ReceiptID: "receipt-id",
		CommandID: "command-id",
		Kind:      mail.EngineSpamTrainingSpam,
		Samples: []mail.EngineSpamTrainingSample{{
			MessageID: "message-opaque",
		}},
	}
	result := mail.EngineSpamTrainingResult{Outcomes: []mail.EngineSpamTrainingOutcome{{
		MessageID: "message-opaque",
		Status:    mail.EngineSpamTrainingAccepted,
	}}}

	if script.Name == "" || len(training.Samples) != 1 || result.Outcomes[0].Status != mail.EngineSpamTrainingAccepted {
		t.Fatal("Sieve and classifier values must stay explicit and typed")
	}
	assertNoFields(t, reflect.TypeOf(mail.SieveScript{}), "RawScript", "Body", "ProviderResponse")
}

func TestMailEngine_DomainStatesExposeOnlyPublicDNSRequirements(t *testing.T) {
	state := mail.EngineDomainState{
		DomainID: "domain-opaque",
		State:    mail.EngineDomainDNSRequired,
		DNSRecords: []mail.EngineDNSRecord{{
			Name:  "_dmarc.mail.example.test",
			Type:  mail.EngineDNSRecordTXT,
			Value: "v=DMARC1; p=none",
		}},
	}

	if state.State != mail.EngineDomainDNSRequired || len(state.DNSRecords) != 1 {
		t.Fatal("domain state must distinguish public DNS-required state")
	}
	if mail.EngineDomainVerified == mail.EngineDomainAbsent || mail.EngineDomainIndeterminate == mail.EngineDomainVerified {
		t.Fatal("domain verified, absent and indeterminate states must remain distinct")
	}
	assertNoFields(t, reflect.TypeOf(mail.EngineDomainState{}), "PrivateKey", "ProviderBody", "Credential", "Token")
}

func TestMailEngine_ReservedKeywordPatchIsDistinctFromPublicFlagsAndLabels(t *testing.T) {
	patch := mail.KeywordPatch{
		Add:    []mail.EngineKeywordMarker{"reserved-marker-add"},
		Remove: []mail.EngineKeywordMarker{"reserved-marker-remove"},
	}

	if len(patch.Add) != 1 || len(patch.Remove) != 1 {
		t.Fatal("reserved keyword patch must retain explicit add and remove sets")
	}
	assertNoFields(t, reflect.TypeOf(mail.KeywordPatch{}), "Label", "Flag", "UnknownKeyword")
	assertNoFields(t, reflect.TypeOf(mail.EngineMessage{}), "Labels", "PublicKeywords")
}

func TestMailEngine_HardLimitsMatchContract018(t *testing.T) {
	want := map[string]int{
		"opaque": 1024, "cursor": 4096, "page": 100,
		"projection_item": 1048576, "projection_page": 8388608,
		"message": 52428800, "attachment": 26214400,
		"projection_addresses": 200, "draft_recipients": 100,
		"draft_total_recipients": 300,
		"references": 100, "folders": 100, "keywords": 200,
		"attachments": 50, "email": 320, "display_name": 640,
		"header": 998, "snippet": 4096, "mime_field": 1020,
		"body_part": 8388608,
	}
	got := map[string]int{
		"opaque": mail.MaxEngineOpaqueIDBytes,
		"cursor": mail.MaxEngineCursorBytes,
		"page": mail.MaxEnginePageItems,
		"projection_item": mail.MaxEngineProjectionItemBytes,
		"projection_page": mail.MaxEngineProjectionPageBytes,
		"message": mail.MaxEngineMessageBytes,
		"attachment": mail.MaxEngineAttachmentBytes,
		"projection_addresses": mail.MaxEngineAddressItems,
		"draft_recipients": mail.MaxEngineDraftRecipientItems,
		"draft_total_recipients": mail.MaxEngineDraftTotalRecipientItems,
		"references": mail.MaxEngineReferenceItems,
		"folders": mail.MaxEngineFolderItems,
		"keywords": mail.MaxEngineKeywordItems,
		"attachments": mail.MaxEngineAttachmentItems,
		"email": mail.MaxEngineAddressEmailBytes,
		"display_name": mail.MaxEngineAddressNameBytes,
		"header": mail.MaxEngineHeaderValueBytes,
		"snippet": mail.MaxEngineSnippetBytes,
		"mime_field": mail.MaxEngineMIMEFieldBytes,
		"body_part": mail.MaxEngineBodyPartBytes,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hard limits = %#v, want %#v", got, want)
	}
}

func TestMailEngine_RequestAndPageShapesAreClosed(t *testing.T) {
	assertExactFields(t, reflect.TypeOf(mail.EnginePageRequest{}), reflect.TypeOf(struct {
		Limit int
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EnginePageBounds{}), reflect.TypeOf(struct {
		Requested int
		Maximum   int
		Returned  int
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineMessageQuery{}), reflect.TypeOf(struct {
		AccountID string
		FolderID  string
		Cursor    string
		Page      mail.EnginePageRequest
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineMessageReadRequest{}), reflect.TypeOf(struct {
		AccountID       string
		MessageID       string
		MaxContentBytes int64
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineBlobReadRequest{}), reflect.TypeOf(struct {
		AccountID string
		BlobID    string
		MaxBytes  int64
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineChangeRequest{}), reflect.TypeOf(struct {
		AccountID  string
		FromCursor string
		Page       mail.EnginePageRequest
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineEnumerationRequest{}), reflect.TypeOf(struct {
		AccountID  string
		BaseCursor string
		Checkpoint string
		Page       mail.EnginePageRequest
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineProjectionPage{}), reflect.TypeOf(struct {
		AccountID  string
		Cursor     string
		NextCursor string
		Bounds     mail.EnginePageBounds
		Messages   []mail.EngineMessageProjection
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineChangePage{}), reflect.TypeOf(struct {
		AccountID        string
		FromCursor       string
		NextCursor       string
		Bounds           mail.EnginePageBounds
		CreatedOrUpdated []mail.EngineMessageProjection
		DestroyedIDs     []string
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineEnumerationPage{}), reflect.TypeOf(struct {
		AccountID      string
		BaseCursor     string
		Checkpoint     string
		NextCheckpoint string
		Complete       bool
		Bounds         mail.EnginePageBounds
		Messages       []mail.EngineMessageProjection
	}{}))
}

func TestMailEngine_ContentAndSubmissionShapesAreClosed(t *testing.T) {
	assertExactFields(t, reflect.TypeOf(mail.EngineAddress{}), reflect.TypeOf(struct {
		Email string
		Name  string
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineMessageFlags{}), reflect.TypeOf(struct {
		Seen     bool
		Answered bool
		Draft    bool
		Flagged  bool
		Junk     bool
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineProjectionAttachment{}), reflect.TypeOf(struct {
		BlobID        string
		AttachmentKey string
		ContentID     string
		Filename      string
		MediaType     string
		SizeBytes     int64
		Disposition   mail.EngineAttachmentDisposition
		SHA256        []byte
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineMessageProjection{}), reflect.TypeOf(struct {
		AccountID               string
		MessageID               string
		Version                 string
		InternetMessageID       string
		InReplyTo               string
		References              []string
		From                    mail.EngineAddress
		To                      []mail.EngineAddress
		CC                      []mail.EngineAddress
		BCC                     []mail.EngineAddress
		ReplyTo                 []mail.EngineAddress
		Subject                 string
		Snippet                 string
		SentAt                  *time.Time
		ReceivedAt              *time.Time
		CalendarMethod          *mail.MailCalendarMethod
		AuthenticationClass     mail.MailAuthenticationClass
		CalendarPartLocatorHash []byte
		SizeBytes               int64
		Priority                mail.EngineMessagePriority
		FolderIDs               []string
		Flags                   mail.EngineMessageFlags
		ReservedKeywords        []mail.EngineKeywordMarker
		Attachments             []mail.EngineProjectionAttachment
		Origin                  mail.EngineProjectionOrigin
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineMessageBody{}), reflect.TypeOf(struct {
		Format    mail.EngineBodyFormat
		PlainBody string
		RichBody  string
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineMessage{}), reflect.TypeOf(struct {
		Projection mail.EngineMessageProjection
		Body       mail.EngineMessageBody
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineDraftAttachment{}), reflect.TypeOf(struct {
		AttachmentKey string
		ContentID     string
		Filename      string
		MediaType     string
		Disposition   mail.EngineAttachmentDisposition
		SHA256        [32]byte
		Bytes         []byte
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineMessageContent{}), reflect.TypeOf(struct {
		From        mail.EngineAddress
		To          []mail.EngineAddress
		CC          []mail.EngineAddress
		BCC         []mail.EngineAddress
		ReplyTo     []mail.EngineAddress
		Subject     string
		InReplyTo   string
		References  []string
		Format      mail.EngineBodyFormat
		PlainBody   string
		RichBody    string
		Priority    mail.EngineMessagePriority
		Attachments []mail.EngineDraftAttachment
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineBlob{}), reflect.TypeOf(struct {
		AccountID string
		BlobID    string
		MediaType string
		SizeBytes int64
		Bytes     []byte
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineDraft{}), reflect.TypeOf(struct {
		AccountID       string
		DraftKey        string
		MessageID       string
		PreviousVersion string
		MutationID      string
		MaxContentBytes int64
		Message         mail.EngineMessageContent
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineDraftReconciliation{}), reflect.TypeOf(struct {
		AccountID       string
		DraftKey        string
		MessageID       string
		PreviousVersion string
		MutationID      string
		MaxContentBytes int64
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineSubmission{}), reflect.TypeOf(struct {
		AccountID       string
		DraftKey        string
		MessageID       string
		ExpectedVersion string
		SubmissionKey   string
		MaxMessageBytes int64
	}{}))
	assertExactFields(t, reflect.TypeOf(mail.EngineSubmissionAccepted{}), reflect.TypeOf(struct {
		SubmissionID string
		MessageID    string
	}{}))
	assertNoFields(t, reflect.TypeOf(mail.EngineMessageProjection{}), "Content", "Body", "Bytes", "RawMIME")
	assertNoFields(t, reflect.TypeOf(mail.EngineSubmission{}), "Message", "Content", "Attachments", "Bytes", "RawMIME")
}

func TestMailEngine_Contract018EnumsAndPayloadErrorAreStable(t *testing.T) {
	if mail.EngineBodyPlain != "plain" || mail.EngineBodyRich != "rich" ||
		mail.EnginePriorityLow != "low" || mail.EnginePriorityNormal != "normal" ||
		mail.EnginePriorityHigh != "high" ||
		mail.EngineAttachmentDispositionAttachment != "attachment" ||
		mail.EngineAttachmentDispositionInline != "inline" ||
		mail.EngineProjectionOriginConfirmedOwnerDraft != "confirmed_owner_draft" {
		t.Fatal("0.1.8 enum value drift")
	}
	if mail.EngineProjectionOriginInboundOrUnknown != "inbound_or_unknown" {
		t.Fatal("inbound-or-unknown projection origin value drift")
	}
	if mail.EngineProjectionOriginConfirmedOwnerSubmission != "confirmed_owner_submission" {
		t.Fatal("confirmed-owner-submission projection origin value drift")
	}
	if mail.EngineErrorPayloadTooLarge != "payload_too_large" {
		t.Fatal("payload-too-large engine error kind value drift")
	}
	err := &mail.EngineError{Kind: mail.EngineErrorPayloadTooLarge}
	if !errors.Is(err, mail.ErrEnginePayloadTooLarge) {
		t.Fatal("payload-too-large must match its stable sentinel")
	}
	if got := err.Error(); got != "mail engine payload too large" {
		t.Fatalf("payload error = %q", got)
	}
}

func assertExactFields(t *testing.T, valueType, expectedType reflect.Type) {
	t.Helper()
	if valueType.NumField() != expectedType.NumField() {
		t.Fatalf("%s field count = %d, want %d", valueType.Name(), valueType.NumField(), expectedType.NumField())
	}
	for index := 0; index < expectedType.NumField(); index++ {
		actual := valueType.Field(index)
		expected := expectedType.Field(index)
		if actual.Name != expected.Name || actual.Type != expected.Type || actual.Tag != expected.Tag || actual.Anonymous != expected.Anonymous {
			t.Fatalf("%s field %d = {Name:%q Type:%s Tag:%q Anonymous:%t}, want {Name:%q Type:%s Tag:%q Anonymous:%t}",
				valueType.Name(), index, actual.Name, actual.Type, actual.Tag, actual.Anonymous,
				expected.Name, expected.Type, expected.Tag, expected.Anonymous)
		}
	}
}

func assertNoFields(t *testing.T, valueType reflect.Type, forbiddenNames ...string) {
	t.Helper()
	for _, forbiddenName := range forbiddenNames {
		if _, found := valueType.FieldByName(forbiddenName); found {
			t.Fatalf("%s must not expose field %s", valueType.Name(), forbiddenName)
		}
	}
}

func assertNoSensitiveFieldNames(t *testing.T, valueType reflect.Type) {
	t.Helper()
	for index := 0; index < valueType.NumField(); index++ {
		fieldName := strings.ToLower(valueType.Field(index).Name)
		for _, forbidden := range []string{"credential", "token", "password", "secret", "apikey"} {
			if strings.Contains(fieldName, forbidden) {
				t.Fatalf("%s must not expose %q", valueType.Name(), valueType.Field(index).Name)
			}
		}
	}
}
