# Project Outbox — shared QA execution plan (P1–P7)

> **Status:** `ROOT_DECISIONS_FROZEN / P0_ACCEPTED / P1A_CLEARED_FOR_ATOMIC_ADMISSION` — design and execution decomposition only.
>
> **Inputs:** frozen `.planning/active/august-2026/07-PROJECT-OUTBOX-SHARED-QA-CONTRACT.md`; the five `TC-OUTBOX-0001`…`TC-OUTBOX-0005` records; current `project` create/outbox path; the QA-only Mail/Calendar control-plane pattern; and the sealed Python integration suite/plugin boundary.
>
> **Evidence boundary:** the independent P0 review accepted the remediated frozen protocol; canonical Docker `python-contract` is **280/280 PASS** (seed `4057944629`, `5.39s`). This clears only P1a's frozen atomic admission. P1a–P7 are `NOT_RUN`; this packet makes no product, migration, build, lint, runtime, merge, or deployment claim.

## Goal

Replace the currently abstract/exclusive `ProjectOutboxHarnessDelegate` expectation with one HTTP-only Python delegate over the fixed shared `qa-*` substrate. A QA-only server control plane owns one exact manifest per case, creates its one durable JetStream consumer before any project publication, registers immutable fixture and project-side-effect UUIDs, serves constrained projections/receipts, and performs manifest-only cleanup.

The product request under test remains `POST /api/projects`; Python must not open a PostgreSQL, Redis, NATS, JetStream, Docker, or Compose client. The private control plane is compiled only with `qa && !release` and is absent from a release build.

## Locked invariants

1. The substrate is shared: a case never owns or deletes `qa-backend`, databases, `EVENTS`, NATS, Redis, caches, containers, networks, or a shared service.
2. Fixture organization UUID and owner UUID are immutable manifest roots. No name, email, nonce, request correlation, prefix, received message, or cleanup result may become a discovery key for data ownership.
3. There is exactly one durable `EVENTS` consumer per case. It is created successfully before fixture readiness is returned and is never replaced during reconnect.
4. All receipt and final-state observation has one monotonic maximum of **3 seconds**, including client retry/polling. A single HTTP operation keeps a 5-second read budget; timeout/absence is never zero evidence.
5. A successful project creation must register its exact project UUID, owner-membership UUID, and outbox-event UUID before the public `201` can be returned to the Python delegate. Receipt equality is exact for `project_id` and `created_by_id`.
6. `TC-OUTBOX-0003` and `TC-OUTBOX-0004` establish zero only through a constrained server projection rooted in immutable fixture UUIDs and the manifest. A missing durable receipt, empty cleanup response, consumer timeout, or general absence from a shared system is not an oracle.
7. Cleanup consumes only enumerated IDs from the manifest, is retry-safe for an HTTP lost response or a partial cleanup, and never expands its target set. A retained `cleanup_failed` manifest is a failure record, not a permission to discover candidates.
8. Existing five test functions and their sealed `conftest.py` remain thin mapped nodes. Factory composition is installed through the public pytest plugin hook if its scope gate proves valid.

## Current facts that constrain implementation

| Fact | Consequence |
|---|---|
| `project.EntRepository.Create` makes the project, owner membership, and `project.created` outbox row inside an Ent transaction, but returns only `*ent.Project`. | The exact membership and outbox UUIDs cannot currently be carried to a manifest.
| `OutboxEvent` has `id`, `event_type`, opaque `payload`, status, attempts, and timestamps, but no organization, owner, project/aggregate, fixture, or operation column. | It cannot produce an authoritative negative/zero projection scoped to a fixture without a new transactional correlation authority. Parsing or scanning payloads would violate P0.
| `ProjectService.CreateProject` invokes tracker workflow/board, TMS suite, and dashboard provisioners after the repository transaction, ignores their failures for the response, and exposes no closed effect receipt. | A normal fixture project has an open-ended post-commit side-effect surface. Deleting by project cascade would not satisfy the frozen requirement to enumerate exact IDs.
| The outbox worker publishes through JetStream `EVENTS` with `InterestPolicy`; `project.created` has no project UUID in its subject. | The consumer must exist first, filter `project.created`, and inspect only a bounded stream of its own consumer deliveries for exact payload equality. It cannot infer zero from no delivery.
| `AuthService.CreateSession` creates a session with a UUID in the JWT `sid` claim, but returns no session ID. Existing `/api/dev/login` is email-based and does not register the created session in a fixture manifest. | The control plane must issue the owner session itself, parse/verify the exact `sid`, and record it before returning an access token. Python must not use email login as the outbox fixture authority.
| `pytest_configure` already has the selected `SuiteBinding` and public stash API; it runs before the sealed integration fixture creates a runtime. | A lazy factory can be installed only for `suite_id == "integration"` without editing that suite's `conftest.py`.

## Frozen root decisions

The following are closed architectural decisions and execution constraints, not open blockers.

### B1 — transactional exact-effect authority (`FROZEN`)

The current API/repository path does not expose the owner-membership or outbox-event UUID, while `outbox_events` cannot be safely searched by fixture roots. Therefore a controller that reads rows after `201`, scans byte payloads, or reconstructs targets from a project key/name is forbidden.

**Frozen decision:** additive untagged `QAProjectOutboxManifest` and `QAProjectOutboxEffect` Ent schemas/tables are permitted in the existing `mail-calendar-v1` Atlas lineage. The QA-only boundary is access, writes, routes, and composition wiring under `qa && !release`, not Ent schema build tags; release builds contain only inert generated types/tables and expose no control surface or authority writes. The authority is written in the same Ent transaction as the project, membership, and outbox row. It records only exact UUID authority and hashed control material; it has no text-discovery field. A narrow transaction-local recorder receives those effects. The publisher returns the exact just-created event UUID. A memory-only map, post-`201` lookup, payload parsing, or generic scan is forbidden. The public response remains unchanged.

The immutable manifest fields are `id`, `scenario`, `nonce_hash` (SHA-256), `control_capability_hash` (SHA-256), `creation_generation`, `organization_id`, `owner_user_id`, `session_id`, `consumer_stream`, `consumer_durable_name`, `consumer_filter_subject`, and `consumer_config_fingerprint` (SHA-256). Its mutable lifecycle is limited to closed `state`; cached receipt `project_id`, `created_by_id`, `cached_at` as an all-or-none triple; per-root cleanup completion timestamps; `cleanup_error_class`; `tombstoned_at`; `created_at`; and `updated_at`. The immutable effect fields are `id`, `manifest_id`, `ordinal`, `organization_id`, `owner_user_id`, `project_id`, `project_member_id`, `outbox_event_id`, and `registered_at`; only exact per-target cleanup completion timestamps are mutable. Every UUID authority is non-zero; each hash/fingerprint is 32 bytes; state/tombstone consistency and cached-receipt triple consistency are checked. Unique manifest fields are `nonce_hash`, `organization_id`, `owner_user_id`, and `session_id`; unique effect constraints are `(manifest_id, ordinal)`, `(manifest_id, project_id)`, `(manifest_id, project_member_id)`, and `(manifest_id, outbox_event_id)`. The sole relation is `Manifest -> Effects` through required immutable `Effect.manifest_id` with `OnDelete(NoAction)`; there are no FKs/edges to `User`, `Organization`, `Session`, `Project`, `ProjectMember`, or `OutboxEvent`. Durable name is an exact resource handle, never lookup authority except manifest ID plus exact recorded stream/name. Raw UUID authority is retained through cleanup/tombstone.

### B2 — closed project side-effect set (`FROZEN`)

The frozen contract requires enumerated project-side-effect IDs. The present `ProjectService` post-commit provisioners create tracker workflow/status/transition/board, TMS suite, and dashboard state after the initial project transaction, without returning a closed receipt.

**Frozen decision:** QA-build-only provisioner suppression applies solely to the admitted exact fixture owner. Public authentication, parsing, project/member/outbox transaction, response, and worker execution remain real. The manifest set is limited to project, owner membership, and outbox event. Any additional created record must be appended to the persisted effect list before use or the operation fails closed. Cascades are never implicit inventory.

### B3 — loss semantics for session and project create (`FROZEN`)

Control-plane commands can be idempotent by their caller-supplied manifest nonce and recorded state. A product `POST /api/projects` currently has no idempotency key semantics; its Python `HttpRequest` explicitly carries no key. A lost public response must therefore be reported as **unknown product outcome**, followed only by manifest-status/cleanup recovery; the delegate must not replay the create and pretend a `409` is a recovered `201`.

**Frozen decision:** recovery is same-process only. An ambiguous/lost public `POST /api/projects` is never replayed; only manifest-status and manifest cleanup are permitted. Server restart and exact-session token reissue are future scope and are not P1–P7 blockers.

## Target architecture

```text
pytest thin case
  -> sealed conftest fixture
  -> plugin-installed ProjectOutboxHTTPFactory
  -> HTTP-only ProjectOutboxHTTPDelegate
       -> private QA control plane       /api/_qa/project-outbox/fixtures/...
       -> real authenticated product API POST /api/projects
       -> private projection/receipt/cleanup endpoints
  -> QA-only controller + durable manifest/effect authority
       -> exact Ent queries and exact session lifecycle
       -> one EVENTS durable pull consumer, FilterSubject=project.created
       -> no Python database/NATS access and no shared-resource lifecycle
```

### Manifest model

The manifest is addressed by a 32-lower-hex caller nonce but **is not constructed from it**. The nonce merely selects the one server-recorded manifest. The record stores no user-visible discovery name or email as an ownership selector.

| Field family | Required immutable or recorded values |
|---|---|
| Identity | `manifest_id`, scenario `project-outbox-v1`, `nonce_hash` (never raw nonce), state, creation generation, hashed control capability, and command receipt state. |
| Fixture roots | exact `organization_id`, `owner_user_id`; each has a server-generated UUID and cannot be changed after fixture creation. |
| Auth | exact preallocated `session_id`; access token is secret/redacted and never appears in a state/diagnostic response. The QA-only issuer verifies the JWT `sid` equals that UUID before fixture readiness. Session recovery is same-process only. |
| Consumer | stream `EVENTS`, durable name `qa-project-outbox-<manifest-id>`, filter `project.created`, consumer configuration fingerprint, connection state, and at most one cached typed receipt. |
| Project effects | ordered entries containing exact `project_id`, `project_member_id`, `outbox_event_id`, and, only if B2 chooses full provisioning, every exact provisioner-effect UUID with kind. |
| Cleanup | immutable original target lists, per-target completion bits, terminal error class, and a retained deleted tombstone. A retry may change completion bits but may not add an ID. |

The server keeps an owner-UUID-to-manifest index only as a direct map/transactional exact lookup. It never lists manifests or chooses one by owner email, project key, or text prefix.

### State machine

| State | Permitted transition | Required condition |
|---|---|---|
| `creating` | `ready` | exact roots, session, consumer, manifest, and every currently created effect are persisted/verified before response. Failure moves to `cleanup_failed` with only already recorded IDs. |
| `ready` | `project_registered` | the normal project route committed; transaction-local recorder persisted exact project/member/outbox IDs before the public `201`. Any additional created record must first be appended to the manifest or the operation fails closed. |
| `ready` or `project_registered` | `observer_disconnected` | same durable consumer exists and is verified; no deletion/replacement occurs. |
| `observer_disconnected` | previous ready/registered state | reconnect verifies the same stream/name/configuration fingerprint and resumes that consumer only. |
| `project_registered` | `receipt_cached` | a bounded consumer delivery passes exact `project_id` and `created_by_id` checks, is cached before acknowledgement, and then is explicitly acknowledged. |
| any nonterminal state | `quiescing` | capability admission closes; no new product create can enter this manifest. |
| `quiescing` | `deleting` | exact consumer is removed or already confirmed absent; no shared stream action is performed. |
| `deleting` | `deleted` | every recorded target has been removed/confirmed absent by exact ID, in the stated order. Retain tombstone for idempotent `DELETE`. |
| `creating`, `quiescing`, or `deleting` | `cleanup_failed` | an exact operation failed or a precondition was unsafe. Retain the same target list; retry resumes only unfinished exact targets. |

Repeated create/session/observer commands with the same nonce return the same manifest state; a conflicting scenario/capability returns `409`. Unknown nonce, unexpected state, consumer fingerprint mismatch, or an attempt to add a target returns `409` after control-token authentication. No endpoint creates a second consumer.

## Private HTTP control-plane contract

All paths below require `X-Hub-QA-Project-Outbox-Token`, a dedicated 64-lower-hex launcher secret. Do not reuse `HUB_QA_MAIL_CALENDAR_TOKEN`: Mail/Calendar and project/outbox are distinct capability scopes. The endpoints reject query strings, unexpected body/content type, duplicate JSON keys, and unrecognized fields; errors are a fixed redacted `{"error":"..."}` body.

| Method and path | Success | Purpose and constraints |
|---|---|---|
| `POST /api/_qa/project-outbox/fixtures` with `{"nonce":"…","scenario":"project-outbox-v1"}` | `201` new, `200` exact idempotent retry | Creates/recovers the manifest, root organization/owner, exact owner session, and durable consumer before returning a redacted fixture snapshot plus bearer access token. The token is never persisted in reports; restart/token reissue is out of scope. |
| `GET /api/_qa/project-outbox/fixtures/:nonce` | `200` | Returns redacted manifest snapshot and state only. It is recovery metadata, not a state projection and not receipt evidence. |
| `POST /api/_qa/project-outbox/fixtures/:nonce/observer/disconnect` | `204` | Marks the sole durable observer disconnected without deletion; a duplicate is `204`. |
| `POST /api/_qa/project-outbox/fixtures/:nonce/observer/reconnect` | `204` | Reuses/verifies the same durable consumer; a duplicate is `204`; no consumer creation call is allowed here. |
| `GET /api/_qa/project-outbox/fixtures/:nonce/receipt` | `200` exact receipt; `408` at 3 seconds | Bounded pull from only the manifest consumer. `408` means observation expired and is never zero evidence. A cached acknowledged receipt is returned idempotently. |
| `GET /api/_qa/project-outbox/fixtures/:nonce/projection` | `200` | Returns only `project_count`, `owner_membership_count`, `outbox_row_count`, and `published_outbox_row_count` from constrained server queries. It may poll exact recorded outbox IDs only until the 3-second deadline. |
| `DELETE /api/_qa/project-outbox/fixtures/:nonce` | `204` | Runs the manifest-only cleanup protocol; terminal tombstone retry is also `204`. A partial failure is `500` with a redacted fixed error and retained manifest. |

Malformed private request is `400`; missing/wrong control token is `401`; lifecycle/capability/fingerprint conflict is `409`; server/internal operation failure is `500`; a 3-second observe expiry is `408`. The ordinary product statuses remain exactly `201`, `400`, `401`, and `409` as consumed by existing outbox models/oracles.

## JetStream and observation rules

1. The controller uses the already ensured shared stream `EVENTS`; it never creates, updates, purges, or deletes the stream.
2. Fixture creation creates one pull durable consumer with `FilterSubject = "project.created"`, explicit acknowledgement, `DeliverNew`, `MaxAckPending = 32`, and a manifest-recorded configuration fingerprint. It is created before owner session response and before any product request is possible.
3. Each receipt/projection observation uses one monotonic 3-second deadline, pull batch size `8`, and an examined-delivery ceiling of `64` across the whole operation. Because the subject is not project-qualified, foreign `project.created` messages can arrive in this consumer. The controller acknowledges foreign deliveries only as copies delivered to this consumer, never treats them as manifest targets/evidence, and accepts a receipt only when its closed JSON payload equals a registered `project_id` and immutable owner UUID. Reaching the ceiling before an exact receipt returns fixed redacted `500`; reaching the deadline returns `408`.
4. `disconnect` changes only controller pull activity. `reconnect` uses the same durable identity; it never calls consumer creation. `TC-OUTBOX-0005` therefore creates while this activity is detached, reconnects the same consumer, then reads the retained message.
5. Receipt caching occurs before acknowledgement so an HTTP response lost after acknowledgement can be retried while the manifest remains available. A receipt timeout, a foreign message, or a cleanup result cannot be used for the zero cases.
6. The outbox worker is not paused, killed, reconfigured, or made a fixture resource. A projection waits only for exact recorded event IDs to reach the stated final condition; a deadline is a test/control failure, never a cross-substrate cleanup action.

## Authoritative server projections

The server returns counts, not raw database rows, to Python. Its implementation is constrained as follows.

| Count | Positive projection | Zero projection |
|---|---|---|
| Project | exact registered `project_id` set, additionally constrained to immutable fixture organization and owner membership | `Project` constrained by immutable fixture organization UUID and owner UUID; expected zero, with no lookup by name/key/nonce. |
| Owner membership | exact registered membership IDs, each additionally bound to its registered project and immutable owner | `ProjectMember` constrained by immutable owner UUID and fixture-organization project scope; expected zero. |
| Outbox row | exact registered `outbox_event_id` set, event type `project.created`, and exact registered correlation/effect record | requires B1's transactional correlation authority; current opaque byte payload is not an admissible negative query. |
| Published outbox row | exact registered `outbox_event_id` set with `status = published` | zero under the same B1 authority; an absent receipt does not contribute. |

For `0001`, `0002`, and `0005`, the expected tuple is `(1, 1, 1, 1)` after the exact receipt. For `0003` and `0004`, the expected tuple is `(0, 0, 0, 0)` solely from this projection. The control plane must return a failure rather than an optimistic zero if the B1 correlation record is unavailable.

## Cleanup protocol and failure handling

Cleanup never discovers a target. It executes this order from the persisted manifest and marks each completed exact ID atomically before the next operation:

The only admissible fixture lifecycle set is: exact organization, owner user, owner session, the auth-service-owned exact Redis refresh/session lifecycle, one durable consumer, and persisted manifest/effect records. Project, owner-membership, and outbox IDs enter only through the same-transaction P1c recorder. Every additional created record must be appended to the manifest before use; if it cannot be appended, creation fails closed and cleanup uses only already recorded IDs.

1. Close fixture admission, require no in-flight registered create, and detach/delete the manifest durable consumer only. If the consumer cannot be proven to match its recorded stream/name/configuration, stop with `cleanup_failed`.
2. For each project effect, delete provisioner leaves in reverse dependency order only if B2 selected full provisioning; otherwise assert no provisioner-effect list exists. Delete the exact owner membership, exact project, and exact outbox row with exact-ID/expected-state preconditions. The controller must not bulk-delete by organization, owner, project key, or a payload scan.
3. Require every recorded outbox row to be terminal before deleting it. A pending/claimed row that cannot be safely resolved is `cleanup_failed`; the controller neither stops the shared worker nor guesses a lease outcome.
4. End the exact owner session through the auth service using the recorded `session_id`, then delete the exact owner and exact organization. Refresh-token/session cleanup must be performed by the auth service, not by a raw broad session delete.
5. Persist a tombstone with the original manifest identity and completed target bits. A repeated `DELETE` returns `204`; it does not recreate/requery resources.

If an HTTP cleanup reply is lost, the same `DELETE` may resume the retained manifest. If a database/NATS/auth operation returns an uncertain result, retain the manifest and return a redacted `500`; a retry is limited to recorded unfinished IDs. Server restart without B1/B3 persistence is an explicit recovery blocker, not permission for a scan.

## P1–P7 executable packets

Every code packet below is atomic, owned by one `gpt-5.6-terra` / `xhigh` implementer, and independently reviewed at `high`. The root remains responsible for the frozen contract, B1–B3 decisions, model provenance, packet admission, and final acceptance. No packet may change `pro-qa-ai/**`, run a host test, or claim runtime success.

### P1a — schema/migration effect authority

**Depends on:** P0 independent acceptance (`ACCEPTED / CLEARED_FOR_P1A`).

**Owner files (exactly six):** `backend/ent/schema/qa_project_outbox_manifest.go`; `backend/ent/schema/qa_project_outbox_effect.go`; `backend/migrations/atlas/mail-calendar-v1/20260831000000_qa_project_outbox_authority.sql`; `backend/migrations/atlas/mail-calendar-v1/atlas.sum`; `backend/docker/entrypoints/mail-calendar-schema.sh`; and `backend/migrations/README.md`. No application code or tests are in this atom.

**Deliverable:** additive untagged Ent schemas/tables in the existing `mail-calendar-v1` Atlas lineage with the frozen immutable/mutable field sets, checks, indexes, and exactly one `Manifest -> Effects` required immutable `OnDelete(NoAction)` relation. Schema generation is available in release as inert generated tables/types, while all QA authority access/writes/routes/wiring remains guarded by `qa && !release`. No raw nonce, payload, name, key, email, or discovery data may be stored; no external domain FK/edge is permitted.

**Static acceptance tests:** generation/migration admission is explicit; only the frozen 32-byte hash/fingerprint checks, non-zero UUID checks, closed state/tombstone and cached-receipt triple checks, unique indexes, and `Manifest -> Effects` no-action edge exist; no discovery field or external-domain FK/edge exists. Launcher admission updates the lineage SQL count `9 -> 10`, total entries `10 -> 11`, latest migration `20260829000001 -> 20260831000000`, and the exact snapshot/copy/allowlist. Foreign-manifest/root matching behavior is accepted only in P1c/P3 transaction/admission tests, not in this schema atom.

### P1b — publisher exposes exact event ID

**Depends on:** P1a.

**Owner files:** `backend/internal/platform/outbox/publisher.go` and focused isolated tests only.

**Deliverable:** add a separate exact-ID-returning API, analogous to `PublishInTxWithID`, for the project transaction recorder. Existing `PublishInTx` behavior and every Calendar/Tracker/Mail caller remain byte/API-semantically unchanged; there is no global signature or return-shape migration.

**Static acceptance tests:** the additive API returns the exact created event UUID; publisher failure exposes no success receipt; existing `PublishInTx` callers compile and retain their current behavior; no payload scan exists.

### P1c — repository transactional recorder

**Depends on:** P1a and P1b.

**Owner files:** `backend/internal/project/repository.go`, approved narrow recorder interface/implementation, and focused genuine isolated project tests only.

**Deliverable:** one transaction-local recorder writes exact manifest/project/member/outbox IDs in the same Ent transaction before commit visibility. It must prove the admitted manifest and immutable organization/owner roots exactly match the authenticated project create before appending an effect; refusal or any write failure rolls back all of them.

**Static acceptance tests:** success commits all three exact IDs only for the admitted matching manifest roots; a foreign manifest or mismatched organization/owner root records no effect and rolls back/rejects; publisher/repository/recorder error rolls back; zero path has no effect record; post-`201` recovery is impossible by construction.

### P2 — close the QA fixture manifest state machine

**Depends on:** P1c.

**Owner files:** new `backend/internal/qaprojectoutbox/config.go`, `manifest.go`, and focused isolated Go tests under `backend/tests/project/...`.

**Deliverable:** validated dedicated token/configuration; nonce addressing; immutable roots; command idempotency; allowed state transitions; target-list append before use; and retained terminal/failed records. The module and every authority writer/query are `qa && !release`; release product code can hold only a nil/no-op observer contract. The module holds no product route, Ent query, or NATS operation.

**Static acceptance tests:** wrong token/nonce/scenario rejected; duplicate create returns same identity; conflicting second scenario/capability rejected; no target can be appended after quiescing; deleted tombstone remains idempotent; snapshot/redaction excludes access and control secrets.

### P3 — make the real project route manifest-aware without changing its public contract

**Depends on:** P1c and P2.

**Owner files:** narrow additions in `backend/internal/project/service.go`, `provisioner.go`, and/or `handler.go`; the approved qaprojectoutbox admission adapter; focused project unit tests.

**Deliverable:** a QA-build-only injected observer keyed by immutable owner UUID and exact manifest root binding. On the authenticated public `POST /api/projects` path it records P1's exact effects before `201` only after the authenticated request organization/owner matches the admitted manifest roots. Anonymous, malformed, foreign-manifest, and root-mismatch paths cannot register effects. If B2 selected suppression, the observer admits the one fixture owner to skip post-commit provisioners only in the QA build; all other callers retain current behavior. Ordinary project files expose only the observer interface and nil/no-op default; no release path writes or queries QA authority.

**Static acceptance tests:** normal unregistered callers are unchanged; authenticated matching fixture owner registers one exact effect set; foreign manifest or organization/owner root mismatch registers none; malformed request registers none; anonymous request never reaches recorder; duplicate key records no second set; provisioner policy cannot fall back silently.

### P4 — consumer, projection, and exact-cleanup core

**Depends on:** P1c, P2, and P3.

**Owner files:** new `backend/internal/qaprojectoutbox/consumer.go`, `projection.go`, `cleanup.go`; focused core tests plus approved fake NATS/Ent interfaces.

**Deliverable:** one `qa && !release` durable pull-consumer lifecycle, 3-second bounded receipt/projection behavior, exact-ID database projections, and retained partial-cleanup recovery. Every QA authority writer/query is absent from release. It exposes a narrow composition interface, not an HTTP route. Reconnect validates and reuses the manifest consumer. Receipt cache precedes acknowledgement. The fixed pull batch, delivery ceiling, deadline, and unsafe outbox-worker handling follow this plan.

**Static acceptance tests:** consumer is created once before publication admission; reconnect cannot create/replace one; batch is 8, `MaxAckPending` is 32, and no more than 64 deliveries are examined; foreign delivery acknowledgement applies only to this consumer copy; ceiling is fixed redacted `500`, deadline is `408`, and neither becomes zero; each count rejects a foreign ID; cleanup cannot target stream/server/database/service; repeated cleanup uses only recorded IDs; a mid-cleanup failure retains the same manifest list.

### P5a — caller-supplied session-ID auth capability

**Depends on:** P0 independent acceptance (`ACCEPTED / CLEARED_FOR_P1A`).

**Owner files:** the smallest additive capability under `backend/internal/auth/` plus focused genuine isolated auth tests only. No project-outbox controller, route, server wiring, or QA signing implementation belongs to this atom.

**Deliverable:** an additive internal-auth capability accepts a caller-supplied non-zero session UUID and then executes the existing single-session replacement, Redis lifecycle, database session persistence, NATS/session event, refresh/access token issuance, expiry, and revocation semantics. Existing `CreateSession` remains unchanged and delegates/shared internals may be reused. Verification uses the existing production auth parser/verifier; QA code neither signs JWTs manually nor receives signing-secret access.

**Static acceptance tests:** zero/caller-mismatched UUID is rejected; the supplied UUID is the persisted session and verified JWT `sid`; existing single-session eviction and Redis/DB/NATS/token semantics remain equal; the standard auth parser validates owner and `sid`; existing `CreateSession` callers are unchanged; no manual signer or secret exposure exists.

### P5b — QA route and server composition over the core

**Depends on:** P4 and P5a.

**Owner files:** new `backend/internal/qaprojectoutbox/controller.go`, `fixture.go`, `http.go`, `session.go`; `backend/cmd/server/wiring_project_outbox_qa.go`; its `!qa || release` no-op stub; `backend/cmd/server/main.go`; and focused route/config/wiring tests.

**Deliverable:** the private route contract and every authority writer/query compile only under `qa && !release`. Fixture create records exact organization/owner/session and every created fixture row before use, allocates a non-zero session UUID, invokes P5a, and verifies the issued token through the existing auth parser so its owner and `sid` exactly match the recorded roots. `main.go` contains exactly one `wireProjectOutboxQA(d)` call after `wireModules(d)` and before `wireIntegrations(d)`. The QA implementation installs the routes/observer; the `!qa || release` stub is a no-op. Ordinary project files depend only on a nil/no-op observer interface and cannot write/query QA authority in release. P4 creates the durable consumer before `ready`.

**Static acceptance tests:** exact call order and exactly-one-call wiring are frozen; release stub exposes no control surface or authority access; missing/invalid config fails closed; no route uses `dev/login`; owner/`sid` mismatch or absent `sid` fails through the existing parser; fixture readiness is impossible if P4 consumer creation fails; all handler shape errors have fixed statuses; unrecorded effects fail closed; no QA manual signing or signing-secret access exists.

### P6 — HTTP-only Python delegate and plugin composition

**Depends on:** P5b route shape is frozen.

**Owner files:** `testframework/src/hub_testkit/harness/project_outbox.py`, `pytest_project_outbox.py`; new `project_outbox_http.py` and its domain/config helpers; `testframework/src/hub_testkit/integration/project_outbox/{api,builders,models,oracles}.py`; `testframework/src/hub_testkit/pytest_plugin/plugin.py`; corresponding framework selftests. Do **not** edit `testframework/tests/integration/redis-nats/project_outbox/conftest.py` unless the P6 selftest proves the public plugin hook cannot install the factory.

**Deliverable:** a lazy `ProjectOutboxHTTPFactory` installed in `pytest_configure` only after a valid `integration` `SuiteBinding` is known. Factory construction is per test; the delegate owns an `httpx` client and invokes only the private QA endpoints plus real `POST /api/projects`. `send`, `provision`, receipt/state, disconnect/reconnect, and cleanup retain the existing opaque runtime shape. `ProvisionedProjectOutbox` carries only redacted exact owner/session/consumer binding and product bearer credentials required by the existing builders.

**Static acceptance tests:** plugin performs no network/I/O at collection; factory is absent outside the integration suite and is installed once inside it; sealed suite conftest is byte/AST-equivalent; delegate has no `nats`, database, Docker, subprocess, or dynamic-import path; requests keep public product status expectations and never retry a lost product create; all secret-bearing representations/diagnostics are redacted.

### P7 — cases, Compose admission, documentation, and Docker-only evidence

**Depends on:** P1a, P1b, P1c, P2–P4, P5a, P5b, and P6 independently accepted; B1–B3 remain frozen constraints.

**Owner files:** five `testframework/cases/outbox/TC-OUTBOX-000*.md` records only if their declared private endpoint/status contract changes; `testframework/selftests/...` for traceability/plugin/harness contracts; `docker-compose.local.yml`; `testframework/{README.md,PACKAGE-CONTRACT.md,PIPELINE.md}`; applicable system/QA documentation; no sealed suite rewrite.

**Deliverable:** `qa-backend` and `python-integration` receive the dedicated project-outbox token; no new service, stream, database, network, or Python NATS/DB dependency is introduced. Documentation states both the exact invocation and the evidence boundary. `STATE` may be updated only after real accepted evidence, never by this source packet.

**Required Docker gates, in order:**

1. `docker compose -f docker-compose.local.yml run --rm backend-unit` for genuine isolated Go units introduced by P1–P5b.
2. `docker compose -f docker-compose.local.yml run --rm python-contract` for P6/P7 framework, plugin, traceability, and import-boundary selftests.
3. `docker compose -f docker-compose.local.yml run --rm python-integration` for exactly the sealed integration profile, including all five outbox cases. A failure is triaged against the frozen P0 contract before changing code/tests.

Each result is recorded separately. `python-contract` does not prove server behavior; a successful focused suite does not license a product/launcher claim beyond its exact execution evidence.

## Acceptance matrix

| Requirement | Packet/gate that proves it |
|---|---|
| immutable fixture roots and one exact manifest | P2 + P4 static review; P7 integration evidence if run |
| durable consumer created before publication and reused after reconnect | P4/P5b unit contracts; `TC-OUTBOX-0005` only after P7 Docker evidence |
| 3-second observation within 5-second HTTP read budget | P5b deadline unit contract and P6 transport contract |
| exact project/member/outbox IDs before any projection or cleanup | P1a/P1b/P1c + P3 transaction contracts |
| authoritative zero state | B1 authority plus P5b negative projection contract; `0003`/`0004` after P7 |
| no shared-resource ownership or deletion | P5b cleanup negative contracts; P7 Compose review |
| HTTP-only Python and sealed conftest | P6 plugin/harness selftests |
| lost-response safety | P2/P5b idempotent commands; B3 explicitly scopes public-create and process-restart limits |

## Explicit non-goals

- No production endpoint, release-build QA route, external infrastructure, exclusive stream/database, Python NATS/PostgreSQL client, broad payload/name/email scan, or cleanup-derived oracle is introduced.
- No claim is made about transaction atomicity beyond the narrow P1 recorder contract, PubAck ordering, worker recovery, whole-product project provisioning, server restart, or successful runtime execution.
- The project API does not gain invented idempotency behavior. If root wants retry-safe lost-response creation, that is a separate API feature with its own contract and tests.

## Root review checklist

- [x] P0 remediated frozen protocol is independently accepted and has Docker `python-contract` evidence: 280/280 PASS (seed `4057944629`, `5.39s`); this clears P1a admission only.
- [x] B1 freezes QA-only same-Ent-transaction persisted manifest/effect authority; payload scans are forbidden.
- [x] B2 freezes QA-build-only provisioner suppression solely for the admitted exact fixture owner.
- [x] B3 freezes same-process recovery; ambiguous/lost public create is never replayed and restart/token reissue are future scope.
- [ ] P1a/P1b/P1c, P2–P4, P5a/P5b, and P6–P7 are dispatched as the stated atomic packets with explicit Terra/xhigh implementation and Terra/high independent review provenance.
- [ ] Docker evidence, if later produced, is recorded as `PASS`, `FAIL`, `NOT_RUN`, or `BLOCKED` independently from static acceptance.

## Self-review

`C/I/M = 0/0/0` for this design artifact: root decisions B1–B3 are frozen; P1 is decomposed to P1a/P1b/P1c; P5 is decomposed to P5a auth capability and P5b QA route/wiring; consumer/projection/cleanup core precedes both; fixed JetStream bounds and narrow exact lifecycle ownership are explicit. The remediated P0 protocol is independently accepted and has Docker `python-contract` evidence (280/280 PASS, seed `4057944629`, `5.39s`); only P1a is cleared for atomic admission, and P1a–P7 remain `NOT_RUN`.
