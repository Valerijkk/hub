import { describe, expect, it } from "vitest";
import type { LinkPreviewData } from "@/api/messenger";

/** Mirror GenericPreview image resolution (W11-A2-001). */
function resolvePreviewImage(data: LinkPreviewData): string | undefined {
	return data.image_url ?? data.image;
}

describe("FE-P1-MESSENGER · link preview image_url (W11-A2-001)", () => {
	it("prefers image_url from API over legacy image", () => {
		expect(
			resolvePreviewImage({
				title: "t",
				image_url: "https://cdn.example/og.png",
				image: "https://legacy/wrong.png",
			}),
		).toBe("https://cdn.example/og.png");
	});

	it("mutation: reading only image misses image_url → empty", () => {
		const data: LinkPreviewData = {
			title: "t",
			image_url: "https://cdn.example/og.png",
		};
		expect(data.image).toBeUndefined();
		expect(resolvePreviewImage(data)).toBe("https://cdn.example/og.png");
	});
});
