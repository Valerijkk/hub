import { Loader2, Maximize2, Minimize2 } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { useCallStore } from "@/stores/call-store";
import { CallRoom } from "./CallRoom";

/**
 * Обёртка звонка в стиле Tinkoff Talk.
 * Поддерживает компактный режим (floating внизу справа) и
 * развёрнутый режим (большая панель). Показывает CallRoom
 * с токеном из стора или индикатор загрузки.
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
				"fixed z-50 overflow-hidden border bg-[#0B0F19] shadow-2xl transition-all duration-300",
				isExpanded
					? "bottom-4 right-4 h-[600px] w-[700px] rounded-2xl"
					: "bottom-4 right-4 h-[500px] w-[420px] rounded-2xl",
			)}
		>
			{/* Ошибка */}
			{error && (
				<div className="flex items-center justify-between bg-red-500/10 border-b border-red-500/20 px-4 py-2 text-sm text-red-400">
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
					className="absolute top-3 right-3 z-10 flex h-8 w-8 items-center justify-center rounded-lg bg-white/10 text-white/70 hover:bg-white/20 hover:text-white transition-colors"
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
					<div className="flex flex-col items-center gap-3">
						<Loader2 className="h-8 w-8 animate-spin text-white/60" />
						<span className="text-sm text-white/60">
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
	);
}
