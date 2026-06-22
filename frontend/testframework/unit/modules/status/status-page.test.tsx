import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import * as statusApi from "@/api/status";
import { StatusPage } from "@/modules/status/status-page";

vi.mock("@/api/status");

describe("StatusPage", () => {
	it("renders operational status", async () => {
		vi.mocked(statusApi.fetchPublicStatus).mockResolvedValue({
			status: "operational",
			updated_at: new Date().toISOString(),
			components: [
				{
					id: "api",
					name: "Hub API",
					status: "operational",
				},
			],
		});
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<StatusPage />
			</QueryClientProvider>,
		);
		await waitFor(() => {
			expect(
				screen.getByRole("heading", { name: /Все системы работают/i }),
			).toBeInTheDocument();
		});
	});
});
