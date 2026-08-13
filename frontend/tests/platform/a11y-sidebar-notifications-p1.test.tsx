import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "@/api/client";
import { SidebarLink } from "@/components/layout/sidebar-link";
import { SidebarUserProfile } from "@/components/layout/sidebar-user";
import { NotificationPanel } from "@/components/notifications/notification-panel";
import { TooltipProvider } from "@/components/ui/tooltip";
import { useAuthStore } from "@/stores/auth-store";
import { renderWithProviders } from "../support/render";

vi.mock("@tanstack/react-router", () => ({
	Link: ({
		children,
		to,
		"aria-label": ariaLabel,
		"aria-current": ariaCurrent,
		...rest
	}: {
		children: React.ReactNode;
		to: string;
		"aria-label"?: string;
		"aria-current"?: string;
	}) => (
		<a href={to} aria-label={ariaLabel} aria-current={ariaCurrent} {...rest}>
			{children}
		</a>
	),
	useMatchRoute: () => () => false,
	useNavigate: () => vi.fn(),
}));

function renderSidebar(ui: React.ReactElement) {
	return render(<TooltipProvider>{ui}</TooltipProvider>);
}

describe("FE-P1-A11Y-CONTROLS (TASK-034)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
		sessionStorage.clear();
		useAuthStore.setState({
			accessToken: "t",
			refreshToken: "r",
			isAuthenticated: true,
			user: {
				id: "u1",
				email: "a@hub.local",
				name: "Alice",
				system_role: "member",
				timezone: "UTC",
				language: "ru",
				onboarding_completed: true,
			},
		});
	});

	it("collapsed SidebarUserProfile is button without nested link", () => {
		const { container } = renderSidebar(
			<SidebarUserProfile collapsed toggle={() => undefined} />,
		);

		const expand = screen.getByRole("button", { name: /развернуть сайдбар/i });
		expect(expand).toBeInTheDocument();
		expect(expand.querySelector("a")).toBeNull();
		expect(container.querySelector("button a")).toBeNull();
	});

	it("expanded SidebarUserProfile is a single Link (no button⊃link)", () => {
		const { container } = renderSidebar(
			<SidebarUserProfile collapsed={false} toggle={() => undefined} />,
		);

		expect(
			screen.getByRole("link", { name: /профиль: alice/i }),
		).toBeInTheDocument();
		expect(container.querySelector("button a")).toBeNull();
		expect(container.querySelector("a button")).toBeNull();
	});

	it("collapsed SidebarLink has accessible name via aria-label", () => {
		renderSidebar(
			<SidebarLink
				to="/dashboard"
				label="Главная"
				icon={<span aria-hidden="true">★</span>}
				collapsed
			/>,
		);

		expect(screen.getByRole("link", { name: "Главная" })).toBeInTheDocument();
	});

	it("notification row has no nested interactive buttons", async () => {
		vi.spyOn(api, "get").mockResolvedValue({
			notifications: [
				{
					id: "n1",
					type: "assigned",
					title: "Assigned you",
					is_read: false,
					is_urgent: false,
					created_at: "2026-07-27T12:00:00Z",
				},
			],
			total: 1,
			page: 1,
			per_page: 20,
		});

		const { container } = renderWithProviders(<NotificationPanel enabled />);

		expect(await screen.findByText("Assigned you")).toBeInTheDocument();

		expect(container.querySelectorAll("button button").length).toBe(0);

		const deleteBtn = screen.getByRole("button", {
			name: /удалить уведомление/i,
			hidden: true,
		});
		expect(deleteBtn.parentElement?.tagName.toLowerCase()).not.toBe("button");
	});
});
