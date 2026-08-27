package mail_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

type stalwartProviderErrorAuthorizer struct{}

func (stalwartProviderErrorAuthorizer) AuthorizeRegistry(context.Context, *http.Request) error {
	return nil
}

func (stalwartProviderErrorAuthorizer) AuthorizeMailbox(context.Context, *http.Request) error {
	return nil
}

type stalwartProviderErrorRegistryAuthorizer struct {
	err error
}

func (authorizer stalwartProviderErrorRegistryAuthorizer) AuthorizeRegistry(context.Context, *http.Request) error {
	return authorizer.err
}

func TestStalwartClient_CallRegistryAuthorizationFailureMappingSkipsHTTP(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		authorizerErr error
		wantKind      stalwart.ErrorKind
		wantSentinel  error
	}{
		{name: `cancelled sentinel`, authorizerErr: stalwart.ErrCancelled, wantKind: stalwart.ErrorCancelled, wantSentinel: stalwart.ErrCancelled},
		{name: `context cancellation`, authorizerErr: context.Canceled, wantKind: stalwart.ErrorCancelled, wantSentinel: stalwart.ErrCancelled},
		{name: `unauthorized`, authorizerErr: stalwart.ErrUnauthorized, wantKind: stalwart.ErrorUnauthorized, wantSentinel: stalwart.ErrUnauthorized},
		{name: `unavailable`, authorizerErr: stalwart.ErrUnavailable, wantKind: stalwart.ErrorUnavailable, wantSentinel: stalwart.ErrUnavailable},
		{name: `deadline`, authorizerErr: context.DeadlineExceeded, wantKind: stalwart.ErrorUnavailable, wantSentinel: stalwart.ErrUnavailable},
		{name: `unknown`, authorizerErr: errors.New(`provider-auth-secret`), wantKind: stalwart.ErrorUnavailable, wantSentinel: stalwart.ErrUnavailable},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				calls++
			}))
			defer server.Close()

			client, err := stalwart.NewClient(stalwart.Config{
				Origin:             server.URL,
				HTTPClient:         server.Client(),
				RegistryAuthorizer: stalwartProviderErrorRegistryAuthorizer{err: testCase.authorizerErr},
				MailboxAuthorizer:  stalwartProviderErrorAuthorizer{},
			})
			if err != nil {
				t.Fatalf(`NewClient() error = %v`, err)
			}

			_, err = client.CallRegistry(context.Background(), stalwart.JMAPRequest{
				Using: []string{`urn:ietf:params:jmap:core`, `urn:stalwart:jmap`},
				MethodCalls: []stalwart.JMAPMethodCall{{
					Name:      `x:Account/get`,
					Arguments: json.RawMessage(`{}`),
					CallID:    `auth-map-1`,
				}},
			})
			kind, known := stalwart.KindOf(err)
			if !known || kind != testCase.wantKind || !errors.Is(err, testCase.wantSentinel) {
				t.Fatalf(`CallRegistry() error = %v, kind = %q, known = %t; want %q`, err, kind, known, testCase.wantKind)
			}
			if calls != 0 {
				t.Fatalf(`CallRegistry() made %d HTTP requests; want 0`, calls)
			}
			if strings.Contains(err.Error(), `provider-auth-secret`) || strings.Contains(err.Error(), server.URL) {
				t.Fatalf(`CallRegistry() exposed authorization detail in %q`, err)
			}
		})
	}
}

func TestStalwartClient_RateLimitKeepsReadsRetryableAndMakesMutationsIndeterminate(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		method       string
		status       int
		wantKind     stalwart.ErrorKind
		wantSentinel error
	}{
		{name: "method read", method: "x:Account/get", wantKind: stalwart.ErrorRateLimited, wantSentinel: stalwart.ErrRateLimited},
		{name: "method mutation", method: "x:Account/set", wantKind: stalwart.ErrorIndeterminate, wantSentinel: stalwart.ErrIndeterminate},
		{name: "http read", method: "x:Account/get", status: http.StatusTooManyRequests, wantKind: stalwart.ErrorRateLimited, wantSentinel: stalwart.ErrRateLimited},
		{name: "http mutation", method: "x:Account/set", status: http.StatusTooManyRequests, wantKind: stalwart.ErrorIndeterminate, wantSentinel: stalwart.ErrIndeterminate},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.Method != http.MethodPost {
					writer.WriteHeader(http.StatusBadRequest)
					return
				}
				if request.URL.Path == "/api" {
					writer.WriteHeader(http.StatusForbidden)
					return
				}
				if request.URL.Path != "/jmap" {
					writer.WriteHeader(http.StatusBadRequest)
					return
				}
				if testCase.status != 0 {
					writer.WriteHeader(testCase.status)
					return
				}
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(`{"methodResponses":[["error",{"type":"rateLimit"},"call-1"]]}`))
			}))
			defer server.Close()

			client, err := stalwart.NewClient(stalwart.Config{
				Origin:             server.URL,
				HTTPClient:         server.Client(),
				RegistryAuthorizer: stalwartProviderErrorAuthorizer{},
				MailboxAuthorizer:  stalwartProviderErrorAuthorizer{},
			})
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			_, err = client.CallRegistry(context.Background(), stalwart.JMAPRequest{
				Using: []string{"urn:ietf:params:jmap:core", "urn:stalwart:jmap"},
				MethodCalls: []stalwart.JMAPMethodCall{{
					Name:      testCase.method,
					Arguments: json.RawMessage("{}"),
					CallID:    "call-1",
				}},
			})
			kind, known := stalwart.KindOf(err)
			if !known || kind != testCase.wantKind || !errors.Is(err, testCase.wantSentinel) {
				t.Fatalf("CallRegistry(%q) error = %v, kind = %q, known = %t; want %q", testCase.method, err, kind, known, testCase.wantKind)
			}
			if strings.Contains(err.Error(), "rateLimit") || strings.Contains(err.Error(), server.URL) {
				t.Fatalf("CallRegistry(%q) exposed provider detail in %q", testCase.method, err)
			}
		})
	}
}

func TestStalwartClient_DiscoveryRateLimitIsTypedAtBothGETStages(t *testing.T) {
	const providerDetail = "provider-credential=discovery-secret"

	for _, testCase := range []struct {
		name       string
		firstStage bool
		wantCalls  int
	}{
		{name: "well-known", firstStage: true, wantCalls: 1},
		{name: "session", wantCalls: 2},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				calls++
				switch request.URL.Path {
				case "/.well-known/jmap":
					if testCase.firstStage {
						writer.WriteHeader(http.StatusTooManyRequests)
						_, _ = writer.Write([]byte(providerDetail))
						return
					}
					writer.Header().Set("Location", "/jmap/session")
					writer.WriteHeader(http.StatusTemporaryRedirect)
				case "/jmap/session":
					writer.WriteHeader(http.StatusTooManyRequests)
					_, _ = writer.Write([]byte(providerDetail))
				default:
					writer.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer server.Close()

			client, err := stalwart.NewClient(stalwart.Config{
				Origin:             server.URL,
				HTTPClient:         server.Client(),
				RegistryAuthorizer: stalwartProviderErrorAuthorizer{},
				MailboxAuthorizer:  stalwartProviderErrorAuthorizer{},
			})
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			_, err = client.DiscoverMailboxSession(context.Background())
			kind, known := stalwart.KindOf(err)
			if !known || kind != stalwart.ErrorRateLimited || !errors.Is(err, stalwart.ErrRateLimited) {
				t.Fatalf("DiscoverMailboxSession() error = %v, kind = %q, known = %t; want rate-limited", err, kind, known)
			}
			if calls != testCase.wantCalls || strings.Contains(err.Error(), providerDetail) || strings.Contains(err.Error(), server.URL) {
				t.Fatalf("DiscoverMailboxSession() calls=%d error=%q, want %d calls and no provider detail", calls, err, testCase.wantCalls)
			}
		})
	}
}

type stalwartProviderErrorMailboxClient struct {
	calls int
}

func (client *stalwartProviderErrorMailboxClient) AccountID() string {
	return "account-a"
}

func (client *stalwartProviderErrorMailboxClient) CallMailbox(_ context.Context, _ stalwart.JMAPRequest) (stalwart.JMAPResponse, error) {
	client.calls++
	if client.calls == 1 {
		return stalwart.JMAPResponse{MethodResponses: []stalwart.JMAPMethodResponse{{
			Name: "Mailbox/get", CallID: "mailbox-get", Arguments: json.RawMessage(`{"accountId":"account-a","state":"state-a","list":[],"notFound":[]} `),
		}}}, nil
	}
	return stalwart.JMAPResponse{}, stalwart.ErrRateLimited
}

func (*stalwartProviderErrorMailboxClient) UploadBlob(context.Context, mail.StalwartMailboxUploadRequest) (mail.StalwartMailboxUploadResult, error) {
	return mail.StalwartMailboxUploadResult{}, stalwart.ErrUnavailable
}

func (*stalwartProviderErrorMailboxClient) DownloadBlob(context.Context, string, int64) (mail.StalwartMailboxBlob, error) {
	return mail.StalwartMailboxBlob{}, stalwart.ErrUnavailable
}

type stalwartProviderErrorMailboxFactory struct {
	client mail.StalwartMailboxClient
}

func (factory stalwartProviderErrorMailboxFactory) OpenMailboxClient(context.Context, string) (mail.StalwartMailboxClient, error) {
	return factory.client, nil
}

func TestStalwartEngine_RateLimitKeepsReadsTypedAndMutationsIndeterminate(t *testing.T) {
	t.Run("read", func(t *testing.T) {
		engine, err := mail.NewStalwartEngine(stalwartProviderErrorMailboxFactory{client: &stalwartProviderErrorMailboxClient{calls: 1}})
		if err != nil {
			t.Fatalf("NewStalwartEngine() error = %v", err)
		}
		_, err = engine.ListFolders(context.Background(), "account-a")
		if !errors.Is(err, mail.ErrEngineRateLimited) || errors.Is(err, mail.ErrEngineIndeterminate) || strings.Contains(err.Error(), "stalwart") {
			t.Fatalf("ListFolders() error = %v, want content-free rate-limited error", err)
		}
	})

	t.Run("mutation", func(t *testing.T) {
		engine, err := mail.NewStalwartEngine(stalwartProviderErrorMailboxFactory{client: &stalwartProviderErrorMailboxClient{}})
		if err != nil {
			t.Fatalf("NewStalwartEngine() error = %v", err)
		}
		_, err = engine.CreateFolder(context.Background(), "account-a", mail.FolderSpec{Name: "Rate limit test"})
		if !errors.Is(err, mail.ErrEngineIndeterminate) || errors.Is(err, mail.ErrEngineRateLimited) || strings.Contains(err.Error(), "stalwart") {
			t.Fatalf("CreateFolder() error = %v, want content-free indeterminate error", err)
		}
	})

	t.Run("submission reconciliation", func(t *testing.T) {
		engine, err := mail.NewStalwartEngine(stalwartProviderErrorMailboxFactory{client: &stalwartProviderErrorMailboxClient{calls: 1}})
		if err != nil {
			t.Fatalf("NewStalwartEngine() error = %v", err)
		}
		result, err := engine.ReconcileSubmission(context.Background(), mail.EngineSubmission{
			AccountID: "account-a", DraftKey: "draft-key", MessageID: "message-a",
			ExpectedVersion: "version-a", SubmissionKey: "submission-key", MaxMessageBytes: 1024,
		})
		if result != nil || !errors.Is(err, mail.ErrEngineRateLimited) || errors.Is(err, mail.ErrEngineIndeterminate) || strings.Contains(err.Error(), "stalwart") {
			t.Fatalf("ReconcileSubmission() = %#v, %v; want content-free rate-limited error", result, err)
		}
	})
}

func TestStalwartSubmissionFailureTaxonomy_SourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/stalwart_engine_drafts.go")
	if err != nil {
		t.Fatalf("read draft source: %v", err)
	}
	source := strings.ReplaceAll(string(sourceBytes), "\r\n", "\n")
	start := strings.Index(source, "func stalwartSubmissionSetFailuresFor")
	end := strings.Index(source, "func stalwartSubmissionMarkerUpdateSucceeded")
	if start == -1 || end == -1 || end <= start {
		t.Fatal("submission failure mapping source markers are missing")
	}
	mapping := source[start:end]

	for _, expected := range []string{
		`case "tooManyRecipients":
		return stalwartEngineError(EngineErrorRecipientLimit)`,
		`case "forbiddenMailFrom", "forbiddenFrom", "forbiddenToSend":
		return stalwartEngineError(EngineErrorIdentityUnavailable)`,
		`case "overQuota", "invalidEmail", "noRecipients", "invalidRecipients":
		return stalwartEngineError(EngineErrorRejectedPolicy)`,
	} {
		if !strings.Contains(mapping, expected) {
			t.Fatalf("submission failure mapping is missing %q", expected)
		}
	}
}
