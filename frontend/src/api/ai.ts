import { api } from "@/api/client";

export interface AIModelInfo {
	id: string;
	provider_id: string;
	provider_name: string;
	ru_resident: boolean;
}

export interface AIModelsResponse {
	models: AIModelInfo[];
}

export function getAIModels() {
	return api.get<AIModelsResponse>("/ai/models");
}

export interface AIBotInfo {
	user_id: string;
	display_name: string;
	email: string;
}

export function getAIBotInfo() {
	return api.get<AIBotInfo>("/ai/bot");
}

export interface ChatMessage {
	role: string;
	content: string;
}

export interface ChatCompletionResponse {
	id: string;
	model: string;
	choices: Array<{
		message: ChatMessage;
		finish_reason: string;
	}>;
	usage: {
		prompt_tokens: number;
		completion_tokens: number;
		total_tokens: number;
	};
}

export interface ChatCompletionRequest {
	model: string;
	messages: ChatMessage[];
	stream?: boolean;
	stream_id?: string;
}

export interface StreamStartResponse {
	stream_id: string;
	status: string;
}

export function chatCompletion(data: ChatCompletionRequest) {
	return api.post<ChatCompletionResponse | StreamStartResponse>(
		"/ai/v1/chat/completions",
		data,
	);
}

export interface JQLAssistRequest {
	query: string;
	project_id?: string;
}

export interface JQLAssistResponse {
	jql: string;
	explanation?: string;
}

export function assistJQL(data: JQLAssistRequest) {
	return api.post<JQLAssistResponse>("/ai/jql/assist", data);
}

export interface AIUsageToday {
	usage_date: string;
	tokens_in: number;
	tokens_out: number;
	request_count: number;
	daily_limit: number;
}

export function getAIUsageToday() {
	return api.get<AIUsageToday>("/ai/usage/today");
}

export interface AIUserPreference {
	default_model?: string;
	allow_external_ai?: boolean;
}

export function getAIPreferences() {
	return api.get<AIUserPreference>("/ai/preferences");
}

export function updateAIPreferences(data: Partial<AIUserPreference>) {
	return api.patch<AIUserPreference>("/ai/preferences", data);
}

export interface ComposerSuggestRequest {
	prefix: string;
	model?: string;
	channel_id?: string;
}

export interface ComposerSuggestResponse {
	suggestion: string;
	model: string;
}

/** Tab-autocomplete в композере мессенджера (M3-024). */
export function suggestComposerCompletion(data: ComposerSuggestRequest) {
	return api.post<ComposerSuggestResponse>("/ai/composer/suggest", data);
}
