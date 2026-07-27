package lifecycle_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/lifecycle"
)

// TestCancelOnListenError_InvokesCancel_A8_008 — Listen fail → cancel, не os.Exit.
func TestCancelOnListenError_InvokesCancel_A8_008(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	lifecycle.CancelOnListenError(cancel, errors.New("bind: address already in use"))

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("A8-008: CancelOnListenError must invoke cancel")
	}
}

// TestCancelOnListenError_NilCancelSafe_A8_008 — nil cancel не паникует.
func TestCancelOnListenError_NilCancelSafe_A8_008(t *testing.T) {
	require.NotPanics(t, func() {
		lifecycle.CancelOnListenError(nil, errors.New("boom"))
	})
}
