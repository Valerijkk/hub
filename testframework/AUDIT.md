# Hub Test Framework — Audit (2026-06-22)

> Аудит применения техник test design и паттернов после миграции в `testframework/`.

## Матрица паттернов

| Паттерн | E2E | FE Unit | BE Unit | Статус |
|---------|-----|---------|---------|--------|
| **Page Object Model** | ✅ `e2e/pages/` | N/A | N/A | Login, AppShell, Messenger |
| **Step definitions** | ✅ `e2e/steps/` | N/A | N/A | AuthSteps, NavigationSteps |
| **AAA (Arrange-Act-Assert)** | ✅ specs | ✅ unit | ✅ *_test.go | Обязательно в новых тестах |
| **Builder** | N/A | N/A | ✅ `testframework/env` | Fluent Ent env |
| **Factory** | N/A | N/A | ✅ `testframework/factory` | UserFactory |
| **TestRenderer (DI)** | N/A | ✅ `frontend/render.tsx` | N/A | renderWithProviders |
| **Fixtures (Playwright)** | ✅ `fixtures.ts` | N/A | N/A | DI, не Singleton |
| **DRY** | ⚠️ partial | ⚠️ partial | ⚠️ partial | 450 BE файлов — постепенная миграция на Builder |
| **KISS** | ✅ | ✅ | ✅ | Specs = orchestration only |
| **SOLID SRP** | ✅ pages/steps/specs | ✅ render vs test | ✅ env/factory/steps |
| **Singleton** | ❌ intentionally avoided | ❌ | ❌ | Scoped per test (correct) |

## Test design techniques

| Техника | Где применена | Пример |
|---------|---------------|--------|
| **Equivalence partitioning** | auth, tracker | valid/invalid magic code, issue types |
| **Boundary values** | ratelimit, validation | title min 3 chars (`issue-form.test.tsx`) |
| **State transition** | tracker workflow | status category todo → in progress |
| **Error guessing** | handler tests | 401/403/404/409/501 paths |
| **Pairwise** | TMS env | dev/staging/prod/custom labels |
| **Decision table** | RBAC middleware | role × resource × action |

## Go unit-тесты (450 файлов)

**Идиома Go:** `*_test.go` остаются в `internal/*` (white-box доступ).

| Anti-pattern (найден) | Рекомендация | Миграция |
|-----------------------|--------------|----------|
| Дубли `newXxxTestEnv` в каждом пакете | `env.New(t, prefix).Build()` | Pilot: `tracker/repository_test.go` |
| `newNotificationTestEnv`, `newProfileTestEnv` | UserFactory + Builder | Phase 121+ |
| Прямой `enttest.Open` без cleanup | Builder encapsulates | ✅ Builder |
| wave*_test.go дубли имён | grep перед добавлением | Process rule |

## E2E (22 specs)

| Spec | POM | Steps | AAA |
|------|-----|-------|-----|
| login.spec.ts | ✅ | ✅ AuthSteps | ✅ |
| critical-gate.spec.ts | ✅ | ✅ | ✅ |
| messenger.spec.ts | ✅ MessengerPage | ✅ AuthSteps | ✅ |
| Остальные 19 | ⚠️ helpers | ⚠️ | ✅ | Миграция на POM — backlog TF-02 |

## Frontend unit (25 files)

| Критерий | До | После |
|----------|-----|-------|
| Расположение | `frontend/src/**` | `frontend/testframework/unit/**` |
| Render setup | дубли QueryClient | `renderWithProviders` |
| Imports | relative `./` | `@/` aliases |

## Backlog миграции (TF-XX)

| ID | Задача | Приоритет |
|----|--------|-----------|
| TF-01 | Migrate remaining 19 e2e specs to POM+Steps | P2 |
| TF-02 | Replace top-20 `newXxxTestEnv` with Builder | P2 |
| TF-03 | `httptest` handlers → `steps.HTTPAct` | P3 |
| TF-04 | Component tests → `renderWithProviders` (all 25) | P2 |
| TF-05 | Page objects: Admin, TMS, Projects | P2 |

## Verdict

**Framework:** ✅ структура создана, паттерны задокументированы, pilot-тесты мигрированы.  
**Full compliance:** ⚠️ ~15% E2E на POM, ~8% FE unit на TestRenderer, BE Builder — pilot only.  
**Next:** постепенная миграция по TF-backlog без big-bang.
