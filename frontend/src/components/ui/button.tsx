import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";

import { cn } from "@/lib/utils";

const buttonVariants = cva(
	"inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-[var(--r2)] text-sm font-bold transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 ease-[var(--spring)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 active:scale-[.96] [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
	{
		variants: {
			variant: {
				default:
					"bg-gradient-to-br from-[var(--blueS)] to-[var(--blue)] text-white shadow-[0_0_20px_rgba(91,156,245,0.25)] hover:shadow-[0_0_28px_rgba(91,156,245,0.4)] hover:-translate-y-px hover:brightness-110",
				destructive:
					"bg-gradient-to-br from-[var(--redS)] to-[var(--red)] text-white shadow-[0_0_16px_rgba(239,68,68,0.2)] hover:shadow-[0_0_24px_rgba(239,68,68,0.35)] hover:-translate-y-px",
				outline:
					"border border-b1 bg-transparent hover:bg-[var(--hover)] hover:border-b2 hover:text-accent-foreground hover:-translate-y-px",
				secondary:
					"bg-raised text-t2 border border-b1 hover:bg-[var(--hover)] hover:text-t1 hover:border-b2 hover:-translate-y-px",
				ghost: "hover:bg-[var(--hover)] hover:text-accent-foreground",
				link: "text-primary underline-offset-4 hover:underline",
			},
			size: {
				default: "h-10 px-4 py-2",
				sm: "h-9 rounded-[var(--r2)] px-3",
				lg: "h-11 rounded-[var(--r2)] px-8",
				icon: "h-10 w-10",
			},
		},
		defaultVariants: {
			variant: "default",
			size: "default",
		},
	},
);

export interface ButtonProps
	extends React.ButtonHTMLAttributes<HTMLButtonElement>,
		VariantProps<typeof buttonVariants> {
	asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
	({ className, variant, size, asChild = false, type, ...props }, ref) => {
		const Comp = asChild ? Slot : "button";
		return (
			<Comp
				className={cn(buttonVariants({ variant, size, className }))}
				ref={ref}
				// Native <button> defaults to type=submit inside forms — force button
				// unless call-site opts into submit/reset. Slot path leaves type alone.
				{...(asChild ? {} : { type: type ?? "button" })}
				{...props}
			/>
		);
	},
);
Button.displayName = "Button";

export { Button, buttonVariants };
