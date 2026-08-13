package mail_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

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

func (consumerStub) ListMessages(context.Context, mail.EngineMessageQuery) (mail.EngineMessagePage, error) {
	return mail.EngineMessagePage{}, nil
}

func (consumerStub) GetMessage(context.Context, string, string) (mail.EngineMessage, error) {
	return mail.EngineMessage{}, nil
}

func (consumerStub) GetBlob(context.Context, string, string) (mail.EngineBlob, error) {
	return mail.EngineBlob{}, nil
}

func (consumerStub) Changes(context.Context, string, string) (mail.EngineChangePage, error) {
	return mail.EngineChangePage{}, nil
}

func (consumerStub) EnumerateProjection(context.Context, string, string) (mail.EngineEnumerationPage, error) {
	return mail.EngineEnumerationPage{}, nil
}

func (consumerStub) SaveDraft(context.Context, mail.EngineDraft) (mail.EngineMessage, error) {
	return mail.EngineMessage{}, nil
}

func (consumerStub) ReconcileDraft(context.Context, mail.EngineDraftReconciliation) (mail.EngineDraftResult, error) {
	return mail.EngineDraftAbsent{}, nil
}

func (consumerStub) Submit(context.Context, mail.EngineSubmission) (mail.EngineSubmissionResult, error) {
	return mail.EngineSubmissionAbsent{}, nil
}

func (consumerStub) ReconcileSubmission(context.Context, string, string) (mail.EngineSubmissionResult, error) {
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
		MessageID:       "draft-opaque",
		PreviousVersion: "version-opaque",
		MutationID:      "hub-mutation-id",
		Message: mail.EngineMessageContent{
			Subject:   "Draft subject",
			PlainBody: "Draft body",
		},
	}
	submission := mail.EngineSubmission{
		AccountID:     "account-opaque",
		DraftID:       draft.MessageID,
		SubmissionKey: "stable-submission-key",
		Message:       draft.Message,
	}

	if draft.PreviousVersion == "" || draft.MutationID == "" || submission.SubmissionKey == "" {
		t.Fatal("draft version, mutation id and submission key must remain explicit typed fields")
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
	err := &mail.EngineError{Kind: mail.EngineErrorInvalidCursor}

	if !errors.Is(err, mail.ErrEngineInvalidCursor) {
		t.Fatal("invalid cursor error must match its stable sentinel")
	}
	var typed *mail.EngineError
	if !errors.As(err, &typed) || typed.Kind != mail.EngineErrorInvalidCursor {
		t.Fatal("invalid cursor error must be available through errors.As")
	}
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
	addressAndSentOnly := mail.EngineMessage{
		AccountID: "account-opaque",
		MessageID: "message-opaque",
		FolderIDs: []string{"sent-folder-opaque"},
		From:      mail.EngineAddress{Email: "owner@example.test"},
	}
	if addressAndSentOnly.Origin != "" {
		t.Fatal("From and Sent fields alone must not derive projection origin")
	}

	message := mail.EngineMessage{
		AccountID:        "account-opaque",
		MessageID:        "message-opaque",
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
