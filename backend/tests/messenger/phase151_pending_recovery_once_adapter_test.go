//go:build hubtest

package messenger_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

var _ messenger.Phase151PendingRecoveryRepository = (*phase151PendingRecoveryOnceRepository)(nil)

type phase151PendingRecoveryOnceRepository struct {
	acquired       bool
	gateErr        error
	page           messenger.Phase151PendingRecoveryTargetPage
	pageErr        error
	authority      messenger.Phase151RecoveryTargetAuthority
	authorityErr   error
	documents      []messenger.MessageDocument
	documentsErr   error
	reconcileErr   error
	calls          []string
	contexts       []context.Context
	minimums       []time.Duration
	targetLimits   []int
	documentLimits []int
	targets        []messenger.Phase151PendingRecoveryTarget
	authorities    []messenger.Phase151RecoveryTargetAuthority
	documentsSeen  []messenger.MessageDocument
}

func (r *phase151PendingRecoveryOnceRepository) TryAcquirePendingRecoveryInterval(ctx context.Context, minimum time.Duration) (bool, error) {
	r.calls = append(r.calls, "gate")
	r.contexts = append(r.contexts, ctx)
	r.minimums = append(r.minimums, minimum)
	return r.acquired, r.gateErr
}

func (r *phase151PendingRecoveryOnceRepository) ListPendingRecoveryTargets(
	ctx context.Context,
	_ *messenger.Phase151RecoveryTargetCursor,
	limit int,
) (messenger.Phase151PendingRecoveryTargetPage, error) {
	r.calls = append(r.calls, "targets")
	r.contexts = append(r.contexts, ctx)
	r.targetLimits = append(r.targetLimits, limit)
	return r.page, r.pageErr
}

func (r *phase151PendingRecoveryOnceRepository) ReauthorizePendingRecoveryTarget(
	ctx context.Context,
	target messenger.Phase151PendingRecoveryTarget,
) (messenger.Phase151RecoveryTargetAuthority, error) {
	r.calls = append(r.calls, "reauthorize")
	r.contexts = append(r.contexts, ctx)
	r.targets = append(r.targets, target)
	return r.authority, r.authorityErr
}

func (r *phase151PendingRecoveryOnceRepository) ListPendingPostCreateDocuments(
	ctx context.Context,
	authority messenger.Phase151RecoveryTargetAuthority,
	limit int,
) ([]messenger.MessageDocument, error) {
	r.calls = append(r.calls, "documents")
	r.contexts = append(r.contexts, ctx)
	r.authorities = append(r.authorities, authority)
	r.documentLimits = append(r.documentLimits, limit)
	return append([]messenger.MessageDocument(nil), r.documents...), r.documentsErr
}

func (r *phase151PendingRecoveryOnceRepository) ReconcilePendingPostCreateEffects(
	ctx context.Context,
	authority messenger.Phase151RecoveryTargetAuthority,
	document messenger.MessageDocument,
) error {
	r.calls = append(r.calls, "reconcile")
	r.contexts = append(r.contexts, ctx)
	r.authorities = append(r.authorities, authority)
	r.documentsSeen = append(r.documentsSeen, document)
	return r.reconcileErr
}

type phase151PendingRecoveryOnceTypedNilContext struct{}

func (*phase151PendingRecoveryOnceTypedNilContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}
func (*phase151PendingRecoveryOnceTypedNilContext) Done() <-chan struct{} { return nil }
func (*phase151PendingRecoveryOnceTypedNilContext) Err() error            { return nil }
func (*phase151PendingRecoveryOnceTypedNilContext) Value(any) any         { return nil }

func phase151PendingRecoveryOnceValidConfig() messenger.Phase151PendingRecoveryConfig {
	return messenger.Phase151PendingRecoveryConfig{
		MaxPages:         1,
		MaxTargets:       3,
		PendingScanLimit: 7,
		MinimumInterval:  2 * time.Minute,
	}
}

func phase151PendingRecoveryOnceRepositoryFixture() (
	*phase151PendingRecoveryOnceRepository,
	messenger.Phase151PendingRecoveryTarget,
	messenger.Phase151RecoveryTargetAuthority,
	messenger.MessageDocument,
) {
	target := messenger.Phase151PendingRecoveryTarget{
		TenantID: uuid.New(), ChannelID: uuid.New(), Kind: "channel",
	}
	authority := messenger.Phase151RecoveryTargetAuthority{
		Principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
		TenantID:  target.TenantID,
		ChannelID: target.ChannelID,
		Kind:      target.Kind,
	}
	document := messenger.MessageDocument{
		ID:                     "phase151-pending-recovery-once",
		PostCreateEffectsState: messenger.PostCreateEffectsStatePending,
	}
	next := messenger.Phase151RecoveryTargetCursor{
		TenantID: uuid.New(), ChannelID: uuid.New(), Kind: "channel",
	}
	return &phase151PendingRecoveryOnceRepository{
		acquired: true,
		page: messenger.Phase151PendingRecoveryTargetPage{
			Targets: []messenger.Phase151PendingRecoveryTarget{target},
			Next:    &next,
		},
		authority: authority,
		documents: []messenger.MessageDocument{document},
	}, target, authority, document
}

func TestPhase151PendingRecoveryOnceAdapterRejectsInvalidConstructionAndPreservesValueCopies(t *testing.T) {
	t.Parallel()

	validConfig := phase151PendingRecoveryOnceValidConfig()
	repository, _, _, _ := phase151PendingRecoveryOnceRepositoryFixture()
	var typedNilRepository *phase151PendingRecoveryOnceRepository

	require.Nil(t, messenger.NewPhase151PendingRecoveryOnceAdapter(nil, validConfig))
	require.Nil(t, messenger.NewPhase151PendingRecoveryOnceAdapter(typedNilRepository, validConfig))

	for _, test := range []struct {
		name   string
		config messenger.Phase151PendingRecoveryConfig
	}{
		{name: "zero max pages", config: messenger.Phase151PendingRecoveryConfig{MaxTargets: 1, PendingScanLimit: 1, MinimumInterval: time.Second}},
		{name: "negative max pages", config: messenger.Phase151PendingRecoveryConfig{MaxPages: -1, MaxTargets: 1, PendingScanLimit: 1, MinimumInterval: time.Second}},
		{name: "zero max targets", config: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, PendingScanLimit: 1, MinimumInterval: time.Second}},
		{name: "negative max targets", config: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: -1, PendingScanLimit: 1, MinimumInterval: time.Second}},
		{name: "zero pending scan limit", config: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: 1, MinimumInterval: time.Second}},
		{name: "negative pending scan limit", config: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: 1, PendingScanLimit: -1, MinimumInterval: time.Second}},
		{name: "zero minimum interval", config: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: 1, PendingScanLimit: 1}},
		{name: "negative minimum interval", config: messenger.Phase151PendingRecoveryConfig{MaxPages: 1, MaxTargets: 1, PendingScanLimit: 1, MinimumInterval: -time.Nanosecond}},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Nil(t, messenger.NewPhase151PendingRecoveryOnceAdapter(repository, test.config))
		})
	}

	adapter := messenger.NewPhase151PendingRecoveryOnceAdapter(repository, validConfig)
	require.NotNil(t, adapter)
	require.Equal(t, validConfig, adapter.PendingRecoveryConfig())

	validConfig.MaxPages = 99
	validConfig.MinimumInterval = time.Second
	returned := adapter.PendingRecoveryConfig()
	returned.MaxTargets = 101
	returned.PendingScanLimit = 103

	require.Equal(t, phase151PendingRecoveryOnceValidConfig(), adapter.PendingRecoveryConfig())
}

func TestPhase151PendingRecoveryOnceAdapterNilReceiverAndInvalidExecutionDoNotCallRepository(t *testing.T) {
	t.Parallel()

	var nilAdapter *messenger.Phase151PendingRecoveryOnceAdapter
	require.Equal(t, messenger.Phase151PendingRecoveryConfig{}, nilAdapter.PendingRecoveryConfig())
	require.Same(t, messenger.ErrPostCreatePipelineInvalid, nilAdapter.Execute(context.Background()))
	cancelledInvalid, cancelInvalid := context.WithCancel(context.Background())
	cancelInvalid()
	require.Same(t, messenger.ErrPostCreatePipelineInvalid, nilAdapter.Execute(cancelledInvalid))

	repository, _, _, _ := phase151PendingRecoveryOnceRepositoryFixture()
	adapter := messenger.NewPhase151PendingRecoveryOnceAdapter(repository, phase151PendingRecoveryOnceValidConfig())
	require.NotNil(t, adapter)
	var typedNilRepository *phase151PendingRecoveryOnceRepository
	rejectedAdapter := messenger.NewPhase151PendingRecoveryOnceAdapter(typedNilRepository, phase151PendingRecoveryOnceValidConfig())
	require.Nil(t, rejectedAdapter)
	require.Same(t, messenger.ErrPostCreatePipelineInvalid, rejectedAdapter.Execute(cancelledInvalid))

	var typedNilContext *phase151PendingRecoveryOnceTypedNilContext
	for _, test := range []struct {
		name string
		ctx  context.Context
	}{
		{name: "nil context"},
		{name: "typed nil context", ctx: typedNilContext},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			require.Same(t, messenger.ErrPostCreatePipelineInvalid, adapter.Execute(test.ctx))
			require.Empty(t, repository.calls)
		})
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	err := adapter.Execute(cancelled)
	require.Same(t, cancelled.Err(), err)
	require.Empty(t, repository.calls)
}

func TestPhase151PendingRecoveryOnceAdapterDelegatesOneExplicitPassWithExactConfiguration(t *testing.T) {
	t.Parallel()

	repository, target, authority, document := phase151PendingRecoveryOnceRepositoryFixture()
	config := phase151PendingRecoveryOnceValidConfig()
	adapter := messenger.NewPhase151PendingRecoveryOnceAdapter(repository, config)
	ctx := context.WithValue(context.Background(), "phase151-pending-recovery-once", "original")

	require.NoError(t, adapter.Execute(ctx))
	require.Equal(t, []string{"gate", "targets", "reauthorize", "documents", "reconcile"}, repository.calls)
	require.Equal(t, []time.Duration{config.MinimumInterval}, repository.minimums)
	require.Equal(t, []int{config.MaxTargets}, repository.targetLimits)
	require.Equal(t, []int{config.PendingScanLimit}, repository.documentLimits)
	require.Equal(t, []messenger.Phase151PendingRecoveryTarget{target}, repository.targets)
	require.Equal(t, []messenger.Phase151RecoveryTargetAuthority{authority, authority}, repository.authorities)
	require.Equal(t, []messenger.MessageDocument{document}, repository.documentsSeen)
	for _, received := range repository.contexts {
		require.Same(t, ctx, received)
	}
}

func TestPhase151PendingRecoveryOnceAdapterReturnsGateDenialOrExecutorErrorWithoutScanning(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		acquired bool
		gateErr  error
	}{
		{name: "denied", acquired: false},
		{
			name:     "sentinel error",
			acquired: true,
			gateErr:  errors.New("pending recovery gate unavailable"),
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			repository, _, _, _ := phase151PendingRecoveryOnceRepositoryFixture()
			repository.acquired = test.acquired
			repository.gateErr = test.gateErr
			adapter := messenger.NewPhase151PendingRecoveryOnceAdapter(repository, phase151PendingRecoveryOnceValidConfig())

			err := adapter.Execute(context.Background())

			if test.gateErr != nil {
				require.ErrorIs(t, err, test.gateErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, []string{"gate"}, repository.calls)
		})
	}
}
