import { useState } from "react";
import type { TokenPairResponse } from "@/api/auth";
import { api } from "@/api/client";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";

const TG_BOT_USERNAME = "hub_notifycation_lol_bot";

/** Кнопка входа через Telegram — бот шлёт код, юзер вводит на сайте. */
export function TelegramButton() {
	const [showCodeInput, setShowCodeInput] = useState(false);
	const [code, setCode] = useState("");
	const [loading, setLoading] = useState(false);
	const setTokens = useAuthStore((s) => s.setTokens);

	const handleGetCode = () => {
		window.open(`https://t.me/${TG_BOT_USERNAME}?start=login`, "_blank");
		setShowCodeInput(true);
	};

	const handleVerify = async () => {
		if (!code.trim()) return;
		setLoading(true);
		try {
			const tokens = await api.post<TokenPairResponse>("/auth/telegram-code", {
				code: code.trim(),
			});
			setTokens(tokens.access_token, tokens.refresh_token);
		} catch (err) {
			toastApiError(err, "Неверный или истёкший код");
		} finally {
			setLoading(false);
		}
	};

	if (!showCodeInput) {
		return (
			<button
				type="button"
				className="w-full h-[50px] bg-[#187AAC] text-white rounded-[var(--r3)] text-[15px] font-bold flex items-center justify-center gap-2.5 hover:brightness-110 hover:-translate-y-px active:scale-[.97] transition-[transform,filter,opacity]"
				onClick={handleGetCode}
			>
				<svg
					className="w-[22px] h-[22px]"
					viewBox="0 0 24 24"
					fill="currentColor"
					aria-hidden="true"
				>
					<path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69a.2.2 0 00-.05-.18c-.06-.05-.14-.03-.21-.02-.09.02-1.49.95-4.22 2.79-.4.27-.76.41-1.08.4-.36-.01-1.04-.2-1.55-.37-.63-.2-1.12-.31-1.08-.66.02-.18.27-.36.74-.55 2.92-1.27 4.86-2.11 5.83-2.51 2.78-1.16 3.35-1.36 3.73-1.36.08 0 .27.02.39.12.1.08.13.19.14.27-.01.06.01.24 0 .38z" />
				</svg>
				Войти через Telegram
			</button>
		);
	}

	return (
		<div className="space-y-2">
			<p className="text-xs text-[var(--t2)] text-center">
				Нажмите Start в боте и введите полученный код
			</p>
			<div className="flex gap-2">
				<input
					value={code}
					onChange={(e) => setCode(e.target.value)}
					placeholder="Код из Telegram"
					maxLength={6}
					className="flex-1 h-[50px] px-[18px] bg-[var(--input)] border border-[var(--b1)] rounded-[var(--r3)] text-center tracking-widest font-mono text-[var(--t1)] focus:border-[rgba(91,156,245,.35)] focus:shadow-[0_0_0_3px_rgba(91,156,245,.08)] transition-colors outline-none"
					onKeyDown={(e) => e.key === "Enter" && handleVerify()}
				/>
				<button
					type="button"
					onClick={handleVerify}
					disabled={loading || !code.trim()}
					className="h-[50px] px-5 bg-[#187AAC] text-white rounded-[var(--r3)] text-[15px] font-bold hover:brightness-110 active:scale-[.97] transition-[transform,filter,opacity]"
				>
					{loading ? "..." : "Войти"}
				</button>
			</div>
		</div>
	);
}
