/**
 * ConfirmDialog — controlled AlertDialog wrapper that replaces window.confirm.
 *
 * Usage:
 *   <ConfirmDialog
 *     open={open}
 *     onOpenChange={setOpen}
 *     title="Удалить вебхук?"
 *     description="Это действие необратимо."
 *     confirmLabel="Удалить"
 *     onConfirm={() => doDelete()}
 *   />
 *
 * Body text and action labels are provided in Russian by the caller — this
 * component intentionally does not localize so each call-site stays explicit.
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
	/** When true, the confirm button uses destructive styling (red). */
	destructive?: boolean;
	/** Called when the user confirms. Dialog closes automatically afterwards. */
	onConfirm: () => void | Promise<void>;
	/** Disable the confirm button (e.g. while a request is in flight). */
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
							// Prevent Radix auto-close so onConfirm can run async; we close manually.
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
