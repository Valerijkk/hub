/**
 * Общий row-link на задачу дашборда (A3W9-2-002/003).
 * Без project_id Link невалиден — рендерим статичный контейнер.
 */
import { Link } from "@tanstack/react-router";
import type { ReactNode } from "react";
import type { TaskSummary } from "@/api/dashboard";

interface TaskRowLinkProps {
	task: TaskSummary;
	className?: string;
	children: ReactNode;
}

export function TaskRowLink({ task, className, children }: TaskRowLinkProps) {
	if (!task.project_id || !task.key) {
		return <div className={className}>{children}</div>;
	}

	return (
		<Link
			to="/projects/$projectId/issues/$issueKey"
			params={{ projectId: task.project_id, issueKey: task.key }}
			className={className}
		>
			{children}
		</Link>
	);
}
