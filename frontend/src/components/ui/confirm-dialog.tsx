/**
 * ConfirmDialog — controlled AlertDialog вместо window.confirm.
 *
 * Использование:
 *   <ConfirmDialog
 *     open={open}
 *     onOpenChange={setOpen}
 *     title="Удалить вебхук?"
 *     description="Это действие необратимо."
 *     confirmLabel="Удалить"
 *     onConfirm={() => doDelete()}
 *   />
 *
 * Тексты передаёт call-site на русском —
 * компонент не локализует, чтобы формулировки были явными.
 */
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "@/components/ui/alert-dialog";

export interface ConfirmDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	title: string;
	description?: string;
	confirmLabel?: string;
	cancelLabel?: string;
	/** true — кнопка подтверждения в destructive-стиле (красная). */
	destructive?: boolean;
	/** Вызывается при подтверждении; диалог закрывается после onConfirm. */
	onConfirm: () => void | Promise<void>;
	/** Блокирует кнопку подтверждения (например, во время запроса). */
	disabled?: boolean;
}

export function ConfirmDialog({
	open,
	onOpenChange,
	title,
	description,
	confirmLabel = "Подтвердить",
	cancelLabel = "Отмена",
	destructive = true,
	onConfirm,
	disabled,
}: ConfirmDialogProps) {
	return (
		<AlertDialog open={open} onOpenChange={onOpenChange}>
			<AlertDialogContent>
				<AlertDialogHeader>
					<AlertDialogTitle>{title}</AlertDialogTitle>
					{description && (
						<AlertDialogDescription>{description}</AlertDialogDescription>
					)}
				</AlertDialogHeader>
				<AlertDialogFooter>
					<AlertDialogCancel disabled={disabled}>
						{cancelLabel}
					</AlertDialogCancel>
					<AlertDialogAction
						disabled={disabled}
						className={
							destructive
								? undefined
								: "bg-blue hover:bg-blue/90 focus:ring-blue/50"
						}
						onClick={(e) => {
							// Блокируем auto-close Radix — onConfirm может быть async; закрываем вручную.
							e.preventDefault();
							void Promise.resolve(onConfirm()).finally(() =>
								onOpenChange(false),
							);
						}}
					>
						{confirmLabel}
					</AlertDialogAction>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	);
}
