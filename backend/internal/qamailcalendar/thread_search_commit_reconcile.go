//go:build !release

package qamailcalendar

const threadSearchFixtureManifestProbeRowCount = 24

// ClassifyThreadSearchFixtureFreshProbe reduces a fresh immutable-manifest
// probe to the closed observation consumed by the commit policy. A probe that
// was not attempted remains distinct from a probe that was attempted but did
// not prove one complete, exact manifest state.
func ClassifyThreadSearchFixtureFreshProbe(
	summary ThreadSearchFixtureFreshProbeSummary,
) ThreadSearchCommitObservation {
	if !summary.Probed {
		return ThreadSearchCommitObservationNotProbed
	}
	if summary.Failed {
		return ThreadSearchCommitObservationProbeError
	}
	if summary.ExactRows < 0 || summary.AbsentRows < 0 || summary.MismatchedRows < 0 {
		return ThreadSearchCommitObservationPartial
	}
	if summary.ExactRows == threadSearchFixtureManifestProbeRowCount &&
		summary.AbsentRows == 0 &&
		summary.MismatchedRows == 0 {
		return ThreadSearchCommitObservationExactPresent
	}
	if summary.AbsentRows == threadSearchFixtureManifestProbeRowCount &&
		summary.ExactRows == 0 &&
		summary.MismatchedRows == 0 {
		return ThreadSearchCommitObservationExactAbsent
	}

	return ThreadSearchCommitObservationPartial
}

// ReconcileThreadSearchFixtureCommit applies the closed commit policy through
// a full-record map CAS. The caller owns synchronization for the fixture map;
// this seam deliberately performs neither I/O nor synchronization itself.
func ReconcileThreadSearchFixtureCommit(
	fixtures map[string]ThreadSearchFixtureLifecycleRecord,
	input ThreadSearchFixtureCommitReconcileInput,
) ThreadSearchFixtureCommitReconcileResult {
	decision := DecideThreadSearchCommitReconciliation(
		input.Operation,
		input.Phase,
		ClassifyThreadSearchFixtureFreshProbe(input.Probe),
	)
	if decision.Action == ThreadSearchCommitActionInvalid ||
		decision.Semantic == ThreadSearchCommitSemanticInvalid ||
		!threadSearchFixtureExpectedRecordAdmitsAction(input.Expected, decision.Action) {
		return threadSearchFixtureCommitCASConflict(input.Operation)
	}

	switch decision.Action {
	case ThreadSearchCommitActionCompleteCreated:
		if !threadSearchFixtureReplaceExactState(
			fixtures,
			input.Expected,
			ThreadSearchFixtureLifecycleStateCreated,
		) {
			return threadSearchFixtureCommitCASConflict(input.Operation)
		}
		return ThreadSearchFixtureCommitReconcileResult{
			Status:      201,
			Disposition: ThreadSearchFixtureLifecycleDispositionCreated,
			Applied:     true,
		}

	case ThreadSearchCommitActionRemoveCreating:
		if !threadSearchFixtureRemoveExact(fixtures, input.Expected) {
			return threadSearchFixtureCommitCASConflict(input.Operation)
		}
		return ThreadSearchFixtureCommitReconcileResult{
			Status:      500,
			Error:       "internal server error",
			Disposition: ThreadSearchFixtureLifecycleDispositionRemoved,
			Applied:     true,
		}

	case ThreadSearchCommitActionRetainCreating:
		if !threadSearchFixtureContainsExact(fixtures, input.Expected) {
			return threadSearchFixtureCommitCASConflict(input.Operation)
		}
		return ThreadSearchFixtureCommitReconcileResult{
			Status:      500,
			Error:       "internal server error",
			Disposition: ThreadSearchFixtureLifecycleDispositionRetained,
		}

	case ThreadSearchCommitActionRemoveDeleted:
		return reconcileThreadSearchFixtureConfirmedDelete(fixtures, input.Expected)

	case ThreadSearchCommitActionRestoreCreated:
		if !threadSearchFixtureReplaceExactState(
			fixtures,
			input.Expected,
			ThreadSearchFixtureLifecycleStateCreated,
		) {
			return threadSearchFixtureCommitCASConflict(input.Operation)
		}
		return ThreadSearchFixtureCommitReconcileResult{
			Status:      409,
			Error:       "fixture lifecycle conflict",
			Disposition: ThreadSearchFixtureLifecycleDispositionRestored,
			Applied:     true,
		}

	case ThreadSearchCommitActionRetainDeleting:
		if !threadSearchFixtureContainsExact(fixtures, input.Expected) {
			return threadSearchFixtureCommitCASConflict(input.Operation)
		}
		return ThreadSearchFixtureCommitReconcileResult{
			Status:      409,
			Error:       "fixture lifecycle conflict",
			Disposition: ThreadSearchFixtureLifecycleDispositionRetained,
		}
	}

	return threadSearchFixtureCommitCASConflict(input.Operation)
}

func reconcileThreadSearchFixtureConfirmedDelete(
	fixtures map[string]ThreadSearchFixtureLifecycleRecord,
	expected ThreadSearchFixtureLifecycleRecord,
) ThreadSearchFixtureCommitReconcileResult {
	observed, found := fixtures[expected.Nonce]
	if !found {
		return ThreadSearchFixtureCommitReconcileResult{
			Status:      204,
			Disposition: ThreadSearchFixtureLifecycleDispositionRemoved,
		}
	}
	if observed == expected {
		delete(fixtures, expected.Nonce)
		return ThreadSearchFixtureCommitReconcileResult{
			Status:      204,
			Disposition: ThreadSearchFixtureLifecycleDispositionRemoved,
			Applied:     true,
		}
	}
	if observed.Nonce != expected.Nonce || observed.Scenario != expected.Scenario {
		return threadSearchFixtureCommitCASConflict(ThreadSearchCommitOperationDelete)
	}
	if observed.State == ThreadSearchFixtureLifecycleStateInvalidated {
		return threadSearchFixtureCommitCASConflict(ThreadSearchCommitOperationDelete)
	}

	invalidated := observed
	invalidated.State = ThreadSearchFixtureLifecycleStateInvalidated
	if current, stillFound := fixtures[expected.Nonce]; !stillFound || current != observed {
		return threadSearchFixtureCommitCASConflict(ThreadSearchCommitOperationDelete)
	}
	fixtures[expected.Nonce] = invalidated

	return ThreadSearchFixtureCommitReconcileResult{
		Status:      409,
		Error:       "fixture lifecycle conflict",
		Disposition: ThreadSearchFixtureLifecycleDispositionInvalidated,
		Applied:     true,
	}
}

func threadSearchFixtureExpectedRecordAdmitsAction(
	record ThreadSearchFixtureLifecycleRecord,
	action ThreadSearchCommitAction,
) bool {
	if record.Nonce == "" ||
		record.Scenario != ThreadSearchFixtureLifecycleScenarioThreadSearchV1 {
		return false
	}

	switch action {
	case ThreadSearchCommitActionCompleteCreated,
		ThreadSearchCommitActionRemoveCreating,
		ThreadSearchCommitActionRetainCreating:
		return record.State == ThreadSearchFixtureLifecycleStateCreating
	case ThreadSearchCommitActionRemoveDeleted,
		ThreadSearchCommitActionRestoreCreated,
		ThreadSearchCommitActionRetainDeleting:
		return record.State == ThreadSearchFixtureLifecycleStateDeleting
	default:
		return false
	}
}

func threadSearchFixtureContainsExact(
	fixtures map[string]ThreadSearchFixtureLifecycleRecord,
	expected ThreadSearchFixtureLifecycleRecord,
) bool {
	current, found := fixtures[expected.Nonce]
	return found && current == expected
}

func threadSearchFixtureReplaceExactState(
	fixtures map[string]ThreadSearchFixtureLifecycleRecord,
	expected ThreadSearchFixtureLifecycleRecord,
	state ThreadSearchFixtureLifecycleState,
) bool {
	if !threadSearchFixtureContainsExact(fixtures, expected) {
		return false
	}

	updated := expected
	updated.State = state
	fixtures[expected.Nonce] = updated
	return true
}

func threadSearchFixtureRemoveExact(
	fixtures map[string]ThreadSearchFixtureLifecycleRecord,
	expected ThreadSearchFixtureLifecycleRecord,
) bool {
	if !threadSearchFixtureContainsExact(fixtures, expected) {
		return false
	}

	delete(fixtures, expected.Nonce)
	return true
}

func threadSearchFixtureCommitCASConflict(
	operation ThreadSearchCommitOperation,
) ThreadSearchFixtureCommitReconcileResult {
	if operation == ThreadSearchCommitOperationDelete {
		return ThreadSearchFixtureCommitReconcileResult{
			Status:      409,
			Error:       "fixture lifecycle conflict",
			Disposition: ThreadSearchFixtureLifecycleDispositionCASConflict,
		}
	}

	return ThreadSearchFixtureCommitReconcileResult{
		Status:      500,
		Error:       "internal server error",
		Disposition: ThreadSearchFixtureLifecycleDispositionCASConflict,
	}
}
