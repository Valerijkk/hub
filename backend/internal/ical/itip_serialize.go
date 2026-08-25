package ical

import (
	"fmt"
	"net/mail"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	itipProductID = "-//Hub//Calendar//EN"
	itipLineLimit = 75
)

type preparedITIPMessage struct {
	message        ITIPMessage
	location       *time.Location
	organizerEmail string
	attendees      []preparedITIPParticipant
	exdates        []ITIPRecurrenceID
}

type preparedITIPParticipant struct {
	participant ITIPParticipant
	email       string
}

// SerializeITIP строит детерминированный RFC 5545/5546 payload без часов, I/O или глобального состояния.
func SerializeITIP(message ITIPMessage) ([]byte, error) {
	prepared, err := prepareITIPMessage(message)
	if err != nil {
		return nil, err
	}

	var builder strings.Builder
	write := func(line string) {
		builder.WriteString(foldITIPContentLine(line))
	}

	write("BEGIN:VCALENDAR")
	write("PRODID:" + itipProductID)
	write("VERSION:2.0")
	write("CALSCALE:GREGORIAN")
	write("METHOD:" + string(prepared.message.Method))
	write("BEGIN:VEVENT")
	write("UID:" + prepared.message.UID)
	write("DTSTAMP:" + prepared.message.Stamp.Format("20060102T150405Z"))
	writeITIPTemporal(&builder, prepared.message.Temporal, prepared.location)
	if prepared.message.RecurrenceID != nil {
		write(formatITIPRecurrence("RECURRENCE-ID", *prepared.message.RecurrenceID, prepared.message.Temporal, prepared.location))
	}
	write(fmt.Sprintf("SEQUENCE:%d", prepared.message.Sequence))
	write("STATUS:" + string(prepared.message.Status))
	write("SUMMARY:" + escapeITIPText(prepared.message.Summary))
	if prepared.message.Description != "" {
		write("DESCRIPTION:" + escapeITIPText(prepared.message.Description))
	}
	if prepared.message.Location != "" {
		write("LOCATION:" + escapeITIPText(prepared.message.Location))
	}
	write(formatITIPOrganizer(prepared.message.Organizer, prepared.organizerEmail))
	for _, attendee := range prepared.attendees {
		write(formatITIPAttendee(attendee.participant, attendee.email))
	}
	if prepared.message.RRULE != "" {
		write("RRULE:" + prepared.message.RRULE)
	}
	if len(prepared.exdates) > 0 {
		write(formatITIPEXDates(prepared.exdates, prepared.message.Temporal, prepared.location))
	}
	write("END:VEVENT")
	write("END:VCALENDAR")

	return []byte(builder.String()), nil
}

func prepareITIPMessage(message ITIPMessage) (preparedITIPMessage, error) {
	if message.Stamp.IsZero() || message.Stamp.Location() != time.UTC {
		return preparedITIPMessage{}, ErrInvalidITIPStamp
	}
	if !isITIPMethod(message.Method) {
		return preparedITIPMessage{}, fmt.Errorf("%w: %q", ErrInvalidITIPMethod, message.Method)
	}
	if message.Sequence < 0 {
		return preparedITIPMessage{}, fmt.Errorf("%w: %d", ErrInvalidITIPSequence, message.Sequence)
	}
	if strings.TrimSpace(message.UID) == "" || hasITIPControl(message.UID) {
		return preparedITIPMessage{}, fmt.Errorf("%w: UID", ErrInvalidITIPMessage)
	}
	if strings.TrimSpace(message.Summary) == "" || hasITIPForbiddenTextControl(message.Summary) {
		return preparedITIPMessage{}, fmt.Errorf("%w: SUMMARY", ErrInvalidITIPMessage)
	}
	if hasITIPForbiddenTextControl(message.Description) || hasITIPForbiddenTextControl(message.Location) ||
		hasITIPControl(message.RRULE) {
		return preparedITIPMessage{}, fmt.Errorf("%w: text", ErrInvalidITIPMessage)
	}
	if !isITIPStatus(message.Status) || !isITIPMethodStatus(message.Method, message.Status) {
		return preparedITIPMessage{}, fmt.Errorf("%w: %s with %s", ErrInvalidITIPStatus, message.Method, message.Status)
	}

	location, err := validateITIPTemporal(message.Temporal)
	if err != nil {
		return preparedITIPMessage{}, err
	}
	organizerEmail, err := normalizeITIPEmail(message.Organizer.Email)
	if err != nil {
		return preparedITIPMessage{}, err
	}
	if hasITIPForbiddenTextControl(message.Organizer.DisplayName) {
		return preparedITIPMessage{}, fmt.Errorf("%w: organizer display name", ErrInvalidITIPMessage)
	}

	attendees, err := prepareITIPAttendees(message.Attendees, message.Method)
	if err != nil {
		return preparedITIPMessage{}, err
	}
	if message.Method == ITIPMethodRequest && message.RecurrenceID != nil {
		return preparedITIPMessage{}, fmt.Errorf("%w: REQUEST cannot carry RECURRENCE-ID", ErrInvalidITIPRecurrence)
	}
	if message.RecurrenceID != nil {
		if err := validateITIPRecurrenceID(*message.RecurrenceID, message.Temporal); err != nil {
			return preparedITIPMessage{}, err
		}
	}

	exdates := append([]ITIPRecurrenceID(nil), message.EXDates...)
	for _, exdate := range exdates {
		if err := validateITIPRecurrenceID(exdate, message.Temporal); err != nil {
			return preparedITIPMessage{}, err
		}
	}
	sort.Slice(exdates, func(left, right int) bool {
		return canonicalITIPRecurrenceKey(exdates[left], message.Temporal, location) < canonicalITIPRecurrenceKey(exdates[right], message.Temporal, location)
	})

	return preparedITIPMessage{
		message:        message,
		location:       location,
		organizerEmail: organizerEmail,
		attendees:      attendees,
		exdates:        exdates,
	}, nil
}

func isITIPMethod(method ITIPMethod) bool {
	switch method {
	case ITIPMethodRequest, ITIPMethodCancel, ITIPMethodReply:
		return true
	default:
		return false
	}
}

func isITIPStatus(status ITIPStatus) bool {
	switch status {
	case ITIPStatusConfirmed, ITIPStatusTentative, ITIPStatusCancelled:
		return true
	default:
		return false
	}
}

func isITIPMethodStatus(method ITIPMethod, status ITIPStatus) bool {
	switch method {
	case ITIPMethodCancel:
		return status == ITIPStatusCancelled
	case ITIPMethodRequest, ITIPMethodReply:
		return status == ITIPStatusConfirmed || status == ITIPStatusTentative
	default:
		return false
	}
}

func validateITIPTemporal(temporal ITIPTemporal) (*time.Location, error) {
	if (temporal.Timed == nil) == (temporal.AllDay == nil) {
		return nil, ErrInvalidITIPTemporal
	}

	if temporal.Timed != nil {
		event := CalendarEvent{
			StartsAt: temporal.Timed.StartsAt,
			EndsAt:   temporal.Timed.EndsAt,
			Timezone: temporal.Timed.Timezone,
		}
		if err := event.Validate(); err != nil {
			return nil, err
		}
		location, err := time.LoadLocation(temporal.Timed.Timezone)
		if err != nil {
			return nil, fmt.Errorf("%w: %q", ErrInvalidTimezone, temporal.Timed.Timezone)
		}
		return location, nil
	}

	event := CalendarEvent{
		AllDay:           true,
		StartDate:        temporal.AllDay.StartDate,
		EndDateExclusive: temporal.AllDay.EndDateExclusive,
		Timezone:         temporal.AllDay.Timezone,
	}
	if err := event.Validate(); err != nil {
		return nil, err
	}
	location, err := time.LoadLocation(temporal.AllDay.Timezone)
	if err != nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidTimezone, temporal.AllDay.Timezone)
	}
	return location, nil
}

func prepareITIPAttendees(attendees []ITIPParticipant, method ITIPMethod) ([]preparedITIPParticipant, error) {
	if method == ITIPMethodReply && len(attendees) != 1 {
		return nil, fmt.Errorf("%w: REPLY requires one attendee", ErrInvalidITIPParticipant)
	}

	prepared := make([]preparedITIPParticipant, 0, len(attendees))
	for _, attendee := range attendees {
		email, err := normalizeITIPEmail(attendee.Email)
		if err != nil {
			return nil, err
		}
		if hasITIPForbiddenTextControl(attendee.DisplayName) || !isITIPParticipantRole(attendee.Role) {
			return nil, fmt.Errorf("%w: participant", ErrInvalidITIPParticipant)
		}
		if err := attendee.Status.Validate(); err != nil {
			return nil, fmt.Errorf("%w: participant status", ErrInvalidITIPParticipant)
		}
		if method == ITIPMethodReply && attendee.Status == RSVPNeedsAction {
			return nil, fmt.Errorf("%w: REPLY PARTSTAT", ErrInvalidITIPParticipant)
		}
		prepared = append(prepared, preparedITIPParticipant{participant: attendee, email: email})
	}
	sort.Slice(prepared, func(left, right int) bool {
		return prepared[left].email < prepared[right].email
	})
	for index := 1; index < len(prepared); index++ {
		if prepared[index-1].email == prepared[index].email {
			return nil, fmt.Errorf("%w: duplicate attendee", ErrInvalidITIPParticipant)
		}
	}
	return prepared, nil
}

func isITIPParticipantRole(role ITIPParticipantRole) bool {
	return role == ITIPParticipantRequired || role == ITIPParticipantOptional
}

func normalizeITIPEmail(value string) (string, error) {
	if value == "" || value != strings.TrimSpace(value) || hasITIPControl(value) {
		return "", fmt.Errorf("%w: %q", ErrInvalidITIPEmail, value)
	}
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value || strings.Count(value, "@") != 1 {
		return "", fmt.Errorf("%w: %q", ErrInvalidITIPEmail, value)
	}
	parts := strings.Split(value, "@")
	if parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("%w: %q", ErrInvalidITIPEmail, value)
	}
	return strings.ToLower(value), nil
}

func validateITIPRecurrenceID(value ITIPRecurrenceID, temporal ITIPTemporal) error {
	if temporal.Timed != nil {
		if value.Time.IsZero() || value.Date != "" {
			return ErrInvalidITIPRecurrence
		}
		return nil
	}
	if !value.Time.IsZero() || value.Date == "" {
		return ErrInvalidITIPRecurrence
	}
	if _, err := parseCalendarDate(value.Date); err != nil {
		return ErrInvalidITIPRecurrence
	}
	return nil
}

func writeITIPTemporal(builder *strings.Builder, temporal ITIPTemporal, location *time.Location) {
	if temporal.Timed != nil {
		builder.WriteString(foldITIPContentLine(formatITIPDateTime("DTSTART", temporal.Timed.StartsAt, temporal.Timed.Timezone, location)))
		builder.WriteString(foldITIPContentLine(formatITIPDateTime("DTEND", temporal.Timed.EndsAt, temporal.Timed.Timezone, location)))
		return
	}
	builder.WriteString(foldITIPContentLine("DTSTART;VALUE=DATE:" + compactITIPDate(temporal.AllDay.StartDate)))
	builder.WriteString(foldITIPContentLine("DTEND;VALUE=DATE:" + compactITIPDate(temporal.AllDay.EndDateExclusive)))
}

func formatITIPRecurrence(name string, value ITIPRecurrenceID, temporal ITIPTemporal, location *time.Location) string {
	if temporal.Timed != nil {
		return formatITIPDateTime(name, value.Time, temporal.Timed.Timezone, location)
	}
	return name + ";VALUE=DATE:" + compactITIPDate(value.Date)
}

func formatITIPEXDates(values []ITIPRecurrenceID, temporal ITIPTemporal, location *time.Location) string {
	if temporal.Timed != nil {
		name := "EXDATE"
		if temporal.Timed.Timezone != "UTC" {
			name += ";TZID=" + temporal.Timed.Timezone
		}
		formatted := make([]string, 0, len(values))
		for _, value := range values {
			if temporal.Timed.Timezone == "UTC" {
				formatted = append(formatted, value.Time.UTC().Format("20060102T150405Z"))
				continue
			}
			formatted = append(formatted, value.Time.In(location).Format("20060102T150405"))
		}
		return name + ":" + strings.Join(formatted, ",")
	}
	formatted := make([]string, 0, len(values))
	for _, value := range values {
		formatted = append(formatted, compactITIPDate(value.Date))
	}
	return "EXDATE;VALUE=DATE:" + strings.Join(formatted, ",")
}

func formatITIPDateTime(name string, value time.Time, timezone string, location *time.Location) string {
	if timezone == "UTC" {
		return name + ":" + value.UTC().Format("20060102T150405Z")
	}
	return name + ";TZID=" + timezone + ":" + value.In(location).Format("20060102T150405")
}

func canonicalITIPRecurrenceKey(value ITIPRecurrenceID, temporal ITIPTemporal, location *time.Location) string {
	if temporal.Timed != nil {
		return value.Time.In(location).Format("20060102T150405.999999999Z07:00")
	}
	return value.Date
}

func compactITIPDate(value string) string {
	return strings.ReplaceAll(value, "-", "")
}

func formatITIPOrganizer(organizer ITIPOrganizer, email string) string {
	return "ORGANIZER" + formatITIPCommonName(organizer.DisplayName) + ":mailto:" + email
}

func formatITIPAttendee(attendee ITIPParticipant, email string) string {
	return "ATTENDEE" +
		formatITIPCommonName(attendee.DisplayName) +
		";ROLE=" + formatITIPRole(attendee.Role) +
		";PARTSTAT=" + formatITIPRSVPStatus(attendee.Status) +
		";RSVP=" + formatITIPBoolean(attendee.RSVP) +
		":mailto:" + email
}

func formatITIPCommonName(value string) string {
	if value == "" {
		return ""
	}
	return `;CN="` + escapeITIPParameter(value) + `"`
}

func formatITIPRole(role ITIPParticipantRole) string {
	if role == ITIPParticipantOptional {
		return "OPT-PARTICIPANT"
	}
	return "REQ-PARTICIPANT"
}

func formatITIPRSVPStatus(status RSVPStatus) string {
	switch status {
	case RSVPNeedsAction:
		return "NEEDS-ACTION"
	case RSVPAccepted:
		return "ACCEPTED"
	case RSVPDeclined:
		return "DECLINED"
	default:
		return "TENTATIVE"
	}
}

func formatITIPBoolean(value bool) string {
	if value {
		return "TRUE"
	}
	return "FALSE"
}

func escapeITIPText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.NewReplacer(
		"\\", "\\\\",
		";", "\\;",
		",", "\\,",
		"\n", "\\n",
	).Replace(value)
}

func escapeITIPParameter(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.NewReplacer(
		"^", "^^",
		"\"", "^'",
		"\n", "^n",
	).Replace(value)
}

func foldITIPContentLine(value string) string {
	var builder strings.Builder
	remaining := value
	first := true
	for {
		limit := itipLineLimit
		if !first {
			builder.WriteByte(' ')
			limit--
		}
		if len(remaining) <= limit {
			builder.WriteString(remaining)
			break
		}
		cut := itipRuneBoundary(remaining, limit)
		builder.WriteString(remaining[:cut])
		builder.WriteString("\r\n")
		remaining = remaining[cut:]
		first = false
	}
	builder.WriteString("\r\n")
	return builder.String()
}

func itipRuneBoundary(value string, limit int) int {
	cut := limit
	for cut > 0 && !utf8.RuneStart(value[cut]) {
		cut--
	}
	return cut
}

func hasITIPControl(value string) bool {
	for _, character := range value {
		if unicode.IsControl(character) {
			return true
		}
	}
	return false
}

func hasITIPForbiddenTextControl(value string) bool {
	for _, character := range value {
		if character != '\r' && character != '\n' && unicode.IsControl(character) {
			return true
		}
	}
	return false
}
