/**
 * Хелперы E2E для AI Hub API (composer, settings).
 */
import { getApiUrl } from "./backend";
import type { DevLoginResult } from "./auth";

export interface AIBotInfo {
	user_id: string;
	display_name: string;
	email: string;
}

export interface AIPreferences {
	default_model?: string;
	allow_external_ai?: boolean;
}

export async function fetchAIBot(
	token: string,
	apiUrl = getApiUrl(),
): Promise<AIBotInfo> {
	const res = await fetch(`${apiUrl}/api/ai/bot`, {
		headers: { Authorization: `Bearer ${token}` },
		signal: AbortSignal.timeout(10_000),
	});
	if (!res.ok) {
		throw new Error(`ai/bot failed: ${res.status} ${await res.text()}`);
	}
	return (await res.json()) as AIBotInfo;
}

export async function fetchAIPreferences(
	token: string,
	apiUrl = getApiUrl(),
): Promise<AIPreferences> {
	const res = await fetch(`${apiUrl}/api/ai/preferences`, {
		headers: { Authorization: `Bearer ${token}` },
		signal: AbortSignal.timeout(10_000),
	});
	if (!res.ok) {
		throw new Error(`ai/preferences GET failed: ${res.status}`);
	}
	return (await res.json()) as AIPreferences;
}

export async function patchAIPreferences(
	token: string,
	body: Partial<AIPreferences>,
	apiUrl = getApiUrl(),
): Promise<{ status: number; body: string }> {
	const res = await fetch(`${apiUrl}/api/ai/preferences`, {
		method: "PATCH",
		headers: {
			Authorization: `Bearer ${token}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify(body),
		signal: AbortSignal.timeout(10_000),
	});
	return { status: res.status, body: await res.text() };
}

export async function chatCompletionRaw(
	login: DevLoginResult,
	model: string,
	apiUrl = getApiUrl(),
): Promise<{ status: number; body: string }> {
	const res = await fetch(`${apiUrl}/api/ai/v1/chat/completions`, {
		method: "POST",
		headers: {
			Authorization: `Bearer ${login.access_token}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify({
			model,
			messages: [{ role: "user", content: "ping" }],
		}),
		signal: AbortSignal.timeout(15_000),
	});
	return { status: res.status, body: await res.text() };
}

export async function createDMChannel(
	token: string,
	targetUserId: string,
	apiUrl = getApiUrl(),
): Promise<{ id: string }> {
	const res = await fetch(`${apiUrl}/api/messenger/dm`, {
		method: "POST",
		headers: {
			Authorization: `Bearer ${token}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify({ target_user_id: targetUserId }),
		signal: AbortSignal.timeout(10_000),
	});
	if (!res.ok) {
		throw new Error(`messenger/dm failed: ${res.status} ${await res.text()}`);
	}
	return (await res.json()) as { id: string };
}

export async function listAIModels(
	token: string,
	apiUrl = getApiUrl(),
): Promise<{ models: Array<{ id: string }> }> {
	const res = await fetch(`${apiUrl}/api/ai/models`, {
		headers: { Authorization: `Bearer ${token}` },
		signal: AbortSignal.timeout(10_000),
	});
	if (!res.ok) {
		throw new Error(`ai/models failed: ${res.status}`);
	}
	return (await res.json()) as { models: Array<{ id: string }> };
}

export interface ComposerSuggestResult {
	suggestion: string;
	model: string;
}

/** POST /api/ai/composer/suggest — статус и тело ответа или текст ошибки. */
export async function composerSuggestRaw(
	token: string,
	prefix: string,
	apiUrl = getApiUrl(),
): Promise<{ status: number; body: ComposerSuggestResult | string }> {
	const res = await fetch(`${apiUrl}/api/ai/composer/suggest`, {
		method: "POST",
		headers: {
			Authorization: `Bearer ${token}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify({ prefix }),
		signal: AbortSignal.timeout(20_000),
	});
	const text = await res.text();
	if (!res.ok) {
		return { status: res.status, body: text };
	}
	try {
		return { status: res.status, body: JSON.parse(text) as ComposerSuggestResult };
	} catch {
		return { status: res.status, body: text };
	}
}
