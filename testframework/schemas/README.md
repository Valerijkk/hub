# Machine-readable schema boundary

## Belongs here

Closed, versioned machine-readable contracts for cases, runs, ownership, and
report metadata belong here.

## Forbidden

Undocumented fields, open-ended property bags, executable behavior, generated
reports, and secret-bearing examples are forbidden.

## First consumer

The first accepted consumer must introduce one minimal closed schema together
with framework selftests before any producer relies on it.

## hard2g backend-unit receipt schemas

The hard2g candidate has exactly three closed schemas:

- `backend-unit-package-plan.schema.v1.json` owns a planned package stage;
- `backend-unit-container-receipt.schema.v1.json` owns container-stage evidence;
- `backend-unit-observer-receipt.schema.v1.json` owns the final receipt
  (`finalReceipt`) published by the observer.

`finalReceipt` is terminal evidence, not a mutable progress record. It may be
`COMPLETE` after the normal container choreography or `ABORTED` only after the
safe bootstrap/execution TERM choreography proves the required terminal facts.
Cold, planning, and uncertain interruption paths publish no receipt and exit
`75`. This contract is **STATIC_ACCEPTED** only; executable evidence is
**EXECUTABLE_NOT_RUN**.
