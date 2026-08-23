package stalwart

import "errors"

// OAuthErrorKind is a content-free OAuth protocol failure classification.
type OAuthErrorKind string

const (
	OAuthErrorInvalidConfiguration OAuthErrorKind = "invalid_configuration"
	OAuthErrorInvalidRequest       OAuthErrorKind = "invalid_request"
	OAuthErrorInvalidGrant         OAuthErrorKind = "invalid_grant"
	OAuthErrorInvalidClient        OAuthErrorKind = "invalid_client"
	OAuthErrorProtocol             OAuthErrorKind = "protocol_violation"
	OAuthErrorPayloadTooLarge      OAuthErrorKind = "payload_too_large"
	OAuthErrorUnavailable          OAuthErrorKind = "unavailable"
	OAuthErrorCancelled            OAuthErrorKind = "cancelled"
)

var (
	ErrOAuthInvalidConfiguration = errors.New("oauth configuration invalid")
	ErrOAuthInvalidRequest       = errors.New("oauth request invalid")
	ErrOAuthInvalidGrant         = errors.New("oauth grant invalid")
	ErrOAuthInvalidClient        = errors.New("oauth client invalid")
	ErrOAuthProtocol             = errors.New("oauth protocol violation")
	ErrOAuthPayloadTooLarge      = errors.New("oauth payload too large")
	ErrOAuthUnavailable          = errors.New("oauth service unavailable")
	ErrOAuthCancelled            = errors.New("oauth request cancelled")
)

// OAuthError never retains provider bodies, URLs, tokens, account identifiers,
// or wrapped source errors.
type OAuthError struct {
	Kind OAuthErrorKind
}

func (err *OAuthError) Error() string {
	if err == nil {
		return "oauth error"
	}
	return oauthErrorForKind(err.Kind).Error()
}

func (err *OAuthError) Unwrap() error {
	if err == nil {
		return nil
	}
	return oauthErrorForKind(err.Kind)
}

// OAuthKindOf returns the stable classification without provider-controlled
// details.
func OAuthKindOf(err error) (OAuthErrorKind, bool) {
	var oauthError *OAuthError
	if !errors.As(err, &oauthError) || oauthError == nil {
		return "", false
	}
	return oauthError.Kind, true
}

func newOAuthError(kind OAuthErrorKind) error {
	return &OAuthError{Kind: kind}
}

func oauthErrorForKind(kind OAuthErrorKind) error {
	switch kind {
	case OAuthErrorInvalidConfiguration:
		return ErrOAuthInvalidConfiguration
	case OAuthErrorInvalidRequest:
		return ErrOAuthInvalidRequest
	case OAuthErrorInvalidGrant:
		return ErrOAuthInvalidGrant
	case OAuthErrorInvalidClient:
		return ErrOAuthInvalidClient
	case OAuthErrorPayloadTooLarge:
		return ErrOAuthPayloadTooLarge
	case OAuthErrorUnavailable:
		return ErrOAuthUnavailable
	case OAuthErrorCancelled:
		return ErrOAuthCancelled
	case OAuthErrorProtocol:
		fallthrough
	default:
		return ErrOAuthProtocol
	}
}
