import { useQueryClient } from "@tanstack/react-query";
import { type KeyboardEvent, useCallback, useRef, useState } from "react";
import { toast } from "sonner";
import { uploadFile } from "@/api/files";
import { sendMessage, sendTypingIndicator } from "@/api/messenger";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";
import { useMessengerStore } from "@/stores/messenger-store";
import { CreatePollDialog } from "./create-poll-dialog";
import { GifPicker } from "./gif-picker";
import { MentionAutocomplete } from "./mention-autocomplete";
import { type StickerData, StickerPicker, buildStickerContentJson } from "./sticker-picker";
import { VoiceRecorder } from "./voice-recorder";

/** Набор популярных эмодзи для быстрого выбора в композере. */
const POPULAR_EMOJIS = [
	"\u{1F44D}", // thumbs up
	"\u{1F44E}", // thumbs down
	"\u{2764}\u{FE0F}", // red heart
	"\u{1F602}", // laughing
	"\u{1F622}", // crying
	"\u{1F632}", // astonished
	"\u{1F525}", // fire
	"\u{1F4AF}", // 100
	"\u{1F440}", // eyes
	"\u{1F389}", // party popper
	"\u{1F680}", // rocket
	"\u{2705}", // check mark
	"\u{274C}", // cross mark
	"\u{1F914}", // thinking
	"\u{1F44F}", // clapping
	"\u{1F60D}", // heart eyes
	"\u{1F64F}", // folded hands
	"\u{1F4A1}", // light bulb
	"\u{26A0}\u{FE0F}", // warning
	"\u{1F3AF}", // bullseye
];

/** Текстовый узел Plate JSON. */
interface PlateTextNode {
	text: string;
	bold?: boolean;
	italic?: boolean;
	code?: boolean;
}

/** Элементный узел Plate JSON. */
interface PlateElementNode {
	type: string;
	children: (PlateTextNode | PlateElementNode)[];
}

/**
 * Парсит строку текста в массив Plate JSON text-узлов с поддержкой
 * базового markdown: **bold**, *italic* / _italic_, `code`.
 */
function parseMarkdownToPlateChildren(line: string): PlateTextNode[] {
	const nodes: PlateTextNode[] = [];
	// Регулярное выражение: `code`, **bold**, *italic*, _italic_
	const regex = /(`[^`]+`|\*\*[^*]+\*\*|\*[^*]+\*|_[^_]+_)/g;
	let lastIndex = 0;

	for (let match = regex.exec(line); match !== null; match = regex.exec(line)) {
		// Текст до совпадения.
		if (match.index > lastIndex) {
			nodes.push({ text: line.slice(lastIndex, match.index) });
		}

		const token = match[0];
		if (token.startsWith("`")) {
			nodes.push({ text: token.slice(1, -1), code: true });
		} else if (token.startsWith("**")) {
			nodes.push({ text: token.slice(2, -2), bold: true });
		} else if (token.startsWith("*")) {
			nodes.push({ text: token.slice(1, -1), italic: true });
		} else if (token.startsWith("_")) {
			nodes.push({ text: token.slice(1, -1), italic: true });
		}

		lastIndex = match.index + token.length;
	}

	// Хвост после последнего совпадения.
	if (lastIndex < line.length) {
		nodes.push({ text: line.slice(lastIndex) });
	}

	// Если ничего не нашлось -- простой текст.
	if (nodes.length === 0) {
		nodes.push({ text: line });
	}

	return nodes;
}

/**
 * Преобразует текст с базовым markdown в Plate JSON документ.
 * Каждая строка -- отдельный параграф.
 */
function textToPlateJson(raw: string): PlateElementNode {
	const lines = raw.split("\n");
	const children: PlateElementNode[] = lines.map((line) => ({
		type: "p" as const,
		children: parseMarkdownToPlateChildren(line),
	}));
	return { type: "doc", children };
}

/** Информация о прикреплённом файле (до отправки). */
interface PendingAttachment {
	/** ID файла в хранилище. */
	fileId: string;
	/** Имя файла. */
	fileName: string;
	/** MIME-тип файла. */
	mimeType: string;
	/** Размер файла в байтах. */
	fileSize: number;
}

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
 * автоподгонка высоты, голосовые сообщения, эмодзи, вложения.
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
	const [emojiOpen, setEmojiOpen] = useState(false);
	const [pendingAttachments, setPendingAttachments] = useState<
		PendingAttachment[]
	>([]);
	const [isUploading, setIsUploading] = useState(false);
	const [showPollDialog, setShowPollDialog] = useState(false);
	/** Позиция курсора для @mention автокомплита. */
	const [cursorPosition, setCursorPosition] = useState(0);
	/** Накопленные user_id упоминаний для отправки с сообщением. */
	const [mentionedUserIds, setMentionedUserIds] = useState<string[]>([]);
	const textareaRef = useRef<HTMLTextAreaElement>(null);
	const fileInputRef = useRef<HTMLInputElement>(null);
	/** Ref на обработчик клавиш MentionAutocomplete. */
	const mentionKeyHandlerRef = useRef<
		((e: KeyboardEvent<HTMLTextAreaElement>) => boolean) | null
	>(null);

	const replyToMessage = useMessengerStore((s) => s.replyToMessage);
	const setReplyTo = useMessengerStore((s) => s.setReplyTo);

	/** Время последней отправки typing-индикатора (троттлинг 3 сек). */
	const lastTypingSentRef = useRef(0);

	/** Вставка эмодзи в текст на позицию курсора. */
	const handleEmojiSelect = useCallback(
		(emoji: string) => {
			const ta = textareaRef.current;
			if (ta) {
				const start = ta.selectionStart;
				const end = ta.selectionEnd;
				const before = text.slice(0, start);
				const after = text.slice(end);
				const newText = before + emoji + after;
				setText(newText);
				// Восстанавливаем позицию курсора после вставки.
				requestAnimationFrame(() => {
					ta.focus();
					const newPos = start + emoji.length;
					ta.setSelectionRange(newPos, newPos);
				});
			} else {
				setText((prev) => prev + emoji);
			}
			setEmojiOpen(false);
		},
		[text],
	);

	/** Открытие диалога выбора файла. */
	const handleAttachClick = useCallback(() => {
		fileInputRef.current?.click();
	}, []);

	/** Загрузка выбранного файла в хранилище. */
	const handleFileChange = useCallback(
		async (e: React.ChangeEvent<HTMLInputElement>) => {
			const files = e.target.files;
			if (!files || files.length === 0) return;

			setIsUploading(true);
			try {
				for (const file of Array.from(files)) {
					const fileId = await uploadFile(
						file,
						"messenger",
						"attachment",
						channelId,
					);
					setPendingAttachments((prev) => [
						...prev,
						{
							fileId,
							fileName: file.name,
							mimeType: file.type || "application/octet-stream",
							fileSize: file.size,
						},
					]);
				}
			} catch {
				toast.error("Не удалось загрузить файл");
			} finally {
				setIsUploading(false);
				// Сбрасываем input, чтобы можно было выбрать тот же файл повторно.
				if (fileInputRef.current) {
					fileInputRef.current.value = "";
				}
			}
		},
		[channelId],
	);

	/** Удалить прикреплённый файл из списка. */
	const removePendingAttachment = useCallback((fileId: string) => {
		setPendingAttachments((prev) => prev.filter((a) => a.fileId !== fileId));
	}, []);

	/** Отправка текстового сообщения. */
	const handleSend = useCallback(async () => {
		const trimmed = text.trim();
		if (!trimmed && pendingAttachments.length === 0) return;

		setIsSending(true);
		try {
			const contentJson = (trimmed
				? textToPlateJson(trimmed)
				: {
						type: "doc",
						children: [{ type: "p", children: [{ text: "" }] }],
					}) as unknown as Record<string, unknown>;

			await sendMessage(channelId, {
				content_json: contentJson,
				content_text: trimmed || undefined,
				thread_id: threadId,
				attachment_ids:
					pendingAttachments.length > 0
						? pendingAttachments.map((a) => a.fileId)
						: undefined,
				mentions:
					mentionedUserIds.length > 0 ? mentionedUserIds : undefined,
			});
			void queryClient.invalidateQueries({
				queryKey: ["messenger", "messages", channelId],
			});
			setText("");
			setPendingAttachments([]);
			setMentionedUserIds([]);
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
	}, [channelId, threadId, text, pendingAttachments, mentionedUserIds, queryClient, setReplyTo]);

	/** Обработка выбора упоминания из автокомплита. */
	const handleMentionSelect = useCallback(
		(newText: string, newCursorPos: number, userId: string) => {
			setText(newText);
			setCursorPosition(newCursorPos);
			// Добавляем user_id в список упоминаний (без дубликатов).
			setMentionedUserIds((prev) =>
				prev.includes(userId) ? prev : [...prev, userId],
			);
			// Восстанавливаем фокус и курсор в textarea.
			requestAnimationFrame(() => {
				const ta = textareaRef.current;
				if (ta) {
					ta.focus();
					ta.setSelectionRange(newCursorPos, newCursorPos);
				}
			});
		},
		[],
	);

	/** Обработка клавиш: сначала проверяем @mention, затем Enter -- отправить. */
	const handleKeyDown = useCallback(
		(e: KeyboardEvent<HTMLTextAreaElement>) => {
			// Если автокомплит @mention обработал клавишу — не делаем ничего.
			if (mentionKeyHandlerRef.current?.(e)) return;

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
			setCursorPosition(e.target.selectionStart);
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

	/** Отправка GIF. */
	const handleGifSend = useCallback(
		async (gifUrl: string) => {
			try {
				const contentJson = {
					type: "doc",
					children: [
						{
							type: "p",
							children: [
								{
									type: "a",
									url: gifUrl,
									children: [{ text: gifUrl }],
								},
							],
						},
					],
				} as unknown as Record<string, unknown>;
				await sendMessage(channelId, {
					content_json: contentJson,
					content_text: gifUrl,
					thread_id: threadId,
				});
				void queryClient.invalidateQueries({
					queryKey: ["messenger", "messages", channelId],
				});
			} catch (err) {
				const message =
					err instanceof Error
						? err.message
						: "Не удалось отправить GIF";
				toast.error(message);
			}
		},
		[channelId, threadId, queryClient],
	);

	/** Отправка стикера. */
	const handleStickerSend = useCallback(
		async (sticker: StickerData) => {
			try {
				const contentJson = buildStickerContentJson(sticker);
				await sendMessage(channelId, {
					content_json: contentJson,
					content_text: `[sticker: ${sticker.label}]`,
					thread_id: threadId,
				});
				void queryClient.invalidateQueries({
					queryKey: ["messenger", "messages", channelId],
				});
			} catch (err) {
				const message =
					err instanceof Error
						? err.message
						: "Не удалось отправить стикер";
				toast.error(message);
			}
		},
		[channelId, threadId, queryClient],
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

	const hasContent = text.trim().length > 0 || pendingAttachments.length > 0;

	return (
		<div className="space-y-0">
			{/* Reply preview */}
			{replyToMessage && (
				<div className="flex items-center gap-2.5 px-3.5 py-2.5 border-l-[3px] border-l-[var(--green)] bg-[var(--raised)] rounded-r-[var(--r2)] mb-2.5 animate-fade-in">
					<div className="flex-1 min-w-0">
						<p className="text-[13px] font-bold text-[var(--green)] truncate">
							{replyToMessage.sender_id}
						</p>
						<p className="text-[12.5px] text-[var(--t3)] truncate flex-1">
							{replyPreviewText}
						</p>
					</div>
					<button
						type="button"
						onClick={() => setReplyTo(null)}
						className="w-[26px] h-[26px] flex items-center justify-center rounded-full text-[var(--t3)] hover:bg-[var(--hover)] hover:text-[var(--t1)] hover:rotate-90 transition-all"
						title="Отменить ответ"
					>
						<svg
							className="h-3.5 w-3.5"
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

			{/* Превью прикреплённых файлов */}
			{pendingAttachments.length > 0 && (
				<div className="flex flex-wrap gap-1.5 px-1 pb-1.5">
					{pendingAttachments.map((att) => (
						<div
							key={att.fileId}
							className="inline-flex items-center gap-1.5 rounded-lg border bg-card px-2.5 py-1.5 text-xs shadow-sm"
						>
							<svg
								className="h-3.5 w-3.5 text-t2/60 shrink-0"
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
							<span className="font-medium truncate max-w-[160px]">
								{att.fileName}
							</span>
							<button
								type="button"
								onClick={() => removePendingAttachment(att.fileId)}
								className="shrink-0 h-4 w-4 flex items-center justify-center rounded-full hover:bg-raised transition-colors"
								title="Убрать вложение"
							>
								<svg
									className="h-3 w-3 text-t2"
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
					))}
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
					className="flex items-end gap-1 bg-[var(--card)] border border-[var(--b1)] rounded-[var(--r4)] px-4 py-1.5 transition-all relative overflow-hidden focus-within:border-[rgba(52,211,153,.25)] focus-within:shadow-[0_0_0_3px_rgba(52,211,153,.05),0_0_24px_rgba(52,211,153,.04)]"
				>
					{/* Top highlight */}
					<div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.03] to-transparent" />

					{/* @Mention автокомплит */}
					<MentionAutocomplete
						channelId={channelId}
						text={text}
						cursorPosition={cursorPosition}
						textareaRef={textareaRef}
						onSelect={handleMentionSelect}
						keyHandlerRef={mentionKeyHandlerRef}
					/>

					{/* Attachment button */}
					<button
						type="button"
						className={cn(
							"w-[38px] h-[38px] flex items-center justify-center rounded-[var(--r2)] text-[var(--t3)] shrink-0 hover:bg-[var(--hover)] hover:text-[var(--t2)] hover:scale-110 transition-all",
							isUploading && "animate-pulse",
						)}
						title="Прикрепить файл"
						onClick={handleAttachClick}
						disabled={isUploading}
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

					{/* Скрытый input для выбора файлов */}
					<input
						ref={fileInputRef}
						type="file"
						multiple
						className="hidden"
						onChange={(e) => void handleFileChange(e)}
					/>

					{/* Text input */}
					<textarea
						ref={textareaRef}
						value={text}
						onChange={handleInput}
						onKeyDown={handleKeyDown}
						onSelect={(e) => setCursorPosition(e.currentTarget.selectionStart)}
						placeholder={placeholder}
						rows={1}
						disabled={isSending}
						className="flex-1 bg-transparent border-none text-sm text-[var(--t1)] resize-none min-h-[38px] max-h-[200px] py-[9px] leading-relaxed outline-none placeholder:text-[var(--t3)]"
					/>

					{/* Emoji button with Popover */}
					<Popover open={emojiOpen} onOpenChange={setEmojiOpen}>
						<PopoverTrigger asChild>
							<button
								type="button"
								className="w-[38px] h-[38px] flex items-center justify-center rounded-[var(--r2)] text-[var(--t3)] shrink-0 hover:bg-[var(--hover)] hover:text-[var(--t2)] hover:scale-110 transition-all"
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
						</PopoverTrigger>
						<PopoverContent
							className="w-auto p-2"
							align="end"
							side="top"
							sideOffset={8}
						>
							<div className="grid grid-cols-5 gap-1">
								{POPULAR_EMOJIS.map((emoji) => (
									<button
										key={emoji}
										type="button"
										className="flex h-8 w-8 items-center justify-center rounded text-lg hover:bg-accent transition-colors"
										onClick={() => handleEmojiSelect(emoji)}
									>
										{emoji}
									</button>
								))}
							</div>
						</PopoverContent>
					</Popover>

					{/* GIF picker */}
					<GifPicker onSelect={(url) => void handleGifSend(url)} />

					{/* Sticker picker */}
					<StickerPicker onSelect={(s) => void handleStickerSend(s)} />

					{/* Poll button */}
					<button
						type="button"
						className="w-[38px] h-[38px] flex items-center justify-center rounded-[var(--r2)] text-[var(--t3)] shrink-0 hover:bg-[var(--hover)] hover:text-[var(--t2)] hover:scale-110 transition-all"
						title="Опрос"
						onClick={() => setShowPollDialog(true)}
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
								d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
							/>
						</svg>
					</button>

					{/* Voice recorder button */}
					<button
						type="button"
						className="w-[38px] h-[38px] flex items-center justify-center rounded-[var(--r2)] text-[var(--t3)] shrink-0 hover:bg-[var(--hover)] hover:text-[var(--t2)] hover:scale-110 transition-all"
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
							"w-[38px] h-[38px] flex items-center justify-center rounded-[var(--r2)] shrink-0 transition-all",
							hasContent
								? "bg-gradient-to-br from-[var(--green)] to-[var(--greenS)] text-white shadow-[0_2px_12px_var(--gG)] hover:scale-[1.08] hover:shadow-[0_4px_20px_rgba(52,211,153,.25)] active:scale-[0.92]"
								: "bg-[var(--input)] text-[var(--t4)]",
						)}
						onClick={() => void handleSend()}
						disabled={isSending || !hasContent}
						title="Отправить"
					>
						<svg
							className={cn(
								"h-5 w-5 transition-transform duration-200",
								hasContent && "translate-x-0.5 -translate-y-0.5",
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

			{/* Диалог создания опроса */}
			<CreatePollDialog
				channelId={channelId}
				open={showPollDialog}
				onOpenChange={setShowPollDialog}
			/>
		</div>
	);
}
