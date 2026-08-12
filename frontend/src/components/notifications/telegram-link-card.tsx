import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import {
	confirmTelegramLink,
	createTelegramBind,
	getTelegramStatus,
	sendTelegramTest,
	type TelegramBindResponse,
	type TelegramStatusResponse,
	unlinkTelegram,
} from "@/api/notifications";
import { QueryErrorAlert } from "@/components/query-state-view";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toastApiError } from "@/lib/format-api-error";

/**
 * TelegramLinkCard — привязка Telegram для уведомлений (NOTIF-02).
 */
export function TelegramLinkCard() {
	const [status, setStatus] = useState<TelegramStatusResponse | null>(null);
	const [bindInfo, setBindInfo] = useState<TelegramBindResponse | null>(null);
	const [code, setCode] = useState("");
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);
	const [busy, setBusy] = useState(false);

	const refresh = useCallback(async () => {
		setLoading(true);
		setLoadError(false);
		try {
			setStatus(await getTelegramStatus());
		} catch (err) {
			// W9-A4-006: fail ≠ «не привязан»
			setStatus(null);
			setLoadError(true);
			toastApiError(err, "Не удалось загрузить статус Telegram");
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		void refresh();
	}, [refresh]);

	const handleBind = async () => {
		setBusy(true);
		try {
			const resp = await createTelegramBind();
			setBindInfo(resp);
			if (resp.deep_link) {
				void navigator.clipboard.writeText(resp.deep_link);
				toast.success("Ссылка для Telegram скопирована");
			} else {
				toast.success("Запросите код у бота командой /link");
			}
		} catch (err) {
			toastApiError(err, "Telegram bind недоступен");
		} finally {
			setBusy(false);
		}
	};

	const handleConfirmCode = async () => {
		if (code.length !== 6) {
			toast.error("Код должен быть 6 цифр");
			return;
		}
		setBusy(true);
		try {
			await confirmTelegramLink(code);
			setCode("");
			setBindInfo(null);
			await refresh();
			toast.success("Telegram привязан");
		} catch (err) {
			toastApiError(err, "Неверный или просроченный код");
		} finally {
			setBusy(false);
		}
	};

	const handleUnlink = async () => {
		setBusy(true);
		try {
			await unlinkTelegram();
			setBindInfo(null);
			await refresh();
			toast.success("Telegram отвязан");
		} catch (err) {
			toastApiError(err, "Не удалось отвязать");
		} finally {
			setBusy(false);
		}
	};

	const handleTest = async () => {
		setBusy(true);
		try {
			await sendTelegramTest();
			toast.success("Тест отправлен в Telegram");
		} catch (err) {
			toastApiError(err, "Не удалось отправить тест");
		} finally {
			setBusy(false);
		}
	};

	return (
		<div className="px-[18px] py-[18px] bg-gradient-to-br from-[var(--card)] to-[var(--raised)] border border-b1 rounded-[var(--r3)] mb-5">
			<h4 className="text-[14px] font-bold mb-1">Telegram</h4>
			<p className="text-[12px] text-t3 mb-3">
				Привязка для push-уведомлений в мессенджер (не для входа)
			</p>

			{loading ? (
				<p className="text-[13px] text-t3">Загрузка...</p>
			) : loadError ? (
				<QueryErrorAlert
					compact
					message="Не удалось загрузить статус Telegram"
					retry={() => void refresh()}
				/>
			) : status?.linked ? (
				<div className="space-y-3">
					<p className="text-[13px] text-green">
						Привязан · ID {status.telegram_id}
					</p>
					<div className="flex flex-wrap gap-2">
						<Button
							size="sm"
							variant="outline"
							disabled={busy}
							onClick={() => void handleTest()}
						>
							Тест
						</Button>
						<Button
							size="sm"
							variant="destructive"
							disabled={busy}
							onClick={() => void handleUnlink()}
						>
							Отвязать
						</Button>
					</div>
				</div>
			) : (
				<div className="space-y-3">
					<Button size="sm" disabled={busy} onClick={() => void handleBind()}>
						Привязать Telegram
					</Button>
					{bindInfo?.deep_link && (
						<p className="text-[11px] text-t3 break-all">
							{bindInfo.deep_link}
						</p>
					)}
					<div className="flex gap-2 items-end">
						<div className="flex-1">
							<label
								htmlFor="tg-link-code"
								className="block text-[11px] text-t3 mb-1"
							>
								Или код из бота (/link)
							</label>
							<Input
								id="tg-link-code"
								inputMode="numeric"
								maxLength={6}
								value={code}
								onChange={(e) =>
									setCode(e.target.value.replace(/\D/g, "").slice(0, 6))
								}
								placeholder="123456"
								className="font-mono"
							/>
						</div>
						<Button
							size="sm"
							disabled={busy || code.length !== 6}
							onClick={() => void handleConfirmCode()}
						>
							Подтвердить
						</Button>
					</div>
				</div>
			)}
		</div>
	);
}
