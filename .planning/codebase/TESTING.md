# Testing Patterns

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

**Run Commands:**
```bash
# Current backend canonical
cd backend
go generate ./ent
go vet -tags hubtest ./...
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...

# Current frontend
cd frontend
npm test                 # watch
npm run test:run         # one pass
npm run test:e2e:smoke
npm run test:e2e

# Current GitHub coverage shape
cd backend
go test -tags hubtest -coverprofile=/tmp/cov -coverpkg=./internal/... ./testframework/...
go tool cover -func=/tmp/cov

# Target after cutover
cd backend && go test -race -count=1 -timeout=600s ./tests/...
cd frontend && npm run test:run
cd testframework && pytest --strict-markers
cd testframework && locust --headless -f load/locustfile.py
```

## Test File Organization

**Location:**
- Current backend unit/API/integration tests are mixed under `backend/testframework/<domain>/`. Shared sqlite setup is `backend/testframework/env/builder.go`; the shared factory layer is `backend/testframework/factory/user.go`.
- Current frontend logic/component tests are under `testframework/stages/unit/frontend-test/domains/<domain>/`; shared render setup is `testframework/stages/unit/frontend-test/render.tsx`.
- Current E2E is under `testframework/stages/e2e/domains/<domain>/` with `testframework/stages/e2e/shared/{fixtures,helpers,pages,steps}/`.
- Target native unit locations are `backend/tests/<domain>/` and
  `frontend/tests/<domain>/`, with frontend helpers in `frontend/tests/support/`.
  Target Python project uses
  `testframework/{core,api,integration,contract,e2e,load,docker,reports}/`.
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
testframework/{core,api,integration,contract,e2e,load,docker,reports}/
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
- Real network in unit tests. `backend/testframework/github/preview_wave103_test.go` has environment-dependent network skips and should be reclassified or use a deterministic local server.

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
- Target Go/TS support: `backend/tests/support/` and `frontend/tests/support/`.
- Target Python: reusable clients/factories in `testframework/core/`; layer/
  domain fixtures in nearest `conftest.py`. Root `conftest.py` owns environment,
  not domain state.
- Fixture acceptance: two identical consecutive runs, zero leaked records, and parallel execution without collisions, per TASK-194/TASK-199.

## Coverage

**Requirements:**
- `.github/workflows/ci.yml` declares overall Go statement coverage of `backend/internal/...` at >=82%; `.planning/COVERAGE.md` records historical security package targets >=80%.
- The current migration backend baseline is fixed at **82.0373%** on 2026-08-07 from a zero-failure Docker Go 1.25 race run; all 6/6 security package floors pass. Older 83.0/81.9/80.2/81.8 values are dated historical attempts.
- Current `coverage-gate` omits `-race` and `-covermode=atomic`, while `.cursor/playbooks/testing.md` requires both together. Freeze one green Docker command before any deletion, then use exactly the same instrumentation after migration.
- Frontend has no enforced coverage configuration/threshold, and `frontend/package.json` does not declare a Vitest coverage provider. TASK-198 must add a pinned provider, include production sources, define exclusions, and record a dated baseline.
- Python black-box tests do not contribute to Go's native `go test -coverpkg` profile. Moving Go handler/repository/integration tests to Python can preserve behavioral coverage while lowering the Go statement metric. Therefore the migration must either retain/add enough Go units to preserve >=82%, or explicitly redefine the gate with an approved black-box coverage collection method and a new non-comparable baseline.

**View Coverage:**
```bash
# Candidate frozen Go method; profile stays outside the repository.
go test -tags hubtest -race -count=1 -covermode=atomic \
  -coverpkg=./internal/... -coverprofile=/tmp/hub-cov.out ./testframework/...
go tool cover -func=/tmp/hub-cov.out

# Target frontend, after provider/config is installed.
cd frontend && npx vitest run --coverage
```

## Test Types

**Unit Tests:**
- Go service logic with manual mocks and pure helpers; move true units to `backend/tests/<domain>/`.
- TypeScript API wrappers, stores, hooks, libraries, and components move to `frontend/tests/<domain>/`; setup/render/fixtures move to `frontend/tests/support/`.
- Ownership remains developers (`be*`, `fe*`) with `be-ver`/`fe-ver`, per `.cursor/playbooks/testing.md`.

**Integration Tests:**
- Current backend tree mixes categories: 184 files use Ent/sqlite helpers, 260 use HTTP/Fiber testing, 58 use miniredis, 16 reference NATS patterns, 8 reference Postgres test env, and 2 reference Mongo test env; categories overlap.
- Before moving, classify every file as Go unit or Python API/integration/contract. Directory alone is not evidence of type.
- GitHub `backend-test` provides Mongo but not Postgres, so PostgreSQL tests can skip on PR; `coverage-gate` starts both only outside pull requests. A required real-Postgres migration job is needed before deletion.

**E2E Tests:**
- Current Playwright TS has 20 specs/53 cases with POM/Steps under `testframework/stages/e2e/`.
- Target Playwright Python must preserve all 20 scenario files' behavior and all 53 collected cases unless an inventory row records an approved consolidation/deletion.
- All 20 current spec files contain conditional/direct skip logic. CI must expose collected/passed/failed/skipped counts and fail on unexpected skips.
- Preserve login/a11y/critical smoke semantics, traces, screenshots, and failure reports from `frontend/playwright.config.ts` and `.github/workflows/ci.yml`.

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

| Gate | Current implementation | Migration acceptance |
|------|------------------------|----------------------|
| Go unit/race | `backend-test` runs `-tags hubtest -race ./testframework/...` | `./tests/...` collects expected tests, race is green, deliberate failing sentinel fails job |
| Go coverage | `coverage-gate` runs only outside PR | Green baseline method frozen; same instrumentation after move, or explicitly approved gate redefinition for Python black-box coverage |
| FE unit | tsc, Vitest, Biome | `frontend/tests/**` included by Vitest, tsc, and Biome; coverage recorded; linker removed |
| Python | Not present | Pinned runtime/dependencies; `pytest --strict-markers` collects non-zero tests; marker jobs separated |
| E2E | TS smoke/full | Python parity inventory complete; exact smoke contract; zero unexpected skips; artifacts uploaded |
| Load | Not present | Locust scheduled/manual only; production target rejected; p50/p95/error rate/degradation point recorded |
| Traceability | Mostly one-way names in `docs/TESTING.md` | Code metadata + exact path/name; Allure upload into Hub TMS; unmapped count zero |
| Entry points | GitHub/GitLab/Makefiles/manifests disagree | All commands agree, or non-canonical mirrors are explicitly retired |

## Migration Acceptance Gates

1. Freeze the 1,047-row inventory at HEAD `ee6d696c`: 830 Go / 3,667 tests (root 829/3,665 + server 1/2; 594 unit + 236 component), 180 Vitest / 519 cases + 3 support, 20 E2E / 53 cases + 14 support, 39 hooks and 0 Python.
2. Establish a clean Docker coverage baseline with zero failed packages. Record command, revision, runtime versions, wall time, overall/security percentages, skips, and profile hash in `.planning/COVERAGE.md`.
3. Dual-run legacy and replacement suites until parity. Each job asserts non-zero collection and reports passed/failed/skipped counts.
4. Preserve behavior with the mutation/failing-sentinel checks required by TASK-193 through TASK-201.
5. Do not delete `backend/testframework/` or root TS stages until every inventory row is accounted for, old/new jobs are green on a real PR, and rollback remains possible.
6. Treat Go coverage separately from Python behavioral coverage. Retain enough Go unit coverage for the current >=82% gate, or obtain explicit approval for a new black-box method/baseline before cutover.
7. Re-run Go race/unit/coverage, frontend type/lint/unit/coverage, pytest layers, Python E2E smoke/full, and non-production Locust. Unexpected skips, unmapped TC results, and coverage regression block cutover.
8. Synchronize `docs/TESTING.md`, `docs/test_plan.md`, `docs/testframework_strategy.md`, `.planning/COVERAGE.md`, `.planning/STATE.md`, `testframework/PIPELINE.md`, all CI files, Makefiles, and package scripts in the same cutover.

---

*Testing analysis: 2026-08-07*
