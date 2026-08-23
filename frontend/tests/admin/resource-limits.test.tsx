import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as adminApi from "@/api/admin";
import { ResourceLimits } from "@/components/admin/resource-limits";

vi.mock("@/api/admin");

describe("ResourceLimits (A3W9-1-002)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(adminApi.getLimits).mockResolvedValue([
			{ key: "limit.file.max_size_mb", value: "25" },
		]);
	});

	it("renders save form after successful load", async () => {
		render(<ResourceLimits />);
		expect(
			await screen.findByLabelText(/максимальный размер файла/i),
		).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /сохранить/i }),
		).toBeInTheDocument();
	});

	it("shows alert+retry on getLimits fail, not save form", async () => {
		vi.mocked(adminApi.getLimits).mockRejectedValue(new Error("network"));
		render(<ResourceLimits />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить лимиты/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: /сохранить/i }),
		).not.toBeInTheDocument();
		expect(
			screen.queryByLabelText(/максимальный размер файла/i),
		).not.toBeInTheDocument();
	});

	it("retry reloads limits form", async () => {
		vi.mocked(adminApi.getLimits)
			.mockRejectedValueOnce(new Error("network"))
			.mockResolvedValueOnce([
				{ key: "limit.file.max_size_mb", value: "25" },
			]);

		render(<ResourceLimits />);
		await waitFor(() => {
			expect(screen.getByRole("alert")).toBeInTheDocument();
		});

		fireEvent.click(screen.getByRole("button", { name: /попробовать снова/i }));

		expect(
			await screen.findByRole("button", { name: /сохранить/i }),
		).toBeInTheDocument();
	});
});
