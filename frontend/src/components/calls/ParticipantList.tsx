import { useIsSpeaking, useParticipants } from "@livekit/components-react";
import type { Participant } from "livekit-client";
import { MicOff, Monitor } from "lucide-react";

import { cn } from "@/lib/utils";

/** Метаданные участника из LiveKit. */
interface ParticipantMeta {
	/** Рука поднята. */
	hand_raised?: boolean;
}

/**
 * Безопасный парсинг метаданных участника.
 * @param metadata - строка JSON из LiveKit
 */
function parseMeta(metadata?: string): ParticipantMeta {
	if (!metadata) return {};
	try {
		return JSON.parse(metadata) as ParticipantMeta;
	} catch {
		return {};
	}
}

/**
 * Карточка одного участника звонка.
 * Аватар 80x80, имя снизу, анимированное зелёное кольцо при говорении,
 * полупрозрачный оверлей при выключенном микрофоне,
 * прыгающий эмодзи руки при поднятой руке.
 */
function ParticipantCard({ participant }: { participant: Participant }) {
	const isSpeaking = useIsSpeaking(participant);
	const meta = parseMeta(participant.metadata);
	const name = participant.name || participant.identity;
	const isMuted = !participant.isMicrophoneEnabled;

	return (
		<div
			className={cn(
				"relative flex flex-col items-center justify-center gap-2.5 rounded-2xl p-4 transition-all duration-300",
				"bg-white/[0.03] border border-white/[0.06]",
				isSpeaking &&
					"ring-2 ring-green-500/60 shadow-[0_0_24px_rgba(34,197,94,0.15)] bg-white/[0.05]",
			)}
		>
			{/* Поднятая рука — прыгающий эмодзи */}
			{meta.hand_raised && (
				<span className="absolute -top-3 left-1/2 -translate-x-1/2 text-lg animate-bounce z-10">
					&#9995;
				</span>
			)}

			{/* Аватар с кольцом говорения */}
			<div className="relative">
				{/* Анимированное зелёное кольцо при говорении */}
				{isSpeaking && (
					<div className="absolute -inset-1.5 rounded-full border-2 border-green-500/60 animate-[pulse_1.5s_ease-in-out_infinite]" />
				)}

				<div
					className={cn(
						"relative flex h-20 w-20 items-center justify-center rounded-full text-2xl font-semibold transition-all duration-300",
						isSpeaking
							? "bg-green-500/15 text-green-400"
							: "bg-white/10 text-white/80",
					)}
				>
					{name.charAt(0).toUpperCase()}

					{/* Muted overlay */}
					{isMuted && (
						<div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/40">
							<MicOff className="h-5 w-5 text-red-400/80" />
						</div>
					)}
				</div>

				{/* Индикатор микрофона — маленький бейдж */}
				{isMuted && (
					<span className="absolute -bottom-1 -right-1 flex h-6 w-6 items-center justify-center rounded-full bg-red-500 shadow-md ring-2 ring-[#0B0F19]">
						<MicOff className="h-3 w-3 text-white" />
					</span>
				)}
			</div>

			{/* Имя участника */}
			<span className="text-xs font-medium text-white/70 truncate max-w-full text-center leading-tight">
				{name}
			</span>

			{/* Индикаторы */}
			<div className="flex items-center gap-1.5 min-h-[16px]">
				{participant.isScreenShareEnabled && (
					<span className="flex items-center gap-1 text-[10px] text-blue-400">
						<Monitor className="h-3 w-3" />
						<span>экран</span>
					</span>
				)}
				{isSpeaking && (
					<span className="text-[10px] text-green-400 font-medium">
						говорит
					</span>
				)}
			</div>
		</div>
	);
}

/**
 * Сетка участников звонка.
 * Адаптивная сетка: 2 колонки для 2-4 человек, 3 для 5-9, 4 для 10+.
 * Участники с поднятой рукой отображаются первыми.
 */
export function ParticipantList() {
	const participants = useParticipants();

	/** Сортировка: поднявшие руку сверху, затем по имени. */
	const sorted = [...participants].sort((a, b) => {
		const aMeta = parseMeta(a.metadata);
		const bMeta = parseMeta(b.metadata);
		if (aMeta.hand_raised && !bMeta.hand_raised) return -1;
		if (!aMeta.hand_raised && bMeta.hand_raised) return 1;
		return (a.name ?? a.identity).localeCompare(b.name ?? b.identity);
	});

	const gridCols =
		participants.length <= 1
			? "grid-cols-1"
			: participants.length <= 4
				? "grid-cols-2"
				: participants.length <= 9
					? "grid-cols-3"
					: "grid-cols-4";

	return (
		<div className="flex-1 overflow-y-auto">
			<div className={cn("grid gap-2.5 auto-rows-min", gridCols)}>
				{sorted.map((p) => (
					<ParticipantCard key={p.identity} participant={p} />
				))}
			</div>
		</div>
	);
}
