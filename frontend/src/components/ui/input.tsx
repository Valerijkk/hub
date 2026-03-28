import * as React from "react";

import { cn } from "@/lib/utils";

const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<"input">>(
	({ className, type, ...props }, ref) => {
		return (
			<input
				type={type}
				className={cn(
					"flex h-[42px] w-full rounded-[var(--r2)] border border-b1 bg-[var(--input)] px-3 py-2 text-base text-t1 ring-offset-background transition-all duration-200 ease-[var(--ease)] file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-t3 focus-visible:outline-none focus-visible:border-[rgba(91,156,245,0.3)] focus-visible:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] focus-visible:scale-[1.005] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
					className,
				)}
				ref={ref}
				{...props}
			/>
		);
	},
);
Input.displayName = "Input";

export { Input };
