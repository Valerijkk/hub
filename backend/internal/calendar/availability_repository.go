package calendar

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendareventexception"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/leaverequest"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/scheduledcall"
	"github.com/valerii/hub/backend/ent/user"
)

const (
	availabilitySourceRelationLimit  = 4096
	availabilitySourceEventLimit     = AvailabilityMaxBusyIntervalsAll
	availabilitySourceAttendeeLimit  = AvailabilityMaxBusyIntervalsAll
	availabilitySourceExceptionLimit = AvailabilityMaxBusyIntervalsAll
	availabilitySourceCallLimit      = AvailabilityMaxBusyIntervalsAll
	availabilitySourceLeaveLimit     = AvailabilityMaxBusyIntervalsAll
	availabilitySourceDatePadding    = 2
	availabilitySourceDaySearch      = 48 * 60
)

// AvailabilitySourceQuery содержит только авторитетные границы и список целевых
// пользователей. ActorID передаётся отдельно из аутентифицированного контекста.
type AvailabilitySourceQuery struct {
	From    time.Time
	To      time.Time
	UserIDs []uuid.UUID
}

// AvailabilitySourceRepository загружает приватные интервалы только после
// полной авторизации всего набора пользователей.
type AvailabilitySourceRepository interface {
	LoadAvailability(context.Context, uuid.UUID, AvailabilitySourceQuery) ([]AvailabilityUserInput, error)
}

// EntAvailabilityRepository читает авторитетные Calendar, Call и HR источники
// в одной read-only Ent-транзакции.
type EntAvailabilityRepository struct {
	client                    *ent.Client
	preCutoverScheduledCalls bool
}

type availabilityRepositoryErrorKind string

const (
	availabilityRepositoryNotFound    availabilityRepositoryErrorKind = "not_found"
	availabilityRepositoryUnavailable availabilityRepositoryErrorKind = "unavailable"
)

// availabilityRepositoryError намеренно не раскрывает причину, пользователя,
// организацию, проект, канал или временной интервал.
type availabilityRepositoryError struct {
	kind availabilityRepositoryErrorKind
}

func (err availabilityRepositoryError) Error() string {
	switch err.kind {
	case availabilityRepositoryNotFound:
		return "calendar availability not found"
	default:
		return "calendar availability unavailable"
	}
}

func (err availabilityRepositoryError) Is(target error) bool {
	switch value := target.(type) {
	case availabilityRepositoryError:
		return err.kind == value.kind
	case *availabilityRepositoryError:
		return value != nil && err.kind == value.kind
	default:
		return false
	}
}

// ErrAvailabilityNotFound является единым generic-результатом для чужого,
// неактивного, удалённого или несвязанного целевого пользователя.
var ErrAvailabilityNotFound error = availabilityRepositoryError{kind: availabilityRepositoryNotFound}

var errAvailabilityUnavailable error = availabilityRepositoryError{kind: availabilityRepositoryUnavailable}

// NewEntAvailabilityRepository создаёт production-адаптер availability. Флаг
// должен быть явно задан composition-слоем на время bridge/cutover.
func NewEntAvailabilityRepository(client *ent.Client, preCutoverScheduledCalls bool) *EntAvailabilityRepository {
	return &EntAvailabilityRepository{
		client:                    client,
		preCutoverScheduledCalls: preCutoverScheduledCalls,
	}
}

var _ AvailabilitySourceRepository = (*EntAvailabilityRepository)(nil)

// LoadAvailability возвращает только доменные AvailabilityUserInput в порядке
// запроса. Ни public DTO, ни деталь источника не покидают репозиторий.
func (repository *EntAvailabilityRepository) LoadAvailability(ctx context.Context, actorID uuid.UUID, query AvailabilitySourceQuery) ([]AvailabilityUserInput, error) {
	if repository == nil || repository.client == nil || actorID == uuid.Nil {
		return nil, ErrAvailabilitySourceInvalid
	}

	normalized, err := normalizeAvailabilitySourceQuery(query)
	if err != nil {
		return nil, err
	}

	transaction, err := repository.client.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:   true,
	})
	if err != nil {
		return nil, errAvailabilityUnavailable
	}
	committed := false
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = transaction.Rollback()
			panic(recovered)
		}
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	inputs, err := repository.loadAvailability(ctx, transaction, actorID, normalized)
	if err != nil {
		return nil, err
	}
	if err := transaction.Commit(); err != nil {
		return nil, errAvailabilityUnavailable
	}
	committed = true
	return inputs, nil
}

func normalizeAvailabilitySourceQuery(query AvailabilitySourceQuery) (AvailabilitySourceQuery, error) {
	from, ok := canonicalAvailabilityTime(query.From)
	if !ok {
		return AvailabilitySourceQuery{}, ErrAvailabilitySourceInvalid
	}
	to, ok := canonicalAvailabilityTime(query.To)
	if !ok || !to.After(from) || to.Sub(from) > AvailabilityMaxWindow {
		return AvailabilitySourceQuery{}, ErrAvailabilitySourceInvalid
	}
	if len(query.UserIDs) == 0 || len(query.UserIDs) > AvailabilityMaxUsers {
		return AvailabilitySourceQuery{}, ErrAvailabilitySourceInvalid
	}

	seen := make(map[uuid.UUID]struct{}, len(query.UserIDs))
	userIDs := make([]uuid.UUID, len(query.UserIDs))
	for index, userID := range query.UserIDs {
		if userID == uuid.Nil {
			return AvailabilitySourceQuery{}, ErrAvailabilitySourceInvalid
		}
		if _, duplicate := seen[userID]; duplicate {
			return AvailabilitySourceQuery{}, ErrAvailabilitySourceInvalid
		}
		seen[userID] = struct{}{}
		userIDs[index] = userID
	}
	return AvailabilitySourceQuery{From: from, To: to, UserIDs: userIDs}, nil
}

func (repository *EntAvailabilityRepository) loadAvailability(ctx context.Context, transaction *ent.Tx, actorID uuid.UUID, query AvailabilitySourceQuery) ([]AvailabilityUserInput, error) {
	targets, actorProjects, targetProjects, err := availabilityAuthorizedTargets(ctx, transaction, actorID, query.UserIDs)
	if err != nil {
		return nil, err
	}

	accumulator := availabilitySourceAccumulator{
		users: make(map[uuid.UUID]*AvailabilityUserInput, len(query.UserIDs)),
	}
	for _, userID := range query.UserIDs {
		target, found := targets[userID]
		if !found || !availabilitySourceTimezoneValid(target.Timezone) {
			return nil, ErrAvailabilitySourceInvalid
		}
		accumulator.users[userID] = &AvailabilityUserInput{UserID: userID, Timezone: target.Timezone}
	}

	firstTarget := targets[query.UserIDs[0]]
	if firstTarget == nil || firstTarget.Edges.Organization == nil || firstTarget.Edges.Organization.ID == uuid.Nil {
		return nil, ErrAvailabilitySourceInvalid
	}
	organizationID := firstTarget.Edges.Organization.ID
	if err := availabilityLoadEventBusy(ctx, transaction, organizationID, query, &accumulator); err != nil {
		return nil, err
	}
	if err := availabilityLoadCallBusy(ctx, transaction, repository.preCutoverScheduledCalls, query, &accumulator); err != nil {
		return nil, err
	}
	if err := availabilityLoadLeaveBusy(ctx, transaction, actorID, actorProjects, targetProjects, query, &accumulator); err != nil {
		return nil, err
	}

	inputs := make([]AvailabilityUserInput, len(query.UserIDs))
	for index, userID := range query.UserIDs {
		input := accumulator.users[userID]
		if input == nil {
			return nil, ErrAvailabilitySourceInvalid
		}
		inputs[index] = AvailabilityUserInput{
			UserID:   input.UserID,
			Timezone: input.Timezone,
			Busy:     append([]AvailabilityBusyInterval(nil), input.Busy...),
		}
	}
	return inputs, nil
}

func availabilityAuthorizedTargets(ctx context.Context, transaction *ent.Tx, actorID uuid.UUID, targetIDs []uuid.UUID) (map[uuid.UUID]*ent.User, map[uuid.UUID]struct{}, map[uuid.UUID]map[uuid.UUID]struct{}, error) {
	actor, err := transaction.User.Query().
		Where(
			user.IDEQ(actorID),
			user.StatusEQ(user.StatusActive),
		).
		WithOrganization().
		Only(ctx)
	if err != nil || actor.Edges.Organization == nil || actor.Edges.Organization.ID == uuid.Nil {
		return nil, nil, nil, ErrAvailabilityNotFound
	}
	organizationID := actor.Edges.Organization.ID

	targetRows, err := transaction.User.Query().
		Where(
			user.IDIn(targetIDs...),
			user.StatusEQ(user.StatusActive),
			user.HasOrganizationWith(organization.IDEQ(organizationID)),
		).
		WithOrganization().
		Order(ent.Asc(user.FieldID)).
		Limit(len(targetIDs) + 1).
		All(ctx)
	if err != nil {
		return nil, nil, nil, errAvailabilityUnavailable
	}
	if len(targetRows) != len(targetIDs) {
		return nil, nil, nil, ErrAvailabilityNotFound
	}

	targets := make(map[uuid.UUID]*ent.User, len(targetRows))
	for _, target := range targetRows {
		if target == nil || target.ID == uuid.Nil || target.Edges.Organization == nil || target.Edges.Organization.ID != organizationID {
			return nil, nil, nil, ErrAvailabilityNotFound
		}
		if _, duplicate := targets[target.ID]; duplicate {
			return nil, nil, nil, ErrAvailabilityNotFound
		}
		targets[target.ID] = target
	}

	members := make([]uuid.UUID, 0, len(targetIDs)+1)
	members = append(members, actorID)
	for _, targetID := range targetIDs {
		if targetID != actorID {
			members = append(members, targetID)
		}
	}

	actorProjects, targetProjects, err := availabilitySharedProjects(ctx, transaction, organizationID, actorID, targets, members)
	if err != nil {
		return nil, nil, nil, err
	}
	actorChannels, targetChannels, err := availabilitySharedChannels(ctx, transaction, actorID, targets, members)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, targetID := range targetIDs {
		if targetID == actorID {
			continue
		}
		if availabilitySharesAny(actorProjects, targetProjects[targetID]) || availabilitySharesAny(actorChannels, targetChannels[targetID]) {
			continue
		}
		return nil, nil, nil, ErrAvailabilityNotFound
	}
	return targets, actorProjects, targetProjects, nil
}

func availabilitySharedProjects(ctx context.Context, transaction *ent.Tx, organizationID, actorID uuid.UUID, targets map[uuid.UUID]*ent.User, memberIDs []uuid.UUID) (map[uuid.UUID]struct{}, map[uuid.UUID]map[uuid.UUID]struct{}, error) {
	rows, err := transaction.ProjectMember.Query().
		Where(
			projectmember.HasUserWith(user.IDIn(memberIDs...)),
			projectmember.HasProjectWith(
				project.DeletedAtIsNil(),
				project.StatusNEQ(project.StatusDeleted),
				project.HasOrganizationWith(organization.IDEQ(organizationID)),
			),
		).
		WithUser().
		WithProject().
		Order(ent.Asc(projectmember.FieldID)).
		Limit(availabilitySourceRelationLimit + 1).
		All(ctx)
	if err != nil {
		return nil, nil, errAvailabilityUnavailable
	}
	if len(rows) > availabilitySourceRelationLimit {
		return nil, nil, ErrAvailabilityLimitExceeded
	}

	actorProjects := make(map[uuid.UUID]struct{})
	targetProjects := make(map[uuid.UUID]map[uuid.UUID]struct{}, len(targets))
	for targetID := range targets {
		targetProjects[targetID] = make(map[uuid.UUID]struct{})
	}
	for _, row := range rows {
		if row == nil || row.Edges.User == nil || row.Edges.Project == nil || row.Edges.User.ID == uuid.Nil || row.Edges.Project.ID == uuid.Nil {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		memberID, projectID := row.Edges.User.ID, row.Edges.Project.ID
		if memberID == actorID {
			actorProjects[projectID] = struct{}{}
			continue
		}
		if values, target := targetProjects[memberID]; target {
			values[projectID] = struct{}{}
		}
	}
	return actorProjects, targetProjects, nil
}

func availabilitySharedChannels(ctx context.Context, transaction *ent.Tx, actorID uuid.UUID, targets map[uuid.UUID]*ent.User, memberIDs []uuid.UUID) (map[uuid.UUID]struct{}, map[uuid.UUID]map[uuid.UUID]struct{}, error) {
	rows, err := transaction.ChannelMember.Query().
		Where(
			channelmember.HasUserWith(user.IDIn(memberIDs...)),
			channelmember.HasChannelWith(channel.DeletedAtIsNil()),
		).
		WithUser().
		WithChannel().
		Order(ent.Asc(channelmember.FieldID)).
		Limit(availabilitySourceRelationLimit + 1).
		All(ctx)
	if err != nil {
		return nil, nil, errAvailabilityUnavailable
	}
	if len(rows) > availabilitySourceRelationLimit {
		return nil, nil, ErrAvailabilityLimitExceeded
	}

	actorChannels := make(map[uuid.UUID]struct{})
	targetChannels := make(map[uuid.UUID]map[uuid.UUID]struct{}, len(targets))
	for targetID := range targets {
		targetChannels[targetID] = make(map[uuid.UUID]struct{})
	}
	for _, row := range rows {
		if row == nil || row.Edges.User == nil || row.Edges.Channel == nil || row.Edges.User.ID == uuid.Nil || row.Edges.Channel.ID == uuid.Nil {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		memberID, channelID := row.Edges.User.ID, row.Edges.Channel.ID
		if memberID == actorID {
			actorChannels[channelID] = struct{}{}
			continue
		}
		if values, target := targetChannels[memberID]; target {
			values[channelID] = struct{}{}
		}
	}
	return actorChannels, targetChannels, nil
}

func availabilitySharesAny(left, right map[uuid.UUID]struct{}) bool {
	for value := range left {
		if _, found := right[value]; found {
			return true
		}
	}
	return false
}

type availabilitySourceAccumulator struct {
	users map[uuid.UUID]*AvailabilityUserInput
	total int
}

func (accumulator *availabilitySourceAccumulator) append(userID uuid.UUID, interval AvailabilityBusyInterval) error {
	if accumulator == nil || interval.Source.Validate() != nil {
		return ErrAvailabilitySourceInvalid
	}
	from, ok := canonicalAvailabilityTime(interval.From)
	if !ok {
		return ErrAvailabilitySourceInvalid
	}
	to, ok := canonicalAvailabilityTime(interval.To)
	if !ok || !to.After(from) {
		return ErrAvailabilitySourceInvalid
	}
	user, found := accumulator.users[userID]
	if !found || user == nil {
		return ErrAvailabilitySourceInvalid
	}
	if len(user.Busy) >= AvailabilityMaxBusyIntervalsUser || accumulator.total >= AvailabilityMaxBusyIntervalsAll {
		return ErrAvailabilityLimitExceeded
	}
	user.Busy = append(user.Busy, AvailabilityBusyInterval{From: from, To: to, Source: interval.Source})
	accumulator.total++
	return nil
}

func availabilityLoadEventBusy(ctx context.Context, transaction *ent.Tx, organizationID uuid.UUID, query AvailabilitySourceQuery, accumulator *availabilitySourceAccumulator) error {
	lowerDate, upperDate := availabilitySourceDateBounds(query.From, query.To)
	events, err := transaction.CalendarEvent.Query().
		Where(
			calendarevent.OrganizationIDEQ(organizationID),
			calendarevent.DeletedAtIsNil(),
			calendarevent.StatusIn(calendarevent.StatusConfirmed, calendarevent.StatusTentative),
			calendarevent.Or(
				calendarevent.CalendarOwnerIDIn(query.UserIDs...),
				calendarevent.OrganizerIDIn(query.UserIDs...),
				calendarevent.HasAttendeesWith(
					calendarattendee.UserIDIn(query.UserIDs...),
					calendarattendee.RemovedAtIsNil(),
				),
			),
			calendarevent.Or(
				calendarevent.And(
					calendarevent.AllDay(false),
					calendarevent.StartsAtLT(query.To),
					calendarevent.Or(
						calendarevent.RruleNotNil(),
						calendarevent.EndsAtGT(query.From),
					),
				),
				calendarevent.And(
					calendarevent.AllDay(true),
					calendarevent.StartDateLT(upperDate),
					calendarevent.Or(
						calendarevent.RruleNotNil(),
						calendarevent.EndDateExclusiveGT(lowerDate),
					),
				),
			),
		).
		Order(ent.Asc(calendarevent.FieldID)).
		Limit(availabilitySourceEventLimit + 1).
		All(ctx)
	if err != nil {
		return errAvailabilityUnavailable
	}
	if len(events) > availabilitySourceEventLimit {
		return ErrAvailabilityLimitExceeded
	}
	if len(events) == 0 {
		return nil
	}

	eventIDs := make([]uuid.UUID, len(events))
	for index, event := range events {
		if event == nil || event.ID == uuid.Nil || event.OrganizationID != organizationID {
			return ErrAvailabilitySourceInvalid
		}
		eventIDs[index] = event.ID
	}

	attendees, err := transaction.CalendarAttendee.Query().
		Where(
			calendarattendee.CalendarEventIDIn(eventIDs...),
			calendarattendee.UserIDIn(query.UserIDs...),
			calendarattendee.RemovedAtIsNil(),
		).
		Order(ent.Asc(calendarattendee.FieldCalendarEventID), ent.Asc(calendarattendee.FieldRecurrenceID), ent.Asc(calendarattendee.FieldID)).
		Limit(availabilitySourceAttendeeLimit + 1).
		All(ctx)
	if err != nil {
		return errAvailabilityUnavailable
	}
	if len(attendees) > availabilitySourceAttendeeLimit {
		return ErrAvailabilityLimitExceeded
	}

	exceptions, err := transaction.CalendarEventException.Query().
		Where(calendareventexception.CalendarEventIDIn(eventIDs...)).
		Order(ent.Asc(calendareventexception.FieldCalendarEventID), ent.Asc(calendareventexception.FieldRecurrenceID), ent.Asc(calendareventexception.FieldID)).
		Limit(availabilitySourceExceptionLimit + 1).
		All(ctx)
	if err != nil {
		return errAvailabilityUnavailable
	}
	if len(exceptions) > availabilitySourceExceptionLimit {
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

	for _, event := range events {
		occurrences, nominal, recurring, err := availabilityEventOccurrences(event, exceptionsByEvent[event.ID])
		if err != nil {
			return err
		}
		attendeeStates, err := availabilityEventAttendeeStates(attendeesByEvent[event.ID], nominal, recurring, accumulator.users)
		if err != nil {
			return err
		}

		for _, occurrence := range occurrences {
			if !occurrence.from.Before(query.To) || !occurrence.to.After(query.From) {
				continue
			}
			for _, userID := range query.UserIDs {
				busy := event.CalendarOwnerID == userID || event.OrganizerID != nil && *event.OrganizerID == userID
				if state, found := attendeeStates[userID]; found {
					status := state.series
					if occurrenceStatus, override := state.occurrences[occurrence.key]; override {
						status = occurrenceStatus
					}
					busy = busy || availabilityBusyRSVPStatus(status)
				}
				if busy {
					if err := accumulator.append(userID, AvailabilityBusyInterval{From: occurrence.from, To: occurrence.to, Source: AvailabilitySourceEvent}); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

type availabilityAttendeeState struct {
	series      RSVPStatus
	hasSeries   bool
	occurrences map[string]RSVPStatus
}

func availabilityEventAttendeeStates(rows []*ent.CalendarAttendee, nominal map[string]struct{}, recurring bool, targets map[uuid.UUID]*AvailabilityUserInput) (map[uuid.UUID]availabilityAttendeeState, error) {
	states := make(map[uuid.UUID]availabilityAttendeeState, len(rows))
	for _, attendee := range rows {
		if attendee == nil || attendee.UserID == nil || *attendee.UserID == uuid.Nil || attendee.RemovedAt != nil {
			return nil, ErrAvailabilitySourceInvalid
		}
		if _, target := targets[*attendee.UserID]; !target {
			return nil, ErrAvailabilitySourceInvalid
		}
		status := RSVPStatus(attendee.RsvpStatus)
		if status.Validate() != nil {
			return nil, ErrAvailabilitySourceInvalid
		}
		state := states[*attendee.UserID]
		if attendee.RecurrenceID == nil {
			if state.hasSeries {
				return nil, ErrAvailabilitySourceInvalid
			}
			state.series, state.hasSeries = status, true
			states[*attendee.UserID] = state
			continue
		}
		if !recurring || *attendee.RecurrenceID == "" {
			return nil, ErrAvailabilitySourceInvalid
		}
		if _, found := nominal[*attendee.RecurrenceID]; !found {
			return nil, ErrAvailabilitySourceInvalid
		}
		if state.occurrences == nil {
			state.occurrences = make(map[string]RSVPStatus)
		}
		if _, duplicate := state.occurrences[*attendee.RecurrenceID]; duplicate {
			return nil, ErrAvailabilitySourceInvalid
		}
		state.occurrences[*attendee.RecurrenceID] = status
		states[*attendee.UserID] = state
	}
	for _, state := range states {
		if len(state.occurrences) != 0 && !state.hasSeries {
			return nil, ErrAvailabilitySourceInvalid
		}
	}
	return states, nil
}

func availabilityBusyRSVPStatus(status RSVPStatus) bool {
	return status == RSVPNeedsAction || status == RSVPAccepted || status == RSVPTentative
}

type availabilityOccurrence struct {
	key  string
	from time.Time
	to   time.Time
}

func availabilityEventOccurrences(event *ent.CalendarEvent, exceptions []*ent.CalendarEventException) ([]availabilityOccurrence, map[string]struct{}, bool, error) {
	if event == nil || event.ID == uuid.Nil || event.CalendarOwnerID == uuid.Nil || event.DeletedAt != nil || CalendarEventStatus(event.Status) == CalendarEventStatusCancelled {
		return nil, nil, false, ErrAvailabilitySourceInvalid
	}
	temporal, err := availabilityCalendarEventTemporal(event)
	if err != nil {
		return nil, nil, false, err
	}
	rule := ""
	if event.Rrule != nil {
		rule = *event.Rrule
	}
	if temporal.AllDay {
		occurrences, nominal, err := availabilityAllDayEventOccurrences(temporal, rule, event.Exdates, exceptions, event.ID)
		return occurrences, nominal, rule != "", err
	}
	occurrences, nominal, err := availabilityTimedEventOccurrences(temporal, rule, event.Exdates, exceptions, event.ID)
	return occurrences, nominal, rule != "", err
}

func availabilityCalendarEventTemporal(event *ent.CalendarEvent) (CalendarEvent, error) {
	if event == nil {
		return CalendarEvent{}, ErrAvailabilitySourceInvalid
	}
	if event.AllDay {
		if event.StartsAt != nil || event.EndsAt != nil || event.StartDate == nil || event.EndDateExclusive == nil {
			return CalendarEvent{}, ErrAvailabilitySourceInvalid
		}
		startDate, err := availabilityDateLabel(*event.StartDate)
		if err != nil {
			return CalendarEvent{}, err
		}
		endDate, err := availabilityDateLabel(*event.EndDateExclusive)
		if err != nil {
			return CalendarEvent{}, err
		}
		temporal := CalendarEvent{AllDay: true, StartDate: startDate.Format(time.DateOnly), EndDateExclusive: endDate.Format(time.DateOnly), Timezone: event.Timezone}
		if temporal.Validate() != nil {
			return CalendarEvent{}, ErrAvailabilitySourceInvalid
		}
		return temporal, nil
	}
	if event.StartsAt == nil || event.EndsAt == nil || event.StartDate != nil || event.EndDateExclusive != nil {
		return CalendarEvent{}, ErrAvailabilitySourceInvalid
	}
	startsAt, ok := canonicalAvailabilityTime(*event.StartsAt)
	if !ok {
		return CalendarEvent{}, ErrAvailabilitySourceInvalid
	}
	endsAt, ok := canonicalAvailabilityTime(*event.EndsAt)
	if !ok {
		return CalendarEvent{}, ErrAvailabilitySourceInvalid
	}
	temporal := CalendarEvent{StartsAt: startsAt, EndsAt: endsAt, Timezone: event.Timezone}
	if temporal.Validate() != nil {
		return CalendarEvent{}, ErrAvailabilitySourceInvalid
	}
	return temporal, nil
}

func availabilityExceptionTemporal(exception *ent.CalendarEventException) (*CalendarEvent, error) {
	if exception == nil {
		return nil, ErrAvailabilitySourceInvalid
	}
	hasTimed := exception.StartsAt != nil || exception.EndsAt != nil
	hasAllDay := exception.StartDate != nil || exception.EndDateExclusive != nil
	if !hasTimed && !hasAllDay && exception.Timezone == nil {
		return nil, nil
	}
	if hasTimed == hasAllDay || exception.Timezone == nil {
		return nil, ErrAvailabilitySourceInvalid
	}
	if hasTimed {
		if exception.StartsAt == nil || exception.EndsAt == nil {
			return nil, ErrAvailabilitySourceInvalid
		}
		startsAt, ok := canonicalAvailabilityTime(*exception.StartsAt)
		if !ok {
			return nil, ErrAvailabilitySourceInvalid
		}
		endsAt, ok := canonicalAvailabilityTime(*exception.EndsAt)
		if !ok {
			return nil, ErrAvailabilitySourceInvalid
		}
		temporal := CalendarEvent{StartsAt: startsAt, EndsAt: endsAt, Timezone: *exception.Timezone}
		if temporal.Validate() != nil {
			return nil, ErrAvailabilitySourceInvalid
		}
		return &temporal, nil
	}
	if exception.StartDate == nil || exception.EndDateExclusive == nil {
		return nil, ErrAvailabilitySourceInvalid
	}
	startDate, err := availabilityDateLabel(*exception.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := availabilityDateLabel(*exception.EndDateExclusive)
	if err != nil {
		return nil, err
	}
	temporal := CalendarEvent{AllDay: true, StartDate: startDate.Format(time.DateOnly), EndDateExclusive: endDate.Format(time.DateOnly), Timezone: *exception.Timezone}
	if temporal.Validate() != nil {
		return nil, ErrAvailabilitySourceInvalid
	}
	return &temporal, nil
}

func availabilityTimedEventOccurrences(temporal CalendarEvent, rule string, exdates []string, exceptions []*ent.CalendarEventException, eventID uuid.UUID) ([]availabilityOccurrence, map[string]struct{}, error) {
	projection := EventCallProjectionEvent{Temporal: temporal, RRULE: rule, Exdates: append([]string(nil), exdates...)}
	nominalRows, kind := eventCallProjectionNominalOccurrences(projection)
	if kind != "" {
		return nil, nil, ErrAvailabilitySourceInvalid
	}
	nominal := make(map[string]availabilityOccurrence, len(nominalRows))
	for _, occurrence := range nominalRows {
		if occurrence.Key == "" || !occurrence.StartsAt.Before(occurrence.EndsAt) {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		nominal[occurrence.Key] = availabilityOccurrence{key: occurrence.Key, from: occurrence.StartsAt.UTC(), to: occurrence.EndsAt.UTC()}
	}
	excluded, overrides, err := availabilityEventExceptions(exdates, exceptions, eventID, nominal)
	if err != nil {
		return nil, nil, err
	}

	occurrences := make([]availabilityOccurrence, 0, len(nominalRows))
	starts := make(map[time.Time]struct{}, len(nominalRows))
	for _, nominalRow := range nominalRows {
		if _, omitted := excluded[nominalRow.Key]; omitted {
			continue
		}
		occurrence := nominal[nominalRow.Key]
		if override, found := overrides[nominalRow.Key]; found {
			if override.cancelled {
				continue
			}
			if override.temporal != nil {
				from, to, err := availabilityTemporalInterval(*override.temporal)
				if err != nil {
					return nil, nil, err
				}
				occurrence.from, occurrence.to = from, to
			}
		}
		if _, duplicate := starts[occurrence.from]; duplicate {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		starts[occurrence.from] = struct{}{}
		occurrences = append(occurrences, occurrence)
	}
	return occurrences, availabilityOccurrenceKeys(nominal), nil
}

type availabilityExceptionOverride struct {
	cancelled bool
	temporal  *CalendarEvent
}

func availabilityEventExceptions(exdates []string, exceptions []*ent.CalendarEventException, eventID uuid.UUID, nominal map[string]availabilityOccurrence) (map[string]struct{}, map[string]availabilityExceptionOverride, error) {
	excluded := make(map[string]struct{}, len(exdates))
	previous := ""
	for _, value := range exdates {
		if value == "" || value <= previous {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		if _, exists := nominal[value]; !exists {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		excluded[value] = struct{}{}
		previous = value
	}

	overrides := make(map[string]availabilityExceptionOverride, len(exceptions))
	for _, exception := range exceptions {
		if exception == nil || exception.CalendarEventID != eventID || exception.RecurrenceID == "" || exception.Status == nil {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		if _, exists := nominal[exception.RecurrenceID]; !exists {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		if _, excludedKey := excluded[exception.RecurrenceID]; excludedKey {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		if _, duplicate := overrides[exception.RecurrenceID]; duplicate {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		status := CalendarEventStatus(*exception.Status)
		if status != CalendarEventStatusConfirmed && status != CalendarEventStatusTentative && status != CalendarEventStatusCancelled {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		temporal, err := availabilityExceptionTemporal(exception)
		if err != nil {
			return nil, nil, err
		}
		overrides[exception.RecurrenceID] = availabilityExceptionOverride{cancelled: status == CalendarEventStatusCancelled, temporal: temporal}
	}
	return excluded, overrides, nil
}

func availabilityOccurrenceKeys(values map[string]availabilityOccurrence) map[string]struct{} {
	keys := make(map[string]struct{}, len(values))
	for key := range values {
		keys[key] = struct{}{}
	}
	return keys
}

func availabilityAllDayEventOccurrences(temporal CalendarEvent, rule string, exdates []string, exceptions []*ent.CalendarEventException, eventID uuid.UUID) ([]availabilityOccurrence, map[string]struct{}, error) {
	dates, err := availabilityAllDayNominalDates(temporal, rule)
	if err != nil {
		return nil, nil, err
	}
	startDate, err := parseCalendarDate(temporal.StartDate)
	if err != nil {
		return nil, nil, ErrAvailabilitySourceInvalid
	}
	endDate, err := parseCalendarDate(temporal.EndDateExclusive)
	if err != nil || !endDate.After(startDate) {
		return nil, nil, ErrAvailabilitySourceInvalid
	}
	days := int(endDate.Sub(startDate) / (24 * time.Hour))
	if days < 1 {
		return nil, nil, ErrAvailabilitySourceInvalid
	}
	nominal := make(map[string]availabilityOccurrence, len(dates))
	for _, date := range dates {
		key := date.Format(time.DateOnly)
		from, to, err := availabilityAllDayDateInterval(date, date.AddDate(0, 0, days), temporal.Timezone)
		if err != nil {
			return nil, nil, err
		}
		if _, duplicate := nominal[key]; duplicate {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		nominal[key] = availabilityOccurrence{key: key, from: from, to: to}
	}
	excluded, overrides, err := availabilityEventExceptions(exdates, exceptions, eventID, nominal)
	if err != nil {
		return nil, nil, err
	}

	occurrences := make([]availabilityOccurrence, 0, len(dates))
	starts := make(map[time.Time]struct{}, len(dates))
	for _, date := range dates {
		key := date.Format(time.DateOnly)
		if _, omitted := excluded[key]; omitted {
			continue
		}
		occurrence := nominal[key]
		if override, found := overrides[key]; found {
			if override.cancelled {
				continue
			}
			if override.temporal != nil {
				from, to, err := availabilityTemporalInterval(*override.temporal)
				if err != nil {
					return nil, nil, err
				}
				occurrence.from, occurrence.to = from, to
			}
		}
		if _, duplicate := starts[occurrence.from]; duplicate {
			return nil, nil, ErrAvailabilitySourceInvalid
		}
		starts[occurrence.from] = struct{}{}
		occurrences = append(occurrences, occurrence)
	}
	return occurrences, availabilityOccurrenceKeys(nominal), nil
}

func availabilityAllDayNominalDates(temporal CalendarEvent, rule string) ([]time.Time, error) {
	startDate, err := parseCalendarDate(temporal.StartDate)
	if err != nil {
		return nil, ErrAvailabilitySourceInvalid
	}
	if rule == "" {
		return []time.Time{startDate}, nil
	}
	normalized, err := normalizeCalendarCreateRRULE(rule, temporal)
	if err != nil || normalized != rule {
		return nil, ErrAvailabilitySourceInvalid
	}
	fields := eventCallProjectionRRuleFields(rule)
	frequency := fields["FREQ"]
	interval, err := parseCalendarCreatePositiveInteger(fields["INTERVAL"], 365)
	if err != nil {
		return nil, ErrAvailabilitySourceInvalid
	}
	count := 0
	if value, found := fields["COUNT"]; found {
		count, err = parseCalendarCreatePositiveInteger(value, calendarCreateOccurrenceCap)
		if err != nil {
			return nil, ErrAvailabilitySourceInvalid
		}
	}
	var until time.Time
	hasUntil := false
	if value, found := fields["UNTIL"]; found {
		until, err = time.Parse("20060102", value)
		if err != nil || until.Format("20060102") != value || until.Before(startDate) {
			return nil, ErrAvailabilitySourceInvalid
		}
		hasUntil = true
	}

	values := []time.Time{startDate}
	work := 0
	appendDate := func(candidate time.Time) (bool, error) {
		if !candidate.After(startDate) {
			return false, nil
		}
		if count > 0 && len(values) >= count {
			return true, nil
		}
		if hasUntil && candidate.After(until) {
			return true, nil
		}
		work++
		if work > calendarCreateCandidateWorkGuard || len(values) >= calendarCreateOccurrenceCap {
			return false, ErrAvailabilityLimitExceeded
		}
		values = append(values, candidate)
		return false, nil
	}

	switch frequency {
	case "DAILY":
		for sequence := 1; ; sequence++ {
			stop, err := appendDate(startDate.AddDate(0, 0, sequence*interval))
			if err != nil || stop {
				return values, err
			}
		}
	case "WEEKLY":
		weekdays, err := calendarCreateWeekdayOffsets(fields["BYDAY"])
		if err != nil {
			return nil, ErrAvailabilitySourceInvalid
		}
		weekStart := calendarCreateMonday(startDate)
		for group := 0; ; group++ {
			base := weekStart.AddDate(0, 0, group*interval*7)
			for _, weekday := range weekdays {
				stop, err := appendDate(base.AddDate(0, 0, weekday))
				if err != nil || stop {
					return values, err
				}
			}
		}
	case "MONTHLY":
		monthDay, err := parseCalendarCreatePositiveInteger(fields["BYMONTHDAY"], 31)
		if err != nil {
			return nil, ErrAvailabilitySourceInvalid
		}
		for sequence := 0; ; sequence++ {
			month := time.Date(startDate.Year(), startDate.Month()+time.Month(sequence*interval), 1, 0, 0, 0, 0, time.UTC)
			if monthDay > calendarCreateDaysInMonth(month.Year(), month.Month()) {
				work++
				if work > calendarCreateCandidateWorkGuard {
					return nil, ErrAvailabilityLimitExceeded
				}
				if hasUntil && month.After(until) {
					return values, nil
				}
				continue
			}
			candidate := time.Date(month.Year(), month.Month(), monthDay, 0, 0, 0, 0, time.UTC)
			stop, err := appendDate(candidate)
			if err != nil || stop {
				return values, err
			}
		}
	default:
		return nil, ErrAvailabilitySourceInvalid
	}
}

func availabilityLoadCallBusy(ctx context.Context, transaction *ent.Tx, preCutover bool, query AvailabilitySourceQuery, accumulator *availabilitySourceAccumulator) error {
	calls, err := transaction.ScheduledCall.Query().
		Where(
			scheduledcall.IsActive(true),
			scheduledcall.DeletedAtIsNil(),
			scheduledcall.Not(scheduledcall.HasCalendarEventLink()),
			scheduledcall.ScheduledAtLT(query.To),
			scheduledcall.Or(
				scheduledcall.RruleNotNil(),
				scheduledcall.ScheduledAtGTE(query.From.Add(-time.Duration(eventCallDueMaxDurationMinutes) * time.Minute)),
			),
			scheduledcall.Or(
				scheduledcall.CreatedByIn(query.UserIDs...),
				scheduledcall.HasParticipantsWith(user.IDIn(query.UserIDs...)),
			),
		).
		WithParticipants(func(participants *ent.UserQuery) {
			participants.Where(user.IDIn(query.UserIDs...))
		}).
		Order(ent.Asc(scheduledcall.FieldScheduledAt), ent.Asc(scheduledcall.FieldID)).
		Limit(availabilitySourceCallLimit + 1).
		All(ctx)
	if err != nil {
		return errAvailabilityUnavailable
	}
	if len(calls) > availabilitySourceCallLimit {
		return ErrAvailabilityLimitExceeded
	}
	if !preCutover && len(calls) != 0 {
		return ErrAvailabilitySourceInvalid
	}
	if !preCutover {
		return nil
	}

	for _, call := range calls {
		if call == nil || call.ID == uuid.Nil || call.DeletedAt != nil || !call.IsActive || call.CreatedBy == uuid.Nil || call.DurationMinutes < 1 || call.DurationMinutes > eventCallDueMaxDurationMinutes {
			return ErrAvailabilitySourceInvalid
		}
		startsAt, ok := canonicalAvailabilityTime(call.ScheduledAt)
		if !ok {
			return ErrAvailabilitySourceInvalid
		}
		endsAt := startsAt.Add(time.Duration(call.DurationMinutes) * time.Minute)
		if canonicalEndsAt, ok := canonicalAvailabilityTime(endsAt); !ok || !canonicalEndsAt.After(startsAt) {
			return ErrAvailabilitySourceInvalid
		} else {
			endsAt = canonicalEndsAt
		}
		temporal := CalendarEvent{StartsAt: startsAt, EndsAt: endsAt, Timezone: "UTC"}
		if temporal.Validate() != nil {
			return ErrAvailabilitySourceInvalid
		}
		occurrences, kind := eventCallProjectionNominalOccurrences(EventCallProjectionEvent{Temporal: temporal, RRULE: call.Rrule})
		if kind != "" {
			return ErrAvailabilitySourceInvalid
		}

		participants := make(map[uuid.UUID]struct{}, len(call.Edges.Participants)+1)
		if _, target := accumulator.users[call.CreatedBy]; target {
			participants[call.CreatedBy] = struct{}{}
		}
		for _, participant := range call.Edges.Participants {
			if participant == nil || participant.ID == uuid.Nil {
				return ErrAvailabilitySourceInvalid
			}
			if _, target := accumulator.users[participant.ID]; !target {
				return ErrAvailabilitySourceInvalid
			}
			participants[participant.ID] = struct{}{}
		}
		if len(participants) == 0 {
			return ErrAvailabilitySourceInvalid
		}

		for _, occurrence := range occurrences {
			if !occurrence.StartsAt.Before(query.To) || !occurrence.EndsAt.After(query.From) {
				continue
			}
			for userID := range participants {
				if err := accumulator.append(userID, AvailabilityBusyInterval{From: occurrence.StartsAt, To: occurrence.EndsAt, Source: AvailabilitySourceCall}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func availabilityLoadLeaveBusy(ctx context.Context, transaction *ent.Tx, actorID uuid.UUID, actorProjects map[uuid.UUID]struct{}, targetProjects map[uuid.UUID]map[uuid.UUID]struct{}, query AvailabilitySourceQuery, accumulator *availabilitySourceAccumulator) error {
	// Leave visibility is project-scoped, unlike the wider channel/project
	// availability authorization. Reuse the bounded non-deleted project relation
	// already resolved before every source query.
	visible := make([]uuid.UUID, 0, len(query.UserIDs))
	for _, userID := range query.UserIDs {
		if userID == actorID || availabilitySharesAny(actorProjects, targetProjects[userID]) {
			visible = append(visible, userID)
		}
	}
	if len(visible) == 0 {
		return nil
	}

	lowerDate, upperDate := availabilitySourceDateBounds(query.From, query.To)
	leaves, err := transaction.LeaveRequest.Query().
		Where(
			leaverequest.StatusEQ(leaverequest.StatusApproved),
			leaverequest.DeletedAtIsNil(),
			leaverequest.StartDateLTE(upperDate),
			leaverequest.EndDateGTE(lowerDate),
			leaverequest.HasRequesterWith(user.IDIn(visible...), user.StatusEQ(user.StatusActive)),
		).
		WithRequester().
		Order(ent.Asc(leaverequest.FieldStartDate), ent.Asc(leaverequest.FieldID)).
		Limit(availabilitySourceLeaveLimit + 1).
		All(ctx)
	if err != nil {
		return errAvailabilityUnavailable
	}
	if len(leaves) > availabilitySourceLeaveLimit {
		return ErrAvailabilityLimitExceeded
	}
	for _, leave := range leaves {
		if leave == nil || leave.DeletedAt != nil || leave.Status != leaverequest.StatusApproved || leave.Edges.Requester == nil || leave.Edges.Requester.ID == uuid.Nil {
			return ErrAvailabilitySourceInvalid
		}
		ownerID := leave.Edges.Requester.ID
		owner, found := accumulator.users[ownerID]
		if !found || owner == nil {
			return ErrAvailabilitySourceInvalid
		}
		startDate, err := availabilityDateLabel(leave.StartDate)
		if err != nil {
			return err
		}
		endDate, err := availabilityDateLabel(leave.EndDate)
		if err != nil || endDate.Before(startDate) {
			return ErrAvailabilitySourceInvalid
		}
		from, to, err := availabilityAllDayDateInterval(startDate, endDate.AddDate(0, 0, 1), owner.Timezone)
		if err != nil {
			return err
		}
		if !from.Before(query.To) || !to.After(query.From) {
			continue
		}
		if err := accumulator.append(ownerID, AvailabilityBusyInterval{From: from, To: to, Source: AvailabilitySourceLeave}); err != nil {
			return err
		}
	}
	return nil
}

func availabilityTemporalInterval(temporal CalendarEvent) (time.Time, time.Time, error) {
	if temporal.Validate() != nil {
		return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
	}
	if !temporal.AllDay {
		from, ok := canonicalAvailabilityTime(temporal.StartsAt)
		if !ok {
			return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
		}
		to, ok := canonicalAvailabilityTime(temporal.EndsAt)
		if !ok || !to.After(from) {
			return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
		}
		return from, to, nil
	}
	startDate, err := parseCalendarDate(temporal.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
	}
	endDate, err := parseCalendarDate(temporal.EndDateExclusive)
	if err != nil || !endDate.After(startDate) {
		return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
	}
	return availabilityAllDayDateInterval(startDate, endDate, temporal.Timezone)
}

func availabilityAllDayDateInterval(startDate, endDate time.Time, timezone string) (time.Time, time.Time, error) {
	if !endDate.After(startDate) || !availabilitySourceTimezoneValid(timezone) {
		return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
	}
	from, err := availabilityFirstInstantOfCivilDate(location, startDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := availabilityFirstInstantOfCivilDate(location, endDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !to.After(from) {
		return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
	}
	return from, to, nil
}

func availabilityFirstInstantOfCivilDate(location *time.Location, date time.Time) (time.Time, error) {
	if location == nil || date.IsZero() {
		return time.Time{}, ErrAvailabilitySourceInvalid
	}
	for minute := 0; minute <= availabilitySourceDaySearch; minute++ {
		candidateDate := date.Add(time.Duration(minute) * time.Minute)
		civil := calendarCreateCivilTime{Year: candidateDate.Year(), Month: candidateDate.Month(), Day: candidateDate.Day(), Hour: candidateDate.Hour(), Minute: candidateDate.Minute()}
		candidate, found := calendarCreateResolveCivilTime(location, civil)
		if !found {
			continue
		}
		utc, ok := canonicalAvailabilityTime(candidate)
		if ok {
			return utc, nil
		}
	}
	return time.Time{}, ErrAvailabilitySourceInvalid
}

func availabilityDateLabel(value time.Time) (time.Time, error) {
	value = value.UTC()
	if value.IsZero() || value.Hour() != 0 || value.Minute() != 0 || value.Second() != 0 || value.Nanosecond() != 0 {
		return time.Time{}, ErrAvailabilitySourceInvalid
	}
	parsed, err := time.Parse(time.DateOnly, value.Format(time.DateOnly))
	if err != nil || parsed.Format(time.DateOnly) != value.Format(time.DateOnly) {
		return time.Time{}, ErrAvailabilitySourceInvalid
	}
	return parsed, nil
}

func availabilitySourceDateBounds(from, to time.Time) (time.Time, time.Time) {
	lower := time.Date(from.UTC().Year(), from.UTC().Month(), from.UTC().Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -availabilitySourceDatePadding)
	upper := time.Date(to.UTC().Year(), to.UTC().Month(), to.UTC().Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, availabilitySourceDatePadding)
	return lower, upper
}

func availabilitySourceTimezoneValid(value string) bool {
	if value == "" || value == "Local" {
		return false
	}
	location, err := time.LoadLocation(value)
	return err == nil && location != time.Local && location.String() == value
}
