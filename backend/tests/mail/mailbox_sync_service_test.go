package mail_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

var (
	syncServiceAccountID = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	syncServiceLeaseID   = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	syncServiceNow       = time.Date(2031, time.February, 3, 4, 5, 6, 0, time.UTC)
)

func TestMailboxSyncService_FrozenABIAndSafeErrors(t *testing.T) {
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncFence{}), []string{"AccountID", "LeaseID", "LeaseEpoch", "SyncEpoch", "MutationEpoch", "Phase", "Cursor", "BaseCursor", "Checkpoint"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncClaimRequest{}), []string{"AccountID", "LeaseID", "Now", "LeaseUntil"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncClaimResult{}), []string{"Claimed", "Fence", "EngineAccountID", "EventCutoff", "NextAttemptAt"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncEventCandidate{}), []string{"EngineMessageID", "ReceivedEventID", "ChangedEventID", "Destroyed"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncPrepareRequest{}), []string{"Fence", "EngineAccountID", "Folders", "Page", "EventCandidates", "OccurredAt", "CorrelationID"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncPreparedPage{}), []string{"Prepared", "Input"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncApplyRequest{}), []string{"Fence", "Result", "Now", "NextAttemptAt"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncApplyResult{}), []string{"Applied", "Replayed", "Phase", "NextAttemptAt"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncResolveRequest{}), []string{"Fence", "Failure", "Now", "NextAttemptAt"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncResolveResult{}), []string{"Resolved", "Phase", "NextAttemptAt"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncServiceConfig{}), []string{"Now", "NewLeaseID", "NewEventID", "Jitter", "LeaseDuration", "PollInterval", "RetryInterval", "PageSize"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncStepResult{}), []string{"Outcome", "NextAttemptAt"})
	requireMailboxSyncFields(t, reflect.TypeOf(mail.MailboxSyncServiceError{}), []string{"Kind"})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncFence{}), []reflect.Type{reflect.TypeOf(uuid.UUID{}), reflect.TypeOf(uuid.UUID{}), reflect.TypeOf(int64(0)), reflect.TypeOf(int64(0)), reflect.TypeOf(int64(0)), reflect.TypeOf(mail.MailboxSyncPhase(0)), reflect.TypeOf(""), reflect.TypeOf(""), reflect.TypeOf("")})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncClaimRequest{}), []reflect.Type{reflect.TypeOf(uuid.UUID{}), reflect.TypeOf(uuid.UUID{}), reflect.TypeOf(time.Time{}), reflect.TypeOf(time.Time{})})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncClaimResult{}), []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(mail.MailboxSyncFence{}), reflect.TypeOf(""), reflect.TypeOf(time.Time{}), reflect.TypeOf(time.Time{})})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncEventCandidate{}), []reflect.Type{reflect.TypeOf(""), reflect.TypeOf(uuid.UUID{}), reflect.TypeOf(uuid.UUID{}), reflect.TypeOf(false)})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncPrepareRequest{}), []reflect.Type{reflect.TypeOf(mail.MailboxSyncFence{}), reflect.TypeOf(""), reflect.TypeOf([]mail.EngineFolder{}), reflect.TypeOf(mail.MailboxSyncPolicyPage{}), reflect.TypeOf([]mail.MailboxSyncEventCandidate{}), reflect.TypeOf(time.Time{}), reflect.TypeOf("")})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncPreparedPage{}), []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(mail.MailboxSyncPolicyInput{})})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncApplyRequest{}), []reflect.Type{reflect.TypeOf(mail.MailboxSyncFence{}), reflect.TypeOf(mail.MailboxSyncPolicyResult{}), reflect.TypeOf(time.Time{}), reflect.TypeOf(time.Time{})})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncApplyResult{}), []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(false), reflect.TypeOf(mail.MailboxSyncPhase(0)), reflect.TypeOf(time.Time{})})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncResolveRequest{}), []reflect.Type{reflect.TypeOf(mail.MailboxSyncFence{}), reflect.TypeOf(mail.MailboxSyncFailureKind(0)), reflect.TypeOf(time.Time{}), reflect.TypeOf(time.Time{})})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncResolveResult{}), []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(mail.MailboxSyncPhase(0)), reflect.TypeOf(time.Time{})})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncServiceConfig{}), []reflect.Type{reflect.TypeOf((func() time.Time)(nil)), reflect.TypeOf((func() uuid.UUID)(nil)), reflect.TypeOf((func() uuid.UUID)(nil)), reflect.TypeOf((func(time.Duration) time.Duration)(nil)), reflect.TypeOf(time.Duration(0)), reflect.TypeOf(time.Duration(0)), reflect.TypeOf(time.Duration(0)), reflect.TypeOf(int(0))})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncStepResult{}), []reflect.Type{reflect.TypeOf(mail.MailboxSyncStepOutcome(0)), reflect.TypeOf(time.Time{})})
	requireMailboxSyncFieldTypes(t, reflect.TypeOf(mail.MailboxSyncServiceError{}), []reflect.Type{reflect.TypeOf(mail.MailboxSyncServiceErrorKind(""))})
	requireMailboxSyncMethods(t, reflect.TypeOf((*mail.MailboxSyncRepository)(nil)).Elem(), []string{"ClaimMailboxSyncStep", "PrepareMailboxSyncPage", "ApplyMailboxSyncPage", "ResolveMailboxSyncStep"})
	requireMailboxSyncMethods(t, reflect.TypeOf((*mail.MailboxSyncService)(nil)), []string{"Step"})
	requireMailboxSyncMethods(t, reflect.TypeOf((*mail.MailboxSyncServiceError)(nil)), []string{"Error", "Unwrap"})
	requireMailboxSyncMethodType(t, reflect.TypeOf((*mail.MailboxSyncRepository)(nil)).Elem(), "ClaimMailboxSyncStep", reflect.TypeOf((func(context.Context, mail.MailboxSyncClaimRequest) (mail.MailboxSyncClaimResult, error))(nil)))
	requireMailboxSyncMethodType(t, reflect.TypeOf((*mail.MailboxSyncRepository)(nil)).Elem(), "PrepareMailboxSyncPage", reflect.TypeOf((func(context.Context, mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error))(nil)))
	requireMailboxSyncMethodType(t, reflect.TypeOf((*mail.MailboxSyncRepository)(nil)).Elem(), "ApplyMailboxSyncPage", reflect.TypeOf((func(context.Context, mail.MailboxSyncApplyRequest) (mail.MailboxSyncApplyResult, error))(nil)))
	requireMailboxSyncMethodType(t, reflect.TypeOf((*mail.MailboxSyncRepository)(nil)).Elem(), "ResolveMailboxSyncStep", reflect.TypeOf((func(context.Context, mail.MailboxSyncResolveRequest) (mail.MailboxSyncResolveResult, error))(nil)))
	requireMailboxSyncMethodType(t, reflect.TypeOf((*mail.MailboxSyncService)(nil)), "Step", reflect.TypeOf((func(*mail.MailboxSyncService, context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error))(nil)))
	requireMailboxSyncMethodType(t, reflect.TypeOf((*mail.MailboxSyncServiceError)(nil)), "Error", reflect.TypeOf((func(*mail.MailboxSyncServiceError) string)(nil)))
	requireMailboxSyncMethodType(t, reflect.TypeOf((*mail.MailboxSyncServiceError)(nil)), "Unwrap", reflect.TypeOf((func(*mail.MailboxSyncServiceError) error)(nil)))
	if signature := reflect.TypeOf(mail.NewMailboxSyncService); signature.NumIn() != 3 || signature.NumOut() != 2 || signature.In(0) != reflect.TypeOf((*mail.MailEngine)(nil)).Elem() || signature.In(1) != reflect.TypeOf((*mail.MailboxSyncRepository)(nil)).Elem() || signature.In(2) != reflect.TypeOf(mail.MailboxSyncServiceConfig{}) || signature.Out(0) != reflect.TypeOf((*mail.MailboxSyncService)(nil)) || signature.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		t.Fatalf("NewMailboxSyncService signature changed: %v", signature)
	}

	if mail.MailboxSyncPhaseIncremental != 1 || mail.MailboxSyncPhaseEnumeration != 2 || mail.MailboxSyncPhaseCatchUp != 3 ||
		mail.MailboxSyncFailureInvalidCursor != 1 || mail.MailboxSyncFailureRestartRequired != 2 || mail.MailboxSyncFailureTransient != 3 || mail.MailboxSyncFailureRejected != 4 ||
		mail.MailboxSyncStepIdle != 1 || mail.MailboxSyncStepApplied != 2 || mail.MailboxSyncStepRebuildScheduled != 3 || mail.MailboxSyncStepRetryScheduled != 4 || mail.MailboxSyncStepLeaseLost != 5 {
		t.Fatal("frozen mailbox sync enum values changed")
	}

	for _, testCase := range []struct {
		kind mail.MailboxSyncServiceErrorKind
		want error
		text string
	}{
		{mail.MailboxSyncServiceErrorInvalid, mail.ErrMailboxSyncServiceInvalid, "mailbox sync service invalid"},
		{mail.MailboxSyncServiceErrorTooLarge, mail.ErrMailboxSyncServiceTooLarge, "mailbox sync service exceeds bounds"},
		{mail.MailboxSyncServiceErrorUnavailable, mail.ErrMailboxSyncServiceUnavailable, "mailbox sync service unavailable"},
	} {
		err := &mail.MailboxSyncServiceError{Kind: testCase.kind}
		if !errors.Is(err, testCase.want) || err.Error() != testCase.text || err.Unwrap() != testCase.want {
			t.Fatalf("safe error %#v did not preserve its exact sentinel contract", err)
		}
	}
	var nilError *mail.MailboxSyncServiceError
	if nilError.Error() != "mailbox sync service error" || nilError.Unwrap() != nil {
		t.Fatal("nil service error must be content-free")
	}
	unknown := &mail.MailboxSyncServiceError{Kind: mail.MailboxSyncServiceErrorKind("unknown")}
	if unknown.Error() != "mailbox sync service error" || unknown.Unwrap() != nil || errors.Is(unknown, mail.ErrMailboxSyncServiceInvalid) || errors.Is(unknown, mail.ErrMailboxSyncServiceTooLarge) || errors.Is(unknown, mail.ErrMailboxSyncServiceUnavailable) {
		t.Fatalf("unknown service error leaked a sentinel or detail: %#v", unknown)
	}
}

func TestNewMailboxSyncService_RejectsInvalidDependenciesAndBounds(t *testing.T) {
	engine := &syncServiceEngineFake{}
	repository := &syncServiceRepositoryFake{}
	valid := (&syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}}).config()
	for _, testCase := range []struct {
		name string
		engine mail.MailEngine
		repository mail.MailboxSyncRepository
		mutate func(*mail.MailboxSyncServiceConfig)
	}{
		{"nil engine", nil, repository, nil},
		{"nil repository", engine, nil, nil},
		{"nil now", engine, repository, func(config *mail.MailboxSyncServiceConfig) { config.Now = nil }},
		{"nil lease id", engine, repository, func(config *mail.MailboxSyncServiceConfig) { config.NewLeaseID = nil }},
		{"nil event id", engine, repository, func(config *mail.MailboxSyncServiceConfig) { config.NewEventID = nil }},
		{"nil jitter", engine, repository, func(config *mail.MailboxSyncServiceConfig) { config.Jitter = nil }},
		{"zero lease", engine, repository, func(config *mail.MailboxSyncServiceConfig) { config.LeaseDuration = 0 }},
		{"negative poll", engine, repository, func(config *mail.MailboxSyncServiceConfig) { config.PollInterval = -time.Second }},
		{"overflow retry", engine, repository, func(config *mail.MailboxSyncServiceConfig) { config.RetryInterval = time.Duration(1<<63 - 1) }},
		{"zero page size", engine, repository, func(config *mail.MailboxSyncServiceConfig) { config.PageSize = 0 }},
		{"page size above maximum", engine, repository, func(config *mail.MailboxSyncServiceConfig) { config.PageSize = mail.MaxEnginePageItems + 1 }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			config := valid
			if testCase.mutate != nil { testCase.mutate(&config) }
			service, err := mail.NewMailboxSyncService(testCase.engine, testCase.repository, config)
			if service != nil || !errors.Is(err, mail.ErrMailboxSyncServiceInvalid) || strings.Contains(err.Error(), "overflow") {
				t.Fatalf("NewMailboxSyncService() = %#v, %v; want content-free invalid", service, err)
			}
		})
	}
}

func TestNewMailboxSyncService_RejectsTypedNilDependenciesWithoutPanicking(t *testing.T) {
	var nilEngine *syncServiceEngineFake
	var nilRepository *syncServiceRepositoryFake
	valid := (&syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}}).config()
	for _, testCase := range []struct {
		name string
		engine mail.MailEngine
		repository mail.MailboxSyncRepository
	}{
		{"typed nil engine", nilEngine, &syncServiceRepositoryFake{}},
		{"typed nil repository", &syncServiceEngineFake{}, nilRepository},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			panicked := false
			var service *mail.MailboxSyncService
			var err error
			func() {
				defer func() { panicked = recover() != nil }()
				service, err = mail.NewMailboxSyncService(testCase.engine, testCase.repository, valid)
			}()
			if panicked || service != nil || !errors.Is(err, mail.ErrMailboxSyncServiceInvalid) {
				t.Fatalf("typed nil dependency yielded service=%#v err=%v panic=%v", service, err, panicked)
			}
		})
	}
}

func TestMailboxSyncServiceStep_IdleAndClaimGuardHaveNoEngineWork(t *testing.T) {
	for _, testCase := range []struct {
		name string
		claim mail.MailboxSyncClaimResult
	}{
		{"not due", mail.MailboxSyncClaimResult{NextAttemptAt: syncServiceNow.Add(time.Minute)}},
		{"live mutation lease", mail.MailboxSyncClaimResult{NextAttemptAt: syncServiceNow.Add(2 * time.Minute)}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			engine := &syncServiceEngineFake{}
			repository := &syncServiceRepositoryFake{claim: testCase.claim}
			trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}}
			service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
			if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
			result, err := service.Step(context.Background(), syncServiceAccountID)
			if err != nil || result.Outcome != mail.MailboxSyncStepIdle || !result.NextAttemptAt.Equal(testCase.claim.NextAttemptAt) {
				t.Fatalf("Step() = %#v, %v; want idle", result, err)
			}
			if repository.claimCalls != 1 || repository.prepareCalls != 0 || repository.applyCalls != 0 || repository.resolveCalls != 0 || engine.listCalls != 0 || engine.changeCalls != 0 || engine.enumerateCalls != 0 || trace.nowCalls != 1 || trace.leaseCalls != 1 || trace.eventCalls != 0 || trace.jitterCalls != 0 {
				t.Fatalf("idle calls claim/prepare/apply/resolve/list/change/enumerate/now/lease/event/jitter = %d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d", repository.claimCalls, repository.prepareCalls, repository.applyCalls, repository.resolveCalls, engine.listCalls, engine.changeCalls, engine.enumerateCalls, trace.nowCalls, trace.leaseCalls, trace.eventCalls, trace.jitterCalls)
			}
		})
	}
}

func TestMailboxSyncServiceStep_AppliesIncrementalPageWithExactConfiguredLimitAndIsolation(t *testing.T) {
	policyInput := validMailboxSyncInput(t)
	policyInput.Page.Change.Bounds.Requested = 17
	fence := syncServiceFence(mail.MailboxSyncPhaseIncremental)
	engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *policyInput.Page.Change}
	trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: []uuid.UUID{syncServiceUUID("received"), syncServiceUUID("changed")}, jitter: time.Second}
	repository := &syncServiceRepositoryFake{claim: syncServiceClaim(fence, policyInput.EventCutoff)}
	repository.prepare = func(request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
		if request.OccurredAt != syncServiceNow.UTC() || request.CorrelationID != syncServiceLeaseID.String() || len(request.EventCandidates) != 1 || request.EventCandidates[0].EngineMessageID != "message-04f" || request.EventCandidates[0].Destroyed || request.EventCandidates[0].ReceivedEventID != trace.eventIDs[0] || request.EventCandidates[0].ChangedEventID != trace.eventIDs[1] {
			t.Fatalf("prepare request = %#v, want canonical occurrence/correlation and deterministic candidate", request)
		}
		input := policyInput
		input.Page = request.Page
		input.Event.OccurredAt = request.OccurredAt
		input.Event.CorrelationID = request.CorrelationID
		input.EventCutoff = policyInput.EventCutoff
		input.SyncEpoch = request.Fence.SyncEpoch
		input.Messages[0].ReceivedEventID = request.EventCandidates[0].ReceivedEventID
		input.Messages[0].ChangedEvent = mail.MailboxSyncPolicyEventReservation{}
		return mail.MailboxSyncPreparedPage{Prepared: true, Input: input}, nil
	}
	repository.apply = func(request mail.MailboxSyncApplyRequest) (mail.MailboxSyncApplyResult, error) {
		if request.Fence != fence || !request.Now.Equal(syncServiceNow) || !request.NextAttemptAt.Equal(syncServiceNow.Add(11*time.Second)) || request.Result.Advance.Change == nil || request.Result.Advance.Change.FromCursor != "cursor-a" || request.Result.Advance.Change.NextCursor != "cursor-b" || len(request.Result.Events) != 1 {
			t.Fatalf("apply request = %#v, want exact incremental result and schedule", request)
		}
		return mail.MailboxSyncApplyResult{Applied: true, Phase: mail.MailboxSyncPhaseIncremental, NextAttemptAt: request.NextAttemptAt}, nil
	}

	service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
	if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
	result, err := service.Step(context.Background(), syncServiceAccountID)
	if err != nil || result.Outcome != mail.MailboxSyncStepApplied || !result.NextAttemptAt.Equal(syncServiceNow.Add(11*time.Second)) {
		t.Fatalf("Step() = %#v, %v; want applied", result, err)
	}
	if engine.listCalls != 1 || engine.changeCalls != 1 || engine.enumerateCalls != 0 || engine.changeRequest.Page.Limit != 17 || repository.prepareCalls != 1 || repository.applyCalls != 1 || repository.resolveCalls != 0 || trace.nowCalls != 1 || trace.leaseCalls != 1 || trace.eventCalls != 2 || trace.jitterCalls != 1 || trace.jitterInput != 10*time.Second {
		t.Fatalf("success call matrix did not preserve one-step exact PageSize flow")
	}
	engine.change.CreatedOrUpdated[0].MessageID = "hostile-after-call"
	if repository.applyRequest.Result.Upserts[0].Projection.MessageID != "message-04f" {
		t.Fatalf("applied result retained engine-owned alias: %#v", repository.applyRequest.Result)
	}
}

func TestMailboxSyncServiceStep_UsesEnumerationAndCatchUpRequestsWithoutSameStepLoop(t *testing.T) {
	for _, testCase := range []struct {
		name string
		phase mail.MailboxSyncPhase
		base string
		checkpoint string
		wantChanges int
		wantEnumeration int
	}{
		{"initial enumeration", mail.MailboxSyncPhaseEnumeration, "", "", 0, 1},
		{"resumed enumeration", mail.MailboxSyncPhaseEnumeration, "base-a", "checkpoint-a", 0, 1},
		{"catch up", mail.MailboxSyncPhaseCatchUp, "", "", 1, 0},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validMailboxSyncInput(t)
			input.Page.Change.Bounds.Requested = 17
			fence := syncServiceFence(testCase.phase)
			fence.BaseCursor, fence.Checkpoint = testCase.base, testCase.checkpoint
			if testCase.phase == mail.MailboxSyncPhaseEnumeration {
				fence.Cursor = ""
				input = mailboxSyncInputWithEnumeration(t, 1)
				input.Page.Enumeration.BaseCursor = "base-a"
				input.Page.Enumeration.Checkpoint = testCase.checkpoint
				input.Page.Enumeration.NextCheckpoint = ""
				input.Page.Enumeration.Complete = true
				input.Page.Enumeration.Bounds.Requested = 17
			} else {
				fence.Cursor = "cursor-a"
			}
			changeInput := validMailboxSyncInput(t)
			changeInput.Page.Change.Bounds.Requested = 17
			engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *changeInput.Page.Change}
			if testCase.phase == mail.MailboxSyncPhaseEnumeration { engine.enumeration = *input.Page.Enumeration }
			trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: []uuid.UUID{syncServiceUUID("received"), syncServiceUUID("changed")}}
			repository := &syncServiceRepositoryFake{claim: syncServiceClaim(fence, input.EventCutoff)}
			repository.prepare = func(request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
				prepared := input
				prepared.Page = request.Page
				prepared.Event.OccurredAt, prepared.Event.CorrelationID = request.OccurredAt, request.CorrelationID
				prepared.SyncEpoch = request.Fence.SyncEpoch
				prepared.Messages[0].ReceivedEventID = request.EventCandidates[0].ReceivedEventID
				return mail.MailboxSyncPreparedPage{Prepared: true, Input: prepared}, nil
			}
			repository.apply = func(request mail.MailboxSyncApplyRequest) (mail.MailboxSyncApplyResult, error) {
				phase := testCase.phase
				if testCase.phase == mail.MailboxSyncPhaseEnumeration { phase = mail.MailboxSyncPhaseCatchUp }
				return mail.MailboxSyncApplyResult{Applied: true, Phase: phase, NextAttemptAt: request.NextAttemptAt}, nil
			}
			service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
			if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
			if _, err := service.Step(context.Background(), syncServiceAccountID); err != nil { t.Fatalf("Step() error = %v", err) }
			if engine.changeCalls != testCase.wantChanges || engine.enumerateCalls != testCase.wantEnumeration || engine.listCalls != 1 {
				t.Fatalf("engine calls list/change/enumerate = %d/%d/%d, want 1/%d/%d", engine.listCalls, engine.changeCalls, engine.enumerateCalls, testCase.wantChanges, testCase.wantEnumeration)
			}
			if testCase.phase == mail.MailboxSyncPhaseEnumeration && (engine.enumerationRequest.Page.Limit != 17 || engine.enumerationRequest.BaseCursor != testCase.base || engine.enumerationRequest.Checkpoint != testCase.checkpoint) {
				t.Fatalf("enumeration request = %#v", engine.enumerationRequest)
			}
			if testCase.phase == mail.MailboxSyncPhaseCatchUp && (engine.changeRequest.Page.Limit != 17 || engine.changeRequest.FromCursor != "cursor-a") {
				t.Fatalf("catch-up request = %#v", engine.changeRequest)
			}
		})
	}
}

func TestMailboxSyncServiceStep_LeaseLossFailureMappingAndCaps(t *testing.T) {
	for _, testCase := range []struct {
		name string
		prepare func(*syncServiceEngineFake, *syncServiceRepositoryFake, *syncServiceCallbackTrace)
		wantOutcome mail.MailboxSyncStepOutcome
		wantError error
		wantResolve mail.MailboxSyncFailureKind
	}{
		{
			name: "prepared false is lease lost",
			prepare: func(_ *syncServiceEngineFake, repository *syncServiceRepositoryFake, _ *syncServiceCallbackTrace) { repository.prepare = func(mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) { return mail.MailboxSyncPreparedPage{}, nil } },
			wantOutcome: mail.MailboxSyncStepLeaseLost,
		},
		{
			name: "apply false is lease lost",
			prepare: func(_ *syncServiceEngineFake, repository *syncServiceRepositoryFake, _ *syncServiceCallbackTrace) { repository.apply = func(mail.MailboxSyncApplyRequest) (mail.MailboxSyncApplyResult, error) { return mail.MailboxSyncApplyResult{}, nil } },
			wantOutcome: mail.MailboxSyncStepLeaseLost,
		},
		{
			name: "invalid cursor resolves rebuild",
			prepare: func(engine *syncServiceEngineFake, _ *syncServiceRepositoryFake, _ *syncServiceCallbackTrace) { engine.changeErr = mail.ErrEngineInvalidCursor },
			wantOutcome: mail.MailboxSyncStepRebuildScheduled, wantResolve: mail.MailboxSyncFailureInvalidCursor,
		},
		{
			name: "unavailable resolves retry",
			prepare: func(engine *syncServiceEngineFake, _ *syncServiceRepositoryFake, _ *syncServiceCallbackTrace) { engine.changeErr = mail.ErrEngineUnavailable },
			wantOutcome: mail.MailboxSyncStepRetryScheduled, wantResolve: mail.MailboxSyncFailureTransient,
		},
		{
			name: "resolve guard loss is lease lost",
			prepare: func(engine *syncServiceEngineFake, repository *syncServiceRepositoryFake, _ *syncServiceCallbackTrace) {
				engine.changeErr = mail.ErrEngineUnavailable
				repository.resolveResult = mail.MailboxSyncResolveResult{}
			},
			wantOutcome: mail.MailboxSyncStepLeaseLost, wantResolve: mail.MailboxSyncFailureTransient,
		},
		{
			name: "hostile engine error is unavailable without detail",
			prepare: func(engine *syncServiceEngineFake, _ *syncServiceRepositoryFake, _ *syncServiceCallbackTrace) { engine.changeErr = errors.New("provider cursor=secret body=leak") },
			wantError: mail.ErrMailboxSyncServiceUnavailable,
		},
		{
			name: "folder cap rejects before ids prepare and apply",
			prepare: func(engine *syncServiceEngineFake, _ *syncServiceRepositoryFake, _ *syncServiceCallbackTrace) { engine.folders = make([]mail.EngineFolder, mail.MaxEngineFolderItems+1) },
			wantError: mail.ErrMailboxSyncServiceTooLarge, wantResolve: mail.MailboxSyncFailureRejected,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validMailboxSyncInput(t)
			input.Page.Change.Bounds.Requested = 17
			fence := syncServiceFence(mail.MailboxSyncPhaseIncremental)
			engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change}
			trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: []uuid.UUID{syncServiceUUID("received"), syncServiceUUID("changed")}}
			repository := &syncServiceRepositoryFake{claim: syncServiceClaim(fence, input.EventCutoff), resolveResult: mail.MailboxSyncResolveResult{Resolved: true, Phase: mail.MailboxSyncPhaseEnumeration}}
			repository.prepare = func(request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
				prepared := input
				prepared.Page = request.Page
				prepared.Event.OccurredAt, prepared.Event.CorrelationID = request.OccurredAt, request.CorrelationID
				prepared.SyncEpoch = request.Fence.SyncEpoch
				prepared.Messages[0].ReceivedEventID = request.EventCandidates[0].ReceivedEventID
				return mail.MailboxSyncPreparedPage{Prepared: true, Input: prepared}, nil
			}
			repository.apply = func(request mail.MailboxSyncApplyRequest) (mail.MailboxSyncApplyResult, error) { return mail.MailboxSyncApplyResult{Applied: true, Phase: mail.MailboxSyncPhaseIncremental, NextAttemptAt: request.NextAttemptAt}, nil }
			testCase.prepare(engine, repository, trace)
			service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
			if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
			result, err := service.Step(context.Background(), syncServiceAccountID)
			if testCase.wantError != nil {
				if !errors.Is(err, testCase.wantError) || strings.Contains(err.Error(), "provider") {
					t.Fatalf("Step() error = %v, want content-free %v", err, testCase.wantError)
				}
			} else if err != nil || result.Outcome != testCase.wantOutcome {
				t.Fatalf("Step() = %#v, %v; want outcome %v", result, err, testCase.wantOutcome)
			}
			if testCase.wantResolve == 0 {
				if repository.resolveCalls != 0 { t.Fatalf("resolve calls = %d, want 0", repository.resolveCalls) }
			} else if repository.resolveCalls != 1 || repository.resolveRequest.Failure != testCase.wantResolve {
				t.Fatalf("resolve request = %#v, calls=%d", repository.resolveRequest, repository.resolveCalls)
			}
			if strings.Contains(errString(err), "secret") { t.Fatalf("raw dependency detail escaped: %v", err) }
		})
	}
}

func TestMailboxSyncServiceStep_ContextCancellationAfterCandidateCallbackStopsBeforePrepare(t *testing.T) {
	input := validMailboxSyncInput(t)
	input.Page.Change.Bounds.Requested = 17
	engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change}
	repository := &syncServiceRepositoryFake{claim: syncServiceClaim(syncServiceFence(mail.MailboxSyncPhaseIncremental), input.EventCutoff)}
	trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}}
	config := trace.config()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eventCalls := 0
	config.NewEventID = func() uuid.UUID {
		eventCalls++
		cancel()
		return syncServiceUUID("cancelled-event")
	}
	service, err := mail.NewMailboxSyncService(engine, repository, config)
	if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
	result, err := service.Step(ctx, syncServiceAccountID)
	if !errors.Is(err, context.Canceled) || result != (mail.MailboxSyncStepResult{}) || eventCalls != 1 || repository.prepareCalls != 0 || repository.applyCalls != 0 || repository.resolveCalls != 0 {
		t.Fatalf("cancellation did not stop at callback boundary: result=%#v err=%v event/prepare/apply/resolve=%d/%d/%d/%d", result, err, eventCalls, repository.prepareCalls, repository.applyCalls, repository.resolveCalls)
	}
}

func TestMailboxSyncServiceStep_RejectsPrepareCandidateSubstitutionBeforeClassifierAndApply(t *testing.T) {
	foreignID := syncServiceUUID("foreign-candidate")
	for _, testCase := range []struct {
		name string
		mutate func(*mail.MailboxSyncPolicyInput, mail.MailboxSyncPrepareRequest)
	}{
		{
			name: "swapped live received ids",
			mutate: func(input *mail.MailboxSyncPolicyInput, _ mail.MailboxSyncPrepareRequest) {
				input.Messages[0].ReceivedEventID, input.Messages[1].ReceivedEventID = input.Messages[1].ReceivedEventID, input.Messages[0].ReceivedEventID
			},
		},
		{
			name: "foreign live received id",
			mutate: func(input *mail.MailboxSyncPolicyInput, _ mail.MailboxSyncPrepareRequest) {
				input.Messages[0].ReceivedEventID = foreignID
			},
		},
		{
			name: "extra live reservation state",
			mutate: func(input *mail.MailboxSyncPolicyInput, _ mail.MailboxSyncPrepareRequest) {
				extra := input.Messages[0]
				extra.EngineMessageID = "extra-message"
				input.Messages = append(input.Messages, extra)
			},
		},
		{
			name: "missing live reservation state",
			mutate: func(input *mail.MailboxSyncPolicyInput, _ mail.MailboxSyncPrepareRequest) {
				input.Messages = input.Messages[:1]
			},
		},
		{
			name: "duplicate selected received id",
			mutate: func(input *mail.MailboxSyncPolicyInput, _ mail.MailboxSyncPrepareRequest) {
				input.Messages[1].ReceivedEventID = input.Messages[0].ReceivedEventID
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := mailboxSyncInputWithLiveMessages(t, 2)
			input.Page.Change.Bounds.Requested = 17
			engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change}
			ids := []uuid.UUID{syncServiceUUID(testCase.name + "-0"), syncServiceUUID(testCase.name + "-1"), syncServiceUUID(testCase.name + "-2"), syncServiceUUID(testCase.name + "-3")}
			trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: ids}
			repository := &syncServiceRepositoryFake{claim: syncServiceClaim(syncServiceFence(mail.MailboxSyncPhaseIncremental), input.EventCutoff), resolveResult: mail.MailboxSyncResolveResult{Resolved: true, Phase: mail.MailboxSyncPhaseIncremental}}
			repository.prepare = func(request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
				prepared := syncServicePreparedInputFromCandidates(t, input, request)
				testCase.mutate(&prepared, request)
				return mail.MailboxSyncPreparedPage{Prepared: true, Input: prepared}, nil
			}
			service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
			if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
			result, err := service.Step(context.Background(), syncServiceAccountID)
			if !errors.Is(err, mail.ErrMailboxSyncServiceInvalid) || result != (mail.MailboxSyncStepResult{}) || repository.prepareCalls != 1 || repository.applyCalls != 0 || repository.resolveCalls != 1 || repository.resolveRequest.Failure != mail.MailboxSyncFailureRejected {
				t.Fatalf("substituted prepared state escaped: result=%#v err=%v prepare/apply/resolve=%d/%d/%d", result, err, repository.prepareCalls, repository.applyCalls, repository.resolveCalls)
			}
		})
	}
}

func TestMailboxSyncServiceStep_UsesImmutablePrepareExpectationAfterRepositoryAliasMutation(t *testing.T) {
	input := validMailboxSyncInput(t)
	input.Page.Change.Bounds.Requested = 17
	engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change}
	trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: []uuid.UUID{syncServiceUUID("received"), syncServiceUUID("changed")}}
	repository := &syncServiceRepositoryFake{claim: syncServiceClaim(syncServiceFence(mail.MailboxSyncPhaseIncremental), input.EventCutoff), resolveResult: mail.MailboxSyncResolveResult{Resolved: true, Phase: mail.MailboxSyncPhaseIncremental}}
	repository.prepare = func(request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
		request.EventCandidates[0].ReceivedEventID = syncServiceUUID("repository-poisoned-candidate")
		prepared := syncServicePreparedInputFromCandidates(t, input, request)
		return mail.MailboxSyncPreparedPage{Prepared: true, Input: prepared}, nil
	}
	service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
	if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
	result, err := service.Step(context.Background(), syncServiceAccountID)
	if !errors.Is(err, mail.ErrMailboxSyncServiceInvalid) || result != (mail.MailboxSyncStepResult{}) || repository.prepareCalls != 1 || repository.applyCalls != 0 || repository.resolveCalls != 1 || repository.resolveRequest.Failure != mail.MailboxSyncFailureRejected {
		t.Fatalf("repository request alias changed service expectation: result=%#v err=%v prepare/apply/resolve=%d/%d/%d", result, err, repository.prepareCalls, repository.applyCalls, repository.resolveCalls)
	}
}

func TestMailboxSyncServiceStep_AllowsOnlyPreexistingCommittedReservations(t *testing.T) {
	for _, testCase := range []struct {
		name string
		committedID func(mail.MailboxSyncPrepareRequest) uuid.UUID
		wantError error
	}{
		{
			name: "fresh changed candidate cannot be committed", committedID: func(request mail.MailboxSyncPrepareRequest) uuid.UUID { return request.EventCandidates[0].ChangedEventID }, wantError: mail.ErrMailboxSyncServiceInvalid,
		},
		{
			name: "preexisting committed reservation is replay safe", committedID: func(mail.MailboxSyncPrepareRequest) uuid.UUID { return syncServiceUUID("already-committed") },
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validMailboxSyncInput(t)
			input.Page.Change.Bounds.Requested = 17
			input.Messages[0].Previous = &mail.MailboxSyncPolicyPreviousMessage{Thread: input.Messages[0].Thread.Thread, FolderIDs: []mail.MailboxEventID{input.Folders[0].Folder}, Flags: mail.EngineMessageFlags{Seen: false}, Labels: []mail.MailboxSyncPolicyLabelState{}}
			engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change}
			trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: []uuid.UUID{syncServiceUUID(testCase.name + "-received"), syncServiceUUID(testCase.name + "-changed")}}
			repository := &syncServiceRepositoryFake{claim: syncServiceClaim(syncServiceFence(mail.MailboxSyncPhaseIncremental), input.EventCutoff), resolveResult: mail.MailboxSyncResolveResult{Resolved: true, Phase: mail.MailboxSyncPhaseIncremental}}
			repository.prepare = func(request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
				prepared := syncServicePreparedInputFromCandidates(t, input, request)
				prepared.Messages[0].ReceivedEventID = uuid.Nil
				prepared.Messages[0].ChangedEvent = mail.MailboxSyncPolicyEventReservation{ID: testCase.committedID(request), Committed: true}
				return mail.MailboxSyncPreparedPage{Prepared: true, Input: prepared}, nil
			}
			repository.apply = func(request mail.MailboxSyncApplyRequest) (mail.MailboxSyncApplyResult, error) {
				if len(request.Result.Events) != 0 { t.Fatalf("committed reservation emitted a duplicate event: %#v", request.Result.Events) }
				return mail.MailboxSyncApplyResult{Applied: true, Phase: mail.MailboxSyncPhaseIncremental, NextAttemptAt: request.NextAttemptAt}, nil
			}
			service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
			if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
			result, err := service.Step(context.Background(), syncServiceAccountID)
			if testCase.wantError != nil {
				if result != (mail.MailboxSyncStepResult{}) || !errors.Is(err, testCase.wantError) || repository.applyCalls != 0 || repository.resolveCalls != 1 {
					t.Fatalf("fresh committed candidate was accepted: result=%#v err=%v apply/resolve=%d/%d", result, err, repository.applyCalls, repository.resolveCalls)
				}
			} else if err != nil || result.Outcome != mail.MailboxSyncStepApplied || repository.applyCalls != 1 {
				t.Fatalf("preexisting committed reservation did not replay safely: result=%#v err=%v apply=%d", result, err, repository.applyCalls)
			}
		})
	}
}

func TestMailboxSyncServiceStep_RejectsResolvePhaseThatDoesNotEchoDisposition(t *testing.T) {
	for _, testCase := range []struct {
		name string
		phase mail.MailboxSyncPhase
		configure func(*syncServiceEngineFake)
		responsePhase mail.MailboxSyncPhase
	}{
		{
			name: "transient must retain incremental phase", phase: mail.MailboxSyncPhaseIncremental,
			configure: func(engine *syncServiceEngineFake) { engine.changeErr = mail.ErrEngineUnavailable }, responsePhase: mail.MailboxSyncPhaseEnumeration,
		},
		{
			name: "rejected must retain incremental phase", phase: mail.MailboxSyncPhaseIncremental,
			configure: func(engine *syncServiceEngineFake) { engine.folders = make([]mail.EngineFolder, mail.MaxEngineFolderItems+1) }, responsePhase: mail.MailboxSyncPhaseEnumeration,
		},
		{
			name: "rebuild must enter enumeration", phase: mail.MailboxSyncPhaseEnumeration,
			configure: func(engine *syncServiceEngineFake) { engine.enumerationErr = mail.ErrEngineRestartRequired }, responsePhase: mail.MailboxSyncPhaseIncremental,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validMailboxSyncInput(t)
			fence := syncServiceFence(testCase.phase)
			if testCase.phase == mail.MailboxSyncPhaseEnumeration { fence.Cursor = "" }
			engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}}
			testCase.configure(engine)
			trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}}
			repository := &syncServiceRepositoryFake{claim: syncServiceClaim(fence, input.EventCutoff), resolveResult: mail.MailboxSyncResolveResult{Resolved: true, Phase: testCase.responsePhase}}
			service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
			if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
			result, err := service.Step(context.Background(), syncServiceAccountID)
			if !errors.Is(err, mail.ErrMailboxSyncServiceInvalid) || result != (mail.MailboxSyncStepResult{}) || repository.resolveCalls != 1 || repository.applyCalls != 0 {
				t.Fatalf("hostile resolve phase was accepted: result=%#v err=%v resolve/apply=%d/%d", result, err, repository.resolveCalls, repository.applyCalls)
			}
		})
	}
}

func TestMailboxSyncServiceStep_RepositoryErrorsAreContentFreeAndNeverResolveAmbiguousApply(t *testing.T) {
	for _, testCase := range []struct {
		name string
		configure func(*syncServiceEngineFake, *syncServiceRepositoryFake)
		wantEngineCalls int
		wantPrepare int
		wantApply int
		wantResolve int
	}{
		{
			name: "claim error", configure: func(_ *syncServiceEngineFake, repository *syncServiceRepositoryFake) { repository.claimErr = errors.New("repository token=claim-secret") },
		},
		{
			name: "prepare error", configure: func(_ *syncServiceEngineFake, repository *syncServiceRepositoryFake) { repository.prepareErr = errors.New("repository token=prepare-secret") },
			wantEngineCalls: 1, wantPrepare: 1,
		},
		{
			name: "ambiguous apply error", configure: func(_ *syncServiceEngineFake, repository *syncServiceRepositoryFake) { repository.applyErr = errors.New("repository token=apply-secret") },
			wantEngineCalls: 1, wantPrepare: 1, wantApply: 1,
		},
		{
			name: "resolve error", configure: func(engine *syncServiceEngineFake, repository *syncServiceRepositoryFake) { engine.changeErr = mail.ErrEngineUnavailable; repository.resolveErr = errors.New("repository token=resolve-secret") },
			wantEngineCalls: 1, wantResolve: 1,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validMailboxSyncInput(t)
			input.Page.Change.Bounds.Requested = 17
			engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change}
			trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: []uuid.UUID{syncServiceUUID(testCase.name + "-received"), syncServiceUUID(testCase.name + "-changed")}}
			repository := &syncServiceRepositoryFake{claim: syncServiceClaim(syncServiceFence(mail.MailboxSyncPhaseIncremental), input.EventCutoff), resolveResult: mail.MailboxSyncResolveResult{Resolved: true, Phase: mail.MailboxSyncPhaseIncremental}}
			repository.prepare = func(request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
				return mail.MailboxSyncPreparedPage{Prepared: true, Input: syncServicePreparedInputFromCandidates(t, input, request)}, nil
			}
			testCase.configure(engine, repository)
			service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
			if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
			result, err := service.Step(context.Background(), syncServiceAccountID)
			if result != (mail.MailboxSyncStepResult{}) || !errors.Is(err, mail.ErrMailboxSyncServiceUnavailable) || strings.Contains(errString(err), "secret") || repository.prepareCalls != testCase.wantPrepare || repository.applyCalls != testCase.wantApply || repository.resolveCalls != testCase.wantResolve || engine.changeCalls != testCase.wantEngineCalls {
				t.Fatalf("repository failure escaped or crossed a forbidden boundary: result=%#v err=%v prepare/apply/resolve/change=%d/%d/%d/%d", result, err, repository.prepareCalls, repository.applyCalls, repository.resolveCalls, engine.changeCalls)
			}
		})
	}
}

func TestMailboxSyncServiceStep_InvalidJitterRejectsWithoutApply(t *testing.T) {
	input := validMailboxSyncInput(t)
	input.Page.Change.Bounds.Requested = 17
	engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change}
	trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: []uuid.UUID{syncServiceUUID("jitter-received"), syncServiceUUID("jitter-changed")}, jitter: 10 * time.Second}
	repository := &syncServiceRepositoryFake{claim: syncServiceClaim(syncServiceFence(mail.MailboxSyncPhaseIncremental), input.EventCutoff), resolveResult: mail.MailboxSyncResolveResult{Resolved: true, Phase: mail.MailboxSyncPhaseIncremental}}
	repository.prepare = func(request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
		return mail.MailboxSyncPreparedPage{Prepared: true, Input: syncServicePreparedInputFromCandidates(t, input, request)}, nil
	}
	service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
	if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
	result, err := service.Step(context.Background(), syncServiceAccountID)
	if result != (mail.MailboxSyncStepResult{}) || !errors.Is(err, mail.ErrMailboxSyncServiceInvalid) || repository.applyCalls != 0 || repository.resolveCalls != 1 || repository.resolveRequest.Failure != mail.MailboxSyncFailureRejected || trace.jitterCalls != 1 || trace.jitterInput != 10*time.Second {
		t.Fatalf("invalid jitter crossed apply boundary: result=%#v err=%v apply/resolve/jitter=%d/%d/%d", result, err, repository.applyCalls, repository.resolveCalls, trace.jitterCalls)
	}
}

func TestMailboxSyncServiceStep_CutoffAndConfirmedOwnerNeverSelectReceivedEvent(t *testing.T) {
	for _, testCase := range []struct {
		name string
		input func(*testing.T) mail.MailboxSyncPolicyInput
		wantEvents int
	}{
		{
			name: "before cutoff", input: func(t *testing.T) mail.MailboxSyncPolicyInput {
				input := validMailboxSyncInput(t)
				received := input.EventCutoff.Add(-time.Second)
				input.Page.Change.CreatedOrUpdated[0].ReceivedAt = &received
				return input
			},
		},
		{
			name: "equal cutoff", input: func(t *testing.T) mail.MailboxSyncPolicyInput {
				input := validMailboxSyncInput(t)
				received := input.EventCutoff
				input.Page.Change.CreatedOrUpdated[0].ReceivedAt = &received
				return input
			},
		},
		{
			name: "after cutoff", input: func(t *testing.T) mail.MailboxSyncPolicyInput { return validMailboxSyncInput(t) }, wantEvents: 1,
		},
		{
			name: "confirmed owner", input: confirmedOwnerSyncPolicyInput, wantEvents: 0,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := testCase.input(t)
			input.Page.Change.Bounds.Requested = 17
			engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change}
			trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: []uuid.UUID{syncServiceUUID(testCase.name + "-received"), syncServiceUUID(testCase.name + "-changed")}}
			repository := &syncServiceRepositoryFake{claim: syncServiceClaim(syncServiceFence(mail.MailboxSyncPhaseIncremental), input.EventCutoff)}
			repository.prepare = func(request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
				prepared := syncServicePreparedInputFromCandidates(t, input, request)
				if testCase.wantEvents == 0 {
					prepared.Messages[0].ReceivedEventID = uuid.Nil
				}
				return mail.MailboxSyncPreparedPage{Prepared: true, Input: prepared}, nil
			}
			repository.apply = func(request mail.MailboxSyncApplyRequest) (mail.MailboxSyncApplyResult, error) {
				if len(request.Result.Events) != testCase.wantEvents || request.Result.Advance.Change == nil || request.Result.Advance.Change.FromCursor != "cursor-a" {
					t.Fatalf("cutoff selection produced events=%d advance=%#v", len(request.Result.Events), request.Result.Advance)
				}
				return mail.MailboxSyncApplyResult{Applied: true, Phase: mail.MailboxSyncPhaseIncremental, NextAttemptAt: request.NextAttemptAt}, nil
			}
			service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
			if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
			result, err := service.Step(context.Background(), syncServiceAccountID)
			if err != nil || result.Outcome != mail.MailboxSyncStepApplied || repository.applyCalls != 1 {
				t.Fatalf("cutoff service result=%#v err=%v apply=%d", result, err, repository.applyCalls)
			}
		})
	}
}

func TestMailboxSyncServiceStep_PersistsExactTerminalAndNonterminalPhaseTransitions(t *testing.T) {
	for _, testCase := range []struct {
		name string
		input func(*testing.T) mail.MailboxSyncPolicyInput
		fence func() mail.MailboxSyncFence
		wantPhase mail.MailboxSyncPhase
		wantChangeNext string
		wantEnumerationComplete bool
	}{
		{
			name: "incremental nonterminal", input: validMailboxSyncInput,
			fence: func() mail.MailboxSyncFence { return syncServiceFence(mail.MailboxSyncPhaseIncremental) }, wantPhase: mail.MailboxSyncPhaseIncremental, wantChangeNext: "cursor-b",
		},
		{
			name: "incremental terminal keeps from cursor", input: func(t *testing.T) mail.MailboxSyncPolicyInput {
				input := validMailboxSyncInput(t)
				input.Page.Change.NextCursor = ""
				return input
			},
			fence: func() mail.MailboxSyncFence { return syncServiceFence(mail.MailboxSyncPhaseIncremental) }, wantPhase: mail.MailboxSyncPhaseIncremental, wantChangeNext: "",
		},
		{
			name: "catchup nonterminal", input: validMailboxSyncInput,
			fence: func() mail.MailboxSyncFence { return syncServiceFence(mail.MailboxSyncPhaseCatchUp) }, wantPhase: mail.MailboxSyncPhaseCatchUp, wantChangeNext: "cursor-b",
		},
		{
			name: "catchup terminal finalizes incremental", input: func(t *testing.T) mail.MailboxSyncPolicyInput {
				input := validMailboxSyncInput(t)
				input.Page.Change.NextCursor = ""
				return input
			},
			fence: func() mail.MailboxSyncFence { return syncServiceFence(mail.MailboxSyncPhaseCatchUp) }, wantPhase: mail.MailboxSyncPhaseIncremental, wantChangeNext: "",
		},
		{
			name: "enumeration checkpoint", input: func(t *testing.T) mail.MailboxSyncPolicyInput {
				input := mailboxSyncInputWithEnumeration(t, 1)
				input.Page.Enumeration.Complete = false
				input.Page.Enumeration.NextCheckpoint = "checkpoint-b"
				return input
			},
			fence: func() mail.MailboxSyncFence {
				fence := syncServiceFence(mail.MailboxSyncPhaseEnumeration)
				fence.Cursor, fence.BaseCursor, fence.Checkpoint = "", "cursor-a", "checkpoint-a"
				return fence
			},
			wantPhase: mail.MailboxSyncPhaseEnumeration,
		},
		{
			name: "enumeration complete enters catchup", input: func(t *testing.T) mail.MailboxSyncPolicyInput { return mailboxSyncInputWithEnumeration(t, 1) },
			fence: func() mail.MailboxSyncFence {
				fence := syncServiceFence(mail.MailboxSyncPhaseEnumeration)
				fence.Cursor, fence.BaseCursor, fence.Checkpoint = "", "cursor-a", "checkpoint-a"
				return fence
			},
			wantPhase: mail.MailboxSyncPhaseCatchUp, wantEnumerationComplete: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := testCase.input(t)
			if input.Page.Change != nil { input.Page.Change.Bounds.Requested = 17 }
			if input.Page.Enumeration != nil { input.Page.Enumeration.Bounds.Requested = 17 }
			fence := testCase.fence()
			engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}}
			if input.Page.Change != nil { engine.change = *input.Page.Change } else { engine.enumeration = *input.Page.Enumeration }
			trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: []uuid.UUID{syncServiceUUID(testCase.name + "-received"), syncServiceUUID(testCase.name + "-changed")}}
			repository := &syncServiceRepositoryFake{claim: syncServiceClaim(fence, input.EventCutoff)}
			repository.prepare = func(request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
				return mail.MailboxSyncPreparedPage{Prepared: true, Input: syncServicePreparedInputFromCandidates(t, input, request)}, nil
			}
			repository.apply = func(request mail.MailboxSyncApplyRequest) (mail.MailboxSyncApplyResult, error) {
				if request.Result.Advance.Change != nil {
					if request.Result.Advance.Change.FromCursor != "cursor-a" || request.Result.Advance.Change.NextCursor != testCase.wantChangeNext {
						t.Fatalf("change advance = %#v", request.Result.Advance.Change)
					}
				} else if request.Result.Advance.Enumeration == nil || request.Result.Advance.Enumeration.Complete != testCase.wantEnumerationComplete {
					t.Fatalf("enumeration advance = %#v", request.Result.Advance.Enumeration)
				}
				return mail.MailboxSyncApplyResult{Applied: true, Phase: testCase.wantPhase, NextAttemptAt: request.NextAttemptAt}, nil
			}
			service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
			if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
			result, err := service.Step(context.Background(), syncServiceAccountID)
			if err != nil || result.Outcome != mail.MailboxSyncStepApplied || engine.listCalls != 1 || engine.changeCalls+engine.enumerateCalls != 1 {
				t.Fatalf("terminal transition result=%#v err=%v list/change/enumerate=%d/%d/%d", result, err, engine.listCalls, engine.changeCalls, engine.enumerateCalls)
			}
		})
	}
}

func TestMailboxSyncServiceStep_DoesNoEngineOrCallbackWorkInsideRepositoryBoundaries(t *testing.T) {
	input := validMailboxSyncInput(t)
	input.Page.Change.Bounds.Requested = 17
	repository := &syncServiceRepositoryFake{claim: syncServiceClaim(syncServiceFence(mail.MailboxSyncPhaseIncremental), input.EventCutoff)}
	engine := &syncServiceEngineFake{repository: repository, folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change}
	trace := &syncServiceCallbackTrace{repository: repository, leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: []uuid.UUID{syncServiceUUID("boundary-received"), syncServiceUUID("boundary-changed")}}
	repository.prepare = func(request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
		return mail.MailboxSyncPreparedPage{Prepared: true, Input: syncServicePreparedInputFromCandidates(t, input, request)}, nil
	}
	repository.apply = func(request mail.MailboxSyncApplyRequest) (mail.MailboxSyncApplyResult, error) {
		return mail.MailboxSyncApplyResult{Applied: true, Phase: mail.MailboxSyncPhaseIncremental, NextAttemptAt: request.NextAttemptAt}, nil
	}
	service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
	if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
	result, err := service.Step(context.Background(), syncServiceAccountID)
	if err != nil || result.Outcome != mail.MailboxSyncStepApplied || repository.externalDuringTransaction {
		t.Fatalf("external work crossed repository boundary: result=%#v err=%v marked=%v", result, err, repository.externalDuringTransaction)
	}
}

func TestMailboxSyncServiceStep_AuthoritativeTransactionTimeExpiresAfterEngineWork(t *testing.T) {
	input := validMailboxSyncInput(t)
	input.Page.Change.Bounds.Requested = 17
	repository := newSyncServiceClockRepository(input)
	engine := &syncServiceClockEngine{repository: repository, folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change, expireAfterChange: true}
	service, err := mail.NewMailboxSyncService(engine, repository, syncServiceClockConfig(repository, []uuid.UUID{syncServiceLeaseID}, []uuid.UUID{syncServiceUUID("expired-received"), syncServiceUUID("expired-changed")}))
	if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
	result, err := service.Step(context.Background(), syncServiceAccountID)
	if err != nil || result.Outcome != mail.MailboxSyncStepLeaseLost || !repository.claimRequest.Now.Equal(syncServiceNow) || repository.prepareCalls != 1 || repository.applyCalls != 0 || repository.resolveCalls != 0 || repository.effectProjection != 0 || repository.effectCounts != 0 || repository.effectOutbox != 0 || repository.externalDuringTransaction {
		t.Fatalf("authoritative expiry was rescued by stale caller time: result=%#v err=%v claim=%#v prepare/apply/resolve=%d/%d/%d effects=%d/%d/%d external=%v", result, err, repository.claimRequest, repository.prepareCalls, repository.applyCalls, repository.resolveCalls, repository.effectProjection, repository.effectCounts, repository.effectOutbox, repository.externalDuringTransaction)
	}
}

func TestMailboxSyncClockRepository_ExpiredOrEqualTransactionTimeRejectsEveryMutation(t *testing.T) {
	input := validMailboxSyncInput(t)
	repository := newSyncServiceClockRepository(input)
	repository.transactionNow = repository.leaseUntil
	fence := repository.fence
	staleNow := repository.leaseUntil.Add(-time.Hour)
	prepared, prepareErr := repository.PrepareMailboxSyncPage(context.Background(), mail.MailboxSyncPrepareRequest{Fence: fence})
	applied, applyErr := repository.ApplyMailboxSyncPage(context.Background(), mail.MailboxSyncApplyRequest{Fence: fence, Now: staleNow, NextAttemptAt: staleNow.Add(time.Second)})
	resolved, resolveErr := repository.ResolveMailboxSyncStep(context.Background(), mail.MailboxSyncResolveRequest{Fence: fence, Failure: mail.MailboxSyncFailureTransient, Now: staleNow, NextAttemptAt: staleNow.Add(time.Second)})
	if prepareErr != nil || applyErr != nil || resolveErr != nil || prepared.Prepared || applied.Applied || applied.Replayed || resolved.Resolved || repository.effectProjection != 0 || repository.effectCounts != 0 || repository.effectOutbox != 0 || repository.resolveEffects != 0 || repository.externalDuringTransaction {
		t.Fatalf("equal transaction time mutated state or trusted request.Now: prepared=%#v applied=%#v resolved=%#v effects=%d/%d/%d/%d external=%v", prepared, applied, resolved, repository.effectProjection, repository.effectCounts, repository.effectOutbox, repository.resolveEffects, repository.externalDuringTransaction)
	}
}

func TestMailboxSyncClockRepository_TakeoverEpochRejectsLateFenceWithoutEffects(t *testing.T) {
	input := validMailboxSyncInput(t)
	repository := newSyncServiceClockRepository(input)
	oldFence := repository.fence
	repository.takeOver(syncServiceUUID("takeover-lease"))
	applied, err := repository.ApplyMailboxSyncPage(context.Background(), mail.MailboxSyncApplyRequest{Fence: oldFence, Now: syncServiceNow, NextAttemptAt: syncServiceNow.Add(time.Second)})
	if err != nil || applied.Applied || applied.Replayed || repository.fence.LeaseEpoch != oldFence.LeaseEpoch+1 || repository.fence.MutationEpoch != oldFence.MutationEpoch+1 || repository.fence.LeaseID == oldFence.LeaseID || repository.effectProjection != 0 || repository.effectCounts != 0 || repository.effectOutbox != 0 {
		t.Fatalf("late owner applied after takeover: applied=%#v err=%v old=%#v current=%#v effects=%d/%d/%d", applied, err, oldFence, repository.fence, repository.effectProjection, repository.effectCounts, repository.effectOutbox)
	}
}

func TestMailboxSyncServiceStep_CrashBeforeApplyThenSameFenceReplayAvoidsDuplicateEffects(t *testing.T) {
	input := validMailboxSyncInput(t)
	input.Page.Change.Bounds.Requested = 17
	repository := newSyncServiceClockRepository(input)
	repository.crashBeforeApply = true
	engine := &syncServiceClockEngine{repository: repository, folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change}
	service, err := mail.NewMailboxSyncService(engine, repository, syncServiceClockConfig(repository, []uuid.UUID{syncServiceLeaseID, syncServiceLeaseID, syncServiceLeaseID}, []uuid.UUID{syncServiceUUID("replay-received"), syncServiceUUID("replay-changed"), syncServiceUUID("replay-received"), syncServiceUUID("replay-changed"), syncServiceUUID("replay-received"), syncServiceUUID("replay-changed")}))
	if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
	crashed, err := service.Step(context.Background(), syncServiceAccountID)
	if err != nil || crashed.Outcome != mail.MailboxSyncStepLeaseLost || repository.committed || repository.effectProjection != 0 || repository.effectCounts != 0 || repository.effectOutbox != 0 {
		t.Fatalf("crash-before-apply persisted effects: result=%#v err=%v committed=%v effects=%d/%d/%d", crashed, err, repository.committed, repository.effectProjection, repository.effectCounts, repository.effectOutbox)
	}
	repository.crashBeforeApply = false
	applied, err := service.Step(context.Background(), syncServiceAccountID)
	if err != nil || applied.Outcome != mail.MailboxSyncStepApplied || !repository.committed || repository.lastApplyReplayed || repository.effectProjection != 1 || repository.effectCounts != 1 || repository.effectOutbox != 1 {
		t.Fatalf("same fence did not commit exactly once after crash: result=%#v err=%v replay=%v effects=%d/%d/%d", applied, err, repository.lastApplyReplayed, repository.effectProjection, repository.effectCounts, repository.effectOutbox)
	}
	projectionEffects, countEffects, outboxEffects := repository.effectProjection, repository.effectCounts, repository.effectOutbox
	replayed, err := service.Step(context.Background(), syncServiceAccountID)
	if err != nil || replayed.Outcome != mail.MailboxSyncStepApplied || !repository.lastApplyReplayed || repository.effectProjection != projectionEffects || repository.effectCounts != countEffects || repository.effectOutbox != outboxEffects || repository.externalDuringTransaction {
		t.Fatalf("same-fence replay duplicated effects or crossed transaction boundary: result=%#v err=%v replay=%v effects=%d/%d/%d external=%v", replayed, err, repository.lastApplyReplayed, repository.effectProjection, repository.effectCounts, repository.effectOutbox, repository.externalDuringTransaction)
	}
}

func TestMailboxSyncServiceStep_AmbiguousApplyRetriesAsSameFenceReplay(t *testing.T) {
	input := validMailboxSyncInput(t)
	input.Page.Change.Bounds.Requested = 17
	repository := newSyncServiceClockRepository(input)
	repository.ambiguousApplyOnce = true
	engine := &syncServiceClockEngine{repository: repository, folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change}
	service, err := mail.NewMailboxSyncService(engine, repository, syncServiceClockConfig(repository, []uuid.UUID{syncServiceLeaseID, syncServiceLeaseID}, []uuid.UUID{syncServiceUUID("ambiguous-received"), syncServiceUUID("ambiguous-changed"), syncServiceUUID("ambiguous-received"), syncServiceUUID("ambiguous-changed")}))
	if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
	first, err := service.Step(context.Background(), syncServiceAccountID)
	if first != (mail.MailboxSyncStepResult{}) || !errors.Is(err, mail.ErrMailboxSyncServiceUnavailable) || !repository.committed || repository.resolveCalls != 0 || repository.effectProjection != 1 || repository.effectCounts != 1 || repository.effectOutbox != 1 {
		t.Fatalf("ambiguous apply escaped or resolved over a possible commit: result=%#v err=%v committed=%v resolve=%d effects=%d/%d/%d", first, err, repository.committed, repository.resolveCalls, repository.effectProjection, repository.effectCounts, repository.effectOutbox)
	}
	retried, err := service.Step(context.Background(), syncServiceAccountID)
	if err != nil || retried.Outcome != mail.MailboxSyncStepApplied || !repository.lastApplyReplayed || repository.effectProjection != 1 || repository.effectCounts != 1 || repository.effectOutbox != 1 || repository.externalDuringTransaction {
		t.Fatalf("ambiguous retry was not an idempotent replay: result=%#v err=%v replay=%v effects=%d/%d/%d external=%v", retried, err, repository.lastApplyReplayed, repository.effectProjection, repository.effectCounts, repository.effectOutbox, repository.externalDuringTransaction)
	}
}

func TestMailboxSyncServiceStep_FolderFailureAndPageCapsStopBeforePrepare(t *testing.T) {
	t.Run("folder failure does not page", func(t *testing.T) {
		input := validMailboxSyncInput(t)
		engine := &syncServiceEngineFake{listErr: mail.ErrEngineUnavailable}
		repository := &syncServiceRepositoryFake{claim: syncServiceClaim(syncServiceFence(mail.MailboxSyncPhaseIncremental), input.EventCutoff), resolveResult: mail.MailboxSyncResolveResult{Resolved: true, Phase: mail.MailboxSyncPhaseIncremental}}
		trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}}
		service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
		if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
		result, err := service.Step(context.Background(), syncServiceAccountID)
		if err != nil || result.Outcome != mail.MailboxSyncStepRetryScheduled || engine.listCalls != 1 || engine.changeCalls != 0 || engine.enumerateCalls != 0 || repository.prepareCalls != 0 || repository.applyCalls != 0 || repository.resolveRequest.Failure != mail.MailboxSyncFailureTransient {
			t.Fatalf("folder failure calls = result %#v err %v list/change/enum/prepare/apply = %d/%d/%d/%d/%d", result, err, engine.listCalls, engine.changeCalls, engine.enumerateCalls, repository.prepareCalls, repository.applyCalls)
		}
	})

	for _, testCase := range []struct {
		name string
		phase mail.MailboxSyncPhase
		page func() (mail.EngineChangePage, mail.EngineEnumerationPage, time.Time)
		wantChanges int
		wantEnumeration int
	}{
		{
			name: "combined changes cap", phase: mail.MailboxSyncPhaseIncremental,
			page: func() (mail.EngineChangePage, mail.EngineEnumerationPage, time.Time) {
				input := validMailboxSyncInput(t)
				change := *input.Page.Change
				change.CreatedOrUpdated = make([]mail.EngineMessageProjection, mail.MaxEnginePageItems+1)
				change.DestroyedIDs = []string{}
				return change, mail.EngineEnumerationPage{}, input.EventCutoff
			}, wantChanges: 1,
		},
		{
			name: "enumeration cap", phase: mail.MailboxSyncPhaseEnumeration,
			page: func() (mail.EngineChangePage, mail.EngineEnumerationPage, time.Time) {
				input := mailboxSyncInputWithEnumeration(t, 1)
				enumeration := *input.Page.Enumeration
				enumeration.Messages = make([]mail.EngineMessageProjection, mail.MaxEnginePageItems+1)
				return mail.EngineChangePage{}, enumeration, input.EventCutoff
			}, wantEnumeration: 1,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			change, enumeration, cutoff := testCase.page()
			fence := syncServiceFence(testCase.phase)
			if testCase.phase == mail.MailboxSyncPhaseEnumeration { fence.Cursor = "" }
			engine := &syncServiceEngineFake{folders: []mail.EngineFolder{}, change: change, enumeration: enumeration}
			repository := &syncServiceRepositoryFake{claim: syncServiceClaim(fence, cutoff), resolveResult: mail.MailboxSyncResolveResult{Resolved: true, Phase: testCase.phase}}
			trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}}
			service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
			if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
			_, err = service.Step(context.Background(), syncServiceAccountID)
			if !errors.Is(err, mail.ErrMailboxSyncServiceTooLarge) || engine.changeCalls != testCase.wantChanges || engine.enumerateCalls != testCase.wantEnumeration || trace.eventCalls != 0 || repository.prepareCalls != 0 || repository.applyCalls != 0 || repository.resolveCalls != 1 || repository.resolveRequest.Failure != mail.MailboxSyncFailureRejected {
				t.Fatalf("cap handling err=%v change/enum/event/prepare/apply/resolve=%d/%d/%d/%d/%d/%d", err, engine.changeCalls, engine.enumerateCalls, trace.eventCalls, repository.prepareCalls, repository.applyCalls, repository.resolveCalls)
			}
		})
	}
}

func TestMailboxSyncServiceStep_EnumerationRestartSchedulesRebuild(t *testing.T) {
	input := mailboxSyncInputWithEnumeration(t, 1)
	fence := syncServiceFence(mail.MailboxSyncPhaseEnumeration)
	fence.Cursor = ""
	engine := &syncServiceEngineFake{folders: []mail.EngineFolder{}, enumerationErr: mail.ErrEngineRestartRequired}
	repository := &syncServiceRepositoryFake{claim: syncServiceClaim(fence, input.EventCutoff), resolveResult: mail.MailboxSyncResolveResult{Resolved: true, Phase: mail.MailboxSyncPhaseEnumeration}}
	trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}}
	service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
	if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
	result, err := service.Step(context.Background(), syncServiceAccountID)
	if err != nil || result.Outcome != mail.MailboxSyncStepRebuildScheduled || engine.changeCalls != 0 || engine.enumerateCalls != 1 || repository.resolveCalls != 1 || repository.resolveRequest.Failure != mail.MailboxSyncFailureRestartRequired {
		t.Fatalf("restart result=%#v err=%v changes/enumerate/resolve=%d/%d/%d", result, err, engine.changeCalls, engine.enumerateCalls, repository.resolveCalls)
	}
}

func TestMailboxSyncServiceStep_RejectsBadCandidateIDsBeforePrepare(t *testing.T) {
	for _, ids := range [][]uuid.UUID{{uuid.Nil}, {syncServiceUUID("duplicate"), syncServiceUUID("duplicate")}} {
		input := validMailboxSyncInput(t)
		input.Page.Change.Bounds.Requested = 17
		engine := &syncServiceEngineFake{folders: []mail.EngineFolder{{AccountID: "engine-account", FolderID: "inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox}}, change: *input.Page.Change}
		trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: ids}
		repository := &syncServiceRepositoryFake{claim: syncServiceClaim(syncServiceFence(mail.MailboxSyncPhaseIncremental), input.EventCutoff), resolveResult: mail.MailboxSyncResolveResult{Resolved: true, Phase: mail.MailboxSyncPhaseIncremental}}
		service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
		if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
		_, err = service.Step(context.Background(), syncServiceAccountID)
		if !errors.Is(err, mail.ErrMailboxSyncServiceInvalid) || repository.prepareCalls != 0 || repository.applyCalls != 0 || repository.resolveCalls != 1 || repository.resolveRequest.Failure != mail.MailboxSyncFailureRejected {
			t.Fatalf("candidate validation did not fail closed before prepare: err=%v prepare/apply/resolve=%d/%d/%d", err, repository.prepareCalls, repository.applyCalls, repository.resolveCalls)
		}
	}
}

func TestMailboxSyncServiceStep_OrdersVolatileCandidatesBeforePrepare(t *testing.T) {
	fence := syncServiceFence(mail.MailboxSyncPhaseIncremental)
	page := mail.EngineChangePage{
		AccountID: "engine-account", FromCursor: "cursor-a", NextCursor: "cursor-b",
		Bounds: mail.EnginePageBounds{Requested: 17, Maximum: mail.MaxEnginePageItems, Returned: 4},
		CreatedOrUpdated: []mail.EngineMessageProjection{{MessageID: "live-z"}, {MessageID: "live-a"}}, DestroyedIDs: []string{"destroyed-z", "destroyed-a"},
	}
	ids := []uuid.UUID{syncServiceUUID("a-received"), syncServiceUUID("a-changed"), syncServiceUUID("z-received"), syncServiceUUID("z-changed"), syncServiceUUID("destroyed-a"), syncServiceUUID("destroyed-z")}
	trace := &syncServiceCallbackTrace{leaseIDs: []uuid.UUID{syncServiceLeaseID}, eventIDs: ids}
	engine := &syncServiceEngineFake{folders: []mail.EngineFolder{}, change: page}
	repository := &syncServiceRepositoryFake{claim: syncServiceClaim(fence, syncServiceNow.Add(-time.Second))}
	repository.prepare = func(request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
		want := []mail.MailboxSyncEventCandidate{
			{EngineMessageID: "live-a", ReceivedEventID: ids[0], ChangedEventID: ids[1]},
			{EngineMessageID: "live-z", ReceivedEventID: ids[2], ChangedEventID: ids[3]},
			{EngineMessageID: "destroyed-a", ChangedEventID: ids[4], Destroyed: true},
			{EngineMessageID: "destroyed-z", ChangedEventID: ids[5], Destroyed: true},
		}
		if !reflect.DeepEqual(request.EventCandidates, want) { t.Fatalf("candidates = %#v, want %#v", request.EventCandidates, want) }
		return mail.MailboxSyncPreparedPage{}, nil
	}
	service, err := mail.NewMailboxSyncService(engine, repository, trace.config())
	if err != nil { t.Fatalf("NewMailboxSyncService() error = %v", err) }
	result, err := service.Step(context.Background(), syncServiceAccountID)
	if err != nil || result.Outcome != mail.MailboxSyncStepLeaseLost || trace.eventCalls != len(ids) || repository.applyCalls != 0 {
		t.Fatalf("Step() = %#v, %v; candidate ordering must remain volatile before false prepare", result, err)
	}
}

type syncServiceEngineFake struct {
	mail.MailEngine
	repository *syncServiceRepositoryFake
	folders []mail.EngineFolder
	listErr error
	change mail.EngineChangePage
	changeErr error
	enumeration mail.EngineEnumerationPage
	enumerationErr error
	listCalls int
	changeCalls int
	enumerateCalls int
	changeRequest mail.EngineChangeRequest
	enumerationRequest mail.EngineEnumerationRequest
}

func (engine *syncServiceEngineFake) ListFolders(context.Context, string) ([]mail.EngineFolder, error) {
	engine.markExternalCall()
	engine.listCalls++
	if engine.listErr != nil { return nil, engine.listErr }
	return append([]mail.EngineFolder(nil), engine.folders...), nil
}

func (engine *syncServiceEngineFake) Changes(_ context.Context, request mail.EngineChangeRequest) (mail.EngineChangePage, error) {
	engine.markExternalCall()
	engine.changeCalls++
	engine.changeRequest = request
	if engine.changeErr != nil { return mail.EngineChangePage{}, engine.changeErr }
	return engine.change, nil
}

func (engine *syncServiceEngineFake) EnumerateProjection(_ context.Context, request mail.EngineEnumerationRequest) (mail.EngineEnumerationPage, error) {
	engine.markExternalCall()
	engine.enumerateCalls++
	engine.enumerationRequest = request
	if engine.enumerationErr != nil { return mail.EngineEnumerationPage{}, engine.enumerationErr }
	return engine.enumeration, nil
}

func (engine *syncServiceEngineFake) markExternalCall() {
	if engine.repository != nil && engine.repository.inTransaction {
		engine.repository.externalDuringTransaction = true
	}
}

type syncServiceRepositoryFake struct {
	claim mail.MailboxSyncClaimResult
	claimErr error
	prepare func(mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error)
	prepareErr error
	apply func(mail.MailboxSyncApplyRequest) (mail.MailboxSyncApplyResult, error)
	applyErr error
	resolveResult mail.MailboxSyncResolveResult
	resolveErr error
	claimCalls int
	prepareCalls int
	applyCalls int
	resolveCalls int
	claimRequest mail.MailboxSyncClaimRequest
	prepareRequest mail.MailboxSyncPrepareRequest
	applyRequest mail.MailboxSyncApplyRequest
	resolveRequest mail.MailboxSyncResolveRequest
	inTransaction bool
	externalDuringTransaction bool
}

func (repository *syncServiceRepositoryFake) ClaimMailboxSyncStep(_ context.Context, request mail.MailboxSyncClaimRequest) (mail.MailboxSyncClaimResult, error) {
	repository.inTransaction = true
	defer func() { repository.inTransaction = false }()
	repository.claimCalls++
	repository.claimRequest = request
	if repository.claimErr != nil { return mail.MailboxSyncClaimResult{}, repository.claimErr }
	return repository.claim, nil
}

func (repository *syncServiceRepositoryFake) PrepareMailboxSyncPage(_ context.Context, request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
	repository.inTransaction = true
	defer func() { repository.inTransaction = false }()
	repository.prepareCalls++
	repository.prepareRequest = request
	if repository.prepareErr != nil { return mail.MailboxSyncPreparedPage{}, repository.prepareErr }
	if repository.prepare == nil { return mail.MailboxSyncPreparedPage{}, nil }
	return repository.prepare(request)
}

func (repository *syncServiceRepositoryFake) ApplyMailboxSyncPage(_ context.Context, request mail.MailboxSyncApplyRequest) (mail.MailboxSyncApplyResult, error) {
	repository.inTransaction = true
	defer func() { repository.inTransaction = false }()
	repository.applyCalls++
	repository.applyRequest = request
	if repository.applyErr != nil { return mail.MailboxSyncApplyResult{}, repository.applyErr }
	if repository.apply == nil { return mail.MailboxSyncApplyResult{}, nil }
	return repository.apply(request)
}

func (repository *syncServiceRepositoryFake) ResolveMailboxSyncStep(_ context.Context, request mail.MailboxSyncResolveRequest) (mail.MailboxSyncResolveResult, error) {
	repository.inTransaction = true
	defer func() { repository.inTransaction = false }()
	repository.resolveCalls++
	repository.resolveRequest = request
	if repository.resolveErr != nil { return mail.MailboxSyncResolveResult{}, repository.resolveErr }
	result := repository.resolveResult
	if result.Resolved && result.NextAttemptAt.IsZero() { result.NextAttemptAt = request.NextAttemptAt }
	return result, nil
}

type syncServiceClockRepository struct {
	input        mail.MailboxSyncPolicyInput
	fence        mail.MailboxSyncFence
	engineAccount string
	eventCutoff  time.Time
	leaseUntil   time.Time
	transactionNow time.Time
	committed    bool
	storedResult mail.MailboxSyncPolicyResult
	storedNextAttemptAt time.Time
	crashBeforeApply bool
	ambiguousApplyOnce bool
	ambiguousReturned bool
	lastApplyReplayed bool
	claimCalls   int
	prepareCalls int
	applyCalls   int
	resolveCalls int
	resolveEffects int
	effectProjection int
	effectCounts int
	effectOutbox int
	claimRequest mail.MailboxSyncClaimRequest
	prepareRequest mail.MailboxSyncPrepareRequest
	applyRequest mail.MailboxSyncApplyRequest
	resolveRequest mail.MailboxSyncResolveRequest
	inTransaction bool
	externalDuringTransaction bool
}

func newSyncServiceClockRepository(input mail.MailboxSyncPolicyInput) *syncServiceClockRepository {
	return &syncServiceClockRepository{
		input: input, fence: syncServiceFence(mail.MailboxSyncPhaseIncremental), engineAccount: "engine-account", eventCutoff: input.EventCutoff,
		leaseUntil: syncServiceNow.Add(time.Minute), transactionNow: syncServiceNow,
	}
}

func (repository *syncServiceClockRepository) ClaimMailboxSyncStep(_ context.Context, request mail.MailboxSyncClaimRequest) (mail.MailboxSyncClaimResult, error) {
	repository.inTransaction = true
	defer func() { repository.inTransaction = false }()
	repository.claimCalls++
	repository.claimRequest = request
	if request.AccountID != repository.fence.AccountID || request.LeaseID != repository.fence.LeaseID || !repository.transactionNow.Before(repository.leaseUntil) {
		return mail.MailboxSyncClaimResult{NextAttemptAt: syncServiceNow.Add(time.Minute)}, nil
	}
	return mail.MailboxSyncClaimResult{Claimed: true, Fence: repository.fence, EngineAccountID: repository.engineAccount, EventCutoff: repository.eventCutoff, NextAttemptAt: syncServiceNow.Add(time.Minute)}, nil
}

func (repository *syncServiceClockRepository) PrepareMailboxSyncPage(_ context.Context, request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPreparedPage, error) {
	repository.inTransaction = true
	defer func() { repository.inTransaction = false }()
	repository.prepareCalls++
	repository.prepareRequest = request
	if !repository.mutationAllowed(request.Fence) {
		return mail.MailboxSyncPreparedPage{}, nil
	}
	input, ok := syncServiceClockPreparedInput(repository.input, request)
	if !ok {
		return mail.MailboxSyncPreparedPage{}, nil
	}
	return mail.MailboxSyncPreparedPage{Prepared: true, Input: input}, nil
}

func (repository *syncServiceClockRepository) ApplyMailboxSyncPage(_ context.Context, request mail.MailboxSyncApplyRequest) (mail.MailboxSyncApplyResult, error) {
	repository.inTransaction = true
	defer func() { repository.inTransaction = false }()
	repository.applyCalls++
	repository.applyRequest = request
	repository.lastApplyReplayed = false
	if !repository.mutationAllowed(request.Fence) {
		return mail.MailboxSyncApplyResult{}, nil
	}
	if repository.committed {
		if reflect.DeepEqual(request.Result, repository.storedResult) && request.NextAttemptAt.Equal(repository.storedNextAttemptAt) {
			repository.lastApplyReplayed = true
			return mail.MailboxSyncApplyResult{Applied: true, Replayed: true, Phase: repository.fence.Phase, NextAttemptAt: repository.storedNextAttemptAt}, nil
		}
		return mail.MailboxSyncApplyResult{}, nil
	}
	if repository.crashBeforeApply {
		return mail.MailboxSyncApplyResult{}, nil
	}
	repository.committed = true
	repository.storedResult = request.Result
	repository.storedNextAttemptAt = request.NextAttemptAt
	repository.effectProjection++
	repository.effectCounts++
	repository.effectOutbox += len(request.Result.Events)
	if repository.ambiguousApplyOnce && !repository.ambiguousReturned {
		repository.ambiguousReturned = true
		return mail.MailboxSyncApplyResult{}, errors.New("ambiguous transaction outcome")
	}
	return mail.MailboxSyncApplyResult{Applied: true, Phase: repository.fence.Phase, NextAttemptAt: request.NextAttemptAt}, nil
}

func (repository *syncServiceClockRepository) ResolveMailboxSyncStep(_ context.Context, request mail.MailboxSyncResolveRequest) (mail.MailboxSyncResolveResult, error) {
	repository.inTransaction = true
	defer func() { repository.inTransaction = false }()
	repository.resolveCalls++
	repository.resolveRequest = request
	if !repository.mutationAllowed(request.Fence) {
		return mail.MailboxSyncResolveResult{}, nil
	}
	repository.resolveEffects++
	return mail.MailboxSyncResolveResult{Resolved: true, Phase: repository.fence.Phase, NextAttemptAt: request.NextAttemptAt}, nil
}

func (repository *syncServiceClockRepository) mutationAllowed(fence mail.MailboxSyncFence) bool {
	return repository.transactionNow.Before(repository.leaseUntil) && fence == repository.fence
}

func (repository *syncServiceClockRepository) takeOver(leaseID uuid.UUID) {
	repository.fence.LeaseID = leaseID
	repository.fence.LeaseEpoch++
	repository.fence.MutationEpoch++
}

func syncServiceClockPreparedInput(source mail.MailboxSyncPolicyInput, request mail.MailboxSyncPrepareRequest) (mail.MailboxSyncPolicyInput, bool) {
	input := source
	input.Page = request.Page
	input.Messages = append([]mail.MailboxSyncPolicyMessage(nil), source.Messages...)
	input.Destroyed = append([]mail.MailboxSyncPolicyDestroyed(nil), source.Destroyed...)
	input.SyncEpoch = request.Fence.SyncEpoch
	input.Event.OccurredAt = request.OccurredAt
	input.Event.CorrelationID = request.CorrelationID
	candidates := make(map[string]mail.MailboxSyncEventCandidate, len(request.EventCandidates))
	for _, candidate := range request.EventCandidates { candidates[candidate.EngineMessageID] = candidate }
	for index := range input.Messages {
		candidate, found := candidates[input.Messages[index].EngineMessageID]
		if !found || candidate.Destroyed { return mail.MailboxSyncPolicyInput{}, false }
		input.Messages[index].ReceivedEventID = candidate.ReceivedEventID
		input.Messages[index].ChangedEvent = mail.MailboxSyncPolicyEventReservation{}
	}
	for index := range input.Destroyed {
		candidate, found := candidates[input.Destroyed[index].EngineMessageID]
		if !found || !candidate.Destroyed { return mail.MailboxSyncPolicyInput{}, false }
		input.Destroyed[index].ChangedEvent = mail.MailboxSyncPolicyEventReservation{ID: candidate.ChangedEventID}
	}
	return input, true
}

type syncServiceClockEngine struct {
	mail.MailEngine
	repository *syncServiceClockRepository
	folders []mail.EngineFolder
	change mail.EngineChangePage
	expireAfterChange bool
	listCalls int
	changeCalls int
}

func (engine *syncServiceClockEngine) ListFolders(context.Context, string) ([]mail.EngineFolder, error) {
	engine.markExternalCall()
	engine.listCalls++
	return append([]mail.EngineFolder(nil), engine.folders...), nil
}

func (engine *syncServiceClockEngine) Changes(_ context.Context, request mail.EngineChangeRequest) (mail.EngineChangePage, error) {
	engine.markExternalCall()
	engine.changeCalls++
	if engine.expireAfterChange { engine.repository.transactionNow = engine.repository.leaseUntil }
	return engine.change, nil
}

func (engine *syncServiceClockEngine) EnumerateProjection(context.Context, mail.EngineEnumerationRequest) (mail.EngineEnumerationPage, error) {
	engine.markExternalCall()
	return mail.EngineEnumerationPage{}, mail.ErrEngineRejectedPolicy
}

func (engine *syncServiceClockEngine) markExternalCall() {
	if engine.repository.inTransaction { engine.repository.externalDuringTransaction = true }
}

func syncServiceClockConfig(repository *syncServiceClockRepository, leaseIDs, eventIDs []uuid.UUID) mail.MailboxSyncServiceConfig {
	leaseIndex, eventIndex := 0, 0
	markExternal := func() { if repository.inTransaction { repository.externalDuringTransaction = true } }
	return mail.MailboxSyncServiceConfig{
		Now: func() time.Time { markExternal(); return syncServiceNow },
		NewLeaseID: func() uuid.UUID { markExternal(); value := leaseIDs[leaseIndex]; leaseIndex++; return value },
		NewEventID: func() uuid.UUID { markExternal(); value := eventIDs[eventIndex]; eventIndex++; return value },
		Jitter: func(time.Duration) time.Duration { markExternal(); return 0 },
		LeaseDuration: time.Minute, PollInterval: 10 * time.Second, RetryInterval: 20 * time.Second, PageSize: 17,
	}
}

type syncServiceCallbackTrace struct {
	repository *syncServiceRepositoryFake
	leaseIDs []uuid.UUID
	eventIDs []uuid.UUID
	jitter time.Duration
	nowCalls int
	leaseCalls int
	eventCalls int
	jitterCalls int
	jitterInput time.Duration
}

func (trace *syncServiceCallbackTrace) config() mail.MailboxSyncServiceConfig {
	return mail.MailboxSyncServiceConfig{
		Now: func() time.Time { trace.markExternalCall(); trace.nowCalls++; return syncServiceNow },
		NewLeaseID: func() uuid.UUID { trace.markExternalCall(); value := trace.leaseIDs[trace.leaseCalls]; trace.leaseCalls++; return value },
		NewEventID: func() uuid.UUID { trace.markExternalCall(); value := trace.eventIDs[trace.eventCalls]; trace.eventCalls++; return value },
		Jitter: func(interval time.Duration) time.Duration { trace.markExternalCall(); trace.jitterCalls++; trace.jitterInput = interval; return trace.jitter },
		LeaseDuration: time.Minute, PollInterval: 10 * time.Second, RetryInterval: 20 * time.Second, PageSize: 17,
	}
}

func (trace *syncServiceCallbackTrace) markExternalCall() {
	if trace.repository != nil && trace.repository.inTransaction {
		trace.repository.externalDuringTransaction = true
	}
}

func syncServiceFence(phase mail.MailboxSyncPhase) mail.MailboxSyncFence {
	return mail.MailboxSyncFence{AccountID: syncServiceAccountID, LeaseID: syncServiceLeaseID, LeaseEpoch: 2, SyncEpoch: 7, MutationEpoch: 3, Phase: phase, Cursor: "cursor-a"}
}

func syncServiceClaim(fence mail.MailboxSyncFence, cutoff time.Time) mail.MailboxSyncClaimResult {
	return mail.MailboxSyncClaimResult{Claimed: true, Fence: fence, EngineAccountID: "engine-account", EventCutoff: cutoff, NextAttemptAt: syncServiceNow.Add(time.Minute)}
}

// syncServicePreparedInputFromCandidates is deliberately test-owned: it builds a
// normal Prepare response by selecting the volatile IDs the request allocated.
func syncServicePreparedInputFromCandidates(t *testing.T, source mail.MailboxSyncPolicyInput, request mail.MailboxSyncPrepareRequest) mail.MailboxSyncPolicyInput {
	t.Helper()
	input := source
	input.Page = request.Page
	input.SyncEpoch = request.Fence.SyncEpoch
	input.Event.OccurredAt = request.OccurredAt
	input.Event.CorrelationID = request.CorrelationID
	candidates := make(map[string]mail.MailboxSyncEventCandidate, len(request.EventCandidates))
	for _, candidate := range request.EventCandidates {
		candidates[candidate.EngineMessageID] = candidate
	}
	for index := range input.Messages {
		candidate, found := candidates[input.Messages[index].EngineMessageID]
		if !found || candidate.Destroyed {
			t.Fatalf("missing live candidate for %q", input.Messages[index].EngineMessageID)
		}
		input.Messages[index].ReceivedEventID = candidate.ReceivedEventID
		input.Messages[index].ChangedEvent = mail.MailboxSyncPolicyEventReservation{}
	}
	for index := range input.Destroyed {
		candidate, found := candidates[input.Destroyed[index].EngineMessageID]
		if !found || !candidate.Destroyed {
			t.Fatalf("missing destroyed candidate for %q", input.Destroyed[index].EngineMessageID)
		}
		input.Destroyed[index].ChangedEvent = mail.MailboxSyncPolicyEventReservation{ID: candidate.ChangedEventID}
	}
	return input
}

func requireMailboxSyncMethods(t *testing.T, value reflect.Type, want []string) {
	t.Helper()
	if value.NumMethod() != len(want) { t.Fatalf("%v methods = %d, want %d", value, value.NumMethod(), len(want)) }
	for _, name := range want {
		if _, found := value.MethodByName(name); !found { t.Fatalf("%v has no %s method", value, name) }
	}
}

func requireMailboxSyncFieldTypes(t *testing.T, value reflect.Type, want []reflect.Type) {
	t.Helper()
	if value.NumField() != len(want) { t.Fatalf("%v fields = %d, want %d", value, value.NumField(), len(want)) }
	for index, wantType := range want {
		field := value.Field(index)
		if field.Type != wantType { t.Fatalf("%v.%s type = %v, want %v", value, field.Name, field.Type, wantType) }
	}
}

func requireMailboxSyncMethodType(t *testing.T, value reflect.Type, name string, want reflect.Type) {
	t.Helper()
	method, found := value.MethodByName(name)
	if !found { t.Fatalf("%v has no %s", value, name) }
	if method.Type != want { t.Fatalf("%v.%s type = %v, want %v", value, name, method.Type, want) }
}

func syncServiceUUID(name string) uuid.UUID { return uuid.NewSHA1(uuid.NameSpaceURL, []byte("mailbox-sync-service-test:"+name)) }
func errString(err error) string { if err == nil { return "" }; return err.Error() }
