package mail

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailmessage"
)

const mailboxEntCalendarPartTimeout = 5 * time.Second

var _ CalendarPartSource = (*MailboxEntCalendarPartSource)(nil)

// MailboxEntCalendarPartSource читает календарную часть письма через MailEngine.
type MailboxEntCalendarPartSource struct {
	repository *MailboxEntRepository
	engine     MailEngine
}

type mailboxEntCalendarPartSnapshot struct {
	organizationID      uuid.UUID
	mailboxID           uuid.UUID
	ownerID             uuid.UUID
	userStatus          string
	accountStatus       string
	desiredEnabled      bool
	projectionState     string
	engineAccountID     string
	engineMessageID     string
	engineVersion       string
	messageLive         bool
	sender              string
	method              MailCalendarMethod
	authenticationClass MailAuthenticationClass
	locatorHash         [sha256.Size]byte
}

// NewMailboxEntCalendarPartSource создаёт источник календарных частей на Ent и MailEngine.
func NewMailboxEntCalendarPartSource(repository *MailboxEntRepository, engine MailEngine) *MailboxEntCalendarPartSource {
	return &MailboxEntCalendarPartSource{
		repository: repository,
		engine:     engine,
	}
}

func (source *MailboxEntCalendarPartSource) ReadCalendarPart(ctx context.Context, request CalendarPartReadRequest) (CalendarPartReadResult, error) {
	var zero CalendarPartReadResult
	if source == nil || mailboxEntNil(source.repository) || mailboxEntNil(source.engine) {
		return zero, newCalendarPartSourceError(CalendarPartSourceErrorRepositoryUnavailable)
	}
	if ctx == nil ||
		request.organizationID == uuid.Nil ||
		request.mailboxID == uuid.Nil ||
		request.messageID == uuid.Nil ||
		request.maxBytes <= 0 ||
		request.maxBytes > MaxCalendarPartBytes ||
		calendarPartLocatorHashZero(request.locatorHash) {
		return zero, newCalendarPartSourceError(CalendarPartSourceErrorInvalid)
	}

	operationContext, cancel := context.WithTimeout(ctx, mailboxEntCalendarPartTimeout)
	defer cancel()

	before, err := source.readCalendarPartSnapshot(operationContext, request)
	if err != nil {
		return zero, calendarPartSourceErrorOr(err, CalendarPartSourceErrorRepositoryUnavailable)
	}

	engineResult, err := source.engine.ReadCalendarPart(operationContext, EngineCalendarPartReadRequest{
		AccountID:   before.engineAccountID,
		MessageID:   before.engineMessageID,
		Version:     before.engineVersion,
		LocatorHash: before.locatorHash,
		Method:      before.method,
		MaxBytes:    request.maxBytes,
	})
	if err != nil {
		if isCalendarPartEngineConflict(err) {
			after, snapshotErr := source.readCalendarPartSnapshot(operationContext, request)
			if snapshotErr != nil {
				return zero, calendarPartSourceErrorOr(snapshotErr, CalendarPartSourceErrorRepositoryUnavailable)
			}
			if before != after {
				return zero, newCalendarPartSourceError(CalendarPartSourceErrorRepositoryUnavailable)
			}
			return zero, newCalendarPartSourceError(CalendarPartSourceErrorInvalid)
		}
		return zero, mapCalendarPartEngineError(err)
	}

	if calendarPartLocatorHashZero(engineResult.LocatorHash) ||
		engineResult.LocatorHash != request.locatorHash ||
		engineResult.LocatorHash != before.locatorHash {
		return zero, newCalendarPartSourceError(CalendarPartSourceErrorInvalid)
	}
	if len(engineResult.ITIP) == 0 {
		return zero, newCalendarPartSourceError(CalendarPartSourceErrorInvalid)
	}
	if len(engineResult.ITIP) > request.maxBytes {
		return zero, newCalendarPartSourceError(CalendarPartSourceErrorPayloadTooLarge)
	}

	after, err := source.readCalendarPartSnapshot(operationContext, request)
	if err != nil {
		return zero, calendarPartSourceErrorOr(err, CalendarPartSourceErrorRepositoryUnavailable)
	}
	if before != after {
		return zero, newCalendarPartSourceError(CalendarPartSourceErrorRepositoryUnavailable)
	}

	result, err := newCalendarPartReadResult(
		before.ownerID,
		before.sender,
		before.method,
		before.authenticationClass,
		engineResult.ITIP,
		request.maxBytes,
	)
	if err != nil {
		return zero, calendarPartSourceErrorOr(err, CalendarPartSourceErrorInvalid)
	}
	return result, nil
}

func (source *MailboxEntCalendarPartSource) readCalendarPartSnapshot(ctx context.Context, request CalendarPartReadRequest) (mailboxEntCalendarPartSnapshot, error) {
	var snapshot mailboxEntCalendarPartSnapshot
	err := source.repository.withMailboxEntReadSnapshot(ctx, func(transaction *ent.Tx) error {
		message, err := transaction.MailMessage.
			Query().
			Where(
				mailmessage.IDEQ(request.messageID),
				mailmessage.MailAccountIDEQ(request.mailboxID),
			).
			WithAccount(func(query *ent.MailAccountQuery) {
				query.WithUser()
			}).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return newCalendarPartSourceError(CalendarPartSourceErrorMessageUnavailable)
			}
			return err
		}

		account := message.Edges.Account
		if account == nil || account.ID != request.mailboxID || account.OrganizationID != request.organizationID {
			return newCalendarPartSourceError(CalendarPartSourceErrorMessageUnavailable)
		}
		if message.DeletedAt != nil || string(account.Status) == "deleting" {
			return newCalendarPartSourceError(CalendarPartSourceErrorDeleting)
		}

		owner := account.Edges.User
		if owner == nil ||
			owner.ID != account.UserID ||
			string(owner.Status) != "active" ||
			string(account.Status) != "active" ||
			!account.DesiredEnabled ||
			string(account.ProjectionState) != "ready" ||
			account.EngineAccountID == nil ||
			*account.EngineAccountID == "" ||
			message.EngineMessageID == "" ||
			message.EngineVersion == "" {
			return newCalendarPartSourceError(CalendarPartSourceErrorMessageUnavailable)
		}
		if message.CalendarMethod == nil ||
			len(message.CalendarPartLocatorHash) != sha256.Size {
			return newCalendarPartSourceError(CalendarPartSourceErrorPartUnavailable)
		}

		method := MailCalendarMethod(string(*message.CalendarMethod))
		if !validCalendarPartMethod(method) {
			return newCalendarPartSourceError(CalendarPartSourceErrorPartUnavailable)
		}
		authenticationClass := MailAuthenticationClass(string(message.AuthenticationClass))
		if !trustedCalendarPartAuthenticationClass(authenticationClass) {
			return newCalendarPartSourceError(CalendarPartSourceErrorUntrusted)
		}
		sender, ok := normalizeCalendarPartSender(message.From.Email)
		if !ok {
			return newCalendarPartSourceError(CalendarPartSourceErrorPartUnavailable)
		}

		var locatorHash [sha256.Size]byte
		copy(locatorHash[:], message.CalendarPartLocatorHash)
		if calendarPartLocatorHashZero(locatorHash) || locatorHash != request.locatorHash {
			return newCalendarPartSourceError(CalendarPartSourceErrorPartUnavailable)
		}

		snapshot = mailboxEntCalendarPartSnapshot{
			organizationID:      account.OrganizationID,
			mailboxID:           account.ID,
			ownerID:             owner.ID,
			userStatus:          string(owner.Status),
			accountStatus:       string(account.Status),
			desiredEnabled:      account.DesiredEnabled,
			projectionState:     string(account.ProjectionState),
			engineAccountID:     *account.EngineAccountID,
			engineMessageID:     message.EngineMessageID,
			engineVersion:       message.EngineVersion,
			messageLive:         message.DeletedAt == nil,
			sender:              sender,
			method:              method,
			authenticationClass: authenticationClass,
			locatorHash:         locatorHash,
		}
		return nil
	})
	if err != nil {
		return mailboxEntCalendarPartSnapshot{}, err
	}
	return snapshot, nil
}

func mapCalendarPartEngineError(err error) error {
	var engineError *EngineError
	if !errors.As(err, &engineError) || engineError == nil {
		return newCalendarPartSourceError(CalendarPartSourceErrorEngineUnavailable)
	}

	switch engineError.Kind {
	case EngineErrorPayloadTooLarge:
		return newCalendarPartSourceError(CalendarPartSourceErrorPayloadTooLarge)
	case EngineErrorNotFound:
		return newCalendarPartSourceError(CalendarPartSourceErrorPartUnavailable)
	case EngineErrorConflict,
		EngineErrorRejectedPolicy,
		EngineErrorUnauthorizedAccount,
		EngineErrorIndeterminate,
		EngineErrorRestartRequired,
		EngineErrorInvalidCursor:
		return newCalendarPartSourceError(CalendarPartSourceErrorInvalid)
	case EngineErrorUnavailable,
		EngineErrorRateLimited,
		EngineErrorIdentityUnavailable,
		EngineErrorRecipientLimit:
		return newCalendarPartSourceError(CalendarPartSourceErrorEngineUnavailable)
	default:
		return newCalendarPartSourceError(CalendarPartSourceErrorEngineUnavailable)
	}
}

func isCalendarPartEngineConflict(err error) bool {
	var engineError *EngineError
	return errors.As(err, &engineError) && engineError != nil && engineError.Kind == EngineErrorConflict
}

func calendarPartSourceErrorOr(err error, fallback CalendarPartSourceErrorKind) error {
	var sourceError *CalendarPartSourceError
	if errors.As(err, &sourceError) {
		return sourceError
	}
	return newCalendarPartSourceError(fallback)
}
