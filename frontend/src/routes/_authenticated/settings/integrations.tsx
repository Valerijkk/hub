import { createFileRoute, Link } from "@tanstack/react-router";
import { Separator } from "@/components/ui/separator";
import { APIKeysPage } from "@/modules/settings/api-keys";

/**
 * Страница интеграций — API-ключи.
 * Вебхуки привязаны к проекту (`/projects/:id/settings/integrations`),
 * поэтому здесь только подсказка — без пустого projectId.
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

			{/* API-ключи */}
			<div className="mb-9">
				<APIKeysPage />
			</div>

			<Separator className="bg-b1 my-9" />

			{/* Webhooks — только на уровне проекта */}
			<div className="mb-9 space-y-2" data-testid="webhooks-project-hint">
				<h3 className="text-[16px] font-bold tracking-[-0.02em]">Вебхуки</h3>
				<p className="text-[13px] text-t2">
					Вебхуки настраиваются в настройках конкретного проекта: откройте
					проект → Настройки → Интеграции.
				</p>
				<p className="text-[13px] text-t3">
					Список проектов:{" "}
					<Link
						to="/projects"
						className="text-blue font-semibold hover:underline"
					>
						Перейти к проектам
					</Link>
				</p>
			</div>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/settings/integrations")({
	component: IntegrationsPage,
});
