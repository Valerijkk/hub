import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { GitLabPipelineStatus } from "@/modules/tracker/gitlab-pipeline-status";
import { renderWithProviders } from "@testframework/render";

const listGitLabPipelines = vi.fn();

vi.mock("@/api/gitlab", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/gitlab")>();
	return {
		...actual,
		listGitLabPipelines: (...args: unknown[]) => listGitLabPipelines(...args),
	};
});

describe("FE · GitLabPipelineStatus sanitizeUrl (A3W10-2-001)", () => {
	beforeEach(() => {
		listGitLabPipelines.mockReset();
	});

	it("sanitizes unsafe pipeline web_url to #", async () => {
		listGitLabPipelines.mockResolvedValue([
			{
				id: 42,
				status: "success",
				ref: "main",
				sha: "abcdef0123456789",
				web_url: "javascript:alert(1)",
				created_at: "2026-07-01T00:00:00Z",
			},
		]);

		renderWithProviders(<GitLabPipelineStatus projectId="proj-1" />);

		const label = await screen.findByText("Pipeline #42");
		expect(label.closest("a")?.getAttribute("href")).toBe("#");

		// Mutation: href={pipeline.web_url} → assert href==="#" FAIL.
	});

	it("keeps safe https web_url", async () => {
		listGitLabPipelines.mockResolvedValue([
			{
				id: 7,
				status: "failed",
				ref: "develop",
				sha: "1234567890abcdef",
				web_url: "https://gitlab.com/org/repo/-/pipelines/7",
				created_at: "2026-07-01T00:00:00Z",
			},
		]);

		renderWithProviders(<GitLabPipelineStatus projectId="proj-1" />);

		await waitFor(() => {
			expect(screen.getByText("Pipeline #7").closest("a")?.getAttribute("href")).toBe(
				"https://gitlab.com/org/repo/-/pipelines/7",
			);
		});
	});
});
