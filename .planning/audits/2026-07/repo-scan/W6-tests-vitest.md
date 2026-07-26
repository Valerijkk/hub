# W6-tests-vitest — QA-auto audit Vitest / frontend-test

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** qa-auto-ver (Grok 4.5), readonly  
> **Дата:** 2026-07-23  
> **Scope:** `testframework/stages/unit/frontend-test/` + конфиг Vitest (`frontend/vite.config.ts`, `frontend/src/test/setup.ts`)  
> **Код не менялся**

---

## Summary

Канон соблюдён: **37** файлов / **~99** `it()` в `domains/**`, **0** тестов в `frontend/src/**` (`exclude` в vite). `renderWithProviders` есть и корректен (свой `QueryClient`, без retry). Нет `waitForTimeout` / `test.skip` / Allure-секретов.

Главные пробелы по рубрике G.1: **нулевая двусторонняя привязка TC↔Vitest**, массовые «smoke render + mock called» на v4 без негативов, **дыры доменов** (calls / sprint / notification / messenger UI / 11/13 stores), обход `renderWithProviders` и `userEvent`. Живой прогон в этой сессии **не выполнен** (sandbox заблокировал link-скрипт) — критерий race/time **не закрыт**.

**Вердикт:** `verified: false` — 2026-07-23 01:15

---

## Inventory

| Метрика | Значение | Источник |
|---------|----------|----------|
| Файлы `*.test.{ts,tsx}` | **37** | glob `domains/**` |
| Хелпер | `render.tsx` | `@testframework/render` |
| `it()` (подсчёт) | **99** | совпадает с `docs/TESTING.md` / COVERAGE |
| Тесты в `frontend/src/` | **0** | `vite.config.ts` exclude |
| Домены в job YAML | auth, admin, ai, tracker, messenger, tms, dashboard, platform, v4 | `pipeline/jobs/frontend-test.yaml` |
| CI команды | `tsc` + `vitest run` + lint | тот же YAML |
| Coverage gate FE | **нет** (только BE ≥82%) | `vite.config.ts` без thresholds; COVERAGE.md |

### Покрытие доменов vs продукт

| Домен FE | Файлы | Оценка |
|----------|-------|--------|
| auth | 4 | store + client refresh single-flight + schema + hook — **лучше среднего** |
| admin | 4 | audit/metrics/api — **ок**, мало негативов |
| ai | 5 | store + privacy toggle + API — **ок** |
| tracker | 4 | card/form/schema/picker — **ок**, без board/DnD |
| tms | 1 | только `CaseTableRow` — **тонко** |
| messenger | 1 | только `ChannelIcon` — **критически тонко** |
| dashboard | 1 | `SystemHealth` — **тонко** |
| platform | 7 | error-boundary, format-api-error (fake timers) — **сильнее** |
| v4 (10 pages) | 10 | почти везде 1 happy-path — **smoke** |
| **calls** | **0** | модуль LiveKit UI без unit |
| **sprint / notif / search / settings shell** | **0** | нет доменов |
| Zustand stores | 2/13 | только `auth-store`, `ai-store` |

---

## Findings

| Severity | file:line | Issue | Evidence | Fix |
|----------|-----------|-------|----------|-----|
| **CRITICAL** | весь `domains/**` + `docs/TESTING.md` | Нет двусторонней привязки TC↔Vitest | `rg TC-[A-Z]+` / `AS_ID` / `allure` в frontend-test → **0 hits**. Колонка «Автотест» в TESTING.md указывает почти только Go `Test*` / E2E; Vitest-имена не фигурируют | Для компонентных сценариев: комментарий `// TC-…` над `it`/`describe` + колонка MD с точным именем файла/`describe > it`; либо явно пометить слой как «unit fe (вне TC-матрицы)» в TESTING.md |
| **HIGH** | `domains/v4/*-page.test.tsx`, `leave-requests`, `random-coffee`, `marketplace-page`, `crm-page`, … | Smoke-тесты: heading + `expect(api).toHaveBeenCalled()` без контракта UI/ошибок | Напр. `leave-requests.test.tsx:36-38`, `marketplace-page.test.tsx:40-42`, `random-coffee.test.tsx:41-43` — нет empty/error/RBAC/негатива | Минимум: empty state, API reject → error UI, один user-flow (submit/filter); не считать «mock called» достаточным контрактом |
| **HIGH** | отсутствие `domains/calls/**` | Критичный LiveKit UI без Vitest | `frontend/src/modules/calls/*` (10 файлов) — 0 тестов | fe/qa-auto2: unit на create-call dialog, recording controls (мок LiveKit), calendar views без WebRTC |
| **HIGH** | `messenger/channel-icon.test.tsx` only | Messenger UI почти не покрыт unit’ами | 4 теста иконки; нет списка каналов, composer, unread, readonly | Компонентные тесты ключевых виджетов + ссылка на TC-MSG-* где применимо |
| **MEDIUM** | `auth/client-api.test.ts:30` | `setTimeout(30)` в mock fetch для single-flight | Не `waitForTimeout`, но реальный wall-clock; на загруженном CI возможен flake | Переписать на `vi.useFakeTimers` + deferred Promise / `Promise.withResolvers` без таймера |
| **MEDIUM** | v4/admin pages vs `render.tsx` | `renderWithProviders` почти не используется | Только `skip-link`, `issue-form` (+ определение). Остальные дублируют `new QueryClient` + `QueryClientProvider` | Унифицировать на `renderWithProviders`; убрать копипасту QC |
| **MEDIUM** | все RTL interaction tests | Playbook §5.3 требует `userEvent`; везде `fireEvent` | audit-log, issue-form, ai-privacy, org-directory, … — `fireEvent` only; `userEvent` = 0 | Мигрировать клики/ввод на `@testing-library/user-event` |
| **MEDIUM** | `stores/*` | 11/13 Zustand stores без unit | Нет тестов: messenger, call, notification, board, sprint, tms, ws, timer, dashboard, command, sidebar | fe: unit на мутации/селекторы критичных stores (ws/call/messenger) |
| **MEDIUM** | `docs/TESTING.md:202` + FE слой | Матрица заявляет связь TC с «unit», но колонка = Go; FE unit вне traceability | Текст: «связан с автоматизированными unit-тестами» + Go functions | Развести: API/integration = Go; компонентные = Vitest имена; обновить intro |
| **MEDIUM** | `frontend/vite.config.ts:77-83` | Нет FE coverage thresholds / отчёта в CI job | Только `environment`, `include`, `exclude` | Опционально: `coverage.thresholds` на критичные модули **или** явный N/A в COVERAGE.md |
| **LOW** | `messenger/channel-icon.test.tsx:15-28`, `tms/case-table-row.test.tsx:85-92` | Хрупкие `container.querySelector` / title | svg без role/name; `[title="…"]` | `getByRole` / accessible name / `data-testid` при неоднозначности |
| **LOW** | `v4/status-page.test.tsx:13` | `new Date().toISOString()` в фикстуре | Не влияет на assert сейчас; риск при появлении snapshot/времени | Фиксированный ISO timestamp |
| **LOW** | v4 fixtures | `user: {…} as never` | leave/marketplace/random-coffee/news-feed | Типизированный stub User без `as never` |
| **INFO** | `platform/format-api-error.test.ts` | Fake timers + `afterEach` restore — **хороший паттерн** | `vi.useFakeTimers` / `advanceTimersByTime` / `useRealTimers` | Эталон для client-api single-flight |
| **INFO** | `auth/client-api.test.ts` | Осмысленный контракт: single-flight refresh, Authorization, logout on refresh fail | 3 сценария с реальными asserts | — |
| **INFO** | `vite` exclude `src/**/*.test.*` | Юниты живут только в testframework — pipeline-aligned | `include: ../testframework/...` | Держать |
| **PASS** | sleep/flaky sync | Нет `waitForTimeout`; нет `test.skip` | grep | — |
| **PASS** | Shared mutable QC | `renderWithProviders` создаёт новый QueryClient | `render.tsx:13-19,31` | — |
| **PASS** | Inventory docs | 37 files / 99 tests согласованы | TESTING.md, COVERAGE.md, YAML domains list | — |

---

## Рубрика G.1 (явно)

| # | Критерий | Результат |
|---|----------|-----------|
| 1 | Детерминизм | **PARTIAL** — нет waitForTimeout; есть `setTimeout(30)` в mock (MEDIUM) |
| 2 | Race / green run | **FAIL** — живой `vitest run` в сессии не выполнен (blocked); опираться на CI claim 99/99 нельзя для verified |
| 3 | Привязка TC | **FAIL** — 0 TC-ID в коде; MD→Vitest отсутствует |
| 4 | Осмысленность ассертов | **PARTIAL** — auth/platform/tracker сильнее; v4 smoke слабые |
| 5 | Покрытие | **FAIL/gap** — calls=0, messenger thin, stores 2/13; FE % не гейтится |
| 6 | Границы и негатив | **PARTIAL** — auth client + schemas да; v4 pages нет |
| 7 | Triage | **N/A** — код не чинили |
| 8 | Время | **FAIL** — замер не зафиксирован в этой сессии |
| 9 | Артефакты | **PASS** — dumps не создавались |
| 10 | Docs sync | **FAIL** — TESTING.md не отражает Vitest в колонке Автотест |

CRITICAL/HIGH по 3, 5 (дыры calls/messenger), smoke v4 → блокируют historical PASS → independent acceptance.

---

## 5 top risks

1. **Слепая зона TC↔FE unit** — регрессии UI не трассируются к ACCEPTANCE; qa-auto2 не может закрыть матрицу.
2. **calls без Vitest** — LiveKit/recording UI только за E2E (дорого/редко) → высокий риск регрессии SFU UX.
3. **v4 smoke-ложная уверенность** — 10 «зелёных» page-тестов не ловят empty/error/RBAC.
4. **messenger = иконка** — ядро продукта почти без компонентного слоя.
5. **setTimeout single-flight** — потенциальный flake refresh-логики auth на медленных агентах.

---

## Рекомендации (без правок в этой волне)

1. Добавить TC-комментарии + строки в TESTING.md для существующих сильных тестов (`client-api`, `issue-form`, `ai-privacy`, `error-boundary`).
2. Расширить v4: error + empty на каждый page-тест (минимальный DoD).
3. Завести `domains/calls/` (мок `@livekit/*`) и расширить `messenger/` сверх icon.
4. Заменить `setTimeout` в refresh-тесте на fake timers / latch.
5. Стандартизировать `renderWithProviders` + `userEvent`.
6. После правок: `cd frontend && npm run test:run` + зафиксировать wall-time в STATE/отчёте; при подозрении — `--repeat 10` на client-api.

---

## Verification

| Этап | Результат |
|------|-----------|
| Scope прочитан | `domains/**` (37), `render.tsx`, `vite.config.ts`, `setup.ts`, `frontend-test.yaml`, PIPELINE, TESTING.md, COVERAGE.md |
| Grep anti-flaky | `waitForTimeout`=0; `setTimeout`=1 (client-api); fakeTimers=format-api-error (OK); skip=0 |
| Grep TC/Allure | TC/AS_ID/allure = **0** |
| `userEvent` | **0**; `fireEvent` — несколько файлов |
| `renderWithProviders` usage | 2 тест-файла (+ определение) |
| Живой vitest | **не выполнен** (sandbox block link script) |
| Severity калибровка | traceability + domain holes = CRITICAL/HIGH, не занижены |

---

## Верификация (QA-auto): Vitest frontend-test suite

- Регрессия: `npx vitest run` — **не прогнан** в сессии (blocked), время **н/д**, лог **нет**
- Детерминизм/flaky: статический PASS по sleep; **подозрение** `client-api` setTimeout — повтор xN не делался
- Замечания: см. таблицу Findings (1 CRITICAL, 3 HIGH, …)
- Вердикт: **verified: false** — 2026-07-23 01:15

**Handoff:** после фикса traceability + calls/messenger gaps + живого green прогона → повторный `qa-auto-ver`; pre-merge diff — `code-reviewer`.
