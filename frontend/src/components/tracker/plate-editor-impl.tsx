import type { Value } from "@udecode/plate";
import { Plate, PlateContent, usePlateEditor } from "@udecode/plate/react";
import { useEffect, useRef } from "react";
import { cn } from "@/lib/utils";

/** Пропсы компонента PlateEditor. */
interface PlateEditorProps {
	/** Начальное значение в формате Plate JSON. */
	value?: Value;
	/** Колбэк при изменении содержимого. */
	onChange?: (value: Value) => void;
	/** Текст-заглушка при пустом редакторе. */
	placeholder?: string;
	/** Режим только для чтения. */
	readOnly?: boolean;
	/** Дополнительные CSS-классы для контейнера. */
	className?: string;
}

const DEFAULT_VALUE: Value = [{ type: "p", children: [{ text: "" }] }];

/**
 * PlateEditor оборачивает Plate rich-text редактор с конфигурацией Hub.
 * Поддерживает базовое форматирование и режим чтения.
 * Хранит содержимое в формате Plate JSON для сохранения на бэкенде.
 */
export function PlateEditor({
	value,
	onChange,
	placeholder = "Введите текст...",
	readOnly = false,
	className,
}: PlateEditorProps) {
	const editor = usePlateEditor({
		value: value ?? DEFAULT_VALUE,
	});
	const lastExternalValue = useRef(JSON.stringify(value ?? DEFAULT_VALUE));

	// Синхронизация внешнего value после mount (смена issue/комментария) — W9-A4-005.
	useEffect(() => {
		const next = value ?? DEFAULT_VALUE;
		const serialized = JSON.stringify(next);
		if (serialized === lastExternalValue.current) return;
		lastExternalValue.current = serialized;
		editor.tf.setValue(next);
	}, [value, editor]);

	return (
		<div className={cn("rounded-md border", className)}>
			<Plate
				editor={editor}
				onValueChange={({ value: newValue }) => {
					lastExternalValue.current = JSON.stringify(newValue);
					onChange?.(newValue);
				}}
				readOnly={readOnly}
			>
				<PlateContent
					placeholder={placeholder}
					className={cn(
						"min-h-[120px] p-3 text-sm focus:outline-none",
						readOnly && "cursor-default",
					)}
				/>
			</Plate>
		</div>
	);
}
