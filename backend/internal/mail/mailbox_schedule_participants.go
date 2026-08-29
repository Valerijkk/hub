package mail

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

const MaxMailboxScheduleParticipantAddresses = 200

type MailboxScheduleParticipantsInput struct {
	MessageID uuid.UUID
	Addresses []string
}

type MailboxScheduleParticipant struct {
	Address string
	UserID  *uuid.UUID
}

type MailboxScheduleParticipantsResult struct {
	Participants []MailboxScheduleParticipant
}

type MailboxScheduleParticipantsRepositoryRequest struct {
	ActorID        uuid.UUID
	OrganizationID uuid.UUID
	AccountID      uuid.UUID
	MessageID      uuid.UUID
	Addresses      []string
}

type MailboxScheduleParticipantsRepositoryResult struct {
	Participants []MailboxScheduleParticipant
}

type MailboxScheduleParticipantsRepository interface {
	ResolveMailboxScheduleParticipants(
		context.Context,
		MailboxScheduleParticipantsRepositoryRequest,
	) (MailboxScheduleParticipantsRepositoryResult, error)
}

type MailboxScheduleParticipantsServiceAPI interface {
	ResolveMailboxScheduleParticipants(
		context.Context,
		uuid.UUID,
		MailboxScheduleParticipantsInput,
	) (MailboxScheduleParticipantsResult, error)
}

type MailboxScheduleParticipantsErrorKind string

const (
	MailboxScheduleParticipantsErrorInvalidRequest MailboxScheduleParticipantsErrorKind = "invalid_request"
	MailboxScheduleParticipantsErrorNotFound       MailboxScheduleParticipantsErrorKind = "not_found"
	MailboxScheduleParticipantsErrorDeleting       MailboxScheduleParticipantsErrorKind = "deleting"
	MailboxScheduleParticipantsErrorUnavailable    MailboxScheduleParticipantsErrorKind = "unavailable"
	MailboxScheduleParticipantsErrorInternal       MailboxScheduleParticipantsErrorKind = "internal"
)

const mailboxScheduleParticipantsErrorText = "mailbox schedule participants error"

type MailboxScheduleParticipantsError struct {
	kind MailboxScheduleParticipantsErrorKind
}

func NewMailboxScheduleParticipantsError(kind MailboxScheduleParticipantsErrorKind) *MailboxScheduleParticipantsError {
	if !validMailboxScheduleParticipantsErrorKind(kind) {
		kind = MailboxScheduleParticipantsErrorInternal
	}
	return &MailboxScheduleParticipantsError{kind: kind}
}

func (*MailboxScheduleParticipantsError) Error() string {
	return mailboxScheduleParticipantsErrorText
}

func (err *MailboxScheduleParticipantsError) Kind() MailboxScheduleParticipantsErrorKind {
	if err == nil || !validMailboxScheduleParticipantsErrorKind(err.kind) {
		return MailboxScheduleParticipantsErrorInternal
	}
	return err.kind
}

func MailboxScheduleParticipantsErrorKindOf(err error) (MailboxScheduleParticipantsErrorKind, bool) {
	var target *MailboxScheduleParticipantsError
	if !errors.As(err, &target) || target == nil || !validMailboxScheduleParticipantsErrorKind(target.kind) {
		return MailboxScheduleParticipantsErrorInternal, false
	}
	return target.kind, true
}

type MailboxScheduleParticipantsService struct {
	owners     mailboxOwnerResolver
	repository MailboxScheduleParticipantsRepository
}

func NewMailboxScheduleParticipantsService(
	owners mailboxOwnerResolver,
	repository MailboxScheduleParticipantsRepository,
) (*MailboxScheduleParticipantsService, error) {
	if mailboxScheduleParticipantsDependencyNil(owners) || mailboxScheduleParticipantsDependencyNil(repository) {
		return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	return &MailboxScheduleParticipantsService{owners: owners, repository: repository}, nil
}

func (service *MailboxScheduleParticipantsService) ResolveMailboxScheduleParticipants(
	ctx context.Context,
	actorID uuid.UUID,
	input MailboxScheduleParticipantsInput,
) (MailboxScheduleParticipantsResult, error) {
	if service == nil ||
		mailboxScheduleParticipantsDependencyNil(service.owners) ||
		mailboxScheduleParticipantsDependencyNil(service.repository) {
		return MailboxScheduleParticipantsResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	if mailboxScheduleParticipantsDependencyNil(ctx) {
		return MailboxScheduleParticipantsResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	if ctx.Err() != nil {
		return MailboxScheduleParticipantsResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorUnavailable)
	}
	if actorID == uuid.Nil || input.MessageID == uuid.Nil {
		return MailboxScheduleParticipantsResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInvalidRequest)
	}

	addresses, valid := mailboxScheduleParticipantsAddresses(input.Addresses)
	if !valid {
		return MailboxScheduleParticipantsResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInvalidRequest)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxScheduleParticipantsResult{}, mailboxScheduleParticipantsOwnerError(err)
	}
	organizationID, accountID, err := mailboxScheduleParticipantsOwnerScope(actorID, owner)
	if err != nil {
		return MailboxScheduleParticipantsResult{}, err
	}

	repositoryResult, err := service.repository.ResolveMailboxScheduleParticipants(ctx, MailboxScheduleParticipantsRepositoryRequest{
		ActorID:        actorID,
		OrganizationID: organizationID,
		AccountID:      accountID,
		MessageID:      input.MessageID,
		Addresses:      addresses,
	})
	if err != nil {
		return MailboxScheduleParticipantsResult{}, mailboxScheduleParticipantsRepositoryError(err)
	}

	result, valid := mailboxScheduleParticipantsResultCopy(repositoryResult, addresses)
	if !valid {
		return MailboxScheduleParticipantsResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	return result, nil
}

func validMailboxScheduleParticipantsErrorKind(kind MailboxScheduleParticipantsErrorKind) bool {
	switch kind {
	case MailboxScheduleParticipantsErrorInvalidRequest,
		MailboxScheduleParticipantsErrorNotFound,
		MailboxScheduleParticipantsErrorDeleting,
		MailboxScheduleParticipantsErrorUnavailable,
		MailboxScheduleParticipantsErrorInternal:
		return true
	default:
		return false
	}
}

func mailboxScheduleParticipantsDependencyNil(dependency any) bool {
	if dependency == nil {
		return true
	}

	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func mailboxScheduleParticipantsAddresses(addresses []string) ([]string, bool) {
	if len(addresses) == 0 || len(addresses) > MaxMailboxScheduleParticipantAddresses {
		return nil, false
	}

	validated := make([]string, 0, len(addresses))
	seen := make(map[string]struct{}, len(addresses))
	for _, address := range addresses {
		canonical, err := NormalizeAddress(EngineAddress{Email: address})
		if err != nil || canonical.Email != address {
			return nil, false
		}
		if _, exists := seen[canonical.Email]; exists {
			return nil, false
		}
		seen[canonical.Email] = struct{}{}
		validated = append(validated, strings.Clone(address))
	}
	return validated, true
}

func mailboxScheduleParticipantsOwnerScope(actorID uuid.UUID, owner MailboxOwner) (uuid.UUID, uuid.UUID, error) {
	organizationID := owner.OrganizationID()
	if actorID == uuid.Nil || organizationID == uuid.Nil {
		return uuid.Nil, uuid.Nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}

	accountID, hasAccount := owner.AccountID()
	projectionState, hasProjectionState := owner.ProjectionState()
	switch owner.State() {
	case MailboxOwnerNotConfigured:
		if hasAccount || accountID != uuid.Nil || hasProjectionState || projectionState != "" {
			return uuid.Nil, uuid.Nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
		return uuid.Nil, uuid.Nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorNotFound)
	case MailboxOwnerDeleting:
		if !mailboxScheduleParticipantsOwnerSnapshotValid(accountID, hasAccount, projectionState, hasProjectionState) {
			return uuid.Nil, uuid.Nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
		return uuid.Nil, uuid.Nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorDeleting)
	case MailboxOwnerProvisioning, MailboxOwnerSuspended, MailboxOwnerError:
		if !mailboxScheduleParticipantsOwnerSnapshotValid(accountID, hasAccount, projectionState, hasProjectionState) {
			return uuid.Nil, uuid.Nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
		return uuid.Nil, uuid.Nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorUnavailable)
	case MailboxOwnerActive:
		if !mailboxScheduleParticipantsOwnerSnapshotValid(accountID, hasAccount, projectionState, hasProjectionState) {
			return uuid.Nil, uuid.Nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
		if projectionState != MailboxProjectionReady {
			return uuid.Nil, uuid.Nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorUnavailable)
		}
		return organizationID, accountID, nil
	default:
		return uuid.Nil, uuid.Nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
}

func mailboxScheduleParticipantsOwnerSnapshotValid(
	accountID uuid.UUID,
	hasAccount bool,
	projectionState MailboxProjectionState,
	hasProjectionState bool,
) bool {
	return hasAccount && accountID != uuid.Nil && hasProjectionState && validMailboxProjectionState(projectionState)
}

func mailboxScheduleParticipantsOwnerError(err error) error {
	var ownerError *MailboxServiceError
	if !errors.As(err, &ownerError) || ownerError == nil {
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}

	switch ownerError.Kind {
	case MailboxServiceErrorNotFound:
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorNotFound)
	case MailboxServiceErrorDeleting:
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorDeleting)
	case MailboxServiceErrorUnavailable:
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorUnavailable)
	case MailboxServiceErrorInvalid:
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	default:
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
}

func mailboxScheduleParticipantsRepositoryError(err error) error {
	kind, ok := MailboxScheduleParticipantsErrorKindOf(err)
	if !ok {
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}

	switch kind {
	case MailboxScheduleParticipantsErrorNotFound,
		MailboxScheduleParticipantsErrorDeleting,
		MailboxScheduleParticipantsErrorUnavailable,
		MailboxScheduleParticipantsErrorInternal:
		return NewMailboxScheduleParticipantsError(kind)
	default:
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
}

func mailboxScheduleParticipantsResultCopy(
	repositoryResult MailboxScheduleParticipantsRepositoryResult,
	addresses []string,
) (MailboxScheduleParticipantsResult, bool) {
	if len(repositoryResult.Participants) != len(addresses) {
		return MailboxScheduleParticipantsResult{}, false
	}

	participants := make([]MailboxScheduleParticipant, len(repositoryResult.Participants))
	for index, participant := range repositoryResult.Participants {
		if participant.Address != addresses[index] || (participant.UserID != nil && *participant.UserID == uuid.Nil) {
			return MailboxScheduleParticipantsResult{}, false
		}

		participants[index].Address = strings.Clone(participant.Address)
		if participant.UserID != nil {
			userID := *participant.UserID
			participants[index].UserID = &userID
		}
	}
	return MailboxScheduleParticipantsResult{Participants: participants}, true
}
