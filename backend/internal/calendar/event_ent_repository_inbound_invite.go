package calendar

import (
	"bytes"
	"context"
	"crypto/sha256"
	"sort"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendareventexception"
	"github.com/valerii/hub/backend/ent/calendarinboundmessagereceipt"
	"github.com/valerii/hub/backend/ent/calendarinboundobjectstate"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/ical"
	"github.com/valerii/hub/backend/internal/platform/events"
	"github.com/valerii/hub/backend/internal/platform/outbox"
)

var _ CalendarInboundInviteTransactionBackend = (*eventEntTransactionBackend)(nil)

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

	clock, err := backend.calendarInboundMessageDatabaseClock(ctx)
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

// PersistCalendarInboundInvite changes only Calendar state for one already
// claimed inbound REQUEST or CANCEL. Receipt completion stays with the caller
// and is intentionally the final operation in the same EventTransaction.
func (backend *eventEntTransactionBackend) PersistCalendarInboundInvite(
	ctx context.Context,
	scope CalendarInboundInviteScope,
	command CalendarInboundInvitePersistenceCommand,
) (CalendarInboundInvitePersistenceResult, error) {
	if ctx == nil ||
		backend == nil ||
		backend.tx == nil ||
		scope.transaction == nil ||
		scope.transaction.backend != backend ||
		!scope.transaction.active ||
		!scope.messageScope.claim.valid() ||
		!scope.authorization.valid() ||
		!command.valid() {
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}

	durable, err := backend.revalidateCalendarInboundInvite(ctx, scope.messageScope.claim, command)
	if err != nil {
		return CalendarInboundInvitePersistenceResult{}, err
	}
	if scope.authorization.ownerID != durable.ownerID ||
		scope.authorization.resourceLimited ||
		durable.resourceLimited ||
		!calendarInboundInviteStatesEqual(scope.authorization.current, durable.current) ||
		!calendarInboundInviteStatesEqual(scope.authorization.series, durable.series) ||
		!calendarInboundInviteStatesEqual(command.input.current, durable.current) ||
		!calendarInboundInviteStatesEqual(command.input.series, durable.series) {
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}

	recheckedInput, err := NewCalendarInboundIngestInput(
		command.input.incoming,
		durable.current,
		durable.series,
	)
	if err != nil || !calendarInboundInviteInputsEqual(command.input, recheckedInput) {
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}
	recheckedDecision, err := DecideCalendarInboundITIP(recheckedInput)
	if err != nil || !calendarInboundInviteDecisionsEqual(command.decision, recheckedDecision) {
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}

	incoming := command.input.incoming
	component := incoming.Component()
	nextState, mutates := command.decision.NextState()
	if !mutates || nextState == "" || !calendarInboundAuthenticationAdmitted(incoming.Authentication()) {
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}

	var (
		eventID         *uuid.UUID
		attendeeIDs     []uuid.UUID
		materialChanged bool
	)
	switch command.Outcome() {
	case CalendarInboundIngestApplyRequest:
		var event *ent.CalendarEvent
		if component.IsSeries() {
			event, attendeeIDs, err = backend.persistCalendarInboundInviteSeriesRequest(ctx, durable, command)
		} else {
			event, attendeeIDs, err = backend.persistCalendarInboundInviteOccurrenceRequest(ctx, durable, command)
		}
		if err != nil {
			return CalendarInboundInvitePersistenceResult{}, err
		}
		if event == nil || event.ID == uuid.Nil {
			return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
		}
		value := event.ID
		eventID = &value
		materialChanged = true
	case CalendarInboundIngestTombstoneSeries:
		var event *ent.CalendarEvent
		event, attendeeIDs, materialChanged, err = backend.persistCalendarInboundInviteSeriesCancel(ctx, durable, command)
		if err != nil {
			return CalendarInboundInvitePersistenceResult{}, err
		}
		if event != nil {
			value := event.ID
			eventID = &value
		}
	case CalendarInboundIngestTombstoneOccurrence:
		var event *ent.CalendarEvent
		event, attendeeIDs, materialChanged, err = backend.persistCalendarInboundInviteOccurrenceCancel(ctx, durable, command)
		if err != nil {
			return CalendarInboundInvitePersistenceResult{}, err
		}
		if event != nil {
			value := event.ID
			eventID = &value
		}
	case CalendarInboundIngestDeferred:
		if component.IsSeries() {
			return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
		}
	default:
		return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
	}

	var expiresAt *time.Time
	if command.Outcome() == CalendarInboundIngestDeferred {
		expiresAt, err = calendarInboundInviteDeferredExpiry(durable.currentRow, durable.clock)
		if err != nil {
			return CalendarInboundInvitePersistenceResult{}, err
		}
	}
	stateID, err := backend.persistCalendarInboundInviteObjectState(
		ctx,
		durable.currentRow,
		durable.organizationID,
		durable.ownerID,
		incoming,
		nextState,
		eventID,
		expiresAt,
	)
	if err != nil {
		return CalendarInboundInvitePersistenceResult{}, err
	}

	if materialChanged {
		if eventID == nil || command.OutboxEnvelopeID() == uuid.Nil {
			return CalendarInboundInvitePersistenceResult{}, ErrEventInvalid
		}
		if err := backend.appendCalendarInboundInviteOutbox(
			ctx,
			command,
			durable.clock,
			durable.organizationID,
			*eventID,
			attendeeIDs,
		); err != nil {
			return CalendarInboundInvitePersistenceResult{}, err
		}
	}

	return NewCalendarInboundInvitePersistenceResult(
		command.Outcome(),
		stateID,
		eventID,
		materialChanged,
		materialChanged,
	)
}

type calendarInboundInviteDurableScope struct {
	organizationID         uuid.UUID
	ownerID                uuid.UUID
	ownerParticipantEmail string
	clock                  time.Time
	currentRow             *ent.CalendarInboundObjectState
	current                *CalendarInboundObjectState
	series                 *CalendarInboundObjectState
	resourceLimited        bool
}

func (backend *eventEntTransactionBackend) revalidateCalendarInboundInvite(
	ctx context.Context,
	claim CalendarInboundMessageClaim,
	command CalendarInboundInvitePersistenceCommand,
) (calendarInboundInviteDurableScope, error) {
	if backend == nil || backend.tx == nil || !claim.valid() || !command.valid() {
		return calendarInboundInviteDurableScope{}, ErrEventInvalid
	}

	receipt, err := backend.tx.CalendarInboundMessageReceipt.Query().
		Where(calendarinboundmessagereceipt.IDEQ(claim.ReceiptID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return calendarInboundInviteDurableScope{}, ErrEventInvalid
		}
		return calendarInboundInviteDurableScope{}, calendarEntSafeError(err)
	}
	if !calendarInboundInviteReceiptReady(receipt, claim) {
		return calendarInboundInviteDurableScope{}, ErrEventInvalid
	}
	clock, err := backend.calendarInboundMessageDatabaseClock(ctx)
	if err != nil {
		return calendarInboundInviteDurableScope{}, err
	}
	if !receipt.LeaseUntil.After(clock) {
		return calendarInboundInviteDurableScope{}, ErrEventInvalid
	}

	candidateAccount, err := backend.tx.MailAccount.Query().
		Where(
			mailaccount.IDEQ(receipt.MailboxID),
			mailaccount.OrganizationIDEQ(receipt.OrganizationID),
		).
		Only(ctx)
	if err != nil {
		return calendarInboundInviteDurableScope{}, calendarEntAuthorizationError(err)
	}
	if candidateAccount.ID != receipt.MailboxID ||
		candidateAccount.OrganizationID != receipt.OrganizationID ||
		candidateAccount.UserID == uuid.Nil {
		return calendarInboundInviteDurableScope{}, ErrEventForbidden
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
		return calendarInboundInviteDurableScope{}, calendarEntAuthorizationError(err)
	}
	if lockedOrganization.ID != receipt.OrganizationID {
		return calendarInboundInviteDurableScope{}, ErrEventForbidden
	}

	lockedUser, err := backend.tx.User.Query().
		Where(
			user.IDEQ(candidateAccount.UserID),
			user.StatusEQ(user.StatusActive),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return calendarInboundInviteDurableScope{}, calendarEntAuthorizationError(err)
	}
	if lockedUser.ID != candidateAccount.UserID {
		return calendarInboundInviteDurableScope{}, ErrEventForbidden
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
		return calendarInboundInviteDurableScope{}, calendarEntAuthorizationError(err)
	}
	if !calendarInboundInviteMailAccountReady(lockedAccount, receipt, candidateAccount.UserID) {
		return calendarInboundInviteDurableScope{}, ErrEventForbidden
	}
	ownerParticipantEmail, err := backend.lockCalendarInboundInviteOwnerParticipantEmail(
		ctx,
		lockedAccount.ID,
		command.request.message.Attendees,
	)
	if err != nil {
		return calendarInboundInviteDurableScope{}, err
	}

	component := command.input.incoming.Component()
	seriesKey, err := NewCalendarInboundObjectKey(candidateAccount.UserID, component.UID(), nil)
	if err != nil {
		return calendarInboundInviteDurableScope{}, ErrEventInvalid
	}
	seriesRow, err := calendarInboundInviteLockObjectStateRow(
		ctx,
		backend.tx,
		candidateAccount.UserID,
		component.UID(),
		nil,
	)
	if err != nil {
		return calendarInboundInviteDurableScope{}, err
	}
	var series *CalendarInboundObjectState
	if seriesRow != nil {
		series, err = calendarInboundInviteMapObjectState(
			seriesRow,
			receipt.OrganizationID,
			candidateAccount.UserID,
			component.UID(),
			nil,
			seriesKey,
		)
		if err != nil {
			return calendarInboundInviteDurableScope{}, err
		}
	}

	currentRow := seriesRow
	current := series
	if !component.IsSeries() {
		recurrenceID := component.RecurrenceID()
		if recurrenceID == nil {
			return calendarInboundInviteDurableScope{}, ErrEventInvalid
		}
		currentKey, keyErr := NewCalendarInboundObjectKey(candidateAccount.UserID, component.UID(), recurrenceID)
		if keyErr != nil {
			return calendarInboundInviteDurableScope{}, ErrEventInvalid
		}
		currentRow, err = calendarInboundInviteLockObjectStateRow(
			ctx,
			backend.tx,
			candidateAccount.UserID,
			component.UID(),
			recurrenceID,
		)
		if err != nil {
			return calendarInboundInviteDurableScope{}, err
		}
		current = nil
		if currentRow != nil {
			current, err = calendarInboundInviteMapObjectState(
				currentRow,
				receipt.OrganizationID,
				candidateAccount.UserID,
				component.UID(),
				recurrenceID,
				currentKey,
			)
			if err != nil {
				return calendarInboundInviteDurableScope{}, err
			}
		}
	} else {
		series = nil
	}

	resourceLimited := false
	if current == nil && series == nil {
		knownUID, knownErr := backend.tx.CalendarInboundObjectState.Query().
			Where(
				calendarinboundobjectstate.CalendarOwnerIDEQ(candidateAccount.UserID),
				calendarinboundobjectstate.UIDEQ(component.UID()),
			).
			Exist(ctx)
		if knownErr != nil {
			return calendarInboundInviteDurableScope{}, calendarEntSafeError(knownErr)
		}
		if !knownUID {
			unmaterializedIDs, limitErr := backend.tx.CalendarInboundObjectState.Query().
				Where(
					calendarinboundobjectstate.CalendarOwnerIDEQ(candidateAccount.UserID),
					calendarinboundobjectstate.CalendarEventIDIsNil(),
				).
				Limit(10000).
				IDs(ctx)
			if limitErr != nil {
				return calendarInboundInviteDurableScope{}, calendarEntSafeError(limitErr)
			}
			resourceLimited = len(unmaterializedIDs) >= 10000
		}
	}

	return calendarInboundInviteDurableScope{
		organizationID:         receipt.OrganizationID,
		ownerID:                candidateAccount.UserID,
		ownerParticipantEmail: ownerParticipantEmail,
		clock:                  clock,
		currentRow:             currentRow,
		current:                current,
		series:                 series,
		resourceLimited:        resourceLimited,
	}, nil
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
		account.DeployedHash == nil ||
		len(*account.DeployedHash) != sha256.Size ||
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

func (backend *eventEntTransactionBackend) lockCalendarInboundInviteOwnerParticipantEmail(
	ctx context.Context,
	mailAccountID uuid.UUID,
	attendees []ITIPParticipant,
) (string, error) {
	if backend == nil || backend.tx == nil || mailAccountID == uuid.Nil {
		return "", ErrEventInvalid
	}
	participants, err := calendarInboundInviteParticipants(attendees)
	if err != nil {
		return "", err
	}
	rows, err := backend.tx.MailIdentity.Query().
		Where(
			mailidentity.MailAccountIDEQ(mailAccountID),
			mailidentity.StatusEQ(mailidentity.StatusActive),
			mailidentity.KindIn(mailidentity.KindPrimary, mailidentity.KindAlias),
		).
		Order(
			ent.Asc(mailidentity.FieldNormalizedAddress),
			ent.Asc(mailidentity.FieldID),
		).
		ForUpdate().
		All(ctx)
	if err != nil {
		return "", calendarEntSafeError(err)
	}
	authoritative := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil || row.MailAccountID != mailAccountID ||
			row.Status != mailidentity.StatusActive ||
			row.Kind != mailidentity.KindPrimary && row.Kind != mailidentity.KindAlias ||
			row.NormalizedAddress == "" || normalizeCalendarCreateEmail(row.NormalizedAddress) != row.NormalizedAddress {
			return "", ErrEventInvalid
		}
		if _, duplicate := authoritative[row.NormalizedAddress]; duplicate {
			return "", ErrEventInvalid
		}
		authoritative[row.NormalizedAddress] = struct{}{}
	}

	ownerParticipantEmail := ""
	for _, participant := range participants {
		if _, matches := authoritative[participant.email]; !matches {
			continue
		}
		if ownerParticipantEmail != "" {
			return "", ErrEventInvalid
		}
		ownerParticipantEmail = participant.email
	}
	if ownerParticipantEmail == "" {
		return "", ErrEventInvalid
	}
	return ownerParticipantEmail, nil
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
	row, err := calendarInboundInviteLockObjectStateRow(ctx, transaction, ownerID, uid, recurrenceID)
	if err != nil || row == nil {
		return nil, err
	}
	return calendarInboundInviteMapObjectState(row, organizationID, ownerID, uid, recurrenceID, key)
}

func calendarInboundInviteLockObjectStateRow(
	ctx context.Context,
	transaction *ent.Tx,
	ownerID uuid.UUID,
	uid string,
	recurrenceID *string,
) (*ent.CalendarInboundObjectState, error) {
	if transaction == nil || ownerID == uuid.Nil || uid == "" {
		return nil, ErrEventInvalid
	}
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
	return rows[0], nil
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

	method, err := calendarInboundInviteITIPMethod(row.Method)
	if err != nil {
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

func calendarInboundInviteITIPMethod(value calendarinboundobjectstate.Method) (ITIPMethod, error) {
	switch value {
	case calendarinboundobjectstate.Method("request"):
		return ITIPMethodRequest, nil
	case calendarinboundobjectstate.Method("cancel"):
		return ITIPMethodCancel, nil
	default:
		return "", ErrEventInvalid
	}
}

func calendarInboundInviteObjectStateMethod(value ITIPMethod) (calendarinboundobjectstate.Method, error) {
	switch value {
	case ITIPMethodRequest:
		return calendarinboundobjectstate.Method("request"), nil
	case ITIPMethodCancel:
		return calendarinboundobjectstate.Method("cancel"), nil
	default:
		return "", ErrEventInvalid
	}
}

func calendarInboundInviteStatesEqual(left, right *CalendarInboundObjectState) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	if !left.valid() || !right.valid() {
		return false
	}
	leftComponent := left.Component()
	rightComponent := right.Component()
	if !leftComponent.equal(rightComponent) ||
		left.OrganizerEmail() != right.OrganizerEmail() ||
		left.Sequence() != right.Sequence() ||
		left.Method() != right.Method() ||
		left.Payload() != right.Payload() ||
		left.Status() != right.Status() {
		return false
	}
	leftEventID := left.EventID()
	rightEventID := right.EventID()
	return leftEventID == nil && rightEventID == nil ||
		leftEventID != nil && rightEventID != nil && *leftEventID == *rightEventID
}

func calendarInboundInviteInputsEqual(left, right CalendarInboundIngestInput) bool {
	if !left.valid() || !right.valid() {
		return false
	}
	leftIncoming := left.incoming
	rightIncoming := right.incoming
	leftComponent := leftIncoming.Component()
	rightComponent := rightIncoming.Component()
	return leftComponent.equal(rightComponent) &&
		leftIncoming.OrganizerEmail() == rightIncoming.OrganizerEmail() &&
		leftIncoming.Authentication() == rightIncoming.Authentication() &&
		leftIncoming.Method() == rightIncoming.Method() &&
		leftIncoming.Sequence() == rightIncoming.Sequence() &&
		leftIncoming.Payload() == rightIncoming.Payload() &&
		calendarInboundInviteStatesEqual(left.current, right.current) &&
		calendarInboundInviteStatesEqual(left.series, right.series)
}

func calendarInboundInviteDecisionsEqual(left, right CalendarInboundIngestResult) bool {
	leftState, leftMutates := left.NextState()
	rightState, rightMutates := right.NextState()
	return left.Outcome() == right.Outcome() &&
		leftMutates == rightMutates &&
		leftState == rightState
}

func calendarInboundInviteDeferredExpiry(
	current *ent.CalendarInboundObjectState,
	clock time.Time,
) (*time.Time, error) {
	if !calendarInboundInviteWholeSecondUTC(clock) {
		return nil, ErrEventInvalid
	}
	if current != nil && string(current.State) == string(CalendarInboundObjectDeferredMissingMaster) {
		if current.ExpiresAt == nil || !calendarInboundInviteWholeSecondUTC(*current.ExpiresAt) {
			return nil, ErrEventInvalid
		}
		return eventEntTimePointer(current.ExpiresAt), nil
	}
	expiresAt := clock.Add(7 * 24 * time.Hour)
	if !calendarInboundInviteWholeSecondUTC(expiresAt) {
		return nil, ErrEventInvalid
	}
	return &expiresAt, nil
}

func (backend *eventEntTransactionBackend) persistCalendarInboundInviteObjectState(
	ctx context.Context,
	current *ent.CalendarInboundObjectState,
	organizationID, ownerID uuid.UUID,
	incoming CalendarInboundITIP,
	status CalendarInboundObjectStatus,
	eventID *uuid.UUID,
	expiresAt *time.Time,
) (uuid.UUID, error) {
	if backend == nil || backend.tx == nil || organizationID == uuid.Nil || ownerID == uuid.Nil ||
		!incoming.valid() || !validCalendarInboundMessageUUIDPointer(eventID) {
		return uuid.Nil, ErrEventInvalid
	}
	component := incoming.Component()
	if component.OwnerID() != ownerID {
		return uuid.Nil, ErrEventInvalid
	}
	method, err := calendarInboundInviteObjectStateMethod(incoming.Method())
	if err != nil {
		return uuid.Nil, err
	}
	sequence, ok := eventEntInt(uint(incoming.Sequence()))
	if !ok {
		return uuid.Nil, ErrEventInvalid
	}
	payload := incoming.Payload()
	payloadHash := payload.Hash()
	if status == CalendarInboundObjectApplied && eventID == nil ||
		status == CalendarInboundObjectDeferredMissingMaster && (component.IsSeries() || eventID != nil || expiresAt == nil) ||
		status != CalendarInboundObjectDeferredMissingMaster && expiresAt != nil {
		return uuid.Nil, ErrEventInvalid
	}
	if expiresAt != nil && !calendarInboundInviteWholeSecondUTC(*expiresAt) {
		return uuid.Nil, ErrEventInvalid
	}

	if current == nil {
		create := backend.tx.CalendarInboundObjectState.Create().
			SetOrganizationID(organizationID).
			SetCalendarOwnerID(ownerID).
			SetUID(component.UID()).
			SetOrganizerEmail(incoming.OrganizerEmail()).
			SetSequence(sequence).
			SetMethod(method).
			SetPayloadHashVersion(payload.CanonicalVersion()).
			SetPayloadHash(payloadHash[:]).
			SetState(calendarinboundobjectstate.State(status))
		if recurrenceID := component.RecurrenceID(); recurrenceID != nil {
			create.SetRecurrenceID(*recurrenceID)
		}
		if eventID != nil {
			create.SetCalendarEventID(*eventID)
		}
		if expiresAt != nil {
			create.SetExpiresAt(*expiresAt)
		}
		saved, saveErr := create.Save(ctx)
		if saveErr != nil {
			return uuid.Nil, calendarEntSafeError(saveErr)
		}
		if saved == nil || saved.ID == uuid.Nil {
			return uuid.Nil, ErrEventInvalid
		}
		return saved.ID, nil
	}

	update := backend.tx.CalendarInboundObjectState.UpdateOne(current).
		SetSequence(sequence).
		SetMethod(method).
		SetPayloadHashVersion(payload.CanonicalVersion()).
		SetPayloadHash(payloadHash[:]).
		SetState(calendarinboundobjectstate.State(status))
	if eventID == nil {
		update.ClearCalendarEventID()
	} else {
		update.SetCalendarEventID(*eventID)
	}
	if expiresAt == nil {
		update.ClearExpiresAt()
	} else {
		update.SetExpiresAt(*expiresAt)
	}
	if err := update.Exec(ctx); err != nil {
		return uuid.Nil, calendarEntSafeError(err)
	}
	return current.ID, nil
}

type calendarInboundInviteEventProjection struct {
	temporal CalendarEvent
	status   CalendarEventStatus
	exdates  []string
}

func calendarInboundInviteProjection(message ITIPMessage) (calendarInboundInviteEventProjection, error) {
	if message.Summary == "" {
		return calendarInboundInviteEventProjection{}, ErrEventInvalid
	}
	projection := calendarInboundInviteEventProjection{}
	switch {
	case message.Temporal.Timed != nil && message.Temporal.AllDay == nil:
		projection.temporal = CalendarEvent{
			StartsAt: message.Temporal.Timed.StartsAt.UTC(),
			EndsAt:   message.Temporal.Timed.EndsAt.UTC(),
			Timezone: message.Temporal.Timed.Timezone,
		}
	case message.Temporal.Timed == nil && message.Temporal.AllDay != nil:
		projection.temporal = CalendarEvent{
			AllDay:           true,
			StartDate:        message.Temporal.AllDay.StartDate,
			EndDateExclusive: message.Temporal.AllDay.EndDateExclusive,
			Timezone:         message.Temporal.AllDay.Timezone,
		}
	default:
		return calendarInboundInviteEventProjection{}, ErrEventInvalid
	}
	if projection.temporal.Validate() != nil {
		return calendarInboundInviteEventProjection{}, ErrEventInvalid
	}
	switch message.Status {
	case ITIPStatusConfirmed:
		projection.status = CalendarEventStatusConfirmed
	case ITIPStatusTentative:
		projection.status = CalendarEventStatusTentative
	default:
		return calendarInboundInviteEventProjection{}, ErrEventInvalid
	}
	projection.exdates = make([]string, len(message.EXDates))
	for index, exdate := range message.EXDates {
		if projection.temporal.AllDay {
			projection.exdates[index] = exdate.Date
		} else {
			projection.exdates[index] = exdate.Time.UTC().Format(time.RFC3339)
		}
		if !validCalendarInboundRecurrenceID(projection.exdates[index]) {
			return calendarInboundInviteEventProjection{}, ErrEventInvalid
		}
	}
	return projection, nil
}

func calendarInboundInviteImportedEventMatches(
	event *ent.CalendarEvent,
	organizationID, ownerID uuid.UUID,
	uid, organizerEmail string,
) bool {
	return event != nil &&
		event.ID != uuid.Nil &&
		event.OrganizationID == organizationID &&
		event.CalendarOwnerID == ownerID &&
		event.UID == uid &&
		event.Source == calendarevent.Source(CalendarEventSourceMailImport) &&
		event.OrganizerEmail == organizerEmail
}

func (backend *eventEntTransactionBackend) lockCalendarInboundInviteEvent(
	ctx context.Context,
	organizationID, ownerID uuid.UUID,
	uid, organizerEmail string,
	eventID *uuid.UUID,
) (*ent.CalendarEvent, error) {
	if backend == nil || backend.tx == nil || organizationID == uuid.Nil || ownerID == uuid.Nil || uid == "" {
		return nil, ErrEventInvalid
	}
	query := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.OrganizationIDEQ(organizationID),
			calendarevent.CalendarOwnerIDEQ(ownerID),
			calendarevent.UIDEQ(uid),
		).
		Order(ent.Asc(calendarevent.FieldID)).
		ForUpdate().
		Limit(2)
	if eventID != nil {
		query = query.Where(calendarevent.IDEQ(*eventID))
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, calendarEntSafeError(err)
	}
	if len(rows) > 1 {
		return nil, ErrEventInvalid
	}
	if len(rows) == 0 {
		if eventID != nil {
			return nil, ErrEventInvalid
		}
		return nil, nil
	}
	if !calendarInboundInviteImportedEventMatches(rows[0], organizationID, ownerID, uid, organizerEmail) {
		return nil, ErrEventInvalid
	}
	return rows[0], nil
}

func (backend *eventEntTransactionBackend) createCalendarInboundInviteEvent(
	ctx context.Context,
	organizationID, ownerID uuid.UUID,
	incoming CalendarInboundITIP,
	message ITIPMessage,
) (*ent.CalendarEvent, error) {
	projection, err := calendarInboundInviteProjection(message)
	if err != nil {
		return nil, err
	}
	sequence, ok := eventEntInt(uint(incoming.Sequence()))
	if !ok {
		return nil, ErrEventInvalid
	}
	payload := incoming.Payload()
	payloadHash := payload.Hash()
	create := backend.tx.CalendarEvent.Create().
		SetOrganizationID(organizationID).
		SetCalendarOwnerID(ownerID).
		SetUID(incoming.Component().UID()).
		SetTitle(message.Summary).
		SetDescription(message.Description).
		SetLocation(message.Location).
		SetAllDay(projection.temporal.AllDay).
		SetTimezone(projection.temporal.Timezone).
		SetExdates(append([]string(nil), projection.exdates...)).
		SetSequence(sequence).
		SetVersion(1).
		SetSource(calendarevent.Source(CalendarEventSourceMailImport)).
		SetStatus(calendarevent.Status(projection.status)).
		SetOrganizerEmail(incoming.OrganizerEmail()).
		SetPayloadHashVersion(payload.CanonicalVersion()).
		SetPayloadHash(payloadHash[:])
	if projection.temporal.AllDay {
		startDate, startErr := time.Parse(time.DateOnly, projection.temporal.StartDate)
		endDate, endErr := time.Parse(time.DateOnly, projection.temporal.EndDateExclusive)
		if startErr != nil || endErr != nil {
			return nil, ErrEventInvalid
		}
		create.SetStartDate(startDate).SetEndDateExclusive(endDate)
	} else {
		create.SetStartsAt(projection.temporal.StartsAt).SetEndsAt(projection.temporal.EndsAt)
	}
	if message.RRULE != "" {
		create.SetRrule(message.RRULE)
	}
	saved, err := create.Save(ctx)
	if err != nil {
		return nil, calendarEntSafeError(err)
	}
	return saved, nil
}

func (backend *eventEntTransactionBackend) updateCalendarInboundInviteEvent(
	ctx context.Context,
	event *ent.CalendarEvent,
	incoming CalendarInboundITIP,
	message ITIPMessage,
) error {
	if event == nil || !calendarInboundInviteImportedEventMatches(
		event,
		event.OrganizationID,
		event.CalendarOwnerID,
		incoming.Component().UID(),
		incoming.OrganizerEmail(),
	) {
		return ErrEventInvalid
	}
	projection, err := calendarInboundInviteProjection(message)
	if err != nil {
		return err
	}
	version, ok := eventEntUint(event.Version)
	if !ok || version == ^uint(0) {
		return ErrEventInvalid
	}
	nextVersion, ok := eventEntInt(version + 1)
	if !ok {
		return ErrEventInvalid
	}
	sequence, ok := eventEntInt(uint(incoming.Sequence()))
	if !ok {
		return ErrEventInvalid
	}
	payload := incoming.Payload()
	payloadHash := payload.Hash()
	update := backend.tx.CalendarEvent.UpdateOne(event).
		SetTitle(message.Summary).
		SetDescription(message.Description).
		SetLocation(message.Location).
		SetAllDay(projection.temporal.AllDay).
		SetTimezone(projection.temporal.Timezone).
		SetExdates(append([]string(nil), projection.exdates...)).
		SetSequence(sequence).
		SetVersion(nextVersion).
		SetStatus(calendarevent.Status(projection.status)).
		SetPayloadHashVersion(payload.CanonicalVersion()).
		SetPayloadHash(payloadHash[:]).
		ClearDeletedAt()
	if projection.temporal.AllDay {
		startDate, startErr := time.Parse(time.DateOnly, projection.temporal.StartDate)
		endDate, endErr := time.Parse(time.DateOnly, projection.temporal.EndDateExclusive)
		if startErr != nil || endErr != nil {
			return ErrEventInvalid
		}
		update.ClearStartsAt().
			ClearEndsAt().
			SetStartDate(startDate).
			SetEndDateExclusive(endDate)
	} else {
		update.ClearStartDate().
			ClearEndDateExclusive().
			SetStartsAt(projection.temporal.StartsAt).
			SetEndsAt(projection.temporal.EndsAt)
	}
	if message.RRULE == "" {
		update.ClearRrule()
	} else {
		update.SetRrule(message.RRULE)
	}
	if err := update.Exec(ctx); err != nil {
		return calendarEntSafeError(err)
	}
	return nil
}

func (backend *eventEntTransactionBackend) persistCalendarInboundInviteSeriesRequest(
	ctx context.Context,
	durable calendarInboundInviteDurableScope,
	command CalendarInboundInvitePersistenceCommand,
) (*ent.CalendarEvent, []uuid.UUID, error) {
	incoming := command.input.incoming
	if !incoming.Component().IsSeries() || incoming.Method() != ITIPMethodRequest {
		return nil, nil, ErrEventInvalid
	}
	var boundID *uuid.UUID
	if durable.current != nil {
		boundID = durable.current.EventID()
	}
	event, err := backend.lockCalendarInboundInviteEvent(
		ctx,
		durable.organizationID,
		durable.ownerID,
		incoming.Component().UID(),
		incoming.OrganizerEmail(),
		boundID,
	)
	if err != nil {
		return nil, nil, err
	}
	if event == nil {
		event, err = backend.createCalendarInboundInviteEvent(
			ctx,
			durable.organizationID,
			durable.ownerID,
			incoming,
			command.request.message,
		)
		if err != nil {
			return nil, nil, err
		}
	} else if err := backend.updateCalendarInboundInviteEvent(ctx, event, incoming, command.request.message); err != nil {
		return nil, nil, err
	}

	attendees, err := backend.lockCalendarInboundInviteAttendees(ctx, event.ID)
	if err != nil {
		return nil, nil, err
	}
	attendeeIDs, err := backend.syncCalendarInboundInviteAttendees(
		ctx,
		event.ID,
		attendees,
		command.request.message.Attendees,
		durable.ownerID,
		durable.ownerParticipantEmail,
		durable.clock,
	)
	if err != nil {
		return nil, nil, err
	}
	return event, attendeeIDs, nil
}

func (backend *eventEntTransactionBackend) persistCalendarInboundInviteSeriesCancel(
	ctx context.Context,
	durable calendarInboundInviteDurableScope,
	command CalendarInboundInvitePersistenceCommand,
) (*ent.CalendarEvent, []uuid.UUID, bool, error) {
	incoming := command.input.incoming
	if !incoming.Component().IsSeries() || incoming.Method() != ITIPMethodCancel {
		return nil, nil, false, ErrEventInvalid
	}
	if durable.current == nil || durable.current.EventID() == nil {
		return nil, nil, false, nil
	}
	event, err := backend.lockCalendarInboundInviteEvent(
		ctx,
		durable.organizationID,
		durable.ownerID,
		incoming.Component().UID(),
		incoming.OrganizerEmail(),
		durable.current.EventID(),
	)
	if err != nil {
		return nil, nil, false, err
	}
	attendees, err := backend.lockCalendarInboundInviteAttendees(ctx, event.ID)
	if err != nil {
		return nil, nil, false, err
	}
	attendeeIDs, err := calendarInboundInviteActiveAttendeeIDs(attendees)
	if err != nil {
		return nil, nil, false, err
	}
	exceptions, err := backend.tx.CalendarEventException.Query().
		Where(calendareventexception.CalendarEventIDEQ(event.ID)).
		Order(
			ent.Asc(calendareventexception.FieldRecurrenceID),
			ent.Asc(calendareventexception.FieldID),
		).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, nil, false, calendarEntSafeError(err)
	}
	version, ok := eventEntUint(event.Version)
	if !ok || version == ^uint(0) {
		return nil, nil, false, ErrEventInvalid
	}
	nextVersion, ok := eventEntInt(version + 1)
	if !ok {
		return nil, nil, false, ErrEventInvalid
	}
	sequence, ok := eventEntInt(uint(incoming.Sequence()))
	if !ok {
		return nil, nil, false, ErrEventInvalid
	}
	payload := incoming.Payload()
	payloadHash := payload.Hash()
	if err := backend.tx.CalendarEvent.UpdateOne(event).
		SetStatus(calendarevent.Status(CalendarEventStatusCancelled)).
		SetSequence(sequence).
		SetVersion(nextVersion).
		SetPayloadHashVersion(payload.CanonicalVersion()).
		SetPayloadHash(payloadHash[:]).
		Exec(ctx); err != nil {
		return nil, nil, false, calendarEntSafeError(err)
	}
	for _, exception := range exceptions {
		if exception == nil || exception.ID == uuid.Nil {
			return nil, nil, false, ErrEventInvalid
		}
		if exception.Status != nil && *exception.Status == calendareventexception.Status(CalendarEventStatusCancelled) {
			continue
		}
		if err := backend.tx.CalendarEventException.UpdateOne(exception).
			SetStatus(calendareventexception.Status(CalendarEventStatusCancelled)).
			Exec(ctx); err != nil {
			return nil, nil, false, calendarEntSafeError(err)
		}
	}
	return event, attendeeIDs, true, nil
}

func (backend *eventEntTransactionBackend) persistCalendarInboundInviteOccurrenceCancel(
	ctx context.Context,
	durable calendarInboundInviteDurableScope,
	command CalendarInboundInvitePersistenceCommand,
) (*ent.CalendarEvent, []uuid.UUID, bool, error) {
	incoming := command.input.incoming
	component := incoming.Component()
	if component.IsSeries() || incoming.Method() != ITIPMethodCancel ||
		durable.current == nil || durable.current.EventID() == nil {
		return nil, nil, false, nil
	}
	recurrenceID := component.RecurrenceID()
	if recurrenceID == nil {
		return nil, nil, false, ErrEventInvalid
	}
	event, err := backend.lockCalendarInboundInviteEvent(
		ctx,
		durable.organizationID,
		durable.ownerID,
		component.UID(),
		incoming.OrganizerEmail(),
		durable.current.EventID(),
	)
	if err != nil {
		return nil, nil, false, err
	}
	attendees, err := backend.lockCalendarInboundInviteAttendees(ctx, event.ID)
	if err != nil {
		return nil, nil, false, err
	}
	attendeeIDs, err := calendarInboundInviteActiveAttendeeIDs(attendees)
	if err != nil {
		return nil, nil, false, err
	}
	exception, err := backend.lockCalendarInboundInviteException(ctx, event.ID, *recurrenceID)
	if err != nil {
		return nil, nil, false, err
	}
	if exception == nil {
		return nil, nil, false, nil
	}
	sequence, ok := eventEntInt(uint(incoming.Sequence()))
	if !ok {
		return nil, nil, false, ErrEventInvalid
	}
	payload := incoming.Payload()
	payloadHash := payload.Hash()
	if err := backend.tx.CalendarEventException.UpdateOne(exception).
		SetSequence(sequence).
		SetStatus(calendareventexception.Status(CalendarEventStatusCancelled)).
		SetPayloadHashVersion(payload.CanonicalVersion()).
		SetPayloadHash(payloadHash[:]).
		Exec(ctx); err != nil {
		return nil, nil, false, calendarEntSafeError(err)
	}
	return event, attendeeIDs, true, nil
}

func (backend *eventEntTransactionBackend) lockCalendarInboundInviteAttendees(
	ctx context.Context,
	eventID uuid.UUID,
) ([]*ent.CalendarAttendee, error) {
	if backend == nil || backend.tx == nil || eventID == uuid.Nil {
		return nil, ErrEventInvalid
	}
	rows, err := backend.tx.CalendarAttendee.Query().
		Where(calendarattendee.CalendarEventIDEQ(eventID)).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, calendarEntSafeError(err)
	}
	return rows, nil
}

func (backend *eventEntTransactionBackend) lockCalendarInboundInviteException(
	ctx context.Context,
	eventID uuid.UUID,
	recurrenceID string,
) (*ent.CalendarEventException, error) {
	if backend == nil || backend.tx == nil || eventID == uuid.Nil || !validCalendarInboundRecurrenceID(recurrenceID) {
		return nil, ErrEventInvalid
	}
	rows, err := backend.tx.CalendarEventException.Query().
		Where(
			calendareventexception.CalendarEventIDEQ(eventID),
			calendareventexception.RecurrenceIDEQ(recurrenceID),
		).
		Order(ent.Asc(calendareventexception.FieldID)).
		ForUpdate().
		Limit(2).
		All(ctx)
	if err != nil {
		return nil, calendarEntSafeError(err)
	}
	if len(rows) > 1 {
		return nil, ErrEventInvalid
	}
	if len(rows) == 0 {
		return nil, nil
	}
	if rows[0] == nil || rows[0].ID == uuid.Nil {
		return nil, ErrEventInvalid
	}
	return rows[0], nil
}

type calendarInboundInviteParticipant struct {
	email       string
	displayName string
	role        CalendarEventAttendeeRole
	status      RSVPStatus
}

func calendarInboundInviteParticipants(values []ITIPParticipant) ([]calendarInboundInviteParticipant, error) {
	if len(values) > calendarCreateMaxAttendees {
		return nil, ErrEventInvalid
	}
	participants := make([]calendarInboundInviteParticipant, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		email := normalizeCalendarCreateEmail(value.Email)
		if email == "" || email != value.Email || value.Status.Validate() != nil {
			return nil, ErrEventInvalid
		}
		var role CalendarEventAttendeeRole
		switch value.Role {
		case ITIPParticipantRequired:
			role = CalendarEventAttendeeRoleRequired
		case ITIPParticipantOptional:
			role = CalendarEventAttendeeRoleOptional
		default:
			return nil, ErrEventInvalid
		}
		if _, exists := seen[email]; exists {
			return nil, ErrEventInvalid
		}
		seen[email] = struct{}{}
		participants = append(participants, calendarInboundInviteParticipant{
			email:       email,
			displayName: value.DisplayName,
			role:        role,
			status:      value.Status,
		})
	}
	sort.Slice(participants, func(left, right int) bool {
		return participants[left].email < participants[right].email
	})
	return participants, nil
}

func (backend *eventEntTransactionBackend) syncCalendarInboundInviteAttendees(
	ctx context.Context,
	eventID uuid.UUID,
	rows []*ent.CalendarAttendee,
	values []ITIPParticipant,
	ownerID uuid.UUID,
	ownerParticipantEmail string,
	clock time.Time,
) ([]uuid.UUID, error) {
	if backend == nil || backend.tx == nil || eventID == uuid.Nil || ownerID == uuid.Nil ||
		ownerParticipantEmail == "" || normalizeCalendarCreateEmail(ownerParticipantEmail) != ownerParticipantEmail ||
		!calendarInboundInviteWholeSecondUTC(clock) {
		return nil, ErrEventInvalid
	}
	participants, err := calendarInboundInviteParticipants(values)
	if err != nil {
		return nil, err
	}
	ownerParticipantCount := 0
	for _, participant := range participants {
		if participant.email == ownerParticipantEmail {
			ownerParticipantCount++
		}
	}
	if ownerParticipantCount != 1 {
		return nil, ErrEventInvalid
	}
	existing := make(map[string]*ent.CalendarAttendee, len(rows))
	activeOwnerBindings := 0
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil || row.CalendarEventID != eventID {
			return nil, ErrEventInvalid
		}
		if row.RecurrenceID != nil {
			continue
		}
		if row.Role == calendarattendee.Role(CalendarEventAttendeeRoleOrganizer) || row.NormalizedEmail == "" ||
			row.Version < 1 || row.Version == int(^uint(0)>>1) {
			return nil, ErrEventInvalid
		}
		if row.UserID != nil {
			if *row.UserID == uuid.Nil {
				return nil, ErrEventInvalid
			}
			if row.NormalizedEmail == ownerParticipantEmail && *row.UserID != ownerID {
				return nil, ErrEventInvalid
			}
			if row.RemovedAt == nil && *row.UserID == ownerID {
				activeOwnerBindings++
				if row.NormalizedEmail != ownerParticipantEmail {
					return nil, ErrEventInvalid
				}
			}
		}
		if _, duplicate := existing[row.NormalizedEmail]; duplicate {
			return nil, ErrEventInvalid
		}
		existing[row.NormalizedEmail] = row
	}
	if activeOwnerBindings > 1 {
		return nil, ErrEventInvalid
	}

	activeIDs := make([]uuid.UUID, 0, len(participants))
	for _, participant := range participants {
		row := existing[participant.email]
		if row == nil {
			create := backend.tx.CalendarAttendee.Create().
				SetCalendarEventID(eventID).
				SetNormalizedEmail(participant.email).
				SetRole(calendarattendee.Role(participant.role)).
				SetRsvpStatus(calendarattendee.RsvpStatus(participant.status)).
				SetVersion(1)
			if participant.email == ownerParticipantEmail {
				create.SetUserID(ownerID)
			}
			if participant.displayName != "" {
				create.SetDisplayName(participant.displayName)
			}
			if participant.status != RSVPNeedsAction {
				create.SetRespondedAt(clock)
			}
			saved, saveErr := create.Save(ctx)
			if saveErr != nil || saved == nil || saved.ID == uuid.Nil {
				if saveErr != nil {
					return nil, calendarEntSafeError(saveErr)
				}
				return nil, ErrEventInvalid
			}
			activeIDs = append(activeIDs, saved.ID)
			continue
		}
		update := backend.tx.CalendarAttendee.UpdateOne(row).
			SetRole(calendarattendee.Role(participant.role)).
			SetRsvpStatus(calendarattendee.RsvpStatus(participant.status)).
			ClearRemovedAt().
			SetVersion(row.Version + 1)
		if participant.email == ownerParticipantEmail {
			update.SetUserID(ownerID)
		} else {
			update.ClearUserID()
		}
		if participant.displayName == "" {
			update.ClearDisplayName()
		} else {
			update.SetDisplayName(participant.displayName)
		}
		if participant.status == RSVPNeedsAction {
			update.ClearRespondedAt()
		} else {
			update.SetRespondedAt(clock)
		}
		if err := update.Exec(ctx); err != nil {
			return nil, calendarEntSafeError(err)
		}
		activeIDs = append(activeIDs, row.ID)
	}

	desired := make(map[string]struct{}, len(participants))
	for _, participant := range participants {
		desired[participant.email] = struct{}{}
	}
	for email, row := range existing {
		if row.RemovedAt != nil {
			continue
		}
		if _, retained := desired[email]; retained {
			continue
		}
		if err := backend.tx.CalendarAttendee.UpdateOne(row).
			SetRemovedAt(clock).
			SetVersion(row.Version + 1).
			Exec(ctx); err != nil {
			return nil, calendarEntSafeError(err)
		}
	}
	return calendarInboundInviteSortUUIDs(activeIDs)
}

func calendarInboundInviteActiveAttendeeIDs(rows []*ent.CalendarAttendee) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil {
			return nil, ErrEventInvalid
		}
		if row.RecurrenceID != nil || row.RemovedAt != nil {
			continue
		}
		if row.Role == calendarattendee.Role(CalendarEventAttendeeRoleOrganizer) {
			return nil, ErrEventInvalid
		}
		if _, duplicate := seen[row.ID]; duplicate {
			return nil, ErrEventInvalid
		}
		seen[row.ID] = struct{}{}
		ids = append(ids, row.ID)
	}
	return calendarInboundInviteSortUUIDs(ids)
}

func calendarInboundInviteSortUUIDs(ids []uuid.UUID) ([]uuid.UUID, error) {
	sorted := append([]uuid.UUID(nil), ids...)
	sort.Slice(sorted, func(left, right int) bool {
		return bytes.Compare(sorted[left][:], sorted[right][:]) < 0
	})
	for index, id := range sorted {
		if id == uuid.Nil || index > 0 && bytes.Equal(sorted[index-1][:], id[:]) {
			return nil, ErrEventInvalid
		}
	}
	return sorted, nil
}

func calendarInboundInviteChangedFields() []CalendarEventChangedField {
	return []CalendarEventChangedField{
		CalendarEventChangedFieldTitle,
		CalendarEventChangedFieldDescription,
		CalendarEventChangedFieldLocation,
		CalendarEventChangedFieldStartsAt,
		CalendarEventChangedFieldEndsAt,
		CalendarEventChangedFieldStartDate,
		CalendarEventChangedFieldEndDateExclusive,
		CalendarEventChangedFieldAllDay,
		CalendarEventChangedFieldTimezone,
		CalendarEventChangedFieldRRULE,
		CalendarEventChangedFieldExdates,
		CalendarEventChangedFieldAttendees,
	}
}

func (backend *eventEntTransactionBackend) appendCalendarInboundInviteOutbox(
	ctx context.Context,
	command CalendarInboundInvitePersistenceCommand,
	clock time.Time,
	organizationID, calendarEventID uuid.UUID,
	attendeeIDs []uuid.UUID,
) error {
	if backend == nil || backend.tx == nil || !command.valid() || command.OutboxEnvelopeID() == uuid.Nil ||
		organizationID == uuid.Nil || calendarEventID == uuid.Nil || !calendarInboundInviteWholeSecondUTC(clock) {
		return ErrEventInvalid
	}
	sortedAttendeeIDs, err := calendarInboundInviteSortUUIDs(attendeeIDs)
	if err != nil {
		return err
	}
	incoming := command.input.incoming
	var envelope events.EnvelopeV1
	switch command.Outcome() {
	case CalendarInboundIngestApplyRequest:
		envelope, err = NewCalendarInviteUpdatedEnvelopeV1(
			command.OutboxEnvelopeID(),
			clock,
			organizationID,
			nil,
			calendarEventID.String(),
			CalendarInviteUpdatedPayload{
				EventID:       calendarEventID,
				UID:           incoming.Component().UID(),
				Sequence:      incoming.Sequence(),
				AttendeeIDs:   sortedAttendeeIDs,
				ResetRSVP:     false,
				ChangedFields: calendarInboundInviteChangedFields(),
			},
		)
	case CalendarInboundIngestTombstoneSeries, CalendarInboundIngestTombstoneOccurrence:
		envelope, err = NewCalendarInviteCancelledEnvelopeV1(
			command.OutboxEnvelopeID(),
			clock,
			organizationID,
			nil,
			calendarEventID.String(),
			CalendarInviteCancelledPayload{
				EventID:     calendarEventID,
				UID:         incoming.Component().UID(),
				Sequence:    incoming.Sequence(),
				AttendeeIDs: sortedAttendeeIDs,
			},
		)
	default:
		return ErrEventInvalid
	}
	if err != nil {
		return ErrEventInvalid
	}
	descriptor, err := NewEventOutboxDescriptor(envelope)
	if err != nil || !descriptor.valid() || descriptor.Envelope().Subject() != envelope.Subject() {
		return ErrEventInvalid
	}
	if err := outbox.NewPublisher().PublishInTx(ctx, backend.tx, envelope.Subject(), descriptor.Envelope()); err != nil {
		return calendarEntSafeError(err)
	}
	return nil
}

func (backend *eventEntTransactionBackend) persistCalendarInboundInviteOccurrenceRequest(
	ctx context.Context,
	durable calendarInboundInviteDurableScope,
	command CalendarInboundInvitePersistenceCommand,
) (*ent.CalendarEvent, []uuid.UUID, error) {
	incoming := command.input.incoming
	component := incoming.Component()
	if component.IsSeries() || incoming.Method() != ITIPMethodRequest ||
		durable.series == nil || durable.series.Status() != CalendarInboundObjectApplied {
		return nil, nil, ErrEventInvalid
	}
	masterID := durable.series.EventID()
	recurrenceID := component.RecurrenceID()
	if masterID == nil || recurrenceID == nil {
		return nil, nil, ErrEventInvalid
	}
	event, err := backend.lockCalendarInboundInviteEvent(
		ctx,
		durable.organizationID,
		durable.ownerID,
		component.UID(),
		incoming.OrganizerEmail(),
		masterID,
	)
	if err != nil {
		return nil, nil, err
	}
	if (event.Status != calendarevent.Status(CalendarEventStatusConfirmed) &&
		event.Status != calendarevent.Status(CalendarEventStatusTentative)) ||
		event.DeletedAt != nil {
		return nil, nil, ErrEventInvalid
	}
	attendees, err := backend.lockCalendarInboundInviteAttendees(ctx, event.ID)
	if err != nil {
		return nil, nil, err
	}
	attendeeIDs, err := calendarInboundInviteActiveAttendeeIDs(attendees)
	if err != nil {
		return nil, nil, err
	}
	exception, err := backend.lockCalendarInboundInviteException(ctx, event.ID, *recurrenceID)
	if err != nil {
		return nil, nil, err
	}
	projection, err := calendarInboundInviteProjection(command.request.message)
	if err != nil {
		return nil, nil, err
	}
	sequence, ok := eventEntInt(uint(incoming.Sequence()))
	if !ok {
		return nil, nil, ErrEventInvalid
	}
	payload := incoming.Payload()
	payloadHash := payload.Hash()
	if exception == nil {
		create := backend.tx.CalendarEventException.Create().
			SetCalendarEventID(event.ID).
			SetRecurrenceID(*recurrenceID).
			SetSequence(sequence).
			SetStatus(calendareventexception.Status(projection.status)).
			SetPayloadHashVersion(payload.CanonicalVersion()).
			SetPayloadHash(payloadHash[:])
		if err := eventEntSetExceptionCreateTemporal(create, &projection.temporal); err != nil {
			return nil, nil, err
		}
		if command.request.message.Location != "" {
			create.SetLocation(command.request.message.Location)
		}
		if _, err := create.Save(ctx); err != nil {
			return nil, nil, calendarEntSafeError(err)
		}
	} else {
		update := backend.tx.CalendarEventException.UpdateOne(exception).
			SetSequence(sequence).
			SetStatus(calendareventexception.Status(projection.status)).
			SetPayloadHashVersion(payload.CanonicalVersion()).
			SetPayloadHash(payloadHash[:])
		if err := eventEntSetExceptionUpdateTemporal(update, &projection.temporal); err != nil {
			return nil, nil, err
		}
		if command.request.message.Location == "" {
			update.ClearLocation()
		} else {
			update.SetLocation(command.request.message.Location)
		}
		if err := update.Exec(ctx); err != nil {
			return nil, nil, calendarEntSafeError(err)
		}
	}
	return event, attendeeIDs, nil
}
