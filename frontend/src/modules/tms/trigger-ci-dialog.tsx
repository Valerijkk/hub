import { useMutation, useQuery } from "@tanstack/react-query";
import { useCallback, useMemo, useState } from "react";
import { toast } from "sonner";
import { listGitLabIntegrations, triggerPipeline } from "@/api/gitlab";
import { getProject } from "@/api/projects";
import { listTestCases, type TestCaseResponse } from "@/api/tms";
import { Badge } from "@/components/ui/badge";
import { toastApiError } from "@/lib/format-api-error";
import { canMutateProject } from "@/lib/permissions";
import { sanitizeUrl } from "@/lib/safe-url";
import { cn } from "@/lib/utils";

/** Лимит кейсов в UI запуска CI — без пагинации (W10-A2-015). */
export const CI_CASES_FETCH_LIMIT = 500;

interface Props {
	projectId: string;
	/** Явный override; иначе берём из `project.my_role`. */
	canMutate?: boolean;
}

/**
 * TriggerCIPanel — компонент для запуска CI-пайплайна с выбранными тест-кейсами.
 * Позволяет выбрать интеграцию, ветку, отметить кейсы и запустить пайплайн.
 */
export function TriggerCIPanel({ projectId, canMutate: canMutateProp }: Props) {
	const [selectedCaseIds, setSelectedCaseIds] = useState<Set<string>>(
		new Set(),
	);
	const [branch, setBranch] = useState("main");
	const [search, setSearch] = useState("");

	const { data: project } = useQuery({
		queryKey: ["project", projectId],
		queryFn: () => getProject(projectId),
		enabled: Boolean(projectId) && canMutateProp === undefined,
	});
	const canMutate = canMutateProp ?? canMutateProject(project?.my_role);

	// Загружаем GitLab-интеграции
	const { data: integrations, isLoading: intLoading } = useQuery({
		queryKey: ["gitlab-integrations", projectId],
		queryFn: () => listGitLabIntegrations(projectId),
		enabled: canMutate,
	});

	const [selectedIntegration, setSelectedIntegration] = useState<string>("");

	// Загружаем тест-кейсы
	const { data: casesData, isLoading: casesLoading } = useQuery({
		queryKey: ["tms-cases", projectId, "ci-trigger"],
		queryFn: () =>
			listTestCases(projectId, {
				limit: CI_CASES_FETCH_LIMIT,
				status: "active",
			}),
		enabled: canMutate,
	});

	const cases: TestCaseResponse[] = casesData?.items ?? [];
	const casesTruncated =
		typeof casesData?.total === "number" &&
		casesData.total > CI_CASES_FETCH_LIMIT;

	// Фильтрация
	const filteredCases = useMemo(() => {
		if (!search.trim()) return cases;
		const q = search.toLowerCase();
		return cases.filter(
			(c) =>
				c.title.toLowerCase().includes(q) ||
				c.case_key.toLowerCase().includes(q),
		);
	}, [cases, search]);

	// Автоматически выбираем первую интеграцию
	const activeIntegrationId =
		selectedIntegration || (integrations?.[0]?.id ?? "");

	// Мутация запуска
	const triggerMutation = useMutation({
		mutationFn: () => {
			const caseKeys = cases
				.filter((c) => selectedCaseIds.has(c.id))
				.map((c) => c.case_key);

			return triggerPipeline(projectId, {
				integration_id: activeIntegrationId,
				ref: branch || undefined,
				variables: {
					HUB_TEST_CASE_IDS: caseKeys.join(","),
					HUB_RUN_SOURCE: "hub_trigger",
					HUB_PROJECT_ID: projectId,
				},
			});
		},
		onSuccess: (data) => {
			toast.success(`Пайплайн #${data.pipeline_id} запущен`, {
				description: `Ветка: ${data.ref}`,
				action: data.web_url
					? {
							label: "Открыть",
							onClick: () => {
								const safe = sanitizeUrl(data.web_url);
								if (safe !== "#") {
									window.open(safe, "_blank", "noopener,noreferrer");
								}
							},
						}
					: undefined,
			});
			setSelectedCaseIds(new Set());
		},
		onError: (err) => toastApiError(err, "Не удалось запустить пайплайн"),
	});

	// Тогглим кейс
	const toggleCase = useCallback((id: string) => {
		setSelectedCaseIds((prev) => {
			const next = new Set(prev);
			if (next.has(id)) next.delete(id);
			else next.add(id);
			return next;
		});
	}, []);

	// Выбрать все / снять все
	const toggleAll = useCallback(() => {
		if (selectedCaseIds.size === filteredCases.length) {
			setSelectedCaseIds(new Set());
		} else {
			setSelectedCaseIds(new Set(filteredCases.map((c) => c.id)));
		}
	}, [filteredCases, selectedCaseIds.size]);

	if (!canMutate) {
		return null;
	}

	const noIntegrations =
		!intLoading && (!integrations || integrations.length === 0);

	return (
		<div className="space-y-4 rounded-lg border p-4">
			<div className="flex items-center justify-between">
				<h3 className="text-base font-semibold">Запуск тестов в CI</h3>
				{selectedCaseIds.size > 0 && (
					<Badge variant="secondary">{selectedCaseIds.size} выбрано</Badge>
				)}
			</div>

			{noIntegrations ? (
				<div className="rounded-lg border border-yellow-200 bg-yellow-50 p-3 text-sm text-yellow-800 dark:border-yellow-900 dark:bg-yellow-950 dark:text-yellow-200">
					Нет подключённых GitLab-интеграций. Подключите репозиторий в{" "}
					<strong>Настройки проекта &rarr; GitLab</strong>.
				</div>
			) : (
				<>
					{/* Интеграция + ветка */}
					<div className="flex gap-3">
						{integrations && integrations.length > 1 && (
							<div className="flex-1">
								<label
									htmlFor="trigger-ci-integration"
									className="mb-1 block text-xs font-medium text-t2"
								>
									Репозиторий
								</label>
								<select
									id="trigger-ci-integration"
									value={activeIntegrationId}
									onChange={(e) => setSelectedIntegration(e.target.value)}
									className="w-full rounded-lg border border-input bg-base px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/50"
								>
									{integrations.map((ig) => (
										<option key={ig.id} value={ig.id}>
											{ig.project_path}
										</option>
									))}
								</select>
							</div>
						)}
						<div
							className={
								integrations && integrations.length > 1 ? "w-48" : "w-64"
							}
						>
							<label
								htmlFor="trigger-ci-branch"
								className="mb-1 block text-xs font-medium text-t2"
							>
								Ветка
							</label>
							<input
								id="trigger-ci-branch"
								type="text"
								value={branch}
								onChange={(e) => setBranch(e.target.value)}
								placeholder="main"
								className="w-full rounded-lg border border-input bg-base px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/50"
							/>
						</div>
					</div>

					{/* Поиск + выбрать все */}
					<div className="flex items-center gap-2">
						<input
							type="text"
							value={search}
							onChange={(e) => setSearch(e.target.value)}
							placeholder="Поиск тест-кейсов..."
							className="flex-1 rounded-lg border border-input bg-base px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring/50"
						/>
						<button
							type="button"
							onClick={toggleAll}
							className="shrink-0 rounded-lg border border-input px-3 py-2 text-xs font-medium hover:bg-accent transition-colors"
						>
							{selectedCaseIds.size === filteredCases.length &&
							filteredCases.length > 0
								? "Снять все"
								: "Выбрать все"}
						</button>
					</div>

					{(casesTruncated || cases.length >= CI_CASES_FETCH_LIMIT) && (
						<div
							className="rounded-lg border border-amber-200 bg-amber-50 p-2 text-xs text-amber-900 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-200"
							role="status"
						>
							Показаны первые {CI_CASES_FETCH_LIMIT} активных тест-кейсов
							{casesTruncated && casesData?.total
								? ` из ${casesData.total}`
								: ""}
							. Остальные недоступны для выбора в этом диалоге.
						</div>
					)}

					{/* Список кейсов */}
					<div className="max-h-72 overflow-y-auto rounded-lg border">
						{casesLoading ? (
							<div className="flex items-center justify-center py-8 text-sm text-t2">
								Загрузка тест-кейсов...
							</div>
						) : filteredCases.length === 0 ? (
							<div className="flex items-center justify-center py-8 text-sm text-t2">
								Нет активных тест-кейсов
							</div>
						) : (
							filteredCases.map((tc) => (
								<label
									key={tc.id}
									className={cn(
										"flex items-center gap-3 px-3 py-2 cursor-pointer transition-colors border-b last:border-b-0",
										selectedCaseIds.has(tc.id)
											? "bg-primary/5"
											: "hover:bg-raised/50",
									)}
								>
									<input
										type="checkbox"
										checked={selectedCaseIds.has(tc.id)}
										onChange={() => toggleCase(tc.id)}
										className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary/50"
									/>
									<Badge
										variant="outline"
										className="shrink-0 text-xs font-mono"
									>
										{tc.case_key}
									</Badge>
									<span className="text-sm truncate">{tc.title}</span>
									<Badge
										variant="secondary"
										className={cn(
											"ml-auto shrink-0 text-[10px]",
											tc.priority === "critical" &&
												"bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300",
											tc.priority === "high" &&
												"bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300",
										)}
									>
										{tc.priority}
									</Badge>
								</label>
							))
						)}
					</div>

					{/* Кнопка запуска */}
					<button
						type="button"
						disabled={
							selectedCaseIds.size === 0 ||
							!activeIntegrationId ||
							triggerMutation.isPending
						}
						onClick={() => triggerMutation.mutate()}
						className={cn(
							"w-full rounded-lg px-4 py-2.5 text-sm font-medium transition-colors",
							selectedCaseIds.size > 0 && activeIntegrationId
								? "bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm"
								: "bg-raised text-t2 cursor-not-allowed",
						)}
					>
						{triggerMutation.isPending
							? "Запускаем..."
							: `Запустить пайплайн (${selectedCaseIds.size} кейсов)`}
					</button>

					{/* Подсказка */}
					<p className="text-xs text-t2">
						Выбранные тест-кейсы передаются в переменную CI{" "}
						<code className="rounded bg-gray-100 px-1 dark:bg-gray-800">
							HUB_TEST_CASE_IDS
						</code>{" "}
						(через запятую). Используйте её в пайплайне для фильтрации тестов.
					</p>
				</>
			)}
		</div>
	);
}
