import { Bell, X } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import i18n from "@/i18n";
import { Button } from "@/components/ui/button";
import {
	dismissPushBanner,
	requestPushPermission,
	shouldShowPushBanner,
	subscribeToPushNotifications,
} from "@/lib/push-notifications";

/**
 * PushPermissionBanner — баннер с предложением включить push-уведомления.
 *
 * Паттерн "soft ask": показываем кастомный UI вместо прямого вызова
 * Notification.requestPermission(). Пользователь может нажать "Включить"
 * или "Позже". При "Позже" баннер не показывается повторно.
 */
export function PushPermissionBanner() {
	const [visible, setVisible] = useState(false);

	useEffect(() => {
		// Показываем баннер с небольшой задержкой чтобы не отвлекать сразу.
		const timer = setTimeout(() => {
			if (shouldShowPushBanner()) {
				setVisible(true);
			}
		}, 3000);

		return () => clearTimeout(timer);
	}, []);

	/** Пользователь нажал "Включить". */
	const handleEnable = useCallback(async () => {
		setVisible(false);

		const granted = await requestPushPermission();
		if (!granted) {
			toast.error(i18n.t("common:validation.pushPermissionDenied"));
			return;
		}

		const success = await subscribeToPushNotifications();
		if (success) {
			toast.success("Push-уведомления включены");
		} else {
			toast.error(i18n.t("common:validation.pushSubscribeFailed"));
		}
	}, []);

	/** Пользователь нажал "Позже". */
	const handleDismiss = useCallback(() => {
		setVisible(false);
		dismissPushBanner();
	}, []);

	if (!visible) return null;

	return (
		<div className="fixed bottom-4 right-4 z-50 flex max-w-sm items-start gap-3 rounded-lg border bg-card p-4 shadow-lg animate-in slide-in-from-bottom-4">
			<div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary/10">
				<Bell className="h-5 w-5 text-primary" />
			</div>
			<div className="flex-1 space-y-2">
				<p className="text-sm font-medium">Включить уведомления?</p>
				<p className="text-xs text-t2">
					Получайте мгновенные уведомления о задачах, сообщениях и звонках прямо
					в браузере.
				</p>
				<div className="flex gap-2">
					<Button size="sm" onClick={handleEnable}>
						Включить
					</Button>
					<Button size="sm" variant="ghost" onClick={handleDismiss}>
						Позже
					</Button>
				</div>
			</div>
			<button
				type="button"
				onClick={handleDismiss}
				className="shrink-0 rounded p-0.5 text-t2 hover:text-t1"
				aria-label="Закрыть"
			>
				<X className="h-4 w-4" />
			</button>
		</div>
	);
}
