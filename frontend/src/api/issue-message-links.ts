import { api } from "./client";

/** Связь задачи с корневым сообщением треда (PROD-03). */
export interface IssueMessageLink {
	id: string;
	issue_id: string;
	issue_key?: string;
	issue_title?: string;
	channel_id: string;
	message_id: string;
	mirror_to_thread: boolean;
	created_by_id: string;
	created_at: string;
}

export interface CreateIssueMessageLinkRequest {
	channel_id: string;
	message_id: string;
	mirror_to_thread?: boolean;
}

export function listIssueMessageLinks(projectId: string, issueKey: string) {
	return api.get<IssueMessageLink[]>(
		`/projects/${projectId}/issues/${issueKey}/message-links`,
	);
}

export function createIssueMessageLink(
	projectId: string,
	issueKey: string,
	data: CreateIssueMessageLinkRequest,
) {
	return api.post<IssueMessageLink>(
		`/projects/${projectId}/issues/${issueKey}/message-links`,
		data,
	);
}

export function deleteIssueMessageLink(
	projectId: string,
	issueKey: string,
	linkId: string,
) {
	return api.delete<void>(
		`/projects/${projectId}/issues/${issueKey}/message-links/${linkId}`,
	);
}

export function listThreadIssueLinks(
	projectId: string,
	channelId: string,
	messageId: string,
) {
	return api.get<IssueMessageLink[]>(
		`/projects/${projectId}/message-links?channel_id=${encodeURIComponent(channelId)}&message_id=${encodeURIComponent(messageId)}`,
	);
}
