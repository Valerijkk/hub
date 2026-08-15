# Implementation 143-04A — event service contract

> **Status:** `STATIC_ACCEPTED` / `EXACT_SHA_PASS`
> **Source commit:** `aa85f833a806a36f6d6ef2d8e7ae36dca2621574`
> **Parent:** `ac9ed8e62edb9064f985a0ad7abf3a2ec42d7304`
> **Tree:** `1e305267cf47f98000ce89d76f9204d9cd84df93`
> **Subject:** `feat(calendar): add event service contract`

> **Downstream receipt:** 04B is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `75dafea889055126f768227a4ae62b28c91b7393`; see `IMPLEMENTATION-143-04B.md`. Only old 04B precommit/unaccepted/current-next clauses are consumed; 04A remains static-only, launcher `NOT_READY`, executable `NOT_RUN`, and 03E Ent generation/migration block unchanged.

## Exact accepted boundary

| Path | Blob |
|---|---|
| `backend/internal/calendar/event_errors.go` | `660578a29c24fe610d3c97072d922c05e876ae83` |
| `backend/internal/calendar/event_repository.go` | `591608e87092ea80308a05923b22d201d4d124a9` |
| `backend/internal/calendar/event_service.go` | `585fe73ecdc5b8b45d41c6f16bebc4cd7b69d5e0` |
| `backend/tests/calendar/event_service_contract_test.go` | `713cb9599baeaaf16c1b95e833e93ee69a50cf08` |

The test contract was authored first. Terra/xhigh authored the packet; Terra/high static
precommit and immutable exact-SHA review both PASS with C/I/M `0/0/0`.
`provenance_source: dispatcher-enforced`; fallback: none; actual telemetry is `UNKNOWN`
unless exposed.

## Acceptance limits

The receipt is static source acceptance only. RED/GREEN, tests, build, gofmt, lint,
validators, Docker and runtime are `NOT_RUN`; the canonical launcher is `NOT_READY`.
It grants no Ent/schema generation, migration, runtime, product or phase authority.

Commit `6386f3ee` is separate hygiene: it corrected only the stale accepted 143-04L test
import to the current module. Its static review PASS C/I/M `0/0/0`; it has no behavior
change.

Clauses claiming 143-03E blocks pure 143-04A or that no further source can proceed are
`CONSUMED/HISTORICAL`. 143-03E remains `BLOCKED/NOT_RUN` for Ent generation and migration
only. The former 143-04G/04B next routes are also `CONSUMED/HISTORICAL`; both packets are
accepted in their implementation receipts. No next Calendar code packet is selected; root
must separately dispatch any later packet. All unrelated blockers remain unchanged.
