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
			<div className="flex flex-1 flex-col items-center justify-center text-muted-foreground">
				<Monitor className="mb-2 h-12 w-12 opacity-30" />
				<p className="text-sm">Нет демонстрации экрана</p>
			</div>
		);
	}

	/** Grid: 1 share = full width, 2+ = grid cols-2. */
	const gridClass =
		tracks.length === 1 ? "grid grid-cols-1" : "grid grid-cols-2";

	return (
		<div className={`flex-1 gap-2 overflow-auto ${gridClass}`}>
			{tracks.map((trackRef) => (
				<div
					key={trackRef.participant.identity + "-screenshare"}
					className="relative overflow-hidden rounded-lg bg-black"
				>
					<VideoTrack
						trackRef={trackRef}
						className="h-full w-full object-contain"
					/>
					<div className="absolute bottom-2 left-2 rounded bg-black/60 px-2 py-0.5 text-xs text-white">
						{trackRef.participant.name || trackRef.participant.identity}
					</div>
				</div>
			))}
		</div>
	);
}
