import type { Message } from "@/api/messenger";
import { cn } from "@/lib/utils";

/** Проверяет, является ли вложение изображением по MIME-типу. */
function isImageAttachment(mimeType: string): boolean {
	return mimeType.startsWith("image/");
}

/** Пропсы компонента MessageAttachments. */
interface MessageAttachmentsProps {
	attachments: NonNullable<Message["attachments"]>;
	isOwn: boolean;
}

/** Отображает вложения сообщения: превью изображений и ссылки для скачивания файлов. */
export function MessageAttachments({
	attachments,
	isOwn,
}: MessageAttachmentsProps) {
	if (attachments.length === 0) return null;

	return (
		<div className="mt-2 space-y-2">
			{/* Image previews */}
			{attachments
				.filter((att) => isImageAttachment(att.mime_type))
				.map((att) => (
					<a
						key={att.file_id}
						href={att.url}
						target="_blank"
						rel="noopener noreferrer"
						className="block overflow-hidden rounded-lg"
					>
						<img
							src={att.url}
							alt={att.file_name}
							className="max-w-full max-h-64 rounded-lg object-cover hover:opacity-90 transition-opacity"
							loading="lazy"
						/>
					</a>
				))}
			{/* File links */}
			<div className="flex flex-wrap gap-1.5">
				{attachments
					.filter((att) => !isImageAttachment(att.mime_type))
					.map((att) => (
						<a
							key={att.file_id}
							href={att.url}
							target="_blank"
							rel="noopener noreferrer"
							className={cn(
								"inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs transition-colors",
								isOwn
									? "bg-white/15 hover:bg-white/25"
									: "border bg-card hover:bg-raised shadow-sm",
							)}
						>
							<svg
								aria-hidden="true"
								className="h-3.5 w-3.5 opacity-60"
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
							<span className="font-medium truncate max-w-[200px]">
								{att.file_name}
							</span>
						</a>
					))}
			</div>
		</div>
	);
}
