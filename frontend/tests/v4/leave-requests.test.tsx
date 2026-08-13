import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as hrApi from "@/api/hr";
import { LeaveRequestsPage } from "@/modules/hr/leave-requests";
import { useAuthStore } from "@/stores/auth-store";

vi.mock("@/api/hr");

describe("LeaveRequestsPage", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		useAuthStore.setState({
			user: {
				id: "u1",
				email: "dev@hub.local",
				system_role: "member",
			} as never,
		});
		vi.mocked(hrApi.listMyLeave).mockResolvedValue({ items: [] });
	});

	it("renders leave form and my requests", async () => {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<LeaveRequestsPage />
			</QueryClientProvider>,
		);

		expect(
			screen.getByRole("heading", { name: /HR — отпуска/i }),
		).toBeInTheDocument();
		await waitFor(() => {
			expect(hrApi.listMyLeave).toHaveBeenCalled();
		});
	});

	it("shows error + retry instead of empty my-leave (W10-A1-017)", async () => {
		vi.mocked(hrApi.listMyLeave).mockRejectedValue(new Error("network"));
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<LeaveRequestsPage />
			</QueryClientProvider>,
		);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить заявки/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(screen.queryByText(/заявок пока нет/i)).not.toBeInTheDocument();
	});
});
