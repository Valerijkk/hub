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
	"go.mongodb.org/mongo-driver/v2/bson"
)

var _ messenger.Phase151PostCreateRecoveryMongoStore = (*phase151PostCreateRecoveryMongoStore)(nil)
var _ messenger.Phase151PostCreateRecoveryMongoRepository = (*messenger.Phase151PostCreateRecoveryMongoAdapter)(nil)
var _ messenger.Phase151PostCreateRecoveryMongoRepository = (*messenger.MongoRepository)(nil)

type phase151PostCreateRecoveryMongoStore struct {
	claimDocument   *messenger.MessageDocument
	claimAcquired   bool
	claimErr        error
	completeMatch   bool
	completeErr     error
	claimFilters    []bson.D
	claimUpdates    []bson.D
	completeFilters []bson.D
	completeUpdates []bson.D
}

func (s *phase151PostCreateRecoveryMongoStore) ApplyPostCreateRecoveryClaim(
	_ context.Context,
	filter bson.D,
	update bson.D,
) (*messenger.MessageDocument, bool, error) {
	s.claimFilters = append(s.claimFilters, filter)
	s.claimUpdates = append(s.claimUpdates, update)
	return s.claimDocument, s.claimAcquired, s.claimErr
}

func (s *phase151PostCreateRecoveryMongoStore) ApplyPostCreateRecoveryCompletion(
	_ context.Context,
	filter bson.D,
	update bson.D,
) (bool, error) {
	s.completeFilters = append(s.completeFilters, filter)
	s.completeUpdates = append(s.completeUpdates, update)
	return s.completeMatch, s.completeErr
}

func phase151ClaimedPostCreateDocument(
	message *messenger.MessageDocument,
	ownerID uuid.UUID,
	fence int64,
	leaseUntil time.Time,
) *messenger.MessageDocument {
	claimed := *message
	claimed.PostCreateRecoveryOwnerID = ownerID
	claimed.PostCreateRecoveryFence = fence
	claimed.PostCreateRecoveryLeaseUntil = &leaseUntil
	return &claimed
}

func TestPhase151PostCreateRecoveryMongoAdapterReturnsExactPostUpdateClaim(t *testing.T) {
	scope, message, now := phase151RecoveryClaimFixture()
	ownerID := uuid.New()
	plan, err := messenger.ExportPhase151BuildPostCreateRecoveryClaimPlan(
		scope, message, ownerID, now, 2*time.Minute,
	)
	require.NoError(t, err)
	claimed := phase151ClaimedPostCreateDocument(message, ownerID, 9, plan.LeaseUntil)
	resolvedMentionID := uuid.New()
	claimed.ResolvedMentionUserIDs = []uuid.UUID{resolvedMentionID}
	store := &phase151PostCreateRecoveryMongoStore{claimDocument: claimed, claimAcquired: true}
	adapter := messenger.NewPhase151PostCreateRecoveryMongoAdapter(store)

	acquisition, acquired, err := adapter.ClaimPendingPostCreateRecovery(
		context.Background(), scope, message, ownerID, now, 2*time.Minute,
	)

	require.NoError(t, err)
	require.True(t, acquired)
	assert.Equal(t, messenger.Phase151PostCreateRecoveryClaim{
		TenantID: scope.TenantID, MessageID: message.ID, ChannelID: scope.ChannelID,
		OwnerID: ownerID, Fence: 9, LeaseUntil: plan.LeaseUntil,
	}, acquisition.Claim)
	assert.Equal(t, *claimed, acquisition.Message)
	require.Equal(t, []bson.D{plan.Filter}, store.claimFilters)
	require.Equal(t, []bson.D{plan.Update}, store.claimUpdates)
	assert.Empty(t, store.completeFilters)

	claimed.PostCreateRecoveryFence = 99
	claimed.ResolvedMentionUserIDs[0] = uuid.New()
	*claimed.PostCreateRecoveryLeaseUntil = claimed.PostCreateRecoveryLeaseUntil.Add(time.Minute)
	assert.Equal(t, int64(9), acquisition.Message.PostCreateRecoveryFence,
		"the acquisition owns a value copy of the store result")
	assert.Equal(t, []uuid.UUID{resolvedMentionID}, acquisition.Message.ResolvedMentionUserIDs,
		"effect input must not alias the store result")
	require.NotNil(t, acquisition.Message.PostCreateRecoveryLeaseUntil)
	assert.Equal(t, plan.LeaseUntil, *acquisition.Message.PostCreateRecoveryLeaseUntil,
		"lease metadata must not alias the store result")
}

func TestPhase151PostCreateRecoveryMongoAdapterHandlesMissErrorsAndCancellation(t *testing.T) {
	scope, message, now := phase151RecoveryClaimFixture()
	ownerID := uuid.New()

	t.Run("claim miss is a successful non-acquisition", func(t *testing.T) {
		store := &phase151PostCreateRecoveryMongoStore{}
		adapter := messenger.NewPhase151PostCreateRecoveryMongoAdapter(store)

		acquisition, acquired, err := adapter.ClaimPendingPostCreateRecovery(
			context.Background(), scope, message, ownerID, now, time.Minute,
		)

		require.NoError(t, err)
		assert.False(t, acquired)
		assert.Equal(t, messenger.Phase151PostCreateRecoveryAcquisition{}, acquisition)
		require.Len(t, store.claimFilters, 1)
	})

	t.Run("store error retains identity", func(t *testing.T) {
		storeErr := errors.New("claim store failed")
		store := &phase151PostCreateRecoveryMongoStore{claimErr: storeErr}
		adapter := messenger.NewPhase151PostCreateRecoveryMongoAdapter(store)

		_, acquired, err := adapter.ClaimPendingPostCreateRecovery(
			context.Background(), scope, message, ownerID, now, time.Minute,
		)

		assert.False(t, acquired)
		require.ErrorIs(t, err, storeErr)
	})

	t.Run("cancelled context stops before store", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		store := &phase151PostCreateRecoveryMongoStore{}
		adapter := messenger.NewPhase151PostCreateRecoveryMongoAdapter(store)

		_, acquired, err := adapter.ClaimPendingPostCreateRecovery(
			ctx, scope, message, ownerID, now, time.Minute,
		)

		assert.False(t, acquired)
		require.ErrorIs(t, err, context.Canceled)
		assert.Empty(t, store.claimFilters)
	})
}

func TestPhase151PostCreateRecoveryMongoAdapterRejectsMalformedClaimResult(t *testing.T) {
	scope, message, now := phase151RecoveryClaimFixture()
	ownerID := uuid.New()
	plan, err := messenger.ExportPhase151BuildPostCreateRecoveryClaimPlan(scope, message, ownerID, now, time.Minute)
	require.NoError(t, err)
	valid := phase151ClaimedPostCreateDocument(message, ownerID, 1, plan.LeaseUntil)
	deletedAt := now
	completedAt := now
	wrongLease := plan.LeaseUntil.Add(time.Second)

	cases := []struct {
		name     string
		document *messenger.MessageDocument
		acquired bool
	}{
		{name: "acquired nil", acquired: true},
		{name: "miss with document", document: valid},
		{name: "wrong id", document: func() *messenger.MessageDocument { value := *valid; value.ID = "foreign"; return &value }(), acquired: true},
		{name: "wrong channel", document: func() *messenger.MessageDocument { value := *valid; value.ChannelID = uuid.New(); return &value }(), acquired: true},
		{name: "wrong sender", document: func() *messenger.MessageDocument { value := *valid; value.SenderID = uuid.New(); return &value }(), acquired: true},
		{name: "created at drift", document: func() *messenger.MessageDocument {
			value := *valid
			value.CreatedAt = value.CreatedAt.Add(time.Second)
			return &value
		}(), acquired: true},
		{name: "deleted", document: func() *messenger.MessageDocument { value := *valid; value.DeletedAt = &deletedAt; return &value }(), acquired: true},
		{name: "complete", document: func() *messenger.MessageDocument {
			value := *valid
			value.PostCreateEffectsState = messenger.PostCreateEffectsStateComplete
			return &value
		}(), acquired: true},
		{name: "completed timestamp", document: func() *messenger.MessageDocument {
			value := *valid
			value.PostCreateEffectsCompletedAt = &completedAt
			return &value
		}(), acquired: true},
		{name: "wrong owner", document: func() *messenger.MessageDocument {
			value := *valid
			value.PostCreateRecoveryOwnerID = uuid.New()
			return &value
		}(), acquired: true},
		{name: "zero fence", document: func() *messenger.MessageDocument { value := *valid; value.PostCreateRecoveryFence = 0; return &value }(), acquired: true},
		{name: "nil lease", document: func() *messenger.MessageDocument {
			value := *valid
			value.PostCreateRecoveryLeaseUntil = nil
			return &value
		}(), acquired: true},
		{name: "wrong lease", document: func() *messenger.MessageDocument {
			value := *valid
			value.PostCreateRecoveryLeaseUntil = &wrongLease
			return &value
		}(), acquired: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &phase151PostCreateRecoveryMongoStore{
				claimDocument: tc.document, claimAcquired: tc.acquired,
			}
			adapter := messenger.NewPhase151PostCreateRecoveryMongoAdapter(store)

			acquisition, acquired, err := adapter.ClaimPendingPostCreateRecovery(
				context.Background(), scope, message, ownerID, now, time.Minute,
			)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.False(t, acquired)
			assert.Equal(t, messenger.Phase151PostCreateRecoveryAcquisition{}, acquisition)
		})
	}
}

func TestPhase151PostCreateRecoveryMongoAdapterCompletesOnlyExactLiveClaim(t *testing.T) {
	scope, message, now := phase151RecoveryClaimFixture()
	claim := messenger.Phase151PostCreateRecoveryClaim{
		TenantID: scope.TenantID, MessageID: message.ID, ChannelID: scope.ChannelID,
		OwnerID: uuid.New(), Fence: 5, LeaseUntil: now.Add(2 * time.Minute),
	}
	completedAt := now.Add(time.Minute)
	expected, err := messenger.ExportPhase151BuildPostCreateRecoveryCompletionPlan(
		scope, message, claim, completedAt,
	)
	require.NoError(t, err)

	t.Run("matched completion delegates exact plan", func(t *testing.T) {
		store := &phase151PostCreateRecoveryMongoStore{completeMatch: true}
		adapter := messenger.NewPhase151PostCreateRecoveryMongoAdapter(store)

		err := adapter.CompleteClaimedPostCreateRecovery(context.Background(), scope, message, claim, completedAt)

		require.NoError(t, err)
		require.Equal(t, []bson.D{expected.Filter}, store.completeFilters)
		require.Equal(t, []bson.D{expected.Update}, store.completeUpdates)
		assert.Empty(t, store.claimFilters)
	})

	t.Run("completion miss is lease loss", func(t *testing.T) {
		store := &phase151PostCreateRecoveryMongoStore{}
		adapter := messenger.NewPhase151PostCreateRecoveryMongoAdapter(store)

		err := adapter.CompleteClaimedPostCreateRecovery(context.Background(), scope, message, claim, completedAt)

		require.ErrorIs(t, err, messenger.ErrPostCreateRecoveryLeaseLost)
	})

	t.Run("completion error retains identity", func(t *testing.T) {
		storeErr := errors.New("completion store failed")
		store := &phase151PostCreateRecoveryMongoStore{completeErr: storeErr}
		adapter := messenger.NewPhase151PostCreateRecoveryMongoAdapter(store)

		err := adapter.CompleteClaimedPostCreateRecovery(context.Background(), scope, message, claim, completedAt)

		require.ErrorIs(t, err, storeErr)
	})

	t.Run("invalid or cancelled input stops before store", func(t *testing.T) {
		store := &phase151PostCreateRecoveryMongoStore{}
		adapter := messenger.NewPhase151PostCreateRecoveryMongoAdapter(store)
		invalid := claim
		invalid.Fence = 0
		err := adapter.CompleteClaimedPostCreateRecovery(context.Background(), scope, message, invalid, completedAt)
		require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err = adapter.CompleteClaimedPostCreateRecovery(ctx, scope, message, claim, completedAt)
		require.ErrorIs(t, err, context.Canceled)
		assert.Empty(t, store.completeFilters)
	})
}

func TestPhase151PostCreateRecoveryMongoAdapterRequiresStore(t *testing.T) {
	assert.Panics(t, func() { messenger.NewPhase151PostCreateRecoveryMongoAdapter(nil) })
	var typedNil *phase151PostCreateRecoveryMongoStore
	assert.Panics(t, func() { messenger.NewPhase151PostCreateRecoveryMongoAdapter(typedNil) })

	var adapter *messenger.Phase151PostCreateRecoveryMongoAdapter
	scope, message, now := phase151RecoveryClaimFixture()
	_, acquired, err := adapter.ClaimPendingPostCreateRecovery(
		context.Background(), scope, message, uuid.New(), now, time.Minute,
	)
	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	assert.False(t, acquired)
}
