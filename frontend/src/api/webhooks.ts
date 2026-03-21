import { api } from "./client";

/** Вебхук-эндпоинт. */
export interface WebhookEndpoint {
  id: string;
  url: string;
  events: string[];
  is_active: boolean;
  created_at: string;
}

/** Запись доставки вебхука. */
export interface WebhookDelivery {
  id: string;
  event_type: string;
  payload: Record<string, unknown>;
  status_code: number;
  response_body?: string;
  error?: string;
  delivered_at: string;
}

/** Запрос на создание вебхук-эндпоинта. */
export interface CreateWebhookRequest {
  url: string;
  secret?: string;
  events: string[];
}

/** Запрос на обновление вебхук-эндпоинта. */
export interface UpdateWebhookRequest {
  url?: string;
  secret?: string;
  events?: string[];
  is_active?: boolean;
}

/** Создать вебхук-эндпоинт. */
export function createWebhook(projectId: string, req: CreateWebhookRequest) {
  return api.post<WebhookEndpoint>(`/projects/${projectId}/webhooks`, req);
}

/** Получить список вебхук-эндпоинтов проекта. */
export function listWebhooks(projectId: string) {
  return api.get<WebhookEndpoint[]>(`/projects/${projectId}/webhooks`);
}

/** Обновить вебхук-эндпоинт. */
export function updateWebhook(id: string, req: UpdateWebhookRequest) {
  return api.put<WebhookEndpoint>(`/webhooks/${id}`, req);
}

/** Удалить вебхук-эндпоинт. */
export function deleteWebhook(id: string) {
  return api.delete<void>(`/webhooks/${id}`);
}

/** Получить историю доставок вебхука. */
export function listDeliveries(id: string) {
  return api.get<WebhookDelivery[]>(`/webhooks/${id}/deliveries`);
}
