package stalwart

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

const registryAPIKeyMaximumBytes = 4096

// RegistryManagementAPIKeyProvider obtains one scoped Stalwart management API
// key for one Registry request. Returned bytes remain owned by the provider.
type RegistryManagementAPIKeyProvider func(context.Context) ([]byte, error)

type registryAuthorizer struct {
	provider RegistryManagementAPIKeyProvider
}

// NewRegistryAuthorizer creates a fail-closed Registry bearer boundary. It
// performs no credential lookup at construction time and never retains a key.
func NewRegistryAuthorizer(provider RegistryManagementAPIKeyProvider) (RegistryAuthorizer, error) {
	if provider == nil {
		return nil, ErrInvalidConfiguration
	}
	return &registryAuthorizer{provider: provider}, nil
}

func (authorizer *registryAuthorizer) AuthorizeRegistry(ctx context.Context, request *http.Request) error {
	if authorizer == nil || authorizer.provider == nil || !validRegistryRequest(request) {
		return ErrUnauthorized
	}
	if err := registryAuthorizationContextError(ctx); err != nil {
		return err
	}

	rawKey, providerErr, panicked := authorizer.lookup(ctx)
	if panicked || providerErr != nil {
		return ErrUnavailable
	}
	if err := registryAuthorizationContextError(ctx); err != nil {
		return err
	}

	if len(rawKey) < 1 || len(rawKey) > registryAPIKeyMaximumBytes {
		return ErrUnauthorized
	}
	if !validRegistryAPIKey(rawKey) {
		return ErrUnauthorized
	}
	key := append([]byte(nil), rawKey...)
	defer clear(key)
	if err := registryAuthorizationContextError(ctx); err != nil {
		return err
	}

	value := "Bearer " + string(key)
	request.Header.Set("Authorization", value)
	value = ""
	return nil
}

func (authorizer *registryAuthorizer) lookup(ctx context.Context) (key []byte, err error, panicked bool) {
	defer func() {
		if recover() != nil {
			key = nil
			err = nil
			panicked = true
		}
	}()
	key, err = authorizer.provider(ctx)
	return key, err, false
}

func validRegistryRequest(request *http.Request) bool {
	if request == nil || request.Header == nil || request.URL == nil || request.Method != http.MethodPost || request.URL.Path != "/jmap" {
		return false
	}
	for name := range request.Header {
		if strings.EqualFold(name, "Authorization") {
			return false
		}
	}
	return request.URL.RawPath == "" && request.URL.RawQuery == "" && !request.URL.ForceQuery && request.URL.Fragment == "" && request.URL.User == nil
}

func registryAuthorizationContextError(ctx context.Context) error {
	if ctx == nil {
		return ErrUnauthorized
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return ErrCancelled
		}
		return ErrUnavailable
	}
	return nil
}

func validRegistryAPIKey(key []byte) bool {
	if len(key) < 1 || len(key) > registryAPIKeyMaximumBytes {
		return false
	}
	for _, value := range key {
		if value < 0x21 || value > 0x7e {
			return false
		}
	}
	return true
}

var _ RegistryAuthorizer = (*registryAuthorizer)(nil)
