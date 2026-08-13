import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as intranetApi from "@/api/intranet";
import { RandomCoffeePanel } from "@/modules/intranet/random-coffee";
import { useAuthStore } from "@/stores/auth-store";

vi.mock("@/api/intranet");

describe("RandomCoffeePanel", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		useAuthStore.setState({
			user: {
				id: "u1",
				email: "dev@hub.local",
				system_role: "member",
			} as never,
		});
		vi.mocked(intranetApi.getCoffeeEnrollment).mockResolvedValue({
			enrolled: false,
			is_active: false,
		});
		vi.mocked(intranetApi.getMyCoffeeMatch).mockResolvedValue({
			unpaired: false,
		});
	});

	it("renders random coffee panel", async () => {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<RandomCoffeePanel />
			</QueryClientProvider>,
		);
		expect(
			screen.getByRole("heading", { name: /Random Coffee/i }),
		).toBeInTheDocument();
		await waitFor(() => {
			expect(intranetApi.getCoffeeEnrollment).toHaveBeenCalled();
		});
	});
});
