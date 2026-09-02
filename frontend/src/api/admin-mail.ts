import { api } from "@/api/client";

export interface MailSettings {
	hub_mail_url: string;
	from_alerts: string;
	alerts_enabled: boolean;
	daily_limit: number;
	resend_ready: boolean;
}

export interface UpdateMailSettingsRequest {
	hub_mail_url?: string;
	from_alerts?: string;
	alerts_enabled?: boolean;
	daily_limit?: number;
}

export function getMailSettings() {
	return api.get<MailSettings>("/admin/mail/settings");
}

export function updateMailSettings(data: UpdateMailSettingsRequest) {
	return api.put<MailSettings>("/admin/mail/settings", data);
}
