/**
 * Публичная страница присоединения к звонку по invite-коду.
 * URL: /call/{code}
 *
 * Авторизованный пользователь автоматически присоединяется к звонку.
 * Неавторизованный — перенаправляется на /login.
 */
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { joinByInviteCode, joinCall } from "@/api/calls";
import { useCallStore } from "@/stores/call-store";
import { useAuthStore } from "@/stores/auth-store";

function JoinCallByCodePage() {
	const { code } = Route.useParams();
	const navigate = useNavigate();
	const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
	const [error, setError] = useState<string | null>(null);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		if (!isAuthenticated) {
			void navigate({ to: "/login" });
			return;
		}

		let cancelled = false;

		async function join() {
			try {
				const call = await joinByInviteCode(code);
				if (cancelled) return;

				const joined = await joinCall(call.id);
				if (cancelled) return;

				// Обновляем store напрямую
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

	if (!isAuthenticated) return null;

	if (error) {
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
