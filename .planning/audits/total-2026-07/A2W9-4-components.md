# A2 W9 агент 4 — frontend/components part 4

> TASK-008 · **Стадия A2** · read-only · 2026-07-28  
> Зеркало: `W9-ASSIGNMENTS.md` §A4 · Stage A: `A9-4-components.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `frontend/src/components/mobile-layout.tsx` | 59 | 0 | чисто |
| `frontend/src/components/mobile-nav.tsx` | 227 | 0 | чисто (A fixed) |
| `frontend/src/components/notifications/notification-bell.tsx` | 57 | 0 | чисто (A fixed) |
| `frontend/src/components/notifications/notification-panel.tsx` | 307 | 2 | замечания |
| `frontend/src/components/notifications/notification-settings.tsx` | 255 | 0 | чисто (A fixed) |
| `frontend/src/components/notifications/push-permission-banner.tsx` | 81 | 0 | чисто |
| `frontend/src/components/notifications/telegram-link-card.tsx` | 182 | 0 | чисто (A fixed) |
| `frontend/src/components/query-state-view.tsx` | 156 | 0 | чисто |
| `frontend/src/components/route-error-boundary.tsx` | 97 | 0 | чисто (A fixed) |
| `frontend/src/components/search/command-palette.tsx` | 262 | 0 | чисто |
| `frontend/src/components/search/jql-editor.tsx` | 433 | 0 | чисто (A fixed) |
| `frontend/src/components/search/jql-search-page.tsx` | 403 | 0 | чисто |
| `frontend/src/components/search/saved-filters.tsx` | 175 | 1 | замечания |
| `frontend/src/components/theme-provider.tsx` | 97 | 0 | чисто |
| `frontend/src/components/tracker/activity-feed.tsx` | 196 | 0 | чисто (A fixed) |
| `frontend/src/components/tracker/automation-rules-panel.tsx` | 226 | 0 | чисто |
| `frontend/src/components/tracker/board-column.tsx` | 161 | 0 | чисто |
| `frontend/src/components/tracker/checklist.tsx` | 152 | 0 | чисто (A fixed) |
| `frontend/src/components/tracker/comment-list.tsx` | 287 | 0 | чисто |
| `frontend/src/components/tracker/issue-attachments.tsx` | 186 | 0 | чисто (A fixed) |
| `frontend/src/components/tracker/issue-card.tsx` | 210 | 0 | чисто |
| `frontend/src/components/tracker/issue-form.tsx` | 314 | 0 | чисто |
| `frontend/src/components/tracker/issue-form-schema.ts` | 32 | 0 | чисто |
| `frontend/src/components/tracker/plate-editor.tsx` | 51 | 0 | чисто |
| `frontend/src/components/tracker/plate-editor-impl.tsx` | 63 | 0 | чисто (A fixed) |
| `frontend/src/components/tracker/time-tracker.tsx` | 242 | 0 | чисто |

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.

## Перепроверка Stage A (`W9-A4-*`)

| Старый ID | Tag | Комментарий |
|-----------|-----|-------------|
| W9-A4-001 | **FIXED-OK** | NotificationItem: sibling buttons в `div` (не nested) |
| W9-A4-002 | **FIXED-OK** | activity-feed `isError` + retry |
| W9-A4-003 | **FIXED-OK** | jql-editor combobox / listbox / activedescendant |
| W9-A4-004 | **RESIDUAL** | saved-filters: 4× пустой `catch {}` (явный TODO в TASK-047) |
| W9-A4-005 | **FIXED-OK** | plate `editor.tf.setValue` sync |
| W9-A4-006 | **FIXED-OK** | telegram `loadError` ≠ unlinked |
| W9-A4-007 | **FIXED-OK** | bell aria-label с unread count |
| W9-A4-008 | **FIXED-OK** | Share/Trash `aria-label` |
| W9-A4-009 | **FIXED-OK** | mobile-nav `aria-expanded` / `aria-controls` |
| W9-A4-010 | **FIXED-OK** | quiet hours `Label htmlFor` |
| W9-A4-011 | **FIXED-OK** | checklist delete `aria-label` |
| W9-A4-012 | **FIXED-OK** | empty copy «Нажмите Добавить файл» (без fake DnD) |
| W9-A4-013 | **FIXED-OK** | route-error `role="alert"` |
| W9-A4-014 | **FIXED-OK** | узкие DTO casts / без `as unknown as Channel[]` в зоне |

## Находки (open)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A2W9-4-001 | RESIDUAL | P2 | `saved-filters.tsx:53-55,75-77,87-89,99-101` | load/create/delete/share — пустой `catch` | W9-A4-004; TASK-047 TODO remaining — fail = silent no-op | `toastApiError` + loadError state |
| A2W9-4-002 | NEW | P2 | `notification-panel.tsx:318-326` | Delete icon: `hidden group-hover:flex` без `group-focus-within` | Клавиатура/SR: кнопка «Удалить» не видна и плохо достижима (паттерн как бывший W9-A3-004) | `group-focus-within:flex` / всегда показывать на focus |
| A2W9-4-003 | NEW | P3 | `notification-panel.tsx:278,302` (+ TYPE_* blue/purple classes) | Raw `border-blue-500` / `bg-blue-500` / `text-*-400` | Token drift после sweep FE-05 в messenger | `var(--blue)` / semantic unread tokens |

## Чисто

nested interactive notifications (структура) · activity error≠empty · JQL combobox · Plate sync · telegram error · bell unread name · quiet hours labels · attachments empty copy · route EB alert · hooks/effects cleanup · `any`/`ts-ignore` = 0.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **26 / 26 / 0** |
| Stage A перепроверено | 14 |
| FIXED-OK | **13** |
| RESIDUAL | 1 |
| REGRESS | 0 |
| NEW | 2 |
| Open Σ | **3** (P2:2 · P3:1) |
