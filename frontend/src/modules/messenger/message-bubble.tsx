import { useMemo } from "react";
import { LinkPreview } from "./link-preview";

// ---------------------------------------------------------------------------
// Типы
// ---------------------------------------------------------------------------

interface MessageBubbleProps {
	/** ID сообщения. */
	messageId: string;
	/** Имя отправителя. */
	senderName: string;
	/** URL аватара отправителя. */
	senderAvatarUrl?: string;
	/** JSON-контент сообщения (Plate editor). */
	contentJson?: Record<string, unknown>;
	/** Текстовое содержимое сообщения. */
	contentText: string;
	/** Является ли отправителем текущий пользователь. */
	isOwn: boolean;
	/** Время отправки. */
	createdAt: string;
	/** Было ли сообщение отредактировано. */
	isEdited?: boolean;
	/** Массив реакций. */
	reactions?: Array<{ emoji: string; count: number }>;
	/** ID пересланного сообщения. */
	forwardedFromId?: string;
}

// ---------------------------------------------------------------------------
// Утилиты
// ---------------------------------------------------------------------------

/** Максимальное количество превью ссылок на одно сообщение. */
const MAX_PREVIEWS = 3;

/** Regex для извлечения URL из текста. */
const urlRegex = /https?:\/\/[^\s<>"{}|\\^`[\]]+/g;

/**
 * extractUrls извлекает уникальные URL из текста.
 * Возвращает не более MAX_PREVIEWS URL с дедупликацией.
 */
function extractUrls(text: string): string[] {
	const matches = text.match(urlRegex);
	if (!matches) return [];

	const seen = new Set<string>();
	const unique: string[] = [];

	for (const url of matches) {
		// Убираем висящую пунктуацию в конце URL
		const clean = url.replace(/[.,;:!?)]+$/, "");
		if (!seen.has(clean)) {
			seen.add(clean);
			unique.push(clean);
			if (unique.length >= MAX_PREVIEWS) break;
		}
	}

	return unique;
}

/**
 * renderPlateContent выполняет простой рендеринг Plate JSON в текст.
 * Для read-only отображения в пузыре сообщения.
 */
function renderPlateContent(content: Record<string, unknown>): string {
	const children = content.children as
		| Array<Record<string, unknown>>
		| undefined;
	if (!children) return "";

	const parts: string[] = [];
	for (const node of children) {
		if (typeof node.text === "string") {
			parts.push(node.text);
		}
		const inner = node.children as Array<Record<string, unknown>> | undefined;
		if (inner) {
			for (const child of inner) {
				if (typeof child.text === "string") {
					parts.push(child.text);
				}
			}
		}
	}
	return parts.join("");
}

// ---------------------------------------------------------------------------
// Компонент
// ---------------------------------------------------------------------------

/**
 * MessageBubble -- пузырь сообщения в мессенджере.
 * После рендеринга текста извлекает URL и показывает LinkPreview (макс. 3, MSG-11).
 */
export function MessageBubble({
	senderName,
	senderAvatarUrl,
	contentJson,
	contentText,
	isOwn,
	createdAt,
	isEdited,
	reactions,
	forwardedFromId,
}: MessageBubbleProps) {
	// Текст для отображения
	const displayText = contentJson
		? renderPlateContent(contentJson)
		: contentText;

	// Извлечение URL для превью (мемоизировано)
	const urls = useMemo(() => extractUrls(contentText), [contentText]);

	// Форматирование времени
	const time = useMemo(() => {
		const d = new Date(createdAt);
		return d.toLocaleTimeString("ru-RU", {
			hour: "2-digit",
			minute: "2-digit",
		});
	}, [createdAt]);

	return (
		<div
			className={`flex gap-2 ${isOwn ? "flex-row-reverse" : "flex-row"} group`}
		>
			{/* Аватар */}
			{!isOwn && (
				<div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-gray-200 text-xs font-medium text-gray-600 dark:bg-gray-700 dark:text-gray-300">
					{senderAvatarUrl ? (
						<img
							src={senderAvatarUrl}
							alt={senderName}
							className="h-8 w-8 rounded-full object-cover"
						/>
					) : (
						(senderName[0]?.toUpperCase() ?? "?")
					)}
				</div>
			)}

			{/* Содержимое */}
			<div className={`max-w-[70%] ${isOwn ? "items-end" : "items-start"}`}>
				{/* Имя отправителя */}
				{!isOwn && (
					<p className="mb-0.5 text-xs font-medium text-gray-500 dark:text-gray-400">
						{senderName}
					</p>
				)}

				{/* Пузырь */}
				<div
					className={`rounded-2xl px-3 py-2 ${
						isOwn
							? "bg-blue-600 text-white"
							: "bg-gray-100 text-gray-900 dark:bg-gray-800 dark:text-gray-100"
					}`}
				>
					{/* Пересланное сообщение */}
					{forwardedFromId && (
						<p className="mb-1 text-xs italic opacity-70">
							Пересланное сообщение
						</p>
					)}

					{/* Текст */}
					<p className="whitespace-pre-wrap break-words text-sm">
						{displayText}
					</p>

					{/* Время + отредактировано */}
					<div
						className={`mt-1 flex items-center gap-1 text-[10px] ${
							isOwn ? "justify-end text-blue-200" : "text-gray-400"
						}`}
					>
						{isEdited && <span>(ред.)</span>}
						<span>{time}</span>
					</div>
				</div>

				{/* Реакции */}
				{reactions && reactions.length > 0 && (
					<div className="mt-1 flex flex-wrap gap-1">
						{reactions.map((r) => (
							<span
								key={r.emoji}
								className="inline-flex items-center gap-0.5 rounded-full bg-gray-100 px-1.5 py-0.5 text-xs dark:bg-gray-800"
							>
								{r.emoji} {r.count}
							</span>
						))}
					</div>
				)}

				{/* Превью ссылок (MSG-11) */}
				{urls.length > 0 && (
					<div className="mt-1 space-y-1">
						{urls.map((u) => (
							<LinkPreview key={u} url={u} />
						))}
					</div>
				)}
			</div>
		</div>
	);
}
