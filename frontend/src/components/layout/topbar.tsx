/**
 * Верхняя панель — glassmorphic topbar with backdrop blur.
 * На всех экранах; гамбургер только на mobile (md:hidden).
 * prop title — gradient-текст названия страницы.
 */
import { Search } from "lucide-react";
import { useTranslation } from "react-i18next";
import { NotificationBell } from "@/components/notifications/notification-bell";
import { Button } from "@/components/ui/button";
import { useCommandStore } from "@/stores/command-store";

interface TopbarProps {
	onMenuToggle?: () => void;
	title?: string;
}

export function Topbar({ onMenuToggle, title }: TopbarProps) {
	const { t } = useTranslation("common");
	const openCommand = useCommandStore((s) => s.setOpen);

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
						aria-hidden="true"
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
				<button
					type="button"
					onClick={() => openCommand(true)}
					aria-label={t("search.globalSearch")}
					aria-keyshortcuts="Control+K Meta+K"
					className="hidden md:flex h-8 min-w-[200px] items-center gap-2 rounded-[var(--r2)] border border-b1 bg-[var(--hover)] px-3 text-sm text-t3 transition-colors hover:text-t2"
				>
					<Search className="h-4 w-4 shrink-0" aria-hidden="true" />
					<span className="flex-1 text-left">
						{t("search.placeholderExtended")}
					</span>
					<span
						className="rounded bg-base px-1.5 py-0.5 font-mono text-[10px] text-t3"
						aria-hidden="true"
					>
						{t("search.hint")}
					</span>
				</button>
				<Button
					type="button"
					variant="ghost"
					size="icon"
					className="h-8 w-8 md:hidden text-t2"
					onClick={() => openCommand(true)}
					aria-label={t("search.globalSearch")}
				>
					<Search className="h-4 w-4" aria-hidden="true" />
				</Button>
				<NotificationBell />
			</div>

			{/* Градиентная нижняя граница */}
			<div
				className="pointer-events-none absolute bottom-[-1px] left-[10%] right-[10%] h-px bg-[linear-gradient(90deg,transparent,rgba(251,191,36,.06),transparent)]"
				aria-hidden="true"
			/>
		</header>
	);
}
