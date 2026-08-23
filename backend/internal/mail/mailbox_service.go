package mail

import (
	"context"
	"errors"
	"reflect"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrMailboxServiceInvalid     = errors.New("mailbox service invalid")
	ErrMailboxServiceNotFound    = errors.New("mailbox service not found")
	ErrMailboxServiceUnavailable = errors.New("mailbox service unavailable")
	ErrMailboxServiceDeleting    = errors.New("mailbox service deleting")
)

type MailboxServiceErrorKind string

const (
	MailboxServiceErrorInvalid     MailboxServiceErrorKind = "invalid"
	MailboxServiceErrorNotFound    MailboxServiceErrorKind = "not_found"
	MailboxServiceErrorUnavailable MailboxServiceErrorKind = "unavailable"
	MailboxServiceErrorDeleting    MailboxServiceErrorKind = "deleting"
)

type MailboxServiceError struct {
	Kind MailboxServiceErrorKind
}

func (errorValue *MailboxServiceError) Error() string {
	if errorValue == nil {
		return "mailbox service error"
	}

	switch errorValue.Kind {
	case MailboxServiceErrorInvalid:
		return ErrMailboxServiceInvalid.Error()
	case MailboxServiceErrorNotFound:
		return ErrMailboxServiceNotFound.Error()
	case MailboxServiceErrorUnavailable:
		return ErrMailboxServiceUnavailable.Error()
	case MailboxServiceErrorDeleting:
		return ErrMailboxServiceDeleting.Error()
	default:
		return "mailbox service error"
	}
}

func (errorValue *MailboxServiceError) Unwrap() error {
	if errorValue == nil {
		return nil
	}

	switch errorValue.Kind {
	case MailboxServiceErrorInvalid:
		return ErrMailboxServiceInvalid
	case MailboxServiceErrorNotFound:
		return ErrMailboxServiceNotFound
	case MailboxServiceErrorUnavailable:
		return ErrMailboxServiceUnavailable
	case MailboxServiceErrorDeleting:
		return ErrMailboxServiceDeleting
	default:
		return nil
	}
}

type MailboxService struct {
	repository MailboxOwnerRepository
}

func NewMailboxService(repository MailboxOwnerRepository) (*MailboxService, error) {
	if mailboxOwnerRepositoryNil(repository) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	return &MailboxService{repository: repository}, nil
}

type MailboxOwner struct {
	state              MailboxOwnerState
	accountID          uuid.UUID
	hasAccount         bool
	organizationID     uuid.UUID
	engineAccountID    string
	hasEngineAccountID bool
	primaryAddress     string
	hasPrimaryAddress  bool
	recoveryCode       MailboxOwnerRecoveryCode
	hasRecoveryCode    bool
	projectionState    MailboxProjectionState
	hasProjectionState bool
}

func (owner MailboxOwner) State() MailboxOwnerState {
	return owner.state
}

func (owner MailboxOwner) AccountID() (uuid.UUID, bool) {
	return owner.accountID, owner.hasAccount
}

func (owner MailboxOwner) OrganizationID() uuid.UUID {
	return owner.organizationID
}

func (owner MailboxOwner) PrimaryAddress() (string, bool) {
	return owner.primaryAddress, owner.hasPrimaryAddress
}

func (owner MailboxOwner) RecoveryCode() (MailboxOwnerRecoveryCode, bool) {
	return owner.recoveryCode, owner.hasRecoveryCode
}

func (owner MailboxOwner) ProjectionState() (MailboxProjectionState, bool) {
	return owner.projectionState, owner.hasProjectionState
}

func (service *MailboxService) ResolveOwner(ctx context.Context, actorID uuid.UUID) (MailboxOwner, error) {
	if service == nil || mailboxOwnerRepositoryNil(service.repository) || ctx == nil || actorID == uuid.Nil {
		return MailboxOwner{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	record, err := service.repository.ResolveMailboxOwner(ctx, actorID)
	if err != nil {
		return MailboxOwner{}, mailboxServiceRepositoryError(err)
	}

	return mailboxOwnerFromRecord(actorID, record)
}

func mailboxOwnerFromRecord(actorID uuid.UUID, record MailboxOwnerRecord) (MailboxOwner, error) {
	if record.ActorID != actorID || record.OrganizationID == uuid.Nil || !validMailboxOwnerState(record.State) {
		return MailboxOwner{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}

	if record.State == MailboxOwnerNotConfigured {
		if record.AccountID != uuid.Nil || record.EngineAccountID != nil || record.ProjectionState != "" || record.PrimaryAddress != nil || record.RecoveryCode != nil {
			return MailboxOwner{}, mailboxServiceError(MailboxServiceErrorNotFound)
		}
		return MailboxOwner{
			state:          record.State,
			organizationID: record.OrganizationID,
		}, nil
	}

	if record.AccountID == uuid.Nil {
		return MailboxOwner{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}

	owner, valid := mailboxOwnerSnapshot(record)
	if !valid {
		return MailboxOwner{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}

	switch record.State {
	case MailboxOwnerDeleting:
		if owner.hasEngineAccountID || owner.hasRecoveryCode {
			return MailboxOwner{}, mailboxServiceError(MailboxServiceErrorNotFound)
		}
	case MailboxOwnerActive:
		if !owner.hasEngineAccountID || !owner.hasPrimaryAddress || owner.hasRecoveryCode {
			return MailboxOwner{}, mailboxServiceError(MailboxServiceErrorNotFound)
		}
	case MailboxOwnerSuspended:
		if !owner.hasEngineAccountID || owner.hasRecoveryCode {
			return MailboxOwner{}, mailboxServiceError(MailboxServiceErrorNotFound)
		}
	case MailboxOwnerError:
		if !owner.hasRecoveryCode {
			return MailboxOwner{}, mailboxServiceError(MailboxServiceErrorNotFound)
		}
	case MailboxOwnerProvisioning:
	default:
		return MailboxOwner{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}

	return owner, nil
}

func mailboxOwnerSnapshot(record MailboxOwnerRecord) (MailboxOwner, bool) {
	if !validMailboxProjectionState(record.ProjectionState) {
		return MailboxOwner{}, false
	}

	owner := MailboxOwner{
		state:              record.State,
		accountID:          record.AccountID,
		hasAccount:         true,
		organizationID:     record.OrganizationID,
		projectionState:    record.ProjectionState,
		hasProjectionState: true,
	}

	if record.EngineAccountID != nil {
		if !validMailboxOwnerOpaque(*record.EngineAccountID) {
			return MailboxOwner{}, false
		}
		owner.engineAccountID = *record.EngineAccountID
		owner.hasEngineAccountID = true
	}
	if record.PrimaryAddress != nil {
		if !validMailboxOwnerPrimaryAddress(*record.PrimaryAddress) {
			return MailboxOwner{}, false
		}
		owner.primaryAddress = *record.PrimaryAddress
		owner.hasPrimaryAddress = true
	}
	if record.RecoveryCode != nil {
		if !validMailboxOwnerRecoveryCode(*record.RecoveryCode) {
			return MailboxOwner{}, false
		}
		owner.recoveryCode = *record.RecoveryCode
		owner.hasRecoveryCode = true
	}

	return owner, true
}

func mailboxOwnerRepositoryNil(repository MailboxOwnerRepository) bool {
	if repository == nil {
		return true
	}

	value := reflect.ValueOf(repository)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func mailboxServiceError(kind MailboxServiceErrorKind) error {
	return &MailboxServiceError{Kind: kind}
}

func mailboxServiceRepositoryError(err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return mailboxServiceError(MailboxServiceErrorUnavailable)
	}

	switch serviceError.Kind {
	case MailboxServiceErrorInvalid, MailboxServiceErrorNotFound, MailboxServiceErrorUnavailable, MailboxServiceErrorDeleting:
		return mailboxServiceError(serviceError.Kind)
	default:
		return mailboxServiceError(MailboxServiceErrorUnavailable)
	}
}

func validMailboxOwnerState(state MailboxOwnerState) bool {
	switch state {
	case MailboxOwnerNotConfigured, MailboxOwnerProvisioning, MailboxOwnerActive, MailboxOwnerSuspended, MailboxOwnerError, MailboxOwnerDeleting:
		return true
	default:
		return false
	}
}

func validMailboxProjectionState(state MailboxProjectionState) bool {
	switch state {
	case MailboxProjectionBootstrapping, MailboxProjectionReady, MailboxProjectionDegraded, MailboxProjectionRebuilding:
		return true
	default:
		return false
	}
}

func validMailboxOwnerRecoveryCode(code MailboxOwnerRecoveryCode) bool {
	switch code {
	case "MAILBOX_ENGINE_DISABLED", "MAILBOX_DOMAIN_REQUIRED", "MAILBOX_PROVISIONING_RETRY", "MAILBOX_SUSPENDED", "MAILBOX_UNAVAILABLE":
		return true
	default:
		return false
	}
}

func validMailboxOwnerOpaque(value string) bool {
	if value == "" || len(value) > MaxEngineOpaqueIDBytes || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validMailboxOwnerPrimaryAddress(value string) bool {
	normalized, err := NormalizeAddress(EngineAddress{Email: value})
	return err == nil && normalized.Email == value
}
