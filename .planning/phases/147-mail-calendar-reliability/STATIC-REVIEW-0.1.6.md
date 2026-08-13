# Phase 147 — Mail 0.1.6 current-provenance static review

> **Reviewed semantic contract SHA:** `5a351d5bd51bd215775f3828bc26979cf985a4ac`
> **Mail 0.1.5 base:** `eae6f5e50d2564e22799e9a795a62603b97a8ed5`
> **Final remediation parent:** `644b2ea894b7645f28ae8fbac647a761b1757276`
> **Branch:** `codex/mail-module-20260813`
> **Root verdict:** `ROOT_STATIC_PASS`
> **Independent verdict:** `TERRA_HIGH_STATIC_PASS`
> **G148-02 state:** `PASS` for the exact reviewed SHA
> **Product/test code:** `NOT_IMPLEMENTED`
> **Runtime/build/test/schema-validator evidence:** `NOT_RUN`

This evidence freezes only static contract coherence. It does not claim source
implementation, executable validation, merge-readiness, engine acceptance or production
readiness. Any semantic contract change makes this evidence `STALE` and reopens G148-02.

## Scope and method

Root reviewed the complete Mail 0.1.6 delta and the current snapshot across ADR, Mail spec,
OpenAPI, unchanged event schema, QA handoff, Phase 143–148 contracts and routed docs. The
review covered mailbox control-plane/recovery, fixed-script rule deployment, safe content,
remote-image SSRF handling, mutation receipts, typed bulk/classifier behavior, exclusive
folder writer gating, ScheduledCall authority/cutover and every calendar conflict path.

Read-only Git/static evidence on the exact SHA:

- the isolated worktree was clean;
- `git diff --check eae6f5e5..5a351d5b`: PASS;
- the complete 0.1.6 delta contains 49 changed paths and zero protected-zone matches;
- added-line secret and forbidden AI/coauthor trailer scan: zero matches;
- OpenAPI blob: `8081f18952343c4be5a8e4f3e2ace40e82bd2789`;
- event schema blob stayed
  `dd98b5587ef38dddf20ba61d895464e001fb3e10` from 0.1.5 through the reviewed SHA;
- generic `Conflict` response references/definition and obsolete
  `AttendeeVersionConflict`: zero matches;
- static component-reference and duplicate-operationId inspection: PASS.

Project policy forbids host tests, builds, linters and validators. Docker Desktop 29.7.2 is
available, but the accepted versioned `/qa-gate` launcher remains `NOT_READY`; therefore no
direct Docker/Compose fallback was used and executable evidence remains `NOT_RUN`.

## Review finding closure

The first independent pass on `644b2ea8` identified two findings. The remediation commit
`5a351d5b` closed both and synchronized all downstream contracts:

| ID | Severity | Finding | Resolution on reviewed SHA |
|---|---|---|---|
| R147-016-01 | P1 | Five calendar/legacy writer 409s still used a generic conflict response | legacy PUT/DELETE now use `SCHEDULED_CALL_NOT_MUTABLE`; event PATCH/DELETE use `CALENDAR_EVENT_VERSION_CONFLICT|CALENDAR_EVENT_NOT_MUTABLE`; occurrence DELETE uses `CALENDAR_EVENT_VERSION_CONFLICT|CALENDAR_OCCURRENCE_NOT_MUTABLE`; NFR-13 and zero-mutation semantics are explicit, and stale v2 paths carry the current quoted event ETag plus same-origin canonical Location |
| R147-016-02 | P2 | The founder-authorized Terra route was not visible from mandatory first-read policy | `AGENTS.md` now records a branch/phase-narrow `gpt-5.6-terra` / `xhigh` exception with no fallback while retaining every other routing, scope and verification rule |

RSVP remains intentionally separate: `CalendarMutationConflict` exposes
`ATTENDEE_VERSION_CONFLICT|IDEMPOTENCY_KEY_REUSED`, with attendee-version ETag and canonical
event Location only for the stale-attendee branch. No endpoint-specific error body was
introduced; every path keeps `{error,request_id}`.

## Independent review

An independent read-only reviewer inspected the exact clean SHA as
`gpt-5.6-terra` / `high` and returned `PASS` with no P0/P1/P2. It explicitly rechecked the
five writer routes, conditional recovery headers, RSVP separation, ADR/spec/OpenAPI/QA/
dispatch/gate agreement, task-local routing scope, generic-conflict absence, event-schema
immutability and honest status claims. Runtime, tests, validators and Docker jobs were not
run by that reviewer.

## Remaining gates

G148-02 is closed only for this semantic SHA. Atomic implementation provenance, source
review, migration/generation authority, selected-engine evidence, accepted launcher,
executable QA, broadcast disposition, final protected-rebase signal and push/PR authority
remain pending, blocked, unresolved, unauthorized or `NOT_RUN` in `GATE-MATRIX.md`.

The only supported claim is:
`Mail 0.1.6 static contract coherence PASS on 5a351d5b`. The Hub MVP and Mail module remain
unfinished, unimplemented, runtime-unverified and not merge/production-ready.
