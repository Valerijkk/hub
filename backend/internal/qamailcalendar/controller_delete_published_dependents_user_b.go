//go:build qa && !release

package qamailcalendar

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/ent/callparticipant"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelinvite"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/coffeeenrollment"
	"github.com/valerii/hub/backend/ent/coffeepair"
	"github.com/valerii/hub/backend/ent/coffeeround"
	"github.com/valerii/hub/backend/ent/comment"
	"github.com/valerii/hub/backend/ent/crmcontact"
	"github.com/valerii/hub/backend/ent/crmdeal"
	"github.com/valerii/hub/backend/ent/dashboardlayout"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/user"
)

// inboundRequestFixturePublishedExternalUserBShard retains the frozen User B
// external-closure observations without Ent rows or dynamic lookup state.
type inboundRequestFixturePublishedExternalUserBShard struct {
	observations [14]inboundRequestFixtureExternalObservation
}

func observeInboundRequestFixturePublishedExternalUserB(
	ctx context.Context,
	tx *ent.Tx,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) (inboundRequestFixturePublishedExternalUserBShard, bool) {
	if ctx == nil || tx == nil || !roots.valid() || !calendar.valid() {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}

	var shard inboundRequestFixturePublishedExternalUserBShard

	policy28, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(28),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation28, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(28),
		func(ctx context.Context) ([]*ent.CallParticipant, error) {
			return tx.CallParticipant.Query().
				Where(
					callparticipant.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserBCallParticipantNoWait(),
				).
				Order(ent.Asc(callparticipant.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CallParticipant) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy28.matches(observation28) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[0] = observation28

	policy29, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(29),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation29, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(29),
		func(ctx context.Context) ([]*ent.Call, error) {
			return tx.Call.Query().
				Where(
					call.CreatedBy(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserBCallNoWait(),
				).
				Order(ent.Asc(call.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Call) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy29.matches(observation29) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[1] = observation29

	policy30, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(30),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation30, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(30),
		func(ctx context.Context) ([]*ent.ChannelInvite, error) {
			return tx.ChannelInvite.Query().
				Where(
					channelinvite.CreatorIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserBChannelInviteNoWait(),
				).
				Order(ent.Asc(channelinvite.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.ChannelInvite) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy30.matches(observation30) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[2] = observation30

	policy31, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(31),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation31, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(31),
		func(ctx context.Context) ([]*ent.ChannelMember, error) {
			return tx.ChannelMember.Query().
				Where(
					channelmember.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserBChannelMemberNoWait(),
				).
				Order(ent.Asc(channelmember.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.ChannelMember) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy31.matches(observation31) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[3] = observation31

	policy32, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(32),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation32, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(32),
		func(ctx context.Context) ([]*ent.Channel, error) {
			return tx.Channel.Query().
				Where(
					channel.CreatedBy(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserBChannelNoWait(),
				).
				Order(ent.Asc(channel.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Channel) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy32.matches(observation32) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[4] = observation32

	policy33, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(33),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation33, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(33),
		func(ctx context.Context) ([]*ent.CoffeeEnrollment, error) {
			return tx.CoffeeEnrollment.Query().
				Where(
					coffeeenrollment.HasUserWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserBCoffeeEnrollmentNoWait(),
				).
				Order(ent.Asc(coffeeenrollment.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CoffeeEnrollment) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy33.matches(observation33) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[5] = observation33

	policy34, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(34),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation34, observed := observeInboundRequestFixturePublishedExternalUserBCoffeePair(
		ctx,
		tx,
		inboundRequestFixtureExternalPolicySlot(34),
		coffeepair.HasParticipantAWith(user.IDEQ(roots.ownerUserID)),
	)
	if !observed || !policy34.matches(observation34) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[6] = observation34

	policy35, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(35),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation35, observed := observeInboundRequestFixturePublishedExternalUserBCoffeePair(
		ctx,
		tx,
		inboundRequestFixtureExternalPolicySlot(35),
		coffeepair.HasParticipantBWith(user.IDEQ(roots.ownerUserID)),
	)
	if !observed || !policy35.matches(observation35) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[7] = observation35

	policy36, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(36),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation36, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(36),
		func(ctx context.Context) ([]*ent.CoffeeRound, error) {
			return tx.CoffeeRound.Query().
				Where(
					coffeeround.HasTriggeredByWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserBCoffeeRoundNoWait(),
				).
				Order(ent.Asc(coffeeround.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CoffeeRound) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy36.matches(observation36) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[8] = observation36

	policy37, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(37),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation37, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(37),
		func(ctx context.Context) ([]*ent.Comment, error) {
			return tx.Comment.Query().
				Where(
					comment.HasAuthorWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserBCommentNoWait(),
				).
				Order(ent.Asc(comment.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Comment) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy37.matches(observation37) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[9] = observation37

	policy38, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(38),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation38, observed := observeInboundRequestFixturePublishedExternalUserBCrmContact(
		ctx,
		tx,
		inboundRequestFixtureExternalPolicySlot(38),
		crmcontact.HasLinkedUserWith(user.IDEQ(roots.ownerUserID)),
	)
	if !observed || !policy38.matches(observation38) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[10] = observation38

	policy39, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(39),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation39, observed := observeInboundRequestFixturePublishedExternalUserBCrmContact(
		ctx,
		tx,
		inboundRequestFixtureExternalPolicySlot(39),
		crmcontact.HasCreatedByWith(user.IDEQ(roots.ownerUserID)),
	)
	if !observed || !policy39.matches(observation39) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[11] = observation39

	policy40, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(40),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation40, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(40),
		func(ctx context.Context) ([]*ent.CrmDeal, error) {
			return tx.CrmDeal.Query().
				Where(
					crmdeal.HasOwnerWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserBCrmDealNoWait(),
				).
				Order(ent.Asc(crmdeal.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CrmDeal) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy40.matches(observation40) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[12] = observation40

	policy41, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(41),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	observation41, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(41),
		func(ctx context.Context) ([]*ent.DashboardLayout, error) {
			return tx.DashboardLayout.Query().
				Where(
					dashboardlayout.HasUserWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserBDashboardLayoutNoWait(),
				).
				Order(ent.Asc(dashboardlayout.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.DashboardLayout) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy41.matches(observation41) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	shard.observations[13] = observation41

	if !shard.valid(roots, calendar) {
		return inboundRequestFixturePublishedExternalUserBShard{}, false
	}
	return shard, true
}

func (shard inboundRequestFixturePublishedExternalUserBShard) valid(
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) bool {
	if !roots.valid() || !calendar.valid() {
		return false
	}

	policy28, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(28),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy28.matches(shard.observations[0]) {
		return false
	}

	policy29, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(29),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy29.matches(shard.observations[1]) {
		return false
	}

	policy30, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(30),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy30.matches(shard.observations[2]) {
		return false
	}

	policy31, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(31),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy31.matches(shard.observations[3]) {
		return false
	}

	policy32, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(32),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy32.matches(shard.observations[4]) {
		return false
	}

	policy33, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(33),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy33.matches(shard.observations[5]) {
		return false
	}

	policy34, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(34),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy34.matches(shard.observations[6]) {
		return false
	}

	policy35, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(35),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy35.matches(shard.observations[7]) {
		return false
	}

	policy36, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(36),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy36.matches(shard.observations[8]) {
		return false
	}

	policy37, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(37),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy37.matches(shard.observations[9]) {
		return false
	}

	policy38, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(38),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy38.matches(shard.observations[10]) {
		return false
	}

	policy39, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(39),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy39.matches(shard.observations[11]) {
		return false
	}

	policy40, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(40),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy40.matches(shard.observations[12]) {
		return false
	}

	policy41, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(41),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	return valid && policy41.matches(shard.observations[13])
}

func observeInboundRequestFixturePublishedExternalUserBCoffeePair(
	ctx context.Context,
	tx *ent.Tx,
	slot inboundRequestFixtureExternalPolicySlot,
	match predicate.CoffeePair,
) (inboundRequestFixtureExternalObservation, bool) {
	if ctx == nil || tx == nil || match == nil {
		return inboundRequestFixtureExternalObservation{}, false
	}
	return observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		slot,
		func(ctx context.Context) ([]*ent.CoffeePair, error) {
			return tx.CoffeePair.Query().
				Where(
					match,
					inboundRequestFixturePublishedExternalUserBCoffeePairNoWait(),
				).
				Order(ent.Asc(coffeepair.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CoffeePair) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
}

func observeInboundRequestFixturePublishedExternalUserBCrmContact(
	ctx context.Context,
	tx *ent.Tx,
	slot inboundRequestFixtureExternalPolicySlot,
	match predicate.CrmContact,
) (inboundRequestFixtureExternalObservation, bool) {
	if ctx == nil || tx == nil || match == nil {
		return inboundRequestFixtureExternalObservation{}, false
	}
	return observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		slot,
		func(ctx context.Context) ([]*ent.CrmContact, error) {
			return tx.CrmContact.Query().
				Where(
					match,
					inboundRequestFixturePublishedExternalUserBCrmContactNoWait(),
				).
				Order(ent.Asc(crmcontact.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CrmContact) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
}

func inboundRequestFixturePublishedExternalUserBCallParticipantNoWait() predicate.CallParticipant {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserBCallNoWait() predicate.Call {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserBChannelInviteNoWait() predicate.ChannelInvite {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserBChannelMemberNoWait() predicate.ChannelMember {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserBChannelNoWait() predicate.Channel {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserBCoffeeEnrollmentNoWait() predicate.CoffeeEnrollment {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserBCoffeePairNoWait() predicate.CoffeePair {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserBCoffeeRoundNoWait() predicate.CoffeeRound {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserBCommentNoWait() predicate.Comment {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserBCrmContactNoWait() predicate.CrmContact {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserBCrmDealNoWait() predicate.CrmDeal {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserBDashboardLayoutNoWait() predicate.DashboardLayout {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
