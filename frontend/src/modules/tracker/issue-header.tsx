/**
 * Заголовок задачи -- хлебные крошки, тип, ключ, кнопка удаления.
 */
import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";

/** Конфигурация иконок типов задач. */
export const TYPE_ICONS: Record<string, string> = {
	bug: "\u{1F41B}",
	task: "\u{2705}",
	story: "\u{1F4D6}",
	epic: "\u{26A1}",
	sub_task: "\u{1F4CB}",
	improvement: "\u{1F680}",
	tech_debt: "\u{1F527}",
	spike: "\u{1F50D}",
	release: "\u{1F3F7}\uFE0F",
};

/** Цвет фона бейджа типа задачи. */
const TYPE_BG: Record<string, string> = {
	bug: "bg-red-500/10 ring-red-500/20",
	task: "bg-blue-500/10 ring-blue-500/20",
	story: "bg-green-500/10 ring-green-500/20",
	epic: "bg-purple-500/10 ring-purple-500/20",
	sub_task: "bg-gray-500/10 ring-gray-500/20",
	improvement: "bg-cyan-500/10 ring-cyan-500/20",
	tech_debt: "bg-amber-500/10 ring-amber-500/20",
	spike: "bg-indigo-500/10 ring-indigo-500/20",
	release: "bg-pink-500/10 ring-pink-500/20",
};

export function IssueHeader({
	projectId,
	issueKey,
	issueType,
	onDelete,
	isDeleting,
	canDelete = true,
}: {
	projectId: string;
	issueKey: string;
	issueType: string;
	onDelete: () => void;
	isDeleting: boolean;
	/** Viewer/guest — скрыть удаление (P2-FE-16). */
	canDelete?: boolean;
}) {
	return (
		<>
			{/* Хлебные крошки */}
			<nav className="mb-4 flex items-center gap-2">
				<Link
					to="/projects/$projectId"
					params={{ projectId }}
					className="text-[13px] text-t3 hover:text-t2 transition-colors"
				>
					Проект
				</Link>
				<span className="text-t4 text-[12px]">/</span>
				<Link
					to="/projects/$projectId/issues"
					params={{ projectId }}
					className="text-[13px] text-t3 hover:text-t2 transition-colors"
				>
					Задачи
				</Link>
				<span className="text-t4 text-[12px]">/</span>
				<span className="font-mono text-[13px] text-blue font-semibold">
					{issueKey}
				</span>
			</nav>

			{/* Заголовок с типом и ключом */}
			<div className="mb-6 flex items-start justify-between">
				<div className="flex items-center gap-3">
					<span
						className={`flex h-9 w-9 items-center justify-center rounded-lg text-xl ring-1 ${TYPE_BG[issueType] ?? "bg-gray-500/10 ring-gray-500/20"}`}
					>
						{TYPE_ICONS[issueType] ?? "\u{1F4CB}"}
					</span>
					<span className="font-mono text-[13px] text-blue font-semibold tracking-tight">
						{issueKey}
					</span>
				</div>
				{canDelete ? (
					<Button
						variant="destructive"
						size="sm"
						onClick={onDelete}
						disabled={isDeleting}
						data-testid="issue-delete-btn"
						className="w-[34px] h-[34px] rounded-[var(--r2)] text-t3 hover:bg-[var(--hover)] hover:text-t1 text-[18px] transition-colors duration-150"
					>
						Удалить
					</Button>
				) : null}
			</div>
		</>
	);
}
