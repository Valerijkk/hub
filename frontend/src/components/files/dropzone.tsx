import { toast } from "sonner";
import { useFileUpload } from "@/hooks/use-file-upload";
import { toastApiError } from "@/lib/format-api-error";
import { cn } from "@/lib/utils";
import { FileUpload } from "./file-upload";

/** Пропсы компонента Dropzone. */
interface DropzoneProps {
	/** ID проекта. */
	projectId: string;
	/** Тип сущности для привязки файла. */
	entityType: string;
	/** ID сущности для привязки файла. */
	entityId: string;
}

/** Зона перетаскивания файлов с индикатором загрузки и поддержкой Ctrl+V. */
export function Dropzone({ projectId, entityType, entityId }: DropzoneProps) {
	const { isDragging, isUploading, upload, dropZoneProps } = useFileUpload({
		projectId,
		entityType,
		entityId,
		onUpload: (_fileId, file) => {
			toast.success(`Файл "${file.name}" загружен`);
		},
		onError: (err) => toastApiError(err, "Не удалось загрузить файл"),
	});

	return (
		<div
			{...dropZoneProps}
			className={cn(
				"rounded-lg border-2 border-dashed p-8 text-center transition-colors",
				isDragging
					? "border-primary bg-primary/5"
					: "border-t2/25 hover:border-t2/50",
				isUploading && "pointer-events-none opacity-60",
			)}
		>
			{isUploading ? (
				<div className="space-y-2">
					<div className="mx-auto h-8 w-8 animate-spin rounded-full border-2 border-t2 border-t-primary" />
					<p className="text-sm text-t2">Загрузка...</p>
				</div>
			) : (
				<div className="space-y-3">
					<svg
						aria-hidden="true"
						className="mx-auto h-10 w-10 text-t2/50"
						width="40"
						height="40"
						viewBox="0 0 40 40"
						fill="none"
						stroke="currentColor"
						strokeWidth="2"
						strokeLinecap="round"
						strokeLinejoin="round"
					>
						<path d="M20 28V12M13 19l7-7 7 7" />
						<path d="M6 34h28" />
					</svg>
					<div>
						<p className="text-sm font-medium">
							{isDragging ? "Отпустите файл" : "Перетащите файлы сюда"}
						</p>
						<p className="text-xs text-t2">
							или используйте Ctrl+V для вставки из буфера обмена
						</p>
					</div>
					<FileUpload
						onFiles={(files) => void upload(files)}
						disabled={isUploading}
					/>
				</div>
			)}
		</div>
	);
}
