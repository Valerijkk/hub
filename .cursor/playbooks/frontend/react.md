# Frontend: React 19 и Error Boundaries

> Часть плейбука. Канон и оглавление — [`../frontend.md`](../frontend.md).

Компоненты, состояние, границы ошибок.

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
