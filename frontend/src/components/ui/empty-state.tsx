import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

/** Общий empty-state: иконка, заголовок, опционально описание. */
export function EmptyState({
	icon: Icon,
	title,
	description,
	className,
}: {
	icon?: LucideIcon;
	title: string;
	description?: string;
	className?: string;
}) {
	return (
		<div
			className={cn(
				"flex flex-col items-center justify-center gap-3 py-12 text-center",
				className,
			)}
		>
			{Icon && (
				<div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted" aria-hidden="true">
					<Icon className="h-6 w-6 text-muted-foreground" aria-hidden="true" />
				</div>
			)}
			<div className="space-y-1">
				<p className="text-sm font-medium text-t2">{title}</p>
				{description && <p className="text-sm text-t3">{description}</p>}
			</div>
		</div>
	);
}
