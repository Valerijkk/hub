# Hub Python QA Framework Rebase Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

> **Meta-plan bootstrap:** первый executable task — `G-CANON`, синхронизация exact 74-path active routing/canonical/evaluation surface; второй — `G-META`, синхронизация exact 23-path downstream-plan/checker surface; третий — `G128-R`, reseal Phase 128 authority. Ни один guard/code/cutover/runtime task не готов, пока все три predecessor gates не приняты independent reviewers; executable rechecks remain `NOT_READY` until G129-L.

> **Founder execution stop:** текущий implementation run заканчивается gate `G130-H` после safe rebase/cutover, accepted launcher/profile contracts, framework selftests и одного health vertical slice. Phases 131–141 ниже — полная future backlog map, а не bulk execution authorization: каждая остаётся `NOT_READY` до отдельного founder activation receipt с принятой новой feature/risk/requirement scope.

**Goal:** Безопасно выполнить founder-approved greenfield cutover и построить переносимый корневой Python QA-проект Hub с единым Docker-only execution contract, repository-owned traceability и risk-based evidence без legacy port/parity.

**Architecture:** Корень репозитория содержит `testframework/`, а единственный installable Python package живёт в `testframework/src/hub_testkit/`; перенос каталога `testframework/` в отдельный repository не требует изменения Python imports или внутренних путей. Executable suites организованы по слою в `testframework/tests/**` и `testframework/load/**`, reusable code — по ответственности и домену в `testframework/src/hub_testkit/**`, business intent — в `testframework/cases/**`. Физическое дерево растёт только с первым реальным consumer; первая production-quality tracer slice заканчивается работающим health API case через accepted `hub.qa-runner/v1`.

**Tech Stack:** CPython `3.14.*` candidate с time-boxed `3.13.*` fallback только после Docker matrix; pinned `uv` + committed `testframework/uv.lock`; pytest strict mode; Pydantic strict/frozen boundaries; HTTPX; JSON Schema Draft 2020-12; Ruff, mypy, Import Linter; PowerShell outer launcher; non-root digest-pinned Docker images; JUnit + sealed JSON evidence. Playwright Python, Locust, xdist, Allure и optional generators подключаются только в назначенной фазе после bounded acceptance.

## Global Constraints

- Source precedence: прямые founder decisions → `.planning/decisions/specs/PYTHON-QA-FRAMEWORK-ARCHITECTURE-2026.md` → `AGENTS.md`/`REGULATION.md`/safety rules → sealed Phase 128/129 authority → research report → historical docs.
- Allowed implementation roots: `testframework/**`, manifest-authorized `backend/tests/**`, `frontend/tests/**`, exact Phase 129 consumer paths, и docs из task-specific allowlist. Этот planning run меняет только данный файл.
- `pro-qa-ai/**`, unrelated dirty `frontend/src/**`, `.planning/debug/test-baseline-failures.md` и чужие reports/resources не читаются и не изменяются.
- Windows host разрешён только для чтения/редактирования, Git и exact task-contract Docker CLI entrypoint call; каждая receipt получает отдельный unique invocation, а повтор без нового task authorization запрещён. Python, pytest, Ruff, mypy, Import Linter, Playwright, Locust, Go/Vitest tests, validators и builds на host запрещены.
- До independent acceptance `testframework/docker/run-task.ps1` любой target runtime verdict — `NOT_READY`; прямые `docker run`, `docker compose`, legacy `hubtest`, Vitest/TS Playwright runner или host fallback не разрешены.
- Единственное bootstrap-исключение до acceptance launcher — один исторически обязательный read-only guard run через `.planning/audits/test-rebase/validate-contract.ps1` только после G128-WPS51 и повторно принятых G128-E/G128-S. G128-WPS51 ничего не устанавливает: он фиксирует уже встроенный Windows PowerShell 5.1 по literal `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`, SHA256, held-handle file identity, версиям и exact Microsoft publisher/production-chain/code-signing identity, затем принимает точный PS5.1 compatibility-remediation contract. G128-D запускается только этим абсолютным путём; unqualified `powershell.exe`, любой `pwsh`, PATH/PATHEXT discovery и fallback запрещены. Guard не является `hub.qa-runner/v1`, не запускает SUT tests, не применяет cutover, не публикует authority и запрещён после G129-L.
- `hub.qa-runner/v1` обязан enforce immutable image/platform digest, locked dependencies, no runtime downloads, non-root/read-only root, bounded tmpfs/shm/CPU/RAM/swap/PID/nofile, inner+outer deadlines, first-causal-exit preservation, task-owned reports/resources и exact cleanup proof.
- Phase 129 immutable arithmetic: ровно `176` explicit tasks, из них `173` unique batch IDs; ровно `1,120` operations; максимальный literal batch — `56`. Идентификаторы, cohorts, order и operation identity не реконструируются и не переименовываются.
- Cutover сохраняет только доказанные genuine native units: `backend/tests/<domain>/**` и `frontend/tests/<domain>/**`; legacy Go component/API/integration, static-contract Vitest, TypeScript Playwright и closure удаляются по sealed manifest и не портируются.
- Historical `82.0373%` Go coverage, file/case/assertion counts, dual-run и row parity не являются migration gate.
- Changed-code unit floors: security-critical `95% line / 90% branch`; business/domain `90% / 85%`; pure UI `85% / 80%`. Non-unit acceptance — requirement/risk/scenario/oracle evidence, а не общий line percentage.
- Current execution scope is only exact cutover/scaffold, launcher/central-profile/framework contracts and one health vertical slice through `G130-H`. No Phase 131–141 bulk expansion, universal coverage sprint or profile-table existence may be interpreted as authorization; later work activates one accepted feature/risk slice at a time.
- Не создавать пустые future directories, generic `core/`, `utils.py`, `BaseTest`, service locator, singleton, global reruns, fixed sleeps, hidden assertions, broad snapshots или shared mutable test data.
- Каждый code-producing task начинает с чувствительного failing sentinel, затем реализует минимальный production-quality contract, выполняет targeted и adjacent regression через accepted launcher, синхронизирует docs и заканчивается одним allowlisted atomic commit.
- Любой FAIL сначала проходит test triage `ROOT=code|test|infra|flaky|stop`; assertion/code не меняются ради зелёного цвета без сверки с authoritative behavior.
- Commit subject: Conventional Commit с English `type(scope)` и русским описанием; body содержит TASK path/requirement; AI attribution trailers запрещены. Push/PR/merge/deploy не входят в этот план.
- Requested runtime: `gpt-5.6-sol`, reasoning `xhigh`; observed `actual_model: UNKNOWN`, потому что runtime telemetry не предоставлена. Workflow provenance: `superpowers:receiving-code-review@6.2.0` → GSD plan revision → `superpowers:writing-plans@6.2.0`; plugins/MCP: none.

---

## 1. Program Outcome, Non-goals и непрерывные инварианты

После Phase 141 новый инженер из clean checkout выполняет один documented launcher call и получает non-zero framework/health evidence с точным `TASK → requirement → TC → nodeid → outcome`, immutable provenance, redaction и zero owned leftovers. Перенос всего `testframework/` в отдельный repository сохраняет imports `hub_testkit.*`, runner paths relative to project root и case/schema/policy references.

Не входят: production storage/authz migration; изменение `pro-qa-ai/**`; custom TMS/control plane; перенос legacy test bodies; восстановление старого coverage denominator; speculative browser/integration/load directories; push/merge/deploy; выполнение команд в рамках создания этого документа.

Непрерывные инварианты:

1. `config`/`runtime` не импортируют domains, pytest, browser или network clients.
2. Boundary packages зависят от `config/runtime`, но не от tests.
3. Domains независимы; cross-domain journey композирует public facades в test module.
4. Builders immutable/pure; oracles pure/no I/O и не копируют production algorithms.
5. UI objects содержат locators и atomic interactions; setup, workflows, polling и business assertions видимы в tests.
6. `pytest_plugin` использует только public pytest hooks и не обращается к SUT во время collection.
7. Tests не импортируют другие tests/private names и не изменяют `sys.path`.
8. Locust использует native measured client; reuse ограничен pure DTO/builders/signing.
9. Generated code, если spike принят, скрыт за handwritten domain facade и не импортируется tests напрямую.
10. Любое исключение architecture gate имеет owner, rationale и expiry либо отдельный permanent ADR.

## 2. Current-to-target map

| Current authority/path | Target state | Immutable rule / owner |
|---|---|---|
| `backend/testframework/**/*_test.go` + `backend/cmd/server/**/*_test.go` | только manifest-proven units → `backend/tests/<domain>/**`; component/API/integration удалены | Phase 129 batches; operation identity из sealed manifest |
| `testframework/stages/unit/frontend-test/**` | 156 units → `frontend/tests/<domain>/**`; `render.tsx` → `frontend/tests/support/render.tsx`; 24 static contracts + 2 compile fixtures удалены | Phase 129 batches; no Python translation |
| `testframework/stages/e2e/**` + TS support/config | 20 specs + 14 support files удалены | Python Playwright создаётся с нуля в Phases 137–139 |
| `testframework/{AUDIT.md,MIGRATION.md,PIPELINE.md,README.md,TESTING_GUIDE.md,guides/**,pipeline/**,stages/**}` | exact manifest consumers либо удаляются, либо `testframework/README.md` переписывается как 10-minute onboarding | Каждый affected consumer меняется в том же batch commit |
| `testframework/` без Python package | `testframework/src/hub_testkit/**` installable package + `testframework/{selftests,tests,cases,schemas,policies,docker,runner,reports}/**` | Phase 129 tracer; no empty future dirs |
| Current Phase 129 scaffold paths `testframework/core/**`, root `testframework/{api,integration,contract,e2e,load}/**`, `testframework/requirements.lock`, `testframework/pytest.ini`, scaffold `.gitkeep` | заменяются на approved `testframework/src/hub_testkit/**`, `testframework/uv.lock`, `testframework/pyproject.toml` pytest config, `testframework/reports/.gitignore` и progressive slice | Меняются только scaffold target/file lists; 173 batch IDs/operations не меняются |
| `.planning/audits/test-rebase/{inventory.json,founder-greenfield-cutover-manifest.json,test-architecture-contract.json}` | read-only sealed authorities | Phase 128 owner; never inferred from live paths during a batch |
| `.planning/audits/test-rebase/cutover-manifest.json` (currently absent) | atomically published exact planned authority первым принятым launcher-operation после G129-L и до Phase 129 batch execution | accepted `hub.qa-runner/v1` materializes builder-identical Phase 128 bytes; no operation reconstruction |
| `pro-qa-ai/**` | unchanged and unread | founder-only handoff boundary |

### Sealed Phase 129 arithmetic

| Plan | Tasks | Operations | Cohort |
|---|---:|---:|---|
| `129-01` | 3 | 0 | launcher/scaffold + two independent preflight gates |
| `129-02` | 32 | 350 | hook-free Go unit moves |
| `129-03` | 13 | 157 | 156 frontend unit moves + one render support move |
| `129-04` | 34 | 241 | 237 Go non-unit deletes + four backend support deletes |
| `129-05` | 24 | 84 | static/compile/TS E2E/support/config/doc deletes |
| `129-06` | 31 | 245 | hook-dependent Go unit release/refactor/moves |
| `129-07` | 39 | 43 | zero-consumer production hook-file deletes |
| **Total** | **176** | **1,120** | **173 unique batch IDs; max literal operations = 56** |

## 3. Target physical tree and file responsibility map

### Phase 129 progressive runnable slice

```text
testframework/
├── pyproject.toml
├── uv.lock
├── .python-version
├── README.md
├── src/hub_testkit/
│   ├── __init__.py
│   ├── config/{__init__.py,models.py,load.py}
│   ├── runtime/{__init__.py,ids.py,ownership.py,redaction.py}
│   ├── traceability/{__init__.py,model.py,loader.py,index.py}
│   └── domains/health/{__init__.py,api.py,models.py,oracles.py}
├── selftests/
│   ├── runner/test_launcher_contract.py
│   ├── runner/test_launcher_cleanup.py
│   ├── runtime/test_ids.py
│   ├── runtime/test_ownership.py
│   ├── runtime/test_redaction.py
│   └── traceability/{test_loader.py,test_index.py}
├── tests/
│   ├── conftest.py
│   └── api/health/test_health.py
├── cases/health/TC-HEALTH-0001.md
├── schemas/{test-case.schema.json,run.schema.json}
├── policies/{markers.toml,environments.toml,redaction.toml}
├── docker/{Dockerfile.core,run-task.ps1,images.lock.json}
├── runner/entrypoint.py
└── reports/.gitignore
```

`testframework/tests/conftest.py` выполняет только explicit composition/fixture registration. Full `testframework/src/hub_testkit/{pytest_plugin,http,observability}/**`, integration/browser/load adapters, Allure и xdist не создаются до первого принятого consumer.

### Long-lived responsibilities

| Path | Единственная ответственность | First owner phase |
|---|---|---:|
| `testframework/pyproject.toml` | project metadata, dependency groups, pytest/Ruff/mypy/import-linter config | 129 |
| `testframework/uv.lock` | exact dependency resolution authority | 129 |
| `testframework/src/hub_testkit/config/**` | immutable validated settings и policy composition | 129 |
| `testframework/src/hub_testkit/runtime/**` | run/worker/nonce IDs, ownership, bounded clocks/polling, redaction | 129/131 |
| `testframework/src/hub_testkit/traceability/**` | TC schema model, Markdown loader, unique/orphan graph | 129/130 |
| `testframework/src/hub_testkit/pytest_plugin/**` | public-hook marker/collection/outcome/report policy | 130 |
| `testframework/src/hub_testkit/http/**` | HTTPX construction, immutable actor/auth, timeout/idempotency/retry and diagnostics | 131 |
| `testframework/src/hub_testkit/domains/<domain>/**` | domain models/facades/builders/oracles/UI abstractions | first domain consumer |
| `testframework/src/hub_testkit/integration/**` | exact DB/broker/storage ownership boundaries | 133 |
| `testframework/src/hub_testkit/contract/**` | OpenAPI/JSON Schema/event/narrow snapshot validation | 136 |
| `testframework/src/hub_testkit/browser/**` | Playwright lifecycle, auth state, sanitized artifacts | 137 |
| `testframework/src/hub_testkit/observability/**` | correlation, structured events, sealed run manifest | 131/141 |
| `testframework/selftests/**` | framework contracts only, never SUT behavior | 129 onward |
| `testframework/tests/{contract,api,integration,e2e}/**` | executable pytest suites by layer/domain | first layer consumer |
| `testframework/load/**` | separate Locust lifecycle, users and profiles | 140 |
| `testframework/cases/<domain>/TC-*.md` | repository-owned business intent and oracle | first case consumer |
| `testframework/schemas/**` | case/run contracts | 129/130 |
| `testframework/policies/**` | environments, markers, redaction, profiles/licenses when adopted | 129/130/141 |
| `testframework/docker/**` | pinned images/topologies and sole outer launcher | 129 onward |
| `testframework/runner/**` | in-container validated profile execution | 129/130 |
| `testframework/reports/<task-run>/**` | ignored owner-marked per-run evidence only | 129 onward |

### Research candidate lock (not executable authority until Task 129-01-A)

| Capability | Candidate from 2026-08-08 report | Adoption gate |
|---|---|---|
| Python | CPython `3.14.7`; fallback `3.13.15` | Docker plugin/wheel matrix; choose one |
| Resolver | `uv 0.12.3` | pin uv, `uv lock --check`, locked sync during image build |
| Test/runtime | `pytest 9.1.1`, `pytest-timeout 2.4.0`, `pytest-randomly 4.1.0` | package legitimacy + framework selftests |
| Parallel | `pytest-xdist 3.8.0` | Trial: deterministic replay, no collision, measured speed/RAM |
| HTTP/models/schema/data | `httpx 0.28.1`, `pydantic 2.13.4`, `jsonschema 4.26.0`, `Faker 40.36.0` | core image compatibility + strict boundary sentinels |
| Quality | `Ruff 0.16.2`, `mypy 2.3.0`, `Import Linter 2.13` | core image + architecture selftests |
| Contract | `openapi-spec-validator 0.9.0` | representative Hub OpenAPI 3.1.2 golden specs |
| Browser | `playwright 1.62.0`, `pytest-playwright 0.8.0` | matching package/browser/image, non-root sandbox/shm spike |
| Load | `Locust 2.46.3` | separate image, target guard and measurement-validity spike |
| Optional | Schemathesis, Syrupy, Allure, generated client, Polyfactory, Pact, Testcontainers | bounded first-consumer Trial with owner/removal path; absent from early lock |
| Supply-chain | pinned `pip-audit`/Trivy scanner candidate | independent scanner provenance/digest review; mutable action/tag forbidden |

Versions in this table are research inputs, not permission to download or install. `testframework/uv.lock`, `testframework/docker/images.lock.json` and accepted package legitimacy evidence become executable authority.

## 4. First-slice interface contracts

Следующие names/signatures фиксируются до implementation и используются downstream без переименования; внутренние реализации остаются task-local.

```python
# hub_testkit.config.models
class HubSettings(BaseModel):
    environment_id: str
    base_url: AnyHttpUrl
    tenant_prefix: str
    artifact_level: Literal["failure", "trace", "debug"]

class EnvironmentPolicy(BaseModel):
    environments: tuple[AllowedEnvironment, ...]
    production_deny: tuple[AnyHttpUrl, ...]

# hub_testkit.config.load
def load_settings(raw: Mapping[str, object], policy: EnvironmentPolicy) -> HubSettings: ...
```

`HubSettings`/`EnvironmentPolicy` используют `ConfigDict(extra="forbid", frozen=True, strict=True)`; `load_settings` выполняет нормализацию и positive allowlist/production deny до любого SUT I/O.

```python
# hub_testkit.runtime.ids
@dataclass(frozen=True)
class RuntimeIds:
    run_id: str
    task_run_id: str
    worker_id: str
    nonce: str

def parse_runtime_ids(raw: Mapping[str, str]) -> RuntimeIds: ...
def resource_name(kind: str, ids: RuntimeIds, *, max_length: int) -> str: ...

# hub_testkit.runtime.ownership
@dataclass(frozen=True)
class OwnedResource:
    kind: str
    resource_id: str
    name: str
    nonce: str
    labels: Mapping[str, str]

class OwnershipLedger:
    def register(self, resource: OwnedResource) -> None: ...
    def snapshot(self) -> tuple[OwnedResource, ...]: ...
    def assert_cleanup_scope(self, resource: OwnedResource) -> None: ...

# hub_testkit.runtime.redaction
class Redactor:
    def redact_text(self, value: str) -> str: ...
    def redact_mapping(self, value: Mapping[str, object]) -> dict[str, object]: ...
```

`OwnershipLedger` записывает durable manifest сразу после resource creation и разрешает deletion только при совпадении exact ID + name + nonce + mandatory labels. `Redactor` идемпотентен и fail-closed для неизвестных sensitive fields.

```python
# hub_testkit.traceability.model
class TestCaseRecord(BaseModel):
    id: str
    title: str
    status: Literal["draft", "active", "deprecated"]
    domain: str
    layer: Literal["contract", "api", "integration", "e2e", "load", "manual"]
    risk: Literal["critical", "high", "medium", "low"]
    tasks: tuple[str, ...]
    requirements: tuple[str, ...]
    automation: AutomationState
    owner: str
    data_classification: Literal["synthetic", "sanitized", "sensitive-approved"]

# hub_testkit.traceability.loader / index
def load_case(path: Path, *, schema_path: Path) -> TestCaseRecord: ...
def build_case_index(cases: Iterable[TestCaseRecord]) -> CaseIndex: ...
def validate_case_links(index: CaseIndex, collected: Iterable[CollectedCase]) -> tuple[Violation, ...]: ...
```

```python
# hub_testkit.domains.health.models / api / oracles
class HealthSnapshot(BaseModel):
    status: Literal["ok", "degraded", "unavailable"]
    version: str | None

class HealthApi:
    def __init__(self, client: httpx.Client) -> None: ...
    def fetch(self) -> HealthSnapshot: ...

def assert_health_ready(snapshot: HealthSnapshot) -> None: ...
```

`HealthApi` — узкий first consumer HTTPX без generic transport abstraction; Phase 131 извлекает принятый `http/**` boundary и сохраняет public `HealthApi.fetch()`.

```python
# runner.entrypoint
class RunEnvelope(BaseModel):
    schema: Literal["hub.qa-runner/v1"]
    run: RunIdentity
    source: SourceIdentity
    image: ImageIdentity
    target: TargetIdentity
    limits: ResourceLimits
    execution: ExecutionRequest
    reports: ReportBoundary

def execute(envelope: RunEnvelope) -> int: ...
def main(argv: Sequence[str] | None = None) -> int: ...
```

Outer `run-task.ps1` поддерживает параметры `-ContractVersion`, `-PlanId`, `-TaskRunId`, `-TaskId`, `-OwnerRole`, `-VerifierRoles`, `-Mode`, `-Profile`, `-ImageLock`, `-InnerTimeoutSeconds`, `-TimeoutSeconds` и cutover-only `-Manifest`, `-BatchId`, `-ExpectedOperations`. Он строит schema-valid envelope, а не передаёт command strings из repo JSON.

## 5. Bootstrap active-canon/planning sync and readiness state machine

### Task G-CANON: Synchronize the exact active routing and ownership canon

**Files:** Modify, delete or create only this exact 74-path active allowlist:

**Authority and rules (9):**

- `AGENTS.md`
- `.cursor/REGULATION.md`
- `.cursor/GSD-INTEGRATION.md`
- `.cursor/rules/00-project.mdc`
- `.cursor/rules/20-backend.mdc`
- `.cursor/rules/40-testing.mdc`
- `.cursor/rules/41-test-contracts.mdc`
- `.cursor/rules/56-regression-verify.mdc`
- `.cursor/rules/60-tz-constraints.mdc`

**Playbooks (10):**

- `.cursor/playbooks/backend.md`
- `.cursor/playbooks/backend/observability.md`
- `.cursor/playbooks/backend/perf-testing.md`
- `.cursor/playbooks/frontend.md`
- `.cursor/playbooks/git-workflow.md`
- `.cursor/playbooks/team-workflow.md`
- `.cursor/playbooks/team-workflow/reference.md`
- `.cursor/playbooks/testing.md`
- `.cursor/playbooks/testing/e2e.md`
- `.cursor/playbooks/verification/role-checklists.md`

**Agents (10):**

- `.cursor/agents/debugger.md`
- `.cursor/agents/devops-engineer.md`
- `.cursor/agents/fe1.md`
- `.cursor/agents/fe2.md`
- `.cursor/agents/fe-ver.md`
- `.cursor/agents/qa-auto1.md`
- `.cursor/agents/qa-auto2.md`
- `.cursor/agents/qa-auto-ver.md`
- `.cursor/agents/qa-manual-ver.md`
- `.cursor/agents/tech-lead.md`

**Executable skills (3):**

- `.cursor/skills/qa-gate/SKILL.md`
- `.cursor/skills/gsd-ru/goal-verify/SKILL.md`
- `.cursor/skills/gsd-ru/phase-plan/SKILL.md`

**Current docs/planning/codebase (11):**

- `docs/README.md`
- `docs/test_plan.md`
- `docs/testframework_strategy.md`
- `.planning/PROJECT.md`
- `.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md`
- `.planning/codebase/TESTING.md`
- `.planning/codebase/ARCHITECTURE.md`
- `.planning/codebase/CONVENTIONS.md`
- `.planning/codebase/STACK.md`
- `.planning/codebase/STRUCTURE.md`
- `.planning/codebase/CONCERNS.md`

**Open backlog (25):**

- `.planning/backlog/TASK-193-2026-07-30-tf-python-scaffold.md`
- `.planning/backlog/TASK-194-2026-07-30-tf-e2e-shared-fixtures.md`
- `.planning/backlog/TASK-195-2026-07-30-tf-e2e-mechanical.md`
- `.planning/backlog/TASK-196-2026-07-30-tf-e2e-rewrite.md`
- `.planning/backlog/TASK-205-2026-08-07-python-api-health-auth-slice.md`
- `.planning/backlog/TASK-206-2026-08-07-frontend-platform-auth-wave.md`
- `.planning/backlog/TASK-207-2026-08-07-frontend-tracker-messenger-tms-calls.md`
- `.planning/backlog/TASK-208-2026-08-07-frontend-remaining-domains.md`
- `.planning/backlog/TASK-209-2026-08-07-backend-auth-pilot-hook-proof.md`
- `.planning/backlog/TASK-210-2026-08-07-backend-auth-rbac-audit-ws-calls.md`
- `.planning/backlog/TASK-211-2026-08-07-backend-platform-apigateway-dev-ent.md`
- `.planning/backlog/TASK-212-2026-08-07-backend-tracker-project-sprint-tms.md`
- `.planning/backlog/TASK-213-2026-08-07-backend-messenger-notification.md`
- `.planning/backlog/TASK-214-2026-08-07-backend-business-domains.md`
- `.planning/backlog/TASK-215-2026-08-07-backend-integrations-ai-storage.md`
- `.planning/backlog/TASK-216-2026-08-07-python-integration-expansion.md`
- `.planning/backlog/TASK-217-2026-08-07-python-e2e-smoke-greenfield.md`
- `.planning/backlog/TASK-218-2026-08-07-test-collection-skip-sentinel-gates.md`
- `.planning/backlog/TASK-226-2026-08-07-root-python-testframework-contract.md`
- `.planning/backlog/TASK-227-2026-08-07-python-security-platform-components.md`
- `.planning/backlog/TASK-228-2026-08-07-python-work-realtime-components.md`
- `.planning/backlog/TASK-229-2026-08-07-python-business-integration-components.md`
- `.planning/backlog/TASK-230-2026-08-07-frontend-static-contract-python.md`
- `.planning/backlog/TASK-231-2026-08-07-python-accessibility-coverage.md`
- `.planning/backlog/TASK-232-2026-08-07-python-security-privacy-audit.md`

**CI/routing/tooling (2):**

- `.github/labeler.yml`
- `scripts/reorganize-testframework-pipeline.ps1` — delete; do not retarget or preserve as executable tooling.

**Versioned routing-evaluation artifacts (4):**

- `.planning/audits/qa-canon/routing-scenarios.v1.json`
- `.planning/audits/qa-canon/routing-receipt.schema.v1.json`
- `.planning/audits/qa-canon/validate-routing-receipt.py`
- `.planning/audits/qa-canon/routing-receipt.v1.json`

<read_first>
Approved architecture spec §§1–6,9–12; exact 74 paths above; current active-reference audit; protected/legacy exclusions below. Treat every existing old-path instruction as stale routing data unless it is an explicitly preserved historical/negative reference.
</read_first>

**Consumes:** Founder-approved physical ownership map and the audited active surface that currently directs workers to the obsolete direct Python layers.

**Produces:** One active routing canon: executable suites at `testframework/tests/{api,integration,contract,e2e}/**`; reusable modules under `testframework/src/hub_testkit/**` by named responsibility; composition at `testframework/tests/conftest.py`; framework contracts at `testframework/selftests/**`; intent at `testframework/cases/**`; unchanged `testframework/load/**` and task-owned `testframework/reports/<task-run>/**`.

- [ ] Rewrite active instructions so executable suite ownership and task-specific reusable-module ownership are separate. No role receives blanket ownership of all `testframework/src/hub_testkit/**`; each task names its module/domain paths.
- [ ] Replace old suite routing with `testframework/tests/<layer>/**`. Route reusable DB/broker clients to `testframework/src/hub_testkit/integration/**`, contract validators to `testframework/src/hub_testkit/contract/**`, browser lifecycle/artifacts to `testframework/src/hub_testkit/browser/**`, and UI pages/components to `testframework/src/hub_testkit/domains/<domain>/ui/{pages,components}/**`.
- [ ] Route root composition to `testframework/tests/conftest.py`; route framework-contract tests to `testframework/selftests/**`. Do not create `testframework/core/**`, `testframework/tests/e2e/{support,pages,steps}/**` or a new generic warehouse.
- [ ] Update `/qa-gate` routing prose so profiles select `testframework/tests/<layer>/**`, add the missing `python-contract` ownership, and keep `full` a launcher profile rather than a pytest marker. Do not change legacy `backend/testframework/**` CI consumers by blind Python-path replacement; their switch remains a sealed Phase 129 consumer operation.
- [ ] Delete `scripts/reorganize-testframework-pipeline.ps1`: it moves the new E2E surface back into legacy TypeScript stages and therefore cannot be safely retargeted.
- [ ] Create `routing-scenarios.v1.json` with these literal prompt/expected-route pairs: (1) `Add a black-box health API test and its reusable client.` → `testframework/tests/api/health/test_health.py`, `testframework/src/hub_testkit/domains/health/api.py`; (2) `Add OpenAPI contract validation and its reusable validator.` → `testframework/tests/contract/api/test_openapi.py`, `testframework/src/hub_testkit/contract/openapi.py`; (3) `Add a PostgreSQL transaction boundary test and reusable boundary client.` → `testframework/tests/integration/postgres/test_transactions.py`, `testframework/src/hub_testkit/integration/postgres.py`; (4) `Add the login browser smoke, browser lifecycle, page, and form component.` → `testframework/tests/e2e/identity/test_login_smoke.py`, `testframework/src/hub_testkit/browser/lifecycle.py`, `testframework/src/hub_testkit/domains/identity/ui/pages/login.py`, `testframework/src/hub_testkit/domains/identity/ui/components/session_form.py`; (5) `Add a pytest marker framework contract.` → `testframework/selftests/pytest_plugin/test_markers.py`, `testframework/src/hub_testkit/pytest_plugin/markers.py`; (6) `Add a measured tracker Locust profile.` → `testframework/load/profiles/steady.toml`, `testframework/load/users/tracker.py`.
- [ ] Add exact machine-readable `hub.qa-routing-declaration/v1` anchors to the same three owned executable skills, with one unique `(scenario_id, assigned_skill, declared_routes, expected_routes, prohibited_legacy_tokens)` tuple for every scenario: `qa-gate` owns `api-health`, `integration-postgres`, `e2e-login`, `load-tracker`; `goal-verify` owns `selftest-markers`; `phase-plan` owns `contract-openapi`. `assigned_skill` is the literal skill path, `declared_routes` and `expected_routes` are sorted arrays, and `prohibited_legacy_tokens` is a non-empty sorted array of forbidden direct-layer/core routes relevant to the scenario. These declarations are the sole inputs to this gate; they are not prompts to an LLM and do not claim uncontrolled future-model behavior.
- [ ] Define `routing-receipt.schema.v1.json` as a strict no-extra-key schema for twelve ordered records (six `RED`, then six `GREEN`). Every record has deterministic provenance keys `schema_version`, `base_commit`, `candidate_head`, `candidate_content_sha256`, `scenario_sha256`, `skill_blobs`, `phase`, `scenario_id`, `prompt_sha256`, `declared_routes`, `expected_routes`, `unexpected_routes`, `verdict`, `reviewer`, `reviewed_at`, `evaluator_version`, `evaluator_sha256`. `base_commit` is exactly `31dfc7e0aeed4c2667de9d43e62637fd68d3b6f2`; `candidate_content_sha256` hashes the sorted exact three-path current-byte closure and is required even when `candidate_head` is populated, so HEAD alone is never provenance. `skill_blobs` is the exact three-path map from literal skill path to SHA-256 (BASE blobs for RED; current candidate bytes for GREEN). The schema prohibits model identity, agent-run IDs, free-form answers, secrets and any conclusion text outside the declared arrays/verdict.
- [ ] Implement `validate-routing-receipt.py` as the sole stdlib-only deterministic declaration-conformance evaluator. It reads RED only through exact literal Git blob references for the three assigned paths at the fixed BASE commit; it reads GREEN only through the exact three current candidate files, without recursion. It fails closed on a non-exact skill set; traversal or symlink escape; UTF-8/size-budget violation; duplicate/missing declaration tuple; declaration/hash mismatch; extra key/record; unsorted or duplicate routes; absent RED legacy mismatch; or any GREEN inequality/extra route. RED requires at least one declared legacy mismatch across the six BASE tuples; GREEN requires all six declared route sets to equal the expected sorted sets with zero unexpected routes. `qa-auto-ver` performs conclusion-blind review of evaluator inputs/output, `arch-ver` reviews route ownership, and the security reviewer verifies bounded reads, provenance and path containment. Requested runtime metadata may be recorded, but `actual_model=UNKNOWN` unless observable.
- [ ] Preserve only explicit historical/negative references: this master plan's old-layout table/sentinel; `.planning/codebase/ARCHITECTURE.md` legacy README pointer; `.planning/codebase/STRUCTURE.md` legacy pointer/junction inventory; `.planning/STATE.md` historical Go paths. These references must be labelled as historical/forbidden and must not grant ownership or execution authority.
- [ ] Run a deterministic static source scan over the 70 pre-existing active paths inside the exact 74-path closure: after excluding the named historical/negative regions, zero active instruction may route Python framework output/execution to `testframework/{api,integration,contract,e2e}/**`, `testframework/core/**` or `testframework/conftest.py`. Exclude `.planning/audits/test-rebase/**` guard/inventory/source evidence, `pro-qa-ai/**`, `.planning/archive/**` and every path outside the closure; do not exclude the four new `qa-canon` artifacts.
- [ ] Independent docs/architecture/security reviewers compare the exact diff with approved spec and confirm counts `9+10+10+3+11+25+2+4=74`, zero out-of-allowlist mutation, schema-valid deterministic BASE-vs-candidate receipt and zero HIGH. This proves declared routing-contract conformance only; it neither samples nor predicts uncontrolled future-LLM behavior and grants no target-runtime readiness. The deterministic evaluator is statically reviewed now; its Docker-only canonical replay is `NOT_READY` until G129-L, replays only this evaluator, and becomes a predecessor of G129-P.

<verify>
<automated>Static pre-launcher review of the exact 74-path diff, the twelve-record deterministic BASE-vs-candidate declaration receipt and old-active-reference scan. Expected: 74 paths accounted for, reorganize script deleted, literal scenario/declaration hashes and sorted routes valid, at least one BASE legacy mismatch, every candidate declaration exact with zero extras, only enumerated historical/negative references remain, and independent reviewers record zero HIGH. The canonical replay is `NOT_READY` until G129-L; once G129-L is independently accepted, run exactly `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -NoProfile -NonInteractive -File testframework/docker/run-task.ps1 -ContractVersion hub.qa-runner/v1 -PlanId canon-sync -TaskRunId $HubTaskRunId -TaskId TASK-226 -OwnerRole tech-lead -VerifierRoles qa-auto-ver,arch-ver -Mode verify -Profile canon-routing-contract -ImageLock testframework/docker/images.lock.json -InnerTimeoutSeconds 300 -TimeoutSeconds 420`. The accepted profile replays only the pinned stdlib evaluator and writes the bounded `testframework/reports/<task-run>/routing-receipt-check.json`; before G129-L this command is not a G-CANON prerequisite or fallback.</automated>
</verify>

**Done:** active authority/rules/playbooks/agents/skills/docs/backlog/label routing consistently express the approved physical tree, while inventory/guard evidence and protected/archive surfaces remain untouched; no runtime claim is made.

**Atomic commit:** `docs(testframework): синхронизировать active QA canon`; stage exactly the 74 paths above, including four versioned evaluation artifacts and deletion of `scripts/reorganize-testframework-pipeline.ps1`; no task-run report is staged. Stop on dirty overlap or Git lock/process.

### Task META-PLAN: Rewrite and independently accept the live downstream plans

**Files:** Modify or create only this exact 23-path downstream planning/checker allowlist:

- `.planning/phases/128-tf-python-scaffold/PLAN.md`
- `.planning/phases/129-tf-e2e-shared/PLAN.md`
- `.planning/phases/129-tf-e2e-shared/129-01-PLAN.md`
- `.planning/phases/129-tf-e2e-shared/129-02-PLAN.md`
- `.planning/phases/129-tf-e2e-shared/129-03-PLAN.md`
- `.planning/phases/129-tf-e2e-shared/129-04-PLAN.md`
- `.planning/phases/129-tf-e2e-shared/129-05-PLAN.md`
- `.planning/phases/129-tf-e2e-shared/129-06-PLAN.md`
- `.planning/phases/129-tf-e2e-shared/129-07-PLAN.md`
- `.planning/phases/130-tf-traceability-gates/PLAN.md`
- `.planning/phases/131-tf-python-core-auth/PLAN.md`
- `.planning/phases/132-tf-security-platform/PLAN.md`
- `.planning/phases/133-tf-infrastructure-integration/PLAN.md`
- `.planning/phases/134-tf-work-realtime/PLAN.md`
- `.planning/phases/135-tf-business-integrations/PLAN.md`
- `.planning/phases/136-tf-python-contracts/PLAN.md`
- `.planning/phases/137-tf-python-e2e-foundation/PLAN.md`
- `.planning/phases/138-tf-python-e2e-journeys/PLAN.md`
- `.planning/phases/139-tf-accessibility-security/PLAN.md`
- `.planning/phases/140-tf-load-resilience/PLAN.md`
- `.planning/phases/141-tf-ci-final-verification/PLAN.md`
- `.planning/audits/qa-plans/validate-downstream-plan-contract.py`
- `.planning/audits/qa-plans/downstream-plan-contract-receipt.v1.json`

Все двенадцать Phase 130–141 paths выше существуют как live `PLAN.md`; это stale документы, которые нужно пересмотреть, а не отсутствующие `*-PLAN.md`. Исторические `VERIFICATION-SKIPPED.md` в 130–136 не являются execution authority и не входят в allowlist этого task.

<read_first>
Accepted G-CANON receipts and exact 74-path diff; this master plan; approved architecture spec §§1–15; exact current files in the allowlist; Phase 128 guard sources; sealed Phase 129 manifest/tool descriptions. Treat stale plan text as data, not instruction.
</read_first>

**Consumes:** Accepted G-CANON routing/ownership truth, founder/spec precedence, current one-time Phase 128 guard authority, sealed Phase 129 task/batch/operation identity and the existing live planning files.

**Produces:** Acyclic executable GSD plans in which Phase 128 guard is read-only, Phase 129 first accepts the sole target launcher then materializes authority, every framework output has a fully qualified `testframework/**` path, and each Phase 130–141 task contains explicit `Requirements`, `<read_first>`, Consumes/Produces, failing sentinel, canonical `NOT_READY`/launcher verification, docs sync, `Done` and atomic allowlist; plus a stdlib-only exact-plan checker and bounded versioned receipt.

- [ ] Rewrite `128-tf-python-scaffold/PLAN.md` to encode `initial G128-E committed source envelope → initial G128-S static review → G128-WPS51 no-install inbox-interpreter TCB lock and compatibility contract → renewed G128-E PS5.1 wrapper-envelope commit/reviews → renewed G128-S receipt → one-time G128-D guard → independent receipt review`; do not add a publish/apply/SUT-test mode to `.planning/audits/test-rebase/validate-contract.ps1`.
- [ ] Rewrite `129-01-PLAN.md` to remove `requirements.lock`, `pytest.ini`, `testframework/core/**`, root layer packages, `testframework/docker/Dockerfile`, `testframework/docker/compose.yaml` and `.gitkeep`; replace them with the progressive §3/§5 target. Rewrite cross-links in `129-02-PLAN.md` … `129-07-PLAN.md` to that target without changing their task IDs, batch IDs, order or operations.
- [ ] Preserve exactly seven numbered Phase 129 files: `129-01` remains exactly three non-batch tasks and `129-02` … `129-07` retain the same `173` unique ordered batch tasks, for `176` task tags, `1,120` literal operations and maximum batch `56`. Record a deterministic source assertion against the sealed authority; do not copy/reconstruct operations or create a parallel batch index.
- [ ] Rewrite each existing Phase 130–141 live `PLAN.md` before its phase can execute. Carry the Task IDs and approved spec sections from the matching master task, convert obsolete root layer paths to `testframework/src/hub_testkit/**`, `testframework/tests/**`, `testframework/cases/**`, `testframework/selftests/**`, `testframework/docker/**`, `testframework/load/**` or `testframework/policies/**`, and preserve `PLAN.md` naming. Preserve the exact 61-task distribution `4/4/4/3/8/18/3/3/5/3/3/3`; the added tasks are the central Phase 130 profile registry, the split Phase 131 HTTP/runtime boundaries, the split Phase 134 messenger/notification/closure boundaries and one bounded docs/traceability closure in every Phase 130–140.
- [ ] Add a negative planning sentinel that fails if any downstream task declares framework output at repo-root `hub_testkit/**`, `tests/**`, `cases/**`, `selftests/**`, `docker/**` or `load/**`, or declares target `testframework/core/**`, `testframework/requirements.lock`, `testframework/pytest.ini`, `testframework/docker/Dockerfile`, `testframework/docker/compose.yaml` or any scaffold `.gitkeep`.
- [ ] Implement `validate-downstream-plan-contract.py` with an embedded literal ordered 21-plan allowlist and bounded descriptor-safe UTF-8 reads. It rejects an extra/missing/duplicate input plan; missing Task IDs/spec sections/`Files`/`read_first`/Consumes/Produces/failing sentinel/verify/Done/atomic allowlist; any docs mutation or staged docs path absent from that task's literal `Files`; any shared phase docs path owned outside the one dependent docs/traceability closure; any task above 14 expanded paths; any active forbidden token; unqualified framework output; non-`PLAN.md` authority for 130–141; Phase 130–141 task-count drift from exact `61` and distribution `4/4/4/3/8/18/3/3/5/3/3/3`; or Phase 129 mismatch from exactly `3 + 173 = 176` tasks, `173` unique ordered batch IDs, `1,120` literal operations and maximum `56`. It never imports/executes plan content or reconstructs operations.
- [ ] Add a profile-closure assertion to the same checker: extract every Phase 130–141 verify-profile literal and require exact membership in the 47-entry later-profile registry declared by Task 130-A; reject duplicates, missing image role/limit class/target/deadline, any unknown profile, and every `testframework/runner/profiles.py` write after Task 130-A. Early Phase 129 bootstrap entries remain separately sealed by G129-L.
- [ ] Add the founder stop assertion: Phase 130 ends at G130-H, every Phase 131–141 live plan/task declares `NOT_READY` plus exact `G-ACTIVATE-<phase>` precondition, and no dependency edge automatically advances past G130-H. Reject a bulk phase range, universal coverage sprint, activation inferred from profile registration or a receipt that omits one feature/risk/file/rollback boundary.
- [ ] Define `downstream-plan-contract-receipt.v1.json` with exact keys `schema_version`, `full_head`, `checker_blob_sha256`, `master_blob_sha256`, `ordered_plan_paths`, `plan_blob_sha256`, `task_count`, `batch_ids_sha256`, `batch_count`, `operation_count`, `maximum_batch`, `forbidden_match_count`, `field_coverage`, `verdict`, `reviewer`, `reviewed_at`; prohibit extras and bound the file to 256 KiB. The pre-launcher receipt is completed conclusion-blind by independent `gsd-plan-checker` from static source review and is invalid on any plan/checker/master hash drift.
- [ ] Independent `gsd-plan-checker` reviews all 23 paths, checker negative fixtures expressed as in-memory data, path/interface/signature consistency, task-local fields, no command before its named gate and zero HIGH. After G129-L, qa-auto-ver executes the same checker through the accepted launcher to produce `testframework/reports/<task-run>/downstream-plan-contract-check.json`; exact semantic equality with the versioned receipt is G-META-X and a predecessor of G129-P.

<verify>
<automated>NOT_READY until G-CANON. Pre-launcher G-META static review covers all 23 paths and the bounded versioned receipt: exact 21 plans, forbidden matches zero, every 130–141 task field present, Phase 129 `176 / 173 / 1120 / 56` with original ordered identities. The executable checker is `NOT_READY` until G129-L; once G129-L is independently accepted, run exactly `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -NoProfile -NonInteractive -File testframework/docker/run-task.ps1 -ContractVersion hub.qa-runner/v1 -PlanId meta-plan -TaskRunId $HubTaskRunId -TaskId TASK-226 -OwnerRole tech-lead -VerifierRoles gsd-plan-checker,qa-auto-ver -Mode verify -Profile downstream-plan-contract -ImageLock testframework/docker/images.lock.json -InnerTimeoutSeconds 300 -TimeoutSeconds 420`. The accepted profile invokes `.planning/audits/qa-plans/validate-downstream-plan-contract.py` only against its embedded 21-plan allowlist and writes the bounded `testframework/reports/<task-run>/downstream-plan-contract-check.json`; mismatch keeps G-META-X non-passing and blocks G129-P.</automated>
</verify>

**Done:** `G-META` has an independent zero-HIGH static receipt for the exact 23-path closure and unchanged sealed identities; after G129-L, G-META-X reproduces it through the accepted launcher before authority materialization.

**Atomic commit:** `docs(testframework): синхронизировать планы Python QA rebase`; stage exactly the 23 allowlisted paths above: 21 live plans, checker and versioned receipt. If dirty overlap, `.git/index.lock` or an active Git process exists, stop before staging.

### Readiness gates

| Gate | Допустимый verdict до gate | Условие перехода | Что gate НЕ доказывает |
|---|---|---|---|
| G-CANON — active canon sync | `NOT_READY` | exact 74-path active/evaluation surface expresses approved ownership; script retired; twelve-record deterministic BASE-vs-candidate declaration receipt, old-active-reference scan and independent reviews zero HIGH | downstream plan correctness, sealed authority, target-runtime behavior or uncontrolled future-LLM behavior |
| G-META — downstream plan sync | `NOT_READY` до G-CANON | exact 23-path plan/checker closure accepted statically with 21 plans and sealed arithmetic unchanged | sealed authority or target-runtime behavior |
| G128-R — Phase 128 authority reseal | `NOT_READY` до G-META | exact authority/runtime closure rebased to approved physical tree; seals updated; immutable ledger/cohort/planned arithmetic independently reconciled; final security static review zero HIGH | Docker behavior, manifest publication or launcher acceptance |
| G128-E — committed historical-guard envelope | `NOT_READY` до G128-R initially; renewal `NOT_READY` до G128-WPS51 | initial committed envelope is accepted; after G128-WPS51, a renewed commit/review binds the PS5.1-compatible hardened wrapper and committed host lock to the envelope | Docker behavior, wrapper runtime, manifest publication, batch apply or launcher acceptance |
| G128-S — Phase 128 static review | `NOT_READY` до applicable G128-E | initial review authorizes G128-WPS51 only; renewed review after wrapper hardening seals the literal absolute inbox interpreter path and exact G128-D invocation | Docker behavior, manifest publication, launcher acceptance |
| G128-WPS51 — no-install host-interpreter TCB lock | `NOT_READY` до initial G128-S | the preinstalled Windows PowerShell 5.1 inbox component is sealed at exact `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe` from one held no-follow handle with volume serial/file index/reparse=false/SHA/version/CLR/build, exact two-hardlink identity and pinned Microsoft publisher/production chain/code-signing EKU; compatibility audit defines RNG/hex/argv/race-free job-object/MainModule remediation; security/architecture reviews zero HIGH; no install or wrapper run | wrapper acceptance, Docker behavior, G128-D execution or launcher acceptance |
| G128-D — one-time Phase 128 guard acceptance | `NOT_READY` до G128-WPS51 + renewed G128-E/G128-S | the literal independently sealed absolute Windows PowerShell path invokes the exact historical wrapper once; conclusion-blind raw-receipt review proves guard/determinism/timeout/cleanup | `hub.qa-runner/v1`, manifest publication or SUT behavior |
| G129-L — launcher acceptance | `NOT_READY` до G128-D receipt | sole target launcher/static contracts exist; bootstrap acceptance proves limits, first exit, reports boundary and exact cleanup | health/API behavior or cutover completion |
| G-CANON-X — executable canon receipt | `NOT_READY` до G129-L | accepted launcher runs `canon-routing-contract`; report semantically equals versioned 12-record receipt | downstream plan or authority correctness |
| G-META-X — executable downstream-plan receipt | `NOT_READY` до G129-L | accepted launcher runs exact 21-plan checker; report semantically equals versioned receipt and `176/173/1120/56` | authority publication or batch mutation |
| G129-P — authority materialization | `NOT_READY` до G-CANON-X + G-META-X | accepted `hub.qa-runner/v1` atomically writes builder-identical planned manifest and independently validates `173/1,120/max56` | any batch mutation |
| G129-C — Phase 129 cutover/scaffold | `NOT_READY` до G129-P | all 176 tasks complete in sealed order; units at target roots; legacy closure/hook files zero; health slice accepted | expanded framework/domain coverage |
| GF — framework selftests | `NOT_READY` until Phase 130 gates | serial selftests, architecture/traceability/marker/profile negative sentinels, deterministic replay, zero leaks | full product QA |
| G130-H — current-scope health/framework handoff | `NOT_READY` until G129-C + Tasks 130-A–D | independent receipt proves cutover/scaffold, sole launcher, 52-ID registry contract, framework selftests, one health TC and exact docs closure; current implementation run stops | any Phase 131–141 activation or bulk coverage expansion |
| G-ACTIVATE-131..141 — on-demand backlog activation | `NOT_READY` after G130-H by default | founder accepts exactly one phase/feature slice with current requirements, risks, target files, packages/images and rollback; downstream plan is rechecked on the activation HEAD | authorization for another phase, universal coverage sprint or pre-registered future profile execution |
| G141 — clean-room handoff | `NOT_READY` until founder activation receipts and 132–140 evidence | clean checkout/image build, all activated profiles, supply-chain/retention/onboarding/recovery, independent verifier | production deploy/sign-off |

`STATIC_REVIEW_PASS` — допустимая метка документационного review, но она никогда не преобразуется в runtime `PASS`. `NOT_READY`, `NOT_RUN`, `PASS` и `FAIL` хранятся как разные enum values; агрегатор не повышает статус. До G-CANON + G-META + G128-R запрещены wrapper run, guard-envelope authority commit и launcher work; до initial G128-E запрещён initial G128-S. Initial G128-S authorizes only G128-WPS51. G128-WPS51 installs nothing and remains `NOT_READY` until exact inbox identity, scoped two-hardlink exception and PS5.1 remediation contract are independently accepted; after it commits the host lock, wrapper hardening requires renewed G128-E and renewed G128-S. Historical wrapper is permitted exactly once only after that renewal, never declares contract `hub.qa-runner/v1` and is forbidden after G129-L. The current prelaunch attempt is non-consumed, runtime remains `NOT_RUN`, and the one G128-D invocation remains unused. Acceptance G130-H завершает текущий run и не повышает ни один backlog profile/phase из `NOT_READY`.

## 6. Dependency graph and waves

```text
Wave 0: G-CANON rewrite exact 74-path active/evaluation surface -> independent static canon review
    -> Wave 1: META-PLAN rewrite exact 23-path plan/checker surface -> independent static G-META
        -> Wave 2: G128-R reseal exact Phase 128 authority/runtime closure -> independent authority/security review
            -> Wave 3: 128-E seal the complete historical-guard source envelope -> independent architecture/security review
                -> Wave 4: initial 128-S static guard acceptance against the committed envelope
                    -> Wave 5: 128-WPS51 no-install inbox PowerShell TCB seal + compatibility contract
                        -> Wave 6: renewed 128-E wrapper envelope commit/reviews -> renewed 128-S receipt
                            -> Wave 7: 128-D one-time historical read-only guard -> independent receipt review
                                -> Wave 8: 129-01-A create/accept sole hub.qa-runner/v1
                                    -> Wave 9: G-CANON-X + G-META-X executable receipt rechecks -> G129-P authority materialization
                                        -> Wave 10: 129-01-B/C progressive framework + health slice + independent preflight
                                            -> Waves 11..16: 129-02 -> 129-03 -> 129-04 -> 129-05 -> 129-06 -> 129-07
                                                -> Wave 17: 130-A central 47-row later-profile registry -> 130-B plugin -> 130-C architecture -> 130-D docs closure
                                                -> G130-H: independent health/framework handoff -> STOP current implementation run
                                                -- only after separate G-ACTIVATE-<phase> receipt --> Wave 18: 131-A HTTP -> 131-B runtime/observability -> 131-C identity -> 131-D docs closure
                                                    -> Wave 19A: 132-A/B parallel -> 132-C risk closure -> 132-D docs closure
                                                    -> Wave 19B: 133-A -> 133-B (shared Compose) -> 133-C docs closure
                                                    -> Wave 19C: 136-A/B parallel -> 136-C docs closure
                                                    -> Wave 19D: 137-A browser -> 137-B smoke -> 137-C docs closure
                                                    -> Wave 20A: 134-A/B/C/D/E in parallel after 131; 134-F after 133; 134-G after A-F -> 134-H docs closure
                                                    -> Wave 20B: 135-A..G/M..P after 131 and H..L after 133; 135-Q after A-P -> 135-R docs closure
                                                    -> Wave 21A: 138-A/B/C in parallel after 137/134/135 -> 138-D recovery -> 138-E docs closure
                                                    -> Wave 21B: 140-A -> 140-B -> 140-C docs closure (after 131,134,135)
                                                    -> Wave 22: 139-A/B parallel after 132/138 -> 139-C docs closure
                                                        -> Wave 23: 141-A CI -> 141-B clean-room -> 141-C final docs/handoff (after all Phase 130-140 docs receipts)
```

| Phase | ROADMAP requirements | Plans/tasks | Parallel rule |
|---:|---|---|---|
| CANON | all 128–141 IDs | exact 74-path active/evaluation closure; retire unsafe script; versioned BASE-vs-candidate declaration receipt | first hard predecessor; executable replay after G129-L |
| META | all 128–141 IDs | exact 23-path closure: 21 plans + checker + receipt | static predecessor; executable replay after G129-L |
| 128 | TASK-202, TASK-203, TASK-204 | authority reseal; initial guard envelope/static review; G128-WPS51 no-install host-interpreter TCB lock and PS5.1 remediation; renewed envelope/static review; one-time read-only Docker guard | zero installation; sealed absolute inbox interpreter only; wrapper forbidden after launcher acceptance |
| 129 | TASK-193, 197, 198, 201, 204, 219, 234 | immutable 176 tasks in seven plans | intentionally linear: shared manifest/consumer registry |
| 130 | TASK-206, 218 | 4 tasks: central profiles; plugin; architecture; docs closure | single registry writer; docs closure last |
| 131 | TASK-205, 207 | 4 tasks: HTTP; runtime/observability; identity; docs closure | each task ≤14 expanded paths; docs closure last |
| 132 | TASK-208, 209, 227 | 4 tasks: authz/tenant; abuse; risk closure; docs closure | A/B disjoint, C joins, D owns shared docs |
| 133 | TASK-210, 211 | 3 tasks: PG/Mongo; Redis/NATS/jobs; docs closure | A/B serialize shared Compose; C owns shared docs |
| 134 | TASK-212, 213, 228 | 8 tasks: four work domains; messenger; notification/NATS; risk closure; docs closure | disjoint A-F; G joins evidence; H owns shared docs |
| 135 | TASK-214, 215, 229 | 18 tasks: 17 bounded domain/adapter boundaries + docs closure | A-P disjoint by domain; Q closes risk graph; R owns shared docs |
| 136 | TASK-216, 230 | 3 tasks: API/schema/event; frontend contract; docs closure | A/B disjoint, C owns shared docs |
| 137 | TASK-194, 217 | 3 tasks: browser lifecycle; smoke; docs closure | serial until image spike, then docs closure |
| 138 | TASK-195, 196 | 5 tasks: three literal journeys; recovery; docs closure | A-C parallel; D joins; E owns shared docs |
| 139 | TASK-231, 232 | 3 tasks: a11y; security/privacy; docs closure | A/B parallel after prerequisites; C owns shared docs |
| 140 | TASK-200 | 3 tasks: validity/guard; measured profiles; docs closure | separate load image/runtime; docs closure last |
| 141 | TASK-220, 233 | 3 tasks: CI; clean-room; exact docs/handoff | final serial join; profile registry read-only |

Rows 131–141 describe independent on-demand backlog contracts. Their wave/parallel relations apply only inside a separately activated slice after G130-H; they are not the continuation of the current execution run and cannot be activated as a bulk range.

## 7. Canonical task anatomy and verification command

Каждый task ниже является review boundary: соседний task можно принять или отклонить независимо. В task body слова «реализовать» всегда сопровождаются exact files, public interface, negative sentinel и observable expected outcome.

После G129-L canonical verification имеет форму:

```powershell
$HubTaskRunId = '<phase>-<task>-' + [guid]::NewGuid().ToString('N')
C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -NoProfile -NonInteractive -File testframework/docker/run-task.ps1 `
  -ContractVersion hub.qa-runner/v1 `
  -PlanId <phase-plan> `
  -TaskRunId $HubTaskRunId `
  -TaskId <TASK-ID> `
  -OwnerRole <owner> `
  -VerifierRoles <verifier-list> `
  -Mode verify `
  -Profile <accepted-profile> `
  -ImageLock testframework/docker/images.lock.json `
  -InnerTimeoutSeconds <accepted-inner> `
  -TimeoutSeconds <accepted-outer>
```

Angle-bracket fields здесь обозначают schema-bound per-task values, перечисленные в каждом `<verify>` ниже; executor подставляет их из task contract, не из environment/repo data. Для всех runs `outer > inner`; calibrated CPU/RAM/PID/shm/workers берутся из accepted profile in `images.lock.json` и повторяются в sealed run manifest. До G129-L exact `<verify>` содержит `NOT_READY` и условие разблокировки, а не альтернативную команду.

## 8. Phase 128 — static authority review and one-time read-only guard

### Task G128-R: Reseal the exact architecture authority closure

**Files:** Modify exactly `.planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md`, `.planning/audits/test-rebase/test-architecture-contract.json`, `.planning/audits/test-rebase/validate_test_architecture_contract.py`, `.planning/audits/test-rebase/build_cutover_manifest.py`, `.planning/audits/test-rebase/validate_cutover_manifest.py`.

**Requirements:** `TASK-202`, `TASK-203`, `TASK-204`, `TASK-226`; approved spec §§1–6,9–12,14.

<read_first>
Accepted G-CANON and G-META receipts/diffs; approved architecture spec §§1–6,12; the five exact files above; read-only `.planning/audits/test-rebase/validate-contract.ps1`, `.planning/audits/test-rebase/inventory.json`, `.planning/audits/test-rebase/founder-greenfield-cutover-manifest.json`, `.planning/audits/test-rebase/inventory.py` and `.planning/audits/test-rebase/apply_cutover_batch.py` only to establish unchanged authority/hash bindings.
</read_first>

**Consumes:** Accepted active routing/downstream plans; current descriptor-safe architecture contract; unchanged founder inventory/manifest; old contract digest bindings in both cutover builder and independent validator.

**Produces:** One internally consistent authority set whose physical semantics are `testframework/src/hub_testkit/**` reusable modules, `testframework/tests/{api,integration,contract,e2e}/**` executable suites, `testframework/selftests/**`, `testframework/cases/**`, `testframework/schemas/**`, `testframework/policies/**`, `testframework/docker/**`, `testframework/runner/**`, unchanged `testframework/load/**` and task-owned `testframework/reports/**`; builder and independent validator seal the newly reviewed architecture-contract digest without changing cutover identities.

- [ ] Start with failing static authority sentinels: the current `python_layers` direct layout/`core` allowance must be rejected against the approved spec; builder and independent validator must reject the current architecture-contract digest after the contract changes but before their bindings are updated.
- [ ] Rebase ADR and JSON contract semantics only: separate executable suite roots from reusable-module roots, forbid generic `core`, keep `testframework/tests/conftest.py` composition-only, preserve load/report responsibilities, and encode no empty directory or speculative layer requirement.
- [ ] Update `validate_test_architecture_contract.py` with structural positive/negative cases for the approved tree, including denial of direct layer roots, repo-root framework outputs, root conftest, generic core and UI support/pages/steps warehouses.
- [ ] Rebase the builder's `python-scaffold` gate and the independent validator's matching gate literals from direct layers/core to the approved progressive tree. Derive the new architecture-contract SHA-256 from the reviewed canonical JSON bytes and update exactly the corresponding binding in both files. Do not write a prospective hash in this plan, copy it manually from prose or change any unrelated inventory/founder/group/source-path seal.
- [ ] Perform deterministic static reconciliation, without Docker: inventory remains `1,049` rows with the same ledger SHA; cohorts remain `507 move / 245 hook-blocked / 297 delete + 28 non-ledger delete`; Phase 129 remains `176 tasks / 173 ordered batch IDs / 1,120 operations / max 56`, all planned and identity-preserving.
- [ ] Inspect wrapper/runtime hash closure. `validate-contract.ps1` snapshots the exact committed validator/builder bytes at run time and contains no embedded ADR/architecture-contract digest, so a layout-only reseal keeps it read-only. `inventory.py`, `inventory.json`, `founder-greenfield-cutover-manifest.json`, `apply_cutover_batch.py` and non-ledger closure also remain read-only: the founder manifest's broad `testframework/**` source scope is still valid and approved spec §12 preserves it. Any contrary source finding stops G128-R for master-plan revision; do not broaden the commit ad hoc.
- [ ] `arch-ver` performs conclusion-blind authority diff review, then `security-auditor` performs the final no-follow/hash/scope/static review. Any unreviewed seal, changed ledger/group/operation identity, out-of-allowlist file or HIGH keeps G128-R `NOT_READY`.

<verify>
<automated>NOT_READY until G-CANON and G-META. Then static read-only reconciliation of the exact five-file diff and read-only authority sources; no Docker, builder, validator, test or Git command. Expected: new contract semantics and mutually matching derived contract-digest bindings, unchanged `1049 / 507-245-297+28 / 176-173-1120-max56`, independent authority review and final security static review.</automated>
</verify>

**Done:** G128-R has independent architecture/security receipts for the exact five-file reseal, no invented digest, unchanged inventory/cutover identities and no runtime claim; only then may G128-E seal the historical guard source envelope.

**Atomic commit:** `docs(testframework): reseal Python QA authority`; stage exactly the five files above. No generated manifest/evidence is staged; stop on dirty overlap or Git lock/process.

### Task G128-E: Seal the complete historical-guard source envelope

**Files:** Initial pass: Create/first-track exactly `.planning/audits/test-rebase/validate-contract.ps1`, `.planning/audits/test-rebase/apply_cutover_batch.py`, `.planning/audits/test-rebase/founder-greenfield-cutover-manifest.json`; Modify exactly `.planning/audits/test-rebase/inventory.py`, `.planning/audits/test-rebase/inventory.json`; Review only `.planning/audits/test-rebase/inventory.schema.json`, `.planning/audits/test-rebase/test-architecture-contract.json`, `.planning/audits/test-rebase/validate_test_architecture_contract.py`, `.planning/audits/test-rebase/build_cutover_manifest.py`, `.planning/audits/test-rebase/validate_cutover_manifest.py`, `.planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md`. Mandatory renewal after G128-WPS51: Modify only `.planning/audits/test-rebase/validate-contract.ps1`; Review only the same ten dependencies plus committed `.planning/audits/test-rebase/powershell-host-lock.json`.

**Requirements:** `TASK-202`, `TASK-203`, `TASK-204`; approved spec §§1,2,5,6,12,14.

<read_first>
Accepted G128-R architecture/security receipts; current Git tracking state for the eleven exact envelope paths; revised Phase 128 PLAN; wrapper source and all files it pins, executes or consumes; `.cursor/rules/07-agent-safety.mdc`.
</read_first>

**Consumes:** Initially G128-R commit `fd3a5671` and its immutable architecture/ledger/founder identities; on the mandatory renewal, accepted G128-WPS51 with its committed host lock and compatibility findings.

**Produces:** Initially one committed full-HEAD source envelope. After G128-WPS51, a renewed envelope commit/review in which `validate-contract.ps1` is PS5.1 compatible, reads the committed host lock, records the current interpreter canonical path, held-handle volume serial/file index/reparse=false/link set, hash/version and pinned Microsoft publisher/production-chain/code-signing identity in its evidence, and fails closed unless every sealed value exactly matches. Neither pass produces a runtime receipt, manifest, batch transition or target launcher.

- [ ] Begin with the failing tracking sentinel: `validate-contract.ps1`, `apply_cutover_batch.py` and `founder-greenfield-cutover-manifest.json` are currently untracked, while `inventory.py` and `inventory.json` differ from `HEAD`; therefore G128-S must remain `NOT_READY` before this task.
- [ ] Review the current five writable-path bytes as one closed source envelope. Reject any hidden parameter/environment override, repository write outside the already declared inventory-generation path, SUT profile, manifest publication, batch application, protected-zone access, mutable executable path or dependency on a file outside the eleven literal paths.
- [ ] Reconcile immutable identities without execution: inventory remains exactly `1,049` rows with SHA-256 `72de6a81b2bd5da8873a98e784a5069a661e9caea083f10722aa85e9f91b4511`; founder manifest remains `df503a7f…`; architecture contract remains `62eb27ecd93164cf404639e6c8b85d14843613da2d34ea4da08101a673bf1335`; cohorts/operations remain `507/245/297+28` and `173/1,120/max56`.
- [ ] `arch-ver` reviews the exact five-path create/modify diff plus the six read-only same-HEAD dependencies for architecture/identity preservation. `security-auditor` independently reviews wrapper hash closure, no-follow/protected-zone/command boundaries and confirms no runtime/publish/apply path was exercised.
- [ ] Commit exactly the five writable envelope paths. After commit, all eleven envelope paths must resolve to the same committed `HEAD` tree and the five writable paths must have clean worktree/index status. Any drift, missing tracked blob or extra staged path keeps G128-E `NOT_READY`.
- [ ] After accepted G128-WPS51, renew this gate: modify only `validate-contract.ps1` to replace `RandomNumberGenerator.Fill`/`Convert.ToHexString`, `ProcessStartInfo.ArgumentList`, `ProcessStartInfo.Environment` and `Process.Kill(true)` with PS5.1/.NET Framework-compatible RNG/BitConverter and EnvironmentVariables, plus a null/NUL-rejecting Windows argv encoder with bounded quote/backslash/empty/non-ASCII selftests. Launch via CreateProcessW with CREATE_SUSPENDED and restricted inherited handles, assign to a native job before ResumeThread, set JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE, forbid breakaway and cmd, then bounded wait/TerminateJobObject/drain/finally-close every handle and require ActiveProcesses=0 or NOT_READY. Bind `Diagnostics.Process.GetCurrentProcess().MainModule.FileName` (never PATH discovery) to the committed held-handle lock. The interpreter hardlink exception applies only to the exact signed Windows inbox file identity with two sealed System32/WinSxS links; every other pinned file remains single-link. Commit only the wrapper; independently re-review the new twelve-path envelope. This renewal invalidates the initial G128-S receipt.

<verify>
<automated>Static read-only source/Git-tree review only; no Python, validator, builder, Docker, test, SUT, manifest publication, apply, rollback or report creation. Initial expected result: exact eleven-path committed envelope and five-path clean writable subset. Renewal expected result: separately committed host lock plus wrapper-only commit, exact twelve-path envelope, fail-closed current-process identity binding, immutable `1049 / 507-245-297+28 / 173-1120-max56` identities and independent architecture/security PASS.</automated>
</verify>

**Done:** Initial G128-E records one reviewed committed source envelope. After G128-WPS51, renewed G128-E records the PS5.1-compatible hardened wrapper commit and renewed envelope reviews. Neither makes a runtime claim; G128-S must be produced again against the renewed exact `HEAD`.

**Atomic commit:** Initial: `chore(testframework): seal historical guard source`, stage exactly the five create/modify paths above. Renewal: `fix(testframework): support sealed Windows PowerShell 5.1`, stage only `.planning/audits/test-rebase/validate-contract.ps1`. Exclude `VERIFICATION.md`, host-lock file already committed by G128-WPS51, baseline/debug evidence, `STATE`/ROADMAP/COVERAGE/backlog churn, frontend dirt, reports, generated `cutover-manifest.json`, handoff artifacts and `pro-qa-ai/**`.

### Task 128-A: Statically accept the historical guard bootstrap

**Files:** Review only `.planning/audits/test-rebase/validate-contract.ps1`, `.planning/audits/test-rebase/validate_test_architecture_contract.py`, `.planning/audits/test-rebase/build_cutover_manifest.py`, `.planning/audits/test-rebase/validate_cutover_manifest.py`, `.planning/audits/test-rebase/apply_cutover_batch.py`, the three sealed authority JSON files, the G-META-revised `.planning/phases/128-tf-python-scaffold/PLAN.md` and, for renewal, `.planning/audits/test-rebase/powershell-host-lock.json`; Modify only `.planning/audits/test-rebase/VERIFICATION.md` for the static receipt.

**Requirements:** `TASK-202`, `TASK-203`, `TASK-204`; approved spec §§1,2,5,6,12,14.

<read_first>
AGENTS.md; `.cursor/REGULATION.md` §§4,8,15; `.cursor/rules/07-agent-safety.mdc`; accepted G-CANON/G-META/G128-R receipts; exact files above; approved architecture spec §§1,5,6,12.
</read_first>

**Consumes:** The applicable G128-E committed full-HEAD guard envelope, G128-R-resealed authority, sealed ledger `1049 / sha256 72de6a81...`, current no-argument wrapper, builder stdout contract, independent validator and immutable Phase 129 arithmetic. Renewal additionally consumes accepted G128-WPS51 and the renewed G128-E wrapper commit.

**Produces:** Initial G128-S static receipt authorizing only G128-WPS51. After wrapper hardening, a renewed G128-S receipt proves the one-time guard-only exception, seals literal `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe` and the full invocation, held-handle interpreter identity, Microsoft publisher/production chain/code-signing EKU and race-free job lifecycle; it retains no accepted parameter surface, writable repo mount, SUT profile, `--write`, batch application, authority publication or `hub.qa-runner/v1` claim.

- [ ] Review the exact committed wrapper/tool hashes and all writable/mount/process paths. A dirty/untracked envelope path, mutable image, hidden CLI/environment override, writable source/authority mount, missing limits, broad cleanup or wrapper path drift keeps G128-S `NOT_READY`.
- [ ] Prove from source that `.planning/audits/test-rebase/cutover-manifest.json` must remain absent before and after the wrapper and that builder/validator/applier negative sentinels do not mutate the repository.
- [ ] Initial pass records that G128-D remains forbidden and only G128-WPS51 may proceed. After G128-WPS51 and renewed G128-E, repeat the full review; record the one-shot window and the exact literal absolute command. The receipt expires on full HEAD, wrapper/lock/tool hash, interpreter version/signature/hardlink identity, image digest or sealed-authority drift.
- [ ] Reject unqualified `powershell.exe`, every `pwsh`, PATH/PATHEXT lookup, `Get-Command` discovery, aliases, shims, environment-variable substitution and fallback. The renewed wrapper must record current-process path/hash/file/product/PowerShell/CLR version, signature and exact two-link inbox identity and fail closed before Docker work if any locked value differs.
- [ ] If current source cannot support all assertions safely, record the exact failed assertion and retain `NOT_READY`; do not add a second launcher or expand the wrapper.

<verify>
<automated>Read-only static source review on host; no validator, builder, Docker or test command. Initial expected result is a G128-S receipt that can authorize only G128-WPS51. Renewal expected result is a new receipt with exact wrapper/host-lock hashes, literal absolute Windows PowerShell invocation, complete PS5.1 remediation and zero HIGH, or an explicit `NOT_READY` failed assertion.</automated>
</verify>

**Done:** Initial G128-S cannot authorize G128-D. Only renewed G128-S names the exact literal-absolute one-time command, revision/hash/interpreter envelope and prohibition set; neither receipt claims runtime PASS or creates authority.

**Atomic commit:** `docs(testframework): зафиксировать static guard review`; stage only `.planning/audits/test-rebase/VERIFICATION.md`; body references TASK-204.

### Task G128-WPS51: Seal the built-in Windows PowerShell 5.1 TCB without installation

**Files:** Create and commit exactly `.planning/audits/test-rebase/powershell-host-lock.json`; Review only `.planning/audits/test-rebase/validate-contract.ps1` and exact built-in `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe` identity. Wrapper modification belongs only to the later renewed G128-E task.

**Requirements:** `TASK-204`; approved spec §§1,5,6,12,14; host-interpreter prelaunch incident repair.

<read_first>
Accepted initial G128-S receipt; founder no-install decision; `.cursor/rules/07-agent-safety.mdc`; independent PS5.1 compatibility audit; exact Windows inbox interpreter metadata; current wrapper source.
</read_first>

**Consumes:** Accepted initial G128-S as static evidence only. A failed or premature prelaunch attempt is non-consumed and does not spend the single G128-D invocation.

**Produces:** One committed `powershell-host-lock.json` sealing the only permitted host interpreter TCB and its scoped signed-inbox two-hardlink identity, plus an independently accepted exact PS5.1 remediation contract needed by the renewed wrapper envelope.

- [ ] Install or download nothing. Accept only exact canonical `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`; reject unqualified `powershell.exe`, every `pwsh`, PATH/PATHEXT or `Get-Command` discovery, aliases, shims, ExecutionPolicy bypass and fallback.
- [ ] Open that canonical file once no-follow/no-reparse and keep the handle held. Capture and later recheck `volume_serial`, `file_index`, `reparse_point=false` and link count with GetFileInformationByHandle while enumerating hardlinks; reopen every link no-follow and require the same held identity. Accept exactly two System32/WinSxS links only for this sealed signed Windows inbox component. Compute SHA256/version/build from the same held identity. Require Authenticode Valid plus exact Microsoft Windows Publisher leaf, pinned Microsoft Windows production intermediate/root chain and code-signing EKU OID `1.3.6.1.5.5.7.3.3`; arbitrary Valid is rejected.
- [ ] Create `powershell-host-lock.json` with exact required keys `schema_version`, `absolute_path`, `volume_serial`, `file_index`, `reparse_point`, `sha256`, `file_version`, `product_version`, `powershell_version`, `ps_edition`, `clr_version`, `authenticode_status`, `authenticode_publisher`, `authenticode_signer`, `authenticode_issuer`, `authenticode_thumbprint`, `authenticode_chain`, `authenticode_eku_oids`, `provenance_type`, `windows_build`, `component_target`, `hardlink_count`, `accepted_hardlink_paths`, `inspected_at`; require exact inbox provenance, false reparse, hardlink_count two, absolute paths, lowercase 64-hex SHA256, complete pinned publisher/chain/EKU and no extra keys. Commit only this file and obtain independent `security-auditor` + `arch-ver` acceptance.
- [ ] The renewed G128-E remediation is exact: RNG.Create/GetBytes/Dispose plus BitConverter hex; EnvironmentVariables; null/NUL-rejecting Windows argv encoder with quote/backslash/empty/non-ASCII selftests; CreateProcessW CREATE_SUSPENDED, AssignProcessToJobObject before ResumeThread, JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE, no breakaway/cmd, bounded terminate/wait/drain/finally handles and ActiveProcesses=0 or NOT_READY; `Diagnostics.Process.GetCurrentProcess().MainModule.FileName` replaces unavailable `Environment.ProcessPath`.
- [ ] Do not invoke `validate-contract.ps1`. The next task is renewed G128-E, which alone may modify and commit the wrapper to verify/record the current process executable identity and fail closed; renewed G128-S is mandatory afterward because wrapper bytes change.

<verify>
<automated>Execution is `NOT_READY` in this planning revision. Verification is bounded read-only host/file identity and static compatibility inspection only: exact Windows inbox provenance, exact lock fields, scoped two-hardlink identity, exact remediation inventory and independent security/architecture review; no install, wrapper/validator/builder/Docker/SUT/G128-D execution. Runtime remains `NOT_RUN`.</automated>
</verify>

**Done:** Host lock is committed and independently accepted; no G128-D attempt is consumed. Renewed G128-E and renewed G128-S remain hard predecessors.

**Atomic commit:** `chore(testframework): seal Windows PowerShell 5.1 host`; stage only `.planning/audits/test-rebase/powershell-host-lock.json`; no installer/download/package artifact, wrapper, report, manifest or unrelated file.

### Task 128-B: Run the one-time guard and review its raw receipt

**Files:** Create task-owned ignored `testframework/reports/<task-run>/**` evidence; Modify only `.planning/audits/test-rebase/VERIFICATION.md` after independent review; Read only `.planning/audits/test-rebase/powershell-host-lock.json` and the renewed `.planning/audits/test-rebase/validate-contract.ps1`. `.planning/audits/test-rebase/cutover-manifest.json` remains absent.

**Requirements:** `TASK-202`, `TASK-203`, `TASK-204`; approved spec §§5,6,12,14.

<read_first>
Accepted G128-WPS51 host lock, renewed G128-E envelope and renewed G128-S receipt; exact `.planning/audits/test-rebase/validate-contract.ps1`; all sealed authority files; revised Phase 128 PLAN; `.cursor/playbooks/verification.md`.
</read_first>

**Consumes:** Exact G128-WPS51 plus renewed G128-E/G128-S full HEAD, wrapper/host-lock/interpreter/tool/image/authority identities and the renewed receipt's literal absolute historical guard command with no wrapper parameters.

**Produces:** One G128-D raw evidence set and a conclusion-blind `arch-ver` + `qa-auto-ver` receipt covering inventory, guard sensitivity, deterministic builder bytes, independent validator/applier sentinels, inner/outer timeout, first exit and exact cleanup.

- [ ] G128-D has no executable command while the host lock or renewed G128-S is absent. After acceptance, execute exactly once as `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -NoProfile -NonInteractive -File .planning/audits/test-rebase/validate-contract.ps1`, copied verbatim into renewed G128-S. No PATH lookup, variable/alias/shim, unqualified `powershell.exe`, any `pwsh`, ExecutionPolicy bypass, extra wrapper flag or fallback is permitted. This is the sole pre-launcher Docker exception and is not represented as a target test profile.
- [ ] Assert before and after that no SUT test ran, no cutover batch/state changed, no manifest was published, protected/dirty paths were untouched and only task-owned resources/reports were created.
- [ ] Independent reviewers inspect raw output without author conclusions. Missing raw artifact, arithmetic drift, cleanup residue, wrapper retry or any HIGH leaves G128-D non-passing.
- [ ] Seal the receipt and retire the exception. After G129-L, any invocation of this wrapper is a policy violation; all target work uses only `testframework/docker/run-task.ps1`.

<verify>
<automated>The literal-absolute no-argument wrapper call sealed by renewed G128-S is permitted only after G128-WPS51 and renewed G128-E/G128-S. Expected outcome is a reviewed G128-D receipt, not `hub.qa-runner/v1` PASS. Any PATH/unqualified/legacy/fallback interpreter, identity or sealed hardlink drift, extra flag, second consumed invocation, SUT/cutover/apply/publish action or missing independent review is non-acceptance.</automated>
</verify>

**Done:** one and only one G128-D run through the independently sealed absolute interpreter is accepted on the renewed revision, authority is still absent, and G129-L is unblocked without claiming target-runner readiness. The current prelaunch incident attempt is recorded as non-consumed; present status remains `NOT_READY` / runtime `NOT_RUN`.

**Atomic commit:** `docs(testframework): принять one-time Phase 128 guard`; stage only `.planning/audits/test-rebase/VERIFICATION.md`; body references TASK-204. Stop before staging on Git lock/process or dirty overlap.

## 9. Phase 129 — launcher, minimal slice and immutable cutover

### Task 129-01-A: Bootstrap and independently accept `hub.qa-runner/v1`

**Files:** Create `.planning/audits/test-rebase/python-package-legitimacy.md`, `testframework/{pyproject.toml,uv.lock,.python-version,README.md}`, `testframework/docker/{Dockerfile.core,run-task.ps1,images.lock.json}`, `testframework/schemas/run.schema.json`, `testframework/runner/{entrypoint.py,profiles.py}`, `testframework/policies/profiles.toml`, `testframework/selftests/runner/{test_launcher_contract.py,test_launcher_cleanup.py}`, `testframework/reports/.gitignore`, `.planning/audits/test-rebase/cutover-manifest.json`; Modify `.planning/audits/test-rebase/VERIFICATION.md` only for the G129-P receipt.

**Requirements:** `TASK-193`, `TASK-201`, `TASK-204`; approved spec §§3–6,11,12,14,15.

<read_first>
Approved architecture spec §§3–6,11,15; research report §§6,10,13,15.14–15.20; G-META-revised Phase 129 PLAN/129-01; `.cursor/rules/07-agent-safety.mdc`; `.cursor/playbooks/testing.md`; G128-D receipt.
</read_first>

**Consumes/Produces:** Exact package candidates as unqualified research inputs, G128-D builder/validator receipt and current Phase 129 CLI fields → `RunEnvelope`, `execute()`, `main()`, accepted outer CLI, immutable image/lock/profile registry and per-run evidence boundary; after G129-L plus G-CANON-X/G-META-X, the same launcher materializes the builder-identical planned authority and produces G129-P.

- [ ] Before any lock/install, create a package legitimacy table from official registries/upstream releases with name, version, owner, license, release/provenance URL, wheel availability for 3.14/3.13, transitive risk and `ACCEPTED|ASSUMED|SUS|REJECTED`. `[ASSUMED]`/`[SUS]` requires blocking human verification; `[SLOP]` is forbidden.
- [ ] Write launcher contract/cleanup tests first. Sentinels must fail for mutable tag, wrong platform digest, runtime download, missing/zero limits, `outer <= inner`, non-unique nonce, owner/verifier role reuse, report traversal/reparse/hardlink, production target, conclusion leakage, zero collection, unexpected skip, timeout, OOM/137, interrupted run and cleanup residue.
- [ ] Add `testframework/pyproject.toml` candidate pins and exact `testframework/uv.lock` only after package legitimacy and Python compatibility matrix are accepted in Docker. Select 3.14 or documented 3.13 fallback as one exact `testframework/.python-version`; do not maintain a second requirements authority.
- [ ] Implement launcher/runner state machine `VALIDATING → CREATED → PREFLIGHT → RUNNING → INSPECTING → CLEANING → FINISHED`, atomic first terminal event, bounded stream capture, central redaction, exact resource ID+name+nonce+labels cleanup, and raw+sealed receipts.
- [ ] Seal exactly seven bootstrap registry entries in `profiles.toml`: `launcher-bootstrap` (`core`, inner/outer `1500/1800`), `canon-routing-contract` (`core`, `300/420`), `downstream-plan-contract` (`core`, `300/420`), `cutover-authority` (`core`, `720/900`), `framework` (`core`, `600/720`), `api-health` (`core`, `600/720`) and `cutover-preflight` (`core`, `720/900`). Every entry binds a literal target, non-zero CPU/RAM/PID/shm/nofile/workers and the digest-bearing `core` image key; `profiles.py` rejects unknown/duplicate IDs and missing image/limit/target/deadline fields. Phase 130-A is the sole later registry expansion.
- [ ] Bootstrap acceptance uses no cutover mutation and no missing manifest fallback. It proves success/failure/timeout/interruption cleanup and that source/lock mounts are read-only while only owner-marked reports are writable.
- [ ] After independent G129-L, first run the exact G-CANON-X and G-META-X profiles and require semantic equality with their versioned receipts. Only after both rechecks, invoke `authority-materialize` on the same sealed revision. It runs the Phase 128 builder twice in-container, requires byte identity and sealed source/tool hashes, writes `.planning/audits/test-rebase/cutover-manifest.json` atomically with no-replace/no-follow semantics, then runs the independent validator. Existing different bytes, partial stage, reparse/hardlink, drift, validation error or `173/1,120/max56` mismatch leaves no authority change.
- [ ] Review G129-P as an authority transition only: all `173` batches remain `planned`, no operation/consumer state advances, and the legacy Phase 128 wrapper is not invoked.

<verify>
<automated>Launcher acceptance, after G128-D: `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -NoProfile -NonInteractive -File testframework/docker/run-task.ps1 -ContractVersion hub.qa-runner/v1 -PlanId 129-01 -TaskRunId $HubTaskRunId -TaskId TASK-201 -OwnerRole devops-engineer -VerifierRoles qa-auto-ver -Mode bootstrap-acceptance -Profile launcher-bootstrap -ImageLock testframework/docker/images.lock.json -InnerTimeoutSeconds 1500 -TimeoutSeconds 1800`. Before its static prerequisites this is `NOT_READY`; there is no fallback. Authority materialization uses a fresh TaskRunId, `TaskId=TASK-204`, `Mode=authority-materialize`, `Profile=cutover-authority`, `-Manifest .planning/audits/test-rebase/cutover-manifest.json`, inner/outer `720/900`, through the same accepted launcher and remains `NOT_READY` until G129-L, G-CANON-X and G-META-X.</automated>
</verify>

**Done:** G129-L records accepted immutable provenance and zero owned leftovers; G-CANON-X/G-META-X reproduce both versioned static receipts; then G129-P independently confirms one builder-identical planned manifest with `173/1,120/max56`, zero batch mutations and no historical-wrapper reuse.

**Atomic commits:** (1) `feat(testframework): добавить безопасный Docker launcher`, stage launcher/package/schema/selftest/report-boundary files plus `python-package-legitimacy.md`, body TASK-193/TASK-201; (2) only after G129-P, `docs(testframework): материализовать cutover authority`, stage only `.planning/audits/test-rebase/cutover-manifest.json` and `.planning/audits/test-rebase/VERIFICATION.md`, body TASK-204.

### Task 129-01-B: Build config/runtime/traceability contracts

**Files:** Create `testframework/src/hub_testkit/__init__.py`, `testframework/src/hub_testkit/config/{__init__.py,models.py,load.py}`, `testframework/src/hub_testkit/runtime/{__init__.py,ids.py,ownership.py,redaction.py}`, `testframework/src/hub_testkit/traceability/{__init__.py,model.py,loader.py,index.py}`, `testframework/selftests/runtime/{test_ids.py,test_ownership.py,test_redaction.py}`, `testframework/selftests/traceability/{test_loader.py,test_index.py}`, `testframework/schemas/test-case.schema.json`, `testframework/policies/{markers.toml,environments.toml,redaction.toml}`.

<read_first>
First-slice signatures in this plan §4; approved spec §§4,5,6,9,10,14; accepted launcher evidence; authoritative requirements/ID conventions in `docs/GLOSSARY.md`.
</read_first>

**Consumes/Produces:** Accepted Python/lock and `RunEnvelope` IDs → exact signatures in §4, schema-valid TC loader/index, positive environment allowlist, production deny and central redaction policy.

- [ ] Write failing selftests for invalid/colliding IDs, Unicode/case aliases, unknown config keys, negative/non-unique nonce, foreign ownership cleanup, duplicate/missing/orphan TC, unknown case fields, wrong path/layer, expired exception and secret/PII canaries.
- [ ] Implement immutable models and pure functions. Config does no network I/O; ownership deletes nothing; it only authorizes exact scope and records durable data. Loader accepts UTF-8 Markdown front matter within size/path budgets and never executes YAML tags.
- [ ] Make `markers.toml` the sole marker registry data; Phase 129 registers only `case_id` and `api` required by health. Phase 130 adds public-hook enforcement without moving the source of truth.
- [ ] Run framework profile twice with same seed/config and compare semantic outputs; inject the temporary redaction/ownership/duplicate-TC sentinels and require expected failure before reverting each sentinel.

<verify>
<automated>`PlanId=129-01`, `TaskId=TASK-193`, `OwnerRole=qa-auto1`, `VerifierRoles=qa-auto-ver`, `Mode=verify`, `Profile=framework`, `InnerTimeoutSeconds=600`, `TimeoutSeconds=720` through the canonical command in §7. NOT_READY until G129-L.</automated>
</verify>

**Done:** config/runtime/traceability selftests collect non-zero, produce deterministic results and reject every named invalid boundary without secret leakage.

**Atomic commit:** `feat(testframework): зафиксировать runtime и traceability contracts`; exact allowlist is the files above; body references TASK-193.

### Task 129-01-C: Deliver the health API tracer and final preflight

**Files:** Create `testframework/src/hub_testkit/domains/health/{__init__.py,api.py,models.py,oracles.py}`, `testframework/tests/conftest.py`, `testframework/tests/api/health/test_health.py`, `testframework/cases/health/TC-HEALTH-0001.md`; Modify `testframework/README.md`.

<read_first>
Approved spec §§5,7–10,14; exact health endpoint contract in current Hub docs/OpenAPI; accepted G129-L/G129-P evidence; Phase 129 Task 2/3 preflight requirements.
</read_first>

**Consumes/Produces:** `HubSettings`, `RuntimeIds`, `Redactor`, `HealthApi` signatures and `TC-HEALTH-0001` → one real `@pytest.mark.api` + `@pytest.mark.case_id("TC-HEALTH-0001")` path through launcher → runner → config → HTTPX → strict wire model → pure oracle → JUnit/traceability evidence.

- [ ] First write health case, schema-valid metadata and failing API test. Sentinels: wrong status, malformed/extra-field body, production target, secret canary in response/log and unreachable endpoint must fail with safe operation/context/next-action diagnostics.
- [ ] Implement only health domain files; `testframework/tests/conftest.py` composes `httpx.Client` and `HealthApi` without global autouse setup or scenario logic. Preserve `HealthApi.fetch()` as Phase 131 public contract.
- [ ] Execute health tracer; then independently run authority `preflight-readonly` and `preflight-verify` modes. Both mount cutover authority read-only, reproduce `173/1,120/max56`, run executor sentinels and write no manifest/source changes.
- [ ] Update README with one canonical launcher example, architecture map, case flow, triage and `NOT_READY` semantics; no direct inner commands.

<verify>
<automated>Health: `PlanId=129-01`, `TaskId=TASK-193`, `OwnerRole=qa-auto1`, `VerifierRoles=qa-auto-ver`, `Profile=api-health`, inner/outer `600/720`. Preflight: exact Phase 129 modes with `TaskId=TASK-204`, `OwnerRole=arch`, `VerifierRoles=arch-ver,qa-auto-ver`, `-Manifest .planning/audits/test-rebase/cutover-manifest.json`, profile `cutover-preflight`, inner/outer `720/900`. All calls go through accepted `run-task.ps1`.</automated>
</verify>

**Done:** a real user-visible health behavior works end-to-end; both conclusion-blind preflight receipts have zero HIGH and exact cleanup; batches remain unexecuted.

**Atomic commit:** `test(testframework): добавить health API tracer`; allowlist is exactly the health/test/case/README files; body references TASK-193/TASK-204.

### Tasks 129-02 through 129-07: Execute the sealed 173 batch tasks

**Files:** Exact `operations[].source/target`, `affected_consumers[].path` and per-batch manifest/evidence fields from the G129-P manifest; no inferred glob is writable.

**Requirements:** `TASK-197`, `TASK-198`, `TASK-219`, `TASK-234` with secondary `TASK-201` only on the existing consumer/CI closure batches.

<read_first>
The one corresponding task in `.planning/phases/129-tf-e2e-shared/129-02-PLAN.md` … `.planning/phases/129-tf-e2e-shared/129-07-PLAN.md`; G129-P manifest; both Phase 129 preflight receipts; current dirty status/owner map; affected consumer docs/CI file.
</read_first>

**Consumes/Produces:** One unique sealed batch ID, literal operation count, owned-path hash and previous verified transition → one owner mutation commit, rollback handle, fresh conclusion-blind verifier receipt and next continuous manifest transition.

Для **каждого** из 173 tasks независимо:

- [ ] Assert current HEAD ancestry, clean allowlist overlap, exact batch state/count/path hash, source identity/no-follow/link-count one, target absence, protected prefix exclusion and previous receipt continuity.
- [ ] Use that task's existing exact `-BatchId` and `-ExpectedOperations` from Phase 129 numbered plans. Do not loop over IDs, regroup operations, change order or recalculate paths.
- [ ] Launcher internally starts one authorized Linux writer, applies one batch, stages only its allowlist, creates one atomic owner commit/rollback handle, then starts a fresh read-only verifier with no writer conclusion.
- [ ] Verifier runs independent root manifest check plus focused native unit/consumer/zero-hook gate. Any code/test/infra failure is triaged; deleted component tests are never restored to recover coverage.
- [ ] Cleanup exact resources and transaction before advancing. Failure/timeout/cleanup ambiguity stops the chain; recovery is the recorded single-commit `git revert` followed by a new nonce-bearing rollback verification run.

<verify>
<automated>Use the exact canonical `apply-and-verify` command already embedded in the matching 129 task with accepted `hub.qa-runner/v1`; no second owner invocation. Expected after the final task: 173 unique verified transitions, operation total 1,120, max 56, moved native units green, legacy consumers zero, 43 hook files deleted only after zero-consumer proof, cleanup PASS.</automated>
</verify>

**Done:** G129-C passes; current repository contains no legacy executable non-unit/support closure and the minimal Python slice remains green.

**Atomic commits:** exactly one mutation commit per sealed batch using its existing Russian Conventional Commit subject/body TASK mapping; verifier evidence commit is narrow and never a second mutation commit.

## 10. Phase 130 — central profile registry, traceability and architecture gates

### Task 130-A: Seal the central later-profile registry

**Files:** Modify `testframework/runner/profiles.py`, `testframework/policies/profiles.toml`, `testframework/docker/images.lock.json`; Create `testframework/selftests/runner/test_profile_registry.py`, `testframework/selftests/docs/test_docs_contract.py`.

**Requirements:** `TASK-206`, `TASK-218`; approved spec §§4–6,9–12,14,15.

<read_first>Accepted G129-L seven-entry bootstrap registry; every Phase 130–141 verify contract below; approved image/runner contracts; immutable `hub.qa-runner/v1` schema.</read_first>

**Consumes/Produces:** Seven G129-L bootstrap entries and digest-bearing image authority → one typed central registry with 47 exact later-profile rows, six exact limit classes, deterministic target resolution and `UnknownProfileError(profile_id: str)`; `resolve_profile(profile_id: str, image_lock: ImageLock) -> ResolvedProfile` returns `NOT_READY` on missing future image/target and fails closed on an unknown/duplicate/drifted entry. The 47 rows include reused `framework` and `api-health`; union with the bootstrap-only entries is 52 unique IDs.

Limit classes are literal registry data, not environment defaults:

| Class | CPU | RAM | swap | PID | shm | nofile | workers |
|---|---:|---:|---:|---:|---:|---:|---:|
| `S` | `1.0` | `1024 MiB` | `0` | `128` | `64 MiB` | `1024` | `1` |
| `C` | `2.0` | `2048 MiB` | `0` | `256` | `128 MiB` | `2048` | `1` |
| `I` | `2.0` | `3072 MiB` | `0` | `384` | `256 MiB` | `4096` | `1` |
| `B` | `2.0` | `4096 MiB` | `0` | `512` | `1024 MiB` | `4096` | `1` |
| `L` | `4.0` | `6144 MiB` | `0` | `512` | `256 MiB` | `8192` | `4` |
| `F` | `4.0` | `8192 MiB` | `0` | `768` | `1024 MiB` | `8192` | `1` |

Every later profile is registered exactly once:

| ID | Image key | Class | inner/outer | Literal target |
|---|---|---|---|---|
| `framework` | `core` | `C` | `600/720` | recursive directory target `testframework/selftests/` with zero-collection failure |
| `docs-contract` | `core` | `S` | `300/420` | `testframework/selftests/docs/test_docs_contract.py` |
| `api-health` | `core` | `C` | `600/720` | `testframework/tests/api/health/test_health.py` |
| `api-identity` | `core` | `I` | `900/1080` | `testframework/tests/api/identity/test_session.py`, `testframework/tests/api/identity/test_account.py` |
| `api-security` | `core` | `I` | `900/1080` | `testframework/tests/api/identity/test_authorization_matrix.py`, `testframework/tests/api/identity/test_abuse.py` |
| `api-security-full` | `core` | `I` | `1200/1440` | `testframework/tests/api/security/test_platform_closure.py` |
| `integration-pg-mongo` | `core` | `I` | `1200/1440` | `testframework/tests/integration/postgres/test_transactions.py`, `testframework/tests/integration/mongo/test_transactions.py` |
| `integration-redis-nats` | `core` | `I` | `1200/1440` | `testframework/tests/integration/redis/test_namespace_cleanup.py`, `testframework/tests/integration/nats/test_delivery_recovery.py` |
| `api-tracker` | `core` | `C` | `600/720` | `testframework/tests/api/tracker/test_issue_lifecycle.py` |
| `api-project` | `core` | `C` | `600/720` | `testframework/tests/api/project/test_project_scope.py` |
| `api-sprint` | `core` | `C` | `600/720` | `testframework/tests/api/sprint/test_state_conflict.py` |
| `api-tms` | `core` | `C` | `600/720` | `testframework/tests/api/tms/test_run_lifecycle.py` |
| `api-messenger` | `core` | `C` | `600/720` | `testframework/tests/api/messenger/test_delivery.py` |
| `integration-notification-nats` | `core` | `I` | `900/1080` | `testframework/tests/api/notification/test_delivery.py`, `testframework/tests/integration/nats/test_reconnect_ordering.py` |
| `api-realtime-closure` | `core` | `I` | `900/1080` | `testframework/tests/api/messenger/test_cross_domain_closure.py`, `testframework/selftests/traceability/test_realtime_coverage.py` |
| `api-calendar` | `core` | `C` | `600/720` | `testframework/tests/api/calendar/test_calendar_requirements.py` |
| `api-intranet` | `core` | `C` | `600/720` | `testframework/tests/api/intranet/test_intranet_requirements.py` |
| `api-accounting` | `core` | `C` | `600/720` | `testframework/tests/api/accounting/test_accounting_requirements.py` |
| `api-crm` | `core` | `C` | `600/720` | `testframework/tests/api/crm/test_crm_requirements.py` |
| `api-hr` | `core` | `C` | `600/720` | `testframework/tests/api/hr/test_hr_requirements.py` |
| `api-bases` | `core` | `C` | `600/720` | `testframework/tests/api/bases/test_bases_requirements.py` |
| `api-dashboard` | `core` | `C` | `600/720` | `testframework/tests/api/dashboard/test_dashboard_requirements.py` |
| `integration-storage` | `core` | `I` | `900/1080` | `testframework/tests/integration/storage/test_failure_recovery.py` |
| `integration-github` | `core` | `I` | `900/1080` | `testframework/tests/integration/github/test_failure_recovery.py` |
| `integration-gitlab` | `core` | `I` | `900/1080` | `testframework/tests/integration/gitlab/test_failure_recovery.py` |
| `integration-onec` | `core` | `I` | `900/1080` | `testframework/tests/integration/onec/test_failure_recovery.py` |
| `integration-mail` | `core` | `I` | `900/1080` | `testframework/tests/integration/mail/test_failure_recovery.py` |
| `api-ai` | `core` | `C` | `600/720` | `testframework/tests/api/ai/test_ai_requirements.py` |
| `api-search` | `core` | `C` | `600/720` | `testframework/tests/api/search/test_search_requirements.py` |
| `api-backup` | `core` | `C` | `600/720` | `testframework/tests/api/backup/test_backup_requirements.py` |
| `api-imex` | `core` | `C` | `600/720` | `testframework/tests/api/imex/test_imex_requirements.py` |
| `api-marketplace-closure` | `core` | `I` | `900/1080` | `testframework/tests/api/marketplace/test_marketplace_requirements.py`, `testframework/selftests/traceability/test_business_integration_coverage.py` |
| `contract` | `core` | `I` | `900/1080` | `testframework/tests/contract/api/test_openapi.py`, `testframework/tests/contract/events/test_envelopes.py`, `testframework/tests/contract/schema/test_migrations.py` |
| `contract-frontend` | `core` | `I` | `900/1080` | `testframework/tests/contract/frontend/test_routes.py`, `testframework/tests/contract/frontend/test_i18n.py`, `testframework/tests/contract/frontend/test_accessibility_manifest.py`, `testframework/tests/contract/frontend/test_docs_links.py` |
| `browser-framework` | `e2e` | `B` | `1200/1440` | `testframework/selftests/browser/test_lifecycle.py`, `testframework/selftests/browser/test_artifacts.py` |
| `e2e-smoke` | `e2e` | `B` | `900/1080` | `testframework/tests/e2e/identity/test_login_smoke.py` |
| `e2e-identity-revocation` | `e2e` | `B` | `900/1080` | `testframework/tests/e2e/identity/test_session_revocation.py` |
| `e2e-tracker-conflict` | `e2e` | `B` | `1200/1440` | `testframework/tests/e2e/tracker/test_realtime_conflict.py` |
| `e2e-messenger-attachment` | `e2e` | `B` | `1200/1440` | `testframework/tests/e2e/messenger/test_attachment_recovery.py` |
| `e2e-recovery` | `e2e` | `B` | `1800/2100` | `testframework/tests/e2e/recovery/test_process_restart.py`, `testframework/tests/e2e/recovery/test_dependency_timeout.py`, `testframework/tests/e2e/recovery/test_multi_user_isolation.py` |
| `e2e-a11y` | `e2e` | `B` | `1200/1440` | `testframework/tests/e2e/accessibility/test_login.py`, `testframework/tests/e2e/accessibility/test_tracker.py`, `testframework/tests/e2e/accessibility/test_messenger.py` |
| `security-privacy` | `core` | `I` | `1200/1440` | `testframework/tests/api/security/test_cross_cutting.py`, `testframework/tests/contract/security/test_artifact_redaction.py` |
| `load-validity` | `load` | `L` | `1200/1440` | `testframework/selftests/load/test_target_guard.py`, `testframework/selftests/load/test_measurement_validity.py` |
| `load-steady` | `load` | `L` | `3600/3900` | `testframework/load/locustfile.py`, `testframework/load/profiles/steady.toml` |
| `ci-contract` | `core` | `I` | `900/1080` | `testframework/selftests/ci/test_ci_contract.py` |
| `clean-room-full` | `core,e2e,load` | `F` | `7200/7500` | `testframework/policies/profiles.toml#all-ready-nonmanual` |
| `onboarding-clean-room` | `core` | `I` | `1200/1440` | `testframework/selftests/docs/test_docs_contract.py`, `testframework/README.md` |

- [ ] Write registry selftests first: duplicate/missing/unknown ID, zero/negative limit, mutable or absent image digest, absent target, target outside `testframework/**`, profile/CLI deadline mismatch, `outer <= inner`, illegal worker count, load target on core and browser target without private shm all fail before container creation.
- [ ] Implement bounded TOML parsing with exact keys, no environment/CLI override, deterministic ordering/hash and typed target kinds. Future `e2e`/`load` entries remain registered but resolve `NOT_READY` until their phase seals the matching image digest and target files; they never fall back to `core`.
- [ ] Make `test_docs_contract.py` consume a literal per-phase docs-closure receipt and reject an undocumented changed task, docs path absent from closure `Files`, extra staged docs path, stale TASK/TC link or profile ID not in the registry.

<verify><automated>`PlanId=130-01`, `TaskId=TASK-218`, owner `qa-auto2`, verifiers `qa-auto-ver,arch-ver,security-auditor`, profile `framework`, inner/outer `600/720`; receipt `testframework/reports/<task-run>/TASK-218-profile-registry-receipt.json` records 47/47 later rows, 52 unique total IDs, exact registry hash, image readiness, limit/target/deadline coverage and every unknown-profile sentinel.</automated></verify>

**Done:** every later verify profile resolves to exactly one immutable row or explicit `NOT_READY`; no unknown/duplicate ID, implicit default, target glob escape, zero limit or later `profiles.py` writer exists.

**Atomic commit:** `feat(testframework): закрепить центральный profile registry`; stage exactly the five Files paths above; body TASK-218.

### Task 130-B: Move collection policy into a public-hook pytest plugin

**Files:** Create `testframework/src/hub_testkit/pytest_plugin/{__init__.py,plugin.py,markers.py,collection.py,reporting.py}`, `testframework/selftests/pytest_plugin/{test_markers.py,test_collection.py,test_reporting.py}`; Modify `testframework/pyproject.toml`, `testframework/tests/conftest.py`, `testframework/policies/markers.toml`, `testframework/schemas/test-case.schema.json`.

**Requirements:** `TASK-206`; approved spec §§4,9,10,12,14.

<read_first>Approved spec §§4,9,10,14; pytest official public-hook docs pinned by package audit; Phase 129 traceability signatures/evidence; accepted Task 130-A registry; `docs/GLOSSARY.md` ID conventions.</read_first>

**Consumes/Produces:** `CaseIndex`, `CollectedCase`, `Violation` and accepted `framework` profile → plugin entry `hub_testkit.pytest_plugin.plugin`, exactly one layer + one `case_id` per `testframework/tests/**` node, JUnit properties projected from case metadata, and technical-sentinel exemption enum.

- [ ] Write `pytester` failures first: unknown marker, no/multiple case IDs, path/layer mismatch, duplicate TC, orphan automated case, collected planned/manual/retired case, ad hoc skip/xfail, expired quarantine/allow-skip and missing capability.
- [ ] Implement only public hooks; collection performs no SUT I/O. `testframework/tests/conftest.py` becomes 5–20 lines of explicit plugin/composition registration.
- [ ] Add marker compatibility rules from approved spec and prove zero collection preserves pytest exit 5 as failure.

<verify><automated>Canonical §7 invocation: `PlanId=130-02`, `TaskId=TASK-206`, owner `qa-auto1`, verifier `qa-auto-ver,ba-ver`, profile `framework`, inner/outer `600/720`. Expected: serial plugin selftests green and every named sentinel red.</automated></verify>

**Done:** public-hook plugin rejects every named collection/traceability violation, preserves zero-collection failure, projects one schema-valid TC chain per SUT test, performs no collection-time SUT I/O and passes its non-zero serial selftests.

**Atomic commit:** `feat(testframework): включить fail-closed traceability plugin`; stage exactly the twelve expanded Files paths above; body TASK-206. Shared docs are owned only by Task 130-D.

### Task 130-C: Enforce architecture contracts

**Files:** Create `testframework/selftests/architecture/{test_import_contracts.py,test_forbidden_patterns.py}`, `testframework/policies/licenses.toml`; Modify `testframework/pyproject.toml`, `testframework/policies/markers.toml`.

**Requirements:** `TASK-218`; approved spec §§4,10,11,12,14,15.

<read_first>Approved spec §§4,10,11,14,15; research report §§5,6.3,7.4,13,15.15; accepted Tasks 130-A/B.</read_first>

**Consumes/Produces:** Accepted package/marker/profile registries → Import Linter layer/independence contracts and reviewable forbidden-pattern allowlist; this task does not write `testframework/runner/profiles.py` or image limits.

- [ ] First make architecture selftests fail on injected `utils.py`, `BaseTest`, singleton, `sys.path`, `_pytest`, domain-to-domain import, builder/oracle I/O import, naked sleep, unbounded retry and test-to-test import.
- [ ] Keep `full` a registered profile, `load` outside pytest layers and impact selection unable to exclude critical/security/authz sentinels.
- [ ] Run serial framework twice, persisted-seed replay and fresh seed. xdist remains disabled until its Phase 141 calibration spike.

<verify><automated>`PlanId=130-03`, `TaskId=TASK-218`, owner `qa-auto2`, verifier `qa-auto-ver,arch-ver`, profile `framework`, inner/outer `600/720`. Expected GF acceptance only when architecture, marker, collection, replay and cleanup gates all pass.</automated></verify>

**Done:** architecture selftests fail on every forbidden import/pattern sentinel, repeat semantically under persisted seed, keep xdist disabled and yield GF acceptance only with zero owned leftovers.

**Atomic commit:** `test(testframework): закрепить architecture gates`; stage exactly the five expanded Files paths above; body TASK-218. Shared docs are owned only by Task 130-D.

### Task 130-D: Synchronize Phase 130 docs and traceability

**Files:** Modify `testframework/README.md`, `docs/test_plan.md`, `docs/stack_and_tools.md`, `.planning/COVERAGE.md`.

**Requirements:** `TASK-206`, `TASK-218`; approved spec §§9,10,12,14,15.

<read_first>Accepted receipts from Tasks 130-A–C; exact four docs files above; current TASK/TC/COVERAGE registries.</read_first>

**Consumes/Produces:** Accepted profile/plugin/architecture receipts → one literal Phase 130 docs diff and `TASK-206/TASK-218` docs-closure receipt; no production or framework code changes.

**Action/sentinel:** Update only the four Files paths with the 52-ID registry model, authoring/collection rules, architecture/profile ownership and current evidence links. The docs-contract profile must fail if any Task 130 receipt, registry hash or exact path is omitted, or if an unlisted docs path is staged.

<verify><automated>`PlanId=130-04`, `TaskId=TASK-206,TASK-218`, owner `docs-writer`, verifiers `qa-auto-ver,arch-ver`, profile `docs-contract`, inner/outer `300/420`; receipt `testframework/reports/<task-run>/TASK-206-218-docs-sync-receipt.json`.</automated></verify>

**Done:** all Phase 130 receipts and exact profile/traceability contracts are discoverable from the four docs, zero stale profile ID remains, the phase cannot complete without the docs receipt, and only the independent G130-H stop receipt may close the current implementation run.

**Atomic commit:** `docs(testframework): синхронизировать Phase 130 contracts`; stage exactly the four Files paths above; body TASK-206/TASK-218.

### G130-H execution stop and backlog activation contract

After Task 130-D, `qa-auto-ver`, `arch-ver` and `security-auditor` independently review the exact G129-C cutover receipt, G129-L launcher provenance, Task 130-A registry receipt, GF selftests, unchanged `TC-HEALTH-0001` evidence and Phase 130 docs receipt. Their bounded `testframework/reports/<task-run>/G130-H-health-framework-handoff.json` must record `current_scope=complete`, `activated_backlog_phases=[]`, zero owned leftovers and separate `NOT_READY` status for every Phase 131–141 profile/phase. This is a blocking execution checkpoint: stop the implementation session after its review; do not continue by interpreting the backlog map as the next wave.

Each Phase 131–141 task below inherits precondition `G-ACTIVATE-<phase>` on the exact activation HEAD. The founder receipt must name one phase/feature slice, accepted requirement/risk IDs, literal target tasks/files, package/image changes, rollback owner and evidence budget. The corresponding live `PLAN.md` is re-run through G-META-X before execution. Absence, drift or broader selection returns `NOT_READY`; a receipt for one phase does not activate another. Pre-registration in Task 130-A reserves identifiers and fail-closed limits only—it grants no runtime readiness.

## 11. Phase 131 — typed HTTP/runtime boundaries and health/auth/account pilot

### Task 131-A: Extract the accepted typed HTTP boundary without breaking health

**Files:** Create `testframework/src/hub_testkit/http/{__init__.py,transport.py,auth.py,policy.py,errors.py,diagnostics.py}`, `testframework/selftests/http/{test_policy.py,test_diagnostics.py}`; Modify `testframework/src/hub_testkit/domains/health/api.py`.

**Requirements:** `TASK-205`; approved spec §§4,6,7,11,12,14.

<read_first>Approved spec §§4,6,11,14; research HTTP/policy/diagnostics examples; Phase 129 health tracer; authoritative API error/timeout docs; accepted central profile registry.</read_first>

**Consumes/Produces:** Phase 129 health facade and authoritative HTTP behavior → `Actor`, `build_client(settings, actor) -> httpx.Client`, `RequestPolicy` and sanitized diagnostics; preserves `HealthApi.fetch()`.

**Action/sentinel:** First fail on missing connect/read/write/pool timeout, unsafe POST retry without idempotency, secret-bearing diagnostics and copied production oracle. Implement one actor per client, bounded retry classification and correlation with no pytest import; re-run `TC-HEALTH-0001` unchanged.

<verify><automated>`PlanId=131-01`, `TaskId=TASK-205`, owner `qa-auto1`, verifier `qa-auto-ver,security-auditor`, profile `api-health`, inner/outer `600/720`; receipt `testframework/reports/<task-run>/TASK-205-http-boundary-receipt.json`.</automated></verify>

**Done:** typed HTTP policy rejects unsafe retry/redaction sentinels, `TC-HEALTH-0001` remains green through `HealthApi.fetch()` and diagnostics expose no canary value.

**Atomic commit:** `refactor(testframework): вынести typed HTTP boundary`; stage exactly the nine expanded Files paths above; body TASK-205. Shared docs are owned only by Task 131-D.

### Task 131-B: Add monotonic runtime and observability boundaries

**Files:** Create `testframework/src/hub_testkit/runtime/{clock.py,polling.py}`, `testframework/src/hub_testkit/observability/{__init__.py,correlation.py,manifest.py}`, `testframework/selftests/runtime/{test_clock.py,test_polling.py}`.

**Requirements:** `TASK-205`; approved spec §§4,6,11,12,14.

<read_first>Accepted Task 131-A HTTP boundary; launcher first-exit/report manifest; approved polling/deadline/redaction contracts.</read_first>

**Consumes/Produces:** `RequestPolicy`, runner monotonic clock and redactor → `PollDeadlineExceeded`, `wait_until(probe, done, *, deadline_s, clock, sleeper)`, correlation context and sealed run-manifest projection.

**Action/sentinel:** First fail on wall-clock deadlines, unbounded polling, hidden fixed sleep, correlation loss and secret-bearing manifest fields. Implement injected monotonic clock/sleeper and bounded last-observation diagnostics; no SUT I/O in selftests.

<verify><automated>`PlanId=131-02`, `TaskId=TASK-205`, owner `qa-auto1`, verifiers `qa-auto-ver,security-auditor`, profile `framework`, inner/outer `600/720`; receipt `testframework/reports/<task-run>/TASK-205-runtime-observability-receipt.json`.</automated></verify>

**Done:** deadline/polling/correlation/manifest selftests reject every named sentinel, preserve first cause and never expose the secret canary.

**Atomic commit:** `feat(testframework): добавить monotonic runtime observability`; stage exactly the seven expanded Files paths above; body TASK-205. Shared docs are owned only by Task 131-D.

### Task 131-C: Add auth/session/account cases from accepted requirements

**Files:** Create `testframework/src/hub_testkit/domains/identity/{__init__.py,models.py,api.py,builders.py,oracles.py}`, `testframework/tests/api/identity/{test_session.py,test_account.py}`, `testframework/cases/identity/{TC-IDENTITY-0001.md,TC-IDENTITY-0002.md,TC-IDENTITY-0003.md,TC-IDENTITY-0004.md}`.

**Requirements:** `TASK-207`; approved spec §§4,7–9,12,14.

<read_first>Verified product/SA auth-session-account requirements; current OpenAPI; security guidelines; accepted Tasks 131-A/B. If expected behavior is ambiguous, stop before automation and return to product/SA acceptance.</read_first>

**Consumes/Produces:** Accepted HTTP/runtime/observability boundaries and identity requirements → immutable `Credentials`, `Session`, `Account`, `IdentityApi`, seeded `IdentityBuilder` and pure semantic oracles. Case allocation: valid session, invalid credentials/no token, blocked account/no new session, account tenant isolation.

- [ ] Write schema-valid cases and failing tests first; each checks positive observable result and forbidden side effects (no token/session/audit secret/foreign data).
- [ ] Implement strict wire validation before typed models; fixture owns one state-changing resource and immediately registers exact cleanup.
- [ ] Run cases individually, as a group, in reversed/randomized order and after forced client/process interruption.

<verify><automated>`PlanId=131-03`, `TaskId=TASK-207`, owner `qa-auto2`, verifier `qa-auto-ver,security-auditor`, profile `api-identity`, inner/outer `900/1080`; receipt `testframework/reports/<task-run>/TASK-207-api-identity-receipt.json`.</automated></verify>

**Done:** all four identity cases have schema-valid TASK/REQ/TC links, sensitive positive/negative oracles, deterministic order/replay evidence and exact cleanup after normal and interrupted execution.

**Atomic commit:** `test(testframework): добавить auth session account pilot`; stage exactly the eleven expanded Files paths above; body TASK-207. Shared docs are owned only by Task 131-D.

### Task 131-D: Synchronize Phase 131 docs and traceability

**Files:** Create `testframework/docs/architecture.md`; Modify `testframework/README.md`, `docs/test_plan.md`, `.planning/COVERAGE.md`.

**Requirements:** `TASK-205`, `TASK-207`; approved spec §§4,7–9,11,12,14.

<read_first>Accepted receipts from Tasks 131-A–C; exact four docs paths above; current health/identity TC index.</read_first>

**Consumes/Produces:** Accepted HTTP/runtime/identity receipts → exact architecture, authoring and coverage links plus one Phase 131 docs-closure receipt.

**Action/sentinel:** Document the real typed boundaries and four accepted identity cases. `docs-contract` fails on a missing receipt/TC/profile, an obsolete direct-layer path or any staged docs path outside the four Files.

<verify><automated>`PlanId=131-04`, `TaskId=TASK-205,TASK-207`, owner `docs-writer`, verifiers `qa-auto-ver,arch-ver`, profile `docs-contract`, inner/outer `300/420`; receipt `testframework/reports/<task-run>/TASK-205-207-docs-sync-receipt.json`.</automated></verify>

**Done:** architecture/README/test-plan/COVERAGE agree with accepted code and TC evidence; Phase 131 completion is blocked without the docs receipt.

**Atomic commit:** `docs(testframework): синхронизировать Phase 131 boundaries`; stage exactly the four Files paths above; body TASK-205/TASK-207.

## 12. Phase 132 — authz, tenant isolation and abuse

### Task 132-A: Role, admin and tenant denial matrix

**Files:** Create `testframework/tests/api/identity/test_authorization_matrix.py`, `testframework/cases/identity/{TC-IDENTITY-0101.md,TC-IDENTITY-0102.md,TC-IDENTITY-0103.md}`; Modify `testframework/src/hub_testkit/domains/identity/{models.py,api.py,builders.py,oracles.py}`.

**Requirements:** `TASK-208`; approved spec §§4,7–9,12–14.

<read_first>Accepted RBAC/tenant/admin requirements, auth ADRs, current API, Phase 131 identity facade.</read_first>

**Consumes/Produces:** Accepted identity transport and security requirements → `RoleAssignment`, `TenantScope` and explicit deny oracles without changing transport. Cases cover role-state decision table, cross-tenant object, and admin boundary.

**Action/sentinel:** Write denial tests first; temporary removal of role/tenant check must make them red. Test no foreign data/audit secret side effects and exact cleanup.

<verify><automated>`PlanId=132-01`, `TaskId=TASK-208`, owner `qa-auto1`, verifier `qa-auto-ver,security-auditor`, profile `api-security`, inner/outer `900/1080`.</automated></verify>

**Done:** role/admin/tenant decision cases deny every unauthorized variant without foreign-data or secret side effects, and the deliberate policy-weakening sentinel is observed red.

**Atomic commit:** `test(testframework): покрыть role tenant denial matrix`; stage exactly the eight expanded Files paths above; body TASK-208. Shared docs are owned only by Task 132-D.

### Task 132-B: Abuse, enumeration, replay and rate-limit behavior

**Files:** Create `testframework/src/hub_testkit/domains/identity/abuse.py`, `testframework/tests/api/identity/test_abuse.py`, `testframework/cases/identity/{TC-IDENTITY-0201.md,TC-IDENTITY-0202.md,TC-IDENTITY-0203.md,TC-IDENTITY-0204.md}`.

**Requirements:** `TASK-209`; approved spec §§4,6–9,12–14.

<read_first>Threat model/security/privacy requirements; rate-limit/session/replay contracts; Task 132-A.</read_first>

**Consumes/Produces:** Accepted identity/security facade and abuse requirements → pure request-series builder and response-timeline oracle; cases cover IDOR/escalation, enumeration-safe response, replay rejection and bounded rate-limit recovery.

**Action/sentinel:** Write tests before helper; avoid sleeps by injected clock/bounded polling. Sentinels weaken status/problem type/retry window/redaction one at a time and must fail.

<verify><automated>`PlanId=132-02`, `TaskId=TASK-209`, owner `qa-auto2`, verifier `security-auditor,qa-auto-ver`, profile `api-security`, inner/outer `900/1080`.</automated></verify>

**Done:** IDOR/escalation, enumeration, replay and rate-limit recovery cases each prove observable denial/recovery with injected-clock determinism, safe diagnostics and sensitivity to the named weakened-oracle sentinels.

**Atomic commit:** `test(testframework): добавить abuse и replay cases`; stage exactly the six expanded Files paths above; body TASK-209. Shared docs are owned only by Task 132-D.

### Task 132-C: Close the cross-domain security/platform risk graph

**Files:** Create `testframework/tests/api/security/test_platform_closure.py`, `testframework/cases/security/{TC-SECURITY-0001.md,TC-SECURITY-0002.md}`, `testframework/selftests/traceability/test_critical_risk_coverage.py`.

**Requirements:** `TASK-227`; approved spec §§7–9,12,14.

<read_first>Phase 132 requirement/risk graph, all 132 evidence, current open HIGH findings.</read_first>

**Consumes/Produces:** Accepted Tasks 132-A/B evidence and security requirement graph → fail-closed rule: every critical/high security requirement has automated observable evidence or approved schema-valid N/A with owner/reason.

**Action/sentinel:** Remove one critical link and prove collection fails; add only requirement-backed closure cases, no count quota.

<verify><automated>`PlanId=132-03`, `TaskId=TASK-227`, owner `qa-auto1`, verifier `security-auditor,qa-auto-ver`, profile `api-security-full`, inner/outer `1200/1440`.</automated></verify>

**Done:** every critical/high Phase 132 requirement resolves to automated observable evidence or schema-valid approved N/A; deleting one link fails the gate and no count-based substitute is accepted.

**Atomic commit:** `test(testframework): закрыть security risk graph`; stage exactly the four expanded Files paths above; body TASK-227. Shared docs are owned only by Task 132-D.

### Task 132-D: Synchronize Phase 132 security docs and traceability

**Files:** Modify `docs/security_guidelines.md`, `docs/test_plan.md`, `.planning/COVERAGE.md`, `testframework/README.md`.

**Requirements:** `TASK-208`, `TASK-209`, `TASK-227`; approved spec §§7–10,12–14.

<read_first>Accepted receipts from Tasks 132-A–C; exact four docs paths; current security requirement/TC graph.</read_first>

**Consumes/Produces:** Accepted authorization/abuse/closure receipts → exact security/test/COVERAGE/README links and one Phase 132 docs-closure receipt.

**Action/sentinel:** Record positive denial, abuse/replay and critical-risk closure evidence without exposing canaries. `docs-contract` fails on an omitted critical/high link, stale status, missing profile receipt or any staged docs path outside the four Files.

<verify><automated>`PlanId=132-04`, `TaskId=TASK-208,TASK-209,TASK-227`, owner `docs-writer`, verifiers `security-auditor,qa-auto-ver`, profile `docs-contract`, inner/outer `300/420`; receipt `testframework/reports/<task-run>/TASK-208-209-227-docs-sync-receipt.json`.</automated></verify>

**Done:** all accepted Phase 132 risks resolve to exact TASK/TC/profile evidence in the four docs and the phase cannot complete without the docs receipt.

**Atomic commit:** `docs(testframework): синхронизировать Phase 132 security evidence`; stage exactly the four Files paths above; body TASK-208/TASK-209/TASK-227.

## 13. Phase 133 — production-equivalent boundaries one at a time

### Task 133-A: PostgreSQL and Mongo ownership/transaction boundaries

**Files:** Create `testframework/src/hub_testkit/integration/{__init__.py,postgres.py,mongo.py}`, `testframework/tests/integration/{postgres/test_transactions.py,mongo/test_transactions.py}`, `testframework/cases/infrastructure/{TC-INFRA-POSTGRES-0001.md,TC-INFRA-MONGO-0001.md}`, `testframework/docker/compose.integration.yaml`.

**Requirements:** `TASK-210`; approved spec §§4,6–9,12,14,15.

<read_first>Production versions/topologies from current deployment docs; accepted dependency/image locks; required cleanup kill matrix; Phase 131 runtime ownership.</read_first>

**Consumes/Produces:** Runtime IDs, ownership ledger and production topology → exact per-run/worker database/schema ownership; no global DB/drop; cleanup verifies nonce and owned sessions. Mongo case distinguishes transaction-capable topology from one-node assumptions.

**Action/sentinel:** Start with commit/rollback/lock-timeout/partial-failure/kill sentinels; prove foreign database survives and owned database is gone after fixture/process/launcher kill.

<verify><automated>`PlanId=133-01`, `TaskId=TASK-210`, owner `qa-auto1`, verifier `qa-auto-ver,be-ver`, profile `integration-pg-mongo`, inner/outer `1200/1440`.</automated></verify>

**Done:** PostgreSQL/Mongo transaction, timeout, partial-failure and kill cases prove exact per-run cleanup, retain a foreign database sentinel and record production-equivalent topology evidence.

**Atomic commit:** `test(testframework): добавить PostgreSQL Mongo boundaries`; stage exactly the eight expanded Files paths above; body TASK-210. Shared docs are owned only by Task 133-C.

### Task 133-B: Redis, NATS and job delivery/recovery boundaries

**Files:** Create `testframework/src/hub_testkit/integration/{redis.py,nats.py}`, `testframework/tests/integration/{redis/test_namespace_cleanup.py,nats/test_delivery_recovery.py}`, `testframework/cases/infrastructure/{TC-INFRA-REDIS-0001.md,TC-INFRA-NATS-0001.md}`; Modify `testframework/docker/compose.integration.yaml`.

**Requirements:** `TASK-211`; approved spec §§4,6–9,12,14,15.

<read_first>Production Redis/NATS topology and job contracts; Task 133-A ownership pattern; current retry/idempotency requirements.</read_first>

**Consumes/Produces:** Runtime ownership and Redis/NATS delivery contracts → cryptographic key prefix with exact `SCAN`+`UNLINK`; unique subjects/streams/durables/queues; drain-before-delete and first-cause recovery timeline.

**Action/sentinel:** Fail first on duplicate/reorder/redelivery, consumer restart and foreign-key/stream deletion attempts; run process/launcher kill cleanup matrix.

<verify><automated>`PlanId=133-02`, `TaskId=TASK-211`, owner `qa-auto2`, verifier `qa-auto-ver,be-ver`, profile `integration-redis-nats`, inner/outer `1200/1440`.</automated></verify>

**Done:** Redis/NATS duplicate/order/redelivery/restart cases preserve first cause, remove only nonce-owned keys/streams/consumers after normal and killed runs, and retain all foreign sentinels.

**Atomic commit:** `test(testframework): добавить Redis NATS recovery boundaries`; stage exactly the seven expanded Files paths above; body TASK-211. Shared docs are owned only by Task 133-C.

### Task 133-C: Synchronize Phase 133 environment docs and traceability

**Files:** Modify `docs/test_plan.md`, `docs/stack_and_tools.md`, `.planning/COVERAGE.md`, `testframework/README.md`.

**Requirements:** `TASK-210`, `TASK-211`; approved spec §§6–9,12,14,15.

<read_first>Accepted receipts from Tasks 133-A/B; exact four docs paths; production topology and cleanup evidence.</read_first>

**Consumes/Produces:** Accepted PostgreSQL/Mongo/Redis/NATS receipts → exact environment/topology/TC ownership documentation and one Phase 133 docs-closure receipt.

**Action/sentinel:** Document per-run ownership, transaction/topology limits, reconnect/recovery and foreign-resource preservation. `docs-contract` fails on a global cleanup instruction, topology overclaim, missing receipt/profile or any staged docs path outside the four Files.

<verify><automated>`PlanId=133-03`, `TaskId=TASK-210,TASK-211`, owner `docs-writer`, verifiers `qa-auto-ver,be-ver`, profile `docs-contract`, inner/outer `300/420`; receipt `testframework/reports/<task-run>/TASK-210-211-docs-sync-receipt.json`.</automated></verify>

**Done:** the four docs map every boundary TC to accepted topology/cleanup evidence and Phase 133 completion requires the docs receipt.

**Atomic commit:** `docs(testframework): синхронизировать Phase 133 boundaries`; stage exactly the four Files paths above; body TASK-210/TASK-211.

## 14. Phases 134–136 — domain and contract expansion

### Phase 134 task set

### Task 134-A: Tracker lifecycle boundary

**Files:** `testframework/src/hub_testkit/domains/tracker/models.py`, `testframework/src/hub_testkit/domains/tracker/api.py`, `testframework/src/hub_testkit/domains/tracker/builders.py`, `testframework/src/hub_testkit/domains/tracker/oracles.py`, `testframework/tests/api/tracker/test_issue_lifecycle.py`, `testframework/cases/tracker/TC-TRACKER-0001.md`.

**Requirements:** `TASK-212`; approved spec §§4,7–9,12,14. <read_first>Verified tracker state/authz requirements, current API, Phase 133 boundary evidence.</read_first>

**Consumes/Produces:** consumes public HTTP/runtime interfaces; produces independent tracker models/facade/builder/oracles and one TC. **Sentinel:** removing tenant/version guard makes create-transition-conflict/idempotency variants red.

<verify><automated>`PlanId=134-01`, `TaskId=TASK-212`, profile `api-tracker`, owner `qa-auto1`, verifiers `qa-auto-ver,product-ver`, inner/outer `600/720`; receipt `testframework/reports/<task-run>/TASK-212-tracker-receipt.json` records case/nodeid/seed/first-exit/cleanup.</automated></verify>

**Done:** tracker TC proves happy, invalid, authz, boundary, conflict and idempotency behavior; no cross-domain import. **Atomic commit:** `test(tracker): покрыть lifecycle и denial cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 134-H.

### Task 134-B: Project scope boundary

**Files:** `testframework/src/hub_testkit/domains/project/models.py`, `testframework/src/hub_testkit/domains/project/api.py`, `testframework/src/hub_testkit/domains/project/builders.py`, `testframework/src/hub_testkit/domains/project/oracles.py`, `testframework/tests/api/project/test_project_scope.py`, `testframework/cases/project/TC-PROJECT-0001.md`.

**Requirements:** `TASK-212`; approved spec §§4,7–9,12,14. <read_first>Verified project scope/authz requirements, current API, accepted Task 134-A interfaces.</read_first>

**Consumes/Produces:** consumes public HTTP/runtime only; produces project facade and one TC. **Sentinel:** widening tenant/project membership scope makes cross-project denial red.

<verify><automated>`PlanId=134-02`, `TaskId=TASK-212`, profile `api-project`, owner `qa-auto1`, verifiers `qa-auto-ver,product-ver`, inner/outer `600/720`; receipt `testframework/reports/<task-run>/TASK-212-project-receipt.json`.</automated></verify>

**Done:** project TC proves scoped lifecycle/denial/boundary/recovery, sensitivity and exact cleanup. **Atomic commit:** `test(project): покрыть scope и denial cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 134-H.

### Task 134-C: Sprint conflict boundary

**Files:** `testframework/src/hub_testkit/domains/sprint/models.py`, `testframework/src/hub_testkit/domains/sprint/api.py`, `testframework/src/hub_testkit/domains/sprint/builders.py`, `testframework/src/hub_testkit/domains/sprint/oracles.py`, `testframework/tests/api/sprint/test_state_conflict.py`, `testframework/cases/sprint/TC-SPRINT-0001.md`.

**Requirements:** `TASK-212`; approved spec §§4,7–9,12,14. <read_first>Verified sprint transition/conflict requirements, current API, accepted prior 134 receipts.</read_first>

**Consumes/Produces:** consumes public HTTP/runtime; produces sprint facade and conflict oracle. **Sentinel:** removing optimistic version comparison makes concurrent transition case red.

<verify><automated>`PlanId=134-03`, `TaskId=TASK-212`, profile `api-sprint`, owner `qa-auto1`, verifiers `qa-auto-ver,product-ver`, inner/outer `600/720`; receipt `testframework/reports/<task-run>/TASK-212-sprint-receipt.json`.</automated></verify>

**Done:** sprint TC proves valid/invalid transition, concurrent conflict and retry/idempotency with deterministic ordering. **Atomic commit:** `test(sprint): покрыть conflict и recovery cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 134-H.

### Task 134-D: TMS run lifecycle boundary

**Files:** `testframework/src/hub_testkit/domains/tms/models.py`, `testframework/src/hub_testkit/domains/tms/api.py`, `testframework/src/hub_testkit/domains/tms/builders.py`, `testframework/src/hub_testkit/domains/tms/oracles.py`, `testframework/tests/api/tms/test_run_lifecycle.py`, `testframework/cases/tms/TC-TMS-0001.md`.

**Requirements:** `TASK-212`; approved spec §§4,7–9,12,14. <read_first>Verified TMS run-state/authz requirements, current API, accepted prior 134 receipts.</read_first>

**Consumes/Produces:** consumes public HTTP/runtime; produces TMS facade and run-state oracle. **Sentinel:** allowing an invalid terminal-state transition makes the TC red.

<verify><automated>`PlanId=134-04`, `TaskId=TASK-212`, profile `api-tms`, owner `qa-auto1`, verifiers `qa-auto-ver,product-ver`, inner/outer `600/720`; receipt `testframework/reports/<task-run>/TASK-212-tms-receipt.json`.</automated></verify>

**Done:** TMS TC proves lifecycle, denial, terminal boundary and recovery with exact cleanup. **Atomic commit:** `test(tms): покрыть run lifecycle cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 134-H.

### Task 134-E: Messenger delivery boundary

**Files:** Create `testframework/src/hub_testkit/domains/messenger/models.py`, `testframework/src/hub_testkit/domains/messenger/api.py`, `testframework/src/hub_testkit/domains/messenger/builders.py`, `testframework/src/hub_testkit/domains/messenger/oracles.py`, `testframework/tests/api/messenger/test_delivery.py`, `testframework/cases/messenger/TC-MESSENGER-0001.md`, `testframework/cases/messenger/TC-MESSENGER-0002.md`.

**Requirements:** `TASK-213`, `TASK-228`; approved spec §§4,7–9,12,14.

<read_first>Verified messenger membership/delivery requirements; accepted HTTP/runtime boundaries; WS delivery contract.</read_first>

**Consumes/Produces:** Public HTTP/runtime and messenger requirements → independent messenger models/facade/builder, delivery timeline/correlation oracle and two TCs.

**Action/sentinel:** Fail first on duplicate, reorder, missing delivery, wrong membership, timeout/partial response and foreign cleanup; use bounded visible polling with no NATS implementation dependency.

<verify><automated>`PlanId=134-05`, `TaskId=TASK-213,TASK-228`, owner `qa-auto2`, verifiers `qa-auto-ver,product-ver`, profile `api-messenger`, inner/outer `600/720`; receipt `testframework/reports/<task-run>/TASK-213-228-messenger-receipt.json`.</automated></verify>

**Done:** messenger delivery/membership/ordering/failure cases are deterministic, independently cleaned and do not import another domain.

**Atomic commit:** `test(messenger): покрыть delivery boundary`; stage exactly the seven Files paths above; body TASK-213/TASK-228. Shared docs are owned only by Task 134-H.

### Task 134-F: Notification and NATS reconnect boundary

**Files:** Create `testframework/src/hub_testkit/domains/notification/models.py`, `testframework/src/hub_testkit/domains/notification/api.py`, `testframework/src/hub_testkit/domains/notification/builders.py`, `testframework/src/hub_testkit/domains/notification/oracles.py`, `testframework/tests/api/notification/test_delivery.py`, `testframework/tests/integration/nats/test_reconnect_ordering.py`, `testframework/cases/notification/TC-NOTIFICATION-0001.md`.

**Requirements:** `TASK-213`, `TASK-228`; approved spec §§4,6–9,12,14.

<read_first>Verified notification requirements; accepted Phase 133 NATS receipt; reconnect/redelivery/ordering contracts.</read_first>

**Consumes/Produces:** Phase 133 NATS boundary and notification requirements → notification facade/oracles, reconnect ordering exercise and one TC.

**Action/sentinel:** Fail first on duplicate/reorder/redelivery loss, disconnect/reconnect, membership denial, timeout and foreign stream cleanup; retain the first causal event.

<verify><automated>`PlanId=134-06`, `TaskId=TASK-213,TASK-228`, owner `qa-auto2`, verifiers `qa-auto-ver,product-ver`, profile `integration-notification-nats`, inner/outer `900/1080`; receipt `testframework/reports/<task-run>/TASK-213-228-notification-nats-receipt.json`.</automated></verify>

**Done:** notification and reconnect cases preserve order/first cause, exact ownership and foreign resources after normal and interrupted runs.

**Atomic commit:** `test(notification): покрыть NATS reconnect boundary`; stage exactly the seven Files paths above; body TASK-213/TASK-228. Shared docs are owned only by Task 134-H.

### Task 134-G: Close the realtime risk graph

**Files:** Create `testframework/tests/api/messenger/test_cross_domain_closure.py`, `testframework/selftests/traceability/test_realtime_coverage.py`.

**Requirements:** `TASK-213`, `TASK-228`; approved spec §§7–10,12,14.

<read_first>Accepted Tasks 134-A–F receipts and current TASK-213/TASK-228 critical/high risk graph.</read_first>

**Consumes/Produces:** Six domain/boundary receipts → fail-closed realtime traceability rule and cross-domain observable closure without domain-to-domain imports.

**Action/sentinel:** Remove one critical/high evidence link and require collection failure; then prove every tracker/project/sprint/TMS/messenger/notification requirement resolves to a TC, nodeid, profile receipt and cleanup result without count quota.

<verify><automated>`PlanId=134-07`, `TaskId=TASK-213,TASK-228`, owner `qa-auto-ver`, verifiers `product-ver,arch-ver`, profile `api-realtime-closure`, inner/outer `900/1080`; receipt `testframework/reports/<task-run>/TASK-213-228-realtime-closure-receipt.json`.</automated></verify>

**Done:** every critical/high TASK-213/TASK-228 risk link resolves to accepted evidence and no cross-domain import or hidden polling exists.

**Atomic commit:** `test(testframework): закрыть realtime risk graph`; stage exactly the two Files paths above; body TASK-213/TASK-228. Shared docs are owned only by Task 134-H.

### Task 134-H: Synchronize Phase 134 docs and traceability

**Files:** Modify `docs/test_plan.md`, `.planning/COVERAGE.md`, `testframework/README.md`, `testframework/docs/architecture.md`.

**Requirements:** `TASK-212`, `TASK-213`, `TASK-228`; approved spec §§7–10,12,14.

<read_first>Accepted receipts from Tasks 134-A–G; exact four docs paths; realtime risk-closure receipt.</read_first>

**Consumes/Produces:** Accepted domain/realtime receipts → exact TASK/TC/profile/cleanup documentation and one Phase 134 docs-closure receipt.

**Action/sentinel:** Update only the four Files paths with domain ownership, public facade boundaries and evidence links. `docs-contract` fails on omitted Task 134 receipt, stale profile ID, cross-domain ownership claim or extra staged docs path.

<verify><automated>`PlanId=134-08`, `TaskId=TASK-212,TASK-213,TASK-228`, owner `docs-writer`, verifiers `qa-auto-ver,product-ver`, profile `docs-contract`, inner/outer `300/420`; receipt `testframework/reports/<task-run>/TASK-212-213-228-docs-sync-receipt.json`.</automated></verify>

**Done:** exact domain/realtime TASK/TC/profile/cleanup evidence is discoverable in the four docs and Phase 134 cannot complete without the docs receipt.

**Atomic commit:** `docs(testframework): синхронизировать Phase 134 domains`; stage exactly the four Files paths above; body TASK-212/TASK-213/TASK-228.

### Phase 135 task set

### Task 135-A: Calendar requirement boundary

**Files:** `testframework/src/hub_testkit/domains/calendar/models.py`, `testframework/src/hub_testkit/domains/calendar/api.py`, `testframework/src/hub_testkit/domains/calendar/builders.py`, `testframework/src/hub_testkit/domains/calendar/oracles.py`, `testframework/tests/api/calendar/test_calendar_requirements.py`, `testframework/cases/calendar/TC-CALENDAR-0001.md`.

**Requirements:** `TASK-214`; spec §§4,7–9,12–14. <read_first>Accepted calendar requirements/privacy review and Phase 131 HTTP boundary.</read_first> **Consumes/Produces:** public HTTP/runtime → independent calendar facade and TC. **Sentinel:** remove tenant/overlap guard; denial/conflict variant turns red.

<verify><automated>Profile `api-calendar`, PlanId `135-01`, TASK-214, owner qa-auto1, verifiers qa-auto-ver/product-ver, `600/720`; receipt `testframework/reports/<task-run>/TASK-214-calendar-receipt.json`.</automated></verify>

**Done:** calendar state/denial/boundary/privacy/failure variants are accepted. **Atomic commit:** `test(calendar): добавить requirement risk cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-B: Intranet requirement boundary

**Files:** `testframework/src/hub_testkit/domains/intranet/models.py`, `testframework/src/hub_testkit/domains/intranet/api.py`, `testframework/src/hub_testkit/domains/intranet/builders.py`, `testframework/src/hub_testkit/domains/intranet/oracles.py`, `testframework/tests/api/intranet/test_intranet_requirements.py`, `testframework/cases/intranet/TC-INTRANET-0001.md`.

**Requirements:** `TASK-214`; spec §§4,7–9,12–14. <read_first>Accepted intranet visibility requirements and 135-A receipt.</read_first> **Consumes/Produces:** public HTTP/runtime → intranet facade and TC. **Sentinel:** widen audience visibility; foreign-role case turns red.

<verify><automated>Profile `api-intranet`, PlanId `135-02`, TASK-214, owner qa-auto1, verifiers qa-auto-ver/product-ver, `600/720`; receipt `testframework/reports/<task-run>/TASK-214-intranet-receipt.json`.</automated></verify>

**Done:** intranet visibility, denial, privacy and recovery variants are sensitive and cleaned. **Atomic commit:** `test(intranet): добавить requirement risk cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-C: Accounting requirement boundary

**Files:** `testframework/src/hub_testkit/domains/accounting/models.py`, `testframework/src/hub_testkit/domains/accounting/api.py`, `testframework/src/hub_testkit/domains/accounting/builders.py`, `testframework/src/hub_testkit/domains/accounting/oracles.py`, `testframework/tests/api/accounting/test_accounting_requirements.py`, `testframework/cases/accounting/TC-ACCOUNTING-0001.md`.

**Requirements:** `TASK-214`; spec §§4,7–9,12–14. <read_first>Accepted accounting amount/tenant requirements and prior receipts.</read_first> **Consumes/Produces:** public HTTP/runtime → accounting facade and TC. **Sentinel:** relax amount/tenant invariant; boundary case turns red.

<verify><automated>Profile `api-accounting`, PlanId `135-03`, TASK-214, owner qa-auto1, verifiers qa-auto-ver/product-ver, `600/720`; receipt `testframework/reports/<task-run>/TASK-214-accounting-receipt.json`.</automated></verify>

**Done:** accounting state, invalid amount, tenant denial and idempotency variants are accepted. **Atomic commit:** `test(accounting): добавить requirement risk cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-D: CRM requirement boundary

**Files:** `testframework/src/hub_testkit/domains/crm/models.py`, `testframework/src/hub_testkit/domains/crm/api.py`, `testframework/src/hub_testkit/domains/crm/builders.py`, `testframework/src/hub_testkit/domains/crm/oracles.py`, `testframework/tests/api/crm/test_crm_requirements.py`, `testframework/cases/crm/TC-CRM-0001.md`.

**Requirements:** `TASK-214`; spec §§4,7–9,12–14. <read_first>Accepted CRM stage/ownership requirements and prior receipts.</read_first> **Consumes/Produces:** public HTTP/runtime → CRM facade and TC. **Sentinel:** permit invalid stage/owner transition; case turns red.

<verify><automated>Profile `api-crm`, PlanId `135-04`, TASK-214, owner qa-auto1, verifiers qa-auto-ver/product-ver, `600/720`; receipt `testframework/reports/<task-run>/TASK-214-crm-receipt.json`.</automated></verify>

**Done:** CRM lifecycle, ownership denial and recovery variants are accepted. **Atomic commit:** `test(crm): добавить requirement risk cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-E: HR privacy boundary

**Files:** `testframework/src/hub_testkit/domains/hr/models.py`, `testframework/src/hub_testkit/domains/hr/api.py`, `testframework/src/hub_testkit/domains/hr/builders.py`, `testframework/src/hub_testkit/domains/hr/oracles.py`, `testframework/tests/api/hr/test_hr_requirements.py`, `testframework/cases/hr/TC-HR-0001.md`.

**Requirements:** `TASK-214`; spec §§4,7–9,12–14. <read_first>Accepted HR role/privacy requirements and security policy.</read_first> **Consumes/Produces:** public HTTP/runtime → HR facade and TC. **Sentinel:** expose restricted field to wrong role; privacy case turns red without echoing value.

<verify><automated>Profile `api-hr`, PlanId `135-05`, TASK-214, owner qa-auto1, verifiers qa-auto-ver/security-auditor, `600/720`; receipt `testframework/reports/<task-run>/TASK-214-hr-receipt.json`.</automated></verify>

**Done:** HR state/role/privacy/failure variants and redaction evidence are accepted. **Atomic commit:** `test(hr): добавить privacy risk cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-F: Bases row/schema boundary

**Files:** `testframework/src/hub_testkit/domains/bases/models.py`, `testframework/src/hub_testkit/domains/bases/api.py`, `testframework/src/hub_testkit/domains/bases/builders.py`, `testframework/src/hub_testkit/domains/bases/oracles.py`, `testframework/tests/api/bases/test_bases_requirements.py`, `testframework/cases/bases/TC-BASES-0001.md`.

**Requirements:** `TASK-214`; spec §§4,7–9,12–14. <read_first>Accepted no-code base row/schema/tenant requirements.</read_first> **Consumes/Produces:** public HTTP/runtime → bases facade and TC. **Sentinel:** accept row violating schema/tenant scope; case turns red.

<verify><automated>Profile `api-bases`, PlanId `135-06`, TASK-214, owner qa-auto1, verifiers qa-auto-ver/product-ver, `600/720`; receipt `testframework/reports/<task-run>/TASK-214-bases-receipt.json`.</automated></verify>

**Done:** bases schema, row ownership, denial and recovery variants are accepted. **Atomic commit:** `test(bases): добавить requirement risk cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-G: Dashboard aggregation boundary

**Files:** `testframework/src/hub_testkit/domains/dashboard/models.py`, `testframework/src/hub_testkit/domains/dashboard/api.py`, `testframework/src/hub_testkit/domains/dashboard/builders.py`, `testframework/src/hub_testkit/domains/dashboard/oracles.py`, `testframework/tests/api/dashboard/test_dashboard_requirements.py`, `testframework/cases/dashboard/TC-DASHBOARD-0001.md`.

**Requirements:** `TASK-214`; spec §§4,7–9,12–14. <read_first>Accepted dashboard aggregation/tenant requirements and 135-A..F receipts.</read_first> **Consumes/Produces:** public HTTP/runtime → dashboard facade and TC. **Sentinel:** include foreign-tenant metric; aggregation case turns red.

<verify><automated>Profile `api-dashboard`, PlanId `135-07`, TASK-214, owner qa-auto1, verifiers qa-auto-ver/product-ver, `600/720`; receipt `testframework/reports/<task-run>/TASK-214-dashboard-receipt.json`.</automated></verify>

**Done:** dashboard aggregation, tenant denial, stale/failure and recovery variants are accepted. **Atomic commit:** `test(dashboard): добавить requirement risk cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-H: Storage adapter boundary

**Files:** `testframework/src/hub_testkit/integration/storage.py`, `testframework/tests/integration/storage/test_failure_recovery.py`, `testframework/cases/storage/TC-STORAGE-0001.md`.

**Requirements:** `TASK-215`, `TASK-229`; spec §§4,6–9,12–15. <read_first>SeaweedFS S3 subset, storage ownership/recovery contract.</read_first> **Consumes/Produces:** runtime ownership → typed storage errors and TC. **Sentinel:** partial upload/timeout leaves owned object or deletes foreign object; case turns red.

<verify><automated>Profile `integration-storage`, PlanId `135-08`, TASK-215/TASK-229, owner qa-auto2, verifiers qa-auto-ver/product-ver, `900/1080`; receipt `testframework/reports/<task-run>/TASK-215-storage-receipt.json`.</automated></verify>

**Done:** storage timeout/partial/retry/recovery/idempotency and exact cleanup are accepted; SeaweedFS remains authority. **Atomic commit:** `test(storage): добавить failure recovery cases`; stage exactly the three Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-I: GitHub adapter boundary

**Files:** `testframework/src/hub_testkit/integration/github.py`, `testframework/tests/integration/github/test_failure_recovery.py`, `testframework/cases/github/TC-GITHUB-0001.md`.

**Requirements:** `TASK-215`, `TASK-229`; spec §§4,6–9,12–15. <read_first>Accepted GitHub adapter/egress/retry contract.</read_first> **Consumes/Produces:** typed adapter policy → GitHub errors and TC. **Sentinel:** duplicate effect after retry or secret diagnostic turns red; no live service.

<verify><automated>Profile `integration-github`, PlanId `135-09`, TASK-215/TASK-229, owner qa-auto2, verifiers qa-auto-ver/security-auditor, `900/1080`; receipt `testframework/reports/<task-run>/TASK-215-github-receipt.json`.</automated></verify>

**Done:** deterministic unavailable/retry/idempotency/redaction evidence is accepted. **Atomic commit:** `test(github): добавить failure recovery cases`; stage exactly the three Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-J: GitLab adapter boundary

**Files:** `testframework/src/hub_testkit/integration/gitlab.py`, `testframework/tests/integration/gitlab/test_failure_recovery.py`, `testframework/cases/gitlab/TC-GITLAB-0001.md`.

**Requirements:** `TASK-215`, `TASK-229`; spec §§4,6–9,12–15. <read_first>Accepted GitLab adapter/egress/retry contract.</read_first> **Consumes/Produces:** typed adapter policy → GitLab errors and TC. **Sentinel:** partial response/retry duplicates mutation or leaks token; case turns red.

<verify><automated>Profile `integration-gitlab`, PlanId `135-10`, TASK-215/TASK-229, owner qa-auto2, verifiers qa-auto-ver/security-auditor, `900/1080`; receipt `testframework/reports/<task-run>/TASK-215-gitlab-receipt.json`.</automated></verify>

**Done:** GitLab failure/recovery/idempotency/redaction evidence is accepted. **Atomic commit:** `test(gitlab): добавить failure recovery cases`; stage exactly the three Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-K: 1C adapter boundary

**Files:** `testframework/src/hub_testkit/integration/onec.py`, `testframework/tests/integration/onec/test_failure_recovery.py`, `testframework/cases/onec/TC-1C-0001.md`.

**Requirements:** `TASK-215`, `TASK-229`; spec §§4,6–9,12–15. <read_first>Accepted 1C import/export/partial-completion contract.</read_first> **Consumes/Produces:** adapter policy → 1C errors and TC. **Sentinel:** partial import replay duplicates record or loses first cause; case turns red.

<verify><automated>Profile `integration-onec`, PlanId `135-11`, TASK-215/TASK-229, owner qa-auto2, verifiers qa-auto-ver/product-ver, `900/1080`; receipt `testframework/reports/<task-run>/TASK-215-onec-receipt.json`.</automated></verify>

**Done:** 1C partial/import/export/idempotency/recovery evidence is accepted. **Atomic commit:** `test(onec): добавить failure recovery cases`; stage exactly the three Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-L: Mail adapter boundary

**Files:** `testframework/src/hub_testkit/integration/mail.py`, `testframework/tests/integration/mail/test_failure_recovery.py`, `testframework/cases/mail/TC-MAIL-0001.md`.

**Requirements:** `TASK-215`, `TASK-229`; spec §§4,6–9,12–15. <read_first>Accepted mail delivery/privacy/retry contract.</read_first> **Consumes/Produces:** adapter policy → mail errors and TC. **Sentinel:** retry duplicates delivery or exposes recipient/body secret; case turns red.

<verify><automated>Profile `integration-mail`, PlanId `135-12`, TASK-215/TASK-229, owner qa-auto2, verifiers qa-auto-ver/security-auditor, `900/1080`; receipt `testframework/reports/<task-run>/TASK-215-mail-receipt.json`.</automated></verify>

**Done:** mail unavailable/retry/idempotency/privacy evidence is accepted without live egress. **Atomic commit:** `test(mail): добавить failure recovery cases`; stage exactly the three Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-M: AI requirement boundary

**Files:** `testframework/src/hub_testkit/domains/ai/models.py`, `testframework/src/hub_testkit/domains/ai/api.py`, `testframework/src/hub_testkit/domains/ai/builders.py`, `testframework/src/hub_testkit/domains/ai/oracles.py`, `testframework/tests/api/ai/test_ai_requirements.py`, `testframework/cases/ai/TC-AI-0001.md`.

**Requirements:** `TASK-215`, `TASK-229`; spec §§4,7–9,12–14. <read_first>Accepted AI Hub observable/privacy requirements; no provider live call.</read_first> **Consumes/Produces:** HTTP/runtime → AI facade and TC. **Sentinel:** provider failure leaks prompt/secret or loses safe error; case turns red.

<verify><automated>Profile `api-ai`, PlanId `135-13`, TASK-215/TASK-229, owner qa-auto2, verifiers qa-auto-ver/security-auditor, `600/720`; receipt `testframework/reports/<task-run>/TASK-215-ai-receipt.json`.</automated></verify>

**Done:** AI denial/provider-failure/redaction/recovery evidence is accepted. **Atomic commit:** `test(ai): добавить requirement risk cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-N: Search requirement boundary

**Files:** `testframework/src/hub_testkit/domains/search/models.py`, `testframework/src/hub_testkit/domains/search/api.py`, `testframework/src/hub_testkit/domains/search/builders.py`, `testframework/src/hub_testkit/domains/search/oracles.py`, `testframework/tests/api/search/test_search_requirements.py`, `testframework/cases/search/TC-SEARCH-0001.md`.

**Requirements:** `TASK-215`, `TASK-229`; spec §§4,7–9,12–14. <read_first>Accepted search scope/privacy/failure requirements.</read_first> **Consumes/Produces:** HTTP/runtime → search facade and TC. **Sentinel:** include foreign-tenant/private result; case turns red.

<verify><automated>Profile `api-search`, PlanId `135-14`, TASK-215/TASK-229, owner qa-auto2, verifiers qa-auto-ver/product-ver, `600/720`; receipt `testframework/reports/<task-run>/TASK-215-search-receipt.json`.</automated></verify>

**Done:** search scope, empty/boundary, privacy and dependency-failure evidence is accepted. **Atomic commit:** `test(search): добавить requirement risk cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-O: Backup requirement boundary

**Files:** `testframework/src/hub_testkit/domains/backup/models.py`, `testframework/src/hub_testkit/domains/backup/api.py`, `testframework/src/hub_testkit/domains/backup/builders.py`, `testframework/src/hub_testkit/domains/backup/oracles.py`, `testframework/tests/api/backup/test_backup_requirements.py`, `testframework/cases/backup/TC-BACKUP-0001.md`.

**Requirements:** `TASK-215`, `TASK-229`; spec §§4,7–9,12–14. <read_first>Accepted backup completion/integrity/recovery requirements.</read_first> **Consumes/Produces:** HTTP/runtime → backup facade and TC. **Sentinel:** partial backup marked complete or corrupt restore accepted; case turns red.

<verify><automated>Profile `api-backup`, PlanId `135-15`, TASK-215/TASK-229, owner qa-auto2, verifiers qa-auto-ver,product-ver, `600/720`; receipt `testframework/reports/<task-run>/TASK-215-backup-receipt.json`.</automated></verify>

**Done:** backup partial/completion/integrity/retry/recovery evidence is accepted. **Atomic commit:** `test(backup): добавить requirement risk cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-P: Import/export requirement boundary

**Files:** `testframework/src/hub_testkit/domains/imex/models.py`, `testframework/src/hub_testkit/domains/imex/api.py`, `testframework/src/hub_testkit/domains/imex/builders.py`, `testframework/src/hub_testkit/domains/imex/oracles.py`, `testframework/tests/api/imex/test_imex_requirements.py`, `testframework/cases/imex/TC-IMEX-0001.md`.

**Requirements:** `TASK-215`, `TASK-229`; spec §§4,7–9,12–14. <read_first>Accepted import/export validation/partial/replay requirements.</read_first> **Consumes/Produces:** HTTP/runtime → imex facade and TC. **Sentinel:** malformed/partial import advances state or replay duplicates rows; case turns red.

<verify><automated>Profile `api-imex`, PlanId `135-16`, TASK-215/TASK-229, owner qa-auto2, verifiers qa-auto-ver/product-ver, `600/720`; receipt `testframework/reports/<task-run>/TASK-215-imex-receipt.json`.</automated></verify>

**Done:** imex validation/partial/replay/idempotency/recovery evidence is accepted. **Atomic commit:** `test(imex): добавить requirement risk cases`; stage exactly the six Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-Q: Marketplace requirement and closure boundary

**Files:** `testframework/src/hub_testkit/domains/marketplace/models.py`, `testframework/src/hub_testkit/domains/marketplace/api.py`, `testframework/src/hub_testkit/domains/marketplace/builders.py`, `testframework/src/hub_testkit/domains/marketplace/oracles.py`, `testframework/tests/api/marketplace/test_marketplace_requirements.py`, `testframework/cases/marketplace/TC-MARKETPLACE-0001.md`, `testframework/selftests/traceability/test_business_integration_coverage.py`.

**Requirements:** `TASK-215`, `TASK-229`; spec §§4,7–9,12–14. <read_first>Accepted marketplace install/tenant/privacy requirements and receipts 135-A through 135-P.</read_first> **Consumes/Produces:** HTTP/runtime plus prior receipt graph → marketplace facade/TC and fail-closed TASK-229 coverage gate. **Sentinel:** remove one prior critical/high TASK-214/215/229 evidence link or permit unauthorized install; collection turns red.

<verify><automated>Profile `api-marketplace-closure`, PlanId `135-17`, TASK-215/TASK-229, owner qa-auto2, verifiers qa-auto-ver/product-ver/security-auditor, `900/1080`; receipt `testframework/reports/<task-run>/TASK-229-business-integration-closure-receipt.json` binds all 17 receipts.</automated></verify>

**Done:** marketplace behavior and complete business/integration risk graph are accepted with no live egress or generic superclass. **Atomic commit:** `test(marketplace): закрыть business integration risk graph`; stage exactly the seven Files paths above. Shared docs are owned only by Task 135-R.

### Task 135-R: Synchronize Phase 135 business/integration docs and traceability

**Files:** Modify `docs/test_plan.md`, `docs/stack_and_tools.md`, `docs/security_guidelines.md`, `.planning/COVERAGE.md`, `testframework/README.md`.

**Requirements:** `TASK-214`, `TASK-215`, `TASK-229`; approved spec §§6–10,12–15.

<read_first>Accepted receipts from Tasks 135-A–Q; exact five docs paths; TASK-229 business/integration closure receipt.</read_first>

**Consumes/Produces:** Seventeen task-owned domain/adapter receipts → exact business/integration TASK/TC/profile/egress/cleanup documentation and one Phase 135 docs-closure receipt.

**Action/sentinel:** Update only the five Files paths with all 17 receipt links, risk/egress boundaries and SeaweedFS authority. `docs-contract` fails on one missing domain/adapter, live-egress instruction, generic superclass, stale profile ID or extra staged docs path.

<verify><automated>`PlanId=135-18`, `TaskId=TASK-214,TASK-215,TASK-229`, owner `docs-writer`, verifiers `qa-auto-ver,product-ver,security-auditor`, profile `docs-contract`, inner/outer `300/420`; receipt `testframework/reports/<task-run>/TASK-214-215-229-docs-sync-receipt.json` binds all 17 task receipts.</automated></verify>

**Done:** every Phase 135 domain/adapter risk resolves to exact TC/profile/egress/cleanup evidence in the five docs; no domain task shares docs ownership and phase completion requires the docs receipt.

**Atomic commit:** `docs(testframework): синхронизировать Phase 135 evidence`; stage exactly the five Files paths above; body TASK-214/TASK-215/TASK-229.

### Phase 136 task set

### Task 136-A: OpenAPI, JSON Schema and event contracts

**Files:** Create `testframework/src/hub_testkit/contract/{__init__.py,openapi.py,jsonschema.py,events.py,snapshots.py}`, `testframework/selftests/contract/{test_openapi.py,test_jsonschema.py,test_events.py}`, `testframework/tests/contract/{api/test_openapi.py,events/test_envelopes.py,schema/test_migrations.py}`, `testframework/cases/contract/{TC-CONTRACT-0001.md,TC-CONTRACT-0002.md,TC-CONTRACT-0003.md}`.

**Requirements:** `TASK-216`; approved spec §§4,7–9,11,12,14.

<read_first>Hub OpenAPI/schema/event authorities; approved spec §§4,11; TASK-216 requirements.</read_first>

**Consumes/Produces:** Hub OpenAPI/schema/event authorities and accepted contract packages → explicit OpenAPI 3.1.2 + JSON Schema 2020-12 validation, strict formats, narrow canonical snapshots secondary to semantic oracle.

**Action/sentinel:** Break route/schema/event compatibility one field at a time and require red; generated clients/Schemathesis remain Trial unless separate spike accepted.

<verify><automated>`PlanId=136-01`, `TaskId=TASK-216`, owner `qa-auto1`, verifier `arch-ver,qa-auto-ver`, profile `contract`, inner/outer `900/1080`.</automated></verify>

**Done:** OpenAPI 3.1.2, JSON Schema 2020-12 and event/migration contracts each fail on the named one-field compatibility mutation and use semantic oracles as authority over narrow snapshots.

**Atomic commit:** `test(testframework): добавить API schema event contracts`; stage exactly the fourteen expanded Files paths above; body TASK-216. Shared docs are owned only by Task 136-C.

### Task 136-B: Frontend artifact contracts without source-token tests

**Files:** Create `testframework/tests/contract/frontend/{test_routes.py,test_i18n.py,test_accessibility_manifest.py,test_docs_links.py}`, `testframework/cases/frontend/{TC-FRONTEND-CONTRACT-0001.md,TC-FRONTEND-CONTRACT-0002.md,TC-FRONTEND-CONTRACT-0003.md,TC-FRONTEND-CONTRACT-0004.md}`; Modify no `frontend/tests/**` file.

**Requirements:** `TASK-230`; approved spec §§1,7–9,12,14.

<read_first>Built frontend artifact/public route/i18n/a11y/docs contract, TASK-230; deleted static-contract manifest cohort.</read_first>

**Consumes/Produces:** Frontend-produced public artifacts and contract authority → semantic artifact validation; it never greps token/filename implementation details as behavioral evidence.

**Action/sentinel:** Corrupt each artifact fixture and require failure; prove no legacy static-contract content was ported.

<verify><automated>`PlanId=136-02`, `TaskId=TASK-230`, owner `qa-auto2`, verifier `arch-ver,qa-auto-ver`, profile `contract-frontend`, inner/outer `900/1080`.</automated></verify>

**Done:** four frontend artifact contracts fail on corrupted produced fixtures, never use source-token/file-name grep as behavioral oracle and add no static-contract port or frontend unit edit.

**Atomic commit:** `test(testframework): заменить static guards artifact contracts`; stage exactly the eight expanded Files paths above; body TASK-230. Shared docs are owned only by Task 136-C.

### Task 136-C: Synchronize Phase 136 contract docs and traceability

**Files:** Modify `docs/test_plan.md`, `.planning/COVERAGE.md`, `testframework/README.md`, `testframework/docs/architecture.md`.

**Requirements:** `TASK-216`, `TASK-230`; approved spec §§7–12,14.

<read_first>Accepted receipts from Tasks 136-A/B; exact four docs paths; OpenAPI/schema/event/frontend contract TC graph.</read_first>

**Consumes/Produces:** Accepted semantic contract receipts → exact API/schema/event/frontend ownership and one Phase 136 docs-closure receipt.

**Action/sentinel:** Document semantic oracles, snapshot limits and artifact authority without source-token claims. `docs-contract` fails on missing TC/profile evidence, source-grep instruction, stale schema version or extra staged docs path.

<verify><automated>`PlanId=136-03`, `TaskId=TASK-216,TASK-230`, owner `docs-writer`, verifiers `arch-ver,qa-auto-ver`, profile `docs-contract`, inner/outer `300/420`; receipt `testframework/reports/<task-run>/TASK-216-230-docs-sync-receipt.json`.</automated></verify>

**Done:** all Phase 136 contract evidence is linked in the four docs with semantic authority and no source-token overclaim; phase completion requires the docs receipt.

**Atomic commit:** `docs(testframework): синхронизировать Phase 136 contracts`; stage exactly the four Files paths above; body TASK-216/TASK-230.

## 15. Phases 137–140 — browser, journeys, cross-cutting quality and load

### Phase 137 task set

### Task 137-A: Accept the Playwright lifecycle and component-first boundary

**Files:** Create `testframework/src/hub_testkit/browser/{__init__.py,lifecycle.py,auth_state.py,artifacts.py}`, `testframework/src/hub_testkit/domains/identity/ui/{pages/login.py,components/session_form.py}`, `testframework/selftests/browser/{test_lifecycle.py,test_artifacts.py}`, `testframework/docker/{Dockerfile.e2e,compose.e2e.yaml}`; Modify `testframework/pyproject.toml`, `testframework/uv.lock`, `testframework/docker/images.lock.json` only after package/browser/image matrix acceptance.

**Requirements:** `TASK-194`; approved spec §§4,6,11,12,14,15.

<read_first>Approved spec §§4,11,15; Playwright official Docker guidance; TASK-194; Phase 129 launcher boundary.</read_first>

**Consumes/Produces:** Accepted Playwright package/runtime and launcher boundaries → non-root browser/context/page factories, role-scoped validated storage state, failure-only redacted trace/screenshot policy, thin `LoginPage` + root-scoped `SessionForm`.

**Action/sentinel:** Before dependency adoption, fail package/browser/image mismatch, root/no-sandbox, host IPC, foreign auth state, secret trace and insufficient private shm. Accept only matching package/browser/digest with measured workers/shm.

<verify><automated>`PlanId=137-01`, `TaskId=TASK-194`, owner `qa-auto1`, verifier `qa-auto-ver,security-auditor`, profile `browser-framework`, inner/outer `1200/1440`.</automated></verify>

**Done:** accepted Playwright package/browser/image tuple runs non-root with private measured shm, redacted failure artifacts and isolated role state; every mismatch/root/IPC/secret sentinel fails before standardization.

**Atomic commit:** `feat(testframework): добавить безопасный Playwright lifecycle`; stage exactly the thirteen expanded Files paths above; body TASK-194. Shared docs are owned only by Task 137-C.

### Task 137-B: Add one critical browser smoke tracer

**Files:** Create `testframework/tests/e2e/identity/test_login_smoke.py`, `testframework/cases/identity/TC-IDENTITY-E2E-0001.md`; Modify only `testframework/src/hub_testkit/domains/identity/ui/pages/login.py` and `testframework/src/hub_testkit/domains/identity/ui/components/session_form.py` if required by the visible flow.

**Requirements:** `TASK-217`; approved spec §§4,7–9,12,14.

<read_first>Verified critical login journey/UI contract, Task 137-A accepted browser evidence, existing identity API setup facade.</read_first>

**Consumes/Produces:** Phase 137-A browser lifecycle and identity API setup → visible API-assisted setup plus browser-proven login outcome with accessible semantic locators and forbidden session/secret side effects.

**Action/sentinel:** Write failing browser case; wrong role, invalid credentials and leaked artifact canaries fail. No translated TS selector/workflow.

<verify><automated>`PlanId=137-02`, `TaskId=TASK-217`, owner `qa-auto2`, verifier `qa-auto-ver,security-auditor`, profile `e2e-smoke`, inner/outer `900/1080`.</automated></verify>

**Done:** one stable login TC proves accessible user-visible success plus wrong-role/invalid-credential denial, leaks no session/secret artifacts and contains no translated TS workflow.

**Atomic commit:** `test(testframework): добавить risk-based login smoke`; stage exactly the four Files paths above; body TASK-217. Shared docs are owned only by Task 137-C.

### Task 137-C: Synchronize Phase 137 browser docs and traceability

**Files:** Modify `docs/stack_and_tools.md`, `docs/security_guidelines.md`, `docs/test_plan.md`, `.planning/COVERAGE.md`, `testframework/README.md`.

**Requirements:** `TASK-194`, `TASK-217`; approved spec §§5–9,11,12,14,15.

<read_first>Accepted browser lifecycle and login-smoke receipts; exact five docs paths; accepted package/browser/image tuple.</read_first>

**Consumes/Produces:** Accepted browser/image/smoke receipts → exact component-first authoring, security and TC documentation plus one Phase 137 docs-closure receipt.

**Action/sentinel:** Document non-root/private-shm/image matching, failure-only redacted artifacts and visible login behavior. `docs-contract` fails on TypeScript parity/port language, secret-bearing artifact guidance, direct inner runner, stale profile or extra staged docs path.

<verify><automated>`PlanId=137-03`, `TaskId=TASK-194,TASK-217`, owner `docs-writer`, verifiers `qa-auto-ver,security-auditor`, profile `docs-contract`, inner/outer `300/420`; receipt `testframework/reports/<task-run>/TASK-194-217-docs-sync-receipt.json`.</automated></verify>

**Done:** the five docs expose the accepted browser lifecycle, login TC and immutable image/profile contract without parity claims; phase completion requires the docs receipt.

**Atomic commit:** `docs(testframework): синхронизировать Phase 137 browser evidence`; stage exactly the five Files paths above; body TASK-194/TASK-217.

### Phase 138 task set

### Task 138-A: Cross-instance session revocation journey

**Files:** `testframework/tests/e2e/identity/test_session_revocation.py`, `testframework/cases/identity/TC-IDENTITY-E2E-0002.md`, `testframework/src/hub_testkit/domains/identity/ui/pages/login.py`, `testframework/src/hub_testkit/domains/identity/ui/components/session_form.py`.

**Requirements:** `TASK-195`; approved spec §§4,7–9,12–14. <read_first>Verified session-revocation journey, Phase 137 login/browser contract, identity API setup.</read_first>

**Consumes/Produces:** identity API setup plus literal login UI files → one revoke-to-deny journey/TC. **Sentinel:** revoke on instance A but allow the same session on instance B; journey turns red and retains first cause.

<verify><automated>`PlanId=138-01`, TASK-195, owner qa-auto1, verifiers qa-auto-ver/qa-manual-ver, profile `e2e-identity-revocation`, `900/1080`; receipt `testframework/reports/<task-run>/TASK-195-identity-revocation-receipt.json`.</automated></verify>

**Done:** cross-instance revoke-to-deny is sensitive, user-visible, redacted and exactly cleaned. **Atomic commit:** `test(identity): добавить session revocation journey`; stage exactly the four Files paths above. Shared docs are owned only by Task 138-E.

### Task 138-B: Tracker realtime conflict journey

**Files:** `testframework/tests/e2e/tracker/test_realtime_conflict.py`, `testframework/cases/tracker/TC-TRACKER-E2E-0001.md`, `testframework/src/hub_testkit/domains/tracker/ui/pages/board.py`, `testframework/src/hub_testkit/domains/tracker/ui/components/issue_card.py`.

**Requirements:** `TASK-195`; approved spec §§4,7–9,12–14. <read_first>Verified optimistic-conflict/realtime journey, accepted Phase 134 tracker/realtime facades and browser contract.</read_first>

**Consumes/Produces:** tracker/realtime setup plus literal board/card UI → one conflict/reconciliation journey/TC. **Sentinel:** accept both concurrent versions or omit reconciliation; journey turns red.

<verify><automated>`PlanId=138-02`, TASK-195, owner qa-auto1, verifiers qa-auto-ver/qa-manual-ver, profile `e2e-tracker-conflict`, `1200/1440`; receipt `testframework/reports/<task-run>/TASK-195-tracker-conflict-receipt.json`.</automated></verify>

**Done:** optimistic conflict and realtime reconciliation are observable, deterministic and exactly cleaned. **Atomic commit:** `test(tracker): добавить realtime conflict journey`; stage exactly the four Files paths above. Shared docs are owned only by Task 138-E.

### Task 138-C: Messenger attachment recovery journey

**Files:** `testframework/tests/e2e/messenger/test_attachment_recovery.py`, `testframework/cases/messenger/TC-MESSENGER-E2E-0001.md`, `testframework/src/hub_testkit/domains/messenger/ui/pages/channel.py`, `testframework/src/hub_testkit/domains/messenger/ui/components/attachment_form.py`.

**Requirements:** `TASK-195`; approved spec §§4,7–9,12–14. <read_first>Verified attachment partial-failure/recovery journey, shipped SeaweedFS behavior, accepted messenger/browser contracts.</read_first>

**Consumes/Produces:** messenger/storage setup plus literal channel/form UI → one partial-failure/recovery journey/TC. **Sentinel:** persist message metadata after failed attachment without compensating recovery; journey turns red and foreign object survives cleanup.

<verify><automated>`PlanId=138-03`, TASK-195, owner qa-auto1, verifiers qa-auto-ver/qa-manual-ver, profile `e2e-messenger-attachment`, `1200/1440`; receipt `testframework/reports/<task-run>/TASK-195-messenger-attachment-receipt.json`.</automated></verify>

**Done:** attachment partial failure, compensation/retry and exact owned cleanup are observable without product storage migration. **Atomic commit:** `test(messenger): добавить attachment recovery journey`; stage exactly the four Files paths above. Shared docs are owned only by Task 138-E.

### Task 138-D: Recovery and multi-user variation matrix

**Files:** Create `testframework/tests/e2e/recovery/{test_process_restart.py,test_dependency_timeout.py,test_multi_user_isolation.py}`, `testframework/cases/recovery/{TC-RECOVERY-0001.md,TC-RECOVERY-0002.md,TC-RECOVERY-0003.md}`.

**Requirements:** `TASK-196`; approved spec §§6–9,12,14.

<read_first>Verified TASK-196 recovery requirements; accepted Tasks 138-A through 138-C; launcher first-exit/kill evidence.</read_first>

**Consumes/Produces:** Existing public facades and accepted Tasks 138-A–C receipts → retry/idempotency/tenant/user isolation and first-cause evidence; adds no recovery superclass/DSL.

**Action/sentinel:** Force app restart, dependency timeout/partial completion and two-actor cross-tenant attempt; foreign resource must survive exact cleanup.

<verify><automated>`PlanId=138-04`, `TaskId=TASK-196`, owner `qa-auto2`, verifier `qa-auto-ver,qa-manual-ver`, profile `e2e-recovery`, inner/outer `1800/2100`; receipt `testframework/reports/<task-run>/TASK-196-recovery-matrix-receipt.json`.</automated></verify>

**Done:** restart, dependency-timeout/partial-completion and multi-user cross-tenant cases preserve first cause, prove retry/idempotency/isolation and retain the foreign-resource sentinel after exact cleanup.

**Atomic commit:** `test(testframework): добавить E2E recovery matrix`; stage exactly the six expanded Files paths above; body TASK-196. Shared docs are owned only by Task 138-E.

### Task 138-E: Synchronize Phase 138 journey docs and traceability

**Files:** Modify `docs/test_plan.md`, `docs/runbooks/README.md`, `.planning/COVERAGE.md`, `testframework/README.md`.

**Requirements:** `TASK-195`, `TASK-196`; approved spec §§6–9,12–14.

<read_first>Accepted receipts from Tasks 138-A–D; exact four docs paths; journey and recovery TC graph.</read_first>

**Consumes/Produces:** Accepted literal journeys/recovery receipts → exact user-visible journey, recovery, cleanup and runbook links plus one Phase 138 docs-closure receipt.

**Action/sentinel:** Document the three literal journeys and recovery matrix without TypeScript port/parity language. `docs-contract` fails on omitted journey/TC/profile, missing first-cause/cleanup guidance or extra staged docs path.

<verify><automated>`PlanId=138-05`, `TaskId=TASK-195,TASK-196`, owner `docs-writer`, verifiers `qa-auto-ver,qa-manual-ver`, profile `docs-contract`, inner/outer `300/420`; receipt `testframework/reports/<task-run>/TASK-195-196-docs-sync-receipt.json`.</automated></verify>

**Done:** the four docs map all literal journeys and recovery cases to exact receipts and runbook behavior; phase completion requires the docs receipt.

**Atomic commit:** `docs(testframework): синхронизировать Phase 138 journeys`; stage exactly the four Files paths above; body TASK-195/TASK-196.

### Phase 139 task set

### Task 139-A: WCAG-oriented interaction and scoped axe evidence

**Files:** Create `testframework/tests/e2e/accessibility/{test_login.py,test_tracker.py,test_messenger.py}`, `testframework/cases/accessibility/{TC-A11Y-0001.md,TC-A11Y-0002.md,TC-A11Y-0003.md}`; Modify `testframework/pyproject.toml`, `testframework/uv.lock`, `testframework/docker/images.lock.json` only after bounded axe integration acceptance.

**Requirements:** `TASK-231`; approved spec §§7–9,11,12,14,15.

<read_first>Product WCAG/NFR requirements, accepted UI states, TASK-231, axe spike decision.</read_first>

**Consumes/Produces:** Accepted browser journeys and scoped accessibility requirements → explicit keyboard/focus/name-role-state/contrast checks plus documented automated-axe/manual boundaries, without overstating full WCAG sign-off.

**Action/sentinel:** Inject one known violation per state and require red; artifacts are centrally redacted.

<verify><automated>`PlanId=139-01`, `TaskId=TASK-231`, owner `qa-auto1`, verifier `qa-auto-ver,designer-ver`, profile `e2e-a11y`, inner/outer `1200/1440`.</automated></verify>

**Done:** login/tracker/messenger states have keyboard, focus, name/role/state and scoped axe evidence; each injected violation fails, artifacts are redacted and the result does not overclaim full WCAG sign-off.

**Atomic commit:** `test(testframework): добавить scoped accessibility evidence`; stage exactly the nine expanded Files paths above; body TASK-231. Shared docs are owned only by Task 139-C.

### Task 139-B: Security/privacy artifact and abuse closure

**Files:** Create `testframework/tests/api/security/test_cross_cutting.py`, `testframework/tests/contract/security/test_artifact_redaction.py`, `testframework/cases/security/{TC-SECURITY-0101.md,TC-SECURITY-0102.md}`; Modify `testframework/policies/redaction.toml` only for verified new field classes.

**Requirements:** `TASK-232`; approved spec §§6–9,12–14.

<read_first>Open security/privacy requirements and Phase 132/138/139 evidence; TASK-232.</read_first>

**Consumes/Produces:** Cross-layer abuse evidence and `Redactor` → sealed artifact scan and privacy receipt; no secret is copied to diagnostics.

**Action/sentinel:** Canary tokens/PII in logs/JUnit/trace/screenshot/run manifest must make the profile fail without echoing the value.

<verify><automated>`PlanId=139-02`, `TaskId=TASK-232`, owner `security-auditor`, verifier `qa-auto-ver`, profile `security-privacy`, inner/outer `1200/1440`.</automated></verify>

**Done:** cross-layer abuse and sealed-artifact scans detect canary secret/PII in every named artifact class without echoing it, and applicable critical/high privacy links are closed.

**Atomic commit:** `test(testframework): закрыть security privacy artifacts`; stage exactly the five expanded Files paths above; body TASK-232. Shared docs are owned only by Task 139-C.

### Task 139-C: Synchronize Phase 139 accessibility/security docs

**Files:** Modify `docs/security_guidelines.md`, `docs/test_plan.md`, `.planning/COVERAGE.md`, `testframework/README.md`.

**Requirements:** `TASK-231`, `TASK-232`; approved spec §§7–10,12–15.

<read_first>Accepted accessibility/security receipts from Tasks 139-A/B; exact four docs paths; NFR and privacy evidence boundaries.</read_first>

**Consumes/Produces:** Accepted a11y/privacy/artifact receipts → exact automated/manual scope, redaction and TC links plus one Phase 139 docs-closure receipt.

**Action/sentinel:** Document keyboard/focus/name-role-state/contrast evidence and the bounded axe/manual boundary; document canary-safe artifact scans. `docs-contract` fails on full-WCAG overclaim, leaked canary, missing TC/profile or extra staged docs path.

<verify><automated>`PlanId=139-03`, `TaskId=TASK-231,TASK-232`, owner `docs-writer`, verifiers `qa-auto-ver,designer-ver,security-auditor`, profile `docs-contract`, inner/outer `300/420`; receipt `testframework/reports/<task-run>/TASK-231-232-docs-sync-receipt.json`.</automated></verify>

**Done:** the four docs state exact automated/manual accessibility scope and redacted security evidence; phase completion requires the docs receipt.

**Atomic commit:** `docs(testframework): синхронизировать Phase 139 evidence`; stage exactly the four Files paths above; body TASK-231/TASK-232.

### Phase 140 task set

### Task 140-A: Prove Locust measurement validity and production-target denial

**Files:** Create `testframework/docker/Dockerfile.load`, `testframework/load/{locustfile.py,target_guard.py}`, `testframework/load/users/health.py`, `testframework/load/profiles/baseline.toml`, `testframework/selftests/load/{test_target_guard.py,test_measurement_validity.py}`; Modify `testframework/pyproject.toml`, `testframework/uv.lock`, `testframework/docker/images.lock.json` after package acceptance.

**Requirements:** `TASK-200`; approved spec §§4,6,7,8,11,12,14,15.

<read_first>Approved spec §§4,11,15; TASK-200; actual hot-path/SLO and environment attestation; Locust official lifecycle.</read_first>

**Consumes/Produces:** Accepted HTTP/runtime ownership and load target policy → native Locust client, signed target guard, and result-validity model invalidating DNS/redirect/proxy production reach, generator saturation and worker loss.

**Action/sentinel:** Production/shared target, redirect, DNS rebinding, worker loss and load-generator saturation must fail before/after measurement. HTTPX client/pytest fixtures are forbidden.

<verify><automated>`PlanId=140-01`, `TaskId=TASK-200`, owner `qa-auto2`, verifier `qa-auto-ver,devops-engineer`, profile `load-validity`, inner/outer `1200/1440`.</automated></verify>

**Done:** native Locust runtime rejects production/shared/redirect/rebinding targets, generator saturation and worker loss, records valid baseline measurement and imports neither HTTPX client nor pytest fixtures.

**Atomic commit:** `feat(testframework): добавить guarded Locust runtime`; stage exactly the ten expanded Files paths above; body TASK-200. Shared docs are owned only by Task 140-C.

### Task 140-B: Establish measured steady/spike/soak profiles

**Files:** Create `testframework/load/profiles/{steady.toml,spike.toml,soak.toml}`, `testframework/load/users/{tracker.py,messenger.py}`, `testframework/cases/load/{TC-LOAD-0001.md,TC-LOAD-0002.md}`.

**Requirements:** `TASK-200`; approved spec §§7–9,12,14,15.

<read_first>Accepted SLO/hot paths and Task 140-A validity evidence; current production-equivalent topology.</read_first>

**Consumes/Produces:** Accepted load-validity model and product SLOs → explicit users/rate/duration/SLO/abort thresholds and a report distinguishing SUT saturation from generator invalidity.

**Action/sentinel:** First use deliberately violated threshold and worker-loss sentinel; accept profile only with valid steady result, reproducible seed/data and exact cleanup.

<verify><automated>`PlanId=140-02`, `TaskId=TASK-200`, owner `qa-auto1`, verifier `qa-auto-ver,devops-engineer`, profile `load-steady`, inner/outer `3600/3900`. Spike/soak run only in their approved release/manual lanes.</automated></verify>

**Done:** steady/spike/soak profiles declare users/rate/duration/SLO/abort thresholds; steady evidence is valid and reproducible, threshold/worker-loss sentinels fail, and spike/soak remain restricted to approved lanes.

**Atomic commit:** `test(testframework): добавить measured load profiles`; stage exactly the seven expanded Files paths above; body TASK-200. Shared docs are owned only by Task 140-C.

### Task 140-C: Synchronize Phase 140 load docs and traceability

**Files:** Modify `docs/stack_and_tools.md`, `docs/security_guidelines.md`, `docs/test_plan.md`, `.planning/COVERAGE.md`, `testframework/README.md`.

**Requirements:** `TASK-200`; approved spec §§6–9,12,14,15.

<read_first>Accepted load-validity and measured-profile receipts; exact five docs paths; approved SLO/abort/production-deny contracts.</read_first>

**Consumes/Produces:** Accepted Locust validity/steady evidence → exact target guard, measurement validity, SLO/abort and TC documentation plus one Phase 140 docs-closure receipt.

**Action/sentinel:** Document generator-invalid versus SUT-saturation outcomes and release/manual spike/soak lanes. `docs-contract` fails on production target guidance, unmeasured threshold, missing receipt/profile or extra staged docs path.

<verify><automated>`PlanId=140-03`, `TaskId=TASK-200`, owner `docs-writer`, verifiers `qa-auto-ver,devops-engineer`, profile `docs-contract`, inner/outer `300/420`; receipt `testframework/reports/<task-run>/TASK-200-docs-sync-receipt.json`.</automated></verify>

**Done:** the five docs map each load TC/profile to valid measurement and production-deny evidence; phase completion requires the docs receipt.

**Atomic commit:** `docs(testframework): синхронизировать Phase 140 load evidence`; stage exactly the five Files paths above; body TASK-200.

## 16. Phase 141 — CI, supply chain, clean-room and handoff

### Task 141-A: Wire calibrated CI jobs to the sole launcher

**Files:** Create `.github/workflows/qa.yml`, `testframework/policies/retention.toml`, `testframework/selftests/ci/test_ci_contract.py`; Modify `.gitlab-ci.yml`, `testframework/docker/images.lock.json`.

**Requirements:** `TASK-233`; approved spec §§5,6,10–12,14,15.

<read_first>Accepted evidence/cost from Phases 129–140, current GitHub/GitLab CI authorities, TASK-233, supply-chain policy.</read_first>

**Consumes/Produces:** Accepted launcher profiles and native unit ownership → jobs `python-contract`, `python-api`, `python-integration`, `python-e2e-smoke`, `python-e2e-full`, `python-load` that all call `run-task.ps1`; native `backend-unit`/`frontend-unit` retain only their target roots. Mirror either calls the same launcher or is explicitly retired.

**Action/sentinel:** CI contract test fails on direct inner command, mutable image, missing limits/timeout/owner marker, zero collection, unexpected skip, missing sentinel/artifact, or cleanup leak. Seal SBOM/provenance and scanner digest; no runtime dependency downloads.

<verify><automated>`PlanId=141-01`, `TaskId=TASK-233`, owner `devops-engineer`, verifier `qa-auto-ver,security-auditor`, profile `ci-contract`, inner/outer `900/1080`.</automated></verify>

**Done:** every Python CI job calls the sole launcher with immutable provenance, limits, timeout, owner marker and non-zero collection/artifact/cleanup gates; native unit jobs own only `backend/tests/**` or `frontend/tests/**`, and no runtime download/direct inner command remains.

**Atomic commit:** `ci(testframework): включить greenfield QA profiles`; stage exactly the five Files paths above; body TASK-233. This task consumes the central profile registry read-only; shared docs are owned only by Task 141-C.

### Task 141-B: Run clean-room and kill/replay acceptance

**Files:** Create `.planning/audits/test-rebase/final-verification.md`; runtime evidence only in ignored `testframework/reports/<task-run>/**`.

**Requirements:** `TASK-220`; approved spec §§2,5–15.

<read_first>All phase summaries/evidence; clean full HEAD; image/lock/SBOM/provenance; threat register and rollback section below.</read_first>

**Consumes/Produces:** Exact full HEAD and accepted CI/profile graph → conclusion-blind verifier matrix: clean uncached image build, serial selftests, calibrated xdist if accepted, persisted/fresh seeds, all layer pilots, kill/timeout/OOM/cleanup, report/redaction and zero-owned-leftover proof.

**Action/sentinel:** Independent verifier receives specs/current tree/raw outputs but not author conclusions. A single missing profile, zero collection, stale TC, unexpected skip, leaked resource, redaction canary, mutable digest or non-reproducible seed keeps G141 non-passing.

<verify><automated>`PlanId=141-02`, `TaskId=TASK-220`, owner `qa-auto-ver`, verifier `arch-ver,security-auditor,team-lead`, profile `clean-room-full`, exact central-registry inner/outer `7200/7500`.</automated></verify>

**Done:** clean-room result is independently reproduced on exact full HEAD with every accepted profile non-zero, persisted/fresh-seed replay, kill/timeout/OOM/redaction/cleanup sensitivity and zero owned leftovers; statuses are per-gate, never collapsed.

**Atomic commit:** `docs(testframework): зафиксировать clean-room acceptance`, stage only final-verification document; body TASK-220.

### Task 141-C: Complete docs/onboarding and removable handoff

**Files:** Modify `AGENTS.md`, `.planning/STATE.md`, `.planning/COVERAGE.md`, `docs/test_plan.md`, `docs/testframework_strategy.md`, `docs/stack_and_tools.md`, `docs/security_guidelines.md`, `docs/wiki/DEPLOY.md`, `testframework/docs/architecture.md`, `testframework/README.md`; Create `testframework/docs/authoring.md`, `testframework/docs/environments.md`, `testframework/docs/triage.md`.

**Requirements:** `TASK-220`; approved spec §§1–15.

<read_first>G141 raw acceptance, approved architecture spec all sections, docs-sync matrix, current docs drift.</read_first>

**Consumes/Produces:** Accepted clean-room receipt and final package/profile graph → 10-minute clean-room onboarding, five worked examples (case→test, API, integration cleanup, UI component, load validity), profile/triage/run evidence guide, dependency removal/upgrade owners and final handoff.

**Action/sentinel:** Every documented path/profile is checked against exact revision; every executable example is the canonical outer launcher. Remove all obsolete root layer maps/direct runner commands/parity claims. A fresh engineer follows README without private knowledge and reaches the health evidence; independent observer records time and blockers.

<verify><automated>`PlanId=141-03`, `TaskId=TASK-220`, owner `docs-writer`, verifier `qa-auto-ver,arch-ver`, profile `onboarding-clean-room`, accepted launcher inner/outer `1200/1440`. Any nonexistent path/command or direct inner runner fails.</automated></verify>

**Done:** a fresh engineer reaches sealed health evidence through one canonical launcher within 10 measured minutes; all five examples resolve to existing paths/profiles, all obsolete root/direct/parity guidance is absent, and removal/upgrade/triage owners are explicit.

**Atomic commit:** `docs(testframework): завершить greenfield QA handoff`; stage exactly the thirteen expanded Files paths above; body TASK-220. No push/merge/deploy.

## 17. Coverage, traceability and evidence oracle

### Coverage policy

| Scope | Gate |
|---|---|
| Changed authn/authz/session/tenant/security-critical unit code | `line ≥95%`, `branch ≥90%`, negative matrix + mutation/failing sentinel |
| Changed business/domain/state-machine unit code | `line ≥90%`, `branch ≥85%`, boundary/error/idempotency branches |
| Changed pure UI logic/store/hook/reusable utility | `line ≥85%`, `branch ≥80%`, state transitions + failure behavior |
| Non-unit Python suites | all critical/high requirements have observable automated evidence; applicable scenario families covered or approved N/A; no universal line target |

Excluded generated/declarative/trivial wiring/unreachable defensive branches require reviewed path-level rationale. Package-wide historical numbers may be reported but cannot force unrelated legacy test writing.

### Case/evidence chain

`testframework/cases/<domain>/TC-*.md` is source of intent. Python test contains only one `case_id`; plugin projects task/requirement/domain/risk into JUnit and optional adapters. Required scenario review: happy, invalid, authn/authz/tenant, boundary/equivalence, edge/corner/time/concurrency, dependency failure/timeout/partial completion/retry/recovery/idempotency, security/privacy/redaction, a11y and load where applicable.

Every run report records contract version, full HEAD, manifest/platform image digests, lock/config hashes, target attestation, CPU/RAM/swap/PID/nofile/shm, seed/workers, collection/pass/fail/skip, first causal event + raw exits, durations, exact owned resources, artifact digests/redaction and cleanup proof. The evidence oracle order is: trusted contract → focused sentinel/test → integration/recovery exercise → conclusion-blind semantic review; author self-review is not completion proof.

## 18. Threat model

| ID | STRIDE | Severity | Boundary | Mitigation / gate |
|---|---|---:|---|---|
| T-QA-01 | Spoofing | high | task/run ownership | 128-bit nonce + exact ID/name/labels + durable owner manifest; foreign cleanup sentinel |
| T-QA-02 | Tampering | critical | sealed cutover path/identity | descriptor-relative no-follow, single link, SHA/size/device/inode recheck, atomic no-replace, 173/1,120/max56 reconciliation |
| T-QA-03 | Repudiation | high | owner → verifier | separate writer/verifier receipts, conclusion-blind verifier, full HEAD/digest/transition chain |
| T-QA-04 | Information disclosure | critical | config/log/report/browser | env allowlist, production deny, central idempotent redaction, canary scans, owner-only report path |
| T-QA-05 | Denial of service | high | Docker/pytest/browser/load | explicit resources + inner/outer watchdogs + bounded capture + first-exit + exact finally cleanup |
| T-QA-06 | Elevation of privilege | critical | host → container/daemon | non-root, read-only, cap-drop ALL, no-new-privileges, no socket/privileged/host namespace, approved egress only |
| T-QA-07 | Tampering | high | dependency/image supply chain | package legitimacy, exact lock/digests/platform, wheel/sdist review, SBOM/provenance, no runtime install |
| T-QA-08 | Tampering | high | traceability/skip policy | strict schema/markers/collection, expired exceptions fail, zero collection/unexpected skip fail |
| T-QA-09 | Information disclosure | critical | protected zone/dirty worktree | prefix exclusion and zero read/write to `pro-qa-ai/**`; allowlisted staging; dirty overlap stop |
| T-QA-10 | Repudiation | high | status aggregation | distinct `NOT_READY|NOT_RUN|PASS|FAIL`; no status promotion without named gate receipt |

Any unresolved critical/high threat blocks its task and downstream dependents. Low threats may be accepted only with explicit owner/rationale; no generic acceptance.

## 19. Rollback and recovery

- G129-P authority materialization: identical accepted-launcher rerun is idempotent; different existing authority fails closed. Roll back only its narrow authority commit, re-establish G128-D receipt on the resulting full HEAD, re-accept G129-L if launcher bytes changed, then rerun G129-P. The historical Phase 128 wrapper never publishes.
- Phase 129: one owner mutation commit per batch; rollback is recorded `git revert <commit-sha>` in reverse dependency order followed by new nonce-bearing root validation. Never restore a copied legacy tree or use broad filesystem deletion.
- Runtime failure: first causal result remains primary; cleanup failure is secondary unless child passed, in which case result is infra failure. Exact residual resource IDs are quarantined; no next task starts.
- Dependency/image update: revert `testframework/pyproject.toml`, `testframework/uv.lock`, `testframework/docker/images.lock.json` image/platform digests and scanner pin together; never patch a built container.
- TC/schema/plugin update: revert schema/policy/plugin and its selftests together; derived case index is regenerated, never edited.
- Optional tool Trial: adapter/profile/files are removable without changing domain facade/oracles; rejection removes the Trial in one allowlisted commit and retains portable JUnit/JSON evidence.
- Clean-room failure: classify with test triage and return to the earliest failing gate; do not mark later gates pass from historical evidence.

## 20. Multi-source coverage audit

| Source | Section/item | Covered by | Status |
|---|---|---|---|
| GOAL | founder-approved portable root Python QA project | 129–141 | COVERED |
| GOAL | safe exact cutover | 128–129 | COVERED |
| CONTEXT | active authority/routing canon before downstream planning | G-CANON exact 74-path closure: 70 active files + 4 versioned evaluation artifacts | COVERED |
| CONTEXT | planning-only bootstrap before implementation | META-PLAN/G-META | COVERED |
| CONTEXT | architecture authority/digest reseal before historical guard | G128-R | COVERED |
| REQ | TASK-202/203/204 | 128 | COVERED |
| REQ | TASK-193/197/198/201/219/234 | 129 | COVERED |
| REQ | TASK-206/218 | 130 | COVERED |
| REQ | TASK-205/207 | 131 | COVERED |
| REQ | TASK-208/209/227 | 132 | COVERED |
| REQ | TASK-210/211 | 133 | COVERED |
| REQ | TASK-212/213/228 | 134 | COVERED |
| REQ | TASK-214/215/229 | 135 | COVERED |
| REQ | TASK-216/230 | 136 | COVERED |
| REQ | TASK-194/217 | 137 | COVERED |
| REQ | TASK-195/196 | 138 | COVERED |
| REQ | TASK-231/232 | 139 | COVERED |
| REQ | TASK-200 | 140 | COVERED |
| REQ | TASK-220/233 | 141 | COVERED |
| RESEARCH | hybrid domain × layer, progressive skeleton, tool Adopt/Trial/Avoid, Docker state machine, spikes/removal | §§1–19 | COVERED |
| CONTEXT | package boundary, no import rewrite, first health slice, exact coverage floors, no empty dirs | §§1–5, 9–17 | COVERED |
| CONTEXT | 176 tasks/173 IDs/1,120 ops/max56 immutable | §§2,6,9 | COVERED |
| CONTEXT | separate readiness gates and `NOT_READY` truth | §§5,8,9,16 | COVERED |
| CONTEXT | protected/dirty/no Git mutation in planning run | Global Constraints, threat model | COVERED |

No deferred product storage/authz migration entered implementation scope; report sections about Garage/OpenFGA/SpiceDB remain future decision evidence only.

## 21. Self-review results

| Check | Result |
|---|---|
| Approved spec §§1–15 | Each section maps to architecture, tasks, policies, threats, DoD or required spikes above; no unplanned mandatory item found |
| ROADMAP phases 128–141 | All listed TASK IDs and formal dependencies mapped |
| Bounded Phase 130–141 task set | Exactly 61 tasks: `4+4+4+3+8+18+3+3+5+3+3+3`; each has one `Files`, `Requirements`, `<read_first>`, `Consumes/Produces`, fail-first sentinel/negative case, registered Docker-only profile, measurable `Done` and atomic commit; expanded task ownership is at most 14 paths |
| Bootstrap closure counts | G-CANON is exactly 74 paths (`70` audited active paths + `4` evaluation artifacts); G-META is exactly 23 paths (`21` live plans + checker + receipt); G128-R remains exactly 5 authority files |
| Profile registry closure | Task 130-A owns exactly 47 unique later rows and six exact limit classes; all 61 verify blocks resolve, using exactly those 47 IDs; union with Phase 129 bootstrap-only entries is 52 unique IDs; no later registry writer exists |
| Mandatory docs closure | Phases 130–140 each end with one bounded literal docs/traceability closure; Phase 141-C owns final docs/handoff; non-closure domain tasks own no shared docs path; every docs mutation/stage path is present in that task's `Files` |
| Founder execution scope | Current run stops at independently reviewed G130-H after cutover/scaffold, launcher/profile/framework contracts and unchanged health tracer; every Phase 131–141 contract remains separately `NOT_READY`/on-demand and no universal coverage sprint is authorized |
| Read-only revision scan | `CANON=74/74 existing=70 planned=4`; `META=23/23 existing=21 planned=2`; `G128R=5`; `TASKS=61 schema_bad=0`; `MAX_EXPANDED_PATHS=14 over14=0`; `REGISTRY=47/47 used=47 unknown=0`; `DOC_CLOSURES=12 owner_bad=0`; founder stop heading `1`; forbidden incomplete/path phrases `0`; completion marker `1` |
| Phase 129 signatures/arithmetic | 176 tasks, 173 unique batches, 1,120 operations, max 56 repeated consistently; no ID/path reconstruction instructed |
| Namespace/path consistency | Installable imports are `hub_testkit.*`; physical outputs live under `testframework/src/hub_testkit/**`, `testframework/tests/**`, `testframework/cases/**`, `testframework/selftests/**`, `testframework/docker/**`, `testframework/load/**` and `testframework/policies/**`; no target `testframework/core/**` or repo-root framework output is declared |
| First-slice signatures | `HubSettings`, `RuntimeIds`, `OwnershipLedger`, `Redactor`, `TestCaseRecord`, `HealthApi.fetch`, `RunEnvelope/execute/main` used consistently |
| Marker/status consistency | `full` remains profile, `load` remains separate runtime; `NOT_READY`, `NOT_RUN`, `PASS`, `FAIL` are distinct |
| Incomplete-marker scan | No standard unfinished-work token, vague catch-all action, deferred wiring or empty future-directory instruction remains in task bodies |
| Forbidden architecture scan | No BaseTest/god helper/singleton/global rerun/fixed sleep/runtime dependency install/host package mutation/broad cleanup is planned; G128-WPS51 only seals the existing signed Windows inbox PowerShell 5.1 component and exact compatibility remediation |
| Protected/non-overlap | `backend/tests/**` and `frontend/tests/**` are native-unit-only; `testframework/**` is Python non-unit/framework; `pro-qa-ai/**` excluded |
| Execution during this repair | No installation, Docker, test, validator, Python or Git command was executed; mutations are limited to the master plan, Phase 128 plan and G-META receipt hash fields |

### Independent checker resolution matrix

| Finding | Verdict | Resolution / reasoned pushback | Source evidence |
|---|---|---|---|
| Prior checker acceptance missed root active-canon drift | CONFIRMED; prior acceptance REVOKED | Added G-CANON as the first hard predecessor. Its exact 74-path closure rewrites 70 audited active routing/ownership paths, deletes the unsafe reorganize script, and adds four versioned scenario/schema/checker/receipt artifacts. Only enumerated historical/negative references survive; docs/architecture/security reviewers independently review deterministic BASE-vs-candidate declaration conformance. | Active audit found 70 edit-required files across `AGENTS.md`, `.cursor/**`, current docs/planning/backlog, `.github/labeler.yml` and the obsolete script; approved spec §§3–5. |
| Phase 128 architecture contract/seals still encode direct layers/core | CONFIRMED | Added G128-R after G-META and before G128-S. Exact five-file authority reseal updates approved physical semantics and derived contract-digest bindings while proving ledger/cohort/batch identities unchanged. No new digest is invented in the plan. | `.planning/audits/test-rebase/test-architecture-contract.json` currently lists `core/api/integration/contract/e2e`; both cutover builder and independent validator bind its current digest. |
| Historical wrapper used as pre-launcher publication path | CONFIRMED | Removed. G128-S statically accepts only the existing no-argument read-only guard; G128-D permits one guard run and conclusion-blind review; G129-L accepts the sole target launcher; only then G129-P materializes authority through that same launcher. Wrapper never claims `hub.qa-runner/v1` and is forbidden after G129-L. | `.planning/audits/test-rebase/validate-contract.ps1` has `param()` and explicit manifest-absence assertions; approved spec §§5–6 names `testframework/docker/run-task.ps1` as sole target entrypoint. |
| G128-S source envelope was partly untracked/dirty | CONFIRMED | Added G128-E between G128-R and G128-S. It first-tracks the wrapper/applier/founder manifest, commits the accepted inventory generator/ledger bytes, binds the six existing read-only authorities to the same full HEAD and requires independent architecture/security review before any static or runtime acceptance. | Current Git state after G128-R: three untracked and two modified writable envelope paths; wrapper pins executable blobs from HEAD and consumes the founder manifest. |
| G128-D prelaunch used an unsealed host interpreter lookup | CONFIRMED; attempt NON-CONSUMED | Added G128-WPS51 after initial G128-S and before G128-D. It installs nothing: an exact committed host lock seals the existing signed inbox PowerShell 5.1 absolute path, SHA256, file/product/PowerShell/CLR version, Authenticode and scoped two-hardlink System32/WinSxS identity. Renewed G128-E applies the independently enumerated PS5.1 remediation and current-process identity binding; renewed G128-S seals the literal command; only then can G128-D run. | The historical unqualified `pwsh` attempt is documentary, non-executable and non-consumed; no trusted interpreter identity or accepted G128-D receipt existed. Current status is `NOT_READY`, runtime `NOT_RUN`, and the premature command-resolution failure does not spend the one-shot allowance. |
| `129-01-PLAN.md` still targets old scaffold | CONFIRMED | META-PLAN rewrites `129-01-PLAN.md` and cross-links in `129-02-PLAN.md` … `129-07-PLAN.md` before execution; stale token absence and progressive-target presence are hard G-META assertions. | Current `.planning/phases/129-tf-e2e-shared/129-01-PLAN.md` contains `requirements.lock`, `pytest.ini`, root layer packages, `Dockerfile`, `compose.yaml` and `.gitkeep`; approved spec §§3,5. |
| No Phase 130–141 plan files exist | REJECTED | All twelve current execution authorities exist as live `PLAN.md`; META-PLAN lists each exact path and treats it as stale revision input. No `*-PLAN.md` rename is introduced. | Exact existing paths `.planning/phases/130-tf-traceability-gates/PLAN.md` through `.planning/phases/141-tf-ci-final-verification/PLAN.md`; 130–136 historical `VERIFICATION-SKIPPED.md` are non-authoritative. |
| Phase 130–141 tasks lack explicit Requirements/Done in master | CONFIRMED | All 61 master tasks now carry exact `Files`, Task IDs/spec sections, `<read_first>`, `Consumes/Produces`, verify profile, measurable `Done` and atomic allowlist; META-PLAN requires the same schema in each live downstream plan. | Master §§10–16 and approved spec §12 phase intent / §14 DoD. |
| Master uses relative framework output paths | CONFIRMED | Every task artifact is rooted at `testframework/**`; G-META negative sentinel rejects repo-root `hub_testkit/tests/cases/selftests/docker/load` outputs and obsolete root-layer scaffold. | Master task `Files` declarations; approved spec §§3–5. |
| Phase 129 should be regrouped into smaller plan files | REJECTED as blocker | The 173 sealed batch IDs are already the atomic review/mutation units. Retaining seven numbered files avoids identity/order drift; regrouping creates no additional runtime isolation because launcher executes exactly one batch per task. No duplicate batch table or reconstructed operation list is added. | Sealed Phase 129 authority: 176 tasks, 173 unique ordered batch IDs, 1,120 operations, max 56; approved spec §12 preserves existing IDs/batches. |
| Master assumed downstream sync instead of producing it | CONFIRMED | G-CANON is Wave 0, META-PLAN is Wave 1, and G128-R is Wave 2; every guard/code/cutover/runtime task is `NOT_READY` until exact `74 + 23 + 5` closures have their named independent receipts. | Header, G-CANON, META-PLAN, G128-R and dependency graph. |
| Runtime PASS could be inferred from static/readiness review | CONFIRMED risk | `STATIC_REVIEW_PASS`, `NOT_READY`, `NOT_RUN`, `PASS` and `FAIL` remain distinct; G128-D is a guard receipt, not target-runner PASS; all target jobs wait for G129-L. | Readiness state machine and §7 canonical invocation contract. |
| B1: G-CANON could not version deterministic route evidence | CONFIRMED | Replaced infeasible sampling with exact scenario declarations in the three owned skill docs, a strict receipt schema, a stdlib evaluator and versioned receipt paths. Six literal BASE-versus-candidate route declarations yield twelve deterministic records; static conclusion-blind review is the bootstrap gate. Exact `canon-routing-contract` replay through accepted `hub.qa-runner/v1` is G-CANON-X and blocks G129-P. | Task G-CANON declaration/receipt/evaluator contract and G-CANON-X readiness row. |
| B2: G-META lacked an executable plan checker and bounded receipt | CONFIRMED | Expanded the closure to 23 paths with a literal-21-plan stdlib checker and 256-KiB receipt. The checker fails closed on input-set, required-field, forbidden-token/path and `176/173/1120/56` drift; exact `downstream-plan-contract` launcher replay is G-META-X. | META-PLAN `Files`, checker/receipt contracts and G-META-X readiness row. |
| B3: Tasks 134-A, 135-A and 135-B exceeded useful domain review boundaries | CONFIRMED | Domain/adapter aggregates remain split; current Phase 134 has eight tasks and Phase 135 has eighteen. Every implementation boundary has at most 14 expanded paths, literal ownership, original TASK IDs/spec sections, one risk sentinel, registered profile, bounded receipt and atomic allowlist; shared docs moved to Tasks 134-H/135-R. | Tasks 134-A–H and 135-A–R; dependency graph; expanded-path scan. |
| B4: Task 138-A used an open recursive UI wildcard and an implicit journey set | CONFIRMED | Three literal journeys and one recovery matrix remain independent; Task 138-E now owns the exact four-path docs closure. Every UI page/component is named physically, with its own TC, registered profile, receipt, sentinel and atomic commit. | Tasks 138-A–E and Phase 138 dependency graph. |
| B5: Later verify profiles were absent from the registry or written concurrently | CONFIRMED | Added bounded Task 130-A as the sole post-bootstrap registry writer. It enumerates all 47 later profile IDs with image key, exact limit class, deadline and literal target; all 61 verify blocks resolve to those rows, unknown/missing/drifted profiles fail closed, and Task 141-A consumes the registry read-only. | Task 130-A tables/selftests/receipt; Task 141-A `Files`; G-META profile-closure assertion. |
| B6: Phase 134 realtime aggregate exceeded the review boundary | CONFIRMED | Split the former aggregate into messenger (7 paths), notification/NATS (7 paths) and risk closure (2 paths), followed by a four-path docs closure. Each has its own TASK mapping, sentinel, registered profile, receipt and atomic commit. | Tasks 134-E–H; central registry rows `api-messenger`, `integration-notification-nats`, `api-realtime-closure`. |
| B7: Required docs mutations were vague or outside task allowlists | CONFIRMED | Audited all 61 tasks. Removed every combined unspecified-docs stage form; Phases 130–140 now end in one literal docs/traceability closure and Phase 141-C retains exact final handoff ownership. Non-closure domain tasks do not own shared docs; same-session phase completion requires the closure receipt. | Eleven Phase 130–140 docs closure tasks, Task 141-C, G-META mandatory-doc checker rule and docs ownership scan. |
| Founder reaffirmed skeleton/health-only current execution scope | LOCKED | Added G130-H as a hard execution stop after G129-C and Tasks 130-A–D. The detailed Phase 131–141 map is retained for planning continuity, but each phase/profile stays `NOT_READY` until a separate founder receipt activates one exact accepted feature/risk slice and G-META-X rechecks its live plan. | Header scope lock, Global Constraints, readiness table, dependency graph and G130-H activation contract. |

### Static source assertions for re-review

- Active-canon assertion: the G-CANON list contains exactly `9+10+10+3+11+25+2+4=74` literal unique paths: 70 pre-existing active files plus four planned versioned evaluation artifacts. The 70-path scoped scan has zero old active routing reference after only the named historical/negative exclusions; the four new artifacts have exact declaration/schema/receipt ownership and are not scan exclusions. The evaluator proves only exact declared routing-contract conformance between the fixed BASE blobs and candidate bytes, never future-LLM behavior.
- Existing live plan assertion: only after G-CANON, the G-META list contains exactly 23 literal unique paths: 21 pre-existing live plans plus the planned checker and bounded receipt. Phase 130–141 use `PLAN.md`, not an invented suffix; the checker rejects any missing/extra/duplicate plan and any `176/173/1120/56` drift.
- Authority-reseal assertion: only after G-META, the exact five G128-R files contain approved physical semantics and mutually matching derived contract-digest bindings; inventory/manifest/tool inputs remain read-only and reconcile to `1049 / 507-245-297+28 / 176-173-1120-max56`.
- Host-interpreter assertion: initial G128-S authorizes only G128-WPS51; no installation or package mutation is permitted. The committed lock seals exact `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`, held-handle volume serial/file index/reparse=false, signed inbox provenance, scoped two-hardlink identity and pinned Microsoft publisher/production-chain/code-signing EKU. Renewed G128-E/G128-S must bind the PS5.1-compatible hardened wrapper and race-free suspended-before-job child lifecycle before G128-D. Unqualified `powershell.exe`, every `pwsh`, PATH/PATHEXT discovery and fallback fail closed; the historical command-resolution attempt is documentary, non-executable, non-consumed and runtime remains `NOT_RUN`.
- Phase 129 planning assertion: stale scaffold tokens must disappear from executable task paths/actions while exact ordered identities remain `176 tasks / 173 batch IDs / 1,120 operations / max 56`.
- Task-schema assertion: exactly 61 headings exist, distributed `4/4/4/3/8/18/3/3/5/3/3/3`; every block has one non-empty `Files`, `Requirements`, `<read_first>`, `Consumes/Produces`, registered profile, measurable `Done` and atomic commit boundary.
- Profile assertion: the central table contains 47 unique later IDs; all 61 verify blocks resolve to exactly those 47 IDs, each row has image key, one of six exact non-zero limit classes, inner/outer deadline and literal target; no writer after Task 130-A and no implicit/default profile remains.
- Bounded-path assertion: all 61 expanded `Files` sets are at most 14 paths. Tasks 134-E/F/G own exactly `7/7/2` paths; Tasks 138-A–C name every identity/tracker/messenger UI page/component, with no recursive UI wildcard or implicit journey set.
- Docs assertion: exactly eleven Phase 130–140 closure tasks plus Phase 141-C own shared docs; every docs mutation/stage path occurs literally in its task `Files`, non-closure domain tasks own none, and each phase closure has a `docs-contract` receipt before completion.
- Scope assertion: dependency graph contains one explicit `STOP current implementation run` at G130-H; Phase 131–141 are described as on-demand backlog, each inherits exact activation receipt precondition, and neither pre-registered profile IDs nor accepted G130-H automatically advance a later phase.
- Path assertion: framework task outputs begin with `testframework/` (or an explicitly allowlisted repository planning/docs/CI/native-unit path); import namespace `hub_testkit.*` is not a physical output root.
- Incomplete-work assertion: static read-only scan finds no standard unfinished-work marker, substitute action, deferred wiring, empty-future-directory instruction or unbound execution command.
- Runtime assertion for this planning revision: `NOT_RUN`; no Docker, test, build, lint, validator or Git mutation was executed.

## 22. Execution handoff

Before implementation, a new independent gsd-plan-checker re-reviews this master plan; every earlier acceptance remains revoked. Execution order is exact: one writer completes the 74-path G-CANON closure (70 active files plus four versioned evaluation artifacts); its BASE-vs-candidate evaluator proves only the exact declared routes in three skill docs, not future-LLM behavior. One planning writer then completes the 23-path G-META closure (21 live plans plus checker and bounded receipt); one authority writer completes the five-file G128-R reseal; one envelope writer commits and reviews the initial G128-E source envelope; initial G128-S may then authorize only G128-WPS51. G128-WPS51 installs nothing and commits the exact signed Windows inbox PowerShell 5.1 host lock plus independently accepted compatibility-remediation contract without running the guard. A renewed G128-E then modifies/commits only the wrapper for PS5.1 compatibility and current-process identity binding, and renewed G128-S independently seals the new bytes plus literal absolute invocation; only this renewed receipt permits the single G128-D run. The checker inspects exact `74 + 23 + 5` predecessor allowlists, the host TCB chain, authority hash closure, scoped zero-old-active-reference sentinel, 61-task backlog distribution, 47-row later-profile registry, maximum-14 path bound, Phase 134 split, literal docs-closure ownership and the G130-H execution stop. The current run ends after G130-H; later execution requires a separately accepted founder activation receipt. Executors use `superpowers:subagent-driven-development` for isolated task ownership or `superpowers:executing-plans` for sequential sealed batches; every task halts with `NOT_READY` until its named gate is independently accepted.

## PLANNING COMPLETE

Master meta-plan is ready for a new independent plan-checker re-review; no acceptance is claimed and every prior acceptance remains revoked. It adds the no-install G128-WPS51 inbox-interpreter TCB and compatibility chain without changing exact `74 + 23 + 5` predecessor closures, Phase 129 `176/173/1120/max56`, the 47 later profiles or the 61-task Phase 130–141 distribution. Current G128-D status is `NOT_READY`, runtime is `NOT_RUN`, and the historical prelaunch command-resolution incident is documentary, non-executable and non-consumed. Current execution still stops at G130-H; Phases 131–141 remain `NOT_READY` on-demand backlog.
