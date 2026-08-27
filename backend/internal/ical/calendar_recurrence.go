package ical

import (
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	CalendarRecurrenceMaxRRuleBytes      = 1000
	CalendarRecurrenceOccurrenceCap      = 1000
	CalendarRecurrenceCandidateWorkGuard = 1001
)

// CalendarRecurrenceCivilTime is an exact local wall-clock value used only by
// the closed Calendar recurrence grammar. It has no location because a caller
// must resolve it against the master event's IANA timezone.
type CalendarRecurrenceCivilTime struct {
	Year       int
	Month      time.Month
	Day        int
	Hour       int
	Minute     int
	Second     int
	Nanosecond int
}

// NominalOccurrence is the canonical recurrence projection before EXDATE and
// exception policy. Temporal preserves the master's timed/all-day form.
type NominalOccurrence struct {
	RecurrenceID string
	Temporal     CalendarEvent
}

// NormalizeCalendarRRULE implements Hub's closed Calendar RRULE grammar. It
// intentionally does not reuse inbound iTIP repair rules.
func NormalizeCalendarRRULE(temporal CalendarEvent, value string) (string, bool) {
	if temporal.Validate() != nil || value == "" || len(value) > CalendarRecurrenceMaxRRuleBytes || !utf8.ValidString(value) {
		return "", false
	}
	fields := make(map[string]string, 6)
	for _, part := range strings.Split(value, ";") {
		key, rawValue, found := strings.Cut(part, "=")
		if !found || key == "" || rawValue == "" || strings.Contains(rawValue, "=") {
			return "", false
		}
		switch key {
		case "FREQ", "INTERVAL", "BYDAY", "BYMONTHDAY", "COUNT", "UNTIL":
		default:
			return "", false
		}
		if _, duplicate := fields[key]; duplicate {
			return "", false
		}
		fields[key] = rawValue
	}

	frequency, found := fields["FREQ"]
	if !found || frequency != "DAILY" && frequency != "WEEKLY" && frequency != "MONTHLY" {
		return "", false
	}
	interval := 1
	if rawInterval, found := fields["INTERVAL"]; found {
		var ok bool
		interval, ok = ParseCalendarRecurrencePositiveInteger(rawInterval, 365)
		if !ok {
			return "", false
		}
	}

	byday, byMonthDay := "", ""
	switch frequency {
	case "DAILY":
		if _, found := fields["BYDAY"]; found {
			return "", false
		}
		if _, found := fields["BYMONTHDAY"]; found {
			return "", false
		}
	case "WEEKLY":
		rawDays, found := fields["BYDAY"]
		if !found {
			return "", false
		}
		if _, found := fields["BYMONTHDAY"]; found {
			return "", false
		}
		normalizedDays, ok := normalizeCalendarRecurrenceWeekdays(rawDays)
		if !ok {
			return "", false
		}
		byday = normalizedDays
	case "MONTHLY":
		rawMonthDay, found := fields["BYMONTHDAY"]
		if !found {
			return "", false
		}
		if _, found := fields["BYDAY"]; found {
			return "", false
		}
		parsed, ok := ParseCalendarRecurrencePositiveInteger(rawMonthDay, 31)
		if !ok {
			return "", false
		}
		byMonthDay = strconv.Itoa(parsed)
	}

	rawCount, hasCount := fields["COUNT"]
	rawUntil, hasUntil := fields["UNTIL"]
	if hasCount == hasUntil {
		return "", false
	}
	count := ""
	if hasCount {
		parsed, ok := ParseCalendarRecurrencePositiveInteger(rawCount, CalendarRecurrenceOccurrenceCap)
		if !ok {
			return "", false
		}
		count = strconv.Itoa(parsed)
	}
	until := ""
	if hasUntil {
		var ok bool
		if temporal.AllDay {
			until, ok = normalizeCalendarRecurrenceAllDayUntil(rawUntil)
		} else {
			until, ok = normalizeCalendarRecurrenceTimedUntil(rawUntil)
		}
		if !ok || !calendarRecurrenceOccurrenceCapValid(temporal, frequency, interval, byday, byMonthDay, until) {
			return "", false
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
	return normalized, true
}

// ResolveNominalOccurrence validates a canonical master RRULE and returns one
// exact nominal member. EXDATE and exception state deliberately remain with
// the authoritative Calendar/Mail caller.
func ResolveNominalOccurrence(event CalendarEvent, rrule, recurrenceID string) (NominalOccurrence, bool) {
	normalized, ok := NormalizeCalendarRRULE(event, rrule)
	if !ok || normalized != rrule {
		return NominalOccurrence{}, false
	}
	if event.AllDay {
		return resolveCalendarRecurrenceAllDayOccurrence(event, rrule, recurrenceID)
	}
	return resolveCalendarRecurrenceTimedOccurrence(event, rrule, recurrenceID)
}

func resolveCalendarRecurrenceAllDayOccurrence(event CalendarEvent, rrule, recurrenceID string) (NominalOccurrence, bool) {
	requested, err := time.Parse(time.DateOnly, recurrenceID)
	if err != nil || requested.Format(time.DateOnly) != recurrenceID {
		return NominalOccurrence{}, false
	}
	if !calendarRecurrenceAllDayMember(event, rrule, requested) {
		return NominalOccurrence{}, false
	}
	start, err := ParseCalendarDate(event.StartDate)
	if err != nil {
		return NominalOccurrence{}, false
	}
	end, err := ParseCalendarDate(event.EndDateExclusive)
	if err != nil || !end.After(start) {
		return NominalOccurrence{}, false
	}
	days := int(end.Sub(start) / (24 * time.Hour))
	if days < 1 {
		return NominalOccurrence{}, false
	}
	temporal := CalendarEvent{
		AllDay:           true,
		StartDate:        recurrenceID,
		EndDateExclusive: requested.AddDate(0, 0, days).Format(time.DateOnly),
		Timezone:         event.Timezone,
	}
	if temporal.Validate() != nil {
		return NominalOccurrence{}, false
	}
	return NominalOccurrence{RecurrenceID: recurrenceID, Temporal: temporal}, true
}

func resolveCalendarRecurrenceTimedOccurrence(event CalendarEvent, rrule, recurrenceID string) (NominalOccurrence, bool) {
	requested, ok := parseCalendarRecurrenceTimedKey(recurrenceID)
	if !ok || !calendarRecurrenceTimedMember(event, rrule, requested) {
		return NominalOccurrence{}, false
	}
	duration := event.EndsAt.Sub(event.StartsAt)
	if duration <= 0 {
		return NominalOccurrence{}, false
	}
	temporal := CalendarEvent{
		StartsAt: requested,
		EndsAt:   requested.Add(duration),
		Timezone: event.Timezone,
	}
	if temporal.Validate() != nil {
		return NominalOccurrence{}, false
	}
	return NominalOccurrence{RecurrenceID: recurrenceID, Temporal: temporal}, true
}

func calendarRecurrenceAllDayMember(event CalendarEvent, rrule string, requested time.Time) bool {
	start, err := ParseCalendarDate(event.StartDate)
	if err != nil || requested.Before(start) {
		return false
	}
	if requested.Equal(start) {
		return true
	}
	fields := calendarRecurrenceFields(rrule)
	interval, ok := calendarRecurrencePositiveRuleField(fields, "INTERVAL", 1)
	if !ok {
		return false
	}
	count := 0
	if value, found := fields["COUNT"]; found {
		count, ok = ParseCalendarRecurrencePositiveInteger(value, CalendarRecurrenceOccurrenceCap)
		if !ok {
			return false
		}
	}
	var until time.Time
	hasUntil := false
	if value, found := fields["UNTIL"]; found {
		until, err = time.Parse("20060102", value)
		if err != nil {
			return false
		}
		hasUntil = true
	}
	seen, work := 1, 1
	check := func(candidate time.Time) (bool, bool) {
		if !candidate.After(start) {
			return false, false
		}
		if hasUntil && candidate.After(until) {
			return false, true
		}
		seen++
		if seen > CalendarRecurrenceOccurrenceCap || count > 0 && seen > count {
			return false, true
		}
		if candidate.Equal(requested) {
			return true, true
		}
		return false, candidate.After(requested)
	}
	switch fields["FREQ"] {
	case "DAILY":
		for sequence := 1; ; sequence++ {
			work++
			if work >= CalendarRecurrenceCandidateWorkGuard {
				return false
			}
			if found, stop := check(start.AddDate(0, 0, sequence*interval)); found || stop {
				return found
			}
		}
	case "WEEKLY":
		days, ok := CalendarRecurrenceWeekdayOffsets(fields["BYDAY"])
		if !ok {
			return false
		}
		weekStart := CalendarRecurrenceMonday(start)
		for group := 0; ; group++ {
			base := weekStart.AddDate(0, 0, group*interval*7)
			for _, day := range days {
				candidate := base.AddDate(0, 0, day)
				if candidate.After(start) {
					work++
					if work >= CalendarRecurrenceCandidateWorkGuard {
						return false
					}
				}
				if found, stop := check(candidate); found || stop {
					return found
				}
			}
		}
	case "MONTHLY":
		monthDay, ok := ParseCalendarRecurrencePositiveInteger(fields["BYMONTHDAY"], 31)
		if !ok {
			return false
		}
		for sequence := 0; ; sequence++ {
			month := time.Date(start.Year(), start.Month()+time.Month(sequence*interval), 1, 0, 0, 0, 0, time.UTC)
			if hasUntil && month.After(until) {
				return false
			}
			if monthDay > CalendarRecurrenceDaysInMonth(month.Year(), month.Month()) {
				work++
				if work >= CalendarRecurrenceCandidateWorkGuard {
					return false
				}
				continue
			}
			candidate := time.Date(month.Year(), month.Month(), monthDay, 0, 0, 0, 0, time.UTC)
			if candidate.After(start) {
				work++
				if work >= CalendarRecurrenceCandidateWorkGuard {
					return false
				}
			}
			if found, stop := check(candidate); found || stop {
				return found
			}
		}
	}
	return false
}

func calendarRecurrenceTimedMember(event CalendarEvent, rrule string, requested time.Time) bool {
	if !validCalendarRecurrenceUTCSecond(requested) || requested.Before(event.StartsAt) {
		return false
	}
	if requested.Equal(event.StartsAt) {
		return true
	}
	fields := calendarRecurrenceFields(rrule)
	interval, ok := calendarRecurrencePositiveRuleField(fields, "INTERVAL", 1)
	if !ok {
		return false
	}
	count := 0
	if value, found := fields["COUNT"]; found {
		count, ok = ParseCalendarRecurrencePositiveInteger(value, CalendarRecurrenceOccurrenceCap)
		if !ok {
			return false
		}
	}
	var until time.Time
	hasUntil := false
	if value, found := fields["UNTIL"]; found {
		parsed, err := time.Parse("20060102T150405Z", value)
		if err != nil || parsed.Location() != time.UTC || parsed.Format("20060102T150405Z") != value {
			return false
		}
		until, hasUntil = parsed, true
	}
	location, err := time.LoadLocation(event.Timezone)
	if err != nil {
		return false
	}
	localStart := CalendarRecurrenceCivilTimeFromTime(event.StartsAt.In(location))
	localUntil := CalendarRecurrenceCivilTime{}
	if hasUntil {
		localUntil = CalendarRecurrenceCivilTimeFromTime(until.In(location))
	}
	accepted, work := 1, 1
	takeWork := func() bool {
		work++
		return work < CalendarRecurrenceCandidateWorkGuard
	}
	check := func(civil CalendarRecurrenceCivilTime) (bool, bool) {
		if CompareCalendarRecurrenceCivilTime(civil, localStart) <= 0 {
			return false, false
		}
		if count > 0 && accepted >= count || !takeWork() {
			return false, true
		}
		candidate, exists := ResolveCalendarRecurrenceCivilTime(location, civil)
		if !exists {
			return false, hasUntil && CompareCalendarRecurrenceCivilTime(civil, localUntil) > 0
		}
		candidate = candidate.UTC()
		if hasUntil && candidate.After(until) || !validCalendarRecurrenceUTCSecond(candidate) {
			return false, true
		}
		accepted++
		if accepted > CalendarRecurrenceOccurrenceCap {
			return false, true
		}
		if candidate.Equal(requested) {
			return true, true
		}
		return false, candidate.After(requested)
	}
	switch fields["FREQ"] {
	case "DAILY":
		for sequence := 1; ; sequence++ {
			date := time.Date(localStart.Year, localStart.Month, localStart.Day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, sequence*interval)
			civil := localStart
			civil.Year, civil.Month, civil.Day = date.Date()
			if found, stop := check(civil); found || stop {
				return found
			}
		}
	case "WEEKLY":
		days, ok := CalendarRecurrenceWeekdayOffsets(fields["BYDAY"])
		if !ok {
			return false
		}
		weekStart := CalendarRecurrenceMonday(time.Date(localStart.Year, localStart.Month, localStart.Day, 0, 0, 0, 0, time.UTC))
		for group := 0; ; group++ {
			for _, day := range days {
				civil := localStart
				civil.Year, civil.Month, civil.Day = weekStart.AddDate(0, 0, group*interval*7+day).Date()
				if found, stop := check(civil); found || stop {
					return found
				}
			}
		}
	case "MONTHLY":
		monthDay, ok := ParseCalendarRecurrencePositiveInteger(fields["BYMONTHDAY"], 31)
		if !ok {
			return false
		}
		for sequence := 0; ; sequence++ {
			month := time.Date(localStart.Year, localStart.Month+time.Month(sequence*interval), 1, 0, 0, 0, 0, time.UTC)
			if monthDay > CalendarRecurrenceDaysInMonth(month.Year(), month.Month()) {
				if hasUntil && CompareCalendarRecurrenceCivilTime(CalendarRecurrenceCivilTime{Year: month.Year(), Month: month.Month(), Day: 1}, localUntil) > 0 {
					return false
				}
				if !takeWork() {
					return false
				}
				continue
			}
			civil := localStart
			civil.Year, civil.Month, civil.Day = month.Year(), month.Month(), monthDay
			if found, stop := check(civil); found || stop {
				return found
			}
		}
	}
	return false
}

func calendarRecurrenceOccurrenceCapValid(temporal CalendarEvent, frequency string, interval int, byday, byMonthDay, until string) bool {
	if temporal.AllDay {
		return calendarRecurrenceAllDayOccurrenceCapValid(temporal, frequency, interval, byday, byMonthDay, until)
	}
	untilTime, err := time.Parse("20060102T150405Z", until)
	if err != nil || temporal.StartsAt.After(untilTime) {
		return false
	}
	location, err := time.LoadLocation(temporal.Timezone)
	if err != nil {
		return false
	}
	return calendarRecurrenceTimedOccurrenceCapValid(temporal.StartsAt, untilTime, location, frequency, interval, byday, byMonthDay)
}

func calendarRecurrenceAllDayOccurrenceCapValid(temporal CalendarEvent, frequency string, interval int, byday, byMonthDay, until string) bool {
	startDate, err := ParseCalendarDate(temporal.StartDate)
	if err != nil {
		return false
	}
	untilDate, err := time.Parse("20060102", until)
	if err != nil || untilDate.Before(startDate) {
		return false
	}
	work, occurrences := 1, 1
	process := func(candidate time.Time) (bool, bool) {
		if !candidate.After(startDate) {
			return true, false
		}
		if candidate.After(untilDate) {
			return true, true
		}
		work++
		if work >= CalendarRecurrenceCandidateWorkGuard {
			return false, false
		}
		occurrences++
		return occurrences <= CalendarRecurrenceOccurrenceCap, false
	}
	switch frequency {
	case "DAILY":
		for sequence := 1; ; sequence++ {
			valid, stop := process(startDate.AddDate(0, 0, sequence*interval))
			if !valid || stop {
				return valid && stop
			}
		}
	case "WEEKLY":
		weekdays, ok := CalendarRecurrenceWeekdayOffsets(byday)
		if !ok {
			return false
		}
		weekStart := CalendarRecurrenceMonday(startDate)
		for group := 0; ; group++ {
			base := weekStart.AddDate(0, 0, group*interval*7)
			for _, weekday := range weekdays {
				valid, stop := process(base.AddDate(0, 0, weekday))
				if !valid || stop {
					return valid && stop
				}
			}
		}
	case "MONTHLY":
		monthDay, err := strconv.Atoi(byMonthDay)
		if err != nil {
			return false
		}
		for sequence := 0; ; sequence++ {
			month := time.Date(startDate.Year(), startDate.Month()+time.Month(sequence*interval), 1, 0, 0, 0, 0, time.UTC)
			if monthDay > CalendarRecurrenceDaysInMonth(month.Year(), month.Month()) {
				if month.After(untilDate) {
					return true
				}
				work++
				if work >= CalendarRecurrenceCandidateWorkGuard {
					return false
				}
				continue
			}
			valid, stop := process(time.Date(month.Year(), month.Month(), monthDay, 0, 0, 0, 0, time.UTC))
			if !valid || stop {
				return valid && stop
			}
		}
	}
	return false
}

func calendarRecurrenceTimedOccurrenceCapValid(start, until time.Time, location *time.Location, frequency string, interval int, byday, byMonthDay string) bool {
	localStart := CalendarRecurrenceCivilTimeFromTime(start.UTC().In(location))
	localUntil := CalendarRecurrenceCivilTimeFromTime(until.In(location))
	work, occurrences := 1, 1
	process := func(civil CalendarRecurrenceCivilTime) (bool, bool) {
		if CompareCalendarRecurrenceCivilTime(civil, localStart) <= 0 {
			return true, false
		}
		candidate, exists := ResolveCalendarRecurrenceCivilTime(location, civil)
		if exists && candidate.After(until) || !exists && CompareCalendarRecurrenceCivilTime(civil, localUntil) > 0 {
			return true, true
		}
		work++
		if work >= CalendarRecurrenceCandidateWorkGuard {
			return false, false
		}
		if !exists {
			return true, false
		}
		occurrences++
		return occurrences <= CalendarRecurrenceOccurrenceCap, false
	}
	switch frequency {
	case "DAILY":
		for sequence := 1; ; sequence++ {
			date := time.Date(localStart.Year, localStart.Month, localStart.Day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, sequence*interval)
			civil := localStart
			civil.Year, civil.Month, civil.Day = date.Date()
			valid, stop := process(civil)
			if !valid || stop {
				return valid && stop
			}
		}
	case "WEEKLY":
		weekdays, ok := CalendarRecurrenceWeekdayOffsets(byday)
		if !ok {
			return false
		}
		weekStart := CalendarRecurrenceMonday(time.Date(localStart.Year, localStart.Month, localStart.Day, 0, 0, 0, 0, time.UTC))
		for group := 0; ; group++ {
			base := weekStart.AddDate(0, 0, group*interval*7)
			for _, weekday := range weekdays {
				civil := localStart
				civil.Year, civil.Month, civil.Day = base.AddDate(0, 0, weekday).Date()
				valid, stop := process(civil)
				if !valid || stop {
					return valid && stop
				}
			}
		}
	case "MONTHLY":
		monthDay, err := strconv.Atoi(byMonthDay)
		if err != nil {
			return false
		}
		for sequence := 0; ; sequence++ {
			month := time.Date(localStart.Year, localStart.Month+time.Month(sequence*interval), 1, 0, 0, 0, 0, time.UTC)
			if monthDay > CalendarRecurrenceDaysInMonth(month.Year(), month.Month()) {
				if CompareCalendarRecurrenceCivilTime(CalendarRecurrenceCivilTime{Year: month.Year(), Month: month.Month(), Day: 1}, localUntil) > 0 {
					return true
				}
				work++
				if work >= CalendarRecurrenceCandidateWorkGuard {
					return false
				}
				continue
			}
			civil := localStart
			civil.Year, civil.Month, civil.Day = month.Year(), month.Month(), monthDay
			valid, stop := process(civil)
			if !valid || stop {
				return valid && stop
			}
		}
	}
	return false
}

func ParseCalendarRecurrencePositiveInteger(value string, maximum int) (int, bool) {
	if value == "" {
		return 0, false
	}
	for index := 0; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return 0, false
		}
	}
	parsed, err := strconv.Atoi(value)
	return parsed, err == nil && parsed >= 1 && parsed <= maximum
}

func CalendarRecurrenceWeekdayOffsets(value string) ([]int, bool) {
	offsets := map[string]int{"MO": 0, "TU": 1, "WE": 2, "TH": 3, "FR": 4, "SA": 5, "SU": 6}
	parts := strings.Split(value, ",")
	result := make([]int, len(parts))
	for index, part := range parts {
		offset, found := offsets[part]
		if !found {
			return nil, false
		}
		result[index] = offset
	}
	return result, true
}

func CalendarRecurrenceMonday(value time.Time) time.Time {
	offset := (int(value.Weekday()) + 6) % 7
	return value.AddDate(0, 0, -offset)
}

func CalendarRecurrenceDaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func CalendarRecurrenceCivilTimeFromTime(value time.Time) CalendarRecurrenceCivilTime {
	return CalendarRecurrenceCivilTime{Year: value.Year(), Month: value.Month(), Day: value.Day(), Hour: value.Hour(), Minute: value.Minute(), Second: value.Second(), Nanosecond: value.Nanosecond()}
}

func CompareCalendarRecurrenceCivilTime(left, right CalendarRecurrenceCivilTime) int {
	for _, pair := range [][2]int{{left.Year, right.Year}, {int(left.Month), int(right.Month)}, {left.Day, right.Day}, {left.Hour, right.Hour}, {left.Minute, right.Minute}, {left.Second, right.Second}, {left.Nanosecond, right.Nanosecond}} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	return 0
}

func ResolveCalendarRecurrenceCivilTime(location *time.Location, civil CalendarRecurrenceCivilTime) (time.Time, bool) {
	if location == nil {
		return time.Time{}, false
	}
	naive := time.Date(civil.Year, civil.Month, civil.Day, civil.Hour, civil.Minute, civil.Second, civil.Nanosecond, time.UTC)
	offsets := make(map[int]struct{}, 17)
	for hours := -48; hours <= 48; hours += 6 {
		_, offset := naive.Add(time.Duration(hours) * time.Hour).In(location).Zone()
		offsets[offset] = struct{}{}
	}
	var earliest time.Time
	found := false
	for offset := range offsets {
		candidate := naive.Add(-time.Duration(offset) * time.Second)
		if CompareCalendarRecurrenceCivilTime(CalendarRecurrenceCivilTimeFromTime(candidate.In(location)), civil) != 0 {
			continue
		}
		if !found || candidate.Before(earliest) {
			earliest, found = candidate, true
		}
	}
	return earliest, found
}

func calendarRecurrenceFields(value string) map[string]string {
	fields := make(map[string]string, 6)
	for _, field := range strings.Split(value, ";") {
		key, fieldValue, found := strings.Cut(field, "=")
		if !found || key == "" || fieldValue == "" {
			return nil
		}
		fields[key] = fieldValue
	}
	return fields
}

func calendarRecurrencePositiveRuleField(fields map[string]string, key string, defaultValue int) (int, bool) {
	value, found := fields[key]
	if !found {
		return defaultValue, true
	}
	return ParseCalendarRecurrencePositiveInteger(value, CalendarRecurrenceOccurrenceCap)
}

func normalizeCalendarRecurrenceWeekdays(value string) (string, bool) {
	order := map[string]int{"MO": 0, "TU": 1, "WE": 2, "TH": 3, "FR": 4, "SA": 5, "SU": 6}
	days := strings.Split(value, ",")
	if len(days) < 1 || len(days) > 7 {
		return "", false
	}
	seen := make(map[string]struct{}, len(days))
	for _, day := range days {
		if _, found := order[day]; !found {
			return "", false
		}
		if _, duplicate := seen[day]; duplicate {
			return "", false
		}
		seen[day] = struct{}{}
	}
	sort.Slice(days, func(left, right int) bool { return order[days[left]] < order[days[right]] })
	return strings.Join(days, ","), true
}

func normalizeCalendarRecurrenceTimedUntil(value string) (string, bool) {
	parsed, err := time.Parse("20060102T150405Z", value)
	if err != nil || parsed.Location() != time.UTC || parsed.Format("20060102T150405Z") != value {
		return "", false
	}
	return parsed.Format("20060102T150405Z"), true
}

func normalizeCalendarRecurrenceAllDayUntil(value string) (string, bool) {
	parsed, err := time.Parse("20060102", value)
	if err != nil || parsed.Format("20060102") != value {
		return "", false
	}
	return parsed.Format("20060102"), true
}

func parseCalendarRecurrenceTimedKey(value string) (time.Time, bool) {
	if len(value) != len("2006-01-02T15:04:05Z") {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil || !validCalendarRecurrenceUTCSecond(parsed) || parsed.Year() < 0 || parsed.Year() > 9999 || parsed.Format("2006-01-02T15:04:05Z") != value {
		return time.Time{}, false
	}
	return parsed, true
}

func validCalendarRecurrenceUTCSecond(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond() == 0
}
