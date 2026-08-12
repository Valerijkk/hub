import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as projectsApi from "@/api/projects";
import { api } from "@/api/client";
import { TrashPage } from "@/routes/_authenticated/trash";
import { renderWithProviders } from "../../render";

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

const toastSuccess = vi.fn();
const toastError = vi.fn();

vi.mock("sonner", () => ({
	toast: {
		success: (...args: unknown[]) => toastSuccess(...args),
		error: (...args: unknown[]) => toastError(...args),
		dismiss: vi.fn(),
	},
}));

const deletedProject = {
	id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
	name: "Gone Project",
	key: "GON",
	status: "deleted",
	member_count: 0,
	created_at: "2026-01-01T00:00:00Z",
	updated_at: "2026-07-20T00:00:00Z",
};

describe("TASK-083 trash hard-delete API", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
		toastSuccess.mockClear();
		toastError.mockClear();
	});

	it("permanentDeleteProject calls DELETE /projects/:id/permanent", async () => {
		const delSpy = vi.spyOn(api, "delete").mockResolvedValue(undefined);
		await projectsApi.permanentDeleteProject(deletedProject.id);
		expect(delSpy).toHaveBeenCalledWith(
			`/projects/${deletedProject.id}/permanent`,
		);
	});

	it("mutation fail: stub API reject → no success toast, item stays", async () => {
		vi.spyOn(projectsApi, "listTrash").mockResolvedValue([deletedProject]);
		vi.spyOn(projectsApi, "permanentDeleteProject").mockRejectedValue({
			code: 500,
			message: "failed to permanently delete project",
		});

		renderWithProviders(<TrashPage />);

		expect(await screen.findByText("Gone Project")).toBeInTheDocument();

		fireEvent.click(screen.getByRole("button", { name: "Удалить" }));
		expect(
			await screen.findByRole("heading", { name: "Удалить навсегда?" }),
		).toBeInTheDocument();

		fireEvent.click(screen.getByTestId("trash-hard-delete-confirm"));

		await waitFor(() => {
			expect(toastError).toHaveBeenCalled();
		});
		expect(toastSuccess).not.toHaveBeenCalled();
		expect(screen.getByText("Gone Project")).toBeInTheDocument();
	});

	it("mutation success: API ok → success toast + list refetch path", async () => {
		const listSpy = vi
			.spyOn(projectsApi, "listTrash")
			.mockResolvedValue([deletedProject]);
		vi.spyOn(projectsApi, "permanentDeleteProject").mockResolvedValue(
			undefined,
		);

		renderWithProviders(<TrashPage />);

		expect(await screen.findByText("Gone Project")).toBeInTheDocument();
		fireEvent.click(screen.getByRole("button", { name: "Удалить" }));
		fireEvent.click(screen.getByTestId("trash-hard-delete-confirm"));

		await waitFor(() => {
			expect(toastSuccess).toHaveBeenCalledWith("Элемент удален навсегда");
		});
		expect(toastError).not.toHaveBeenCalled();
		await waitFor(() => {
			expect(listSpy.mock.calls.length).toBeGreaterThanOrEqual(2);
		});
	});
});
