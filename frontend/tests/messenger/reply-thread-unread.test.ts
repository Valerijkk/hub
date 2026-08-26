import { describe, expect, it } from "vitest";
import type { Message, UnreadSnapshot } from "@/api/messenger";
import { findParentInMessagesCache } from "@/components/messenger/thread-panel";
import { useMessengerStore } from "@/stores/messenger-store";

function msg(id: string): Message {
	return {
		id,
		channel_id: "ch-1",
		sender_id: "u-1",
		content_text: "msg " + id,
		reply_count: 0,
		created_at: "2026-07-28T12:00:00Z",
	};
}

describe("Messenger thread parent cache", () => {
	it("finds parent in infinite pages, not only the first message", () => {
		const cache = {
			pages: [[msg("newest")], [msg("parent-42"), msg("older")]],
		};
		expect(findParentInMessagesCache(cache, "parent-42")?.id).toBe(
			"parent-42",
		);
		expect(findParentInMessagesCache(cache, "missing")).toBeNull();
	});
});

describe("Phase 151 unread snapshot replacement", () => {
	it("drops every category key absent from the latest server snapshot", () => {
		const first: UnreadSnapshot = {
			channels: { "ch-a": 3, "ch-b": 5 },
			dms: { "dm-a": 2 },
			mentions: { "ch-a": 1 },
			threads: { "root-a": 4 },
			any_unread: true,
		};
		const second: UnreadSnapshot = {
			channels: { "ch-a": 1 },
			dms: {},
			mentions: {},
			threads: {},
			any_unread: true,
		};
		useMessengerStore.getState().setUnreadSnapshot(first);
		useMessengerStore.getState().setUnreadSnapshot(second);

		expect(useMessengerStore.getState().unreadSnapshot).toEqual(second);
		expect(
			useMessengerStore.getState().unreadSnapshot.channels["ch-b"],
		).toBeUndefined();
	});
});