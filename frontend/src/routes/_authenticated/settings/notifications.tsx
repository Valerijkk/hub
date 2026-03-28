import { createFileRoute } from "@tanstack/react-router";
import { NotificationSettings } from "@/components/notifications/notification-settings";

function NotificationsSettingsPage() {
	return (
		<div className="mb-9">
			<h2 className="text-[18px] font-extrabold tracking-[-0.02em] mb-1">Уведомления</h2>
			<p className="text-[13px] text-t2 mb-5">
				Настройте каналы доставки для каждого типа события
			</p>
			<NotificationSettings />
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/settings/notifications",
)({
	component: NotificationsSettingsPage,
});
