# A2 W9 агент 2 — frontend/components part 2

> TASK-008 · **Стадия A2** · read-only · 2026-07-28  
> Зеркало: `W9-ASSIGNMENTS.md` §A2 · Stage A: `A9-2-components.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `frontend/src/components/dashboard/widgets/burndown-chart.tsx` | 109 | 0 | чисто (A fixed) |
| `frontend/src/components/dashboard/widgets/deadlines.tsx` | 89 | 2 | замечания |
| `frontend/src/components/dashboard/widgets/my-tasks.tsx` | 93 | 2 | замечания |
| `frontend/src/components/dashboard/widgets/storage-usage.tsx` | 49 | 0 | чисто |
| `frontend/src/components/dashboard/widgets/system-health.tsx` | 102 | 0 | чисто |
| `frontend/src/components/dashboard/widgets/task-status-chart.tsx` | 109 | 0 | чисто (A fixed) |
| `frontend/src/components/dashboard/widgets/team-activity.tsx` | 58 | 0 | чисто |
| `frontend/src/components/dashboard/widgets/test-trends.tsx` | 111 | 0 | чисто (A fixed) |
| `frontend/src/components/dashboard/widgets/unread-messages.tsx` | 44 | 0 | чисто |
| `frontend/src/components/dashboard/widgets/upcoming-calls.tsx` | 87 | 0 | чисто (A fixed) |
| `frontend/src/components/dashboard/widgets/velocity-chart.tsx` | 107 | 0 | чисто (A fixed) |
| `frontend/src/components/files/dropzone.tsx` | 78 | 0 | чисто (A fixed) |
| `frontend/src/components/files/file-upload.tsx` | 70 | 0 | чисто (A fixed) |
| `frontend/src/components/help/context-help-tip.tsx` | 47 | 0 | чисто |
| `frontend/src/components/help/shortcuts-dialog.tsx` | 59 | 0 | чисто |
| `frontend/src/components/language-switcher.tsx` | 86 | 0 | чисто (A fixed) |
| `frontend/src/components/layout/app-shell.tsx` | 187 | 0 | чисто (A fixed) |
| `frontend/src/components/layout/sidebar.tsx` | 148 | 0 | чисто (A fixed) |
| `frontend/src/components/layout/sidebar-link.tsx` | 125 | 0 | чисто (A fixed) |
| `frontend/src/components/layout/sidebar-nav.tsx` | 341 | 0 | чисто |
| `frontend/src/components/layout/sidebar-projects.tsx` | 305 | 0 | чисто (A fixed) |
| `frontend/src/components/layout/sidebar-user.tsx` | 104 | 0 | чисто (A fixed) |
| `frontend/src/components/layout/skip-link.tsx` | 13 | 0 | чисто |
| `frontend/src/components/layout/topbar.tsx` | 86 | 0 | чисто |
| `frontend/src/components/messenger/bookmarks-panel.tsx` | 112 | 1 | замечания |
| `frontend/src/components/messenger/bot-markdown.tsx` | 173 | 0 | чисто (A fixed) |

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.

## Перепроверка Stage A (`W9-A2-*`)

| Старый ID | Tag | Комментарий |
|-----------|-----|-------------|
| W9-A2-001 | **FIXED-OK** | `joinCall(call.id)` (`upcoming-calls.tsx:84`) |
| W9-A2-002 | **FIXED-OK** | collapsed: button expand без вложенного Link |
| W9-A2-003 | **FIXED-OK** | collapsed `aria-label` (+ unread count) |
| W9-A2-004 | **FIXED-OK** | createProject Link `aria-label` |
| W9-A2-005 | **FIXED-OK** | unread в `aria-label` collapsed |
| W9-A2-006 | **FIXED-OK** | убран claim «Онлайн»; muted presence dot |
| W9-A2-007 | **FIXED-OK** | `userCollapsedIds` позволяет свернуть active |
| W9-A2-008 | **FIXED-OK** | mobile menu close on route + Escape |
| W9-A2-009 | **FIXED-OK** | bookmarks `QueryStateView` + error |
| W9-A2-010 | **RESIDUAL** | UI всё ещё показывает сырые UUID (U03 / P2-FE-U03) |
| W9-A2-011 | **FIXED-OK** | dropzone `role`/`aria-label`/`aria-live` |
| W9-A2-012 | **FIXED-OK** | file input `aria-label` |
| W9-A2-013 | **FIXED-OK** | language `aria-pressed` + group label |
| W9-A2-014 | **FIXED-OK** | charts `role="img"` + `aria-label` summary |
| W9-A2-015 | **FIXED-OK** | priority `role="img"` + `aria-label` |
| W9-A2-016 | **FIXED-OK** | `DashboardWidgetData` + type guards (`widget-data.ts`) |
| W9-A2-017 | **FIXED-OK** | bot-markdown copy timer cleanup |
| W9-A2-018 | **FIXED-OK** | bookmarks delete `toastApiError` |
| W9-A2-019 | **RESIDUAL** | deadlines rows всё ещё без Link (TASK-075 false-close) |

## Находки (open)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A2W9-2-001 | RESIDUAL | P2 | `bookmarks-panel.tsx:78-80` | Закладки рендерят `channel_id` / `message_id` UUID | Preview так и не появился (W9-A2-010 / P2-FE-U03); список бесполезен | BE preview DTO **или** FE fetch message/channel |
| A2W9-2-002 | RESIDUAL | P3 | `deadlines.tsx:66-91` | Строки дедлайнов не кликабельны | W9-A2-019; TASK-075 «Исправлена», код без Link | `Link` на issue по `key` (+ `project_id` в DTO при необходимости) |
| A2W9-2-003 | NEW | P3 | `my-tasks.tsx:73-97` | Строки «Мои задачи» тоже без Link (тот же паттерн, что deadlines) | Виджет информативный, сценарий «открыть задачу» отсутствует | `Link` / navigate по `task.key` |
| A2W9-2-004 | NEW | P3 | `deadlines.tsx:32-39,72`; `my-tasks.tsx:13-18` | Raw Tailwind `red-500`/`amber-500`/`orange-500`/… | Обход `--hub` / semantic tokens; drift light/dark | `var(--red)` / `var(--amber)` / map токенов |

## Чисто

layout a11y (nested interactive, collapsed names, mobile Escape), upcoming-calls join, charts text alternative, files dropzone/upload labels, language switcher, bookmarks error≠empty (кроме UUID preview), bot-markdown timer, typed widget data.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **26 / 26 / 0** |
| Stage A перепроверено | 19 |
| FIXED-OK | **17** |
| RESIDUAL | 2 |
| REGRESS | 0 |
| NEW | 2 |
| Open Σ | **4** (P2:1 · P3:3) |
