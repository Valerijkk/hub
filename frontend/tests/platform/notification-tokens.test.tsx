import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { NotificationDTO } from "@/api/notifications";
import { NotificationPanel } from "@/components/notifications/notification-panel";
import { renderWithProviders } from "../support/render";

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

describe("NotificationPanel tokens (A2W9-4-003 / TASK-093)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("unread row uses design tokens, not raw Tailwind blue-500", async () => {
		const { container } = renderWithProviders(<NotificationPanel />);

		await waitFor(() => {
			expect(screen.getByText(/Вам назначена задача/)).toBeInTheDocument();
		});

		const unreadDot = screen.getByRole("img", { name: /непрочитано/i });
		expect(unreadDot.className).toMatch(/bg-\[var\(--blue\)\]/);
		expect(unreadDot.className).not.toMatch(/bg-blue-500/);

		const html = container.innerHTML;
		expect(html).toMatch(/border-\[var\(--blue\)\]/);
		expect(html).not.toMatch(/border-blue-500/);
		expect(html).not.toMatch(/text-blue-400/);
	});
});
