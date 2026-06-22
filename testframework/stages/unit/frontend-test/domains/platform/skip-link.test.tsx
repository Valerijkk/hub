import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { SkipLink } from "@/components/layout/skip-link";
import { renderWithProviders } from "@testframework/render";

describe("SkipLink", () => {
	it("renders skip link with href to main content", () => {
		renderWithProviders(<SkipLink />);
		const link = screen.getByRole("link");
		expect(link).toHaveAttribute("href", "#main-content");
	});
});
