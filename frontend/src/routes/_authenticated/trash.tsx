import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import type { ProjectResponse } from "@/api/projects";
import { listTrash, restoreProject } from "@/api/projects";
import { toastApiError } from "@/lib/format-api-error";

/**
 * Страница корзины — список удалённых/архивных проектов с возможностью восстановления.
 * Вкладки: элементы (таблица) + «О корзине» (lifecycle + типы).
 */

/** Дней до удаления из корзины (retention 30 дней от updated_at). */
function daysRemaining(updatedAt: string): number {
	const deleted = new Date(updatedAt).getTime();
	const now = Date.now();
	const elapsed = Math.floor((now - deleted) / (1000 * 60 * 60 * 24));
	return Math.max(0, 30 - elapsed);
}

/** Класс бейджа по оставшимся дням retention. */
function daysBadgeClass(days: number): string {
	if (days < 10) return "bg-red/[0.08] text-red";
	if (days <= 20) return "bg-orange/[0.08] text-orange";
	return "bg-blue/[0.08] text-blue";
}

function TrashPage() {
	const queryClient = useQueryClient();
	const [activeTab, setActiveTab] = useState<"items" | "about">("items");
	const [confirmProject, setConfirmProject] = useState<ProjectResponse | null>(
		null,
	);

	const {
		data: deletedProjects,
		isLoading,
		isError,
		refetch,
	} = useQuery({
		queryKey: ["projects", "trash"],
		queryFn: listTrash,
	});

	const restoreMutation = useMutation({
		mutationFn: restoreProject,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["projects"] });
			toast.success("Проект восстановлен");
		},
		onError: (err) => toastApiError(err, "Не удалось восстановить проект"),
	});

	return (
		<div className="flex-1 overflow-y-auto p-6 animate-[eu_450ms_var(--spring)_both]">
			{/* Заголовок в стиле topbar */}
			<div className="flex items-center gap-4 mb-6">
				<span
					className="text-[18px] font-extrabold tracking-[-0.03em] text-t1"
					data-testid="trash-page-heading"
				>
					Корзина
				</span>
				<div className="flex gap-[3px] ml-3">
					<button
						type="button"
						className={`px-[14px] py-[7px] text-[13px] font-medium rounded-[var(--r2)] transition-colors duration-200 cursor-pointer relative ${
							activeTab === "items"
								? "text-t1 font-semibold bg-[rgba(107,114,128,0.06)]"
								: "text-t3 hover:text-t2 hover:bg-[var(--hover)]"
						}`}
						onClick={() => setActiveTab("items")}
					>
						Удаленные
						{activeTab === "items" && (
							<span className="absolute bottom-0 left-1/4 right-1/4 h-[2px] bg-[var(--gray)] rounded-[1px]" />
						)}
					</button>
					<button
						type="button"
						className={`px-[14px] py-[7px] text-[13px] font-medium rounded-[var(--r2)] transition-colors duration-200 cursor-pointer relative ${
							activeTab === "about"
								? "text-t1 font-semibold bg-[rgba(107,114,128,0.06)]"
								: "text-t3 hover:text-t2 hover:bg-[var(--hover)]"
						}`}
						onClick={() => setActiveTab("about")}
					>
						О корзине
						{activeTab === "about" && (
							<span className="absolute bottom-0 left-1/4 right-1/4 h-[2px] bg-[var(--gray)] rounded-[1px]" />
						)}
					</button>
				</div>
				<div className="flex-1" />
				<button
					type="button"
					className="text-[13px] text-red font-semibold px-[14px] py-[7px] rounded-[var(--r2)] transition-colors duration-180 hover:bg-red/[0.06]"
				>
					Очистить все
				</button>
			</div>

			{/* Вкладка элементов */}
			{activeTab === "items" && (
				<div>
					{/* Загрузка */}
					{isLoading && (
						<div className="space-y-3">
							{["sk-1", "sk-2"].map((sk) => (
								<div
									key={sk}
									className="rounded-[var(--r4)] border border-b1 bg-card p-5"
								>
									<div className="space-y-2">
										<div className="h-5 w-1/3 rounded-md animate-shimmer" />
										<div className="h-4 w-1/5 rounded-md animate-shimmer" />
									</div>
								</div>
							))}
						</div>
					)}

					{/* Ошибка загрузки — не путать с пустой корзиной */}
					{!isLoading && isError && (
						<div
							role="alert"
							className="flex flex-col items-center justify-center rounded-[var(--r4)] border border-b1 bg-card/50 py-16 px-6 gap-3"
						>
							<p className="text-[14px] font-medium text-t1">
								Не удалось загрузить корзину
							</p>
							<button
								type="button"
								className="text-[13px] text-blue font-semibold hover:underline"
								onClick={() => void refetch()}
							>
								Попробовать снова
							</button>
						</div>
					)}

					{/* Пустое состояние */}
					{!isLoading &&
						!isError &&
						(!deletedProjects || deletedProjects.length === 0) && (
							<div className="flex flex-col items-center justify-center rounded-[var(--r4)] border border-dashed border-b2 bg-card/50 py-16 px-6">
								<div className="flex h-12 w-12 items-center justify-center rounded-full bg-raised mb-4">
									<svg
										aria-hidden="true"
										className="h-6 w-6 text-t3"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										strokeWidth="1.8"
									>
										<polyline points="3 6 5 6 21 6" />
										<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
										<line x1="10" y1="11" x2="10" y2="17" />
										<line x1="14" y1="11" x2="14" y2="17" />
									</svg>
								</div>
								<h3 className="text-[18px] font-bold mb-2">Корзина пуста</h3>
								<p className="text-[14px] text-t2 text-center max-w-sm">
									Здесь будут отображаться удалённые элементы
								</p>
							</div>
						)}

					{/* Таблица */}
					{!isError && deletedProjects && deletedProjects.length > 0 && (
						<table className="w-full border-collapse-separate border-spacing-0">
							<thead>
								<tr>
									<th
										className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-[18px] py-[14px] border-b border-b2"
										style={{ width: 44 }}
									/>
									<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-[18px] py-[14px] border-b border-b2">
										Название
									</th>
									<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-[18px] py-[14px] border-b border-b2">
										Удалил
									</th>
									<th
										className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-[18px] py-[14px] border-b border-b2"
										style={{ width: 120 }}
									>
										Осталось
									</th>
									<th
										className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-[18px] py-[14px] border-b border-b2"
										style={{ width: 220 }}
									>
										Действия
									</th>
								</tr>
							</thead>
							<tbody>
								{deletedProjects.map(
									(project: ProjectResponse, idx: number) => {
										const days = daysRemaining(project.updated_at);
										return (
											<tr
												key={project.id}
												className="cursor-default hover:bg-[var(--hover)] animate-[eu_350ms_var(--spring)_both]"
												style={{ animationDelay: `${idx * 40}ms` }}
											>
												<td className="px-[18px] py-4 border-b border-b1 align-middle transition-[background] duration-150">
													<span className="w-5 h-5 inline-block">
														<svg
															aria-hidden="true"
															className="w-5 h-5"
															viewBox="0 0 24 24"
															fill="none"
															stroke="var(--blue)"
															strokeWidth="1.8"
														>
															<rect x="2" y="2" width="5" height="5" rx="1" />
															<rect x="9" y="2" width="5" height="5" rx="1" />
															<rect x="2" y="9" width="5" height="5" rx="1" />
															<rect x="9" y="9" width="5" height="5" rx="1" />
														</svg>
													</span>
												</td>
												<td className="px-[18px] py-4 border-b border-b1 align-middle transition-[background] duration-150">
													<span className="text-[14px] font-semibold">
														{project.name}
													</span>
												</td>
												<td className="px-[18px] py-4 border-b border-b1 align-middle transition-[background] duration-150">
													<span className="text-[13px] text-t2">
														{new Date(project.updated_at).toLocaleDateString(
															"ru-RU",
														)}
													</span>
												</td>
												<td className="px-[18px] py-4 border-b border-b1 align-middle transition-[background] duration-150">
													<span
														className={`px-3 py-[3px] rounded-[var(--r1)] text-[11px] font-bold inline-block ${daysBadgeClass(days)}`}
													>
														{days} дней
													</span>
												</td>
												<td className="px-[18px] py-4 border-b border-b1 align-middle transition-[background] duration-150">
													<button
														type="button"
														onClick={() => restoreMutation.mutate(project.id)}
														disabled={restoreMutation.isPending}
														className="px-4 py-[6px] rounded-[var(--r2)] text-[12px] font-bold bg-blue/[0.08] text-blue border border-blue/[0.15] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 mr-2 hover:bg-blue/[0.15] hover:-translate-y-px"
													>
														Восстановить
													</button>
													<button
														type="button"
														onClick={() => setConfirmProject(project)}
														className="px-4 py-[6px] rounded-[var(--r2)] text-[12px] font-bold bg-red/[0.08] text-red border border-red/[0.15] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 hover:bg-red/[0.15] hover:-translate-y-px"
													>
														Удалить
													</button>
												</td>
											</tr>
										);
									},
								)}
							</tbody>
						</table>
					)}
				</div>
			)}

			{/* Вкладка «О корзине» */}
			{activeTab === "about" && (
				<div>
					<h2 className="text-[18px] font-extrabold mb-6 tracking-[-0.02em]">
						Жизненный цикл
					</h2>

					{/* Шаги жизненного цикла */}
					<div className="flex gap-6 mb-10">
						{[
							{
								n: "1",
								title: "Удаление",
								desc: "Элемент перемещается в корзину, ссылки сохраняются",
							},
							{
								n: "2",
								title: "30 дней",
								desc: "Можно восстановить. Обратный отсчет виден.",
							},
							{
								n: "3",
								title: "Навсегда",
								desc: "Автоматическое безвозвратное удаление данных",
							},
						].map((s, idx) => (
							<div key={s.n} className="contents">
								{idx > 0 && (
									<span className="flex items-center pt-[10px] text-t4 text-[20px] shrink-0">
										&rarr;
									</span>
								)}
								<div className="flex-1 text-center">
									<div className="w-11 h-11 rounded-full bg-gradient-to-br from-[var(--blueS)] to-[var(--purpleS)] flex items-center justify-center text-[17px] font-extrabold text-white mx-auto mb-[14px] shadow-[0_4px_16px_var(--gB)]">
										{s.n}
									</div>
									<h4 className="text-[14px] font-bold mb-[5px] tracking-[-0.01em]">
										{s.title}
									</h4>
									<p className="text-[12.5px] text-t2 leading-[1.5]">
										{s.desc}
									</p>
								</div>
							</div>
						))}
					</div>

					{/* Карточки типов сущностей */}
					<h3 className="text-[16px] font-bold mb-4">Что можно удалить</h3>
					<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-[14px]">
						{[
							{
								title: "Задачи",
								desc: "Issues, story points, comments",
								icon: (
									<svg
										aria-hidden="true"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										strokeWidth="1.8"
									>
										<rect x="4" y="4" width="6" height="16" rx="1" />
										<rect x="14" y="4" width="6" height="10" rx="1" />
									</svg>
								),
							},
							{
								title: "Каналы",
								desc: "Каналы и история сообщений",
								icon: (
									<svg
										aria-hidden="true"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										strokeWidth="1.8"
									>
										<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
									</svg>
								),
							},
							{
								title: "Тест-прогоны",
								desc: "Результаты и статистика",
								icon: (
									<svg
										aria-hidden="true"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										strokeWidth="1.8"
									>
										<path d="M9 11l3 3L22 4" />
										<path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11" />
									</svg>
								),
							},
							{
								title: "Папки TMS",
								desc: "Папки с тест-кейсами",
								icon: (
									<svg
										aria-hidden="true"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										strokeWidth="1.8"
									>
										<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
									</svg>
								),
							},
							{
								title: "Вложения",
								desc: "Файлы, документы, медиа",
								icon: (
									<svg
										aria-hidden="true"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										strokeWidth="1.8"
									>
										<path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z" />
										<polyline points="13 2 13 9 20 9" />
									</svg>
								),
							},
							{
								title: "Проекты",
								desc: "Весь проект со всеми данными",
								icon: (
									<svg
										aria-hidden="true"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										strokeWidth="1.8"
									>
										<path d="M12 2L2 7l10 5 10-5-10-5z" />
										<path d="M2 17l10 5 10-5" />
										<path d="M2 12l10 5 10-5" />
									</svg>
								),
							},
						].map((card) => (
							<div
								key={card.title}
								className="onb-tour-card bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-[22px] text-center transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-250 hover:border-b2 hover:-translate-y-[2px] hover:shadow-[var(--sh2)]"
							>
								<div className="w-7 h-7 text-t2 mb-[10px] mx-auto">
									{card.icon}
								</div>
								<h4 className="text-[14px] font-bold mb-1">{card.title}</h4>
								<p className="text-[12px] text-t3 leading-[1.4]">{card.desc}</p>
							</div>
						))}
					</div>
				</div>
			)}

			{/* Диалог подтверждения */}
			{confirmProject && (
				<div className="fixed inset-0 bg-black/65 backdrop-blur-[8px] z-[1000] flex items-center justify-center">
					<button
						type="button"
						className="absolute inset-0 cursor-default border-0 bg-transparent p-0"
						aria-label="Закрыть"
						onClick={() => setConfirmProject(null)}
					/>
					<div
						role="dialog"
						aria-modal="true"
						className="relative bg-base border border-b2 rounded-[var(--r4)] shadow-[var(--sh4)] max-w-[420px] w-[90%] p-7 text-center animate-[scIn_250ms_var(--spring)]"
					>
						<div className="text-[44px] mb-[14px]">&#9888;&#65039;</div>
						<h3 className="text-[18px] font-bold mb-2">Удалить навсегда?</h3>
						<p className="text-[14px] text-t2 mb-[22px] leading-[1.6]">
							Это действие необратимо. Данные будут полностью уничтожены.
						</p>
						<div className="flex gap-[10px] justify-center">
							<button
								type="button"
								onClick={() => setConfirmProject(null)}
								className="px-[22px] py-[9px] rounded-[var(--r2)] text-[14px] font-semibold transition-colors duration-200 border border-b2 text-t1 hover:bg-[var(--hover)]"
							>
								Отмена
							</button>
							<button
								type="button"
								onClick={() => {
									setConfirmProject(null);
									toast.success("Элемент удален навсегда");
								}}
								className="px-[22px] py-[9px] rounded-[var(--r2)] text-[14px] font-bold bg-[var(--redS)] text-white transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 shadow-[0_2px_12px_var(--gR)] hover:brightness-110 hover:-translate-y-px"
							>
								Удалить навсегда
							</button>
						</div>
					</div>
				</div>
			)}
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/trash")({
	component: TrashPage,
});
