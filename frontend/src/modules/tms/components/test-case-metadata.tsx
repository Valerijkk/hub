import type { TestCaseResponse } from "@/api/tms";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

/** Конфигурация приоритетов для бейджей (цвет точки, фон, текст, подпись). */
const PRIORITY_CONFIG: Record<
	string,
	{ dot: string; bg: string; text: string; label: string }
> = {
	critical: {
		dot: "bg-red-500",
		bg: "bg-red-500/10",
		text: "text-red-600 dark:text-red-400",
		label: "Критический",
	},
	high: {
		dot: "bg-orange-500",
		bg: "bg-orange-500/10",
		text: "text-orange-600 dark:text-orange-400",
		label: "Высокий",
	},
	medium: {
		dot: "bg-yellow-500",
		bg: "bg-yellow-500/10",
		text: "text-yellow-700 dark:text-yellow-400",
		label: "Средний",
	},
	low: {
		dot: "bg-blue-500",
		bg: "bg-blue-500/10",
		text: "text-blue-600 dark:text-blue-400",
		label: "Низкий",
	},
	trivial: {
		dot: "bg-gray-400",
		bg: "bg-gray-500/10",
		text: "text-gray-500",
		label: "Тривиальный",
	},
};

/** Конфигурация статусов тест-кейса (черновик, активный, устаревший). */
const STATUS_CONFIG: Record<
	string,
	{ bg: string; text: string; label: string }
> = {
	draft: {
		bg: "bg-gray-500/10",
		text: "text-gray-600 dark:text-gray-400",
		label: "Черновик",
	},
	active: {
		bg: "bg-green-500/10",
		text: "text-green-700 dark:text-green-400",
		label: "Активный",
	},
	deprecated: {
		bg: "bg-red-500/10",
		text: "text-red-600 dark:text-red-400",
		label: "Устаревший",
	},
};

/** Форматирует относительное время на русском языке ("5 мин. назад", "вчера"). */
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

/** Реэкспорт конфигураций для использования в родительском роуте. */
export { PRIORITY_CONFIG, STATUS_CONFIG };

/** Пропсы компонента TestCaseMetadata. */
interface TestCaseMetadataProps {
	/** Данные тест-кейса. */
	tc: TestCaseResponse;
}

/** Боковая панель метаданных тест-кейса: тип, приоритет, статус, автор, даты, папка, компоненты, метки. */
export function TestCaseMetadata({ tc }: TestCaseMetadataProps) {
	// biome-ignore lint/style/noNonNullAssertion: PRIORITY_CONFIG.medium is always defined
	const priorityCfg = (PRIORITY_CONFIG[tc.priority] ?? PRIORITY_CONFIG.medium)!;
	// biome-ignore lint/style/noNonNullAssertion: STATUS_CONFIG.draft is always defined
	const statusCfg = (STATUS_CONFIG[tc.status] ?? STATUS_CONFIG.draft)!;

	return (
		<div className="space-y-4">
			{/* Metadata card */}
			<Card className="shadow-card">
				<CardHeader className="pb-2 px-4 pt-4">
					<CardTitle className="text-[11px] font-bold uppercase tracking-wider text-t2">
						Метаданные
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-3 px-4 pb-4">
					<div className="grid grid-cols-2 gap-x-3 gap-y-2.5">
						<div>
							<div className="text-[10px] text-t2/70 uppercase tracking-wider font-medium">
								Приоритет
							</div>
							<div className="flex items-center gap-1 mt-0.5">
								<span className={cn("h-2 w-2 rounded-full", priorityCfg.dot)} />
								<span className="text-xs font-medium">{priorityCfg.label}</span>
							</div>
						</div>
						<div>
							<div className="text-[10px] text-t2/70 uppercase tracking-wider font-medium">
								Статус
							</div>
							<Badge
								variant="secondary"
								className={cn(
									"text-[10px] font-medium mt-0.5",
									statusCfg.bg,
									statusCfg.text,
								)}
							>
								{statusCfg.label}
							</Badge>
						</div>
						<div>
							<div className="text-[10px] text-t2/70 uppercase tracking-wider font-medium">
								Тип
							</div>
							<span className="text-xs font-medium mt-0.5">
								{tc.case_type === "bdd" ? "BDD" : "Classic"}
							</span>
						</div>
						<div>
							<div className="text-[10px] text-t2/70 uppercase tracking-wider font-medium">
								Версия
							</div>
							<span className="text-xs font-medium mt-0.5 tabular-nums">
								v{tc.version}
							</span>
						</div>
					</div>

					<div className="border-t pt-2.5 space-y-2.5">
						{tc.created_by && (
							<div>
								<div className="text-[10px] text-t2/70 uppercase tracking-wider font-medium">
									Автор
								</div>
								<div className="text-xs font-medium mt-0.5">
									{tc.created_by.name}
								</div>
							</div>
						)}
						<div>
							<div className="text-[10px] text-t2/70 uppercase tracking-wider font-medium">
								Создан
							</div>
							<div className="text-xs text-t2 mt-0.5">
								{formatRelativeTime(tc.created_at)}
							</div>
						</div>
						<div>
							<div className="text-[10px] text-t2/70 uppercase tracking-wider font-medium">
								Обновлён
							</div>
							<div className="text-xs text-t2 mt-0.5">
								{formatRelativeTime(tc.updated_at)}
							</div>
						</div>
						{tc.suite && (
							<div>
								<div className="text-[10px] text-t2/70 uppercase tracking-wider font-medium">
									Папка
								</div>
								<div className="text-xs font-medium mt-0.5 flex items-center gap-1">
									<svg
										className="h-3 w-3 text-t2"
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
									{tc.suite.name}
								</div>
							</div>
						)}
					</div>
				</CardContent>
			</Card>

			{/* Components */}
			{tc.components && tc.components.length > 0 && (
				<Card className="shadow-card">
					<CardHeader className="pb-2 px-4 pt-4">
						<CardTitle className="text-[11px] font-bold uppercase tracking-wider text-t2">
							Компоненты
						</CardTitle>
					</CardHeader>
					<CardContent className="px-4 pb-4">
						<div className="flex flex-wrap gap-1.5">
							{tc.components.map((c) => (
								<Badge
									key={c.id}
									variant="outline"
									className="text-[10px] font-medium"
								>
									{c.name}
								</Badge>
							))}
						</div>
					</CardContent>
				</Card>
			)}

			{/* Labels */}
			{tc.labels && tc.labels.length > 0 && (
				<Card className="shadow-card">
					<CardHeader className="pb-2 px-4 pt-4">
						<CardTitle className="text-[11px] font-bold uppercase tracking-wider text-t2">
							Метки
						</CardTitle>
					</CardHeader>
					<CardContent className="px-4 pb-4">
						<div className="flex flex-wrap gap-1.5">
							{tc.labels.map((l) => (
								<Badge
									key={l.id}
									variant="secondary"
									className="text-[10px] font-medium"
									style={{ backgroundColor: l.color + "20", color: l.color }}
								>
									{l.name}
								</Badge>
							))}
						</div>
					</CardContent>
				</Card>
			)}
		</div>
	);
}
