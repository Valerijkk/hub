# Frontend: TanStack Router и Query

> Часть плейбука. Канон и оглавление — [`../frontend.md`](../frontend.md).

Маршрутизация, серверное состояние, кэш и инвалидация.

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
