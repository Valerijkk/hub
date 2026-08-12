import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as projectsApi from "@/api/projects";
import { api } from "@/api/client";
import { WebhooksPage } from "@/modules/settings/webhooks";
import { renderWithProviders } from "../../render";

// TrashPage uses createFileRoute — mock router for isolated render of query path.
vi.mock("@tanstack/react-router", async () => {
	const actual = await vi.importActual<typeof import("@tanstack/react-router")>(
		"@tanstack/react-router",
	);
	return {
		...actual,
		createFileRoute: () => (opts: { component: React.ComponentType }) => ({
			...opts,
			useParams: () => ({}),
		}),
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

describe("FE-P1-TRASH-WEBHOOKS (TASK-032)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("listTrash calls GET /projects/trash — not listProjects", async () => {
		const getSpy = vi.spyOn(api, "get").mockResolvedValue([]);
		await projectsApi.listTrash();
		expect(getSpy).toHaveBeenCalledWith("/projects/trash");
		expect(getSpy).not.toHaveBeenCalledWith("/projects");
	});

	it("mutation guard: listProjects path would miss soft-deleted (API contract)", async () => {
		// Soft-deleted live only on ListTrash; active list excludes them.
		const getSpy = vi
			.spyOn(api, "get")
			.mockImplementation(async (path: string) => {
				if (path === "/projects") {
					return [
						{
							id: "active-1",
							name: "Active",
							key: "ACT",
							status: "active",
							member_count: 1,
							created_at: "2026-01-01T00:00:00Z",
							updated_at: "2026-01-01T00:00:00Z",
						},
					];
				}
				if (path === "/projects/trash") {
					return [
						{
							id: "deleted-1",
							name: "Gone",
							key: "GON",
							status: "deleted",
							member_count: 0,
							created_at: "2026-01-01T00:00:00Z",
							updated_at: "2026-07-01T00:00:00Z",
						},
					];
				}
				throw new Error(`unexpected ${path}`);
			});

		const active = await projectsApi.listProjects();
		const trash = await projectsApi.listTrash();

		expect(active.some((p) => p.status === "deleted")).toBe(false);
		expect(trash.some((p) => p.id === "deleted-1")).toBe(true);
		expect(getSpy).toHaveBeenCalledWith("/projects");
		expect(getSpy).toHaveBeenCalledWith("/projects/trash");
	});

	it("WebhooksPage with empty projectId does not call /projects//webhooks", async () => {
		const getSpy = vi.spyOn(api, "get").mockResolvedValue([]);

		renderWithProviders(<WebhooksPage projectId="" />);

		expect(
			await screen.findByTestId("webhooks-missing-project"),
		).toBeInTheDocument();

		await waitFor(() => {
			expect(getSpy).not.toHaveBeenCalled();
		});
		expect(
			getSpy.mock.calls.some(
				(c) => typeof c[0] === "string" && c[0].includes("/projects//"),
			),
		).toBe(false);
	});

	it("WebhooksPage with projectId lists webhooks for that project", async () => {
		const getSpy = vi.spyOn(api, "get").mockResolvedValue([]);

		renderWithProviders(
			<WebhooksPage projectId="11111111-1111-4111-8111-111111111111" />,
		);

		await waitFor(() => {
			expect(getSpy).toHaveBeenCalledWith(
				"/projects/11111111-1111-4111-8111-111111111111/webhooks",
			);
		});
		expect(screen.queryByTestId("webhooks-missing-project")).toBeNull();
	});
});
