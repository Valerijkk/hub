package mail

import (
	"context"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const mailboxOAuthRuntimeClientIDMaximum = 256

// MailboxOAuthClientSecretLookup retrieves one secret for one fixed public
// OAuth client ID. The returned bytes remain owned by the lookup.
type MailboxOAuthClientSecretLookup func(context.Context, string) ([]byte, error)

type mailboxSyncWorkerRuntimeTimer struct{}

type stalwartMailboxRegistryAuthorizer struct{}

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

var _ MailboxSyncWorkerTimer = mailboxSyncWorkerRuntimeTimer{}
var _ stalwart.RegistryAuthorizer = stalwartMailboxRegistryAuthorizer{}
var _ stalwart.OAuthClientSecretProvider = (*stalwartOAuthClientSecretProvider)(nil)
