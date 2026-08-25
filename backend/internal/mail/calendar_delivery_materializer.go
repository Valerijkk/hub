package mail

import (
	"strings"
	"unicode"

	ical "github.com/valerii/hub/backend/internal/ical"
)

type calendarMailMaterializer struct{}

var _ CalendarMailMaterializer = calendarMailMaterializer{}

// NewCalendarMailMaterializer creates a deterministic calendar mail materializer.
func NewCalendarMailMaterializer() CalendarMailMaterializer {
	return calendarMailMaterializer{}
}

func (calendarMailMaterializer) MaterializeCalendarMail(
	input CalendarMailDeliveryMaterializationInput,
) (CalendarMailMaterialization, error) {
	if !validCalendarMailDeliveryMaterializationInput(input) {
		return CalendarMailMaterialization{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	message, _, ok := calendarMailDeliveryCanonicalITIPMessage(input.message)
	if !ok || calendarMailDeliveryMethodFromITIP(message.Method) != input.method {
		return CalendarMailMaterialization{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	if message.Location != "" &&
		(!validCalendarMailDeliveryMaterializedSingleLine(message.Location) ||
			ical.HasITIPControl(message.Location)) {
		return CalendarMailMaterialization{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	if !validCalendarMailDeliveryMaterializedDescription(message.Description) {
		return CalendarMailMaterialization{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	when, ok := calendarMailDeliveryMaterializedWhen(
		message.Temporal.Timed,
		message.Temporal.AllDay,
	)
	if !ok {
		return CalendarMailMaterialization{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	organizer, ok := calendarMailDeliveryMaterializedAddress(
		message.Organizer.DisplayName,
		message.Organizer.Email,
	)
	if !ok {
		return CalendarMailMaterialization{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	subject := ""
	status := ""
	switch input.method {
	case CalendarMailDeliveryMethodRequest:
		subject = "Invitation: " + message.Summary
		status = "You are invited."
	case CalendarMailDeliveryMethodCancel:
		subject = "Canceled: " + message.Summary
		status = "This event has been canceled."
	case CalendarMailDeliveryMethodReply:
		if len(message.Attendees) != 1 {
			return CalendarMailMaterialization{},
				newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
		}

		attendee := message.Attendees[0]
		responder, valid := calendarMailDeliveryMaterializedAddress(
			attendee.DisplayName,
			attendee.Email,
		)
		if !valid {
			return CalendarMailMaterialization{},
				newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
		}

		switch attendee.Status {
		case ical.RSVPAccepted:
			subject = "Accepted: " + message.Summary
			status = "Response: Accepted by " + responder + "."
		case ical.RSVPDeclined:
			subject = "Declined: " + message.Summary
			status = "Response: Declined by " + responder + "."
		case ical.RSVPTentative:
			subject = "Tentative: " + message.Summary
			status = "Response: Tentative by " + responder + "."
		case ical.RSVPNeedsAction:
			return CalendarMailMaterialization{},
				newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
		default:
			return CalendarMailMaterialization{},
				newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
		}
	default:
		return CalendarMailMaterialization{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	if !validCalendarMailDeliveryMaterializedSingleLine(subject) {
		return CalendarMailMaterialization{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	body, ok := calendarMailDeliveryMaterializedBody(
		status,
		when,
		message.Location,
		organizer,
		message.Description,
	)
	if !ok {
		return CalendarMailMaterialization{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	materialization, err := NewCalendarMailMaterialization(subject, MailboxContentInput{
		SourceFormat: MailboxContentFormatPlain,
		Body:         body,
	})
	if err != nil {
		return CalendarMailMaterialization{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	return materialization, nil
}

func calendarMailDeliveryMaterializedAddress(displayName, email string) (string, bool) {
	if email == "" ||
		!validCalendarMailDeliveryMaterializedSingleLine(email) ||
		!validCalendarMailDeliveryMaterializedSingleLine(displayName) ||
		ical.HasITIPControl(email) ||
		ical.HasITIPControl(displayName) ||
		strings.ContainsAny(displayName, "<>") {
		return "", false
	}
	if displayName != "" {
		return displayName + " <" + email + ">", true
	}
	return email, true
}

func validCalendarMailDeliveryMaterializedSingleLine(value string) bool {
	for _, character := range value {
		if unicode.IsControl(character) ||
			unicode.Is(unicode.Zl, character) ||
			unicode.Is(unicode.Zp, character) {
			return false
		}
	}
	return true
}

func validCalendarMailDeliveryMaterializedDescription(value string) bool {
	for _, character := range value {
		if (unicode.IsControl(character) && character != '\n') ||
			unicode.Is(unicode.Zl, character) ||
			unicode.Is(unicode.Zp, character) {
			return false
		}
	}
	return true
}

func calendarMailDeliveryMaterializedBody(
	status string,
	when string,
	location string,
	organizer string,
	description string,
) (string, bool) {
	var body calendarMailDeliveryBoundedBody
	if !body.append(status) ||
		!body.append("\n\nWhen: ") ||
		!body.append(when) {
		return "", false
	}
	if location != "" &&
		(!body.append("\nWhere: ") || !body.append(location)) {
		return "", false
	}
	if !body.append("\nOrganizer: ") || !body.append(organizer) {
		return "", false
	}
	if description != "" &&
		(!body.append("\n\n") || !body.append(description)) {
		return "", false
	}
	return body.string(), true
}

type calendarMailDeliveryBoundedBody struct {
	builder strings.Builder
	used    int
}

func (body *calendarMailDeliveryBoundedBody) append(value string) bool {
	if len(value) > MailboxContentMaxInputBytes-body.used {
		return false
	}
	body.builder.WriteString(value)
	body.used += len(value)
	return true
}

func (body *calendarMailDeliveryBoundedBody) string() string {
	return body.builder.String()
}

func calendarMailDeliveryMaterializedWhen(
	timed *ical.ITIPTimed,
	allDay *ical.ITIPAllDay,
) (string, bool) {
	switch {
	case timed != nil && allDay == nil:
		return calendarMailDeliveryMaterializedTimedWhen(timed)
	case timed == nil && allDay != nil:
		return calendarMailDeliveryMaterializedAllDayWhen(allDay)
	default:
		return "", false
	}
}

func calendarMailDeliveryMaterializedTimedWhen(
	timed *ical.ITIPTimed,
) (string, bool) {
	if timed.Timezone == "" ||
		!validCalendarMailDeliveryMaterializedSingleLine(timed.Timezone) ||
		ical.HasITIPControl(timed.Timezone) ||
		timed.StartsAt.IsZero() ||
		timed.EndsAt.IsZero() ||
		!timed.EndsAt.After(timed.StartsAt) {
		return "", false
	}

	var value strings.Builder
	value.WriteString(timed.StartsAt.UTC().Format("2006-01-02T15:04:05Z"))
	value.WriteString(" - ")
	value.WriteString(timed.EndsAt.UTC().Format("2006-01-02T15:04:05Z"))
	value.WriteString(" (UTC; event timezone: ")
	value.WriteString(timed.Timezone)
	value.WriteString(")")
	return value.String(), true
}

func calendarMailDeliveryMaterializedAllDayWhen(
	allDay *ical.ITIPAllDay,
) (string, bool) {
	if allDay.Timezone == "" ||
		!validCalendarMailDeliveryMaterializedSingleLine(allDay.Timezone) ||
		ical.HasITIPControl(allDay.Timezone) {
		return "", false
	}

	startDate, err := ical.ParseCalendarDate(allDay.StartDate)
	if err != nil {
		return "", false
	}
	endDateExclusive, err := ical.ParseCalendarDate(allDay.EndDateExclusive)
	if err != nil || !endDateExclusive.After(startDate) {
		return "", false
	}

	endDate := endDateExclusive.AddDate(0, 0, -1)
	if endDate.Before(startDate) {
		return "", false
	}

	var value strings.Builder
	value.WriteString(startDate.Format("2006-01-02"))
	if !startDate.Equal(endDate) {
		value.WriteString(" - ")
		value.WriteString(endDate.Format("2006-01-02"))
	}
	value.WriteString(" (all day, ")
	value.WriteString(allDay.Timezone)
	value.WriteString(")")
	return value.String(), true
}
