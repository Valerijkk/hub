import { api } from "./client";

export interface StatusComponent {
	id: string;
	name: string;
	status: "operational" | "degraded" | "outage" | "unknown";
	description?: string;
}

export interface StatusResponse {
	status: "operational" | "degraded";
	updated_at: string;
	components: StatusComponent[];
}

export function fetchPublicStatus(): Promise<StatusResponse> {
	return api.get<StatusResponse>("/status");
}
