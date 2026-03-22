import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { toast } from "sonner";
import type { ProjectResponse } from "@/api/projects";
import { listProjects, restoreProject } from "@/api/projects";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";

/**
 * Страница корзины — список удалённых/архивных проектов с возможностью восстановления.
 */
function TrashPage() {
	const queryClient = useQueryClient();

	const { data: allProjects, isLoading } = useQuery({
		queryKey: ["projects", "all"],
		queryFn: listProjects,
	});

	const deletedProjects = allProjects?.filter(
		(p: ProjectResponse) => p.status === "deleted" || p.status === "archived",
	);

	const restoreMutation = useMutation({
		mutationFn: restoreProject,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["projects"] });
			toast.success("Проект восстановлен");
		},
		onError: () => toast.error("Не удалось восстановить проект"),
	});

	return (
		<div className="mx-auto max-w-3xl animate-fade-in">
			{/* Заголовок */}
			<div className="mb-8">
				<h1 className="text-2xl font-semibold tracking-tight">Корзина</h1>
				<p className="mt-1 text-sm text-muted-foreground">
					Удаленные проекты хранятся 30 дней. После этого удаляются
					безвозвратно.
				</p>
			</div>

			{/* Загрузка */}
			{isLoading && (
				<div className="space-y-3">
					{[1, 2].map((i) => (
						<div key={i} className="rounded-xl border bg-card p-5">
							<div className="space-y-2">
								<div className="h-5 w-1/3 rounded-md animate-shimmer" />
								<div className="h-4 w-1/5 rounded-md animate-shimmer" />
							</div>
						</div>
					))}
				</div>
			)}

			{/* Пустое состояние */}
			{!isLoading && (!deletedProjects || deletedProjects.length === 0) && (
				<div className="flex flex-col items-center justify-center rounded-xl border border-dashed bg-card/50 py-16 px-6">
					<div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted mb-4">
						<svg
							className="h-6 w-6 text-muted-foreground"
							viewBox="0 0 16 16"
							fill="none"
							stroke="currentColor"
							strokeWidth="1.5"
							strokeLinecap="round"
							strokeLinejoin="round"
						>
							<path d="M2 4h12M5.3 4V2.7a1.3 1.3 0 011.4-1.4h2.6a1.3 1.3 0 011.4 1.4V4M12.7 4v9.3a1.3 1.3 0 01-1.4 1.4H4.7a1.3 1.3 0 01-1.4-1.4V4" />
						</svg>
					</div>
					<h3 className="text-lg font-medium mb-1">Корзина пуста</h3>
					<p className="text-sm text-muted-foreground text-center max-w-sm">
						Здесь будут отображаться удалённые проекты
					</p>
				</div>
			)}

			{/* Список удалённых проектов */}
			{deletedProjects && deletedProjects.length > 0 && (
				<div className="space-y-2">
					{deletedProjects.map((project: ProjectResponse) => (
						<Card key={project.id} className="rounded-xl border shadow-sm">
							<div className="flex items-center justify-between p-4">
								<div className="flex items-start gap-3 min-w-0">
									<div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-muted">
										<svg
											className="h-4 w-4 text-muted-foreground"
											viewBox="0 0 16 16"
											fill="none"
											stroke="currentColor"
											strokeWidth="1.5"
											strokeLinecap="round"
										>
											<rect x="2" y="2" width="5" height="5" rx="1" />
											<rect x="9" y="2" width="5" height="5" rx="1" />
											<rect x="2" y="9" width="5" height="5" rx="1" />
											<rect x="9" y="9" width="5" height="5" rx="1" />
										</svg>
									</div>
									<div className="min-w-0">
										<h3 className="text-sm font-medium truncate">
											{project.name}
										</h3>
										<div className="flex items-center gap-2 mt-0.5">
											<span className="text-xs font-mono text-muted-foreground">
												{project.key}
											</span>
											<span className="text-xs text-muted-foreground">
												&middot; Удален{" "}
												{new Date(project.updated_at).toLocaleDateString(
													"ru-RU",
												)}
											</span>
										</div>
									</div>
								</div>
								<Button
									variant="outline"
									size="sm"
									onClick={() => restoreMutation.mutate(project.id)}
									disabled={restoreMutation.isPending}
									className="shrink-0"
								>
									Восстановить
								</Button>
							</div>
						</Card>
					))}
				</div>
			)}
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/trash")({
	component: TrashPage,
});
