package ical

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvalidTimedEvent РѕР·РЅР°С‡Р°РµС‚ РЅРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ РёРЅС‚РµСЂРІР°Р» РёР»Рё СЃРјРµС€РµРЅРёРµ РІСЂРµРјРµРЅРЅРѕРіРѕ СЃРѕР±С‹С‚РёСЏ СЃ РґР°С‚Р°РјРё all-day.
	ErrInvalidTimedEvent = errors.New("calendar: invalid timed event")
	// ErrInvalidAllDayEvent РѕР·РЅР°С‡Р°РµС‚ РЅРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ РґРёР°РїР°Р·РѕРЅ РґР°С‚ all-day РёР»Рё РёСЃРїРѕР»СЊР·РѕРІР°РЅРёРµ РІ РЅС‘Рј instants.
	ErrInvalidAllDayEvent = errors.New("calendar: invalid all-day event")
	ErrInvalidTimezone = errors.New("calendar: invalid timezone")
	ErrInvalidRSVPStatus = errors.New("calendar: invalid RSVP status")
)

// CalendarEvent СЃРѕРґРµСЂР¶РёС‚ РІСЂРµРјРµРЅРЅСѓСЋ С„РѕСЂРјСѓ СЃРѕР±С‹С‚РёСЏ, РЅРµРѕР±С…РѕРґРёРјСѓСЋ С‡РёСЃС‚РѕР№ РєР°Р»РµРЅРґР°СЂРЅРѕР№ РїРѕР»РёС‚РёРєРµ.
type CalendarEvent struct {
	AllDay           bool
	StartsAt         time.Time
	EndsAt           time.Time
	StartDate        string
	EndDateExclusive string
	Timezone         string
}

// Validate РїСЂРѕРІРµСЂСЏРµС‚ РІР·Р°РёРјРѕРёСЃРєР»СЋС‡Р°СЋС‰РёРµ timed/all-day РїСЂРµРґСЃС‚Р°РІР»РµРЅРёСЏ Рё IANA-РІСЂРµРјРµРЅРЅСѓСЋ Р·РѕРЅСѓ СЃРѕР±С‹С‚РёСЏ.
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

// ParseCalendarDate СЂР°Р·Р±РёСЂР°РµС‚ СЃС‚СЂРѕРіРѕРµ РєР°Р»РµРЅРґР°СЂРЅРѕРµ Р·РЅР°С‡РµРЅРёРµ YYYY-MM-DD.
func ParseCalendarDate(value string) (time.Time, error) {
	return parseCalendarDate(value)
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

// ValidateTimezone РїСЂРѕРІРµСЂСЏРµС‚ IANA-РёРјСЏ РІСЂРµРјРµРЅРЅРѕР№ Р·РѕРЅС‹ РєР°Р»РµРЅРґР°СЂРЅРѕРіРѕ СЃРѕР±С‹С‚РёСЏ.
func ValidateTimezone(value string) error {
	return validateTimezone(value)
}

type RSVPStatus string

const (
	RSVPNeedsAction RSVPStatus = "needs_action"
	RSVPAccepted    RSVPStatus = "accepted"
	RSVPDeclined    RSVPStatus = "declined"
	RSVPTentative   RSVPStatus = "tentative"
)

func (status RSVPStatus) Validate() error {
	switch status {
	case RSVPNeedsAction, RSVPAccepted, RSVPDeclined, RSVPTentative:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidRSVPStatus, status)
	}
}

const (
	MaxITIPEmailBytes = itipMaxEmailBytes
	MaxITIPSequence = itipMaxSequence
	CanonicalITIPVersion = itipCanonicalVersion
)

func ValidateITIPUID(value string) error {
	return validateITIPUID(value)
}

func NormalizeITIPEmail(value string) (string, error) {
	return normalizeSafeITIPEmail(value)
}

func NormalizeITIPMessage(message ITIPMessage) (ITIPMessage, error) {
	return normalizeITIPIdentityMessage(message)
}

func HasITIPControl(value string) bool {
	return hasAnyITIPControl(value)
}

func HasLeadingOrTrailingASCIIWhitespace(value string) bool {
	return hasLeadingOrTrailingASCIIWhitespace(value)
}
