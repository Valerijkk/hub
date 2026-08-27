package messenger_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

func TestPhase151ExistingCanonicalFingerprintMismatchIsTypedNonceConflict(t *testing.T) {
	for _, scenario := range []struct {
		name          string
		configure     func(*phase151PostCreateResultRecorder, messenger.CreateMessageResult)
		expectedCalls func(messenger.PostCreatePipelineRequest) []string
	}{
		{
			name: "lookup hit",
			configure: func(recorder *phase151PostCreateResultRecorder, canonical messenger.CreateMessageResult) {
				recorder.findFound = true
				recorder.findResult = canonical
			},
			expectedCalls: func(request messenger.PostCreatePipelineRequest) []string {
				return []string{"idempotency:" + request.Message.ClientNonce}
			},
		},
		{
			name: "persist race",
			configure: func(recorder *phase151PostCreateResultRecorder, canonical messenger.CreateMessageResult) {
				recorder.persistResult = &canonical
			},
			expectedCalls: func(request messenger.PostCreatePipelineRequest) []string {
				return []string{
					"idempotency:" + request.Message.ClientNonce,
					"resolve-mentions",
					"persist-pending",
				}
			},
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope, request := phase151PostCreateFingerprintFixture()
			canonical := phase151PostCreateCanonicalResult(
				t, scope, request, false, messenger.PostCreateEffectsStateComplete,
			)
			canonical.Message = phase151MessageDocumentPointer(phase151MessageDocumentWithBSONOverrides(
				t,
				*canonical.Message,
				map[string]interface{}{
					"client_nonce_fingerprint": "0000000000000000000000000000000000000000000000000000000000000000",
				},
			))
			recorder := &phase151PostCreateResultRecorder{
				phase151PostCreateRecorder: newPhase151PostCreateRecorder(),
			}
			scenario.configure(recorder, canonical)

			result, err := messenger.ExportPhase151ExecutePostCreatePipeline(
				context.Background(), scope, request, recorder,
			)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineNonceConflict)
			assert.NotErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Equal(t, messenger.CreateMessageResult{}, result)
			assert.Equal(t, scenario.expectedCalls(request), recorder.calls)
			phase151AssertPostCreateNoEffects(t, recorder.phase151PostCreateRecorder)
		})
	}
}

func TestPhase151ExistingDeletedCanonicalReservesNonceBeforeMetadataValidation(t *testing.T) {
	for _, scenario := range []struct {
		name          string
		configure     func(*phase151PostCreateResultRecorder, messenger.CreateMessageResult)
		expectedCalls func(messenger.PostCreatePipelineRequest) []string
	}{
		{
			name: "lookup hit",
			configure: func(recorder *phase151PostCreateResultRecorder, canonical messenger.CreateMessageResult) {
				recorder.findFound = true
				recorder.findResult = canonical
			},
			expectedCalls: func(request messenger.PostCreatePipelineRequest) []string {
				return []string{"idempotency:" + request.Message.ClientNonce}
			},
		},
		{
			name: "persist race",
			configure: func(recorder *phase151PostCreateResultRecorder, canonical messenger.CreateMessageResult) {
				recorder.persistResult = &canonical
			},
			expectedCalls: func(request messenger.PostCreatePipelineRequest) []string {
				return []string{
					"idempotency:" + request.Message.ClientNonce,
					"resolve-mentions",
					"persist-pending",
				}
			},
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope, request := phase151PostCreateFingerprintFixture()
			canonical := phase151PostCreateCanonicalResult(
				t, scope, request, false, messenger.PostCreateEffectsStateComplete,
			)
			deletedAt := time.Date(2026, time.August, 27, 5, 0, 0, 0, time.UTC)
			canonical.Message.DeletedAt = &deletedAt
			canonical.Message = phase151MessageDocumentPointer(phase151MessageDocumentWithBSONOverrides(
				t,
				*canonical.Message,
				map[string]interface{}{
					"client_nonce_fingerprint": nil,
					"client_nonce_fingerprint_version": nil,
				},
			))
			recorder := &phase151PostCreateResultRecorder{
				phase151PostCreateRecorder: newPhase151PostCreateRecorder(),
			}
			scenario.configure(recorder, canonical)

			result, err := messenger.ExportPhase151ExecutePostCreatePipeline(
				context.Background(), scope, request, recorder,
			)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineNonceDeleted)
			assert.NotErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.NotErrorIs(t, err, messenger.ErrPostCreatePipelineNonceConflict)
			assert.Equal(t, messenger.CreateMessageResult{}, result)
			assert.Equal(t, scenario.expectedCalls(request), recorder.calls)
			phase151AssertPostCreateNoEffects(t, recorder.phase151PostCreateRecorder)
		})
	}
}

func TestPhase151NewlyCreatedDeletedCanonicalIsRepositoryInvariantFailure(t *testing.T) {
	scope, request := phase151PostCreateFingerprintFixture()
	canonical := phase151PostCreateCanonicalResult(
		t, scope, request, true, messenger.PostCreateEffectsStatePending,
	)
	deletedAt := time.Date(2026, time.August, 27, 6, 0, 0, 0, time.UTC)
	canonical.Message.DeletedAt = &deletedAt
	recorder := &phase151PostCreateResultRecorder{
		phase151PostCreateRecorder: newPhase151PostCreateRecorder(),
		persistResult: &canonical,
	}

	result, err := messenger.ExportPhase151ExecutePostCreatePipeline(
		context.Background(), scope, request, recorder,
	)

	require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	assert.NotErrorIs(t, err, messenger.ErrPostCreatePipelineNonceDeleted)
	assert.NotErrorIs(t, err, messenger.ErrPostCreatePipelineNonceConflict)
	assert.Equal(t, messenger.CreateMessageResult{}, result)
	assert.Equal(t, []string{
		"idempotency:" + request.Message.ClientNonce,
		"resolve-mentions",
		"persist-pending",
	}, recorder.calls)
	phase151AssertPostCreateNoEffects(t, recorder.phase151PostCreateRecorder)
}

func TestPhase151DeletedLookupStillRejectsStructuralInvalidityBeforeNonceOutcome(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		mutate func(*messenger.CreateMessageResult)
	}{
		{
			name: "lookup cannot report created",
			mutate: func(canonical *messenger.CreateMessageResult) {
				canonical.Created = true
			},
		},
		{
			name: "empty canonical id",
			mutate: func(canonical *messenger.CreateMessageResult) {
				canonical.Message.ID = ""
			},
		},
		{
			name: "unknown canonical state",
			mutate: func(canonical *messenger.CreateMessageResult) {
				canonical.Message.PostCreateEffectsState = "unknown"
				canonical.PostCreateEffectsState = "unknown"
			},
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope, request := phase151PostCreateFingerprintFixture()
			canonical := phase151PostCreateCanonicalResult(
				t, scope, request, false, messenger.PostCreateEffectsStateComplete,
			)
			deletedAt := time.Date(2026, time.August, 27, 7, 0, 0, 0, time.UTC)
			canonical.Message.DeletedAt = &deletedAt
			scenario.mutate(&canonical)
			recorder := &phase151PostCreateResultRecorder{
				phase151PostCreateRecorder: newPhase151PostCreateRecorder(),
				findFound:                   true,
				findResult:                  canonical,
			}

			result, err := messenger.ExportPhase151ExecutePostCreatePipeline(
				context.Background(), scope, request, recorder,
			)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.NotErrorIs(t, err, messenger.ErrPostCreatePipelineNonceDeleted)
			assert.NotErrorIs(t, err, messenger.ErrPostCreatePipelineNonceConflict)
			assert.Equal(t, messenger.CreateMessageResult{}, result)
			assert.Equal(t, []string{"idempotency:" + request.Message.ClientNonce}, recorder.calls)
			phase151AssertPostCreateNoEffects(t, recorder.phase151PostCreateRecorder)
		})
	}
}
