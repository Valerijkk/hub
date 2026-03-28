import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createLazyFileRoute, Link } from "@tanstack/react-router";
import { useCallback, useState } from "react";
import { toast } from "sonner";
import {
	createTestCase,
	type ListTestCasesParams,
	listSuites,
	listTestCases,
	type SuiteResponse,
	type TestCaseResponse,
	type TestCaseType,
} from "@/api/tms";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { formatRelativeTime } from "@/lib/format-time";
import { cn } from "@/lib/utils";
import { useTmsStore } from "@/stores/tms-store";

/** Color config for test case priorities. */
const PRIORITY_CONFIG: Record<
	string,
	{ dot: string; bg: string; text: string; label: string }
> = {
	critical: {
		dot: "bg-red",
		bg: "bg-red/[0.08]",
		text: "text-red",
		label: "Critical",
	},
	high: {
		dot: "bg-orange",
		bg: "bg-orange/[0.08]",
		text: "text-orange",
		label: "High",
	},
	medium: {
		dot: "bg-amber",
		bg: "bg-amber/[0.08]",
		text: "text-amber",
		label: "Medium",
	},
	low: {
		dot: "bg-blue",
		bg: "bg-blue/[0.08]",
		text: "text-blue",
		label: "Low",
	},
	trivial: {
		dot: "bg-gray",
		bg: "bg-gray/[0.08]",
		text: "text-gray",
		label: "Trivial",
	},
};

/** Test case status config. */
const STATUS_CONFIG: Record<
	string,
	{ label: string; bg: string; text: string }
> = {
	draft: {
		label: "Черновик",
		bg: "bg-gray/[0.08]",
		text: "text-gray",
	},
	active: {
		label: "Активный",
		bg: "bg-green/[0.08]",
		text: "text-green",
	},
	deprecated: {
		label: "Устаревший",
		bg: "bg-red/[0.08]",
		text: "text-red",
	},
};

/**
 * Recursive suite tree item component.
 */
function SuiteTreeItem({
	suite,
	selectedId,
	onSelect,
	depth = 0,
}: {
	suite: SuiteResponse;
	selectedId?: string;
	onSelect: (id: string | undefined) => void;
	depth?: number;
}) {
	const [expanded, setExpanded] = useState(depth < 2);
	const hasChildren = suite.children && suite.children.length > 0;
	const isSelected = selectedId === suite.id;

	return (
		<div>
			<button
				className={cn(
					"group flex w-full items-center gap-2 rounded-[var(--r2)] px-[10px] py-2 text-left text-[13px] font-medium cursor-pointer transition-colors duration-[180ms] text-t2 relative",
					"hover:bg-[var(--hover)] hover:text-t1",
					isSelected && "bg-purple/[0.05] text-white shadow-[0_0_0_1px_rgba(167,139,250,0.08)_inset]",
				)}
				style={{ paddingLeft: `${depth * 26 + 10}px` }}
				onClick={() => onSelect(isSelected ? undefined : suite.id)}
			>
				{isSelected && (
					<span className="absolute left-0 top-0 bottom-0 w-10 bg-gradient-to-r from-purple/[0.06] to-transparent rounded-[var(--r2)] pointer-events-none" />
				)}
				{hasChildren ? (
					<span
						className={cn(
							"w-4 h-4 text-[8px] text-t4 transition-transform duration-[250ms] shrink-0 flex items-center justify-center",
							expanded && "rotate-90",
						)}
						onClick={(e) => {
							e.stopPropagation();
							setExpanded(!expanded);
						}}
					>
						<svg
							className="h-[10px] w-[10px]"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={3}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M9 5l7 7-7 7"
							/>
						</svg>
					</span>
				) : (
					<span className="w-4 shrink-0" style={{ visibility: "hidden" }} />
				)}
				{hasChildren ? (
					<span
						className={cn(
							"shrink-0 flex items-center justify-center w-[22px] h-[22px] rounded-[7px] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200",
							"group-hover:scale-110 group-hover:shadow-[0_0_10px_rgba(167,139,250,0.1)]",
						)}
						style={{
							background: isSelected ? "rgba(167,139,250,0.1)" : "rgba(255,255,255,0.035)",
							color: isSelected ? "var(--purple)" : "var(--text-3)",
						}}
					>
						<svg
							className="h-3 w-3"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
							/>
						</svg>
					</span>
				) : (
					<span
						className="w-[7px] h-[7px] rounded-full shrink-0 ml-[1px] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 group-hover:shadow-[0_0_8px_currentColor] group-hover:scale-[1.3]"
						style={{
							background: isSelected ? "var(--purple)" : "var(--text-3)",
							color: isSelected ? "var(--purple)" : "var(--text-3)",
						}}
					/>
				)}
				<span className="truncate flex-1">{suite.name}</span>
			</button>
			{expanded && hasChildren && (
				<div>
					{suite.children?.map((child) => (
						<SuiteTreeItem
							key={child.id}
							suite={child}
							selectedId={selectedId}
							onSelect={onSelect}
							depth={depth + 1}
						/>
					))}
				</div>
			)}
		</div>
	);
}

/**
 * Test cases list page -- Allure TestOps level design.
 */
function TestCasesPage() {
	const { projectId } = Route.useParams();
	const queryClient = useQueryClient();
	const { filters, updateFilter, resetFilters } = useTmsStore();
	const [offset, setOffset] = useState(0);
	const [createOpen, setCreateOpen] = useState(false);
	const [newTitle, setNewTitle] = useState("");
	const [newType, setNewType] = useState<TestCaseType>("classic");
	const [newPriority, setNewPriority] = useState("medium");
	const [newSuiteId, setNewSuiteId] = useState("");
	const [newDescription, setNewDescription] = useState("");
	const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
	const limit = 25;

	const params: ListTestCasesParams = {
		offset,
		limit,
		suite_id: filters.suiteId,
		status: filters.status,
		priority: filters.priority,
		search: filters.search,
	};

	const { data, isLoading } = useQuery({
		queryKey: ["tms-cases", projectId, params],
		queryFn: () => listTestCases(projectId, params),
	});

	const { data: suites } = useQuery({
		queryKey: ["tms-suites", projectId],
		queryFn: () => listSuites(projectId),
	});

	const createMutation = useMutation({
		mutationFn: () =>
			createTestCase(projectId, {
				title: newTitle,
				case_type: newType,
				priority: newPriority as TestCaseResponse["priority"],
				suite_id: newSuiteId || undefined,
				description: newDescription || undefined,
			}),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["tms-cases", projectId],
			});
			setCreateOpen(false);
			setNewTitle("");
			setNewDescription("");
			toast.success("Тест-кейс создан");
		},
		onError: () => toast.error("Ошибка создания тест-кейса"),
	});

	const handleSuiteSelect = useCallback(
		(id: string | undefined) => {
			updateFilter({ suiteId: id });
			setOffset(0);
		},
		[updateFilter],
	);

	const cases = data?.items ?? [];
	const total = data?.total ?? 0;
	const totalPages = Math.ceil(total / limit);
	const currentPage = Math.floor(offset / limit) + 1;

	const toggleSelect = useCallback((id: string) => {
		setSelectedIds((prev) => {
			const next = new Set(prev);
			if (next.has(id)) next.delete(id);
			else next.add(id);
			return next;
		});
	}, []);

	const toggleSelectAll = useCallback(() => {
		if (selectedIds.size === cases.length && cases.length > 0) {
			setSelectedIds(new Set());
		} else {
			setSelectedIds(new Set(cases.map((c) => c.id)));
		}
	}, [cases, selectedIds.size]);

	const hasActiveFilters = !!(
		filters.search ||
		filters.status ||
		filters.priority ||
		filters.suiteId
	);

	return (
		<div className="flex flex-col md:flex-row flex-1 overflow-hidden">
			{/* Suite tree sidebar */}
			<div className="w-full md:w-[290px] md:min-w-[290px] border-r border-b1 flex flex-col bg-void animate-[enterRight_400ms_var(--spring)_both]">
				<div className="px-[18px] py-4 border-b border-b1 flex items-center justify-between">
					<span className="text-[10.5px] font-bold uppercase tracking-[0.08em] text-t4">
						Тест-сьюты
					</span>
				</div>
				<div className="flex-1 overflow-y-auto p-2">
					<button
						className={cn(
							"group flex w-full items-center gap-2 rounded-[var(--r2)] px-[10px] py-2 text-left text-[13px] font-medium cursor-pointer transition-colors duration-[180ms] text-t2 relative mb-1",
							"hover:bg-[var(--hover)] hover:text-t1",
							!filters.suiteId && "bg-purple/[0.05] text-white shadow-[0_0_0_1px_rgba(167,139,250,0.08)_inset]",
						)}
						onClick={() => handleSuiteSelect(undefined)}
					>
						{!filters.suiteId && (
							<span className="absolute left-0 top-0 bottom-0 w-10 bg-gradient-to-r from-purple/[0.06] to-transparent rounded-[var(--r2)] pointer-events-none" />
						)}
						<svg
							className="h-3.5 w-3.5 shrink-0"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
							/>
						</svg>
						<span className="flex-1">Все кейсы</span>
						<span className="text-[10.5px] font-semibold text-t4 bg-white/[0.035] px-2 py-[1px] rounded-full transition-colors duration-200 group-hover:bg-white/[0.06] group-hover:text-t3">
							{total}
						</span>
					</button>
					{suites?.map((suite) => (
						<SuiteTreeItem
							key={suite.id}
							suite={suite}
							selectedId={filters.suiteId}
							onSelect={handleSuiteSelect}
						/>
					))}
				</div>
			</div>

			{/* Case table panel */}
			<div className="flex-1 flex flex-col overflow-hidden">
				{/* Header */}
				<div className="px-6 py-4 border-b border-b1 flex items-center gap-[14px] bg-[rgba(13,14,18,0.5)] backdrop-blur-[12px]">
					<h2 className="text-[15px] font-bold tracking-[-0.015em]">Тест-кейсы</h2>
					<span className="text-[11.5px] font-semibold text-t3 bg-white/[0.035] px-3 py-[3px] rounded-full">
						{total} {total === 1 ? "кейс" : total < 5 ? "кейса" : "кейсов"}
					</span>
					<div className="flex-1" />

					{/* Filters */}
					<div className="relative">
						<svg
							className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-t4"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
							/>
						</svg>
						<Input
							placeholder="Поиск тест-кейсов..."
							className="w-56 pl-8 h-8 text-sm bg-white/[0.03] border-b1 rounded-[var(--r2)]"
							value={filters.search ?? ""}
							onChange={(e) => {
								updateFilter({ search: e.target.value || undefined });
								setOffset(0);
							}}
						/>
					</div>

					<Select
						value={filters.status ?? "all"}
						onValueChange={(v) => {
							updateFilter({ status: (v === "all" ? undefined : v) as ListTestCasesParams["status"] });
							setOffset(0);
						}}
					>
						<SelectTrigger className="w-[140px] h-8 text-xs rounded-[var(--r2)] border-b1 bg-white/[0.03]">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="all">Все статусы</SelectItem>
							<SelectItem value="draft">Черновик</SelectItem>
							<SelectItem value="active">Активный</SelectItem>
							<SelectItem value="deprecated">Устаревший</SelectItem>
						</SelectContent>
					</Select>

					<Select
						value={filters.priority ?? "all"}
						onValueChange={(v) => {
							updateFilter({ priority: (v === "all" ? undefined : v) as ListTestCasesParams["priority"] });
							setOffset(0);
						}}
					>
						<SelectTrigger className="w-[160px] h-8 text-xs rounded-[var(--r2)] border-b1 bg-white/[0.03]">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="all">Все приоритеты</SelectItem>
							<SelectItem value="critical">Критический</SelectItem>
							<SelectItem value="high">Высокий</SelectItem>
							<SelectItem value="medium">Средний</SelectItem>
							<SelectItem value="low">Низкий</SelectItem>
							<SelectItem value="trivial">Тривиальный</SelectItem>
						</SelectContent>
					</Select>

					{hasActiveFilters && (
						<Button
							variant="ghost"
							size="sm"
							className="h-8 text-xs text-t3 hover:text-t1"
							onClick={() => {
								resetFilters();
								setOffset(0);
							}}
						>
							<svg
								className="mr-1 h-3 w-3"
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
							Сбросить
						</Button>
					)}

					{/* Bulk panel */}
					{selectedIds.size > 0 && (
						<div className="flex items-center gap-2 px-3 py-1 rounded-[var(--r2)] bg-purple/[0.06] border border-purple/[0.15]">
							<span className="text-xs font-medium text-purple">
								Выбрано: {selectedIds.size}
							</span>
							<Button
								variant="ghost"
								size="sm"
								className="h-6 text-xs px-2 text-t3"
								onClick={() => setSelectedIds(new Set())}
							>
								Снять
							</Button>
						</div>
					)}

					<Dialog open={createOpen} onOpenChange={setCreateOpen}>
						<DialogTrigger asChild>
							<Button
								size="sm"
								className="gradient-tms text-white shadow-sm h-8"
							>
								<svg
									className="mr-1.5 h-3.5 w-3.5"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									strokeWidth={2}
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d="M12 4v16m8-8H4"
									/>
								</svg>
								Создать кейс
							</Button>
						</DialogTrigger>
						<DialogContent>
							<DialogHeader>
								<DialogTitle>Новый тест-кейс</DialogTitle>
							</DialogHeader>
							<div className="space-y-4">
								<Input
									placeholder="Название тест-кейса"
									value={newTitle}
									onChange={(e) => setNewTitle(e.target.value)}
								/>
								<div className="grid grid-cols-2 gap-3">
									<Select
										value={newType}
										onValueChange={(v) =>
											setNewType(v as TestCaseType)
										}
									>
										<SelectTrigger className="w-full rounded-xl border bg-card px-3 py-2 text-sm">
											<SelectValue />
										</SelectTrigger>
										<SelectContent>
											<SelectItem value="classic">Classic</SelectItem>
											<SelectItem value="bdd">BDD</SelectItem>
										</SelectContent>
									</Select>
									<Select
										value={newPriority}
										onValueChange={(v) => setNewPriority(v)}
									>
										<SelectTrigger className="w-full rounded-xl border bg-card px-3 py-2 text-sm">
											<SelectValue />
										</SelectTrigger>
										<SelectContent>
											<SelectItem value="critical">Критический</SelectItem>
											<SelectItem value="high">Высокий</SelectItem>
											<SelectItem value="medium">Средний</SelectItem>
											<SelectItem value="low">Низкий</SelectItem>
											<SelectItem value="trivial">Тривиальный</SelectItem>
										</SelectContent>
									</Select>
								</div>
								{suites && suites.length > 0 && (
									<Select
										value={newSuiteId || "none"}
										onValueChange={(v) => setNewSuiteId(v === "none" ? "" : v)}
									>
										<SelectTrigger className="w-full rounded-xl border bg-card px-3 py-2 text-sm">
											<SelectValue />
										</SelectTrigger>
										<SelectContent>
											<SelectItem value="none">Без папки</SelectItem>
											{suites.map((s) => (
												<SelectItem key={s.id} value={s.id}>
													{s.name}
												</SelectItem>
											))}
										</SelectContent>
									</Select>
								)}
								<textarea
									className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
									placeholder="Описание (необязательно)"
									rows={3}
									value={newDescription}
									onChange={(e) => setNewDescription(e.target.value)}
								/>
								<Button
									className="w-full"
									disabled={!newTitle.trim() || createMutation.isPending}
									onClick={() => createMutation.mutate()}
								>
									{createMutation.isPending ? "Создание..." : "Создать"}
								</Button>
							</div>
						</DialogContent>
					</Dialog>
				</div>

				{/* Test cases table */}
				{isLoading ? (
					<div className="flex-1 overflow-y-auto">
						<table className="w-full border-separate border-spacing-0">
							<thead>
								<tr>
									<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-5 py-[14px] border-b border-b2" colSpan={8}>
										<div className="h-4 w-full rounded animate-shimmer" />
									</th>
								</tr>
							</thead>
							<tbody>
								{[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
									<tr key={i}>
										<td className="px-5 h-14 border-b border-b1" colSpan={8}>
											<div className="flex items-center gap-4">
												<div className="h-4 w-4 rounded animate-shimmer" />
												<div className="h-3.5 w-16 rounded animate-shimmer" />
												<div className="h-3.5 w-48 rounded animate-shimmer flex-1" />
												<div className="h-5 w-14 rounded-full animate-shimmer" />
												<div className="h-5 w-20 rounded-full animate-shimmer" />
											</div>
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				) : cases.length === 0 ? (
					<div className="flex flex-col items-center justify-center py-20 flex-1 animate-[enterUp_450ms_var(--spring)_both]">
						<div className="mx-auto mb-5 flex h-24 w-24 items-center justify-center rounded-full bg-gradient-to-br from-blue/[0.06] to-purple/[0.06] shadow-[0_0_60px_rgba(167,139,250,0.06)] animate-[breathe_4s_ease-in-out_infinite]">
							{filters.suiteId &&
							!filters.search &&
							!filters.status &&
							!filters.priority ? (
								<svg
									className="h-10 w-10 text-t4"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									strokeWidth={1.5}
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
									/>
								</svg>
							) : (
								<svg
									className="h-10 w-10 text-t4"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									strokeWidth={1.5}
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
									/>
								</svg>
							)}
						</div>
						<h3 className="text-[19px] font-bold mb-2 tracking-[-0.02em]">
							{filters.suiteId &&
							!filters.search &&
							!filters.status &&
							!filters.priority
								? "В этой папке пока нет тест-кейсов"
								: "Нет тест-кейсов"}
						</h3>
						<p className="text-[14px] text-t2 mb-6 max-w-[340px] text-center leading-[1.65]">
							{filters.suiteId &&
							!filters.search &&
							!filters.status &&
							!filters.priority
								? "Создайте новый тест-кейс в выбранной папке или перенесите существующие"
								: hasActiveFilters
									? "По вашим фильтрам ничего не найдено. Попробуйте изменить параметры поиска."
									: "Создайте первый тест-кейс, чтобы начать управление тестированием"}
						</p>
						{hasActiveFilters ? (
							<div className="flex gap-2">
								<Button
									variant="outline"
									onClick={() => {
										resetFilters();
										setOffset(0);
									}}
								>
									Сбросить фильтры
								</Button>
								{filters.suiteId && (
									<Button
										className="gradient-tms text-white"
										onClick={() => setCreateOpen(true)}
									>
										Создать кейс
									</Button>
								)}
							</div>
						) : (
							<Button
								className="gradient-tms text-white"
								onClick={() => setCreateOpen(true)}
							>
								Создать первый кейс
							</Button>
						)}
					</div>
				) : (
					<>
						<div className="flex-1 overflow-y-auto">
							<table className="w-full border-separate border-spacing-0">
								<thead>
									<tr>
										<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-5 py-[14px] border-b border-b2 w-10">
											<input
												type="checkbox"
												className="rounded border-b2"
												checked={
													selectedIds.size === cases.length && cases.length > 0
												}
												onChange={toggleSelectAll}
											/>
										</th>
										<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-5 py-[14px] border-b border-b2" style={{ width: 82 }}>
											Ключ
										</th>
										<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-5 py-[14px] border-b border-b2">
											Название
										</th>
										<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-5 py-[14px] border-b border-b2" style={{ width: 100 }}>
											Приоритет
										</th>
										<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-5 py-[14px] border-b border-b2" style={{ width: 110 }}>
											Статус
										</th>
										<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-5 py-[14px] border-b border-b2" style={{ width: 44 }}>
										</th>
										<th className="sticky top-0 z-[5] bg-base text-right text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-5 py-[14px] border-b border-b2">
											Обновлено
										</th>
									</tr>
								</thead>
								<tbody>
									{cases.map((tc, index) => {
										const priority = (PRIORITY_CONFIG[tc.priority] ??
											PRIORITY_CONFIG.medium)!;
										const statusCfg = (STATUS_CONFIG[tc.status] ??
											STATUS_CONFIG.draft)!;
										const isSelected = selectedIds.has(tc.id);
										const hasAutotest = !!(
											tc.full_name ||
											tc.class_name ||
											tc.external_id
										);

										return (
											<tr
												key={tc.id}
												className={cn(
													"group cursor-pointer",
													isSelected && "bg-purple/[0.03]",
												)}
												style={{
													animation: `enterUp 350ms var(--spring) both`,
													animationDelay: `${index * 40}ms`,
												}}
											>
												<td className="px-5 h-14 border-b border-b1 align-middle transition-colors duration-[180ms] group-hover:bg-[var(--hover)]">
													<input
														type="checkbox"
														className="rounded border-b2"
														checked={isSelected}
														onChange={() => toggleSelect(tc.id)}
													/>
												</td>
												<td className="px-5 h-14 border-b border-b1 align-middle transition-colors duration-[180ms] group-hover:bg-[var(--hover)] group-first:group-hover:shadow-[-3px_0_0_var(--purple)_inset]">
													<Link
														to="/projects/$projectId/tms/cases/$caseKey"
														params={{ projectId, caseKey: tc.case_key }}
														className="font-mono text-[12px] font-medium text-t3 hover:underline whitespace-nowrap"
													>
														{tc.case_key}
													</Link>
												</td>
												<td className="px-5 h-14 border-b border-b1 align-middle transition-colors duration-[180ms] group-hover:bg-[var(--hover)] max-w-md">
													<Link
														to="/projects/$projectId/tms/cases/$caseKey"
														params={{ projectId, caseKey: tc.case_key }}
														className="text-[13.5px] font-medium text-t1 transition-colors duration-150 group-hover:text-white line-clamp-1"
													>
														{tc.title}
													</Link>
												</td>
												<td className="px-5 h-14 border-b border-b1 align-middle transition-colors duration-[180ms] group-hover:bg-[var(--hover)]">
													<span className="inline-flex items-center gap-[7px] text-[12px] font-medium text-t2">
														<span
															className={cn(
																"w-2 h-2 rounded-[3px] transition-[box-shadow,transform] duration-200 group-hover:shadow-[0_0_8px_currentColor] group-hover:scale-[1.2]",
															)}
															style={{
																background: `var(--${tc.priority === "critical" ? "red" : tc.priority === "high" ? "orange" : tc.priority === "medium" ? "amber" : tc.priority === "low" ? "blue" : "gray"})`,
																color: `var(--${tc.priority === "critical" ? "red" : tc.priority === "high" ? "orange" : tc.priority === "medium" ? "amber" : tc.priority === "low" ? "blue" : "gray"})`,
															}}
														/>
														{priority.label}
													</span>
												</td>
												<td className="px-5 h-14 border-b border-b1 align-middle transition-colors duration-[180ms] group-hover:bg-[var(--hover)]">
													<Badge
														variant="secondary"
														className={cn(
															"text-[11px] font-semibold",
															statusCfg.bg,
															statusCfg.text,
															"border",
															tc.status === "active" ? "border-green/[0.15]" : tc.status === "deprecated" ? "border-red/[0.15]" : "border-gray/[0.15]",
														)}
													>
														{statusCfg.label}
													</Badge>
												</td>
												<td className="px-5 h-14 border-b border-b1 align-middle transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-[180ms] group-hover:bg-[var(--hover)]">
													{hasAutotest ? (
														<span
															className="text-[16px] opacity-[0.55] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 group-hover:opacity-100 group-hover:scale-[1.15] inline-block"
															title="Автоматизирован"
														>
															🤖
														</span>
													) : (
														<span
															className="text-[16px] opacity-[0.25] inline-block"
															title="Ручной"
														>
															📝
														</span>
													)}
												</td>
												<td className="px-5 h-14 border-b border-b1 align-middle transition-colors duration-[180ms] group-hover:bg-[var(--hover)] text-right">
													<span className="text-[11px] text-t3 whitespace-nowrap">
														{formatRelativeTime(tc.updated_at)}
													</span>
												</td>
											</tr>
										);
									})}
								</tbody>
							</table>
						</div>

						{/* Pagination */}
						<div className="px-6 py-3 border-t border-b1 flex items-center justify-between bg-[rgba(13,14,18,0.5)]">
							<span className="text-[12px] text-t3 tabular-nums">
								Показано {offset + 1}--{Math.min(offset + limit, total)} из{" "}
								{total}
							</span>
							<div className="flex items-center gap-1">
								<Button
									variant="outline"
									size="sm"
									className="h-7 w-7 p-0"
									disabled={offset === 0}
									onClick={() => setOffset(0)}
								>
									<svg
										className="h-3.5 w-3.5"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										strokeWidth={2}
									>
										<path
											strokeLinecap="round"
											strokeLinejoin="round"
											d="M11 19l-7-7 7-7m8 14l-7-7 7-7"
										/>
									</svg>
								</Button>
								<Button
									variant="outline"
									size="sm"
									className="h-7 w-7 p-0"
									disabled={offset === 0}
									onClick={() => setOffset(Math.max(0, offset - limit))}
								>
									<svg
										className="h-3.5 w-3.5"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										strokeWidth={2}
									>
										<path
											strokeLinecap="round"
											strokeLinejoin="round"
											d="M15 19l-7-7 7-7"
										/>
									</svg>
								</Button>
								<span className="px-2 text-xs text-t3 tabular-nums">
									{currentPage} / {totalPages || 1}
								</span>
								<Button
									variant="outline"
									size="sm"
									className="h-7 w-7 p-0"
									disabled={offset + limit >= total}
									onClick={() => setOffset(offset + limit)}
								>
									<svg
										className="h-3.5 w-3.5"
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
								</Button>
								<Button
									variant="outline"
									size="sm"
									className="h-7 w-7 p-0"
									disabled={offset + limit >= total}
									onClick={() =>
										setOffset(Math.max(0, (totalPages - 1) * limit))
									}
								>
									<svg
										className="h-3.5 w-3.5"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										strokeWidth={2}
									>
										<path
											strokeLinecap="round"
											strokeLinejoin="round"
											d="M13 5l7 7-7 7M5 5l7 7-7 7"
										/>
									</svg>
								</Button>
							</div>
						</div>
					</>
				)}
			</div>
		</div>
	);
}

export const Route = createLazyFileRoute(
	"/_authenticated/projects/$projectId/tms/cases",
)({
	component: TestCasesPage,
});
