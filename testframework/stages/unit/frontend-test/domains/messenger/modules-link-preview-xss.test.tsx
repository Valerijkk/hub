import { render, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { isSafeUrl, sanitizeUrl } from "@/lib/safe-url";

vi.mock("@/api/messenger", () => ({
	getLinkPreview: vi.fn().mockResolvedValue({
		url: "https://example.com",
		type: "external",
		title: "Example",
	}),
}));

import { LinkPreview } from "@/modules/messenger/link-preview";

/**
 * C-W19-FE-002: modules/messenger LinkPreview — unsafe URL → null / no href.
 * Mutation: убрать isSafeUrl gate + href={url} → javascript: в DOM → FAIL.
 */
describe("modules LinkPreview URL whitelist (C-W19-FE-002)", () => {
	it("does not render preview for javascript: URL", async () => {
		const { container } = render(<LinkPreview url="javascript:alert(1)" />);
		await waitFor(() => {
			expect(container.querySelector("a")).toBeNull();
		});
		expect(container).toBeEmptyDOMElement();
	});

	it("does not render preview for data: URL", async () => {
		const { container } = render(
			<LinkPreview url="data:text/html,<script>alert(1)</script>" />,
		);
		await waitFor(() => {
			expect(container.querySelector("a")).toBeNull();
		});
		expect(container).toBeEmptyDOMElement();
	});

	it("safe https URL uses sanitizeUrl href (not raw unsafe)", async () => {
		const url = "https://example.com/page";
		const { container } = render(<LinkPreview url={url} />);
		await waitFor(() => {
			expect(container.querySelector("a")).not.toBeNull();
		});
		const link = container.querySelector("a");
		expect(link).toHaveAttribute("href", sanitizeUrl(url));
		expect(link?.getAttribute("href")).not.toMatch(/^javascript:/i);
	});

	it("mutation evidence: isSafeUrl rejects javascript:", () => {
		expect(isSafeUrl("javascript:alert(1)")).toBe(false);
		expect(sanitizeUrl("javascript:alert(1)")).toBe("#");
		// Without early isSafeUrl return, sanitizeUrl alone → href="#" still,
		// but fetch/card must not mount for unsafe — empty DOM is the bar.
	});
});
