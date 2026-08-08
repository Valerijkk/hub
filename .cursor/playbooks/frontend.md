# Hub — Эталонный Playbook по фронтенд-разработке

> **Статус:** единый эталон для AI-команды фронтенда Hub.
> **На кого рассчитан:** агенты `fe1`, `fe2` (исполнители), `fe-ver` (верификатор).
> **Как использовать:** читать целиком перед первой фронтенд-задачей сессии; далее — как справочник по разделам.
> **Связанные документы:**
> - `.cursor/rules/30-frontend.mdc` — краткие правила (источник истины по паттернам)
> - `.cursor/rules/10-style.mdc` — стиль кода Go/TS
> - `.cursor/rules/60-tz-constraints.mdc` — mandatory/forbidden ограничения ТЗ
> - `docs/code_style_guide.md` — полный гайд по стилю (TS-часть)
> - `docs/architecture.md` — архитектура, C4, frontend-раздел
> - `.cursor/playbooks/design.md` — дизайн-система, токены `--hub-*`, визуальный язык
> - `.cursor/playbooks/testing.md` — тестовая стратегия (native units + greenfield Python non-unit QA)
> - `.cursor/playbooks/code-comments.md`, `.cursor/playbooks/observability.md`, `.cursor/playbooks/git-workflow.md`
>
> **Правило приоритета:** если этот playbook расходится с `.cursor/rules/*.mdc` или `.cursor/rules/60-tz-constraints.mdc` — **правила и ТЗ побеждают**. Playbook расширяет и объясняет, но не отменяет их.

---

## Оглавление

Канон здесь. Библиотеки и специальные темы — отдельными файлами: они нужны
при работе с конкретной технологией, а не на каждой задаче.

**В этом файле:** роль и объём, структура и границы модулей, безопасность,
документация, Definition of Done, чек-лист перед `fe-ver`.

| Файл | Когда открывать |
|------|-----------------|
| [`frontend/api-layer.md`](frontend/api-layer.md) | API-слой и где живёт бизнес-логика |
| [`frontend/a11y-styles.md`](frontend/a11y-styles.md) | Доступность WCAG 2.2 AA и система стилей |
| [`frontend/react.md`](frontend/react.md) | Компоненты React 19, состояние, Error Boundaries |
| [`frontend/tanstack.md`](frontend/tanstack.md) | Router и Query: маршруты, кэш, инвалидация |
| [`frontend/state-realtime.md`](frontend/state-realtime.md) | Zustand и живые данные по WebSocket |
| [`frontend/perf-i18n.md`](frontend/perf-i18n.md) | Производительность и интернационализация |
| [`frontend/reference.md`](frontend/reference.md) | Сквозной пример модуля и обновления 2026 |

Нумерация разделов сохранена — ссылки вида «§5» остаются рабочими.
## Содержание

1. [Кто и когда использует playbook (fe1 / fe2 / fe-ver)](#1-кто-и-когда-использует-playbook)
2. [Организация файлов и границы модулей](#2-организация-файлов-и-границы-модулей)
3. [Компоненты React 19](#3-компоненты-react-19)
4. [TanStack Router](#4-tanstack-router)
5. [TanStack Query](#5-tanstack-query)
6. [Zustand](#6-zustand)
7. [API-слой](#7-api-слой)
8. [Дизайн-система на практике](#8-дизайн-система-на-практике)
9. [Доступность (a11y, WCAG 2.2 AA)](#9-доступность-a11y-wcag-22-aa)
10. [Формы и валидация](#10-формы-и-валидация)
11. [Производительность](#11-производительность)
12. [Error Boundaries](#12-error-boundaries)
13. [Реальное время (WebSocket)](#13-реальное-время-websocket)
14. [Интернационализация (i18n)](#14-интернационализация-i18n)
15. [Типобезопасность](#15-типобезопасность)
16. [Тестирование фронтенда](#16-тестирование-фронтенда)
17. [Комментарии, логирование, git](#17-комментарии-логирование-git)
18. [Обязательная регрессия и docs-sync](#18-обязательная-регрессия-и-docs-sync)
19. [Definition of Done и что проверяет fe-ver](#19-definition-of-done-и-что-проверяет-fe-ver)
20. [Антипаттерны Hub](#20-антипаттерны-hub)
21. [Быстрый чек-лист перед передачей fe-ver](#21-быстрый-чек-лист-перед-передачей-fe-ver)

---
## Стек (факты, не обсуждается без ADR)

| Слой | Технология | Ограничение |
|------|-----------|-------------|
| UI-библиотека | React 19 | Функциональные компоненты, никаких классов кроме Error Boundary |
| Язык | TypeScript 5.7 | `strict: true`, `any` запрещён без `// reason` |
| Сборка | Vite | ESM, code-splitting через динамический `import()` |
| Роутер | TanStack Router | **Единственный**, file-based, 39 routes. Другой роутер запрещён |
| Server state | TanStack Query | Всё, что приходит с бэка |
| Client state | Zustand (11 stores) | UI-состояние, выбор, драг, сессия |
| UI-kit | Shadcn/ui + Tailwind CSS 4 | Менять только `className`; поведение Radix не трогать |
| i18n | i18next + react-i18next | Нет хардкода user-visible строк |
| HTTP | Кастомный `ApiClient` (fetch) | `api/client.ts`, single-flight refresh на 401 |

**Жёстко запрещено (см. `60-tz-constraints.mdc` и `10-style.mdc`):**

- другой роутер или UI-framework вместо TanStack Router / Shadcn;
- `!important` в CSS (Phase 8 вычистил 56 штук — новые не добавлять);
- `key={index}` в списках;
- non-null assertion `!` без явного guard;
- вызов хука после условного `return` (нарушение Rules of Hooks);
- inline `style={{}}` вместо дизайн-токенов `--hub-*`;
- `any` без пояснительного комментария.

---
## 1. Кто и когда использует playbook

### 1.1. Роли

| Роль | Agent | Skill | Когда включается |
|------|-------|-------|------------------|
| Frontend Dev #1 | `fe1` | `.cursor/agents/fe1.md` / `fe2.md` | Реализация модуля/страницы A |
| Frontend Dev #2 | `fe2` | `.cursor/agents/fe1.md` / `fe2.md` | Реализация модуля/страницы B (параллельно A) |
| Frontend Verifier | `fe-ver` | `.cursor/project_skills/FE_VER_skill.md` | Верификация после fe1/fe2 |

`fe1` и `fe2` — **исполнители одного профиля**. Они пишут одинаково качественный код по этому playbook. Разница только в распределении зон ответственности внутри одной волны. `fe-ver` — **независимый контролёр**: он не пишет фичу, а проверяет её по разделу [19](#19-definition-of-done-и-что-проверяет-fe-ver).

### 1.2. Как fe1 и fe2 делят работу без конфликтов

Главный риск параллельной работы двух агентов — merge-конфликты и рассинхрон типов API. Правила разделения:

**Правило 1. Один модуль — один агент за волну.** Модули живут в `src/modules/<module>/` и `src/components/<feature>/`. Волна планируется так, чтобы `fe1` и `fe2` не касались одного и того же каталога модуля.

```
fe1 -> src/modules/tracker/**, src/components/tracker/**, src/api/tracker.ts
fe2 -> src/modules/messenger/**, src/components/messenger/**, src/api/messenger.ts
```

**Правило 2. Общие ui-компоненты (`src/components/ui/*`) — только по договорённости.** Файлы Shadcn (`button.tsx`, `input.tsx`, `select.tsx`, `dialog.tsx`) используют оба агента. Их **нельзя** менять «под свою фичу»: если нужен новый вариант — расширяй через `className` в потребителе, а не правь общий компонент. Если общий компонент действительно нужно доработать (новый проп) — это отдельная атомарная задача, согласованная с обоими агентами, и она идёт **первой**, до параллельной работы.

**Правило 3. Контракты API фиксируются заранее.** Файл `src/api/<module>.ts` — это граница между FE и BE. Перед реализацией UI оба агента сверяются с backend-контрактом (`docs/system_specification.md`, `docs/wiki/API.md`) и с Response/Request DTO. Тип из `api/*.ts` — **единственный источник истины** для формы данных. UI не выдумывает собственные типы для тех же сущностей.

**Правило 4. Общие точки (роутер, query client, stores index) трогаем осторожно.** `main.tsx`, `routeTree.gen.ts` (генерируется), глобальный `QueryClient` — потенциальные конфликты. `routeTree.gen.ts` генерируется автоматически плагином TanStack Router — **не редактировать руками**, только добавлять route-файлы.

**Правило 5. Каждый агент делает атомарные коммиты по слою.** Не «всё сразу»: сначала `api/`, потом `hooks/`, потом `components/`, потом `route`. Это уменьшает площадь конфликта и упрощает ревью для `fe-ver`.

### 1.3. Хороший и плохой раздел работы

<good-example>
Волна «Tracker + Messenger» разбита так, что пересечений нет:

- `fe1`: `src/modules/tracker/`, `src/components/tracker/`, `src/api/tracker.ts`, `src/routes/_authenticated/projects/$projectId/board/`
- `fe2`: `src/modules/messenger/`, `src/components/messenger/`, `src/api/messenger.ts`, `src/routes/_authenticated/messenger/`
- Общее (`src/components/ui/dialog.tsx` нужен обоим) — уже существует, никто не правит.
</good-example>

<bad-example>
`fe1` и `fe2` оба добавляют проп `size` в `src/components/ui/button.tsx` в одной волне -> конфликт при merge, дважды переопределённый вариант, риск сломать Radix-поведение. Правильно: сначала одна задача на `button.tsx`, потом параллель.
</bad-example>

### 1.4. Чек-лист старта фронтенд-задачи

- [ ] Прочитаны `docs/*` (vision, system_specification, architecture — frontend раздел).
- [ ] Определён затронутый модуль и его каталог; нет пересечения с зоной другого агента.
- [ ] Сверён контракт API (`src/api/<module>.ts` + `docs/wiki/API.md`).
- [ ] Определены затронутые REQ/US (для docs-sync в конце).
- [ ] Понятно, где server-state (Query), где client-state (Zustand).
- [ ] Понятно, нужны ли Error Boundary, lazy-загрузка, i18n-ключи.

---
## 2. Организация файлов и границы модулей

### 2.1. Каноническая структура

```
frontend/src/
├── api/<module>.ts          # API-клиент: типы DTO + функции запросов (один файл на backend-модуль)
├── api/client.ts            # Базовый HTTP-клиент (fetch + refresh mutex)
├── api/types.ts             # Общие типы (PaginatedResponse и т. п.)
├── components/ui/           # Shadcn-компоненты (только className-изменения)
├── components/<feature>/    # Feature-компоненты (tracker/, messenger/, notifications/, ...)
├── hooks/use<Name>.ts       # Кросс-модульные хуки (use-websocket, use-debounced-value, ...)
├── modules/<module>/        # Самодостаточные фичи
│   ├── components/
│   ├── hooks/
│   └── stores/
├── routes/                  # TanStack Router file-based routes
├── stores/<module>.ts       # Глобальные Zustand-сторы (auth, ws, board, ...)
├── lib/                     # Утилиты (format-api-error, utils, onboarding-схемы)
└── i18n/                    # Инициализация i18next + locales/{ru,en}/*.json
```

### 2.2. Границы: где что живёт

| Что | Где | Пример |
|-----|-----|--------|
| Типы и функции запросов | `api/<module>.ts` | `createIssue`, `IssueResponse` |
| Server-state хуки (Query) | `hooks/` или `modules/<m>/hooks/` | `useMessages`, `useChannels` |
| Client-state (Zustand) | `stores/<module>.ts` | `useBoardStore`, `useAuthStore` |
| Переиспользуемый UI-примитив | `components/ui/` | `Button`, `Input`, `Select` |
| Feature-компонент | `components/<feature>/` или `modules/<m>/components/` | `IssueForm`, `BoardColumn` |
| Страница (маршрут) | `routes/` | `_authenticated/projects/$projectId/board/index.tsx` |
| Утилита без React | `lib/` | `buildQueryString`, `toastApiError` |

### 2.3. `components/<feature>/` vs `modules/<module>/`

- **`modules/<module>/`** — крупная самодостаточная фича с собственными компонентами, хуками и (иногда) сторами. Например `modules/tracker/kanban-board.tsx`, `modules/calls/`.
- **`components/<feature>/`** — набор feature-компонентов, которые вставляются в маршруты и в модули. Например `components/tracker/issue-form.tsx`, `components/tracker/board-column.tsx`.

На практике в Hub оба каталога сосуществуют исторически. **Правило для нового кода:** новую крупную фичу оформляй как `modules/<module>/`. Мелкие компоненты внутри уже существующего домена клади рядом с существующими (не создавай третий каталог).

### 2.4. Правила именования (из `docs/code_style_guide.md`)

| Сущность | Конвенция | Пример файла |
|----------|-----------|--------------|
| Компонент | PascalCase | `IssueCard.tsx` |
| Хук | `use` prefix, kebab-файл | `use-messages.ts` |
| Store | `useXStore`, kebab-файл | `board-store.ts` -> `useBoardStore` |
| Route | kebab/`$param` в path | `$issueKey.tsx` |
| API-тип | PascalCase, суффикс роли | `CreateIssueRequest`, `IssueResponse` |
| Тест | `*.test.ts(x)` | `auth-store.test.ts` |

### 2.5. Импорты: alias `@/`

Используется alias `@/` на `src/`. Порядок импортов Biome сортирует автоматически, но логически:

```typescript
// 1. Внешние пакеты
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

// 2. Внутренние через @/
import { createIssue, type IssueResponse } from "@/api/tracker";
import { Button } from "@/components/ui/button";
import { useBoardStore } from "@/stores/board-store";
```

<good-example>
Импорт типа через `import type` — не тянет рантайм-значение:

```typescript
import { createIssue, type IssueResponse } from "@/api/tracker";
```
</good-example>

<bad-example>
Относительные лесенки `../../../` вместо alias — хрупко при переносе файла:

```typescript
import { Button } from "../../../components/ui/button";
```
</bad-example>

### 2.6. Чек-лист организации файлов

- [ ] API-функции и DTO — в `api/<module>.ts`, а не разбросаны по компонентам.
- [ ] Общие UI-примитивы берутся из `components/ui/`, не дублируются.
- [ ] Новый крупный модуль — в `modules/<module>/` с подкаталогами.
- [ ] Импорты через alias `@/`, типы через `import type`.
- [ ] Файл-роут положен в правильную ветку `routes/` (см. раздел 4).
- [ ] Нет пересечения каталога с зоной другого агента.

---
## 15. Типобезопасность

### 15.1. Строгий TypeScript

Проект на `strict: true`. Цель — «если компилируется, значит форма данных верна».
Регрессия начинается с typecheck-шага task-owned Docker job `frontend-unit`
(раздел 18), не с host-команды.

### 15.2. Запрет `any` без причины

`any` отключает проверки. Запрещён без пояснительного комментария. Для «неизвестно что» используй `unknown` и сужай типами/гардами.

<good-example>
`unknown` + сужение (паттерн из WS-обработчиков):

```tsx
const handleEvent = (event: string, payload: unknown) => {
	const data = payload as Record<string, unknown>;
	const id = data.message_id as string | undefined;
	if (!id) return;
	// ...
};
```
</good-example>

<bad-example>
`any` глушит ошибки:

```tsx
const handleEvent = (event: string, payload: any) => {
	doThing(payload.message_id.toUpperCase()); // упадёт в рантайме, TS молчит
};
```
</bad-example>

Если `any` реально неизбежен — оставь причину: `// biome-ignore lint/suspicious/noExplicitAny: сторонний тип без типизации`.

### 15.3. Optional chaining `?.` + nullish `??`

Для потенциально отсутствующих значений — `?.` и `??`. Non-null `!` — только с доказанным guard.

<good-example>
```tsx
const name = user?.profile?.name ?? "Unknown";
const first = lastPage[0]?.created_at;
```
</good-example>

<bad-example>
```tsx
const name = user!.profile!.name; // рантайм-краш, если null
```
</bad-example>

**Про `!` с guard.** В Hub допускается точечный `!` там, где выполнение гарантировано флагом `enabled`, но обязательно с `biome-ignore`-комментарием, объясняющим гарантию:

```tsx
queryFn: ({ pageParam }) =>
	// biome-ignore lint/style/noNonNullAssertion: enabled гарантирует наличие channelId
	getMessages(channelId!, { before: pageParam, limit: PAGE_SIZE }),
```

Без такого guard/комментария `!` — запрещён.

### 15.4. Дискриминированные юнионы

Для «одно из нескольких» используем discriminated unions с полем-дискриминатором — TS сузит тип в `switch`.

```typescript
export type AuthSyncEvent =
	| { type: "logout" }
	| { type: "login"; userId: string | null };

function handle(event: AuthSyncEvent) {
	switch (event.type) {
		case "login":
			return connectAs(event.userId); // здесь userId доступен
		case "logout":
			return clear();
	}
}
```

Для перечислений-строк — литеральные юнионы (как в `tracker.ts`): `type IssuePriority = "critical" | "high" | "medium" | "low" | "trivial"`.

### 15.5. Типы из API — единый источник истины

Не дублируй серверные типы. Импортируй DTO из `api/<module>.ts`. Формы выводят тип из Zod-схемы (`z.infer`).

```tsx
import type { IssueResponse, CreateIssueRequest } from "@/api/tracker";
import type { IssueFormValues } from "@/components/tracker/issue-form-schema";
```

<bad-example>
Локальный «свой» тип задачи, дублирующий `IssueResponse` — рассинхрон при изменении контракта:

```tsx
interface MyIssue { id: string; title: string } // дубль API-типа
```
</bad-example>

### 15.6. Прочие правила типизации

- Возвращаемые типы публичных функций/хуков — явные, где неочевидно.
- Массивы/объекты в стейте — типизированы (`useState<Value | undefined>`), не выводятся в `never[]`.
- Приведение `as` — минимально; предпочитай гарды (`typeof`, проверки полей).
- `React.FC` не обязателен; типизируй пропсы напрямую.

### 15.7. Чек-лист типобезопасности

- [ ] `tsc --noEmit` без ошибок.
- [ ] Нет `any` без `biome-ignore`-причины; неизвестное — `unknown` + сужение.
- [ ] `?.`/`??` вместо `!`; каждый `!` имеет guard + комментарий.
- [ ] «Одно из» — discriminated unions; перечисления — литеральные юнионы.
- [ ] Типы данных импортированы из `api/*`; форма — `z.infer`; нет дублей API-типов.

---
## 16. Тестирование фронтенда

> Полная стратегия — в `.cursor/playbooks/testing.md`. Здесь — факты и минимум для fe1/fe2.

### 16.1. Где лежат тесты (0 тестов в `src/`)

| Тип | Расположение | Инструмент |
|-----|--------------|-----------|
| Genuine unit/component behavior | `frontend/tests/<domain>/` | Vitest + React Testing Library |
| Unit support | `frontend/tests/support/` | Локальные builders/providers; MSW только для hermetic unit boundary |
| API/contract/integration | `testframework/tests/{api,contract,integration}/` | Python, владельцы `qa-auto1/2` |
| Browser journeys | `testframework/tests/e2e/` | Python Playwright; `smoke` marker, `full` launcher profile |

Тесты **не** живут в `frontend/src/`. Static/source-contract Vitest и
TypeScript browser suites не портируются. POM/Steps — не обязательная форма:
Component/Page Object появляется только после повторения meaningful UI-knowledge.

### 16.2. Перед запуском — изолированные зависимости

Зависимости устанавливаются в task-owned Docker volume канонического
`frontend-unit` runner из `AGENTS.md`. Host `node_modules` и legacy link-script
не используются как evidence.

### 16.3. Команды

Каноничная команда — task-owned limited Docker job `frontend-unit` из
`AGENTS.md`: install dependencies в named volume, typecheck, lint, запуск
`frontend/tests/**` и build. Evidence включает source SHA, image, inner command,
CPU/RAM, collection, duration и cleanup.

### 16.4. Что и как тестировать (минимум)

- **Компоненты**: RTL — рендер, взаимодействие (`userEvent`), проверка доступного текста/ролей (`getByRole`), а не деталей реализации.
- **Хуки/Query**: оборачивай в `QueryClientProvider`; API мокай через MSW, не патчь `fetch` руками.
- **Zustand**: сбрасывай стор между тестами; тестируй действия и селекторы.
- **A11y-регресс**: unit проверяет семантику компонента; browser outcome через
Python E2E в `testframework/tests/e2e/**` после верифицированного TC; browser lifecycle/artifacts use exact task paths in `testframework/src/hub_testkit/browser/**`, and UI helpers in `testframework/src/hub_testkit/domains/<domain>/ui/{pages,components}/**`.

```tsx
// Набросок компонентного теста (RTL)
render(
	<QueryClientProvider client={queryClient}>
		<IssueForm projectId="p1" mode="create" />
	</QueryClientProvider>,
);
await userEvent.type(screen.getByLabelText(/title/i), "New issue");
await userEvent.click(screen.getByRole("button", { name: /create/i }));
expect(await screen.findByText(/created/i)).toBeInTheDocument();
```

### 16.5. Линтер — Biome

Линтинг и форматирование — Biome внутри Docker job `frontend-unit`. Он ловит
`key={index}`, non-null без причины, неиспользуемое, порядок импортов и правила
хуков. `biome-ignore` — только с причиной.

### 16.6. Чек-лист тестирования

- [ ] Genuine units созданы в `frontend/tests/<domain>/`, не в `src/`.
- [ ] Запуск выполнен task-owned limited Docker job `frontend-unit`.
- [ ] API мокается через MSW; Query обёрнут в провайдер.
- [ ] Тестируем поведение/доступные роли, не детали реализации.
- [ ] Typecheck/lint/Vitest/build внутри Docker зелёные; exact evidence и cleanup приложены.
- [ ] Non-unit browser/API risk передан `qa-auto*` как Python TC; POM не назначен ритуальным gate.

---
## 17. Комментарии, логирование, git

> Детали — в `.cursor/playbooks/code-comments.md`, `.cursor/playbooks/observability.md`, `.cursor/playbooks/git-workflow.md`. Кратко для фронта:

### 17.1. Комментарии — WHY, не WHAT

Комментарии на **английском** в коде (docstring-стиль в Hub допускает русские пояснения бизнес-логики, но новые технические комментарии — по правилу WHY). Объясняем неочевидное намерение/ограничение, а не пересказываем код. Без эмодзи.

<good-example>
Объясняет ПОЧЕМУ (гонка refresh):

```tsx
// Single-flight: параллельные 401 не должны запускать несколько refresh —
// иначе первый инвалидирует токен и остальные разлогинят пользователя (FE-H1).
```
</good-example>

<bad-example>
Пересказ кода:

```tsx
// increment counter
count += 1;
```
</bad-example>

### 17.2. Логирование

- Пользователю — тосты (`sonner`), не `console`.
- `console.error` — только для действительно неожиданных ошибок (как в Error Boundary), не как обычный поток.
- **Не логировать** секреты: токены, magic-link коды, refresh-токены.
- `request_id` из `ApiError` можно логировать для отладки.

### 17.3. Git

Conventional Commits, scope фронта — `frontend` (или доменный: `tracker`, `messenger`). Атомарные коммиты по слою.

```
feat(tracker): add issue form validation with zod
fix(messenger): dedupe ws message_sent by id
refactor(frontend): extract useDebouncedValue hook
```

Не коммить: `.env`, секреты, артефакты сборки (`frontend/dist/`). Ветки: `feat/<module>/<desc>`, `fix/<module>/<desc>`.

---
## 18. Обязательная регрессия и docs-sync

### 18.1. Обязательная регрессия (перед handoff)

После любого изменения фронта, **до передачи `fe-ver`/коммита**, выполнить
task-owned limited Docker job `frontend-unit` из `AGENTS.md`. Host-run запрещён.
Job обязан дать non-zero collection для `frontend/tests/**`, typecheck/lint/
Vitest/build PASS и подтверждённый cleanup.

Правило `56-regression-verify.mdc`: `verified: true` — только при зелёной регрессии с приложенным логом. Останавливаемся на первой FAIL и чиним (сначала проверяем корректность функционала по `docs/`, потом решаем — код или тест; см. `.cursor/skills/test-triage/SKILL.md`).

### 18.2. Docs-sync (обязательно, в той же сессии)

После функционального изменения обновляем документацию (правило `55-docs-sync.mdc`):

| Изменение | Документ |
|-----------|----------|
| Любая фича/фикс | `.planning/STATE.md` (session note) |
| Новый UI flow / hotkey | `docs/wiki/SHORTCUTS.md`, `docs/wiki/QUICKSTART.md`; `docs/vision_and_scope.md` если US |
| Новый API-контракт (потребляемый) | сверка с `docs/system_specification.md`, `docs/wiki/API.md` |
| Тесты/coverage | `.planning/COVERAGE.md`, `docs/test_plan.md` |

**Конкретно для фронта:** новый хоткей -> `SHORTCUTS.md`; новый пользовательский сценарий -> `QUICKSTART.md`. Не оставляй drift: wiki не должна обещать то, чего нет в UI.

### 18.3. Артефакты

Не коммить временные дампы. Coverage/артефакты чистятся по `45-artifacts-cleanup.mdc`.

### 18.4. Чек-лист регрессии и docs

- [ ] Docker job `frontend-unit` — typecheck/lint/Vitest/build PASS.
- [ ] Source/image/resources/collection/duration/cleanup приложены.
- [ ] `.planning/STATE.md` обновлён.
- [ ] Новый flow/hotkey отражён в `SHORTCUTS.md`/`QUICKSTART.md`.
- [ ] Нет закоммиченных артефактов/секретов.

---
## 19. Definition of Done и что проверяет fe-ver

### 19.1. Definition of Done фронтенд-задачи

Задача считается готовой, когда **всё** выполнено:

- [ ] Функционал соответствует требованиям (US/REQ из `docs/`).
- [ ] Server-state через Query, client-state через Zustand, форма через RHF+Zod — по назначению.
- [ ] Компоненты соблюдают Rules of Hooks; ключи по стабильному `id`.
- [ ] Роуты корректны; защищённые — под `_authenticated`; навигация типобезопасна.
- [ ] Мутации инвалидируют кэш; optimistic updates имеют rollback.
- [ ] Дизайн — только токены `--hub-*`/Tailwind/Shadcn className; нет inline-стилей/`!important`.
- [ ] A11y: семантика, `aria-label` на иконках, focus-visible, keyboard nav, контраст; axe чист.
- [ ] i18n: нет хардкода; ключи в ru+en; даты в TZ пользователя.
- [ ] Типобезопасность: нет `any` без причины; `?.`/`??`; типы из API.
- [ ] Тяжёлое — lazy; списки виртуализированы; debounce где нужно.
- [ ] Крупные модули обёрнуты в Error Boundary.
- [ ] WS-интеграция иммутабельна, с дедупом и отпиской.
- [ ] Genuine units добавлены/обновлены в `frontend/tests/**`; non-unit risk маршрутизирован в Python `testframework/**`; Docker regression зелёная.
- [ ] Docs-sync выполнен (STATE + SHORTCUTS/QUICKSTART при новом flow/hotkey).
- [ ] Атомарные коммиты, Conventional Commits, без секретов/артефактов.

### 19.2. Что проверяет fe-ver

`fe-ver` (`.cursor/agents/fe-ver.md`) верифицирует **независимо**:

1. **Соответствие требованиям** — фича делает то, что в US/REQ.
2. **Rules of Hooks / React-паттерны** — нет хуков после return, нет `key={index}`, нет `!` без guard.
3. **State-разделение** — нет server-state в Zustand; мутации инвалидируют; optimistic с rollback.
4. **A11y (axe)** — 0 critical/serious; keyboard nav (Tab/Enter/Escape) работает; `aria-label` на иконках; контраст.
5. **Lighthouse** — производительность/лучшие практики на ключевых экранах (в Hub Lighthouse CI — artifact job, non-blocking, но fe-ver смотрит регресс).
6. **Дизайн-система** — токены, нет `!important`/inline-стилей, Shadcn не сломан.
7. **i18n** — нет хардкода, ключи в обоих языках, даты локализованы.
8. **Типобезопасность и линт** — `tsc --noEmit` + Biome зелёные.
9. **Тесты и регрессия** — приложен зелёный лог; новый flow покрыт.
10. **Docs-sync** — STATE и wiki обновлены.

`fe-ver` ставит `verified: true` только при зелёной регрессии с логом и закрытых CRITICAL/HIGH замечаниях.

---
## 20. Антипаттерны Hub

| Нельзя | Почему | Как правильно |
|--------|--------|---------------|
| `key={index}` в списках | Ломает реконсиляцию при вставке/удалении/сортировке | Стабильный `key={item.id}` |
| Хук после условного `return` | Нарушение Rules of Hooks -> краш «fewer hooks» | Все хуки до return; ветвление через `enabled` |
| `user!.profile!.name` без guard | Рантайм-краш на null | `user?.profile?.name ?? fallback`; `!` только с guard+комментом |
| `any` без причины | Отключает проверки типов | `unknown` + сужение; `biome-ignore` с причиной если неизбежно |
| Server-state в Zustand | Рассинхрон, нет кэша/инвалидации | TanStack Query |
| Optimistic update без rollback | Навсегда неверное состояние при ошибке | `onMutate` snapshot + `onError` rollback + `onSettled` invalidate |
| Refresh без single-flight mutex | Гонка 401 -> ложные разлогины | Модульный promise-lock (FE-H1) |
| Ручная сборка URL `` `/x/${id}` `` в навигации | Теряется типобезопасность роутера | `<Link to params>` / `navigate({ to, params })` |
| Инлайн `style={{}}` поверх дизайна | Несогласованность, ломает тему | Токены `--hub-*` + Tailwind |
| `!important` в CSS | Техдолг, войны специфичности | Tailwind-слои/утилиты |
| Изменение поведения Radix в Shadcn | Ломает a11y/фокус/Esc | Только `className`; вариант через cva в ui-компоненте |
| `<div onClick>` как кнопка | Не фокусируется, не читается СР | `<button type="button">` |
| Иконочная кнопка без `aria-label` | Скринридер читает «button» | `aria-label` + `aria-hidden` на svg |
| Хардкод user-visible строк | Непереводимо | `t("ns:key")`, ключ в ru+en |
| Сырой UTC-ISO пользователю | Неверное локальное время | `Intl.DateTimeFormat(i18n.language, ...)` |
| Подписка на весь Zustand-стор для 1 поля | Лишние ререндеры | Узкий селектор `(s) => s.field` |
| `useEffect` для производного значения | Лишний рендер, рассинхрон | Вычисление при рендере / `useMemo` |
| Мемоизация «на всякий» | Шум, лишняя работа | Только для дорогого/стабильности deps |
| Рендер всего длинного списка | Фриз, огромный DOM | Виртуализация + infinite scroll |
| Тяжёлая либа (LiveKit/charts) в основном бандле | Раздутый бандл | `lazy()` + Suspense + Error Boundary |
| Правка `routeTree.gen.ts` вручную | Затирается генератором | Добавлять route-файлы, не трогать gen |
| Тесты в `frontend/src/` | Нарушение pipeline-структуры | `frontend/tests/<domain>/` |
| WS-событие: мутация кэша на месте | Баг реконсиляции, дубли | `setQueryData` иммутабельно + дедуп по `id` |
| Логирование токенов/секретов | Утечка | Не логировать; можно `request_id` |

---
## 21. Быстрый чек-лист перед передачей fe-ver

Прогони по пунктам перед handoff. Если хоть один не выполнен — не передавай.

**Код и React**
- [ ] Все хуки до любого `return`; управление запросом через `enabled`.
- [ ] `key` по стабильному `id` (скелетоны — составной строковый ключ).
- [ ] `type="button"` на неотправляющих кнопках.
- [ ] Нет `any` без причины; `?.`/`??`; каждый `!` с guard+комментом.
- [ ] Мемоизация только по делу.

**State и данные**
- [ ] Server-state в Query, client-state в Zustand, форма в RHF+Zod.
- [ ] `queryKey` по конвенции; `staleTime` осмыслен.
- [ ] Мутации инвалидируют минимальный набор; optimistic с rollback (4 шага).
- [ ] Узкие Zustand-селекторы; persist только для нужного.
- [ ] Форма редактирования шлёт `version`.

**Роутинг и API**
- [ ] Путь `createFileRoute` = положение файла; защищённое под `_authenticated`.
- [ ] Навигация типобезопасна (`Link`/`navigate` + `params`).
- [ ] Новые запросы в `api/<module>.ts`, типизированы; single-flight refresh не тронут.

**UI и дизайн**
- [ ] Только токены `--hub-*`/Tailwind/Shadcn className; нет inline-стиля/`!important`.
- [ ] Responsive; тёмная тема работает.

**A11y**
- [ ] Семантика; `aria-label` на иконках; focus-visible; Tab/Enter/Escape.
- [ ] Ошибки полей: `aria-invalid`/`aria-describedby`/`role="alert"`.
- [ ] axe: 0 critical/serious.

**i18n**
- [ ] Нет хардкода; ключи в ru+en; даты в TZ пользователя (UTC -> Intl).

**Устойчивость и перф**
- [ ] Крупные модули в Error Boundary; тяжёлое — lazy + Suspense.
- [ ] Длинные списки виртуализированы; поиск дебаунсится.
- [ ] WS: иммутабельные апдейты, дедуп по `id`, отписка в cleanup, reconnect с backoff.

**Качество и процесс**
- [ ] Docker job `frontend-unit` — typecheck/lint/Vitest/build PASS.
- [ ] Genuine units добавлены/обновлены в `frontend/tests/<domain>/`; non-unit Python QA маршрутизирован отдельно.
- [ ] Exact Docker evidence и cleanup приложены.
- [ ] `.planning/STATE.md` + `SHORTCUTS.md`/`QUICKSTART.md` (при новом flow/hotkey) обновлены.
- [ ] Атомарные Conventional Commits; без секретов/артефактов.

> Готово к `fe-ver`, когда все чекбоксы отмечены и приложен зелёный лог регрессии.

---
