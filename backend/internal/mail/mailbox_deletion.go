package mail

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/google/uuid"
)

// EngineAccountDeletionPort deletes or reconciles one provisioned engine account.
type EngineAccountDeletionPort interface {
	DeleteAccount(context.Context, EngineAccountDeleteRequest) (EngineAccountDeleteResult, error)
	ReconcileAccountAbsence(context.Context, EngineAccountAbsenceRequest) (EngineAccountAbsenceResult, error)
}

// MailObjectDeletionPort deletes or reconciles one mail object.
type MailObjectDeletionPort interface {
	DeleteObject(context.Context, MailObjectDeleteRequest) (MailObjectDeleteResult, error)
	ReconcileObjectAbsence(context.Context, MailObjectAbsenceRequest) (MailObjectAbsenceResult, error)
}

// EngineAccountDeleteRequest identifies one provisioned engine account deletion.
type EngineAccountDeleteRequest struct {
	ProvisioningKey string
	EngineAccountID string
	DeletionID      uuid.UUID
}

// MailObjectDeleteRequest identifies one mail object deletion.
type MailObjectDeleteRequest struct {
	Bucket     string
	ObjectKey  string
	DeletionID uuid.UUID
}

type EngineAccountAbsenceRequest = EngineAccountDeleteRequest
type MailObjectAbsenceRequest = MailObjectDeleteRequest

type EngineAccountDeletionState string

const (
	EngineAccountDeletionStateDeleted               EngineAccountDeletionState = "deleted"
	EngineAccountDeletionStateAuthoritativelyAbsent EngineAccountDeletionState = "authoritatively_absent"
)

type EngineAccountAbsenceState string

const (
	EngineAccountAbsenceStatePresent               EngineAccountAbsenceState = "present"
	EngineAccountAbsenceStateAuthoritativelyAbsent EngineAccountAbsenceState = "authoritatively_absent"
)

type MailObjectDeletionState string

const (
	MailObjectDeletionStateDeleted               MailObjectDeletionState = "deleted"
	MailObjectDeletionStateAuthoritativelyAbsent MailObjectDeletionState = "authoritatively_absent"
)

type MailObjectAbsenceState string

const (
	MailObjectAbsenceStatePresent               MailObjectAbsenceState = "present"
	MailObjectAbsenceStateAuthoritativelyAbsent MailObjectAbsenceState = "authoritatively_absent"
)

type EngineAccountDeleteResult struct {
	State EngineAccountDeletionState
}

type EngineAccountAbsenceResult struct {
	State EngineAccountAbsenceState
}

type MailObjectDeleteResult struct {
	State MailObjectDeletionState
}

type MailObjectAbsenceResult struct {
	State MailObjectAbsenceState
}

type MailObjectDeletionErrorKind string

const (
	MailObjectDeletionErrorUnavailable       MailObjectDeletionErrorKind = "unavailable"
	MailObjectDeletionErrorIndeterminate     MailObjectDeletionErrorKind = "indeterminate"
	MailObjectDeletionErrorConflict          MailObjectDeletionErrorKind = "conflict"
	MailObjectDeletionErrorRejectedPolicy    MailObjectDeletionErrorKind = "rejected_policy"
	MailObjectDeletionErrorUnauthorizedScope MailObjectDeletionErrorKind = "unauthorized_scope"
)

type MailObjectDeletionError struct {
	kind MailObjectDeletionErrorKind
}

func NewMailObjectDeletionError(kind MailObjectDeletionErrorKind) (*MailObjectDeletionError, error) {
	if !validMailObjectDeletionErrorKind(kind) {
		return nil, errors.New("invalid mail object deletion error kind")
	}
	return &MailObjectDeletionError{kind: kind}, nil
}

func (*MailObjectDeletionError) Error() string { return "mail object deletion failed" }

func (e *MailObjectDeletionError) Kind() MailObjectDeletionErrorKind {
	if e == nil || !validMailObjectDeletionErrorKind(e.kind) {
		return ""
	}
	return e.kind
}

func MailObjectDeletionErrorKindOf(err error) (MailObjectDeletionErrorKind, bool) {
	var target *MailObjectDeletionError
	if !errors.As(err, &target) || target == nil {
		return "", false
	}
	if kind := target.Kind(); kind != "" {
		return kind, true
	}
	return "", false
}

func validateEngineAccountDeleteRequest(request EngineAccountDeleteRequest) error {
	if request.DeletionID == uuid.Nil || !validMailDeletionUTF8(request.ProvisioningKey, MaxEngineOpaqueIDBytes) || !validMailDeletionUTF8(request.EngineAccountID, MaxEngineOpaqueIDBytes) {
		return &EngineError{Kind: EngineErrorRejectedPolicy}
	}
	return nil
}

func validateMailObjectDeleteRequest(request MailObjectDeleteRequest) error {
	if request.DeletionID == uuid.Nil || !validMailDeletionUTF8(request.Bucket, 63) || !validMailDeletionUTF8(request.ObjectKey, 512) {
		return mailObjectDeletionError(MailObjectDeletionErrorRejectedPolicy)
	}
	return nil
}

func validEngineAccountDeletionIdentity(provisioningKey, engineAccountID string) bool {
	return validMailDeletionUTF8(provisioningKey, MaxEngineOpaqueIDBytes) && validMailDeletionUTF8(engineAccountID, MaxEngineOpaqueIDBytes)
}

func validMailObjectDeletionIdentity(bucket, objectKey string) bool {
	return validMailDeletionUTF8(bucket, 63) && validMailDeletionUTF8(objectKey, 512)
}

func validEngineAccountDeletionState(state EngineAccountDeletionState) bool {
	switch state {
	case EngineAccountDeletionStateDeleted, EngineAccountDeletionStateAuthoritativelyAbsent:
		return true
	default:
		return false
	}
}

func validEngineAccountAbsenceState(state EngineAccountAbsenceState) bool {
	switch state {
	case EngineAccountAbsenceStatePresent, EngineAccountAbsenceStateAuthoritativelyAbsent:
		return true
	default:
		return false
	}
}

func validMailObjectDeletionState(state MailObjectDeletionState) bool {
	switch state {
	case MailObjectDeletionStateDeleted, MailObjectDeletionStateAuthoritativelyAbsent:
		return true
	default:
		return false
	}
}

func validMailObjectAbsenceState(state MailObjectAbsenceState) bool {
	switch state {
	case MailObjectAbsenceStatePresent, MailObjectAbsenceStateAuthoritativelyAbsent:
		return true
	default:
		return false
	}
}

func validMailObjectDeletionErrorKind(kind MailObjectDeletionErrorKind) bool {
	switch kind {
	case MailObjectDeletionErrorUnavailable,
		MailObjectDeletionErrorIndeterminate,
		MailObjectDeletionErrorConflict,
		MailObjectDeletionErrorRejectedPolicy,
		MailObjectDeletionErrorUnauthorizedScope:
		return true
	default:
		return false
	}
}

func validMailDeletionUTF8(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value)
}

func mailObjectDeletionError(kind MailObjectDeletionErrorKind) *MailObjectDeletionError {
	return &MailObjectDeletionError{kind: kind}
}
