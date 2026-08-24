package mail

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	mailboxObjectStoreMaximumPutLifetime = 10 * time.Minute
	mailboxObjectStoreMaximumGetLifetime = 15 * time.Minute
	mailboxObjectStoreMaximumURLBytes    = 8_192
	mailboxObjectStoreMaximumHeaders     = 64
	mailboxObjectStoreMaximumHeaderName  = 256
	mailboxObjectStoreMaximumHeaderValue = 4_096
)

// MailboxObjectStore is the streaming boundary for one mail object.
type MailboxObjectStore interface {
	PresignPut(context.Context, MailboxObjectPresignPutRequest) (MailboxObjectPresignPutResult, error)
	Open(context.Context, MailboxObjectReadRequest) (MailboxObjectReadStream, error)
	PresignGet(context.Context, MailboxObjectPresignGetRequest) (MailboxObjectPresignGetResult, error)
}

// MailboxObjectLocator identifies one mail-owned object.
type MailboxObjectLocator struct {
	Bucket string
	Key    string
}

// MailboxObjectPresignPutRequest binds a pending object to one bounded upload.
type MailboxObjectPresignPutRequest struct {
	Object        MailboxObjectLocator
	ContentLength int64
	ContentType   string
	IssuedAt      time.Time
	ExpiresAt     time.Time
}

// MailboxObjectPresignPutResult returns the bounded upload URL and its headers.
type MailboxObjectPresignPutResult struct {
	URL       string
	Headers   map[string]string
	ExpiresAt time.Time
}

// MailboxObjectReadRequest identifies an object opened for streaming.
type MailboxObjectReadRequest struct {
	Object MailboxObjectLocator
}

// MailboxObjectReadStream is one caller-owned object stream.
type MailboxObjectReadStream struct {
	Body          io.ReadCloser
	ContentLength int64
}

// MailboxObjectPresignGetRequest binds a download URL to one bounded lifetime.
type MailboxObjectPresignGetRequest struct {
	Object    MailboxObjectLocator
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// MailboxObjectPresignGetResult returns one bounded download URL.
type MailboxObjectPresignGetResult struct {
	URL       string
	ExpiresAt time.Time
}

// MailboxObjectStoreErrorKind is the closed safe failure kind of MailboxObjectStore.
type MailboxObjectStoreErrorKind string

const (
	MailboxObjectStoreErrorInvalid     MailboxObjectStoreErrorKind = "invalid"
	MailboxObjectStoreErrorNotFound    MailboxObjectStoreErrorKind = "not_found"
	MailboxObjectStoreErrorConflict    MailboxObjectStoreErrorKind = "conflict"
	MailboxObjectStoreErrorRejected    MailboxObjectStoreErrorKind = "rejected"
	MailboxObjectStoreErrorUnavailable MailboxObjectStoreErrorKind = "unavailable"
)

// MailboxObjectStoreError retains only a safe kind and an internal cause.
type MailboxObjectStoreError struct {
	kind  MailboxObjectStoreErrorKind
	cause error
}

// Error returns a stable content-free message.
func (*MailboxObjectStoreError) Error() string {
	return "mailbox object store error"
}

// Unwrap retains the internal cause for adapter-local classification.
func (e *MailboxObjectStoreError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Kind returns a closed safe kind and fails closed for malformed errors.
func (e *MailboxObjectStoreError) Kind() MailboxObjectStoreErrorKind {
	if e == nil || !validMailboxObjectStoreErrorKind(e.kind) {
		return MailboxObjectStoreErrorInvalid
	}
	return e.kind
}

// MailboxObjectStoreErrorKindOf reports a safe kind without exposing its cause.
func MailboxObjectStoreErrorKindOf(err error) (MailboxObjectStoreErrorKind, bool) {
	var target *MailboxObjectStoreError
	if !errors.As(err, &target) || target == nil {
		return "", false
	}
	return target.Kind(), true
}

func newMailboxObjectStoreError(kind MailboxObjectStoreErrorKind, cause error) *MailboxObjectStoreError {
	if !validMailboxObjectStoreErrorKind(kind) {
		kind = MailboxObjectStoreErrorInvalid
	}
	return &MailboxObjectStoreError{kind: kind, cause: cause}
}

func validMailboxObjectPresignPutRequest(request MailboxObjectPresignPutRequest) bool {
	return validMailboxObjectLocator(request.Object) &&
		request.ContentLength >= 1 && request.ContentLength <= MaxEngineAttachmentBytes &&
		validMailboxObjectStoreContentType(request.ContentType) &&
		validMailboxObjectStoreLifetime(request.IssuedAt, request.ExpiresAt, mailboxObjectStoreMaximumPutLifetime)
}

func validMailboxObjectPresignPutResult(request MailboxObjectPresignPutRequest, result MailboxObjectPresignPutResult) bool {
	return validMailboxObjectStoreURL(result.URL) &&
		validMailboxObjectStoreHeaders(result.Headers) &&
		validMailboxObjectStoreResultExpiry(request.IssuedAt, request.ExpiresAt, result.ExpiresAt)
}

func cloneMailboxObjectPresignPutResult(result MailboxObjectPresignPutResult) MailboxObjectPresignPutResult {
	result.Headers = cloneMailboxObjectStoreHeaders(result.Headers)
	return result
}

func validMailboxObjectReadRequest(request MailboxObjectReadRequest) bool {
	return validMailboxObjectLocator(request.Object)
}

func validMailboxObjectReadStream(stream MailboxObjectReadStream) bool {
	return stream.Body != nil && stream.ContentLength >= 1 && stream.ContentLength <= MaxEngineAttachmentBytes
}

func validMailboxObjectPresignGetRequest(request MailboxObjectPresignGetRequest) bool {
	return validMailboxObjectLocator(request.Object) &&
		validMailboxObjectStoreLifetime(request.IssuedAt, request.ExpiresAt, mailboxObjectStoreMaximumGetLifetime)
}

func validMailboxObjectPresignGetResult(request MailboxObjectPresignGetRequest, result MailboxObjectPresignGetResult) bool {
	return validMailboxObjectStoreURL(result.URL) &&
		validMailboxObjectStoreResultExpiry(request.IssuedAt, request.ExpiresAt, result.ExpiresAt)
}

func validMailboxObjectLocator(locator MailboxObjectLocator) bool {
	return validMailObjectDeletionIdentity(locator.Bucket, locator.Key) &&
		mailboxObjectStoreControlFree(locator.Bucket) && mailboxObjectStoreControlFree(locator.Key)
}

func validMailboxObjectStoreContentType(value string) bool {
	if !validMailboxObjectStoreText(value, MaxEngineMIMEFieldBytes, true) {
		return false
	}
	mediaType, parameters, err := mime.ParseMediaType(value)
	if err != nil || !validMailboxObjectStoreMediaType(mediaType) {
		return false
	}
	return mime.FormatMediaType(mediaType, parameters) == value
}

func validMailboxObjectStoreLifetime(issuedAt, expiresAt time.Time, maximum time.Duration) bool {
	if !validMailboxObjectStoreTime(issuedAt) || !validMailboxObjectStoreTime(expiresAt) {
		return false
	}
	lifetime := expiresAt.Sub(issuedAt)
	return lifetime > 0 && lifetime <= maximum
}

func validMailboxObjectStoreResultExpiry(requestIssuedAt, requestExpiry, resultExpiry time.Time) bool {
	return validMailboxObjectStoreTime(resultExpiry) && resultExpiry.After(requestIssuedAt) && !resultExpiry.After(requestExpiry)
}

func validMailboxObjectStoreTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC
}

func validMailboxObjectStoreURL(value string) bool {
	if !validMailboxObjectStoreText(value, mailboxObjectStoreMaximumURLBytes, true) {
		return false
	}
	parsed, err := url.ParseRequestURI(value)
	return err == nil && parsed != nil && parsed.IsAbs() && strings.EqualFold(parsed.Scheme, "https") &&
		parsed.Host != "" && parsed.Hostname() != "" && parsed.User == nil && parsed.Fragment == "" && parsed.Opaque == ""
}

func validMailboxObjectStoreHeaders(headers map[string]string) bool {
	if len(headers) > mailboxObjectStoreMaximumHeaders {
		return false
	}
	for name, value := range headers {
		if !validMailboxObjectStoreHeaderName(name) ||
			!validMailboxObjectStoreText(value, mailboxObjectStoreMaximumHeaderValue, false) {
			return false
		}
	}
	return true
}

func cloneMailboxObjectStoreHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	result := make(map[string]string, len(headers))
	for name, value := range headers {
		result[name] = value
	}
	return result
}

func validMailboxObjectStoreErrorKind(kind MailboxObjectStoreErrorKind) bool {
	switch kind {
	case MailboxObjectStoreErrorInvalid,
		MailboxObjectStoreErrorNotFound,
		MailboxObjectStoreErrorConflict,
		MailboxObjectStoreErrorRejected,
		MailboxObjectStoreErrorUnavailable:
		return true
	default:
		return false
	}
}

func validMailboxObjectStoreMediaType(value string) bool {
	typeName, subtype, found := strings.Cut(value, "/")
	return found && typeName != "" && subtype != "" && !strings.Contains(subtype, "/")
}

func validMailboxObjectStoreHeaderName(value string) bool {
	if !validMailboxObjectStoreText(value, mailboxObjectStoreMaximumHeaderName, true) {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' {
			continue
		}
		switch character {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			continue
		default:
			return false
		}
	}
	return true
}

func validMailboxObjectStoreText(value string, maximum int, required bool) bool {
	return (!required || value != "") && len(value) <= maximum && utf8.ValidString(value) && mailboxObjectStoreControlFree(value)
}

func mailboxObjectStoreControlFree(value string) bool {
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
