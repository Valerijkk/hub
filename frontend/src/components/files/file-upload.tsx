import { useId, useRef } from "react";
import { Button } from "@/components/ui/button";

/** Пропсы компонента FileUpload. */
interface FileUploadProps {
	/** Колбэк при выборе файлов. */
	onFiles: (files: FileList) => void;
	/** Допустимые MIME-типы/расширения файлов. */
	accept?: string;
	/** Разрешить выбор нескольких файлов. */
	multiple?: boolean;
	/** Заблокировать кнопку. */
	disabled?: boolean;
}

/** Кнопка загрузки файла со скрытым input[type=file]. */
export function FileUpload({
	onFiles,
	accept,
	multiple = true,
	disabled = false,
}: FileUploadProps) {
	const inputRef = useRef<HTMLInputElement>(null);
	const inputId = useId();

	function handleClick() {
		inputRef.current?.click();
	}

	function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
		if (e.target.files && e.target.files.length > 0) {
			onFiles(e.target.files);
			// Сброс input — иначе повторный выбор того же файла не сработает
			e.target.value = "";
		}
	}

	return (
		<>
			<Button
				type="button"
				variant="outline"
				onClick={handleClick}
				disabled={disabled}
			>
				<svg
					aria-hidden="true"
					className="mr-2 h-4 w-4"
					width="16"
					height="16"
					viewBox="0 0 16 16"
					fill="none"
					stroke="currentColor"
					strokeWidth="2"
					strokeLinecap="round"
					strokeLinejoin="round"
				>
					<path d="M8 12V3M4.5 6.5L8 3l3.5 3.5M2 14h12" />
				</svg>
				Загрузить файл
			</Button>
			<input
				ref={inputRef}
				id={inputId}
				type="file"
				accept={accept}
				multiple={multiple}
				onChange={handleChange}
				aria-label="Загрузить файл"
				tabIndex={-1}
				className="hidden"
			/>
		</>
	);
}
