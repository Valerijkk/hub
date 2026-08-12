# E2E: целевой Python-suite и исторический Playwright-контекст

> Часть плейбука тестирования. Канон и оглавление — [`../testing.md`](../testing.md).

Целевой E2E-контракт Hub — единый Python-suite в
`testframework/tests/e2e/**`. `smoke` — pytest marker для выбора быстрых
сценариев из этого единственного suite. `full` — не pytest marker: это профиль
будущего outer launcher, выбирающий полный тот же `testframework/tests/e2e/**`
suite. Отдельных каталогов smoke/full не создаём.

Пакет, outer launcher и целевые suites пока `NOT_READY`; для этой
документационной волны runtime — `NOT_RUN`. До принятой реализации никакая
прямая host/Docker/legacy команда не является fallback, gate или runtime
evidence. Далее сохранены исторические Playwright-практики POM, шагов и
стабильности как контекст миграции, а не как инструкция для целевого запуска.

---

## 6. E2E (исторический Playwright + POM)

Исторические E2E-тесты (20 Playwright spec) проверяли критические сквозные
пути в реальном браузере и находились в
`testframework/stages/e2e/domains/{domain}/`; общий код был в `shared/`.
Эта топология и её прежние job не определяют целевой Python E2E-contract и не
могут использоваться как fallback или доказательство runtime.

### 6.1. Архитектура: BasePage → Page Object → Steps → Spec

Слоистая структура, каждый слой со своей ответственностью (SRP):

```
BasePage (общая навигация)
   ▲
   │ extends
Page Object (локаторы + действия одной страницы)   ← shared/pages/
   ▲
   │ использует
Steps (бизнес-шаги поверх POM)                       ← shared/steps/
   ▲
   │ использует
Spec (сценарий + ассерты)                            ← domains/{domain}/
```

**BasePage** — общий предок (реальный `base.page.ts`):

```typescript
export abstract class BasePage {
	constructor(protected readonly page: Page) {}

	async goto(path: string): Promise<void> {
		await this.page.goto(path);
	}

	protected locator(selector: string): Locator {
		return this.page.locator(selector);
	}
}

// Хелпер этапа Assert (AAA).
export async function expectVisible(loc: Locator, timeout?: number): Promise<void> {
	await expect(loc).toBeVisible(timeout ? { timeout } : undefined);
}
```

**Page Object** — локаторы и действия конкретной страницы (реальный `login.page.ts`):

```typescript
export class LoginPage extends BasePage {
	readonly emailInput = this.page.getByLabel("Email");
	readonly hubTitle = this.page.getByText("Hub").first();
	readonly devPanel = this.page.getByText("Демо-доступ");

	async open(): Promise<void> {
		await this.goto("/login");
		await expect(this.page).toHaveURL(/\/login/);
	}

	async devLoginAs(buttonName: string): Promise<void> {
		await expectVisible(this.devPanel, 15_000);
		await this.page.getByRole("button", { name: buttonName }).click();
		await expect(this.page).toHaveURL(/\/projects/, { timeout: 15_000 });
	}
}
```

**Steps** — бизнес-шаги (реальный `auth.steps.ts`): `AuthSteps.loginAsAdmin`, `loginAsDev`, `logoutFromProfile`. Spec оперирует шагами, не локаторами.

**(2026) Бизнес-шаги оборачивай в нативный `test.step()`.** Шаги видны в HTML-отчёте, trace viewer и UI Mode, а `allure-playwright` конвертирует их в Allure-шаги автоматически — один API покрывает оба отчёта (вместо отдельного `allure.step` в Steps-слое):

```typescript
export class AuthSteps {
	async loginAsAdmin(page: Page): Promise<void> {
		await test.step("Login as admin (seed + dev-login)", async () => {
			const login = await devLogin("admin@hub.local");
			await authenticatePage(page, login);
		});
	}
}
```

### 6.2. Устойчивые локаторы

Приоритет локаторов (как в RTL, отражает пользовательское восприятие):

| Приоритет | Локатор | Пример |
|-----------|---------|--------|
| 1 | `getByRole` | `page.getByRole("button", { name: "Создать проект" })` |
| 2 | `getByLabel` | `page.getByLabel("Email")` |
| 3 | `getByText` | `page.getByText("Демо-доступ")` |
| 4 | `data-testid` / `#id` | `page.locator("#proj-name")` — когда роль/текст неоднозначны |

`data-testid`/`#id` уместны для полей форм без стабильного label и контейнеров (`#main-content`, `#proj-name`, `#proj-key` в реальном `project-lifecycle.spec.ts`). Не завязываться на CSS-классы и порядок в DOM.

**Good — роль + имя:**

```typescript
await page.getByRole("button", { name: "Создать проект" }).click();
await expect(page.getByRole("tab", { name: "Обзор" })).toBeVisible({ timeout: 15_000 });
```

**Bad — CSS-цепочка:**

```typescript
await page.locator("div.form > button.primary:nth-child(3)").click(); // ломается от любой правки вёрстки
```

### 6.3. Ожидания: auto-wait и web-first assertions, без sleep

Playwright сам ждёт актуальности (видимость, кликабельность) в `expect(...).toBeVisible()` и действиях. Никогда не использовать `page.waitForTimeout` (sleep) для синхронизации.

**Good — web-first assertions с явным таймаутом на медленные операции:**

```typescript
await page.getByRole("button", { name: "Создать проект" }).click();
// Ждём навигацию по URL, а не фиксированную паузу.
await expect(page).toHaveURL(/\/projects\/[0-9a-f-]+/, { timeout: 30_000 });
await expect(page.locator("#main-content")).toBeVisible({ timeout: 15_000 });
```

**Bad — sleep:**

```typescript
await page.getByRole("button", { name: "Создать проект" }).click();
await page.waitForTimeout(3000); // флейк: иногда мало, иногда медленно; всегда тратит время
expect(page.url()).toContain("/projects/");
```

Типовые ожидания: `toHaveURL`, `toBeVisible`, `toHaveText`, `toHaveCount`. Для навигации — `toHaveURL(regex)`.

### 6.4. Фикстуры и условный пропуск

E2E-фикстуры (реальный `fixtures.ts`) сообщают, доступен ли backend/dev-mode, и позволяют корректно пропускать тесты, требующие инфраструктуры:

```typescript
export const test = base.extend<HubFixtures>({
	backendAvailable: async ({}, use) => { await use(await isBackendReachable()); },
	devModeAvailable: async ({}, use) => { await use(await isDevModeAvailable()); },
});

export function skipUnlessBackend(backendAvailable: boolean): void {
	test.skip(!backendAvailable, "Backend not running — start with: docker compose ...");
}
```

Smoke-тесты (job `e2e-smoke`) идут без backend, поэтому spec с backend-зависимостью объявляют пропуск в начале:

```typescript
test("create project ...", async ({ page, backendAvailable, devModeAvailable }) => {
	skipUnlessBackend(backendAvailable);
	skipUnlessDevMode(devModeAvailable);
	// ...основной сценарий
});
```

### 6.5. Dev-login (аутентификация в E2E)

E2E не проходит magic-link через email. Вместо этого — dev-login (только при `DEV_MODE=true` на backend). Реальный `helpers/auth.ts` сидит демо-данные, логинится и инъектит сессию в sessionStorage:

```typescript
export async function devLogin(email: string, apiUrl = getApiUrl()): Promise<DevLoginResult> {
	await seedDevData(apiUrl); // POST /api/dev/seed
	const res = await fetch(`${apiUrl}/api/dev/login`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ email }),
		signal: AbortSignal.timeout(10_000),
	});
	if (!res.ok) throw new Error(`dev login failed: ${res.status}`);
	return (await res.json()) as DevLoginResult;
}

export async function authenticatePage(page: Page, login: DevLoginResult): Promise<void> {
	await page.addInitScript((payload) => {
		sessionStorage.setItem("hub-auth", JSON.stringify(payload));
	}, { state: { accessToken: login.access_token, isAuthenticated: true, user: login.user }, version: 0 });
}
```

Через `AuthSteps` это выглядит одной строкой в spec:

```typescript
const auth = new AuthSteps(page);
await auth.loginAsAdmin(page); // seed + dev-login + inject session
```

**Важно (TZ-инвариант):** dev-login/DEV_MODE — только для локали и CI e2e-full. В prod `DEV_MODE=true` запрещён (rule 00, deploy без DEV_MODE).

**(2026) `storageState`: логин один раз вместо dev-login в каждом тесте.** Канон Playwright — auth setup-project: логинится один раз, сохраняет состояние в json; остальные projects декларируют его через `use` и стартуют уже залогиненными (экономия секунд на каждый spec):

```typescript
// auth.setup.ts (project "setup")
setup("authenticate", async ({ page }) => {
	const login = await devLogin("admin@hub.local");
	await authenticatePage(page, login);
	await page.goto("/projects");
	await page.context().storageState({ path: "playwright/.auth/admin.json" });
});

// playwright.config.ts
projects: [
	{ name: "setup", testMatch: /auth\.setup\.ts/ },
	{
		name: "chromium-full",
		dependencies: ["setup"],
		use: { storageState: "playwright/.auth/admin.json" },
	},
],
```

Нюанс Hub: сессия хранится в **sessionStorage**, который `storageState` из коробки не сохраняет (только cookies + localStorage) — нужна адаптация (инъекция sessionStorage через `addInitScript` из сохранённого json) или storageState-совместимое хранение для тестов. До адаптации per-test dev-login (выше) остаётся рабочим fallback; тесты, мутирующие сессию (logout), выполняют логин сами.

### 6.6. Изоляция E2E

- **Уникальные данные на прогон:** суффикс из времени, чтобы параллельные worker'ы и повторные прогоны не конфликтовали:

```typescript
const suffix = Date.now().toString(36).toUpperCase().slice(-6);
const projectName = `E2E Project ${suffix}`;
const projectKey = `E${suffix}`.slice(0, 10);
```

- **Каждый тест логинится сам** (в `beforeEach` или в начале теста) — не полагаться на состояние предыдущего.
- **Не шарить созданные сущности** между тестами; чистка — через восстановление/удаление внутри сценария (как в `project-lifecycle`: создать → удалить → восстановить).

### 6.7. Smoke marker и full launcher profile

| Термин | Значение | Область выбора |
|--------|----------|----------------|
| `smoke` | pytest marker для быстрых риск-ориентированных E2E-сценариев | `testframework/tests/e2e/**` |
| `full` | профиль будущего outer launcher, выбирающий полный E2E-suite | тот же `testframework/tests/e2e/**` |

`full` не выражает marker-семантику и не создаёт отдельный путь, каталог или
pytest-выборку. Новый критический user-path проектируется риск-ориентированно в
едином E2E-suite; попадание в smoke определяется только marker-стратегией, а
полный охват — launcher profile после его отдельной реализации и приёмки.

Пока launcher и package `NOT_READY`, а runtime `NOT_RUN`, нельзя заменять этот
контракт историческими Playwright jobs, legacy runners, прямыми host/Docker
командами или заявлением об успешном прогоне.

### 6.8. Исторический контекст: шардинг и merge-reports

> Неисполняемое историческое описание. Оно не является текущей CI-механикой,
> gate, fallback или runtime evidence и не даёт команды для запуска.

В прежнем Playwright-контуре полный набор из 20 spec предполагал шардинг между
машинами с blob-отчётами и последующим объединением в HTML-отчёт. Этот подход
мог сокращать wall-time, но не заменяет будущий outer launcher и не переносит
его authority на legacy Playwright tooling.

- Параллелизм внутри одного исторического файла допускался только для независимых тестов; изоляция данных из раздела 6.6 оставалась обязательной.
- Исторический CI-бюджет ограничивал применение шардирования. Для целевого Python-suite будущая стратегия параллелизма, артефактов и resource limits принадлежит только outer launcher после отдельной реализации и приёмки.

### 6.9. Чек-лист E2E-теста

- [ ] Логика вынесена в Page Object/Steps, spec читается как сценарий.
- [ ] Локаторы by role/label; `data-testid`/`#id` только при необходимости.
- [ ] Ожидания — web-first assertions (`toHaveURL`, `toBeVisible`), ноль `waitForTimeout`.
- [ ] Backend-зависимые spec вызывают `skipUnlessBackend/DevMode` в начале.
- [ ] Аутентификация через `AuthSteps`/dev-login (цель 2026 — setup-project + `storageState`), не ручной magic-link.
- [ ] (2026) Бизнес-шаги обёрнуты в `test.step()` — видны в HTML/trace и конвертируются в Allure.
- [ ] Уникальные данные на прогон (суффикс времени), тест самодостаточен.
- [ ] Новый критический путь расположен в едином `testframework/tests/e2e/**`; smoke отмечен pytest marker только при соответствии его риск-стратегии, а full остаётся launcher profile.
- [ ] Прогон стабилен (нет флейков при повторе) — раздел 13.

---
