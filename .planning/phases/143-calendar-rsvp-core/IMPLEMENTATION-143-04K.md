# 143-04K — ScheduledCall due-policy implementation receipt

> **Status:** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT15_COMMIT_REQUIRED /
> RUNTIME_NOT_RUN`. This documentation candidate grants no authority. The immutable source receipt
> below activates only after one atomic commit of precisely the named exact15 documentation
> manifest and its independent immutable acceptance review. Until then it is not a source
> acceptance, no subsequent packet is selected, and executable evidence remains `NOT_RUN`.

## Ordered exact15 acceptance-sync manifest

The one atomic acceptance-sync commit must change **exactly** these canonical paths, in this order;
its diff must equal this manifest with no omitted, substituted or additional path:

1. `.planning/STATE.md`
2. `.planning/AUTONOMOUS-QUESTIONS-2026-08-21.md`
3. `.planning/decisions/specs/MAIL-SPEC.md`
4. `.planning/phases/143-calendar-rsvp-core/.continue-here.md`
5. `.planning/phases/143-calendar-rsvp-core/DISPATCH-INDEX.md`
6. `.planning/phases/143-calendar-rsvp-core/DISPATCH-REMAINING.md`
7. `.planning/phases/143-calendar-rsvp-core/IMPLEMENTATION-143-04K.md`
8. `.planning/phases/143-calendar-rsvp-core/PLAN.md`
9. `.planning/phases/143-calendar-rsvp-core/PREIMPLEMENTATION-READINESS.md`
10. `.planning/phases/143-calendar-rsvp-core/REVIEW-143-04K-CALL-DUE-AUTHORITY.md`
11. `.planning/phases/143-calendar-rsvp-core/RUNTIME-BLOCKER.md`
12. `docs/mail/CONTRACT-CHANGELOG.md`
13. `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`
14. `docs/mail/QA-HANDOFF.md`
15. `docs/mail/events-v1.schema.json`

This is fourteen modified paths plus this new receipt. A dirty tree or a commit with any other
diff remains non-authoritative; the conditional source-acceptance status and all blockers below
are unchanged.

## Proposed immutable source receipt

| Fact | Value |
| --- | --- |
| Source commit | `572f911227b9d50e21aa73b87c530ddcaf0f8934` |
| Parent | `26b377eb288ac8e363cf6d53c744b60f097e9804` |
| Tree | `ac96e2f5c03256b1e00a391600325a524a840673` |
| Subject | `feat(calendar): add call due policy` |
| Authored-first test blob | `437db3af0263cedb10dcf217cd53fbbcac03cf23` — `backend/tests/calendar/event_call_due_service_test.go` |
| Production blob | `376f46c1cc79956fab3479efc38750fd88161b16` — `backend/internal/calendar/event_call_due_service.go` |

The immutable source diff is exactly those two paths, in that test-first order. It implements the
accepted 04K pure due-policy ABI over accepted 04J projection, 04A `EventOutboxDescriptor` and 04L
`EnvelopeV1`: caller-time input only, no I/O/clock/DB/NATS/CAS/reservation, typed content-free
errors, cloned input and opaque descriptor decision. The existing plural `calls.channel.*.scheduled_due.v1`
payload contract is consumed without a schema semantic change.

## Proposed static acceptance facts

Contract/Spec/Oracle, Security/Adversarial, and Lifecycle/Dependency/Boundary precommit and
immutable exact-SHA reviews are `PASS C/I/M 0/0/0`. Author provenance is
`gpt-5.6-terra/xhigh`, `dispatcher-enforced`, no fallback; telemetry is `UNKNOWN`.

Only after the atomic exact15 receipt-sync commit and those recorded immutable facts is this source
`STATIC_ACCEPTED / EXACT_SHA_PASS`. Then the earlier 04K authority-candidate, precommit and
test-first-dispatch clauses are `CONSUMED/HISTORICAL`; this receipt itself selects no next packet.
Source is not accepted by this dirty candidate.

## Boundaries preserved

04F-K alone remains responsible for locking, CAS, tuple reservation, same-transaction outbox append,
rollback and recovery. AQ-02 is backfill-only. Ent generation/migration, adapters, API/UI, raw iTIP,
legacy `calls.scheduled.due` cutover, launcher and runtime are outside this receipt; launcher remains
`NOT_READY` and all executable/runtime evidence is `NOT_RUN`.
