import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "@/api/client";
import { UserManagement } from "@/components/admin/user-management";
import { renderWithProviders } from "../../render";

const getOrgInvite = vi.fn().mockResolvedValue({
	invite_token: "tok",
	invite_url: "http://localhost/login?invite=tok",
});
const getRegistrationMode = vi.fn().mockResolvedValue({ mode: "invite_only" });
const setRegistrationMode = vi.fn().mockResolvedValue({ mode: "open" });

vi.mock("@/api/admin-org", () => ({
	getOrgInvite: (...args: unknown[]) => getOrgInvite(...args),
	getRegistrationMode: (...args: unknown[]) => getRegistrationMode(...args),
	setRegistrationMode: (...args: unknown[]) => setRegistrationMode(...args),
}));

vi.mock("@/components/admin/pending-users-panel", () => ({
	PendingUsersPanel: () => <div data-testid="pending-users" />,
}));

describe("UserManagement registration mode (A11-013)", () => {
	beforeEach(() => {
		getOrgInvite.mockClear();
		getRegistrationMode.mockClear();
		setRegistrationMode.mockClear();
		getOrgInvite.mockResolvedValue({
			invite_token: "tok",
			invite_url: "http://localhost/login?invite=tok",
		});
		getRegistrationMode.mockResolvedValue({ mode: "invite_only" });
		setRegistrationMode.mockResolvedValue({ mode: "open" });
		vi.spyOn(api, "get").mockResolvedValue([]);
	});

	it("loads registration mode and saves via setRegistrationMode", async () => {
		renderWithProviders(<UserManagement />);

		await waitFor(() => {
			expect(getRegistrationMode).toHaveBeenCalled();
		});

		const select = await screen.findByLabelText("Режим регистрации");
		fireEvent.change(select, { target: { value: "open" } });
		fireEvent.click(screen.getByRole("button", { name: /^Сохранить$/i }));

		await waitFor(() => {
			expect(setRegistrationMode).toHaveBeenCalledWith("open");
		});
	});
});

describe("UserManagement (A10-006)", () => {
	beforeEach(() => {
		getOrgInvite.mockClear();
		getRegistrationMode.mockClear();
		getOrgInvite.mockResolvedValue({
			invite_token: "tok",
			invite_url: "http://localhost/login?invite=tok",
		});
		getRegistrationMode.mockResolvedValue({ mode: "invite_only" });
	});

	it("shows error UI instead of empty when /admin/users fails", async () => {
		vi.spyOn(api, "get").mockRejectedValue(new Error("network down"));

		renderWithProviders(<UserManagement />);

		expect(await screen.findByRole("alert")).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(
			screen.queryByText("Пользователи не найдены"),
		).not.toBeInTheDocument();
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
