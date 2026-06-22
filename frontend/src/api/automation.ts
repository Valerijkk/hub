import { api } from "./client";

export interface AutomationRule {
	id: string;
	project_id: string;
	name: string;
	enabled: boolean;
	trigger_type: "status_changed" | "issue_created" | "issue_assigned";
	trigger_config: Record<string, unknown>;
	action_type: "post_to_channel" | "notify_user";
	action_config: Record<string, unknown>;
	created_by_id: string;
	created_at: string;
	updated_at: string;
}

export interface CreateAutomationRuleRequest {
	name: string;
	enabled?: boolean;
	trigger_type: AutomationRule["trigger_type"];
	trigger_config: Record<string, unknown>;
	action_type: AutomationRule["action_type"];
	action_config: Record<string, unknown>;
}

export function listAutomationRules(projectId: string) {
	return api.get<AutomationRule[]>(`/projects/${projectId}/automation-rules`);
}

export function createAutomationRule(
	projectId: string,
	data: CreateAutomationRuleRequest,
) {
	return api.post<AutomationRule>(
		`/projects/${projectId}/automation-rules`,
		data,
	);
}

export function updateAutomationRule(
	projectId: string,
	ruleId: string,
	data: Partial<CreateAutomationRuleRequest> & { enabled?: boolean },
) {
	return api.patch<AutomationRule>(
		`/projects/${projectId}/automation-rules/${ruleId}`,
		data,
	);
}

export function deleteAutomationRule(projectId: string, ruleId: string) {
	return api.delete<void>(`/projects/${projectId}/automation-rules/${ruleId}`);
}
