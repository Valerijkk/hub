package messenger_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

var _ messenger.Phase151PendingRecoveryRepository = (*messenger.Phase151ClaimedRecoveryCoordinatorAdapter)(nil)
var _ messenger.Phase151PendingRecoveryScanRepository = (*phase151ClaimedRecoveryCoordinatorScanner)(nil)
var _ messenger.Phase151ClaimedPostCreateRecoveryRepository = (*phase151ClaimedRecoveryCoordinatorEffects)(nil)

type phase151ClaimedRecoveryCoordinatorScanner struct {
	target       messenger.Phase151PendingRecoveryTarget
	scope        messenger.AuthorizedMessageProducerScope
	documents    []messenger.MessageDocument
	calls        []string
	intervals    []time.Duration
	intervalErr  error
	pageErr      error
	authorityErr error
	documentsErr error
	lastCursor   *messenger.Phase151RecoveryTargetCursor
	lastLimit    int
	lastScope    messenger.AuthorizedMessageProducerScope
	lastTarget   messenger.Phase151PendingRecoveryTarget
}

func (s *phase151ClaimedRecoveryCoordinatorScanner) TryAcquirePendingRecoveryInterval(
	_ context.Context,
	minimum time.Duration,
) (bool, error) {
	s.calls = append(s.calls, "interval")
	s.intervals = append(s.intervals, minimum)
	return s.intervalErr == nil, s.intervalErr
}

func (s *phase151ClaimedRecoveryCoordinatorScanner) ListPendingRecoveryTargets(
	_ context.Context,
	after *messenger.Phase151RecoveryTargetCursor,
	limit int,
) (messenger.Phase151PendingRecoveryTargetPage, error) {
	s.calls = append(s.calls, "targets")
	s.lastCursor = after
	s.lastLimit = limit
	if s.pageErr != nil {
		return messenger.Phase151PendingRecoveryTargetPage{}, s.pageErr
	}
	return messenger.Phase151PendingRecoveryTargetPage{
		Targets: []messenger.Phase151PendingRecoveryTarget{s.target},
	}, nil
}

func (s *phase151ClaimedRecoveryCoordinatorScanner) ReauthorizePendingRecoveryTarget(
	_ context.Context,
	target messenger.Phase151PendingRecoveryTarget,
) (messenger.AuthorizedMessageProducerScope, error) {
	s.calls = append(s.calls, "reauthorize")
	s.lastTarget = target
	return s.scope, s.authorityErr
}

func (s *phase151ClaimedRecoveryCoordinatorScanner) ListPendingPostCreateDocuments(
	_ context.Context,
	scope messenger.AuthorizedMessageProducerScope,
	target messenger.Phase151PendingRecoveryTarget,
	limit int,
) ([]messenger.MessageDocument, error) {
	s.calls = append(s.calls, "documents")
	s.lastScope = scope
	s.lastTarget = target
	s.lastLimit = limit
	return append([]messenger.MessageDocument(nil), s.documents...), s.documentsErr
}

type phase151ClaimedRecoveryCoordinatorEffects struct {
	calls     []string
	miss      map[string]bool
	followErr error
	owners    []uuid.UUID
	leases    []time.Duration
}

func (e *phase151ClaimedRecoveryCoordinatorEffects) ClaimPendingPostCreateRecovery(
	_ context.Context,
	scope messenger.AuthorizedMessageProducerScope,
	message *messenger.MessageDocument,
	ownerID uuid.UUID,
	now time.Time,
	lease time.Duration,
) (messenger.Phase151PostCreateRecoveryAcquisition, bool, error) {
	e.calls = append(e.calls, "claim:"+message.ID)
	e.owners = append(e.owners, ownerID)
	e.leases = append(e.leases, lease)
	if e.miss[message.ID] {
		return messenger.Phase151PostCreateRecoveryAcquisition{}, false, nil
	}
	leaseUntil := now.Add(lease).UTC()
	claimed := *message
	claimed.PostCreateRecoveryOwnerID = ownerID
	claimed.PostCreateRecoveryFence = 7
	claimed.PostCreateRecoveryLeaseUntil = &leaseUntil
	if message.PostCreateProducerProvenance != nil {
		provenance := *message.PostCreateProducerProvenance
		claimed.PostCreateProducerProvenance = &provenance
	}
	return messenger.Phase151PostCreateRecoveryAcquisition{
		Claim: messenger.Phase151PostCreateRecoveryClaim{
			TenantID: scope.TenantID, MessageID: message.ID, ChannelID: scope.ChannelID,
			OwnerID: ownerID, Fence: claimed.PostCreateRecoveryFence, LeaseUntil: leaseUntil,
		},
		Message: claimed,
	}, true, nil
}

func (e *phase151ClaimedRecoveryCoordinatorEffects) EnsureAutoThreadFollows(
	_ context.Context,
	_ messenger.AuthorizedMessageProducerScope,
	message *messenger.MessageDocument,
) error {
	e.calls = append(e.calls, "auto-follow:"+message.ID)
	return e.followErr
}

func (e *phase151ClaimedRecoveryCoordinatorEffects) ReconcileReplyCount(
	_ context.Context,
	_ messenger.AuthorizedMessageProducerScope,
	message *messenger.MessageDocument,
) error {
	e.calls = append(e.calls, "reply-count:"+message.ID)
	return nil
}

func (e *phase151ClaimedRecoveryCoordinatorEffects) CompleteClaimedPostCreateRecovery(
	_ context.Context,
	_ messenger.AuthorizedMessageProducerScope,
	message *messenger.MessageDocument,
	_ messenger.Phase151PostCreateRecoveryClaim,
	_ time.Time,
) error {
	e.calls = append(e.calls, "complete:"+message.ID)
	return nil
}

func (e *phase151ClaimedRecoveryCoordinatorEffects) PublishUnreadInvalidation(
	_ context.Context,
	_ messenger.AuthorizedMessageProducerScope,
	message *messenger.MessageDocument,
) error {
	e.calls = append(e.calls, "publish-unread:"+message.ID)
	return nil
}

func phase151ClaimedRecoveryCoordinatorFixture() (
	messenger.Phase151PendingRecoveryTarget,
	messenger.AuthorizedMessageProducerScope,
	messenger.MessageDocument,
	uuid.UUID,
	time.Time,
	time.Time,
	time.Duration,
) {
	targetAuthority, message, ownerID, claimNow, completedAt, lease, _ :=
		phase151ClaimedRecoveryRunnerFixture()
	target := messenger.Phase151PendingRecoveryTarget{
		TenantID:  targetAuthority.TenantID,
		ChannelID: targetAuthority.ChannelID,
		Kind:      "channel",
	}
	return target, targetAuthority, *message, ownerID, claimNow, completedAt, lease
}

func TestPhase151ClaimedRecoveryCoordinatorAdapterBindsBoundedScanToRunner(t *testing.T) {
	target, scope, message, ownerID, claimNow, completedAt, lease :=
		phase151ClaimedRecoveryCoordinatorFixture()
	scanner := &phase151ClaimedRecoveryCoordinatorScanner{
		target: target, scope: scope, documents: []messenger.MessageDocument{message},
	}
	effects := &phase151ClaimedRecoveryCoordinatorEffects{}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}
	adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(
		scanner, ownerID, lease, clock, effects,
	)

	err := messenger.ExportPhase151ExecutePendingPostCreateRecovery(
		context.Background(),
		adapter,
		messenger.Phase151PendingRecoveryConfig{
			MaxPages: 1, MaxTargets: 1, PendingScanLimit: 10, MinimumInterval: time.Minute,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, []string{"interval", "targets", "reauthorize", "documents"}, scanner.calls)
	assert.Equal(t, []string{
		"claim:" + message.ID,
		"auto-follow:" + message.ID,
		"complete:" + message.ID,
		"publish-unread:" + message.ID,
	}, effects.calls)
	assert.Equal(t, []uuid.UUID{ownerID}, effects.owners)
	assert.Equal(t, []time.Duration{lease}, effects.leases)
	assert.Equal(t, 2, clock.calls)
	assert.Equal(t, scope, scanner.lastScope)
	assert.Equal(t, target, scanner.lastTarget)
}

func TestPhase151ClaimedRecoveryCoordinatorAdapterContinuesAfterClaimMiss(t *testing.T) {
	target, scope, first, ownerID, claimNow, completedAt, lease :=
		phase151ClaimedRecoveryCoordinatorFixture()
	second := first
	second.ID = "phase151-claimed-coordinator-second"
	second.ClientNonce = second.ID
	second.CreatedAt = second.CreatedAt.Add(time.Second)
	provenance := *first.PostCreateProducerProvenance
	second.PostCreateProducerProvenance = &provenance
	scanner := &phase151ClaimedRecoveryCoordinatorScanner{
		target: target, scope: scope, documents: []messenger.MessageDocument{first, second},
	}
	effects := &phase151ClaimedRecoveryCoordinatorEffects{miss: map[string]bool{first.ID: true}}
	clock := &phase151ClaimedRecoveryRunnerClock{
		times: []time.Time{claimNow, claimNow.Add(time.Second), completedAt},
	}
	adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(
		scanner, ownerID, lease, clock, effects,
	)

	err := messenger.ExportPhase151ExecutePendingPostCreateRecovery(
		context.Background(),
		adapter,
		messenger.Phase151PendingRecoveryConfig{
			MaxPages: 1, MaxTargets: 1, PendingScanLimit: 10, MinimumInterval: time.Minute,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, []string{
		"claim:" + first.ID,
		"claim:" + second.ID,
		"auto-follow:" + second.ID,
		"complete:" + second.ID,
		"publish-unread:" + second.ID,
	}, effects.calls)
	assert.Equal(t, 3, clock.calls)
}

func TestPhase151ClaimedRecoveryCoordinatorAdapterPreservesRunnerError(t *testing.T) {
	target, scope, message, ownerID, claimNow, completedAt, lease :=
		phase151ClaimedRecoveryCoordinatorFixture()
	stageErr := errors.New("phase151 claimed coordinator effect failed")
	scanner := &phase151ClaimedRecoveryCoordinatorScanner{
		target: target, scope: scope, documents: []messenger.MessageDocument{message},
	}
	effects := &phase151ClaimedRecoveryCoordinatorEffects{followErr: stageErr}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}
	adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(
		scanner, ownerID, lease, clock, effects,
	)

	err := messenger.ExportPhase151ExecutePendingPostCreateRecovery(
		context.Background(),
		adapter,
		messenger.Phase151PendingRecoveryConfig{
			MaxPages: 1, MaxTargets: 1, PendingScanLimit: 10, MinimumInterval: time.Minute,
		},
	)

	require.ErrorIs(t, err, stageErr)
	assert.Equal(t, []string{"claim:" + message.ID, "auto-follow:" + message.ID}, effects.calls)
	assert.Equal(t, 1, clock.calls)
}

func TestPhase151ClaimedRecoveryCoordinatorAdapterRejectsInvalidConstruction(t *testing.T) {
	_, _, _, ownerID, claimNow, completedAt, lease := phase151ClaimedRecoveryCoordinatorFixture()
	scanner := &phase151ClaimedRecoveryCoordinatorScanner{}
	effects := &phase151ClaimedRecoveryCoordinatorEffects{}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}
	var typedNilScanner *phase151ClaimedRecoveryCoordinatorScanner
	var typedNilEffects *phase151ClaimedRecoveryCoordinatorEffects
	var typedNilClock *phase151ClaimedRecoveryRunnerClock

	assert.Nil(t, messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(nil, ownerID, lease, clock, effects))
	assert.Nil(t, messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(typedNilScanner, ownerID, lease, clock, effects))
	assert.Nil(t, messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, uuid.Nil, lease, clock, effects))
	assert.Nil(t, messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, ownerID, 0, clock, effects))
	assert.Nil(t, messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, ownerID, -time.Second, clock, effects))
	assert.Nil(t, messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, ownerID, lease, nil, effects))
	assert.Nil(t, messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, ownerID, lease, typedNilClock, effects))
	assert.Nil(t, messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, ownerID, lease, clock, nil))
	assert.Nil(t, messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, ownerID, lease, clock, typedNilEffects))
	assert.NotNil(t, messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, ownerID, lease, clock, effects))
}

func TestPhase151ClaimedRecoveryCoordinatorAdapterRejectsDriftingFreshAuthority(t *testing.T) {
	target, validScope, _, ownerID, claimNow, completedAt, lease :=
		phase151ClaimedRecoveryCoordinatorFixture()
	effects := &phase151ClaimedRecoveryCoordinatorEffects{}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}

	t.Run("malformed target stops before reauthorization", func(t *testing.T) {
		scanner := &phase151ClaimedRecoveryCoordinatorScanner{target: target, scope: validScope}
		adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(
			scanner, ownerID, lease, clock, effects,
		)
		malformed := target
		malformed.Kind = ""

		_, err := adapter.ReauthorizePendingRecoveryTarget(context.Background(), malformed)

		require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
		assert.Empty(t, scanner.calls)
	})

	for _, scenario := range []struct {
		name   string
		mutate func(*messenger.AuthorizedMessageProducerScope)
	}{
		{name: "foreign tenant", mutate: func(scope *messenger.AuthorizedMessageProducerScope) {
			scope.TenantID = uuid.New()
		}},
		{name: "foreign channel", mutate: func(scope *messenger.AuthorizedMessageProducerScope) {
			scope.ChannelID = uuid.New()
		}},
		{name: "nil actor", mutate: func(scope *messenger.AuthorizedMessageProducerScope) {
			scope.ActorID = uuid.Nil
		}},
		{name: "invalid producer proof", mutate: func(scope *messenger.AuthorizedMessageProducerScope) {
			scope.Producer = messenger.TrustedMessageProducer("unknown")
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			drifted := validScope
			scenario.mutate(&drifted)
			scanner := &phase151ClaimedRecoveryCoordinatorScanner{target: target, scope: drifted}
			adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(
				scanner, ownerID, lease, clock, effects,
			)

			_, err := adapter.ReauthorizePendingRecoveryTarget(context.Background(), target)

			require.Error(t, err)
			assert.Equal(t, []string{"reauthorize"}, scanner.calls)
			assert.Empty(t, effects.calls)
		})
	}
}

func TestPhase151ClaimedRecoveryCoordinatorAdapterForwardsScannerAndFailsClosed(t *testing.T) {
	target, scope, message, ownerID, claimNow, completedAt, lease :=
		phase151ClaimedRecoveryCoordinatorFixture()
	scanErr := errors.New("phase151 recovery scan failed")
	scanner := &phase151ClaimedRecoveryCoordinatorScanner{
		target: target, scope: scope, documents: []messenger.MessageDocument{message},
	}
	effects := &phase151ClaimedRecoveryCoordinatorEffects{}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}
	adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(
		scanner, ownerID, lease, clock, effects,
	)

	scanner.intervalErr = scanErr
	_, err := adapter.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
	require.ErrorIs(t, err, scanErr)
	scanner.intervalErr = nil

	scanner.pageErr = scanErr
	cursor := &messenger.Phase151RecoveryTargetCursor{
		TenantID: target.TenantID, ChannelID: target.ChannelID, Kind: target.Kind,
	}
	_, err = adapter.ListPendingRecoveryTargets(context.Background(), cursor, 7)
	require.ErrorIs(t, err, scanErr)
	assert.Same(t, cursor, scanner.lastCursor)
	assert.Equal(t, 7, scanner.lastLimit)
	scanner.pageErr = nil

	scanner.authorityErr = scanErr
	_, err = adapter.ReauthorizePendingRecoveryTarget(context.Background(), target)
	require.ErrorIs(t, err, scanErr)
	scanner.authorityErr = nil

	scanner.documentsErr = scanErr
	_, err = adapter.ListPendingPostCreateDocuments(context.Background(), scope, target, 9)
	require.ErrorIs(t, err, scanErr)
	assert.Equal(t, scope, scanner.lastScope)
	assert.Equal(t, target, scanner.lastTarget)
	assert.Equal(t, 9, scanner.lastLimit)

	var nilAdapter *messenger.Phase151ClaimedRecoveryCoordinatorAdapter
	_, err = nilAdapter.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	_, err = nilAdapter.ListPendingRecoveryTargets(context.Background(), nil, 1)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	_, err = nilAdapter.ReauthorizePendingRecoveryTarget(context.Background(), target)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	_, err = nilAdapter.ListPendingPostCreateDocuments(context.Background(), scope, target, 1)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	err = nilAdapter.ReconcilePendingPostCreateEffects(context.Background(), scope, message)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)

	_, err = adapter.TryAcquirePendingRecoveryInterval(nil, time.Minute)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
}
