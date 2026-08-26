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
			`{"methodResponses":[["x:Account/set",{"created":{"hub-local-user":{"id":"account-2"}}},"registry-account-set"]]}`,
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

	t.Run("rejects not created mutation result", func(t *testing.T) {
		fixture := newRegistryAccountHTTPFixture(t, []string{
			`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":[]},"registry-account-query"]]}`,
			`{"methodResponses":[["x:Account/set",{"notCreated":{"hub-local-user":{"type":"stateMismatch"}}},"registry-account-set"]]}`,
		})

		_, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), stalwart.LocalUserAccountSpec{EmailLocalPart: "ada", DomainID: "domain-1"})
		if !errors.Is(err, stalwart.ErrIndeterminate) {
			t.Fatalf("EnsureLocalUserAccount() error = %v, want redacted ErrIndeterminate", err)
		}
		if len(fixture.requests) != 2 {
			t.Fatalf("JMAP request count = %d, want query plus one failed set", len(fixture.requests))
		}
	})

	t.Run("maps state mismatch method error", func(t *testing.T) {
		fixture := newRegistryAccountHTTPFixture(t, []string{
			`{"methodResponses":[["x:Account/query",{"accountId":"0","queryState":"account-state-1","canCalculateChanges":false,"position":0,"ids":[]},"registry-account-query"]]}`,
			`{"methodResponses":[["error",{"type":"stateMismatch","description":"provider-secret"},"registry-account-set"]]}`,
		})

		_, err := fixture.accountClient.EnsureLocalUserAccount(context.Background(), stalwart.LocalUserAccountSpec{EmailLocalPart: "ada", DomainID: "domain-1"})
		if !errors.Is(err, stalwart.ErrStateMismatch) || strings.Contains(err.Error(), "provider-secret") {
			t.Fatalf("EnsureLocalUserAccount() error = %v, want redacted ErrStateMismatch", err)
		}
		if len(fixture.requests) != 2 {
			t.Fatalf("JMAP request count = %d, want query plus one state-mismatched set", len(fixture.requests))
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
		RegistryAuthorizer: registryAccountTestAuthorizer{},
		MailboxAuthorizer:  registryAccountTestAuthorizer{},
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

type registryAccountTestAuthorizer struct{}

func (registryAccountTestAuthorizer) AuthorizeRegistry(context.Context, *http.Request) error { return nil }

func (registryAccountTestAuthorizer) AuthorizeMailbox(context.Context, *http.Request) error { return nil }
