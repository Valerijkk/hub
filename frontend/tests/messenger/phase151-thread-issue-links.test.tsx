import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { IssueMessageLink } from "@/api/issue-message-links";
import {
	ThreadIssueLinks,
	threadIssueLinksQueryKey,
} from "@/components/messenger/thread-issue-links";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

const api = vi.hoisted(() => ({
	createIssueMessageLink: vi.fn(),
	deleteIssueMessageLink: vi.fn(),
	listThreadIssueLinks: vi.fn(),
}));

vi.mock("@/api/issue-message-links", () => api);

vi.mock("@tanstack/react-router", () => ({
	Link: ({ children }: { children: ReactNode }) => <a href="/issues">{children}</a>,
}));

const projectId = "project-1";
const channelId = "shared-channel";
const messageId = "shared-thread-root";
const userA = {
	id: "thread-links-user-a",
	email: "thread-links-user-a@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};
const userB = {
	...userA,
	id: "thread-links-user-b",
	email: "thread-links-user-b@example.test",
};

const linkA: IssueMessageLink = {
	id: "link-a",
	issue_id: "issue-a",
	issue_key: "PRIVATE-A",
	channel_id: channelId,
	message_id: messageId,
	mirror_to_thread: true,
	created_by_id: userA.id,
	created_at: "2026-08-26T00:00:00Z",
};
const linkB: IssueMessageLink = {
	...linkA,
	id: "link-b",
	issue_id: "issue-b",
	issue_key: "PRIVATE-B",
	created_by_id: userB.id,
};

function deferred<T>() {
	let resolve: (value: T) => void = () => undefined;
	const promise = new Promise<T>((done) => {
		resolve = done;
	});
	return { promise, resolve };
}

function createQueryClient() {
	return new QueryClient({
		defaultOptions: {
			queries: { retry: false },
			mutations: { retry: false },
		},
	});
}

describe("Phase 151 ThreadIssueLinks identity boundary", () => {
	let queryClient: QueryClient;

	beforeEach(() => {
		queryClient = createQueryClient();
		api.createIssueMessageLink.mockReset();
		api.deleteIssueMessageLink.mockReset();
		api.listThreadIssueLinks.mockReset();
		useAuthStore.setState({ user: userA });
		useMessengerStore.getState().reset();
	});

	afterEach(() => {
		cleanup();
		queryClient.clear();
		useAuthStore.setState({ user: null });
		useMessengerStore.getState().reset();
		vi.clearAllMocks();
	});

	function TestProvider({ children }: { children: ReactNode }) {
		return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
	}

	it("keeps delayed A issue links out of the same thread rendered for B", async () => {
		const fromA = deferred<IssueMessageLink[]>();
		const fromB = deferred<IssueMessageLink[]>();
		api.listThreadIssueLinks
			.mockImplementationOnce(() => fromA.promise)
			.mockImplementationOnce(() => fromB.promise);

		render(
			<ThreadIssueLinks
				projectId={projectId}
				channelId={channelId}
				messageId={messageId}
			/>,
			{ wrapper: TestProvider },
		);
		await waitFor(() => expect(api.listThreadIssueLinks).toHaveBeenCalledTimes(1));

		act(() => {
			useAuthStore.setState({ user: userB });
		});
		await waitFor(() => expect(api.listThreadIssueLinks).toHaveBeenCalledTimes(2));

		await act(async () => {
			fromA.resolve([linkA]);
			await fromA.promise;
		});
		expect(screen.queryByText(linkA.issue_key as string)).not.toBeInTheDocument();

		await act(async () => {
			fromB.resolve([linkB]);
			await fromB.promise;
		});
		expect(await screen.findByText(linkB.issue_key as string)).toBeInTheDocument();
		expect(screen.queryByText(linkA.issue_key as string)).not.toBeInTheDocument();
		expect(
			queryClient.getQueryData(
				threadIssueLinksQueryKey(projectId, channelId, messageId, userA.id),
			),
		).toEqual([linkA]);
		expect(
			queryClient.getQueryData(
				threadIssueLinksQueryKey(projectId, channelId, messageId, userB.id),
			),
		).toEqual([linkB]);
	});

	it("masks resident no-user issue-link data and public mutations", async () => {
		queryClient.setQueryData(
			threadIssueLinksQueryKey(projectId, channelId, messageId, "__no-user__"),
			[linkA],
		);
		useAuthStore.setState({ user: null });

		render(
			<ThreadIssueLinks
				projectId={projectId}
				channelId={channelId}
				messageId={messageId}
			/>,
			{ wrapper: TestProvider },
		);
		await act(async () => Promise.resolve());

		expect(api.listThreadIssueLinks).not.toHaveBeenCalled();
		expect(api.createIssueMessageLink).not.toHaveBeenCalled();
		expect(api.deleteIssueMessageLink).not.toHaveBeenCalled();
		expect(screen.queryByText(linkA.issue_key as string)).not.toBeInTheDocument();
		expect(screen.queryByRole("button")).not.toBeInTheDocument();
	});

	it("masks resident issue-link data and public mutations after channel revoke", async () => {
		queryClient.setQueryData(
			threadIssueLinksQueryKey(projectId, channelId, messageId, userA.id),
			[linkA],
		);
		useMessengerStore.getState().revokeChannel(channelId);

		render(
			<ThreadIssueLinks
				projectId={projectId}
				channelId={channelId}
				messageId={messageId}
			/>,
			{ wrapper: TestProvider },
		);
		await act(async () => Promise.resolve());

		expect(api.listThreadIssueLinks).not.toHaveBeenCalled();
		expect(api.createIssueMessageLink).not.toHaveBeenCalled();
		expect(api.deleteIssueMessageLink).not.toHaveBeenCalled();
		expect(screen.queryByText(linkA.issue_key as string)).not.toBeInTheDocument();
		expect(screen.queryByRole("button")).not.toBeInTheDocument();
	});
});