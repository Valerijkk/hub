/**
 * Lazy-обёртка над тяжёлым Plate-редактором (@udecode/plate ~317 KB).
 *
 * Импорт plate-editor-impl.tsx через React.lazy исключает Plate из main-бандла:
 * vendor-plate грузится только когда пользователь реально открывает форму с
 * редактором (issue-form, comment-list, issue-description).
 */
import type { Value } from "@udecode/plate";
import { lazy, Suspense } from "react";
import { cn } from "@/lib/utils";

/** Пропсы компонента PlateEditor (зеркалят props plate-editor-impl). */
export interface PlateEditorProps {
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

const LazyPlateEditor = lazy(() =>
	import("./plate-editor-impl").then((m) => ({ default: m.PlateEditor })),
);

/** Скелет на время загрузки chunk'а Plate. */
function PlateEditorSkeleton({ className }: { className?: string }) {
	return (
		<div
			className={cn(
				"rounded-md border min-h-[120px] p-3 text-sm text-t3",
				className,
			)}
			aria-busy="true"
		>
			<div className="h-4 w-1/2 rounded animate-shimmer opacity-50" />
		</div>
	);
}

/**
 * PlateEditor — public façade. Динамически подгружает Plate при первом рендере.
 * API идентичен `plate-editor-impl#PlateEditor`.
 */
export function PlateEditor(props: PlateEditorProps) {
	return (
		<Suspense fallback={<PlateEditorSkeleton className={props.className} />}>
			<LazyPlateEditor {...props} />
		</Suspense>
	);
}
