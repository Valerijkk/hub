import { describe, expect, it } from "vitest";
import type { Message } from "@/api/messenger";
import { findParentInMessagesCache } from "@/components/messenger/thread-panel";
import { useMessengerStore } from "@/stores/messenger-store";

function msg(id: string): Message {
	return {
		id,
		channel_id: "ch-1",
		sender_id: "u-1",
		content_text: `msg ${id}`,
		reply_count: 0,
		created_at: "2026-07-28T12:00:00Z",
	};
}

describe("FE-P1-MESSENGER · thread parent cache (W9-A3-001)", () => {
	it("finds parent in infinite pages, not only first message", () => {
		const cache = {
			pages: [[msg("newest")], [msg("parent-42"), msg("older")]],
		};
		expect(findParentInMessagesCache(cache, "parent-42")?.id).toBe(
			"parent-42",
		);
		expect(findParentInMessagesCache(cache, "missing")).toBeNull();
	});
});

describe("FE-P1-MESSENGER · unread map replace (W11-A4-001)", () => {
	it("setUnreadCounts drops channels absent from poll (MarkRead HDel)", () => {
		useMessengerStore.setState({
			unreadCounts: { "ch-a": 3, "ch-b": 5 },
		});
		useMessengerStore.getState().setUnreadCounts({ "ch-a": 1 });
		expect(useMessengerStore.getState().unreadCounts).toEqual({ "ch-a": 1 });
		expect(useMessengerStore.getState().unreadCounts["ch-b"]).toBeUndefined();
	});

	it("updateUnread alone would leave sticky badge (mutation criterion)", () => {
		useMessengerStore.setState({
			unreadCounts: { "ch-a": 3, "ch-b": 5 },
		});
		useMessengerStore.getState().updateUnread("ch-a", 1);
		expect(useMessengerStore.getState().unreadCounts["ch-b"]).toBe(5);
	});
});
