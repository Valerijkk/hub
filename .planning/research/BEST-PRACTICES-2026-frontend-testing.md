# Hub — Best Practices 2026: Frontend + Testing (дельта-аудит playbooks)

> **Дата:** 2026-07-03 · **Тип:** research-отчёт (код и playbooks не менялись)
> **Сверялись:** `.cursor/playbooks/frontend.md` (21 раздел), `.cursor/playbooks/testing.md` (17 разделов + приложения), `docs/stack_and_tools.md`, факт-конфиги `frontend/tsconfig.json`, `frontend/biome.json`, `frontend/playwright.config.ts`.
> **Источники:** web-research 2026 — React 19.2 / React Compiler v1.0, TanStack Router+Query v5, TS 5.8–5.9, Tailwind 4, Playwright best practices 2026, Vitest 4 Browser Mode, Allure 3, Biome 2.x, Core Web Vitals (INP), WCAG 2.2 / EN 301 549 v4.1.1.

---

## 1. Резюме

**Всего дельт: 21** (7 — high, 9 — medium, 5 — low). Playbooks в целом свежие и сильные: дисциплина мемоизации «по причине», Rules of Hooks, RTL-приоритеты локаторов, web-first assertions без sleep, optimistic updates с rollback, MSW с `onUnhandledRequest: "error"`, trace `on-first-retry` — всё это соответствует состоянию 2026 года. Основные разрывы — не «устаревшее», а **отсутствующее**: новые API React 19, интеграция Router↔Query на уровне preload, WCAG 2.2, INP-метрики и CI-механика Playwright (шардинг/merge-reports).

### Топ-5 дельт

1. **WCAG 2.2 AA вместо 2.1 AA** — 2.2 уже стандарт (ISO/IEC 40500:2025; EN 301 549 v4.1.1 с WCAG 2.2 ожидается в 2026). Для Hub критичны: **2.5.7 Dragging Movements** (Kanban DnD обязан иметь альтернативу без drag), **2.5.8 Target Size ≥24×24px** (иконочные кнопки), **2.4.11 Focus Not Obscured** (sticky-панели). Magic Link уже удовлетворяет 3.3.8 Accessible Authentication.
2. **TanStack Router preloading не используется** — канон 2026: `defaultPreload: 'intent'` + `defaultPreloadStaleTime: 0` + `ensureQueryData` в loader + `useSuspenseQuery` в компоненте (queryClient через router context). Playbook описывает loader как «опционально» и не упоминает intent-preload — бесплатный prefetch-on-hover не задействован.
3. **Playwright CI-механика устарела относительно масштаба** — `workers: 1` в CI, нет шардинга (`--shard` + blob reporter + `merge-reports`), нет `test.step()` (нативные шаги в HTML/trace-отчёте), нет `storageState` (сессия сохраняется один раз в setup-project вместо dev-login в каждом тесте).
4. **React 19 Actions-триада отсутствует** — `useActionState`, `useOptimistic`, `useFormStatus` не упомянуты; React Compiler v1.0 (production-stable с октября 2025, проверен в Meta) не оценён — playbook пишет «компилятор может быть не включён», пора принять явное решение (пилот на одном route-subtree).
5. **Perf-раздел не знает про INP и Core Web Vitals** — INP ≤200ms — главный failing-метрик 2026; Lighthouse-score как прокси CWV — признанный антипаттерн. Нужны бюджеты (JS <300KB gzip, INP <150ms) и переход Lighthouse CI из artifact-only (NFR-17) в blocking assertions.

---

## 2. Таблица дельт

Приоритет: 🔴 high · 🟡 medium · 🟢 low.

| # | Тема | Что в playbook сейчас | Свежая практика 2026 | Рекомендация | Куда вписать |
|---|------|----------------------|----------------------|--------------|--------------|
| 1 | 🔴 WCAG 2.2 | `frontend.md` §9: «Hub целится в WCAG 2.1 AA» | WCAG 2.2 — текущая рекомендация W3C, ISO/IEC 40500:2025; EN 301 549 v4.1.1 (2026) переходит на 2.2; обратно совместим с 2.1 | Сменить цель на **2.2 AA**. Добавить чек-пункты: Dragging Movements (Kanban/DnD — альтернатива кнопками), Target Size ≥24px, Focus Not Obscured (sticky header/панели), Redundant Entry, Consistent Help. Отметить: 4.1.1 Parsing упразднён | `frontend.md` §9 (заголовок + чек-лист 9.9), `docs/security_guidelines.md`/NFR-16 |
| 2 | 🔴 Router preload | `frontend.md` §4.4: loader «опционально», `ensureQueryData` упомянут одной строкой | `defaultPreload: 'intent'` (prefetch on hover/touch), `defaultPreloadStaleTime: 0` (обязателен при Query, иначе 30s-кэш роутера ломает staleTime Query), `createRootRouteWithContext` c queryClient | Добавить подраздел «Preloading» в §4: канон-паттерн loader `ensureQueryData` + `useSuspenseQuery`; предупредить про `defaultPreloadStaleTime: 0` | `frontend.md` §4 (новый 4.9), настройка router в `main.tsx` |
| 3 | 🔴 Playwright шардинг | `playwright.config.ts`: `workers: 1` в CI; reporter `list`+`html`; merge-reports нигде | Шардинг `--shard=x/y` по CI-машинам + blob reporter + `npx playwright merge-reports` → единый HTML; параллелизм внутри файла через `test.describe.configure({ mode: 'parallel' })` | Для e2e-full (20 spec, 30 min timeout): 2–3 шарда + merge-reports artifact. Осторожно с бюджетом GitHub Actions (~3000 мин/мес) — шардинг умножает job-минуты незначительно, но сокращает wall-time | `testing.md` §6 (новый подраздел), `ci.yml` e2e-full |
| 4 | 🔴 `test.step()` | `testing.md` §8: шаги только через `allure.step` | Нативный `test.step()` Playwright — шаги видны в HTML-отчёте, trace viewer и UI Mode; allure-playwright конвертирует их в Allure-шаги автоматически | В Steps-слое POM оборачивать бизнес-шаги в `test.step()` — один API покрывает и Playwright-отчёт, и Allure | `testing.md` §6.1 (Steps), §8.4 |
| 5 | 🔴 `storageState` | `testing.md` §6.5–6.6: dev-login + inject sessionStorage **в каждом тесте** | Setup-project логинится один раз, сохраняет `storageState` (json), остальные projects декларируют `use: { storageState }` — экономия секунд на каждый spec | Ввести auth setup-project для e2e-full; sessionStorage-инъекцию оставить как fallback (Hub хранит сессию в sessionStorage — потребуется адаптация или переход на storageState-совместимое хранение для тестов) | `testing.md` §6.5, `playwright.config.ts` projects |
| 6 | 🔴 React 19 Actions | `frontend.md` §3.6: только `use()`/Suspense; `useActionState`/`useOptimistic` не упомянуты | Actions-модель стабильна: `useActionState` (pending+result форм), `useOptimistic` (optimistic UI c авто-откатом), `useFormStatus`. Для Vite-SPA работают без Server Actions — с обычными async-функциями | Добавить подраздел в §3: когда `useOptimistic` уместнее ручного optimistic-паттерна Query (локальный компонентный оптимизм: лайки, тогглы), а когда остаётся `onMutate`/rollback (мультиэкранное состояние кэша) | `frontend.md` §3 (новый 3.10), §5.5 (перекрёстная ссылка) |
| 7 | 🔴 React Compiler | §3.7: «компилятор может быть не включён — не полагайся» | React Compiler **v1.0 production-stable** (окт 2025), включается через `babel-plugin-react-compiler` в `@vitejs/plugin-react`; ESLint/Biome-plugin проверяет чистоту рендера; rollout по одному поддереву | Провести пилот: включить на одном модуле (например `tracker`), профилировать, при успехе — глобально и ослабить ручную мемоизацию. Зафиксировать решение ADR | `frontend.md` §3.7 + §11.5, ADR + `docs/stack_and_tools.md` |
| 8 | 🟡 useSuspenseQuery | §3.6 «точечно»; §5 весь на `useQuery`+`isLoading` | v5: `useSuspenseQuery` стабилен, `data` не-undefined на уровне типов; границы Suspense/Error — на маршрутах (Router даёт их из коробки). Ловушка: два `useSuspenseQuery` в одном компоненте фетчат последовательно → `useSuspenseQueries` | Расширить §5: критичные данные маршрута — suspense-паттерн + loader; вторичные — `useQuery`. Предупредить про waterfall | `frontend.md` §5 (новый подраздел) |
| 9 | 🟡 Simplified optimistic | §5.5: только паттерн `onMutate`/snapshot/rollback | v5 добавил облегчённый вариант: читать `mutation.variables` при `isPending` и рисовать оптимистичный элемент без записи в кэш (не нужен rollback) — для «одного места на экране» | Дополнить §5.5 выбором: variables-паттерн (одно место) vs кэш-паттерн (несколько мест) | `frontend.md` §5.5 |
| 10 | 🟡 Offline / persist | Нигде; Hub — PWA (`vite-plugin-pwa`) | `persistQueryClient` + `PersistQueryClientProvider` (требование `gcTime ≥ maxAge`), `networkMode: 'offlineFirst'` для SW-кэша | Оценить для PWA-сценария: минимум — задокументировать осознанный отказ; максимум — persist для справочников (channels, projects) | `frontend.md` §5 + §13 |
| 11 | 🟡 TS-конфиг | `tsconfig.json`: `strict`, `noUncheckedIndexedAccess` есть; нет `verbatimModuleSyntax`, `exactOptionalPropertyTypes`, `noUncheckedSideEffectImports`, `erasableSyntaxOnly` | Рекомендуемый строгий набор 2026: + `verbatimModuleSyntax` (честный `import type`), `noUncheckedSideEffectImports` (ловит опечатки в side-effect импортах), `erasableSyntaxOnly` (запрет enum/namespace — Hub на esbuild/Vite), `exactOptionalPropertyTypes` (самый шумный — последним) | Включать поэтапно: `verbatimModuleSyntax` + `noUncheckedSideEffectImports` + `erasableSyntaxOnly` сразу (дешёвые), `exactOptionalPropertyTypes` — отдельной задачей с прогоном | `frontend.md` §15.1, `frontend/tsconfig.json` |
| 12 | 🟡 TS-версия | `~5.7` в package.json; в 5.9 `--strictInference` входит в `strict` | TS 5.8: `erasableSyntaxOnly`; TS 5.9 (Q1 2026): `strictInference`, стабильные decorator metadata | Плановый апгрейд до 5.9 с прогоном `tsc --noEmit`; ожидать новые ошибки от strictInference | `docs/stack_and_tools.md`, chore-задача |
| 13 | 🟡 Tailwind 4 `@theme` | §8: токены `--hub-*` живут как CSS-переменные, применяются через arbitrary values `bg-[var(--card)]` | v4 CSS-first: токены объявляются в `@theme` → автоматически и CSS-переменная, и утилита (`bg-card` вместо `bg-[var(--card)]`); `@utility`/`@custom-variant` вместо plugins | Прописать в playbook правило: новые дизайн-токены — через `@theme`, чтобы получать чистые утилиты; постепенно уходить от arbitrary-values для стандартных токенов | `frontend.md` §8.1–8.2, `design.md` |
| 14 | 🟡 Container queries | §8.5 Responsive: только viewport-брейкпоинты (`sm:`, `md:`) | v4: container queries first-class (`@container`, `@sm`, `@md`) — правильный примитив для компонентной отзывчивости (панели, split-view, сайдбары) | Добавить в §8.5: для компонентов внутри resizable-панелей (messenger, board) — container queries вместо viewport | `frontend.md` §8.5 |
| 15 | 🟡 INP / CWV | §11: lazy, виртуализация, debounce — есть; метрик и бюджетов нет | INP ≤200ms заменил FID и является самым проваливаемым CWV; long tasks >50ms бить через `scheduler.yield()`; `content-visibility: auto`; поле (CrUX/RUM) важнее лабораторного Lighthouse-score | Добавить в §11 подраздел «Метрики»: цели INP<150ms/LCP<2.0s/CLS<0.05 (буфер к порогам Google), `web-vitals` lib для RUM, разбивка длинных обработчиков (drag, фильтрация board) | `frontend.md` §11, `observability.md` |
| 16 | 🟡 Perf-бюджеты в CI | CI: lighthouse job — artifact, non-blocking (NFR-17) | Lighthouse CI assertions / size-limit: сборка падает при превышении бюджета (JS <300KB gzip на маршрут, TBT, bundle-size diff на PR) | Перевести lighthouse job в blocking с минимальными assertions + добавить size-limit на основной чанк | `testing.md` §11, `ci.yml`, закрывает NFR-17 |
| 17 | 🟢 Vite 7 | Vite `^6.0.0` | Vite 7 (rolldown-vite) — Rust-бандлер, быстрее билд; миграция с v6 малоболезненная | Плановый chore-апгрейд после стабилизации экосистемы плагинов (pwa, tanstack router plugin) | `docs/stack_and_tools.md`, chore |
| 18 | 🟢 Vitest 4 Browser Mode | Vitest `^3.0.0`, jsdom; component-тесты только в jsdom | Vitest 4 (дек 2025): Browser Mode стабилен — компонентные тесты в реальном Chromium через provider `playwright()`; jsdom-моки (`matchMedia`, `ResizeObserver`) не нужны | Не мигрировать массово; точечно — для компонентов, страдающих от jsdom (DnD, виртуализация, Radix-поповеры). Сначала апгрейд Vitest 3→4 | `testing.md` §5 (примечание), backlog |
| 19 | 🟢 Allure 3 | §8.6: просмотр через `npx allure generate` (Allure 2 CLI подразумевается); allure-vitest не упомянут | Allure Report 3: `npm i -D allure` → `npx allure`; `allure-playwright`/`allure-vitest` совместимы с Report 2 и 3; для Vitest — reporter `allure-vitest/reporter` (осторожно с `describe.concurrent` в browser mode) | Добавить `allure-vitest` для frontend-unit (сейчас в TMS льются только e2e/Go); зафиксировать wrapper-функцию для `AS_ID`-ссылок вместо ручного `allure.label` в каждом тесте (рекомендация Qameta) | `testing.md` §8.4–8.6 |
| 20 | 🟢 Biome-строгость | `biome.json` (schema 2.4.8 — **уже Biome 2.x** ✅): почти все правила `warn`; domains не настроены; plugins нет | Biome 2.x: domains (`react`, `test`, `project`) включаются по package.json; type-aware `noFloatingPromises` (без tsc); GritQL-плагины для проектных запретов | Поднять a11y-правила и `noExplicitAny` до `error` (иначе gate декоративный — warn не валит CI); включить `noFloatingPromises`; GritQL-плагин для Hub-запретов (`key={index}`, inline style) — nice-to-have | `frontend.md` §16.5, `frontend/biome.json` |
| 21 | 🟢 `ref` как проп | §3: не упомянуто, `forwardRef` в кодовой базе Shadcn | React 19: `ref` передаётся как обычный проп, `forwardRef` не нужен для новых компонентов | Правило для нового кода: без `forwardRef`; существующие Shadcn-обёртки не трогать до апгрейда компонентов | `frontend.md` §3 |

---

## 3. Что уже соответствует 2026 (менять не нужно)

- **Мемоизация «по причине»** (§3.7) — точно совпадает с консенсусом 2026 «не пиши useMemo рефлекторно»; с React Compiler станет ещё проще.
- **Rules of Hooks + `enabled` вместо раннего return** — канон.
- **RTL-приоритет `getByRole` → `getByLabelText` → `getByText` → `testid`** и те же приоритеты в Playwright-локаторах — ровно официальные best practices.
- **Web-first assertions, ноль `waitForTimeout`**, `trace: 'on-first-retry'`, retries только в CI, POM+Steps — совпадает с Playwright guidance 2026.
- **MSW с `onUnhandledRequest: "error"`**, свой QueryClient на тест с `retry: false`, сброс синглтонов в `beforeEach` — актуально.
- **Optimistic updates c snapshot/rollback/`onSettled` invalidate** (§5.5) — это и есть официальный кэш-паттерн v5 (дельта #9 лишь добавляет облегчённый вариант).
- **Zustand: узкие селекторы, persist, слайсы** — без изменений в 2026.
- **Magic Link** удовлетворяет WCAG 2.2 **3.3.8 Accessible Authentication** из коробки (нет когнитивного теста).
- **Biome 2.x уже внедрён** (schema 2.4.8) как единый lint+format gate — сам выбор инструмента опережает многие команды; дельта только в строгости конфига.
- **Tailwind 4 уже на `@tailwindcss/vite`**, `!important` запрещён, дизайн-токены централизованы — миграционные breaking changes v3→v4 пройдены.
- **tsconfig**: `strict` + `noUncheckedIndexedAccess` + `moduleResolution: bundler` — сильная база (дельта #11 — только дополнение).
- **Snapshot-осторожность, антипаттерны флейков, детерминизм, traceability TC↔автотест, Allure-матчинг по `AS_ID`** — состояние искусства для in-house TMS.

---

## 4. Инструменты и gates для добавления в CI

| Инструмент / gate | Что даёт | Статус в CI | Приоритет |
|---|---|---|---|
| **Lighthouse CI assertions** (blocking) | Падение сборки при регрессии perf/a11y-скоров и бюджетов | Сейчас artifact-only (PR #27, NFR-17) | 🔴 |
| **size-limit / bundlesize** на PR | Diff размера основного чанка; ловит «случайно затянули LiveKit в общий бандл» | Нет | 🔴 |
| **Playwright shard + blob + merge-reports** для e2e-full | Wall-time 30 мин → ~10–12 мин; единый отчёт | Нет (`workers: 1`) | 🔴 |
| **axe-core на все ключевые экраны** (сейчас только login) + чек-лист WCAG 2.2 | Путь к NFR-16 sign-off на 2.2 AA | Частично (login axe PASS) | 🟡 |
| **Biome: warn→error** для a11y/`noExplicitAny` + `noFloatingPromises` | Gate перестаёт быть декоративным; ловит незавершённые промисы без tsc | Biome — gate, но правила warn | 🟡 |
| **web-vitals RUM** (p75 INP/LCP/CLS в observability-стек) | Полевые метрики вместо лабораторных; алерт при p75 INP >200ms | Нет | 🟡 |
| **allure-vitest reporter** | Frontend-unit результаты в TMS Hub (сейчас только e2e/Go) | Нет | 🟢 |
| **auth setup-project + storageState** | Секунды с каждого e2e-spec | Нет | 🟢 |

**Замечание по бюджету CI:** студенческий лимит ~3000 мин/мес — шардинг e2e-full и blocking-Lighthouse добавляют job'ы, поэтому включать под условие «e2e-full только на master push + label», как уже сделано для coverage-gate.

---

## 5. Предлагаемый порядок внедрения (для team-lead)

1. **Волна quick-wins (без риска):** дельты #11 (три дешёвых tsc-флага), #20 (Biome error-уровни), #4 (`test.step`), #21 (`ref` prop) — правки playbook + конфиги, полдня.
2. **Волна CI:** #3 (шардинг+merge), #16 (Lighthouse blocking + size-limit), #5 (storageState) — закрывает NFR-17.
3. **Волна WCAG 2.2 (#1):** аудит Kanban DnD (2.5.7), target-size пас по иконочным кнопкам (2.5.8), sticky-панели (2.4.11) → обновление §9 playbook и NFR-16.
4. **Волна Router/Query (#2, #8, #9):** intent-preload + suspense-паттерн на 2–3 тяжёлых маршрутах (board, issue detail), замер.
5. **Пилоты:** React Compiler (#7, ADR), `useOptimistic` (#6) на messenger-send, persistQueryClient (#10) для PWA.

---

*Отчёт создан 2026-07-03. Playbooks и код не изменялись — только этот файл.*
