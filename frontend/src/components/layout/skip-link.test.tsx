import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import "@/i18n";
import { SkipLink } from "./skip-link";

describe("SkipLink", () => {
	it("renders skip link with href to main content", () => {
		render(<SkipLink />);
		const link = screen.getByRole("link");
		expect(link).toHaveAttribute("href", "#main-content");
	});
});
