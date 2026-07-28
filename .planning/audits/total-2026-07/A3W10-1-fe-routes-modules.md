# A3 W10 агент A1 — routes+modules part 1 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W10-1-fe-routes-modules.md](A2W10-1-fe-routes-modules.md)  
> Stage A: [A10-1-fe-routes-modules.md](A10-1-fe-routes-modules.md) · [W10-ASSIGNMENTS.md](W10-ASSIGNMENTS.md) §A1  
> Remediation: TASK-093 (`A2W10-A1-001…007` → FIXED)  
> Периметр: 29 файлов · Прочитано: 29 · Пропущено: 0  
> Фокус: TanStack Router only · auth guards · IDOR via client params · XSS · error boundaries

## Сверка с A2 (A2W10-A1-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W10-A1-001 | RESIDUAL → **FIXED-OK** (ядро) + residual P3 | accounts/journal `isError` early return | `accounting-page.tsx:33–41,43–46,156–185` — `accountsLoading` / `accountsLoadError` + `QueryErrorAlert`. Вторичные query (invoices/trial/P&L) без `isError` → A3W10-1-001 |
| A2W10-A1-002 | RESIDUAL → **FIXED-OK** | закрыт | `backups.tsx:31–43,50–80,154–160` — catch + `loadError` + toast create/restore/delete |
| A2W10-A1-003 | RESIDUAL → **FIXED-OK** | закрыт | `crm-page.tsx:42–50,205–236` — `contactsLoading` / `contactsError` + `QueryErrorAlert` ≠ «Нет контактов» |
| A2W10-A1-004 | RESIDUAL → **FIXED-OK** | закрыт | `recordings-list.tsx:36–68,107–113` — load/download `toastApiError` + `loadError` alert. Отдельно: `window.open` без noopener → A3W10-1-003 |
| A2W10-A1-005 | RESIDUAL → **FIXED-OK** | закрыт | `api-keys.tsx:31–43,50–69,152–158` — catch list/create/revoke + `QueryErrorAlert`; timer cleanup `22–28` |
| A2W10-A1-006 | RESIDUAL → **FIXED-OK** | закрыт | `ai-privacy.tsx:17–30,49–60` — `loadError \|\| !loaded` → alert, форма не с default `allowExternal` |
| A2W10-A1-007 | RESIDUAL → **FIXED-OK** (page) | page закрыт | `project-templates.tsx:95–133` — `ProjectTemplatesPage` loadError+toast. `TemplateSelector` toast-only → A3W10-1-002 (W10-A1-012 residual) |

**REGRESS:** 0 · **RESIDUAL (частичные):** 2 · **NEW:** 3

## Фокус-чеклист (слот modules, не routes)

| Тема | Вердикт |
|------|---------|
| TanStack Router only | ✅ `useNavigate`/`useSearch` только из `@tanstack/react-router` (`intranet-hub`, `unified-calendar`, `deals-pipeline`, `org-directory`). `react-router` / BrowserRouter — нет |
| Auth guards | N/A в modules — layout `_authenticated` вне слота A1. Admin UI: `isSystemAdmin` в marketplace/coffee (client gate; API — server) |
| IDOR via client params | ✅ IDs (`projectId`, `recordingId`, base/table) уходят в API; UI не подменяет authz. CreateCall: Select `projectId` + `isValid` требует `!!projectId` |
| XSS | ✅ текст React-escaped; `extractUrls` только `https?:`; `LinkPreview` `rel="noopener noreferrer"`; нет `dangerouslySetInnerHTML` |
| Error boundaries | N/A в этих файлах — обёртки на route-level (слоты A3–A5). Модули сами boundary не ставят |

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/modules/accounting/accounting-page.tsx` | 422 | 1 | RESIDUAL (partial) |
| `frontend/src/modules/admin/backups.tsx` | 239 | 0 | FIXED-OK |
| `frontend/src/modules/bases/bases-page.tsx` | 476 | 1 | замечание (detail query) |
| `frontend/src/modules/calendar/unified-calendar.tsx` | 462 | 0 | чисто |
| `frontend/src/modules/calls/calendar-agenda-view.tsx` | 101 | 0 | чисто |
| `frontend/src/modules/calls/calendar-day-view.tsx` | 125 | 0 | чисто |
| `frontend/src/modules/calls/calendar-month-view.tsx` | 155 | 0 | чисто |
| `frontend/src/modules/calls/calendar-types.ts` | 213 | 0 | чисто |
| `frontend/src/modules/calls/calendar-week-view.tsx` | 170 | 0 | чисто |
| `frontend/src/modules/calls/call-calendar.tsx` | 530 | 0 | FIXED-OK |
| `frontend/src/modules/calls/create-call-dialog.tsx` | 573 | 0 | FIXED-OK |
| `frontend/src/modules/calls/recording-controls.tsx` | 130 | 0 | FIXED-OK |
| `frontend/src/modules/calls/recordings-list.tsx` | 221 | 1 | FIXED-OK + NEW tabnabbing |
| `frontend/src/modules/calls/scheduled-calls.tsx` | 317 | 0 | FIXED-OK |
| `frontend/src/modules/crm/crm-page.tsx` | 393 | 0 | FIXED-OK |
| `frontend/src/modules/crm/deals-pipeline.tsx` | 150 | 0 | чисто |
| `frontend/src/modules/hr/leave-requests.tsx` | 280 | 0 | FIXED-OK |
| `frontend/src/modules/intranet/intranet-hub.tsx` | 67 | 0 | FIXED-OK |
| `frontend/src/modules/intranet/kudos-feed.tsx` | 204 | 0 | FIXED-OK |
| `frontend/src/modules/intranet/news-feed.tsx` | 178 | 0 | чисто |
| `frontend/src/modules/intranet/org-directory.tsx` | 267 | 0 | чисто |
| `frontend/src/modules/intranet/random-coffee.tsx` | 130 | 0 | FIXED-OK |
| `frontend/src/modules/marketplace/marketplace-page.tsx` | 501 | 1 | FIXED-OK browse/mine + NEW pending |
| `frontend/src/modules/messenger/link-preview.tsx` | 238 | 0 | чисто |
| `frontend/src/modules/messenger/message-bubble.tsx` | 224 | 0 | FIXED-OK |
| `frontend/src/modules/settings/ai-privacy.tsx` | 109 | 0 | FIXED-OK |
| `frontend/src/modules/settings/api-keys.tsx` | 233 | 0 | FIXED-OK |
| `frontend/src/modules/settings/appearance.tsx` | 139 | 0 | чисто |
| `frontend/src/modules/settings/project-templates.tsx` | 186 | 1 | page FIXED-OK; TemplateSelector RESIDUAL |

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W10-1-001 | RESIDUAL | P3 | `accounting-page.tsx:48–61` (+ UI `:196–417`) | `trialBalance` / `invoices` / `incomeStatement` без `isError`; early-return только на accounts+journal | Частичный фикс A2W10-A1-001 / W10-A1-005: fail вторичных API = пустые таблицы/скрытие P&L как empty ledger | `isError` на трёх query + banner/retry per section (или общий `accountingLoadError`) |
| A3W10-1-002 | RESIDUAL | P3 | `project-templates.tsx:27–36,54–83` | `TemplateSelector`: есть `.catch`+toast, нет loading/error UI — fail → пустая сетка | Residual W10-A1-012: toast не заменяет empty≠error при создании проекта | `loadError` + `QueryErrorAlert` / skeleton; не рендерить empty grid как «нет шаблонов» |
| A3W10-1-003 | NEW | P3 | `recordings-list.tsx:65` | `window.open(data.url, "_blank")` без `noopener`/`noreferrer` | Reverse tabnabbing на presigned download (Chromium часто noopener по умолчанию; Safari/legacy — нет). `LinkPreview` уже ставит `rel` | `window.open(url, "_blank", "noopener,noreferrer")` или `<a download target=_blank rel=noopener>` |
| A3W10-1-004 | NEW | P3 | `marketplace-page.tsx:67–71,236` | Admin `listPendingListings` без `isError`; секция только при `pendingItems.length > 0` | Fail mod-queue = секция исчезает (как «нет pending»), в отличие от browse/mine с `listError` | `isError` + alert/retry в admin block; не скрывать очередь при fail |
| A3W10-1-005 | NEW | P3 | `bases-page.tsx:52–72,267–274` | `getBase` / `getTableView` без `isError`/`isLoading` в main | При выбранной базе и fail/loading main пустой (ни «выберите», ни alert) — soft error≠empty | `isError`/`isPending` на detail/table + `QueryErrorAlert` в `<main>` |

## FIXED-OK (сверка A2 + Stage A)

| Зона | Проверено в HEAD |
|------|------------------|
| CreateCall projectId | Select + `isValid` требует `projectId`; `onError` toast (`create-call-dialog.tsx:200–237`) |
| ScheduledCalls Query | `useQuery` `["calls","scheduled",channelId]` + isError UI; invalidate после create |
| call-calendar | `scheduledLoading` / `scheduledError` |
| bases list | `basesError` + `QueryErrorAlert` (W10-A1-006) |
| api-keys timer | `copyTimerRef` + unmount clearTimeout |
| recording-controls | `toastApiError` start/stop |
| intranet-hub URL tab | controlled `useSearch` + `navigate({ search })` |
| leave / kudos | `isError` ≠ empty |
| coffee / marketplace mutations | `onError` toast; browse/mine `isError` |
| message-bubble URLs | `extractUrls(displayText)` ∪ `contentText` |
| XSS / router | нет raw HTML; только TanStack |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 5 |
| **Total** | **5** |

| Метрика | Значение |
|---------|----------|
| Файлов | **29 / 29 / 0** |
| Активных | **5** (NEW 3 · REGRESS 0 · RESIDUAL 2) |
| FIXED-OK (A2 actives 001–007) | **7/7** ядро (2 partial → P3 residual) |
| FIXED-OK (Stage A W10-A1 via A2) | **14+** без регресса P1/P2 |

### Топ-3

1. **A3W10-1-001 (P3, RESIDUAL)** — accounting: invoices/trial/P&L всё ещё error≠empty.
2. **A3W10-1-004 (P3, NEW)** — marketplace pending queue silently disappears on fail.
3. **A3W10-1-003 (P3, NEW)** — recordings `window.open` без noopener.

---

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=0 · P3=5 · Total=5.  
Топ: A3W10-1-001 · A3W10-1-004 · A3W10-1-003.
