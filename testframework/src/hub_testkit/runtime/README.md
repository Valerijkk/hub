# Runtime boundary

## Belongs here

Run IDs, task ownership, bounded clocks and polling, cleanup-aware lifecycle
state, and redaction primitives belong here.

## Forbidden

A global runtime singleton, unbounded waits, hidden ownership, secret-bearing
logs, and domain behavior are forbidden.

## First consumer

The first accepted consumer must introduce one explicit runtime contract
test-first, including its bound and cleanup semantics. No runtime object exists
in this structure-only slice.
