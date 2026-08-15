# Implementation 143-04G — calendar EVENTS-stream admission

> **Status:** `STATIC_ACCEPTED` / `EXACT_SHA_PASS`
> **Source commit:** `48a7273d100c07788def88e2f1b8d69c7173580f`
> **Parent:** `6386f3eeb4025fc0aca55f49aaa8083ff856f8a9`
> **Tree:** `65c5074a9793d24bac5f9b18aa214fbc942ac922`
> **Subject:** `feat(calendar): admit calendar event subjects`

> **Downstream receipt:** 04B is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `75dafea889055126f768227a4ae62b28c91b7393`; see `IMPLEMENTATION-143-04B.md`. Only old 04B precommit/unaccepted/current-next clauses are consumed; 04G remains static-only, launcher `NOT_READY`, executable `NOT_RUN`, and 03E Ent generation/migration block unchanged.

## Exact accepted boundary

| Path | Blob |
|---|---|
| `backend/internal/platform/nats/client.go` | `e838704e6b4f5ebb9d1bb52d7d36b776c4453738` |
| `backend/tests/calendar/calendar_subjects_test.go` | `71f5f68280eeb6b64e590b809c6df6b618695c9b` |

The external-package test was authored first. Terra/xhigh authored the packet; Terra/high static
precommit and immutable exact-SHA review both PASS with C/I/M `0/0/0`.
`provenance_source: dispatcher-enforced`; fallback: none; actual telemetry is `UNKNOWN`
unless exposed.

## Acceptance limits

The receipt is static source acceptance only. RED/GREEN, tests, build, gofmt, lint,
validators, Docker and runtime are `NOT_RUN`; the canonical launcher is `NOT_READY`.
It grants no Ent/schema generation, migration, runtime, product or phase authority.

Former 143-04G/04B next-route clauses are `CONSUMED/HISTORICAL`. 143-03E remains
`BLOCKED/NOT_RUN` for Ent generation and migration only. 143-04B is accepted in
`IMPLEMENTATION-143-04B.md`. No next Calendar code packet is selected; root must separately
dispatch any later packet. All unrelated blockers remain unchanged.
