import { render, screen } from "@testing-library/react";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it, vi } from "vitest";
import type { PersonalData } from "@/api/dashboard";
import { DeadlinesWidget } from "@/components/dashboard/widgets/deadlines";
import { MyTasksWidget } from "@/components/dashboard/widgets/my-tasks";
import { SystemHealthWidget } from "@/components/dashboard/widgets/system-health";
import { UnreadMessagesWidget } from "@/components/dashboard/widgets/unread-messages";

vi.mock("@tanstack/react-router", () => ({
	Link: ({
		children,
		to,
		params,
		className,
	}: {
		children: React.ReactNode;
		to?: string;
		params?: Record<string, string>;
		className?: string;
	}) => (
		<a
			href={`${to}?projectId=${params?.projectId ?? ""}&issueKey=${params?.issueKey ?? ""}`}
			data-to={to}
			data-project-id={params?.projectId}
			data-issue-key={params?.issueKey}
			className={className}
		>
			{children}
		</a>
	),
}));

const PROJECT_ID = "proj-aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee";

const personalBase: PersonalData = {
	my_tasks: [],
	unread_count: 0,
	upcoming_calls: [],
	recent_runs: [],
	deadlines: [],
	time_today: 0,
};

function widgetsSrc(file: string): string {
	return readFileSync(
		resolve(process.cwd(), `src/components/dashboard/widgets/${file}`),
		"utf8",
	);
}

describe("Dashboard widgets A3W9-2 (TASK-173)", () => {
	it("deadlines row links to issue route with project_id + key (A3W9-2-002)", () => {
		const data: PersonalData = {
			...personalBase,
			deadlines: [
				{
					id: "iss-1",
					key: "HUB-42",
					title: "Ship deadlines link",
					status: "In Progress",
					priority: 3,
					project_id: PROJECT_ID,
					due_date: new Date(Date.now() + 2 * 86400000).toISOString(),
				},
			],
		};

		render(<DeadlinesWidget data={data} />);

		const link = screen.getByRole("link");
		expect(link).toHaveAttribute(
			"data-to",
			"/projects/$projectId/issues/$issueKey",
		);
		expect(link).toHaveAttribute("data-project-id", PROJECT_ID);
		expect(link).toHaveAttribute("data-issue-key", "HUB-42");
		expect(screen.getByText("Ship deadlines link")).toBeInTheDocument();
	});

	it("my-tasks row uses the same Link pattern (A3W9-2-003)", () => {
		const data: PersonalData = {
			...personalBase,
			my_tasks: [
				{
					id: "iss-2",
					key: "HUB-7",
					title: "Fix my-tasks link",
					status: "Open",
					priority: 4,
					project_id: PROJECT_ID,
				},
			],
		};

		render(<MyTasksWidget data={data} />);

		const link = screen.getByRole("link");
		expect(link).toHaveAttribute(
			"data-to",
			"/projects/$projectId/issues/$issueKey",
		);
		expect(link).toHaveAttribute("data-project-id", PROJECT_ID);
		expect(link).toHaveAttribute("data-issue-key", "HUB-7");
	});

	it("deadlines/my-tasks source has no raw Tailwind priority colors (A3W9-2-004)", () => {
		const forbidden =
			/\b(bg|text)-(red|amber|orange|yellow|blue|slate)-[0-9]{3}\b/;
		expect(widgetsSrc("deadlines.tsx")).not.toMatch(forbidden);
		expect(widgetsSrc("my-tasks.tsx")).not.toMatch(forbidden);
		expect(widgetsSrc("deadlines.tsx")).toContain("var(--red)");
		expect(widgetsSrc("my-tasks.tsx")).toContain("var(--amber)");
	});

	it("system-health/unread use semantic tokens, no *-500 or inline rgba glow (A3W9-2-005)", () => {
		const healthSrc = widgetsSrc("system-health.tsx");
		const unreadSrc = widgetsSrc("unread-messages.tsx");
		const raw500 = /\b(bg|text)-(green|red|yellow|amber)-500\b/;
		expect(healthSrc).not.toMatch(raw500);
		expect(unreadSrc).not.toMatch(raw500);
		expect(healthSrc).not.toMatch(/rgba\(/);
		expect(healthSrc).not.toMatch(/style=\{\{\s*boxShadow/);
		expect(healthSrc).toContain("var(--green)");
		expect(unreadSrc).toContain("var(--amber)");

		render(
			<SystemHealthWidget
				data={{
					total_users: 1,
					active_users: 1,
					total_projects: 1,
					total_issues: 1,
					dau: 1,
					wau: 1,
					mau: 1,
					storage_used: 0,
					service_health: { postgres: "healthy", redis: "degraded" },
					top_projects: [],
				}}
			/>,
		);
		expect(screen.getByText("Работает")).toBeInTheDocument();
		expect(screen.getByText("Деградация")).toBeInTheDocument();

		render(<UnreadMessagesWidget data={{ ...personalBase, unread_count: 3 }} />);
		expect(screen.getByText("непрочитанных сообщений")).toBeInTheDocument();
		expect(screen.getAllByText("3").length).toBeGreaterThanOrEqual(1);
	});
});
