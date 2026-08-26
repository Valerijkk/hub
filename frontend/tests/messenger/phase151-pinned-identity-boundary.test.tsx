import { QueryClient } from "@tanstack/react-query";
import { act, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
	getChannelPins,
	togglePin,
	type Channel,
	type Message,
	type PinnedMessage,
} from "@/api/messenger";
import { PinnedMessages } from "@/components/messenger/pinned-messages";
import { messagesQueryKey } from "@/hooks/use-messages";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/messenger", () => ({
	getChannelPins: vi.fn(),
	togglePin: vi.fn(),
}));

const channelId = "channel-shared";
const pinCapabilities = { can_list_pins: true, can_pin: true } as const;
const userA = {
	id: "user-a",
	email: "user-a@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};
const userB = {
	...userA,
	id: "user-b",
	email: "user-b@example.test",
};

const pinA: PinnedMessage = {
	id: "pin-a",
	message_id: "message-a",
	channel_id: channelId,
	pinned_by: userA.id,
	pinned_at: "2026-08-26T00:00:00Z",
};
const pinB: PinnedMessage = {
	...pinA,
	id: "pin-b",
	message_id: "message-b",
	pinned_by: userB.id,
};
const previewA = "Private pin preview A";
const previewB = "Private pin preview B";
const messageA = {
	id: pinA.message_id,
	content_text: previewA,
} as Message;
const messageB = {
	id: pinB.message_id,
	content_text: previewB,
} as Message;

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

function seedMessagePreviews(queryClient: QueryClient) {
	queryClient.setQueryData(messagesQueryKey(channelId, userA.id), {
		pages: [[messageA]],
	});
	queryClient.setQueryData(messagesQueryKey(channelId, userB.id), {
		pages: [[messageB]],
	});
}

function openPinnedMessages() {
	fireEvent.click(
		screen.getByRole("button", { name: "Закреплённые сообщения" }),
	);
}

describe("Phase 151 pinned-message identity boundary", () => {
	const queryClients: QueryClient[] = [];

	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(getChannelPins).mockReset().mockResolvedValue([]);
		vi.mocked(togglePin).mockReset().mockResolvedValue({ pinned: false });
		useAuthStore.setState({ user: userA });
		useMessengerStore.getState().setChannels([
			{ id: channelId, capabilities: pinCapabilities } as Channel,
		]);
	});

	afterEach(() => {
		for (const queryClient of queryClients.splice(0)) {
			queryClient.clear();
		}
		useAuthStore.setState({ user: null });
		useMessengerStore.getState().reset();
	});

	it("uses an identity-scoped pins cache and never renders a delayed A response after switching to B", async () => {
		const fromA = deferred<PinnedMessage[]>();
		const fromB = deferred<PinnedMessage[]>();
		vi.mocked(getChannelPins)
			.mockImplementationOnce(() => fromA.promise)
			.mockImplementationOnce(() => fromB.promise);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		seedMessagePreviews(queryClient);

		renderWithProviders(
			<PinnedMessages channelId={channelId} capabilities={pinCapabilities} />,
			{ queryClient },
		);
		openPinnedMessages();

		await waitFor(() => expect(getChannelPins).toHaveBeenCalledTimes(1));
		act(() => {
			useAuthStore.setState({ user: userB });
		});
		await waitFor(() => expect(getChannelPins).toHaveBeenCalledTimes(2));

		await act(async () => {
			fromA.resolve([pinA]);
			await fromA.promise;
		});
		expect(screen.queryByText(previewA)).not.toBeInTheDocument();

		await act(async () => {
			fromB.resolve([pinB]);
			await fromB.promise;
		});
		expect(await screen.findByText(previewB)).toBeInTheDocument();
		expect(screen.queryByText(previewA)).not.toBeInTheDocument();
		expect(
			queryClient.getQueryData([
				"messenger",
				"pins",
				channelId,
				userA.id,
			]),
		).toEqual([pinA]);
		expect(
			queryClient.getQueryData([
				"messenger",
				"pins",
				channelId,
				userB.id,
			]),
		).toEqual([pinB]);
	});

	it("fails closed without identity when the legacy generic pins cache contains private rows", async () => {
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		seedMessagePreviews(queryClient);
		queryClient.setQueryData(["messenger", "pins", channelId], [pinA]);
		useAuthStore.setState({ user: null });

		renderWithProviders(
			<PinnedMessages channelId={channelId} capabilities={pinCapabilities} />,
			{ queryClient },
		);
		openPinnedMessages();
		await act(async () => Promise.resolve());

		expect(getChannelPins).not.toHaveBeenCalled();
		expect(screen.queryByText(previewA)).not.toBeInTheDocument();
		expect(screen.queryByRole("button", { name: "Открепить" })).not.toBeInTheDocument();
	});

	it("never renders private rows from the disabled no-user pins cache", async () => {
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		seedMessagePreviews(queryClient);
		queryClient.setQueryData(
			["messenger", "pins", channelId, "__no-user__"],
			[pinA],
		);
		useAuthStore.setState({ user: null });

		renderWithProviders(
			<PinnedMessages channelId={channelId} capabilities={pinCapabilities} />,
			{ queryClient },
		);
		openPinnedMessages();
		await act(async () => Promise.resolve());

		expect(getChannelPins).not.toHaveBeenCalled();
		expect(screen.queryByText(previewA)).not.toBeInTheDocument();
		expect(screen.queryByRole("button", { name: "Открепить" })).not.toBeInTheDocument();
	});

	it("does not refetch or replace B pins when an in-flight A unpin completes", async () => {
		const unpinFromA = deferred<{ pinned: boolean }>();
		vi.mocked(getChannelPins)
			.mockResolvedValueOnce([pinA])
			.mockResolvedValueOnce([pinB]);
		vi.mocked(togglePin).mockImplementationOnce(() => unpinFromA.promise);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		seedMessagePreviews(queryClient);

		renderWithProviders(
			<PinnedMessages channelId={channelId} capabilities={pinCapabilities} />,
			{ queryClient },
		);
		openPinnedMessages();
		expect(await screen.findByText(previewA)).toBeInTheDocument();
		fireEvent.click(screen.getByRole("button", { name: "Открепить" }));
		await waitFor(() => {
			expect(togglePin).toHaveBeenCalledWith(channelId, pinA.message_id);
		});

		act(() => {
			useAuthStore.setState({ user: userB });
		});
		expect(await screen.findByText(previewB)).toBeInTheDocument();
		expect(getChannelPins).toHaveBeenCalledTimes(2);

		await act(async () => {
			unpinFromA.resolve({ pinned: false });
			await unpinFromA.promise;
			await Promise.resolve();
		});

		expect(getChannelPins).toHaveBeenCalledTimes(2);
		expect(screen.getByText(previewB)).toBeInTheDocument();
		expect(screen.queryByText(previewA)).not.toBeInTheDocument();
		expect(
			queryClient.getQueryData([
				"messenger",
				"pins",
				channelId,
				userB.id,
			]),
		).toEqual([pinB]);
	});
});
