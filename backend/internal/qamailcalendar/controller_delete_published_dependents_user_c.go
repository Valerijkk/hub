//go:build qa && !release

package qamailcalendar

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/ent/hubbase"
	"github.com/valerii/hub/backend/ent/hubrecord"
	"github.com/valerii/hub/backend/ent/invite"
	"github.com/valerii/hub/backend/ent/invoice"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/issuemessagelink"
	"github.com/valerii/hub/backend/ent/journalentry"
	"github.com/valerii/hub/backend/ent/kudo"
	"github.com/valerii/hub/backend/ent/leaverequest"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/user"
)

// inboundRequestFixturePublishedExternalUserCShard retains the frozen User C
// external-closure observations without Ent rows or dynamic lookup state.
type inboundRequestFixturePublishedExternalUserCShard struct {
	observations [14]inboundRequestFixtureExternalObservation
}

func observeInboundRequestFixturePublishedExternalUserC(
	ctx context.Context,
	tx *ent.Tx,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) (inboundRequestFixturePublishedExternalUserCShard, bool) {
	if ctx == nil || tx == nil || !roots.valid() || !calendar.valid() {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}

	var shard inboundRequestFixturePublishedExternalUserCShard

	policy42, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(42),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation42, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(42),
		func(ctx context.Context) ([]*ent.FileMetadata, error) {
			return tx.FileMetadata.Query().
				Where(
					filemetadata.HasUploaderWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserCFileMetadataNoWait(),
				).
				Order(ent.Asc(filemetadata.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.FileMetadata) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy42.matches(observation42) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[0] = observation42

	policy43, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(43),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation43, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(43),
		func(ctx context.Context) ([]*ent.HubBase, error) {
			return tx.HubBase.Query().
				Where(
					hubbase.HasOwnerWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserCHubBaseNoWait(),
				).
				Order(ent.Asc(hubbase.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.HubBase) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy43.matches(observation43) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[1] = observation43

	policy44, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(44),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation44, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(44),
		func(ctx context.Context) ([]*ent.HubRecord, error) {
			return tx.HubRecord.Query().
				Where(
					hubrecord.HasCreatedByWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserCHubRecordNoWait(),
				).
				Order(ent.Asc(hubrecord.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.HubRecord) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy44.matches(observation44) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[2] = observation44

	policy45, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(45),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation45, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(45),
		func(ctx context.Context) ([]*ent.Invite, error) {
			return tx.Invite.Query().
				Where(
					invite.HasCreatedByWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserCInviteNoWait(),
				).
				Order(ent.Asc(invite.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Invite) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy45.matches(observation45) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[3] = observation45

	policy46, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(46),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation46, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(46),
		func(ctx context.Context) ([]*ent.Invoice, error) {
			return tx.Invoice.Query().
				Where(
					invoice.HasCreatedByWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserCInvoiceNoWait(),
				).
				Order(ent.Asc(invoice.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Invoice) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy46.matches(observation46) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[4] = observation46

	policy47, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(47),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation47, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(47),
		func(ctx context.Context) ([]*ent.IssueMessageLink, error) {
			return tx.IssueMessageLink.Query().
				Where(
					issuemessagelink.CreatedByIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserCIssueMessageLinkNoWait(),
				).
				Order(ent.Asc(issuemessagelink.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.IssueMessageLink) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy47.matches(observation47) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[5] = observation47

	policy48, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(48),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation48, observed := observeInboundRequestFixturePublishedExternalUserCIssue(
		ctx,
		tx,
		inboundRequestFixtureExternalPolicySlot(48),
		issue.HasAssigneeWith(user.IDEQ(roots.ownerUserID)),
	)
	if !observed || !policy48.matches(observation48) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[6] = observation48

	policy49, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(49),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation49, observed := observeInboundRequestFixturePublishedExternalUserCIssue(
		ctx,
		tx,
		inboundRequestFixtureExternalPolicySlot(49),
		issue.HasReporterWith(user.IDEQ(roots.ownerUserID)),
	)
	if !observed || !policy49.matches(observation49) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[7] = observation49

	policy50, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(50),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation50, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(50),
		func(ctx context.Context) ([]*ent.JournalEntry, error) {
			return tx.JournalEntry.Query().
				Where(
					journalentry.HasCreatedByWith(user.IDEQ(roots.ownerUserID)),
					inboundRequestFixturePublishedExternalUserCJournalEntryNoWait(),
				).
				Order(ent.Asc(journalentry.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.JournalEntry) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy50.matches(observation50) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[8] = observation50

	policy51, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(51),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation51, observed := observeInboundRequestFixturePublishedExternalUserCKudo(
		ctx,
		tx,
		inboundRequestFixtureExternalPolicySlot(51),
		kudo.HasRecipientWith(user.IDEQ(roots.ownerUserID)),
	)
	if !observed || !policy51.matches(observation51) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[9] = observation51

	policy52, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(52),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation52, observed := observeInboundRequestFixturePublishedExternalUserCKudo(
		ctx,
		tx,
		inboundRequestFixtureExternalPolicySlot(52),
		kudo.HasSenderWith(user.IDEQ(roots.ownerUserID)),
	)
	if !observed || !policy52.matches(observation52) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[10] = observation52

	policy53, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(53),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation53, observed := observeInboundRequestFixturePublishedExternalUserCLeaveRequest(
		ctx,
		tx,
		inboundRequestFixtureExternalPolicySlot(53),
		leaverequest.HasRequesterWith(user.IDEQ(roots.ownerUserID)),
	)
	if !observed || !policy53.matches(observation53) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[11] = observation53

	policy54, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(54),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation54, observed := observeInboundRequestFixturePublishedExternalUserCLeaveRequest(
		ctx,
		tx,
		inboundRequestFixtureExternalPolicySlot(54),
		leaverequest.HasReviewerWith(user.IDEQ(roots.ownerUserID)),
	)
	if !observed || !policy54.matches(observation54) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[12] = observation54

	policy55, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(55),
		inboundRequestFixtureExternalPolicyRootKindUser,
		inboundRequestFixtureExternalPolicySourceRootProofMailAccountID,
		roots,
		calendar,
	)
	if !valid {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	observation55, observed := observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		inboundRequestFixtureExternalPolicySlot(55),
		func(ctx context.Context) ([]*ent.MailAccount, error) {
			return tx.MailAccount.Query().
				Where(
					mailaccount.UserIDEQ(roots.ownerUserID),
					inboundRequestFixturePublishedExternalUserCMailAccountNoWait(),
				).
				Order(ent.Asc(mailaccount.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.MailAccount) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
	if !observed || !policy55.matches(observation55) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	shard.observations[13] = observation55

	if !shard.valid(roots, calendar) {
		return inboundRequestFixturePublishedExternalUserCShard{}, false
	}
	return shard, true
}

func (shard inboundRequestFixturePublishedExternalUserCShard) valid(
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) bool {
	if !roots.valid() || !calendar.valid() {
		return false
	}

	policy42, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(42),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy42.matches(shard.observations[0]) {
		return false
	}

	policy43, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(43),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy43.matches(shard.observations[1]) {
		return false
	}

	policy44, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(44),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy44.matches(shard.observations[2]) {
		return false
	}

	policy45, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(45),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy45.matches(shard.observations[3]) {
		return false
	}

	policy46, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(46),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy46.matches(shard.observations[4]) {
		return false
	}

	policy47, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(47),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy47.matches(shard.observations[5]) {
		return false
	}

	policy48, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(48),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy48.matches(shard.observations[6]) {
		return false
	}

	policy49, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(49),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy49.matches(shard.observations[7]) {
		return false
	}

	policy50, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(50),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy50.matches(shard.observations[8]) {
		return false
	}

	policy51, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(51),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy51.matches(shard.observations[9]) {
		return false
	}

	policy52, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(52),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy52.matches(shard.observations[10]) {
		return false
	}

	policy53, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(53),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy53.matches(shard.observations[11]) {
		return false
	}

	policy54, valid := newInboundRequestFixtureExternalPolicyForbidden(
		inboundRequestFixtureExternalPolicySlot(54),
		inboundRequestFixtureExternalPolicyRootKindUser,
	)
	if !valid || !policy54.matches(shard.observations[12]) {
		return false
	}

	policy55, valid := newInboundRequestFixtureExternalPolicyAllowlisted(
		inboundRequestFixtureExternalPolicySlot(55),
		inboundRequestFixtureExternalPolicyRootKindUser,
		inboundRequestFixtureExternalPolicySourceRootProofMailAccountID,
		roots,
		calendar,
	)
	return valid && policy55.matches(shard.observations[13])
}

func observeInboundRequestFixturePublishedExternalUserCIssue(
	ctx context.Context,
	tx *ent.Tx,
	slot inboundRequestFixtureExternalPolicySlot,
	match predicate.Issue,
) (inboundRequestFixtureExternalObservation, bool) {
	if ctx == nil || tx == nil || match == nil {
		return inboundRequestFixtureExternalObservation{}, false
	}
	return observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		slot,
		func(ctx context.Context) ([]*ent.Issue, error) {
			return tx.Issue.Query().
				Where(
					match,
					inboundRequestFixturePublishedExternalUserCIssueNoWait(),
				).
				Order(ent.Asc(issue.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Issue) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
}

func observeInboundRequestFixturePublishedExternalUserCKudo(
	ctx context.Context,
	tx *ent.Tx,
	slot inboundRequestFixtureExternalPolicySlot,
	match predicate.Kudo,
) (inboundRequestFixtureExternalObservation, bool) {
	if ctx == nil || tx == nil || match == nil {
		return inboundRequestFixtureExternalObservation{}, false
	}
	return observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		slot,
		func(ctx context.Context) ([]*ent.Kudo, error) {
			return tx.Kudo.Query().
				Where(
					match,
					inboundRequestFixturePublishedExternalUserCKudoNoWait(),
				).
				Order(ent.Asc(kudo.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.Kudo) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
}

func observeInboundRequestFixturePublishedExternalUserCLeaveRequest(
	ctx context.Context,
	tx *ent.Tx,
	slot inboundRequestFixtureExternalPolicySlot,
	match predicate.LeaveRequest,
) (inboundRequestFixtureExternalObservation, bool) {
	if ctx == nil || tx == nil || match == nil {
		return inboundRequestFixtureExternalObservation{}, false
	}
	return observeInboundRequestFixturePublishedExternalLockedExactSet(
		ctx,
		slot,
		func(ctx context.Context) ([]*ent.LeaveRequest, error) {
			return tx.LeaveRequest.Query().
				Where(
					match,
					inboundRequestFixturePublishedExternalUserCLeaveRequestNoWait(),
				).
				Order(ent.Asc(leaverequest.FieldID)).
				Limit(2).
				All(ctx)
		},
		func(row *ent.LeaveRequest) (uuid.UUID, bool) {
			if row == nil {
				return uuid.Nil, false
			}
			return row.ID, row.ID != uuid.Nil
		},
	)
}

func inboundRequestFixturePublishedExternalUserCFileMetadataNoWait() predicate.FileMetadata {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserCHubBaseNoWait() predicate.HubBase {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserCHubRecordNoWait() predicate.HubRecord {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserCInviteNoWait() predicate.Invite {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserCInvoiceNoWait() predicate.Invoice {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserCIssueMessageLinkNoWait() predicate.IssueMessageLink {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserCIssueNoWait() predicate.Issue {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserCJournalEntryNoWait() predicate.JournalEntry {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserCKudoNoWait() predicate.Kudo {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserCLeaveRequestNoWait() predicate.LeaveRequest {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedExternalUserCMailAccountNoWait() predicate.MailAccount {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
