import { useQuery } from "@tanstack/react-query";
import { fetchPublicStatus } from "@/api/status";
import { cn } from "@/lib/utils";

const STATUS_LABEL: Record<string, string> = {
	operational: "Работает",
	degraded: "Частичный сбой",
	outage: "Недоступен",
	unknown: "Неизвестно",
};

const STATUS_COLOR: Record<string, string> = {
	operational: "bg-emerald-500",
	degraded: "bg-amber-500",
	outage: "bg-rose-500",
	unknown: "bg-zinc-500",
};

/** V4-06: публичная status page — без авторизации. */
export function StatusPage() {
	const { data, isLoading, isError, dataUpdatedAt } = useQuery({
		queryKey: ["public", "status"],
		queryFn: fetchPublicStatus,
		refetchInterval: 30_000,
		retry: 1,
	});

	const overall = data?.status ?? (isError ? "degraded" : "operational");

	return (
		<div className="min-h-screen bg-[var(--base)] text-t1">
			<div className="mx-auto max-w-2xl px-4 py-16 space-y-8">
				<header className="text-center space-y-2">
					<p className="text-sm font-medium text-t3 uppercase tracking-widest">
						Hub Status
					</p>
					<h1 className="text-3xl font-bold tracking-tight">
						{overall === "operational"
							? "Все системы работают"
							: "Обнаружены проблемы"}
					</h1>
					{dataUpdatedAt > 0 && (
						<p className="text-sm text-t3">
							Обновлено: {new Date(dataUpdatedAt).toLocaleString("ru-RU")}
						</p>
					)}
				</header>

				{isLoading && (
					<p className="text-center text-sm text-t3" role="status">
						Проверка компонентов…
					</p>
				)}

				<ul className="space-y-3">
					{(data?.components ?? []).map((c) => (
						<li
							key={c.id}
							className="flex items-center gap-4 rounded-xl border border-b1 px-4 py-3"
						>
							<span
								className={cn(
									"h-3 w-3 shrink-0 rounded-full",
									STATUS_COLOR[c.status] ?? STATUS_COLOR.unknown,
								)}
								aria-hidden="true"
							/>
							<div className="flex-1 min-w-0">
								<p className="font-medium">{c.name}</p>
								{c.description && (
									<p className="text-sm text-t3">{c.description}</p>
								)}
							</div>
							<span className="text-sm text-t2 shrink-0">
								{STATUS_LABEL[c.status] ?? c.status}
							</span>
						</li>
					))}
				</ul>

				<p className="text-center text-xs text-t3">
					Публичная страница статуса Hub · автообновление каждые 30 сек
				</p>
			</div>
		</div>
	);
}
