import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as searchApi from "@/api/search";
import { SavedFilters } from "@/components/search/saved-filters";
import { renderWithProviders } from "../../render";

vi.mock("@/api/search");
vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

import { toastApiError } from "@/lib/format-api-error";

describe("SavedFilters error≠empty (A2W9-4-001 / TASK-093)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows alert + retry instead of silent empty when listSavedFilters fails", async () => {
		vi.mocked(searchApi.listSavedFilters).mockRejectedValue(
			new Error("network"),
		);

		renderWithProviders(<SavedFilters projectId="proj-1" />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить сохранённые фильтры/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(toastApiError).toHaveBeenCalled();
	});

	it("toasts on create failure — silent catch would leave mutation invisible", async () => {
		vi.mocked(searchApi.listSavedFilters).mockResolvedValue({ filters: [] });
		vi.mocked(searchApi.createSavedFilter).mockRejectedValue(
			new Error("create failed"),
		);

		renderWithProviders(
			<SavedFilters
				projectId="proj-1"
				currentFilterConfig={{ status: "open" }}
			/>,
		);

		await waitFor(() => {
			expect(
				screen.getByRole("button", { name: /сохранить фильтр/i }),
			).toBeInTheDocument();
		});

		fireEvent.click(screen.getByRole("button", { name: /сохранить фильтр/i }));
		fireEvent.change(screen.getByPlaceholderText(/название фильтра/i), {
			target: { value: "Мой фильтр" },
		});
		fireEvent.click(screen.getByRole("button", { name: /^сохранить$/i }));

		await waitFor(() => {
			expect(toastApiError).toHaveBeenCalled();
		});
	});
});
