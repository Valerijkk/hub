/**
 * Форма входа через magic link — mockup-matching design.
 *
 * Два шага: ввод email -> ввод кода подтверждения.
 * Inputs and buttons match 01-login mockup CSS values exactly.
 */
import { type FormEvent, useState } from "react";
import { toast } from "sonner";
import { sendMagicLink, verifyMagicLink } from "@/api/auth";
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
			<form onSubmit={handleSendCode}>
				<div className="mb-4">
					<label
						htmlFor="email"
						className="block text-xs font-semibold text-[var(--t2)] mb-1.5"
					>
						Email
					</label>
					<input
						id="email"
						type="email"
						placeholder="name@company.com"
						value={email}
						onChange={(e) => setEmail(e.target.value)}
						autoComplete="email"
						autoFocus
						className="w-full h-[50px] px-[18px] bg-[var(--input)] border border-[var(--b1)] rounded-[var(--r3)] text-[15px] text-[var(--t1)] placeholder:text-[var(--t3)] focus:border-[rgba(91,156,245,.35)] focus:shadow-[0_0_0_3px_rgba(91,156,245,.08),0_0_20px_rgba(91,156,245,.04)] focus:scale-[1.005] transition-all outline-none"
					/>
				</div>
				<button
					type="submit"
					className="w-full h-[50px] bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white rounded-[var(--r3)] text-[15px] font-bold shadow-[0_2px_16px_var(--gB)] hover:shadow-[0_4px_28px_rgba(91,156,245,.35)] hover:-translate-y-px active:scale-[.97] transition-all relative overflow-hidden mt-2"
					disabled={loading || !isEmailValid}
				>
					{loading ? (
						<span className="flex items-center justify-center gap-2">
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
				</button>
			</form>
		);
	}

	return (
		<form onSubmit={handleVerifyCode}>
			<p className="text-sm text-[var(--t2)] text-center mb-2">
				Код отправлен на{" "}
				<strong className="text-[var(--t1)]">{email}</strong>
			</p>
			<div className="mb-4">
				<label
					htmlFor="code"
					className="block text-xs font-semibold text-[var(--t2)] mb-1.5"
				>
					Код подтверждения
				</label>
				<input
					id="code"
					type="text"
					inputMode="numeric"
					placeholder="000000"
					maxLength={6}
					value={code}
					onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
					autoFocus
					className="w-full h-[58px] bg-[var(--input)] border border-[var(--b1)] rounded-[var(--r2)] text-center text-2xl font-bold font-mono text-[var(--t1)] focus:border-[rgba(91,156,245,.35)] focus:shadow-[0_0_0_3px_rgba(91,156,245,.08)] transition-all outline-none"
				/>
			</div>
			<button
				type="submit"
				className="w-full h-[50px] bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white rounded-[var(--r3)] text-[15px] font-bold shadow-[0_2px_16px_var(--gB)] hover:shadow-[0_4px_28px_rgba(91,156,245,.35)] hover:-translate-y-px active:scale-[.97] transition-all relative overflow-hidden mt-2"
				disabled={loading || !code.trim()}
			>
				{loading ? "Проверка..." : "Войти"}
			</button>
			<button
				type="button"
				className="block w-full text-center mt-3 text-[13px] text-[var(--t3)] hover:text-[var(--t2)] transition-colors cursor-pointer bg-transparent border-none"
				onClick={() => {
					setStep("email");
					setCode("");
				}}
			>
				&larr; Изменить email
			</button>
		</form>
	);
}
