/**
 * Верхняя панель — минимальная, только hamburger для мобильных.
 * Профиль и уведомления перенесены в sidebar.
 */
import { Button } from "@/components/ui/button";

interface TopbarProps {
	onMenuToggle?: () => void;
}

export function Topbar({ onMenuToggle }: TopbarProps) {
	return (
		<header className="flex h-14 items-center justify-between border-b bg-card/70 backdrop-blur-md px-4 lg:px-6 shadow-[0_1px_3px_rgba(0,0,0,0.02)]">
			<div className="flex items-center gap-2">
				{onMenuToggle && (
					<Button
						variant="ghost"
						size="icon"
						className="h-8 w-8 md:hidden text-muted-foreground"
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
			</div>
			<div />
		</header>
	);
}
