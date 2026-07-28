/**
 * Хелперы E2E: dev-login, seed и инъекция сессии в memory (без JWT в sessionStorage).
 */
import type { Page } from "@playwright/test";

import { getApiUrl } from "./backend";

export const ADMIN_EMAIL = "admin@hub.local";
export const DEV_EMAIL = "dev1@hub.local";

export interface DevLoginResult {
	access_token: string;
	refresh_token: string;
	expires_in: number;
	user: {
		id: string;
		email: string;
		name?: string;
		system_role: string;
		onboarding_completed: boolean;
		timezone: string;
		language: string;
	};
}

export async function seedDevData(apiUrl = getApiUrl()): Promise<void> {
	const res = await fetch(`${apiUrl}/api/dev/seed`, {
		method: "POST",
		signal: AbortSignal.timeout(30_000),
	});
	if (!res.ok) {
		throw new Error(`dev seed failed: ${res.status} ${await res.text()}`);
	}
}

export async function devLogin(
	email: string,
	apiUrl = getApiUrl(),
): Promise<DevLoginResult> {
	await seedDevData(apiUrl);

	const res = await fetch(`${apiUrl}/api/dev/login`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ email }),
		signal: AbortSignal.timeout(10_000),
	});
	if (!res.ok) {
		throw new Error(`dev login failed: ${res.status} ${await res.text()}`);
	}
	return (await res.json()) as DevLoginResult;
}

/**
 * Инъекция сессии до загрузки приложения (ADR-WS-JWT-HANDSHAKE-2026):
 * JWT только через memory bootstrap, не sessionStorage.
 */
export async function authenticatePage(
	page: Page,
	login: DevLoginResult,
): Promise<void> {
	const bootstrap = {
		accessToken: login.access_token,
		refreshToken: login.refresh_token,
		user: login.user,
		isAuthenticated: true,
	};

	await page.addInitScript((payload) => {
		localStorage.setItem("i18nextLng", "ru");
		(
			window as unknown as { __HUB_AUTH_BOOTSTRAP__: typeof payload }
		).__HUB_AUTH_BOOTSTRAP__ = payload;
	}, bootstrap);
}
