# A3 W11 агент 5 — fe lib+docs part 5 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W11-5-fe-lib-docs.md](A2W11-5-fe-lib-docs.md)  
> Stage A: [A11-5-fe-lib-docs.md](A11-5-fe-lib-docs.md)  
> Периметр: 17 файлов (W11-ASSIGNMENTS A5) · Прочитано: 17 · Пропущено: 0  
> Фокус: docs drift vs код/метрики · Tags `NEW`|`REGRESS`|`RESIDUAL`|`FIXED-OK` · ID `A3W11-5-001`…

## Сверка с A2 (A2W11-5-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W11-5-001 | RESIDUAL P1 | **FIXED-OK** | TASK-105: `vision_and_scope.md:28` G-07 → **Partial**; KPI Wiki «планируется» (`:428-433`); `wiki/API.md:953-955`, `HOME.md:26` — product wiki planned; `Test-Path backend/internal/wiki` → False |
| A2W11-5-002 | RESIDUAL P2 | **FIXED-OK** (канон) | TASK-104: `tz_constraints.md:44` cursor list-issues; rule `60-tz-constraints.mdc`; NFR-14 `system_specification.md:60`; `wiki/API.md:1145-1158`; FE `tracker.ts` `cursor?` — **но** пример-таблица `:144` ещё offset → A3W11-5-003 |
| A2W11-5-003 | FIXED (TASK-123) | **FIXED-OK** | `security_guidelines.md:99` NF-11 → `cmd/server/wiring_fiber.go` |
| A2W11-5-004 | FIXED (TASK-123) | **FIXED-OK** | `system_specification.md:27,160` ~252 OpenAPI / ~286 inventory; swagger paths = **252** |

**REGRESS:** 0 · **RESIDUAL (A2 actives):** 0 закрыты · **NEW:** 3 · **RESIDUAL (false-close Stage A/A2 counts):** 1 (Vitest line в strategy)

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `docs/security_guidelines.md` | 358 | 0 | FIXED-OK NF-11 wiring; ARCH-H1 FIXED |
| `docs/stack_and_tools.md` | 236 | 0 | FIXED-OK 12 stores / motion+pwa pins (=факт) |
| `docs/system_specification.md` | 696 | 1 | NFR-14 cursor OK; пример `:144` offset → A3W11-5-003 |
| `docs/TECHNICAL.md` | 1899 | 0 | FIXED-OK Magic Link; CI `./testframework/...` |
| `docs/testframework_strategy.md` | 354 | 2 | BE/hooks stale + Vitest 37/99 → 001/002 |
| `docs/TESTING.md` | 511 | 1 | inventory 652/37/94 → A3W11-5-001 |
| `docs/test_plan.md` | 216 | 1 | inventory 652/37/94 → A3W11-5-001 |
| `docs/TROUBLESHOOTING.md` | 246 | 1 | 37 hooks / 652 tests → A3W11-5-001 |
| `docs/tz_constraints.md` | 153 | 0 | FIXED-OK cursor (TASK-104) |
| `docs/user_needs_analysis_2026.md` | 263 | 1 | Vitest 94 → A3W11-5-001; Calls camera OK |
| `docs/vision_and_scope.md` | 638 | 1 | G-07 Partial OK; Vitest 94 → A3W11-5-001 |
| `docs/wiki/API.md` | 1190 | 0 | FIXED-OK 252 + cursor + wiki planned |
| `docs/wiki/ARCHITECTURE.md` | 192 | 0 | FIXED-OK 77 routes (=факт) |
| `docs/wiki/DEPLOY.md` | 262 | 0 | чисто (pins PG17 / seaweed 3.97 / LK v1.8) |
| `docs/wiki/HOME.md` | 89 | 0 | FIXED-OK product wiki ≠ docs/wiki |
| `docs/wiki/QUICKSTART.md` | 120 | 0 | FIXED-OK 4 шага; только `./testframework/...` |
| `docs/wiki/SHORTCUTS.md` | 93 | 0 | чисто |

Файлов в списке: **17**. Прочитано: **17**. Пропущено: **0**.

### Замер HEAD (2026-07-29, `ee353287`)

| Метрика | Docs (2026-07-28 snapshot) | Факт HEAD |
|---------|---------------------------|-----------|
| `backend/testframework/**/*_test.go` | **652** | **759** |
| `export_hooks.go` (`hubtest`) | **37** | **39** |
| Vitest `*.test.ts(x)` under `frontend-test/` | **94** (strategy ещё **37/99**) | **125** (54 `.ts` + 71 `.tsx`) |
| `frontend/src/stores/*.ts` | **12** | **12** |
| `frontend/src/routes/**/*.tsx` | **77** | **77** |
| `swagger.json` paths | **252** | **252** |

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A3W11-5-001 | NEW | P2 | `TESTING.md:33-36`; `test_plan.md:52-57`; `TROUBLESHOOTING.md:21,48`; `vision_and_scope.md:42`; `user_needs_analysis_2026.md:61`; `testframework_strategy.md:76-77` | Inventory «measured 2026-07-28»: BE **652** / hooks **37** / Vitest **94** | После remediation TASK-081…123 факт HEAD: **759** / **39** / **125**. Docs sync rule нарушен; KPI/troubleshooting вводят в заблуждение | Переснять одним каноном (дата + команда подсчёта) во всех шести файлах; footer `test_plan`/`TESTING` |
| A3W11-5-002 | RESIDUAL | P2 | `testframework_strategy.md:78` | Vitest всё ещё «**37 файлов, 99 тестов**» | A2 пометил strategy FIXED-OK counts, но Vitest-строка **не** выровнена даже под соседние docs (**94**). False-close W11-A5-007 | Заменить на **125** files (или «N files / M tests» после `npx vitest run`); синхрон с 001 |
| A3W11-5-003 | NEW | P2 | `system_specification.md:144` | Пример: `GET /projects/:id/issues?offset=0&limit=50` | Тот же файл NFR-14 (`:60`) + `wiki/API.md` + FE `ListIssuesParams.cursor` — **cursor**. Пример противоречит ADR-PAGINATION-CURSOR-ISSUES / TASK-104 (недочистка) | `?cursor=…&limit=50` (+ note `sprint_id=none`) |

## FIXED-OK (сверка A2 + Stage A фокус)

| Зона | Проверено в HEAD |
|------|------------------|
| G-07 / US-WIKI | Partial + planned; нет `backend/internal/wiki` / FE module |
| list-issues TZ/NFR/wiki | cursor канон (кроме примера `:144`) |
| NF-11 | `wiring_fiber.go` |
| OpenAPI count | swagger **252**; inventory methodology ~286 |
| Auth docs | Magic Link only; Telegram notify; no `./internal/...` test path |
| Stores / routes / pins | 12 stores; 77 routes; motion `^12.40.0`; pwa `^1.3.0` |
| QUICKSTART | 4 шага онбординг; regression `./testframework/...` |
| DEPLOY / SHORTCUTS / HOME | compose pins; Ctrl+K; product-wiki disclaimer |
| Security packages KPI | auth/ws/rbac/audit/calls/**platform/middleware** |
| Product score | **89/100** в user_needs |
| Calls camera | `CallControls` wording, не «webcams=Production» gap |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 3 |
| P3 | 0 |
| **Total** | **3** |

| Метрика | Значение |
|---------|----------|
| Файлов | **17 / 17 / 0** |
| Активных | **3** (NEW 2 · RESIDUAL 1 · REGRESS 0) |
| FIXED-OK (A2 actives 001/002 + TASK-123) | **4/4** |
| FIXED-OK (Stage A via A2) | **20+** без регресса команд/auth/G-07 |

### Топ-3

1. **A3W11-5-001 (P2, NEW)** — stale test inventory 652/37/94 vs HEAD 759/39/125.
2. **A3W11-5-003 (P2, NEW)** — `system_specification` пример list-issues ещё `offset`.
3. **A3W11-5-002 (P2, RESIDUAL)** — strategy Vitest «37/99» false-close A2.

---

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=3 · P3=0 · Total=3.  
Топ: A3W11-5-001.
