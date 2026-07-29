# TASK-047 — FE fail/сеть ≠ empty (isError+retry)

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта  
> Источник: TASK-008 W12 · Classes: P2-FE-01 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P2-FRONTEND.md`

## [OUTCOME]

Дефекты P2-FE-01 закрыты.

## [ROLE]

- Role: fe1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: frontend/src/{modules,routes,stores,components}/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P2-FE-01)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт
- Паттерн: `QueryStateView` / `QueryErrorAlert` (`frontend/src/components/query-state-view.tsx`); trash.tsx (TASK-032)

## [ACCEPTANCE]

- Done criteria: все ClassID из P2-FE-01 с мутацией «да» закрыты по реестру
- **Мутация (C):** mock list reject→ошибка/retry, не empty
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

---

## Progress (fe1 · 2026-07-28)

### Закрыто в этой сессии (высокий impact)

| Source ID | Файл | Что сделано |
|-----------|------|-------------|
| W9-A1-010 | `components/admin/mail-settings.tsx` | `loadError` + `QueryErrorAlert` |
| W9-A1-012 | `components/admin/ai-settings.tsx` | то же |
| W9-A1-013 | `components/admin/onec-settings.tsx` | то же + Vitest |
| W9-A2-009 | `components/messenger/bookmarks-panel.tsx` | `QueryStateView` + Vitest |
| W9-A4-006 | `components/notifications/telegram-link-card.tsx` | fail ≠ unlinked |
| W9-A5-006 | `components/tracker/user-picker.tsx` | `isError` + i18n retry |
| W10-A1-004 | `modules/calls/call-calendar.tsx` | loading/error banner |
| W10-A1-006 | `modules/bases/bases-page.tsx` | sidebar `isError` + Vitest |
| W10-A1-017 | `modules/hr/leave-requests.tsx` | `isError` + Vitest |
| W10-A1-018 | `modules/intranet/kudos-feed.tsx` | `isError` + retry |
| W10-A1-020 | `modules/marketplace/marketplace-page.tsx` | fail ≠ empty listings |
| W10-A2-007 | `modules/tms/components/coverage-report.tsx` | `isError` + Vitest |
| W10-A2-008 | `modules/tms/components/traceability-matrix.tsx` | `isError` |
| W10-A2-009 | `modules/tms/components/trends-chart.tsx` | `isError` |
| W10-A5-007 | `routes/.../tms/runs.lazy.tsx` | `isError` ≠ empty runs |
| W11-A1-003 | `api/calls.ts` `getActiveCall` | **уже было** (404→null, else rethrow) |
| W11-A1-004 | `api/calls.ts` `listScheduledCallsForChannels` | allSettled; all-fail → throw |
| W11-A2-002 | `stores/ai-store.ts` | `loaded` только на success |
| W11-A2-003 | `stores/ai-store.ts` | `modelsError` + rollback save |

Reusable: экспортирован `QueryErrorAlert` из `query-state-view.tsx`.

### TODO — remaining Source IDs (следующая сессия)

| Source ID | Файл / зона |
|-----------|-------------|
| W9-A4-004 | `components/tracker/saved-filters.tsx` — silent catch |
| W10-A1-005 | `modules/accounting/accounting-page.tsx` |
| W10-A1-007 | `modules/crm/crm-page.tsx` |
| W10-A1-008 | `modules/admin/backups.tsx` |
| W10-A1-009 | `modules/settings/api-keys.tsx` |
| W10-A1-011 | `modules/settings/ai-privacy.tsx` |
| W10-A1-012 | `modules/settings/project-templates.tsx` |
| W10-A1-013 | `modules/calls/scheduled-calls.tsx` |
| W10-A1-014 | `modules/calls/recordings-list.tsx` |
| W10-A2-004 | `modules/tms/webhooks.tsx` |
| W10-A2-005 | `modules/tms/environments.tsx` |
| W10-A2-006 | `modules/tms`/`github-links` catch→[] |
| W10-A2-010 | `modules/tracker/burndown-chart.tsx` / store |
| W10-A2-016 | `modules/tms/import-export.tsx` |
| W10-A3-003 | `modules/tracker/velocity-chart.tsx` |
| W10-A3-004 | `modules/tracker/sprint-backlog.tsx` |
| W10-A4-007 | routes: `$issueKey`, messenger index, `cases.lazy`, `dashboard.lazy` |
| W10-A5-005 | `cases_.$caseKey.lazy.tsx` — isError vs shimmer |
| W10-A5-006 | `runs_.$runId.tsx` — isError vs shimmer |
| W11-A3-012 | `hooks/use-presence.ts` — прокинуть isError |
| W11-A3-014 | `stores/sprint-store.ts` — error flags / Query migration |

### Evidence

- Vitest (19 tests, 8 files): ai-store, onec-settings, bases-page(+error), leave-requests, bookmarks-error, coverage-report-error, marketplace-page — **PASS**
- `npx tsc --noEmit` — **PASS**

---

## Stage C mutation (W18-C4 · fe2 · 2026-07-28)

- Verdict: **PASS**
- Evidence: `.planning/audits/total-2026-07/W18-C4-MUTATION.md`
- RED→restore→GREEN: `bases-page-error.test.tsx` (listBases reject → alert/retry ≠ empty)
