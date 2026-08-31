//go:build qa && !release

package qamailcalendar

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/recentcontact"
	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/ent/savedfilter"
	"github.com/valerii/hub/backend/ent/scheduledcall"
	"github.com/valerii/hub/backend/ent/session"
	"github.com/valerii/hub/backend/ent/teammember"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/testcaseversion"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/ent/timelog"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/userblock"
)

// inboundRequestFixturePublishedExternalUserEShard retains the frozen User E
// external-closure observations without Ent rows or dynamic lookup state.
type inboundRequestFixturePublishedExternalUserEShard struct {
	observations [13]inboundRequestFixtureExternalObservation
}

func observeInboundRequestFixturePublishedExternalUserE(
	ctx context.Context,
	tx *ent.Tx,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) (inboundRequestFixturePublishedExternalUserEShard, bool) {
	if ctx == nil || tx == nil || !roots.valid() || !calendar.valid() {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}

	var shard inboundRequestFixturePublishedExternalUserEShard

	policy70, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(70),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation70, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(70),
		func(ctx context.Context) ([]*ent.RecentContact, error) {
			return tx.RecentContact.Query().
				Where(
					recentcontact.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserERecentContactNoWait(),
				).
				Order(ent.Asc(recentcontact.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.RecentContact) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy70.matches(observation70) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[0] = observation70

	policy71, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(71),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation71, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(71),
		func(ctx context.Context) ([]*ent.Recording, error) {
			return tx.Recording.Query().
				Where(
					recording.StartedByEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserERecordingNoWait(),
				).
				Order(ent.Asc(recording.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Recording) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy71.matches(observation71) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[1] = observation71

	policy72, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(72),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation72, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(72),
		func(ctx context.Context) ([]*ent.SavedFilter, error) {
			return tx.SavedFilter.Query().
				Where(
					savedfilter.OwnerIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserESavedFilterNoWait(),
				).
				Order(ent.Asc(savedfilter.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.SavedFilter) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy72.matches(observation72) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[2] = observation72

	policy73, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(73),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation73, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(73),
		func(ctx context.Context) ([]*ent.ScheduledCall, error) {
			return tx.ScheduledCall.Query().
				Where(
					scheduledcall.CreatedByEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserEScheduledCallNoWait(),
				).
				Order(ent.Asc(scheduledcall.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.ScheduledCall) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy73.matches(observation73) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[3] = observation73

	policy74, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(74),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation74, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(74),
		func(ctx context.Context) ([]*ent.Session, error) {
			return tx.Session.Query().
				Where(
					session.HasUserWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserESessionNoWait(),
				).
				Order(ent.Asc(session.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Session) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy74.matches(observation74) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[4] = observation74

	policy75, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(75),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation75, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(75),
		func(ctx context.Context) ([]*ent.TeamMember, error) {
			return tx.TeamMember.Query().
				Where(
					teammember.HasUserWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserETeamMemberNoWait(),
				).
				Order(ent.Asc(teammember.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.TeamMember) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy75.matches(observation75) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[5] = observation75

	policy76, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(76),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation76, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(76),
		func(ctx context.Context) ([]*ent.TestCaseVersion, error) {
			return tx.TestCaseVersion.Query().
				Where(
					testcaseversion.HasCreatedByWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserETestCaseVersionNoWait(),
				).
				Order(ent.Asc(testcaseversion.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.TestCaseVersion) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy76.matches(observation76) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[6] = observation76

	policy77, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(77),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation77, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(77),
		func(ctx context.Context) ([]*ent.TestCase, error) {
			return tx.TestCase.Query().
				Where(
					testcase.HasCreatedByWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserETestCaseNoWait(),
				).
				Order(ent.Asc(testcase.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.TestCase) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy77.matches(observation77) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[7] = observation77

	policy78, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(78),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation78, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(78),
		func(ctx context.Context) ([]*ent.TestResult, error) {
			return tx.TestResult.Query().
				Where(
					testresult.HasExecutedByWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserETestResultNoWait(),
				).
				Order(ent.Asc(testresult.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.TestResult) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy78.matches(observation78) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[8] = observation78

	policy79, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(79),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation79, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(79),
		func(ctx context.Context) ([]*ent.TestRun, error) {
			return tx.TestRun.Query().
				Where(
					testrun.HasCreatedByWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserETestRunNoWait(),
				).
				Order(ent.Asc(testrun.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.TestRun) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy79.matches(observation79) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[9] = observation79

	policy80, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(80),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation80, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(80),
		func(ctx context.Context) ([]*ent.TimeLog, error) {
			return tx.TimeLog.Query().
				Where(
					timelog.HasUserWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserETimeLogNoWait(),
				).
				Order(ent.Asc(timelog.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.TimeLog) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy80.matches(observation80) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[10] = observation80

	policy81, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(81),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation81, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(81),
		func(ctx context.Context) ([]*ent.UserBlock, error) {
			return tx.UserBlock.Query().
				Where(
					userblock.BlockedIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserEUserBlockBlockedNoWait(),
				).
				Order(ent.Asc(userblock.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.UserBlock) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy81.matches(observation81) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[11] = observation81

	policy82, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(82),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	observation82, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(82),
		func(ctx context.Context) ([]*ent.UserBlock, error) {
			return tx.UserBlock.Query().
				Where(
					userblock.BlockerIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserEUserBlockBlockerNoWait(),
				).
				Order(ent.Asc(userblock.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.UserBlock) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy82.matches(observation82) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	shard.observations[12] = observation82

	if !shard.valid(roots, calendar) {
		return inboundRequestFixturePublishedExternalUserEShard{}, false
	}
	return shard, true
}

func (shard inboundRequestFixturePublishedExternalUserEShard) valid(
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) bool {
	if !roots.valid() || !calendar.valid() {
		return false
	}

	policy70, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(70),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy70.matches(shard.observations[0]) {
		return false
	}

	policy71, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(71),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy71.matches(shard.observations[1]) {
		return false
	}

	policy72, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(72),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy72.matches(shard.observations[2]) {
		return false
	}

	policy73, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(73),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy73.matches(shard.observations[3]) {
		return false
	}

	policy74, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(74),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy74.matches(shard.observations[4]) {
		return false
	}

	policy75, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(75),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy75.matches(shard.observations[5]) {
		return false
	}

	policy76, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(76),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy76.matches(shard.observations[6]) {
		return false
	}

	policy77, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(77),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy77.matches(shard.observations[7]) {
		return false
	}

	policy78, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(78),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy78.matches(shard.observations[8]) {
		return false
	}

	policy79, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(79),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy79.matches(shard.observations[9]) {
		return false
	}

	policy80, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(80),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy80.matches(shard.observations[10]) {
		return false
	}

	policy81, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(81),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy81.matches(shard.observations[11]) {
		return false
	}

	policy82, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(82),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	return valid && policy82.matches(shard.observations[12])
}

func inboundRequestFixturePublishedExternalUserERecentContactNoWait() predicate.RecentContact {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserERecordingNoWait() predicate.Recording {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserESavedFilterNoWait() predicate.SavedFilter {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserEScheduledCallNoWait() predicate.ScheduledCall {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserESessionNoWait() predicate.Session {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserETeamMemberNoWait() predicate.TeamMember {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserETestCaseVersionNoWait() predicate.TestCaseVersion {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserETestCaseNoWait() predicate.TestCase {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserETestResultNoWait() predicate.TestResult {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserETestRunNoWait() predicate.TestRun {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserETimeLogNoWait() predicate.TimeLog {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserEUserBlockBlockedNoWait() predicate.UserBlock {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserEUserBlockBlockerNoWait() predicate.UserBlock {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
