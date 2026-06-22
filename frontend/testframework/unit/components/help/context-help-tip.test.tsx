import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import "@/i18n";
import { ContextHelpTip } from "@/components/help/context-help-tip";

describe("ContextHelpTip", () => {
	it("renders accessible help trigger for JQL tip", () => {
		render(<ContextHelpTip tipKey="jql" />);
		expect(
			screen.getByRole("button", { name: /context help|подсказка/i }),
		).toBeInTheDocument();
	});

	it("renders WIP tip trigger", () => {
		render(<ContextHelpTip tipKey="wip" />);
		expect(screen.getByRole("button")).toBeInTheDocument();
	});
});
