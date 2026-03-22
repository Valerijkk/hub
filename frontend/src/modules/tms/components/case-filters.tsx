/**
 * Панель фильтров для списка тест-кейсов.
 */

import type { ListTestCasesParams } from "@/api/tms";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

/** Пропсы панели фильтров тест-кейсов. */
interface CaseFiltersProps {
	filters: {
		search?: string;
		status?: string;
		priority?: string;
		suiteId?: string;
	};
	updateFilter: (patch: Record<string, unknown>) => void;
	resetFilters: () => void;
	resetOffset: () => void;
	selectedCount: number;
	onClearSelection: () => void;
	total: number;
	onCreateClick: () => void;
	createButton: React.ReactNode;
}

/** Панель фильтров: поиск, статус, приоритет, bulk-действия и кнопка создания. */
export function CaseFilters({
	filters,
	updateFilter,
	resetFilters,
	resetOffset,
	selectedCount,
	onClearSelection,
	total,
	createButton,
}: CaseFiltersProps) {
	const hasActiveFilters = !!(
		filters.search ||
		filters.status ||
		filters.priority ||
		filters.suiteId
	);

	return (
		<div className="mb-3 flex flex-wrap items-center gap-2">
			<div className="relative">
				<svg
					className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground/60"
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
					className="w-56 pl-8 h-8 text-sm"
					value={filters.search ?? ""}
					onChange={(e) => {
						updateFilter({ search: e.target.value || undefined });
						resetOffset();
					}}
				/>
			</div>

			<select
				className="rounded-lg border bg-card px-2.5 py-1.5 text-xs h-8"
				value={filters.status ?? ""}
				onChange={(e) => {
					updateFilter({
						status: (e.target.value ||
							undefined) as ListTestCasesParams["status"],
					});
					resetOffset();
				}}
			>
				<option value="">Все статусы</option>
				<option value="draft">Черновик</option>
				<option value="active">Активный</option>
				<option value="deprecated">Устаревший</option>
			</select>

			<select
				className="rounded-lg border bg-card px-2.5 py-1.5 text-xs h-8"
				value={filters.priority ?? ""}
				onChange={(e) => {
					updateFilter({
						priority: (e.target.value ||
							undefined) as ListTestCasesParams["priority"],
					});
					resetOffset();
				}}
			>
				<option value="">Все приоритеты</option>
				<option value="critical">Критический</option>
				<option value="high">Высокий</option>
				<option value="medium">Средний</option>
				<option value="low">Низкий</option>
				<option value="trivial">Тривиальный</option>
			</select>

			{hasActiveFilters && (
				<Button
					variant="ghost"
					size="sm"
					className="h-8 text-xs"
					onClick={() => {
						resetFilters();
						resetOffset();
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

			{/* Bulk-панель при выбранных кейсах */}
			{selectedCount > 0 && (
				<div className="flex items-center gap-2 ml-2 px-3 py-1 rounded-lg bg-primary/10 border border-primary/20">
					<span className="text-xs font-medium text-primary">
						Выбрано: {selectedCount}
					</span>
					<Button
						variant="ghost"
						size="sm"
						className="h-6 text-xs px-2"
						onClick={onClearSelection}
					>
						Снять
					</Button>
				</div>
			)}

			<div className="ml-auto flex items-center gap-2">
				<span className="text-xs text-muted-foreground tabular-nums">
					{total} {total === 1 ? "кейс" : total < 5 ? "кейса" : "кейсов"}
				</span>
				{createButton}
			</div>
		</div>
	);
}
