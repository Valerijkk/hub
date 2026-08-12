# A3 W10 агент A2 — routes+modules part 2 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W10-2-fe-routes-modules.md](A2W10-2-fe-routes-modules.md) · Stage A: [A10-2-fe-routes-modules.md](A10-2-fe-routes-modules.md)  
> Периметр: 29 файлов (W10-ASSIGNMENTS §A2) · Прочитано: 29 · Пропущено: 0  
> Фокус: auth guards · XSS · client-side IDOR assumptions

## Сверка с A2 (A2W10-A2-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W10-A2-001 | RESIDUAL → FIXED-OK | закрыт | TASK-093: `webhooks.tsx:64–75` catch + `loadError` + `QueryErrorAlert` + toast create |
| A2W10-A2-002 | RESIDUAL → FIXED-OK | закрыт | TASK-093: `environments.tsx:52–64` catch + `loadError` + toast; `formatEnvConfigValue` |
| A2W10-A2-003 | RESIDUAL → FIXED-OK | закрыт | TASK-093: `github-links.tsx:230–272` `loadError` ≠ «Нет привязанных PR» |
| A2W10-A2-004 | RESIDUAL → FIXED-OK* | error закрыт | chart `error && burndown.length===0` + retry (`burndown-chart.tsx:34–56`); store set error (`sprint-store.ts:109–118`). **Loading residual → A3W10-2-005** |
| A2W10-A2-005 | RESIDUAL → FIXED-OK | закрыт | `import-export.tsx` — `api.post`/`api.getBlob` + toast на preview/commit/export |
| A2W10-A2-006 | RESIDUAL → FIXED-OK | закрыт | `issue-description.tsx:16–37` zod `plateValueSchema.safeParse` |
| A2W10-A2-007 | RESIDUAL → FIXED-OK | закрыт | `coverage-report.tsx:46–53` `role=button` + `tabIndex` + Enter/Space |
| A2W10-A2-008 | NEW → FIXED-OK | закрыт | `CaseTreePanel` wired in `cases.lazy.tsx:151+`; `onCreateSuite?` hides dead «+» |

**REGRESS:** 0 · **RESIDUAL (тонкий):** 1 (burndown loading) · **NEW (этот слот):** 4

### Cascade FIXED-OK (Stage A W10-A2-* / A2 FIXED-OK)

| Stage A / A2 FIXED | Проверено в HEAD |
|--------------------|------------------|
| W10-A2-001 JWT import | `api.post` / `api.getBlob` (Bearer) |
| W10-A2-002 messenger links | `to="/messenger/$channelId"` + `channel_id` |
| W10-A2-003 env comparison | `submitted && comparison` (`environment-comparison.tsx:322`) |
| W10-A2-007…009 isError | coverage / traceability / trends |
| W10-A2-011 flaky key | `["tms-flaky", projectId]` |
| W10-A2-012 create suite | `onCreateSuite?` |
| W10-A2-013 clearTimeout | `autotest-setup.tsx:18–31` |
| W10-A2-014 canMutate | webhooks / environments; issue-header `canDelete`; version-history `canRollback` |
| W10-A2-015 CI limit warn | `casesTruncated` + `CI_CASES_FETCH_LIMIT` |
| W10-A2-017 Partial filters | `Partial<CaseFiltersProps["filters"]>` |
| W10-A2-021…025 | mojibake / aria / unlink toast / stepResults removed |
| W10-A2-026 formatEnvConfig | `formatEnvConfigValue` |

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/modules/settings/webhooks.tsx` | 337 | 0 | FIXED-OK authz+error |
| `frontend/src/modules/status/status-page.tsx` | 87 | 0 | чисто (public status) |
| `frontend/src/modules/tms/autotest-setup.tsx` | 917 | 0 | FIXED-OK timer; embeds TriggerCI |
| `frontend/src/modules/tms/components/case-filters.tsx` | 179 | 0 | FIXED-OK |
| `frontend/src/modules/tms/components/case-table-row.tsx` | 209 | 0 | FIXED-OK a11y |
| `frontend/src/modules/tms/components/case-tree.tsx` | 211 | 0 | FIXED-OK wired |
| `frontend/src/modules/tms/components/coverage-report.tsx` | 192 | 0 | FIXED-OK isError+a11y |
| `frontend/src/modules/tms/components/create-case-dialog.tsx` | 155 | 0 | чисто |
| `frontend/src/modules/tms/components/environment-comparison.tsx` | 474 | 0 | FIXED-OK stale+error |
| `frontend/src/modules/tms/components/flaky-tests-report.tsx` | 126 | 0 | FIXED-OK |
| `frontend/src/modules/tms/components/import-dialog.tsx` | 189 | 0 | чисто |
| `frontend/src/modules/tms/components/run-execution-panel.tsx` | 167 | 0 | FIXED-OK |
| `frontend/src/modules/tms/components/run-header.tsx` | 277 | 0 | чисто |
| `frontend/src/modules/tms/components/run-results-table.tsx` | 293 | 0 | FIXED-OK UTF-8 |
| `frontend/src/modules/tms/components/step-editor.tsx` | 250 | 0 | FIXED-OK aria |
| `frontend/src/modules/tms/components/test-case-metadata.tsx` | 256 | 0 | чисто |
| `frontend/src/modules/tms/components/traceability-matrix.tsx` | 375 | 0 | FIXED-OK isError |
| `frontend/src/modules/tms/components/trends-chart.tsx` | 152 | 0 | FIXED-OK isError |
| `frontend/src/modules/tms/components/version-history.tsx` | 108 | 0 | FIXED-OK canRollback |
| `frontend/src/modules/tms/environments.tsx` | 256 | 0 | FIXED-OK authz+error |
| `frontend/src/modules/tms/trigger-ci-dialog.tsx` | 300 | 2 | NEW authz + noopener |
| `frontend/src/modules/tracker/burndown-chart.tsx` | 155 | 1 | RESIDUAL loading |
| `frontend/src/modules/tracker/gantt-chart.tsx` | 275 | 0 | чисто QueryStateView |
| `frontend/src/modules/tracker/github-links.tsx` | 358 | 2 | NEW XSS + authz UI |
| `frontend/src/modules/tracker/gitlab-pipeline-status.tsx` | 314 | 1 | NEW XSS (shared) |
| `frontend/src/modules/tracker/import-export.tsx` | 189 | 1 | NEW dead export |
| `frontend/src/modules/tracker/issue-description.tsx` | 195 | 0 | FIXED-OK zod |
| `frontend/src/modules/tracker/issue-header.tsx` | 101 | 0 | FIXED-OK canDelete |
| `frontend/src/modules/tracker/issue-message-links-panel.tsx` | 88 | 0 | FIXED-OK channel link |

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W10-2-001 | NEW | P2 | `github-links.tsx:305–309,324–328`; `gitlab-pipeline-status.tsx:272–276` | `href={link.url}` / `href={pipeline.web_url}` без whitelist схем | Messenger уже санитизирует (`render-content.tsx` `SAFE_URL_RE` → `#`). Здесь сырой URL из API → при `javascript:`/`data:` в DTO возможен XSS/навигация. Defense-in-depth: FE не должен доверять схеме | Вынести `sanitizeUrl` в `@/lib/safe-url`; применять ко всем external `href`/`window.open` |
| A3W10-2-002 | NEW | P2 | `trigger-ci-dialog.tsx:21–32,268–286`; `github-links.tsx:284–291` | Viewer/guest видит «Запустить пайплайн» и «+ Создать ветку» без `canMutateProject` | W10-A2-014 закрыт для webhooks/envs/delete/rollback; тот же класс на CI/branch — UX 403 без объяснения; authz только на backend | `canMutate` prop / `canMutateProject(project.my_role)`; скрыть destructive/expensive actions |
| A3W10-2-003 | NEW | P3 | `trigger-ci-dialog.tsx:89` | `window.open(data.web_url, "_blank")` без `noopener` и без scheme-check | Reverse tabnabbing; тот же класс URL, что A3W10-2-001 | `window.open(sanitizeUrl(url), "_blank", "noopener,noreferrer")` или `<a rel="noopener noreferrer">` |
| A3W10-2-004 | NEW | P3 | `import-export.tsx:6` (весь модуль) | `ImportExportToolbar` нигде не импортируется (только определение) | Мёртвый export в периметре; JWT/toast уже FIXED, но authz UI отсутствует — при будущем wire viewer увидит Импорт/Экспорт | Либо удалить/wire в issues toolbar с `canMutate`, либо пометить deprecated |
| A3W10-2-005 | RESIDUAL | P3 | `burndown-chart.tsx:28–72`; `sprint-store.ts:109–118` | `fetchBurndown` не ставит `loading`; chart не читает `loading` → empty «Нет данных» на первом fetch | A2W10-A2-004 error≠empty закрыт; loading-ветка осталась | `set({ loading: true })` в store + chart: loading → skeleton, не empty |

## Client-side IDOR / auth assumptions (проверено, не reopen)

| Зона | FE assumption | Backend evidence | Вердикт |
|------|---------------|------------------|---------|
| `listGitHubLinks(issueKey)` без `projectId` в path | auth + membership на сервере | `github/service.go:474–489` `IsProjectMember` | FIXED-OK (BE) |
| `updateWebhook`/`deleteWebhook`/`listDeliveries` by id | project из БД, не client | `apigateway/handler.go:81–93` `authorizeEndpoint` | FIXED-OK (BE) |
| Route `projectId` props | UI trust; API enforces | стандарт FE pattern | OK (не FE IDOR) |
| Plate `descriptionJson` | zod guard before editor | `parsePlateValue` safeParse | FIXED-OK |
| Status page public | no JWT | intentional V4-06 | OK |

## FIXED-OK (сводка фокуса)

| Зона | Проверено в HEAD |
|------|------------------|
| Authz UI webhooks/envs | `canMutateProject` + hide create/toggle/delete |
| Authz UI issue delete / version rollback | `canDelete` / `canRollback` from parent route |
| XSS text nodes | React escape; webhook URL as text (не href); status/description text |
| JWT import/export API | `api` client Bearer (даже если toolbar dead) |
| GitHub links error≠empty | loadError + toast |
| Messenger deep-link | channel_id в URL |
| No `dangerouslySetInnerHTML` | в 29 файлах слота — 0 |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 2 |
| P3 | 3 |
| **Total** | **5** |

| Метрика | Значение |
|---------|----------|
| Файлов | **29 / 29 / 0** |
| Активных | **5** (NEW 4 · REGRESS 0 · RESIDUAL 1) |
| FIXED-OK (A2 actives) | **8/8** (burndown error; loading → residual) |
| FIXED-OK (Stage A cascade) | **14+** без регресса |

### Топ-3

1. **A3W10-2-001 (P2, NEW)** — external `href` без `sanitizeUrl` (github-links + gitlab-pipeline).
2. **A3W10-2-002 (P2, NEW)** — нет authz UI на Trigger CI и Create branch (viewer→403).
3. **A3W10-2-005 (P3, RESIDUAL)** — burndown loading всё ещё = empty.

---

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=2 · P3=3 · Total=5.  
Топ: A3W10-2-001 · A3W10-2-002 · A3W10-2-005.
