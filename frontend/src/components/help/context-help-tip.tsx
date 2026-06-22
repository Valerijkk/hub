import { HelpCircle } from "lucide-react";
import { useTranslation } from "react-i18next";
import {
	Tooltip,
	TooltipContent,
	TooltipProvider,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

interface ContextHelpTipProps {
	/** Ключ i18n в help:contextTips (например "jql", "wip", "allure"). */
	tipKey: string;
	className?: string;
}

/**
 * Контекстная подсказка с tooltip (UX-P1-04).
 * Текст в help.json → contextTips.*
 */
export function ContextHelpTip({ tipKey, className }: ContextHelpTipProps) {
	const { t } = useTranslation("help");
	const label = t("contextTips.label");
	const content = t(`contextTips.${tipKey}`);

	return (
		<TooltipProvider delayDuration={200}>
			<Tooltip>
				<TooltipTrigger asChild>
					<button
						type="button"
						className={cn(
							"inline-flex shrink-0 items-center justify-center rounded-full text-t3 transition-colors hover:text-t2 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue/30",
							className,
						)}
						aria-label={label}
					>
						<HelpCircle className="h-3.5 w-3.5" aria-hidden="true" />
					</button>
				</TooltipTrigger>
				<TooltipContent
					side="top"
					className="max-w-xs text-[12px] leading-relaxed"
				>
					{content}
				</TooltipContent>
			</Tooltip>
		</TooltipProvider>
	);
}
