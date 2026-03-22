import { useCallback, useEffect, useRef } from "react";
import { toast } from "sonner";
import { loginTelegram } from "@/api/auth";
import { useAuthStore } from "@/stores/auth-store";

const TG_BOT_USERNAME = "hub_notifycation_lol_bot";

/** Кнопка входа через Telegram Login Widget. */
export function TelegramButton() {
	const containerRef = useRef<HTMLDivElement>(null);
	const setTokens = useAuthStore((s) => s.setTokens);

	const handleAuth = useCallback(
		async (user: Record<string, string | number>) => {
			// Telegram Login Widget возвращает объект с полями:
			// id, first_name, last_name, username, photo_url, auth_date, hash
			const params = new URLSearchParams();
			for (const [key, value] of Object.entries(user)) {
				if (value != null) params.set(key, String(value));
			}
			try {
				const tokens = await loginTelegram({ init_data: params.toString() });
				setTokens(tokens.access_token, tokens.refresh_token);
			} catch {
				toast.error("Не удалось войти через Telegram");
			}
		},
		[setTokens],
	);

	useEffect(() => {
		// Регистрируем глобальный callback для виджета.
		(window as unknown as Record<string, unknown>).__tg_auth_callback = handleAuth;

		const script = document.createElement("script");
		script.src = "https://telegram.org/js/telegram-widget.js?22";
		script.async = true;
		script.setAttribute("data-telegram-login", TG_BOT_USERNAME);
		script.setAttribute("data-size", "large");
		script.setAttribute("data-radius", "8");
		script.setAttribute("data-request-access", "write");
		script.setAttribute("data-onauth", "__tg_auth_callback(user)");
		script.setAttribute("data-userpic", "false");

		const container = containerRef.current;
		if (container) {
			container.innerHTML = "";
			container.appendChild(script);
		}

		return () => {
			delete (window as unknown as Record<string, unknown>).__tg_auth_callback;
		};
	}, [handleAuth]);

	return <div ref={containerRef} className="flex w-full justify-center" />;
}
