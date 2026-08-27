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
	lifetime, ok := mailboxS3ObjectStoreRemainingLifetime(request.ExpiresAt, mailboxObjectStoreMaximumPutLifetime)
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
	if presigned == nil {
		return MailboxObjectPresignPutResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorUnavailable, nil)
	}
	proof, ok := mailboxS3ObjectStoreValidatePresign(
		store.publicOrigin,
		request.Object,
		presigned.URL,
		presigned.Method,
		http.MethodPut,
		lifetime,
		mailboxObjectStoreMaximumPutLifetime,
		request.ExpiresAt,
	)
	if !ok {
		return MailboxObjectPresignPutResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorUnavailable, nil)
	}
	headers, ok := mailboxS3ObjectStorePutHeaders(presigned.SignedHeader, request, proof)
	if !ok || !validMailboxObjectPresignPutResult(request, MailboxObjectPresignPutResult{
		URL:       presigned.URL,
		Headers:   headers,
		ExpiresAt: proof.expiresAt,
	}) {
		return MailboxObjectPresignPutResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorUnavailable, nil)
	}

	return MailboxObjectPresignPutResult{
		URL:       presigned.URL,
		Headers:   headers,
		ExpiresAt: proof.expiresAt,
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
	lifetime, ok := mailboxS3ObjectStoreRemainingLifetime(request.ExpiresAt, mailboxObjectStoreMaximumGetLifetime)
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
	if presigned == nil {
		return MailboxObjectPresignGetResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorUnavailable, nil)
	}
	proof, ok := mailboxS3ObjectStoreValidatePresign(
		store.publicOrigin,
		request.Object,
		presigned.URL,
		presigned.Method,
		http.MethodGet,
		lifetime,
		mailboxObjectStoreMaximumGetLifetime,
		request.ExpiresAt,
	)
	if !ok || !validMailboxObjectPresignGetResult(request, MailboxObjectPresignGetResult{URL: presigned.URL, ExpiresAt: proof.expiresAt}) {
		return MailboxObjectPresignGetResult{}, newMailboxObjectStoreError(MailboxObjectStoreErrorUnavailable, nil)
	}

	return MailboxObjectPresignGetResult{URL: presigned.URL, ExpiresAt: proof.expiresAt}, nil
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

func mailboxS3ObjectStoreRemainingLifetime(expiresAt time.Time, maximum time.Duration) (time.Duration, bool) {
	if !validMailboxObjectStoreTime(expiresAt) || maximum < time.Second || maximum%time.Second != 0 {
		return 0, false
	}
	remaining := expiresAt.Sub(time.Now().UTC()).Truncate(time.Second)
	if remaining < time.Second || remaining > maximum {
		return 0, false
	}
	return remaining, true
}

type mailboxS3ObjectStorePresignProof struct {
	expiresAt     time.Time
	signedHeaders map[string]struct{}
	authority     string
}

func mailboxS3ObjectStoreValidatePresign(origin *url.URL, object MailboxObjectLocator, value, method, expectedMethod string, lifetime, maximum time.Duration, requestExpiresAt time.Time) (mailboxS3ObjectStorePresignProof, bool) {
	if origin == nil || !validMailboxObjectStoreURL(value) || !validMailboxObjectStoreTime(requestExpiresAt) ||
		method != expectedMethod || lifetime < time.Second || lifetime > maximum || lifetime%time.Second != 0 {
		return mailboxS3ObjectStorePresignProof{}, false
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil || !strings.EqualFold(parsed.Scheme, origin.Scheme) || !strings.EqualFold(parsed.Host, origin.Host) ||
		!mailboxS3ObjectStorePresignPathValid(parsed, object) {
		return mailboxS3ObjectStorePresignProof{}, false
	}
	authority, ok := mailboxS3ObjectStoreCanonicalAuthority(parsed)
	if !ok {
		return mailboxS3ObjectStorePresignProof{}, false
	}
	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return mailboxS3ObjectStorePresignProof{}, false
	}
	signature, ok := mailboxS3ObjectStoreSigV4Query(query)
	if !ok || signature.algorithm != `AWS4-HMAC-SHA256` ||
		signature.expires != int64(lifetime/time.Second) || signature.expires > int64(maximum/time.Second) {
		return mailboxS3ObjectStorePresignProof{}, false
	}
	expiresAt := signature.issuedAt.Add(time.Duration(signature.expires) * time.Second)
	if !expiresAt.After(signature.issuedAt) || expiresAt.After(requestExpiresAt) {
		return mailboxS3ObjectStorePresignProof{}, false
	}
	return mailboxS3ObjectStorePresignProof{expiresAt: expiresAt, signedHeaders: signature.signedHeaders, authority: authority}, true
}

type mailboxS3ObjectStoreSigV4Signature struct {
	algorithm     string
	issuedAt      time.Time
	expires       int64
	signedHeaders map[string]struct{}
}

func mailboxS3ObjectStoreSigV4Query(query url.Values) (mailboxS3ObjectStoreSigV4Signature, bool) {
	values := make(map[string]string)
	for name, entries := range query {
		lowerName := strings.ToLower(name)
		if !strings.HasPrefix(lowerName, `x-amz-`) {
			continue
		}
		if _, exists := values[lowerName]; exists || len(entries) != 1 || !validMailboxObjectStoreText(entries[0], mailboxObjectStoreMaximumURLBytes, true) {
			return mailboxS3ObjectStoreSigV4Signature{}, false
		}
		values[lowerName] = entries[0]
	}
	algorithm, algorithmOK := values[`x-amz-algorithm`]
	credential, credentialOK := values[`x-amz-credential`]
	date, dateOK := values[`x-amz-date`]
	expires, expiresOK := values[`x-amz-expires`]
	signedHeaders, signedHeadersOK := values[`x-amz-signedheaders`]
	signature, signatureOK := values[`x-amz-signature`]
	if !algorithmOK || !credentialOK || !dateOK || !expiresOK || !signedHeadersOK || !signatureOK || credential == `` || signature == `` {
		return mailboxS3ObjectStoreSigV4Signature{}, false
	}
	issuedAt, err := time.Parse(`20060102T150405Z`, date)
	if err != nil || issuedAt.Location() != time.UTC {
		return mailboxS3ObjectStoreSigV4Signature{}, false
	}
	seconds, err := strconv.ParseInt(expires, 10, 64)
	if err != nil || seconds < 1 || strconv.FormatInt(seconds, 10) != expires {
		return mailboxS3ObjectStoreSigV4Signature{}, false
	}
	headers, ok := mailboxS3ObjectStoreSignedHeaders(signedHeaders)
	if !ok {
		return mailboxS3ObjectStoreSigV4Signature{}, false
	}
	return mailboxS3ObjectStoreSigV4Signature{algorithm: algorithm, issuedAt: issuedAt, expires: seconds, signedHeaders: headers}, true
}

func mailboxS3ObjectStoreSignedHeaders(value string) (map[string]struct{}, bool) {
	if !validMailboxObjectStoreText(value, mailboxObjectStoreMaximumHeaderValue, true) {
		return nil, false
	}
	result := make(map[string]struct{})
	for _, name := range strings.Split(value, `;`) {
		if name == `` || name != strings.ToLower(name) || !validMailboxObjectStoreHeaderName(name) {
			return nil, false
		}
		if _, duplicate := result[name]; duplicate {
			return nil, false
		}
		result[name] = struct{}{}
	}
	_, hasHost := result[`host`]
	return result, hasHost
}
func mailboxS3ObjectStorePresignPathValid(parsed *url.URL, object MailboxObjectLocator) bool {
	if parsed == nil || !validMailboxObjectLocator(object) {
		return false
	}
	expected := (&url.URL{Path: `/` + object.Bucket + `/` + object.Key}).EscapedPath()
	return expected != `` && parsed.EscapedPath() == expected
}

func mailboxS3ObjectStoreCanonicalAuthority(parsed *url.URL) (string, bool) {
	if parsed == nil || parsed.Host == `` || parsed.Hostname() == `` || parsed.User != nil {
		return ``, false
	}
	authority := strings.ToLower(parsed.Host)
	if authority == `` || strings.ContainsAny(authority, string([]byte{32, 9, 13, 10})) {
		return ``, false
	}
	return authority, true
}


func mailboxS3ObjectStorePutHeaders(headers http.Header, request MailboxObjectPresignPutRequest, proof mailboxS3ObjectStorePresignProof) (map[string]string, bool) {
	if headers == nil || len(headers) == 0 || len(headers) > mailboxObjectStoreMaximumHeaders {
		return nil, false
	}
	result := make(map[string]string, len(headers))
	seen := make(map[string]struct{}, len(headers))
	contentLength := strconv.FormatInt(request.ContentLength, 10)
	hasHost := false
	hasContentLength := false
	hasContentType := false
	for name, values := range headers {
		canonicalName := http.CanonicalHeaderKey(name)
		if canonicalName == `` || !validMailboxObjectStoreHeaderName(canonicalName) || len(values) != 1 || !validMailboxObjectStoreText(values[0], mailboxObjectStoreMaximumHeaderValue, false) {
			return nil, false
		}
		if _, duplicate := seen[canonicalName]; duplicate {
			return nil, false
		}
		seen[canonicalName] = struct{}{}
		lowerName := strings.ToLower(canonicalName)
		if _, signed := proof.signedHeaders[lowerName]; !signed {
			return nil, false
		}
		switch canonicalName {
		case `Host`:
			if values[0] != proof.authority {
				return nil, false
			}
			hasHost = true
		case `Content-Length`:
			if values[0] != contentLength {
				return nil, false
			}
			hasContentLength = true
		case `Content-Type`:
			if values[0] != request.ContentType {
				return nil, false
			}
			hasContentType = true
		default:
			if !mailboxS3ObjectStoreBrowserHeaderSafe(canonicalName) {
				return nil, false
			}
		}
		if canonicalName != `Host` && canonicalName != `Content-Length` {
			if !mailboxS3ObjectStoreBrowserHeaderSafe(canonicalName) {
				return nil, false
			}
			result[canonicalName] = values[0]
		}
	}
	if !hasHost || !hasContentLength || !hasContentType {
		return nil, false
	}
	for name := range proof.signedHeaders {
		if name == `host` || name == `content-length` {
			continue
		}
		canonicalName := http.CanonicalHeaderKey(name)
		if canonicalName == `` || !mailboxS3ObjectStoreBrowserHeaderSafe(canonicalName) {
			return nil, false
		}
		if _, found := result[canonicalName]; !found {
			return nil, false
		}
	}
	return result, true
}


func mailboxS3ObjectStoreBrowserHeaderSafe(name string) bool {
	lowerName := strings.ToLower(name)
	if strings.HasPrefix(lowerName, `proxy-`) || strings.HasPrefix(lowerName, `sec-`) || strings.HasPrefix(lowerName, `x-forwarded-`) {
		return false
	}
	switch lowerName {
	case `accept-charset`, `accept-encoding`, `access-control-request-headers`, `access-control-request-method`, `authorization`, `connection`, `content-length`, `cookie`, `cookie2`, `date`, `dnt`, `expect`, `forwarded`, `host`, `keep-alive`, `origin`, `permissions-policy`, `referer`, `te`, `trailer`, `transfer-encoding`, `upgrade`, `user-agent`, `via`, `x-amz-acl`, `x-amz-bypass-governance-retention`, `x-amz-expected-bucket-owner`, `x-amz-grant-full-control`, `x-amz-grant-read`, `x-amz-grant-read-acp`, `x-amz-grant-write-acp`, `x-amz-object-lock-legal-hold`, `x-amz-object-lock-mode`, `x-amz-object-lock-retain-until-date`, `x-amz-request-payer`, `x-amz-security-token`, `x-amz-server-side-encryption`, `x-amz-server-side-encryption-aws-kms-key-id`, `x-amz-server-side-encryption-bucket-key-enabled`, `x-amz-server-side-encryption-context`, `x-amz-storage-class`, `x-amz-tagging`, `x-amz-website-redirect-location`:
		return false
	case `content-type`:
		return true
	default:
		return strings.HasPrefix(lowerName, `x-amz-checksum-`)
	}
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
