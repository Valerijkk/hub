package mail

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/platform/events"
)

type MailboxSyncServiceConfig struct {
	Now           func() time.Time
	NewLeaseID    func() uuid.UUID
	NewEventID    func() uuid.UUID
	Jitter        func(time.Duration) time.Duration
	LeaseDuration time.Duration
	PollInterval  time.Duration
	RetryInterval time.Duration
	PageSize      int
}

type MailboxSyncStepOutcome uint8

const (
	MailboxSyncStepIdle MailboxSyncStepOutcome = iota + 1
	MailboxSyncStepApplied
	MailboxSyncStepRebuildScheduled
	MailboxSyncStepRetryScheduled
	MailboxSyncStepLeaseLost
)

type MailboxSyncStepResult struct {
	Outcome       MailboxSyncStepOutcome
	NextAttemptAt time.Time
}

type MailboxSyncServiceErrorKind string

const (
	MailboxSyncServiceErrorInvalid     MailboxSyncServiceErrorKind = "invalid"
	MailboxSyncServiceErrorTooLarge    MailboxSyncServiceErrorKind = "too_large"
	MailboxSyncServiceErrorUnavailable MailboxSyncServiceErrorKind = "unavailable"
)

var (
	ErrMailboxSyncServiceInvalid     = errors.New("mailbox sync service invalid")
	ErrMailboxSyncServiceTooLarge    = errors.New("mailbox sync service exceeds bounds")
	ErrMailboxSyncServiceUnavailable = errors.New("mailbox sync service unavailable")
)

type MailboxSyncServiceError struct {
	Kind MailboxSyncServiceErrorKind
}

func (errorValue *MailboxSyncServiceError) Error() string {
	if errorValue == nil {
		return "mailbox sync service error"
	}
	switch errorValue.Kind {
	case MailboxSyncServiceErrorInvalid:
		return ErrMailboxSyncServiceInvalid.Error()
	case MailboxSyncServiceErrorTooLarge:
		return ErrMailboxSyncServiceTooLarge.Error()
	case MailboxSyncServiceErrorUnavailable:
		return ErrMailboxSyncServiceUnavailable.Error()
	default:
		return "mailbox sync service error"
	}
}

func (errorValue *MailboxSyncServiceError) Unwrap() error {
	if errorValue == nil {
		return nil
	}
	switch errorValue.Kind {
	case MailboxSyncServiceErrorInvalid:
		return ErrMailboxSyncServiceInvalid
	case MailboxSyncServiceErrorTooLarge:
		return ErrMailboxSyncServiceTooLarge
	case MailboxSyncServiceErrorUnavailable:
		return ErrMailboxSyncServiceUnavailable
	default:
		return nil
	}
}

// MailboxSyncEngine reads one bounded account folder snapshot and projection page.
type MailboxSyncEngine interface {
	ListFolders(context.Context, string) ([]EngineFolder, error)
	Changes(context.Context, EngineChangeRequest) (EngineChangePage, error)
	EnumerateProjection(context.Context, EngineEnumerationRequest) (EngineEnumerationPage, error)
}

var _ MailboxSyncEngine = (*StalwartEngine)(nil)

type MailboxSyncService struct {
	engine     MailboxSyncEngine
	repository MailboxSyncRepository
	config     MailboxSyncServiceConfig
}

func NewMailboxSyncService(engine MailboxSyncEngine, repository MailboxSyncRepository, config MailboxSyncServiceConfig) (*MailboxSyncService, error) {
	if mailboxSyncNilDependency(engine) || mailboxSyncNilDependency(repository) || !validMailboxSyncServiceConfig(config) {
		return nil, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid)
	}
	return &MailboxSyncService{engine: engine, repository: repository, config: config}, nil
}

func (service *MailboxSyncService) Step(ctx context.Context, accountID uuid.UUID) (MailboxSyncStepResult, error) {
	if ctx == nil || service == nil || mailboxSyncNilDependency(service.engine) || mailboxSyncNilDependency(service.repository) || !validMailboxSyncServiceConfig(service.config) || accountID == uuid.Nil {
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return MailboxSyncStepResult{}, err
	}
	now := service.config.Now()
	if !syncValidTime(now) {
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return MailboxSyncStepResult{}, err
	}
	leaseID := service.config.NewLeaseID()
	if leaseID == uuid.Nil {
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return MailboxSyncStepResult{}, err
	}
	leaseUntil, ok := mailboxSyncCheckedTimeAdd(now, service.config.LeaseDuration)
	if !ok {
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid)
	}

	claim, err := service.repository.ClaimMailboxSyncStep(ctx, MailboxSyncClaimRequest{AccountID: accountID, LeaseID: leaseID, Now: now, LeaseUntil: leaseUntil})
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil { return MailboxSyncStepResult{}, contextErr }
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorUnavailable)
	}
	if err := ctx.Err(); err != nil { return MailboxSyncStepResult{}, err }
	if !claim.Claimed {
		if !validMailboxSyncIdleClaim(claim) {
			return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid)
		}
		return MailboxSyncStepResult{Outcome: MailboxSyncStepIdle, NextAttemptAt: claim.NextAttemptAt}, nil
	}
	if !validMailboxSyncClaim(accountID, leaseID, claim) {
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid)
	}

	folders, err := service.engine.ListFolders(ctx, claim.EngineAccountID)
	if err != nil {
		return service.handleFolderError(ctx, claim.Fence, now, err)
	}
	if err := ctx.Err(); err != nil { return MailboxSyncStepResult{}, err }
	if len(folders) > MaxEngineFolderItems {
		return service.reject(ctx, claim.Fence, now, MailboxSyncServiceErrorTooLarge)
	}
	folders = cloneMailboxSyncFolders(folders)

	page, err := service.enginePage(ctx, claim)
	if err != nil {
		return service.handlePageError(ctx, claim.Fence, now, err)
	}
	if err := ctx.Err(); err != nil { return MailboxSyncStepResult{}, err }
	candidates, candidateError, err := service.eventCandidates(ctx, page)
	if err != nil {
		return MailboxSyncStepResult{}, err
	}
	if candidateError != "" {
		return service.reject(ctx, claim.Fence, now, candidateError)
	}
	if err := ctx.Err(); err != nil {
		return MailboxSyncStepResult{}, err
	}
	page = cloneMailboxSyncPolicyPage(page)

	prepareRequest := MailboxSyncPrepareRequest{
		Fence:           claim.Fence,
		EngineAccountID: claim.EngineAccountID,
		Folders:         folders,
		Page:            page,
		EventCandidates: cloneMailboxSyncEventCandidates(candidates),
		OccurredAt:      now.UTC(),
		CorrelationID:   leaseID.String(),
	}
	expectedPrepareRequest := cloneMailboxSyncPrepareRequest(prepareRequest)
	prepared, err := service.repository.PrepareMailboxSyncPage(ctx, cloneMailboxSyncPrepareRequest(expectedPrepareRequest))
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil { return MailboxSyncStepResult{}, contextErr }
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorUnavailable)
	}
	if err := ctx.Err(); err != nil { return MailboxSyncStepResult{}, err }
	if !prepared.Prepared {
		if !mailboxSyncPolicyInputZero(prepared.Input) {
			return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid)
		}
		return MailboxSyncStepResult{Outcome: MailboxSyncStepLeaseLost}, nil
	}
	input := cloneMailboxSyncPolicyInput(prepared.Input)
	if !validMailboxSyncPreparedInput(input, claim, expectedPrepareRequest) {
		return service.reject(ctx, claim.Fence, now, MailboxSyncServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return MailboxSyncStepResult{}, err
	}
	policyResult, err := ClassifyMailboxSyncPage(input)
	if err != nil {
		if errors.Is(err, ErrMailboxSyncPolicyTooLarge) {
			return service.reject(ctx, claim.Fence, now, MailboxSyncServiceErrorTooLarge)
		}
		return service.reject(ctx, claim.Fence, now, MailboxSyncServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return MailboxSyncStepResult{}, err
	}
	policyResult = cloneMailboxSyncPolicyResult(policyResult)
	expectedPhase, ok := mailboxSyncExpectedPhase(claim.Fence, policyResult.Advance)
	if !ok {
		return service.reject(ctx, claim.Fence, now, MailboxSyncServiceErrorInvalid)
	}
	nextAttemptAt, err := service.nextAttemptAt(now, service.config.PollInterval)
	if err != nil {
		return service.reject(ctx, claim.Fence, now, MailboxSyncServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return MailboxSyncStepResult{}, err
	}
	applyRequest := MailboxSyncApplyRequest{CanonicalVersion: MailboxSyncCanonicalVersionV3, Fence: claim.Fence, Result: policyResult, Now: now, NextAttemptAt: nextAttemptAt}
	applied, err := service.repository.ApplyMailboxSyncPage(ctx, applyRequest)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil { return MailboxSyncStepResult{}, contextErr }
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorUnavailable)
	}
	if err := ctx.Err(); err != nil { return MailboxSyncStepResult{}, err }
	if !applied.Applied {
		if applied.Replayed { return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid) }
		return MailboxSyncStepResult{Outcome: MailboxSyncStepLeaseLost}, nil
	}
	if !validMailboxSyncApplyResult(applied, expectedPhase, nextAttemptAt) {
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid)
	}
	return MailboxSyncStepResult{Outcome: MailboxSyncStepApplied, NextAttemptAt: applied.NextAttemptAt}, nil
}

func (service *MailboxSyncService) enginePage(ctx context.Context, claim MailboxSyncClaimResult) (MailboxSyncPolicyPage, error) {
	requestPage := EnginePageRequest{Limit: service.config.PageSize}
	switch claim.Fence.Phase {
	case MailboxSyncPhaseIncremental, MailboxSyncPhaseCatchUp:
		page, err := service.engine.Changes(ctx, EngineChangeRequest{AccountID: claim.EngineAccountID, FromCursor: claim.Fence.Cursor, Page: requestPage})
		if err != nil { return MailboxSyncPolicyPage{}, err }
		if len(page.CreatedOrUpdated)+len(page.DestroyedIDs) > MaxEnginePageItems {
			return MailboxSyncPolicyPage{}, ErrEnginePayloadTooLarge
		}
		if page.AccountID != claim.EngineAccountID || page.FromCursor != claim.Fence.Cursor || page.Bounds.Requested != service.config.PageSize || page.Bounds.Maximum != MaxEnginePageItems || page.Bounds.Returned != len(page.CreatedOrUpdated)+len(page.DestroyedIDs) {
			return MailboxSyncPolicyPage{}, ErrEngineRejectedPolicy
		}
		return MailboxSyncPolicyPage{Change: cloneMailboxSyncChangePage(&page)}, nil
	case MailboxSyncPhaseEnumeration:
		page, err := service.engine.EnumerateProjection(ctx, EngineEnumerationRequest{AccountID: claim.EngineAccountID, BaseCursor: claim.Fence.BaseCursor, Checkpoint: claim.Fence.Checkpoint, Page: requestPage})
		if err != nil { return MailboxSyncPolicyPage{}, err }
		if len(page.Messages) > MaxEnginePageItems {
			return MailboxSyncPolicyPage{}, ErrEnginePayloadTooLarge
		}
		if page.AccountID != claim.EngineAccountID || page.BaseCursor == "" || (claim.Fence.BaseCursor != "" && page.BaseCursor != claim.Fence.BaseCursor) || page.Checkpoint != claim.Fence.Checkpoint || page.Bounds.Requested != service.config.PageSize || page.Bounds.Maximum != MaxEnginePageItems || page.Bounds.Returned != len(page.Messages) || (!page.Complete && page.NextCheckpoint == "") || (page.Complete && page.NextCheckpoint != "") {
			return MailboxSyncPolicyPage{}, ErrEngineRejectedPolicy
		}
		return MailboxSyncPolicyPage{Enumeration: cloneMailboxSyncEnumerationPage(&page)}, nil
	default:
		return MailboxSyncPolicyPage{}, ErrEngineRejectedPolicy
	}
}

func (service *MailboxSyncService) eventCandidates(ctx context.Context, page MailboxSyncPolicyPage) ([]MailboxSyncEventCandidate, MailboxSyncServiceErrorKind, error) {
	candidates := make([]MailboxSyncEventCandidate, 0)
	seenIDs := make(map[uuid.UUID]struct{})
	newID := func() (uuid.UUID, MailboxSyncServiceErrorKind, error) {
		if err := ctx.Err(); err != nil {
			return uuid.Nil, "", err
		}
		value := service.config.NewEventID()
		if err := ctx.Err(); err != nil {
			return uuid.Nil, "", err
		}
		if value == uuid.Nil {
			return uuid.Nil, MailboxSyncServiceErrorInvalid, nil
		}
		if _, found := seenIDs[value]; found {
			return uuid.Nil, MailboxSyncServiceErrorInvalid, nil
		}
		seenIDs[value] = struct{}{}
		return value, "", nil
	}
	if page.Change != nil && page.Enumeration == nil {
		if len(page.Change.CreatedOrUpdated)+len(page.Change.DestroyedIDs) > MaxEnginePageItems {
			return nil, MailboxSyncServiceErrorTooLarge, nil
		}
		live := append([]EngineMessageProjection(nil), page.Change.CreatedOrUpdated...)
		sort.Slice(live, func(left, right int) bool { return live[left].MessageID < live[right].MessageID })
		for _, projection := range live {
			if !syncValidOpaque(projection.MessageID, MaxEngineOpaqueIDBytes) {
				return nil, MailboxSyncServiceErrorInvalid, nil
			}
			received, candidateError, err := newID()
			if err != nil || candidateError != "" {
				return nil, candidateError, err
			}
			changed, candidateError, err := newID()
			if err != nil || candidateError != "" {
				return nil, candidateError, err
			}
			candidates = append(candidates, MailboxSyncEventCandidate{EngineMessageID: projection.MessageID, ReceivedEventID: received, ChangedEventID: changed})
		}
		destroyed := append([]string(nil), page.Change.DestroyedIDs...)
		sort.Strings(destroyed)
		for _, engineMessageID := range destroyed {
			if !syncValidOpaque(engineMessageID, MaxEngineOpaqueIDBytes) {
				return nil, MailboxSyncServiceErrorInvalid, nil
			}
			changed, candidateError, err := newID()
			if err != nil || candidateError != "" {
				return nil, candidateError, err
			}
			candidates = append(candidates, MailboxSyncEventCandidate{EngineMessageID: engineMessageID, ChangedEventID: changed, Destroyed: true})
		}
		return candidates, "", nil
	}
	if page.Enumeration != nil && page.Change == nil {
		if len(page.Enumeration.Messages) > MaxEnginePageItems {
			return nil, MailboxSyncServiceErrorTooLarge, nil
		}
		live := append([]EngineMessageProjection(nil), page.Enumeration.Messages...)
		sort.Slice(live, func(left, right int) bool { return live[left].MessageID < live[right].MessageID })
		for _, projection := range live {
			if !syncValidOpaque(projection.MessageID, MaxEngineOpaqueIDBytes) {
				return nil, MailboxSyncServiceErrorInvalid, nil
			}
			received, candidateError, err := newID()
			if err != nil || candidateError != "" {
				return nil, candidateError, err
			}
			changed, candidateError, err := newID()
			if err != nil || candidateError != "" {
				return nil, candidateError, err
			}
			candidates = append(candidates, MailboxSyncEventCandidate{EngineMessageID: projection.MessageID, ReceivedEventID: received, ChangedEventID: changed})
		}
		return candidates, "", nil
	}
	return nil, MailboxSyncServiceErrorInvalid, nil
}

func (service *MailboxSyncService) handleFolderError(ctx context.Context, fence MailboxSyncFence, now time.Time, err error) (MailboxSyncStepResult, error) {
	if contextErr := ctx.Err(); contextErr != nil {
		return MailboxSyncStepResult{}, contextErr
	}
	switch {
	case errors.Is(err, ErrEngineUnavailable), errors.Is(err, ErrEngineRateLimited), errors.Is(err, ErrEngineIndeterminate):
		return service.resolve(ctx, fence, now, MailboxSyncFailureTransient, MailboxSyncServiceErrorUnavailable)
	case errors.Is(err, ErrEnginePayloadTooLarge):
		return service.reject(ctx, fence, now, MailboxSyncServiceErrorTooLarge)
	case errors.Is(err, ErrEngineRejectedPolicy), errors.Is(err, ErrEngineUnauthorizedAccount):
		return service.reject(ctx, fence, now, MailboxSyncServiceErrorInvalid)
	default:
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorUnavailable)
	}
}

func (service *MailboxSyncService) handlePageError(ctx context.Context, fence MailboxSyncFence, now time.Time, err error) (MailboxSyncStepResult, error) {
	if contextErr := ctx.Err(); contextErr != nil { return MailboxSyncStepResult{}, contextErr }
	switch {
	case errors.Is(err, ErrEngineInvalidCursor) && (fence.Phase == MailboxSyncPhaseIncremental || fence.Phase == MailboxSyncPhaseCatchUp):
		return service.resolve(ctx, fence, now, MailboxSyncFailureInvalidCursor, MailboxSyncServiceErrorInvalid)
	case errors.Is(err, ErrEngineRestartRequired) && fence.Phase == MailboxSyncPhaseEnumeration:
		return service.resolve(ctx, fence, now, MailboxSyncFailureRestartRequired, MailboxSyncServiceErrorInvalid)
	case errors.Is(err, ErrEngineUnavailable), errors.Is(err, ErrEngineRateLimited), errors.Is(err, ErrEngineIndeterminate):
		return service.resolve(ctx, fence, now, MailboxSyncFailureTransient, MailboxSyncServiceErrorUnavailable)
	case errors.Is(err, ErrEnginePayloadTooLarge):
		return service.reject(ctx, fence, now, MailboxSyncServiceErrorTooLarge)
	case errors.Is(err, ErrEngineRejectedPolicy), errors.Is(err, ErrEngineUnauthorizedAccount):
		return service.reject(ctx, fence, now, MailboxSyncServiceErrorInvalid)
	default:
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorUnavailable)
	}
}

func (service *MailboxSyncService) reject(ctx context.Context, fence MailboxSyncFence, now time.Time, kind MailboxSyncServiceErrorKind) (MailboxSyncStepResult, error) {
	result, err := service.resolve(ctx, fence, now, MailboxSyncFailureRejected, kind)
	if err != nil { return MailboxSyncStepResult{}, err }
	if result.Outcome == MailboxSyncStepLeaseLost { return result, nil }
	return MailboxSyncStepResult{}, mailboxSyncServiceError(kind)
}

func (service *MailboxSyncService) resolve(ctx context.Context, fence MailboxSyncFence, now time.Time, failure MailboxSyncFailureKind, rejected MailboxSyncServiceErrorKind) (MailboxSyncStepResult, error) {
	nextAttemptAt := now
	if failure == MailboxSyncFailureTransient {
		var err error
		nextAttemptAt, err = service.nextAttemptAt(now, service.config.RetryInterval)
		if err != nil { return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid) }
	}
	if err := ctx.Err(); err != nil {
		return MailboxSyncStepResult{}, err
	}
	resolved, err := service.repository.ResolveMailboxSyncStep(ctx, MailboxSyncResolveRequest{Fence: fence, Failure: failure, Now: now, NextAttemptAt: nextAttemptAt})
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil { return MailboxSyncStepResult{}, contextErr }
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorUnavailable)
	}
	if err := ctx.Err(); err != nil { return MailboxSyncStepResult{}, err }
	if !resolved.Resolved {
		return MailboxSyncStepResult{Outcome: MailboxSyncStepLeaseLost}, nil
	}
	expectedPhase := fence.Phase
	if failure == MailboxSyncFailureInvalidCursor || failure == MailboxSyncFailureRestartRequired {
		expectedPhase = MailboxSyncPhaseEnumeration
	}
	if resolved.Phase != expectedPhase || !syncValidTime(resolved.NextAttemptAt) || !resolved.NextAttemptAt.Equal(nextAttemptAt) {
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid)
	}
	switch failure {
	case MailboxSyncFailureInvalidCursor, MailboxSyncFailureRestartRequired:
		if resolved.Phase != MailboxSyncPhaseEnumeration { return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid) }
		return MailboxSyncStepResult{Outcome: MailboxSyncStepRebuildScheduled, NextAttemptAt: resolved.NextAttemptAt}, nil
	case MailboxSyncFailureTransient:
		return MailboxSyncStepResult{Outcome: MailboxSyncStepRetryScheduled, NextAttemptAt: resolved.NextAttemptAt}, nil
	case MailboxSyncFailureRejected:
		return MailboxSyncStepResult{}, mailboxSyncServiceError(rejected)
	default:
		return MailboxSyncStepResult{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid)
	}
}

func (service *MailboxSyncService) nextAttemptAt(now time.Time, interval time.Duration) (time.Time, error) {
	jitter := service.config.Jitter(interval)
	if jitter < 0 || jitter >= interval { return time.Time{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid) }
	total, ok := mailboxSyncCheckedDurationAdd(interval, jitter)
	if !ok { return time.Time{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid) }
	next, ok := mailboxSyncCheckedTimeAdd(now, total)
	if !ok { return time.Time{}, mailboxSyncServiceError(MailboxSyncServiceErrorInvalid) }
	return next, nil
}

func validMailboxSyncServiceConfig(config MailboxSyncServiceConfig) bool {
	if config.Now == nil || config.NewLeaseID == nil || config.NewEventID == nil || config.Jitter == nil || config.LeaseDuration <= 0 || config.PollInterval <= 0 || config.RetryInterval <= 0 || config.PageSize < 1 || config.PageSize > MaxEnginePageItems {
		return false
	}
	_, pollSafe := mailboxSyncCheckedDurationAdd(config.PollInterval, config.PollInterval)
	_, retrySafe := mailboxSyncCheckedDurationAdd(config.RetryInterval, config.RetryInterval)
	return pollSafe && retrySafe
}

func validMailboxSyncIdleClaim(claim MailboxSyncClaimResult) bool {
	return claim.Fence == (MailboxSyncFence{}) && claim.EngineAccountID == "" && claim.EventCutoff.IsZero() && syncValidTime(claim.NextAttemptAt)
}

func validMailboxSyncClaim(accountID, leaseID uuid.UUID, claim MailboxSyncClaimResult) bool {
	return claim.Claimed && claim.Fence.AccountID == accountID && claim.Fence.LeaseID == leaseID && validMailboxSyncFence(claim.Fence) && syncValidOpaque(claim.EngineAccountID, MaxEngineOpaqueIDBytes) && syncValidTime(claim.EventCutoff)
}

func validMailboxSyncFence(fence MailboxSyncFence) bool {
	if fence.AccountID == uuid.Nil || fence.LeaseID == uuid.Nil || fence.LeaseEpoch <= 0 || fence.SyncEpoch < 0 || fence.MutationEpoch < 0 || !validMailboxSyncPhase(fence.Phase) {
		return false
	}
	switch fence.Phase {
	case MailboxSyncPhaseIncremental, MailboxSyncPhaseCatchUp:
		return syncValidCursor(fence.Cursor, false) && fence.BaseCursor == "" && fence.Checkpoint == ""
	case MailboxSyncPhaseEnumeration:
		return fence.Cursor == "" && ((fence.BaseCursor == "" && fence.Checkpoint == "") || (syncValidCursor(fence.BaseCursor, false) && syncValidCursor(fence.Checkpoint, false)))
	default:
		return false
	}
}

func validMailboxSyncPhase(phase MailboxSyncPhase) bool {
	return phase == MailboxSyncPhaseIncremental || phase == MailboxSyncPhaseEnumeration || phase == MailboxSyncPhaseCatchUp
}

func validMailboxSyncPreparedInput(input MailboxSyncPolicyInput, claim MailboxSyncClaimResult, request MailboxSyncPrepareRequest) bool {
	if input.SyncEpoch != claim.Fence.SyncEpoch || !input.EventCutoff.Equal(claim.EventCutoff) || input.Event.Mailbox.AccountID != claim.Fence.AccountID || input.Event.Mailbox.ID == uuid.Nil || input.Event.OrganizationID == uuid.Nil || input.Event.ActorID != nil || !input.Event.OccurredAt.Equal(request.OccurredAt) || input.Event.CorrelationID != request.CorrelationID {
		return false
	}
	if !validMailboxSyncPrepareSelection(input, request.EventCandidates) {
		return false
	}
	expectedPage, ok := mailboxSyncNormalizedPreparedPage(request.Page, input.Messages)
	return ok && reflect.DeepEqual(input.Page, expectedPage)
}

func mailboxSyncNormalizedPreparedPage(page MailboxSyncPolicyPage, messages []MailboxSyncPolicyMessage) (MailboxSyncPolicyPage, bool) {
	messageIndexByEngineID := make(map[string]int, len(messages))
	for index := range messages {
		if _, duplicate := messageIndexByEngineID[messages[index].EngineMessageID]; duplicate {
			return MailboxSyncPolicyPage{}, false
		}
		messageIndexByEngineID[messages[index].EngineMessageID] = index
	}

	result := cloneMailboxSyncPolicyPage(page)
	normalize := func(projections []EngineMessageProjection) bool {
		for index := range projections {
			messageIndex, found := messageIndexByEngineID[projections[index].MessageID]
			if !found {
				return false
			}
			provenance := messages[messageIndex].Provenance
			if provenance.ConfirmedSubmission != nil && provenance.ConfirmedDraft != nil {
				return false
			}
			switch {
			case provenance.ConfirmedSubmission != nil:
				projections[index].Origin = EngineProjectionOriginConfirmedOwnerSubmission
			case provenance.ConfirmedDraft != nil:
				projections[index].Origin = EngineProjectionOriginConfirmedOwnerDraft
			default:
				projections[index].Origin = EngineProjectionOriginInboundOrUnknown
			}
		}
		return true
	}

	if result.Change != nil && result.Enumeration == nil {
		return result, normalize(result.Change.CreatedOrUpdated)
	}
	if result.Enumeration != nil && result.Change == nil {
		return result, normalize(result.Enumeration.Messages)
	}
	return MailboxSyncPolicyPage{}, false
}

func validMailboxSyncPrepareSelection(input MailboxSyncPolicyInput, candidates []MailboxSyncEventCandidate) bool {
	live, destroyed, ok := mailboxSyncPageCandidateShapes(input.Page)
	if !ok || len(candidates) != len(live)+len(destroyed) || len(input.Messages) != len(live) || len(input.Destroyed) != len(destroyed) {
		return false
	}
	candidateByMessageID := make(map[string]MailboxSyncEventCandidate, len(candidates))
	seenEventIDs := make(map[uuid.UUID]struct{}, len(candidates)*2)
	for _, candidate := range candidates {
		if !syncValidOpaque(candidate.EngineMessageID, MaxEngineOpaqueIDBytes) || candidate.ChangedEventID == uuid.Nil || (!candidate.Destroyed && candidate.ReceivedEventID == uuid.Nil) || (candidate.Destroyed && candidate.ReceivedEventID != uuid.Nil) {
			return false
		}
		if _, found := candidateByMessageID[candidate.EngineMessageID]; found {
			return false
		}
		if _, found := seenEventIDs[candidate.ChangedEventID]; found {
			return false
		}
		seenEventIDs[candidate.ChangedEventID] = struct{}{}
		if !candidate.Destroyed {
			if _, found := seenEventIDs[candidate.ReceivedEventID]; found {
				return false
			}
			seenEventIDs[candidate.ReceivedEventID] = struct{}{}
		}
		candidateByMessageID[candidate.EngineMessageID] = candidate
	}
	seenLive := make(map[string]struct{}, len(input.Messages))
	for _, message := range input.Messages {
		candidate, found := candidateByMessageID[message.EngineMessageID]
		if !found || candidate.Destroyed || !live[message.EngineMessageID] {
			return false
		}
		if _, duplicate := seenLive[message.EngineMessageID]; duplicate || !mailboxSyncSelectedReceivedID(message.ReceivedEventID, candidate.ReceivedEventID) || !mailboxSyncSelectedChangedEvent(message.ChangedEvent, candidate.ChangedEventID, seenEventIDs) {
			return false
		}
		seenLive[message.EngineMessageID] = struct{}{}
	}
	seenDestroyed := make(map[string]struct{}, len(input.Destroyed))
	for _, value := range input.Destroyed {
		candidate, found := candidateByMessageID[value.EngineMessageID]
		if !found || !candidate.Destroyed || !destroyed[value.EngineMessageID] {
			return false
		}
		if _, duplicate := seenDestroyed[value.EngineMessageID]; duplicate || !mailboxSyncSelectedChangedEvent(value.ChangedEvent, candidate.ChangedEventID, seenEventIDs) {
			return false
		}
		seenDestroyed[value.EngineMessageID] = struct{}{}
	}
	return len(seenLive) == len(live) && len(seenDestroyed) == len(destroyed)
}

func mailboxSyncPageCandidateShapes(page MailboxSyncPolicyPage) (map[string]bool, map[string]bool, bool) {
	live := make(map[string]bool)
	destroyed := make(map[string]bool)
	if page.Change != nil && page.Enumeration == nil {
		for _, projection := range page.Change.CreatedOrUpdated {
			if !syncValidOpaque(projection.MessageID, MaxEngineOpaqueIDBytes) || live[projection.MessageID] || destroyed[projection.MessageID] {
				return nil, nil, false
			}
			live[projection.MessageID] = true
		}
		for _, engineMessageID := range page.Change.DestroyedIDs {
			if !syncValidOpaque(engineMessageID, MaxEngineOpaqueIDBytes) || live[engineMessageID] || destroyed[engineMessageID] {
				return nil, nil, false
			}
			destroyed[engineMessageID] = true
		}
		return live, destroyed, true
	}
	if page.Enumeration != nil && page.Change == nil {
		for _, projection := range page.Enumeration.Messages {
			if !syncValidOpaque(projection.MessageID, MaxEngineOpaqueIDBytes) || live[projection.MessageID] {
				return nil, nil, false
			}
			live[projection.MessageID] = true
		}
		return live, destroyed, true
	}
	return nil, nil, false
}

func mailboxSyncSelectedReceivedID(value, candidate uuid.UUID) bool {
	return value == uuid.Nil || value == candidate
}

func mailboxSyncSelectedChangedEvent(value MailboxSyncPolicyEventReservation, candidate uuid.UUID, generated map[uuid.UUID]struct{}) bool {
	if value.ID == uuid.Nil {
		return !value.Committed
	}
	if value.Committed {
		_, isFreshCandidate := generated[value.ID]
		return !isFreshCandidate
	}
	return value.ID == candidate
}

func validMailboxSyncApplyResult(result MailboxSyncApplyResult, expectedPhase MailboxSyncPhase, nextAttemptAt time.Time) bool {
	return result.Applied && (!result.Replayed || result.Applied) && result.Phase == expectedPhase && result.NextAttemptAt.Equal(nextAttemptAt) && syncValidTime(result.NextAttemptAt)
}

func mailboxSyncExpectedPhase(fence MailboxSyncFence, advance MailboxSyncPolicyAdvance) (MailboxSyncPhase, bool) {
	if advance.SyncEpoch != fence.SyncEpoch { return 0, false }
	if advance.Change != nil && advance.Enumeration == nil {
		if (fence.Phase != MailboxSyncPhaseIncremental && fence.Phase != MailboxSyncPhaseCatchUp) || advance.Change.FromCursor != fence.Cursor {
			return 0, false
		}
		if fence.Phase == MailboxSyncPhaseCatchUp && advance.Change.NextCursor == "" { return MailboxSyncPhaseIncremental, true }
		return fence.Phase, true
	}
	if advance.Enumeration != nil && advance.Change == nil {
		if fence.Phase != MailboxSyncPhaseEnumeration || advance.Enumeration.BaseCursor == "" || advance.Enumeration.Checkpoint != fence.Checkpoint {
			return 0, false
		}
		if fence.BaseCursor != "" && advance.Enumeration.BaseCursor != fence.BaseCursor { return 0, false }
		if advance.Enumeration.Complete {
			return MailboxSyncPhaseCatchUp, advance.Enumeration.NextCheckpoint == ""
		}
		return MailboxSyncPhaseEnumeration, advance.Enumeration.NextCheckpoint != ""
	}
	return 0, false
}

func mailboxSyncPolicyInputZero(input MailboxSyncPolicyInput) bool {
	return reflect.DeepEqual(input, MailboxSyncPolicyInput{})
}

func mailboxSyncCheckedDurationAdd(left, right time.Duration) (time.Duration, bool) {
	if left < 0 || right < 0 || left > time.Duration(1<<63-1)-right { return 0, false }
	return left + right, true
}

func mailboxSyncCheckedTimeAdd(now time.Time, duration time.Duration) (time.Time, bool) {
	if !syncValidTime(now) || duration <= 0 { return time.Time{}, false }
	next := now.Add(duration)
	return next, syncValidTime(next) && next.After(now)
}

func mailboxSyncServiceError(kind MailboxSyncServiceErrorKind) error {
	return &MailboxSyncServiceError{Kind: kind}
}

func mailboxSyncNilDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func cloneMailboxSyncFolders(values []EngineFolder) []EngineFolder {
	return append([]EngineFolder(nil), values...)
}

func cloneMailboxSyncEventCandidates(values []MailboxSyncEventCandidate) []MailboxSyncEventCandidate {
	return append([]MailboxSyncEventCandidate(nil), values...)
}

func cloneMailboxSyncPrepareRequest(value MailboxSyncPrepareRequest) MailboxSyncPrepareRequest {
	result := value
	result.Folders = cloneMailboxSyncFolders(value.Folders)
	result.Page = cloneMailboxSyncPolicyPage(value.Page)
	result.EventCandidates = cloneMailboxSyncEventCandidates(value.EventCandidates)
	return result
}

func cloneMailboxSyncPolicyPage(value MailboxSyncPolicyPage) MailboxSyncPolicyPage {
	return MailboxSyncPolicyPage{Change: cloneMailboxSyncChangePage(value.Change), Enumeration: cloneMailboxSyncEnumerationPage(value.Enumeration)}
}

func cloneMailboxSyncChangePage(value *EngineChangePage) *EngineChangePage {
	if value == nil { return nil }
	result := *value
	result.CreatedOrUpdated = cloneMailboxSyncProjections(value.CreatedOrUpdated)
	result.DestroyedIDs = append([]string(nil), value.DestroyedIDs...)
	return &result
}

func cloneMailboxSyncEnumerationPage(value *EngineEnumerationPage) *EngineEnumerationPage {
	if value == nil { return nil }
	result := *value
	result.Messages = cloneMailboxSyncProjections(value.Messages)
	return &result
}

func cloneMailboxSyncProjections(values []EngineMessageProjection) []EngineMessageProjection {
	result := make([]EngineMessageProjection, len(values))
	for index, value := range values { result[index] = cloneMailboxSyncProjection(value) }
	return result
}

func cloneMailboxSyncProjection(value EngineMessageProjection) EngineMessageProjection {
	result := value
	result.References = append([]string(nil), value.References...)
	result.To = append([]EngineAddress(nil), value.To...)
	result.CC = append([]EngineAddress(nil), value.CC...)
	result.BCC = append([]EngineAddress(nil), value.BCC...)
	result.ReplyTo = append([]EngineAddress(nil), value.ReplyTo...)
	result.FolderIDs = append([]string(nil), value.FolderIDs...)
	result.ReservedKeywords = append([]EngineKeywordMarker(nil), value.ReservedKeywords...)
	result.Attachments = make([]EngineProjectionAttachment, len(value.Attachments))
	for index, attachment := range value.Attachments {
		result.Attachments[index] = attachment
		result.Attachments[index].SHA256 = append([]byte(nil), attachment.SHA256...)
	}
	if value.SentAt != nil { copied := *value.SentAt; result.SentAt = &copied }
	if value.ReceivedAt != nil { copied := *value.ReceivedAt; result.ReceivedAt = &copied }
	result.CalendarMethod = copyMailCalendarMethod(value.CalendarMethod)
	if value.CalendarPartLocatorHash != nil {
		result.CalendarPartLocatorHash = append([]byte{}, value.CalendarPartLocatorHash...)
	}
	return result
}

func cloneMailboxSyncPolicyInput(value MailboxSyncPolicyInput) MailboxSyncPolicyInput {
	result := value
	result.Page = cloneMailboxSyncPolicyPage(value.Page)
	result.Folders = append([]MailboxSyncPolicyFolder(nil), value.Folders...)
	result.Messages = make([]MailboxSyncPolicyMessage, len(value.Messages))
	for index, message := range value.Messages {
		result.Messages[index] = message
		result.Messages[index].Thread.Ancestors = append([]ThreadAncestor(nil), message.Thread.Ancestors...)
		result.Messages[index].Previous = cloneMailboxSyncPrevious(message.Previous)
		result.Messages[index].Provenance.ConfirmedSubmission = cloneMailboxSyncProof(message.Provenance.ConfirmedSubmission)
		result.Messages[index].Provenance.ConfirmedDraft = cloneMailboxSyncProof(message.Provenance.ConfirmedDraft)
	}
	result.Destroyed = make([]MailboxSyncPolicyDestroyed, len(value.Destroyed))
	for index, destroyed := range value.Destroyed {
		result.Destroyed[index] = destroyed
		result.Destroyed[index].Labels = append([]MailboxSyncPolicyLabelState(nil), destroyed.Labels...)
	}
	result.Labels = append([]MailboxSyncPolicyLabelDefinition(nil), value.Labels...)
	if value.Event.ActorID != nil { actorID := *value.Event.ActorID; result.Event.ActorID = &actorID }
	return result
}

func cloneMailboxSyncPrevious(value *MailboxSyncPolicyPreviousMessage) *MailboxSyncPolicyPreviousMessage {
	if value == nil { return nil }
	result := *value
	result.FolderIDs = append([]MailboxEventID(nil), value.FolderIDs...)
	result.Labels = append([]MailboxSyncPolicyLabelState(nil), value.Labels...)
	if value.ReceivedEventID != nil { eventID := *value.ReceivedEventID; result.ReceivedEventID = &eventID }
	return &result
}

func cloneMailboxSyncProof(value *MailboxSyncPolicyProvenanceProof) *MailboxSyncPolicyProvenanceProof {
	if value == nil { return nil }
	result := *value
	return &result
}

func cloneMailboxSyncPolicyResult(value MailboxSyncPolicyResult) MailboxSyncPolicyResult {
	result := value
	if value.Advance.Change != nil { advance := *value.Advance.Change; result.Advance.Change = &advance }
	if value.Advance.Enumeration != nil { advance := *value.Advance.Enumeration; result.Advance.Enumeration = &advance }
	result.Folders = append([]MailboxSyncPolicyFolderIntent(nil), value.Folders...)
	result.Upserts = make([]MailboxSyncPolicyMessageIntent, len(value.Upserts))
	for index, upsert := range value.Upserts {
		result.Upserts[index] = upsert
		result.Upserts[index].ThreadIdentity.CanonicalParticipants = append([]string(nil), upsert.ThreadIdentity.CanonicalParticipants...)
		result.Upserts[index].Projection = cloneMailboxSyncProjection(upsert.Projection)
		result.Upserts[index].FolderIDs = append([]MailboxEventID(nil), upsert.FolderIDs...)
		result.Upserts[index].Labels = append([]MailboxSyncPolicyLabelState(nil), upsert.Labels...)
		result.Upserts[index].CalendarMethod = copyMailCalendarMethod(upsert.CalendarMethod)
		result.Upserts[index].CalendarPartLocatorHash = syncCloneBytes(upsert.CalendarPartLocatorHash)
		if upsert.ChangedFields != nil {
			result.Upserts[index].ChangedFields = make([]MailChangedField, len(upsert.ChangedFields))
			copy(result.Upserts[index].ChangedFields, upsert.ChangedFields)
		}
		if upsert.ReceivedEventID != nil { eventID := *upsert.ReceivedEventID; result.Upserts[index].ReceivedEventID = &eventID }
	}
	result.Tombstones = make([]MailboxSyncPolicyTombstoneIntent, len(value.Tombstones))
	for index, tombstone := range value.Tombstones {
		result.Tombstones[index] = tombstone
		result.Tombstones[index].Labels = append([]MailboxSyncPolicyLabelState(nil), tombstone.Labels...)
	}
	result.Counts.Folders = append([]MailboxEventID(nil), value.Counts.Folders...)
	result.Counts.Threads = append([]MailboxEventID(nil), value.Counts.Threads...)
	result.Events = append([]events.EnvelopeV1(nil), value.Events...)
	return result
}
