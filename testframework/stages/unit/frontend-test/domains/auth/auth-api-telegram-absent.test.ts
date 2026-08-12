import { describe, expect, it } from "vitest";
import * as authApi from "@/api/auth";

describe("auth API telegram login absent (A11-004)", () => {
	it("does not export loginTelegram or any telegram auth request helper", () => {
		expect(authApi).not.toHaveProperty("loginTelegram");
		expect(Object.keys(authApi).some((k) => /telegram/i.test(k))).toBe(false);
	});

	it("still exports magic-link and profile endpoints", () => {
		expect(typeof authApi.sendMagicLink).toBe("function");
		expect(typeof authApi.verifyMagicLink).toBe("function");
		expect(typeof authApi.updateProfile).toBe("function");
	});
});
