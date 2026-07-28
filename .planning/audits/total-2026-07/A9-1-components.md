# W9 агент A1 — frontend/components part 1

> TASK-008 · Стадия A · read-only · 2026-07-28  
> Фокус FE: useEffect leaks · hook deps · a11y · Shadcn className-only · any/ts-ignore · loading/empty/error

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| frontend/src/components/admin/ai-settings.tsx | 364 | 1 | замечания |
| frontend/src/components/admin/audit-log.tsx | 574 | 1 | замечания |
| frontend/src/components/admin/mail-settings.tsx | 200 | 1 | замечания |
| frontend/src/components/admin/onec-settings.tsx | 175 | 1 | замечания |
| frontend/src/components/admin/operation-metrics.tsx | 139 | 0 | чисто |
| frontend/src/components/admin/pending-users-panel.tsx | 144 | 0 | чисто |
| frontend/src/components/admin/resource-limits.tsx | 110 | 1 | замечания |
| frontend/src/components/admin/trash-manager.tsx | 204 | 1 | замечания |
| frontend/src/components/admin/user-management.tsx | 690 | 2 | замечания |
| frontend/src/components/ai/ai-chat-panel.tsx | 300 | 1 | замечания |
| frontend/src/components/ai/byok-settings.tsx | 101 | 1 | замечания |
| frontend/src/components/ai/model-picker.tsx | 46 | 0 | чисто |
| frontend/src/components/auth/dev-login-panel.tsx | 285 | 0 | чисто |
| frontend/src/components/auth/magic-link-form.tsx | 242 | 0 | чисто |
| frontend/src/components/calls/CallButton.tsx | 101 | 0 | чисто |
| frontend/src/components/calls/CallChat.tsx | 208 | 4 | замечания |
| frontend/src/components/calls/CallControls.tsx | 312 | 0 | чисто |
| frontend/src/components/calls/CallPanel.tsx | 96 | 1 | замечания |
| frontend/src/components/calls/CallReactions.tsx | 246 | 1 | замечания |
| frontend/src/components/calls/CallRoom.tsx | 404 | 1 | замечания |
| frontend/src/components/calls/CallSettingsModal.tsx | 307 | 1 | замечания |
| frontend/src/components/calls/DeviceSelector.tsx | 159 | 0 | чисто |
| frontend/src/components/calls/ParticipantList.tsx | 179 | 0 | чисто |
| frontend/src/components/calls/ScreenShareView.tsx | 54 | 0 | чисто |
| frontend/src/components/dashboard/dashboard-templates.ts | 174 | 0 | чисто |
| frontend/src/components/dashboard/widget-grid.tsx | 415 | 1 | замечания |

Файлов в списке: 26. Прочитано: 26. Пропущено: 0.

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W9-A1-001 | P1 | frontend/src/components/calls/CallChat.tsx:55 | Нет `subscribe("call:"+callId)` на комнату WS, куда bridge шлёт `chat.message` (`backend/internal/ws/bridge.go:212–227`). По репо нет ни одной подписки на `call:` | Real-time сообщения чата звонка не доходят; `sendMutation` только чистит input и не пишет в cache — отправитель тоже не видит своё сообщение без refetch | В `useEffect` вызывать `subscribe`/`unsubscribe` на `call:${callId}`; в `onSuccess` мутации добавлять сообщение в query cache или invalidate |
| W9-A1-002 | P1 | frontend/src/components/calls/CallChat.tsx:56 | Handler игнорирует имя event (`_event`), принимает любой payload с `call_id === callId` | Глобальный `addListener` получает все WS-события клиента. `call.joined`/`call.left` (room `channel:*`) тоже несут `call_id` → в ленту чата попадают «сообщения» без `text`/`id` | Фильтровать `event === "chat.message"` (+ проверка полей `id`/`text`) |
| W9-A1-003 | P2 | frontend/src/components/admin/user-management.tsx:407 | Кнопка «Пригласить» без `onClick` | Мёртвый UI: админ жмёт — ничего не происходит; invite уже копируется отдельной кнопкой выше | Убрать кнопку или привязать к копированию invite / созданию инвайта |
| W9-A1-004 | P2 | frontend/src/components/admin/user-management.tsx:69 | `ReassignDialog` / `GdprDeleteDialog` / ban-modal (стр. 126, 621) — кастомный overlay без `role="dialog"`, `aria-modal`, focus trap | Screen reader / клавиатура: нет семантики модалки (в `audit-log` ban-dialog ARIA есть — расхождение) | Перевести на `Dialog`/`ConfirmDialog` или добавить ARIA + focus/Escape как в Shadcn Dialog |
| W9-A1-005 | P2 | frontend/src/components/admin/resource-limits.tsx:48 | `setTimeout(..., 3000)` для toast без cleanup; `loading` стартует `false` (стр. 19) | Unmount до 3s → setState на размонтированном; первый paint — пустая форма без «Загрузка» | Хранить timer id + clear в cleanup; `useState(true)` для loading; abort/cancelled flag на fetch |
| W9-A1-006 | P2 | frontend/src/components/calls/CallSettingsModal.tsx:280 | `ToggleRow` — кастомный switch без `role="switch"` / `aria-checked` | a11y: состояние тоггла недоступно AT | Добавить `role="switch"` + `aria-checked={checked}` на button |
| W9-A1-007 | P2 | frontend/src/components/calls/CallChat.tsx:193 | Кнопка submit только с иконкой `Send`, без `aria-label` | Icon-only control без accessible name | `aria-label="Отправить"` |
| W9-A1-008 | P2 | frontend/src/components/dashboard/widget-grid.tsx:374 | Кнопка меню виджета (`Ellipsis`) без `aria-label` | Icon-only, недоступна screen reader | `aria-label={`Меню виджета ${name}`}` |
| W9-A1-009 | P2 | frontend/src/components/ai/ai-chat-panel.tsx:203 | `role="dialog"` без `aria-modal="true"` | Неполная семантика модальной панели | Добавить `aria-modal="true"` (+ focus management при открытии) |
| W9-A1-010 | P2 | frontend/src/components/admin/mail-settings.tsx:35 | После ошибки `getMailSettings` `loading=false`, `settings` остаётся `null`, форма рендерится как пустые настройки | Нет error/empty-error UI — админ правит «пустой» стейт, думая что это сервер | Error-state + retry (как `operation-metrics.tsx:49–61`); то же в `ai-settings.tsx:57`, `onec-settings.tsx:42` |
| W9-A1-011 | P2 | frontend/src/components/admin/audit-log.tsx:94 | `fetchLogs` без abort/cancelled; interval 30s + смена фильтров | Гонка: устаревший ответ перетирает свежий список | AbortController или generation counter / cancelled flag |
| W9-A1-012 | P2 | frontend/src/components/admin/ai-settings.tsx:57 | То же, что W9-A1-010: ошибка загрузки → форма «пустого» реестра без error UI | Путает empty и fail | Error + retry; не сбрасывать список в тишину |
| W9-A1-013 | P2 | frontend/src/components/admin/onec-settings.tsx:42 | То же, что W9-A1-010 для 1C-панели | Путает empty и fail | Error + retry |
| W9-A1-014 | P3 | frontend/src/components/calls/CallRoom.tsx:203 | `setTimeout(() => setCopied(false), 3000)` в `InviteBlock` без cleanup | Unmount → setState на размонтированном | clearTimeout в cleanup / ref |
| W9-A1-015 | P3 | frontend/src/components/ai/byok-settings.tsx:23 | `refresh` глотает ошибки (`catch { /* … */ }`); нет loading | Сбой API BYOK невидим; пустой список неотличим от ошибки | toastApiError + loading/empty |
| W9-A1-016 | P3 | frontend/src/components/calls/CallPanel.tsx:55 | Кнопка fullscreen/minimize — только `title`, нет `aria-label` | Icon-only a11y | `aria-label` по состоянию |
| W9-A1-017 | P3 | frontend/src/components/calls/CallReactions.tsx:104 | Кнопки реакций — только `title`, нет `aria-label` | AT может игнорировать `title` | `aria-label={label}` |
| W9-A1-018 | P3 | frontend/src/components/admin/trash-manager.tsx:76 | Вкладки категорий — обычные `button` без `role="tablist"` / `aria-selected` | Слабая a11y навигации вкладок | tablist/tab + `aria-selected` |
| W9-A1-019 | P3 | frontend/src/components/calls/CallChat.tsx:45 | `useQuery` без loading/error UI | Пока грузится история — пустой empty «Нет сообщений»; ошибка сети невидима | isPending/isError fallbacks |

## Сводка

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 2 |
| P2 | 11 |
| P3 | 6 |
| **Σ** | **19** |

### Топ-3
1. **W9-A1-001** — CallChat не подписан на WS-комнату `call:{id}` → real-time чат звонка фактически не работает.
2. **W9-A1-002** — тот же handler принимает любые события с `call_id` → мусор в ленте из `call.joined`/`left`.
3. **W9-A1-003** — мёртвая кнопка «Пригласить» в user-management.

## Чисто
Проверено по чек-листу FE (утечки useEffect, deps, a11y, Shadcn className-only, any/ts-ignore, loading/empty/error):

| Файл | Что проверено |
|------|----------------|
| operation-metrics.tsx | loading/error/empty + retry; deps `[load]` ок; interval/listener нет |
| pending-users-panel.tsx | load + ConfirmDialog; loading/empty; cancelled не нужен (короткий mount admin) |
| model-picker.tsx | Shadcn Select только className; empty «Нет моделей»; loadModels в effect |
| dev-login-panel.tsx | DEV_MODE gate; error UI; aria-label на карточках; без any/ts-ignore |
| magic-link-form.tsx | labels/htmlFor; loading; toast errors; без таймеров/effects |
| CallButton.tsx | cancelled flag; TooltipTrigger asChild (Radix не ломается); aria-label |
| CallControls.tsx | aria-label на ControlButton; Dialog Shadcn className-only |
| DeviceSelector.tsx | devicechange + removeEventListener cleanup; Select className-only |
| ParticipantList.tsx | стабильные key=`identity`; hooks order ок |
| ScreenShareView.tsx | early null; key по identity |
| dashboard-templates.ts | чистые данные LayoutItem[], логики UI нет |

**Shadcn / Radix:** в зоне A1 нет правок поведения Radix — только `className` / `asChild` на Trigger.  
**any / @ts-ignore / @ts-expect-error:** в 26 файлах не найдено. Есть узкие assertions (`as CallChatMessage`, `as RegistrationMode`, `import.meta.env.VITE_LIVEKIT_URL as string`) — не заводились как дефекты без runtime-багов.  
**Хорошие паттерны для копирования:** `audit-log` interval+cleanup; `CallReactionsOverlay` clearTimeout map; `CallSettingsModal`/`DeviceSelector` devicechange cleanup; `user-management` invite/regMode cancelled flags; `operation-metrics` error+retry.
