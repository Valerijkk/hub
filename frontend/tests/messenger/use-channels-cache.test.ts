import { describe, expect, it } from "vitest";
import type { Channel } from "@/api/messenger";
import {
	patchChannelInList,
	upsertChannelInList,
} from "@/hooks/use-channels";

function makeChannel(partial: Partial<Channel> & Pick<Channel, "id" | "name">): Channel {
	return {
		channel_type: "public",
		is_readonly: false,
		created_by: "user-1",
		is_starred: false,
		created_at: "2026-07-01T00:00:00Z",
		updated_at: "2026-07-01T00:00:00Z",
		...partial,
	};
}

describe("use-channels Query cache helpers (A10-001)", () => {
	it("upsertChannelInList prepends a new channel", () => {
		const existing = [makeChannel({ id: "a", name: "Alpha" })];
		const created = makeChannel({ id: "b", name: "Beta" });

		const next = upsertChannelInList(existing, created);

		expect(next.map((c) => c.id)).toEqual(["b", "a"]);
	});

	it("upsertChannelInList dedupes by id (merge)", () => {
		const existing = [makeChannel({ id: "a", name: "Old" })];
		const created = makeChannel({ id: "a", name: "New" });

		const next = upsertChannelInList(existing, created);

		expect(next).toHaveLength(1);
		expect(next[0]?.name).toBe("New");
	});

	it("upsertChannelInList handles empty cache", () => {
		const created = makeChannel({ id: "c", name: "Gamma" });
		expect(upsertChannelInList(undefined, created)).toEqual([created]);
	});

	it("patchChannelInList updates matching channel only", () => {
		const existing = [
			makeChannel({ id: "a", name: "Alpha" }),
			makeChannel({ id: "b", name: "Beta" }),
		];

		const next = patchChannelInList(existing, "b", { name: "Beta Renamed" });

		expect(next?.[0]?.name).toBe("Alpha");
		expect(next?.[1]?.name).toBe("Beta Renamed");
	});

	it("patchChannelInList returns undefined for empty cache", () => {
		expect(patchChannelInList(undefined, "x", { name: "Nope" })).toBeUndefined();
	});
});
