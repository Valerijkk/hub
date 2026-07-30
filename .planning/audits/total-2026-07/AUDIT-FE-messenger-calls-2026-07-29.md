# AUDIT FE — messenger + calls modules

> TASK-008 · File-auditor · READ-ONLY product · 2026-07-29  
> Ветка: `cursor-grok-4.5` · HEAD: `617af223`  
> Scope: `frontend/src/modules/messenger/**/*.{ts,tsx}` · `frontend/src/modules/calls/**/*.{ts,tsx}`  
> Фокус: XSS · IDOR UI · error≠empty · tokens vs hardcoded · noopener · a11y critical  
> Выход: **только этот файл** · коммит **не** делать · **TASK-008 не закрывать**

---

## Охват

| Каталог | Найдено файлов | Прочитано | Пустых / пропущено |
|---------|---------------:|----------:|-------------------:|
| `modules/messenger` | 2 | 2 | 0 |
| `modules/calls` | 10 | 10 | 0 |
| **Σ** | **12** | **12** | **0** |

Примечание: основная UI-логика мессенджера живёт вне `modules/messenger/` (routes/components/stores); в scope слота — только эти 2 файла.

---

## Таблица файлов

| file | status | severity | note |
|------|--------|----------|------|
| `frontend/src/modules/messenger/message-bubble.tsx` | finding | P3 | Текст через React children (XSS нет). Цвета `bg-blue-600` / `gray-*` вместо `--hub-*` |
| `frontend/src/modules/messenger/link-preview.tsx` | clean | — | `rel="noopener noreferrer"`; `error→null` осознанно для MSG-11 unfurl; XSS нет |
| `frontend/src/modules/calls/calendar-types.ts` | clean | — | Только типы/утилиты RRULE; UI/секретов нет |
| `frontend/src/modules/calls/calendar-agenda-view.tsx` | clean | — | Empty = нет событий; error обрабатывает parent; a11y кнопок ок |
| `frontend/src/modules/calls/calendar-day-view.tsx` | clean | — | aria-label на слотах/событиях; XSS нет |
| `frontend/src/modules/calls/calendar-month-view.tsx` | clean | — | aria-label; stable keys; XSS нет |
| `frontend/src/modules/calls/calendar-week-view.tsx` | finding | P3 | aria-label слота без дня недели (только час) — путаница в сетке 7 дней |
| `frontend/src/modules/calls/call-calendar.tsx` | finding | P2 | `scheduled*` error≠empty ок; **каналы** без `isError` → fail = пустой календарь; close `×` без accessible name |
| `frontend/src/modules/calls/create-call-dialog.tsx` | finding | P2 | `projects`/`channels` без `isError` → fail = «нет проектов/каналов»; Channel Select без aria-label; remove-X без имени |
| `frontend/src/modules/calls/recording-controls.tsx` | clean | — | toastApiError; aria-label start/stop; authz на API |
| `frontend/src/modules/calls/recordings-list.tsx` | clean | — | `loadError`+`QueryErrorAlert`; `window.open(..., "noopener,noreferrer")` |
| `frontend/src/modules/calls/scheduled-calls.tsx` | clean | — | `isError` ≠ empty; delete/create toastApiError; custom RRULE = текст, не HTML |

---

## Findings (активные)

| ID | Sev | file:line | Категория | Проблема |
|----|-----|-----------|-----------|----------|
| FE-MC-001 | P2 | `call-calendar.tsx:144–147`, `:154–164`, `:311–330` | error≠empty | `useQuery(getChannels)` без `isError`/`refetch`. При fail `channels` → `channelIds=[]` → `enabled: false` у scheduled → UI без alert (только пустая сетка). `scheduledError` срабатывает лишь при `channelIds.length > 0` |
| FE-MC-002 | P2 | `create-call-dialog.tsx:106–114`, `:281–327` | error≠empty | `projects` / `channels` без `isError`. Fail `listProjects` → `projectList.length===0` → «Нет доступных проектов». Fail `getChannels` → пустой Select «Выберите канал». Неотличимо от легитимного empty |
| FE-MC-003 | P2 | `call-calendar.tsx:471–477` | a11y critical | Кнопка закрытия popover (`&times;`) без `aria-label` / текста — нет accessible name (WCAG 4.1.2). Backdrop-кнопка подписана (`:460`), сама `×` — нет |
| FE-MC-004 | P3 | `create-call-dialog.tsx:316–318` | a11y | `SelectTrigger` канала без `aria-label` (у проекта есть `:293`) |
| FE-MC-005 | P3 | `create-call-dialog.tsx:361–367` | a11y | Кнопка удаления участника (`<X />`) без `aria-label` |
| FE-MC-006 | P3 | `calendar-week-view.tsx:104–106` | a11y | `aria-label={`Создать встречу на HH:00`}` без даты/дня — 7 одинаковых подписей в строке |
| FE-MC-007 | P3 | `message-bubble.tsx:169–174`, `:146` | tokens | Hardcoded `bg-blue-600` / `bg-gray-*` / `dark:bg-gray-*` вместо токенов `--hub-*` / semantic surfaces |

---

## Фокус-чеклист

| Тема | Вердикт |
|------|---------|
| **XSS** | В 12 файлах нет `dangerouslySetInnerHTML` / `innerHTML` / `eval`. User/OG текст через React children. Plate → plain text. |
| **IDOR UI** | Списки/мутации опираются на API (`projectId`/`channelId`/`callId`/`recordingId` из props или Query). Клиентской подмены чужого ресурса без ответа API нет. Пара channel↔project в CreateCallDialog не валидируется на FE (ожидаемо BE). |
| **error≠empty** | Хорошо: `call-calendar` scheduled, `recordings-list`, `scheduled-calls`. Дыры: каналы в календаре (FE-MC-001), projects/channels в диалоге (FE-MC-002). `link-preview` `error→null` — by design unfurl. |
| **tokens vs hardcoded** | Смешанно: agenda/month ближе к semantic (`bg-card`, `text-t2`); bubble/day/week/recording — raw Tailwind color utilities (FE-MC-007 + аналоги без эскалации). |
| **noopener** | `link-preview.tsx:228` — `rel="noopener noreferrer"`. `recordings-list.tsx:65` — `noopener,noreferrer` в `window.open` (A3W10-1-003 закрыт на HEAD). |
| **a11y critical** | FE-MC-003 (безымянный close). Остальное — P3 labels. Критичных keyboard-trap / невидимых focus-ловушек нет. |

---

## Сводка

| Severity | Count |
|----------|------:|
| P0 / P1 | 0 |
| P2 | 3 |
| P3 | 4 |
| Файлов finding | 4 |
| Файлов clean | 8 |

**Не коммитить. TASK-008 не закрывать.**
