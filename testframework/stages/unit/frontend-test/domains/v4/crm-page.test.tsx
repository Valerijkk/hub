import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
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

const contact = {
	id: "c-1",
	name: "Ada Lovelace",
	email: "ada@hub.local",
	version: 2,
	created_at: "2026-07-01T00:00:00Z",
	updated_at: "2026-07-01T00:00:00Z",
};

function renderCrm() {
	const qc = new QueryClient({
		defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
	});
	return render(
		<QueryClientProvider client={qc}>
			<CrmPage />
		</QueryClientProvider>,
	);
}

describe("CrmPage contact update/delete (A11-009)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(crmApi.listContacts).mockResolvedValue({ contacts: [contact] });
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
		vi.mocked(crmApi.updateContact).mockResolvedValue({
			...contact,
			name: "Ada Updated",
			version: 3,
		});
		vi.mocked(crmApi.deleteContact).mockResolvedValue(undefined);
	});

	it("renders CRM header and pipeline", async () => {
		renderCrm();

		expect(screen.getByRole("heading", { name: /^CRM$/i })).toBeInTheDocument();
		await waitFor(() => {
			expect(crmApi.getPipeline).toHaveBeenCalled();
		});
	});

	it("calls updateContact when saving selected contact", async () => {
		renderCrm();

		await screen.findByText("Ada Lovelace");
		fireEvent.click(screen.getByRole("button", { name: /Ada Lovelace/i }));

		const nameInput = screen.getByLabelText(/^Имя$/i);
		fireEvent.change(nameInput, { target: { value: "Ada Updated" } });
		fireEvent.click(screen.getByRole("button", { name: /^Сохранить$/i }));

		await waitFor(() => {
			expect(crmApi.updateContact).toHaveBeenCalledWith("c-1", {
				name: "Ada Updated",
				email: "ada@hub.local",
				version: 2,
			});
		});
	});

	it("calls deleteContact after confirm", async () => {
		renderCrm();

		await screen.findByText("Ada Lovelace");
		fireEvent.click(screen.getByRole("button", { name: /Ada Lovelace/i }));
		fireEvent.click(
			screen.getByRole("button", { name: /Удалить контакт Ada Lovelace/i }),
		);
		fireEvent.click(screen.getByRole("button", { name: /^Удалить$/i }));

		await waitFor(() => {
			expect(crmApi.deleteContact).toHaveBeenCalledWith("c-1");
		});
	});
});
