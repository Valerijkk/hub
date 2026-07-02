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
> - `.cursor/playbooks/testing.md` — тестовая стратегия (Vitest, RTL, MSW, Playwright)
> - `.cursor/playbooks/comments.md`, `.cursor/playbooks/logging.md`, `.cursor/playbooks/git.md`
>
> **Правило приоритета:** если этот playbook расходится с `.cursor/rules/*.mdc` или `.cursor/rules/60-tz-constraints.mdc` — **правила и ТЗ побеждают**. Playbook расширяет и объясняет, но не отменяет их.

---

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
| Frontend Dev #1 | `fe1` | `.cursor/project_skills/FE_skill.md` | Реализация модуля/страницы A |
| Frontend Dev #2 | `fe2` | `.cursor/project_skills/FE_skill.md` | Реализация модуля/страницы B (параллельно A) |
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

## 3. Компоненты React 19

### 3.1. Базовая форма компонента

Только функциональные компоненты. Пропсы типизируем явно — либо инлайн-объект, либо `interface Props`.

```tsx
interface IssueCardProps {
	issue: IssueBriefResponse;
	onSelect: (issueKey: string) => void;
	isSelected?: boolean;
}

export function IssueCard({ issue, onSelect, isSelected = false }: IssueCardProps) {
	return (
		<button
			type="button"
			onClick={() => onSelect(issue.issue_key)}
			aria-pressed={isSelected}
			className="w-full rounded-[var(--r2)] border border-[var(--b1)] bg-[var(--card)] p-3 text-left"
		>
			<span className="text-[13px] font-medium text-[var(--t1)]">{issue.title}</span>
		</button>
	);
}
```

В Hub встречается и инлайн-типизация пропсов прямо в сигнатуре (см. `IssueForm`) — это допустимо для небольших списков пропсов:

```tsx
export function IssueForm({
	projectId,
	onSuccess,
	initialData,
	mode,
}: {
	projectId: string;
	onSuccess?: (issue: IssueResponse) => void;
	initialData?: IssueResponse;
	mode: "create" | "edit";
}) {
	// ...
}
```

**Правило:** до ~4 пропсов — можно инлайн; больше — выноси в именованный `interface`, чтобы тип переиспользовался и читался.

### 3.2. Rules of Hooks — все хуки ДО любого return

Это **самый частый источник багов**, который ловит `fe-ver`. Все вызовы хуков должны идти на верхнем уровне компонента, безусловно, и **до** любого раннего `return`.

<good-example>
Все хуки вызываются до условного выхода:

```tsx
export function IssuePanel({ issueId }: { issueId: string | null }) {
	// Хуки — первыми, всегда, безусловно
	const { data: issue } = useQuery({
		queryKey: ["issue", issueId],
		queryFn: () => getIssueById(issueId as string),
		enabled: !!issueId, // управляем выполнением через enabled, не через return
	});
	const setSelected = useBoardStore((s) => s.setActiveBoard);

	// Условный выход — ПОСЛЕ всех хуков
	if (!issueId) return null;
	if (!issue) return <PanelSkeleton />;

	return <IssueDetails issue={issue} />;
}
```
</good-example>

<bad-example>
Хук после раннего return — React выбросит "Rendered fewer hooks than expected" и приложение упадёт:

```tsx
export function IssuePanel({ issueId }: { issueId: string | null }) {
	if (!issueId) return null; // ранний выход

	// Этот хук вызывается НЕ на всех рендерах -> нарушение Rules of Hooks
	const { data: issue } = useQuery({
		queryKey: ["issue", issueId],
		queryFn: () => getIssueById(issueId),
	});

	return <IssueDetails issue={issue} />;
}
```
</bad-example>

**Ключевой приём:** вместо раннего `return` для управления запросом используем флаг `enabled` в TanStack Query. Компонент всегда вызывает `useQuery`, а сам запрос стартует только при валидных параметрах. Так сделано в `useMessages` (`enabled: !!channelId`).

### 3.3. Стабильные ключи списков (никогда не `key={index}`)

<good-example>
Ключ — стабильный `id` сущности:

```tsx
{columns.map((col) => (
	<BoardColumn key={col.status.id} status={col.status} issues={col.issues} />
))}
```
</good-example>

<bad-example>
Индекс как ключ ломает реконсиляцию при вставке/удалении/сортировке:

```tsx
{columns.map((col, idx) => (
	<BoardColumn key={idx} status={col.status} issues={col.issues} />
))}
```
</bad-example>

**Исключение:** для статических скелетонов, где элементы не переставляются и не удаляются, допустим составной строковый ключ (не голый индекс):

```tsx
{[1, 2, 3, 4, 5, 6].map((i) => (
	<div key={`board-col-sk-${i}`} className="h-28 animate-shimmer" />
))}
```

Это реальный паттерн из `board/index.tsx`: ключ уникален и не является просто `key={i}` от `.map((_, i) =>)`.

### 3.4. Композиция вместо «мега-компонентов»

Дели по ответственности: страница-контейнер грузит данные, дочерние — рендерят. Пример: `BoardPage` (маршрут) грузит доску и мутации, `BoardColumn` рендерит колонку, `IssueCard` — карточку. Не пиши компонент на 800 строк с 15 `useState` и вложенной логикой.

**Признаки, что пора дробить:**
- больше ~5 `useState`/`useRef` в одном компоненте;
- JSX глубже 5–6 уровней вложенности;
- один компонент грузит несколько несвязанных доменов данных.

### 3.5. Controlled vs uncontrolled

- **Controlled** (значение хранит React) — для полей, где нужен live-контроль: поиск с фильтрацией, форма с валидацией onChange.

```tsx
<Input
	value={filters.search ?? ""}
	onChange={(e) => setFilters({ search: e.target.value || undefined })}
/>
```

- **Uncontrolled** через `react-hook-form` `register` — для больших форм: RHF держит значения вне React-стейта, меньше ререндеров.

```tsx
<Input id="title" {...register("title")} />
```

**Правило:** не смешивай на одном поле `value`+`onChange` с `register`. Для RHF-полей, которым нужен контроль (Select, кастомные пикеры), используй `<Controller>` — так сделано в `IssueForm` для `issue_type`, `priority`, `assignee_id`.

### 3.6. Suspense и `use()`

React 19 добавляет `use()` для чтения промисов/контекста. В Hub основной способ загрузки — TanStack Query, поэтому `use()` для данных применяем ограниченно. Suspense используем для code-splitting (`lazy`) и для Query с `useSuspenseQuery` — для критичных данных маршрута это канон в связке с loader/`ensureQueryData` (2026, см. раздел 4.4).

```tsx
const CallsRoom = lazy(() => import("@/modules/calls/calls-room"));

function CallsRoute() {
	return (
		<Suspense fallback={<CallsRoomSkeleton />}>
			<CallsRoom />
		</Suspense>
	);
}
```

**Правило:** не превращай весь app в Suspense-водопад. Suspense — для тяжёлых lazy-чанков и там, где skeleton логичнее ручного `isLoading`. Для большинства списков остаётся привычный паттерн `useQuery` + `QueryStateView`/`isLoading`.

### 3.7. `memo` / `useMemo` / `useCallback` — только когда реально нужно

**React Compiler (2026):** v1.0 production-stable с октября 2025, но в Hub **не включён**. Статус — «пилот за ADR»: включение (сначала на одном route-subtree, например `tracker`) — только после принятого решения `arch` с профилированием; до этого не полагайся на компилятор и не ослабляй ручную мемоизацию. Мемоизация — инструмент против **доказанной** проблемы, а не рефлекс.

**Когда `useMemo` оправдан:**
- дорогое вычисление (фильтрация/сортировка большого массива), пересчитываемое на каждый рендер;
- стабильная ссылка на объект/массив, который идёт в зависимости другого хука или в `key`/пропс мемоизированного дочернего.

<good-example>
Фильтрация колонок доски мемоизирована — пересчёт только при смене данных или фильтров:

```tsx
const filteredColumns = useMemo(() => {
	if (!boardData?.columns) return [];
	return boardData.columns.map((col) => ({
		...col,
		issues: (col.issues ?? []).filter((issue) => matchesFilters(issue, filters)),
	}));
}, [boardData?.columns, filters]);
```
</good-example>

**Когда `useCallback` оправдан:**
- колбэк уходит в зависимость `useEffect`/другого хука (иначе эффект перезапускается каждый рендер);
- колбэк передаётся в мемоизированный дочерний компонент.

<good-example>
`handleDrop` в зависимости эффекта drag-and-drop — стабилизируем ссылку:

```tsx
const handleDrop = useCallback(
	(issueKey: string, statusId: string) => {
		moveCardMutation.mutate({ issueKey, statusId, rank: "0|aaaaaa:" });
	},
	[moveCardMutation],
);
```
</good-example>

<bad-example>
Мемоизация примитива и тривиального колбэка — только шум и лишняя работа:

```tsx
const title = useMemo(() => issue.title, [issue.title]); // бессмысленно
const onClick = useCallback(() => setOpen(true), []); // если не уходит в deps/мемо-дочерний — не нужно
```
</bad-example>

**Правило `fe-ver`:** каждый `memo`/`useMemo`/`useCallback` должен иметь причину (дорогое вычисление, стабильность ссылки для deps/мемо-дочернего). Мемоизация «на всякий случай» — замечание.

### 3.8. Эффекты: правильные зависимости и очистка

- Указывай **все** реально используемые значения в массиве зависимостей (Biome/ESLint это проверяет).
- Всегда возвращай функцию очистки для подписок, таймеров, слушателей.

<good-example>
Подписка на WS-комнату с корректной очисткой (из `useMessages`):

```tsx
useEffect(() => {
	if (!connected || !channelId) return;
	const room = `channel:${channelId}`;
	subscribe(room);
	return () => {
		unsubscribe(room);
	};
}, [connected, channelId, subscribe, unsubscribe]);
```
</good-example>

<bad-example>
Нет очистки таймера -> утечка и «двойные» срабатывания:

```tsx
useEffect(() => {
	setTimeout(() => setDebounced(value), 300);
	// забыли clearTimeout -> старые таймеры продолжают тикать
}, [value]);
```
</bad-example>

Правильный debounce — `useDebouncedValue` с `clearTimeout` в cleanup (раздел 11).

### 3.9. React 19 Actions: `useActionState` / `useOptimistic` / `useFormStatus` (2026)

Actions-триада React 19 стабильна и работает в Vite-SPA с обычными async-функциями (Server Actions не нужны). Где что уместно:

| Хук | Что даёт | Когда в Hub |
|-----|----------|-------------|
| `useActionState` | pending + result async-действия формы | Простые формы без RHF/Zod (одиночное действие: подписка, инвайт) |
| `useOptimistic` | Оптимистичное значение с авто-откатом при завершении action | Локальный компонентный оптимизм: лайк, тоггл, отправка сообщения |
| `useFormStatus` | pending родительской `<form>` из дочернего компонента | Общая submit-кнопка, не знающая о конкретной форме |

<good-example>
Локальный оптимизм через `useOptimistic` — реакция мгновенная, откат при ошибке автоматический:

```tsx
function ReactionButton({ message }: { message: MessageResponse }) {
	const toggleMutation = useToggleReaction(message.channel_id);
	const [optimisticLiked, setOptimisticLiked] = useOptimistic(message.liked_by_me);

	const handleToggle = () => {
		startTransition(async () => {
			setOptimisticLiked(!optimisticLiked); // мгновенно в UI
			await toggleMutation.mutateAsync(message.id); // при ошибке значение откатится само
		});
	};

	return (
		<button type="button" aria-pressed={optimisticLiked} onClick={handleToggle}>
			{optimisticLiked ? "★" : "☆"}
		</button>
	);
}
```
</good-example>

<bad-example>
Ручной `useState` + откат в `catch` — то, что `useOptimistic` делает сам; легко забыть откат:

```tsx
const [liked, setLiked] = useState(message.liked_by_me);
const handleToggle = async () => {
	setLiked(!liked);
	try {
		await toggleReaction(message.id);
	} catch {
		setLiked(liked); // ручной откат: хрупко, при параллельных кликах — рассинхрон
	}
};
```
</bad-example>

**Граница с TanStack Query (важно):** `useOptimistic` — для оптимизма, видимого **в одном месте на экране** (компонентный стейт). Если оптимистичное состояние должно отразиться в **нескольких местах** (доска + счётчики + списки) — остаётся кэш-паттерн Query `onMutate`/snapshot/rollback (раздел 5.5). Не смешивай оба подхода на одних и тех же данных.

Основной стек форм Hub — RHF + Zod (раздел 10): для сложных форм с валидацией он остаётся каноном, `useActionState` его не заменяет.

### 3.10. Чек-лист компонентов

- [ ] Все хуки — до любого `return`; условный выход после хуков.
- [ ] Управление запросом через `enabled`, а не ранний return перед `useQuery`.
- [ ] `key` по стабильному `id`; скелетоны — с составным строковым ключом, не голым индексом.
- [ ] `type="button"` на всех неотправляющих кнопках.
- [ ] `memo`/`useMemo`/`useCallback` — только с обоснованием.
- [ ] Все `useEffect` имеют полный массив зависимостей и cleanup для подписок/таймеров.
- [ ] Крупный компонент разбит по ответственности.

---

## 4. TanStack Router

### 4.1. File-based routing

Маршруты живут в `src/routes/` и соответствуют URL по структуре папок. `routeTree.gen.ts` генерируется плагином автоматически — **руками не редактируем**. Всего в проекте 39 маршрутов.

Структура на примере:

```
routes/
├── __root.tsx                                   # корневой layout
├── login.tsx                                    # /login
├── _authenticated.tsx                           # guard-layout (все защищённые под ним)
└── _authenticated/
    ├── index.tsx                                # /
    ├── messenger/
    ├── calls.tsx                                # /calls
    └── projects/
        ├── index.tsx                            # /projects
        └── $projectId/
            ├── board/index.tsx                  # /projects/:projectId/board
            └── issues/$issueKey.tsx             # /projects/:projectId/issues/:issueKey
```

### 4.2. `createFileRoute`

Каждый файл-маршрут экспортирует `Route`, созданный через `createFileRoute` с точным путём.

```tsx
export const Route = createFileRoute("/_authenticated/projects/$projectId/board/")({
	component: BoardPage,
});
```

Строка пути должна **точно** совпадать с положением файла в дереве. Ошибка в строке -> route не подхватится.

### 4.3. Параметры маршрута

Читаем params через `Route.useParams()` внутри компонента:

```tsx
function BoardPage() {
	const { projectId } = Route.useParams();
	// ...
}
```

Для search-параметров есть `Route.useSearch()` с валидацией через `validateSearch`.

### 4.4. Loader + preloading — канон для критичных данных маршрута (2026)

Loader предзагружает данные до рендера. Для критичных данных маршрута (доска, деталка задачи) канон 2026 — связка Router↔Query: `queryClient.ensureQueryData` в loader + `useSuspenseQuery` в компоненте. Раньше loader считался «опциональным» — теперь для тяжёлых маршрутов это рекомендуемый паттерн; для вторичных данных остаётся `useQuery` внутри компонента.

**Настройка router (`main.tsx`)** — intent-preload включает бесплатный prefetch по hover/touch на `<Link>`:

```tsx
const router = createRouter({
	routeTree,
	context: { queryClient }, // queryClient через createRootRouteWithContext
	defaultPreload: "intent",
	// ОБЯЗАТЕЛЬНО при интеграции с Query: 0 отдаёт управление свежестью
	// staleTime самого Query, иначе 30s-кэш роутера ломает его семантику.
	defaultPreloadStaleTime: 0,
});
```

**Канон-паттерн loader + suspense:**

```tsx
const issueQueryOptions = (issueKey: string) =>
	queryOptions({
		queryKey: ["issue", issueKey],
		queryFn: () => getIssueByKey(issueKey),
		staleTime: 30_000,
	});

export const Route = createFileRoute("/_authenticated/projects/$projectId/issues/$issueKey")({
	component: IssueDetailPage,
	loader: ({ context: { queryClient }, params }) =>
		queryClient.ensureQueryData(issueQueryOptions(params.issueKey)),
});

function IssueDetailPage() {
	const { issueKey } = Route.useParams();
	// data гарантированно не-undefined на уровне типов; fallback — Suspense/Error boundary маршрута
	const { data: issue } = useSuspenseQuery(issueQueryOptions(issueKey));
	return <IssueDetails issue={issue} />;
}
```

**Правила:**
- loader и Query не дублируют друг друга: единый кэш через `ensureQueryData` + общие `queryOptions`.
- `defaultPreloadStaleTime: 0` обязателен при Query — свежестью управляет `staleTime` Query, не роутер.
- Ловушка waterfall: два `useSuspenseQuery` в одном компоненте фетчат **последовательно** — для нескольких критичных запросов используй `useSuspenseQueries` (или предзагружай оба в loader).
- Вторичные/некритичные данные — обычный `useQuery` + `isLoading`/`QueryStateView` (раздел 5), не тяни всё в suspense.

### 4.5. Protected routes через `_authenticated`

Ключевой паттерн Hub. Layout `_authenticated.tsx` проверяет аутентификацию в `beforeLoad` и редиректит на `/login`, если пользователь не залогинен. Все защищённые маршруты — дети этого layout.

```tsx
export const Route = createFileRoute("/_authenticated")({
	beforeLoad: () => {
		const { isAuthenticated } = useAuthStore.getState();
		if (!isAuthenticated) {
			throw redirect({ to: "/login" });
		}
	},
	component: AuthenticatedLayout,
});
```

Обрати внимание: в `beforeLoad` читаем стор через `useAuthStore.getState()` (не хук — это не компонент). `AuthenticatedLayout` дополнительно поднимает WebSocket (`useWebSocket()`) и multi-tab sync — то есть всё, что нужно только залогиненному пользователю.

### 4.6. Навигация

Декларативно — `<Link>`; императивно — `useNavigate()`.

```tsx
import { Link } from "@tanstack/react-router";

<Link to="/projects/$projectId/issues/new" params={{ projectId }} className="...">
	Задача
</Link>
```

```tsx
const navigate = useNavigate();
// ...
void navigate({ to: "/login" });
```

**Правило:** параметры маршрута передаём через `params`, а не конкатенацией строк URL. Типобезопасность TanStack Router проверит существование пути и наличие нужных params на этапе компиляции.

<bad-example>
Строковая сборка URL теряет типобезопасность и легко ломается:

```tsx
<a href={`/projects/${projectId}/issues/new`}>Задача</a>
```
</bad-example>

### 4.7. Code-splitting через lazy

Тяжёлые страницы/модули грузим лениво. В file-based routing применяется суффикс `.lazy.tsx` (например `onboarding.lazy.tsx`) либо `lazy(() => import(...))` для тяжёлого поддерева.

```tsx
const CallsRoom = lazy(() => import("@/modules/calls/calls-room"));
```

Обязательно оборачивай lazy-компонент в `<Suspense fallback={...}>` (см. 3.6) и в Error Boundary (см. 12).

### 4.8. Чек-лист роутинга

- [ ] Путь в `createFileRoute` точно совпадает с расположением файла.
- [ ] Защищённые страницы — под `_authenticated`.
- [ ] Params читаются через `Route.useParams()`.
- [ ] Навигация — `<Link>`/`useNavigate` с `params`, без ручной сборки URL.
- [ ] `routeTree.gen.ts` не редактировался вручную.
- [ ] Тяжёлые маршруты грузятся лениво и обёрнуты в Suspense + Error Boundary.
- [ ] (2026) Критичные данные маршрута — `ensureQueryData` в loader + `useSuspenseQuery`; router с `defaultPreload: 'intent'` и `defaultPreloadStaleTime: 0`.

---

## 5. TanStack Query

TanStack Query — **единственный** держатель server-state. Всё, что приходит с бэка, живёт в кэше Query, а не в `useState`/Zustand.

### 5.1. Разделение server-state vs client-state

| Тип данных | Инструмент | Примеры |
|------------|-----------|---------|
| Пришло с сервера, может устареть | **TanStack Query** | список задач, сообщения, профиль, доска |
| Локальный UI-стейт | **Zustand** | выбранный элемент, открыт ли модал, фильтры, drag-state |
| Данные формы | **react-hook-form** | значения полей до submit |

<bad-example>
Класть серверные данные в Zustand и синхронизировать вручную — источник рассинхрона:

```tsx
// ПЛОХО: дублируем серверный список в Zustand
const issues = useIssuesStore((s) => s.issues);
useEffect(() => {
	listIssues(projectId).then((r) => setIssues(r.items)); // ручная синхронизация, нет кэша, нет invalidation
}, [projectId]);
```
</bad-example>

<good-example>
Серверные данные — только через Query:

```tsx
const { data } = useQuery({
	queryKey: ["issues", projectId, filters],
	queryFn: () => listIssues(projectId, filters),
	staleTime: 30_000,
});
```
</good-example>

### 5.2. Конвенции `queryKey`

Ключ — массив от общего к частному: `[домен, сущность, ...идентификаторы, ...параметры]`. Так работает частичная инвалидация.

```typescript
["issues", projectId, filters]                        // список задач проекта с фильтрами
["issue", projectId, issueKey]                        // одна задача
["board", projectId, activeBoardId]                   // данные доски
["boards", projectId]                                 // список досок
["messenger", "messages", channelId]                  // сообщения канала
["messenger", "thread", channelId, messageId]         // ответы в треде
["messenger", "channels"]                             // список каналов
["projects"]                                          // все проекты
```

**Правила ключей:**
- Первый элемент — домен (`"messenger"`, `"tracker"` в виде корневых имён сущностей).
- Идентификаторы — в стабильном порядке (родитель раньше ребёнка).
- Параметры (фильтры, cursor) — последними; при их смене Query автоматически перезапросит.
- Ключ, по которому инвалидируешь, должен быть **префиксом** используемых ключей (см. 5.4).

### 5.3. `staleTime` и базовая настройка запроса

`staleTime` определяет, сколько данные считаются свежими (не рефетчатся при повторном монтировании/фокусе). Подбирай по природе данных.

| Данные | staleTime | Почему |
|--------|-----------|--------|
| Редко меняющийся справочник (workflow statuses, labels) | `5 * 60_000` | Меняются редко |
| Список задач/досок | `30_000` | Компромисс свежести/нагрузки |
| Реальновременные (сообщения) | `0` + обновление из WS | Свежесть критична, апдейты приходят через WS |

```tsx
const { data: issue } = useQuery({
	queryKey: ["issues", issueId],
	queryFn: () => issueApi.getById(issueId),
	staleTime: 30_000,
});
```

### 5.4. Mutations + invalidateQueries

После мутации инвалидируем затронутые ключи, чтобы Query перезапросил актуальные данные.

```tsx
const createMutation = useMutation({
	mutationFn: (data: CreateIssueRequest) => createIssue(projectId, data),
	onSuccess: (result) => {
		void queryClient.invalidateQueries({ queryKey: ["board", projectId] });
		void queryClient.invalidateQueries({ queryKey: ["issues", projectId] });
		toast.success(t("form.createSuccess"));
		onSuccess?.(result);
	},
	onError: (err) => toastApiError(err, t("form.createError")),
});
```

**Правила:**
- Инвалидируй **минимально достаточный** набор ключей. `["board", projectId]` инвалидирует все доски проекта; `["issues", projectId]` — списки задач.
- Оборачивай `invalidateQueries` в `void`, если не ждёшь промис (Biome требует явно обозначить проигнорированный промис).
- Ошибку показывай через `toastApiError` (общий помощник, разбирает `ApiError`), а не голым `alert`.

### 5.5. Optimistic updates + rollback

Для мгновенного UX (drag-and-drop, лайки) применяем оптимистичное обновление: правим кэш до ответа сервера, при ошибке откатываем. Канонический паттерн из `board/index.tsx`:

```tsx
const moveCardMutation = useMutation({
	mutationFn: (data: { issueKey: string; statusId: string; rank: string }) =>
		moveIssue(projectId, data.issueKey, { status_id: data.statusId, rank: data.rank }),

	onMutate: async (data) => {
		// 1. Останавливаем текущие рефетчи, чтобы они не перезаписали оптимистичное состояние
		await queryClient.cancelQueries({ queryKey: ["board", projectId, activeBoardId] });

		// 2. Снимаем снапшот для отката
		const previous = queryClient.getQueryData<BoardDataResponse>([
			"board", projectId, activeBoardId,
		]);

		// 3. Оптимистично обновляем кэш
		queryClient.setQueryData<BoardDataResponse>(
			["board", projectId, activeBoardId],
			(old) => (old ? reorderBoard(old, data) : old),
		);

		// 4. Возвращаем контекст с предыдущим состоянием
		return { previous };
	},

	onError: (err, _data, context) => {
		// 5. Откат при ошибке
		if (context?.previous) {
			queryClient.setQueryData(["board", projectId, activeBoardId], context.previous);
		}
		toastApiError(err, t("kanban.moveError"));
	},

	onSettled: () => {
		// 6. В любом случае синхронизируемся с сервером
		void queryClient.invalidateQueries({ queryKey: ["board", projectId, activeBoardId] });
	},
});
```

**Обязательные шаги optimistic update (все четыре):**
1. `cancelQueries` — иначе фоновый рефетч затрёт оптимистичное состояние.
2. Снапшот `previous` в `onMutate`.
3. Откат в `onError` из `context.previous`.
4. `invalidateQueries` в `onSettled` — финальная сверка с сервером.

<bad-example>
Оптимистичное обновление без снапшота/отката — при 409/сетевой ошибке доска остаётся в неверном состоянии навсегда:

```tsx
onMutate: (data) => {
	queryClient.setQueryData(key, (old) => reorder(old, data)); // нет previous, нет onError-rollback
},
```
</bad-example>

> (2026) Для оптимизма, видимого только в одном месте на экране (лайк, тоггл), есть более лёгкий `useOptimistic` из React 19 — см. раздел 3.9. Кэш-паттерн выше остаётся каноном, когда оптимистичное состояние должно отразиться в нескольких местах кэша.

### 5.6. Обработка ошибок

- Мутации: `onError` -> `toastApiError(err, fallbackMessage)`.
- Запросы: показываем через состояние (`isError`, `error`) и общий `QueryStateView` (в Hub есть компонент-обёртка, который отображает loading/empty/error).

```tsx
<QueryStateView
	query={boardQuery}
	loadingFallback={boardLoadingFallback}
	isEmpty={() => filteredColumns.length === 0}
	emptyFallback={<EmptyBoard projectId={projectId} />}
>
	{() => <BoardColumns columns={filteredColumns} />}
</QueryStateView>
```

Структура серверной ошибки фиксирована ТЗ: `{ "error": message, "request_id": uuid }`. Клиентский `ApiError` (`api/client.ts`) содержит `code`, `message`, `request_id?`, `retry_after?`. `request_id` полезен для отладки — прокидывай его в тост/лог при необходимости.

### 5.7. Refetch, `enabled`, зависимые запросы

**`enabled`** — управляет тем, стартует ли запрос. Обязателен, когда параметр может быть `undefined`:

```tsx
const boardQuery = useQuery({
	queryKey: ["board", projectId, activeBoardId],
	queryFn: () => getBoardData(projectId, activeBoardId as string),
	enabled: !!activeBoardId, // не запускаем без активной доски
});
```

**Зависимые запросы** — второй запрос ждёт данные первого через `enabled`:

```tsx
const { data: project } = useQuery({
	queryKey: ["project", projectId],
	queryFn: () => getProject(projectId),
});

const { data: members } = useQuery({
	queryKey: ["project", projectId, "members"],
	queryFn: () => getMembers(project!.id),
	enabled: !!project, // ждём загрузки проекта
});
```

**Refetch** — ручной `query.refetch()` (кнопка «Обновить») или автоматический по инвалидации/фокусу. Для infinite scroll — `fetchNextPage` (см. 5.8).

### 5.8. Infinite queries (курсорная пагинация)

Для чатов и длинных лент — `useInfiniteQuery` с курсором. Канон из `useMessages`:

```tsx
const query = useInfiniteQuery({
	queryKey: ["messenger", "messages", channelId],
	queryFn: ({ pageParam }) => getMessages(channelId!, { before: pageParam, limit: PAGE_SIZE }),
	initialPageParam: undefined as string | undefined,
	getNextPageParam: (lastPage) => {
		if (lastPage.length < PAGE_SIZE) return undefined; // страниц больше нет
		return lastPage[0]?.created_at; // курсор — created_at первого элемента (скролл вверх)
	},
	enabled: !!channelId,
});

const messages = query.data?.pages.flat() ?? [];
```

Бэкенд ожидает cursor в RFC3339 для infinite scroll вверх. `getNextPageParam` возвращает `undefined`, когда данных больше нет — это останавливает пагинацию.

### 5.9. Прямое обновление кэша из WS (setQueryData)

Реальновременные события правят кэш Query напрямую, без рефетча. Подробно — в разделе 13, но ключевая идея: `queryClient.setQueryData(key, updater)` с иммутабельным обновлением и защитой от дублей по `id`.

### 5.10. Чек-лист Query

- [ ] Server-state только в Query; client-state в Zustand; данные формы в RHF.
- [ ] `queryKey` по конвенции (домен -> id -> параметры).
- [ ] `staleTime` подобран под природу данных.
- [ ] Мутации инвалидируют минимально достаточный набор ключей.
- [ ] Optimistic update содержит все 4 шага (cancel/snapshot/rollback/settled invalidate).
- [ ] `enabled` для запросов с потенциально `undefined` параметрами.
- [ ] Ошибки мутаций -> `toastApiError`; запросов -> `QueryStateView`/состояние.
- [ ] Промисы `invalidateQueries`/`navigate` помечены `void`, если не ожидаются.

---

## 6. Zustand

Zustand держит **client-state**: то, что не приходит с сервера. В проекте 11+ сторов: `auth-store`, `ws-store`, `board-store`, `messenger-store`, `notification-store`, `timer-store`, `call-store`, `command-store`, `sidebar-store`, `ai-store`, `dashboard-store`, `sprint-store`, `tms-store`.

### 6.1. Базовый паттерн стора

Интерфейс состояния (данные + действия), затем `create<T>()`.

```typescript
import { create } from "zustand";

interface BoardState {
	activeBoardId: string | null;
	filters: BoardFilters;
	isDragging: boolean;

	setActiveBoard: (boardId: string) => void;
	setFilters: (filters: Partial<BoardFilters>) => void;
	resetFilters: () => void;
	setDragging: (isDragging: boolean, issueId?: string | null) => void;
}

const defaultFilters: BoardFilters = {};

export const useBoardStore = create<BoardState>()((set) => ({
	activeBoardId: null,
	filters: defaultFilters,
	isDragging: false,

	setActiveBoard: (boardId) => set({ activeBoardId: boardId }),
	setFilters: (filters) => set((state) => ({ filters: { ...state.filters, ...filters } })),
	resetFilters: () => set({ filters: defaultFilters }),
	setDragging: (isDragging, issueId = null) => set({ isDragging, draggingIssueId: issueId }),
}));
```

**Правила:**
- Действия обновляют состояние **иммутабельно** (`{ ...state.filters, ...filters }`), не мутируют объект.
- Функция `set((state) => ...)` — когда новое состояние зависит от текущего; `set({ ... })` — когда нет.
- `get()` — для чтения актуального состояния внутри действий (см. `ws-store`).

### 6.2. Селекторы: минимизация ререндеров

**Критично для производительности.** Подписывайся на минимально нужный срез состояния, а не на весь стор.

<good-example>
Селектор -> ререндер только при смене `accessToken`:

```tsx
const accessToken = useAuthStore((s) => s.accessToken);
const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
```
</good-example>

<bad-example>
Деструктуризация всего стора -> ререндер на ЛЮБОЕ изменение стора:

```tsx
const { accessToken } = useAuthStore(); // подписка на весь стор
```
</bad-example>

**Нюанс.** В Hub встречается и деструктуризация всего стора для действий (`useWsStore()` в `useWebSocket`), потому что действия стабильны и берутся один раз. Но для **часто меняющихся данных** всегда используй узкий селектор. Если нужно несколько полей — бери несколько отдельных селекторов или используй `useShallow` для поверхностного сравнения:

```tsx
import { useShallow } from "zustand/react/shallow";

const { subscribe, unsubscribe } = useWsStore(
	useShallow((s) => ({ subscribe: s.subscribe, unsubscribe: s.unsubscribe })),
);
```

### 6.3. Действия и вызов вне компонентов

Zustand-стор доступен вне React через `useStore.getState()`. Это используется в guard-ах маршрутов и в API-клиенте.

```typescript
// В beforeLoad маршрута (не компонент — хук нельзя)
const { isAuthenticated } = useAuthStore.getState();

// В API-клиенте: чтение токена и logout
const token = useAuthStore.getState().accessToken;
useAuthStore.getState().logout();
```

**Правило:** внутри компонента — хук `useStore(selector)`; вне компонента (loader, api, утилита) — `useStore.getState()`.

### 6.4. Persist

Middleware `persist` сохраняет стор в storage. Auth-стор персистится в **sessionStorage** (сессия завершается при закрытии вкладки):

```typescript
export const useAuthStore = create<AuthState>()(
	persist(
		(set) => ({
			accessToken: null,
			// ...
			setTokens: (access, refresh) => {
				set({ accessToken: access, refreshToken: refresh, isAuthenticated: true });
				broadcastAuth({ type: "login", userId: null });
			},
			logout: () => {
				set({ accessToken: null, refreshToken: null, user: null, isAuthenticated: false });
				broadcastAuth({ type: "logout" });
			},
		}),
		{ name: "hub-auth", storage: createJSONStorage(() => sessionStorage) },
	),
);
```

**Правила persist:**
- Выбор storage осознанный: `sessionStorage` для сессии (auth), `localStorage` для настроек, переживающих перезапуск (тема, язык — но язык уже держит i18next-detector в localStorage).
- Не персисти чувствительное дольше необходимого. Токены — sessionStorage, не localStorage.
- Не персисти server-state (это работа Query) и эфемерное (`isDragging`, `socket`). `ws-store` намеренно **без** persist — соединение пересоздаётся при каждой загрузке.
- Для эволюции формы стора используй `version` + `migrate`.

### 6.5. Multi-tab синхронизация (продвинутый паттерн)

Auth-стор транслирует login/logout между вкладками через `BroadcastChannel` (FE-H2). Это не обязательный паттерн для каждого стора, но эталон для критичных состояний:

```typescript
function broadcastAuth(event: AuthSyncEvent): void {
	if (!authChannel) return;
	try {
		authChannel.postMessage(event);
	} catch {
		// канал мог закрыться — не критично
	}
}
```

Подписка (`subscribeAuthSync`) вешается в `_authenticated.tsx`: при `logout` в другой вкладке — чистим Query-кэш и редиректим на `/login`.

### 6.6. Срезы (slices) для крупных сторов

Когда стор разрастается, дели его на слайсы — функции, возвращающие часть состояния, объединяемые в `create`. Это держит один логический стор, но модульно.

```typescript
interface FiltersSlice {
	filters: BoardFilters;
	setFilters: (f: Partial<BoardFilters>) => void;
}

const createFiltersSlice: StateCreator<BoardState, [], [], FiltersSlice> = (set) => ({
	filters: {},
	setFilters: (f) => set((s) => ({ filters: { ...s.filters, ...f } })),
});

// Композиция в create<BoardState>()((...a) => ({ ...createFiltersSlice(...a), ...createDragSlice(...a) }))
```

**Правило:** не плоди 20 микро-сторов там, где логично один доменный стор со слайсами. И наоборот — не суй несвязанные домены в один гигантский стор.

### 6.7. Чек-лист Zustand

- [ ] В сторе только client-state; server-state — в Query.
- [ ] Обновления иммутабельны (`{ ...state, ... }`), без мутаций.
- [ ] Подписка через узкий селектор; для данных, меняющихся часто — не деструктуризация всего стора.
- [ ] Вне компонента — `getState()`, внутри — `useStore(selector)`.
- [ ] `persist` только для нужного; чувствительное — sessionStorage; эфемерное не персистится.
- [ ] Крупный стор разбит на слайсы, а не на десяток микро-сторов.

---

## 7. API-слой

### 7.1. Один файл на backend-модуль

`api/<module>.ts` — граница FE/BE. В нём: типы Request/Response DTO + функции запросов. Один файл на backend-модуль (`tracker.ts`, `messenger.ts`, `auth.ts`, `tms.ts`, ...). Общий низкоуровневый клиент — `api/client.ts`, общие типы — `api/types.ts`.

Функции — простые обёртки над `api.get/post/put/patch/delete`, типизированные дженериком ответа:

```typescript
import { api } from "./client";
import type { PaginatedResponse } from "./types";

export function createIssue(projectId: string, data: CreateIssueRequest) {
	return api.post<IssueResponse>(`/projects/${projectId}/issues`, data);
}

export function listIssues(projectId: string, params?: ListIssuesParams) {
	const qs = buildQueryString(params);
	return api.get<PaginatedResponse<IssueResponse>>(`/projects/${projectId}/issues${qs}`);
}

export function updateIssue(projectId: string, issueKey: string, data: UpdateIssueRequest) {
	return api.put<IssueResponse>(`/projects/${projectId}/issues/${issueKey}`, data);
}
```

**Правила:**
- Каждая функция явно типизирует ответ: `api.get<IssueResponse>(...)`.
- Query-строку собираем через общий `buildQueryString(params)` (`@/lib/utils`), не конкатенацией руками.
- DTO именуем по роли: `CreateIssueRequest`, `UpdateIssueRequest`, `IssueResponse`, `IssueBriefResponse`.
- Функции не знают про React/Query — это чистый слой доступа к данным. Кэширование и состояние — задача Query-хуков поверх.

### 7.2. Базовый клиент `api/client.ts`

Кастомный класс `ApiClient` на `fetch`. Отвечает за: базовый префикс `/api`, заголовки, Authorization из auth-стора, обработку 401, 429, 204/пустого тела.

```typescript
class ApiClient {
	private getHeaders(): Record<string, string> {
		const headers: Record<string, string> = { "Content-Type": "application/json" };
		const token = useAuthStore.getState().accessToken;
		if (token) headers.Authorization = `Bearer ${token}`;
		return headers;
	}

	async request<T>(method: string, path: string, body?: unknown): Promise<T> {
		const response = await fetch(`${API_BASE}${path}`, {
			method,
			headers: this.getHeaders(),
			body: body ? JSON.stringify(body) : undefined,
		});

		if (response.status === 401) {
			const refreshed = await this.tryRefresh();
			if (refreshed) return this.request(method, path, body);
			useAuthStore.getState().logout();
			throw new Error("Session expired");
		}

		if (!response.ok) {
			// разбор ApiError + Retry-After для 429
			// ...
			throw error;
		}

		if (response.status === 204) return undefined as T;
		const text = await response.text();
		if (!text) return undefined as T;
		return JSON.parse(text) as T;
	}
}

export const api = new ApiClient();
```

### 7.3. Ключевой паттерн: single-flight refresh на 401 (FE-H1)

**Самый важный паттерн auth.** При параллельных 401-ответах несколько запросов могут одновременно попытаться обновить токен. Первый успешный refresh инвалидирует старый refresh-токен, и остальные refresh-запросы получают 401 -> пользователя выкидывает.

Решение — модульный promise-lock (mutex): пока refresh в работе, все последующие вызовы ждут **тот же** промис.

```typescript
// Модульный single-flight lock
let refreshing: Promise<boolean> | null = null;

private tryRefresh(): Promise<boolean> {
	if (refreshing) return refreshing; // уже идёт — переиспользуем промис
	refreshing = this.doRefresh().finally(() => {
		refreshing = null; // сброс по завершении — следующий 401 запустит новый refresh
	});
	return refreshing;
}

private async doRefresh(): Promise<boolean> {
	const refreshToken = useAuthStore.getState().refreshToken;
	if (!refreshToken) return false;
	try {
		const res = await fetch(`${API_BASE}/auth/refresh`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ refresh_token: refreshToken }),
		});
		if (!res.ok) return false;
		const data = (await res.json()) as { access_token: string; refresh_token: string };
		useAuthStore.getState().setTokens(data.access_token, data.refresh_token);
		return true;
	} catch {
		return false;
	}
}
```

**Логика на 401 в `request`:** пробуем refresh; при успехе — повторяем исходный запрос; при неудаче — `logout()` + throw. Это связано с ТЗ: access-токен 15 минут, refresh 30 дней с instant revocation. Нарушение single-flight -> массовые ложные разлогины.

<bad-example>
Refresh без mutex — гонка на параллельных 401:

```typescript
if (response.status === 401) {
	await this.doRefresh(); // каждый 401 запускает свой refresh -> гонка -> ложный logout
	return this.request(method, path, body);
}
```
</bad-example>

### 7.4. `PaginatedResponse<T>`

Единый тип пагинации поддерживает и cursor-, и offset-режимы (`api/types.ts`):

```typescript
export interface PaginatedResponse<T> {
	items: T[];
	next_cursor?: string;   // cursor-пагинация (чат, infinite scroll)
	has_more?: boolean;
	total: number;
	offset?: number;        // offset-пагинация (задачи, таблицы)
	limit?: number;
}
```

ТЗ: задачи — offset/limit, чат — cursor. Тип покрывает оба; используй нужные поля.

### 7.5. Ошибки, `request_id`, коды

Клиентский `ApiError`:

```typescript
export interface ApiError {
	code: number;         // HTTP-код
	message: string;      // человекочитаемое сообщение
	request_id?: string;  // для отладки/логов
	retry_after?: number; // секунды до повтора (429, из Retry-After)
}
```

- 429 -> читаем заголовок `Retry-After`, кладём в `retry_after` — UI показывает «повторите через N сек».
- `request_id` — прокидывай в тост/лог при отладке, чтобы сматчить с backend-логом.
- 204/пустое тело -> `undefined` (не пытайся `JSON.parse("")`).

### 7.6. Отмена запросов

Для запросов, которые могут устареть (быстрый набор в поиске, смена вкладки), используем отмену. TanStack Query передаёт `AbortSignal` в `queryFn`:

```typescript
useQuery({
	queryKey: ["search", query],
	queryFn: ({ signal }) => searchApi.search(query, { signal }),
});
```

Чтобы это работало, низкоуровневый клиент должен прокидывать `signal` в `fetch`. Для новых endpoint-ов, где отмена важна, добавляй перегрузку с опциями (`{ signal }`) и прокидывай в `fetch({ signal })`. Для optimistic-мутаций отмена конкурирующих рефетчей делается через `queryClient.cancelQueries` (раздел 5.5), а не через AbortSignal.

**Debounce vs cancel:** для поиска сначала `useDebouncedValue` (меньше запросов), затем — отмена устаревших (актуальность последнего).

### 7.7. Чек-лист API-слоя

- [ ] Новые запросы — в `api/<module>.ts`, типизированы дженериком ответа.
- [ ] Query-строка через `buildQueryString`, не ручная конкатенация.
- [ ] DTO именованы по роли (`*Request`/`*Response`), совпадают с backend-контрактом.
- [ ] Не трогали single-flight refresh mutex без явной причины и тестов.
- [ ] Пагинация использует `PaginatedResponse<T>` с нужным режимом (cursor/offset).
- [ ] 429 обрабатывает `Retry-After`; 204/пустое тело -> `undefined`.
- [ ] Где нужна свежесть последнего запроса — debounce + отмена устаревших.

---

## 8. Дизайн-система на практике

> Полный визуальный язык, палитра, spacing и компоненты — в `.cursor/playbooks/design.md`. Здесь — практика применения во фронтенд-коде.

### 8.1. CSS-переменные `--hub-*` и алиасы

Единственный источник цветов/поверхностей — дизайн-токены. Не хардкодь hex в компонентах.

Группы токенов (из `30-frontend.mdc`):

```
Поверхности: --hub-void, --hub-base, --hub-raised, --hub-card, --hub-input
Текст:       --hub-text-primary / secondary / muted / disabled
Границы:     --hub-border, --hub-border-subtle, --hub-border-strong
```

В коде Hub встречаются короткие алиасы токенов (`--t1`, `--t2`, `--t3`, `--b1`, `--card`, `--input`, `--r2`, `--blueS`, `--gB`) — это тот же слой токенов. Применяются через Tailwind arbitrary values или семантические классы (`text-t1`, `bg-card`, `border-b1`).

**(2026) Новые дизайн-токены — через `@theme` (Tailwind 4 CSS-first).** Токен, объявленный в `@theme`, автоматически даёт и CSS-переменную, и чистую утилиту (`bg-card` вместо `bg-[var(--card)]`):

```css
@theme {
	--color-card: #191a20; /* даёт утилиты bg-card, text-card, border-card */
}
```

Правило: новые токены объявляй в `@theme`; для стандартных токенов постепенно предпочитай чистые утилиты вместо arbitrary values (`bg-[var(--card)]` в существующем коде не переписываем массово). Кастомные утилиты/варианты — через `@utility`/`@custom-variant`, не через JS-plugins.

```tsx
// Через CSS-переменную в arbitrary value
<div className="bg-[var(--hub-raised)] border border-[var(--hub-border)]" />

// Через семантический alias-класс
<span className="text-t1" />
<div className="bg-card border border-b1 rounded-[var(--r2)]" />
```

### 8.2. Tailwind-утилиты — основной способ стилизации

Стилизуем через Tailwind-классы. Никаких отдельных CSS-файлов «на компонент» без веской причины.

<good-example>
Токены + утилиты:

```tsx
<div className="flex items-center gap-2.5 px-6 h-[50px] border-b border-[var(--b1)] shrink-0">
```
</good-example>

<bad-example>
Inline-стиль поверх дизайн-системы (запрещён `10-style.mdc`):

```tsx
<div style={{ background: "#1a1a1a", borderColor: "#333" }} />
```
</bad-example>

### 8.3. Shadcn — только `className`, поведение Radix не трогать

Компоненты `components/ui/*` построены на Radix. **Правило:** меняем только визуал через `className`; пропсы/поведение Radix (открытие, фокус-менеджмент, роли) не переопределяем.

<good-example>
Расширяем вид кнопки через className:

```tsx
<Button className="bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_20px_var(--gB)]">
	Задача
</Button>
```
</good-example>

<bad-example>
Хак внутренностей Radix (переопределение обработчиков/data-состояний, форк логики Dialog) — ломает доступность и обновляемость:

```tsx
<DialogPrimitive.Content onEscapeKeyDown={(e) => e.preventDefault()} /> // ломает Esc-закрытие без причины
```
</bad-example>

Если нужен новый вариант компонента ui — добавляй его как проп-вариант (cva) в самом ui-компоненте отдельной согласованной задачей, а не хакай в потребителе.

### 8.4. Запрет inline-стилей и `!important`

- `style={{}}` поверх дизайн-системы — запрещён. Исключение — вычисляемые в рантайме величины, которые нельзя выразить классом (например, позиция драг-оверлея по координатам курсора). Но цвета/типографика — только токены.
- `!important` в CSS — **запрещён** (`60-tz-constraints.mdc`: Phase 8 удалил 56 штук). Если стиль «не перебивается» — проблема в специфичности/порядке слоёв, решай через Tailwind-слои, а не `!important`.

### 8.5. Responsive

Mobile-first: базовые классы для мобильного, `sm:`/`md:`/`lg:` для больших экранов.

```tsx
<Input className="w-full sm:w-[220px] h-[34px] ..." />
```

Для навигации есть отдельные компоненты (`mobile-nav.tsx`). Не дублируй логику — переиспользуй адаптивные обёртки.

**(2026) Container queries для компонентной отзывчивости.** Viewport-брейкпоинты отвечают на вопрос «какой экран», но не «сколько места у компонента». Для компонентов внутри resizable-панелей и split-view (messenger, board, сайдбары) правильный примитив — container queries Tailwind 4 (first-class, без плагина):

```tsx
<div className="@container">
	{/* раскладка зависит от ширины ПАНЕЛИ, а не окна */}
	<div className="grid grid-cols-1 @md:grid-cols-2 gap-3">...</div>
</div>
```

Правило: компонент живёт в панели с переменной шириной — `@container`/`@sm:`/`@md:`; раскладка страницы целиком — viewport-брейкпоинты.

### 8.6. Dark theme

Тема — через токены и класс на корне (`data-theme`/`.dark`). Поскольку цвета берутся из `--hub-*`, поддержка тёмной темы автоматическая: переменные переопределяются на уровне темы. **Не хардкодь** «тёмные» цвета в компонентах — это сломает переключение темы. Тумблер темы — в `modules/settings/appearance.tsx`.

### 8.7. Чек-лист дизайн-системы

- [ ] Цвета/поверхности/границы — только токены `--hub-*` (или их алиасы), без hex в компонентах.
- [ ] Стилизация через Tailwind-утилиты.
- [ ] Shadcn — только `className`; поведение Radix не тронуто.
- [ ] Нет inline `style={{}}` поверх дизайн-системы (кроме рантайм-координат).
- [ ] Нет `!important`.
- [ ] Responsive: mobile-first + брейкпоинты.
- [ ] Тёмная тема работает (нет хардкод-цветов, ломающих переключение).

---

## 9. Доступность (a11y, WCAG 2.2 AA)

Hub целится в **WCAG 2.2 AA** *(2026: переход с 2.1; WCAG 2.2 — ISO/IEC 40500:2025, обратно совместим с 2.1)*. `fe-ver` гоняет axe и проверяет клавиатурную навигацию. A11y — не опция.

Новые для 2.2 критерии, критичные для Hub (детали и дизайн-требования — `.cursor/playbooks/design.md` §5.7): **2.5.7 Dragging Movements** (Kanban DnD обязан иметь альтернативу без drag — перемещение через меню/кнопки), **2.5.8 Target Size ≥24×24px** (иконочные кнопки), **2.4.11 Focus Not Obscured** (фокус не перекрыт sticky-панелями). Magic Link уже удовлетворяет 3.3.8 Accessible Authentication; 4.1.1 Parsing упразднён.

### 9.1. Семантика

Используй правильные элементы: `<button>` для действий, `<a>`/`<Link>` для навигации, заголовки `<h1>..<h3>` по иерархии, списки для списков.

<good-example>
Кнопка — это `<button>` с типом:

```tsx
<button type="button" onClick={handleClose}>Cancel</button>
```
</good-example>

<bad-example>
`<div onClick>` вместо кнопки — не фокусируется с клавиатуры, не читается скринридером:

```tsx
<div onClick={handleClose}>Cancel</div>
```
</bad-example>

### 9.2. `type="button"` обязателен

Любая кнопка внутри `<form>`, которая **не** отправляет форму, должна иметь `type="button"` — иначе по умолчанию `type="submit"` и она сабмитит форму. Кнопка отправки — `type="submit"`.

```tsx
<button type="button" onClick={resetFilters}>Сбросить</button>
<Button type="submit" disabled={isPending || !isValid}>Сохранить</Button>
```

### 9.3. `aria-label` на иконочных кнопках

Кнопка без текста (только иконка) обязана иметь `aria-label`, а декоративная svg — `aria-hidden="true"`.

<good-example>
```tsx
<button type="button" aria-label={t("issue.delete")} onClick={onDelete}>
	<svg aria-hidden="true" className="h-4 w-4" /* ... */ />
</button>
```
</good-example>

<bad-example>
Иконочная кнопка без метки — скринридер читает «button», пользователь не знает назначения:

```tsx
<button type="button" onClick={onDelete}>
	<TrashIcon />
</button>
```
</bad-example>

Также `aria-label` на элементах управления без видимой подписи (например, Select выбора доски):

```tsx
<SelectTrigger aria-label="Выбор доски" className="...">
```

(В проде текст `aria-label` должен идти через i18n — см. раздел 14.)

### 9.4. Focus ring и видимый фокус

Не убирай фокус-обводку без замены. Используй `focus-visible:` утилиты. Ошибочные поля — `aria-invalid` + связанное сообщение.

```tsx
<Input
	id="title"
	aria-invalid={errors.title ? true : undefined}
	aria-describedby={errors.title ? "title-error" : undefined}
	className={errors.title ? "border-red-500 focus-visible:ring-red-500" : undefined}
/>
{errors.title ? (
	<p id="title-error" role="alert" className="text-sm text-red-500">
		{fieldError("title")}
	</p>
) : null}
```

Это реальный паттерн из `IssueForm`: `aria-invalid`, `aria-describedby`, `role="alert"` на сообщении об ошибке.

<bad-example>
Снятие обводки без замены — клавиатурный пользователь теряет позицию:

```tsx
<button className="outline-none" /> // нет focus-visible замены
```
</bad-example>

### 9.5. Клавиатурная навигация (Tab / Enter / Escape)

- Всё интерактивное достижимо `Tab` (нативные элементы — из коробки; кастомные — `tabIndex={0}` + обработчики клавиш).
- `Enter`/`Space` активируют кнопки; для кастомных — обрабатывай `onKeyDown`.
- `Escape` закрывает модалы/поповеры — у Shadcn/Radix это встроено, **не отключай**.
- Модалки: фокус-trap и возврат фокуса — уже в Radix Dialog, не ломай.

```tsx
function handleKeyDown(e: React.KeyboardEvent) {
	if (e.key === "Enter" || e.key === " ") {
		e.preventDefault();
		onActivate();
	}
	if (e.key === "Escape") onClose();
}
```

### 9.6. Контраст 4.5:1

Текст должен иметь контраст ≥ 4.5:1 к фону (для крупного текста ≥ 3:1). Токены `--hub-text-*` подобраны под это — поэтому и нельзя хардкодить цвета: легко получить нечитаемую пару. Проверяй нестандартные сочетания (текст на градиенте, на цветном бейдже).

### 9.7. Skip link и лендмарки

Есть `components/layout/skip-link.tsx` — ссылка «к основному контенту». Основные регионы размечай лендмарками: `<nav>`, `<main>`, `<aside>`, `<header>`. Один `<h1>` на страницу.

### 9.8. axe-проверки

`fe-ver` запускает axe (в т.ч. в E2E на ключевых экранах — login axe PASS уже есть). Перед сдачей сам прогоняй axe DevTools на новом экране: 0 critical/serious violations.

### 9.9. Чек-лист a11y

- [ ] Семантические элементы (`button`/`a`/заголовки/списки), не `div` с onClick.
- [ ] `type="button"` на неотправляющих кнопках; `type="submit"` на отправке.
- [ ] Иконочные кнопки имеют `aria-label`; декоративные svg — `aria-hidden`.
- [ ] Видимый focus-visible; поля ошибок — `aria-invalid` + `aria-describedby` + `role="alert"`.
- [ ] Tab/Enter/Space/Escape работают; Radix-поведение (Esc, focus-trap) не сломано.
- [ ] Контраст ≥ 4.5:1 (через токены).
- [ ] Лендмарки и один `<h1>`; skip-link на месте.
- [ ] axe: 0 critical/serious на новом экране.
- [ ] WCAG 2.2 (2026): drag-действия имеют альтернативу без drag (2.5.7); интерактивные цели ≥24×24px (2.5.8); фокус не перекрыт sticky-контентом (2.4.11).

---

## 10. Формы и валидация

Стек форм в Hub: **react-hook-form** + **@hookform/resolvers/zod** + **zod**. Схема валидации — отдельным файлом рядом с формой. (2026) Для простейших одноактных форм без схемной валидации допустима Actions-триада React 19 (`useActionState`/`useFormStatus`) — см. раздел 3.9; сложные формы — только RHF + Zod.

### 10.1. Схема Zod (отдельный файл)

Ключи сообщений — **i18n-ключи**, а не готовые тексты. Форма переводит их через `t(...)`.

```typescript
import { z } from "zod";

export const issueFormSchema = z.object({
	title: z.string().trim().min(3, "form.titleMinLength").max(200, "form.titleMaxLength"),
	issue_type: z.enum(["bug", "task", "story", "epic", "sub_task", "improvement", "tech_debt", "spike", "release"]),
	priority: z.enum(["critical", "high", "medium", "low", "trivial"]),
	assignee_id: z.string().optional(),
	estimate_minutes: z
		.string()
		.optional()
		.refine((v) => !v || (!Number.isNaN(Number(v)) && Number(v) >= 0), "form.estimateInvalid"),
	story_points: z.string().optional(),
	due_date: z.string().optional(),
});

export type IssueFormValues = z.infer<typeof issueFormSchema>;
```

**Правила:**
- Тип значений выводим из схемы: `z.infer<typeof schema>` — не дублируем руками.
- Сообщения — i18n-ключи (`"form.titleMinLength"`).
- Кастомная логика — через `.refine`.

### 10.2. Подключение RHF + zodResolver

```tsx
const {
	register,
	control,
	handleSubmit,
	formState: { errors, isValid, isSubmitting },
} = useForm<IssueFormValues>({
	resolver: zodResolver(issueFormSchema),
	mode: "onChange", // валидация по мере ввода -> можно дизейблить submit по isValid
	defaultValues: {
		title: initialData?.title ?? "",
		issue_type: initialData?.issue_type ?? "task",
		priority: initialData?.priority ?? "medium",
		// ...
	},
});
```

### 10.3. Простые поля через `register`, сложные через `Controller`

- Нативные `<input>` — `register("field")`.
- Кастомные контролы (Shadcn Select, кастомный picker) — `<Controller>`, потому что у них нет нативного `ref`/`onChange` в формате RHF.

```tsx
{/* Нативный input */}
<Input id="title" {...register("title")} aria-invalid={errors.title ? true : undefined} />

{/* Кастомный Select через Controller */}
<Controller
	name="priority"
	control={control}
	render={({ field }) => (
		<Select value={field.value} onValueChange={field.onChange}>
			<SelectTrigger id="priority"><SelectValue /></SelectTrigger>
			<SelectContent>
				{PRIORITY_KEYS.map((p) => (
					<SelectItem key={p} value={p}>{t(`priorities.${p}`)}</SelectItem>
				))}
			</SelectContent>
		</Select>
	)}
/>
```

### 10.4. Отправка + связь с мутацией

`handleSubmit` валидирует и вызывает обработчик с типизированными значениями. Дальше — мутация Query (раздел 5).

```tsx
const onSubmit = (values: IssueFormValues) => {
	if (mode === "create") {
		createMutation.mutate({ title: values.title.trim(), /* ... */ });
	} else {
		updateMutation.mutate(values);
	}
};

<form onSubmit={(e) => void handleSubmit(onSubmit)(e)} className="space-y-6">
```

Промис `handleSubmit(...)()` помечаем `void` (Biome).

### 10.5. Отображение ошибок

Показываем ошибку под полем, переводя i18n-ключ из схемы:

```tsx
const fieldError = (key: keyof IssueFormValues) => {
	const err = errors[key];
	if (!err?.message) return null;
	return t(err.message as string);
};

{errors.title ? (
	<p id="title-error" role="alert" className="text-sm text-red-500">
		{fieldError("title")}
	</p>
) : null}
```

### 10.6. Состояния disabled / loading

Кнопка отправки дизейблится, пока идёт запрос **или** форма невалидна; показывает индикатор.

```tsx
const isPending = isSubmitting || createMutation.isPending || updateMutation.isPending;

<Button type="submit" disabled={isPending || !isValid} className="w-full">
	{isPending ? t("common:btn.saving") : mode === "create" ? t("form.submitCreate") : t("form.submitEdit")}
</Button>
```

**Правило:** запрещай двойной сабмит (дизейбл на `isPending`). Не полагайся только на `isValid` — сеть тоже может падать, обрабатывай `onError`.

### 10.7. Оптимистическая блокировка (version) в формах редактирования

ТЗ требует `version` в каждой изменяемой сущности. Форма редактирования **обязана** отправлять `version` из загруженных данных — бэк вернёт 409 при конфликте:

```tsx
return updateIssue(projectId, initialData.issue_key, {
	title: values.title.trim(),
	// ...
	version: initialData.version, // optimistic locking — обязательно
});
```

При 409 (`ApiError.code === 409`) показывай пользователю, что данные устарели, и предложи перезагрузить.

### 10.8. Чек-лист форм

- [ ] Схема Zod в отдельном файле; тип из `z.infer`; сообщения — i18n-ключи.
- [ ] RHF + `zodResolver`; `mode: "onChange"` где нужен live-контроль валидности.
- [ ] Нативные поля — `register`; кастомные контролы — `Controller`.
- [ ] Ошибки под полем + `aria-invalid`/`aria-describedby`/`role="alert"`.
- [ ] Submit дизейблится на `isPending`/`!isValid`; нет двойного сабмита; есть индикатор.
- [ ] Форма редактирования шлёт `version` (optimistic locking); 409 обрабатывается.
- [ ] Успех/ошибка -> `toast` (`toastApiError` для ошибок).

---

## 11. Производительность

### 11.0. Метрики и бюджеты Core Web Vitals (2026)

Производительность меряется метриками, а не ощущениями. **INP (Interaction to Next Paint) ≤200ms** заменил FID и является самым проваливаемым CWV-метриком 2026 года; Lighthouse-score как прокси CWV — признанный антипаттерн (лабораторный скор ≠ полевые данные).

Целевые бюджеты Hub (с буфером к порогам Google):

| Метрика | Цель Hub | Порог Google «good» |
|---------|----------|---------------------|
| INP | < 150 ms | ≤ 200 ms |
| LCP | < 2.0 s | ≤ 2.5 s |
| CLS | < 0.05 | ≤ 0.1 |
| JS основного маршрута | < 300 KB gzip | — |

Практика:
- Длинные обработчики (drag на доске, фильтрация больших списков) бей на куски: long task >50ms блокирует отклик — используй `scheduler.yield()` / `startTransition` между фазами.
- `content-visibility: auto` для тяжёлых оффскрин-секций.
- Полевые метрики важнее лабораторных: RUM через библиотеку `web-vitals` (p75 INP/LCP/CLS) — план в observability-стек; алерт при p75 INP >200ms.
- Lighthouse CI сейчас artifact-only (NFR-17); перевод в blocking assertions — на обсуждении (Приложение B).

### 11.1. Lazy-загрузка тяжёлых зависимостей (обязательно)

Тяжёлые библиотеки не должны попадать в основной бандл. Грузим их через динамический `import()` только когда нужны.

**Обязательно lazy:** LiveKit (звонки), чарты (dashboard), редакторы (Plate editor), крупные модули (calls, dashboard).

```tsx
const CallsRoom = lazy(() => import("@/modules/calls/calls-room"));
const Dashboard = lazy(() => import("@/routes/dashboard"));
```

Всегда с `<Suspense fallback>` и Error Boundary (LiveKit может упасть на несовместимом браузере).

### 11.2. Виртуализация длинных списков

Списки на сотни+ элементов (сообщения, большие таблицы задач) виртуализируем (рендерим только видимое). Для чата это сочетается с infinite scroll (`useInfiniteQuery`). Используй виртуализацию (например `@tanstack/react-virtual`), когда список стабильно > ~100 строк и есть тормоза при скролле.

<bad-example>
Рендер 5000 сообщений разом — фриз UI, гигантский DOM:

```tsx
{messages.map((m) => <MessageRow key={m.id} message={m} />)} // 5000 узлов
```
</bad-example>

<good-example>
Виртуализация: в DOM только видимые строки + буфер. `key` — по `m.id`.
</good-example>

### 11.3. Debounce / throttle

Частые события (ввод в поиск, resize, скролл) throttle/debounce. Канон — `useDebouncedValue`:

```tsx
export function useDebouncedValue<T>(value: T, delayMs = 300): T {
	const [debounced, setDebounced] = useState(value);
	useEffect(() => {
		const id = setTimeout(() => setDebounced(value), delayMs);
		return () => clearTimeout(id); // критично: чистим предыдущий таймер
	}, [value, delayMs]);
	return debounced;
}
```

Применение: дебаунсим строку поиска, затем кладём в `queryKey`, чтобы не слать запрос на каждый символ:

```tsx
const debouncedSearch = useDebouncedValue(search, 300);
const { data } = useQuery({
	queryKey: ["search", debouncedSearch],
	queryFn: ({ signal }) => searchApi.search(debouncedSearch, { signal }),
	enabled: debouncedSearch.length >= 2,
});
```

### 11.4. Разбиение бандла

- Route-level split: тяжёлые маршруты — `.lazy.tsx`/`lazy()`.
- Vendor-чанки: крупные библиотеки выносятся сборкой автоматически, но не тяни тяжёлое в общий модуль без нужды.
- Не импортируй весь пакет ради одной функции (`import { thing } from "lib"`, не `import * as lib`).

### 11.5. Мемоизация (по делу)

См. раздел 3.7. Кратко: `useMemo` для дорогих вычислений (фильтрация больших массивов), `useCallback` для колбэков в deps/мемо-дочерних, `memo` для дочернего, который часто получает те же пропсы, но перерисовывается из-за родителя.

### 11.6. Избегание лишних эффектов и ререндеров

- Не клади в `useEffect` то, что можно вычислить при рендере (производные значения — просто переменные или `useMemo`).
- Узкие Zustand-селекторы (раздел 6.2) вместо подписки на весь стор.
- Не создавай новые объекты/массивы в пропсах мемоизированных дочерних без мемоизации (иначе `memo` бесполезен).

<bad-example>
Эффект для производного значения — лишний рендер и рассинхрон:

```tsx
const [fullName, setFullName] = useState("");
useEffect(() => setFullName(`${first} ${last}`), [first, last]); // не нужно
```
</bad-example>

<good-example>
```tsx
const fullName = `${first} ${last}`; // просто вычисляем при рендере
```
</good-example>

### 11.7. Skeletons вместо спиннеров

Для восприятия скорости используем skeleton-заглушки (см. `boardLoadingFallback`), а не голый спиннер на весь экран. Skeleton отражает будущий layout — меньше «прыжков».

### 11.8. Чек-лист производительности

- [ ] (2026) Бюджеты CWV соблюдены: INP <150ms, LCP <2.0s, CLS <0.05; долгие обработчики разбиты (`scheduler.yield()`/`startTransition`).
- [ ] Тяжёлые зависимости (LiveKit, charts, editor) — lazy + Suspense + Error Boundary.
- [ ] Длинные списки виртуализированы (> ~100 строк) со стабильными ключами.
- [ ] Поиск/resize/скролл — debounce/throttle с очисткой таймеров.
- [ ] Нет `import * as` ради одной функции; тяжёлое не в общем модуле.
- [ ] Мемоизация обоснована; производные значения не через эффект.
- [ ] Узкие селекторы Zustand.
- [ ] Loading через skeleton, отражающий layout.

---

## 12. Error Boundaries

### 12.1. Зачем и где обязательны

Error Boundary ловит ошибки рендера дочернего дерева и показывает fallback вместо белого экрана (FE-H4). **Обязательны** для крупных модулей: `calls`, `dashboard`, `tracker`, `tms`, `messenger` (обёртки маршрутов/модулей).

### 12.2. Реализация (единственный допустимый класс-компонент)

Error Boundary — единственное место, где в Hub используется class-компонент (React пока не даёт хук-эквивалента).

```tsx
import type { ReactNode } from "react";
import { Component } from "react";

interface Props {
	children: ReactNode;
	fallback?: ReactNode;
}
interface State {
	hasError: boolean;
}

export class ErrorBoundary extends Component<Props, State> {
	state: State = { hasError: false };

	static getDerivedStateFromError(): State {
		return { hasError: true };
	}

	componentDidCatch(error: Error): void {
		console.error("UI error boundary:", error);
	}

	render() {
		if (this.state.hasError) {
			return this.props.fallback ?? <DefaultErrorFallback />;
		}
		return this.props.children;
	}
}
```

### 12.3. Применение

Оборачивай модуль/маршрут, давая осмысленный fallback:

```tsx
<ErrorBoundary fallback={<ModuleError module="tracker" />}>
	<TrackerModule />
</ErrorBoundary>
```

Для lazy-модулей комбинируй с Suspense:

```tsx
<ErrorBoundary fallback={<CallsError />}>
	<Suspense fallback={<CallsSkeleton />}>
		<CallsRoom />
	</Suspense>
</ErrorBoundary>
```

### 12.4. Хороший fallback

- Понятное сообщение (через i18n в проде), не стек-трейс пользователю.
- Действие восстановления: «Обновить» (`window.location.reload()`) или сброс состояния.
- Стилизация по дизайн-токенам.

```tsx
<div className="flex flex-col items-center justify-center gap-3 p-8 text-center">
	<p className="text-[15px] font-semibold text-t1">{t("errors.title")}</p>
	<p className="text-[13px] text-t3 max-w-md">{t("errors.description")}</p>
	<button type="button" onClick={() => window.location.reload()} className="rounded-[var(--r2)] border border-b1 px-4 py-2 text-[13px] text-t2 hover:text-t1">
		{t("errors.reload")}
	</button>
</div>
```

### 12.5. Что Error Boundary НЕ ловит

- Ошибки в обработчиках событий (`onClick`) — оборачивай в try/catch или обрабатывай через Query `onError`.
- Асинхронные ошибки (fetch) — это зона TanStack Query (`isError`, `onError`), не boundary.
- Ошибки в SSR/во время самого рендера boundary.

Поэтому Error Boundary — страховка от неожиданных ошибок рендера, а не замена нормальной обработке ошибок запросов.

### 12.6. Чек-лист Error Boundaries

- [ ] Крупные модули (calls/dashboard/tracker/tms/messenger) обёрнуты в Error Boundary.
- [ ] Lazy-модули: Error Boundary + Suspense.
- [ ] Fallback осмысленный, с действием восстановления, по токенам, без стек-трейса.
- [ ] Ошибки запросов обрабатываются через Query, а не «ожидаются» boundary.

---

## 13. Реальное время (WebSocket)

### 13.1. Архитектура WS в Hub

Единое WS-соединение на всё приложение живёт в `ws-store` (Zustand, без persist). Соединение поднимается в `_authenticated.tsx` через `useWebSocket()` при аутентификации. Слушатели событий регистрируются в наборе (`Set<EventHandler>`), фичи подписываются на комнаты.

ТЗ: WebSocket через Fiber (fasthttp) на бэке; pub/sub — NATS JetStream. На фронте — нативный `WebSocket`.

### 13.2. WS-стор: соединение и рассылка событий

```typescript
export const useWsStore = create<WsState>((set, get) => ({
	connected: false,
	socket: null,
	listeners: new Set(),

	connect: (token: string) => {
		const { socket } = get();
		if (socket) return; // одно соединение

		const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
		const ws = new WebSocket(`${protocol}//${window.location.host}/ws?token=${encodeURIComponent(token)}`);

		ws.onopen = () => set({ connected: true, socket: ws });
		ws.onclose = () => set({ connected: false, socket: null });
		ws.onerror = () => ws.close();
		ws.onmessage = (event) => {
			try {
				const msg = JSON.parse(event.data as string) as ServerMessage;
				if (msg.type === "event" && msg.event) {
					for (const handler of get().listeners) handler(msg.event, msg.payload);
				}
			} catch {
				// битый JSON от WS — пропускаем, не роняем обработчики
			}
		};

		set({ socket: ws });
	},

	subscribe: (room) => {
		const { socket } = get();
		if (socket?.readyState === WebSocket.OPEN) {
			socket.send(JSON.stringify({ type: "subscribe", room }));
		}
	},
	// unsubscribe, addListener, removeListener аналогично
}));
```

### 13.3. Подписка на комнаты из хука фичи

Комната — строка вида `channel:{id}`, `project:{id}`, `issue:{key}`. Подписка — в `useEffect` с очисткой (отписка при размонтировании/смене id).

```tsx
useEffect(() => {
	if (!connected || !channelId) return;
	const room = `channel:${channelId}`;
	subscribe(room);
	return () => {
		unsubscribe(room);
	};
}, [connected, channelId, subscribe, unsubscribe]);
```

Для набора комнат есть `useRooms(rooms: string[])` — сам вычисляет дельту (на что подписаться, от чего отписаться).

**Безопасность (ТЗ NF-02):** подписка на комнату без проверки membership/RBAC на бэке запрещена. Фронт шлёт `subscribe`, но авторизацию комнаты проверяет сервер — не считай, что клиентская подписка = доступ.

### 13.4. Обновление кэша Query из WS-событий

Главный паттерн интеграции WS + Query: событие правит кэш через `setQueryData`, **без** рефетча. Иммутабельно, с защитой от дублей по `id`.

```tsx
const handleEvent = useCallback(
	(event: string, payload: unknown) => {
		if (!channelId) return;
		const data = payload as Record<string, unknown>;
		if (data.channel_id && data.channel_id !== channelId) return; // чужой канал — игнор

		const queryKey = ["messenger", "messages", channelId];

		if (event === "message_sent") {
			const message = buildMessageFromWsEvent(data, channelId);
			if (!message) return;
			queryClient.setQueryData(queryKey, (old: unknown) => {
				if (!old) return old;
				const typed = old as { pages: Message[][]; pageParams: unknown[] };
				// WS может прислать дубликат -> не добавляем повторно по id
				const exists = typed.pages.some((page) => page.some((m) => m.id === message.id));
				if (exists) return old;
				const lastPage = typed.pages[typed.pages.length - 1] ?? [];
				return { ...typed, pages: [...typed.pages.slice(0, -1), [...lastPage, message]] };
			});
		}
	},
	[channelId, queryClient],
);

useWsEventListener(handleEvent);
```

**Правила обновления из WS:**
- Иммутабельность: новые массивы/объекты, не мутация кэша.
- Дедупликация по `id` (WS и оптимистичное создание могут дать дубль).
- Фильтрация чужих событий (по `channel_id`/`room`).
- Для infinite-структуры уважай форму `{ pages, pageParams }`.
- Обработчик мемоизирован (`useCallback`) — иначе `useWsEventListener` будет переподписываться каждый рендер.

### 13.5. `useWsEventListener` — безопасная подписка на события

Хук управляет регистрацией/снятием обработчика и переустанавливает его при смене ссылки:

```tsx
export function useWsEventListener(handler: EventHandler) {
	const addListener = useWsStore((s) => s.addListener);
	const removeListener = useWsStore((s) => s.removeListener);
	const handlerRef = useRef<EventHandler | null>(null);

	useEffect(() => {
		if (handlerRef.current) removeListener(handlerRef.current);
		handlerRef.current = handler;
		addListener(handler);
		return () => removeListener(handler);
	}, [handler, addListener, removeListener]);
}
```

### 13.6. Reconnect

Текущий `ws-store` при `onclose` сбрасывает `connected/socket`. Поскольку `useWebSocket` завязан на `isAuthenticated`/`accessToken`, повторное подключение инициируется при изменении токена/состояния. Для устойчивого reconnect с backoff (сеть моргнула) паттерн:

```typescript
// В connect: при onclose, если пользователь всё ещё аутентифицирован — переподключение с экспоненциальной задержкой
let reconnectAttempts = 0;
ws.onclose = () => {
	set({ connected: false, socket: null });
	const { isAuthenticated, accessToken } = useAuthStore.getState();
	if (isAuthenticated && accessToken) {
		const delay = Math.min(1000 * 2 ** reconnectAttempts, 30_000);
		reconnectAttempts += 1;
		setTimeout(() => get().connect(accessToken), delay);
	}
};
// при onopen: reconnectAttempts = 0
```

**Правило:** не делай агрессивный reconnect без backoff (заддосишь свой же WS-гейт). После reconnect нужно **переподписаться** на активные комнаты — это делают `useRooms`/эффекты подписки, потому что они завязаны на `connected`.

### 13.7. Чек-лист реального времени

- [ ] Одно WS-соединение через `ws-store`; поднимается в `_authenticated`.
- [ ] Подписка на комнаты в `useEffect` с отпиской в cleanup (или через `useRooms`).
- [ ] WS-события правят кэш через `setQueryData` иммутабельно, с дедупом по `id` и фильтром чужих.
- [ ] Обработчик событий мемоизирован (`useCallback`).
- [ ] Битый JSON от WS не роняет обработчики (try/catch).
- [ ] Reconnect с backoff; после reconnect — переподписка на комнаты.
- [ ] Не полагаемся на клиентскую подписку как на авторизацию (RBAC — на бэке).

---

## 14. Интернационализация (i18n)

Стек: **i18next** + **react-i18next** + **i18next-browser-languagedetector**. Языки: `ru` (fallback) и `en`. Namespaces: `common`, `tracker`, `messenger`, `tms`, `calls`, `dashboard`, `settings`, `help`, `onboarding`.

### 14.1. Никакого хардкода user-visible строк

Любой текст, который видит пользователь, идёт через `t(...)`. Хардкод-строки в новом коде — замечание `fe-ver`.

<good-example>
```tsx
const { t } = useTranslation(["tracker", "common"]);
<Label htmlFor="title">{t("form.title")}</Label>
<Button type="submit">{t("form.submitCreate")}</Button>
```
</good-example>

<bad-example>
```tsx
<Label htmlFor="title">Название задачи</Label> // хардкод -> непереводимо
```
</bad-example>

> Примечание: в исторических участках Hub встречаются русские строки прямо в JSX (например фильтры доски). Это техдолг. **Новый** код пишем через i18n; при существенной правке legacy-экрана — выносим строки в ключи.

### 14.2. Ключи и namespaces

- Ключ — по смыслу и месту: `form.title`, `priorities.high`, `kanban.moveError`.
- Namespace выбираем по домену: `useTranslation("tracker")` или список `["tracker", "common"]`.
- Кросс-namespace ключ — с префиксом: `t("common:btn.saving")`.

```tsx
const { t } = useTranslation(["tracker", "common"]);
t("form.createSuccess");      // из tracker
t("common:btn.saving");       // из common
```

Файлы переводов: `src/i18n/locales/ru/<ns>.json` и `.../en/<ns>.json`. Новый ключ добавляется **в оба** языка (ru и en) синхронно — иначе один язык покажет ключ вместо текста.

### 14.3. Интерполяция и плюрализация

Интерполяция:

```json
{ "issues.count": "Задач: {{count}}" }
```

```tsx
t("issues.count", { count: total });
```

Плюрализация (i18next поддерживает формы по `count`):

```json
{
	"comments_zero": "Нет комментариев",
	"comments_one": "{{count}} комментарий",
	"comments_few": "{{count}} комментария",
	"comments_many": "{{count}} комментариев"
}
```

```tsx
t("comments", { count: comments.length });
```

**Правило:** не собирай множественное число конкатенацией (`count + " задач"`) — используй плюрал-формы i18next.

### 14.4. Форматирование дат в TZ пользователя

ТЗ (жёстко): **время в БД всегда UTC**; фронтенд конвертирует в локальный timezone пользователя. Бэк отдаёт ISO/RFC3339 в UTC — на фронте форматируем в локаль.

```tsx
// Дата из API (UTC ISO) -> локальный формат пользователя
const formatted = new Intl.DateTimeFormat(i18n.language, {
	dateStyle: "medium",
	timeStyle: "short",
}).format(new Date(issue.created_at));
```

**Правила дат:**
- Никогда не показывай сырой UTC-ISO пользователю.
- Не форматируй дату руками через конкатенацию — используй `Intl.DateTimeFormat`/общий хелпер (`lib/`), учитывающий локаль (`i18n.language`).
- Относительное время («5 минут назад») — через локале-зависимый форматтер (`Intl.RelativeTimeFormat`).
- Отправляя дату на бэк — отправляй в UTC/ISO (`toISOString()`), не в локальном виде.

### 14.5. Переключение языка

Есть `components/language-switcher.tsx`; выбор кэшируется в localStorage (`i18nextLng`) детектором. При смене языка перерисовываются переведённые компоненты — не кэшируй перевод в `useState`.

### 14.6. Чек-лист i18n

- [ ] Нет хардкода user-visible строк; всё через `t(...)`.
- [ ] Ключ добавлен в **оба** языка (ru + en), нужный namespace.
- [ ] Интерполяция через `{{var}}`; множественное — через плюрал-формы, не конкатенацию.
- [ ] Даты: UTC из бэка -> `Intl.DateTimeFormat` в локали пользователя; на бэк — ISO/UTC.
- [ ] Строки в `aria-label`/placeholder/тостах — тоже через i18n.

---

## 15. Типобезопасность

### 15.1. Строгий TypeScript

Проект на `strict: true`. Цель — «если компилируется, значит форма данных верна». Регрессия начинается с `npx tsc --noEmit` (раздел 18).

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
| Unit/компонентные | `testframework/stages/unit/frontend-test/domains/<domain>/` | Vitest + React Testing Library |
| API-моки | там же | MSW (Mock Service Worker) |
| E2E | `testframework/stages/e2e/domains/<domain>/` | Playwright (POM) |
| E2E shared | `testframework/stages/e2e/shared/` | POM, steps, fixtures |

Тесты **не** живут в `frontend/src/`. Домены фронта: `auth`, `admin`, `ai`, `v4` и др.

### 16.2. Перед запуском — линк node_modules

Обязательный шаг перед Vitest/E2E (иначе не найдутся зависимости):

```bash
node frontend/scripts/link-testframework-node-modules.mjs
```

### 16.3. Команды

```bash
# Регрессия фронта (из frontend/)
cd frontend
node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit
npm run lint          # Biome
npm run test:run      # Vitest прогон
```

### 16.4. Что и как тестировать (минимум)

- **Компоненты**: RTL — рендер, взаимодействие (`userEvent`), проверка доступного текста/ролей (`getByRole`), а не деталей реализации.
- **Хуки/Query**: оборачивай в `QueryClientProvider`; API мокай через MSW, не патчь `fetch` руками.
- **Zustand**: сбрасывай стор между тестами; тестируй действия и селекторы.
- **A11y-регресс**: ключевые экраны — через axe в E2E.

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

Линтинг и форматирование — Biome (`npm run lint`). Он ловит `key={index}`, non-null без причины, неиспользуемое, порядок импортов, правила хуков. `biome-ignore` — только с причиной.

### 16.6. Чек-лист тестирования

- [ ] Тесты созданы в `testframework/...`, не в `src/`.
- [ ] Перед прогоном выполнен линк node_modules.
- [ ] API мокается через MSW; Query обёрнут в провайдер.
- [ ] Тестируем поведение/доступные роли, не детали реализации.
- [ ] `npm run test:run` и `npm run lint` зелёные.
- [ ] Новый flow покрыт E2E (POM) при необходимости.

---

## 17. Комментарии, логирование, git

> Детали — в `.cursor/playbooks/comments.md`, `.cursor/playbooks/logging.md`, `.cursor/playbooks/git.md`. Кратко для фронта:

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

После любого изменения фронта, **до передачи `fe-ver`/коммита**, прогнать:

```bash
cd frontend
node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit
npm run lint
npx vitest run
```

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

- [ ] `tsc --noEmit` зелёный.
- [ ] `npm run lint` (Biome) зелёный.
- [ ] `npx vitest run` зелёный (после линка node_modules).
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
- [ ] Тесты добавлены/обновлены в `testframework/`; регрессия зелёная.
- [ ] Docs-sync выполнен (STATE + SHORTCUTS/QUICKSTART при новом flow/hotkey).
- [ ] Атомарные коммиты, Conventional Commits, без секретов/артефактов.

### 19.2. Что проверяет fe-ver

`fe-ver` (skill `.cursor/project_skills/FE_VER_skill.md`) верифицирует **независимо**:

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

## Приложение A. Сквозной пример новой фичи

Чтобы связать все разделы, разберём реализацию фичи «список меток проекта с добавлением» от API до маршрута. Это эталон порядка работы и слоёв для fe1/fe2.

### A.1. Шаг 1 — API-слой (`api/tracker.ts`)

Сначала фиксируем контракт: DTO и функции. Они уже есть в `tracker.ts`, используем как есть:

```typescript
export interface LabelResponse {
	id: string;
	name: string;
	color: string;
	group_name?: string;
}

export interface CreateLabelRequest {
	name: string;
	color: string;
	group_id?: string;
}

export function listLabels(projectId: string) {
	return api.get<LabelResponse[]>(`/projects/${projectId}/labels`);
}

export function createLabel(projectId: string, data: CreateLabelRequest) {
	return api.post<LabelResponse>(`/projects/${projectId}/labels`, data);
}

export function deleteLabel(projectId: string, labelId: string) {
	return api.delete<void>(`/projects/${projectId}/labels/${labelId}`);
}
```

**Что важно:** типы — источник истины для UI; функции — тонкие обёртки над `api`; ответы типизированы дженериком.

### A.2. Шаг 2 — Query-хук (`components/tracker/use-labels.ts` или `modules/tracker/hooks/`)

Оборачиваем API в Query-хуки. Server-state целиком в Query, ключи по конвенции.

```typescript
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	type CreateLabelRequest,
	createLabel,
	deleteLabel,
	listLabels,
} from "@/api/tracker";

/** Список меток проекта (редко меняется -> длинный staleTime). */
export function useLabels(projectId: string) {
	return useQuery({
		queryKey: ["labels", projectId],
		queryFn: () => listLabels(projectId),
		staleTime: 5 * 60_000,
		enabled: !!projectId,
	});
}

/** Создание метки с инвалидацией списка. */
export function useCreateLabel(projectId: string) {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: (data: CreateLabelRequest) => createLabel(projectId, data),
		onSuccess: () => {
			void qc.invalidateQueries({ queryKey: ["labels", projectId] });
		},
	});
}

/** Удаление метки с оптимистичным обновлением и откатом. */
export function useDeleteLabel(projectId: string) {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: (labelId: string) => deleteLabel(projectId, labelId),
		onMutate: async (labelId) => {
			await qc.cancelQueries({ queryKey: ["labels", projectId] });
			const previous = qc.getQueryData<import("@/api/tracker").LabelResponse[]>([
				"labels",
				projectId,
			]);
			qc.setQueryData<import("@/api/tracker").LabelResponse[]>(
				["labels", projectId],
				(old) => old?.filter((l) => l.id !== labelId) ?? old,
			);
			return { previous };
		},
		onError: (_err, _labelId, context) => {
			if (context?.previous) {
				qc.setQueryData(["labels", projectId], context.previous);
			}
		},
		onSettled: () => {
			void qc.invalidateQueries({ queryKey: ["labels", projectId] });
		},
	});
}
```

**Что важно:** `useLabels` — чтение; `useCreateLabel` — простая мутация с инвалидацией; `useDeleteLabel` — полный optimistic-цикл (cancel/snapshot/rollback/settled). UI не знает про `fetch` — только про хуки.

### A.3. Шаг 3 — Валидация формы (`components/tracker/label-form-schema.ts`)

```typescript
import { z } from "zod";

/** Схема формы метки: имя 1-40, цвет — hex. Сообщения — i18n-ключи. */
export const labelFormSchema = z.object({
	name: z.string().trim().min(1, "labels.nameRequired").max(40, "labels.nameMax"),
	color: z
		.string()
		.regex(/^#[0-9a-fA-F]{6}$/, "labels.colorInvalid"),
});

export type LabelFormValues = z.infer<typeof labelFormSchema>;
```

### A.4. Шаг 4 — Компонент формы (`components/tracker/label-form.tsx`)

```tsx
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { useCreateLabel } from "@/components/tracker/use-labels";
import {
	type LabelFormValues,
	labelFormSchema,
} from "@/components/tracker/label-form-schema";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toastApiError } from "@/lib/format-api-error";

/** Форма добавления метки проекта (RHF + Zod). */
export function LabelForm({ projectId }: { projectId: string }) {
	const { t } = useTranslation(["tracker", "common"]);
	const createMutation = useCreateLabel(projectId);

	const {
		register,
		handleSubmit,
		reset,
		formState: { errors, isValid, isSubmitting },
	} = useForm<LabelFormValues>({
		resolver: zodResolver(labelFormSchema),
		mode: "onChange",
		defaultValues: { name: "", color: "#5B9CF5" },
	});

	const onSubmit = (values: LabelFormValues) => {
		createMutation.mutate(values, {
			onSuccess: () => {
				toast.success(t("labels.created"));
				reset();
			},
			onError: (err) => toastApiError(err, t("labels.createError")),
		});
	};

	const isPending = isSubmitting || createMutation.isPending;

	return (
		<form onSubmit={(e) => void handleSubmit(onSubmit)(e)} className="space-y-4">
			<div className="space-y-2">
				<Label htmlFor="label-name">{t("labels.name")}</Label>
				<Input
					id="label-name"
					{...register("name")}
					aria-invalid={errors.name ? true : undefined}
					aria-describedby={errors.name ? "label-name-error" : undefined}
				/>
				{errors.name ? (
					<p id="label-name-error" role="alert" className="text-sm text-red-500">
						{t(errors.name.message as string)}
					</p>
				) : null}
			</div>

			<div className="space-y-2">
				<Label htmlFor="label-color">{t("labels.color")}</Label>
				<Input id="label-color" type="color" {...register("color")} className="h-9 w-16 p-1" />
				{errors.color ? (
					<p role="alert" className="text-sm text-red-500">
						{t(errors.color.message as string)}
					</p>
				) : null}
			</div>

			<Button type="submit" disabled={isPending || !isValid}>
				{isPending ? t("common:btn.saving") : t("labels.add")}
			</Button>
		</form>
	);
}
```

### A.5. Шаг 5 — Список с удалением (`components/tracker/label-list.tsx`)

```tsx
import { useTranslation } from "react-i18next";
import { useDeleteLabel, useLabels } from "@/components/tracker/use-labels";

/** Список меток проекта с удалением. */
export function LabelList({ projectId }: { projectId: string }) {
	const { t } = useTranslation("tracker");
	const { data: labels, isLoading, isError } = useLabels(projectId);
	const deleteMutation = useDeleteLabel(projectId);

	// Хуки выше — далее условный рендер
	if (isLoading) return <LabelListSkeleton />;
	if (isError) return <p className="text-sm text-t3">{t("labels.loadError")}</p>;
	if (!labels || labels.length === 0) {
		return <p className="text-sm text-t3">{t("labels.empty")}</p>;
	}

	return (
		<ul className="space-y-2">
			{labels.map((label) => (
				<li
					key={label.id}
					className="flex items-center justify-between rounded-[var(--r2)] border border-b1 bg-card px-3 py-2"
				>
					<span className="flex items-center gap-2 text-[13px] text-t1">
						<span
							aria-hidden="true"
							className="h-3 w-3 rounded-full"
							style={{ backgroundColor: label.color }}
						/>
						{label.name}
					</span>
					<button
						type="button"
						aria-label={t("labels.deleteAria", { name: label.name })}
						onClick={() => deleteMutation.mutate(label.id)}
						disabled={deleteMutation.isPending}
						className="text-t3 hover:text-red-500"
					>
						<svg aria-hidden="true" className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
							<path strokeLinecap="round" strokeLinejoin="round" d="M6 6l12 12M6 18L18 6" />
						</svg>
					</button>
				</li>
			))}
		</ul>
	);
}

function LabelListSkeleton() {
	return (
		<ul className="space-y-2">
			{[1, 2, 3].map((i) => (
				<li key={`label-sk-${i}`} className="h-10 rounded-[var(--r2)] animate-shimmer" />
			))}
		</ul>
	);
}
```

**Разбор по playbook:**
- Все хуки (`useLabels`, `useDeleteLabel`) — до условного рендера (раздел 3.2).
- `key={label.id}` — стабильный (3.3); скелетон — `label-sk-${i}`, не голый индекс.
- Цветовой индикатор — единственный оправданный `style` (рантайм-цвет из данных, не из дизайн-токена; раздел 8.4), при этом `aria-hidden`.
- Иконочная кнопка удаления имеет `aria-label` через i18n с интерполяцией (разделы 9.3, 14.3).
- Кнопка дизейблится на `isPending` (10.6).
- Строки через `t(...)` (14.1).

### A.6. Шаг 6 — Встраивание в маршрут

```tsx
import { createFileRoute } from "@tanstack/react-router";
import { ErrorBoundary } from "@/components/ui/error-boundary";
import { LabelForm } from "@/components/tracker/label-form";
import { LabelList } from "@/components/tracker/label-list";

function ProjectLabelsPage() {
	const { projectId } = Route.useParams();
	return (
		<ErrorBoundary>
			<section className="mx-auto max-w-lg space-y-6 p-6">
				<LabelForm projectId={projectId} />
				<LabelList projectId={projectId} />
			</section>
		</ErrorBoundary>
	);
}

export const Route = createFileRoute("/_authenticated/projects/$projectId/settings/labels/")({
	component: ProjectLabelsPage,
});
```

### A.7. Шаг 7 — Регрессия, тесты, docs-sync

```bash
cd frontend
node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit && npm run lint && npx vitest run
```

- Тест в `testframework/stages/unit/frontend-test/domains/tracker/` — рендер формы, ввод имени, сабмит (MSW мокает POST), проверка тоста; список — рендер, удаление (MSW), проверка оптимистичного исчезновения и отката при 500.
- Новый экран настроек -> при необходимости обновить `QUICKSTART.md`; `.planning/STATE.md` — session note.
- Атомарные коммиты: `feat(tracker): add project labels api hooks`, затем `feat(tracker): add label form and list ui`, затем `feat(tracker): add project labels settings route`.

Эта последовательность (API -> хуки -> схема -> компоненты -> маршрут -> регрессия/docs) — рекомендуемый порядок для любой фронтенд-фичи в Hub.

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
| Тесты в `frontend/src/` | Нарушение pipeline-структуры | `testframework/stages/unit/frontend-test/...` |
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
- [ ] `npx tsc --noEmit` — зелёный.
- [ ] `npm run lint` (Biome) — зелёный.
- [ ] `npx vitest run` — зелёный (после `link-testframework-node-modules.mjs`).
- [ ] Тесты добавлены/обновлены в `testframework/`.
- [ ] `.planning/STATE.md` + `SHORTCUTS.md`/`QUICKSTART.md` (при новом flow/hotkey) обновлены.
- [ ] Атомарные Conventional Commits; без секретов/артефактов.

> Готово к `fe-ver`, когда все чекбоксы отмечены и приложен зелёный лог регрессии.

---

## Приложение B. Дельты 2026 на обсуждение

> Источник: `.planning/research/BEST-PRACTICES-2026-frontend-testing.md` (2026-07-03). Ниже — дельты medium/low со статусом «обсудить»: в playbook не внесены, решение — за team-lead/arch. Не применять без решения.

| # | Тема | Суть | Предлагаемое действие |
|---|------|------|----------------------|
| 9 | Simplified optimistic (Query v5) | Читать `mutation.variables` при `isPending` и рисовать оптимистичный элемент без записи в кэш (rollback не нужен) | Дополнить §5.5 выбором: variables-паттерн (одно место) vs кэш-паттерн (несколько мест) |
| 10 | Offline / persist для PWA | `persistQueryClient` + `PersistQueryClientProvider` (`gcTime ≥ maxAge`), `networkMode: 'offlineFirst'` | Оценить для PWA: минимум — задокументировать осознанный отказ; максимум — persist для справочников (channels, projects) |
| 11 | Строгие TS-флаги | `verbatimModuleSyntax`, `noUncheckedSideEffectImports`, `erasableSyntaxOnly` (дешёвые), `exactOptionalPropertyTypes` (шумный, последним) | Включать поэтапно в `frontend/tsconfig.json` с прогоном `tsc --noEmit` |
| 12 | TypeScript 5.9 | 5.8: `erasableSyntaxOnly`; 5.9: `strictInference` входит в `strict` | Плановый chore-апгрейд с прогоном; ожидать новые ошибки от strictInference |
| 16 | Perf-бюджеты в CI | Lighthouse CI assertions (blocking) + size-limit на PR (JS <300KB gzip, diff чанка) | Перевести lighthouse job из artifact-only в blocking; закрывает NFR-17. Учесть бюджет GitHub Actions |
| 17 | Vite 7 (rolldown) | Rust-бандлер, быстрее билд; миграция с v6 малоболезненная | Chore-апгрейд после стабилизации экосистемы плагинов (pwa, router plugin) |
| 20 | Biome-строгость | a11y-правила и `noExplicitAny` — `warn` → `error` (иначе gate декоративный); `noFloatingPromises`; GritQL-плагин для Hub-запретов | Поднять уровни в `frontend/biome.json`; плагин — nice-to-have |
| 21 | `ref` как проп (React 19) | `forwardRef` не нужен для новых компонентов | Правило для нового кода: без `forwardRef`; существующие Shadcn-обёртки не трогать |

---

*Playbook фронтенда Hub. Источники истины: `.cursor/rules/30-frontend.mdc`, `.cursor/rules/10-style.mdc`, `.cursor/rules/60-tz-constraints.mdc`, `docs/code_style_guide.md`. При расхождении — правила и ТЗ приоритетны.*
