import { useCallback, useEffect, useRef, useState } from "react";
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
	/** Пропсы для зоны перетаскивания (drag-and-drop). */
	dropZoneProps: {
		onDragOver: (e: React.DragEvent) => void;
		onDragLeave: (e: React.DragEvent) => void;
		onDrop: (e: React.DragEvent) => void;
	};
}

/**
 * useFileUpload предоставляет функциональность drag-and-drop и вставки файлов (Ctrl+V).
 *
 * Обрабатывает события dragover, drop и paste. Загружает файлы
 * через presigned URL (получить URL -> PUT -> подтвердить).
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

	// Обработка глобальных событий вставки (Ctrl+V).
	useEffect(() => {
		const handlePaste = (e: ClipboardEvent) => {
			const items = e.clipboardData?.items;
			if (!items) return;

			const files: File[] = [];
			for (const item of items) {
				if (item.kind === "file") {
					const file = item.getAsFile();
					if (file) files.push(file);
				}
			}

			if (files.length > 0) {
				e.preventDefault();
				void upload(files);
			}
		};

		document.addEventListener("paste", handlePaste);
		return () => document.removeEventListener("paste", handlePaste);
	}, [upload]);

	const onDragOver = useCallback((e: React.DragEvent) => {
		e.preventDefault();
		e.stopPropagation();
		dragCounter.current++;
		setIsDragging(true);
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

	return {
		isDragging,
		isUploading,
		upload,
		dropZoneProps: { onDragOver, onDragLeave, onDrop },
	};
}
