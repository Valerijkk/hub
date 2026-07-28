# Phase 999.1 — Unified Test Framework & Test Tree Refactor

> **Status:** BACKLOG (captured 2026-06-21)  
> **Trigger:** promote after Phase 15 Wave 8 (coverage gate) or when E2E debt blocks velocity  
> **Agents:** `architect` → `qa` → `refactor-specialist` → `qa-ver` + `be-ver` + `fe-ver`

---

## Goal

Вынести **все** тесты Hub в отдельную структуру (`tests/` или `hub-tests/`) и построить **единый тестовый фреймворк** с Page Object Model, OOP-слоем, SOLID/DRY/KISS и явными техниками test design — без дублирования setup/fixtures между BE/FE/E2E.

---

## Problem (as-is)

| Layer | Location today | Pain |
|-------|----------------|------|
| Go unit/integration | `backend/internal/*/*_test.go` (co-located) | 100+ файлов, разные env helpers, дубли enttest/miniredis |
| Vitest/RTL | `frontend/src/**/*.test.tsx` (co-located) | 0 files, нет конвенции |
| Playwright E2E | `frontend/e2e/` | POM частично, fixtures разрознены |
| Coverage artifacts | `backend/cov_*`, `cover*` в корне backend | шум в git, нет единого отчёта |

Co-located tests работают для Go-идиомы, но **cross-cutting test infra** (shared builders, POM, data factories, CI matrix) не масштабируется к Phase 15 exit (50% BE + E2E FB-07..09 + Vitest wave).

---

## Target architecture

```
tests/
├── README.md                 # entry point, how to run each layer
├── framework/                # shared test framework (language-agnostic contracts)
│   ├── core/                 # base types, lifecycle, tagging (@smoke @regression)
│   ├── builders/             # test data builders (User, Project, Issue…) — Factory pattern
│   ├── fixtures/             # session-scoped fixtures (DB seed, auth tokens)
│   └── assertions/           # custom matchers / soft asserts
├── backend/                  # Go integration & contract tests (NOT co-located)
│   ├── suite/                # testify suite base, enttest wrapper, NATS/Redis helpers
│   ├── modules/              # mirror internal/ modules (auth, tracker, …)
│   └── smoke/                # API smoke against running server
├── frontend/                 # Vitest + RTL
│   ├── setup/                # renderWithProviders, MSW handlers
│   ├── components/
│   └── hooks/
├── e2e/                      # Playwright (move from frontend/e2e)
│   ├── pages/                # Page Objects (LoginPage, KanbanPage, …)
│   ├── flows/                # Business flows (UserJourney, not raw clicks)
│   ├── fixtures/             # auth, backendAvailable, devMode
│   └── specs/                # thin spec files → delegate to flows/pages
└── contracts/                # optional: OpenAPI/consumer-driven snapshots
```

**Go note:** unit tests *may* remain co-located (`*_test.go` рядом с кодом) — Go toolchain ожидает package-local tests. Phase scope:

- **Move:** integration, handler_flow, enttest-heavy, cross-module tests → `tests/backend/`
- **Keep co-located:** pure unit tests без I/O (fast, white-box)
- Document split in `docs/test_plan.md` § «Test tree layout»

---

## Design principles (mandatory)

| Principle | Application |
|-----------|-------------|
| **Page Object (POM)** | E2E: pages expose intent (`loginAsAdmin()`), specs — orchestration only |
| **Flow Object** | Multi-step journeys (onboarding → create project → invite) in `flows/` |
| **SOLID** | Single-responsibility pages/builders; Open/Closed via base `Page` + composition |
| **DRY** | One `AuthFixture`, one `TestDB`, one MSW profile — no copy-paste per spec |
| **KISS** | Specs ≤ 30 lines; complexity in framework, not in every test file |
| **Factory / Builder** | `UserBuilder`, `ProjectBuilder` — readable Given setup |
| **Arrange-Act-Assert** | Enforced structure; optional AAA comments only when non-obvious |
| **Equivalence partitioning** | Document input classes in builder tags (valid/invalid/boundary) |
| **Boundary value** | Rate limits, pagination limits, magic link TTL — explicit BVA cases |
| **State transition** | Issue workflow, user pending→active, call states — table-driven tests |
| **Pairwise** | Config matrices (roles × permissions × modes) where combinatorial explosion |

---

## Test design catalog (framework must support)

1. **Smoke** — `@smoke` tag, CI fast path (< 5 min)
2. **Regression** — full suite nightly / pre-release
3. **Contract** — handler response shape vs OpenAPI (Phase 15 swagger wave)
4. **Security** — auth bypass, RBAC negative paths, rate limit (reuse `.planning/audits/code-review-final/`)
5. **A11y** — axe in E2E layer (Phase 15 Wave 6)
6. **Flaky quarantine** — `@quarantine` + retry policy (Phase 16 prep)

---

## Migration waves (when promoted)

| Wave | Scope | Exit |
|------|-------|------|
| M1 | `tests/README` + `framework/core` + CI path filters | Docs + skeleton green |
| M2 | Extract Playwright → `tests/e2e/` + POM refactor | All existing e2e pass |
| M3 | Vitest setup + 10 component tests | `npm run test:unit` in CI |
| M4 | Go integration suite in `tests/backend/` | Auth, project, tracker smoke |
| M5 | Builders/fixtures parity BE↔E2E | Same seed data both layers |
| M6 | Deprecate duplicate helpers in `internal/*_test.go` | grep audit clean |

---

## Success criteria

1. Новый разработчик находит «как писать тест» за 1 README
2. E2E spec не содержит CSS-селекторов — только Page/Flow API
3. Auth setup — один вызов `fixtures.auth.loginAs(role)` во всех слоях
4. CI: `make test-smoke` / `make test-regression` / `make test-e2e`
5. `docs/test_plan.md` и `.planning/COVERAGE.md` обновлены под новое дерево
6. Zero regression: все существующие green tests остаются green после миграции

---

## Out of scope

- Замена Go co-located **pure unit** tests (остаются в `internal/`)
- Новый runner вместо Go test / Vitest / Playwright
- Test data in production DB

---

## Dependencies

- Phase 15 Wave 5 (E2E FB-07..09) — лучше мигрировать POM до добавления 3 новых specs
- Phase 15 Wave 8 (coverage 50%) — не блокировать, но align builders с coverage gaps

---

## References

- `docs/test_plan.md`, `docs/TESTING.md`
- `.cursor/rules/40-testing.mdc`
- `.cursor/skills/test-triage/SKILL.md`
- Playwright POM: `frontend/e2e/` (current partial impl)
