import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Send } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import {
	type CallChatMessage,
	listChatMessages,
	sendChatMessage,
} from "@/api/calls";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";
import { useWsStore } from "@/stores/ws-store";

/** Пропсы компонента эфемерного чата звонка. */
interface CallChatProps {
	/** ID активного звонка. */
	callId: string;
}

/**
 * Форматирование времени сообщения.
 */
function formatTime(dateStr: string): string {
	try {
		const d = new Date(dateStr);
		return d.toLocaleTimeString("ru-RU", {
			hour: "2-digit",
			minute: "2-digit",
		});
	} catch {
		return "";
	}
}

/** Эфемерный чат звонка — сворачиваемая панель справа, 280px. */
export function CallChat({ callId }: CallChatProps) {
	const [text, setText] = useState("");
	const scrollRef = useRef<HTMLDivElement>(null);
	const queryClient = useQueryClient();
	const user = useAuthStore((s) => s.user);

	/** Загрузка истории сообщений. */
	const { data: messages = [] } = useQuery({
		queryKey: ["call-chat", callId],
		queryFn: () => listChatMessages(callId),
		refetchInterval: false,
	});

	/** Подписка на новые сообщения через WebSocket. */
	const addListener = useWsStore((s) => s.addListener);
	const removeListener = useWsStore((s) => s.removeListener);

	useEffect(() => {
		const handler = (_event: string, payload: unknown) => {
			const msg = payload as CallChatMessage;
			if (msg.call_id === callId) {
				queryClient.setQueryData<CallChatMessage[]>(
					["call-chat", callId],
					(old = []) => [...old, msg],
				);
			}
		};
		addListener(handler);
		return () => removeListener(handler);
	}, [callId, queryClient, addListener, removeListener]);

	/** Отправка сообщения. */
	const sendMutation = useMutation({
		mutationFn: (msg: string) => sendChatMessage(callId, msg),
		onSuccess: () => setText(""),
	});

	/** Авто-скролл при новых сообщениях. */
	useEffect(() => {
		scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight });
	}, [messages.length]);

	/** Обработка отправки по Enter. */
	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (text.trim()) sendMutation.mutate(text.trim());
	};

	const isOwn = (msg: CallChatMessage) => msg.user_id === user?.id;

	return (
		<div className="flex h-full w-[280px] flex-shrink-0 flex-col border-l border-white/5 bg-black/20">
			{/* Заголовок */}
			<div className="flex items-center justify-between border-b border-white/5 px-4 py-3">
				<span className="text-sm font-semibold text-white/80">
					Чат звонка
				</span>
				<span className="text-[10px] text-white/30">
					{messages.length} сообщ.
				</span>
			</div>

			{/* Сообщения */}
			<div
				ref={scrollRef}
				className="flex-1 overflow-y-auto px-3 py-3 space-y-3"
			>
				{messages.length === 0 && (
					<p className="text-center text-xs text-white/20 mt-8">
						Нет сообщений
					</p>
				)}
				{messages.map((msg) => {
					const own = isOwn(msg);
					return (
						<div
							key={msg.id}
							className={cn(
								"flex flex-col gap-0.5",
								own ? "items-end" : "items-start",
							)}
						>
							{/* Имя + время */}
							<div
								className={cn(
									"flex items-center gap-2 px-1",
									own ? "flex-row-reverse" : "",
								)}
							>
								<span className="text-[10px] font-medium text-white/40">
									{own ? "Вы" : msg.user_name}
								</span>
								{msg.created_at && (
									<span className="text-[10px] text-white/20">
										{formatTime(msg.created_at)}
									</span>
								)}
							</div>

							{/* Пузырёк сообщения */}
							<div
								className={cn(
									"max-w-[85%] rounded-2xl px-3 py-2 text-sm leading-relaxed",
									own
										? "bg-blue-500/20 text-blue-100 rounded-br-md"
										: "bg-white/[0.06] text-white/80 rounded-bl-md",
								)}
							>
								{msg.text}
							</div>
						</div>
					);
				})}
			</div>

			{/* Ввод сообщения */}
			<form
				onSubmit={handleSubmit}
				className="flex items-center gap-2 border-t border-white/5 px-3 py-2.5"
			>
				<input
					value={text}
					onChange={(e) => setText(e.target.value)}
					placeholder="Сообщение..."
					className={cn(
						"flex-1 h-9 rounded-xl bg-white/5 border-0 px-3 text-sm text-white/90",
						"placeholder:text-white/25",
						"focus:outline-none focus:ring-1 focus:ring-blue-500/40",
						"transition-all duration-200",
					)}
				/>
				<button
					type="submit"
					disabled={!text.trim()}
					className={cn(
						"flex h-9 w-9 items-center justify-center rounded-xl transition-all duration-200",
						text.trim()
							? "bg-blue-500 text-white hover:bg-blue-600 hover:scale-105"
							: "bg-white/5 text-white/20 cursor-not-allowed",
					)}
				>
					<Send className="h-4 w-4" />
				</button>
			</form>
		</div>
	);
}
