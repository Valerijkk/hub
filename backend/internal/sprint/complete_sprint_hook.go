//go:build !hubtest

package sprint

import "context"

// completeSprintAfterMoveHook — no-op вне hubtest (release/prod).
func completeSprintAfterMoveHook(_ context.Context) error { return nil }
