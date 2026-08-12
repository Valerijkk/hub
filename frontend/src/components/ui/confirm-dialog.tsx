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
 *
 * W9-A5-003 / P2-FE-02: диалог закрывается только после успешного
 * onConfirm; при reject остаётся open (повторный confirm возможен).
 */
import { useState } from "react";
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
	/**
	 * Вызывается при подтверждении.
	 * Диалог закрывается только при успешном завершении (resolve);
	 * при reject остаётся открытым.
	 */
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
	const [confirming, setConfirming] = useState(false);
	const busy = Boolean(disabled) || confirming;

	return (
		<AlertDialog
			open={open}
			onOpenChange={(next) => {
				// Не закрывать по Escape/overlay, пока идёт async confirm.
				if (confirming && !next) return;
				onOpenChange(next);
			}}
		>
			<AlertDialogContent>
				<AlertDialogHeader>
					<AlertDialogTitle>{title}</AlertDialogTitle>
					{/* Always render Description so Radix associates aria-describedby (W9-A5-004). */}
					<AlertDialogDescription
						className={description ? undefined : "sr-only"}
					>
						{description ?? title}
					</AlertDialogDescription>
				</AlertDialogHeader>
				<AlertDialogFooter>
					<AlertDialogCancel disabled={busy}>{cancelLabel}</AlertDialogCancel>
					<AlertDialogAction
						disabled={busy}
						className={
							destructive
								? undefined
								: "bg-blue hover:bg-blue/90 focus:ring-blue/50"
						}
						onClick={(e) => {
							// Блокируем auto-close Radix — onConfirm может быть async.
							e.preventDefault();
							setConfirming(true);
							void (async () => {
								try {
									await onConfirm();
									onOpenChange(false);
								} catch {
									// Ошибка — диалог остаётся open (W9-A5-003).
								} finally {
									setConfirming(false);
								}
							})();
						}}
					>
						{confirmLabel}
					</AlertDialogAction>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	);
}
