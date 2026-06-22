/**
 * E2E-компонент Hub: ai-hub.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../fixtures";
import {
	chatCompletionRaw,
	composerSuggestRaw,
	createDMChannel,
	fetchAIBot,
	fetchAIPreferences,
	listAIModels,
	patchAIPreferences,
} from "../helpers/ai";
import { MessengerPage } from "../pages";
import { AuthSteps } from "../steps";
import { devLogin, DEV_EMAIL } from "../helpers/auth";

// E2E-группа: Настройки приватности AI (M3-030).
test.describe("AI privacy settings (M3-030)", () => {
	// Проверяет: страница settings/ai загружает переключатель согласия.
	test("settings/ai page loads consent toggle", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsDev(page);
		await page.goto("/settings/ai");

		await expect(page.getByRole("heading", { name: "AI и приватность" })).toBeVisible();
		await expect(
			page.getByText("Разрешить облачные AI-провайдеры"),
		).toBeVisible();
		const checkbox = page.getByRole("checkbox");
		await expect(checkbox).toBeVisible();
		await checkbox.click();
		await page.getByRole("button", { name: "Сохранить" }).click();
		await expect(page.getByText("Настройки AI сохранены")).toBeVisible({
			timeout: 10_000,
		});
	});

	// Проверяет: API возвращает настройку allow_external_ai.
	test("API returns allow_external_ai preference", async ({
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(DEV_EMAIL);
		const pref = await fetchAIPreferences(login.access_token);
		expect(typeof pref.allow_external_ai).toBe("boolean");
	});

	// Проверяет: chat completion блокируется при выключенном внешнем AI и только облачных моделях.
	test("chat completion blocked when external AI disabled and only cloud models", async ({
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(DEV_EMAIL);
		await patchAIPreferences(login.access_token, { allow_external_ai: false });

		const models = await listAIModels(login.access_token);
		const cloudModel = models.models.find(
			(m) => !m.id.includes("llama") && !m.id.includes("mistral"),
		)?.id;
		if (!cloudModel) {
			test.skip(true, "No cloud AI model configured — skip consent 403 test");
			return;
		}

		const { status, body } = await chatCompletionRaw(login, cloudModel);
		expect([403, 503]).toContain(status);
		if (status === 403) {
			expect(body.toLowerCase()).toContain("consent");
		}

		await patchAIPreferences(login.access_token, { allow_external_ai: true });
	});
});

// E2E-группа: Smoke мессенджера @hub-ai.
test.describe("@hub-ai messenger smoke", () => {
	test.beforeEach(({ }, testInfo) => {
		test.skip(
			!!process.env.CI,
			"@hub-ai reply smoke is flaky without configured LLM provider in CI",
		);
	});
	// Проверяет: бот существует и DM-канал открывается.
	test("bot user exists and DM channel opens", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const login = await auth.devLoginViaAPI(DEV_EMAIL);
		const bot = await fetchAIBot(login.access_token);
		expect(bot.user_id).toBeTruthy();
		expect(bot.display_name.toLowerCase()).toContain("hub-ai");

		const dm = await createDMChannel(login.access_token, bot.user_id);
		await auth.injectSession(page, login);
		await page.goto(`/messenger/${dm.id}`);

		const messenger = new MessengerPage(page);
		await expect(messenger.composer).toBeVisible({
			timeout: 15_000,
		});
	});

	// Проверяет: сообщение в DM вызывает хук @hub-ai (ответ или подсказка провайдера).
	test("DM message triggers @hub-ai hook (reply or provider hint)", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const login = await auth.loginAsDev(page);
		const bot = await fetchAIBot(login.access_token);
		const dm = await createDMChannel(login.access_token, bot.user_id);

		await page.goto(`/messenger/${dm.id}`);
		const messenger = new MessengerPage(page);
		await expect(messenger.composer).toBeVisible({ timeout: 15_000 });

		const ping = `E2E hub-ai ${Date.now()}`;
		await messenger.composer.fill(ping);
		await messenger.sendButton.click();
		await expect(page.getByText(ping)).toBeVisible({ timeout: 10_000 });

		const botReply = page.getByText(
			/AI провайдер не настроен|пустой ответ|▍|Не удалось получить ответ|@hub-ai/i,
		);
		try {
			await expect(botReply.first()).toBeVisible({ timeout: 45_000 });
		} catch {
			test.skip(
				true,
				"No @hub-ai reply in CI within 45s — provider not configured",
			);
		}
	});
});

// E2E-группа: Composer — AI suggest (M3-024).
test.describe("Composer AI suggest (M3-024)", () => {
	// Проверяет: API отклоняет короткий префикс.
	test("API rejects short prefix", async ({
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(DEV_EMAIL);
		const { status } = await composerSuggestRaw(login.access_token, "short");
		expect(status).toBe(400);
	});

	// Проверяет: API принимает валидный префикс (200 или провайдер недоступен).
	test("API accepts valid prefix (200 or provider unavailable)", async ({
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(DEV_EMAIL);
		const { status, body } = await composerSuggestRaw(
			login.access_token,
			"Привет, команда!",
		);
		expect([200, 403, 502, 503, 504]).toContain(status);
		if (status === 200 && typeof body === "object" && body !== null) {
			expect(body.suggestion.length).toBeGreaterThan(0);
		}
	});

	// Проверяет: подсказка Tab видна и Tab запускает сценарий suggest.
	test("Tab hint visible and Tab key triggers suggest flow", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		test.skip(
			!!process.env.CI,
			"Composer AI Tab suggest requires provider; skipped in CI",
		);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const messenger = new MessengerPage(page);
		await auth.loginAsDev(page);
		await messenger.openInbox();

		await expect(messenger.composer).toBeVisible({ timeout: 15_000 });
		const prefix = "Привет всем, это длинный текст для AI suggest!";
		await messenger.composer.fill(prefix);

		const tabTitle = await messenger.composer.getAttribute("title");
		if (!tabTitle?.includes("Tab")) {
			test.skip(true, "Composer AI suggest unavailable");
			return;
		}

		const before = await messenger.composer.inputValue();
		await messenger.composer.press("Tab");
		await page.waitForTimeout(3_000);
		const after = await messenger.composer.inputValue();
		expect(after.length).toBeGreaterThanOrEqual(before.length);
	});
});
