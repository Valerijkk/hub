# Testing Patterns

> **Superseded inventory note (2026-08-07):** числовой snapshot ниже был
> построен до двух последних unit-файлов. Текущий action authority —
> `.planning/audits/test-rebase/founder-greenfield-cutover-manifest.json`
> (1,049 rows / 4,246 declared, revision `5708671f`, SHA `72de6a81…`).
> Parity/dual-run отменены; 237 legacy component rows delete-not-port, а 245
> hook-dependent genuine units защищены TASK-234.

**Analysis Date:** 2026-08-07

## Test Framework

**Runner:**
- Go `testing` + `testify` 1.11.1 on Go 1.25.0: 830 total / 3,667 tests = legacy root 829/3,665 + server 1/2; taxonomy 594 unit + 236 component (31 API + 205 integration).
- Vitest 3.2.4 + RTL + jest-dom: 180 files / 519 cases = 156 unit + 24 static contract, plus 3 support rows.
- Playwright TypeScript 1.52.x: 20 specs / 53 cases plus 14 support rows.
- Target: native units in `backend/tests/<domain>/**` and
  `frontend/tests/<domain>/**`; pytest/httpx/Playwright Python/Locust directly
  in root `testframework/**`.
- Python target is absent: no root `testframework/pyproject.toml`/pytest config or `{core,api,integration,contract,e2e,load,docker,reports}/` layer exists.
- Config: `backend/go.mod`, `frontend/vite.config.ts`, `frontend/playwright.config.ts`, `.github/workflows/ci.yml`; target config belongs in `testframework/pyproject.toml` with registered strict markers.

**Assertion Library:**
- Go: `require` for prerequisites and `assert` for independent value checks.
- Frontend: Vitest `expect` plus jest-dom loaded from `frontend/src/test/setup.ts`.
- Browser: Playwright web-first `expect`.
- Target Python: pytest assertions and Playwright `expect`; Allure provides metadata/reporting, not pass/fail truth.

**Execution contract:**

- Legacy Go/Vitest/TypeScript Playwright commands are inventory only and must
  not be run as migration evidence or fallback.
- Target jobs run only through `/qa-gate` and the versioned
  `testframework/docker/run-task.ps1` contract (`hub.qa-runner/v1`).
- Until that launcher and `images.lock.json` exist and receive independent
  acceptance, runtime status is `NOT_READY`; host or inline Docker execution is
  forbidden.
- Every accepted invocation records immutable image provenance, outer timeout,
  CPU/RAM/swap/PID/network policy, first exit code, task-run owner marker and
  exact `finally` cleanup proof.

## Test File Organization

**Location:**
- Current backend unit/API/integration tests are mixed under `backend/testframework/<domain>/`. Shared sqlite setup is `backend/testframework/env/builder.go`; the shared factory layer is `backend/testframework/factory/user.go`.
- Current frontend logic/component tests are under `testframework/stages/unit/frontend-test/domains/<domain>/`; shared render setup is `testframework/stages/unit/frontend-test/render.tsx`.
- Current E2E is under `testframework/stages/e2e/domains/<domain>/` with `testframework/stages/e2e/shared/{fixtures,helpers,pages,steps}/`.
- Target native unit locations are `backend/tests/<domain>/` and
  `frontend/tests/<domain>/`, with frontend helpers in `frontend/tests/support/`.
  Target Python project uses
`testframework/{src/hub_testkit,tests/{api,integration,contract,e2e},selftests,cases,load,docker,reports}/`.
- The target above supersedes `backend/internal/<domain>/unit/` and `frontend/src/<module>/unit/` in `.planning/active/august-2026/05-AUGUST-DETAILED.md` and TASK-197/TASK-198; those plans must be corrected before execution.

**Naming:**
- Go: `snake_case_test.go`; `Test<Subject>_<Scenario>`; meaningful table-row names.
- Vitest: kebab-case `*.test.ts` / `*.test.tsx`; subject in `describe`, observable outcome in `it`.
- Current E2E: kebab-case `*.spec.ts`; Page Objects `*.page.ts`; business steps `*.steps.ts`.
- Target pytest: `test_<feature>.py`, `test_<behavior>_<expected_result>`; stable `HUB-TC-*` identity via Allure `AS_ID`/TMS link.

**Structure:**
```text
Current
backend/testframework/<domain>/*_test.go
testframework/stages/unit/frontend-test/domains/<domain>/*.test.ts(x)
testframework/stages/e2e/{shared,domains}/

Target
backend/tests/<domain>/*_test.go
frontend/tests/<domain>/*.test.ts(x)
frontend/tests/support/
testframework/{src/hub_testkit,tests/{api,integration,contract,e2e},selftests,cases,load,docker,reports}/
```

## Test Structure

**Suite Organization:**
```go
func TestUpdateIssue_RejectsStaleVersion(t *testing.T) {
	svc, repo := newTestService()
	issue := repo.issueWithVersion(2)

	_, err := svc.UpdateIssue(context.Background(), issue.ID, staleRequest(1))

	require.Error(t, err)
	assertAppErrorCode(t, err, http.StatusConflict)
}
```

```typescript
describe("auth-store", () => {
	beforeEach(() => {
		resetAuthStore();
		vi.restoreAllMocks();
	});

	it("logout clears tokens and authenticated state", () => {
		useAuthStore.getState().logout();
		expect(useAuthStore.getState().isAuthenticated).toBe(false);
	});
});
```

```python
@pytest.mark.api
@allure.label("AS_ID", "HUB-TC-TRCK-04")
def test_update_issue_rejects_stale_version(api_client, issue_factory):
    issue = issue_factory(version=2)
    response = api_client.update_issue(issue.id, version=1)
    assert response.status_code == 409
```

**Patterns:**
- Arrange with a new mock/resource namespace; act through a public contract; assert prerequisites before values and verify side effects/error paths.
- Use table-driven Go subtests and parametrized pytest cases for equivalence classes and boundaries.
- Cleanup belongs to `t.Cleanup`, `beforeEach`/`afterEach`, or the fixture that creates the resource.
- A compose/browser lifetime may be session-scoped; mutable domain data must be function-scoped and xdist-safe.
- Mutation evidence is required by `.planning/backlog/TASK-193-2026-07-30-tf-python-scaffold.md` through `.planning/backlog/TASK-201-2026-07-30-tf-ci-cleanup.md`.

## Mocking

**Framework:**
- Go uses hand-written repository mocks; there is no canonical generator. See `backend/testframework/tracker/` and `.cursor/playbooks/testing/backend-go.md`.
- Frontend uses `vi.mock` and `vi.spyOn`: 107 current files use `vi.mock`, 45 use `vi.spyOn` under `testframework/stages/unit/frontend-test/domains/`.
- MSW is recommended in `.cursor/playbooks/testing/frontend.md`, but zero current test files use `setupServer(...)`; do not claim MSW protection until a shared strict server exists.
- E2E uses real Hub HTTP/browser behavior with helpers in `testframework/stages/e2e/shared/helpers/`.

**Patterns:**
```go
type mockRepo struct {
	issues map[uuid.UUID]*ent.Issue
	updateVersionError bool
}

func newTestService() (*tracker.TrackerService, *mockRepo) {
	repo := newMockRepo()
	return tracker.NewTrackerService(repo, nil), repo
}
```

```typescript
vi.mock("@tanstack/react-router", () => ({
	Link: ({ children }: { children: React.ReactNode }) => <a href="#">{children}</a>,
}));
const logoutSpy = vi.spyOn(authApi, "logout").mockResolvedValue(undefined);
```

**What to Mock:**
- Unit layer: repositories, external clients, clock, browser APIs, animation/DnD/router dependencies outside the behavior under test.
- Mocks must reproduce not-found, conflict, forbidden, and infrastructure failures.
- Target Python system tests mock only third-party boundaries; Hub API, DB, NATS, and browser behavior are the integration target.

**What NOT to Mock:**
- The tested function/component.
- Ent/sqlite when verifying constraints, filters, cascades, or optimistic locking.
- Browser-visible behavior in E2E.
- Real network in unit tests. `backend/testframework/github/preview_wave103_test.go` is frozen legacy discovery and follows its sealed operation; do not remediate it. New target tests use deterministic doubles/local boundaries where the requirement permits.

## Fixtures and Factories

**Test Data:**
```go
e := env.New(t, "tracker").Build()
users := factory.NewUserFactory(t, e.Client, e.Ctx)
alice := users.Create("alice")
```

```typescript
const { queryClient } = renderWithProviders(<IssueCard issue={issue} />);
```

```python
@pytest.fixture
def project(project_factory):
    resource = project_factory.create(unique_suffix=uuid4().hex[:8])
    yield resource
    project_factory.delete(resource.id)
```

**Location:**
- Current Go: `backend/testframework/env/builder.go`, `backend/testframework/factory/user.go`, and domain-local helpers.
- Current frontend: `testframework/stages/unit/frontend-test/render.tsx`; global jest-dom setup is `frontend/src/test/setup.ts`.
- Current E2E: `testframework/stages/e2e/shared/fixtures.ts` plus helpers/POM/Steps.
- Target support: frontend-only helpers живут в `frontend/tests/support/`. Четыре legacy Go support-файла удаляются; новые Go unit helpers допускаются только domain-local и не мигрируют из старого support tree.
- Target Python: reusable clients/factories in exact task-named modules below `testframework/src/hub_testkit/`; layer/
  domain fixtures in nearest `conftest.py`. Root `conftest.py` owns environment,
  not domain state.
- Fixture acceptance: two identical consecutive runs, zero leaked records, and parallel execution without collisions, per TASK-194/TASK-199.

## Coverage

**Requirements:**
- `.github/workflows/ci.yml` currently contains a legacy >=82% job. Its
  **82.0373%** result from 2026-08-07 is dated discovery evidence, not a
  migration/acceptance gate.
- After cutover publish a new honest native-unit baseline. Do not preserve the
  historical denominator or infer Python behavior from Go statement coverage.
- Frontend has no enforced coverage configuration/threshold, and `frontend/package.json` does not declare a Vitest coverage provider. TASK-198 must add a pinned provider, include production sources, define exclusions, and record a dated baseline.
- Python black-box tests do not contribute to Go's native profile. Their gate is
  task/requirement/risk/TC coverage plus behavior/oracle evidence, not a promise
  to preserve the legacy Go percentage.

**Execution status:** target coverage jobs run only through the accepted
`hub.qa-runner/v1` outer launcher. Until it exists, verdict is `NOT_READY`; no
host command or inline Docker substitute is canonical evidence.

## Test Types

**Unit Tests:**
- Go service logic with manual mocks and pure helpers; move true units to `backend/tests/<domain>/`.
- Exactly 156 sealed TypeScript/React unit rows move to `frontend/tests/<domain>/` and only `render.tsx` moves to `frontend/tests/support/`. The 24 static contracts, two other support rows and legacy config closure delete; later helpers are greenfield.
- Ownership remains developers (`be*`, `fe*`) with `be-ver`/`fe-ver`, per `.cursor/playbooks/testing.md`.

**Integration Tests:**
- Current backend tree mixes categories: 184 files use Ent/sqlite helpers, 260 use HTTP/Fiber testing, 58 use miniredis, 16 reference NATS patterns, 8 reference Postgres test env, and 2 reference Mongo test env; categories overlap.
- Before cutover, classify every file only to select its sealed operation: genuine Go unit `move`, otherwise legacy `delete`. Directory alone is not evidence; no legacy row is ported into Python.
- GitHub `backend-test` historically provided Mongo but not Postgres, while `coverage-gate` started both only outside pull requests. Это discovery evidence, не условие удаления: новый real-Postgres profile создаётся greenfield в Phase 133 после принятия launcher/scaffold.

**E2E Tests:**
- Current Playwright TS has 20 specs/53 cases with POM/Steps under `testframework/stages/e2e/`.
- Target Playwright Python is designed from current requirements, risks and
  verified TC; the 20/53 historical counts are not preservation authority.
- All 20 current spec files contain conditional/direct skip logic; this is historical discovery, not a remediation queue. New Python E2E jobs independently expose collected/passed/failed/skipped counts and fail on unexpected skips.
- Reuse useful ideas only as new test design; do not port the TypeScript files or
  treat their artifacts/counts as acceptance.

## Common Patterns

**Async Testing:**
```typescript
const user = userEvent.setup();
await user.click(screen.getByRole("button", { name: "Сохранить" }));
await screen.findByRole("status");
```

```python
@pytest.mark.e2e
async def test_project_lifecycle(authenticated_page):
    await authenticated_page.get_by_role("button", name="Создать проект").click()
    await expect(authenticated_page).to_have_url(re.compile(r"/projects/[0-9a-f-]+"))
```

**Error Testing:**
```go
_, err := svc.GetIssue(ctx, projectID, "MISSING-1")
require.Error(t, err)
assertAppErrorCode(t, err, http.StatusNotFound)
```

```python
response = api_client.update_issue(issue_id, version=stale_version)
assert response.status_code == 409
assert response.json()["request_id"]
```

## CI Gates

The `Current implementation` column below is non-executable legacy inventory.
Every target gate is a named profile of the sole accepted `hub.qa-runner/v1`;
the table does not authorize raw commands, provider-specific runners or host
execution.

| Gate | Current implementation | Migration acceptance |
|------|------------------------|----------------------|
| Go unit/race | Historical `backend-test` | `backend-unit` launcher profile collects expected tests, race is green, deliberate failing sentinel fails job |
| Go coverage | legacy `coverage-gate` outside PR | Historical evidence only; publish a new native-unit baseline without a preservation threshold |
| FE unit | tsc, Vitest, Biome | `frontend/tests/**` included by Vitest, tsc, and Biome; coverage recorded; linker removed |
| Python | Not present | Launcher-owned Python profiles use pinned runtime/dependencies, strict registered markers and non-zero collection |
| E2E | TS smoke/full legacy inventory | New risk-derived Python smoke/full; zero unexpected skips; traceable TC and sanitized artifacts |
| Load | Not present | Locust scheduled/manual only; production target rejected; p50/p95/error rate/degradation point recorded |
| Traceability | Mostly one-way names in `docs/TESTING.md` | Code metadata + exact path/name; Allure upload into Hub TMS; unmapped count zero |
| Entry points | GitHub/GitLab/Makefiles/manifests disagree | All retained consumers call `hub.qa-runner/v1`; non-canonical entrypoints are retired |

## Migration Acceptance Gates

1. Seal the current 1,049-row ledger and exact move/keep/delete manifest with an
   independent descriptor-safe validator; no file operation before PASS.
2. Execute Phase 129 literal batches: move only genuine native units, delete
   legacy non-units/static-contract/TS-E2E closure, protect hook-dependent units
   until TASK-234, and switch exact consumers atomically.
3. Create the root Python project and `hub.qa-runner/v1`; until independent
   launcher acceptance every runtime verdict is `NOT_READY`.
4. Design new Python cases from current tasks/requirements/risks/verified TC;
   require happy/negative/authz/boundary/edge/failure-recovery analysis and a
   sensitive oracle.
5. Run target jobs twice where relevant, fail on zero collection/unexpected
   skip/orphan metadata/leaked secret or resource, and isolate artifacts in
   `reports/<task-run>/` with an owner marker.
6. Phase 141 performs clean-room verification, publishes new native-unit and
   requirement/risk baselines, and synchronizes all live docs/CI entrypoints.

---

*Testing analysis: 2026-08-07*
