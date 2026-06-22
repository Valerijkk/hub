/**
 * Провайдер темы приложения: светлая/тёмная/системная + акцентный цвет.
 *
 * Использует next-themes для переключения class на html (Tailwind dark mode).
 * Акцентный цвет хранится в localStorage и устанавливается через data-accent на html.
 * Предустановленные акцентные цвета: red (default), blue, green, purple, orange.
 *
 * @example Пример подключения
 * // В main.tsx
 * <ThemeProvider>
 *   <App />
 * </ThemeProvider>
 *
 * // В компоненте
 * const { accent, setAccent } = useAccentColor();
 */
import {
	ThemeProvider as NextThemesProvider,
	useTheme as useNextTheme,
} from "next-themes";
import type { ReactNode } from "react";
import {
	createContext,
	useCallback,
	useContext,
	useEffect,
	useState,
} from "react";

/** Допустимые акцентные цвета */
export type AccentColor = "red" | "blue" | "green" | "purple" | "orange";

/** Все доступные акцентные цвета */
export const ACCENT_COLORS: AccentColor[] = [
	"red",
	"blue",
	"green",
	"purple",
	"orange",
];

const ACCENT_STORAGE_KEY = "hub-accent-color";
const DEFAULT_ACCENT: AccentColor = "red";

interface AccentContextValue {
	/** Текущий акцентный цвет */
	accent: AccentColor;
	/** Установить акцентный цвет */
	setAccent: (color: AccentColor) => void;
}

const AccentContext = createContext<AccentContextValue>({
	accent: DEFAULT_ACCENT,
	setAccent: () => {},
});

/**
 * Хук для доступа к акцентному цвету
 */
export function useAccentColor() {
	return useContext(AccentContext);
}

/**
 * Реэкспорт useTheme из next-themes для удобства
 */
export { useNextTheme as useTheme };

/**
 * Провайдер акцентного цвета -- устанавливает data-accent на html
 */
function AccentProvider({ children }: { children: ReactNode }) {
	const [accent, setAccentState] = useState<AccentColor>(() => {
		if (typeof window === "undefined") return DEFAULT_ACCENT;
		const stored = localStorage.getItem(ACCENT_STORAGE_KEY);
		if (stored && ACCENT_COLORS.includes(stored as AccentColor)) {
			return stored as AccentColor;
		}
		return DEFAULT_ACCENT;
	});

	const setAccent = useCallback((color: AccentColor) => {
		setAccentState(color);
		localStorage.setItem(ACCENT_STORAGE_KEY, color);
		document.documentElement.setAttribute("data-accent", color);
	}, []);

	useEffect(() => {
		document.documentElement.setAttribute("data-accent", accent);
	}, [accent]);

	return (
		<AccentContext.Provider value={{ accent, setAccent }}>
			{children}
		</AccentContext.Provider>
	);
}

/**
 * Комбинированный провайдер темы (next-themes) и акцентного цвета.
 * Оборачивает приложение, обеспечивая переключение dark/light/system + accent.
 */
export function ThemeProvider({ children }: { children: ReactNode }) {
	return (
		<NextThemesProvider attribute="class" defaultTheme="system" enableSystem>
			<AccentProvider>{children}</AccentProvider>
		</NextThemesProvider>
	);
}
