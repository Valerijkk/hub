//go:build hubtest

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

type phase151NilContext struct{}

func (*phase151NilContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (*phase151NilContext) Done() <-chan struct{} {
	return nil
}

func (*phase151NilContext) Err() error {
	return nil
}

func (*phase151NilContext) Value(any) any {
	return nil
}

type phase151PendingRecoveryLeaderLeaseRepositorySpy struct {
	acquireCalls int
	renewCalls   int
	releaseCalls int
	advanceCalls int

	acquireHolderID uuid.UUID
	acquireDuration time.Duration
	acquisition     messenger.Phase151PendingRecoveryLeaderLeaseAcquisition
	acquireOwned    bool
	acquireErr      error
	afterAcquire    func()

	renewLease    messenger.Phase151PendingRecoveryLeaderLease
	renewDuration time.Duration
	renewResult   messenger.Phase151PendingRecoveryLeaderLease
	renewOwned    bool
	renewErr      error
	afterRenew    func()

	releaseLease messenger.Phase151PendingRecoveryLeaderLease
	releaseOwned bool
	releaseErr   error
	afterRelease func()

	advanceLease      messenger.Phase151PendingRecoveryLeaderLease
	advanceCheckpoint messenger.Phase151PendingRecoveryLeaderCheckpoint
	advanceOwned      bool
	advanceErr        error
	mutateAdvance     bool
	afterAdvance      func()
}

var _ messenger.Phase151PendingRecoveryLeaderLeaseRepository = (*phase151PendingRecoveryLeaderLeaseRepositorySpy)(nil)

func (r *phase151PendingRecoveryLeaderLeaseRepositorySpy) AcquirePendingRecoveryLeaderLease(
	_ context.Context,
	holderID uuid.UUID,
	duration time.Duration,
) (messenger.Phase151PendingRecoveryLeaderLeaseAcquisition, bool, error) {
	r.acquireCalls++
	r.acquireHolderID = holderID
	r.acquireDuration = duration
	if r.afterAcquire != nil {
		r.afterAcquire()
	}
	return r.acquisition, r.acquireOwned, r.acquireErr
}

func (r *phase151PendingRecoveryLeaderLeaseRepositorySpy) RenewPendingRecoveryLeaderLease(
	_ context.Context,
	lease messenger.Phase151PendingRecoveryLeaderLease,
	duration time.Duration,
) (messenger.Phase151PendingRecoveryLeaderLease, bool, error) {
	r.renewCalls++
	r.renewLease = lease
	r.renewDuration = duration
	if r.afterRenew != nil {
		r.afterRenew()
	}
	return r.renewResult, r.renewOwned, r.renewErr
}

func (r *phase151PendingRecoveryLeaderLeaseRepositorySpy) ReleasePendingRecoveryLeaderLease(
	_ context.Context,
	lease messenger.Phase151PendingRecoveryLeaderLease,
) (bool, error) {
	r.releaseCalls++
	r.releaseLease = lease
	if r.afterRelease != nil {
		r.afterRelease()
	}
	return r.releaseOwned, r.releaseErr
}

func (r *phase151PendingRecoveryLeaderLeaseRepositorySpy) AdvancePendingRecoveryLeaderCheckpoint(
	_ context.Context,
	lease messenger.Phase151PendingRecoveryLeaderLease,
	checkpoint messenger.Phase151PendingRecoveryLeaderCheckpoint,
) (bool, error) {
	r.advanceCalls++
	r.advanceLease = lease
	r.advanceCheckpoint = checkpoint
	if r.mutateAdvance && checkpoint.Cursor != nil {
		checkpoint.Cursor.Kind = "repository-mutated"
	}
	if r.afterAdvance != nil {
		r.afterAdvance()
	}
	return r.advanceOwned, r.advanceErr
}

func (r *phase151PendingRecoveryLeaderLeaseRepositorySpy) callCount() int {
	return r.acquireCalls + r.renewCalls + r.releaseCalls + r.advanceCalls
}

func phase151ValidPendingRecoveryLeaderLease(holderID uuid.UUID) messenger.Phase151PendingRecoveryLeaderLease {
	return messenger.Phase151PendingRecoveryLeaderLease{
		HolderID:   holderID,
		Fence:      7,
		LeaseUntil: time.Date(2001, time.January, 2, 3, 4, 5, 0, time.UTC),
	}
}

func phase151ValidPendingRecoveryLeaderCheckpoint() messenger.Phase151PendingRecoveryLeaderCheckpoint {
	return messenger.Phase151PendingRecoveryLeaderCheckpoint{
		Cursor: &messenger.Phase151RecoveryTargetCursor{
			TenantID:  uuid.New(),
			ChannelID: uuid.New(),
			Kind:      "channel",
		},
	}
}

func TestPhase151PendingRecoveryLeaderLeaseWrappersRejectNilDependenciesWithoutDelegation(t *testing.T) {
	holderID := uuid.New()
	lease := phase151ValidPendingRecoveryLeaderLease(holderID)
	checkpoint := phase151ValidPendingRecoveryLeaderCheckpoint()
	operations := []struct {
		name   string
		invoke func(context.Context, messenger.Phase151PendingRecoveryLeaderLeaseRepository) error
	}{
		{
			name: "acquire",
			invoke: func(ctx context.Context, repository messenger.Phase151PendingRecoveryLeaderLeaseRepository) error {
				_, _, err := messenger.ExportPhase151AcquirePendingRecoveryLeaderLease(ctx, repository, holderID, time.Second)
				return err
			},
		},
		{
			name: "renew",
			invoke: func(ctx context.Context, repository messenger.Phase151PendingRecoveryLeaderLeaseRepository) error {
				_, _, err := messenger.ExportPhase151RenewPendingRecoveryLeaderLease(ctx, repository, lease, time.Second)
				return err
			},
		},
		{
			name: "release",
			invoke: func(ctx context.Context, repository messenger.Phase151PendingRecoveryLeaderLeaseRepository) error {
				_, err := messenger.ExportPhase151ReleasePendingRecoveryLeaderLease(ctx, repository, lease)
				return err
			},
		},
		{
			name: "advance",
			invoke: func(ctx context.Context, repository messenger.Phase151PendingRecoveryLeaderLeaseRepository) error {
				_, err := messenger.ExportPhase151AdvancePendingRecoveryLeaderCheckpoint(ctx, repository, lease, checkpoint)
				return err
			},
		},
	}

	for _, operation := range operations {
		operation := operation
		t.Run(operation.name+"/nil_repository", func(t *testing.T) {
			var repository messenger.Phase151PendingRecoveryLeaderLeaseRepository
			err := operation.invoke(context.Background(), repository)
			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
		})

		t.Run(operation.name+"/typed_nil_repository", func(t *testing.T) {
			var repository *phase151PendingRecoveryLeaderLeaseRepositorySpy
			err := operation.invoke(context.Background(), repository)
			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
		})

		t.Run(operation.name+"/cancelled_context_with_nil_repository", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			var repository messenger.Phase151PendingRecoveryLeaderLeaseRepository

			err := operation.invoke(ctx, repository)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.False(t, errors.Is(err, context.Canceled))
		})

		t.Run(operation.name+"/cancelled_context_with_typed_nil_repository", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			var repository *phase151PendingRecoveryLeaderLeaseRepositorySpy

			err := operation.invoke(ctx, repository)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.False(t, errors.Is(err, context.Canceled))
		})

		t.Run(operation.name+"/nil_context", func(t *testing.T) {
			repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{}
			err := operation.invoke(nil, repository)
			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Zero(t, repository.callCount())
		})

		t.Run(operation.name+"/typed_nil_context", func(t *testing.T) {
			repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{}
			var ctx *phase151NilContext
			err := operation.invoke(ctx, repository)
			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Zero(t, repository.callCount())
		})
	}
}

func TestPhase151PendingRecoveryLeaderLeaseCancelledContextWinsBeforeInputValidation(t *testing.T) {
	operations := []struct {
		name   string
		invoke func(context.Context, messenger.Phase151PendingRecoveryLeaderLeaseRepository) error
	}{
		{
			name: "acquire",
			invoke: func(ctx context.Context, repository messenger.Phase151PendingRecoveryLeaderLeaseRepository) error {
				_, _, err := messenger.ExportPhase151AcquirePendingRecoveryLeaderLease(ctx, repository, uuid.Nil, 0)
				return err
			},
		},
		{
			name: "renew",
			invoke: func(ctx context.Context, repository messenger.Phase151PendingRecoveryLeaderLeaseRepository) error {
				_, _, err := messenger.ExportPhase151RenewPendingRecoveryLeaderLease(ctx, repository, messenger.Phase151PendingRecoveryLeaderLease{}, 0)
				return err
			},
		},
		{
			name: "release",
			invoke: func(ctx context.Context, repository messenger.Phase151PendingRecoveryLeaderLeaseRepository) error {
				_, err := messenger.ExportPhase151ReleasePendingRecoveryLeaderLease(ctx, repository, messenger.Phase151PendingRecoveryLeaderLease{})
				return err
			},
		},
		{
			name: "advance",
			invoke: func(ctx context.Context, repository messenger.Phase151PendingRecoveryLeaderLeaseRepository) error {
				_, err := messenger.ExportPhase151AdvancePendingRecoveryLeaderCheckpoint(
					ctx,
					repository,
					messenger.Phase151PendingRecoveryLeaderLease{},
					messenger.Phase151PendingRecoveryLeaderCheckpoint{},
				)
				return err
			},
		},
	}

	for _, operation := range operations {
		operation := operation
		t.Run(operation.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{}

			err := operation.invoke(ctx, repository)

			require.ErrorIs(t, err, context.Canceled)
			assert.Zero(t, repository.callCount())
		})
	}
}

func TestPhase151PendingRecoveryLeaderLeaseCancelledContextWinsAfterDelegation(t *testing.T) {
	holderID := uuid.New()
	lease := phase151ValidPendingRecoveryLeaderLease(holderID)
	checkpoint := phase151ValidPendingRecoveryLeaderCheckpoint()

	t.Run("acquire", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		storeErr := errors.New("acquire storage error")
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{
			acquisition: messenger.Phase151PendingRecoveryLeaderLeaseAcquisition{
				Lease:      lease,
				Checkpoint: checkpoint,
			},
			acquireOwned: true,
			acquireErr:   storeErr,
			afterAcquire: cancel,
		}

		_, _, err := messenger.ExportPhase151AcquirePendingRecoveryLeaderLease(ctx, repository, holderID, time.Second)

		require.ErrorIs(t, err, context.Canceled)
		assert.False(t, errors.Is(err, storeErr))
		assert.Equal(t, 1, repository.acquireCalls)
	})

	t.Run("renew", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		storeErr := errors.New("renew storage error")
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{
			renewResult: lease,
			renewOwned:  true,
			renewErr:    storeErr,
			afterRenew:  cancel,
		}

		_, _, err := messenger.ExportPhase151RenewPendingRecoveryLeaderLease(ctx, repository, lease, time.Second)

		require.ErrorIs(t, err, context.Canceled)
		assert.False(t, errors.Is(err, storeErr))
		assert.Equal(t, 1, repository.renewCalls)
	})

	t.Run("release", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		storeErr := errors.New("release storage error")
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{
			releaseOwned: true,
			releaseErr:   storeErr,
			afterRelease: cancel,
		}

		_, err := messenger.ExportPhase151ReleasePendingRecoveryLeaderLease(ctx, repository, lease)

		require.ErrorIs(t, err, context.Canceled)
		assert.False(t, errors.Is(err, storeErr))
		assert.Equal(t, 1, repository.releaseCalls)
	})

	t.Run("advance", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		storeErr := errors.New("advance storage error")
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{
			advanceOwned: true,
			advanceErr:   storeErr,
			afterAdvance: cancel,
		}

		_, err := messenger.ExportPhase151AdvancePendingRecoveryLeaderCheckpoint(ctx, repository, lease, checkpoint)

		require.ErrorIs(t, err, context.Canceled)
		assert.False(t, errors.Is(err, storeErr))
		assert.Equal(t, 1, repository.advanceCalls)
	})
}

func TestPhase151PendingRecoveryLeaderLeaseWrappersRejectInvalidInputWithoutDelegation(t *testing.T) {
	holderID := uuid.New()
	validLease := phase151ValidPendingRecoveryLeaderLease(holderID)
	validCheckpoint := phase151ValidPendingRecoveryLeaderCheckpoint()

	t.Run("acquire", func(t *testing.T) {
		for _, scenario := range []struct {
			name     string
			holderID uuid.UUID
			duration time.Duration
		}{
			{name: "nil holder", holderID: uuid.Nil, duration: time.Second},
			{name: "zero duration", holderID: holderID, duration: 0},
			{name: "negative duration", holderID: holderID, duration: -time.Second},
		} {
			scenario := scenario
			t.Run(scenario.name, func(t *testing.T) {
				repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{}
				_, _, err := messenger.ExportPhase151AcquirePendingRecoveryLeaderLease(
					context.Background(), repository, scenario.holderID, scenario.duration,
				)
				require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
				assert.Zero(t, repository.callCount())
			})
		}
	})

	invalidLeases := []struct {
		name  string
		lease messenger.Phase151PendingRecoveryLeaderLease
	}{
		{name: "nil holder", lease: messenger.Phase151PendingRecoveryLeaderLease{Fence: 7, LeaseUntil: validLease.LeaseUntil}},
		{name: "zero fence", lease: messenger.Phase151PendingRecoveryLeaderLease{HolderID: holderID, LeaseUntil: validLease.LeaseUntil}},
		{name: "negative fence", lease: messenger.Phase151PendingRecoveryLeaderLease{HolderID: holderID, Fence: -1, LeaseUntil: validLease.LeaseUntil}},
		{name: "zero lease until", lease: messenger.Phase151PendingRecoveryLeaderLease{HolderID: holderID, Fence: 7}},
	}

	t.Run("renew", func(t *testing.T) {
		for _, scenario := range append(invalidLeases, struct {
			name  string
			lease messenger.Phase151PendingRecoveryLeaderLease
		}{name: "zero duration", lease: validLease}) {
			scenario := scenario
			t.Run(scenario.name, func(t *testing.T) {
				repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{}
				duration := time.Second
				if scenario.name == "zero duration" {
					duration = 0
				}
				_, _, err := messenger.ExportPhase151RenewPendingRecoveryLeaderLease(
					context.Background(), repository, scenario.lease, duration,
				)
				require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
				assert.Zero(t, repository.callCount())
			})
		}
	})

	t.Run("release", func(t *testing.T) {
		for _, scenario := range invalidLeases {
			scenario := scenario
			t.Run(scenario.name, func(t *testing.T) {
				repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{}
				_, err := messenger.ExportPhase151ReleasePendingRecoveryLeaderLease(
					context.Background(), repository, scenario.lease,
				)
				require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
				assert.Zero(t, repository.callCount())
			})
		}
	})

	t.Run("advance", func(t *testing.T) {
		scenarios := []struct {
			name       string
			lease      messenger.Phase151PendingRecoveryLeaderLease
			checkpoint messenger.Phase151PendingRecoveryLeaderCheckpoint
		}{
			{name: "invalid lease", lease: messenger.Phase151PendingRecoveryLeaderLease{}, checkpoint: validCheckpoint},
			{name: "nil cursor", lease: validLease, checkpoint: messenger.Phase151PendingRecoveryLeaderCheckpoint{}},
			{name: "nil cursor tenant", lease: validLease, checkpoint: messenger.Phase151PendingRecoveryLeaderCheckpoint{Cursor: &messenger.Phase151RecoveryTargetCursor{ChannelID: uuid.New(), Kind: "channel"}}},
			{name: "nil cursor channel", lease: validLease, checkpoint: messenger.Phase151PendingRecoveryLeaderCheckpoint{Cursor: &messenger.Phase151RecoveryTargetCursor{TenantID: uuid.New(), Kind: "channel"}}},
			{name: "empty cursor kind", lease: validLease, checkpoint: messenger.Phase151PendingRecoveryLeaderCheckpoint{Cursor: &messenger.Phase151RecoveryTargetCursor{TenantID: uuid.New(), ChannelID: uuid.New()}}},
		}
		for _, scenario := range scenarios {
			scenario := scenario
			t.Run(scenario.name, func(t *testing.T) {
				repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{}
				_, err := messenger.ExportPhase151AdvancePendingRecoveryLeaderCheckpoint(
					context.Background(), repository, scenario.lease, scenario.checkpoint,
				)
				require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
				assert.Zero(t, repository.callCount())
			})
		}
	})
}

func TestPhase151PendingRecoveryLeaderLeaseAcquirePreservesOwnershipAndCopiesCheckpoint(t *testing.T) {
	holderID := uuid.New()
	sharedCheckpoint := phase151ValidPendingRecoveryLeaderCheckpoint()
	repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{
		acquisition: messenger.Phase151PendingRecoveryLeaderLeaseAcquisition{
			Lease:      phase151ValidPendingRecoveryLeaderLease(holderID),
			Checkpoint: sharedCheckpoint,
		},
		acquireOwned: true,
	}

	acquisition, owned, err := messenger.ExportPhase151AcquirePendingRecoveryLeaderLease(
		context.Background(), repository, holderID, time.Second,
	)

	require.NoError(t, err)
	require.True(t, owned)
	assert.Equal(t, holderID, acquisition.Lease.HolderID)
	assert.Equal(t, int64(7), acquisition.Lease.Fence)
	assert.Equal(t, 2001, acquisition.Lease.LeaseUntil.Year(), "the wrapper must not make a local expiry judgement")
	require.NotNil(t, acquisition.Checkpoint.Cursor)
	assert.Equal(t, "channel", acquisition.Checkpoint.Cursor.Kind)
	assert.NotSame(t, repository.acquisition.Checkpoint.Cursor, acquisition.Checkpoint.Cursor)
	assert.Equal(t, holderID, repository.acquireHolderID)
	assert.Equal(t, time.Second, repository.acquireDuration)

	repository.acquisition.Checkpoint.Cursor.Kind = "repository-mutated"
	assert.Equal(t, "channel", acquisition.Checkpoint.Cursor.Kind)

	t.Run("nil checkpoint is an allowed successful acquisition", func(t *testing.T) {
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{
			acquisition: messenger.Phase151PendingRecoveryLeaderLeaseAcquisition{
				Lease: phase151ValidPendingRecoveryLeaderLease(holderID),
			},
			acquireOwned: true,
		}

		acquisition, owned, err := messenger.ExportPhase151AcquirePendingRecoveryLeaderLease(
			context.Background(), repository, holderID, time.Second,
		)

		require.NoError(t, err)
		require.True(t, owned)
		assert.Nil(t, acquisition.Checkpoint.Cursor)
	})

	t.Run("ownership loss returns zero acquisition without error", func(t *testing.T) {
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{
			acquisition: messenger.Phase151PendingRecoveryLeaderLeaseAcquisition{
				Lease: phase151ValidPendingRecoveryLeaderLease(uuid.New()),
				Checkpoint: messenger.Phase151PendingRecoveryLeaderCheckpoint{
					Cursor: &messenger.Phase151RecoveryTargetCursor{},
				},
			},
		}

		acquisition, owned, err := messenger.ExportPhase151AcquirePendingRecoveryLeaderLease(
			context.Background(), repository, holderID, time.Second,
		)

		require.NoError(t, err)
		assert.False(t, owned)
		assert.Equal(t, messenger.Phase151PendingRecoveryLeaderLeaseAcquisition{}, acquisition)
	})

	t.Run("repository error is preserved", func(t *testing.T) {
		storeErr := errors.New("leader lease store unavailable")
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{acquireErr: storeErr}

		_, _, err := messenger.ExportPhase151AcquirePendingRecoveryLeaderLease(
			context.Background(), repository, holderID, time.Second,
		)

		require.ErrorIs(t, err, storeErr)
	})

	for _, scenario := range []struct {
		name        string
		acquisition messenger.Phase151PendingRecoveryLeaderLeaseAcquisition
	}{
		{name: "foreign holder", acquisition: messenger.Phase151PendingRecoveryLeaderLeaseAcquisition{Lease: phase151ValidPendingRecoveryLeaderLease(uuid.New())}},
		{name: "zero fence", acquisition: messenger.Phase151PendingRecoveryLeaderLeaseAcquisition{Lease: messenger.Phase151PendingRecoveryLeaderLease{HolderID: holderID, LeaseUntil: phase151ValidPendingRecoveryLeaderLease(holderID).LeaseUntil}}},
		{name: "zero lease until", acquisition: messenger.Phase151PendingRecoveryLeaderLeaseAcquisition{Lease: messenger.Phase151PendingRecoveryLeaderLease{HolderID: holderID, Fence: 7}}},
		{name: "invalid checkpoint", acquisition: messenger.Phase151PendingRecoveryLeaderLeaseAcquisition{Lease: phase151ValidPendingRecoveryLeaderLease(holderID), Checkpoint: messenger.Phase151PendingRecoveryLeaderCheckpoint{Cursor: &messenger.Phase151RecoveryTargetCursor{}}}},
	} {
		scenario := scenario
		t.Run("invalid successful result/"+scenario.name, func(t *testing.T) {
			repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{
				acquisition: scenario.acquisition,
				acquireOwned: true,
			}

			acquisition, owned, err := messenger.ExportPhase151AcquirePendingRecoveryLeaderLease(
				context.Background(), repository, holderID, time.Second,
			)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.False(t, owned)
			assert.Equal(t, messenger.Phase151PendingRecoveryLeaderLeaseAcquisition{}, acquisition)
			assert.Equal(t, 1, repository.acquireCalls)
		})
	}
}

func TestPhase151PendingRecoveryLeaderLeaseRenewReleaseAndAdvancePreserveLossAndErrors(t *testing.T) {
	holderID := uuid.New()
	lease := phase151ValidPendingRecoveryLeaderLease(holderID)
	checkpoint := phase151ValidPendingRecoveryLeaderCheckpoint()

	t.Run("renew ownership loss returns zero lease without error", func(t *testing.T) {
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{
			renewResult: messenger.Phase151PendingRecoveryLeaderLease{
				HolderID:   uuid.New(),
				Fence:      lease.Fence + 1,
				LeaseUntil: lease.LeaseUntil,
			},
		}

		renewed, owned, err := messenger.ExportPhase151RenewPendingRecoveryLeaderLease(
			context.Background(), repository, lease, time.Second,
		)

		require.NoError(t, err)
		assert.False(t, owned)
		assert.Equal(t, messenger.Phase151PendingRecoveryLeaderLease{}, renewed)
	})

	t.Run("renew error is preserved", func(t *testing.T) {
		storeErr := errors.New("renew store unavailable")
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{renewErr: storeErr}

		_, _, err := messenger.ExportPhase151RenewPendingRecoveryLeaderLease(
			context.Background(), repository, lease, time.Second,
		)

		require.ErrorIs(t, err, storeErr)
	})

	for _, scenario := range []struct {
		name  string
		lease messenger.Phase151PendingRecoveryLeaderLease
	}{
		{name: "foreign holder", lease: phase151ValidPendingRecoveryLeaderLease(uuid.New())},
		{name: "foreign fence", lease: messenger.Phase151PendingRecoveryLeaderLease{HolderID: holderID, Fence: 8, LeaseUntil: lease.LeaseUntil}},
		{name: "zero lease until", lease: messenger.Phase151PendingRecoveryLeaderLease{HolderID: holderID, Fence: 7}},
	} {
		scenario := scenario
		t.Run("renew rejects invalid successful result/"+scenario.name, func(t *testing.T) {
			repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{
				renewResult: scenario.lease,
				renewOwned:  true,
			}

			renewed, owned, err := messenger.ExportPhase151RenewPendingRecoveryLeaderLease(
				context.Background(), repository, lease, time.Second,
			)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.False(t, owned)
			assert.Equal(t, messenger.Phase151PendingRecoveryLeaderLease{}, renewed)
			assert.Equal(t, 1, repository.renewCalls)
		})
	}

	t.Run("renew keeps an expired but nonzero lease as repository authority", func(t *testing.T) {
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{
			renewResult: lease,
			renewOwned:  true,
		}

		renewed, owned, err := messenger.ExportPhase151RenewPendingRecoveryLeaderLease(
			context.Background(), repository, lease, time.Second,
		)

		require.NoError(t, err)
		require.True(t, owned)
		assert.Equal(t, lease, renewed)
		assert.Equal(t, lease, repository.renewLease)
		assert.Equal(t, time.Second, repository.renewDuration)
	})

	t.Run("release ownership loss is normal", func(t *testing.T) {
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{}

		owned, err := messenger.ExportPhase151ReleasePendingRecoveryLeaderLease(context.Background(), repository, lease)

		require.NoError(t, err)
		assert.False(t, owned)
	})

	t.Run("release error is preserved", func(t *testing.T) {
		storeErr := errors.New("release store unavailable")
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{releaseErr: storeErr}

		_, err := messenger.ExportPhase151ReleasePendingRecoveryLeaderLease(context.Background(), repository, lease)

		require.ErrorIs(t, err, storeErr)
	})

	t.Run("release forwards lease without local expiry judgement", func(t *testing.T) {
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{releaseOwned: true}

		owned, err := messenger.ExportPhase151ReleasePendingRecoveryLeaderLease(context.Background(), repository, lease)

		require.NoError(t, err)
		assert.True(t, owned)
		assert.Equal(t, lease, repository.releaseLease)
	})

	t.Run("advance ownership loss is normal", func(t *testing.T) {
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{}

		owned, err := messenger.ExportPhase151AdvancePendingRecoveryLeaderCheckpoint(
			context.Background(), repository, lease, checkpoint,
		)

		require.NoError(t, err)
		assert.False(t, owned)
	})

	t.Run("advance error is preserved", func(t *testing.T) {
		storeErr := errors.New("advance store unavailable")
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{advanceErr: storeErr}

		_, err := messenger.ExportPhase151AdvancePendingRecoveryLeaderCheckpoint(
			context.Background(), repository, lease, checkpoint,
		)

		require.ErrorIs(t, err, storeErr)
	})

	t.Run("advance gives repository an independent checkpoint copy", func(t *testing.T) {
		repository := &phase151PendingRecoveryLeaderLeaseRepositorySpy{
			advanceOwned:  true,
			mutateAdvance: true,
		}

		owned, err := messenger.ExportPhase151AdvancePendingRecoveryLeaderCheckpoint(
			context.Background(), repository, lease, checkpoint,
		)

		require.NoError(t, err)
		assert.True(t, owned)
		assert.Equal(t, lease, repository.advanceLease)
		require.NotNil(t, repository.advanceCheckpoint.Cursor)
		assert.Equal(t, "repository-mutated", repository.advanceCheckpoint.Cursor.Kind)
		require.NotNil(t, checkpoint.Cursor)
		assert.Equal(t, "channel", checkpoint.Cursor.Kind)
		assert.NotSame(t, checkpoint.Cursor, repository.advanceCheckpoint.Cursor)
	})
}
