import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { TestCaseResponse } from "@/api/tms";
import { CaseTableRow } from "./case-table-row";

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
}));

const mockCase: TestCaseResponse = {
	id: "tc-1",
	case_number: 1,
	case_key: "TC-1",
	title: "Login flow test",
	case_type: "classic",
	priority: "high",
	status: "active",
	version: 1,
	created_at: "2026-06-17T10:00:00Z",
	updated_at: "2026-06-17T10:00:00Z",
};

function renderRow(
	props: Partial<{
		tc: TestCaseResponse;
		isSelected: boolean;
		onToggleSelect: (id: string) => void;
	}> = {},
) {
	const onToggleSelect = props.onToggleSelect ?? vi.fn();
	return render(
		<table>
			<tbody>
				<CaseTableRow
					tc={props.tc ?? mockCase}
					projectId="proj-1"
					index={0}
					isSelected={props.isSelected ?? false}
					onToggleSelect={onToggleSelect}
				/>
			</tbody>
		</table>,
	);
}

describe("CaseTableRow", () => {
	it("renders case key and title", () => {
		renderRow();

		expect(screen.getByText("TC-1")).toBeInTheDocument();
		expect(screen.getByText("Login flow test")).toBeInTheDocument();
	});

	it("renders priority and status badges", () => {
		renderRow();

		expect(screen.getByText("High")).toBeInTheDocument();
		expect(screen.getByText("Активный")).toBeInTheDocument();
	});

	it("calls onToggleSelect when checkbox clicked", () => {
		const onToggleSelect = vi.fn();
		renderRow({ onToggleSelect });

		fireEvent.click(screen.getByRole("checkbox"));

		expect(onToggleSelect).toHaveBeenCalledWith("tc-1");
	});

	it("shows autotest icon when case has full_name", () => {
		const tc = { ...mockCase, full_name: "com.hub.LoginTest" };
		const { container } = renderRow({ tc });

		expect(
			container.querySelector('[title="Автоматизирован"]'),
		).toBeInTheDocument();
	});

	it("shows manual icon when case has no autotest metadata", () => {
		const { container } = renderRow();

		expect(container.querySelector('[title="Ручной"]')).toBeInTheDocument();
	});
});
