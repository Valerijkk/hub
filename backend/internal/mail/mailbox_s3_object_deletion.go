package mail

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

var _ MailObjectDeletionPort = (*MailboxS3ObjectStore)(nil)

// DeleteObject applies one idempotent object-removal command. A successful
// provider response confirms only command application, not prior existence.
func (store *MailboxS3ObjectStore) DeleteObject(ctx context.Context, request MailObjectDeleteRequest) (MailObjectDeleteResult, error) {
	if !mailboxS3ObjectStoreReady(store) {
		return MailObjectDeleteResult{}, mailObjectDeletionError(MailObjectDeletionErrorUnavailable)
	}
	if ctx == nil || !mailboxS3ObjectDeletionRequestValid(request) {
		return MailObjectDeleteResult{}, mailObjectDeletionError(MailObjectDeletionErrorRejectedPolicy)
	}
	if request.Bucket != store.bucket {
		return MailObjectDeleteResult{}, mailObjectDeletionError(MailObjectDeletionErrorUnauthorizedScope)
	}

	output, err := store.objectClient.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(request.Bucket),
		Key:    aws.String(request.ObjectKey),
	})
	if err != nil {
		return MailObjectDeleteResult{}, mailboxS3ObjectDeletionDeleteFailure(ctx, err)
	}
	if output == nil {
		return MailObjectDeleteResult{}, mailObjectDeletionError(MailObjectDeletionErrorUnavailable)
	}
	return MailObjectDeleteResult{State: MailObjectDeletionStateDeleted}, nil
}

// ReconcileObjectAbsence observes one object without mutating provider state.
func (store *MailboxS3ObjectStore) ReconcileObjectAbsence(ctx context.Context, request MailObjectAbsenceRequest) (MailObjectAbsenceResult, error) {
	if !mailboxS3ObjectStoreReady(store) {
		return MailObjectAbsenceResult{}, mailObjectDeletionError(MailObjectDeletionErrorUnavailable)
	}
	if ctx == nil || !mailboxS3ObjectDeletionRequestValid(MailObjectDeleteRequest(request)) {
		return MailObjectAbsenceResult{}, mailObjectDeletionError(MailObjectDeletionErrorRejectedPolicy)
	}
	if request.Bucket != store.bucket {
		return MailObjectAbsenceResult{}, mailObjectDeletionError(MailObjectDeletionErrorUnauthorizedScope)
	}

	output, err := store.objectClient.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(request.Bucket),
		Key:    aws.String(request.ObjectKey),
	})
	if err != nil {
		if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return MailObjectAbsenceResult{}, mailObjectDeletionError(MailObjectDeletionErrorUnavailable)
		}
		if mailboxS3ObjectDeletionAbsenceAuthoritative(err) {
			return MailObjectAbsenceResult{State: MailObjectAbsenceStateAuthoritativelyAbsent}, nil
		}
		return MailObjectAbsenceResult{}, mailboxS3ObjectDeletionAbsenceFailure(ctx, err)
	}
	if output == nil {
		return MailObjectAbsenceResult{}, mailObjectDeletionError(MailObjectDeletionErrorUnavailable)
	}
	return MailObjectAbsenceResult{State: MailObjectAbsenceStatePresent}, nil
}

func mailboxS3ObjectDeletionRequestValid(request MailObjectDeleteRequest) bool {
	if validateMailObjectDeleteRequest(request) != nil {
		return false
	}
	return validMailboxObjectLocator(MailboxObjectLocator{Bucket: request.Bucket, Key: request.ObjectKey})
}

// mailboxS3ObjectDeletionDeleteFailure fails closed after a mutating provider
// call: an error cannot prove whether the provider applied deletion.
func mailboxS3ObjectDeletionDeleteFailure(ctx context.Context, err error) error {
	if err == nil {
		return mailObjectDeletionError(MailObjectDeletionErrorUnavailable)
	}
	if ctx != nil && ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return mailObjectDeletionError(MailObjectDeletionErrorIndeterminate)
	}

	var noSuchKey *s3types.NoSuchKey
	if errors.As(err, &noSuchKey) && noSuchKey != nil {
		return mailObjectDeletionError(MailObjectDeletionErrorIndeterminate)
	}
	if kind, matched := mailboxS3ObjectDeletionAPIErrorKind(err, true); matched {
		return mailObjectDeletionError(kind)
	}
	if kind, matched := mailboxS3ObjectDeletionResponseErrorKind(err, true); matched {
		return mailObjectDeletionError(kind)
	}
	return mailObjectDeletionError(MailObjectDeletionErrorIndeterminate)
}

func mailboxS3ObjectDeletionAbsenceFailure(ctx context.Context, err error) error {
	if err == nil {
		return mailObjectDeletionError(MailObjectDeletionErrorUnavailable)
	}
	if ctx != nil && ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return mailObjectDeletionError(MailObjectDeletionErrorUnavailable)
	}

	var noSuchKey *s3types.NoSuchKey
	if errors.As(err, &noSuchKey) && noSuchKey != nil {
		return mailObjectDeletionError(MailObjectDeletionErrorUnavailable)
	}
	if kind, matched := mailboxS3ObjectDeletionAPIErrorKind(err, false); matched {
		return mailObjectDeletionError(kind)
	}
	if kind, matched := mailboxS3ObjectDeletionResponseErrorKind(err, false); matched {
		return mailObjectDeletionError(kind)
	}
	return mailObjectDeletionError(MailObjectDeletionErrorUnavailable)
}

func mailboxS3ObjectDeletionAbsenceAuthoritative(err error) bool {
	var noSuchKey *s3types.NoSuchKey
	if errors.As(err, &noSuchKey) && noSuchKey != nil {
		return true
	}
	var apiError smithy.APIError
	if errors.As(err, &apiError) && apiError != nil {
		switch apiError.ErrorCode() {
		case "NoSuchKey", "NotFound":
			return true
		}
	}
	var responseError *smithyhttp.ResponseError
	return errors.As(err, &responseError) && responseError != nil && responseError.HTTPResponse() != nil && responseError.HTTPStatusCode() == http.StatusNotFound
}

func mailboxS3ObjectDeletionAPIErrorKind(err error, deleteCall bool) (MailObjectDeletionErrorKind, bool) {
	var apiError smithy.APIError
	if !errors.As(err, &apiError) || apiError == nil {
		return "", false
	}
	switch apiError.ErrorCode() {
	case "NoSuchKey", "NotFound":
		if deleteCall {
			return MailObjectDeletionErrorIndeterminate, true
		}
		return "", false
	case "NoSuchBucket", "NoSuchObject":
		if deleteCall {
			return MailObjectDeletionErrorIndeterminate, true
		}
		return "", false
	case "OperationAborted", "ConditionalRequestConflict", "PreconditionFailed":
		return MailObjectDeletionErrorConflict, true
	case "AccessDenied", "InvalidAccessKeyId", "SignatureDoesNotMatch", "AuthorizationHeaderMalformed", "InvalidToken", "ExpiredToken", "RequestExpired":
		return MailObjectDeletionErrorUnauthorizedScope, true
	case "InvalidRequest", "InvalidArgument", "InvalidBucketName", "InvalidObjectState", "MalformedXML", "MalformedACLError":
		return MailObjectDeletionErrorRejectedPolicy, true
	default:
		return "", false
	}
}

func mailboxS3ObjectDeletionResponseErrorKind(err error, deleteCall bool) (MailObjectDeletionErrorKind, bool) {
	var responseError *smithyhttp.ResponseError
	if !errors.As(err, &responseError) || responseError == nil || responseError.HTTPResponse() == nil {
		return "", false
	}
	switch responseError.HTTPStatusCode() {
	case http.StatusNotFound:
		if deleteCall {
			return MailObjectDeletionErrorIndeterminate, true
		}
		return "", false
	case http.StatusGone:
		if deleteCall {
			return MailObjectDeletionErrorIndeterminate, true
		}
		return MailObjectDeletionErrorUnavailable, true
	case http.StatusConflict, http.StatusPreconditionFailed:
		return MailObjectDeletionErrorConflict, true
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return MailObjectDeletionErrorRejectedPolicy, true
	case http.StatusUnauthorized, http.StatusForbidden:
		return MailObjectDeletionErrorUnauthorizedScope, true
	default:
		return "", false
	}
}
