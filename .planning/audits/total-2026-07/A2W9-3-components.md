# A2 W9 агент 3 — frontend/components part 3 (messenger)

> TASK-008 · **Стадия A2** · read-only · 2026-07-28  
> Зеркало: `W9-ASSIGNMENTS.md` §A3 · Stage A: `A9-3-components.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `frontend/src/components/messenger/channel-drawer-context.tsx` | 25 | 0 | чисто |
| `frontend/src/components/messenger/channel-icon.tsx` | 65 | 0 | чисто |
| `frontend/src/components/messenger/channel-list.tsx` | 791 | 1 | замечания |
| `frontend/src/components/messenger/chat-view.tsx` | 434 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/create-channel-dialog.tsx` | 101 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/create-poll-dialog.tsx` | 232 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/forward-dialog.tsx` | 88 | 0 | чисто |
| `frontend/src/components/messenger/gif-picker.tsx` | 358 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/invite-link-dialog.tsx` | 428 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/link-preview.tsx` | 192 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/mention-autocomplete.tsx` | 224 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/message-actions.tsx` | 217 | 0 | чисто |
| `frontend/src/components/messenger/message-attachments.tsx` | 78 | 0 | чисто |
| `frontend/src/components/messenger/message-composer.tsx` | 775 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/message-item.tsx` | 550 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/new-dm-dialog.tsx` | 166 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/pinned-messages.tsx` | 101 | 1 | замечания |
| `frontend/src/components/messenger/poll-message.tsx` | 268 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/presence-indicator.tsx` | 45 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/reaction-picker.tsx` | 99 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/render-content.tsx` | 183 | 0 | чисто (A fixed*) |
| `frontend/src/components/messenger/search-dialog.tsx` | 167 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/sticker-picker.tsx` | 303 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/thread-issue-links.tsx` | 108 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/thread-panel.tsx` | 140 | 0 | чисто (A fixed) |
| `frontend/src/components/messenger/voice-recorder.tsx` | 185 | 0 | чисто (A fixed) |

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.  
\* `render-content` cast сужен до `Record<string, unknown>` после Array.isArray — приемлемо; Stage A P3 закрыт как style.

## Перепроверка Stage A (`W9-A3-*`) — кратко

| Группа | Tag | Комментарий |
|--------|-----|-------------|
| W9-A3-001 | **FIXED-OK** | parent из cache + `getMessages(limit:50)` fallback |
| W9-A3-002 | **FIXED-OK** | `effectiveThreadId = threadId ?? replyToMessage?.id` |
| W9-A3-003 | **FIXED-OK** | `onstop` до `stop()` |
| W9-A3-004…012,018,019,024…030,038,039 | **FIXED-OK** | a11y names / focus-within toolbar / labels |
| W9-A3-005 | **FIXED-OK** | `useState` до throw (hooks order) |
| W9-A3-006 | **FIXED-OK** | gif debounce cleanup |
| W9-A3-013 | **RESIDUAL** | section collapse **без** `aria-expanded` (`channel-list.tsx:188-211`) — false-close TASK-049 |
| W9-A3-014…017,031…035 | **FIXED-OK** | semantic `var(--*)` tokens |
| W9-A3-020…022 | **FIXED-OK** | `PlateDoc` / narrowed casts |
| W9-A3-023 | **FIXED-OK** | mention Esc → `suppressed` |
| W9-A3-036 | **RESIDUAL** | pins: `ID: {message_id}` (U01 / P3-FE-U01) |
| W9-A3-037 | **FIXED-OK** | poll `channelId` + toastApiError |

Полный перечень 39 ID: 37 FIXED-OK · 2 RESIDUAL · 0 REGRESS.

## Находки (open)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A2W9-3-001 | RESIDUAL | P2 | `channel-list.tsx:188-211` | Collapse section: нет `aria-expanded` / `aria-controls` | W9-A3-013; TASK-049 заявлен закрытым, атрибут отсутствует | `aria-expanded={!collapsed}` + `aria-controls` |
| A2W9-3-002 | RESIDUAL | P3 | `pinned-messages.tsx:80` | Список pins показывает `ID: {message_id}` | W9-A3-036 / P3-FE-U01 — нет preview | BE `content_text` **или** product accept ids |

## Чисто

P1 reply/thread/voice — OK. a11y icon triggers, sticker hooks/tabs, gif debounce, tokens presence/link-preview/bot/voice/poll, poll toast, PlateDoc composer.  
Эталоны: `message-actions` aria-labels; `mention-autocomplete` listbox; `sticker-picker` tablist.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **26 / 26 / 0** |
| Stage A перепроверено | 39 |
| FIXED-OK | **37** |
| RESIDUAL | 2 |
| REGRESS | 0 |
| NEW | 0 |
| Open Σ | **2** (P2:1 · P3:1) |
