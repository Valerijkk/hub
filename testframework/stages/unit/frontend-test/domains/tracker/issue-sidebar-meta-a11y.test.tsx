import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { IssueSidebarMeta } from "@/modules/tracker/issue-sidebar-meta";

vi.mock("@tanstack/react-router", () => ({
	Link: ({
		children,
		...props
	}: {
		children: React.ReactNode;
		to?: string;
		params?: Record<string, string>;
		className?: string;
	}) => <a {...props}>{children}</a>,
}));

const baseIssue = {
	status: { id: "s1", name: "Open" },
	priority: "medium",
	assignee: null,
	reporter: { name: "Alice" },
	labels: [],
	children_count: 0,
	created_at: "2026-01-01T00:00:00Z",
	updated_at: "2026-01-02T00:00:00Z",
	version: 1,
};

describe("IssueSidebarMeta a11y (A2W10-A3-004/008 / TASK-093)", () => {
	it("binds Status/Priority labels via htmlFor", () => {
		render(
			<IssueSidebarMeta
				projectId="p1"
				issue={baseIssue}
				statuses={[{ id: "s1", name: "Open" }]}
				onUpdateField={vi.fn()}
			/>,
		);

		const status = screen.getByLabelText("Статус");
		expect(status.id).toBe("issue-sidebar-status");
		const priority = screen.getByLabelText("Приоритет");
		expect(priority.id).toBe("issue-sidebar-priority");
	});

	it("assignee empty state is not a dead cursor-pointer control", () => {
		const { container } = render(
			<IssueSidebarMeta
				projectId="p1"
				issue={baseIssue}
				onUpdateField={vi.fn()}
			/>,
		);

		expect(screen.getByText("Не назначен")).toBeInTheDocument();
		const assigneeBlock = screen.getByText("Не назначен").closest("div");
		expect(assigneeBlock?.innerHTML ?? "").not.toMatch(/cursor-pointer/);
		expect(container.querySelectorAll("[class*=cursor-pointer]").length).toBeGreaterThan(0);
	});
});
