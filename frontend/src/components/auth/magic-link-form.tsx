/**
 * Форма входа через magic link — чистый, минимальный дизайн.
 *
 * Два шага: ввод email → ввод кода подтверждения.
 * Кнопки с градиентом, тонкие фокус-кольца на инпутах.
 */
import { type FormEvent, useState } from "react";
import { toast } from "sonner";
import { sendMagicLink, verifyMagicLink } from "@/api/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuthStore } from "@/stores/auth-store";

/** Пропсы компонента MagicLinkForm. */
interface MagicLinkFormProps {
	/** Колбэк при успешной авторизации. */
	onSuccess: () => void;
}

/** Форма входа через magic link с двумя шагами: ввод email и подтверждение кодом. */
export function MagicLinkForm({ onSuccess }: MagicLinkFormProps) {
	const [email, setEmail] = useState("");
	const [code, setCode] = useState("");
	const [step, setStep] = useState<"email" | "code">("email");
	const [loading, setLoading] = useState(false);
	const { setTokens } = useAuthStore();

	const isEmailValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);

	async function handleSendCode(e: FormEvent) {
		e.preventDefault();
		if (!isEmailValid) {
			toast.error("Введите корректный email");
			return;
		}
		setLoading(true);
		try {
			const res = await sendMagicLink({ email });
			setStep("code");
			// В dev mode бэкенд возвращает код в ответе (email не настроен)
			if (res?.code) {
				setCode(res.code);
				toast.success("Код подставлен автоматически (dev mode)");
			} else {
				toast.success("Код отправлен на " + email);
			}
		} catch (err: unknown) {
			const message =
				err && typeof err === "object" && "message" in err
					? (err as { message: string }).message
					: "Не удалось отправить код";
			toast.error(message);
		} finally {
			setLoading(false);
		}
	}

	async function handleVerifyCode(e: FormEvent) {
		e.preventDefault();
		if (!code.trim()) {
			toast.error("Введите код");
			return;
		}
		setLoading(true);
		try {
			const tokens = await verifyMagicLink({ email, code: code.trim() });
			setTokens(tokens.access_token, tokens.refresh_token);
			onSuccess();
		} catch (err: unknown) {
			const message =
				err && typeof err === "object" && "message" in err
					? (err as { message: string }).message
					: "Неверный код";
			toast.error(message);
		} finally {
			setLoading(false);
		}
	}

	if (step === "email") {
		return (
			<form onSubmit={handleSendCode} className="space-y-5">
				<div className="space-y-2">
					<Label htmlFor="email" className="text-sm font-medium">
						Email
					</Label>
					<Input
						id="email"
						type="email"
						placeholder="you@example.com"
						value={email}
						onChange={(e) => setEmail(e.target.value)}
						autoComplete="email"
						autoFocus
						className="h-12 rounded-xl text-base transition-shadow duration-200 focus-visible:ring-2 focus-visible:ring-blue-500/40 focus-visible:ring-offset-0"
					/>
				</div>
				<Button
					type="submit"
					className="w-full h-12 rounded-xl bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white font-semibold text-base shadow-md shadow-blue-500/25 transition-all duration-200 hover:shadow-lg hover:shadow-blue-500/30"
					disabled={loading || !isEmailValid}
				>
					{loading ? (
						<span className="flex items-center gap-2">
							<svg
								className="h-4 w-4 animate-spin"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								strokeWidth="2"
							>
								<path d="M12 2v4m0 12v4m-7.07-3.93l2.83-2.83m8.49-8.49l2.83-2.83M2 12h4m12 0h4m-3.93 7.07l-2.83-2.83M7.76 7.76L4.93 4.93" />
							</svg>
							Отправка...
						</span>
					) : (
						"Получить код"
					)}
				</Button>
			</form>
		);
	}

	return (
		<form onSubmit={handleVerifyCode} className="space-y-5">
			<div className="space-y-2">
				<Label htmlFor="code" className="text-sm font-medium">
					Код подтверждения
				</Label>
				<Input
					id="code"
					type="text"
					inputMode="numeric"
					placeholder="000000"
					maxLength={6}
					value={code}
					onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
					autoFocus
					className="h-12 rounded-xl text-center text-lg tracking-[0.3em] font-mono transition-shadow duration-200 focus-visible:ring-2 focus-visible:ring-blue-500/40 focus-visible:ring-offset-0"
				/>
				<p className="text-xs text-muted-foreground">
					Код отправлен на {email}
				</p>
			</div>
			<Button
				type="submit"
				className="w-full h-12 rounded-xl bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white font-semibold text-base shadow-md shadow-blue-500/25 transition-all duration-200 hover:shadow-lg hover:shadow-blue-500/30"
				disabled={loading || !code.trim()}
			>
				{loading ? "Проверка..." : "Войти"}
			</Button>
			<Button
				type="button"
				variant="ghost"
				className="w-full text-sm text-muted-foreground hover:text-foreground"
				onClick={() => {
					setStep("email");
					setCode("");
				}}
			>
				Изменить email
			</Button>
		</form>
	);
}
