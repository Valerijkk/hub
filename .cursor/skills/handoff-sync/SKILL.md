---
name: handoff-sync
description: Обновить документы рабочей памяти после сессии — STATE.md, COVERAGE.md, domain docs по матрице docs-sync. Запускай в конце каждой значимой сессии.
disable-model-invocation: true
---

# /handoff-sync — Session Handoff Sync

Обновляет документы рабочей памяти для следующей сессии.

**Обязательное правило:** `.cursor/rules/55-docs-sync.mdc` — после **любого** изменения кода обновляй затронутую документацию в **той же сессии**, не только в handoff.

## Step 0: Docs-sync matrix (при изменениях в сессии)

| Изменение | Файлы |
|-----------|-------|
| API | `docs/system_specification.md`, `docs/wiki/API.md` |
| UI / UX / hotkeys | `docs/wiki/SHORTCUTS.md`, `docs/wiki/QUICKSTART.md`, `docs/competitive_audit.md` |
| Tests / coverage | `.planning/COVERAGE.md`, `docs/test_plan.md`, `docs/TESTING.md` |
| CI / deploy | `docs/stack_and_tools.md`, `deployments/.env.example`, `docs/wiki/DEPLOY.md` |
| Security | `docs/security_guidelines.md`, при закрытии finding — `.planning/audits/2026-06/AUDIT-SCORECARD.md` |
| Waves / gaps | `.planning/tasks/TASKS-CLOSE-GAPS.md`, `.planning/tasks/PLAN-CLOSE-GAPS.md`, `.planning/tasks/TASKS-PHASES-121-FINAL-AUDIT.md` |
| Roadmap-решение / gray-zone | `.planning/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md`, доменный `.planning/gray-zones/AREA-*.md` |
| Архитектурное решение | `.planning/decisions/ADR-*.md`, `docs/architecture.md` |
| July Rebuild / QA-HARDENING статус | `.planning/july-2026/PROJECT-STATUS-2026-07-10.md`, `.planning/july-2026/INDEX-JULY-2026.md` |
| Прод-операции / секреты | `.planning/production/PRODUCTION-HANDOFF.md`, `.planning/production/FA-06-EXECUTION-LOG.md` |
| Спека модуля (Intranet/Calendar/CRM/HR/AI/no-code DB) | `.planning/specs/*.md`, `docs/system_specification.md` |
| Завершение milestone | `.planning/milestones/<name>/SUMMARY.md` |
| Всегда | `.planning/STATE.md`, `AGENTS.md` (если менился статус) |

Полная карта `.planning/` (все тематические папки с примерами файлов): `.cursor/REGULATION.md` §12, навигация: `.planning/README.md`.

## Step 1: Прочитать текущее состояние

```
Прочитай:
- .planning/STATE.md         — текущая позиция в roadmap
- .planning/README.md        — навигация по тематическим папкам
- .planning/ROADMAP.md       — статус фаз (исторический)
- .planning/july-2026/ROADMAP-FINAL-2026-07-03.md — актуальный roadmap волн
- .planning/COVERAGE.md      — метрики coverage
- .planning/code-review-final/SUMMARY.md  — открытые findings
```

## Step 2: Обновить STATE.md

Обновить секции:
```markdown
## Current Position
Phase: <N>
Plan: <что делали>
Status: <что завершено / что осталось>
Last activity: <YYYY-MM-DD>

## Pending Todos
- [ ] <задача 1> — <приоритет>
- [ ] <задача 2>

## Blockers/Concerns
- <что может заблокировать следующий шаг>
```

## Step 3: Обновить ROADMAP.md статусы

Если фаза завершена — отметить `Complete` с датой.

## Step 4: Обновить COVERAGE.md / TASKS (если менялись тесты или gaps)

- PG-xx closed → статус ✅ + дата
- Новые метрики coverage → таблица Wave 1 security packages
- REQ matrix → Code/Tests колонки

## Step 5: Проверить открытые criticals

Сверить список CRITICAL findings с `.planning/code-review-final/SUMMARY.md`.

## Step 6: Git commit handoff (только по запросу пользователя)

```bash
cd C:/Users/Valerii/Desktop/hub
git add .planning/ docs/ AGENTS.md
git commit -m "docs(planning): sync session state $(date +%Y-%m-%d)"
```

## Шаблон записи в STATE.md

```markdown
## Session Continuity

Last session: <ISO timestamp>
Stopped at: <что делали>
Completed this session:
- <изменение 1> + docs: <файлы>
Resume file: None

## Next session priorities
1. <самое важное следующее действие>
2. ...
```

## Пример реальной session note (`.planning/STATE.md`, формат Hub)

```markdown
## Session note — 2026-07-18 (be1: WIP-лимиты в tracker Kanban)

- Реализован WipLimit в `repo_boards.go` + `service_boards.go`, guard в `handler_boards.go`.
- Цикл нового поведения: `TestBoards_WipLimit*` — FAIL→triage(ROOT=code)→fix→PASS.
- Регрессия домена tracker: `TestAutomationEngine_TriggerOnStatusChange` — FAIL→triage(ROOT=test, сценарий не учитывал лимит)→fix теста→PASS.
- Docs: `docs/system_specification.md` §FR-TRACKER (WIP-лимиты), `.planning/COVERAGE.md` (tracker 82.4%→82.6%).
- verified: true (be-ver, 2026-07-18 23:40) + лог обоих циклов.
```

## Пример handoff при незавершённой задаче (Blockers)

```markdown
## Blockers/Concerns
- TestBasesService_ColumnTypeCast — ROOT неясен (Stop→clarify), эскалировано sa2, ждём уточнения contract в docs/system_specification.md §Bases
- PG-03 secret rotation — manual ops, единственный prod-блокер (см. .planning/production/PG-03-SECRET-ROTATION.md)
```
