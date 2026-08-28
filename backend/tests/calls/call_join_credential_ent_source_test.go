//go:build hubtest

package calls_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"

	"github.com/valerii/hub/backend/internal/calls"
)

var _ calls.CallJoinCredentialProjectionSource = (*calls.EntRepository)(nil)

type typedNilCallJoinCredentialProjectionContext struct {
	context.Context
}

var _ context.Context = (*typedNilCallJoinCredentialProjectionContext)(nil)

func TestEntRepositoryCallJoinCredentialProjectionSourceFailsClosedForInvalidInputs(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	repository := calls.NewEntRepository(nil)
	var nilRepository *calls.EntRepository

	for _, testCase := range []struct {
		name   string
		invoke func() error
	}{
		{
			name: "load active call with nil client",
			invoke: func() error {
				_, err := repository.LoadActiveCall(context.Background(), callID)
				return err
			},
		},
		{
			name: "load active call with nil receiver",
			invoke: func() error {
				_, err := nilRepository.LoadActiveCall(context.Background(), callID)
				return err
			},
		},
		{
			name: "load active call with nil context",
			invoke: func() error {
				_, err := repository.LoadActiveCall(nil, callID)
				return err
			},
		},
		{
			name: "load active call with nil call ID",
			invoke: func() error {
				_, err := repository.LoadActiveCall(context.Background(), uuid.Nil)
				return err
			},
		},
		{
			name: "load server known user name with nil client",
			invoke: func() error {
				_, err := repository.GetServerKnownCallUserName(context.Background(), callID)
				return err
			},
		},
		{
			name: "load server known user name with nil receiver",
			invoke: func() error {
				_, err := nilRepository.GetServerKnownCallUserName(context.Background(), callID)
				return err
			},
		},
		{
			name: "load server known user name with nil context",
			invoke: func() error {
				_, err := repository.GetServerKnownCallUserName(nil, callID)
				return err
			},
		},
		{
			name: "load server known user name with nil user ID",
			invoke: func() error {
				_, err := repository.GetServerKnownCallUserName(context.Background(), uuid.Nil)
				return err
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var err error

			require.NotPanics(t, func() {
				err = testCase.invoke()
			})
			require.ErrorIs(t, err, calls.ErrCallJoinCredentialInvalid)
		})
	}
}

func TestEntRepositoryCallJoinCredentialProjectionSourceRejectsTypedNilContext(t *testing.T) {
	validID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	repository := calls.NewEntRepository(&ent.Client{})
	var typedNilContext *typedNilCallJoinCredentialProjectionContext
	var ctx context.Context = typedNilContext

	for _, testCase := range []struct {
		name   string
		invoke func() error
	}{
		{
			name: "load active call",
			invoke: func() error {
				_, err := repository.LoadActiveCall(ctx, validID)
				return err
			},
		},
		{
			name: "load server known user name",
			invoke: func() error {
				_, err := repository.GetServerKnownCallUserName(ctx, validID)
				return err
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var err error

			require.NotPanics(t, func() {
				err = testCase.invoke()
			})
			require.ErrorIs(t, err, calls.ErrCallJoinCredentialInvalid)
		})
	}
}
