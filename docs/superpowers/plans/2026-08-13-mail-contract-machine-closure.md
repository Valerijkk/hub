# Mail 0.1.8 Machine-contract Closure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remediate the exact current Phase 142 machine-contract findings while keeping every implementation and runtime gate honest.

**Architecture:** OpenAPI owns the closed HTTP DTOs, examples and policy-id vocabulary; the JSON Schema owns event-envelope shape and cross-field metadata; engine/persistence documents own the reachable recipient ceiling. QA, specification, verification, changelog and state mirror those same values and gates.

**Tech Stack:** OpenAPI 3.1 YAML, JSON Schema draft 2020-12, Markdown planning artifacts, Git.

## Global Constraints

- Work only on `codex/mail-module-20260813` in its linked worktree.
- Do not read or change `testframework/**`, `backend/testframework/**`, `pro-qa-ai/**`, `.planning/audits/test-rebase/**`, Phases 128-141, `.env.local` or `.git/index.lock`.
- Use `apply_patch` for edits.
- Do not run host tests, validators, generators, linters, builds or application processes.
- Do not use direct Docker/Compose while the accepted launcher is `NOT_READY`.
- Preserve `NOT_RUN` and actual-model telemetry `UNKNOWN`; neither may be upgraded by static review.
- Commit only explicit remediation paths and include no AI/co-author attribution.

---

### Task 1: Close the public HTTP schemas and policy vocabulary

**Files:**
- Modify: `docs/mail/openapi.yaml`

**Interfaces:**
- Consumes: existing bearer security, `x-hub-rbac`, NFR-13 `ErrorEnvelope`, mailbox/calendar/CRM prose boundaries.
- Produces: `x-hub-authz-vocabulary` v1 and one `x-hub-authz` policy id on every non-admin operation.

- [ ] **Step 1:** Add the default-deny policy vocabulary with mailbox-owner, scheduled-call, calendar organizer/reader/attendee/availability and CRM deal+message owner policies.
- [ ] **Step 2:** Apply exactly one declared policy id to every non-admin operation and retain admin `403`/audit metadata.
- [ ] **Step 3:** Split thread/message field bases from closed public summary/detail schemas; close every remaining public object and nested attendee/health object.
- [ ] **Step 4:** Bound the legacy broadcast request to 8,388,608 UTF-8 bytes, bound counters to signed int32, close both objects and retain its blocked governance status.
- [ ] **Step 5:** Change `AdminMailboxSettings.max_recipients` and its update request maximum from 500 to 300.
- [ ] **Step 6:** Add standard examples for bounded cursor continuation, event-create request plus byte-identical replay response, recoverable calendar `409`, and mailbox `423`.

### Task 2: Align the event schema and cross-field oracle

**Files:**
- Modify: `docs/mail/events-v1.schema.json`
- Modify: `docs/mail/QA-HANDOFF.md`

**Interfaces:**
- Consumes: canonical event attendee maximum 200 and existing channel subject grammar.
- Produces: an event schema whose calendar payloads cannot exceed the HTTP/domain attendee cap and whose channel UUID equality is machine-addressable.

- [ ] **Step 1:** Change both mail received/sent `to.maxItems` and both calendar invite `attendee_ids.maxItems` values from 500 to 200.
- [ ] **Step 2:** Add one structured `x-hub-cross-field-invariants` entry to `callsScheduledDue` using source pointer `/subject`, capture group 1 and target pointer `/payload/channel_id`.
- [ ] **Step 3:** Add QA oracles requiring producer and consumer rejection before outbox/domain mutation when the captured subject UUID differs from the payload UUID.

### Task 3: Make the recipient policy reachable

**Files:**
- Modify: `docs/mail/ENGINE-CONTRACT-0.1.8.md`
- Modify: `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`
- Modify: `.planning/decisions/specs/MAIL-SPEC.md`
- Modify: `docs/mail/QA-HANDOFF.md`

**Interfaces:**
- Consumes: 100-item caps for each draft `to`, `cc` and `bcc` collection.
- Produces: a named 300-item combined hard ceiling and policy/check maxima that cannot exceed it.

- [ ] **Step 1:** Add `MaxEngineDraftTotalRecipientItems = 300` and define the combined-count rule.
- [ ] **Step 2:** Change persistence prose and `mailbox_policies_bounds` from 500 to 300.
- [ ] **Step 3:** Synchronize the canonical specification and QA boundary oracle to reject the 301st combined recipient before engine I/O.

### Task 4: Record review disposition and honest gates

**Files:**
- Create: `.planning/phases/142-mail-contract-foundation/REMEDIATION-142-MACHINE-CONTRACT-2026-08-13.md`
- Modify: `.planning/phases/142-mail-contract-foundation/VERIFICATION.md`
- Modify: `docs/mail/CONTRACT-CHANGELOG.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: Terra/high review at source SHA `c48046b876f429c7ab31f58fbc3ee0b960529144` with actual telemetry `UNKNOWN`.
- Produces: a traceable remediation record that remains pending independent provenance and executable validation.

- [ ] **Step 1:** Record every finding, technical disposition and changed contract surface.
- [ ] **Step 2:** Mark prior blanket PASS language superseded and keep the Phase 142 verdict pending independent rereview.
- [ ] **Step 3:** Add the amendment to the 0.1.8 changelog and current state without changing implementation/runtime status.

### Task 5: Static verification and atomic commit

**Files:**
- Verify only the explicit paths from Tasks 1-4.

**Interfaces:**
- Consumes: the complete remediation diff.
- Produces: static evidence and one atomic documentation commit.

- [ ] **Step 1:** Prove every non-admin operation has one declared `x-hub-authz` policy and no admin operation lost `x-hub-rbac`.
- [ ] **Step 2:** Prove no open public object remains except intentionally open internal field-base schemas used under closed wrappers.
- [ ] **Step 3:** Prove attendee and recipient limits are synchronized and no stale `max_recipients` maximum 500 remains in the mail contract set.
- [ ] **Step 4:** Run `git diff --check`, review protected-path absence and inspect the complete diff.
- [ ] **Step 5:** Stage only explicit paths, run `git diff --cached --check`, and commit as `docs(mail): close machine contract gaps`.
- [ ] **Step 6:** Keep executable OpenAPI/JSON Schema/runtime evidence `NOT_RUN`; do not claim formal Terra acceptance while actual telemetry is `UNKNOWN`.
