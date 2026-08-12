# Architecture

> **Superseded inventory note (2026-08-07):** counts ниже — pre-final snapshot.
> Для любых move/delete действий использовать только current founder manifest
> (revision `5708671f`, 1,049 rows, SHA `72de6a81…`); parity/dual-run не gate,
> hook-dependent genuine units нельзя удалять.

**Analysis Date:** 2026-08-07

## Pattern Overview

**Overall:** Мультиязычная тестовая архитектура, физически разделённая между Go-модулем `backend/`, внешним TypeScript-деревом `testframework/stages/` и исполняемыми конфигурациями в `frontend/` и `.github/workflows/ci.yml`.

**Key Characteristics:**
- Backend-тесты оформлены как внешние Go-пакеты `*_test` внутри того же модуля `github.com/valerii/hub/backend`: 829 файлов / 3,665 tests находятся в `backend/testframework/`, ещё один server test содержит 2 cases; total 830/3,667 и taxonomy 594 genuine unit + 236 component (31 API + 205 integration).
- Frontend tests физически вынесены из `frontend/src/`: 180 файлов / 519 cases находятся в `testframework/stages/unit/frontend-test/domains/` (156 unit + 24 static contract), ещё 3 support rows — рядом; запускаются из `frontend/` через `frontend/vite.config.ts`.
- E2E — 20 TypeScript Playwright specs / 53 cases в `testframework/stages/e2e/domains/` + 14 support rows; фактическая конфигурация и зависимости принадлежат `frontend/playwright.config.ts` и `frontend/package.json`.
- Корневой `testframework/` сейчас не является Python-проектом: в нём нет `pyproject.toml`, `conftest.py` и файлов `*.py`; его 241 tracked-файл описывают старую pipeline-oriented схему.
- Founder target для этого mapping: units в `backend/tests/<domain>/**` и `frontend/tests/<domain>/**`, frontend support в `frontend/tests/support/**`, а единый Python QA-проект — прямо в корневом `testframework/`. Пути `backend/internal/<домен>/unit/` и `frontend/src/<модуль>/unit/` в `.planning/active/august-2026/05-AUGUST-DETAILED.md` и TASK-197/198 требуют docs sync до исполнения.

## Layers

**Backend Go test layer (current):**
- Purpose: Unit, handler, repository, schema и интеграционные проверки backend в одном дереве.
- Location: `backend/testframework/`
- Contains: 829 `*_test.go` / 3,665 tests и четыре support-файла: `backend/testframework/env/builder.go`, `backend/testframework/factory/user.go`, `backend/testframework/mongotest/setup.go`, `backend/testframework/steps/http.go`.
- Depends on: `backend/internal/**`, `backend/ent/**`, testify, Fiber test API, SQLite/Ent, MongoDB, Redis и NATS в зависимости от теста.
- Used by: `.github/workflows/ci.yml` jobs `backend-test` и `coverage-gate`; declarative mirrors в `testframework/pipeline/jobs/backend-test.yaml` и `testframework/pipeline/jobs/coverage-gate.yaml`.

**Backend test-only production seam:**
- Purpose: Открывает приватные функции и поля внешним пакетам `backend/testframework/*_test`.
- Location: 39 файлов `backend/internal/**/export_hooks.go`; всего 45 Go-файлов в `backend/internal/` имеют build constraint `//go:build hubtest`.
- Contains: функции `Export*`, а также дополнительные tagged-файлы вроде `backend/internal/platform/middleware/export_requestid_hooks.go` и `backend/internal/tms/export_hooks_repo.go`.
- Depends on: сборочный тег `hubtest` во всех канонических backend-командах.
- Used by: 328 ledger rows имеют разрешённые hook consumers (889 links); 168 test-файлов также имеют `//go:build hubtest`.

**Frontend Vitest layer (current):**
- Purpose: API, stores, hooks, lib и React component tests.
- Location: `testframework/stages/unit/frontend-test/domains/`
- Contains: 180 test-файлов (60 `.test.ts`, 120 `.test.tsx`; 519 cases), из них 156 unit и 24 static contract; ещё 3 support rows (`render.tsx` + 2 fixtures).
- Depends on: 171 test-файл импортирует production-код через alias `@/`; 18 используют `@testframework`; 62 используют общий renderer/query-client из `testframework/stages/unit/frontend-test/render.tsx`.
- Used by: `frontend/vite.config.ts` с external include и alias, scripts `test`/`test:run` в `frontend/package.json`, job `frontend-test` в `.github/workflows/ci.yml`.

**Playwright E2E layer (current):**
- Purpose: Browser smoke и full critical-path regression.
- Location: 20 specs / 53 cases в `testframework/stages/e2e/domains/`; 14 support rows (13 shared-файлов + `frontend/playwright.config.ts`).
- Contains: Page Objects в `testframework/stages/e2e/shared/pages/`, Steps в `testframework/stages/e2e/shared/steps/`, availability/auth helpers в `testframework/stages/e2e/shared/fixtures.ts` и `testframework/stages/e2e/shared/helpers/`.
- Depends on: `@playwright/test` из `frontend/node_modules`, dev-login/seed API из `testframework/stages/e2e/shared/helpers/auth.ts`, backend/frontend URLs.
- Used by: `frontend/playwright.config.ts`, scripts `test:e2e*` в `frontend/package.json`, jobs `e2e-smoke`/`e2e-full` в `.github/workflows/ci.yml`.

**Pipeline metadata layer (current):**
- Purpose: Документирует stage → job → domain topology.
- Location: `testframework/pipeline/stages.yaml`, `testframework/pipeline/jobs/*.yaml`, `testframework/PIPELINE.md`.
- Contains: Пять job manifests и stage ordering.
- Depends on: ручную синхронизацию с CI.
- Used by: документация; исполняемых non-document consumers для `testframework/pipeline/` не обнаружено. Исполняемый источник правды — `.github/workflows/ci.yml`.

**Language-native unit trees (approved target, not implemented):**
- Purpose: Отделить developer-owned Go/Vitest unit tests от независимого QA framework.
- Location: founder target `backend/tests/<domain>/**`, `frontend/tests/<domain>/**` и `frontend/tests/support/**`; каталоги сейчас отсутствуют.
- Contains: только доказанные genuine Go `*_test.go` под `backend/tests/<domain>/**`; четыре legacy support-файла не переносятся. Новые минимальные Go helpers допустимы только рядом с unit-доменом и только если их явно создаёт принятый план. Vitest `*.test.ts(x)` и frontend-only support живут под `frontend/tests/`.
- Depends on: `backend/go.mod`/`backend/internal/**` для Go и `frontend/package.json`/`frontend/src/**` для TypeScript.
- Used by: целевые language-native CI jobs; `.github/workflows/ci.yml`, `.planning/backlog/TASK-197-2026-07-30-tf-backend-units-colocated.md` и `.planning/backlog/TASK-198-2026-07-30-tf-frontend-units-colocated.md` ещё не синхронизированы с этими exact paths.

**Python QA framework (approved target, not implemented):**
- Purpose: Integration, API, system E2E, full-flow и load tests вне языковых unit-слоёв.
- Location: целевые `testframework/{src/hub_testkit,tests/{api,integration,contract,e2e},selftests,cases,load,docker,reports}/`, `testframework/pyproject.toml` и `testframework/tests/conftest.py`.
- Contains: pytest fixtures, Playwright Python, Locust; сейчас эти пути отсутствуют.
- Depends on: Docker-managed Postgres/Mongo и приложение через публичные API; Python-слой не должен импортировать Go packages или frontend source.
- Used by: целевая отдельная CI job из `.planning/backlog/TASK-201-2026-07-30-tf-ci-cleanup.md`; такой job сейчас отсутствует в `.github/workflows/ci.yml`.

## Data Flow

**Backend test flow:**

1. `.github/workflows/ci.yml` выполняет `go generate ./ent` и `go vet -tags hubtest ./...` из `backend/`.
2. `go test -tags hubtest ... ./testframework/...` компилирует внешние пакеты `backend/testframework/*_test` внутри модуля из `backend/go.mod`.
3. Тест импортирует `backend/internal/<domain>`; Go internal visibility разрешена, потому что importer остаётся под `backend/`.
4. При обращении к приватным деталям tagged-файл `backend/internal/**/export_hooks*.go` предоставляет `Export*` API.
5. Coverage job в `.github/workflows/ci.yml` добавляет `-coverpkg=./internal/...`, поэтому покрытие production-пакетов измеряется из внешнего test tree.

**Frontend unit flow:**

1. `frontend/package.json` запускает `frontend/scripts/link-testframework-node-modules.mjs`.
2. Скрипт создаёт ignored junction/symlink в `testframework/stages/unit/frontend-test/node_modules`, указывающий на `frontend/node_modules`.
3. `frontend/vite.config.ts` включает только `../testframework/stages/unit/frontend-test/domains/**/*.test.{ts,tsx}`, исключает `src/**/*.test.{ts,tsx}` и резолвит `@/` в `frontend/src/`.
4. Vitest трансформирует внешние test-файлы; `frontend/tsconfig.json` включает только `frontend/src`, поэтому отдельный `npx tsc --noEmit` не проверяет root test tree.

**E2E flow:**

1. `frontend/package.json` создаёт symlink `testframework/stages/e2e/node_modules` и запускает `frontend/playwright.config.ts`.
2. Config назначает `testframework/stages/e2e/domains/` как `testDir`; smoke project выбирает три spec, full project — все 20.
3. `testframework/stages/e2e/shared/helpers/auth.ts` использует DEV seed/login и внедряет session bootstrap в browser context.
4. `e2e-full` в `.github/workflows/ci.yml` поднимает инфраструктуру и backend через `scripts/ci-e2e-backend.sh`; `e2e-smoke` запускает frontend-only subset.

**State Management:**
- Go SQLite state изолируется UUID DSN и `t.Cleanup` в `backend/testframework/env/builder.go`; Mongo helper создаёт отдельную БД и удаляет её в `backend/testframework/mongotest/setup.go`.
- Vitest renderer создаёт новый QueryClient на тест в `testframework/stages/unit/frontend-test/render.tsx`.
- Playwright shared layer проверяет доступность окружения и использует dev seed/session injection в `testframework/stages/e2e/shared/fixtures.ts` и `testframework/stages/e2e/shared/helpers/auth.ts`.

## Dependency Directions

```text
backend/testframework/*_test
    -> backend/internal/<domain>
    -> backend/internal/**/export_hooks*.go  [only with -tags hubtest]
    -> backend/ent + SQLite/Mongo/Redis/NATS/HTTP as needed

testframework/stages/unit/frontend-test
    -> frontend/vite.config.ts
    -> frontend/src via @/*
    -> frontend/node_modules via generated junction

testframework/stages/e2e
    -> frontend/playwright.config.ts
    -> frontend/node_modules via generated junction
    -> browser -> frontend -> backend -> infrastructure

testframework/pipeline/*.yaml
    -X-> CI executor (documentation only)
.github/workflows/ci.yml
    -> actual Go/Vitest/Playwright commands
```

Use only these directions while the current topology is active. Do not add imports from production source into `testframework/pipeline/`, and do not place Go code that imports `backend/internal/**` outside the `backend/` tree.

## Key Abstractions

**Go external-package tests:**
- Purpose: Проверяют public API и не компилируются как часть production package.
- Examples: `backend/testframework/auth/handler_admin_extra_test.go`, `backend/testframework/tracker/automation_consumer_test.go`.
- Pattern: `package auth_test` / `package tracker_test` плюс импорт соответствующего `backend/internal/<domain>`.

**Hubtest export hooks:**
- Purpose: Ограниченный test-only доступ к private implementation.
- Examples: `backend/internal/auth/export_hooks.go`, `backend/internal/tracker/export_hooks.go`, `backend/internal/tms/export_hooks_repo.go`.
- Pattern: `//go:build hubtest`; production `-tags release` не включает hooks.

**Backend builders/factories:**
- Purpose: DRY setup для Ent, users, Mongo и HTTP assertions.
- Examples: `backend/testframework/env/builder.go`, `backend/testframework/factory/user.go`, `backend/testframework/mongotest/setup.go`, `backend/testframework/steps/http.go`.
- Pattern: unique test data + cleanup; Mongo absence сейчас приводит к `t.Skip`, а не hard failure.

**Frontend TestRenderer:**
- Purpose: Единый React Query provider без shared cache между тестами.
- Examples: `testframework/stages/unit/frontend-test/render.tsx` и его consumers через `@testframework`.
- Pattern: factory per test, optional wrapper injection.

**Page Object + Step:**
- Purpose: Отделяет browser mechanics от business flow.
- Examples: `testframework/stages/e2e/shared/pages/base.page.ts`, `testframework/stages/e2e/shared/steps/auth.steps.ts`.
- Pattern: domain spec → Step/POM → Playwright Page.

## Entry Points

**GitHub backend unit/coverage:**
- Location: `.github/workflows/ci.yml`
- Triggers: PR/push/manual rules workflow.
- Responsibilities: `backend-test` исполняет только `backend/testframework/...`; `coverage-gate` измеряет `backend/internal/...` через `-coverpkg`.

**GitHub frontend unit:**
- Location: `.github/workflows/ci.yml`
- Triggers: PR/push.
- Responsibilities: npm install, `tsc`, symlink setup, Vitest, Biome.

**GitHub E2E:**
- Location: `.github/workflows/ci.yml`
- Triggers: smoke после frontend test, full после backend+frontend test.
- Responsibilities: Playwright install, infrastructure/backend lifecycle для full, report artifacts.

**Local frontend scripts:**
- Location: `frontend/package.json`
- Triggers: `npm run test:run`, `npm run test:e2e:smoke`, `npm run test:e2e`.
- Responsibilities: Создать symlink и вызвать Vitest/Playwright с `frontend/vite.config.ts` / `frontend/playwright.config.ts`.

**Secondary entry points with drift:**
- Location: `.gitlab-ci.yml`, `Makefile`.
- Triggers: GitLab pipeline или local `make test`.
- Responsibilities: Сейчас используют `go test ./...` без `-tags hubtest`; `.gitlab-ci.yml` также ссылается на несуществующие spec в `frontend/e2e/`. Не считать их эквивалентными GitHub CI до синхронизации.

## Error Handling

**Strategy:** Тестовый слой использует fail-fast assertions для проверок, но часть environment checks допускает skip.

**Patterns:**
- Go tests используют `testing.T`, testify `require`/`assert` и `t.Cleanup`; примеры находятся в `backend/testframework/env/builder.go` и `backend/testframework/steps/http.go`.
- `backend/testframework/mongotest/setup.go` вызывает `t.Skip`, если Mongo недоступна; такой тест не доказывает интеграционный gate без проверки числа skipped tests.
- `testframework/stages/e2e/shared/fixtures.ts` предоставляет `skipUnlessBackend`/`skipUnlessDevMode`/`skipUnlessFrontend`; full CI должен дополнительно доказывать доступность сервисов до Playwright.
- Python target из `.planning/backlog/TASK-193-2026-07-30-tf-python-scaffold.md` требует явной ошибки подключения вместо timeout/skip для обязательной инфраструктуры.

## Cross-Cutting Concerns

**Logging:** Go test output идёт через `go test`; Playwright report создаётся под `frontend/playwright-report/` и загружается `.github/workflows/ci.yml` только при failure.
**Validation:** Go — `go vet -tags hubtest`; frontend production source — `frontend/tsconfig.json` и Biome по `frontend/src`; root TypeScript tests валидируются Vitest, но не `tsc` include.
**Authentication:** Текущий E2E использует DEV seed/login в `testframework/stages/e2e/shared/helpers/auth.ts`; этот helper удаляется. Greenfield auth fixture выводится из актуального auth/security contract и обеспечивает per-test data isolation.
**Coverage:** Старый `.github/workflows/ci.yml` хранит только historical discovery evidence. Значение 82.0373% не является migration gate; после cutover target jobs публикуют новый baseline через accepted launcher.

## Current-to-Target Gaps

| Area | Current evidence | Approved target | Required seam |
|---|---|---|---|
| Go unit placement | 832 total / 3,674 declared; 595 classified genuine unit, 237 component; 0 in `backend/tests/` | Founder target `backend/tests/<domain>/**` only for units | Sealed manifest moves 350 hook-free units and keeps 245 hook-dependent units blocked until TASK-234 |
| Frontend unit placement | 180 files (156 unit + 24 static contract) + 3 support; 0 under `frontend/tests/` | Exactly 156 units move to `frontend/tests/<domain>/**` and one `render.tsx` moves to `frontend/tests/support/` | Delete 24 static contracts and the remaining two legacy support rows; new Python contracts are independent greenfield tests |
| Root framework | 241 tracked legacy files; 0 Python files | Python project in `testframework/` | Phase 129 deletes the exact legacy closure and creates the root scaffold; no coexistence/parity gate |
| E2E source | 20 TS specs in `testframework/stages/e2e/domains/` | New risk-derived Playwright Python in `testframework/tests/e2e/` | Delete TS rows by manifest; do not port files or preserve historical counts |
| Integration classification | Unit/integration/API mixed in `backend/testframework/` | Legacy non-unit rows удаляются; новые Python suites создаются независимо в `testframework/tests/integration/` и `testframework/tests/api/` | Classification выбирает только literal `move` genuine unit или `delete`; Python не является destination старого row |
| Load | No files/config under `testframework/load/` | Locust in `testframework/load/` | Add only local/scheduled execution per `TASK-200` |
| CI | Go + Vitest + TS Playwright | language-native unit jobs + Python layer jobs + scheduled load | TASK-201 switches consumers per cutover batch; TASK-219 deletes legacy immediately by manifest, without dual-run |

## Migration Decision Gates

**G1 — Go directory semantics:**
- `backend/tests/**` остаётся отдельным Go package tree, как и текущий `backend/testframework/**`; перенос туда не даёт доступа к private symbols в `backend/internal/<domain>`.
- Поскольку `backend/tests/**` находится внутри модуля из `backend/go.mod`, он может импортировать `backend/internal/**`, но должен сохранить `package <domain>_test`/public imports и необходимые `backend/internal/**/export_hooks*.go`.
- Удаляйте export hook только после нулевого поиска его consumers и зелёного доменного прогона; утверждение в `.planning/backlog/TASK-197-2026-07-30-tf-backend-units-colocated.md` о прямом private-доступе после перемещения неприменимо к founder target `backend/tests/**`.

**G2 — Test classification:**
- `backend/testframework/` is not a pure unit layer: overlapping scans find 182 files using Ent test/open patterns, 27 with Postgres patterns, 12 with Mongo patterns, 60 with Redis patterns, 49 with NATS patterns and 259 with `httptest` patterns.
- Use direct and transitive dependency evidence only to select the sealed
  operation: service-free genuine unit `move`, hook-dependent unit
  `keep-blocked`, service-backed/whole-API legacy row `delete`. New Python
  integration/API/contract suites are independent greenfield artifacts.

**G3 — Source path authority:**
- **Historical/forbidden pointer only:** actual legacy E2E source was `testframework/stages/e2e/domains/`; `frontend/e2e/` is a tracked README whose old direct-layer pointer is not execution or ownership authority.
- `.planning/audits/planning-audit-2026-07/02-TEST-FRAMEWORK-MIGRATION.md` uses `/tests`, while the later active plan and TASK-193 use root `testframework/`. Use `testframework/` unless the founder explicitly changes the active plan.
- `.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md` is the current routing source. Older microservice/test-path sections in `05-AUGUST-DETAILED.md` are explicitly non-executable provenance.

**G4 — Destructive cleanup:**
- Phase 128 seals an exact current-SHA move/keep/delete authority; Phase 129 is
  the only destructive cutover owner.
- Legacy Go non-units, static-contract Vitest, TypeScript Playwright and their
  support/config/entrypoints are delete-no-port rows. Only genuine native units
  move; 245 hook-dependent Go units remain protected until TASK-234 proves a
  public/injected seam and zero consumers.

## Safe Migration Order

1. Accept Phase 128 descriptor-safe authority, independent validator, negative
   sentinels and sealed cutover manifest on the current SHA.
2. Execute Phase 129 as ordered batches: move 350 hook-free Go units and 157
   frontend unit/support rows, delete 325 legacy rows/closure files, and update
   every affected consumer atomically.
3. Refactor and move the remaining 245 genuine Go units only after TASK-234
   zero-consumer evidence; delete 43 hook files only through their guarded
   batches.
4. In the same Phase 129 program create the single root Python scaffold and its
   versioned outer launcher. Until independent launcher acceptance, runtime is
   `NOT_READY`; legacy commands are not fallback.
5. Build new contract/API/integration/E2E/load coverage from current
   requirements, risks and verified TC. Historical TS/Go cases may inform test
   design but are not port/parity authority.
6. Phase 141 runs clean-room target jobs, publishes honest new baselines and
   proves zero stale paths/entrypoints/resources.

## Reproducibility

Snapshot: HEAD `ee6d696c`, measured 2026-08-07; ledger SHA-256 `01dae4e2…`.

Run from repository root; `rg --files` respects `.gitignore`, and the explicit globs exclude generated/vendor dependency trees:

```powershell
@(rg --files backend/testframework -g '*_test.go' -g '!**/vendor/**' -g '!**/node_modules/**' -g '!**/generated/**').Count
@(rg --files backend/internal -g '*_test.go' -g '!**/generated/**').Count
@(rg --files testframework/stages/unit/frontend-test -g '*.test.ts' -g '*.test.tsx' -g '!**/node_modules/**').Count
@(rg --files testframework/stages/e2e/domains -g '*.spec.ts' -g '!**/node_modules/**').Count
@(rg --files testframework -g '*.py' -g '!**/node_modules/**' -g '!**/generated/**').Count
@(rg --files backend/internal -g 'export_hooks.go' -g '!**/generated/**').Count
@(git ls-files -- testframework).Count
```

Expected counts for this snapshot: `830`, `0`, `180`, `20`, `0`, `39`, `241`; ledger total `1,047`, SHA-256 `01dae4e2…`.

---

*Architecture analysis: 2026-08-07*
