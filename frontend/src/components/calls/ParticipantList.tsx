import { useIsSpeaking, useParticipants } from "@livekit/components-react";
import type { Participant } from "livekit-client";
import { MicOff, Monitor } from "lucide-react";

import { cn } from "@/lib/utils";

/** Метаданные участника из LiveKit. */
interface ParticipantMeta {
	hand_raised?: boolean;
}

function parseMeta(metadata?: string): ParticipantMeta {
	if (!metadata) return {};
	try {
		return JSON.parse(metadata) as ParticipantMeta;
	} catch {
		return {};
	}
}

/**
 * Плитка участника — Телемост-стиль.
 * Большой кубик с аватаром по центру, имя внизу, индикаторы.
 */
function ParticipantTile({
	participant,
	isSolo,
}: {
	participant: Participant;
	isSolo: boolean;
}) {
	const isSpeaking = useIsSpeaking(participant);
	const meta = parseMeta(participant.metadata);
	const name = participant.name || participant.identity;
	const initials = name
		.split(" ")
		.map((w) => w[0])
		.join("")
		.slice(0, 2)
		.toUpperCase();
	const isMuted = !participant.isMicrophoneEnabled;

	return (
		<div
			className={cn(
				"relative flex flex-col items-center justify-center rounded-2xl transition-all duration-300 overflow-hidden",
				"bg-[#1a1f2e] border border-white/[0.08]",
				isSolo ? "min-h-[400px]" : "min-h-[180px]",
				isSpeaking &&
					"ring-2 ring-green-500/50 shadow-[0_0_30px_rgba(34,197,94,0.12)]",
			)}
		>
			{/* Поднятая рука */}
			{meta.hand_raised && (
				<span className="absolute top-3 left-1/2 -translate-x-1/2 text-2xl animate-bounce z-10">
					✋
				</span>
			)}

			{/* Аватар */}
			<div className="relative">
				{isSpeaking && (
					<div className="absolute -inset-3 rounded-full border-2 border-green-500/50 animate-[pulse_1.5s_ease-in-out_infinite]" />
				)}

				<div
					className={cn(
						"relative flex items-center justify-center rounded-full font-bold transition-all duration-300",
						isSolo ? "h-32 w-32 text-5xl" : "h-20 w-20 text-2xl",
						isSpeaking
							? "bg-green-500/15 text-green-400"
							: "bg-white/10 text-white/80",
					)}
				>
					{initials}

					{isMuted && !isSolo && (
						<div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/40">
							<MicOff className="h-5 w-5 text-red-400/80" />
						</div>
					)}
				</div>

				{isMuted && (
					<span
						className={cn(
							"absolute -bottom-1 -right-1 flex items-center justify-center rounded-full bg-red-500 shadow-md ring-2 ring-[#1a1f2e]",
							isSolo ? "h-8 w-8" : "h-6 w-6",
						)}
					>
						<MicOff className={cn("text-white", isSolo ? "h-4 w-4" : "h-3 w-3")} />
					</span>
				)}
			</div>

			{/* Имя */}
			<span
				className={cn(
					"font-medium text-white/80 truncate max-w-[90%] text-center mt-3",
					isSolo ? "text-lg" : "text-sm",
				)}
			>
				{name}
			</span>

			{/* Индикаторы */}
			<div className="flex items-center gap-2 mt-1.5 min-h-[18px]">
				{participant.isScreenShareEnabled && (
					<span className="flex items-center gap-1 text-xs text-blue-400">
						<Monitor className="h-3.5 w-3.5" />
						экран
					</span>
				)}
				{isSpeaking && (
					<span className="text-xs text-green-400 font-medium">говорит</span>
				)}
			</div>
		</div>
	);
}

/**
 * Сетка участников — Телемост-стиль.
 * 1 участник: одна большая плитка по центру.
 * 2-4: сетка 2 колонки, плитки на всю ширину.
 * 5-9: 3 колонки. 10+: 4 колонки.
 */
export function ParticipantList() {
	const participants = useParticipants();

	const sorted = [...participants].sort((a, b) => {
		const aMeta = parseMeta(a.metadata);
		const bMeta = parseMeta(b.metadata);
		if (aMeta.hand_raised && !bMeta.hand_raised) return -1;
		if (!aMeta.hand_raised && bMeta.hand_raised) return 1;
		return (a.name ?? a.identity).localeCompare(b.name ?? b.identity);
	});

	const isSolo = sorted.length <= 1;

	const gridCols = isSolo
		? "grid-cols-1 place-items-center"
		: sorted.length <= 4
			? "grid-cols-2"
			: sorted.length <= 9
				? "grid-cols-3"
				: "grid-cols-4";

	return (
		<div className="flex-1 overflow-y-auto p-1">
			<div
				className={cn(
					"grid gap-3 h-full",
					gridCols,
					isSolo && "content-center",
				)}
			>
				{sorted.map((p) => (
					<ParticipantTile
						key={p.identity}
						participant={p}
						isSolo={isSolo}
					/>
				))}
			</div>
		</div>
	);
}
