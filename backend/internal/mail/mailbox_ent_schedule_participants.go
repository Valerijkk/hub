package mail

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/ent/user"
)

var _ MailboxScheduleParticipantsRepository = (*MailboxEntRepository)(nil)

// ResolveMailboxScheduleParticipants proves visible header membership before
// the single bounded same-organization user lookup.
func (repository *MailboxEntRepository) ResolveMailboxScheduleParticipants(
	ctx context.Context,
	request MailboxScheduleParticipantsRepositoryRequest,
) (MailboxScheduleParticipantsRepositoryResult, error) {
	if repository == nil || repository.client == nil || mailboxScheduleParticipantsDependencyNil(ctx) {
		return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	if ctx.Err() != nil {
		return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorUnavailable)
	}
	addresses, valid := mailboxEntScheduleParticipantsRequestAddresses(request)
	if !valid {
		return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}

	var result MailboxScheduleParticipantsRepositoryResult
	err := repository.withMailboxEntReadSnapshot(ctx, func(transaction *ent.Tx) error {
		if err := mailboxEntScheduleParticipantsAccountScope(ctx, transaction, request); err != nil {
			return err
		}
		candidates, err := mailboxEntScheduleParticipantsHeaderCandidates(ctx, transaction, request)
		if err != nil {
			return err
		}
		selfAddresses, err := mailboxEntScheduleParticipantsSelfAddresses(ctx, transaction, request.AccountID)
		if err != nil {
			return err
		}
		for address := range selfAddresses {
			delete(candidates, address)
		}
		for _, address := range addresses {
			if _, found := candidates[address]; !found {
				return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorNotFound)
			}
		}

		result, err = mailboxEntScheduleParticipantsUserResolution(ctx, transaction, request, addresses)
		return err
	})
	if err != nil {
		return MailboxScheduleParticipantsRepositoryResult{}, mailboxEntScheduleParticipantsSafeError(err)
	}
	return mailboxEntScheduleParticipantsResultCopy(result, addresses)
}

func mailboxEntScheduleParticipantsRequestAddresses(request MailboxScheduleParticipantsRepositoryRequest) ([]string, bool) {
	if request.ActorID == uuid.Nil || request.OrganizationID == uuid.Nil || request.AccountID == uuid.Nil || request.MessageID == uuid.Nil {
		return nil, false
	}
	return mailboxScheduleParticipantsAddresses(request.Addresses)
}

func mailboxEntScheduleParticipantsAccountScope(
	ctx context.Context,
	transaction *ent.Tx,
	request MailboxScheduleParticipantsRepositoryRequest,
) error {
	if ctx == nil || transaction == nil {
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	account, err := transaction.MailAccount.Query().
		Where(
			mailaccount.IDEQ(request.AccountID),
			mailaccount.UserIDEQ(request.ActorID),
			mailaccount.OrganizationIDEQ(request.OrganizationID),
		).
		Select(
			mailaccount.FieldID,
			mailaccount.FieldOrganizationID,
			mailaccount.FieldUserID,
			mailaccount.FieldStatus,
			mailaccount.FieldDesiredEnabled,
			mailaccount.FieldProjectionState,
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorNotFound)
		}
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorUnavailable)
	}
	if account == nil || account.ID != request.AccountID || account.UserID != request.ActorID || account.OrganizationID != request.OrganizationID ||
		!validMailboxProjectionState(MailboxProjectionState(account.ProjectionState)) {
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	switch account.Status {
	case mailaccount.StatusDeleting:
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorDeleting)
	case mailaccount.StatusActive:
		if !account.DesiredEnabled {
			return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
		if MailboxProjectionState(account.ProjectionState) != MailboxProjectionReady {
			return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorUnavailable)
		}
		return nil
	case mailaccount.StatusProvisioning, mailaccount.StatusSuspended, mailaccount.StatusError:
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorUnavailable)
	default:
		return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
}

func mailboxEntScheduleParticipantsHeaderCandidates(
	ctx context.Context,
	transaction *ent.Tx,
	request MailboxScheduleParticipantsRepositoryRequest,
) (map[string]struct{}, error) {
	if ctx == nil || transaction == nil {
		return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	message, err := transaction.MailMessage.Query().
		Where(
			mailmessage.IDEQ(request.MessageID),
			mailmessage.MailAccountIDEQ(request.AccountID),
			mailmessage.DeletedAtIsNil(),
		).
		Select(
			mailmessage.FieldID,
			mailmessage.FieldMailAccountID,
			mailmessage.FieldFrom,
			mailmessage.FieldTo,
			mailmessage.FieldCc,
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorNotFound)
		}
		return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorUnavailable)
	}
	if message == nil || message.ID != request.MessageID || message.MailAccountID != request.AccountID {
		return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	from, valid := mailboxEntScheduleParticipantsHeaderAddress(message.From)
	if !valid {
		return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	candidates := make(map[string]struct{}, 1+len(message.To)+len(message.Cc))
	candidates[from] = struct{}{}
	if !mailboxEntScheduleParticipantsAddHeaderAddresses(candidates, message.To) ||
		!mailboxEntScheduleParticipantsAddHeaderAddresses(candidates, message.Cc) {
		return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	return candidates, nil
}

func mailboxEntScheduleParticipantsHeaderAddress(address schema.MailStoredAddress) (string, bool) {
	if err := address.Validate(); err != nil {
		return "", false
	}
	canonical, err := NormalizeAddress(EngineAddress{Email: address.Email})
	if err != nil || canonical.Email != address.Email {
		return "", false
	}
	return strings.Clone(address.Email), true
}

func mailboxEntScheduleParticipantsAddHeaderAddresses(candidates map[string]struct{}, addresses schema.MailStoredAddresses) bool {
	if candidates == nil || addresses.Validate() != nil {
		return false
	}
	for _, address := range addresses {
		canonical, valid := mailboxEntScheduleParticipantsHeaderAddress(address)
		if !valid {
			return false
		}
		candidates[canonical] = struct{}{}
	}
	return true
}

func mailboxEntScheduleParticipantsSelfAddresses(
	ctx context.Context,
	transaction *ent.Tx,
	accountID uuid.UUID,
) (map[string]struct{}, error) {
	if ctx == nil || transaction == nil || accountID == uuid.Nil {
		return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	identities, err := transaction.MailIdentity.Query().
		Where(
			mailidentity.MailAccountIDEQ(accountID),
			mailidentity.StatusEQ(mailidentity.StatusActive),
			mailidentity.KindIn(mailidentity.KindPrimary, mailidentity.KindAlias),
		).
		Select(
			mailidentity.FieldID,
			mailidentity.FieldMailAccountID,
			mailidentity.FieldNormalizedAddress,
			mailidentity.FieldStatus,
			mailidentity.FieldKind,
		).
		Order(ent.Asc(mailidentity.FieldNormalizedAddress), ent.Asc(mailidentity.FieldID)).
		Limit(maxMailboxIdentityLimit + 1).
		All(ctx)
	if err != nil {
		return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorUnavailable)
	}
	if len(identities) > maxMailboxIdentityLimit {
		return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	selfAddresses := make(map[string]struct{}, len(identities))
	for _, identity := range identities {
		if identity == nil || identity.ID == uuid.Nil || identity.MailAccountID != accountID ||
			identity.Status != mailidentity.StatusActive ||
			identity.Kind != mailidentity.KindPrimary && identity.Kind != mailidentity.KindAlias {
			return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
		canonical, err := NormalizeAddress(EngineAddress{Email: identity.NormalizedAddress})
		if err != nil || canonical.Email != identity.NormalizedAddress {
			return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
		if _, duplicate := selfAddresses[canonical.Email]; duplicate {
			return nil, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
		selfAddresses[strings.Clone(canonical.Email)] = struct{}{}
	}
	return selfAddresses, nil
}

func mailboxEntScheduleParticipantsUserResolution(
	ctx context.Context,
	transaction *ent.Tx,
	request MailboxScheduleParticipantsRepositoryRequest,
	addresses []string,
) (MailboxScheduleParticipantsRepositoryResult, error) {
	if ctx == nil || transaction == nil || len(addresses) == 0 || len(addresses) > MaxMailboxScheduleParticipantAddresses {
		return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	rows, err := transaction.User.Query().
		Where(
			user.Or(
				user.IDEQ(request.ActorID),
				user.EmailIn(addresses...),
			),
			user.StatusEQ(user.StatusActive),
			user.HasOrganizationWith(organization.IDEQ(request.OrganizationID)),
		).
		Select(
			user.FieldID,
			user.FieldEmail,
			user.FieldStatus,
		).
		Order(ent.Asc(user.FieldID)).
		Limit(len(addresses) + 2).
		All(ctx)
	if err != nil {
		return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorUnavailable)
	}
	if len(rows) > len(addresses)+1 {
		return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	requested := make(map[string]struct{}, len(addresses))
	for _, address := range addresses {
		requested[address] = struct{}{}
	}
	resolved := make(map[string]uuid.UUID, len(rows))
	actorEmail := ""
	actorFound := false
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil || row.Status != user.StatusActive {
			return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
		if row.ID == request.ActorID {
			if actorFound {
				return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
			}
			canonical, err := NormalizeAddress(EngineAddress{Email: row.Email})
			if err != nil || canonical.Email != row.Email {
				return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
			}
			actorEmail = strings.Clone(row.Email)
			actorFound = true
			continue
		}
		canonical, err := NormalizeAddress(EngineAddress{Email: row.Email})
		if err != nil || canonical.Email != row.Email {
			continue
		}
		if _, selected := requested[row.Email]; !selected {
			return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
		if _, duplicate := resolved[row.Email]; duplicate {
			return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
		resolved[strings.Clone(row.Email)] = row.ID
	}
	if !actorFound {
		return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	if _, selected := requested[actorEmail]; selected {
		return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorNotFound)
	}

	participants := make([]MailboxScheduleParticipant, len(addresses))
	for index, address := range addresses {
		participants[index].Address = strings.Clone(address)
		if userID, found := resolved[address]; found {
			copiedID := userID
			participants[index].UserID = &copiedID
		}
	}
	return MailboxScheduleParticipantsRepositoryResult{Participants: participants}, nil
}

func mailboxEntScheduleParticipantsResultCopy(
	result MailboxScheduleParticipantsRepositoryResult,
	addresses []string,
) (MailboxScheduleParticipantsRepositoryResult, error) {
	if len(result.Participants) != len(addresses) {
		return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
	}
	participants := make([]MailboxScheduleParticipant, len(result.Participants))
	for index, participant := range result.Participants {
		if participant.Address != addresses[index] || participant.UserID != nil && *participant.UserID == uuid.Nil {
			return MailboxScheduleParticipantsRepositoryResult{}, NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
		participants[index].Address = strings.Clone(participant.Address)
		if participant.UserID != nil {
			userID := *participant.UserID
			participants[index].UserID = &userID
		}
	}
	return MailboxScheduleParticipantsRepositoryResult{Participants: participants}, nil
}

func mailboxEntScheduleParticipantsSafeError(err error) error {
	if kind, ok := MailboxScheduleParticipantsErrorKindOf(err); ok {
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
	var serviceError *MailboxServiceError
	if errors.As(err, &serviceError) && serviceError != nil {
		switch serviceError.Kind {
		case MailboxServiceErrorNotFound:
			return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorNotFound)
		case MailboxServiceErrorDeleting:
			return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorDeleting)
		case MailboxServiceErrorUnavailable:
			return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorUnavailable)
		default:
			return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
		}
	}
	return NewMailboxScheduleParticipantsError(MailboxScheduleParticipantsErrorInternal)
}
