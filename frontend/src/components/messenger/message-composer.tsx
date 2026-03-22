import { useQueryClient } from "@tanstack/react-query";
import { type KeyboardEvent, useCallback, useRef, useState } from "react";
import { toast } from "sonner";
import { sendMessage, sendTypingIndicator } from "@/api/messenger";
import { cn } from "@/lib/utils";
import { useMessengerStore } from "@/stores/messenger-store";
import { VoiceRecorder } from "./voice-recorder";

/** Пропсы компонента MessageComposer. */
interface MessageComposerProps {
	/** ID канала для отправки сообщения. */
	channelId: string;
	/** ID треда (если отвечаем в тред). */
	threadId?: string;
	/** Заглушка-подсказка. */
	placeholder?: string;
}

/**
 * MessageComposer -- редактор сообщений в стиле Slack / Tinkoff Talk.
 * Округлённый контейнер с кнопками внутри, reply preview сверху,
 * автоподгонка высоты, голосовые сообщения.
 */
export function MessageComposer({
	channelId,
	threadId,
	placeholder = "Написать сообщение...",
}: MessageComposerProps) {
	const queryClient = useQueryClient();
	const [text, setText] = useState("");
	const [isSending, setIsSending] = useState(false);
	const [showVoice, setShowVoice] = useState(false);
	const textareaRef = useRef<HTMLTextAreaElement>(null);

	const replyToMessage = useMessengerStore((s) => s.replyToMessage);
	const setReplyTo = useMessengerStore((s) => s.setReplyTo);

	/** Время последней отправки typing-индикатора (троттлинг 3 сек). */
	const lastTypingSentRef = useRef(0);

	/** Отправка текстового сообщения. */
	const handleSend = useCallback(async () => {
		const trimmed = text.trim();
		if (!trimmed) return;

		setIsSending(true);
		try {
			await sendMessage(channelId, {
				content_json: {
					type: "doc",
					children: [{ type: "p", children: [{ text: trimmed }] }],
				},
				content_text: trimmed,
				thread_id: threadId,
			});
			void queryClient.invalidateQueries({
				queryKey: ["messenger", "messages", channelId],
			});
			setText("");
			setReplyTo(null);
			if (textareaRef.current) {
				textareaRef.current.style.height = "auto";
			}
		} catch (err) {
			const message =
				err instanceof Error ? err.message : "Не удалось отправить сообщение";
			toast.error(message);
		} finally {
			setIsSending(false);
		}
	}, [channelId, threadId, text, queryClient, setReplyTo]);

	/** Обработка клавиш: Enter -- отправить, Shift+Enter -- перенос строки. */
	const handleKeyDown = useCallback(
		(e: KeyboardEvent<HTMLTextAreaElement>) => {
			if (e.key === "Enter" && !e.shiftKey) {
				e.preventDefault();
				void handleSend();
			}
		},
		[handleSend],
	);

	/** Автоматическая подгонка высоты textarea (макс 200px). */
	const handleInput = useCallback(
		(e: React.ChangeEvent<HTMLTextAreaElement>) => {
			const value = e.target.value;
			setText(value);
			const el = e.target;
			el.style.height = "auto";
			el.style.height = `${Math.min(el.scrollHeight, 200)}px`;

			// Отправка typing-индикатора (троттлинг: не чаще 1 раза в 3 сек).
			if (value.trim()) {
				const now = Date.now();
				if (now - lastTypingSentRef.current > 3000) {
					lastTypingSentRef.current = now;
					void sendTypingIndicator(channelId);
				}
			}
		},
		[channelId],
	);

	/** Отправка голосового сообщения. */
	const handleVoiceSend = useCallback(
		async (voiceUrl: string, durationMs: number) => {
			try {
				await sendMessage(channelId, {
					content_json: {
						type: "doc",
						children: [{ type: "p", children: [{ text: "" }] }],
					},
					voice_url: voiceUrl,
					voice_duration_ms: durationMs,
					thread_id: threadId,
				});
				void queryClient.invalidateQueries({
					queryKey: ["messenger", "messages", channelId],
				});
			} catch (err) {
				const message =
					err instanceof Error
						? err.message
						: "Не удалось отправить голосовое сообщение";
				toast.error(message);
			}
			setShowVoice(false);
		},
		[channelId, threadId, queryClient],
	);

	/** Текст превью сообщения для reply. */
	const replyPreviewText = replyToMessage?.content_text
		? replyToMessage.content_text.slice(0, 80) +
			(replyToMessage.content_text.length > 80 ? "..." : "")
		: "Вложение";

	const hasText = text.trim().length > 0;

	return (
		<div className="space-y-0">
			{/* Reply preview */}
			{replyToMessage && (
				<div className="flex items-center gap-2 px-1 py-2 mb-1.5 animate-fade-in">
					<div className="w-0.5 h-8 rounded-full bg-primary shrink-0" />
					<div className="flex-1 min-w-0">
						<p className="text-xs font-semibold text-primary truncate">
							{replyToMessage.sender_id}
						</p>
						<p className="text-xs text-muted-foreground truncate">
							{replyPreviewText}
						</p>
					</div>
					<button
						type="button"
						onClick={() => setReplyTo(null)}
						className="shrink-0 h-6 w-6 flex items-center justify-center rounded-full hover:bg-muted transition-colors"
						title="Отменить ответ"
					>
						<svg
							className="h-3.5 w-3.5 text-muted-foreground"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M6 18L18 6M6 6l12 12"
							/>
						</svg>
					</button>
				</div>
			)}

			{showVoice ? (
				<VoiceRecorder
					onSend={handleVoiceSend}
					onCancel={() => setShowVoice(false)}
				/>
			) : (
				/* Rounded container with all controls inside */
				<div
					className={cn(
						"flex items-end rounded-2xl border bg-background/80 backdrop-blur-sm shadow-sm",
						"transition-all duration-200",
						"focus-within:shadow-md focus-within:border-ring/50 focus-within:bg-background",
					)}
				>
					{/* Attachment button */}
					<button
						type="button"
						className="flex h-11 w-11 shrink-0 items-center justify-center text-muted-foreground/60 hover:text-muted-foreground transition-colors"
						title="Прикрепить файл"
					>
						<svg
							className="h-5 w-5"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"
							/>
						</svg>
					</button>

					{/* Text input */}
					<textarea
						ref={textareaRef}
						value={text}
						onChange={handleInput}
						onKeyDown={handleKeyDown}
						placeholder={placeholder}
						rows={1}
						disabled={isSending}
						className={cn(
							"flex-1 resize-none bg-transparent py-3 text-sm focus:outline-none",
							"min-h-[44px] max-h-[200px]",
							"placeholder:text-muted-foreground/40",
						)}
					/>

					{/* Emoji button */}
					<button
						type="button"
						className="flex h-11 w-9 shrink-0 items-center justify-center text-muted-foreground/60 hover:text-muted-foreground transition-colors"
						title="Эмодзи"
					>
						<svg
							className="h-5 w-5"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
					</button>

					{/* Voice recorder button */}
					<button
						type="button"
						className="flex h-11 w-9 shrink-0 items-center justify-center text-muted-foreground/60 hover:text-muted-foreground transition-colors"
						title="Голосовое сообщение"
						onClick={() => setShowVoice(true)}
					>
						<svg
							className="h-5 w-5"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"
							/>
						</svg>
					</button>

					{/* Send button */}
					<button
						type="button"
						className={cn(
							"flex h-11 w-11 shrink-0 items-center justify-center rounded-r-2xl transition-all duration-200",
							hasText
								? "text-primary hover:bg-primary/10"
								: "text-muted-foreground/30 cursor-not-allowed",
						)}
						onClick={() => void handleSend()}
						disabled={isSending || !hasText}
						title="Отправить"
					>
						<svg
							className={cn(
								"h-5 w-5 transition-transform duration-200",
								hasText && "translate-x-0.5 -translate-y-0.5",
							)}
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
							/>
						</svg>
					</button>
				</div>
			)}
		</div>
	);
}
