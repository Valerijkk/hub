//go:build !release

package qamailcalendar

// ThreadSearchCommitOperation identifies the lifecycle mutation whose commit
// result is being reconciled.
type ThreadSearchCommitOperation uint8

const (
	ThreadSearchCommitOperationInvalid ThreadSearchCommitOperation = iota
	ThreadSearchCommitOperationCreate
	ThreadSearchCommitOperationDelete
)

// ThreadSearchCommitPhase identifies whether the mutation failed before its
// commit boundary, completed, or has an unknown commit outcome.
type ThreadSearchCommitPhase uint8

const (
	ThreadSearchCommitPhaseInvalid ThreadSearchCommitPhase = iota
	ThreadSearchCommitPhaseSucceeded
	ThreadSearchCommitPhasePreCommit
	ThreadSearchCommitPhaseOutcomeUnknown
)

// ThreadSearchCommitObservation is the closed classification produced by a
// fresh fixture-manifest probe.
type ThreadSearchCommitObservation uint8

const (
	ThreadSearchCommitObservationInvalid ThreadSearchCommitObservation = iota
	ThreadSearchCommitObservationNotProbed
	ThreadSearchCommitObservationExactPresent
	ThreadSearchCommitObservationExactAbsent
	ThreadSearchCommitObservationPartial
	ThreadSearchCommitObservationProbeError
)

// ThreadSearchCommitAction is the lifecycle-state action selected by the
// reconciliation table.
type ThreadSearchCommitAction uint8

const (
	ThreadSearchCommitActionInvalid ThreadSearchCommitAction = iota
	ThreadSearchCommitActionCompleteCreated
	ThreadSearchCommitActionRemoveCreating
	ThreadSearchCommitActionRetainCreating
	ThreadSearchCommitActionRemoveDeleted
	ThreadSearchCommitActionRestoreCreated
	ThreadSearchCommitActionRetainDeleting
)

// ThreadSearchCommitSemantic is the controller-visible semantic paired with a
// reconciliation action.
type ThreadSearchCommitSemantic uint8

const (
	ThreadSearchCommitSemanticInvalid ThreadSearchCommitSemantic = iota
	ThreadSearchCommitSemanticCreated
	ThreadSearchCommitSemanticInternalError
	ThreadSearchCommitSemanticNoContent
	ThreadSearchCommitSemanticConflict
)

// ThreadSearchCommitReconcileDecision is the complete result of the closed
// commit-outcome reconciliation table. Its zero value intentionally means
// invalid input and must not be interpreted as an actionable decision.
type ThreadSearchCommitReconcileDecision struct {
	Action   ThreadSearchCommitAction
	Semantic ThreadSearchCommitSemantic
}

// DecideThreadSearchCommitReconciliation selects the only permitted lifecycle
// action for a commit result and fresh-manifest observation. Every unlisted
// combination fails closed with the zero decision.
func DecideThreadSearchCommitReconciliation(
	operation ThreadSearchCommitOperation,
	phase ThreadSearchCommitPhase,
	observation ThreadSearchCommitObservation,
) ThreadSearchCommitReconcileDecision {
	switch operation {
	case ThreadSearchCommitOperationCreate:
		switch phase {
		case ThreadSearchCommitPhaseSucceeded:
			if observation == ThreadSearchCommitObservationNotProbed {
				return ThreadSearchCommitReconcileDecision{
					Action:   ThreadSearchCommitActionCompleteCreated,
					Semantic: ThreadSearchCommitSemanticCreated,
				}
			}
		case ThreadSearchCommitPhasePreCommit:
			if observation == ThreadSearchCommitObservationNotProbed {
				return ThreadSearchCommitReconcileDecision{
					Action:   ThreadSearchCommitActionRemoveCreating,
					Semantic: ThreadSearchCommitSemanticInternalError,
				}
			}
		case ThreadSearchCommitPhaseOutcomeUnknown:
			switch observation {
			case ThreadSearchCommitObservationExactPresent:
				return ThreadSearchCommitReconcileDecision{
					Action:   ThreadSearchCommitActionCompleteCreated,
					Semantic: ThreadSearchCommitSemanticCreated,
				}
			case ThreadSearchCommitObservationExactAbsent:
				return ThreadSearchCommitReconcileDecision{
					Action:   ThreadSearchCommitActionRemoveCreating,
					Semantic: ThreadSearchCommitSemanticInternalError,
				}
			case ThreadSearchCommitObservationPartial, ThreadSearchCommitObservationProbeError:
				return ThreadSearchCommitReconcileDecision{
					Action:   ThreadSearchCommitActionRetainCreating,
					Semantic: ThreadSearchCommitSemanticInternalError,
				}
			}
		}
	case ThreadSearchCommitOperationDelete:
		switch phase {
		case ThreadSearchCommitPhaseSucceeded:
			if observation == ThreadSearchCommitObservationNotProbed {
				return ThreadSearchCommitReconcileDecision{
					Action:   ThreadSearchCommitActionRemoveDeleted,
					Semantic: ThreadSearchCommitSemanticNoContent,
				}
			}
		case ThreadSearchCommitPhasePreCommit:
			if observation == ThreadSearchCommitObservationNotProbed {
				return ThreadSearchCommitReconcileDecision{
					Action:   ThreadSearchCommitActionRestoreCreated,
					Semantic: ThreadSearchCommitSemanticConflict,
				}
			}
		case ThreadSearchCommitPhaseOutcomeUnknown:
			switch observation {
			case ThreadSearchCommitObservationExactPresent:
				return ThreadSearchCommitReconcileDecision{
					Action:   ThreadSearchCommitActionRestoreCreated,
					Semantic: ThreadSearchCommitSemanticConflict,
				}
			case ThreadSearchCommitObservationExactAbsent:
				return ThreadSearchCommitReconcileDecision{
					Action:   ThreadSearchCommitActionRemoveDeleted,
					Semantic: ThreadSearchCommitSemanticNoContent,
				}
			case ThreadSearchCommitObservationPartial, ThreadSearchCommitObservationProbeError:
				return ThreadSearchCommitReconcileDecision{
					Action:   ThreadSearchCommitActionRetainDeleting,
					Semantic: ThreadSearchCommitSemanticConflict,
				}
			}
		}
	}

	return ThreadSearchCommitReconcileDecision{}
}
