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
	page         *messenger.Phase151PendingRecoveryTargetPage
	authority    messenger.Phase151RecoveryTargetAuthority
	documents    []messenger.MessageDocument
	calls        []string
	intervals    []time.Duration
	intervalErr  error
	pageErr      error
	authorityErr error
	documentsErr error
	lastCursor   *messenger.Phase151RecoveryTargetCursor
	lastLimit    int
	lastAuthority messenger.Phase151RecoveryTargetAuthority
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
	if s.page != nil {
		return *s.page, s.pageErr
	}
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
) (messenger.Phase151RecoveryTargetAuthority, error) {
	s.calls = append(s.calls, "reauthorize")
	s.lastTarget = target
	return s.authority, s.authorityErr
}

func (s *phase151ClaimedRecoveryCoordinatorScanner) ListPendingPostCreateDocuments(
	_ context.Context,
	authority messenger.Phase151RecoveryTargetAuthority,
	limit int,
) ([]messenger.MessageDocument, error) {
	s.calls = append(s.calls, "documents")
	s.lastAuthority = authority
	s.lastLimit = limit
	return append([]messenger.MessageDocument(nil), s.documents...), s.documentsErr
}

type phase151ClaimedRecoveryCoordinatorEffects struct {
	calls              []string
	miss               map[string]bool
	followErr          error
	owners             []uuid.UUID
	leases             []time.Duration
	claimScopes        []messenger.AuthorizedMessageProducerScope
	followScopes       []messenger.AuthorizedMessageProducerScope
	replyScopes        []messenger.AuthorizedMessageProducerScope
	completeScopes     []messenger.AuthorizedMessageProducerScope
	invalidationScopes []messenger.AuthorizedMessageProducerScope
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
	e.claimScopes = append(e.claimScopes, scope)
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
	scope messenger.AuthorizedMessageProducerScope,
	message *messenger.MessageDocument,
) error {
	e.calls = append(e.calls, "auto-follow:"+message.ID)
	e.followScopes = append(e.followScopes, scope)
	return e.followErr
}

func (e *phase151ClaimedRecoveryCoordinatorEffects) ReconcileReplyCount(
	_ context.Context,
	scope messenger.AuthorizedMessageProducerScope,
	message *messenger.MessageDocument,
) error {
	e.calls = append(e.calls, "reply-count:"+message.ID)
	e.replyScopes = append(e.replyScopes, scope)
	return nil
}

func (e *phase151ClaimedRecoveryCoordinatorEffects) CompleteClaimedPostCreateRecovery(
	_ context.Context,
	scope messenger.AuthorizedMessageProducerScope,
	message *messenger.MessageDocument,
	_ messenger.Phase151PostCreateRecoveryClaim,
	_ time.Time,
) error {
	e.calls = append(e.calls, "complete:"+message.ID)
	e.completeScopes = append(e.completeScopes, scope)
	return nil
}

func (e *phase151ClaimedRecoveryCoordinatorEffects) PublishUnreadInvalidation(
	_ context.Context,
	scope messenger.AuthorizedMessageProducerScope,
	message *messenger.MessageDocument,
) error {
	e.calls = append(e.calls, "publish-unread:"+message.ID)
	e.invalidationScopes = append(e.invalidationScopes, scope)
	return nil
}

func phase151ClaimedRecoveryCoordinatorFixture() (
	messenger.Phase151PendingRecoveryTarget,
	messenger.Phase151RecoveryTargetAuthority,
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
		Kind:      targetAuthority.Kind,
	}
	return target, targetAuthority, *message, ownerID, claimNow, completedAt, lease
}

func TestPhase151ClaimedRecoveryCoordinatorAdapterDerivesEffectScopesFromCanonicalMessageNotScanProof(t *testing.T) {
	target, authority, message, ownerID, claimNow, completedAt, lease :=
		phase151ClaimedRecoveryCoordinatorFixture()
	scanner := &phase151ClaimedRecoveryCoordinatorScanner{
		target: target, authority: authority, documents: []messenger.MessageDocument{message},
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
	assert.Equal(t, authority, scanner.lastAuthority)
	assert.Equal(t, target, scanner.lastTarget)

	canonicalEffectScope := messenger.AuthorizedMessageProducerScope{
		TenantID: target.TenantID, ActorID: message.SenderID, ChannelID: target.ChannelID,
		Producer: messenger.TrustedMessageProducerUser,
	}
	assert.Equal(t, []messenger.AuthorizedMessageProducerScope{canonicalEffectScope}, effects.claimScopes)
	assert.Equal(t, []messenger.AuthorizedMessageProducerScope{canonicalEffectScope}, effects.followScopes)
	assert.Equal(t, []messenger.AuthorizedMessageProducerScope{canonicalEffectScope}, effects.completeScopes)
	assert.Equal(t, []messenger.AuthorizedMessageProducerScope{canonicalEffectScope}, effects.invalidationScopes)

}

func TestPhase151ClaimedRecoveryCoordinatorAdapterContinuesAfterClaimMiss(t *testing.T) {
	target, authority, first, ownerID, claimNow, completedAt, lease :=
		phase151ClaimedRecoveryCoordinatorFixture()
	second := first
	second.ID = "phase151-claimed-coordinator-second"
	second.ClientNonce = second.ID
	second.CreatedAt = second.CreatedAt.Add(time.Second)
	provenance := *first.PostCreateProducerProvenance
	second.PostCreateProducerProvenance = &provenance
	scanner := &phase151ClaimedRecoveryCoordinatorScanner{
		target: target, authority: authority, documents: []messenger.MessageDocument{first, second},
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
	target, authority, message, ownerID, claimNow, completedAt, lease :=
		phase151ClaimedRecoveryCoordinatorFixture()
	stageErr := errors.New("phase151 claimed coordinator effect failed")
	scanner := &phase151ClaimedRecoveryCoordinatorScanner{
		target: target, authority: authority, documents: []messenger.MessageDocument{message},
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
	target, validAuthority, _, ownerID, claimNow, completedAt, lease :=
		phase151ClaimedRecoveryCoordinatorFixture()

	t.Run("malformed target stops before reauthorization", func(t *testing.T) {
		scanner := &phase151ClaimedRecoveryCoordinatorScanner{target: target, authority: validAuthority}
		adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(
			scanner, ownerID, lease, &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}, &phase151ClaimedRecoveryCoordinatorEffects{},
		)
		malformed := target
		malformed.Kind = ""

		_, err := adapter.ReauthorizePendingRecoveryTarget(context.Background(), malformed)

		require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
		assert.Empty(t, scanner.calls)
	})

	reauthorizationErr := errors.New("reauthorization revoked")
	for _, scenario := range []struct {
		name         string
		mutate       func(*messenger.Phase151RecoveryTargetAuthority)
		authorityErr error
	}{
		{name: "foreign principal", mutate: func(authority *messenger.Phase151RecoveryTargetAuthority) { authority.Principal = messenger.Phase151RecoveryServicePrincipal("messenger.user_actor") }},
		{name: "foreign tenant", mutate: func(authority *messenger.Phase151RecoveryTargetAuthority) { authority.TenantID = uuid.New() }},
		{name: "foreign channel", mutate: func(authority *messenger.Phase151RecoveryTargetAuthority) { authority.ChannelID = uuid.New() }},
		{name: "foreign kind", mutate: func(authority *messenger.Phase151RecoveryTargetAuthority) { authority.Kind = "dm" }},
		{name: "reauthorization error", authorityErr: reauthorizationErr},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			drifted := validAuthority
			if scenario.mutate != nil { scenario.mutate(&drifted) }
			scanner := &phase151ClaimedRecoveryCoordinatorScanner{
				target: target, authority: drifted, authorityErr: scenario.authorityErr,
			}
			adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(
				scanner, ownerID, lease, &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}, &phase151ClaimedRecoveryCoordinatorEffects{},
			)

			_, err := adapter.ReauthorizePendingRecoveryTarget(context.Background(), target)

			if scenario.authorityErr != nil {
				require.ErrorIs(t, err, scenario.authorityErr)
			} else {
				require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			}
			assert.Equal(t, []string{"reauthorize"}, scanner.calls)
		})
	}
}
func TestPhase151ClaimedRecoveryCoordinatorAdapterRejectsInvalidAuthorityOrDocumentDriftBeforeClaim(t *testing.T) {
	for _, scenario := range []struct {
		name            string
		mutateAuthority func(*messenger.Phase151RecoveryTargetAuthority)
		mutateDoc       func(*messenger.MessageDocument)
	}{
		{name: "empty authority kind", mutateAuthority: func(authority *messenger.Phase151RecoveryTargetAuthority) { authority.Kind = "" }},
		{name: "canonical document channel drift", mutateDoc: func(document *messenger.MessageDocument) { document.ChannelID = uuid.New() }},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			target, authority, message, ownerID, claimNow, completedAt, lease :=
				phase151ClaimedRecoveryCoordinatorFixture()
			if scenario.mutateAuthority != nil { scenario.mutateAuthority(&authority) }
			if scenario.mutateDoc != nil { scenario.mutateDoc(&message) }
			scanner := &phase151ClaimedRecoveryCoordinatorScanner{target: target, authority: authority}
			effects := &phase151ClaimedRecoveryCoordinatorEffects{}
			clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}
			adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, ownerID, lease, clock, effects)

			err := adapter.ReconcilePendingPostCreateEffects(context.Background(), authority, message)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Empty(t, effects.calls)
			assert.Zero(t, clock.calls)
		})
	}
}
func TestPhase151ClaimedRecoveryCoordinatorAdapterForwardsScannerAndFailsClosed(t *testing.T) {
	target, authority, message, ownerID, claimNow, completedAt, lease :=
		phase151ClaimedRecoveryCoordinatorFixture()
	scanErr := errors.New("phase151 recovery scan failed")
	scanner := &phase151ClaimedRecoveryCoordinatorScanner{
		target: target, authority: authority, documents: []messenger.MessageDocument{message},
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
	require.NotNil(t, scanner.lastCursor)
	assert.Equal(t, *cursor, *scanner.lastCursor)
	assert.NotSame(t, cursor, scanner.lastCursor)
	assert.Equal(t, 7, scanner.lastLimit)
	scanner.pageErr = nil

	scanner.authorityErr = scanErr
	_, err = adapter.ReauthorizePendingRecoveryTarget(context.Background(), target)
	require.ErrorIs(t, err, scanErr)
	scanner.authorityErr = nil

	scanner.documentsErr = scanErr
	_, err = adapter.ListPendingPostCreateDocuments(context.Background(), authority, 9)
	require.ErrorIs(t, err, scanErr)
	assert.Equal(t, authority, scanner.lastAuthority)
	assert.Equal(t, target, scanner.lastTarget)
	assert.Equal(t, 9, scanner.lastLimit)

	var nilAdapter *messenger.Phase151ClaimedRecoveryCoordinatorAdapter
	_, err = nilAdapter.TryAcquirePendingRecoveryInterval(context.Background(), time.Minute)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	_, err = nilAdapter.ListPendingRecoveryTargets(context.Background(), nil, 1)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	_, err = nilAdapter.ReauthorizePendingRecoveryTarget(context.Background(), target)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	_, err = nilAdapter.ListPendingPostCreateDocuments(context.Background(), authority, 1)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	err = nilAdapter.ReconcilePendingPostCreateEffects(context.Background(), authority, message)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)

	_, err = adapter.TryAcquirePendingRecoveryInterval(nil, time.Minute)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
}
func TestPhase151ClaimedRecoveryCoordinatorAdapterOwnsValidatedOrderedTargetPage(t *testing.T) {
	target, authority, _, ownerID, claimNow, completedAt, lease := phase151ClaimedRecoveryCoordinatorFixture()
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	firstChannelID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	secondChannelID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	after := &messenger.Phase151RecoveryTargetCursor{TenantID: tenantID, ChannelID: firstChannelID, Kind: "alpha"}
	first := messenger.Phase151PendingRecoveryTarget{TenantID: tenantID, ChannelID: firstChannelID, Kind: "beta"}
	second := messenger.Phase151PendingRecoveryTarget{TenantID: tenantID, ChannelID: secondChannelID, Kind: "future_kind"}
	next := &messenger.Phase151RecoveryTargetCursor{TenantID: second.TenantID, ChannelID: second.ChannelID, Kind: second.Kind}
	scannerTargets := []messenger.Phase151PendingRecoveryTarget{first, second}
	scanner := &phase151ClaimedRecoveryCoordinatorScanner{
		target: target, authority: authority,
		page: &messenger.Phase151PendingRecoveryTargetPage{Targets: scannerTargets, Next: next},
	}
	effects := &phase151ClaimedRecoveryCoordinatorEffects{}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}
	adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, ownerID, lease, clock, effects)

	page, err := adapter.ListPendingRecoveryTargets(context.Background(), after, 2)

	require.NoError(t, err)
	assert.Equal(t, []messenger.Phase151PendingRecoveryTarget{first, second}, page.Targets)
	require.NotNil(t, page.Next)
	assert.Equal(t, *next, *page.Next)
	require.NotNil(t, scanner.lastCursor)
	assert.NotSame(t, after, scanner.lastCursor)
	assert.NotSame(t, next, page.Next)

	after.Kind = "mutated-after"
	scannerTargets[0].Kind = "mutated-target"
	next.Kind = "mutated-next"
	assert.Equal(t, "alpha", scanner.lastCursor.Kind)
	assert.Equal(t, "beta", page.Targets[0].Kind)
	assert.Equal(t, "future_kind", page.Next.Kind)
}

func TestPhase151ClaimedRecoveryCoordinatorAdapterRejectsInvalidTargetPageInputsBeforeScanner(t *testing.T) {
	target, authority, _, ownerID, claimNow, completedAt, lease := phase151ClaimedRecoveryCoordinatorFixture()
	effects := &phase151ClaimedRecoveryCoordinatorEffects{}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}

	for _, scenario := range []struct {
		name  string
		after *messenger.Phase151RecoveryTargetCursor
		limit int
	}{
		{name: "zero limit", limit: 0},
		{name: "negative limit", limit: -1},
		{name: "cursor without tenant", limit: 1, after: &messenger.Phase151RecoveryTargetCursor{ChannelID: target.ChannelID, Kind: target.Kind}},
		{name: "cursor without channel", limit: 1, after: &messenger.Phase151RecoveryTargetCursor{TenantID: target.TenantID, Kind: target.Kind}},
		{name: "cursor without kind", limit: 1, after: &messenger.Phase151RecoveryTargetCursor{TenantID: target.TenantID, ChannelID: target.ChannelID}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scanner := &phase151ClaimedRecoveryCoordinatorScanner{target: target, authority: authority}
			adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, ownerID, lease, clock, effects)

			page, err := adapter.ListPendingRecoveryTargets(context.Background(), scenario.after, scenario.limit)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Equal(t, messenger.Phase151PendingRecoveryTargetPage{}, page)
			assert.Empty(t, scanner.calls)
		})
	}
}

func TestPhase151ClaimedRecoveryCoordinatorAdapterRejectsMalformedOrNonAdvancingTargetPages(t *testing.T) {
	target, authority, _, ownerID, claimNow, completedAt, lease := phase151ClaimedRecoveryCoordinatorFixture()
	effects := &phase151ClaimedRecoveryCoordinatorEffects{}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	firstChannelID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	secondChannelID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	first := messenger.Phase151PendingRecoveryTarget{TenantID: tenantID, ChannelID: firstChannelID, Kind: "channel"}
	second := messenger.Phase151PendingRecoveryTarget{TenantID: tenantID, ChannelID: secondChannelID, Kind: "channel"}
	afterFirst := &messenger.Phase151RecoveryTargetCursor{TenantID: first.TenantID, ChannelID: first.ChannelID, Kind: first.Kind}

	for _, scenario := range []struct {
		name  string
		after *messenger.Phase151RecoveryTargetCursor
		limit int
		page  messenger.Phase151PendingRecoveryTargetPage
	}{
		{name: "oversized", limit: 1, page: messenger.Phase151PendingRecoveryTargetPage{Targets: []messenger.Phase151PendingRecoveryTarget{first, second}}},
		{name: "malformed target", limit: 1, page: messenger.Phase151PendingRecoveryTargetPage{Targets: []messenger.Phase151PendingRecoveryTarget{{ChannelID: first.ChannelID, Kind: first.Kind}}}},
		{name: "duplicate tuple", limit: 2, page: messenger.Phase151PendingRecoveryTargetPage{Targets: []messenger.Phase151PendingRecoveryTarget{first, first}}},
		{name: "reversed tuples", limit: 2, page: messenger.Phase151PendingRecoveryTargetPage{Targets: []messenger.Phase151PendingRecoveryTarget{second, first}}},
		{name: "reversed tenant tuple", limit: 2, page: messenger.Phase151PendingRecoveryTargetPage{Targets: []messenger.Phase151PendingRecoveryTarget{{TenantID: uuid.MustParse("44444444-4444-4444-8444-444444444444"), ChannelID: firstChannelID, Kind: "channel"}, first}}},
		{name: "reversed kind tuple", limit: 2, page: messenger.Phase151PendingRecoveryTargetPage{Targets: []messenger.Phase151PendingRecoveryTarget{{TenantID: tenantID, ChannelID: firstChannelID, Kind: "beta"}, {TenantID: tenantID, ChannelID: firstChannelID, Kind: "alpha"}}}},
		{name: "first tuple equals after", after: afterFirst, limit: 1, page: messenger.Phase151PendingRecoveryTargetPage{Targets: []messenger.Phase151PendingRecoveryTarget{first}}},
		{name: "first tuple before after", after: &messenger.Phase151RecoveryTargetCursor{TenantID: second.TenantID, ChannelID: second.ChannelID, Kind: second.Kind}, limit: 1, page: messenger.Phase151PendingRecoveryTargetPage{Targets: []messenger.Phase151PendingRecoveryTarget{first}}},
		{name: "empty page advances", limit: 1, page: messenger.Phase151PendingRecoveryTargetPage{Next: &messenger.Phase151RecoveryTargetCursor{TenantID: first.TenantID, ChannelID: first.ChannelID, Kind: first.Kind}}},
		{name: "invalid next", limit: 2, page: messenger.Phase151PendingRecoveryTargetPage{Targets: []messenger.Phase151PendingRecoveryTarget{first, second}, Next: &messenger.Phase151RecoveryTargetCursor{TenantID: second.TenantID, ChannelID: second.ChannelID}}},
		{name: "next is not last", limit: 2, page: messenger.Phase151PendingRecoveryTargetPage{Targets: []messenger.Phase151PendingRecoveryTarget{first, second}, Next: &messenger.Phase151RecoveryTargetCursor{TenantID: first.TenantID, ChannelID: first.ChannelID, Kind: first.Kind}}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scannerPage := scenario.page
			scanner := &phase151ClaimedRecoveryCoordinatorScanner{target: target, authority: authority, page: &scannerPage}
			adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, ownerID, lease, clock, effects)

			page, err := adapter.ListPendingRecoveryTargets(context.Background(), scenario.after, scenario.limit)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Equal(t, messenger.Phase151PendingRecoveryTargetPage{}, page)
			assert.Equal(t, []string{"targets"}, scanner.calls)
		})
	}
}

func TestPhase151ClaimedRecoveryCoordinatorAdapterPreservesTargetPageScannerErrorWithoutPartialResult(t *testing.T) {
	target, authority, _, ownerID, claimNow, completedAt, lease := phase151ClaimedRecoveryCoordinatorFixture()
	scanErr := errors.New("phase151 target page scan failed")
	partial := messenger.Phase151PendingRecoveryTargetPage{Targets: []messenger.Phase151PendingRecoveryTarget{target}}
	scanner := &phase151ClaimedRecoveryCoordinatorScanner{
		target: target, authority: authority, page: &partial, pageErr: scanErr,
	}
	effects := &phase151ClaimedRecoveryCoordinatorEffects{}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}
	adapter := messenger.NewPhase151ClaimedRecoveryCoordinatorAdapter(scanner, ownerID, lease, clock, effects)

	page, err := adapter.ListPendingRecoveryTargets(context.Background(), nil, 1)

	require.ErrorIs(t, err, scanErr)
	assert.Equal(t, messenger.Phase151PendingRecoveryTargetPage{}, page)
	assert.Equal(t, []string{"targets"}, scanner.calls)
}
