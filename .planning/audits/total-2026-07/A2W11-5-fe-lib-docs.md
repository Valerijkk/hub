# A2 W11 агент 5 — fe lib+docs part 5 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W11-ASSIGNMENTS.md` A5 · Stage A: `A11-5-fe-lib-docs.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `docs/security_guidelines.md` | 353 | 1 | RESIDUAL NF-11 → main.go |
| `docs/stack_and_tools.md` | 182 | 0 | FIXED-OK stores/pins |
| `docs/system_specification.md` | 692 | 1 | RESIDUAL «286» в summary vs 252 |
| `docs/TECHNICAL.md` | 1899 | 0 | FIXED-OK Magic Link + TF path |
| `docs/testframework_strategy.md` | 258 | 0 | FIXED-OK counts |
| `docs/TESTING.md` | 511 | 0 | FIXED-OK только `./testframework/...` |
| `docs/test_plan.md` | 163 | 0 | FIXED-OK inventory 652 |
| `docs/TROUBLESHOOTING.md` | 162 | 0 | FIXED-OK 37 hooks / 652 tests |
| `docs/tz_constraints.md` | 153 | 1 | RESIDUAL issues offset vs cursor |
| `docs/user_needs_analysis_2026.md` | 263 | 0 | FIXED-OK 89/100; calls camera wording |
| `docs/vision_and_scope.md` | 636 | 0 | **FIXED** G-07 Partial (TASK-105) |
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
| A2W11-5-001 | FIXED | P1 | `vision_and_scope.md` G-07 + Wiki KPI; `docs/wiki/{API,HOME}.md` | G-07: «US-WIKI реализованы» ↔ KPI wiki planned | ~~RESIDUAL~~ → **FIXED** docs+PO (TASK-105, 2026-07-28): G-07 = **Partial** (INTR+CAL base shipped, US-WIKI planned); KPI Wiki остаются «планируется» | G-07 Partial; не утверждать «реализованы» до ship wiki |
| A2W11-5-002 | RESIDUAL | P2 | `tz_constraints.md:44` | «Задачи — offset/limit» при cursor list issues в FE/BE | W11-A2-009 / P2-FE-U04; правило 60 зеркалит drift | Обновить TZ+rule 60 под cursor **или** вернуть offset в API |
| A2W11-5-003 | RESIDUAL | P3 | `security_guidelines.md:99` | NF-11 evidence всё ещё `cmd/server/main.go` | W11-A5-010 почти закрыт wiring-refs; один leftover | Ссылка → `wiring_fiber.go` / фактический EnableTrustedProxyCheck |
| A2W11-5-004 | RESIDUAL | P3 | `system_specification.md:27` | Summary table «REST **286**» при `:160` каноне **~252** OpenAPI | W11-A5-013 частично: деталь поправлена, шапка нет | Выровнять `:27` под методику `:160` |

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
| Открытых находок | **3** (после TASK-105: A2W11-5-001 FIXED) |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 3 (P2×1, P3×2) — было 4; P1 G-07 закрыт |
| FIXED-OK | 21 (+ A2W11-5-001) |
