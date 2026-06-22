import { createFileRoute } from "@tanstack/react-router";
import { Shield } from "lucide-react";

/**
 * Страница настроек безопасности.
 * Magic Link авторизация — пароль не используется.
 */
function SecurityPage() {
	return (
		<div className="space-y-0">
			<div className="mb-9">
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em] mb-1">
					Безопасность
				</h2>
				<p className="text-[13px] text-t2 mb-5">
					Управление безопасностью аккаунта
				</p>
			</div>

			{/* Секция смены пароля (legacy UI, вход только magic-link) */}
			<div className="mb-9">
				<h3 className="text-[14px] font-bold text-t1 mb-2">Смена пароля</h3>
				<div className="px-[18px] py-[18px] bg-gradient-to-br from-[var(--card)] to-[var(--raised)] border border-b1 rounded-[var(--r3)]">
					<div className="flex items-center gap-3 text-t2">
						<Shield className="w-5 h-5 text-t3 shrink-0" />
						<p className="text-[13px] leading-[1.5]">
							Авторизация происходит через Magic Link на вашу почту. Пароль не
							используется — это обеспечивает высокий уровень безопасности без
							необходимости запоминать и менять пароли.
						</p>
					</div>
				</div>
			</div>

			{/* Двухфакторная аутентификация (placeholder) */}
			<div className="mb-9">
				<h3 className="text-[14px] font-bold text-t1 mb-2">
					Двухфакторная аутентификация
				</h3>
				<div className="px-[18px] py-[18px] bg-gradient-to-br from-[var(--card)] to-[var(--raised)] border border-b1 rounded-[var(--r3)]">
					<div className="flex items-center gap-3 text-t2">
						<Shield className="w-5 h-5 text-t3 shrink-0" />
						<p className="text-[13px] leading-[1.5]">
							Magic Link авторизация обеспечивает безопасность без пароля.
							Каждый вход подтверждается через ссылку на вашу почту — это
							эквивалент двухфакторной аутентификации.
						</p>
					</div>
				</div>
			</div>

			{/* Активные сессии */}
			<div className="mb-9">
				<h3 className="text-[14px] font-bold text-t1 mb-2">Активные сессии</h3>
				<div className="px-[18px] py-[18px] bg-gradient-to-br from-[var(--card)] to-[var(--raised)] border border-b1 rounded-[var(--r3)]">
					<p className="text-[13px] text-t2 leading-[1.5]">
						Управление сессиями доступно в разделе{" "}
						<a
							href="/settings/profile"
							className="text-[var(--blue)] font-semibold hover:underline"
						>
							Профиль
						</a>
						.
					</p>
				</div>
			</div>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/settings/security")({
	component: SecurityPage,
});
