# 04 — Frontend Vitest + RTL

> Часть [Testing Guide](../TESTING_GUIDE.md). Playbook: [`.cursor/playbooks/testing.md`](../../.cursor/playbooks/testing.md) §5.  
> Путь: `testframework/stages/unit/frontend-test/domains/{domain}/`.  
> Владелец automation: **`qa-auto2`**. Unit рядом с фичей пишут **`fe1`/`fe2`**.

---

## Структура

```
stages/unit/frontend-test/
├── render.tsx                 # renderWithProviders
├── domains/
│   ├── auth/                  # auth-store, use-auth, client-api, …
│   ├── tracker/
│   ├── ai/
│   ├── admin/
│   ├── platform/
│   └── …
└── README.md
```

Перед прогоном: `node frontend/scripts/link-testframework-node-modules.mjs`.

---

## TestRenderer (обязательный паттерн)

Реальный код: `render.tsx`.

```typescript
export function renderWithProviders(ui: ReactElement, options: TestRenderOptions = {}) {
	const { queryClient = createTestQueryClient(), wrapper: ExtraWrapper, ...renderOptions } = options;
	// QueryClient: queries/mutations retry: false — детерминизм
	// Не Singleton: каждый вызов — свой client
	return { queryClient, ...render(ui, { wrapper: Wrapper, ...renderOptions }) };
}
```

| Good | Bad |
|------|-----|
| `renderWithProviders(<X />)` | Голый `render` без QueryClient для query-компонентов |
| Свой client на тест | Один глобальный `QueryClient` на suite |
| `retry: false` | Реальные ретраи → флейки / таймауты |

---

## Store / hooks (пример из репо)

`domains/auth/auth-store.test.ts`:

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

describe("auth-store", () => {
	beforeEach(() => {
		resetAuthStore();
	});

	it("logout clears tokens and isAuthenticated", () => {
		useAuthStore.getState().setTokens("access-abc", "refresh-xyz");
		useAuthStore.getState().logout();
		expect(useAuthStore.getState().isAuthenticated).toBe(false);
	});
});
```

**Изоляция:** `beforeEach` сбрасывает Zustand + `sessionStorage`. Без этого — зависимость от порядка.

---

## Queries и userEvent

Приоритет (как у пользователя):

1. `getByRole` / `findByRole`
2. `getByLabelText`
3. `getByText`
4. `getByTestId` — запасной

Взаимодействие — **`userEvent`**, не сырой `fireEvent` (кроме редких низкоуровневых кейсов).

```typescript
// Good
await userEvent.click(screen.getByRole("button", { name: /save/i }));

// Bad
fireEvent.click(container.querySelector(".btn-primary")!);
```

---

## Моки сети

| Способ | Когда |
|--------|-------|
| **MSW** | HTTP-контракт компонента / хука с fetch |
| **`vi.spyOn` / `vi.mock`** | узкий модуль, не HTTP |

```typescript
// Good — MSW, необработанный запрос = ошибка
server.listen({ onUnhandledRequest: "error" });

// Bad — реальный backend в unit
await fetch("http://localhost:8080/api/…");
```

Подробности MSW vs spy: playbook §5.4.

---

## Snapshot

Допустимы точечные; **запрещён** snapshot всего дерева «чтобы было покрытие». Обновление `--update` без ревью UI — anti-pattern (playbook §5.7).

---

## Именование

| Файл | Содержание |
|------|------------|
| `auth-store.test.ts` | Zustand store |
| `use-auth.test.ts` | хук |
| `*-api.test.ts` | клиент API / схемы |
| `*-schema.test.ts` | zod/валидация |

`describe` = субъект; `it` = поведение на английском.

---

## Happy / negative на FE unit

| Тип | Что |
|-----|-----|
| Happy | рендер, успешный submit, store setTokens |
| Negative | ошибка API → сообщение; disabled кнопка; validation fail |
| Edge | пустой список, длинная строка, null optional fields |

Не дублировать E2E: unit проверяет компонент/стор; E2E — сквозной путь.

---

## Прогон

```bash
cd frontend && node scripts/link-testframework-node-modules.mjs
cd frontend && npx tsc --noEmit && npm run test:run
cd frontend && npx vitest run ../testframework/stages/unit/frontend-test/domains/auth
```

---

## Чек-лист Vitest

- [ ] Путь `domains/{domain}/`, владелец слоя соблюдён
- [ ] `renderWithProviders` / сброс store в `beforeEach`
- [ ] Нет реальной сети; MSW или mock
- [ ] Queries by role/label; `userEvent`
- [ ] TC-ID при automation-задаче + колонка в `TESTING.md`
- [ ] Biome / tsc чисты
