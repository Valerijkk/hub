import { describe, expect, it } from "vitest";
import type { Channel, ChannelCapability } from "@/api/messenger";
import { canChannel } from "@/api/messenger";

function channel(partial: Partial<Channel> = {}): Channel {
	return {
		id: "ch-1",
		channel_type: "public",
		name: "General",
		is_readonly: false,
		created_by: "owner-looking-user",
		is_starred: false,
		created_at: "2026-08-26T00:00:00Z",
		updated_at: "2026-08-26T00:00:00Z",
		...partial,
	};
}

describe("Phase 151 recipient channel capabilities", () => {
	it("fails closed when the server capability snapshot is absent", () => {
		const legacy = channel();
		expect(canChannel(legacy, "can_read")).toBe(false);
		expect(canChannel(legacy, "can_post")).toBe(false);
		expect(canChannel(legacy, "can_join")).toBe(false);
	});

	it("does not infer a permission from readonly, creator, type, or project metadata", () => {
		const metadataOnly = channel({
			is_readonly: false,
			project_id: "project-1",
			created_by: "current-user",
		});
		const capabilities: ChannelCapability[] = [
			"can_post",
			"can_reply",
			"can_follow",
		];

		for (const capability of capabilities) {
			expect(canChannel(metadataOnly, capability)).toBe(false);
		}
	});

	it("allows only an explicit true value from the recipient DTO", () => {
		const browse = channel({
			capabilities: {
				is_member: false,
				has_active_project_membership: false,
				can_read: true,
				can_join: true,
				can_post: false,
				can_reply: false,
				can_follow: false,
			},
		});

		expect(canChannel(browse, "can_read")).toBe(true);
		expect(canChannel(browse, "can_join")).toBe(true);
		expect(canChannel(browse, "can_post")).toBe(false);
		expect(canChannel(browse, "can_follow")).toBe(false);
	});
});
