package mail_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	mail "github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const identityReconcileAccountID = "account-a"

func TestStalwartEngineObserveMailboxIdentityReconcile_ReadsBoundedCanonicalSnapshot(t *testing.T) {
	client := &identityReconcileFixtureClient{
		accountID: identityReconcileAccountID,
		response: identityReconcileResponse(`{"accountId":"account-a","state":"identity-state-a","list":[{"id":"identity-z","email":"Zed@EXAMPLE.test"},{"id":"identity-a","email":"Ada@EXAMPLE.test"}],"notFound":[]}`),
	}
	factory := &identityReconcileFixtureFactory{client: client}
	engine, err := mail.NewStalwartEngine(factory)
	if err != nil {
		t.Fatalf("NewStalwartEngine() error = %v", err)
	}

	snapshot, err := engine.ObserveMailboxIdentityReconcile(context.Background(), mail.MailboxIdentityReconcileRequest{EngineAccountID: identityReconcileAccountID})
	if err != nil {
		t.Fatalf("ObserveMailboxIdentityReconcile() error = %v", err)
	}
	if snapshot.EngineAccountID != identityReconcileAccountID {
		t.Fatalf("EngineAccountID = %q, want %q", snapshot.EngineAccountID, identityReconcileAccountID)
	}
	want := []mail.MailboxIdentityReconcileIdentity{
		{EngineIdentityID: "identity-a", Email: "ada@example.test"},
		{EngineIdentityID: "identity-z", Email: "zed@example.test"},
	}
	if !equalIdentityReconcileIdentities(snapshot.Identities, want) {
		t.Fatalf("Identities = %#v, want %#v", snapshot.Identities, want)
	}
	if got := strings.Join(factory.opened, ","); got != identityReconcileAccountID {
		t.Fatalf("opened accounts = %q, want exact account", got)
	}
	if client.unboundedCalls != 0 {
		t.Fatalf("CallMailbox calls = %d, want 0", client.unboundedCalls)
	}
	if len(client.boundedCalls) != 1 {
		t.Fatalf("CallMailboxBounded calls = %d, want 1", len(client.boundedCalls))
	}
	assertIdentityReconcileRequest(t, client.boundedCalls[0])
}

func TestStalwartEngineObserveMailboxIdentityReconcile_RejectsUntrustedSnapshot(t *testing.T) {
	overlongState := strings.Repeat("s", mail.MaxCalendarPartBytes)
	tooMany := make([]identityReconcileFixtureIdentity, mail.MaxEnginePageItems+1)
	for index := range tooMany {
		tooMany[index] = identityReconcileFixtureIdentity{ID: "identity-" + string(rune('a'+index)), Email: "person" + string(rune('a'+index)) + "@example.test"}
	}
	tooManyPayload, err := json.Marshal(struct {
		AccountID string                              `json:"accountId"`
		State     string                              `json:"state"`
		List      []identityReconcileFixtureIdentity `json:"list"`
		NotFound  []string                            `json:"notFound"`
	}{AccountID: identityReconcileAccountID, State: "identity-state-a", List: tooMany, NotFound: []string{}})
	if err != nil {
		t.Fatalf("marshal too-many fixture: %v", err)
	}

	for _, testCase := range []struct {
		name      string
		client    mail.StalwartMailboxClient
		wantError error
	}{
		{
			name: "missing bounded capability",
			client: &identityReconcileUnboundedFixtureClient{accountID: identityReconcileAccountID},
			wantError: mail.ErrEngineUnavailable,
		},
		{
			name: "foreign account",
			client: &identityReconcileFixtureClient{accountID: identityReconcileAccountID, response: identityReconcileResponse(`{"accountId":"other-account","state":"identity-state-a","list":[],"notFound":[]}`)},
			wantError: mail.ErrEngineUnauthorizedAccount,
		},
		{
			name: "missing account",
			client: &identityReconcileFixtureClient{accountID: identityReconcileAccountID, response: identityReconcileResponse(`{"state":"identity-state-a","list":[],"notFound":[]}`)},
			wantError: mail.ErrEngineUnavailable,
		},
		{
			name: "unknown field",
			client: &identityReconcileFixtureClient{accountID: identityReconcileAccountID, response: identityReconcileResponse(`{"accountId":"account-a","state":"identity-state-a","list":[],"notFound":[],"providerDetail":"must-not-pass"}`)},
			wantError: mail.ErrEngineUnavailable,
		},
		{
			name: "trailing value",
			client: &identityReconcileFixtureClient{accountID: identityReconcileAccountID, response: identityReconcileResponse(`{"accountId":"account-a","state":"identity-state-a","list":[],"notFound":[]} {}`)},
			wantError: mail.ErrEngineUnavailable,
		},
		{
			name: "wrong list type",
			client: &identityReconcileFixtureClient{accountID: identityReconcileAccountID, response: identityReconcileResponse(`{"accountId":"account-a","state":"identity-state-a","list":{},"notFound":[]}`)},
			wantError: mail.ErrEngineUnavailable,
		},
		{
			name: "duplicate canonical provider email",
			client: &identityReconcileFixtureClient{accountID: identityReconcileAccountID, response: identityReconcileResponse(`{"accountId":"account-a","state":"identity-state-a","list":[{"id":"identity-a","email":"Ada@example.test"},{"id":"identity-b","email":"ada@EXAMPLE.test"}],"notFound":[]}`)},
			wantError: mail.ErrEngineUnavailable,
		},
		{
			name: "duplicate provider identity id",
			client: &identityReconcileFixtureClient{accountID: identityReconcileAccountID, response: identityReconcileResponse(`{"accountId":"account-a","state":"identity-state-a","list":[{"id":"identity-a","email":"ada@example.test"},{"id":"identity-a","email":"bea@example.test"}],"notFound":[]}`)},
			wantError: mail.ErrEngineUnavailable,
		},
		{
			name: "too many identities",
			client: &identityReconcileFixtureClient{accountID: identityReconcileAccountID, response: identityReconcileResponse(string(tooManyPayload))},
			wantError: mail.ErrEngineUnavailable,
		},
		{
			name: "oversized method payload",
			client: &identityReconcileFixtureClient{accountID: identityReconcileAccountID, response: identityReconcileResponse(`{"accountId":"account-a","state":"` + overlongState + `","list":[],"notFound":[]}`)},
			wantError: mail.ErrEngineUnavailable,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			factory := &identityReconcileFixtureFactory{client: testCase.client}
			engine, err := mail.NewStalwartEngine(factory)
			if err != nil {
				t.Fatalf("NewStalwartEngine() error = %v", err)
			}
			_, err = engine.ObserveMailboxIdentityReconcile(context.Background(), mail.MailboxIdentityReconcileRequest{EngineAccountID: identityReconcileAccountID})
			if !errors.Is(err, testCase.wantError) {
				t.Fatalf("ObserveMailboxIdentityReconcile() error = %v, want %v", err, testCase.wantError)
			}
			if err != nil && strings.Contains(err.Error(), "providerDetail") {
				t.Fatalf("ObserveMailboxIdentityReconcile() leaked provider payload in %q", err)
			}
			if bounded, ok := testCase.client.(*identityReconcileFixtureClient); ok {
				if len(bounded.boundedCalls) != 1 || bounded.unboundedCalls != 0 {
					t.Fatalf("identity transport calls = bounded:%d unbounded:%d, want exactly one bounded Identity/get", len(bounded.boundedCalls), bounded.unboundedCalls)
				}
				assertIdentityReconcileRequest(t, bounded.boundedCalls[0])
			}
		})
	}
}

type identityReconcileFixtureIdentity struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type identityReconcileBoundedCall struct {
	request stalwart.JMAPRequest
	maximum int64
}

type identityReconcileFixtureClient struct {
	accountID      string
	response       stalwart.JMAPResponse
	boundedErr     error
	boundedCalls   []identityReconcileBoundedCall
	unboundedCalls int
}

func (client *identityReconcileFixtureClient) AccountID() string {
	return client.accountID
}

func (client *identityReconcileFixtureClient) CallMailbox(context.Context, stalwart.JMAPRequest) (stalwart.JMAPResponse, error) {
	client.unboundedCalls++
	return stalwart.JMAPResponse{}, errors.New("unbounded mailbox call is forbidden")
}

func (client *identityReconcileFixtureClient) CallMailboxBounded(_ context.Context, request stalwart.JMAPRequest, maximum int64) (stalwart.JMAPResponse, error) {
	client.boundedCalls = append(client.boundedCalls, identityReconcileBoundedCall{request: request, maximum: maximum})
	if client.boundedErr != nil {
		return stalwart.JMAPResponse{}, client.boundedErr
	}
	return client.response, nil
}

func (*identityReconcileFixtureClient) UploadBlob(context.Context, mail.StalwartMailboxUploadRequest) (mail.StalwartMailboxUploadResult, error) {
	return mail.StalwartMailboxUploadResult{}, errors.New("upload is forbidden")
}

func (*identityReconcileFixtureClient) DownloadBlob(context.Context, string, int64) (mail.StalwartMailboxBlob, error) {
	return mail.StalwartMailboxBlob{}, errors.New("download is forbidden")
}

type identityReconcileUnboundedFixtureClient struct {
	accountID      string
	unboundedCalls int
}

func (client *identityReconcileUnboundedFixtureClient) AccountID() string {
	return client.accountID
}

func (client *identityReconcileUnboundedFixtureClient) CallMailbox(context.Context, stalwart.JMAPRequest) (stalwart.JMAPResponse, error) {
	client.unboundedCalls++
	return stalwart.JMAPResponse{}, errors.New("unbounded mailbox call is forbidden")
}

func (*identityReconcileUnboundedFixtureClient) UploadBlob(context.Context, mail.StalwartMailboxUploadRequest) (mail.StalwartMailboxUploadResult, error) {
	return mail.StalwartMailboxUploadResult{}, errors.New("upload is forbidden")
}

func (*identityReconcileUnboundedFixtureClient) DownloadBlob(context.Context, string, int64) (mail.StalwartMailboxBlob, error) {
	return mail.StalwartMailboxBlob{}, errors.New("download is forbidden")
}

type identityReconcileFixtureFactory struct {
	client mail.StalwartMailboxClient
	opened []string
}

func (factory *identityReconcileFixtureFactory) OpenMailboxClient(_ context.Context, accountID string) (mail.StalwartMailboxClient, error) {
	factory.opened = append(factory.opened, accountID)
	return factory.client, nil
}

func identityReconcileResponse(arguments string) stalwart.JMAPResponse {
	return stalwart.JMAPResponse{MethodResponses: []stalwart.JMAPMethodResponse{{
		Name: "Identity/get", Arguments: json.RawMessage(arguments), CallID: "identity-reconcile-get",
	}}}
}

func assertIdentityReconcileRequest(t *testing.T, call identityReconcileBoundedCall) {
	t.Helper()
	if call.maximum != mail.MaxCalendarPartBytes {
		t.Fatalf("bounded response maximum = %d, want %d", call.maximum, mail.MaxCalendarPartBytes)
	}
	if !equalIdentityReconcileStrings(call.request.Using, []string{
		"urn:ietf:params:jmap:core",
		"urn:ietf:params:jmap:mail",
		"urn:ietf:params:jmap:submission",
	}) {
		t.Fatalf("JMAP capabilities = %#v, want exact mailbox capabilities", call.request.Using)
	}
	if len(call.request.MethodCalls) != 1 {
		t.Fatalf("JMAP methods = %#v, want one read-only identity request", call.request.MethodCalls)
	}
	method := call.request.MethodCalls[0]
	if method.Name != "Identity/get" || method.CallID != "identity-reconcile-get" {
		t.Fatalf("JMAP method = %#v, want exact Identity/get", method)
	}
	var arguments struct {
		AccountID  string   `json:"accountId"`
		Properties []string `json:"properties"`
	}
	if err := json.Unmarshal(method.Arguments, &arguments); err != nil {
		t.Fatalf("Identity/get arguments: %v", err)
	}
	if arguments.AccountID != identityReconcileAccountID || !equalIdentityReconcileStrings(arguments.Properties, []string{"id", "email"}) {
		t.Fatalf("Identity/get arguments = %#v, want exact account/id/email", arguments)
	}
}

func equalIdentityReconcileIdentities(left, right []mail.MailboxIdentityReconcileIdentity) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalIdentityReconcileStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
