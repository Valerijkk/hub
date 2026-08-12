import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { NotificationBell } from "@/components/notifications/notification-bell";
import { QueryErrorAlert } from "@/components/query-state-view";

vi.mock("@/hooks/use-notifications", () => ({
	useUnreadNotificationCount: () => ({ data: 3 }),
	useNotificationRealtime: () => undefined,
}));

vi.mock("@/components/notifications/notification-panel", () => ({
	NotificationPanel: () => null,
}));

describe("Shared hub tokens (A3W9-4-004 / TASK-175)", () => {
	it("NotificationBell badge uses var(--red), not bg-red-500 / #D1D2D3", () => {
		const { container } = render(<NotificationBell />);

		expect(
			screen.getByRole("button", { name: /непрочитанных: 3/i }),
		).toBeInTheDocument();
		const html = container.innerHTML;
		expect(html).toMatch(/bg-\[var\(--red\)\]/);
		expect(html).not.toMatch(/bg-red-500/);
		expect(html).not.toMatch(/#D1D2D3/);
	});

	it("QueryErrorAlert full variant uses var(--red), not bg-red-500", () => {
		const { container } = render(
			<QueryErrorAlert message="fail" retry={vi.fn()} />,
		);

		expect(screen.getByRole("alert")).toBeInTheDocument();
		const html = container.innerHTML;
		expect(html).toMatch(/bg-\[var\(--red\)\]/);
		expect(html).toMatch(/text-\[var\(--red\)\]/);
		expect(html).not.toMatch(/bg-red-500/);
		expect(html).not.toMatch(/text-red-500/);
	});
});
