# TASK-235 — Mail contract foundation

> Date: **2026-08-13** · Priority: **P0** · Status: **Contract Complete (static)** · Phase: **142**

## Outcome

Freeze the engine boundary, ERD, HTTP API, NATS event schemas, security model, UI behavior,
traceability, and QA handoff before product code is written.

## Owned paths

- `.planning/decisions/ADR-MAIL-01-mail-engine-calendar-sync.md`
- `.planning/decisions/specs/MAIL-SPEC.md`
- `.planning/phases/142-mail-contract-foundation/**`
- `docs/mail/**`
- `docs/superpowers/specs/2026-08-13-mail-module-design.md`
- `docs/superpowers/plans/2026-08-13-mail-module.md`
- mail-owned additions to `.planning/ROADMAP.md` and `.planning/STATE.md`

## Acceptance

- Stalwart, Maddy, and docker-mailserver are compared from current primary sources.
- Preferred engine has explicit legal, provenance, resource, persistence, and API gates.
- Source of truth is unambiguous; existing Resend alert mail is preserved.
- Calendar sequence, RSVP reset, recurrence, duplicate UID, timezone/all-day, guest, and
  spoof-defense policies are frozen.
- OpenAPI covers mail, calendar, and admin endpoints; event JSON Schema is versioned.
- `FR-MAIL-*`, `FR-CALSYNC-*`, and named critical `TC-*` are mapped for QA.
- Runtime claims remain `NOT_RUN`; no protected test-rebase paths are changed.

## Verification

Static review of links, file references, requirements, and repository diff. Executable
schema validation is deferred until the accepted Docker-only launcher exists.
