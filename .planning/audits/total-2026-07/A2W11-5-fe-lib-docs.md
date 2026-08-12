# A2 W11 агент 5 — fe lib+docs part 5 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W11-ASSIGNMENTS.md` A5 · Stage A: `A11-5-fe-lib-docs.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `docs/security_guidelines.md` | 353 | 0 | FIXED TASK-123 — NF-11 → wiring_fiber.go |
| `docs/stack_and_tools.md` | 182 | 0 | FIXED-OK stores/pins |
| `docs/system_specification.md` | 692 | 0 | FIXED TASK-123 — summary ~252 OpenAPI |
| `docs/TECHNICAL.md` | 1899 | 0 | FIXED-OK Magic Link + TF path |
| `docs/testframework_strategy.md` | 258 | 0 | FIXED-OK counts |
| `docs/TESTING.md` | 511 | 0 | FIXED-OK только `./testframework/...` |
| `docs/test_plan.md` | 163 | 0 | FIXED-OK inventory 652 |
| `docs/TROUBLESHOOTING.md` | 162 | 0 | FIXED-OK 37 hooks / 652 tests |
| `docs/tz_constraints.md` | 153 | 1 | RESIDUAL issues offset vs cursor |
| `docs/user_needs_analysis_2026.md` | 263 | 0 | FIXED-OK 89/100; calls camera wording |
| `docs/vision_and_scope.md` | 636 | 1 | RESIDUAL G-07 vs US-WIKI planned |
| `docs/wiki/API.md` | 1172 | 0 | FIXED-OK 252 + ApiError note |
| `docs/wiki/ARCHITECTURE.md` | 192 | 0 | FIXED-OK / routes metric |
| `docs/wiki/DEPLOY.md` | 258 | 0 | чисто |
| `docs/wiki/HOME.md` | 87 | 0 | чисто |
| `docs/wiki/QUICKSTART.md` | 120 | 0 | FIXED-OK 4 шага; без `./internal/...` |
| `docs/wiki/SHORTCUTS.md` | 93 | 0 | чисто |

Файлов в списке: **17**. Прочитано: **17**. Пропущено: **0**.

## Находки (открытые)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A2W11-5-001 | RESIDUAL | P1 | `vision_and_scope.md:28` vs `:428-431` | G-07: «US-INTR/US-CAL/**US-WIKI реализованы**» ↔ KPI wiki — **планируется** | W11-A5-006 не закрыт PO; wiki-модуля как shipped G-07 нет | G-07 → Partial (INTR+CAL shipped, WIKI planned) |
| A2W11-5-002 | RESIDUAL | P2 | `tz_constraints.md:44` | «Задачи — offset/limit» при cursor list issues в FE/BE | W11-A2-009 / P2-FE-U04; правило 60 зеркалит drift | Обновить TZ+rule 60 под cursor **или** вернуть offset в API |

## FIXED (TASK-123 · 2026-07-28)

| ID | Evidence |
|----|----------|
| A2W11-5-003 | `docs/security_guidelines.md` NF-11 → `cmd/server/wiring_fiber.go` |
| A2W11-5-004 | `docs/system_specification.md:27` summary → REST **~252** OpenAPI / ~286 inventory (как `:160`) |

## FIXED-OK (Stage A → закрыто после B/C)

| Stage A ID | Evidence |
|------------|----------|
| W11-A5-001…004 | TESTING / TECHNICAL / QUICKSTART — только `./testframework/...` |
| W11-A5-002 | inventory переписан под testframework (не Telegram auth) |
| W11-A5-005 | user_needs Calls: camera через CallControls, не «webcams=Production» gap |
| W11-A5-007…009 | test counts 652 / Vitest 94 / export_hooks 37 |
| W11-A5-010 | большинство wiring line-refs (кроме NF-11) |
| W11-A5-011 | ARCH-H1 → FIXED |
| W11-A5-012 | health/metrics → wiring_* |
| W11-A5-013 | wiki/API + `:160` методика 252/~286 |
| W11-A5-014 | vision KPI **77** routes |
| W11-A5-015 | security packages wording (см. AGENTS; scorecard sync) |
| W11-A5-016 | product score **89/100** |
| W11-A5-017 | TECHNICAL auth: Magic Link; Telegram notify |
| W11-A5-018 | tz Gantt UI shipped |
| W11-A5-019 | QUICKSTART онбординг **4** шага |
| W11-A5-020 | stack: 12 stores, motion/pwa pins |
| W11-A5-021 | API.md уточнение ports (Caddy/backend) |

## Чисто

DEPLOY/SHORTCUTS/HOME — compose pins, hotkeys, magic TTL согласованы с кодом.

## Сводка агента

| Метрика | Значение |
|---------|----------|
| Файлов | 17/17/0 |
| Открытых находок | **2** (001/002 вне TASK-123) |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 2 (P1×1, P2×1); P3×2 закрыты TASK-123 |
| FIXED-OK | 20 + TASK-123 (2) |
