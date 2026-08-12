import { useCallback, useRef, useState } from "react";
import { uploadFile } from "@/api/files";

/** Параметры хука useFileUpload. */
interface UseFileUploadOptions {
	/** ID проекта для привязки файла. */
	projectId: string;
	/** Тип сущности (issue, comment, test_case и т.д.). */
	entityType: string;
	/** ID сущности, к которой привязывается файл. */
	entityId: string;
	/** Callback после успешной загрузки файла. */
	onUpload?: (fileId: string, file: File) => void;
	/** Callback при ошибке загрузки. */
	onError?: (error: unknown) => void;
}

/** Возвращаемое значение хука useFileUpload. */
interface UseFileUploadReturn {
	/** Файл перетаскивается над зоной. */
	isDragging: boolean;
	/** Идёт загрузка файлов. */
	isUploading: boolean;
	/** Функция загрузки файлов программно. */
	upload: (files: FileList | File[]) => Promise<void>;
	/** Пропсы для зоны перетаскивания (drag-and-drop + paste). */
	dropZoneProps: {
		onDragEnter: (e: React.DragEvent) => void;
		onDragOver: (e: React.DragEvent) => void;
		onDragLeave: (e: React.DragEvent) => void;
		onDrop: (e: React.DragEvent) => void;
		onPaste: (e: React.ClipboardEvent) => void;
	};
}

/**
 * Извлекает файлы из clipboardData (paste).
 */
function filesFromClipboard(clipboardData: DataTransfer | null): File[] {
	const items = clipboardData?.items;
	if (!items) return [];

	const files: File[] = [];
	for (const item of items) {
		if (item.kind === "file") {
			const file = item.getAsFile();
			if (file) files.push(file);
		}
	}
	return files;
}

/**
 * useFileUpload предоставляет функциональность drag-and-drop и вставки файлов (Ctrl+V).
 *
 * dragCounter инкрементируется только на dragenter (не на dragover) — иначе
 * частые dragover-события «залипают» isDragging (W11-A3-006 / P2-FE-11).
 * Paste слушается только на drop-zone, не на document (W11-A3-007).
 *
 * @param options - параметры: проект, тип сущности, ID сущности, callbacks
 * @returns isDragging, isUploading, upload, dropZoneProps
 */
export function useFileUpload(
	options: UseFileUploadOptions,
): UseFileUploadReturn {
	const { projectId, entityType, entityId, onUpload, onError } = options;
	const [isDragging, setIsDragging] = useState(false);
	const [isUploading, setIsUploading] = useState(false);
	const dragCounter = useRef(0);

	const upload = useCallback(
		async (files: FileList | File[]) => {
			const fileArray = Array.from(files);
			if (fileArray.length === 0) return;

			setIsUploading(true);
			try {
				for (const file of fileArray) {
					const fileId = await uploadFile(
						file,
						projectId,
						entityType,
						entityId,
					);
					onUpload?.(fileId, file);
				}
			} catch (err) {
				onError?.(err);
			} finally {
				setIsUploading(false);
			}
		},
		[projectId, entityType, entityId, onUpload, onError],
	);

	const onDragEnter = useCallback((e: React.DragEvent) => {
		e.preventDefault();
		e.stopPropagation();
		dragCounter.current++;
		setIsDragging(true);
	}, []);

	// dragover только preventDefault — иначе counter ≫ enter (W11-A3-006).
	const onDragOver = useCallback((e: React.DragEvent) => {
		e.preventDefault();
		e.stopPropagation();
	}, []);

	const onDragLeave = useCallback((e: React.DragEvent) => {
		e.preventDefault();
		e.stopPropagation();
		dragCounter.current--;
		if (dragCounter.current === 0) {
			setIsDragging(false);
		}
	}, []);

	const onDrop = useCallback(
		(e: React.DragEvent) => {
			e.preventDefault();
			e.stopPropagation();
			dragCounter.current = 0;
			setIsDragging(false);

			if (e.dataTransfer.files.length > 0) {
				void upload(e.dataTransfer.files);
			}
		},
		[upload],
	);

	// Paste только на зоне (не document) — иначе hijack полей чата/форм (W11-A3-007).
	const onPaste = useCallback(
		(e: React.ClipboardEvent) => {
			const files = filesFromClipboard(e.clipboardData);
			if (files.length === 0) return;
			e.preventDefault();
			void upload(files);
		},
		[upload],
	);

	return {
		isDragging,
		isUploading,
		upload,
		dropZoneProps: { onDragEnter, onDragOver, onDragLeave, onDrop, onPaste },
	};
}
