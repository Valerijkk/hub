package ical

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	itipMaxUIDBytes         = 255
	itipMaxSummaryBytes     = 500
	itipMaxDescriptionBytes = 200000
	itipMaxLocationBytes    = 500
	itipMaxEmailBytes       = 320
	itipMaxDisplayNameBytes = 160
	itipMaxTimezoneBytes    = 64
	itipMaxRRULEBytes       = 1000
	itipMaxSequence         = 2147483647
	itipCanonicalVersion    = 1
)

type parsedITIPProperty struct {
	name   string
	params map[string]string
	value  string
}

type collectedITIPProperties struct {
	calendar map[string][]parsedITIPProperty
	event    map[string][]parsedITIPProperty
}

type parsedITIPValueForm uint8

const (
	parsedITIPValueUTC parsedITIPValueForm = iota + 1
	parsedITIPValueZoned
	parsedITIPValueDate
)

type parsedITIPTemporalValue struct {
	form     parsedITIPValueForm
	timezone string
	instant  time.Time
	date     string
}

// ParseITIP разбирает один ограниченный VCALENDAR/VEVENT payload без I/O и мутаций.
func ParseITIP(payload []byte, limits ITIPParseLimits) (ITIPMessage, error) {
	if !limits.valid() {
		return ITIPMessage{}, ErrInvalidITIPLimits
	}
	if len(payload) > limits.MaxBytes {
		return ITIPMessage{}, ErrITIPLimitExceeded
	}

	lines, err := unfoldITIPLines(payload, limits)
	if err != nil {
		return ITIPMessage{}, err
	}
	properties, err := collectITIPProperties(lines, limits)
	if err != nil {
		return ITIPMessage{}, err
	}
	return buildParsedITIPMessage(properties, limits)
}

func unfoldITIPLines(payload []byte, limits ITIPParseLimits) ([]string, error) {
	if len(payload) == 0 {
		return nil, ErrInvalidITIPSyntax
	}

	lines := make([]string, 0, min(limits.MaxProperties, 64))
	var current []byte
	logicalCount := 0
	start := 0
	for start < len(payload) {
		newlineOffset := -1
		for index := start; index < len(payload); index++ {
			if payload[index] == '\n' {
				newlineOffset = index
				break
			}
		}

		end := len(payload)
		next := len(payload)
		if newlineOffset >= 0 {
			end = newlineOffset
			next = newlineOffset + 1
			if end > start && payload[end-1] == '\r' {
				end--
			}
		}
		for index := start; index < end; index++ {
			if payload[index] == '\r' {
				return nil, ErrInvalidITIPSyntax
			}
		}

		physical := payload[start:end]
		if len(physical) == 0 {
			return nil, ErrInvalidITIPSyntax
		}
		continuation := len(physical) > 0 && (physical[0] == ' ' || physical[0] == '\t')
		if continuation {
			if current == nil {
				return nil, ErrInvalidITIPSyntax
			}
			continuationBytes := len(physical) - 1
			if continuationBytes > limits.MaxUnfoldedLineBytes-len(current) {
				return nil, ErrITIPLimitExceeded
			}
			current = append(current, physical[1:]...)
		} else {
			if current != nil {
				if !utf8.Valid(current) {
					return nil, ErrInvalidITIPSyntax
				}
				lines = append(lines, string(current))
			}
			logicalCount++
			if logicalCount > limits.MaxProperties {
				return nil, ErrITIPLimitExceeded
			}
			if len(physical) > limits.MaxUnfoldedLineBytes {
				return nil, ErrITIPLimitExceeded
			}
			current = append([]byte(nil), physical...)
		}
		start = next
	}
	if current != nil {
		if !utf8.Valid(current) {
			return nil, ErrInvalidITIPSyntax
		}
		lines = append(lines, string(current))
	}
	if len(lines) == 0 {
		return nil, ErrInvalidITIPSyntax
	}
	return lines, nil
}

func collectITIPProperties(lines []string, limits ITIPParseLimits) (collectedITIPProperties, error) {
	result := collectedITIPProperties{
		calendar: make(map[string][]parsedITIPProperty),
		event:    make(map[string][]parsedITIPProperty),
	}
	state := uint8(0)
	eventCount := 0
	exdateCount := 0
	attendeeCount := 0

	for index, line := range lines {
		property, err := parseITIPContentLine(line)
		if err != nil {
			return collectedITIPProperties{}, err
		}

		if property.name == "BEGIN" {
			if len(property.params) != 0 {
				return collectedITIPProperties{}, ErrUnsupportedITIPContent
			}
			component := strings.ToUpper(property.value)
			switch state {
			case 0:
				if index != 0 || component != "VCALENDAR" {
					return collectedITIPProperties{}, ErrInvalidITIPSyntax
				}
				state = 1
			case 1:
				if component != "VEVENT" || eventCount != 0 {
					return collectedITIPProperties{}, ErrUnsupportedITIPContent
				}
				eventCount++
				state = 2
			default:
				return collectedITIPProperties{}, ErrUnsupportedITIPContent
			}
			continue
		}

		if property.name == "END" {
			if len(property.params) != 0 {
				return collectedITIPProperties{}, ErrUnsupportedITIPContent
			}
			component := strings.ToUpper(property.value)
			switch state {
			case 2:
				if component != "VEVENT" {
					return collectedITIPProperties{}, ErrInvalidITIPSyntax
				}
				state = 1
			case 1:
				if component != "VCALENDAR" || index != len(lines)-1 {
					return collectedITIPProperties{}, ErrInvalidITIPSyntax
				}
				state = 3
			default:
				return collectedITIPProperties{}, ErrInvalidITIPSyntax
			}
			continue
		}

		switch state {
		case 1:
			if !isSupportedITIPCalendarProperty(property.name) {
				return collectedITIPProperties{}, ErrUnsupportedITIPContent
			}
			result.calendar[property.name] = append(result.calendar[property.name], property)
		case 2:
			if !isSupportedITIPEventProperty(property.name) {
				return collectedITIPProperties{}, ErrUnsupportedITIPContent
			}
			if property.name == "ATTENDEE" {
				attendeeCount++
				if attendeeCount > limits.MaxAttendees {
					return collectedITIPProperties{}, ErrITIPLimitExceeded
				}
			}
			if property.name == "EXDATE" {
				exdateCount += 1 + strings.Count(property.value, ",")
				if exdateCount > limits.MaxExDates {
					return collectedITIPProperties{}, ErrITIPLimitExceeded
				}
			}
			result.event[property.name] = append(result.event[property.name], property)
		default:
			return collectedITIPProperties{}, ErrInvalidITIPSyntax
		}
	}

	if state != 3 || eventCount != 1 {
		return collectedITIPProperties{}, ErrInvalidITIPSyntax
	}
	return result, nil
}

func parseITIPContentLine(line string) (parsedITIPProperty, error) {
	if line == "" {
		return parsedITIPProperty{}, ErrInvalidITIPSyntax
	}
	nameEnd := strings.IndexAny(line, ";:")
	if nameEnd <= 0 || !isITIPToken(line[:nameEnd]) {
		return parsedITIPProperty{}, ErrInvalidITIPSyntax
	}
	propertyName := strings.ToUpper(line[:nameEnd])
	maximumParameters, supported := itipParameterPolicy(propertyName)
	if !supported {
		return parsedITIPProperty{}, ErrUnsupportedITIPContent
	}

	colon := -1
	quoted := false
	parameterCount := 0
	for index := nameEnd; index < len(line); index++ {
		switch line[index] {
		case '"':
			quoted = !quoted
		case ';':
			if !quoted {
				parameterCount++
				if parameterCount > maximumParameters || parameterCount > itipHardMaxParameterSlots {
					return parsedITIPProperty{}, ErrUnsupportedITIPContent
				}
			}
		case ':':
			if !quoted {
				colon = index
				index = len(line)
			}
		}
	}
	if colon <= 0 || quoted {
		return parsedITIPProperty{}, ErrInvalidITIPSyntax
	}

	headerParts, err := splitITIPHeader(line[:colon], parameterCount)
	if err != nil || len(headerParts) == 0 || !isITIPToken(headerParts[0]) {
		return parsedITIPProperty{}, ErrInvalidITIPSyntax
	}
	property := parsedITIPProperty{
		name:   propertyName,
		params: make(map[string]string, len(headerParts)-1),
		value:  line[colon+1:],
	}
	for _, rawParameter := range headerParts[1:] {
		equals := strings.IndexByte(rawParameter, '=')
		if equals <= 0 || equals == len(rawParameter)-1 {
			return parsedITIPProperty{}, ErrInvalidITIPSyntax
		}
		key := rawParameter[:equals]
		if !isITIPToken(key) {
			return parsedITIPProperty{}, ErrInvalidITIPSyntax
		}
		key = strings.ToUpper(key)
		if !isAllowedITIPParameter(property.name, key) {
			return parsedITIPProperty{}, ErrUnsupportedITIPContent
		}
		if _, exists := property.params[key]; exists {
			return parsedITIPProperty{}, ErrInvalidITIPSyntax
		}
		rawValue := rawParameter[equals+1:]
		if rawValue[0] == '"' {
			if len(rawValue) < 2 || rawValue[len(rawValue)-1] != '"' || strings.Contains(rawValue[1:len(rawValue)-1], `"`) {
				return parsedITIPProperty{}, ErrInvalidITIPSyntax
			}
			rawValue = rawValue[1 : len(rawValue)-1]
		} else if strings.Contains(rawValue, `"`) {
			return parsedITIPProperty{}, ErrInvalidITIPSyntax
		}
		value, decodeErr := decodeITIPParameter(rawValue)
		if decodeErr != nil {
			return parsedITIPProperty{}, decodeErr
		}
		property.params[key] = value
	}
	return property, nil
}

func splitITIPHeader(header string, parameterCount int) ([]string, error) {
	parts := make([]string, 0, 1+parameterCount)
	quoted := false
	start := 0
	for index := 0; index < len(header); index++ {
		switch header[index] {
		case '"':
			quoted = !quoted
		case ';':
			if !quoted {
				if index == start {
					return nil, ErrInvalidITIPSyntax
				}
				parts = append(parts, header[start:index])
				start = index + 1
			}
		}
	}
	if quoted || start == len(header) {
		return nil, ErrInvalidITIPSyntax
	}
	parts = append(parts, header[start:])
	return parts, nil
}

func itipParameterPolicy(propertyName string) (int, bool) {
	switch propertyName {
	case "BEGIN", "END", "VERSION", "PRODID", "CALSCALE", "METHOD", "UID", "DTSTAMP",
		"SEQUENCE", "STATUS", "SUMMARY", "DESCRIPTION", "LOCATION", "RRULE":
		return 0, true
	case "DTSTART", "DTEND", "RECURRENCE-ID", "EXDATE", "ORGANIZER":
		return 1, true
	case "ATTENDEE":
		return itipHardMaxParameterSlots, true
	default:
		return 0, false
	}
}

func isAllowedITIPParameter(propertyName, parameterName string) bool {
	switch propertyName {
	case "DTSTART", "DTEND", "RECURRENCE-ID", "EXDATE":
		return parameterName == "VALUE" || parameterName == "TZID"
	case "ORGANIZER":
		return parameterName == "CN"
	case "ATTENDEE":
		switch parameterName {
		case "CN", "ROLE", "PARTSTAT", "RSVP":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func isITIPToken(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '-' {
			continue
		}
		return false
	}
	return true
}

func isSupportedITIPCalendarProperty(name string) bool {
	switch name {
	case "VERSION", "PRODID", "CALSCALE", "METHOD":
		return true
	default:
		return false
	}
}

func isSupportedITIPEventProperty(name string) bool {
	switch name {
	case "UID", "DTSTAMP", "DTSTART", "DTEND", "RECURRENCE-ID", "SEQUENCE", "STATUS", "SUMMARY",
		"DESCRIPTION", "LOCATION", "ORGANIZER", "ATTENDEE", "RRULE", "EXDATE":
		return true
	default:
		return false
	}
}

func buildParsedITIPMessage(properties collectedITIPProperties, limits ITIPParseLimits) (ITIPMessage, error) {
	version, err := requireSingleITIPProperty(properties.calendar, "VERSION", true)
	if err != nil {
		return ITIPMessage{}, err
	}
	if err := requireNoITIPParameters(version); err != nil {
		return ITIPMessage{}, err
	}
	if version.value != "2.0" {
		return ITIPMessage{}, ErrInvalidITIPMessage
	}

	product, err := requireSingleITIPProperty(properties.calendar, "PRODID", true)
	if err != nil {
		return ITIPMessage{}, err
	}
	if err := requireNoITIPParameters(product); err != nil {
		return ITIPMessage{}, err
	}
	productID, err := decodeITIPText(product.value)
	if err != nil {
		return ITIPMessage{}, err
	}
	if productID == "" || len(productID) > itipMaxUIDBytes || hasAnyITIPControl(productID) {
		return ITIPMessage{}, ErrInvalidITIPMessage
	}

	calendarScale, err := requireSingleITIPProperty(properties.calendar, "CALSCALE", false)
	if err != nil {
		return ITIPMessage{}, err
	}
	if calendarScale.name != "" {
		if err := requireNoITIPParameters(calendarScale); err != nil {
			return ITIPMessage{}, err
		}
		if !strings.EqualFold(calendarScale.value, "GREGORIAN") {
			return ITIPMessage{}, ErrInvalidITIPMessage
		}
	}

	methodProperty, err := requireSingleITIPProperty(properties.calendar, "METHOD", true)
	if err != nil {
		return ITIPMessage{}, err
	}
	if err := requireNoITIPParameters(methodProperty); err != nil {
		return ITIPMessage{}, err
	}
	method, err := parseITIPMethod(methodProperty.value)
	if err != nil {
		return ITIPMessage{}, err
	}

	uidProperty, err := requireSingleITIPProperty(properties.event, "UID", true)
	if err != nil {
		return ITIPMessage{}, err
	}
	if err := requireNoITIPParameters(uidProperty); err != nil {
		return ITIPMessage{}, err
	}
	uid, err := decodeITIPText(uidProperty.value)
	if err != nil {
		return ITIPMessage{}, err
	}
	if err := validateITIPUID(uid); err != nil {
		return ITIPMessage{}, err
	}

	stampProperty, err := requireSingleITIPProperty(properties.event, "DTSTAMP", true)
	if err != nil {
		return ITIPMessage{}, err
	}
	if err := requireNoITIPParameters(stampProperty); err != nil {
		return ITIPMessage{}, err
	}
	stamp, err := parseITIPUTCInstant(stampProperty.value, ErrInvalidITIPStamp)
	if err != nil {
		return ITIPMessage{}, err
	}

	startProperty, err := requireSingleITIPProperty(properties.event, "DTSTART", true)
	if err != nil {
		return ITIPMessage{}, err
	}
	endProperty, err := requireSingleITIPProperty(properties.event, "DTEND", true)
	if err != nil {
		return ITIPMessage{}, err
	}
	temporal, startValue, err := parseITIPTemporal(startProperty, endProperty)
	if err != nil {
		return ITIPMessage{}, err
	}

	sequenceProperty, err := requireSingleITIPProperty(properties.event, "SEQUENCE", true)
	if err != nil {
		return ITIPMessage{}, err
	}
	if err := requireNoITIPParameters(sequenceProperty); err != nil {
		return ITIPMessage{}, err
	}
	sequence, err := parseITIPSequence(sequenceProperty.value)
	if err != nil {
		return ITIPMessage{}, err
	}

	statusProperty, err := requireSingleITIPProperty(properties.event, "STATUS", true)
	if err != nil {
		return ITIPMessage{}, err
	}
	if err := requireNoITIPParameters(statusProperty); err != nil {
		return ITIPMessage{}, err
	}
	status, err := parseITIPStatus(statusProperty.value)
	if err != nil {
		return ITIPMessage{}, err
	}
	if !isITIPMethodStatus(method, status) {
		return ITIPMessage{}, ErrInvalidITIPStatus
	}

	summary, err := parseSingleITIPText(properties.event, "SUMMARY", true, itipMaxSummaryBytes)
	if err != nil {
		return ITIPMessage{}, err
	}
	description, err := parseSingleITIPText(properties.event, "DESCRIPTION", false, itipMaxDescriptionBytes)
	if err != nil {
		return ITIPMessage{}, err
	}
	location, err := parseSingleITIPText(properties.event, "LOCATION", false, itipMaxLocationBytes)
	if err != nil {
		return ITIPMessage{}, err
	}

	organizerProperty, err := requireSingleITIPProperty(properties.event, "ORGANIZER", true)
	if err != nil {
		return ITIPMessage{}, err
	}
	organizer, err := parseITIPOrganizer(organizerProperty)
	if err != nil {
		return ITIPMessage{}, err
	}

	attendees := make([]ITIPParticipant, 0, len(properties.event["ATTENDEE"]))
	for _, attendeeProperty := range properties.event["ATTENDEE"] {
		attendee, attendeeErr := parseITIPAttendee(attendeeProperty)
		if attendeeErr != nil {
			return ITIPMessage{}, attendeeErr
		}
		attendees = append(attendees, attendee)
	}
	if err := normalizeAndValidateITIPAttendees(attendees, method); err != nil {
		return ITIPMessage{}, err
	}
	sort.Slice(attendees, func(left, right int) bool {
		return attendees[left].Email < attendees[right].Email
	})

	rruleProperty, err := requireSingleITIPProperty(properties.event, "RRULE", false)
	if err != nil {
		return ITIPMessage{}, err
	}
	rrule := ""
	if rruleProperty.name != "" {
		if err := requireNoITIPParameters(rruleProperty); err != nil {
			return ITIPMessage{}, err
		}
		rrule, err = normalizeITIPRRULE(rruleProperty.value, temporal)
		if err != nil {
			return ITIPMessage{}, err
		}
	}

	var recurrenceID *ITIPRecurrenceID
	recurrenceProperty, err := requireSingleITIPProperty(properties.event, "RECURRENCE-ID", false)
	if err != nil {
		return ITIPMessage{}, err
	}
	if recurrenceProperty.name != "" {
		value, recurrenceErr := parseITIPRecurrenceValue(recurrenceProperty, temporal, startValue)
		if recurrenceErr != nil {
			return ITIPMessage{}, recurrenceErr
		}
		recurrenceID = &value
	}

	exdates := make([]ITIPRecurrenceID, 0, min(limits.MaxExDates, len(properties.event["EXDATE"])))
	for _, exdateProperty := range properties.event["EXDATE"] {
		values := strings.Split(exdateProperty.value, ",")
		for _, rawValue := range values {
			if rawValue == "" {
				return ITIPMessage{}, ErrInvalidITIPRecurrence
			}
			individual := exdateProperty
			individual.value = rawValue
			value, exdateErr := parseITIPRecurrenceValue(individual, temporal, startValue)
			if exdateErr != nil {
				return ITIPMessage{}, exdateErr
			}
			exdates = append(exdates, value)
		}
	}
	if err := sortAndRejectDuplicateITIPRecurrences(exdates, temporal); err != nil {
		return ITIPMessage{}, err
	}

	return ITIPMessage{
		Method:       method,
		UID:          uid,
		Stamp:        stamp,
		Sequence:     sequence,
		Status:       status,
		Summary:      summary,
		Description:  description,
		Location:     location,
		Temporal:     temporal,
		Organizer:    organizer,
		Attendees:    attendees,
		RRULE:        rrule,
		EXDates:      exdates,
		RecurrenceID: recurrenceID,
	}, nil
}

func requireSingleITIPProperty(properties map[string][]parsedITIPProperty, name string, required bool) (parsedITIPProperty, error) {
	values := properties[name]
	if len(values) > 1 {
		return parsedITIPProperty{}, ErrInvalidITIPSyntax
	}
	if len(values) == 0 {
		if required {
			return parsedITIPProperty{}, ErrInvalidITIPMessage
		}
		return parsedITIPProperty{}, nil
	}
	return values[0], nil
}

func requireNoITIPParameters(property parsedITIPProperty) error {
	if len(property.params) != 0 {
		return ErrUnsupportedITIPContent
	}
	return nil
}

func parseSingleITIPText(properties map[string][]parsedITIPProperty, name string, required bool, maximum int) (string, error) {
	property, err := requireSingleITIPProperty(properties, name, required)
	if err != nil {
		return "", err
	}
	if property.name == "" {
		return "", nil
	}
	if err := requireNoITIPParameters(property); err != nil {
		return "", err
	}
	value, err := decodeITIPText(property.value)
	if err != nil {
		return "", err
	}
	if (required && value == "") || len(value) > maximum || hasForbiddenITIPTextControl(value) {
		return "", ErrInvalidITIPMessage
	}
	return value, nil
}

func parseITIPMethod(value string) (ITIPMethod, error) {
	switch strings.ToUpper(value) {
	case string(ITIPMethodRequest):
		return ITIPMethodRequest, nil
	case string(ITIPMethodCancel):
		return ITIPMethodCancel, nil
	case string(ITIPMethodReply):
		return ITIPMethodReply, nil
	default:
		return "", ErrInvalidITIPMethod
	}
}

func parseITIPStatus(value string) (ITIPStatus, error) {
	switch strings.ToUpper(value) {
	case string(ITIPStatusConfirmed):
		return ITIPStatusConfirmed, nil
	case string(ITIPStatusTentative):
		return ITIPStatusTentative, nil
	case string(ITIPStatusCancelled):
		return ITIPStatusCancelled, nil
	default:
		return "", ErrInvalidITIPStatus
	}
}

func parseITIPSequence(value string) (int, error) {
	if value == "" || hasAnyITIPControl(value) {
		return 0, ErrInvalidITIPSequence
	}
	for index := 0; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return 0, ErrInvalidITIPSequence
		}
	}
	sequence, err := strconv.ParseInt(value, 10, 32)
	if err != nil || sequence < 0 || sequence > itipMaxSequence {
		return 0, ErrInvalidITIPSequence
	}
	return int(sequence), nil
}

func validateITIPUID(value string) error {
	if value == "" || !utf8.ValidString(value) || len(value) > itipMaxUIDBytes || hasAnyITIPControl(value) ||
		hasLeadingOrTrailingASCIIWhitespace(value) {
		return ErrInvalidITIPMessage
	}
	return nil
}

func hasLeadingOrTrailingASCIIWhitespace(value string) bool {
	if value == "" {
		return false
	}
	return isASCIIWhitespace(value[0]) || isASCIIWhitespace(value[len(value)-1])
}

func isASCIIWhitespace(value byte) bool {
	switch value {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}

func parseITIPTemporal(startProperty, endProperty parsedITIPProperty) (ITIPTemporal, parsedITIPTemporalValue, error) {
	start, err := parseITIPTemporalValue(startProperty, ErrInvalidITIPTemporal)
	if err != nil {
		return ITIPTemporal{}, parsedITIPTemporalValue{}, err
	}
	end, err := parseITIPTemporalValue(endProperty, ErrInvalidITIPTemporal)
	if err != nil {
		return ITIPTemporal{}, parsedITIPTemporalValue{}, err
	}
	if start.form != end.form || (start.form == parsedITIPValueZoned && start.timezone != end.timezone) {
		return ITIPTemporal{}, parsedITIPTemporalValue{}, ErrInvalidITIPTemporal
	}

	if start.form == parsedITIPValueDate {
		startDate, parseErr := parseCanonicalITIPDate(start.date)
		if parseErr != nil {
			return ITIPTemporal{}, parsedITIPTemporalValue{}, ErrInvalidITIPTemporal
		}
		endDate, parseErr := parseCanonicalITIPDate(end.date)
		if parseErr != nil || !endDate.After(startDate) {
			return ITIPTemporal{}, parsedITIPTemporalValue{}, ErrInvalidITIPTemporal
		}
		return ITIPTemporal{AllDay: &ITIPAllDay{
			StartDate:        start.date,
			EndDateExclusive: end.date,
			Timezone:         "UTC",
		}}, start, nil
	}
	if !end.instant.After(start.instant) {
		return ITIPTemporal{}, parsedITIPTemporalValue{}, ErrInvalidITIPTemporal
	}
	timezone := start.timezone
	if start.form == parsedITIPValueUTC {
		timezone = "UTC"
	}
	return ITIPTemporal{Timed: &ITIPTimed{
		StartsAt: start.instant.UTC(),
		EndsAt:   end.instant.UTC(),
		Timezone: timezone,
	}}, start, nil
}

func parseITIPTemporalValue(property parsedITIPProperty, semanticError error) (parsedITIPTemporalValue, error) {
	for parameter := range property.params {
		if parameter != "VALUE" && parameter != "TZID" {
			return parsedITIPTemporalValue{}, ErrUnsupportedITIPContent
		}
	}
	valueType, hasValueType := property.params["VALUE"]
	timezone, hasTimezone := property.params["TZID"]
	if hasValueType {
		if !strings.EqualFold(valueType, "DATE") || hasTimezone {
			return parsedITIPTemporalValue{}, semanticError
		}
		date, err := parseCompactITIPDate(property.value)
		if err != nil {
			return parsedITIPTemporalValue{}, semanticError
		}
		return parsedITIPTemporalValue{form: parsedITIPValueDate, date: date}, nil
	}
	if hasTimezone {
		location, err := loadCanonicalITIPLocation(timezone)
		if err != nil {
			return parsedITIPTemporalValue{}, err
		}
		instant, err := parseITIPLocalInstant(property.value, location)
		if err != nil {
			return parsedITIPTemporalValue{}, semanticError
		}
		return parsedITIPTemporalValue{
			form:     parsedITIPValueZoned,
			timezone: timezone,
			instant:  instant.UTC(),
		}, nil
	}
	instant, err := parseITIPUTCInstant(property.value, semanticError)
	if err != nil {
		return parsedITIPTemporalValue{}, err
	}
	return parsedITIPTemporalValue{form: parsedITIPValueUTC, timezone: "UTC", instant: instant}, nil
}

func loadCanonicalITIPLocation(timezone string) (*time.Location, error) {
	if timezone == "" || timezone == "Local" || !utf8.ValidString(timezone) ||
		len(timezone) > itipMaxTimezoneBytes || hasAnyITIPControl(timezone) {
		return nil, ErrInvalidTimezone
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, ErrInvalidTimezone
	}
	return location, nil
}

func parseITIPUTCInstant(value string, semanticError error) (time.Time, error) {
	if len(value) != len("20060102T150405Z") || value[len(value)-1] != 'Z' || hasAnyITIPControl(value) {
		return time.Time{}, semanticError
	}
	instant, err := time.Parse("20060102T150405Z", value)
	if err != nil || instant.Format("20060102T150405Z") != value {
		return time.Time{}, semanticError
	}
	return instant.UTC(), nil
}

func parseITIPLocalInstant(value string, location *time.Location) (time.Time, error) {
	if len(value) != len("20060102T150405") || hasAnyITIPControl(value) {
		return time.Time{}, ErrInvalidITIPTemporal
	}
	instant, err := time.ParseInLocation("20060102T150405", value, location)
	if err != nil || instant.In(location).Format("20060102T150405") != value {
		return time.Time{}, ErrInvalidITIPTemporal
	}
	return instant, nil
}

func parseCompactITIPDate(value string) (string, error) {
	if len(value) != 8 || hasAnyITIPControl(value) {
		return "", ErrInvalidITIPTemporal
	}
	parsed, err := time.Parse("20060102", value)
	if err != nil || parsed.Format("20060102") != value {
		return "", ErrInvalidITIPTemporal
	}
	return parsed.Format("2006-01-02"), nil
}

func parseCanonicalITIPDate(value string) (time.Time, error) {
	if len(value) != 10 {
		return time.Time{}, ErrInvalidITIPTemporal
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Format("2006-01-02") != value {
		return time.Time{}, ErrInvalidITIPTemporal
	}
	return parsed, nil
}

func parseITIPRecurrenceValue(property parsedITIPProperty, temporal ITIPTemporal, eventStart parsedITIPTemporalValue) (ITIPRecurrenceID, error) {
	parsed, err := parseITIPTemporalValue(property, ErrInvalidITIPRecurrence)
	if err != nil {
		if err == ErrInvalidITIPTemporal || err == ErrInvalidITIPStamp {
			return ITIPRecurrenceID{}, ErrInvalidITIPRecurrence
		}
		return ITIPRecurrenceID{}, err
	}

	if temporal.AllDay != nil {
		if parsed.form != parsedITIPValueDate {
			return ITIPRecurrenceID{}, ErrInvalidITIPRecurrence
		}
		return ITIPRecurrenceID{Date: parsed.date}, nil
	}
	if parsed.form == parsedITIPValueDate {
		return ITIPRecurrenceID{}, ErrInvalidITIPRecurrence
	}
	if parsed.form == parsedITIPValueZoned {
		if eventStart.form != parsedITIPValueZoned || parsed.timezone != eventStart.timezone {
			return ITIPRecurrenceID{}, ErrInvalidITIPRecurrence
		}
	}
	return ITIPRecurrenceID{Time: parsed.instant.UTC()}, nil
}

func parseITIPOrganizer(property parsedITIPProperty) (ITIPOrganizer, error) {
	for parameter := range property.params {
		if parameter != "CN" {
			return ITIPOrganizer{}, ErrUnsupportedITIPContent
		}
	}
	displayName := property.params["CN"]
	if len(displayName) > itipMaxDisplayNameBytes || hasForbiddenITIPTextControl(displayName) {
		return ITIPOrganizer{}, ErrInvalidITIPMessage
	}
	email, err := parseITIPMailto(property.value)
	if err != nil {
		return ITIPOrganizer{}, err
	}
	return ITIPOrganizer{Email: email, DisplayName: displayName}, nil
}

func parseITIPAttendee(property parsedITIPProperty) (ITIPParticipant, error) {
	for parameter := range property.params {
		switch parameter {
		case "CN", "ROLE", "PARTSTAT", "RSVP":
		default:
			return ITIPParticipant{}, ErrUnsupportedITIPContent
		}
	}
	email, err := parseITIPMailto(property.value)
	if err != nil {
		return ITIPParticipant{}, err
	}
	displayName := property.params["CN"]
	if len(displayName) > itipMaxDisplayNameBytes || hasForbiddenITIPTextControl(displayName) {
		return ITIPParticipant{}, ErrInvalidITIPParticipant
	}

	role := ITIPParticipantRequired
	if value, exists := property.params["ROLE"]; exists {
		switch strings.ToUpper(value) {
		case "REQ-PARTICIPANT":
			role = ITIPParticipantRequired
		case "OPT-PARTICIPANT":
			role = ITIPParticipantOptional
		default:
			return ITIPParticipant{}, ErrInvalidITIPParticipant
		}
	}
	status := RSVPNeedsAction
	if value, exists := property.params["PARTSTAT"]; exists {
		switch strings.ToUpper(value) {
		case "NEEDS-ACTION":
			status = RSVPNeedsAction
		case "ACCEPTED":
			status = RSVPAccepted
		case "DECLINED":
			status = RSVPDeclined
		case "TENTATIVE":
			status = RSVPTentative
		default:
			return ITIPParticipant{}, ErrInvalidITIPParticipant
		}
	}
	rsvp := false
	if value, exists := property.params["RSVP"]; exists {
		switch strings.ToUpper(value) {
		case "TRUE":
			rsvp = true
		case "FALSE":
			rsvp = false
		default:
			return ITIPParticipant{}, ErrInvalidITIPParticipant
		}
	}
	return ITIPParticipant{
		Email:       email,
		DisplayName: displayName,
		Role:        role,
		Status:      status,
		RSVP:        rsvp,
	}, nil
}

func parseITIPMailto(value string) (string, error) {
	if len(value) < len("mailto:") || !strings.EqualFold(value[:len("mailto:")], "mailto:") || hasAnyITIPControl(value) {
		return "", ErrInvalidITIPEmail
	}
	rawEmail := value[len("mailto:"):]
	if len(rawEmail) > itipMaxEmailBytes {
		return "", ErrInvalidITIPEmail
	}
	email, err := normalizeSafeITIPEmail(rawEmail)
	if err != nil || len(email) > itipMaxEmailBytes {
		return "", ErrInvalidITIPEmail
	}
	return email, nil
}

func normalizeSafeITIPEmail(value string) (string, error) {
	normalized, err := normalizeITIPEmail(value)
	if err != nil {
		return "", ErrInvalidITIPEmail
	}
	return normalized, nil
}

func normalizeAndValidateITIPAttendees(attendees []ITIPParticipant, method ITIPMethod) error {
	if method == ITIPMethodReply && len(attendees) != 1 {
		return ErrInvalidITIPParticipant
	}
	for index := range attendees {
		if method == ITIPMethodReply && attendees[index].Status == RSVPNeedsAction {
			return ErrInvalidITIPParticipant
		}
	}
	sort.Slice(attendees, func(left, right int) bool {
		return attendees[left].Email < attendees[right].Email
	})
	for index := 1; index < len(attendees); index++ {
		if attendees[index-1].Email == attendees[index].Email {
			return ErrInvalidITIPParticipant
		}
	}
	return nil
}

func decodeITIPText(value string) (string, error) {
	var builder strings.Builder
	builder.Grow(len(value))
	for index := 0; index < len(value); index++ {
		if value[index] != '\\' {
			builder.WriteByte(value[index])
			continue
		}
		index++
		if index >= len(value) {
			return "", ErrInvalidITIPSyntax
		}
		switch value[index] {
		case '\\':
			builder.WriteByte('\\')
		case ',':
			builder.WriteByte(',')
		case ';':
			builder.WriteByte(';')
		case 'n', 'N':
			builder.WriteByte('\n')
		default:
			return "", ErrInvalidITIPSyntax
		}
	}
	return builder.String(), nil
}

func decodeITIPParameter(value string) (string, error) {
	var builder strings.Builder
	builder.Grow(len(value))
	for index := 0; index < len(value); index++ {
		if value[index] != '^' {
			builder.WriteByte(value[index])
			continue
		}
		index++
		if index >= len(value) {
			return "", ErrInvalidITIPSyntax
		}
		switch value[index] {
		case '^':
			builder.WriteByte('^')
		case 'n', 'N':
			builder.WriteByte('\n')
		case '\'':
			builder.WriteByte('"')
		default:
			return "", ErrInvalidITIPSyntax
		}
	}
	return builder.String(), nil
}

func hasAnyITIPControl(value string) bool {
	for _, character := range value {
		if unicode.IsControl(character) {
			return true
		}
	}
	return false
}

func hasForbiddenITIPTextControl(value string) bool {
	for _, character := range value {
		if unicode.IsControl(character) && character != '\n' {
			return true
		}
	}
	return false
}

func normalizeITIPRRULE(value string, temporal ITIPTemporal) (string, error) {
	if value == "" || len(value) > itipHardMaxUnfoldedLineBytes || !utf8.ValidString(value) || hasAnyITIPControl(value) {
		return "", ErrInvalidITIPRecurrence
	}
	parts := strings.Split(value, ";")
	fields := make(map[string]string, len(parts))
	for _, part := range parts {
		equals := strings.IndexByte(part, '=')
		if equals <= 0 || equals == len(part)-1 {
			return "", ErrInvalidITIPRecurrence
		}
		key := strings.ToUpper(part[:equals])
		if _, exists := fields[key]; exists {
			return "", ErrInvalidITIPRecurrence
		}
		switch key {
		case "FREQ", "INTERVAL", "BYDAY", "BYMONTHDAY", "COUNT", "UNTIL":
		default:
			return "", ErrInvalidITIPRecurrence
		}
		fields[key] = part[equals+1:]
	}

	frequency := strings.ToUpper(fields["FREQ"])
	if frequency != "DAILY" && frequency != "WEEKLY" && frequency != "MONTHLY" {
		return "", ErrInvalidITIPRecurrence
	}
	interval := 1
	if rawInterval, exists := fields["INTERVAL"]; exists {
		parsed, err := parseBoundedPositiveITIPInt(rawInterval, 365)
		if err != nil {
			return "", ErrInvalidITIPRecurrence
		}
		interval = parsed
	}

	byday := ""
	byMonthDay := ""
	switch frequency {
	case "DAILY":
		if fields["BYDAY"] != "" || fields["BYMONTHDAY"] != "" {
			return "", ErrInvalidITIPRecurrence
		}
	case "WEEKLY":
		if fields["BYDAY"] == "" || fields["BYMONTHDAY"] != "" {
			return "", ErrInvalidITIPRecurrence
		}
		days, err := normalizeITIPWeekdays(fields["BYDAY"])
		if err != nil {
			return "", err
		}
		byday = days
	case "MONTHLY":
		if fields["BYMONTHDAY"] == "" || fields["BYDAY"] != "" {
			return "", ErrInvalidITIPRecurrence
		}
		day, err := parseBoundedPositiveITIPInt(fields["BYMONTHDAY"], 31)
		if err != nil {
			return "", ErrInvalidITIPRecurrence
		}
		byMonthDay = strconv.Itoa(day)
	}

	if fields["COUNT"] != "" && fields["UNTIL"] != "" {
		return "", ErrInvalidITIPRecurrence
	}
	count := ""
	if rawCount := fields["COUNT"]; rawCount != "" {
		parsed, err := parseBoundedPositiveITIPInt(rawCount, 1000)
		if err != nil {
			return "", ErrInvalidITIPRecurrence
		}
		count = strconv.Itoa(parsed)
	}
	until := ""
	if rawUntil := fields["UNTIL"]; rawUntil != "" {
		if temporal.Timed != nil {
			if _, err := parseITIPUTCInstant(rawUntil, ErrInvalidITIPRecurrence); err != nil {
				return "", ErrInvalidITIPRecurrence
			}
			until = rawUntil
		} else {
			date, err := parseCompactITIPDate(rawUntil)
			if err != nil {
				return "", ErrInvalidITIPRecurrence
			}
			until = strings.ReplaceAll(date, "-", "")
		}
	}

	normalized := "FREQ=" + frequency + ";INTERVAL=" + strconv.Itoa(interval)
	if byday != "" {
		normalized += ";BYDAY=" + byday
	}
	if byMonthDay != "" {
		normalized += ";BYMONTHDAY=" + byMonthDay
	}
	if count != "" {
		normalized += ";COUNT=" + count
	}
	if until != "" {
		normalized += ";UNTIL=" + until
	}
	if len(normalized) > itipMaxRRULEBytes {
		return "", ErrInvalidITIPRecurrence
	}
	return normalized, nil
}

func parseBoundedPositiveITIPInt(value string, maximum int) (int, error) {
	if value == "" {
		return 0, ErrInvalidITIPRecurrence
	}
	for index := 0; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return 0, ErrInvalidITIPRecurrence
		}
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 || parsed > maximum {
		return 0, ErrInvalidITIPRecurrence
	}
	return parsed, nil
}

func normalizeITIPWeekdays(value string) (string, error) {
	order := map[string]int{"MO": 0, "TU": 1, "WE": 2, "TH": 3, "FR": 4, "SA": 5, "SU": 6}
	days := strings.Split(value, ",")
	if len(days) < 1 || len(days) > 7 {
		return "", ErrInvalidITIPRecurrence
	}
	normalized := make([]string, 0, len(days))
	seen := make(map[string]struct{}, len(days))
	for _, day := range days {
		day = strings.ToUpper(day)
		if _, exists := order[day]; !exists {
			return "", ErrInvalidITIPRecurrence
		}
		if _, exists := seen[day]; exists {
			return "", ErrInvalidITIPRecurrence
		}
		seen[day] = struct{}{}
		normalized = append(normalized, day)
	}
	sort.Slice(normalized, func(left, right int) bool {
		return order[normalized[left]] < order[normalized[right]]
	})
	return strings.Join(normalized, ","), nil
}

func sortAndRejectDuplicateITIPRecurrences(values []ITIPRecurrenceID, temporal ITIPTemporal) error {
	sort.Slice(values, func(left, right int) bool {
		return canonicalParsedITIPRecurrence(values[left], temporal) < canonicalParsedITIPRecurrence(values[right], temporal)
	})
	for index := 1; index < len(values); index++ {
		if canonicalParsedITIPRecurrence(values[index-1], temporal) == canonicalParsedITIPRecurrence(values[index], temporal) {
			return ErrInvalidITIPRecurrence
		}
	}
	return nil
}

func canonicalParsedITIPRecurrence(value ITIPRecurrenceID, temporal ITIPTemporal) string {
	if temporal.Timed != nil {
		return value.Time.UTC().Format(time.RFC3339)
	}
	return value.Date
}

// CanonicalITIPIdentity строит CVE-v1 identity только из уже нормализованных доменных значений.
func CanonicalITIPIdentity(message ITIPMessage) (ITIPIdentity, error) {
	normalized, err := normalizeITIPIdentityMessage(message)
	if err != nil {
		return ITIPIdentity{}, err
	}

	recurrence := ""
	hasRecurrence := normalized.RecurrenceID != nil
	if hasRecurrence {
		recurrence = canonicalParsedITIPRecurrence(*normalized.RecurrenceID, normalized.Temporal)
	}
	return ITIPIdentity{
		CanonicalVersion: itipCanonicalVersion,
		Method:           normalized.Method,
		UID:              normalized.UID,
		HasRecurrenceID:  hasRecurrence,
		RecurrenceID:     recurrence,
		Sequence:         normalized.Sequence,
		PayloadHash:      hashCanonicalITIPMessage(normalized),
	}, nil
}

func normalizeITIPIdentityMessage(message ITIPMessage) (ITIPMessage, error) {
	if !isITIPMethod(message.Method) {
		return ITIPMessage{}, ErrInvalidITIPMethod
	}
	if message.Sequence < 0 || message.Sequence > itipMaxSequence {
		return ITIPMessage{}, ErrInvalidITIPSequence
	}
	if err := validateITIPUID(message.UID); err != nil {
		return ITIPMessage{}, err
	}
	if message.Stamp.IsZero() || message.Stamp.Location() != time.UTC || message.Stamp.Nanosecond() != 0 {
		return ITIPMessage{}, ErrInvalidITIPStamp
	}
	if !isITIPStatus(message.Status) || !isITIPMethodStatus(message.Method, message.Status) {
		return ITIPMessage{}, ErrInvalidITIPStatus
	}
	if message.Summary == "" || !utf8.ValidString(message.Summary) || len(message.Summary) > itipMaxSummaryBytes || hasForbiddenITIPTextControl(message.Summary) {
		return ITIPMessage{}, ErrInvalidITIPMessage
	}
	if !utf8.ValidString(message.Description) || len(message.Description) > itipMaxDescriptionBytes || hasForbiddenITIPTextControl(message.Description) ||
		!utf8.ValidString(message.Location) || len(message.Location) > itipMaxLocationBytes || hasForbiddenITIPTextControl(message.Location) {
		return ITIPMessage{}, ErrInvalidITIPMessage
	}

	normalized := message
	normalized.Stamp = message.Stamp.UTC()
	if err := normalizeITIPIdentityTemporal(&normalized); err != nil {
		return ITIPMessage{}, err
	}

	if len(message.Organizer.Email) > itipMaxEmailBytes {
		return ITIPMessage{}, ErrInvalidITIPEmail
	}
	organizerEmail, err := normalizeSafeITIPEmail(message.Organizer.Email)
	if err != nil || len(organizerEmail) > itipMaxEmailBytes {
		return ITIPMessage{}, ErrInvalidITIPEmail
	}
	if !utf8.ValidString(message.Organizer.DisplayName) || len(message.Organizer.DisplayName) > itipMaxDisplayNameBytes || hasForbiddenITIPTextControl(message.Organizer.DisplayName) {
		return ITIPMessage{}, ErrInvalidITIPMessage
	}
	normalized.Organizer.Email = organizerEmail

	if len(message.Attendees) > itipHardMaxAttendees {
		return ITIPMessage{}, ErrITIPLimitExceeded
	}
	normalized.Attendees = append([]ITIPParticipant(nil), message.Attendees...)
	for index := range normalized.Attendees {
		attendee := &normalized.Attendees[index]
		if len(attendee.Email) > itipMaxEmailBytes {
			return ITIPMessage{}, ErrInvalidITIPEmail
		}
		email, emailErr := normalizeSafeITIPEmail(attendee.Email)
		if emailErr != nil || len(email) > itipMaxEmailBytes {
			return ITIPMessage{}, ErrInvalidITIPEmail
		}
		attendee.Email = email
		if !utf8.ValidString(attendee.DisplayName) || len(attendee.DisplayName) > itipMaxDisplayNameBytes || hasForbiddenITIPTextControl(attendee.DisplayName) ||
			!isITIPParticipantRole(attendee.Role) {
			return ITIPMessage{}, ErrInvalidITIPParticipant
		}
		if err := attendee.Status.Validate(); err != nil {
			return ITIPMessage{}, ErrInvalidITIPParticipant
		}
	}
	if err := normalizeAndValidateITIPAttendees(normalized.Attendees, normalized.Method); err != nil {
		return ITIPMessage{}, err
	}

	if message.RRULE != "" {
		normalized.RRULE, err = normalizeITIPRRULE(message.RRULE, normalized.Temporal)
		if err != nil {
			return ITIPMessage{}, err
		}
	}

	if message.RecurrenceID != nil {
		recurrence, recurrenceErr := normalizeDomainITIPRecurrence(*message.RecurrenceID, normalized.Temporal)
		if recurrenceErr != nil {
			return ITIPMessage{}, recurrenceErr
		}
		normalized.RecurrenceID = &recurrence
	}
	if len(message.EXDates) > itipHardMaxExDates {
		return ITIPMessage{}, ErrITIPLimitExceeded
	}
	normalized.EXDates = make([]ITIPRecurrenceID, len(message.EXDates))
	for index, exdate := range message.EXDates {
		normalizedExdate, exdateErr := normalizeDomainITIPRecurrence(exdate, normalized.Temporal)
		if exdateErr != nil {
			return ITIPMessage{}, exdateErr
		}
		normalized.EXDates[index] = normalizedExdate
	}
	if err := sortAndRejectDuplicateITIPRecurrences(normalized.EXDates, normalized.Temporal); err != nil {
		return ITIPMessage{}, err
	}
	return normalized, nil
}

func normalizeITIPIdentityTemporal(message *ITIPMessage) error {
	temporal := message.Temporal
	if (temporal.Timed == nil) == (temporal.AllDay == nil) {
		return ErrInvalidITIPTemporal
	}
	if temporal.Timed != nil {
		value := *temporal.Timed
		if value.StartsAt.IsZero() || value.EndsAt.IsZero() || value.StartsAt.Nanosecond() != 0 || value.EndsAt.Nanosecond() != 0 ||
			!value.EndsAt.After(value.StartsAt) {
			return ErrInvalidITIPTemporal
		}
		if _, err := loadCanonicalITIPLocation(value.Timezone); err != nil {
			return err
		}
		value.StartsAt = value.StartsAt.UTC()
		value.EndsAt = value.EndsAt.UTC()
		message.Temporal = ITIPTemporal{Timed: &value}
		return nil
	}

	value := *temporal.AllDay
	start, err := parseCanonicalITIPDate(value.StartDate)
	if err != nil {
		return ErrInvalidITIPTemporal
	}
	end, err := parseCanonicalITIPDate(value.EndDateExclusive)
	if err != nil || !end.After(start) || value.Timezone != "UTC" {
		return ErrInvalidITIPTemporal
	}
	message.Temporal = ITIPTemporal{AllDay: &ITIPAllDay{
		StartDate:        value.StartDate,
		EndDateExclusive: value.EndDateExclusive,
		Timezone:         "UTC",
	}}
	return nil
}

func normalizeDomainITIPRecurrence(value ITIPRecurrenceID, temporal ITIPTemporal) (ITIPRecurrenceID, error) {
	if temporal.Timed != nil {
		if value.Time.IsZero() || value.Time.Nanosecond() != 0 || value.Date != "" {
			return ITIPRecurrenceID{}, ErrInvalidITIPRecurrence
		}
		return ITIPRecurrenceID{Time: value.Time.UTC()}, nil
	}
	if !value.Time.IsZero() || value.Date == "" {
		return ITIPRecurrenceID{}, ErrInvalidITIPRecurrence
	}
	if _, err := parseCanonicalITIPDate(value.Date); err != nil {
		return ITIPRecurrenceID{}, ErrInvalidITIPRecurrence
	}
	return ITIPRecurrenceID{Date: value.Date}, nil
}

type itipCVEWriter struct {
	hash    hash.Hash
	scratch [8]byte
}

func hashCanonicalITIPMessage(message ITIPMessage) [32]byte {
	writer := itipCVEWriter{hash: sha256.New()}
	writer.writeRaw([]byte("hub.calendar.itip.payload"))
	writer.writeByte(0)
	writer.writeByte(itipCanonicalVersion)

	writer.writeFrameHeader(1, 1)
	writer.writeByte(canonicalITIPMethodByte(message.Method))
	writer.writeRawFrame(2, message.UID)
	if message.RecurrenceID == nil {
		writer.writeFrameHeader(3, 1)
		writer.writeByte(0)
	} else {
		length := uint32(1) + canonicalITIPRecurrenceRecordLength(*message.RecurrenceID, message.Temporal)
		writer.writeFrameHeader(3, length)
		writer.writeByte(1)
		writer.writeRecurrenceRecord(*message.RecurrenceID, message.Temporal)
	}
	writer.writeFrameHeader(4, 8)
	writer.writeUint64(uint64(message.Sequence))
	writer.writeFrameHeader(5, 8)
	writer.writeInt64(message.Stamp.Unix())
	writer.writeFrameHeader(6, 1)
	writer.writeByte(canonicalITIPStatusByte(message.Status))

	writer.writeCanonicalTemporalFrame(message.Temporal)
	organizerLength := canonicalITIPStringLength(message.Organizer.Email) + canonicalITIPStringLength(message.Organizer.DisplayName)
	writer.writeFrameHeader(8, organizerLength)
	writer.writeString(message.Organizer.Email)
	writer.writeString(message.Organizer.DisplayName)
	writer.writeCanonicalAttendeeFrame(message.Attendees)
	writer.writeRawFrame(10, message.Summary)
	writer.writeRawFrame(11, message.Description)
	writer.writeRawFrame(12, message.Location)
	if message.RRULE == "" {
		writer.writeFrameHeader(13, 1)
		writer.writeByte(0)
	} else {
		writer.writeFrameHeader(13, 1+canonicalITIPStringLength(message.RRULE))
		writer.writeByte(1)
		writer.writeString(message.RRULE)
	}
	writer.writeCanonicalExdateFrame(message.EXDates, message.Temporal)

	var result [32]byte
	copy(result[:], writer.hash.Sum(nil))
	return result
}

func canonicalITIPMethodByte(method ITIPMethod) byte {
	switch method {
	case ITIPMethodRequest:
		return 1
	case ITIPMethodCancel:
		return 2
	default:
		return 3
	}
}

func canonicalITIPStatusByte(status ITIPStatus) byte {
	switch status {
	case ITIPStatusConfirmed:
		return 1
	case ITIPStatusTentative:
		return 2
	default:
		return 3
	}
}

func canonicalITIPRoleByte(role ITIPParticipantRole) byte {
	if role == ITIPParticipantRequired {
		return 1
	}
	return 2
}

func canonicalITIPPartstatByte(status RSVPStatus) byte {
	switch status {
	case RSVPNeedsAction:
		return 1
	case RSVPAccepted:
		return 2
	case RSVPDeclined:
		return 3
	default:
		return 4
	}
}

func canonicalITIPStringLength(value string) uint32 {
	return 4 + uint32(len(value))
}

func canonicalITIPRecurrenceRecordLength(value ITIPRecurrenceID, temporal ITIPTemporal) uint32 {
	if temporal.Timed != nil {
		return 1 + 8
	}
	return 1 + canonicalITIPStringLength(value.Date)
}

func (writer *itipCVEWriter) writeRaw(value []byte) {
	_, _ = writer.hash.Write(value)
}

func (writer *itipCVEWriter) writeByte(value byte) {
	writer.scratch[0] = value
	_, _ = writer.hash.Write(writer.scratch[:1])
}

func (writer *itipCVEWriter) writeUint32(value uint32) {
	binary.BigEndian.PutUint32(writer.scratch[:4], value)
	_, _ = writer.hash.Write(writer.scratch[:4])
}

func (writer *itipCVEWriter) writeUint64(value uint64) {
	binary.BigEndian.PutUint64(writer.scratch[:], value)
	_, _ = writer.hash.Write(writer.scratch[:])
}

func (writer *itipCVEWriter) writeInt64(value int64) {
	writer.writeUint64(uint64(value))
}

func (writer *itipCVEWriter) writeFrameHeader(tag byte, length uint32) {
	writer.writeByte(tag)
	writer.writeUint32(length)
}

func (writer *itipCVEWriter) writeRawFrame(tag byte, value string) {
	writer.writeFrameHeader(tag, uint32(len(value)))
	writer.writeRaw([]byte(value))
}

func (writer *itipCVEWriter) writeString(value string) {
	writer.writeUint32(uint32(len(value)))
	writer.writeRaw([]byte(value))
}

func (writer *itipCVEWriter) writeBool(value bool) {
	if value {
		writer.writeByte(1)
		return
	}
	writer.writeByte(0)
}

func (writer *itipCVEWriter) writeRecurrenceRecord(value ITIPRecurrenceID, temporal ITIPTemporal) {
	if temporal.Timed != nil {
		writer.writeByte(1)
		writer.writeInt64(value.Time.UTC().Unix())
		return
	}
	writer.writeByte(2)
	writer.writeString(value.Date)
}

func (writer *itipCVEWriter) writeCanonicalTemporalFrame(temporal ITIPTemporal) {
	if temporal.Timed != nil {
		length := uint32(1) + canonicalITIPStringLength(temporal.Timed.Timezone) + 8 + 8
		writer.writeFrameHeader(7, length)
		writer.writeByte(1)
		writer.writeString(temporal.Timed.Timezone)
		writer.writeInt64(temporal.Timed.StartsAt.UTC().Unix())
		writer.writeInt64(temporal.Timed.EndsAt.UTC().Unix())
		return
	}
	length := uint32(1) + canonicalITIPStringLength("UTC") +
		canonicalITIPStringLength(temporal.AllDay.StartDate) +
		canonicalITIPStringLength(temporal.AllDay.EndDateExclusive)
	writer.writeFrameHeader(7, length)
	writer.writeByte(2)
	writer.writeString("UTC")
	writer.writeString(temporal.AllDay.StartDate)
	writer.writeString(temporal.AllDay.EndDateExclusive)
}

func (writer *itipCVEWriter) writeCanonicalAttendeeFrame(attendees []ITIPParticipant) {
	length := uint32(4)
	for _, attendee := range attendees {
		recordLength := canonicalITIPStringLength(attendee.Email) + canonicalITIPStringLength(attendee.DisplayName) + 3
		length += 4 + recordLength
	}
	writer.writeFrameHeader(9, length)
	writer.writeUint32(uint32(len(attendees)))
	for _, attendee := range attendees {
		recordLength := canonicalITIPStringLength(attendee.Email) + canonicalITIPStringLength(attendee.DisplayName) + 3
		writer.writeUint32(recordLength)
		writer.writeString(attendee.Email)
		writer.writeString(attendee.DisplayName)
		writer.writeByte(canonicalITIPRoleByte(attendee.Role))
		writer.writeByte(canonicalITIPPartstatByte(attendee.Status))
		writer.writeBool(attendee.RSVP)
	}
}

func (writer *itipCVEWriter) writeCanonicalExdateFrame(exdates []ITIPRecurrenceID, temporal ITIPTemporal) {
	length := uint32(4)
	for _, exdate := range exdates {
		length += 4 + canonicalITIPRecurrenceRecordLength(exdate, temporal)
	}
	writer.writeFrameHeader(14, length)
	writer.writeUint32(uint32(len(exdates)))
	for _, exdate := range exdates {
		recordLength := canonicalITIPRecurrenceRecordLength(exdate, temporal)
		writer.writeUint32(recordLength)
		writer.writeRecurrenceRecord(exdate, temporal)
	}
}
