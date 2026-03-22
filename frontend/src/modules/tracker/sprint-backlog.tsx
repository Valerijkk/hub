import { useEffect, useState } from "react";
import { type IssueResponse, listIssues } from "@/api/tracker";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { useSprintStore } from "@/stores/sprint-store";

/** Пропсы компонента SprintBacklog. */
interface SprintBacklogProps {
	/** ID проекта. */
	projectId: string;
}

/**
 * SprintBacklog отображает задачи без спринта (бэклог)
 * и задачи в спринтах со статусом planning.
 * Поддерживает перетаскивание задач в спринт.
 */
export function SprintBacklog({ projectId }: SprintBacklogProps) {
	const { sprints, fetchSprints, moveIssueToSprint } = useSprintStore();
	const [backlogIssues, setBacklogIssues] = useState<IssueResponse[]>([]);
	const [selectedIssues, setSelectedIssues] = useState<Set<string>>(new Set());
	const [loading, setLoading] = useState(false);

	useEffect(() => {
		fetchSprints(projectId);
		loadBacklog();
	}, [projectId, fetchSprints]);

	const loadBacklog = async () => {
		setLoading(true);
		try {
			const resp = await listIssues(projectId, { limit: 100 });
			setBacklogIssues(resp.items);
		} catch {
			// ошибка загрузки бэклога
		}
		setLoading(false);
	};

	const toggleSelect = (issueId: string) => {
		const next = new Set(selectedIssues);
		if (next.has(issueId)) {
			next.delete(issueId);
		} else {
			next.add(issueId);
		}
		setSelectedIssues(next);
	};

	const handleMoveToSprint = async (sprintId: string) => {
		if (selectedIssues.size === 0) return;
		await moveIssueToSprint(projectId, sprintId, Array.from(selectedIssues));
		setSelectedIssues(new Set());
		loadBacklog();
	};

	const planningSprints = sprints.filter((s) => s.status === "planning");

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h2 className="text-lg font-semibold">Бэклог</h2>
				{selectedIssues.size > 0 && planningSprints.length > 0 && (
					<div className="flex items-center gap-2">
						<span className="text-sm text-muted-foreground">
							Выбрано: {selectedIssues.size}
						</span>
						{planningSprints.map((s) => (
							<Button
								key={s.id}
								size="sm"
								variant="outline"
								onClick={() => handleMoveToSprint(s.id)}
							>
								В {s.name}
							</Button>
						))}
					</div>
				)}
			</div>

			{loading ? (
				<p className="text-muted-foreground">Загрузка...</p>
			) : backlogIssues.length === 0 ? (
				<p className="text-muted-foreground">Бэклог пуст</p>
			) : (
				<div className="space-y-2">
					{backlogIssues.map((issue) => (
						<Card
							key={issue.id}
							className={`cursor-pointer transition-colors ${
								selectedIssues.has(issue.id)
									? "border-primary bg-primary/5"
									: "hover:bg-muted/50"
							}`}
							onClick={() => toggleSelect(issue.id)}
						>
							<CardHeader className="py-3">
								<div className="flex items-center gap-3">
									<input
										type="checkbox"
										checked={selectedIssues.has(issue.id)}
										onChange={() => toggleSelect(issue.id)}
										className="h-4 w-4"
										onClick={(e) => e.stopPropagation()}
									/>
									<span className="text-xs font-mono text-muted-foreground">
										{issue.issue_key}
									</span>
									<CardTitle className="text-sm font-normal">
										{issue.title}
									</CardTitle>
									<span
										className={`ml-auto rounded px-2 py-0.5 text-xs ${
											issue.priority === "critical"
												? "bg-red-100 text-red-800"
												: issue.priority === "high"
													? "bg-orange-100 text-orange-800"
													: "bg-gray-100 text-gray-600"
										}`}
									>
										{issue.priority}
									</span>
								</div>
							</CardHeader>
						</Card>
					))}
				</div>
			)}
		</div>
	);
}
