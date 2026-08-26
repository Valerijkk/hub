import { describe, expect, it } from "vitest";
import type { Message } from "@/api/messenger";
import { getMessageCacheTargets } from "@/hooks/use-messages";

function message(partial: Partial<Message> = {}): Message {
	return {
		id: "message-1",
		channel_id: "ch-1",
		sender_id: "user-1",
		reply_count: 0,
		created_at: "2026-08-26T00:00:00Z",
		...partial,
	};
}

describe("Phase 151 canonical message cache projection", () => {
	it("projects a root message to the timeline only", () => {
		expect(getMessageCacheTargets(message(), "ch-1", "root-1")).toEqual({
			timeline: true,
			thread: false,
		});
	});

	it("projects an ordinary direct reply to its thread only", () => {
		expect(
			getMessageCacheTargets(
				message({ thread_id: "root-1" }),
				"ch-1",
				"root-1",
			),
		).toEqual({ timeline: false, thread: true });
	});

	it("projects one canonical also-send reply id to timeline and thread", () => {
		const alsoSend = message({
			thread_id: "root-1",
			also_send_to_channel: true,
		});
		const targets = getMessageCacheTargets(alsoSend, "ch-1", "root-1");

		expect(targets).toEqual({ timeline: true, thread: true });
		expect(alsoSend.id).toBe("message-1");
	});

	it("does not project another channel or another root", () => {
		expect(
			getMessageCacheTargets(
				message({ channel_id: "ch-2", thread_id: "root-2" }),
				"ch-1",
				"root-1",
			),
		).toEqual({ timeline: false, thread: false });
	});
});
