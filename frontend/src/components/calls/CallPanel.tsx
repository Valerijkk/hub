import { Loader2, Maximize2, Minimize2 } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { useCallStore } from "@/stores/call-store";
import { CallRoom } from "./CallRoom";

/**
 * Обёртка звонка.
 * По умолчанию — большая панель (800x650) внизу справа.
 * Развёрнутый режим — полный экран поверх всего.
 */
export function CallPanel() {
	const { isInCall, isConnecting, token, leaveCall, error, setError } =
		useCallStore();
	const [isFullscreen, setIsFullscreen] = useState(false);

	if (!isInCall && !isConnecting && !error) return null;

	const handleDisconnect = () => {
		setIsFullscreen(false);
		void leaveCall();
	};

	return (
		<div
			className={cn(
				"fixed z-50 overflow-hidden shadow-2xl transition-[inset,border-radius,box-shadow,width,height] duration-500 ease-out",
				isFullscreen
					? "inset-0 rounded-none"
					: "bottom-4 right-4 h-[650px] w-[800px] rounded-2xl ring-1 ring-white/10 shadow-[0_0_40px_rgba(59,130,246,0.15)]",
			)}
		>
			<div className="relative h-full w-full bg-[#0B0F19] overflow-hidden">
				{/* Градиентный акцент сверху */}
				{!isFullscreen && (
					<div className="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-blue-500/40 to-transparent" />
				)}

				{/* Ошибка */}
				{error && (
					<div className="flex items-center justify-between bg-red-500/10 border-b border-red-500/20 px-4 py-2.5 text-sm text-red-400">
						<span>{error}</span>
						<button
							type="button"
							onClick={() => setError(null)}
							className="ml-2 text-xs underline hover:text-red-300 transition-colors"
						>
							Закрыть
						</button>
					</div>
				)}

				{/* Кнопка fullscreen/minimize */}
				{isInCall && (
					<button
						type="button"
						onClick={() => setIsFullscreen(!isFullscreen)}
						className={cn(
							"absolute top-3 right-3 z-10 flex h-8 w-8 items-center justify-center",
							"rounded-lg bg-white/5 text-white/60",
							"hover:bg-white/15 hover:text-white hover:scale-105",
							"transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 backdrop-blur-sm",
						)}
						title={isFullscreen ? "Свернуть" : "На весь экран"}
					>
						{isFullscreen ? (
							<Minimize2 className="h-4 w-4" />
						) : (
							<Maximize2 className="h-4 w-4" />
						)}
					</button>
				)}

				{/* Загрузка */}
				{isConnecting && (
					<div className="flex flex-1 h-full items-center justify-center">
						<div className="flex flex-col items-center gap-4">
							<div className="relative">
								<div className="absolute inset-0 rounded-full bg-blue-500/20 animate-ping" />
								<Loader2 className="relative h-10 w-10 animate-spin text-blue-400" />
							</div>
							<span className="text-sm font-medium text-white/60">
								Подключение к звонку...
							</span>
						</div>
					</div>
				)}

				{/* Комната звонка */}
				{isInCall && token && (
					<CallRoom token={token} onDisconnect={handleDisconnect} />
				)}
			</div>
		</div>
	);
}
