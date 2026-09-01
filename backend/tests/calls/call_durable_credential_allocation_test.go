//go:build hubtest

package calls_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	entcall "github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/internal/calls"
)

var _ calls.DurableCallJoinCredentialAllocationSource = (*durableAllocationSource)(nil)
var _ calls.DurableCallJoinCredentialAllocationTx = (*durableAllocationTx)(nil)
var _ calls.DurableCallJoinCredentialIssuer = (*durableAllocationIssuer)(nil)
var _ func(calls.DurableCallJoinCredentialAllocationSource, calls.DurableCallJoinCredentialIssuer) (*calls.DurableCallJoinCredentialAllocator, error) = calls.NewDurableCallJoinCredentialAllocator
var _ interface {
	IssueForAuthenticatedUser(context.Context, uuid.UUID, uuid.UUID) (calls.JoinCredential, error)
} = (*calls.DurableCallJoinCredentialAllocator)(nil)

func TestDurableCredentialAllocatorCommitsBeforeExposingCredential(t *testing.T) {
	trace := &durableAllocationTrace{}
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	channelID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	creatorID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	ctx := context.WithValue(context.Background(), durableAllocationContextKey{}, "happy")
	lockedCall := &ent.Call{
		ID:        callID,
		RoomName:  "durable-room",
		ChannelID: &channelID,
		CreatedBy: creatorID,
		Status:    entcall.StatusActive,
	}
	tx := &durableAllocationTx{
		trace: trace,
		name:  " Ada Lovelace ",
		call:  lockedCall,
	}
	issuer := &durableAllocationIssuer{
		trace:      trace,
		credential: calls.JoinCredential{Token: "opaque"},
	}
	source := &durableAllocationSource{trace: trace, tx: tx}
	allocator := newDurableCredentialAllocator(t, source, issuer)

	credential, err := allocator.IssueForAuthenticatedUser(ctx, callID, userID)
	require.NoError(t, err)
	require.Equal(t, issuer.credential, credential)
	require.Equal(t, []string{"begin", "lock_user", "lock_call", "authorize", "create", "issue", "commit"}, trace.stages)
	require.Equal(t, 1, tx.createCalls)
	require.Equal(t, 1, issuer.calls)
	require.Equal(t, 1, tx.commitCalls)
	require.Zero(t, tx.rollbackCalls)
	require.Equal(t, callID, tx.createdKey.CallID())
	require.Equal(t, userID, tx.createdKey.UserID())
	require.NotEqual(t, uuid.Nil, tx.createdKey.Epoch())
	require.True(t, tx.createdKey.IsValid())
	require.Equal(t, tx.createdKey, issuer.identity.ConnectionKey())
	require.Equal(t, "hub.call-connection.v1/"+callID.String()+"/"+userID.String()+"/"+tx.createdKey.Epoch().String(), tx.createdKey.Canonical())
	require.Equal(t, "hub.call-participant.v1/"+callID.String()+"/"+userID.String()+"/"+tx.createdKey.Epoch().String(), issuer.identity.Canonical())
	require.NotEqual(t, tx.createdKey.Canonical(), issuer.identity.Canonical())
	require.Same(t, tx.call, issuer.call)
	require.Equal(t, "Ada Lovelace", issuer.userName)
	require.Same(t, ctx, source.beginCtx)
	require.Same(t, ctx, tx.lockUserCtx)
	require.Equal(t, userID, tx.lockUserID)
	require.Same(t, ctx, tx.lockCallCtx)
	require.Equal(t, callID, tx.lockCallID)
	require.Same(t, ctx, tx.authorizeCtx)
	require.Same(t, lockedCall, tx.authorizedCall)
	require.Equal(t, userID, tx.authorizeUserID)
	require.Same(t, ctx, tx.createCtx)
}

func TestDurableCredentialAllocatorRejectsInvalidBoundaryBeforeBegin(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	require.Equal(t, "call join credential unavailable", calls.ErrCallJoinCredentialUnavailable.Error())
	require.Equal(t, "call join credential commit outcome ambiguous", calls.ErrCallJoinCredentialCommitAmbiguous.Error())
	validAllocator := func() (*calls.DurableCallJoinCredentialAllocator, *durableAllocationSource) {
		trace := &durableAllocationTrace{}
		source := &durableAllocationSource{trace: trace, tx: &durableAllocationTx{trace: trace, name: "Ada", call: durableAllocationCall(callID, userID)}}
		return newDurableCredentialAllocator(t, source, &durableAllocationIssuer{trace: trace, credential: calls.JoinCredential{Token: "opaque"}}), source
	}

	t.Run("constructor rejects nil and typed nil dependencies", func(t *testing.T) {
		issuer := &durableAllocationIssuer{}
		_, err := calls.NewDurableCallJoinCredentialAllocator(nil, issuer)
		requireExactAllocationError(t, calls.ErrCallJoinCredentialInvalid, err, calls.JoinCredential{}, nil, nil)
		var typedNilSource *durableAllocationSource
		_, err = calls.NewDurableCallJoinCredentialAllocator(typedNilSource, issuer)
		requireExactAllocationError(t, calls.ErrCallJoinCredentialInvalid, err, calls.JoinCredential{}, nil, nil)
		_, err = calls.NewDurableCallJoinCredentialAllocator(&durableAllocationSource{}, nil)
		requireExactAllocationError(t, calls.ErrCallJoinCredentialInvalid, err, calls.JoinCredential{}, nil, nil)
		var typedNilIssuer *durableAllocationIssuer
		_, err = calls.NewDurableCallJoinCredentialAllocator(&durableAllocationSource{}, typedNilIssuer)
		requireExactAllocationError(t, calls.ErrCallJoinCredentialInvalid, err, calls.JoinCredential{}, nil, nil)
	})

	var typedNilContext *durableTypedNilContext
	tests := []struct {
		name string
		call uuid.UUID
		user uuid.UUID
		ctx  context.Context
	}{
		{name: "nil context", call: callID, user: userID, ctx: nil},
		{name: "typed nil context", call: callID, user: userID, ctx: typedNilContext},
		{name: "zero call", call: uuid.Nil, user: userID, ctx: context.Background()},
		{name: "zero user", call: callID, user: uuid.Nil, ctx: context.Background()},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			allocator, source := validAllocator()
			credential, err := allocator.IssueForAuthenticatedUser(testCase.ctx, testCase.call, testCase.user)
			requireExactAllocationError(t, calls.ErrCallJoinCredentialInvalid, err, credential, nil, nil)
			require.Zero(t, source.begins)
		})
	}

	t.Run("already cancelled context", func(t *testing.T) {
		allocator, source := validAllocator()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		credential, err := allocator.IssueForAuthenticatedUser(ctx, callID, userID)
		requireExactAllocationError(t, calls.ErrCallJoinCredentialUnavailable, err, credential, ctx.Err(), nil)
		require.Zero(t, source.begins)
	})
}

func TestDurableCredentialAllocatorRejectsUnusableTransactionWithoutRollback(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	for _, testCase := range []struct {
		name string
		tx   calls.DurableCallJoinCredentialAllocationTx
	}{
		{name: "nil transaction", tx: nil},
		{name: "typed nil transaction", tx: (*durableAllocationTx)(nil)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			trace := &durableAllocationTrace{}
			source := &durableAllocationSource{trace: trace, tx: testCase.tx}
			allocator := newDurableCredentialAllocator(t, source, &durableAllocationIssuer{trace: trace})
			credential, err := allocator.IssueForAuthenticatedUser(context.Background(), callID, userID)
			requireExactAllocationError(t, calls.ErrCallJoinCredentialUnavailable, err, credential, nil, nil)
			require.Equal(t, []string{"begin"}, trace.stages)
			require.Equal(t, 1, source.begins)
		})
	}
}

func TestDurableCredentialAllocatorRejectsMalformedLockedFacts(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	wrongCallID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	zeroChannelID := uuid.Nil
	tests := []struct {
		name          string
		lockedName    string
		lockedCall    *ent.Call
		expectedTrace []string
	}{
		{name: "empty locked name", lockedName: "", lockedCall: durableAllocationCall(callID, userID), expectedTrace: []string{"begin", "lock_user", "rollback"}},
		{name: "whitespace locked name", lockedName: " \t\r\n ", lockedCall: durableAllocationCall(callID, userID), expectedTrace: []string{"begin", "lock_user", "rollback"}},
		{name: "nil locked call", lockedName: "Ada", lockedCall: nil, expectedTrace: []string{"begin", "lock_user", "lock_call", "rollback"}},
		{name: "zero locked call ID", lockedName: "Ada", lockedCall: &ent.Call{RoomName: "durable-room", CreatedBy: userID, Status: entcall.StatusActive}, expectedTrace: []string{"begin", "lock_user", "lock_call", "rollback"}},
		{name: "wrong locked call ID", lockedName: "Ada", lockedCall: durableAllocationCall(wrongCallID, userID), expectedTrace: []string{"begin", "lock_user", "lock_call", "rollback"}},
		{name: "inactive locked call", lockedName: "Ada", lockedCall: &ent.Call{ID: callID, RoomName: "durable-room", CreatedBy: userID}, expectedTrace: []string{"begin", "lock_user", "lock_call", "rollback"}},
		{name: "ended locked call", lockedName: "Ada", lockedCall: &ent.Call{ID: callID, RoomName: "durable-room", CreatedBy: userID, Status: entcall.StatusEnded}, expectedTrace: []string{"begin", "lock_user", "lock_call", "rollback"}},
		{name: "empty locked room", lockedName: "Ada", lockedCall: &ent.Call{ID: callID, CreatedBy: userID, Status: entcall.StatusActive}, expectedTrace: []string{"begin", "lock_user", "lock_call", "rollback"}},
		{name: "whitespace locked room", lockedName: "Ada", lockedCall: &ent.Call{ID: callID, RoomName: " \t\r\n ", CreatedBy: userID, Status: entcall.StatusActive}, expectedTrace: []string{"begin", "lock_user", "lock_call", "rollback"}},
		{name: "padded locked room", lockedName: "Ada", lockedCall: &ent.Call{ID: callID, RoomName: " durable-room ", CreatedBy: userID, Status: entcall.StatusActive}, expectedTrace: []string{"begin", "lock_user", "lock_call", "rollback"}},
		{name: "zero channel origin", lockedName: "Ada", lockedCall: &ent.Call{ID: callID, RoomName: "durable-room", CreatedBy: userID, ChannelID: &zeroChannelID, Status: entcall.StatusActive}, expectedTrace: []string{"begin", "lock_user", "lock_call", "rollback"}},
		{name: "missing instant creator", lockedName: "Ada", lockedCall: &ent.Call{ID: callID, RoomName: "durable-room", Status: entcall.StatusActive}, expectedTrace: []string{"begin", "lock_user", "lock_call", "rollback"}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			trace := &durableAllocationTrace{}
			tx := &durableAllocationTx{trace: trace, name: testCase.lockedName, call: testCase.lockedCall, rollbackErr: errors.New("rollback cause")}
			allocator := newDurableCredentialAllocator(t, &durableAllocationSource{trace: trace, tx: tx}, &durableAllocationIssuer{trace: trace})
			credential, err := allocator.IssueForAuthenticatedUser(context.Background(), callID, userID)
			requireExactAllocationError(t, calls.ErrCallJoinCredentialInvalid, err, credential, nil, tx.rollbackErr)
			require.Equal(t, testCase.expectedTrace, trace.stages)
			require.Equal(t, 1, tx.rollbackCalls)
		})
	}
}

func TestDurableCredentialAllocatorMapsDependencyFailuresAndRollsBack(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	tests := []struct {
		name          string
		configure     func(*durableAllocationSource, *durableAllocationTx, *durableAllocationIssuer, error)
		want           error
		expectedTrace []string
		wantRollback  int
	}{
		{
			name: "begin failure",
			configure: func(source *durableAllocationSource, _ *durableAllocationTx, _ *durableAllocationIssuer, cause error) {
				source.err = cause
			},
			want: calls.ErrCallJoinCredentialUnavailable, expectedTrace: []string{"begin"}, wantRollback: 0,
		},
		{
			name: "lock user failure",
			configure: func(_ *durableAllocationSource, tx *durableAllocationTx, _ *durableAllocationIssuer, cause error) { tx.nameErr = cause },
			want: calls.ErrCallJoinCredentialUnavailable, expectedTrace: []string{"begin", "lock_user", "rollback"}, wantRollback: 1,
		},
		{
			name: "lock call failure",
			configure: func(_ *durableAllocationSource, tx *durableAllocationTx, _ *durableAllocationIssuer, cause error) { tx.callErr = cause },
			want: calls.ErrCallJoinCredentialUnavailable, expectedTrace: []string{"begin", "lock_user", "lock_call", "rollback"}, wantRollback: 1,
		},
		{
			name: "forbidden authorization is exact",
			configure: func(_ *durableAllocationSource, tx *durableAllocationTx, _ *durableAllocationIssuer, _ error) { tx.authorizeErr = fmt.Errorf("wrapped: %w", calls.ErrCallForbidden) },
			want: calls.ErrCallForbidden, expectedTrace: []string{"begin", "lock_user", "lock_call", "authorize", "rollback"}, wantRollback: 1,
		},
		{
			name: "other authorization failure",
			configure: func(_ *durableAllocationSource, tx *durableAllocationTx, _ *durableAllocationIssuer, cause error) { tx.authorizeErr = cause },
			want: calls.ErrCallJoinCredentialUnavailable, expectedTrace: []string{"begin", "lock_user", "lock_call", "authorize", "rollback"}, wantRollback: 1,
		},
		{
			name: "unknown connection failure",
			configure: func(_ *durableAllocationSource, tx *durableAllocationTx, _ *durableAllocationIssuer, cause error) { tx.createErr = cause },
			want: calls.ErrCallJoinCredentialUnavailable, expectedTrace: []string{"begin", "lock_user", "lock_call", "authorize", "create", "rollback"}, wantRollback: 1,
		},
		{
			name: "issuer failure suppresses partial credential",
			configure: func(_ *durableAllocationSource, _ *durableAllocationTx, issuer *durableAllocationIssuer, cause error) {
				issuer.credential = calls.JoinCredential{Token: "partial-credential"}
				issuer.err = cause
			},
			want: calls.ErrCallJoinCredentialUnavailable, expectedTrace: []string{"begin", "lock_user", "lock_call", "authorize", "create", "issue", "rollback"}, wantRollback: 1,
		},
		{
			name: "zero issuer credential is unavailable",
			configure: func(_ *durableAllocationSource, _ *durableAllocationTx, issuer *durableAllocationIssuer, _ error) {
				issuer.credential = calls.JoinCredential{}
				issuer.err = nil
			},
			want: calls.ErrCallJoinCredentialUnavailable, expectedTrace: []string{"begin", "lock_user", "lock_call", "authorize", "create", "issue", "rollback"}, wantRollback: 1,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			trace := &durableAllocationTrace{}
			cause := errors.New("raw dependency cause")
			tx := &durableAllocationTx{trace: trace, name: "Ada", call: durableAllocationCall(callID, userID), rollbackErr: errors.New("raw rollback cause")}
			issuer := &durableAllocationIssuer{trace: trace, credential: calls.JoinCredential{Token: "opaque"}}
			source := &durableAllocationSource{trace: trace, tx: tx}
			testCase.configure(source, tx, issuer, cause)
			allocator := newDurableCredentialAllocator(t, source, issuer)
			credential, err := allocator.IssueForAuthenticatedUser(context.Background(), callID, userID)
			rawCause := cause
			if testCase.want == calls.ErrCallForbidden {
				rawCause = tx.authorizeErr
			}
			requireExactAllocationError(t, testCase.want, err, credential, rawCause, tx.rollbackErr)
			require.Equal(t, testCase.expectedTrace, trace.stages)
			require.Equal(t, testCase.wantRollback, tx.rollbackCalls)
			require.Zero(t, tx.commitCalls)
		})
	}
}

func TestDurableCredentialAllocatorChecksContextBeforeCommitAndPreservesAmbiguity(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")

	t.Run("cancellation after issuer rolls back before commit", func(t *testing.T) {
		trace := &durableAllocationTrace{}
		tx := &durableAllocationTx{trace: trace, name: "Ada", call: durableAllocationCall(callID, userID), rollbackErr: errors.New("rollback cause")}
		ctx, cancel := context.WithCancel(context.Background())
		issuer := &durableAllocationIssuer{trace: trace, credential: calls.JoinCredential{Token: "opaque"}, afterIssue: cancel}
		allocator := newDurableCredentialAllocator(t, &durableAllocationSource{trace: trace, tx: tx}, issuer)
		credential, err := allocator.IssueForAuthenticatedUser(ctx, callID, userID)
		requireExactAllocationError(t, calls.ErrCallJoinCredentialUnavailable, err, credential, ctx.Err(), tx.rollbackErr)
		require.Equal(t, []string{"begin", "lock_user", "lock_call", "authorize", "create", "issue", "rollback"}, trace.stages)
		require.Equal(t, 1, tx.rollbackCalls)
		require.Zero(t, tx.commitCalls)
	})

	t.Run("commit failure is ambiguous without rollback retry or reissue", func(t *testing.T) {
		trace := &durableAllocationTrace{}
		commitCause := errors.New("commit cause")
		tx := &durableAllocationTx{trace: trace, name: "Ada", call: durableAllocationCall(callID, userID), commitErr: commitCause}
		issuer := &durableAllocationIssuer{trace: trace, credential: calls.JoinCredential{Token: "opaque"}}
		allocator := newDurableCredentialAllocator(t, &durableAllocationSource{trace: trace, tx: tx}, issuer)
		credential, err := allocator.IssueForAuthenticatedUser(context.Background(), callID, userID)
		requireExactAllocationError(t, calls.ErrCallJoinCredentialCommitAmbiguous, err, credential, commitCause, nil)
		require.Equal(t, []string{"begin", "lock_user", "lock_call", "authorize", "create", "issue", "commit"}, trace.stages)
		require.Equal(t, 1, issuer.calls)
		require.Equal(t, 1, tx.commitCalls)
		require.Zero(t, tx.rollbackCalls)
	})
}

func newDurableCredentialAllocator(
	t *testing.T,
	source calls.DurableCallJoinCredentialAllocationSource,
	issuer calls.DurableCallJoinCredentialIssuer,
) *calls.DurableCallJoinCredentialAllocator {
	t.Helper()
	allocator, err := calls.NewDurableCallJoinCredentialAllocator(source, issuer)
	require.NoError(t, err)
	return allocator
}

func durableAllocationCall(callID, userID uuid.UUID) *ent.Call {
	return &ent.Call{ID: callID, RoomName: "durable-room", CreatedBy: userID, Status: entcall.StatusActive}
}

func requireExactAllocationError(t *testing.T, want, got error, credential calls.JoinCredential, rawCause, rollbackCause error) {
	t.Helper()
	require.Same(t, want, got)
	require.Equal(t, want.Error(), got.Error())
	require.Equal(t, calls.JoinCredential{}, credential)
	if rawCause != nil {
		require.False(t, errors.Is(got, rawCause))
	}
	if rollbackCause != nil {
		require.False(t, errors.Is(got, rollbackCause))
	}
}

type durableAllocationTrace struct {
	stages []string
}

func (t *durableAllocationTrace) add(stage string) {
	t.stages = append(t.stages, stage)
}

type durableAllocationSource struct {
	trace  *durableAllocationTrace
	tx     calls.DurableCallJoinCredentialAllocationTx
	err    error
	begins int
	beginCtx context.Context
}

func (s *durableAllocationSource) BeginDurableCallJoinCredentialAllocation(ctx context.Context) (calls.DurableCallJoinCredentialAllocationTx, error) {
	s.begins++
	s.beginCtx = ctx
	s.trace.add("begin")
	return s.tx, s.err
}

type durableAllocationTx struct {
	trace        *durableAllocationTrace
	name         string
	nameErr      error
	call         *ent.Call
	callErr      error
	authorizeErr error
	createErr    error
	commitErr    error
	rollbackErr  error
	createdKey    calls.CallParticipantConnectionKey
	createCtx     context.Context
	createCalls   int
	commitCalls   int
	rollbackCalls int
	lockUserCtx   context.Context
	lockUserID    uuid.UUID
	lockCallCtx   context.Context
	lockCallID    uuid.UUID
	authorizeCtx  context.Context
	authorizedCall *ent.Call
	authorizeUserID uuid.UUID
}

func (tx *durableAllocationTx) LockActiveUserName(ctx context.Context, userID uuid.UUID) (string, error) {
	tx.lockUserCtx = ctx
	tx.lockUserID = userID
	tx.trace.add("lock_user")
	return tx.name, tx.nameErr
}

func (tx *durableAllocationTx) LockActiveCall(ctx context.Context, callID uuid.UUID) (*ent.Call, error) {
	tx.lockCallCtx = ctx
	tx.lockCallID = callID
	tx.trace.add("lock_call")
	return tx.call, tx.callErr
}

func (tx *durableAllocationTx) AuthorizeLockedCall(ctx context.Context, call *ent.Call, userID uuid.UUID) error {
	tx.authorizeCtx = ctx
	tx.authorizedCall = call
	tx.authorizeUserID = userID
	tx.trace.add("authorize")
	return tx.authorizeErr
}

func (tx *durableAllocationTx) CreateUnknownParticipantConnection(ctx context.Context, key calls.CallParticipantConnectionKey) error {
	tx.createCtx = ctx
	tx.trace.add("create")
	tx.createCalls++
	tx.createdKey = key
	return tx.createErr
}

func (tx *durableAllocationTx) Commit() error {
	tx.trace.add("commit")
	tx.commitCalls++
	return tx.commitErr
}

func (tx *durableAllocationTx) Rollback() error {
	tx.trace.add("rollback")
	tx.rollbackCalls++
	return tx.rollbackErr
}

type durableAllocationIssuer struct {
	trace      *durableAllocationTrace
	credential calls.JoinCredential
	err        error
	afterIssue func()
	calls      int
	call       *ent.Call
	identity   calls.ProviderParticipantIdentity
	userName   string
}

func (i *durableAllocationIssuer) IssueDurableCallJoinCredential(call *ent.Call, identity calls.ProviderParticipantIdentity, userName string) (calls.JoinCredential, error) {
	i.trace.add("issue")
	i.calls++
	i.call = call
	i.identity = identity
	i.userName = userName
	if i.afterIssue != nil {
		i.afterIssue()
	}
	return i.credential, i.err
}

type durableTypedNilContext struct{}

type durableAllocationContextKey struct{}

func (*durableTypedNilContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (*durableTypedNilContext) Done() <-chan struct{}       { return nil }
func (*durableTypedNilContext) Err() error                  { return nil }
func (*durableTypedNilContext) Value(any) any               { return nil }
