# Codebase Concerns

**Analysis Date:** 2026-08-07

## Tech Debt

**Неоднозначный baseline и coverage gate:**
- Issue: В репозитории одновременно зафиксированы 80.2%, 81.9% и 83.0%; показатель 80.2% получен из запуска с 11 упавшими пакетами, а GitHub coverage job не воспроизводит канонический `-race -covermode=atomic` профиль.
- Files: `.planning/COVERAGE.md`, `.planning/STATE.md`, `docs/test_plan.md`, `.github/workflows/ci.yml`, `.cursor/playbooks/testing.md`
- Impact: Нельзя доказать отсутствие регрессии при миграции тестов или объективно применить порог ≥82%; невалидный профиль может быть принят как релизный baseline.
- Fix approach: До перемещения файлов получить один полностью зелёный Docker baseline, сохранить команды, commit SHA, pass/fail/skip и per-package coverage; затем унифицировать локальную, GitHub и документированную команды.

**Смешанная таксономия backend-тестов:**
- Issue: Все 828 Go-файлов находятся под общим `backend/testframework/`, хотя среди них есть unit-тесты, HTTP/Fiber black-box, SQLite/Ent integration, Redis/NATS и тесты, требующие PostgreSQL/MongoDB.
- Files: `backend/testframework/`, `backend/testframework/testutil/env.go`, `backend/testframework/testutil/factories.go`, `docs/testframework_strategy.md`
- Impact: Механическое перемещение всей директории в `backend/tests/unit/` закрепит неверную классификацию и потребует повторной миграции black-box сценариев в Python.
- Fix approach: До переноса построить machine-readable inventory каждого файла по категории `unit`, `integration`, `api`, `contract`, `e2e`; переносить в `backend/tests/unit/<domain>/` только настоящие unit-тесты, а black-box/API сценарии сразу проектировать в Python-слоях корневого `testframework/`.

**Расхождение целевых путей и объёмов в планах:**
- Issue: Принятые целевые пути — `backend/tests/**` и `frontend/tests/**`, но активные августовские планы и задачи всё ещё указывают `backend/internal/<domain>/unit`, `frontend/src/<module>/unit`, устаревший `frontend/e2e` и разные количества тестов (515, 812, 824 вместо фактических 828; 179 вместо 180).
- Files: `.planning/active/august-2026/00-AUGUST-PLAN.md`, `.planning/active/august-2026/05-AUGUST-DETAILED.md`, `.planning/backlog/TASK-193-2026-07-30-tf-python-scaffold.md`, `.planning/backlog/TASK-195-2026-07-30-tf-e2e-mechanical.md`, `.planning/backlog/TASK-196-2026-07-30-tf-e2e-rewrite.md`, `.planning/backlog/TASK-197-2026-07-30-tf-backend-units-colocated.md`, `.planning/backlog/TASK-198-2026-07-30-tf-frontend-units-colocated.md`, `AGENTS.md`
- Impact: Исполнители могут создать третью схему размещения, потерять тесты при удалении старого дерева или считать parity по неверному знаменателю.
- Fix approach: Перед реализацией синхронизировать план и задачи с одним контрактом: native unit в `backend/tests/**` и `frontend/tests/**`, Python-only framework в `testframework/**`; закрепить фактический inventory 828/180/20 и mapping старый путь → новый путь.

**Дублирующиеся и расходящиеся точки запуска:**
- Issue: GitHub Actions, GitLab CI, два Makefile, pipeline YAML и документация задают разные команды; GitLab и Makefile не используют обязательный тег `hubtest`, GitLab разрешает frontend lint падать через `|| true` и ссылается на несуществующие E2E пути.
- Files: `.github/workflows/ci.yml`, `.gitlab-ci.yml`, `Makefile`, `backend/Makefile`, `testframework/pipeline/jobs/backend-test.yaml`, `testframework/pipeline/jobs/frontend-test.yaml`, `testframework/pipeline/jobs/e2e-smoke.yaml`, `testframework/pipeline/jobs/e2e-full.yaml`, `docs/TROUBLESHOOTING.md`
- Impact: Локально или в зеркальном CI можно получить ложный green, пропустить 828 backend-тестов либо вообще не запустить E2E контракт.
- Fix approach: Сделать один канонический скрипт/entry point для каждого job и вызывать его из GitHub, GitLab и Makefile; запретить fail-open lint и добавить smoke sentinel, который доказывает реальное выполнение ожидаемого набора.

**Frontend-тесты вне typecheck и lint:**
- Issue: 180 Vitest-файлов находятся в `testframework/stages/unit/frontend-test/domains/`, но `frontend/tsconfig.json` включает только `src`, а `frontend/biome.json` проверяет только `src/**`; запуск зависит от ручного junction `node_modules`.
- Files: `testframework/stages/unit/frontend-test/domains/`, `frontend/tsconfig.json`, `frontend/biome.json`, `frontend/vite.config.ts`, `frontend/scripts/link-testframework-node-modules.mjs`, `frontend/package.json`
- Impact: Ошибки типов и стиля в тестах не блокируют CI, а перенос может сломать module resolution на другой ОС или runner.
- Fix approach: При переносе включить `frontend/tests/**` в TypeScript, Biome и Vitest; вынести support utilities в `frontend/tests/support/`; удалить junction только после зелёного native запуска без него.

**Traceability существует преимущественно в документации:**
- Issue: `docs/TESTING.md` содержит 143 TC-строки, но только 8 тестовых файлов имеют code-side `TC-*`; часть pathless имён неоднозначна из-за 33 повторяющихся Go test names, а пять точных ссылок на `Test...` не найдены.
- Files: `docs/TESTING.md`, `backend/testframework/`, `testframework/stages/unit/frontend-test/domains/`, `testframework/stages/e2e/domains/`, `.cursor/playbooks/testing.md`
- Impact: Нельзя автоматически доказать связь TC ↔ автотест, выявить orphan/missing тесты или публиковать надёжную Allure/TMS матрицу.
- Fix approach: Ввести стабильные code-side идентификаторы (`AS_ID`/TC link), полный путь + test name для Go, обязательный reverse-index и CI gate `unmapped = 0` для выбранного acceptance scope.

**Pipeline manifests не являются исполняемым источником истины:**
- Issue: YAML-файлы в `testframework/pipeline/jobs/` дублируют workflow и содержат списки/пути, которые не проверяются автоматически на соответствие реальным тестам.
- Files: `testframework/pipeline/jobs/`, `testframework/PIPELINE.md`, `.github/workflows/ci.yml`
- Impact: Документированный pipeline незаметно расходится с реально исполняемым, а оценки покрытия и состава jobs становятся недостоверными.
- Fix approach: Генерировать manifest из исполняемых конфигураций либо добавить schema/parity check, который сравнивает пути, команды и ожидаемые счётчики.

## Known Bugs

**`make test` не запускает канонический backend suite:**
- Symptoms: Цель запускает `go test ./...` без `-tags hubtest`; экспортные hooks под build tag недоступны, и внешний `backend/testframework/` набор не компилируется или не исполняется как ожидается.
- Files: `Makefile`, `backend/Makefile`, `backend/internal/*/export_hooks.go`, `backend/testframework/`, `docs/TROUBLESHOOTING.md`
- Trigger: Выполнить `make test` или `make -C backend test` вместо документированной команды с `-tags hubtest`.
- Workaround: Использовать каноническую Docker-команду из `AGENTS.md` до исправления общей точки запуска.

**GitLab pipeline не эквивалентен GitHub pipeline:**
- Symptoms: Backend запускается без `hubtest`, frontend не выполняет linker, lint не блокирует job, E2E job ссылается на устаревшие файлы.
- Files: `.gitlab-ci.yml`, `.github/workflows/ci.yml`, `frontend/scripts/link-testframework-node-modules.mjs`, `testframework/stages/e2e/domains/`
- Trigger: Запуск pipeline в GitLab.
- Workaround: Считать GitHub workflow и ручную каноническую команду временным источником результата; не использовать GitLab green как доказательство parity.

**Coverage status может публиковаться после неуспешного прогона:**
- Symptoms: В `.planning/COVERAGE.md` сохранено 80.2% одновременно с 11 failed packages; другие документы продолжают показывать 81.9% или исторические 83.0%.
- Files: `.planning/COVERAGE.md`, `.planning/STATE.md`, `docs/test_plan.md`
- Trigger: Обновление агрегата из частичного/красного запуска или сравнение несогласованных профилей.
- Workaround: Не принимать ни один из трёх показателей как миграционный baseline без нового полного green run.

**Условные skip могут дать ложный E2E green:**
- Symptoms: Все 20 E2E specs содержат direct/conditional skip logic; отсутствие fixture, dev-login или внешней зависимости может уменьшить реально проверенный набор без явного падения job.
- Files: `testframework/stages/e2e/domains/`, `frontend/playwright.config.ts`, `testframework/pipeline/jobs/e2e-smoke.yaml`, `testframework/pipeline/jobs/e2e-full.yaml`
- Trigger: Запуск в окружении без требуемых пользователей, feature flags, API или seed data.
- Workaround: Сверять ожидаемые pass/skip counts и artifacts; для обязательного smoke scope использовать fail-closed preflight.

## Security Considerations

**Fail-open зависимости интеграционных тестов:**
- Risk: Недоступная БД, Redis, NATS или fixture может приводить к `t.Skip`/conditional skip и скрывать регрессию auth, RBAC или audit flows.
- Files: `backend/testframework/`, `testframework/stages/e2e/domains/auth/`, `testframework/stages/e2e/domains/platform/`, `.cursor/REGULATION.md`
- Current mitigation: GitHub smoke/full jobs поднимают сервисы и используют отдельные тестовые credentials; обязательный `hubtest` раскрывает test-only hooks.
- Recommendations: Разделить optional и release-blocking dependencies; для security-critical suites сделать preflight fail-closed и установить максимальный допустимый skip count равным нулю.

**Секреты и персональные данные в Python/Allure artifacts:**
- Risk: API-тела, cookies, токены, session storage и ответы CRM/HR могут попасть в pytest/Allure/Locust attachments.
- Files: `frontend/playwright.config.ts`, `testframework/stages/e2e/domains/`, `.cursor/playbooks/testing.md`, `docs/security_guidelines.md`
- Current mitigation: Playwright trace сохраняется только при retry; тестовая среда использует выделенные credentials.
- Recommendations: Добавить централизованную redaction-функцию, запрещённые header/body fields, retention policy и artifact scanning до публикации Allure.

**Нагрузочные тесты могут быть направлены на production:**
- Risk: Будущий Locust runner без allowlist способен создать нагрузку или тестовые данные в production.
- Files: `.planning/backlog/TASK-199-2026-07-30-tf-integration-python.md`, `.planning/active/august-2026/05-AUGUST-DETAILED.md`, `testframework/`
- Current mitigation: Python/load scaffold отсутствует, поэтому риск ещё не материализован.
- Recommendations: По умолчанию разрешать только localhost/CI hostnames, требовать явный non-production marker и немедленно прекращать run при production-like URL.

## Performance Bottlenecks

**Серийный E2E и повторные retry:**
- Problem: Playwright ограничен одним worker и двумя retry в CI; 20 specs и 53 теста выполняются последовательно.
- Files: `frontend/playwright.config.ts`, `testframework/stages/e2e/domains/`
- Cause: Shared fixtures/state и недостаточно изолированные пользователи делают параллелизм рискованным.
- Improvement path: Сначала изолировать accounts/data namespaces, затем распараллелить по независимым доменам и сохранить serial lane только для stateful сценариев.

**Фиксированные ожидания вместо событийной синхронизации:**
- Problem: В 21 Go test-файле используется `time.Sleep`, а в E2E есть `waitForTimeout`.
- Files: `backend/testframework/`, `testframework/stages/e2e/domains/`
- Cause: Асинхронные события, websocket/NATS и UI transitions проверяются по времени, а не по наблюдаемому условию.
- Improvement path: Использовать bounded polling/event acknowledgements в Go и locator/response/state expectations в Playwright/Python.

**Dual-run временно удваивает стоимость suite:**
- Problem: Требуемая parity-фаза запускает старое и новое расположение, а Python black-box добавляет отдельный runtime и инфраструктуру.
- Files: `.github/workflows/ci.yml`, `backend/testframework/`, `testframework/stages/unit/frontend-test/`, `testframework/`
- Cause: Безопасная миграция требует доказать одинаковый test inventory и outcome до удаления старых entry points.
- Improvement path: Разделить быстрый PR subset и nightly/full parity, кэшировать Go/Node/Python зависимости и завершать dual-run только после зафиксированных acceptance gates.

## Fragile Areas

**Test-only экспортные hooks:**
- Files: `backend/internal/*/export_hooks.go`, `backend/testframework/`
- Why fragile: 39 hook-файлов доступны только с `hubtest`, а 338 тестовых файлов используют `Export*`; принятый путь `backend/tests/**` остаётся отдельным Go package path и сам по себе не устраняет зависимость от hooks.
- Safe modification: Сохранить build tag и каноническую команду при первом переносе; удалять конкретный hook только после переписывания всех его consumers на публичный контракт или действительно co-located внутренний тест.
- Test coverage: Отдельного gate, проверяющего orphan hooks и missing `hubtest`, нет.

**Внешний frontend test tree и junction `node_modules`:**
- Files: `frontend/scripts/link-testframework-node-modules.mjs`, `frontend/vite.config.ts`, `testframework/stages/unit/frontend-test/domains/`
- Why fragile: Module resolution зависит от platform-specific junction и ручного pre-step; TypeScript/Biome этот каталог не валидируют.
- Safe modification: Сначала добавить `frontend/tests/**` в конфиги и доказать green native run, затем удалить external include/linker одним cutover.
- Test coverage: Нет enforced frontend coverage и coverage provider.

**Удаление корневого legacy tree до извлечения сценариев:**
- Files: `.planning/backlog/TASK-193-2026-07-30-tf-python-scaffold.md`, `.planning/backlog/TASK-195-2026-07-30-tf-e2e-mechanical.md`, `.planning/backlog/TASK-196-2026-07-30-tf-e2e-rewrite.md`, `.planning/backlog/TASK-197-2026-07-30-tf-backend-units-colocated.md`, `.planning/backlog/TASK-198-2026-07-30-tf-frontend-units-colocated.md`, `testframework/`
- Why fragile: TASK-193 планирует очистку дерева, данные из которого нужны последующим задачам; handoff описан текстом, но полный file-level manifest не закреплён.
- Safe modification: Сначала сохранить inventory с хэшами, классификацией и целевым путём каждого файла; физическое удаление разрешать только после dual-run parity и восстановления из manifest.
- Test coverage: Нет automated sentinel на потерянные/дублированные test cases.

**Параллельные Python integration tests с общей БД:**
- Files: `testframework/`, `.planning/backlog/TASK-194-2026-07-30-tf-e2e-shared-fixtures.md`, `.planning/active/august-2026/05-AUGUST-DETAILED.md`
- Why fragile: Python scaffold ещё отсутствует; не определены isolation keys, transaction boundaries, cleanup и xdist-safe fixtures.
- Safe modification: Ввести уникальный namespace/run ID, идемпотентные factories, deterministic cleanup и отдельные serial markers для конфликтующих сценариев.
- Test coverage: Нет реализации или pilot suite, подтверждающих безопасность параллелизма.

## Scaling Limits

**Единый backend regression lane:**
- Current capacity: 828 test files и 3,663 `Test*` functions запускаются одним `go test -tags hubtest -race ... ./testframework/...` контуром.
- Limit: Рост suite увеличивает PR latency; повторный полный запуск после флейка дорог, а разбивка без package ownership может скрыть cross-domain regressions.
- Scaling path: Сохранить обязательный full/nightly gate, добавить детерминированный affected-package PR matrix и публиковать объединённый coverage profile.

**E2E shared-state capacity:**
- Current capacity: 20 specs, 53 tests, один worker.
- Limit: Новые домены линейно увеличивают время; shared users/data препятствуют параллелизму.
- Scaling path: Перейти на isolated test tenants/users, shard-safe setup и domain shards с отдельным smoke contract из трёх specs.

**Python framework отсутствует как масштабируемый слой:**
- Current capacity: В корневом `testframework/` нет `pyproject.toml`, pytest configuration, Python API/integration/load suites или lockfile.
- Limit: Невозможно оценить duration, flaky rate, parallel safety и стоимость будущего black-box слоя.
- Scaling path: Начать с одного вертикального pilot domain, измерить timings и лишь затем расширять категории и workers.

## Dependencies at Risk

**Неопределённый Python toolchain:**
- Risk: Нет зафиксированных Python version, pytest/plugins, HTTP client, schema validation, Allure adapter, Locust и lock strategy.
- Impact: Невоспроизводимые окружения, конфликтующие плагины и разные результаты локально/CI.
- Migration plan: Добавить `testframework/pyproject.toml` и lockfile, зафиксировать runtime, минимальный набор plugins и одну команду установки/запуска; затем валидировать на чистом runner.

**Frontend coverage provider отсутствует:**
- Risk: В `frontend/package.json` нет Vitest coverage provider и нет threshold configuration в `frontend/vite.config.ts`.
- Impact: Перенос 180 тестов может сохранить pass count, но незаметно потерять покрываемые модули.
- Migration plan: Выбрать поддерживаемый Vitest provider, получить baseline до переноса, зафиксировать thresholds и сравнивать file/module coverage в dual-run.

**Allure Playwright/pytest integration не установлена:**
- Risk: Документация требует Allure/TMS traceability, но `frontend/package.json` и `frontend/playwright.config.ts` не содержат Allure reporter; Python package ещё отсутствует.
- Impact: Acceptance по отчётности и TC links невозможно выполнить только переносом файлов.
- Migration plan: Зафиксировать reporter/adapters, redaction и artifact contract; добавить тест на ненулевое число результатов и gate `unmapped = 0`.

## Missing Critical Features

**Python-only корневой framework:**
- Problem: Целевой `testframework/` ещё не имеет Python package, `api/`, `integration/`, `load/`, `full/`, fixtures, markers или CI jobs.
- Blocks: Перенос black-box сценариев из Go/TypeScript и доказательство новой архитектуры.

**Стратегия сохранения Go coverage после переноса black-box в Python:**
- Problem: `go test -coverpkg=./internal/...` учитывает покрытие только от Go test binary; Python API/integration tests не попадут в Go coverage profile.
- Blocks: Нельзя одновременно обещать перенос black-box сценариев и неизменный ≥82% Go gate без сохранения достаточного Go unit слоя либо явного нового измерения.

**Machine-readable migration inventory:**
- Problem: Нет единого списка 828 Go, 180 frontend unit и 20 E2E файлов с категорией, владельцем, старым/новым путём, test count, TC IDs и статусом parity.
- Blocks: Безопасное удаление legacy paths, обнаружение потерь/дубликатов и воспроизводимый audit.

**Exact execution sentinels:**
- Problem: Jobs не валидируют ожидаемое число collected/executed/skipped tests; для E2E conditional skips широко распространены.
- Blocks: Fail-closed доказательство, что green означает запуск полного обязательного scope.

## Test Coverage Gaps

**Чистый общий backend baseline:**
- What's not tested: Один воспроизводимый green run всего backend suite с race detector, точным package list и согласованным coverage profile.
- Files: `.github/workflows/ci.yml`, `.planning/COVERAGE.md`, `backend/testframework/`
- Risk: Регрессия покрытия или потеря тестов во время миграции останется недоказанной.
- Priority: High

**Frontend line/branch/function coverage:**
- What's not tested: Coverage не собирается и thresholds не применяются к 180 Vitest-файлам.
- Files: `frontend/package.json`, `frontend/vite.config.ts`, `testframework/stages/unit/frontend-test/domains/`
- Risk: Удалённый или переставший исполняться тест может не изменить pass count.
- Priority: High

**Code-to-TC reverse traceability:**
- What's not tested: Большинство тестов не имеет стабильного code-side TC ID; orphan и missing mappings не блокируют CI.
- Files: `docs/TESTING.md`, `backend/testframework/`, `testframework/stages/unit/frontend-test/domains/`, `testframework/stages/e2e/domains/`
- Risk: Acceptance claims и TMS/Allure отчёты не подтверждают реальное покрытие требований.
- Priority: High

**Реальная PostgreSQL семантика в PR gate:**
- What's not tested: Значительная часть Ent-тестов использует SQLite; лишь ограниченное число тестов явно зависит от `POSTGRES_TEST_URL`.
- Files: `backend/testframework/`, `.github/workflows/ci.yml`, `deployments/docker-compose.test.yml`
- Risk: PostgreSQL-specific SQL, locking, transaction и type behavior может сломаться после зелёных SQLite тестов.
- Priority: High

**Unexpected skip budget:**
- What's not tested: Нет gate на нулевое число неожиданных Go/E2E skips и нет разделения expected/optional skips.
- Files: `backend/testframework/`, `testframework/stages/e2e/domains/`, `.github/workflows/ci.yml`
- Risk: Обязательный security или smoke сценарий может не выполниться при green job.
- Priority: High

## Migration Risk Register

| Risk | Probability | Impact | Required control | Evidence |
|---|---|---|---|---|
| Потеря Go coverage при переносе black-box тестов в Python | High | Critical | Сохранить достаточный Go unit слой либо утвердить отдельную black-box метрику и новый baseline; не снижать порог молча | `.github/workflows/ci.yml`, `backend/testframework/`, `testframework/` |
| Потеря/дублирование файлов из-за удаления legacy tree | Medium | Critical | Inventory с хэшами и old→new mapping, dual-run, deletion только после parity | `.planning/backlog/TASK-193-2026-07-30-tf-python-scaffold.md`, `.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md` |
| Неверная классификация всех Go tests как unit | High | High | File-level taxonomy до переноса | `backend/testframework/`, `docs/testframework_strategy.md` |
| Сохранение зависимости `backend/tests/**` от test-only hooks | High | High | Канонический `hubtest`; удаление hooks только consumer-by-consumer | `backend/internal/*/export_hooks.go`, `backend/testframework/` |
| Ложный green из-за разных CI entry points | High | High | Общие executable scripts и parity checks GitHub/GitLab/Makefile | `.github/workflows/ci.yml`, `.gitlab-ci.yml`, `Makefile` |
| Frontend tests проходят без typecheck/lint/coverage | High | High | Включить `frontend/tests/**` во все три gate до cutover | `frontend/tsconfig.json`, `frontend/biome.json`, `frontend/vite.config.ts` |
| E2E green с пропущенными обязательными сценариями | Medium | High | Fail-closed preflight и exact pass/skip sentinel | `testframework/stages/e2e/domains/`, `frontend/playwright.config.ts` |
| Невоспроизводимый Python framework | Medium | High | Зафиксированный runtime, lockfile и clean-run validation | `testframework/` |

## Migration Acceptance Gates

1. Зафиксирован один зелёный pre-migration baseline с inventory 828 Go files / 3,663 Go tests, 180 Vitest files, 20 E2E specs / 53 tests и отдельным списком skip.
2. Для каждого legacy-файла существует machine-readable old→new mapping, категория и checksum; число unmapped и accidentally duplicated файлов равно нулю.
3. Native unit suites исполняются из `backend/tests/**` и `frontend/tests/**`; корневой `testframework/**` содержит только Python framework и его данные/конфигурацию.
4. Старые и новые entry points работают в dual-run до совпадения collected tests, результатов и согласованных coverage deltas; deletion выполняется только после parity.
5. Go coverage ≥82% подтверждён чистым green run. Если Python заменяет Go black-box тесты, отдельно одобрена и документирована метрика их покрытия; Python execution не считается частью Go profile.
6. `frontend/tests/**` входит в TypeScript, Biome, Vitest и coverage; junction script больше не нужен и удаляется только после clean-run validation.
7. GitHub, GitLab, Makefile, pipeline manifests, `AGENTS.md`, `docs/test_plan.md`, `docs/TESTING.md` и `.planning/STATE.md` используют одни пути, команды и актуальные счётчики.
8. E2E сохраняет 20 specs / 53 tests и трёх-spec smoke contract; неожиданные skips равны нулю, а обязательные dependencies fail closed.
9. TC/AS_ID ↔ code mapping имеет полный путь, уникальный ID и reverse index; CI показывает `unmapped = 0` в обязательном scope.
10. Python framework проходит clean install, unit/API/integration pilot, artifact redaction, deterministic cleanup и xdist/serial policy; load runner блокирует production targets.

---

*Concerns audit: 2026-08-07*
