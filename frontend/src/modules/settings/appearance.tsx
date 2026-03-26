/**
 * Страница настроек внешнего вида: тема, акцентный цвет, язык.
 *
 * Позволяет переключить тему (светлая/тёмная/системная),
 * выбрать акцентный цвет из 5 предустановленных,
 * и переключить язык интерфейса (RU/EN).
 */
import { useTranslation } from "react-i18next";
import { LanguageSwitcher } from "@/components/language-switcher";
import type { AccentColor } from "@/components/theme-provider";
import {
	ACCENT_COLORS,
	useAccentColor,
	useTheme,
} from "@/components/theme-provider";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";

/** CSS цвета для превью акцентных кружков */
const ACCENT_PREVIEWS: Record<AccentColor, string> = {
	red: "#ef4444",
	blue: "#3b82f6",
	green: "#22c55e",
	purple: "#a855f7",
	orange: "#f97316",
};

/** Конфигурация темы для карточки выбора */
const THEME_OPTIONS = [
	{ value: "light", icon: "sun" },
	{ value: "dark", icon: "moon" },
	{ value: "system", icon: "monitor" },
] as const;

/**
 * Компонент настроек внешнего вида
 */
export function AppearanceSettings() {
	const { t } = useTranslation("settings");
	const { theme, setTheme } = useTheme();
	const { accent, setAccent } = useAccentColor();

	return (
		<div className="space-y-6">
			{/* Тема */}
			<Card>
				<CardHeader>
					<CardTitle>{t("theme.title")}</CardTitle>
					<CardDescription>{t("theme.description")}</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="flex gap-[10px]">
						{THEME_OPTIONS.map((opt) => (
							<button
								key={opt.value}
								type="button"
								onClick={() => setTheme(opt.value)}
								className={cn(
									"flex flex-col items-center gap-2 rounded-lg border-2 p-3 transition-colors hover:bg-accent/50",
									theme === opt.value
										? "border-blue-500 bg-blue-500/[0.08]"
										: "border-transparent",
								)}
							>
								<ThemeIcon icon={opt.icon} />
								<span className="text-sm font-medium">
									{t(`theme.${opt.value}`)}
								</span>
							</button>
						))}
					</div>
				</CardContent>
			</Card>

			{/* Акцентный цвет */}
			<Card>
				<CardHeader>
					<CardTitle>{t("accent.title")}</CardTitle>
					<CardDescription>{t("accent.description")}</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="flex gap-3">
						{ACCENT_COLORS.map((color) => (
							<button
								key={color}
								type="button"
								onClick={() => setAccent(color)}
								className={cn(
									"flex items-center justify-center rounded-full transition-transform hover:scale-110",
									accent === color
										? "ring-2 ring-white ring-offset-2 ring-offset-background"
										: "",
								)}
								style={accent === color ? { boxShadow: `0 0 12px ${ACCENT_PREVIEWS[color]}` } : undefined}
								title={t(`accent.${color}`)}
							>
								<span
									className="h-7 w-7 rounded-full"
									style={{ backgroundColor: ACCENT_PREVIEWS[color] }}
								/>
							</button>
						))}
					</div>
				</CardContent>
			</Card>

			{/* Язык */}
			<Card>
				<CardHeader>
					<CardTitle>{t("lang.title")}</CardTitle>
					<CardDescription>{t("lang.description")}</CardDescription>
				</CardHeader>
				<CardContent>
					<LanguageSwitcher />
				</CardContent>
			</Card>
		</div>
	);
}

/** Иконки для карточек выбора темы */
function ThemeIcon({ icon }: { icon: string }) {
	if (icon === "sun") {
		return (
			<svg
				width="24"
				height="24"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				strokeWidth="2"
				strokeLinecap="round"
				strokeLinejoin="round"
			>
				<circle cx="12" cy="12" r="5" />
				<path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
			</svg>
		);
	}
	if (icon === "moon") {
		return (
			<svg
				width="24"
				height="24"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				strokeWidth="2"
				strokeLinecap="round"
				strokeLinejoin="round"
			>
				<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
			</svg>
		);
	}
	return (
		<svg
			width="24"
			height="24"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			strokeWidth="2"
			strokeLinecap="round"
			strokeLinejoin="round"
		>
			<rect x="2" y="3" width="20" height="14" rx="2" />
			<path d="M8 21h8M12 17v4" />
		</svg>
	);
}
