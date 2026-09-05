# Mail Occurrence RSVP Implementation Plan
> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** deliver durable, tenant-safe Mail RSVP targeting for calendar series and individual recurrence instances.

**Scope lock:** do not parse Mail content in the client; do not backfill existing receipts; do not change outbound invitation behavior; preserve unrelated dirty worktree changes.

## 2026-09-05 implementation status

Tasks 1--3 source changes are committed through code commit `76a5afb3`. The RSVP target now carries required `attendee_version` and `rsvp_status`; occurrence state uses its effective active override, then the active series attendee fallback. The client validates the complete target fail-closed, passes canonical recurrence IDs to Calendar's existing raw-base64url route encoder, uses canonical Mail query-key factories, and keeps returned occurrence attendee state component-local. Independent backend review passed after adding the target-event/detail-event equality invariant; independent frontend review passed after adding the undefined-target no-button case. Schema-maintenance evidence is recorded in `backend/migrations/README.md`.

At code commit `76a5afb3`, clean Docker `backend-unit` final receipts are recorded under `.tmp/acceptance-5937e3ef/testframework/reports/backend-unit/`: build PASS at `2026-09-05T07:41:39Z`; broad vet/unit/race FAIL only on the known unrelated baseline of 1 vet plus 35 legacy failures. The exact parent tests `TestCalendarInboundMessageReceiptCompletionPersistsOccurrenceSnapshotWithCAS` and `TestMailboxEntRepositoryResolveCalendarEventForMessageSuppressesTargetForMissingSeriesAttendee` passed in unit at `07:46:37Z` / `07:50:32Z` and race at `07:59:40Z` / `08:03:57Z`. This is targeted packet evidence, not broad-green or full-Hub acceptance. Latest production backend and hub-mail image builds PASS.

The final clean tmpfs stack started the explicit Mail/Calendar set (postgres, mongo, redis, nats, seaweedfs, hub-mail, backend, frontend; LiveKit excluded because host port 7880 belongs to parallel work) and auto-seeded on its first `POST` 200 attempt, ready in 72.169s at `2026-09-05T08:29:12Z`. Runtime occurrence RSVP returned 200 with byte-identical same-key replay 200 and only the occurrence became accepted/v2 while the linked series remained needs_action/v1. After an exact stack reset without rebuild, a pristine first-attempt auto-seed again passed in 72.169s; the final read-only pair shared one event and both were needs_action/v1, with 7 threads at `localhost:8088`. backend/frontend/hub-mail/PostgreSQL/Mongo/Redis/NATS were healthy and SeaweedFS was up. Browser visual and real external Stalwart/JMAP remain `NOT_RUN`; push is blocked/pending and merge/release remain pending.

## Global Constraints

- Docker-only build, validation, tests, and runtime checks; no host validation.
- Preserve the dirty worktree and never read or change protected `pro-qa-ai/**`.
- Stage only exact owned paths; inspect the staged diff before commit.
- Do not merge to `master`; after the authorized commit, push the branch immediately.

## Task 1: Add the durable inbound receipt snapshot

- [ ] Implement the durable inbound receipt snapshot.
  - Files: `backend/ent/schema/calendarinboundmessagereceipt.go`, `backend/internal/calendar/event_ent_inbound_message.go`, receipt/domain contracts that carry normalized `ITIPIdentity`, and a new additive migration under `backend/migrations/atlas/mail-calendar-v1/` with its `atlas.sum` hash.
  - Implement nullable `calendar_component_scope` (`series|occurrence`) plus nullable `recurrence_id`; write tri-state unknown/series/occurrence from normalized identity, preserve matching existing data, and reject contradictory overwrite. Old, unparsed, and nonterminal rows remain unknown; a successfully parsed `deferred_missing_master` receipt stores the canonical occurrence pair, but has no Mail RSVP authority until an allowed terminal `applied|duplicate|ignored_stale` result.
  - Add genuine isolated Go units under `backend/tests/calendar/` covering unknown, series, occurrence canonicalization, preserved snapshot, and incompatible overwrite. No static/source-text tests.
  - Acceptance: schema/migration contracts validate in the schema-maintenance Docker service; focused calendar units pass in `docker compose -f docker-compose.local.yml run --rm backend-unit` using the exact focused package selector supported by the service.

## Task 2: Project a fail-closed RSVP target through Mail

- [ ] Implement the Mail RSVP target projection.
  - Files: `backend/internal/mail/mailbox_calendar_correlation.go`, `backend/internal/mail/mailbox_message_repository.go`, `backend/internal/mail/mailbox_message_service.go`, Mail HTTP DTO/handler/OpenAPI files that expose message detail, and affected Mail domain documentation.
  - Add nullable typed `calendar_rsvp_target { event_id, scope, recurrence_id }` beside existing `calendar_event_id`. Emit it only for correlated inbound terminal `applied|duplicate|ignored_stale` receipts with a valid known snapshot and existing tenant/account/actor/visibility fences; otherwise emit null. Keep event navigation bound to the master event.
  - Add genuine isolated Go units under `backend/tests/mail/` for series projection, occurrence projection, invalid pair, nonterminal/outbound/legacy null paths, and visibility fence failure.
  - Acceptance: focused Mail units pass through `backend-unit`; generated/API contract is refreshed by the project Docker-only maintenance route where required.

## Task 3: Make the invitation UI select the authoritative endpoint

- [ ] Implement authoritative series-versus-occurrence invitation RSVP selection.
  - Files: `frontend/src/api/mail.ts`, `frontend/src/modules/mail/mail-calendar-invitation.tsx`, `frontend/src/modules/mail/mail-message-detail.tsx` (or its exact caller), `frontend/src/modules/mail/mail-query-keys.ts`, `frontend/src/routes/_authenticated/mail/message/$messageId.tsx`, and `frontend/src/modules/mail/mail-composer.tsx`.
  - Extend the Mail detail type with `calendar_rsvp_target`; pass it intact to the invitation component. Use series RSVP only for `scope=series`, existing occurrence RSVP only for `scope=occurrence`, and never issue an RSVP when target is null/invalid. On occurrence success, render Calendar's returned effective attendee/CAS state; master-event navigation remains unchanged. Introduce canonical `mailKeys` factories for `threadsRoot`, `drafts`, `directMessage`, and `identities`, then replace ad-hoc equivalents.
  - Add genuine Vitest/RTL units under `frontend/tests/mail/` for series endpoint selection, occurrence endpoint selection, null target non-actionability, and returned effective attendee rendering. No static contract tests.
  - Acceptance: focused frontend units run only through `docker compose -f docker-compose.local.yml run --rm frontend-unit` with the service's supported selector.

## Task 4: Extend deterministic dev data and prove the vertical slice in Docker

- [x] Extend deterministic dev data and prove the vertical slice in Docker.
  - Files: the existing Mail/Calendar dev autoseed fixture and its documented recovery path, plus `docs/mail/QA-HANDOFF.md` and `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`.
  - Add one deterministic inbound occurrence REQUEST fixture that produces an actionable occurrence target without relying on manual database mutation. Preserve current series/demo messages.
  - Run one final low-disk Docker compose stack with automatic seed. Verify raw API detail data for series and occurrence, the corresponding `/rsvp` and `/instances/{recurrence_id}/rsvp` paths, truthful returned attendee state, repeat/idempotency behavior, master-event navigation ID, and no startup error regressions. Browser verification is recorded as `NOT_RUN` if the host browser sandbox remains unavailable.
  - Acceptance: PASS for the final pristine clean tmpfs stack: first automatic seed attempt succeeded in 72.169s at `2026-09-05T08:29:12Z`; this supersedes the earlier 72.186s bounded measurement. The explicit Mail/Calendar set (postgres, mongo, redis, nats, seaweedfs, hub-mail, backend, frontend; LiveKit excluded because host port 7880 belongs to parallel work) was healthy/ready, and the API through actual nginx frontend `localhost:8088` proved both targets/action paths. Before the exact no-rebuild reset, series began `series/null/v1/needs_action`; occurrence began `occurrence/2026-09-10T10:00:00Z/v1/needs_action`, accepted `POST` 200, and same-key replay returned byte-identical 200. Reload showed occurrence `accepted/v2` while series stayed `needs_action/v1`; PM outbound retained nonnull `calendar_event_id` and null target. DB confirmed the applied occurrence receipt, exact recurrence, effective occurrence attendee accepted/v2 with unchanged series attendee, detached-exception reply delivery sequence 1 over master sequence 0 (`accepted/pending`), and completed/200 matching-SHA mutation receipt. The final read-only post-reset pair shared one event and both were needs_action/v1, with 7 threads. Private Ent receipt behavior remains private: this autoseed/runtime proof replaces a widened production API solely for an isolated unit seam. Browser visual remains `NOT_RUN`; external real Stalwart/JMAP remains `NOT_RUN`.

## Task 5: Review, synchronize docs, and publish the bounded packet

- [ ] Review, synchronize docs, and publish the bounded packet.
  - Re-read changed diffs against [the design](../specs/2026-09-05-mail-occurrence-rsvp-design.md); obtain independent backend and frontend review with findings recorded and fixed or explicitly retained.
  - Run the exact Docker-only focused checks from steps 1-4 once after the final diff; do not run host build/test/lint commands or broad destructive cleanup.
  - Update affected Mail/Calendar API and operational documentation in the same session. Stage only exact owned paths, inspect `git diff --cached`, commit without AI attribution, and immediately push `codex/mail-module-20260813`. Do not merge to `master`.
