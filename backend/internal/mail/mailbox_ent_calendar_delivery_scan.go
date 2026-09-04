package mail

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarinvitedelivery"
	"github.com/valerii/hub/backend/ent/calendarreplydelivery"
)

type calendarMailDeliveryScanStage string

const (
	calendarMailDeliveryScanStageSnapshotBegin        calendarMailDeliveryScanStage = "snapshot_begin"
	calendarMailDeliveryScanStageWallClock            calendarMailDeliveryScanStage = "wall_clock"
	calendarMailDeliveryScanStageInviteRetryQuery     calendarMailDeliveryScanStage = "invite_retry_query"
	calendarMailDeliveryScanStageInviteRetryCandidate calendarMailDeliveryScanStage = "invite_retry_candidate"
	calendarMailDeliveryScanStageInviteStaleQuery     calendarMailDeliveryScanStage = "invite_stale_query"
	calendarMailDeliveryScanStageInviteStaleCandidate calendarMailDeliveryScanStage = "invite_stale_candidate"
	calendarMailDeliveryScanStageReplyRetryQuery      calendarMailDeliveryScanStage = "reply_retry_query"
	calendarMailDeliveryScanStageReplyRetryCandidate  calendarMailDeliveryScanStage = "reply_retry_candidate"
	calendarMailDeliveryScanStageReplyStaleQuery      calendarMailDeliveryScanStage = "reply_stale_query"
	calendarMailDeliveryScanStageReplyStaleCandidate  calendarMailDeliveryScanStage = "reply_stale_candidate"
	calendarMailDeliveryScanStageSnapshotCommit       calendarMailDeliveryScanStage = "snapshot_commit"
	calendarMailDeliveryScanStageResultValidation     calendarMailDeliveryScanStage = "result_validation"
)

// ScanDueCalendarMailDeliveries returns the next bounded, globally ordered
// calendar delivery intents that are eligible for repository-owned claiming.
// It deliberately takes one read-only repeatable snapshot: no delivery state
// is changed and the database wall clock is the sole due-time authority.
func (repository *MailboxEntRepository) ScanDueCalendarMailDeliveries(ctx context.Context, request CalendarMailDeliveryDueScanRequest) (CalendarMailDeliveryDueScanResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !validCalendarMailDeliveryDueScanRequest(request) {
		return CalendarMailDeliveryDueScanResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}

	stage := calendarMailDeliveryScanStageSnapshotBegin
	candidates := make([]mailboxEntCalendarDueDelivery, 0, request.Limit)
	err := repository.withMailboxEntReadSnapshot(ctx, func(transaction *ent.Tx) error {
		stage = calendarMailDeliveryScanStageWallClock
		databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}

		stage = calendarMailDeliveryScanStageInviteRetryQuery
		inviteRetryRows, err := transaction.CalendarInviteDelivery.Query().
			Where(
				calendarinvitedelivery.StateIn(calendarinvitedelivery.StatePending, calendarinvitedelivery.StateUnknown),
				calendarinvitedelivery.NextAttemptAtNotNil(),
				calendarinvitedelivery.NextAttemptAtLTE(databaseWallClock),
			).
			Order(ent.Asc(calendarinvitedelivery.FieldNextAttemptAt), ent.Asc(calendarinvitedelivery.FieldID)).
			Limit(request.Limit).
			All(ctx)
		if err != nil {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		for _, row := range inviteRetryRows {
			stage = calendarMailDeliveryScanStageInviteRetryCandidate
			candidate, err := mailboxEntCalendarRetryDueCandidate(CalendarMailDeliveryKindInvite, row.ID, CalendarMailDeliveryState(row.State), row.NextAttemptAt, databaseWallClock)
			if err != nil {
				return err
			}
			candidates = append(candidates, candidate)
		}

		stage = calendarMailDeliveryScanStageInviteStaleQuery
		inviteStaleRows, err := transaction.CalendarInviteDelivery.Query().
			Where(
				calendarinvitedelivery.StateEQ(calendarinvitedelivery.StateSubmitting),
				calendarinvitedelivery.LeaseUntilNotNil(),
				calendarinvitedelivery.LeaseUntilLTE(databaseWallClock),
			).
			Order(ent.Asc(calendarinvitedelivery.FieldLeaseUntil), ent.Asc(calendarinvitedelivery.FieldID)).
			Limit(request.Limit).
			All(ctx)
		if err != nil {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		for _, row := range inviteStaleRows {
			stage = calendarMailDeliveryScanStageInviteStaleCandidate
			candidate, err := mailboxEntCalendarSubmittingDueCandidate(CalendarMailDeliveryKindInvite, row.ID, CalendarMailDeliveryState(row.State), row.LeaseUntil, databaseWallClock)
			if err != nil {
				return err
			}
			candidates = append(candidates, candidate)
		}

		stage = calendarMailDeliveryScanStageReplyRetryQuery
		replyRetryRows, err := transaction.CalendarReplyDelivery.Query().
			Where(
				calendarreplydelivery.StateIn(calendarreplydelivery.StatePending, calendarreplydelivery.StateUnknown),
				calendarreplydelivery.NextAttemptAtNotNil(),
				calendarreplydelivery.NextAttemptAtLTE(databaseWallClock),
			).
			Order(ent.Asc(calendarreplydelivery.FieldNextAttemptAt), ent.Asc(calendarreplydelivery.FieldID)).
			Limit(request.Limit).
			All(ctx)
		if err != nil {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		for _, row := range replyRetryRows {
			stage = calendarMailDeliveryScanStageReplyRetryCandidate
			candidate, err := mailboxEntCalendarRetryDueCandidate(CalendarMailDeliveryKindReply, row.ID, CalendarMailDeliveryState(row.State), row.NextAttemptAt, databaseWallClock)
			if err != nil {
				return err
			}
			candidates = append(candidates, candidate)
		}

		stage = calendarMailDeliveryScanStageReplyStaleQuery
		replyStaleRows, err := transaction.CalendarReplyDelivery.Query().
			Where(
				calendarreplydelivery.StateEQ(calendarreplydelivery.StateSubmitting),
				calendarreplydelivery.LeaseUntilNotNil(),
				calendarreplydelivery.LeaseUntilLTE(databaseWallClock),
			).
			Order(ent.Asc(calendarreplydelivery.FieldLeaseUntil), ent.Asc(calendarreplydelivery.FieldID)).
			Limit(request.Limit).
			All(ctx)
		if err != nil {
			return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
		for _, row := range replyStaleRows {
			stage = calendarMailDeliveryScanStageReplyStaleCandidate
			candidate, err := mailboxEntCalendarSubmittingDueCandidate(CalendarMailDeliveryKindReply, row.ID, CalendarMailDeliveryState(row.State), row.LeaseUntil, databaseWallClock)
			if err != nil {
				return err
			}
			candidates = append(candidates, candidate)
		}
		stage = calendarMailDeliveryScanStageSnapshotCommit
		return nil
	})
	if err != nil {
		slog.Warn("calendar mail delivery repository scan failed", "stage", string(stage))
		return CalendarMailDeliveryDueScanResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}

	sort.SliceStable(candidates, func(leftIndex, rightIndex int) bool {
		left := candidates[leftIndex]
		right := candidates[rightIndex]
		if left.dueAt.Before(right.dueAt) {
			return true
		}
		if left.dueAt.After(right.dueAt) {
			return false
		}
		leftUUID := left.deliveryID.UUID().String()
		rightUUID := right.deliveryID.UUID().String()
		if leftUUID != rightUUID {
			return leftUUID < rightUUID
		}
		return left.deliveryID.Kind() < right.deliveryID.Kind()
	})

	result := CalendarMailDeliveryDueScanResult{DeliveryIDs: make([]CalendarMailDeliveryID, 0, request.Limit)}
	seen := make(map[CalendarMailDeliveryID]struct{}, len(candidates))
	for _, candidate := range candidates {
		if _, duplicate := seen[candidate.deliveryID]; duplicate {
			continue
		}
		seen[candidate.deliveryID] = struct{}{}
		if len(result.DeliveryIDs) < request.Limit {
			result.DeliveryIDs = append(result.DeliveryIDs, candidate.deliveryID)
		}
	}
	if !validCalendarMailDeliveryDueScanResult(request, result) {
		stage = calendarMailDeliveryScanStageResultValidation
		slog.Warn("calendar mail delivery repository scan failed", "stage", string(stage))
		return CalendarMailDeliveryDueScanResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return mailboxEntCalendarDueScanResultCopy(result), nil
}

type mailboxEntCalendarDueDelivery struct {
	deliveryID CalendarMailDeliveryID
	dueAt      time.Time
}

func mailboxEntCalendarRetryDueCandidate(kind CalendarMailDeliveryKind, id uuid.UUID, state CalendarMailDeliveryState, dueAt *time.Time, databaseWallClock time.Time) (mailboxEntCalendarDueDelivery, error) {
	if state != CalendarMailDeliveryStatePending && state != CalendarMailDeliveryStateUnknown {
		return mailboxEntCalendarDueDelivery{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return mailboxEntCalendarDueCandidate(kind, id, dueAt, databaseWallClock)
}

func mailboxEntCalendarSubmittingDueCandidate(kind CalendarMailDeliveryKind, id uuid.UUID, state CalendarMailDeliveryState, dueAt *time.Time, databaseWallClock time.Time) (mailboxEntCalendarDueDelivery, error) {
	if state != CalendarMailDeliveryStateSubmitting {
		return mailboxEntCalendarDueDelivery{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return mailboxEntCalendarDueCandidate(kind, id, dueAt, databaseWallClock)
}

func mailboxEntCalendarDueCandidate(kind CalendarMailDeliveryKind, id uuid.UUID, dueAt *time.Time, databaseWallClock time.Time) (mailboxEntCalendarDueDelivery, error) {
	if dueAt == nil || !mailboxEntValidTime(*dueAt) || !mailboxEntValidTime(databaseWallClock) || dueAt.After(databaseWallClock) {
		return mailboxEntCalendarDueDelivery{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	deliveryID, err := NewCalendarMailDeliveryID(kind, id)
	if err != nil {
		return mailboxEntCalendarDueDelivery{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return mailboxEntCalendarDueDelivery{deliveryID: deliveryID, dueAt: dueAt.UTC()}, nil
}

func mailboxEntCalendarDueScanResultCopy(result CalendarMailDeliveryDueScanResult) CalendarMailDeliveryDueScanResult {
	return CalendarMailDeliveryDueScanResult{
		DeliveryIDs: append(make([]CalendarMailDeliveryID, 0, len(result.DeliveryIDs)), result.DeliveryIDs...),
	}
}
