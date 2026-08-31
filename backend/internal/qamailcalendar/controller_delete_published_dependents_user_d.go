//go:build qa && !release

package qamailcalendar

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailobject"
	"github.com/valerii/hub/backend/ent/mailsendratebucket"
	"github.com/valerii/hub/backend/ent/marketplacelisting"
	"github.com/valerii/hub/backend/ent/notification"
	"github.com/valerii/hub/backend/ent/notificationpreference"
	"github.com/valerii/hub/backend/ent/newspollvote"
	"github.com/valerii/hub/backend/ent/newspost"
	"github.com/valerii/hub/backend/ent/poll"
	"github.com/valerii/hub/backend/ent/pollvote"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/projecttemplate"
	"github.com/valerii/hub/backend/ent/pushsubscription"
	"github.com/valerii/hub/backend/ent/reaction"
	"github.com/valerii/hub/backend/ent/user"
)

// inboundRequestFixturePublishedExternalUserDShard retains the frozen User D
// external-closure observations without Ent rows or dynamic lookup state.
type inboundRequestFixturePublishedExternalUserDShard struct {
	observations [14]inboundRequestFixtureExternalObservation
}

func observeInboundRequestFixturePublishedExternalUserD(
	ctx context.Context,
	tx *ent.Tx,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) (inboundRequestFixturePublishedExternalUserDShard, bool) {
	if ctx == nil || tx == nil || !roots.valid() || !calendar.valid() {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}

	var shard inboundRequestFixturePublishedExternalUserDShard

	policy56, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(56),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation56, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(56),
		func(ctx context.Context) ([]*ent.MailObject, error) {
			return tx.MailObject.Query().
				Where(
					mailobject.UploaderUserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserDMailObjectNoWait(),
				).
				Order(ent.Asc(mailobject.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.MailObject) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy56.matches(observation56) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[0] = observation56

	policy57, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(57),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation57, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(57),
		func(ctx context.Context) ([]*ent.MailSendRateBucket, error) {
			return tx.MailSendRateBucket.Query().
				Where(
					mailsendratebucket.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserDMailSendRateBucketNoWait(),
				).
				Order(ent.Asc(mailsendratebucket.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.MailSendRateBucket) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy57.matches(observation57) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[1] = observation57

	policy58, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(58),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation58, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(58),
		func(ctx context.Context) ([]*ent.MarketplaceListing, error) {
			return tx.MarketplaceListing.Query().
				Where(
					marketplacelisting.HasAuthorWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserDMarketplaceListingNoWait(),
				).
				Order(ent.Asc(marketplacelisting.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.MarketplaceListing) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy58.matches(observation58) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[2] = observation58

	policy59, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(59),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation59, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(59),
		func(ctx context.Context) ([]*ent.MarketplaceListing, error) {
			return tx.MarketplaceListing.Query().
				Where(
					marketplacelisting.HasReviewerWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserDMarketplaceListingNoWait(),
				).
				Order(ent.Asc(marketplacelisting.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.MarketplaceListing) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy59.matches(observation59) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[3] = observation59

	policy60, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(60),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation60, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(60),
		func(ctx context.Context) ([]*ent.NewsPollVote, error) {
			return tx.NewsPollVote.Query().
				Where(
					newspollvote.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserDNewsPollVoteNoWait(),
				).
				Order(ent.Asc(newspollvote.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.NewsPollVote) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy60.matches(observation60) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[4] = observation60

	policy61, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(61),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation61, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(61),
		func(ctx context.Context) ([]*ent.NewsPost, error) {
			return tx.NewsPost.Query().
				Where(
					newspost.HasAuthorWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserDNewsPostNoWait(),
				).
				Order(ent.Asc(newspost.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.NewsPost) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy61.matches(observation61) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[5] = observation61

	policy62, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(62),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation62, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(62),
		func(ctx context.Context) ([]*ent.NotificationPreference, error) {
			return tx.NotificationPreference.Query().
				Where(
					notificationpreference.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserDNotificationPreferenceNoWait(),
				).
				Order(ent.Asc(notificationpreference.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.NotificationPreference) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy62.matches(observation62) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[6] = observation62

	policy63, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(63),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation63, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(63),
		func(ctx context.Context) ([]*ent.Notification, error) {
			return tx.Notification.Query().
				Where(
					notification.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserDNotificationNoWait(),
				).
				Order(ent.Asc(notification.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Notification) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy63.matches(observation63) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[7] = observation63

	policy64, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(64),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation64, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(64),
		func(ctx context.Context) ([]*ent.PollVote, error) {
			return tx.PollVote.Query().
				Where(
					pollvote.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserDPollVoteNoWait(),
				).
				Order(ent.Asc(pollvote.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.PollVote) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy64.matches(observation64) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[8] = observation64

	policy65, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(65),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation65, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(65),
		func(ctx context.Context) ([]*ent.Poll, error) {
			return tx.Poll.Query().
				Where(
					poll.CreatorIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserDPollNoWait(),
				).
				Order(ent.Asc(poll.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Poll) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy65.matches(observation65) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[9] = observation65

	policy66, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(66),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation66, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(66),
		func(ctx context.Context) ([]*ent.ProjectMember, error) {
			return tx.ProjectMember.Query().
				Where(
					projectmember.HasUserWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserDProjectMemberNoWait(),
				).
				Order(ent.Asc(projectmember.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.ProjectMember) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy66.matches(observation66) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[10] = observation66

	policy67, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(67),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation67, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(67),
		func(ctx context.Context) ([]*ent.ProjectTemplate, error) {
			return tx.ProjectTemplate.Query().
				Where(
					projecttemplate.HasCreatedByWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserDProjectTemplateNoWait(),
				).
				Order(ent.Asc(projecttemplate.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.ProjectTemplate) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy67.matches(observation67) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[11] = observation67

	policy68, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(68),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation68, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(68),
		func(ctx context.Context) ([]*ent.PushSubscription, error) {
			return tx.PushSubscription.Query().
				Where(
					pushsubscription.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserDPushSubscriptionNoWait(),
				).
				Order(ent.Asc(pushsubscription.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.PushSubscription) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy68.matches(observation68) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[12] = observation68

	policy69, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(69),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	observation69, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(69),
		func(ctx context.Context) ([]*ent.Reaction, error) {
			return tx.Reaction.Query().
				Where(
					reaction.HasUserWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserDReactionNoWait(),
				).
				Order(ent.Asc(reaction.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Reaction) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy69.matches(observation69) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	shard.observations[13] = observation69

	if !shard.valid(roots, calendar) {
		return inboundRequestFixturePublishedExternalUserDShard{}, false
	}
	return shard, true
}

func (shard inboundRequestFixturePublishedExternalUserDShard) valid(
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) bool {
	if !roots.valid() || !calendar.valid() {
		return false
	}

	policy56, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(56),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy56.matches(shard.observations[0]) {
		return false
	}

	policy57, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(57),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy57.matches(shard.observations[1]) {
		return false
	}

	policy58, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(58),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy58.matches(shard.observations[2]) {
		return false
	}

	policy59, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(59),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy59.matches(shard.observations[3]) {
		return false
	}

	policy60, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(60),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy60.matches(shard.observations[4]) {
		return false
	}

	policy61, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(61),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy61.matches(shard.observations[5]) {
		return false
	}

	policy62, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(62),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy62.matches(shard.observations[6]) {
		return false
	}

	policy63, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(63),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy63.matches(shard.observations[7]) {
		return false
	}

	policy64, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(64),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy64.matches(shard.observations[8]) {
		return false
	}

	policy65, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(65),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy65.matches(shard.observations[9]) {
		return false
	}

	policy66, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(66),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy66.matches(shard.observations[10]) {
		return false
	}

	policy67, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(67),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy67.matches(shard.observations[11]) {
		return false
	}

	policy68, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(68),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy68.matches(shard.observations[12]) {
		return false
	}

	policy69, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(69),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	return valid && policy69.matches(shard.observations[13])
}

func inboundRequestFixturePublishedExternalUserDMailObjectNoWait() predicate.MailObject {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserDMailSendRateBucketNoWait() predicate.MailSendRateBucket {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserDMarketplaceListingNoWait() predicate.MarketplaceListing {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserDNewsPollVoteNoWait() predicate.NewsPollVote {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserDNewsPostNoWait() predicate.NewsPost {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserDNotificationPreferenceNoWait() predicate.NotificationPreference {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserDNotificationNoWait() predicate.Notification {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserDPollVoteNoWait() predicate.PollVote {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserDPollNoWait() predicate.Poll {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserDProjectMemberNoWait() predicate.ProjectMember {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserDProjectTemplateNoWait() predicate.ProjectTemplate {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserDPushSubscriptionNoWait() predicate.PushSubscription {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserDReactionNoWait() predicate.Reaction {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
