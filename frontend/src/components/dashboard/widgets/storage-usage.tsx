/**
 * Виджет «Хранилище» — использование дискового пространства.
 */
import { HardDrive } from "lucide-react";
import {
	type DashboardWidgetData,
	isAdminStats,
} from "@/components/dashboard/widget-data";

interface Props {
	data?: DashboardWidgetData;
}

/** Форматирование байт в человекочитаемый вид. */
function formatBytes(bytes: number): string {
	if (bytes === 0) return "0 B";
	const k = 1024;
	const sizes = ["B", "KB", "MB", "GB", "TB"];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	return `${(bytes / k ** i).toFixed(1)} ${sizes[i]}`;
}

/** Виджет использования хранилища. */
export function StorageUsageWidget({ data }: Props) {
	const adminData = isAdminStats(data) ? data : undefined;
	const storageUsed = adminData?.storage_used ?? 0;

	return (
		<div className="flex flex-col items-center justify-center h-full gap-3">
			<div className="relative">
				<HardDrive className="h-10 w-10 text-t2 opacity-60" />
			</div>
			<div className="text-center">
				<span className="text-3xl font-bold tabular-nums">
					{formatBytes(storageUsed)}
				</span>
				<p className="text-sm text-t2 mt-0.5">Использовано</p>
			</div>
			{adminData?.top_projects && adminData.top_projects.length > 0 && (
				<div className="mt-1 w-full space-y-1.5 pt-2 border-t border-border/50">
					{adminData.top_projects.slice(0, 3).map((project) => (
						<div key={project.id} className="flex justify-between text-xs px-1">
							<span className="truncate text-t2">{project.name}</span>
							<span className="tabular-nums text-t2/80 shrink-0 ml-2">
								{project.issue_count} задач
							</span>
						</div>
					))}
				</div>
			)}
		</div>
	);
}
