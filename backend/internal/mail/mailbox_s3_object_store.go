package mail

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// MailboxS3ObjectStore keeps internal reads and browser-visible presigns on
// separate path-style S3 clients. The public endpoint is used only to sign
// bounded URLs; object bytes always use the internal endpoint.
type MailboxS3ObjectStore struct {
	bucket       string
	publicOrigin *url.URL
	objectClient *s3.Client
	presign      *s3.PresignClient
}

var _ MailboxObjectStore = (*MailboxS3ObjectStore)(nil)

// NewMailboxS3ObjectStore constructs the mail object boundary from the pinned
// S3 SDK. presignOrigin must be the browser-reachable HTTPS origin and is kept
// separate from cfg.Endpoint, which is used only for internal reads.
func NewMailboxS3ObjectStore(ctx context.Context, cfg config.S3Config, presignOrigin string) (*MailboxS3ObjectStore, error) {
	internalOrigin, internalOK := mailboxS3ObjectStoreOrigin(cfg.Endpoint, false)
	publicOrigin, publicOK := mailboxS3ObjectStoreOrigin(presignOrigin, true)
	if ctx == nil || !internalOK || !publicOK || !mailboxS3ObjectStoreConfigValid(cfg) {
		return nil, newMailboxObjectStoreError(MailboxObjectStoreErrorInvalid, nil)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, mailboxS3ObjectStoreProviderError(ctx, err)
	}

	objectClient := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(internalOrigin.String())
		options.UsePathStyle = true
	})
	presignClient := s3.NewPresignClient(s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(publicOrigin.String())
		options.UsePathStyle = true
	}))

	return &MailboxS3ObjectStore{
		bucket:       cfg.Bucket,
		publicOrigin: publicOrigin,
		objectClient: objectClient,
		presign:      presignClient,
	}, nil
}

// PresignPut returns one exact-size PUT contract on the public S3 origin.
func (store *MailboxS3ObjectStore) PresignPut(ctx context.Context, request MailboxObjectPresignPutRequest) (MailboxObjectPresignPutResult, error) {
	if !mailboxS3ObjectStoreReady(store) || ctx == nil || !validMailboxObjectPresignPutRequest(request) {
		return MailboxObjectPresignPutResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorInvalid, nil)
	}
	if request.Object.Bucket != store.bucket {
		return MailboxObjectPresignPutResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorRejected, nil)
	}
	lifetime, ok := mailboxS3ObjectStoreExactLifetime(request.IssuedAt, request.ExpiresAt, mailboxObjectStoreMaximumPutLifetime)
	if !ok {
		return MailboxObjectPresignPutResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorInvalid, nil)
	}

	presigned, err := store.presign.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:        aws.String(request.Object.Bucket),
			Key:           aws.String(request.Object.Key),
			ContentLength: aws.Int64(request.ContentLength),
			ContentType:   aws.String(request.ContentType),
		},
		s3.WithPresignExpires(lifetime),
	)
	if err != nil {
		return MailboxObjectPresignPutResult{}, mailboxS3ObjectStoreProviderError(ctx, err)
	}
	if presigned == nil || !mailboxS3ObjectStorePresignURLValid(store.publicOrigin, presigned.URL, lifetime) {
		return MailboxObjectPresignPutResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorUnavailable, nil)
	}
	headers, ok := mailboxS3ObjectStorePutHeaders(presigned.SignedHeader, request)
	if !ok || !validMailboxObjectPresignPutResult(request, MailboxObjectPresignPutResult{
		URL:       presigned.URL,
		Headers:   headers,
		ExpiresAt: request.ExpiresAt,
	}) {
		return MailboxObjectPresignPutResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorUnavailable, nil)
	}

	return MailboxObjectPresignPutResult{
		URL:       presigned.URL,
		Headers:   headers,
		ExpiresAt: request.ExpiresAt,
	}, nil
}

// Open streams a mail-owned object over the internal S3 endpoint. The caller
// owns the returned body and must close it.
func (store *MailboxS3ObjectStore) Open(ctx context.Context, request MailboxObjectReadRequest) (MailboxObjectReadStream, error) {
	if !mailboxS3ObjectStoreReady(store) || ctx == nil || !validMailboxObjectReadRequest(request) {
		return MailboxObjectReadStream{}, newMailboxObjectStoreError(MailboxObjectStoreErrorInvalid, nil)
	}
	if request.Object.Bucket != store.bucket {
		return MailboxObjectReadStream{}, newMailboxObjectStoreError(MailboxObjectStoreErrorRejected, nil)
	}

	output, err := store.objectClient.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(request.Object.Bucket),
		Key:    aws.String(request.Object.Key),
	})
	if err != nil {
		return MailboxObjectReadStream{}, mailboxS3ObjectStoreProviderError(ctx, err)
	}
	if output == nil || output.Body == nil || output.ContentLength == nil || *output.ContentLength < 1 || *output.ContentLength > MaxEngineAttachmentBytes {
		if output != nil && output.Body != nil {
			_ = output.Body.Close()
		}
		return MailboxObjectReadStream{}, newMailboxObjectStoreError(MailboxObjectStoreErrorUnavailable, nil)
	}

	return MailboxObjectReadStream{Body: output.Body, ContentLength: *output.ContentLength}, nil
}

// PresignGet returns one bounded GET URL on the public S3 origin.
func (store *MailboxS3ObjectStore) PresignGet(ctx context.Context, request MailboxObjectPresignGetRequest) (MailboxObjectPresignGetResult, error) {
	if !mailboxS3ObjectStoreReady(store) || ctx == nil || !validMailboxObjectPresignGetRequest(request) {
		return MailboxObjectPresignGetResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorInvalid, nil)
	}
	if request.Object.Bucket != store.bucket {
		return MailboxObjectPresignGetResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorRejected, nil)
	}
	lifetime, ok := mailboxS3ObjectStoreExactLifetime(request.IssuedAt, request.ExpiresAt, mailboxObjectStoreMaximumGetLifetime)
	if !ok {
		return MailboxObjectPresignGetResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorInvalid, nil)
	}

	presigned, err := store.presign.PresignGetObject(
		ctx,
		&s3.GetObjectInput{Bucket: aws.String(request.Object.Bucket), Key: aws.String(request.Object.Key)},
		s3.WithPresignExpires(lifetime),
	)
	if err != nil {
		return MailboxObjectPresignGetResult{}, mailboxS3ObjectStoreProviderError(ctx, err)
	}
	if presigned == nil || !mailboxS3ObjectStorePresignURLValid(store.publicOrigin, presigned.URL, lifetime) || !validMailboxObjectPresignGetResult(request, MailboxObjectPresignGetResult{URL: presigned.URL, ExpiresAt: request.ExpiresAt}) {
		return MailboxObjectPresignGetResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorUnavailable, nil)
	}

	return MailboxObjectPresignGetResult{URL: presigned.URL, ExpiresAt: request.ExpiresAt}, nil
}

func mailboxS3ObjectStoreReady(store *MailboxS3ObjectStore) bool {
	if store == nil || store.objectClient == nil || store.presign == nil || !mailboxS3ObjectStoreBucketValid(store.bucket) || store.publicOrigin == nil {
		return false
	}
	_, ok := mailboxS3ObjectStoreOrigin(store.publicOrigin.String(), true)
	return ok
}

func mailboxS3ObjectStoreConfigValid(cfg config.S3Config) bool {
	return mailboxS3ObjectStoreBucketValid(cfg.Bucket) &&
		mailboxS3ObjectStoreRegionValid(cfg.Region) &&
		mailboxS3ObjectStoreSecretValid(cfg.AccessKey) &&
		mailboxS3ObjectStoreSecretValid(cfg.SecretKey)
}

func mailboxS3ObjectStoreOrigin(value string, requireHTTPS bool) (*url.URL, bool) {
	if !mailboxS3ObjectStoreTextValid(value, mailboxObjectStoreMaximumURLBytes) || strings.TrimSpace(value) != value {
		return nil, false
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil || !parsed.IsAbs() || parsed.Opaque != "" || parsed.Host == "" || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.RawPath != "" || parsed.Path != "" && parsed.Path != "/" {
		return nil, false
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" || requireHTTPS && scheme != "https" {
		return nil, false
	}
	return &url.URL{Scheme: scheme, Host: strings.ToLower(parsed.Host)}, true
}

func mailboxS3ObjectStoreBucketValid(value string) bool {
	if len(value) < 3 || len(value) > 63 || net.ParseIP(value) != nil {
		return false
	}
	previousSeparator := false
	for index := 0; index < len(value); index++ {
		character := value[index]
		isLowerAlpha := character >= 'a' && character <= 'z'
		isDigit := character >= '0' && character <= '9'
		isSeparator := character == '-' || character == '.'
		if !isLowerAlpha && !isDigit && !isSeparator || index == 0 && isSeparator || index == len(value)-1 && isSeparator || isSeparator && previousSeparator && character == '.' {
			return false
		}
		previousSeparator = isSeparator
	}
	return true
}

func mailboxS3ObjectStoreRegionValid(value string) bool {
	if !mailboxS3ObjectStoreTextValid(value, 128) || strings.TrimSpace(value) != value {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		if character != '-' && (character < 'a' || character > 'z') && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}

func mailboxS3ObjectStoreSecretValid(value string) bool {
	return mailboxS3ObjectStoreTextValid(value, 4_096) && strings.TrimSpace(value) == value
}

func mailboxS3ObjectStoreTextValid(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) && mailboxObjectStoreControlFree(value)
}

func mailboxS3ObjectStoreExactLifetime(issuedAt, expiresAt time.Time, maximum time.Duration) (time.Duration, bool) {
	if !validMailboxObjectStoreLifetime(issuedAt, expiresAt, maximum) {
		return 0, false
	}
	lifetime := expiresAt.Sub(issuedAt)
	return lifetime, lifetime%time.Second == 0
}

func mailboxS3ObjectStorePresignURLValid(origin *url.URL, value string, lifetime time.Duration) bool {
	if origin == nil || !validMailboxObjectStoreURL(value) || lifetime < time.Second || lifetime%time.Second != 0 {
		return false
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil || !strings.EqualFold(parsed.Scheme, origin.Scheme) || !strings.EqualFold(parsed.Host, origin.Host) {
		return false
	}
	expires := parsed.Query()["X-Amz-Expires"]
	return len(expires) == 1 && expires[0] == strconv.FormatInt(int64(lifetime/time.Second), 10)
}

func mailboxS3ObjectStorePutHeaders(headers http.Header, request MailboxObjectPresignPutRequest) (map[string]string, bool) {
	if headers == nil || len(headers) == 0 || len(headers) > mailboxObjectStoreMaximumHeaders {
		return nil, false
	}
	result := make(map[string]string, len(headers))
	contentLength := strconv.FormatInt(request.ContentLength, 10)
	hasContentLength := false
	for name, values := range headers {
		canonicalName := http.CanonicalHeaderKey(name)
		if canonicalName == "" || !validMailboxObjectStoreHeaderName(canonicalName) || len(values) != 1 || !validMailboxObjectStoreText(values[0], mailboxObjectStoreMaximumHeaderValue, false) {
			return nil, false
		}
		if _, duplicate := result[canonicalName]; duplicate {
			return nil, false
		}
		if canonicalName == "Content-Length" {
			if values[0] != contentLength {
				return nil, false
			}
			hasContentLength = true
		}
		if canonicalName == "Content-Type" && values[0] != request.ContentType {
			return nil, false
		}
		result[canonicalName] = values[0]
	}
	return result, hasContentLength
}

func mailboxS3ObjectStoreProviderError(ctx context.Context, err error) error {
	if err == nil {
		return newMailboxObjectStoreError(MailboxObjectStoreErrorUnavailable, nil)
	}
	if ctx != nil && ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return newMailboxObjectStoreError(MailboxObjectStoreErrorUnavailable, err)
	}
	var noSuchKey *s3types.NoSuchKey
	if errors.As(err, &noSuchKey) && noSuchKey != nil {
		return newMailboxObjectStoreError(MailboxObjectStoreErrorNotFound, err)
	}
	var apiError smithy.APIError
	if errors.As(err, &apiError) && apiError != nil {
		switch apiError.ErrorCode() {
		case "NoSuchKey", "NoSuchBucket", "NotFound", "NoSuchObject":
			return newMailboxObjectStoreError(MailboxObjectStoreErrorNotFound, err)
		case "OperationAborted", "ConditionalRequestConflict", "PreconditionFailed":
			return newMailboxObjectStoreError(MailboxObjectStoreErrorConflict, err)
		case "AccessDenied", "InvalidAccessKeyId", "SignatureDoesNotMatch", "AuthorizationHeaderMalformed", "InvalidToken", "ExpiredToken", "RequestExpired":
			return newMailboxObjectStoreError(MailboxObjectStoreErrorRejected, err)
		case "InvalidRequest", "InvalidArgument", "InvalidBucketName", "InvalidObjectState", "MalformedXML", "MalformedACLError":
			return newMailboxObjectStoreError(MailboxObjectStoreErrorInvalid, err)
		}
	}
	var responseError *smithyhttp.ResponseError
	if errors.As(err, &responseError) && responseError != nil && responseError.HTTPResponse() != nil {
		switch responseError.HTTPStatusCode() {
		case http.StatusNotFound, http.StatusGone:
			return newMailboxObjectStoreError(MailboxObjectStoreErrorNotFound, err)
		case http.StatusConflict, http.StatusPreconditionFailed:
			return newMailboxObjectStoreError(MailboxObjectStoreErrorConflict, err)
		case http.StatusBadRequest, http.StatusUnprocessableEntity:
			return newMailboxObjectStoreError(MailboxObjectStoreErrorInvalid, err)
		case http.StatusUnauthorized, http.StatusForbidden:
			return newMailboxObjectStoreError(MailboxObjectStoreErrorRejected, err)
		}
	}
	return newMailboxObjectStoreError(MailboxObjectStoreErrorUnavailable, err)
}
