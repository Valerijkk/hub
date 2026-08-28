//go:build hubtest

package messenger_test

import (
	"context"
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151EntRecoveryTargetPager interface {
	ListPendingRecoveryTargets(
		context.Context,
		*messenger.Phase151RecoveryTargetCursor,
		int,
	) (messenger.Phase151PendingRecoveryTargetPage, error)
}

var _ phase151EntRecoveryTargetPager = (*messenger.EntRepository)(nil)

type phase151EntRecoveryTargetPagerTypedNilContext struct {
	context.Context
}

var _ context.Context = (*phase151EntRecoveryTargetPagerTypedNilContext)(nil)

func TestEntRepositoryPendingRecoveryTargetPagerFailsClosedBeforeSingletonOrEntQuery(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	channelID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	validRepository := messenger.NewEntRepository(&ent.Client{})
	nilClientRepository := messenger.NewEntRepository(nil)
	validCursor := &messenger.Phase151RecoveryTargetCursor{
		TenantID: tenantID, ChannelID: channelID, Kind: "channel",
	}
	var nilRepository *messenger.EntRepository
	var typedNilContext *phase151EntRecoveryTargetPagerTypedNilContext
	cancelledContext, cancel := context.WithCancel(context.Background())
	cancel()

	for _, testCase := range []struct {
		name   string
		invoke func() (messenger.Phase151PendingRecoveryTargetPage, error)
		want   error
	}{
		{
			name: "nil receiver",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return nilRepository.ListPendingRecoveryTargets(context.Background(), validCursor, 1)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "nil client",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return nilClientRepository.ListPendingRecoveryTargets(context.Background(), validCursor, 1)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "nil context",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(nil, validCursor, 1)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "typed nil context",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(typedNilContext, validCursor, 1)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "cancelled context",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(cancelledContext, validCursor, 1)
			},
			want: context.Canceled,
		},
		{
			name: "cancelled context wins over malformed limit",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(cancelledContext, validCursor, 0)
			},
			want: context.Canceled,
		},
		{
			name: "cancelled context wins over malformed cursor",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(cancelledContext, &messenger.Phase151RecoveryTargetCursor{}, 1)
			},
			want: context.Canceled,
		},
		{
			name: "zero limit",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(context.Background(), validCursor, 0)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "negative limit",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(context.Background(), validCursor, -1)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "overflowing limit",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(context.Background(), validCursor, math.MaxInt)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "cursor without tenant",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(context.Background(), &messenger.Phase151RecoveryTargetCursor{ChannelID: channelID, Kind: "channel"}, 1)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "cursor without channel",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(context.Background(), &messenger.Phase151RecoveryTargetCursor{TenantID: tenantID, Kind: "channel"}, 1)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "cursor empty kind",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(context.Background(), &messenger.Phase151RecoveryTargetCursor{TenantID: tenantID, ChannelID: channelID}, 1)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "cursor whitespace kind",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(context.Background(), &messenger.Phase151RecoveryTargetCursor{TenantID: tenantID, ChannelID: channelID, Kind: " "}, 1)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "cursor padded kind",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(context.Background(), &messenger.Phase151RecoveryTargetCursor{TenantID: tenantID, ChannelID: channelID, Kind: " channel"}, 1)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "cursor case drifted kind",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(context.Background(), &messenger.Phase151RecoveryTargetCursor{TenantID: tenantID, ChannelID: channelID, Kind: "Channel"}, 1)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name: "cursor unknown kind",
			invoke: func() (messenger.Phase151PendingRecoveryTargetPage, error) {
				return validRepository.ListPendingRecoveryTargets(context.Background(), &messenger.Phase151RecoveryTargetCursor{TenantID: tenantID, ChannelID: channelID, Kind: "future_kind"}, 1)
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var page messenger.Phase151PendingRecoveryTargetPage
			var err error

			require.NotPanics(t, func() {
				page, err = testCase.invoke()
			})
			require.Equal(t, messenger.Phase151PendingRecoveryTargetPage{}, page)
			require.Same(t, testCase.want, err)
		})
	}
}
