import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ThreadPanel } from "@/components/messenger/thread-panel";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

const issueLinks = vi.hoisted(() =>
	vi.fn(() => <div data-testid="thread-issue-links" />),
);

vi.mock("@/components/messenger/thread-issue-links", () => ({
	ThreadIssueLinks: issueLinks,
}));

vi.mock("@/hooks/use-messages", () => ({
	threadDetailQueryKey: (...parts: string[]) => ["messenger", "thread", ...parts],
	useThreadMessages: () => ({ messages: [], isLoading: false }),
}));

const user = {
	id: "thread-panel-user",
	email: "thread-panel-user@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};

describe("Phase 151 ThreadPanel issue-link mount guard", () => {
	let queryClient: QueryClient;

	beforeEach(() => {
		queryClient = new QueryClient({
			defaultOptions: { queries: { retry: false } },
		});
		issueLinks.mockClear();
		useAuthStore.setState({ user });
		useMessengerStore.getState().reset();
	});

	afterEach(() => {
		cleanup();
		queryClient.clear();
		useAuthStore.setState({ user: null });
		useMessengerStore.getState().reset();
	});

	function TestProvider({ children }: { children: ReactNode }) {
		return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
	}

	it("does not mount ThreadIssueLinks without an authenticated user", () => {
		useAuthStore.setState({ user: null });

		render(
			<ThreadPanel
				projectId="project-1"
				channelId="channel-1"
				messageId="message-1"
				onClose={vi.fn()}
			/>,
			{ wrapper: TestProvider },
		);

		expect(issueLinks).not.toHaveBeenCalled();
		expect(screen.queryByTestId("thread-issue-links")).not.toBeInTheDocument();
	});

	it("does not mount ThreadIssueLinks after the channel is revoked", () => {
		useMessengerStore.getState().revokeChannel("channel-1");

		render(
			<ThreadPanel
				projectId="project-1"
				channelId="channel-1"
				messageId="message-1"
				onClose={vi.fn()}
			/>,
			{ wrapper: TestProvider },
		);

		expect(issueLinks).not.toHaveBeenCalled();
		expect(screen.queryByTestId("thread-issue-links")).not.toBeInTheDocument();
	});
});