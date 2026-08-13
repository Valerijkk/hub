import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as adminOrg from "@/api/admin-org";
import { PendingUsersPanel } from "@/components/admin/pending-users-panel";

vi.mock("@/api/admin-org");

describe("PendingUsersPanel (A3W9-1-001)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(adminOrg.listPendingUsers).mockResolvedValue([]);
	});

	it("shows empty copy only after successful empty load", async () => {
		render(<PendingUsersPanel />);
		expect(
			await screen.findByText("Нет заявок на регистрацию"),
		).toBeInTheDocument();
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});

	it("shows QueryErrorAlert on load fail, not empty copy", async () => {
		vi.mocked(adminOrg.listPendingUsers).mockRejectedValue(
			new Error("network"),
		);
		render(<PendingUsersPanel />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить очередь/i,
			);
		});
		expect(
			screen.queryByText("Нет заявок на регистрацию"),
		).not.toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});

	it("retry reloads pending queue", async () => {
		vi.mocked(adminOrg.listPendingUsers)
			.mockRejectedValueOnce(new Error("network"))
			.mockResolvedValueOnce([]);

		render(<PendingUsersPanel />);
		await waitFor(() => {
			expect(screen.getByRole("alert")).toBeInTheDocument();
		});

		fireEvent.click(screen.getByRole("button", { name: /попробовать снова/i }));

		expect(
			await screen.findByText("Нет заявок на регистрацию"),
		).toBeInTheDocument();
		expect(adminOrg.listPendingUsers).toHaveBeenCalledTimes(2);
	});
});
