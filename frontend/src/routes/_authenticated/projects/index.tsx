import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { listProjects } from "@/api/projects";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

/**
 * Страница списка проектов — сетка карточек, пустое состояние,
 * скелетон-загрузка, CTA для создания нового проекта.
 */
function ProjectsPage() {
	const { data: projects, isLoading } = useQuery({
		queryKey: ["projects"],
		queryFn: listProjects,
	});

	return (
		<div className="mx-auto max-w-5xl animate-fade-in">
			{/* Заголовок страницы — градиентный */}
			<div className="mb-8 flex items-center justify-between">
				<div>
					<h1 className="page-title">Проекты</h1>
					<p className="mt-2 text-sm text-muted-foreground leading-relaxed">
						Управляйте проектами вашей команды
					</p>
				</div>
				<Button
					asChild
					className="bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white shadow-md shadow-blue-500/25 hover-pulse"
				>
					<Link to="/projects/new">
						<svg
							className="mr-2 h-4 w-4"
							viewBox="0 0 16 16"
							fill="none"
							stroke="currentColor"
							strokeWidth="2"
							strokeLinecap="round"
						>
							<path d="M8 2v12M2 8h12" />
						</svg>
						Создать проект
					</Link>
				</Button>
			</div>

			{/* Скелетон-загрузка */}
			{isLoading && (
				<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
					{[1, 2, 3].map((i) => (
						<div key={i} className="rounded-xl border bg-card p-5">
							<div className="space-y-3">
								<div className="h-5 w-2/3 rounded-md animate-shimmer" />
								<div className="h-4 w-1/4 rounded-md animate-shimmer" />
								<div className="h-4 w-full rounded-md animate-shimmer" />
								<div className="h-3 w-1/3 rounded-md animate-shimmer" />
							</div>
						</div>
					))}
				</div>
			)}

			{/* Пустое состояние */}
			{!isLoading && projects?.length === 0 && (
				<div className="flex flex-col items-center justify-center rounded-xl border border-dashed bg-card/50 py-20 px-6 animate-fade-in">
					{/* Illustration */}
					<div className="relative mb-6">
						<div className="flex h-20 w-20 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-500/10 to-purple-500/10">
							<svg
								className="h-10 w-10 text-primary/60"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								strokeWidth="1.5"
								strokeLinecap="round"
								strokeLinejoin="round"
							>
								<rect x="3" y="3" width="7" height="7" rx="2" />
								<rect x="14" y="3" width="7" height="7" rx="2" />
								<rect x="3" y="14" width="7" height="7" rx="2" />
								<rect x="14" y="14" width="7" height="7" rx="2" />
							</svg>
						</div>
						<div className="absolute -top-1 -right-1 h-5 w-5 rounded-full bg-blue-500/20 animate-pulse" />
						<div
							className="absolute -bottom-1 -left-1 h-3 w-3 rounded-full bg-purple-500/20 animate-pulse"
							style={{ animationDelay: "0.5s" }}
						/>
					</div>
					<h3 className="text-xl font-semibold mb-2">Нет проектов</h3>
					<p className="mb-8 text-sm text-muted-foreground text-center max-w-sm leading-relaxed">
						Создайте первый проект для вашей команды, чтобы начать работу с
						задачами, тестами и общением.
					</p>
					<Button
						asChild
						className="bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white shadow-md shadow-blue-500/25 hover-pulse"
					>
						<Link to="/projects/new">
							<svg
								className="mr-2 h-4 w-4"
								viewBox="0 0 16 16"
								fill="none"
								stroke="currentColor"
								strokeWidth="2"
								strokeLinecap="round"
							>
								<path d="M8 2v12M2 8h12" />
							</svg>
							Создать первый проект
						</Link>
					</Button>
				</div>
			)}

			{/* Сетка проектов */}
			{projects && projects.length > 0 && (
				<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
					{projects.map((project, index) => (
						<Link
							key={project.id}
							to="/projects/$projectId"
							params={{ projectId: project.id }}
							className="group block animate-fade-in-up"
							style={{ animationDelay: `${index * 50}ms` }}
						>
							{/* Gradient border wrapper */}
							<div className="rounded-xl p-[1px] bg-border transition-all duration-300 group-hover:bg-gradient-to-br group-hover:from-blue-500/50 group-hover:via-purple-500/30 group-hover:to-blue-500/50">
								<Card className="h-full rounded-[11px] border-0 bg-card transition-all duration-200 hover:shadow-lg hover:-translate-y-1">
									<CardHeader className="pb-3 pt-5">
										<div className="flex items-start justify-between gap-2">
											<div className="flex items-center gap-3 min-w-0 flex-1">
												{/* Иконка проекта — ключ проекта */}
												<div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-primary/15 to-primary/5 text-primary font-bold text-xs tracking-tight ring-1 ring-primary/10">
													{project.key.substring(0, 3)}
												</div>
												<div className="min-w-0 flex-1">
													<CardTitle className="text-base font-semibold truncate group-hover:text-primary transition-colors duration-200">
														{project.name}
													</CardTitle>
													<CardDescription className="mt-0.5">
														<span className="inline-flex items-center rounded-md bg-muted/80 px-1.5 py-0.5 text-xs font-mono font-semibold text-muted-foreground/70 ring-1 ring-border/50">
															{project.key}
														</span>
													</CardDescription>
												</div>
											</div>
											<Badge
												variant={
													project.status === "active" ? "default" : "secondary"
												}
												className={
													project.status === "active"
														? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-0"
														: ""
												}
											>
												{project.status === "active"
													? "Активный"
													: project.status}
											</Badge>
										</div>
									</CardHeader>
									<CardContent>
										{project.description && (
											<p className="mb-3 line-clamp-2 text-sm text-muted-foreground leading-relaxed">
												{project.description}
											</p>
										)}
										<div className="flex items-center gap-1.5 text-xs text-muted-foreground">
											<svg
												className="h-3.5 w-3.5"
												viewBox="0 0 16 16"
												fill="none"
												stroke="currentColor"
												strokeWidth="1.5"
												strokeLinecap="round"
											>
												<circle cx="8" cy="5" r="3" />
												<path d="M2 14c0-3 2.7-5 6-5s6 2 6 5" />
											</svg>
											<span>
												{project.member_count}{" "}
												{project.member_count === 1 ? "участник" : "участников"}
											</span>
										</div>
									</CardContent>
								</Card>
							</div>
						</Link>
					))}
				</div>
			)}
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/projects/")({
	component: ProjectsPage,
});
