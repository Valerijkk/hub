# W12 — реестр уникальных P2-классов (frontend, волны 9–11)

> TASK-008 · Стадия B · W12 · 2026-07-28  
> Источники: `A{9-11}-*.md`, сводки `W9-SUMMARY`…`W11-SUMMARY`, [`00-INDEX.md`](00-INDEX.md)  
> Scope: **только frontend** (W9–W11) + docs, попавшие в W11 A4/A5.  
> Backend W1–W8 — вне артефакта: [`W12-REGISTRY-P2-BACKEND.md`](W12-REGISTRY-P2-BACKEND.md).  
> **TASK-\*.md не создаются** — только реестр + предложение группировки.  
> Формат: зеркало P2-backend + таблица классов как у [`W12-REGISTRY-P1-FRONTEND.md`](W12-REGISTRY-P1-FRONTEND.md).

## Сводка

| Метрика | Значение |
|---------|---------:|
| Сырых P2 (W9–W11) | **165** |
| Уникальных классов | **20** |
| Из них с критерием мутации «да» | **16** |
| В секции «требует уточнения» | **4** |
| Предлагаемых задач бэклога | **10** |

Сырые по волнам: W9=52 · W10=64 · W11=49 → **165**  
(полный Stage A P2=422 = 257 backend + 165 frontend).

P0/P1 FE **не** в таблице: [`W12-REGISTRY-P1-FRONTEND.md`](W12-REGISTRY-P1-FRONTEND.md) (P0 FE = 0).

---

## Уникальные классы дефектов

| ClassID | Severity | Title | Source IDs | Primary file:line | Mutation criterion | Suggested TASK slug |
|---------|----------|-------|------------|-------------------|--------------------|---------------------|
| P2-FE-01 | P2 | Fail/сеть = empty: нет `isError` / catch→`[]` | W9-A1-010, W9-A1-012, W9-A1-013, W9-A2-009, W9-A4-004, W9-A4-006, W9-A5-006, W10-A1-004, W10-A1-005, W10-A1-006, W10-A1-007, W10-A1-008, W10-A1-009, W10-A1-011, W10-A1-012, W10-A1-013, W10-A1-014, W10-A1-017, W10-A1-018, W10-A1-020, W10-A2-004, W10-A2-005, W10-A2-006, W10-A2-007, W10-A2-008, W10-A2-009, W10-A2-010, W10-A2-016, W10-A3-003, W10-A3-004, W10-A4-007, W10-A5-005, W10-A5-006, W10-A5-007, W11-A1-003, W11-A1-004, W11-A2-002, W11-A2-003, W11-A3-012, W11-A3-014 | `frontend/src/modules/bases/bases-page.tsx:41` | **да** — mock listBases reject → UI «ошибка/retry», не empty sidebar; убрать `isError` → снова «нет баз» | `fe-p2-error-vs-empty` |
| P2-FE-02 | P2 | Мутации без feedback / optimistic без rollback / confirm закрывает на reject | W9-A5-003, W10-A1-015, W10-A1-019, W10-A1-021, W10-A3-010, W10-A3-011, W10-A3-015, W11-A3-008, W11-A3-009, W11-A4-004 | `frontend/src/components/ui/confirm-dialog.tsx:74` | **да** — `onConfirm` reject → диалог остаётся open; закрыть в `finally` → тест «диалог закрыт при ошибке» краснеет; markRead fail → badge откатывается | `fe-p2-mutation-feedback` |
| P2-FE-03 | P2 | Icon-only / controls без accessible name (aria-label, label) | W9-A1-006, W9-A1-007, W9-A1-008, W9-A2-004, W9-A2-005, W9-A2-011, W9-A2-012, W9-A2-013, W9-A2-014, W9-A2-015, W9-A3-004, W9-A3-007, W9-A3-008, W9-A3-009, W9-A3-010, W9-A3-011, W9-A3-012, W9-A3-013, W9-A3-018, W9-A3-019, W9-A4-007, W9-A4-008, W9-A5-008, W10-A2-019, W10-A3-005, W10-A3-006, W10-A3-008 | `frontend/src/components/messenger/message-item.tsx:546` | **да** — axe/`getByRole` name пустой на Send/Pin/Search; добавить `aria-label` → name появляется; убрать → FAIL | `fe-p2-a11y-accessible-names` |
| P2-FE-04 | P2 | Dialog/tabs/combobox/nested interactive + Button default `type` | W9-A1-004, W9-A1-009, W9-A4-003, W9-A5-002, W9-A5-004, W9-A5-005, W10-A3-007, W10-A3-012, W10-A4-008 | `frontend/src/components/ui/button.tsx:43` | **да** — `Button` в `<form>` без `type` → submit; с default `type="button"` → не submit; button⊂Link → axe nested-interactive FAIL | `fe-p2-a11y-dialogs-structure` |
| P2-FE-05 | P2 | Raw Tailwind colors / динамический JIT (`bg-${color}`) | W9-A3-014, W9-A3-015, W9-A3-016, W9-A3-017, W10-A3-014, W10-A5-009 | `frontend/src/routes/_authenticated/calls.tsx:117` | **да** — динамический `bg-${s.color}-500` → computed style без фона; статический map → класс в DOM/CSS | `fe-p2-tokens-tailwind-jit` |
| P2-FE-06 | P2 | Dead UI / no-op клики / фейковые KPI / фильтр не читается | W9-A1-003, W10-A2-012, W10-A3-013, W10-A3-016, W10-A4-004, W10-A5-004, W10-A5-010, W10-A5-011, W10-A5-012 | `frontend/src/routes/_auth/board/index.lazy.tsx:239` | **да** — чип «Мои задачи» без filter → список не меняется; добавить assignee filter → чужие скрыты; dead `onClick` отсутствует → клик no-op | `fe-p2-dead-ui-noop` |
| P2-FE-07 | P2 | Timer/fetch race / Rules of Hooks / setState-in-render / Plate stale / imperative poll | W9-A1-005, W9-A1-011, W9-A3-005, W9-A3-006, W9-A4-005, W10-A1-010, W10-A2-013, W10-A3-009, W10-A5-008 | `frontend/src/components/messenger/sticker-picker.tsx:198` | **да** — `useState` после early return → hooks lint/runtime FAIL; timer без cleanup → setState after unmount warning; plate value prop change без remount → stale JSON; poker `setInterval`+empty catch → silent fail | `fe-p2-effects-hooks-race` |
| P2-FE-08 | P2 | API DTO gaps / type-lie / unsafe cast / missing client methods | W9-A5-007, W10-A2-017, W10-A2-018, W10-A2-020, W10-A5-013, W11-A1-005, W11-A2-006, W11-A2-007, W11-A2-008, W11-A2-010 | `frontend/src/api/bases.ts:74` | **да** — `updateBase` без `version` → вызов проходит TS; сделать `version: number` required → compile fail на omit; guest `as CallResponse` → узкий тип ловит пустые поля | `fe-p2-api-dto-gaps` |
| P2-FE-09 | P2 | Auth session local: orphaned refresh + remote logout re-broadcast | W11-A1-007, W11-A2-004 | `frontend/src/api/client.ts:84` | **да** — failed refresh → нет `POST /auth/logout` (spy); добавить best-effort logout → spy called; remote logout → `clearSessionLocal` без второго broadcast | `fe-p2-auth-session-local` |
| P2-FE-10 | P2 | WS handlers: payload без null/object/field guard | W11-A3-010, W11-A3-011, W11-A3-015 | `frontend/src/hooks/use-messages.ts:51` | **да** — `payload: null` → TypeError в listener; guard → early return, соседние handlers живы | `fe-p2-ws-payload-guards` |
| P2-FE-11 | P2 | Dropzone: dragCounter на dragover + global paste hijack | W11-A3-006, W11-A3-007 | `frontend/src/hooks/use-file-upload.ts:100` | **да** — N dragover → counter ≫ enter; leave не сбрасывает `isDragging`; paste в input → upload Dropzone; fix → paste только на zone | `fe-p2-upload-drag-paste` |
| P2-FE-12 | P2 | Docs drift: counts, wiring line-refs, SSO/Telegram, onboarding steps | W11-A1-006, W11-A4-009, W11-A4-010, W11-A4-011, W11-A4-012, W11-A4-013, W11-A4-014, W11-A4-015, W11-A5-007, W11-A5-008, W11-A5-009, W11-A5-010, W11-A5-011, W11-A5-012, W11-A5-013, W11-A5-014, W11-A5-015, W11-A5-016, W11-A5-017, W11-A5-018, W11-A5-019 | `docs/TROUBLESHOOTING.md:21` | **да** — вернуть «515 / 39 hooks» → расхождение с `go list`/`export_hooks` count; канон 580/36 → grep docs зелёный; Telegram как login → assert «notify bind» | `fe-p2-docs-metrics-drift` |
| P2-FE-13 | P2 | Shell/sidebar: fake Online, collapse lock, mobile menu, intranet tab URL | W9-A2-006, W9-A2-007, W9-A2-008, W10-A1-016 | `frontend/src/components/layout/sidebar-user.tsx:39` | **да** — всегда «Онлайн» без presence → ложный claim; mobile nav без reset on route → overlay остаётся; `?tab=` без sync → back/forward ломает вкладку | `fe-p2-shell-nav-ux` |
| P2-FE-14 | P2 | Tracker residuals: attachments=[], soft-delete copy, queryKey/CI limit | W10-A2-011, W10-A2-015, W10-A4-005, W10-A4-006 | `frontend/src/routes/_auth/issues/$issueKey.tsx:210` | **да** — `attachments={[]}` → вложения с API не видны; передать `issue.attachments` → список; текст «нельзя отменить» → заменить на «корзина 30д» и snapshot | `fe-p2-tracker-attachments-copy` |
| P2-FE-15 | P2 | i18n relative time + push «Позже» forever + dead `useRooms` | W11-A4-005, W11-A4-006, W11-A4-007, W11-A4-008 | `frontend/src/lib/format-time.ts:18` | **да** — locale=en → всё равно RU «Только что»; future date → «Только что»; dismissLater без TTL → banner never; TTL → banner через N дней | `fe-p2-i18n-push-rooms` |
| P2-FE-16 | P2 | Authz UI: мутации всегда видны viewer/guest | W10-A2-014 | `frontend/src/modules/tms/webhooks.tsx:89` | **да** — role=viewer → create/delete webhook скрыты/disabled; убрать guard → кнопки видны и бьют 403 | `fe-p2-authz-ui-hide` |

**Классы с мутацией «да»: 16** (P2-FE-01…16).  
**Всего уникальных классов: 20** (= 16 + 4 в секции уточнения).

---

## Требует уточнения

Находки/классы, для которых **нельзя** завести TASK по правилу 57 (нет однозначного критерия мутации / нужен ADR, BE или product-scope). После уточнения — вернуть в очередь задач.

| ClassID | Source IDs | Суть | Чего не хватает для мутации / задачи |
|---------|------------|------|--------------------------------------|
| P2-FE-U01 | W11-A2-005 | Access+refresh JWT в `sessionStorage` (`hub-auth`) | ADR: httpOnly cookie refresh vs memory-only access. Без решения «минимум не persist refresh» ≠ полный фикс XSS |
| P2-FE-U02 | W11-A3-013 | Access JWT в `/ws?token=` query | Нужен BE: subprotocol / first-message auth. FE-only «не логировать» не закрывает утечку в access log |
| P2-FE-U03 | W9-A2-010 | Bookmarks UI показывает сырые UUID | BE: preview text/channel в DTO **или** product «оставить ids». Без контракта FE-fetch по id — отдельная кросс-задача |
| P2-FE-U04 | W11-A2-009 | `PaginatedResponse`/TZ offset vs tracker cursor | Решение PO/ADR: выровнять FE+BE на offset **или** поправить `tz_constraints`/комментарий под cursor. Иначе scope бесконечен |

Покрытие сырых ID: все **165** входят ровно в один класс из P2-FE-01…16 или P2-FE-U01…U04.

Проверка:

| Группа | Σ сырых |
|--------|--------:|
| FE-01…02 (error empty + mutation feedback) | 50 |
| FE-03…04 (a11y names + dialogs/structure) | 36 |
| FE-05…07 (tokens / dead UI / effects) | 24 |
| FE-08…11 (API / auth local / WS / upload) | 17 |
| FE-12…16 (docs / shell / tracker / i18n / authz UI) | 34 |
| **Σ с мутацией** | **161** |
| U01…U04 | 4 |
| **Итого** | **165** |

---

## Предлагаемая группировка в задачи бэклога (~10)

Не создавать файлы сейчас — только карта ClassID → будущий TASK slug. Нумерация TASK после закрытия P1 FE-волны (свободный слот — у team-lead).

| # | Suggested slug | Priority | ClassIDs | Сырых ID (оценка) | Заметка |
|---|---------------|----------|----------|------------------:|---------|
| 1 | `fe-p2-error-vs-empty` | P2 | FE-01 | 40 | Единый паттерн `QueryStateView` / `isError`+retry по modules/routes/admin/stores |
| 2 | `fe-p2-mutation-feedback` | P2 | FE-02 | 10 | ConfirmDialog + toast/rollback на mutate |
| 3 | `fe-p2-a11y-names-dialogs` | P2 | FE-03, FE-04 | 36 | aria-label sweep + dialog/tabs/Button type |
| 4 | `fe-p2-tokens-dead-ui` | P2 | FE-05, FE-06, FE-16 | 16 | `--hub-*` / JIT map + убрать/починить no-op + authz hide |
| 5 | `fe-p2-effects-hooks-race` | P2 | FE-07 | 9 | cleanup timers, hooks order, Plate controlled, poker poll |
| 6 | `fe-p2-api-dto-gaps` | P2 | FE-08 | 10 | version required, missing CRUD wrappers, type guards |
| 7 | `fe-p2-auth-ws-upload` | P2 | FE-09, FE-10, FE-11 | 7 | local session + WS guards + dropzone; **U01/U02** — после ADR/BE |
| 8 | `fe-p2-docs-metrics-drift` | P2 | FE-12 | 21 | counts, wiring refs, SSO/Telegram wording (docs-only) |
| 9 | `fe-p2-shell-tracker-ux` | P2 | FE-13, FE-14 | 8 | sidebar/mobile/intranet tab + attachments/copy/queryKey |
| 10 | `fe-p2-i18n-push-rooms` | P2 | FE-15 | 4 | format-time locale + push TTL + useRooms |

**Не в задачах до уточнения:** U01, U02, U03, U04 (4 класса / 4 сырых ID).

---

## Инвентарь сырых P2 (трассировка по волнам)

### W9 — components (52)

| ID | Кратко | → |
|----|--------|---|
| W9-A1-003 | Invite button no-op | FE-06 |
| W9-A1-004 | Custom modal без dialog ARIA | FE-04 |
| W9-A1-005 | setTimeout toast без cleanup | FE-07 |
| W9-A1-006 | ToggleRow без role=switch | FE-03 |
| W9-A1-007 | CallChat Send без aria-label | FE-03 |
| W9-A1-008 | Widget menu без aria-label | FE-03 |
| W9-A1-009 | dialog без aria-modal | FE-04 |
| W9-A1-010 | mail-settings error→empty | FE-01 |
| W9-A1-011 | audit-log fetch race | FE-07 |
| W9-A1-012 | ai-settings error→empty | FE-01 |
| W9-A1-013 | onec-settings error→empty | FE-01 |
| W9-A2-004 | compose Link без aria-label | FE-03 |
| W9-A2-005 | collapsed unread color-only | FE-03 |
| W9-A2-006 | fake Online | FE-13 |
| W9-A2-007 | active project нельзя свернуть | FE-13 |
| W9-A2-008 | mobile menu не закрывается | FE-13 |
| W9-A2-009 | bookmarks no isError | FE-01 |
| W9-A2-010 | bookmarks UUID only | **U03** |
| W9-A2-011 | dropzone без role/label | FE-03 |
| W9-A2-012 | file input без label | FE-03 |
| W9-A2-013 | language switcher aria-pressed | FE-03 |
| W9-A2-014 | charts без text alternative | FE-03 |
| W9-A2-015 | priority color-only | FE-03 |
| W9-A3-004 | toolbar opacity keyboard | FE-03 |
| W9-A3-005 | sticker hooks order | FE-07 |
| W9-A3-006 | gif debounce leak | FE-07 |
| W9-A3-007…013 | messenger icon a11y | FE-03 |
| W9-A3-014…017 | raw purple/blue/gray | FE-05 |
| W9-A3-018 | EditMode focus/aria | FE-03 |
| W9-A3-019 | reaction emoji names | FE-03 |
| W9-A4-003 | JQL combobox ARIA | FE-04 |
| W9-A4-004 | saved-filters silent catch | FE-01 |
| W9-A4-005 | Plate value stale | FE-07 |
| W9-A4-006 | telegram status catch→unlinked | FE-01 |
| W9-A4-007 | bell unread в aria-label | FE-03 |
| W9-A4-008 | share/trash icon names | FE-03 |
| W9-A5-002 | Button default type | FE-04 |
| W9-A5-003 | confirm closes on reject | FE-02 |
| W9-A5-004 | AlertDialog Description | FE-04 |
| W9-A5-005 | CommandDialog aria conflict | FE-04 |
| W9-A5-006 | user-picker error→empty | FE-01 |
| W9-A5-007 | user-picker enabled projectId | FE-08 |
| W9-A5-008 | form-field aria-describedby | FE-03 |

### W10 — routes/modules (64)

| ID range / ID | Кратко | → |
|---------------|--------|---|
| W10-A1-004…009,011…014,017,018,020 | modules fail→empty | FE-01 |
| W10-A1-010 | api-keys copied timer | FE-07 |
| W10-A1-015,019,021 | recording/coffee/moderation silent | FE-02 |
| W10-A1-016 | intranet tab URL sync | FE-13 |
| W10-A2-004…010,016 | TMS/reports fail→empty | FE-01 |
| W10-A2-011 | flaky queryKey | FE-14 |
| W10-A2-012 | create-suite no-op | FE-06 |
| W10-A2-013 | autotest copied timer | FE-07 |
| W10-A2-014 | mutations always visible | FE-16 |
| W10-A2-015 | CI cases limit 500 | FE-14 |
| W10-A2-017,018,020 | filter any / Plate cast / error cast | FE-08 |
| W10-A2-019 | coverage row keyboard | FE-03 |
| W10-A3-003,004 | velocity/backlog empty | FE-01 |
| W10-A3-005,006,008 | sprint a11y labels | FE-03 |
| W10-A3-007,012 | nested Link/tabs | FE-04 |
| W10-A3-009 | poker interval+store silent | FE-07 |
| W10-A3-010,011,015 | preset/retro/instant silent | FE-02 |
| W10-A3-013,016 | fake KPI / dead search | FE-06 |
| W10-A3-014 | calls JIT colors | FE-05 |
| W10-A4-004 | myTasks filter no-op | FE-06 |
| W10-A4-005,006 | soft-delete copy / attachments=[] | FE-14 |
| W10-A4-007 | routes error→empty | FE-01 |
| W10-A4-008 | ChatErrorBoundary retry | FE-04 |
| W10-A5-004,010…012 | trash/onboarding dead | FE-06 |
| W10-A5-005…007 | TMS detail error→shimmer/empty | FE-01 |
| W10-A5-008 | profile setState-in-render | FE-07 |
| W10-A5-009 | projects JIT colors | FE-05 |
| W10-A5-013 | guest CallResponse type-lie | FE-08 |
### W11 — api/stores/hooks/lib + docs (49)

| ID | Кратко | → |
|----|--------|---|
| W11-A1-003,004 | getActiveCall / scheduled swallow | FE-01 |
| W11-A1-005 | bases version optional | FE-08 |
| W11-A1-006 | ApiError vs docs `error` | FE-12 |
| W11-A1-007 | refresh fail без logout POST | FE-09 |
| W11-A2-002,003 | ai-store loaded on error | FE-01 |
| W11-A2-004 | logout re-broadcast | FE-09 |
| W11-A2-005 | JWT sessionStorage | **U01** |
| W11-A2-006…008,010 | missing API methods / WS cast | FE-08 |
| W11-A2-009 | cursor vs offset TZ | **U04** |
| W11-A3-006,007 | dragCounter / paste | FE-11 |
| W11-A3-008,009 | hand/layout silent | FE-02 |
| W11-A3-010,011,015 | WS payload guards | FE-10 |
| W11-A3-012 | presence no isError | FE-01 |
| W11-A3-013 | WS token query | **U02** |
| W11-A3-014 | sprint-store no error | FE-01 |
| W11-A4-004 | markRead no rollback | FE-02 |
| W11-A4-005…008 | i18n / push / useRooms | FE-15 |
| W11-A4-009…015 | architecture/BUSINESS/style drift | FE-12 |
| W11-A5-007…019 | counts/wiring/SSO/onboarding docs | FE-12 |

---

## Handoff

- Следующий шаг W12 (другие агенты): P3-сводка FE/BE — по назначениям оркестратора.
- Team-lead: завести TASK **только** по строкам таблицы группировки с мутацией «да»; U\* — в backlog README как «Blocked: needs ADR/BE/PO».
- Исполнение: после green P0/P1 FE; порядок — error-vs-empty и mutation-feedback раньше cosmetic a11y/docs; authz-ui и API DTO — парно с соответствующими BE P1/P2.
- Не путать с P1 FE: CallChat WS / board DnD / api-client 401 — уже в [`W12-REGISTRY-P1-FRONTEND.md`](W12-REGISTRY-P1-FRONTEND.md).

---

*Артефакт W12 Stage B · read-only по product code · 2026-07-28*
