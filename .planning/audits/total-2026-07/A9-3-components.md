# W9 агент A3 — components part 3 (messenger)

> Фокус: hooks · a11y · `any`/casts · design tokens (`--t*`/`--green`/…; raw Tailwind palette)  
> TASK-008 · Стадия A · READ-ONLY

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| frontend/src/components/messenger/channel-drawer-context.tsx | 29 | 0 | чисто |
| frontend/src/components/messenger/channel-icon.tsx | 66 | 0 | чисто |
| frontend/src/components/messenger/channel-list.tsx | 843 | 3 | замечания |
| frontend/src/components/messenger/chat-view.tsx | 452 | 1 | замечания |
| frontend/src/components/messenger/create-channel-dialog.tsx | 101 | 1 | замечания |
| frontend/src/components/messenger/create-poll-dialog.tsx | 250 | 1 | замечания |
| frontend/src/components/messenger/forward-dialog.tsx | 94 | 0 | чисто |
| frontend/src/components/messenger/gif-picker.tsx | 392 | 2 | замечания |
| frontend/src/components/messenger/invite-link-dialog.tsx | 448 | 3 | замечания |
| frontend/src/components/messenger/link-preview.tsx | 200 | 2 | замечания |
| frontend/src/components/messenger/mention-autocomplete.tsx | 235 | 1 | замечания |
| frontend/src/components/messenger/message-actions.tsx | 226 | 0 | чисто |
| frontend/src/components/messenger/message-attachments.tsx | 82 | 0 | чисто |
| frontend/src/components/messenger/message-composer.tsx | 817 | 4 | замечания |
| frontend/src/components/messenger/message-item.tsx | 581 | 3 | замечания |
| frontend/src/components/messenger/new-dm-dialog.tsx | 173 | 1 | замечания |
| frontend/src/components/messenger/pinned-messages.tsx | 105 | 2 | замечания |
| frontend/src/components/messenger/poll-message.tsx | 281 | 2 | замечания |
| frontend/src/components/messenger/presence-indicator.tsx | 49 | 1 | замечания |
| frontend/src/components/messenger/reaction-picker.tsx | 102 | 1 | замечания |
| frontend/src/components/messenger/render-content.tsx | 200 | 1 | замечания |
| frontend/src/components/messenger/search-dialog.tsx | 170 | 1 | замечания |
| frontend/src/components/messenger/sticker-picker.tsx | 307 | 4 | замечания |
| frontend/src/components/messenger/thread-issue-links.tsx | 114 | 1 | замечания |
| frontend/src/components/messenger/thread-panel.tsx | 125 | 2 | замечания |
| frontend/src/components/messenger/voice-recorder.tsx | 205 | 2 | замечания |

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.

## Counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 3 |
| P2 | 16 |
| P3 | 20 |
| **Σ** | **39** |

| Категория (фокус) | Count |
|-------------------|------:|
| hooks / effect leaks / RoH | 5 |
| a11y | 18 |
| `any` / unsafe casts | 4 |
| design tokens (raw palette) | 8 |
| correctness / UX (вне узкого фокуса, но в файлах) | 4 |

### Топ-3

1. **W9-A3-001 (P1)** — `thread-panel.tsx:39-47`: родитель треда грузится через `getMessages(…, { limit: 1 })` + `find(id)` — почти всегда miss.
2. **W9-A3-002 (P1)** — `message-composer.tsx:320-329`: ReplyPreview есть, но `replyToMessage.id` не уходит в `thread_id` при send.
3. **W9-A3-003 (P1)** — `voice-recorder.tsx:84-96`: `recorder.stop()` до переназначения `onstop` — гонка, Promise может не resolve / chunks потеряны.

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W9-A3-001 | P1 | `frontend/src/components/messenger/thread-panel.tsx:39-47` | Parent: `getMessages(channelId, { limit: 1 })`, затем `find(m => m.id === messageId)` | `limit: 1` = одно новейшее сообщение канала; parent почти никогда не совпадает → шапка треда пустая | Брать parent из infinite-cache `["messenger","messages",channelId]` или API `GET …/messages/:id`; fallback — сообщение из стора/пропсов ChatView |
| W9-A3-002 | P1 | `frontend/src/components/messenger/message-composer.tsx:320-329` (+ `227-228`, `519-523`) | `replyToMessage` только в UI-превью; `sendMessage` шлёт `thread_id: threadId` без `replyToMessage?.id` | «Ответить» визуально работает, сообщение уходит как обычное в канал — reply-сценарий сломан | `thread_id: threadId ?? replyToMessage?.id`; после успеха уже есть `setReplyTo(null)` |
| W9-A3-003 | P1 | `frontend/src/components/messenger/voice-recorder.tsx:84-96` | `recorder.stop()` вызывается **до** `recorder.onstop = …` внутри Promise | Событие `onstop` может сработать со старым handler’ом (startRecording) → Promise зависает, upload не стартует, UI «Запись» ломается | Сначала назначить `onstop` (или `addEventListener('stop')`), потом `stop()`; общий cleanup треков/таймера |
| W9-A3-004 | P2 | `frontend/src/components/messenger/message-item.tsx:546-549` | Toolbar `opacity-0` / `group-hover:…`; без focus-within / keyboard | Клавиатурный/SR пользователь не видит Reply/React/Pin — WCAG 2.1.1 / operable | `group-focus-within:opacity-100` + `focus-within:pointer-events-auto`; или всегда показывать actions для `:focus-within` |
| W9-A3-005 | P2 | `frontend/src/components/messenger/sticker-picker.tsx:198-203` | `useState(activePackId)` **после** `if (!defaultPack) throw` | Conditional до второго hook — паттерн Rules of Hooks (eslint-plugin-react-hooks) | `const [activePackId, setActivePackId] = useState(STICKER_PACKS[0]?.id ?? "")` до любой ветки; throw/guard после всех хуков |
| W9-A3-006 | P2 | `frontend/src/components/messenger/gif-picker.tsx:146-183` | `debounceRef` timeout не чистится на unmount | После закрытия поповера/размонта `searchGifs` → `setState` на unmounted; лишний fetch | `useEffect(() => () => clearTimeout(debounceRef.current), [])` |
| W9-A3-007 | P2 | `frontend/src/components/messenger/pinned-messages.tsx:44-48` | Icon-button Pin: только `title`, нет `aria-label` | Screen reader слышит безымянную кнопку (контраст с Bookmarks/Search-соседями с label) | `aria-label="Закреплённые сообщения"` |
| W9-A3-008 | P2 | `frontend/src/components/messenger/search-dialog.tsx:83` | Trigger Search: `title="Поиск"`, нет `aria-label` | Та же a11y-дыра, что у Pin | `aria-label="Поиск сообщений"` |
| W9-A3-009 | P2 | `frontend/src/components/messenger/chat-view.tsx:363-394` | FAB «вниз» без `aria-label`; badge unread только визуально | Недоступно для SR; счётчик не озвучен | `aria-label={unreadBelowCount ? \`Вниз, ${unreadBelowCount} новых\` : "К последним сообщениям"}` |
| W9-A3-010 | P2 | `frontend/src/components/messenger/invite-link-dialog.tsx:418-421` | Trigger «Пригласить»: `title`, нет `aria-label` | На `sm` текст скрыт (`hidden sm:inline`) — остаётся иконка без accessible name | `aria-label="Пригласить в канал"` |
| W9-A3-011 | P2 | `frontend/src/components/messenger/thread-panel.tsx:54-74` | Close thread: icon-only, нет `aria-label` | Кнопка закрытия треда безымянна для SR | `aria-label="Закрыть тред"` |
| W9-A3-012 | P2 | `frontend/src/components/messenger/channel-list.tsx:790-796` | Поиск каналов: `<input>` без `aria-label` / associated label | Placeholder не замена label (WCAG) | `aria-label="Поиск каналов"` |
| W9-A3-013 | P2 | `frontend/src/components/messenger/channel-list.tsx:188-211` | Collapse section: нет `aria-expanded` | Состояние свёрнутости не объявлено AT | `aria-expanded={!collapsed}` + `aria-controls` |
| W9-A3-014 | P2 | `frontend/src/components/messenger/presence-indicator.tsx:16-18` | Статусы: `bg-green-500` / `bg-gray-400` / `bg-yellow-500` | Обход дизайн-токенов (`--green`, `--gray`, `--amber`) — drift light/dark | `bg-[var(--green)]`, `bg-[var(--gray)]`, `bg-[var(--amber)]` |
| W9-A3-015 | P2 | `frontend/src/components/messenger/link-preview.tsx:33-52` | Internal issue card: `border-blue-500/*`, `text-blue-600` | Не `--blue` / `--hub` surfaces; ломает тему | `border-[var(--blue)]/20`, `text-[var(--blue)]`, `bg-[var(--gB)]` |
| W9-A3-016 | P2 | `frontend/src/components/messenger/link-preview.tsx:81-83` | GitHub card: `border-gray-500/*`, `bg-gray-500/*` | То же — raw Tailwind gray | `border-[var(--b2)]`, `bg-[var(--raised)]` |
| W9-A3-017 | P2 | `frontend/src/components/messenger/message-item.tsx:370` | Bot name: `text-purple-500` | Не токен `--purple` | `text-[var(--purple)]` |
| W9-A3-018 | P2 | `frontend/src/components/messenger/message-item.tsx:105-135` | EditMode: `inputRef` не фокусируется; Save/Cancel без `aria-label` | При старте edit фокус остаётся вне поля; icon-кнопки безымянны | `useEffect(() => inputRef.current?.focus(), [])`; `aria-label` на Save/Cancel |
| W9-A3-019 | P2 | `frontend/src/components/messenger/reaction-picker.tsx:89-96` | Emoji grid buttons — только glyph, без `aria-label` | SR: «кнопка» без имени ×20 | `aria-label={emoji}` или человекочитаемые имена |
| W9-A3-020 | P3 | `frontend/src/components/messenger/message-composer.tsx:318` (+ `460`) | `as unknown as Record<string, unknown>` на Plate JSON | Обход типизации без узкого Plate-типа | `PlateDoc` / shared type вместо double-cast |
| W9-A3-021 | P3 | `frontend/src/components/messenger/sticker-picker.tsx:144` (+ `154`) | `contentJson.children as unknown[]` | Unsafe cast; нет runtime-guard до index | Type guard `Array.isArray` + element narrow (частично уже есть ниже) |
| W9-A3-022 | P3 | `frontend/src/components/messenger/render-content.tsx:160` | `contentJson as unknown[]` при Array.isArray | Лишний cast после guard | `const nodes: unknown[] = contentJson` / generic helper |
| W9-A3-023 | P3 | `frontend/src/components/messenger/mention-autocomplete.tsx:164-166` | Escape → `return true`, но UI не закрывается | Handler глотает Esc; список остаётся (query всё ещё `@…`) | Сброс через callback `onDismiss` / временный «suppressed» state |
| W9-A3-024 | P3 | `frontend/src/components/messenger/gif-picker.tsx:235-246` | Search input без `aria-label` | Placeholder ≠ accessible name | `aria-label="Поиск GIF"` |
| W9-A3-025 | P3 | `frontend/src/components/messenger/message-composer.tsx:698-706` | Emoji popover buttons без `aria-label` | Как W9-A3-019 | `aria-label` на каждый emoji |
| W9-A3-026 | P3 | `frontend/src/components/messenger/sticker-picker.tsx:257-270` | Pack tabs: `title`, нет `aria-label` / `aria-selected` | Табы паков слабо озвучены | `role="tablist"` / `aria-selected` / `aria-label={pack.name}` |
| W9-A3-027 | P3 | `frontend/src/components/messenger/sticker-picker.tsx:285-299` | Sticker cells: `title`, нет `aria-label` | Дубль title для AT недостаточен в части AT | `aria-label={sticker.label}` |
| W9-A3-028 | P3 | `frontend/src/components/messenger/create-channel-dialog.tsx:75-79` | Name `Input` без `Label`/`htmlFor` | Поле без программной метки | `<Label htmlFor="channel-name">` + id |
| W9-A3-029 | P3 | `frontend/src/components/messenger/new-dm-dialog.tsx:123-128` | Search `Input` без label | То же | `aria-label="Поиск пользователей"` |
| W9-A3-030 | P3 | `frontend/src/components/messenger/create-poll-dialog.tsx:188-194` | Remove option: `title`, нет `aria-label` | Icon Trash безымянен для SR | `aria-label="Удалить вариант"` |
| W9-A3-031 | P3 | `frontend/src/components/messenger/invite-link-dialog.tsx:245` (+ `267`) | `bg-orange-500/10`, `text-orange-600`, `text-green-500` | Raw palette vs `--orange` / `--green` | Токены `--orange`, `--green` |
| W9-A3-032 | P3 | `frontend/src/components/messenger/poll-message.tsx:165` (+ `273`) | `text-orange-500` / `hover:text-orange-600` | Не `--orange` | `text-[var(--orange)]` |
| W9-A3-033 | P3 | `frontend/src/components/messenger/voice-recorder.tsx:186` | Recording dot: `bg-red-500` | Не `--red` | `bg-[var(--red)]` |
| W9-A3-034 | P3 | `frontend/src/components/messenger/channel-list.tsx:356` | Star filled: `text-yellow-400` | Не semantic token (`--amber`) | `text-[var(--amber)]` |
| W9-A3-035 | P3 | `frontend/src/components/messenger/thread-issue-links.tsx:85` | Unlink hover: `hover:text-red-500` | Не `--red` | `hover:text-[var(--red)]` |
| W9-A3-036 | P3 | `frontend/src/components/messenger/pinned-messages.tsx:79` | Pin list показывает `ID: {message_id}` | UX/a11y: UUID вместо текста сообщения | Подтянуть `content_text`/preview из API pins |
| W9-A3-037 | P3 | `frontend/src/components/messenger/poll-message.tsx:22` (+ `87-91`) | `channelId` в props не используется; ошибки vote/close глотаются | Мёртвый prop; silent fail без feedback | Использовать `channelId` для WS-room scope или убрать; toast на error |
| W9-A3-038 | P3 | `frontend/src/components/messenger/message-composer.tsx:554-556` | Remove attachment: `title`, нет `aria-label` | Icon-only remove | `aria-label="Убрать вложение"` |
| W9-A3-039 | P3 | `frontend/src/components/messenger/invite-link-dialog.tsx:258-263` (+ `298-304`) | Copy/Revoke icon buttons: `title` only | Нет `aria-label` | `aria-label="Копировать ссылку"` / `"Отозвать"` |

> IDs: W9-A3-001…039. Σ = 39 (P1=3 · P2=16 · P3=20).

## Чисто

Файлы без находок — что проверялось:

- **channel-drawer-context.tsx** — Context+hook, noop default, без side-effects; типов `any` нет.
- **channel-icon.tsx** — decorative SVG с `aria-hidden`; цвета через `var(--t2)`.
- **forward-dialog.tsx** — Dialog+hooks deps OK; empty state; keys по `ch.id`; токены `bg-raised`/`text-t2`.
- **message-actions.tsx** — все icon-кнопки с `aria-label` + `aria-hidden` на SVG; без raw palette.
- **message-attachments.tsx** — `alt={file_name}`, `rel=noopener`, lazy images; keys `file_id`.

Дополнительно (замечания, но не отдельный ID): `message-composer`/`chat-view` в целом хорошо держат `--t*`/`--b*`/`--card`; `mention-autocomplete` listbox/option/`aria-selected` — образец; `render-content` URL sanitize whitelist — security OK; `channel-list` mute aria + virtualizer isolation — hooks-корректны.
