package mail_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const stalwartRegistryRuntimeFactoryOrigin = "https://stalwart.runtime.test"

type stalwartRegistryRuntimeFactoryAuthorizer struct {
	calls int
}

func (authorizer *stalwartRegistryRuntimeFactoryAuthorizer) AuthorizeRegistry(_ context.Context, request *http.Request) error {
	authorizer.calls++
	request.Header.Set("Authorization", "Bearer registry-scope")
	return nil
}

type stalwartRegistryRuntimeFactoryAllowRegistryAuthorizer struct{}

func (stalwartRegistryRuntimeFactoryAllowRegistryAuthorizer) AuthorizeRegistry(context.Context, *http.Request) error {
	return nil
}

type stalwartRegistryRuntimeFactoryAllowMailboxAuthorizer struct{}

func (stalwartRegistryRuntimeFactoryAllowMailboxAuthorizer) AuthorizeMailbox(context.Context, *http.Request) error {
	return nil
}

type stalwartRegistryRuntimeFactoryRequest struct {
	method        string
	path          string
	authorization string
}

type stalwartRegistryRuntimeFactoryTransport struct {
	calls    int
	requests []stalwartRegistryRuntimeFactoryRequest
}

func (transport *stalwartRegistryRuntimeFactoryTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	transport.calls++
	transport.requests = append(transport.requests, stalwartRegistryRuntimeFactoryRequest{
		method:        request.Method,
		path:          request.URL.Path,
		authorization: request.Header.Get("Authorization"),
	})

	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/.well-known/jmap":
		return stalwartRegistryRuntimeFactoryResponse(request, http.StatusTemporaryRedirect, http.Header{"Location": []string{"/jmap/session"}}, ""), nil
	case request.Method == http.MethodGet && request.URL.Path == "/jmap/session":
		return stalwartRegistryRuntimeFactoryResponse(request, http.StatusOK, http.Header{"Content-Type": []string{"application/json"}}, stalwartRegistryRuntimeFactorySessionJSON()), nil
	case request.Method == http.MethodPost && request.URL.Path == "/jmap":
		return stalwartRegistryRuntimeFactoryResponse(request, http.StatusOK, http.Header{"Content-Type": []string{"application/json"}}, `{"methodResponses":[["x:Account/get",{},"registry-get"]]}`), nil
	default:
		return stalwartRegistryRuntimeFactoryResponse(request, http.StatusBadRequest, nil, ""), nil
	}
}

func (transport *stalwartRegistryRuntimeFactoryTransport) reset() {
	transport.calls = 0
	transport.requests = nil
}

func stalwartRegistryRuntimeFactoryResponse(request *http.Request, status int, header http.Header, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    request,
	}
}

func stalwartRegistryRuntimeFactorySessionJSON() string {
	return `{"capabilities":{"urn:ietf:params:jmap:core":{},"urn:ietf:params:jmap:mail":{},"urn:ietf:params:jmap:submission":{}},"accounts":{"account-a":{}},"apiUrl":"` + stalwartRegistryRuntimeFactoryOrigin + `/jmap/","uploadUrl":"` + stalwartRegistryRuntimeFactoryOrigin + `/jmap/upload/{accountId}/","downloadUrl":"` + stalwartRegistryRuntimeFactoryOrigin + `/jmap/download/{accountId}/{blobId}/{name}?type={type}"}`
}

func TestNewStalwartRegistryClient_RejectsNilDependencies(t *testing.T) {
	validHTTPClient := &http.Client{}
	validAuthorizer := &stalwartRegistryRuntimeFactoryAuthorizer{}
	var typedNilAuthorizer *stalwartRegistryRuntimeFactoryAuthorizer

	for _, testCase := range []struct {
		name               string
		origin             string
		httpClient         *http.Client
		registryAuthorizer stalwart.RegistryAuthorizer
	}{
		{name: "nil HTTP client", origin: stalwartRegistryRuntimeFactoryOrigin, registryAuthorizer: validAuthorizer},
		{name: "nil registry authorizer", origin: stalwartRegistryRuntimeFactoryOrigin, httpClient: validHTTPClient},
		{name: "typed nil registry authorizer", origin: stalwartRegistryRuntimeFactoryOrigin, httpClient: validHTTPClient, registryAuthorizer: typedNilAuthorizer},
		{name: "origin with path", origin: stalwartRegistryRuntimeFactoryOrigin + "/forbidden", httpClient: validHTTPClient, registryAuthorizer: validAuthorizer},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			client, err := mail.NewStalwartRegistryClient(testCase.origin, testCase.httpClient, testCase.registryAuthorizer)
			if client != nil || !errors.Is(err, stalwart.ErrInvalidConfiguration) {
				t.Fatalf("NewStalwartRegistryClient() = (%v, %v); want nil, ErrInvalidConfiguration", client, err)
			}
		})
	}
}

func TestNewStalwartRegistryClient_UsesRegistryAuthorizerForRegistryCall(t *testing.T) {
	transport := &stalwartRegistryRuntimeFactoryTransport{}
	authorizer := &stalwartRegistryRuntimeFactoryAuthorizer{}
	client, err := mail.NewStalwartRegistryClient(stalwartRegistryRuntimeFactoryOrigin, &http.Client{Transport: transport}, authorizer)
	if err != nil {
		t.Fatalf("NewStalwartRegistryClient() error = %v", err)
	}
	if transport.calls != 0 || authorizer.calls != 0 {
		t.Fatalf("constructor transport/authorizer calls = %d/%d; want 0/0", transport.calls, authorizer.calls)
	}

	_, err = client.CallRegistry(context.Background(), stalwart.JMAPRequest{
		Using: []string{"urn:ietf:params:jmap:core", "urn:stalwart:jmap"},
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "x:Account/get",
			Arguments: json.RawMessage(`{}`),
			CallID:    "registry-get",
		}},
	})
	if err != nil {
		t.Fatalf("CallRegistry() error = %v", err)
	}
	if authorizer.calls != 1 || transport.calls != 1 || len(transport.requests) != 1 {
		t.Fatalf("registry authorizer/transport/requests = %d/%d/%d; want 1/1/1", authorizer.calls, transport.calls, len(transport.requests))
	}
	request := transport.requests[0]
	if request.method != http.MethodPost || request.path != "/jmap" || request.authorization != "Bearer registry-scope" {
		t.Fatalf("registry request = %#v; want POST /jmap with scoped registry authorization", request)
	}
}

func TestNewStalwartRegistryClient_RejectsMailboxCallBeforeTransport(t *testing.T) {
	transport := &stalwartRegistryRuntimeFactoryTransport{}
	httpClient := &http.Client{Transport: transport}
	sessionSource, err := stalwart.NewClient(stalwart.Config{
		Origin:             stalwartRegistryRuntimeFactoryOrigin,
		HTTPClient:         httpClient,
		RegistryAuthorizer: stalwartRegistryRuntimeFactoryAllowRegistryAuthorizer{},
		MailboxAuthorizer:  stalwartRegistryRuntimeFactoryAllowMailboxAuthorizer{},
	})
	if err != nil {
		t.Fatalf("session source NewClient() error = %v", err)
	}
	session, err := sessionSource.DiscoverMailboxSession(context.Background())
	if err != nil {
		t.Fatalf("DiscoverMailboxSession() error = %v", err)
	}
	transport.reset()

	authorizer := &stalwartRegistryRuntimeFactoryAuthorizer{}
	client, err := mail.NewStalwartRegistryClient(stalwartRegistryRuntimeFactoryOrigin, httpClient, authorizer)
	if err != nil {
		t.Fatalf("NewStalwartRegistryClient() error = %v", err)
	}
	if transport.calls != 0 || authorizer.calls != 0 {
		t.Fatalf("constructor transport/registry-authorizer calls = %d/%d; want 0/0", transport.calls, authorizer.calls)
	}

	_, err = client.CallMailbox(context.Background(), session, stalwart.JMAPRequest{
		Using: []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail", "urn:ietf:params:jmap:submission"},
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Mailbox/get",
			Arguments: json.RawMessage(`{"accountId":"account-a"}`),
			CallID:    "mailbox-get",
		}},
	})
	if !errors.Is(err, stalwart.ErrUnauthorized) {
		t.Fatalf("CallMailbox() error = %v; want ErrUnauthorized", err)
	}
	if transport.calls != 0 || len(transport.requests) != 0 || authorizer.calls != 0 {
		t.Fatalf("mailbox transport/requests/registry-authorizer calls = %d/%d/%d; want 0/0/0", transport.calls, len(transport.requests), authorizer.calls)
	}
}
