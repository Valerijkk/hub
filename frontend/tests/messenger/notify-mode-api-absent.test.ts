import { describe, expect, it } from "vitest";
import * as messengerApi from "@/api/messenger";

describe("messenger notify-mode API absent (A11-005)", () => {
	it("does not export updateChannelNotifyMode (no BE route)", () => {
		expect(messengerApi).not.toHaveProperty("updateChannelNotifyMode");
	});

	it("still exports channel mute helpers that map to registered routes", () => {
		expect(typeof messengerApi.muteChannel).toBe("function");
		expect(typeof messengerApi.unmuteChannel).toBe("function");
	});
});
