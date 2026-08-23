import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as adminApi from "@/api/admin";
import { OperationMetricsPanel } from "@/components/admin/operation-metrics";

const mockGetOperationMetrics = vi.spyOn(adminApi, "getOperationMetrics");

describe("OperationMetricsPanel", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		mockGetOperationMetrics.mockResolvedValue([]);
	});

	it("loads metrics on mount", async () => {
		render(<OperationMetricsPanel />);

		await waitFor(() => {
			expect(mockGetOperationMetrics).toHaveBeenCalledTimes(1);
		});

		expect(screen.getByText("Метрики операций")).toBeInTheDocument();
	});

	it("renders metric rows from API", async () => {
		mockGetOperationMetrics.mockResolvedValue([
			{
				operation: "auth.Login",
				count: 42,
				error_count: 2,
				total_ms: 840,
				max_ms: 120,
				avg_ms: 20,
				last_seen_utc: "2026-06-20T12:00:00Z",
			},
		]);

		render(<OperationMetricsPanel />);

		expect(await screen.findByText("auth.Login")).toBeInTheDocument();
		expect(screen.getByText("42")).toBeInTheDocument();
		expect(screen.getByText("2")).toBeInTheDocument();
	});

	it("shows empty state when no metrics", async () => {
		render(<OperationMetricsPanel />);

		expect(await screen.findByText(/Пока нет данных/)).toBeInTheDocument();
	});

	it("refresh button reloads metrics", async () => {
		render(<OperationMetricsPanel />);

		await waitFor(() => {
			expect(mockGetOperationMetrics).toHaveBeenCalledTimes(1);
		});

		fireEvent.click(screen.getByRole("button", { name: "Обновить" }));

		await waitFor(() => {
			expect(mockGetOperationMetrics).toHaveBeenCalledTimes(2);
		});
	});
});
