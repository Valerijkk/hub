/**
 * Widget "Upcoming Calls" -- list of scheduled calls.
 */
import { Phone, PhoneOff } from "lucide-react";
import type { CallSummary, PersonalData } from "@/api/dashboard";
import { Button } from "@/components/ui/button";

interface Props {
	data?: unknown;
}

/** UUID/generated-ID pattern (e.g. "call-d69b75ce-..." or raw UUIDs). */
const GENERATED_ID_RE = /^(call-)?[0-9a-f]{8}-/i;

/** Returns a human-readable call name, preferring channel_name over raw IDs. */
function getCallDisplayName(call: CallSummary): string {
	const titleIsRaw = !call.title || GENERATED_ID_RE.test(call.title);
	if (titleIsRaw && call.channel_name) {
		return call.channel_name;
	}
	return call.title || call.channel_name || "Звонок";
}

/** Upcoming calls widget. */
export function UpcomingCallsWidget({ data }: Props) {
	const personalData = data as PersonalData | undefined;
	const nextCall = (personalData?.upcoming_calls ?? [])[0];

	if (!nextCall) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-muted-foreground">
				<PhoneOff className="h-8 w-8 opacity-40" />
				<p className="text-sm">Нет предстоящих звонков</p>
				<p className="text-xs opacity-60">Запланированные звонки появятся здесь</p>
			</div>
		);
	}

	const displayName = getCallDisplayName(nextCall);
	const scheduledDate = new Date(nextCall.scheduled_at);
	const timeStr = scheduledDate.toLocaleString("ru-RU", { hour: "2-digit", minute: "2-digit" });
	const dateStr = scheduledDate.toLocaleString("ru-RU", { day: "numeric", month: "long" });

	return (
		<div className="flex flex-col items-center justify-center h-full gap-3 text-center">
			<div className="h-12 w-12 rounded-full bg-violet-500/10 flex items-center justify-center">
				<Phone className="h-6 w-6 text-violet-500" />
			</div>
			<div>
				<p className="font-semibold text-[15px]">{displayName}</p>
				<p className="text-[13px] text-muted-foreground mt-0.5">
					{dateStr} в {timeStr}
				</p>
				{nextCall.channel_name && (
					<p className="text-[11px] text-muted-foreground/60 mt-0.5">
						{nextCall.channel_name}
					</p>
				)}
			</div>
			<Button
				size="sm"
				className="btn-glow gap-1.5 bg-gradient-to-r from-violet-600 to-purple-600 text-white border-0 hover:from-violet-500 hover:to-purple-500 mt-1"
			>
				<Phone className="h-3.5 w-3.5" />
				Войти в звонок
			</Button>
		</div>
	);
}
