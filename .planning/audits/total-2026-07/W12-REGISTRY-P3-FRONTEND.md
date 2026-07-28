# W12 — реестр уникальных P3-классов (frontend, волны 9–11)

> TASK-008 · Стадия B · W12 · 2026-07-28  
> Источники: `A{9-11}-*.md`, сводки `W9-SUMMARY`…`W11-SUMMARY`, [`00-INDEX.md`](00-INDEX.md)  
> Scope: **только frontend** (W9–W11) + docs, попавшие в W11 A4/A5.  
> Backend W1–W8 — вне артефакта (отдельный P3-backend, если заведён оркестратором).  
> **TASK-\*.md не создаются** — только реестр + предложение группировки.  
> Формат: зеркало [`W12-REGISTRY-P2-FRONTEND.md`](W12-REGISTRY-P2-FRONTEND.md) / [`W12-REGISTRY-P1-FRONTEND.md`](W12-REGISTRY-P1-FRONTEND.md).

## Сводка

| Метрика | Значение |
|---------|---------:|
| Сырых P3 (W9–W11) | **81** |
| Уникальных классов | **17** |
| Из них с критерием мутации «да» | **16** |
| В секции «требует уточнения» | **1** |
| Предлагаемых задач бэклога | **10** |

Сырые по волнам: W9=47 · W10=20 · W11=14 → **81**

P0/P1/P2 FE **не** в таблице: [`W12-REGISTRY-P1-FRONTEND.md`](W12-REGISTRY-P1-FRONTEND.md), [`W12-REGISTRY-P2-FRONTEND.md`](W12-REGISTRY-P2-FRONTEND.md) (P0 FE = 0).

---

## Уникальные классы дефектов

| ClassID | Severity | Title | Source IDs | Primary file:line | Mutation criterion | Suggested TASK slug |
|---------|----------|-------|------------|-------------------|--------------------|---------------------|
| P3-FE-01 | P3 | Icon-only / inputs: `title` only или нет `aria-label`/`Label` | W9-A1-016, W9-A1-017, W9-A3-024, W9-A3-025, W9-A3-026, W9-A3-027, W9-A3-028, W9-A3-029, W9-A3-030, W9-A3-038, W9-A3-039, W9-A4-009, W9-A4-010, W9-A4-011, W10-A2-022, W10-A2-023, W10-A4-011 | `frontend/src/components/calls/CallPanel.tsx:55` | **да** — `getByRole('button', { name: … })` пустой на fullscreen/реакциях/GIF search; добавить `aria-label`/`Label` → name появляется; убрать → FAIL | `fe-p3-a11y-names` |
| P3-FE-02 | P3 | a11y structure: tabs/semantics/декор. иконки/`aria-required` | W9-A1-018, W9-A4-013, W9-A5-012, W9-A5-013, W9-A5-014, W9-A5-015, W9-A5-016, W9-A5-017, W9-A5-019 | `frontend/src/components/admin/trash-manager.tsx:76` | **да** — вкладки без `tablist`/`aria-selected` → axe tab FAIL; `CardTitle`=`div` → нет heading; decorative lucide без `aria-hidden` → лишний AT noise; `aria-required` отсутствует на `*` → добавить → AT видит required | `fe-p3-a11y-structure` |
| P3-FE-03 | P3 | Raw Tailwind palette / shadcn defaults ≠ `--hub-*` | W9-A3-031, W9-A3-032, W9-A3-033, W9-A3-034, W9-A3-035, W9-A5-018 | `frontend/src/components/messenger/voice-recorder.tsx:186` | **да** — `bg-red-500`/`text-orange-500` в DOM; заменить на `var(--red)`/`var(--orange)` → raw palette grep пуст; Select `rounded-md`/`border-input` → hub tokens | `fe-p3-tokens-palette` |
| P3-FE-04 | P3 | Hardcoded RU/EN в ui-примитивах (обход i18n) | W9-A5-009, W9-A5-010, W9-A5-011 | `frontend/src/components/ui/dialog.tsx:46` | **да** — locale=en → Dialog Close всё равно «Close» EN / Command «Глобальный поиск» RU; ключи i18n → строки из catalog; EB fallback без i18n → snapshot RU-only | `fe-p3-i18n-primitives` |
| P3-FE-05 | P3 | `setTimeout` copied/saved без cleanup | W9-A1-014, W9-A2-017, W10-A3-018 | `frontend/src/components/calls/CallRoom.tsx:203` | **да** — unmount до timeout → React «setState on unmounted»; `clearTimeout` в cleanup → warning исчезает; убрать cleanup → warning снова | `fe-p3-timer-cleanup` |
| P3-FE-06 | P3 | Silent fail / loading=empty (P3-tier) | W9-A1-015, W9-A1-019, W9-A2-018, W9-A3-037, W10-A2-024 | `frontend/src/components/ai/byok-settings.tsx:23` | **да** — BYOK/CallChat reject → empty «нет данных»; добавить `isError`/toast → error UI; poll vote catch silent → toast spy; bookmark delete без try → toast на reject | `fe-p3-silent-fail-empty` |
| P3-FE-07 | P3 | Unsafe cast / type-lie / missing generics | W9-A2-016, W9-A3-020, W9-A3-021, W9-A3-022, W9-A4-014, W10-A2-026, W11-A1-010, W11-A2-014, W11-A4-016 | `frontend/src/components/messenger/message-composer.tsx:318` | **да** — `as unknown as Record` на Plate → узкий `PlateDoc` ловит omit; marketplace `api.get` без `<T>` → снять annotation → `unknown` compile fail; `extractApiError` глотает `Error` → guard `request_id`/`code` | `fe-p3-type-casts` |
| P3-FE-08 | P3 | Dead UI / noop affordance / мёртвый state | W9-A2-019, W9-A3-023, W9-A4-012, W10-A2-025, W10-A3-019, W10-A4-009, W10-A4-010, W10-A5-014, W10-A5-016, W10-A5-017, W11-A3-016 | `frontend/src/routes/_auth/issues/index.tsx:53` | **да** — `statusFilter` в state без Select → reset трогает невидимое; Esc mention не закрывает список; avatar `cursor-pointer` без open file → клик no-op; dead `placeholder` NavItem / tms-store fields → удалить → coverage dead-code | `fe-p3-dead-ui-state` |
| P3-FE-09 | P3 | Perf: always-poll / stale clock / O(n²) Set | W10-A3-017, W10-A3-020, W10-A5-019 | `frontend/src/modules/tracker/pipeline-status.tsx:200` | **да** — `refetchInterval` всегда 30s без in_progress → лишние fetch (spy); interval только на active → spy тише; `formatTime` в `useMemo([])` → часы застыли; `new Set` внутри `.map` → вынести до map | `fe-p3-perf-stale` |
| P3-FE-10 | P3 | API client: `ApiError` / JSON parse resilience | W11-A1-008, W11-A1-009 | `frontend/src/api/client.ts:112` | **да** — 2xx invalid JSON → сырой `SyntaxError`; try/catch → `ApiError` code 502; calendar export `throw new Error` → нет `request_id` в toast; `ApiError` → toast видит code | `fe-p3-api-error-parse` |
| P3-FE-11 | P3 | Auth store: login broadcast `userId` + storage fallback | W11-A2-011, W11-A2-012 | `frontend/src/stores/auth-store.ts:112` | **да** — `setTokens` → event `userId: null` всегда; передать userId → соседняя вкладка различает same-user; fallback `storage` без mirror write → dead path; mirror localStorage **или** удалить listener | `fe-p3-auth-broadcast` |
| P3-FE-12 | P3 | Duplicate API / `use-*` naming drift | W11-A2-013, W11-A3-017 | `frontend/src/api/github.ts:183` | **да** — два `/link-preview` клиента → расхождение полей; один модуль → импорт единый; `use-composer-ai-suggest` без hook → перенос в `lib/` **или** реальный `use*` | `fe-p3-api-naming-dup` |
| P3-FE-13 | P3 | Docs metrics/pins drift (P3) | W11-A4-017, W11-A5-020, W11-A5-021 | `docs/architecture.md:105` | **да** — «~60+ UI pages» / Zustand «11+» / curl `:3000` → расхождение с 72 routes / 12 stores / Caddy `:8080`; канон → grep docs зелёный | `fe-p3-docs-metrics` |
| P3-FE-14 | P3 | SPA `Link` + misleading column / mojibake | W10-A2-021, W10-A5-015, W10-A5-018 | `frontend/src/routes/_auth/settings/security.tsx:58` | **да** — `<a href="/settings/profile">` → full reload; `<Link>` → SPA nav; trash «Удалил» показывает дату → rename «Удалено»; «Испол��итель» → «Исполнитель» snapshot | `fe-p3-spa-copy-utf8` |
| P3-FE-15 | P3 | Link-preview URLs: `contentText` ≠ `displayText` | W10-A1-022 | `frontend/src/modules/messenger/message-bubble.tsx:117` | **да** — URL только в Plate JSON → preview пуст при `extractUrls(contentText)`; `extractUrls(displayText)` / union → preview появляется | `fe-p3-link-preview-urls` |
| P3-FE-16 | P3 | Timer store: битый `startedAt` → `NaN` elapsed | W11-A3-018 | `frontend/src/stores/timer-store.ts:56` | **да** — persist `startedAt: "bad"` → UI `NaN`; `Number.isFinite` guard → `stopTimer()` / 0; убрать guard → NaN снова | `fe-p3-timer-nan-guard` |

**Классы с мутацией «да»: 16** (P3-FE-01…16).  
**Всего уникальных классов: 17** (= 16 + 1 в секции уточнения).

---

## Требует уточнения

Находки/классы, для которых **нельзя** завести TASK по правилу 57 (нет однозначного критерия мутации / нужен ADR, BE или product-scope). После уточнения — вернуть в очередь задач.

| ClassID | Source IDs | Суть | Чего не хватает для мутации / задачи |
|---------|------------|------|--------------------------------------|
| P3-FE-U01 | W9-A3-036 | Pinned messages показывают `ID: {message_id}` (UUID) | BE: `content_text`/preview в pins DTO **или** product «оставить ids». Без контракта FE-only truncate UUID — косметика, не полноценный preview (аналог P2-FE-U03 bookmarks) |

Покрытие сырых ID: все **81** входят ровно в один класс из P3-FE-01…16 или P3-FE-U01.

Проверка:

| Группа | Σ сырых |
|--------|--------:|
| FE-01…02 (a11y names + structure) | 26 |
| FE-03…04 (tokens + i18n) | 9 |
| FE-05…06 (timers + silent fail) | 8 |
| FE-07…08 (casts + dead UI) | 20 |
| FE-09…12 (perf / api / auth / naming) | 9 |
| FE-13…16 (docs / spa / preview / timer NaN) | 8 |
| **Σ с мутацией** | **80** |
| U01 | 1 |
| **Итого** | **81** |

---

## Предлагаемая группировка в задачи бэклога (~10)

Не создавать файлы сейчас — только карта ClassID → будущий TASK slug. Нумерация TASK после закрытия P1/P2 FE-волн (свободный слот — у team-lead).

| # | Suggested slug | Priority | ClassIDs | Сырых ID (оценка) | Заметка |
|---|---------------|----------|----------|------------------:|---------|
| 1 | `fe-p3-a11y-names` | P3 | FE-01 | 17 | Sweep `aria-label`/`Label` на calls/messenger/TMS icon controls |
| 2 | `fe-p3-a11y-structure` | P3 | FE-02 | 9 | tablist, CardTitle heading, decorative `aria-hidden`, `aria-required`, alert role |
| 3 | `fe-p3-tokens-i18n` | P3 | FE-03, FE-04 | 9 | `--hub-*` / semantic colors + i18n keys в dialog/command/EB |
| 4 | `fe-p3-timers-cleanup` | P3 | FE-05, FE-16 | 4 | setTimeout cleanup + timer-store NaN guard |
| 5 | `fe-p3-silent-fail-empty` | P3 | FE-06 | 5 | BYOK/CallChat/bookmarks/poll/issue-links: isError + toast |
| 6 | `fe-p3-type-casts` | P3 | FE-07 | 9 | PlateDoc, widget union, marketplace generics, ApiError guard |
| 7 | `fe-p3-dead-ui-state` | P3 | FE-08 | 11 | statusFilter, bulk panel, cursor-pointer, dead store fields |
| 8 | `fe-p3-perf-stale` | P3 | FE-09 | 3 | pipeline interval, dashboard clock, Set hoist |
| 9 | `fe-p3-api-auth-naming` | P3 | FE-10, FE-11, FE-12 | 6 | ApiError parse + auth broadcast + link-preview/`use-*` |
| 10 | `fe-p3-docs-spa-ux` | P3 | FE-13, FE-14, FE-15 | 7 | docs pins + Link/mojibake/trash copy + preview URLs |

**Не в задачах до уточнения:** U01 (1 класс / 1 сырой ID).

---

## Инвентарь сырых P3 (трассировка по волнам)

### W9 — components (47)

| ID | Кратко | → |
|----|--------|---|
| W9-A1-014 | InviteBlock setTimeout no cleanup | FE-05 |
| W9-A1-015 | BYOK refresh silent catch | FE-06 |
| W9-A1-016 | CallPanel fullscreen title-only | FE-01 |
| W9-A1-017 | CallReactions title-only | FE-01 |
| W9-A1-018 | trash-manager tabs no tablist | FE-02 |
| W9-A1-019 | CallChat no loading/error UI | FE-06 |
| W9-A2-016 | widgets `data?: unknown` cast | FE-07 |
| W9-A2-017 | bot-markdown copy timeout | FE-05 |
| W9-A2-018 | bookmarks delete no try/catch | FE-06 |
| W9-A2-019 | deadlines rows not clickable | FE-08 |
| W9-A3-020…022 | Plate/sticker/render casts | FE-07 |
| W9-A3-023 | mention Esc не закрывает | FE-08 |
| W9-A3-024…030 | messenger a11y labels/tabs | FE-01 |
| W9-A3-031…035 | raw orange/red/yellow | FE-03 |
| W9-A3-036 | pins UUID only | **U01** |
| W9-A3-037 | poll dead channelId + silent | FE-06 |
| W9-A3-038,039 | attachment/invite icon names | FE-01 |
| W9-A4-009 | mobile-nav «Ещё» aria-expanded | FE-01 |
| W9-A4-010 | quiet hours Input без Label | FE-01 |
| W9-A4-011 | checklist × title-only | FE-01 |
| W9-A4-012 | attachments «Перетащите» no DnD | FE-08 |
| W9-A4-013 | route-error без role=alert | FE-02 |
| W9-A4-014 | `as unknown as` DTO casts | FE-07 |
| W9-A5-009…011 | Dialog/Command/EB i18n | FE-04 |
| W9-A5-012…017,019 | avatar/card/badge/icons/required | FE-02 |
| W9-A5-018 | Select shadcn token drift | FE-03 |

### W10 — routes/modules (20)

| ID | Кратко | → |
|----|--------|---|
| W10-A1-022 | extractUrls(contentText) vs displayText | FE-15 |
| W10-A2-021 | mojibake «Исполнитель» | FE-14 |
| W10-A2-022,023 | step/case icon a11y | FE-01 |
| W10-A2-024 | issue-message-links silent | FE-06 |
| W10-A2-025 | stepResults dead state | FE-08 |
| W10-A2-026 | env config `[object Object]` | FE-07 |
| W10-A3-017 | pipeline always poll | FE-09 |
| W10-A3-018 | sprint-retro setTimeout | FE-05 |
| W10-A3-019 | sidebar meta fake click | FE-08 |
| W10-A3-020 | dashboard clock useMemo[] | FE-09 |
| W10-A4-009 | statusFilter dead | FE-08 |
| W10-A4-010 | bulk select no actions | FE-08 |
| W10-A4-011 | color swatch no aria | FE-01 |
| W10-A5-014 | settings placeholder dead | FE-08 |
| W10-A5-015 | security `<a>` full reload | FE-14 |
| W10-A5-016 | avatar cursor-pointer no-op | FE-08 |
| W10-A5-017 | department/duration dead state | FE-08 |
| W10-A5-018 | trash «Удалил» = date | FE-14 |
| W10-A5-019 | Set inside map O(n²) | FE-09 |

### W11 — api/stores/hooks/lib + docs (14)

| ID | Кратко | → |
|----|--------|---|
| W11-A1-008 | calendar export plain Error | FE-10 |
| W11-A1-009 | JSON.parse 2xx no try | FE-10 |
| W11-A1-010 | AIProvider type duplicate | FE-07 |
| W11-A2-011 | login broadcast userId null | FE-11 |
| W11-A2-012 | storage fallback dead | FE-11 |
| W11-A2-013 | github link-preview dup | FE-12 |
| W11-A2-014 | marketplace api no `<T>` | FE-07 |
| W11-A3-016 | tms-store dead server fields | FE-08 |
| W11-A3-017 | use-composer-ai-suggest no hook | FE-12 |
| W11-A3-018 | timer NaN from persist | FE-16 |
| W11-A4-016 | extractApiError loose guard | FE-07 |
| W11-A4-017 | architecture «~60+» routes | FE-13 |
| W11-A5-020 | stack stores/motion/pwa pins | FE-13 |
| W11-A5-021 | API.md curl :3000 vs Caddy | FE-13 |

---

## Handoff

- Следующий шаг W12 (другие агенты): P3-сводка backend — по назначениям оркестратора.
- Team-lead: завести TASK **только** по строкам таблицы группировки с мутацией «да»; U01 — в backlog README как «Blocked: needs BE/PO» (после pins DTO).
- Исполнение: после green P0/P1/P2 FE; порядок — silent-fail и type-casts раньше cosmetic a11y/docs; a11y-names можно параллелить с tokens-i18n.
- Не путать с P2 FE: error-vs-empty / a11y-names P2 / docs-metrics P2 — уже в [`W12-REGISTRY-P2-FRONTEND.md`](W12-REGISTRY-P2-FRONTEND.md); здесь **остаточный** P3-хвост (title-only, shadcn drift, perf, store quirks).

---

*Артефакт W12 Stage B · read-only по product code · 2026-07-28*
