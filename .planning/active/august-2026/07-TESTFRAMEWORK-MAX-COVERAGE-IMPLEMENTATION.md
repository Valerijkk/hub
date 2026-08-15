# Testframework Maximum-Coverage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: use `superpowers:subagent-driven-development` or `superpowers:executing-plans` task by task. Checkboxes are deliberately retained for execution tracking; this document itself grants source-authoring scope only.

**Goal:** Author the greenfield `testframework` package, closed test-case catalog, and maximum-risk suite source in ordered waves without claiming that a runner, image, dependency resolution, Docker execution, or product behavior is ready.

**Architecture:** One installable `hub_testkit` package owns reusable Python QA behavior. Repository-owned cases, schemas, policies, framework selftests, and SUT suites remain separate roots and join only through validated traceability metadata and the sole future launcher. The waves intentionally maximize coverage of implemented, observable Hub behavior while keeping product proposals and unverified infrastructure out of the executable source surface.

**Tech Stack:** CPython 3.12.13; `requires-python = ">=3.12,<3.13"`; Hatchling; uv lock generation in Docker only; pytest framework selftests; Pydantic, JSON Schema, HTTPX, OpenAPI validation, Playwright, and Locust in explicit dependency groups.

## Global Constraints

- This is the founder's 2026-08-15 override of `G129-C`, `G130-H`, and per-feature `G-ACTIVATE-*` **only for source authoring in F0--F3**. It does not create an execution receipt or alter `NOT_READY` / `NOT_RUN` runtime truth.
- Runtime is still prohibited until all of A1 runner admission, R Docker resolver admission, immutable image evidence, generated `uv.lock`, and Docker-only evidence are separately accepted on one exact revision. File existence, source review, or a source commit never supplies any of those proofs.
- Do not run host or Docker tests, linters, package builders, dependency resolution, services, browser automation, load, or network requests while authoring this plan. Future execution always uses the sole accepted launcher and task-owned Docker resources.
- The canonical new-case grammar is `^TC-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*-[0-9]{4}$`. `TC-HEALTH-0001` is valid. Older `TC-AUTH-03` examples and `HUB-TC-*` inventory identifiers are historical evidence only and are rejected for new `testframework/cases/**` files.
- Keep the exact 47 profile identifiers, image roles, limits, targets, and deadlines in Phase 130 unchanged. A registered profile is capacity metadata, not runtime activation or proof that a target suite exists.
- No generic `utils.py`, `BaseTest`, global mutable configuration, `sys.path` mutation, copied suite-local reusable helpers, direct `src/` imports, package-to-suite imports, unbounded retry/polling, raw secret diagnostics, or a second launcher is permitted.
- No test may assert an unimplemented product proposal. Magic-link TTL behavior is excluded pending a product decision. “Tenant” cases are limited to an actual implemented ownership/scope invariant; they must not invent multi-tenant behavior for a single-tenant deployment. Current read-only Calendar aggregation/export and implemented call API authz/signature behavior are in scope. Calendar-v2 writes/resources/availability/subscriptions, Wiki, WebRTC/media capacity, and other unimplemented call proposals remain planned-only.
- Every future source slice has one bounded feature/risk, literal owned paths, a schema-valid case or explicit framework-selftest rationale, an observable oracle, exact cleanup, a rollback-by-revert boundary, and independent review before the next slice.

## 1. Authority and status separation

This document supersedes the former missing-predecessor stops only as described above. It does not reopen Phase 129 deletion/cutover work, accept the private runner, or authorize edits outside the F0--F3 source plan.

| Surface | Source-authoring status | Execution/runtime status |
|---|---|---|
| F0 package, traceability, plugin, health, framework selftests | Authorized in the order below | `NOT_READY` / `NOT_RUN` until A1, R, image, lock, and Docker evidence |
| F1 HTTP, identity, authorization/platform, OpenAPI/security | Authorized after F0 source review | `NOT_READY` / `NOT_RUN` under the same gates |
| F2 integration/API spine | Authorized after F1 source review, only for implemented behavior | `NOT_READY` / `NOT_RUN` under the same gates |
| F3 browser, accessibility, privacy, load | Authorized after its source prerequisites, never as a product-readiness claim | `NOT_READY` / `NOT_RUN` under the same gates |
| Phase 129 cutover/deletion, CI publication, deployments | Not authorized by this document | Unchanged |

`A1` is the accepted public versioned launcher admission; `R` is the accepted Docker dependency-resolution admission. Both must record exact source revision, approved request/profile binding, image digest, resource limits, failure/cleanup behavior, and provenance. `uv.lock` is generated only by R's Docker-owned resolver from the committed `pyproject.toml`; it is never hand-authored and may remain absent before R.

## 2. Fixed physical contract and F0 bootstrap

### 2.1 Bootstrap files and responsibilities

| Path | Responsibility | First owning atomic slice |
|---|---|---|
| `testframework/pyproject.toml` | sole package metadata, Hatchling build backend, Python range, `src` discovery, pytest strict configuration, and dependency-group authority | F0-01 |
| `testframework/.python-version` | literal `3.12.13` only | F0-01 |
| `testframework/uv.lock` | generated exact resolution only after R; intentionally absent before R | F0-02/R |
| `testframework/PACKAGE-CONTRACT.md` | current source/package/runtime status updated in the same atomic slice that changes package artifacts | F0-01 and F0 closure |
| `testframework/src/hub_testkit/__init__.py` | sole public namespace boundary | F0-01 |
| `testframework/src/hub_testkit/py.typed` | PEP 561 marker shipped with typed public surfaces | F0-01 |
| `testframework/src/hub_testkit/config/{__init__.py,models.py,load.py}` | frozen strict settings and environment admission before SUT I/O | F0-03 |
| `testframework/src/hub_testkit/runtime/{__init__.py,ids.py,ownership.py,redaction.py}` | run identity, exact ownership ledger, cleanup scope, fail-closed redaction | F0-03 |
| `testframework/src/hub_testkit/traceability/{__init__.py,model.py,loader.py,index.py}` | closed case records, schema loading, case index, collected-test link validation | F0-04 |
| `testframework/src/hub_testkit/pytest_plugin/{__init__.py,plugin.py,markers.py,collection.py,reporting.py}` | public pytest-hook collection and reporting policy only | F0-05 |
| `testframework/src/hub_testkit/domains/health/{__init__.py,models.py,api.py,oracles.py}` | narrow health facade and observable health oracle | F0-06 |
| `testframework/schemas/test-case.schema.json` | closed machine-readable frontmatter/case contract | F0-04 |
| `testframework/policies/{markers.toml,environments.toml,redaction.toml,profiles.toml}` | closed marker, environment, redaction, and exact 47-profile policy data | F0-02/F0-04/F0-05 |
| `testframework/selftests/architecture/{package,config,policies}/**` | installation, metadata, configuration, profile, namespace, marker, and no-service contracts collected by `architecture-selftest` | F0-01 through F0-03 |
| `testframework/selftests/collection/{traceability,pytest_plugin}/**` | case and public-hook collection contracts collected by `collection-gates` | F0-04/F0-05 |
| `testframework/selftests/runtime/**` | ownership, cleanup, deadline, and redaction contracts collected by `runtime-observability` | F0-03 |
| `testframework/selftests/architecture/imports/**` | import-direction and forbidden-pattern contracts collected by `architecture-selftest` | F0-05 |
| `testframework/tests/api/health/**` and `testframework/cases/health/TC-HEALTH-0001.md` | first schema-valid API tracer for `GET /api/live` | F0-06 |

The package owns no generic product behavior. SUT tests only import installed `hub_testkit.*`; framework selftests may validate package contracts but never contact Hub, read credentials, create reports, or access network services.

### 2.2 `pyproject.toml` package groups

The committed project metadata contains one group authority with the following membership. Version resolution is deliberately deferred to R and becomes immutable only in generated `uv.lock`.

| Group | Exact packages | First source consumer |
|---|---|---|
| project runtime | `pydantic`, `jsonschema` | F0 config and traceability |
| `framework` | `pytest`, `pytest-timeout`, `pytest-randomly` | F0 selftests and plugin |
| `quality` | `ruff`, `mypy`, `import-linter` | F0 architecture/source contracts |
| `http` | `httpx` | F1 HTTP and health extraction |
| `contract` | `openapi-spec-validator` | F1 OpenAPI contract |
| `browser` | `playwright`, `pytest-playwright` | F3 browser/a11y only |
| `load` | `locust` | F3 load only |

Hatchling is the only build backend. `pytest-xdist`, Faker, Allure, Schemathesis, Pact, Testcontainers, code generators, and browser/load dependencies outside their named groups are not bootstrap dependencies. A package request that requires one of them must amend this plan and the group table before source authoring.

## 3. Traceability, cases, and plugin contract

Every new case is strict UTF-8 Markdown with exactly one leading TOML frontmatter block delimited by `+++` lines and parsed only by CPython 3.12 `tomllib`. YAML frontmatter and a second parser dependency are forbidden. The mapping is then validated by `test-case.schema.json`; bytes before the opening delimiter, an absent closing delimiter, duplicate TOML keys, invalid UTF-8, or an empty Markdown body fail closed.

The case model is closed: `id`, `title`, `status`, `domain`, `layer`, `risk`, `tasks`, `requirements`, `automation`, `owner`, `data_classification`, `preconditions`, `test_data`, `oracle`, and `cleanup`. Unknown fields fail. Scalar fields are non-empty strings; `tasks`, `requirements`, `preconditions`, `test_data`, `oracle`, and `cleanup` are non-empty duplicate-free string arrays. `status` is `draft`, `active`, or `deprecated`; `layer` is `contract`, `api`, `integration`, `e2e`, `load`, or `manual`; `risk` is `critical`, `high`, `medium`, or `low`; `data_classification` is `public`, `internal`, `confidential`, or `restricted`.

`automation` is a closed tagged object. Active `contract|api|integration|e2e` cases require exactly `{kind = "pytest", node_id = "<repo-relative pytest node id>"}`. Active `load` cases require exactly `{kind = "locust", entrypoint = "<repo-relative callable>"}`. `manual` cases require exactly `{kind = "manual", evidence_owner = "<owner>"}`. Cross-kind fields are forbidden. Each active pytest case has one unique node ID, and collection must prove a bijection between that node ID, the sole `case_id` marker, and the loaded active case. Missing active items, duplicate node IDs, and collected items not referenced by an active case fail.

Collection uses one `@pytest.mark.case_id("TC-...")` and exactly one registered layer marker per executable test. The plugin uses public pytest hooks only and must fail closed on malformed/new-grammar ID, duplicate case ID, missing or unknown marker, path/layer mismatch, missing risk/oracle/cleanup, orphan requirement/task link, inactive case collected, unexpected `skip`/`xfail`, or zero collection. It emits only redacted JUnit properties for case/task/requirement/domain/risk. It performs no Hub or network I/O during collection.

The 47 profile registry remains the literal list in `130-tf-traceability-gates/PLAN.md`: it is copied verbatim to `policies/profiles.toml` by F0-02 and receives a parser/duplicate/field-closure selftest. No runner code writes that file.

## 4. Ordered source-authoring waves

### F0 — package bootstrap, traceability, plugin, health, framework selftests

1. **F0-01 package contract.** Add the project metadata, Python file, namespace marker, and package selftests. Prove source-level package discovery, distribution metadata intent, typed marker inclusion intent, strict pytest configuration, no root-path hack, and no service/network/reports dependency.
2. **F0-02 profile and resolver boundary.** Add the literal 47-row profile policy plus selftest. Preserve the documented absent-before-R status for `uv.lock`; do not create or ignore a lock before R.
3. **F0-03 strict config/runtime safety.** Add immutable `HubSettings`/environment policy admission; add nonce-bearing resource IDs, ownership ledger, exact cleanup scope, and idempotent recursive redaction. Selftests cover unknown config fields, production-deny admission, cross-run ownership, secret canaries, and no broad deletion selector.
4. **F0-04 closed traceability.** Add the TOML-frontmatter case schema/model/loader/index and fixtures. Selftests prove strict UTF-8/`+++` parsing through stdlib `tomllib`, the four-digit grammar, historical-ID rejection, duplicate/orphan/node-ID rejection, and required task/requirement/risk/oracle/cleanup links.
5. **F0-05 plugin/architecture.** Add public-hook plugin behavior and architecture selftests. Reject forbidden import directions, `sys.path`, `utils.py`, `BaseTest`, package-to-suite imports, naked sleep, unbounded polling, and unredacted reporting.
6. **F0-06 health tracer.** Add `TC-HEALTH-0001` and its API test against the implemented `GET /api/live` contract: no authorization, `200`, JSON `status == "ok"`. `/api/ready` and `/api/health` remain separate readiness behavior, including possible `503/degraded`; they are not substituted for liveness and no nonexistent `version` field is asserted.

### F1 — HTTP, identity/authz/platform, OpenAPI/security API and contract surfaces

F1 begins only after F0 independent source review records that the package boundary and 47-profile policy are closed.

**Founder source-authoring receipt (2026-08-15).** The instruction to build the
test framework and maximize tests before any Docker execution activates F1-01
source authoring on the reviewed F0 closure. For F1-01 only, this receipt
supersedes the older Phase 131-A literal path list with the following exact
allowlist:

- `testframework/src/hub_testkit/http/{__init__.py,transport.py,auth.py,policy.py,errors.py,diagnostics.py}`;
- `testframework/selftests/http/{test_policy.py,test_diagnostics.py}`;
- `testframework/src/hub_testkit/domains/health/api.py`.

F1-01 creates no case, SUT-suite, report, lock, image, runner, or execution
evidence. `TC-HEALTH-0001` and its fake-only API test remain byte-unchanged.
This receipt grants no later F1 slice and does not change `NOT_READY` /
`NOT_RUN` runtime truth.

**Founder F1-02 source-authoring receipt (2026-08-15).** The same instruction
activates the identity/session source slice only after the independently
reviewed F1-01 boundary. It is split into two revertable commits with these
literal allowlists:

1. JSON request closure may modify only
   `testframework/src/hub_testkit/http/transport.py` and
   `testframework/selftests/http/test_policy.py`. It adds the sole optional
   `Content-Type: application/json` request contract; arbitrary headers and
   all existing retry/auth/redaction rules remain closed.
2. Identity source may create only
   `testframework/src/hub_testkit/domains/identity/{__init__.py,models.py,builders.py,api.py,oracles.py}`,
   `testframework/src/hub_testkit/harness/{__init__.py,identity.py,pytest_identity.py}`,
   `testframework/tests/api/identity/{conftest.py,test_identity_api.py}`, and
   `testframework/cases/identity/{TC-IDENTITY-0001.md,TC-IDENTITY-0002.md,TC-IDENTITY-0003.md,TC-IDENTITY-0004.md}`.

Commit 2 may be authored only after commit 1 has received an independent
source-review `ACCEPT`; otherwise stop after commit 1. Each commit has its own
revert-only rollback boundary.

The identity harness is an explicit launcher-injected protocol, never a fake
SUT. The config stash contains an `IdentityHarnessFactory`, not a reusable
harness instance: the function-scoped fixture creates one fresh harness and
owner ledger per test, then calls that harness's exact cleanup in `finally`.
Missing injection fails before a request. Future A1 admission for the
`identity-api` profile must bind target `testframework/tests/api/identity` to
the exact `--hub-case-root testframework/cases/identity` and reject any other
case root; loading the global case catalog for a profile-local collection is
forbidden. This receipt creates no concrete provisioner, environment lookup,
dev login, magic-link delivery, report, lock, image, runner, Docker execution,
or runtime evidence. It grants no F1-03 or later slice and preserves
`NOT_READY` / `NOT_RUN`.

**F1-03 source-authoring receipt (2026-08-15).** F1-03 is split into two
independently revertible source-only commits. `authz-tenant` is only the frozen
profile identifier; no test, case, model, diagnostic, or documentation may
claim a tenant boundary in the current single-instance product. Both profiles
use a launcher-injected factory stored in pytest config. A function-scoped
fixture calls that factory exactly once per test, owns a fresh nonce-bound
ledger, and calls that harness's exact cleanup in `finally`. Missing,
duplicate, invalid, or cleanup-failing injection raises a generic
`pytest.UsageError` with no secret-bearing preserved cause. There is no
environment discovery, default HTTP client, fake SUT response, dev login,
magic-link delivery, host execution, runner authority, or runtime evidence.

F1-03A owns exactly these paths:

```text
testframework/src/hub_testkit/domains/authz/__init__.py
testframework/src/hub_testkit/domains/authz/models.py
testframework/src/hub_testkit/domains/authz/builders.py
testframework/src/hub_testkit/domains/authz/api.py
testframework/src/hub_testkit/domains/authz/oracles.py
testframework/src/hub_testkit/harness/__init__.py
testframework/src/hub_testkit/harness/authz.py
testframework/src/hub_testkit/harness/pytest_authz.py
testframework/tests/api/authz/conftest.py
testframework/tests/api/authz/test_authz_api.py
testframework/cases/authz/TC-AUTHZ-0001.md
testframework/cases/authz/TC-AUTHZ-0002.md
testframework/cases/authz/TC-AUTHZ-0003.md
testframework/cases/authz/TC-AUTHZ-0004.md
```

The future `authz-tenant` A1 profile binds only
`testframework/tests/api/authz` to exact
`--hub-case-root testframework/cases/authz`; any global or different case root
is rejected before collection. The four active `TASK-208` cases map one-to-one
to four literal, non-parametrized pytest nodes:

- `TC-AUTHZ-0001` / `test_system_admin_read_matrix`: anonymous registration
  mode is `401`, an active member is `403`, and founder plus system-admin are
  `200` with the sole `mode` field in `open|invite_only`.
- `TC-AUTHZ-0002` /
  `test_pending_approval_and_rejection_are_admin_only`: a member receives
  `403` for pending-list, approve, and reject; an admin sees both owned pending
  identities, approves one and rejects the other with the exact current success
  messages, and neither remains in the later pending list. Setup provisioning
  is not magic-link behavior and is never used as the oracle.
- `TC-AUTHZ-0003` /
  `test_deactivation_revokes_credentials_and_preserves_founder`: an admin
  deactivates a fresh ordinary active identity with `200` and `success: true`;
  that identity's access receives `/api/auth/me` `403` and its refresh receives
  `/api/auth/refresh` `401`. Founder deactivation is `403` and the founder's
  `/api/auth/me` remains `200`. Localized message text is not an oracle.
- `TC-AUTHZ-0004` /
  `test_pending_admin_endpoint_does_not_enumerate_active_users`: approving a
  known active identity and an unknown UUID both produce `404` with the same
  generic error shape; neither response nor diagnostic contains owned email,
  ID, token, or nonce canaries.

`AuthzHarness` exposes only `send`, `provision_principal`, and `cleanup`.
Principal setup explicitly distinguishes `founder|admin|member` and
`active|pending`, permits zero credentials for pending identities, redacts all
credential/PII representations, and never emulates an API response.

F1-03B owns exactly these paths:

```text
testframework/src/hub_testkit/domains/platform_abuse/__init__.py
testframework/src/hub_testkit/domains/platform_abuse/models.py
testframework/src/hub_testkit/domains/platform_abuse/builders.py
testframework/src/hub_testkit/domains/platform_abuse/api.py
testframework/src/hub_testkit/domains/platform_abuse/oracles.py
testframework/src/hub_testkit/harness/__init__.py
testframework/src/hub_testkit/harness/platform_abuse.py
testframework/src/hub_testkit/harness/pytest_platform_abuse.py
testframework/tests/api/platform/conftest.py
testframework/tests/api/platform/test_platform_abuse.py
testframework/cases/platform/TC-PLATFORM-0001.md
testframework/cases/platform/TC-PLATFORM-0002.md
```

The future `abuse-platform` A1 profile binds only
`testframework/tests/api/platform` to exact
`--hub-case-root testframework/cases/platform`. Its two active `TASK-209`
cases map one-to-one to two literal, non-parametrized pytest nodes:

- `TC-PLATFORM-0001` /
  `test_magic_verify_has_bounded_production_rate_limit`: with explicit
  production-like `DevMode=false`, isolated Redis, controlled client identity,
  and one owned nonce email, four identical valid-shape unknown-code requests
  yield `401,401,401,429`; the fourth has exact `Retry-After: 600`. Raw
  response, headers, diagnostics, and failures omit email, code, and nonce.
- `TC-PLATFORM-0002` /
  `test_magic_verify_rejects_replay_without_secret_leak`: explicit future
  launcher authority provisions one owned active identity and one valid
  one-time code without delivery or dev login; first public verification is
  `200` with a strict token pair and byte-identical replay is `401`. Tokens,
  code, email, nonce, and authorization never enter repr, failure text, case
  markdown, or diagnostics.

The one-time-code provisioner, isolated Redis binding, and production-like
server remain future launcher/A1 responsibilities. The config stash holds a
`PlatformAbuseHarnessFactory`; each call must bind a healthy exclusive Redis
instance or database and one stable server-observed client identity before the
first request. Its nonce-bound cleanup ledger records and may delete only the
exact owned `rl:auth:ip:<server-derived-hash>`,
`rl:auth-verify-email:email:<normalized-owned-email>`,
`magiclink:fails:<normalized-owned-email>`, and owned one-time-code state.
Prefix scan/delete is forbidden. Missing, foreign, unhealthy, or non-exclusive
state fails before a request, and cleanup never removes an unrecorded key.
Rate-limit expiry/recovery is excluded: no accepted clock-control API exists,
sleeping ten minutes is forbidden, and cleanup of exact fixture-owned Redis
keys is not an observed product oracle.

`TC-AUTHZ-0005` is blocked on a separate product repair. The current project
route registration places protected handlers before `rbac.RequirePermission`,
and those handlers terminate the request without `Next`, so the RBAC middleware
is unreachable. F1-03 must not manufacture a passing cross-project denial.
After a separately reviewed backend fix, a follow-up may assert outsider member
read `403`, developer owner-escalation `403`, unchanged developer role, owner
read `200`, and canary-free denials. Organization invite success is also
excluded because the current response exposes an invite secret. F1-03A and
F1-03B create no profile binding, provisioner, image, lock, runner, Docker
execution, report, or runtime evidence and preserve `NOT_READY` / `NOT_RUN`.

1. **F1-01 HTTP policy and diagnostics.** Add immutable actor/client policy, explicit connect/read/write/pool timeout values, bounded retry classification, idempotency protection for unsafe methods, correlation, and redacted diagnostics. The first failures are mutable authorization, timeout/retry leakage, unsafe POST retry, and secret-bearing request evidence.
2. **F1-02 identity/session/account.** Add only observable current auth/session/account behavior: successful valid flow, invalid credentials/no token, deactivated/blocked account denial where implemented, refresh/session revocation behavior where current routes support it, and exact fixture cleanup. Magic-link transport/code creation is not assumed; a safe provisioning path must be explicit. `/api/dev/login` is development-only and is not a default auth fixture.
3. **F1-03 authorization/platform abuse.** Author only the six active cases in the source-authoring receipt above. Add admin/account RBAC denial, privilege preservation, enumeration-safe responses, one-time-code replay rejection, the observable production rate bound, and artifact redaction. Do not claim rate-limit expiry/recovery without clock control, and do not add cross-project denial until the separately reviewed middleware-order product fix exists. The reserved `authz-tenant` profile name is not evidence of a true multi-tenant product.
4. **F1-04 OpenAPI/security contracts.** Validate implemented public OpenAPI/schema/event compatibility and critical/high security traceability. Bound snapshots; reject incompatible schema/event changes, unredacted public artifacts, and a missing critical/high risk-to-case edge.

### F2 — implemented integration/API spine

F2 is source-authoring for active, observable implementations only. Each service gets isolated seed/nonce ownership and a source-level cleanup oracle before a SUT test is authored.

1. **Data and broker boundaries:** PostgreSQL/Mongo transaction/migration/isolation; Redis/NATS message bleed, duplicate job, timeout, reconnect, and recovery; storage partial failure/idempotency/recovery; WebSocket authorization/order/reconnect. Use only existing profile identities (`postgres-integration`, `redis-nats-integration`, `notifications-nats`, and the relevant API/work profile); do not invent a 48th profile.
2. **Implemented work domains:** tracker lifecycle, TMS state/run lifecycle, project scope/idempotency, messenger membership/order/privacy, notification delivery/retry, and implemented search/audit/backup/import-export boundaries. Every suite covers allowed behavior plus invalid, denied, boundary, recovery, and cleanup variants that are observable in current handlers.
3. **Implemented Calendar reads:** under the existing `calendar-api` profile add `cases/calendar/**`, an optional narrow `hub_testkit.domains.calendar` facade, and `tests/api/calendar/**` for authentication; missing/malformed/reversed/>366-day ranges; user/project-scoped call/sprint/deadline/approved-leave aggregation; UTC/order invariants; recurring scheduled-call reads; and `export.ics` content type, disposition, RFC 5545 escaping, and PII-safe leave titles.
4. **Implemented business/adapter domains:** CRM, HR, accounting, bases, intranet, files/storage, automation, GitHub, GitLab, mail, marketplace, AI, and adapters only where their current API/contract and safe fixture topology are accepted. Implemented call API authz and webhook-signature behavior may be covered under the existing authz/security profiles; no live media/provider egress. External delivery/integration tests use a controlled fake or contract boundary.
5. **Explicit exclusions:** Calendar-v2 writes/resources/availability/subscriptions, Wiki permission/tree/versioning, WebRTC/media capacity, unimplemented call proposals, and any route present only in a forward-looking specification remain planned-only until product scope, implementation, seed strategy, and an observable oracle are accepted.

### F3 — browser, accessibility, privacy, load

1. **Browser foundation and critical smoke:** lifecycle, immutable browser image contract, sanitized screenshots/traces, login/session only through the accepted fixture boundary, and one implemented critical journey.
2. **Journey/recovery:** work, business, admin, and recovery journeys only for F2-backed implemented features. No UI flow becomes a substitute for API/integration coverage.
3. **Accessibility/privacy:** scoped axe/WCAG interactions for accepted UI pages; PII/secret artifact inspection and authorization isolation across API/browser evidence.
4. **Load/resilience:** target guard first (non-production allowlist, capped users/rate/duration, task-owned cleanup), then steady/spike/soak source profiles. No measured baseline, capacity statement, or production-readiness conclusion is allowed before Docker execution evidence.

## 5. Maximum-risk suite catalog and source references

The catalog is risk-first, not a case-count target. It consolidates the live 47 profile registry and the detailed suite candidates in `docs/superpowers/plans/2026-08-08-python-qa-framework-rebase.md` §§10--15. That older document supplies candidate suite names only; the current implementation, API/docs, and this founder directive determine whether a case is authored.

| Catalog family | Required risk coverage | Profile/reference |
|---|---|---|
| Framework F0 | package installation boundary, config admission, ownership/cleanup, secret redaction, metadata closure, collection loss | `collection-gates`, `architecture-selftest`, `health-api`; detailed-plan §§10 and 17 |
| HTTP/identity F1 | mutable credentials, timeout/retry, token leakage, invalid/session/revocation behavior, role/object denial, replay/enumeration/rate recovery | `http-api`, `runtime-observability`, `identity-api`, `authz-tenant`, `abuse-platform`, `security-risk`; detailed-plan §§11--12 |
| Contract/security F1 | incompatible OpenAPI/schema/event, public-artifact redaction, critical/high risk graph | `openapi-contract`, `frontend-contract`, `privacy-security`; detailed-plan §§14 and 15 |
| Data/broker F2 | transaction/migration isolation, cross-run data leak, message bleed, duplicate/reordered delivery, timeout/recovery, exact cleanup | `postgres-integration`, `redis-nats-integration`, `notifications-nats`; detailed-plan §13 and §§14--15 |
| Work F2 | lifecycle/state conflict, scope/idempotency, membership/order/privacy, notification recovery | `tracker-api`, `tms-api`, `projects-api`, `messenger-api`, `work-risk`; live Phase 134 tasks A--G |
| Calendar F2 | authenticated scoped aggregation, range boundaries, UTC/order, recurrence reads, RFC 5545 export and PII-safe titles | `calendar-api`; current `GET /api/calendar/events` and `GET /api/calendar/export.ics` only |
| Business/adapters F2 | financial idempotency, HR privacy, row/schema ownership, storage recovery, webhook signature, unavailable/retry/redaction | `crm-api`, `hr-api`, `accounting-api`, `bases-api`, `intranet-api`, `files-api`, `automation-api`, `github-adapter`, `gitlab-adapter`, `mail-adapter`, `business-risk`; live Phase 135 and detailed-plan §14 |
| Browser/privacy F3 | browser leak, auth artifact leak, critical accessibility regression, PII/secret exposure, recovery journey | `browser-lifecycle`, `e2e-smoke`, `journey-*`, `accessibility`, `privacy-security`; live Phases 137--139 |
| Load F3 | production target, uncapped rate, baseline ambiguity, data leak, unsafe cleanup | `load-validity`, `load-measured`; live Phase 140 and detailed-plan §15 |

## 6. Atomic commit and review order

One atomic source commit contains its source, cases, and selftests for one row below; it never mixes domains, dependency/image changes, unrelated docs, or reports. The author writes the negative contract before the minimum implementation in the same working slice. Until runtime gates exist, an independent reviewer may accept source closure only; the commit must retain `NOT_READY` and make no pass/green claim.

| Order | Atomic commit subject | Literal source ownership | Mandatory independent review before next source commit |
|---:|---|---|---|
| 0 | `docs(testframework): activate founder max coverage source plan` | the four planning/contract files named in this document's change set | architecture + security review for scope/status/grammar |
| 1 | `build(testframework): bootstrap installable hub testkit` | F0-01 paths including the factual `PACKAGE-CONTRACT.md` status update | package/import-contract review |
| 2 | `feat(testframework): add profile and resolver policy boundary` | F0-02 paths, never `uv.lock` before R | profile/launcher-boundary review |
| 3 | `feat(testframework): add strict config runtime ownership` | F0-03 paths | security/redaction review |
| 4 | `feat(testframework): add closed test case traceability` | F0-04 paths | traceability/schema review |
| 5 | `feat(testframework): add collection and architecture gates` | F0-05 paths | pytest-public-hook and architecture review |
| 6 | `test(testframework): add health liveness tracer` | F0-06 paths | API-contract review |
| 7 | `feat(testframework): add HTTP policy boundary` | F1-01 paths | transport/auth/redaction review |
| 8 | `test(testframework): add identity session account cases` | F1-02 paths | auth/security/product-behavior review |
| 9A | `test(testframework): add admin account authorization cases` | exact F1-03A 14-path allowlist | security/product review; reject invented tenant semantics |
| 9B | `test(testframework): add platform abuse replay cases` | exact F1-03B 12-path allowlist | security/privacy review; provisioning and profile remain unbound |
| 10 | `test(testframework): add OpenAPI security contracts` | F1-04 paths | API/schema/security review |
| 11 | `test(testframework): add data broker isolation spine` | F2 data/broker paths, one service boundary per commit | integration/cleanup review per service |
| 12 | `test(testframework): add implemented work domain cases` | exactly one work domain per commit | domain + traceability review |
| 13 | `test(testframework): add current calendar read contracts` | `cases/calendar/**`, optional `src/hub_testkit/domains/calendar/**`, `tests/api/calendar/**` | calendar API/privacy/RFC 5545 review |
| 14 | `test(testframework): add implemented business adapter cases` | exactly one business/adapter boundary per commit | domain/security review; no live egress |
| 15 | `feat(testframework): add browser lifecycle and smoke source` | F3 browser foundation paths | browser artifact/privacy review |
| 16 | `test(testframework): add accepted journey and a11y privacy cases` | one journey or cross-cutting boundary per commit | product/a11y/security review |
| 17 | `test(testframework): add bounded load source profiles` | F3 load paths | production-safety and cleanup review |

At each review, reject: an unlisted path, a missing case/risk edge, an unobservable oracle, a shared mutable fixture, broad cleanup, a secret-bearing artifact, an old-format TC ID, a new profile, an unsupported dependency group, or a runtime-readiness claim. Rollback is a normal revert of that single commit, followed by a fresh source review; no legacy suite is copied back into the target tree.

## 7. Deferred resolver and execution admission

After source commits have independent review, the separate R admission may build a task-owned Docker resolver image, generate and review `uv.lock`, inspect resolved artifact provenance and hashes, and bind the lock to an immutable `images.lock.v3.json`/image digest. A1 then binds an accepted versioned launcher request to those immutable bytes and the exact 47 profiles. Only after both admissions may Docker-only framework selftests or SUT suites produce runtime evidence. The first runtime evidence records source SHA, lock hash, image digest, profile, CPU/RAM/PID limits, command, collection/pass/fail/skip, duration, artifact URI, and exact cleanup result.

Until then the valid wording is: **source-authoring planned or reviewed; package/launcher/runtime `NOT_READY` or `NOT_RUN`; no test result exists.**

## 8. Planning self-review

- F0 through F3 cover the requested package/bootstrap, API/contract, integration/API spine, and browser/a11y/privacy/load sequence.
- The exact new TC grammar accepts `TC-HEALTH-0001` and rejects the two historical formats.
- The 47-profile registry is retained as-is, with no profile addition or deadline change.
- The source override is narrowly limited; runtime gates, cutover recovery, Docker-only evidence, and product delivery remain separate.
- Magic TTL, unimplemented true tenancy, Calendar-v2 writes/resources/availability/subscriptions, Wiki, and WebRTC/media-capacity behavior are explicitly excluded; current Calendar reads/export remain in scope.
- Every source slice has a bounded atomic commit, independent review, and revert-only rollback boundary.
