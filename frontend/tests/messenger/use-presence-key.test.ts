import { describe, expect, it } from "vitest";
import { presenceQueryKeyPart } from "@/hooks/use-presence";

describe("presenceQueryKeyPart (A10-004)", () => {
	it("does not mutate the input array", () => {
		const ids = ["b", "a", "c"];
		const snapshot = [...ids];

		const key = presenceQueryKeyPart(ids);

		expect(key).toBe("a,b,c");
		expect(ids).toEqual(snapshot);
	});

	it("returns stable order for the same set", () => {
		expect(presenceQueryKeyPart(["z", "m", "a"])).toBe(
			presenceQueryKeyPart(["a", "z", "m"]),
		);
	});
});
