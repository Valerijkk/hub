# A3 W9 агент 3 — frontend/components part 3 (messenger) FINAL

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W9-3-components.md](A2W9-3-components.md) · Stage A: [A9-3-components.md](A9-3-components.md)  
> Assignment: [W9-ASSIGNMENTS.md](W9-ASSIGNMENTS.md) § A3  
> Фокус: XSS · a11y · hooks · design tokens  
> Периметр: 26 файлов · Прочитано: **26/26** · Пропущено: **0**

## Сверка с A2 (A2W9-3-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W9-3-001 | RESIDUAL | **RESIDUAL** | `channel-list.tsx:188-211` — collapse `<button>` без `aria-expanded` / `aria-controls` (W9-A3-013; TASK-049 false-close) |
| A2W9-3-002 | RESIDUAL | **RESIDUAL** | `pinned-messages.tsx:80` — `ID: {pin.message_id}` (W9-A3-036 / P3-FE-U01) |

### Cascade FIXED-OK (Stage A W9-A3-* / A2)

| Группа | Tag | Evidence |
|--------|-----|----------|
| W9-A3-001 | FIXED-OK | `thread-panel.tsx:52-68` — parent из cache + `getMessages(limit:50)` fallback |
| W9-A3-002 | FIXED-OK | `message-composer.tsx:232,329` — `effectiveThreadId = threadId ?? replyToMessage?.id` |
| W9-A3-003 | FIXED-OK | `voice-recorder.tsx:86-95` — `onstop` до `stop()` |
| W9-A3-004 | FIXED-OK | `message-item.tsx:555` — `group-focus-within` / `focus-within` toolbar |
| W9-A3-005 | FIXED-OK | `sticker-picker.tsx:202-209` — `useState` до throw |
| W9-A3-006 | FIXED-OK | `gif-picker.tsx:149-155` — debounce cleanup on unmount |
| W9-A3-007…012,018,019,024…030,038,039 | FIXED-OK | aria-labels / Label htmlFor / focus edit |
| W9-A3-014…017,031…035 | FIXED-OK | semantic `var(--*)` вместо raw Tailwind palette (grep `bg-green-500` / `text-orange-500` / … → 0) |
| W9-A3-020…022 | FIXED-OK | PlateDoc / narrowed casts |
| W9-A3-023 | FIXED-OK | mention Esc → `suppressed` (`mention-autocomplete.tsx:92,113-115,175-178`) |
| W9-A3-037 | FIXED-OK | poll `channelId` WS-scope + `toastApiError` |

**REGRESS:** 0 · **RESIDUAL (A2 actives):** 2 · **NEW (этот слот):** 2

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `frontend/src/components/messenger/channel-drawer-context.tsx` | 29 | 0 | чисто |
| `frontend/src/components/messenger/channel-icon.tsx` | 66 | 0 | чисто |
| `frontend/src/components/messenger/channel-list.tsx` | 851 | 2 | RESIDUAL a11y + NEW tokens |
| `frontend/src/components/messenger/chat-view.tsx` | 457 | 1 | NEW tokens (rgba header) |
| `frontend/src/components/messenger/create-channel-dialog.tsx` | 106 | 0 | FIXED-OK Label |
| `frontend/src/components/messenger/create-poll-dialog.tsx` | 251 | 0 | FIXED-OK aria |
| `frontend/src/components/messenger/forward-dialog.tsx` | 94 | 0 | чисто |
| `frontend/src/components/messenger/gif-picker.tsx` | 401 | 0 | FIXED-OK debounce + aria |
| `frontend/src/components/messenger/invite-link-dialog.tsx` | 467 | 0 | FIXED-OK tokens/aria |
| `frontend/src/components/messenger/link-preview.tsx` | 205 | 1 | NEW XSS (href без sanitize) |
| `frontend/src/components/messenger/mention-autocomplete.tsx` | 248 | 0 | FIXED-OK Esc suppress |
| `frontend/src/components/messenger/message-actions.tsx` | 226 | 0 | чисто (эталон aria) |
| `frontend/src/components/messenger/message-attachments.tsx` | 82 | 0 | чисто* |
| `frontend/src/components/messenger/message-composer.tsx` | 825 | 1 | FIXED-OK reply; NEW tokens (rgba focus) |
| `frontend/src/components/messenger/message-item.tsx` | 587 | 1 | NEW XSS (Plate `a.url` → LinkPreview) |
| `frontend/src/components/messenger/new-dm-dialog.tsx` | 174 | 0 | FIXED-OK aria |
| `frontend/src/components/messenger/pinned-messages.tsx` | 106 | 1 | RESIDUAL U01 |
| `frontend/src/components/messenger/poll-message.tsx` | 291 | 0 | FIXED-OK WS+toast |
| `frontend/src/components/messenger/presence-indicator.tsx` | 60 | 0 | FIXED-OK tokens + unavailable |
| `frontend/src/components/messenger/reaction-picker.tsx` | 103 | 0 | FIXED-OK aria |
| `frontend/src/components/messenger/render-content.tsx` | 196 | 0 | FIXED-OK `sanitizeUrl` whitelist |
| `frontend/src/components/messenger/search-dialog.tsx` | 176 | 0 | FIXED-OK aria |
| `frontend/src/components/messenger/sticker-picker.tsx` | 321 | 0 | FIXED-OK hooks/tabs/aria |
| `frontend/src/components/messenger/thread-issue-links.tsx` | 114 | 0 | FIXED-OK `--red` |
| `frontend/src/components/messenger/thread-panel.tsx` | 150 | 0 | FIXED-OK parent cache |
| `frontend/src/components/messenger/voice-recorder.tsx` | 205 | 0 | FIXED-OK onstop + `--red` |

Файлов в списке: **26**. Прочитано: **26**. Пропущено: **0**.  
\* `message-attachments` — `href={att.url}` без FE-sanitize; URL с storage/API (доверие BE). Не заводим отдельный ID; см. remediation рядом с A3W9-3-003.

## Находки (активные A3)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A3W9-3-001 | RESIDUAL (A2W9-3-001 / W9-A3-013) | P2 | `channel-list.tsx:188-211` | Section collapse: нет `aria-expanded` / `aria-controls` | AT не знает свёрнутость; TASK-049 заявлен закрытым, атрибут отсутствует | `aria-expanded={!collapsed}` + `id` на панели + `aria-controls` |
| A3W9-3-002 | RESIDUAL (A2W9-3-002 / W9-A3-036) | P3 | `pinned-messages.tsx:80` | Pins: `ID: {message_id}` | P3-FE-U01 — UUID вместо preview текста | BE `content_text`/preview в pins DTO **или** product accept ids |
| A3W9-3-003 | NEW | P2 | `link-preview.tsx:32,78,140` + `message-item.tsx:204-206,476-478` | `LinkPreview` ставит `href={url}` без whitelist; Plate `a.url` попадает в `extractedUrls` без фильтра схемы | `RenderContent` уже режет `javascript:`/`data:` (`render-content.tsx:7-14`), но карточка превью — обход: клик по preview → XSS | Общий `sanitizeUrl` (как в render-content) в `LinkPreview` **и** фильтр схем при extract в `message-item`; unsafe → не рендерить preview |
| A3W9-3-004 | NEW | P3 | `channel-list.tsx:556,799` · `chat-view.tsx:223,236,264` · `message-composer.tsx:591,788` · `message-item.tsx:496` | Hardcoded `rgba(52,211,153,…)` / `rgba(13,14,18,.7)` вместо токенов | После sweep raw Tailwind (W9-A3-014…) остался brand-green / void в литералах — drift light/dark | `var(--green)` / `var(--gG)` / `var(--void)` (+ opacity modifiers) |

## Чисто (перепроверено)

- **XSS (основной путь):** `render-content` — `SAFE_URL_RE` + `sanitizeUrl`; нет `dangerouslySetInnerHTML` в слоте; text children через React escape / `safeString`.
- **Hooks:** sticker RoH, gif debounce cleanup, channel-list `useMemo` до early return, voice onstop order, virtualizer isolation — OK.
- **a11y (кроме RESIDUAL):** icon triggers, mention listbox, sticker tablist, toolbar focus-within, poll/create labels — OK.
- **Tokens (Stage A palette):** raw `bg-green-500` / `text-orange-500` / `text-purple-500` / `bg-red-500` — отсутствуют; presence/invite/poll/voice/thread на `var(--*)`.
- **Correctness P1:** reply `thread_id`, thread parent cache, voice recorder — без регресса.

Эталоны: `message-actions` aria-labels; `mention-autocomplete` listbox+Esc; `render-content` URL whitelist (эталон для A3W9-3-003).

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **26 / 26 / 0** |
| A2 open перепроверено | 2 |
| Stage A FIXED-OK cascade | 37 (без регресса) |
| FIXED-OK (этот слот, A2→закрыто) | **0** (оба A2 open живы) |
| RESIDUAL | **2** |
| REGRESS | **0** |
| NEW | **2** |
| Open Σ | **4** (P0:0 · P1:0 · P2:2 · P3:2) |
