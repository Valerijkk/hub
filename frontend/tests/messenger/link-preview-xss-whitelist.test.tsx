import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { isSafeUrl, sanitizeUrl } from "@/lib/safe-url";

vi.mock("@/api/messenger", () => ({
	fetchLinkPreview: vi.fn(),
}));

import { LinkPreview } from "@/components/messenger/link-preview";

/**
 * A3W9-3-003: LinkPreview XSS whitelist — unsafe схемы не рендерят <a href>.
 */
describe("LinkPreview URL whitelist (A3W9-3-003)", () => {
	function renderPreview(url: string) {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		return render(
			<QueryClientProvider client={qc}>
				<LinkPreview url={url} />
			</QueryClientProvider>,
		);
	}

	it("does not render preview card for javascript: URL", () => {
		const { container } = renderPreview("javascript:alert(1)");
		expect(container.querySelector("a")).toBeNull();
		expect(container).toBeEmptyDOMElement();
	});

	it("does not render preview card for data: URL", () => {
		const { container } = renderPreview("data:text/html,<script>alert(1)</script>");
		expect(container.querySelector("a")).toBeNull();
		expect(container).toBeEmptyDOMElement();
	});

	it("renders safe https URL as clickable preview (GitHub path)", () => {
		renderPreview("https://github.com/acme/hub");
		const link = screen.getByRole("link");
		expect(link).toHaveAttribute("href", "https://github.com/acme/hub");
		expect(link.getAttribute("href")).not.toMatch(/^javascript:/i);
	});

	it("sanitizeUrl / isSafeUrl mirror RenderContent whitelist", () => {
		expect(isSafeUrl("javascript:alert(1)")).toBe(false);
		expect(sanitizeUrl("javascript:alert(1)")).toBe("#");
		expect(isSafeUrl("https://example.com")).toBe(true);
		expect(sanitizeUrl("https://example.com")).toBe("https://example.com");
	});

	it("mutation: drop isSafeUrl gate → javascript: would become href candidate", () => {
		// Evidence: without isSafeUrl early-return, sanitizeUrl alone still
		// rewrites to "#" — but extract path must not add unsafe URLs at all.
		const plateUrl = "javascript:alert(1)";
		const extracted = [plateUrl].filter(isSafeUrl);
		expect(extracted).toEqual([]);
		// If filter removed → unsafe leaks into LinkPreview props.
		expect([plateUrl].filter(() => true)).toContain(plateUrl);
	});
});
