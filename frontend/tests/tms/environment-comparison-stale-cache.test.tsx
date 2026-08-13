import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type {
	EnvironmentComparisonResponse,
	EnvironmentResponse,
} from "@/api/tms";
import { EnvironmentComparison } from "@/modules/tms/components/environment-comparison";
import { renderWithProviders } from "../support/render";

const ENVS: EnvironmentResponse[] = [
	{
		id: "env-dev",
		name: "Dev",
		label: "dev",
		created_at: "2026-07-01T00:00:00Z",
	},
	{
		id: "env-stg",
		name: "Staging",
		label: "staging",
		created_at: "2026-07-01T00:00:00Z",
	},
	{
		id: "env-prod",
		name: "Prod",
		label: "prod",
		created_at: "2026-07-01T00:00:00Z",
	},
];

const STALE_MARKER = "TC-STALE-MARKER";

const COMPARISON: EnvironmentComparisonResponse = {
	env1: ENVS[0],
	env2: ENVS[1],
	items: [
		{
			case_key: STALE_MARKER,
			case_title: "Stale comparison row",
			status_env1: "passed",
			status_env2: "failed",
		},
	],
};

const listEnvironments = vi.fn();
const compareEnvironments = vi.fn();

vi.mock("@/api/tms", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/tms")>();
	return {
		...actual,
		listEnvironments: (...args: unknown[]) => listEnvironments(...args),
		compareEnvironments: (...args: unknown[]) => compareEnvironments(...args),
	};
});

/** Упрощённый Select: Radix Portal плохо дружит с jsdom для mutation-теста. */
vi.mock("@/components/ui/select", async () => {
	const React = await import("react");

	type SelectCtx = {
		value?: string;
		onValueChange?: (v: string) => void;
	};

	const Ctx = React.createContext<SelectCtx>({});

	return {
		Select: ({
			value,
			onValueChange,
			children,
		}: {
			value?: string;
			onValueChange?: (v: string) => void;
			children: React.ReactNode;
		}) => (
			<Ctx.Provider value={{ value, onValueChange }}>{children}</Ctx.Provider>
		),
		SelectTrigger: ({
			children,
			...props
		}: React.HTMLAttributes<HTMLDivElement> & {
			children?: React.ReactNode;
		}) => (
			<div role="combobox" {...props}>
				{children}
			</div>
		),
		SelectValue: () => null,
		SelectContent: ({ children }: { children: React.ReactNode }) => (
			<div role="listbox">{children}</div>
		),
		SelectItem: ({
			value,
			children,
			disabled,
		}: {
			value: string;
			children: React.ReactNode;
			disabled?: boolean;
		}) => {
			const { onValueChange } = React.useContext(Ctx);
			return (
				<button
					type="button"
					role="option"
					disabled={disabled}
					onClick={() => onValueChange?.(value)}
				>
					{children}
				</button>
			);
		},
	};
});

describe("EnvironmentComparison stale cache guard (W10-A2-003)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		listEnvironments.mockResolvedValue(ENVS);
		compareEnvironments.mockResolvedValue(COMPARISON);
	});

	it("hides comparison table after env change until Compare is clicked again", async () => {
		renderWithProviders(<EnvironmentComparison projectId="proj-1" />);

		await waitFor(() => {
			expect(screen.getByLabelText("Окружение 1")).toBeInTheDocument();
		});

		const env1 = screen.getByLabelText("Окружение 1").parentElement;
		const env2 = screen.getByLabelText("Окружение 2").parentElement;
		expect(env1).toBeTruthy();
		expect(env2).toBeTruthy();

		fireEvent.click(within(env1 as HTMLElement).getByRole("option", { name: /Dev/ }));
		fireEvent.click(
			within(env2 as HTMLElement).getByRole("option", { name: /Staging/ }),
		);

		fireEvent.click(screen.getByRole("button", { name: "Сравнить" }));

		await waitFor(() => {
			expect(screen.getByText(STALE_MARKER)).toBeInTheDocument();
		});
		expect(compareEnvironments).toHaveBeenCalledWith(
			"proj-1",
			"env-dev",
			"env-stg",
		);

		// Мутация: смена env сбрасывает submitted; без guard таблица осталась бы из кэша
		fireEvent.click(
			within(env1 as HTMLElement).getByRole("option", { name: /Prod/ }),
		);

		expect(screen.queryByText(STALE_MARKER)).not.toBeInTheDocument();
		expect(
			screen.getByText(/Выберите два окружения и нажмите "Сравнить"/),
		).toBeInTheDocument();
	});
});
