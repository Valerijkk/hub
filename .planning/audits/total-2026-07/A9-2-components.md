# W9 агент A2 — components part 2

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус FE: hooks (Rules of Hooks / deps / cleanup), a11y, `any`/`unknown` casts, Shadcn className-only, state bugs.  
> Список: W9-ASSIGNMENTS.md §A2 (26 файлов).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/components/dashboard/widgets/burndown-chart.tsx` | 106 | 2 | замечания |
| `frontend/src/components/dashboard/widgets/deadlines.tsx` | 92 | 2 | замечания |
| `frontend/src/components/dashboard/widgets/my-tasks.tsx` | 95 | 2 | замечания |
| `frontend/src/components/dashboard/widgets/storage-usage.tsx` | 50 | 1 | замечания |
| `frontend/src/components/dashboard/widgets/system-health.tsx` | 108 | 1 | замечания |
| `frontend/src/components/dashboard/widgets/task-status-chart.tsx` | 107 | 2 | замечания |
| `frontend/src/components/dashboard/widgets/team-activity.tsx` | 58 | 1 | замечания |
| `frontend/src/components/dashboard/widgets/test-trends.tsx` | 108 | 2 | замечания |
| `frontend/src/components/dashboard/widgets/unread-messages.tsx` | 45 | 1 | замечания |
| `frontend/src/components/dashboard/widgets/upcoming-calls.tsx` | 92 | 2 | замечания |
| `frontend/src/components/dashboard/widgets/velocity-chart.tsx` | 104 | 2 | замечания |
| `frontend/src/components/files/dropzone.tsx` | 78 | 1 | замечания |
| `frontend/src/components/files/file-upload.tsx` | 71 | 1 | замечания |
| `frontend/src/components/help/context-help-tip.tsx` | 50 | 0 | чисто |
| `frontend/src/components/help/shortcuts-dialog.tsx` | 64 | 0 | чисто |
| `frontend/src/components/language-switcher.tsx` | 85 | 1 | замечания |
| `frontend/src/components/layout/app-shell.tsx` | 189 | 1 | замечания |
| `frontend/src/components/layout/sidebar.tsx` | 153 | 1 | замечания |
| `frontend/src/components/layout/sidebar-link.tsx` | 119 | 2 | замечания |
| `frontend/src/components/layout/sidebar-nav.tsx` | 342 | 0 | чисто |
| `frontend/src/components/layout/sidebar-projects.tsx` | 273 | 1 | замечания |
| `frontend/src/components/layout/sidebar-user.tsx` | 81 | 2 | замечания |
| `frontend/src/components/layout/skip-link.tsx` | 14 | 0 | чисто |
| `frontend/src/components/layout/topbar.tsx` | 93 | 0 | чисто |
| `frontend/src/components/messenger/bookmarks-panel.tsx` | 107 | 3 | замечания |
| `frontend/src/components/messenger/bot-markdown.tsx` | 176 | 1 | замечания |

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W9-A2-001 | P1 | `upcoming-calls.tsx:79-81` | «Войти» вызывает `startCall(call.channel_id)` только при truthy `channel_id`; бэкенд `dashboard/service.go` **не заполняет** `channel_id` в `CallSummary` (только `id` / title / channel_name) | Кнопка **всегда no-op** на реальных данных. Даже при `channel_id` `startCall` **создаёт** новый звонок, а не `joinCall(call.id)` — расходится с `CallButton` | `void joinCall(call.id)`; disable/toast при ошибке; синхронизировать DTO при необходимости |
| W9-A2-002 | P1 | `sidebar-user.tsx:46-55` + `:27-40` | В collapsed-режиме `<button onClick={toggle}>` оборачивает `avatarBlock`, который сам — `<Link to="/settings/profile">` | Вложенный interactive (button ⊃ link): невалидный HTML, непредсказуемый фокус/клик/AT (WCAG 4.1.1 / nested controls) | Либо только Link (клик → профиль), либо только button (toggle) + отдельная ссылка; не вкладывать |
| W9-A2-003 | P1 | `sidebar-link.tsx:59-104`; `:98` | При `collapsed` у `Link` нет текста и **нет `aria-label`**; иконки в nav с `aria-hidden` | Свёрнутый сайдбар: пункты без accessible name (screen reader / voice control) | `aria-label={label}` на `Link`; `aria-current="page"` при active |
| W9-A2-004 | P2 | `sidebar.tsx:58-74` | Icon-only compose `Link` (`to="/projects/new"`): SVG `aria-hidden`, на Link нет `aria-label` (есть только Tooltip) | Триггер без accessible name; Tooltip ≠ имя для всех AT | `aria-label={t("sidebar.createProject")}` на Link |
| W9-A2-005 | P2 | `sidebar-link.tsx:94-96` | Collapsed unread: только красная точка без текста/`aria` | Индикатор непрочитанного недоступен AT (информация только цветом) | `aria-label={`${label}, ${badge} непрочитанных`}` или sr-only счётчик |
| W9-A2-006 | P2 | `sidebar-user.tsx:39`; `:75` | Всегда зелёный индикатор + текст «Онлайн»; нет связи с presence/WS | Ложное состояние online для любого пользователя | Presence store / убрать claim «Онлайн» / «статус неизвестен» |
| W9-A2-007 | P2 | `sidebar-projects.tsx:187-188` | `isExpanded = expandedProjectId === id \|\| isProjectActive` | Активный проект **нельзя свернуть** — `toggleProject` бесполезен, пока route совпадает | Разделить auto-expand on navigate и ручной collapse (`userCollapsed` set) |
| W9-A2-008 | P2 | `app-shell.tsx:60`; `:128-145` | `mobileMenuOpen` не сбрасывается при смене route и нет Escape | Overlay остаётся после навигации; ловушка фокуса/UX на mobile | `useRouterState` → close; `keydown Escape` → close |
| W9-A2-009 | P2 | `bookmarks-panel.tsx:16-19`; `:41-47` | `useQuery` без `isError` / error UI | Сбой API неотличим от пустого списка после loading | Ветка `isError` + retry; toast на remove fail |
| W9-A2-010 | P2 | `bookmarks-panel.tsx:67-70` | UI показывает сырые `channel_id` / `message_id` (UUID) | Закладки бесполезны без превью текста/имени канала (DTO Bookmark только ids) | Расширить API preview **или** fetch message/channel по id в панели |
| W9-A2-011 | P2 | `dropzone.tsx:30-38` | Drop-зона: только drag handlers, нет `role` / `aria-label` / live region для uploading | DnD-зона не объявлена AT; статус «Загрузка...» без `aria-live` | `role="region"` + `aria-label`; `aria-busy` / `aria-live="polite"` при upload |
| W9-A2-012 | P2 | `file-upload.tsx:61-68` | Скрытый `<input type="file">` без `id`/`aria-label`; не связан с Button через `htmlFor` | Связь button→input только через click(); часть AT/тестов не видит control | `id` + `aria-labelledby` / `aria-label="Загрузить файл"`; `tabIndex={-1}` |
| W9-A2-013 | P2 | `language-switcher.tsx:39-62` | Compact RU/EN: нет `aria-pressed` / `role="group"` / `aria-label` | Состояние выбранного языка не экспонируется AT | `aria-pressed={current===…}`; group `aria-label="Язык"` |
| W9-A2-014 | P2 | `burndown-chart.tsx:67-104`; `task-status-chart.tsx:76-105`; `test-trends.tsx:67-106`; `velocity-chart.tsx:67-102` | Recharts без текстовой альтернативы / `aria-label` на контейнере | WCAG 1.1.1: данные только в canvas/SVG без summary | `aria-label` + sr-only table/summary по точкам |
| W9-A2-015 | P2 | `my-tasks.tsx:75-78` | Приоритет — цветная точка + `title`; нет `aria-label` | Color-only + `title` ненадёжен для AT | `role="img"` + `aria-label={PRIORITY_LABELS[…]}` |
| W9-A2-016 | P3 | все 11 widgets `Props.data?: unknown` + `as ProjectStats\|PersonalData\|AdminStats` (напр. `burndown-chart.tsx:17-49`) | Обход типизации через `unknown` cast | Нарушение style (`any`/`unknown` без контракта); runtime mismatch → пустые виджеты без сигнала | Дискриминированный union / generics от `widget-grid` registry |
| W9-A2-017 | P3 | `bot-markdown.tsx:93` | `setTimeout(() => setCopied(false), 2000)` без cleanup | Unmount до 2s → setState на размонтированном + таймер-leak | `useEffect` cleanup / `let cancelled` в `onCopy` |
| W9-A2-018 | P3 | `bookmarks-panel.tsx:21-28` | `handleRemove`: нет try/catch; ошибка delete глотается Promise | UI не обновляется / нет toast при fail delete | try/catch + `toastApiError`; optimistic optional |
| W9-A2-019 | P3 | `deadlines.tsx:62-87` | Строки дедлайнов не кликабельны (нет Link на issue) | Виджет информативный, но сценарий «открыть задачу» отсутствует | `Link` на `/projects/.../issues/...` по `task.id`/`key` |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `context-help-tip.tsx` | `aria-label`, `aria-hidden` на иконке, Tooltip+button, className-only Shadcn |
| `shortcuts-dialog.tsx` | Dialog controlled, i18n groups, безопасный `asLines` для `returnObjects`, без `any` |
| `sidebar-nav.tsx` | Хуки до return, `useMemo` unread, иконки `aria-hidden`, только className на Link-обёртках |
| `skip-link.tsx` | `#main-content`, focus-visible styles, i18n — эталон NFR-16 |
| `topbar.tsx` | `aria-label` / `aria-keyshortcuts` на search, hamburger `aria-label`, Shadcn Button className-only |
| Charts empty states | Empty UI без падений при `[]` |
| `team-activity` / `unread` / `storage` | Стабильные keys / empty; storage только отображение |
| Shadcn usage (Badge/Button/Dialog/Tooltip/ScrollArea) | В этих файлах только `className` / composition — Radix behavior не патчится |
| `app-shell` hotkeys | Cleanup `removeEventListener`; deps `[toggleCommand, toggleAIPanel, navigate]`; `isEditableTarget` guard |
| `bot-markdown` XSS | Нет `dangerouslySetInnerHTML`; текст как React children |

**Фокус-вердикт:** критичный UX — **«Войти» в upcoming-calls мёртв** (нет `channel_id` + неверный API). a11y: **nested button/link** в sidebar-user и **безымянные collapsed links**. Остальное — query error/empty UX, dropzone a11y, chart alternatives, `unknown` widget props.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **26 / 26 / 0** |
| Находок | **19** |
| P0 | 0 |
| P1 | 3 |
| P2 | 12 |
| P3 | 4 |

### Топ-3

1. **W9-A2-001 (P1)** — `upcoming-calls`: «Войти» всегда no-op (`channel_id` не приходит) и при наличии звал бы `startCall` вместо `joinCall(call.id)`.
2. **W9-A2-002 (P1)** — `sidebar-user` collapsed: `<button>` оборачивает `<Link>` (nested interactive).
3. **W9-A2-003 (P1)** — `sidebar-link` collapsed: пункты навигации без accessible name.
