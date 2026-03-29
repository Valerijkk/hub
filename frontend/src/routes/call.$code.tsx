/**
 * Публичная страница присоединения к звонку по invite-коду.
 * URL: /call/{code}
 *
 * Авторизованный пользователь автоматически присоединяется к звонку.
 * Неавторизованный — видит форму для ввода имени (гостевой доступ).
 */
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useRef, useState } from "react";
import { joinByInviteCode, joinCall, guestJoinCall } from "@/api/calls";
import { useCallStore } from "@/stores/call-store";
import { useAuthStore } from "@/stores/auth-store";
import { LiveKitRoom, RoomAudioRenderer } from "@livekit/components-react";
import "@livekit/components-styles";

/** URL сервера LiveKit. */
const LIVEKIT_URL =
	(import.meta.env.VITE_LIVEKIT_URL as string) ||
	(window.location.protocol === "https:"
		? `wss://${window.location.host}/lk`
		: `ws://${window.location.hostname}:7880`);

function JoinCallByCodePage() {
	const { code } = Route.useParams();
	const navigate = useNavigate();
	const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
	const [error, setError] = useState<string | null>(null);
	const [loading, setLoading] = useState(true);

	// Guest state
	const [guestMode, setGuestMode] = useState(false);
	const [guestName, setGuestName] = useState("");
	const [guestSubmitting, setGuestSubmitting] = useState(false);
	const [guestToken, setGuestToken] = useState<string | null>(null);
	const [guestLivekitUrl, setGuestLivekitUrl] = useState<string | null>(null);
	const nameInputRef = useRef<HTMLInputElement>(null);

	// Authenticated flow
	useEffect(() => {
		if (!isAuthenticated) {
			setGuestMode(true);
			setLoading(false);
			return;
		}

		let cancelled = false;

		async function join() {
			try {
				const call = await joinByInviteCode(code);
				if (cancelled) return;

				const joined = await joinCall(call.id);
				if (cancelled) return;

				useCallStore.setState({
					activeCall: joined,
					token: joined.token ?? null,
					isInCall: true,
					isConnecting: false,
				});

				void navigate({ to: "/messenger" });
			} catch {
				if (!cancelled) {
					setError("Не удалось присоединиться к звонку. Возможно, звонок уже завершён.");
					setLoading(false);
				}
			}
		}

		void join();
		return () => { cancelled = true; };
	}, [code, isAuthenticated, navigate]);

	// Focus name input on guest mode
	useEffect(() => {
		if (guestMode && nameInputRef.current) {
			nameInputRef.current.focus();
		}
	}, [guestMode]);

	const handleGuestJoin = useCallback(async () => {
		const trimmed = guestName.trim();
		if (!trimmed) return;

		setGuestSubmitting(true);
		setError(null);

		try {
			const res = await guestJoinCall(code, trimmed);
			setGuestToken(res.token);
			setGuestLivekitUrl(res.livekit_url || LIVEKIT_URL);
		} catch {
			setError("Не удалось войти в звонок. Проверьте ссылку или попробуйте позже.");
		} finally {
			setGuestSubmitting(false);
		}
	}, [code, guestName]);

	const handleKeyDown = useCallback(
		(e: React.KeyboardEvent) => {
			if (e.key === "Enter") {
				void handleGuestJoin();
			}
		},
		[handleGuestJoin],
	);

	// Error state
	if (error && !guestMode) {
		return (
			<div className="min-h-screen flex items-center justify-center bg-void">
				<div className="max-w-md text-center p-8">
					<div className="w-16 h-16 rounded-full bg-red/10 flex items-center justify-center mx-auto mb-4">
						<svg className="w-8 h-8 text-red" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
							<path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
						</svg>
					</div>
					<h1 className="text-xl font-bold text-t1 mb-2">Звонок недоступен</h1>
					<p className="text-t2 mb-6">{error}</p>
					<button
						type="button"
						onClick={() => void navigate({ to: "/messenger" })}
						className="px-6 py-2.5 rounded-[var(--r2)] bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white font-bold text-sm"
					>
						Перейти в мессенджер
					</button>
				</div>
			</div>
		);
	}

	// Guest is connected to LiveKit
	if (guestToken && guestLivekitUrl) {
		return (
			<div className="min-h-screen bg-void flex flex-col">
				<LiveKitRoom
					serverUrl={guestLivekitUrl}
					token={guestToken}
					connect={true}
					onDisconnected={() => {
						setGuestToken(null);
						setGuestLivekitUrl(null);
						setGuestMode(true);
						setError("Звонок завершён.");
					}}
					style={{ height: "100vh" }}
				>
					<RoomAudioRenderer />
					<div className="flex-1 flex items-center justify-center">
						<div className="text-center p-8">
							<div className="w-20 h-20 rounded-full bg-[var(--blueS)]/20 flex items-center justify-center mx-auto mb-6">
								<svg className="w-10 h-10 text-[var(--blueS)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
									<path strokeLinecap="round" strokeLinejoin="round" d="M12 18.75a6 6 0 006-6v-1.5m-6 7.5a6 6 0 01-6-6v-1.5m6 7.5v3.75m-3.75 0h7.5M12 15.75a3 3 0 01-3-3V4.5a3 3 0 116 0v8.25a3 3 0 01-3 3z" />
								</svg>
							</div>
							<h2 className="text-xl font-bold text-t1 mb-2">Вы в звонке</h2>
							<p className="text-t2 text-sm mb-6">Подключено как {guestName.trim()}</p>
							<button
								type="button"
								onClick={() => {
									setGuestToken(null);
									setGuestLivekitUrl(null);
								}}
								className="px-6 py-2.5 rounded-[var(--r2)] bg-red/90 hover:bg-red text-white font-bold text-sm transition-colors"
							>
								Покинуть звонок
							</button>
						</div>
					</div>
				</LiveKitRoom>
			</div>
		);
	}

	// Guest form
	if (guestMode) {
		return (
			<div className="min-h-screen flex items-center justify-center bg-void">
				<div className="max-w-sm w-full mx-4">
					<div className="bg-[var(--b2)] rounded-[var(--r3)] border border-[var(--border)] p-8 shadow-xl">
						{/* Icon */}
						<div className="w-16 h-16 rounded-full bg-[var(--blueS)]/10 flex items-center justify-center mx-auto mb-6">
							<svg className="w-8 h-8 text-[var(--blueS)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
								<path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
							</svg>
						</div>

						<h1 className="text-xl font-bold text-t1 text-center mb-2">
							Присоединиться к звонку
						</h1>
						<p className="text-t2 text-sm text-center mb-6">
							Введите ваше имя, чтобы войти как гость
						</p>

						{error && (
							<div className="bg-red/10 border border-red/20 rounded-[var(--r2)] px-4 py-3 mb-4">
								<p className="text-red text-sm">{error}</p>
							</div>
						)}

						<div className="space-y-4">
							<input
								ref={nameInputRef}
								type="text"
								value={guestName}
								onChange={(e) => setGuestName(e.target.value)}
								onKeyDown={handleKeyDown}
								placeholder="Ваше имя"
								maxLength={50}
								disabled={guestSubmitting}
								className="w-full px-4 py-3 rounded-[var(--r2)] bg-[var(--b1)] border border-[var(--border)] text-t1 placeholder:text-t3 text-sm outline-none focus:border-[var(--blueS)] focus:ring-1 focus:ring-[var(--blueS)] transition-colors disabled:opacity-50"
							/>
							<button
								type="button"
								onClick={() => void handleGuestJoin()}
								disabled={!guestName.trim() || guestSubmitting}
								className="w-full px-6 py-3 rounded-[var(--r2)] bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white font-bold text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:brightness-110 transition-all"
							>
								{guestSubmitting ? (
									<span className="flex items-center justify-center gap-2">
										<span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
										Подключение...
									</span>
								) : (
									"Войти в звонок"
								)}
							</button>
						</div>

						<div className="mt-6 pt-4 border-t border-[var(--border)]">
							<p className="text-t3 text-xs text-center">
								Есть аккаунт?{" "}
								<button
									type="button"
									onClick={() => void navigate({ to: "/login" })}
									className="text-[var(--blueS)] hover:underline font-medium"
								>
									Войти
								</button>
							</p>
						</div>
					</div>
				</div>
			</div>
		);
	}

	// Loading state (authenticated flow)
	if (loading) {
		return (
			<div className="min-h-screen flex items-center justify-center bg-void">
				<div className="text-center">
					<div className="w-10 h-10 border-2 border-blue border-t-transparent rounded-full animate-spin mx-auto mb-4" />
					<p className="text-t2 text-sm">Подключение к звонку...</p>
				</div>
			</div>
		);
	}

	return null;
}

export const Route = createFileRoute("/call/$code")({
	component: JoinCallByCodePage,
});
