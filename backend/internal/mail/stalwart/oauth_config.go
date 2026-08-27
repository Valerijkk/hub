package stalwart

import (
	"context"
	"crypto/rand"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	oauthDiscoveryPath          = "/.well-known/oauth-authorization-server"
	oauthAuthorizationPath      = "/login"
	oauthTokenPath              = "/auth/token"
	oauthScope                  = "offline_access mail"
	oauthStateBytes             = 32
	oauthVerifierBytes          = 64
	oauthStateLength            = 43
	oauthVerifierLength         = 86
	oauthChallengeLength        = 43
	oauthMaxClientIDBytes       = 256
	oauthMaxClientSecretBytes   = 4096
	oauthMaxRedirectURIBytes    = 2048
	oauthMaxAuthorizationBytes  = 8192
	oauthMaxCodeBytes           = 4096
	oauthMaxTokenBytes          = 16384
	oauthMaxDiscoveryBytes     int64 = 32 << 10
	oauthMaxTokenResponseBytes int64 = 64 << 10
	oauthMaxExpiresSeconds     int64 = 365 * 24 * 60 * 60
)

// OAuthClientSecretProvider returns a fresh caller-owned secret for one exact
// public client ID. Implementations must not provide a fallback credential;
// the caller clears the returned byte slice after constructing Basic auth.
type OAuthClientSecretProvider interface {
	OAuthClientSecret(ctx context.Context, clientID string) ([]byte, error)
}

// OAuthConfig contains fixed public protocol coordinates and injected private
// dependencies. It contains neither client secrets nor user tokens.
type OAuthConfig struct {
	IssuerOrigin         string
	ResourceOrigin       string
	TransportOrigin      string
	ClientID             string
	ClientSecretProvider OAuthClientSecretProvider
	RedirectURI          string
	HTTPClient           *http.Client
	Random               io.Reader
}

// OAuthClient retains only validated public values and opaque dependency
// boundaries. Construction performs no network or credential-provider I/O.
type OAuthClient struct {
	issuerOrigin    string
	resourceOrigin  string
	transportOrigin string
	clientID        string
	clientSecrets   OAuthClientSecretProvider
	redirectURI     string
	httpClient      *http.Client
	random          io.Reader
}

// NewOAuthClient validates the complete immutable OAuth configuration.
func NewOAuthClient(config OAuthConfig) (*OAuthClient, error) {
	issuerOrigin, ok := canonicalOAuthOrigin(config.IssuerOrigin)
	if !ok {
		return nil, newOAuthError(OAuthErrorInvalidConfiguration)
	}
	resourceOrigin, ok := canonicalOAuthOrigin(config.ResourceOrigin)
	if !ok {
		return nil, newOAuthError(OAuthErrorInvalidConfiguration)
	}
	transportOrigin, ok := canonicalOAuthOrigin(config.TransportOrigin)
	if !ok || !validOAuthClientID(config.ClientID) || config.ClientSecretProvider == nil || config.HTTPClient == nil {
		return nil, newOAuthError(OAuthErrorInvalidConfiguration)
	}
	redirectURI, ok := canonicalOAuthRedirectURI(config.RedirectURI)
	if !ok {
		return nil, newOAuthError(OAuthErrorInvalidConfiguration)
	}
	randomSource := config.Random
	if randomSource == nil {
		randomSource = rand.Reader
	}
	return &OAuthClient{
		issuerOrigin:    issuerOrigin,
		resourceOrigin:  resourceOrigin,
		transportOrigin: transportOrigin,
		clientID:        config.ClientID,
		clientSecrets:   config.ClientSecretProvider,
		redirectURI:     redirectURI,
		httpClient:      config.HTTPClient,
		random:          randomSource,
	}, nil
}

func canonicalOAuthOrigin(value string) (string, bool) {
	parsed, ok := parseOrigin(value)
	if !ok {
		return "", false
	}
	return parsed.String(), true
}

func canonicalOAuthRedirectURI(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > oauthMaxRedirectURIBytes || !utf8.ValidString(value) {
		return "", false
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || parsed.Opaque != "" || parsed.Hostname() == "" || parsed.Path == "" || parsed.Path == "/" || parsed.RawPath != "" {
		return "", false
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return "", false
	}
	return value, true
}

func validOAuthClientID(value string) bool {
	if value == "" || len(value) > oauthMaxClientIDBytes || !utf8.ValidString(value) || strings.ContainsRune(value, ':') {
		return false
	}
	for _, character := range value {
		if character <= 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}
