# Frontend: unit на Vitest и RTL

> Часть плейбука тестирования. Канон и оглавление — [`../testing.md`](../testing.md).
> Все Node/Vitest/tsc/Biome команды ниже выполняются только внутри task-owned
> Docker container с CPU/RAM limits и read-only source mount.
> Legacy `testframework/stages/unit/frontend-test/**` frozen. Новые Vitest/RTL
> tests — только genuine units в `frontend/tests/<domain>/**`.

Канон Hub для фронтовых юнит-тестов и рендер-хелперов.

---

## 5. Frontend unit (Vitest + RTL)

Frontend-тесты (99 тестов, Vitest + React Testing Library) живут в `testframework/stages/unit/frontend-test/domains/{domain}/`, job `frontend-test`. Перед прогоном линкуются node_modules:

```bash
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run
# один домен:
npx vitest run ../testframework/stages/unit/frontend-test/domains/auth
```

### 5.1. Рендер через TestRenderer

Компоненты, которым нужны провайдеры (TanStack Query и т.п.), рендерятся через `renderWithProviders` из `render.tsx`. Он даёт каждому тесту свой `QueryClient` с отключёнными retry:

```typescript
import { renderWithProviders } from "../../render";
import { screen } from "@testing-library/react";
import { IssueList } from "@/components/tracker/issue-list";

describe("IssueList", () => {
	it("renders issues from query", async () => {
		renderWithProviders(<IssueList projectId="proj-1" />);
		expect(await screen.findByText("HUB-42")).toBeInTheDocument();
	});
});
```

Для простых компонентов без провайдеров допустим прямой `render` из RTL (как в `issue-card.test.tsx`).

> **(2026) Vitest 4 Browser Mode.** В Vitest 4 (дек 2025) Browser Mode стабилен: компонентные тесты можно гонять в реальном Chromium через provider `playwright()` — jsdom-моки (`matchMedia`, `ResizeObserver`) не нужны. Массово не мигрируем; применяем точечно для компонентов, страдающих от jsdom (DnD, виртуализация, Radix-поповеры), после планового апгрейда Vitest 3→4. Остальное остаётся в jsdom.

**Почему свой QueryClient на тест:** отключённые `retry: false` делают ошибки детерминированными; изоляция клиента исключает утечку кэша между тестами (иначе — флейки от порядка).

### 5.2. Queries: by role / by label — приоритет

RTL-запросы отражают то, как пользователь и скринридер видят UI. Приоритет (сверху вниз):

| Приоритет | Query | Когда |
|-----------|-------|-------|
| 1 | `getByRole` | кнопки, ссылки, заголовки, поля (`getByRole('button', { name: /edit/i })`) |
| 2 | `getByLabelText` | поля формы с label |
| 3 | `getByText` | некликабельный текст, бейджи |
| 4 | `getByTestId` | крайний случай, когда роль/текст недоступны |

**Good — by role/label (реальный `issue-card.test.tsx`):**

```typescript
it("renders issue key and title", () => {
	render(<IssueCard issue={baseIssue} projectId="proj-1" />);
	expect(screen.getByText("HUB-42")).toBeInTheDocument();
	expect(screen.getByText("Fix login bug")).toBeInTheDocument();
});
```

**Bad — хрупкие CSS-селекторы:**

```typescript
// container.querySelector завязан на структуру DOM, ломается при рефакторинге верстки
const { container } = render(<IssueCard issue={baseIssue} projectId="proj-1" />);
expect(container.querySelector(".issue-card__title span")).toHaveTextContent("Fix login bug");
```

### 5.3. Взаимодействие через userEvent

`userEvent` эмулирует реальные события (focus, keydown, click) точнее, чем `fireEvent`. Всегда `await`:

```typescript
import userEvent from "@testing-library/user-event";

it("calls onEdit when edit button clicked", async () => {
	const onEdit = vi.fn();
	render(<IssueCard issue={baseIssue} projectId="proj-1" onEdit={onEdit} />);
	await userEvent.click(screen.getByRole("button", { name: /edit/i }));
	expect(onEdit).toHaveBeenCalledWith(baseIssue.id);
});
```

**Bad:** синхронный `fireEvent.click` без await для сложных взаимодействий (ввод текста, комбо клавиш) — пропускает промежуточные события и даёт ложные результаты.

### 5.4. Мок API: MSW vs vi.spyOn

Два подхода к изоляции сети:

**vi.spyOn на api-клиент** — когда тестируем логику, вызывающую клиент (реальный `admin-api.test.ts`):

```typescript
import { api } from "@/api/client";
import { banUser } from "@/api/admin";

it("banUser posts reason to /admin/users/:id/ban", async () => {
	const postSpy = vi.spyOn(api, "post").mockResolvedValue({ success: true, message: "User banned" });
	const result = await banUser("user-42", "spam");
	expect(postSpy).toHaveBeenCalledWith("/admin/users/user-42/ban", { reason: "spam" });
	expect(result.success).toBe(true);
});
```

**MSW (Mock Service Worker)** — когда компонент делает fetch на уровне HTTP и мы хотим проверить весь путь query → render:

```typescript
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";

const server = setupServer(
	http.get("/api/projects/:id/issues", ({ params }) => {
		return HttpResponse.json({ data: mockIssues, total: mockIssues.length });
	}),
);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers()); // сброс переопределений между тестами
afterAll(() => server.close());
```

`onUnhandledRequest: "error"` ловит незамоканные запросы — защита от «случайно ходили в реальный бэкенд».

**Выбор:** тестируем чистую функцию/api-обёртку → `vi.spyOn`; тестируем компонент с data-fetching через TanStack Query → MSW.

### 5.5. Тестирование хуков

Хуки тестируются через `renderHook` + `act`. Реальный `use-auth.test.ts`:

```typescript
import { act, renderHook } from "@testing-library/react";
import * as authApi from "@/api/auth";
import { useAuth } from "@/hooks/use-auth";
import { useAuthStore } from "@/stores/auth-store";

it("logout calls POST /auth/logout before clearing store", async () => {
	const logoutSpy = vi.spyOn(authApi, "logout").mockResolvedValue(undefined);
	useAuthStore.getState().setTokens("access-token", "refresh-token");

	const { result } = renderHook(() => useAuth());
	await act(async () => {
		await result.current.logout();
	});

	expect(logoutSpy).toHaveBeenCalledOnce();
	expect(useAuthStore.getState().isAuthenticated).toBe(false);
});
```

`act` оборачивает вызовы, меняющие состояние, чтобы React применил обновления до ассертов.

### 5.6. Тестирование стора (Zustand)

Стор — синглтон, поэтому сбрасываем его в `beforeEach` (реальный паттерн из `use-auth.test.ts`):

```typescript
function resetAuthStore(): void {
	sessionStorage.clear();
	useAuthStore.setState({
		accessToken: null,
		refreshToken: null,
		user: null,
		isAuthenticated: false,
	});
}

describe("useAuth", () => {
	beforeEach(() => {
		resetAuthStore();
		vi.restoreAllMocks(); // снять все spy/mock между тестами
	});
	// ...
});
```

**Правило:** любой глобальный синглтон (стор, sessionStorage) сбрасывается в `beforeEach`; иначе тесты зависят от порядка → флейки.

### 5.7. Snapshot — осторожно

Snapshot-тесты допустимы **только** для стабильного, редко меняющегося вывода (например, форматтер ошибки). Для компонентов с активной вёрсткой они дают шум: любая правка класса ломает снепшот, и разработчики привыкают бездумно обновлять (`--update`), теряя ценность.

**Good — snapshot на детерминированную структуру данных:**

```typescript
it("formats api error to stable shape", () => {
	expect(formatApiError(raw)).toMatchInlineSnapshot(`{ "message": "...", "requestId": "..." }`);
});
```

**Bad — snapshot всего компонента:**

```typescript
// Ломается при любой правке className/структуры; ревьюер не понимает, баг это или ожидаемо
expect(container).toMatchSnapshot();
```

Предпочитай точечные ассерты (`getByRole`/`getByText`) вместо снепшота всего дерева.

### 5.8. Мок внешних зависимостей (vi.mock)

Тяжёлые/несущественные для теста зависимости мокаются на уровне модуля. Реальный `issue-card.test.tsx` мокает роутер и drag-and-drop:

```typescript
vi.mock("@tanstack/react-router", () => ({
	Link: ({ children, onClick, className }: {
		children: React.ReactNode; onClick?: (e: React.MouseEvent) => void; className?: string;
	}) => (
		<a href="#" onClick={onClick} className={className}>{children}</a>
	),
}));

vi.mock("@atlaskit/pragmatic-drag-and-drop/element/adapter", () => ({
	draggable: vi.fn(() => vi.fn()),
}));
```

**Правило:** мокать только то, что мешает изоляции (роутер, DnD, анимации). Не мокать сам тестируемый компонент.

### 5.9. Чек-лист frontend-теста

- [ ] node_modules слинкованы (`link-testframework-node-modules.mjs`) перед прогоном.
- [ ] Провайдеры — через `renderWithProviders`, свой QueryClient.
- [ ] Запросы by role/label, не CSS-селекторы.
- [ ] `userEvent` с `await`, не сырой `fireEvent` для сложных действий.
- [ ] API замокан (MSW с `onUnhandledRequest: "error"` или `vi.spyOn`) — нет реальной сети.
- [ ] Хуки — `renderHook` + `act`.
- [ ] Глобальные сторы/sessionStorage сброшены в `beforeEach`, `vi.restoreAllMocks()`.
- [ ] Snapshot — только на стабильные структуры, не на всё дерево.
- [ ] `npx tsc --noEmit` и Biome (`npm run lint`) зелёные.

---
