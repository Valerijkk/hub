# Frontend Audit — Hub (post-Milestone v2.0)

**Дата:** 2026-05-15
**Стек:** React 19 + Vite 6 + TanStack Router/Query + Tailwind 4 + Shadcn + LiveKit + Zustand
**Codebase:** 249 TS/TSX файлов, 45 396 строк, 264 проверяемых файлов (биомом)

## TL;DR — Score: **76 / 100**

- TS strict: **0 errors / 0 warnings** в `tsc --noEmit`.
- Biome: **575 errors, 109 warnings, 13 infos** (47 из 575 — generated `routeTree.gen.ts` с `@ts-nocheck`, реальных ≈ 528).
- Архитектура целостная: ApiClient + refresh-token, ErrorBoundary в root + `messenger/$channelId`, code-splitting вендоров настроен.
- Дыры: a11y/keyboard в кастомных кнопках/menu, hooks rules-of-hooks нарушения, отсутствие multi-tab sync для auth, локальные edge cases в WebRTC.

## 1. TypeScript

| Метрика | Значение |
|---|---|
| `tsc --noEmit` errors | **0** |
| `@ts-ignore` / `@ts-nocheck` | 1 (generated `routeTree.gen.ts`) |
| `as any` / `: any` в ручном коде | **1** (`lib/utils.ts:15` — `buildQueryString(params?: Record<string, any>)`) |
| `as any` в generated | 46 (игнорируется) |

**Вывод:** Type safety **отличная**, TS strict проходит. Один реальный `any` в utility — тривиальный фикс на `unknown`.

## 2. React Anti-patterns (biome)

| Правило | Кол-во | Severity |
|---|---|---|
| `useExhaustiveDependencies` | 15 | HIGH |
| `useHookAtTopLevel` | **8** (все в `message-item.tsx`) | **CRITICAL** |
| `noNonNullAssertion` (`x!`) | 50 | MEDIUM |
| `noArrayIndexKey` | 16 | MEDIUM |
| `noExplicitAny` (ручной) | 1 | LOW |

**CRITICAL:** `components/messenger/message-item.tsx:167-260` — `useMemo`/`useCallback` вызываются **после** условных return / в коллбэках, ломает Rules of Hooks. Может дать React #185 (что и подтверждает наличие `ChatErrorBoundary` в `$channelId.tsx`). **Это причина существующего error boundary** — симптом баги, а не профилактика.

**HIGH:** Лишние deps в useEffect (`CallButton`, `CallChat`, `DeviceSelector`, `chat-view` x3, `mention-autocomplete`, `notification-settings`, `jql-editor` x2, `kanban-board` x3, `sprint-backlog`, `trash-manager`) — реренды и потенциальные двойные subscribe.

## 3. Accessibility

- Радиксовые `Dialog/DropdownMenu/Popover` дают focus-trap/Esc/aria из коробки — OK.
- **8 button-type missing** (биом `useButtonType`) — нарушает поведение внутри `<form>`.
- **0 случаев** `tabIndex`/`onKeyDown` в `components/calls/*` — кастомные кнопки чата/контролов **не работают с клавиатурой** (Tab+Enter/Space).
- `routes/login.tsx`: декоративные feature-cards без `aria-hidden`, mesh-gradient div'ы рендерятся как контент для screen reader.
- Эмодзи-реакции в `message-item.tsx:425` — `<button>` без `aria-label` (только эмодзи + count).

## 4. Performance / Bundle

| Chunk | Size |
|---|---|
| `index.js` (main) | **1.1 MB** ⚠️ |
| `vendor-livekit` | 478 KB |
| `vendor-charts` | 389 KB |
| `vendor-plate` | 317 KB |
| `vendor-tanstack` | 161 KB |
| `vendor-ui` | 145 KB |
| `vendor-motion` | 94 KB |

**Проблема:** main bundle ≈ 1.1 MB до минимум-gzip — нет route-level code-splitting кроме `dashboard.lazy`, `cases.lazy`, `runs.lazy`. **`livekit-client` НЕ lazy-loaded** — он попадает в vendor-livekit, но импортируется напрямую в `CallControls`/`CallRoom`/`ParticipantList`, значит eager. Для пользователя без звонков 478 KB лишних.

**Динамический import** найден **1 раз** (`message-item.tsx:433` для toggleReaction) — антипаттерн, должно быть на уровне route, не на клик.

`React.memo` использован **3 раза** на 249 файлов — для тяжёлых списков (`message-item`, `kanban-board`, `notification-panel`) это **недостаточно**. `MessageItem` ре-рендерится на каждый apdate чата.

## 5. Error Boundaries

- `routes/__root.tsx`: `RootErrorBoundary` (`errorComponent` TanStack) — OK.
- `routes/_authenticated/messenger/$channelId.tsx`: локальный `ChatErrorBoundary` — компенсирует Rules-of-Hooks нарушения в `message-item.tsx` (см. §2 CRITICAL).
- **Нет EB** вокруг: calls (`CallRoom`, `/call/$code`), dashboard widgets, tracker boards, TMS routes. Любой ран-тайм краш = весь экран сломан.
- Отсутствует Sentry/телеметрия — `console.error` в `componentDidCatch` без отправки.

## 6. API Integration

`src/api/client.ts` — единый ApiClient:

- 401 → автоматический refresh → retry. **OK**, но есть race: при параллельных запросах N штук получат 401, все запустят `tryRefresh` параллельно — N лишних refresh-вызовов. **Нет single-flight mutex.**
- 403/5xx → пробрасывается как `ApiError`. UI обрабатывает в 2 файлах (`isError` встречается в `link-preview.tsx`). **Нет глобального toast при 5xx** — теряется UX.
- `request<T>` рекурсивно вызывает себя при refresh — если refresh вернёт 401 → бесконечная рекурсия (защита есть через `refreshed=false → logout`, но обработка edge `refresh_token` отозван не реализована).

## 7. State Management

- **Auth:** zustand `auth-store` с `sessionStorage` (правильно — сессия в табе). `setTokens` обновляет state, но **multi-tab broadcast отсутствует** (нет `BroadcastChannel`, нет `storage` event listener) → logout в одной вкладке оставляет другую залогиненной до 401.
- **TanStack Query:** root `QueryClient` с `staleTime: 30s`, `retry: 1` — разумно.
- **Конфликт truth-source:** `message-item.tsx:230` делает `queryClient.setQueryData(["messenger","messages",channelId], ...)` напрямую + invalidateMessages в catch — оптимистический апдейт без `useMutation` (нет автоматического rollback).
- Стейт-сторов **18+** (`auth`, `messenger`, `notification`, `dashboard`, `sidebar`, `call`...), но границы чёткие.

## 8. CSS / Design System

- `app.css` 1024 строки. `!important` встречается **4 раза** — все в `@media (prefers-reduced-motion: reduce)` (валидное использование). Цель Phase 8 «0 !important в bizness CSS» — **выполнена**.
- Хардкод HSL/RGB цветов в `routes/login.tsx` (`rgba(91,156,245,.1)` для иконок) — должны быть `var(--blue-alpha-10)`.
- `CallRoom.tsx:196` — `bg-[hsl(220,11%,12%)]` хардкод серого в звонках.
- 12 биом-варнингов в `app.css` (включая `useTemplate`, `noImportantStyles`).

## 9. Calls / WebRTC

- `livekit-client` импортируется в 6 файлах **синхронно** — попадает в main render path. См. §4.
- `voice-recorder.tsx:37` обрабатывает getUserMedia в try/catch → юзер-фрэндли error message. **OK.**
- `CallSettingsModal` (недавно правленный) — `enumerateDevices` обёрнут в try, devicechange listener cleanup-нут. **OK.**
- **Cleanup проблема:** `CallRoom.tsx:47-52` — useEffect мьютит микрофон, но deps включает `localParticipant.isMicrophoneEnabled` который меняется → может дёргать `setMicrophoneEnabled(false)` повторно. Использует `didMuteRef` как защиту, но логика хрупкая.
- `MediaRecorder` в `voice-recorder.tsx`: stream stop в `onstop`, но **если компонент анмаунтится во время записи** — stream остаётся открытым (нет cleanup в useEffect return).

## 10. Auth Flow

- Login через Telegram + dev panel. `useAuthStore.setTokens` синхронно → ApiClient берёт через `getState()`.
- **Refresh edge cases:**
  - Race condition (см. §6).
  - При invalid `refresh_token` (отозван) — `tryRefresh` вернёт false, logout. **OK.**
  - **Нет refresh заранее** (proactive) — токены обновляются только по 401, значит первый запрос после истечения всегда падает и ретраится.
- **Logout edge cases:**
  - `logout()` чистит store, но не делает `POST /auth/logout` на бэк → refresh token остаётся валидным на сервере (security risk).
  - **Нет multi-tab синхронизации** logout.
- `onboarding_completed` flag читается в `login.tsx:69` — OK, но user object может быть null после refresh (стор персистит токены, но не user).

## Top-10 файлов по issues (без generated)

| # | Файл | Issues | LOC | Заметки |
|---|---|---|---|---|
| 1 | `components/messenger/message-item.tsx` | 16 | 513 | **Rules-of-Hooks нарушения x8** |
| 2 | `routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.tsx` | 16 | ? | |
| 3 | `routes/_authenticated/projects/$projectId/tms/cases.lazy.tsx` | 16 | ? | |
| 4 | `routes/onboarding.tsx` | 16 | ? | |
| 5 | `components/messenger/chat-view.tsx` | 13 | 415 | non-null asserts на typingUsers |
| 6 | `routes/_authenticated/trash.tsx` | 12 | ? | |
| 7 | `modules/calls/call-calendar.tsx` | 12 | ? | |
| 8 | `app.css` | 12 | 1024 | reduce-motion !important OK |
| 9 | `components/search/jql-editor.tsx` | 10 | 456 | exhaustive-deps x2 |
| 10 | `components/messenger/message-composer.tsx` | 9 | **742** | **самый большой компонент** |

**File-size code smells (>300 lines):** 19 файлов. Топ: `message-composer` (742), `channel-list` (619), `message-item` (513), `jql-editor` (456), `invite-link-dialog` (451), `chat-view` (415), `gif-picker` (391), `widget-grid` (381), `jql-search-page` (378), `user-management` (373).

## Severity-classified findings

### CRITICAL (блокер)
1. **Rules-of-Hooks нарушения в `message-item.tsx`** (8 шт). Это причина существующего ChatErrorBoundary — баги, не профилактика. Может крашить React в продакшене.
2. **Refresh-token race** в `api/client.ts` — N параллельных 401 → N refresh-запросов, возможна потеря токенов если сервер инвалидирует старый.

### HIGH
3. **Bundle 1.1 MB main** + 478 KB livekit eager — медленный first-paint для пользователей без звонков.
4. **Нет multi-tab auth sync** — logout/login не пробрасывается через BroadcastChannel.
5. **Нет server-side logout** — refresh_token остаётся валидным после клиент-логаута.
6. **Нет Error Boundaries** вокруг calls/dashboard/tracker/TMS — runtime крах = белый экран.
7. **15 exhaustive-deps нарушений** в эффектах (потенциальные двойные subscribe/реренды).

### MEDIUM
8. **50 non-null assertions** (`x!`) — потенциальные runtime crashes.
9. **16 array-index keys** — некорректные ре-рендеры в списках.
10. **a11y в calls** — нет keyboard nav в custom buttons.
11. **8 button без `type=`** — submit-by-default баг в формах.
12. **MediaRecorder stream cleanup** в voice-recorder при unmount.
13. **Хардкод цветов** вне CSS-переменных (login, CallRoom).

### LOW
14. 1 `any` в `lib/utils.ts`.
15. console.log/warn/error 25 occurrences в 8 файлах (дев-стори, не блокер).
16. **3 React.memo** на 249 файлов — недо-оптимизация списков.
17. 7 `useTemplate` биом-варнингов.
18. 0 TODO/FIXME — чистый код.

## Рекомендации (приоритет)

1. **Срочно** разрулить hooks в `message-item.tsx` → можно убрать `ChatErrorBoundary` как костыль.
2. Добавить single-flight mutex в `ApiClient.tryRefresh` (Promise dedup).
3. Route-level `lazy()` для `/calls`, `/dashboard`, `/tracker/board` — выкинуть livekit/recharts/plate из main.
4. `BroadcastChannel("hub-auth")` в `auth-store` для multi-tab logout.
5. `POST /auth/logout` на бэк перед очисткой стора.
6. Error Boundary HOC и обернуть calls/dashboard/board.
7. Прогон `biome check --write --unsafe` для авто-фиксов (575 → ~100 ручных).
8. `React.memo(MessageItem)` + `useMemo` для `messages` в chat-view.

---

**Артефакты:** `tsc-output.txt` (пустой = 0 ошибок), `biome-output.txt` (полный лог).
