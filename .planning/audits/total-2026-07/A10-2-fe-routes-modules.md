# W10 агент A2 — routes+modules part 2

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус FE: Query keys / инвалидация, authz UI, hooks leaks, `any`/`unknown`.  
> Список: W10-ASSIGNMENTS.md §A2 (29 файлов).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/modules/settings/webhooks.tsx` | 269 | 2 | замечания |
| `frontend/src/modules/status/status-page.tsx` | 87 | 0 | чисто |
| `frontend/src/modules/tms/autotest-setup.tsx` | 908 | 1 | замечания |
| `frontend/src/modules/tms/components/case-filters.tsx` | 179 | 1 | замечания |
| `frontend/src/modules/tms/components/case-table-row.tsx` | 204 | 1 | замечания |
| `frontend/src/modules/tms/components/case-tree.tsx` | 205 | 1 | замечания |
| `frontend/src/modules/tms/components/coverage-report.tsx` | 162 | 2 | замечания |
| `frontend/src/modules/tms/components/create-case-dialog.tsx` | 155 | 0 | чисто |
| `frontend/src/modules/tms/components/environment-comparison.tsx` | 467 | 2 | замечания |
| `frontend/src/modules/tms/components/flaky-tests-report.tsx` | 126 | 1 | замечания |
| `frontend/src/modules/tms/components/import-dialog.tsx` | 189 | 0 | чисто |
| `frontend/src/modules/tms/components/run-execution-panel.tsx` | 175 | 1 | замечания |
| `frontend/src/modules/tms/components/run-header.tsx` | 277 | 0 | чисто |
| `frontend/src/modules/tms/components/run-results-table.tsx` | 293 | 1 | замечания |
| `frontend/src/modules/tms/components/step-editor.tsx` | 246 | 1 | замечания |
| `frontend/src/modules/tms/components/test-case-metadata.tsx` | 256 | 0 | чисто |
| `frontend/src/modules/tms/components/traceability-matrix.tsx` | 353 | 1 | замечания |
| `frontend/src/modules/tms/components/trends-chart.tsx` | 130 | 1 | замечания |
| `frontend/src/modules/tms/components/version-history.tsx` | 102 | 1 | замечания |
| `frontend/src/modules/tms/environments.tsx` | 218 | 2 | замечания |
| `frontend/src/modules/tms/trigger-ci-dialog.tsx` | 277 | 1 | замечания |
| `frontend/src/modules/tracker/burndown-chart.tsx` | 130 | 1 | замечания |
| `frontend/src/modules/tracker/gantt-chart.tsx` | 275 | 0 | чисто |
| `frontend/src/modules/tracker/github-links.tsx` | 338 | 1 | замечания |
| `frontend/src/modules/tracker/gitlab-pipeline-status.tsx` | 314 | 0 | чисто |
| `frontend/src/modules/tracker/import-export.tsx` | 179 | 2 | замечания |
| `frontend/src/modules/tracker/issue-description.tsx` | 168 | 1 | замечания |
| `frontend/src/modules/tracker/issue-header.tsx` | 95 | 1 | замечания |
| `frontend/src/modules/tracker/issue-message-links-panel.tsx` | 65 | 2 | замечания |

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W10-A2-001 | P1 | `import-export.tsx:22-31`; `:41-46` | Preview/commit импорта через голый `fetch("/api/...")` без `Authorization` | `api/client.ts` кладёт Bearer JWT из auth-store; без него эндпоинт → 401. Сценарий «Импорт задач» сломан для авторизованных пользователей | Использовать `api`/`apiForm` из `@/api/client` (или общий upload-helper с теми же заголовками) |
| W10-A2-002 | P1 | `issue-message-links-panel.tsx:47-52` | `Link to="/messenger"`; в DTO есть `channel_id`, но в URL не передаётся | Сценарий «открыть связанный тред» ведёт на корень мессенджера, не в канал/тред — ссылка бесполезна | `to="/messenger/$channelId"` + `params={{ channelId: link.channel_id }}` (+ search/hash на `message_id` если есть) |
| W10-A2-003 | P1 | `environment-comparison.tsx:224-227`; `:315-454` | При смене env: `setSubmitted(false)`, query `enabled=false`, но UI рендерит `{comparison && …}` из кэша | После выбора других окружений таблица показывает **старое** сравнение, пока снова не нажмут «Сравнить» — ложные различия | Сбрасывать `comparison` (`placeholderData: undefined` + key reset) или рендерить таблицу только при `submitted && comparison` с match env ids |
| W10-A2-004 | P2 | `webhooks.tsx:34-42`; `:48-64` | `fetchEndpoints`/мутации: `try/finally` без `catch`; ошибки create/toggle/delete глотаются | Сбой API неотличим от пустого списка; пользователь не видит 403/сеть | `catch` + toast; лучше `useQuery`/`useMutation` с `onError` |
| W10-A2-005 | P2 | `environments.tsx:34-43`; `:50-71` | Тот же manual-fetch паттерн без обработки ошибок; нет TanStack Query | Нет cache/invalidation с `EnvironmentComparison` (`["tms-environments", projectId]`) — списки расходятся | Перейти на `listEnvironments` + общий queryKey; toast на ошибки |
| W10-A2-006 | P2 | `github-links.tsx:235-237` | `catch { /* пустой список */ }` без UI ошибки | Ошибка сети/403 выглядит как «Нет привязанных PR» | Разделить empty vs error; toast/`isError` |
| W10-A2-007 | P2 | `coverage-report.tsx:104-123` | `useQuery` без `isError`; `!data` → empty «Нет задач…» | Ошибка API маскируется под empty state | Ветка `isError` + retry |
| W10-A2-008 | P2 | `traceability-matrix.tsx:105-206` | То же: нет `isError`, error → empty card | Ложный empty UX | Как W10-A2-007 |
| W10-A2-009 | P2 | `trends-chart.tsx:64-80` | То же для трендов | Ложный empty UX | Как W10-A2-007 |
| W10-A2-010 | P2 | `burndown-chart.tsx:28-47` | Zustand `fetchBurndown` в `useEffect`; `burndown.length === 0` = empty | Нет loading/error; ошибка fetch = «Нет данных»; дублирование Query↔store | Query key `["sprint-burndown", projectId, sprintId]` + `QueryStateView` / флаги в store |
| W10-A2-011 | P2 | `flaky-tests-report.tsx:34` | `queryKey: ["flaky-tests", projectId]` вне префикса `tms-*` | Непоследовательные ключи → риск пропустить invalidation при мутациях TMS | Унифицировать: `["tms-flaky", projectId]` + обновить инвалидации |
| W10-A2-012 | P2 | `case-tree.tsx:153-159` | Кнопка «+» / `aria-label="Создать сьют"` без `onClick` | Мёртвый UI: пользователь ждёт создания сьюта | Подключить create-suite dialog или скрыть кнопку до готовности |
| W10-A2-013 | P2 | `autotest-setup.tsx:19-23` | `setTimeout(() => setCopied(false), 2000)` без cleanup | Unmount до 2s → setState на размонтированном + timer leak | `useEffect` cleanup / `let cancelled` / `useRef` timer |
| W10-A2-014 | P2 | `webhooks.tsx:89-95`; `environments.tsx:94-100`; `issue-header.tsx:83-91`; `version-history.tsx:76-84` | Мутации (create/delete webhook/env, delete issue, rollback) всегда видны | Нет authz UI: viewer/guest видит действия → 403 без объяснения (backend-only guard) | Скрывать/disable по project role / permission hook |
| W10-A2-015 | P2 | `trigger-ci-dialog.tsx:34-37` | `listTestCases(..., { limit: 500 })` без пагинации/предупреждения | Проекты >500 active cases: неполный выбор для CI; silent truncation | Cursor/infinite или явный warning «показаны первые 500» |
| W10-A2-016 | P2 | `import-export.tsx:29-31`; `:48-52` | При `!resp.ok` preview/import молча no-op | Даже после фикса JWT пользователь не увидит 4xx/5xx | Toast + текст ошибки из тела ответа |
| W10-A2-017 | P2 | `case-filters.tsx:24` | `updateFilter: (patch: Record<string, unknown>)` | Обход типизации фильтров (`any`-эквивалент) | Типизировать `Partial<CaseFiltersState>` |
| W10-A2-018 | P2 | `issue-description.tsx:77`; `:145`; `:155` | `descriptionJson: unknown` + `as Value` / `as unknown as Value` | Небезопасный cast Plate Value; битый JSON → runtime crash в editor | Zod/type-guard или валидация перед cast |
| W10-A2-019 | P2 | `coverage-report.tsx:43-46` | Раскрытие строки через `tr.onClick` без `role`/`tabIndex`/keyboard | Недоступно с клавиатуры (a11y) | `button` в ячейке или `role="button"` + Enter/Space |
| W10-A2-020 | P2 | `environment-comparison.tsx:309` | `(compError as { message?: string }).message` | Cast вместо нормализации API error | `toastApiError` / helper `getErrorMessage(err)` |
| W10-A2-021 | P3 | `run-results-table.tsx:220` | Заголовок колонки «Испол��итель» (mojibake) | Битая UTF-8 в UI | Исправить на «Исполнитель» |
| W10-A2-022 | P3 | `step-editor.tsx:130-150` | Кнопка удаления шага: icon-only, нет `aria-label` | AT не объявляет действие | `aria-label="Удалить шаг"` |
| W10-A2-023 | P3 | `case-table-row.tsx:158-195` | Иконки auto/manual: SVG `aria-hidden`, только `title` | Индикатор автоматизации слабо доступен AT | `aria-label` на обёртке |
| W10-A2-024 | P3 | `issue-message-links-panel.tsx:19-31`; `:53-58` | Нет `isError`; unlink без toast/`onError` | Тихий fail при отвязке; ошибка списка → пустой null | Error UI + `toastApiError` |
| W10-A2-025 | P3 | `run-execution-panel.tsx:80-84` | `stepResults` в state, setter не используется | Мёртвое состояние; step results никогда не редактируются в панели | Убрать или подключить UI шагов |
| W10-A2-026 | P3 | `environments.tsx:17`; `:180-183` | `config?: Record<string, unknown>`; `${v}` в UI | Non-string values → `[object Object]` | Типизировать config / `JSON.stringify` |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `status-page.tsx` | Query key `["public", "status"]`; loading/`isError`→overall degraded; refetch 30s; без auth (публичная) — ок |
| `create-case-dialog.tsx` | Mutation + invalidate `["tms-cases", projectId]`; `toastApiError`; controlled Dialog |
| `import-dialog.tsx` | Invalidate prefix `["tms-cases"]`; error toast; file input с `htmlFor` |
| `run-header.tsx` | Presentational; status badges с текстом; кнопки по `run.status` |
| `test-case-metadata.tsx` | Read-only meta; labels/components keys по id |
| `gantt-chart.tsx` | Infinite query keys; `QueryStateView`; batch links; virtualization threshold |
| `gitlab-pipeline-status.tsx` | Query + error/empty/loading; refetchInterval; `rel=noopener` |
| Query keys (хорошие) | `tms-cases`, `tms-coverage`, `tms-env-comparison`, `tms-run-detail`, `gantt-issues`, `gitlab-pipelines` — стабильные, с projectId |
| Hooks order | Во всех 29 файлах хуки до conditional return; Rules of Hooks не нарушены |
| Shadcn | Select/Dialog/Button/Badge — только `className`/composition |

**Фокус-вердикт:** критично — **импорт задач без JWT** и **битые ссылки на треды** + **stale env comparison**. Системно — manual fetch без Query/ошибок (webhooks/environments/github) и empty↔error в TMS-отчётах. Authz UI почти нигде не гейтит destructive actions.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **29 / 29 / 0** |
| Находок | **26** |
| P0 | 0 |
| P1 | 3 |
| P2 | 17 |
| P3 | 6 |

### Топ-3

1. **W10-A2-001 (P1)** — `import-export`: голый `fetch` без Bearer → импорт задач не работает под JWT-клиентом.
2. **W10-A2-002 (P1)** — `issue-message-links-panel`: ссылка на `/messenger` без `channel_id` — тред не открывается.
3. **W10-A2-003 (P1)** — `environment-comparison`: после смены окружений UI показывает устаревшее сравнение из кэша.
