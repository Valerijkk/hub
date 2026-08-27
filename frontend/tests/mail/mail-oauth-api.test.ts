import { beforeEach, describe, expect, it, vi } from "vitest";
import {
	beginMailOAuth,
	beginMailSetupOAuth,
	completeMailOAuth,
	completeMailSetupOAuth,
} from "@/api/mail";
import { api } from "@/api/client";

describe("mail OAuth API", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("starts setup and regular OAuth with no body and forwards options", async () => {
		const setupOptions = { signal: new AbortController().signal };
		const regularOptions = { signal: new AbortController().signal };
		const authorization = {
			authorization_url: "https://mail.example.test/login",
			expires_at: "2026-08-27T12:00:00Z",
		};
		const postSpy = vi
			.spyOn(api, "post")
			.mockResolvedValueOnce(authorization)
			.mockResolvedValueOnce(authorization);

		await expect(beginMailSetupOAuth(setupOptions)).resolves.toEqual(
			authorization,
		);
		await expect(beginMailOAuth(regularOptions)).resolves.toEqual(authorization);

		expect(postSpy).toHaveBeenNthCalledWith(
			1,
			"/mail/setup/oauth/authorize",
			undefined,
			setupOptions,
		);
		expect(postSpy).toHaveBeenNthCalledWith(
			2,
			"/mail/oauth/authorize",
			undefined,
			regularOptions,
		);
	});

	it("completes setup and regular callbacks with exact requests and options", async () => {
		const setupRequest = { state: "setup-state", code: "setup-code" };
		const regularRequest = { state: "regular-state", code: "regular-code" };
		const setupOptions = { signal: new AbortController().signal };
		const regularOptions = { signal: new AbortController().signal };
		const postSpy = vi
			.spyOn(api, "post")
			.mockResolvedValueOnce(undefined)
			.mockResolvedValueOnce(undefined);

		await expect(
			completeMailSetupOAuth(setupRequest, setupOptions),
		).resolves.toBeUndefined();
		await expect(
			completeMailOAuth(regularRequest, regularOptions),
		).resolves.toBeUndefined();

		expect(postSpy).toHaveBeenNthCalledWith(
			1,
			"/mail/setup/oauth/callback",
			setupRequest,
			setupOptions,
		);
		expect(postSpy).toHaveBeenNthCalledWith(
			2,
			"/mail/oauth/callback",
			regularRequest,
			regularOptions,
		);
	});
});
