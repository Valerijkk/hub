package calendar

import (
	"time"

	"github.com/valerii/hub/backend/internal/ical"
)

var (
	// ErrInvalidTimedEvent означает некорректный интервал или смешение временного события с датами all-day.
	ErrInvalidTimedEvent = ical.ErrInvalidTimedEvent
	// ErrInvalidAllDayEvent означает некорректный диапазон дат all-day или использование в нём instants.
	ErrInvalidAllDayEvent = ical.ErrInvalidAllDayEvent
	ErrInvalidTimezone = ical.ErrInvalidTimezone
	ErrInvalidRSVPStatus = ical.ErrInvalidRSVPStatus
	// ErrInvalidTimezone означает отсутствующее или не загружаемое IANA-имя временной зоны.

	// ErrInvalidRSVPStatus означает состояние RSVP вне закрытого календарного набора.

)

// CalendarEvent содержит временную форму события, необходимую чистой календарной политике.
type CalendarEvent = ical.CalendarEvent

// Validate проверяет взаимоисключающие timed/all-day представления и IANA-временную зону события.
func parseCalendarDate(value string) (time.Time, error) {
	return ical.ParseCalendarDate(value)
}

func validateTimezone(value string) error {
	return ical.ValidateTimezone(value)
}

// RSVPStatus задаёт закрытый набор ответов участника календарного события.
type RSVPStatus = ical.RSVPStatus

const (
	// RSVPNeedsAction означает, что участник ещё не ответил на приглашение.
	RSVPNeedsAction = ical.RSVPNeedsAction
	// RSVPAccepted означает, что участник подтвердил участие.
	RSVPAccepted = ical.RSVPAccepted
	// RSVPDeclined означает, что участник отказался от участия.
	RSVPDeclined = ical.RSVPDeclined
	// RSVPTentative означает предварительное согласие участника.
	RSVPTentative = ical.RSVPTentative
)

// Validate проверяет принадлежность статуса RSVP к каноническому набору.
