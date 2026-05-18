import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { listProjects } from "@/api/projects";
import { QueryStateView } from "@/components/query-state-view";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

/**
 * Страница списка проектов — сетка карточек, пустое состояние,
 * скелетон-загрузка, CTA для создания нового проекта.
 */
function ProjectsPage() {
	const projectsQuery = useQuery({
		queryKey: ["projects"],
		queryFn: listProjects,
	});

	return (
		<div className="mx-auto max-w-5xl animate-[eu_450ms_var(--spring)_both]">
			{/* Заголовок страницы — градиентный */}
			<div className="mb-8 flex items-center justify-between">
				<div className="flex items-center gap-3">
					<div className="flex h-10 w-10 items-center justify-center rounded-[var(--r3)] bg-gradient-to-br from-[var(--blueS)]/20 to-[var(--purpleS)]/20 text-blue">
						<svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
							<rect x="3" y="3" width="7" height="7" rx="2" />
							<rect x="14" y="3" width="7" height="7" rx="2" />
							<rect x="3" y="14" width="7" height="7" rx="2" />
							<rect x="14" y="14" width="7" height="7" rx="2" />
						</svg>
					</div>
					<div>
						<h1 className="text-[18px] font-extrabold tracking-[-0.03em] text-t1">Проекты</h1>
						<p className="mt-0.5 text-[13px] text-t2 leading-[1.5]">
							Управляйте проектами вашей команды
						</p>
					</div>
				</div>
				<Button
					asChild
					className="bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] hover:from-blue hover:to-[var(--blueS)] text-white shadow-md shadow-blue/25 hover-pulse"
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

			<Tabs defaultValue="list">
				<TabsList>
					<TabsTrigger value="list">Проекты</TabsTrigger>
					<TabsTrigger value="info">О проектах</TabsTrigger>
				</TabsList>

				<TabsContent value="info">
					<section className="rounded-[var(--r4)] border border-b1 bg-gradient-to-b from-card to-raised p-6 space-y-6">
						<div>
							<h2 className="text-[14px] font-bold text-t1 mb-4">Как создать проект</h2>
							<div className="grid gap-4 md:grid-cols-3">
								{[
									{ n: "1", color: "blue", title: "Нажмите \"Создать проект\"", desc: "Укажите название и ключ (KEY-1)" },
									{ n: "2", color: "green", title: "Настройте команду", desc: "Добавьте участников и роли" },
									{ n: "3", color: "violet", title: "Начинайте работать!", desc: "Задачи, тесты, мессенджер и звонки" },
								].map((s) => (
									<div key={s.n} className="flex gap-3">
										<div className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-[var(--r2)] bg-${s.color}-500/10 text-${s.color}-500 text-sm font-bold`}>
											{s.n}
										</div>
										<div>
											<p className="font-medium text-[13px] text-t1">{s.title}</p>
											<p className="text-[12px] text-t3 mt-0.5">{s.desc}</p>
										</div>
									</div>
								))}
							</div>
						</div>
						<div className="grid gap-3 md:grid-cols-2">
							{[
								{ icon: "\u{1F4CB}", title: "Трекер задач", desc: "Kanban-доска, бэклог, спринты" },
								{ icon: "\u{1F9EA}", title: "Тест-менеджмент", desc: "Тест-кейсы, прогоны, отчёты" },
								{ icon: "\u{1F4AC}", title: "Мессенджер", desc: "Каналы, DM, голосовые, стикеры" },
								{ icon: "\u{1F4DE}", title: "Созвоны", desc: "Аудио, screen share, запись" },
								{ icon: "\u{1F517}", title: "GitLab / GitHub", desc: "Ветки, MR, пайплайны из Hub" },
								{ icon: "\u{1F4CA}", title: "Дашборды", desc: "Виджеты, графики, метрики" },
							].map((f) => (
								<div key={f.title} className="flex gap-3 rounded-[var(--r3)] border border-b1 p-3 hover:bg-[var(--hover)] transition-colors duration-150">
									<span className="text-xl">{f.icon}</span>
									<div>
										<p className="font-medium text-[13px] text-t1">{f.title}</p>
										<p className="text-[12px] text-t3">{f.desc}</p>
									</div>
								</div>
							))}
						</div>
					</section>
				</TabsContent>

				<TabsContent value="list" className="space-y-0">

			<QueryStateView
				query={projectsQuery}
				isEmpty={(p) => p.length === 0}
				loadingFallback={
					<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
						{[1, 2, 3].map((i) => (
							<div key={`proj-sk-${i}`} className="rounded-[var(--r4)] border border-b1 bg-gradient-to-b from-card to-raised p-5">
								<div className="space-y-3">
									<div className="h-5 w-2/3 rounded-md animate-shimmer" />
									<div className="h-4 w-1/4 rounded-md animate-shimmer" />
									<div className="h-4 w-full rounded-md animate-shimmer" />
									<div className="h-3 w-1/3 rounded-md animate-shimmer" />
								</div>
							</div>
						))}
					</div>
				}
				emptyFallback={
					<div className="flex flex-col items-center justify-center rounded-[var(--r4)] border border-dashed border-b1 bg-card/50 py-20 px-6 animate-[eu_450ms_var(--spring)_both]">
						{/* Illustration */}
						<div className="relative mb-6">
							<div className="flex h-20 w-20 items-center justify-center rounded-2xl bg-gradient-to-br from-[var(--blueS)]/10 to-[var(--purpleS)]/10">
								<svg
									className="h-10 w-10 text-blue/60"
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
							<div className="absolute -top-1 -right-1 h-5 w-5 rounded-full bg-blue/20 animate-pulse" />
							<div
								className="absolute -bottom-1 -left-1 h-3 w-3 rounded-full bg-[var(--purpleS)]/20 animate-pulse"
								style={{ animationDelay: "0.5s" }}
							/>
						</div>
						<h3 className="text-[18px] font-extrabold tracking-[-0.03em] text-t1 mb-2">Нет проектов</h3>
						<p className="mb-8 text-[13px] text-t2 text-center max-w-sm leading-[1.5]">
							Создайте первый проект для вашей команды, чтобы начать работу с
							задачами, тестами и общением.
						</p>
						<Button
							asChild
							className="bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] hover:from-blue hover:to-[var(--blueS)] text-white shadow-md shadow-blue/25 hover-pulse"
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
				}
			>
				{(projects) => (
					<div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
						{projects.map((project, index) => (
						<Link
							key={project.id}
							to="/projects/$projectId"
							params={{ projectId: project.id }}
							className="group block animate-fade-in-up"
							style={{ animationDelay: `${index * 50}ms` }}
						>
							<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5 transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-250 relative overflow-hidden hover:border-b2 hover:-translate-y-[2px] hover:shadow-[var(--sh2)]">
								<div className="flex items-start justify-between gap-2 mb-3">
									<div className="flex items-center gap-3 min-w-0 flex-1">
										{/* Иконка проекта — ключ проекта */}
										<div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-[var(--r3)] bg-gradient-to-br from-[var(--blueS)]/15 to-[var(--purpleS)]/5 text-blue font-bold text-xs tracking-tight ring-1 ring-blue/10">
											{project.key.substring(0, 3)}
										</div>
										<div className="min-w-0 flex-1">
											<p className="text-[14px] font-bold tracking-[-0.01em] text-t1 truncate group-hover:text-blue transition-colors duration-200">
												{project.name}
											</p>
											<span className="inline-flex items-center rounded-md bg-input px-1.5 py-0.5 text-xs font-mono font-semibold text-t3 ring-1 ring-b1">
												{project.key}
											</span>
										</div>
									</div>
									<Badge
										variant={
											project.status === "active" ? "default" : "secondary"
										}
										className={
											project.status === "active"
												? "bg-emerald-500/[0.08] text-emerald-400 border border-emerald-500/[0.15] px-[10px] py-[3px] rounded-full text-[11px] font-[650]"
												: ""
										}
									>
										{project.status === "active"
											? "Активный"
											: project.status}
									</Badge>
								</div>
								{project.description && (
									<p className="mb-3 line-clamp-2 text-[13px] text-t2 leading-[1.5]">
										{project.description}
									</p>
								)}
								<div className="flex items-center gap-1.5 text-[12px] text-t3 font-medium">
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
							</div>
						</Link>
					))}
					</div>
				)}
			</QueryStateView>
				</TabsContent>
			</Tabs>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/projects/")({
	component: ProjectsPage,
});
