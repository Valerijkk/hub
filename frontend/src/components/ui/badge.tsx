import { cva, type VariantProps } from "class-variance-authority";
import type * as React from "react";

import { cn } from "@/lib/utils";

const badgeVariants = cva(
	"inline-flex items-center rounded-full border px-[7px] py-[2px] text-[11px] font-bold leading-tight transition-colors",
	{
		variants: {
			variant: {
				default:
					"border-transparent bg-[var(--gB)] text-[var(--blue)] hover:bg-[var(--gB)]",
				secondary: "border-transparent bg-raised text-t2 hover:bg-raised/80",
				destructive:
					"border-transparent bg-[var(--gR)] text-[var(--red)] hover:bg-[var(--gR)]",
				outline: "text-foreground",
				success: "border-transparent bg-[var(--gG)] text-[var(--green)]",
				warning: "border-transparent bg-[var(--gP)] text-[var(--purple)]",
			},
		},
		defaultVariants: {
			variant: "default",
		},
	},
);

export interface BadgeProps
	extends React.HTMLAttributes<HTMLSpanElement>,
		VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
	return (
		<span className={cn(badgeVariants({ variant }), className)} {...props} />
	);
}

export { Badge, badgeVariants };
