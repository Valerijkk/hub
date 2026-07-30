import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { PipelineStatus } from "@/modules/tracker/pipeline-status";
import { renderWithProviders } from "@testframework/render";

const listWorkflowRuns = vi.fn();

vi.mock("@/api/github", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/github")>();
	return {
		...actual,
		listWorkflowRuns: (...args: unknown[]) => listWorkflowRuns(...args),
	};
});

describe("FE · PipelineStatus sanitizeUrl (FE-TA-001 / TASK-187)", () => {
	beforeEach(() => {
		listWorkflowRuns.mockReset();
	});

	it("mutation: javascript: html_url → нет href (не RED при sanitize)", async () => {
		listWorkflowRuns.mockResolvedValue([
			{
				id: 99,
				name: "CI",
				status: "completed",
				conclusion: "success",
				branch: "main",
				event: "push",
				html_url: "javascript:alert(1)",
				created_at: "2026-07-01T00:00:00Z",
				updated_at: "2026-07-01T00:00:00Z",
				head_sha: "abc123",
				run_number: 99,
			},
		]);

		renderWithProviders(<PipelineStatus projectId="proj-1" />);

		const label = await screen.findByText("CI");
		const row = label.closest("a, div");
		expect(row).toBeTruthy();
		// Unsafe → div без href; raw href={run.html_url} → assert FAIL (javascript:).
		expect(row?.getAttribute("href")).toBeNull();
		expect(row?.tagName.toLowerCase()).not.toBe("a");
		expect(document.querySelector('a[href^="javascript:"]')).toBeNull();
	});

	it("source guard: html_url проходит isSafeUrl/sanitizeUrl (не raw href)", async () => {
		const fs = await import("node:fs");
		const path = await import("node:path");
		const src = fs.readFileSync(
			path.resolve(
				__dirname,
				"../../../../../../frontend/src/modules/tracker/pipeline-status.tsx",
			),
			"utf8",
		);
		expect(src).toMatch(/isSafeUrl\(run\.html_url\)/);
		expect(src).toMatch(/sanitizeUrl\(run\.html_url\)/);
		expect(src).not.toMatch(/href=\{run\.html_url\}/);
	});

	it("source guard: hub tokens — нет text-green/red-600 / bg-gray-* (FE-TA-005)", async () => {
		const fs = await import("node:fs");
		const path = await import("node:path");
		const src = fs.readFileSync(
			path.resolve(
				__dirname,
				"../../../../../../frontend/src/modules/tracker/pipeline-status.tsx",
			),
			"utf8",
		);
		// FE-TA-005: raw Tailwind palette → design-system tokens (var(--green|red|gG|gR|…)).
		expect(src).not.toMatch(/text-(green|red)-600/);
		expect(src).not.toMatch(/bg-gray-/);
		expect(src).not.toMatch(/text-gray-/);
		expect(src).not.toMatch(/text-yellow-/);
		expect(src).not.toMatch(/bg-(green|red|yellow)-/);
		expect(src).toMatch(/text-\[var\(--green\)\]/);
		expect(src).toMatch(/text-\[var\(--red\)\]/);
		expect(src).toMatch(/bg-\[var\(--gG\)\]/);
		expect(src).toMatch(/bg-\[var\(--gR\)\]/);
	});

	it("keeps safe https html_url as href", async () => {
		listWorkflowRuns.mockResolvedValue([
			{
				id: 7,
				name: "Deploy",
				status: "completed",
				conclusion: "failure",
				branch: "develop",
				event: "push",
				html_url: "https://github.com/org/repo/actions/runs/7",
				created_at: "2026-07-01T00:00:00Z",
				run_number: 7,
			},
		]);

		renderWithProviders(<PipelineStatus projectId="proj-1" />);

		await waitFor(() => {
			expect(screen.getByText("Deploy").closest("a")?.getAttribute("href")).toBe(
				"https://github.com/org/repo/actions/runs/7",
			);
		});
	});
});
