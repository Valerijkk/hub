# Calendar 143-04H occurrence-cancel source receipt

## Conditional receipt status

**Current receipt status:** `STATIC_ACCEPTED / EXACT_SHA_PASS` for only immutable source `73940810046bb3186f2746705a83565ab29368c5`. Exact16 receipt `f9f5d436c08737eeb5a3e965137f329f3ae7f627` is ancestral to HEAD; Contract/Security/Lifecycle reviews are PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback, telemetry `UNKNOWN`. The conditional status block below is `CONSUMED/HISTORICAL` only for 04H. No next packet is selected; generation, migration, adapter, handler, API/UI, launcher and runtime remain blocked or `NOT_RUN`.

`SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT16_COMMIT_REQUIRED / RUNTIME_NOT_RUN`.
These dirty documentation bytes grant no source acceptance. Only one atomic commit whose diff equals
the ordered manifest below, followed by immutable Contract, Security and Lifecycle PASS for that
exact commit, may mark the source `STATIC_ACCEPTED / EXACT_SHA_PASS`.

The source receipt is immutable commit `73940810046bb3186f2746705a83565ab29368c5`:

- parent `9a638f75745aa0fb516a1177a33cbd1c657432fe`
- tree `9f994a8805b43a84c00ed50b8e5544739425edaa`
- subject `feat(calendar): add occurrence cancellation service`

Its ordered source blobs are:

1. authored-first `backend/tests/calendar/event_service_occurrence_cancel_test.go` — `80e19d3b918ae9ddbe3b3ea8803c6062c2f8a713`
2. `backend/internal/calendar/event_repository.go` — `e0da66bf95d4b8193eb5d4034053a353bd154b66`
3. `backend/internal/calendar/event_service_occurrence_cancel.go` — `2b500e5cdda34d7d80d0ccc0ba9c94427167716f`

Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high precommit plus immutable
exact-SHA reviews PASS C/I/M `0/0/0`. Author provenance is Terra/xhigh,
dispatcher-enforced/no fallback; telemetry `UNKNOWN`.

## Exact16 acceptance-sync manifest

The atomic receipt-sync commit must modify exactly, and in this ordered manifest:

1. `.planning/STATE.md`
2. `.planning/decisions/specs/MAIL-SPEC.md`
3. `.planning/phases/143-calendar-rsvp-core/.continue-here.md`
4. `.planning/phases/143-calendar-rsvp-core/DISPATCH-143-04H.md`
5. `.planning/phases/143-calendar-rsvp-core/DISPATCH-INDEX.md`
6. `.planning/phases/143-calendar-rsvp-core/DISPATCH-REMAINING.md`
7. `.planning/phases/143-calendar-rsvp-core/IMPLEMENTATION-143-04H.md`
8. `.planning/phases/143-calendar-rsvp-core/PLAN.md`
9. `.planning/phases/143-calendar-rsvp-core/PREIMPLEMENTATION-READINESS.md`
10. `.planning/phases/143-calendar-rsvp-core/REVIEW-143-04H-OCCURRENCE-CANCEL-AUTHORITY.md`
11. `.planning/phases/143-calendar-rsvp-core/RUNTIME-BLOCKER.md`
12. `docs/mail/CONTRACT-CHANGELOG.md`
13. `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`
14. `docs/mail/QA-HANDOFF.md`
15. `docs/mail/events-v1.schema.json`
16. `docs/mail/openapi.yaml`

No other path is in scope. The schema and OpenAPI changes are status metadata only and do not alter
event, payload or API semantics.

## Post-activation boundary

After the atomic exact16 commit and immutable triple PASS, this exact source alone is
`STATIC_ACCEPTED / EXACT_SHA_PASS`; former 04H authority, precommit, source-unaccepted and
current-next clauses are `CONSUMED/HISTORICAL`. No next packet is selected automatically.

The receipt does not authorize 04J or 04K coupling, 04F-H persistence work, Ent/concrete adapters,
handlers/wiring, generation, migration, API/UI, raw iTIP/transport or runtime. Launcher remains
`NOT_READY`; executables and runtime remain `NOT_RUN`.
