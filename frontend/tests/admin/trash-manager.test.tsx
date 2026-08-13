import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as adminApi from "@/api/admin";
import { TrashManager } from "@/components/admin/trash-manager";

vi.mock("@/api/admin");

describe("TrashManager (A3W9-1-003)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(adminApi.getTrash).mockResolvedValue({ items: [], total: 0 });
		vi.mocked(adminApi.restoreItem).mockResolvedValue(undefined);
	});

	it("shows empty copy only after successful empty load", async () => {
		render(<TrashManager />);
		expect(await screen.findByText("Корзина пуста")).toBeInTheDocument();
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
		expect(screen.getByRole("tablist")).toBeInTheDocument();
	});

	it("shows alert+retry on getTrash fail, not empty copy", async () => {
		vi.mocked(adminApi.getTrash).mockRejectedValue(new Error("network"));
		render(<TrashManager />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить корзину/i,
			);
		});
		expect(screen.queryByText("Корзина пуста")).not.toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		// tabs ARIA (W9-A1-018) preserved
		expect(screen.getByRole("tablist")).toBeInTheDocument();
	});

	it("retry reloads trash table", async () => {
		vi.mocked(adminApi.getTrash)
			.mockRejectedValueOnce(new Error("network"))
			.mockResolvedValueOnce({ items: [], total: 0 });

		render(<TrashManager />);
		await waitFor(() => {
			expect(screen.getByRole("alert")).toBeInTheDocument();
		});

		fireEvent.click(screen.getByRole("button", { name: /попробовать снова/i }));

		expect(await screen.findByText("Корзина пуста")).toBeInTheDocument();
	});
});
