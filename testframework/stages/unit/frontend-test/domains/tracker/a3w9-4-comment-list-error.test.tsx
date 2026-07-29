import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as trackerApi from "@/api/tracker";
import { CommentList } from "@/components/tracker/comment-list";
import { renderWithProviders } from "../../render";

vi.mock("@/api/tracker", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/tracker")>();
	return {
		...actual,
		listComments: vi.fn(),
		createComment: vi.fn(),
		deleteComment: vi.fn(),
		toggleReaction: vi.fn(),
	};
});

vi.mock("@/components/tracker/plate-editor", () => ({
	PlateEditor: () => <div data-testid="plate-editor" />,
}));

vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

describe("CommentList error≠empty (A3W9-4-003 / TASK-175)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows alert+retry on list fail; composer remains; not silent empty", async () => {
		vi.mocked(trackerApi.listComments).mockRejectedValue(new Error("network"));

		renderWithProviders(
			<CommentList projectId="proj-1" issueKey="HUB-1" />,
		);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить комментарии/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(screen.queryByText(/нет комментариев/i)).not.toBeInTheDocument();
		expect(screen.getByTestId("plate-editor")).toBeInTheDocument();
	});
});
