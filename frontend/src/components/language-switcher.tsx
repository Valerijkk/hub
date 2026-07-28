/**
 * Переключатель языка интерфейса (RU / EN).
 *
 * Компактный вариант для topbar и полный для настроек.
 * При переключении вызывает i18n.changeLanguage и сохраняет в localStorage.
 */
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface LanguageSwitcherProps {
	/** Компактный вид (только аббревиатуры) для topbar */
	compact?: boolean;
	className?: string;
}

/**
 * Переключатель языка RU/EN
 */
export function LanguageSwitcher({
	compact = false,
	className,
}: LanguageSwitcherProps) {
	const { i18n } = useTranslation();
	const currentLang = i18n.language?.startsWith("en") ? "en" : "ru";

	function handleChange(lang: string) {
		void i18n.changeLanguage(lang);
	}

	if (compact) {
		return (
			<div
				role="group"
				aria-label="Язык"
				className={cn(
					"flex items-center gap-0.5 rounded-md border p-0.5",
					className,
				)}
			>
				<button
					type="button"
					aria-pressed={currentLang === "ru"}
					aria-label="Русский"
					onClick={() => handleChange("ru")}
					className={cn(
						"rounded px-1.5 py-0.5 text-xs font-medium transition-colors",
						currentLang === "ru"
							? "bg-primary text-primary-foreground"
							: "text-t2 hover:text-t1",
					)}
				>
					RU
				</button>
				<button
					type="button"
					aria-pressed={currentLang === "en"}
					aria-label="English"
					onClick={() => handleChange("en")}
					className={cn(
						"rounded px-1.5 py-0.5 text-xs font-medium transition-colors",
						currentLang === "en"
							? "bg-primary text-primary-foreground"
							: "text-t2 hover:text-t1",
					)}
				>
					EN
				</button>
			</div>
		);
	}

	return (
		<div className={cn("flex gap-2", className)}>
			<Button
				variant={currentLang === "ru" ? "default" : "outline"}
				size="sm"
				onClick={() => handleChange("ru")}
			>
				Русский
			</Button>
			<Button
				variant={currentLang === "en" ? "default" : "outline"}
				size="sm"
				onClick={() => handleChange("en")}
			>
				English
			</Button>
		</div>
	);
}
