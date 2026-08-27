package mail

import (
	"context"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	mailboxOAuthRuntimeClientIDMaximum       = 256
	mailboxRegistryAPIKeyOriginMaximumBytes = 2048
)

// MailboxOAuthClientSecretLookup retrieves one secret for one fixed public
// OAuth client ID. The returned bytes remain owned by the lookup.
type MailboxOAuthClientSecretLookup func(context.Context, string) ([]byte, error)

// MailboxRegistryAPIKeyLookup retrieves one scoped Registry API key for one
// exact trusted Stalwart origin. The returned bytes remain owned by the lookup.
type MailboxRegistryAPIKeyLookup func(context.Context) ([]byte, error)

type mailboxSyncWorkerRuntimeTimer struct{}

type stalwartMailboxRegistryAuthorizer struct{}

type stalwartRegistryMailboxAuthorizer struct{}

type stalwartRegistryAPIKeyAuthorizer struct {
	origin   *url.URL
	delegate stalwart.RegistryAuthorizer
}

type stalwartOAuthClientSecretProvider struct {
	fixedClientID string
	lookup        MailboxOAuthClientSecretLookup
}

// NewMailboxSyncWorkerTimer supplies the production context-aware timer for a
// MailboxSyncWorker. It has no background goroutine or retained state.
func NewMailboxSyncWorkerTimer() MailboxSyncWorkerTimer {
	return mailboxSyncWorkerRuntimeTimer{}
}

func (mailboxSyncWorkerRuntimeTimer) Wait(ctx context.Context, duration time.Duration) error {
	if ctx == nil {
		return context.Canceled
	}
	timer := time.NewTimer(duration)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (mailboxSyncWorkerRuntimeTimer) WithTimeout(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return nil, nil
	}
	return context.WithTimeout(ctx, duration)
}

// NewStalwartMailboxRegistryAuthorizer returns the fail-closed Registry slot
// required by the typed client used only for mailbox operations.
func NewStalwartMailboxRegistryAuthorizer() stalwart.RegistryAuthorizer {
	return stalwartMailboxRegistryAuthorizer{}
}

func (stalwartMailboxRegistryAuthorizer) AuthorizeRegistry(context.Context, *http.Request) error {
	return stalwart.ErrUnauthorized
}

// NewStalwartRegistryClient constructs the fixed-origin Registry transport.
// Mailbox traffic is deliberately denied because this composition has no
// account-scoped mailbox credential boundary.
func NewStalwartRegistryClient(origin string, httpClient *http.Client, registryAuthorizer stalwart.RegistryAuthorizer) (*stalwart.Client, error) {
	if stalwartRegistryAuthorizerNil(registryAuthorizer) {
		return nil, stalwart.ErrInvalidConfiguration
	}
	return stalwart.NewClient(stalwart.Config{
		Origin:             origin,
		HTTPClient:         httpClient,
		RegistryAuthorizer: registryAuthorizer,
		MailboxAuthorizer:  stalwartRegistryMailboxAuthorizer{},
	})
}

func (stalwartRegistryMailboxAuthorizer) AuthorizeMailbox(context.Context, *http.Request) error {
	return stalwart.ErrUnauthorized
}

func stalwartRegistryAuthorizerNil(authorizer stalwart.RegistryAuthorizer) bool {
	if authorizer == nil {
		return true
	}
	value := reflect.ValueOf(authorizer)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// NewStalwartRegistryAPIKeyAuthorizer binds one scoped API-key lookup to one
// canonical Stalwart Registry origin. Construction performs no lookup or I/O.
func NewStalwartRegistryAPIKeyAuthorizer(origin string, lookup MailboxRegistryAPIKeyLookup) (stalwart.RegistryAuthorizer, error) {
	trustedOrigin, ok := canonicalMailboxRegistryAPIKeyOrigin(origin)
	if !ok || lookup == nil {
		return nil, stalwart.ErrInvalidConfiguration
	}
	delegate, err := stalwart.NewRegistryAuthorizer(stalwart.RegistryManagementAPIKeyProvider(lookup))
	if err != nil {
		return nil, err
	}
	return &stalwartRegistryAPIKeyAuthorizer{origin: trustedOrigin, delegate: delegate}, nil
}

func (authorizer *stalwartRegistryAPIKeyAuthorizer) AuthorizeRegistry(ctx context.Context, request *http.Request) error {
	if authorizer == nil || authorizer.origin == nil || authorizer.delegate == nil || !validMailboxRegistryAPIKeyRequest(request, authorizer.origin) {
		return stalwart.ErrUnauthorized
	}
	return authorizer.delegate.AuthorizeRegistry(ctx, request)
}

// NewStalwartOAuthClientSecretProvider binds secret retrieval to one exact
// OAuth client ID. Construction performs no credential lookup.
func NewStalwartOAuthClientSecretProvider(fixedClientID string, lookup MailboxOAuthClientSecretLookup) (stalwart.OAuthClientSecretProvider, error) {
	if !validMailboxOAuthRuntimeClientID(fixedClientID) || lookup == nil {
		return nil, stalwart.ErrOAuthInvalidConfiguration
	}
	return &stalwartOAuthClientSecretProvider{fixedClientID: fixedClientID, lookup: lookup}, nil
}

func (provider *stalwartOAuthClientSecretProvider) OAuthClientSecret(ctx context.Context, clientID string) ([]byte, error) {
	if provider == nil || provider.lookup == nil || clientID != provider.fixedClientID {
		return nil, stalwart.ErrOAuthInvalidClient
	}
	if err := mailboxOAuthRuntimeContextError(ctx); err != nil {
		return nil, err
	}

	raw, lookupErr, panicked := provider.lookupSecret(ctx)
	if err := mailboxOAuthRuntimeContextError(ctx); err != nil {
		return nil, err
	}
	if panicked || lookupErr != nil {
		return nil, stalwart.ErrOAuthUnavailable
	}

	validated := append([]byte(nil), raw...)
	if !validMailboxOAuthRuntimeClientSecret(validated) {
		clear(validated)
		return nil, stalwart.ErrOAuthInvalidClient
	}
	if err := mailboxOAuthRuntimeContextError(ctx); err != nil {
		clear(validated)
		return nil, err
	}
	secret := append([]byte(nil), validated...)
	clear(validated)
	return secret, nil
}

func (provider *stalwartOAuthClientSecretProvider) lookupSecret(ctx context.Context) (secret []byte, err error, panicked bool) {
	defer func() {
		if recover() != nil {
			secret = nil
			err = nil
			panicked = true
		}
	}()
	secret, err = provider.lookup(ctx, provider.fixedClientID)
	return secret, err, false
}

func mailboxOAuthRuntimeContextError(ctx context.Context) error {
	if ctx == nil {
		return stalwart.ErrOAuthInvalidRequest
	}
	if err := ctx.Err(); err != nil {
		if err == context.Canceled {
			return stalwart.ErrOAuthCancelled
		}
		return stalwart.ErrOAuthUnavailable
	}
	return nil
}

func validMailboxOAuthRuntimeClientID(value string) bool {
	if value == "" || len(value) > mailboxOAuthRuntimeClientIDMaximum || !utf8.ValidString(value) || strings.ContainsRune(value, ':') {
		return false
	}
	for _, character := range value {
		if character <= 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}

func validMailboxOAuthRuntimeClientSecret(value []byte) bool {
	if len(value) < 1 || len(value) > 4096 || !utf8.Valid(value) {
		return false
	}
	for len(value) > 0 {
		character, size := utf8.DecodeRune(value)
		if unicode.IsControl(character) {
			return false
		}
		value = value[size:]
	}
	return true
}

func canonicalMailboxRegistryAPIKeyOrigin(value string) (*url.URL, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > mailboxRegistryAPIKeyOriginMaximumBytes || !utf8.ValidString(value) {
		return nil, false
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || parsed.Opaque != "" {
		return nil, false
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return nil, false
	}
	if parsed.Path != "" && parsed.Path != "/" || parsed.RawPath != "" && parsed.RawPath != "/" || parsed.Hostname() == "" {
		return nil, false
	}
	return &url.URL{Scheme: strings.ToLower(parsed.Scheme), Host: strings.ToLower(parsed.Host)}, true
}

func validMailboxRegistryAPIKeyRequest(request *http.Request, origin *url.URL) bool {
	if request == nil || request.Header == nil || request.URL == nil || origin == nil || request.Method != http.MethodPost || request.URL.Scheme == "" || request.URL.Host == "" || request.URL.Path != "/jmap" || !strings.EqualFold(request.URL.Scheme, origin.Scheme) || !strings.EqualFold(request.URL.Host, origin.Host) {
		return false
	}
	for name := range request.Header {
		if strings.EqualFold(name, "Authorization") {
			return false
		}
	}
	return request.URL.RawPath == "" && request.URL.RawQuery == "" && !request.URL.ForceQuery && request.URL.Fragment == "" && request.URL.RawFragment == "" && request.URL.User == nil && request.URL.Opaque == ""
}

var _ MailboxSyncWorkerTimer = mailboxSyncWorkerRuntimeTimer{}
var _ stalwart.RegistryAuthorizer = stalwartMailboxRegistryAuthorizer{}
var _ stalwart.MailboxAuthorizer = stalwartRegistryMailboxAuthorizer{}
var _ stalwart.RegistryAuthorizer = (*stalwartRegistryAPIKeyAuthorizer)(nil)
var _ stalwart.OAuthClientSecretProvider = (*stalwartOAuthClientSecretProvider)(nil)
