# 02 — Page Object Model (канон Hub)

> Часть [Testing Guide](../TESTING_GUIDE.md). Playbook: [`.cursor/playbooks/testing.md`](../../.cursor/playbooks/testing.md) §6.  
> E2E-обзор: [05-e2e-playwright.md](05-e2e-playwright.md).

---

## Слои

```
BasePage          stages/e2e/shared/pages/base.page.ts
   ▲
Page Object       stages/e2e/shared/pages/{login,app-shell,messenger}.page.ts
   ▲
Steps             stages/e2e/shared/steps/{auth}.steps.ts
   ▲
Spec              stages/e2e/domains/{domain}/*.spec.ts
```

Реэкспорт: `stages/e2e/shared/pages/index.ts`, `shared/steps/index.ts`.

---

## BasePage

Реальный код: `base.page.ts`.

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

export async function expectVisible(loc: Locator, timeout?: number): Promise<void> {
	await expect(loc).toBeVisible(timeout ? { timeout } : undefined);
}
```

**Правило:** общая навигация и assert-хелперы — здесь; доменные локаторы — нет.

---

## Page Object

Реальный код: `login.page.ts`.

```typescript
export class LoginPage extends BasePage {
	readonly emailInput = this.page.getByLabel("Email");
	readonly hubTitle = this.page.getByText("Hub").first();
	readonly devPanel = this.page.getByText("Демо-доступ");

	async open(): Promise<void> {
		await this.goto("/login");
		await expect(this.page).toHaveURL(/\/login/);
	}

	async expectBrandingLoaded(): Promise<void> {
		await expectVisible(this.hubTitle);
		await expectVisible(this.emailInput);
	}

	async devLoginAs(buttonName: string): Promise<void> {
		await expectVisible(this.devPanel, 15_000);
		await this.page.getByRole("button", { name: buttonName }).click();
		await expect(this.page).toHaveURL(/\/projects/, { timeout: 15_000 });
	}
}
```

| Good | Bad |
|------|-----|
| Локаторы как `readonly` поля класса | Селекторы размазаны по spec |
| Методы = действия пользователя (`open`, `devLoginAs`) | Методы = «кликни третий div» |
| `getByRole` / `getByLabel` / `getByText` | `page.locator("#app > div:nth-child(2)")` |
| Одна страница = один файл | `HubPages.ts` на 2000 строк |

Существующие POM: `LoginPage`, `AppShellPage`, `MessengerPage`.

---

## Steps

Реальный код: `auth.steps.ts` — SRP: только аутентификация / навигация после входа.

```typescript
export class AuthSteps {
	private readonly loginPage: LoginPage;
	private readonly shell: AppShellPage;

	constructor(page: Page) {
		this.loginPage = new LoginPage(page);
		this.shell = new AppShellPage(page);
	}

	async openLoginAndExpectBranding(): Promise<void> {
		await this.loginPage.open();
		await this.loginPage.expectBrandingLoaded();
	}

	async devLoginAsAdminUI(): Promise<void> {
		await this.devLoginViaUI("Админ Хабов");
	}

	async loginAsAdmin(page: Page): Promise<DevLoginResult> {
		return this.loginAndInjectSession(page, ADMIN_EMAIL);
	}
}
```

Spec использует Steps, не Page напрямую (кроме простых smoke без бизнес-цепочки).

Предпочтительно оборачивать бизнес-шаги в `test.step(…)` — видно в HTML/trace и уходит в Allure (playbook §6.1).

---

## Spec

Реальный код: `domains/auth/login.spec.ts`.

```typescript
import { skipUnlessBackend, skipUnlessDevMode, test } from "../../shared/fixtures";
import { AuthSteps } from "../../shared/steps";

test.describe("Login flow", () => {
	test("login page loads with Hub branding", async ({ page }) => {
		const auth = new AuthSteps(page);
		await auth.openLoginAndExpectBranding();
	});

	test("dev login panel authenticates and redirects to projects", async ({
		page, backendAvailable, devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);
		const auth = new AuthSteps(page);
		await auth.devLoginAsAdminUI();
	});
});
```

---

## Приоритет локаторов

| # | Способ | Пример |
|---|--------|--------|
| 1 | Role | `getByRole("button", { name: "…" })` |
| 2 | Label | `getByLabel("Email")` |
| 3 | Text | `getByText("Hub")` |
| 4 | Test id | `getByTestId("…")` — только если нет role/label |
| 5 | CSS | последний resort |

---

## Чек-лист нового POM

- [ ] Файл в `shared/pages/`, экспорт в `index.ts`
- [ ] extends `BasePage`
- [ ] Нет `waitForTimeout`
- [ ] Steps обновлены, если появился новый бизнес-flow
- [ ] Spec не содержит сырых селекторов страницы
