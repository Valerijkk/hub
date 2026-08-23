package calendar

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

const (
	AvailabilityMaxWindow            = 31 * 24 * time.Hour
	AvailabilityMinSlotMinutes       = 5
	AvailabilityMaxSlotMinutes       = 240
	AvailabilityMaxUsers             = 50
	AvailabilityMaxBusyIntervalsUser = 1000
	AvailabilityMaxBusyIntervalsAll  = 5000
)

// AvailabilityErrorKind is a safe category for availability policy failures.
type AvailabilityErrorKind string

const (
	AvailabilityErrorLimitExceeded AvailabilityErrorKind = "limit_exceeded"
	AvailabilityErrorSourceInvalid AvailabilityErrorKind = "source_invalid"
)

var (
	ErrAvailabilityLimitExceeded = AvailabilityError{kind: AvailabilityErrorLimitExceeded}
	ErrAvailabilitySourceInvalid = AvailabilityError{kind: AvailabilityErrorSourceInvalid}
)

// AvailabilityError intentionally exposes no source, user, or temporal metadata.
type AvailabilityError struct {
	kind AvailabilityErrorKind
}

func (err AvailabilityError) Error() string {
	switch err.kind {
	case AvailabilityErrorLimitExceeded:
		return "calendar availability limit exceeded"
	case AvailabilityErrorSourceInvalid:
		return "calendar availability source invalid"
	default:
		return "calendar availability error"
	}
}

func (err AvailabilityError) Is(target error) bool {
	switch value := target.(type) {
	case AvailabilityError:
		return err.kind == value.kind
	case *AvailabilityError:
		return value != nil && err.kind == value.kind
	default:
		return false
	}
}

// Kind returns the safe machine-readable availability failure category.
func (err AvailabilityError) Kind() AvailabilityErrorKind {
	return err.kind
}

// AvailabilitySource is the closed public attribution set for a busy interval.
type AvailabilitySource string

const (
	AvailabilitySourceEvent AvailabilitySource = "event"
	AvailabilitySourceCall  AvailabilitySource = "call"
	AvailabilitySourceLeave AvailabilitySource = "leave"
)

func (source AvailabilitySource) Validate() error {
	switch source {
	case AvailabilitySourceEvent, AvailabilitySourceCall, AvailabilitySourceLeave:
		return nil
	default:
		return ErrAvailabilitySourceInvalid
	}
}

// AvailabilityBusyInterval is an already-authorized, recurrence-expanded busy span.
// ResolveAvailabilityPolicy converts every emitted timestamp to canonical UTC.
type AvailabilityBusyInterval struct {
	From   time.Time
	To     time.Time
	Source AvailabilitySource
}

// AvailabilityFreeInterval is a private free span with no source attribution.
type AvailabilityFreeInterval struct {
	From time.Time
	To   time.Time
}

// AvailabilityUserInput contains one requested user's validated source intervals.
// Timezone must be a loadable IANA location even though this policy receives instants.
type AvailabilityUserInput struct {
	UserID   uuid.UUID
	Timezone string
	Busy     []AvailabilityBusyInterval
}

// AvailabilityRequest is a bounded private availability query.
type AvailabilityRequest struct {
	From        time.Time
	To          time.Time
	SlotMinutes int
	Users       []AvailabilityUserInput
}

// AvailabilityUserResult preserves the corresponding AvailabilityRequest.Users order.
type AvailabilityUserResult struct {
	UserID uuid.UUID
	Busy   []AvailabilityBusyInterval
	Free   []AvailabilityFreeInterval
}

// AvailabilityResult contains no title, reason, identifier, or project metadata.
type AvailabilityResult struct {
	Users []AvailabilityUserResult
}

// ResolveAvailabilityPolicy normalizes already-authorized source intervals into public
// source-attributed busy spans and slot-filtered free spans. It performs no I/O, lookup,
// recurrence expansion, or authorization decision.
func ResolveAvailabilityPolicy(request AvailabilityRequest) (AvailabilityResult, error) {
	from, to, err := validateAvailabilityRequest(request)
	if err != nil {
		return AvailabilityResult{}, err
	}

	totalBusy := 0
	seenUsers := make(map[uuid.UUID]struct{}, len(request.Users))
	result := AvailabilityResult{Users: make([]AvailabilityUserResult, 0, len(request.Users))}
	for _, user := range request.Users {
		if err := validateAvailabilityUser(user, seenUsers); err != nil {
			return AvailabilityResult{}, err
		}
		if len(user.Busy) > AvailabilityMaxBusyIntervalsUser {
			return AvailabilityResult{}, ErrAvailabilityLimitExceeded
		}
		totalBusy += len(user.Busy)
		if totalBusy > AvailabilityMaxBusyIntervalsAll {
			return AvailabilityResult{}, ErrAvailabilityLimitExceeded
		}

		busy, err := normalizeAvailabilityBusy(user.Busy, from, to)
		if err != nil {
			return AvailabilityResult{}, err
		}
		result.Users = append(result.Users, AvailabilityUserResult{
			UserID: user.UserID,
			Busy:   mergeAvailabilityBusyBySource(busy),
			Free:   availabilityFreeIntervals(busy, from, to, request.SlotMinutes),
		})
	}

	return result, nil
}

func validateAvailabilityRequest(request AvailabilityRequest) (time.Time, time.Time, error) {
	from, ok := canonicalAvailabilityTime(request.From)
	if !ok {
		return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
	}
	to, ok := canonicalAvailabilityTime(request.To)
	if !ok || !to.After(from) || to.Sub(from) > AvailabilityMaxWindow {
		return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
	}
	if request.SlotMinutes < AvailabilityMinSlotMinutes || request.SlotMinutes > AvailabilityMaxSlotMinutes {
		return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
	}
	if len(request.Users) == 0 || len(request.Users) > AvailabilityMaxUsers {
		return time.Time{}, time.Time{}, ErrAvailabilitySourceInvalid
	}
	return from, to, nil
}

func validateAvailabilityUser(user AvailabilityUserInput, seen map[uuid.UUID]struct{}) error {
	if user.UserID == uuid.Nil {
		return ErrAvailabilitySourceInvalid
	}
	if _, exists := seen[user.UserID]; exists {
		return ErrAvailabilitySourceInvalid
	}
	seen[user.UserID] = struct{}{}
	if user.Timezone == "" || user.Timezone == "Local" {
		return ErrAvailabilitySourceInvalid
	}
	location, err := time.LoadLocation(user.Timezone)
	if err != nil || location == time.Local || location.String() != user.Timezone {
		return ErrAvailabilitySourceInvalid
	}
	return nil
}

func canonicalAvailabilityTime(value time.Time) (time.Time, bool) {
	if value.IsZero() || value.Year() < 0 || value.Year() > 9999 || value.Nanosecond() != 0 {
		return time.Time{}, false
	}
	utc := value.UTC()
	encoded := utc.Format(time.RFC3339)
	parsed, err := time.Parse(time.RFC3339, encoded)
	if err != nil || parsed.Location() != time.UTC || !parsed.Equal(utc) {
		return time.Time{}, false
	}
	return utc, true
}

func normalizeAvailabilityBusy(values []AvailabilityBusyInterval, from, to time.Time) ([]AvailabilityBusyInterval, error) {
	normalized := make([]AvailabilityBusyInterval, 0, len(values))
	for _, value := range values {
		if err := value.Source.Validate(); err != nil {
			return nil, err
		}
		start, ok := canonicalAvailabilityTime(value.From)
		if !ok {
			return nil, ErrAvailabilitySourceInvalid
		}
		end, ok := canonicalAvailabilityTime(value.To)
		if !ok || !end.After(start) {
			return nil, ErrAvailabilitySourceInvalid
		}
		if !start.Before(to) || !end.After(from) {
			continue
		}
		if start.Before(from) {
			start = from
		}
		if end.After(to) {
			end = to
		}
		normalized = append(normalized, AvailabilityBusyInterval{From: start, To: end, Source: value.Source})
	}
	return normalized, nil
}

func mergeAvailabilityBusyBySource(values []AvailabilityBusyInterval) []AvailabilityBusyInterval {
	grouped := append([]AvailabilityBusyInterval(nil), values...)
	sort.Slice(grouped, func(left, right int) bool {
		if grouped[left].Source != grouped[right].Source {
			return grouped[left].Source < grouped[right].Source
		}
		if !grouped[left].From.Equal(grouped[right].From) {
			return grouped[left].From.Before(grouped[right].From)
		}
		return grouped[left].To.Before(grouped[right].To)
	})

	merged := make([]AvailabilityBusyInterval, 0, len(grouped))
	for _, current := range grouped {
		if len(merged) == 0 {
			merged = append(merged, current)
			continue
		}
		last := &merged[len(merged)-1]
		if last.Source == current.Source && !current.From.After(last.To) {
			if current.To.After(last.To) {
				last.To = current.To
			}
			continue
		}
		merged = append(merged, current)
	}

	sort.Slice(merged, func(left, right int) bool {
		if !merged[left].From.Equal(merged[right].From) {
			return merged[left].From.Before(merged[right].From)
		}
		if !merged[left].To.Equal(merged[right].To) {
			return merged[left].To.Before(merged[right].To)
		}
		return merged[left].Source < merged[right].Source
	})
	return merged
}

func availabilityFreeIntervals(values []AvailabilityBusyInterval, from, to time.Time, slotMinutes int) []AvailabilityFreeInterval {
	union := availabilityBusyUnion(values)
	minimumGap := time.Duration(slotMinutes) * time.Minute
	free := make([]AvailabilityFreeInterval, 0, len(union)+1)
	cursor := from
	for _, interval := range union {
		if cursor.Before(interval.From) && interval.From.Sub(cursor) >= minimumGap {
			free = append(free, AvailabilityFreeInterval{From: cursor, To: interval.From})
		}
		if cursor.Before(interval.To) {
			cursor = interval.To
		}
	}
	if cursor.Before(to) && to.Sub(cursor) >= minimumGap {
		free = append(free, AvailabilityFreeInterval{From: cursor, To: to})
	}
	return free
}

func availabilityBusyUnion(values []AvailabilityBusyInterval) []AvailabilityFreeInterval {
	ordered := append([]AvailabilityBusyInterval(nil), values...)
	sort.Slice(ordered, func(left, right int) bool {
		if !ordered[left].From.Equal(ordered[right].From) {
			return ordered[left].From.Before(ordered[right].From)
		}
		if !ordered[left].To.Equal(ordered[right].To) {
			return ordered[left].To.Before(ordered[right].To)
		}
		return ordered[left].Source < ordered[right].Source
	})

	union := make([]AvailabilityFreeInterval, 0, len(ordered))
	for _, current := range ordered {
		if len(union) == 0 {
			union = append(union, AvailabilityFreeInterval{From: current.From, To: current.To})
			continue
		}
		last := &union[len(union)-1]
		if !current.From.After(last.To) {
			if current.To.After(last.To) {
				last.To = current.To
			}
			continue
		}
		union = append(union, AvailabilityFreeInterval{From: current.From, To: current.To})
	}
	return union
}
