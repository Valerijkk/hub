import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as intranetApi from "@/api/intranet";
import { OrgDirectory } from "@/modules/intranet/org-directory";

vi.mock("@/api/intranet");
vi.mock("@tanstack/react-router", () => ({
	useNavigate: () => vi.fn(),
}));

const mockDirectory: intranetApi.DirectoryResponse = {
	total: 2,
	departments: [
		{
			department: "Engineering",
			users: [
				{
					id: "u1",
					email: "alice@hub.local",
					display_name: "Alice Eng",
					position: "Backend Dev",
					department: "Engineering",
				},
			],
		},
		{
			department: "HR",
			users: [
				{
					id: "u2",
					email: "bob@hub.local",
					display_name: "Bob HR",
					position: "HR Manager",
					department: "HR",
				},
			],
		},
	],
};

describe("OrgDirectory", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(intranetApi.listDirectory).mockResolvedValue(mockDirectory);
	});

	it("renders departments and users", async () => {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<OrgDirectory />
			</QueryClientProvider>,
		);

		expect(await screen.findByText("Engineering")).toBeInTheDocument();
		expect(screen.getByText("Alice Eng")).toBeInTheDocument();
		expect(screen.getByText("HR")).toBeInTheDocument();
		expect(screen.getByText("Bob HR")).toBeInTheDocument();
	});

	it("filters users by search input", async () => {
		const qc = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		render(
			<QueryClientProvider client={qc}>
				<OrgDirectory />
			</QueryClientProvider>,
		);

		await screen.findByText("Alice Eng");
		const input = screen.getByLabelText(/Поиск сотрудников/i);
		fireEvent.change(input, { target: { value: "bob" } });

		await waitFor(() => {
			expect(screen.queryByText("Alice Eng")).not.toBeInTheDocument();
			expect(screen.getByText("Bob HR")).toBeInTheDocument();
		});
	});
});
