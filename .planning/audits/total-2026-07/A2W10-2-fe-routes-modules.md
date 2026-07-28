# A2-W10 агент A2 — routes+modules part 2 (re-audit)

> TASK-008 · Stage A2 · read-only · 2026-07-28  
> HEAD: `0d6a659f` · Зеркало: [`A10-2-fe-routes-modules.md`](A10-2-fe-routes-modules.md) · [`W10-ASSIGNMENTS.md`](W10-ASSIGNMENTS.md) §A2  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| frontend/src/modules/settings/webhooks.tsx | 298 | 1 | RESIDUAL |
| frontend/src/modules/status/status-page.tsx | 79 | 0 | чисто |
| frontend/src/modules/tms/autotest-setup.tsx | 855 | 0 | FIXED-OK |
| frontend/src/modules/tms/components/case-filters.tsx | 170 | 0 | FIXED-OK |
| frontend/src/modules/tms/components/case-table-row.tsx | 204 | 0 | FIXED-OK |
| frontend/src/modules/tms/components/case-tree.tsx | 208 | 0 | FIXED-OK |
| frontend/src/modules/tms/components/coverage-report.tsx | 174 | 1 | RESIDUAL |
| frontend/src/modules/tms/components/create-case-dialog.tsx | 150 | 0 | чисто |
| frontend/src/modules/tms/components/environment-comparison.tsx | 445 | 0 | FIXED-OK |
| frontend/src/modules/tms/components/flaky-tests-report.tsx | 117 | 0 | FIXED-OK |
| frontend/src/modules/tms/components/import-dialog.tsx | 177 | 0 | чисто |
| frontend/src/modules/tms/components/run-execution-panel.tsx | 159 | 0 | FIXED-OK |
| frontend/src/modules/tms/components/run-header.tsx | 271 | 0 | чисто |
| frontend/src/modules/tms/components/run-results-table.tsx | 283 | 0 | FIXED-OK |
| frontend/src/modules/tms/components/step-editor.tsx | 242 | 0 | FIXED-OK |
| frontend/src/modules/tms/components/test-case-metadata.tsx | 245 | 0 | чисто |
| frontend/src/modules/tms/components/traceability-matrix.tsx | 352 | 0 | FIXED-OK |
| frontend/src/modules/tms/components/trends-chart.tsx | 144 | 0 | FIXED-OK |
| frontend/src/modules/tms/components/version-history.tsx | 104 | 0 | FIXED-OK* |
| frontend/src/modules/tms/environments.tsx | 225 | 1 | RESIDUAL |
| frontend/src/modules/tms/trigger-ci-dialog.tsx | 279 | 0 | FIXED-OK |
| frontend/src/modules/tracker/burndown-chart.tsx | 125 | 1 | RESIDUAL |
| frontend/src/modules/tracker/gantt-chart.tsx | 258 | 0 | чисто |
| frontend/src/modules/tracker/github-links.tsx | 309 | 1 | RESIDUAL |
| frontend/src/modules/tracker/gitlab-pipeline-status.tsx | 298 | 0 | чисто |
| frontend/src/modules/tracker/import-export.tsx | 174 | 1 | RESIDUAL |
| frontend/src/modules/tracker/issue-description.tsx | 168 | 1 | RESIDUAL |
| frontend/src/modules/tracker/issue-header.tsx | 97 | 0 | FIXED-OK* |
| frontend/src/modules/tracker/issue-message-links-panel.tsx | 81 | 0 | FIXED-OK |

\* authz UI (`canMutate`) закрыт для webhooks/environments; header/version-history — destructive без role-гейта смотрит parent route (не reopen).

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | A1 ref |
|----|-----|-----|-------------|----------|--------|
| A2W10-A2-001 | RESIDUAL | P2 | webhooks.tsx:54–67 | `canMutate` ✅; `fetchEndpoints` всё ещё `try/finally` без catch → fail=empty | W10-A2-004 |
| A2W10-A2-002 | RESIDUAL | P2 | environments.tsx:49–57 | Manual `listEnvironments` без error UI; не shared Query key с comparison | W10-A2-005 |
| A2W10-A2-003 | RESIDUAL | P2 | github-links.tsx:235 | `catch {}` → «Нет привязанных PR» | W10-A2-006 |
| A2W10-A2-004 | RESIDUAL | P2 | burndown-chart.tsx:28–34 | Zustand `fetchBurndown`; `length===0` = empty, нет loading/error | W10-A2-010 |
| A2W10-A2-005 | RESIDUAL | P2 | import-export.tsx:30–32,48–49,67–69 | JWT через `api` ✅; preview/commit/export catch молчаливы | W10-A2-016 |
| A2W10-A2-006 | RESIDUAL | P2 | issue-description.tsx:21,86 | `parsePlateValue` + `as Value`; `descriptionJson: unknown` без zod-guard | W10-A2-018 |
| A2W10-A2-007 | RESIDUAL | P2 | coverage-report.tsx:43–45 | `tr.onClick` expand без `role`/`tabIndex`/keyboard | W10-A2-019 |
| A2W10-A2-008 | NEW | P3 | case-tree.tsx:136–149 | `CaseTreePanel` нигде не импортируется — мёртвый export | — |

## FIXED-OK

| A1 ID | Что проверено |
|-------|----------------|
| W10-A2-001 | import-export → `api.post` / `api.getBlob` (Bearer) |
| W10-A2-002 | issue-message-links → `/messenger/$channelId` + `channel_id` |
| W10-A2-003 | env-comparison: `submitted && comparison` + reset on env change |
| W10-A2-007/008/009 | coverage/traceability/trends: `isError` |
| W10-A2-011 | flaky key `tms-flaky` |
| W10-A2-012 | «+» сьют только если `onCreateSuite` |
| W10-A2-013 | autotest `clearTimeout` |
| W10-A2-014 | webhooks/envs: `canMutateProject` |
| W10-A2-015 | CI limit + `casesTruncated` warning |
| W10-A2-017 | `Partial<CaseFilters…>` |
| W10-A2-021–025 | mojibake / aria / unlink toast / stepResults removed |
| W10-A2-026 | `formatEnvConfigValue` |

## Сводка

| Метрика | Значение |
|---------|----------:|
| Файлов | 29 / 29 / 0 |
| NEW | 1 |
| REGRESS | 0 |
| RESIDUAL | 7 |
| P0 / P1 / P2 / P3 (open) | 0 / 0 / 7 / 1 |
