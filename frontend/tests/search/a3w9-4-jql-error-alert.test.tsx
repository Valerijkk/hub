import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as searchApi from "@/api/search";
import { JQLSearchPage } from "@/components/search/jql-search-page";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/search");
vi.mock("@/api/ai", () => ({
	assistJQL: vi.fn(),
}));
vi.mock("@/lib/format-api-error", () => ({
	formatApiError: (err: unknown) =>
		err instanceof Error ? err.message : "error",
	toastApiError: vi.fn(),
}));
vi.mock("@/components/search/jql-editor", () => ({
	JQLEditor: ({
		value,
		onChange,
	}: {
		value: string;
		onChange: (v: string) => void;
		onSubmit?: (q: string) => void;
		className?: string;
	}) => (
		<input
			aria-label="JQL"
			value={value}
			onChange={(e) => onChange(e.target.value)}
		/>
	),
}));
vi.mock("@/components/help/context-help-tip", () => ({
	ContextHelpTip: () => null,
}));

describe("JQLSearchPage error alert (A3W9-4-005 / TASK-175)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(searchApi.listSavedFilters).mockResolvedValue({ filters: [] });
	});

	it("error block has role=alert when searchJQL fails", async () => {
		vi.mocked(searchApi.searchJQL).mockRejectedValue(new Error("bad jql"));

		renderWithProviders(<JQLSearchPage projectId="proj-1" />);

		fireEvent.change(screen.getByLabelText("JQL"), {
			target: { value: "status = Open" },
		});
		fireEvent.click(screen.getByRole("button", { name: /найти/i }));

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(/bad jql/i);
		});
	});
});
