package calendar

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendaravailabilityreservation"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendareventexception"
	"github.com/valerii/hub/backend/ent/leaverequest"
	"github.com/valerii/hub/backend/ent/scheduledcall"
	"github.com/valerii/hub/backend/ent/user"
)

const (
	calendarAvailabilityReservationConstraintName = "calendar_availability_reservations_active_window_excl"
	calendarAvailabilityReservationFinalEventLimit = 1000
	calendarAvailabilityReservationFinalComparisonLimit = 5_000_000
)

var _ AvailabilityReservationTransactionBackend = (*eventEntTransactionBackend)(nil)

type calendarAvailabilityReservationIdentity struct {
	userID        uuid.UUID
	occurrenceKey string
}

// ReconcileCalendarAvailabilityReservations locks every participant in a
// deterministic order, verifies all authoritative busy sources in the same
// transaction, and then reconciles the exact active reservation set. The
// exclusion constraint remains the authoritative race boundary.
func (backend *eventEntTransactionBackend) ReconcileCalendarAvailabilityReservations(ctx context.Context, scope EventMutationScope, set CalendarAvailabilityReservationSet) error {
	if backend == nil || backend.tx == nil || backend.now == nil {
		return ErrAvailabilitySourceInvalid
	}
	if scope.ActorID() != set.OrganizerID() || scope.OrganizationID() != set.OrganizationID() {
		return ErrEventForbidden
	}
	if !set.valid() {
		return ErrAvailabilitySourceInvalid
	}
	participants := set.Participants()
	occurrences := set.Occurrences()
	releaseAll := len(occurrences) == 0
	if err := backend.acquireCalendarAvailabilityAdvisoryLocks(ctx, participants); err != nil {
		return err
	}
	if err := backend.lockCalendarAvailabilityReservationEvent(ctx, scope, set, releaseAll); err != nil {
		return err
	}
	if err := backend.requireCalendarAvailabilityReservationConstraint(ctx); err != nil {
		return err
	}

	if releaseAll {
		active, err := backend.lockActiveCalendarAvailabilityReservations(ctx, set)
		if err != nil {
			return err
		}
		desired := calendarAvailabilityReservationDesired(set, participants, occurrences)
		if err := calendarAvailabilityReservationActiveRowsMatch(active, desired); err != nil {
			return err
		}
		now, ok := canonicalAvailabilityTime(backend.now().UTC().Truncate(time.Second))
		if !ok {
			return ErrAvailabilitySourceInvalid
		}
		return backend.releaseObsoleteCalendarAvailabilityReservations(ctx, active, desired, now)
	}

	members, err := backend.lockCalendarAvailabilityReservationMembers(ctx, set.OrganizationID(), participants)
	if err != nil {
		return err
	}
	active, err := backend.lockActiveCalendarAvailabilityReservations(ctx, set)
	if err != nil {
		return err
	}
	desired := calendarAvailabilityReservationDesired(set, participants, occurrences)
	if err := calendarAvailabilityReservationActiveRowsMatch(active, desired); err != nil {
		return err
	}
	if err := backend.checkCalendarAvailabilityReservationLeaves(ctx, members, occurrences); err != nil {
		return err
	}
	if err := backend.checkCalendarAvailabilityReservationEvents(ctx, set, participants, occurrences); err != nil {
		return err
	}
	if err := backend.checkCalendarAvailabilityReservationRows(ctx, set, participants, occurrences); err != nil {
		return err
	}
	if err := backend.checkCalendarAvailabilityReservationCalls(ctx, participants, occurrences); err != nil {
		return err
	}

	now, ok := canonicalAvailabilityTime(backend.now().UTC().Truncate(time.Second))
	if !ok {
		return ErrAvailabilitySourceInvalid
	}
	if err := backend.releaseObsoleteCalendarAvailabilityReservations(ctx, active, desired, now); err != nil {
		return err
	}
	return backend.insertMissingCalendarAvailabilityReservations(ctx, set, desired)
}

func (backend *eventEntTransactionBackend) acquireCalendarAvailabilityAdvisoryLocks(ctx context.Context, participants []uuid.UUID) error {
	if backend == nil || backend.tx == nil || len(participants) == 0 {
		return ErrAvailabilitySourceInvalid
	}
	driver := backend.tx.Driver()
	if driver == nil {
		return ErrAvailabilitySourceInvalid
	}
	previous := ""
	for _, participantID := range participants {
		if participantID == uuid.Nil || participantID.String() <= previous {
			return ErrAvailabilitySourceInvalid
		}
		previous = participantID.String()
		var result sql.Result
		if err := driver.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", []any{calendarAvailabilityReservationAdvisoryKey(participantID)}, &result); err != nil {
			return calendarEntSafeError(err)
		}
	}
	return nil
}

func calendarAvailabilityReservationAdvisoryKey(participantID uuid.UUID) int64 {
	digest := sha256.Sum256(participantID[:])
	return int64(binary.BigEndian.Uint64(digest[:8]))
}

func (backend *eventEntTransactionBackend) requireCalendarAvailabilityReservationConstraint(ctx context.Context) error {
	if backend == nil || backend.tx == nil || backend.tx.Driver() == nil {
		return ErrAvailabilitySourceInvalid
	}
	var rows sql.Rows
	if err := backend.tx.Driver().Query(ctx, "SELECT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = $1 AND contype = 'x' AND conrelid = 'calendar_availability_reservations'::regclass)", []any{calendarAvailabilityReservationConstraintName}, &rows); err != nil {
		return calendarEntSafeError(err)
	}
	var found bool
	scanErr := sql.ScanOne(&rows, &found)
	closeErr := rows.Close()
	if scanErr != nil || closeErr != nil || !found {
		return ErrAvailabilitySourceInvalid
	}
	return nil
}

func (backend *eventEntTransactionBackend) lockCalendarAvailabilityReservationMembers(ctx context.Context, organizationID uuid.UUID, participants []uuid.UUID) (map[uuid.UUID]*ent.User, error) {
	if organizationID == uuid.Nil || len(participants) == 0 {
		return nil, ErrAvailabilitySourceInvalid
	}
	members := make(map[uuid.UUID]*ent.User, len(participants))
	for _, participantID := range participants {
		member, err := backend.activeOrganizationMemberForUpdate(ctx, participantID, organizationID)
		if err != nil {
			return nil, err
		}
		if member == nil || member.ID != participantID || !availabilitySourceTimezoneValid(member.Timezone) {
			return nil, ErrAvailabilitySourceInvalid
		}
		members[participantID] = member
	}
	return members, nil
}

func (backend *eventEntTransactionBackend) lockCalendarAvailabilityReservationEvent(ctx context.Context, scope EventMutationScope, set CalendarAvailabilityReservationSet, releaseAll bool) error {
	query := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(set.EventID()),
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.CalendarOwnerIDEQ(scope.ActorID()),
			calendarevent.OrganizerIDEQ(scope.ActorID()),
			calendarevent.DeletedAtIsNil(),
		)
	if !releaseAll {
		query = query.Where(
			calendarevent.AllDay(false),
			calendarevent.StatusIn(calendarevent.StatusConfirmed, calendarevent.StatusTentative),
		)
	}
	event, err := query.ForUpdate().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrEventForbidden
		}
		return calendarEntSafeError(err)
	}
	if event == nil || event.ID != set.EventID() || event.OrganizationID != scope.OrganizationID() || event.CalendarOwnerID != scope.ActorID() || event.OrganizerID == nil || *event.OrganizerID != scope.ActorID() {
		return ErrEventForbidden
	}
	if releaseAll {
		return nil
	}
	if event.AllDay || event.DeletedAt != nil {
		return ErrAvailabilitySourceInvalid
	}
	attendees, err := backend.tx.CalendarAttendee.Query().
		Where(
			calendarattendee.CalendarEventIDEQ(event.ID),
			calendarattendee.UserIDNotNil(),
			calendarattendee.RemovedAtIsNil(),
			calendarattendee.RecurrenceIDIsNil(),
		).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		Limit(calendarAvailabilityReservationMaxAttendees + 2).
		All(ctx)
	if err != nil {
		return calendarEntSafeError(err)
	}
	if !calendarAvailabilityReservationAttendeesMatch(set, attendees) {
		return ErrAvailabilitySourceInvalid
	}
	exceptions, err := backend.tx.CalendarEventException.Query().
		Where(calendareventexception.CalendarEventIDEQ(event.ID)).
		Order(ent.Asc(calendareventexception.FieldRecurrenceID), ent.Asc(calendareventexception.FieldID)).
		ForUpdate().
		Limit(calendarAvailabilityReservationMaxOccurrences + 1).
		All(ctx)
	if err != nil {
		return calendarEntSafeError(err)
	}
	if len(exceptions) > calendarAvailabilityReservationMaxOccurrences {
		return ErrAvailabilityLimitExceeded
	}
	occurrences, _, _, occurrenceErr := availabilityEventOccurrences(event, exceptions)
	if occurrenceErr != nil || !calendarAvailabilityReservationOccurrencesMatch(occurrences, set.Occurrences()) {
		return ErrAvailabilitySourceInvalid
	}
	return nil
}

func calendarAvailabilityReservationAttendeesMatch(set CalendarAvailabilityReservationSet, rows []*ent.CalendarAttendee) bool {
	if len(rows) > calendarAvailabilityReservationMaxAttendees+1 {
		return false
	}
	expected := make(map[uuid.UUID]struct{}, len(set.participants)-1)
	for _, participantID := range set.Participants() {
		if participantID != set.OrganizerID() {
			expected[participantID] = struct{}{}
		}
	}
	actual := make(map[uuid.UUID]struct{}, len(rows))
	organizerRows := 0
	for _, attendee := range rows {
		if attendee == nil || attendee.ID == uuid.Nil || attendee.RemovedAt != nil || attendee.RecurrenceID != nil {
			return false
		}
		if attendee.UserID == nil {
			continue
		}
		if *attendee.UserID == uuid.Nil {
			return false
		}
		role := CalendarEventAttendeeRole(attendee.Role)
		status := RSVPStatus(attendee.RsvpStatus)
		if !validCalendarEventStoredAttendeeRoleState(role, status, "") {
			return false
		}
		organizerIdentity := *attendee.UserID == set.OrganizerID()
		if role == CalendarEventAttendeeRoleOrganizer || organizerIdentity {
			if !validCalendarEventStoredOrganizerAttendeeState(role, status, "", attendee.UserID, set.OrganizerID(), set.OrganizerID()) || organizerRows != 0 {
				return false
			}
			organizerRows++
			continue
		}
		if _, duplicate := actual[*attendee.UserID]; duplicate {
			return false
		}
		actual[*attendee.UserID] = struct{}{}
	}
	if organizerRows != 1 || len(actual) != len(expected) {
		return false
	}
	for participantID := range expected {
		if _, found := actual[participantID]; !found {
			return false
		}
	}
	return true
}

func calendarAvailabilityReservationOccurrencesMatch(actual []availabilityOccurrence, expected []CalendarAvailabilityReservationOccurrence) bool {
	if len(actual) != len(expected) {
		return false
	}
	byKey := make(map[string]availabilityOccurrence, len(actual))
	for _, occurrence := range actual {
		if occurrence.key == "" || occurrence.from.IsZero() || occurrence.from.Location() != time.UTC || occurrence.to.IsZero() || !occurrence.to.After(occurrence.from) {
			return false
		}
		effectiveKey := occurrence.from.UTC().Format(time.RFC3339Nano)
		if _, duplicate := byKey[effectiveKey]; duplicate {
			return false
		}
		byKey[effectiveKey] = occurrence
	}
	for _, occurrence := range expected {
		actualOccurrence, found := byKey[occurrence.OccurrenceKey]
		if !found || !actualOccurrence.from.Equal(occurrence.StartsAt) || !actualOccurrence.to.Equal(occurrence.EndsAt) {
			return false
		}
	}
	return true
}

func (backend *eventEntTransactionBackend) lockActiveCalendarAvailabilityReservations(ctx context.Context, set CalendarAvailabilityReservationSet) ([]*ent.CalendarAvailabilityReservation, error) {
	rows, err := backend.tx.CalendarAvailabilityReservation.Query().
		Where(
			calendaravailabilityreservation.OrganizationIDEQ(set.OrganizationID()),
			calendaravailabilityreservation.CalendarEventIDEQ(set.EventID()),
			calendaravailabilityreservation.StateEQ(calendaravailabilityreservation.StateActive),
		).
		Order(ent.Asc(calendaravailabilityreservation.FieldUserID), ent.Asc(calendaravailabilityreservation.FieldOccurrenceKey), ent.Asc(calendaravailabilityreservation.FieldID)).
		ForUpdate().
		Limit(calendarAvailabilityReservationMaxRows + 1).
		All(ctx)
	if err != nil {
		return nil, calendarEntSafeError(err)
	}
	if len(rows) > calendarAvailabilityReservationMaxRows {
		return nil, ErrAvailabilityLimitExceeded
	}
	return rows, nil
}

func calendarAvailabilityReservationDesired(set CalendarAvailabilityReservationSet, participants []uuid.UUID, occurrences []CalendarAvailabilityReservationOccurrence) map[calendarAvailabilityReservationIdentity]CalendarAvailabilityReservationOccurrence {
	desired := make(map[calendarAvailabilityReservationIdentity]CalendarAvailabilityReservationOccurrence, len(participants)*len(occurrences))
	for _, participantID := range participants {
		for _, occurrence := range occurrences {
			desired[calendarAvailabilityReservationIdentity{userID: participantID, occurrenceKey: occurrence.OccurrenceKey}] = occurrence
		}
	}
	return desired
}

func calendarAvailabilityReservationActiveRowsMatch(rows []*ent.CalendarAvailabilityReservation, desired map[calendarAvailabilityReservationIdentity]CalendarAvailabilityReservationOccurrence) error {
	seen := make(map[calendarAvailabilityReservationIdentity]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil || row.OrganizationID == uuid.Nil || row.UserID == uuid.Nil || row.CalendarEventID == uuid.Nil || row.State != calendaravailabilityreservation.StateActive || row.ReleasedAt != nil || row.Version < 1 {
			return ErrAvailabilitySourceInvalid
		}
		startsAt, endsAt, ok := canonicalCalendarAvailabilityReservationInterval(row.StartsAt, row.EndsAt)
		if !ok || row.OccurrenceKey != startsAt.Format(time.RFC3339Nano) {
			return ErrAvailabilitySourceInvalid
		}
		identity := calendarAvailabilityReservationIdentity{userID: row.UserID, occurrenceKey: row.OccurrenceKey}
		if _, duplicate := seen[identity]; duplicate {
			return ErrAvailabilitySourceInvalid
		}
		seen[identity] = struct{}{}
		if expected, wanted := desired[identity]; wanted && (!startsAt.Equal(expected.StartsAt) || !endsAt.Equal(expected.EndsAt)) {
			return ErrAvailabilitySourceInvalid
		}
	}
	return nil
}

func (backend *eventEntTransactionBackend) checkCalendarAvailabilityReservationLeaves(ctx context.Context, members map[uuid.UUID]*ent.User, occurrences []CalendarAvailabilityReservationOccurrence) error {
	minimum, maximum := calendarAvailabilityReservationBounds(occurrences)
	lowerDate, upperDate := availabilitySourceDateBounds(minimum, maximum)
	participantIDs := calendarAvailabilityReservationMemberIDs(members)
	leaves, err := backend.tx.LeaveRequest.Query().
		Where(
			leaverequest.StatusEQ(leaverequest.StatusApproved),
			leaverequest.DeletedAtIsNil(),
			leaverequest.StartDateLTE(upperDate),
			leaverequest.EndDateGTE(lowerDate),
			leaverequest.HasRequesterWith(user.IDIn(participantIDs...), user.StatusEQ(user.StatusActive)),
		).
		WithRequester().
		Order(ent.Asc(leaverequest.FieldStartDate), ent.Asc(leaverequest.FieldID)).
		ForUpdate().
		Limit(calendarAvailabilityReservationMaxRows + 1).
		All(ctx)
	if err != nil {
		return calendarEntSafeError(err)
	}
	if len(leaves) > calendarAvailabilityReservationMaxRows {
		return ErrAvailabilityLimitExceeded
	}
	for _, leave := range leaves {
		if leave == nil || leave.DeletedAt != nil || leave.Status != leaverequest.StatusApproved || leave.Edges.Requester == nil {
			return ErrAvailabilitySourceInvalid
		}
		member, found := members[leave.Edges.Requester.ID]
		if !found || member == nil || member.ID != leave.Edges.Requester.ID || !availabilitySourceTimezoneValid(member.Timezone) {
			return ErrAvailabilitySourceInvalid
		}
		startDate, startErr := availabilityDateLabel(leave.StartDate)
		endDate, endErr := availabilityDateLabel(leave.EndDate)
		if startErr != nil || endErr != nil || endDate.Before(startDate) {
			return ErrAvailabilitySourceInvalid
		}
		from, to, intervalErr := calendarAvailabilityReservationLeaveInterval(startDate, endDate, member.Timezone)
		if intervalErr != nil {
			return intervalErr
		}
		if calendarAvailabilityReservationOverlapsAny(from, to, occurrences) {
			return ErrAvailabilitySlotUnavailable
		}
	}
	return nil
}

func calendarAvailabilityReservationLeaveInterval(startDate, endDate time.Time, timezone string) (time.Time, time.Time, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil || location == time.Local {
		return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
	}
	return availabilityAllDayDateInterval(startDate, endDate.AddDate(0, 0, 1), timezone)
}

func (backend *eventEntTransactionBackend) checkCalendarAvailabilityReservationEvents(ctx context.Context, set CalendarAvailabilityReservationSet, participants []uuid.UUID, desired []CalendarAvailabilityReservationOccurrence) error {
	minimum, maximum := calendarAvailabilityReservationBounds(desired)
	lowerDate, upperDate := availabilitySourceDateBounds(minimum, maximum)
	events, err := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.OrganizationIDEQ(set.OrganizationID()),
			calendarevent.DeletedAtIsNil(),
			calendarevent.StatusIn(calendarevent.StatusConfirmed, calendarevent.StatusTentative),
			calendarevent.Or(
				calendarevent.And(
					calendarevent.AllDay(false),
					calendarevent.StartsAtLT(maximum),
					calendarevent.Or(calendarevent.RruleNotNil(), calendarevent.EndsAtGT(minimum)),
				),
				calendarevent.And(
					calendarevent.AllDay(true),
					calendarevent.StartDateLT(upperDate),
					calendarevent.Or(calendarevent.RruleNotNil(), calendarevent.EndDateExclusiveGT(lowerDate)),
				),
			),
			calendarevent.Or(
				calendarevent.CalendarOwnerIDIn(participants...),
				calendarevent.OrganizerIDIn(participants...),
				calendarevent.HasAttendeesWith(calendarattendee.UserIDIn(participants...), calendarattendee.RemovedAtIsNil()),
			),
		).
		Order(ent.Asc(calendarevent.FieldID)).
		ForUpdate().
		Limit(calendarAvailabilityReservationFinalEventLimit + 1).
		All(ctx)
	if err != nil {
		return calendarEntSafeError(err)
	}
	if len(events) > calendarAvailabilityReservationFinalEventLimit {
		return ErrAvailabilityLimitExceeded
	}
	if len(events) == 0 {
		return nil
	}

	eventIDs := make([]uuid.UUID, 0, len(events))
	for _, event := range events {
		if event == nil || event.ID == uuid.Nil || event.OrganizationID != set.OrganizationID() {
			return ErrAvailabilitySourceInvalid
		}
		eventIDs = append(eventIDs, event.ID)
	}
	attendees, err := backend.tx.CalendarAttendee.Query().
		Where(
			calendarattendee.CalendarEventIDIn(eventIDs...),
			calendarattendee.UserIDIn(participants...),
			calendarattendee.RemovedAtIsNil(),
		).
		Order(ent.Asc(calendarattendee.FieldCalendarEventID), ent.Asc(calendarattendee.FieldRecurrenceID), ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		Limit(calendarAvailabilityReservationMaxRows + 1).
		All(ctx)
	if err != nil {
		return calendarEntSafeError(err)
	}
	if len(attendees) > calendarAvailabilityReservationMaxRows {
		return ErrAvailabilityLimitExceeded
	}
	exceptions, err := backend.tx.CalendarEventException.Query().
		Where(calendareventexception.CalendarEventIDIn(eventIDs...)).
		Order(ent.Asc(calendareventexception.FieldCalendarEventID), ent.Asc(calendareventexception.FieldRecurrenceID), ent.Asc(calendareventexception.FieldID)).
		ForUpdate().
		Limit(calendarAvailabilityReservationMaxRows + 1).
		All(ctx)
	if err != nil {
		return calendarEntSafeError(err)
	}
	if len(exceptions) > calendarAvailabilityReservationMaxRows {
		return ErrAvailabilityLimitExceeded
	}

	attendeesByEvent := make(map[uuid.UUID][]*ent.CalendarAttendee, len(events))
	for _, attendee := range attendees {
		if attendee == nil || attendee.CalendarEventID == uuid.Nil {
			return ErrAvailabilitySourceInvalid
		}
		attendeesByEvent[attendee.CalendarEventID] = append(attendeesByEvent[attendee.CalendarEventID], attendee)
	}
	exceptionsByEvent := make(map[uuid.UUID][]*ent.CalendarEventException, len(events))
	for _, exception := range exceptions {
		if exception == nil || exception.CalendarEventID == uuid.Nil {
			return ErrAvailabilitySourceInvalid
		}
		exceptionsByEvent[exception.CalendarEventID] = append(exceptionsByEvent[exception.CalendarEventID], exception)
	}
	targets := make(map[uuid.UUID]*AvailabilityUserInput, len(participants))
	for _, participantID := range participants {
		targets[participantID] = &AvailabilityUserInput{UserID: participantID, Timezone: "UTC"}
	}
	comparisons := 0
	for _, event := range events {
		if event.ID == set.EventID() {
			continue
		}
		occurrences, nominal, recurring, occurrenceErr := availabilityEventOccurrences(event, exceptionsByEvent[event.ID])
		if occurrenceErr != nil {
			return occurrenceErr
		}
		states, stateErr := availabilityEventAttendeeStates(attendeesByEvent[event.ID], nominal, recurring, targets)
		if stateErr != nil {
			return stateErr
		}
		overlaps := calendarAvailabilityReservationOccurrenceOverlaps(occurrences, desired)
		for index, occurrence := range occurrences {
			if !overlaps[index] {
				continue
			}
			for _, participantID := range participants {
				comparisons++
				if comparisons > calendarAvailabilityReservationFinalComparisonLimit {
					return ErrAvailabilityLimitExceeded
				}
				if calendarAvailabilityReservationEventBusy(event, participantID, occurrence, states) {
					return ErrAvailabilitySlotUnavailable
				}
			}
		}
	}
	return nil
}

func calendarAvailabilityReservationEventBusy(event *ent.CalendarEvent, participantID uuid.UUID, occurrence availabilityOccurrence, states map[uuid.UUID]availabilityAttendeeState) bool {
	if event == nil || participantID == uuid.Nil {
		return false
	}
	if event.CalendarOwnerID == participantID || event.OrganizerID != nil && *event.OrganizerID == participantID {
		return true
	}
	state, found := states[participantID]
	if !found || !state.hasSeries {
		return false
	}
	status := state.series
	if override, exists := state.occurrences[occurrence.key]; exists {
		status = override
	}
	return availabilityBusyRSVPStatus(status)
}

func (backend *eventEntTransactionBackend) checkCalendarAvailabilityReservationRows(ctx context.Context, set CalendarAvailabilityReservationSet, participants []uuid.UUID, desired []CalendarAvailabilityReservationOccurrence) error {
	minimum, maximum := calendarAvailabilityReservationBounds(desired)
	rows, err := backend.tx.CalendarAvailabilityReservation.Query().
		Where(
			calendaravailabilityreservation.OrganizationIDEQ(set.OrganizationID()),
			calendaravailabilityreservation.UserIDIn(participants...),
			calendaravailabilityreservation.CalendarEventIDNEQ(set.EventID()),
			calendaravailabilityreservation.StateEQ(calendaravailabilityreservation.StateActive),
			calendaravailabilityreservation.StartsAtLT(maximum),
			calendaravailabilityreservation.EndsAtGT(minimum),
		).
		Order(ent.Asc(calendaravailabilityreservation.FieldUserID), ent.Asc(calendaravailabilityreservation.FieldStartsAt), ent.Asc(calendaravailabilityreservation.FieldID)).
		ForUpdate().
		Limit(calendarAvailabilityReservationMaxRows + 1).
		All(ctx)
	if err != nil {
		return calendarEntSafeError(err)
	}
	if len(rows) > calendarAvailabilityReservationMaxRows {
		return ErrAvailabilityLimitExceeded
	}
	for _, row := range rows {
		if row == nil || row.OrganizationID != set.OrganizationID() || row.UserID == uuid.Nil || row.CalendarEventID == uuid.Nil || row.State != calendaravailabilityreservation.StateActive || row.ReleasedAt != nil {
			return ErrAvailabilitySourceInvalid
		}
		startsAt, endsAt, ok := canonicalCalendarAvailabilityReservationInterval(row.StartsAt, row.EndsAt)
		if !ok {
			return ErrAvailabilitySourceInvalid
		}
		if calendarAvailabilityReservationOverlapsAny(startsAt, endsAt, desired) {
			return ErrAvailabilitySlotUnavailable
		}
	}
	return nil
}

func (backend *eventEntTransactionBackend) checkCalendarAvailabilityReservationCalls(ctx context.Context, participants []uuid.UUID, desired []CalendarAvailabilityReservationOccurrence) error {
	minimum, maximum := calendarAvailabilityReservationBounds(desired)
	calls, err := backend.tx.ScheduledCall.Query().
		Where(
			scheduledcall.IsActive(true),
			scheduledcall.DeletedAtIsNil(),
			scheduledcall.Not(scheduledcall.HasCalendarEvent()),
			scheduledcall.Not(scheduledcall.HasCalendarEventReference()),
			scheduledcall.ScheduledAtLT(maximum),
			scheduledcall.Or(
				scheduledcall.RruleNotNil(),
				scheduledcall.ScheduledAtGTE(minimum.Add(-time.Duration(eventCallDueMaxDurationMinutes)*time.Minute)),
			),
			scheduledcall.Or(
				scheduledcall.CreatedByIn(participants...),
				scheduledcall.HasParticipantsWith(user.IDIn(participants...)),
			),
		).
		WithParticipants(func(query *ent.UserQuery) {
			query.Where(user.IDIn(participants...))
		}).
		Order(ent.Asc(scheduledcall.FieldScheduledAt), ent.Asc(scheduledcall.FieldID)).
		ForUpdate().
		Limit(calendarAvailabilityReservationMaxRows + 1).
		All(ctx)
	if err != nil {
		return calendarEntSafeError(err)
	}
	if len(calls) > calendarAvailabilityReservationMaxRows {
		return ErrAvailabilityLimitExceeded
	}
	participantSet := make(map[uuid.UUID]struct{}, len(participants))
	for _, participantID := range participants {
		participantSet[participantID] = struct{}{}
	}
	for _, call := range calls {
		if call == nil || call.ID == uuid.Nil || call.DeletedAt != nil || !call.IsActive || call.CreatedBy == uuid.Nil || call.DurationMinutes < 1 || call.DurationMinutes > eventCallDueMaxDurationMinutes {
			return ErrAvailabilitySourceInvalid
		}
		startsAt, startsOK := canonicalAvailabilityTime(call.ScheduledAt)
		endsAt, endsOK := canonicalAvailabilityTime(startsAt.Add(time.Duration(call.DurationMinutes) * time.Minute))
		if !startsOK || !endsOK || !endsAt.After(startsAt) {
			return ErrAvailabilitySourceInvalid
		}
		temporal := CalendarEvent{StartsAt: startsAt, EndsAt: endsAt, Timezone: "UTC"}
		if temporal.Validate() != nil {
			return ErrAvailabilitySourceInvalid
		}
		occurrences, kind := eventCallProjectionNominalOccurrences(EventCallProjectionEvent{Temporal: temporal, RRULE: call.Rrule})
		if kind != "" {
			return ErrAvailabilitySourceInvalid
		}
		overlaps, overlapErr := calendarAvailabilityReservationEventCallOccurrencesOverlap(occurrences, desired)
		if overlapErr != nil {
			return overlapErr
		}
		if _, organizerBusy := participantSet[call.CreatedBy]; organizerBusy && overlaps {
			return ErrAvailabilitySlotUnavailable
		}
		for _, participant := range call.Edges.Participants {
			if participant == nil || participant.ID == uuid.Nil {
				return ErrAvailabilitySourceInvalid
			}
			if _, included := participantSet[participant.ID]; !included {
				return ErrAvailabilitySourceInvalid
			}
			if overlaps {
				return ErrAvailabilitySlotUnavailable
			}
		}
	}
	return nil
}

func (backend *eventEntTransactionBackend) releaseObsoleteCalendarAvailabilityReservations(ctx context.Context, active []*ent.CalendarAvailabilityReservation, desired map[calendarAvailabilityReservationIdentity]CalendarAvailabilityReservationOccurrence, releasedAt time.Time) error {
	for _, row := range active {
		identity := calendarAvailabilityReservationIdentity{userID: row.UserID, occurrenceKey: row.OccurrenceKey}
		if _, wanted := desired[identity]; wanted {
			continue
		}
		if row.Version < 1 || row.Version == int(^uint(0)>>1) {
			return ErrAvailabilitySourceInvalid
		}
		if err := backend.tx.CalendarAvailabilityReservation.UpdateOneID(row.ID).
			Where(calendaravailabilityreservation.VersionEQ(row.Version), calendaravailabilityreservation.StateEQ(calendaravailabilityreservation.StateActive)).
			SetState(calendaravailabilityreservation.StateReleased).
			SetReleasedAt(releasedAt).
			SetVersion(row.Version + 1).
			Exec(ctx); err != nil {
			return calendarEntCASFailure(err)
		}
	}
	return nil
}

func (backend *eventEntTransactionBackend) insertMissingCalendarAvailabilityReservations(ctx context.Context, set CalendarAvailabilityReservationSet, desired map[calendarAvailabilityReservationIdentity]CalendarAvailabilityReservationOccurrence) error {
	existing, err := backend.activeCalendarAvailabilityReservationIdentities(ctx, set)
	if err != nil {
		return err
	}
	identities := make([]calendarAvailabilityReservationIdentity, 0, len(desired))
	for identity := range desired {
		identities = append(identities, identity)
	}
	sort.Slice(identities, func(left, right int) bool {
		if identities[left].userID != identities[right].userID {
			return identities[left].userID.String() < identities[right].userID.String()
		}
		return identities[left].occurrenceKey < identities[right].occurrenceKey
	})
	for _, identity := range identities {
		if _, found := existing[identity]; found {
			continue
		}
		occurrence := desired[identity]
		if _, err := backend.tx.CalendarAvailabilityReservation.Create().
			SetOrganizationID(set.OrganizationID()).
			SetUserID(identity.userID).
			SetCalendarEventID(set.EventID()).
			SetOccurrenceKey(occurrence.OccurrenceKey).
			SetStartsAt(occurrence.StartsAt).
			SetEndsAt(occurrence.EndsAt).
			SetState(calendaravailabilityreservation.StateActive).
			SetVersion(1).
			Save(ctx); err != nil {
			return calendarEntSafeError(err)
		}
	}
	return nil
}

func (backend *eventEntTransactionBackend) activeCalendarAvailabilityReservationIdentities(ctx context.Context, set CalendarAvailabilityReservationSet) (map[calendarAvailabilityReservationIdentity]struct{}, error) {
	rows, err := backend.tx.CalendarAvailabilityReservation.Query().
		Where(
			calendaravailabilityreservation.OrganizationIDEQ(set.OrganizationID()),
			calendaravailabilityreservation.CalendarEventIDEQ(set.EventID()),
			calendaravailabilityreservation.StateEQ(calendaravailabilityreservation.StateActive),
		).
		Order(ent.Asc(calendaravailabilityreservation.FieldUserID), ent.Asc(calendaravailabilityreservation.FieldOccurrenceKey), ent.Asc(calendaravailabilityreservation.FieldID)).
		ForUpdate().
		Limit(calendarAvailabilityReservationMaxRows + 1).
		All(ctx)
	if err != nil {
		return nil, calendarEntSafeError(err)
	}
	if len(rows) > calendarAvailabilityReservationMaxRows {
		return nil, ErrAvailabilityLimitExceeded
	}
	identities := make(map[calendarAvailabilityReservationIdentity]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || row.UserID == uuid.Nil || row.OccurrenceKey == "" {
			return nil, ErrAvailabilitySourceInvalid
		}
		identity := calendarAvailabilityReservationIdentity{userID: row.UserID, occurrenceKey: row.OccurrenceKey}
		if _, duplicate := identities[identity]; duplicate {
			return nil, ErrAvailabilitySourceInvalid
		}
		identities[identity] = struct{}{}
	}
	return identities, nil
}

func calendarAvailabilityReservationBounds(occurrences []CalendarAvailabilityReservationOccurrence) (time.Time, time.Time) {
	minimum, maximum := occurrences[0].StartsAt, occurrences[0].EndsAt
	for _, occurrence := range occurrences[1:] {
		if occurrence.StartsAt.Before(minimum) {
			minimum = occurrence.StartsAt
		}
		if occurrence.EndsAt.After(maximum) {
			maximum = occurrence.EndsAt
		}
	}
	return minimum, maximum
}

func calendarAvailabilityReservationMemberIDs(members map[uuid.UUID]*ent.User) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(members))
	for memberID := range members {
		ids = append(ids, memberID)
	}
	sort.Slice(ids, func(left, right int) bool { return ids[left].String() < ids[right].String() })
	return ids
}

func calendarAvailabilityReservationOverlapsAny(from, to time.Time, occurrences []CalendarAvailabilityReservationOccurrence) bool {
	for _, occurrence := range occurrences {
		if from.Before(occurrence.EndsAt) && occurrence.StartsAt.Before(to) {
			return true
		}
	}
	return false
}

func calendarAvailabilityReservationOccurrenceOverlaps(occurrences []availabilityOccurrence, desired []CalendarAvailabilityReservationOccurrence) []bool {
	ordered := make([]int, len(occurrences))
	for index := range occurrences {
		ordered[index] = index
	}
	sort.Slice(ordered, func(left, right int) bool {
		return occurrences[ordered[left]].from.Before(occurrences[ordered[right]].from)
	})
	matched := make([]bool, len(occurrences))
	desiredIndex := 0
	for _, index := range ordered {
		occurrence := occurrences[index]
		for desiredIndex < len(desired) && !desired[desiredIndex].EndsAt.After(occurrence.from) {
			desiredIndex++
		}
		for candidate := desiredIndex; candidate < len(desired) && desired[candidate].StartsAt.Before(occurrence.to); candidate++ {
			if occurrence.from.Before(desired[candidate].EndsAt) {
				matched[index] = true
				break
			}
		}
	}
	return matched
}

func calendarAvailabilityReservationEventCallOccurrencesOverlap(occurrences []EventCallProjectionOccurrence, desired []CalendarAvailabilityReservationOccurrence) (bool, error) {
	for _, occurrence := range occurrences {
		if occurrence.StartsAt.IsZero() || occurrence.EndsAt.IsZero() || !occurrence.EndsAt.After(occurrence.StartsAt) {
			return false, ErrAvailabilitySourceInvalid
		}
		if calendarAvailabilityReservationOverlapsAny(occurrence.StartsAt.UTC(), occurrence.EndsAt.UTC(), desired) {
			return true, nil
		}
	}
	return false, nil
}
