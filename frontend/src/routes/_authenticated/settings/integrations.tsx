import { createFileRoute } from "@tanstack/react-router";
import { Separator } from "@/components/ui/separator";
import { APIKeysPage } from "@/modules/settings/api-keys";
import { WebhooksPage } from "@/modules/settings/webhooks";

/**
 * Страница интеграций — API-ключи и вебхуки.
 * Компоненты уже существуют, здесь только маршрут и layout.
 */
function IntegrationsPage() {
	return (
		<div className="space-y-0">
			<div className="mb-9">
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em] mb-1">
					Интеграции
				</h2>
				<p className="text-[13px] text-t2 mb-5">
					API-ключи и вебхуки для интеграции с внешними сервисами
				</p>
			</div>

			{/* API Keys */}
			<div className="mb-9">
				<APIKeysPage />
			</div>

			<Separator className="bg-b1 my-9" />

			{/* Webhooks — projectId not available in global settings, use empty */}
			<div className="mb-9">
				<WebhooksPage projectId="" />
			</div>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/settings/integrations")({
	component: IntegrationsPage,
});
