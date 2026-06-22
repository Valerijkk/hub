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
		<div className="space-y-0">
			{/* Тема */}
			<div className="mb-9">
				<h2
					className="text-[18px] font-extrabold tracking-[-0.02em] mb-1"
					data-testid="appearance-theme-heading"
				>
					{t("theme.title")}
				</h2>
				<p className="text-[13px] text-[var(--t2)] mb-5">
					{t("theme.description")}
				</p>
				<div
					className="flex gap-[14px] mb-6"
					role="radiogroup"
					aria-label={t("theme.title")}
				>
					{THEME_OPTIONS.map((opt) => (
						<button
							key={opt.value}
							type="button"
							aria-pressed={theme === opt.value}
							onClick={() => setTheme(opt.value)}
							className={cn(
								"flex-1 p-[18px] bg-[var(--card)] border-2 border-b1 rounded-[var(--r3)] text-center cursor-pointer transition-colors duration-250 hover:border-b2",
								theme === opt.value &&
									"border-[var(--blue)] shadow-[0_0_0_3px_rgba(91,156,245,0.1),_0_0_16px_rgba(91,156,245,0.06)]",
							)}
						>
							<div
								className={cn(
									"w-full h-[52px] rounded-[var(--r2)] mb-[10px]",
									opt.value === "light" && "bg-white border border-[#ddd]",
									opt.value === "dark" && "bg-[var(--void)] border border-b1",
									opt.value === "system" &&
										"bg-gradient-to-r from-white to-[var(--void)] border border-[#ddd]",
								)}
							/>
							<span className="text-[13px] font-semibold">
								{t(`theme.${opt.value}`)}
							</span>
						</button>
					))}
				</div>
			</div>

			{/* Акцентный цвет */}
			<div className="mb-9">
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em] mb-1">
					{t("accent.title")}
				</h2>
				<p className="text-[13px] text-[var(--t2)] mb-5">
					{t("accent.description")}
				</p>
				<div
					className="flex gap-[14px] justify-center mb-6"
					role="radiogroup"
					aria-label={t("accent.title")}
				>
					{ACCENT_COLORS.map((color) => (
						<button
							key={color}
							type="button"
							aria-pressed={accent === color}
							aria-label={t(`accent.${color}`)}
							onClick={() => setAccent(color)}
							className={cn(
								"w-10 h-10 rounded-full cursor-pointer transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 border-[3px] border-transparent hover:scale-115",
								accent === color &&
									"border-white/40 shadow-[0_0_0_2px_currentColor,_0_0_16px_currentColor]",
							)}
							style={{
								backgroundColor: ACCENT_PREVIEWS[color],
								color: ACCENT_PREVIEWS[color],
							}}
						/>
					))}
				</div>
			</div>

			{/* Язык */}
			<div className="mb-9">
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em] mb-1">
					{t("lang.title")}
				</h2>
				<p className="text-[13px] text-[var(--t2)] mb-5">
					{t("lang.description")}
				</p>
				<LanguageSwitcher />
			</div>

			<p className="text-[12px] text-[var(--t2)]">
				Изменения применяются автоматически
			</p>
		</div>
	);
}
