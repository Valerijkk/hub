package mail

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	stalwartMailboxOriginMaxBytes = 2048
	stalwartBlobDownloadName      = "attachment"
	stalwartBlobDownloadType      = "application/octet-stream"
)

// MailboxAuthorizerBinding proves which exact mailbox account an opaque OAuth
// bearer authorizer belongs to. The bearer token itself never crosses this
// boundary.
type MailboxAuthorizerBinding struct {
	AccountID  string
	Authorizer stalwart.MailboxAuthorizer
}

// MailboxAuthorizerSource resolves only a delegated OAuth bearer authorizer
// for the requested account. Implementations must reject passwords, client
// credentials, cross-account reuse, and every fallback credential path.
type MailboxAuthorizerSource interface {
	OAuthBearerAuthorizer(context.Context, string) (MailboxAuthorizerBinding, error)
}

type productionStalwartMailboxClientFactory struct {
	origin               string
	httpClient           *http.Client
	registryAuthorizer   stalwart.RegistryAuthorizer
	mailboxAuthorizers   MailboxAuthorizerSource
}

type accountBoundStalwartMailboxClient struct {
	accountID string
	client    *stalwart.Client
}

// NewStalwartMailboxClientFactory validates fixed dependencies without I/O.
// The shared HTTP client remains bounded by the transport's per-call contexts;
// the Registry authorizer is retained only because Stalwart's typed client
// requires the provisioning boundary and is never reachable through this
// mailbox-only wrapper.
func NewStalwartMailboxClientFactory(origin string, httpClient *http.Client, registryAuthorizer stalwart.RegistryAuthorizer, mailboxAuthorizers MailboxAuthorizerSource) (StalwartMailboxClientFactory, error) {
	canonicalOrigin, ok := canonicalStalwartMailboxOrigin(origin)
	if !ok || httpClient == nil || registryAuthorizer == nil || mailboxAuthorizers == nil {
		return nil, stalwart.ErrInvalidConfiguration
	}
	return &productionStalwartMailboxClientFactory{
		origin:             canonicalOrigin,
		httpClient:         httpClient,
		registryAuthorizer: registryAuthorizer,
		mailboxAuthorizers: mailboxAuthorizers,
	}, nil
}

func (factory *productionStalwartMailboxClientFactory) OpenMailboxClient(ctx context.Context, accountID string) (StalwartMailboxClient, error) {
	if factory == nil || factory.httpClient == nil || factory.registryAuthorizer == nil || factory.mailboxAuthorizers == nil || ctx == nil || !validAccountID(accountID) {
		return nil, stalwart.ErrInvalidRequest
	}
	if err := stalwartMailboxContextError(ctx); err != nil {
		return nil, err
	}
	binding, err := factory.mailboxAuthorizers.OAuthBearerAuthorizer(ctx, accountID)
	if err != nil {
		return nil, safeStalwartAuthorizerSourceError(ctx, err)
	}
	if err := stalwartMailboxContextError(ctx); err != nil {
		return nil, err
	}
	if binding.AccountID != accountID || binding.Authorizer == nil {
		return nil, stalwart.ErrUnauthorized
	}
	client, err := stalwart.NewClient(stalwart.Config{
		Origin:             factory.origin,
		HTTPClient:         factory.httpClient,
		RegistryAuthorizer: factory.registryAuthorizer,
		MailboxAuthorizer:  binding.Authorizer,
	})
	if err != nil {
		return nil, stalwart.ErrInvalidConfiguration
	}
	return &accountBoundStalwartMailboxClient{accountID: accountID, client: client}, nil
}

func (client *accountBoundStalwartMailboxClient) AccountID() string {
	if client == nil {
		return ""
	}
	return client.accountID
}

func (client *accountBoundStalwartMailboxClient) CallMailbox(ctx context.Context, request stalwart.JMAPRequest) (stalwart.JMAPResponse, error) {
	if client == nil || client.client == nil || !validAccountID(client.accountID) || ctx == nil {
		return stalwart.JMAPResponse{}, stalwart.ErrInvalidRequest
	}
	if err := stalwartMailboxContextError(ctx); err != nil {
		return stalwart.JMAPResponse{}, err
	}
	session, err := client.client.DiscoverMailboxSession(ctx)
	if err != nil {
		return stalwart.JMAPResponse{}, err
	}
	if err := stalwartMailboxContextError(ctx); err != nil {
		return stalwart.JMAPResponse{}, err
	}
	return client.client.CallMailbox(ctx, session, request)
}

func (client *accountBoundStalwartMailboxClient) DownloadBlob(ctx context.Context, blobID string, maxBytes int64) (StalwartMailboxBlob, error) {
	if client == nil || client.client == nil || !validAccountID(client.accountID) || ctx == nil || !validOpaqueID(blobID) || maxBytes < 1 || maxBytes > MaxEngineAttachmentBytes {
		return StalwartMailboxBlob{}, stalwart.ErrInvalidRequest
	}
	if err := stalwartMailboxContextError(ctx); err != nil {
		return StalwartMailboxBlob{}, err
	}
	session, err := client.client.DiscoverMailboxSession(ctx)
	if err != nil {
		return StalwartMailboxBlob{}, err
	}
	if err := stalwartMailboxContextError(ctx); err != nil {
		return StalwartMailboxBlob{}, err
	}
	result, err := client.client.DownloadMailboxBlob(ctx, stalwart.DownloadRequest{
		Session:      session,
		AccountID:    client.accountID,
		BlobID:       blobID,
		Name:         stalwartBlobDownloadName,
		Type:         stalwartBlobDownloadType,
		MaximumBytes: maxBytes,
	})
	if err != nil {
		return StalwartMailboxBlob{}, err
	}
	contentSize := int64(len(result.Content))
	if contentSize > maxBytes || contentSize > MaxEngineAttachmentBytes {
		return StalwartMailboxBlob{}, stalwart.ErrPayloadTooLarge
	}
	mediaType, err := canonicalStalwartBlobMediaType(result.ContentType)
	if err != nil {
		return StalwartMailboxBlob{}, err
	}
	return StalwartMailboxBlob{
		AccountID: client.accountID,
		BlobID:    blobID,
		MediaType: mediaType,
		SizeBytes: contentSize,
		Bytes:     result.Content,
	}, nil
}

func canonicalStalwartMailboxOrigin(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > stalwartMailboxOriginMaxBytes || !utf8.ValidString(value) {
		return "", false
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || parsed.Opaque != "" {
		return "", false
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return "", false
	}
	if parsed.Path != "" && parsed.Path != "/" || parsed.RawPath != "" && parsed.RawPath != "/" || parsed.Hostname() == "" {
		return "", false
	}
	return (&url.URL{Scheme: strings.ToLower(parsed.Scheme), Host: strings.ToLower(parsed.Host)}).String(), true
}

func canonicalStalwartBlobMediaType(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return stalwartBlobDownloadType, nil
	}
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return "", stalwart.ErrProtocol
	}
	normalized, err := normalizeAttachmentMediaType(mediaType)
	if err != nil {
		return "", stalwart.ErrProtocol
	}
	return normalized, nil
}

func stalwartMailboxContextError(ctx context.Context) error {
	if ctx == nil {
		return stalwart.ErrInvalidRequest
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return stalwart.ErrCancelled
		}
		return stalwart.ErrUnavailable
	}
	return nil
}

func safeStalwartAuthorizerSourceError(ctx context.Context, err error) error {
	if contextErr := stalwartMailboxContextError(ctx); contextErr != nil {
		return contextErr
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, stalwart.ErrCancelled) {
		return stalwart.ErrCancelled
	}
	if errors.Is(err, stalwart.ErrUnauthorized) {
		return stalwart.ErrUnauthorized
	}
	return stalwart.ErrUnavailable
}
