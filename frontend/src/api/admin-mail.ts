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

export interface BroadcastMailRequest {
	subject: string;
	html_body: string;
	dry_run?: boolean;
}

export interface BroadcastMailResponse {
	recipients: number;
	sent: number;
	dry_run: boolean;
}

export function sendMailBroadcast(data: BroadcastMailRequest) {
	return api.post<BroadcastMailResponse>("/admin/mail/broadcast", data);
}
