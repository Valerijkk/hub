import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as tmsApi from "@/api/tms";
import { CaseTreePanel } from "@/modules/tms/components/case-tree";
import { TestCasesPage } from "@/routes/_authenticated/projects/$projectId/tms/cases.lazy";
import { renderWithProviders } from "../../render";

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
		Link: ({ children, ...props }: { children: React.ReactNode }) => (
			<a {...props}>{children}</a>
		),
	};
});

vi.mock("@/stores/tms-store", () => ({
	useTmsStore: () => ({
		filters: {
			suiteId: undefined,
			status: undefined,
			priority: undefined,
			search: undefined,
		},
		updateFilter: vi.fn(),
		resetFilters: vi.fn(),
	}),
}));

vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

const caseTreeMountSpy = vi.hoisted(() =>
	vi.fn(
		(_props: {
			suites?: unknown;
			selectedSuiteId?: string;
			onSelectSuite: (id: string | undefined) => void;
			total: number;
		}) => null as unknown as React.ReactElement,
	),
);

vi.mock("@/modules/tms/components/case-tree", async () => {
	const actual = await vi.importActual<
		typeof import("@/modules/tms/components/case-tree")
	>("@/modules/tms/components/case-tree");
	return {
		...actual,
		CaseTreePanel: (props: {
			suites?: unknown;
			selectedSuiteId?: string;
			onSelectSuite: (id: string | undefined) => void;
			total: number;
		}) => {
			caseTreeMountSpy(props);
			return actual.CaseTreePanel(
				props as Parameters<typeof actual.CaseTreePanel>[0],
			);
		},
	};
});

const suiteFixture = {
	id: "su1",
	name: "Auth",
	sort_order: 0,
	created_at: "2026-01-01T00:00:00Z",
	updated_at: "2026-01-01T00:00:00Z",
	children: [],
};

const casesLazyPath = resolve(
	__dirname,
	"../../../../../../frontend/src/routes/_authenticated/projects/$projectId/tms/cases.lazy.tsx",
);

describe("CaseTreePanel live export (A2W10-A2-008 / TASK-093)", () => {
	beforeEach(() => {
		caseTreeMountSpy.mockClear();
	});

	it("renders suite tree and all-cases control", () => {
		const onSelect = vi.fn();
		render(
			<CaseTreePanel
				suites={[suiteFixture]}
				selectedSuiteId={undefined}
				onSelectSuite={onSelect}
				total={4}
			/>,
		);

		expect(screen.getByText("Тест-сьюты")).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /все кейсы/i }),
		).toBeInTheDocument();
		expect(screen.getByText("Auth")).toBeInTheDocument();
		expect(screen.getByText("4")).toBeInTheDocument();
	});

	it("cases.lazy wires CaseTreePanel into TestCasesPage", async () => {
		const src = readFileSync(casesLazyPath, "utf8");
		expect(src).toMatch(/<CaseTreePanel[\s\n]/);
		expect(src).toContain(
			'import { CaseTreePanel } from "@/modules/tms/components/case-tree"',
		);
		expect(src).toMatch(
			/suites=\{suites\}[\s\S]*selectedSuiteId=\{filters\.suiteId\}[\s\S]*onSelectSuite=\{handleSuiteSelect\}/,
		);

		vi.spyOn(tmsApi, "listTestCases").mockResolvedValue({
			items: [],
			total: 4,
		});
		vi.spyOn(tmsApi, "listSuites").mockResolvedValue([suiteFixture]);

		renderWithProviders(<TestCasesPage />);

		await waitFor(() => {
			expect(caseTreeMountSpy).toHaveBeenCalled();
		});
		expect(screen.getByText("Тест-сьюты")).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /все кейсы/i }),
		).toBeInTheDocument();
	});
});
