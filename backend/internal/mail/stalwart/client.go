package stalwart

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	sessionDeadline  = 10 * time.Second
	jmapDeadline     = 15 * time.Second
	transferDeadline = 60 * time.Second
)

// ContextReadCloser is the upload pre-read boundary. ReadContext must return
// when ctx is cancelled; the transport does not start helper goroutines to
// interrupt an opaque io.Reader. UploadMailboxBlob owns and closes it after
// the bounded pre-read, including every cancellation path.
type ContextReadCloser interface {
	io.ReadCloser
	ReadContext(context.Context, []byte) (int, error)
}

// UploadRequest makes the caller select a bounded body size and provide a
// cancellation-aware stream. The transport additionally caps it at 50 MiB and
// never records the content after the call.
type UploadRequest struct {
	Session      *MailboxSession
	AccountID    string
	ContentType  string
	Content      ContextReadCloser
	MaximumBytes int64
}

// UploadResult contains only the provider-private blob identifiers needed by
// later adapter mapping. It has no String method and must not be logged.
type UploadResult struct {
	AccountID string
	BlobID    string
	Type      string
	Size      int64
}

// DownloadRequest makes the caller select a bounded blob size. The hard local
// limit is 25 MiB, independent of the generic mail package.
type DownloadRequest struct {
	Session      *MailboxSession
	AccountID    string
	BlobID       string
	Name         string
	Type         string
	MaximumBytes int64
}

// DownloadResult owns the bounded content for the lifetime of the caller. It
// deliberately contains no source URL or authorization state.
type DownloadResult struct {
	Content     []byte
	ContentType string
}

// DiscoverMailboxSession performs the only discovery flow accepted for the
// pinned Stalwart version: unauthenticated well-known GET, an exact same-origin
// 307 to /jmap/session, then mailbox-authorized session GET. Redirect following
// is disabled for every request; no other redirect is accepted.
func (c *Client) DiscoverMailboxSession(ctx context.Context) (*MailboxSession, error) {
	wellKnown := c.fixedURL("/.well-known/jmap")
	status, headers, _, err := c.doRequest(ctx, sessionDeadline, http.MethodGet, wellKnown, nil, "", nil, maxSessionBytes)
	if err != nil {
		return nil, err
	}
	if status != http.StatusTemporaryRedirect {
		return nil, errorFromStatus(status)
	}
	if !c.validSessionRedirect(headers.Get("Location")) {
		return nil, newError(ErrorProtocol)
	}

	sessionURL := c.fixedURL("/jmap/session")
	status, _, payload, err := c.doRequest(ctx, sessionDeadline, http.MethodGet, sessionURL, nil, "", c.authorizeMailbox, maxSessionBytes)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, errorFromStatus(status)
	}
	return parseMailboxSession(c, payload)
}

// CallRegistry sends an allowlisted Stalwart Registry batch to exact POST
// /jmap. Registry methods and credentials can never cross into mailbox calls.
func (c *Client) CallRegistry(ctx context.Context, request JMAPRequest) (JMAPResponse, error) {
	payload, mutation, err := validateRequest(jmapRegistry, request, nil)
	if err != nil {
		return JMAPResponse{}, err
	}
	endpoint := c.fixedURL("/jmap")
	status, _, response, err := c.doRequest(ctx, jmapDeadline, http.MethodPost, endpoint, bytes.NewReader(payload), "application/json", c.authorizeRegistry, maxJMAPBytes)
	if err != nil {
		return JMAPResponse{}, mutationError(err, mutation)
	}
	if status != http.StatusOK {
		return JMAPResponse{}, mutationError(errorFromStatus(status), mutation)
	}
	parsed, err := parseJMAPResponse(jmapRegistry, request, response, nil)
	if err != nil {
		return JMAPResponse{}, mutationError(err, mutation)
	}
	return parsed, nil
}

// CallMailbox sends a standards-only batch to the exact session-confirmed
// mailbox endpoint /jmap/. A transport ambiguity after a mutation is always
// reported as indeterminate; this method performs no retry.
func (c *Client) CallMailbox(ctx context.Context, session *MailboxSession, request JMAPRequest) (JMAPResponse, error) {
	payload, mutation, err := validateRequest(jmapMailbox, request, session)
	if err != nil {
		return JMAPResponse{}, err
	}
	if !c.validMailboxAPIURL(session) {
		return JMAPResponse{}, newError(ErrorProtocol)
	}
	status, _, response, err := c.doRequest(ctx, jmapDeadline, http.MethodPost, session.apiURL, bytes.NewReader(payload), "application/json", c.authorizeMailbox, maxJMAPBytes)
	if err != nil {
		return JMAPResponse{}, mutationError(err, mutation)
	}
	if status != http.StatusOK {
		return JMAPResponse{}, mutationError(errorFromStatus(status), mutation)
	}
	parsed, err := parseJMAPResponse(jmapMailbox, request, response, session)
	if err != nil {
		return JMAPResponse{}, mutationError(err, mutation)
	}
	return parsed, nil
}

// UploadMailboxBlob posts a bounded body to the exact session upload template.
// It pre-reads at most the permitted bytes so an oversize caller stream causes
// no partial provider write.
func (c *Client) UploadMailboxBlob(ctx context.Context, request UploadRequest) (UploadResult, error) {
	maximum, ok := boundedLimit(request.MaximumBytes, maxUploadContentBytes)
	if ctx == nil || !ok || request.Session == nil || !validOpaque(request.AccountID) || !validContentType(request.ContentType) || request.Content == nil || !c.validUploadURL(request.Session) {
		return UploadResult{}, newError(ErrorInvalidRequest)
	}
	defer request.Content.Close()
	content, err := readBounded(ctx, request.Content, maximum)
	if err != nil {
		return UploadResult{}, err
	}
	if ctx.Err() != nil {
		return UploadResult{}, requestError(ctx)
	}
	target, ok := c.renderUploadURL(request.AccountID)
	if !ok {
		return UploadResult{}, newError(ErrorProtocol)
	}
	status, _, payload, err := c.doRequest(ctx, transferDeadline, http.MethodPost, target, bytes.NewReader(content), request.ContentType, c.authorizeMailbox, maxJMAPBytes)
	if err != nil {
		return UploadResult{}, newError(ErrorIndeterminate)
	}
	if status != http.StatusCreated {
		return UploadResult{}, mutationError(errorFromStatus(status), true)
	}
	return parseUploadResult(request.AccountID, payload)
}

// DownloadMailboxBlob reads a bounded blob through the exact session download
// template. It never follows redirects or returns a provider URL.
func (c *Client) DownloadMailboxBlob(ctx context.Context, request DownloadRequest) (DownloadResult, error) {
	maximum, ok := boundedLimit(request.MaximumBytes, maxDownloadedBlobBytes)
	if !ok || request.Session == nil || !validOpaque(request.AccountID) || !validOpaque(request.BlobID) || !validDownloadName(request.Name) || !validContentType(request.Type) || !c.validDownloadURL(request.Session) {
		return DownloadResult{}, newError(ErrorInvalidRequest)
	}
	target, ok := c.renderDownloadURL(request.AccountID, request.BlobID, request.Name, request.Type)
	if !ok {
		return DownloadResult{}, newError(ErrorProtocol)
	}
	status, headers, content, err := c.doRequest(ctx, transferDeadline, http.MethodGet, target, nil, "", c.authorizeMailbox, maximum)
	if err != nil {
		return DownloadResult{}, err
	}
	if status != http.StatusOK {
		return DownloadResult{}, errorFromStatus(status)
	}
	contentType := headers.Get("Content-Type")
	if contentType != "" && !validContentType(contentType) {
		return DownloadResult{}, newError(ErrorProtocol)
	}
	return DownloadResult{Content: content, ContentType: contentType}, nil
}

func (c *Client) authorizeRegistry(ctx context.Context, request *http.Request) error {
	if c == nil || c.registryAuthorizer == nil || request == nil {
		return newError(ErrorUnauthorized)
	}
	if err := c.registryAuthorizer.AuthorizeRegistry(ctx, request); err != nil {
		return newError(ErrorUnauthorized)
	}
	return nil
}

func (c *Client) authorizeMailbox(ctx context.Context, request *http.Request) error {
	if c == nil || c.mailboxAuthorizer == nil || request == nil {
		return newError(ErrorUnauthorized)
	}
	if err := c.mailboxAuthorizer.AuthorizeMailbox(ctx, request); err != nil {
		return newError(ErrorUnauthorized)
	}
	return nil
}

type requestAuthorizer func(context.Context, *http.Request) error

func (c *Client) doRequest(ctx context.Context, deadline time.Duration, method string, target *url.URL, body io.Reader, contentType string, authorize requestAuthorizer, maximum int64) (int, http.Header, []byte, error) {
	if c == nil || c.httpClient == nil || target == nil || !c.sameOrigin(target) || target.User != nil || target.Fragment != "" || target.ForceQuery || ctx == nil || maximum < 1 {
		return 0, nil, nil, newError(ErrorInvalidRequest)
	}
	requestContext, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, method, target.String(), body)
	if err != nil {
		return 0, nil, nil, newError(ErrorInvalidRequest)
	}
	request.Header.Set("Accept", "application/json")
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if authorize != nil {
		if err := authorize(requestContext, request); err != nil {
			return 0, nil, nil, err
		}
	}
	if !sameExactURL(request.URL, target) {
		return 0, nil, nil, newError(ErrorProtocol)
	}

	client := *c.httpClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	response, err := client.Do(request)
	if err != nil {
		return 0, nil, nil, requestError(requestContext)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1024))
		return response.StatusCode, response.Header.Clone(), nil, nil
	}
	payload, err := readHTTPBody(response.Body, maximum)
	if err != nil {
		return response.StatusCode, response.Header.Clone(), nil, err
	}
	return response.StatusCode, response.Header.Clone(), payload, nil
}

func sameExactURL(left, right *url.URL) bool {
	return left != nil && right != nil && left.String() == right.String() && left.User == nil && right.User == nil && left.Fragment == "" && right.Fragment == "" && !left.ForceQuery && !right.ForceQuery
}

func (c *Client) validSessionRedirect(location string) bool {
	if c == nil || c.origin == nil || location == "" || len(location) > maxOriginBytes || !utf8Valid(location) {
		return false
	}
	parsed, err := url.Parse(location)
	if err != nil || parsed == nil || parsed.User != nil || parsed.Fragment != "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Opaque != "" || parsed.RawPath != "" {
		return false
	}
	if !parsed.IsAbs() {
		if parsed.Host != "" || parsed.Path != "/jmap/session" {
			return false
		}
		parsed = c.origin.ResolveReference(parsed)
	}
	return c.sameOrigin(parsed) && parsed.Path == "/jmap/session" && parsed.RawPath == "" && parsed.RawQuery == "" && parsed.Fragment == "" && parsed.User == nil
}

func (c *Client) validMailboxAPIURL(session *MailboxSession) bool {
	return session != nil && session.apiURL != nil && c.sameOrigin(session.apiURL) && session.apiURL.Path == "/jmap/" && session.apiURL.RawPath == "" && session.apiURL.RawQuery == "" && session.apiURL.Fragment == "" && session.apiURL.User == nil
}

func (c *Client) validUploadURL(session *MailboxSession) bool {
	return session != nil && session.uploadURL != nil && c.sameOrigin(session.uploadURL) && session.uploadURL.Path == "/jmap/upload/{accountId}/" && session.uploadURL.RawPath == "" && session.uploadURL.RawQuery == "" && session.uploadURL.Fragment == "" && session.uploadURL.User == nil
}

func (c *Client) validDownloadURL(session *MailboxSession) bool {
	return session != nil && session.downloadURL != nil && c.sameOrigin(session.downloadURL) && session.downloadURL.Path == "/jmap/download/{accountId}/{blobId}/{name}" && session.downloadURL.RawPath == "" && session.downloadURL.RawQuery == "type={type}" && session.downloadURL.Fragment == "" && session.downloadURL.User == nil
}

func (c *Client) renderUploadURL(accountID string) (*url.URL, bool) {
	if c == nil || !validOpaque(accountID) {
		return nil, false
	}
	return url.Parse(c.origin.String() + "/jmap/upload/" + url.PathEscape(accountID) + "/")
}

func (c *Client) renderDownloadURL(accountID, blobID, name, contentType string) (*url.URL, bool) {
	if c == nil || !validOpaque(accountID) || !validOpaque(blobID) || !validDownloadName(name) || !validContentType(contentType) {
		return nil, false
	}
	return url.Parse(c.origin.String() + "/jmap/download/" + url.PathEscape(accountID) + "/" + url.PathEscape(blobID) + "/" + url.PathEscape(name) + "?type=" + url.QueryEscape(contentType))
}

func boundedLimit(value, hardLimit int64) (int64, bool) {
	return value, value > 0 && value <= hardLimit
}

func readBounded(ctx context.Context, source ContextReadCloser, maximum int64) ([]byte, error) {
	if ctx == nil || source == nil || maximum < 1 {
		return nil, newError(ErrorInvalidRequest)
	}
	const chunkSize = 32 << 10
	buffer := make([]byte, chunkSize)
	content := bytes.NewBuffer(make([]byte, 0, chunkSize))
	for {
		if ctx.Err() != nil {
			return nil, requestError(ctx)
		}
		remaining := maximum + 1 - int64(content.Len())
		if remaining < 1 {
			return nil, newError(ErrorPayloadTooLarge)
		}
		readSize := int64(len(buffer))
		if remaining < readSize {
			readSize = remaining
		}
		n, err := source.ReadContext(ctx, buffer[:readSize])
		if n < 0 || n > int(readSize) {
			return nil, newError(ErrorUnavailable)
		}
		if n > 0 {
			_, _ = content.Write(buffer[:n])
			if int64(content.Len()) > maximum {
				return nil, newError(ErrorPayloadTooLarge)
			}
		}
		if ctx.Err() != nil {
			return nil, requestError(ctx)
		}
		if err == io.EOF {
			return content.Bytes(), nil
		}
		if err != nil || n == 0 {
			return nil, newError(ErrorUnavailable)
		}
	}
}

func readHTTPBody(source io.Reader, maximum int64) ([]byte, error) {
	if source == nil || maximum < 1 {
		return nil, newError(ErrorProtocol)
	}
	payload, err := io.ReadAll(io.LimitReader(source, maximum+1))
	if err != nil {
		return nil, newError(ErrorUnavailable)
	}
	if int64(len(payload)) > maximum {
		return nil, newError(ErrorPayloadTooLarge)
	}
	return payload, nil
}

func parseUploadResult(accountID string, data []byte) (UploadResult, error) {
	if int64(len(data)) > maxJMAPBytes || validateJSON(data) != nil {
		return UploadResult{}, newError(ErrorIndeterminate)
	}
	var result UploadResult
	if err := json.Unmarshal(data, &result); err != nil || result.AccountID != accountID || !validOpaque(result.AccountID) || !validOpaque(result.BlobID) || !validContentType(result.Type) || result.Size < 0 || result.Size > maxUploadContentBytes {
		return UploadResult{}, newError(ErrorIndeterminate)
	}
	return result, nil
}

func validDownloadName(value string) bool {
	if value == "" || len(value) > 255 || !utf8Valid(value) {
		return false
	}
	for _, character := range value {
		if character <= 0x1f || character == 0x7f || character == '/' || character == '\\' {
			return false
		}
	}
	return true
}

func validContentType(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxContentTypeBytes || !utf8Valid(value) || strings.ContainsAny(value, "\r\n") {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(value)
	return err == nil && strings.Contains(mediaType, "/")
}

func requestError(ctx context.Context) error {
	if ctx != nil && ctx.Err() != nil {
		if ctx.Err() == context.Canceled {
			return newError(ErrorCancelled)
		}
		return newError(ErrorUnavailable)
	}
	return newError(ErrorUnavailable)
}

func errorFromStatus(status int) error {
	if status >= http.StatusMultipleChoices && status < http.StatusBadRequest {
		return newError(ErrorProtocol)
	}
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return newError(ErrorUnauthorized)
	case http.StatusNotFound:
		return newError(ErrorNotFound)
	case http.StatusConflict, http.StatusPreconditionFailed:
		return newError(ErrorStateMismatch)
	case http.StatusRequestEntityTooLarge, http.StatusRequestHeaderFieldsTooLarge:
		return newError(ErrorPayloadTooLarge)
	case http.StatusRequestTimeout, http.StatusTooManyRequests:
		return newError(ErrorUnavailable)
	}
	if status >= http.StatusInternalServerError {
		return newError(ErrorUnavailable)
	}
	return newError(ErrorInvalidRequest)
}

func mutationError(err error, mutation bool) error {
	if !mutation || err == nil {
		return err
	}
	kind, known := KindOf(err)
	if !known {
		return newError(ErrorIndeterminate)
	}
	switch kind {
	case ErrorUnavailable, ErrorProtocol, ErrorCancelled:
		return newError(ErrorIndeterminate)
	default:
		return err
	}
}
