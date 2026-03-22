import { useTracks, VideoTrack } from "@livekit/components-react";
import { Track } from "livekit-client";
import { Monitor } from "lucide-react";

/**
 * Отображение демонстраций экрана участников.
 * Поддерживает несколько одновременных демонстраций (CALL-03).
 * Если нет активных демонстраций — показывает placeholder.
 */
export function ScreenShareView() {
	const tracks = useTracks([Track.Source.ScreenShare]);

	if (tracks.length === 0) {
		return (
			<div className="flex flex-1 flex-col items-center justify-center text-white/30">
				<Monitor className="mb-3 h-14 w-14 opacity-20" />
				<p className="text-sm font-medium">Нет демонстрации экрана</p>
			</div>
		);
	}

	/** Grid: 1 share = full width, 2+ = grid cols-2. */
	const gridClass =
		tracks.length === 1 ? "grid grid-cols-1" : "grid grid-cols-2";

	return (
		<div className={`flex-1 gap-3 overflow-auto ${gridClass}`}>
			{tracks.map((trackRef) => {
				const presenterName =
					trackRef.participant.name || trackRef.participant.identity;

				return (
					<div
						key={trackRef.participant.identity + "-screenshare"}
						className="relative overflow-hidden rounded-2xl bg-black ring-1 ring-white/10"
					>
						{/* Лейбл "Демонстрация экрана" */}
						<div className="absolute top-3 left-3 z-10 flex items-center gap-2 rounded-full bg-blue-500/80 backdrop-blur-sm px-3 py-1 text-xs font-medium text-white shadow-lg">
							<Monitor className="h-3 w-3" />
							<span>Демонстрация экрана</span>
						</div>

						<VideoTrack
							trackRef={trackRef}
							className="h-full w-full object-contain"
						/>

						{/* Имя презентера — стеклянный оверлей */}
						<div className="absolute bottom-3 left-3 z-10 flex items-center gap-2 rounded-xl bg-black/50 backdrop-blur-xl px-3 py-1.5 text-sm font-medium text-white shadow-lg ring-1 ring-white/10">
							<span className="h-2 w-2 rounded-full bg-green-500" />
							{presenterName}
						</div>
					</div>
				);
			})}
		</div>
	);
}
