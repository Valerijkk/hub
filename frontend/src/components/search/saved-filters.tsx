import { Filter, Plus, Share2, Trash2 } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import {
	createSavedFilter,
	deleteSavedFilter,
	listSavedFilters,
	type SavedFilter,
	shareFilter,
} from "@/api/search";
import { QueryErrorAlert } from "@/components/query-state-view";
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
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { toastApiError } from "@/lib/format-api-error";

interface SavedFiltersProps {
	/** ID проекта для фильтрации. */
	projectId: string;
	/** Текущая конфигурация фильтра трекера (для сохранения). */
	currentFilterConfig?: Record<string, unknown>;
	/** Callback при применении сохранённого фильтра. */
	onApplyFilter?: (config: Record<string, unknown>) => void;
}

/**
 * Компонент управления сохранёнными фильтрами трекера.
 * Позволяет создавать, удалять, расшаривать и применять фильтры.
 */
export function SavedFilters({
	projectId,
	currentFilterConfig,
	onApplyFilter,
}: SavedFiltersProps) {
	const [filters, setFilters] = useState<SavedFilter[]>([]);
	const [newFilterName, setNewFilterName] = useState("");
	const [dialogOpen, setDialogOpen] = useState(false);
	const [loading, setLoading] = useState(false);
	const [loadError, setLoadError] = useState(false);

	/** Загрузка списка фильтров. */
	const loadFilters = useCallback(async () => {
		try {
			const response = await listSavedFilters(projectId);
			setFilters(response.filters);
			setLoadError(false);
		} catch (err) {
			setLoadError(true);
			toastApiError(err, "Не удалось загрузить сохранённые фильтры");
		}
	}, [projectId]);

	useEffect(() => {
		void loadFilters();
	}, [loadFilters]);

	/** Создание нового фильтра. */
	const handleCreate = async () => {
		if (!newFilterName.trim() || !currentFilterConfig) return;
		setLoading(true);
		try {
			await createSavedFilter({
				name: newFilterName.trim(),
				filter_config: currentFilterConfig,
				project_id: projectId,
			});
			setNewFilterName("");
			setDialogOpen(false);
			await loadFilters();
		} catch (err) {
			toastApiError(err, "Не удалось сохранить фильтр");
		} finally {
			setLoading(false);
		}
	};

	/** Удаление фильтра. */
	const handleDelete = async (id: string) => {
		try {
			await deleteSavedFilter(id);
			setFilters((prev) => prev.filter((f) => f.id !== id));
		} catch (err) {
			toastApiError(err, "Не удалось удалить фильтр");
		}
	};

	/** Переключение расшаривания. */
	const handleToggleShare = async (filter: SavedFilter) => {
		try {
			const updated = await shareFilter(filter.id, !filter.shared);
			setFilters((prev) =>
				prev.map((f) => (f.id === updated.id ? updated : f)),
			);
		} catch (err) {
			toastApiError(err, "Не удалось изменить доступ к фильтру");
		}
	};

	if (loadError) {
		return (
			<QueryErrorAlert
				compact
				message="Не удалось загрузить сохранённые фильтры"
				retry={() => void loadFilters()}
			/>
		);
	}

	return (
		<div className="flex items-center gap-2">
			{/* Список сохранённых фильтров */}
			{filters.map((filter) => (
				<div
					key={filter.id}
					className="flex items-center gap-1 rounded-md border bg-base px-2 py-1"
				>
					<button
						type="button"
						className="text-xs font-medium hover:text-primary"
						onClick={() => onApplyFilter?.(filter.filter_config)}
					>
						{filter.name}
					</button>
					{filter.shared && <Share2 className="h-3 w-3 text-t2" />}
					<Tooltip>
						<TooltipTrigger asChild>
							<button
								type="button"
								className="ml-1 text-t2 hover:text-primary"
								aria-label={filter.shared ? "Сделать приватным" : "Расшарить"}
								onClick={() => void handleToggleShare(filter)}
							>
								<Share2 className="h-3 w-3" aria-hidden="true" />
							</button>
						</TooltipTrigger>
						<TooltipContent>
							{filter.shared ? "Сделать приватным" : "Расшарить"}
						</TooltipContent>
					</Tooltip>
					<Tooltip>
						<TooltipTrigger asChild>
							<button
								type="button"
								className="text-t2 hover:text-destructive"
								aria-label="Удалить фильтр"
								onClick={() => void handleDelete(filter.id)}
							>
								<Trash2 className="h-3 w-3" aria-hidden="true" />
							</button>
						</TooltipTrigger>
						<TooltipContent>Удалить фильтр</TooltipContent>
					</Tooltip>
				</div>
			))}

			{/* Диалог создания нового фильтра */}
			<Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
				<DialogTrigger asChild>
					<Button variant="outline" size="sm" className="h-7 gap-1 text-xs">
						<Plus className="h-3 w-3" />
						<Filter className="h-3 w-3" />
						Сохранить фильтр
					</Button>
				</DialogTrigger>
				<DialogContent className="sm:max-w-[400px]">
					<DialogHeader>
						<DialogTitle>Сохранить текущий фильтр</DialogTitle>
					</DialogHeader>
					<div className="flex flex-col gap-4 pt-2">
						<Input
							placeholder="Название фильтра"
							value={newFilterName}
							onChange={(e) => setNewFilterName(e.target.value)}
							onKeyDown={(e) => {
								if (e.key === "Enter") void handleCreate();
							}}
						/>
						<Button
							onClick={() => void handleCreate()}
							disabled={!newFilterName.trim() || loading}
						>
							Сохранить
						</Button>
					</div>
				</DialogContent>
			</Dialog>
		</div>
	);
}
