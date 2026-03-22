import { useEffect, useState } from "react";
import type { SprintResponse } from "@/api/sprint";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { useSprintStore } from "@/stores/sprint-store";

/** Пропсы компонента SprintBoard. */
interface SprintBoardProps {
	/** ID проекта. */
	projectId: string;
}

/**
 * SprintBoard отображает активный спринт с задачами.
 * Кнопки "Начать спринт" и "Завершить спринт" управляют жизненным циклом.
 */
export function SprintBoard({ projectId }: SprintBoardProps) {
	const {
		sprints,
		activeSprint,
		fetchSprints,
		startSprint,
		completeSprint,
		loading,
	} = useSprintStore();

	const [confirmComplete, setConfirmComplete] = useState(false);
	const [selectedSprint, setSelectedSprint] = useState<SprintResponse | null>(
		null,
	);

	useEffect(() => {
		fetchSprints(projectId);
	}, [projectId, fetchSprints]);

	const handleStartSprint = async (sprintId: string) => {
		await startSprint(projectId, sprintId);
	};

	const handleCompleteSprint = async () => {
		if (!selectedSprint) return;
		await completeSprint(projectId, selectedSprint.id);
		setConfirmComplete(false);
		setSelectedSprint(null);
	};

	const openCompleteDialog = (sprint: SprintResponse) => {
		setSelectedSprint(sprint);
		setConfirmComplete(true);
	};

	if (loading && sprints.length === 0) {
		return (
			<div className="flex items-center justify-center p-8">
				<p className="text-muted-foreground">Загрузка спринтов...</p>
			</div>
		);
	}

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h2 className="text-lg font-semibold">Спринты</h2>
			</div>

			{activeSprint && (
				<Card className="border-primary">
					<CardHeader>
						<div className="flex items-center justify-between">
							<div>
								<CardTitle>{activeSprint.name}</CardTitle>
								<CardDescription>
									{new Date(activeSprint.start_date).toLocaleDateString()} -{" "}
									{new Date(activeSprint.end_date).toLocaleDateString()}
									{activeSprint.goal && ` | ${activeSprint.goal}`}
								</CardDescription>
							</div>
							<div className="flex gap-2">
								<span className="rounded-full bg-green-100 px-3 py-1 text-sm font-medium text-green-800">
									Активный
								</span>
								<Button
									variant="outline"
									size="sm"
									onClick={() => openCompleteDialog(activeSprint)}
								>
									Завершить спринт
								</Button>
							</div>
						</div>
					</CardHeader>
					<CardContent>
						<p className="text-sm text-muted-foreground">
							Задач в спринте: {activeSprint.issue_count}
						</p>
					</CardContent>
				</Card>
			)}

			<div className="space-y-3">
				{sprints
					.filter((s) => s.status === "planning")
					.map((sprint) => (
						<Card key={sprint.id}>
							<CardHeader>
								<div className="flex items-center justify-between">
									<div>
										<CardTitle className="text-base">{sprint.name}</CardTitle>
										<CardDescription>
											{new Date(sprint.start_date).toLocaleDateString()} -{" "}
											{new Date(sprint.end_date).toLocaleDateString()}
											{sprint.goal && ` | ${sprint.goal}`}
										</CardDescription>
									</div>
									<div className="flex gap-2">
										<span className="rounded-full bg-gray-100 px-3 py-1 text-sm font-medium text-gray-600">
											Планирование
										</span>
										<Button
											variant="default"
											size="sm"
											onClick={() => handleStartSprint(sprint.id)}
											disabled={!!activeSprint}
										>
											Начать спринт
										</Button>
									</div>
								</div>
							</CardHeader>
							<CardContent>
								<p className="text-sm text-muted-foreground">
									Задач: {sprint.issue_count}
								</p>
							</CardContent>
						</Card>
					))}

				{sprints
					.filter((s) => s.status === "completed")
					.map((sprint) => (
						<Card key={sprint.id} className="opacity-60">
							<CardHeader>
								<div className="flex items-center justify-between">
									<div>
										<CardTitle className="text-base">{sprint.name}</CardTitle>
										<CardDescription>
											{new Date(sprint.start_date).toLocaleDateString()} -{" "}
											{new Date(sprint.end_date).toLocaleDateString()}
										</CardDescription>
									</div>
									<span className="rounded-full bg-blue-100 px-3 py-1 text-sm font-medium text-blue-800">
										Завершён
									</span>
								</div>
							</CardHeader>
						</Card>
					))}
			</div>

			{/* Диалог подтверждения завершения */}
			<Dialog open={confirmComplete} onOpenChange={setConfirmComplete}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Завершить спринт?</DialogTitle>
						<DialogDescription>
							Незавершённые задачи будут автоматически перенесены в следующий
							спринт.
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<Button variant="outline" onClick={() => setConfirmComplete(false)}>
							Отмена
						</Button>
						<Button onClick={handleCompleteSprint}>Завершить</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
