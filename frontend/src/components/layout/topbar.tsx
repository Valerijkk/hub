/**
 * Верхняя панель — glassmorphic topbar with backdrop blur.
 * Visible on all screen sizes. Hamburger shown only on mobile (md:hidden).
 * Optional title prop renders gradient text for page name.
 */
import { NotificationBell } from "@/components/notifications/notification-bell";
import { Button } from "@/components/ui/button";

interface TopbarProps {
	onMenuToggle?: () => void;
	title?: string;
}

export function Topbar({ onMenuToggle, title }: TopbarProps) {
	return (
		<header className="relative flex h-14 shrink-0 items-center gap-4 border-b border-b1 bg-[rgba(13,14,18,.75)] px-6 backdrop-blur-[24px] backdrop-saturate-[180%] z-50 sticky top-0">
			{onMenuToggle && (
				<Button
					variant="ghost"
					size="icon"
					className="h-8 w-8 md:hidden text-t2"
					onClick={onMenuToggle}
					aria-label="Меню"
				>
					<svg
						width="16"
						height="16"
						viewBox="0 0 16 16"
						fill="none"
						stroke="currentColor"
						strokeWidth="2"
						strokeLinecap="round"
					>
						<path d="M2 4h12M2 8h12M2 12h12" />
					</svg>
				</Button>
			)}

			{title && (
				<span className="text-[18px] font-[800] tracking-[-0.03em] bg-gradient-to-br from-amber-400 to-orange-400 bg-clip-text text-transparent">
					{title}
				</span>
			)}

			<div className="flex-1" />

			<div className="flex items-center gap-2">
				<NotificationBell />
			</div>

			{/* Gradient bottom edge */}
			<div
				className="pointer-events-none absolute bottom-[-1px] left-[10%] right-[10%] h-px bg-[linear-gradient(90deg,transparent,rgba(251,191,36,.06),transparent)]"
				aria-hidden="true"
			/>
		</header>
	);
}
