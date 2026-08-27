package mail_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

func TestRegistryAccountClient_EnsureLocalUserAccount(t *testing.T) {
	t.Run("returns exactly one existing matching user", func(t *testing.T) {
		fixture := newRegistryAccountHTTPFixture(t, []string{
			`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":["account-1"]},"registry-account-query"]]}`,
			`{"methodResponses":[["x:Account/get",{"accountId":"0","list":[{"id":"account-1","@type":"User","name":"ada","domainId":"domain-1","emailAddress":"ada@example.test"}],"notFound":[]},"registry-account-get"]]}`,
		})

		account, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), stalwart.LocalUserAccountSpec{
			EmailLocalPart: "ada",
			DomainID:        "domain-1",
			DomainName:      "example.test",
		})
		if err != nil {
			t.Fatalf("EnsureLocalUserAccount() error = %v", err)
		}
		if account.AccountID != "account-1" || account.EmailAddress != "ada@example.test" {
			t.Fatalf("EnsureLocalUserAccount() = %#v, want account-1 / ada@example.test", account)
		}
		fixture.requirePayloads(t,
			`{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/query",{"filter":{"operator":"AND","conditions":[{"name":"ada"},{"domainId":"domain-1"}]},"limit":2,"sort":[{"property":"id","isAscending":true}]},"registry-account-query"]]}`,
			`{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/get",{"ids":["account-1"],"properties":["id","@type","name","domainId","emailAddress"]},"registry-account-get"]]}`,
		)
	})

	t.Run("creates the documented minimal user then verifies it", func(t *testing.T) {
		fixture := newRegistryAccountHTTPFixture(t, []string{
			`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":[]},"registry-account-query"]]}`,
			`{"methodResponses":[["x:Account/set",{"accountId":"0","created":{"hub-local-user":{"id":"account-2"}}},"registry-account-set"]]}`,
			`{"methodResponses":[["x:Account/get",{"accountId":"0","list":[{"id":"account-2","@type":"User","name":"ada","domainId":"domain-1","emailAddress":"ada@example.test"}],"notFound":[]},"registry-account-get"]]}`,
		})

		account, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), stalwart.LocalUserAccountSpec{
			EmailLocalPart: "ada",
			DomainID:        "domain-1",
			DomainName:      "example.test",
		})
		if err != nil {
			t.Fatalf("EnsureLocalUserAccount() error = %v", err)
		}
		if account.AccountID != "account-2" || account.EmailAddress != "ada@example.test" {
			t.Fatalf("EnsureLocalUserAccount() = %#v, want account-2 / ada@example.test", account)
		}
		fixture.requirePayloads(t,
			`{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/query",{"filter":{"operator":"AND","conditions":[{"name":"ada"},{"domainId":"domain-1"}]},"limit":2,"sort":[{"property":"id","isAscending":true}]},"registry-account-query"]]}`,
			`{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/set",{"create":{"hub-local-user":{"@type":"User","name":"ada","domainId":"domain-1","credentials":{},"memberGroupIds":{},"roles":{"@type":"User"},"permissions":{"@type":"Inherit"},"quotas":{},"aliases":{},"encryptionAtRest":{"@type":"Disabled"}}}},"registry-account-set"]]}`,
			`{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/get",{"ids":["account-2"],"properties":["id","@type","name","domainId","emailAddress"]},"registry-account-get"]]}`,
		)
	})

	t.Run("returns the server address when no safe domain name was supplied", func(t *testing.T) {
		fixture := newRegistryAccountHTTPFixture(t, []string{
			`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":["account-1"]},"registry-account-query"]]}`,
			`{"methodResponses":[["x:Account/get",{"accountId":"0","list":[{"id":"account-1","@type":"User","name":"ada","domainId":"domain-1","emailAddress":"provider@example.test"}],"notFound":[]},"registry-account-get"]]}`,
		})

		account, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), stalwart.LocalUserAccountSpec{EmailLocalPart: "ada", DomainID: "domain-1"})
		if err != nil {
			t.Fatalf("EnsureLocalUserAccount() error = %v", err)
		}
		if account.AccountID != "account-1" || account.EmailAddress != "provider@example.test" {
			t.Fatalf("EnsureLocalUserAccount() = %#v, want server-confirmed address without fabricated comparison", account)
		}
	})

	t.Run("rejects duplicate provider identities without creating", func(t *testing.T) {
		fixture := newRegistryAccountHTTPFixture(t, []string{
			`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":["account-1","account-2"]},"registry-account-query"]]}`,
		})

		_, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), stalwart.LocalUserAccountSpec{EmailLocalPart: "ada", DomainID: "domain-1"})
		if !errors.Is(err, stalwart.ErrProtocol) || strings.Contains(err.Error(), "account-1") {
			t.Fatalf("EnsureLocalUserAccount() error = %v, want redacted ErrProtocol", err)
		}
		fixture.requirePayloads(t,
			`{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/query",{"filter":{"operator":"AND","conditions":[{"name":"ada"},{"domainId":"domain-1"}]},"limit":2,"sort":[{"property":"id","isAscending":true}]},"registry-account-query"]]}`,
		)
	})

	t.Run("rejects a user that does not match the requested identity", func(t *testing.T) {
		for _, account := range []string{
			`{"id":"account-1","@type":"Group","name":"ada","domainId":"domain-1","emailAddress":"provider-secret@example.test"}`,
			`{"id":"account-1","@type":"User","name":"other","domainId":"domain-1","emailAddress":"provider-secret@example.test"}`,
			`{"id":"account-1","@type":"User","name":"ada","domainId":"other-domain","emailAddress":"provider-secret@example.test"}`,
		} {
			fixture := newRegistryAccountHTTPFixture(t, []string{
				`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":["account-1"]},"registry-account-query"]]}`,
				`{"methodResponses":[["x:Account/get",{"accountId":"0","list":[` + account + `],"notFound":[]},"registry-account-get"]]}`,
			})

			_, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), stalwart.LocalUserAccountSpec{EmailLocalPart: "ada", DomainID: "domain-1"})
			if !errors.Is(err, stalwart.ErrStateMismatch) || strings.Contains(err.Error(), "provider-secret") {
				t.Fatalf("EnsureLocalUserAccount() error = %v, want redacted ErrStateMismatch", err)
			}
		}
	})

	t.Run("rejects malformed JMAP response", func(t *testing.T) {
		fixture := newRegistryAccountHTTPFixture(t, []string{
			`{"methodResponses":[["x:Account/query",{"accountId":"0","ids":["account-1"]},"registry-account-query"]]}`,
		})

		_, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), stalwart.LocalUserAccountSpec{EmailLocalPart: "ada", DomainID: "domain-1"})
		if !errors.Is(err, stalwart.ErrProtocol) {
			t.Fatalf("EnsureLocalUserAccount() error = %v, want ErrProtocol", err)
		}
	})

	t.Run("rejects omitted and null response account IDs", func(t *testing.T) {
		const queryRequest = `{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/query",{"filter":{"operator":"AND","conditions":[{"name":"ada"},{"domainId":"domain-1"}]},"limit":2,"sort":[{"property":"id","isAscending":true}]},"registry-account-query"]]}`
		const getRequest = `{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/get",{"ids":["account-1"],"properties":["id","@type","name","domainId","emailAddress"]},"registry-account-get"]]}`
		const setRequest = `{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/set",{"create":{"hub-local-user":{"@type":"User","name":"ada","domainId":"domain-1","credentials":{},"memberGroupIds":{},"roles":{"@type":"User"},"permissions":{"@type":"Inherit"},"quotas":{},"aliases":{},"encryptionAtRest":{"@type":"Disabled"}}}},"registry-account-set"]]}`

		for _, testCase := range []struct {
			name             string
			responses        []string
			wantErr          error
			wantPayloads     []string
			setIndeterminate bool
		}{
			{
				name: "query omitted",
				responses: []string{
					`{"methodResponses":[["x:Account/query",{"queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":[]},"registry-account-query"]]}`,
				},
				wantErr:      stalwart.ErrProtocol,
				wantPayloads: []string{queryRequest},
			},
			{
				name: "query null",
				responses: []string{
					`{"methodResponses":[["x:Account/query",{"accountId":null,"queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":[]},"registry-account-query"]]}`,
				},
				wantErr:      stalwart.ErrProtocol,
				wantPayloads: []string{queryRequest},
			},
			{
				name: "get omitted",
				responses: []string{
					`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":["account-1"]},"registry-account-query"]]}`,
					`{"methodResponses":[["x:Account/get",{"list":[{"id":"account-1","@type":"User","name":"ada","domainId":"domain-1","emailAddress":"ada@example.test"}],"notFound":[]},"registry-account-get"]]}`,
				},
				wantErr:      stalwart.ErrProtocol,
				wantPayloads: []string{queryRequest, getRequest},
			},
			{
				name: "get null",
				responses: []string{
					`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":["account-1"]},"registry-account-query"]]}`,
					`{"methodResponses":[["x:Account/get",{"accountId":null,"list":[{"id":"account-1","@type":"User","name":"ada","domainId":"domain-1","emailAddress":"ada@example.test"}],"notFound":[]},"registry-account-get"]]}`,
				},
				wantErr:      stalwart.ErrProtocol,
				wantPayloads: []string{queryRequest, getRequest},
			},
			{
				name: "set omitted",
				responses: []string{
					`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":[]},"registry-account-query"]]}`,
					`{"methodResponses":[["x:Account/set",{"created":{"hub-local-user":{"id":"account-2"}}},"registry-account-set"]]}`,
				},
				wantErr:          stalwart.ErrIndeterminate,
				wantPayloads:     []string{queryRequest, setRequest},
				setIndeterminate: true,
			},
			{
				name: "set null",
				responses: []string{
					`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":[]},"registry-account-query"]]}`,
					`{"methodResponses":[["x:Account/set",{"accountId":null,"created":{"hub-local-user":{"id":"account-2"}}},"registry-account-set"]]}`,
				},
				wantErr:          stalwart.ErrIndeterminate,
				wantPayloads:     []string{queryRequest, setRequest},
				setIndeterminate: true,
			},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				fixture := newRegistryAccountHTTPFixture(t, testCase.responses)
				_, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), stalwart.LocalUserAccountSpec{EmailLocalPart: "ada", DomainID: "domain-1"})
				if !errors.Is(err, testCase.wantErr) {
					t.Fatalf("EnsureLocalUserAccount() error = %v, want %v", err, testCase.wantErr)
				}
				if testCase.setIndeterminate && len(fixture.requests) != 2 {
					t.Fatalf("invalid set response issued %d requests, want initial query and exactly one mutation without recovery", len(fixture.requests))
				}
				fixture.requirePayloads(t, testCase.wantPayloads...)
			})
		}
	})

	t.Run("recovers the concurrent create winner after not created state mismatch", func(t *testing.T) {
		fixture := newRegistryAccountHTTPFixture(t, []string{
			`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":[]},"registry-account-query"]]}`,
			`{"methodResponses":[["x:Account/set",{"accountId":"0","notCreated":{"hub-local-user":{"type":"stateMismatch"}}},"registry-account-set"]]}`,
			`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-2","canCalculateChanges":false,"position":0,"ids":["account-3"]},"registry-account-query"]]}`,
			`{"methodResponses":[["x:Account/get",{"accountId":"0","list":[{"id":"account-3","@type":"User","name":"ada","domainId":"domain-1","emailAddress":"ada@example.test"}],"notFound":[]},"registry-account-get"]]}`,
		})

		account, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), stalwart.LocalUserAccountSpec{EmailLocalPart: "ada", DomainID: "domain-1", DomainName: "example.test"})
		if err != nil || account.AccountID != "account-3" {
			t.Fatalf("EnsureLocalUserAccount() = %#v, %v; want concurrent winner account-3", account, err)
		}
		fixture.requirePayloads(t,
			`{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/query",{"filter":{"operator":"AND","conditions":[{"name":"ada"},{"domainId":"domain-1"}]},"limit":2,"sort":[{"property":"id","isAscending":true}]},"registry-account-query"]]}`,
			`{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/set",{"create":{"hub-local-user":{"@type":"User","name":"ada","domainId":"domain-1","credentials":{},"memberGroupIds":{},"roles":{"@type":"User"},"permissions":{"@type":"Inherit"},"quotas":{},"aliases":{},"encryptionAtRest":{"@type":"Disabled"}}}},"registry-account-set"]]}`,
			`{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/query",{"filter":{"operator":"AND","conditions":[{"name":"ada"},{"domainId":"domain-1"}]},"limit":2,"sort":[{"property":"id","isAscending":true}]},"registry-account-query"]]}`,
			`{"using":["urn:ietf:params:jmap:core","urn:stalwart:jmap"],"methodCalls":[["x:Account/get",{"ids":["account-3"],"properties":["id","@type","name","domainId","emailAddress"]},"registry-account-get"]]}`,
		)
	})

	t.Run("recovers the concurrent create winner after method state mismatch", func(t *testing.T) {
		fixture := newRegistryAccountHTTPFixture(t, []string{
			`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":[]},"registry-account-query"]]}`,
			`{"methodResponses":[["error",{"type":"stateMismatch","description":"provider-secret"},"registry-account-set"]]}`,
			`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-2","canCalculateChanges":false,"position":0,"ids":["account-4"]},"registry-account-query"]]}`,
			`{"methodResponses":[["x:Account/get",{"accountId":"0","list":[{"id":"account-4","@type":"User","name":"ada","domainId":"domain-1","emailAddress":"provider@example.test"}],"notFound":[]},"registry-account-get"]]}`,
		})

		account, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), stalwart.LocalUserAccountSpec{EmailLocalPart: "ada", DomainID: "domain-1"})
		if err != nil || account.AccountID != "account-4" {
			t.Fatalf("EnsureLocalUserAccount() = %#v, %v; want concurrent winner account-4", account, err)
		}
		if len(fixture.requests) != 4 {
			t.Fatalf("JMAP request count = %d, want query, failed set, recovery query and get", len(fixture.requests))
		}
	})

	t.Run("maps provider method errors without provider detail", func(t *testing.T) {
		fixture := newRegistryAccountHTTPFixture(t, []string{
			`{"methodResponses":[["error",{"type":"forbidden","description":"provider-secret"},"registry-account-query"]]}`,
		})

		_, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), stalwart.LocalUserAccountSpec{EmailLocalPart: "ada", DomainID: "domain-1"})
		if !errors.Is(err, stalwart.ErrUnauthorized) || strings.Contains(err.Error(), "provider-secret") {
			t.Fatalf("EnsureLocalUserAccount() error = %v, want redacted ErrUnauthorized", err)
		}
	})

	t.Run("does not send a request after cancellation", func(t *testing.T) {
		fixture := newRegistryAccountHTTPFixture(t, nil)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := fixture.accountClient.EnsureLocalUserAccount(ctx, stalwart.LocalUserAccountSpec{EmailLocalPart: "ada", DomainID: "domain-1"})
		if !errors.Is(err, stalwart.ErrCancelled) {
			t.Fatalf("EnsureLocalUserAccount() error = %v, want ErrCancelled", err)
		}
		fixture.requirePayloads(t)
	})

	t.Run("rejects invalid local account input before I O", func(t *testing.T) {
		for _, spec := range []stalwart.LocalUserAccountSpec{
			{EmailLocalPart: "Ada", DomainID: "domain-1"},
			{EmailLocalPart: "ada", DomainID: "domain\n1"},
			{EmailLocalPart: strings.Repeat("a", 65), DomainID: "domain-1"},
			{EmailLocalPart: "ada", DomainID: "domain-1", DomainName: "Example.test"},
		} {
			fixture := newRegistryAccountHTTPFixture(t, nil)
			_, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), spec)
			if !errors.Is(err, stalwart.ErrInvalidRequest) {
				t.Fatalf("EnsureLocalUserAccount(%#v) error = %v, want ErrInvalidRequest", spec, err)
			}
			fixture.requirePayloads(t)
		}
	})
}

type registryAccountHTTPFixture struct {
	accountClient *stalwart.RegistryAccountClient
	responses     []string
	requests      []string
	registryCalls int
	mailboxCalls  int
}

func newRegistryAccountHTTPFixture(t *testing.T, responses []string) *registryAccountHTTPFixture {
	t.Helper()
	fixture := &registryAccountHTTPFixture{responses: append([]string(nil), responses...)}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/jmap" || request.URL.RawQuery != "" || request.URL.Fragment != "" {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		fixture.requests = append(fixture.requests, string(body))
		if len(fixture.responses) < len(fixture.requests) {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(fixture.responses[len(fixture.requests)-1]))
	}))
	t.Cleanup(server.Close)

	client, err := stalwart.NewClient(stalwart.Config{
		Origin:             server.URL,
		HTTPClient:         server.Client(),
		RegistryAuthorizer: registryAccountTestAuthorizer{calls: &fixture.registryCalls},
		MailboxAuthorizer:  registryAccountTestAuthorizer{calls: &fixture.mailboxCalls},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	fixture.accountClient, err = stalwart.NewRegistryAccountClient(client)
	if err != nil {
		t.Fatalf("NewRegistryAccountClient() error = %v", err)
	}
	return fixture
}

func (fixture *registryAccountHTTPFixture) requirePayloads(t *testing.T, want ...string) {
	t.Helper()
	if len(fixture.requests) != len(want) {
		t.Fatalf("JMAP request count = %d, want %d; requests = %#v", len(fixture.requests), len(want), fixture.requests)
	}
	if fixture.registryCalls != len(want) || fixture.mailboxCalls != 0 {
		t.Fatalf("authorizer calls = registry %d, mailbox %d; want registry %d, mailbox 0", fixture.registryCalls, fixture.mailboxCalls, len(want))
	}
	for index := range want {
		var gotValue any
		if err := json.Unmarshal([]byte(fixture.requests[index]), &gotValue); err != nil {
			t.Fatalf("request %d invalid JSON: %v", index, err)
		}
		var wantValue any
		if err := json.Unmarshal([]byte(want[index]), &wantValue); err != nil {
			t.Fatalf("test expected request %d invalid JSON: %v", index, err)
		}
		if !reflect.DeepEqual(gotValue, wantValue) {
			t.Fatalf("JMAP request %d = %s, want exact payload %s", index, fixture.requests[index], want[index])
		}
	}
}

type registryAccountTestAuthorizer struct {
	calls *int
}

func (authorizer registryAccountTestAuthorizer) AuthorizeRegistry(context.Context, *http.Request) error {
	if authorizer.calls != nil {
		(*authorizer.calls)++
	}
	return nil
}

func (authorizer registryAccountTestAuthorizer) AuthorizeMailbox(context.Context, *http.Request) error {
	if authorizer.calls != nil {
		(*authorizer.calls)++
	}
	return nil
}
