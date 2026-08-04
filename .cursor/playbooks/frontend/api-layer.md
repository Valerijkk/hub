# Frontend: API-слой и бизнес-логика

> Часть плейбука. Канон и оглавление — [`../frontend.md`](../frontend.md).

Клиенты, контракты, где живёт логика домена.

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
