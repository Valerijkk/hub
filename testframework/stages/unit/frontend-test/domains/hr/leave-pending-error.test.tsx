import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as hrApi from "@/api/hr";
import { LeaveRequestsPage } from "@/modules/hr/leave-requests";
import { useAuthStore } from "@/stores/auth-store";

vi.mock("@/api/hr");

/**
 * C-W19-FE-005 / TASK-182: admin pending leave fail ≠ скрытая очередь.
 * Mutation: убрать pendingError из gate секции → alert/heading пропадут (RED).
 */
describe("LeaveRequestsPage pending error≠empty (C-W19-FE-005 / TASK-182)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		useAuthStore.setState({
			user: {
				id: "u1",
				email: "admin@hub.local",
				system_role: "admin",
			} as never,
		});
		vi.mocked(hrApi.listMyLeave).mockResolvedValue({ items: [] });
		vi.mocked(hrApi.listPendingLeave).mockRejectedValue(new Error("network"));
	});

	it("shows approval queue alert when listPendingLeave fails, not silent empty", async () => {
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
				/не удалось загрузить очередь согласования/i,
			);
		});
		expect(
			screen.getByRole("heading", { name: /очередь согласования/i }),
		).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});

	it("hides approval queue on success+empty (no false alert)", async () => {
		vi.mocked(hrApi.listPendingLeave).mockResolvedValue({ items: [] });
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<LeaveRequestsPage />
			</QueryClientProvider>,
		);

		await waitFor(() => {
			expect(screen.getByText(/заявок пока нет/i)).toBeInTheDocument();
		});
		expect(
			screen.queryByRole("heading", { name: /очередь согласования/i }),
		).not.toBeInTheDocument();
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});
});
