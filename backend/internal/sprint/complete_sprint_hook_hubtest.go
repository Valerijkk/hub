//go:build hubtest

package sprint

import "context"

// CompleteSprintAfterMoveHook — inject mid-fail после move issues (P1-BE-05 Ent proof).
// Тест ставит ошибку → tx Rollback; снятие tx в prod → partial state → RED.
var CompleteSprintAfterMoveHook func(context.Context) error

func completeSprintAfterMoveHook(ctx context.Context) error {
	if CompleteSprintAfterMoveHook != nil {
		return CompleteSprintAfterMoveHook(ctx)
	}
	return nil
}
