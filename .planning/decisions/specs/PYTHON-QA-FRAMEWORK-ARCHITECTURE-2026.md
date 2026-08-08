# Python QA Framework Architecture 2026

**Status:** founder-approved design direction; implementation planning pending
**Decision date:** 2026-08-08
**Scope:** root `testframework/`, native unit-test policy, QA lifecycle and execution contract
**Supersedes after final founder review:** direct layer-only Python layout from the 2026-08-07 rebase draft

## 1. Source and decision

This design adapts the independent research report
`C:\Users\Valerii\Desktop\Hub_QA_Architecture_Report_2026-08-08.md`
(SHA-256 `28d5e9076527164338887d229f0acf90fcc20833852f0eb883e6a7463a1023c1`).
The report is evidence, not an executable plan. Hub keeps its founder-locked
greenfield cutover, Docker-only policy and protected `pro-qa-ai/**` boundary,
while replacing the provisional Python directory layout with a maintainable
hybrid domain-by-layer architecture.

The final choice is:

- genuine isolated units stay in `backend/tests/<domain>/**` and
  `frontend/tests/<domain>/**`;
- all new non-unit QA is Python in one installable root project;
- executable suites are layer-oriented; reusable code and cases are
  responsibility/domain-oriented;
- the physical tree grows only when a real consumer exists;
- legacy non-unit tests are not ported and historical parity is not a gate.

## 2. Goals and non-goals

### Goals

1. A new engineer can understand a test from business intent through evidence
   without reading hidden fixture graphs or framework internals.
2. Tests run identically in local and CI workflows through one Docker-only
   launcher contract.
3. Every stateful resource is isolated by run, worker and nonce and is cleaned
   without broad deletion.
4. Unit coverage is created with the feature; non-unit evidence is completed in
   the same feature train after product acceptance, not deferred indefinitely.
5. TASK, requirement, risk and stable TC traceability have one repository-owned
   source of truth.
6. Dependencies and abstractions are replaceable and justified by measured use.

### Non-goals

- no one-to-one legacy migration or row-count parity;
- no enterprise testing control plane or custom TMS replacement;
- no universal `BaseTest`, `BasePage`, service locator, singleton or `utils.py`;
- no global reruns, arbitrary sleeps, broad snapshots or ignored skips;
- no host Python, pytest, Playwright, Locust, lint, type-check or build process;
- no product storage or authorization migration as part of QA framework work.

## 3. Canonical architecture

```text
testframework/
├── pyproject.toml
├── uv.lock
├── .python-version
├── README.md
├── src/hub_testkit/
│   ├── config/                 # immutable validated settings; no I/O after composition
│   ├── runtime/                # ids, ownership, clocks, polling, redaction
│   ├── http/                   # HTTPX transport, auth boundary, timeout/retry policy
│   ├── contract/               # OpenAPI, JSON Schema, events, narrow snapshots
│   ├── browser/                # Playwright lifecycle and sanitized artifacts
│   ├── integration/            # PostgreSQL/Mongo/Redis/NATS/storage boundaries
│   ├── observability/          # correlation, structured events, run manifest
│   ├── traceability/           # TC schema, loader and graph validation
│   ├── pytest_plugin/          # only proven public-hook policy integration
│   └── domains/<domain>/
│       ├── models.py
│       ├── api.py
│       ├── builders.py
│       ├── oracles.py
│       └── ui/{pages,components}/
├── selftests/                  # tests of framework contracts, not SUT tests
├── tests/
│   ├── conftest.py             # plugin registration/composition only
│   ├── contract/<domain>/
│   ├── api/<domain>/
│   ├── integration/<boundary>/
│   └── e2e/<domain>/
├── load/                       # separate Locust lifecycle and measured client
├── cases/<domain>/TC-*.md
├── schemas/                    # test-case and run-contract schemas
├── policies/                   # environments, markers, redaction, profiles, licenses
├── docker/                     # images, Compose topology, lock and outer launcher
├── runner/                     # in-container validated entrypoint/profile loading
├── reports/                    # ignored task-owned run directories only
└── docs/                       # onboarding, authoring, environments and triage
```

This is the final namespace map, not the Phase 129 creation list. Empty
directories are not committed. The first runnable slice creates only the
packages needed by launcher validation, ownership, redaction, traceability and
one health-domain API test.

There is no generic `core/` directory. “Core” is a conceptual layer represented
by narrowly named packages such as `config`, `runtime`, `http` and
`traceability`. There is no root `pages/`, `steps/`, `fixtures/`, `data/` or
`assertions/` warehouse.

## 4. Dependency and responsibility rules

1. `config` and `runtime` import no domains, pytest, browser or network clients.
2. Boundary packages may depend on `config/runtime`, never on tests.
3. Domains are mutually independent. A cross-domain journey composes public
   facades in the test, not in a shared superclass.
4. Domain builders are immutable and pure; domain oracles perform no I/O and do
   not copy production algorithms.
5. Domain UI code contains locators and atomic interactions only. API setup,
   workflows, retries and business assertions stay visible in the test.
6. `pytest_plugin` uses public pytest hooks only and performs no SUT I/O during
   collection.
7. Tests do not import other test modules, private names or patch `sys.path`.
8. Locust uses its native measured client. It may reuse pure DTO/builders and
   auth signing, but not the HTTPX test client or pytest fixtures.
9. Generated clients, if a bounded spike accepts them, live behind handwritten
   domain facades and are never imported directly by tests.
10. Import Linter and architecture selftests enforce these rules. Any exception
    needs owner, rationale, expiry or explicit permanent ADR.

## 5. Progressive physical skeleton

Phase 129 creates the smallest runnable vertical slice:

```text
testframework/
├── pyproject.toml
├── uv.lock
├── .python-version
├── README.md
├── src/hub_testkit/
│   ├── config/
│   ├── runtime/{ids.py,ownership.py,redaction.py}
│   ├── traceability/{model.py,loader.py,index.py}
│   └── domains/health/{api.py,models.py,oracles.py}
├── selftests/{runner,traceability}/
├── tests/{conftest.py,api/health/test_health.py}
├── cases/health/TC-HEALTH-0001.md
├── schemas/{test-case.schema.json,run.schema.json}
├── policies/{markers.toml,environments.toml,redaction.toml}
├── docker/{Dockerfile.core,run-task.ps1,images.lock.json}
├── runner/entrypoint.py
└── reports/.gitignore
```

`testframework/docker/run-task.ps1` remains the only host entrypoint because Hub
forbids host Python execution. It performs Docker CLI orchestration only. The
Python runner executes inside the pinned container. Browser, integration, load,
Allure, xdist, generated clients and additional domains are not included until
their first accepted consumer/spike.

## 6. Execution and security contract

The versioned outer contract remains `hub.qa-runner/v1` until a deliberate
schema migration. Every run performs:

```text
validate signed/run spec
→ verify immutable image and lock
→ create task-owned resources and durable ownership manifest
→ capability health and collect-only policy gate
→ execute exact profile under inner and outer deadlines
→ preserve first causal terminal event and raw exits
→ redact and seal evidence
→ remove exact name/ID/label/nonce-owned resources
→ prove zero owned leftovers
```

Required controls:

- non-root runtime, read-only source/root filesystem and bounded tmpfs;
- explicit CPU, RAM, PID, file-descriptor, shm and wall-clock limits;
- pinned platform image digest, locked dependencies and no runtime downloads;
- no Docker socket, privileged mode, host network/PID/IPC or broad cleanup;
- task-owned Compose project, networks, volumes and report directory;
- environment allowlist/attestation and production deny before any test data;
- first-exit preservation: cleanup/rerun never overwrites the primary cause;
- zero collection, unexpected/expired skip, invalid evidence or cleanup failure
  is non-zero;
- no `docker system prune`, wildcard deletion or label-only destructive action.

Exact limits and worker counts are calibrated by Docker spikes; they are not
invented as permanent values in this design.

## 7. Feature testing lifecycle

Every feature follows one feature train with four gates:

1. **Implementation gate:** production code and genuine units are written
   together. Unit-eligible logic cannot be accepted with a “tests later” note.
2. **Product acceptance:** product/BA/SA and relevant verifier confirm intended
   behavior and update authoritative requirements/contracts.
3. **QA expansion:** QA designs and automates every applicable non-unit layer
   from requirements and risks: contract, API, integration, E2E, accessibility,
   security/privacy, recovery and load.
4. **Final verification:** independent verifier checks oracle sensitivity,
   traceability, deterministic/parallel/kill behavior, evidence and cleanup.

This sequencing prevents premature automation against an unstable feature but
also prevents a feature from being considered release-complete with units only.

## 8. Coverage policy

Coverage is a diagnostic and minimum safety floor, not the definition of
quality. The old 82.0373% denominator is discovery evidence only.

### Unit coverage for new or changed unit-eligible code

| Risk class | Line floor | Branch floor | Additional gate |
|---|---:|---:|---|
| authn/authz/session/tenant/security-critical policy | 95% | 90% | negative matrix + mutation/failing sentinel |
| business/domain logic and state machines | 90% | 85% | boundary/error/idempotency branches |
| pure UI logic, stores, hooks and reusable utilities | 85% | 80% | state-transition and failure behavior |

Generated code, declarative schemas, trivial wiring and unreachable defensive
platform branches are excluded only through a reviewed path-level rationale.
Thin adapters are not padded with meaningless unit mocks; their behavior is
proved at contract/integration boundaries.

These are changed-code floors. A package-wide percentage may be reported, but
cannot force unrelated legacy test-writing or preserve an obsolete denominator.
Critical branches must demonstrate failure sensitivity; reaching a line is not
enough.

### Non-unit coverage

There is no universal 50–60% line target for black-box tests. Acceptance is:

- all critical and high-risk requirements have automated observable evidence;
- applicable happy, negative, authn/authz, boundary, edge/corner,
  concurrency/time, dependency-failure, retry/recovery/idempotency and
  security/privacy scenarios are covered or carry an approved `N/A` rationale;
- critical user journeys have bounded E2E evidence;
- hot paths have valid measured load/SLO evidence;
- schema/contract and cleanup mutations make the suite fail as expected.

Medium/low risks may remain manual or planned only with owner and explicit
reason. Counts, assertions and file parity are not coverage substitutes.

## 9. Traceability and test-case contract

`cases/<domain>/TC-*.md` is the repository source of business intent. YAML front
matter is schema-validated and contains:

- stable `id`, title, status, domain, layer and risk;
- TASK and requirement references;
- automation state, owner and data classification.

The body contains intent, preconditions, observable scenario/oracles and cleanup
ownership, not selectors or duplicated request payloads. A Python test carries
only `@pytest.mark.case_id("TC-...")`; TASK/REQ/domain/risk are projected from
the case index into JUnit and optional Allure output.

Collection fails on duplicate TC, missing case, orphan automated case, wrong
path/layer, unknown marker, expired quarantine/skip or contradictory metadata.
Technical framework sentinels may use an explicit schema-defined exemption.

## 10. Marker and profile policy

The closed marker set begins with:

- layer: `contract`, `api`, `integration`, `e2e`;
- execution: `smoke`, `nightly`;
- quality: `security`, `authz`, `a11y`, `resilience`;
- capability: `postgres`, `mongo`, `redis`, `nats`, `storage`, `browser`;
- controlled exceptions: `serial(reason)`, `destructive`, `external`,
  `quarantine(issue,expires)`, `allow_skip(issue,expires,reason)`.

`full` is a launcher profile, not a marker. `load` is a Locust profile, not a
pytest layer. Domain, owner, priority and risk live in case metadata. Ad hoc
`skip`, non-strict xfail, `flaky`, `slow` and arbitrary team/domain markers are
forbidden.

## 11. Tool decisions

### Adopt after Docker compatibility acceptance

- CPython 3.14 candidate with time-boxed 3.13 fallback;
- pinned uv and committed `uv.lock`;
- pytest strict config with plugin autoload disabled;
- Ruff, mypy and Import Linter;
- HTTPX behind a thin Hub adapter;
- raw JSON Schema/OpenAPI validation before strict Pydantic models;
- pytest-timeout plus outer launcher watchdog;
- pytest-randomly with persisted seed;
- task-owned Docker Compose;
- Playwright Python with matching package/browser/image versions;
- Locust in a separate dependency group/image;
- frozen builders and seeded Faker;
- JUnit plus sealed run/traceability manifest as portable core output.

### Trial only through a bounded spike

- xdist with explicit workers, never `-n auto`;
- Allure as a replaceable Hub TMS projection, not source of truth;
- Schemathesis, generated OpenAPI client, Syrupy and axe integration;
- pytest-asyncio only for real async tests;
- Polyfactory, Pact and Testcontainers only with demonstrated need and a
  removal path.

### Avoid

- global reruns and green-by-retry;
- root god `conftest.py`, mutable globals and hidden assertions;
- Screenplay/business DSL before repeated workflows;
- giant all-tools image and runtime dependency installation;
- shared staging data, fixed container names/ports and broad cleanup;
- snapshots/generated clients as primary semantic oracles.

## 12. Revised phase intent

- **128:** accept only the descriptor-safe cutover authority and guard.
- **129:** execute exact native-unit/legacy cutover and create the minimal
  launcher + health vertical slice above.
- **130:** establish case schemas, traceability, marker/profile and architecture
  gates.
- **131:** typed runtime/HTTP boundaries and health/auth/account API pilot.
- **132:** security, authn/authz, cross-instance/project isolation and abuse.
- **133:** one production-equivalent boundary at a time for PG/Mongo/Redis/NATS.
- **134–136:** work/realtime, business/integrations and contracts by risk.
- **137:** Playwright lifecycle, component-first UI foundation and one smoke.
- **138:** three high-risk domain/recovery journeys.
- **139:** accessibility and cross-cutting security/privacy.
- **140:** valid Locust measurement, target guard and resilience profiles.
- **141:** calibrated CI profiles, supply chain, retention, onboarding,
  clean-room verification and handoff.

Existing Phase 129 operation IDs, exact cutover batches and source manifest are
preserved. Only Python target paths, scaffold tasks and later phase contracts are
replanned.

## 13. Product architecture boundary

The report's storage and external authorization research is recorded as future
decision evidence, not accepted product migration:

- SeaweedFS remains Hub's current shipped S3-compatible store;
- a Garage/SeaweedFS capability and recovery bake-off may be planned only when
  product storage work is explicitly reopened;
- typed deny-by-default Go authorization and SQL predicates remain current;
- SpiceDB/OpenFGA/Cerbos are Trial candidates only after measured graph,
  revocation/list or independent-service triggers;
- `pro-qa-ai/**` remains untouched without a separate founder handoff.

## 14. Definition of done for a framework backlog item

- TASK/REQ/TC link or documented technical-sentinel exemption;
- positive and negative/failure-sensitivity evidence;
- deterministic seed/time/namespace and exact cleanup;
- parallel/order/kill behavior checked proportional to risk;
- no forbidden import, marker, secret, skip or mutable image;
- Docker limits, timeout and target policy declared;
- diagnostic message contains operation, safe context and next action;
- dependency/abstraction has owner, evidence, upgrade test and removal path;
- docs/example updated when newcomer contract changes;
- report records image, lock, config, seed, workers and owned resources.

## 15. Required spikes before standardization

1. Python 3.14/3.13 image and plugin compatibility matrix.
2. xdist isolation, deterministic replay, speedup and memory calibration.
3. exact cleanup kill matrix for each service boundary.
4. Playwright non-root sandbox, private shm and redacted artifact profile.
5. Locust production-target guard, worker-loss and generator-validity checks.
6. optional tool trials only when their first real consumer appears.

No spike is allowed to weaken Docker-only execution, immutable provenance,
resource limits, exact cleanup, first-exit preservation or repository-owned
traceability.
