package calendar

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvalidTimedEvent означает некорректный интервал или смешение временного события с датами all-day.
	ErrInvalidTimedEvent = errors.New("calendar: invalid timed event")
	// ErrInvalidAllDayEvent означает некорректный диапазон дат all-day или использование в нём instants.
	ErrInvalidAllDayEvent = errors.New("calendar: invalid all-day event")
	// ErrInvalidTimezone означает отсутствующее или не загружаемое IANA-имя временной зоны.
	ErrInvalidTimezone = errors.New("calendar: invalid timezone")
	// ErrInvalidRSVPStatus означает состояние RSVP вне закрытого календарного набора.
	ErrInvalidRSVPStatus = errors.New("calendar: invalid RSVP status")
)

// CalendarEvent содержит временную форму события, необходимую чистой календарной политике.
type CalendarEvent struct {
	AllDay           bool
	StartsAt         time.Time
	EndsAt           time.Time
	StartDate        string
	EndDateExclusive string
	Timezone         string
}

// Validate проверяет взаимоисключающие timed/all-day представления и IANA-временную зону события.
func (event CalendarEvent) Validate() error {
	if event.AllDay {
		return event.validateAllDay()
	}

	if event.StartsAt.IsZero() || event.EndsAt.IsZero() || !event.EndsAt.After(event.StartsAt) {
		return ErrInvalidTimedEvent
	}
	if event.StartDate != "" || event.EndDateExclusive != "" {
		return ErrInvalidTimedEvent
	}

	return validateTimezone(event.Timezone)
}

func (event CalendarEvent) validateAllDay() error {
	if !event.StartsAt.IsZero() || !event.EndsAt.IsZero() {
		return ErrInvalidAllDayEvent
	}

	startDate, err := parseCalendarDate(event.StartDate)
	if err != nil {
		return ErrInvalidAllDayEvent
	}
	endDate, err := parseCalendarDate(event.EndDateExclusive)
	if err != nil || !endDate.After(startDate) {
		return ErrInvalidAllDayEvent
	}

	return validateTimezone(event.Timezone)
}

func parseCalendarDate(value string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, value)
	if err != nil || date.Format(time.DateOnly) != value {
		return time.Time{}, fmt.Errorf("%w: %q", ErrInvalidAllDayEvent, value)
	}
	return date, nil
}

func validateTimezone(value string) error {
	if value == "" || value == "Local" {
		return ErrInvalidTimezone
	}
	if _, err := time.LoadLocation(value); err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidTimezone, value)
	}
	return nil
}

// RSVPStatus задаёт закрытый набор ответов участника календарного события.
type RSVPStatus string

const (
	// RSVPNeedsAction означает, что участник ещё не ответил на приглашение.
	RSVPNeedsAction RSVPStatus = "needs_action"
	// RSVPAccepted означает, что участник подтвердил участие.
	RSVPAccepted RSVPStatus = "accepted"
	// RSVPDeclined означает, что участник отказался от участия.
	RSVPDeclined RSVPStatus = "declined"
	// RSVPTentative означает предварительное согласие участника.
	RSVPTentative RSVPStatus = "tentative"
)

// Validate проверяет принадлежность статуса RSVP к каноническому набору.
func (status RSVPStatus) Validate() error {
	switch status {
	case RSVPNeedsAction, RSVPAccepted, RSVPDeclined, RSVPTentative:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidRSVPStatus, status)
	}
}
