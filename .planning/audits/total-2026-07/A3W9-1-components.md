# A3 W9 агент 1 — frontend/components part 1 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W9-1-components.md](A2W9-1-components.md)  
> Stage A: [A9-1-components.md](A9-1-components.md)  
> Периметр: 26 файлов (W9-ASSIGNMENTS A1) · Прочитано: 26 · Пропущено: 0  
> Фокус FE: XSS/`dangerouslySetInnerHTML` · a11y · Rules of Hooks · `--hub-*` / токены · Shadcn className-only · secrets in UI

## Сверка с A2 (A2W9-1-*)

| A2 / Stage A ID | Tag (A2) | Статус в HEAD | Доказательство |
|-----------------|----------|---------------|----------------|
| W9-A1-001 | FIXED-OK | без регресса | `CallChat.tsx:78-85` — `subscribe`/`unsubscribe` на `call:${callId}` |
| W9-A1-002 | FIXED-OK | без регресса | `CallChat.tsx:87-90` + `isCallChatMessage` (`:35-48`) — фильтр `chat.message` |
| W9-A1-003 | FIXED-OK | без регресса | `user-management.tsx:450-459` — «Пригласить» → clipboard invite |
| W9-A1-004 | FIXED-OK | без регресса | Reassign/GDPR/ban: `role="dialog"` + `aria-modal` (`:79-81`, `:151-153`, `:671-673`) |
| W9-A1-005 | FIXED-OK | без регресса | `resource-limits.tsx:22-52` — toast timer ref + clear + cancelled fetch; `loading` init `true` |
| W9-A1-006 | FIXED-OK | без регресса | `CallSettingsModal.tsx:284-286` — `role="switch"` + `aria-checked` |
| W9-A1-007 | FIXED-OK | без регресса | `CallChat.tsx:265` — `aria-label="Отправить"` |
| W9-A1-008 | FIXED-OK | без регресса | `widget-grid.tsx:377` — `aria-label={`Меню виджета ${name}`}` |
| W9-A1-009 | FIXED-OK | без регресса | `ai-chat-panel.tsx:203-205` — `aria-modal="true"` |
| W9-A1-010 | FIXED-OK | без регресса | `mail-settings.tsx:25,74-79` — `loadError` + `QueryErrorAlert` |
| W9-A1-011 | FIXED-OK | без регресса | `audit-log.tsx:94-117` — cancelled signal + interval cleanup |
| W9-A1-012 | FIXED-OK | без регресса | `ai-settings.tsx:29,112-117` — `loadError` + `QueryErrorAlert` |
| W9-A1-013 | FIXED-OK | без регресса | `onec-settings.tsx:31,100-105` — `loadError` + `QueryErrorAlert` |
| W9-A1-014 | FIXED-OK | без регресса | `CallRoom.tsx:196-212` — `copyTimerRef` + clear on unmount |
| W9-A1-015 | FIXED-OK | без регресса | `byok-settings.tsx:17,30,70-75` — toast + loading/error/`QueryErrorAlert` |
| W9-A1-016 | FIXED-OK | без регресса | `CallPanel.tsx:65` — fullscreen `aria-label` |
| W9-A1-017 | FIXED-OK | без регресса | `CallReactions.tsx:115` — `aria-label={label}` |
| W9-A1-018 | FIXED-OK | без регресса | `trash-manager.tsx:76-82` — `role="tablist"` / `aria-selected` |
| W9-A1-019 | FIXED-OK | без регресса | `CallChat.tsx:62-63,176-200` — `isPending` / `isError` / empty разделены |

**REGRESS:** 0 · **RESIDUAL (A2 actives):** 0 · **NEW (этот слот):** 5

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/components/admin/ai-settings.tsx` | 367 | 0 | чисто (A1 fixed; API key `type=password`, не echo) |
| `frontend/src/components/admin/audit-log.tsx` | 564 | 1 | замечания (error≠empty) |
| `frontend/src/components/admin/mail-settings.tsx` | 200 | 0 | чисто (A1 fixed) |
| `frontend/src/components/admin/onec-settings.tsx` | 175 | 0 | чисто (A1 fixed; bearer `type=password`, mask placeholder) |
| `frontend/src/components/admin/operation-metrics.tsx` | 130 | 0 | чисто (эталон error+retry) |
| `frontend/src/components/admin/pending-users-panel.tsx` | 136 | 1 | замечания (error≠empty) |
| `frontend/src/components/admin/resource-limits.tsx` | 120 | 1 | замечания (error≠empty после A1 timer-fix) |
| `frontend/src/components/admin/trash-manager.tsx` | 194 | 1 | замечания (error≠empty; tabs ARIA OK) |
| `frontend/src/components/admin/user-management.tsx` | 713 | 0 | чисто (A1 fixed; QueryStateView) |
| `frontend/src/components/ai/ai-chat-panel.tsx` | 280 | 0 | чисто (A1 fixed; BotMarkdown text nodes, без `dangerouslySetInnerHTML`) |
| `frontend/src/components/ai/byok-settings.tsx` | 119 | 0 | чисто (A1 fixed; key `type=password`) |
| `frontend/src/components/ai/model-picker.tsx` | 42 | 0 | чисто (Shadcn Select className-only) |
| `frontend/src/components/auth/dev-login-panel.tsx` | 262 | 0 | чисто (gate `/dev/status`; hooks до early return) |
| `frontend/src/components/auth/magic-link-form.tsx` | 231 | 0 | чисто (labels/htmlFor; tokens только в store) |
| `frontend/src/components/calls/CallButton.tsx` | 113 | 0 | чисто |
| `frontend/src/components/calls/CallChat.tsx` | 262 | 1 | замечания (logging toggle a11y) |
| `frontend/src/components/calls/CallControls.tsx` | 289 | 0 | чисто |
| `frontend/src/components/calls/CallPanel.tsx` | 89 | 0 | чисто (A1 fixed) |
| `frontend/src/components/calls/CallReactions.tsx` | 217 | 0 | чисто (A1 fixed) |
| `frontend/src/components/calls/CallRoom.tsx` | 385 | 0 | чисто (A1 fixed) |
| `frontend/src/components/calls/CallSettingsModal.tsx` | 281 | 0 | чисто (A1 fixed) |
| `frontend/src/components/calls/DeviceSelector.tsx` | 146 | 0 | чисто (devicechange cleanup) |
| `frontend/src/components/calls/ParticipantList.tsx` | 164 | 0 | чисто (`key={identity}`) |
| `frontend/src/components/calls/ScreenShareView.tsx` | 47 | 0 | чисто |
| `frontend/src/components/dashboard/dashboard-templates.ts` | 168 | 0 | чисто (данные LayoutItem[]) |
| `frontend/src/components/dashboard/widget-grid.tsx` | 394 | 0 | чисто (A1 fixed; lazy charts) |

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W9-1-001 | NEW | P2 | `pending-users-panel.tsx:26-28` (+ `:86-87`) | `catch` → `setPending([])` → UI «Нет заявок на регистрацию» | Fail = empty: админ не видит сбой API, путает с пустой очередью (класс дефекта W9-A1-010 / TASK-093, но этот файл A1/A2 пометили «чисто») | `loadError` + `QueryErrorAlert`/retry; не сбрасывать в empty без error-ветки (эталон: `operation-metrics.tsx:49-61`) |
| A3W9-1-002 | NEW | P2 | `resource-limits.tsx:43-48` (+ render `:82+`) | После fail `getLimits` — только toast; `limits=[]`, форма без полей / как «нет лимитов» | W9-A1-005 закрыл timer/cancelled/loading, но **не** error≠empty; соседние admin-панели уже на `QueryErrorAlert` | `loadError` + retry; не показывать save-форму при fail |
| A3W9-1-003 | NEW | P2 | `trash-manager.tsx:38-40` (+ `:113-117`) | Fail `getTrash` → toast; таблица «Корзина пуста» | То же: empty ≠ error; вкладки/ARIA (W9-A1-018) OK, load-fail UI нет | `loadError` + alert/retry; empty только при `!error && items.length===0` |
| A3W9-1-004 | NEW | P3 | `audit-log.tsx:101-104` (+ `:414-421`) | Fail `fetchLogs` → toast; при пустом `entries` — «Записи не найдены» | Interval+cancelled (W9-A1-011) OK; первичная/полная ошибка всё ещё выглядит как empty filter | Флаг `loadError` / banner+retry над таблицей; empty copy только без error |
| A3W9-1-005 | NEW | P3 | `CallChat.tsx:143-164` | Toggle логирования чата: только `title`, нет `aria-label` / `aria-pressed` | Icon+короткий текст; AT может не получить имя/состояние (симметрия с W9-A1-007/017) | `aria-label` по состоянию + `aria-pressed={chatLoggingEnabled}` |

## FIXED-OK (сверка A2 + FE-фокус)

| Зона | Проверено в HEAD |
|------|------------------|
| XSS | В 26 файлах **нет** `dangerouslySetInnerHTML` / `innerHTML`. AI assistant → `BotMarkdown` (React text nodes). CallChat `msg.text` — text children |
| Secrets in UI | AI/BYOK/1C keys: `type="password"`; `has_api_key` / `bearer_token_set` без echo секрета; JWT только через `setTokens` store; invite URL — clipboard, не hardcode |
| A1 a11y/leak fixes | CallChat WS+filter+send label · dialogs ARIA · ToggleRow switch · widget menu · trash tabs · timers cancelled |
| Hooks | Early returns после всех hooks (`ai-chat-panel`, `CallButton`, `dev-login`); biome-ignore на intentional deps документированы |
| Shadcn | Select/Button/Dialog/Dropdown — только `className` / `asChild`; Radix behavior не ломается |
| Tokens | Широко `var(--*)` / semantic Tailwind; legacy hex в call chrome (`#0B0F19`) / gradient `to-[#2563EB]` — hygiene, не security; не заводилось как open без UX-регресса |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 3 |
| P3 | 2 |
| **Total** | **5** |

| Метрика | Значение |
|---------|----------|
| Файлов | **26 / 26 / 0** |
| Активных | **5** (NEW 5 · REGRESS 0 · RESIDUAL 0) |
| FIXED-OK (Stage A via A2) | **19/19** без регресса |

### Топ-3

1. **A3W9-1-001 (P2, NEW)** — `pending-users-panel`: fail load → «Нет заявок».
2. **A3W9-1-002 (P2, NEW)** — `resource-limits`: после A1 timer-fix всё ещё нет error UI.
3. **A3W9-1-003 (P2, NEW)** — `trash-manager`: fail → «Корзина пуста».

---

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=3 · P3=2 · Total=5.  
Топ: A3W9-1-001 … A3W9-1-003 (error≠empty admin).
