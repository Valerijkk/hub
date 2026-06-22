/**
 * E2E-компонент Hub: messenger.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../../shared/fixtures";
import { MessengerPage } from "../../shared/pages";
import { AuthSteps } from "../../shared/steps";

// E2E-группа: Мессенджер (FB-08).
test.describe("Messenger (FB-08)", () => {
	// Проверяет: отправка сообщения в seed-канале.
	test("send message in seeded channel", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const messenger = new MessengerPage(page);
		await auth.loginAsDev(page);

		await messenger.openInbox();
		await expect(messenger.composer).toBeVisible({ timeout: 30_000 });
		const messageText = `E2E ping ${Date.now()}`;
		await messenger.sendMessage(messageText);
	});
});
