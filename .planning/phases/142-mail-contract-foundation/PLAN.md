# Phase 142 — Mail Contract Foundation

> Task: TASK-235 · Status: CONTRACT_REOPENED_P1_ENGINE_BYTE_BOUNDS_AND_SCHEMA_MATRIX · Runtime: NOT_RUN

## Goal

Freeze implementation-ready product, architecture, domain, HTTP, event, security, UI and
QA contracts without touching product runtime or protected test paths.

## Tasks

### 142-01 Engine and repository research

- Inspect current Hub mail/calendar/storage/search/notification/ws/audit seams.
- Compare Stalwart, Maddy and docker-mailserver from current primary sources.
- Record version volatility and separate facts from inferences/resource unknowns.

### 142-02 Architecture and product design

- Create `.planning/decisions/ADR-MAIL-01-mail-engine-calendar-sync.md`.
- Create `docs/superpowers/specs/2026-08-13-mail-module-design.md`.
- Freeze engine authority, internal delivery, calendar sequence/RSVP and security policy.

### 142-03 Machine-readable contracts

- Create `.planning/decisions/specs/MAIL-SPEC.md` with ERD/state/traceability.
- Create `docs/mail/openapi.yaml` with every v1 endpoint.
- Create `docs/mail/events-v1.schema.json` with versioned NATS envelopes.
- Create `docs/mail/QA-HANDOFF.md` and `docs/mail/FOUNDER-QUESTIONS.md`.

### 142-04 Planning registration

- Create TASK-235…241 and phases 142…148 without modifying protected 128…141 artifacts.
- Add the mail workstream to `.planning/ROADMAP.md` and a separate state section.
- Create the implementation plan and phase plans.

### 142-05 Static verification and commit

- Review `git diff --check`, changed paths, unresolved placeholders, accidental secrets,
  protected-path changes, and internal links by inspection.
- Do not run host validators. If accepted Docker validation is unavailable, record it.
- Commit only contract/planning artifacts with a Conventional Commit and Russian body.

## Completion gate

All listed artifacts exist and agree; runtime remains honestly `NOT_RUN`. The phase can be
marked contract-complete without implying product implementation or engine approval.
