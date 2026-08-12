/**
 * Виджет «Ближайшие звонки» — компактный список по прототипу.
 * Shows up to 3 calls: time left | name | "Войти" button.
 */
import { Phone, PhoneOff } from "lucide-react";
import type { CallSummary } from "@/api/dashboard";
import {
	type DashboardWidgetData,
	isPersonalData,
} from "@/components/dashboard/widget-data";
import { useCallStore } from "@/stores/call-store";

interface Props {
	data?: DashboardWidgetData;
}

const GENERATED_ID_RE = /^(call-)?[0-9a-f]{8}-/i;

function getCallDisplayName(call: CallSummary): string {
	const titleIsRaw = !call.title || GENERATED_ID_RE.test(call.title);
	// channel_name предпочтительнее для коротких/числовых заголовков
	if (titleIsRaw && call.channel_name) return call.channel_name;
	if (call.title && call.title.length <= 2 && call.channel_name)
		return call.channel_name;
	return call.title || call.channel_name || "Звонок";
}

function formatCallTime(isoDate: string): string {
	const d = new Date(isoDate);
	const now = new Date();
	const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
	const tomorrow = new Date(today);
	tomorrow.setDate(tomorrow.getDate() + 1);
	const callDay = new Date(d.getFullYear(), d.getMonth(), d.getDate());

	const time = d.toLocaleTimeString("ru-RU", {
		hour: "2-digit",
		minute: "2-digit",
	});

	if (callDay.getTime() === today.getTime()) return time;
	if (callDay.getTime() === tomorrow.getTime()) return "Завтра";
	return d.toLocaleDateString("ru-RU", { day: "numeric", month: "short" });
}

export function UpcomingCallsWidget({ data }: Props) {
	const personalData = isPersonalData(data) ? data : undefined;
	const calls = (personalData?.upcoming_calls ?? []).slice(0, 3);
	const joinCall = useCallStore((s) => s.joinCall);

	if (calls.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-t2">
				<PhoneOff className="h-6 w-6 opacity-40" />
				<p className="text-xs">Нет предстоящих звонков</p>
			</div>
		);
	}

	return (
		<div className="space-y-1">
			{calls.map((call: CallSummary, i: number) => (
				<div
					key={call.id}
					className="flex items-center gap-2.5 py-1.5 px-1 rounded-md hover:bg-white/[0.04] transition-colors"
				>
					<span className="text-[12px] text-t2 shrink-0 w-[52px] tabular-nums">
						{formatCallTime(call.scheduled_at)}
					</span>
					<div className="h-6 w-6 rounded-full bg-violet-500/15 flex items-center justify-center shrink-0">
						<Phone className="h-3 w-3 text-violet-400" />
					</div>
					<span className="text-[13px] truncate flex-1">
						{getCallDisplayName(call)}
					</span>
					<button
						type="button"
						className={`shrink-0 text-[11px] font-medium px-2.5 py-1 rounded-md transition-colors ${
							i === 0
								? "bg-gradient-to-r from-violet-600 to-purple-600 text-white btn-glow"
								: "border border-white/[0.10] text-t2 hover:text-white hover:border-white/20"
						}`}
						onClick={() => {
							void joinCall(call.id);
						}}
					>
						<span className="flex items-center gap-1">
							<Phone className="h-2.5 w-2.5" />
							Войти
						</span>
					</button>
				</div>
			))}
		</div>
	);
}
