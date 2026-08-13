import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as githubApi from "@/api/github";
import * as gitlabApi from "@/api/gitlab";
import {
	GitHubSection,
	GitLabSection,
} from "@/routes/_authenticated/projects/$projectId/settings/integrations.lazy";
import { renderWithProviders } from "../support/render";

vi.mock("@tanstack/react-router", async () => {
	const actual = await vi.importActual<typeof import("@tanstack/react-router")>(
		"@tanstack/react-router",
	);
	return {
		...actual,
		createLazyFileRoute: () => (opts: { component: React.ComponentType }) => ({
			...opts,
			useParams: () => ({ projectId: "proj-1" }),
		}),
	};
});

vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

describe("Integrations error≠empty (A3W10-4-006 / TASK-127)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("GitHub: alert + no empty copy + no connect form on list fail", async () => {
		vi.spyOn(githubApi, "listIntegrations").mockRejectedValue(
			new Error("network"),
		);

		renderWithProviders(<GitHubSection projectId="proj-1" />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить github-интеграции/i,
			);
		});
		expect(
			screen.queryByText(/нет подключённых github-репозиториев/i),
		).not.toBeInTheDocument();
		expect(
			screen.queryByRole("heading", { name: /подключить репозиторий/i }),
		).not.toBeInTheDocument();
	});

	it("GitLab: alert + no empty copy + no connect form on list fail", async () => {
		vi.spyOn(gitlabApi, "listGitLabIntegrations").mockRejectedValue(
			new Error("network"),
		);

		renderWithProviders(<GitLabSection projectId="proj-1" />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить gitlab-интеграции/i,
			);
		});
		expect(
			screen.queryByText(/нет подключённых gitlab-репозиториев/i),
		).not.toBeInTheDocument();
		expect(
			screen.queryByRole("heading", { name: /подключить репозиторий/i }),
		).not.toBeInTheDocument();
	});
});
