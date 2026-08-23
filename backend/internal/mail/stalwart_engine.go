package mail

import (
	"context"
	"errors"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

// StalwartMailboxClient is the account-bound JMAP boundary used by
// StalwartEngine. AccountID must be the exact account passed to the factory;
// implementations must not select credentials from a JMAP request body.
//
// The composition layer owns session discovery and the credential lifecycle.
// This keeps the parent mail package independent of Stalwart's private session
// details while making account separation an explicit dependency.
type StalwartMailboxClient interface {
	AccountID() string
	CallMailbox(context.Context, stalwart.JMAPRequest) (stalwart.JMAPResponse, error)
	DownloadBlob(context.Context, string, int64) (StalwartMailboxBlob, error)
}

// StalwartMailboxBlob is the provider-neutral result of one bounded download.
// Implementations must use their already validated account/session binding,
// return no provider URL or authorization material, and perform no retry. They
// own any fixed safe download name/type required by the provider transport.
type StalwartMailboxBlob struct {
	AccountID string
	BlobID    string
	MediaType string
	SizeBytes int64
	Bytes     []byte
}

// StalwartMailboxClientFactory validates one bounded mailbox account ID and
// returns a client already bound to that exact account. It must never return a
// shared, globally authorized client for arbitrary accounts.
type StalwartMailboxClientFactory interface {
	OpenMailboxClient(context.Context, string) (StalwartMailboxClient, error)
}

// StalwartEngine is the parent-package adapter for the standards-only Stalwart
// JMAP transport. It intentionally implements only operations whose provider
// semantics are fully represented here; a full MailEngine assertion belongs
// only after every interface operation is implemented.
type StalwartEngine struct {
	mailboxClients StalwartMailboxClientFactory
}

// NewStalwartEngine constructs an adapter without performing network I/O.
func NewStalwartEngine(mailboxClients StalwartMailboxClientFactory) (*StalwartEngine, error) {
	if mailboxClients == nil {
		return nil, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return &StalwartEngine{mailboxClients: mailboxClients}, nil
}

// Capabilities reports only properties proven by this partial adapter. Folder
// mutation authority, projection enumeration, and draft/submission recovery
// remain disabled until their independent provider and deployment proofs exist.
func (engine *StalwartEngine) Capabilities(context.Context) (EngineCapabilities, error) {
	if engine == nil || engine.mailboxClients == nil {
		return EngineCapabilities{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return EngineCapabilities{}, nil
}

// Health deliberately reports an unknown global state. This account-scoped
// adapter has no provider-wide probe or representative mailbox credential, so
// it must not synthesize a ready result from construction alone.
func (engine *StalwartEngine) Health(context.Context) (EngineHealth, error) {
	if engine == nil || engine.mailboxClients == nil {
		return EngineHealth{State: EngineHealthUnknown}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return EngineHealth{State: EngineHealthUnknown}, nil
}

func (engine *StalwartEngine) mailboxClient(ctx context.Context, accountID string) (StalwartMailboxClient, error) {
	if engine == nil || engine.mailboxClients == nil || ctx == nil || !validAccountID(accountID) {
		return nil, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	client, err := engine.mailboxClients.OpenMailboxClient(ctx, accountID)
	if err != nil {
		return nil, mapStalwartEngineError(err, false)
	}
	if client == nil || client.AccountID() != accountID {
		return nil, stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	return client, nil
}

func (engine *StalwartEngine) callMailbox(ctx context.Context, accountID string, request stalwart.JMAPRequest, mutation bool) (stalwart.JMAPResponse, error) {
	client, err := engine.mailboxClient(ctx, accountID)
	if err != nil {
		return stalwart.JMAPResponse{}, err
	}
	response, err := client.CallMailbox(ctx, request)
	if err != nil {
		return stalwart.JMAPResponse{}, mapStalwartEngineError(err, mutation)
	}
	return response, nil
}

func stalwartEngineError(kind EngineErrorKind) error {
	return &EngineError{Kind: kind}
}

// mapStalwartEngineError intentionally drops all provider-controlled details.
// A mutation transport ambiguity is never retry-labelled: callers receive an
// indeterminate result and must reconcile from the provider's authoritative
// state before attempting another mutation.
func mapStalwartEngineError(err error, mutation bool) error {
	if err == nil {
		return nil
	}
	providerKind, known := stalwartProviderErrorKind(err)
	if !known {
		if mutation {
			return stalwartEngineError(EngineErrorIndeterminate)
		}
		return stalwartEngineError(EngineErrorUnavailable)
	}
	switch providerKind {
	case stalwart.ErrorUnauthorized:
		return stalwartEngineError(EngineErrorUnauthorizedAccount)
	case stalwart.ErrorNotFound:
		return stalwartEngineError(EngineErrorNotFound)
	case stalwart.ErrorStateMismatch:
		return stalwartEngineError(EngineErrorConflict)
	case stalwart.ErrorRestartRequired:
		return stalwartEngineError(EngineErrorRestartRequired)
	case stalwart.ErrorPayloadTooLarge:
		return stalwartEngineError(EngineErrorPayloadTooLarge)
	case stalwart.ErrorIndeterminate:
		return stalwartEngineError(EngineErrorIndeterminate)
	case stalwart.ErrorUnavailable, stalwart.ErrorCancelled, stalwart.ErrorProtocol:
		if mutation {
			return stalwartEngineError(EngineErrorIndeterminate)
		}
		return stalwartEngineError(EngineErrorUnavailable)
	case stalwart.ErrorInvalidConfiguration, stalwart.ErrorInvalidRequest, stalwart.ErrorUnsupported:
		return stalwartEngineError(EngineErrorRejectedPolicy)
	default:
		if mutation {
			return stalwartEngineError(EngineErrorIndeterminate)
		}
		return stalwartEngineError(EngineErrorUnavailable)
	}
}

func stalwartProviderErrorKind(err error) (stalwart.ErrorKind, bool) {
	if kind, ok := stalwart.KindOf(err); ok {
		return kind, true
	}
	switch {
	case errors.Is(err, stalwart.ErrInvalidConfiguration):
		return stalwart.ErrorInvalidConfiguration, true
	case errors.Is(err, stalwart.ErrInvalidRequest):
		return stalwart.ErrorInvalidRequest, true
	case errors.Is(err, stalwart.ErrUnauthorized):
		return stalwart.ErrorUnauthorized, true
	case errors.Is(err, stalwart.ErrNotFound):
		return stalwart.ErrorNotFound, true
	case errors.Is(err, stalwart.ErrStateMismatch):
		return stalwart.ErrorStateMismatch, true
	case errors.Is(err, stalwart.ErrRestartRequired):
		return stalwart.ErrorRestartRequired, true
	case errors.Is(err, stalwart.ErrPayloadTooLarge):
		return stalwart.ErrorPayloadTooLarge, true
	case errors.Is(err, stalwart.ErrUnsupported):
		return stalwart.ErrorUnsupported, true
	case errors.Is(err, stalwart.ErrProtocol):
		return stalwart.ErrorProtocol, true
	case errors.Is(err, stalwart.ErrUnavailable):
		return stalwart.ErrorUnavailable, true
	case errors.Is(err, stalwart.ErrIndeterminate):
		return stalwart.ErrorIndeterminate, true
	case errors.Is(err, stalwart.ErrCancelled):
		return stalwart.ErrorCancelled, true
	default:
		return "", false
	}
}
