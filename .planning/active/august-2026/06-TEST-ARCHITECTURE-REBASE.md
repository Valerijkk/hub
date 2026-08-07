# Test Architecture Rebase — август 2026

> **Статус:** CANONICAL · **Дата:** 2026-08-07 · **Scope:** фазы 128–141
>
> Этот документ отменяет только раздел «Приоритет 1 — тестовый фреймворк»,
> его недельную раскладку, пути и destructive-cutover предположения в
> [`05-AUGUST-DETAILED.md`](05-AUGUST-DETAILED.md). Остальные приоритеты того
> документа сохраняются с уточнённым post-test порядком из §8 ниже.

## 1. Решение и граница программы

Историческая **pipeline migration COMPLETE** (перенос старых Go/Vitest/E2E в
pipeline-oriented каталоги) не открывается заново. Новая **Python/native test
architecture migration NOT_STARTED** — отдельная программа. Её целевая схема:

**Status distinction:** prior pipeline migration = `COMPLETE`; Python migration
= `NOT_STARTED`. **Live baseline:** 828 Go / 180 Vitest / 20 E2E specs / 53 E2E
cases / 0 Python.

```text
backend/tests/**           Go unit/component tests, owner: backend
frontend/tests/**          Vitest/RTL tests, owner: frontend
testframework/python/**    pytest/httpx/Playwright: API/integration/E2E/full
testframework/load/**      Locust load profiles, owner: QA automation
```

Миграция только через dual-run. `backend/testframework/**`,
`testframework/stages/unit/frontend-test/**` и
`testframework/stages/e2e/**` остаются исполняемыми до принятого parity каждой
inventory-строки. Корневой `testframework/` не очищается при создании Python
scaffold: новые Python-файлы сосуществуют со старым деревом до Phase 141.

## 2. Текущая стадия и evidence

Hub — **широкая внутренняя beta**, не production-ready. Программные milestone
могут быть исторически отмечены COMPLETE, но это не является production
acceptance; ручной prod-blocker PG-03 также остаётся вне этой программы.

| Поверхность | Live inventory 2026-08-07 | Runtime verdict |
|---|---:|---|
| Go `backend/testframework/**/*_test.go` | **828 файлов**, 3 663 `Test*` | current green/coverage **UNKNOWN** |
| Vitest legacy | **180 файлов** (60 TS + 120 TSX) | baseline не завершён |
| Playwright TS | **20 specs / 53 cases** | исторический suite; live parity не доказан |
| Python tests | **0** | Python отсутствует в PATH |
| `export_hooks.go` / все `hubtest` production files | **39 / 45** | 338 Go tests используют `Export*` |
| Docker | compose parse доступен | daemon недоступен; service run NOT_RUN |

Значения Go coverage 83.0%, 81.9% и 80.2% получены разными или загрязнёнными
прогонами и не являются взаимозаменяемым baseline. До первого перемещения Phase
128 обязан получить один воспроизводимый green baseline с Go coverage **≥82%**
либо заранее получить явное founder-решение о новом denominator/threshold в
TASK-204. Иначе Phase 128 остаётся BLOCKED. Timeout, missing runtime,
unavailable Docker и skip обязательной зависимости — не PASS.

## 3. Неподвижные контракты

1. **Phase 128 блокирует все перемещения и scaffold.** До него разрешены только
   read-only inventory и planning.
2. Для каждого legacy test-файла создаётся machine-readable строка:
   `source_path`, checksum, runner, domain, stable IDs/test count, taxonomy,
   target path, coverage-bearing, hook consumers, baseline result, replacement
   result, parity status, disposition и evidence URI.
3. Таксономия: `go-unit`, `go-component`, `python-api`, `python-integration`,
   `python-contract`, `python-e2e`, `python-full`, `python-load`,
   `frontend-unit`. Каталог-источник сам по себе тип теста не определяет.
4. Перемещение считается принятым только на одном immutable build при равных
   assertions/negative paths, явном pass/fail/skip count, forced-failure или
   mutation evidence и решении независимого verifier `PARITY_ACCEPTED`.
5. `backend/tests/**` остаётся внешним package tree внутри Go module. Оно может
   импортировать `backend/internal/**`, но не получает private access. Hooks под
   `hubtest` не удаляются механически; каждый hook требует consumer=0 и зелёного
   доменного доказательства.
   **Go hooks cannot be mechanically deleted.**
6. Python black-box execution не попадает в `go test -coverpkg` и не может
   автоматически компенсировать Go statement coverage. Порог ≥82% сохраняется,
   пока founder до перемещения явно не утвердит новый denominator/threshold.
7. Legacy и new jobs используют раздельные имена, данные и artifacts. Неожиданный
   skip, нулевая collection или неработающий failing sentinel блокируют cutover.
8. Protected external QA zone исключена из discovery, зависимостей и evidence.

## 4. Каноническая фазовая матрица

| Phase | Outcome | TASK |
|---:|---|---|
| [128](../../phases/128-tf-python-scaffold/PLAN.md) | Status/baseline recovery | TASK-202 inventory ledger; TASK-203 canonical baseline; TASK-204 taxonomy + coverage contract |
| [129](../../phases/129-tf-e2e-shared/PLAN.md) | Python scaffold + API vertical slice | TASK-193; TASK-205 |
| [130](../../phases/130-tf-e2e-mechanical/PLAN.md) | Frontend native foundation | TASK-198; TASK-206 |
| [131](../../phases/131-tf-e2e-rewrite/PLAN.md) | Frontend domain migration | TASK-207; TASK-208 |
| [132](../../phases/132-tf-backend-units-colocated/PLAN.md) | Backend native foundation | TASK-197; TASK-209 |
| [133](../../phases/133-tf-frontend-units-colocated/PLAN.md) | Backend security/platform | TASK-210; TASK-211 |
| [134](../../phases/134-tf-integration-python/PLAN.md) | Backend work/realtime | TASK-212; TASK-213 |
| [135](../../phases/135-tf-locust-load/PLAN.md) | Backend business/integrations | TASK-214; TASK-215 |
| [136](../../phases/136-tf-ci-cleanup/PLAN.md) | Python API/integration | TASK-199; TASK-216 |
| [137](../../phases/137-tf-python-e2e-foundation/PLAN.md) | Python E2E foundation + smoke parity | TASK-194; TASK-217 |
| [138](../../phases/138-tf-python-e2e-migration/PLAN.md) | Python E2E full migration | TASK-195; TASK-196 |
| [139](../../phases/139-tf-load/PLAN.md) | Locust local harness/baseline | TASK-200 |
| [140](../../phases/140-tf-ci-dual-run/PLAN.md) | CI dual-run + execution sentinels | TASK-201; TASK-218 |
| [141](../../phases/141-tf-cutover-cleanup/PLAN.md) | Safe cutover, cleanup, final verification | TASK-219; TASK-220 |

Каталоги 128–136 сохраняют исторические slugs; номер и содержимое PLAN являются
источником смысла. Это избегает отдельной разрушительной операции переименования.

## 5. Dependency graph и волны

```text
128 baseline/status
 ├─129 Python scaffold/API ───────────────┬─136 Python API/integration
 │                                       ├─137 E2E foundation ─138 E2E full
 │                                       └─139 load
 ├─130 FE foundation ─131 FE domains
 └─132 BE foundation ─133 BE sec/platform ─134 BE work/realtime ─135 BE rest

{131,135,136,138,139} → 140 CI dual-run → 141 cutover/cleanup/verification
```

Domain waves могут выполняться параллельно только после Phase 128 и только при
непересекающихся target/evidence files. Каждая фаза имеет бюджет 1–3 дня. Если
inventory доказывает больший объём, фаза останавливается и перепланируется; её
нельзя закрывать частичным переносом или скрытым defer.

## 6. Coverage и CI contract

- Phase 128 фиксирует command, toolchain, services, source revision, profile
  hash, overall/per-package coverage и expected skip budget.
- Старый и новый Go suites меряются одной instrumentation. Python coverage
  публикуется отдельно как behavioral/API evidence.
- Frontend до перемещения получает native typecheck, Biome, Vitest discovery и
  coverage baseline для `frontend/tests/**`.
- GitHub Actions — текущий executable authority; GitLab, Makefile и
  `testframework/pipeline/jobs/*.yaml` считаются drift до Phase 140.
- Phase 140 только добавляет dual-run и gates. Required checks переключаются и
  legacy удаляется лишь в Phase 141 после реального CI green.

## 7. Cutover gate Phase 141

Удаление разрешено, только если одновременно:

- inventory `unmapped = 0`, `PARITY_REQUIRED = 0`, accidental duplicates = 0;
- legacy/new behavior, counts, expected skips и artifacts приняты verifiers;
- принято явное решение по текущему Go coverage contract и green профиль
  получен тем же способом;
- real GitHub CI dual-run green (не локальная симуляция), зеркала согласованы;
- failing sentinels доказали чувствительность всех новых jobs;
- rollback описан и проверен: предыдущие job/path можно восстановить одним
  revert до истечения retention window;
- hooks удаляются только поштучно при consumer=0; root framework целиком не
  удаляется, удаляются лишь allowlisted legacy entries.

## 8. Порядок после test rebase

1. Свежий evidence-based re-audit B1–B6.
2. Productization по результатам аудита: **calendar → intranet → accounting →
   CRM → HR → bases**.
3. Hardening modular-monolith boundaries и schema ownership.
4. Production acceptance, включая ops/security/manual gates.
5. Затем **Wiki → OIDC SSO → AI MCP**.

До re-audit запрещены подробные «готовые» feature-планы, основанные на старых
оценках: аудит сначала доказывает фактические gaps, затем создаются фазы.

## 9. Главные риски и владельцы решений

| Риск | Gate / mitigation |
|---|---|
| Ложный green из-за skips/zero collection | exact count + unexpected-skip + failing-sentinel gates |
| Потеря Go coverage при Python replacement | separate metrics; retain/add Go units or founder-approved contract |
| Раннее удаление legacy/root | manifest allowlist; deletion only Phase 141; rollback |
| Сломанные hooks/private access | hook consumer ledger; no mechanical deletion |
| CI drift | one job contract across GitHub/GitLab/Makefile/manifests in Phase 140 |
| Нет Python/Docker локально | NOT_RUN/BLOCKED evidence; real CI proof required before cutover |

---

*Канон фаз 128–141. Исторические completion claims не заменяют live evidence.*
