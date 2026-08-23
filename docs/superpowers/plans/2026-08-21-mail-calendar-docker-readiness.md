# Mail + Calendar Docker readiness implementation plan

> **P3-D0 static acceptance (2026-08-23):** immutable source `47faf3a705e5641f4ceb7ee365c3dfeeeace725d` is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly `backend/ent/schema/calendarinboundmessagereceipt.go`. It adds required immutable `organization_id`/`mailbox_id` UUID scalars and database nonzero checks, with no Ent edges or raw data; existing uniqueness, state-shape, and due behavior are unchanged. Independent Contract/high and Security/xhigh review is APPROVED C/I/M `0/0/0`. The D0-next route is `CONSUMED/HISTORICAL`. P3-C remains blocked on Mail Ent generation; 147-02B/147-02C and D1/D2 remain blocked on active Calendar generation/dependencies; P3-E is conjunction-blocked. No next unblocked P3 source is selected. Generation, migration, SQL, build, tests, Docker, and runtime are `NOT_RUN`; no readiness, merge, or product claim.

> Execute atomically in the listed order. Every code packet is test-first, independently reviewed,
> committed with exact-path staging, and verified only through the accepted Docker launcher.

**Design:** `docs/superpowers/specs/2026-08-21-mail-calendar-docker-readiness-design.md`
**Branch:** `codex/mail-module-20260813`
**Do not touch:** `master`, remotes, deploy targets, shared/prod databases, `pro-qa-ai/**`

## Task 1 — Reconcile current accepted packet receipts

**Files:** current Phase 143/144 state carriers only, exact manifest frozen before editing.

- Record immutable accepted source/receipt commits already present at current HEAD.
- Remove only active stale candidate routes; preserve historical provenance.
- Keep generation, migration, adapter, API, launcher, and runtime as `NOT_READY` / `NOT_RUN`.
- Review lifecycle truth, commit exact docs, and perform immutable exact-SHA review.

## Task 2 — Freeze launcher bootstrap revision

**Primary files:**

- `testframework/runner/run-task.ps1`
- new closed public launcher in `testframework/runner/`
- `testframework/selftests/runner/**`
- `testframework/docker/{build-manifest.v2.json,source-policy.v1.json,compose.unit.yaml}`
- `.cursor/skills/qa-gate/SKILL.md`
- runner/docker/package documentation and affected planning state

**Behavior:**

- derive worktree-local canonical roots from the launcher location;
- accept only fixed `target`, `task`, and `run`;
- preserve the four-operation private dependency seam;
- rebase source/control provenance to the exact accepted commit;
- expose no arbitrary command/path/image/Compose/environment input;
- leave targets `NOT_READY` until dependency and image admission completes.

**Test first:** extend runner selftests for linked-worktree containment, typed path attacks, report
collision, public/private argument closure, exit propagation, and cleanup ownership.

## Task 3 — Admit locked dependencies and build offline G0 images

**Primary files:**

- `testframework/docker/admit-local-dependency-assets.ps1`
- `testframework/docker/source-image-attestor.ps1`
- `.cursor/skills/qa-gate/profiles/local-dependency-admission.v1.json`
- three target Dockerfiles and image-baked entrypoints
- `testframework/docker/images.lock.v3.json` produced only by the accepted attestor

**Behavior:**

- consume only manifest/lock-pinned dependencies;
- verify exact local base IDs and `linux/amd64`;
- bake Ent generation before backend commands;
- bake fixed frontend and OpenAPI commands;
- publish a strict v3 lock without overwrite;
- run with no network, pull, source bind, package install, or checkout read.

**Evidence:** D0 receipts, image attestations, nine-label equality, Compose raw-byte equality,
task-owned report, and cleanup zero proof.

## Task 4 — Accept and run G0 targets

- Run `backend-unit`, `frontend-unit`, and `backend-openapi` through the public launcher.
- Triage every failure against current docs before changing code or tests.
- Require non-zero discovery, zero unexpected skips, no timeout, and exact cleanup.
- Repeat the failing target after each fix and all three on the final G0 revision.

## Task 5 — Bootstrap Python QA and pre-author closed vertical targets

- Create `testframework/pyproject.toml`, generated `uv.lock`, installed `hub_testkit`, `py.typed`,
  and failing-first framework selftests.
- Accept closed `python-selftest` and `python-contract` catalog entries. Later vertical targets are
  added one at a time, in D4 order, immediately before their first authored-first oracle and SUT
  packet. A target may initially collect only its authored-first oracle and fail for the missing
  SUT; zero collection is never accepted.
- Keep all pytest execution behind the public launcher; no direct pytest or Compose path.
- Freeze literal evidence, skip, marker, artifact, timeout, resource, and cleanup rules per target.

## Task 6 — Add Atlas migration target

**Primary files:**

- new `backend/atlas.hcl`
- new immutable versioned pre-feature baseline fixture and ordered
  `backend/migrations/mail-calendar/**`
- reviewed Atlas CLI tool lock/build/attestation metadata
- migration image/entrypoint and closed launcher target
- contract/integration fixtures under `testframework/**`
- migration runbook and state/docs

**Behavior:**

- build the exact pinned Atlas CLI inside the offline admitted image and attest its binary/version;
- load the exact accepted pre-feature full-schema baseline into an empty database, then apply only
  the Mail/Calendar feature lineage;
- apply the same lineage to a realistic pre-feature fixture;
- compare generated Ent metadata only for the allowlisted owned Mail/Calendar schema and shared
  `scheduled_calls` delta, while proving unrelated baseline schema did not change;
- verify Calendar partial uniques, exception/receipt constraints, ScheduledCall edge, and all Mail
  tables/indexes/delete actions;
- reject legacy recurring ScheduledCall backfill without mutation;
- prove no historical outbox/event side effects;
- use forward-only correction after apply.

## Task 7 — Calendar concrete persistence

**Primary files:** new Ent adapters under `backend/internal/calendar/`, existing frozen ports, new
genuine units, and Python integration tests.

- First accept the closed `calendar-integration` target, author its failing persistence oracle, and
  prove the target collects/fails for the missing adapter before implementation.
- Implement create, update/master cancel, and occurrence-cancel transaction capabilities.
- Preserve event-before-call locking, organization masking, version/fence precedence, exact
  persistence echo, attendee tombstones, and one-transaction outbox semantics.
- Add canonical read/availability repository without modifying the legacy aggregate implicitly.
- Verify rollback, replay, duplicate receipts, stale versions, recurrence exception conflicts,
  cross-org privacy, and two-replica races in Docker PostgreSQL.

## Task 8 — Calendar RSVP and v2 HTTP vertical

- First accept the closed `calendar-api` target and author failing RSVP/v2 contract and API oracles.
- Freeze and review the missing RSVP command/result/transaction port.
- Implement series and occurrence RSVP with optimistic versioning and receipt replay.
- Add dedicated v2 DTO/handler mapping for create, update, cancel, occurrence cancel, RSVP, reads,
  and availability.
- Regenerate and compare backend OpenAPI through `backend-openapi`.
- Preserve legacy endpoints until the bridge cutover is proven.

## Task 9 — Calendar ScheduledCall transaction bridges

- Implement 04F-H occurrence persistence, 04F-J event-to-call projection, and 04F-K due-job
  reservation as separate adapters.
- Replace wall-clock/direct-NATS legacy cron behavior only after same-transaction outbox and replay
  recovery pass.
- Verify call/event lock order, occurrence key dedupe, EXDATE/cancel behavior, retry/redelivery, and
  no double-counted aggregate rows.

## Task 10 — Mail concrete sync persistence and Phase 145 engine runtime

**Primary files:** new `mailbox_ent_repository_sync.go`, a separately frozen real engine adapter,
worker composition, genuine units, and Python integration tests.

- First accept the closed `mail-integration` target and author its failing persistence/engine
  oracles.
- Implement claim/prepare/apply/resolve plus due scan/fault reschedule with DB transaction time and
  full account/lease/sync/mutation fences.
- Make projection, counts, cursor, and outbox atomic.
- Add a real engine adapter only after authentication and DNS readiness contracts are frozen.
- Verify stale leases, ambiguous apply, replay, rebuild without pre-clear, engine outage/restart,
  account isolation, and two-replica races.
- Admit and attest the selected Stalwart image/license, implement authenticated provisioning/JMAP
  adapters, and prove bounded persistent profile health, restart, shutdown, and resource limits.
- Keep external-domain delivery disabled until DNS/PTR/network acceptance is independently green.

## Task 11 — Mail provisioning, product services, and public API

- First accept the closed `mail-api` target and author failing provisioning/owner API oracles.
- Implement account/domain/identity provisioning before enabling mailbox sync traffic.
- Add folder/label/message, draft/submission, settings/signature, rule, storage, search/contact,
  outbox/realtime/audit seams in their existing phase order.
- Add owner-default-deny handlers with opaque 404, `Idempotency-Key`, `If-Match`, bounded DTOs, and
  no raw-body persistence.
- Wire repository, engine, services, worker timer/config, root-context shutdown, and readiness in
  the server composition root.

## Task 12 — Phase 146 Mail workspace

- Implement the responsive three-pane Mail workspace and genuine units under
  `frontend/tests/mail/**` only.
- Cover triage/read/compose/search/RSVP, autosave/send ambiguity, scan states, remote-image consent,
  projection rebuilding, provisioning, CRUD and admin-alias states across desktop/tablet/mobile.
- Verify accessibility and API journeys through accepted `frontend-unit`, `mail-api`, and bounded
  smoke targets; preserve honest unavailable/degraded states.

## Task 13 — Phase 147 reliability and integration closure

- Implement or explicitly founder-defer recurrence/import recovery, busy/free with approved leave,
  reminders, L1 feeds, and CRM send/manual-log seams according to the Phase 147 contract.
- Prove out-of-order/redelivery convergence, detail non-disclosure, linked event/call dedupe, SSRF
  and bounds, and reuse of accepted policy boundaries.
- An explicit deferral remains visible and prevents a phase-complete claim; it is not silently
  converted to PASS.

## Task 14 — Combined Docker stack and full QA

- Add one task-owned `mail-calendar-smoke` topology with PostgreSQL, NATS, required storage,
  backend, frontend if exercised, and the real mail engine fixture.
- Apply migrations, start services, wait on bounded readiness, run smoke journeys, capture logs and
  receipts, then perform unconditional exact cleanup.
- Run selected regression targets and independent code/security/lifecycle reviews on the exact
  final commit.
- Run non-zero `python-contract`, Calendar/Mail API and integration, frontend Mail unit, and smoke
  collections through the same public launcher; zero unexpected skips and exact cleanup are gates.

## Task 15 — Phase 148 final branch gate

- Synchronize `.planning/STATE.md`, coverage/test docs, architecture, deploy/runbook, Mail/Calendar
  contracts, and final evidence links.
- Reconcile independent security/recovery findings and produce a current-SHA evidence manifest.
- Stop with a clean merge-ready branch. Do not rebase onto a protected-program result, merge, push,
  deploy, or apply to shared/prod state until the founder gives the separate signal.
