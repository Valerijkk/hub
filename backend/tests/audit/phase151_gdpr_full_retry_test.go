package audit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/audit"
)

type phase151MessengerEraseRetryRecorder struct {
	steps   *[]string
	users   []uuid.UUID
	actors  []uuid.UUID
	err     error
}

func (r *phase151MessengerEraseRetryRecorder) EnqueueMessengerErase(
	_ context.Context,
	userID uuid.UUID,
	actorID uuid.UUID,
) error {
	r.users = append(r.users, userID)
	r.actors = append(r.actors, actorID)
	if r.steps != nil {
		*r.steps = append(*r.steps, "messenger_retry:"+userID.String())
	}
	return r.err
}

func TestPhase151DeleteUserGDPR_FullFailureEnqueuesOnlyFullRetryAfterCommit(t *testing.T) {
	steps := []string{}
	repo := &phase151GDPRCommitOrderRepo{steps: &steps}
	full := &phase151GDPRMessengerDataEraser{steps: &steps, err: errGDPRCascadeStep}
	legacy := &stubMessageEraser{}
	retry := &phase151MessengerEraseRetryRecorder{steps: &steps}
	service := audit.NewAuditService(repo, nil)
	service.SetMessageEraser(legacy)
	service.SetMessengerDataEraser(full)
	service.SetMessengerEraseRetry(retry)
	targetID := uuid.New()
	actorID := uuid.New()

	err := service.DeleteUserGDPR(context.Background(), actorID, targetID, nil)

	require.ErrorIs(t, err, errGDPRCascadeStep)
	assert.Empty(t, legacy.calls, "a full erase failure must never downgrade to sender-only deletion")
	assert.Equal(t, []uuid.UUID{targetID}, retry.users)
	assert.Equal(t, []uuid.UUID{actorID}, retry.actors)
	assert.Equal(t, []string{
		"ent_commit:" + targetID.String(),
		"messenger_erase:" + targetID.String(),
		"messenger_retry:" + targetID.String(),
	}, steps)
}

func TestPhase151DeleteUserGDPR_FullRetryEnqueueFailureIsReturned(t *testing.T) {
	steps := []string{}
	repo := &phase151GDPRCommitOrderRepo{steps: &steps}
	full := &phase151GDPRMessengerDataEraser{steps: &steps, err: errGDPRCascadeStep}
	retryErr := errors.New("retry unavailable")
	retry := &phase151MessengerEraseRetryRecorder{steps: &steps, err: retryErr}
	service := audit.NewAuditService(repo, nil)
	service.SetMessengerDataEraser(full)
	service.SetMessengerEraseRetry(retry)

	err := service.DeleteUserGDPR(context.Background(), uuid.New(), uuid.New(), nil)

	require.Error(t, err)
	assert.ErrorIs(t, err, errGDPRCascadeStep)
	assert.Contains(t, err.Error(), retryErr.Error())
}

func TestPhase151DispatchMessengerEraseRetry_ExactVersionedOperation(t *testing.T) {
	steps := []string{}
	full := &phase151GDPRMessengerDataEraser{steps: &steps}
	targetID := uuid.New()
	event := audit.GDPRMessengerEraseRetryEvent{
		Version: audit.GDPRMessengerEraseEventVersion,
		Operation: audit.GDPRMessengerEraseOperation,
		SubjectID: targetID,
		ActorID: uuid.New(),
	}

	require.NoError(t, audit.ExportPhase151DispatchMessengerEraseRetry(context.Background(), event, full))
	assert.Equal(t, []uuid.UUID{targetID}, full.calls)
}

func TestPhase151LegacyMongoEraseRetry_V0PayloadCallsOnlyLegacy(t *testing.T) {
	legacy := &stubMessageEraser{}
	targetID := uuid.New()

	deleted, err := audit.ExportPhase151DispatchLegacyMongoEraseRetry(
		context.Background(),
		audit.MongoEraseRetryEvent{SenderID: targetID, ActorID: uuid.New()},
		legacy,
	)

	require.NoError(t, err)
	assert.EqualValues(t, 3, deleted)
	assert.Equal(t, []uuid.UUID{targetID}, legacy.calls)
}


func TestPhase151DispatchMessengerEraseRetry_InvalidSelectorCallsNoEraser(t *testing.T) {
	tests := []struct {
		name  string
		event audit.GDPRMessengerEraseRetryEvent
	}{
		{name: "missing version", event: audit.GDPRMessengerEraseRetryEvent{Operation: audit.GDPRMessengerEraseOperation, SubjectID: uuid.New()}},
		{name: "unknown version", event: audit.GDPRMessengerEraseRetryEvent{Version: 2, Operation: audit.GDPRMessengerEraseOperation, SubjectID: uuid.New()}},
		{name: "unknown operation", event: audit.GDPRMessengerEraseRetryEvent{Version: audit.GDPRMessengerEraseEventVersion, Operation: "legacy", SubjectID: uuid.New()}},
		{name: "empty subject", event: audit.GDPRMessengerEraseRetryEvent{Version: audit.GDPRMessengerEraseEventVersion, Operation: audit.GDPRMessengerEraseOperation}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := []string{}
			full := &phase151GDPRMessengerDataEraser{steps: &steps}

			err := audit.ExportPhase151DispatchMessengerEraseRetry(context.Background(), tt.event, full)

			assert.ErrorIs(t, err, audit.ErrInvalidGDPRMessengerEraseEvent)
			assert.Empty(t, full.calls)
		})
	}
}
