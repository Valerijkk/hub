package mail_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

func TestStalwartRegistryAPIKeyAuthorizer_AuthorizeRegistry(t *testing.T) {
	t.Run("authorizes only the exact canonical registry request", func(t *testing.T) {
		lookup := &registryAPIKeyLookupRecorder{key: []byte("scoped-key_123")}
		authorizer, err := mail.NewStalwartRegistryAPIKeyAuthorizer(" HTTPS://Registry.Example.Test/ ", lookup.lookup)
		if err != nil {
			t.Fatalf("NewStalwartRegistryAPIKeyAuthorizer() error = %v", err)
		}
		request := newRegistryAPIKeyRequest(http.MethodPost, "https", "registry.example.test", "/jmap")

		if err := authorizer.AuthorizeRegistry(context.Background(), request); err != nil {
			t.Fatalf("AuthorizeRegistry() error = %v", err)
		}
		if lookup.calls != 1 {
			t.Fatalf("lookup calls = %d; want 1", lookup.calls)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer scoped-key_123" {
			t.Fatalf("Authorization = %q; want bearer API key", got)
		}
	})

	for _, testCase := range []struct {
		name    string
		ctx     context.Context
		request *http.Request
		wantErr error
	}{
		{
			name: "nil context", request: newRegistryAPIKeyRequest(http.MethodPost, "https", "registry.example.test", "/jmap"), wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "cancelled context", ctx: cancelledRegistryAPIKeyContext(), request: newRegistryAPIKeyRequest(http.MethodPost, "https", "registry.example.test", "/jmap"), wantErr: stalwart.ErrCancelled,
		},
		{
			name: "wrong method", ctx: context.Background(), request: newRegistryAPIKeyRequest(http.MethodGet, "https", "registry.example.test", "/jmap"), wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "wrong scheme", ctx: context.Background(), request: newRegistryAPIKeyRequest(http.MethodPost, "http", "registry.example.test", "/jmap"), wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "wrong host", ctx: context.Background(), request: newRegistryAPIKeyRequest(http.MethodPost, "https", "other.example.test", "/jmap"), wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "wrong path", ctx: context.Background(), request: newRegistryAPIKeyRequest(http.MethodPost, "https", "registry.example.test", "/api"), wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "encoded path", ctx: context.Background(), request: &http.Request{Method: http.MethodPost, URL: &url.URL{Scheme: "https", Host: "registry.example.test", Path: "/jmap", RawPath: "/%6amap"}, Header: make(http.Header)}, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "query", ctx: context.Background(), request: &http.Request{Method: http.MethodPost, URL: &url.URL{Scheme: "https", Host: "registry.example.test", Path: "/jmap", RawQuery: "next=1"}, Header: make(http.Header)}, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "fragment", ctx: context.Background(), request: &http.Request{Method: http.MethodPost, URL: &url.URL{Scheme: "https", Host: "registry.example.test", Path: "/jmap", Fragment: "fragment"}, Header: make(http.Header)}, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "userinfo", ctx: context.Background(), request: &http.Request{Method: http.MethodPost, URL: &url.URL{Scheme: "https", Host: "registry.example.test", Path: "/jmap", User: url.User("user")}, Header: make(http.Header)}, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "preexisting authorization", ctx: context.Background(), request: registryAPIKeyRequestWithHeader("authorization", "Bearer previous"), wantErr: stalwart.ErrUnauthorized,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			lookup := &registryAPIKeyLookupRecorder{key: []byte("scoped-key_123")}
			authorizer, err := mail.NewStalwartRegistryAPIKeyAuthorizer("https://registry.example.test", lookup.lookup)
			if err != nil {
				t.Fatalf("NewStalwartRegistryAPIKeyAuthorizer() error = %v", err)
			}
			before := registryAPIKeyAuthorizationValue(testCase.request)

			err = authorizer.AuthorizeRegistry(testCase.ctx, testCase.request)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("AuthorizeRegistry() error = %v; want %v", err, testCase.wantErr)
			}
			if lookup.calls != 0 {
				t.Fatalf("lookup calls = %d; want request rejected before lookup", lookup.calls)
			}
			if got := registryAPIKeyAuthorizationValue(testCase.request); got != before {
				t.Fatalf("Authorization = %q; want unchanged %q", got, before)
			}
		})
	}

	for _, testCase := range []struct {
		name       string
		lookup     func(context.Context) ([]byte, error)
		wantErr    error
		wantDetail string
	}{
		{
			name: "lookup error", lookup: func(context.Context) ([]byte, error) { return nil, errors.New("registry secret must not leak") }, wantErr: stalwart.ErrUnavailable, wantDetail: "registry secret",
		},
		{
			name: "lookup panic", lookup: func(context.Context) ([]byte, error) { panic("registry secret must not leak") }, wantErr: stalwart.ErrUnavailable, wantDetail: "registry secret",
		},
		{
			name: "empty key", lookup: func(context.Context) ([]byte, error) { return nil, nil }, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "oversize key", lookup: func(context.Context) ([]byte, error) { return []byte(strings.Repeat("a", 4097)), nil }, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "space in key", lookup: func(context.Context) ([]byte, error) { return []byte("key value"), nil }, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "control in key", lookup: func(context.Context) ([]byte, error) { return []byte("key\nvalue"), nil }, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "non ascii key", lookup: func(context.Context) ([]byte, error) { return []byte("key\xc3\xb1"), nil }, wantErr: stalwart.ErrUnauthorized,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			calls := 0
			authorizer, err := mail.NewStalwartRegistryAPIKeyAuthorizer("https://registry.example.test", func(ctx context.Context) ([]byte, error) {
				calls++
				return testCase.lookup(ctx)
			})
			if err != nil {
				t.Fatalf("NewStalwartRegistryAPIKeyAuthorizer() error = %v", err)
			}
			request := newRegistryAPIKeyRequest(http.MethodPost, "https", "registry.example.test", "/jmap")

			err = authorizer.AuthorizeRegistry(context.Background(), request)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("AuthorizeRegistry() error = %v; want %v", err, testCase.wantErr)
			}
			if calls != 1 {
				t.Fatalf("lookup calls = %d; want 1", calls)
			}
			if got := request.Header.Get("Authorization"); got != "" {
				t.Fatalf("Authorization = %q; want no mutation", got)
			}
			if testCase.wantDetail != "" && strings.Contains(err.Error(), testCase.wantDetail) {
				t.Fatalf("AuthorizeRegistry() exposed lookup detail in %q", err)
			}
		})
	}
}

func TestStalwartRegistryAPIKeyAuthorizer_ContextCancellationAfterLookupLeavesRequestUntouched(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	authorizer, err := mail.NewStalwartRegistryAPIKeyAuthorizer("https://registry.example.test", func(context.Context) ([]byte, error) {
		cancel()
		return []byte("scoped-key_123"), nil
	})
	if err != nil {
		t.Fatalf("NewStalwartRegistryAPIKeyAuthorizer() error = %v", err)
	}
	request := newRegistryAPIKeyRequest(http.MethodPost, "https", "registry.example.test", "/jmap")

	if err := authorizer.AuthorizeRegistry(ctx, request); !errors.Is(err, stalwart.ErrCancelled) {
		t.Fatalf("AuthorizeRegistry() error = %v; want ErrCancelled", err)
	}
	if got := request.Header.Get("Authorization"); got != "" {
		t.Fatalf("Authorization = %q; want no mutation after cancellation", got)
	}
}

func TestStalwartRegistryAPIKeyAuthorizer_LeavesLookupBufferCallerOwned(t *testing.T) {
	key := []byte("scoped-key_123")
	before := append([]byte(nil), key...)
	authorizer, err := mail.NewStalwartRegistryAPIKeyAuthorizer("https://registry.example.test", func(context.Context) ([]byte, error) { return key, nil })
	if err != nil {
		t.Fatalf("NewStalwartRegistryAPIKeyAuthorizer() error = %v", err)
	}

	if err := authorizer.AuthorizeRegistry(context.Background(), newRegistryAPIKeyRequest(http.MethodPost, "https", "registry.example.test", "/jmap")); err != nil {
		t.Fatalf("AuthorizeRegistry() error = %v", err)
	}
	if string(key) != string(before) {
		t.Fatalf("lookup key = %q; want caller-owned buffer unchanged", key)
	}
}

func TestNewStalwartRegistryAPIKeyAuthorizer_RejectsInvalidConfigurationWithoutLookup(t *testing.T) {
	for _, origin := range []string{"", "registry.example.test", "ftp://registry.example.test", "https://user@registry.example.test", "https://registry.example.test/api", "https://registry.example.test?next=1", "https://registry.example.test#fragment"} {
		t.Run(origin, func(t *testing.T) {
			calls := 0
			authorizer, err := mail.NewStalwartRegistryAPIKeyAuthorizer(origin, func(context.Context) ([]byte, error) {
				calls++
				return []byte("scoped-key_123"), nil
			})
			if authorizer != nil || !errors.Is(err, stalwart.ErrInvalidConfiguration) {
				t.Fatalf("NewStalwartRegistryAPIKeyAuthorizer(%q) = (%v, %v); want nil, ErrInvalidConfiguration", origin, authorizer, err)
			}
			if calls != 0 {
				t.Fatalf("lookup calls = %d; want no constructor lookup", calls)
			}
		})
	}

	authorizer, err := mail.NewStalwartRegistryAPIKeyAuthorizer("https://registry.example.test", nil)
	if authorizer != nil || !errors.Is(err, stalwart.ErrInvalidConfiguration) {
		t.Fatalf("NewStalwartRegistryAPIKeyAuthorizer(nil lookup) = (%v, %v); want nil, ErrInvalidConfiguration", authorizer, err)
	}
}

func TestStalwartMailboxRegistryAuthorizer_RemainsFailClosed(t *testing.T) {
	request := newRegistryAPIKeyRequest(http.MethodPost, "https", "registry.example.test", "/jmap")
	if err := mail.NewStalwartMailboxRegistryAuthorizer().AuthorizeRegistry(context.Background(), request); !errors.Is(err, stalwart.ErrUnauthorized) {
		t.Fatalf("AuthorizeRegistry() error = %v; want ErrUnauthorized", err)
	}
	if got := request.Header.Get("Authorization"); got != "" {
		t.Fatalf("Authorization = %q; want no mailbox registry credential", got)
	}
}

type registryAPIKeyLookupRecorder struct {
	key   []byte
	calls int
}

func (recorder *registryAPIKeyLookupRecorder) lookup(context.Context) ([]byte, error) {
	recorder.calls++
	return recorder.key, nil
}

func newRegistryAPIKeyRequest(method, scheme, host, path string) *http.Request {
	return &http.Request{Method: method, URL: &url.URL{Scheme: scheme, Host: host, Path: path}, Header: make(http.Header)}
}

func registryAPIKeyRequestWithHeader(name, value string) *http.Request {
	request := newRegistryAPIKeyRequest(http.MethodPost, "https", "registry.example.test", "/jmap")
	request.Header[name] = []string{value}
	return request
}

func registryAPIKeyAuthorizationValue(request *http.Request) string {
	if request == nil || request.Header == nil {
		return ""
	}
	return request.Header.Get("Authorization")
}

func cancelledRegistryAPIKeyContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
