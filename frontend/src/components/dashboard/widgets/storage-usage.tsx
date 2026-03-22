/**
 * Виджет "Хранилище" -- визуализация использования хранилища.
 */
import type { AdminStats } from "@/api/dashboard";

interface Props {
	data?: unknown;
}

/** Форматирует байты в человекочитаемый вид. */
function formatBytes(bytes: number): string {
	if (bytes === 0) return "0 B";
	const k = 1024;
	const sizes = ["B", "KB", "MB", "GB", "TB"];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	return `${(bytes / k ** i).toFixed(1)} ${sizes[i]}`;
}

/** Виджет использования хранилища. */
export function StorageUsageWidget({ data }: Props) {
	const adminData = data as AdminStats | undefined;
	const storageUsed = adminData?.storage_used ?? 0;

	return (
		<div className="flex flex-col items-center justify-center h-full gap-2">
			<span className="text-3xl font-bold">{formatBytes(storageUsed)}</span>
			<span className="text-sm text-muted-foreground">Использовано</span>
			{adminData?.top_projects && adminData.top_projects.length > 0 && (
				<div className="mt-2 w-full space-y-1">
					{adminData.top_projects.slice(0, 3).map((project) => (
						<div
							key={project.id}
							className="flex justify-between text-xs text-muted-foreground"
						>
							<span className="truncate">{project.name}</span>
							<span>{project.issue_count} задач</span>
						</div>
					))}
				</div>
			)}
		</div>
	);
}
