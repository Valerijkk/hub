# Frontend: Zustand и живые данные

> Часть плейбука. Канон и оглавление — [`../frontend.md`](../frontend.md).

Клиентское состояние и обновления по WebSocket.

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
