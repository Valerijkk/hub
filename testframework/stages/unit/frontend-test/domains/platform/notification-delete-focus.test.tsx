import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { NotificationDTO } from "@/api/notifications";
import { NotificationPanel } from "@/components/notifications/notification-panel";
import { renderWithProviders } from "../../render";

const sample: NotificationDTO = {
	id: "n1",
	type: "assigned",
	title: "Вам назначена задача",
	body: "HUB-1",
	is_read: false,
	is_urgent: false,
	created_at: new Date().toISOString(),
	entity_type: "issue",
	entity_id: "i1",
};

vi.mock("@/hooks/use-notifications", () => ({
	flattenNotificationPages: () => [sample],
	useNotificationList: () => ({
		data: { pages: [[sample]] },
		isLoading: false,
		isFetchingNextPage: false,
		isError: false,
		hasNextPage: false,
		fetchNextPage: vi.fn(),
		refetch: vi.fn(),
	}),
	useNotificationMutations: () => ({
		markRead: vi.fn(),
		markAllRead: vi.fn(),
		clearAll: vi.fn(),
		deleteNotification: vi.fn(),
	}),
}));

vi.mock("@tanstack/react-router", () => ({
	useNavigate: () => vi.fn(),
}));

describe("NotificationPanel delete focus (A2W9-4-002 / TASK-093)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("delete button uses group-focus-within so keyboard focus reveals control", async () => {
		renderWithProviders(<NotificationPanel />);

		await waitFor(() => {
			expect(
				screen.getByRole("button", { name: /удалить уведомление/i }),
			).toBeInTheDocument();
		});

		const del = screen.getByRole("button", { name: /удалить уведомление/i });
		expect(del.className).toMatch(/group-focus-within:flex/);
		expect(del.className).toMatch(/group-hover:flex/);
	});
});
