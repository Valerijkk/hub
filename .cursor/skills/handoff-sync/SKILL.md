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
| Tests / coverage | `.planning/COVERAGE.md`, `docs/test_plan.md` |
| CI / deploy | `docs/stack_and_tools.md`, `deployments/.env.example`, `docs/wiki/DEPLOY.md` |
| Security | `docs/security_guidelines.md` |
| Waves / gaps | `.planning/TASKS-CLOSE-GAPS.md`, `.planning/PLAN-CLOSE-GAPS.md` |
| Всегда | `.planning/STATE.md`, `AGENTS.md` (если менился статус) |

## Step 1: Прочитать текущее состояние

```
Прочитай:
- .planning/STATE.md         — текущая позиция в roadmap
- .planning/ROADMAP.md       — статус фаз
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
