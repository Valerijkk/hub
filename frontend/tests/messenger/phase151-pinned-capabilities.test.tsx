import { QueryClient } from "@tanstack/react-query";
import { act, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
	getChannelPins,
	togglePin,
	type Channel,
	type ChannelCapabilities,
	type Message,
	type PinnedMessage,
} from "@/api/messenger";
import { PinnedMessages } from "@/components/messenger/pinned-messages";
import { messagesQueryKey } from "@/hooks/use-messages";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getChannelPins: vi.fn(),
		togglePin: vi.fn(),
	};
});

const channelId = "pins-capability-channel";
const user = {
	id: "pins-capability-user",
	email: "pins-capability@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};
const pin: PinnedMessage = {
	id: "pin-1",
	message_id: "message-1",
	channel_id: channelId,
	pinned_by: user.id,
	pinned_at: "2026-08-26T00:00:00Z",
};
const preview = "Capability-bound pinned preview";

function capabilities(
	overrides: Partial<ChannelCapabilities> = {},
): ChannelCapabilities {
	return {
		is_member: true,
		has_active_project_membership: false,
		can_read: true,
		can_post: false,
		can_reply: false,
		can_join: false,
		can_follow: false,
		can_unfollow: false,
		can_view_unread: false,
		can_type: false,
		can_edit: false,
		can_forward: false,
		can_react: false,
		can_delete_message: false,
		can_pin: false,
		can_list_pins: false,
		can_bookmark: false,
		can_promote: false,
		...overrides,
	};
}

function channel(currentCapabilities: ChannelCapabilities): Channel {
	return {
		id: channelId,
		name: "Pins capabilities",
		channel_type: "private",
		is_readonly: false,
		capabilities: currentCapabilities,
		created_by: user.id,
		is_starred: false,
		created_at: "2026-08-26T00:00:00Z",
		updated_at: "2026-08-26T00:00:00Z",
	};
}

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

function seedPreview(queryClient: QueryClient) {
	queryClient.setQueryData<{ pages: Message[][] }>(
		messagesQueryKey(channelId, user.id),
		{
			pages: [[{ id: pin.message_id, content_text: preview } as Message]],
		},
	);
}

function openPinnedMessages() {
	fireEvent.click(
		screen.getByRole("button", { name: "Закреплённые сообщения" }),
	);
}

describe("Phase 151 pinned-message capability boundary", () => {
	const queryClients: QueryClient[] = [];

	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(getChannelPins).mockReset().mockResolvedValue([pin]);
		vi.mocked(togglePin).mockReset().mockResolvedValue({ pinned: false });
		useAuthStore.setState({ user });
		useMessengerStore.getState().reset();
	});

	afterEach(() => {
		for (const queryClient of queryClients.splice(0)) queryClient.clear();
		useAuthStore.setState({ user: null });
		useMessengerStore.getState().reset();
	});

	it("fails closed without explicit can_list_pins and never starts a pins request", async () => {
		const queryClient = createQueryClient();
		queryClients.push(queryClient);

		renderWithProviders(<PinnedMessages channelId={channelId} />, {
			queryClient,
		});
		await act(async () => Promise.resolve());

		expect(
			screen.queryByRole("button", { name: "Закреплённые сообщения" }),
		).not.toBeInTheDocument();
		expect(getChannelPins).not.toHaveBeenCalled();
	});

	it("fetches lazily with can_list_pins but hides unpin without can_pin", async () => {
		const currentCapabilities = capabilities({ can_list_pins: true });
		useMessengerStore.getState().setChannels([channel(currentCapabilities)]);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		seedPreview(queryClient);

		renderWithProviders(
			<PinnedMessages
				channelId={channelId}
				capabilities={currentCapabilities}
			/>,
			{ queryClient },
		);
		expect(getChannelPins).not.toHaveBeenCalled();
		openPinnedMessages();

		expect(await screen.findByText(preview)).toBeInTheDocument();
		expect(getChannelPins).toHaveBeenCalledTimes(1);
		expect(
			screen.queryByRole("button", { name: "Открепить" }),
		).not.toBeInTheDocument();
	});

	it("rechecks current can_pin at click time before calling togglePin", async () => {
		const initialCapabilities = capabilities({
			can_list_pins: true,
			can_pin: true,
		});
		useMessengerStore.getState().setChannels([channel(initialCapabilities)]);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		seedPreview(queryClient);

		renderWithProviders(
			<PinnedMessages
				channelId={channelId}
				capabilities={initialCapabilities}
			/>,
			{ queryClient },
		);
		openPinnedMessages();
		const unpin = await screen.findByRole("button", { name: "Открепить" });
		act(() => {
			useMessengerStore.getState().updateChannel(channelId, {
				capabilities: capabilities({ can_list_pins: true, can_pin: false }),
			});
		});
		fireEvent.click(unpin);
		await act(async () => Promise.resolve());

		expect(togglePin).not.toHaveBeenCalled();
	});

	it("does not invalidate pins when the channel is revoked during unpin", async () => {
		const initialCapabilities = capabilities({
			can_list_pins: true,
			can_pin: true,
		});
		useMessengerStore.getState().setChannels([channel(initialCapabilities)]);
		const unpin = deferred<{ pinned: boolean }>();
		vi.mocked(togglePin).mockReturnValueOnce(unpin.promise);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		seedPreview(queryClient);
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");

		renderWithProviders(
			<PinnedMessages
				channelId={channelId}
				capabilities={initialCapabilities}
			/>,
			{ queryClient },
		);
		openPinnedMessages();
		fireEvent.click(await screen.findByRole("button", { name: "Открепить" }));
		await waitFor(() =>
			expect(togglePin).toHaveBeenCalledWith(channelId, pin.message_id),
		);

		act(() => useMessengerStore.getState().revokeChannel(channelId));
		await waitFor(() =>
			expect(
				screen.queryByRole("dialog", { name: "Закреплённые сообщения" }),
			).not.toBeInTheDocument(),
		);
		await act(async () => {
			unpin.resolve({ pinned: false });
			await unpin.promise;
			await Promise.resolve();
		});

		expect(invalidate).not.toHaveBeenCalled();
		expect(getChannelPins).toHaveBeenCalledTimes(1);
	});

	it("does not invalidate pins when the current channel DTO loses list permission during unpin", async () => {
		const initialCapabilities = capabilities({
			can_list_pins: true,
			can_pin: true,
		});
		useMessengerStore.getState().setChannels([channel(initialCapabilities)]);
		const unpin = deferred<{ pinned: boolean }>();
		vi.mocked(togglePin).mockReturnValueOnce(unpin.promise);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		seedPreview(queryClient);
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");

		const { rerender } = renderWithProviders(
			<PinnedMessages
				channelId={channelId}
				capabilities={initialCapabilities}
			/>, 
			{ queryClient },
		);
		openPinnedMessages();
		fireEvent.click(await screen.findByRole("button", { name: "Открепить" }));
		await waitFor(() =>
			expect(togglePin).toHaveBeenCalledWith(channelId, pin.message_id),
		);

		rerender(
			<PinnedMessages
				channelId={channelId}
				capabilities={capabilities()}
			/>, 
		);
		await waitFor(() =>
			expect(
				screen.queryByRole("dialog", { name: "Закреплённые сообщения" }),
			).not.toBeInTheDocument(),
		);
		await act(async () => {
			unpin.resolve({ pinned: false });
			await unpin.promise;
			await Promise.resolve();
		});

		expect(invalidate).not.toHaveBeenCalled();
		expect(getChannelPins).toHaveBeenCalledTimes(1);
	});
});
