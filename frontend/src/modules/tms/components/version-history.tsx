import type { TestCaseVersionResponse } from "@/api/tms";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

/** Форматирует относительное время на русском языке. */
function formatRelativeTime(dateStr: string): string {
	const date = new Date(dateStr);
	const now = new Date();
	const diffMs = now.getTime() - date.getTime();
	const diffMin = Math.floor(diffMs / 60000);
	const diffHours = Math.floor(diffMs / 3600000);
	const diffDays = Math.floor(diffMs / 86400000);

	if (diffMin < 1) return "только что";
	if (diffMin < 60) return `${diffMin} мин. назад`;
	if (diffHours < 24) return `${diffHours} ч. назад`;
	if (diffDays === 1) return "вчера";
	if (diffDays < 7) return `${diffDays} дн. назад`;
	return date.toLocaleDateString("ru-RU");
}

/** Пропсы компонента VersionHistory. */
interface VersionHistoryProps {
	/** Раскрыта ли панель истории. */
	versionsOpen: boolean;
	/** Переключение видимости панели. */
	onToggle: () => void;
	/** Список версий тест-кейса. */
	versions?: TestCaseVersionResponse[];
	/** Колбэк отката к указанной версии. */
	onRollback: (versionId: string) => void;
	/** Процесс отката в данный момент. */
	isRollingBack: boolean;
	/** Viewer — скрыть кнопку отката (P2-FE-16). */
	canRollback?: boolean;
}

/** Панель истории версий тест-кейса с возможностью отката. */
export function VersionHistory({
	versionsOpen,
	onToggle,
	versions,
	onRollback,
	isRollingBack,
	canRollback = true,
}: VersionHistoryProps) {
	return (
		<Card className="shadow-card">
			<CardHeader className="pb-2 px-4 pt-4">
				<div className="flex items-center justify-between">
					<CardTitle className="text-[11px] font-bold uppercase tracking-wider text-t2">
						История версий
					</CardTitle>
					<Button
						variant="ghost"
						size="sm"
						className="h-6 text-[11px] px-2"
						onClick={onToggle}
					>
						{versionsOpen ? "Скрыть" : "Показать"}
					</Button>
				</div>
			</CardHeader>
			{versionsOpen && versions && (
				<CardContent className="px-4 pb-4">
					{versions.length === 0 ? (
						<p className="text-xs text-t2">Нет предыдущих версий</p>
					) : (
						<div className="space-y-2">
							{versions.map((v: TestCaseVersionResponse) => (
								<div
									key={v.id}
									className="rounded-lg border p-2.5 hover:bg-raised/30 transition-colors"
								>
									<div className="flex items-center justify-between">
										<span className="text-xs font-semibold tabular-nums">
											v{v.version_number}
										</span>
										{canRollback ? (
											<Button
												variant="ghost"
												size="sm"
												className="h-5 text-[10px] px-1.5"
												disabled={isRollingBack}
												data-testid="version-rollback-btn"
												onClick={() => onRollback(v.id)}
											>
												Откатить
											</Button>
										) : null}
									</div>
									{v.change_description && (
										<p className="text-[11px] text-t2 mt-0.5">
											{v.change_description}
										</p>
									)}
									<div className="text-[10px] text-t2/60 mt-1">
										{v.created_by?.name} -- {formatRelativeTime(v.created_at)}
									</div>
								</div>
							))}
						</div>
					)}
				</CardContent>
			)}
		</Card>
	);
}
