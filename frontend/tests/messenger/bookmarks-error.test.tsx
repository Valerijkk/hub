import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { BookmarksPanel } from "@/components/messenger/bookmarks-panel";
import { queryError } from "../support/query-result";
import { renderWithProviders } from "../support/render";

const refetch = vi.fn();

vi.mock("@tanstack/react-query", async () => {
	const actual = await vi.importActual<typeof import("@tanstack/react-query")>(
		"@tanstack/react-query",
	);
	return {
		...actual,
		useQuery: vi.fn(),
	};
});

vi.mock("@tanstack/react-router", () => ({
	useNavigate: () => vi.fn(),
}));

import { useQuery } from "@tanstack/react-query";

describe("BookmarksPanel error≠empty (W9-A2-009 / TASK-047)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows alert + retry instead of «Нет закладок» when query fails", async () => {
		vi.mocked(useQuery).mockReturnValue(queryError(new Error("network"), refetch));

		renderWithProviders(<BookmarksPanel />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toBeInTheDocument();
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(screen.queryByText(/нет закладок/i)).not.toBeInTheDocument();
	});
});
