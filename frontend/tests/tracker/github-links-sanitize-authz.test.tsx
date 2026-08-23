import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { GitHubLinks } from "@/modules/tracker/github-links";
import { renderWithProviders } from "../support/render";

const listGitHubLinks = vi.fn();
const listIntegrations = vi.fn();

vi.mock("@/api/github", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/github")>();
	return {
		...actual,
		listGitHubLinks: (...args: unknown[]) => listGitHubLinks(...args),
		listIntegrations: (...args: unknown[]) => listIntegrations(...args),
		createBranch: vi.fn(),
	};
});

vi.mock("@/api/projects", () => ({
	getProject: vi.fn(),
}));

describe("FE · GitHubLinks sanitize + canMutate (A3W10-2-001/002)", () => {
	beforeEach(() => {
		listGitHubLinks.mockReset();
		listIntegrations.mockReset();
		listIntegrations.mockResolvedValue([
			{
				id: "ig-1",
				repo_owner: "org",
				repo_name: "repo",
			},
		]);
	});

	it("sanitizes unsafe PR/commit href to #", async () => {
		listGitHubLinks.mockResolvedValue([
			{
				id: "l1",
				link_type: "pull_request",
				url: "javascript:alert(1)",
				title: "Evil PR",
				status: "open",
				external_id: "1",
				author: "x",
			},
			{
				id: "l2",
				link_type: "commit",
				url: "data:text/html,x",
				title: "Evil commit",
				status: "none",
				external_id: "abcdef0123456789",
				author: "y",
			},
		]);

		renderWithProviders(
			<GitHubLinks issueKey="HUB-1" projectId="proj-1" canMutate={true} />,
		);

		const pr = await screen.findByText("Evil PR");
		expect(pr.closest("a")?.getAttribute("href")).toBe("#");

		const commitSha = await screen.findByText("abcdef0");
		expect(commitSha.closest("a")?.getAttribute("href")).toBe("#");

		// Mutation: href={link.url} без sanitize → assert href==="#" FAIL.
	});

	it("hides «+ Создать ветку» for viewer (canMutate=false)", async () => {
		listGitHubLinks.mockResolvedValue([]);

		renderWithProviders(
			<GitHubLinks issueKey="HUB-1" projectId="proj-1" canMutate={false} />,
		);

		await waitFor(() => {
			expect(screen.getByText("GitHub")).toBeInTheDocument();
		});
		expect(
			screen.queryByRole("button", { name: "+ Создать ветку" }),
		).toBeNull();
	});

	it("shows «+ Создать ветку» when canMutate=true and integrations exist", async () => {
		listGitHubLinks.mockResolvedValue([]);

		renderWithProviders(
			<GitHubLinks issueKey="HUB-1" projectId="proj-1" canMutate={true} />,
		);

		expect(
			await screen.findByRole("button", { name: "+ Создать ветку" }),
		).toBeInTheDocument();

		// Mutation: убрать canMutate && → viewer увидит кнопку → assert queryByRole Null FAIL.
	});
});
