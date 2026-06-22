import { useCallback, useRef, useState } from "react";
import { toast } from "sonner";
import { uploadFile } from "@/api/files";
import { toastApiError } from "@/lib/format-api-error";
import { cn } from "@/lib/utils";

interface Attachment {
	file_id: string;
	file_name: string;
	file_size: number;
	mime_type: string;
	url?: string;
}

interface IssueAttachmentsProps {
	projectId: string;
	issueId: string;
	attachments: Attachment[];
	onAttachmentAdded?: () => void;
}

function formatFileSize(bytes: number): string {
	if (bytes < 1024) return `${bytes} B`;
	if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
	return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function getFileIcon(mimeType: string): string {
	if (mimeType.startsWith("image/")) return "\u{1F5BC}\uFE0F";
	if (mimeType.startsWith("video/")) return "\u{1F3AC}";
	if (mimeType.includes("pdf")) return "\u{1F4C4}";
	if (mimeType.includes("spreadsheet") || mimeType.includes("excel"))
		return "\u{1F4CA}";
	if (mimeType.includes("zip") || mimeType.includes("archive"))
		return "\u{1F4E6}";
	return "\u{1F4CE}";
}

export function IssueAttachments({
	projectId,
	issueId,
	attachments,
	onAttachmentAdded,
}: IssueAttachmentsProps) {
	const [isUploading, setIsUploading] = useState(false);
	const fileInputRef = useRef<HTMLInputElement>(null);

	const handleFileChange = useCallback(
		async (e: React.ChangeEvent<HTMLInputElement>) => {
			const files = e.target.files;
			if (!files || files.length === 0) return;

			setIsUploading(true);
			try {
				for (const file of Array.from(files)) {
					await uploadFile(file, projectId, "issue", issueId);
				}
				toast.success(
					files.length === 1
						? "Файл загружен"
						: `${files.length} файлов загружено`,
				);
				onAttachmentAdded?.();
			} catch (err) {
				toastApiError(err, "Не удалось загрузить файл");
			} finally {
				setIsUploading(false);
				if (fileInputRef.current) fileInputRef.current.value = "";
			}
		},
		[projectId, issueId, onAttachmentAdded],
	);

	return (
		<div className="space-y-3">
			<div className="flex items-center justify-between">
				<h3 className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3">
					Вложения
					{attachments.length > 0 && (
						<span className="ml-1.5 text-[11px] text-t4">
							({attachments.length})
						</span>
					)}
				</h3>
				<button
					type="button"
					onClick={() => fileInputRef.current?.click()}
					disabled={isUploading}
					className={cn(
						"inline-flex items-center gap-1.5 rounded-lg border px-3 py-1.5 text-xs font-medium transition-colors",
						"hover:bg-accent hover:text-accent-foreground",
						isUploading && "opacity-50 cursor-not-allowed",
					)}
				>
					{isUploading ? (
						<>
							<svg
								aria-hidden="true"
								className="h-3.5 w-3.5 animate-spin"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								strokeWidth={2}
							>
								<path d="M12 2v4m0 12v4m-7.07-3.93l2.83-2.83m8.49-8.49l2.83-2.83M2 12h4m12 0h4" />
							</svg>
							Загрузка...
						</>
					) : (
						<>
							<svg
								aria-hidden="true"
								className="h-3.5 w-3.5"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								strokeWidth={2}
							>
								<path
									strokeLinecap="round"
									strokeLinejoin="round"
									d="M12 4v16m8-8H4"
								/>
							</svg>
							Добавить файл
						</>
					)}
				</button>
				<input
					ref={fileInputRef}
					type="file"
					multiple
					className="hidden"
					onChange={(e) => void handleFileChange(e)}
				/>
			</div>

			{attachments.length > 0 && (
				<div className="flex gap-[10px] flex-wrap">
					{attachments.map((att) => (
						<a
							key={att.file_id}
							href={att.url || `/api/files/${att.file_id}/download`}
							target="_blank"
							rel="noopener noreferrer"
							className="flex items-center gap-[10px] px-4 py-[10px] bg-gradient-to-br from-card to-raised border border-b1 rounded-[var(--r3)] text-[13px] cursor-pointer transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 hover:border-b2 hover:-translate-y-px hover:shadow-[var(--sh2)]"
						>
							<span className="w-[18px] h-[18px] text-blue">
								{getFileIcon(att.mime_type)}
							</span>
							<div className="flex-1 min-w-0">
								<p className="text-[13px] font-medium truncate text-t1 transition-colors">
									{att.file_name}
								</p>
								<p className="text-t3 text-[11px]">
									{formatFileSize(att.file_size)}
								</p>
							</div>
							<svg
								aria-hidden="true"
								className="h-4 w-4 text-t2/50 group-hover:text-primary shrink-0 transition-colors"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								strokeWidth={2}
							>
								<path
									strokeLinecap="round"
									strokeLinejoin="round"
									d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
								/>
							</svg>
						</a>
					))}
				</div>
			)}

			{attachments.length === 0 && (
				<p className="text-xs text-t2 py-2">
					Перетащите файлы или нажмите "Добавить файл"
				</p>
			)}
		</div>
	);
}
