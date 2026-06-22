import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as crmApi from "@/api/crm";
import { CrmPage } from "@/modules/crm/crm-page";

vi.mock("@/api/crm");
vi.mock("@/api/projects", () => ({
	listProjects: vi.fn().mockResolvedValue([]),
}));
vi.mock("@tanstack/react-router", () => ({
	useNavigate: () => vi.fn(),
}));

describe("CrmPage", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(crmApi.listContacts).mockResolvedValue({ contacts: [] });
		vi.mocked(crmApi.getPipeline).mockResolvedValue({
			stages: [
				{ stage: "lead", deals: [] },
				{ stage: "qualified", deals: [] },
				{ stage: "proposal", deals: [] },
				{ stage: "negotiation", deals: [] },
				{ stage: "won", deals: [] },
				{ stage: "lost", deals: [] },
			],
		});
	});

	it("renders CRM header and pipeline", async () => {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<CrmPage />
			</QueryClientProvider>,
		);

		expect(screen.getByRole("heading", { name: /^CRM$/i })).toBeInTheDocument();
		await waitFor(() => {
			expect(crmApi.getPipeline).toHaveBeenCalled();
		});
	});
});
