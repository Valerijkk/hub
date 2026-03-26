import { createFileRoute } from "@tanstack/react-router";
import { NotificationSettings } from "@/components/notifications/notification-settings";

function NotificationsSettingsPage() {
	return (
		<div>
			<h2 className="text-lg font-semibold mb-1">Уведомления</h2>
			<p className="text-sm text-muted-foreground mb-6">
				Настройте каналы доставки для каждого типа событий
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
