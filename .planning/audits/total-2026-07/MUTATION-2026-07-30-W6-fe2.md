# MUTATION W6 slot 2 — TASK-127 (fe2) · 2026-07-30

> Checker: **fe2** (!= fe1 criteria author) · Env: local Vitest · Criteria: `.planning/backlog/TASK-127-2026-07-29-a3-10-4-fe-error-empty-settings.md` `[MUTATION CRITERIA]`

## Verdict

**6/6 PROVEN** (runnable) · transitions twin gap → **N/A** (нет отдельного `it`) · статус TASK-127 → **Закрыта**

| ID | Verdict | RED test (`it`) | Mutation (BREAK) | Layer |
|----|---------|-----------------|------------------|-------|
| A3W10-4-001 | **PROVEN** | `shows alert + retry instead of welcoming empty when channels fail` | `if (false && isError)` — fail → welcome empty | frontend |
| A3W10-4-002 | **PROVEN** | `shows alert + retry instead of «Проект не найден» when getProject fails` | `if (false && isError)` — fail → «Проект не найден» | frontend |
| A3W10-4-003 | **PROVEN** | `shows alert + retry instead of «Нет тест-кейсов» when listTestCases fails` | `false && isError ?` — fail → empty «Нет тест-кейсов» | frontend |
| A3W10-4-004 | **PROVEN** | `shows alert, hides empty copy and create UI when listLabels fails` | disable `labelsError`/`groupsError` alerts; `loadError = false` | frontend |
| A3W10-4-005 | **PROVEN** | `shows alert and blocks create when statuses query fails` | disable `statusesError` alert; `loadError = false` | frontend |
| A3W10-4-005 transitions twin | **N/A** | нет отдельного `it` на transitions-only fail | — | — |
| A3W10-4-006 | **PROVEN** | `GitHub: …` + `GitLab: …` | убрать `isError && QueryErrorAlert`; empty + connect при fail | frontend |

Production restored (`git checkout --`). Mock-gap: нет (RTL на prod-роутах).

## Per-ID evidence

### A3W10-4-001 — messenger index isError

```text
BREAK: messenger/index.tsx — if (false && isError) { QueryErrorAlert }
→ FAIL shows alert + retry instead of welcoming empty when channels fail
RESTORE
→ PASS same test
```

### A3W10-4-002 — project overview isError

```text
BREAK: projects/$projectId/index.lazy.tsx — if (false && isError) { QueryErrorAlert }
→ FAIL shows alert + retry instead of «Проект не найден» when getProject fails
RESTORE
→ PASS same test
```

### A3W10-4-003 — TMS cases list isError

```text
BREAK: tms/cases.lazy.tsx — false && isError ? (alert) : empty
→ FAIL shows alert + retry instead of «Нет тест-кейсов» when listTestCases fails
RESTORE
→ PASS same test
```

### A3W10-4-004 — labels settings isError + create block

```text
BREAK: settings/labels.tsx — false && labelsError alert; groupsError alert off; loadError = false
→ FAIL shows alert, hides empty copy and create UI when listLabels fails
RESTORE
→ PASS same test
```

### A3W10-4-005 — workflow statuses isError + create block

```text
BREAK: settings/workflow.tsx — false && statusesError alert; loadError = false
→ FAIL shows alert and blocks create when statuses query fails
RESTORE
→ PASS same test
```

**Gap:** transitions-only fail — **N/A** (criteria note: primary `it` = statuses; отдельного `it` на transitions twin нет).

### A3W10-4-006 — GitHub + GitLab list isError

```text
BREAK: settings/integrations.lazy.tsx — disable GH+GL QueryErrorAlert; allow empty + connect on fail
→ FAIL GitHub: alert + no empty copy + no connect form on list fail
→ FAIL GitLab: alert + no empty copy + no connect form on list fail
RESTORE
→ PASS both
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run \
  ../testframework/stages/unit/frontend-test/domains/messenger/messenger-index-error.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/project-overview-error.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tms/cases-list-error.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/settings-labels-error.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/settings-workflow-error.test.tsx \
  ../testframework/stages/unit/frontend-test/domains/tracker/settings-integrations-error.test.tsx
→ Test Files 6 passed · Tests 7 passed (~4.4s) after full restore
```

## Assumptions / unresolved

- Коммит не создавался (инструкция слота).
- Criteria authored by fe1; mutation prove by fe2 (W6 slot 2).
- transitions twin (workflow) = N/A, не BLOCKED.

## Handoff

TASK-127 **Закрыта** · evidence этот файл · README синхронизирован.
