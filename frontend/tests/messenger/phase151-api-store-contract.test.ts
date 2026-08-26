import { beforeEach, describe, expect, it, vi } from "vitest";
import {
	followThread,
	getThread,
	markThreadRead,
	type UnreadSnapshot,
	unfollowThread,
} from "@/api/messenger";
import { api } from "@/api/client";
import { useMessengerStore } from "@/stores/messenger-store";

vi.mock("@/api/client", () => ({
	api: {
		get: vi.fn(),
		post: vi.fn(),
		delete: vi.fn(),
	},
}));

const firstSnapshot: UnreadSnapshot = {
	channels: { "ch-1": 2 },
	dms: { "dm-1": 3 },
	mentions: { "ch-1": 1 },
	threads: { "root-1": 4 },
	any_unread: true,
};

describe("Phase 151 exact thread API", () => {
	beforeEach(() => vi.clearAllMocks());

	it("uses the exact thread resource and receipt routes", async () => {
		vi.mocked(api.get).mockResolvedValue({ parent: {}, replies: [] });
		vi.mocked(api.post).mockResolvedValue({ advanced: true });
		vi.mocked(api.delete).mockResolvedValue({
			changed: true,
			mode: "manual_unfollow",
		});

		await getThread("ch-1", "root-1");
		await markThreadRead("ch-1", "root-1", "last-message-1");
		await followThread("ch-1", "root-1");
		await unfollowThread("ch-1", "root-1");

		expect(api.get).toHaveBeenCalledWith(
			"/messenger/channels/ch-1/threads/root-1",
		);
		expect(api.post).toHaveBeenNthCalledWith(
			1,
			"/messenger/channels/ch-1/threads/root-1/read",
			{ category: "followed_thread", message_id: "last-message-1" },
		);
		expect(api.post).toHaveBeenNthCalledWith(
			2,
			"/messenger/channels/ch-1/threads/root-1/follow",
		);
		expect(api.delete).toHaveBeenCalledWith(
			"/messenger/channels/ch-1/threads/root-1/follow",
		);
		expect(api.get).not.toHaveBeenCalledWith(
			expect.stringContaining("/messages/root-1/replies"),
		);
	});
});

describe("Phase 151 authoritative messenger state", () => {
	beforeEach(() => useMessengerStore.getState().reset());

	it("replaces all unread categories as one server snapshot", () => {
		useMessengerStore.getState().setUnreadSnapshot(firstSnapshot);
		expect(useMessengerStore.getState().unreadSnapshot).toEqual(firstSnapshot);

		const replacement: UnreadSnapshot = {
			channels: {},
			dms: {},
			mentions: {},
			threads: { "root-2": 1 },
			any_unread: true,
		};
		useMessengerStore.getState().setUnreadSnapshot(replacement);

		expect(useMessengerStore.getState().unreadSnapshot).toEqual(replacement);
		expect(
			useMessengerStore.getState().unreadSnapshot.channels["ch-1"],
		).toBeUndefined();
	});

	it("opens and closes a thread as an atomic channel/root pair", () => {
		useMessengerStore.getState().openThread("ch-1", "root-1");
		expect(useMessengerStore.getState().activeThread).toEqual({
			channelId: "ch-1",
			rootId: "root-1",
		});

		useMessengerStore.getState().closeThread();
		expect(useMessengerStore.getState().activeThread).toBeNull();
	});
});
