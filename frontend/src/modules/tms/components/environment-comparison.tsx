import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { compareEnvironments, listEnvironments } from "@/api/tms";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Конфигурация
// ---------------------------------------------------------------------------

const STATUS_COLORS: Record<
	string,
	{ bg: string; text: string; label: string }
> = {
	passed: {
		bg: "bg-green-100 dark:bg-green-900/30",
		text: "text-green-700 dark:text-green-400",
		label: "Passed",
	},
	failed: {
		bg: "bg-red-100 dark:bg-red-900/30",
		text: "text-red-600 dark:text-red-400",
		label: "Failed",
	},
	broken: {
		bg: "bg-orange-100 dark:bg-orange-900/30",
		text: "text-orange-600 dark:text-orange-400",
		label: "Broken",
	},
	skipped: {
		bg: "bg-yellow-100 dark:bg-yellow-900/30",
		text: "text-yellow-700 dark:text-yellow-400",
		label: "Skipped",
	},
	blocked: {
		bg: "bg-gray-200 dark:bg-gray-700/40",
		text: "text-gray-600 dark:text-gray-400",
		label: "Blocked",
	},
	unknown: {
		bg: "bg-gray-100 dark:bg-gray-800/30",
		text: "text-gray-500",
		label: "Unknown",
	},
	no_results: {
		bg: "bg-gray-50 dark:bg-gray-800/20",
		text: "text-gray-400",
		label: "Нет результатов",
	},
};

function getStatusCfg(status: string) {
	// biome-ignore lint/style/noNonNullAssertion: STATUS_COLORS.unknown is always defined
	return STATUS_COLORS[status] ?? STATUS_COLORS.unknown!;
}

const ENV_LABEL_CONFIG: Record<
	string,
	{ bg: string; text: string; label: string }
> = {
	dev: {
		bg: "bg-blue-500/10",
		text: "text-blue-700 dark:text-blue-400",
		label: "Dev",
	},
	staging: {
		bg: "bg-yellow-500/10",
		text: "text-yellow-700 dark:text-yellow-400",
		label: "Staging",
	},
	prod: {
		bg: "bg-red-500/10",
		text: "text-red-600 dark:text-red-400",
		label: "Prod",
	},
	custom: {
		bg: "bg-gray-500/10",
		text: "text-gray-600 dark:text-gray-400",
		label: "Custom",
	},
};

// ---------------------------------------------------------------------------
// Статус-бейдж
// ---------------------------------------------------------------------------

function StatusBadge({ status }: { status: string }) {
	const cfg = getStatusCfg(status);
	return (
		<Badge
			variant="secondary"
			className={cn("text-[10px] font-semibold px-2 py-0.5", cfg.bg, cfg.text)}
		>
			{cfg.label}
		</Badge>
	);
}

// ---------------------------------------------------------------------------
// Основной компонент
// ---------------------------------------------------------------------------

/** Пропсы компонента EnvironmentComparison. */
interface EnvironmentComparisonProps {
	/** ID проекта. */
	projectId: string;
}

/**
 * Сравнение результатов тестов по двум окружениям.
 * Выбираются два окружения, затем отображается таблица
 * с результатами по каждому тест-кейсу с подсветкой различий.
 */
export function EnvironmentComparison({
	projectId,
}: EnvironmentComparisonProps) {
	const [envId1, setEnvId1] = useState<string>("");
	const [envId2, setEnvId2] = useState<string>("");
	const [submitted, setSubmitted] = useState(false);

	/** Загрузка списка окружений. */
	const { data: environments, isLoading: envsLoading } = useQuery({
		queryKey: ["tms-environments", projectId],
		queryFn: () => listEnvironments(projectId),
	});

	/** Загрузка сравнения. */
	const {
		data: comparison,
		isLoading: compLoading,
		error: compError,
	} = useQuery({
		queryKey: ["tms-env-comparison", projectId, envId1, envId2],
		queryFn: () => compareEnvironments(projectId, envId1, envId2),
		enabled: submitted && !!envId1 && !!envId2 && envId1 !== envId2,
	});

	/** Фильтр: только различия. */
	const [showDiffsOnly, setShowDiffsOnly] = useState(false);

	/** Отфильтрованные результаты. */
	const items = useMemo(() => {
		if (!comparison?.items) return [];
		if (!showDiffsOnly) return comparison.items;
		return comparison.items.filter(
			(item) => item.status_env1 !== item.status_env2,
		);
	}, [comparison, showDiffsOnly]);

	/** Подсчёт различий. */
	const diffCount = useMemo(() => {
		if (!comparison?.items) return 0;
		return comparison.items.filter(
			(item) => item.status_env1 !== item.status_env2,
		).length;
	}, [comparison]);

	const handleCompare = () => {
		if (envId1 && envId2 && envId1 !== envId2) {
			setSubmitted(true);
		}
	};

	const envs = environments ?? [];

	// Загрузка окружений
	if (envsLoading) {
		return (
			<div className="py-8 text-center text-t2">Загрузка окружений...</div>
		);
	}

	// Нет окружений
	if (envs.length < 2) {
		return (
			<Card>
				<CardContent className="py-8 text-center text-t2">
					<div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-raised/60 to-raised/30 border border-dashed border-t2/20">
						<svg
							aria-hidden="true"
							className="h-8 w-8 text-t2/30"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={1.5}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2z"
							/>
						</svg>
					</div>
					<p className="font-medium mb-1">
						Недостаточно окружений для сравнения
					</p>
					<p className="text-sm">
						{envs.length === 0
							? "Создайте хотя бы два тестовых окружения"
							: "Создайте ещё одно окружение для сравнения"}
					</p>
				</CardContent>
			</Card>
		);
	}

	return (
		<div className="space-y-4">
			{/* Выбор окружений */}
			<div className="flex flex-wrap items-end gap-3 rounded-lg border bg-card/50 p-4">
				<div className="space-y-1.5">
					<label className="text-xs font-medium text-t2">Окружение 1</label>
					<Select
						value={envId1 || "placeholder"}
						onValueChange={(v) => {
							setEnvId1(v === "placeholder" ? "" : v);
							setSubmitted(false);
						}}
					>
						<SelectTrigger className="rounded-xl border bg-card px-3 py-2 text-sm min-w-[200px]">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="placeholder">Выберите окружение</SelectItem>
							{envs.map((env) => {
								const labelCfg = ENV_LABEL_CONFIG[env.label];
								return (
									<SelectItem
										key={env.id}
										value={env.id}
										disabled={env.id === envId2}
									>
										{env.name} ({labelCfg?.label ?? env.label})
									</SelectItem>
								);
							})}
						</SelectContent>
					</Select>
				</div>

				<div className="flex items-center pb-2">
					<svg
						aria-hidden="true"
						className="h-5 w-5 text-t2"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2}
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
						/>
					</svg>
				</div>

				<div className="space-y-1.5">
					<label className="text-xs font-medium text-t2">Окружение 2</label>
					<Select
						value={envId2 || "placeholder"}
						onValueChange={(v) => {
							setEnvId2(v === "placeholder" ? "" : v);
							setSubmitted(false);
						}}
					>
						<SelectTrigger className="rounded-xl border bg-card px-3 py-2 text-sm min-w-[200px]">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="placeholder">Выберите окружение</SelectItem>
							{envs.map((env) => {
								const labelCfg = ENV_LABEL_CONFIG[env.label];
								return (
									<SelectItem
										key={env.id}
										value={env.id}
										disabled={env.id === envId1}
									>
										{env.name} ({labelCfg?.label ?? env.label})
									</SelectItem>
								);
							})}
						</SelectContent>
					</Select>
				</div>

				<Button
					onClick={handleCompare}
					disabled={!envId1 || !envId2 || envId1 === envId2 || compLoading}
					className="shrink-0"
				>
					{compLoading ? "Сравнение..." : "Сравнить"}
				</Button>
			</div>

			{/* Ошибка */}
			{compError && (
				<div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950/30 p-3 text-sm text-red-700 dark:text-red-400">
					{(compError as { message?: string }).message ??
						"Ошибка сравнения окружений"}
				</div>
			)}

			{/* Результаты */}
			{comparison && (
				<>
					{/* Заголовки окружений + фильтр */}
					<div className="flex flex-wrap items-center gap-4">
						<div className="flex items-center gap-2">
							<Badge
								variant="secondary"
								className={cn(
									"text-xs font-semibold",
									ENV_LABEL_CONFIG[comparison.env1.label]?.bg ??
										"bg-gray-500/10",
									ENV_LABEL_CONFIG[comparison.env1.label]?.text ??
										"text-gray-600",
								)}
							>
								{comparison.env1.name}
							</Badge>
							<span className="text-xs text-t2">vs</span>
							<Badge
								variant="secondary"
								className={cn(
									"text-xs font-semibold",
									ENV_LABEL_CONFIG[comparison.env2.label]?.bg ??
										"bg-gray-500/10",
									ENV_LABEL_CONFIG[comparison.env2.label]?.text ??
										"text-gray-600",
								)}
							>
								{comparison.env2.name}
							</Badge>
						</div>

						<span className="text-xs text-t2 tabular-nums">
							Всего: {comparison.items.length} | Различий: {diffCount}
						</span>

						<label className="ml-auto flex items-center gap-2 text-xs cursor-pointer">
							<input
								type="checkbox"
								className="rounded border-t2/30"
								checked={showDiffsOnly}
								onChange={(e) => setShowDiffsOnly(e.target.checked)}
							/>
							Только различия
						</label>
					</div>

					{/* Таблица */}
					{items.length > 0 ? (
						<div className="rounded-xl border shadow-card overflow-auto max-h-[calc(100vh-350px)]">
							<table className="w-full">
								<thead className="sticky top-0 z-10">
									<tr className="bg-raised/60 backdrop-blur-sm">
										<th className="border-b px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-t2">
											Тест-кейс
										</th>
										<th className="border-b px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-t2">
											Название
										</th>
										<th className="border-b px-4 py-2.5 text-center text-[11px] font-bold uppercase tracking-wider text-t2">
											{comparison.env1.name}
										</th>
										<th className="border-b px-4 py-2.5 text-center text-[11px] font-bold uppercase tracking-wider text-t2">
											{comparison.env2.name}
										</th>
										<th className="border-b px-4 py-2.5 text-center text-[11px] font-bold uppercase tracking-wider text-t2">
											Статус
										</th>
									</tr>
								</thead>
								<tbody>
									{items.map((item) => {
										const isDiff = item.status_env1 !== item.status_env2;

										return (
											<tr
												key={item.case_key}
												className={cn(
													"border-b transition-colors",
													isDiff
														? "bg-amber-50/50 dark:bg-amber-900/5 hover:bg-amber-50 dark:hover:bg-amber-900/10"
														: "hover:bg-raised/30",
												)}
											>
												<td className="px-4 py-2.5">
													<span className="font-mono text-xs font-bold text-primary">
														{item.case_key}
													</span>
												</td>
												<td className="max-w-md px-4 py-2.5">
													<span className="text-sm line-clamp-1">
														{item.case_title}
													</span>
												</td>
												<td className="px-4 py-2.5 text-center">
													<StatusBadge status={item.status_env1} />
												</td>
												<td className="px-4 py-2.5 text-center">
													<StatusBadge status={item.status_env2} />
												</td>
												<td className="px-4 py-2.5 text-center">
													{isDiff ? (
														<span className="inline-flex items-center gap-1 text-[10px] font-medium text-amber-600 dark:text-amber-400">
															<svg
																aria-hidden="true"
																className="h-3 w-3"
																fill="none"
																viewBox="0 0 24 24"
																stroke="currentColor"
																strokeWidth={2}
															>
																<path
																	strokeLinecap="round"
																	strokeLinejoin="round"
																	d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"
																/>
															</svg>
															Различие
														</span>
													) : (
														<span className="text-[10px] text-t2/50">
															Совпадает
														</span>
													)}
												</td>
											</tr>
										);
									})}
								</tbody>
							</table>
						</div>
					) : (
						<div className="py-12 text-center text-t2">
							{showDiffsOnly
								? "Различий не обнаружено -- результаты идентичны"
								: "Нет результатов для сравнения"}
						</div>
					)}
				</>
			)}

			{/* Начальное состояние */}
			{!submitted && !comparison && (
				<div className="rounded-lg border bg-card/50 p-8 text-center">
					<p className="text-sm text-t2">
						Выберите два окружения и нажмите "Сравнить" для просмотра различий в
						результатах тестирования
					</p>
				</div>
			)}
		</div>
	);
}
