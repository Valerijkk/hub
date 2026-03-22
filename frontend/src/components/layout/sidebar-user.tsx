/**
 * Кнопка развернуть/свернуть сайдбар — нижний блок.
 */
import { Button } from "@/components/ui/button";

/** Кнопка развернуть сайдбар, отображаемая только в свёрнутом состоянии. */
export function SidebarExpandButton({
	collapsed,
	toggle,
	ariaLabel,
}: {
	collapsed: boolean;
	toggle: () => void;
	ariaLabel: string;
}) {
	if (!collapsed) return null;

	return (
		<div className="px-3 pb-3">
			<Button
				variant="ghost"
				size="icon"
				onClick={toggle}
				className="w-full h-7 text-slate-500 hover:text-white hover:bg-white/[0.06] rounded-md"
				aria-label={ariaLabel}
			>
				<svg
					width="14"
					height="14"
					viewBox="0 0 16 16"
					fill="none"
					stroke="currentColor"
					strokeWidth="2"
					strokeLinecap="round"
					strokeLinejoin="round"
				>
					<path d="M6 12l4-4-4-4" />
				</svg>
			</Button>
		</div>
	);
}
