import * as React from "react";

import { cn } from "@/lib/utils";

const Textarea = React.forwardRef<
	HTMLTextAreaElement,
	React.ComponentProps<"textarea">
>(({ className, ...props }, ref) => {
	return (
		<textarea
			className={cn(
				"flex min-h-[80px] w-full rounded-[var(--r2)] border border-b1 bg-[var(--input)] px-3 py-2 text-base text-t1 ring-offset-background transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 ease-[var(--ease)] placeholder:text-t3 focus-visible:outline-none focus-visible:border-[rgba(91,156,245,0.3)] focus-visible:shadow-[0_0_0_3px_rgba(91,156,245,0.06)] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
				className,
			)}
			ref={ref}
			{...props}
		/>
	);
});
Textarea.displayName = "Textarea";

export { Textarea };
