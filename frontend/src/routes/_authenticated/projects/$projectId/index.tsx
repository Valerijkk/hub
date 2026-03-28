import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import {
	archiveProject,
	deleteProject,
	getProject,
	listMembers,
} from "@/api/projects";
import { Dropzone } from "@/components/files/dropzone";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

/**
 * Страница обзора проекта.
 * Отображает основную информацию, участников, файлы и настройки.
 */
function ProjectPage() {
	const { projectId } = Route.useParams();
	const navigate = useNavigate();
	const queryClient = useQueryClient();

	const { data: project, isLoading } = useQuery({
		queryKey: ["project", projectId],
		queryFn: () => getProject(projectId),
	});

	const { data: members } = useQuery({
		queryKey: ["project", projectId, "members"],
		queryFn: () => listMembers(projectId),
	});

	const archiveMutation = useMutation({
		mutationFn: () => archiveProject(projectId),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["projects"] });
			toast.success("Проект архивирован");
			void navigate({ to: "/projects" });
		},
		onError: () => toast.error("Не удалось архивировать проект"),
	});

	const deleteMutation = useMutation({
		mutationFn: () => deleteProject(projectId),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["projects"] });
			toast.success("Проект удален");
			void navigate({ to: "/projects" });
		},
		onError: () => toast.error("Не удалось удалить проект"),
	});

	if (isLoading) {
		return (
			<div className="animate-[eu_450ms_var(--spring)_both] space-y-6 py-4">
				<div className="flex items-center justify-between">
					<div className="flex items-center gap-4">
						<div className="h-12 w-12 rounded-[var(--r3)] animate-shimmer" />
						<div className="space-y-2">
							<div className="h-6 w-48 rounded animate-shimmer" />
							<div className="h-4 w-24 rounded animate-shimmer" />
						</div>
					</div>
					<div className="h-9 w-32 rounded-[var(--r2)] animate-shimmer" />
				</div>
				<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
					{[1, 2, 3, 4].map((i) => (
						<div key={i} className="rounded-[var(--r4)] border border-b1 bg-gradient-to-b from-card to-raised p-5 space-y-2">
							<div className="h-4 w-20 rounded animate-shimmer" />
							<div className="h-8 w-16 rounded animate-shimmer" />
						</div>
					))}
				</div>
			</div>
		);
	}

	if (!project) {
		return (
			<div className="flex items-center justify-center py-12">
				<p className="text-t2">Проект не найден</p>
			</div>
		);
	}

	const isOwnerOrAdmin =
		project.my_role === "owner" || project.my_role === "admin";

	return (
		<div className="animate-[eu_450ms_var(--spring)_both]">
			<div className="mb-6 flex items-center justify-between">
				<div className="flex items-center gap-4">
					<div>
						<h1 className="text-[18px] font-extrabold tracking-[-0.03em] text-t1">{project.name}</h1>
						<p className="text-[13px] text-t3">{project.key}</p>
					</div>
					<Badge
						variant={project.status === "active" ? "default" : "secondary"}
						className={
							project.status === "active"
								? "bg-emerald-500/[0.08] text-emerald-400 border border-emerald-500/[0.15] px-[10px] py-[3px] rounded-full text-[11px] font-[650]"
								: ""
						}
					>
						{project.status === "active" ? "Активный" : project.status}
					</Badge>
				</div>
				<div className="flex gap-2">
					<Button variant="outline" asChild className="border-b1 text-t2 hover:text-t1 hover:bg-[var(--hover)]">
						<Link to="/projects/$projectId/board" params={{ projectId }}>
							Доска
						</Link>
					</Button>
					<Button variant="outline" asChild className="border-b1 text-t2 hover:text-t1 hover:bg-[var(--hover)]">
						<Link to="/projects/$projectId/issues" params={{ projectId }}>
							Задачи
						</Link>
					</Button>
				</div>
			</div>

			<Tabs defaultValue="overview">
				<TabsList>
					<TabsTrigger value="overview">Обзор</TabsTrigger>
					<TabsTrigger value="members">Участники</TabsTrigger>
					<TabsTrigger value="files">Файлы</TabsTrigger>
					{isOwnerOrAdmin && (
						<TabsTrigger value="settings">Настройки</TabsTrigger>
					)}
				</TabsList>

				<TabsContent value="overview" className="mt-4 space-y-4">
					<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5">
						<h3 className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3 mb-3">О проекте</h3>
						<p className="text-[13px] text-t2 leading-[1.5]">
							{project.description || "Описание не указано"}
						</p>
						<Separator className="my-4 bg-b1" />
						<div className="grid grid-cols-2 gap-4 text-[13px]">
							<div>
								<span className="text-t3">Участников:</span>{" "}
								<span className="text-[24px] font-black tracking-[-0.03em] text-t1">{project.member_count}</span>
							</div>
							<div>
								<span className="text-t3">Роль:</span>{" "}
								<span className="text-t1 font-medium">{project.my_role ?? "Не указана"}</span>
							</div>
							<div>
								<span className="text-t3">Создан:</span>{" "}
								<span className="text-t2">{new Date(project.created_at).toLocaleDateString("ru-RU")}</span>
							</div>
							<div>
								<span className="text-t3">Обновлен:</span>{" "}
								<span className="text-t2">{new Date(project.updated_at).toLocaleDateString("ru-RU")}</span>
							</div>
						</div>
					</div>
				</TabsContent>

				<TabsContent value="members" className="mt-4">
					<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5">
						<h3 className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3 mb-3">Участники</h3>
						<p className="text-[12px] text-t3 mb-4">
							{members?.length ?? 0} участников в проекте
						</p>
						<div className="space-y-1">
							{members?.map((member) => (
								<div
									key={member.user_id}
									className="flex items-center justify-between rounded-[var(--r3)] border border-b1 px-4 py-3 hover:bg-[var(--hover)] transition-colors duration-150"
								>
									<div>
										<p className="text-[14px] font-bold text-t1">
											{member.name || member.email}
										</p>
										<p className="text-[13px] text-t2">
											{member.email}
										</p>
									</div>
									<Badge variant="outline" className="border-b1 text-t2">{member.role}</Badge>
								</div>
							))}
							{(!members || members.length === 0) && (
								<p className="text-[13px] text-t3">
									Нет участников
								</p>
							)}
						</div>
					</div>
				</TabsContent>

				<TabsContent value="files" className="mt-4">
					<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5">
						<h3 className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3 mb-3">Файлы</h3>
						<p className="text-[12px] text-t3 mb-4">
							Загрузите файлы через drag and drop или Ctrl+V
						</p>
						<Dropzone
							projectId={projectId}
							entityType="project"
							entityId={projectId}
						/>
					</div>
				</TabsContent>

				{isOwnerOrAdmin && (
					<TabsContent value="settings" className="mt-4 space-y-4">
						<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5">
							<h3 className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3 mb-3">Настройки проекта</h3>
							<p className="text-[12px] text-t3 mb-5">Управление проектом</p>
							<div className="space-y-6">
								<div>
									<h4 className="mb-2 text-[14px] font-bold text-t1">Архивация</h4>
									<p className="mb-2 text-[13px] text-t2">
										Архивированный проект доступен только для чтения.
									</p>
									<Button
										variant="outline"
										onClick={() => archiveMutation.mutate()}
										disabled={
											archiveMutation.isPending || project.status !== "active"
										}
										className="border-b1 text-t2 hover:text-t1 hover:bg-[var(--hover)]"
									>
										Архивировать
									</Button>
								</div>
								<Separator className="bg-b1" />
								<div>
									<h4 className="mb-2 text-[14px] font-bold text-red">
										Удаление
									</h4>
									<p className="mb-2 text-[13px] text-t2">
										Проект будет перемещен в корзину на 30 дней.
									</p>
									<Button
										className="bg-red hover:bg-red/80 text-white"
										onClick={() => {
											if (window.confirm("Удалить проект?")) {
												deleteMutation.mutate();
											}
										}}
										disabled={deleteMutation.isPending}
									>
										Удалить
									</Button>
								</div>
							</div>
						</div>
					</TabsContent>
				)}
			</Tabs>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/projects/$projectId/")({
	component: ProjectPage,
});
