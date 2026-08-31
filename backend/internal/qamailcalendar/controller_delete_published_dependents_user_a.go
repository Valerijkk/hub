//go:build qa && !release

package qamailcalendar

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/activitylog"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/ent/aiusagedaily"
	"github.com/valerii/hub/backend/ent/aiuserpreference"
	"github.com/valerii/hub/backend/ent/apikey"
	"github.com/valerii/hub/backend/ent/automationrule"
	"github.com/valerii/hub/backend/ent/bookmark"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendaravailabilityreservation"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendarinboundobjectstate"
	"github.com/valerii/hub/backend/ent/calendarmutationreceipt"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/user"
)

// inboundRequestFixturePublishedExternalUserAShard retains the frozen User A
// external-closure observations without Ent rows or dynamic lookup state.
type inboundRequestFixturePublishedExternalUserAShard struct {
	observations [13]inboundRequestFixtureExternalObservation
}

func observeInboundRequestFixturePublishedExternalUserA(
	ctx context.Context,
	tx *ent.Tx,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) (inboundRequestFixturePublishedExternalUserAShard, bool) {
	if ctx == nil || tx == nil || !roots.valid() || !calendar.valid() {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}

	var shard inboundRequestFixturePublishedExternalUserAShard

	policy15, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlotUserFirst,
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation15, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlotUserFirst,
		func(ctx context.Context) ([]*ent.ActivityLog, error) {
			return tx.ActivityLog.Query().
				Where(
					activitylog.HasUserWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserAActivityLogNoWait(),
				).
				Order(ent.Asc(activitylog.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.ActivityLog) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy15.matches(observation15) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[0] = observation15

	policy16, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(16),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation16, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(16),
		func(ctx context.Context) ([]*ent.AIProvider, error) {
			return tx.AIProvider.Query().
				Where(
					aiprovider.HasOwnerWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserAAIProviderNoWait(),
				).
				Order(ent.Asc(aiprovider.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.AIProvider) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy16.matches(observation16) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[1] = observation16

	policy17, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(17),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation17, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(17),
		func(ctx context.Context) ([]*ent.AIUsageDaily, error) {
			return tx.AIUsageDaily.Query().
				Where(
					aiusagedaily.HasUserWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserAAIUsageDailyNoWait(),
				).
				Order(ent.Asc(aiusagedaily.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.AIUsageDaily) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy17.matches(observation17) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[2] = observation17

	policy18, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(18),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation18, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(18),
		func(ctx context.Context) ([]*ent.AIUserPreference, error) {
			return tx.AIUserPreference.Query().
				Where(
					aiuserpreference.HasUserWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserAAIUserPreferenceNoWait(),
				).
				Order(ent.Asc(aiuserpreference.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.AIUserPreference) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy18.matches(observation18) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[3] = observation18

	policy19, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(19),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation19, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(19),
		func(ctx context.Context) ([]*ent.APIKey, error) {
			return tx.APIKey.Query().
				Where(
					apikey.HasUserWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserAAPIKeyNoWait(),
				).
				Order(ent.Asc(apikey.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.APIKey) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy19.matches(observation19) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[4] = observation19

	policy20, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(20),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation20, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(20),
		func(ctx context.Context) ([]*ent.AutomationRule, error) {
			return tx.AutomationRule.Query().
				Where(
					automationrule.CreatedByIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserAAutomationRuleNoWait(),
				).
				Order(ent.Asc(automationrule.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.AutomationRule) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy20.matches(observation20) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[5] = observation20

	policy21, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(21),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation21, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(21),
		func(ctx context.Context) ([]*ent.Bookmark, error) {
			return tx.Bookmark.Query().
				Where(
					bookmark.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserABookmarkNoWait(),
				).
				Order(ent.Asc(bookmark.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Bookmark) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy21.matches(observation21) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[6] = observation21

	policy22, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(22),
		inboundRequestFixtureExternalPolicyRootKindUser,
		inboundRequestFixtureExternalPolicySourceCalendarProofAttendeeID,
		roots,
		calendar,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation22, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(22),
		func(ctx context.Context) ([]*ent.CalendarAttendee, error) {
			return tx.CalendarAttendee.Query().
				Where(
					calendarattendee.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserACalendarAttendeeNoWait(),
				).
				Order(ent.Asc(calendarattendee.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CalendarAttendee) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy22.matches(observation22) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[7] = observation22

	policy23, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(23),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation23, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(23),
		func(ctx context.Context) ([]*ent.CalendarAvailabilityReservation, error) {
			return tx.CalendarAvailabilityReservation.Query().
				Where(
					calendaravailabilityreservation.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserACalendarAvailabilityReservationNoWait(),
				).
				Order(ent.Asc(calendaravailabilityreservation.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CalendarAvailabilityReservation) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy23.matches(observation23) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[8] = observation23

	policy24, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(24),
		inboundRequestFixtureExternalPolicyRootKindUser,
		inboundRequestFixtureExternalPolicySourceCalendarProofEventID,
		roots,
		calendar,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation24, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(24),
		func(ctx context.Context) ([]*ent.CalendarEvent, error) {
			return tx.CalendarEvent.Query().
				Where(
					calendarevent.CalendarOwnerIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserACalendarEventCalendarOwnerNoWait(),
				).
				Order(ent.Asc(calendarevent.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CalendarEvent) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy24.matches(observation24) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[9] = observation24

	policy25, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(25),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation25, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(25),
		func(ctx context.Context) ([]*ent.CalendarEvent, error) {
			return tx.CalendarEvent.Query().
				Where(
					calendarevent.OrganizerIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserACalendarEventOrganizerNoWait(),
				).
				Order(ent.Asc(calendarevent.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CalendarEvent) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy25.matches(observation25) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[10] = observation25

	policy26, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(26),
		inboundRequestFixtureExternalPolicyRootKindUser,
		inboundRequestFixtureExternalPolicySourceCalendarProofInboundObjectStateID,
		roots,
		calendar,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation26, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(26),
		func(ctx context.Context) ([]*ent.CalendarInboundObjectState, error) {
			return tx.CalendarInboundObjectState.Query().
				Where(
					calendarinboundobjectstate.CalendarOwnerIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserACalendarInboundObjectStateNoWait(),
				).
				Order(ent.Asc(calendarinboundobjectstate.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CalendarInboundObjectState) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy26.matches(observation26) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[11] = observation26

	policy27, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(27),
		inboundRequestFixtureExternalPolicyRootKindUser,
		inboundRequestFixtureExternalPolicySourceCalendarProofMutationReceiptID,
		roots,
		calendar,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	observation27, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(27),
		func(ctx context.Context) ([]*ent.CalendarMutationReceipt, error) {
			return tx.CalendarMutationReceipt.Query().
				Where(
					calendarmutationreceipt.ActorUserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserACalendarMutationReceiptNoWait(),
				).
				Order(ent.Asc(calendarmutationreceipt.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CalendarMutationReceipt) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy27.matches(observation27) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	shard.observations[12] = observation27

	if !shard.valid(roots, calendar) {
		return inboundRequestFixturePublishedExternalUserAShard{}, false
	}
	return shard, true
}

func (shard inboundRequestFixturePublishedExternalUserAShard) valid(
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) bool {
	if !roots.valid() || !calendar.valid() {
		return false
	}

	policy15, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlotUserFirst,
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy15.matches(shard.observations[0]) {
		return false
	}

	policy16, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(16),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy16.matches(shard.observations[1]) {
		return false
	}

	policy17, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(17),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy17.matches(shard.observations[2]) {
		return false
	}

	policy18, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(18),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy18.matches(shard.observations[3]) {
		return false
	}

	policy19, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(19),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy19.matches(shard.observations[4]) {
		return false
	}

	policy20, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(20),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy20.matches(shard.observations[5]) {
		return false
	}

	policy21, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(21),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy21.matches(shard.observations[6]) {
		return false
	}

	policy22, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(22),
		inboundRequestFixtureExternalPolicyRootKindUser,
		inboundRequestFixtureExternalPolicySourceCalendarProofAttendeeID,
		roots,
		calendar,
	)
	if !valid || !policy22.matches(shard.observations[7]) {
		return false
	}

	policy23, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(23),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy23.matches(shard.observations[8]) {
		return false
	}

	policy24, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(24),
		inboundRequestFixtureExternalPolicyRootKindUser,
		inboundRequestFixtureExternalPolicySourceCalendarProofEventID,
		roots,
		calendar,
	)
	if !valid || !policy24.matches(shard.observations[9]) {
		return false
	}

	policy25, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(25),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy25.matches(shard.observations[10]) {
		return false
	}

	policy26, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(26),
		inboundRequestFixtureExternalPolicyRootKindUser,
		inboundRequestFixtureExternalPolicySourceCalendarProofInboundObjectStateID,
		roots,
		calendar,
	)
	if !valid || !policy26.matches(shard.observations[11]) {
		return false
	}

	policy27, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(27),
		inboundRequestFixtureExternalPolicyRootKindUser,
		inboundRequestFixtureExternalPolicySourceCalendarProofMutationReceiptID,
		roots,
		calendar,
	)
	return valid && policy27.matches(shard.observations[12])
}

func inboundRequestFixturePublishedExternalUserAActivityLogNoWait() predicate.ActivityLog {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserAAIProviderNoWait() predicate.AIProvider {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserAAIUsageDailyNoWait() predicate.AIUsageDaily {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserAAIUserPreferenceNoWait() predicate.AIUserPreference {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserAAPIKeyNoWait() predicate.APIKey {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserAAutomationRuleNoWait() predicate.AutomationRule {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserABookmarkNoWait() predicate.Bookmark {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserACalendarAttendeeNoWait() predicate.CalendarAttendee {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserACalendarAvailabilityReservationNoWait() predicate.CalendarAvailabilityReservation {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserACalendarEventCalendarOwnerNoWait() predicate.CalendarEvent {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserACalendarEventOrganizerNoWait() predicate.CalendarEvent {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserACalendarInboundObjectStateNoWait() predicate.CalendarInboundObjectState {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserACalendarMutationReceiptNoWait() predicate.CalendarMutationReceipt {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
