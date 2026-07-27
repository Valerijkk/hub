import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "@/api/client";
import { UserManagement } from "@/components/admin/user-management";
import { renderWithProviders } from "../../render";

vi.mock("@/api/admin-org", () => ({
	getOrgInvite: vi.fn().mockResolvedValue({
		invite_token: "tok",
		invite_url: "http://localhost/login?invite=tok",
	}),
}));

vi.mock("@/components/admin/pending-users-panel", () => ({
	PendingUsersPanel: () => <div data-testid="pending-users" />,
}));

describe("UserManagement (A10-006)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("shows error UI instead of empty when /admin/users fails", async () => {
		vi.spyOn(api, "get").mockRejectedValue(new Error("network down"));

		renderWithProviders(<UserManagement />);

		expect(
			await screen.findByRole("alert"),
		).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(screen.queryByText("Пользователи не найдены")).not.toBeInTheDocument();
	});

	it("renders users when API succeeds", async () => {
		vi.spyOn(api, "get").mockResolvedValue([
			{
				id: "u1",
				email: "alice@hub.local",
				name: "Alice",
				display_name: "Alice",
				system_role: "member",
				status: "active",
			},
		]);

		renderWithProviders(<UserManagement />);

		expect(await screen.findByText("Alice")).toBeInTheDocument();
		expect(screen.getByText("alice@hub.local")).toBeInTheDocument();
	});

	it("shows empty state when API returns []", async () => {
		vi.spyOn(api, "get").mockResolvedValue([]);

		renderWithProviders(<UserManagement />);

		await waitFor(() => {
			expect(screen.getByText("Пользователи не найдены")).toBeInTheDocument();
		});
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});
});
