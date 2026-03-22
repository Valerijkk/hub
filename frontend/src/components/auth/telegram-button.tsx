import { useState } from "react";
import { toast } from "sonner";
import { api } from "@/api/client";
import type { TokenPairResponse } from "@/api/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
			const tokens = await api.post<TokenPairResponse>(
				"/auth/telegram-code",
				{ code: code.trim() },
			);
			setTokens(tokens.access_token, tokens.refresh_token);
		} catch {
			toast.error("Неверный или истёкший код");
		} finally {
			setLoading(false);
		}
	};

	if (!showCodeInput) {
		return (
			<Button
				type="button"
				variant="outline"
				className="w-full"
				onClick={handleGetCode}
			>
				<svg
					className="mr-2 h-5 w-5"
					viewBox="0 0 24 24"
					fill="currentColor"
					aria-hidden="true"
				>
					<path d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z" />
				</svg>
				Войти через Telegram
			</Button>
		);
	}

	return (
		<div className="space-y-2">
			<p className="text-xs text-muted-foreground text-center">
				Нажмите Start в боте и введите полученный код
			</p>
			<div className="flex gap-2">
				<Input
					value={code}
					onChange={(e) => setCode(e.target.value)}
					placeholder="Код из Telegram"
					maxLength={6}
					className="text-center tracking-widest font-mono"
					onKeyDown={(e) => e.key === "Enter" && handleVerify()}
				/>
				<Button onClick={handleVerify} disabled={loading || !code.trim()}>
					{loading ? "..." : "Войти"}
				</Button>
			</div>
		</div>
	);
}
