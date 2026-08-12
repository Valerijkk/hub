import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "@/api/client";
import { NotificationPanel } from "@/components/notifications/notification-panel";
import { renderWithProviders } from "../../render";

vi.mock("@tanstack/react-router", () => ({
	useNavigate: () => vi.fn(),
}));

describe("NotificationPanel Query errors (A10-007)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("shows alert + retry on fetch failure — not empty state and not console.warn", async () => {
		const warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
		vi.spyOn(api, "get").mockRejectedValue(new Error("network down"));

		renderWithProviders(<NotificationPanel enabled />);

		expect(await screen.findByRole("alert")).toHaveTextContent(
			/не удалось загрузить уведомления/i,
		);
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(screen.queryByText(/нет уведомлений/i)).not.toBeInTheDocument();

		await waitFor(() => {
			expect(warnSpy).not.toHaveBeenCalled();
		});

		warnSpy.mockRestore();
	});

	it("renders list when API succeeds", async () => {
		vi.spyOn(api, "get").mockResolvedValue({
			notifications: [
				{
					id: "n1",
					type: "assigned",
					title: "You were assigned",
					is_read: false,
					is_urgent: false,
					created_at: "2026-07-27T12:00:00Z",
				},
			],
			total: 1,
			page: 1,
			per_page: 20,
		});

		renderWithProviders(<NotificationPanel enabled />);

		expect(await screen.findByText("You were assigned")).toBeInTheDocument();
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});

	it("shows empty state when API returns no notifications", async () => {
		vi.spyOn(api, "get").mockResolvedValue({
			notifications: [],
			total: 0,
			page: 1,
			per_page: 20,
		});

		renderWithProviders(<NotificationPanel enabled />);

		expect(await screen.findByText(/нет уведомлений/i)).toBeInTheDocument();
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});

	it("retry button refetches after failure", async () => {
		const getSpy = vi
			.spyOn(api, "get")
			.mockRejectedValueOnce(new Error("network down"))
			.mockResolvedValueOnce({
				notifications: [
					{
						id: "n2",
						type: "mentioned",
						title: "Mentioned you",
						is_read: false,
						is_urgent: false,
						created_at: "2026-07-27T12:00:00Z",
					},
				],
				total: 1,
				page: 1,
				per_page: 20,
			});

		renderWithProviders(<NotificationPanel enabled />);

		expect(await screen.findByRole("alert")).toBeInTheDocument();
		fireEvent.click(screen.getByRole("button", { name: /попробовать снова/i }));

		expect(await screen.findByText("Mentioned you")).toBeInTheDocument();
		expect(getSpy.mock.calls.length).toBeGreaterThanOrEqual(2);
	});
});

describe("notification hooks (A10-007 source)", () => {
	it("uses InfiniteQuery and has no console.warn fetch path", async () => {
		const fs = await import("node:fs");
		const path = await import("node:path");
		const hook = fs.readFileSync(
			path.resolve(
				__dirname,
				"../../../../../../frontend/src/hooks/use-notifications.ts",
			),
			"utf8",
		);
		expect(hook).toContain("useInfiniteQuery");
		expect(hook).toContain("toastApiError");
		expect(hook).not.toMatch(/console\.warn\s*\(/);
		expect(
			fs.existsSync(
				path.resolve(
					__dirname,
					"../../../../../../frontend/src/stores/notification-store.ts",
				),
			),
		).toBe(false);
	});
});
