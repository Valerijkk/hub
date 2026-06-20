import {
	useIsSpeaking,
	useParticipants,
	VideoTrack,
} from "@livekit/components-react";
import type { Participant, TrackPublication } from "livekit-client";
import { Track } from "livekit-client";
import { MicOff } from "lucide-react";

import { cn } from "@/lib/utils";

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
 * Плитка участника — Телемост стиль.
 * Занимает всю доступную площадь ячейки, аватар крупный по центру.
 */
function ParticipantTile({ participant }: { participant: Participant }) {
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

	// Находим камерный трек участника
	const cameraPublication: TrackPublication | undefined = participant
		.getTrackPublications()
		.find(
			(pub) =>
				pub.source === Track.Source.Camera &&
				pub.track !== undefined &&
				!pub.isMuted,
		);
	const hasVideo = !!cameraPublication;

	return (
		<div
			className={cn(
				"relative flex flex-col items-center justify-center rounded-2xl overflow-hidden",
				"bg-[#2a2f3e] border border-white/[0.06]",
				isSpeaking && "border-green-500 shadow-[0_0_16px_rgba(34,197,94,0.35)]",
			)}
		>
			{/* Рука */}
			{meta.hand_raised && (
				<span className="absolute top-3 left-1/2 -translate-x-1/2 text-2xl animate-bounce z-10">
					✋
				</span>
			)}

			{/* Видео с камеры или аватар */}
			{hasVideo && cameraPublication ? (
				<div className="absolute inset-0">
					<VideoTrack
						trackRef={{
							participant,
							publication: cameraPublication,
							source: Track.Source.Camera,
						}}
						className="h-full w-full object-cover"
					/>
					{/* Имя поверх видео */}
					<div className="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/70 to-transparent px-3 py-2">
						<span className="text-[13px] font-medium text-white truncate block">
							{name}
						</span>
					</div>
					{/* Индикатор мьюта поверх видео */}
					{isMuted && (
						<span className="absolute bottom-2 right-2 flex h-7 w-7 items-center justify-center rounded-full bg-red-500 shadow-lg">
							<MicOff className="h-3.5 w-3.5 text-white" />
						</span>
					)}
				</div>
			) : (
				<>
					{/* Аватар — крупный, адаптивный */}
					<div className="relative">
						{isSpeaking && (
							<div className="absolute -inset-3 rounded-full border-2 border-green-500/50 animate-speaking" />
						)}

						<div
							className={cn(
								"relative flex items-center justify-center rounded-full font-bold",
								"h-20 w-20 text-3xl",
								isSpeaking
									? "bg-green-500/15 text-green-300"
									: "bg-white/10 text-white/70",
							)}
						>
							{initials}
						</div>

						{isMuted && (
							<span className="absolute -bottom-1 -right-1 flex h-7 w-7 items-center justify-center rounded-full bg-red-500 shadow-lg ring-2 ring-[#2a2f3e]">
								<MicOff className="h-3.5 w-3.5 text-white" />
							</span>
						)}
					</div>

					{/* Имя */}
					<span className="mt-3 text-[13px] font-medium text-white/80 truncate max-w-[90%]">
						{name}
					</span>
				</>
			)}
		</div>
	);
}

/**
 * Сетка участников — Телемост стиль.
 * Плитки растягиваются на ВСЮ доступную площадь.
 * 1 чел = 1 колонка. 2 = 1x2. 3-4 = 2x2. 5-6 = 2x3. 7-9 = 3x3. 10+ = 4x(auto)
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

	const count = sorted.length;

	// Рассчитываем колонки и строки чтобы плитки заполняли всё пространство
	let cols: number;
	let rows: number;
	if (count <= 1) {
		cols = 1;
		rows = 1;
	} else if (count <= 4) {
		cols = 2;
		rows = Math.ceil(count / 2);
	} else if (count <= 6) {
		cols = 3;
		rows = 2;
	} else if (count <= 9) {
		cols = 3;
		rows = Math.ceil(count / 3);
	} else {
		cols = 4;
		rows = Math.ceil(count / 4);
	}

	return (
		<div
			className="flex-1 grid gap-1 p-1"
			style={{
				gridTemplateColumns: `repeat(${cols}, 1fr)`,
				gridTemplateRows: `repeat(${rows}, 1fr)`,
			}}
		>
			{sorted.map((p) => (
				<ParticipantTile key={p.identity} participant={p} />
			))}
		</div>
	);
}
