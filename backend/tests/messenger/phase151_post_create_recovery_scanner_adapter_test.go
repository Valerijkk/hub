//go:build hubtest

package messenger_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

var _ messenger.Phase151PendingRecoveryScanRepository = (*messenger.Phase151PostCreateRecoveryScannerAdapter)(nil)

type phase151RecoveryScannerAdapterPager struct {
	calls int
	ctx   context.Context
	after *messenger.Phase151RecoveryTargetCursor
	limit int
	page  messenger.Phase151PendingRecoveryTargetPage
	err   error
}

func (p *phase151RecoveryScannerAdapterPager) ListPendingRecoveryTargets(
	ctx context.Context,
	after *messenger.Phase151RecoveryTargetCursor,
	limit int,
) (messenger.Phase151PendingRecoveryTargetPage, error) {
	p.calls++
	p.ctx = ctx
	p.after = after
	p.limit = limit
	return p.page, p.err
}

type phase151RecoveryScannerAdapterReauthorizer struct {
	calls     int
	ctx       context.Context
	target    messenger.Phase151PendingRecoveryTarget
	authority messenger.Phase151RecoveryTargetAuthority
	err       error
}

func (r *phase151RecoveryScannerAdapterReauthorizer) ReauthorizePendingRecoveryTarget(
	ctx context.Context,
	target messenger.Phase151PendingRecoveryTarget,
) (messenger.Phase151RecoveryTargetAuthority, error) {
	r.calls++
	r.ctx = ctx
	r.target = target
	return r.authority, r.err
}

type phase151RecoveryScannerAdapterDocuments struct {
	calls     int
	ctx       context.Context
	authority messenger.Phase151RecoveryTargetAuthority
	limit     int
	documents []messenger.MessageDocument
	err       error
}

func (d *phase151RecoveryScannerAdapterDocuments) ListPendingPostCreateDocuments(
	ctx context.Context,
	authority messenger.Phase151RecoveryTargetAuthority,
	limit int,
) ([]messenger.MessageDocument, error) {
	d.calls++
	d.ctx = ctx
	d.authority = authority
	d.limit = limit
	return d.documents, d.err
}

type phase151RecoveryScannerAdapterTypedNilContext struct {
	context.Context
}

var _ context.Context = (*phase151RecoveryScannerAdapterTypedNilContext)(nil)

func TestPhase151PostCreateRecoveryScannerAdapterRejectsNilConstruction(t *testing.T) {
	pager := &phase151RecoveryScannerAdapterPager{}
	reauthorizer := &phase151RecoveryScannerAdapterReauthorizer{}
	documents := &phase151RecoveryScannerAdapterDocuments{}
	var typedNilPager *phase151RecoveryScannerAdapterPager
	var typedNilReauthorizer *phase151RecoveryScannerAdapterReauthorizer
	var typedNilDocuments *phase151RecoveryScannerAdapterDocuments

	assert.Nil(t, messenger.NewPhase151PostCreateRecoveryScannerAdapter(nil, reauthorizer, documents))
	assert.Nil(t, messenger.NewPhase151PostCreateRecoveryScannerAdapter(typedNilPager, reauthorizer, documents))
	assert.Nil(t, messenger.NewPhase151PostCreateRecoveryScannerAdapter(pager, nil, documents))
	assert.Nil(t, messenger.NewPhase151PostCreateRecoveryScannerAdapter(pager, typedNilReauthorizer, documents))
	assert.Nil(t, messenger.NewPhase151PostCreateRecoveryScannerAdapter(pager, reauthorizer, nil))
	assert.Nil(t, messenger.NewPhase151PostCreateRecoveryScannerAdapter(pager, reauthorizer, typedNilDocuments))
	assert.NotNil(t, messenger.NewPhase151PostCreateRecoveryScannerAdapter(pager, reauthorizer, documents))
}

func TestPhase151PostCreateRecoveryScannerAdapterFailsClosedBeforeDependencyAccess(t *testing.T) {
	pager := &phase151RecoveryScannerAdapterPager{}
	reauthorizer := &phase151RecoveryScannerAdapterReauthorizer{}
	documents := &phase151RecoveryScannerAdapterDocuments{}
	adapter := messenger.NewPhase151PostCreateRecoveryScannerAdapter(pager, reauthorizer, documents)
	target := messenger.Phase151PendingRecoveryTarget{TenantID: uuid.New(), ChannelID: uuid.New(), Kind: "unknown-kind"}
	authority := messenger.Phase151RecoveryTargetAuthority{TenantID: uuid.New(), ChannelID: uuid.New(), Kind: "unknown-kind"}
	var typedNilContext *phase151RecoveryScannerAdapterTypedNilContext
	cancelledContext, cancel := context.WithCancel(context.Background())
	cancel()

	for _, testCase := range []struct {
		name string
		ctx  context.Context
		want error
	}{
		{name: "nil context", want: messenger.ErrPostCreatePipelineInvalid},
		{name: "typed nil context", ctx: typedNilContext, want: messenger.ErrPostCreatePipelineInvalid},
		{name: "cancelled context", ctx: cancelledContext, want: context.Canceled},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			page, err := adapter.ListPendingRecoveryTargets(testCase.ctx, nil, 0)
			require.Same(t, testCase.want, err)
			assert.Equal(t, messenger.Phase151PendingRecoveryTargetPage{}, page)

			gotAuthority, err := adapter.ReauthorizePendingRecoveryTarget(testCase.ctx, target)
			require.Same(t, testCase.want, err)
			assert.Equal(t, messenger.Phase151RecoveryTargetAuthority{}, gotAuthority)

			gotDocuments, err := adapter.ListPendingPostCreateDocuments(testCase.ctx, authority, -1)
			require.Same(t, testCase.want, err)
			assert.Nil(t, gotDocuments)
		})
	}

	assert.Zero(t, pager.calls)
	assert.Zero(t, reauthorizer.calls)
	assert.Zero(t, documents.calls)

	var nilAdapter *messenger.Phase151PostCreateRecoveryScannerAdapter
	page, err := nilAdapter.ListPendingRecoveryTargets(cancelledContext, nil, 1)
	require.Same(t, messenger.ErrPostCreatePipelineInvalid, err)
	assert.Equal(t, messenger.Phase151PendingRecoveryTargetPage{}, page)
	gotAuthority, err := nilAdapter.ReauthorizePendingRecoveryTarget(cancelledContext, target)
	require.Same(t, messenger.ErrPostCreatePipelineInvalid, err)
	assert.Equal(t, messenger.Phase151RecoveryTargetAuthority{}, gotAuthority)
	gotDocuments, err := nilAdapter.ListPendingPostCreateDocuments(cancelledContext, authority, 1)
	require.Same(t, messenger.ErrPostCreatePipelineInvalid, err)
	assert.Nil(t, gotDocuments)
	assert.Zero(t, pager.calls)
	assert.Zero(t, reauthorizer.calls)
	assert.Zero(t, documents.calls)
}

func TestPhase151PostCreateRecoveryScannerAdapterForwardsOnlyTheMatchingPortUnchanged(t *testing.T) {
	ctx := context.WithValue(context.Background(), "phase151 recovery scanner adapter", "context identity")
	after := &messenger.Phase151RecoveryTargetCursor{TenantID: uuid.New(), ChannelID: uuid.New(), Kind: "future-kind"}
	target := messenger.Phase151PendingRecoveryTarget{TenantID: uuid.New(), ChannelID: uuid.New(), Kind: "future-kind"}
	authority := messenger.Phase151RecoveryTargetAuthority{TenantID: uuid.New(), ChannelID: uuid.New(), Kind: "future-kind"}
	next := &messenger.Phase151RecoveryTargetCursor{TenantID: uuid.New(), ChannelID: uuid.New(), Kind: "later-kind"}
	wantPage := messenger.Phase151PendingRecoveryTargetPage{Targets: []messenger.Phase151PendingRecoveryTarget{target}, Next: next}
	wantAuthority := authority
	wantDocuments := []messenger.MessageDocument{{ID: "phase151-scanner-adapter-document"}}
	pagerErr := errors.New("phase151 scanner pager sentinel")
	reauthorizerErr := errors.New("phase151 scanner reauthorizer sentinel")
	documentsErr := errors.New("phase151 scanner documents sentinel")
	pager := &phase151RecoveryScannerAdapterPager{page: wantPage, err: pagerErr}
	reauthorizer := &phase151RecoveryScannerAdapterReauthorizer{authority: wantAuthority, err: reauthorizerErr}
	documents := &phase151RecoveryScannerAdapterDocuments{documents: wantDocuments, err: documentsErr}
	adapter := messenger.NewPhase151PostCreateRecoveryScannerAdapter(pager, reauthorizer, documents)

	gotPage, err := adapter.ListPendingRecoveryTargets(ctx, after, 0)
	require.Same(t, pagerErr, err)
	assert.Equal(t, wantPage, gotPage)
	assert.Same(t, &wantPage.Targets[0], &gotPage.Targets[0])
	assert.Same(t, wantPage.Next, gotPage.Next)
	assert.Same(t, ctx, pager.ctx)
	assert.Same(t, after, pager.after)
	assert.Zero(t, pager.limit)
	assert.Equal(t, 1, pager.calls)
	assert.Zero(t, reauthorizer.calls)
	assert.Zero(t, documents.calls)

	gotAuthority, err := adapter.ReauthorizePendingRecoveryTarget(ctx, target)
	require.Same(t, reauthorizerErr, err)
	assert.Equal(t, wantAuthority, gotAuthority)
	assert.Same(t, ctx, reauthorizer.ctx)
	assert.Equal(t, target, reauthorizer.target)
	assert.Equal(t, 1, pager.calls)
	assert.Equal(t, 1, reauthorizer.calls)
	assert.Zero(t, documents.calls)

	gotDocuments, err := adapter.ListPendingPostCreateDocuments(ctx, authority, -1)
	require.Same(t, documentsErr, err)
	assert.Equal(t, wantDocuments, gotDocuments)
	assert.Same(t, &wantDocuments[0], &gotDocuments[0])
	assert.Same(t, ctx, documents.ctx)
	assert.Equal(t, authority, documents.authority)
	assert.Equal(t, -1, documents.limit)
	assert.Equal(t, 1, pager.calls)
	assert.Equal(t, 1, reauthorizer.calls)
	assert.Equal(t, 1, documents.calls)
}
