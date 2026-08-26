package mail_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

func TestRegistryAuthorizer_AuthorizeRegistry(t *testing.T) {
	for _, testCase := range []struct {
		name               string
		ctx                context.Context
		provider           stalwart.RegistryManagementAPIKeyProvider
		request            *http.Request
		wantErr            error
		wantAuthorization  string
		wantProviderCalled bool
	}{
		{
			name: "exact registry request", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("scoped-key_123"), nil },
			request: newRegistryRequest(), wantAuthorization: "Bearer scoped-key_123", wantProviderCalled: true,
		},
		{
			name: "nil context", provider: func(context.Context) ([]byte, error) { return []byte("key"), nil },
			request: newRegistryRequest(), wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "cancelled context", ctx: cancelledRegistryContext(),
			provider: func(context.Context) ([]byte, error) { return []byte("key"), nil }, request: newRegistryRequest(), wantErr: stalwart.ErrCancelled,
		},
		{
			name: "method other than post", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key"), nil },
			request: &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "/api"}, Header: make(http.Header)}, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "jmap path is rejected", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key"), nil },
			request: &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/jmap"}, Header: make(http.Header)}, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "encoded path", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key"), nil },
			request: &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/api", RawPath: "/%61pi"}, Header: make(http.Header)}, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "query", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key"), nil },
			request: &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/api", RawQuery: "next=1"}, Header: make(http.Header)}, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "fragment", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key"), nil },
			request: &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/api", Fragment: "fragment"}, Header: make(http.Header)}, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "userinfo", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key"), nil },
			request: &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/api", User: url.User("user")}, Header: make(http.Header)}, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "pre-existing authorization", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key"), nil },
			request: registryRequestWithAuthorization("Bearer previous"), wantErr: stalwart.ErrUnauthorized, wantAuthorization: "Bearer previous",
		},
		{
			name: "empty lower case authorization", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key"), nil },
			request: registryRequestWithHeader("authorization", ""), wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "nil request", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key"), nil }, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "nil url", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key"), nil },
			request: &http.Request{Method: http.MethodPost, Header: make(http.Header)}, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "nil header", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key"), nil },
			request: &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/api"}}, wantErr: stalwart.ErrUnauthorized,
		},
		{
			name: "provider error", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return nil, errors.New("provider detail must not leak") },
			request: newRegistryRequest(), wantErr: stalwart.ErrUnavailable, wantProviderCalled: true,
		},
		{
			name: "provider panic", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { panic("provider detail must not leak") },
			request: newRegistryRequest(), wantErr: stalwart.ErrUnavailable, wantProviderCalled: true,
		},
		{
			name: "empty key", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return nil, nil }, request: newRegistryRequest(), wantErr: stalwart.ErrUnauthorized, wantProviderCalled: true,
		},
		{
			name: "oversize key", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte(strings.Repeat("a", 4097)), nil }, request: newRegistryRequest(), wantErr: stalwart.ErrUnauthorized, wantProviderCalled: true,
		},
		{
			name: "control key", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key\n"), nil }, request: newRegistryRequest(), wantErr: stalwart.ErrUnauthorized, wantProviderCalled: true,
		},
		{
			name: "whitespace key", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key value"), nil }, request: newRegistryRequest(), wantErr: stalwart.ErrUnauthorized, wantProviderCalled: true,
		},
		{
			name: "non ascii key", ctx: context.Background(),
			provider: func(context.Context) ([]byte, error) { return []byte("key\xc3\xb1"), nil }, request: newRegistryRequest(), wantErr: stalwart.ErrUnauthorized, wantProviderCalled: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			providerCalls := 0
			provider := testCase.provider
			if provider != nil {
				testCase.provider = func(ctx context.Context) ([]byte, error) {
					providerCalls++
					return provider(ctx)
				}
			}
			authorizer, err := stalwart.NewRegistryAuthorizer(testCase.provider)
			if err != nil {
				t.Fatalf("NewRegistryAuthorizer() error = %v", err)
			}
			before := authorizationValue(testCase.request)
			err = authorizer.AuthorizeRegistry(testCase.ctx, testCase.request)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("AuthorizeRegistry() error = %v; want %v", err, testCase.wantErr)
			}
			if testCase.wantErr != nil && err != nil && strings.Contains(err.Error(), "provider detail") {
				t.Fatalf("AuthorizeRegistry() exposed provider detail in %q", err)
			}
			if providerCalls != boolToInt(testCase.wantProviderCalled) {
				t.Fatalf("provider calls = %d; want %d", providerCalls, boolToInt(testCase.wantProviderCalled))
			}
			if got := authorizationValue(testCase.request); got != testCase.wantAuthorization {
				if testCase.wantErr != nil && got != before {
					t.Fatalf("AuthorizeRegistry() mutated Authorization from %q to %q on failure", before, got)
				}
				t.Fatalf("Authorization = %q; want %q", got, testCase.wantAuthorization)
			}
		})
	}
}

func TestRegistryAuthorizer_ProviderBufferRemainsCallerOwned(t *testing.T) {
	key := []byte("scoped-key_123")
	before := append([]byte(nil), key...)
	authorizer, err := stalwart.NewRegistryAuthorizer(func(context.Context) ([]byte, error) { return key, nil })
	if err != nil {
		t.Fatalf("NewRegistryAuthorizer() error = %v", err)
	}
	request := newRegistryRequest()
	if err := authorizer.AuthorizeRegistry(context.Background(), request); err != nil {
		t.Fatalf("AuthorizeRegistry() error = %v", err)
	}
	if string(key) != string(before) {
		t.Fatalf("provider key = %q; want caller-owned buffer unchanged", key)
	}
}

func TestRegistryAuthorizer_ContextCancellationAfterLookupLeavesRequestUntouched(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	authorizer, err := stalwart.NewRegistryAuthorizer(func(context.Context) ([]byte, error) {
		cancel()
		return []byte("scoped-key_123"), nil
	})
	if err != nil {
		t.Fatalf("NewRegistryAuthorizer() error = %v", err)
	}
	request := newRegistryRequest()
	if err := authorizer.AuthorizeRegistry(ctx, request); !errors.Is(err, stalwart.ErrCancelled) {
		t.Fatalf("AuthorizeRegistry() error = %v; want ErrCancelled", err)
	}
	if got := authorizationValue(request); got != "" {
		t.Fatalf("Authorization = %q; want no mutation after cancellation", got)
	}
}

func TestRegistryAuthorizer_NewRejectsNilProvider(t *testing.T) {
	authorizer, err := stalwart.NewRegistryAuthorizer(nil)
	if authorizer != nil || !errors.Is(err, stalwart.ErrInvalidConfiguration) {
		t.Fatalf("NewRegistryAuthorizer(nil) = (%v, %v); want nil, ErrInvalidConfiguration", authorizer, err)
	}
}

func newRegistryRequest() *http.Request {
	return &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/api"}, Header: make(http.Header)}
}

func registryRequestWithAuthorization(value string) *http.Request {
	return registryRequestWithHeader("Authorization", value)
}

func registryRequestWithHeader(name, value string) *http.Request {
	request := newRegistryRequest()
	request.Header[name] = []string{value}
	return request
}

func cancelledRegistryContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func authorizationValue(request *http.Request) string {
	if request == nil || request.Header == nil {
		return ""
	}
	return request.Header.Get("Authorization")
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
