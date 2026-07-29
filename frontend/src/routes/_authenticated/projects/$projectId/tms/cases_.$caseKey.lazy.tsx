import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createLazyFileRoute, Link } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { getProject } from "@/api/projects";
import {
	getTestCase,
	linkTestCaseToIssue,
	listVersions,
	rollbackVersion,
	stripStepUid,
	type TestCasePriority,
	type TestCaseStatus,
	type TestStep,
	unlinkTestCaseFromIssue,
	updateTestCase,
} from "@/api/tms";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { toastApiError } from "@/lib/format-api-error";
import { canMutateProject } from "@/lib/permissions";
import { cn } from "@/lib/utils";
import { StepEditor } from "@/modules/tms/components/step-editor";
import {
	PRIORITY_CONFIG,
	STATUS_CONFIG,
	TestCaseMetadata,
} from "@/modules/tms/components/test-case-metadata";
import { VersionHistory } from "@/modules/tms/components/version-history";

/** Client-only stable id для React key строк шагов. */
function makeStepUid(): string {
	if (
		typeof crypto !== "undefined" &&
		typeof crypto.randomUUID === "function"
	) {
		return crypto.randomUUID();
	}
	return `step-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

/** Подписи приоритетов. */
const PRIORITY_LABELS: Record<string, string> = {
	critical: "Критический",
	high: "Высокий",
	medium: "Средний",
	low: "Низкий",
	trivial: "Тривиальный",
};

/** Подписи статусов. */
const STATUS_LABELS: Record<string, string> = {
	draft: "Черновик",
	active: "Активный",
	deprecated: "Устаревший",
};

/**
 * Детальная страница кейса — UI уровня Allure TestOps.
 * Шапка с key и статусами, вкладки Steps / Linked Issues / Autotest,
 * правый sidebar: метаданные и история версий.
 */
/** Страница деталей тест-кейса (экспорт для Vitest authz UI). */
export function TestCaseDetailPage() {
	const { projectId, caseKey } = Route.useParams();
	const queryClient = useQueryClient();
	const [activeTab, setActiveTab] = useState<"steps" | "links" | "autotest">(
		"steps",
	);
	const [versionsOpen, setVersionsOpen] = useState(false);
	const [linkIssueId, setLinkIssueId] = useState("");
	const [editing, setEditing] = useState(false);

	// Поля редактирования
	const [editTitle, setEditTitle] = useState("");
	const [editDescription, setEditDescription] = useState("");
	const [editPrecondition, setEditPrecondition] = useState("");
	const [editPostcondition, setEditPostcondition] = useState("");
	const [editPriority, setEditPriority] = useState<TestCasePriority>("medium");
	const [editStatus, setEditStatus] = useState<TestCaseStatus>("draft");
	const [editSteps, setEditSteps] = useState<TestStep[]>([]);

	// Загрузка тест-кейса
	const {
		data: tc,
		isLoading,
		isError,
		refetch,
	} = useQuery({
		queryKey: ["tms-case", projectId, caseKey],
		queryFn: () => getTestCase(projectId, caseKey),
	});

	const { data: project } = useQuery({
		queryKey: ["project", projectId],
		queryFn: () => getProject(projectId),
	});
	/** Edit / link / unlink / rollback — скрыты для viewer/guest (P2-FE-16). */
	const canMutate = canMutateProject(project?.my_role);
	const canRollback = canMutate;

	// Инициализация формы после загрузки данных
	useEffect(() => {
		if (tc) {
			setEditTitle(tc.title);
			setEditDescription(tc.description ?? "");
			setEditPrecondition(tc.precondition ?? "");
			setEditPostcondition(tc.postcondition ?? "");
			setEditPriority(tc.priority);
			setEditStatus(tc.status);
			setEditSteps(
				tc.steps ? tc.steps.map((s) => ({ ...s, _uid: makeStepUid() })) : [],
			);
		}
	}, [tc]);

	// Загрузка версий
	const { data: versions } = useQuery({
		queryKey: ["tms-case-versions", projectId, caseKey],
		queryFn: () => listVersions(projectId, caseKey),
		enabled: versionsOpen,
	});

	// Обновление тест-кейса
	const updateMutation = useMutation({
		mutationFn: () => {
			if (!tc) {
				return Promise.reject(new Error("test case not loaded"));
			}
			return updateTestCase(projectId, caseKey, {
				title: editTitle,
				description: editDescription || undefined,
				precondition: editPrecondition || undefined,
				postcondition: editPostcondition || undefined,
				priority: editPriority,
				status: editStatus,
				steps: editSteps.length > 0 ? stripStepUid(editSteps) : undefined,
				version: tc.version,
			});
		},
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["tms-case", projectId, caseKey],
			});
			void queryClient.invalidateQueries({
				queryKey: ["tms-cases", projectId],
			});
			setEditing(false);
			toast.success("Тест-кейс обновлён");
		},
		onError: (err) => toastApiError(err, "Ошибка обновления тест-кейса"),
	});

	// Привязка к задаче
	const linkMutation = useMutation({
		mutationFn: (issueId: string) =>
			linkTestCaseToIssue(projectId, caseKey, issueId),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["tms-case", projectId, caseKey],
			});
			setLinkIssueId("");
			toast.success("Задача привязана");
		},
		onError: (err) => toastApiError(err, "Ошибка привязки"),
	});

	// Отвязка от задачи
	const unlinkMutation = useMutation({
		mutationFn: (issueId: string) =>
			unlinkTestCaseFromIssue(projectId, caseKey, issueId),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["tms-case", projectId, caseKey],
			});
			toast.success("Задача отвязана");
		},
		onError: (err) => toastApiError(err, "Ошибка отвязки"),
	});

	// Откат версии
	const rollbackMutation = useMutation({
		mutationFn: (versionId: string) =>
			rollbackVersion(projectId, caseKey, versionId),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["tms-case", projectId, caseKey],
			});
			void queryClient.invalidateQueries({
				queryKey: ["tms-case-versions", projectId, caseKey],
			});
			toast.success("Тест-кейс откачен к версии");
		},
		onError: (err) => toastApiError(err, "Ошибка отката"),
	});

	/** Начать редактирование. */
	const startEditing = () => {
		if (!tc) return;
		setEditTitle(tc.title);
		setEditDescription(tc.description ?? "");
		setEditPrecondition(tc.precondition ?? "");
		setEditPostcondition(tc.postcondition ?? "");
		setEditPriority(tc.priority);
		setEditStatus(tc.status);
		setEditSteps(
			tc.steps ? tc.steps.map((s) => ({ ...s, _uid: makeStepUid() })) : [],
		);
		setEditing(true);
	};

	/** Отменить редактирование. */
	const cancelEditing = () => setEditing(false);

	/** Добавить шаг в конец списка. */
	const addStep = () => {
		setEditSteps([
			...editSteps,
			{
				_uid: makeStepUid(),
				order: editSteps.length + 1,
				action: "",
				expected_result: "",
			},
		]);
	};

	/** Удалить шаг и пересчитать order. */
	const removeStep = (index: number) => {
		const updated = editSteps.filter((_, i) => i !== index);
		setEditSteps(updated.map((s, i) => ({ ...s, order: i + 1 })));
	};

	/** Обновить поле шага. */
	const updateStep = (
		index: number,
		field: "action" | "expected_result",
		value: string,
	) => {
		setEditSteps((prev) =>
			prev.map((s, i) => (i === index ? { ...s, [field]: value } : s)),
		);
	};

	/** Переместить шаг и пересчитать order. */
	const moveStep = (index: number, direction: "up" | "down") => {
		setEditSteps((prev) => {
			const target = direction === "up" ? index - 1 : index + 1;
			if (target < 0 || target >= prev.length) return prev;
			const stepA = prev[index];
			const stepB = prev[target];
			if (!stepA || !stepB) return prev;
			const next = [...prev];
			next[index] = stepB;
			next[target] = stepA;
			return next.map((s, i) => ({ ...s, order: i + 1 }));
		});
	};

	if (isLoading) {
		return (
			<div className="mx-auto max-w-6xl animate-fade-in">
				<div className="mb-4 flex items-center gap-2">
					<div className="h-4 w-20 rounded animate-shimmer" />
					<div className="h-4 w-4 rounded animate-shimmer" />
					<div className="h-4 w-24 rounded animate-shimmer" />
				</div>
				<div className="flex gap-6">
					<div className="flex-1 space-y-4">
						<div className="h-8 w-48 rounded animate-shimmer" />
						<div className="h-6 w-96 rounded animate-shimmer" />
						<div className="h-64 w-full rounded-xl animate-shimmer" />
					</div>
					<div className="w-72 shrink-0">
						<div className="h-80 w-full rounded-xl animate-shimmer" />
					</div>
				</div>
			</div>
		);
	}

	if (isError) {
		return (
			<div
				role="alert"
				className="mx-auto flex max-w-6xl flex-col items-center justify-center gap-3 py-16 text-center"
			>
				<p className="text-[14px] font-semibold text-t1">
					Не удалось загрузить тест-кейс
				</p>
				<button
					type="button"
					onClick={() => void refetch()}
					className="h-9 px-3 rounded-[var(--r2)] border border-b1 text-[13px] text-t2 hover:text-t1"
				>
					Попробовать снова
				</button>
			</div>
		);
	}

	if (!tc) {
		return (
			<div className="mx-auto flex max-w-6xl items-center justify-center py-16">
				<p className="text-t2">Тест-кейс не найден</p>
			</div>
		);
	}

	// biome-ignore lint/style/noNonNullAssertion: PRIORITY_CONFIG.medium is always defined
	const priorityCfg = (PRIORITY_CONFIG[tc.priority] ?? PRIORITY_CONFIG.medium)!;
	// biome-ignore lint/style/noNonNullAssertion: STATUS_CONFIG.draft is always defined
	const statusCfg = (STATUS_CONFIG[tc.status] ?? STATUS_CONFIG.draft)!;
	const hasAutotest = !!(
		tc.full_name ||
		tc.class_name ||
		tc.external_id ||
		tc.automation_id
	);

	/** Конфиг вкладок детальной страницы. */
	const detailTabs = [
		{ id: "steps" as const, label: "Шаги", count: tc.steps?.length ?? 0 },
		{
			id: "links" as const,
			label: "Связанные задачи",
			count: tc.linked_issues?.length ?? 0,
		},
		{
			id: "autotest" as const,
			label: "Автотест",
			show: hasAutotest || tc.case_type === "bdd",
		},
	];

	return (
		<div className="mx-auto max-w-6xl animate-fade-in">
			{/* Хлебные крошки */}
			<div className="mb-4 flex items-center justify-between">
				<div className="flex items-center gap-1.5 text-sm">
					<Link
						to="/projects/$projectId/tms/cases"
						params={{ projectId }}
						className="text-t2 hover:text-t1 transition-colors"
					>
						Тест-кейсы
					</Link>
					<svg
						aria-hidden="true"
						className="h-3.5 w-3.5 text-t2/40"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2}
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M9 5l7 7-7 7"
						/>
					</svg>
					<span className="font-mono font-medium text-primary">
						{tc.case_key}
					</span>
				</div>
				<div className="flex gap-2">
					{editing ? (
						<>
							<Button
								variant="outline"
								size="sm"
								onClick={cancelEditing}
								disabled={updateMutation.isPending}
							>
								Отмена
							</Button>
							<Button
								size="sm"
								className="gradient-tms text-white"
								onClick={() => updateMutation.mutate()}
								disabled={!editTitle.trim() || updateMutation.isPending}
							>
								{updateMutation.isPending ? "Сохранение..." : "Сохранить"}
							</Button>
						</>
					) : canMutate ? (
						<Button variant="outline" size="sm" onClick={startEditing}>
							<svg
								aria-hidden="true"
								className="mr-1.5 h-3.5 w-3.5"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								strokeWidth={2}
							>
								<path
									strokeLinecap="round"
									strokeLinejoin="round"
									d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
								/>
							</svg>
							Редактировать
						</Button>
					) : null}
				</div>
			</div>

			<div className="flex gap-6">
				{/* Основная секция */}
				<div className="flex-1 min-w-0 space-y-5">
					{/* Заголовок кейса */}
					<div>
						<div className="mb-2 flex items-center gap-2.5 flex-wrap">
							<span className="font-mono text-lg font-bold text-primary">
								{tc.case_key}
							</span>
							{editing ? (
								<div className="flex items-center gap-2">
									<Select
										value={editPriority}
										onValueChange={(v) =>
											setEditPriority(v as TestCasePriority)
										}
									>
										<SelectTrigger className="rounded-xl border bg-card px-2 py-1 text-xs h-auto w-auto min-w-[100px]">
											<SelectValue />
										</SelectTrigger>
										<SelectContent>
											{Object.entries(PRIORITY_LABELS).map(([value, label]) => (
												<SelectItem key={value} value={value}>
													{label}
												</SelectItem>
											))}
										</SelectContent>
									</Select>
									<Select
										value={editStatus}
										onValueChange={(v) => setEditStatus(v as TestCaseStatus)}
									>
										<SelectTrigger className="rounded-xl border bg-card px-2 py-1 text-xs h-auto w-auto min-w-[100px]">
											<SelectValue />
										</SelectTrigger>
										<SelectContent>
											{Object.entries(STATUS_LABELS).map(([value, label]) => (
												<SelectItem key={value} value={value}>
													{label}
												</SelectItem>
											))}
										</SelectContent>
									</Select>
								</div>
							) : (
								<>
									<Badge
										variant="secondary"
										className={cn(
											"text-[10px] font-semibold px-2",
											statusCfg.bg,
											statusCfg.text,
										)}
									>
										{statusCfg.label}
									</Badge>
									<div className="flex items-center gap-1">
										<span
											className={cn("h-2 w-2 rounded-full", priorityCfg.dot)}
										/>
										<span
											className={cn("text-xs font-medium", priorityCfg.text)}
										>
											{priorityCfg.label}
										</span>
									</div>
									<Badge variant="outline" className="text-[10px] font-medium">
										{tc.case_type === "bdd" ? "BDD" : "Classic"}
									</Badge>
								</>
							)}
						</div>
						{editing ? (
							<Input
								className="text-xl font-semibold"
								value={editTitle}
								onChange={(e) => setEditTitle(e.target.value)}
								placeholder="Название тест-кейса"
							/>
						) : (
							<h2 className="text-xl font-semibold leading-tight">
								{tc.title}
							</h2>
						)}
						{editing ? (
							<textarea
								className="mt-2 w-full rounded-lg border bg-card px-3 py-2 text-sm"
								placeholder="Описание (необязательно)"
								rows={3}
								value={editDescription}
								onChange={(e) => setEditDescription(e.target.value)}
							/>
						) : (
							tc.description && (
								<p className="mt-2 text-sm text-t2 leading-relaxed">
									{tc.description}
								</p>
							)
						)}
					</div>

					{/* Предусловие */}
					{editing ? (
						<div className="rounded-lg border bg-card/50 p-3">
							<div className="text-xs font-semibold text-t2 uppercase tracking-wider mb-2">
								Предусловие
							</div>
							<textarea
								className="w-full rounded-md border bg-card px-3 py-2 text-sm"
								placeholder="Предусловие (необязательно)"
								rows={2}
								value={editPrecondition}
								onChange={(e) => setEditPrecondition(e.target.value)}
							/>
						</div>
					) : (
						tc.precondition && (
							<div className="tms-precond-card px-[18px] py-4 bg-gradient-to-br from-card to-raised border border-b1 rounded-[var(--r3)] text-[13px] text-t2 leading-[1.65] relative overflow-hidden">
								<strong className="text-t1 font-semibold">Предусловие:</strong>{" "}
								{tc.precondition}
							</div>
						)
					)}

					{/* Вкладки */}
					<div className="border-b">
						<div className="flex gap-0.5">
							{detailTabs.map((tab) => {
								if (tab.show === false) return null;
								return (
									<button
										key={tab.id}
										type="button"
										className={cn(
											"relative px-4 py-2 text-sm font-medium transition-colors duration-200 border-b-2",
											activeTab === tab.id
												? "border-primary text-primary"
												: "border-transparent text-t2 hover:text-t1 hover:border-t2/30",
										)}
										onClick={() => setActiveTab(tab.id)}
									>
										<span className="flex items-center gap-1.5">
											{tab.label}
											{"count" in tab && (tab.count ?? 0) > 0 && (
												<span
													className={cn(
														"text-[10px] font-semibold px-1.5 py-0 rounded-full",
														activeTab === tab.id
															? "bg-primary/10 text-primary"
															: "bg-raised text-t2",
													)}
												>
													{tab.count}
												</span>
											)}
										</span>
									</button>
								);
							})}
						</div>
					</div>

					{/* Вкладка: Steps */}
					{activeTab === "steps" && (
						<div className="space-y-4">
							{editing ? (
								<StepEditor
									editing={true}
									editSteps={editSteps}
									onAddStep={addStep}
									onRemoveStep={removeStep}
									onUpdateStep={updateStep}
									onMoveStep={moveStep}
								/>
							) : (
								<StepEditor
									editing={false}
									tc={tc}
									onStartEditing={canMutate ? startEditing : undefined}
								/>
							)}

							{/* Постусловие */}
							{editing ? (
								<div className="rounded-lg border bg-card/50 p-3">
									<div className="text-xs font-semibold text-t2 uppercase tracking-wider mb-2">
										Постусловие
									</div>
									<textarea
										className="w-full rounded-md border bg-card px-3 py-2 text-sm"
										placeholder="Постусловие (необязательно)"
										rows={2}
										value={editPostcondition}
										onChange={(e) => setEditPostcondition(e.target.value)}
									/>
								</div>
							) : (
								tc.postcondition && (
									<div className="tms-precond-card px-[18px] py-4 bg-gradient-to-br from-card to-raised border border-b1 rounded-[var(--r3)] text-[13px] text-t2 leading-[1.65] relative overflow-hidden">
										<strong className="text-t1 font-semibold">
											Постусловие:
										</strong>{" "}
										{tc.postcondition}
									</div>
								)
							)}
						</div>
					)}

					{/* Вкладка: Linked Issues */}
					{activeTab === "links" && (
						<div className="space-y-3">
							{tc.linked_issues && tc.linked_issues.length > 0 ? (
								<div className="rounded-xl border overflow-hidden shadow-card">
									<table className="w-full">
										<thead>
											<tr className="border-b bg-raised/30">
												<th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-t2">
													Ключ
												</th>
												<th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-t2">
													Название
												</th>
												<th className="w-10 px-3 py-2.5" />
											</tr>
										</thead>
										<tbody>
											{tc.linked_issues.map((issue) => (
												<tr
													key={issue.id}
													className="border-b last:border-0 hover:bg-raised/20 transition-colors"
												>
													<td className="px-3 py-2.5">
														<Link
															to="/projects/$projectId/issues/$issueKey"
															params={{ projectId, issueKey: issue.issue_key }}
															className="font-mono text-xs font-bold text-primary hover:underline"
														>
															{issue.issue_key}
														</Link>
													</td>
													<td className="px-3 py-2.5 text-sm">{issue.title}</td>
													<td className="px-3 py-2.5">
														{canMutate ? (
															<Button
																variant="ghost"
																size="sm"
																className="h-6 w-6 p-0 text-t2 hover:text-destructive"
																aria-label="Отвязать задачу"
																disabled={unlinkMutation.isPending}
																onClick={() => unlinkMutation.mutate(issue.id)}
															>
																<svg
																	aria-hidden="true"
																	className="h-3.5 w-3.5"
																	fill="none"
																	viewBox="0 0 24 24"
																	stroke="currentColor"
																	strokeWidth={2}
																>
																	<path
																		strokeLinecap="round"
																		strokeLinejoin="round"
																		d="M6 18L18 6M6 6l12 12"
																	/>
																</svg>
															</Button>
														) : null}
													</td>
												</tr>
											))}
										</tbody>
									</table>
								</div>
							) : (
								<div className="flex flex-col items-center justify-center py-12 text-center border rounded-xl border-dashed">
									<svg
										aria-hidden="true"
										className="h-10 w-10 text-t2/20 mb-3"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										strokeWidth={1.5}
									>
										<path
											strokeLinecap="round"
											strokeLinejoin="round"
											d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
										/>
									</svg>
									<p className="text-sm text-t2">Нет связанных задач</p>
								</div>
							)}

							{canMutate ? (
								<div className="flex gap-2">
									<Input
										placeholder="ID или ключ задачи"
										className="flex-1 h-8 text-sm"
										value={linkIssueId}
										onChange={(e) => setLinkIssueId(e.target.value)}
										onKeyDown={(e) => {
											if (e.key === "Enter" && linkIssueId.trim()) {
												linkMutation.mutate(linkIssueId.trim());
											}
										}}
									/>
									<Button
										variant="outline"
										size="sm"
										className="h-8"
										disabled={!linkIssueId.trim() || linkMutation.isPending}
										onClick={() => linkMutation.mutate(linkIssueId.trim())}
									>
										<svg
											aria-hidden="true"
											className="mr-1 h-3 w-3"
											fill="none"
											viewBox="0 0 24 24"
											stroke="currentColor"
											strokeWidth={2}
										>
											<path
												strokeLinecap="round"
												strokeLinejoin="round"
												d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
											/>
										</svg>
										Привязать
									</Button>
								</div>
							) : null}
						</div>
					)}

					{/* Вкладка: Autotest */}
					{activeTab === "autotest" && (
						<div className="space-y-4">
							{/* Бейдж статуса автоматизации */}
							<div className="flex items-center gap-2">
								{hasAutotest ? (
									<span className="inline-flex items-center gap-1.5 rounded-full bg-green-100 dark:bg-green-900/30 px-3 py-1 text-xs font-semibold text-green-700 dark:text-green-400">
										<svg
											aria-hidden="true"
											className="h-3.5 w-3.5"
											fill="currentColor"
											viewBox="0 0 16 16"
										>
											<path d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7.25 7.25a.75.75 0 0 1-1.06 0L2.22 9.28a.751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018L6 10.94l6.72-6.72a.75.75 0 0 1 1.06 0Z" />
										</svg>
										Автоматизирован
									</span>
								) : (
									<span className="inline-flex items-center gap-1.5 rounded-full bg-gray-100 dark:bg-gray-800 px-3 py-1 text-xs font-semibold text-gray-500 dark:text-gray-400">
										<svg
											aria-hidden="true"
											className="h-3.5 w-3.5"
											fill="currentColor"
											viewBox="0 0 16 16"
										>
											<path d="M2 7.75A.75.75 0 0 1 2.75 7h10a.75.75 0 0 1 0 1.5h-10A.75.75 0 0 1 2 7.75Z" />
										</svg>
										Не автоматизирован
									</span>
								)}
							</div>

							{/* ID для автотестов */}
							<div className="rounded-xl border overflow-hidden shadow-card">
								<div className="p-4 space-y-4">
									<div>
										<div className="text-[10px] font-bold text-t2 uppercase tracking-wider mb-1.5">
											ID для автотестов
										</div>
										<div className="flex items-center gap-2">
											<code className="flex-1 text-sm font-mono bg-raised/50 px-3 py-1.5 rounded border select-all">
												{tc.case_key}
											</code>
											<Button
												variant="outline"
												size="sm"
												className="h-8 px-2.5"
												onClick={() => {
													void navigator.clipboard.writeText(tc.case_key);
													toast.success("ID скопирован");
												}}
											>
												<svg
													aria-hidden="true"
													className="h-3.5 w-3.5"
													fill="none"
													viewBox="0 0 24 24"
													stroke="currentColor"
													strokeWidth={2}
												>
													<path
														strokeLinecap="round"
														strokeLinejoin="round"
														d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
													/>
												</svg>
											</Button>
										</div>
										<p className="mt-2 text-xs text-t2/70 leading-relaxed">
											Используйте этот ID в названии или аннотации автотеста для
											автоматической привязки результатов.
										</p>
									</div>

									{tc.automation_id && (
										<div>
											<div className="text-[10px] font-bold text-t2 uppercase tracking-wider mb-1">
												Automation ID
											</div>
											<div className="flex items-center gap-2">
												<code className="flex-1 text-sm font-mono bg-raised/50 px-2 py-1 rounded">
													{tc.automation_id}
												</code>
												<Button
													variant="ghost"
													size="sm"
													className="h-7 w-7 p-0"
													onClick={() => {
														if (!tc.automation_id) return;
														void navigator.clipboard.writeText(
															tc.automation_id,
														);
														toast.success("Automation ID скопирован");
													}}
												>
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
															d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
														/>
													</svg>
												</Button>
											</div>
										</div>
									)}
									{tc.full_name && (
										<div>
											<div className="text-[10px] font-bold text-t2 uppercase tracking-wider mb-1">
												Полное имя
											</div>
											<code className="text-sm font-mono bg-raised/50 px-2 py-1 rounded">
												{tc.full_name}
											</code>
										</div>
									)}
									{tc.class_name && (
										<div>
											<div className="text-[10px] font-bold text-t2 uppercase tracking-wider mb-1">
												Класс
											</div>
											<code className="text-sm font-mono bg-raised/50 px-2 py-1 rounded">
												{tc.class_name}
											</code>
										</div>
									)}
									{tc.method_name && (
										<div>
											<div className="text-[10px] font-bold text-t2 uppercase tracking-wider mb-1">
												Метод
											</div>
											<code className="text-sm font-mono bg-raised/50 px-2 py-1 rounded">
												{tc.method_name}
											</code>
										</div>
									)}
									{tc.external_id && (
										<div>
											<div className="text-[10px] font-bold text-t2 uppercase tracking-wider mb-1">
												External ID
											</div>
											<code className="text-sm font-mono bg-raised/50 px-2 py-1 rounded">
												{tc.external_id}
											</code>
										</div>
									)}
								</div>
							</div>
						</div>
					)}
				</div>

				{/* Правый sidebar — метаданные */}
				<div className="w-72 shrink-0 space-y-4">
					<TestCaseMetadata tc={tc} />
					<VersionHistory
						versionsOpen={versionsOpen}
						onToggle={() => setVersionsOpen(!versionsOpen)}
						versions={versions}
						onRollback={(versionId) => rollbackMutation.mutate(versionId)}
						isRollingBack={rollbackMutation.isPending}
						canRollback={canRollback}
					/>
				</div>
			</div>
		</div>
	);
}

export const Route = createLazyFileRoute(
	"/_authenticated/projects/$projectId/tms/cases_/$caseKey",
)({
	component: TestCaseDetailPage,
});
