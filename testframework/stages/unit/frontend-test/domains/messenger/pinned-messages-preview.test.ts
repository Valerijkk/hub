import { describe, expect, it } from "vitest";
import type { Message } from "@/api/messenger";
import { resolvePinPreview } from "@/components/messenger/pinned-messages";

/**
 * A3W9-3-002: pins label — preview from cache, not UUID-only.
 */
describe("resolvePinPreview (A3W9-3-002)", () => {
	const messageId = "11111111-1111-4111-8111-111111111111";

	it("uses truncated content_text from message cache", () => {
		const pages: Message[][] = [
			[
				{
					id: messageId,
					channel_id: "ch-1",
					sender_id: "u-1",
					content_text: "Закреплённый текст сообщения",
					reply_count: 0,
					created_at: "2026-07-01T00:00:00Z",
				},
			],
		];
		expect(resolvePinPreview(pages, messageId)).toBe(
			"Закреплённый текст сообщения",
		);
	});

	it("falls back to neutral label instead of UUID", () => {
		const label = resolvePinPreview(undefined, messageId);
		expect(label).toBe("Закреплённое сообщение");
		expect(label).not.toContain(messageId);
		expect(label).not.toMatch(/^ID:/);
	});

	it("mutation: returning ID:uuid would fail neutral-label assert", () => {
		const bad = `ID: ${messageId}`;
		expect(resolvePinPreview([], messageId)).not.toBe(bad);
	});
});
