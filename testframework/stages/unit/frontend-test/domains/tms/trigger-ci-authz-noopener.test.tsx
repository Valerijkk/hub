import { act, fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { TriggerCIPanel } from "@/modules/tms/trigger-ci-dialog";
import { renderWithProviders } from "@testframework/render";

const listGitLabIntegrations = vi.fn();
const triggerPipeline = vi.fn();
const listTestCases = vi.fn();

vi.mock("@/api/gitlab", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/gitlab")>();
	return {
		...actual,
		listGitLabIntegrations: (...args: unknown[]) =>
			listGitLabIntegrations(...args),
		triggerPipeline: (...args: unknown[]) => triggerPipeline(...args),
	};
});

vi.mock("@/api/tms", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/tms")>();
	return {
		...actual,
		listTestCases: (...args: unknown[]) => listTestCases(...args),
	};
});

vi.mock("@/api/projects", () => ({
	getProject: vi.fn(),
}));

vi.mock("sonner", () => ({
	toast: {
		success: vi.fn(),
		error: vi.fn(),
	},
}));

describe("FE · TriggerCIPanel canMutate + noopener (A3W10-2-002/003)", () => {
	beforeEach(() => {
		listGitLabIntegrations.mockReset();
		triggerPipeline.mockReset();
		listTestCases.mockReset();
		listGitLabIntegrations.mockResolvedValue([
			{ id: "ig-1", project_path: "org/repo" },
		]);
		listTestCases.mockResolvedValue({
			items: [
				{
					id: "c1",
					case_key: "TC-1",
					title: "Login",
					priority: "high",
					status: "active",
				},
			],
			total: 1,
		});
	});

	it("returns null for viewer (canMutate=false) — no trigger CTA", async () => {
		const { container } = renderWithProviders(
			<TriggerCIPanel projectId="proj-1" canMutate={false} />,
		);

		await waitFor(() => {
			expect(container).toBeEmptyDOMElement();
		});
		expect(screen.queryByText("Запуск тестов в CI")).toBeNull();
		expect(
			screen.queryByRole("button", { name: /Запустить пайплайн/ }),
		).toBeNull();
	});

	it("shows trigger UI when canMutate=true", async () => {
		renderWithProviders(
			<TriggerCIPanel projectId="proj-1" canMutate={true} />,
		);

		expect(await screen.findByText("Запуск тестов в CI")).toBeInTheDocument();
		expect(
			await screen.findByRole("button", { name: /Запустить пайплайн/ }),
		).toBeInTheDocument();

		// Mutation: убрать early return !canMutate → viewer увидит CTA → assert Null FAIL.
	});

	it("toast «Открыть» calls window.open with noopener; unsafe URL skips open", async () => {
		const { toast } = await import("sonner");
		const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);

		triggerPipeline.mockResolvedValue({
			pipeline_id: 99,
			ref: "main",
			web_url: "https://gitlab.com/org/repo/-/pipelines/99",
		});

		renderWithProviders(
			<TriggerCIPanel projectId="proj-1" canMutate={true} />,
		);

		const checkbox = await screen.findByRole("checkbox");
		fireEvent.click(checkbox);
		fireEvent.click(
			screen.getByRole("button", { name: /Запустить пайплайн/ }),
		);

		await waitFor(() => {
			expect(toast.success).toHaveBeenCalled();
		});

		const successCall = vi.mocked(toast.success).mock.calls[0];
		const opts = successCall?.[1] as {
			action?: { onClick: () => void };
		};
		expect(opts?.action?.onClick).toBeTypeOf("function");

		openSpy.mockClear();
		act(() => {
			opts?.action?.onClick();
		});
		expect(openSpy).toHaveBeenCalledWith(
			"https://gitlab.com/org/repo/-/pipelines/99",
			"_blank",
			expect.stringContaining("noopener"),
		);

		// Unsafe URL: open не вызывается
		triggerPipeline.mockResolvedValue({
			pipeline_id: 100,
			ref: "main",
			web_url: "javascript:alert(1)",
		});
		vi.mocked(toast.success).mockClear();
		openSpy.mockClear();

		fireEvent.click(screen.getByRole("checkbox"));
		fireEvent.click(
			screen.getByRole("button", { name: /Запустить пайплайн/ }),
		);
		await waitFor(() => {
			expect(toast.success).toHaveBeenCalled();
		});
		const unsafeOpts = vi.mocked(toast.success).mock.calls[0]?.[1] as {
			action?: { onClick: () => void };
		};
		act(() => {
			unsafeOpts?.action?.onClick();
		});
		expect(openSpy).not.toHaveBeenCalled();

		openSpy.mockRestore();

		// Mutation: window.open(raw) без sanitize/noopener → assert stringContaining noopener FAIL
		// или unsafe open вызывается → not.toHaveBeenCalled FAIL.
	});
});
