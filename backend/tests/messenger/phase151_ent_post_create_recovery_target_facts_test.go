//go:build hubtest

package messenger_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
)

var _ messenger.Phase151PendingRecoveryTargetFactSource = (*messenger.EntRepository)(nil)

type phase151EntRecoveryTargetTypedNilContext struct {
	context.Context
}

var _ context.Context = (*phase151EntRecoveryTargetTypedNilContext)(nil)

func TestEntRepositoryPendingRecoveryTargetFactsFailsClosedBeforeEntQuery(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	channelID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	validRepository := messenger.NewEntRepository(&ent.Client{})
	nilClientRepository := messenger.NewEntRepository(nil)
	var nilRepository *messenger.EntRepository
	var typedNilContext *phase151EntRecoveryTargetTypedNilContext
	cancelledContext, cancel := context.WithCancel(context.Background())
	cancel()

	for _, testCase := range []struct {
		name   string
		invoke func() (messenger.Phase151PendingRecoveryTargetFacts, error)
		want   error
	}{
		{
			name: "nil receiver",
			invoke: func() (messenger.Phase151PendingRecoveryTargetFacts, error) {
				return nilRepository.ResolvePendingRecoveryTargetFacts(context.Background(), tenantID, channelID)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "nil client",
			invoke: func() (messenger.Phase151PendingRecoveryTargetFacts, error) {
				return nilClientRepository.ResolvePendingRecoveryTargetFacts(context.Background(), tenantID, channelID)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "nil context",
			invoke: func() (messenger.Phase151PendingRecoveryTargetFacts, error) {
				return validRepository.ResolvePendingRecoveryTargetFacts(nil, tenantID, channelID)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "typed nil context",
			invoke: func() (messenger.Phase151PendingRecoveryTargetFacts, error) {
				return validRepository.ResolvePendingRecoveryTargetFacts(typedNilContext, tenantID, channelID)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "zero tenant",
			invoke: func() (messenger.Phase151PendingRecoveryTargetFacts, error) {
				return validRepository.ResolvePendingRecoveryTargetFacts(context.Background(), uuid.Nil, channelID)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "zero channel",
			invoke: func() (messenger.Phase151PendingRecoveryTargetFacts, error) {
				return validRepository.ResolvePendingRecoveryTargetFacts(context.Background(), tenantID, uuid.Nil)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "cancelled context with valid IDs",
			invoke: func() (messenger.Phase151PendingRecoveryTargetFacts, error) {
				return validRepository.ResolvePendingRecoveryTargetFacts(cancelledContext, tenantID, channelID)
			},
			want: context.Canceled,
		},
		{
			name: "cancelled context wins before zero UUID validation",
			invoke: func() (messenger.Phase151PendingRecoveryTargetFacts, error) {
				return validRepository.ResolvePendingRecoveryTargetFacts(cancelledContext, uuid.Nil, uuid.Nil)
			},
			want: context.Canceled,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var facts messenger.Phase151PendingRecoveryTargetFacts
			var err error

			require.NotPanics(t, func() {
				facts, err = testCase.invoke()
			})
			require.Equal(t, messenger.Phase151PendingRecoveryTargetFacts{}, facts)
			require.Same(t, testCase.want, err)
		})
	}
}
