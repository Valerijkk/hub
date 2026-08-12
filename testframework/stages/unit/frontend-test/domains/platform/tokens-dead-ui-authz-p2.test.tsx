import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { api } from "@/api/client";
import { PresenceIndicator } from "@/components/messenger/presence-indicator";
import { canMutateProject } from "@/lib/permissions";
import { onboardingProfileSchema } from "@/lib/onboarding/profile-schema";
import { IssueHeader } from "@/modules/tracker/issue-header";
import { WebhooksPage } from "@/modules/settings/webhooks";
import { VersionHistory } from "@/modules/tms/components/version-history";
import { renderWithProviders } from "../../render";

vi.mock("@tanstack/react-router", async () => {
	const actual = await vi.importActual<typeof import("@tanstack/react-router")>(
		"@tanstack/react-router",
	);
	return {
		...actual,
		Link: ({
			children,
			to,
			...rest
		}: {
			children: React.ReactNode;
			to: string;
		}) => (
			<a href={to} {...rest}>
				{children}
			</a>
		),
	};
});

describe("TASK-050 P2-FE-05/06/16 tokens + dead UI + authz", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	describe("P2-FE-05 tokens", () => {
		it("PresenceIndicator uses design tokens, not raw Tailwind green-500", () => {
			const { container } = render(<PresenceIndicator status="online" />);
			const el = container.querySelector("[role='status']");
			expect(el?.className).toContain("bg-[var(--green)]");
			expect(el?.className).not.toContain("bg-green-500");
		});
	});

	describe("P2-FE-06 schema / filter contract", () => {
		it("onboarding profile schema requires timezone", () => {
			const withoutTz = onboardingProfileSchema.safeParse({
				name: "Иван",
				language: "ru",
			});
			expect(withoutTz.success).toBe(false);

			const withTz = onboardingProfileSchema.safeParse({
				name: "Иван",
				timezone: "Europe/Moscow",
				language: "ru",
			});
			expect(withTz.success).toBe(true);
		});
	});

	describe("P2-FE-16 authz UI hide", () => {
		it("canMutateProject is false for viewer/guest", () => {
			expect(canMutateProject("viewer")).toBe(false);
			expect(canMutateProject("guest")).toBe(false);
			expect(canMutateProject("developer")).toBe(true);
			expect(canMutateProject("owner")).toBe(true);
			expect(canMutateProject(undefined)).toBe(true);
		});

		it("WebhooksPage hides create for viewer (canMutate=false)", async () => {
			vi.spyOn(api, "get").mockResolvedValue([]);

			renderWithProviders(
				<WebhooksPage
					projectId="11111111-1111-4111-8111-111111111111"
					canMutate={false}
				/>,
			);

			await waitFor(() => {
				expect(screen.queryByText("Загрузка...")).not.toBeInTheDocument();
			});

			expect(screen.queryByTestId("webhook-create-btn")).toBeNull();
			expect(screen.queryByRole("button", { name: "Добавить вебхук" })).toBeNull();
		});

		it("WebhooksPage shows create when canMutate=true", async () => {
			vi.spyOn(api, "get").mockResolvedValue([]);

			renderWithProviders(
				<WebhooksPage
					projectId="11111111-1111-4111-8111-111111111111"
					canMutate={true}
				/>,
			);

			expect(
				await screen.findByTestId("webhook-create-btn"),
			).toBeInTheDocument();
		});

		it("mutation: without canMutate=false guard create button is visible", async () => {
			vi.spyOn(api, "get").mockImplementation(async (path: string) => {
				if (path.includes("/webhooks")) return [];
				if (path.match(/\/projects\/[^/]+$/)) {
					return {
						id: "11111111-1111-4111-8111-111111111111",
						name: "P",
						key: "P",
						status: "active",
						member_count: 1,
						my_role: "developer",
						created_at: "2026-01-01T00:00:00Z",
						updated_at: "2026-01-01T00:00:00Z",
					};
				}
				return [];
			});

			renderWithProviders(
				<WebhooksPage projectId="11111111-1111-4111-8111-111111111111" />,
			);

			expect(
				await screen.findByTestId("webhook-create-btn"),
			).toBeInTheDocument();
		});

		it("IssueHeader hides delete when canDelete=false", () => {
			renderWithProviders(
				<IssueHeader
					projectId="p1"
					issueKey="HUB-1"
					issueType="task"
					onDelete={() => {}}
					isDeleting={false}
					canDelete={false}
				/>,
			);
			expect(screen.queryByTestId("issue-delete-btn")).toBeNull();
			expect(screen.queryByRole("button", { name: "Удалить" })).toBeNull();
		});

		it("VersionHistory hides rollback when canRollback=false", () => {
			renderWithProviders(
				<VersionHistory
					versionsOpen
					onToggle={() => {}}
					versions={[
						{
							id: "v1",
							version_number: 1,
							snapshot: {},
							created_at: "2026-01-01T00:00:00Z",
							created_by: { id: "u1", name: "A" },
						},
					]}
					onRollback={() => {}}
					isRollingBack={false}
					canRollback={false}
				/>,
			);
			expect(screen.queryByTestId("version-rollback-btn")).toBeNull();
			expect(screen.queryByRole("button", { name: "Откатить" })).toBeNull();
		});
	});
});
