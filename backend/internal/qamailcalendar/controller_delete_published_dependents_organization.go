//go:build qa && !release

package qamailcalendar

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendaravailabilityreservation"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendarinboundobjectstate"
	"github.com/valerii/hub/backend/ent/calendarinvitedelivery"
	"github.com/valerii/hub/backend/ent/calendarmutationreceipt"
	"github.com/valerii/hub/backend/ent/calendarreplydelivery"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailboxpolicy"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailmutationreceipt"
	"github.com/valerii/hub/backend/ent/mailobject"
	"github.com/valerii/hub/backend/ent/mailsendratebucket"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
)

// inboundRequestFixturePublishedExternalOrganizationShard retains the frozen
// Organization external-closure observations without Ent rows or dynamic
// lookup state.
type inboundRequestFixturePublishedExternalOrganizationShard struct {
	observations [14]inboundRequestFixtureExternalObservation
}

func observeInboundRequestFixturePublishedExternalOrganization(
	ctx context.Context,
	tx *ent.Tx,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) (inboundRequestFixturePublishedExternalOrganizationShard, bool) {
	if ctx == nil || tx == nil || !roots.valid() || !calendar.valid() {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}

	var shard inboundRequestFixturePublishedExternalOrganizationShard

	policy01, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlotOrganizationFirst,
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation01, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlotOrganizationFirst,
		func(ctx context.Context) ([]*ent.CalendarAvailabilityReservation, error) {
			return tx.CalendarAvailabilityReservation.Query().
				Where(
					calendaravailabilityreservation.OrganizationIDEQ(roots.organizationID),
					inboundRequestFixturePublishedExternalCalendarAvailabilityReservationNoWait(),
				).
				Order(ent.Asc(calendaravailabilityreservation.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CalendarAvailabilityReservation) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy01.matches(observation01) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[0] = observation01

	policy02, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(2),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceCalendarProofEventID,
		roots,
		calendar,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation02, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(2),
		func(ctx context.Context) ([]*ent.CalendarEvent, error) {
			return tx.CalendarEvent.Query().
				Where(
					calendarevent.OrganizationIDEQ(roots.organizationID),
					inboundRequestFixturePublishedExternalCalendarEventNoWait(),
				).
				Order(ent.Asc(calendarevent.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CalendarEvent) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy02.matches(observation02) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[1] = observation02

	policy03, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(3),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceCalendarProofInboundObjectStateID,
		roots,
		calendar,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation03, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(3),
		func(ctx context.Context) ([]*ent.CalendarInboundObjectState, error) {
			return tx.CalendarInboundObjectState.Query().
				Where(
					calendarinboundobjectstate.OrganizationIDEQ(roots.organizationID),
					inboundRequestFixturePublishedExternalCalendarInboundObjectStateNoWait(),
				).
				Order(ent.Asc(calendarinboundobjectstate.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CalendarInboundObjectState) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy03.matches(observation03) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[2] = observation03

	policy04, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(4),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation04, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(4),
		func(ctx context.Context) ([]*ent.CalendarInviteDelivery, error) {
			return tx.CalendarInviteDelivery.Query().
				Where(
					calendarinvitedelivery.OrganizationIDEQ(roots.organizationID),
					inboundRequestFixturePublishedExternalCalendarInviteDeliveryNoWait(),
				).
				Order(ent.Asc(calendarinvitedelivery.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CalendarInviteDelivery) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy04.matches(observation04) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[3] = observation04

	policy05, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(5),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceCalendarProofMutationReceiptID,
		roots,
		calendar,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation05, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(5),
		func(ctx context.Context) ([]*ent.CalendarMutationReceipt, error) {
			return tx.CalendarMutationReceipt.Query().
				Where(
					calendarmutationreceipt.OrganizationIDEQ(roots.organizationID),
					inboundRequestFixturePublishedExternalCalendarMutationReceiptNoWait(),
				).
				Order(ent.Asc(calendarmutationreceipt.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CalendarMutationReceipt) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy05.matches(observation05) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[4] = observation05

	policy06, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(6),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceCalendarProofReplyDeliveryID,
		roots,
		calendar,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation06, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(6),
		func(ctx context.Context) ([]*ent.CalendarReplyDelivery, error) {
			return tx.CalendarReplyDelivery.Query().
				Where(
					calendarreplydelivery.OrganizationIDEQ(roots.organizationID),
					inboundRequestFixturePublishedExternalCalendarReplyDeliveryNoWait(),
				).
				Order(ent.Asc(calendarreplydelivery.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.CalendarReplyDelivery) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy06.matches(observation06) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[5] = observation06

	policy07, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(7),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceRootProofMailAccountID,
		roots,
		calendar,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation07, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(7),
		func(ctx context.Context) ([]*ent.MailAccount, error) {
			return tx.MailAccount.Query().
				Where(
					mailaccount.OrganizationIDEQ(roots.organizationID),
					inboundRequestFixturePublishedExternalMailAccountNoWait(),
				).
				Order(ent.Asc(mailaccount.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.MailAccount) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy07.matches(observation07) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[6] = observation07

	policy08, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(8),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceRootProofMailDomainID,
		roots,
		calendar,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation08, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(8),
		func(ctx context.Context) ([]*ent.MailDomain, error) {
			return tx.MailDomain.Query().
				Where(
					maildomain.OrganizationIDEQ(roots.organizationID),
					inboundRequestFixturePublishedExternalMailDomainNoWait(),
				).
				Order(ent.Asc(maildomain.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.MailDomain) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy08.matches(observation08) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[7] = observation08

	policy09, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(9),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation09, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(9),
		func(ctx context.Context) ([]*ent.MailMutationReceipt, error) {
			return tx.MailMutationReceipt.Query().
				Where(
					mailmutationreceipt.OrganizationIDEQ(roots.organizationID),
					inboundRequestFixturePublishedExternalMailMutationReceiptNoWait(),
				).
				Order(ent.Asc(mailmutationreceipt.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.MailMutationReceipt) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy09.matches(observation09) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[8] = observation09

	policy10, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(10),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation10, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(10),
		func(ctx context.Context) ([]*ent.MailObject, error) {
			return tx.MailObject.Query().
				Where(
					mailobject.OrganizationIDEQ(roots.organizationID),
					inboundRequestFixturePublishedExternalMailObjectNoWait(),
				).
				Order(ent.Asc(mailobject.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.MailObject) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy10.matches(observation10) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[9] = observation10

	policy11, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(11),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation11, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(11),
		func(ctx context.Context) ([]*ent.MailSendRateBucket, error) {
			return tx.MailSendRateBucket.Query().
				Where(
					mailsendratebucket.OrganizationIDEQ(roots.organizationID),
					inboundRequestFixturePublishedExternalMailSendRateBucketNoWait(),
				).
				Order(ent.Asc(mailsendratebucket.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.MailSendRateBucket) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy11.matches(observation11) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[10] = observation11

	policy12, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(12),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation12, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(12),
		func(ctx context.Context) ([]*ent.MailboxPolicy, error) {
			return tx.MailboxPolicy.Query().
				Where(
					mailboxpolicy.OrganizationIDEQ(roots.organizationID),
					inboundRequestFixturePublishedExternalMailboxPolicyNoWait(),
				).
				Order(ent.Asc(mailboxpolicy.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.MailboxPolicy) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy12.matches(observation12) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[11] = observation12

	policy13, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(13),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation13, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(13),
		func(ctx context.Context) ([]*ent.Project, error) {
			return tx.Project.Query().
				Where(
					project.HasOrganizationWith(organization.IDEQ(roots.organizationID)),
					inboundRequestFixturePublishedExternalProjectNoWait(),
				).
				Order(ent.Asc(project.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Project) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy13.matches(observation13) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[12] = observation13

	policy14, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlotOrganizationLast,
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceRootProofOwnerUserID,
		roots,
		calendar,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	observation14, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlotOrganizationLast,
		func(ctx context.Context) ([]*ent.User, error) {
			return tx.User.Query().
				Where(
					user.HasOrganizationWith(organization.IDEQ(roots.organizationID)),
					inboundRequestFixturePublishedExternalUserNoWait(),
				).
				Order(ent.Asc(user.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.User) (uuid.UUID, bool) {
			return row.ID, row != nil && row.ID != uuid.Nil
		},
	)
	if !observed || !policy14.matches(observation14) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	shard.observations[13] = observation14

	if !shard.valid(roots, calendar) {
		return inboundRequestFixturePublishedExternalOrganizationShard{}, false
	}
	return shard, true
}

func (shard inboundRequestFixturePublishedExternalOrganizationShard) valid(
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) bool {
	if !roots.valid() || !calendar.valid() {
		return false
	}

	policy01, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlotOrganizationFirst,
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid || !policy01.matches(shard.observations[0]) {
		return false
	}

	policy02, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(2),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceCalendarProofEventID,
		roots,
		calendar,
	)
	if !valid || !policy02.matches(shard.observations[1]) {
		return false
	}

	policy03, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(3),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceCalendarProofInboundObjectStateID,
		roots,
		calendar,
	)
	if !valid || !policy03.matches(shard.observations[2]) {
		return false
	}

	policy04, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(4),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid || !policy04.matches(shard.observations[3]) {
		return false
	}

	policy05, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(5),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceCalendarProofMutationReceiptID,
		roots,
		calendar,
	)
	if !valid || !policy05.matches(shard.observations[4]) {
		return false
	}

	policy06, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(6),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceCalendarProofReplyDeliveryID,
		roots,
		calendar,
	)
	if !valid || !policy06.matches(shard.observations[5]) {
		return false
	}

	policy07, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(7),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceRootProofMailAccountID,
		roots,
		calendar,
	)
	if !valid || !policy07.matches(shard.observations[6]) {
		return false
	}

	policy08, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(8),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceRootProofMailDomainID,
		roots,
		calendar,
	)
	if !valid || !policy08.matches(shard.observations[7]) {
		return false
	}

	policy09, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(9),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid || !policy09.matches(shard.observations[8]) {
		return false
	}

	policy10, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(10),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid || !policy10.matches(shard.observations[9]) {
		return false
	}

	policy11, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(11),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid || !policy11.matches(shard.observations[10]) {
		return false
	}

	policy12, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(12),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid || !policy12.matches(shard.observations[11]) {
		return false
	}

	policy13, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(13),
		inboundRequestFixtureExternalPolicyRootKindOrganization,
	)
	if !valid || !policy13.matches(shard.observations[12]) {
		return false
	}

	policy14, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlotOrganizationLast,
		inboundRequestFixtureExternalPolicyRootKindOrganization,
		inboundRequestFixtureExternalPolicySourceRootProofOwnerUserID,
		roots,
		calendar,
	)
	return valid && policy14.matches(shard.observations[13])
}

func inboundRequestFixturePublishedExternalCalendarAvailabilityReservationNoWait() predicate.CalendarAvailabilityReservation {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalCalendarEventNoWait() predicate.CalendarEvent {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalCalendarInboundObjectStateNoWait() predicate.CalendarInboundObjectState {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalCalendarInviteDeliveryNoWait() predicate.CalendarInviteDelivery {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalCalendarMutationReceiptNoWait() predicate.CalendarMutationReceipt {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalCalendarReplyDeliveryNoWait() predicate.CalendarReplyDelivery {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalMailAccountNoWait() predicate.MailAccount {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalMailDomainNoWait() predicate.MailDomain {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalMailMutationReceiptNoWait() predicate.MailMutationReceipt {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalMailObjectNoWait() predicate.MailObject {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalMailSendRateBucketNoWait() predicate.MailSendRateBucket {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalMailboxPolicyNoWait() predicate.MailboxPolicy {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalProjectNoWait() predicate.Project {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserNoWait() predicate.User {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
