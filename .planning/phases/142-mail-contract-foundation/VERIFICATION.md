# Phase 142 verification — Mail Contract Foundation

> **Verdict:** PASS for the contract-only phase goal after independent remediation
> **Date:** 2026-08-13
> **Runtime:** NOT_RUN
> **Contract:** 0.1.1

## Goal-backward evidence

| Required outcome | Evidence | Result |
|---|---|---|
| Engine/authority/scope decision | `ADR-MAIL-01-mail-engine-calendar-sync.md` | PASS |
| ERD/state/requirements | `MAIL-SPEC.md` | PASS |
| Complete v1 HTTP surface | `docs/mail/openapi.yaml` | PASS by static review; executable validation NOT_RUN |
| Versioned event contract | `docs/mail/events-v1.schema.json` | PASS by static review; executable validation NOT_RUN |
| Product/UI interaction design | `docs/superpowers/specs/2026-08-13-mail-module-design.md` | PASS |
| QA traceability handoff | `docs/mail/QA-HANDOFF.md` | PASS |
| Planning registration | TASK-235…241, Phases 142…148, roadmap/state/plan | PASS |
| No protected-zone mutation | staged path review | PASS |
| No whitespace errors | `git diff --cached --check` after staging | PASS |

## Independent review and remediation

The first post-commit Terra contract review returned `BLOCK` with four P1 and two P2
findings. Contract 0.1.1 resolves them without product-code changes:

- legacy `GET /api/calendar/events` retains and now specifies `{events,from,to}`;
- canonical CRUD/RSVP/availability moves to `/api/calendar/v2/**`;
- timed/all-day/IANA constraints and stale mutation recovery are machine-readable while
  retaining mandatory NFR-13 `{error,request_id}`;
- legacy flat Resend settings remain unchanged and mailbox settings use a separate path;
- event cancellation has 409, and admin operations declare RBAC/audit actions;
- UID/recurrence partial uniqueness, persisted HTTP idempotency receipts, ScheduledCall
  mapping and migration-authority uncertainty are explicit.

The final independent contract rereview returned static `PASS` with only a P2 asking that
RSVP `Location` be constrained. The contract now requires a same-origin
`/api/calendar/v2/events/<uuid>` URI-reference. A separate docs review exposed the current
legacy `/admin/mail/broadcast` contradiction and stale Ent counts; current truth, the
deprecated OpenAPI route, Phase 148 gate and 85 entity types / 87-file measurement are now
synchronized. Its repeat review returned static `PASS` with no new P0/P1/P2. Executable
validation remains unavailable.

## Honesty boundaries

- No host tests, validators, generators, builds, linters, or application processes ran.
- The accepted versioned Docker launcher remains `NOT_READY`, so executable OpenAPI/JSON
  Schema validation and all runtime behavior remain `NOT_RUN`.
- Stalwart license, immutable image provenance, resource use, persistence, JMAP/SMTP and
  external DNS/deliverability gates are not part of this contract-only PASS.
- This verdict does not declare the mail feature implemented, stable, or production-ready.

## Residual risks routed forward

- Phase 143: calendar implementation and deterministic RFC evidence.
- Phase 144: mailbox implementation and fake adapter contract.
- Phase 145: legal/provenance/launcher/runtime gates.
- Phase 148: independent cross-discipline and executable verification.
