import { describe, expect, it } from "vitest";

describe("FE-P1-MESSENGER · voice onstop order (W9-A3-003)", () => {
	it("resolves when onstop is set before stop", async () => {
		let onstop: (() => void) | null = null;
		const recorder = {
			stop() {
				onstop?.();
			},
			set onstop(h: (() => void) | null) {
				onstop = h;
			},
		};
		const result = await new Promise<"resolved">((resolve) => {
			recorder.onstop = () => resolve("resolved");
			recorder.stop();
		});
		expect(result).toBe("resolved");
	});

	it("mutation: stop before onstop leaves Promise hung", async () => {
		let onstop: (() => void) | null = null;
		const recorder = {
			stop() {
				onstop?.();
			},
			set onstop(h: (() => void) | null) {
				onstop = h;
			},
		};

		const result = await Promise.race([
			new Promise<"resolved">((resolve) => {
				recorder.stop();
				recorder.onstop = () => resolve("resolved");
			}),
			new Promise<"hung">((resolve) => {
				setTimeout(() => resolve("hung"), 30);
			}),
		]);

		expect(result).toBe("hung");
	});
});
