/**
 * TASK-128 / A3W10-5-001…002 — TMS canMutate UI gates (P2-FE-16).
 *
 * Mutation notes:
 * - 001/002: убрать `canMutate` / `canMutate={false}` gate → assert «кнопки null» краснеет
 * - 002 list: убрать `{canMutate ? <Dialog>…` → viewer видит «Новый прогон» → FAIL
 */
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { TestCaseResponse } from "@/api/tms";
import { RunHeader } from "@/modules/tms/components/run-header";
import { RunResultsTable } from "@/modules/tms/components/run-results-table";
import { StepEditor } from "@/modules/tms/components/step-editor";
import {
	createTestQueryClient,
	renderWithProviders,
} from "../../render";

const { useParamsMock } = vi.hoisted(() => ({
	useParamsMock: vi.fn(() => ({
		projectId: "proj-1",
		caseKey: "TC-1",
		runId: "run-1",
	})),
}));

vi.mock("@tanstack/react-router", () => ({
	Link: ({
		children,
		className,
	}: {
		children: React.ReactNode;
		className?: string;
	}) => (
		<a href="#" className={className}>
			{children}
		</a>
	),
	createLazyFileRoute: () => (opts: Record<string, unknown>) => ({
		useParams: useParamsMock,
		...opts,
	}),
}));

vi.mock("@/api/tms", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/tms")>();
	return {
		...actual,
		listTestRuns: vi.fn(),
		createTestRun: vi.fn(),
		getTestCase: vi.fn(),
		listVersions: vi.fn(),
		updateTestCase: vi.fn(),
		linkTestCaseToIssue: vi.fn(),
		unlinkTestCaseFromIssue: vi.fn(),
		rollbackVersion: vi.fn(),
	};
});

vi.mock("@/api/projects", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/projects")>();
	return {
		...actual,
		getProject: vi.fn(),
	};
});

import { getProject } from "@/api/projects";
import { getTestCase, listTestRuns } from "@/api/tms";
import { TestCaseDetailPage } from "@/routes/_authenticated/projects/$projectId/tms/cases_.$caseKey.lazy";
import { TestRunsPage } from "@/routes/_authenticated/projects/$projectId/tms/runs.lazy";

const baseRun = {
	title: "Smoke run",
	status: "created" as const,
	total_count: 1,
	passed_count: 0,
	failed_count: 0,
	broken_count: 0,
	skipped_count: 0,
	results: [{ id: "r1" }],
};

const mockCase: TestCaseResponse = {
	id: "tc-1",
	case_number: 1,
	case_key: "TC-1",
	title: "Login flow",
	case_type: "classic",
	priority: "medium",
	status: "active",
	version: 1,
	created_at: "2026-06-17T10:00:00Z",
	updated_at: "2026-06-17T10:00:00Z",
	steps: [{ order: 1, action: "Open", expected_result: "OK" }],
	linked_issues: [
		{
			id: "iss-1",
			issue_key: "HUB-1",
			title: "Linked issue",
		},
	],
};

describe("A3W10-5 TMS canMutate UI (TASK-128)", () => {
	beforeEach(() => {
		vi.mocked(listTestRuns).mockReset();
		vi.mocked(getProject).mockReset();
		vi.mocked(getTestCase).mockReset();
		useParamsMock.mockReturnValue({
			projectId: "proj-1",
			caseKey: "TC-1",
			runId: "run-1",
		});
	});

	describe("A3W10-5-002 RunHeader", () => {
		it("viewer (canMutate=false): hides Start / Add cases", () => {
			renderWithProviders(
				<RunHeader
					projectId="proj-1"
					run={baseRun}
					canMutate={false}
					onStart={vi.fn()}
					onComplete={vi.fn()}
					onReopen={vi.fn()}
					onAddCases={vi.fn()}
					isStarting={false}
					isCompleting={false}
					isReopening={false}
				/>,
			);

			expect(screen.queryByRole("button", { name: /Запустить/i })).toBeNull();
			expect(
				screen.queryByRole("button", { name: /Добавить кейсы/i }),
			).toBeNull();
		});

		it("member (canMutate=true): shows Start / Add cases", () => {
			renderWithProviders(
				<RunHeader
					projectId="proj-1"
					run={baseRun}
					canMutate={true}
					onStart={vi.fn()}
					onComplete={vi.fn()}
					onReopen={vi.fn()}
					onAddCases={vi.fn()}
					isStarting={false}
					isCompleting={false}
					isReopening={false}
				/>,
			);

			expect(
				screen.getByRole("button", { name: /Запустить/i }),
			).toBeInTheDocument();
			expect(
				screen.getByRole("button", { name: /Добавить кейсы/i }),
			).toBeInTheDocument();
		});

		it("viewer: hides Complete / Reopen", () => {
			const { rerender } = renderWithProviders(
				<RunHeader
					projectId="proj-1"
					run={{ ...baseRun, status: "in_progress" }}
					canMutate={false}
					onStart={vi.fn()}
					onComplete={vi.fn()}
					onReopen={vi.fn()}
					onAddCases={vi.fn()}
					isStarting={false}
					isCompleting={false}
					isReopening={false}
				/>,
			);
			expect(screen.queryByRole("button", { name: /Завершить/i })).toBeNull();

			rerender(
				<RunHeader
					projectId="proj-1"
					run={{ ...baseRun, status: "completed" }}
					canMutate={false}
					onStart={vi.fn()}
					onComplete={vi.fn()}
					onReopen={vi.fn()}
					onAddCases={vi.fn()}
					isStarting={false}
					isCompleting={false}
					isReopening={false}
				/>,
			);
			expect(
				screen.queryByRole("button", { name: /Переоткрыть/i }),
			).toBeNull();
		});
	});

	describe("A3W10-5-002 RunResultsTable", () => {
		it("hides Add when onAddCases omitted", () => {
			renderWithProviders(
				<RunResultsTable
					results={[]}
					runStatus="created"
					executingResultId={null}
					onSelectResult={vi.fn()}
				/>,
			);
			expect(
				screen.queryByRole("button", { name: /Добавить тест-кейсы/i }),
			).toBeNull();
		});

		it("shows Add when onAddCases provided", () => {
			renderWithProviders(
				<RunResultsTable
					results={[]}
					runStatus="created"
					executingResultId={null}
					onSelectResult={vi.fn()}
					onAddCases={vi.fn()}
				/>,
			);
			expect(
				screen.getByRole("button", { name: /Добавить тест-кейсы/i }),
			).toBeInTheDocument();
		});
	});

	describe("A3W10-5-001 StepEditor", () => {
		const emptyCase: TestCaseResponse = {
			...mockCase,
			steps: [],
			linked_issues: undefined,
		};

		it("viewer: no Добавить шаги without onStartEditing", () => {
			renderWithProviders(<StepEditor editing={false} tc={emptyCase} />);
			expect(
				screen.queryByRole("button", { name: /Добавить шаги/i }),
			).toBeNull();
		});

		it("member: shows Добавить шаги", () => {
			const onStart = vi.fn();
			renderWithProviders(
				<StepEditor editing={false} tc={emptyCase} onStartEditing={onStart} />,
			);
			fireEvent.click(screen.getByRole("button", { name: /Добавить шаги/i }));
			expect(onStart).toHaveBeenCalled();
		});
	});

	describe("A3W10-5-002 TestRunsPage", () => {
		it("viewer: hides Новый прогон", async () => {
			vi.mocked(listTestRuns).mockResolvedValue({ items: [], total: 0 });
			const project = {
				id: "proj-1",
				name: "P",
				key: "P",
				my_role: "viewer",
				status: "active",
				member_count: 1,
				version: 1,
				created_at: "2026-01-01T00:00:00Z",
				updated_at: "2026-01-01T00:00:00Z",
			};
			vi.mocked(getProject).mockResolvedValue(project);

			const { createTestQueryClient, renderWithProviders: renderP } =
				await import("../../render");
			const qc = createTestQueryClient();
			qc.setQueryData(["project", "proj-1"], project);
			qc.setQueryData(
				["tms-runs", "proj-1", { offset: 0, limit: 50, status: undefined }],
				{ items: [], total: 0 },
			);

			renderP(<TestRunsPage />, { queryClient: qc });

			await waitFor(() => {
				expect(screen.getByText("Прогоны")).toBeInTheDocument();
			});

			expect(
				screen.queryByRole("button", { name: /Новый прогон/i }),
			).toBeNull();
		});

		it("member: shows Новый прогон", async () => {
			const project = {
				id: "proj-1",
				name: "P",
				key: "P",
				my_role: "developer",
				status: "active",
				member_count: 1,
				version: 1,
				created_at: "2026-01-01T00:00:00Z",
				updated_at: "2026-01-01T00:00:00Z",
			};
			vi.mocked(listTestRuns).mockResolvedValue({ items: [], total: 0 });
			vi.mocked(getProject).mockResolvedValue(project);

			const { createTestQueryClient, renderWithProviders: renderP } =
				await import("../../render");
			const qc = createTestQueryClient();
			qc.setQueryData(["project", "proj-1"], project);
			qc.setQueryData(
				["tms-runs", "proj-1", { offset: 0, limit: 50, status: undefined }],
				{ items: [], total: 0 },
			);

			renderP(<TestRunsPage />, { queryClient: qc });

			await waitFor(() => {
				expect(
					screen.getByRole("button", { name: /Новый прогон/i }),
				).toBeInTheDocument();
			});
		});
	});

	describe("A3W10-5-001 TestCaseDetailPage", () => {
		it("viewer: hides Edit / link / unlink", async () => {
			const project = {
				id: "proj-1",
				name: "P",
				key: "P",
				my_role: "viewer",
				status: "active",
				member_count: 1,
				version: 1,
				created_at: "2026-01-01T00:00:00Z",
				updated_at: "2026-01-01T00:00:00Z",
			};
			vi.mocked(getTestCase).mockResolvedValue(mockCase);
			vi.mocked(getProject).mockResolvedValue(project);

			const { createTestQueryClient, renderWithProviders: renderP } =
				await import("../../render");
			const qc = createTestQueryClient();
			qc.setQueryData(["project", "proj-1"], project);
			qc.setQueryData(["tms-case", "proj-1", "TC-1"], mockCase);

			renderP(<TestCaseDetailPage />, { queryClient: qc });

			await waitFor(() => {
				expect(screen.getByText("Login flow")).toBeInTheDocument();
			});

			expect(
				screen.queryByRole("button", { name: /Редактировать/i }),
			).toBeNull();
			expect(
				screen.queryByRole("button", { name: /Привязать/i }),
			).toBeNull();
			expect(screen.queryByLabelText(/Отвязать задачу/i)).toBeNull();
		});

		it("member: shows Edit; link/unlink on Links tab", async () => {
			const project = {
				id: "proj-1",
				name: "P",
				key: "P",
				my_role: "developer",
				status: "active",
				member_count: 1,
				version: 1,
				created_at: "2026-01-01T00:00:00Z",
				updated_at: "2026-01-01T00:00:00Z",
			};
			vi.mocked(getTestCase).mockResolvedValue(mockCase);
			vi.mocked(getProject).mockResolvedValue(project);

			const { createTestQueryClient, renderWithProviders: renderP } =
				await import("../../render");
			const qc = createTestQueryClient();
			qc.setQueryData(["project", "proj-1"], project);
			qc.setQueryData(["tms-case", "proj-1", "TC-1"], mockCase);

			renderP(<TestCaseDetailPage />, { queryClient: qc });

			await waitFor(() => {
				expect(
					screen.getByRole("button", { name: /Редактировать/i }),
				).toBeInTheDocument();
			});

			fireEvent.click(
				screen.getByRole("button", { name: /Связанные задачи/i }),
			);

			await waitFor(() => {
				expect(
					screen.getByRole("button", { name: /Привязать/i }),
				).toBeInTheDocument();
				expect(screen.getByLabelText(/Отвязать задачу/i)).toBeInTheDocument();
			});
		});
	});
});
