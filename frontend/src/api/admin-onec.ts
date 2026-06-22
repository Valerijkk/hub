import { api } from "@/api/client";

export type OnecEmployeeIDField = "email" | "hub_user_id" | "phone";

export interface OnecSettings {
	enabled: boolean;
	webhook_url: string;
	bearer_token_set: boolean;
	employee_id_field: OnecEmployeeIDField;
}

export interface UpdateOnecSettingsRequest {
	enabled?: boolean;
	webhook_url?: string;
	bearer_token?: string;
	employee_id_field?: OnecEmployeeIDField;
}

export interface TestOnecWebhookResponse {
	ok: boolean;
	status_code?: number;
	message?: string;
}

export function getOnecSettings() {
	return api.get<OnecSettings>("/admin/integrations/onec/settings");
}

export function updateOnecSettings(data: UpdateOnecSettingsRequest) {
	return api.put<OnecSettings>("/admin/integrations/onec/settings", data);
}

export function testOnecWebhook() {
	return api.post<TestOnecWebhookResponse>("/admin/integrations/onec/test", {});
}
