import * as DialogPrimitive from "@radix-ui/react-dialog";
import { X } from "lucide-react";
import * as React from "react";

import { cn } from "@/lib/utils";

/**
 * Sheet — slide-in панель сбоку, построенная на @radix-ui/react-dialog.
 *
 * Зачем отдельно от Dialog: Dialog центрирует контент по экрану, а на мобилке
 * для боковых панелей (например, список каналов в мессенджере) нужен
 * полноэкранный слайд от края. Поддерживается focus-trap, Escape, ARIA — всё
 * от Radix.
 */

const Sheet = DialogPrimitive.Root;

const SheetTrigger = DialogPrimitive.Trigger;

const SheetClose = DialogPrimitive.Close;

const SheetPortal = DialogPrimitive.Portal;

const SheetOverlay = React.forwardRef<
	React.ElementRef<typeof DialogPrimitive.Overlay>,
	React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
	<DialogPrimitive.Overlay
		ref={ref}
		className={cn(
			"fixed inset-0 z-50 bg-black/65 backdrop-blur-[6px] data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
			className,
		)}
		{...props}
	/>
));
SheetOverlay.displayName = "SheetOverlay";

/** Поддерживаемые стороны панели. */
type SheetSide = "left" | "right" | "top" | "bottom";

const SIDE_CLASSES: Record<SheetSide, string> = {
	left: "left-0 top-0 bottom-0 h-full w-[85vw] max-w-[320px] border-r border-[var(--b1)] data-[state=open]:slide-in-from-left data-[state=closed]:slide-out-to-left",
	right:
		"right-0 top-0 bottom-0 h-full w-[85vw] max-w-[320px] border-l border-[var(--b1)] data-[state=open]:slide-in-from-right data-[state=closed]:slide-out-to-right",
	top: "left-0 right-0 top-0 w-full border-b border-[var(--b1)] data-[state=open]:slide-in-from-top data-[state=closed]:slide-out-to-top",
	bottom:
		"left-0 right-0 bottom-0 w-full border-t border-[var(--b1)] data-[state=open]:slide-in-from-bottom data-[state=closed]:slide-out-to-bottom",
};

interface SheetContentProps
	extends React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content> {
	/** Сторона, с которой выезжает панель. */
	side?: SheetSide;
	/** Показывать ли крестик-close в углу (по умолчанию false — обычно у Sheet свой header). */
	showClose?: boolean;
}

const SheetContent = React.forwardRef<
	React.ElementRef<typeof DialogPrimitive.Content>,
	SheetContentProps
>(
	(
		{ className, children, side = "left", showClose = false, ...props },
		ref,
	) => (
		<SheetPortal>
			<SheetOverlay />
			<DialogPrimitive.Content
				ref={ref}
				className={cn(
					"fixed z-50 flex flex-col bg-[var(--void)] shadow-2xl outline-none duration-[220ms] ease-[var(--spring)] data-[state=open]:animate-in data-[state=closed]:animate-out",
					SIDE_CLASSES[side],
					className,
				)}
				{...props}
			>
				{children}
				{showClose && (
					<DialogPrimitive.Close className="absolute right-3 top-3 rounded-[var(--r1)] text-t3 transition-colors duration-150 hover:text-t1 hover:bg-[var(--hover)] focus:outline-none focus:ring-2 focus:ring-ring">
						<X className="h-4 w-4" />
						<span className="sr-only">Закрыть</span>
					</DialogPrimitive.Close>
				)}
			</DialogPrimitive.Content>
		</SheetPortal>
	),
);
SheetContent.displayName = "SheetContent";

const SheetTitle = React.forwardRef<
	React.ElementRef<typeof DialogPrimitive.Title>,
	React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title>
>(({ className, ...props }, ref) => (
	<DialogPrimitive.Title
		ref={ref}
		className={cn("text-base font-semibold text-t1", className)}
		{...props}
	/>
));
SheetTitle.displayName = "SheetTitle";

const SheetDescription = React.forwardRef<
	React.ElementRef<typeof DialogPrimitive.Description>,
	React.ComponentPropsWithoutRef<typeof DialogPrimitive.Description>
>(({ className, ...props }, ref) => (
	<DialogPrimitive.Description
		ref={ref}
		className={cn("text-sm text-t2", className)}
		{...props}
	/>
));
SheetDescription.displayName = "SheetDescription";

export {
	Sheet,
	SheetClose,
	SheetContent,
	SheetDescription,
	SheetOverlay,
	SheetPortal,
	SheetTitle,
	SheetTrigger,
};
