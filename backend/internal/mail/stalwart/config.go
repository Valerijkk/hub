package stalwart

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	// Version is the selected upstream protocol source, not a runtime approval.
	Version = "v0.16.17"

	maxOriginBytes          = 2048
	maxOpaqueIDBytes        = 1024
	maxCallIDBytes          = 128
	maxContentTypeBytes     = 256
	maxJSONDepth            = 64
	maxJSONTokens           = 8192
	maxJMAPMethodCalls      = 16
	maxSessionBytes   int64 = 1 << 20
	maxJMAPBytes      int64 = 8 << 20

	// Upload and download limits remain local so the transport cannot create an
	// import cycle with the generic mail engine package.
	maxUploadContentBytes int64 = 50 << 20
	maxDownloadedBlobBytes int64 = 25 << 20
)

// RegistryAuthorizer owns the least-privilege Registry credential lifecycle.
// It may apply a scoped credential to request, but must not retain it in this
// transport. Registry authorizers are never used for mailbox traffic.
type RegistryAuthorizer interface {
	AuthorizeRegistry(context.Context, *http.Request) error
}

// MailboxAuthorizer is an account-scoped credential boundary supplied by
// composition. It is the only place an OAuth access token may be obtained or
// attached; this transport does not implement passwords or client credentials.
type MailboxAuthorizer interface {
	AuthorizeMailbox(context.Context, *http.Request) error
}

// Config contains no credentials. Origin is one trusted absolute origin with
// no path, query, fragment or userinfo. HTTPClient is injected for composition
// and tests; redirects are still disabled per individual request.
type Config struct {
	Origin            string
	HTTPClient        *http.Client
	RegistryAuthorizer RegistryAuthorizer
	MailboxAuthorizer  MailboxAuthorizer
}

// Client has no network side effects at construction time. It retains only the
// fixed trusted origin, injected client and typed authorizer boundaries.
type Client struct {
	origin             *url.URL
	httpClient         *http.Client
	registryAuthorizer RegistryAuthorizer
	mailboxAuthorizer  MailboxAuthorizer
}

// NewClient validates the only base URL accepted by the transport. Both typed
// authorizers are required, preventing accidental fallback or credential reuse.
func NewClient(config Config) (*Client, error) {
	origin, ok := parseOrigin(config.Origin)
	if !ok || config.HTTPClient == nil || config.RegistryAuthorizer == nil || config.MailboxAuthorizer == nil {
		return nil, newError(ErrorInvalidConfiguration)
	}
	return &Client{
		origin:             origin,
		httpClient:         config.HTTPClient,
		registryAuthorizer: config.RegistryAuthorizer,
		mailboxAuthorizer:  config.MailboxAuthorizer,
	}, nil
}

func parseOrigin(value string) (*url.URL, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxOriginBytes || !utf8.ValidString(value) {
		return nil, false
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || parsed.Opaque != "" {
		return nil, false
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return nil, false
	}
	if parsed.Path != "" && parsed.Path != "/" || parsed.RawPath != "" && parsed.RawPath != "/" {
		return nil, false
	}
	if parsed.Hostname() == "" {
		return nil, false
	}
	return &url.URL{Scheme: strings.ToLower(parsed.Scheme), Host: strings.ToLower(parsed.Host)}, true
}

func (c *Client) sameOrigin(candidate *url.URL) bool {
	return c != nil && c.origin != nil && candidate != nil && strings.EqualFold(c.origin.Scheme, candidate.Scheme) && strings.EqualFold(c.origin.Host, candidate.Host)
}

func (c *Client) fixedURL(path string) *url.URL {
	if c == nil || c.origin == nil {
		return nil
	}
	return &url.URL{Scheme: c.origin.Scheme, Host: c.origin.Host, Path: path}
}

func validOpaque(value string) bool {
	if value == "" || len(value) > maxOpaqueIDBytes || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if character <= 0x1f || character == 0x7f {
			return false
		}
	}
	return true
}

func validCallID(value string) bool {
	return validOpaque(value) && len(value) <= maxCallIDBytes
}
