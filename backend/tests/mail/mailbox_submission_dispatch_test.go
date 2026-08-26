package mail_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/mail"
)

type mailboxSubmissionDispatchRepositoryStub struct{}

func (*mailboxSubmissionDispatchRepositoryStub) ClaimMailboxSubmission(context.Context, mail.MailboxSubmissionClaimRequest) (mail.MailboxSubmissionClaimResult, error) {
	return mail.MailboxSubmissionClaimResult{}, nil
}

func (*mailboxSubmissionDispatchRepositoryStub) AuthorizeMailboxSubmission(context.Context, mail.MailboxSubmissionAuthorizeRequest) (mail.MailboxSubmissionAuthorizeResult, error) {
	return mail.MailboxSubmissionAuthorizeResult{}, nil
}

func (*mailboxSubmissionDispatchRepositoryStub) ResolveMailboxSubmissionPreflightFailure(context.Context, mail.MailboxSubmissionPreflightFailureRequest) (mail.MailboxSubmissionAuthorizeResult, error) {
	return mail.MailboxSubmissionAuthorizeResult{}, nil
}

func (*mailboxSubmissionDispatchRepositoryStub) ResolveMailboxSubmissionDispatch(context.Context, mail.MailboxSubmissionDispatchResolutionRequest) error {
	return nil
}

var _ mail.MailboxSubmissionDispatchRepository = (*mailboxSubmissionDispatchRepositoryStub)(nil)

func TestMailboxSubmissionDispatchContract_SealsProviderMutation(t *testing.T) {
	resolutionType := reflect.TypeOf(mail.MailboxSubmissionDispatchResolution(""))
	if mail.MailboxSubmissionDispatchResolutionAccepted != "accepted" || mail.MailboxSubmissionDispatchResolutionUnknown != "unknown" {
		t.Fatalf("dispatch resolutions = %q, %q, want accepted and unknown", mail.MailboxSubmissionDispatchResolutionAccepted, mail.MailboxSubmissionDispatchResolutionUnknown)
	}

	requestType := reflect.TypeOf(mail.MailboxSubmissionDispatchResolutionRequest{})
	wantFields := []struct {
		name  string
		type_ reflect.Type
	}{
		{name: "Authorization", type_: reflect.TypeOf(mail.MailboxSubmissionDispatchAuthorization{})},
		{name: "Resolution", type_: resolutionType},
		{name: "ProviderSubmissionID", type_: reflect.TypeOf("")},
		{name: "ProviderMessageID", type_: reflect.TypeOf("")},
	}
	if requestType.NumField() != len(wantFields) {
		t.Fatalf("MailboxSubmissionDispatchResolutionRequest field count = %d, want %d", requestType.NumField(), len(wantFields))
	}
	for index, want := range wantFields {
		field := requestType.Field(index)
		if field.Name != want.name || field.Type != want.type_ || field.PkgPath != "" {
			t.Fatalf("MailboxSubmissionDispatchResolutionRequest field %d = %s %s exported=%t, want %s %s exported", index, field.Name, field.Type, field.PkgPath == "", want.name, want.type_)
		}
	}

	repositoryType := reflect.TypeOf((*mail.MailboxSubmissionDispatchRepository)(nil)).Elem()
	resolveMethod, exists := repositoryType.MethodByName("ResolveMailboxSubmissionDispatch")
	if !exists || resolveMethod.Type.NumIn() != 2 || resolveMethod.Type.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() || resolveMethod.Type.In(1) != requestType || resolveMethod.Type.NumOut() != 1 || resolveMethod.Type.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
		t.Fatalf("ResolveMailboxSubmissionDispatch signature = %v, want func(context.Context, MailboxSubmissionDispatchResolutionRequest) error", resolveMethod.Type)
	}

	engineType := reflect.TypeOf((*mail.MailEngine)(nil)).Elem()
	if _, exists := engineType.MethodByName("Submit"); exists {
		t.Fatal("MailEngine must not expose provider submission")
	}
	for _, candidate := range []reflect.Type{
		reflect.TypeOf((*mail.FakeMailEngine)(nil)),
		reflect.TypeOf((*mail.StalwartEngine)(nil)),
	} {
		if _, exists := candidate.MethodByName("Submit"); exists {
			t.Fatalf("%s must not expose provider submission", candidate)
		}
	}

	dispatcherType := reflect.TypeOf((*mail.MailboxSubmissionDispatcher)(nil))
	dispatchMethod, exists := dispatcherType.MethodByName("Dispatch")
	if !exists || dispatchMethod.Type.NumIn() != 3 || dispatchMethod.Type.In(1) != reflect.TypeOf((*context.Context)(nil)).Elem() || dispatchMethod.Type.In(2) != reflect.TypeOf(uuid.UUID{}) || dispatchMethod.Type.NumOut() != 1 || dispatchMethod.Type.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
		t.Fatalf("Dispatch signature = %v, want func(*MailboxSubmissionDispatcher, context.Context, uuid.UUID) error", dispatchMethod.Type)
	}
}

func TestMailboxSubmissionDispatcherRejectsTypedNilDependencies(t *testing.T) {
	repository := &mailboxSubmissionDispatchRepositoryStub{}
	var typedNilEngine *mail.FakeMailEngine
	if _, err := mail.NewMailboxSubmissionDispatcher(repository, typedNilEngine); !mailboxSubmissionDispatchHasKind(err, mail.MailboxSubmissionDispatchErrorInvalid) {
		t.Fatalf("typed-nil engine error = %v, want invalid dispatcher error", err)
	}

	engine := mail.NewFakeMailEngine(mail.FakeMailEngineConfig{})
	var typedNilRepository *mailboxSubmissionDispatchRepositoryStub
	if _, err := mail.NewMailboxSubmissionDispatcher(typedNilRepository, engine); !mailboxSubmissionDispatchHasKind(err, mail.MailboxSubmissionDispatchErrorInvalid) {
		t.Fatalf("typed-nil repository error = %v, want invalid dispatcher error", err)
	}
}

func mailboxSubmissionDispatchHasKind(err error, want mail.MailboxSubmissionDispatchErrorKind) bool {
	var dispatchError *mail.MailboxSubmissionDispatchError
	return errors.As(err, &dispatchError) && dispatchError.Kind() == want
}

func mailboxSubmissionDispatchSource(t *testing.T, path string) string {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read dispatcher source %q: %v", path, err)
	}
	return string(source)
}

func mailboxSubmissionDispatchSourceSlice(t *testing.T, source, start, end string) string {
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

func requireMailboxSubmissionDispatchOrdered(t *testing.T, source string, fragments ...string) {
	t.Helper()
	previous := -1
	for _, fragment := range fragments {
		position := strings.Index(source, fragment)
		if position < 0 {
			t.Fatalf("source is missing %q", fragment)
		}
		if position <= previous {
			t.Fatalf("source places %q before its required predecessor", fragment)
		}
		previous = position
	}
}

func TestMailboxSubmissionDispatcherSourceContract(t *testing.T) {
	dispatcherSource := mailboxSubmissionDispatchSource(t, "../../internal/mail/mailbox_submission_dispatcher.go")
	dispatch := mailboxSubmissionDispatchSourceSlice(t, dispatcherSource,
		"func (dispatcher *MailboxSubmissionDispatcher) Dispatch",
		"func (dispatcher *MailboxSubmissionDispatcher) resolvePreflightFailure",
	)
	requireMailboxSubmissionDispatchOrdered(t, dispatch,
		"ClaimMailboxSubmission(",
		"claim.Policy().PreflightRequest(",
		"dispatcher.engine.PreflightSubmission(",
		"AuthorizeMailboxSubmission(",
		"dispatcher.dispatchAuthorized(",
	)
	busy := strings.Index(dispatch, "claimResult.BusyUntil()")
	if busy < 0 || strings.Index(dispatch[busy:], "return nil") < 0 {
		t.Fatal("busy claim must be a no-op before any provider mutation")
	}

	authorized := mailboxSubmissionDispatchSourceSlice(t, dispatcherSource,
		"func (dispatcher *MailboxSubmissionDispatcher) dispatchAuthorized",
		"func (dispatcher *MailboxSubmissionDispatcher) resolve",
	)
	if calls := strings.Count(authorized, "dispatcher.engine.submitMailboxSubmission("); calls != 1 {
		t.Fatalf("authorized dispatch private submit calls = %d, want exactly one", calls)
	}
	for _, required := range []string{
		"result.(EngineSubmissionAccepted)",
		"validOpaqueID(accepted.SubmissionID)",
		"validOpaqueID(accepted.MessageID)",
		"MailboxSubmissionDispatchResolutionAccepted",
	} {
		if !strings.Contains(authorized, required) {
			t.Fatalf("authorized dispatch is missing %q", required)
		}
	}
	if unknowns := strings.Count(authorized, "MailboxSubmissionDispatchResolutionUnknown"); unknowns < 3 {
		t.Fatalf("authorized dispatch unknown resolutions = %d, want cancellation/error/malformed coverage", unknowns)
	}

	resolve := mailboxSubmissionDispatchSourceSlice(t, dispatcherSource,
		"func (dispatcher *MailboxSubmissionDispatcher) resolve(ctx",
		"\\n}",
	)
	if !strings.Contains(resolve, "context.WithoutCancel(ctx)") || !strings.Contains(resolve, "MailboxSubmissionDispatchErrorReconciliationRequired") {
		t.Fatal("post-seal resolution must persist with cancellation removed and surface reconciliation_required only")
	}

	dispatchSource := mailboxSubmissionDispatchSource(t, "../../internal/mail/mailbox_submission_dispatch.go")
	preflightMapping := mailboxSubmissionDispatchSourceSlice(t, dispatchSource,
		"func mailboxSubmissionDispatchPreflightFailure",
		"\\n}",
	)
	for _, required := range []string{
		"ErrEngineNotFound", "ErrEngineConflict", "MailboxSubmissionErrorCodeRejected",
		"ErrEngineRejectedPolicy", "MailboxSubmissionErrorCodePolicy",
		"ErrEnginePayloadTooLarge", "MailboxSubmissionErrorCodeSizeLimit",
		"ErrEngineRecipientLimit", "MailboxSubmissionErrorCodeRecipientLimit",
		"ErrEngineIdentityUnavailable", "MailboxSubmissionErrorCodeIdentityUnavailable",
		"ErrEngineRateLimited", "MailboxSubmissionPreflightFailureDispositionRateLimited",
	} {
		if !strings.Contains(preflightMapping, required) {
			t.Fatalf("deterministic preflight mapping is missing %q", required)
		}
	}
	if delays := strings.Count(preflightMapping, "request.RetryAfterSeconds = 60"); delays != 2 {
		t.Fatalf("deterministic preflight delays = %d, want rate-limit and retry at 60 seconds", delays)
	}

	engineInterface := mailboxSubmissionDispatchSourceSlice(t, dispatcherSource,
		"type mailboxSubmissionEngine interface {",
		"\\n}",
	)
	for _, required := range []string{"PreflightSubmission", "submitMailboxSubmission"} {
		if !strings.Contains(engineInterface, required) {
			t.Fatalf("private submission engine is missing %q", required)
		}
	}
	constructor := mailboxSubmissionDispatchSourceSlice(t, dispatcherSource,
		"func NewMailboxSubmissionDispatcher",
		"func mailboxSubmissionTypedNil",
	)
	for _, required := range []string{"mailboxSubmissionTypedNil(repository)", "mailboxSubmissionTypedNil(engine)"} {
		if !strings.Contains(constructor, required) {
			t.Fatalf("constructor typed-nil guard is missing %q", required)
		}
	}

	for path, forbidden := range map[string][]string{
		"../../internal/mail/mailbox_engine.go":          {"\\tSubmit(context.Context, EngineSubmission)"},
		"../../internal/mail/mailbox_fake.go":            {"func (f *FakeMailEngine) Submit(", "SubmitForTest"},
		"../../internal/mail/stalwart_engine_drafts.go": {"func (engine *StalwartEngine) Submit("},
	} {
		source := mailboxSubmissionDispatchSource(t, path)
		for _, token := range forbidden {
			if strings.Contains(source, token) {
				t.Fatalf("%s still exposes forbidden public submission token %q", path, token)
			}
		}
	}
}
