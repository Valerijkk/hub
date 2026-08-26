package schema

import (
	"errors"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

const (
	calendarAvailabilityReservationOccurrenceKeyMaxLen = 35

	// calendarAvailabilityReservationActiveWindowExclusionConstraint is reserved
	// for the dedicated PostgreSQL Atlas delta. Ent cannot faithfully model the
	// required partial EXCLUDE USING gist constraint. That delta must create
	// this exact constraint over organization_id, user_id and
	// tstzrange(starts_at, ends_at, '[)') for active reservations.
	calendarAvailabilityReservationActiveWindowExclusionConstraint = "calendar_availability_reservations_active_window_excl"
)

// CalendarAvailabilityReservation holds a per-user, per-occurrence busy-window
// reservation. Meeting lifecycle code releases a reservation instead of
// mutating its scheduling identity, which makes retry and conflict handling
// auditable.
type CalendarAvailabilityReservation struct {
	ent.Schema
}

// Fields describes the immutable reservation identity, its time interval and
// the minimal lifecycle state needed to release a previously held slot.
func (CalendarAvailabilityReservation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("organization_id", uuid.UUID{}).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Immutable(),
		field.UUID("calendar_event_id", uuid.UUID{}).
			Immutable(),
		field.String("occurrence_key").
			NotEmpty().
			MaxLen(calendarAvailabilityReservationOccurrenceKeyMaxLen).
			Immutable().
			Validate(validateCalendarAvailabilityReservationOccurrenceKey),
		field.Time("starts_at").
			Immutable(),
		field.Time("ends_at").
			Immutable(),
		field.Enum("state").
			Values("active", "released").
			Default("active"),
		field.Time("released_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Mixin adds the common optimistic-lock version used by mutable reservation
// lifecycle transitions.
func (CalendarAvailabilityReservation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

// Annotations defines checks that Ent can faithfully project. The active-window
// overlap exclusion is intentionally not represented here; see
// calendarAvailabilityReservationActiveWindowExclusionConstraint.
func (CalendarAvailabilityReservation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"calendar_availability_reservations_interval": "ends_at > starts_at",
			"calendar_availability_reservations_state_release": "(state = 'active' AND released_at IS NULL) OR (state = 'released' AND released_at IS NOT NULL)",
			"calendar_availability_reservations_occurrence_key_utc": "char_length(occurrence_key) BETWEEN 20 AND 35 AND occurrence_key ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\\.[0-9]{1,9})?Z$'",
		}),
	}
}

// Edges binds a reservation to its tenant, reserved colleague and canonical
// event. Tenant and user deletion are deliberately restricted; a hard event
// deletion removes only the dependent reservation.
func (CalendarAvailabilityReservation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organization", Organization.Type).
			Field("organization_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("calendar_event", CalendarEvent.Type).
			Field("calendar_event_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes supports idempotent reservation creation and the bounded free/busy
// lookup path. The active overlap exclusion remains a later Atlas concern.
func (CalendarAvailabilityReservation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("calendar_event_id", "user_id", "occurrence_key").
			Unique(),
		index.Fields("organization_id", "user_id", "state", "starts_at"),
	}
}

func validateCalendarAvailabilityReservationOccurrenceKey(value string) error {
	occurrence, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || occurrence.Location() != time.UTC || occurrence.Format(time.RFC3339Nano) != value {
		return errors.New("occurrence key must be a canonical RFC3339Nano UTC instant")
	}
	return nil
}
