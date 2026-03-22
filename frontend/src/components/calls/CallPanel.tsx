import { Loader2, Maximize2, Minimize2 } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { useCallStore } from "@/stores/call-store";
import { CallRoom } from "./CallRoom";

/**
 * Обёртка звонка — floating-панель внизу справа.
 * Компактный и развёрнутый режимы с плавными анимациями
 * и градиентным свечением по краю.
 */
export function CallPanel() {
	const { isInCall, isConnecting, token, leaveCall, error, setError } =
		useCallStore();
	const [isExpanded, setIsExpanded] = useState(false);

	/** Не показывать панель если не в звонке, не подключаемся и нет ошибки. */
	if (!isInCall && !isConnecting && !error) return null;

	/** Обработчик отключения. */
	const handleDisconnect = () => {
		void leaveCall();
	};

	return (
		<div
			className={cn(
				"fixed z-50 overflow-hidden shadow-2xl transition-all duration-500 ease-out",
				"rounded-2xl",
				/* Gradient border glow */
				"ring-1 ring-white/10",
				"shadow-[0_0_40px_rgba(59,130,246,0.15),0_0_80px_rgba(59,130,246,0.05)]",
				isExpanded
					? "bottom-4 right-4 h-[650px] w-[800px]"
					: "bottom-4 right-4 h-[560px] w-[480px]",
			)}
		>
			{/* Inner container with background */}
			<div className="relative h-full w-full bg-[#0B0F19] rounded-2xl overflow-hidden">
				{/* Subtle top gradient accent */}
				<div className="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-blue-500/40 to-transparent" />

				{/* Ошибка */}
				{error && (
					<div className="flex items-center justify-between bg-red-500/10 border-b border-red-500/20 px-4 py-2.5 text-sm text-red-400">
						<span>{error}</span>
						<button
							onClick={() => setError(null)}
							className="ml-2 text-xs underline hover:text-red-300 transition-colors"
						>
							Закрыть
						</button>
					</div>
				)}

				{/* Кнопка expand/collapse */}
				{isInCall && (
					<button
						type="button"
						onClick={() => setIsExpanded(!isExpanded)}
						className={cn(
							"absolute top-3 right-3 z-10 flex h-8 w-8 items-center justify-center",
							"rounded-lg bg-white/5 text-white/60",
							"hover:bg-white/15 hover:text-white hover:scale-105",
							"transition-all duration-200",
							"backdrop-blur-sm",
						)}
						title={isExpanded ? "Свернуть" : "Развернуть"}
					>
						{isExpanded ? (
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
