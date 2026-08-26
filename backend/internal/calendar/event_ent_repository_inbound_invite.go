package calendar

import (
	"context"
	"crypto/sha256"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarinboundmessagereceipt"
	"github.com/valerii/hub/backend/ent/calendarinboundobjectstate"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/ical"
)

func (backend *eventEntTransactionBackend) ResolveCalendarInboundInvite(
	ctx context.Context,
	messageScope CalendarInboundMessageScope,
	identity ITIPIdentity,
) (CalendarInboundInviteAuthorization, error) {
	if ctx == nil ||
		backend == nil ||
		backend.tx == nil ||
		messageScope.transaction == nil ||
		messageScope.transaction.backend != backend ||
		!messageScope.transaction.active ||
		!messageScope.claim.valid() ||
		!validCalendarInboundInviteIdentity(identity) {
		return CalendarInboundInviteAuthorization{}, ErrEventInvalid
	}

	receipt, err := backend.tx.CalendarInboundMessageReceipt.Query().
		Where(calendarinboundmessagereceipt.IDEQ(messageScope.claim.ReceiptID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return CalendarInboundInviteAuthorization{}, ErrEventInvalid
		}
		return CalendarInboundInviteAuthorization{}, calendarEntSafeError(err)
	}
	if !calendarInboundInviteReceiptReady(receipt, messageScope.claim) {
		return CalendarInboundInviteAuthorization{}, ErrEventInvalid
	}

	candidateAccount, err := backend.tx.MailAccount.Query().
		Where(
			mailaccount.IDEQ(receipt.MailboxID),
			mailaccount.OrganizationIDEQ(receipt.OrganizationID),
		).
		Only(ctx)
	if err != nil {
		return CalendarInboundInviteAuthorization{}, calendarEntAuthorizationError(err)
	}
	if candidateAccount.ID != receipt.MailboxID ||
		candidateAccount.OrganizationID != receipt.OrganizationID ||
		candidateAccount.UserID == uuid.Nil {
		return CalendarInboundInviteAuthorization{}, ErrEventForbidden
	}

	clock, err := calendarInboundInviteDatabaseClock(ctx, backend.tx)
	if err != nil {
		return CalendarInboundInviteAuthorization{}, err
	}
	if !receipt.LeaseUntil.After(clock) {
		return CalendarInboundInviteAuthorization{}, ErrEventInvalid
	}

	lockedOrganization, err := backend.tx.Organization.Query().
		Where(
			organization.IDEQ(receipt.OrganizationID),
			organization.HasUsersWith(
				user.IDEQ(candidateAccount.UserID),
				user.StatusEQ(user.StatusActive),
			),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return CalendarInboundInviteAuthorization{}, calendarEntAuthorizationError(err)
	}
	if lockedOrganization.ID != receipt.OrganizationID {
		return CalendarInboundInviteAuthorization{}, ErrEventForbidden
	}

	lockedUser, err := backend.tx.User.Query().
		Where(
			user.IDEQ(candidateAccount.UserID),
			user.StatusEQ(user.StatusActive),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return CalendarInboundInviteAuthorization{}, calendarEntAuthorizationError(err)
	}
	if lockedUser.ID != candidateAccount.UserID {
		return CalendarInboundInviteAuthorization{}, ErrEventForbidden
	}

	lockedAccount, err := backend.tx.MailAccount.Query().
		Where(
			mailaccount.IDEQ(receipt.MailboxID),
			mailaccount.OrganizationIDEQ(receipt.OrganizationID),
			mailaccount.UserIDEQ(candidateAccount.UserID),
			mailaccount.StatusEQ(mailaccount.Status("active")),
			mailaccount.DesiredEnabledEQ(true),
			mailaccount.ProjectionStateEQ(mailaccount.ProjectionState("ready")),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return CalendarInboundInviteAuthorization{}, calendarEntAuthorizationError(err)
	}
	if !calendarInboundInviteMailAccountReady(lockedAccount, receipt, candidateAccount.UserID) {
		return CalendarInboundInviteAuthorization{}, ErrEventForbidden
	}

	seriesKey, err := NewCalendarInboundObjectKey(candidateAccount.UserID, identity.UID, nil)
	if err != nil {
		return CalendarInboundInviteAuthorization{}, ErrEventInvalid
	}
	series, err := calendarInboundInviteLockObjectState(
		ctx,
		backend.tx,
		receipt.OrganizationID,
		candidateAccount.UserID,
		identity.UID,
		nil,
		seriesKey,
	)
	if err != nil {
		return CalendarInboundInviteAuthorization{}, err
	}

	var current *CalendarInboundObjectState
	if identity.HasRecurrenceID {
		recurrenceID := identity.RecurrenceID
		currentKey, err := NewCalendarInboundObjectKey(candidateAccount.UserID, identity.UID, &recurrenceID)
		if err != nil {
			return CalendarInboundInviteAuthorization{}, ErrEventInvalid
		}
		current, err = calendarInboundInviteLockObjectState(
			ctx,
			backend.tx,
			receipt.OrganizationID,
			candidateAccount.UserID,
			identity.UID,
			&recurrenceID,
			currentKey,
		)
		if err != nil {
			return CalendarInboundInviteAuthorization{}, err
		}
	} else {
		current = series
		series = nil
	}

	if current != nil || series != nil {
		authorization, err := NewCalendarInboundInviteAuthorization(
			candidateAccount.UserID,
			current,
			series,
			false,
		)
		if err != nil {
			return CalendarInboundInviteAuthorization{}, ErrEventInvalid
		}
		return authorization, nil
	}

	knownUID, err := backend.tx.CalendarInboundObjectState.Query().
		Where(
			calendarinboundobjectstate.CalendarOwnerIDEQ(candidateAccount.UserID),
			calendarinboundobjectstate.UIDEQ(identity.UID),
		).
		Exist(ctx)
	if err != nil {
		return CalendarInboundInviteAuthorization{}, calendarEntSafeError(err)
	}
	if knownUID {
		authorization, err := NewCalendarInboundInviteAuthorization(candidateAccount.UserID, nil, nil, false)
		if err != nil {
			return CalendarInboundInviteAuthorization{}, ErrEventInvalid
		}
		return authorization, nil
	}

	unmaterializedIDs, err := backend.tx.CalendarInboundObjectState.Query().
		Where(
			calendarinboundobjectstate.CalendarOwnerIDEQ(candidateAccount.UserID),
			calendarinboundobjectstate.CalendarEventIDIsNil(),
		).
		Limit(10000).
		IDs(ctx)
	if err != nil {
		return CalendarInboundInviteAuthorization{}, calendarEntSafeError(err)
	}

	authorization, err := NewCalendarInboundInviteAuthorization(
		candidateAccount.UserID,
		nil,
		nil,
		len(unmaterializedIDs) >= 10000,
	)
	if err != nil {
		return CalendarInboundInviteAuthorization{}, ErrEventInvalid
	}
	return authorization, nil
}

func calendarInboundInviteDatabaseClock(ctx context.Context, transaction *ent.Tx) (time.Time, error) {
	rows := &entsql.Rows{}
	if err := transaction.Driver().Query(
		ctx,
		`SELECT date_trunc('second', clock_timestamp())`,
		[]any{},
		rows,
	); err != nil {
		return time.Time{}, calendarEntSafeError(err)
	}

	var clock time.Time
	scanErr := entsql.ScanOne(rows, &clock)
	closeErr := rows.Close()
	if scanErr != nil {
		return time.Time{}, calendarEntSafeError(scanErr)
	}
	if closeErr != nil {
		return time.Time{}, calendarEntSafeError(closeErr)
	}
	clock = clock.UTC()
	if !calendarInboundInviteWholeSecondUTC(clock) {
		return time.Time{}, ErrEventInvalid
	}
	return clock, nil
}

func calendarInboundInviteReceiptReady(
	receipt *ent.CalendarInboundMessageReceipt,
	claim CalendarInboundMessageClaim,
) bool {
	return receipt != nil &&
		receipt.ID == claim.ReceiptID &&
		receipt.OrganizationID != uuid.Nil &&
		receipt.MailboxID != uuid.Nil &&
		string(receipt.State) == "processing" &&
		receipt.Attempts >= 1 &&
		receipt.Attempts <= 8 &&
		receipt.LeaseToken != nil &&
		*receipt.LeaseToken == claim.LeaseToken &&
		receipt.LeaseGeneration == claim.LeaseGeneration &&
		receipt.LeaseUntil != nil &&
		calendarInboundInviteWholeSecondUTC(*receipt.LeaseUntil) &&
		receipt.NextAttemptAt == nil &&
		receipt.ErrorCode == nil
}

func calendarInboundInviteMailAccountReady(
	account *ent.MailAccount,
	receipt *ent.CalendarInboundMessageReceipt,
	ownerID uuid.UUID,
) bool {
	if account == nil ||
		receipt == nil ||
		account.ID == uuid.Nil ||
		account.OrganizationID == uuid.Nil ||
		account.UserID == uuid.Nil ||
		account.ID != receipt.MailboxID ||
		account.OrganizationID != receipt.OrganizationID ||
		account.UserID != ownerID ||
		string(account.Status) != "active" ||
		!account.DesiredEnabled ||
		string(account.ProjectionState) != "ready" ||
		account.EngineAccountID == nil ||
		*account.EngineAccountID == "" ||
		account.DeployedGeneration == nil ||
		account.DeployedHashVersion == nil ||
		len(account.DeployedHash) != sha256.Size ||
		*account.DeployedGeneration < 1 ||
		*account.DeployedHashVersion < 1 ||
		*account.DeployedGeneration > account.ControlGeneration ||
		account.SyncCursor == nil ||
		*account.SyncCursor == "" ||
		account.LastSyncedAt == nil ||
		!calendarInboundInviteWholeSecondUTC(*account.LastSyncedAt) ||
		account.RebuildCheckpoint != nil ||
		account.RebuildBaseCursor != nil ||
		account.SyncErrorCode != nil {
		return false
	}
	return true
}

func calendarInboundInviteLockObjectState(
	ctx context.Context,
	transaction *ent.Tx,
	organizationID uuid.UUID,
	ownerID uuid.UUID,
	uid string,
	recurrenceID *string,
	key CalendarInboundObjectKey,
) (*CalendarInboundObjectState, error) {
	query := transaction.CalendarInboundObjectState.Query().
		Where(
			calendarinboundobjectstate.CalendarOwnerIDEQ(ownerID),
			calendarinboundobjectstate.UIDEQ(uid),
		).
		Order(ent.Asc(calendarinboundobjectstate.FieldID)).
		ForUpdate().
		Limit(2)
	if recurrenceID == nil {
		query = query.Where(calendarinboundobjectstate.RecurrenceIDIsNil())
	} else {
		query = query.Where(calendarinboundobjectstate.RecurrenceIDEQ(*recurrenceID))
	}

	rows, err := query.All(ctx)
	if err != nil {
		return nil, calendarEntSafeError(err)
	}
	if len(rows) > 1 {
		return nil, ErrEventInvalid
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return calendarInboundInviteMapObjectState(rows[0], organizationID, ownerID, uid, recurrenceID, key)
}

func calendarInboundInviteMapObjectState(
	row *ent.CalendarInboundObjectState,
	organizationID uuid.UUID,
	ownerID uuid.UUID,
	uid string,
	recurrenceID *string,
	key CalendarInboundObjectKey,
) (*CalendarInboundObjectState, error) {
	if row == nil ||
		row.ID == uuid.Nil ||
		row.OrganizationID != organizationID ||
		row.CalendarOwnerID == uuid.Nil ||
		row.CalendarOwnerID != ownerID ||
		row.UID != uid ||
		row.Sequence < 0 ||
		row.Sequence > ical.MaxITIPSequence ||
		row.PayloadHashVersion == 0 ||
		len(row.PayloadHash) != sha256.Size {
		return nil, ErrEventInvalid
	}
	if recurrenceID == nil {
		if row.RecurrenceID != nil {
			return nil, ErrEventInvalid
		}
	} else if row.RecurrenceID == nil || *row.RecurrenceID != *recurrenceID {
		return nil, ErrEventInvalid
	}
	if row.CalendarEventID != nil && *row.CalendarEventID == uuid.Nil {
		return nil, ErrEventInvalid
	}

	method := ITIPMethod(row.Method)
	if method != ITIPMethodRequest && method != ITIPMethodCancel {
		return nil, ErrEventInvalid
	}
	status := CalendarInboundObjectStatus(row.State)
	switch status {
	case CalendarInboundObjectApplied, CalendarInboundObjectTombstone:
		if row.ExpiresAt != nil {
			return nil, ErrEventInvalid
		}
	case CalendarInboundObjectDeferredMissingMaster:
		if recurrenceID == nil ||
			method != ITIPMethodRequest ||
			row.CalendarEventID != nil ||
			row.ExpiresAt == nil ||
			row.ExpiresAt.IsZero() ||
			!calendarInboundInviteWholeSecondUTC(*row.ExpiresAt) {
			return nil, ErrEventInvalid
		}
	default:
		return nil, ErrEventInvalid
	}

	var payloadHash [sha256.Size]byte
	copy(payloadHash[:], row.PayloadHash)
	payload, err := NewCalendarInboundCanonicalPayload(row.PayloadHashVersion, payloadHash)
	if err != nil {
		return nil, ErrEventInvalid
	}
	state, err := NewCalendarInboundObjectState(
		key,
		row.OrganizerEmail,
		uint32(row.Sequence),
		method,
		payload,
		status,
		row.CalendarEventID,
	)
	if err != nil {
		return nil, ErrEventInvalid
	}
	return &state, nil
}

func calendarInboundInviteWholeSecondUTC(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond() == 0
}
