// Package stalwart contains the deliberately narrow Stalwart v0.16.17 JMAP
// transport. Provider response data never becomes part of an error value.
package stalwart

import "errors"

// ErrorKind is a stable, content-free classification for transport failures.
// It deliberately carries no upstream URL, response body, identifier, token or
// SMTP diagnostic.
type ErrorKind string

const (
	ErrorInvalidConfiguration ErrorKind = "invalid_configuration"
	ErrorInvalidRequest       ErrorKind = "invalid_request"
	ErrorUnauthorized         ErrorKind = "unauthorized"
	ErrorNotFound             ErrorKind = "not_found"
	ErrorStateMismatch        ErrorKind = "state_mismatch"
	ErrorRestartRequired      ErrorKind = "restart_required"
	ErrorPayloadTooLarge      ErrorKind = "payload_too_large"
	ErrorUnsupported          ErrorKind = "unsupported"
	ErrorProtocol             ErrorKind = "protocol_violation"
	ErrorUnavailable          ErrorKind = "unavailable"
	ErrorIndeterminate        ErrorKind = "indeterminate"
	ErrorCancelled            ErrorKind = "cancelled"
)

var (
	ErrInvalidConfiguration = errors.New("stalwart transport configuration invalid")
	ErrInvalidRequest       = errors.New("stalwart transport request invalid")
	ErrUnauthorized         = errors.New("stalwart transport unauthorized")
	ErrNotFound             = errors.New("stalwart transport object not found")
	ErrStateMismatch        = errors.New("stalwart transport state mismatch")
	ErrRestartRequired      = errors.New("stalwart transport restart required")
	ErrPayloadTooLarge      = errors.New("stalwart transport payload too large")
	ErrUnsupported          = errors.New("stalwart transport unsupported")
	ErrProtocol             = errors.New("stalwart transport protocol violation")
	ErrUnavailable          = errors.New("stalwart transport unavailable")
	ErrIndeterminate        = errors.New("stalwart transport result indeterminate")
	ErrCancelled            = errors.New("stalwart transport cancelled")
)

// Error is intentionally detail-free. Callers can branch with KindOf or
// errors.Is without receiving provider-controlled data.
type Error struct {
	Kind ErrorKind
}

func (e *Error) Error() string {
	if e == nil {
		return "stalwart transport error"
	}
	return errorForKind(e.Kind).Error()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return errorForKind(e.Kind)
}

// KindOf returns the stable classification without exposing an upstream cause.
func KindOf(err error) (ErrorKind, bool) {
	var transportError *Error
	if !errors.As(err, &transportError) || transportError == nil {
		return "", false
	}
	return transportError.Kind, true
}

func newError(kind ErrorKind) error {
	return &Error{Kind: kind}
}

func errorForKind(kind ErrorKind) error {
	switch kind {
	case ErrorInvalidConfiguration:
		return ErrInvalidConfiguration
	case ErrorInvalidRequest:
		return ErrInvalidRequest
	case ErrorUnauthorized:
		return ErrUnauthorized
	case ErrorNotFound:
		return ErrNotFound
	case ErrorStateMismatch:
		return ErrStateMismatch
	case ErrorRestartRequired:
		return ErrRestartRequired
	case ErrorPayloadTooLarge:
		return ErrPayloadTooLarge
	case ErrorUnsupported:
		return ErrUnsupported
	case ErrorProtocol:
		return ErrProtocol
	case ErrorUnavailable:
		return ErrUnavailable
	case ErrorIndeterminate:
		return ErrIndeterminate
	case ErrorCancelled:
		return ErrCancelled
	default:
		return ErrProtocol
	}
}

func jmapErrorKind(errorType string) ErrorKind {
	switch errorType {
	case "stateMismatch":
		return ErrorStateMismatch
	case "cannotCalculateChanges", "anchorNotFound":
		return ErrorRestartRequired
	case "notFound":
		return ErrorNotFound
	case "accountNotFound", "forbidden", "accountNotSupported":
		return ErrorUnauthorized
	case "tooLarge", "requestTooLarge":
		return ErrorPayloadTooLarge
	case "unknownMethod", "unknownCapability", "unsupportedFilter", "unsupportedSort":
		return ErrorUnsupported
	case "serverFail", "serverUnavailable", "rateLimit":
		return ErrorUnavailable
	case "invalidArguments", "invalidProperties", "invalidPatch", "willDestroy", "invalidResultReference":
		return ErrorInvalidRequest
	default:
		return ErrorProtocol
	}
}
