# Coding Conventions

**Analysis Date:** 2026-08-07

## Naming Patterns

**Files:**
- Текущие Go-тесты используют `snake_case_test.go` внутри доменных пакетов, например `backend/testframework/tracker/handler_flow_test.go` и `backend/testframework/auth/magiclink_test.go`. Идентификаторы задач иногда встроены в имя (`backend/testframework/tracker/link_target_scope_a3011_test.go`, `backend/testframework/tracker/automation_soft_delete_task191_test.go`); при механическом переносе сохраняйте имена для проверки parity.
- Новые Go unit-файлы называйте по устойчивому поведению: `service_issue_test.go`, `handler_issue_test.go`. Идентификатор задачи или TC указывайте в имени теста/метаданных, но не используйте его как единственный смысл имени файла. Принятый target — `backend/tests/<domain>/`.
- Текущие frontend-тесты используют kebab-case `*.test.ts` / `*.test.tsx`, например `testframework/stages/unit/frontend-test/domains/tracker/safe-url.test.ts` и `testframework/stages/unit/frontend-test/domains/tracker/issue-card.test.tsx`. Сохраняйте это правило после переноса в `frontend/tests/<domain>/`; этот target заменяет `frontend/src/<module>/unit/` из `.planning/active/august-2026/05-AUGUST-DETAILED.md` и `.planning/backlog/TASK-198-2026-07-30-tf-frontend-units-colocated.md`.
- Текущие browser specs используют kebab-case `*.spec.ts`, Page Object — `*.page.ts`, бизнес-шаги — `*.steps.ts`; примеры: `testframework/stages/e2e/domains/tracker/project-lifecycle.spec.ts`, `testframework/stages/e2e/shared/pages/login.page.ts`, `testframework/stages/e2e/shared/steps/auth.steps.ts`.
- Целевые Python system tests именуйте по pytest discovery: `test_<feature>.py`, функции `test_<behavior>_<expected_result>`, fixtures — в `conftest.py` или `fixtures/<domain>.py`. Размещайте их прямо в `testframework/{api,integration,contract,e2e}/`; Locust-код — в `testframework/load/`.

**Functions:**
- Go-тесты именуйте `Test<Subject>_<Scenario>`; образцы: `TestHandler_CreateAndGetIssue` в `backend/testframework/tracker/handler_flow_test.go` и `TestEntRepository_CreateIssue_IncrementsCounter` в `backend/testframework/tracker/repository_test.go`.
- Go helpers используют lowerCamelCase и вызывают `t.Helper()`, если ошибка должна указывать на caller; образцы: `setupTrackerHandlerApp` в `backend/testframework/tracker/handler_flow_test.go`, `New`/`Build` в `backend/testframework/env/builder.go`.
- Vitest suite называет subject в `describe(...)`, наблюдаемое поведение — в `it(...)`; следуйте `testframework/stages/unit/frontend-test/domains/auth/auth-store.test.ts` и `testframework/stages/unit/frontend-test/domains/tracker/safe-url.test.ts`.
- Python-тесты называйте по наблюдаемому результату, а не по внутреннему шагу: `test_update_issue_rejects_stale_version`, не `test_call_update_method`.

**Variables:**
- Go использует короткие общеупотребимые locals (`ctx`, `svc`, `repo`, `req`, `resp`) и содержательные fixture IDs (`projectID`, `userID`, `statusID`), как в `backend/testframework/env/builder.go`.
- TypeScript использует camelCase, неизменяемые fixtures через `const` и данные доменной формы, например `baseIssue` в `testframework/stages/unit/frontend-test/domains/tracker/issue-card.test.tsx`.
- Python fixtures называйте по выдаваемому ресурсу (`api_client`, `project_factory`, `authenticated_page`) и не создавайте анонимное глобальное состояние.

**Types:**
- Локальные Go test structs используют lowerCamelCase (`trackerTestEnv`, `mockRepo`), общие helpers — экспортируемый PascalCase (`Env`, `Builder`, `UserFactory`) в `backend/testframework/env/builder.go` и `backend/testframework/factory/user.go`.
- TypeScript types/interfaces используют PascalCase; образец опций тестового render — `TestRenderOptions` в `testframework/stages/unit/frontend-test/render.tsx`.
- Python fixture/result models используют PascalCase-классы и snake_case-поля; API response schemas должны быть явными, не произвольными untyped dictionaries.

## Code Style

**Formatting:**
- Go-код должен проходить `gofmt`; imports группируются как standard library, third-party, project. В текущем CI нет отдельного formatting gate для test tree, поэтому TASK-197 должен добавить его для `backend/tests/**` рядом с `go vet` в `.github/workflows/ci.yml`.
- TypeScript-тесты следуют Biome-настройкам из `frontend/biome.json`; текущий внешний test tree форматирован табами.
- `frontend/biome.json` включает только `src/**`, поэтому 180 файлов в `testframework/stages/unit/frontend-test/domains/` и 20 specs в `testframework/stages/e2e/domains/` не lintятся. После TASK-198 включите `frontend/tests/**`. Python formatter/linter не обнаружен; выберите и зафиксируйте его в `testframework/pyproject.toml` до приёмки TASK-193.
- Комментарии и prose пишите по-русски, identifiers и test code — на идиоматичном английском, согласно `docs/code_style_guide.md` и `.cursor/playbooks/testing.md`.

**Linting:**
- Текущий frontend lint — `npm run lint`, то есть `biome check ./src` из `frontend/package.json`; значимые правила accessibility, unused variables, non-null assertions и explicit `any` заданы в `frontend/biome.json`.
- Текущая backend static check — `go vet -tags hubtest ./...` в `.github/workflows/ci.yml`. Сохраняйте tag `hubtest`, пока ссылки на `backend/internal/**/export_hooks.go` не удалены осознанно.
- Не оставляйте `fmt.Println`/`console.log` как постоянный test output. Используйте assertion messages, `t.Logf` для диагностического контекста, captured logging pytest и CI artifacts согласно `.cursor/playbooks/testing.md`.
- Зарегистрируйте все Python markers и включите strict checking. Минимальный словарь: `api`, `integration`, `contract`, `e2e`, `smoke`, `regression`; Locust запускается отдельно из `testframework/load/`.

## Import Organization

**Order:**
1. Standard library (`context`, `testing`, `node:path`).
2. Внешние библиотеки (`testify`, React Testing Library, Vitest, Playwright, pytest/allure/httpx).
3. Hub imports (`github.com/valerii/hub/backend/...`, `@/...`, локальные Python packages).

**Path Aliases:**
- Frontend production imports используют `@/* -> frontend/src/*` из `frontend/tsconfig.json` и `frontend/vite.config.ts`.
- Текущие test helpers могут импортироваться через `@testframework` из `frontend/vite.config.ts`; удалите alias вместе с внешним TypeScript tree, если реальных consumers не останется.
- Сейчас внешний TS tree получает зависимости через junctions из `frontend/scripts/link-testframework-node-modules.mjs`. Целевые `frontend/tests/**` находятся внутри frontend package и должны использовать `frontend/node_modules` напрямую; удаление linker — acceptance check `.planning/backlog/TASK-201-2026-07-30-tf-ci-cleanup.md`.
- Python должен импортироваться из корневого проекта `testframework/` без правок `sys.path` в отдельных тестах. Layout и invocation фиксируются в `testframework/pyproject.toml`.

## Error Handling

**Patterns:**
- В Go применяйте `require.NoError`/`require.Error` до dereference, `assert` — для независимых value checks. Typed Hub errors проверяйте по `AppError.Code`, не по хрупкому тексту, согласно `.cursor/playbooks/testing/backend-go.md`.
- Test doubles обязаны воспроизводить repository contracts: not-found, conflict, forbidden, infrastructure errors. Mock, всегда возвращающий `nil, nil`, недопустим; ориентиры — `backend/testframework/tracker/repository_test.go` и `.cursor/playbooks/testing/backend-go.md`.
- В TypeScript async tests ожидайте user actions и observable state. Rejected promises должны проверяться, а не подавляться ради green suite.
- Python fixtures при недоступной обязательной инфраструктуре должны завершать CI setup error, не молча `skip`. Optional local-only capabilities требуют зарегистрированного marker и allowlisted причины skip.
- Каждый FAIL сначала triage по требованиям с фиксацией `ROOT=code|test|unclear|flaky`, как требуют `.cursor/REGULATION.md` и `.cursor/playbooks/testing.md`.

## Logging

**Framework:** assertion output, Go `testing.T`, Playwright HTML/trace, целевой pytest/Allure captured logging.

**Patterns:**
- Логируйте только данные для воспроизведения: test ID, request ID, sanitized endpoint, fixture identity и timing.
- Никогда не прикладывайте JWT, magic-link codes, passwords, API keys, webhook secrets или PII. Это относится к Playwright traces, pytest logs, Allure attachments и Locust output; правила — `.cursor/playbooks/testing/tooling.md` и `docs/security_guidelines.md`.
- При переводе E2E на Python сохраните trace-on-first-retry и screenshot-on-failure из `frontend/playwright.config.ts`.
- Для значимых suite runs записывайте wall time и самые медленные тесты в `.planning/COVERAGE.md`; деградация более 20% требует заметки в `.planning/STATE.md`, согласно `docs/testframework_strategy.md`.

## Comments

**When to Comment:**
- Объясняйте бизнес-инвариант, причину fixture, mutation proof или trace ID. Ориентиры: комментарий об storage invariant в `testframework/stages/unit/frontend-test/domains/auth/auth-store.test.ts` и mutation note в `testframework/stages/unit/frontend-test/domains/tracker/safe-url.test.ts`.
- Не пересказывайте setup/action/assert и имя функции. Общие автогенерированные комментарии вроде «проверяет сценарий» не несут полезного контекста.
- Каждый automated acceptance test должен иметь стабильный `TC-<MODULE>-NN`/`HUB-TC-*` в code metadata или соседнем комментарии; точный путь и test name фиксируются в `docs/TESTING.md`.

**JSDoc/TSDoc:**
- Используйте JSDoc для общих fixtures, Page Objects, Steps и неочевидных helpers, например `testframework/stages/e2e/shared/fixtures.ts` и `testframework/stages/unit/frontend-test/render.tsx`.
- Самим component test cases обычно JSDoc не нужен: поведение выражают `describe`/`it`.
- Python shared fixtures получают docstrings, если scope, cleanup или security boundary неочевидны; обычный test function должен объяснять себя именем.

## Function Design

**Size:** Один тест — один behavioral contract. Разбивайте multi-endpoint flow, только если flow сам не является acceptance contract. Для равнозначных cases используйте table-driven Go subtests и parametrized pytest.

**Parameters:**
- Go helpers принимают `*testing.T` первым и вызывают `t.Helper()`; ресурсы очищаются через `t.Cleanup()`, как в `backend/testframework/env/builder.go`.
- React render helpers принимают явные options и создают новый QueryClient, как в `testframework/stages/unit/frontend-test/render.tsx`.
- Python fixture scope соответствует стоимости: session scope для одного compose lifecycle/browser installation, function scope для mutable domain data, worker-safe namespaces для pytest-xdist.

**Return Values:**
- Helpers возвращают typed resources; assertions остаются в тесте, кроме явно названных assertion helpers, связанных с test object.
- Не возвращайте неоднозначные `nil, nil`/`None` из mocks для error-bearing contracts.
- Ответственность за cleanup принадлежит fixture, создавшей ресурс.

## Module Design

**Exports:**
- Текущие Go-тесты — внешние packages (`<domain>_test`); целевой
  `backend/tests/<domain>/` остаётся отдельным Go package path и не получает
  private access. Сохраняйте hooks, пока конкретный unit не переписан на public behavior.
- Экспортируйте только минимальный white-box seam для поведения. Не расширяйте production API ради тестов.
- Общие frontend helpers `renderWithProviders` и `createTestQueryClient` сейчас находятся в `testframework/stages/unit/frontend-test/render.tsx`; после переноса держите одну каноническую реализацию в `frontend/tests/support/`.
- Общие Python helpers размещайте в `testframework/core/` или layer-specific
  fixture modules. Domain setup не должен накапливаться в root `conftest.py`.

**Barrel Files:**
- E2E Page Objects и Steps сейчас используют barrels `testframework/stages/e2e/shared/pages/index.ts` и `testframework/stages/e2e/shared/steps/index.ts`.
- В Python избегайте широких fixture barrels; основной механизм — scoped `conftest.py` discovery.

## Migration-Preserving Rules

- До перемещения создайте machine-readable inventory: каждый текущий путь сопоставляется с `go-unit`, `python-api`, `python-integration`, `python-contract`, `python-e2e` или `delete-with-reason`. Источники: `backend/testframework/`, `testframework/stages/unit/frontend-test/domains/`, `testframework/stages/e2e/domains/`.
- На механическом этапе сохраняйте test names и assertions. Refactoring/translation разрешается после parity старого и нового запусков.
- Не классифицируйте тест по текущей директории: в `backend/testframework/` 184 файла используют Ent/SQLite helpers, 260 — HTTP/Fiber testing, 58 — miniredis, 16 — NATS patterns; категории пересекаются.
- Каждый новый CI job доказывает non-empty collection и failure sensitivity через намеренно падающий sentinel, как требует `.planning/backlog/TASK-201-2026-07-30-tf-ci-cleanup.md`.
- Удаляйте старый путь только после parity всех inventory rows, green replacement job на реальном PR, проверки frozen coverage method и нулевого unexpected skip count.

---

*Convention analysis: 2026-08-07*
